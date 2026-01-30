package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apiserver/pkg/authentication/authenticatorfactory"
	"k8s.io/apiserver/pkg/authorization/authorizerfactory"
	"k8s.io/apiserver/pkg/endpoints/filters"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	cliflag "k8s.io/component-base/cli/flag"
	"k8s.io/klog/v2"

	server "github.com/openshift/operator-framework-olm/pkg/lifecycle-server"
)

const (
	defaultFBCPath     = "/catalog/configs"
	defaultListenAddr  = ":8443"
	defaultHealthAddr  = "localhost:8081"
	defaultTLSCertPath = "/var/run/secrets/serving-cert/tls.crt"
	defaultTLSKeyPath  = "/var/run/secrets/serving-cert/tls.key"
	shutdownTimeout    = 10 * time.Second
)

var (
	fbcPath         string
	listenAddr      string
	healthAddr      string
	tlsCertPath     string
	tlsKeyPath      string
	tlsMinVersion   string
	tlsCipherSuites []string
)

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "start",
		Short:        "Start the Lifecycle Server",
		SilenceUsage: true,
		RunE:         run,
	}

	cmd.Flags().StringVar(&fbcPath, "fbc-path", defaultFBCPath, "path to FBC catalog data")
	cmd.Flags().StringVar(&listenAddr, "listen", defaultListenAddr, "address to listen on for HTTPS API")
	cmd.Flags().StringVar(&healthAddr, "health", defaultHealthAddr, "address to listen on for health checks")
	cmd.Flags().StringVar(&tlsCertPath, "tls-cert", defaultTLSCertPath, "path to TLS certificate")
	cmd.Flags().StringVar(&tlsKeyPath, "tls-key", defaultTLSKeyPath, "path to TLS private key")
	cmd.Flags().StringVar(&tlsMinVersion, "tls-min-version", "", "minimum TLS version (VersionTLS12 or VersionTLS13)")
	cmd.Flags().StringSliceVar(&tlsCipherSuites, "tls-cipher-suites", nil, "comma-separated list of cipher suites")

	return cmd
}

func run(_ *cobra.Command, _ []string) error {
	log := klog.NewKlogr()
	log.Info("starting lifecycle-server")

	// Parse TLS configuration
	var tlsMinVersionID uint16
	var err error
	if tlsMinVersion != "" {
		tlsMinVersionID, err = cliflag.TLSVersion(tlsMinVersion)
		if err != nil {
			return fmt.Errorf("invalid tls-min-version: %w", err)
		}
	}

	var tlsCipherSuiteIDs []uint16
	if len(tlsCipherSuites) > 0 {
		tlsCipherSuiteIDs, err = cliflag.TLSCipherSuites(tlsCipherSuites)
		if err != nil {
			return fmt.Errorf("invalid tls-cipher-suites: %w", err)
		}
	}

	// Create Kubernetes client for authn/authz
	config, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to get in-cluster config: %w", err)
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	// Create delegating authenticator (uses TokenReview)
	authnConfig := authenticatorfactory.DelegatingAuthenticatorConfig{
		Anonymous:               nil, // disable anonymous auth
		TokenAccessReviewClient: kubeClient.AuthenticationV1(),
		CacheTTL:                2 * time.Minute,
	}
	authenticator, _, err := authnConfig.New()
	if err != nil {
		return fmt.Errorf("failed to create authenticator: %w", err)
	}

	// Create delegating authorizer (uses SubjectAccessReview)
	authzConfig := authorizerfactory.DelegatingAuthorizerConfig{
		SubjectAccessReviewClient: kubeClient.AuthorizationV1(),
		AllowCacheTTL:             5 * time.Minute,
		DenyCacheTTL:              30 * time.Second,
	}
	authorizer, err := authzConfig.New()
	if err != nil {
		return fmt.Errorf("failed to create authorizer: %w", err)
	}

	// Load lifecycle data from FBC
	log.Info("loading lifecycle data from FBC", "path", fbcPath)
	data, err := server.LoadLifecycleData(fbcPath)
	if err != nil {
		log.Error(err, "failed to load lifecycle data, starting with empty data")
		data = make(server.LifecycleIndex)
	}
	log.Info("loaded lifecycle data",
		"blobCount", server.CountBlobs(data),
		"versionCount", len(data),
		"versions", server.ListVersions(data),
	)

	// Create HTTP handler with authn/authz middleware
	baseHandler := server.NewHandler(data, log)

	// Wrap with authorization
	authorizedHandler := filters.WithAuthorization(baseHandler, authorizer, nil)

	// Wrap with authentication
	handler := filters.WithAuthentication(
		authorizedHandler,
		authenticator,
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		}),
		nil,
		nil,
	)

	// Wrap with request info (required by authorization filter)
	requestInfoResolver := &request.RequestInfoFactory{
		APIPrefixes:          sets.NewString("api"),
		GrouplessAPIPrefixes: sets.NewString("api"),
	}
	handler = filters.WithRequestInfo(handler, requestInfoResolver)

	// Create health handler (no auth required)
	healthHandler := http.NewServeMux()
	healthHandler.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Load TLS certificate
	cert, err := tls.LoadX509KeyPair(tlsCertPath, tlsKeyPath)
	if err != nil {
		return fmt.Errorf("failed to load TLS certificate: %w", err)
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tlsMinVersionID,
		CipherSuites: tlsCipherSuiteIDs,
	}

	// Create servers
	apiServer := &http.Server{
		Addr:      listenAddr,
		Handler:   handler,
		TLSConfig: tlsConfig,
	}
	healthServer := &http.Server{
		Addr:    healthAddr,
		Handler: healthHandler,
	}

	// Start servers
	errCh := make(chan error, 2)
	go func() {
		log.Info("starting API server (HTTPS)", "addr", listenAddr)
		// Cert paths are empty since TLSConfig already has certificates loaded
		if err := apiServer.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("API server error: %w", err)
		}
	}()
	go func() {
		log.Info("starting health server", "addr", healthAddr)
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("health server error: %w", err)
		}
	}()

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Info("received shutdown signal", "signal", sig)
	case err := <-errCh:
		log.Error(err, "server error")
		return err
	}

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	log.Info("shutting down servers")
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Error(err, "API server shutdown error")
	}
	if err := healthServer.Shutdown(ctx); err != nil {
		log.Error(err, "health server shutdown error")
	}

	return nil
}
