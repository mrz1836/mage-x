// Package mage provides CI runner for test output interception
package mage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// ErrCIOutputProcessingPanic is returned when a panic occurs during CI output processing
var ErrCIOutputProcessingPanic = errors.New("panic during CI output processing")

// maxStderrSize limits stderr buffer to 10MB to prevent OOM from massive stack dumps
const maxStderrSize = 10 * 1024 * 1024

// limitedBuffer is a size-bounded buffer that prevents unbounded memory growth.
// When the limit is exceeded, additional writes are silently discarded and the
// truncated flag is set.
type limitedBuffer struct {
	buf       bytes.Buffer
	limit     int
	truncated bool
}

// newLimitedBuffer creates a new size-limited buffer
func newLimitedBuffer(limit int) *limitedBuffer {
	return &limitedBuffer{limit: limit}
}

// Write implements io.Writer with size limiting
func (lb *limitedBuffer) Write(p []byte) (n int, err error) {
	if lb.buf.Len() >= lb.limit {
		lb.truncated = true
		return len(p), nil // Report full write to avoid broken pipe
	}
	remaining := lb.limit - lb.buf.Len()
	if len(p) > remaining {
		lb.truncated = true
		_, err = lb.buf.Write(p[:remaining])
		return len(p), err // Report full write
	}
	return lb.buf.Write(p)
}

// String returns the buffer contents as a string
func (lb *limitedBuffer) String() string {
	s := lb.buf.String()
	if lb.truncated {
		s += "\n...[stderr truncated - exceeded " + fmt.Sprintf("%dMB", lb.limit/(1024*1024)) + " limit]..."
	}
	return s
}

// Len returns the current buffer size
func (lb *limitedBuffer) Len() int {
	return lb.buf.Len()
}

// CIRunner wraps CommandRunner to intercept test output for CI mode
type CIRunner interface {
	CommandRunner

	// WithContext returns a context-aware runner
	WithContext(ctx context.Context) CIRunner

	// GetResults returns the collected CI results
	GetResults() *CIResult

	// GenerateReport writes the final report
	GenerateReport() error
}

// CIRunnerOptions configures CI runner behavior
type CIRunnerOptions struct {
	Mode     CIMode
	Reporter CIReporter
	Detector CIDetector
}

// ciRunner implements CIRunner interface
type ciRunner struct {
	base      CommandRunner
	mode      CIMode
	parser    StreamParser
	reporter  CIReporter
	detector  CIDetector
	results   *CIResult
	startTime time.Time
	mu        sync.Mutex
	getCtx    func() context.Context // Function to retrieve context instead of storing it
	started   bool                   // Track if reporter has been started (prevents duplicate start entries)
}

// NewCIRunner wraps a CommandRunner with CI capabilities
func NewCIRunner(base CommandRunner, opts CIRunnerOptions) CIRunner {
	// Use defaults if not provided
	if opts.Detector == nil {
		opts.Detector = NewCIDetector()
	}

	return &ciRunner{
		base:     base,
		mode:     opts.Mode,
		parser:   NewStreamParser(opts.Mode.ContextLines, opts.Mode.Dedup),
		reporter: opts.Reporter,
		detector: opts.Detector,
		results:  &CIResult{},
		getCtx:   func() context.Context { return context.Background() },
	}
}

// WithContext returns a context-aware runner.
// Each runner created via WithContext has its own parser and results
// to prevent race conditions when multiple runners execute concurrently.
// The reporter is shared (it has its own mutex for thread safety).
func (r *ciRunner) WithContext(ctx context.Context) CIRunner {
	return &ciRunner{
		base:     r.base,
		mode:     r.mode,
		parser:   NewStreamParser(r.mode.ContextLines, r.mode.Dedup), // Fresh parser to avoid race
		reporter: r.reporter,                                         // Reporter has its own mutex - safe to share
		detector: r.detector,
		results:  &CIResult{}, // Fresh results to avoid race
		getCtx:   func() context.Context { return ctx },
	}
}

// RunCmd executes a command, intercepting go test -json output
func (r *ciRunner) RunCmd(name string, args ...string) error {
	// Only intercept go test commands when CI mode is enabled
	if r.mode.Enabled && name == "go" && len(args) > 0 && args[0] == "test" {
		ctx := context.Background()
		if r.getCtx != nil {
			ctx = r.getCtx()
		}
		return r.runTestWithCI(ctx, name, args...)
	}

	// Pass through to base runner
	return r.base.RunCmd(name, args...)
}

