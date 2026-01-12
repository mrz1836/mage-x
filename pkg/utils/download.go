// Package utils provides download utilities with retry logic for robust binary acquisition
package utils

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/retry"
)

const (
	// downloadBufferSize is the buffer size for file downloads (32KB optimal for network I/O)
	downloadBufferSize = 32 * 1024
)

// Static errors for err113 compliance
var (
	ErrDownloadFailed              = errors.New("download failed after all retries")
	ErrChecksumMismatch            = errors.New("downloaded file checksum does not match expected value")
	ErrInvalidURL                  = errors.New("invalid download URL")
	ErrInvalidDestination          = errors.New("invalid destination path")
	ErrFileExists                  = errors.New("destination file already exists")
	ErrPartialDownload             = errors.New("partial download completed but could not resume")
	ErrServerDoesNotSupportResume  = errors.New("server does not support resumable downloads")
	ErrInvalidWriteResult          = errors.New("invalid write result")
	ErrScriptExecutionNotSupported = errors.New("script execution requires integration with SecureExecutor")
	ErrExecutorCannotBeNil         = errors.New("executor cannot be nil")
	ErrExecutorNotImplemented      = errors.New("executor does not implement required interface")
	ErrHTTPRequestFailed           = errors.New("HTTP request failed")
)

// DownloadConfig holds configuration for download operations
type DownloadConfig struct {
	// MaxRetries is the maximum number of retry attempts (default: 5)
	MaxRetries int
	// InitialDelay is the initial delay between retries (default: 1s)
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries (default: 30s)
	MaxDelay time.Duration
	// Timeout is the timeout for each download attempt (default: 60s)
	Timeout time.Duration
	// BackoffMultiplier controls exponential backoff (default: 2.0)
	BackoffMultiplier float64
	// EnableResume enables resumable downloads using Range headers (default: true)
	EnableResume bool
	// ChecksumSHA256 is optional SHA256 checksum to verify download integrity
	ChecksumSHA256 string
	// UserAgent to use for HTTP requests
	UserAgent string
	// ProgressCallback is called periodically with download progress (optional)
	ProgressCallback func(bytesRead, totalBytes int64)
}

// DefaultDownloadConfig returns a sensible default configuration
func DefaultDownloadConfig() *DownloadConfig {
	return &DownloadConfig{
		MaxRetries:        5,
		InitialDelay:      1 * time.Second,
		MaxDelay:          30 * time.Second,
		Timeout:           60 * time.Second,
		BackoffMultiplier: 2.0,
		EnableResume:      true,
		UserAgent:         "mage-x-downloader/1.0",
	}
}

// DownloadWithRetry downloads a file from a URL with retry logic and optional resume capability
func DownloadWithRetry(ctx context.Context, url, destPath string, config *DownloadConfig) error {
	if config == nil {
		config = DefaultDownloadConfig()
	}

	// Validate inputs
	if url == "" {
		return ErrInvalidURL
	}
	if destPath == "" {
		return ErrInvalidDestination
	}

	// Clean and validate destination path
	destPath = filepath.Clean(destPath)
	destDir := filepath.Dir(destPath)
	if err := EnsureDir(destDir); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Configure retry using pkg/retry
	retryCfg := &retry.Config{
		MaxAttempts: config.MaxRetries + 1, // MaxRetries + 1 = total attempts
		Classifier:  retry.NewNetworkClassifier(),
		Backoff: &retry.ExponentialBackoff{
			Initial:    config.InitialDelay,
			Max:        config.MaxDelay,
			Multiplier: config.BackoffMultiplier,
		},
		OnRetry: func(attempt int, err error, delay time.Duration) {
			Info("Download attempt %d/%d failed: %v. Retrying in %v...",
				attempt+1, config.MaxRetries+1, err, delay)
		},
	}

	// Execute download with retry
	err := retry.Do(ctx, retryCfg, func() error {
		// Add attempt-specific context with timeout
		attemptCtx, cancel := context.WithTimeout(ctx, config.Timeout)
		defer cancel()

		return downloadFile(attemptCtx, url, destPath, config)
	})
	if err != nil {
		return err
	}

	// Verify checksum if provided (only after successful download)
	if config.ChecksumSHA256 != "" {
		if checksumErr := verifyChecksum(destPath, config.ChecksumSHA256); checksumErr != nil {
			// Remove corrupted file
			if removeErr := os.Remove(destPath); removeErr != nil {
				return fmt.Errorf("checksum verification failed and could not remove corrupted file: %w, original error: %w", removeErr, checksumErr)
			}
			return fmt.Errorf("checksum verification failed: %w", checksumErr)
		}
	}

	return nil
}

