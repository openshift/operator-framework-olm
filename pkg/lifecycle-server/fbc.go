/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"context"
	"encoding/json"
	"os"
	"regexp"
	"sync"

	"github.com/operator-framework/operator-registry/alpha/declcfg"
	"k8s.io/apimachinery/pkg/util/sets"
)

// versionPattern matches API versions like v1, v1alpha1, v2beta3
// Matches: v1, v1alpha1, v1beta1, v200beta300
// Does not match: 1, v0, v1beta0
const versionPattern = `v[1-9][0-9]*(?:(?:alpha|beta)[1-9][0-9]*)?`

// schemaVersionRegex matches lifecycle schema versions in FBC blobs
var schemaVersionRegex = regexp.MustCompile(`^io\.openshift\.operators\.lifecycles\.(` + versionPattern + `)$`)

// LifecycleIndex maps schema version -> package name -> raw JSON blob
type LifecycleIndex map[string]map[string]json.RawMessage

// LoadLifecycleData loads lifecycle blobs from FBC files at the given path
func LoadLifecycleData(fbcPath string) (LifecycleIndex, error) {
	result := make(LifecycleIndex)
	var mu sync.Mutex

	// Check if path exists
	if _, err := os.Stat(fbcPath); os.IsNotExist(err) {
		return result, nil
	}

	root := os.DirFS(fbcPath)
	err := declcfg.WalkMetasFS(context.Background(), root, func(path string, meta *declcfg.Meta, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}
		if meta == nil {
			return nil
		}

		// Check if schema matches our pattern
		matches := schemaVersionRegex.FindStringSubmatch(meta.Schema)
		if matches == nil {
			return nil
		}
		schemaVersion := matches[1] // e.g., "v1alpha1"

		if meta.Package == "" {
			return nil
		}

		// Store in index (thread-safe)
		mu.Lock()
		if result[schemaVersion] == nil {
			result[schemaVersion] = make(map[string]json.RawMessage)
		}
		result[schemaVersion][meta.Package] = meta.Blob
		mu.Unlock()

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// CountBlobs returns the total number of blobs in the index
func (index LifecycleIndex) CountBlobs() int {
	count := 0
	for _, packages := range index {
		count += len(packages)
	}
	return count
}

func (index LifecycleIndex) CountPackages() int {
	pkgs := sets.New[string]()
	for _, packages := range index {
		for pkg := range packages {
			pkgs.Insert(pkg)
		}
	}
	return pkgs.Len()
}

// ListVersions returns the list of versions available in the index
func (index LifecycleIndex) ListVersions() []string {
	versions := make([]string, 0, len(index))
	for v := range index {
		versions = append(versions, v)
	}
	return versions
}
