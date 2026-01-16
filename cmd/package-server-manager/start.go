package main

import (
	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "start",
		Short:        "Start the PackageServer manager",
		SilenceUsage: true,
		RunE:         run,
	}

	cmd.Flags().String("name", defaultName, "configures the metadata.name for the packageserver csv resource")
	cmd.Flags().String("namespace", defaultNamespace, "configures the metadata.namespace that contains the packageserver csv resource")
	cmd.Flags().String("health", defaultHealthCheckPort, "configures the health check port that the kubelet is configured to probe")
	cmd.Flags().String("pprof", defaultPprofPort, "configures the pprof port that the process exposes")
	cmd.Flags().String("interval", defaultInterval, "configures the wakeup interval for the packageserver csc resource")
	cmd.Flags().String("metrics", defaultMetricsPort, "configures the metrics port that the process exposes")
	cmd.Flags().Bool("disable-leader-election", false, "configures whether leader election will be disabled")
	cmd.Flags().String("tls-cert", "", "path to use for certificate key (requires tls-key)")
	cmd.Flags().String("tls-key", "", "path to use for private key (requires tls-cert)")
	cmd.Flags().String("client-ca", "", "path to watch for client ca bundle")

	return cmd
}
