// Package olmv0util provides utilities for OLM v0 operator testing
// This file contains OperatorGroup management utilities for configuring
// operator installation scopes and permissions
package olmv0util

import (
	"strings"

	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// OperatorGroupDescription represents an OperatorGroup resource configuration
// OperatorGroups define the scope and permissions for operator installations,
// controlling which namespaces operators can manage
type OperatorGroupDescription struct {
	Name               string // Name of the OperatorGroup
	Namespace          string // Namespace where the OperatorGroup will be created
	Multinslabel       string // Label selector for multi-namespace targeting
	Template           string // Template file path for creating the OperatorGroup
	ServiceAccountName string // Service account for operator permissions
	UpgradeStrategy    string // Strategy for operator upgrades
	ClusterType        string // Target cluster type (e.g., "microshift")
}

// CreateWithCheck checks for existing OperatorGroup and creates one if none exists
// This method ensures that exactly one OperatorGroup exists in the namespace,
// which is required for operator installations to function properly
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
//   - itName: Test iteration name for resource tracking
//   - dr: Resource descriptor for test cleanup management
func (og *OperatorGroupDescription) CreateWithCheck(oc *exutil.CLI, itName string, dr DescriberResrouce) {
	// Check if any OperatorGroup already exists in the current namespace
	output, err := exutil.OcAction(oc, "get", exutil.AsAdmin, false, "operatorgroup")
	o.Expect(err).NotTo(o.HaveOccurred())
	if strings.Contains(output, "No resources found") {
		// No OperatorGroup exists, create a new one
		e2e.Logf("No operatorgroup in project: %s, create one: %s", og.Namespace, og.Name)
		og.Create(oc, itName, dr)
	} else {
		// OperatorGroup already exists, skip creation
		e2e.Logf("Already exist operatorgroup in project: %s", og.Namespace)
	}

}

// Create creates an OperatorGroup resource using the provided template and configuration
// The method supports different OperatorGroup types based on the configuration:
// - Single namespace (ownnamespace): Operators manage only their installation namespace
// - Multi namespace: Operators manage multiple specified namespaces
// - All namespaces: Operators manage cluster-wide resources
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
//   - itName: Test iteration name for resource tracking
//   - dr: Resource descriptor for test cleanup management
func (og *OperatorGroupDescription) Create(oc *exutil.CLI, itName string, dr DescriberResrouce) {
	var err error
	// Choose appropriate template application function based on cluster type
	applyFn := ApplyResourceFromTemplate
	if strings.Compare(og.ClusterType, "microshift") == 0 {
		applyFn = ApplyResourceFromTemplateOnMicroshift
	}

	// Apply template with different parameter combinations based on configuration
	if strings.Compare(og.Multinslabel, "") != 0 && strings.Compare(og.ServiceAccountName, "") != 0 {
		// Multi-namespace with custom service account
		err = applyFn(oc, "--ignore-unknown-parameters=true", "-f", og.Template, "-p", "NAME="+og.Name, "NAMESPACE="+og.Namespace, "MULTINSLABEL="+og.Multinslabel, "SERVICE_ACCOUNT_NAME="+og.ServiceAccountName)
	} else if strings.Compare(og.Multinslabel, "") == 0 && strings.Compare(og.ServiceAccountName, "") == 0 && strings.Compare(og.UpgradeStrategy, "") == 0 {
		// Basic OperatorGroup with minimal configuration
		err = applyFn(oc, "--ignore-unknown-parameters=true", "-f", og.Template, "-p", "NAME="+og.Name, "NAMESPACE="+og.Namespace)
	} else if strings.Compare(og.Multinslabel, "") != 0 {
		// Multi-namespace configuration
		err = applyFn(oc, "--ignore-unknown-parameters=true", "-f", og.Template, "-p", "NAME="+og.Name, "NAMESPACE="+og.Namespace, "MULTINSLABEL="+og.Multinslabel)
	} else if strings.Compare(og.UpgradeStrategy, "") != 0 {
		// OperatorGroup with custom upgrade strategy
		err = applyFn(oc, "--ignore-unknown-parameters=true", "-f", og.Template, "-p", "NAME="+og.Name, "NAMESPACE="+og.Namespace, "UPGRADESTRATEGY="+og.UpgradeStrategy)
	} else {
		// OperatorGroup with custom service account
		err = applyFn(oc, "--ignore-unknown-parameters=true", "-f", og.Template, "-p", "NAME="+og.Name, "NAMESPACE="+og.Namespace, "SERVICE_ACCOUNT_NAME="+og.ServiceAccountName)
	}
	o.Expect(err).NotTo(o.HaveOccurred())
	// Register the created resource for test cleanup
	dr.GetIr(itName).Add(NewResource(oc, "og", og.Name, exutil.RequireNS, og.Namespace))
	e2e.Logf("create og %s success", og.Name)
}

// Delete removes the OperatorGroup resource from the cluster
// This method unregisters the OperatorGroup from the test cleanup framework
//
// Parameters:
//   - itName: Test iteration name for resource tracking
//   - dr: Resource descriptor for test cleanup management
func (og *OperatorGroupDescription) Delete(itName string, dr DescriberResrouce) {
	e2e.Logf("delete og %s, ns is %s", og.Name, og.Namespace)
	dr.GetIr(itName).Remove(og.Name, "og", og.Namespace)
}
