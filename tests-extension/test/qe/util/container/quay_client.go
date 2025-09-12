package container

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	e2e "k8s.io/kubernetes/test/e2e/framework"
)

// AuthInfo returns the error info
type AuthInfo struct {
	Authorization string `json:"authorization"`
}

// TagInfo returns the images tag info
type TagInfo struct {
	Name           string `json:"name"`
	Reversion      bool   `json:"reversion"`
	StartTs        int64  `json:"start_ts"`
	EndTs          int64  `json:"end_ts"`
	ManifestDigest string `json:"manifest_digest"`
	ImageID        string `json:"image_id"`
	LastModified   string `json:"last_modified"`
	Expiration     string `json:"expiration"`
	DockerImageID  string `json:"docker_image_id"`
	IsManifestList bool   `json:"is_manifest_list"`
	Size           int64  `json:"size"`
}

// TagsResult returns the images tag info
type TagsResult struct {
	HasAdditional bool      `json:"has_additional"`
	Page          int       `json:"page"`
	Tags          []TagInfo `json:"tags"`
}

// QuayCLI provides function to run the quay command
type QuayCLI struct {
	EndPointPre   string
	Authorization string
	client        *http.Client
}

// validImageName validates repository and tag names to prevent injection
var validImageName = regexp.MustCompile(`^[a-z0-9]+(?:[._-][a-z0-9]+)*(?:/[a-z0-9]+(?:[._-][a-z0-9]+)*)*(?::[a-zA-Z0-9._-]+)?$`)

// isValidImageName checks if an image name is safe
func isValidImageName(name string) bool {
	return name != "" && validImageName.MatchString(name)
}

// sanitizeURL validates and escapes URL components
func sanitizeURL(baseURL, path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}
	u, err := url.Parse(baseURL + path)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}
	return u.String(), nil
}

// loadAuthFromFile loads authentication from file and returns auth string
func loadAuthFromFile(authFilepath string) string {
	if _, err := os.Stat(authFilepath); err != nil {
		e2e.Logf("Quay auth file does not exist: %s", authFilepath)
		return ""
	}

	content, err := os.ReadFile(authFilepath)
	if err != nil {
		e2e.Logf("Failed to read auth file %s: %v", authFilepath, err)
		return ""
	}

	var authJSON AuthInfo
	if err := json.Unmarshal(content, &authJSON); err != nil {
		e2e.Logf("Failed to parse auth JSON from %s: %v", authFilepath, err)
		return ""
	}

	if authJSON.Authorization == "" {
		return ""
	}

	e2e.Logf("Successfully loaded auth from file")
	return "Bearer " + authJSON.Authorization
}

// buildTagEndpoint constructs the appropriate API endpoint based on image index format
func (c *QuayCLI) buildTagEndpoint(imageIndex string) (string, error) {
	if strings.Contains(imageIndex, ":") {
		return c.buildRepositoryTagEndpoint(imageIndex)
	}
	if strings.Contains(imageIndex, "/tag/") {
		return c.buildDirectTagEndpoint(imageIndex)
	}
	return "", fmt.Errorf("invalid image format: %s (expected 'repo:tag' or 'repo/tag/')", imageIndex)
}

// buildRepositoryTagEndpoint builds endpoint for repository:tag format
func (c *QuayCLI) buildRepositoryTagEndpoint(imageIndex string) (string, error) {
	parts := strings.SplitN(imageIndex, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid image format: %s", imageIndex)
	}

	indexRepository := parts[0] + "/tag"
	specificTag := parts[1]

	if specificTag == "" {
		// GET /api/v1/repository/{repository}/tag?onlyActiveTags=true
		return sanitizeURL(c.EndPointPre, indexRepository+"?onlyActiveTags=true")
	}

	// GET /api/v1/repository/{repository}/tag?specificTag={tag}
	escapedTag := url.QueryEscape(specificTag)
	return sanitizeURL(c.EndPointPre, indexRepository+"?specificTag="+escapedTag)
}

// buildDirectTagEndpoint builds endpoint for direct tag path format
func (c *QuayCLI) buildDirectTagEndpoint(imageIndex string) (string, error) {
	parts := strings.SplitN(imageIndex, "tag/", 2)
	if len(parts) < 1 {
		return "", fmt.Errorf("invalid tag format: %s", imageIndex)
	}

	tagPath := parts[0] + "tag/"
	return sanitizeURL(c.EndPointPre, tagPath)
}

// NewQuayCLI initializes and returns a new QuayCLI instance with authentication configured from environment or file
// Returns:
//   - *QuayCLI: new QuayCLI instance with endpoint and authorization configured
func NewQuayCLI() *QuayCLI {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	newclient := &QuayCLI{
		EndPointPre: "https://quay.io/api/v1/repository/",
		client:      client,
	}

	authString := ""
	authFilepath := os.Getenv("QUAY_AUTH_FILE")
	if authFilepath == "" {
		// Use user's home directory instead of hardcoded path
		homeDir, err := os.UserHomeDir()
		if err != nil {
			e2e.Logf("Failed to get user home directory: %v", err)
			authFilepath = "/home/cloud-user/.docker/auto/quay_auth.json"
		} else {
			authFilepath = filepath.Join(homeDir, ".docker", "auto", "quay_auth.json")
		}
	}

	authString = loadAuthFromFile(authFilepath)

	// Environment variable takes precedence
	if envAuth := os.Getenv("QUAY_AUTH"); envAuth != "" {
		e2e.Logf("Using auth from QUAY_AUTH environment variable")
		authString = "Bearer " + envAuth
	}

	if authString == "" || authString == "Bearer " {
		e2e.Failf("Quay authentication not found. Set QUAY_AUTH environment variable or provide auth file")
	}

	newclient.Authorization = authString
	return newclient
}

