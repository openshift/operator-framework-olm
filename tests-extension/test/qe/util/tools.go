package util

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	o "github.com/onsi/gomega"
	authorizationv1 "k8s.io/api/authorization/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	kclientset "k8s.io/client-go/kubernetes"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// DuplicateFileToPath copies a file from source path to destination path with error handling
// Parameters:
//   - srcPath: path to the source file to copy
//   - destPath: path where the file will be copied (created with mode 0666 if not exists, truncated if exists)
func DuplicateFileToPath(srcPath string, destPath string) {
	// Validate and clean paths to prevent path traversal attacks
	srcPath = filepath.Clean(srcPath)
	destPath = filepath.Clean(destPath)

	// Ensure source file exists and is readable
	srcInfo, err := os.Stat(srcPath)
	o.Expect(err).NotTo(o.HaveOccurred(), "Source file does not exist or is not accessible")
	o.Expect(srcInfo.Mode().IsRegular()).To(o.BeTrue(), "Source path is not a regular file")

	srcFile, err := os.Open(srcPath)
	o.Expect(err).NotTo(o.HaveOccurred())
	defer func() {
		o.Expect(srcFile.Close()).NotTo(o.HaveOccurred())
	}()

	// Create destination file with restrictive permissions (0644)
	destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	o.Expect(err).NotTo(o.HaveOccurred())
	defer func() {
		o.Expect(destFile.Close()).NotTo(o.HaveOccurred())
	}()

	_, err = io.Copy(destFile, srcFile)
	o.Expect(err).NotTo(o.HaveOccurred())
	o.Expect(destFile.Sync()).NotTo(o.HaveOccurred())
}

// DuplicateFileToTemp creates a temporary duplicate of a file with a specified prefix
// Parameters:
//   - srcPath: path to the source file to duplicate
//   - destPrefix: prefix for the temporary file name
//
// Returns:
//   - string: path to the created temporary file
func DuplicateFileToTemp(srcPath string, destPrefix string) string {
	destFile, err := os.CreateTemp(os.TempDir(), destPrefix)
	o.Expect(err).NotTo(o.HaveOccurred(), "Failed to create temporary file")
	o.Expect(destFile.Close()).NotTo(o.HaveOccurred(), "Failed to close temporary file")

	destPath := destFile.Name()
	DuplicateFileToPath(srcPath, destPath)
	return destPath
}

// MoveFileToPath moves a file from source to destination, handling cross-device moves gracefully
// Parameters:
//   - srcPath: path to the source file to move
//   - destPath: destination path for the file
func MoveFileToPath(srcPath string, destPath string) {
	switch err := os.Rename(srcPath, destPath); {
	case err == nil:
		return
	case strings.Contains(err.Error(), "invalid cross-device link"):
		e2e.Logf("Failed to rename file from %s to %s: %v, attempting an alternative", srcPath, destPath, err)
		DuplicateFileToPath(srcPath, destPath)
		o.Expect(os.Remove(srcPath)).NotTo(o.HaveOccurred(), "Failed to remove source file")
	default:
		o.Expect(err).NotTo(o.HaveOccurred(), "Failed to rename source file")
	}
}

// DeleteLabelsFromSpecificResource removes multiple labels from a Kubernetes resource
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - resourceKindAndName: resource type and name (e.g., "pod/my-pod", "deployment/my-app")
//   - resourceNamespace: namespace of the resource (empty string for cluster-scoped resources)
//   - labelNames: variable number of label keys to remove
//
// Returns:
//   - string: output from the label deletion command
//   - error: error if label deletion fails, nil on success
func DeleteLabelsFromSpecificResource(oc *CLI, resourceKindAndName string, resourceNamespace string, labelNames ...string) (string, error) {
	var cargs []string
	if resourceNamespace != "" {
		cargs = append(cargs, "-n", resourceNamespace)
	}
	cargs = append(cargs, resourceKindAndName)
	cargs = append(cargs, StringsSliceElementsAddSuffix(labelNames, "-")...)
	return oc.AsAdmin().WithoutNamespace().Run("label").Args(cargs...).Output()
}

