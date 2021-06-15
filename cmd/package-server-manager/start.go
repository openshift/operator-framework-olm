package main

import (
	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the PackageServer manager",
		RunE:  run,
	}

	var (
		name      string
		namespace string
	)
	cmd.Flags().StringVar(&name, "name", defaultName, "configures the packageserver deployment metadata.name")
	cmd.Flags().StringVar(&namespace, "namespace", defaultNamespace, "configures the namespace for managing the packageserver deployment")

	return cmd
}