// TryDeleteTag attempts to delete a single image tag from Quay.io registry
// Parameters:
//   - imageIndex: image name with tag in format "repository:tag" or "repository/tag/tagname"
//
// Returns:
//   - bool: true if deletion was successful (HTTP 204), false otherwise
//   - error: error if HTTP request fails, nil on success
func (c *QuayCLI) TryDeleteTag(imageIndex string) (bool, error) {
	if !isValidImageName(imageIndex) {
		return false, fmt.Errorf("invalid image name: %s", imageIndex)
	}

	imageIndex = strings.Replace(imageIndex, ":", "/tag/", 1)
	endpoint, err := sanitizeURL(c.EndPointPre, imageIndex)
	if err != nil {
		return false, fmt.Errorf("failed to construct endpoint: %w", err)
	}
	e2e.Logf("DELETE endpoint: %s", endpoint)

	request, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	if c.Authorization != "" {
		request.Header.Add("Authorization", c.Authorization)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return false, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			e2e.Logf("Failed to close response body: %v", closeErr)
		}
	}()

	if response.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(response.Body)
		e2e.Logf("Delete %s failed, status: %d, body: %s", imageIndex, response.StatusCode, string(body))
		return false, nil
	}
	return true, nil
}

// DeleteTag deletes an image tag from Quay.io registry with retry logic on failure
// Parameters:
//   - imageIndex: image name with tag in format "repository:tag"
//
// Returns:
//   - bool: true if deletion was successful, fails test if deletion fails after retry
//   - error: error from the deletion attempt, nil on success
func (c *QuayCLI) DeleteTag(imageIndex string) (bool, error) {
	if imageIndex == "" {
		return false, fmt.Errorf("image name cannot be empty")
	}

	rc, err := c.TryDeleteTag(imageIndex)
	if !rc && err == nil {
		e2e.Logf("Retrying deletion of %s", imageIndex)
		rc, err = c.TryDeleteTag(imageIndex)
		if !rc {
			e2e.Failf("Failed to delete tag %s on quay.io after retry", imageIndex)
		}
	}
	return rc, err
}

// CheckTagNotExist checks if an image tag does not exist in the Quay.io registry
// Parameters:
//   - imageIndex: image name with tag in format "repository:tag"
//
// Returns:
//   - bool: true if tag does not exist (HTTP 404), false if tag exists
//   - error: error if HTTP request fails, nil on success
func (c *QuayCLI) CheckTagNotExist(imageIndex string) (bool, error) {
	if !isValidImageName(imageIndex) {
		return false, fmt.Errorf("invalid image name: %s", imageIndex)
	}

	imageIndex = strings.Replace(imageIndex, ":", "/tag/", 1)
	endpoint, err := sanitizeURL(c.EndPointPre, imageIndex+"/images")
	if err != nil {
		return false, fmt.Errorf("failed to construct endpoint: %w", err)
	}
	e2e.Logf("GET endpoint: %s", endpoint)

	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	if c.Authorization != "" {
		request.Header.Add("Authorization", c.Authorization)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return false, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			e2e.Logf("Failed to close response body: %v", closeErr)
		}
	}()

	if response.StatusCode == http.StatusNotFound {
		e2e.Logf("Tag %s does not exist", imageIndex)
		return true, nil
	}

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		e2e.Logf("Failed to read response body: %v", err)
	} else {
		e2e.Logf("Response: %s", string(contents))
	}
	return false, nil
}

// GetTagNameList retrieves a list of tag names for a given repository from Quay.io
// Parameters:
//   - imageIndex: repository name or image with tag to get tags for
//
// Returns:
//   - []string: slice of tag names found in the repository
//   - error: error if unable to retrieve tags, nil on success
func (c *QuayCLI) GetTagNameList(imageIndex string) ([]string, error) {
	if imageIndex == "" {
		return nil, fmt.Errorf("image name cannot be empty")
	}

	tags, err := c.GetTags(imageIndex)
	if err != nil {
		return nil, err
	}

	tagNameList := make([]string, 0, len(tags))
	for _, tag := range tags {
		tagNameList = append(tagNameList, tag.Name)
	}
	return tagNameList, nil
}

