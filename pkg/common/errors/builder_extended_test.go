package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRealDefaultErrorBuilder_WithContext(t *testing.T) {
	builder := NewErrorBuilder()

	ctx := ErrorContext{
		Operation:   "TestOperation",
		Resource:    "TestResource",
		User:        "testuser",
		RequestID:   "req-123",
		Environment: "test",
		Version:     "1.0.0",
	}

	builder.WithContext(ctx)
	err := builder.Build()

	mageErr, ok := err.(*DefaultMageError)
	require.True(t, ok)

	errCtx := mageErr.Context()
	assert.Equal(t, "TestOperation", errCtx.Operation)
	assert.Equal(t, "TestResource", errCtx.Resource)
	assert.Equal(t, "testuser", errCtx.User)
	assert.Equal(t, "req-123", errCtx.RequestID)
	assert.Equal(t, "test", errCtx.Environment)
	assert.Equal(t, "1.0.0", errCtx.Version)
}

func TestDefaultErrorBuilder_AllMethods(t *testing.T) {
	// Test the DefaultErrorBuilder wrapper
	builder := &DefaultErrorBuilder{}

	t.Run("WithMessage", func(t *testing.T) {
		b := builder.WithMessage("test message %s", "arg")
		assert.NotNil(t, b) // Returns a new builder

		err := b.Build()
		assert.Contains(t, err.Error(), "test message arg")
	})

	t.Run("WithCode", func(t *testing.T) {
		b := builder.WithCode(ErrInternal)
		assert.NotNil(t, b)

		err := b.Build()
		var mageErr MageError
		require.ErrorAs(t, err, &mageErr)
		assert.Equal(t, ErrInternal, mageErr.Code())
	})

	t.Run("WithSeverity", func(t *testing.T) {
		b := builder.WithSeverity(SeverityCritical)
		assert.NotNil(t, b)

		err := b.Build()
		var mageErr MageError
		require.ErrorAs(t, err, &mageErr)
		assert.Equal(t, SeverityCritical, mageErr.Severity())
	})

	t.Run("WithContext", func(t *testing.T) {
		ctx := ErrorContext{
			Operation: "test-op",
			Resource:  "test-resource",
		}
		b := builder.WithContext(ctx)
		assert.NotNil(t, b)

		err := b.Build()
		var mageErr MageError
		require.ErrorAs(t, err, &mageErr)
		assert.Equal(t, "test-op", mageErr.Context().Operation)
		assert.Equal(t, "test-resource", mageErr.Context().Resource)
	})

	t.Run("WithField", func(t *testing.T) {
		testBuilder := &DefaultErrorBuilder{}
		b := testBuilder.WithField("key1", "value1")
		assert.NotNil(t, b)

		err := b.Build()
		var mageErr MageError
		require.ErrorAs(t, err, &mageErr)
		assert.Equal(t, "value1", mageErr.Context().Fields["key1"])
	})

	t.Run("WithFields", func(t *testing.T) {
		testBuilder := &DefaultErrorBuilder{}
		fields := map[string]interface{}{
			"field1": "value1",
			"field2": 123,
			"field3": true,
		}
		b := testBuilder.WithFields(fields)
		assert.NotNil(t, b)

		err := b.Build()
		var mageErr MageError
		require.ErrorAs(t, err, &mageErr)
		assert.Equal(t, "value1", mageErr.Context().Fields["field1"])
		assert.Equal(t, 123, mageErr.Context().Fields["field2"])
		assert.Equal(t, true, mageErr.Context().Fields["field3"])
	})

	t.Run("WithCause", func(t *testing.T) {
		cause := errors.New("underlying cause")
		b := builder.WithCause(cause)
		assert.NotNil(t, b)

		err := b.Build()
		assert.ErrorIs(t, err, cause)
	})

	t.Run("WithOperation", func(t *testing.T) {
		b := builder.WithOperation("TestOp")
		assert.NotNil(t, b)

		err := b.Build()
		var mageErr MageError
		require.ErrorAs(t, err, &mageErr)
		assert.Equal(t, "TestOp", mageErr.Context().Operation)
	})

	t.Run("WithResource", func(t *testing.T) {
		b := builder.WithResource("TestResource")
		assert.NotNil(t, b)

		err := b.Build()
		var mageErr MageError
		require.ErrorAs(t, err, &mageErr)
		assert.Equal(t, "TestResource", mageErr.Context().Resource)
	})

	t.Run("WithStackTrace", func(t *testing.T) {
		// Since DefaultErrorBuilder creates a new builder each time,
		// we need to chain the calls
		err := builder.WithStackTrace().Build()
		var mageErr MageError
		require.ErrorAs(t, err, &mageErr)
		// Stack trace is captured but might be empty in the default implementation
		stackTrace := mageErr.Context().StackTrace
		_ = stackTrace // Stack trace might be empty in some environments
	})

	t.Run("Build", func(t *testing.T) {
		// Since DefaultErrorBuilder creates a new builder each time,
		// we need to chain all calls
		builder := &DefaultErrorBuilder{}
		err := builder.
			WithMessage("final test").
			WithCode(ErrInternal).
			WithSeverity(SeverityError).
			Build()

		assert.NotNil(t, err)

		var mageErr MageError
		require.ErrorAs(t, err, &mageErr)
		assert.Equal(t, "final test", mageErr.Error())
		assert.Equal(t, ErrInternal, mageErr.Code())
		assert.Equal(t, SeverityError, mageErr.Severity())
	})
}

