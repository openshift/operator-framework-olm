package leaderelection

import (
	"context"
	"time"

	"github.com/go-logr/logr"

	configv1 "github.com/openshift/api/config/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	infraResourceName = "cluster"

	// Defaults follow conventions
	// https://github.com/openshift/enhancements/blob/master/CONVENTIONS.md#high-availability
	// Impl Calculations: https://github.com/openshift/library-go/commit/7e7d216ed91c3119800219c9194e5e57113d059a
	defaultLeaseDuration = 137 * time.Second
	defaultRenewDeadline = 107 * time.Second
	defaultRetryPeriod   = 26 * time.Second

	// Default leader election for SNO environments
	// Impl Calculations:
	// https://github.com/openshift/library-go/commit/2612981f3019479805ac8448b997266fc07a236a#diff-61dd95c7fd45fa18038e825205fbfab8a803f1970068157608b6b1e9e6c27248R127
	defaultSingleNodeLeaseDuration = 270 * time.Second
	defaultSingleNodeRenewDeadline = 240 * time.Second
	defaultSingleNodeRetryPeriod   = 60 * time.Second

)

var (
	// Retry configuration for fetching infrastructure status
	// These are variables to allow overriding in tests
	infraStatusRetryInterval = 2 * time.Second
	infraStatusRetryTimeout  = 30 * time.Second
)

var (
	defaultLeaderElectionConfig = configv1.LeaderElection{
		LeaseDuration: metav1.Duration{Duration: defaultLeaseDuration},
		RenewDeadline: metav1.Duration{Duration: defaultRenewDeadline},
		RetryPeriod:   metav1.Duration{Duration: defaultRetryPeriod},
	}
)

func GetLeaderElectionConfig(log logr.Logger, restConfig *rest.Config, enabled bool) (defaultConfig configv1.LeaderElection) {
	client, err := client.New(restConfig, client.Options{})
	if err != nil {
		log.Error(err, "unable to create client to detect SNO cluster, defaulting to HA leader election values")
		return defaultLeaderElectionConfig
	}
	configv1.AddToScheme(client.Scheme())
	return getLeaderElectionConfig(log, client, enabled)
}

func getLeaderElectionConfig(log logr.Logger, client client.Client, enabled bool) (config configv1.LeaderElection) {
	config = defaultLeaderElectionConfig
	config.Disable = !enabled
	if enabled {
		var infra *configv1.InfrastructureStatus
		var lastErr error

		ctx, cancel := context.WithTimeout(context.Background(), infraStatusRetryTimeout)
		defer cancel()

		err := wait.PollUntilContextTimeout(ctx, infraStatusRetryInterval, infraStatusRetryTimeout, true, func(ctx context.Context) (bool, error) {
			var err error
			infra, err = getClusterInfraStatus(ctx, client)
			if err != nil {
				lastErr = err
				log.Info("retrying to detect SNO cluster", "error", err)
				return false, nil // retry
			}
			return true, nil // success
		})

		if err != nil {
			log.Error(lastErr, "unable to detect SNO cluster, defaulting to HA leader election values")
			return
		}
		if infra != nil && infra.ControlPlaneTopology == configv1.SingleReplicaTopologyMode {
			return leaderElectionSNOConfig(config)
		}
	}
	return
}

func leaderElectionSNOConfig(config configv1.LeaderElection) configv1.LeaderElection {
	ret := *(&config).DeepCopy()
	ret.LeaseDuration.Duration = defaultSingleNodeLeaseDuration
	ret.RenewDeadline.Duration = defaultSingleNodeRenewDeadline
	ret.RetryPeriod.Duration = defaultSingleNodeRetryPeriod
	return ret
}

// Retrieve the cluster status, used to determine if we should use different leader election.
func getClusterInfraStatus(ctx context.Context, client client.Client) (*configv1.InfrastructureStatus, error) {
	infra := &configv1.Infrastructure{}
	err := client.Get(ctx, types.NamespacedName{Name: infraResourceName}, infra)
	if err != nil {
		return nil, err
	}
	return &infra.Status, nil
}
