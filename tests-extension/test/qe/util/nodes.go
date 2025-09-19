package util

import (
	"context"
	"fmt"
	"strings"
	"time"

	o "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// GetFirstLinuxWorkerNode returns the first Linux worker node in the cluster (CoreOS or RHEL)
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - string: name of the first Linux worker node found
//   - error: error if no Linux worker node is found, nil on success
func GetFirstLinuxWorkerNode(oc *CLI) (string, error) {
	var (
		workerNode string
		err        error
	)
	workerNode, err = getFirstWorkerNodeByOsID(oc, "rhcos")
	if len(workerNode) == 0 {
		workerNode, err = getFirstWorkerNodeByOsID(oc, "rhel")
	}
	return workerNode, err
}

// GetAllNodesbyOSType returns a list of node names filtered by operating system type
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - ostype: operating system type to filter by (e.g., "linux", "windows")
//
// Returns:
//   - []string: slice of node names matching the specified OS type
//   - error: error if node retrieval fails, nil on success
func GetAllNodesbyOSType(oc *CLI, ostype string) ([]string, error) {
	if oc == nil {
		return nil, fmt.Errorf("CLI client cannot be nil")
	}
	if ostype == "" {
		return nil, fmt.Errorf("OS type cannot be empty")
	}

	var nodesArray []string
	nodes, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("node", "-l", "kubernetes.io/os="+ostype, "-o", "jsonpath='{.items[*].metadata.name}'").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes by OS type %s: %w", ostype, err)
	}

	nodesStr := strings.Trim(nodes, "'")
	// If split an empty string to string array, the default length string array is 1
	// So need to check if string is empty.
	if len(nodesStr) == 0 {
		return nodesArray, nil
	}
	nodesArray = strings.Split(nodesStr, " ")
	return nodesArray, nil
}

// GetAllNodes returns a list of all node names in the cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - []string: slice of all node names in the cluster
//   - error: error if node retrieval fails, nil on success
func GetAllNodes(oc *CLI) ([]string, error) {
	if oc == nil {
		return nil, fmt.Errorf("CLI client cannot be nil")
	}

	nodes, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("node", "-o", "jsonpath='{.items[*].metadata.name}'").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get all nodes: %w", err)
	}

	trimmedNodes := strings.Trim(nodes, "'")
	if len(trimmedNodes) == 0 {
		return []string{}, nil
	}
	return strings.Split(trimmedNodes, " "), nil
}

// GetFirstWorkerNode returns the first worker node in the cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - string: name of the first worker node
//   - error: error if no worker node is found, nil on success
func GetFirstWorkerNode(oc *CLI) (string, error) {
	if oc == nil {
		return "", fmt.Errorf("CLI client cannot be nil")
	}

	workerNodes, err := GetClusterNodesBy(oc, "worker")
	if err != nil {
		return "", fmt.Errorf("failed to get worker nodes: %w", err)
	}
	if len(workerNodes) == 0 {
		return "", fmt.Errorf("no worker nodes found")
	}
	return workerNodes[0], nil
}

// GetFirstMasterNode returns the first master node in the cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - string: name of the first master node
//   - error: error if no master node is found, nil on success
func GetFirstMasterNode(oc *CLI) (string, error) {
	if oc == nil {
		return "", fmt.Errorf("CLI client cannot be nil")
	}

	masterNodes, err := GetClusterNodesBy(oc, "master")
	if err != nil {
		return "", fmt.Errorf("failed to get master nodes: %w", err)
	}
	if len(masterNodes) == 0 {
		return "", fmt.Errorf("no master nodes found")
	}
	return masterNodes[0], nil
}

// GetClusterNodesBy returns cluster nodes filtered by their role
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - role: node role to filter by (e.g., "master", "worker")
//
// Returns:
//   - []string: slice of node names with the specified role
//   - error: error if node retrieval fails, nil on success
func GetClusterNodesBy(oc *CLI, role string) ([]string, error) {
	if oc == nil {
		return nil, fmt.Errorf("CLI client cannot be nil")
	}
	if role == "" {
		return nil, fmt.Errorf("node role cannot be empty")
	}

	nodes, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("node", "-l", "node-role.kubernetes.io/"+role, "-o", "jsonpath='{.items[*].metadata.name}'").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes with role %s: %w", role, err)
	}

	trimmedNodes := strings.Trim(nodes, "'")
	if len(trimmedNodes) == 0 {
		return []string{}, nil
	}
	return strings.Split(trimmedNodes, " "), nil
}

