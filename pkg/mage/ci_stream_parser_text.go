// Package mage provides text-mode streaming parser for fuzz test output
//
// Fuzz tests cannot use `go test -json` flag, so this parser handles the plain text
// output format. It detects failure patterns like `--- FAIL: FuzzTestName (0.34s)`
// and extracts error details, producing the same structured output as the JSON parser.
package mage

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// TextStreamParser parses plain text test output (for fuzz tests)
type TextStreamParser interface {
	// ParseLine processes a single line of text output
	ParseLine(line string) error

	// Parse processes an entire reader of text output
	Parse(r io.Reader) error

	// Flush finalizes parsing and returns collected failures
	Flush() []CITestFailure

	// GetFailures returns all collected failures
	GetFailures() []CITestFailure

	// GetStats returns test statistics
	GetStats() (total, passed, failed, skipped int)
}

// textStreamParser implements TextStreamParser interface
type textStreamParser struct {
	mu           sync.Mutex
	contextLines int
	dedup        bool

	// Test tracking
	failures    []CITestFailure
	seenTests   map[string]*CITestFailure // test name -> failure (for dedup by elapsed time)
	currentTest *textTestState            // Current test being processed

	// Statistics
	testCount int
	passCount int
	failCount int
}

// textTestState tracks state for a running test in text mode
type textTestState struct {
	pkg           string
	test          string
	elapsed       float64 // seconds
	outputBuffer  *RingBuffer
	failureType   FailureType
	inFailure     bool // We've seen a FAIL line for this test
	errorMessage  string
	file          string
	line          int
	fuzzCorpus    string
	panicDetected bool
}

// Regex patterns for text parsing
var (
	// --- FAIL: FuzzTestName (0.34s)
	failLinePattern = regexp.MustCompile(`^---\s*FAIL:\s+([A-Za-z][A-Za-z0-9_]*)\s*\(([0-9.]+)([a-z]*)\)`)

	// --- PASS: TestName (0.01s)
	passLinePattern = regexp.MustCompile(`^---\s*PASS:\s+([A-Za-z][A-Za-z0-9_]*)\s*\(([0-9.]+)([a-z]*)\)`)

	// FAIL package/path 0.123s
	failPackagePattern = regexp.MustCompile(`^FAIL\s+(\S+)\s+([0-9.]+)s`)

	// ok package/path 0.123s
	okPackagePattern = regexp.MustCompile(`^ok\s+(\S+)\s+([0-9.]+)s`)

	// Failing input written to testdata/fuzz/FuzzFoo/582528ddfad69eb5
	fuzzCorpusPattern = regexp.MustCompile(`Failing input written to (\S+)`)

	// panic: message
	textPanicPattern = regexp.MustCompile(`^panic:\s*(.+)`)

	// fatal error: message
	textFatalPattern = regexp.MustCompile(`^fatal error:\s*(.+)`)

	// runtime error: message
	textRuntimePattern = regexp.MustCompile(`^runtime error:\s*(.+)`)

	// WARNING: DATA RACE
	textRacePattern = regexp.MustCompile(`WARNING:\s*DATA\s*RACE`)

	// file.go:42: error message
	textLocationPattern = regexp.MustCompile(`^\s*(\S+\.go):(\d+):\s*(.*)`)

	// /path/to/file.go:42 +0x39 (panic stack trace)
	textStackLocationPattern = regexp.MustCompile(`^\s*(/?\S+\.go):(\d+)\s+\+0x`)
)

// NewTextStreamParser creates a new text stream parser
func NewTextStreamParser(contextLines int, dedup bool) TextStreamParser {
	if contextLines <= 0 {
		contextLines = 20
	}
	return &textStreamParser{
		contextLines: contextLines,
		dedup:        dedup,
		failures:     make([]CITestFailure, 0),
		seenTests:    make(map[string]*CITestFailure),
	}
}

// ParseLine processes a single line of text output
func (p *textStreamParser) ParseLine(line string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Trim trailing whitespace but preserve leading (for indentation detection)
	line = strings.TrimRight(line, "\r\n")

	// Check for test failure line: --- FAIL: TestName (0.34s)
	if matches := failLinePattern.FindStringSubmatch(line); len(matches) >= 4 {
		p.handleFailLine(matches[1], matches[2], matches[3])
		return nil
	}

	// Check for test pass line: --- PASS: TestName (0.01s)
	if matches := passLinePattern.FindStringSubmatch(line); len(matches) >= 3 {
		p.handlePassLine(matches[1])
		return nil
	}

	// Check for package-level failure: FAIL package/path
	if matches := failPackagePattern.FindStringSubmatch(line); len(matches) >= 2 {
		p.handlePackageResult(matches[1], false)
		return nil
	}

	// Check for package-level pass: ok package/path
	if matches := okPackagePattern.FindStringSubmatch(line); len(matches) >= 2 {
		p.handlePackageResult(matches[1], true)
		return nil
	}

	// If we're tracking a failed test, capture output
	if p.currentTest != nil && p.currentTest.inFailure {
		p.captureTestOutput(line)
	}

	return nil
}

