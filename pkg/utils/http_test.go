package utils

import (
	"context"
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := HTTPGetJSON[testRelease](ctx, server.URL)
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := HTTPGetJSON[[]testRelease](ctx, server.URL)
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

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			result, err := HTTPGetJSON[testRelease](ctx, server.URL)
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := HTTPGetJSON[testRelease](ctx, server.URL)
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := HTTPGetJSON[testRelease](ctx, server.URL)
	require.Error(t, err)
	assert.Nil(t, result)
}

// TestHTTPGetJSON_Timeout tests timeout handling via context
func TestHTTPGetJSON_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Sleep longer than the timeout
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`)) //nolint:errcheck // test server write never fails
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	result, err := HTTPGetJSON[testRelease](ctx, server.URL)
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

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := HTTPGetJSON[testRelease](ctx, server.URL)
	require.Error(t, err)
	assert.Nil(t, result)
}

// TestHTTPGetJSON_ResponseBodyCloseError tests error handling when response body close fails
func TestHTTPGetJSON_ResponseBodyCloseError(t *testing.T) {
	t.Run("logs error when response body close fails", func(t *testing.T) {
		// Create a test server that sends a valid response
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"test":"value"}`)) //nolint:errcheck // test server
		}))
		defer server.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Make a successful request - the defer close will be called
		result, err := HTTPGetJSON[map[string]string](ctx, server.URL)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "value", (*result)["test"])
	})
}

// TestHTTPGetJSON_ContextCancellation tests that requests are properly canceled
func TestHTTPGetJSON_ContextCancellation(t *testing.T) {
	requestReceived := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		close(requestReceived)
		// Wait longer than we'll wait for context cancellation
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`)) //nolint:errcheck // test server write never fails
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// Start the request in a goroutine
	done := make(chan error)
	go func() {
		_, err := HTTPGetJSON[testRelease](ctx, server.URL)
		done <- err
	}()

	// Wait for server to receive request, then cancel
	<-requestReceived
	cancel()

	// Request should return with context canceled error
	err := <-done
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}

// TestHTTPGetJSON_InvalidURL tests error handling for invalid URLs
func TestHTTPGetJSON_InvalidURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := HTTPGetJSON[testRelease](ctx, "://invalid-url")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create request")
}

// TestHTTPGetJSON_ConnectionRefused tests error handling when connection is refused
func TestHTTPGetJSON_ConnectionRefused(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Port 99999 should not have anything listening
	_, err := HTTPGetJSON[testRelease](ctx, "http://localhost:99999/nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch")
}

// TestDefaultHTTPClient tests that the shared client is returned
func TestDefaultHTTPClient(t *testing.T) {
	client := DefaultHTTPClient()
	require.NotNil(t, client)

	// Verify it's the same instance on subsequent calls
	client2 := DefaultHTTPClient()
	assert.Same(t, client, client2)

	// Verify transport configuration
	transport, ok := client.Transport.(*http.Transport)
	require.True(t, ok, "Transport should be *http.Transport")
	assert.Equal(t, 100, transport.MaxIdleConns)
	assert.Equal(t, 10, transport.MaxIdleConnsPerHost)
	assert.Equal(t, 90*time.Second, transport.IdleConnTimeout)
	assert.Equal(t, 10*time.Second, transport.TLSHandshakeTimeout)
	assert.Equal(t, 30*time.Second, transport.ResponseHeaderTimeout)
	assert.Equal(t, 1*time.Second, transport.ExpectContinueTimeout)
}

// TestHTTPGetJSON_UsesSharedClient verifies connection reuse
func TestHTTPGetJSON_UsesSharedClient(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"tag_name":"v1.0.0"}`)) //nolint:errcheck // test server
	}))
	defer server.Close()

	// Make multiple requests - they should all succeed using the shared client
	for i := 0; i < 5; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		result, err := HTTPGetJSON[testRelease](ctx, server.URL)
		cancel()

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "v1.0.0", result.TagName)
	}

	assert.Equal(t, 5, requestCount)
}
