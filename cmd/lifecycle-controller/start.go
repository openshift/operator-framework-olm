package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/spf13/cobra"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsfilters "sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/source"

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
	initialTLSConfig := controllers.GetClusterTLSConfig(ctx, tempClient, setupLog)

	// Create a TLS config provider for dynamic updates
	tlsProvider := controllers.NewTLSConfigProvider(initialTLSConfig)

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
					cfg.GetConfigForClient = func(info *tls.ClientHelloInfo) (*tls.Config, error) {
						cfg := tlsProvider.Get()
						return cfg.Clone(), nil
					}
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

	tlsProfileLog := ctrl.Log.WithName("controllers").WithName("tlsprofile-controller")
	tlsProfileReconciler := controllers.ClusterTLSProfileReconciler{
		Client:      mgr.GetClient(),
		Log:         tlsProfileLog,
		TLSProvider: tlsProvider,
		OnChange: func(prev, cur *tls.Config) {
			// Trigger reconciliation of all CatalogSources to update lifecycle-server deployments
			var catalogSources operatorsv1alpha1.CatalogSourceList
			if err := mgr.GetClient().List(ctx, &catalogSources); err != nil {
				tlsProfileLog.Error(err, "failed to list CatalogSources to requeue for TLS reconfiguration; CatalogSources will not receive new TLS configuration until their next reconciliation")
				return
			}

			tlsProfileLog.Info("requeueing CatalogSources TLS reconfiguration", "count", len(catalogSources.Items))

			// Send events to trigger reconciliation
			for i := range catalogSources.Items {
				cs := &catalogSources.Items[i]
				tlsChangeChan <- event.GenericEvent{Object: cs}
			}
		},
	}
	// Set up TLSProfileReconciler to reconcile TLS profile changes.
	if err := tlsProfileReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "failed to setup TLSProfile watcher")
		return err
	}

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

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		return err
	}

	return nil
}