// AddLabelsToSpecificResource adds or updates labels on a Kubernetes resource
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - resourceKindAndName: resource type and name (e.g., "pod/my-pod", "deployment/my-app")
//   - resourceNamespace: namespace of the resource (empty string for cluster-scoped resources)
//   - labels: variable number of label assignments in format "key=value"
//
// Returns:
//   - string: output from the label addition command
//   - error: error if label addition fails, nil on success
func AddLabelsToSpecificResource(oc *CLI, resourceKindAndName string, resourceNamespace string, labels ...string) (string, error) {
	var cargs []string
	if resourceNamespace != "" {
		cargs = append(cargs, "-n", resourceNamespace)
	}
	cargs = append(cargs, resourceKindAndName)
	cargs = append(cargs, labels...)
	cargs = append(cargs, "--overwrite")
	return oc.AsAdmin().WithoutNamespace().Run("label").Args(cargs...).Output()
}

// GetResourceSpecificLabelValue retrieves the value of a specific label from a Kubernetes resource
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - resourceKindAndName: resource type and name (e.g., "pod/my-pod", "deployment/my-app")
//   - resourceNamespace: namespace of the resource (empty string for cluster-scoped resources)
//   - labelName: name of the label to retrieve (use escaped format for special characters)
//
// Returns:
//   - string: value of the specified label
//   - error: error if label retrieval fails, nil on success
func GetResourceSpecificLabelValue(oc *CLI, resourceKindAndName string, resourceNamespace string, labelName string) (string, error) {
	var cargs []string
	if resourceNamespace != "" {
		cargs = append(cargs, "-n", resourceNamespace)
	}
	cargs = append(cargs, resourceKindAndName, "-o=jsonpath={.metadata.labels."+labelName+"}")
	return oc.AsAdmin().WithoutNamespace().Run("get").Args(cargs...).Output()
}

// AddAnnotationsToSpecificResource adds or updates annotations on a Kubernetes resource
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - resourceKindAndName: resource type and name (e.g., "pod/my-pod", "deployment/my-app")
//   - resourceNamespace: namespace of the resource (empty string for cluster-scoped resources)
//   - annotations: variable number of annotation assignments in format "key=value"
//
// Returns:
//   - string: output from the annotation addition command
//   - error: error if annotation addition fails, nil on success
func AddAnnotationsToSpecificResource(oc *CLI, resourceKindAndName, resourceNamespace string, annotations ...string) (string, error) {
	var cargs []string
	if resourceNamespace != "" {
		cargs = append(cargs, "-n", resourceNamespace)
	}
	cargs = append(cargs, resourceKindAndName)
	cargs = append(cargs, annotations...)
	cargs = append(cargs, "--overwrite")
	return oc.AsAdmin().WithoutNamespace().Run("annotate").Args(cargs...).Output()
}

// RemoveAnnotationFromSpecificResource removes a specific annotation from a Kubernetes resource
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - resourceKindAndName: resource type and name (e.g., "pod/my-pod", "deployment/my-app")
//   - resourceNamespace: namespace of the resource (empty string for cluster-scoped resources)
//   - annotationName: name of the annotation key to remove
//
// Returns:
//   - string: output from the annotation removal command
//   - error: error if annotation removal fails, nil on success
func RemoveAnnotationFromSpecificResource(oc *CLI, resourceKindAndName, resourceNamespace string, annotationName string) (string, error) {
	var cargs []string
	if resourceNamespace != "" {
		cargs = append(cargs, "-n", resourceNamespace)
	}
	cargs = append(cargs, resourceKindAndName)
	cargs = append(cargs, annotationName+"-")
	return oc.AsAdmin().WithoutNamespace().Run("annotate").Args(cargs...).Output()
}

