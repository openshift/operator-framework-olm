package main

import (
	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "start",
		Short:        "Start the Lifecycle Controller",
		SilenceUsage: true,
		RunE:         run,
	}

	cmd.Flags().String("namespace", defaultNamespace, "namespace where the controller runs")
	cmd.Flags().String("health", defaultHealthCheckPort, "health check port")
	cmd.Flags().String("pprof", defaultPprofPort, "pprof port")
	cmd.Flags().String("metrics", defaultMetricsPort, "metrics port")
	cmd.Flags().Bool("disable-leader-election", false, "disable leader election")
	cmd.Flags().String("catalog-source-selector", defaultCatalogSourceSelector, "label selector for catalog sources to manage (e.g., 'olm.operatorframework.io/lifecycle-server=true')")

	return cmd
}
