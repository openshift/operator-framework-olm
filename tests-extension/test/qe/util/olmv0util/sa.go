// Package olmv0util provides utilities for OLM v0 operator testing
// This file contains ServiceAccount management utilities for configuring
// operator service accounts and testing RBAC permissions
package olmv0util

import (
	"context"
	"fmt"
	"strings"
	"time"

	g "github.com/onsi/ginkgo/v2"
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
// It waits for the ServiceAccount to be fully deleted before returning to avoid race conditions
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
func (sa *serviceAccountDescription) Delete(oc *exutil.CLI) {
	e2e.Logf("delete sa %s, ns is %s", sa.name, sa.namespace)
	_, err := exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "sa", sa.name, "-n", sa.namespace)
	o.Expect(err).NotTo(o.HaveOccurred())

	// Wait for SA to be actually deleted to avoid race conditions
	// Kubernetes deletion is asynchronous, so we need to poll until the resource is gone
	err = wait.PollUntilContextTimeout(context.TODO(), 500*time.Millisecond, 30*time.Second, true, func(ctx context.Context) (bool, error) {
		output, getErr := exutil.OcAction(oc, "get", exutil.AsAdmin, exutil.WithoutNamespace, "sa", sa.name, "-n", sa.namespace)
		if getErr != nil {
			// Check if error is due to resource not found (which means successfully deleted)
			// The error message or output will contain "not found" or "NotFound"
			errMsg := strings.ToLower(getErr.Error())
			outputMsg := strings.ToLower(output)
			if strings.Contains(errMsg, "not found") || strings.Contains(outputMsg, "not found") || strings.Contains(errMsg, "notfound") || strings.Contains(outputMsg, "notfound") {
				e2e.Logf("SA %s successfully deleted from namespace %s", sa.name, sa.namespace)
				return true, nil
			}
			// Other errors (network, permission, etc.) should be retried
			e2e.Logf("Error checking SA %s (will retry): %v", sa.name, getErr)
			return false, nil
		}
		e2e.Logf("Waiting for SA %s to be fully deleted from namespace %s...", sa.name, sa.namespace)
		return false, nil
	})
	if err != nil {
		g.Skip("skip because of sa not deleted")
	}
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
