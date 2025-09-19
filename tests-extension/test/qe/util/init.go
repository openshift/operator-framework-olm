package util

import (
	"fmt"
	"os"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// InitClusterEnv initializes the cluster environment for testing by setting up test framework and configuration
// This function will panic if initialization fails
func InitClusterEnv() {
	if err := initTestFramework(false); err != nil {
		panic(fmt.Sprintf("Failed to initialize test framework: %v", err))
	}

	if TestContext == nil {
		panic("TestContext is nil, cannot proceed with initialization")
	}

	e2e.AfterReadingAllFlags(TestContext)
	e2e.TestContext.OutputDir = os.TempDir()

	// Configure logging behavior
	e2e.TestContext.DumpLogsOnFailure = true
	TestContext.DumpLogsOnFailure = true
}

// Configuration constants for test framework
const (
	// DefaultAllowedNotReadyNodes defines the default number of nodes allowed to be not ready
	DefaultAllowedNotReadyNodes = 100
	// DefaultMaxNodesToGather defines the default maximum number of nodes to gather information from
	DefaultMaxNodesToGather = 0
)

// initTestFramework initializes the test framework with configuration settings and Ginkgo integration
// Parameters:
//   - dryRun: if true, runs in dry-run mode without executing actual tests
//
// Returns:
//   - error: error if framework initialization fails, nil on success
func initTestFramework(dryRun bool) error {
	if TestContext == nil {
		return fmt.Errorf("TestContext is not initialized")
	}

	// Configure test framework with sensible defaults
	TestContext.AllowedNotReadyNodes = DefaultAllowedNotReadyNodes
	TestContext.MaxNodesToGather = DefaultMaxNodesToGather

	// Register Gomega failure handler
	gomega.RegisterFailHandler(ginkgo.Fail)

	// Initialize the test framework
	if err := InitTest(dryRun); err != nil {
		return fmt.Errorf("failed to initialize test: %w", err)
	}

	return nil
}
