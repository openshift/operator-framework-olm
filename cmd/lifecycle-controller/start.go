package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"
	"k8s.io/klog/v2"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"

	configv1 "github.com/openshift/api/config/v1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsfilters "sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/apiserver"

	controllers "github.com/openshift/operator-framework-olm/pkg/lifecycle-controller"
)

const (
	defaultMetricsAddr     = ":8443"
	defaultHealthCheckAddr = "localhost:8081"
	leaderElectionID       = "lifecycle-controller-lock"

	// Leader election defaults per OpenShift conventions
	// https://github.com/openshift/enhancements/blob/master/CONVENTIONS.md#high-availability
	defaultLeaseDuration = 137 * time.Second
	defaultRenewDeadline = 107 * time.Second
	defaultRetryPeriod   = 26 * time.Second

	// Name of the cluster-scoped APIServer resource
	clusterAPIServerName = "cluster"
)

var (
	disableLeaderElection      bool
	healthCheckAddr            string
	metricsAddr                string
	catalogSourceLabelSelector string
	catalogSourceFieldSelector string
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

	return cmd
}

// catalogPodLabelSelector returns a label selector matching pods with olm.catalogSource label
func catalogPodLabelSelector() labels.Selector {
	// This call cannot fail: the label key is valid and selection.Exists requires no values.
	req, err := labels.NewRequirement("olm.catalogSource", selection.Exists, nil)
	if err != nil {
		// Panic on impossible error to satisfy static analysis and catch programming errors
		panic(fmt.Sprintf("BUG: failed to create label requirement: %v", err))
	}
	return labels.NewSelector().Add(*req)
}

// tlsConfig holds the TLS configuration extracted from the APIServer resource
type tlsConfig struct {
	minVersion   uint16
	cipherSuites []uint16
	// String representations for passing to lifecycle-server
	minVersionString   string
	cipherSuiteStrings []string
}

// getInitialTLSConfig reads the APIServer "cluster" resource and extracts TLS settings.
// Falls back to Intermediate profile defaults if the resource doesn't exist.
func getInitialTLSConfig(ctx context.Context, c client.Client, log logr.Logger) (*tlsConfig, error) {
	var apiServer configv1.APIServer
	err := c.Get(ctx, types.NamespacedName{Name: clusterAPIServerName}, &apiServer)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("APIServer 'cluster' not found, using Intermediate TLS profile defaults")
			minVersion, cipherSuites := apiserver.GetSecurityProfileConfig(nil)
			return &tlsConfig{
				minVersion:         minVersion,
				cipherSuites:       cipherSuites,
				minVersionString:   tlsVersionToString(minVersion),
				cipherSuiteStrings: cipherSuiteIDsToNames(cipherSuites),
			}, nil
		}
		return nil, fmt.Errorf("failed to get APIServer 'cluster': %w", err)
	}

	minVersion, cipherSuites := apiserver.GetSecurityProfileConfig(apiServer.Spec.TLSSecurityProfile)
	cfg := &tlsConfig{
		minVersion:         minVersion,
		cipherSuites:       cipherSuites,
		minVersionString:   tlsVersionToString(minVersion),
		cipherSuiteStrings: cipherSuiteIDsToNames(cipherSuites),
	}

	log.Info("loaded TLS configuration from APIServer",
		"profile", getTLSProfileName(apiServer.Spec.TLSSecurityProfile),
		"minVersion", cfg.minVersionString,
		"cipherCount", len(cfg.cipherSuites),
	)

	return cfg, nil
}

// tlsVersionToString converts a TLS version constant to its string name
func tlsVersionToString(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "VersionTLS10"
	case tls.VersionTLS11:
		return "VersionTLS11"
	case tls.VersionTLS12:
		return "VersionTLS12"
	case tls.VersionTLS13:
		return "VersionTLS13"
	default:
		return "VersionTLS12"
	}
}