// GetAnnotationsFromSpecificResource retrieves all annotations from a Kubernetes resource
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - resourceKindAndName: resource type and name (e.g., "pod/my-pod", "deployment/my-app")
//   - resourceNamespace: namespace of the resource (empty string for cluster-scoped resources)
//
// Returns:
//   - []string: slice of annotation strings in "key=value" format
//   - error: error if annotation retrieval fails, nil on success
func GetAnnotationsFromSpecificResource(oc *CLI, resourceKindAndName, resourceNamespace string) ([]string, error) {
	var cargs []string
	if resourceNamespace != "" {
		cargs = append(cargs, "-n", resourceNamespace)
	}
	cargs = append(cargs, resourceKindAndName, "--list")
	annotationsStr, getAnnotationsErr := oc.AsAdmin().WithoutNamespace().Run("annotate").Args(cargs...).Output()
	if getAnnotationsErr != nil {
		e2e.Logf(`Failed to get annotations from /%s in namespace %s: "%v"`, resourceKindAndName, resourceNamespace, getAnnotationsErr)
	}
	return strings.Fields(annotationsStr), getAnnotationsErr
}

// IsSpecifiedAnnotationKeyExist checks if a specific annotation key exists on a Kubernetes resource
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - resourceKindAndName: resource type and name (e.g., "pod/my-pod", "deployment/my-app")
//   - resourceNamespace: namespace of the resource (empty string for cluster-scoped resources)
//   - annotationKey: annotation key to check for existence
//
// Returns:
//   - bool: true if the annotation key exists, false otherwise
func IsSpecifiedAnnotationKeyExist(oc *CLI, resourceKindAndName, resourceNamespace, annotationKey string) bool {
	resourceAnnotations, getResourceAnnotationsErr := GetAnnotationsFromSpecificResource(oc, resourceKindAndName, resourceNamespace)
	o.Expect(getResourceAnnotationsErr).NotTo(o.HaveOccurred())
	isAnnotationKeyExist, _ := StringsSliceElementsHasPrefix(resourceAnnotations, annotationKey+"=", true)
	return isAnnotationKeyExist
}

// StringsSliceContains checks if a string slice contains a specific element
// Parameters:
//   - stringsSlice: slice of strings to search in
//   - element: exact string to search for
//
// Returns:
//   - bool: true if element is found, false otherwise
//   - int: index of the first matching element (0 if not found)
func StringsSliceContains(stringsSlice []string, element string) (bool, int) {
	for index, strElement := range stringsSlice {
		if strElement == element {
			return true, index
		}
	}
	return false, 0
}

// StringsSliceElementsHasPrefix checks if any element in a string slice has a specific prefix
// Parameters:
//   - stringsSlice: slice of strings to search in
//   - elementPrefix: prefix to search for
//   - sequentialFlag: true for forward search, false for reverse search
//
// Returns:
//   - bool: true if an element with the prefix is found, false otherwise
//   - int: index of the first matching element (0 if not found)
func StringsSliceElementsHasPrefix(stringsSlice []string, elementPrefix string, sequentialFlag bool) (bool, int) {
	if len(stringsSlice) == 0 {
		return false, 0
	}
	if sequentialFlag {
		for index, strElement := range stringsSlice {
			if strings.HasPrefix(strElement, elementPrefix) {
				return true, index
			}
		}
	} else {
		for i := len(stringsSlice) - 1; i >= 0; i-- {
			if strings.HasPrefix(stringsSlice[i], elementPrefix) {
				return true, i
			}
		}
	}
	return false, 0
}

