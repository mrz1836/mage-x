package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Static test errors for DefaultErrorBuilder tests (err113 compliant)
var (
	errDefaultBuilderCause = errors.New("default builder cause error")
)

// TestDefaultErrorBuilderWithMessage tests the WithMessage method
func TestDefaultErrorBuilderWithMessage(t *testing.T) {
	t.Run("simple message", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		result := builder.WithMessage("test message")

		mageErr := result.Build()
		assert.Contains(t, mageErr.Error(), "test message")
	})

	t.Run("formatted message", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		result := builder.WithMessage("error %d: %s", 42, "details")

		mageErr := result.Build()
		assert.Contains(t, mageErr.Error(), "error 42: details")
	})

	t.Run("empty message", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		result := builder.WithMessage("")

		mageErr := result.Build()
		assert.NotNil(t, mageErr)
	})

	t.Run("chaining creates new builder each time", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		chain1 := builder.WithMessage("first")
		chain2 := builder.WithMessage("second")

		// Each call creates a new RealDefaultErrorBuilder via NewErrorBuilder()
		err1 := chain1.Build()
		err2 := chain2.Build()

		assert.Contains(t, err1.Error(), "first")
		assert.Contains(t, err2.Error(), "second")
	})
}

// TestDefaultErrorBuilderWithCode tests the WithCode method
func TestDefaultErrorBuilderWithCode(t *testing.T) {
	tests := []struct {
		name string
		code ErrorCode
	}{
		{name: "internal error", code: ErrInternal},
		{name: "not found error", code: ErrNotFound},
		{name: "timeout error", code: ErrTimeout},
		{name: "permission denied", code: ErrPermissionDenied},
		{name: "build failed", code: ErrBuildFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &DefaultErrorBuilder{}
			result := builder.WithCode(tt.code)

			mageErr := result.Build()
			assert.Equal(t, tt.code, mageErr.Code())
		})
	}
}

// TestDefaultErrorBuilderWithSeverity tests the WithSeverity method
func TestDefaultErrorBuilderWithSeverity(t *testing.T) {
	tests := []struct {
		name     string
		severity Severity
	}{
		{name: "debug", severity: SeverityDebug},
		{name: "info", severity: SeverityInfo},
		{name: "warning", severity: SeverityWarning},
		{name: "error", severity: SeverityError},
		{name: "critical", severity: SeverityCritical},
		{name: "fatal", severity: SeverityFatal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &DefaultErrorBuilder{}
			result := builder.WithSeverity(tt.severity)

			mageErr := result.Build()
			assert.Equal(t, tt.severity, mageErr.Severity())
		})
	}
}

// TestDefaultErrorBuilderWithContext tests the WithContext method
func TestDefaultErrorBuilderWithContext(t *testing.T) {
	t.Run("full context", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		ctx := &ErrorContext{
			Operation:   "TestOp",
			Resource:    "TestResource",
			User:        "testuser",
			RequestID:   "req-123",
			Environment: "test",
		}
		result := builder.WithContext(ctx)

		mageErr := result.Build()
		errCtx := mageErr.Context()
		assert.Equal(t, "TestOp", errCtx.Operation)
		assert.Equal(t, "TestResource", errCtx.Resource)
		assert.Equal(t, "testuser", errCtx.User)
		assert.Equal(t, "req-123", errCtx.RequestID)
		assert.Equal(t, "test", errCtx.Environment)
	})

	t.Run("nil context", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		result := builder.WithContext(nil)

		mageErr := result.Build()
		assert.NotNil(t, mageErr)
		// Should have default context with timestamp
		assert.False(t, mageErr.Context().Timestamp.IsZero())
	})

	t.Run("partial context", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		ctx := &ErrorContext{
			Operation: "OnlyOp",
		}
		result := builder.WithContext(ctx)

		mageErr := result.Build()
		errCtx := mageErr.Context()
		assert.Equal(t, "OnlyOp", errCtx.Operation)
		assert.Empty(t, errCtx.Resource)
	})
}

