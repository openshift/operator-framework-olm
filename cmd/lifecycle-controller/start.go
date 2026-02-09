package main

import (
	"cmp"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/go-logr/logr"
	configv1 "github.com/openshift/api/config/v1"
	tlsutil "github.com/openshift/controller-runtime-common/pkg/tls"
	"github.com/openshift/library-go/pkg/crypto"
	"github.com/openshift/operator-framework-olm/pkg/leaderelection"
	controllers "github.com/openshift/operator-framework-olm/pkg/lifecycle-controller"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsfilters "sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
)

const (
	defaultMetricsAddr     = ":8443"
	defaultHealthCheckAddr = ":8081"
	leaderElectionID       = "lifecycle-controller-lock"
)

var (
	disableLeaderElection      bool
	healthCheckAddr            string
	metricsAddr                string
	catalogSourceLabelSelector string
	catalogSourceFieldSelector string
	tlsCertFile                string
	tlsKeyFile                 string
)

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "start",
		Short:        "Start the Lifecycle Controller",
		SilenceUsage: true,
		RunE:         run,
	}

	cmd.Flags().StringVar(&healthCheckAddr, "health", defaultHealthCheckAddr, "health check address")
	cmd.Flags().StringVar(&metricsAddr, "metrics", defaultMetricsAddr, "metrics address")
	cmd.Flags().BoolVar(&disableLeaderElection, "disable-leader-election", false, "disable leader election")
	cmd.Flags().StringVar(&catalogSourceLabelSelector, "catalog-source-label-selector", "", "label selector for catalog sources to manage (empty means all)")
	cmd.Flags().StringVar(&catalogSourceFieldSelector, "catalog-source-field-selector", "", "field selector for catalog sources to manage (empty means all)")
	cmd.Flags().StringVar(&tlsCertFile, "tls-cert", "", "path to TLS certificate file for metrics server")
	cmd.Flags().StringVar(&tlsKeyFile, "tls-key", "", "path to TLS key file for metrics server")
	_ = cmd.MarkFlagRequired("tls-cert")
	_ = cmd.MarkFlagRequired("tls-key")
	return cmd
}

func run(_ *cobra.Command, _ []string) error {
	ctx := ctrl.SetupSignalHandler()
	ctrl.SetLogger(klog.NewKlogr())
	setupLog := ctrl.Log.WithName("setup")

	cfg, err := loadStartConfig(ctx)
	if err != nil {
		return fmt.Errorf("unable to load startup configuration: %v", err)
	}
	logConfig(cfg, setupLog)

	mgr, err := setupManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to setup manager instance: %v", err)
	}

	tlsProfileChan, err := setupTLSProfileWatcher(mgr, cfg)
	if err != nil {
		return fmt.Errorf("unable to setup TLS profile watcher: %v", err)
	}
	defer close(tlsProfileChan)

	if err := setupLifecycleServerController(mgr, cfg, tlsProfileChan); err != nil {
		return fmt.Errorf("unable to setup lifecycle server controller: %v", err)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("unable to start manager: %v", err)
	}

	return nil
}

type startConfig struct {
	Namespace string
	Version   string

	ServerImage                string
	CatalogSourceFieldSelector fields.Selector
	CatalogSourceLabelSelector labels.Selector
	RESTConfig                 *rest.Config
	Scheme                     *runtime.Scheme

	LeaderElection configv1.LeaderElection

	InitialTLSProfileSpec   configv1.TLSProfileSpec
	TLSConfigProvider       *controllers.TLSConfigProvider
	EnableTLSProfileWatcher bool
}