// RunCmdOutput executes a command and returns output
func (r *ciRunner) RunCmdOutput(name string, args ...string) (string, error) {
	// For output capture, just pass through
	return r.base.RunCmdOutput(name, args...)
}

// runTestWithCI runs go test with CI mode output processing
func (r *ciRunner) runTestWithCI(ctx context.Context, name string, args ...string) error {
	r.startTime = time.Now()

	// Print CI mode startup banner
	printCIModeBanner(r.detector.Platform(), r.mode)

	// Start the reporter (only once - prevents duplicate start entries when tests run multiple times)
	// Use mutex to avoid race condition on the started flag
	r.mu.Lock()
	shouldStart := r.reporter != nil && !r.started
	if shouldStart {
		r.started = true
	}
	r.mu.Unlock()

	if shouldStart {
		if d, ok := r.detector.(*ciDetector); ok {
			metadata := d.GetMetadata()
			if err := r.reporter.Start(metadata); err != nil {
				// Log but don't fail - continue with test
				fmt.Fprintf(os.Stderr, "Warning: Failed to start CI reporter: %v\n", err)
			}
		}
	}

	// Ensure -json flag is present for output parsing
	hasJSON := false
	for _, arg := range args {
		if arg == "-json" {
			hasJSON = true
			break
		}
	}
	if !hasJSON {
		// Create a new slice to avoid modifying the caller's slice
		// The original code could corrupt the caller's data if the slice had spare capacity
		newArgs := make([]string, 0, len(args)+1)
		newArgs = append(newArgs, args[0], "-json")
		newArgs = append(newArgs, args[1:]...)
		args = newArgs
	}

	// Create command
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = os.Environ()

	// Create pipes for stdout
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	// Track if we need to clean up the pipe on early error
	pipeNeedsClose := true
	defer func() {
		// Close pipe only if Start() failed - otherwise Wait() handles it
		if pipeNeedsClose && stdout != nil {
			_ = stdout.Close() //nolint:errcheck // Best effort cleanup, error doesn't affect logic
		}
	}()

	// Capture stderr with bounded buffer to prevent OOM from massive stack dumps
	stderrBuf := newLimitedBuffer(maxStderrSize)
	cmd.Stderr = stderrBuf

	// Start command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command: %w", err)
	}
	pipeNeedsClose = false // Wait() will handle pipe cleanup

	// Process output in real-time with tee to stdout
	if err := r.processOutput(stdout); err != nil {
		// Log but continue - don't fail due to parsing errors
		fmt.Fprintf(os.Stderr, "Warning: Error processing test output: %v\n", err)
	}

	// Wait for command to complete
	cmdErr := cmd.Wait()

	// Collect results
	r.collectResults()

	// Note: Failures are reported in GenerateReport() to avoid duplicates
	// when tests run multiple times (e.g., with different build tags)

	// Handle stderr for crashes
	if stderrBuf.Len() > 0 && cmdErr != nil {
		r.handleCrash(stderrBuf.String())
	}

	return cmdErr
}

// processOutput reads and parses test output, teeing to stdout for terminal display
func (r *ciRunner) processOutput(stdout io.Reader) (retErr error) {
	// Recovery handler for panics in parsing - ensures test output still displays
	// and prevents deadlock from undrained pipe
	defer func() {
		if p := recover(); p != nil {
			// CRITICAL: Drain remaining pipe data to prevent subprocess deadlock
			// The subprocess may block on write if the pipe buffer is full
			_, _ = io.Copy(io.Discard, stdout) //nolint:errcheck // Best effort drain, error doesn't affect recovery
			retErr = fmt.Errorf("%w: %v", ErrCIOutputProcessingPanic, p)
		}
	}()

	// Tee output to both parser and stdout for backwards compatibility
	tee := io.TeeReader(stdout, os.Stdout)

	return r.parser.Parse(tee)
}

// collectResults gathers test results from parser
func (r *ciRunner) collectResults() {
	r.mu.Lock()
	defer r.mu.Unlock()

	total, passed, failed, skipped := r.parser.GetStats()
	uniqueTotal := r.parser.GetUniqueTestCount()
	failures := r.parser.GetFailures()

	// Determine status
	status := TestStatusPassed
	if failed > 0 {
		status = TestStatusFailed
	}

	duration := time.Since(r.startTime)

	r.results = &CIResult{
		Summary: CISummary{
			Status:      status,
			Total:       total,
			UniqueTotal: uniqueTotal,
			Passed:      passed,
			Failed:      len(failures), // Use deduplicated count, not raw count which includes parent tests
			Skipped:     skipped,
			Duration:    formatDurationForSummary(duration),
		},
		Failures:  failures,
		Timestamp: r.startTime,
		Duration:  duration,
	}

	// Add metadata if detector available
	if d, ok := r.detector.(*ciDetector); ok {
		r.results.Metadata = d.GetMetadata()
	}
}

