package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"

	"k8s.io/apimachinery/pkg/fields"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/openshift/operator-framework-olm/pkg/leaderelection"
	controllers "github.com/openshift/operator-framework-olm/pkg/package-server-manager"
	//+kubebuilder:scaffold:imports
)

const (
	defaultName                 = "packageserver"
	defaultNamespace            = "openshift-operator-lifecycle-manager"
	defaultMetricsPort          = "0"
	defaultHealthCheckPort      = ":8080"
	defaultPprofPort            = ":6060"
	defaultInterval             = "5m"
	leaderElectionConfigmapName = "packageserver-controller-lock"
)

func main() {
	cmd := newStartCmd()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "encountered an error while executing the binary: %v\n", err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
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
	interval, err := cmd.Flags().GetString("interval")
	if err != nil {
		return err
	}

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	setupLog := ctrl.Log.WithName("setup")

	restConfig := ctrl.GetConfigOrDie()
	le := leaderelection.GetLeaderElectionConfig(setupLog, restConfig, !disableLeaderElection)

	packageserverCSVFields := fields.Set{"metadata.name": name}
	mgr, err := ctrl.NewManager(restConfig, manager.Options{
		Scheme:                  setupScheme(),
		Namespace:               namespace,
		MetricsBindAddress:      defaultMetricsPort,
		LeaderElection:          !disableLeaderElection,
		LeaderElectionNamespace: namespace,
		LeaderElectionID:        leaderElectionConfigmapName,
		LeaseDuration:           &le.LeaseDuration.Duration,
		RenewDeadline:           &le.RenewDeadline.Duration,
		RetryPeriod:             &le.RetryPeriod.Duration,
		HealthProbeBindAddress:  healthCheckAddr,
		PprofBindAddress:        pprofAddr,
		Cache: cache.Options{
			ByObject: map[client.Object]cache.ByObject{
				&olmv1alpha1.ClusterServiceVersion{}: {
					Field: packageserverCSVFields.AsSelector(),
				},
			},
		},
	})
	if err != nil {
		setupLog.Error(err, "failed to setup manager instance")
		return err
	}

	if err := (&controllers.PackageServerCSVReconciler{
		Name:      name,
		Namespace: namespace,
		Image:     os.Getenv("PACKAGESERVER_IMAGE"),
		Interval:  interval,
		Client:    mgr.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName(name),
		Scheme:    mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", name)
		return err
	}

	if err := mgr.AddReadyzCheck("ping", healthz.Ping); err != nil {
		setupLog.Error(err, "failed to establish a readyz check")
		return err
	}
	if err := mgr.AddHealthzCheck("ping", healthz.Ping); err != nil {
		setupLog.Error(err, "failed to establish a healthz check")
		return err
	}
	// +kubebuilder:scaffold:builder
	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		return err
	}

	return nil
}
