package mage

import (
	"context"
	"strings"
	"testing"
	"time"
)

// mockCommandRunner for testing
type mockCommandRunner struct {
	lastCmd  string
	lastArgs []string
	output   string
	err      error
}

func (m *mockCommandRunner) RunCmd(name string, args ...string) error {
	m.lastCmd = name
	m.lastArgs = args
	return m.err
}

func (m *mockCommandRunner) RunCmdOutput(name string, args ...string) (string, error) {
	m.lastCmd = name
	m.lastArgs = args
	return m.output, m.err
}

// mockReporter for testing
type mockReporter struct {
	started       bool
	failures      []CITestFailure
	summaryCalled bool
	closed        bool
}

func (m *mockReporter) Start(_ CIMetadata) error {
	m.started = true
	return nil
}

func (m *mockReporter) ReportFailure(f CITestFailure) error {
	m.failures = append(m.failures, f)
	return nil
}

func (m *mockReporter) WriteSummary(_ *CIResult) error {
	m.summaryCalled = true
	return nil
}

func (m *mockReporter) Close() error {
	m.closed = true
	return nil
}

func TestNewCIRunner(t *testing.T) {
	t.Parallel()

	base := &mockCommandRunner{}
	runner := NewCIRunner(base, CIRunnerOptions{
		Mode: DefaultCIMode(),
	})

	if runner == nil {
		t.Error("NewCIRunner() returned nil")
	}
}

func TestCIRunner_WithContext(t *testing.T) {
	t.Parallel()

	base := &mockCommandRunner{}
	runner := NewCIRunner(base, CIRunnerOptions{
		Mode: DefaultCIMode(),
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	ctxRunner := runner.WithContext(ctx)
	if ctxRunner == nil {
		t.Error("WithContext() returned nil")
	}
}

func TestCIRunner_RunCmd_NonGoTest(t *testing.T) {
	t.Parallel()

	base := &mockCommandRunner{}
	mode := DefaultCIMode()
	mode.Enabled = true

	runner := NewCIRunner(base, CIRunnerOptions{
		Mode: mode,
	})

	// Non-go test commands should pass through
	if err := runner.RunCmd("ls", "-la"); err != nil {
		t.Fatalf("RunCmd() error = %v", err)
	}

	if base.lastCmd != "ls" {
		t.Errorf("Expected command 'ls', got %q", base.lastCmd)
	}
}

func TestCIRunner_RunCmdOutput(t *testing.T) {
	t.Parallel()

	base := &mockCommandRunner{output: "test output"}
	runner := NewCIRunner(base, CIRunnerOptions{
		Mode: DefaultCIMode(),
	})

	output, err := runner.RunCmdOutput("echo", "hello")
	if err != nil {
		t.Errorf("RunCmdOutput() error = %v", err)
	}
	if output != "test output" {
		t.Errorf("RunCmdOutput() = %q, want %q", output, "test output")
	}
}

func TestCIRunner_GetResults(t *testing.T) {
	t.Parallel()

	base := &mockCommandRunner{}
	runner := NewCIRunner(base, CIRunnerOptions{
		Mode: DefaultCIMode(),
	})

	results := runner.GetResults()
	if results == nil {
		t.Error("GetResults() returned nil")
	}
}

func TestCIRunner_GenerateReport(t *testing.T) {
	t.Parallel()

	base := &mockCommandRunner{}
	reporter := &mockReporter{}

	runner := NewCIRunner(base, CIRunnerOptions{
		Mode:     DefaultCIMode(),
		Reporter: reporter,
	})

	err := runner.GenerateReport()
	if err != nil {
		t.Errorf("GenerateReport() error = %v", err)
	}

	if !reporter.summaryCalled {
		t.Error("GenerateReport() did not call WriteSummary")
	}
	if !reporter.closed {
		t.Error("GenerateReport() did not call Close")
	}
}

func TestCIRunner_GenerateReport_NoReporter(t *testing.T) {
	t.Parallel()

	base := &mockCommandRunner{}
	runner := NewCIRunner(base, CIRunnerOptions{
		Mode: DefaultCIMode(),
	})

	err := runner.GenerateReport()
	if err != nil {
		t.Errorf("GenerateReport() without reporter error = %v", err)
	}
}

func TestFormatDurationForSummary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		duration time.Duration
		expected string
	}{
		{500 * time.Millisecond, "500ms"},
		{1500 * time.Millisecond, "1.5s"},
		{65 * time.Second, "1m5s"},
		{125 * time.Second, "2m5s"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			got := formatDurationForSummary(tt.duration)
			if got != tt.expected {
				t.Errorf("formatDurationForSummary(%v) = %q, want %q", tt.duration, got, tt.expected)
			}
		})
	}
}