// handleCrash handles test binary crashes
func (r *ciRunner) handleCrash(stderr string) {
	// Look for crash indicators
	if strings.Contains(stderr, "SIGSEGV") ||
		strings.Contains(stderr, "fatal error:") ||
		strings.Contains(stderr, "unexpected signal") ||
		strings.Contains(stderr, "signal:") ||
		strings.Contains(stderr, "panic:") {

		failure := CITestFailure{
			Type:   FailureTypeFatal,
			Error:  "Test binary crashed",
			Output: stderr,
			Stack:  stderr, // Include full crash dump as stack
		}

		// Try to extract location from crash dump
		extractCrashLocation(&failure, stderr)

		r.mu.Lock()
		r.results.Failures = append(r.results.Failures, failure)
		r.results.Summary.Status = TestStatusError
		r.mu.Unlock()

		if r.reporter != nil {
			if err := r.reporter.ReportFailure(failure); err != nil {
				// Log but continue - already in error handling path
				// Error writing to stderr is non-critical in error path
				fmt.Fprintf(os.Stderr, "Warning: Failed to report crash: %v\n", err)
			}
		}
	}
}

// extractCrashLocation tries to extract file:line from crash dump
func extractCrashLocation(failure *CITestFailure, stderr string) {
	// Look for goroutine stack trace format
	// Stack traces look like:
	// goroutine 1 [running]:
	// main.foo()
	//     /path/to/file.go:42 +0x39
	lines := strings.Split(stderr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for the typical stack trace location format: /path/file.go:42 +0x...
		if panicLocationPattern.MatchString(line) {
			matches := panicLocationPattern.FindStringSubmatch(line)
			if len(matches) >= 3 {
				failure.File = matches[1]
				// Best effort line number parse - errors are intentionally ignored as line numbers are optional
				//nolint:errcheck,gosec // Best effort parsing - failure is acceptable
				fmt.Sscanf(matches[2], "%d", &failure.Line)
				return
			}
		}
	}
}

// GetResults returns the collected CI results
func (r *ciRunner) GetResults() *CIResult {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.results
}

// GenerateReport writes the final report
func (r *ciRunner) GenerateReport() error {
	if r.reporter == nil {
		return nil
	}

	// Create a defensive copy of failures under lock to avoid race condition
	// with handleCrash which may append to results.Failures concurrently
	r.mu.Lock()
	results := r.results
	failures := make([]CITestFailure, len(results.Failures))
	copy(failures, results.Failures)
	r.mu.Unlock()

	// Print CI mode completion summary to stdout
	printCIModeSummary(results, r.mode.OutputPath)

	// Report failures to JSONL (done here to ensure each failure is written only once,
	// even when tests run multiple times with different build tags)
	for _, failure := range failures {
		if err := r.reporter.ReportFailure(failure); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to report failure: %v\n", err)
		}
	}

	if err := r.reporter.WriteSummary(results); err != nil {
		return fmt.Errorf("failed to write summary: %w", err)
	}

	return r.reporter.Close()
}

// formatDurationForSummary formats duration for human-readable summary
func formatDurationForSummary(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
}

// printCIModeBanner outputs CI mode configuration to stdout
func printCIModeBanner(platform string, mode CIMode) {
	dedupStatus := "disabled"
	if mode.Dedup {
		dedupStatus = "enabled"
	}

	lines := []string{
		"",
		"============================================================",
		"🤖 MAGE-X CI MODE ACTIVE",
		"============================================================",
		fmt.Sprintf("📍 Platform:       %s", platform),
		fmt.Sprintf("📋 Format:         %s", mode.Format),
		fmt.Sprintf("📊 Output File:    %s", mode.OutputPath),
		fmt.Sprintf("🔍 Context Lines:  %d", mode.ContextLines),
		fmt.Sprintf("💾 Max Memory:     %dMB", mode.MaxMemoryMB),
		fmt.Sprintf("🔄 Deduplication:  %s", dedupStatus),
		"============================================================",
		"",
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(os.Stdout, line); err != nil {
			// Log to stderr so users know the banner was truncated
			fmt.Fprintf(os.Stderr, "Warning: failed to write CI banner: %v\n", err)
			return
		}
	}
}