// DebugNodeWithChroot creates a debugging session on a node using chroot to access the host filesystem
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - nodeName: name of the node to debug
//   - cmd: variable arguments representing the command to execute in the debug session
//
// Returns:
//   - string: combined stdout and stderr output from the debug session
//   - error: error if debug session fails, nil on success
func DebugNodeWithChroot(oc *CLI, nodeName string, cmd ...string) (string, error) {
	stdOut, stdErr, err := debugNode(oc, nodeName, []string{}, true, true, cmd...)
	return strings.Join([]string{stdOut, stdErr}, "\n"), err
}

// DebugNodeWithOptions launches a debug container with custom options (e.g., --image)
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - nodeName: name of the node to debug
//   - options: slice of additional options for the debug command
//   - cmd: variable arguments representing the command to execute
//
// Returns:
//   - string: combined stdout and stderr output from the debug session
//   - error: error if debug session fails, nil on success
func DebugNodeWithOptions(oc *CLI, nodeName string, options []string, cmd ...string) (string, error) {
	stdOut, stdErr, err := debugNode(oc, nodeName, options, false, true, cmd...)
	return strings.Join([]string{stdOut, stdErr}, "\n"), err
}

// DebugNodeWithOptionsAndChroot launches a debug container with both custom options and chroot access
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - nodeName: name of the node to debug
//   - options: slice of additional options for the debug command
//   - cmd: variable arguments representing the command to execute
//
// Returns:
//   - string: combined stdout and stderr output from the debug session
//   - error: error if debug session fails, nil on success
func DebugNodeWithOptionsAndChroot(oc *CLI, nodeName string, options []string, cmd ...string) (string, error) {
	stdOut, stdErr, err := debugNode(oc, nodeName, options, true, true, cmd...)
	return strings.Join([]string{stdOut, stdErr}, "\n"), err
}

// DebugNodeRetryWithOptionsAndChroot launches a debug container with retry logic for pod creation failures
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - nodeName: name of the node to debug
//   - options: slice of additional options for the debug command
//   - cmd: variable arguments representing the command to execute
//
// Returns:
//   - string: combined stdout and stderr output from the debug session
//   - error: error if debug session fails after retries, nil on success
func DebugNodeRetryWithOptionsAndChroot(oc *CLI, nodeName string, options []string, cmd ...string) (string, error) {
	var stdErr string
	var stdOut string
	var err error
	errWait := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 30*time.Second, false, func(ctx context.Context) (bool, error) {
		stdOut, stdErr, err = debugNode(oc, nodeName, options, true, true, cmd...)
		if err != nil {
			return false, nil
		}
		return true, nil
	})
	AssertWaitPollNoErr(errWait, fmt.Sprintf("Failed to debug node : %v", errWait))
	return strings.Join([]string{stdOut, stdErr}, "\n"), err
}

// DebugNodeWithOptionsAndChrootWithoutRecoverNsLabel launches debug container without recovering namespace labels
// This function does not restore pod security labels, useful for 4.12+ clusters with changed pod security
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - nodeName: name of the node to debug
//   - options: slice of additional options for the debug command
//   - cmd: variable arguments representing the command to execute
//
// Returns:
//   - stdOut: stdout output from the debug session
//   - stdErr: stderr output from the debug session
//   - err: error if debug session fails, nil on success
func DebugNodeWithOptionsAndChrootWithoutRecoverNsLabel(oc *CLI, nodeName string, options []string, cmd ...string) (string, string, error) {
	return debugNode(oc, nodeName, options, true, false, cmd...)
}

// DebugNode creates a basic debugging session on a node without chroot
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - nodeName: name of the node to debug
//   - cmd: variable arguments representing the command to execute
//
// Returns:
//   - string: combined stdout and stderr output from the debug session
//   - error: error if debug session fails, nil on success
func DebugNode(oc *CLI, nodeName string, cmd ...string) (string, error) {
	stdOut, stdErr, err := debugNode(oc, nodeName, []string{}, false, true, cmd...)
	return strings.Join([]string{stdOut, stdErr}, "\n"), err
}

