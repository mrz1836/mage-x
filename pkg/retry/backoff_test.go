package retry

import (
	"testing"
	"time"
)

func TestConstantBackoff(t *testing.T) {
	b := &ConstantBackoff{Delay: 100 * time.Millisecond}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 100 * time.Millisecond},
		{1, 100 * time.Millisecond},
		{5, 100 * time.Millisecond},
		{100, 100 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := b.Duration(tt.attempt)
			if got != tt.want {
				t.Errorf("Duration(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}

	// Test Reset (should be no-op)
	b.Reset()
}

func TestExponentialBackoff_Duration(t *testing.T) {
	b := &ExponentialBackoff{
		Initial:    100 * time.Millisecond,
		Max:        10 * time.Second,
		Multiplier: 2.0,
		Jitter:     0, // Disable jitter for deterministic tests
	}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 100 * time.Millisecond},  // 100ms * 2^0 = 100ms
		{1, 200 * time.Millisecond},  // 100ms * 2^1 = 200ms
		{2, 400 * time.Millisecond},  // 100ms * 2^2 = 400ms
		{3, 800 * time.Millisecond},  // 100ms * 2^3 = 800ms
		{4, 1600 * time.Millisecond}, // 100ms * 2^4 = 1600ms
		{5, 3200 * time.Millisecond}, // 100ms * 2^5 = 3200ms
		{6, 6400 * time.Millisecond}, // 100ms * 2^6 = 6400ms
		{7, 10 * time.Second},        // Capped at Max
		{10, 10 * time.Second},       // Still capped
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := b.Duration(tt.attempt)
			if got != tt.want {
				t.Errorf("Duration(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

func TestExponentialBackoff_DefaultMultiplier(t *testing.T) {
	b := &ExponentialBackoff{
		Initial:    100 * time.Millisecond,
		Multiplier: 0, // Should default to 2.0
	}

	got := b.Duration(1)
	want := 200 * time.Millisecond // 100ms * 2^1
	if got != want {
		t.Errorf("Duration(1) with default multiplier = %v, want %v", got, want)
	}
}

func TestExponentialBackoff_NegativeAttempt(t *testing.T) {
	b := &ExponentialBackoff{
		Initial:    100 * time.Millisecond,
		Multiplier: 2.0,
	}

	got := b.Duration(-1)
	want := 100 * time.Millisecond // Should treat as attempt 0
	if got != want {
		t.Errorf("Duration(-1) = %v, want %v", got, want)
	}
}

func TestExponentialBackoff_Jitter(t *testing.T) {
	b := &ExponentialBackoff{
		Initial:    1 * time.Second,
		Multiplier: 1.0, // No exponential growth, just jitter
		Jitter:     0.5, // ±50%
	}

	// Run multiple times and check that values are within expected range
	minExpected := 500 * time.Millisecond  // 1s - 50%
	maxExpected := 1500 * time.Millisecond // 1s + 50%

	for i := 0; i < 100; i++ {
		got := b.Duration(0)
		if got < minExpected || got > maxExpected {
			t.Errorf("Duration(0) with jitter = %v, expected between %v and %v", got, minExpected, maxExpected)
		}
	}
}

func TestExponentialBackoff_JitterDistribution(t *testing.T) {
	b := &ExponentialBackoff{
		Initial:    1 * time.Second,
		Multiplier: 1.0,
		Jitter:     0.1, // ±10%
	}

	// Collect samples
	samples := make([]time.Duration, 100)
	for i := range samples {
		samples[i] = b.Duration(0)
	}

	// Check that we get some variation (not all the same value)
	allSame := true
	for i := 1; i < len(samples); i++ {
		if samples[i] != samples[0] {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("Jitter should produce variation in delays")
	}
}

func TestLinearBackoff(t *testing.T) {
	b := &LinearBackoff{
		Initial:   100 * time.Millisecond,
		Increment: 50 * time.Millisecond,
		Max:       500 * time.Millisecond,
	}

	tests := []struct {
		attempt int
		want    time.Duration
	}{
		{0, 100 * time.Millisecond},  // 100ms + 0*50ms
		{1, 150 * time.Millisecond},  // 100ms + 1*50ms
		{2, 200 * time.Millisecond},  // 100ms + 2*50ms
		{3, 250 * time.Millisecond},  // 100ms + 3*50ms
		{8, 500 * time.Millisecond},  // Capped at Max
		{10, 500 * time.Millisecond}, // Still capped
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := b.Duration(tt.attempt)
			if got != tt.want {
				t.Errorf("Duration(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

func TestLinearBackoff_NegativeAttempt(t *testing.T) {
	b := &LinearBackoff{
		Initial:   100 * time.Millisecond,
		Increment: 50 * time.Millisecond,
	}

	got := b.Duration(-1)
	want := 100 * time.Millisecond // Should treat as attempt 0
	if got != want {
		t.Errorf("Duration(-1) = %v, want %v", got, want)
	}
}

func TestDefaultBackoff(t *testing.T) {
	b := DefaultBackoff()

	if b.Initial != 100*time.Millisecond {
		t.Errorf("DefaultBackoff().Initial = %v, want 100ms", b.Initial)
	}

	if b.Max != 30*time.Second {
		t.Errorf("DefaultBackoff().Max = %v, want 30s", b.Max)
	}

	if b.Multiplier != 2.0 {
		t.Errorf("DefaultBackoff().Multiplier = %v, want 2.0", b.Multiplier)
	}

	if b.Jitter != 0.1 {
		t.Errorf("DefaultBackoff().Jitter = %v, want 0.1", b.Jitter)
	}
}

func TestFastBackoff(t *testing.T) {
	b := FastBackoff()

	if b.Initial != 10*time.Millisecond {
		t.Errorf("FastBackoff().Initial = %v, want 10ms", b.Initial)
	}

	if b.Max != 1*time.Second {
		t.Errorf("FastBackoff().Max = %v, want 1s", b.Max)
	}
}

func TestSlowBackoff(t *testing.T) {
	b := SlowBackoff()

	if b.Initial != 1*time.Second {
		t.Errorf("SlowBackoff().Initial = %v, want 1s", b.Initial)
	}

	if b.Max != 5*time.Minute {
		t.Errorf("SlowBackoff().Max = %v, want 5m", b.Max)
	}
}

func TestNoDelay(t *testing.T) {
	b := NoDelay()

	for i := 0; i < 10; i++ {
		got := b.Duration(i)
		if got != 0 {
			t.Errorf("NoDelay().Duration(%d) = %v, want 0", i, got)
		}
	}
}

func BenchmarkExponentialBackoff(b *testing.B) {
	backoff := DefaultBackoff()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backoff.Duration(i % 10)
	}
}

func BenchmarkLinearBackoff(b *testing.B) {
	backoff := &LinearBackoff{
		Initial:   100 * time.Millisecond,
		Increment: 50 * time.Millisecond,
		Max:       1 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		backoff.Duration(i % 10)
	}
}
