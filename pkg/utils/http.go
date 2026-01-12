// Package utils provides utility functions for HTTP operations
package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// ErrHTTPAPIError is returned when an HTTP API request returns a non-200 status code.
var ErrHTTPAPIError = errors.New("HTTP API error")

// defaultClient is a shared HTTP client with connection pooling for improved performance.
// This avoids creating new TCP/TLS connections for each request, reducing latency by 50-200ms.
//
//nolint:gochecknoglobals // intentional shared client for connection pooling
var defaultClient = &http.Client{
	Timeout: 30 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
}

// DefaultHTTPClient returns the shared HTTP client for reuse across the application.
// Using a shared client enables connection pooling and significantly improves performance
// for repeated requests to the same hosts.
func DefaultHTTPClient() *http.Client {
	return defaultClient
}

// HTTPGetJSON fetches JSON from a URL and decodes it into the target type.
// The context controls cancellation and timeout - use context.WithTimeout for request timeouts.
// Returns the decoded value or an error with response body details on non-200 status.
func HTTPGetJSON[T any](ctx context.Context, url string) (*T, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", url, err)
	}

	resp, err := defaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", url, err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("failed to close response body for %s: %v", url, closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("%w: GET %s returned status %d (body unreadable: %w)", ErrHTTPAPIError, url, resp.StatusCode, readErr)
		}
		return nil, fmt.Errorf("%w: GET %s returned status %d: %s", ErrHTTPAPIError, url, resp.StatusCode, body)
	}

	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode JSON response from %s: %w", url, err)
	}
	return &result, nil
}