// cipherSuiteIDsToNames converts TLS cipher suite IDs to their IANA names
func cipherSuiteIDsToNames(ids []uint16) []string {
	names := make([]string, 0, len(ids))
	for _, id := range ids {
		if suite := tls.CipherSuiteName(id); suite != "" {
			names = append(names, suite)
		}
	}
	return names
}

// getTLSProfileName returns the TLS profile name for logging
func getTLSProfileName(profile *configv1.TLSSecurityProfile) string {
	if profile == nil {
		return "Intermediate (default)"
	}
	if profile.Type == "" {
		return "Intermediate (default)"
	}
	return string(profile.Type)
}

func run(_ *cobra.Command, _ []string) error {
	serverImage := os.Getenv("LIFECYCLE_SERVER_IMAGE")
	if serverImage == "" {
		return fmt.Errorf("LIFECYCLE_SERVER_IMAGE environment variable is required")
	}

	namespace := os.Getenv("NAMESPACE")
	if !disableLeaderElection && namespace == "" {
		return fmt.Errorf("NAMESPACE environment variable is required when leader election is enabled")
	}

	ctrl.SetLogger(klog.NewKlogr())
	setupLog := ctrl.Log.WithName("setup")

	version := os.Getenv("RELEASE_VERSION")
	if version == "" {
		version = "unknown"
	}
	setupLog.Info("starting lifecycle-controller", "version", version)

	// Parse the catalog source label selector
	labelSelector, err := labels.Parse(catalogSourceLabelSelector)
	if err != nil {
		setupLog.Error(err, "failed to parse catalog-source-label-selector", "selector", catalogSourceLabelSelector)
		return fmt.Errorf("invalid catalog-source-label-selector %q: %w", catalogSourceLabelSelector, err)
	}
	setupLog.Info("using catalog source label selector", "selector", labelSelector.String())

	// Parse the catalog source field selector
	fieldSelector, err := fields.ParseSelector(catalogSourceFieldSelector)
	if err != nil {
		setupLog.Error(err, "failed to parse catalog-source-field-selector", "selector", catalogSourceFieldSelector)
		return fmt.Errorf("invalid catalog-source-field-selector %q: %w", catalogSourceFieldSelector, err)
	}
	setupLog.Info("using catalog source field selector", "selector", fieldSelector.String())

	restConfig := ctrl.GetConfigOrDie()
	scheme := setupScheme()

	// Create a temporary client to read initial TLS configuration
	tempClient, err := client.New(restConfig, client.Options{Scheme: scheme})
	if err != nil {
		setupLog.Error(err, "failed to create temporary client for TLS config")
		return err
	}

	// Get initial TLS configuration from APIServer "cluster"
	ctx := context.Background()
	initialTLSConfig, err := getInitialTLSConfig(ctx, tempClient, setupLog)
	if err != nil {
		setupLog.Error(err, "failed to get initial TLS configuration")
		return err
	}

	// Leader election timing defaults
	leaseDuration := defaultLeaseDuration
	renewDeadline := defaultRenewDeadline
	retryPeriod := defaultRetryPeriod

	mgr, err := ctrl.NewManager(restConfig, manager.Options{
		Scheme: scheme,
		Metrics: metricsserver.Options{
			BindAddress:    metricsAddr,
			SecureServing:  true,
			FilterProvider: metricsfilters.WithAuthenticationAndAuthorization,
			TLSOpts: []func(*tls.Config){
				func(cfg *tls.Config) {
					cfg.MinVersion = initialTLSConfig.minVersion
					cfg.CipherSuites = initialTLSConfig.cipherSuites
				},
			},
		},
		LeaderElection:                !disableLeaderElection,
		LeaderElectionNamespace:       namespace,
		LeaderElectionID:              leaderElectionID,
		LeaseDuration:                 &leaseDuration,
		RenewDeadline:                 &renewDeadline,
		RetryPeriod:                   &retryPeriod,
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
				&configv1.APIServer{}: {},
			},
		},
	})
	if err != nil {
		setupLog.Error(err, "failed to setup manager instance")
		return err
	}

	if err := (&controllers.LifecycleControllerReconciler{
		Client:                     mgr.GetClient(),
		Log:                        ctrl.Log.WithName("controllers").WithName("lifecycle-controller"),
		Scheme:                     mgr.GetScheme(),
		ServerImage:                serverImage,
		CatalogSourceLabelSelector: labelSelector,
		CatalogSourceFieldSelector: fieldSelector,
		TLSMinVersion:              initialTLSConfig.minVersionString,
		TLSCipherSuites:            initialTLSConfig.cipherSuiteStrings,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "lifecycle-controller")
		return err
	}

	// Set up APIServer watcher to exit on TLS config change
	if err := setupAPIServerWatcher(mgr, initialTLSConfig, setupLog); err != nil {
		setupLog.Error(err, "failed to setup APIServer watcher")
		return err
	}

	// Add health check endpoint (used for both liveness and readiness probes)
	if err := mgr.AddHealthzCheck("healthz", func(req *http.Request) error {
		return nil
	}); err != nil {
		setupLog.Error(err, "unable to set up health check")
		return err
	}

	// Set up signal handler context
	signalCtx := ctrl.SetupSignalHandler()

	setupLog.Info("starting manager")
	if err := mgr.Start(signalCtx); err != nil {
		setupLog.Error(err, "problem running manager")
		return err
	}

	return nil
}

