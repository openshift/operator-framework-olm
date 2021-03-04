package alpha

import (
	"github.com/openshift/operator-framework-olm/staging/operator-registry/cmd/opm/alpha/bundle"
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Hidden: true,
		Use:    "alpha",
		Short:  "Run an alpha subcommand",
	}

	runCmd.AddCommand(bundle.NewCmd())
	return runCmd
}
