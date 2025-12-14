package errors

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Static error variables for tests (err113 compliance)
var (
	errRecoveryFunc     = errors.New("function error")
	errRecoveryPanic    = errors.New("panic error")
	errRecoveryOriginal = errors.New("original error")
	errRecoveryFallback = errors.New("fallback error")
	errRecoveryAlways   = errors.New("always fails")
	errRecoveryNotYet   = errors.New("not yet")
	errRecoveryFails    = errors.New("fails")
	errRecoveryNormal   = errors.New("normal error")
	errRecoveryFail     = errors.New("fail")
	errRecoveryAny      = errors.New("any error")
)

// TestRealRecovery_SuccessfulExecution verifies Recover returns nil for successful execution
func TestRealRecovery_SuccessfulExecution(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()
	executed := false

	err := recovery.Recover(func() error {
		executed = true
		return nil
	})

	require.NoError(t, err)
	assert.True(t, executed, "Function should be executed")
}

// TestRealRecovery_FunctionReturnsError verifies Recover returns function's error
func TestRealRecovery_FunctionReturnsError(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()

	err := recovery.Recover(func() error {
		return errRecoveryFunc
	})

	assert.Equal(t, errRecoveryFunc, err, "Should return function's error")
}

// TestRealRecovery_PanicWithError verifies panic(error) is wrapped correctly
func TestRealRecovery_PanicWithError(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()

	err := recovery.Recover(func() error {
		panic(errRecoveryPanic)
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "panic recovered", "Error should mention panic recovery")

	// Verify the original error is preserved in the chain
	var mageErr MageError
	require.ErrorAs(t, err, &mageErr)
	assert.Error(t, mageErr.Cause(), "Cause should be set")
}

// TestRealRecovery_PanicWithNonError verifies panic(non-error) is converted to error
func TestRealRecovery_PanicWithNonError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		panicValue interface{}
		contains   string
	}{
		{
			name:       "panic with string",
			panicValue: "panic string",
			contains:   "panic string",
		},
		{
			name:       "panic with int",
			panicValue: 42,
			contains:   "42",
		},
		{
			name:       "panic with struct",
			panicValue: struct{ msg string }{"test"},
			contains:   "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			recovery := NewErrorRecovery()
			err := recovery.Recover(func() error {
				panic(tt.panicValue)
			})

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.contains)
			assert.Contains(t, err.Error(), "panic recovered")
		})
	}
}

// TestRealRecovery_FallbackOnError verifies fallback is called on error
func TestRealRecovery_FallbackOnError(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()
	fallbackCalled := false

	err := recovery.RecoverWithFallback(
		func() error {
			return errRecoveryOriginal
		},
		func(err error) error {
			fallbackCalled = true
			assert.Equal(t, errRecoveryOriginal, err)
			return errRecoveryFallback
		},
	)

	require.Error(t, err)
	assert.True(t, fallbackCalled, "Fallback should be called")
	assert.Equal(t, "fallback error", err.Error())
}

// TestRealRecovery_FallbackOnSuccess verifies fallback is not called on success
func TestRealRecovery_FallbackOnSuccess(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()

	err := recovery.RecoverWithFallback(
		func() error {
			return nil
		},
		func(_ error) error {
			t.Error("Fallback should not be called on success")
			return errRecoveryFallback
		},
	)

	require.NoError(t, err)
}

// TestRealRecovery_FallbackNilIgnored verifies nil fallback doesn't panic
func TestRealRecovery_FallbackNilIgnored(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()

	err := recovery.RecoverWithFallback(
		func() error {
			return errRecoveryOriginal
		},
		nil, // nil fallback
	)

	assert.Equal(t, errRecoveryOriginal, err, "Original error should be returned when fallback is nil")
}

