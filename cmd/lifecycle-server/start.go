package main

import (
	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "start",
		Short:        "Start the Lifecycle Server",
		SilenceUsage: true,
		RunE:         run,
	}

	cmd.Flags().String("fbc-path", defaultFBCPath, "path to FBC catalog data")
	cmd.Flags().String("listen", defaultListenAddr, "address to listen on for HTTP API")
	cmd.Flags().String("health", defaultHealthAddr, "address to listen on for health checks")

	return cmd
}
