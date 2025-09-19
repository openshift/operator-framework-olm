// Package olmv0util provides utilities for OLM v0 operator testing
// This file contains RBAC (Role-Based Access Control) management utilities
// for creating and managing roles and role bindings in operator testing
package olmv0util

import (
	"encoding/json"
	"strings"

	o "github.com/onsi/gomega"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// RoleDescription represents a Role resource configuration
// Roles define sets of permissions within a specific namespace
type RoleDescription struct {
	Name      string // Name of the Role
	Namespace string // Namespace where the Role will be created
	Template  string // Template file path for creating the Role
}

// Create creates a Role resource using the provided template
// Roles define permissions for resources within a specific namespace
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
func (role *RoleDescription) Create(oc *exutil.CLI) {
	err := ApplyResourceFromTemplate(oc, "--ignore-unknown-parameters=true", "-f", role.Template,
		"-p", "NAME="+role.Name, "NAMESPACE="+role.Namespace)
	o.Expect(err).NotTo(o.HaveOccurred())
}

// NewRole constructs a new Role descriptor for RBAC testing
// This function creates a Role configuration object for use in tests
//
// Parameters:
//   - name: Name of the role
//   - namespace: Namespace where the role will be created
//
// Returns:
//   - *RoleDescription: Role configuration object
func NewRole(name string, namespace string) *RoleDescription {
	return &RoleDescription{
		Name:      name,
		Namespace: namespace,
	}
}

// Patch modifies the Role resource with the provided patch
// This method applies JSON merge patches to update Role permissions
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
//   - patch: JSON patch string to apply to the Role
func (role *RoleDescription) Patch(oc *exutil.CLI, patch string) {
	PatchResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "role", role.Name, "-n", role.Namespace, "--type", "merge", "-p", patch)
}

// GetRules retrieves the permission rules from the Role
// This method extracts the current RBAC rules defined in the Role
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
//
// Returns:
//   - string: JSON representation of the Role's permission rules
func (role *RoleDescription) GetRules(oc *exutil.CLI) string {
	return role.getRulesWithDelete(oc, "nodelete")
}

// getRulesWithDelete retrieves Role rules with optional API group filtering
// This method allows excluding specific API groups from the returned rules
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
//   - delete: API group to exclude from rules ("nodelete" to include all)
//
// Returns:
//   - string: JSON representation of filtered permission rules
func (role *RoleDescription) getRulesWithDelete(oc *exutil.CLI, delete string) string {
	var roleboday map[string]interface{}
	output := GetResource(oc, exutil.AsAdmin, exutil.WithoutNamespace, "role", role.Name, "-n", role.Namespace, "-o=json")
	err := json.Unmarshal([]byte(output), &roleboday)
	o.Expect(err).NotTo(o.HaveOccurred())
	rules := roleboday["rules"].([]interface{})

	handleRuleAttribute := func(rc *strings.Builder, rt string, r map[string]interface{}) {
		rc.WriteString("\"" + rt + "\":[")
		items := r[rt].([]interface{})
		e2e.Logf("%s:%v, and the len:%v", rt, items, len(items))
		for i, v := range items {
			vc := v.(string)
			rc.WriteString("\"" + vc + "\"")
			if i != len(items)-1 {
				rc.WriteString(",")
			}
		}
		rc.WriteString("]")
		if strings.Compare(rt, "verbs") != 0 {
			rc.WriteString(",")
		}
	}

	var rc strings.Builder
	rc.WriteString("[")
	for _, rv := range rules {
		rule := rv.(map[string]interface{})
		if strings.Compare(delete, "nodelete") != 0 && strings.Compare(rule["apiGroups"].([]interface{})[0].(string), delete) == 0 {
			continue
		}

		rc.WriteString("{")
		handleRuleAttribute(&rc, "apiGroups", rule)
		handleRuleAttribute(&rc, "resources", rule)
		handleRuleAttribute(&rc, "verbs", rule)
		rc.WriteString("},")
	}
	result := strings.TrimSuffix(rc.String(), ",") + "]"
	e2e.Logf("rc:%v", result)
	return result
}

// RolebindingDescription represents a RoleBinding resource configuration
// RoleBindings associate Roles with subjects (users, groups, or service accounts)
type RolebindingDescription struct {
	Name      string // Name of the RoleBinding
	Namespace string // Namespace where the RoleBinding will be created
	Rolename  string // Name of the Role to bind
	Saname    string // Name of the ServiceAccount to bind to the Role
	Template  string // Template file path for creating the RoleBinding
}

// Create creates a RoleBinding resource using the provided template
// RoleBindings grant the permissions defined in a Role to specific subjects
//
// Parameters:
//   - oc: OpenShift CLI client for executing commands
func (rolebinding *RolebindingDescription) Create(oc *exutil.CLI) {
	err := ApplyResourceFromTemplate(oc, "--ignore-unknown-parameters=true", "-f", rolebinding.Template,
		"-p", "NAME="+rolebinding.Name, "NAMESPACE="+rolebinding.Namespace, "SA_NAME="+rolebinding.Saname, "ROLE_NAME="+rolebinding.Rolename)
	o.Expect(err).NotTo(o.HaveOccurred())
}
