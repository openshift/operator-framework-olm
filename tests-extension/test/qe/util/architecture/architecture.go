package architecture

import (
	"fmt"
	"strings"

	g "github.com/onsi/ginkgo/v2"
	o "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/sets"
	e2e "k8s.io/kubernetes/test/e2e/framework"

	exutil "github.com/openshift/operator-framework-olm/tests-extension/test/qe/util"
)

type Architecture int

const (
	AMD64 Architecture = iota
	ARM64
	PPC64LE
	S390X
	MULTI
	UNKNOWN
)

const (
	NodeArchitectureLabel = "kubernetes.io/arch"
)

// SkipIfNoNodeWithArchitectures skips the test if the cluster does not have nodes with all of the specified architectures
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - architectures: variable number of Architecture values to check for in the cluster
func SkipIfNoNodeWithArchitectures(oc *exutil.CLI, architectures ...Architecture) {
	if sets.New(
		GetAvailableArchitecturesSet(oc)...).IsSuperset(
		sets.New(architectures...)) {
		return
	}
	g.Skip("Skip for no nodes with requested architectures")
}

// SkipArchitectures skips the test if the cluster architecture matches any of the specified architectures
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//   - architectures: variable number of Architecture values to skip the test for
//
// Returns:
//   - architecture: the detected cluster architecture
func SkipArchitectures(oc *exutil.CLI, architectures ...Architecture) Architecture {
	architecture := ClusterArchitecture(oc)
	for _, arch := range architectures {
		if arch == architecture {
			g.Skip(fmt.Sprintf("Skip for cluster architecture: %s", arch.String()))
		}
	}
	return architecture
}

// SkipNonAmd64SingleArch skips the test if the cluster is not a single-architecture AMD64 cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - Architecture: the detected cluster architecture (only returned if test is not skipped)
func SkipNonAmd64SingleArch(oc *exutil.CLI) Architecture {
	architecture := ClusterArchitecture(oc)
	if architecture != AMD64 {
		g.Skip(fmt.Sprintf("Skip for cluster architecture: %s", architecture.String()))
	}
	return architecture
}

// getNodeArchitectures is a helper function that retrieves all node architectures from the cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - []string: slice of architecture strings from all cluster nodes
func getNodeArchitectures(oc *exutil.CLI) []string {
	output, err := oc.WithoutNamespace().AsAdmin().Run("get").Args("nodes", "-o=jsonpath={.items[*].status.nodeInfo.architecture}").Output()
	if err != nil {
		g.Skip(fmt.Sprintf("unable to get cluster node architectures: %v", err))
	}
	if output == "" {
		g.Skip("no nodes found or architecture information missing")
	}
	return strings.Fields(output) // Use Fields instead of Split to handle multiple spaces
}

// GetAvailableArchitecturesSet returns a list of unique architectures available in the cluster nodes
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - []Architecture: slice of unique architectures found across all cluster nodes
func GetAvailableArchitecturesSet(oc *exutil.CLI) []Architecture {
	architectureStrings := getNodeArchitectures(oc)
	if len(architectureStrings) == 0 {
		g.Skip("no node architectures found")
	}

	// Use map for deduplication with Architecture as key
	archSet := make(map[Architecture]struct{})
	for _, archStr := range architectureStrings {
		if archStr != "" { // Skip empty strings
			arch := FromString(archStr)
			archSet[arch] = struct{}{}
		}
	}

	// Convert map keys to slice
	architectures := make([]Architecture, 0, len(archSet))
	for arch := range archSet {
		architectures = append(architectures, arch)
	}
	return architectures
}

// SkipNonMultiArchCluster skips the test if the cluster is not a multi-architecture cluster
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
func SkipNonMultiArchCluster(oc *exutil.CLI) {
	if !IsMultiArchCluster(oc) {
		g.Skip("This cluster is not multi-arch cluster, skip this case!")
	}
}

// IsMultiArchCluster checks if the cluster is a multi-architecture cluster (has nodes with different architectures)
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - bool: true if cluster has nodes with different architectures, false otherwise
func IsMultiArchCluster(oc *exutil.CLI) bool {
	architectures := GetAvailableArchitecturesSet(oc)
	return len(architectures) > 1
}

