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
)

// Static error definitions for err113 compliance
var (
	errTestConnectionRefused = errors.New("connection refused")
	errTestTimeout           = errors.New("timeout occurred")
	errTestNoSuchHost        = errors.New("no such host: example.com")
	errTestFileNotFound      = errors.New("file not found")
	errTestHTTP404           = errors.New("HTTP request failed with status 404")
	errTestHTTP500           = errors.New("HTTP request failed with status 500")
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

	if !strings.Contains(err.Error(), "permanent download error") {
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

func TestIsRetriableError(t *testing.T) {
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
			result := isRetriableError(tc.err)
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
	// Test contains function
	if !contains("hello world", "world") {
		t.Error("contains should find substring")
	}

	if contains("hello", "xyz") {
		t.Error("contains should not find non-existent substring")
	}

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
