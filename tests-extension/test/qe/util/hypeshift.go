package util

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/blang/semver/v4"
	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/wait"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

type HostedClusterPlatformType = string

const (
	// AWSPlatform represents Amazon Web Services infrastructure.
	AWSPlatform HostedClusterPlatformType = "AWS"

	// NonePlatform represents user supplied (e.g. bare metal) infrastructure.
	NonePlatform HostedClusterPlatformType = "None"

	// IBMCloudPlatform represents IBM Cloud infrastructure.
	IBMCloudPlatform HostedClusterPlatformType = "IBMCloud"

	// AgentPlatform represents user supplied insfrastructure booted with agents.
	AgentPlatform HostedClusterPlatformType = "Agent"

	// KubevirtPlatform represents Kubevirt infrastructure.
	KubevirtPlatform HostedClusterPlatformType = "KubeVirt"

	// AzurePlatform represents Azure infrastructure.
	AzurePlatform HostedClusterPlatformType = "Azure"

	// PowerVSPlatform represents PowerVS infrastructure.
	PowerVSPlatform HostedClusterPlatformType = "PowerVS"
)

var (
	hypershiftBinaryOnce  sync.Once
	hypershiftBinarySetup error
)

// EnsureHypershiftBinary ensures hypershift binary is available with cross-process synchronization
func EnsureHypershiftBinary(oc *CLI) error {
	hypershiftBinaryOnce.Do(func() {
		hypershiftBinarySetup = ensureHypershiftBinaryWithLock(oc)
	})
	return hypershiftBinarySetup
}

// isArchitectureSupported checks if the current architecture supports hypershift binary
func isArchitectureSupported() error {
	arch := runtime.GOARCH
	os := runtime.GOOS

	e2e.Logf("Current runtime: OS=%s, ARCH=%s", os, arch)

	// Hypershift binary is x86-64 Linux ELF executable
	if os != "linux" {
		return fmt.Errorf("hypershift binary only supports Linux, current OS: %s", os)
	}

	if arch != "amd64" {
		return fmt.Errorf("hypershift binary only supports x86-64 architecture, current architecture: %s", arch)
	}

	return nil
}

func ensureHypershiftBinaryWithLock(oc *CLI) error {
	e2e.Logf("Setting up hypershift binary...")

	_, err := exec.LookPath("hypershift")
	if err == nil {
		e2e.Logf("hypershift command is found in PATH")
		return nil
	}

	// Check architecture compatibility first
	if err := isArchitectureSupported(); err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	var hypershiftPath string
	var hypershiftDir string
	var lockPath string

	if cwd == "/tmp" {
		hypershiftPath = filepath.Join(cwd, "hypershift")
		hypershiftDir = cwd
		lockPath = filepath.Join(cwd, "hypershift.lock")
	} else {
		hypershiftPath = "/tmp/hypershift"
		hypershiftDir = "/tmp"
		lockPath = "/tmp/hypershift.lock"
	}

	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := lockFile.Close(); closeErr != nil {
			e2e.Logf("Failed to close lock file: %v", closeErr)
		}
	}()

	e2e.Logf("Acquiring file lock for hypershift binary installation...")
	maxRetries := 90
	for i := 0; i < maxRetries; i++ {
		err = syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if err != syscall.EWOULDBLOCK {
			return err
		}
		e2e.Logf("Lock is held by another process, retrying in 1 second... (%d/%d)", i+1, maxRetries)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		return err
	}
	defer func() {
		if unlockErr := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_UN); unlockErr != nil {
			e2e.Logf("Failed to unlock file: %v", unlockErr)
		}
	}()

	e2e.Logf("File lock acquired, checking if hypershift binary exists...")

	if _, err := os.Stat(hypershiftPath); err == nil {
		e2e.Logf("Hypershift binary already exists at %s", hypershiftPath)
		return setupHypershiftEnv(hypershiftDir)
	}

	e2e.Logf("Extracting hypershift binary from container image...")
	err = oc.WithoutNamespace().Run("image").Args("extract", "quay.io/hypershift/hypershift-operator:latest", "--file=/usr/bin/hypershift").Execute()
	if err != nil {
		return err
	}

	if hypershiftDir != cwd {
		// Move hypershift binary to target directory
		err = exec.Command("mv", "hypershift", hypershiftPath).Run()
		if err != nil {
			return fmt.Errorf("failed to move hypershift binary: %v", err)
		}
		// Set executable permissions
		err = exec.Command("chmod", "755", hypershiftPath).Run()
		if err != nil {
			return fmt.Errorf("failed to set permissions on hypershift binary: %v", err)
		}
	} else {
		// Set executable permissions in current directory
		err = exec.Command("chmod", "755", "hypershift").Run()
		if err != nil {
			return fmt.Errorf("failed to set permissions on hypershift binary: %v", err)
		}
	}

	e2e.Logf("Hypershift binary installed at %s", hypershiftPath)
	return setupHypershiftEnv(hypershiftDir)
}