// debugNode is the internal implementation for node debugging with comprehensive configuration options
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - nodeName: name of the node to debug
//   - cmdOptions: slice of command-line options for the debug command
//   - needChroot: whether to use chroot to access the host filesystem
//   - recoverNsLabels: whether to recover namespace security labels after debugging
//   - cmd: variable arguments representing the command to execute
//
// Returns:
//   - stdOut: stdout output from the debug session
//   - stdErr: stderr output from the debug session
//   - err: error if debug session fails, nil on success
func debugNode(oc *CLI, nodeName string, cmdOptions []string, needChroot bool, recoverNsLabels bool, cmd ...string) (string, string, error) {
	if oc == nil {
		return "", "", fmt.Errorf("CLI client cannot be nil")
	}
	if nodeName == "" {
		return "", "", fmt.Errorf("node name cannot be empty")
	}
	if len(cmd) == 0 {
		return "", "", fmt.Errorf("command cannot be empty")
	}

	var (
		debugNodeNamespace string
		isNsPrivileged     bool
		cargs              []string
		outputError        error
	)
	cargs = []string{"node/" + nodeName}
	// Enhance for debug node namespace used logic
	// if "--to-namespace=" option is used, then uses the input options' namespace, otherwise use oc.Namespace()
	// if oc.Namespace() is empty, uses "default" namespace instead
	hasToNamespaceInCmdOptions, index := StringsSliceElementsHasPrefix(cmdOptions, "--to-namespace=", false)
	if hasToNamespaceInCmdOptions {
		debugNodeNamespace = strings.TrimPrefix(cmdOptions[index], "--to-namespace=")
	} else {
		debugNodeNamespace = oc.Namespace()
		if debugNodeNamespace == "" {
			debugNodeNamespace = "default"
		}
	}
	// Running oc debug node command in normal projects
	// (normal projects mean projects that are not clusters default projects like: "openshift-xxx" et al)
	// need extra configuration on 4.12+ ocp test clusters
	// https://github.com/openshift/oc/blob/master/pkg/helpers/cmd/errors.go#L24-L29
	if !strings.HasPrefix(debugNodeNamespace, "openshift-") { // nolint:nestif
		isNsPrivileged, outputError = IsNamespacePrivileged(oc, debugNodeNamespace)
		if outputError != nil {
			return "", "", outputError
		}
		if !isNsPrivileged {
			if recoverNsLabels {
				defer func() {
					if recoErr := RecoverNamespaceRestricted(oc, debugNodeNamespace); recoErr != nil {
						e2e.Logf("Error recovery NamespaceRestricted: %v", recoErr)
					}
				}()
			}
			outputError = SetNamespacePrivileged(oc, debugNodeNamespace)
			if outputError != nil {
				return "", "", outputError
			}
		}
	}

	// For default nodeSelector enabled test clusters we need to add the extra annotation to avoid the debug pod's
	// nodeSelector overwritten by the scheduler
	if IsDefaultNodeSelectorEnabled(oc) && !IsWorkerNode(oc, nodeName) && !IsSpecifiedAnnotationKeyExist(oc, "ns/"+debugNodeNamespace, "", `openshift.io/node-selector`) {
		if _, addErr := AddAnnotationsToSpecificResource(oc, "ns/"+debugNodeNamespace, "", `openshift.io/node-selector=`); addErr != nil {
			e2e.Logf("Error adding annotation: %v", addErr)
		}
		defer func() {
			if _, removeErr := RemoveAnnotationFromSpecificResource(oc, "ns/"+debugNodeNamespace, "", `openshift.io/node-selector`); removeErr != nil {
				e2e.Logf("Error removing annotation: %v", removeErr)
			}
		}()
	}

	if len(cmdOptions) > 0 {
		cargs = append(cargs, cmdOptions...)
	}
	if !hasToNamespaceInCmdOptions {
		cargs = append(cargs, "--to-namespace="+debugNodeNamespace)
	}
	if needChroot {
		cargs = append(cargs, "--", "chroot", "/host")
	} else {
		cargs = append(cargs, "--")
	}
	cargs = append(cargs, cmd...)
	return oc.AsAdmin().WithoutNamespace().Run("debug").Args(cargs...).Outputs()
}

// DeleteLabelFromNode removes a custom label from the specified node
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - node: name of the node to remove label from
//   - label: label key to remove
//
// Returns:
//   - string: output from the label deletion command
//   - error: error if label deletion fails, nil on success
func DeleteLabelFromNode(oc *CLI, node string, label string) (string, error) {
	return oc.AsAdmin().WithoutNamespace().Run("label").Args("node", node, label+"-").Output()
}

