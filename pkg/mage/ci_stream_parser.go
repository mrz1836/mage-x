// Package mage provides streaming parser for go test -json output
package mage

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
)

// TestEvent represents a single event from `go test -json` output
type TestEvent struct {
	Time    string  `json:"Time"`
	Action  string  `json:"Action"`
	Package string  `json:"Package"`
	Test    string  `json:"Test"`
	Elapsed float64 `json:"Elapsed"`
	Output  string  `json:"Output"`
}

// TestEventHandler processes test events from go test -json
type TestEventHandler interface {
	// OnTestStart called when a test begins
	OnTestStart(pkg, test string)

	// OnTestPass called when a test passes
	OnTestPass(pkg, test string, elapsed float64)

	// OnTestFail called when a test fails
	OnTestFail(pkg, test string, elapsed float64, output string)

	// OnTestSkip called when a test is skipped
	OnTestSkip(pkg, test string, elapsed float64)

	// OnOutput called for each output line
	OnOutput(pkg, test, output string)

	// GetFailures returns all collected failures
	GetFailures() []CITestFailure

	// GetStats returns test statistics
	GetStats() (total, passed, failed, skipped int)
}

// StreamParser parses go test -json output line by line
type StreamParser interface {
	TestEventHandler

	// ParseLine processes a single line of JSON output
	ParseLine(line []byte) error

	// Parse processes an entire reader of JSON output
	Parse(r io.Reader) error

	// Flush finalizes parsing and returns any pending failures
	Flush() []CITestFailure

	// GetUniqueTestCount returns the number of unique tests seen
	GetUniqueTestCount() int
}

// RingBuffer implements a fixed-size circular buffer for context lines
type RingBuffer struct {
	lines []string
	size  int
	head  int
	count int
}

// NewRingBuffer creates a new ring buffer with the given capacity
func NewRingBuffer(size int) *RingBuffer {
	if size <= 0 {
		size = 20 // default
	}
	return &RingBuffer{
		lines: make([]string, size),
		size:  size,
	}
}

// Add adds a line to the buffer, overwriting oldest if full
func (rb *RingBuffer) Add(line string) {
	rb.lines[rb.head] = line
	rb.head = (rb.head + 1) % rb.size
	if rb.count < rb.size {
		rb.count++
	}
}

// GetLines returns all lines in order (oldest to newest)
func (rb *RingBuffer) GetLines() []string {
	if rb.count == 0 {
		return nil
	}

	result := make([]string, rb.count)
	start := 0
	if rb.count == rb.size {
		start = rb.head
	}

	for i := 0; i < rb.count; i++ {
		idx := (start + i) % rb.size
		result[i] = rb.lines[idx]
	}
	return result
}

// Clear resets the buffer
func (rb *RingBuffer) Clear() {
	rb.head = 0
	rb.count = 0
}

// Default limits for large suite handling
const (
	// MaxOutputSize is the maximum size of test output to capture (10MB)
	MaxOutputSize = 10 * 1024 * 1024
	// TruncationMarker is added when output is truncated
	TruncationMarker = "\n...[output truncated]...\n"
	// AdaptiveStrategyThreshold is the test count at which to re-evaluate strategy
	AdaptiveStrategyThreshold = 100
)

// streamParser implements StreamParser interface
type streamParser struct {
	contextLines  int
	strategy      CaptureStrategy
	maxOutputSize int
	mu            sync.Mutex
	failures      []CITestFailure
	signatures    map[string]bool     // for deduplication
	uniqueTests   map[uint64]struct{} // hash-based unique test tracking (memory efficient)
	testCount     atomic.Int32
	passCount     atomic.Int32
	failCount     atomic.Int32
	skipCount     atomic.Int32
	currentTest   map[string]*testState // pkg:test -> state
	dedup         bool
	adaptiveMode  bool // Whether to adapt strategy based on test count
}

