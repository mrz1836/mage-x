package utils

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mrz1836/mage-x/pkg/retry"
)

// Static error definitions for err113 compliance
var (
	errTestConnectionRefused = errors.New("connection refused")
	errTestTimeout           = errors.New("timeout occurred")
	errTestNoSuchHost        = errors.New("no such host: example.com")
	errTestFileNotFound      = errors.New("file not found")
	errTestHTTP404           = errors.New("HTTP request failed with status 404")
	errTestHTTP500           = errors.New("HTTP request failed with status 500")
	errTestExecutionFailed   = errors.New("execution failed")
)

func TestDefaultDownloadConfig(t *testing.T) {
	config := DefaultDownloadConfig()

	if config.MaxRetries != 5 {
		t.Errorf("Expected MaxRetries to be 5, got %d", config.MaxRetries)
	}
	if config.InitialDelay != 1*time.Second {
		t.Errorf("Expected InitialDelay to be 1s, got %v", config.InitialDelay)
	}
	if config.BackoffMultiplier != 2.0 {
		t.Errorf("Expected BackoffMultiplier to be 2.0, got %f", config.BackoffMultiplier)
	}
	if !config.EnableResume {
		t.Error("Expected EnableResume to be true")
	}
}

func TestDownloadWithRetry_Success(t *testing.T) {
	// Create test content
	testContent := "This is a test file content for download testing."
	expectedSHA256 := sha256.Sum256([]byte(testContent))
	expectedChecksum := hex.EncodeToString(expectedSHA256[:])

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(testContent)); err != nil {
			t.Errorf("Failed to write test content: %v", err)
		}
	}))
	defer server.Close()

	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "download_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			t.Errorf("Failed to remove temp directory: %v", removeErr)
		}
	}()

	destPath := filepath.Join(tempDir, "test_file.txt")

	config := &DownloadConfig{
		MaxRetries:     2,
		InitialDelay:   10 * time.Millisecond,
		MaxDelay:       100 * time.Millisecond,
		Timeout:        5 * time.Second,
		ChecksumSHA256: expectedChecksum,
		UserAgent:      "test-agent",
	}

	ctx := context.Background()
	err = DownloadWithRetry(ctx, server.URL, destPath, config)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Verify file exists and content is correct
	//nolint:gosec // G304: Test file path is controlled
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != testContent {
		t.Errorf("Downloaded content mismatch. Expected: %q, Got: %q", testContent, string(content))
	}
}

func TestDownloadWithRetry_ChecksumMismatch(t *testing.T) {
	testContent := "Test content"
	wrongChecksum := "incorrect_checksum"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(testContent)); err != nil {
			t.Errorf("Failed to write test content: %v", err)
		}
	}))
	defer server.Close()

	tempDir, err := os.MkdirTemp("", "download_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			t.Errorf("Failed to remove temp directory: %v", removeErr)
		}
	}()

	destPath := filepath.Join(tempDir, "test_file.txt")

	config := &DownloadConfig{
		MaxRetries:     1,
		InitialDelay:   10 * time.Millisecond,
		Timeout:        5 * time.Second,
		ChecksumSHA256: wrongChecksum,
	}

	ctx := context.Background()
	err = DownloadWithRetry(ctx, server.URL, destPath, config)
	if err == nil {
		t.Fatal("Expected download to fail due to checksum mismatch")
	}

	if !strings.Contains(err.Error(), "checksum verification failed") {
		t.Errorf("Expected checksum error, got: %v", err)
	}
}