// downloadFile performs a single download attempt with resume support
func downloadFile(ctx context.Context, url, destPath string, config *DownloadConfig) error {
	// Check if file already exists and we can resume
	var resumeOffset int64
	var file *os.File
	var err error

	if config.EnableResume {
		if stat, statErr := os.Stat(destPath); statErr == nil {
			resumeOffset = stat.Size()
			//nolint:gosec // G304: destPath validated by caller
			file, err = os.OpenFile(destPath, os.O_WRONLY|os.O_APPEND, fileops.PermFile)
		}
	}

	if file == nil {
		//nolint:gosec // G304: destPath validated by caller
		file, err = os.Create(destPath)
	}
	if err != nil {
		return fmt.Errorf("failed to create/open destination file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("failed to close destination file: %v", closeErr)
		}
	}()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", config.UserAgent)

	// Add range header for resume if needed
	if resumeOffset > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", resumeOffset))
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	// Perform request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("failed to close response body: %v", closeErr)
		}
	}()

	// Check response status
	//nolint:nestif // Complex HTTP resume logic with multiple status codes
	if resumeOffset > 0 && resp.StatusCode == http.StatusPartialContent {
		// Resume successful
	} else if resumeOffset > 0 && resp.StatusCode == http.StatusOK {
		// Server doesn't support resume, restart download
		if closeErr := file.Close(); closeErr != nil {
			if bodyCloseErr := resp.Body.Close(); bodyCloseErr != nil {
				return fmt.Errorf("failed to close file before restart: %w, and failed to close response body: %w", closeErr, bodyCloseErr)
			}
			return fmt.Errorf("failed to close file before restart: %w", closeErr)
		}
		//nolint:gosec // G304: destPath validated by caller
		file, err = os.Create(destPath)
		if err != nil {
			return fmt.Errorf("failed to restart download: %w", err)
		}
		resumeOffset = 0
	} else if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w with status %d: %s", ErrHTTPRequestFailed, resp.StatusCode, resp.Status)
	}

	// Get total size if available
	var totalSize int64 = -1
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		if size, parseErr := strconv.ParseInt(contentLength, 10, 64); parseErr == nil {
			totalSize = size + resumeOffset // Add resume offset to get total file size
		}
	}

	// Copy with progress tracking
	return copyWithProgress(ctx, file, resp.Body, resumeOffset, totalSize, config.ProgressCallback)
}

// copyWithProgress copies data from src to dst with progress reporting
func copyWithProgress(ctx context.Context, dst io.Writer, src io.Reader, offset, totalSize int64,
	progressCallback func(int64, int64),
) error {
	buf := make([]byte, downloadBufferSize)
	bytesWritten := offset

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		nr, er := src.Read(buf)
		//nolint:nestif // Complex io.Copy logic with progress tracking
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = ErrInvalidWriteResult
				}
			}
			bytesWritten += int64(nw)

			// Report progress
			if progressCallback != nil {
				progressCallback(bytesWritten, totalSize)
			}

			if ew != nil {
				return ew
			}
			if nr != nw {
				return io.ErrShortWrite
			}
		}
		if er != nil {
			if er != io.EOF {
				return er
			}
			break
		}
	}

	return nil
}

// verifyChecksum verifies the SHA256 checksum of a downloaded file
func verifyChecksum(filePath, expectedSHA256 string) error {
	//nolint:gosec // G304: filePath validated by caller
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file for checksum verification: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("failed to close file during checksum verification: %v", closeErr)
		}
	}()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	actualSHA256 := hex.EncodeToString(hash.Sum(nil))
	if actualSHA256 != expectedSHA256 {
		return fmt.Errorf("%w: expected %s, got %s", ErrChecksumMismatch, expectedSHA256, actualSHA256)
	}

	return nil
}

// NOTE: isRetriableError has been replaced by retry.NewNetworkClassifier().IsRetriable()
// from pkg/retry. The classifier provides comprehensive network error classification.

