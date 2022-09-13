package controllers

import (
	"reflect"

	"github.com/go-logr/logr"
	configv1 "github.com/openshift/api/config/v1"
	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	"github.com/openshift/operator-framework-olm/pkg/manifests"
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

func getAntiAffinityConfig(ha bool) *corev1.Affinity {
	if !ha {
		return &corev1.Affinity{}
	}
	return &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
				{
					TopologyKey: "kubernetes.io/hostname",
					LabelSelector: &metav1.LabelSelector{
						MatchLabels: map[string]string{
							"app": "packageserver",
						},
					},
				},
			},
		},
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
// resource matches that of the codified defaults and high availability configurations, where
// codified defaults are defined by the csv returned by the manifests.NewPackageServerCSV
// function.
func ensureCSV(log logr.Logger, image string, csv *olmv1alpha1.ClusterServiceVersion, highlyAvailableMode bool) (bool, error) {
	expectedCSV, err := manifests.NewPackageServerCSV(
		manifests.WithName(csv.Name),
		manifests.WithNamespace(csv.Namespace),
		manifests.WithImage(image),
	)
	if err != nil {
		return false, err
	}

	ensureCSVHighAvailability(image, expectedCSV, highlyAvailableMode)

	var modified bool

	for k, v := range expectedCSV.GetLabels() {
		if csv.GetLabels() == nil {
			csv.SetLabels(make(map[string]string))
		}
		if vv, ok := csv.GetLabels()[k]; !ok || vv != v {
			log.Info("setting expected label", "key", k, "value", v)
			csv.ObjectMeta.Labels[k] = v
			modified = true
		}
	}

	for k, v := range expectedCSV.GetAnnotations() {
		if csv.GetAnnotations() == nil {
			csv.SetAnnotations(make(map[string]string))
		}
		if vv, ok := csv.GetAnnotations()[k]; !ok || vv != v {
			log.Info("setting expected annotation", "key", k, "value", v)
			csv.ObjectMeta.Annotations[k] = v
			modified = true
		}
	}

	if !reflect.DeepEqual(expectedCSV.Spec, csv.Spec) {
		log.Info("updating csv spec")
		csv.Spec = expectedCSV.Spec
		modified = true
	}

	if modified {
		log.V(3).Info("csv has been modified")
	}

	return modified, err
}

// ensureCSVHighAvailability is responsible for ensuring the state of the @csv ClusterServiceVersion custom
// resource matches the expected state based on any high availability expectations being exposed.
func ensureCSVHighAvailability(image string, csv *olmv1alpha1.ClusterServiceVersion, highlyAvailableMode bool) {
	var modified bool

	deploymentSpecs := csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs
	deployment := &deploymentSpecs[0].Spec

	currentImage := deployment.Template.Spec.Containers[0].Image
	if currentImage != image {
		deployment.Template.Spec.Containers[0].Image = image
		modified = true
	}

	expectedReplicas := getReplicas(highlyAvailableMode)
	if *deployment.Replicas != expectedReplicas {
		deployment.Replicas = pointer.Int32Ptr(expectedReplicas)
		modified = true
	}

	expectedRolloutConfiguration := getRolloutStrategy(highlyAvailableMode)
	if !reflect.DeepEqual(deployment.Strategy.RollingUpdate, expectedRolloutConfiguration) {
		deployment.Strategy.RollingUpdate = expectedRolloutConfiguration
		modified = true
	}

	expectedAffinityConfiguration := getAntiAffinityConfig(highlyAvailableMode)
	if !reflect.DeepEqual(deployment.Template.Spec.Affinity, expectedAffinityConfiguration) {
		deployment.Template.Spec.Affinity = expectedAffinityConfiguration
		modified = true
	}

	if modified {
		csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[0].Spec = *deployment
	}
}