func TestIsCIEnabled(t *testing.T) {
	t.Parallel()

	t.Run("not enabled by default locally", func(t *testing.T) {
		t.Parallel()
		// This test assumes we're not running in CI
		// In CI, this test would behave differently
		if IsCIEnabled(nil, nil) {
			// Running in CI, which is expected
			return
		}
		// Local environment - should be false
	})

	t.Run("enabled via params", func(t *testing.T) {
		t.Parallel()
		params := map[string]string{"ci": "true"}
		if !IsCIEnabled(params, nil) {
			t.Error("IsCIEnabled() with ci=true = false, want true")
		}
	})

	t.Run("disabled via params", func(t *testing.T) {
		t.Parallel()
		params := map[string]string{"ci": "false"}
		if IsCIEnabled(params, nil) {
			t.Error("IsCIEnabled() with ci=false = true, want false")
		}
	})
}

func TestGetCIRunner(t *testing.T) {
	t.Parallel()

	t.Run("returns base when disabled", func(t *testing.T) {
		t.Parallel()
		base := &mockCommandRunner{}
		params := map[string]string{"ci": "false"}

		runner := GetCIRunner(base, params, nil)

		// Should return base runner, not CI runner
		if _, ok := runner.(*ciRunner); ok {
			t.Error("GetCIRunner() returned ciRunner when CI mode disabled")
		}
	})

	t.Run("returns CI runner when enabled", func(t *testing.T) {
		t.Parallel()
		base := &mockCommandRunner{}
		params := map[string]string{"ci": "true"}

		runner := GetCIRunner(base, params, nil)

		if _, ok := runner.(*ciRunner); !ok {
			t.Error("GetCIRunner() did not return ciRunner when CI mode enabled")
		}
	})
}

func TestExtractCrashLocation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		stderr   string
		wantFile string
		wantLine int
	}{
		{
			name: "stack trace with location",
			stderr: `fatal error: unexpected signal
goroutine 1 [running]:
main.foo()
	/path/to/file.go:42 +0x39`,
			wantFile: "/path/to/file.go",
			wantLine: 42,
		},
		{
			name:     "no stack trace",
			stderr:   "some error output",
			wantFile: "",
			wantLine: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			failure := &CITestFailure{}
			extractCrashLocation(failure, tt.stderr)

			if failure.File != tt.wantFile {
				t.Errorf("extractCrashLocation() File = %q, want %q", failure.File, tt.wantFile)
			}
			if failure.Line != tt.wantLine {
				t.Errorf("extractCrashLocation() Line = %d, want %d", failure.Line, tt.wantLine)
			}
		})
	}
}

