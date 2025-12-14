package mage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewGitHubReporter(t *testing.T) {
	t.Parallel()

	reporter := NewGitHubReporter()
	if reporter == nil {
		t.Error("NewGitHubReporter() returned nil")
	}
}

func TestGitHubReporter_Start(t *testing.T) {
	t.Parallel()

	reporter := NewGitHubReporter()
	err := reporter.Start(CIMetadata{
		Platform: "github",
		Branch:   "main",
	})
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}
}

func TestGitHubReporter_ReportFailure(t *testing.T) {
	// Cannot use t.Parallel() because we capture stdout

	reporter := NewGitHubReporter()

	// Capture stdout to verify annotation format
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = w

	failure := CITestFailure{
		Test:    "TestExample",
		Package: "github.com/example/pkg",
		File:    "example_test.go",
		Line:    42,
		Error:   "assertion failed",
		Type:    FailureTypeTest,
	}

	err = reporter.ReportFailure(failure)
	if err != nil {
		t.Errorf("ReportFailure() error = %v", err)
	}

	err = w.Close()
	if err != nil {
		t.Errorf("w.Close() error = %v", err)
	}
	os.Stdout = old

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil {
		t.Fatalf("r.Read() error = %v", err)
	}
	output := string(buf[:n])

	// Verify the annotation format
	if !strings.Contains(output, "::error") {
		t.Errorf("Expected ::error in output, got %q", output)
	}
	if !strings.Contains(output, "file=example_test.go") {
		t.Errorf("Expected file= in output, got %q", output)
	}
	if !strings.Contains(output, "line=42") {
		t.Errorf("Expected line= in output, got %q", output)
	}
	if !strings.Contains(output, "title=TestExample") {
		t.Errorf("Expected title= in output, got %q", output)
	}
	if !strings.Contains(output, "assertion failed") {
		t.Errorf("Expected error message in output, got %q", output)
	}
}

func TestGitHubReporter_ReportFailure_NoLocation(t *testing.T) {
	// Cannot use t.Parallel() because we capture stdout

	reporter := NewGitHubReporter()

	// Capture stdout
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	os.Stdout = w

	failure := CITestFailure{
		Test:  "TestExample",
		Error: "something went wrong",
		Type:  FailureTypeBuild,
	}

	err = reporter.ReportFailure(failure)
	if err != nil {
		t.Errorf("ReportFailure() error = %v", err)
	}

	err = w.Close()
	if err != nil {
		t.Errorf("w.Close() error = %v", err)
	}
	os.Stdout = old

	buf := make([]byte, 1024)
	n, err := r.Read(buf)
	if err != nil {
		t.Fatalf("r.Read() error = %v", err)
	}
	output := string(buf[:n])

	// Should still have ::error but no file/line
	if !strings.Contains(output, "::error") {
		t.Errorf("Expected ::error in output, got %q", output)
	}
	// Should not have file= since no file provided
	if strings.Contains(output, "file=") {
		t.Errorf("Expected no file= in output when file is empty, got %q", output)
	}
}

func TestGitHubReporter_WriteStepSummary(t *testing.T) {
	// Create temp file for GITHUB_STEP_SUMMARY
	tmpDir := t.TempDir()
	summaryFile := filepath.Join(tmpDir, "summary.md")
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)

	r := NewGitHubReporter()
	_, ok := r.(*githubReporter)
	if !ok {
		t.Fatal("failed to cast to *githubReporter")
	}
	// Need to recreate to pick up env var
	reporter := &githubReporter{
		stepSummaryFile: summaryFile,
	}

	result := &CIResult{
		Summary: CISummary{
			Status:   TestStatusFailed,
			Total:    10,
			Passed:   8,
			Failed:   2,
			Skipped:  0,
			Duration: "1.5s",
		},
		Failures: []CITestFailure{
			{
				Test:    "TestFoo",
				Package: "pkg/foo",
				File:    "foo_test.go",
				Line:    10,
				Type:    FailureTypeTest,
				Error:   "expected 1, got 2",
			},
		},
	}

	err := reporter.WriteStepSummary(result)
	if err != nil {
		t.Errorf("WriteStepSummary() error = %v", err)
	}

	// Read the file and verify markdown content
	content, err := os.ReadFile(summaryFile) //nolint:gosec // Test file reading is safe
	if err != nil {
		t.Fatalf("Failed to read summary file: %v", err)
	}

	summary := string(content)

	// Check for expected markdown elements
	if !strings.Contains(summary, "## Test Results") {
		t.Error("Expected '## Test Results' header")
	}
	if !strings.Contains(summary, "❌") {
		t.Error("Expected failure emoji")
	}
	if !strings.Contains(summary, "| ✅ Passed | 8 |") {
		t.Error("Expected passed count in table")
	}
	if !strings.Contains(summary, "| ❌ Failed | 2 |") {
		t.Error("Expected failed count in table")
	}
	if !strings.Contains(summary, "### Failed Tests") {
		t.Error("Expected 'Failed Tests' section")
	}
	if !strings.Contains(summary, "TestFoo") {
		t.Error("Expected test name in failures")
	}
}

