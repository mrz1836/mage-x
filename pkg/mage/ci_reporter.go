// Package mage provides CI reporter interface for test output
package mage

// CIReporter writes CI output in various formats
type CIReporter interface {
	// Start begins the test run report
	Start(metadata CIMetadata) error

	// ReportFailure reports a single test failure
	ReportFailure(failure CITestFailure) error

	// WriteSummary writes the final summary
	WriteSummary(result *CIResult) error

	// Close finalizes and closes the reporter
	Close() error
}

// GitHubReporterInterface extends CIReporter with GitHub-specific methods
type GitHubReporterInterface interface {
	CIReporter

	// WriteStepSummary writes markdown to GITHUB_STEP_SUMMARY
	WriteStepSummary(result *CIResult) error

	// WriteOutputs writes key-value pairs to GITHUB_OUTPUT
	WriteOutputs(outputs map[string]string) error
}

// JSONReporterInterface extends CIReporter with JSON-specific methods
type JSONReporterInterface interface {
	CIReporter

	// Flush forces any buffered output to be written
	Flush() error

	// GetOutputPath returns the path to the output file
	GetOutputPath() string
}

// MultiReporter combines multiple reporters into one
type MultiReporter struct {
	reporters []CIReporter
}

// NewMultiReporter creates a reporter that writes to multiple destinations
func NewMultiReporter(reporters ...CIReporter) *MultiReporter {
	return &MultiReporter{
		reporters: reporters,
	}
}

// Start begins the test run report on all reporters
func (m *MultiReporter) Start(metadata CIMetadata) error {
	for _, r := range m.reporters {
		if err := r.Start(metadata); err != nil {
			return err
		}
	}
	return nil
}

// ReportFailure reports a failure to all reporters
func (m *MultiReporter) ReportFailure(failure CITestFailure) error {
	for _, r := range m.reporters {
		if err := r.ReportFailure(failure); err != nil {
			return err
		}
	}
	return nil
}

// WriteSummary writes summary to all reporters
func (m *MultiReporter) WriteSummary(result *CIResult) error {
	for _, r := range m.reporters {
		if err := r.WriteSummary(result); err != nil {
			return err
		}
	}
	return nil
}

// Close finalizes all reporters
func (m *MultiReporter) Close() error {
	var lastErr error
	for _, r := range m.reporters {
		if err := r.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// NullReporter is a no-op reporter for testing
type NullReporter struct{}

// Start is a no-op
func (NullReporter) Start(_ CIMetadata) error { return nil }

// ReportFailure is a no-op
func (NullReporter) ReportFailure(_ CITestFailure) error { return nil }

// WriteSummary is a no-op
func (NullReporter) WriteSummary(_ *CIResult) error { return nil }

// Close is a no-op
func (NullReporter) Close() error { return nil }
