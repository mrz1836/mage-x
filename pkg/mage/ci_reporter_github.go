// Package mage provides GitHub Actions reporter for CI test output
package mage

import (
	"fmt"
	"os"
	"strings"

	"github.com/mrz1836/mage-x/pkg/common/fileops"
)

// githubReporter implements GitHubReporterInterface for GitHub Actions
type githubReporter struct {
	stepSummaryFile string
	outputFile      string
}

// NewGitHubReporter creates a GitHub Actions reporter
func NewGitHubReporter() GitHubReporterInterface {
	return &githubReporter{
		stepSummaryFile: os.Getenv("GITHUB_STEP_SUMMARY"),
		outputFile:      os.Getenv("GITHUB_OUTPUT"),
	}
}

// Start begins the test run report
func (r *githubReporter) Start(_ CIMetadata) error {
	// No initialization needed for GitHub
	return nil
}

// ReportFailure reports a single test failure as GitHub annotation
func (r *githubReporter) ReportFailure(failure CITestFailure) error {
	// Format: ::error file={name},line={line},col={col},title={title}::{message}
	var parts []string

	if failure.File != "" {
		parts = append(parts, fmt.Sprintf("file=%s", failure.File))
	}
	if failure.Line > 0 {
		parts = append(parts, fmt.Sprintf("line=%d", failure.Line))
	}
	if failure.Column > 0 {
		parts = append(parts, fmt.Sprintf("col=%d", failure.Column))
	}

	title := failure.Test
	if title == "" {
		title = string(failure.Type) + " error"
	}
	parts = append(parts, fmt.Sprintf("title=%s", escapeGitHubValue(title)))

	// Construct annotation
	annotation := "::error"
	if len(parts) > 0 {
		annotation += " " + strings.Join(parts, ",")
	}
	annotation += "::" + escapeGitHubMessage(failure.Error)

	// Write to stdout (GitHub Actions captures this)
	// Using fmt.Fprintln to avoid forbidigo linter warning about fmt.Println
	_, err := fmt.Fprintln(os.Stdout, annotation)
	if err != nil {
		return fmt.Errorf("failed to write GitHub annotation: %w", err)
	}

	return nil
}

// WriteSummary writes the final summary (delegates to WriteStepSummary)
func (r *githubReporter) WriteSummary(result *CIResult) error {
	return r.WriteStepSummary(result)
}

// WriteStepSummary writes markdown to GITHUB_STEP_SUMMARY
func (r *githubReporter) WriteStepSummary(result *CIResult) error {
	if r.stepSummaryFile == "" {
		return nil // Not running in GitHub Actions
	}

	// Allow workflows to skip step summary writing to avoid duplicate "Test Results" blocks.
	// The completion report workflow writes a single consolidated summary instead.
	// Check both env var names for flexibility (workflow uses MAGE_X_ prefix)
	if os.Getenv("MAGE_X_CI_SKIP_STEP_SUMMARY") == "true" || os.Getenv("MAGEX_CI_SKIP_STEP_SUMMARY") == "true" {
		return nil
	}

	// Skip writing empty summaries - these create confusing duplicate blocks with zeros
	// A valid summary must have either: test results (Total > 0), a defined status, or failures
	if result == nil ||
		(result.Summary.Total == 0 && result.Summary.Status == "" && len(result.Failures) == 0) {
		return nil
	}

	// Open file in append mode
	f, err := os.OpenFile(r.stepSummaryFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileops.PermFileSensitive)
	if err != nil {
		return fmt.Errorf("failed to open step summary file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	// Write markdown summary
	var sb strings.Builder

	sb.WriteString("## Test Results\n\n")

	// Status badge
	statusEmoji := "✅"
	switch result.Summary.Status {
	case TestStatusPassed:
		statusEmoji = "✅"
	case TestStatusFailed:
		statusEmoji = "❌"
	case TestStatusError:
		statusEmoji = "💥"
	}
	sb.WriteString(fmt.Sprintf("**Status**: %s %s\n\n", statusEmoji, result.Summary.Status))

	// Summary table
	sb.WriteString("| Metric | Count |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| ✅ Passed | %d |\n", result.Summary.Passed))
	sb.WriteString(fmt.Sprintf("| ❌ Failed | %d |\n", result.Summary.Failed))
	sb.WriteString(fmt.Sprintf("| ⏭️ Skipped | %d |\n", result.Summary.Skipped))
	sb.WriteString(fmt.Sprintf("| ⏱️ Duration | %s |\n", result.Summary.Duration))
	sb.WriteString("\n")

	// Failed tests details
	if len(result.Failures) > 0 {
		sb.WriteString("### Failed Tests\n\n")

		for _, failure := range result.Failures {
			sb.WriteString(fmt.Sprintf("<details>\n<summary>%s (%s)</summary>\n\n", failure.Test, failure.Package))

			if failure.File != "" {
				sb.WriteString(fmt.Sprintf("**File**: `%s", failure.File))
				if failure.Line > 0 {
					sb.WriteString(fmt.Sprintf(":%d", failure.Line))
				}
				sb.WriteString("`\n")
			}

			sb.WriteString(fmt.Sprintf("**Type**: %s\n", failure.Type))

			if failure.Error != "" {
				sb.WriteString(fmt.Sprintf("**Error**: %s\n", failure.Error))
			}

			if len(failure.Context) > 0 {
				sb.WriteString("\n```go\n")
				sb.WriteString(strings.Join(failure.Context, "\n"))
				sb.WriteString("\n```\n")
			}

			sb.WriteString("\n</details>\n\n")
		}
	}

	_, err = f.WriteString(sb.String())
	return err
}

// WriteOutputs writes key-value pairs to GITHUB_OUTPUT
func (r *githubReporter) WriteOutputs(outputs map[string]string) error {
	if r.outputFile == "" {
		return nil // Not running in GitHub Actions
	}

	// Open file in append mode
	f, err := os.OpenFile(r.outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileops.PermFileSensitive)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	for k, v := range outputs {
		// For multiline values, use delimiter syntax
		if strings.Contains(v, "\n") {
			delimiter := "EOF_" + k
			if _, err := fmt.Fprintf(f, "%s<<%s\n%s\n%s\n", k, delimiter, v, delimiter); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
		} else {
			if _, err := fmt.Fprintf(f, "%s=%s\n", k, v); err != nil {
				return fmt.Errorf("failed to write output: %w", err)
			}
		}
	}

	return nil
}

// Close finalizes the reporter
func (r *githubReporter) Close() error {
	return nil
}

// escapeGitHubValue escapes a value for use in GitHub annotation parameters
func escapeGitHubValue(s string) string {
	s = strings.ReplaceAll(s, "%", "%25")
	s = strings.ReplaceAll(s, "\r", "%0D")
	s = strings.ReplaceAll(s, "\n", "%0A")
	s = strings.ReplaceAll(s, ",", "%2C")
	return s
}

// escapeGitHubMessage escapes a message for GitHub annotations
func escapeGitHubMessage(s string) string {
	s = strings.ReplaceAll(s, "%", "%25")
	s = strings.ReplaceAll(s, "\r", "%0D")
	s = strings.ReplaceAll(s, "\n", "%0A")
	return s
}
