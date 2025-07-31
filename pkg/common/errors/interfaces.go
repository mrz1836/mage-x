// Package errors provides enhanced error handling with structured errors, error codes, and error chains
package errors

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrorCode represents a unique error code
type ErrorCode string

// Common error codes
const (
	// General errors
	ErrUnknown           ErrorCode = "UNKNOWN"
	ErrInternal          ErrorCode = "INTERNAL"
	ErrInvalidArgument   ErrorCode = "INVALID_ARGUMENT"
	ErrNotFound          ErrorCode = "NOT_FOUND"
	ErrAlreadyExists     ErrorCode = "ALREADY_EXISTS"
	ErrPermissionDenied  ErrorCode = "PERMISSION_DENIED"
	ErrResourceExhausted ErrorCode = "RESOURCE_EXHAUSTED"
	ErrCancelled         ErrorCode = "CANCELED"
	ErrTimeout           ErrorCode = "TIMEOUT"
	ErrNotImplemented    ErrorCode = "NOT_IMPLEMENTED"
	ErrUnavailable       ErrorCode = "UNAVAILABLE"

	// Build errors
	ErrBuildFailed     ErrorCode = "BUILD_FAILED"
	ErrCompileFailed   ErrorCode = "COMPILE_FAILED"
	ErrTestFailed      ErrorCode = "TEST_FAILED"
	ErrLintFailed      ErrorCode = "LINT_FAILED"
	ErrPackageFailed   ErrorCode = "PACKAGE_FAILED"
	ErrDependencyError ErrorCode = "DEPENDENCY_ERROR"

	// Environment errors
	ErrEnvVarNotSet    ErrorCode = "ENV_VAR_NOT_SET"
	ErrInvalidEnvValue ErrorCode = "INVALID_ENV_VALUE"

	// File operation errors
	ErrFileNotFound     ErrorCode = "FILE_NOT_FOUND"
	ErrFileAccessDenied ErrorCode = "FILE_ACCESS_DENIED"
	ErrFileExists       ErrorCode = "FILE_EXISTS"
	ErrDirectoryExists  ErrorCode = "DIRECTORY_EXISTS"
	ErrNotADirectory    ErrorCode = "NOT_A_DIRECTORY"
	ErrNotAFile         ErrorCode = "NOT_A_FILE"

	// Configuration errors
	ErrConfigNotFound    ErrorCode = "CONFIG_NOT_FOUND"
	ErrConfigInvalid     ErrorCode = "CONFIG_INVALID"
	ErrConfigParseFailed ErrorCode = "CONFIG_PARSE_FAILED"

	// Command errors
	ErrCommandFailed    ErrorCode = "COMMAND_FAILED"
	ErrCommandNotFound  ErrorCode = "COMMAND_NOT_FOUND"
	ErrCommandTimeout   ErrorCode = "COMMAND_TIMEOUT"
	ErrCommandCancelled ErrorCode = "COMMAND_CANCELED"

	// Security errors
	ErrSecurityFailed ErrorCode = "SECURITY_FAILED"
	ErrUnauthorized   ErrorCode = "UNAUTHORIZED"

	// Format errors
	ErrFormatCheckFailed ErrorCode = "FORMAT_CHECK_FAILED"
)

// Severity represents the severity level of an error
type Severity int

const (
	// SeverityDebug represents debug-level severity.
	SeverityDebug Severity = iota
	// SeverityInfo represents informational-level severity.
	SeverityInfo
	// SeverityWarning represents warning-level severity.
	SeverityWarning
	// SeverityError represents error-level severity.
	SeverityError
	// SeverityCritical represents critical-level severity.
	SeverityCritical
	// SeverityFatal represents fatal-level severity.
	SeverityFatal
)

// ErrorContext contains contextual information about an error
type ErrorContext struct {
	Operation   string                 // Operation being performed
	Resource    string                 // Resource being operated on
	User        string                 // User who triggered the operation
	RequestID   string                 // Request identifier for tracing
	Timestamp   time.Time              // When the error occurred
	StackTrace  string                 // Stack trace at error point
	Fields      map[string]interface{} // Additional context fields
	Environment string                 // Environment (dev, staging, prod)
	Version     string                 // Application version
}

// MageError is the main error interface for the mage system
type MageError interface {
	error
	Code() ErrorCode
	Severity() Severity
	Context() ErrorContext
	Cause() error
	Unwrap() error
	WithCode(code ErrorCode) MageError
	WithSeverity(severity Severity) MageError
	WithContext(ctx *ErrorContext) MageError
	WithField(key string, value interface{}) MageError
	WithFields(fields map[string]interface{}) MageError
	WithCause(cause error) MageError
	WithOperation(operation string) MageError
	WithResource(resource string) MageError
	Format(includeStack bool) string
	Is(target error) bool
	As(target interface{}) bool
}