// FromString converts a string representation of architecture to the corresponding Architecture enum value
// Parameters:
//   - arch: string representation of the architecture (e.g., "amd64", "arm64", "ppc64le", "s390x", "multi")
//
// Returns:
//   - Architecture: corresponding Architecture enum value, UNKNOWN for unrecognized architectures
func FromString(arch string) Architecture {
	switch strings.ToLower(strings.TrimSpace(arch)) {
	case "amd64", "x86_64":
		return AMD64
	case "arm64", "aarch64":
		return ARM64
	case "ppc64le":
		return PPC64LE
	case "s390x":
		return S390X
	case "multi":
		return MULTI
	case "":
		e2e.Failf("Empty architecture string provided")
		return UNKNOWN
	default:
		e2e.Logf("Unknown architecture '%s', treating as UNKNOWN", arch)
		return UNKNOWN
	}
}

// String returns the string representation of the Architecture enum value
// Parameters:
//   - a: Architecture enum value to convert
//
// Returns:
//   - string: string representation of the architecture
func (a Architecture) String() string {
	switch a {
	case AMD64:
		return "amd64"
	case ARM64:
		return "arm64"
	case PPC64LE:
		return "ppc64le"
	case S390X:
		return "s390x"
	case MULTI:
		return "multi"
	case UNKNOWN:
		return "unknown"
	default:
		e2e.Failf("Unhandled architecture enum value: %d", a)
		return "unknown"
	}
}

// ClusterArchitecture determines and returns the cluster's architecture type, returning MULTI if nodes have different architectures
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - Architecture: Architecture enum representing the cluster's architecture (MULTI if mixed architectures)
func ClusterArchitecture(oc *exutil.CLI) Architecture {
	architectureStrings := getNodeArchitectures(oc)
	if len(architectureStrings) == 0 {
		g.Skip("no node architectures found")
	}

	// Filter out empty strings and convert to Architecture
	var architectures []Architecture
	for _, archStr := range architectureStrings {
		if archStr != "" {
			architectures = append(architectures, FromString(archStr))
		}
	}

	if len(architectures) == 0 {
		g.Skip("no valid node architectures found")
	}

	// Check if all architectures are the same
	firstArch := architectures[0]
	for _, arch := range architectures[1:] {
		if arch != firstArch {
			e2e.Logf("Found multi-architecture cluster with architectures: %v", architectureStrings)
			return MULTI
		}
	}

	return firstArch
}

// GNUString returns the GNU/Linux standard string representation of the Architecture enum value
// Parameters:
//   - a: Architecture enum value to convert
//
// Returns:
//   - string: GNU/Linux standard architecture string (e.g., "x86_64" for AMD64, "aarch64" for ARM64)
func (a Architecture) GNUString() string {
	switch a {
	case AMD64:
		return "x86_64"
	case ARM64:
		return "aarch64"
	case PPC64LE:
		return "ppc64le"
	case S390X:
		return "s390x"
	case MULTI:
		return "multi"
	case UNKNOWN:
		return "unknown"
	default:
		e2e.Failf("Unhandled architecture enum value: %d", a)
		return "unknown"
	}
}

// GetControlPlaneArch retrieves the architecture of the control plane (master) node
// Parameters:
//   - oc: CLI client for interacting with the OpenShift cluster
//
// Returns:
//   - Architecture: Architecture enum value of the first master node
func GetControlPlaneArch(oc *exutil.CLI) Architecture {
	masterNode, err := exutil.GetFirstMasterNode(oc)
	o.Expect(err).NotTo(o.HaveOccurred())

	architectureStr, err := oc.AsAdmin().WithoutNamespace().Run("get").Args("node", masterNode, "-o=jsonpath={.status.nodeInfo.architecture}").Output()
	o.Expect(err).NotTo(o.HaveOccurred())

	architectureStr = strings.TrimSpace(architectureStr)
	if architectureStr == "" {
		g.Skip(fmt.Sprintf("Control plane node %s has no architecture information", masterNode))
	}

	return FromString(architectureStr)
}