// testState tracks state for a running test
type testState struct {
	pkg        string
	test       string
	output     *RingBuffer
	started    bool
	outputSize int  // Accumulated output size in bytes
	truncated  bool // Whether output was truncated
}

// NewStreamParser creates a new stream parser.
// contextLines is clamped to valid range [0, 100].
func NewStreamParser(contextLines int, dedup bool) StreamParser {
	// Validate and clamp contextLines to prevent negative values
	// and excessively large buffers
	if contextLines < 0 {
		contextLines = 0
	}
	if contextLines > 100 {
		contextLines = 100
	}
	return &streamParser{
		contextLines:  contextLines,
		strategy:      StrategySmartCapture,
		maxOutputSize: MaxOutputSize,
		failures:      make([]CITestFailure, 0),
		signatures:    make(map[string]bool),
		uniqueTests:   make(map[uint64]struct{}),
		currentTest:   make(map[string]*testState),
		dedup:         dedup,
		adaptiveMode:  true,
	}
}

// NewStreamParserWithOptions creates a stream parser with custom options.
// contextLines is clamped to valid range [0, 100].
func NewStreamParserWithOptions(opts StreamParserOptions) StreamParser {
	maxOutput := opts.MaxOutputSize
	if maxOutput <= 0 {
		maxOutput = MaxOutputSize
	}
	// Validate and clamp contextLines
	contextLines := opts.ContextLines
	if contextLines < 0 {
		contextLines = 0
	}
	if contextLines > 100 {
		contextLines = 100
	}
	return &streamParser{
		contextLines:  contextLines,
		strategy:      opts.Strategy,
		maxOutputSize: maxOutput,
		failures:      make([]CITestFailure, 0),
		signatures:    make(map[string]bool),
		uniqueTests:   make(map[uint64]struct{}),
		currentTest:   make(map[string]*testState),
		dedup:         opts.Dedup,
		adaptiveMode:  opts.AdaptiveMode,
	}
}

// hashTestKey generates a 64-bit hash for unique test tracking.
// Uses FNV-1a for fast, well-distributed hashes with minimal memory overhead.
//
// Trade-off: Using a hash instead of storing full strings reduces memory usage
// significantly for large test suites. The collision probability is acceptable:
// - 10,000 tests: ~0.00027% collision probability (birthday problem)
// - 100,000 tests: ~0.027% collision probability
//
// This is acceptable for counting unique tests but NOT for deduplication of failures,
// which uses the full string signatures map instead.
func hashTestKey(key string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(key)) // hash.Write never returns error per Go docs
	return h.Sum64()
}

// StreamParserOptions contains options for creating a StreamParser
type StreamParserOptions struct {
	ContextLines  int
	Strategy      CaptureStrategy
	MaxOutputSize int
	Dedup         bool
	AdaptiveMode  bool
}

// ParseLine processes a single line of JSON output
func (p *streamParser) ParseLine(line []byte) error {
	if len(line) == 0 {
		return nil
	}

	var event TestEvent
	if err := json.Unmarshal(line, &event); err != nil {
		// Not valid JSON - might be a non-JSON line from go test, which is expected
		// Intentionally ignore the error and continue processing
		//nolint:nilerr // Non-JSON lines from go test are expected, not errors
		return nil
	}

	p.processEvent(&event)
	return nil
}

// Parse processes an entire reader of JSON output
func (p *streamParser) Parse(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	// Increase buffer size for potentially long output lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // 1MB max line size

	for scanner.Scan() {
		if err := p.ParseLine(scanner.Bytes()); err != nil {
			return err
		}
	}
	return scanner.Err()
}

// processEvent handles a single test event
func (p *streamParser) processEvent(event *TestEvent) {
	key := event.Package + ":" + event.Test

	switch event.Action {
	case "run":
		p.OnTestStart(event.Package, event.Test)
	case "pass":
		p.OnTestPass(event.Package, event.Test, event.Elapsed)
	case "fail":
		p.OnTestFail(event.Package, event.Test, event.Elapsed, p.getTestOutput(key))
	case "skip":
		p.OnTestSkip(event.Package, event.Test, event.Elapsed)
	case "output":
		p.OnOutput(event.Package, event.Test, event.Output)
	}
}

