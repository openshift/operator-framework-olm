# AGENTS.md

This file provides AI agents with comprehensive context about the OLM v0 QE Test Extension project to enable effective test development, debugging, and maintenance.

## Scope and Working Directory

### Applicability
This AGENTS.md applies to the **OLM v0 QE Test Extension** project located at:
```
operator-framework-olm/tests-extension/
```

**IMPORTANT**: This file is specifically for the **test code** in the `tests-extension/` directory, not for the OLM product code in other directories.

### Required Working Directory
For this AGENTS.md to be effective, ensure your working directory is set to:
```bash
<repo-root>/operator-framework-olm/tests-extension/
```

Or any subdirectory within `tests-extension/`, such as:
- `tests-extension/test/qe/`
- `tests-extension/test/qe/specs/`
- `tests-extension/cmd/`

### Working Directory Verification for AI Agents

**Context Awareness**: This AGENTS.md may be loaded even when not actively working with test files (e.g., user briefly entered `tests-extension/` directory and left). Apply these guidelines intelligently based on the actual task.

#### When to Apply This AGENTS.md

**ONLY apply this AGENTS.md when the user is working with test files**, identified by:
- File paths containing `tests-extension/test/`
- File paths containing `tests-extension/cmd/`
- Tasks explicitly about "OLM v0 tests", "test extension", "olmv0-tests-ext"

**DO NOT apply this AGENTS.md when**:
- Working with files outside `tests-extension/` (e.g., product code)
- User is in a different part of the repository
- Even if this AGENTS.md was previously loaded

#### Directory Check (Only for Test File Operations)

When the user asks to work with test files (files under `tests-extension/`):

1. **Check current working directory**:
   ```bash
   pwd
   ```

2. **Verify directory alignment**:
   - Preferred: Current directory should be `tests-extension/` or subdirectory
   - This ensures AGENTS.md context is automatically available

