package container

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"runtime/debug"
	"strings"

	g "github.com/onsi/ginkgo/v2"
	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// contains checks if any string in the slice contains the specified substring
// Parameters:
//   - s: slice of strings to search through
//   - e: substring to search for
//
// Returns:
//   - bool: true if any string in the slice contains the substring, false otherwise
func contains(s []string, e string) bool {
	for _, a := range s {
		if strings.Contains(a, e) {
			return true
		}
	}
	return false
}

// ExitError returns the error info
type ExitError struct {
	Cmd    string
	StdErr string
	*exec.ExitError
}

// Error implements the error interface
func (e *ExitError) Error() string {
	return fmt.Sprintf("command '%s' failed: %s", e.Cmd, e.StdErr)
}

// validCommandArg validates command arguments to prevent injection attacks
var validCommandArg = regexp.MustCompile(`^[a-zA-Z0-9._/-]+$`)

// isValidArg checks if a command argument is safe
func isValidArg(arg string) bool {
	if arg == "" {
		return false
	}
	// Allow common podman flags and values
	if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
		return true
	}
	return validCommandArg.MatchString(arg)
}

// FatalErr exits the test in case a fatal error has occurred, printing stack trace and failing the test
// Parameters:
//   - msg: error message or object to display before exiting
func FatalErr(msg any) {
	// the path that leads to this being called isn't always clear...
	if _, err := fmt.Fprintln(g.GinkgoWriter, string(debug.Stack())); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write stack trace: %v\n", err)
	}
	e2e.Failf("%v", msg)
}

// PodmanImage podman image
type PodmanImage struct {
	ID         string            `json:"Id"`
	Size       int64             `json:"Size"`
	Labels     map[string]string `json:"Labels"`
	Names      []string          `json:"Names"`
	Digest     string            `json:"Digest"`
	Digests    []string          `json:"Digests"`
	Dangling   bool              `json:"Dangling"`
	History    []string          `json:"History"`
	Containers int64             `json:"Containers"`
}

// PodmanCLI provides function to run the docker command
type PodmanCLI struct {
	execPath        string
	ExecCommandPath string
	globalArgs      []string
	commandArgs     []string
	finalArgs       []string
	verbose         bool
	stdin           *bytes.Buffer
	stdout          io.Writer
	stderr          io.Writer
	showInfo        bool
	UnsetProxy      bool
	env             []string
}

// NewPodmanCLI initializes and returns a new PodmanCLI instance with default settings
// Returns:
//   - *PodmanCLI: new PodmanCLI instance configured with default values
func NewPodmanCLI() *PodmanCLI {
	return &PodmanCLI{
		execPath:   "podman",
		showInfo:   true,
		UnsetProxy: false,
	}
}

// Run prepares a Podman command for execution with the specified command arguments
// Parameters:
//   - commands: variable number of command arguments to pass to podman
//
// Returns:
//   - *PodmanCLI: new PodmanCLI instance configured for the command execution
func (c *PodmanCLI) Run(commands ...string) *PodmanCLI {
	// Validate commands for security
	for _, cmd := range commands {
		if !isValidArg(cmd) {
			e2e.Logf("Warning: potentially unsafe command argument: %s", cmd)
		}
	}

	in, out, errout := &bytes.Buffer{}, &bytes.Buffer{}, &bytes.Buffer{}
	podman := &PodmanCLI{
		execPath:        c.execPath,
		ExecCommandPath: c.ExecCommandPath,
		UnsetProxy:      c.UnsetProxy,
		showInfo:        c.showInfo,
		env:             c.env,
	}
	podman.globalArgs = commands
	podman.stdin, podman.stdout, podman.stderr = in, out, errout
	return podman.setOutput(c.stdout)
}

// setOutput allows overriding the default command output destination
// Parameters:
//   - out: io.Writer to redirect command output to
//
// Returns:
//   - *PodmanCLI: the PodmanCLI instance with updated output destination
func (c *PodmanCLI) setOutput(out io.Writer) *PodmanCLI {
	if out != nil {
		c.stdout = out
	}
	return c
}

