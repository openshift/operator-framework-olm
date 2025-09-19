// Package olmv0util provides utilities for OLM v0 operator testing
// This file contains CustomResourceDefinition (CRD) management utilities
// for creating and managing custom resource schemas
package olmv0util

import (
	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// CrdDescription represents a CustomResourceDefinition resource configuration
// CRDs define the schema and validation rules for custom resources used by operators
type CrdDescription struct {
	Name     string // Name of the CustomResourceDefinition
	Template string // Template file path for creating the CRD
}

// Create creates a CustomResourceDefinition using the provided template
// CRDs are cluster-scoped resources that define new API types for the cluster
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
//   - itName: Test iteration name for resource tracking
//   - dr: Resource descriptor for test cleanup management
func (crd *CrdDescription) Create(oc *exutil.CLI, itName string, dr DescriberResrouce) {
	err := ApplyResourceFromTemplate(oc, "--ignore-unknown-parameters=true", "-f", crd.Template, "-p", "NAME="+crd.Name)
	o.Expect(err).NotTo(o.HaveOccurred())
	dr.GetIr(itName).Add(NewResource(oc, "crd", crd.Name, exutil.NotRequireNS, ""))
	e2e.Logf("create crd %s SUCCESS", crd.Name)
}

// Delete removes the CustomResourceDefinition from the cluster
// This method performs direct deletion without using the resource tracking framework
// Note: CRD deletion also removes all custom resource instances of that type
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
func (crd *CrdDescription) Delete(oc *exutil.CLI) {
	e2e.Logf("remove crd %s, WithoutNamespace is %v", crd.Name, exutil.WithoutNamespace)
	removeResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "crd", crd.Name)
}
