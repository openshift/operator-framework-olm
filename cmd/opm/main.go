package main

import (
	"os"

	"github.com/spf13/cobra"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"

	"github.com/operator-framework/operator-registry/cmd/opm/root"
	registrylib "github.com/operator-framework/operator-registry/pkg/registry"

	"github.com/openshift/operator-framework-olm/cmd/opm/validate"
)

func main() {
	override := map[string]*cobra.Command{"validate <directory>": validate.NewCmd()}
	cmd := root.NewCmd()
	for _, c := range cmd.Commands() {
		if newCmd, ok := override[c.Use]; ok {
			cmd.RemoveCommand(c)
			cmd.AddCommand(newCmd)
		}
	}

	if err := cmd.Execute(); err != nil {
		agg, ok := err.(utilerrors.Aggregate)
		if !ok {
			os.Exit(1)
		}
		for _, e := range agg.Errors() {
			if _, ok := e.(registrylib.BundleImageAlreadyAddedErr); ok {
				os.Exit(2)
			}
			if _, ok := e.(registrylib.PackageVersionAlreadyAddedErr); ok {
				os.Exit(3)
			}
		}
		os.Exit(1)
	}
}
