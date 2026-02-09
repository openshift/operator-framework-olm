package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "lifecycle-server",
		Short: "Lifecycle Metadata Server for OLM",
	}

	rootCmd.AddCommand(newStartCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error running lifecycle-server: %v\n", err)
		os.Exit(1)
	}
}
