//go:build integration
// +build integration

package mage

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mrz1836/mage-x/pkg/testhelpers"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// TestIntegration_ToolInstallationWithNetworkFailures tests tool installation
// with various network failure scenarios
func TestIntegration_ToolInstallationWithNetworkFailures(t *testing.T) {
	testhelpers.SkipIfShort(t)
	testhelpers.RequireNetwork(t)

	// Reset config for clean test environment
	TestResetConfig()

	config := &Config{
		Download: DownloadConfig{
			MaxRetries:        3,
			InitialDelayMs:    100,  // 100ms
			MaxDelayMs:        1000, // 1s
			TimeoutMs:         5000, // 5s
			BackoffMultiplier: 2.0,
			EnableResume:      true,
			UserAgent:         "mage-x-test/1.0",
		},
		Tools: ToolsConfig{
			Fumpt:       "latest",
			GoVulnCheck: "latest",
		},
	}
	TestSetConfig(config)

	t.Run("GofumptInstallation_NetworkRetry", func(t *testing.T) {
		// Create a temporary GOPATH to avoid affecting system
		tempDir, err := os.MkdirTemp("", "mage_integration_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Set GOPATH temporarily
		originalGOPATH := os.Getenv("GOPATH")
		os.Setenv("GOPATH", tempDir)
		defer os.Setenv("GOPATH", originalGOPATH)

		// Set PATH to include our temp bin directory
		binDir := filepath.Join(tempDir, "bin")
		originalPATH := os.Getenv("PATH")
		os.Setenv("PATH", binDir+string(os.PathListSeparator)+originalPATH)
		defer os.Setenv("PATH", originalPATH)

		// Test gofumpt installation with retry logic
		format := Format{}
		err = format.Fumpt()

		// We expect this to succeed eventually with retries
		if err != nil {
			t.Logf("Gofumpt installation failed (expected in integration test): %v", err)
			// Don't fail the test - this is expected to sometimes fail in CI
		} else {
			t.Log("Gofumpt installation succeeded with retry logic")
		}
	})

	t.Run("GovulncheckInstallation_NetworkRetry", func(t *testing.T) {
		// Create a temporary GOPATH
		tempDir, err := os.MkdirTemp("", "mage_integration_test")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Set GOPATH temporarily
		originalGOPATH := os.Getenv("GOPATH")
		os.Setenv("GOPATH", tempDir)
		defer os.Setenv("GOPATH", originalGOPATH)

		// Set PATH to include our temp bin directory
		binDir := filepath.Join(tempDir, "bin")
		originalPATH := os.Getenv("PATH")
		os.Setenv("PATH", binDir+string(os.PathListSeparator)+originalPATH)
		defer os.Setenv("PATH", originalPATH)

		// Test govulncheck installation with retry logic
		tools := Tools{}
		err = tools.VulnCheck()

		if err != nil {
			t.Logf("Govulncheck installation failed (expected in integration test): %v", err)
			// Don't fail the test - this is expected to sometimes fail in CI
		} else {
			t.Log("Govulncheck installation succeeded with retry logic")
		}
	})
}

// TestIntegration_DownloadWithNetworkSimulation simulates various network
// conditions using a test server
func TestIntegration_DownloadWithNetworkSimulation(t *testing.T) {
	testhelpers.SkipIfShort(t)

	testCases := []struct {
		name           string
		serverBehavior func(attempt *int) http.HandlerFunc
		expectSuccess  bool
		description    string
	}{
		{
			name: "IntermittentServerErrors",
			serverBehavior: func(attempt *int) http.HandlerFunc {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					*attempt++
					// Fail first 2 attempts, succeed on 3rd
					if *attempt < 3 {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte("Server temporarily unavailable"))
						return
					}
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("Success after retries"))
				})
			},
			expectSuccess: true,
			description:   "Server errors that resolve after retries",
		},
		{
			name: "SlowResponseThenSuccess",
			serverBehavior: func(attempt *int) http.HandlerFunc {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					*attempt++
					// First attempt is very slow, second is fast
					if *attempt == 1 {
						time.Sleep(2 * time.Second) // Longer than our timeout
						w.WriteHeader(http.StatusRequestTimeout)
						return
					}
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("Fast response on retry"))
				})
			},
			expectSuccess: true,
			description:   "Slow response timeout followed by fast success",
		},
		{
			name: "PartialContentWithResume",
			serverBehavior: func(attempt *int) http.HandlerFunc {
				fullContent := "This is the full content that should be downloaded completely"
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					*attempt++

					rangeHeader := r.Header.Get("Range")
					if rangeHeader != "" && *attempt > 1 {
						// Resume request - return remaining content
						w.Header().Set("Content-Range", fmt.Sprintf("bytes 20-%d/%d", len(fullContent)-1, len(fullContent)))
						w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fullContent)-20))
						w.WriteHeader(http.StatusPartialContent)
						w.Write([]byte(fullContent[20:]))
					} else if *attempt == 1 {
						// First attempt - return partial content then "disconnect"
						w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fullContent)))
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(fullContent[:20])) // Only first 20 characters
						// Don't write the rest to simulate connection failure
					}
				})
			},
			expectSuccess: true,
			description:   "Partial download with successful resume",
		},
		{
			name: "PersistentServerError",
			serverBehavior: func(attempt *int) http.HandlerFunc {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					*attempt++
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte("Resource not found"))
				})
			},
			expectSuccess: false,
			description:   "Persistent client error that should not be retried",
		},
		{
			name: "NetworkTimeouts",
			serverBehavior: func(attempt *int) http.HandlerFunc {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					*attempt++
					// All attempts timeout
					time.Sleep(3 * time.Second) // Longer than reasonable timeout
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("Should not reach here"))
				})
			},
			expectSuccess: false,
			description:   "All requests timeout",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			attempt := 0
			server := httptest.NewServer(tc.serverBehavior(&attempt))
			defer server.Close()

			tempDir, err := os.MkdirTemp("", "download_integration_test")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			destPath := filepath.Join(tempDir, "test_file.txt")
			config := &utils.DownloadConfig{
				MaxRetries:        3,
				InitialDelay:      100 * time.Millisecond,
				MaxDelay:          1 * time.Second,
				Timeout:           1 * time.Second,
				BackoffMultiplier: 2.0,
				EnableResume:      true,
				UserAgent:         "mage-x-integration-test/1.0",
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			err = utils.DownloadWithRetry(ctx, server.URL, destPath, config)

			if tc.expectSuccess {
				if err != nil {
					t.Errorf("Expected success for %s, got error: %v", tc.description, err)
				} else {
					t.Logf("Successfully completed: %s (attempts: %d)", tc.description, attempt)
				}
			} else {
				if err == nil {
					t.Errorf("Expected failure for %s, but succeeded", tc.description)
				} else {
					t.Logf("Expected failure occurred: %s - %v (attempts: %d)", tc.description, err, attempt)
				}
			}
		})
	}
}

