package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "lifecycle-controller",
		Short: "Lifecycle Metadata Controller for OLM",
	}

	rootCmd.AddCommand(newStartCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error running lifecycle-controller: %v\n", err)
		os.Exit(1)
	}
}