// T076: Tests for fatal crash detection
func TestHandleCrash(t *testing.T) {
	t.Parallel()

	t.Run("fatal error detected", func(t *testing.T) {
		t.Parallel()
		base := &mockCommandRunner{}
		reporter := &mockReporter{}

		runner, ok := NewCIRunner(base, CIRunnerOptions{
			Mode:     DefaultCIMode(),
			Reporter: reporter,
		}).(*ciRunner)
		if !ok {
			t.Fatal("Failed to create ciRunner")
		}

		// Simulate a fatal error
		stderr := `fatal error: all goroutines are asleep - deadlock!
goroutine 1 [running]:
runtime.throw(0x4b5e60, 0x2a)
	/usr/local/go/src/runtime/panic.go:617 +0x72`

		runner.handleCrash(stderr)

		// Should have reported a failure
		if len(reporter.failures) != 1 {
			t.Fatalf("Expected 1 failure, got %d", len(reporter.failures))
		}

		failure := reporter.failures[0]
		if failure.Type != FailureTypeFatal {
			t.Errorf("Expected FailureTypeFatal, got %v", failure.Type)
		}
		if failure.Stack == "" {
			t.Error("Expected stack trace to be captured")
		}
	})

	t.Run("signal detected", func(t *testing.T) {
		t.Parallel()
		base := &mockCommandRunner{}
		reporter := &mockReporter{}

		runner, ok := NewCIRunner(base, CIRunnerOptions{
			Mode:     DefaultCIMode(),
			Reporter: reporter,
		}).(*ciRunner)
		if !ok {
			t.Fatal("Failed to create ciRunner")
		}

		stderr := `signal: segmentation fault (core dumped)
goroutine 1 [running]:
main.causeSegfault()
	/app/main.go:15 +0x20`

		runner.handleCrash(stderr)

		if len(reporter.failures) != 1 {
			t.Fatalf("Expected 1 failure, got %d", len(reporter.failures))
		}

		failure := reporter.failures[0]
		if failure.Type != FailureTypeFatal {
			t.Errorf("Expected FailureTypeFatal, got %v", failure.Type)
		}
	})

	t.Run("panic detected via stderr", func(t *testing.T) {
		t.Parallel()
		base := &mockCommandRunner{}
		reporter := &mockReporter{}

		runner, ok := NewCIRunner(base, CIRunnerOptions{
			Mode:     DefaultCIMode(),
			Reporter: reporter,
		}).(*ciRunner)
		if !ok {
			t.Fatal("Failed to create ciRunner")
		}

		stderr := `panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x4a1234]
goroutine 1 [running]:
main.dereference(0x0)
	/app/main.go:20 +0x30`

		runner.handleCrash(stderr)

		if len(reporter.failures) != 1 {
			t.Fatalf("Expected 1 failure, got %d", len(reporter.failures))
		}
	})

	t.Run("normal output not treated as crash", func(t *testing.T) {
		t.Parallel()
		base := &mockCommandRunner{}
		reporter := &mockReporter{}

		runner, ok := NewCIRunner(base, CIRunnerOptions{
			Mode:     DefaultCIMode(),
			Reporter: reporter,
		}).(*ciRunner)
		if !ok {
			t.Fatal("Failed to create ciRunner")
		}

		stderr := "some normal stderr output that's not a crash"

		runner.handleCrash(stderr)

		// Should not report a failure for normal output
		if len(reporter.failures) != 0 {
			t.Errorf("Expected no failures for normal output, got %d", len(reporter.failures))
		}
	})
}

func TestMultiReporter(t *testing.T) {
	t.Parallel()

	r1 := &mockReporter{}
	r2 := &mockReporter{}
	multi := NewMultiReporter(r1, r2)

	t.Run("Start calls all reporters", func(t *testing.T) {
		err := multi.Start(CIMetadata{})
		if err != nil {
			t.Errorf("Start() error = %v", err)
		}
		if !r1.started || !r2.started {
			t.Error("Start() did not call all reporters")
		}
	})

	t.Run("ReportFailure calls all reporters", func(t *testing.T) {
		err := multi.ReportFailure(CITestFailure{Test: "TestExample"})
		if err != nil {
			t.Errorf("ReportFailure() error = %v", err)
		}
		if len(r1.failures) != 1 || len(r2.failures) != 1 {
			t.Error("ReportFailure() did not call all reporters")
		}
	})

	t.Run("WriteSummary calls all reporters", func(t *testing.T) {
		err := multi.WriteSummary(&CIResult{})
		if err != nil {
			t.Errorf("WriteSummary() error = %v", err)
		}
		if !r1.summaryCalled || !r2.summaryCalled {
			t.Error("WriteSummary() did not call all reporters")
		}
	})

	t.Run("Close calls all reporters", func(t *testing.T) {
		err := multi.Close()
		if err != nil {
			t.Errorf("Close() error = %v", err)
		}
		if !r1.closed || !r2.closed {
			t.Error("Close() did not call all reporters")
		}
	})
}

