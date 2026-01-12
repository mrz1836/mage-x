// Package errors provides a comprehensive error handling framework for mage-x,
// including structured errors, error chaining, recovery mechanisms, and formatting.
//
// # Core Types
//
// The package provides several key types:
//
//   - MageError: Rich error type with codes, severity, and context
//   - ErrorChain: Chains multiple errors together
//   - ErrorBuilder: Fluent API for constructing errors
//
// # Error Codes
//
// Predefined error codes for common scenarios:
//
//   - ErrBuildFailed, ErrTestFailed, ErrLintFailed
//   - ErrCommandFailed, ErrConfigFailed
//   - ErrSecurityFailed, ErrNetworkFailed
//
// # Severity Levels
//
// Errors can be classified by severity:
//
//   - SeverityDebug, SeverityInfo, SeverityWarning
//   - SeverityError, SeverityCritical, SeverityFatal
//
// # Building Errors
//
// Use the builder pattern for complex errors:
//
//	err := errors.NewBuilder().
//	    WithCode(errors.ErrBuildFailed).
//	    WithMessage("compilation failed").
//	    WithSeverity(errors.SeverityError).
//	    WithField("file", "main.go").
//	    Build()
//
// # Error Checking
//
// Helper functions for error classification:
//
//   - IsNotFound(err): Check for not-found errors
//   - IsTimeout(err): Check for timeout errors
//   - IsRetryable(err): Check if error should be retried
//   - IsCritical(err): Check if error is critical
//
// # Recovery
//
// Safe execution with panic recovery:
//
//	err := errors.SafeExecute(func() error {
//	    // potentially panicking code
//	    return nil
//	})
//
// # Integration
//
// This package is designed to work seamlessly with Go's standard error handling
// and supports errors.Is() and errors.As() for error inspection.
package errors
