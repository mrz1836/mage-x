package mage

import (
	"bytes"
	"strings"
	"testing"
)

func TestNewTerminalReporter(t *testing.T) {
	t.Parallel()

	reporter := NewTerminalReporter()
	if reporter == nil {
		t.Error("NewTerminalReporter() returned nil")
	}
}

func TestTerminalReporter_Start(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	reporter := NewTerminalReporterWithWriter(&buf)

	metadata := CIMetadata{
		Platform: "local",
		Branch:   "feature-branch",
	}

	err := reporter.Start(metadata)
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "CI Mode Preview") {
		t.Errorf("Expected 'CI Mode Preview' in output, got %q", output)
	}
	if !strings.Contains(output, "Platform: local") {
		t.Errorf("Expected 'Platform: local' in output, got %q", output)
	}
	if !strings.Contains(output, "Branch: feature-branch") {
		t.Errorf("Expected branch in output, got %q", output)
	}
}

func TestTerminalReporter_ReportFailure(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	r := NewTerminalReporterWithWriter(&buf)
	reporter, ok := r.(*terminalReporter)
	if !ok {
		t.Fatal("failed to cast to *terminalReporter")
	}

	failure := CITestFailure{
		Test:    "TestExample",
		Package: "pkg/foo",
		File:    "foo_test.go",
		Line:    42,
		Type:    FailureTypeTest,
		Error:   "expected 1, got 2",
	}

	err := reporter.ReportFailure(failure)
	if err != nil {
		t.Errorf("ReportFailure() error = %v", err)
	}

	if len(reporter.failures) != 1 {
		t.Errorf("Expected 1 failure stored, got %d", len(reporter.failures))
	}
}

