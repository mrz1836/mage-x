package errors

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMageError_Basic(t *testing.T) {
	// Test basic error creation
	err := New("test error")
	assert.Equal(t, "test error", err.Error(), "Error() should return correct message")

	// Test formatted error
	err = Newf("test error %d", 42)
	assert.Equal(t, "test error 42", err.Error(), "Newf() should format correctly")

	// Test with code
	err = WithCode(ErrBuildFailed, "build failed")
	assert.Equal(t, ErrBuildFailed, err.Code(), "WithCode() should set correct code")

	// Test with code and format
	err = WithCodef(ErrTestFailed, "test %s failed", "unit")
	assert.Equal(t, ErrTestFailed, err.Code(), "WithCodef() should set correct code")
	assert.Equal(t, "test unit failed", err.Error(), "WithCodef() should format message correctly")
}

func TestMageError_Wrapping(t *testing.T) {
	// Test wrapping nil error
	err := Wrap(nil, "wrapper")
	assert.Nil(t, err, "Wrap(nil) should return nil")

	// Test basic wrapping
	baseErr := New("base error")
	wrapped := Wrap(baseErr, "wrapped")
	assert.Contains(t, wrapped.Error(), "wrapped", "Wrapped error should contain wrapper message")
	require.ErrorIs(t, wrapped.Cause(), baseErr, "Cause() should return original error")

	// Test formatted wrapping
	wrapped = Wrapf(baseErr, "wrapped with %s", "context")
	assert.Contains(t, wrapped.Error(), "wrapped with context", "Wrapped error should contain formatted message")
}

func TestMageError_Context(t *testing.T) {
	err := NewBuilder().
		WithMessage("test error").
		WithCode(ErrBuildFailed).
		WithSeverity(SeverityCritical).
		WithOperation("build").
		WithResource("main.go").
		WithField("exitCode", 1).
		WithFields(map[string]interface{}{
			"command":  "go build",
			"duration": "5s",
		}).
		Build()

	// Check all fields
	assert.Equal(t, ErrBuildFailed, err.Code(), "Code() should be ErrBuildFailed")
	assert.Equal(t, SeverityCritical, err.Severity(), "Severity() should be SeverityCritical")

	ctx := err.Context()
	assert.Equal(t, "build", ctx.Operation, "Operation should be 'build'")
	assert.Equal(t, "main.go", ctx.Resource, "Resource should be 'main.go'")
	assert.Equal(t, 1, ctx.Fields["exitCode"], "Field exitCode should be 1")
	assert.Equal(t, "go build", ctx.Fields["command"], "Field command should be 'go build'")
}

func TestMageError_Is(t *testing.T) {
	// Test same error
	err1 := WithCode(ErrBuildFailed, "build failed")
	assert.True(t, err1.Is(err1), "Error should match itself")

	// Test same code
	err2 := WithCode(ErrBuildFailed, "different message")
	assert.True(t, err1.Is(err2), "Errors with same code should match")

	// Test different code
	err3 := WithCode(ErrTestFailed, "test failed")
	assert.False(t, err1.Is(err3), "Errors with different codes should not match")

	// Test wrapped error
	wrapped := Wrap(err1, "wrapped")
	target := WithCode(ErrBuildFailed, "target")
	assert.True(t, wrapped.Is(target), "Wrapped error should match by code")
}

func TestMageError_As(t *testing.T) {
	err := WithCode(ErrBuildFailed, "build failed")

	// Test As with DefaultMageError
	var mageErr *DefaultMageError
	assert.True(t, err.As(&mageErr), "As should work with *DefaultMageError")
	assert.NotNil(t, mageErr, "As should properly assign the error")
	assert.Equal(t, ErrBuildFailed, mageErr.Code(), "As should properly assign the error")

	// Test As with MageError interface
	var iface MageError
	assert.True(t, err.As(&iface), "As should work with MageError interface")
	assert.Equal(t, ErrBuildFailed, iface.Code(), "As should properly assign the interface")
}

func TestMageError_StackTrace(t *testing.T) {
	err := NewBuilder().
		WithMessage("test error").
		WithStackTrace().
		Build()

	formatted := err.Format(true)
	assert.Contains(t, formatted, "Stack trace:", "Formatted error should include stack trace when requested")

	// Test without stack trace
	formattedNoStack := err.Format(false)
	assert.NotContains(t, formattedNoStack, "Stack trace:", "Formatted error should not include stack trace when not requested")
}

