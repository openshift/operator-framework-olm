package server

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSchemaVersionRegex(t *testing.T) {
	tt := []struct {
		name    string
		input   string
		matches bool
		version string
	}{
		{
			name:    "v1",
			input:   "io.openshift.operators.lifecycles.v1",
			matches: true,
			version: "v1",
		},
		{
			name:    "v1alpha1",
			input:   "io.openshift.operators.lifecycles.v1alpha1",
			matches: true,
			version: "v1alpha1",
		},
		{
			name:    "v1beta1",
			input:   "io.openshift.operators.lifecycles.v1beta1",
			matches: true,
			version: "v1beta1",
		},
		{
			name:    "v2beta3",
			input:   "io.openshift.operators.lifecycles.v2beta3",
			matches: true,
			version: "v2beta3",
		},
		{
			name:    "v200beta300",
			input:   "io.openshift.operators.lifecycles.v200beta300",
			matches: true,
			version: "v200beta300",
		},
		{
			name:    "missing v prefix",
			input:   "io.openshift.operators.lifecycles.1",
			matches: false,
		},
		{
			name:    "v0 not allowed",
			input:   "io.openshift.operators.lifecycles.v0",
			matches: false,
		},
		{
			name:    "v1beta0 not allowed",
			input:   "io.openshift.operators.lifecycles.v1beta0",
			matches: false,
		},
		{
			name:    "v0alpha1 not allowed",
			input:   "io.openshift.operators.lifecycles.v0alpha1",
			matches: false,
		},
		{
			name:    "random schema",
			input:   "olm.package",
			matches: false,
		},
		{
			name:    "empty string",
			input:   "",
			matches: false,
		},
		{
			name:    "partial prefix match",
			input:   "io.openshift.operators.lifecycles.",
			matches: false,
		},
		{
			name:    "wrong prefix",
			input:   "io.openshift.operators.lifecycle.v1",
			matches: false,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			matches := schemaVersionRegex.FindStringSubmatch(tc.input)
			if tc.matches {
				require.NotNil(t, matches, "expected %q to match", tc.input)
				require.Equal(t, tc.version, matches[1])
			} else {
				require.Nil(t, matches, "expected %q not to match", tc.input)
			}
		})
	}
}

// writeFBCFile writes a JSON file containing FBC meta objects to the given directory.
// Each object must be a map with "schema", "package", and other fields that become the blob.
func writeFBCFile(t *testing.T, dir, filename string, objects ...map[string]any) {
	t.Helper()
	var data []byte
	for _, obj := range objects {
		b, err := json.Marshal(obj)
		require.NoError(t, err)
		data = append(data, b...)
		data = append(data, '\n')
	}
	err := os.WriteFile(filepath.Join(dir, filename), data, 0644)
	require.NoError(t, err)
}

