package errors

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Error definitions for mage error operations
var (
	ErrUnknownSeverity = errors.New("unknown severity")
)

// DefaultMageError is the default implementation of MageError
type DefaultMageError struct {
	mu         sync.RWMutex
	message    string
	code       ErrorCode
	severity   Severity
	context    ErrorContext
	cause      error
	stackTrace string
}

// NewMageError creates a new DefaultMageError
func NewMageError(message string) *DefaultMageError {
	return &DefaultMageError{
		message:  message,
		code:     ErrUnknown,
		severity: SeverityError,
		context: ErrorContext{
			Timestamp: time.Now(),
			Fields:    make(map[string]interface{}),
		},
	}
}

// Error returns the error message
func (e *DefaultMageError) Error() string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

// Code returns the error code
func (e *DefaultMageError) Code() ErrorCode {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.code
}

// Severity returns the error severity
func (e *DefaultMageError) Severity() Severity {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.severity
}

// Context returns the error context
func (e *DefaultMageError) Context() ErrorContext {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// Return a copy to prevent external modification
	ctx := e.context
	if e.context.Fields != nil {
		ctx.Fields = make(map[string]interface{})
		for k, v := range e.context.Fields {
			ctx.Fields[k] = v
		}
	}
	return ctx
}

// Cause returns the underlying cause
func (e *DefaultMageError) Cause() error {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.cause
}

// Unwrap returns the wrapped error (for errors.Is and errors.As)
func (e *DefaultMageError) Unwrap() error {
	return e.Cause()
}

// WithCode sets the error code
func (e *DefaultMageError) WithCode(code ErrorCode) MageError {
	e.mu.Lock()
	defer e.mu.Unlock()

	newErr := e.clone()
	newErr.code = code
	return newErr
}

// WithSeverity sets the error severity
func (e *DefaultMageError) WithSeverity(severity Severity) MageError {
	e.mu.Lock()
	defer e.mu.Unlock()

	newErr := e.clone()
	newErr.severity = severity
	return newErr
}

// WithContext sets the error context
func (e *DefaultMageError) WithContext(ctx *ErrorContext) MageError {
	e.mu.Lock()
	defer e.mu.Unlock()

	newErr := e.clone()
	if ctx != nil {
		newErr.context = *ctx
	}
	if newErr.context.Fields == nil {
		newErr.context.Fields = make(map[string]interface{})
	}
	return newErr
}

// WithField adds a context field
func (e *DefaultMageError) WithField(key string, value interface{}) MageError {
	e.mu.Lock()
	defer e.mu.Unlock()

	newErr := e.clone()
	if newErr.context.Fields == nil {
		newErr.context.Fields = make(map[string]interface{})
	}
	newErr.context.Fields[key] = value
	return newErr
}

// WithFields adds multiple context fields
func (e *DefaultMageError) WithFields(fields map[string]interface{}) MageError {
	e.mu.Lock()
	defer e.mu.Unlock()

	newErr := e.clone()
	if newErr.context.Fields == nil {
		newErr.context.Fields = make(map[string]interface{})
	}
	for k, v := range fields {
		newErr.context.Fields[k] = v
	}
	return newErr
}

// WithCause sets the underlying cause
func (e *DefaultMageError) WithCause(cause error) MageError {
	e.mu.Lock()
	defer e.mu.Unlock()

	newErr := e.clone()
	newErr.cause = cause
	return newErr
}

// WithOperation sets the operation in context
func (e *DefaultMageError) WithOperation(operation string) MageError {
	e.mu.Lock()
	defer e.mu.Unlock()

	newErr := e.clone()
	newErr.context.Operation = operation
	return newErr
}

// WithResource sets the resource in context
func (e *DefaultMageError) WithResource(resource string) MageError {
	e.mu.Lock()
	defer e.mu.Unlock()

	newErr := e.clone()
	newErr.context.Resource = resource
	return newErr
}

