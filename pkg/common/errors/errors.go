// Package errors provides enhanced error handling with structured errors
package errors

import (
	"errors"
	"fmt"
	"os"
)

// Global instances
var (
	// DefaultRegistry is the global error registry
	DefaultRegistry = NewRegistry() //nolint:gochecknoglobals // Package-level default

	// DefaultHandler is the global error handler
	DefaultHandler = NewHandler() //nolint:gochecknoglobals // Package-level default

	// DefaultFormatter is the global error formatter
	DefaultFormatter = NewFormatter() //nolint:gochecknoglobals // Package-level default

	// DefaultRecovery is the global error recovery handler
	DefaultRecovery = NewRecovery() //nolint:gochecknoglobals // Package-level default

	// DefaultMetrics is the global error metrics collector
	DefaultMetrics = NewMetrics() //nolint:gochecknoglobals // Package-level default

	// DefaultLogger is the global error logger
	DefaultLogger = NewErrorLogger() //nolint:gochecknoglobals // Package-level default

	// DefaultNotifier is the global error notifier
	DefaultNotifier = NewErrorNotifier() //nolint:gochecknoglobals // Package-level default

	// DefaultTransformer is the global error transformer
	DefaultTransformer = NewErrorTransformer() //nolint:gochecknoglobals // Package-level default
)

// Package initialization
//
//nolint:gochecknoinits // Required for package-level initialization
func init() {
	// Set up default error handlers
	setupDefaultHandlers()

	// Register additional error codes if needed
	registerAdditionalErrors()
}

// setupDefaultHandlers configures default error handlers
func setupDefaultHandlers() {
	// Handle fatal errors
	DefaultHandler.OnSeverity(SeverityFatal, func(err MageError) error {
		fmt.Fprintf(os.Stderr, "FATAL: %v\n", err.Format(true))
		os.Exit(1)
		return nil
	})

	// Handle critical errors
	DefaultHandler.OnSeverity(SeverityCritical, func(err MageError) error {
		fmt.Fprintf(os.Stderr, "CRITICAL: %v\n", err.Format(false))
		return err
	})

	// Set default handler
	DefaultHandler.SetDefault(func(err error) error {
		var mageErr MageError
		if errors.As(err, &mageErr) {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", mageErr.Format(false))
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		}
		return err
	})
}

// registerAdditionalErrors registers additional error codes
func registerAdditionalErrors() {
	// Register deployment errors
	DefaultRegistry.RegisterWithSeverity(
		"DEPLOY_FAILED",
		"Deployment process failed",
		SeverityError,
	)

	DefaultRegistry.RegisterWithSeverity(
		"DEPLOY_ROLLBACK",
		"Deployment rolled back",
		SeverityWarning,
	)

	// Register security errors
	DefaultRegistry.RegisterWithSeverity(
		"SECURITY_VIOLATION",
		"Security violation detected",
		SeverityCritical,
	)

	DefaultRegistry.RegisterWithSeverity(
		"AUTH_FAILED",
		"Authentication failed",
		SeverityError,
	)
}

// Common error creation helpers

// NewBuildError creates a new build error
func NewBuildError(message string, cause error) MageError {
	return NewBuilder().
		WithCode(ErrBuildFailed).
		WithMessage(message).
		WithCause(cause).
		WithOperation("build").
		WithStackTrace().
		Build()
}

// NewConfigError creates a new configuration error
func NewConfigError(message string, configFile string) MageError {
	return NewBuilder().
		WithCode(ErrConfigInvalid).
		WithMessage(message).
		WithResource(configFile).
		WithOperation("config").
		Build()
}

// NewFileError creates a new file operation error
func NewFileError(code ErrorCode, message string, path string) MageError {
	return NewBuilder().
		WithCode(code).
		WithMessage(message).
		WithResource(path).
		WithOperation("file").
		Build()
}

// NewCommandError creates a new command execution error
func NewCommandError(command string, exitCode int, output string) MageError {
	return NewBuilder().
		WithCode(ErrCommandFailed).
		WithMessage("command failed: %s", command).
		WithField("command", command).
		WithField("exitCode", exitCode).
		WithField("output", output).
		WithOperation("exec").
		Build()
}

