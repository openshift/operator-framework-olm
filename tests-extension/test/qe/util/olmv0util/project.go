// Package olmv0util provides utilities for OLM v0 operator testing
// This file contains project (namespace) management utilities for creating
// and managing OpenShift projects for operator testing
package olmv0util

import (
	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// ProjectDescription represents an OpenShift project (namespace) configuration
// Projects provide namespace isolation and RBAC boundaries for operator testing
type ProjectDescription struct {
	Name            string // Name of the project to create or manage
	TargetNamespace string // Target namespace to switch to after operations
}

// CreateWithCheck checks for existing project and creates one if it doesn't exist
// This method ensures the project exists and switches to it for subsequent operations
// If the project already exists, it simply switches to it without modification
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
//   - itName: Test iteration name for resource tracking
//   - dr: Resource descriptor for test cleanup management
func (p *ProjectDescription) CreateWithCheck(oc *exutil.CLI, itName string, dr DescriberResrouce) {
	// Check if the project already exists
	output, err := exutil.OcAction(oc, "get", exutil.AsAdmin, exutil.WithoutNamespace, "project", p.Name)
	if err != nil {
		// Project doesn't exist, create it
		e2e.Logf("Output: %s, cannot find the %s project, create one", output, p.Name)
		_, err := exutil.OcAction(oc, "adm", exutil.AsAdmin, exutil.WithoutNamespace, "new-project", p.Name)
		o.Expect(err).NotTo(o.HaveOccurred())
		// Register project for cleanup
		dr.GetIr(itName).Add(NewResource(oc, "project", p.Name, exutil.NotRequireNS, ""))
		// Switch to the newly created project
		_, err = exutil.OcAction(oc, "project", exutil.AsAdmin, exutil.WithoutNamespace, p.Name)
		o.Expect(err).NotTo(o.HaveOccurred())

	} else {
		// Project already exists
		e2e.Logf("project: %s already exist!", p.Name)
	}
}

// Create forcibly recreates a project by deleting the existing one and creating a new one
// This method ensures a clean project state by removing any existing project with the same name
// After creation, it switches back to the target namespace
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
//   - itName: Test iteration name for resource tracking
//   - dr: Resource descriptor for test cleanup management
func (p *ProjectDescription) Create(oc *exutil.CLI, itName string, dr DescriberResrouce) {
	// Remove any existing project with the same name
	removeResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "project", p.Name)
	// Create new project without updating local config
	_, err := exutil.OcAction(oc, "new-project", exutil.AsAdmin, exutil.WithoutNamespace, p.Name, "--skip-config-write")
	o.Expect(err).NotTo(o.HaveOccurred())
	// Register project for cleanup
	dr.GetIr(itName).Add(NewResource(oc, "project", p.Name, exutil.NotRequireNS, ""))
	// Switch back to target namespace
	_, err = exutil.OcAction(oc, "project", exutil.AsAdmin, exutil.WithoutNamespace, p.TargetNamespace)
	o.Expect(err).NotTo(o.HaveOccurred())
}

// Label adds a label to the project namespace
// This method applies environment or category labels to projects for organization and selection
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
//   - label: Label value to apply (will be applied as env=<label>)
func (p *ProjectDescription) Label(oc *exutil.CLI, label string) {
	_, err := exutil.OcAction(oc, "label", exutil.AsAdmin, exutil.WithoutNamespace, "ns", p.Name, "env="+label)
	o.Expect(err).NotTo(o.HaveOccurred())
}

// Delete removes the project from the cluster
// This method performs a standard project deletion
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
func (p *ProjectDescription) Delete(oc *exutil.CLI) {
	_, err := exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "project", p.Name)
	o.Expect(err).NotTo(o.HaveOccurred())
}

// DeleteWithForce performs a forced deletion of the project and all its resources
// This method forcibly removes all resources in the project before deleting the project itself
// It bypasses finalizers and grace periods to ensure complete cleanup
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
func (p *ProjectDescription) DeleteWithForce(oc *exutil.CLI) {
	// Forcibly delete all standard resources in the project
	_, err := exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "all", "--all", "-n", p.Name, "--force", "--grace-period=0", "--wait=false")
	o.Expect(err).NotTo(o.HaveOccurred())
	// Forcibly delete all CSV resources (ClusterServiceVersions)
	_, err = exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "csv", "--all", "-n", p.Name, "--force", "--grace-period=0", "--wait=false")
	o.Expect(err).NotTo(o.HaveOccurred())
	// Forcibly delete the project itself
	_, err = exutil.OcAction(oc, "delete", exutil.AsAdmin, exutil.WithoutNamespace, "project", p.Name, "--force", "--grace-period=0", "--wait=false")
	o.Expect(err).NotTo(o.HaveOccurred())
}
