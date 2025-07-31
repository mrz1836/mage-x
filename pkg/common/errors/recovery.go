package errors

import (
	"context"
	"crypto/rand"
	"math/big"
	"time"
)

// RealDefaultErrorRecovery is the actual implementation of ErrorRecovery
type RealDefaultErrorRecovery struct{}

// NewErrorRecovery creates a new error recovery handler
func NewErrorRecovery() *RealDefaultErrorRecovery {
	return &RealDefaultErrorRecovery{}
}

// Recover executes a function and recovers from panics
func (r *RealDefaultErrorRecovery) Recover(fn func() error) (err error) {
	defer func() {
		if rec := recover(); rec != nil {
			if recErr, ok := rec.(error); ok {
				err = Wrap(recErr, "panic recovered")
			} else {
				err = Newf("panic recovered: %v", rec)
			}
		}
	}()

	return fn()
}

// RecoverWithFallback executes a function with panic recovery and fallback
func (r *RealDefaultErrorRecovery) RecoverWithFallback(fn func() error, fallback func(error) error) error {
	err := r.Recover(fn)
	if err != nil && fallback != nil {
		return fallback(err)
	}
	return err
}

// RecoverWithRetry executes a function with panic recovery and retries
func (r *RealDefaultErrorRecovery) RecoverWithRetry(fn func() error, retries int, delay time.Duration) error {
	var lastErr error

	for i := 0; i <= retries; i++ {
		err := r.Recover(fn)
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't delay after the last attempt
		if i < retries {
			time.Sleep(delay)
		}
	}

	return Wrapf(lastErr, "failed after %d retries", retries+1)
}

// RecoverWithBackoff executes a function with exponential backoff
func (r *RealDefaultErrorRecovery) RecoverWithBackoff(fn func() error, config BackoffConfig) error {
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}
	if config.InitialDelay <= 0 {
		config.InitialDelay = 100 * time.Millisecond
	}
	if config.MaxDelay <= 0 {
		config.MaxDelay = 30 * time.Second
	}
	if config.Multiplier <= 0 {
		config.Multiplier = 2.0
	}

	var lastErr error
	delay := config.InitialDelay

	for i := 0; i <= config.MaxRetries; i++ {
		err := r.Recover(fn)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if config.RetryIf != nil && !config.RetryIf(err) {
			return err
		}

		// Don't delay after the last attempt
		if i < config.MaxRetries {
			// Add jitter to prevent thundering herd using crypto/rand
			maxJitter := int64(float64(delay) * 0.1)
			actualDelay := delay
			if maxJitter > 0 {
				jitterBig, err := rand.Int(rand.Reader, big.NewInt(maxJitter))
				if err != nil {
					// Fallback to no jitter if crypto/rand fails
					jitterBig = big.NewInt(0)
				}
				jitter := time.Duration(jitterBig.Int64())
				actualDelay = delay + jitter
			}

			time.Sleep(actualDelay)

			// Calculate next delay with backoff
			delay = time.Duration(float64(delay) * config.Multiplier)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}
	}

	return Wrapf(lastErr, "failed after %d retries with backoff", config.MaxRetries+1)
}

// RecoverWithContext executes a function with panic recovery and context
func (r *RealDefaultErrorRecovery) RecoverWithContext(ctx context.Context, fn func() error) error {
	type result struct {
		err error
	}

	resultChan := make(chan result, 1)

	go func() {
		resultChan <- result{err: r.Recover(fn)}
	}()

	select {
	case <-ctx.Done():
		return Wrap(ctx.Err(), "operation canceled")
	case res := <-resultChan:
		return res.err
	}
}

// Recover executes a function and recovers from panics, converting them to errors.
func (r *DefaultErrorRecovery) Recover(fn func() error) error {
	recovery := NewErrorRecovery()
	return recovery.Recover(fn)
}

// RecoverWithFallback executes a function with panic recovery and calls fallback on error.
func (r *DefaultErrorRecovery) RecoverWithFallback(fn func() error, fallback func(error) error) error {
	recovery := NewErrorRecovery()
	return recovery.RecoverWithFallback(fn, fallback)
}

// RecoverWithRetry executes a function with panic recovery and retries on failure.
func (r *DefaultErrorRecovery) RecoverWithRetry(fn func() error, retries int, delay time.Duration) error {
	recovery := NewErrorRecovery()
	return recovery.RecoverWithRetry(fn, retries, delay)
}

// RecoverWithBackoff executes a function with panic recovery and exponential backoff retries.
func (r *DefaultErrorRecovery) RecoverWithBackoff(fn func() error, config BackoffConfig) error {
	recovery := NewErrorRecovery()
	return recovery.RecoverWithBackoff(fn, config)
}

// RecoverWithContext executes a function with panic recovery and context support.
func (r *DefaultErrorRecovery) RecoverWithContext(ctx context.Context, fn func() error) error {
	recovery := NewErrorRecovery()
	return recovery.RecoverWithContext(ctx, fn)
}

// Helper functions for backoff strategies

// LinearBackoff creates a linear backoff configuration
func LinearBackoff(initialDelay time.Duration, increment time.Duration, maxRetries int) BackoffConfig {
	return BackoffConfig{
		InitialDelay: initialDelay,
		MaxDelay:     initialDelay + (increment * time.Duration(maxRetries)),
		Multiplier:   1.0 + (float64(increment) / float64(initialDelay)),
		MaxRetries:   maxRetries,
	}
}

// ExponentialBackoff creates an exponential backoff configuration
func ExponentialBackoff(initialDelay time.Duration, maxRetries int) BackoffConfig {
	return BackoffConfig{
		InitialDelay: initialDelay,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		MaxRetries:   maxRetries,
	}
}

// FibonacciBackoff creates a Fibonacci sequence backoff configuration
func FibonacciBackoff(initialDelay time.Duration, maxRetries int) BackoffConfig {
	return BackoffConfig{
		InitialDelay: initialDelay,
		MaxDelay:     30 * time.Second,
		Multiplier:   1.618, // Golden ratio
		MaxRetries:   maxRetries,
	}
}

// RetryIfRetryable returns a retry predicate that checks if error is retryable
func RetryIfRetryable() func(error) bool {
	return func(err error) bool {
		return IsRetryable(err)
	}
}

// RetryIfTimeout returns a retry predicate that checks for timeout errors
func RetryIfTimeout() func(error) bool {
	return func(err error) bool {
		return IsTimeout(err)
	}
}

// RetryIfNotCritical returns a retry predicate that excludes critical errors
func RetryIfNotCritical() func(error) bool {
	return func(err error) bool {
		return !IsCritical(err)
	}
}

// CombineRetryPredicates combines multiple retry predicates with AND logic
func CombineRetryPredicates(predicates ...func(error) bool) func(error) bool {
	return func(err error) bool {
		for _, predicate := range predicates {
			if !predicate(err) {
				return false
			}
		}
		return true
	}
}

// RetryWithExponentialBackoff is a convenience function for exponential backoff retry
func RetryWithExponentialBackoff(fn func() error, maxRetries int) error {
	return DefaultRecovery.RecoverWithBackoff(fn, ExponentialBackoff(100*time.Millisecond, maxRetries))
}

// RetryWithLinearBackoff is a convenience function for linear backoff retry
func RetryWithLinearBackoff(fn func() error, delay time.Duration, maxRetries int) error {
	return DefaultRecovery.RecoverWithBackoff(fn, LinearBackoff(delay, delay, maxRetries))
}
