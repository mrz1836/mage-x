package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testRelease is a simple struct for testing HTTPGetJSON
type testRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Prerelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
}

// TestHTTPGetJSON_Success tests successful JSON fetching and decoding
func TestHTTPGetJSON_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tag_name":"v1.0.0","name":"Release 1.0.0","prerelease":false,"draft":false}`)) //nolint:errcheck // test server write never fails
	}))
	defer server.Close()

	result, err := HTTPGetJSON[testRelease](server.URL, 5*time.Second)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "v1.0.0", result.TagName)
	assert.Equal(t, "Release 1.0.0", result.Name)
	assert.False(t, result.Prerelease)
	assert.False(t, result.Draft)
}

// TestHTTPGetJSON_SliceResult tests fetching a slice of JSON objects
func TestHTTPGetJSON_SliceResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"tag_name":"v2.0.0","prerelease":true},{"tag_name":"v1.0.0","prerelease":false}]`)) //nolint:errcheck // test server write never fails
	}))
	defer server.Close()

	result, err := HTTPGetJSON[[]testRelease](server.URL, 5*time.Second)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, *result, 2)
	assert.Equal(t, "v2.0.0", (*result)[0].TagName)
	assert.True(t, (*result)[0].Prerelease)
	assert.Equal(t, "v1.0.0", (*result)[1].TagName)
	assert.False(t, (*result)[1].Prerelease)
}

// TestHTTPGetJSON_HTTPError tests error handling for non-200 status codes
func TestHTTPGetJSON_HTTPError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{
			name:       "404 Not Found",
			statusCode: http.StatusNotFound,
			body:       `{"message":"Not Found"}`,
		},
		{
			name:       "500 Internal Server Error",
			statusCode: http.StatusInternalServerError,
			body:       `{"error":"server error"}`,
		},
		{
			name:       "403 Forbidden",
			statusCode: http.StatusForbidden,
			body:       `{"message":"rate limit exceeded"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body)) //nolint:errcheck // test server write never fails
			}))
			defer server.Close()

			result, err := HTTPGetJSON[testRelease](server.URL, 5*time.Second)
			require.Error(t, err)
			assert.Nil(t, result)
			require.ErrorIs(t, err, ErrHTTPAPIError)
			assert.Contains(t, err.Error(), tt.body)
			// Verify URL is included in error message
			assert.Contains(t, err.Error(), server.URL)
		})
	}
}

// TestHTTPGetJSON_ErrorContainsURL tests that errors include the URL for debugging
func TestHTTPGetJSON_ErrorContainsURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Not Found"}`)) //nolint:errcheck // test server write never fails
	}))
	defer server.Close()

	_, err := HTTPGetJSON[testRelease](server.URL, 5*time.Second)
	require.Error(t, err)
	// Verify error message includes the URL for debugging
	assert.Contains(t, err.Error(), "GET")
	assert.Contains(t, err.Error(), server.URL)
}

// TestHTTPGetJSON_InvalidJSON tests error handling for malformed JSON
func TestHTTPGetJSON_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`)) //nolint:errcheck // test server write never fails
	}))
	defer server.Close()

	result, err := HTTPGetJSON[testRelease](server.URL, 5*time.Second)
	require.Error(t, err)
	assert.Nil(t, result)
}

// TestHTTPGetJSON_Timeout tests timeout handling
func TestHTTPGetJSON_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Sleep longer than the timeout
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`)) //nolint:errcheck // test server write never fails
	}))
	defer server.Close()

	result, err := HTTPGetJSON[testRelease](server.URL, 50*time.Millisecond)
	require.Error(t, err)
	assert.Nil(t, result)
}

// TestHTTPGetJSON_EmptyBody tests handling of empty response body
func TestHTTPGetJSON_EmptyBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Empty body - should fail to decode
	}))
	defer server.Close()

	result, err := HTTPGetJSON[testRelease](server.URL, 5*time.Second)
	require.Error(t, err)
	assert.Nil(t, result)
}
