package mage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewJSONReporter(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "ci-results.jsonl")

	reporter, err := NewJSONReporter(outputPath)
	if err != nil {
		t.Fatalf("NewJSONReporter() error = %v", err)
	}
	t.Cleanup(func() {
		if closeErr := reporter.Close(); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	})

	if reporter == nil {
		t.Error("NewJSONReporter() returned nil")
	}
}

func TestNewJSONReporter_CreatesDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "subdir", "deep", "ci-results.jsonl")

	reporter, err := NewJSONReporter(outputPath)
	if err != nil {
		t.Fatalf("NewJSONReporter() error = %v", err)
	}
	t.Cleanup(func() {
		if closeErr := reporter.Close(); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	})

	// Verify directory was created
	if _, err := os.Stat(filepath.Dir(outputPath)); os.IsNotExist(err) {
		t.Error("Expected directory to be created")
	}
}

func TestJSONReporter_Start(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "ci-results.jsonl")

	reporter, err := NewJSONReporter(outputPath)
	if err != nil {
		t.Fatalf("NewJSONReporter() error = %v", err)
	}
	t.Cleanup(func() {
		if closeErr := reporter.Close(); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	})

	metadata := CIMetadata{
		Platform: "github",
		Branch:   "main",
		Commit:   "abc123",
		RunID:    "12345",
		Workflow: "test",
	}

	err = reporter.Start(metadata)
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}

	// Read and verify the start line
	content, err := os.ReadFile(outputPath) //nolint:gosec // Test file reading is safe
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	var line jsonLine
	if err := json.Unmarshal([]byte(lines[0]), &line); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if line.Type != jsonLineTypeStart {
		t.Errorf("Expected type 'start', got %q", line.Type)
	}
	if line.Metadata == nil {
		t.Error("Expected metadata to be present")
	}
	if line.Metadata.Platform != string(CIFormatGitHub) {
		t.Errorf("Expected platform 'github', got %q", line.Metadata.Platform)
	}
}

func TestJSONReporter_ReportFailure(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "ci-results.jsonl")

	reporter, err := NewJSONReporter(outputPath)
	if err != nil {
		t.Fatalf("NewJSONReporter() error = %v", err)
	}
	t.Cleanup(func() {
		if closeErr := reporter.Close(); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	})

	failure := CITestFailure{
		Test:    "TestExample",
		Package: "pkg/example",
		File:    "example_test.go",
		Line:    42,
		Column:  5,
		Type:    FailureTypeTest,
		Error:   "assertion failed",
		Output:  "expected 1, got 2",
		Context: []string{"line 41", "line 42", "line 43"},
	}

	err = reporter.ReportFailure(failure)
	if err != nil {
		t.Errorf("ReportFailure() error = %v", err)
	}

	// Read and verify
	content, err := os.ReadFile(outputPath) //nolint:gosec // Test file reading is safe
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	var line jsonLine
	if err := json.Unmarshal([]byte(lines[0]), &line); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if line.Type != jsonLineTypeFailure {
		t.Errorf("Expected type 'failure', got %q", line.Type)
	}
	if line.Failure == nil {
		t.Fatal("Expected failure to be present")
	}
	if line.Failure.Test != "TestExample" {
		t.Errorf("Expected test 'TestExample', got %q", line.Failure.Test)
	}
	if line.Failure.Line != 42 {
		t.Errorf("Expected line 42, got %d", line.Failure.Line)
	}
}

func TestJSONReporter_WriteSummary(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "ci-results.jsonl")

	reporter, err := NewJSONReporter(outputPath)
	if err != nil {
		t.Fatalf("NewJSONReporter() error = %v", err)
	}
	t.Cleanup(func() {
		if closeErr := reporter.Close(); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	})

	result := &CIResult{
		Summary: CISummary{
			Status:   TestStatusFailed,
			Total:    10,
			Passed:   8,
			Failed:   2,
			Skipped:  0,
			Duration: "1.5s",
		},
		Timestamp: time.Now(),
		Duration:  1500 * time.Millisecond,
	}

	err = reporter.WriteSummary(result)
	if err != nil {
		t.Errorf("WriteSummary() error = %v", err)
	}

	// Read and verify
	content, err := os.ReadFile(outputPath) //nolint:gosec // Test file reading is safe
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 1 {
		t.Fatalf("Expected 1 line, got %d", len(lines))
	}

	var line jsonLine
	if err := json.Unmarshal([]byte(lines[0]), &line); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if line.Type != jsonLineTypeSummary {
		t.Errorf("Expected type 'summary', got %q", line.Type)
	}
	if line.Summary == nil {
		t.Fatal("Expected summary to be present")
	}
	if line.Summary.Total != 10 {
		t.Errorf("Expected total 10, got %d", line.Summary.Total)
	}
	if line.Summary.Failed != 2 {
		t.Errorf("Expected failed 2, got %d", line.Summary.Failed)
	}
}

