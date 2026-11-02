package util

import (
	"reflect"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	psapi "k8s.io/pod-security-admission/api"
)

func TestSanitizePSAPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "namespace", input: "ns/example-123", want: "ns-example-123"},
		{name: "empty", input: "", want: "unknown"},
		{name: "safe characters", input: "namespace_1.test", want: "namespace_1.test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizePSAPath(tt.input); got != tt.want {
				t.Fatalf("sanitizePSAPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestParsePSAWarnings is ported from openshift/cluster-debug-tools
// (pkg/cmd/psa/parser_test.go) to keep the in-process audit's parsing
// behaviour aligned with the kubectl-dev_tool psa-check command, extended to
// cover the other-violating-pods count the parser captures for the build-pod
// exemption.
func TestParsePSAWarnings(t *testing.T) {
	tests := []struct {
		name     string
		warnings []string
		expected *podSecurityViolation
	}{
		{
			name:     "empty warnings",
			warnings: []string{},
			expected: nil,
		},
		{
			name: "invalid warnings",
			warnings: []string{
				`this is some warning that is not related to pod security admission`,
				`this is some more detail on that unknown warning`,
			},
			expected: nil,
		},
		{
			name: "valid warnings",
			warnings: []string{
				`existing pods in namespace "p0t-sekurity-namespace" violate the new PodSecurity enforce level "restricted:latest"`,
				`p0t-sekurity-pod (and 11 other pods): allowPrivilegeEscalation != false, unrestricted capabilities, runAsNonRoot != true, seccompProfile`,
			},
			expected: &podSecurityViolation{
				Namespace:          "p0t-sekurity-namespace",
				Level:              "restricted:latest",
				PodName:            "p0t-sekurity-pod",
				Violations:         []string{"allowPrivilegeEscalation != false", "unrestricted capabilities", "runAsNonRoot != true", "seccompProfile"},
				OtherViolatingPods: 11,
			},
		},
		{
			name: "valid warnings with a singular other-pod hint",
			warnings: []string{
				`existing pods in namespace "p0t-sekurity-namespace" violate the new PodSecurity enforce level "restricted:latest"`,
				`p0t-sekurity-pod (and 1 other pod): privileged, runAsUser=0`,
			},
			expected: &podSecurityViolation{
				Namespace:          "p0t-sekurity-namespace",
				Level:              "restricted:latest",
				PodName:            "p0t-sekurity-pod",
				Violations:         []string{"privileged", "runAsUser=0"},
				OtherViolatingPods: 1,
			},
		},
		{
			name: "valid warnings without the other-pods hint",
			warnings: []string{
				`existing pods in namespace "p0t-sekurity-namespace" violate the new PodSecurity enforce level "restricted:latest"`,
				`p0t-sekurity-pod: allowPrivilegeEscalation != false, runAsNonRoot != true`,
			},
			expected: &podSecurityViolation{
				Namespace:  "p0t-sekurity-namespace",
				Level:      "restricted:latest",
				PodName:    "p0t-sekurity-pod",
				Violations: []string{"allowPrivilegeEscalation != false", "runAsNonRoot != true"},
			},
		},
		{
			name: "valid warnings, mixed with invalid warnings",
			warnings: []string{
				`this is some warning that is not related to pod security admission`,
				`this is some more detail on that unknown warning`,
				`existing pods in namespace "p0t-sekurity-namespace" violate the new PodSecurity enforce level "restricted:latest"`,
				`p0t-sekurity-pod (and 11 other pods): forbidden AppArmor profiles, forbidden sysctls, host namespaces, hostPath volumes, hostPort, hostProcess, non-default capabilities, privileged, procMount, restricted volume types, runAsUser=0, seLinuxOptions`,
				`this is some more warning that is not related to pod security admission`,
				`this is some even more detail on that unknown warning`,
			},
			expected: &podSecurityViolation{
				Namespace:          "p0t-sekurity-namespace",
				Level:              "restricted:latest",
				PodName:            "p0t-sekurity-pod",
				Violations:         []string{"forbidden AppArmor profiles", "forbidden sysctls", "host namespaces", "hostPath volumes", "hostPort", "hostProcess", "non-default capabilities", "privileged", "procMount", "restricted volume types", "runAsUser=0", "seLinuxOptions"},
				OtherViolatingPods: 11,
			},
		},
		{
			name: "pod warning with unknown violation reasons only is ignored",
			warnings: []string{
				`existing pods in namespace "p0t-sekurity-namespace" violate the new PodSecurity enforce level "restricted:latest"`,
				`p0t-sekurity-pod: some future violation reason`,
			},
			expected: &podSecurityViolation{
				Namespace: "p0t-sekurity-namespace",
				Level:     "restricted:latest",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := parsePSAWarnings(tc.warnings)
			if tc.expected == nil && actual != nil {
				t.Errorf("expected: %+v, actual: %+v", tc.expected, actual)
			}

			if tc.expected != nil && !reflect.DeepEqual(actual, tc.expected) {
				t.Errorf("expected: %v, actual: %v", tc.expected, actual)
			}
		})
	}
}

func TestPSAAuditMode(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want string
	}{
		{name: "unset", env: map[string]string{}, want: psaAuditOff},
		{name: "legacy required", env: map[string]string{"PSA_CHECK_REQUIRED": "true"}, want: psaAuditEnforce},
		{name: "legacy not required", env: map[string]string{"PSA_CHECK_REQUIRED": "false"}, want: psaAuditOff},
		{name: "legacy garbage value", env: map[string]string{"PSA_CHECK_REQUIRED": "garbage"}, want: psaAuditOff},
		{name: "mode off", env: map[string]string{"PSA_AUDIT_MODE": "off"}, want: psaAuditOff},
		{name: "mode warn", env: map[string]string{"PSA_AUDIT_MODE": "warn"}, want: psaAuditWarn},
		{name: "mode enforce", env: map[string]string{"PSA_AUDIT_MODE": "enforce"}, want: psaAuditEnforce},
		{name: "mode is case insensitive", env: map[string]string{"PSA_AUDIT_MODE": "WARN"}, want: psaAuditWarn},
		{name: "mode is trimmed", env: map[string]string{"PSA_AUDIT_MODE": " enforce "}, want: psaAuditEnforce},
		{name: "mode overrides legacy required", env: map[string]string{"PSA_AUDIT_MODE": "warn", "PSA_CHECK_REQUIRED": "true"}, want: psaAuditWarn},
		{name: "unknown mode fails closed", env: map[string]string{"PSA_AUDIT_MODE": "yes please"}, want: psaAuditEnforce},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for key := range tc.env {
				t.Setenv(key, tc.env[key])
			}
			if got := psaAuditMode(); got != tc.want {
				t.Fatalf("psaAuditMode() = %q, want %q (env: %v)", got, tc.want, tc.env)
			}
		})
	}
}

