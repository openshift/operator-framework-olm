package controllers

import (
	"reflect"

	"github.com/go-logr/logr"
	configv1 "github.com/openshift/api/config/v1"
	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
)

func getReplicas(ha bool) int32 {
	if !ha {
		return singleReplicaCount
	}
	return defaultReplicaCount
}

func getRolloutStrategy(ha bool) *appsv1.RollingUpdateDeployment {
	if !ha {
		return &appsv1.RollingUpdateDeployment{}
	}

	intStr := intstr.FromInt(defaultRolloutCount)
	return &appsv1.RollingUpdateDeployment{
		MaxUnavailable: &intStr,
		MaxSurge:       &intStr,
	}
}

func getTopologyModeFromInfra(infra *configv1.Infrastructure) bool {
	var highAvailabilityMode bool
	if infra.Status.ControlPlaneTopology != configv1.SingleReplicaTopologyMode {
		highAvailabilityMode = true
	}
	return highAvailabilityMode
}

// ensureCSV is responsible for ensuring the state of the @csv ClusterServiceVersion custom
// resource matches the expected state based on any high availability expectations being exposed.
func ensureCSV(log logr.Logger, image string, csv *olmv1alpha1.ClusterServiceVersion, highlyAvailableMode bool) bool {
	var modified bool

	deploymentSpecs := csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs
	deployment := &deploymentSpecs[0].Spec

	currentImage := deployment.Template.Spec.Containers[0].Image
	if currentImage != image {
		log.Info("updating the image", "old", currentImage, "new", image)
		deployment.Template.Spec.Containers[0].Image = image
		modified = true
	}

	expectedReplicas := getReplicas(highlyAvailableMode)
	if *deployment.Replicas != expectedReplicas {
		log.Info("updating the replica count", "old", deployment.Replicas, "new", expectedReplicas)
		deployment.Replicas = pointer.Int32Ptr(expectedReplicas)
		modified = true
	}

	expectedRolloutConfiguration := getRolloutStrategy(highlyAvailableMode)
	if !reflect.DeepEqual(deployment.Strategy.RollingUpdate, expectedRolloutConfiguration) {
		log.Info("updating the rollout strategy")
		deployment.Strategy.RollingUpdate = expectedRolloutConfiguration
		modified = true
	}

	if modified {
		log.V(3).Info("csv has been modified")
		csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[0].Spec = *deployment
	}

	return modified
}

func validateCSV(log logr.Logger, csv *olmv1alpha1.ClusterServiceVersion) bool {
	deploymentSpecs := csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs
	if len(deploymentSpecs) != 1 {
		log.Info("csv contains more than one or zero nested deployment specs")
		return false
	}

	deployment := &deploymentSpecs[0].Spec
	if len(deployment.Template.Spec.Containers) != 1 {
		log.Info("csv contains more than one containers")
		return false
	}

	return true
}
