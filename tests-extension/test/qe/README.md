## Overview

When creating test cases based on OTE (OpenShift Tests Extension) in operator-framework-olm, there are two sources:

### 1. Migrated Cases from Origin
- These cases are all robust and stable, meeting OpenShift CI requirements
- All of them are contributed to openshift-tests and used in operator-controller PR presubmit jobs

### 2. Migrated Cases from tests-private
- Some of these cases are stable, others are not
- Only those meeting OpenShift CI requirements can be contributed to openshift-tests and used in operator-controller PR presubmit jobs
- The remaining cases that don't meet OpenShift CI requirements are run by QE-created periodic jobs

## Suite Selection Logic

Test suites for openshift-tests and PR presubmit jobs select all cases by default, then exclude unwanted ones:
- **migrated cases from Origin**: All can be contributed to openshift-tests (fits this logic)
- **migrated cases from tests-private**: Not all can be contributed to openshift-tests by default (doesn't fit this logic)

We need to identify all cases from tests-private among all cases, then mark which cases can be contributed to openshift-tests and PR presubmit jobs.

> **Note**: For OpenShift CI requirements, refer to: [Choosing a Test Suite](https://docs.google.com/document/d/1cFZj9QdzW8hbHc3H0Nce-2xrJMtpDJrwAse9H7hLiWk/edit?tab=t.0#heading=h.tjtqedd47nnu)

## Implementation Strategy

### Test Case Organization

1. **Case from tests-private Identification**: These cases must be implemented under `tests-extension/test/qe/specs/`
   - Test framework automatically adds `Extended` label to the cases
   - Enables automatic cases identification without requiring authors to add labels

2. **Case from Origin Placement**: These cases should NOT be implemented under `tests-extension/test/qe/specs/`

3. **OpenShift CI Compatibility for cases from tests-private**: If the author believes a case meets OpenShift CI requirements, add the `ReleaseGate` label:
   ```go
   g.It("xxxxxx", g.Label("ReleaseGate"), func() {
   ```
   - This makes the case equivalent to origin cases for openshift-tests
   - For the cases with `ReleaseGate` that need `Informing`, add:
     ```go
     import oteg "github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"
     g.It("xxxxxx", g.Label("ReleaseGate"), oteg.Informing(), func() {
     ```

## Suite Definitions

### Suites for openshift-tests and PR presubmit jobs:

#### Parallel Suite
```go
	ext.AddSuite(e.Suite{
		Name:    "olmv0/parallel",
		Parents: []string{"openshift/conformance/parallel"},
		Qualifiers: []string{
			`((!labels.exists(l, l=="Extended")) || (labels.exists(l, l=="Extended") && labels.exists(l, l=="ReleaseGate"))) &&
			!(name.contains("[Serial]") || name.contains("[Slow]"))`,
		},
	})
```

#### Serial Suite
```go
	ext.AddSuite(e.Suite{
		Name:    "olmv0/serial",
		Parents: []string{"openshift/conformance/serial"},
		Qualifiers: []string{
			`((!labels.exists(l, l=="Extended")) || (labels.exists(l, l=="Extended") && labels.exists(l, l=="ReleaseGate"))) &&
			(name.contains("[Serial]") && !name.contains("[Disruptive]") && !name.contains("[Slow]"))`,
			// refer to https://github.com/openshift/origin/blob/main/pkg/testsuites/standard_suites.go#L456
		},
	})
```

#### Slow Suite
```go
	ext.AddSuite(e.Suite{
		Name:    "olmv0/slow",
		Parents: []string{"openshift/optional/slow"},
		Qualifiers: []string{
			`((!labels.exists(l, l=="Extended")) || (labels.exists(l, l=="Extended") && labels.exists(l, l=="ReleaseGate"))) &&
			name.contains("[Slow]")`,
		},
	})
```

#### All Suite
```go
	ext.AddSuite(e.Suite{
		Name: "olmv0/all",
		Qualifiers: []string{
			`(!labels.exists(l, l=="Extended")) || (labels.exists(l, l=="Extended") && labels.exists(l, l=="ReleaseGate"))`,
		},
	})
```

### Suites for Custom Prow jobs:

#### Extended All Suite
```go
	ext.AddSuite(e.Suite{
		Name: "olmv0/extended",
		Qualifiers: []string{
			`labels.exists(l, l=="Extended")`,
		},
	})
```

#### Extended ReleaseGate Suite
```go
	ext.AddSuite(e.Suite{
		Name: "olmv0/extended/releasegate",
		Qualifiers: []string{
			`labels.exists(l, l=="Extended") && labels.exists(l, l=="ReleaseGate")`,
		},
	})
```

#### Extended Candidate Suite
```go
	ext.AddSuite(e.Suite{
		Name: "olmv0/extended/candidate",
		Qualifiers: []string{
			`labels.exists(l, l=="Extended") && !labels.exists(l, l=="ReleaseGate")`,
		},
	})
```

#### Extended Candidate Parallel Suite
```go
	ext.AddSuite(e.Suite{
		Name: "olmv0/extended/candidate/parallel",
		Qualifiers: []string{
			`(labels.exists(l, l=="Extended") && !labels.exists(l, l=="ReleaseGate") && !labels.exists(l, l=="StressTest")) &&
			!(name.contains("[Serial]") || name.contains("[Slow]"))`,
		},
	})
```

#### CExtended Candidate Serial Suite
```go
	ext.AddSuite(e.Suite{
		Name: "olmv0/extended/candidate/serial",
		Qualifiers: []string{
			`(labels.exists(l, l=="Extended") && !labels.exists(l, l=="ReleaseGate") && !labels.exists(l, l=="StressTest")) &&
			(name.contains("[Serial]") && !name.contains("[Slow]"))`,
		},
	})
```

#### Extended Candidate Slow Suite
```go
	ext.AddSuite(e.Suite{
		Name: "olmv0/extended/candidate/slow",
		Qualifiers: []string{
			`(labels.exists(l, l=="Extended") && !labels.exists(l, l=="ReleaseGate") && !labels.exists(l, l=="StressTest")) &&
			name.contains("[Slow]")`,
		},
	})
```

## Test Case Migration Guide

**Required For all QE cases**:
- Do not use `&|!,()/` in case title
- Do NOT remove the PolarionID number from the `original-name` label. The PolarionID in `g.Label("original-name:...")` must include the case ID number.
  - ✅ **Correct**: `g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 optional should PolarionID:68679-[Skipped:Disconnected]catalogsource with invalid name is created")`
  - ❌ **Wrong**: `g.Label("original-name:[sig-operator][Jira:OLM] OLMv0 optional should PolarionID:[Skipped:Disconnected]catalogsource with invalid name is created")` (missing case ID)

### A. Code Changes for Migrated Cases

All migrated test case code needs the following changes to run in the new test framework:

1. Change `exutil.By()` to `g.By()`
2. Change `newCheck()` to `olmv0util.NewCheck()`, change `check(oc)` to `Check(oc)`
3. Change `getResource()` to `olmv0util.GetResource()`
4. Change `patchResource()` to `olmv0util.PatchResource()`
5. Change `getRandomString()` to `exutil.GetRandomString()`
6.   - asAdmin → exutil.AsAdmin
     - asUser → exutil.AsUser
     - withNamespace → exutil.WithNamespace
     - withoutNamespace → exutil.WithoutNamespace
     - present → exutil.Present
     - compare → exutil.Compare
     - contain → exutil.Contain
     - ok → exutil.Ok
     - nok → exutil.Nok
7.  - describerResrouce → olmv0util.DescriberResrouce
    - operatorGroupDescription → olmv0util.OperatorGroupDescription
    - catalogSourceDescription → olmv0util.CatalogSourceDescription
    - subscriptionDescription → olmv0util.SubscriptionDescription
    - projectDescription → olmv0util.ProjectDescription
    - configMapDescription → olmv0util.ConfigMapDescription
    - csvDescription → olmv0util.CsvDescription
    - crdDescription → olmv0util.CrdDescription
    - checkList → olmv0util.CheckList

### B. Label Requirements for Migrated and New Cases

#### Required Labels in case title
1. **Component annotation**: Add `[sig-operator]` in case title
2. **Jira Component**: Add `[Jira:OLM]` in case title
3. **OpenShift CI compatibility**: If you believe the case meets OpenShift CI requirements, add `ReleaseGate` label to Ginkgo
   - **Note**: Don't add `ReleaseGate` if case title contains `Disruptive` or `Slow`, or labels contain `StressTest`
4. **Required For Migrated case from test-private**: Add `[OTP]` in case title

#### Optional Labels in Migration/New test cases' title
1. **LEVEL0**: Add `[Level0]` in the case title as a title tag. Do NOT use `g.Label("LEVEL0")`.
   - ✅ **Correct**: `g.It("PolarionID:72192-[Level0][OTP]-description", func() { ... })`
   - ❌ **Wrong**: `g.It("PolarionID:72192-[OTP]-description", g.Label("LEVEL0"), func() { ... })`
2. **Author**: Deprecated, remove it.
3. **ConnectedOnly**: Add `[Skipped:Disconnected]` in title
4. **DisconnectedOnly**: Add `[Skipped:Connected][Skipped:Proxy]` in title
5. **Case ID**: change it to `PolarionID:xxxxxx` format, and remove the old one from the case title. Such as `-72017-` strings.
   - **IMPORTANT**: The PolarionID number should only appear ONCE in the test title - at the beginning as `PolarionID:xxxxx`. Do NOT repeat the number anywhere else in the title.
   - **IMPORTANT**: Do NOT add `-` between two consecutive square brackets. Adjacent tags should be written directly together.
   - ✅ **Correct**: `PolarionID:73201-[OTP][Skipped:Disconnected]catalog pods do not recover from node failure [Disruptive][Serial]`
   - ❌ **Wrong**: `PolarionID:73201-[OTP]-[Skipped:Disconnected]catalog pods do not recover from node failure [Disruptive][Serial]` (dash between brackets)
   - ❌ **Wrong**: `PolarionID:73201-[OTP][Skipped:Disconnected]73201-catalog pods do not recover from node failure [Disruptive][Serial]` (repeated ID)
   - ❌ **Wrong**: `PolarionID:22070-[OTP][Skipped:Disconnected]22070-support grpc sourcetype [Serial]` (repeated ID)
6. **Importance**: Deprecated, remove it. Such as `Critical`, `High`, `Medium` and `Low` strings.
7. **NonPrerelease**: Deprecated, remove it.
    - **Longduration**: Change it to `[Slow]` in case title.
    - **ChkUpg**: Deprecated, remove it. Not supported (openshift-tests upgrade differs from OpenShift QE)
8.  **VMonly**: Deprecated, and don't migrate the `VMonly` test cases to here. 
9.  **Slow, Serial, Disruptive**: Preserved, but add them in the end of the title. Such as `"[sig-operator][Jira:OLM] OLMv0 optional should PolarionID: xxx ...[Slow][Serial][Disruptive]"`
10. **DEPRECATED**: Deprecated, don't add this kind of case to here. But, if your test case has been merged into this repo, please add this case into the [IgnoreObsoleteTests](https://github.com/openshift/operator-framework-olm/blob/main/tests-extension/cmd/main.go#L272).
11. **CPaasrunOnly, CPaasrunBoth, StagerunOnly, StagerunBoth, ProdrunOnly, ProdrunBoth**: Deprecated, remove them.
12. **StressTest**: Use Ginkgo label `g.Label("StressTest")`
13. **NonHyperShiftHOST**: Use Ginkgo label `g.Label("NonHyperShiftHOST")` or use `IsHypershiftHostedCluster` judgment, then skip
14. **HyperShiftMGMT**: Deprecated. Use `g.Label("NonHyperShiftHOST")` and `ValidHypershiftAndGetGuestKubeConf` validation instead.
15. **MicroShiftOnly**: Deprecated. Use `SkipMicroshift` instead.
16. **ROSA**: Deprecated. Three ROSA job types:
    - `rosa-sts-ovn`: equivalent to OCP
    - `rosa-sts-hypershift-ovn`: equivalent to hypershift hosted
    - `rosa-classic-sts`: doesn't use openshift-tests
17. **ARO**: Deprecated. All ARO jobs based on HCP are equivalent to hypershift hosted (don't actually use openshift-test)
18. **OSD_CCS**: Deprecated. Only one job type: `osd-ccs-gcp` equivalent to OCP
19. **Feature Gates**: Handle test cases based on their feature gate requirements:

    **Case 1: Test only runs when feature gate is enabled**
    - The test should not execute if the feature gate is disabled
    - Add `[OCPFeatureGate:xxxx]` in `g.It` title (where xxxx is feature gate name)
    - Or use `IsFeaturegateEnabled` check, then skip if disabled
    - Remove label/check when feature no longer requires gate
    
    **Case 2: Test runs with/without feature gate but with different behaviors**
    - The test executes regardless of feature gate status, but behaves differently
    - Use `IsFeaturegateEnabled` check to handle different behaviors
    - Do NOT add `[OCPFeatureGate:xxxx]` label
    - Remove `IsFeaturegateEnabled` check when feature no longer requires gate
    
    **Case 3: Test runs with/without feature gate with same behavior**
    - The test executes the same way regardless of feature gate status
    - Do NOT use `IsFeaturegateEnabled` check
    - Do NOT add `[OCPFeatureGate:xxxx]` label
20. **Exclusive**: change to `Serial`

### C. Don't output the sensitive info in the log

## Disconnected Environment Support for Migrated QE cases

**IMPORTANT**: With IDMS/ITMS mirror configuration in place, disconnected environments work exactly like connected environments.

**What this means:**
- Write test cases the same way you would for connected environments
- Create ClusterCatalogs directly - no environment detection needed
- IDMS/ITMS automatically redirects image pulls to mirror registry
- No special helper functions or conditional logic required

**Image Requirements for Migrated QE Cases:**
- All operator images (bundle, base, index) must be hosted under `quay.io/openshifttest` or `quay.io/olmqe`
- This ensures images are mirrored to disconnected environments via IDMS/ITMS configuration
- Images from other registries will not be available in disconnected clusters

**Environment Validation for Disconnected-Supporting Migrated Test Cases:**

**When to use `ValidateAccessEnvironment`:**

1. **Test cases that create CatalogSource or Subscription**:
   - If your test supports disconnected environments (both connected+disconnected, or disconnected-only)
   - AND your test creates CatalogSource or Subscription resources
   - **MUST** call `ValidateAccessEnvironment(oc)` at the beginning of the test
   - This applies to both newly created test cases and migrated test cases

2. **Test cases that do NOT create both CatalogSource and Subscription**:
   - Optional to use `ValidateAccessEnvironment(oc)`
   - Using it won't cause errors, but it's not required
   - The validation is primarily for ensuring catalog images can be mirrored

**Usage example:**

```go
g.It("test case supporting disconnected", func() {
    olmv0util.ValidateAccessEnvironment(oc)  // MUST call if creating CatalogSource/Subscription
    // rest of test code
})
```

**What ValidateAccessEnvironment does:**
- **Proxy clusters**: Returns immediately (no validation needed, proxy provides external access)
- **Connected clusters**: Returns immediately after quick network check (no validation needed)
- **Disconnected clusters**: Validates that ImageTagMirrorSet `image-policy-aosqe` is configured
  - If ITMS is configured: Test proceeds normally
  - If ITMS is missing: Test is skipped with clear message explaining what's missing


## Test Automation Code Requirements

Consider these requirements when writing and reviewing code:

### Security Considerations
- Does the test case generate sensitive information in logs?
- Does the code contain sensitive information in output or commands?

### Test Isolation
- Will this test case affect other test executions?
- Will this test case be affected by other test executions?

### Labeling and Cleanup
- Are correct labels applied?
- What changes does this case make to the cluster?
- Can changes be restored for both normal and abnormal exits?
- During recovery, are both actions and results correct?
- Should recovery restore to predetermined or dynamically determined values?

### Logging Best Practices
- Avoid excessive logs or large error messages
- Don't put large log outputs in error messages(use proper log messages instead). Don't use `o.Expect` to assert large messages (appears in error message on failure)
- Avoid logging `oc logs` output directly

### Code Quality
- Don't modify shared libraries (e.g., Ginkgo) or global settings affecting other tests
- Don't execute logic code in `g.Describe` except for initing oc, and move to `g.BeforeEach`
- Don't use single/double quotes in case titles (causes XML parse failures)
- **JSONPath field names must use lowercase**: When using `-o=jsonpath={}` with `oc` commands, all field names must be lowercase
  ```go
  // Wrong - capitalized field names:
  .metadata.Name              // ❌
  .spec.Template              // ❌
  @.Name                      // ❌
  .subjects[0].Name           // ❌
  .ownerReferences[0].Name    // ❌

  // Correct - lowercase field names:
  .metadata.name              // ✅
  .spec.template              // ✅
  @.name                      // ✅
  .subjects[0].name           // ✅
  .ownerReferences[0].name    // ✅

  // Examples:
  // Wrong:
  oc.Run("get").Args("pod", "mypod", "-o=jsonpath={.metadata.Name}").Output()
  oc.Run("get").Args("deploy", "mydeploy", "-o=jsonpath={.spec.Template.spec.containers[0].image}").Output()

  // Correct:
  oc.Run("get").Args("pod", "mypod", "-o=jsonpath={.metadata.name}").Output()
  oc.Run("get").Args("deploy", "mydeploy", "-o=jsonpath={.spec.template.spec.containers[0].image}").Output()
  ```
- Avoid `o.Expect` in `wait.Poll`:
  ```go
  // Wrong:
  wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, false, func(ctx context.Context) (bool, error) {
		response, err := c.AuthorizationV1().SelfSubjectAccessReviews().Create(context.Background(), review, metav1.CreateOptions{})
		o.Expect(err).NotTo(o.HaveOccurred()) // in wait.Poll
		return response.Status.Allowed == allowed, nil
	})
  
  // Correct:
  wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, false, func(ctx context.Context) (bool, error) {
		response, err := c.AuthorizationV1().SelfSubjectAccessReviews().Create(context.Background(), review, metav1.CreateOptions{})
		if err != nil {
			return false, err
		}
		return response.Status.Allowed == allowed, nil
	})
  ```

## Using Claude Code for Development of QE Case

If you are using Claude Code as your AI coding assistant:

1. **Start Claude Code from tests-extension directory**:
   ```bash
   cd tests-extension/
   # Then launch Claude Code from this directory
   claude
   ```

2. **Load AGENTS.md**:
   - If starting from `tests-extension/` directory: AGENTS.md auto-loads (no prompt)
   - If starting from a subdirectory (e.g., `tests-extension/test/qe/`): On first launch, Claude Code will prompt you to load the parent AGENTS.md - select **Yes** (subsequent launches will auto-load)
   - Use `/memory` to verify AGENTS.md is loaded and view its content

This ensures Claude Code has access to:
- Test framework architecture and patterns
- Migration guidelines from tests-private
- Suite definitions and label requirements
- Code quality standards and anti-patterns

## Local Development Workflow

### Environment Configuration for Migrated QE cases

**IMPORTANT**: With IDMS/ITMS in place, tests work the same in both connected and disconnected environments. No special configuration is needed.

### Before Submitting PR

1. **Build and compile**:
   ```bash
   $ cd tests-extension
   $ make bindata
   $ make build
   ```

2. **Check test name**:
   ```bash
   $ ./bin/olmv0-tests-ext -h
OLMv0 Tests Extension

Usage:
   [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  images      List test images
  info        Display extension metadata
  list        List items
  run-suite   Run a group of tests by suite. This is more limited than origin, and intended for light local development use. Orchestration parameters, scheduling, isolation, etc are not obeyed, and Ginkgo tests are executed serially.
  run-test    Runs tests by name
  update      Update test metadata

Flags:
  -h, --help   help for this command

Use " [command] --help" for more information about a command.

   # List all test names and search for your test using a keyword
   $ ./bin/olmv0-tests-ext list -o names | grep "keyword_from_your_test_name"
   
   # Example: If your test is about "catalog installation", search for:
   $ ./bin/olmv0-tests-ext list -o names | grep "catalog"
   # This will show the full test name like:
   # [sig-operator][Jira:OLM] OLMv0 catalog installation should succeed
   ```

3. **Run test locally**:
   ```bash
   $ ./bin/olmv0-tests-ext run-test <full test name>
   ```
   For example, 
   ```console
   jiazha-mac:tests-extension jiazha$ ./bin/olmv0-tests-ext list -o names | grep 43271
   [sig-operator][Jira:OLM] OLMv0 optional should PolarionID:43271-[OTP]-Bundle Content Compression

   jiazha-mac:tests-extension jiazha$ ./bin/olmv0-tests-ext run-test -n "[sig-operator][Jira:OLM] OLMv0 optional should PolarionID:43271-[OTP]-Bundle Content Compression"
   I1120 12:13:20.979754 47868 test_context.go:566] The --provider flag is not set. Continuing as if --provider=skeleton had been used.
   Running Suite:  - /Users/jiazha/goproject/operator-framework-olm/tests-extension
   ================================================================================
   Random Seed: 1763612000 - will randomize all specs

   Will run 1 of 1 specs
   ------------------------------
   [sig-operator][Jira:OLM] OLMv0 optional should PolarionID:43271-[OTP]-Bundle Content Compression [original-name:[sig-operator][Jira:OLM] OLMv0 optional should PolarionID:43271-[OTP]-Medium-43191-Medium-43271-Bundle Content Compression]
   ...
   ```
   
   **Keep generated temporary project for Debugging**
   Add the Env Var: `export DELETE_NAMESPACE=false`. These random namespaces will be kept, like below:
   ```console
   jiazha-mac:tests-extension jiazha$ oc get ns 
   NAME                                               STATUS   AGE
   default                                            Active   76m
   e2e-test-default-1a2fc8d6-2jr22                    Active   119s
   e2e-test-default-1a2fc8d6-fg54z                    Active   104s
   ```

4. **Test with openshift-tests**:
   - Switch to origin repo
   - Follow [test extensions documentation](https://github.com/openshift/origin/blob/main/docs/test_extensions.md)
   - Set environment variables:
     ```bash
     OPENSHIFT_TESTS_DISABLE_CACHE=1
     EXTENSION_BINARY_OVERRIDE_INCLUDE_TAGS=tests,olm-operator-controller
     EXTENSION_BINARY_OVERRIDE_OLM_OPERATOR_CONTROLLER=<path to repo>/operator-framework-olm/tests-extension/bin/olmv0-tests-ext
     EXTENSIONS_PAYLOAD_OVERRIDE=<ocp arm payload> # For AMD cluster with ARM laptop:
     # EXTENSIONS_PAYLOAD_OVERRIDE=registry.ci.openshift.org/ocp-arm64/release-arm64:4.20.0-0.nightly-arm64-2025-08-31-123924
     ```
   - Run appropriate suite based on your test characteristics:
     ```bash
     # Choose the suite that matches your test type:
     
     # For parallel tests (most common):
     ./openshift-tests run olmv0/parallel --monitor watch-namespaces
     
     # For parallel tests which does not contributed to openshift-tests
     ./openshift-tests run olmv0/extended/candidate/parallel --monitor watch-namespaces
     ```

5. **Please update metadata if test case title changed**:
   ```bash
   make update-metadata
   ```
   - If test case title changed, refer to "How to Keep Test Names Unique"

6. **Create PR**

### PR Submission Requirements

#### Pre-submission Checks
1. Check failed presubmit jobs - verify both your new cases and whether other case failures are caused by your changes

#### Stability Testing
2. **For parallel cases** contributing to openshift-tests:
   ```bash
   /payload-aggregate periodic-ci-openshift-release-master-ci-<release version>-e2e-gcp-ovn-techpreview 5
   # Example: /payload-aggregate periodic-ci-openshift-release-master-ci-4.20-e2e-gcp-ovn-techpreview 5
   
   /payload-aggregate periodic-ci-openshift-release-master-ci-<release version>-e2e-gcp-ovn 5
   ```

3. **For serial cases** contributing to openshift-tests:
   ```bash
   /payload-aggregate periodic-ci-openshift-release-master-ci-<release version>-e2e-gcp-ovn-techpreview-serial 5
   # Example: /payload-aggregate periodic-ci-openshift-release-master-ci-4.20-e2e-gcp-ovn-techpreview-serial 5
   
   /payload-aggregate periodic-ci-openshift-release-master-ci-<release version>-e2e-gcp-ovn-serial 5
   ```

## Test Case Migration Mapping

The following table shows the mapping between the original g.Describe blocks in `openshift-tests-private/test/extended/operators/olm.go` and the corresponding migrated spec files in this repository:

### Binary Dependencies

This test framework automatically installs required binaries before running tests:

#### OPM Binary Support
- **Automatic Installation**: The OPM binary is automatically downloaded and installed before running OPM-related tests
- **Multi-Architecture Support**: Supports Linux (amd64, arm64, ppc64le, s390x) and macOS (amd64 only)
- **Concurrent Safe**: Uses file locking to prevent conflicts when multiple test processes run simultaneously
- **Architecture Detection**: Automatically selects the correct binary for the current platform

**Special Note for macOS ARM64 Developers**:
If you're developing OPM test cases on an ARM64 macOS machine, the framework will skip tests since OPM binaries are only available for macOS amd64. To work around this:
1. Build OPM from source for ARM64 macOS
2. Place the compiled `opm` binary in your system PATH
3. The framework will detect the existing binary and skip automatic installation

#### HyperShift Binary Support
- **Automatic Installation**: The HyperShift binary is automatically downloaded and installed for HyperShift management cluster tests
- **Architecture Requirements**: Only supports Linux x86-64 architecture
- **Concurrent Safe**: Uses file locking similar to OPM binary installation

### Test File Mapping

#### From openshift-tests-private/test/extended/operators/olm.go

| Original g.Describe in olm.go | Mapped Spec File | Line in olm.go |
|-------------------------------|------------------|----------------|
| `[sig-operators] OLM optional` | `olmv0_defaultoption.go` | 34 |
| `[sig-operators] OLM should` | `olmv0_defaultoption.go` | 150 |
| `[sig-operators] OLM for an end user use` | `olmv0_common.go` | 5302 |
| `[sig-operators] OLM for an end user handle common object` | `olmv0_common.go` | 5349 |
| `[sig-operators] OLM for an end user handle within a namespace` | `olmv0_nonallns.go` | 5555 |
| `[sig-operators] OLM for an end user handle to support` | `olmv0_multins.go` | 11544 |
| `[sig-operators] OLM for an end user handle within all namespace` | `olmv0_allns.go` | 12657 |
| `[sig-operators] OLM on hypershift` | `olmv0_hypershiftmgmt.go` | 14814 |

#### From openshift-tests-private/test/extended/opm/opm.go

| Original Content | Mapped Spec File | Coverage |
|------------------|------------------|----------|
| All OPM CLI functionality tests | `olmv0_opm.go` | Complete |

### Notes:
- Each mapped spec file contains a comment indicating its relationship to the original olm.go g.Describe block or source file
- Some spec files map to multiple g.Describe blocks (e.g., `olmv0_common.go` and `olmv0_defaultoption.go`)
- `olmv0_opm.go` maps to OPM CLI functionality tests from openshift-tests-private

## How to Keep Test Names Unique

OTE requires unique test names. If you want to modify a test name after merging, refer to the [Makefile implementation](https://github.com/openshift/operator-framework-operator-controller/blob/main/openshift/tests-extension/Makefile#L104) for proper modification procedures.