func loadStartConfig(ctx context.Context) (*startConfig, error) {
	cfg := &startConfig{
		Namespace:   os.Getenv("NAMESPACE"),
		Version:     cmp.Or(os.Getenv("RELEASE_VERSION"), "unknown"),
		ServerImage: os.Getenv("LIFECYCLE_SERVER_IMAGE"),
	}
	if cfg.Namespace == "" && !disableLeaderElection {
		return nil, fmt.Errorf("NAMESPACE environment variable is required when leader election is enabled")
	}
	if cfg.ServerImage == "" {
		return nil, fmt.Errorf("LIFECYCLE_SERVER_IMAGE environment variable is required")
	}

	// Using a function to load the keypair each time means that we automatically pick up the new certificate when it reloads.
	getCertificate := func(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
		cert, err := tls.LoadX509KeyPair(tlsCertFile, tlsKeyFile)
		if err != nil {
			return nil, err
		}
		return &cert, nil
	}
	_, err := getCertificate(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate/key: %v", err)
	}
	cfg.CatalogSourceFieldSelector, err = fields.ParseSelector(catalogSourceFieldSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to parse catalog source field selector %q: %v", catalogSourceFieldSelector, err)
	}
	cfg.CatalogSourceLabelSelector, err = labels.Parse(catalogSourceLabelSelector)
	if err != nil {
		return nil, fmt.Errorf("failed to parse catalog source label selector %q: %v", catalogSourceLabelSelector, err)
	}
	cfg.RESTConfig, err = ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get rest config: %v", err)
	}
	cfg.Scheme = setupScheme()
	cfg.LeaderElection = leaderelection.GetLeaderElectionConfig(ctrl.Log.WithName("leaderelection"), cfg.RESTConfig, !disableLeaderElection)

	cfg.InitialTLSProfileSpec, cfg.EnableTLSProfileWatcher, err = getInitialTLSProfile(ctx, cfg.RESTConfig, cfg.Scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to get initial TLS security profile: %v", err)
	}
	cfg.TLSConfigProvider = controllers.NewTLSConfigProvider(getCertificate, cfg.InitialTLSProfileSpec)
	return cfg, nil
}

func logConfig(cfg *startConfig, log logr.Logger) {
	log.Info("starting lifecycle-controller", "version", cfg.Version)
	log.Info("config", "lifecycleServerImage", cfg.ServerImage)
	if !cfg.CatalogSourceLabelSelector.Empty() {
		log.Info("config", "catalogSourceLabelSelector", cfg.CatalogSourceLabelSelector.String())
	}
	if !cfg.CatalogSourceFieldSelector.Empty() {
		log.Info("config", "catalogSourceFieldSelector", cfg.CatalogSourceFieldSelector.String())
	}
	tlsProfile, unsupportedCiphers := cfg.TLSConfigProvider.Get()
	log.Info("config", "tlsMinVersion", crypto.TLSVersionToNameOrDie(tlsProfile.MinVersion))
	log.Info("config", "tlsCipherSuites", crypto.CipherSuitesToNamesOrDie(tlsProfile.CipherSuites))
	if len(unsupportedCiphers) > 0 {
		log.Error(errors.New("ignored config"), "unsupported TLS cipher suites", "tlsCipherSuites", unsupportedCiphers)
	}
}

func getInitialTLSProfile(ctx context.Context, restConfig *rest.Config, sch *runtime.Scheme) (configv1.TLSProfileSpec, bool, error) {
	cl, err := client.New(restConfig, client.Options{Scheme: sch})
	if err != nil {
		return configv1.TLSProfileSpec{}, false, fmt.Errorf("failed to create client: %v", err)
	}
	initialTLSProfileSpec, err := tlsutil.FetchAPIServerTLSProfile(ctx, cl)
	if err != nil {
		return *configv1.TLSProfiles[crypto.DefaultTLSProfileType], false, nil
	}
	return initialTLSProfileSpec, true, nil
}

