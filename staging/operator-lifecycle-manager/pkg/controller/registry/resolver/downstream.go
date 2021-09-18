package resolver

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/blang/semver/v4"
	configv1client "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/operators/openshift"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/registry/resolver/solver"
)

var openshiftCluster cluster

func downstreamConstraints(entry *Operator) ([]solver.Constraint, error) {
	clusterVersion, err := openshiftCluster.version()
	if err != nil {
		// We failed to lazy-load
		return nil, err
	}

	max, err := maxOpenShiftVersion(entry)
	if err != nil {
		// All parsing errors should prohibit the entry from being installed
		return []solver.Constraint{PrettyConstraint(
			solver.Prohibited(),
			fmt.Sprintf("invalid %q property: %s", openshift.MaxOpenShiftVersionProperty, err),
		)}, nil // Don't bubble up err -- this allows resolution to continue
	}

	if max == nil || max.GTE(clusterVersion) {
		// No max version declared, don't prohibit
		return nil, nil
	}

	return []solver.Constraint{PrettyConstraint(
		solver.Prohibited(),
		fmt.Sprintf("bundle incompatible with openshift cluster, %q < cluster version: (%d.%d < %d.%d)", openshift.MaxOpenShiftVersionProperty, max.Major, max.Minor, clusterVersion.Major, clusterVersion.Minor),
	)}, nil
}

func maxOpenShiftVersion(entry *Operator) (*semver.Version, error) {
	// Get the max property -- if defined -- and check for duplicates
	var max *string
	for _, property := range entry.Properties() {
		if property.Type != openshift.MaxOpenShiftVersionProperty {
			continue
		}

		if max != nil {
			return nil, fmt.Errorf("defining more than one %q property is not allowed", openshift.MaxOpenShiftVersionProperty)
		}

		max = &property.Value
	}

	if max == nil {
		return nil, nil
	}

	// Account for any additional quoting
	value := strings.Trim(*max, "\"")
	if value == "" {
		// Handle "" separately, so parse doesn't treat it as a zero
		return nil, fmt.Errorf(`value cannot be "" (an empty string)`)
	}

	version, err := semver.ParseTolerant(value)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %q as semver: %w", value, err)
	}

	truncatedVersion := semver.Version{Major: version.Major, Minor: version.Minor}
	if !version.EQ(truncatedVersion) {
		return nil, fmt.Errorf("property %q must specify only <major>.<minor> version, got invalid value %s", openshift.MaxOpenShiftVersionProperty, version)
	}
	return &truncatedVersion, nil
}

func desiredVersion(ctx context.Context, cli configv1client.ClusterVersionsGetter) (*semver.Version, error) {
	var desired semver.Version
	cv, err := cli.ClusterVersions().Get(ctx, "version", metav1.GetOptions{})
	if err != nil { // "version" is the name of OpenShift's ClusterVersion singleton
		return nil, fmt.Errorf("failed to get cluster version: %w", err)
	}

	if cv == nil {
		// Note: A nil return without an error breaks the client's contract.
		// If this happens it's probably due to a client fake with ill-defined behavior.
		// TODO(njhale): Should this panic to indicate developer error?
		return nil, fmt.Errorf("incorrect client behavior observed")
	}

	v := cv.Status.Desired.Version
	if v == "" {
		// The release version hasn't been set yet
		return nil, fmt.Errorf("desired release missing from resource")
	}

	desired, err = semver.ParseTolerant(v)
	if err != nil {
		return nil, fmt.Errorf("resource has invalid desired release: %w", err)
	}

	return &desired, nil
}

type cluster struct {
	mu  sync.Mutex
	ver *semver.Version
	cli configv1client.ClusterVersionsGetter
}

func (c *cluster) version() (semver.Version, error) {
	var (
		err error
		ver semver.Version
	)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ver == nil {
		// Note: We lazy-load cluster.ver so instantiating a cluster struct doesn't require
		// a running OpenShift cluster; i.e. we don't want command options like --version failing
		// because we can't talk to a cluster.
		c.ver, err = desiredVersion(context.TODO(), c.cli)
	}

	if c.ver != nil {
		ver = *c.ver
	}

	return ver, err
}