// StringsSliceElementsAddSuffix creates a new string slice with a suffix added to all elements
// Parameters:
//   - stringsSlice: original slice of strings
//   - suffix: suffix to append to each element
//
// Returns:
//   - []string: new slice with suffix added to all elements
func StringsSliceElementsAddSuffix(stringsSlice []string, suffix string) []string {
	if len(stringsSlice) == 0 {
		return []string{}
	}
	// Pre-allocate slice with exact capacity for better memory efficiency
	newStringsSlice := make([]string, 0, len(stringsSlice))
	for _, element := range stringsSlice {
		newStringsSlice = append(newStringsSlice, element+suffix)
	}
	return newStringsSlice
}

const (
	AsAdmin          = true
	AsUser           = false
	WithoutNamespace = true
	WithNamespace    = false
	Immediately      = true
	NotImmediately   = false
	AllowEmpty       = true
	NotAllowEmpty    = false
	Appear           = true
	Disappear        = false
	Compare          = true
	Contain          = false
	RequireNS        = true
	NotRequireNS     = false
	Present          = true
	NotPresent       = false
	Ok               = true
	Nok              = false
	monitorNamespace = "openshift-monitoring"
	prometheusK8s    = "prometheus-k8s"
)

// GetFieldWithJsonpath retrieves a specific field from a Kubernetes resource using JSONPath with polling
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - interval: polling interval between attempts
//   - timeout: maximum time to wait for successful retrieval
//   - immediately: true to wait first interval before starting, false to start immediately
//   - allowEmpty: true to accept empty string results, false to retry on empty results
//   - asAdmin: true to use admin privileges, false for user privileges
//   - withoutNamespace: true to use WithoutNamespace(), false to use current namespace context
//   - parameters: resource specification and JSONPath query (must include "jsonpath" parameter)
//
// Returns:
//   - string: field value extracted using JSONPath
//   - error: error if retrieval fails or times out, nil on success
func GetFieldWithJsonpath(oc *CLI, interval, timeout time.Duration, immediately, allowEmpty, asAdmin, withoutNamespace bool, parameters ...string) (string, error) {
	var result string
	var err error
	usingJsonpath := false
	for _, parameter := range parameters {
		if strings.Contains(parameter, "jsonpath") {
			usingJsonpath = true
		}
	}
	if !usingJsonpath {
		return "", fmt.Errorf("you do not use jsonpath to get field")
	}
	errWait := wait.PollUntilContextTimeout(context.TODO(), interval, timeout, immediately, func(ctx context.Context) (bool, error) {
		result, err = OcAction(oc, "get", asAdmin, withoutNamespace, parameters...)
		if err != nil || (!allowEmpty && strings.TrimSpace(result) == "") {
			e2e.Logf("output is %v, error is %v, and try next", result, err)
			return false, nil
		}
		return true, nil
	})
	e2e.Logf("$oc get %v, the returned resource:%v", parameters, result)
	// replace errWait because it is always timeout if it happned with wait.Poll
	if errWait != nil {
		errWait = fmt.Errorf("can not get resource with %v", parameters)
	}
	return result, errWait
}

// CheckAppearance checks if a Kubernetes resource appears or disappears within a specified timeframe
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - interval: polling interval between checks
//   - timeout: maximum time to wait for the expected state
//   - immediately: true to wait first interval before starting, false to start immediately
//   - asAdmin: true to use admin privileges, false for user privileges
//   - withoutNamespace: true to use WithoutNamespace(), false to use current namespace context
//   - appear: true to wait for resource appearance, false to wait for disappearance
//   - parameters: resource specification (e.g., "pod", "name" or "-n", "namespace", "pod", "name")
//
// Returns:
//   - bool: true if the expected appearance/disappearance occurred within timeout, false otherwise
func CheckAppearance(oc *CLI, interval, timeout time.Duration, immediately, asAdmin, withoutNamespace, appear bool, parameters ...string) bool {
	if !appear {
		parameters = append(parameters, "--ignore-not-found")
	}
	err := wait.PollUntilContextTimeout(context.TODO(), interval, timeout, immediately, func(ctx context.Context) (bool, error) {
		output, err := OcAction(oc, "get", asAdmin, withoutNamespace, parameters...)
		if err != nil {
			e2e.Logf("the get error is %v, and try next", err)
			return false, nil
		}
		e2e.Logf("output: %v", output)
		if !appear && output == "" {
			return true, nil
		}
		if appear && output != "" && !strings.Contains(strings.ToLower(output), "no resources found") {
			return true, nil
		}
		return false, nil
	})
	return err == nil
}