func TestJSONReporter_FullWorkflow(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "ci-results.jsonl")

	reporter, err := NewJSONReporter(outputPath)
	if err != nil {
		t.Fatalf("NewJSONReporter() error = %v", err)
	}

	// Start
	err = reporter.Start(CIMetadata{Platform: "github"})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Report failures
	err = reporter.ReportFailure(CITestFailure{Test: "Test1", Error: "error1"})
	if err != nil {
		t.Fatalf("ReportFailure() error = %v", err)
	}
	err = reporter.ReportFailure(CITestFailure{Test: "Test2", Error: "error2"})
	if err != nil {
		t.Fatalf("ReportFailure() error = %v", err)
	}

	// Summary
	err = reporter.WriteSummary(&CIResult{
		Summary: CISummary{Status: TestStatusFailed, Total: 10, Failed: 2},
	})
	if err != nil {
		t.Fatalf("WriteSummary() error = %v", err)
	}

	// Close
	err = reporter.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Read and verify all lines
	content, err := os.ReadFile(outputPath) //nolint:gosec // Test file reading is safe
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 4 {
		t.Fatalf("Expected 4 lines, got %d: %v", len(lines), lines)
	}

	// Verify line types
	expectedTypes := []jsonLineType{jsonLineTypeStart, jsonLineTypeFailure, jsonLineTypeFailure, jsonLineTypeSummary}
	for i, expectedType := range expectedTypes {
		var line jsonLine
		if err := json.Unmarshal([]byte(lines[i]), &line); err != nil {
			t.Fatalf("Failed to parse line %d: %v", i, err)
		}
		if line.Type != expectedType {
			t.Errorf("Line %d: expected type %q, got %q", i, expectedType, line.Type)
		}
	}
}

func TestJSONReporter_Flush(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "ci-results.jsonl")

	reporter, err := NewJSONReporter(outputPath)
	if err != nil {
		t.Fatalf("NewJSONReporter() error = %v", err)
	}
	t.Cleanup(func() {
		if closeErr := reporter.Close(); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	})

	// Write some data
	err = reporter.Start(CIMetadata{Platform: "test"})
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Flush should not error
	err = reporter.Flush()
	if err != nil {
		t.Errorf("Flush() error = %v", err)
	}
}

func TestJSONReporter_CloseIdempotent(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "ci-results.jsonl")

	reporter, err := NewJSONReporter(outputPath)
	if err != nil {
		t.Fatalf("NewJSONReporter() error = %v", err)
	}

	// Close multiple times should not error
	err = reporter.Close()
	if err != nil {
		t.Errorf("First Close() error = %v", err)
	}

	err = reporter.Close()
	if err != nil {
		t.Errorf("Second Close() error = %v", err)
	}
}

func TestJSONReporter_GetOutputPath(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "ci-results.jsonl")

	reporter, err := NewJSONReporter(outputPath)
	if err != nil {
		t.Fatalf("NewJSONReporter() error = %v", err)
	}
	t.Cleanup(func() {
		if closeErr := reporter.Close(); closeErr != nil {
			t.Errorf("Close() error = %v", closeErr)
		}
	})

	jr, ok := reporter.(*jsonReporter)
	if !ok {
		t.Fatal("failed to cast to *jsonReporter")
	}
	if got := jr.GetOutputPath(); got != outputPath {
		t.Errorf("GetOutputPath() = %q, want %q", got, outputPath)
	}
}

func TestJSONReporter_WriteAfterClose(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	outputPath := filepath.Join(tmpDir, "ci-results.jsonl")

	reporter, err := NewJSONReporter(outputPath)
	if err != nil {
		t.Fatalf("NewJSONReporter() error = %v", err)
	}

	err = reporter.Close()
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Writing after close should error
	err = reporter.Start(CIMetadata{})
	if err == nil {
		t.Error("Expected error when writing after close")
	}

	err = reporter.ReportFailure(CITestFailure{})
	if err == nil {
		t.Error("Expected error when writing after close")
	}

	err = reporter.WriteSummary(&CIResult{})
	if err == nil {
		t.Error("Expected error when writing after close")
	}
}