// AddLabelToNode adds a custom label with value to the specified node
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - node: name of the node to add label to
//   - label: label key to add
//   - value: label value to set
//
// Returns:
//   - string: output from the label addition command
//   - error: error if label addition fails, nil on success
func AddLabelToNode(oc *CLI, node string, label string, value string) (string, error) {
	return oc.AsAdmin().WithoutNamespace().Run("label").Args("node", node, label+"="+value).Output()
}

// GetFirstCoreOsWorkerNode returns the first CoreOS (RHCOS) worker node in the cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - string: name of the first CoreOS worker node
//   - error: error if no CoreOS worker node is found, nil on success
func GetFirstCoreOsWorkerNode(oc *CLI) (string, error) {
	return getFirstWorkerNodeByOsID(oc, "rhcos")
}

// GetFirstRhelWorkerNode returns the first Red Hat Enterprise Linux worker node in the cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - string: name of the first RHEL worker node
//   - error: error if no RHEL worker node is found, nil on success
func GetFirstRhelWorkerNode(oc *CLI) (string, error) {
	return getFirstWorkerNodeByOsID(oc, "rhel")
}

// getFirstWorkerNodeByOsID returns the first cluster node matching OS ID
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - osID: operating system ID to match (e.g., "rhcos", "rhel")
//
// Returns:
//   - string: name of the first matching node
//   - error: error if no matching node is found, nil on success
func getFirstWorkerNodeByOsID(oc *CLI, osID string) (string, error) {
	nodes, err := GetClusterNodesBy(oc, "worker")
	for _, node := range nodes {
		stdout, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("node/"+node, "-o", "jsonpath=\"{.metadata.labels.node\\.openshift\\.io/os_id}\"").Output()
		if strings.Trim(stdout, "\"") == osID {
			return node, err
		}
	}
	return "", err
}

// GetNodeHostname returns the hostname of the specified cluster node
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - node: name of the node to get hostname for
//
// Returns:
//   - string: hostname of the node
//   - error: error if hostname retrieval fails, nil on success
func GetNodeHostname(oc *CLI, node string) (string, error) {
	hostname, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("node", node, "-o", "jsonpath='{..kubernetes\\.io/hostname}'").Output()
	return strings.Trim(hostname, "'"), err
}

// GetClusterNodesByRoleInHostedCluster returns cluster nodes by role in a hosted cluster environment
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - role: node role to filter by (e.g., "master", "worker")
//
// Returns:
//   - []string: slice of node names with the specified role in hosted cluster
//   - error: error if node retrieval fails, nil on success
func GetClusterNodesByRoleInHostedCluster(oc *CLI, role string) ([]string, error) {
	nodes, err := oc.AsAdmin().AsGuestKubeconf().Run("get").Args("node", "-l", "node-role.kubernetes.io/"+role, "-o", "jsonpath='{.items[*].metadata.name}'").Output()
	return strings.Split(strings.Trim(nodes, "'"), " "), err
}

// getFirstNodeByOsIDInHostedCluster returns the first cluster node by role and OS ID in hosted cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - role: node role to filter by (e.g., "worker", "master")
//   - osID: operating system ID to match (e.g., "rhcos", "rhel")
//
// Returns:
//   - string: name of the first matching node in hosted cluster
//   - error: error if no matching node is found, nil on success
func getFirstNodeByOsIDInHostedCluster(oc *CLI, role string, osID string) (string, error) {
	nodes, err := GetClusterNodesByRoleInHostedCluster(oc, role)
	for _, node := range nodes {
		stdout, err := oc.AsAdmin().AsGuestKubeconf().Run("get").Args("node/"+node, "-o", "jsonpath=\"{.metadata.labels.node\\.openshift\\.io/os_id}\"").Output()
		if strings.Trim(stdout, "\"") == osID {
			return node, err
		}
	}
	return "", err
}

// GetFirstLinuxWorkerNodeInHostedCluster returns the first Linux worker node in a hosted cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - string: name of the first Linux worker node in hosted cluster
//   - error: error if no Linux worker node is found, nil on success
func GetFirstLinuxWorkerNodeInHostedCluster(oc *CLI) (string, error) {
	var (
		workerNode string
		err        error
	)
	workerNode, err = getFirstNodeByOsIDInHostedCluster(oc, "worker", "rhcos")
	if len(workerNode) == 0 {
		workerNode, err = getFirstNodeByOsIDInHostedCluster(oc, "worker", "rhel")
	}
	return workerNode, err
}