// ErrorBuilder provides a fluent interface for building errors
type ErrorBuilder interface {
	WithMessage(format string, args ...interface{}) ErrorBuilder
	WithCode(code ErrorCode) ErrorBuilder
	WithSeverity(severity Severity) ErrorBuilder
	WithContext(ctx *ErrorContext) ErrorBuilder
	WithField(key string, value interface{}) ErrorBuilder
	WithFields(fields map[string]interface{}) ErrorBuilder
	WithCause(cause error) ErrorBuilder
	WithOperation(operation string) ErrorBuilder
	WithResource(resource string) ErrorBuilder
	WithStackTrace() ErrorBuilder
	Build() MageError
}

// ErrorChain represents a chain of errors
type ErrorChain interface {
	error
	Add(err error) ErrorChain
	AddWithContext(err error, ctx *ErrorContext) ErrorChain
	Errors() []error
	First() error
	Last() error
	Count() int
	HasError(code ErrorCode) bool
	FindByCode(code ErrorCode) MageError
	ForEach(fn func(error) error) error
	Filter(predicate func(error) bool) []error
	ToSlice() []error
}

// ErrorHandler handles errors in different ways
type ErrorHandler interface {
	Handle(err error) error
	HandleWithContext(ctx context.Context, err error) error
	OnError(code ErrorCode, handler func(MageError) error) ErrorHandler
	OnSeverity(severity Severity, handler func(MageError) error) ErrorHandler
	SetDefault(handler func(error) error) ErrorHandler
	SetFallback(handler func(error) error) ErrorHandler
}

// ErrorLogger logs errors with appropriate formatting
type ErrorLogger interface {
	LogError(err error)
	LogErrorWithContext(ctx context.Context, err error)
	LogMageError(err MageError)
	SetLogLevel(severity Severity)
	SetFormatter(formatter ErrorFormatter)
}

// ErrorFormatter formats errors for display
type ErrorFormatter interface {
	Format(err error) string
	FormatMageError(err MageError) string
	FormatChain(chain ErrorChain) string
	FormatWithOptions(err error, opts FormatOptions) string
}

// FormatOptions controls error formatting
type FormatOptions struct {
	IncludeStack     bool
	IncludeContext   bool
	IncludeCause     bool
	IncludeTimestamp bool
	IncludeFields    bool
	IndentLevel      int
	MaxDepth         int
	TimeFormat       string
	FieldsSeparator  string
	UseColor         bool
	CompactMode      bool
}

// ErrorRegistry manages error code definitions
type ErrorRegistry interface {
	Register(code ErrorCode, description string) error
	RegisterWithSeverity(code ErrorCode, description string, severity Severity) error
	Unregister(code ErrorCode) error
	Get(code ErrorCode) (ErrorDefinition, bool)
	List() []ErrorDefinition
	ListByPrefix(prefix string) []ErrorDefinition
	ListBySeverity(severity Severity) []ErrorDefinition
	Contains(code ErrorCode) bool
	Clear() error
}

// ErrorDefinition defines an error code
type ErrorDefinition struct {
	Code        ErrorCode
	Description string
	Severity    Severity
	Category    string
	Retryable   bool
	UserMessage string
	HelpURL     string
	Tags        []string
}

// ErrorRecovery provides error recovery mechanisms
type ErrorRecovery interface {
	Recover(fn func() error) error
	RecoverWithFallback(fn func() error, fallback func(error) error) error
	RecoverWithRetry(fn func() error, retries int, delay time.Duration) error
	RecoverWithBackoff(fn func() error, config BackoffConfig) error
	RecoverWithContext(ctx context.Context, fn func() error) error
}

// BackoffConfig configures exponential backoff
type BackoffConfig struct {
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	MaxRetries   int
	RetryIf      func(error) bool
}

// ErrorMetrics tracks error metrics
type ErrorMetrics interface {
	RecordError(err error)
	RecordMageError(err MageError)
	GetCount(code ErrorCode) int64
	GetCountBySeverity(severity Severity) int64
	GetRate(code ErrorCode, duration time.Duration) float64
	GetTopErrors(limit int) []ErrorStat
	Reset() error
}

// ErrorStat represents error statistics
type ErrorStat struct {
	Code        ErrorCode
	Count       int64
	LastSeen    time.Time
	FirstSeen   time.Time
	AverageRate float64
}

// ErrorNotifier sends error notifications
type ErrorNotifier interface {
	Notify(err error) error
	NotifyWithContext(ctx context.Context, err error) error
	SetThreshold(severity Severity)
	SetRateLimit(duration time.Duration, count int)
	AddChannel(channel NotificationChannel) error
	RemoveChannel(name string) error
}

// NotificationChannel represents a notification channel
type NotificationChannel interface {
	Name() string
	Send(ctx context.Context, notification *ErrorNotification) error
	IsEnabled() bool
	SetEnabled(enabled bool)
}

// ErrorNotification represents an error notification
type ErrorNotification struct {
	Error       MageError
	Timestamp   time.Time
	Environment string
	Hostname    string
	Service     string
	Metadata    map[string]interface{}
}

