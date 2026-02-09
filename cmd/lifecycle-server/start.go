package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/openshift/library-go/pkg/crypto"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"

	"k8s.io/klog/v2"

	server "github.com/openshift/operator-framework-olm/pkg/lifecycle-server"
)

const (
	defaultFBCPath     = "/catalog/configs"
	defaultListenAddr  = ":8443"
	defaultHealthAddr  = ":8081"
	defaultTLSCertPath = "/var/run/secrets/serving-cert/tls.crt"
	defaultTLSKeyPath  = "/var/run/secrets/serving-cert/tls.key"
	shutdownTimeout    = 10 * time.Second
)

var (
	fbcPath            string
	listenAddr         string
	healthAddr         string
	tlsCertPath        string
	tlsKeyPath         string
	tlsMinVersionStr   string
	tlsCipherSuiteStrs []string
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
	cmd.Flags().StringVar(&tlsMinVersionStr, "tls-min-version", "", "minimum TLS version")
	cmd.Flags().StringSliceVar(&tlsCipherSuiteStrs, "tls-cipher-suites", nil, "comma-separated list of cipher suites")

	return cmd
}

func parseTLSFlags(certPath, keyPath, minVersionStr string, cipherSuiteStrs []string) (*tls.Config, error) {
	// Using a function to load the keypair each time means that we automatically pick up the new certificate when it reloads.
	getCertificate := func(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
		cert, err := tls.LoadX509KeyPair(certPath, keyPath)
		if err != nil {
			return nil, err
		}
		return &cert, nil
	}
	if _, err := getCertificate(nil); err != nil {
		return nil, fmt.Errorf("unable to load TLS certificate: %v", err)
	}

	minVersion, err := crypto.TLSVersion(minVersionStr)
	if err != nil {
		return nil, fmt.Errorf("invalid TLS minimum version: %s", minVersionStr)
	}

	var (
		cipherSuites    []uint16
		cipherSuiteErrs []error
	)
	for _, tlsCipherSuiteStr := range cipherSuiteStrs {
		tlsCipherSuite, err := crypto.CipherSuite(tlsCipherSuiteStr)
		if err != nil {
			cipherSuiteErrs = append(cipherSuiteErrs, err)
		} else {
			cipherSuites = append(cipherSuites, tlsCipherSuite)
		}
	}
	if len(cipherSuiteErrs) != 0 {
		return nil, fmt.Errorf("invalid TLS cipher suites: %v", errors.Join(cipherSuiteErrs...))
	}

	return &tls.Config{
		GetCertificate: getCertificate,
		MinVersion:     minVersion,
		CipherSuites:   cipherSuites,
	}, nil
}

func run(_ *cobra.Command, _ []string) error {
	log := klog.NewKlogr()
	log.Info("starting lifecycle-server")

	tlsConfig, err := parseTLSFlags(tlsCertPath, tlsKeyPath, tlsMinVersionStr, tlsCipherSuiteStrs)
	if err != nil {
		return fmt.Errorf("failed to parse tls flags: %w", err)
	}

	// Create Kubernetes client for authn/authz
	restCfg := ctrl.GetConfigOrDie()
	httpClient, err := rest.HTTPClientFor(restCfg)
	if err != nil {
		log.Error(err, "failed to create http client")
		return err
	}

	authnzFilter, err := filters.WithAuthenticationAndAuthorization(restCfg, httpClient)
	if err != nil {
		log.Error(err, "failed to create authorization filter")
		return err
	}

	// Load lifecycle data from FBC
	log.Info("loading lifecycle data from FBC", "path", fbcPath)
	data, err := server.LoadLifecycleData(fbcPath)
	if err != nil {
		log.Error(err, "failed to load lifecycle data, starting with empty data")
		data = make(server.LifecycleIndex)
	}
	log.Info("loaded lifecycle data",
		"packageCount", data.CountPackages(),
		"blobCount", data.CountBlobs(),
		"versions", data.ListVersions(),
	)

	// Create HTTP apiHandler with authn/authz middleware
	baseHandler := server.NewHandler(data, log)
	apiHandler, err := authnzFilter(log, baseHandler)
	if err != nil {
		log.Error(err, "failed to create api handler")
		return err
	}

	// Create health apiHandler (no auth required)
	healthHandler := http.NewServeMux()
	healthHandler.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Create servers
	apiServer := cancelableServer{
		Server: &http.Server{
			Addr:      listenAddr,
			Handler:   apiHandler,
			TLSConfig: tlsConfig,
		},
		ShutdownTimeout: shutdownTimeout,
	}
	healthServer := cancelableServer{
		Server: &http.Server{
			Addr:    healthAddr,
			Handler: healthHandler,
		},
		ShutdownTimeout: shutdownTimeout,
	}

	eg, ctx := errgroup.WithContext(ctrl.SetupSignalHandler())
	eg.Go(func() error {
		if err := apiServer.ListenAndServeTLS(ctx, "", ""); err != nil {
			return fmt.Errorf("api server error: %w", err)
		}
		return nil
	})
	eg.Go(func() error {
		if err := healthServer.ListenAndServe(ctx); err != nil {
			return fmt.Errorf("health server error: %w", err)
		}
		return nil
	})
	return eg.Wait()
}

type cancelableServer struct {
	*http.Server
	ShutdownTimeout time.Duration
}

func (s *cancelableServer) ListenAndServe(ctx context.Context) error {
	return s.listenAndServe(ctx,
		func() error {
			return s.Server.ListenAndServe()
		},
		s.Server.Shutdown,
	)
}
func (s *cancelableServer) ListenAndServeTLS(ctx context.Context, certFile, keyFile string) error {
	return s.listenAndServe(ctx,
		func() error {
			return s.Server.ListenAndServeTLS(certFile, keyFile)
		},
		s.Server.Shutdown,
	)
}

func (s *cancelableServer) listenAndServe(ctx context.Context, runFunc func() error, cancelFunc func(context.Context) error) error {
	errChan := make(chan error, 1)
	go func() {
		errChan <- runFunc()
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), s.ShutdownTimeout)
		defer cancel()
		if err := cancelFunc(shutdownCtx); err != nil {
			return err
		}
		return nil
	}
}
