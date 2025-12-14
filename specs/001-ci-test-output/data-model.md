# Data Model: CI Test Output Mode

**Feature Branch**: `001-ci-test-output`
**Date**: 2025-12-12

## Entity Definitions

### 1. TestFailure (Extended)

Extends existing `TestFailure` struct in `pkg/mage/interfaces.go` with CI-specific fields.

```go
// TestFailure represents a failed test with detailed location information
type TestFailure struct {
    // Existing fields (backwards compatible)
    Package string `json:"package"`          // Go package path
    Test    string `json:"test"`             // Test name (full hierarchy for subtests)
    Error   string `json:"error"`            // Error message
    Output  string `json:"output"`           // Full test output

    // New CI fields
    Type      FailureType `json:"type"`      // Failure classification
    File      string      `json:"file"`      // Source file path (relative)
    Line      int         `json:"line"`      // Line number
    Column    int         `json:"column"`    // Column (for build errors)
    Stack     string      `json:"stack"`     // Stack trace (for panics)
    Context     []string  `json:"context"`               // Surrounding lines of code
    Signature   string    `json:"signature"`             // Deduplication key (pkg:test:file:line:type)
    Duration    string    `json:"duration"`              // Test duration
    RaceRelated bool      `json:"race_related,omitempty"` // True if panic was triggered by race detector
    FuzzInfo    *FuzzInfo `json:"fuzz_info,omitempty"`   // Fuzz-specific details (only for fuzz failures)
}

// FuzzInfo contains fuzz test specific failure details
type FuzzInfo struct {
    CorpusPath string `json:"corpus_path,omitempty"` // Path to failing input file
    Input      string `json:"input,omitempty"`       // Failing input (base64 if binary)
    IsBinary   bool   `json:"is_binary,omitempty"`   // True if input is binary
    FromSeed   bool   `json:"from_seed,omitempty"`   // True if from seed corpus
}

// FailureType classifies the type of test failure
type FailureType string

const (
    FailureTypeTest    FailureType = "test"    // Test assertion failure
    FailureTypeBuild   FailureType = "build"   // Compilation error
    FailureTypePanic   FailureType = "panic"   // Runtime panic
    FailureTypeRace    FailureType = "race"    // Data race detected
    FailureTypeFuzz    FailureType = "fuzz"    // Fuzz test failure
    FailureTypeTimeout FailureType = "timeout" // Test timeout
    FailureTypeFatal   FailureType = "fatal"   // Test binary crash (SIGSEGV, cgo crash)
)
```

**Validation Rules**:
- `Package`: Required, must be valid Go package path
- `Test`: Required for test/panic/race/timeout, empty for build errors
- `File`: Required, relative path from project root
- `Line`: Required, must be > 0
- `Column`: Optional, only set for build errors
- `Signature`: Auto-generated as `Package:Test:File:Line:Type` (includes Type to preserve multiple failure types for same location)

**State Transitions**: N/A (immutable after creation)

---

### 2. CIResult

Aggregated test results for CI output.

```go
// CIResult contains aggregated test execution results
type CIResult struct {
    Summary   CISummary     `json:"summary"`
    Failures  []TestFailure `json:"failures"`
    Timestamp time.Time     `json:"timestamp"`
    Duration  time.Duration `json:"duration"`
    Metadata  CIMetadata    `json:"metadata"`
}

// CISummary contains test count statistics
type CISummary struct {
    Status   TestStatus `json:"status"`   // passed, failed, error
    Total    int        `json:"total"`    // Total tests run
    Passed   int        `json:"passed"`   // Tests passed
    Failed   int        `json:"failed"`   // Tests failed
    Skipped  int        `json:"skipped"`  // Tests skipped
    Duration string     `json:"duration"` // Total duration
}

// TestStatus represents overall test run status
type TestStatus string

const (
    TestStatusPassed TestStatus = "passed"
    TestStatusFailed TestStatus = "failed"
    TestStatusError  TestStatus = "error" // Build/setup errors
)

// CIMetadata contains execution context
type CIMetadata struct {
    Branch    string `json:"branch,omitempty"`
    Commit    string `json:"commit,omitempty"`
    RunID     string `json:"run_id,omitempty"`
    Workflow  string `json:"workflow,omitempty"`
    Platform  string `json:"platform"`
    GoVersion string `json:"go_version"`
}
```

**Validation Rules**:
- `Summary.Total` >= `Summary.Passed` + `Summary.Failed` + `Summary.Skipped`
- `Summary.Status` = "passed" iff `Summary.Failed` == 0
- `Timestamp` required, RFC3339 format
- `Failures` slice may be empty (if all tests pass)

---

### 3. CIMode

Configuration for CI behavior.

```go
// CIMode represents CI mode configuration
type CIMode struct {
    Enabled      bool      `yaml:"enabled" json:"enabled"`
    Format       CIFormat  `yaml:"format" json:"format"`
    ContextLines int       `yaml:"context_lines" json:"context_lines"`
    MaxMemoryMB  int       `yaml:"max_memory_mb" json:"max_memory_mb"`
    Dedup        bool      `yaml:"dedup" json:"dedup"`
    OutputPath   string    `yaml:"output_path" json:"output_path"`
}

// CIFormat specifies the output format
type CIFormat string

const (
    CIFormatAuto   CIFormat = "auto"   // Auto-detect (github in GHA, json otherwise)
    CIFormatGitHub CIFormat = "github" // GitHub Actions annotations
    CIFormatJSON   CIFormat = "json"   // JSON Lines file only
)

// DefaultCIMode returns default CI configuration
func DefaultCIMode() CIMode {
    return CIMode{
        Enabled:      false,      // Auto-detect enables it
        Format:       CIFormatAuto,
        ContextLines: 20,
        MaxMemoryMB:  100,
        Dedup:        true,
        OutputPath:   ".mage-x/ci-results.jsonl",
    }
}
```

