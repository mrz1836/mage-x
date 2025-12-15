package errors

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Static error variables for tests (err113 compliance)
var (
	errHandlerStandard = errors.New("standard error")
	errHandlerError    = errors.New("handler error")
	errSeverityHandler = errors.New("severity handler error")
	errDefaultHandler  = errors.New("default handler error")
	errFallback        = errors.New("fallback error")
)

// TestRealHandler_NilError verifies that handling nil error returns nil
func TestRealHandler_NilError(t *testing.T) {
	t.Parallel()

	handler := NewErrorHandler()
	result := handler.Handle(nil)
	assert.NoError(t, result, "Handle(nil) should return nil")
}

// TestRealHandler_CodeHandlerPriority verifies code handlers run before severity handlers
func TestRealHandler_CodeHandlerPriority(t *testing.T) {
	t.Parallel()

	var callOrder []string

	handler := NewErrorHandler()
	handler.OnError(ErrBuildFailed, func(_ MageError) error {
		callOrder = append(callOrder, "code")
		return nil
	})
	handler.OnSeverity(SeverityError, func(_ MageError) error {
		callOrder = append(callOrder, "severity")
		return nil
	})

	err := NewBuilder().
		WithMessage("build failed").
		WithCode(ErrBuildFailed).
		WithSeverity(SeverityError).
		Build()

	result := handler.Handle(err)
	require.NoError(t, result)

	assert.Equal(t, []string{"code"}, callOrder, "Code handler should be called first and exclusively")
}

// TestRealHandler_SeverityHandlerFallback verifies severity handler runs when no code handler matches
func TestRealHandler_SeverityHandlerFallback(t *testing.T) {
	t.Parallel()

	severityHandled := false

	handler := NewErrorHandler()
	handler.OnError(ErrBuildFailed, func(_ MageError) error {
		t.Error("Code handler should not be called")
		return nil
	})
	handler.OnSeverity(SeverityCritical, func(_ MageError) error {
		severityHandled = true
		return nil
	})

	// Use a different error code that doesn't have a handler
	err := NewBuilder().
		WithMessage("critical error").
		WithCode(ErrTestFailed). // No handler for this code
		WithSeverity(SeverityCritical).
		Build()

	result := handler.Handle(err)
	require.NoError(t, result)
	assert.True(t, severityHandled, "Severity handler should be called when no code handler matches")
}

// TestRealHandler_DefaultHandlerForNonMageError verifies default handler is used for standard errors
func TestRealHandler_DefaultHandlerForNonMageError(t *testing.T) {
	t.Parallel()

	defaultHandled := false

	handler := NewErrorHandler()
	handler.SetDefault(func(_ error) error {
		defaultHandled = true
		return nil
	})

	result := handler.Handle(errHandlerStandard)

	require.NoError(t, result)
	assert.True(t, defaultHandled, "Default handler should be called for non-MageError")
}

// TestRealHandler_FallbackOnHandlerError verifies fallback is called when handler returns error
func TestRealHandler_FallbackOnHandlerError(t *testing.T) {
	t.Parallel()

	fallbackCalled := false

	handler := NewErrorHandler()
	handler.OnError(ErrBuildFailed, func(_ MageError) error {
		return errHandlerError
	})
	handler.SetFallback(func(err error) error {
		fallbackCalled = true
		assert.Equal(t, errHandlerError, err, "Fallback should receive handler's error")
		return nil
	})

	err := WithCode(ErrBuildFailed, "build failed")
	result := handler.Handle(err)

	require.NoError(t, result)
	assert.True(t, fallbackCalled, "Fallback should be called when handler returns error")
}

// TestRealHandler_NoHandlerReturnsOriginal verifies original error returned when no handlers match
func TestRealHandler_NoHandlerReturnsOriginal(t *testing.T) {
	t.Parallel()

	handler := NewErrorHandler()
	// Don't register any handlers

	err := WithCode(ErrBuildFailed, "build failed")
	result := handler.Handle(err)

	assert.Equal(t, err, result, "Original error should be returned when no handlers match")
}

// TestRealHandler_FallbackWithNoHandlers verifies fallback works when no specific handlers exist
func TestRealHandler_FallbackWithNoHandlers(t *testing.T) {
	t.Parallel()

	fallbackCalled := false

	handler := NewErrorHandler()
	handler.SetFallback(func(err error) error {
		fallbackCalled = true
		return nil
	})

	err := WithCode(ErrBuildFailed, "build failed")
	result := handler.Handle(err)

	require.NoError(t, result)
	assert.True(t, fallbackCalled, "Fallback should be called when no handlers match")
}