// printCIModeSummary outputs test results summary to stdout
func printCIModeSummary(results *CIResult, outputPath string) {
	lines := []string{
		"",
		"============================================================",
		"📊 MAGE-X CI TEST SUMMARY",
		"============================================================",
		fmt.Sprintf("Total Tests:       %d", results.Summary.Total),
		fmt.Sprintf("├── Passed:        %d", results.Summary.Passed),
		fmt.Sprintf("├── Failed:        %d", results.Summary.Failed),
		fmt.Sprintf("└── Skipped:       %d", results.Summary.Skipped),
		fmt.Sprintf("Duration:          %s", results.Summary.Duration),
		fmt.Sprintf("Failures Detected: %d", len(results.Failures)),
	}
	for _, f := range results.Failures {
		lines = append(lines, fmt.Sprintf("  └── %s (%s) - %s", f.Test, f.Package, f.Type))
	}
	lines = append(lines,
		fmt.Sprintf("Output Written:    %s", outputPath),
		"============================================================",
		"",
	)
	for _, line := range lines {
		if _, err := fmt.Fprintln(os.Stdout, line); err != nil {
			// Log to stderr so users know the summary was truncated
			fmt.Fprintf(os.Stderr, "Warning: failed to write CI summary: %v\n", err)
			return
		}
	}
}

// printCIFuzzSummary outputs fuzz test results summary to stdout
func printCIFuzzSummary(results []fuzzTestResult, totalDuration time.Duration) {
	total := len(results)
	passed := 0
	failed := 0
	var failures []fuzzTestResult

	for _, r := range results {
		if r.Error == nil {
			passed++
		} else {
			failed++
			failures = append(failures, r)
		}
	}

	lines := []string{
		"",
		"============================================================",
		"📊 MAGE-X CI FUZZ TEST SUMMARY",
		"============================================================",
		fmt.Sprintf("Total Fuzz Tests:  %d", total),
		fmt.Sprintf("├── Passed:        %d", passed),
		fmt.Sprintf("└── Failed:        %d", failed),
		fmt.Sprintf("Duration:          %s", formatDurationForSummary(totalDuration)),
		fmt.Sprintf("Failures Detected: %d", len(failures)),
	}
	for _, f := range failures {
		lines = append(lines, fmt.Sprintf("  └── %s (%s)", f.Test, f.Package))
	}
	lines = append(lines, "============================================================", "")

	for _, line := range lines {
		if _, err := fmt.Fprintln(os.Stdout, line); err != nil {
			// Log to stderr so users know the fuzz summary was truncated
			fmt.Fprintf(os.Stderr, "Warning: failed to write CI fuzz summary: %v\n", err)
			return
		}
	}
}

// IsCIEnabled checks if CI mode should be enabled
func IsCIEnabled(params map[string]string, cfg *Config) bool {
	detector := NewCIDetector()
	mode := detector.GetConfig(params, cfg)
	return mode.Enabled
}

// GetCIRunner returns a CI runner if CI mode is enabled, otherwise returns the base runner
func GetCIRunner(base CommandRunner, params map[string]string, cfg *Config) CommandRunner {
	detector := NewCIDetector()
	mode := detector.GetConfig(params, cfg)

	if !mode.Enabled {
		return base
	}

	// Create appropriate reporters based on mode
	var reporters []CIReporter

	// Always add JSON reporter for structured output
	if mode.OutputPath != "" {
		jsonReporter, err := NewJSONReporter(mode.OutputPath)
		if err == nil {
			reporters = append(reporters, jsonReporter)
		}
	}

	// Add platform-specific reporter
	platform := detector.Platform()
	if mode.Format == CIFormatGitHub || (mode.Format == CIFormatAuto && platform == "github") {
		reporters = append(reporters, NewGitHubReporter())
	} else if platform == "local" {
		// Add terminal reporter for local CI mode preview
		reporters = append(reporters, NewTerminalReporter())
	}

	// Combine reporters
	var reporter CIReporter
	switch len(reporters) {
	case 0:
		reporter = NullReporter{}
	case 1:
		reporter = reporters[0]
	default:
		reporter = NewMultiReporter(reporters...)
	}

	return NewCIRunner(base, CIRunnerOptions{
		Mode:     mode,
		Reporter: reporter,
		Detector: detector,
	})
}

// PrintCIBannerIfEnabled prints CI mode banner if CI is enabled.
// This is useful for test functions that don't use the full CI runner
// (e.g., fuzz tests which have non-JSON output that can't be parsed).
func PrintCIBannerIfEnabled(params map[string]string, cfg *Config) {
	detector := NewCIDetector()
	mode := detector.GetConfig(params, cfg)

	if mode.Enabled {
		printCIModeBanner(detector.Platform(), mode)
	}
}
