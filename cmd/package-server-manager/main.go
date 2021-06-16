package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	configv1 "github.com/openshift/api/config/v1"
	operatorsv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	controllers "github.com/openshift/operator-framework-olm/pkg/package-server"
	//+kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

const (
	defaultName      = "packageserver"
	defaultNamespace = "openshift-operator-lifecycle-manager"
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(configv1.Install(scheme))
	utilruntime.Must(operatorsv1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {
	cmd := newStartCmd()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "encountered an error while executing the binary: %v", err)
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

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	setupLog := ctrl.Log.WithName("setup")
	// TODO(tflannag): Setup leader election?
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), manager.Options{
		Scheme:             scheme,
		Namespace:          namespace,
		MetricsBindAddress: "0",
	})
	if err != nil {
		setupLog.Error(err, "failed to setup manager instance")
		return err
	}

	if err := (&controllers.PackageServerReconciler{
		Name:      name,
		Namespace: namespace,
		Client:    mgr.GetClient(),
		Log:       ctrl.Log.WithName("controllers").WithName(name),
		Scheme:    mgr.GetScheme(),
		Recorder:  mgr.GetEventRecorderFor(name),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", name)
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