func TestLoadLifecycleData(t *testing.T) {
	tt := []struct {
		name          string
		setup         func(t *testing.T) string
		expectedIndex LifecycleIndex
		expectErr     bool
	}{
		{
			name: "non-existent path returns empty index",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "does-not-exist")
			},
			expectedIndex: LifecycleIndex{},
		},
		{
			name: "empty directory returns empty index",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			expectedIndex: LifecycleIndex{},
		},
		{
			name: "lifecycle blob is indexed correctly",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				writeFBCFile(t, dir, "catalog.json",
					map[string]any{
						"schema":  "io.openshift.operators.lifecycles.v1alpha1",
						"package": "my-operator",
						"data":    "test-value",
					},
				)
				return dir
			},
			expectedIndex: LifecycleIndex{
				"v1alpha1": {
					"my-operator": json.RawMessage(`{"data":"test-value","package":"my-operator","schema":"io.openshift.operators.lifecycles.v1alpha1"}`),
				},
			},
		},
		{
			name: "non-lifecycle schemas are skipped",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				writeFBCFile(t, dir, "catalog.json",
					map[string]any{
						"schema":  "olm.package",
						"package": "my-operator",
						"name":    "my-operator",
					},
					map[string]any{
						"schema":  "olm.channel",
						"package": "my-operator",
						"name":    "stable",
					},
				)
				return dir
			},
			expectedIndex: LifecycleIndex{},
		},
		{
			name: "multiple versions and packages",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				writeFBCFile(t, dir, "catalog.json",
					map[string]any{
						"schema":  "io.openshift.operators.lifecycles.v1alpha1",
						"package": "operator-a",
						"status":  "active",
					},
					map[string]any{
						"schema":  "io.openshift.operators.lifecycles.v1alpha1",
						"package": "operator-b",
						"status":  "deprecated",
					},
					map[string]any{
						"schema":  "io.openshift.operators.lifecycles.v1",
						"package": "operator-a",
						"level":   "ga",
					},
				)
				return dir
			},
			expectedIndex: LifecycleIndex{
				"v1alpha1": {
					"operator-a": json.RawMessage(`{"package":"operator-a","schema":"io.openshift.operators.lifecycles.v1alpha1","status":"active"}`),
					"operator-b": json.RawMessage(`{"package":"operator-b","schema":"io.openshift.operators.lifecycles.v1alpha1","status":"deprecated"}`),
				},
				"v1": {
					"operator-a": json.RawMessage(`{"level":"ga","package":"operator-a","schema":"io.openshift.operators.lifecycles.v1"}`),
				},
			},
		},
		{
			name: "empty package name is skipped",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				writeFBCFile(t, dir, "catalog.json",
					map[string]any{
						"schema": "io.openshift.operators.lifecycles.v1alpha1",
						"data":   "should-be-skipped",
					},
				)
				return dir
			},
			expectedIndex: LifecycleIndex{},
		},
		{
			name: "mixed lifecycle and non-lifecycle schemas",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				writeFBCFile(t, dir, "catalog.json",
					map[string]any{
						"schema":  "olm.package",
						"package": "my-operator",
						"name":    "my-operator",
					},
					map[string]any{
						"schema":  "io.openshift.operators.lifecycles.v1alpha1",
						"package": "my-operator",
						"eol":     "2025-12-31",
					},
				)
				return dir
			},
			expectedIndex: LifecycleIndex{
				"v1alpha1": {
					"my-operator": json.RawMessage(`{"eol":"2025-12-31","package":"my-operator","schema":"io.openshift.operators.lifecycles.v1alpha1"}`),
				},
			},
		},
		{
			name: "corrupted entries are silently skipped, valid entries still loaded",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				// Write a valid lifecycle blob
				writeFBCFile(t, dir, "valid.json",
					map[string]any{
						"schema":  "io.openshift.operators.lifecycles.v1alpha1",
						"package": "good-operator",
						"status":  "active",
					},
				)
				// Write a file with invalid JSON (corrupted entry)
				err := os.WriteFile(filepath.Join(dir, "corrupted.json"), []byte("not valid json{{{"), 0644)
				require.NoError(t, err)
				return dir
			},
			// WalkMetasFS passes per-meta errors to the callback, where LoadLifecycleData
			// silently skips them (fbc.go:53-54). No error is returned overall, and
			// valid entries from other files are still loaded successfully.
			expectedIndex: LifecycleIndex{
				"v1alpha1": {
					"good-operator": json.RawMessage(`{"package":"good-operator","schema":"io.openshift.operators.lifecycles.v1alpha1","status":"active"}`),
				},
			},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			path := tc.setup(t)
			result, err := LoadLifecycleData(path)
			if tc.expectErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Compare version keys
			require.Equal(t, len(tc.expectedIndex), len(result), "version count mismatch")
			for version, expectedPkgs := range tc.expectedIndex {
				resultPkgs, ok := result[version]
				require.True(t, ok, "missing version %q in result", version)
				require.Equal(t, len(expectedPkgs), len(resultPkgs), "package count mismatch for version %q", version)
				for pkg, expectedBlob := range expectedPkgs {
					resultBlob, ok := resultPkgs[pkg]
					require.True(t, ok, "missing package %q in version %q", pkg, version)
					// Compare as unmarshalled maps since JSON key order is not guaranteed
					var expectedMap, resultMap map[string]any
					require.NoError(t, json.Unmarshal(expectedBlob, &expectedMap))
					require.NoError(t, json.Unmarshal(resultBlob, &resultMap))
					require.Equal(t, expectedMap, resultMap)
				}
			}
		})
	}
}

func TestLifecycleIndex_CountBlobs(t *testing.T) {
	tt := []struct {
		name     string
		index    LifecycleIndex
		expected int
	}{
		{
			name:     "empty index",
			index:    LifecycleIndex{},
			expected: 0,
		},
		{
			name: "single version single package",
			index: LifecycleIndex{
				"v1": {"pkg-a": json.RawMessage(`{}`)},
			},
			expected: 1,
		},
		{
			name: "multiple versions and packages",
			index: LifecycleIndex{
				"v1alpha1": {
					"pkg-a": json.RawMessage(`{}`),
					"pkg-b": json.RawMessage(`{}`),
				},
				"v1": {
					"pkg-a": json.RawMessage(`{}`),
				},
			},
			expected: 3,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.index.CountBlobs())
		})
	}
}

func TestLifecycleIndex_CountPackages(t *testing.T) {
	tt := []struct {
		name     string
		index    LifecycleIndex
		expected int
	}{
		{
			name:     "empty index",
			index:    LifecycleIndex{},
			expected: 0,
		},
		{
			name: "same package across versions counted once",
			index: LifecycleIndex{
				"v1alpha1": {"pkg-a": json.RawMessage(`{}`)},
				"v1":       {"pkg-a": json.RawMessage(`{}`)},
			},
			expected: 1,
		},
		{
			name: "different packages counted separately",
			index: LifecycleIndex{
				"v1alpha1": {
					"pkg-a": json.RawMessage(`{}`),
					"pkg-b": json.RawMessage(`{}`),
				},
				"v1": {
					"pkg-c": json.RawMessage(`{}`),
				},
			},
			expected: 3,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, tc.index.CountPackages())
		})
	}
}

func TestLifecycleIndex_ListVersions(t *testing.T) {
	tt := []struct {
		name     string
		index    LifecycleIndex
		expected []string
	}{
		{
			name:     "empty index",
			index:    LifecycleIndex{},
			expected: []string{},
		},
		{
			name: "multiple versions",
			index: LifecycleIndex{
				"v1alpha1": {"pkg-a": json.RawMessage(`{}`)},
				"v1":       {"pkg-a": json.RawMessage(`{}`)},
				"v2beta1":  {"pkg-b": json.RawMessage(`{}`)},
			},
			expected: []string{"v1", "v1alpha1", "v2beta1"},
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.index.ListVersions()
			sort.Strings(result)
			sort.Strings(tc.expected)
			require.Equal(t, tc.expected, result)
		})
	}
}