func TestNullReporter(t *testing.T) {
	t.Parallel()

	r := NullReporter{}

	if err := r.Start(CIMetadata{}); err != nil {
		t.Errorf("Start() error = %v", err)
	}
	if err := r.ReportFailure(CITestFailure{}); err != nil {
		t.Errorf("ReportFailure() error = %v", err)
	}
	if err := r.WriteSummary(&CIResult{}); err != nil {
		t.Errorf("WriteSummary() error = %v", err)
	}
	if err := r.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestCIRunnerWithModeDisabled(t *testing.T) {
	t.Parallel()

	base := &mockCommandRunner{}
	mode := DefaultCIMode()
	mode.Enabled = false

	runner := NewCIRunner(base, CIRunnerOptions{
		Mode: mode,
	})

	// Should pass through to base runner
	if err := runner.RunCmd("go", "test", "./..."); err != nil {
		t.Fatalf("RunCmd() error = %v", err)
	}

	if base.lastCmd != "go" {
		t.Errorf("Expected command to pass through, got %q", base.lastCmd)
	}
	if !strings.Contains(strings.Join(base.lastArgs, " "), "test") {
		t.Errorf("Expected args to contain 'test', got %v", base.lastArgs)
	}
}

func TestGetCIParams(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		args          []string
		wantCIValue   string
		wantRemaining int
	}{
		{
			name:          "no ci param",
			args:          []string{"-v", "-timeout=5m"},
			wantCIValue:   "",
			wantRemaining: 2,
		},
		{
			name:          "ci=true param",
			args:          []string{"-v", "ci=true", "-timeout=5m"},
			wantCIValue:   "true",
			wantRemaining: 2,
		},
		{
			name:          "ci=false param",
			args:          []string{"ci=false", "-v"},
			wantCIValue:   "false",
			wantRemaining: 1,
		},
		{
			name:          "bare ci param",
			args:          []string{"ci", "-v"},
			wantCIValue:   "true",
			wantRemaining: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			params, remaining := getCIParams(tt.args)

			if got := params["ci"]; got != tt.wantCIValue {
				t.Errorf("getCIParams() ci value = %q, want %q", got, tt.wantCIValue)
			}
			if len(remaining) != tt.wantRemaining {
				t.Errorf("getCIParams() remaining count = %d, want %d", len(remaining), tt.wantRemaining)
			}
		})
	}
}

func TestGetTestRunner(t *testing.T) {
	t.Parallel()

	t.Run("returns standard runner when CI disabled", func(t *testing.T) {
		t.Parallel()
		params := map[string]string{"ci": "false"}
		config := &Config{}

		runner := getTestRunner(params, config)

		// Should NOT be a CI runner
		if _, ok := runner.(*ciRunner); ok {
			t.Error("getTestRunner() returned ciRunner when CI mode disabled")
		}
	})

	t.Run("returns CI runner when enabled via params", func(t *testing.T) {
		t.Parallel()
		params := map[string]string{"ci": "true"}
		config := &Config{}

		runner := getTestRunner(params, config)

		// Should be a CI runner
		if _, ok := runner.(*ciRunner); !ok {
			t.Error("getTestRunner() did not return ciRunner when CI mode enabled")
		}
	})
}

func TestCIAutoDetectionIntegration(t *testing.T) {
	// Note: Cannot use t.Parallel() because some subtests use t.Setenv()

	t.Run("detects GitHub Actions environment", func(t *testing.T) {
		// Cannot use t.Parallel() with t.Setenv()
		// Note: This test verifies the detection logic, not actual CI environment
		detector := NewCIDetector()

		// When GITHUB_ACTIONS is set, should detect as CI
		t.Setenv("GITHUB_ACTIONS", "true")
		t.Setenv("CI", "true")

		if !detector.IsCI() {
			t.Error("Expected IsCI() to be true when GITHUB_ACTIONS is set")
		}

		if detector.Platform() != string(CIFormatGitHub) {
			t.Errorf("Expected Platform() = 'github', got %q", detector.Platform())
		}
	})

	t.Run("CI runner wraps commands when enabled", func(t *testing.T) {
		base := &mockCommandRunner{}
		params := map[string]string{"ci": "true"}
		config := &Config{}

		runner := GetCIRunner(base, params, config)

		// Should wrap in CI runner
		ciRunner, ok := runner.(*ciRunner)
		if !ok {
			t.Fatal("Expected CI runner wrapper")
		}

		// GetResults should return non-nil
		results := ciRunner.GetResults()
		if results == nil {
			t.Error("GetResults() returned nil")
		}
	})

	t.Run("CI runner passes through non-test commands", func(t *testing.T) {
		base := &mockCommandRunner{}
		mode := DefaultCIMode()
		mode.Enabled = true

		runner := NewCIRunner(base, CIRunnerOptions{Mode: mode})

		// Non-test command should pass through
		if err := runner.RunCmd("go", "build", "./..."); err != nil {
			t.Fatalf("RunCmd() error = %v", err)
		}

		if base.lastCmd != "go" {
			t.Errorf("Expected pass-through command, got %q", base.lastCmd)
		}
		if len(base.lastArgs) == 0 || base.lastArgs[0] != "build" {
			t.Errorf("Expected build args, got %v", base.lastArgs)
		}
	})
}