func TestGitHubReporter_WriteStepSummary_NoFile(t *testing.T) {
	t.Parallel()

	// When GITHUB_STEP_SUMMARY is not set, should not error
	reporter := &githubReporter{
		stepSummaryFile: "",
	}

	result := &CIResult{
		Summary: CISummary{Status: TestStatusPassed},
	}

	err := reporter.WriteStepSummary(result)
	if err != nil {
		t.Errorf("WriteStepSummary() should not error when no summary file: %v", err)
	}
}

func TestGitHubReporter_WriteStepSummary_SkipsEmpty(t *testing.T) {
	// Create temp file for GITHUB_STEP_SUMMARY
	tmpDir := t.TempDir()
	summaryFile := filepath.Join(tmpDir, "summary.md")
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)

	reporter := &githubReporter{
		stepSummaryFile: summaryFile,
	}

	// Empty result should be skipped - no file should be created
	emptyResult := &CIResult{}
	err := reporter.WriteStepSummary(emptyResult)
	if err != nil {
		t.Errorf("WriteStepSummary() error = %v", err)
	}

	// File should not exist because empty summary was skipped
	if _, statErr := os.Stat(summaryFile); !os.IsNotExist(statErr) {
		t.Error("Expected summary file to not be created for empty result")
	}

	// Nil result should also be skipped
	err = reporter.WriteStepSummary(nil)
	if err != nil {
		t.Errorf("WriteStepSummary(nil) error = %v", err)
	}

	// Result with zeros but no status should be skipped
	zeroResult := &CIResult{
		Summary: CISummary{
			Total:   0,
			Passed:  0,
			Failed:  0,
			Skipped: 0,
			Status:  "",
		},
	}
	err = reporter.WriteStepSummary(zeroResult)
	if err != nil {
		t.Errorf("WriteStepSummary() error = %v", err)
	}

	// File should still not exist
	if _, statErr := os.Stat(summaryFile); !os.IsNotExist(statErr) {
		t.Error("Expected summary file to not be created for zero result with no status")
	}

	// Result with status but zero tests should be written (has meaningful status)
	statusOnlyResult := &CIResult{
		Summary: CISummary{
			Status: TestStatusPassed,
		},
	}
	err = reporter.WriteStepSummary(statusOnlyResult)
	if err != nil {
		t.Errorf("WriteStepSummary() error = %v", err)
	}

	// File should now exist because we had a valid status
	if _, statErr := os.Stat(summaryFile); os.IsNotExist(statErr) {
		t.Error("Expected summary file to be created for result with valid status")
	}
}

