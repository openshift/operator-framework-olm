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
	"net/http"
	"regexp"

	"github.com/sirupsen/logrus"
)

const apiPrefix = "/api/"

// pathRegex matches /api/<version>/lifecycles/<packageName>.json
// version pattern: v[1-9][0-9]*((?:alpha|beta)[1-9][0-9]*)?
// Matches: v1, v1alpha1, v1beta1, v200beta300
// Does not match: 1, v0, v1beta0
var pathRegex = regexp.MustCompile(`^/api/(v[1-9][0-9]*(?:(?:alpha|beta)[1-9][0-9]*)?)/lifecycles/([^/]+)\.json$`)

// NewHandler creates a new HTTP handler for the lifecycle API
func NewHandler(data LifecycleIndex, log *logrus.Logger) http.Handler {
	mux := http.NewServeMux()

	// Handle GET /api/<version>/lifecycles/<packageName>.json
	mux.HandleFunc(apiPrefix, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// If no lifecycle data is available, return 503 Service Unavailable
		if len(data) == 0 {
			log.Debug("no lifecycle data available, returning 503")
			http.Error(w, "No lifecycle data available", http.StatusServiceUnavailable)
			return
		}

		// Parse the path
		matches := pathRegex.FindStringSubmatch(r.URL.Path)
		if matches == nil {
			http.NotFound(w, r)
			return
		}

		version := matches[1]  // e.g., "v1alpha1"
		pkg := matches[2]      // package name

		// Look up version in index
		versionData, ok := data[version]
		if !ok {
			log.WithFields(logrus.Fields{
				"version": version,
				"package": pkg,
			}).Debug("version not found")
			http.NotFound(w, r)
			return
		}

		// Look up package in version
		rawJSON, ok := versionData[pkg]
		if !ok {
			log.WithFields(logrus.Fields{
				"version": version,
				"package": pkg,
			}).Debug("package not found")
			http.NotFound(w, r)
			return
		}

		log.WithFields(logrus.Fields{
			"version": version,
			"package": pkg,
		}).Debug("returning lifecycle data")

		w.Header().Set("Content-Type", "application/json")
		w.Write(rawJSON)
	})

	// List available versions at /api/
	mux.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Redirect /api to /api/
		if r.URL.Path == "/api" {
			http.Redirect(w, r, "/api/", http.StatusMovedPermanently)
			return
		}
	})

	return mux
}

// CountBlobs returns the total number of blobs in the index
func CountBlobs(index LifecycleIndex) int {
	count := 0
	for _, packages := range index {
		count += len(packages)
	}
	return count
}

// ListVersions returns the list of versions available in the index
func ListVersions(index LifecycleIndex) []string {
	versions := make([]string, 0, len(index))
	for v := range index {
		versions = append(versions, v)
	}
	return versions
}