// Args sets additional arguments for the podman CLI command and finalizes the argument list
// Parameters:
//   - args: variable number of additional arguments to append to the command
//
// Returns:
//   - *PodmanCLI: the PodmanCLI instance with updated arguments
func (c *PodmanCLI) Args(args ...string) *PodmanCLI {
	// Validate arguments for security
	for _, arg := range args {
		if !isValidArg(arg) {
			e2e.Logf("Warning: potentially unsafe argument: %s", arg)
		}
	}

	c.commandArgs = args
	c.finalArgs = append(c.globalArgs, c.commandArgs...)
	return c
}

// printCmd returns the complete command string that will be executed
// Returns:
//   - string: space-separated string of all command arguments
func (c *PodmanCLI) printCmd() string {
	return strings.Join(c.finalArgs, " ")
}

// Output executes the prepared podman command and returns combined stdout/stderr output
// Returns:
//   - string: trimmed combined output from stdout and stderr
//   - error: execution error if command fails, nil on success
func (c *PodmanCLI) Output() (string, error) {
	if c.verbose {
		e2e.Logf("DEBUG: podman %s\n", c.printCmd())
	}
	cmd := exec.Command(c.execPath, c.finalArgs...) // nolint:gosec // G204: execPath is hardcoded to "podman" and finalArgs come from test code
	cmd.Env = os.Environ()
	if c.UnsetProxy {
		var envCmd []string
		proxyVars := []string{"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY", "ALL_PROXY", "FTP_PROXY", "SOCKS_PROXY"}
		for _, envVar := range cmd.Env {
			upperEnv := strings.ToUpper(envVar)
			isProxy := false
			for _, proxy := range proxyVars {
				if strings.HasPrefix(upperEnv, proxy+"=") {
					isProxy = true
					break
				}
			}
			if !isProxy {
				envCmd = append(envCmd, envVar)
			}
		}
		cmd.Env = envCmd
	}
	if c.env != nil {
		cmd.Env = append(cmd.Env, c.env...)
	}
	if c.ExecCommandPath != "" {
		e2e.Logf("set exec command path is %s\n", c.ExecCommandPath)
		cmd.Dir = c.ExecCommandPath
	}
	cmd.Stdin = c.stdin
	if c.showInfo {
		e2e.Logf("Running '%s %s'", c.execPath, strings.Join(c.finalArgs, " "))
	}
	out, err := cmd.CombinedOutput()
	trimmed := strings.TrimSpace(string(out))
	if err == nil {
		c.stdout = bytes.NewBuffer(out)
		return trimmed, nil
	}

	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		e2e.Logf("Error running %v:\n%s", cmd, trimmed)
		return trimmed, &ExitError{ExitError: exitErr, Cmd: c.execPath + " " + strings.Join(c.finalArgs, " "), StdErr: trimmed}
	}

	FatalErr(fmt.Errorf("unable to execute %q: %v", c.execPath, err))
	return "", nil
}

// GetImageList retrieves a list of image names from the podman registry
// Returns:
//   - []string: slice of comma-separated image names
//   - error: error if unable to retrieve images, nil on success
func (c *PodmanCLI) GetImageList() ([]string, error) {
	images, err := c.GetImages()
	if err != nil {
		return nil, err
	}

	imageList := make([]string, 0, len(images))
	for _, image := range images {
		e2e.Logf("ID %s, name: %s", image.ID, strings.Join(image.Names, ","))
		imageList = append(imageList, strings.Join(image.Names, ","))
	}
	return imageList, nil
}

// GetImages retrieves detailed information about all podman images in JSON format
// Returns:
//   - []PodmanImage: slice of PodmanImage structs containing image details
//   - error: error if unable to retrieve or parse images, nil on success
func (c *PodmanCLI) GetImages() ([]PodmanImage, error) {
	output, err := c.Run("images").Args("--format", "json").Output()
	if err != nil {
		e2e.Logf("Failed to run 'podman images --format json'")
		return nil, err
	}

	images, err := c.GetImagesByJSON(output)
	if err != nil {
		return nil, err
	}
	return images, nil
}

