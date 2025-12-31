// Package errors provides enhanced error handling with structured errors
package errors

import (
	"errors"
	"fmt"
	"os"
	"sync"
)

// Lazy-initialized singleton instances using sync.Once pattern.
// This avoids init() functions which can cause unpredictable initialization order.
//
//nolint:gochecknoglobals // Package-level singletons with lazy initialization
var (
	defaultRegistryOnce sync.Once
	defaultRegistry     ErrorRegistry

	defaultHandlerOnce sync.Once
	defaultHandler     ErrorHandler

	defaultFormatterOnce sync.Once
	defaultFormatter     *DefaultErrorFormatter

	defaultRecoveryOnce sync.Once
	defaultRecovery     ErrorRecovery

	defaultMetricsOnce sync.Once
	defaultMetrics     ErrorMetrics

	defaultLoggerOnce sync.Once
	defaultLogger     ErrorLogger

	defaultNotifierOnce sync.Once
	defaultNotifier     ErrorNotifier

	defaultTransformerOnce sync.Once
	defaultTransformer     ErrorTransformer
)

// GetDefaultRegistry returns the global error registry (lazy initialized)
func GetDefaultRegistry() ErrorRegistry {
	defaultRegistryOnce.Do(func() {
		defaultRegistry = NewRegistry()
		registerAdditionalErrors(defaultRegistry)
	})
	return defaultRegistry
}

// GetDefaultHandler returns the global error handler (lazy initialized)
func GetDefaultHandler() ErrorHandler {
	defaultHandlerOnce.Do(func() {
		defaultHandler = NewHandler()
		setupDefaultHandlers(defaultHandler)
	})
	return defaultHandler
}

// GetDefaultFormatter returns the global error formatter (lazy initialized)
func GetDefaultFormatter() *DefaultErrorFormatter {
	defaultFormatterOnce.Do(func() {
		defaultFormatter = NewFormatter()
	})
	return defaultFormatter
}

// GetDefaultRecovery returns the global error recovery handler (lazy initialized)
func GetDefaultRecovery() ErrorRecovery {
	defaultRecoveryOnce.Do(func() {
		defaultRecovery = NewRecovery()
	})
	return defaultRecovery
}

// GetDefaultMetrics returns the global error metrics collector (lazy initialized)
func GetDefaultMetrics() ErrorMetrics {
	defaultMetricsOnce.Do(func() {
		defaultMetrics = NewMetrics()
	})
	return defaultMetrics
}

// GetDefaultLogger returns the global error logger (lazy initialized)
func GetDefaultLogger() ErrorLogger {
	defaultLoggerOnce.Do(func() {
		defaultLogger = NewErrorLogger()
	})
	return defaultLogger
}

// GetDefaultNotifier returns the global error notifier (lazy initialized)
func GetDefaultNotifier() ErrorNotifier {
	defaultNotifierOnce.Do(func() {
		defaultNotifier = NewErrorNotifier()
	})
	return defaultNotifier
}

// GetDefaultTransformer returns the global error transformer (lazy initialized)
func GetDefaultTransformer() ErrorTransformer {
	defaultTransformerOnce.Do(func() {
		defaultTransformer = NewErrorTransformer()
	})
	return defaultTransformer
}

// setupDefaultHandlers configures default error handlers on the given handler
func setupDefaultHandlers(handler ErrorHandler) {
	// Handle fatal errors
	handler.OnSeverity(SeverityFatal, func(err MageError) error {
		fmt.Fprintf(os.Stderr, "FATAL: %v\n", err.Format(true))
		os.Exit(1)
		return nil
	})

	// Handle critical errors
	handler.OnSeverity(SeverityCritical, func(err MageError) error {
		fmt.Fprintf(os.Stderr, "CRITICAL: %v\n", err.Format(false))
		return err
	})

	// Set default handler
	handler.SetDefault(func(err error) error {
		var mageErr MageError
		if errors.As(err, &mageErr) {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", mageErr.Format(false))
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		}
		return err
	})
}

// registerAdditionalErrors registers additional error codes on the given registry
func registerAdditionalErrors(registry ErrorRegistry) {
	// Register deployment errors
	if err := registry.RegisterWithSeverity(
		"DEPLOY_FAILED",
		"Deployment process failed",
		SeverityError,
	); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register error DEPLOY_FAILED: %v\n", err)
	}

	if err := registry.RegisterWithSeverity(
		"DEPLOY_ROLLBACK",
		"Deployment rolled back",
		SeverityWarning,
	); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register error DEPLOY_ROLLBACK: %v\n", err)
	}

	// Register security errors
	if err := registry.RegisterWithSeverity(
		"SECURITY_VIOLATION",
		"Security violation detected",
		SeverityCritical,
	); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register error SECURITY_VIOLATION: %v\n", err)
	}

	if err := registry.RegisterWithSeverity(
		"AUTH_FAILED",
		"Authentication failed",
		SeverityError,
	); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to register error AUTH_FAILED: %v\n", err)
	}
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
func NewConfigError(message, configFile string) MageError {
	return NewBuilder().
		WithCode(ErrConfigInvalid).
		WithMessage(message).
		WithResource(configFile).
		WithOperation("config").
		Build()
}

// NewFileError creates a new file operation error
func NewFileError(code ErrorCode, message, path string) MageError {
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
	if def, exists := GetDefaultRegistry().Get(GetCode(err)); exists {
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
func Combine(errs ...error) error {
	chain := NewChain()
	for _, err := range errs {
		if err != nil {
			chain = chain.Add(err)
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
func FirstError(errs ...error) error {
	for _, err := range errs {
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
	return GetDefaultHandler().Handle(err)
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

// RecoverTo recovers from panics and returns an error
func RecoverTo() error {
	if r := recover(); r != nil {
		if err, ok := r.(error); ok {
			return Wrap(err, "panic recovered")
		}
		return Newf("panic recovered: %v", r)
	}
	return nil
}

// SafeExecute executes a function and recovers from panics
func SafeExecute(fn func() error) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if err != nil {
			return
		}
		if recoveredErr, ok := r.(error); ok {
			err = Wrap(recoveredErr, "panic recovered")
			return
		}
		err = Newf("panic recovered: %v", r)
	}()
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
	GetDefaultLogger().LogError(err)
}

// LogMageError logs a MageError using the default logger
func LogMageError(err MageError) {
	GetDefaultLogger().LogMageError(err)
}

// Notification convenience functions

// NotifyError sends a notification for an error using the default notifier
func NotifyError(err error) error {
	return GetDefaultNotifier().Notify(err)
}

// Transform convenience functions

// TransformError transforms an error using the default transformer
func TransformError(err error) error {
	return GetDefaultTransformer().Transform(err)
}

// Configuration functions

// SetErrorLogger sets the global error logger.
// Note: This must be called before any logging occurs to take effect.
func SetErrorLogger(logger ErrorLogger) {
	defaultLoggerOnce.Do(func() {}) // Ensure initialization has happened
	defaultLogger = logger
}

// SetErrorNotifier sets the global error notifier.
// Note: This must be called before any notifications occur to take effect.
func SetErrorNotifier(notifier ErrorNotifier) {
	defaultNotifierOnce.Do(func() {}) // Ensure initialization has happened
	defaultNotifier = notifier
}

// SetErrorTransformer sets the global error transformer.
// Note: This must be called before any transformations occur to take effect.
func SetErrorTransformer(transformer ErrorTransformer) {
	defaultTransformerOnce.Do(func() {}) // Ensure initialization has happened
	defaultTransformer = transformer
}