func TestErrorChain(t *testing.T) {
	chain := NewChain()

	// Test empty chain
	assert.Equal(t, 0, chain.Count(), "Empty chain should have count 0")
	assert.NoError(t, chain.First(), "Empty chain First() should return nil")
	assert.NoError(t, chain.Last(), "Empty chain Last() should return nil")

	// Add errors
	err1 := WithCode(ErrBuildFailed, "build failed")
	err2 := WithCode(ErrTestFailed, "test failed")
	err3 := New("generic error")

	chain = chain.Add(err1).Add(err2).Add(err3)

	// Test count
	assert.Equal(t, 3, chain.Count(), "Chain should have count 3")

	// Test first/last
	require.ErrorIs(t, chain.First(), err1, "First() should return first error")
	require.ErrorIs(t, chain.Last(), err3, "Last() should return last error")

	// Test HasError
	assert.True(t, chain.HasError(ErrBuildFailed), "HasError should find ErrBuildFailed")
	assert.False(t, chain.HasError(ErrConfigInvalid), "HasError should not find ErrConfigInvalid")

	// Test FindByCode
	found := chain.FindByCode(ErrTestFailed)
	assert.NotNil(t, found, "FindByCode should find ErrTestFailed")
	assert.Equal(t, ErrTestFailed, found.Code(), "FindByCode should return correct error")

	// Test ForEach
	count := 0
	forEachErr := chain.ForEach(func(_ error) error {
		count++
		return nil
	})
	require.NoError(t, forEachErr, "ForEach should not fail")
	assert.Equal(t, 3, count, "ForEach should iterate over all errors")

	// Test Filter
	filtered := chain.Filter(func(err error) bool {
		var mageErr MageError
		if errors.As(err, &mageErr) {
			return mageErr.Code() == ErrBuildFailed || mageErr.Code() == ErrTestFailed
		}
		return false
	})
	assert.Len(t, filtered, 2, "Filter should return 2 errors")
}

func TestErrorHandler(t *testing.T) {
	handler := NewHandler()

	// Test code handler
	handled := false
	handler.OnError(ErrBuildFailed, func(_ MageError) error {
		handled = true
		return nil
	})

	err := WithCode(ErrBuildFailed, "build failed")
	result := handler.Handle(err)
	require.NoError(t, result, "Handle should not return an error")

	assert.True(t, handled, "Code handler should have been called")

	// Test severity handler
	severityHandled := false
	handler.OnSeverity(SeverityCritical, func(_ MageError) error {
		severityHandled = true
		return nil
	})

	criticalErr := NewBuilder().
		WithMessage("critical error").
		WithSeverity(SeverityCritical).
		Build()
	result = handler.Handle(criticalErr)
	require.NoError(t, result, "Handle should not return an error")

	assert.True(t, severityHandled, "Severity handler should have been called")

	// Test default handler
	defaultHandled := false
	handler.SetDefault(func(_ error) error {
		defaultHandled = true
		return nil
	})

	genericErr := fmt.Errorf("generic error")
	result = handler.Handle(genericErr)
	require.NoError(t, result, "Handle should not return an error")

	assert.True(t, defaultHandled, "Default handler should have been called")

	// Test context handler
	type contextKey string
	const requestIDKey contextKey = "requestID"
	ctx := context.WithValue(context.Background(), requestIDKey, "12345")
	contextErr := New("context error")

	result = handler.HandleWithContext(ctx, contextErr)
	assert.NoError(t, result, "HandleWithContext should not return an error when handled by default handler")
}

