// Package olmv0util provides utilities for OLM v0 operator testing
// This file contains ConfigMap management utilities for creating and managing
// configuration data for operator testing scenarios
package olmv0util

import (
	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// ConfigMapDescription represents a ConfigMap resource configuration
// ConfigMaps store configuration data that can be consumed by operators and applications
type ConfigMapDescription struct {
	Name      string // Name of the ConfigMap
	Namespace string // Namespace where the ConfigMap will be created
	Template  string // Template file path for creating the ConfigMap
}

// Create creates a ConfigMap resource using the provided template
// This method applies the ConfigMap template and registers the resource for cleanup
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
//   - itName: Test iteration name for resource tracking
//   - dr: Resource descriptor for test cleanup management
func (cm *ConfigMapDescription) Create(oc *exutil.CLI, itName string, dr DescriberResrouce) {
	err := ApplyResourceFromTemplate(oc, "--ignore-unknown-parameters=true", "-f", cm.Template, "-p", "NAME="+cm.Name, "NAMESPACE="+cm.Namespace)
	o.Expect(err).NotTo(o.HaveOccurred())
	dr.GetIr(itName).Add(NewResource(oc, "cm", cm.Name, exutil.RequireNS, cm.Namespace))
	e2e.Logf("create cm %s SUCCESS", cm.Name)
}

// Patch modifies the ConfigMap resource with the provided patch
// This method applies JSON merge patches to update ConfigMap data or metadata
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
//   - patch: JSON patch string to apply to the ConfigMap
func (cm *ConfigMapDescription) Patch(oc *exutil.CLI, patch string) {
	PatchResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "cm", cm.Name, "-n", cm.Namespace, "--type", "merge", "-p", patch)
}

// Delete removes the ConfigMap resource from the cluster
// This method unregisters the ConfigMap from the test cleanup framework
//
// Parameters:
//   - itName: Test iteration name for resource tracking
//   - dr: Resource descriptor for test cleanup management
func (cm *ConfigMapDescription) Delete(itName string, dr DescriberResrouce) {
	e2e.Logf("remove cm %s, ns is %v", cm.Name, cm.Namespace)
	dr.GetIr(itName).Remove(cm.Name, "cm", cm.Namespace)
}
