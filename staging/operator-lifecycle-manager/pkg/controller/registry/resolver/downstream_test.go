package resolver

import (
	"testing"

	"github.com/blang/semver/v4"
	configfake "github.com/openshift/client-go/config/clientset/versioned/fake"
	configv1client "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	"github.com/stretchr/testify/require"
)

func TestClusterVersion(t *testing.T) {
	type expected struct {
		version semver.Version
		err     bool
	}

	for _, tt := range []struct {
		name     string
		cli      configv1client.ClusterVersionsGetter
		ver      *semver.Version
		expected expected
	}{
		{
			name: "NoVersion",
			cli:  configfake.NewSimpleClientset().ConfigV1(),
			expected: expected{
				err: true,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c := cluster{
				cli: tt.cli,
				ver: tt.ver,
			}

			version, err := c.version()
			require.Equal(t, tt.expected.version, version)
			if tt.expected.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

		})
	}

}
