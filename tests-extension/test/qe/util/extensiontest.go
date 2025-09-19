package util

import (
	"strings"

	et "github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
)

// Olmv1QeTestsOnly returns a SelectFunction that filters tests to include only
// Extended-authored tests located in the specs directory.
//
// This function is used to identify tests that were migrated from the
// tests-private repository, as opposed to tests migrated from the dev origin
// repository. Tests in the "/qe/specs/" directory are automatically
// tagged with the "Extended" label by the test framework.
//
// Returns:
//   - et.SelectFunction: A function that takes an ExtensionTestSpec and returns
//     true if the test is located in the specs directory, false otherwise.
//
// Usage:
//
//	This function is typically used in test suite definitions to create
//	extended-specific test suites or to filter tests based on their origin.
func Olmv1QeTestsOnly() et.SelectFunction {
	return func(spec *et.ExtensionTestSpec) bool {
		// Iterate through all code locations associated with this test spec
		for _, cl := range spec.CodeLocations {
			// Check if any code location contains the extended test directory path
			if strings.Contains(cl, "/qe/specs/") {
				return true
			}
		}
		// Return false if no extended test directory paths are found
		return false
	}
}