// TestRealHandler_WithContext_ExtractsRequestID verifies context values are extracted
func TestRealHandler_WithContext_ExtractsRequestID(t *testing.T) {
	t.Parallel()

	var capturedErr MageError

	handler := NewErrorHandler()
	handler.OnError(ErrBuildFailed, func(err MageError) error {
		capturedErr = err
		return nil
	})

	ctx := context.WithValue(context.Background(), "requestID", "req-123") //nolint:staticcheck // handler expects string keys
	ctx = context.WithValue(ctx, "userID", "user-456")                     //nolint:staticcheck // handler expects string keys

	err := WithCode(ErrBuildFailed, "build failed")
	result := handler.HandleWithContext(ctx, err)

	require.NoError(t, result)
	require.NotNil(t, capturedErr)

	// Check that context values were extracted into fields
	fields := capturedErr.Context().Fields
	assert.Equal(t, "req-123", fields["requestID"], "requestID should be extracted from context")
	assert.Equal(t, "user-456", fields["userID"], "userID should be extracted from context")
}

// TestRealHandler_WithContext_CanceledContext verifies canceled context returns ctx.Err()
func TestRealHandler_WithContext_CanceledContext(t *testing.T) {
	t.Parallel()

	handler := NewErrorHandler()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := WithCode(ErrBuildFailed, "build failed")
	result := handler.HandleWithContext(ctx, err)

	assert.ErrorIs(t, result, context.Canceled, "HandleWithContext should return context.Canceled")
}

// TestRealHandler_WithContext_NilError verifies HandleWithContext(ctx, nil) returns nil
func TestRealHandler_WithContext_NilError(t *testing.T) {
	t.Parallel()

	handler := NewErrorHandler()
	ctx := context.Background()

	result := handler.HandleWithContext(ctx, nil)
	assert.NoError(t, result, "HandleWithContext(ctx, nil) should return nil")
}

// TestRealHandler_Concurrent verifies concurrent handler calls are safe
func TestRealHandler_Concurrent(t *testing.T) {
	t.Parallel()

	var callCount atomic.Int64

	handler := NewErrorHandler()
	handler.OnError(ErrBuildFailed, func(_ MageError) error {
		callCount.Add(1)
		return nil
	})

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			err := WithCode(ErrBuildFailed, "build failed")
			_ = handler.Handle(err) //nolint:errcheck // testing concurrent calls, error not relevant
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(numGoroutines), callCount.Load(), "All handlers should be called")
}

// TestRealHandler_ChainedRegistration verifies fluent API returns same handler
func TestRealHandler_ChainedRegistration(t *testing.T) {
	t.Parallel()

	handler := NewErrorHandler()

	// Chain multiple registrations
	result := handler.
		OnError(ErrBuildFailed, func(_ MageError) error { return nil }).
		OnError(ErrTestFailed, func(_ MageError) error { return nil }).
		OnSeverity(SeverityCritical, func(_ MageError) error { return nil }).
		SetDefault(func(_ error) error { return nil }).
		SetFallback(func(_ error) error { return nil })

	assert.Same(t, handler, result, "Chained methods should return same handler")
}

// TestRealHandler_MultipleCodeHandlers verifies only registered code handler is called
func TestRealHandler_MultipleCodeHandlers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		errorCode    ErrorCode
		expectedCall string
	}{
		{
			name:         "calls ErrBuildFailed handler",
			errorCode:    ErrBuildFailed,
			expectedCall: "build",
		},
		{
			name:         "calls ErrTestFailed handler",
			errorCode:    ErrTestFailed,
			expectedCall: "test",
		},
		{
			name:         "calls ErrConfigInvalid handler",
			errorCode:    ErrConfigInvalid,
			expectedCall: "config",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var called string

			handler := NewErrorHandler()
			handler.OnError(ErrBuildFailed, func(_ MageError) error {
				called = "build"
				return nil
			})
			handler.OnError(ErrTestFailed, func(_ MageError) error {
				called = "test"
				return nil
			})
			handler.OnError(ErrConfigInvalid, func(_ MageError) error {
				called = "config"
				return nil
			})

			err := WithCode(tt.errorCode, "error message")
			_ = handler.Handle(err) //nolint:errcheck // testing handler selection, error not relevant

			assert.Equal(t, tt.expectedCall, called, "Expected handler should be called")
		})
	}
}