// Format formats the error with options
func (e *DefaultMageError) Format(includeStack bool) string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var sb strings.Builder

	// Error code and message
	sb.WriteString(fmt.Sprintf("[%s] %s", e.code, e.message))

	// Context information
	if e.context.Operation != "" {
		sb.WriteString(fmt.Sprintf("\n  Operation: %s", e.context.Operation))
	}
	if e.context.Resource != "" {
		sb.WriteString(fmt.Sprintf("\n  Resource: %s", e.context.Resource))
	}

	// Fields
	if len(e.context.Fields) > 0 {
		sb.WriteString("\n  Fields:")
		for k, v := range e.context.Fields {
			sb.WriteString(fmt.Sprintf("\n    %s: %v", k, v))
		}
	}

	// Cause
	if e.cause != nil {
		sb.WriteString(fmt.Sprintf("\n  Caused by: %v", e.cause))
	}

	// Stack trace
	if includeStack && e.stackTrace != "" {
		sb.WriteString("\n  Stack trace:\n")
		sb.WriteString(e.stackTrace)
	}

	return sb.String()
}

// Is reports whether the target error is in the error chain
func (e *DefaultMageError) Is(target error) bool {
	if errors.Is(e.cause, target) {
		return true
	}

	// Check if target is a MageError with the same code
	if targetMage, ok := target.(MageError); ok {
		if e.Code() == targetMage.Code() {
			return true
		}
	}

	// Check the cause chain
	if e.cause != nil {
		if causeErr, ok := e.cause.(interface{ Is(target error) bool }); ok {
			return causeErr.Is(target)
		}
		return errors.Is(e.cause, target)
	}

	return false
}

// As finds the first error in the chain that matches target
func (e *DefaultMageError) As(target interface{}) bool {
	if target == nil {
		return false
	}

	// Try to assign this error
	if targetErr, ok := target.(**DefaultMageError); ok {
		*targetErr = e
		return true
	}

	if targetErr, ok := target.(*MageError); ok {
		*targetErr = e
		return true
	}

	// Try the cause chain
	if e.cause != nil {
		if causeErr, ok := e.cause.(interface{ As(target interface{}) bool }); ok {
			return causeErr.As(target)
		}
	}

	return false
}

// clone creates a copy of the error
func (e *DefaultMageError) clone() *DefaultMageError {
	newErr := &DefaultMageError{
		message:    e.message,
		code:       e.code,
		severity:   e.severity,
		context:    e.context,
		cause:      e.cause,
		stackTrace: e.stackTrace,
	}

	// Deep copy fields
	if e.context.Fields != nil {
		newErr.context.Fields = make(map[string]interface{})
		for k, v := range e.context.Fields {
			newErr.context.Fields[k] = v
		}
	}

	return newErr
}

// captureStackTrace captures the current stack trace
func captureStackTrace(skip int) string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip+2, pcs[:])

	var sb strings.Builder
	frames := runtime.CallersFrames(pcs[:n])

	for {
		frame, more := frames.Next()

		// Skip runtime frames
		if strings.Contains(frame.Function, "runtime.") {
			if !more {
				break
			}
			continue
		}

		sb.WriteString(fmt.Sprintf("    %s\n        %s:%d\n",
			frame.Function, frame.File, frame.Line))

		if !more {
			break
		}
	}

	return sb.String()
}

// String returns a string representation of the severity
func (s Severity) String() string {
	switch s {
	case SeverityDebug:
		return "DEBUG"
	case SeverityInfo:
		return "INFO"
	case SeverityWarning:
		return "WARNING"
	case SeverityError:
		return "ERROR"
	case SeverityCritical:
		return "CRITICAL"
	case SeverityFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// MarshalText implements encoding.TextMarshaler
func (s Severity) MarshalText() ([]byte, error) {
	return []byte(s.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler for Severity.
// It parses severity levels from their string representation.
func (s *Severity) UnmarshalText(text []byte) error {
	str := string(text)
	switch strings.ToUpper(str) {
	case "DEBUG":
		*s = SeverityDebug
	case "INFO":
		*s = SeverityInfo
	case "WARNING":
		*s = SeverityWarning
	case "ERROR":
		*s = SeverityError
	case "CRITICAL":
		*s = SeverityCritical
	case "FATAL":
		*s = SeverityFatal
	default:
		return fmt.Errorf("%w: %s", ErrUnknownSeverity, str)
	}
	return nil
}
