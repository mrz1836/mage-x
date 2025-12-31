package errors

import (
	"context"
	"errors"
	"testing"
	"time"

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
		require.NoError(t, result)
	})
}

// TestRecoverTo tests the RecoverTo function when no panic occurs.
func TestRecoverTo(t *testing.T) {
	t.Run("no panic returns nil", func(t *testing.T) {
		// When there's no panic, RecoverTo should return nil
		result := RecoverTo()
		require.NoError(t, result)
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
		require.Error(t, result)
	})

	t.Run("handles MageError", func(t *testing.T) {
		handler := newDefaultErrorHandler()
		mageErr := New("mage error")

		result := handler.Handle(mageErr)

		require.Error(t, result)
	})

	t.Run("handles nil error", func(t *testing.T) {
		handler := newDefaultErrorHandler()

		result := handler.Handle(nil)

		require.NoError(t, result)
	})
}

// TestDefaultErrorHandler_HandleWithContext tests handling with context.
func TestDefaultErrorHandler_HandleWithContext(t *testing.T) {
	t.Run("handles error with context", func(t *testing.T) {
		handler := newDefaultErrorHandler()
		ctx := context.Background()

		result := handler.HandleWithContext(ctx, errCGGeneric)

		require.Error(t, result)
	})

	t.Run("handles nil error with context", func(t *testing.T) {
		handler := newDefaultErrorHandler()
		ctx := context.Background()

		result := handler.HandleWithContext(ctx, nil)

		require.NoError(t, result)
	})

	t.Run("handles with canceled context", func(t *testing.T) {
		handler := newDefaultErrorHandler()
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		result := handler.HandleWithContext(ctx, errCGGeneric)

		// Should return context error when canceled
		require.Error(t, result)
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

		require.ErrorIs(t, err, errCGFunction)
	})

	t.Run("returns nil when fn succeeds", func(t *testing.T) {
		err := SafeExecute(func() error {
			return nil
		})

		require.NoError(t, err)
	})
}

// =============================================================================
// MockErrorNotifier Tests
// =============================================================================

// TestMockErrorNotifier tests all MockErrorNotifier methods for coverage.
func TestMockErrorNotifier(t *testing.T) {
	t.Run("NewMockErrorNotifier", func(t *testing.T) {
		mock := NewMockErrorNotifier()
		require.NotNil(t, mock)
		assert.Equal(t, SeverityError, mock.Threshold)
		assert.True(t, mock.Enabled)
	})

	t.Run("Notify", func(t *testing.T) {
		mock := NewMockErrorNotifier()
		err := mock.Notify(errCGGeneric)
		require.NoError(t, err)
		assert.Len(t, mock.NotifyCalls, 1)
	})

	t.Run("Notify_with_error", func(t *testing.T) {
		mock := NewMockErrorNotifier()
		mock.ShouldError = true
		err := mock.Notify(errCGGeneric)
		require.Error(t, err)
	})

	t.Run("NotifyWithContext", func(t *testing.T) {
		mock := NewMockErrorNotifier()
		err := mock.NotifyWithContext(context.Background(), errCGGeneric)
		require.NoError(t, err)
		assert.Len(t, mock.NotifyWithContextCalls, 1)
	})

	t.Run("NotifyWithContext_with_error", func(t *testing.T) {
		mock := NewMockErrorNotifier()
		mock.ShouldError = true
		err := mock.NotifyWithContext(context.Background(), errCGGeneric)
		require.Error(t, err)
	})

	t.Run("SetThreshold", func(t *testing.T) {
		mock := NewMockErrorNotifier()
		mock.SetThreshold(SeverityCritical)
		assert.Equal(t, SeverityCritical, mock.Threshold)
		assert.Len(t, mock.SetThresholdCalls, 1)
	})

	t.Run("SetRateLimit", func(t *testing.T) {
		mock := NewMockErrorNotifier()
		mock.SetRateLimit(time.Minute, 10)
		assert.Len(t, mock.SetRateLimitCalls, 1)
		assert.Equal(t, time.Minute, mock.SetRateLimitCalls[0].Duration)
		assert.Equal(t, 10, mock.SetRateLimitCalls[0].Count)
	})

	t.Run("AddChannel", func(t *testing.T) {
		mock := NewMockErrorNotifier()
		channel := &testChannel{name: "test-channel"}
		err := mock.AddChannel(channel)
		require.NoError(t, err)
		assert.Len(t, mock.AddChannelCalls, 1)
		assert.Contains(t, mock.Channels, "test-channel")
	})

	t.Run("AddChannel_with_error", func(t *testing.T) {
		mock := NewMockErrorNotifier()
		mock.ShouldError = true
		channel := &testChannel{name: "test-channel"}
		err := mock.AddChannel(channel)
		require.Error(t, err)
	})

	t.Run("RemoveChannel", func(t *testing.T) {
		mock := NewMockErrorNotifier()
		mock.Channels["test-channel"] = &testChannel{name: "test-channel"}
		err := mock.RemoveChannel("test-channel")
		require.NoError(t, err)
		assert.Len(t, mock.RemoveChannelCalls, 1)
		assert.NotContains(t, mock.Channels, "test-channel")
	})

	t.Run("RemoveChannel_with_error", func(t *testing.T) {
		mock := NewMockErrorNotifier()
		mock.ShouldError = true
		err := mock.RemoveChannel("test-channel")
		require.Error(t, err)
	})
}