func TestDownloadWithRetry_ServerErrors(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		// Fail first 2 attempts, succeed on 3rd
		if attempt < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte("Server error")); err != nil {
				t.Errorf("Failed to write server error: %v", err)
			}
			return
		}
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("Success content")); err != nil {
			t.Errorf("Failed to write success content: %v", err)
		}
	}))
	defer server.Close()

	tempDir, err := os.MkdirTemp("", "download_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			t.Errorf("Failed to remove temp directory: %v", removeErr)
		}
	}()

	destPath := filepath.Join(tempDir, "test_file.txt")

	config := &DownloadConfig{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Timeout:      2 * time.Second,
	}

	ctx := context.Background()
	err = DownloadWithRetry(ctx, server.URL, destPath, config)
	if err != nil {
		t.Fatalf("Download should have succeeded on retry: %v", err)
	}

	if attempt != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempt)
	}

	// Verify content
	//nolint:gosec // G304: Test file path is controlled
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read downloaded file: %v", err)
	}

	if string(content) != "Success content" {
		t.Errorf("Unexpected content: %q", string(content))
	}
}

func TestDownloadWithRetry_PermanentFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte("Not found")); err != nil {
			t.Errorf("Failed to write not found response: %v", err)
		}
	}))
	defer server.Close()

	tempDir, err := os.MkdirTemp("", "download_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			t.Errorf("Failed to remove temp directory: %v", removeErr)
		}
	}()

	destPath := filepath.Join(tempDir, "test_file.txt")

	config := &DownloadConfig{
		MaxRetries:   2,
		InitialDelay: 10 * time.Millisecond,
		Timeout:      1 * time.Second,
	}

	ctx := context.Background()
	err = DownloadWithRetry(ctx, server.URL, destPath, config)
	if err == nil {
		t.Fatal("Expected download to fail permanently")
	}

	// Error message is now from retry package: "permanent error (not retriable)"
	if !strings.Contains(err.Error(), "permanent error") && !strings.Contains(err.Error(), "not retriable") {
		t.Errorf("Expected permanent error, got: %v", err)
	}
}

func TestDownloadWithRetry_Resume(t *testing.T) {
	fullContent := "This is the full content that should be downloaded in parts."
	partialContent := fullContent[:20] // First 20 characters

	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++

		// Check for Range header
		rangeHeader := r.Header.Get("Range")

		if rangeHeader != "" {
			// Resume request - return the rest of the content
			w.Header().Set("Content-Range", fmt.Sprintf("bytes 20-%d/%d", len(fullContent)-1, len(fullContent)))
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fullContent)-20))
			w.WriteHeader(http.StatusPartialContent)
			if _, err := w.Write([]byte(fullContent[20:])); err != nil {
				t.Errorf("Failed to write partial content: %v", err)
			}
		} else {
			// Normal request - return full content
			w.Header().Set("Content-Length", fmt.Sprintf("%d", len(fullContent)))
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(fullContent)); err != nil {
				t.Errorf("Failed to write full content: %v", err)
			}
		}
	}))
	defer server.Close()

	tempDir, err := os.MkdirTemp("", "download_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			t.Errorf("Failed to remove temp directory: %v", removeErr)
		}
	}()

	destPath := filepath.Join(tempDir, "test_file.txt")

	// First, simulate a partial download
	err = os.WriteFile(destPath, []byte(partialContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create partial file: %v", err)
	}

	config := &DownloadConfig{
		MaxRetries:   2,
		InitialDelay: 10 * time.Millisecond,
		EnableResume: true,
		Timeout:      2 * time.Second,
	}

	ctx := context.Background()
	err = DownloadWithRetry(ctx, server.URL, destPath, config)
	if err != nil {
		t.Fatalf("Resume download failed: %v", err)
	}

	// Verify full content
	//nolint:gosec // G304: Test file path is controlled
	content, err := os.ReadFile(destPath)
	if err != nil {
		t.Fatalf("Failed to read resumed file: %v", err)
	}

	if string(content) != fullContent {
		t.Errorf("Resume failed. Expected: %q, Got: %q", fullContent, string(content))
	}

	if attempt < 1 {
		t.Errorf("Expected at least 1 attempt for resume, got %d", attempt)
	}
}

