package util

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	g "github.com/onsi/ginkgo/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	e2e "k8s.io/kubernetes/test/e2e/framework"
	psapi "k8s.io/pod-security-admission/api"
)

const (
	// podSecurityLabelSyncControlLabel is the OpenShift SCC-to-PSA label-sync
	// controller sentinel. When set on a namespace, the controller may mutate
	// the namespace's pod security labels after initial evaluation (see
	// OCPBUGS-115142), so its value is recorded with the audit result for
	// context.
	podSecurityLabelSyncControlLabel = "security.openshift.io/scc.podSecurityLabelSync"

	// psaAuditTimeout bounds a single in-process PSA audit so a hung request
	// cannot stall the test cleanup indefinitely.
	psaAuditTimeout = 2 * time.Minute

	// psaAuditLevel is the pod security level the audit checks namespaces
	// against.
	psaAuditLevel = psapi.LevelRestricted
)

// PSA audit modes, selected through the PSA_AUDIT_MODE environment variable.
const (
	// psaAuditOff skips the audit entirely.
	psaAuditOff = "off"
	// psaAuditWarn records violations (log + artifact) without failing the
	// test.
	psaAuditWarn = "warn"
	// psaAuditEnforce fails the test when violations are found.
	psaAuditEnforce = "enforce"
)

// psaAuditMode resolves the audit mode for the current run.
//
// PSA_AUDIT_MODE selects the mode directly: "enforce" fails scenarios that
// violate the audit level, "warn" records violations without failing, and
// "off" skips the audit. Unrecognized values fail closed to "enforce".
//
// When PSA_AUDIT_MODE is unset, the audit stays backwards compatible with the
// release-repo step integration (openshift/release#84940): PSA_CHECK_REQUIRED
// set to true selects "enforce" and anything else skips the audit. The
// PSA_CHECK_BIN environment variable is no longer used: the audit runs
// in-process instead of executing an external kubectl-dev_tool binary, so the
// tool no longer needs to be staged into the test execution context.
func psaAuditMode() string {
	if mode := strings.TrimSpace(os.Getenv("PSA_AUDIT_MODE")); mode != "" {
		switch strings.ToLower(mode) {
		case psaAuditOff, psaAuditWarn:
			return strings.ToLower(mode)
		default:
			return psaAuditEnforce
		}
	}
	if required, err := strconv.ParseBool(os.Getenv("PSA_CHECK_REQUIRED")); err == nil && required {
		return psaAuditEnforce
	}
	return psaAuditOff
}

// podSecurityViolation is the structured result of a PSA audit for one
// namespace. It mirrors the output shape of the kubectl-dev_tool psa-check
// command from openshift/cluster-debug-tools, which this in-process audit
// replaces.
type podSecurityViolation struct {
	// Namespace where the violation happened.
	Namespace string `json:"namespace"`
	// Level is the pod security level that was violated.
	Level string `json:"level"`
	// PodName is the name of a pod that violates the level.
	PodName string `json:"podName"`
	// Violations lists the violations that pods in the namespace made.
	Violations []string `json:"violations"`
	// Labels contains the pod security labels present on the namespace.
	Labels map[string]string `json:"labels,omitempty"`
	// SyncControlLabel signals that the OpenShift label syncer is enabled for
	// this namespace.
	SyncControlLabel string `json:"syncControlLabel,omitempty"`
}

// summary renders a compact description of the violation for test failure
// messages. Only parsed, allowlisted fields are included: no raw API server
// output is propagated into logs or junit results.
func (v *podSecurityViolation) summary() string {
	var b strings.Builder
	fmt.Fprintf(&b, "pod %q violates PodSecurity level %q: %s", v.PodName, v.Level, strings.Join(v.Violations, ", "))
	if v.SyncControlLabel != "" {
		fmt.Fprintf(&b, " (podSecurityLabelSync=%q: the OpenShift label syncer may still be converging the namespace labels)", v.SyncControlLabel)
	}
	return b.String()
}

// warningCollector collects the Warning headers returned by the API server.
// It implements k8s.io/client-go/rest.WarningHandler. No locking is needed:
// client-go invokes the handler while processing the response of the request
// that owns the collector, and each audit performs a single request.
type warningCollector struct {
	warnings []string
}

// HandleWarningHeader implements rest.WarningHandler.
func (w *warningCollector) HandleWarningHeader(code int, agent string, text string) {
	if text == "" {
		return
	}
	w.warnings = append(w.warnings, text)
}

// popAll returns all collected warnings and resets the collector.
func (w *warningCollector) popAll() []string {
	warnings := w.warnings
	w.warnings = nil
	return warnings
}