func setupHypershiftEnv(hypershiftDir string) error {
	currentPath := os.Getenv("PATH")
	if !strings.Contains(currentPath, hypershiftDir) {
		newPath := hypershiftDir + ":" + currentPath
		err := os.Setenv("PATH", newPath)
		if err != nil {
			return err
		}
		e2e.Logf("Added %s to PATH: %s", hypershiftDir, newPath)
	}
	return nil
}

// IsHypershiftMgmtCluster checks if the current cluster is a hypershift management cluster
// Returns true if both hypershift operator and hosted cluster namespace exist
func IsHypershiftMgmtCluster(oc *CLI) bool {
	operatorNS := GetHyperShiftOperatorNameSpace(oc)
	hostedclusterNS := GetHyperShiftHostedClusterNameSpace(oc)
	return len(operatorNS) > 0 && len(hostedclusterNS) > 0
}

// ValidHypershiftAndGetGuestKubeConf check if it is hypershift env and get kubeconf of the hosted cluster
// the first return is hosted cluster name
// the second return is the file of kubeconfig of the hosted cluster
// the third return is the hostedcluster namespace in mgmt cluster which contains the generated resources
// if it is not hypershift env, it will skip test.
func ValidHypershiftAndGetGuestKubeConf(oc *CLI) (string, string, string) {
	if !IsHypershiftMgmtCluster(oc) {
		g.Skip("this is not a hypershift management cluster, skip test run")
	}

	operatorNS := GetHyperShiftOperatorNameSpace(oc)
	hostedclusterNS := GetHyperShiftHostedClusterNameSpace(oc)

	clusterNames, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(
		"-n", hostedclusterNS, "hostedclusters", "-o=jsonpath={.items[*].metadata.name}").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	if len(clusterNames) <= 0 {
		g.Skip("there is no hosted cluster, skip test run")
	}

	// Verify HyperShift operator is running
	hypershiftPodStatus, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(
		"-n", operatorNS, "pod", "-l", "hypershift.openshift.io/operator-component=operator", "-l", "app=operator", "-o=jsonpath={.items[*].status.phase}").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	o.Expect(hypershiftPodStatus).To(o.ContainSubstring("Running"))

	//get first hosted cluster to run test
	e2e.Logf("the hosted cluster names: %s, and will select the first", clusterNames)
	clusterName := strings.Split(clusterNames, " ")[0]

	var hostedClusterKubeconfigFile string
	if os.Getenv("GUEST_KUBECONFIG") != "" {
		e2e.Logf("the kubeconfig you set GUEST_KUBECONFIG must be that of the hosted cluster %s in namespace %s", clusterName, hostedclusterNS)
		hostedClusterKubeconfigFile = os.Getenv("GUEST_KUBECONFIG")
		e2e.Logf("use a known hosted cluster kubeconfig: %v", hostedClusterKubeconfigFile)
	} else {
		// Check if hypershift command is available
		_, err := exec.LookPath("hypershift")
		if err != nil {
			g.Skip("hypershift command not found in PATH, cannot create kubeconfig for hosted cluster")
		}

		hostedClusterKubeconfigFile = "/tmp/guestcluster-kubeconfig-" + clusterName + "-" + GetRandomString()
		output, err := exec.Command("bash", "-c", fmt.Sprintf("hypershift create kubeconfig --name %s --namespace %s > %s",
			clusterName, hostedclusterNS, hostedClusterKubeconfigFile)).Output()
		e2e.Logf("the cmd output: %s", string(output))
		o.Expect(err).NotTo(o.HaveOccurred())
		e2e.Logf("create a new hosted cluster kubeconfig: %v", hostedClusterKubeconfigFile)
	}
	e2e.Logf("if you want hostedcluster controlplane namespace, you could get it by combining %s and %s with -", hostedclusterNS, clusterName)
	return clusterName, hostedClusterKubeconfigFile, hostedclusterNS
}

