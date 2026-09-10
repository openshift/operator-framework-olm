package util

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	g "github.com/onsi/ginkgo/v2"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// AuditAndTeardownProject audits the current test namespace before removing
// the resources created by the test. PSA_CHECK_BIN is supplied by the release
// step-registry integration.
func (c *CLI) AuditAndTeardownProject() {
	err := c.auditPSA()
	c.TeardownProject()
	if err != nil {
		g.Fail(err.Error())
	}
}

func (c *CLI) auditPSA() error {
	bin := strings.TrimSpace(os.Getenv("PSA_CHECK_BIN"))
	if bin == "" {
		return nil
	}

	namespace := c.Namespace()
	if namespace == "" {
		e2e.Logf("Skipping PSA audit because the test has no namespace")
		return nil
	}

	if _, err := exec.LookPath(bin); err != nil {
		if psaCheckRequired() {
			return fmt.Errorf("PSA checker %q is unavailable: %w", bin, err)
		}
		e2e.Logf("Skipping PSA audit because checker %q is unavailable: %v", bin, err)
		return nil
	}

	cmd := exec.CommandContext(context.Background(), bin,
		"psa-check",
		"--namespace", namespace,
		"--level", "restricted",
		"--output", "json",
	)
	configPath := c.adminConfigPath
	if c.tempAdmConfPath != "" {
		configPath = c.tempAdmConfPath
	}
	cmd.Env = append(os.Environ(), fmt.Sprintf("KUBECONFIG=%s", configPath))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	if artifactPath := psaArtifactPath(namespace); artifactPath != "" {
		if writeErr := os.WriteFile(artifactPath, stdout.Bytes(), 0o600); writeErr != nil {
			e2e.Logf("Failed to write PSA audit result: %v", writeErr)
		}
		if stderr.Len() > 0 {
			stderrPath := strings.TrimSuffix(artifactPath, filepath.Ext(artifactPath)) + ".stderr"
			if writeErr := os.WriteFile(stderrPath, stderr.Bytes(), 0o600); writeErr != nil {
				e2e.Logf("Failed to write PSA audit diagnostics: %v", writeErr)
			}
		}
	}

	if err == nil {
		return nil
	}

	message := strings.TrimSpace(stdout.String())
	if stderrMessage := strings.TrimSpace(stderr.String()); stderrMessage != "" {
		if message != "" {
			message += "; "
		}
		message += stderrMessage
	}
	if message == "" {
		message = "no diagnostic output"
	}
	return fmt.Errorf("PSA audit failed for namespace %q: %w: %s", namespace, err, message)
}

func psaCheckRequired() bool {
	required, err := strconv.ParseBool(os.Getenv("PSA_CHECK_REQUIRED"))
	return err == nil && required
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
