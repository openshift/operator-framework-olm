package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/apiserver"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/lib/server"

	"k8s.io/apimachinery/pkg/fields"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	"github.com/openshift/operator-framework-olm/pkg/leaderelection"
	controllers "github.com/openshift/operator-framework-olm/pkg/package-server-manager"
	//+kubebuilder:scaffold:imports
)

const (
	defaultName                 = "packageserver"
	defaultNamespace            = "openshift-operator-lifecycle-manager"
	defaultMetricsPort          = "0" // Disable controller-runtime metrics (using pkg/lib/server instead)
	defaultHealthCheckPort      = ""  // Disable controller-runtime health (using pkg/lib/server instead)
	defaultPprofPort            = ":6060"
	defaultInterval             = ""
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
	metricsAddr, err := cmd.Flags().GetString("metrics")
	if err != nil {
		return err
	}
	tlsCertPath, err := cmd.Flags().GetString("tls-cert")
	if err != nil {
		return err
	}
	tlsKeyPath, err := cmd.Flags().GetString("tls-key")
	if err != nil {
		return err
	}
	clientCAPath, err := cmd.Flags().GetString("client-ca")
	if err != nil {
		return err
	}

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	setupLog := ctrl.Log.WithName("setup")

	restConfig := ctrl.GetConfigOrDie()

	// Create logrus logger for the server library
	logger := logrus.New()

	// Setup APIServer TLS configuration for HTTPS servers
	apiServerTLSQuerier, apiServerFactory, err := apiserver.SetupAPIServerTLSConfig(logger, restConfig)
	if err != nil {
		setupLog.Error(err, "failed to setup APIServer TLS configuration")
		return err
	}

	// Start HTTPS server with metrics/health endpoints
	listenAndServe, err := server.GetListenAndServeFunc(
		server.WithLogger(logger),
		server.WithTLS(&tlsCertPath, &tlsKeyPath, &clientCAPath),
		server.WithKubeConfig(restConfig),
		server.WithAPIServerTLSQuerier(apiServerTLSQuerier),
	)
	if err != nil {
		setupLog.Error(err, "failed to setup health/metric/pprof service")
		return err
	}

	go func() {
		if err := listenAndServe(); err != nil {
			setupLog.Error(err, "server error")
		}
	}()
	le := leaderelection.GetLeaderElectionConfig(setupLog, restConfig, !disableLeaderElection)

	packageserverCSVFields := fields.Set{"metadata.name": name}
	serviceaccountFields := fields.Set{"metadata.name": "olm-operator-serviceaccount"}
	clusterroleFields := fields.Set{"metadata.name": "system:controller:operator-lifecycle-manager"}
	clusterrolebindingFields := fields.Set{"metadata.name": "olm-operator-binding-openshift-operator-lifecycle-manager"}
	mgr, err := ctrl.NewManager(restConfig, manager.Options{
		Scheme:                        setupScheme(),
		Metrics:                       metricsserver.Options{BindAddress: metricsAddr},
		LeaderElection:                !disableLeaderElection,
		LeaderElectionNamespace:       namespace,
		LeaderElectionID:              leaderElectionConfigmapName,
		LeaseDuration:                 &le.LeaseDuration.Duration,
		RenewDeadline:                 &le.RenewDeadline.Duration,
		RetryPeriod:                   &le.RetryPeriod.Duration,
		HealthProbeBindAddress:        healthCheckAddr,
		PprofBindAddress:              pprofAddr,
		LeaderElectionReleaseOnCancel: true,
		Cache: cache.Options{
			DefaultNamespaces: map[string]cache.Config{
				namespace: {},
			},
			ByObject: map[client.Object]cache.ByObject{
				&olmv1alpha1.ClusterServiceVersion{}: {
					Field: packageserverCSVFields.AsSelector(),
				},
				&corev1.ServiceAccount{}: {
					Field: serviceaccountFields.AsSelector(),
				},
				&rbacv1.ClusterRole{}: {
					Field: clusterroleFields.AsSelector(),
				},
				&rbacv1.ClusterRoleBinding{}: {
					Field: clusterrolebindingFields.AsSelector(),
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

	// Health checks are now handled by pkg/lib/server (not controller-runtime)
	// +kubebuilder:scaffold:builder

	// Set up signal handler context
	ctx := ctrl.SetupSignalHandler()

	// Start APIServer informer factory if on OpenShift
	if apiServerFactory != nil {
		apiServerFactory.Start(ctx.Done())
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running manager")
		return err
	}

	return nil
}