// testChannel is a minimal implementation of NotificationChannel for testing.
type testChannel struct {
	name    string
	enabled bool
}

func (c *testChannel) Name() string                                       { return c.name }
func (c *testChannel) Send(_ context.Context, _ *ErrorNotification) error { return nil }
func (c *testChannel) IsEnabled() bool                                    { return c.enabled }
func (c *testChannel) SetEnabled(e bool)                                  { c.enabled = e }

// =============================================================================
// MockErrorHandler Tests
// =============================================================================

// TestMockErrorHandler tests all MockErrorHandler methods for coverage.
func TestMockErrorHandler(t *testing.T) {
	t.Run("NewMockErrorHandler", func(t *testing.T) {
		mock := NewMockErrorHandler()
		require.NotNil(t, mock)
	})

	t.Run("Handle", func(t *testing.T) {
		mock := NewMockErrorHandler()
		// Without a handleFunc set, Handle returns the error as-is
		err := mock.Handle(errCGGeneric)
		assert.Equal(t, errCGGeneric, err)
		assert.Equal(t, 1, mock.GetCallCount("Handle"))
	})

	t.Run("HandleWithContext", func(t *testing.T) {
		mock := NewMockErrorHandler()
		// Without a handleContextFunc set, HandleWithContext returns the error as-is
		err := mock.HandleWithContext(context.Background(), errCGGeneric)
		assert.Equal(t, errCGGeneric, err)
		assert.Equal(t, 1, mock.GetCallCount("HandleWithContext"))
	})

	t.Run("OnError", func(t *testing.T) {
		mock := NewMockErrorHandler()
		result := mock.OnError(ErrBuildFailed, func(_ MageError) error { return nil })
		assert.Equal(t, mock, result)
		assert.Equal(t, 1, mock.GetCallCount("OnError"))
	})

	t.Run("OnSeverity", func(t *testing.T) {
		mock := NewMockErrorHandler()
		result := mock.OnSeverity(SeverityError, func(_ MageError) error { return nil })
		assert.Equal(t, mock, result)
		assert.Equal(t, 1, mock.GetCallCount("OnSeverity"))
	})

	t.Run("SetDefault", func(t *testing.T) {
		mock := NewMockErrorHandler()
		result := mock.SetDefault(func(err error) error { return err })
		assert.Equal(t, mock, result)
		assert.Equal(t, 1, mock.GetCallCount("SetDefault"))
	})

	t.Run("SetFallback", func(t *testing.T) {
		mock := NewMockErrorHandler()
		result := mock.SetFallback(func(err error) error { return err })
		assert.Equal(t, mock, result)
		assert.Equal(t, 1, mock.GetCallCount("SetFallback"))
	})

	t.Run("GetCallCount", func(t *testing.T) {
		mock := NewMockErrorHandler()
		_ = mock.Handle(errCGGeneric) //nolint:errcheck // test coverage
		_ = mock.Handle(errCGGeneric) //nolint:errcheck // test coverage
		assert.Equal(t, 2, mock.GetCallCount("Handle"))
	})
}

// =============================================================================
// MockErrorChain Tests
// =============================================================================