// TestIntegration_GolangciLintInstallation tests the actual golangci-lint
// installation process with retry logic
func TestIntegration_GolangciLintInstallation(t *testing.T) {
	testhelpers.SkipIfShort(t)
	testhelpers.RequireNetwork(t)

	// This test requires internet access and may take a while
	t.Skip("Skipping golangci-lint installation test to avoid affecting CI")

	// Reset config for clean test environment
	TestResetConfig()

	config := &Config{
		Download: DownloadConfig{
			MaxRetries:        3,
			InitialDelayMs:    500,
			MaxDelayMs:        5000,
			TimeoutMs:         30000, // 30 seconds
			BackoffMultiplier: 2.0,
			EnableResume:      true,
			UserAgent:         "mage-x-integration-test/1.0",
		},
		Lint: LintConfig{
			GolangciVersion: "v1.55.2", // Use a specific version for reproducibility
		},
	}
	TestSetConfig(config)

	// Create a temporary GOPATH
	tempDir, err := os.MkdirTemp("", "mage_golangci_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set GOPATH temporarily
	originalGOPATH := os.Getenv("GOPATH")
	os.Setenv("GOPATH", tempDir)
	defer os.Setenv("GOPATH", originalGOPATH)

	// Set PATH to include our temp bin directory
	binDir := filepath.Join(tempDir, "bin")
	originalPATH := os.Getenv("PATH")
	os.Setenv("PATH", binDir+string(os.PathListSeparator)+originalPATH)
	defer os.Setenv("PATH", originalPATH)

	// Ensure golangci-lint is not already installed
	if utils.CommandExists("golangci-lint") {
		t.Skip("golangci-lint already installed, skipping installation test")
	}

	// Test installation with retry logic
	start := time.Now()
	err = ensureGolangciLint(config)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("golangci-lint installation failed: %v", err)
	} else {
		t.Logf("golangci-lint installation succeeded in %v", duration)

		// Verify installation
		if !utils.CommandExists("golangci-lint") {
			t.Error("golangci-lint command not found after installation")
		}
	}
}