func TestDownloadWithRetry_Context_Cancellation(t *testing.T) {
	// Create a server that responds slowly
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Longer than context timeout
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("Should not reach here")); err != nil {
			t.Errorf("Failed to write timeout response: %v", err)
		}
	}))
	defer server.Close()

	tempDir, err := os.MkdirTemp("", "download_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			t.Errorf("Failed to remove temp directory: %v", removeErr)
		}
	}()

	destPath := filepath.Join(tempDir, "test_file.txt")

	config := &DownloadConfig{
		MaxRetries:   1,
		InitialDelay: 10 * time.Millisecond,
		Timeout:      500 * time.Millisecond, // Short timeout
	}

	ctx := context.Background()
	err = DownloadWithRetry(ctx, server.URL, destPath, config)
	if err == nil {
		t.Fatal("Expected download to fail due to timeout")
	}

	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("Expected context deadline exceeded error, got: %v", err)
	}
}

// TestNetworkClassifier tests that the retry.NetworkClassifier correctly
// classifies errors. This validates the integration between download.go
// and pkg/retry.
func TestNetworkClassifier(t *testing.T) {
	classifier := retry.NewNetworkClassifier()

	testCases := []struct {
		name      string
		err       error
		retriable bool
	}{
		{
			name:      "nil error",
			err:       nil,
			retriable: false,
		},
		{
			name:      "connection refused",
			err:       errTestConnectionRefused,
			retriable: true,
		},
		{
			name:      "timeout error",
			err:       errTestTimeout,
			retriable: true,
		},
		{
			name:      "no such host",
			err:       errTestNoSuchHost,
			retriable: true,
		},
		{
			name:      "file not found error",
			err:       errTestFileNotFound,
			retriable: false,
		},
		{
			name:      "status 404 error",
			err:       errTestHTTP404,
			retriable: false,
		},
		{
			name:      "status 500 error",
			err:       errTestHTTP500,
			retriable: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := classifier.IsRetriable(tc.err)
			if result != tc.retriable {
				t.Errorf("Expected retriable=%v for error %q, got %v", tc.retriable, tc.err, result)
			}
		})
	}
}

func TestDownloadWithRetry_InvalidInputs(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "download_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			t.Errorf("Failed to remove temp directory: %v", removeErr)
		}
	}()

	ctx := context.Background()
	config := DefaultDownloadConfig()

	// Test empty URL
	err = DownloadWithRetry(ctx, "", filepath.Join(tempDir, "test.txt"), config)
	if err == nil || !strings.Contains(err.Error(), "invalid download URL") {
		t.Errorf("Expected invalid download URL error, got: %v", err)
	}

	// Test empty destination
	err = DownloadWithRetry(ctx, "http://example.com/test.txt", "", config)
	if err == nil || !strings.Contains(err.Error(), "invalid destination") {
		t.Errorf("Expected invalid destination error, got: %v", err)
	}
}

func TestProgressCallback(t *testing.T) {
	testContent := strings.Repeat("A", 1000) // 1KB of content
	var progressUpdates []int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		w.WriteHeader(http.StatusOK)

		// Write content in chunks to trigger progress updates
		chunks := 10
		chunkSize := len(testContent) / chunks
		for i := 0; i < chunks; i++ {
			start := i * chunkSize
			end := start + chunkSize
			if i == chunks-1 {
				end = len(testContent)
			}
			if _, err := w.Write([]byte(testContent[start:end])); err != nil {
				t.Errorf("Failed to write progress content: %v", err)
			}
			// Force flush to trigger progress callback
			if flusher, ok := w.(http.Flusher); ok {
				flusher.Flush()
			}
			time.Sleep(10 * time.Millisecond) // Small delay to allow progress tracking
		}
	}))
	defer server.Close()

	tempDir, err := os.MkdirTemp("", "download_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if removeErr := os.RemoveAll(tempDir); removeErr != nil {
			t.Logf("Failed to remove temp directory: %v", removeErr)
		}
	}()

	destPath := filepath.Join(tempDir, "test_file.txt")

	config := &DownloadConfig{
		MaxRetries:   1,
		InitialDelay: 10 * time.Millisecond,
		Timeout:      5 * time.Second,
		ProgressCallback: func(bytesRead, totalBytes int64) {
			progressUpdates = append(progressUpdates, bytesRead)
		},
	}

	ctx := context.Background()
	err = DownloadWithRetry(ctx, server.URL, destPath, config)
	if err != nil {
		t.Fatalf("Download failed: %v", err)
	}

	// Check that we received progress updates
	if len(progressUpdates) == 0 {
		t.Error("Expected progress updates, got none")
	}

	// Check that progress increases
	for i := 1; i < len(progressUpdates); i++ {
		if progressUpdates[i] < progressUpdates[i-1] {
			t.Errorf("Progress should increase, got %d after %d", progressUpdates[i], progressUpdates[i-1])
		}
	}

	// Final progress should match file size
	if len(progressUpdates) > 0 {
		finalProgress := progressUpdates[len(progressUpdates)-1]
		if finalProgress != int64(len(testContent)) {
			t.Errorf("Final progress %d should equal content length %d", finalProgress, len(testContent))
		}
	}
}