// CleanupResource deletes a Kubernetes resource and waits for it to be completely removed
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - interval: polling interval to check for resource removal
//   - timeout: maximum time to wait for resource deletion
//   - asAdmin: true to use admin privileges, false for user privileges
//   - withoutNamespace: true to use WithoutNamespace(), false to use current namespace context
//   - parameters: resource specification (e.g., "pod", "name" or "-n", "namespace", "pod", "name")
func CleanupResource(oc *CLI, interval, timeout time.Duration, asAdmin, withoutNamespace bool, parameters ...string) {
	output, err := OcAction(oc, "delete", asAdmin, withoutNamespace, parameters...)
	if err != nil && (strings.Contains(output, "NotFound") || strings.Contains(output, "No resources found")) {
		e2e.Logf("the resource is deleted already")
		return
	}
	o.Expect(err).NotTo(o.HaveOccurred())

	err = wait.PollUntilContextTimeout(context.TODO(), interval, timeout, false, func(ctx context.Context) (bool, error) {
		output, err := OcAction(oc, "get", asAdmin, withoutNamespace, parameters...)
		if err != nil && (strings.Contains(output, "NotFound") || strings.Contains(output, "No resources found")) {
			e2e.Logf("the resource is delete successfully")
			return true, nil
		}
		return false, nil
	})
	AssertWaitPollNoErr(err, fmt.Sprintf("can not remove %v", parameters))
}

// OcAction executes OpenShift CLI commands with configurable privilege and namespace context
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - action: oc command action (e.g., "get", "create", "delete", "patch")
//   - asAdmin: true to use admin privileges, false for user privileges
//   - withoutNamespace: true to use WithoutNamespace(), false to use current namespace context
//   - parameters: command arguments and options
//
// Returns:
//   - string: command output
//   - error: error if command execution fails, nil on success
func OcAction(oc *CLI, action string, asAdmin, withoutNamespace bool, parameters ...string) (string, error) {
	if asAdmin && withoutNamespace {
		return oc.AsAdmin().WithoutNamespace().Run(action).Args(parameters...).Output()
	}
	if asAdmin && !withoutNamespace {
		return oc.AsAdmin().Run(action).Args(parameters...).Output()
	}
	if !asAdmin && withoutNamespace {
		return oc.WithoutNamespace().Run(action).Args(parameters...).Output()
	}
	if !asAdmin && !withoutNamespace {
		return oc.Run(action).Args(parameters...).Output()
	}
	return "", nil
}

// WaitForResourceUpdate waits for a Kubernetes resource's resourceVersion to change, indicating an update
// Parameters:
//   - ctx: context for cancellation and timeout control
//   - oc: CLI client for interacting with the OpenShift cluster
//   - interval: polling interval to check for resourceVersion changes
//   - timeout: maximum time to wait for resource update
//   - kindAndName: resource type and name (e.g., "pod/my-pod", "deployment/my-app")
//   - namespace: namespace of the resource (empty string for cluster-scoped resources)
//   - oldResourceVersion: the current resourceVersion to compare against
//
// Returns:
//   - error: error if update is not detected within timeout, nil if resource was updated
func WaitForResourceUpdate(ctx context.Context, oc *CLI, interval, timeout time.Duration, kindAndName, namespace, oldResourceVersion string) error {
	args := []string{kindAndName}
	if len(namespace) > 0 {
		args = append(args, "-n", namespace)
	}
	args = append(args, "-o=jsonpath={.metadata.resourceVersion}")
	return wait.PollUntilContextTimeout(ctx, interval, timeout, true, func(ctx context.Context) (bool, error) {
		resourceVersion, _, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(args...).Outputs()
		if err != nil {
			e2e.Logf("Error getting current resourceVersion: %v", err)
			return false, nil
		}
		if len(resourceVersion) == 0 {
			return false, errors.New("obtained empty resourceVersion")
		}
		if resourceVersion == oldResourceVersion {
			e2e.Logf("resourceVersion unchanged, keep polling")
			return false, nil
		}
		return true, nil
	})
}

