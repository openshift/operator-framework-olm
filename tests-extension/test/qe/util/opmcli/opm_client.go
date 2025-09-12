// Package opmcli provides utilities for interacting with the OPM (Operator Package Manager) CLI tool.
// It includes functionality to execute OPM commands, manage authentication, and handle command output
// for testing OLM (Operator Lifecycle Manager) operators and catalogs.
package opmcli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"

	g "github.com/onsi/ginkgo/v2"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// CLI provides functionality to call the OPM CLI tool and manage command execution.
// It encapsulates command configuration, authentication, and output handling for OPM operations.
type CLI struct {
	// execPath is the path to the executable (e.g., "opm", "initializer")
	execPath string
	// ExecCommandPath is the working directory for command execution
	ExecCommandPath string
	// verb is the main command verb (e.g., "index", "render")
	verb string
	// username for authentication (defaults to "admin")
	username string
	// globalArgs contains the main command and global flags
	globalArgs []string
	// commandArgs contains command-specific arguments
	commandArgs []string
	// finalArgs is the combination of globalArgs and commandArgs
	finalArgs []string
	// stdin buffer for command input
	stdin *bytes.Buffer
	// stdout writer for command output
	stdout io.Writer
	// stderr writer for command errors
	stderr io.Writer
	// verbose enables debug output
	verbose bool
	// showInfo enables command execution logging
	showInfo bool
	// skipTLS enables TLS verification skipping
	skipTLS bool
	// podmanAuthfile path to authentication file for registry access
	podmanAuthfile string
}

// NewOpmCLI initializes a new OPM CLI client with default configuration.
// Sets up the client to use the "opm" executable with admin privileges and info logging enabled.
func NewOpmCLI() *CLI {
	client := &CLI{}
	client.username = "admin"
	client.execPath = "opm"
	client.showInfo = true
	return client
}

// NewInitializer creates a new CLI client configured for the "initializer" tool.
// The initializer is used for setting up and configuring operator catalogs.
func NewInitializer() *CLI {
	client := &CLI{}
	client.username = "admin"
	client.execPath = "initializer"
	client.showInfo = true
	return client
}

// Run prepares and configures a CLI command for execution with the given command verb and arguments.
// Creates a new CLI instance with the specified commands and inherits configuration from the parent.
// Automatically adds TLS skip flag if configured.
func (c *CLI) Run(commands ...string) *CLI {
	// Initialize buffers for command I/O
	in, out, errout := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	// Create new CLI instance inheriting parent configuration
	client := &CLI{
		execPath:        c.execPath,
		verb:            commands[0],
		username:        c.username,
		ExecCommandPath: c.ExecCommandPath,
		podmanAuthfile:  c.podmanAuthfile,
	}
	// Add TLS skip flag if configured
	if c.skipTLS {
		client.globalArgs = append([]string{"--skip-tls=true"}, commands...)
	} else {
		client.globalArgs = commands
	}
	// Set up I/O buffers
	client.stdin, client.stdout, client.stderr = in, out, errout
	return client.setOutput(c.stdout)
}

// setOutput allows overriding the default command output writer.
// This is useful for redirecting command output to custom writers or files.
func (c *CLI) setOutput(out io.Writer) *CLI {
	c.stdout = out
	return c
}

// Args sets additional arguments for the OPM CLI command.
// Combines global arguments with command-specific arguments to form the final argument list.
func (c *CLI) Args(args ...string) *CLI {
	c.commandArgs = args
	c.finalArgs = append(c.globalArgs, c.commandArgs...)
	return c
}

// printCmd returns the complete command string for logging and debugging purposes.
// Joins all final arguments into a single space-separated string.
func (c *CLI) printCmd() string {
	return strings.Join(c.finalArgs, " ")
}

// ExitError represents an error that occurred during command execution.
// It extends exec.ExitError with additional context including the command and stderr output.
type ExitError struct {
	// Cmd is the complete command that was executed
	Cmd string
	// StdErr contains the standard error output from the command
	StdErr string
	// Embedded ExitError provides the underlying execution error details
	*exec.ExitError
}

