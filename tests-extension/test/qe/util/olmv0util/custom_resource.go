// Package olmv0util provides utilities for OLM v0 operator testing
// This file contains custom resource management utilities for creating
// and managing instances of custom resources defined by CRDs
package olmv0util

import (
	"context"
	"time"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// CustomResourceDescription represents a custom resource instance configuration
// Custom resources are instances of types defined by CustomResourceDefinitions (CRDs)
type CustomResourceDescription struct {
	Name      string // Name of the custom resource instance
	Namespace string // Namespace where the custom resource will be created
	TypeName  string // Type name of the custom resource (e.g., "etcdcluster")
	Template  string // Template file path for creating the custom resource
}

// Create creates a custom resource instance using the provided template
// This method applies the custom resource template with retry logic to handle
// potential timing issues with CRD availability
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
//   - itName: Test iteration name for resource tracking
//   - dr: Resource descriptor for test cleanup management
func (crinstance *CustomResourceDescription) Create(oc *exutil.CLI, itName string, dr DescriberResrouce) {
	// Retry custom resource creation to handle CRD availability timing
	errCR := wait.PollUntilContextTimeout(context.TODO(), 30*time.Second, 60*time.Second, false, func(ctx context.Context) (bool, error) {
		err := ApplyResourceFromTemplate(oc, "--ignore-unknown-parameters=true", "-f", crinstance.Template,
			"-p", "NAME="+crinstance.Name, "NAMESPACE="+crinstance.Namespace)
		if err != nil {
			e2e.Logf("met error: %v, try next round ...", err.Error())
			return false, nil
		}
		return true, nil
	})
	exutil.AssertWaitPollNoErr(errCR, "cr etcdcluster is not created")

	dr.GetIr(itName).Add(NewResource(oc, crinstance.TypeName, crinstance.Name, exutil.RequireNS, crinstance.Namespace))
	e2e.Logf("create CR %s SUCCESS", crinstance.Name)
}

// Delete removes the custom resource instance from the cluster
// This method unregisters the custom resource from the test cleanup framework
//
// Parameters:
//   - itName: Test iteration name for resource tracking
//   - dr: Resource descriptor for test cleanup management
func (crinstance *CustomResourceDescription) Delete(itName string, dr DescriberResrouce) {
	e2e.Logf("delete crinstance %s, type is %s, ns is %s", crinstance.Name, crinstance.TypeName, crinstance.Namespace)
	dr.GetIr(itName).Remove(crinstance.Name, crinstance.TypeName, crinstance.Namespace)
}
