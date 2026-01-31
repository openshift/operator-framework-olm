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

	"github.com/go-logr/logr"
)

// NewHandler creates a new HTTP handler for the lifecycle API
func NewHandler(data LifecycleIndex, log logr.Logger) http.Handler {
	mux := http.NewServeMux()

	// GET /api/{version}/lifecycles/{package}
	mux.HandleFunc("GET /api/{version}/lifecycles/{package}", func(w http.ResponseWriter, r *http.Request) {
		version := r.PathValue("version")
		pkg := r.PathValue("package")

		// If no lifecycle data is available, return 503 Service Unavailable
		if len(data) == 0 {
			log.V(1).Info("no lifecycle data available, returning 503")
			http.Error(w, "No lifecycle data available", http.StatusServiceUnavailable)
			return
		}

		// Look up version in index
		versionData, ok := data[version]
		if !ok {
			log.V(1).Info("version not found", "version", version, "package", pkg)
			http.NotFound(w, r)
			return
		}

		// Look up package in version
		rawJSON, ok := versionData[pkg]
		if !ok {
			log.V(1).Info("package not found", "version", version, "package", pkg)
			http.NotFound(w, r)
			return
		}

		log.V(1).Info("returning lifecycle data", "version", version, "package", pkg)

		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(rawJSON); err != nil {
			log.V(1).Error(err, "failed to write response")
		}
	})

	return mux
}
