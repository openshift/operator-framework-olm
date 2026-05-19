package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
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
			require.Equal(t, tc.expectedStatus, resp.StatusCode, "unexpected status code")

			if tc.expectedBody != "" {
				body, err := io.ReadAll(resp.Body)
				require.NoError(t, err, "failed to read response body")
				require.Equal(t, tc.expectedBody, string(body), "unexpected response body")
			}

			if tc.expectedCT != "" {
				require.Equal(t, tc.expectedCT, resp.Header.Get("Content-Type"), "unexpected Content-Type header")
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
	require.NoError(t, err, "failed to read response body")
	require.Equal(t, string(originalBlob), string(body), "response body should be byte-for-byte identical to the stored blob")
}

func TestNewHandler_ConcurrentRequests(t *testing.T) {
	testBlob := json.RawMessage(`{"status":"active","eol":"2025-12-31"}`)
	data := LifecycleIndex{
		"v1alpha1": {
			"my-operator": testBlob,
		},
	}
	handler := NewHandler(data, logr.Discard())

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)

	errCh := make(chan error, goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/api/v1alpha1/lifecycles/my-operator", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				errCh <- fmt.Errorf("expected status 200, got %d", resp.StatusCode)
				return
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				errCh <- fmt.Errorf("failed to read body: %w", err)
				return
			}
			if string(body) != string(testBlob) {
				errCh <- fmt.Errorf("body mismatch: got %q, want %q", string(body), string(testBlob))
			}
		}()
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}

func TestNewHealthHandler(t *testing.T) {
	tt := []struct {
		name           string
		data           LifecycleIndex
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "healthz always returns 200",
			data:           LifecycleIndex{},
			path:           "/healthz",
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
		{
			name:           "healthz returns 200 with data",
			data:           LifecycleIndex{"v1": {"pkg": json.RawMessage(`{}`)}},
			path:           "/healthz",
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
		{
			name:           "readyz returns 503 when empty",
			data:           LifecycleIndex{},
			path:           "/readyz",
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   "no lifecycle data loaded",
		},
		{
			name:           "readyz returns 503 when nil",
			data:           nil,
			path:           "/readyz",
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   "no lifecycle data loaded",
		},
		{
			name:           "readyz returns 200 when data loaded",
			data:           LifecycleIndex{"v1alpha1": {"my-operator": json.RawMessage(`{}`)}},
			path:           "/readyz",
			expectedStatus: http.StatusOK,
			expectedBody:   "ok",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			handler := NewHealthHandler(tc.data)
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			resp := rec.Result()
			defer resp.Body.Close()
			require.Equal(t, tc.expectedStatus, resp.StatusCode, "unexpected status code")

			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err, "failed to read response body")
			require.Contains(t, string(body), tc.expectedBody, "unexpected response body")
		})
	}
}

func TestNewHandler_MultipleVersions(t *testing.T) {
	blobV1Alpha1 := json.RawMessage(`{"version":"v1alpha1","status":"active"}`)
	blobV1Beta1 := json.RawMessage(`{"version":"v1beta1","status":"deprecated"}`)

	data := LifecycleIndex{
		"v1alpha1": {"my-operator": blobV1Alpha1},
		"v1beta1":  {"my-operator": blobV1Beta1},
	}
	handler := NewHandler(data, logr.Discard())

	// Query v1alpha1
	req := httptest.NewRequest(http.MethodGet, "/api/v1alpha1/lifecycles/my-operator", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	resp := rec.Result()
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode, "v1alpha1 request should return 200")
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "failed to read v1alpha1 response body")
	require.Equal(t, string(blobV1Alpha1), string(body), "v1alpha1 response body mismatch")

	// Query v1beta1
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1beta1/lifecycles/my-operator", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	resp2 := rec2.Result()
	defer resp2.Body.Close()
	require.Equal(t, http.StatusOK, resp2.StatusCode, "v1beta1 request should return 200")
	body2, err := io.ReadAll(resp2.Body)
	require.NoError(t, err, "failed to read v1beta1 response body")
	require.Equal(t, string(blobV1Beta1), string(body2), "v1beta1 response body mismatch")
}
