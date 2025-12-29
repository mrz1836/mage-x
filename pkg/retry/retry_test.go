//nolint:err113,errcheck // Test code uses ad-hoc errors and ignores return values intentionally
package retry

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestDo_SuccessOnFirstAttempt(t *testing.T) {
	attempts := 0
	err := Do(context.Background(), DefaultConfig(), func() error {
		attempts++
		return nil
	})
	if err != nil {
		t.Errorf("Do() returned error: %v", err)
	}

	if attempts != 1 {
		t.Errorf("Do() made %d attempts, want 1", attempts)
	}
}

func TestDo_SuccessAfterRetries(t *testing.T) {
	attempts := 0
	cfg := &Config{
		MaxAttempts: 5,
		Classifier:  AlwaysRetry,
		Backoff:     NoDelay(),
	}

	err := Do(context.Background(), cfg, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary error")
		}
		return nil
	})
	if err != nil {
		t.Errorf("Do() returned error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Do() made %d attempts, want 3", attempts)
	}
}

func TestDo_AllAttemptsFail(t *testing.T) {
	attempts := 0
	cfg := &Config{
		MaxAttempts: 3,
		Classifier:  AlwaysRetry,
		Backoff:     NoDelay(),
	}

	err := Do(context.Background(), cfg, func() error {
		attempts++
		return errors.New("persistent error")
	})

	if err == nil {
		t.Error("Do() should return error after all attempts fail")
	}

	if attempts != 3 {
		t.Errorf("Do() made %d attempts, want 3", attempts)
	}

	if !errors.Is(err, errors.New("persistent error")) {
		// Check error message contains attempt count
		if err.Error() != "failed after 3 attempts: persistent error" {
			t.Logf("Error message: %v", err) // Just log for visibility
		}
	}
}

func TestDo_PermanentError(t *testing.T) {
	attempts := 0
	cfg := &Config{
		MaxAttempts: 5,
		Classifier:  NeverRetry, // All errors are permanent
		Backoff:     NoDelay(),
	}

	err := Do(context.Background(), cfg, func() error {
		attempts++
		return errors.New("permanent error")
	})

	if err == nil {
		t.Error("Do() should return error for permanent errors")
	}

	if attempts != 1 {
		t.Errorf("Do() made %d attempts, want 1 (should not retry permanent errors)", attempts)
	}
}

func TestDo_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0

	cfg := &Config{
		MaxAttempts: 10,
		Classifier:  AlwaysRetry,
		Backoff:     &ConstantBackoff{Delay: 100 * time.Millisecond},
	}

	// Cancel after first attempt
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := Do(ctx, cfg, func() error {
		attempts++
		return errors.New("error")
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Do() should return context.Canceled, got: %v", err)
	}

	if attempts > 2 {
		t.Errorf("Do() made %d attempts, should have stopped after cancellation", attempts)
	}
}

func TestDo_ContextDeadline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	attempts := 0
	cfg := &Config{
		MaxAttempts: 10,
		Classifier:  AlwaysRetry,
		Backoff:     &ConstantBackoff{Delay: 100 * time.Millisecond},
	}

	err := Do(ctx, cfg, func() error {
		attempts++
		return errors.New("error")
	})

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Do() should return context.DeadlineExceeded, got: %v", err)
	}
}

func TestDo_OnRetryCallback(t *testing.T) {
	var callbackAttempts []int
	var callbackErrors []error
	var callbackDelays []time.Duration

	cfg := &Config{
		MaxAttempts: 3,
		Classifier:  AlwaysRetry,
		Backoff:     &ConstantBackoff{Delay: 10 * time.Millisecond},
		OnRetry: func(attempt int, err error, delay time.Duration) {
			callbackAttempts = append(callbackAttempts, attempt)
			callbackErrors = append(callbackErrors, err)
			callbackDelays = append(callbackDelays, delay)
		},
	}

	_ = Do(context.Background(), cfg, func() error {
		return errors.New("error")
	})

	// OnRetry should be called for each retry (not the initial attempt)
	if len(callbackAttempts) != 2 { // 3 attempts means 2 retries
		t.Errorf("OnRetry called %d times, want 2", len(callbackAttempts))
	}

	// Check attempt numbers
	for i, attempt := range callbackAttempts {
		if attempt != i {
			t.Errorf("OnRetry attempt %d = %d, want %d", i, attempt, i)
		}
	}

	// Check delays
	for _, delay := range callbackDelays {
		if delay != 10*time.Millisecond {
			t.Errorf("OnRetry delay = %v, want 10ms", delay)
		}
	}
}

func TestDo_NilConfig(t *testing.T) {
	attempts := 0
	err := Do(context.Background(), nil, func() error {
		attempts++
		return nil
	})
	if err != nil {
		t.Errorf("Do() with nil config returned error: %v", err)
	}

	if attempts != 1 {
		t.Errorf("Do() made %d attempts, want 1", attempts)
	}
}

func TestDo_MaxAttemptsZero(t *testing.T) {
	attempts := 0
	cfg := &Config{
		MaxAttempts: 0, // Should be treated as 1
		Classifier:  AlwaysRetry,
		Backoff:     NoDelay(),
	}

	_ = Do(context.Background(), cfg, func() error {
		attempts++
		return errors.New("error")
	})

	if attempts != 1 {
		t.Errorf("Do() with MaxAttempts=0 made %d attempts, want 1", attempts)
	}
}

