package exec

import (
	"context"
	"time"

	"github.com/mrz1836/mage-x/pkg/retry"
)

// RetryingExecutor wraps an executor with retry support using the retry package
type RetryingExecutor struct {
	wrapped Executor

	// MaxRetries is the maximum number of retry attempts (not including the initial attempt)
	MaxRetries int

	// Classifier determines if an error is retriable
	Classifier retry.Classifier

	// Backoff determines the delay between retries
	Backoff retry.Backoff

	// OnRetry is called before each retry attempt (optional)
	OnRetry func(attempt int, err error, delay time.Duration)
}

// RetryOption configures a RetryingExecutor
type RetryOption func(*RetryingExecutor)

// WithMaxRetries sets the maximum retry count
func WithMaxRetries(n int) RetryOption {
	return func(r *RetryingExecutor) {
		r.MaxRetries = n
	}
}

// WithClassifier sets the error classifier
func WithClassifier(c retry.Classifier) RetryOption {
	return func(r *RetryingExecutor) {
		r.Classifier = c
	}
}

// WithBackoff sets the backoff strategy
func WithBackoff(b retry.Backoff) RetryOption {
	return func(r *RetryingExecutor) {
		r.Backoff = b
	}
}

// WithOnRetry sets the callback for retry attempts
func WithOnRetry(fn func(attempt int, err error, delay time.Duration)) RetryOption {
	return func(r *RetryingExecutor) {
		r.OnRetry = fn
	}
}

// NewRetryingExecutor creates a new retrying executor
func NewRetryingExecutor(wrapped Executor, opts ...RetryOption) *RetryingExecutor {
	r := &RetryingExecutor{
		wrapped:    wrapped,
		MaxRetries: 3,
		Classifier: retry.NewNetworkClassifier(),
		Backoff:    retry.DefaultBackoff(),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Execute runs a command with retry support
func (r *RetryingExecutor) Execute(ctx context.Context, name string, args ...string) error {
	cfg := &retry.Config{
		MaxAttempts: r.MaxRetries + 1,
		Classifier:  r.Classifier,
		Backoff:     r.Backoff,
		OnRetry:     r.OnRetry,
	}

	return retry.Do(ctx, cfg, func() error {
		return r.wrapped.Execute(ctx, name, args...)
	})
}

// ExecuteOutput runs a command with retry support and returns output
func (r *RetryingExecutor) ExecuteOutput(ctx context.Context, name string, args ...string) (string, error) {
	cfg := &retry.Config{
		MaxAttempts: r.MaxRetries + 1,
		Classifier:  r.Classifier,
		Backoff:     r.Backoff,
		OnRetry:     r.OnRetry,
	}

	return retry.DoWithData(ctx, cfg, func() (string, error) {
		return r.wrapped.ExecuteOutput(ctx, name, args...)
	})
}

// CommandClassifier is a retry.Classifier specialized for command execution errors.
// It extends the DefaultClassifier with additional command-specific patterns.
//
//nolint:gochecknoglobals // Intentional package-level singleton for convenience
var CommandClassifier = retry.NewCommandClassifier()

// Ensure RetryingExecutor implements Executor
var _ Executor = (*RetryingExecutor)(nil)

// ExecuteWithRetry executes a command with retry support, allowing per-call retry configuration.
// This is a convenience function for cases where you need different retry settings per call.
func ExecuteWithRetry(ctx context.Context, executor Executor, maxRetries int, initialDelay time.Duration, name string, args ...string) error {
	cfg := &retry.Config{
		MaxAttempts: maxRetries + 1,
		Classifier:  CommandClassifier,
		Backoff: &retry.ExponentialBackoff{
			Initial:    initialDelay,
			Max:        30 * time.Second,
			Multiplier: 2.0,
		},
	}
	return retry.Do(ctx, cfg, func() error {
		return executor.Execute(ctx, name, args...)
	})
}

// ExecuteOutputWithRetry executes a command with retry support and returns output.
// This is a convenience function for cases where you need different retry settings per call.
func ExecuteOutputWithRetry(ctx context.Context, executor Executor, maxRetries int, initialDelay time.Duration, name string, args ...string) (string, error) {
	cfg := &retry.Config{
		MaxAttempts: maxRetries + 1,
		Classifier:  CommandClassifier,
		Backoff: &retry.ExponentialBackoff{
			Initial:    initialDelay,
			Max:        30 * time.Second,
			Multiplier: 2.0,
		},
	}
	return retry.DoWithData(ctx, cfg, func() (string, error) {
		return executor.ExecuteOutput(ctx, name, args...)
	})
}
