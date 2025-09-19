// Package olmv0util provides utilities for OLM v0 operator testing
// This file contains resource management utilities for tracking and cleaning up
// test resources in a structured way during operator testing
package olmv0util

import (
	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// resourceDescription represents a Kubernetes/OpenShift resource for cleanup tracking
// This structure stores all necessary information to identify and delete resources
// created during test execution
type resourceDescription struct {
	oc               *exutil.CLI // OpenShift CLI client for resource operations
	AsAdmin          bool        // Whether to use admin privileges for operations
	WithoutNamespace bool        // Whether to execute without namespace context
	kind             string      // Resource type (e.g., "pod", "service", "csv")
	name             string      // Resource name
	RequireNS        bool        // Whether the resource requires namespace specification
	namespace        string      // Namespace where the resource exists
}

// NewResource constructs a resource descriptor for test cleanup tracking
// This function creates a structured representation of a Kubernetes/OpenShift resource
// that can be later deleted during test cleanup procedures
//
// Parameters:
//   - oc: OpenShift CLI client for resource operations
//   - kind: Resource type (e.g., "pod", "service", "csv", "subscription")
//   - name: Name of the resource
//   - nsflag: Whether the resource requires namespace specification (true for namespaced resources)
//   - namespace: Namespace where the resource exists (empty for cluster-scoped resources)
//
// Returns:
//   - resourceDescription: Structured resource information for cleanup
func NewResource(oc *exutil.CLI, kind string, name string, nsflag bool, namespace string) resourceDescription {
	return resourceDescription{
		oc:               oc,
		AsAdmin:          exutil.AsAdmin,
		WithoutNamespace: exutil.WithoutNamespace,
		kind:             kind,
		name:             name,
		RequireNS:        nsflag,
		namespace:        namespace,
	}
}

// delete removes the resource from the cluster using the stored configuration
// This method handles both namespaced and cluster-scoped resource deletion
func (r resourceDescription) delete() {
	if r.WithoutNamespace && r.RequireNS {
		removeResource(r.oc, r.AsAdmin, r.WithoutNamespace, r.kind, r.name, "-n", r.namespace)
	} else {
		removeResource(r.oc, r.AsAdmin, r.WithoutNamespace, r.kind, r.name)
	}
}

// itResource maps resource identifiers to their descriptions for a single test iteration
// The key format is "name+kind+namespace" to ensure unique identification of resources
// This allows tracking of all resources created within a single Ginkgo "It" block
type itResource map[string]resourceDescription

// Add registers a resource for cleanup tracking within the test iteration
// The resource is indexed by a composite key to ensure uniqueness
func (ir itResource) Add(r resourceDescription) {
	ir[r.name+r.kind+r.namespace] = r
}

// get retrieves a resource description by its identifying components
// This method is currently unused but provides resource lookup capability
//
//nolint:unused
func (ir itResource) get(name string, kind string, namespace string) resourceDescription {
	r, ok := ir[name+kind+namespace]
	o.Expect(ok).To(o.BeTrue())
	return r
}

// Remove deletes a specific resource from the cluster and unregisters it from tracking
// This method performs the actual resource deletion and removes it from the cleanup list
func (ir itResource) Remove(name string, kind string, namespace string) {
	rKey := name + kind + namespace
	if r, ok := ir[rKey]; ok {
		r.delete()
		delete(ir, rKey)
	}
}

// Cleanup removes all tracked resources from the cluster
// This method is typically called during test teardown to ensure clean environment
func (ir itResource) Cleanup() {
	for _, r := range ir {
		e2e.Logf("cleanup resource %s,   %s", r.kind, r.name)
		ir.Remove(r.name, r.kind, r.namespace)
	}
}

// DescriberResrouce maps test iteration names to their resource collections
// This structure organizes resources by Ginkgo test iterations ("It" blocks)
// within a test suite ("Describe" block), enabling granular cleanup control
type DescriberResrouce map[string]itResource

// AddIr initializes a new resource collection for a test iteration
// This method prepares resource tracking for a new Ginkgo "It" block
func (dr DescriberResrouce) AddIr(itName string) {
	dr[itName] = itResource{}
}

// GetIr retrieves the resource collection for a specific test iteration
// This method provides access to all resources created within a named "It" block
func (dr DescriberResrouce) GetIr(itName string) itResource {
	ir, ok := dr[itName]
	if !ok {
		e2e.Logf("!!! couldn't find the itName:%s, print the DescriberResrouce:%v", itName, dr)
	}
	o.Expect(ok).To(o.BeTrue())
	return ir
}

// RmIr removes a test iteration's resource collection from tracking
// This method cleans up the resource tracking data structure after test completion
func (dr DescriberResrouce) RmIr(itName string) {
	delete(dr, itName)
}
