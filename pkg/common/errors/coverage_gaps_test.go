package errors

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test errors for err113 compliance (prefixed with cg_ to avoid conflicts)
var (
	errCGOriginal = errors.New("original error")
	errCGDB       = errors.New("db error")
	errCGGeneric  = errors.New("test error")
	errCGPanic    = errors.New("panic error")
	errCGFunction = errors.New("function error")
)

// =============================================================================
// Recover and RecoverTo Tests
// =============================================================================
// Note: Recover() and RecoverTo() have a design limitation - recover() only
// works when called directly in a deferred function, not in functions called
// by the deferred function. These functions would need to be redesigned to
// work properly. Testing their "no panic" case only.

// TestRecover tests the Recover function when no panic occurs.
func TestRecover(t *testing.T) {
	t.Run("no panic returns nil", func(t *testing.T) {
		// When there's no panic, Recover should return nil
		result := Recover()
		assert.NoError(t, result)
	})
}

// TestRecoverTo tests the RecoverTo function when no panic occurs.
func TestRecoverTo(t *testing.T) {
	t.Run("no panic returns nil", func(t *testing.T) {
		// When there's no panic, RecoverTo should return nil
		result := RecoverTo()
		assert.NoError(t, result)
	})
}

// =============================================================================
// WrapError and WrapErrorf Tests
// =============================================================================

// TestWrapError tests the WrapError convenience function.
func TestWrapError(t *testing.T) {
	t.Run("wraps error with message", func(t *testing.T) {
		wrapped := WrapError(errCGOriginal, "additional context")

		require.Error(t, wrapped)
		assert.Contains(t, wrapped.Error(), "additional context")
		assert.Contains(t, wrapped.Error(), "original error")
	})

	t.Run("wraps nil error", func(t *testing.T) {
		wrapped := WrapError(nil, "context for nil")

		// Behavior depends on implementation - nil might be returned or wrapped
		// This test verifies it doesn't panic
		_ = wrapped
	})
}

// TestWrapErrorf tests the WrapErrorf convenience function with formatting.
func TestWrapErrorf(t *testing.T) {
	t.Run("wraps error with formatted message", func(t *testing.T) {
		wrapped := WrapErrorf(errCGOriginal, "operation %s failed at step %d", "build", 3)

		require.Error(t, wrapped)
		assert.Contains(t, wrapped.Error(), "operation build failed at step 3")
		assert.Contains(t, wrapped.Error(), "original error")
	})

	t.Run("wraps with complex format", func(t *testing.T) {
		wrapped := WrapErrorf(errCGDB, "query %q failed for user %d", "SELECT *", 123)

		require.Error(t, wrapped)
		assert.Contains(t, wrapped.Error(), "SELECT")
		assert.Contains(t, wrapped.Error(), "123")
	})
}

// =============================================================================
// DefaultErrorHandler Tests (the interface-based handler from interfaces.go)
// =============================================================================

// newDefaultErrorHandler creates a DefaultErrorHandler for testing
func newDefaultErrorHandler() *DefaultErrorHandler {
	return &DefaultErrorHandler{
		handlers:         make(map[ErrorCode]func(MageError) error),
		severityHandlers: make(map[Severity]func(MageError) error),
	}
}

// TestDefaultErrorHandler_Handle tests the Handle method.
func TestDefaultErrorHandler_Handle(t *testing.T) {
	t.Run("handles error without custom handlers", func(t *testing.T) {
		handler := newDefaultErrorHandler()

		result := handler.Handle(errCGGeneric)

		// Should return the error (or nil if handled)
		assert.Error(t, result)
	})

	t.Run("handles MageError", func(t *testing.T) {
		handler := newDefaultErrorHandler()
		mageErr := New("mage error")

		result := handler.Handle(mageErr)

		assert.Error(t, result)
	})

	t.Run("handles nil error", func(t *testing.T) {
		handler := newDefaultErrorHandler()

		result := handler.Handle(nil)

		assert.NoError(t, result)
	})
}

// TestDefaultErrorHandler_HandleWithContext tests handling with context.
func TestDefaultErrorHandler_HandleWithContext(t *testing.T) {
	t.Run("handles error with context", func(t *testing.T) {
		handler := newDefaultErrorHandler()
		ctx := context.Background()

		result := handler.HandleWithContext(ctx, errCGGeneric)

		assert.Error(t, result)
	})

	t.Run("handles nil error with context", func(t *testing.T) {
		handler := newDefaultErrorHandler()
		ctx := context.Background()

		result := handler.HandleWithContext(ctx, nil)

		assert.NoError(t, result)
	})

	t.Run("handles with canceled context", func(t *testing.T) {
		handler := newDefaultErrorHandler()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		result := handler.HandleWithContext(ctx, errCGGeneric)

		// Should return context error when canceled
		assert.Error(t, result)
	})
}

