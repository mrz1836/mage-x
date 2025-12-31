// Package mage provides JSON Lines reporter for CI test output
package mage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/mrz1836/mage-x/pkg/common/fileops"
)

var (
	// ErrReporterClosed is returned when operations are attempted on a closed reporter
	ErrReporterClosed = errors.New("reporter is closed")
	// ErrOutputFileNil is returned when the output file is nil
	ErrOutputFileNil = errors.New("output file is nil")
)

// jsonLineType represents the type of JSON line in output
type jsonLineType string

const (
	jsonLineTypeStart   jsonLineType = "start"
	jsonLineTypeFailure jsonLineType = "failure"
	jsonLineTypeSummary jsonLineType = "summary"
)

// jsonLine represents a single line in the JSONL output
type jsonLine struct {
	Type      jsonLineType   `json:"type"`
	Timestamp string         `json:"timestamp,omitempty"`
	Metadata  *CIMetadata    `json:"metadata,omitempty"`
	Failure   *CITestFailure `json:"failure,omitempty"`
	Summary   *CISummary     `json:"summary,omitempty"`
}

// jsonReporter implements JSONReporterInterface for JSON Lines output
type jsonReporter struct {
	path        string
	file        *os.File
	mu          sync.Mutex
	closed      bool
	diskError   bool   // Track if disk full or other I/O error occurred
	fallbackDir string // Alternative directory if primary path fails
}

// NewJSONReporter creates a JSON Lines reporter
func NewJSONReporter(path string) (JSONReporterInterface, error) {
	return NewJSONReporterWithFallback(path, "")
}

// NewJSONReporterWithFallback creates a JSON Lines reporter with a fallback directory
func NewJSONReporterWithFallback(path, fallbackDir string) (JSONReporterInterface, error) {
	reporter := &jsonReporter{
		path:        path,
		fallbackDir: fallbackDir,
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		err := os.MkdirAll(dir, fileops.PermDirSensitive)
		if err == nil {
			// Directory created successfully
		} else if fallbackDir != "" {
			// Try fallback directory if provided
			fallbackPath := filepath.Join(fallbackDir, filepath.Base(path))
			if mkdirErr := os.MkdirAll(fallbackDir, fileops.PermDirSensitive); mkdirErr != nil {
				return nil, fmt.Errorf("failed to create fallback directory: %w", err)
			}
			reporter.path = fallbackPath
			fmt.Fprintf(os.Stderr, "Warning: Using fallback path %s\n", fallbackPath)
		} else {
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	// Create or truncate file
	f, err := os.Create(reporter.path)
	if err != nil && fallbackDir != "" && reporter.path == path {
		// Try fallback if primary fails
		fallbackPath := filepath.Join(fallbackDir, filepath.Base(path))
		mkdirErr := os.MkdirAll(fallbackDir, fileops.PermDirSensitive)
		if mkdirErr == nil {
			f, err = os.Create(fallbackPath) //nolint:gosec // Fallback path is controlled
			if err == nil {
				reporter.path = fallbackPath
				fmt.Fprintf(os.Stderr, "Warning: Using fallback path %s\n", fallbackPath)
			}
		}
	}
	if f == nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}

	reporter.file = f
	return reporter, nil
}

// Start begins the test run report
func (r *jsonReporter) Start(metadata CIMetadata) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return ErrReporterClosed
	}

	line := jsonLine{
		Type:      jsonLineTypeStart,
		Timestamp: time.Now().Format(time.RFC3339),
		Metadata:  &metadata,
	}

	return r.writeLine(line)
}

// ReportFailure reports a single test failure
func (r *jsonReporter) ReportFailure(failure CITestFailure) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return ErrReporterClosed
	}

	line := jsonLine{
		Type:    jsonLineTypeFailure,
		Failure: &failure,
	}

	return r.writeLine(line)
}

// WriteSummary writes the final summary
func (r *jsonReporter) WriteSummary(result *CIResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return ErrReporterClosed
	}

	line := jsonLine{
		Type:    jsonLineTypeSummary,
		Summary: &result.Summary,
	}

	return r.writeLine(line)
}

// Flush forces any buffered output to be written
func (r *jsonReporter) Flush() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.file != nil {
		if err := r.file.Sync(); err != nil {
			return fmt.Errorf("failed to sync file: %w", err)
		}
	}
	return nil
}

// Close finalizes and closes the reporter
func (r *jsonReporter) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true

	if r.file != nil {
		if err := r.file.Sync(); err != nil {
			// Sync failed - close file but return sync error as it's more important
			closeErr := r.file.Close()
			if closeErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to close file after sync error: %v\n", closeErr)
			}
			return fmt.Errorf("failed to sync file: %w", err)
		}
		if err := r.file.Close(); err != nil {
			return fmt.Errorf("failed to close file: %w", err)
		}
	}

	return nil
}

// writeLine writes a single JSON line to the output.
// IMPORTANT: Caller MUST hold r.mu before calling this method.
// All public methods (Start, ReportFailure, WriteSummary) acquire the lock
// before calling writeLine.
func (r *jsonReporter) writeLine(line jsonLine) error {
	// Skip writing if disk error already occurred
	// (r.diskError is safe to read here since caller holds r.mu)
	if r.diskError {
		return nil // Silently skip - already warned user
	}

	if r.file == nil {
		return ErrOutputFileNil
	}

	data, err := json.Marshal(line)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write line with newline
	_, err = r.file.Write(append(data, '\n'))
	if err != nil {
		// Check for disk full or other I/O errors
		if isDiskFullError(err) {
			r.diskError = true
			fmt.Fprintf(os.Stderr, "Warning: Disk full, skipping remaining JSON Lines output\n")
			return nil // Don't fail the test run due to disk full
		}
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

// isDiskFullError checks if an error indicates disk is full
func isDiskFullError(err error) bool {
	if err == nil {
		return false
	}

	// Check for ENOSPC (no space left on device)
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errno == syscall.ENOSPC
	}

	// Check for common disk full error messages
	errStr := err.Error()
	return errStr == "no space left on device" ||
		errStr == "disk quota exceeded"
}

// GetOutputPath returns the path to the output file
func (r *jsonReporter) GetOutputPath() string {
	return r.path
}
