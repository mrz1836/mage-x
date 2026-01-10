package retry

import (
	"math"
	"math/rand"
	"time"
)

// Backoff computes the delay before the next retry attempt.
type Backoff interface {
	// Duration returns the delay for the given attempt number (0-indexed).
	Duration(attempt int) time.Duration

	// Reset resets any internal state (for stateful backoff strategies).
	Reset()
}

// ConstantBackoff returns the same delay for every attempt.
type ConstantBackoff struct {
	Delay time.Duration
}

// Duration implements Backoff.
func (b *ConstantBackoff) Duration(_ int) time.Duration {
	return b.Delay
}

// Reset implements Backoff. No-op for ConstantBackoff as it has no internal state.
func (b *ConstantBackoff) Reset() {}

// ExponentialBackoff implements exponential backoff with optional jitter.
type ExponentialBackoff struct {
	// Initial is the delay for the first retry (attempt 0).
	Initial time.Duration

	// Max is the maximum delay (cap).
	Max time.Duration

	// Multiplier is the factor by which the delay increases each attempt.
	// Default is 2.0 if not set.
	Multiplier float64

	// Jitter is the maximum random jitter to add (0.0 to 1.0).
	// A jitter of 0.1 means ±10% randomization.
	Jitter float64
}

// Duration implements Backoff.
func (b *ExponentialBackoff) Duration(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	multiplier := b.Multiplier
	if multiplier <= 0 {
		multiplier = 2.0
	}

	// Calculate base delay: initial * multiplier^attempt
	delay := float64(b.Initial) * math.Pow(multiplier, float64(attempt))

	// Apply maximum cap
	if b.Max > 0 && time.Duration(delay) > b.Max {
		delay = float64(b.Max)
	}

	// Apply jitter
	if b.Jitter > 0 {
		jitterRange := delay * b.Jitter
		//nolint:gosec // Cryptographic randomness not needed for jitter
		jitter := (rand.Float64()*2 - 1) * jitterRange
		delay += jitter
	}

	// Ensure non-negative
	if delay < 0 {
		delay = 0
	}

	return time.Duration(delay)
}

// Reset implements Backoff. No-op for ExponentialBackoff as it calculates delays
// purely from the attempt number passed to Duration().
func (b *ExponentialBackoff) Reset() {}

// LinearBackoff increases the delay linearly with each attempt.
type LinearBackoff struct {
	// Initial is the delay for the first retry.
	Initial time.Duration

	// Increment is added to the delay for each subsequent attempt.
	Increment time.Duration

	// Max is the maximum delay (cap).
	Max time.Duration
}

// Duration implements Backoff.
func (b *LinearBackoff) Duration(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	delay := b.Initial + time.Duration(attempt)*b.Increment

	if b.Max > 0 && delay > b.Max {
		delay = b.Max
	}

	return delay
}

// Reset implements Backoff. No-op for LinearBackoff as it calculates delays
// purely from the attempt number passed to Duration().
func (b *LinearBackoff) Reset() {}

// DefaultBackoff returns a standard exponential backoff configuration:
// - Initial delay: 100ms
// - Maximum delay: 30s
// - Multiplier: 2.0
// - Jitter: 0.1 (±10%)
func DefaultBackoff() *ExponentialBackoff {
	return &ExponentialBackoff{
		Initial:    100 * time.Millisecond,
		Max:        30 * time.Second,
		Multiplier: 2.0,
		Jitter:     0.1,
	}
}

// FastBackoff returns an aggressive backoff for quick retries:
// - Initial delay: 10ms
// - Maximum delay: 1s
// - Multiplier: 2.0
// - Jitter: 0.2 (±20%)
func FastBackoff() *ExponentialBackoff {
	return &ExponentialBackoff{
		Initial:    10 * time.Millisecond,
		Max:        1 * time.Second,
		Multiplier: 2.0,
		Jitter:     0.2,
	}
}

// SlowBackoff returns a conservative backoff for expensive operations:
// - Initial delay: 1s
// - Maximum delay: 5m
// - Multiplier: 2.0
// - Jitter: 0.1 (±10%)
func SlowBackoff() *ExponentialBackoff {
	return &ExponentialBackoff{
		Initial:    1 * time.Second,
		Max:        5 * time.Minute,
		Multiplier: 2.0,
		Jitter:     0.1,
	}
}

// NoDelay returns a backoff that always returns zero delay.
// Useful for testing or when delay is handled externally.
func NoDelay() *ConstantBackoff {
	return &ConstantBackoff{Delay: 0}
}
