package controllers

import (
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/operator-framework-olm/pkg/manifests"
	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	logger    = zap.New()
	name      = "packageserver"
	namespace = "openshift-operator-lifecycle-manager"
	image     = getImageFromManifest()
)

func TestHighlyAvailableFromInstructure(t *testing.T) {
	const (
		singleReplicaHA  = false
		defaultReplicaHA = true
	)
	tt := []struct {
		name  string
		want  bool
		infra *configv1.Infrastructure
	}{
		{
			name: "SingleReplicaTopologyMode/non-HA",
			want: singleReplicaHA,
			infra: &configv1.Infrastructure{
				Status: configv1.InfrastructureStatus{
					ControlPlaneTopology: configv1.SingleReplicaTopologyMode,
				},
			},
		},
		{
			name: "HighlyAvailableTopologyMode/HA",
			want: defaultReplicaHA,
			infra: &configv1.Infrastructure{
				Status: configv1.InfrastructureStatus{
					ControlPlaneTopology: configv1.HighlyAvailableTopologyMode,
				},
			},
		},
		{
			name: "EmptyTopologyMode/HA",
			want: defaultReplicaHA,
			infra: &configv1.Infrastructure{
				Status: configv1.InfrastructureStatus{
					ControlPlaneTopology: "",
				},
			},
		},
	}

	for _, tc := range tt {
		got := getTopologyModeFromInfra(tc.infra)
		require.EqualValues(t, tc.want, got)
	}
}

func intOrStr(val int) *intstr.IntOrString {
	tmp := intstr.FromInt(val)
	return &tmp
}

func newTestCSV(replicas *int32, strategy *appsv1.RollingUpdateDeployment) *olmv1alpha1.ClusterServiceVersion {
	csv, err := manifests.NewPackageServerCSV(
		manifests.WithName(name),
		manifests.WithNamespace(namespace),
	)
	if err != nil {
		return nil
	}
	deployment := csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[0].Spec
	deployment.Replicas = replicas
	deployment.Strategy.RollingUpdate = strategy
	csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[0].Spec = deployment

	return csv
}

func getImageFromManifest() string {
	csv, err := manifests.NewPackageServerCSV(
		manifests.WithName(name),
		manifests.WithNamespace(namespace),
	)
	if err != nil {
		return ""
	}
	return csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[0].Spec.Template.Spec.Containers[0].Image
}

func TestEnsureCSV(t *testing.T) {
	defaultRollout := &appsv1.RollingUpdateDeployment{
		MaxUnavailable: intOrStr(defaultRolloutCount),
		MaxSurge:       intOrStr(defaultRolloutCount),
	}
	emptyRollout := &appsv1.RollingUpdateDeployment{}

	defaultReplicas := pointer.Int32(defaultReplicaCount)
	singleReplicas := pointer.Int32(singleReplicaCount)
	image := getImageFromManifest()

	tt := []struct {
		name            string
		inputCSV        *olmv1alpha1.ClusterServiceVersion
		expectedCSV     *olmv1alpha1.ClusterServiceVersion
		highlyAvailable bool
		want            bool
	}{
		{
			name:            "Modified/HighlyAvailable/CorrectReplicasIncorrectRolling",
			want:            true,
			highlyAvailable: true,
			inputCSV:        newTestCSV(defaultReplicas, emptyRollout),
			expectedCSV:     newTestCSV(defaultReplicas, defaultRollout),
		},
		{
			name:            "Modified/HighlyAvailable/IncorrectReplicasCorrectRolling",
			want:            true,
			highlyAvailable: true,
			inputCSV:        newTestCSV(singleReplicas, defaultRollout),
			expectedCSV:     newTestCSV(defaultReplicas, defaultRollout),
		},
		{
			name:            "NotModified/HighlyAvailable",
			want:            false,
			highlyAvailable: true,
			inputCSV:        newTestCSV(defaultReplicas, defaultRollout),
			expectedCSV:     newTestCSV(defaultReplicas, defaultRollout),
		},
		{
			name:            "Modified/SingleReplica/CorrectReplicasIncorrectRolling",
			want:            true,
			highlyAvailable: false,
			inputCSV:        newTestCSV(singleReplicas, defaultRollout),
			expectedCSV:     newTestCSV(singleReplicas, emptyRollout),
		},
		{
			name:            "Modified/SingleReplica/IncorrectReplicasCorrectRolling",
			want:            true,
			highlyAvailable: false,
			inputCSV:        newTestCSV(defaultReplicas, emptyRollout),
			expectedCSV:     newTestCSV(singleReplicas, emptyRollout),
		},
		{
			name:            "NotModified/SingleReplica",
			want:            false,
			highlyAvailable: false,
			inputCSV:        newTestCSV(singleReplicas, emptyRollout),
			expectedCSV:     newTestCSV(singleReplicas, emptyRollout),
		},
	}

	for _, tc := range tt {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			got := ensureCSV(logger, image, tc.inputCSV, tc.highlyAvailable)
			require.EqualValues(t, tc.want, got)
			require.EqualValues(t, tc.inputCSV.Spec, tc.expectedCSV.Spec)
		})
	}
}

func TestReplicaChanges(t *testing.T) {
	tt := []struct {
		name             string
		ha               bool
		expectedReplicas int32
	}{
		{
			name:             "HighlyAvailable/DefaultReplicas",
			ha:               true,
			expectedReplicas: defaultReplicaCount,
		},
		{
			name:             "NonHighlyAvailable/SingleReplica",
			ha:               false,
			expectedReplicas: singleReplicaCount,
		},
	}

	for _, tc := range tt {
		val := getReplicas(tc.ha)
		require.EqualValues(t, tc.expectedReplicas, val)
	}
}

func TestRolloutStrategy(t *testing.T) {
	tt := []struct {
		name            string
		ha              bool
		expectedRollout *appsv1.RollingUpdateDeployment
	}{
		{
			name: "HighlyAvailable/DefaultRollingUpdate",
			ha:   true,
			expectedRollout: &appsv1.RollingUpdateDeployment{
				MaxUnavailable: intOrStr(1),
				MaxSurge:       intOrStr(1),
			},
		},
		{
			name:            "NonHighlyAvailable/EmptyRollingUpdate",
			ha:              false,
			expectedRollout: &appsv1.RollingUpdateDeployment{},
		},
	}

	for _, tc := range tt {
		actual := getRolloutStrategy(tc.ha)
		require.EqualValues(t, actual, tc.expectedRollout)
	}
}