// ErrorTransformer transforms errors
type ErrorTransformer interface {
	Transform(err error) error
	TransformCode(from, to ErrorCode) ErrorTransformer
	TransformSeverity(from, to Severity) ErrorTransformer
	AddTransformer(fn func(error) error) ErrorTransformer
	RemoveTransformer(name string) ErrorTransformer
}

// ErrorMatcher matches errors based on criteria
type ErrorMatcher interface {
	Match(err error) bool
	MatchCode(code ErrorCode) ErrorMatcher
	MatchSeverity(severity Severity) ErrorMatcher
	MatchMessage(pattern string) ErrorMatcher
	MatchType(errType error) ErrorMatcher
	MatchField(key string, value interface{}) ErrorMatcher
	MatchAny(matchers ...ErrorMatcher) ErrorMatcher
	MatchAll(matchers ...ErrorMatcher) ErrorMatcher
	Not() ErrorMatcher
}

// Helper functions

// New creates a new MageError with the given message
func New(message string) MageError {
	return NewBuilder().WithMessage(message).Build()
}

// Newf creates a new MageError with a formatted message
func Newf(format string, args ...interface{}) MageError {
	return NewBuilder().WithMessage(format, args...).Build()
}

// WithCode creates a new MageError with the given code and message
func WithCode(code ErrorCode, message string) MageError {
	return NewBuilder().WithCode(code).WithMessage(message).Build()
}

// WithCodef creates a new MageError with the given code and formatted message
func WithCodef(code ErrorCode, format string, args ...interface{}) MageError {
	return NewBuilder().WithCode(code).WithMessage(format, args...).Build()
}

// Wrap wraps an error with additional context
func Wrap(err error, message string) MageError {
	if err == nil {
		return nil
	}
	return NewBuilder().WithCause(err).WithMessage(message).Build()
}

// Wrapf wraps an error with a formatted message
func Wrapf(err error, format string, args ...interface{}) MageError {
	if err == nil {
		return nil
	}
	return NewBuilder().WithCause(err).WithMessage(format, args...).Build()
}

// Is reports whether any error in err's chain matches target
func Is(err, target error) bool {
	var mageErr MageError
	if errors.As(err, &mageErr) {
		return mageErr.Is(target)
	}
	return errors.Is(fmt.Errorf("%w", err), target)
}

// As finds the first error in err's chain that matches target
func As(err error, target interface{}) bool {
	var mageErr MageError
	if errors.As(err, &mageErr) {
		return mageErr.As(target)
	}
	return fmt.Errorf("%w", err) != nil
}

// GetCode returns the error code from a MageError
func GetCode(err error) ErrorCode {
	var mageErr MageError
	if errors.As(err, &mageErr) {
		return mageErr.Code()
	}
	return ErrUnknown
}

// GetSeverity returns the severity from a MageError
func GetSeverity(err error) Severity {
	var mageErr MageError
	if errors.As(err, &mageErr) {
		return mageErr.Severity()
	}
	return SeverityError
}

// Factory functions

// NewBuilder creates a new error builder
func NewBuilder() ErrorBuilder {
	return NewErrorBuilder()
}

// NewChain creates a new error chain
func NewChain() ErrorChain {
	return NewErrorChain()
}

// NewHandler creates a new error handler
func NewHandler() ErrorHandler {
	return NewErrorHandler()
}

// NewRegistry creates a new error registry
func NewRegistry() ErrorRegistry {
	return NewErrorRegistry()
}

// NewRecovery creates a new error recovery handler
func NewRecovery() ErrorRecovery {
	return &DefaultErrorRecovery{}
}

// NewMetrics creates a new error metrics collector
func NewMetrics() ErrorMetrics {
	return NewErrorMetrics()
}

// NewMatcher creates a new error matcher
func NewMatcher() ErrorMatcher {
	return &DefaultErrorMatcher{
		matchers: make([]func(error) bool, 0),
	}
}

// Default implementations (defined in separate files)
type (
	// DefaultErrorBuilder provides default error building functionality.
	DefaultErrorBuilder struct{}
	// DefaultChainError provides default error chaining functionality.
	DefaultChainError struct{ errors []error }
	// DefaultErrorHandler provides default error handling functionality.
	DefaultErrorHandler struct {
		handlers         map[ErrorCode]func(MageError) error
		severityHandlers map[Severity]func(MageError) error
	}
)

type (
	// DefaultErrorRegistry provides default error registry functionality.
	DefaultErrorRegistry struct{ definitions map[ErrorCode]ErrorDefinition }
	// DefaultErrorRecovery provides default error recovery functionality.
	DefaultErrorRecovery struct{}
	// DefaultErrorMetrics provides default error metrics functionality.
	DefaultErrorMetrics struct{ counts map[ErrorCode]*ErrorStat }
	// DefaultErrorMatcher provides default error matching functionality.
	DefaultErrorMatcher struct{ matchers []func(error) bool }
)
