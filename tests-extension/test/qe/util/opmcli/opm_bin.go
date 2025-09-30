package opmcli

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	e2e "k8s.io/kubernetes/test/e2e/framework"
)

var (
	opmBinaryOnce  sync.Once
	opmBinarySetup error
)

// EnsureOPMBinary ensures opm binary is available with cross-process synchronization
func EnsureOPMBinary() error {
	opmBinaryOnce.Do(func() {
		opmBinarySetup = ensureOPMBinaryWithLock()
	})
	return opmBinarySetup
}

func ensureOPMBinaryWithLock() error {
	e2e.Logf("Setting up opm binary...")

	// Check if opm command is already available in PATH
	_, err := exec.LookPath("opm")
	if err == nil {
		e2e.Logf("opm command is found in PATH")
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

	var opmPath string
	var opmDir string
	var lockPath string

	if cwd == "/tmp" {
		opmPath = filepath.Join(cwd, "opm")
		opmDir = cwd
		lockPath = filepath.Join(cwd, "opm.lock")
	} else {
		opmPath = "/tmp/opm"
		opmDir = "/tmp"
		lockPath = "/tmp/opm.lock"
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

	e2e.Logf("Acquiring file lock for opm binary installation...")
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

	e2e.Logf("File lock acquired, checking if opm binary exists...")

	if _, err := os.Stat(opmPath); err == nil {
		e2e.Logf("OPM binary already exists at %s", opmPath)
		return setupOPMEnv(opmDir)
	}

	e2e.Logf("Downloading and installing opm binary...")
	err = downloadAndInstallOPM(opmPath, opmDir, cwd)
	if err != nil {
		return err
	}

	e2e.Logf("OPM binary installed at %s", opmPath)
	return setupOPMEnv(opmDir)
}

// isArchitectureSupported checks if the current architecture supports opm binary
func isArchitectureSupported() error {
	arch := runtime.GOARCH
	osType := runtime.GOOS

	e2e.Logf("Current runtime: OS=%s, ARCH=%s", osType, arch)

	// Check supported architectures
	supportedArchs := map[string]bool{
		"amd64":   true,
		"arm64":   true,
		"ppc64le": true,
		"s390x":   true,
	}

	if !supportedArchs[arch] {
		return fmt.Errorf("opm binary does not support architecture: %s", arch)
	}

	// Check OS support
	if osType == "linux" {
		return nil // All architectures support Linux
	}

	if osType == "darwin" {
		// Only amd64 (x86_64) supports macOS
		if arch == "amd64" {
			return nil
		}
		return fmt.Errorf("macOS only supports amd64 (x86_64) architecture, current: %s", arch)
	}

	return fmt.Errorf("opm binary only supports Linux and macOS, current OS: %s", osType)
}

func downloadAndInstallOPM(opmPath, opmDir, cwd string) error {
	// Get the appropriate URL and filename based on architecture and OS
	tarballURL, tarballFile, binaryName, err := getOPMDownloadInfo()
	if err != nil {
		return err
	}

	// Download the tarball with automatic fallback to "candidate" on 404
	e2e.Logf("Downloading opm tarball from %s", tarballURL)
	err = downloadFileWithFallback(tarballURL, tarballFile)
	if err != nil {
		return fmt.Errorf("failed to download opm tarball: %v", err)
	}
	defer func() {
		if removeErr := os.Remove(tarballFile); removeErr != nil {
			e2e.Logf("Failed to remove tarball file: %v", removeErr)
		}
	}()

	// Extract the tarball
	e2e.Logf("Extracting opm tarball...")
	err = exec.Command("tar", "-xzf", tarballFile).Run()
	if err != nil {
		return fmt.Errorf("failed to extract opm tarball: %v", err)
	}

	// Move and rename the binary if necessary
	if opmDir != cwd {
		// Move binary to target directory and rename to opm
		err = exec.Command("mv", binaryName, opmPath).Run()
		if err != nil {
			return fmt.Errorf("failed to move opm binary: %v", err)
		}
	} else {
		// Only rename if the binary name is different from "opm"
		if binaryName != "opm" {
			err = exec.Command("mv", binaryName, "opm").Run()
			if err != nil {
				return fmt.Errorf("failed to rename opm binary: %v", err)
			}
		}
	}

	// Set executable permissions
	targetPath := opmPath
	if opmDir == cwd {
		targetPath = "opm"
	}
	err = exec.Command("chmod", "755", targetPath).Run()
	if err != nil {
		return fmt.Errorf("failed to set permissions on opm binary: %v", err)
	}

	// Verify installation using absolute path
	verifyPath := opmPath
	if opmDir == cwd {
		verifyPath = filepath.Join(cwd, "opm")
	}
	err = exec.Command(verifyPath, "version").Run()
	if err != nil {
		e2e.Logf("Warning: opm version check failed: %v", err)
	}

	return nil
}

// downloadFileWithFallback downloads a file from the given URL with automatic fallback to "candidate" on 404
// If the URL contains "candidate-X.Y" and returns 404, it will retry with "candidate"
// Example: .../candidate-4.21/opm-linux.tar.gz -> .../candidate/opm-linux.tar.gz
func downloadFileWithFallback(url, filepath string) error {
	err := downloadFile(url, filepath)
	if err == nil {
		return nil
	}

	// Check if the error is a 404 and the URL contains "candidate-"
	if strings.Contains(err.Error(), "404") && strings.Contains(url, "/candidate-") {
		// Find the position of "/candidate-"
		idx := strings.Index(url, "/candidate-")
		if idx != -1 {
			// Find the next "/" after "candidate-X.Y"
			restURL := url[idx+len("/candidate-"):]
			slashIdx := strings.Index(restURL, "/")
			if slashIdx != -1 {
				// Construct fallback URL: everything before "/candidate-" + "/candidate" + everything after version
				fallbackURL := url[:idx] + "/candidate" + restURL[slashIdx:]
				e2e.Logf("Download failed with 404, retrying with fallback URL: %s", fallbackURL)
				return downloadFile(fallbackURL, filepath)
			}
		}
	}

	return err
}

func downloadFile(url, filepath string) error {
	e2e.Logf("Attempting to download from URL: %s", url)
	e2e.Logf("Target file path: %s", filepath)

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", filepath, err)
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			e2e.Logf("Failed to close output file: %v", closeErr)
		}
	}()

	// Create HTTP client with redirect support
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			e2e.Logf("Following redirect to: %s", req.URL.String())
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	// Get the data
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to make HTTP request to %s: %v", url, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			e2e.Logf("Failed to close response body: %v", closeErr)
		}
	}()

	e2e.Logf("HTTP response status: %s (code: %d)", resp.Status, resp.StatusCode)
	e2e.Logf("Final URL after redirects: %s", resp.Request.URL.String())

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s (URL: %s)", resp.Status, url)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write response body to file: %v", err)
	}

	e2e.Logf("Successfully downloaded file to %s", filepath)
	return nil
}