// psaNamespaceViolationRegexp matches the namespace-level warning emitted by
// the API server's pod-security-admission plugin when existing pods violate a
// new enforce level, e.g.:
//
//	existing pods in namespace "ns" violate the new PodSecurity enforce level "restricted:latest"
var psaNamespaceViolationRegexp = regexp.MustCompile(`existing pods in namespace "([^"]+)" violate the new PodSecurity enforce level "([^"]+)"`)

// psaViolationReasons are the violation reasons the pod-security-admission
// plugin reports for individual pods, e.g.:
//
//	pod-name (and 2 other pods): allowPrivilegeEscalation != false, runAsNonRoot != true
var psaViolationReasons = map[string]struct{}{
	"allowPrivilegeEscalation != false": {},
	"forbidden AppArmor profile":        {},
	"forbidden AppArmor profiles":       {},
	"forbidden sysctls":                 {},
	"host namespaces":                   {},
	"hostPath volumes":                  {},
	"hostPort":                          {},
	"hostProcess":                       {},
	"non-default capabilities":          {},
	"privileged":                        {},
	"procMount":                         {},
	"restricted volume types":           {},
	"runAsNonRoot != true":              {},
	"runAsUser=0":                       {},
	"seccompProfile":                    {},
	"seLinuxOptions":                    {},
	"unrestricted capabilities":         {},
}

// parsePSAWarnings parses the warnings returned by the API server into a
// podSecurityViolation containing the namespace, level, pod name and
// violations. Ported from openshift/cluster-debug-tools (pkg/cmd/psa) so the
// in-process audit reports the same data as the kubectl-dev_tool psa-check
// command.
func parsePSAWarnings(warnings []string) *podSecurityViolation {
	if len(warnings) == 0 {
		return nil
	}

	var psv podSecurityViolation

	// There should be exactly 2 warnings, but lets not rely on the order nor
	// it to stay that way.
	for _, warning := range warnings {
		// Namespace warning message: existing pods in namespace "ns" violate
		// the new PodSecurity enforce level "level:version"
		if matches := psaNamespaceViolationRegexp.FindStringSubmatch(warning); len(matches) == 3 {
			psv.Namespace = matches[1]
			psv.Level = matches[2]
			continue
		}

		// Pod warning message: {pod name}{optional " (and N other pods)"}: {reason A}, {reason B}, ...
		textSplit := strings.Split(warning, ": ")
		if len(textSplit) != 2 {
			continue
		}

		matchingList := strings.Split(textSplit[1], ", ")
		matchedCandidates := []string{}
		for _, candidate := range matchingList {
			// It could be that there are new violations in the
			// pod-security-admission code.
			if _, ok := psaViolationReasons[candidate]; ok {
				matchedCandidates = append(matchedCandidates, candidate)
			}
		}

		if len(matchedCandidates) == 0 {
			continue
		}

		// Get rid of the potential " (and x other pods)" hint.
		psv.PodName = strings.Split(textSplit[0], " ")[0]
		psv.Violations = matchedCandidates
	}

	// Check if any properties of the psv struct have been set. If not, return
	// nil.
	if psv.Namespace == "" && psv.Level == "" && psv.PodName == "" && len(psv.Violations) == 0 {
		return nil
	}

	return &psv
}

// AuditAndTeardownProject audits the current test namespace against the
// restricted PodSecurity level before removing the resources created by the
// test.
//
// The audit runs in-process: the API server evaluates the namespace's existing
// pods during a server-side dry-run update that raises the namespace's
// enforce label, and reports violations through the response Warning headers
// (the same mechanism the kubectl-dev_tool psa-check command from
// openshift/cluster-debug-tools uses). Compiling the audit into the
// olmv0-tests-ext binary makes it available wherever the tests execute,
// without staging an external binary into the test environment (see
// openshift/release#84940 and openshift/release#87635).
func (c *CLI) AuditAndTeardownProject() {
	err := c.auditPSA()
	c.TeardownProject()
	if err != nil {
		g.Fail(err.Error())
	}
}

// auditPSA audits the CLI's namespace per the configured audit mode. In warn
// mode, infrastructure errors are logged and violations are recorded without
// failing the test; in enforce mode both surface as test failures.
func (c *CLI) auditPSA() error {
	mode := psaAuditMode()
	if mode == psaAuditOff {
		return nil
	}

	namespace := c.Namespace()
	if namespace == "" {
		e2e.Logf("Skipping PSA audit because the test has no namespace")
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), psaAuditTimeout)
	defer cancel()

	violation, err := c.checkNamespacePodSecurity(ctx, namespace)
	if err != nil {
		if mode == psaAuditWarn {
			e2e.Logf("PSA audit (warn mode) could not complete for namespace %q: %v", namespace, err)
			return nil
		}
		return fmt.Errorf("PSA audit failed for namespace %q: %w", namespace, err)
	}
	if violation == nil {
		return nil
	}

	if artifactPath := psaArtifactPath(namespace); artifactPath != "" {
		psaWriteArtifact(artifactPath, violation)
	}

	if mode == psaAuditWarn {
		e2e.Logf("PSA audit (warn mode) found violations in namespace %q: %s", namespace, violation.summary())
		return nil
	}
	return fmt.Errorf("PSA audit failed for namespace %q: %s", namespace, violation.summary())
}