// OnTestStart called when a test begins
func (p *streamParser) OnTestStart(pkg, test string) {
	if test == "" {
		return // Package-level event
	}

	key := pkg + ":" + test
	p.mu.Lock()
	defer p.mu.Unlock()

	// Calculate context buffer size based on strategy
	contextBufferSize := p.contextLines * 2
	switch p.strategy {
	case StrategyFullCapture:
		// Full context for small suites - use default
	case StrategySmartCapture:
		// Full context for medium suites - use default
	case StrategyEfficientCapture:
		// Limited context for large suites
		contextBufferSize = min(contextBufferSize, 10)
	case StrategyStreamingCapture:
		// Minimal context for very large suites
		contextBufferSize = min(contextBufferSize, 5)
	}

	p.currentTest[key] = &testState{
		pkg:     pkg,
		test:    test,
		output:  NewRingBuffer(contextBufferSize),
		started: true,
	}
	newCount := p.testCount.Add(1)

	// Track unique tests using hash for memory efficiency
	hash := hashTestKey(key)
	if _, exists := p.uniqueTests[hash]; !exists {
		p.uniqueTests[hash] = struct{}{}
	}

	// Adaptive strategy selection: re-evaluate at thresholds
	if p.adaptiveMode && newCount%AdaptiveStrategyThreshold == 0 {
		p.updateStrategy(int(newCount))
	}
}

// updateStrategy updates the capture strategy based on test count.
// Must be called with p.mu held or from within a locked context.
func (p *streamParser) updateStrategy(testCount int) {
	newStrategy := SelectStrategy(testCount)
	// Note: p.mu is already held by the caller (OnTestStart)
	if newStrategy != p.strategy {
		p.strategy = newStrategy
	}
}

// GetStrategy returns the current capture strategy.
// Thread-safe via mutex protection.
func (p *streamParser) GetStrategy() CaptureStrategy {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.strategy
}

// SetStrategy sets the capture strategy explicitly
func (p *streamParser) SetStrategy(strategy CaptureStrategy) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.strategy = strategy
}

// OnTestPass called when a test passes
func (p *streamParser) OnTestPass(pkg, test string, _ float64) {
	if test == "" {
		return // Package-level event
	}

	p.passCount.Add(1)

	key := pkg + ":" + test
	p.mu.Lock()
	delete(p.currentTest, key)
	p.mu.Unlock()
}

// OnTestFail called when a test fails
func (p *streamParser) OnTestFail(pkg, test string, elapsed float64, output string) {
	if test == "" {
		return // Package-level event - will be handled by individual test failures
	}

	p.failCount.Add(1)

	// Extract failure details from output
	failure := p.extractFailure(pkg, test, elapsed, output)

	// Check for deduplication
	if p.dedup && failure.Signature != "" {
		p.mu.Lock()
		if p.signatures[failure.Signature] {
			p.mu.Unlock()
			return // Already reported this failure
		}
		p.signatures[failure.Signature] = true
		p.mu.Unlock()
	}

	p.mu.Lock()
	p.failures = append(p.failures, failure)
	delete(p.currentTest, pkg+":"+test)
	p.mu.Unlock()
}

// OnTestSkip called when a test is skipped
func (p *streamParser) OnTestSkip(pkg, test string, _ float64) {
	if test == "" {
		return // Package-level event
	}

	p.skipCount.Add(1)

	key := pkg + ":" + test
	p.mu.Lock()
	delete(p.currentTest, key)
	p.mu.Unlock()
}