// apiServerWatcher watches the APIServer "cluster" resource and exits if TLS config changes
type apiServerWatcher struct {
	client         client.Client
	log            logr.Logger
	initialMinVer  uint16
	initialCiphers []uint16
}

func setupAPIServerWatcher(mgr manager.Manager, initialCfg *tlsConfig, log logr.Logger) error {
	watcher := &apiServerWatcher{
		client:         mgr.GetClient(),
		log:            log.WithName("apiserver-watcher"),
		initialMinVer:  initialCfg.minVersion,
		initialCiphers: initialCfg.cipherSuites,
	}

	// Create a controller that watches APIServer and triggers reconcile
	return ctrl.NewControllerManagedBy(mgr).
		Named("apiserver-tls-watcher").
		WatchesRawSource(source.Kind(mgr.GetCache(), &configv1.APIServer{},
			handler.TypedEnqueueRequestsFromMapFunc(func(ctx context.Context, obj *configv1.APIServer) []reconcile.Request {
				if obj.Name == clusterAPIServerName {
					return []reconcile.Request{{NamespacedName: types.NamespacedName{Name: clusterAPIServerName}}}
				}
				return nil
			}),
		)).
		Complete(watcher)
}

func (w *apiServerWatcher) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	if req.Name != clusterAPIServerName {
		return reconcile.Result{}, nil
	}

	var apiServer configv1.APIServer
	if err := w.client.Get(ctx, req.NamespacedName, &apiServer); err != nil {
		if errors.IsNotFound(err) {
			// APIServer deleted - check if we had a non-default config
			defaultMin, defaultCiphers := apiserver.GetSecurityProfileConfig(nil)
			if w.initialMinVer != defaultMin || !cipherSuitesEqual(w.initialCiphers, defaultCiphers) {
				w.log.Info("APIServer 'cluster' deleted and initial config was non-default, exiting to pick up new defaults")
				os.Exit(0)
			}
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Get current TLS config
	currentMinVer, currentCiphers := apiserver.GetSecurityProfileConfig(apiServer.Spec.TLSSecurityProfile)

	// Compare with initial config
	if w.initialMinVer != currentMinVer || !cipherSuitesEqual(w.initialCiphers, currentCiphers) {
		w.log.Info("TLS security profile changed, exiting to pick up new configuration",
			"oldMinVersion", tlsVersionToString(w.initialMinVer),
			"newMinVersion", tlsVersionToString(currentMinVer),
			"oldCipherCount", len(w.initialCiphers),
			"newCipherCount", len(currentCiphers),
		)
		os.Exit(0)
	}

	return reconcile.Result{}, nil
}

// cipherSuitesEqual compares two cipher suite slices for equality
func cipherSuitesEqual(a, b []uint16) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}