// GetAllNodesByNodePoolNameInHostedCluster returns all node names belonging to a specific node pool in hosted cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - nodePoolName: name of the node pool to filter by
//
// Returns:
//   - []string: slice of node names in the specified node pool
//   - error: error if node retrieval fails, nil on success
func GetAllNodesByNodePoolNameInHostedCluster(oc *CLI, nodePoolName string) ([]string, error) {
	nodes, err := oc.AsAdmin().AsGuestKubeconf().Run("get").Args("node", "-l", "hypershift.openshift.io/nodePool="+nodePoolName, "-ojsonpath='{.items[*].metadata.name}'").Output()
	return strings.Split(strings.Trim(nodes, "'"), " "), err
}

// GetFirstWorkerNodeByNodePoolNameInHostedCluster returns the first worker node from a specific node pool in hosted cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - nodePoolName: name of the node pool to get worker node from
//
// Returns:
//   - string: name of the first worker node in the specified node pool
//   - error: error if no worker node is found in the node pool, nil on success
func GetFirstWorkerNodeByNodePoolNameInHostedCluster(oc *CLI, nodePoolName string) (string, error) {
	if oc == nil {
		return "", fmt.Errorf("CLI client cannot be nil")
	}
	if nodePoolName == "" {
		return "", fmt.Errorf("node pool name cannot be empty")
	}

	workerNodes, err := GetAllNodesByNodePoolNameInHostedCluster(oc, nodePoolName)
	if err != nil {
		return "", fmt.Errorf("failed to get nodes from pool %s: %w", nodePoolName, err)
	}
	if len(workerNodes) == 0 {
		return "", fmt.Errorf("no worker nodes found in node pool %s", nodePoolName)
	}
	return workerNodes[0], nil
}

// GetSchedulableLinuxWorkerNodes returns Linux worker nodes that are ready and schedulable
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - []v1.Node: slice of Node objects that are Linux, worker role, ready status, and schedulable
//   - error: error if node retrieval fails, nil on success
func GetSchedulableLinuxWorkerNodes(oc *CLI) ([]corev1.Node, error) {
	if oc == nil {
		return nil, fmt.Errorf("CLI client cannot be nil")
	}

	var nodes, workers []corev1.Node
	linuxNodes, err := oc.AdminKubeClient().CoreV1().Nodes().List(context.Background(), metav1.ListOptions{LabelSelector: "kubernetes.io/os=linux"})
	// get schedulable linux worker nodes
	for _, node := range linuxNodes.Items {
		if _, ok := node.Labels["node-role.kubernetes.io/worker"]; ok && !node.Spec.Unschedulable {
			workers = append(workers, node)
		}
	}
	// get ready nodes
	for _, worker := range workers {
		for _, con := range worker.Status.Conditions {
			if con.Type == "Ready" && con.Status == "True" {
				nodes = append(nodes, worker)
				break
			}
		}
	}
	return nodes, err
}

// GetPodsNodesMap returns a mapping of node names to their running pods
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - nodes: slice of Node objects to map pods for
//
// Returns:
//   - map[string][]v1.Pod: map where keys are node names and values are slices of pods running on each node
func GetPodsNodesMap(oc *CLI, nodes []corev1.Node) map[string][]corev1.Pod {
	if oc == nil {
		e2e.Logf("CLI client is nil, returning empty map")
		return make(map[string][]corev1.Pod)
	}

	podsMap := make(map[string][]corev1.Pod)
	projects, err := oc.AdminKubeClient().CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	o.Expect(err).NotTo(o.HaveOccurred())

	// get pod list in each node
	for _, project := range projects.Items {
		pods, err := oc.AdminKubeClient().CoreV1().Pods(project.Name).List(context.Background(), metav1.ListOptions{})
		o.Expect(err).NotTo(o.HaveOccurred())
		for _, pod := range pods.Items {
			if pod.Status.Phase != "Failed" && pod.Status.Phase != "Succeeded" {
				podsMap[pod.Spec.NodeName] = append(podsMap[pod.Spec.NodeName], pod)
			}
		}
	}

	nodeNames := make([]string, 0, len(nodes))
	for _, node := range nodes {
		nodeNames = append(nodeNames, node.Name)
	}
	// Helper function to check if a string exists in a slice
	contains := func(slice []string, item string) bool {
		for _, element := range slice {
			if element == item {
				return true
			}
		}
		return false
	}
	// if the key is not in nodes list, remove the element from the map
	for podmap := range podsMap {
		if !contains(nodeNames, podmap) {
			delete(podsMap, podmap)
		}
	}
	return podsMap
}