// WaitForSelfSAR waits for a SelfSubjectAccessReview to be allowed, indicating permission availability
// Parameters:
//   - interval: polling interval to check permission status
//   - timeout: maximum time to wait for permission to be granted
//   - c: Kubernetes client interface for API calls
//   - selfSAR: SelfSubjectAccessReview specification defining the permission to check
//
// Returns:
//   - error: error if permission is not granted within timeout, nil if permission is allowed
func WaitForSelfSAR(interval, timeout time.Duration, c kclientset.Interface, selfSAR authorizationv1.SelfSubjectAccessReviewSpec) error {
	err := wait.PollUntilContextTimeout(context.TODO(), interval, timeout, true, func(ctx context.Context) (bool, error) {
		res, err := c.AuthorizationV1().SelfSubjectAccessReviews().Create(context.Background(),
			&authorizationv1.SelfSubjectAccessReview{
				Spec: selfSAR,
			}, metav1.CreateOptions{})
		if err != nil {
			return false, err
		}

		if !res.Status.Allowed {
			e2e.Logf("Waiting for SelfSAR (ResourceAttributes: %#v, NonResourceAttributes: %#v) to be allowed, current Status: %#v", selfSAR.ResourceAttributes, selfSAR.NonResourceAttributes, res.Status)
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		return fmt.Errorf("failed to wait for SelfSAR (ResourceAttributes: %#v, NonResourceAttributes: %#v), err: %v", selfSAR.ResourceAttributes, selfSAR.NonResourceAttributes, err)
	}

	return nil
}

// GetSAToken retrieves a service account token for prometheus-k8s from openshift-monitoring namespace
// Handles version compatibility by trying 'oc create token' first, falling back to 'oc sa get-token'
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - string: service account token for prometheus-k8s
//   - error: error if token retrieval fails, nil on success
func GetPrometheusSAToken(oc *CLI) (string, error) {
	return GetSAToken(oc, prometheusK8s, monitorNamespace)
}

func GetSAToken(oc *CLI, sa, ns string) (string, error) {
	e2e.Logf("Getting a token assgined to specific serviceaccount from %s namespace...", ns)

	token, err := oc.AsAdmin().WithoutNamespace().Run("create").Args("token", sa, "-n", ns).Output()
	if err != nil {
		if strings.Contains(token, "unknown command") || strings.Contains(err.Error(), "unknown command") {
			e2e.Logf("oc create token is not supported by current client, use oc sa get-token instead")
			token, err = oc.AsAdmin().WithoutNamespace().Run("sa").Args("get-token", sa, "-n", ns).Output()
			if err != nil {
				return "", fmt.Errorf("failed to get service account token: %w", err)
			}
		} else {
			return "", fmt.Errorf("failed to create token: %w", err)
		}
	}

	return token, err
}

// PatchResource applies a patch to a Kubernetes resource using oc patch command
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - asAdmin: true to use admin privileges, false for user privileges
//   - withoutNamespace: true to use WithoutNamespace(), false to use current namespace context
//   - parameters: patch command arguments including resource specification and patch data
func PatchResource(oc *CLI, asAdmin bool, withoutNamespace bool, parameters ...string) {
	_, err := OcAction(oc, "patch", asAdmin, withoutNamespace, parameters...)
	o.Expect(err).NotTo(o.HaveOccurred())
}

// Is3MasterNoDedicatedWorkerNode checks if the OpenShift cluster has a 3-node configuration
// where all three nodes serve as both master and worker nodes (no dedicated worker nodes)
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - bool: true if cluster has exactly 3 nodes that are both master and worker, false otherwise
func Is3MasterNoDedicatedWorkerNode(oc *CLI) bool {
	masterNodes, err := GetClusterNodesBy(oc, "master")
	o.Expect(err).NotTo(o.HaveOccurred())
	workerNodes, err := GetClusterNodesBy(oc, "worker")
	o.Expect(err).NotTo(o.HaveOccurred())
	if len(masterNodes) != 3 || len(workerNodes) != 3 {
		return false
	}

	matchCount := 0
	for i := 0; i < len(workerNodes); i++ {
		for j := 0; j < len(masterNodes); j++ {
			if workerNodes[i] == masterNodes[j] {
				matchCount++
			}
		}
	}
	return matchCount == 3
}

// IsSNOCluster checks if the OpenShift cluster is a Single Node OpenShift (SNO) deployment
// SNO is a deployment topology where a single node serves as both master and worker
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - bool: true if cluster has exactly 1 node that is both master and worker, false otherwise
func IsSNOCluster(oc *CLI) bool {
	masterNodes, _ := GetClusterNodesBy(oc, "master")
	workerNodes, _ := GetClusterNodesBy(oc, "worker")
	if len(masterNodes) == 1 && len(workerNodes) == 1 && masterNodes[0] == workerNodes[0] {
		return true
	}
	return false
}

func IsPodReady(oc *CLI, ns, label string) bool {
	pods, err := oc.AdminKubeClient().CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{LabelSelector: label})
	if err != nil {
		return true
	}
	if len(pods.Items) == 0 {
		return true
	}
	isReady := true
	for _, pod := range pods.Items {
		if pod.Status.Phase != corev1.PodRunning {
			isReady = false
		}
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodReady && cond.Status != corev1.ConditionTrue {
				isReady = false
			}
		}
	}
	return isReady
}