// GetImagesByJSON parses a JSON string containing image information into PodmanImage structs
// Parameters:
//   - jsonStr: JSON string containing image data from podman images command
//
// Returns:
//   - []PodmanImage: slice of parsed PodmanImage structs
//   - error: error if JSON parsing fails, nil on success
func (c *PodmanCLI) GetImagesByJSON(jsonStr string) ([]PodmanImage, error) {
	if jsonStr == "" {
		return nil, fmt.Errorf("empty JSON string provided")
	}

	var images []PodmanImage
	if err := json.Unmarshal([]byte(jsonStr), &images); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON data: %w", err)
	}

	return images, nil
}

// CheckImageExist checks if a specific image exists in the podman registry
// Parameters:
//   - imageIndex: image name or identifier to search for
//
// Returns:
//   - bool: true if image exists, false otherwise
//   - error: error if unable to retrieve image list, nil on success
func (c *PodmanCLI) CheckImageExist(imageIndex string) (bool, error) {
	if imageIndex == "" {
		return false, fmt.Errorf("image name cannot be empty")
	}

	e2e.Logf("checking if image %s exists", imageIndex)
	imageList, err := c.GetImageList()
	if err != nil {
		return false, err
	}
	return contains(imageList, imageIndex), nil
}

// GetImageID retrieves the unique image ID for a given image tag
// Parameters:
//   - imageTag: image tag or name to get the ID for
//
// Returns:
//   - string: image ID if found, empty string if not found
//   - error: error if command fails, nil on success
func (c *PodmanCLI) GetImageID(imageTag string) (string, error) {
	if imageTag == "" {
		return "", fmt.Errorf("image tag cannot be empty")
	}

	imageID, err := c.Run("images").Args(imageTag, "--format", "{{.ID}}").Output()
	if err != nil {
		e2e.Logf("Failed to run 'podman images --format {{.ID}}' for tag %s", imageTag)
		return "", err
	}
	return imageID, nil
}

// RemoveImage removes an image from the podman registry by image name or tag
// Parameters:
//   - imageIndex: image name, tag, or identifier to remove
//
// Returns:
//   - bool: true if removal was successful or image didn't exist, false on failure
//   - error: error if removal fails, nil on success
func (c *PodmanCLI) RemoveImage(imageIndex string) (bool, error) {
	if imageIndex == "" {
		return false, fmt.Errorf("image name cannot be empty")
	}

	imageID, err := c.GetImageID(imageIndex)
	if err != nil {
		return false, err
	}
	if imageID == "" {
		e2e.Logf("image %s not found, considering as already removed", imageIndex)
		return true, nil
	}

	e2e.Logf("removing image with ID: %s", imageID)
	_, err = c.Run("image").Args("rm", "-f", imageID).Output()
	if err != nil {
		e2e.Logf("failed to remove image %s: %v", imageID, err)
		return false, err
	}
	e2e.Logf("successfully removed image %s", imageID)
	return true, nil
}

// ContainerCreate creates a new container from the specified image with given configuration
// Parameters:
//   - imageName: name of the image to create container from
//   - containerName: name to assign to the new container
//   - entrypoint: custom entrypoint command for the container
//   - openStdin: whether to keep stdin open for interactive use
//
// Returns:
//   - string: container ID of the created container
//   - error: error if container creation fails, nil on success
func (c *PodmanCLI) ContainerCreate(imageName string, containerName string, entrypoint string, openStdin bool) (string, error) {
	if imageName == "" {
		return "", fmt.Errorf("image name cannot be empty")
	}
	if containerName == "" {
		return "", fmt.Errorf("container name cannot be empty")
	}

	interactiveStr := "--interactive=false"
	if openStdin {
		interactiveStr = "--interactive=true"
	}

	args := []string{interactiveStr}
	if entrypoint != "" {
		args = append(args, "--entrypoint="+entrypoint)
	}
	args = append(args, "--name="+containerName, imageName)

	output, err := c.Run("create").Args(args...).Output()
	if err != nil {
		e2e.Logf("failed to run podman create: %v", err)
		return "", err
	}

	if output == "" {
		return "", fmt.Errorf("empty output from podman create command")
	}

	outputLines := strings.Split(strings.TrimSpace(output), "\n")
	if len(outputLines) == 0 {
		return "", fmt.Errorf("no container ID returned from podman create")
	}

	containerID := outputLines[len(outputLines)-1]
	return containerID, nil
}

