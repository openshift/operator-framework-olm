package validate

import (
	"fmt"
	"io/fs"

	"k8s.io/apimachinery/pkg/util/json"

	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"github.com/operator-framework/operator-registry/alpha/model"
	"github.com/operator-framework/operator-registry/pkg/api"
)

func Validate(root fs.FS) error {
	// Load config files and convert them to declcfg objects
	cfg, err := declcfg.LoadFS(root)
	if err != nil {
		return err
	}
	// Validate the config using model validation:
	// This will convert declcfg objects to intermediate model objects that are
	// also used for serve and add commands. The conversion process will run
	// validation for the model objects and ensure they are valid.
	mdl, err := declcfg.ConvertToModel(*cfg)
	if err != nil {
		return err
	}

	if err = validatePackageManifest(mdl); err != nil {
		return err
	}
	return nil
}

func validatePackageManifest(mdl model.Model) error {
	for _, pkg := range mdl {
		for _, channel := range pkg.Channels {
			head, err := channel.Head()
			if err != nil {
				return err
			}

			if len(head.CsvJSON) == 0 {
				return fmt.Errorf("missing head CSV on package %s, channel %s head %s: ensure valid csv under 'olm.bundle.object' properties", pkg.Name, channel.Name, head.Name)
			}
			bundle, err := api.ConvertModelBundleToAPIBundle(*head)
			if err != nil {
				return err
			}

			csv := operatorsv1alpha1.ClusterServiceVersion{}
			err = json.Unmarshal([]byte(bundle.GetCsvJson()), &csv)
			if err != nil {
				return fmt.Errorf("invalid head CSV on package %s, channel %s head %s: failed to unmarshal any 'olm.bundle.object' property as CSV JSON: %v", pkg.Name, channel.Name, head.Name, err)
			}
		}
	}
	return nil
}