// ValidHypershiftAndGetGuestKubeConfWithNoSkip check if it is hypershift env and get kubeconf of the hosted cluster
// the first return is hosted cluster name
// the second return is the file of kubeconfig of the hosted cluster
// the third return is the hostedcluster namespace in mgmt cluster which contains the generated resources
// if it is not hypershift env, it will not skip the testcase and return null string.
func ValidHypershiftAndGetGuestKubeConfWithNoSkip(oc *CLI) (string, string, string) {
	if !IsHypershiftMgmtCluster(oc) {
		return "", "", ""
	}

	operatorNS := GetHyperShiftOperatorNameSpace(oc)
	hostedclusterNS := GetHyperShiftHostedClusterNameSpace(oc)

	clusterNames, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(
		"-n", hostedclusterNS, "hostedclusters", "-o=jsonpath={.items[*].metadata.name}").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	if len(clusterNames) <= 0 {
		return "", "", ""
	}

	// Verify HyperShift operator is running
	hypershiftPodStatus, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(
		"-n", operatorNS, "pod", "-l", "hypershift.openshift.io/operator-component=operator", "-l", "app=operator", "-o=jsonpath={.items[*].status.phase}").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	o.Expect(hypershiftPodStatus).To(o.ContainSubstring("Running"))

	//get first hosted cluster to run test
	e2e.Logf("the hosted cluster names: %s, and will select the first", clusterNames)
	clusterName := strings.Split(clusterNames, " ")[0]

	var hostedClusterKubeconfigFile string
	if os.Getenv("GUEST_KUBECONFIG") != "" {
		e2e.Logf("the kubeconfig you set GUEST_KUBECONFIG must be that of the guestcluster %s in namespace %s", clusterName, hostedclusterNS)
		hostedClusterKubeconfigFile = os.Getenv("GUEST_KUBECONFIG")
		e2e.Logf("use a known hosted cluster kubeconfig: %v", hostedClusterKubeconfigFile)
	} else {
		// Check if hypershift command is available
		_, err := exec.LookPath("hypershift")
		if err != nil {
			return "", "", ""
		}

		hostedClusterKubeconfigFile = "/tmp/guestcluster-kubeconfig-" + clusterName + "-" + GetRandomString()
		output, err := exec.Command("bash", "-c", fmt.Sprintf("hypershift create kubeconfig --name %s --namespace %s > %s",
			clusterName, hostedclusterNS, hostedClusterKubeconfigFile)).Output()
		e2e.Logf("the cmd output: %s", string(output))
		o.Expect(err).NotTo(o.HaveOccurred())
		e2e.Logf("create a new hosted cluster kubeconfig: %v", hostedClusterKubeconfigFile)
	}
	e2e.Logf("if you want hostedcluster controlplane namespace, you could get it by combining %s and %s with -", hostedclusterNS, clusterName)
	return clusterName, hostedClusterKubeconfigFile, hostedclusterNS
}

// GetHyperShiftOperatorNameSpace get hypershift operator namespace
// if not exist, it will return empty string.
func GetHyperShiftOperatorNameSpace(oc *CLI) string {
	args := []string{
		"pods", "-A",
		"-l", "hypershift.openshift.io/operator-component=operator",
		"-l", "app=operator",
		"--ignore-not-found",
		"-ojsonpath={.items[0].metadata.namespace}",
	}
	namespace, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(args...).Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	return strings.TrimSpace(namespace)
}

// GetHyperShiftHostedClusterNameSpace get hypershift hostedcluster namespace
// if not exist, it will return empty string. If more than one exists, it will return the first one.
func GetHyperShiftHostedClusterNameSpace(oc *CLI) string {
	// First check if HostedCluster CRD exists
	_, crdErr := oc.AsAdmin().WithoutNamespace().Run("get").Args("crd", "hostedclusters.hypershift.openshift.io", "--ignore-not-found").Output()
	if crdErr != nil {
		e2e.Logf("HostedCluster CRD not found, this is not a hypershift management cluster: %v", crdErr)
		return ""
	}

	namespace, err := oc.AsAdmin().WithoutNamespace().Run("get").Args(
		"hostedcluster", "-A", "--ignore-not-found", "-ojsonpath={.items[*].metadata.namespace}").Output()

	if err != nil && !strings.Contains(namespace, "the server doesn't have a resource type") {
		o.Expect(err).NotTo(o.HaveOccurred(), "get hostedcluster fail: %v", err)
	}

	if len(namespace) <= 0 {
		return namespace
	}
	namespaces := strings.Fields(namespace)
	if len(namespaces) == 1 {
		return namespaces[0]
	}
	ns := ""
	for _, ns = range namespaces {
		if ns != "clusters" {
			break
		}
	}
	return ns
}

// GetHostedClusterPlatformType returns a hosted cluster platform type
// oc is the management cluster client to query the hosted cluster platform type based on hostedcluster CR obj
func GetHostedClusterPlatformType(oc *CLI, clusterName, clusterNamespace string) (HostedClusterPlatformType, error) {
	if IsHypershiftHostedCluster(oc) {
		return "", fmt.Errorf("this is a hosted cluster env. You should use oc of the management cluster")
	}
	return oc.AsAdmin().WithoutNamespace().Run("get").Args("hostedcluster", clusterName, "-n", clusterNamespace, `-ojsonpath={.spec.platform.type}`).Output()
}