func setupOPMEnv(opmDir string) error {
	currentPath := os.Getenv("PATH")
	if !strings.Contains(currentPath, opmDir) {
		newPath := opmDir + ":" + currentPath
		err := os.Setenv("PATH", newPath)
		if err != nil {
			return err
		}
		e2e.Logf("Added %s to PATH: %s", opmDir, newPath)
	}
	return nil
}

// isRHEL9 checks if the current Linux system is RHEL9
func isRHEL9() bool {
	// Check /etc/os-release for RHEL 9
	content, err := os.ReadFile("/etc/os-release")
	if err != nil {
		e2e.Logf("Cannot read /etc/os-release: %v, defaulting to non-RHEL9", err)
		return false
	}

	osRelease := string(content)
	isRHEL := strings.Contains(osRelease, "ID=\"rhel\"") || strings.Contains(osRelease, "ID=rhel")
	isVersion9 := strings.Contains(osRelease, "VERSION_ID=\"9") || strings.Contains(osRelease, "VERSION_ID=9")

	e2e.Logf("OS release info: RHEL=%v, Version9=%v", isRHEL, isVersion9)
	return isRHEL && isVersion9
}

// getOPMMirrorCandidatePath determines the candidate path for OPM mirror
// It returns "candidate" for the current development version or "candidate-<version>" for released versions
// Priority:
// 1. Environment variable OCP_VERSION (e.g., "4.20" will use "candidate-4.20")
// 2. Auto-detect from cluster using 'oc version' command
// 3. Default to "candidate" if detection fails
//
// Supported version formats:
// - 4.21.0-ec.1 (early candidate)
// - 4.20.0-rc.3 (release candidate)
// - 4.18.26 (GA release)
// - 4.21.0-0.ci-2025-10-13-134809 (CI build)
// - 4.21.0-0.nightly-2025-10-12-115019 (nightly build)
func getOPMMirrorCandidatePath() string {
	// Check environment variable first
	if ocpVersion := os.Getenv("OCP_VERSION"); ocpVersion != "" {
		e2e.Logf("Using OCP_VERSION from environment: %s", ocpVersion)
		return fmt.Sprintf("candidate-%s", ocpVersion)
	}

	// Try to detect cluster version using oc command
	cmd := exec.Command("oc", "version", "-o", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		e2e.Logf("Failed to detect cluster version via 'oc version': %v, defaulting to 'candidate'", err)
		return "candidate"
	}

	// Parse version from output using simple string manipulation
	// The output format is JSON with openshiftVersion field
	// Example: {"openshiftVersion": "4.21.0-0.nightly-2024-01-15-123456"}
	outputStr := string(output)
	versionStart := strings.Index(outputStr, `"openshiftVersion"`)
	if versionStart == -1 {
		e2e.Logf("Could not find openshiftVersion in oc version output, defaulting to 'candidate'")
		return "candidate"
	}

	// Extract version string between quotes after "openshiftVersion":
	versionSubstr := outputStr[versionStart:]
	colonIdx := strings.Index(versionSubstr, ":")
	if colonIdx == -1 {
		e2e.Logf("Invalid openshiftVersion format, defaulting to 'candidate'")
		return "candidate"
	}

	versionValue := versionSubstr[colonIdx+1:]
	quoteStart := strings.Index(versionValue, `"`)
	if quoteStart == -1 {
		e2e.Logf("Invalid openshiftVersion value format, defaulting to 'candidate'")
		return "candidate"
	}

	versionValue = versionValue[quoteStart+1:]
	quoteEnd := strings.Index(versionValue, `"`)
	if quoteEnd == -1 {
		e2e.Logf("Invalid openshiftVersion value format, defaulting to 'candidate'")
		return "candidate"
	}

	fullVersion := versionValue[:quoteEnd]
	e2e.Logf("Detected cluster version: %s", fullVersion)

	// Extract major.minor version from various formats:
	// - "4.21.0-ec.1" -> "4.21"
	// - "4.20.0-rc.3" -> "4.20"
	// - "4.18.26" -> "4.18"
	// - "4.21.0-0.ci-2025-10-13-134809" -> "4.21"
	// - "4.21.0-0.nightly-2025-10-12-115019" -> "4.21"

	// Split by "." and take first two parts
	parts := strings.Split(fullVersion, ".")
	if len(parts) < 2 {
		e2e.Logf("Invalid version format: %s, defaulting to 'candidate'", fullVersion)
		return "candidate"
	}

	majorMinor := fmt.Sprintf("%s.%s", parts[0], parts[1])
	e2e.Logf("Extracted major.minor version: %s", majorMinor)

	// For now, we use "candidate" for the current development version
	// If you want to use "candidate-<version>" for specific versions, you can set OCP_VERSION environment variable
	// Example: export OCP_VERSION=4.20
	return "candidate"
}