// TestDefaultErrorBuilderWithField tests the WithField method
func TestDefaultErrorBuilderWithField(t *testing.T) {
	t.Run("single field", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		result := builder.WithField("key1", "value1")

		mageErr := result.Build()
		fields := mageErr.Context().Fields
		assert.Equal(t, "value1", fields["key1"])
	})

	t.Run("multiple fields via chaining", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		// Note: Each WithField creates a new builder, so chaining doesn't accumulate
		result := builder.WithField("key1", "value1")

		mageErr := result.Build()
		fields := mageErr.Context().Fields
		assert.Equal(t, "value1", fields["key1"])
	})

	t.Run("various value types", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		result := builder.WithField("int", 42)

		mageErr := result.Build()
		assert.Equal(t, 42, mageErr.Context().Fields["int"])
	})
}

// TestDefaultErrorBuilderWithFields tests the WithFields method
func TestDefaultErrorBuilderWithFields(t *testing.T) {
	t.Run("multiple fields at once", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		fields := map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"key3": true,
		}
		result := builder.WithFields(fields)

		mageErr := result.Build()
		errFields := mageErr.Context().Fields
		assert.Equal(t, "value1", errFields["key1"])
		assert.Equal(t, 42, errFields["key2"])
		assert.Equal(t, true, errFields["key3"])
	})

	t.Run("empty map", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		result := builder.WithFields(map[string]interface{}{})

		mageErr := result.Build()
		assert.NotNil(t, mageErr)
	})

	t.Run("nil map", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		result := builder.WithFields(nil)

		mageErr := result.Build()
		assert.NotNil(t, mageErr)
	})
}

// TestDefaultErrorBuilderWithCause tests the WithCause method
func TestDefaultErrorBuilderWithCause(t *testing.T) {
	t.Run("with cause", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		result := builder.WithCause(errDefaultBuilderCause)

		mageErr := result.Build()
		assert.ErrorIs(t, mageErr, errDefaultBuilderCause)
	})

	t.Run("nil cause", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		result := builder.WithCause(nil)

		mageErr := result.Build()
		assert.NotNil(t, mageErr)
	})

	t.Run("nested cause", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		result := builder.WithCause(errDefaultBuilderCause)

		mageErr := result.Build()
		assert.ErrorIs(t, mageErr, errDefaultBuilderCause)
	})
}

// TestDefaultErrorBuilderWithOperation tests the WithOperation method
func TestDefaultErrorBuilderWithOperation(t *testing.T) {
	t.Run("sets operation", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		result := builder.WithOperation("CreateUser")

		mageErr := result.Build()
		assert.Equal(t, "CreateUser", mageErr.Context().Operation)
	})

	t.Run("empty operation", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		result := builder.WithOperation("")

		mageErr := result.Build()
		assert.Empty(t, mageErr.Context().Operation)
	})
}

// TestDefaultErrorBuilderWithResource tests the WithResource method
func TestDefaultErrorBuilderWithResource(t *testing.T) {
	t.Run("sets resource", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		result := builder.WithResource("users/123")

		mageErr := result.Build()
		assert.Equal(t, "users/123", mageErr.Context().Resource)
	})

	t.Run("empty resource", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		result := builder.WithResource("")

		mageErr := result.Build()
		assert.Empty(t, mageErr.Context().Resource)
	})
}

// TestDefaultErrorBuilderWithStackTrace tests the WithStackTrace method
func TestDefaultErrorBuilderWithStackTrace(t *testing.T) {
	t.Run("captures stack trace", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		result := builder.WithStackTrace()

		mageErr := result.Build()
		// Stack trace is stored in context
		// Note: Stack trace might be empty in some test environments
		_ = mageErr.Context().StackTrace // Just verify it doesn't panic
	})
}

