// Package filters provides utilities for generating test suite qualifiers
// used in the OpenShift OLMv1 test extension framework.
package filters

import (
	"fmt"
	"strings"
)

// and combines multiple filters using logical AND operator.
// Returns a parenthesized expression joining all filters with " && ".
func and(filters ...string) string {
	return fmt.Sprintf("(%s)", strings.Join(filters, " && "))
}

// buildFilter combines a base filter with an optional additional filter.
// If additionalFilter is empty, returns only the base filter wrapped in parentheses.
// Otherwise, combines both filters using logical AND.
func buildFilter(baseFilter, additionalFilter string) string {
	if additionalFilter == "" {
		return fmt.Sprintf("(%s)", baseFilter)
	}
	return and(fmt.Sprintf("(%s)", baseFilter), fmt.Sprintf("(%s)", additionalFilter))
}

// BasedStandardTests generates a qualifier for standard tests.
// Includes: non-Extended tests OR Extended tests marked as ReleaseGate.
// Additional filter can be applied to further narrow the selection.
func BasedStandardTests(filter string) string {
	standardFilter := `(!labels.exists(l, l=="Extended")) || (labels.exists(l, l=="Extended") && labels.exists(l, l=="ReleaseGate"))`
	return buildFilter(standardFilter, filter)
}

// BasedExtendedTests generates a qualifier for all extended tests.
// Includes: all tests marked with "Extended" label.
// Additional filter can be applied to further narrow the selection.
func BasedExtendedTests(filter string) string {
	extendedFilter := `labels.exists(l, l=="Extended")`
	return buildFilter(extendedFilter, filter)
}

// BasedExtendedReleaseGateTests generates a qualifier for extended release gate tests.
// Includes: Extended tests that are also marked as ReleaseGate.
// Additional filter can be applied to further narrow the selection.
func BasedExtendedReleaseGateTests(filter string) string {
	extendedReleaseGateFilter := `labels.exists(l, l=="Extended") && labels.exists(l, l=="ReleaseGate")`
	return buildFilter(extendedReleaseGateFilter, filter)
}

// BasedExtendedCandidateTests generates a qualifier for extended candidate tests.
// Includes: Extended tests that are NOT marked as ReleaseGate.
// Additional filter can be applied to further narrow the selection.
func BasedExtendedCandidateTests(filter string) string {
	extendedCandidateFilter := `labels.exists(l, l=="Extended") && !labels.exists(l, l=="ReleaseGate")`
	return buildFilter(extendedCandidateFilter, filter)
}

// BasedExtendedCandidateFuncTests generates a qualifier for extended candidate functional tests.
// Includes: Extended tests that are NOT ReleaseGate and NOT StressTest.
// Additional filter can be applied to further narrow the selection.
func BasedExtendedCandidateFuncTests(filter string) string {
	extendedCandidateFuncFilter := `labels.exists(l, l=="Extended") && !labels.exists(l, l=="ReleaseGate") && !labels.exists(l, l=="StressTest")`
	return buildFilter(extendedCandidateFuncFilter, filter)
}