func MakeArtifactDir(subdir string) string {
	dirPath := os.Getenv("ARTIFACT_DIR")
	if dirPath == "" {
		dirPath = "/tmp"
	}
	e2e.Logf("the log dir path: %s", dirPath)
	logSubDir := dirPath + "/" + subdir
	err := os.MkdirAll(logSubDir, 0755)
	if err != nil {
		e2e.Logf("failed to create %s", logSubDir)
		return ""
	}
	return logSubDir
}
func WriteErrToArtifactDir(oc *CLI, ns, podName, pattern, expattern, caseid string, minutes int) bool {
	logFile, errLog := oc.AsAdmin().WithoutNamespace().Run("logs").Args("-n", ns, podName, "--since", fmt.Sprintf("%dm", minutes)).OutputToFile(podName + ".log")
	if errLog != nil {
		e2e.Logf("can not get log of pod %s in %s", podName, ns)
		return false
	}
	cmd := fmt.Sprintf(`grep -iE '%s' %s | grep -vE '%s' || true`, pattern, logFile, expattern)
	errLogs, errExec := exec.Command("bash", "-c", cmd).Output()
	if errExec != nil {
		e2e.Logf("can not cat error log of pod %s in %s", podName, ns)
		return false
	}

	if len(errLogs) == 0 {
		e2e.Logf("no error log of pod %s in %s", podName, ns)
		return false
	}

	subdir := MakeArtifactDir("podLog")
	if len(subdir) == 0 {
		e2e.Logf("can not make sub dir for log of pod %s in %s", podName, ns)
		return false
	}
	errLogFile := subdir + "/" + caseid + "-" + podName + "-errors.log"
	if writeErr := os.WriteFile(errLogFile, errLogs, 0644); writeErr != nil {
		e2e.Logf("failed to write error logs to %s: %v\n", errLogFile, writeErr)
		return false
	}

	return true
}