// FatalErr exits the test in case a fatal error has occurred.
// Prints the stack trace to help with debugging and fails the test using the e2e framework.
func FatalErr(msg interface{}) {
	// Print stack trace for debugging - the path that leads to this being called isn't always clear
	if _, err := fmt.Fprintln(g.GinkgoWriter, string(debug.Stack())); err != nil {
		e2e.Logf("Failed to write debug stack: %v", err)
	}
	e2e.Failf("%v", msg)
}

// SetAuthFile configures the authentication file for registry access.
// The auth file is used by podman/docker for authenticating with container registries.
func (c *CLI) SetAuthFile(authfile string) *CLI {
	c.podmanAuthfile = authfile
	return c
}

// Output executes the configured command and returns the combined stdout/stderr output.
// Sets up the execution environment including authentication and working directory,
// then executes the command and handles different types of errors appropriately.
func (c *CLI) Output() (string, error) {
	// Log debug information if verbose mode is enabled
	if c.verbose {
		e2e.Logf("DEBUG: %s %s\n", c.execPath, c.printCmd())
	}
	// Create the command with the executable and arguments
	cmd := exec.Command(c.execPath, c.finalArgs...)
	// Set registry authentication file if configured
	if c.podmanAuthfile != "" {
		cmd.Env = append(os.Environ(), "REGISTRY_AUTH_FILE="+c.podmanAuthfile)
	}
	// Set working directory if specified
	if c.ExecCommandPath != "" {
		e2e.Logf("set exec command path is %s\n", c.ExecCommandPath)
		cmd.Dir = c.ExecCommandPath
	}
	// Set stdin buffer
	cmd.Stdin = c.stdin
	// Log command execution if info logging is enabled
	if c.showInfo {
		e2e.Logf("Running '%s %s'", c.execPath, strings.Join(c.finalArgs, " "))
	}
	// Execute command and capture combined output
	out, err := cmd.CombinedOutput()
	trimmed := strings.TrimSpace(string(out))
	// Handle different types of execution results
	switch exitErr := err.(type) {
	case nil:
		// Successful execution
		c.stdout = bytes.NewBuffer(out)
		return trimmed, nil
	case *exec.ExitError:
		// Command executed but returned non-zero exit code
		e2e.Logf("Error running %v:\n%s", cmd, trimmed)
		return trimmed, &ExitError{ExitError: exitErr, Cmd: c.execPath + " " + strings.Join(c.finalArgs, " "), StdErr: trimmed}
	default:
		// Fatal error preventing command execution
		FatalErr(fmt.Errorf("unable to execute %q: %v", c.execPath, err))
		// unreachable code
		return "", nil
	}
}

// GetDirPath recursively searches for a directory containing a file/directory with the specified prefix.
// Traverses up the directory tree from the given path until it finds a match or reaches the root.
// Returns the full path to the directory containing the matching file/directory, or empty string if not found.
func GetDirPath(filePathStr string, filePre string) string {
	// Check if we've reached an invalid path or root directory
	if !strings.Contains(filePathStr, "/") || filePathStr == "/" {
		return ""
	}
	// Split the path into directory and file components
	dir, file := filepath.Split(filePathStr)
	// Check if the current file/directory matches the prefix
	if strings.HasPrefix(file, filePre) {
		return filePathStr
	} else {
		// Recursively search in the parent directory
		return GetDirPath(filepath.Dir(dir), filePre)
	}
}

// DeleteDir finds and deletes a directory that contains a file/directory with the specified prefix.
// Uses GetDirPath to locate the target directory, then removes it and all its contents.
// Returns true if the directory was successfully deleted, false otherwise.
func DeleteDir(filePathStr string, filePre string) bool {
	// Find the directory path containing the specified prefix
	filePathToDelete := GetDirPath(filePathStr, filePre)
	if filePathToDelete == "" || !strings.Contains(filePathToDelete, filePre) {
		e2e.Logf("there is no such dir %s", filePre)
		return false
	} else {
		e2e.Logf("remove dir %s", filePathToDelete)
		// Remove the directory and all its contents
		if err := os.RemoveAll(filePathToDelete); err != nil {
			e2e.Logf("Failed to remove directory %s: %v", filePathToDelete, err)
		}
		// Verify the directory was actually deleted
		if _, err := os.Stat(filePathToDelete); err == nil {
			e2e.Logf("delele dir %s failed", filePathToDelete)
			return false
		}
		return true
	}
}