func TestDownloadHelperFunctions(t *testing.T) {
	// Test splitArgs function
	args := splitArgs("arg1 arg2 \"quoted arg\" 'single quoted'")
	expected := []string{"arg1", "arg2", "quoted arg", "single quoted"}

	if len(args) != len(expected) {
		t.Errorf("Expected %d args, got %d", len(expected), len(args))
	}

	for i, arg := range args {
		if i < len(expected) && arg != expected[i] {
			t.Errorf("Arg %d: expected %q, got %q", i, expected[i], arg)
		}
	}
}

// Benchmark tests
func BenchmarkDownloadSmallFile(b *testing.B) {
	content := "Small test content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(content)); err != nil {
			http.Error(w, "Failed to write content", http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	config := &DownloadConfig{
		MaxRetries:   1,
		InitialDelay: 1 * time.Millisecond,
		Timeout:      1 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tempDir, err := os.MkdirTemp("", "bench")
		if err != nil {
			b.Fatalf("Failed to create temp directory: %v", err)
		}
		destPath := filepath.Join(tempDir, "test.txt")

		ctx := context.Background()
		if err := DownloadWithRetry(ctx, server.URL, destPath, config); err != nil {
			b.Logf("Download failed: %v", err)
		}

		if err := os.RemoveAll(tempDir); err != nil {
			b.Logf("Failed to remove temp directory: %v", err)
		}
	}
}

func TestDownloadScript(t *testing.T) {
	// Create a test server that serves a simple shell script
	scriptContent := "#!/bin/bash\necho 'Hello World'\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/x-shellscript")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(scriptContent)); err != nil {
			t.Errorf("Failed to write script content: %v", err)
		}
	}))
	defer server.Close()

	t.Run("DownloadScript with nil executor returns error", func(t *testing.T) {
		ctx := context.Background()
		config := &DownloadConfig{
			MaxRetries:   1,
			InitialDelay: 10 * time.Millisecond,
			Timeout:      5 * time.Second,
		}

		err := DownloadScript(ctx, server.URL+"/script.sh", "", config, nil)

		if !errors.Is(err, ErrExecutorCannotBeNil) {
			t.Errorf("Expected ErrExecutorCannotBeNil, got: %v", err)
		}
	})

	t.Run("DownloadScript with executor executes script", func(t *testing.T) {
		ctx := context.Background()
		config := &DownloadConfig{
			MaxRetries:   1,
			InitialDelay: 10 * time.Millisecond,
			Timeout:      5 * time.Second,
		}

		var executedScript string
		var executedArgs []string
		executor := func(ctx context.Context, name string, args ...string) error {
			executedScript = name
			executedArgs = args
			return nil
		}

		err := DownloadScript(ctx, server.URL+"/script.sh", "-- -b /usr/local/bin v1.0.0", config, executor)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		// Script path should be a temp file
		if executedScript == "" {
			t.Error("Expected script to be executed")
		}
		if !strings.HasPrefix(executedScript, os.TempDir()) && !strings.Contains(executedScript, "mage-download") {
			t.Errorf("Expected temp file path, got: %s", executedScript)
		}
		// Args should be parsed correctly
		expectedArgs := []string{"--", "-b", "/usr/local/bin", "v1.0.0"}
		if len(executedArgs) != len(expectedArgs) {
			t.Errorf("Expected %d args, got %d: %v", len(expectedArgs), len(executedArgs), executedArgs)
		}
		for i, arg := range executedArgs {
			if i < len(expectedArgs) && arg != expectedArgs[i] {
				t.Errorf("Arg %d: expected %q, got %q", i, expectedArgs[i], arg)
			}
		}
	})

	t.Run("DownloadScript with nil config uses defaults", func(t *testing.T) {
		ctx := context.Background()

		executed := false
		executor := func(ctx context.Context, name string, args ...string) error {
			executed = true
			return nil
		}

		err := DownloadScript(ctx, server.URL+"/script.sh", "", nil, executor)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if !executed {
			t.Error("Expected script to be executed")
		}
	})

	t.Run("DownloadScript with invalid URL fails", func(t *testing.T) {
		ctx := context.Background()
		config := &DownloadConfig{
			MaxRetries:   1,
			InitialDelay: 10 * time.Millisecond,
			Timeout:      1 * time.Second,
		}

		executor := func(ctx context.Context, name string, args ...string) error {
			return nil
		}

		err := DownloadScript(ctx, "http://invalid.example.test:12345/script.sh", "", config, executor)

		// Should fail during download
		if err == nil {
			t.Error("Expected error for invalid URL")
		}
		if !strings.Contains(err.Error(), "failed to download script") {
			t.Errorf("Expected download error, got: %v", err)
		}
	})

	t.Run("DownloadScript executor error propagates", func(t *testing.T) {
		ctx := context.Background()
		config := &DownloadConfig{
			MaxRetries:   1,
			InitialDelay: 10 * time.Millisecond,
			Timeout:      5 * time.Second,
		}

		executor := func(ctx context.Context, name string, args ...string) error {
			return errTestExecutionFailed
		}

		err := DownloadScript(ctx, server.URL+"/script.sh", "", config, executor)

		if err == nil {
			t.Error("Expected error from executor")
		}
		if !strings.Contains(err.Error(), "failed to execute script") {
			t.Errorf("Expected execute error, got: %v", err)
		}
	})

	t.Run("DownloadScript cleans up temp file", func(t *testing.T) {
		ctx := context.Background()
		config := &DownloadConfig{
			MaxRetries:   1,
			InitialDelay: 10 * time.Millisecond,
			Timeout:      5 * time.Second,
		}

		var capturedScript string
		executor := func(ctx context.Context, name string, args ...string) error {
			capturedScript = name
			// Verify file exists during execution
			if _, err := os.Stat(name); err != nil {
				t.Errorf("Script file should exist during execution: %v", err)
			}
			return nil
		}

		err := DownloadScript(ctx, server.URL+"/script.sh", "", config, executor)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		// After execution, file should be cleaned up
		if _, err := os.Stat(capturedScript); !os.IsNotExist(err) {
			t.Errorf("Script file should be deleted after execution, got: %v", err)
		}
	})
}