// OnOutput called for each output line
func (p *streamParser) OnOutput(pkg, test, output string) {
	if test == "" {
		return // Package-level output
	}

	key := pkg + ":" + test
	p.mu.Lock()
	defer p.mu.Unlock()

	if state, ok := p.currentTest[key]; ok {
		// Check if we've already exceeded the output cap
		if state.truncated {
			return // Don't add more output once truncated
		}

		outputLen := len(output)
		newSize := state.outputSize + outputLen

		// Check if adding this output would exceed the cap
		if newSize > p.maxOutputSize {
			state.truncated = true
			// Add truncation marker instead
			state.output.Add(TruncationMarker)
			return
		}

		state.outputSize = newSize
		state.output.Add(output)
	}
}

// GetFailures returns all collected failures with parent tests filtered out
func (p *streamParser) GetFailures() []CITestFailure {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Filter out parent tests to avoid duplicate reporting of nested subtests
	return filterParentTests(p.failures)
}

// filterParentTests removes parent test failures, keeping only leaf failures.
// When TestParent/Child fails, Go reports both as failed - we only want the child.
// This prevents inflated failure counts from nested subtests.
func filterParentTests(failures []CITestFailure) []CITestFailure {
	if len(failures) <= 1 {
		return failures
	}

	// Build set of all test names (per package) for quick lookup
	testNames := make(map[string]bool) // key: "pkg:test"
	for _, f := range failures {
		testNames[f.Package+":"+f.Test] = true
	}

	// Filter out parents - a test is a parent if another test starts with its name + "/"
	result := make([]CITestFailure, 0, len(failures))
	for _, f := range failures {
		key := f.Package + ":" + f.Test
		isParent := false

		// Check if any other test starts with this test + "/"
		for otherKey := range testNames {
			if otherKey != key && strings.HasPrefix(otherKey, key+"/") {
				isParent = true
				break
			}
		}

		if !isParent {
			result = append(result, f)
		}
	}

	return result
}

// GetStats returns test statistics
func (p *streamParser) GetStats() (total, passed, failed, skipped int) {
	return int(p.testCount.Load()), int(p.passCount.Load()), int(p.failCount.Load()), int(p.skipCount.Load())
}

// GetUniqueTestCount returns the number of unique tests seen.
// This counts each test only once, even if it runs multiple times
// (e.g., with different build tags).
func (p *streamParser) GetUniqueTestCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.uniqueTests)
}

// Flush finalizes parsing and returns any pending failures
func (p *streamParser) Flush() []CITestFailure {
	return p.GetFailures()
}

// getTestOutput retrieves accumulated output for a test
func (p *streamParser) getTestOutput(key string) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	if state, ok := p.currentTest[key]; ok {
		lines := state.output.GetLines()
		return strings.Join(lines, "")
	}
	return ""
}

// extractFailure extracts failure details from test output
func (p *streamParser) extractFailure(pkg, test string, elapsed float64, output string) CITestFailure {
	failure := CITestFailure{
		Package:  pkg,
		Test:     test,
		Output:   output,
		Type:     FailureTypeTest,
		Duration: formatDurationSeconds(elapsed),
	}

	// Detect failure type and extract location
	p.detectFailureType(&failure, output)
	p.extractLocation(&failure, output)

	// Capture context from source file if we have file/line information
	if failure.File != "" && failure.Line > 0 && p.contextLines > 0 {
		context, err := CaptureContext(failure.File, failure.Line, p.contextLines)
		if err == nil && len(context) > 0 {
			failure.Context = context
		}
	}

	// Generate signature for deduplication
	failure.Signature = generateSignature(pkg, test, failure.File, failure.Line, failure.Type)

	// Extract error message
	failure.Error = extractErrorMessage(output)

	return failure
}

