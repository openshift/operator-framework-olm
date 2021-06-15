package manifests

import (
	_ "embed"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

// Note: can also just create a helper function that inlines the default
// packageserver deployment as a Go structure w/o having to serialize
// that YAML manifest.

var (
	// TODO(tflannag): Worth building up a embed.FS to avoid hardlinking this
	// manifest here?
	//go:embed deployment.yaml
	packageServerDeployment []byte
)

func NewPackageServerDeployment(namespace string) (*appsv1.Deployment, error) {
	var deployment appsv1.Deployment
	if err := yaml.Unmarshal(packageServerDeployment, &deployment); err != nil {
		return nil, err
	}
	deployment.SetNamespace(namespace)

	return &deployment, nil
}