// NodeResources contains the resources of CPU and Memory in a node
type NodeResources struct {
	CPU    int64
	Memory int64
}

// GetRequestedResourcesNodesMap calculates the total requested CPU and memory resources for each node
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - nodes: slice of Node objects to calculate requested resources for
//
// Returns:
//   - map[string]NodeResources: map where keys are node names and values contain total requested CPU and memory
func GetRequestedResourcesNodesMap(oc *CLI, nodes []corev1.Node) map[string]NodeResources {
	rmap := make(map[string]NodeResources)
	podsMap := GetPodsNodesMap(oc, nodes)
	for nodeName := range podsMap {
		var totalRequestedCPU, totalRequestedMemory int64
		for _, pod := range podsMap[nodeName] {
			for _, container := range pod.Spec.Containers {
				totalRequestedCPU += container.Resources.Requests.Cpu().MilliValue()
				totalRequestedMemory += container.Resources.Requests.Memory().MilliValue()
			}
		}
		rmap[nodeName] = NodeResources{totalRequestedCPU, totalRequestedMemory}
	}
	return rmap
}

// GetAllocatableResourcesNodesMap returns the total allocatable CPU and memory resources for each node
// Parameters:
//   - nodes: slice of Node objects to get allocatable resources for
//
// Returns:
//   - map[string]NodeResources: map where keys are node names and values contain allocatable CPU and memory
func GetAllocatableResourcesNodesMap(nodes []corev1.Node) map[string]NodeResources {
	rmap := make(map[string]NodeResources)
	for _, node := range nodes {
		rmap[node.Name] = NodeResources{node.Status.Allocatable.Cpu().MilliValue(), node.Status.Allocatable.Memory().MilliValue()}
	}
	return rmap
}

// GetRemainingResourcesNodesMap calculates the remaining available CPU and memory resources for each node
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - nodes: slice of Node objects to calculate remaining resources for
//
// Returns:
//   - map[string]NodeResources: map where keys are node names and values contain remaining CPU and memory (allocatable - requested)
func GetRemainingResourcesNodesMap(oc *CLI, nodes []corev1.Node) map[string]NodeResources {
	rmap := make(map[string]NodeResources)
	requested := GetRequestedResourcesNodesMap(oc, nodes)
	allocatable := GetAllocatableResourcesNodesMap(nodes)

	for _, node := range nodes {
		rmap[node.Name] = NodeResources{allocatable[node.Name].CPU - requested[node.Name].CPU, allocatable[node.Name].Memory - requested[node.Name].Memory}
	}
	return rmap
}

// getNodesByRoleAndOsID returns a list of node names filtered by both role and OS ID
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - role: node role to filter by (e.g., "worker", "master")
//   - osID: operating system ID to filter by (e.g., "rhcos", "rhel")
//
// Returns:
//   - []string: slice of node names matching both role and OS ID criteria
//   - error: error if node retrieval fails, nil on success
func getNodesByRoleAndOsID(oc *CLI, role string, osID string) ([]string, error) {
	var nodesList []string
	nodes, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("node", "-l", "node-role.kubernetes.io/"+role+",node.openshift.io/os_id="+osID, "-o", "jsonpath='{.items[*].metadata.name}'").Output()
	nodes = strings.Trim(nodes, "'")
	if len(nodes) != 0 {
		nodesList = strings.Split(nodes, " ")
	}
	return nodesList, err
}

// GetAllWorkerNodesByOSID returns a list of all worker nodes filtered by OS ID
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - osID: operating system ID to filter by (e.g., "rhcos", "rhel")
//
// Returns:
//   - []string: slice of worker node names matching the OS ID
//   - error: error if node retrieval fails, nil on success
func GetAllWorkerNodesByOSID(oc *CLI, osID string) ([]string, error) {
	return getNodesByRoleAndOsID(oc, "worker", osID)
}

