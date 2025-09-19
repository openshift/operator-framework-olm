package util

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	o "github.com/onsi/gomega"
	"github.com/tidwall/pretty"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// ApplyClusterResourceFromTemplateWithError applies cluster resources from template and returns errors instead of failing
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - parameters: template processing parameters (e.g., "--ignore-unknown-parameters=true", "-f", "template_file")
//
// Returns:
//   - error: error if template processing or resource application fails, nil on success
func ApplyClusterResourceFromTemplateWithError(oc *CLI, parameters ...string) error {
	return resourceFromTemplate(oc, false, true, "", parameters...)
}

// ApplyClusterResourceFromTemplate applies cluster-scoped resources from template, fails test on error
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - parameters: template processing parameters (e.g., "--ignore-unknown-parameters=true", "-f", "template_file")
func ApplyClusterResourceFromTemplate(oc *CLI, parameters ...string) {
	err := resourceFromTemplate(oc, false, false, "", parameters...)

	o.Expect(err).ShouldNot(o.HaveOccurred())
}

// ApplyNsResourceFromTemplate applies namespace-scoped resources from template
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - namespace: target namespace for the resources (overrides template namespace if specified)
//   - parameters: template processing parameters (e.g., "--ignore-unknown-parameters=true", "-f", "template_file")
func ApplyNsResourceFromTemplate(oc *CLI, namespace string, parameters ...string) {
	err := resourceFromTemplate(oc, false, false, namespace, parameters...)

	o.Expect(err).ShouldNot(o.HaveOccurred())
}

// CreateClusterResourceFromTemplateWithError creates cluster resources from template and returns errors instead of failing
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - parameters: template processing parameters (e.g., "--ignore-unknown-parameters=true", "-f", "template_file")
//
// Returns:
//   - error: error if template processing or resource creation fails, nil on success
func CreateClusterResourceFromTemplateWithError(oc *CLI, parameters ...string) error {
	return resourceFromTemplate(oc, true, true, "", parameters...)
}

// CreateClusterResourceFromTemplate creates cluster-scoped resources from template, fails test on error
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - parameters: template processing parameters (e.g., "--ignore-unknown-parameters=true", "-f", "template_file")
func CreateClusterResourceFromTemplate(oc *CLI, parameters ...string) {
	err := resourceFromTemplate(oc, true, false, "", parameters...)
	o.Expect(err).ShouldNot(o.HaveOccurred())
}

// CreateNsResourceFromTemplate creates namespace-scoped resources from template
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - namespace: target namespace for the resources (overrides template namespace if specified)
//   - parameters: template processing parameters (e.g., "--ignore-unknown-parameters=true", "-f", "template_file")
func CreateNsResourceFromTemplate(oc *CLI, namespace string, parameters ...string) {
	err := resourceFromTemplate(oc, true, false, namespace, parameters...)
	o.Expect(err).ShouldNot(o.HaveOccurred())
}

// resourceFromTemplate is the internal implementation for template processing and resource management
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - create: true to use 'oc create', false to use 'oc apply'
//   - returnError: true to return errors instead of failing the test
//   - namespace: target namespace (empty string for cluster-scoped resources)
//   - parameters: template processing parameters
//
// Returns:
//   - error: error if processing fails and returnError is true, nil on success
func resourceFromTemplate(oc *CLI, create bool, returnError bool, namespace string, parameters ...string) error {
	if oc == nil {
		return fmt.Errorf("CLI client cannot be nil")
	}
	if len(parameters) == 0 {
		return fmt.Errorf("template parameters cannot be empty")
	}

	var configFile string
	err := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 15*time.Second, false, func(ctx context.Context) (bool, error) {
		fileName := GetRandomString() + "config.json"
		stdout, _, err := oc.AsAdmin().Run("process").Args(parameters...).OutputsToFiles(fileName)
		if err != nil {
			e2e.Logf("template processing failed (will retry): %v", err)
			return false, nil
		}

		// Validate that the output file exists and is readable
		if stdout == "" {
			e2e.Logf("template processing returned empty output file path")
			return false, nil
		}
		if _, err := os.Stat(stdout); err != nil {
			e2e.Logf("template output file does not exist or is not accessible: %v", err)
			return false, nil
		}

		configFile = stdout
		return true, nil
	})
	if returnError && err != nil {
		e2e.Logf("failed to process template parameters %v: %v", parameters, err)
		return fmt.Errorf("template processing failed: %w", err)
	}
	AssertWaitPollNoErr(err, fmt.Sprintf("failed to process template parameters %v", parameters))

	e2e.Logf("generated resource configuration file: %s", configFile)

	// Apply or create the resources with better error handling
	resourceErr := applyOrCreateResource(oc, create, namespace, configFile)
	if returnError && resourceErr != nil {
		e2e.Logf("failed to create/apply resource: %v", resourceErr)
		return fmt.Errorf("resource operation failed: %w", resourceErr)
	}
	AssertWaitPollNoErr(resourceErr, fmt.Sprintf("failed to create/apply resource from %s", configFile))
	return nil
}

// GetRandomString generates a cryptographically secure random 8-character hex string for unique identifiers
// Returns:
//   - string: 8-character random hex string using cryptographically secure randomness
func GetRandomString() string {
	bytes := make([]byte, 4) // 4 bytes = 8 hex characters
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based unique string if crypto/rand fails
		e2e.Logf("Warning: crypto/rand failed, using timestamp fallback: %v", err)
		return fmt.Sprintf("%08x", time.Now().UnixNano()&0xFFFFFFFF)
	}
	return hex.EncodeToString(bytes)
}