func TestConfigToDownloadConfig(t *testing.T) {
	t.Run("nil config returns default", func(t *testing.T) {
		result := ConfigToDownloadConfig(nil)

		if result == nil {
			t.Fatal("Expected non-nil config")
		}
		// Verify default values
		if result.MaxRetries != 5 {
			t.Errorf("Expected MaxRetries=5, got %d", result.MaxRetries)
		}
		if result.InitialDelay != 1*time.Second {
			t.Errorf("Expected InitialDelay=1s, got %v", result.InitialDelay)
		}
	})

	t.Run("non-nil config returns default", func(t *testing.T) {
		// This is a simplified implementation that always returns default
		input := &struct{ SomeField string }{SomeField: "test"}
		result := ConfigToDownloadConfig(input)

		if result == nil {
			t.Fatal("Expected non-nil config")
		}
		// Currently returns default regardless of input
		if result.MaxRetries != 5 {
			t.Errorf("Expected MaxRetries=5, got %d", result.MaxRetries)
		}
	})
}

func TestExecuteCommandWithRetry(t *testing.T) {
	t.Run("nil executor returns error", func(t *testing.T) {
		ctx := context.Background()

		err := ExecuteCommandWithRetry(ctx, nil, 3, 100, "echo", "hello")

		if !errors.Is(err, ErrExecutorCannotBeNil) {
			t.Errorf("Expected ErrExecutorCannotBeNil, got: %v", err)
		}
	})

	t.Run("executor without interface returns error", func(t *testing.T) {
		ctx := context.Background()

		// Pass an object that doesn't implement the retryExecutor interface
		invalidExecutor := struct{ name string }{name: "invalid"}

		err := ExecuteCommandWithRetry(ctx, invalidExecutor, 3, 100, "echo", "hello")

		if !errors.Is(err, ErrExecutorNotImplemented) {
			t.Errorf("Expected ErrExecutorNotImplemented, got: %v", err)
		}
	})

	t.Run("executor with interface succeeds", func(t *testing.T) {
		ctx := context.Background()

		// Create a mock executor that implements the interface
		mockExecutor := &mockRetryExecutor{
			executeFunc: func(ctx context.Context, maxRetries int, initialDelay time.Duration,
				name string, args ...string,
			) error {
				// Verify parameters were passed correctly
				if maxRetries != 3 {
					t.Errorf("Expected maxRetries=3, got %d", maxRetries)
				}
				if initialDelay != 100*time.Millisecond {
					t.Errorf("Expected initialDelay=100ms, got %v", initialDelay)
				}
				if name != "echo" {
					t.Errorf("Expected name=echo, got %s", name)
				}
				if len(args) != 1 || args[0] != "hello" {
					t.Errorf("Expected args=[hello], got %v", args)
				}
				return nil
			},
		}

		err := ExecuteCommandWithRetry(ctx, mockExecutor, 3, 100, "echo", "hello")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})

	t.Run("executor with interface returns error", func(t *testing.T) {
		ctx := context.Background()

		mockExecutor := &mockRetryExecutor{
			executeFunc: func(ctx context.Context, maxRetries int, initialDelay time.Duration,
				name string, args ...string,
			) error {
				return errTestExecutionFailed
			},
		}

		err := ExecuteCommandWithRetry(ctx, mockExecutor, 3, 100, "echo", "hello")

		if !errors.Is(err, errTestExecutionFailed) {
			t.Errorf("Expected %v, got: %v", errTestExecutionFailed, err)
		}
	})
}