// detectFailureType determines the type of failure from output
func (p *streamParser) detectFailureType(failure *CITestFailure, output string) {
	switch {
	// Check timeout before panic (timeout often includes "panic:" keyword)
	case timeoutPattern.MatchString(output):
		failure.Type = FailureTypeTimeout
	case racePattern.MatchString(output):
		failure.Type = FailureTypeRace
		// Extract stack trace for race conditions
		failure.Stack = extractStackTrace(output)
	case panicPattern.MatchString(output):
		failure.Type = FailureTypePanic
		if racePattern.MatchString(output) {
			failure.RaceRelated = true
		}
		// Extract stack trace
		failure.Stack = extractStackTrace(output)
	case buildErrorPattern.MatchString(output):
		failure.Type = FailureTypeBuild
	case fuzzFailPattern.MatchString(output):
		failure.Type = FailureTypeFuzz
		failure.FuzzInfo = extractFuzzInfo(output)
	default:
		failure.Type = FailureTypeTest
	}
}

// extractLocation extracts file:line from test output
func (p *streamParser) extractLocation(failure *CITestFailure, output string) {
	// Try test assertion pattern first (most common)
	if matches := testLocationPattern.FindStringSubmatch(output); len(matches) >= 3 {
		failure.File = matches[1]
		// Best effort parse - errors are intentionally ignored as line numbers are optional
		//nolint:errcheck,gosec // Best effort parsing - failure is acceptable
		fmt.Sscanf(matches[2], "%d", &failure.Line)
		return
	}

	// Try panic location pattern
	if matches := panicLocationPattern.FindStringSubmatch(output); len(matches) >= 3 {
		failure.File = matches[1]
		// Best effort parse - errors are intentionally ignored as line numbers are optional
		//nolint:errcheck,gosec // Best effort parsing - failure is acceptable
		fmt.Sscanf(matches[2], "%d", &failure.Line)
		return
	}

	// Try build error pattern
	if matches := buildLocationPattern.FindStringSubmatch(output); len(matches) >= 4 {
		failure.File = matches[1]
		// Best effort parse - errors are intentionally ignored as line/column numbers are optional
		//nolint:errcheck,gosec // Best effort parsing - failure is acceptable
		fmt.Sscanf(matches[2], "%d", &failure.Line)
		//nolint:errcheck,gosec // Best effort parsing - failure is acceptable
		fmt.Sscanf(matches[3], "%d", &failure.Column)
		return
	}

	// Try race location pattern
	if matches := raceLocationPattern.FindStringSubmatch(output); len(matches) >= 3 {
		failure.File = matches[1]
		// Best effort parse - errors are intentionally ignored as line numbers are optional
		//nolint:errcheck,gosec // Best effort parsing - failure is acceptable
		fmt.Sscanf(matches[2], "%d", &failure.Line)
		return
	}
}

// Regex patterns for failure detection
var (
	// Test assertion: filename_test.go:42: message
	testLocationPattern = regexp.MustCompile(`^\s*(\S+\.go):(\d+):`)

	// Panic: /full/path/file.go:15 +0x39
	panicLocationPattern = regexp.MustCompile(`^\s*(/?\S+\.go):(\d+)\s+\+0x`)

	// Build error: ./file.go:10:5: undefined: foo
	buildLocationPattern = regexp.MustCompile(`^\.?/?(\S+\.go):(\d+):(\d+):`)

	// Race location: similar to panic
	raceLocationPattern = regexp.MustCompile(`^\s+(\S+\.go):(\d+)\s+\+0x`)

	// Failure type patterns
	panicPattern      = regexp.MustCompile(`panic:`)
	racePattern       = regexp.MustCompile(`WARNING: DATA RACE`)
	buildErrorPattern = regexp.MustCompile(`^\./\S+\.go:\d+:\d+:`)
	fuzzFailPattern   = regexp.MustCompile(`Failing input written to`)
	timeoutPattern    = regexp.MustCompile(`test timed out after|panic: test timed out`)

	// Fuzz corpus path: Failing input written to testdata/fuzz/FuzzFoo/582528ddfad69eb5
	corpusPathPattern = regexp.MustCompile(`Failing input written to (\S+)`)
)

// generateSignature creates a deduplication key
func generateSignature(pkg, test, file string, line int, failureType FailureType) string {
	data := fmt.Sprintf("%s:%s:%s:%d:%s", pkg, test, file, line, failureType)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8]) // Use first 8 bytes for compact signature
}

