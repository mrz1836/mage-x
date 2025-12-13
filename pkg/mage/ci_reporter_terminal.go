// Package mage provides terminal reporter for local CI mode preview
package mage

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// terminalReporter provides terminal-friendly output for local CI mode
type terminalReporter struct {
	output   io.Writer
	failures []CITestFailure
	started  bool
}

// write is a helper that writes to output and logs errors to stderr
// Terminal output errors are non-critical and shouldn't fail the test run
func (r *terminalReporter) write(format string, args ...interface{}) {
	if _, err := fmt.Fprintf(r.output, format, args...); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to write to terminal: %v\n", err)
	}
}

// writeFailure writes a single test failure with formatting
func (r *terminalReporter) writeFailure(num int, f CITestFailure) {
	r.write("%d. %s\n", num, colorRed(f.Test))
	if f.Package != "" {
		r.write("   Package: %s\n", f.Package)
	}
	if f.File != "" {
		location := f.File
		if f.Line > 0 {
			location = fmt.Sprintf("%s:%d", f.File, f.Line)
		}
		r.write("   Location: %s\n", colorCyan(location))
	}
	if f.Type != FailureTypeTest {
		r.write("   Type: %s\n", colorYellow(string(f.Type)))
	}
	if f.Error != "" {
		// Truncate long error messages
		errMsg := f.Error
		if len(errMsg) > 100 {
			errMsg = errMsg[:100] + "..."
		}
		r.write("   Error: %s\n", errMsg)
	}

	// Show context if available
	if len(f.Context) > 0 {
		r.write("   Context:\n")
		for _, line := range f.Context {
			// Highlight the error line
			if strings.HasPrefix(line, "> ") {
				r.write("   %s\n", colorRed(line))
			} else {
				r.write("   %s\n", line)
			}
		}
	}
	r.write("\n")
}

// NewTerminalReporter creates a new terminal reporter for local CI preview
func NewTerminalReporter() CIReporter {
	return &terminalReporter{
		output:   os.Stdout,
		failures: make([]CITestFailure, 0),
	}
}

// NewTerminalReporterWithWriter creates a terminal reporter with custom output
func NewTerminalReporterWithWriter(w io.Writer) CIReporter {
	return &terminalReporter{
		output:   w,
		failures: make([]CITestFailure, 0),
	}
}

// Start begins the CI report
func (r *terminalReporter) Start(metadata CIMetadata) error {
	r.started = true
	r.write("\n%s CI Mode Preview %s\n", colorBold("==="), colorBold("==="))
	if metadata.Platform != "" {
		r.write("Platform: %s\n", metadata.Platform)
	}
	if metadata.Branch != "" {
		r.write("Branch: %s\n", metadata.Branch)
	}
	r.write("\n")
	return nil
}

// ReportFailure reports a test failure
func (r *terminalReporter) ReportFailure(failure CITestFailure) error {
	r.failures = append(r.failures, failure)
	return nil
}

// WriteSummary writes the final summary
func (r *terminalReporter) WriteSummary(result *CIResult) error {
	r.write("\n")
	r.write("%s Test Results Summary %s\n", colorBold("==="), colorBold("==="))
	r.write("\n")

	// Status emoji and color
	statusStr := colorGreen("✅ PASSED")
	if result.Summary.Status == TestStatusFailed {
		statusStr = colorRed("❌ FAILED")
	}

	r.write("Status: %s\n", statusStr)
	r.write("Total:   %d\n", result.Summary.Total)
	r.write("Passed:  %s\n", colorGreen(fmt.Sprintf("%d", result.Summary.Passed)))
	r.write("Failed:  %s\n", colorRed(fmt.Sprintf("%d", result.Summary.Failed)))
	if result.Summary.Skipped > 0 {
		r.write("Skipped: %s\n", colorYellow(fmt.Sprintf("%d", result.Summary.Skipped)))
	}
	if result.Summary.Duration != "" {
		r.write("Duration: %s\n", result.Summary.Duration)
	}

	// List failures if any
	if len(result.Failures) > 0 {
		r.write("\n")
		r.write("%s Failed Tests %s\n", colorBold("---"), colorBold("---"))
		r.write("\n")

		for i, f := range result.Failures {
			r.writeFailure(i+1, f)
		}
	}

	return nil
}

// Close finalizes the reporter
func (r *terminalReporter) Close() error {
	return nil
}

// Color helper functions using ANSI escape codes
// These check for terminal capability
func colorBold(s string) string {
	if !isTerminal() {
		return s
	}
	return "\033[1m" + s + "\033[0m"
}

func colorRed(s string) string {
	if !isTerminal() {
		return s
	}
	return "\033[31m" + s + "\033[0m"
}

func colorGreen(s string) string {
	if !isTerminal() {
		return s
	}
	return "\033[32m" + s + "\033[0m"
}

func colorYellow(s string) string {
	if !isTerminal() {
		return s
	}
	return "\033[33m" + s + "\033[0m"
}

func colorCyan(s string) string {
	if !isTerminal() {
		return s
	}
	return "\033[36m" + s + "\033[0m"
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	// Check TERM environment variable
	if os.Getenv("TERM") == "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	// Check NO_COLOR environment variable (standard for disabling color)
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	// Simple heuristic: check if stdout is a file
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
