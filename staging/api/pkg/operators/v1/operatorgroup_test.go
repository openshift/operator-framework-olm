package v1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpgradeStrategy(t *testing.T) {
	tests := []struct {
		description string
		og          *OperatorGroup
		expected    upgradeStrategyName
	}{
		{
			description: "NoSpec",
			og:          &OperatorGroup{},
			expected:    DefaultUpgradeStrategy,
		},
		{
			description: "NoUpgradeStrategy",
			og: &OperatorGroup{
				Spec: OperatorGroupSpec{},
			},
			expected: DefaultUpgradeStrategy,
		},

		{
			description: "NoUpgradeStrategyName",
			og: &OperatorGroup{
				Spec: OperatorGroupSpec{
					UpgradeStrategy: UpgradeStrategy{},
				},
			},
			expected: DefaultUpgradeStrategy,
		},
		{
			description: "NonSupportedUpgradeStrategyName",
			og: &OperatorGroup{
				Spec: OperatorGroupSpec{
					UpgradeStrategy: UpgradeStrategy{
						Name: "",
					},
				},
			},
			expected: DefaultUpgradeStrategy,
		},
		{
			description: "UnsafeFailForwardUpgradeStrategyName",
			og: &OperatorGroup{
				Spec: OperatorGroupSpec{
					UpgradeStrategy: UpgradeStrategy{
						Name: "UnsafeFailForward",
					},
				},
			},
			expected: UnsafeFailForwardUpgradeStrategy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.description, func(t *testing.T) {
			require.EqualValues(t, tt.expected, tt.og.UpgradeStrategy())
		})
	}
}
