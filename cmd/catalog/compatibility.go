package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/blang/semver/v4"
	configv1client "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/operators/openshift"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/registry/resolver"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/registry/resolver/solver"
)

func prohibitIncompatible(ctx context.Context, cli configv1client.ClusterVersionsGetter) resolver.ConstraintProviderFunc {
	var (
		once sync.Once
		cv   semver.Version
	)

	return func(entry *resolver.Operator) ([]solver.Constraint, error) {
		var err error
		once.Do(func() { // once.Do is thread-safe and blocks all invocations until the first invocation returns
			// Note: We lazy-load clusterVersion so invoking prohibitIncompatible doesn't require
			// a running OpenShift cluster; i.e. we don't want command options like --version failing
			// because we can't talk to a cluster.

			// TODO(njhale): inject signals context
			cv, err = clusterVersion(ctx, versionClient)

			// Drop Z, since we only care about compatibility with Y
			cv.Minor = 0
		})
		if err != nil {
			// We failed to lazy-load
			return nil, err
		}

		max, err := maxOpenShiftVersion(entry)
		if err != nil {
			// All parsing errors should prohibit the entry from being installed
			return []solver.Constraint{resolver.PrettyConstraint(
				solver.Prohibited(),
				fmt.Sprintf("invalid %q property: %s", openshift.MaxOpenShiftVersionProperty, err),
			)}, nil // Don't bubble up err -- this allows resolution to continue
		}

		if max == nil || max.GTE(cv) {
			// No max version declared, don't prohibit
			return nil, nil
		}

		return []solver.Constraint{resolver.PrettyConstraint(
			solver.Prohibited(),
			fmt.Sprintf("bundle incompatible with openshift cluster, %q < cluster version: (%s.%s < %s.%s)", openshift.MaxOpenShiftVersionProperty, max.Major, max.Minor, cv.Major, cv.Minor),
		)}, nil
	}
}

func maxOpenShiftVersion(entry *resolver.Operator) (*semver.Version, error) {
	// Get the max property -- if defined -- and check for duplicates
	var max *string
	for _, property := range entry.Properties() {
		if property.Type != openshift.MaxOpenShiftVersionProperty {
			continue
		}

		if max != nil {
			return nil, fmt.Errorf(`Defining more than one "%s" property is not allowed`, openshift.MaxOpenShiftVersionProperty)
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
		return nil, fmt.Errorf(`Value cannot be "" (an empty string)`)
	}

	version, err := semver.ParseTolerant(value)
	if err != nil {
		return nil, fmt.Errorf(`Failed to parse "%s" as semver: %w`, value, err)
	}

	truncatedVersion := semver.Version{Major: version.Major, Minor: version.Minor}
	if !version.EQ(truncatedVersion) {
		return nil, fmt.Errorf("property %s must specify only <major>.<minor> version, got invalid value %s", openshift.MaxOpenShiftVersionProperty, version)
	}
	return &truncatedVersion, nil
}

func clusterVersion(ctx context.Context, cli configv1client.ClusterVersionsGetter) (semver.Version, error) {
	var desired semver.Version
	cv, err := cli.ClusterVersions().Get(ctx, "version", metav1.GetOptions{})
	if err != nil { // "version" is the name of OpenShift's ClusterVersion singleton
		return desired, fmt.Errorf("Failed to get ClusterVersion: %w", err)
	}

	if cv == nil {
		// Note: A nil return without an error breaks the client's contract.
		// If this happens it's probably due to a client fake with ill-defined behavior.
		// TODO(njhale): Should this panic to indicate developer error?
		return desired, nil
	}

	v := cv.Status.Desired.Version
	if v == "" {
		// The release version hasn't been set yet
		return desired, fmt.Errorf("Desired release version missing from ClusterVersion")
	}

	desired, err = semver.ParseTolerant(v)
	if err != nil {
		return desired, fmt.Errorf("ClusterVersion has invalid desired release version: %w", err)
	}

	return desired, nil
}