// GetTags retrieves detailed tag information for a repository or specific tag from Quay.io
// Parameters:
//   - imageIndex: repository name, "repository:tag" for specific tag, or "repository/tag/" for all tags
//
// Returns:
//   - []TagInfo: slice of TagInfo structs containing detailed tag information
//   - error: error if HTTP request fails or JSON parsing fails, nil on success
func (c *QuayCLI) GetTags(imageIndex string) ([]TagInfo, error) {
	if imageIndex == "" {
		return nil, fmt.Errorf("image name cannot be empty")
	}

	endpoint, err := c.buildTagEndpoint(imageIndex)

	if err != nil {
		return nil, fmt.Errorf("failed to construct endpoint: %w", err)
	}

	e2e.Logf("GET endpoint: %s", endpoint)

	request, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.Authorization != "" {
		request.Header.Add("Authorization", c.Authorization)
	}

	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			e2e.Logf("Failed to close response body: %v", closeErr)
		}
	}()

	e2e.Logf("Response status: %s", response.Status)
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		e2e.Logf("Get %s failed, status: %d, body: %s", imageIndex, response.StatusCode, string(body))
		return nil, fmt.Errorf("HTTP %d: %s", response.StatusCode, response.Status)
	}

	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var tagsResult TagsResult
	if err := json.Unmarshal(contents, &tagsResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	return tagsResult.Tags, nil
}

// GetImageDigest retrieves the manifest digest for a specific image tag from Quay.io
// Parameters:
//   - imageIndex: image name with tag in format "repository:tag"
//
// Returns:
//   - string: manifest digest of the specified image tag, empty string if not found
//   - error: error if unable to retrieve tags, nil on success
func (c *QuayCLI) GetImageDigest(imageIndex string) (string, error) {
	if !strings.Contains(imageIndex, ":") {
		return "", fmt.Errorf("image must include tag in format 'repository:tag'")
	}

	parts := strings.SplitN(imageIndex, ":", 2)
	if len(parts) != 2 || parts[1] == "" {
		return "", fmt.Errorf("invalid image format: %s", imageIndex)
	}

	tags, err := c.GetTags(imageIndex)
	if err != nil {
		e2e.Logf("Failed to get tags for digest lookup: %v", err)
		return "", err
	}

	imageTag := parts[1]
	for _, tag := range tags {
		if tag.Name == imageTag {
			if tag.ManifestDigest == "" {
				e2e.Logf("Empty manifest digest for tag %s", imageTag)
				return "", nil
			}
			return tag.ManifestDigest, nil
		}
	}

	e2e.Logf("Manifest digest not found for tag %s", imageTag)
	return "", nil
}

// TryChangeTag attempts to update an image tag to point to a different manifest digest
// Parameters:
//   - imageTag: image tag in format "repository:tag" to update
//   - manifestDigest: new manifest digest to associate with the tag
//
// Returns:
//   - bool: true if tag change was successful (HTTP 201), false otherwise
//   - error: error if HTTP request fails, nil on success
func (c *QuayCLI) TryChangeTag(imageTag, manifestDigest string) (bool, error) {
	if !isValidImageName(imageTag) {
		return false, fmt.Errorf("invalid image tag: %s", imageTag)
	}
	if manifestDigest == "" {
		return false, fmt.Errorf("manifest digest cannot be empty")
	}

	imageTag = strings.Replace(imageTag, ":", "/tag/", 1)
	endpoint, err := sanitizeURL(c.EndPointPre, imageTag)
	if err != nil {
		return false, fmt.Errorf("failed to construct endpoint: %w", err)
	}
	e2e.Logf("PUT endpoint: %s", endpoint)

	payload := map[string]string{
		"manifest_digest": manifestDigest,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return false, fmt.Errorf("failed to marshal payload: %w", err)
	}

	request, err := http.NewRequest("PUT", endpoint, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	if c.Authorization != "" {
		request.Header.Add("Authorization", c.Authorization)
	}
	request.Header.Set("Content-Type", "application/json")

	response, err := c.client.Do(request)
	if err != nil {
		return false, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		if closeErr := response.Body.Close(); closeErr != nil {
			e2e.Logf("Failed to close response body: %v", closeErr)
		}
	}()

	if response.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(response.Body)
		e2e.Logf("Change %s failed, status: %d, body: %s", imageTag, response.StatusCode, string(body))
		return false, nil
	}
	return true, nil
}

// ChangeTag updates an image tag to point to a different manifest digest with retry logic on failure
// Parameters:
//   - imageTag: image tag in format "repository:tag" to update
//   - manifestDigest: new manifest digest to associate with the tag
//
// Returns:
//   - bool: true if tag change was successful, logs failure if unsuccessful after retry
//   - error: error from the change attempt, nil on success
func (c *QuayCLI) ChangeTag(imageTag, manifestDigest string) (bool, error) {
	if imageTag == "" {
		return false, fmt.Errorf("image tag cannot be empty")
	}
	if manifestDigest == "" {
		return false, fmt.Errorf("manifest digest cannot be empty")
	}

	rc, err := c.TryChangeTag(imageTag, manifestDigest)
	if !rc && err == nil {
		e2e.Logf("Retrying tag change for %s", imageTag)
		rc, err = c.TryChangeTag(imageTag, manifestDigest)
		if !rc {
			e2e.Logf("Failed to change tag %s on quay.io after retry", imageTag)
		}
	}
	return rc, err
}