**Validation Rules**:
- `ContextLines` must be >= 0 and <= 100
- `MaxMemoryMB` must be >= 10 and <= 1000
- `OutputPath` must be a valid file path (relative or absolute)

---

### 4. StreamParser

Real-time processor of test output (internal).

```go
// StreamParser processes Go test JSON output in real-time
type StreamParser struct {
    contextBuffer *RingBuffer     // Circular buffer for context lines
    currentTest   string          // Currently running test name
    currentPkg    string          // Currently running package
    failures      []TestFailure   // Collected failures
    testCount     int32           // Total tests seen (atomic)
    strategy      CaptureStrategy // Memory strategy
    mu            sync.Mutex      // Thread safety for failures
}

// CaptureStrategy determines memory usage pattern
type CaptureStrategy int

const (
    StrategyFullCapture      CaptureStrategy = iota // Keep everything (<100 tests)
    StrategySmartCapture                            // Full for failures only (<1000 tests)
    StrategyEfficientCapture                        // Limited context (<5000 tests)
    StrategyStreamingCapture                        // Disk buffer (5000+ tests)
)

// RingBuffer implements a fixed-size circular buffer for context lines
type RingBuffer struct {
    lines []string
    size  int
    head  int
    count int
}
```

**State Transitions**:
```
Initial -> Parsing -> Complete
              |
              v
           Failed (on error)
```

---

### 5. CIRunner

CommandRunner wrapper for CI mode (internal).

```go
// CIRunner wraps CommandRunner to intercept test output
type CIRunner struct {
    base     CommandRunner
    parser   *StreamParser
    ciMode   CIMode
    results  *CIResult
    reporter CIReporter
}

// CIReporter interface for different output formats
type CIReporter interface {
    ReportFailure(failure TestFailure) error
    WriteSummary(result *CIResult) error
    Close() error
}
```

---

## Relationships

```
┌─────────────────────────────────────────────────────────────────┐
│                        Test Namespace                           │
│                                                                 │
│  ┌──────────┐    uses     ┌──────────┐                         │
│  │ Unit()   │────────────>│ CIRunner │                         │
│  │ Race()   │             └────┬─────┘                         │
│  │ Cover()  │                  │                               │
│  └──────────┘                  │ wraps                         │
│                                ▼                               │
│                         ┌──────────────┐                       │
│                         │CommandRunner │                       │
│                         └──────────────┘                       │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                      CI Output Pipeline                         │
│                                                                 │
│  go test -json ──> StreamParser ──> []TestFailure              │
│                         │                  │                    │
│                         │                  │                    │
│                         ▼                  ▼                    │
│                    CIResult ──────> CIReporter                 │
│                                          │                      │
│                         ┌────────────────┼────────────────┐    │
│                         ▼                ▼                ▼    │
│                   GitHubReporter   JSONReporter    (future)    │
│                         │                │                      │
│                         ▼                ▼                      │
│                   ::error::         .jsonl file                │
│                   STEP_SUMMARY                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Configuration Integration

Extends `TestConfig` in `config.go`:

```go
// TestConfig contains test-specific settings
type TestConfig struct {
    // ... existing fields ...

    // CI Mode configuration
    CIMode CIMode `yaml:"ci_mode"`
}
```

Example `.mage.yaml`:

```yaml
test:
  timeout: "10m"
  parallel: 4
  ci_mode:
    enabled: auto          # auto/on/off
    format: github         # auto/github/json
    context_lines: 20
    max_memory_mb: 100
    dedup: true
    output_path: ".mage-x/ci-results.jsonl"
```

---

## Output File Formats

### JSON Lines (.jsonl)

Each line is a complete JSON object:

```jsonl
{"type":"start","timestamp":"2025-12-12T10:30:00Z","metadata":{"branch":"main","commit":"abc123"}}
{"type":"failure","failure":{"package":"pkg/mage","test":"TestBuild","type":"test","file":"build_test.go","line":42}}
{"type":"failure","failure":{"package":"pkg/mage","test":"TestLint","type":"panic","file":"lint_test.go","line":156}}
{"type":"summary","summary":{"status":"failed","total":1523,"passed":1521,"failed":2,"skipped":0,"duration":"2m15s"}}
```

### GitHub Annotations

```
::error file=pkg/mage/build_test.go,line=42,title=TestBuild::Expected 'linux', got 'darwin'
::error file=pkg/mage/lint_test.go,line=156,title=TestLint::panic: runtime error: invalid memory address
```

### GitHub Step Summary

Written to `$GITHUB_STEP_SUMMARY`:

```markdown
## Test Results

| Status | Count |
|--------|-------|
| ✅ Passed | 1521 |
| ❌ Failed | 2 |
| ⏭️ Skipped | 0 |

### Failed Tests

<details>
<summary>TestBuild (pkg/mage)</summary>

**File**: `build_test.go:42`
**Error**: Expected 'linux', got 'darwin'

</details>
```