func TestTerminalReporter_WriteSummary_Passed(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	reporter := NewTerminalReporterWithWriter(&buf)

	result := &CIResult{
		Summary: CISummary{
			Status:   TestStatusPassed,
			Total:    10,
			Passed:   10,
			Failed:   0,
			Skipped:  0,
			Duration: "1.5s",
		},
		Failures: nil,
	}

	err := reporter.WriteSummary(result)
	if err != nil {
		t.Errorf("WriteSummary() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Test Results Summary") {
		t.Errorf("Expected summary header in output, got %q", output)
	}
	if !strings.Contains(output, "PASSED") {
		t.Errorf("Expected PASSED status in output, got %q", output)
	}
	if !strings.Contains(output, "Total:   10") {
		t.Errorf("Expected total count in output, got %q", output)
	}
	if !strings.Contains(output, "Duration: 1.5s") {
		t.Errorf("Expected duration in output, got %q", output)
	}
}

func TestTerminalReporter_WriteSummary_Failed(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	reporter := NewTerminalReporterWithWriter(&buf)

	// First report a failure
	if err := reporter.ReportFailure(CITestFailure{
		Test:    "TestFoo",
		Package: "pkg/foo",
		File:    "foo_test.go",
		Line:    42,
		Type:    FailureTypeTest,
		Error:   "assertion failed: expected true",
	}); err != nil {
		t.Fatalf("ReportFailure() error = %v", err)
	}

	result := &CIResult{
		Summary: CISummary{
			Status:   TestStatusFailed,
			Total:    10,
			Passed:   9,
			Failed:   1,
			Skipped:  0,
			Duration: "2.5s",
		},
		Failures: []CITestFailure{
			{
				Test:    "TestFoo",
				Package: "pkg/foo",
				File:    "foo_test.go",
				Line:    42,
				Type:    FailureTypeTest,
				Error:   "assertion failed: expected true",
			},
		},
	}

	err := reporter.WriteSummary(result)
	if err != nil {
		t.Errorf("WriteSummary() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "FAILED") {
		t.Errorf("Expected FAILED status in output, got %q", output)
	}
	if !strings.Contains(output, "Failed Tests") {
		t.Errorf("Expected 'Failed Tests' section in output, got %q", output)
	}
	if !strings.Contains(output, "TestFoo") {
		t.Errorf("Expected test name in output, got %q", output)
	}
	if !strings.Contains(output, "foo_test.go:42") {
		t.Errorf("Expected location in output, got %q", output)
	}
}

func TestTerminalReporter_WriteSummary_WithContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	reporter := NewTerminalReporterWithWriter(&buf)

	result := &CIResult{
		Summary: CISummary{
			Status: TestStatusFailed,
			Total:  1,
			Failed: 1,
		},
		Failures: []CITestFailure{
			{
				Test:    "TestWithContext",
				Package: "pkg",
				File:    "test.go",
				Line:    10,
				Type:    FailureTypeTest,
				Error:   "error",
				Context: []string{
					"     9 | func Test() {",
					">   10 | 	assert(false)",
					"    11 | }",
				},
			},
		},
	}

	err := reporter.WriteSummary(result)
	if err != nil {
		t.Errorf("WriteSummary() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Context:") {
		t.Errorf("Expected context section in output, got %q", output)
	}
	if !strings.Contains(output, "assert(false)") {
		t.Errorf("Expected context line in output, got %q", output)
	}
}

func TestTerminalReporter_WriteSummary_WithSkipped(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	reporter := NewTerminalReporterWithWriter(&buf)

	result := &CIResult{
		Summary: CISummary{
			Status:  TestStatusPassed,
			Total:   10,
			Passed:  8,
			Failed:  0,
			Skipped: 2,
		},
	}

	err := reporter.WriteSummary(result)
	if err != nil {
		t.Errorf("WriteSummary() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Skipped:") {
		t.Errorf("Expected Skipped in output when skipped > 0, got %q", output)
	}
}

func TestTerminalReporter_WriteSummary_PanicType(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	reporter := NewTerminalReporterWithWriter(&buf)

	result := &CIResult{
		Summary: CISummary{
			Status: TestStatusFailed,
			Total:  1,
			Failed: 1,
		},
		Failures: []CITestFailure{
			{
				Test:  "TestPanic",
				Type:  FailureTypePanic,
				Error: "panic: nil pointer",
			},
		},
	}

	err := reporter.WriteSummary(result)
	if err != nil {
		t.Errorf("WriteSummary() error = %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Type: panic") {
		t.Errorf("Expected 'Type: panic' in output for panic failure, got %q", output)
	}
}

func TestTerminalReporter_Close(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	reporter := NewTerminalReporterWithWriter(&buf)

	err := reporter.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestColorHelpers(t *testing.T) {
	t.Parallel()

	// These just verify the functions don't panic
	_ = colorBold("test")
	_ = colorRed("test")
	_ = colorGreen("test")
	_ = colorYellow("test")
	_ = colorCyan("test")
}

// T069: Integration test for local CI mode preview
func TestLocalCIModePreview(t *testing.T) {
	t.Parallel()

	t.Run("ci param enables local preview mode", func(t *testing.T) {
		t.Parallel()
		base := &mockCommandRunner{}
		params := map[string]string{"ci": "true"}

		// Mock detector that returns "local" platform
		runner := GetCIRunner(base, params, nil)

		// When CI mode is forced via param and not on actual CI platform,
		// the runner should be a CI runner
		if _, ok := runner.(*ciRunner); !ok {
			t.Error("Expected CI runner when ci=true parameter is set")
		}
	})

	t.Run("ci=false disables CI mode even in CI env", func(t *testing.T) {
		t.Parallel()
		base := &mockCommandRunner{}
		params := map[string]string{"ci": "false"}

		runner := GetCIRunner(base, params, nil)

		// Should return base runner, not CI runner
		if _, ok := runner.(*ciRunner); ok {
			t.Error("Expected base runner when ci=false")
		}
	})

	t.Run("terminal reporter used for local preview", func(t *testing.T) {
		t.Parallel()
		// Verify terminal reporter can be created and produces output
		var buf bytes.Buffer
		reporter := NewTerminalReporterWithWriter(&buf)

		if err := reporter.Start(CIMetadata{Platform: "local"}); err != nil {
			t.Fatalf("Start() error = %v", err)
		}
		if err := reporter.ReportFailure(CITestFailure{
			Test:  "TestLocal",
			Error: "local error",
		}); err != nil {
			t.Fatalf("ReportFailure() error = %v", err)
		}
		if err := reporter.WriteSummary(&CIResult{
			Summary: CISummary{
				Status: TestStatusFailed,
				Total:  1,
				Failed: 1,
			},
			Failures: []CITestFailure{
				{Test: "TestLocal", Error: "local error"},
			},
		}); err != nil {
			t.Fatalf("WriteSummary() error = %v", err)
		}
		if err := reporter.Close(); err != nil {
			t.Fatalf("Close() error = %v", err)
		}

		output := buf.String()
		if output == "" {
			t.Error("Expected non-empty output from terminal reporter")
		}
		if !strings.Contains(output, "TestLocal") {
			t.Errorf("Expected test name in output, got %q", output)
		}
	})
}
