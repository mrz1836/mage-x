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

// HTTPGetJSON fetches JSON from a URL and decodes it into the target type.
// Returns the decoded value or an error with response body details on non-200 status.
func HTTPGetJSON[T any](url string, timeout time.Duration) (*T, error) {
	client := &http.Client{Timeout: timeout}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", url, err)
	}

	resp, err := client.Do(req)
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