3. **If working directory is not `tests-extension/` or subdirectory**:

   **Inform (don't block) the user**:
   ```
   💡 Note: Working Directory Suggestion

   You're working with files under tests-extension/, but your current
   directory is elsewhere. For better context and auto-completion:

   Consider running: cd tests-extension/

   I can still help you, but setting the working directory correctly
   ensures I have full access to the test documentation.

   Do you want to continue in the current directory, or should I wait
   for you to switch?
   ```

**Important**: This is a suggestion, not a blocker. If the user wants to proceed, assist them normally.

### Path Structure Reference
```
operator-framework-olm/                          ← Parent repo (product code)
└── tests-extension/                             ← THIS AGENTS.MD APPLIES HERE
    ├── AGENTS.md                                ← This file
    ├── cmd/main.go                              ← Test binary entry point
    ├── test/qe/                                 ← Test code
    │   ├── specs/                               ← Test specifications
    │   └── util/                                ← Test utilities
    ├── Makefile                                 ← Build automation
    └── bin/olmv0-tests-ext                      ← Compiled test binary
```

## Project Overview

This is a **Quality Engineering (QE) test extension** for OLM v0 (Operator Lifecycle Manager v0) on OpenShift. It provides end-to-end functional tests that validate OLM v0 features and functionality in real OpenShift clusters.

### Purpose
- Validate OLM v0 functionality across different OpenShift topologies
- Test operator installation, upgrade, and lifecycle management scenarios
- Ensure OLM v0 works correctly in various cluster configurations (SNO, HyperShift, Microshift, etc.)
- Provide regression testing for OLM v0 bug fixes and enhancements

### Key Characteristics
- **Framework**: Built on Ginkgo v2 BDD testing framework and OpenShift Tests Extension (OTE)
- **Test Organization**: Polarion-ID based test case management
- **Integration**: Extends `openshift-tests-extension` framework

## Test Case Sources and Organization

### Two Types of Test Cases

#### 1. Migrated Cases from Origin
- **Characteristics**: All robust and stable, meeting OpenShift CI requirements
- **Contribution**: ALL contributed to openshift-tests and used in operator-controller PR presubmit jobs
- **Location**: Should NOT be implemented under `tests-extension/test/qe/specs/`

#### 2. Migrated Cases from tests-private
- **Characteristics**: Some stable, others not
- **Contribution**: Only those meeting OpenShift CI requirements can be contributed to openshift-tests
- **Location**: MUST be implemented under `tests-extension/test/qe/specs/`
- **Auto-Labeling**: Framework automatically adds `Extended` label to these cases
- **Quality Gate**: Cases not meeting CI requirements run in QE-created periodic jobs

### Suite Selection Logic

**For OpenShift General Jobs and PR Presubmit Jobs**:
- Select all cases by default, then exclude unwanted ones
- Migrated cases from Origin: All fit this logic
- Migrated cases from tests-private: Not all fit by default (hence the `Extended` label mechanism)
  - **IMPORTANT**: Only cases with **`Extended` AND `ReleaseGate`** labels can be used in OpenShift General Jobs and PR Presubmit Jobs
  - Cases with only `Extended` (no `ReleaseGate`) can only be used in OLM QE-defined periodic jobs

**Reference**: For OpenShift CI requirements, see [Choosing a Test Suite](https://docs.google.com/document/d/1cFZj9QdzW8hbHc3H0Nce-2xrJMtpDJrwAse9H7hLiWk/edit?tab=t.0#heading=h.tjtqedd47nnu)

## Test Suite Definitions

**IMPORTANT**: These suite definitions are sourced from `cmd/main.go` and may change over time. Always refer to `cmd/main.go` and `test/qe/util/filters/filters.go` for the most current definitions.

### Understanding the Filter Functions

The qualifiers use helper functions from `test/qe/util/filters/filters.go`:
- `BasedStandardTests()`: Non-Extended OR (Extended with ReleaseGate)
- `BasedExtendedTests()`: All Extended tests
- `BasedExtendedReleaseGateTests()`: Extended AND ReleaseGate
- `BasedExtendedCandidateTests()`: Extended AND NOT ReleaseGate
- `BasedExtendedCandidateFuncTests()`: Extended AND NOT ReleaseGate AND NOT StressTest

### Suites for OpenShift General Jobs and PR Presubmit Jobs

**Note**: These suites are used by OpenShift-defined jobs (not defined by OLM QE). All jobs execute tests via `openshift-tests` command.

**Current Status**: PR Presubmit Jobs currently execute only these suites.

**Suite names**: `olmv0/parallel`, `olmv0/serial`, `olmv0/slow`, `olmv0/all`

For complete and current suite definitions and qualifiers, refer to:
- **Suite registration**: `cmd/main.go` (search for `olmv0/` suite names)
- **Filter functions**: `test/qe/util/filters/filters.go` (`BasedStandardTests` and related)

### Suites for Custom Prow Jobs (OLM QE Periodic)

**Note**: These suites are defined by OLM QE team for periodic testing. All jobs execute tests via `openshift-tests` command.

**Current Status**: These suites are currently used only in OLM QE-created periodic jobs. In the future, PR Presubmit QE Jobs will be created to execute these suites as well.

**Suite hierarchy**:
```
olmv0/extended                           # All Extended tests
├── olmv0/extended/releasegate          # Extended + ReleaseGate
└── olmv0/extended/candidate            # Extended without ReleaseGate
    ├── function                        # Functional tests
    │   ├── parallel                    # Can run concurrently
    │   ├── serial                      # Must run one at a time
    │   ├── fast                        # Non-slow (parallel + serial)
    │   └── slow                        # [Slow] tests
    └── stress                          # StressTest label
```

**Key categories**:
- **`candidate/function`**: Functional tests (currently the majority)
- **`candidate/stress`**: Stress tests for resource exhaustion

For complete and current suite definitions and qualifiers, refer to:
- **Suite registration**: `cmd/main.go` (search for `olmv0/extended` suite names)
- **Filter functions**: `test/qe/util/filters/filters.go` (`BasedExtendedTests`, `BasedExtendedCandidateTests`, etc.)

## Directory Structure

```
tests-extension/
├── cmd/
│   └── main.go                   # Test binary entry point
│
├── test/
│   └── qe/
│       ├── specs/                # Test specifications (*.go)
│       │   ├── olmv0_common.go   # Common OLM tests
│       │   ├── olmv0_allns.go    # AllNamespaces install mode tests
│       │   ├── olmv0_multins.go  # MultiNamespace install mode tests
│       │   ├── olmv0_nonallns.go # OwnNamespace/SingleNamespace tests
│       │   ├── olmv0_opm.go      # OPM (operator-package-manager) tests
│       │   ├── olmv0_defaultoption.go # Default option tests
│       │   ├── olmv0_microshift.go    # Microshift-specific tests
│       │   └── olmv0_hypershiftmgmt.go # HyperShift management tests
│       │
│       ├── util/                 # Test utilities and helpers
│       │   ├── client.go         # OpenShift client wrappers
│       │   ├── framework.go      # Test framework setup
│       │   ├── tools.go          # Common test tools
│       │   ├── clusters.go       # Cluster detection utilities
│       │   ├── extensiontest.go  # Extension test helpers
│       │   ├── template.go       # Template processing
│       │   ├── yaml.go           # YAML manipulation
│       │   ├── architecture/     # Architecture detection
│       │   ├── container/        # Container client (Podman/Quay)
│       │   ├── db/               # Database utilities (SQLite)
│       │   ├── filters/          # Test filters
│       │   ├── opmcli/           # OPM CLI wrapper
│       │   └── olmv0util/        # OLM v0 specific utilities
│       │       ├── subscription.go    # Subscription helpers
│       │       ├── check.go           # Resource validation helpers
│       │       ├── catalogsource.go   # CatalogSource utilities
│       │       ├── installplan.go     # InstallPlan helpers
│       │       └── csv.go             # CSV (ClusterServiceVersion) utilities
│       │
│       ├── testdata/             # Test fixtures and manifests
│       └── README.md             # Comprehensive project documentation
│
├── pkg/
│   └── bindata/                  # Embedded test data
│       └── qe/                   # QE test bindata
│
├── bin/                          # Compiled binaries
├── .bingo/                       # Tool dependency management
└── Makefile                      # Build and test automation
```

## Test Case Migration Guide

### A. Code Changes for Migrated Cases

For detailed code changes required when migrating test cases from openshift-tests-private, refer to **[README.md](./README.md)** section "Test Case Migration".

**Key changes summary**:
- Ginkgo functions: `exutil.By()` → `g.By()`
- Check functions: `newCheck().check()` → `olmv0util.NewCheck().Check()`
- Package prefixes: Add `exutil.` and `olmv0util.` to appropriate functions and types

### B. Label Requirements for Migrated and New Cases

For complete label requirements and migration guidelines, refer to **[README.md](./README.md)** section "Label Requirements".

**Essential labels**:
- `[sig-operator]` - Component annotation (required in title)
- `[Jira:OLM]` - Jira component (required in title)
- `PolarionID:xxxxx` - Case ID format (required in title)
- `g.Label("ReleaseGate")` - For cases meeting OpenShift CI requirements
- `g.Label("LEVEL0")` - Level 0 priority tests
- `g.Label("StressTest")` - Stress testing

**Common title tags**:
- `[Skipped:Disconnected]`, `[Skipped:Connected]`, `[Skipped:Proxy]` - Network requirements
- `[Serial]`, `[Slow]`, `[Disruptive]` - Execution characteristics
- `[OCPFeatureGate:xxx]` - Feature gate dependencies

## Test Architecture and Patterns

### Test Structure Pattern

For complete test structure examples, refer to existing test files:
- **Standard tests**: `test/qe/specs/olmv0_common.go`, `test/qe/specs/olmv0_allns.go`
- **Microshift tests**: `test/qe/specs/olmv0_microshift.go`
- **HyperShift management tests**: `test/qe/specs/olmv0_hypershiftmgmt.go`
- **Key patterns**: Look for `g.Describe`, `g.BeforeEach`, `g.AfterEach`, `g.It` blocks

**Basic structure**:
```go
var _ = g.Describe("[sig-operator][Jira:OLM] feature description", func() {
    defer g.GinkgoRecover()
    var oc = exutil.NewCLI("test-name", exutil.KubeConfigPath())

    g.BeforeEach(func() {
        // Setup resources, skip conditions
        exutil.SkipMicroshift(oc)  // For non-Microshift tests
        exutil.SkipNoOLMCore(oc)   // Skip if OLM not installed
    })

    g.AfterEach(func() {
        // Cleanup resources (use defer)
    })

    g.It("PolarionID:xxxxx-test description", g.Label("ReleaseGate"), func() {
        // Test implementation
    })
})
```

### Skip Functions and Cluster Detection

For complete list of skip and detection functions, refer to:
- **Source code**: `test/qe/util/clusters.go`, `test/qe/util/framework.go`
- **Usage examples**: See existing test files in `test/qe/specs/olmv0_*.go`

**Common functions**:
- `SkipMicroshift(oc)` - Skip on Microshift clusters
- `IsMicroshiftCluster(oc)` - Check if running on Microshift
- `IsHypershiftMgmtCluster(oc)` - Check if HyperShift management cluster
- `SkipForSNOCluster(oc)` - Skip on Single Node OpenShift
- `SkipNoOLMCore(oc)` - Skip if OLM not installed
- `IsFeaturegateEnabled(oc, "name")` - Check feature gate status
- `ValidHypershiftAndGetGuestKubeConf(oc)` - Get guest cluster kubeconfig
- `IsAKSCluster(ctx, oc)` - Detect AKS cluster

## Local Development Workflow

For complete local development workflow, build instructions, testing procedures, and PR submission requirements, refer to **[README.md](./README.md)**.

**Quick reference**:
- Build: `make bindata && make build && make update-metadata`
- Find test: `./bin/olmv0-tests-ext list -o names | grep <keyword>`
- Run test: `./bin/olmv0-tests-ext run-test "<full test name>"`
- See README.md for openshift-tests integration and PR submission requirements

## Binary Dependencies

For complete information about binary dependencies (OPM, HyperShift), automatic installation, architecture support, and troubleshooting, refer to **[README.md](./README.md)** section "Binary Dependencies".

**Key points**:
- OPM binary: Auto-downloaded for OPM tests, multi-architecture support
- HyperShift binary: Auto-downloaded for HyperShift management tests
- macOS ARM64: See README for workaround (OPM binaries only available for macOS amd64)

## Test Automation Code Requirements

For complete code quality guidelines, best practices, and common pitfalls, refer to **[README.md](./README.md)** section "Test Automation Code Requirements".

**Critical rules**:
- ✅ Use `defer` for cleanup (before resource creation)
- ✅ Use unique namespaces with random suffixes
- ❌ Don't use `o.Expect` inside `wait.Poll` loops
- ❌ Don't execute logic in `g.Describe` blocks (only initialization)
- ❌ Don't use quotes in test titles (breaks XML parsing)

## Key Utilities

For complete utility APIs and usage examples, refer to the source code and existing tests:

### `exutil` Package
**Location**: `test/qe/util/` directory (e.g., `util/client.go`, `util/framework.go`, `util/tools.go`, `util/clusters.go`)

**Key functions**:
- CLI management: `NewCLI()`, `KubeConfigPath()`
- Resource operations: `OcAction()`, `OcCreate()`, `OcDelete()`, `PatchResource()`
- Cluster detection: `IsMicroshift()`, `IsSNOCluster()`, `IsHyperShiftHostedCluster()`, `IsROSA()`, `IsTechPreviewNoUpgrade()`
- Skip functions: `SkipMicroshift()`, `SkipForSNOCluster()`

### `olmv0util` Package
**Location**: `test/qe/util/olmv0util/` directory (e.g., `util/olmv0util/subscription.go`, `util/olmv0util/catalogsource.go`, `util/olmv0util/helper.go`)

**Key types and methods**:
- `SubscriptionDescription`: Create, Delete, GetCurrentCSV, Wait methods
- `CatalogSourceDescription`: Create, Delete, Wait methods
- `CSVDescription`: WaitSucceeded, Delete methods
- `NewCheck()`: Validation helper for resource state checking

**Usage examples**: See existing test files in `test/qe/specs/olmv0_*.go`

## Anti-Patterns to Avoid

For complete anti-patterns and code examples, refer to **[README.md](./README.md)** and existing test files in `test/qe/specs/olmv0_*.go`.

**Common mistakes**:
- ❌ No cleanup: Always use `defer resource.Delete(oc)` before `resource.Create(oc)`
- ❌ Hardcoded namespaces: Use `namespace := "test-ns-" + exutil.GetRandomString()`
- ❌ Missing timeouts: Always specify timeout for Wait functions
- ❌ Hard sleeps: Use Wait functions instead of `time.Sleep()`

## Quick Reference

### Build and Run
```bash
make build                     # Build test binary
make bindata                   # Regenerate embedded data
make update-metadata          # Update test metadata

# List tests
./bin/olmv0-tests-ext list -o names | grep "keyword"

# Run test
./bin/olmv0-tests-ext run-test "<full test name>"
```

### Test Naming Convention
```
[sig-operator][Jira:OLM] OLMv0 <feature> PolarionID:XXXXX-[Skipped:XXX]description[Serial|Slow|Disruptive]
```

### Key Labels
- `ReleaseGate` - Promotes Extended case to openshift-tests
- `Extended` - Auto-added to cases under test/qe/specs/
- `LEVEL0` - Level 0 priority test
- `StressTest` - Stress testing
- `NonHyperShiftHOST` - Skip on HyperShift hosted clusters

## Resources

- [OLM v0 Source Code](https://github.com/openshift/operator-framework-olm/tree/main/staging/operator-lifecycle-manager)
- [Ginkgo v2 Documentation](https://onsi.github.io/ginkgo/)
- [OpenShift Tests Extension](https://github.com/openshift-eng/openshift-tests-extension)
- [Test Extensions in Origin](https://github.com/openshift/origin/blob/main/docs/test_extensions.md)
- [OpenShift CI Requirements](https://docs.google.com/document/d/1cFZj9QdzW8hbHc3H0Nce-2xrJMtpDJrwAse9H7hLiWk/edit?tab=t.0#heading=h.tjtqedd47nnu)

## Debugging

**Investigation Priority** when tests fail:
1. **First**: Check test code in `tests-extension/test/qe/`
2. **Second**: Check test utilities in `tests-extension/test/qe/util/olmv0util/`
3. **Third**: Check resource status and conditions via `oc describe`
4. **Fourth**: Check OLM operator and catalog operator logs
5. **Last**: Refer to product code to understand expected behavior

**For deeper investigation** (when you need to refer to product code):
1. **Identify test category**: Check if it's an OPM test (`olmv0_opm.go`) or OLM core test
2. **Locate product code**: See **Product Code References** section below
3. **Read product AGENTS.md**: For OLM core issues, read the [OLM AGENTS.md](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/AGENTS.md)
4. **Trace code flow**: Use product code to understand expected behavior
5. **Compare implementation**: Check if test expectations match product implementation
6. **Check recent changes**: Look for recent commits that might have changed behavior

**Key Namespaces** (OpenShift):
- `openshift-operator-lifecycle-manager`: OLM operator and catalog operator
- `openshift-marketplace`: Marketplace and catalog sources
- `openshift-operators`: Default operator installation namespace

**Common Debugging Commands**:
```bash
# Check resource status
oc get csv -A
oc get subscription -A
oc get installplan -A
oc describe csv <name> -n <namespace>
oc describe subscription <name> -n <namespace>

# Check logs
oc logs -n openshift-operator-lifecycle-manager deployment/olm-operator -f
oc logs -n openshift-operator-lifecycle-manager deployment/catalog-operator -f
```

## Product Code References for Debugging

**IMPORTANT**: Most test failures are due to issues in the test code itself, not the product code. Always investigate and fix the test code first. Only refer to product code when you need to understand expected product behavior or verify assumptions about how the product should work.

This test repository contains two main categories of test cases. When writing or debugging tests, you may occasionally need to understand the corresponding product code:

### Test Case Categories and Code Repositories

#### 1. OPM Test Cases
**Product Code**: [operator-registry](https://github.com/openshift/operator-framework-olm/tree/main/staging/operator-registry)
- Contains OPM (Operator Package Manager) CLI implementation
- Tests for `opm` commands: index, serve, alpha, etc.
- **Note**: This repository does not have an AGENTS.md yet

**Test Files** (current list, may expand over time):
- `test/qe/specs/olmv0_opm.go`
- Additional test files may be added as new OPM features are covered

#### 2. Non-OPM Test Cases (OLM Core Functionality)
**Product Code**: [operator-lifecycle-manager](https://github.com/openshift/operator-framework-olm/tree/main/staging/operator-lifecycle-manager)
- Contains OLM v0 core implementation
- Manages operators, CSVs, Subscriptions, InstallPlans, CatalogSources
- **AGENTS.md Available**: [OLM AGENTS.md](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/AGENTS.md) - highly recommended for understanding OLM internals

**Test Files** (current list, may expand over time):
- `test/qe/specs/olmv0_common.go`
- `test/qe/specs/olmv0_allns.go`
- `test/qe/specs/olmv0_multins.go`
- `test/qe/specs/olmv0_nonallns.go`
- `test/qe/specs/olmv0_defaultoption.go`
- `test/qe/specs/olmv0_microshift.go`
- `test/qe/specs/olmv0_hypershiftmgmt.go`
- Additional test files may be added as new OLM features are covered

#### 3. Additional Reference
**Product Code**: [operator-marketplace](https://github.com/operator-framework/operator-marketplace)
- Legacy marketplace operator (mostly deprecated in OLM v0)
- May still be referenced for historical context or specific marketplace features

## Notes for AI Agents

### Suggesting Test Locations

When discussing whether a feature needs testing:

**✅ DO**: Provide simple, focused guidance on QE test placement
- Example: "If you need to write QE tests for this functionality, they should go in `tests-extension/test/qe/specs/`."
- Keep suggestions within the scope of this AGENTS.md (QE tests only)

**❌ DON'T**:
- Discuss DEV test locations (e.g., unit tests in product code directories)
- Explain the difference between QE and DEV tests unless explicitly asked
- Provide detailed test categorization unless the user is actively writing tests

**Remember**: This AGENTS.md is for QE test code in `tests-extension/` only. Product code testing (DEV tests) is outside this scope.

### Critical Points

- **Test Location Matters**: Cases under `test/qe/specs/` auto-get `Extended` label
- **ReleaseGate is Critical**: Determines if Extended case can be used in OpenShift General Jobs and PR Presubmit Jobs (all cases are executed via `openshift-tests` command)
- **Most Failures are Test Code Issues**: Always investigate test code first before looking at product code
- **Suite Definitions Change**: Always refer to `cmd/main.go` for current suite definitions, not just this document

### Test Development

- **Suite Logic**: Understand the qualifier logic for different test suites
- **Migration Mapping**: Use the mapping table to find original test cases
- **Code Changes**: Follow the A/B migration guide strictly
- **Label Requirements**: All tests MUST have `[sig-operator][Jira:OLM]` and PolarionID
- **Cleanup is Mandatory**: Always use defer for resource deletion
- **Random Namespaces**: Use `exutil.GetRandomString()` for unique namespace names

### Cluster Topologies

- **Topology Awareness**: Consider SNO, HyperShift, Microshift, ROSA, OSD, ARO
- **Skip Functions**: Use appropriate skip functions for different topologies
  - Standard tests: `exutil.SkipMicroshift(oc)`
  - Microshift tests: `if !exutil.IsMicroshiftCluster(oc) { g.Skip(...) }`
  - HyperShift mgmt tests: Use `g.Label("NonHyperShiftHOST")` and `IsHypershiftMgmtCluster()`

- **Network Connectivity**: Use skip labels in test titles for network-dependent tests
  - `[Skipped:Disconnected]`: Test requires internet access, skip on disconnected clusters
  - `[Skipped:Connected]`: Test requires disconnected environment, skip on connected clusters
  - `[Skipped:Proxy]`: Test incompatible with proxy configuration, skip on proxy clusters
  - Example: `g.It("PolarionID:12345-[Skipped:Disconnected]test description", func() {...})`

### Common Pitfalls

- **Feature Gates**: Three distinct handling patterns based on test behavior
- **Binary Dependencies**: OPM and HyperShift binaries auto-installed
- **No Quotes in Titles**: Causes XML parsing failures
- **No Expect in Poll**: Return errors, don't assert in wait loops
- **No Logic in g.Describe**: Move all logic to g.BeforeEach or g.It

### Build and Run

- **Before PR**: Run `make bindata && make build && make update-metadata`
- **Local Testing**: Use `./bin/olmv0-tests-ext list -o names | grep <id>` to find test names
- **Test File Changes**: If adding test files under `test/qe/specs/`, they auto-get `Extended` label