func TestErrorRegistry(t *testing.T) {
	registry := NewRegistry()

	// Test registration
	err := registry.Register("CUSTOM_ERROR", "Custom error description")
	require.NoError(t, err, "Register should not fail")

	// Test duplicate registration
	err = registry.Register("CUSTOM_ERROR", "Duplicate")
	require.Error(t, err, "Duplicate registration should fail")

	// Test Get
	def, exists := registry.Get("CUSTOM_ERROR")
	assert.True(t, exists, "Get should find registered error")
	assert.Equal(t, "Custom error description", def.Description, "Description should match")

	// Test Contains
	assert.True(t, registry.Contains("CUSTOM_ERROR"), "Contains should return true for registered error")

	// Test List
	list := registry.List()
	assert.NotEmpty(t, list, "List should return registered errors")

	// Test ListByPrefix
	err = registry.Register("PREFIX_TEST_1", "Test error 1")
	require.NoError(t, err, "Register should not fail")
	err = registry.Register("PREFIX_TEST_2", "Test error 2")
	require.NoError(t, err, "Register should not fail")

	testErrors := registry.ListByPrefix("PREFIX_TEST_")
	assert.Len(t, testErrors, 2, "ListByPrefix should return 2 errors")

	// Test Unregister
	err = registry.Unregister("CUSTOM_ERROR")
	require.NoError(t, err, "Unregister should not fail")

	assert.False(t, registry.Contains("CUSTOM_ERROR"), "Unregistered error should not exist")
}

func TestErrorRecovery(t *testing.T) {
	recovery := NewRecovery()

	// Test successful execution
	err := recovery.Recover(func() error {
		return nil
	})
	require.NoError(t, err, "Recover should return nil for successful execution")

	// Test panic recovery
	err = recovery.Recover(func() error {
		panic("test panic")
	})
	require.Error(t, err, "Recover should return error for panic")
	assert.Contains(t, err.Error(), "panic recovered", "Recovered error should mention panic")

	// Test retry
	attempts := 0
	err = recovery.RecoverWithRetry(func() error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("attempt %d failed", attempts)
		}
		return nil
	}, 5, 10*time.Millisecond)
	require.NoError(t, err, "RecoverWithRetry should succeed after retries")
	assert.Equal(t, 3, attempts, "Should make exactly 3 attempts")

	// Test backoff
	backoffAttempts := 0
	err = recovery.RecoverWithBackoff(func() error {
		backoffAttempts++
		if backoffAttempts < 2 {
			return fmt.Errorf("backoff attempt %d", backoffAttempts)
		}
		return nil
	}, BackoffConfig{
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		MaxRetries:   3,
	})
	require.NoError(t, err, "RecoverWithBackoff should succeed")

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err = recovery.RecoverWithContext(ctx, func() error {
		time.Sleep(100 * time.Millisecond)
		return nil
	})

	require.Error(t, err, "RecoverWithContext should return cancellation error")
	assert.Contains(t, err.Error(), "canceled", "Error should mention cancellation")
}

func TestErrorMetrics(t *testing.T) {
	metrics := NewMetrics()

	// Record some errors
	err1 := WithCode(ErrBuildFailed, "build failed")
	err2 := WithCode(ErrTestFailed, "test failed")
	err3 := fmt.Errorf("generic error")

	metrics.RecordError(err1)
	metrics.RecordError(err1)
	metrics.RecordError(err2)
	metrics.RecordError(err3)

	// Test counts
	assert.Equal(t, int64(2), metrics.GetCount(ErrBuildFailed), "GetCount(ErrBuildFailed) should be 2")
	assert.Equal(t, int64(1), metrics.GetCount(ErrTestFailed), "GetCount(ErrTestFailed) should be 1")

	// Test top errors
	top := metrics.GetTopErrors(2)
	assert.NotEmpty(t, top, "GetTopErrors should return results")

	// The top error should be ErrBuildFailed with count 2
	if len(top) > 0 {
		assert.Equal(t, ErrBuildFailed, top[0].Code, "Top error should be ErrBuildFailed")
	}

	// Test reset
	err := metrics.Reset()
	require.NoError(t, err, "Reset should not fail")

	assert.Equal(t, int64(0), metrics.GetCount(ErrBuildFailed), "Count should be 0 after reset")
}