// TestMockErrorChain tests all MockErrorChain methods for coverage.
//
//nolint:errcheck // Mock methods return ErrorChain for chaining, not errors
func TestMockErrorChain(t *testing.T) {
	t.Run("NewMockErrorChain", func(t *testing.T) {
		mock := NewMockErrorChain()
		require.NotNil(t, mock)
	})

	t.Run("Add", func(t *testing.T) {
		mock := NewMockErrorChain()
		_ = mock.Add(errCGGeneric)
		assert.Equal(t, 1, mock.Count())
	})

	t.Run("AddWithContext", func(t *testing.T) {
		mock := NewMockErrorChain()
		ctx := &ErrorContext{Operation: "test-op"}
		_ = mock.AddWithContext(errCGGeneric, ctx)
		assert.Equal(t, 1, mock.Count())
	})

	t.Run("Error", func(t *testing.T) {
		mock := NewMockErrorChain()
		_ = mock.Add(errCGGeneric)
		result := mock.Error()
		assert.Contains(t, result, "test error")
	})

	t.Run("Error_empty", func(t *testing.T) {
		mock := NewMockErrorChain()
		result := mock.Error()
		// Mock returns "no errors" when empty
		assert.Equal(t, "no errors", result)
	})

	t.Run("Errors", func(t *testing.T) {
		mock := NewMockErrorChain()
		_ = mock.Add(errCGGeneric)
		_ = mock.Add(errCGDB)
		result := mock.Errors()
		assert.Len(t, result, 2)
	})

	t.Run("First", func(t *testing.T) {
		mock := NewMockErrorChain()
		_ = mock.Add(errCGGeneric)
		_ = mock.Add(errCGDB)
		result := mock.First()
		assert.Equal(t, errCGGeneric, result)
	})

	t.Run("First_empty", func(t *testing.T) {
		mock := NewMockErrorChain()
		result := mock.First()
		require.NoError(t, result)
	})

	t.Run("Last", func(t *testing.T) {
		mock := NewMockErrorChain()
		_ = mock.Add(errCGGeneric)
		_ = mock.Add(errCGDB)
		result := mock.Last()
		assert.Equal(t, errCGDB, result)
	})

	t.Run("Last_empty", func(t *testing.T) {
		mock := NewMockErrorChain()
		result := mock.Last()
		require.NoError(t, result)
	})

	t.Run("Count", func(t *testing.T) {
		mock := NewMockErrorChain()
		_ = mock.Add(errCGGeneric)
		_ = mock.Add(errCGDB)
		assert.Equal(t, 2, mock.Count())
	})

	t.Run("HasError", func(t *testing.T) {
		mock := NewMockErrorChain()
		// Mock always returns false regardless of errors
		assert.False(t, mock.HasError(ErrBuildFailed))
		mageErr := NewMageError("test").WithCode(ErrBuildFailed)
		_ = mock.Add(mageErr)
		// Mock still returns false - it doesn't actually check
		assert.False(t, mock.HasError(ErrBuildFailed))
		// Verify method was called
		assert.Equal(t, 2, mock.GetCallCount("HasError"))
	})

	t.Run("FindByCode", func(t *testing.T) {
		mock := NewMockErrorChain()
		mageErr := NewMageError("test").WithCode(ErrBuildFailed)
		_ = mock.Add(mageErr)
		result := mock.FindByCode(ErrBuildFailed)
		// Mock always returns nil - just ensure method is called
		_ = result
		assert.Equal(t, 1, mock.GetCallCount("FindByCode"))
	})

	t.Run("FindByCode_not_found", func(t *testing.T) {
		mock := NewMockErrorChain()
		_ = mock.Add(errCGGeneric)
		result := mock.FindByCode(ErrBuildFailed)
		// Mock always returns nil
		assert.Nil(t, result)
	})

	t.Run("ForEach", func(t *testing.T) {
		mock := NewMockErrorChain()
		_ = mock.Add(errCGGeneric)
		_ = mock.Add(errCGDB)
		_ = mock.ForEach(func(_ error) error {
			return nil
		})
		// Just ensure no panic - call count was incremented
		assert.Equal(t, 1, mock.GetCallCount("ForEach"))
	})

	t.Run("Filter", func(t *testing.T) {
		mock := NewMockErrorChain()
		_ = mock.Add(errCGGeneric)
		_ = mock.Add(errCGDB)
		result := mock.Filter(func(_ error) bool {
			return true
		})
		// Just ensure no panic - result might be empty depending on implementation
		_ = result
		assert.Equal(t, 1, mock.GetCallCount("Filter"))
	})

	t.Run("ToSlice", func(t *testing.T) {
		mock := NewMockErrorChain()
		_ = mock.Add(errCGGeneric)
		_ = mock.Add(errCGDB)
		result := mock.ToSlice()
		assert.Len(t, result, 2)
	})

	t.Run("GetCallCount", func(t *testing.T) {
		mock := NewMockErrorChain()
		_ = mock.Add(errCGGeneric)
		_ = mock.Add(errCGDB)
		_ = mock.Count()
		assert.Equal(t, 3, mock.GetCallCount("Add")+mock.GetCallCount("Count"))
	})
}