// handleFailLine processes a test failure line
func (p *textStreamParser) handleFailLine(testName, elapsedStr, unit string) {
	// Finalize previous test if any
	p.finalizeCurrentTest()

	elapsed := parseElapsed(elapsedStr, unit)

	// Check if we've seen this test before (for deduplication)
	// We always create the new test state - deduplication happens in dedupFailures()
	// by keeping the entry with the longest elapsed time (parent test)
	if existing, ok := p.seenTests[testName]; ok {
		// If this one has longer elapsed time, it will replace existing in dedupFailures()
		if elapsed <= existing.elapsedSeconds() {
			// Skip this one (it's a child/nested test with shorter duration)
			return
		}
	}

	// New or replacing test failure
	p.currentTest = &textTestState{
		test:         testName,
		elapsed:      elapsed,
		outputBuffer: NewRingBuffer(p.contextLines * 2),
		failureType:  FailureTypeFuzz, // Default to fuzz for text parser
		inFailure:    true,
	}

	p.failCount++
	p.testCount++
}

// handlePassLine processes a test pass line
func (p *textStreamParser) handlePassLine(testName string) {
	// Finalize previous test if any
	p.finalizeCurrentTest()

	// If this test was previously marked as failed, it might be a retry that passed
	// In text mode, we just count it as passed
	if _, ok := p.seenTests[testName]; !ok {
		p.passCount++
		p.testCount++
	}
}

// handlePackageResult processes package-level result line
func (p *textStreamParser) handlePackageResult(pkg string, _ bool) {
	// Finalize any current test
	p.finalizeCurrentTest()

	// Update package info on any pending failures without a package
	for i := range p.failures {
		if p.failures[i].Package == "" {
			p.failures[i].Package = pkg
		}
	}

	// Update current test package if present
	if p.currentTest != nil && p.currentTest.pkg == "" {
		p.currentTest.pkg = pkg
	}
}

// captureTestOutput captures output lines for the current failing test
func (p *textStreamParser) captureTestOutput(line string) {
	if p.currentTest == nil {
		return
	}

	// Add to output buffer
	p.currentTest.outputBuffer.Add(line)

	// Check for panic
	if matches := textPanicPattern.FindStringSubmatch(line); len(matches) >= 2 {
		p.currentTest.panicDetected = true
		p.currentTest.failureType = FailureTypePanic
		if p.currentTest.errorMessage == "" {
			p.currentTest.errorMessage = matches[1]
		}
	}

	// Check for fatal error
	if matches := textFatalPattern.FindStringSubmatch(line); len(matches) >= 2 {
		p.currentTest.failureType = FailureTypeFatal
		if p.currentTest.errorMessage == "" {
			p.currentTest.errorMessage = matches[1]
		}
	}

	// Check for runtime error
	if matches := textRuntimePattern.FindStringSubmatch(line); len(matches) >= 2 {
		p.currentTest.failureType = FailureTypePanic
		if p.currentTest.errorMessage == "" {
			p.currentTest.errorMessage = "runtime error: " + matches[1]
		}
	}

	// Check for data race
	if textRacePattern.MatchString(line) {
		p.currentTest.failureType = FailureTypeRace
		if p.currentTest.errorMessage == "" {
			p.currentTest.errorMessage = "data race detected"
		}
	}

	// Check for fuzz corpus path
	if matches := fuzzCorpusPattern.FindStringSubmatch(line); len(matches) >= 2 {
		p.currentTest.fuzzCorpus = matches[1]
		p.currentTest.failureType = FailureTypeFuzz
	}

	// Try to extract file:line location
	p.extractFileLocation(line)
}

// extractFileLocation tries to extract file:line information from output line
func (p *textStreamParser) extractFileLocation(line string) {
	if p.currentTest == nil || p.currentTest.file != "" {
		return
	}

	// Try standard file.go:42: message format
	if matches := textLocationPattern.FindStringSubmatch(line); len(matches) >= 4 {
		p.currentTest.file = matches[1]
		if lineNum, err := strconv.Atoi(matches[2]); err == nil {
			p.currentTest.line = lineNum
		}
		if p.currentTest.errorMessage == "" && matches[3] != "" {
			p.currentTest.errorMessage = strings.TrimSpace(matches[3])
		}
		return
	}

	// Try stack trace format: /path/to/file.go:42 +0x39
	if matches := textStackLocationPattern.FindStringSubmatch(line); len(matches) >= 3 {
		p.currentTest.file = matches[1]
		if lineNum, err := strconv.Atoi(matches[2]); err == nil {
			p.currentTest.line = lineNum
		}
	}
}