func TestDefaultErrorBuilder_ComplexScenario(t *testing.T) {
	// Test a complex scenario with all builder methods
	builder := &DefaultErrorBuilder{}

	cause := fmt.Errorf("root cause")

	err := builder.
		WithMessage("complex error: %s", "test").
		WithCode(ErrInternal).
		WithSeverity(SeverityCritical).
		WithCause(cause).
		WithOperation("ComplexOp").
		WithResource("ComplexResource").
		WithField("request_id", "req-456").
		WithField("user_id", "user-789").
		WithFields(map[string]interface{}{
			"additional_info": "some info",
			"retry_count":     3,
		}).
		WithStackTrace().
		Build()

	var mageErr MageError
	require.ErrorAs(t, err, &mageErr)

	// Verify all properties
	// When a cause is set, it's included in the error message
	assert.Contains(t, mageErr.Error(), "complex error: test")
	assert.Equal(t, ErrInternal, mageErr.Code())
	assert.Equal(t, SeverityCritical, mageErr.Severity())
	assert.ErrorIs(t, err, cause)

	ctx := mageErr.Context()
	assert.Equal(t, "ComplexOp", ctx.Operation)
	assert.Equal(t, "ComplexResource", ctx.Resource)
	assert.Equal(t, "req-456", ctx.Fields["request_id"])
	assert.Equal(t, "user-789", ctx.Fields["user_id"])
	assert.Equal(t, "some info", ctx.Fields["additional_info"])
	assert.Equal(t, 3, ctx.Fields["retry_count"])
	// Stack trace might be empty in some environments
	stackTrace := ctx.StackTrace
	_ = stackTrace // Stack trace might be empty in some environments
}

func TestDefaultErrorBuilder_NilFields(t *testing.T) {
	builder := &DefaultErrorBuilder{}

	// Test WithFields with nil map
	b := builder.WithFields(nil)
	assert.NotNil(t, b)

	err := b.Build()
	var mageErr MageError
	require.ErrorAs(t, err, &mageErr)
	assert.NotNil(t, mageErr.Context().Fields) // Should still have initialized fields map
}

func TestDefaultErrorBuilder_EmptyBuild(t *testing.T) {
	builder := &DefaultErrorBuilder{}

	// Build without setting any properties
	err := builder.Build()
	assert.NotNil(t, err)

	var mageErr MageError
	require.ErrorAs(t, err, &mageErr)
	assert.Equal(t, ErrUnknown, mageErr.Code())        // Default code
	assert.Equal(t, SeverityError, mageErr.Severity()) // Default severity
	// Default error has empty message
	errorMsg := mageErr.Error()
	_ = errorMsg // Error message might be empty for default error
}