func TestPSAViolationSummary(t *testing.T) {
	violation := &podSecurityViolation{
		Namespace:  "olmv0-e2e",
		Level:      "restricted:latest",
		PodName:    "catalog-operator-abc",
		Violations: []string{"allowPrivilegeEscalation != false", "runAsNonRoot != true"},
	}
	summary := violation.summary()
	for _, want := range []string{
		`pod "catalog-operator-abc" violates PodSecurity level "restricted:latest"`,
		"allowPrivilegeEscalation != false, runAsNonRoot != true",
	} {
		if !strings.Contains(summary, want) {
			t.Errorf("summary %q does not contain %q", summary, want)
		}
	}
	if strings.Contains(summary, "other pods") {
		t.Errorf("summary %q should not mention other pods when none are reported", summary)
	}
	if strings.Contains(summary, "podSecurityLabelSync") {
		t.Errorf("summary %q should not mention the label syncer when unset", summary)
	}

	violation.SyncControlLabel = "true"
	if summary := violation.summary(); !strings.Contains(summary, "podSecurityLabelSync") {
		t.Errorf("summary %q should mention the label syncer when set", summary)
	}

	violation.OtherViolatingPods = 2
	if summary := violation.summary(); !strings.Contains(summary, "2 other pods also violate") {
		t.Errorf("summary %q should mention other violating pods", summary)
	}
}

func TestPSALabels(t *testing.T) {
	ns := &corev1.Namespace{}
	if got := psaLabels(ns); got != nil {
		t.Errorf("psaLabels() = %v, want nil for a namespace without PSA labels", got)
	}

	ns.Labels = map[string]string{
		psapi.EnforceLevelLabel: "restricted",
		psapi.WarnLevelLabel:    "audit",
		"unrelated":             "value",
	}
	got := psaLabels(ns)
	want := map[string]string{
		psapi.EnforceLevelLabel: "restricted",
		psapi.WarnLevelLabel:    "audit",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("psaLabels() = %v, want %v", got, want)
	}
}

func TestIsBuildPodName(t *testing.T) {
	tests := []struct {
		name    string
		podName string
		want    bool
	}{
		{name: "build pod", podName: "custom-schema-3d2a6f2c-1-build", want: true},
		{name: "plain build suffix", podName: "my-build-42-build", want: true},
		{name: "catalog pod", podName: "custom-schema-3d2a6f2c-jfb9w", want: false},
		{name: "no suffix", podName: "catalog-operator-abc", want: false},
		{name: "empty", podName: "", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isBuildPodName(tc.podName); got != tc.want {
				t.Fatalf("isBuildPodName(%q) = %t, want %t", tc.podName, got, tc.want)
			}
		})
	}
}

func TestSoleBuildPodViolation(t *testing.T) {
	tests := []struct {
		name      string
		violation podSecurityViolation
		want      bool
	}{
		{
			name: "single build pod violation",
			violation: podSecurityViolation{
				PodName: "custom-schema-3d2a6f2c-1-build",
			},
			want: true,
		},
		{
			name: "build pod with other violating pods",
			violation: podSecurityViolation{
				PodName:            "custom-schema-3d2a6f2c-1-build",
				OtherViolatingPods: 2,
			},
			want: false,
		},
		{
			name: "non-build pod",
			violation: podSecurityViolation{
				PodName: "custom-schema-3d2a6f2c-jfb9w",
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.violation.soleBuildPodViolation(); got != tc.want {
				t.Fatalf("soleBuildPodViolation() = %t, want %t", got, tc.want)
			}
		})
	}
}