// GetNodeArchByName retrieves the architecture of a node by its name
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - nodeName: name of the node to get architecture for
//
// Returns:
//   - string: node architecture (e.g., "amd64", "arm64", "ppc64le", "s390x")
func GetNodeArchByName(oc *CLI, nodeName string) string {
	nodeArch, err := GetResourceSpecificLabelValue(oc, "node/"+nodeName, "", "kubernetes\\.io/arch")
	o.Expect(err).NotTo(o.HaveOccurred(), "Fail to get node/%s arch: %v\n", nodeName, err)
	e2e.Logf(`The node/%s arch is "%s"`, nodeName, nodeArch)
	return nodeArch
}

// GetNodeListByLabel retrieves a list of node names that have the specified label
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - labelKey: label key to filter nodes by
//
// Returns:
//   - []string: slice of node names that have the specified label
func GetNodeListByLabel(oc *CLI, labelKey string) []string {
	output, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("node", "-l", labelKey, "-o=jsonpath={.items[*].metadata.name}").Output()
	o.Expect(err).NotTo(o.HaveOccurred(), "Fail to get node with label %v, got error: %v\n", labelKey, err)
	nodeNameList := strings.Fields(output)
	return nodeNameList
}

// IsDefaultNodeSelectorEnabled determines if the cluster has default node selector enabled
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - bool: true if default node selector is configured, false otherwise
func IsDefaultNodeSelectorEnabled(oc *CLI) bool {
	defaultNodeSelector, getNodeSelectorErr := oc.AsAdmin().WithoutNamespace().Run("get").Args("scheduler", "cluster", "-o=jsonpath={.spec.defaultNodeSelector}").Output()
	if getNodeSelectorErr != nil && strings.Contains(defaultNodeSelector, `the server doesn't have a resource type`) {
		e2e.Logf("WARNING: The scheduler API is not supported on the test cluster")
		return false
	}
	o.Expect(getNodeSelectorErr).NotTo(o.HaveOccurred(), "Fail to get cluster scheduler defaultNodeSelector got error: %v\n", getNodeSelectorErr)
	return !strings.EqualFold(defaultNodeSelector, "")
}

// IsWorkerNode determines if the specified node has the worker role
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - nodeName: name of the node to check
//
// Returns:
//   - bool: true if the node has worker role, false otherwise
func IsWorkerNode(oc *CLI, nodeName string) bool {
	isWorker, _ := StringsSliceContains(GetNodeListByLabel(oc, `node-role.kubernetes.io/worker`), nodeName)
	return isWorker
}

// WaitForNodeToDisappear waits for a node to be removed from the cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - nodeName: name of the node to wait for disappearance
//   - timeout: maximum time to wait for node disappearance
//   - interval: polling interval for checking node status
func WaitForNodeToDisappear(oc *CLI, nodeName string, timeout, interval time.Duration) {
	o.Eventually(func() bool {
		_, err := oc.AdminKubeClient().CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
		if errors.IsNotFound(err) {
			return true
		}
		o.Expect(err).ShouldNot(o.HaveOccurred(), fmt.Sprintf("Unexpected error: %s", errors.ReasonForError(err)))
		e2e.Logf("Still waiting for node %s to disappear", nodeName)
		return false
	}).WithTimeout(timeout).WithPolling(interval).Should(o.BeTrue())
}

// DebugNodeRetryWithOptionsAndChrootWithStdErr launches debug container with retry and separate stdout/stderr handling
// This function handles warning messages and separates stdout from stderr for better error diagnosis
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - nodeName: name of the node to debug
//   - options: slice of additional options for the debug command
//   - cmd: variable arguments representing the command to execute
//
// Returns:
//   - string: stdout output from the debug session
//   - string: stderr output from the debug session
//   - error: error if debug session fails after retries, nil on success
func DebugNodeRetryWithOptionsAndChrootWithStdErr(oc *CLI, nodeName string, options []string, cmd ...string) (string, string, error) {
	var stdErr string
	var stdOut string
	var err error
	errWait := wait.PollUntilContextTimeout(context.TODO(), 3*time.Second, 30*time.Second, false, func(ctx context.Context) (bool, error) {
		stdOut, stdErr, err = debugNode(oc, nodeName, options, true, true, cmd...)
		if err != nil {
			return false, nil
		}
		return true, nil
	})
	AssertWaitPollNoErr(errWait, fmt.Sprintf("Failed to debug node : %v", errWait))
	return stdOut, stdErr, err
}
