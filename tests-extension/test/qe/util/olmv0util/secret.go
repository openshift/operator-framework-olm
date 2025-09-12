// Package olmv0util provides utilities for OLM v0 operator testing
// This file contains Secret management utilities for creating and managing
// sensitive data storage for operator testing scenarios
package olmv0util

import (
	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
)

// SecretDescription represents a Secret resource configuration
// Secrets store sensitive data such as passwords, tokens, and certificates
type SecretDescription struct {
	Name      string // Name of the Secret
	Namespace string // Namespace where the Secret will be created
	Saname    string // ServiceAccount name associated with the Secret
	Sectype   string // Type of Secret (e.g., "Opaque", "kubernetes.io/service-account-token")
	Template  string // Template file path for creating the Secret
}

// Create creates a Secret resource using the provided template
// This method applies the Secret template with the configured parameters
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
func (secret *SecretDescription) Create(oc *exutil.CLI) {
	err := ApplyResourceFromTemplate(oc, "--ignore-unknown-parameters=true", "-f", secret.Template,
		"-p", "NAME="+secret.Name, "NAMESPACE="+secret.Namespace, "SANAME="+secret.Saname, "TYPE="+secret.Sectype)
	o.Expect(err).NotTo(o.HaveOccurred())
}