func TestGitHubReporter_WriteStepSummary_EnvSkip(t *testing.T) {
	// Create temp file for GITHUB_STEP_SUMMARY
	tmpDir := t.TempDir()
	summaryFile := filepath.Join(tmpDir, "summary.md")
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)
	t.Setenv("MAGE_X_CI_SKIP_STEP_SUMMARY", "true")

	reporter := &githubReporter{
		stepSummaryFile: summaryFile,
	}

	// Result with valid data should still be skipped due to env var
	result := &CIResult{
		Summary: CISummary{
			Status:  TestStatusPassed,
			Total:   10,
			Passed:  10,
			Failed:  0,
			Skipped: 0,
		},
	}
	err := reporter.WriteStepSummary(result)
	if err != nil {
		t.Errorf("WriteStepSummary() error = %v", err)
	}

	// File should not exist because env var skipped writing
	if _, statErr := os.Stat(summaryFile); !os.IsNotExist(statErr) {
		t.Error("Expected summary file to not be created when MAGE_X_CI_SKIP_STEP_SUMMARY is set")
	}
}

func TestGitHubReporter_WriteOutputs(t *testing.T) {
	// Create temp file for GITHUB_OUTPUT
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output")
	t.Setenv("GITHUB_OUTPUT", outputFile)

	reporter := &githubReporter{
		outputFile: outputFile,
	}

	outputs := map[string]string{
		"test_result": "failed",
		"test_count":  "10",
	}

	err := reporter.WriteOutputs(outputs)
	if err != nil {
		t.Errorf("WriteOutputs() error = %v", err)
	}

	// Read and verify
	content, err := os.ReadFile(outputFile) //nolint:gosec // Test file reading is safe
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)
	if !strings.Contains(output, "test_result=failed") {
		t.Errorf("Expected test_result output, got %q", output)
	}
	if !strings.Contains(output, "test_count=10") {
		t.Errorf("Expected test_count output, got %q", output)
	}
}

func TestGitHubReporter_WriteOutputs_Multiline(t *testing.T) {
	// Create temp file for GITHUB_OUTPUT
	tmpDir := t.TempDir()
	outputFile := filepath.Join(tmpDir, "output")
	t.Setenv("GITHUB_OUTPUT", outputFile)

	reporter := &githubReporter{
		outputFile: outputFile,
	}

	outputs := map[string]string{
		"errors": "line 1\nline 2\nline 3",
	}

	err := reporter.WriteOutputs(outputs)
	if err != nil {
		t.Errorf("WriteOutputs() error = %v", err)
	}

	// Read and verify delimiter syntax
	content, err := os.ReadFile(outputFile) //nolint:gosec // Test file reading is safe
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	output := string(content)
	// Should use delimiter syntax for multiline
	if !strings.Contains(output, "errors<<EOF_errors") {
		t.Errorf("Expected multiline delimiter syntax, got %q", output)
	}
}

func TestGitHubReporter_Close(t *testing.T) {
	t.Parallel()

	reporter := NewGitHubReporter()
	err := reporter.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestEscapeGitHubValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with%percent", "with%25percent"},
		{"with\nnewline", "with%0Anewline"},
		{"with\rcarriage", "with%0Dcarriage"},
		{"with,comma", "with%2Ccomma"},
		{"multi%\n,chars", "multi%25%0A%2Cchars"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := escapeGitHubValue(tt.input)
			if got != tt.expected {
				t.Errorf("escapeGitHubValue(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEscapeGitHubMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected string
	}{
		{"simple message", "simple message"},
		{"error:\n  details", "error:%0A  details"},
		{"100% complete", "100%25 complete"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got := escapeGitHubMessage(tt.input)
			if got != tt.expected {
				t.Errorf("escapeGitHubMessage(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestGitHubReporter_WriteSummary(t *testing.T) {
	// WriteSummary should delegate to WriteStepSummary
	tmpDir := t.TempDir()
	summaryFile := filepath.Join(tmpDir, "summary.md")
	t.Setenv("GITHUB_STEP_SUMMARY", summaryFile)

	reporter := &githubReporter{
		stepSummaryFile: summaryFile,
	}

	result := &CIResult{
		Summary: CISummary{
			Status: TestStatusPassed,
			Total:  5,
			Passed: 5,
		},
	}

	err := reporter.WriteSummary(result)
	if err != nil {
		t.Errorf("WriteSummary() error = %v", err)
	}

	// Verify file was written
	if _, err := os.Stat(summaryFile); os.IsNotExist(err) {
		t.Error("Expected summary file to be created")
	}
}
