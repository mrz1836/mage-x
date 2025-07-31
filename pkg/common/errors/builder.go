package errors

import (
	"fmt"
	"time"
)

// RealDefaultErrorBuilder is the actual implementation of ErrorBuilder
type RealDefaultErrorBuilder struct {
	err *DefaultMageError
}

// NewErrorBuilder creates a new error builder instance
func NewErrorBuilder() *RealDefaultErrorBuilder {
	return &RealDefaultErrorBuilder{
		err: &DefaultMageError{
			code:     ErrUnknown,
			severity: SeverityError,
			context: ErrorContext{
				Timestamp: time.Now(),
				Fields:    make(map[string]interface{}),
			},
		},
	}
}

// WithMessage sets the error message
func (b *RealDefaultErrorBuilder) WithMessage(format string, args ...interface{}) ErrorBuilder {
	if len(args) > 0 {
		b.err.message = fmt.Sprintf(format, args...)
	} else {
		b.err.message = format
	}
	return b
}

// WithCode sets the error code
func (b *RealDefaultErrorBuilder) WithCode(code ErrorCode) ErrorBuilder {
	b.err.code = code
	return b
}

// WithSeverity sets the error severity
func (b *RealDefaultErrorBuilder) WithSeverity(severity Severity) ErrorBuilder {
	b.err.severity = severity
	return b
}

// WithContext sets the error context
func (b *RealDefaultErrorBuilder) WithContext(ctx *ErrorContext) ErrorBuilder {
	if ctx != nil {
		b.err.context = *ctx
	}
	if b.err.context.Fields == nil {
		b.err.context.Fields = make(map[string]interface{})
	}
	if b.err.context.Timestamp.IsZero() {
		b.err.context.Timestamp = time.Now()
	}
	return b
}

// WithField adds a field to the error context
func (b *RealDefaultErrorBuilder) WithField(key string, value interface{}) ErrorBuilder {
	if b.err.context.Fields == nil {
		b.err.context.Fields = make(map[string]interface{})
	}
	b.err.context.Fields[key] = value
	return b
}

// WithFields adds multiple fields to the error context
func (b *RealDefaultErrorBuilder) WithFields(fields map[string]interface{}) ErrorBuilder {
	if b.err.context.Fields == nil {
		b.err.context.Fields = make(map[string]interface{})
	}
	for k, v := range fields {
		b.err.context.Fields[k] = v
	}
	return b
}

// WithCause sets the underlying cause
func (b *RealDefaultErrorBuilder) WithCause(cause error) ErrorBuilder {
	b.err.cause = cause
	return b
}

// WithOperation sets the operation in the error context
func (b *RealDefaultErrorBuilder) WithOperation(operation string) ErrorBuilder {
	b.err.context.Operation = operation
	return b
}

// WithResource sets the resource in the error context
func (b *RealDefaultErrorBuilder) WithResource(resource string) ErrorBuilder {
	b.err.context.Resource = resource
	return b
}

// WithStackTrace captures the current stack trace
func (b *RealDefaultErrorBuilder) WithStackTrace() ErrorBuilder {
	b.err.stackTrace = captureStackTrace(1)
	return b
}

// Build creates the final MageError
func (b *RealDefaultErrorBuilder) Build() MageError {
	// Create a copy to ensure immutability
	result := &DefaultMageError{
		message:    b.err.message,
		code:       b.err.code,
		severity:   b.err.severity,
		context:    b.err.context,
		cause:      b.err.cause,
		stackTrace: b.err.stackTrace,
	}

	// Deep copy fields
	if b.err.context.Fields != nil {
		result.context.Fields = make(map[string]interface{})
		for k, v := range b.err.context.Fields {
			result.context.Fields[k] = v
		}
	}

	return result
}

// WithMessage adds a formatted message to the error builder
func (b *DefaultErrorBuilder) WithMessage(format string, args ...interface{}) ErrorBuilder {
	return NewErrorBuilder().WithMessage(format, args...)
}

// WithCode adds an error code to the error builder
func (b *DefaultErrorBuilder) WithCode(code ErrorCode) ErrorBuilder {
	return NewErrorBuilder().WithCode(code)
}

// WithSeverity adds a severity level to the error builder
func (b *DefaultErrorBuilder) WithSeverity(severity Severity) ErrorBuilder {
	return NewErrorBuilder().WithSeverity(severity)
}

// WithContext adds context information to the error builder
func (b *DefaultErrorBuilder) WithContext(ctx *ErrorContext) ErrorBuilder {
	return NewErrorBuilder().WithContext(ctx)
}

// WithField adds a key-value field to the error builder
func (b *DefaultErrorBuilder) WithField(key string, value interface{}) ErrorBuilder {
	return NewErrorBuilder().WithField(key, value)
}

// WithFields adds multiple key-value fields to the error builder
func (b *DefaultErrorBuilder) WithFields(fields map[string]interface{}) ErrorBuilder {
	return NewErrorBuilder().WithFields(fields)
}

// WithCause adds a cause error to the error builder
func (b *DefaultErrorBuilder) WithCause(cause error) ErrorBuilder {
	return NewErrorBuilder().WithCause(cause)
}

// WithOperation adds an operation name to the error builder
func (b *DefaultErrorBuilder) WithOperation(operation string) ErrorBuilder {
	return NewErrorBuilder().WithOperation(operation)
}

// WithResource adds a resource identifier to the error builder
func (b *DefaultErrorBuilder) WithResource(resource string) ErrorBuilder {
	return NewErrorBuilder().WithResource(resource)
}

// WithStackTrace adds stack trace information to the error builder
func (b *DefaultErrorBuilder) WithStackTrace() ErrorBuilder {
	return NewErrorBuilder().WithStackTrace()
}

// Build creates a MageError from the error builder
func (b *DefaultErrorBuilder) Build() MageError {
	return NewErrorBuilder().Build()
}
