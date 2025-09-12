// Package olmv0util provides utilities for OLM v0 operator testing
// This file contains ServiceAccount management utilities for configuring
// operator service accounts and testing RBAC permissions
package olmv0util

import (
	"context"
	"fmt"
	"strings"
	"time"

	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// serviceAccountDescription represents a ServiceAccount resource configuration
// ServiceAccounts provide identity for processes running in pods and define RBAC permissions
type serviceAccountDescription struct {
	name           string // Name of the service account
	namespace      string // Namespace where the service account exists
	definitionfile string // Path to the exported service account definition file
}

// NewSa constructs a new ServiceAccount descriptor for testing
// This function creates a ServiceAccount configuration object for use in tests
//
// Parameters:
//   - name: Name of the service account
//   - namespace: Namespace where the service account will be created
//
// Returns:
//   - *serviceAccountDescription: ServiceAccount configuration object
func NewSa(name, namespace string) *serviceAccountDescription {
	return &serviceAccountDescription{
		name:           name,
		namespace:      namespace,
		definitionfile: "",
	}
}

// GetDefinition exports the ServiceAccount resource definition to a file
// This method retrieves the current ServiceAccount configuration and saves it
// for later reapplication or modification during tests
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
func (sa *serviceAccountDescription) GetDefinition(oc *exutil.CLI) {
	// Export the ServiceAccount to a JSON file
	parameters := []string{"sa", sa.name, "-n", sa.namespace, "-o=json"}
	definitionfile, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(parameters...).OutputToFile("sa-config.json")
	o.Expect(err).NotTo(o.HaveOccurred())
	e2e.Logf("getDefinition: definitionfile is %s", definitionfile)
	// Store the file path for later use
	sa.definitionfile = definitionfile
}

// Delete removes the ServiceAccount from the cluster
// This method performs cleanup of ServiceAccount resources created during tests
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
func (sa *serviceAccountDescription) Delete(oc *exutil.CLI) {
	e2e.Logf("delete sa %s, ns is %s", sa.name, sa.namespace)
	_, err := exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "sa", sa.name, "-n", sa.namespace)
	o.Expect(err).NotTo(o.HaveOccurred())
}

// Reapply recreates the ServiceAccount using the previously exported definition
// This method restores a ServiceAccount from its saved configuration file
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
func (sa *serviceAccountDescription) Reapply(oc *exutil.CLI) {
	err := oc.AsAdmin().WithoutNamespace().Run("apply").Args("-f", sa.definitionfile).Execute()
	o.Expect(err).NotTo(o.HaveOccurred())
}

// checkAuth verifies ServiceAccount permissions for specific resource operations
// This method tests whether the ServiceAccount has the expected permissions
// to perform operations on specified custom resources
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
//   - expected: Expected permission result (e.g., "yes", "no")
//   - cr: Custom resource type to test permissions for
//
//nolint:unused
func (sa *serviceAccountDescription) checkAuth(oc *exutil.CLI, expected string, cr string) {
	// Test ServiceAccount permissions with retry logic
	err := wait.PollUntilContextTimeout(context.TODO(), 20*time.Second, 420*time.Second, false, func(ctx context.Context) (bool, error) {
		// Impersonate the ServiceAccount and test permissions
		output, _ := exutil.OcAction(oc, "auth", exutil.AsAdmin, exutil.WithNamespace, "--as", fmt.Sprintf("system:serviceaccount:%s:%s", sa.namespace, sa.name), "can-i", "create", cr)
		e2e.Logf("the result of checkAuth:%v", output)
		if strings.Contains(output, expected) {
			return true, nil
		}
		return false, nil
	})
	exutil.AssertWaitPollNoErr(err, fmt.Sprintf("sa %s expects %s permssion to create %s, but no", sa.name, expected, cr))
}