// finalizeCurrentTest converts the current test state to a failure and stores it
func (p *textStreamParser) finalizeCurrentTest() {
	if p.currentTest == nil || !p.currentTest.inFailure {
		return
	}

	// Build output string
	outputLines := p.currentTest.outputBuffer.GetLines()
	output := strings.Join(outputLines, "\n")

	// Create failure entry
	failure := CITestFailure{
		Package:  p.currentTest.pkg,
		Test:     p.currentTest.test,
		Type:     p.currentTest.failureType,
		Error:    p.currentTest.errorMessage,
		Output:   output,
		File:     p.currentTest.file,
		Line:     p.currentTest.line,
		Duration: formatDurationSeconds(p.currentTest.elapsed),
	}

	// Add fuzz info if applicable
	if p.currentTest.fuzzCorpus != "" {
		failure.FuzzInfo = &FuzzInfo{
			CorpusPath: p.currentTest.fuzzCorpus,
		}
	}

	// Generate signature for deduplication
	failure.Signature = generateSignature(
		failure.Package,
		failure.Test,
		failure.File,
		failure.Line,
		failure.Type,
	)

	// Capture context from source file if available
	if failure.File != "" && failure.Line > 0 && p.contextLines > 0 {
		if context, err := CaptureContext(failure.File, failure.Line, p.contextLines); err == nil && len(context) > 0 {
			failure.Context = context
		}
	}

	// Extract stack trace for panic/fatal types
	if p.currentTest.panicDetected || p.currentTest.failureType == FailureTypeFatal {
		failure.Stack = extractStackTrace(output)
	}

	// Store for deduplication tracking
	p.seenTests[p.currentTest.test] = &failure

	// Add to failures list
	p.failures = append(p.failures, failure)

	// Reset current test
	p.currentTest = nil
}

// Parse processes an entire reader of text output
func (p *textStreamParser) Parse(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	// Increase buffer size for potentially long output lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // 1MB max line size

	for scanner.Scan() {
		if err := p.ParseLine(scanner.Text()); err != nil {
			return fmt.Errorf("failed to parse line: %w", err)
		}
	}

	// Finalize any remaining test
	p.mu.Lock()
	p.finalizeCurrentTest()
	p.mu.Unlock()

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to scan input: %w", err)
	}
	return nil
}

// Flush finalizes parsing and returns collected failures
func (p *textStreamParser) Flush() []CITestFailure {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Finalize any remaining test
	p.finalizeCurrentTest()

	return p.dedupFailures()
}

// GetFailures returns all collected failures with deduplication applied
func (p *textStreamParser) GetFailures() []CITestFailure {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.dedupFailures()
}

// dedupFailures returns failures with duplicates removed
// For tests with same name, keeps the one with longest elapsed time (parent test)
func (p *textStreamParser) dedupFailures() []CITestFailure {
	if !p.dedup || len(p.failures) <= 1 {
		return p.failures
	}

	// Build map by test name, keeping longest elapsed
	bestByTest := make(map[string]CITestFailure)
	for _, f := range p.failures {
		existing, ok := bestByTest[f.Test]
		if !ok {
			bestByTest[f.Test] = f
			continue
		}

		// Keep the one with longer duration (parent test)
		if f.elapsedSeconds() > existing.elapsedSeconds() {
			bestByTest[f.Test] = f
		}
	}

	// Convert back to slice
	result := make([]CITestFailure, 0, len(bestByTest))
	for _, f := range bestByTest {
		result = append(result, f)
	}

	return result
}

// GetStats returns test statistics
func (p *textStreamParser) GetStats() (total, passed, failed, skipped int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.testCount, p.passCount, p.failCount, 0 // text mode doesn't track skipped
}

// parseElapsed converts elapsed string and unit to seconds
func parseElapsed(elapsedStr, unit string) float64 {
	elapsed, err := strconv.ParseFloat(elapsedStr, 64)
	if err != nil {
		return 0
	}

	// Convert based on unit
	switch unit {
	case "ms":
		return elapsed / 1000
	case "s", "":
		return elapsed
	case "m":
		return elapsed * 60
	default:
		return elapsed
	}
}

// elapsedSeconds returns elapsed time in seconds from duration string
func (f *CITestFailure) elapsedSeconds() float64 {
	if f.Duration == "" {
		return 0
	}

	// Try to parse as duration
	if d, err := time.ParseDuration(f.Duration); err == nil {
		return d.Seconds()
	}

	// Try parsing just the number (assuming seconds)
	s := strings.TrimSuffix(f.Duration, "s")
	s = strings.TrimSuffix(s, "ms")
	if parsed, err := strconv.ParseFloat(s, 64); err == nil {
		return parsed
	}

	return 0
}