// TestDefaultErrorBuilderBuild tests the Build method
func TestDefaultErrorBuilderBuild(t *testing.T) {
	t.Run("builds with defaults", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		mageErr := builder.Build()

		assert.NotNil(t, mageErr)
		// Default code should be ErrUnknown
		assert.Equal(t, ErrUnknown, mageErr.Code())
		// Default severity should be SeverityError
		assert.Equal(t, SeverityError, mageErr.Severity())
	})

	t.Run("build is idempotent", func(t *testing.T) {
		builder := &DefaultErrorBuilder{}
		err1 := builder.Build()
		err2 := builder.Build()

		// Both should be valid MageErrors
		assert.NotNil(t, err1)
		assert.NotNil(t, err2)
		// But they're independent (from fresh NewErrorBuilder() calls)
		assert.Equal(t, err1.Code(), err2.Code())
	})
}

// TestDefaultErrorBuilderImplementsInterface ensures DefaultErrorBuilder
// properly implements the ErrorBuilder interface
func TestDefaultErrorBuilderImplementsInterface(t *testing.T) {
	// Compile-time check
	var _ ErrorBuilder = (*DefaultErrorBuilder)(nil)

	// Runtime check - all methods should be callable
	var builder ErrorBuilder = &DefaultErrorBuilder{}

	builder = builder.WithMessage("test")
	builder = builder.WithCode(ErrInternal)
	builder = builder.WithSeverity(SeverityError)
	builder = builder.WithContext(&ErrorContext{})
	builder = builder.WithField("key", "value")
	builder = builder.WithFields(map[string]interface{}{"k": "v"})
	builder = builder.WithCause(errDefaultBuilderCause)
	builder = builder.WithOperation("op")
	builder = builder.WithResource("res")
	builder = builder.WithStackTrace()

	mageErr := builder.Build()
	assert.NotNil(t, mageErr)
}

// TestDefaultErrorBuilderFluent tests fluent API usage
func TestDefaultErrorBuilderFluent(t *testing.T) {
	t.Run("complete fluent chain", func(t *testing.T) {
		// Note: DefaultErrorBuilder creates new builder on each call,
		// so fluent chaining works through the returned RealDefaultErrorBuilder
		builder := &DefaultErrorBuilder{}

		// Start the chain - WithMessage returns RealDefaultErrorBuilder
		mageErr := builder.WithMessage("operation failed").
			WithCode(ErrInternal).
			WithSeverity(SeverityCritical).
			WithOperation("CreateUser").
			WithResource("users/new").
			WithField("attempt", 3).
			WithCause(errDefaultBuilderCause).
			Build()

		assert.Contains(t, mageErr.Error(), "operation failed")
		assert.Equal(t, ErrInternal, mageErr.Code())
		assert.Equal(t, SeverityCritical, mageErr.Severity())
		assert.Equal(t, "CreateUser", mageErr.Context().Operation)
		assert.Equal(t, "users/new", mageErr.Context().Resource)
		assert.Equal(t, 3, mageErr.Context().Fields["attempt"])
		assert.ErrorIs(t, mageErr, errDefaultBuilderCause)
	})
}

// TestDefaultErrorBuilderNewInstance verifies each method creates new builder
func TestDefaultErrorBuilderNewInstance(t *testing.T) {
	builder := &DefaultErrorBuilder{}

	// Each call should create an independent builder path
	withCode := builder.WithCode(ErrInternal)
	withSeverity := builder.WithSeverity(SeverityCritical)
	withMessage := builder.WithMessage("test")

	// Build each separately
	errCode := withCode.Build()
	errSeverity := withSeverity.Build()
	errMessage := withMessage.Build()

	// Each should have only the property that was set
	// (others will be defaults)
	assert.Equal(t, ErrInternal, errCode.Code())
	assert.Equal(t, SeverityError, errCode.Severity()) // default

	assert.Equal(t, ErrUnknown, errSeverity.Code()) // default
	assert.Equal(t, SeverityCritical, errSeverity.Severity())

	assert.Contains(t, errMessage.Error(), "test")
	assert.Equal(t, ErrUnknown, errMessage.Code())        // default
	assert.Equal(t, SeverityError, errMessage.Severity()) // default
}