func setupManager(cfg *startConfig) (manager.Manager, error) {
	mgr, err := ctrl.NewManager(cfg.RESTConfig, manager.Options{
		Scheme: cfg.Scheme,
		Metrics: metricsserver.Options{
			BindAddress:    metricsAddr,
			SecureServing:  true,
			FilterProvider: metricsfilters.WithAuthenticationAndAuthorization,
			TLSOpts: []func(*tls.Config){func(tlsConfig *tls.Config) {
				tlsConfig.GetConfigForClient = func(*tls.ClientHelloInfo) (*tls.Config, error) {
					tlsCfg, _ := cfg.TLSConfigProvider.Get()
					return tlsCfg, nil
				}
			}},
		},
		LeaderElection:                !cfg.LeaderElection.Disable,
		LeaderElectionNamespace:       cfg.Namespace,
		LeaderElectionID:              leaderElectionID,
		LeaseDuration:                 &cfg.LeaderElection.LeaseDuration.Duration,
		RenewDeadline:                 &cfg.LeaderElection.RenewDeadline.Duration,
		RetryPeriod:                   &cfg.LeaderElection.RetryPeriod.Duration,
		HealthProbeBindAddress:        healthCheckAddr,
		LeaderElectionReleaseOnCancel: true,
		Cache: cache.Options{
			ByObject: map[client.Object]cache.ByObject{
				&operatorsv1alpha1.CatalogSource{}: {},
				&corev1.Pod{}: {
					Label: catalogPodLabelSelector(),
				},
				&appsv1.Deployment{}: {
					Label: controllers.LifecycleServerLabelSelector(),
				},
				&corev1.ServiceAccount{}: {
					Label: controllers.LifecycleServerLabelSelector(),
				},
				&corev1.Service{}: {
					Label: controllers.LifecycleServerLabelSelector(),
				},
				&networkingv1.NetworkPolicy{}: {
					Label: controllers.LifecycleServerLabelSelector(),
				},
				&configv1.APIServer{}: {
					Field: fields.SelectorFromSet(fields.Set{"metadata.name": "cluster"}),
				},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create manager: %v", err)
	}

	// Add health check endpoint (used for both liveness and readiness probes)
	if err := mgr.AddHealthzCheck("healthz", func(req *http.Request) error {
		return nil
	}); err != nil {
		return nil, fmt.Errorf("failed to configure health check handler: %v", err)
	}
	return mgr, nil
}

func setupTLSProfileWatcher(mgr manager.Manager, cfg *startConfig) (chan event.TypedGenericEvent[configv1.TLSProfileSpec], error) {
	tlsChangeChan := make(chan event.TypedGenericEvent[configv1.TLSProfileSpec])

	if !cfg.EnableTLSProfileWatcher {
		return tlsChangeChan, nil
	}

	log := ctrl.Log.WithName("tls-profile")
	tlsProfileReconciler := tlsutil.SecurityProfileWatcher{
		Client:                mgr.GetClient(),
		InitialTLSProfileSpec: cfg.InitialTLSProfileSpec,
		OnProfileChange: func(ctx context.Context, oldTLSProfileSpec, newTLSProfileSpec configv1.TLSProfileSpec) {
			cfg.TLSConfigProvider.UpdateProfile(newTLSProfileSpec)
			log.Info("applying new TLS profile spec",
				"minVersion", newTLSProfileSpec.MinTLSVersion,
				"cipherSuites", newTLSProfileSpec.Ciphers,
			)

			_, unsupportedCiphers := cfg.TLSConfigProvider.Get()
			if len(unsupportedCiphers) > 0 {
				log.Info("ignoring unsupported ciphers found in TLS profile", "unsupportedCiphers", unsupportedCiphers)
			}
			tlsChangeChan <- event.TypedGenericEvent[configv1.TLSProfileSpec]{Object: newTLSProfileSpec}
		},
	}

	if err := tlsProfileReconciler.SetupWithManager(mgr); err != nil {
		return nil, err
	}
	return tlsChangeChan, nil
}

func setupLifecycleServerController(mgr manager.Manager, cfg *startConfig, tlsProfileChan <-chan event.TypedGenericEvent[configv1.TLSProfileSpec]) error {
	reconciler := &controllers.LifecycleServerReconciler{
		Client:                     mgr.GetClient(),
		Log:                        ctrl.Log.WithName("controllers").WithName("lifecycle-server"),
		Scheme:                     mgr.GetScheme(),
		ServerImage:                cfg.ServerImage,
		CatalogSourceLabelSelector: cfg.CatalogSourceLabelSelector,
		CatalogSourceFieldSelector: cfg.CatalogSourceFieldSelector,
		TLSConfigProvider:          cfg.TLSConfigProvider,
	}

	if err := reconciler.SetupWithManager(mgr, tlsProfileChan); err != nil {
		return fmt.Errorf("unable to setup lifecycle server controller: %v", err)
	}
	return nil
}
