package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	controllers "github.com/openshift/operator-framework-olm/pkg/lifecycle-controller"
)

const (
	defaultNamespace              = "openshift-operator-lifecycle-manager"
	defaultMetricsPort            = "0"
	defaultHealthCheckPort        = ":8081"
	defaultPprofPort              = ":6060"
	leaderElectionID              = "lifecycle-controller-lock"
	defaultCatalogSourceSelector  = "olm.operatorframework.io/lifecycle-server=true"

	// Leader election defaults per OpenShift conventions
	// https://github.com/openshift/enhancements/blob/master/CONVENTIONS.md#high-availability
	defaultLeaseDuration = 137 * time.Second
	defaultRenewDeadline = 107 * time.Second
	defaultRetryPeriod   = 26 * time.Second
)

func main() {
	cmd := newStartCmd()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "encountered an error while executing the binary: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	namespace, err := cmd.Flags().GetString("namespace")
	if err != nil {
		return err
	}
	disableLeaderElection, err := cmd.Flags().GetBool("disable-leader-election")
	if err != nil {
		return err
	}
	healthCheckAddr, err := cmd.Flags().GetString("health")
	if err != nil {
		return err
	}
	pprofAddr, err := cmd.Flags().GetString("pprof")
	if err != nil {
		return err
	}
	metricsAddr, err := cmd.Flags().GetString("metrics")
	if err != nil {
		return err
	}
	catalogSourceSelectorStr, err := cmd.Flags().GetString("catalog-source-selector")
	if err != nil {
		return err
	}

	serverImage := os.Getenv("LIFECYCLE_SERVER_IMAGE")
	if serverImage == "" {
		return fmt.Errorf("LIFECYCLE_SERVER_IMAGE environment variable is required")
	}

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	setupLog := ctrl.Log.WithName("setup")

	// Parse the catalog source label selector
	catalogSourceSelector, err := labels.Parse(catalogSourceSelectorStr)
	if err != nil {
		setupLog.Error(err, "failed to parse catalog-source-selector", "selector", catalogSourceSelectorStr)
		return fmt.Errorf("invalid catalog-source-selector %q: %w", catalogSourceSelectorStr, err)
	}
	setupLog.Info("using catalog source selector", "selector", catalogSourceSelector.String())

	restConfig := ctrl.GetConfigOrDie()

	// Leader election timing defaults
	leaseDuration := defaultLeaseDuration
	renewDeadline := defaultRenewDeadline
	retryPeriod := defaultRetryPeriod

	mgr, err := ctrl.NewManager(restConfig, manager.Options{
		Scheme:                        setupScheme(),
		Metrics:                       metricsserver.Options{BindAddress: metricsAddr},
		LeaderElection:                !disableLeaderElection,
		LeaderElectionNamespace:       namespace,
		LeaderElectionID:              leaderElectionID,
		LeaseDuration:                 &leaseDuration,
		RenewDeadline:                 &renewDeadline,
		RetryPeriod:                   &retryPeriod,
		HealthProbeBindAddress:        healthCheckAddr,
		PprofBindAddress:              pprofAddr,
		LeaderElectionReleaseOnCancel: true,
		Cache: cache.Options{
			ByObject: map[client.Object]cache.ByObject{
				// Watch all CatalogSources cluster-wide
				&operatorsv1alpha1.CatalogSource{}: {},
				// Watch all Pods cluster-wide (for catalog pods)
				&corev1.Pod{}: {},
				// Watch the lifecycle-server Deployment
				&appsv1.Deployment{}: {},
			},
		},
	})
	if err != nil {
		setupLog.Error(err, "failed to setup manager instance")
		return err
	}

	if err := (&controllers.LifecycleControllerReconciler{
		Client:                mgr.GetClient(),
		Log:                   ctrl.Log.WithName("controllers").WithName("lifecycle-controller"),
		Scheme:                mgr.GetScheme(),
		ServerImage:           serverImage,
		CatalogSourceSelector: catalogSourceSelector,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "lifecycle-controller")
		return err
	}

	// Add health check endpoint
	if err := mgr.AddHealthzCheck("healthz", func(req *http.Request) error {
		return nil
	}); err != nil {
		setupLog.Error(err, "unable to set up health check")
		return err
	}

	// Set up signal handler context
	ctx := ctrl.SetupSignalHandler()

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		return err
	}

	return nil
}