func TestErrorFormatter(t *testing.T) {
	formatter := NewFormatter()

	// Test basic formatting
	err := WithCode(ErrBuildFailed, "build failed")
	formatted := formatter.Format(err)

	assert.Contains(t, formatted, "BUILD_FAILED", "Formatted error should contain error code")
	assert.Contains(t, formatted, "build failed", "Formatted error should contain message")

	// Test chain formatting
	chain := NewChain()
	chain = chain.Add(WithCode(ErrBuildFailed, "build failed"))
	chain = chain.Add(WithCode(ErrTestFailed, "test failed"))

	chainFormatted := formatter.FormatChain(chain)
	assert.Contains(t, chainFormatted, "2 errors", "Chain format should show error count")

	// Test compact mode
	opts := FormatOptions{
		CompactMode:   true,
		IncludeFields: true,
	}

	complexErr := NewBuilder().
		WithMessage("complex error").
		WithCode(ErrBuildFailed).
		WithField("key", "value").
		Build()

	compact := formatter.FormatWithOptions(complexErr, opts)
	assert.NotContains(t, compact, "\n", "Compact format should be single line")
	assert.Contains(t, compact, "key=value", "Compact format should include fields")
}

func TestHelperFunctions(t *testing.T) {
	// Test IsNotFound
	err := WithCode(ErrNotFound, "not found")
	assert.True(t, IsNotFound(err), "IsNotFound should return true for ErrNotFound")

	err = WithCode(ErrFileNotFound, "file not found")
	assert.True(t, IsNotFound(err), "IsNotFound should return true for ErrFileNotFound")

	// Test IsTimeout
	err = WithCode(ErrTimeout, "timeout")
	assert.True(t, IsTimeout(err), "IsTimeout should return true for ErrTimeout")

	// Test IsBuildError
	err = WithCode(ErrBuildFailed, "build failed")
	assert.True(t, IsBuildError(err), "IsBuildError should return true for ErrBuildFailed")

	// Test Combine
	err1 := New("error 1")
	err2 := New("error 2")
	combined := Combine(err1, nil, err2)

	var chain ErrorChain
	if assert.ErrorAs(t, combined, &chain, "Combine should return ErrorChain for multiple errors") {
		assert.Equal(t, 2, chain.Count(), "Combined errors should have 2 errors")
	}

	// Test FirstError
	first := FirstError(nil, nil, err1, err2)
	require.ErrorIs(t, first, err1, "FirstError should return first non-nil error")

	// Test SafeExecute
	executed := false
	safeErr := SafeExecute(func() error {
		executed = true
		return nil
	})
	require.NoError(t, safeErr, "SafeExecute should execute successfully")
	assert.True(t, executed, "SafeExecute should execute function")

	// Test SafeExecute with panic
	safeErr = SafeExecute(func() error {
		panic("test panic")
	})
	assert.Error(t, safeErr, "SafeExecute should recover from panic")
}

func TestSeverity(t *testing.T) {
	// Test String representation
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityDebug, "DEBUG"},
		{SeverityInfo, "INFO"},
		{SeverityWarning, "WARNING"},
		{SeverityError, "ERROR"},
		{SeverityCritical, "CRITICAL"},
		{SeverityFatal, "FATAL"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.severity.String(), "Severity.String() should return correct value")
	}

	// Test marshaling
	data, err := SeverityError.MarshalText()
	require.NoError(t, err, "MarshalText should not fail")
	assert.Equal(t, "ERROR", string(data), "MarshalText should return ERROR")

	// Test unmarshaling
	var sev Severity
	err = sev.UnmarshalText([]byte("CRITICAL"))
	require.NoError(t, err, "UnmarshalText should not fail")
	assert.Equal(t, SeverityCritical, sev, "UnmarshalText should set correct value")
}

// Benchmark tests

func BenchmarkErrorCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := New("test error")
		if err == nil {
			b.Fatal("expected error, got nil")
		}
	}
}

func BenchmarkErrorBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := NewBuilder().
			WithMessage("test error").
			WithCode(ErrBuildFailed).
			WithSeverity(SeverityError).
			WithField("key", "value").
			Build()
		if err == nil {
			b.Fatal("expected error, got nil")
		}
	}
}

func BenchmarkErrorFormatting(b *testing.B) {
	err := NewBuilder().
		WithMessage("test error").
		WithCode(ErrBuildFailed).
		WithStackTrace().
		Build()

	formatter := NewFormatter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = formatter.Format(err)
	}
}

func BenchmarkErrorChain(b *testing.B) {
	errs := make([]error, 10)
	for i := range errs {
		errs[i] = Newf("error %d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chain := NewChain()
		for _, err := range errs {
			chain = chain.Add(err)
		}
		_ = chain.Count()
	}
}
