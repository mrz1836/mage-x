package mage

import (
	"errors"
	"testing"
)

func TestFailureTypeConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		ft       FailureType
		expected string
	}{
		{"test failure", FailureTypeTest, "test"},
		{"build failure", FailureTypeBuild, "build"},
		{"panic failure", FailureTypePanic, "panic"},
		{"race failure", FailureTypeRace, "race"},
		{"fuzz failure", FailureTypeFuzz, "fuzz"},
		{"timeout failure", FailureTypeTimeout, "timeout"},
		{"fatal failure", FailureTypeFatal, "fatal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.ft) != tt.expected {
				t.Errorf("FailureType %s = %q, want %q", tt.name, tt.ft, tt.expected)
			}
		})
	}
}

func TestCIFormatConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		format   CIFormat
		expected string
	}{
		{"auto format", CIFormatAuto, "auto"},
		{"github format", CIFormatGitHub, "github"},
		{"json format", CIFormatJSON, "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.format) != tt.expected {
				t.Errorf("CIFormat %s = %q, want %q", tt.name, tt.format, tt.expected)
			}
		})
	}
}

func TestDefaultCIMode(t *testing.T) {
	t.Parallel()

	mode := DefaultCIMode()

	if mode.Enabled {
		t.Error("DefaultCIMode().Enabled = true, want false")
	}
	if mode.Format != CIFormatAuto {
		t.Errorf("DefaultCIMode().Format = %q, want %q", mode.Format, CIFormatAuto)
	}
	if mode.ContextLines != 20 {
		t.Errorf("DefaultCIMode().ContextLines = %d, want 20", mode.ContextLines)
	}
	if mode.MaxMemoryMB != 100 {
		t.Errorf("DefaultCIMode().MaxMemoryMB = %d, want 100", mode.MaxMemoryMB)
	}
	if !mode.Dedup {
		t.Error("DefaultCIMode().Dedup = false, want true")
	}
	if mode.OutputPath != ".mage-x/ci-results.jsonl" {
		t.Errorf("DefaultCIMode().OutputPath = %q, want %q", mode.OutputPath, ".mage-x/ci-results.jsonl")
	}
}

func TestCIModeValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		mode    CIMode
		wantErr bool
		errType error
	}{
		{
			name:    "valid default mode",
			mode:    DefaultCIMode(),
			wantErr: false,
		},
		{
			name: "valid custom mode",
			mode: CIMode{
				ContextLines: 50,
				MaxMemoryMB:  500,
			},
			wantErr: false,
		},
		{
			name: "context_lines too low",
			mode: CIMode{
				ContextLines: -1,
				MaxMemoryMB:  100,
			},
			wantErr: true,
			errType: ErrCIContextLinesOutOfRange,
		},
		{
			name: "context_lines too high",
			mode: CIMode{
				ContextLines: 101,
				MaxMemoryMB:  100,
			},
			wantErr: true,
			errType: ErrCIContextLinesOutOfRange,
		},
		{
			name: "max_memory_mb too low",
			mode: CIMode{
				ContextLines: 20,
				MaxMemoryMB:  5,
			},
			wantErr: true,
			errType: ErrCIMaxMemoryOutOfRange,
		},
		{
			name: "max_memory_mb too high",
			mode: CIMode{
				ContextLines: 20,
				MaxMemoryMB:  2000,
			},
			wantErr: true,
			errType: ErrCIMaxMemoryOutOfRange,
		},
		{
			name: "boundary values valid",
			mode: CIMode{
				ContextLines: 0,
				MaxMemoryMB:  10,
			},
			wantErr: false,
		},
		{
			name: "upper boundary values valid",
			mode: CIMode{
				ContextLines: 100,
				MaxMemoryMB:  1000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.mode.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("CIMode.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errType != nil && !errors.Is(err, tt.errType) {
				t.Errorf("CIMode.Validate() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestTestStatusConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		status   TestStatus
		expected string
	}{
		{"passed status", TestStatusPassed, "passed"},
		{"failed status", TestStatusFailed, "failed"},
		{"error status", TestStatusError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.status) != tt.expected {
				t.Errorf("TestStatus %s = %q, want %q", tt.name, tt.status, tt.expected)
			}
		})
	}
}

func TestCaptureStrategyString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		strategy CaptureStrategy
		expected string
	}{
		{StrategyFullCapture, "full"},
		{StrategySmartCapture, "smart"},
		{StrategyEfficientCapture, "efficient"},
		{StrategyStreamingCapture, "streaming"},
		{CaptureStrategy(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			if got := tt.strategy.String(); got != tt.expected {
				t.Errorf("CaptureStrategy.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSelectStrategy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		testCount int
		expected  CaptureStrategy
	}{
		{"very small suite", 10, StrategyFullCapture},
		{"small suite boundary", 99, StrategyFullCapture},
		{"medium suite lower", 100, StrategySmartCapture},
		{"medium suite", 500, StrategySmartCapture},
		{"medium suite boundary", 999, StrategySmartCapture},
		{"large suite lower", 1000, StrategyEfficientCapture},
		{"large suite", 3000, StrategyEfficientCapture},
		{"large suite boundary", 4999, StrategyEfficientCapture},
		{"very large suite lower", 5000, StrategyStreamingCapture},
		{"very large suite", 10000, StrategyStreamingCapture},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := SelectStrategy(tt.testCount); got != tt.expected {
				t.Errorf("SelectStrategy(%d) = %v, want %v", tt.testCount, got, tt.expected)
			}
		})
	}
}

func TestCITestFailureToTestFailure(t *testing.T) {
	t.Parallel()

	ciFailure := CITestFailure{
		Package:   "github.com/example/pkg",
		Test:      "TestSomething",
		Error:     "assertion failed",
		Output:    "test output here",
		Type:      FailureTypeTest,
		File:      "pkg_test.go",
		Line:      42,
		Signature: "github.com/example/pkg:TestSomething:pkg_test.go:42:test",
	}

	tf := ciFailure.ToTestFailure()

	if tf.Package != ciFailure.Package {
		t.Errorf("ToTestFailure().Package = %q, want %q", tf.Package, ciFailure.Package)
	}
	if tf.Test != ciFailure.Test {
		t.Errorf("ToTestFailure().Test = %q, want %q", tf.Test, ciFailure.Test)
	}
	if tf.Error != ciFailure.Error {
		t.Errorf("ToTestFailure().Error = %q, want %q", tf.Error, ciFailure.Error)
	}
	if tf.Output != ciFailure.Output {
		t.Errorf("ToTestFailure().Output = %q, want %q", tf.Output, ciFailure.Output)
	}
}

func TestFuzzInfo(t *testing.T) {
	t.Parallel()

	info := &FuzzInfo{
		CorpusPath: "testdata/fuzz/FuzzFoo/582528ddfad69eb5",
		Input:      "test input",
		IsBinary:   false,
		FromSeed:   true,
	}

	if info.CorpusPath != "testdata/fuzz/FuzzFoo/582528ddfad69eb5" {
		t.Errorf("FuzzInfo.CorpusPath = %q, want %q", info.CorpusPath, "testdata/fuzz/FuzzFoo/582528ddfad69eb5")
	}
	if info.Input != "test input" {
		t.Errorf("FuzzInfo.Input = %q, want %q", info.Input, "test input")
	}
	if info.IsBinary {
		t.Error("FuzzInfo.IsBinary = true, want false")
	}
	if !info.FromSeed {
		t.Error("FuzzInfo.FromSeed = false, want true")
	}
}

func TestCISummary(t *testing.T) {
	t.Parallel()

	summary := CISummary{
		Status:   TestStatusFailed,
		Total:    100,
		Passed:   98,
		Failed:   2,
		Skipped:  0,
		Duration: "2m30s",
	}

	if summary.Status != TestStatusFailed {
		t.Errorf("CISummary.Status = %q, want %q", summary.Status, TestStatusFailed)
	}
	if summary.Total != 100 {
		t.Errorf("CISummary.Total = %d, want 100", summary.Total)
	}
	if summary.Passed != 98 {
		t.Errorf("CISummary.Passed = %d, want 98", summary.Passed)
	}
	if summary.Failed != 2 {
		t.Errorf("CISummary.Failed = %d, want 2", summary.Failed)
	}
}

func TestCIMetadata(t *testing.T) {
	t.Parallel()

	metadata := CIMetadata{
		Branch:    "main",
		Commit:    "abc123",
		RunID:     "12345",
		Workflow:  "CI",
		Platform:  "github",
		GoVersion: "go1.24",
	}

	if metadata.Branch != "main" {
		t.Errorf("CIMetadata.Branch = %q, want %q", metadata.Branch, "main")
	}
	if metadata.Commit != "abc123" {
		t.Errorf("CIMetadata.Commit = %q, want %q", metadata.Commit, "abc123")
	}
	if metadata.Platform != string(CIFormatGitHub) {
		t.Errorf("CIMetadata.Platform = %q, want %q", metadata.Platform, string(CIFormatGitHub))
	}
}

func TestErrorTypes(t *testing.T) {
	t.Parallel()

	t.Run("ErrCIContextLinesOutOfRange", func(t *testing.T) {
		t.Parallel()
		err := ErrCIContextLinesOutOfRange
		expected := "context_lines must be between 0 and 100"
		if err.Error() != expected {
			t.Errorf("ErrCIContextLinesOutOfRange.Error() = %q, want %q", err.Error(), expected)
		}
	})

	t.Run("ErrCIMaxMemoryOutOfRange", func(t *testing.T) {
		t.Parallel()
		err := ErrCIMaxMemoryOutOfRange
		expected := "max_memory_mb must be between 10 and 1000"
		if err.Error() != expected {
			t.Errorf("ErrCIMaxMemoryOutOfRange.Error() = %q, want %q", err.Error(), expected)
		}
	})
}