// TestRealHandler_MultipleSeverityHandlers verifies correct severity handler is called
func TestRealHandler_MultipleSeverityHandlers(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		severity     Severity
		expectedCall string
	}{
		{
			name:         "calls Warning handler",
			severity:     SeverityWarning,
			expectedCall: "warning",
		},
		{
			name:         "calls Error handler",
			severity:     SeverityError,
			expectedCall: "error",
		},
		{
			name:         "calls Critical handler",
			severity:     SeverityCritical,
			expectedCall: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var called string

			handler := NewErrorHandler()
			handler.OnSeverity(SeverityWarning, func(_ MageError) error {
				called = "warning"
				return nil
			})
			handler.OnSeverity(SeverityError, func(_ MageError) error {
				called = "error"
				return nil
			})
			handler.OnSeverity(SeverityCritical, func(_ MageError) error {
				called = "critical"
				return nil
			})

			err := NewBuilder().
				WithMessage("error").
				WithCode("UNREGISTERED_CODE"). // Use code without handler
				WithSeverity(tt.severity).
				Build()

			_ = handler.Handle(err) //nolint:errcheck // testing handler selection, error not relevant

			assert.Equal(t, tt.expectedCall, called, "Expected severity handler should be called")
		})
	}
}

// TestRealHandler_DefaultHandlerAfterMageErrorNoMatch verifies default handler for MageError without matching handlers
func TestRealHandler_DefaultHandlerAfterMageErrorNoMatch(t *testing.T) {
	t.Parallel()

	defaultCalled := false

	handler := NewErrorHandler()
	handler.OnError(ErrBuildFailed, func(_ MageError) error {
		t.Error("Code handler should not be called")
		return nil
	})
	handler.SetDefault(func(_ error) error {
		defaultCalled = true
		return nil
	})

	// Use error code that doesn't have a handler
	err := WithCode(ErrTestFailed, "test failed")
	_ = handler.Handle(err) //nolint:errcheck // testing handler selection, error not relevant

	assert.True(t, defaultCalled, "Default handler should be called when no specific handler matches")
}

// TestRealHandler_SeverityHandlerError verifies fallback is called on severity handler error
func TestRealHandler_SeverityHandlerError(t *testing.T) {
	t.Parallel()

	fallbackCalled := false

	handler := NewErrorHandler()
	handler.OnSeverity(SeverityError, func(_ MageError) error {
		return errSeverityHandler
	})
	handler.SetFallback(func(err error) error {
		fallbackCalled = true
		assert.Equal(t, errSeverityHandler, err)
		return nil
	})

	err := NewBuilder().
		WithMessage("error").
		WithCode("UNREGISTERED").
		WithSeverity(SeverityError).
		Build()

	_ = handler.Handle(err) //nolint:errcheck // testing fallback, error not relevant

	assert.True(t, fallbackCalled, "Fallback should be called on severity handler error")
}

// TestRealHandler_DefaultHandlerError verifies fallback is called on default handler error
func TestRealHandler_DefaultHandlerError(t *testing.T) {
	t.Parallel()

	fallbackCalled := false

	handler := NewErrorHandler()
	handler.SetDefault(func(_ error) error {
		return errDefaultHandler
	})
	handler.SetFallback(func(err error) error {
		fallbackCalled = true
		assert.Equal(t, errDefaultHandler, err)
		return nil
	})

	_ = handler.Handle(errHandlerStandard) //nolint:errcheck // testing fallback, error not relevant

	assert.True(t, fallbackCalled, "Fallback should be called on default handler error")
}

// TestRealHandler_FallbackReturnsError verifies fallback error is returned
func TestRealHandler_FallbackReturnsError(t *testing.T) {
	t.Parallel()

	handler := NewErrorHandler()
	handler.SetFallback(func(_ error) error {
		return errFallback
	})

	err := WithCode(ErrBuildFailed, "build failed")
	result := handler.Handle(err)

	assert.Equal(t, errFallback, result, "Fallback error should be returned")
}

// TestRealHandler_ConcurrentRegistrationAndHandling verifies thread safety during registration
func TestRealHandler_ConcurrentRegistrationAndHandling(t *testing.T) {
	t.Parallel()

	handler := NewErrorHandler()
	var wg sync.WaitGroup

	// Concurrent registrations
	for i := 0; i < 10; i++ {
		wg.Add(1)
		code := ErrorCode("CODE_" + string(rune('A'+i)))
		go func(c ErrorCode) {
			defer wg.Done()
			handler.OnError(c, func(_ MageError) error { return nil })
		}(code)
	}

	// Concurrent handling
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := WithCode(ErrBuildFailed, "test")
			_ = handler.Handle(err) //nolint:errcheck // testing concurrent calls, error not relevant
		}()
	}

	wg.Wait()
	// Test passes if no race conditions occur
}