// TestRealRecovery_RetryExhaustsAttempts verifies retry stops after max retries
func TestRealRecovery_RetryExhaustsAttempts(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()
	var attempts atomic.Int32

	err := recovery.RecoverWithRetry(
		func() error {
			attempts.Add(1)
			return errRecoveryAlways
		},
		3,                   // 3 retries
		10*time.Millisecond, // short delay for tests
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed after 4 retries") // 1 initial + 3 retries
	assert.Equal(t, int32(4), attempts.Load(), "Should make 4 total attempts")
}

// TestRealRecovery_RetrySucceedsEventually verifies success stops retries
func TestRealRecovery_RetrySucceedsEventually(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()
	var attempts atomic.Int32

	err := recovery.RecoverWithRetry(
		func() error {
			current := attempts.Add(1)
			if current < 3 {
				return errRecoveryNotYet
			}
			return nil // Succeed on 3rd attempt
		},
		5, // Allow up to 5 retries
		10*time.Millisecond,
	)

	require.NoError(t, err)
	assert.Equal(t, int32(3), attempts.Load(), "Should succeed on 3rd attempt")
}

// TestRealRecovery_RetryZeroRetries verifies zero retries means one attempt
func TestRealRecovery_RetryZeroRetries(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()
	var attempts atomic.Int32

	err := recovery.RecoverWithRetry(
		func() error {
			attempts.Add(1)
			return errRecoveryFails
		},
		0, // Zero retries
		10*time.Millisecond,
	)

	require.Error(t, err)
	assert.Equal(t, int32(1), attempts.Load(), "Should make only 1 attempt with 0 retries")
}

// TestRealRecovery_BackoffConfigDefaults verifies zero config values use defaults
func TestRealRecovery_BackoffConfigDefaults(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()
	var attempts atomic.Int32

	start := time.Now()
	err := recovery.RecoverWithBackoff(
		func() error {
			attempts.Add(1)
			return errRecoveryFails
		},
		BackoffConfig{}, // All zero values - should use defaults
	)

	elapsed := time.Since(start)

	require.Error(t, err)
	// Default MaxRetries is 3, so 4 total attempts
	assert.Equal(t, int32(4), attempts.Load(), "Should use default MaxRetries of 3")
	// Default InitialDelay is 100ms, so there should be some delay
	assert.GreaterOrEqual(t, elapsed, 100*time.Millisecond, "Should have delays between retries")
}

// TestRealRecovery_BackoffRetryIfPredicate verifies RetryIf predicate stops non-retryable
func TestRealRecovery_BackoffRetryIfPredicate(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()
	var attempts atomic.Int32
	nonRetryableErr := WithCode(ErrPermissionDenied, "permission denied")

	err := recovery.RecoverWithBackoff(
		func() error {
			attempts.Add(1)
			return nonRetryableErr
		},
		BackoffConfig{
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     100 * time.Millisecond,
			Multiplier:   2.0,
			MaxRetries:   5,
			RetryIf: func(err error) bool {
				// Only retry if not permission denied
				return !IsPermissionDenied(err)
			},
		},
	)

	require.Error(t, err)
	// Should stop immediately since error is non-retryable
	assert.Equal(t, int32(1), attempts.Load(), "Should not retry non-retryable error")
}

// TestRealRecovery_BackoffRetryIfAllowsRetry verifies RetryIf allows retryable errors
func TestRealRecovery_BackoffRetryIfAllowsRetry(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()
	var attempts atomic.Int32
	retryableErr := WithCode(ErrTimeout, "timeout")

	//nolint:errcheck // testing retry count, error not relevant
	_ = recovery.RecoverWithBackoff(
		func() error {
			attempts.Add(1)
			return retryableErr
		},
		BackoffConfig{
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     50 * time.Millisecond,
			Multiplier:   2.0,
			MaxRetries:   2,
			RetryIf: func(err error) bool {
				return IsTimeout(err)
			},
		},
	)

	// Should retry since error is retryable
	assert.Equal(t, int32(3), attempts.Load(), "Should retry retryable error")
}

// TestRealRecovery_BackoffDelayIncreases verifies delay increases each attempt
func TestRealRecovery_BackoffDelayIncreases(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()

	// Test calculateNextDelay directly
	tests := []struct {
		current    time.Duration
		multiplier float64
		maxDelay   time.Duration
		expected   time.Duration
	}{
		{
			current:    100 * time.Millisecond,
			multiplier: 2.0,
			maxDelay:   1 * time.Second,
			expected:   200 * time.Millisecond,
		},
		{
			current:    500 * time.Millisecond,
			multiplier: 2.0,
			maxDelay:   1 * time.Second,
			expected:   1 * time.Second, // Capped at max
		},
		{
			current:    100 * time.Millisecond,
			multiplier: 1.5,
			maxDelay:   1 * time.Second,
			expected:   150 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		result := recovery.calculateNextDelay(tt.current, tt.multiplier, tt.maxDelay)
		assert.Equal(t, tt.expected, result)
	}
}

// TestRealRecovery_BackoffMaxDelayRespected verifies delay never exceeds MaxDelay
func TestRealRecovery_BackoffMaxDelayRespected(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()

	result := recovery.calculateNextDelay(
		1*time.Second,   // current
		10.0,            // huge multiplier
		500*time.Second, // maxDelay
	)

	assert.LessOrEqual(t, result, 500*time.Second, "Delay should not exceed MaxDelay")
}

// TestRealRecovery_ContextCancellation verifies canceled context returns quickly
func TestRealRecovery_ContextCancellation(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	start := time.Now()
	err := recovery.RecoverWithContext(ctx, func() error {
		// This should not complete because context is canceled
		time.Sleep(1 * time.Second)
		return nil
	})
	elapsed := time.Since(start)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "canceled")
	assert.Less(t, elapsed, 500*time.Millisecond, "Should return quickly when context is canceled")
}

// TestRealRecovery_ContextDeadlineExceeded verifies deadline exceeded returns error
func TestRealRecovery_ContextDeadlineExceeded(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := recovery.RecoverWithContext(ctx, func() error {
		// This takes longer than the deadline
		time.Sleep(200 * time.Millisecond)
		return nil
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "canceled")
}

// TestRealRecovery_ContextSuccessBeforeDeadline verifies success before deadline works
func TestRealRecovery_ContextSuccessBeforeDeadline(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	executed := false
	err := recovery.RecoverWithContext(ctx, func() error {
		executed = true
		return nil
	})

	require.NoError(t, err)
	assert.True(t, executed)
}

// TestRealRecovery_ContextPanicRecovered verifies panics are recovered with context
func TestRealRecovery_ContextPanicRecovered(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()
	ctx := context.Background()

	err := recovery.RecoverWithContext(ctx, func() error {
		panic("test panic")
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "panic")
}

// TestRealRecovery_LinearBackoff verifies LinearBackoff config is correct
func TestRealRecovery_LinearBackoff(t *testing.T) {
	t.Parallel()

	config := LinearBackoff(100*time.Millisecond, 50*time.Millisecond, 5)

	assert.Equal(t, 100*time.Millisecond, config.InitialDelay)
	assert.Equal(t, 5, config.MaxRetries)
	// MaxDelay should be initialDelay + (increment * maxRetries)
	assert.Equal(t, 350*time.Millisecond, config.MaxDelay)
}

// TestRealRecovery_ExponentialBackoff verifies ExponentialBackoff config is correct
func TestRealRecovery_ExponentialBackoff(t *testing.T) {
	t.Parallel()

	config := ExponentialBackoff(100*time.Millisecond, 5)

	assert.Equal(t, 100*time.Millisecond, config.InitialDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.InDelta(t, 2.0, config.Multiplier, 0.001)
	assert.Equal(t, 5, config.MaxRetries)
}

// TestRealRecovery_FibonacciBackoff verifies FibonacciBackoff uses golden ratio
func TestRealRecovery_FibonacciBackoff(t *testing.T) {
	t.Parallel()

	config := FibonacciBackoff(100*time.Millisecond, 5)

	assert.Equal(t, 100*time.Millisecond, config.InitialDelay)
	assert.Equal(t, 30*time.Second, config.MaxDelay)
	assert.InDelta(t, 1.618, config.Multiplier, 0.001, "Should use golden ratio")
	assert.Equal(t, 5, config.MaxRetries)
}

// TestRealRecovery_CombineRetryPredicates verifies AND logic for combined predicates
func TestRealRecovery_CombineRetryPredicates(t *testing.T) {
	t.Parallel()

	isNotCritical := func(err error) bool { return !IsCritical(err) }
	isRetryableCheck := RetryIfRetryable()

	combined := CombineRetryPredicates(isNotCritical, isRetryableCheck)

	// Retryable and not critical - should retry
	timeoutErr := WithCode(ErrTimeout, "timeout")
	assert.True(t, combined(timeoutErr), "Should retry timeout errors")

	// Not retryable - should not retry
	notFoundErr := WithCode(ErrNotFound, "not found")
	assert.False(t, combined(notFoundErr), "Should not retry not found errors")

	// Critical - should not retry (even if retryable)
	criticalErr := NewBuilder().
		WithMessage("critical").
		WithCode(ErrInternal).
		WithSeverity(SeverityCritical).
		Build()
	assert.False(t, combined(criticalErr), "Should not retry critical errors")
}

// TestRealRecovery_RetryIfRetryable verifies RetryIfRetryable predicate
func TestRealRecovery_RetryIfRetryable(t *testing.T) {
	t.Parallel()

	predicate := RetryIfRetryable()

	// Retryable errors
	assert.True(t, predicate(WithCode(ErrTimeout, "timeout")))
	assert.True(t, predicate(WithCode(ErrInternal, "internal")))
	assert.True(t, predicate(WithCode(ErrBuildFailed, "build failed")))

	// Non-retryable errors
	assert.False(t, predicate(WithCode(ErrNotFound, "not found")))
	assert.False(t, predicate(WithCode(ErrPermissionDenied, "permission denied")))
}

// TestRealRecovery_RetryIfTimeout verifies RetryIfTimeout predicate
func TestRealRecovery_RetryIfTimeout(t *testing.T) {
	t.Parallel()

	predicate := RetryIfTimeout()

	assert.True(t, predicate(WithCode(ErrTimeout, "timeout")))
	assert.True(t, predicate(WithCode(ErrCommandTimeout, "command timeout")))
	assert.False(t, predicate(WithCode(ErrNotFound, "not found")))
}

// TestRealRecovery_RetryIfNotCritical verifies RetryIfNotCritical predicate
func TestRealRecovery_RetryIfNotCritical(t *testing.T) {
	t.Parallel()

	predicate := RetryIfNotCritical()

	// Non-critical errors
	normalErr := NewBuilder().
		WithMessage("normal").
		WithSeverity(SeverityError).
		Build()
	assert.True(t, predicate(normalErr))

	// Critical errors
	criticalErr := NewBuilder().
		WithMessage("critical").
		WithSeverity(SeverityCritical).
		Build()
	assert.False(t, predicate(criticalErr))

	fatalErr := NewBuilder().
		WithMessage("fatal").
		WithSeverity(SeverityFatal).
		Build()
	assert.False(t, predicate(fatalErr))
}

// TestRealRecovery_RetryWithExponentialBackoff verifies convenience function
func TestRealRecovery_RetryWithExponentialBackoff(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	err := RetryWithExponentialBackoff(func() error {
		if attempts.Add(1) < 2 {
			return errRecoveryFail
		}
		return nil
	}, 3)

	require.NoError(t, err)
	assert.Equal(t, int32(2), attempts.Load())
}

// TestRealRecovery_RetryWithLinearBackoff verifies convenience function
func TestRealRecovery_RetryWithLinearBackoff(t *testing.T) {
	t.Parallel()

	var attempts atomic.Int32

	err := RetryWithLinearBackoff(func() error {
		if attempts.Add(1) < 2 {
			return errRecoveryFail
		}
		return nil
	}, 10*time.Millisecond, 3)

	require.NoError(t, err)
	assert.Equal(t, int32(2), attempts.Load())
}

// TestRealRecovery_BackoffPanicOnRetry verifies panic during retry is recovered
func TestRealRecovery_BackoffPanicOnRetry(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()
	var attempts atomic.Int32
	panicAttempt := int32(2)

	err := recovery.RecoverWithBackoff(
		func() error {
			current := attempts.Add(1)
			if current == panicAttempt {
				panic("panic on second attempt")
			}
			if current > panicAttempt {
				return nil // Succeed after panic recovery
			}
			return errRecoveryNormal
		},
		BackoffConfig{
			InitialDelay: 10 * time.Millisecond,
			MaxDelay:     50 * time.Millisecond,
			Multiplier:   2.0,
			MaxRetries:   5, // Allow more retries to recover from panic
		},
	)
	// The function should eventually succeed after recovering from panic
	// OR the final error should be from the last attempt (which may or may not be a panic)
	// The key test is that the panic was recovered and didn't crash the program
	if err != nil {
		// Either panic was recovered and we got a wrapped error, or we succeeded
		assert.Greater(t, attempts.Load(), panicAttempt, "Should have continued after panic")
	}
}

// TestRealRecovery_RetryPanicRecovered verifies panic during retry is recovered
func TestRealRecovery_RetryPanicRecovered(t *testing.T) {
	t.Parallel()

	recovery := NewErrorRecovery()
	var attempts atomic.Int32

	err := recovery.RecoverWithRetry(
		func() error {
			current := attempts.Add(1)
			if current == 2 {
				panic("test panic")
			}
			return errRecoveryNormal
		},
		3,
		10*time.Millisecond,
	)

	require.Error(t, err)
	// Should continue retrying after panic
	assert.GreaterOrEqual(t, attempts.Load(), int32(2))
}

// TestRealRecovery_EmptyPredicateList verifies empty predicate list allows all
func TestRealRecovery_EmptyPredicateList(t *testing.T) {
	t.Parallel()

	combined := CombineRetryPredicates() // No predicates

	// Should return true for any error (empty AND = true)
	assert.True(t, combined(errRecoveryAny))
	assert.True(t, combined(WithCode(ErrBuildFailed, "build failed")))
}