// ContainerStart starts a previously created container
// Parameters:
//   - id: container ID or name to start
//
// Returns:
//   - error: error if container start fails, nil on success
func (c *PodmanCLI) ContainerStart(id string) error {
	if id == "" {
		return fmt.Errorf("container ID cannot be empty")
	}

	_, err := c.Run("start").Args(id).Output()
	if err != nil {
		e2e.Logf("failed to start container %s: %v", id, err)
	}
	return err
}

// ContainerStop stops a running container
// Parameters:
//   - id: container ID or name to stop
//
// Returns:
//   - error: error if container stop fails, nil on success
func (c *PodmanCLI) ContainerStop(id string) error {
	if id == "" {
		return fmt.Errorf("container ID cannot be empty")
	}

	_, err := c.Run("stop").Args(id).Output()
	if err != nil {
		e2e.Logf("failed to stop container %s: %v", id, err)
	}
	return err
}

// ContainerRemove forcefully removes a container from the system
// Parameters:
//   - id: container ID or name to remove
//
// Returns:
//   - error: error if container removal fails, nil on success
func (c *PodmanCLI) ContainerRemove(id string) error {
	if id == "" {
		return fmt.Errorf("container ID cannot be empty")
	}

	_, err := c.Run("rm").Args(id, "-f").Output()
	if err != nil {
		e2e.Logf("failed to remove container %s: %v", id, err)
	}
	return err
}

// Exec executes a command inside a running container and waits for completion
// Parameters:
//   - id: container ID or name to execute command in
//   - commands: slice of command and arguments to execute
//
// Returns:
//   - string: output from the executed command
//   - error: error if command execution fails, nil on success
func (c *PodmanCLI) Exec(id string, commands []string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("container ID cannot be empty")
	}
	if len(commands) == 0 {
		return "", fmt.Errorf("commands cannot be empty")
	}

	execArgs := append([]string{id}, commands...)
	output, err := c.Run("exec").Args(execArgs...).Output()
	if err != nil {
		e2e.Logf("failed to run podman exec %v: %v", execArgs, err)
		return "", err
	}
	return output, nil
}

// ExecBackground executes a command inside a running container in detached mode (background)
// Parameters:
//   - id: container ID or name to execute command in
//   - commands: slice of command and arguments to execute
//
// Returns:
//   - string: output from the command execution (typically execution ID)
//   - error: error if command execution fails, nil on success
func (c *PodmanCLI) ExecBackground(id string, commands []string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("container ID cannot be empty")
	}
	if len(commands) == 0 {
		return "", fmt.Errorf("commands cannot be empty")
	}

	execArgs := append([]string{"--detach", id}, commands...)
	output, err := c.Run("exec").Args(execArgs...).Output()
	if err != nil {
		e2e.Logf("failed to run podman exec in background %v: %v", execArgs, err)
		return "", err
	}
	return output, nil
}

// CopyFile copies a file from the host filesystem to a container
// Parameters:
//   - id: container ID or name to copy file to
//   - src: source file path on the host
//   - target: destination path inside the container
//
// Returns:
//   - error: error if file copy fails, nil on success
func (c *PodmanCLI) CopyFile(id string, src string, target string) error {
	if id == "" {
		return fmt.Errorf("container ID cannot be empty")
	}
	if src == "" {
		return fmt.Errorf("source path cannot be empty")
	}
	if target == "" {
		return fmt.Errorf("target path cannot be empty")
	}

	_, err := c.Run("cp").Args(src, id+":"+target).Output()
	if err != nil {
		e2e.Logf("failed to copy file from %s to %s:%s: %v", src, id, target, err)
	}
	return err
}