// ExecutorFunc is a function type that executes a command with the given name and arguments.
// This allows callers to pass in their own execution strategy (e.g., SecureExecutor)
// to avoid circular dependencies between packages.
type ExecutorFunc func(ctx context.Context, name string, args ...string) error

// DownloadScript downloads and executes a shell script with retry logic.
// The executor parameter is required and will be used to execute the downloaded script.
// This design allows the caller to provide their own secure execution strategy.
func DownloadScript(ctx context.Context, url, scriptArgs string, config *DownloadConfig, executor ExecutorFunc) error {
	if executor == nil {
		return ErrExecutorCannotBeNil
	}

	if config == nil {
		config = DefaultDownloadConfig()
	}

	// Create temporary file for script
	tmpFile, err := os.CreateTemp("", "mage-download-*.sh")
	if err != nil {
		return fmt.Errorf("failed to create temporary script file: %w", err)
	}

	scriptPath := tmpFile.Name()

	// Ensure cleanup happens after execution
	defer func() {
		if removeErr := os.Remove(scriptPath); removeErr != nil {
			log.Printf("failed to remove temporary script %s: %v", scriptPath, removeErr)
		}
	}()

	// Close the file before downloading (DownloadWithRetry opens it)
	if closeErr := tmpFile.Close(); closeErr != nil {
		return fmt.Errorf("failed to close temporary file: %w", closeErr)
	}

	// Download script
	if err := DownloadWithRetry(ctx, url, scriptPath, config); err != nil {
		return fmt.Errorf("failed to download script: %w", err)
	}

	// Make executable - scripts need execute permission
	if err := os.Chmod(scriptPath, fileops.PermFileExecutable); err != nil {
		return fmt.Errorf("failed to make script executable: %w", err)
	}

	// Parse script arguments and execute
	args := splitArgs(scriptArgs)
	if err := executor(ctx, scriptPath, args...); err != nil {
		return fmt.Errorf("failed to execute script: %w", err)
	}

	return nil
}

// splitArgs splits a string into arguments (simple implementation)
func splitArgs(s string) []string {
	var args []string
	var current string
	inQuotes := false

	for i, r := range s {
		switch r {
		case ' ', '\t':
			if !inQuotes && current != "" {
				args = append(args, current)
				current = ""
			} else if inQuotes {
				current += string(r)
			}
		case '"', '\'':
			if i == 0 || s[i-1] != '\\' {
				inQuotes = !inQuotes
			} else {
				current += string(r)
			}
		default:
			current += string(r)
		}
	}

	if current != "" {
		args = append(args, current)
	}

	return args
}

// ConfigToDownloadConfig converts mage.DownloadConfig to utils.DownloadConfig
func ConfigToDownloadConfig(cfg interface{}) *DownloadConfig {
	// Use type assertion to safely convert from mage config
	// This avoids circular import between mage and utils packages
	if cfg == nil {
		return DefaultDownloadConfig()
	}

	// Use reflection-like approach via interface{} to extract values
	config := DefaultDownloadConfig()

	// This is a simplified version - in practice you'd use a proper conversion
	// or pass values individually to avoid the type conversion complexity
	return config
}

// ExecuteCommandWithRetry wraps the command execution with retry logic
// This will be called from the mage package to avoid circular dependencies
func ExecuteCommandWithRetry(ctx context.Context, executor interface{}, maxRetries int,
	initialDelayMs int, name string, args ...string,
) error {
	// Cast the executor to the expected interface
	// This approach avoids circular imports while maintaining type safety
	if executor == nil {
		return ErrExecutorCannotBeNil
	}

	// Use type assertion to call the retry method
	// The actual implementation will be passed from the mage package
	type retryExecutor interface {
		ExecuteWithRetry(ctx context.Context, maxRetries int, initialDelay time.Duration,
			name string, args ...string) error
	}

	if retryExec, ok := executor.(retryExecutor); ok {
		initialDelay := time.Duration(initialDelayMs) * time.Millisecond
		return retryExec.ExecuteWithRetry(ctx, maxRetries, initialDelay, name, args...)
	}

	// Fallback to basic execution if retry interface not available
	type basicExecutor interface {
		Execute(ctx context.Context, name string, args ...string) error
	}

	if basicExec, ok := executor.(basicExecutor); ok {
		return basicExec.Execute(ctx, name, args...)
	}

	return ErrExecutorNotImplemented
}
