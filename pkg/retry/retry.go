package retry

import (
	"context"
	"fmt"
	"time"
)

// Config holds retry configuration.
type Config struct {
	// MaxAttempts is the maximum number of attempts (including the initial attempt).
	// A value of 1 means no retries, 3 means up to 2 retries after the initial attempt.
	MaxAttempts int

	// Classifier determines which errors are retriable.
	// If nil, defaults to NewNetworkClassifier().
	Classifier Classifier

	// Backoff determines the delay between retries.
	// If nil, defaults to DefaultBackoff().
	Backoff Backoff

	// OnRetry is called before each retry attempt (optional).
	// The attempt parameter is 0-indexed (0 = first retry).
	OnRetry func(attempt int, err error, delay time.Duration)
}

// DefaultConfig returns a standard retry configuration:
// - MaxAttempts: 3
// - Classifier: NewNetworkClassifier()
// - Backoff: DefaultBackoff()
func DefaultConfig() *Config {
	return &Config{
		MaxAttempts: 3,
		Classifier:  NewNetworkClassifier(),
		Backoff:     DefaultBackoff(),
	}
}

// CommandConfig returns a retry configuration optimized for command execution:
// - MaxAttempts: 3
// - Classifier: NewCommandClassifier()
// - Backoff: DefaultBackoff()
func CommandConfig() *Config {
	return &Config{
		MaxAttempts: 3,
		Classifier:  NewCommandClassifier(),
		Backoff:     DefaultBackoff(),
	}
}

// Do executes the given function with retry logic.
// It returns the error from the last attempt if all attempts fail.
func Do(ctx context.Context, cfg *Config, fn func() error) error {
	_, err := DoWithData(ctx, cfg, func() (struct{}, error) {
		return struct{}{}, fn()
	})
	return err
}

// DoWithData executes the given function with retry logic and returns both data and error.
// This is useful when the function returns a value that you want to preserve even on failure.
func DoWithData[T any](ctx context.Context, cfg *Config, fn func() (T, error)) (T, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	classifier := cfg.Classifier
	if classifier == nil {
		classifier = NewNetworkClassifier()
	}

	backoff := cfg.Backoff
	if backoff == nil {
		backoff = DefaultBackoff()
	}

	maxAttempts := cfg.MaxAttempts
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastResult T
	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		// Check context before attempting
		if ctx.Err() != nil {
			return lastResult, ctx.Err()
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastResult = result
		lastErr = err

		// Check if error is retriable
		if !classifier.IsRetriable(err) {
			return lastResult, fmt.Errorf("permanent error (not retriable): %w", err)
		}

		// Don't delay after the last attempt
		if attempt < maxAttempts-1 {
			delay := backoff.Duration(attempt)

			// Call OnRetry callback if provided
			if cfg.OnRetry != nil {
				cfg.OnRetry(attempt, err, delay)
			}

			// Wait with context cancellation support
			select {
			case <-ctx.Done():
				return lastResult, ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return lastResult, fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}

// WithMaxAttempts creates a new Config with the specified max attempts.
func WithMaxAttempts(attempts int) *Config {
	cfg := DefaultConfig()
	cfg.MaxAttempts = attempts
	return cfg
}

// WithClassifier creates a new Config with the specified classifier.
func WithClassifier(c Classifier) *Config {
	cfg := DefaultConfig()
	cfg.Classifier = c
	return cfg
}

// WithBackoff creates a new Config with the specified backoff.
func WithBackoff(b Backoff) *Config {
	cfg := DefaultConfig()
	cfg.Backoff = b
	return cfg
}

// Quick executes with a fast retry configuration (3 attempts, fast backoff).
func Quick(ctx context.Context, fn func() error) error {
	cfg := &Config{
		MaxAttempts: 3,
		Classifier:  NewNetworkClassifier(),
		Backoff:     FastBackoff(),
	}
	return Do(ctx, cfg, fn)
}

// Once executes with no retries (useful for consistency in code that conditionally retries).
func Once(_ context.Context, fn func() error) error {
	return fn()
}