// getOPMDownloadInfo returns the appropriate download URL, filename, and binary name based on current OS and architecture
func getOPMDownloadInfo() (tarballURL, tarballFile, binaryName string, err error) {
	arch := runtime.GOARCH
	osType := runtime.GOOS

	e2e.Logf("Determining OPM download info for OS=%s, ARCH=%s", osType, arch)

	// Map Go architecture names to OpenShift architecture names
	archMapping := map[string]string{
		"amd64":   "amd64",
		"arm64":   "arm64",
		"ppc64le": "ppc64le",
		"s390x":   "s390x",
	}

	// Handle special cases for mirror URL architecture naming
	mirrorArch, exists := archMapping[arch]
	if !exists {
		return "", "", "", fmt.Errorf("unsupported architecture: %s", arch)
	}

	switch arch {
	case "amd64":
		// Use x86_64 for the mirror URL
		mirrorArch = "x86_64"
		e2e.Logf("Mapped amd64 to x86_64 for mirror URL")
	case "arm64":
		// Keep arm64 as is (some mirrors might also use aarch64, but arm64 is more common)
		mirrorArch = "arm64"
		e2e.Logf("Using arm64 for mirror URL")
	default:
		e2e.Logf("Using %s for mirror URL", mirrorArch)
	}

	// Construct base URL with /clients path
	// Automatically determine whether to use "candidate" or "candidate-<version>"
	candidatePath := getOPMMirrorCandidatePath()
	baseURL := fmt.Sprintf("https://mirror2.openshift.com/pub/openshift-v4/%s/clients/ocp/%s", mirrorArch, candidatePath)
	e2e.Logf("Constructed base URL: %s", baseURL)

	// Determine filename and binary name based on OS
	switch osType {
	case "linux":
		// Check if this is RHEL9 to determine which variant to use
		if isRHEL9() {
			tarballFile = "opm-linux-rhel9.tar.gz"
			tarballURL = fmt.Sprintf("%s/opm-linux-rhel9.tar.gz", baseURL)
			binaryName = "opm-rhel9"
			e2e.Logf("RHEL9 Linux detected: tarballFile=%s, binaryName=%s", tarballFile, binaryName)
		} else {
			tarballFile = "opm-linux.tar.gz"
			tarballURL = fmt.Sprintf("%s/opm-linux.tar.gz", baseURL)
			binaryName = "opm-rhel8"
			e2e.Logf("Non-RHEL9 Linux detected: tarballFile=%s, binaryName=%s", tarballFile, binaryName)
		}
	case "darwin":
		// macOS only supports amd64 (x86_64), not arm64
		if arch != "amd64" {
			return "", "", "", fmt.Errorf("macOS only supports amd64 (x86_64) architecture, current: %s", arch)
		}
		// macOS uses separate mac tarball
		tarballFile = "opm-mac.tar.gz"
		tarballURL = fmt.Sprintf("%s/opm-mac.tar.gz", baseURL)
		binaryName = "darwin-amd64-opm"
		e2e.Logf("macOS detected: tarballFile=%s, binaryName=%s", tarballFile, binaryName)
	default:
		return "", "", "", fmt.Errorf("unsupported OS: %s", osType)
	}

	e2e.Logf("Final download URL: %s", tarballURL)
	return tarballURL, tarballFile, binaryName, nil
}