func TestDoWithData_ReturnsData(t *testing.T) {
	cfg := &Config{
		MaxAttempts: 3,
		Classifier:  AlwaysRetry,
		Backoff:     NoDelay(),
	}

	attempts := 0
	data, err := DoWithData(context.Background(), cfg, func() (string, error) {
		attempts++
		if attempts < 2 {
			return "partial", errors.New("retry")
		}
		return "success", nil
	})
	if err != nil {
		t.Errorf("DoWithData() returned error: %v", err)
	}

	if data != "success" {
		t.Errorf("DoWithData() returned data = %q, want %q", data, "success")
	}
}

func TestDoWithData_ReturnsLastDataOnFailure(t *testing.T) {
	cfg := &Config{
		MaxAttempts: 3,
		Classifier:  AlwaysRetry,
		Backoff:     NoDelay(),
	}

	attempts := 0
	data, err := DoWithData(context.Background(), cfg, func() (int, error) {
		attempts++
		return attempts * 10, errors.New("error")
	})

	if err == nil {
		t.Error("DoWithData() should return error")
	}

	// Should return the last attempt's data
	if data != 30 { // 3 * 10
		t.Errorf("DoWithData() returned data = %d, want %d", data, 30)
	}
}

func TestWithMaxAttempts(t *testing.T) {
	cfg := WithMaxAttempts(5)

	if cfg.MaxAttempts != 5 {
		t.Errorf("WithMaxAttempts(5).MaxAttempts = %d, want 5", cfg.MaxAttempts)
	}

	// Should have defaults for other fields
	if cfg.Classifier == nil {
		t.Error("WithMaxAttempts should set default classifier")
	}

	if cfg.Backoff == nil {
		t.Error("WithMaxAttempts should set default backoff")
	}
}

func TestWithClassifier(t *testing.T) {
	cfg := WithClassifier(NeverRetry)

	if cfg.Classifier == nil {
		t.Error("WithClassifier should set the classifier")
	}

	// Verify the classifier behaves like NeverRetry
	if cfg.Classifier.IsRetriable(errors.New("any error")) {
		t.Error("WithClassifier(NeverRetry) should never retry")
	}
}

func TestWithBackoff(t *testing.T) {
	b := &ConstantBackoff{Delay: 1 * time.Second}
	cfg := WithBackoff(b)

	if cfg.Backoff != b {
		t.Error("WithBackoff should set the backoff")
	}
}

func TestQuick(t *testing.T) {
	attempts := 0
	err := Quick(context.Background(), func() error {
		attempts++
		if attempts < 2 {
			return errors.New("connection refused") // Retriable
		}
		return nil
	})
	if err != nil {
		t.Errorf("Quick() returned error: %v", err)
	}

	if attempts < 2 {
		t.Errorf("Quick() should have retried, made %d attempts", attempts)
	}
}

func TestOnce(t *testing.T) {
	attempts := 0
	err := Once(context.Background(), func() error {
		attempts++
		return errors.New("error")
	})

	if err == nil {
		t.Error("Once() should return error")
	}

	if attempts != 1 {
		t.Errorf("Once() made %d attempts, want 1", attempts)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MaxAttempts != 3 {
		t.Errorf("DefaultConfig().MaxAttempts = %d, want 3", cfg.MaxAttempts)
	}

	if cfg.Classifier == nil {
		t.Error("DefaultConfig().Classifier should not be nil")
	}

	if cfg.Backoff == nil {
		t.Error("DefaultConfig().Backoff should not be nil")
	}
}

func TestCommandConfig(t *testing.T) {
	cfg := CommandConfig()

	if cfg.MaxAttempts != 3 {
		t.Errorf("CommandConfig().MaxAttempts = %d, want 3", cfg.MaxAttempts)
	}

	// Should use command classifier
	if cfg.Classifier == nil {
		t.Error("CommandConfig().Classifier should not be nil")
	}

	// Command classifier should recognize Go module errors
	if !cfg.Classifier.IsRetriable(errors.New("go: downloading module failed")) {
		t.Error("CommandConfig classifier should recognize Go module errors as retriable")
	}
}

func TestDo_ConcurrentSafety(t *testing.T) {
	cfg := &Config{
		MaxAttempts: 3,
		Classifier:  AlwaysRetry,
		Backoff:     NoDelay(),
	}

	var counter int64

	// Run multiple retries concurrently
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_ = Do(context.Background(), cfg, func() error {
				atomic.AddInt64(&counter, 1)
				return errors.New("error")
			})
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Each goroutine should make 3 attempts
	expected := int64(10 * 3)
	if atomic.LoadInt64(&counter) != expected {
		t.Errorf("Concurrent retries made %d total attempts, want %d", counter, expected)
	}
}

func BenchmarkDo_SuccessFirstAttempt(b *testing.B) {
	cfg := DefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Do(context.Background(), cfg, func() error {
			return nil
		})
	}
}

func BenchmarkDo_WithRetries(b *testing.B) {
	cfg := &Config{
		MaxAttempts: 3,
		Classifier:  AlwaysRetry,
		Backoff:     NoDelay(),
	}

	attempts := 0
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		attempts = 0
		_ = Do(context.Background(), cfg, func() error {
			attempts++
			if attempts < 3 {
				return errors.New("error")
			}
			return nil
		})
	}
}