// checkNamespacePodSecurity evaluates the namespace's existing pods against
// the restricted PodSecurity level by issuing a server-side dry-run update
// that raises the namespace's enforce label. The update is never persisted; it
// only triggers the API server's pod-security-admission evaluation, whose
// findings are returned as Warning headers on the response.
func (c *CLI) checkNamespacePodSecurity(ctx context.Context, namespace string) (*podSecurityViolation, error) {
	config, err := psaRestConfig(c.adminKubeConfigPath())
	if err != nil {
		return nil, fmt.Errorf("failed to create rest config: %w", err)
	}
	warnings := &warningCollector{}
	config.WarningHandler = warnings

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	ns, err := client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace %q: %w", namespace, err)
	}

	nsCopy := ns.DeepCopy()
	if nsCopy.Labels == nil {
		nsCopy.Labels = map[string]string{}
	}
	nsCopy.Labels[psapi.EnforceLevelLabel] = string(psaAuditLevel)

	if _, err := client.CoreV1().Namespaces().Update(ctx, nsCopy, metav1.UpdateOptions{DryRun: []string{metav1.DryRunAll}}); err != nil {
		return nil, fmt.Errorf("failed to dry-run update namespace %q with enforce level %q: %w", namespace, psaAuditLevel, err)
	}

	collected := warnings.popAll()
	if len(collected) == 0 {
		return nil, nil
	}

	violation := parsePSAWarnings(collected)
	if violation == nil {
		return nil, nil
	}

	// Take note of the pod security related labels and the label syncer
	// sentinel for context in the audit result.
	violation.Labels = psaLabels(ns)
	if syncControl, ok := ns.Labels[podSecurityLabelSyncControlLabel]; ok {
		violation.SyncControlLabel = syncControl
	}

	return violation, nil
}

// adminKubeConfigPath returns the kubeconfig the audit should use: the
// per-test temporary admin config when one is in use, the admin config path
// otherwise.
func (c *CLI) adminKubeConfigPath() string {
	if c.tempAdmConfPath != "" {
		return c.tempAdmConfPath
	}
	return c.adminConfigPath
}

// psaRestConfig loads a rest config from the given kubeconfig path, falling
// back to the default loading rules when no path is set (matching the
// behaviour of an external tool run without KUBECONFIG).
func psaRestConfig(path string) (*rest.Config, error) {
	if path != "" {
		return clientcmd.BuildConfigFromFlags("", path)
	}
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{}).ClientConfig()
}

// psaLabels records the pod security labels present on the namespace.
func psaLabels(ns *corev1.Namespace) map[string]string {
	var labels map[string]string
	for _, label := range []string{
		psapi.EnforceLevelLabel,
		psapi.AuditLevelLabel,
		psapi.WarnLevelLabel,
	} {
		if value, ok := ns.Labels[label]; ok {
			if labels == nil {
				labels = map[string]string{}
			}
			labels[label] = value
		}
	}
	return labels
}

// psaWriteArtifact writes the parsed violation as JSON next to the other test
// artifacts.
func psaWriteArtifact(artifactPath string, violation *podSecurityViolation) {
	data, err := json.MarshalIndent(violation, "", "  ")
	if err != nil {
		e2e.Logf("Failed to serialize PSA audit result: %v", err)
		return
	}
	if err := os.WriteFile(artifactPath, append(data, '\n'), 0o600); err != nil {
		e2e.Logf("Failed to write PSA audit result: %v", err)
	}
}

func psaArtifactPath(namespace string) string {
	basePath := strings.TrimSpace(os.Getenv("ARTIFACT_DIR"))
	if basePath == "" {
		return ""
	}

	dir := filepath.Join(basePath, "psa", sanitizePSAPath(namespace))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		e2e.Logf("Failed to create PSA audit artifact directory %q: %v", dir, err)
		return ""
	}
	return filepath.Join(dir, "psa.json")
}

func sanitizePSAPath(value string) string {
	var result strings.Builder
	for _, r := range strings.TrimSpace(value) {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			result.WriteRune(r)
		default:
			result.WriteByte('-')
		}
	}
	if result.Len() == 0 {
		return "unknown"
	}
	safe := strings.Trim(result.String(), "-")
	if safe == "" {
		return "unknown"
	}
	return safe
}