// GetNodePoolNamesbyHostedClusterName gets the nodepools names of the hosted cluster
func GetNodePoolNamesbyHostedClusterName(oc *CLI, hostedClusterName, hostedClusterNS string) []string {
	var nodePoolName []string
	nodePoolNameList, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("nodepool", "-n", hostedClusterNS, "-ojsonpath={.items[*].metadata.name}").Output()
	o.Expect(err).NotTo(o.HaveOccurred())
	o.Expect(nodePoolNameList).NotTo(o.BeEmpty())

	nodePoolName = strings.Fields(nodePoolNameList)
	e2e.Logf("\n\nGot nodepool(s) for the hosted cluster %s: %v\n", hostedClusterName, nodePoolName)
	return nodePoolName
}

// GetHostedClusterVersion gets a HostedCluster's version from the management cluster.
func GetHostedClusterVersion(mgmtOc *CLI, hostedClusterName, hostedClusterNs string) semver.Version {
	hcVersionStr, _, err := mgmtOc.
		AsAdmin().
		WithoutNamespace().
		Run("get").
		Args("hostedcluster", hostedClusterName, "-n", hostedClusterNs, `-o=jsonpath={.status.version.history[?(@.state!="")].version}`).
		Outputs()
	o.Expect(err).NotTo(o.HaveOccurred())

	hcVersion := semver.MustParse(hcVersionStr)
	e2e.Logf("Found hosted cluster %s version = %q", hostedClusterName, hcVersion)
	return hcVersion
}

func CheckHypershiftOperatorExistence(mgmtOC *CLI) (bool, error) {
	stdout, _, err := mgmtOC.AsAdmin().WithoutNamespace().Run("get").
		Args("pods", "-n", "hypershift", "-o=jsonpath={.items[*].metadata.name}").Outputs()
	if err != nil {
		return false, fmt.Errorf("failed to get HO Pods: %v", err)
	}
	return len(stdout) > 0, nil
}

func SkipOnHypershiftOperatorExistence(mgmtOC *CLI, expectHO bool) {
	HOExist, err := CheckHypershiftOperatorExistence(mgmtOC)
	if err != nil {
		e2e.Logf("failed to check Hypershift Operator existence: %v, defaulting to not found", err)
	}

	if HOExist && !expectHO {
		g.Skip("Not expecting Hypershift Operator but it is found, skip the test")
	}
	if !HOExist && expectHO {
		g.Skip("Expecting Hypershift Operator but it is not found, skip the test")
	}
}

// WaitForHypershiftHostedClusterReady waits for the hostedCluster ready
func WaitForHypershiftHostedClusterReady(oc *CLI, hostedClusterName, hostedClusterNS string) {
	pollWaitErr := wait.PollUntilContextTimeout(context.Background(), 20*time.Second, 10*time.Minute, false, func(cxt context.Context) (bool, error) {
		hostedClusterAvailable, getStatusErr := oc.AsAdmin().WithoutNamespace().Run("get").Args("hostedclusters", "-n", hostedClusterNS, "--ignore-not-found", hostedClusterName, `-ojsonpath='{.status.conditions[?(@.type=="Available")].status}'`).Output()
		if getStatusErr != nil {
			e2e.Logf("Failed to get hosted cluster %q status: %v, try next round", hostedClusterName, getStatusErr)
			return false, nil
		}
		if !strings.Contains(hostedClusterAvailable, "True") {
			e2e.Logf("Hosted cluster %q status: Available=%s, try next round", hostedClusterName, hostedClusterAvailable)
			return false, nil
		}

		hostedClusterProgressState, getStateErr := oc.AsAdmin().WithoutNamespace().Run("get").Args("hostedclusters", "-n", hostedClusterNS, "--ignore-not-found", hostedClusterName, `-ojsonpath={.status.version.history[?(@.state!="")].state}`).Output()
		if getStateErr != nil {
			e2e.Logf("Failed to get hosted cluster %q progress state: %v, try next round", hostedClusterName, getStateErr)
			return false, nil
		}
		if !strings.Contains(hostedClusterProgressState, "Completed") {
			e2e.Logf("Hosted cluster %q progress state: %q, try next round", hostedClusterName, hostedClusterProgressState)
			return false, nil
		}
		e2e.Logf("Hosted cluster %q is ready now", hostedClusterName)
		return true, nil
	})
	AssertWaitPollNoErr(pollWaitErr, fmt.Sprintf("Hosted cluster %q still not ready", hostedClusterName))

}
