package controllers

import (
	"testing"

	semver "github.com/blang/semver/v4"
	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/operator-framework-olm/pkg/manifests"
	"github.com/operator-framework/api/pkg/lib/version"
	olmv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type testCSVOption func(*olmv1alpha1.ClusterServiceVersion)

func withVersion(v semver.Version) func(*olmv1alpha1.ClusterServiceVersion) {
	return func(csv *olmv1alpha1.ClusterServiceVersion) {
		csv.Spec.Version = version.OperatorVersion{v}
	}
}

func withDescription(description string) func(*olmv1alpha1.ClusterServiceVersion) {
	return func(csv *olmv1alpha1.ClusterServiceVersion) {
		csv.Spec.Description = description
	}
}

func withAffinity(affinity *corev1.Affinity) func(*olmv1alpha1.ClusterServiceVersion) {
	return func(csv *olmv1alpha1.ClusterServiceVersion) {
		csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[0].Spec.Template.Spec.Affinity = affinity
	}
}
func withRollingUpdateStrategy(strategy *appsv1.RollingUpdateDeployment) func(*olmv1alpha1.ClusterServiceVersion) {
	return func(csv *olmv1alpha1.ClusterServiceVersion) {
		csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[0].Spec.Strategy.RollingUpdate = strategy
	}
}

func withReplicas(replicas *int32) func(*olmv1alpha1.ClusterServiceVersion) {
	return func(csv *olmv1alpha1.ClusterServiceVersion) {
		csv.Spec.InstallStrategy.StrategySpec.DeploymentSpecs[0].Spec.Replicas = replicas
	}
}

func newTestCSV(
	options ...testCSVOption,
) *olmv1alpha1.ClusterServiceVersion {
	csv, err := manifests.NewPackageServerCSV(
		manifests.WithName(name),
		manifests.WithNamespace(namespace),
	)
	if err != nil {
		return nil
	}

	for _, o := range options {
		o(csv)
	}

	return csv
}

func newPodAffinity(antiAffinity *corev1.PodAntiAffinity) *corev1.Affinity {
	return &corev1.Affinity{
		PodAntiAffinity: antiAffinity,
	}
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
	defaultAffinity := newPodAffinity(&corev1.PodAntiAffinity{
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
	})
	defaultRollout := &appsv1.RollingUpdateDeployment{
		MaxUnavailable: intOrStr(defaultRolloutCount),
		MaxSurge:       intOrStr(defaultRolloutCount),
	}
	emptyRollout := &appsv1.RollingUpdateDeployment{}

	defaultReplicas := pointer.Int32(defaultReplicaCount)
	singleReplicas := pointer.Int32(singleReplicaCount)
	image := getImageFromManifest()

	type wanted struct {
		expectedBool bool
		expectedErr  error
	}

	tt := []struct {
		name            string
		inputCSV        *olmv1alpha1.ClusterServiceVersion
		expectedCSV     *olmv1alpha1.ClusterServiceVersion
		highlyAvailable bool
		want            wanted
	}{
		{
			name:            "Modified/HighlyAvailable/CorrectReplicasIncorrectRolling",
			want:            wanted{true, nil},
			highlyAvailable: true,
			inputCSV:        newTestCSV(withReplicas(defaultReplicas), withRollingUpdateStrategy(emptyRollout), withAffinity(defaultAffinity)),
			expectedCSV:     newTestCSV(withReplicas(defaultReplicas), withRollingUpdateStrategy(defaultRollout), withAffinity(defaultAffinity)),
		},
		{
			name:            "Modified/HighlyAvailable/IncorrectReplicasCorrectRolling",
			want:            wanted{true, nil},
			highlyAvailable: true,
			inputCSV:        newTestCSV(withReplicas(singleReplicas), withRollingUpdateStrategy(defaultRollout), withAffinity(defaultAffinity)),
			expectedCSV:     newTestCSV(withReplicas(defaultReplicas), withRollingUpdateStrategy(defaultRollout), withAffinity(defaultAffinity)),
		},
		{
			name:            "Modified/HighlyAvailable/IncorrectPodAntiAffinity",
			want:            wanted{true, nil},
			highlyAvailable: true,
			inputCSV: newTestCSV(withReplicas(singleReplicas), withRollingUpdateStrategy(defaultRollout), withAffinity(newPodAffinity(&corev1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
					{
						Weight: 1,
					},
				},
			}))),
			expectedCSV: newTestCSV(withReplicas(defaultReplicas), withRollingUpdateStrategy(defaultRollout), withAffinity(defaultAffinity)),
		},
		{
			name:            "NotModified/HighlyAvailable",
			want:            wanted{false, nil},
			highlyAvailable: true,
			inputCSV:        newTestCSV(withReplicas(defaultReplicas), withRollingUpdateStrategy(defaultRollout), withAffinity(defaultAffinity)),
			expectedCSV:     newTestCSV(withReplicas(defaultReplicas), withRollingUpdateStrategy(defaultRollout), withAffinity(defaultAffinity)),
		},

		{
			name:            "Modified/SingleReplica/CorrectReplicasIncorrectRolling",
			want:            wanted{true, nil},
			highlyAvailable: false,
			inputCSV:        newTestCSV(withReplicas(singleReplicas), withRollingUpdateStrategy(defaultRollout), withAffinity(&corev1.Affinity{})),
			expectedCSV:     newTestCSV(withReplicas(singleReplicas), withRollingUpdateStrategy(emptyRollout), withAffinity(&corev1.Affinity{})),
		},
		{
			name:            "Modified/SingleReplica/IncorrectReplicasCorrectRolling",
			want:            wanted{true, nil},
			highlyAvailable: false,
			inputCSV:        newTestCSV(withReplicas(defaultReplicas), withRollingUpdateStrategy(emptyRollout), withAffinity(&corev1.Affinity{})),
			expectedCSV:     newTestCSV(withReplicas(singleReplicas), withRollingUpdateStrategy(emptyRollout), withAffinity(&corev1.Affinity{})),
		},
		{
			name:            "Modified/SingleReplica/IncorrectPodAntiAffinity",
			want:            wanted{true, nil},
			highlyAvailable: false,
			inputCSV:        newTestCSV(withReplicas(singleReplicas), withRollingUpdateStrategy(emptyRollout), withAffinity(defaultAffinity)),
			expectedCSV:     newTestCSV(withReplicas(singleReplicas), withRollingUpdateStrategy(emptyRollout), withAffinity(&corev1.Affinity{})),
		},
		{
			name:            "Modified/SingleReplica/IncorrectVersion",
			want:            wanted{true, nil},
			highlyAvailable: false,
			inputCSV:        newTestCSV(withReplicas(singleReplicas), withRollingUpdateStrategy(emptyRollout), withAffinity(&corev1.Affinity{}), withVersion(semver.Version{Major: 0, Minor: 0, Patch: 0})),
			expectedCSV:     newTestCSV(withReplicas(singleReplicas), withRollingUpdateStrategy(emptyRollout), withAffinity(&corev1.Affinity{})),
		},
		{
			name:            "Modified/SingleReplica/IncorrectDescription",
			want:            wanted{true, nil},
			highlyAvailable: false,
			inputCSV:        newTestCSV(withReplicas(singleReplicas), withRollingUpdateStrategy(emptyRollout), withAffinity(&corev1.Affinity{}), withDescription("foo")),
			expectedCSV:     newTestCSV(withReplicas(singleReplicas), withRollingUpdateStrategy(emptyRollout), withAffinity(&corev1.Affinity{})),
		},
		{
			name:            "NotModified/SingleReplica",
			want:            wanted{false, nil},
			highlyAvailable: false,
			inputCSV:        newTestCSV(withReplicas(singleReplicas), withRollingUpdateStrategy(emptyRollout), withAffinity(&corev1.Affinity{})),
			expectedCSV:     newTestCSV(withReplicas(singleReplicas), withRollingUpdateStrategy(emptyRollout), withAffinity(&corev1.Affinity{})),
		},
	}

	for _, tc := range tt {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			gotBool, gotErr := ensureCSV(logger, image, tc.inputCSV, tc.highlyAvailable)
			require.EqualValues(t, tc.want.expectedBool, gotBool)
			require.EqualValues(t, tc.want.expectedErr, gotErr)
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
