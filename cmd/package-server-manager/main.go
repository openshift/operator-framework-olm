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
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	controllers "github.com/openshift/operator-framework-olm/pkg/package-server-manager"
	//+kubebuilder:scaffold:imports
)

const (
	defaultName                 = "packageserver"
	defaultNamespace            = "openshift-operator-lifecycle-manager"
	defaultMetricsPort          = "0"
	defaultHealthCheckPort      = ":8080"
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

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	setupLog := ctrl.Log.WithName("setup")

	packageserverCSVFields := fields.Set{"metadata.name": name}
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), manager.Options{
		Scheme:                  setupScheme(),
		Namespace:               namespace,
		MetricsBindAddress:      defaultMetricsPort,
		LeaderElection:          !disableLeaderElection,
		LeaderElectionNamespace: namespace,
		LeaderElectionID:        leaderElectionConfigmapName,
		RetryPeriod:             timeDurationPtr(defaultRetryPeriod),
		RenewDeadline:           timeDurationPtr(defaultRenewDeadline),
		LeaseDuration:           timeDurationPtr(defaultLeaseDuration),
		HealthProbeBindAddress:  healthCheckAddr,
		NewCache: cache.BuilderWithOptions(cache.Options{
			SelectorsByObject: cache.SelectorsByObject{
				&olmv1alpha1.ClusterServiceVersion{}: {
					Field: packageserverCSVFields.AsSelector(),
				},
			},
		}),
	})
	if err != nil {
		setupLog.Error(err, "failed to setup manager instance")
		return err
	}

	if err := (&controllers.PackageServerCSVReconciler{
		Name:      name,
		Namespace: namespace,
		Image:     os.Getenv("PACKAGESERVER_IMAGE"),
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
