package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	testBlob := json.RawMessage(`{"eol":"2025-12-31","status":"active"}`)

	tt := []struct {
		name           string
		data           LifecycleIndex
		method         string
		path           string
		expectedStatus int
		expectedBody   string
		expectedCT     string
	}{
		{
			name: "valid version and package returns 200 with JSON",
			data: LifecycleIndex{
				"v1alpha1": {
					"my-operator": testBlob,
				},
			},
			method:         http.MethodGet,
			path:           "/api/v1alpha1/lifecycles/my-operator",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"eol":"2025-12-31","status":"active"}`,
			expectedCT:     "application/json",
		},
		{
			name:           "empty data returns 503",
			data:           LifecycleIndex{},
			method:         http.MethodGet,
			path:           "/api/v1alpha1/lifecycles/my-operator",
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name: "unknown version returns 404",
			data: LifecycleIndex{
				"v1alpha1": {
					"my-operator": testBlob,
				},
			},
			method:         http.MethodGet,
			path:           "/api/v2/lifecycles/my-operator",
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "known version unknown package returns 404",
			data: LifecycleIndex{
				"v1alpha1": {
					"my-operator": testBlob,
				},
			},
			method:         http.MethodGet,
			path:           "/api/v1alpha1/lifecycles/other-operator",
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "POST method not allowed",
			data: LifecycleIndex{
				"v1alpha1": {
					"my-operator": testBlob,
				},
			},
			method:         http.MethodPost,
			path:           "/api/v1alpha1/lifecycles/my-operator",
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name: "wrong path returns 404",
			data: LifecycleIndex{
				"v1alpha1": {
					"my-operator": testBlob,
				},
			},
			method:         http.MethodGet,
			path:           "/wrong/path",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "nil data (nil map) returns 503",
			data:           nil,
			method:         http.MethodGet,
			path:           "/api/v1alpha1/lifecycles/my-operator",
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewHandler(tc.data, logr.Discard())

			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()
			require.Equal(t, tc.expectedStatus, resp.StatusCode)

			if tc.expectedBody != "" {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				require.Equal(t, tc.expectedBody, string(body))
			}

			if tc.expectedCT != "" {
				require.Equal(t, tc.expectedCT, resp.Header.Get("Content-Type"))
			}
		})
	}
}

func TestNewHandler_RawBlobReturnedByteForByte(t *testing.T) {
	// Verify that the raw JSON blob is returned exactly as stored, not re-serialized.
	// This matters because the handler writes rawJSON directly with w.Write(rawJSON).
	originalBlob := json.RawMessage(`{"keys":"in-specific-order","numbers":42,"nested":{"a":1}}`)

	data := LifecycleIndex{
		"v1alpha1": {
			"test-pkg": originalBlob,
		},
	}

	handler := NewHandler(data, logr.Discard())
	req := httptest.NewRequest(http.MethodGet, "/api/v1alpha1/lifecycles/test-pkg", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	resp := rec.Result()
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, string(originalBlob), string(body), "response body should be byte-for-byte identical to the stored blob")
}
