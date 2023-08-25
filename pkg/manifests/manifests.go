package manifests

import (
	_ "embed"

	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"

	"k8s.io/apimachinery/pkg/util/yaml"
)

var (
	//go:embed csv.yaml
	packageServerCSV []byte
)

type CSVOption func(*olmv1alpha1.ClusterServiceVersion)

// NewPackageServerCSV is responsible for serializing the PackageServer csv.yaml
// YAML manifest into a populated ClusterServiceVersion Go structure that contains a
// metadata.namespace that matches the @namespace value.
func NewPackageServerCSV(opts ...CSVOption) (*olmv1alpha1.ClusterServiceVersion, error) {
	var csv olmv1alpha1.ClusterServiceVersion
	if err := yaml.Unmarshal(packageServerCSV, &csv); err != nil {
		return nil, err
	}

	for _, opt := range opts {
		opt(&csv)
	}

	return &csv, nil
}

func WithRunFlags(flags []string) CSVOption {
	return func(csv *olmv1alpha1.ClusterServiceVersion) {
		for i, deployment := range csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
			for j, container := range deployment.Spec.Template.Spec.Containers {
				// TODO: Should be fine to hardcode this for now, but likely want
				// to pass this as a parameter?
				if container.Name == "packageserver" {
					for _, flag := range flags {
						container.Command = append(container.Command, flag)
					}
					csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[i].Spec.Template.Spec.Containers[j].Command = container.Command
					break
				}
			}
		}

	}
}

func WithName(name string) CSVOption {
	return func(csv *olmv1alpha1.ClusterServiceVersion) {
		csv.Name = name
	}
}

func WithNamespace(namespace string) CSVOption {
	return func(csv *olmv1alpha1.ClusterServiceVersion) {
		csv.Namespace = namespace
	}
}

func WithImage(image string) CSVOption {
	return func(csv *olmv1alpha1.ClusterServiceVersion) {
		for _, deployment := range csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs {
			for _, container := range deployment.Spec.Template.Spec.Containers {
				// TODO(tflannag): Should be fine to hardcode this for now, but likely want
				// to pass this as a parameter?
				if container.Name == "packageserver" {
					container.Image = image
					break
				}
			}
		}
	}
}
