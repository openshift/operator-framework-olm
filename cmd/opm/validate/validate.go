package validate

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	dsvalidate "github.com/openshift/operator-framework-olm/pkg/validate"
	"github.com/operator-framework/operator-registry/cmd/opm/validate"
)

func NewCmd() *cobra.Command {
	logger := logrus.New()
	validateCmd := validate.NewCmd()
	validateFn := validateCmd.RunE
	validateCmd.RunE = func(c *cobra.Command, args []string) error {
		if err := validateFn(c, args); err != nil {
			logger.Fatal(err)
		}

		directory := args[0]
		s, err := os.Stat(directory)
		if err != nil {
			return err
		}
		if !s.IsDir() {
			return fmt.Errorf("%q is not a directory", directory)
		}

		if err := dsvalidate.Validate(os.DirFS(directory)); err != nil {
			logger.Fatal(err)
		}
		return nil
	}

	return validateCmd
}