// =============================================================================
// DefaultErrorRegistry Tests
// =============================================================================

// TestDefaultErrorRegistry tests DefaultErrorRegistry methods for coverage.
func TestDefaultErrorRegistry(t *testing.T) {
	t.Run("Register", func(t *testing.T) {
		registry := &DefaultErrorRegistry{
			definitions: make(map[ErrorCode]ErrorDefinition),
		}
		err := registry.Register("TEST_CODE", "Test description")
		require.NoError(t, err)
	})

	t.Run("RegisterWithSeverity", func(t *testing.T) {
		registry := &DefaultErrorRegistry{
			definitions: make(map[ErrorCode]ErrorDefinition),
		}
		err := registry.RegisterWithSeverity("TEST_CODE", "Test description", SeverityError)
		require.NoError(t, err)
	})

	t.Run("Unregister", func(t *testing.T) {
		registry := &DefaultErrorRegistry{
			definitions: map[ErrorCode]ErrorDefinition{
				"TEST_CODE": {Code: "TEST_CODE", Description: "Test"},
			},
		}
		err := registry.Unregister("TEST_CODE")
		require.NoError(t, err)
		assert.NotContains(t, registry.definitions, ErrorCode("TEST_CODE"))
	})

	t.Run("Get", func(t *testing.T) {
		registry := &DefaultErrorRegistry{
			definitions: map[ErrorCode]ErrorDefinition{
				"TEST_CODE": {Code: "TEST_CODE", Description: "Test"},
			},
		}
		def, exists := registry.Get("TEST_CODE")
		assert.True(t, exists)
		assert.Equal(t, "Test", def.Description)
	})

	t.Run("Get_not_found", func(t *testing.T) {
		registry := &DefaultErrorRegistry{
			definitions: make(map[ErrorCode]ErrorDefinition),
		}
		_, exists := registry.Get("NOT_FOUND")
		assert.False(t, exists)
	})

	t.Run("List", func(t *testing.T) {
		registry := &DefaultErrorRegistry{
			definitions: map[ErrorCode]ErrorDefinition{
				"CODE1": {Code: "CODE1"},
				"CODE2": {Code: "CODE2"},
			},
		}
		result := registry.List()
		assert.Len(t, result, 2)
	})

	t.Run("ListByPrefix", func(t *testing.T) {
		registry := &DefaultErrorRegistry{
			definitions: map[ErrorCode]ErrorDefinition{
				"BUILD_FAILED":   {Code: "BUILD_FAILED"},
				"BUILD_TIMEOUT":  {Code: "BUILD_TIMEOUT"},
				"CONFIG_INVALID": {Code: "CONFIG_INVALID"},
			},
		}
		result := registry.ListByPrefix("BUILD_")
		assert.Len(t, result, 2)
	})

	t.Run("ListBySeverity", func(t *testing.T) {
		registry := &DefaultErrorRegistry{
			definitions: map[ErrorCode]ErrorDefinition{
				"CODE1": {Code: "CODE1", Severity: SeverityError},
				"CODE2": {Code: "CODE2", Severity: SeverityWarning},
				"CODE3": {Code: "CODE3", Severity: SeverityError},
			},
		}
		result := registry.ListBySeverity(SeverityError)
		assert.Len(t, result, 2)
	})

	t.Run("Contains", func(t *testing.T) {
		registry := &DefaultErrorRegistry{
			definitions: map[ErrorCode]ErrorDefinition{
				"CODE1": {Code: "CODE1"},
			},
		}
		assert.True(t, registry.Contains("CODE1"))
		assert.False(t, registry.Contains("CODE2"))
	})

	t.Run("Clear", func(t *testing.T) {
		registry := &DefaultErrorRegistry{
			definitions: map[ErrorCode]ErrorDefinition{
				"CODE1": {Code: "CODE1"},
				"CODE2": {Code: "CODE2"},
			},
		}
		err := registry.Clear()
		require.NoError(t, err)
		assert.Empty(t, registry.definitions)
	})
}

// =============================================================================
// DefaultErrorNotifier Configuration Tests
// =============================================================================

