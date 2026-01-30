package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"sync"
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
	"sigs.k8s.io/controller-runtime/pkg/event"
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
	defaultHealthCheckAddr = ":8081"
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

// TLSConfigProvider provides thread-safe access to dynamically updated TLS configuration.
// It implements controllers.TLSConfigProvider interface.
type TLSConfigProvider struct {
	mu     sync.RWMutex
	config *tlsConfig
}

// NewTLSConfigProvider creates a new TLSConfigProvider with the given initial config.
func NewTLSConfigProvider(initial *tlsConfig) *TLSConfigProvider {
	return &TLSConfigProvider{config: initial}
}

// Get returns the current TLS configuration.
func (p *TLSConfigProvider) Get() *tlsConfig {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config
}

// Update sets a new TLS configuration.
func (p *TLSConfigProvider) Update(cfg *tlsConfig) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config = cfg
}

// GetMinVersion returns the current TLS minimum version string.
func (p *TLSConfigProvider) GetMinVersion() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config.minVersionString
}

// GetCipherSuites returns the current TLS cipher suites.
func (p *TLSConfigProvider) GetCipherSuites() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config.cipherSuiteStrings
}

// GetConfigForClient returns a TLS config callback for dynamic TLS configuration.
func (p *TLSConfigProvider) GetConfigForClient() func(*tls.ClientHelloInfo) (*tls.Config, error) {
	return func(*tls.ClientHelloInfo) (*tls.Config, error) {
		cfg := p.Get()
		return &tls.Config{
			MinVersion:   cfg.minVersion,
			CipherSuites: cfg.cipherSuites,
		}, nil
	}
}

// getInitialTLSConfig reads the APIServer "cluster" resource and extracts TLS settings.
// Falls back to Intermediate profile defaults if the resource doesn't exist.
func getInitialTLSConfig(ctx context.Context, c client.Client, log logr.Logger) (*tlsConfig, error) {
	var apiServer configv1.APIServer
	err := c.Get(ctx, types.NamespacedName{Name: clusterAPIServerName}, &apiServer)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("APIServer 'cluster' not found, using TLS profile defaults")
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

	// Create a TLS config provider for dynamic updates
	tlsProvider := NewTLSConfigProvider(initialTLSConfig)

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
					// Use GetConfigForClient for dynamic TLS configuration
					cfg.GetConfigForClient = tlsProvider.GetConfigForClient()
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

	// Create channel for TLS config change notifications
	// The apiServerWatcher sends events to this channel after updating the TLS provider
	tlsChangeChan := make(chan event.GenericEvent)
	tlsChangeSource := source.Channel(tlsChangeChan, &handler.EnqueueRequestForObject{})

	reconciler := &controllers.LifecycleControllerReconciler{
		Client:                     mgr.GetClient(),
		Log:                        ctrl.Log.WithName("controllers").WithName("lifecycle-controller"),
		Scheme:                     mgr.GetScheme(),
		ServerImage:                serverImage,
		CatalogSourceLabelSelector: labelSelector,
		CatalogSourceFieldSelector: fieldSelector,
		TLSConfigProvider:          tlsProvider,
	}

	if err := reconciler.SetupWithManager(mgr, tlsChangeSource); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "lifecycle-controller")
		return err
	}

	// Add health check endpoint (used for both liveness and readiness probes)
	if err := mgr.AddHealthzCheck("healthz", func(req *http.Request) error {
		return nil
	}); err != nil {
		setupLog.Error(err, "unable to set up health check")
		return err
	}

	// Set up APIServer watcher to update TLS config and trigger CatalogSource reconciliation
	if err := setupAPIServerWatcher(mgr, tlsProvider, tlsChangeChan, setupLog); err != nil {
		setupLog.Error(err, "failed to setup APIServer watcher")
		return err
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		return err
	}

	return nil
}

// apiServerWatcher watches the APIServer "cluster" resource and updates TLS config dynamically
type apiServerWatcher struct {
	client        client.Client
	log           logr.Logger
	tlsProvider   *TLSConfigProvider
	tlsChangeChan chan<- event.GenericEvent
}

func setupAPIServerWatcher(mgr manager.Manager, tlsProvider *TLSConfigProvider, tlsChangeChan chan<- event.GenericEvent, log logr.Logger) error {
	watcher := &apiServerWatcher{
		client:        mgr.GetClient(),
		log:           log.WithName("apiserver-watcher"),
		tlsProvider:   tlsProvider,
		tlsChangeChan: tlsChangeChan,
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

	var newConfig *tlsConfig

	var apiServer configv1.APIServer
	if err := w.client.Get(ctx, req.NamespacedName, &apiServer); err != nil {
		if errors.IsNotFound(err) {
			// APIServer deleted - use defaults
			w.log.Info("APIServer 'cluster' deleted, using TLS profile defaults")
			minVersion, cipherSuites := apiserver.GetSecurityProfileConfig(nil)
			newConfig = &tlsConfig{
				minVersion:         minVersion,
				cipherSuites:       cipherSuites,
				minVersionString:   tlsVersionToString(minVersion),
				cipherSuiteStrings: cipherSuiteIDsToNames(cipherSuites),
			}
		} else {
			return reconcile.Result{}, err
		}
	} else {
		// Get current TLS config from APIServer
		minVersion, cipherSuites := apiserver.GetSecurityProfileConfig(apiServer.Spec.TLSSecurityProfile)
		newConfig = &tlsConfig{
			minVersion:         minVersion,
			cipherSuites:       cipherSuites,
			minVersionString:   tlsVersionToString(minVersion),
			cipherSuiteStrings: cipherSuiteIDsToNames(cipherSuites),
		}
	}

	// Check if config changed
	currentConfig := w.tlsProvider.Get()
	if currentConfig.minVersion == newConfig.minVersion && cipherSuitesEqual(currentConfig.cipherSuites, newConfig.cipherSuites) {
		// No change
		return reconcile.Result{}, nil
	}

	w.log.Info("TLS security profile changed, updating configuration and triggering reconciliation",
		"oldMinVersion", currentConfig.minVersionString,
		"newMinVersion", newConfig.minVersionString,
		"oldCipherCount", len(currentConfig.cipherSuites),
		"newCipherCount", len(newConfig.cipherSuites),
	)

	// Update the provider
	w.tlsProvider.Update(newConfig)

	// Trigger reconciliation of all CatalogSources to update lifecycle-server deployments
	var catalogSources operatorsv1alpha1.CatalogSourceList
	if err := w.client.List(ctx, &catalogSources); err != nil {
		w.log.Error(err, "failed to list CatalogSources for reconciliation")
		return reconcile.Result{}, err
	}

	w.log.Info("triggering reconciliation for CatalogSources", "count", len(catalogSources.Items))

	// Send events to trigger reconciliation
	for i := range catalogSources.Items {
		cs := &catalogSources.Items[i]
		w.tlsChangeChan <- event.GenericEvent{Object: cs}
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
