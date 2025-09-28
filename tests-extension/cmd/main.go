/*
This command is used to run the OLMv0 tests extension for OpenShift.
It registers the OLMv0 tests with the OpenShift Tests Extension framework
and provides a command-line interface to execute them.

For further information, please refer to the documentation at:
https://github.com/openshift-eng/openshift-tests-extension/blob/main/cmd/example-tests/main.go
*/
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/openshift-eng/openshift-tests-extension/pkg/cmd"
	e "github.com/openshift-eng/openshift-tests-extension/pkg/extension"
	et "github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
	g "github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"
	"github.com/spf13/cobra"

	_ "github.com/openshift/operator-framework-olm/tests-extension/test/qe/specs"
	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util/filters"
)

func main() {
	registry := e.NewRegistry()
	ext := e.NewExtension("openshift", "payload", "olmv0")

	// Register the OLMv0 test suites with OpenShift Tests Extension.
	// These suites determine how test cases are grouped and executed in various jobs.
	//
	// Definitions of labels:
	// - [Serial]: test must run in isolation, one at a time. Typically used for disruptive cases (e.g., kill nodes).
	// - [Slow]: test takes a long time to execute (i.e. >5 min.). Cannot be included in fast/parallel suites.
	//
	// IMPORTANT:
	// Even though a suite is marked "parallel", all tests run serially when using the *external binary*
	// (`run-suite`, `run-test`) because it executes within a single process and Ginkgo
	// cannot parallelize within a single process.
	// See: https://github.com/openshift-eng/openshift-tests-extension/blob/main/pkg/ginkgo/util.go#L50
	//
	// For actual parallel test execution (e.g., in CI), use `openshift-tests`, which launches one process per test:
	// https://github.com/openshift/origin/blob/main/pkg/test/ginkgo/test_runner.go#L294

	// Suite: olmv0/parallel
	// ---------------------------------------------------------------
	// Contains fast, parallel-safe test cases only.
	// Excludes any tests labeled [Serial] or [Slow].
	// Note: Tests with [Serial] and [Slow] cannot run with openshift/conformance/parallel.
	// Note: It includes the extended cases which match the openshift-tests.
	ext.AddSuite(e.Suite{
		Name:    "olmv0/parallel",
		Parents: []string{"openshift/conformance/parallel"},
		Qualifiers: []string{
			filters.BasedStandardTests(`!(name.contains("[Serial]") || name.contains("[Slow]"))`),
		},
	})

	// Suite: olmv0/serial
	// ---------------------------------------------------------------
	// Contains tests explicitly labeled [Serial].
	// These tests are typically disruptive and must run one at a time.
	// Note: It includes the extended cases which match the openshift-tests.
	ext.AddSuite(e.Suite{
		Name:    "olmv0/serial",
		Parents: []string{"openshift/conformance/serial"},
		Qualifiers: []string{
			filters.BasedStandardTests(`(name.contains("[Serial]") && !name.contains("[Disruptive]") && !name.contains("[Slow]"))`),
			// refer to https://github.com/openshift/origin/blob/main/pkg/testsuites/standard_suites.go#L456
		},
	})

	// Suite: olmv0/slow
	// 	// ---------------------------------------------------------------
	// Contains tests labeled [Slow], which take significant time to run.
	// These are not allowed in fast/parallel suites, and should run in optional/slow jobs.
	// Note: It includes the extended cases which match the openshift-tests.
	ext.AddSuite(e.Suite{
		Name:    "olmv0/slow",
		Parents: []string{"openshift/optional/slow"},
		Qualifiers: []string{
			filters.BasedStandardTests(`name.contains("[Slow]")`),
		},
	})

	// Suite: olmv0/all
	// ---------------------------------------------------------------
	// All tests in one suite: includes [Serial], [Slow], [Disruptive], etc.
	// Note: It includes the extended cases which match the openshift-tests.
	ext.AddSuite(e.Suite{
		Name: "olmv0/all",
		Qualifiers: []string{
			filters.BasedStandardTests(``),
		},
	})

	// Extended Suite: All extended tests
	// Contains all extended tests migrated from tests-private repository
	ext.AddSuite(e.Suite{
		Name: "olmv0/extended",
		Qualifiers: []string{
			filters.BasedExtendedTests(``),
		},
	})

	// Extended ReleaseGate Suite: extended tests that meet OpenShift CI requirements
	// Contains extended tests marked as release gate for openshift-tests
	ext.AddSuite(e.Suite{
		Name: "olmv0/extended/releasegate",
		Qualifiers: []string{
			filters.BasedExtendedReleaseGateTests(``),
		},
	})

	// Extended Candidate Suite: Extended tests that don't meet OpenShift CI requirements
	// Contains extended tests that are not for openshift-tests (run in custom periodic jobs only)
	ext.AddSuite(e.Suite{
		Name: "olmv0/extended/candidate",
		Qualifiers: []string{
			filters.BasedExtendedCandidateTests(``),
		},
	})

	//
	// Categorization of Extended Candidate Tests:
	// ===========================================
	// The extended/candidate tests are categorized by test purpose and characteristics:
	//
	// 1. By Test Type:
	//    - function: Functional tests that verify feature behavior and business logic
	//    - stress:   Stress tests that verify system behavior under resource pressure and load
	//
	// Relationship: candidate = function + stress + (other specialized test types)

	// Extended Candidate Function Suite: Extended functional tests that don't meet OpenShift CI requirements
	// Contains extended tests that are not for openshift-tests and exclude stress tests
	ext.AddSuite(e.Suite{
		Name: "olmv0/extended/candidate/function",
		Qualifiers: []string{
			filters.BasedExtendedCandidateFuncTests(``),
		},
	})

	//
	// Categorization of Extended Candidate Functional Tests:
	// =====================================================
	// The extended/candidate/function tests are categorized using two complementary approaches:
	//
	// 1. By Execution Model:
	//    - parallel: Tests that can run concurrently (excludes [Serial] and [Slow])
	//    - serial:   Tests that must run one at a time ([Serial] but not [Slow])
	//    - slow:     Tests that take significant time to execute ([Slow])
	//
	// 2. By Execution Speed:
	//    - fast:     All non-slow functional tests (includes both parallel and serial, excludes [Slow])
	//    - slow:     Tests marked as [Slow] (same as above)
	//
	// Relationship: function = parallel + serial + slow = fast + slow

	// Extended Candidate Suite Parallel Suite: extended tests that can run in parallel
	// Contains extended tests that can run concurrently (excludes Serial, Slow, and StressTest)
	ext.AddSuite(e.Suite{
		Name: "olmv0/extended/candidate/parallel",
		Qualifiers: []string{
			filters.BasedExtendedCandidateFuncTests(`!(name.contains("[Serial]") || name.contains("[Slow]"))`),
		},
	})
	// Extended Candidate Serial Suite: extended tests that must run one at a time
	// Contains extended tests marked as [Serial] (includes Disruptive tests since not used for openshift-tests)
	ext.AddSuite(e.Suite{
		Name: "olmv0/extended/candidate/serial",
		Qualifiers: []string{
			filters.BasedExtendedCandidateFuncTests(`(name.contains("[Serial]") && !name.contains("[Slow]"))`),
			// it is not used for openshift-tests, so it does not exclude Disruptive, so that we could use
			// olmv0/extended/candidate/serial to run all serial case including Disruptive cases
		},
	})

	// Extended Candidate Fast Suite: extended functional tests excluding slow cases
	// Contains all extended functional tests that are not marked as [Slow] (includes both Serial and Parallel)
	// This provides a comprehensive functional test coverage with reasonable execution time
	ext.AddSuite(e.Suite{
		Name: "olmv0/extended/candidate/fast",
		Qualifiers: []string{
			filters.BasedExtendedCandidateFuncTests(`!name.contains("[Slow]")`),
		},
	})
	// Extended Candidate Slow Suite: extended tests that take significant time to run
	// Contains extended tests marked as [Slow] (long-running tests not suitable for fast CI)
	ext.AddSuite(e.Suite{
		Name: "olmv0/extended/candidate/slow",
		Qualifiers: []string{
			filters.BasedExtendedCandidateFuncTests(`name.contains("[Slow]")`),
		},
	})
	// Extended Candidate Stress Suite: extended stress tests
	// Contains extended tests designed for stress testing and resource exhaustion scenarios
	ext.AddSuite(e.Suite{
		Name: "olmv0/extended/candidate/stress",
		Qualifiers: []string{
			filters.BasedExtendedCandidateTests(`labels.exists(l, l=="StressTest")`),
		},
	})

	specs, err := g.BuildExtensionTestSpecsFromOpenShiftGinkgoSuite()
	if err != nil {
		panic(fmt.Sprintf("couldn't build extension test specs from ginkgo: %+v", err.Error()))
	}

	// Automatically identify and label extended tests: select all tests from the specs directory and add the "Extended" label
	specs.Select(exutil.Olmv1QeTestsOnly()).AddLabel("Extended")
	// Process all test specs to apply extended-specific transformations and topology exclusions
	specs = specs.Walk(func(spec *et.ExtensionTestSpec) {
		if spec.Labels.Has("Extended") {
			// Change blocking tests to informing unless marked as ReleaseGate
			if !spec.Labels.Has("ReleaseGate") && spec.Lifecycle == "blocking" {
				spec.Lifecycle = "informing"
			}
			// Exclude External topology for NonHyperShiftHOST tests
			if spec.Labels.Has("NonHyperShiftHOST") {
				spec.Exclude(et.TopologyEquals("External"))
			}
			// Include External Connecttivity for Disconnected only tests
			if strings.Contains(spec.Name, "[Skipped:Connected]") {
				spec.Include(et.ExternalConnectivityEquals("Disconnected"))
			}
		}
	})

	// Ensure `[Disruptive]` tests are always also marked `[Serial]`.
	// This prevents them from running in parallel suites, which could cause flaky failures
	// due to disruptive behavior.
	specs = specs.Walk(func(spec *et.ExtensionTestSpec) {
		if strings.Contains(spec.Name, "[Disruptive]") && !strings.Contains(spec.Name, "[Serial]") {
			spec.Name = strings.ReplaceAll(
				spec.Name,
				"[Disruptive]",
				"[Serial][Disruptive]",
			)
		}
	})

	// To handle renames and preserve test ID by setting the original-name.
	// This logic looks for a custom Ginkgo label in the format:
	//   Label("original-name:<full old test name>")
	// When found, it sets spec.OriginalName = <old name>.
	// **Example**
	// It("should pass a renamed sanity check",
	//		Label("original-name:[sig-operator] OLMv0 should pass a trivial sanity check"),
	//  	func(ctx context.Context) {
	//  		Expect(len("test")).To(BeNumerically(">", 0))
	// 	    })
	specs = specs.Walk(func(spec *et.ExtensionTestSpec) {
		for label := range spec.Labels {
			if strings.HasPrefix(label, "original-name:") {
				parts := strings.SplitN(label, "original-name:", 2)
				if len(parts) > 1 {
					spec.OriginalName = parts[1]
				}
			}
		}
	})

	// To delete tests you must mark them as obsolete.
	// These tests will be excluded from metadata validation during OTE update.
	// 1 - To get the full name of the test you want to remove run: make list-test-names
	// 2 - Add the test name here to avoid validation errors
	// 3 - Remove the test in your test file.
	// 4 - Run make build-update
	ext.IgnoreObsoleteTests(
	// "[sig-operator] OLMv0 should pass a trivial sanity check",
	// Add more removed test names below
	)

	// Initialize the environment before running any tests.
	specs.AddBeforeAll(func() {
		exutil.InitClusterEnv()
	})

	ext.AddSpecs(specs)
	registry.Register(ext)

	root := &cobra.Command{
		Long: "OLMv0 Tests Extension",
	}

	root.AddCommand(cmd.DefaultExtensionCommands(registry)...)

	if err := func() error {
		return root.Execute()
	}(); err != nil {
		os.Exit(1)
	}
}