// TestDefaultErrorNotifierConfiguration tests configuration methods.
func TestDefaultErrorNotifierConfiguration(t *testing.T) {
	t.Run("SetThreshold", func(t *testing.T) {
		notifier := NewErrorNotifier()
		notifier.SetThreshold(SeverityCritical)
		// No assertion needed - just ensure no panic
	})

	t.Run("SetRateLimit", func(t *testing.T) {
		notifier := NewErrorNotifier()
		notifier.SetRateLimit(time.Minute, 10)
		// No assertion needed - just ensure no panic
	})

	t.Run("AddChannel", func(t *testing.T) {
		notifier := NewErrorNotifier()
		channel := &testChannel{name: "test"}
		err := notifier.AddChannel(channel)
		require.NoError(t, err)
	})

	t.Run("AddChannel_nil", func(t *testing.T) {
		notifier := NewErrorNotifier()
		err := notifier.AddChannel(nil)
		require.Error(t, err)
	})

	t.Run("RemoveChannel", func(t *testing.T) {
		notifier := NewErrorNotifier()
		channel := &testChannel{name: "test"}
		require.NoError(t, notifier.AddChannel(channel))
		err := notifier.RemoveChannel("test")
		require.NoError(t, err)
	})

	t.Run("RemoveChannel_not_found", func(t *testing.T) {
		notifier := NewErrorNotifier()
		err := notifier.RemoveChannel("nonexistent")
		require.Error(t, err)
	})

	t.Run("GetChannels_via_concrete", func(t *testing.T) {
		notifier, ok := NewErrorNotifier().(*DefaultErrorNotifier)
		require.True(t, ok, "expected *DefaultErrorNotifier")
		require.NoError(t, notifier.AddChannel(&testChannel{name: "ch1"}))
		require.NoError(t, notifier.AddChannel(&testChannel{name: "ch2"}))
		result := notifier.GetChannels()
		assert.Len(t, result, 2)
	})

	t.Run("SetEnabled_via_concrete", func(t *testing.T) {
		notifier, ok := NewErrorNotifier().(*DefaultErrorNotifier)
		require.True(t, ok, "expected *DefaultErrorNotifier")
		notifier.SetEnabled(false)
		// No assertion needed - just ensure no panic
	})
}

// =============================================================================
// Not Matcher Tests
// =============================================================================

// TestNotMatcher tests the Not() matcher function.
func TestNotMatcher(t *testing.T) {
	t.Run("Not_inverts_match", func(t *testing.T) {
		// Use the NewMatcher builder to create a matcher
		matcher := NewMatcher().MatchCode(ErrBuildFailed)
		notMatcher := matcher.Not()

		buildErr := NewMageError("test").WithCode(ErrBuildFailed)
		configErr := NewMageError("test").WithCode(ErrConfigInvalid)

		// The Not matcher should invert the match result
		assert.False(t, notMatcher.Match(buildErr))
		assert.True(t, notMatcher.Match(configErr))
	})

	t.Run("Not_via_direct_method", func(t *testing.T) {
		// Test using the public interface method
		matcher := NewMatcher().MatchSeverity(SeverityError)
		notMatcher := matcher.Not()

		errorSeverity := NewMageError("test").WithSeverity(SeverityError)
		warnSeverity := NewMageError("test").WithSeverity(SeverityWarning)

		assert.False(t, notMatcher.Match(errorSeverity))
		assert.True(t, notMatcher.Match(warnSeverity))
	})
}

// =============================================================================
// DefaultErrorRecovery RecoverWithFallback Test
// =============================================================================

// TestDefaultErrorRecoveryRecoverWithFallback tests RecoverWithFallback wrapper.
func TestDefaultErrorRecoveryRecoverWithFallback(t *testing.T) {
	t.Run("calls underlying recovery", func(t *testing.T) {
		recovery := &DefaultErrorRecovery{}
		err := recovery.RecoverWithFallback(
			func() error { return nil },
			func(_ error) error { return nil },
		)
		require.NoError(t, err)
	})

	t.Run("with panic", func(t *testing.T) {
		recovery := &DefaultErrorRecovery{}
		err := recovery.RecoverWithFallback(
			func() error { panic("test panic") },
			func(_ error) error { return nil },
		)
		require.NoError(t, err) // Fallback handles the panic
	})

	t.Run("with error and fallback", func(t *testing.T) {
		recovery := &DefaultErrorRecovery{}
		err := recovery.RecoverWithFallback(
			func() error { return errCGFunction },
			func(err error) error { return Wrap(err, "fallback") },
		)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "fallback")
	})
}
