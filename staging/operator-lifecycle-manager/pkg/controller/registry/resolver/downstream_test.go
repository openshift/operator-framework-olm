package resolver

import (
	"fmt"
	"testing"

	"github.com/blang/semver/v4"
	configfake "github.com/openshift/client-go/config/clientset/versioned/fake"
	configv1client "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	"github.com/stretchr/testify/require"

	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/operators/openshift"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/controller/registry/resolver/solver"
	"github.com/operator-framework/operator-registry/pkg/api"
)

func TestClusterConstraints(t *testing.T) {
	type expected struct {
		constraints []solver.Constraint
		err         bool
	}

	for _, tt := range []struct {
		name     string
		entry    *Operator
		expected expected
	}{
		{
			name:  "foo",
			entry: &Operator{},
			expected: expected{
				constraints: []solver.Constraint{},
				err:         false,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {

		})
	}

}

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

func TestMaxOpenShiftVersion(t *testing.T) {
	mustParse := func(s string) *semver.Version {
		version, err := semver.ParseTolerant(s)
		if err != nil {
			panic(fmt.Sprintf("bad version given for test case: %s", err))
		}
		return &version
	}

	type expect struct {
		err bool
		max *semver.Version
	}
	for _, tt := range []struct {
		description string
		in          []string
		expect      expect
	}{
		{
			description: "None",
			expect: expect{
				err: false,
				max: nil,
			},
		},
		{
			description: "Nothing",
			in:          []string{`""`},
			expect: expect{
				err: true,
				max: nil,
			},
		},
		{
			description: "Garbage",
			in:          []string{`"bad_version"`},
			expect: expect{
				err: true,
				max: nil,
			},
		},
		{
			description: "Single",
			in:          []string{`"1.0.0"`},
			expect: expect{
				err: false,
				max: mustParse("1.0.0"),
			},
		},
		{
			description: "Multiple",
			in: []string{
				`"1.0.0"`,
				`"2.0.0"`,
			},
			expect: expect{
				err: true,
				max: nil,
			},
		},
		{
			// Ensure unquoted short strings are accepted; e.g. X.Y
			description: "Unquoted/Short",
			in:          []string{"4.8"},
			expect: expect{
				err: false,
				max: mustParse("4.8"),
			},
		},
	} {
		t.Run(tt.description, func(t *testing.T) {
			var properties []*api.Property
			for _, max := range tt.in {
				properties = append(properties, &api.Property{
					Type:  openshift.MaxOpenShiftVersionProperty,
					Value: max,
				})
			}

			entry := &Operator{
				properties: properties,
			}
			max, err := maxOpenShiftVersion(entry)
			if tt.expect.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tt.expect.max, max)
		})
	}
}
