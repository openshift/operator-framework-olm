// Package db provides database-related utilities for testing OLM operators.
// It includes functions for managing Kubernetes pods and executing commands
// within containers for database operations.
package db

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kcoreclient "k8s.io/client-go/kubernetes/typed/core/v1"
)

// PodConfig holds configuration for a pod.
// It contains the container name and environment variables
// extracted from the pod specification.
type PodConfig struct {
	// Container is the name of the first container in the pod
	Container string
	// Env contains all environment variables from all containers in the pod
	Env map[string]string
}

// getPodConfig retrieves the configuration of a pod including container name and environment variables.
// It fetches the pod by name and extracts environment variables from all containers.
// Returns a PodConfig struct containing the first container's name and all environment variables.
//
//nolint:unused // utility function kept for future use
func getPodConfig(c kcoreclient.PodInterface, podName string) (conf *PodConfig, err error) {
	pod, err := c.Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	// Collect environment variables from all containers
	env := make(map[string]string)
	for _, container := range pod.Spec.Containers {
		for _, e := range container.Env {
			env[e.Name] = e.Value
		}
	}
	return &PodConfig{pod.Spec.Containers[0].Name, env}, nil
}

// firstContainerName retrieves the name of the first container in a pod.
// This is useful when you need to identify the primary container for executing commands.
//
//nolint:unused // utility function kept for future use
func firstContainerName(c kcoreclient.PodInterface, podName string) (string, error) {
	pod, err := c.Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return pod.Spec.Containers[0].Name, nil
}

// isReady checks if a service or component is ready by executing a ping command
// and verifying that the output contains the expected response.
// This is commonly used to verify database connectivity or service health.
//
//nolint:unused // utility function kept for future use
func isReady(oc *util.CLI, podName string, pingCommand, expectedOutput string) (bool, error) {
	out, err := executeShellCommand(oc, podName, pingCommand)
	ok := strings.Contains(out, expectedOutput)
	if !ok {
		err = fmt.Errorf("expected output: %q but actual: %q", expectedOutput, out)
	}
	return ok, err
}

// executeShellCommand executes a shell command inside a pod using kubectl exec.
// It runs the command with bash and returns the output.
// Exit errors are handled gracefully by returning empty output instead of an error.
//
//nolint:unused // utility function kept for future use
func executeShellCommand(oc *util.CLI, podName string, command string) (string, error) {
	// Execute command in pod using kubectl exec with bash
	out, err := oc.Run("exec").Args(podName, "--", "bash", "-c", command).Output()
	if err != nil {
		// Handle exit errors gracefully - return empty output instead of error
		switch err.(type) {
		case *util.ExitError, *exec.ExitError:
			return "", nil
		default:
			return "", err
		}
	}

	return out, nil
}