// applyOrCreateResource applies or creates resources from the given config file
func applyOrCreateResource(oc *CLI, create bool, namespace, configFile string) error {
	var args []string
	action := "apply"
	if create {
		action = "create"
	}

	args = []string{"-f", configFile}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	return oc.AsAdmin().WithoutNamespace().Run(action).Args(args...).Execute()
}

// ApplyResourceFromTemplateWithNonAdminUser processes template and applies resources using non-admin user privileges
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - parameters: template processing parameters (e.g., "-f", "template_file", "-p", "PARAM=value")
//
// Returns:
//   - error: error if template processing or resource application fails, nil on success
func ApplyResourceFromTemplateWithNonAdminUser(oc *CLI, parameters ...string) error {
	var configFile string
	err := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 15*time.Second, false, func(ctx context.Context) (bool, error) {
		output, err := oc.Run("process").Args(parameters...).OutputToFile(GetRandomString() + "config.json")
		if err != nil {
			e2e.Logf("the err:%v, and try next round", err)
			return false, nil
		}
		configFile = output
		return true, nil
	})
	AssertWaitPollNoErr(err, fmt.Sprintf("fail to process %v", parameters))

	e2e.Logf("the file of resource is %s", configFile)
	return oc.WithoutNamespace().Run("apply").Args("-f", configFile).Execute()
}

// ProcessTemplate processes a template with given parameters and returns the output file path
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - parameters: template processing parameters (e.g., "-f", "template_file", "-p", "PARAM=value")
//
// Returns:
//   - string: path to the generated configuration file
func ProcessTemplate(oc *CLI, parameters ...string) string {
	var configFile string

	err := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 15*time.Second, false, func(ctx context.Context) (bool, error) {
		output, err := oc.Run("process").Args(parameters...).OutputToFile(GetRandomString() + "config.json")
		if err != nil {
			e2e.Logf("the err:%v, and try next round", err)
			return false, nil
		}
		configFile = output
		return true, nil
	})

	AssertWaitPollNoErr(err, fmt.Sprintf("fail to process %v", parameters))
	e2e.Logf("the file of resource is %s", configFile)
	return configFile
}

// ParameterizedTemplateByReplaceToFile processes template by direct string replacement and saves to file
// This function manually replaces template parameters without using 'oc process'
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - parameters: template parameters including "-f template_file" and "-p PARAM=value" entries
//
// Returns:
//   - string: path to the generated JSON configuration file
func ParameterizedTemplateByReplaceToFile(oc *CLI, parameters ...string) string {
	if oc == nil {
		o.Expect(fmt.Errorf("CLI client cannot be nil")).ShouldNot(o.HaveOccurred())
	}
	if len(parameters) == 0 {
		o.Expect(fmt.Errorf("parameters cannot be empty")).ShouldNot(o.HaveOccurred())
	}

	// Find template file parameter
	isParameterExist, pIndex := StringsSliceElementsHasPrefix(parameters, "-f", true)
	o.Expect(isParameterExist).Should(o.BeTrue(), "Template file parameter (-f) is required")
	o.Expect(pIndex+1).Should(o.BeNumerically("<", len(parameters)), "Template file path is missing after -f flag")
	templateFileName := parameters[pIndex+1]

	// Validate and read template file
	templateFileName = filepath.Clean(templateFileName) // Prevent path traversal
	templateContentByte, readFileErr := os.ReadFile(templateFileName)
	o.Expect(readFileErr).ShouldNot(o.HaveOccurred(), "Failed to read template file: %s", templateFileName)
	templateContentStr := string(templateContentByte)

	// Find parameter substitution values
	isParameterExist, pIndex = StringsSliceElementsHasPrefix(parameters, "-p", true)
	o.Expect(isParameterExist).Should(o.BeTrue(), "Parameter substitution (-p) is required")

	// Process parameter substitutions with validation
	for i := pIndex + 1; i < len(parameters); i++ {
		if strings.Contains(parameters[i], "=") {
			tempSlice := strings.Split(parameters[i], "=")
			o.Expect(tempSlice).Should(o.HaveLen(2), "Parameter format should be KEY=VALUE")
			paramKey := strings.TrimSpace(tempSlice[0])
			paramValue := tempSlice[1] // Don't trim value as it might be intentionally contain spaces

			// Validate parameter key is not empty
			o.Expect(paramKey).ShouldNot(o.BeEmpty(), "Parameter key cannot be empty")

			// Replace template placeholder with value
			placeholder := "${" + paramKey + "}"
			templateContentStr = strings.ReplaceAll(templateContentStr, placeholder, paramValue)
		}
	}

	// Convert YAML to JSON with error handling
	templateContentJSON, convertErr := yaml.YAMLToJSON([]byte(templateContentStr))
	o.Expect(convertErr).NotTo(o.HaveOccurred(), "Failed to convert YAML template to JSON")

	// Generate secure output file path
	namespace := oc.Namespace()
	if namespace == "" {
		namespace = "default"
	}
	configFile := filepath.Join(e2e.TestContext.OutputDir, namespace+"-"+GetRandomString()+"config.json")
	// Use more restrictive file permissions (0600) for security
	o.Expect(os.WriteFile(configFile, pretty.Pretty(templateContentJSON), 0600)).ShouldNot(o.HaveOccurred(), "Failed to write configuration file")
	return configFile
}
