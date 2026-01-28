package leaderelection

import (
	"context"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	configv1 "github.com/openshift/api/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

func TestGetLeaderElectionConfigWithRetry(t *testing.T) {
	// Save original values and restore after test
	originalInterval := infraStatusRetryInterval
	originalTimeout := infraStatusRetryTimeout
	defer func() {
		infraStatusRetryInterval = originalInterval
		infraStatusRetryTimeout = originalTimeout
	}()

	// Use shorter timeouts for testing
	infraStatusRetryInterval = 10 * time.Millisecond
	infraStatusRetryTimeout = 100 * time.Millisecond

	sch := runtime.NewScheme()
	configv1.AddToScheme(sch)

	t.Run("falls back to HA values when infrastructure not found", func(t *testing.T) {
		// Create client without infrastructure object
		c := fake.NewClientBuilder().WithScheme(sch).Build()
		setupLog := ctrl.Log.WithName("leaderelection_config_testing")

		result := getLeaderElectionConfig(setupLog, c, true)

		expected := configv1.LeaderElection{
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
		}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("expected %+v but got %+v", expected, result)
		}
	})

	t.Run("succeeds after retry", func(t *testing.T) {
		infra := &configv1.Infrastructure{
			ObjectMeta: metav1.ObjectMeta{Name: infraResourceName},
			Status: configv1.InfrastructureStatus{
				ControlPlaneTopology: configv1.SingleReplicaTopologyMode,
			},
		}
		c := fake.NewClientBuilder().WithScheme(sch).WithRuntimeObjects(infra).Build()

		// Wrap client to fail first 2 attempts
		mockClient := &failingClient{
			Client:       c,
			failuresLeft: 2,
		}

		setupLog := ctrl.Log.WithName("leaderelection_config_testing")
		result := getLeaderElectionConfig(setupLog, mockClient, true)

		expected := configv1.LeaderElection{
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
		}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("expected %+v but got %+v", expected, result)
		}

		// Verify retries happened (2 failures + 1 success = 3 calls, so counter should be -1)
		if atomic.LoadInt32(&mockClient.failuresLeft) != -1 {
			t.Errorf("expected failuresLeft to be -1 after retries, but got %d", mockClient.failuresLeft)
		}
	})
}

// failingClient wraps a client and fails the first N Get calls
type failingClient struct {
	client.Client
	failuresLeft int32
}

func (f *failingClient) Get(ctx context.Context, key types.NamespacedName, obj client.Object, opts ...client.GetOption) error {
	if atomic.AddInt32(&f.failuresLeft, -1) >= 0 {
		return context.DeadlineExceeded
	}
	return f.Client.Get(ctx, key, obj, opts...)
}