// NewValidationError creates a new validation error
func NewValidationError(field string, value interface{}, reason string) MageError {
	return NewBuilder().
		WithCode(ErrInvalidArgument).
		WithMessage("validation failed for field '%s': %s", field, reason).
		WithField("field", field).
		WithField("value", value).
		WithField("reason", reason).
		WithOperation("validate").
		Build()
}

// Error checking helpers

// IsNotFound returns true if the error is a not found error
func IsNotFound(err error) bool {
	return GetCode(err) == ErrNotFound || GetCode(err) == ErrFileNotFound
}

// IsTimeout returns true if the error is a timeout error
func IsTimeout(err error) bool {
	return GetCode(err) == ErrTimeout || GetCode(err) == ErrCommandTimeout
}

// IsPermissionDenied returns true if the error is a permission denied error
func IsPermissionDenied(err error) bool {
	return GetCode(err) == ErrPermissionDenied || GetCode(err) == ErrFileAccessDenied
}

// IsBuildError returns true if the error is a build-related error
func IsBuildError(err error) bool {
	code := GetCode(err)
	return code == ErrBuildFailed ||
		code == ErrCompileFailed ||
		code == ErrTestFailed ||
		code == ErrLintFailed ||
		code == ErrPackageFailed
}

// IsRetryable returns true if the error is retryable
func IsRetryable(err error) bool {
	if def, exists := DefaultRegistry.Get(GetCode(err)); exists {
		return def.Retryable
	}
	return false
}

// IsCritical returns true if the error has critical or fatal severity
func IsCritical(err error) bool {
	severity := GetSeverity(err)
	return severity == SeverityCritical || severity == SeverityFatal
}

// Error aggregation helpers

// Combine combines multiple errors into an error chain
func Combine(errors ...error) error {
	chain := NewChain()
	for _, err := range errors {
		if err != nil {
			chain.Add(err)
		}
	}

	if chain.Count() == 0 {
		return nil
	}
	if chain.Count() == 1 {
		return chain.First()
	}
	return chain
}

// FirstError returns the first non-nil error
func FirstError(errors ...error) error {
	for _, err := range errors {
		if err != nil {
			return err
		}
	}
	return nil
}

// Must panics if the error is not nil
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// MustValue returns the value or panics if error is not nil
func MustValue[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

// Handle handles an error using the default handler
func Handle(err error) error {
	return DefaultHandler.Handle(err)
}

// Recover recovers from panics and returns as error
func Recover() error {
	if r := recover(); r != nil {
		if err, ok := r.(error); ok {
			return Wrap(err, "panic recovered")
		}
		return Newf("panic recovered: %v", r)
	}
	return nil
}

// RecoverTo recovers from panics and stores in the provided error pointer
func RecoverTo(errPtr *error) {
	if r := recover(); r != nil {
		if err, ok := r.(error); ok {
			*errPtr = Wrap(err, "panic recovered")
		} else {
			*errPtr = Newf("panic recovered: %v", r)
		}
	}
}

// SafeExecute executes a function and recovers from panics
func SafeExecute(fn func() error) (err error) {
	defer RecoverTo(&err)
	return fn()
}

// SafeExecuteWithFallback executes a function with panic recovery and fallback
func SafeExecuteWithFallback(fn func() error, fallback func(error) error) error {
	err := SafeExecute(fn)
	if err != nil && fallback != nil {
		return fallback(err)
	}
	return err
}

// Logging convenience functions

// LogError logs an error using the default logger
func LogError(err error) {
	DefaultLogger.LogError(err)
}

// LogMageError logs a MageError using the default logger
func LogMageError(err MageError) {
	DefaultLogger.LogMageError(err)
}

// Notification convenience functions

// NotifyError sends a notification for an error using the default notifier
func NotifyError(err error) error {
	return DefaultNotifier.Notify(err)
}

// Transform convenience functions

// TransformError transforms an error using the default transformer
func TransformError(err error) error {
	return DefaultTransformer.Transform(err)
}

// Configuration functions

// SetErrorLogger sets the global error logger
func SetErrorLogger(logger ErrorLogger) {
	DefaultLogger = logger
}

// SetErrorNotifier sets the global error notifier
func SetErrorNotifier(notifier ErrorNotifier) {
	DefaultNotifier = notifier
}

// SetErrorTransformer sets the global error transformer
func SetErrorTransformer(transformer ErrorTransformer) {
	DefaultTransformer = transformer
}
