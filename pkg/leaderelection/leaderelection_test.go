package leaderelection

import (
	"reflect"
	"testing"

	configv1 "github.com/openshift/api/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetLeaderElectionConfig(t *testing.T) {
	sch := runtime.NewScheme()
	configv1.AddToScheme(sch)
	testCases := []struct {
		desc         string
		enabled      bool
		clusterInfra configv1.Infrastructure
		expected     configv1.LeaderElection
	}{
		{
			desc:    "single node leader election values when ControlPlaneTopology is SingleReplicaTopologyMode",
			enabled: true,
			clusterInfra: configv1.Infrastructure{
				ObjectMeta: metav1.ObjectMeta{Name: infraResourceName},
				Status: configv1.InfrastructureStatus{
					ControlPlaneTopology: configv1.SingleReplicaTopologyMode,
				}},
			expected: configv1.LeaderElection{
				Disable: false,
				LeaseDuration: metav1.Duration{
					Duration: defaultSingleNodeLeaseDuration,
				},
				RenewDeadline: metav1.Duration{
					Duration: defaultSingleNodeRenewDeadline,
				},
				RetryPeriod: metav1.Duration{
					Duration: defaultSingleNodeRetryPeriod,
				},
			},
		},
		{
			desc:    "ha leader election values when ControlPlaneTopology is HighlyAvailableTopologyMode",
			enabled: true,
			clusterInfra: configv1.Infrastructure{
				ObjectMeta: metav1.ObjectMeta{Name: infraResourceName},
				Status: configv1.InfrastructureStatus{
					ControlPlaneTopology: configv1.HighlyAvailableTopologyMode,
				}},
			expected: configv1.LeaderElection{
				Disable: false,
				LeaseDuration: metav1.Duration{
					Duration: defaultLeaseDuration,
				},
				RenewDeadline: metav1.Duration{
					Duration: defaultRenewDeadline,
				},
				RetryPeriod: metav1.Duration{
					Duration: defaultRetryPeriod,
				},
			},
		},
		{
			desc:    "when disabled the default HA values should be returned",
			enabled: false,
			clusterInfra: configv1.Infrastructure{
				ObjectMeta: metav1.ObjectMeta{Name: infraResourceName},
				Status: configv1.InfrastructureStatus{
					ControlPlaneTopology: configv1.SingleReplicaTopologyMode,
				}},
			expected: configv1.LeaderElection{
				Disable: true,
				LeaseDuration: metav1.Duration{
					Duration: defaultLeaseDuration,
				},
				RenewDeadline: metav1.Duration{
					Duration: defaultRenewDeadline,
				},
				RetryPeriod: metav1.Duration{
					Duration: defaultRetryPeriod,
				},
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			client := fake.NewClientBuilder().
				WithRuntimeObjects(&tC.clusterInfra).WithScheme(sch).Build()

			setupLog := ctrl.Log.WithName("leaderelection_config_testing")

			result := getLeaderElectionConfig(setupLog, client, tC.enabled)
			if !reflect.DeepEqual(result, tC.expected) {
				t.Errorf("expected %+v but got %+v", tC.expected, result)
			}
		})
	}
}