// extractStackTrace extracts stack trace from panic output
func extractStackTrace(output string) string {
	lines := strings.Split(output, "\n")
	var stack []string
	inStack := false

	for _, line := range lines {
		if strings.HasPrefix(line, "goroutine ") {
			inStack = true
		}
		if inStack {
			stack = append(stack, line)
			// Stop at empty line or next panic marker
			if line == "" || (len(stack) > 1 && strings.HasPrefix(line, "panic:")) {
				break
			}
		}
	}

	return strings.Join(stack, "\n")
}

// extractFuzzInfo extracts fuzz test specific information
func extractFuzzInfo(output string) *FuzzInfo {
	info := &FuzzInfo{}

	// Extract corpus path using pre-compiled regex (see corpusPathPattern at package level)
	if matches := corpusPathPattern.FindStringSubmatch(output); len(matches) >= 2 {
		info.CorpusPath = matches[1]
	}

	// Check if it's a seed corpus failure
	if strings.Contains(output, "/seed#") {
		info.FromSeed = true
	}

	// Try to read the failing input if corpus path exists
	if info.CorpusPath != "" {
		if data, err := os.ReadFile(info.CorpusPath); err == nil {
			// Check if binary (contains non-printable characters)
			isBinary := false
			for _, b := range data {
				if b < 32 && b != '\n' && b != '\r' && b != '\t' {
					isBinary = true
					break
				}
			}
			info.IsBinary = isBinary
			if !isBinary {
				info.Input = string(data)
			}
		}
	}

	return info
}

// extractErrorMessage extracts a concise error message from output
func extractErrorMessage(output string) string {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for assertion messages
		if strings.Contains(line, "Error:") ||
			strings.Contains(line, "error:") ||
			strings.Contains(line, "FAIL:") ||
			strings.Contains(line, "panic:") {
			return line
		}
		// Look for test assertion format: file.go:42: message
		if testLocationPattern.MatchString(line) {
			// Extract just the message part after file:line:
			parts := strings.SplitN(line, ":", 3)
			if len(parts) >= 3 {
				msg := strings.TrimSpace(parts[2])
				if msg != "" {
					return msg
				}
				// Fall through to continue searching if message is empty
			}
		}
	}

	// Return first non-empty line as fallback
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "---") && !strings.HasPrefix(line, "===") {
			if len(line) > 200 {
				return line[:200] + "..."
			}
			return line
		}
	}

	return "test failed"
}

// formatDurationSeconds formats elapsed seconds as duration string
func formatDurationSeconds(seconds float64) string {
	if seconds < 0.001 {
		return "<1ms"
	}
	if seconds < 1 {
		return fmt.Sprintf("%.0fms", seconds*1000)
	}
	return fmt.Sprintf("%.2fs", seconds)
}

// CaptureContext reads surrounding lines from a source file
func CaptureContext(file string, line, contextLines int) ([]string, error) {
	if file == "" || line <= 0 || contextLines <= 0 {
		return nil, nil
	}

	f, err := os.Open(file) //nolint:gosec // Source file from test output
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := f.Close(); err != nil {
			// Log but don't fail - we're reading for context only
			fmt.Fprintf(os.Stderr, "Warning: failed to close file %s: %v\n", file, err)
		}
	}()

	scanner := bufio.NewScanner(f)
	var lines []string
	lineNum := 0

	startLine := line - contextLines
	if startLine < 1 {
		startLine = 1
	}
	endLine := line + contextLines

	for scanner.Scan() {
		lineNum++
		if lineNum >= startLine && lineNum <= endLine {
			prefix := "  "
			if lineNum == line {
				prefix = "> "
			}
			lines = append(lines, fmt.Sprintf("%s%4d | %s", prefix, lineNum, scanner.Text()))
		}
		if lineNum > endLine {
			break
		}
	}

	return lines, scanner.Err()
}