// mockRetryExecutor implements the retryExecutor interface for testing
type mockRetryExecutor struct {
	executeFunc func(ctx context.Context, maxRetries int, initialDelay time.Duration,
		name string, args ...string) error
}

func (m *mockRetryExecutor) ExecuteWithRetry(ctx context.Context, maxRetries int, initialDelay time.Duration,
	name string, args ...string,
) error {
	return m.executeFunc(ctx, maxRetries, initialDelay, name, args...)
}

func BenchmarkDownloadWithRetries(b *testing.B) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt%3 == 0 { // Succeed every 3rd attempt
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("Success")); err != nil {
				http.Error(w, "Failed to write content", http.StatusInternalServerError)
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	config := &DownloadConfig{
		MaxRetries:   3,
		InitialDelay: 1 * time.Millisecond,
		Timeout:      1 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tempDir, err := os.MkdirTemp("", "bench")
		if err != nil {
			b.Fatalf("Failed to create temp directory: %v", err)
		}
		destPath := filepath.Join(tempDir, "test.txt")

		ctx := context.Background()
		if err := DownloadWithRetry(ctx, server.URL, destPath, config); err != nil {
			b.Logf("Download failed: %v", err)
		}

		if err := os.RemoveAll(tempDir); err != nil {
			b.Logf("Failed to remove temp directory: %v", err)
		}
	}
}
