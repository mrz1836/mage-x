// Package mage provides CI mode types and constants for test output processing
package mage

import (
	"time"
)

// FailureType classifies the type of test failure
type FailureType string

const (
	// FailureTypeTest represents a test assertion failure
	FailureTypeTest FailureType = "test"
	// FailureTypeBuild represents a compilation error
	FailureTypeBuild FailureType = "build"
	// FailureTypePanic represents a runtime panic
	FailureTypePanic FailureType = "panic"
	// FailureTypeRace represents a data race detected by the race detector
	FailureTypeRace FailureType = "race"
	// FailureTypeFuzz represents a fuzz test failure
	FailureTypeFuzz FailureType = "fuzz"
	// FailureTypeTimeout represents a test timeout
	FailureTypeTimeout FailureType = "timeout"
	// FailureTypeFatal represents a test binary crash (SIGSEGV, cgo crash)
	FailureTypeFatal FailureType = "fatal"
)

// FuzzInfo contains fuzz test specific failure details
type FuzzInfo struct {
	CorpusPath string `json:"corpus_path,omitempty"` // Path to failing input file
	Input      string `json:"input,omitempty"`       // Failing input (base64 if binary)
	IsBinary   bool   `json:"is_binary,omitempty"`   // True if input is binary
	FromSeed   bool   `json:"from_seed,omitempty"`   // True if from seed corpus
}

// CIFormat specifies the output format for CI mode
type CIFormat string

const (
	// CIFormatAuto auto-detects format (github in GHA, json otherwise)
	CIFormatAuto CIFormat = "auto"
	// CIFormatGitHub outputs GitHub Actions annotations
	CIFormatGitHub CIFormat = "github"
	// CIFormatJSON outputs JSON Lines file only
	CIFormatJSON CIFormat = "json"
)

// CIMode represents CI mode configuration
type CIMode struct {
	Enabled      bool     `yaml:"enabled" json:"enabled"`
	Format       CIFormat `yaml:"format" json:"format"`
	ContextLines int      `yaml:"context_lines" json:"context_lines"`
	MaxMemoryMB  int      `yaml:"max_memory_mb" json:"max_memory_mb"`
	Dedup        bool     `yaml:"dedup" json:"dedup"`
	OutputPath   string   `yaml:"output_path" json:"output_path"`
}

// DefaultCIMode returns default CI configuration
func DefaultCIMode() CIMode {
	return CIMode{
		Enabled:      false, // Auto-detect enables it
		Format:       CIFormatAuto,
		ContextLines: 20,
		MaxMemoryMB:  100,
		Dedup:        true,
		OutputPath:   ".mage-x/ci-results.jsonl",
	}
}

// Validate validates the CIMode configuration
func (m *CIMode) Validate() error {
	if m.ContextLines < 0 || m.ContextLines > 100 {
		return ErrCIContextLinesOutOfRange
	}
	if m.MaxMemoryMB < 10 || m.MaxMemoryMB > 1000 {
		return ErrCIMaxMemoryOutOfRange
	}
	return nil
}

// TestStatus represents overall test run status
type TestStatus string

const (
	// TestStatusPassed indicates all tests passed
	TestStatusPassed TestStatus = "passed"
	// TestStatusFailed indicates one or more tests failed
	TestStatusFailed TestStatus = "failed"
	// TestStatusError indicates build/setup errors
	TestStatusError TestStatus = "error"
)

// CISummary contains test count statistics
type CISummary struct {
	Status      TestStatus `json:"status"`
	Total       int        `json:"total"`        // Total test runs (may include duplicates from build tag runs)
	UniqueTotal int        `json:"unique_total"` // Unique test cases (deduplicated across build tag runs)
	Passed      int        `json:"passed"`
	Failed      int        `json:"failed"`
	Skipped     int        `json:"skipped"`
	Duration    string     `json:"duration"`
}

// CIMetadata contains execution context information
type CIMetadata struct {
	Branch    string `json:"branch,omitempty"`
	Commit    string `json:"commit,omitempty"`
	RunID     string `json:"run_id,omitempty"`
	Workflow  string `json:"workflow,omitempty"`
	Platform  string `json:"platform"`
	GoVersion string `json:"go_version"`
}

// CIResult contains aggregated test execution results
type CIResult struct {
	Summary   CISummary       `json:"summary"`
	Failures  []CITestFailure `json:"failures"`
	Timestamp time.Time       `json:"timestamp"`
	Duration  time.Duration   `json:"duration"`
	Metadata  CIMetadata      `json:"metadata"`
}

// CITestFailure represents a test failure with detailed CI information
// This extends the basic TestFailure with CI-specific fields
type CITestFailure struct {
	// Core fields from TestFailure (backwards compatible)
	Package string `json:"package"`
	Test    string `json:"test"`
	Error   string `json:"error"`
	Output  string `json:"output"`

	// CI-specific fields
	Type        FailureType `json:"type"`
	File        string      `json:"file"`
	Line        int         `json:"line"`
	Column      int         `json:"column,omitempty"`
	Stack       string      `json:"stack,omitempty"`
	Context     []string    `json:"context,omitempty"`
	Signature   string      `json:"signature"`
	Duration    string      `json:"duration,omitempty"`
	RaceRelated bool        `json:"race_related,omitempty"`
	FuzzInfo    *FuzzInfo   `json:"fuzz_info,omitempty"`
}

// ToTestFailure converts CITestFailure to basic TestFailure for backwards compatibility
func (f CITestFailure) ToTestFailure() TestFailure {
	return TestFailure{
		Package: f.Package,
		Test:    f.Test,
		Error:   f.Error,
		Output:  f.Output,
	}
}

// CaptureStrategy determines memory usage pattern for stream processing
type CaptureStrategy int

const (
	// StrategyFullCapture keeps everything (<100 tests)
	StrategyFullCapture CaptureStrategy = iota
	// StrategySmartCapture captures full details for failures only (<1000 tests)
	StrategySmartCapture
	// StrategyEfficientCapture uses limited context (<5000 tests)
	StrategyEfficientCapture
	// StrategyStreamingCapture uses disk buffer for very large suites (5000+ tests)
	StrategyStreamingCapture
)

// String returns the string representation of the capture strategy
func (s CaptureStrategy) String() string {
	switch s {
	case StrategyFullCapture:
		return "full"
	case StrategySmartCapture:
		return "smart"
	case StrategyEfficientCapture:
		return "efficient"
	case StrategyStreamingCapture:
		return "streaming"
	default:
		return "unknown"
	}
}

// SelectStrategy selects the appropriate capture strategy based on test count
func SelectStrategy(testCount int) CaptureStrategy {
	switch {
	case testCount < 100:
		return StrategyFullCapture
	case testCount < 1000:
		return StrategySmartCapture
	case testCount < 5000:
		return StrategyEfficientCapture
	default:
		return StrategyStreamingCapture
	}
}

// CI mode static errors
var (
	ErrCIContextLinesOutOfRange = ciContextLinesOutOfRangeError{}
	ErrCIMaxMemoryOutOfRange    = ciMaxMemoryOutOfRangeError{}
)

type ciContextLinesOutOfRangeError struct{}

func (ciContextLinesOutOfRangeError) Error() string {
	return "context_lines must be between 0 and 100"
}

type ciMaxMemoryOutOfRangeError struct{}

func (ciMaxMemoryOutOfRangeError) Error() string {
	return "max_memory_mb must be between 10 and 1000"
}
