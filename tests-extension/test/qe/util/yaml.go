// Package util provides utility functions for OpenShift testing,
// specifically for YAML file manipulation and modification.
package util

import (
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// YamlReplace defines a YAML modification instruction.
// It specifies a dot-separated path and the value to set at that path.
//
// Example:
//
//	YamlReplace{
//	  Path:  "spec.template.spec.imagePullSecrets",
//	  Value: "- name: notmatch-secret",
//	}
type YamlReplace struct {
	// Path is the dot-separated path to the YAML node to modify or create.
	// For example: "spec.template.spec.containers.0.image"
	Path string

	// Value is the string representation of the value to set at the specified path.
	// It can be a scalar value or a YAML fragment (e.g., "name: frontend").
	Value string
}

// ModifyYamlFileContent modifies a YAML file in-place by applying a series of replacements.
// It reads the file, applies all the specified modifications, and writes the result back.
//
// Parameters:
//   - file: path to the YAML file to modify
//   - replacements: slice of YamlReplace instructions to apply
//
// Example:
//
//	ModifyYamlFileContent("deployment.yaml", []YamlReplace{
//	  {
//	    Path:  "spec.template.spec.imagePullSecrets",
//	    Value: "- name: notmatch-secret",
//	  },
//	})
//
// The function will fail the test if any file operations or YAML parsing fails.
func ModifyYamlFileContent(file string, replacements []YamlReplace) {
	// Read the YAML file content
	input, err := os.ReadFile(file)
	if err != nil {
		e2e.Failf("read file %s failed: %v", file, err)
	}

	// Parse the YAML content into a node tree
	var doc yaml.Node
	if err = yaml.Unmarshal(input, &doc); err != nil {
		e2e.Failf("unmarshal yaml for file %s failed: %v", file, err)
	}

	// Apply each replacement to the document
	for _, replacement := range replacements {
		// Split the dot-separated path into components
		path := strings.Split(replacement.Path, ".")

		// Create a YAML node for the replacement value
		value := yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: replacement.Value,
		}

		// Apply the modification to the document
		setYamlValue(&doc, path, value)
	}

	// Marshal the modified document back to YAML
	output, err := yaml.Marshal(doc.Content[0])
	if err != nil {
		e2e.Failf("marshal yaml for file %s failed: %v", file, err)
	}

	// Write the modified content back to the file
	if err = os.WriteFile(file, output, 0o644); err != nil {
		e2e.Failf("write file %s failed: %v", file, err)
	}
}

// setYamlValue sets or creates a YAML node value at the specified path.
// It recursively traverses the YAML node tree, creating intermediate nodes as needed.
//
// Parameters:
//   - root: the root YAML node to start from
//   - path: slice of path components (e.g., ["spec", "template", "spec", "containers", "0"])
//   - value: the YAML node value to set at the target path
//
// The function handles different YAML node types:
//   - DocumentNode: delegates to the document's content
//   - MappingNode: searches for keys and creates new key-value pairs if needed
//   - SequenceNode: accesses elements by numeric index
func setYamlValue(root *yaml.Node, path []string, value yaml.Node) {
	// Base case: we've reached the target location
	if len(path) == 0 {
		// Try to parse the value as YAML first (for complex values)
		var valueParsed yaml.Node
		if err := yaml.Unmarshal([]byte(value.Value), &valueParsed); err == nil {
			// Successfully parsed as YAML, use the parsed structure
			*root = *valueParsed.Content[0]
		} else {
			// Failed to parse as YAML, treat as scalar value
			*root = value
		}
		return
	}
	// Extract the current path component and remaining path
	key := path[0]
	rest := path[1:]

	// Handle different node types appropriately
	switch root.Kind {
	case yaml.DocumentNode:
		// For document nodes, delegate to the actual document content
		setYamlValue(root.Content[0], path, value)
	case yaml.MappingNode:
		// For mapping nodes, search for the key in key-value pairs
		// Content is stored as [key1, value1, key2, value2, ...]
		for i := 0; i < len(root.Content); i += 2 {
			if root.Content[i].Value == key {
				// Found the key, recurse into its value
				setYamlValue(root.Content[i+1], rest, value)
				return
			}
		}

		// Key not found, create a new key-value pair
		root.Content = append(root.Content,
			&yaml.Node{
				Kind:  yaml.ScalarNode,
				Value: key,
			},
			&yaml.Node{
				Kind: yaml.MappingNode,
			},
		)

		// Recurse into the newly created value node
		lastIndex := len(root.Content) - 1
		setYamlValue(root.Content[lastIndex], rest, value)
	case yaml.SequenceNode:
		// For sequence nodes, treat the key as an array index
		index, err := strconv.Atoi(key)
		if err != nil {
			e2e.Failf("failed to convert sequence index %q to integer: %v", key, err)
		}

		// Only proceed if the index is within bounds
		if index >= 0 && index < len(root.Content) {
			setYamlValue(root.Content[index], rest, value)
		} else {
			e2e.Failf("sequence index %d is out of bounds (length: %d)", index, len(root.Content))
		}
	}
}