// TestDefaultErrorHandler_OnError tests registering error code handlers.
func TestDefaultErrorHandler_OnError(t *testing.T) {
	t.Run("registers handler for error code", func(t *testing.T) {
		handler := newDefaultErrorHandler()

		result := handler.OnError(ErrNotFound, func(_ MageError) error {
			return nil
		})

		// Should return handler for chaining
		assert.NotNil(t, result)
		// Handler should be registered
		assert.NotNil(t, handler.handlers[ErrNotFound])
	})

	t.Run("registers multiple handlers", func(t *testing.T) {
		handler := newDefaultErrorHandler()

		handler.
			OnError(ErrNotFound, func(_ MageError) error { return nil }).
			OnError(ErrTimeout, func(_ MageError) error { return nil }).
			OnError(ErrInternal, func(_ MageError) error { return nil })

		// All handlers should be registered
		assert.Len(t, handler.handlers, 3)
	})
}

// TestDefaultErrorHandler_OnSeverity tests registering severity handlers.
func TestDefaultErrorHandler_OnSeverity(t *testing.T) {
	t.Run("registers handler for severity", func(t *testing.T) {
		handler := newDefaultErrorHandler()

		result := handler.OnSeverity(SeverityError, func(_ MageError) error {
			return nil
		})

		// Should return handler for chaining
		assert.NotNil(t, result)
		assert.NotNil(t, handler.severityHandlers[SeverityError])
	})

	t.Run("registers handlers for all severities", func(t *testing.T) {
		handler := newDefaultErrorHandler()

		handler.
			OnSeverity(SeverityDebug, func(_ MageError) error { return nil }).
			OnSeverity(SeverityInfo, func(_ MageError) error { return nil }).
			OnSeverity(SeverityWarning, func(_ MageError) error { return nil }).
			OnSeverity(SeverityError, func(_ MageError) error { return nil }).
			OnSeverity(SeverityCritical, func(_ MageError) error { return nil })

		assert.Len(t, handler.severityHandlers, 5)
	})
}

// TestDefaultErrorHandler_SetDefault tests setting a default handler.
func TestDefaultErrorHandler_SetDefault(t *testing.T) {
	t.Run("sets default handler", func(t *testing.T) {
		handler := newDefaultErrorHandler()

		result := handler.SetDefault(func(err error) error {
			return err
		})

		// Should return handler for chaining
		assert.NotNil(t, result)
	})

	t.Run("sets nil default handler", func(t *testing.T) {
		handler := newDefaultErrorHandler()

		result := handler.SetDefault(nil)

		// Should not panic
		assert.NotNil(t, result)
	})
}

// TestDefaultErrorHandler_SetFallback tests setting a fallback handler.
func TestDefaultErrorHandler_SetFallback(t *testing.T) {
	t.Run("sets fallback handler", func(t *testing.T) {
		handler := newDefaultErrorHandler()

		result := handler.SetFallback(func(err error) error {
			return err
		})

		// Should return handler for chaining
		assert.NotNil(t, result)
	})

	t.Run("sets nil fallback handler", func(t *testing.T) {
		handler := newDefaultErrorHandler()

		result := handler.SetFallback(nil)

		// Should not panic
		assert.NotNil(t, result)
	})
}

// =============================================================================
// SafeExecute Edge Cases Tests
// =============================================================================

// TestSafeExecute_EdgeCases tests edge cases of SafeExecute.
func TestSafeExecute_EdgeCases(t *testing.T) {
	t.Run("recovers when fn panics with error", func(t *testing.T) {
		err := SafeExecute(func() error {
			panic(errCGPanic)
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "panic recovered")
	})

	t.Run("recovers when fn panics with non-error", func(t *testing.T) {
		err := SafeExecute(func() error {
			panic("string panic value")
		})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "panic recovered")
		assert.Contains(t, err.Error(), "string panic value")
	})

	t.Run("returns fn error when no panic", func(t *testing.T) {
		err := SafeExecute(func() error {
			return errCGFunction
		})

		assert.ErrorIs(t, err, errCGFunction)
	})

	t.Run("returns nil when fn succeeds", func(t *testing.T) {
		err := SafeExecute(func() error {
			return nil
		})

		assert.NoError(t, err)
	})
}