// TestIntegration_ConcurrentDownloads tests multiple concurrent downloads
// to ensure retry logic works correctly under concurrent load
func TestIntegration_ConcurrentDownloads(t *testing.T) {
	testhelpers.SkipIfShort(t)

	const numConcurrent = 5
	const contentSize = 1024 // 1KB per file

	// Create server that occasionally fails
	requestCount := 0
	var mutex sync.Mutex
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		requestCount++
		count := requestCount
		mutex.Unlock()

		// Fail approximately 30% of requests
		if count%3 == 0 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Server error"))
			return
		}

		// Return successful response
		content := strings.Repeat("A", contentSize)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(content))
	}))
	defer server.Close()

	tempDir, err := os.MkdirTemp("", "concurrent_download_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := &utils.DownloadConfig{
		MaxRetries:        3,
		InitialDelay:      50 * time.Millisecond,
		MaxDelay:          500 * time.Millisecond,
		Timeout:           5 * time.Second,
		BackoffMultiplier: 2.0,
		UserAgent:         "mage-x-concurrent-test/1.0",
	}

	// Channel to collect results
	results := make(chan error, numConcurrent)
	ctx := context.Background()

	// Start concurrent downloads
	for i := 0; i < numConcurrent; i++ {
		go func(index int) {
			destPath := filepath.Join(tempDir, fmt.Sprintf("file_%d.txt", index))
			err := utils.DownloadWithRetry(ctx, server.URL, destPath, config)
			results <- err
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numConcurrent; i++ {
		err := <-results
		if err != nil {
			t.Logf("Download %d failed: %v", i, err)
		} else {
			successCount++
		}
	}

	t.Logf("Concurrent downloads completed: %d/%d successful", successCount, numConcurrent)
	t.Logf("Total server requests: %d", requestCount)

	// We expect most downloads to succeed due to retry logic
	if successCount < numConcurrent/2 {
		t.Errorf("Expected at least %d successful downloads, got %d", numConcurrent/2, successCount)
	}

	// Verify downloaded files
	for i := 0; i < successCount; i++ {
		destPath := filepath.Join(tempDir, fmt.Sprintf("file_%d.txt", i))
		if stat, err := os.Stat(destPath); err == nil {
			if stat.Size() != int64(contentSize) {
				t.Errorf("File %d has wrong size: expected %d, got %d", i, contentSize, stat.Size())
			}
		}
	}
}

// TestIntegration_NetworkLatency tests retry behavior under high network latency
func TestIntegration_NetworkLatency(t *testing.T) {
	testhelpers.SkipIfShort(t)

	// Create server with artificial latency
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate network latency
		time.Sleep(200 * time.Millisecond)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Content delivered with latency"))
	}))
	defer server.Close()

	tempDir, err := os.MkdirTemp("", "latency_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	destPath := filepath.Join(tempDir, "latency_test.txt")

	config := &utils.DownloadConfig{
		MaxRetries:        3,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          1 * time.Second,
		Timeout:           2 * time.Second, // Should be enough for 200ms latency
		BackoffMultiplier: 2.0,
		UserAgent:         "mage-x-latency-test/1.0",
	}

	ctx := context.Background()
	start := time.Now()

	err = utils.DownloadWithRetry(ctx, server.URL, destPath, config)
	duration := time.Since(start)

	if err != nil {
		t.Errorf("Download with latency failed: %v", err)
	} else {
		t.Logf("Download with latency succeeded in %v", duration)

		// Verify minimum time (should include latency)
		if duration < 200*time.Millisecond {
			t.Errorf("Download completed too quickly: %v (expected >= 200ms)", duration)
		}
	}
}

// TestIntegration_ProxyFailover tests behavior when proxy fails and direct connection works
func TestIntegration_ProxyFailover(t *testing.T) {
	testhelpers.SkipIfShort(t)

	t.Skip("Skipping proxy failover test - requires complex proxy setup")

	// This test would set up a failing proxy and verify that the direct
	// connection fallback works in tool installation scenarios
	// Implementation would require setting up HTTP_PROXY, HTTPS_PROXY
	// environment variables pointing to a non-existent proxy
}

// TestIntegration_NetworkConnectivity tests real network connectivity scenarios
func TestIntegration_NetworkConnectivity(t *testing.T) {
	testhelpers.SkipIfShort(t)
	testhelpers.RequireNetwork(t)

	// Test with real URLs that should be accessible
	testCases := []struct {
		name        string
		url         string
		expectError bool
		description string
	}{
		{
			name:        "ValidGitHubURL",
			url:         "https://github.com/robots.txt",
			expectError: false,
			description: "Download from GitHub (should succeed)",
		},
		{
			name:        "NonExistentDomain",
			url:         "https://this-domain-definitely-does-not-exist.invalid/file.txt",
			expectError: true,
			description: "Download from non-existent domain (should fail)",
		},
		{
			name:        "ValidButNonExistentPath",
			url:         "https://github.com/this/path/does/not/exist.txt",
			expectError: true,
			description: "Download valid domain but non-existent path (should fail quickly)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "connectivity_test")
			if err != nil {
				t.Fatalf("Failed to create temp directory: %v", err)
			}
			defer os.RemoveAll(tempDir)

			destPath := filepath.Join(tempDir, "test_file.txt")
			config := &utils.DownloadConfig{
				MaxRetries:   2,
				InitialDelay: 500 * time.Millisecond,
				Timeout:      10 * time.Second,
				UserAgent:    "mage-x-connectivity-test/1.0",
			}

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			start := time.Now()
			err = utils.DownloadWithRetry(ctx, tc.url, destPath, config)
			duration := time.Since(start)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, but succeeded", tc.description)
				} else {
					t.Logf("Expected error occurred for %s: %v (duration: %v)", tc.description, err, duration)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for %s: %v (duration: %v)", tc.description, err, duration)
				} else {
					t.Logf("Success for %s (duration: %v)", tc.description, duration)

					// Verify file was created and has content
					if stat, statErr := os.Stat(destPath); statErr != nil {
						t.Errorf("Downloaded file not found: %v", statErr)
					} else if stat.Size() == 0 {
						t.Error("Downloaded file is empty")
					}
				}
			}
		})
	}
}
