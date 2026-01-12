package retry

import (
	"context"
	"errors"
	"net"
	"syscall"
	"testing"
	"time"
)

// TestReset_Coverage tests Reset() methods for coverage
func TestReset_Coverage(t *testing.T) {
	t.Run("ConstantBackoff", func(t *testing.T) {
		b := &ConstantBackoff{Delay: 100 * time.Millisecond}
		b.Reset() // Call Reset for coverage
		if d := b.Duration(0); d != 100*time.Millisecond {
			t.Errorf("Duration = %v, want 100ms", d)
		}
	})

	t.Run("ExponentialBackoff", func(t *testing.T) {
		b := &ExponentialBackoff{
			Initial: 10 * time.Millisecond,
			Max:     1 * time.Second,
		}
		b.Reset() // Call Reset for coverage
		if d := b.Duration(0); d != 10*time.Millisecond {
			t.Errorf("Duration = %v, want 10ms", d)
		}
	})

	t.Run("LinearBackoff", func(t *testing.T) {
		b := &LinearBackoff{
			Initial:   50 * time.Millisecond,
			Increment: 25 * time.Millisecond,
		}
		b.Reset() // Call Reset for coverage
		if d := b.Duration(0); d != 50*time.Millisecond {
			t.Errorf("Duration = %v, want 50ms", d)
		}
	})
}

// TestIsRetriable_DNSErrors tests DNS error handling
func TestIsRetriable_DNSErrors(t *testing.T) {
	classifier := &DefaultClassifier{
		CheckDNSErrors: true,
	}

	t.Run("DNS no such host - not retriable", func(t *testing.T) {
		dnsErr := &net.DNSError{
			Err:         "no such host",
			IsTemporary: false,
		}
		if classifier.IsRetriable(dnsErr) {
			t.Error("Should not retry permanent 'no such host' DNS error")
		}
	})

	t.Run("DNS temporary failure - retriable", func(t *testing.T) {
		dnsErr := &net.DNSError{
			Err:         "no such host: temporary failure",
			IsTemporary: true,
		}
		if !classifier.IsRetriable(dnsErr) {
			t.Error("Should retry temporary DNS error")
		}
	})

	t.Run("DNS timeout - retriable", func(t *testing.T) {
		dnsErr := &net.DNSError{
			Err:         "i/o timeout",
			IsTimeout:   true,
			IsTemporary: true,
		}
		if !classifier.IsRetriable(dnsErr) {
			t.Error("Should retry DNS timeout error")
		}
	})
}

// TestIsRetriable_OpErrors tests operation error handling
func TestIsRetriable_OpErrors(t *testing.T) {
	classifier := &DefaultClassifier{
		CheckOpErrors: true,
	}

	t.Run("OpError - retriable", func(t *testing.T) {
		// Use syscall error instead of dynamic error
		opErr := &net.OpError{
			Op:  "dial",
			Net: "tcp",
			Err: syscall.ECONNREFUSED,
		}
		if !classifier.IsRetriable(opErr) {
			t.Error("Should retry net.OpError")
		}
	})
}

// TestIsRetriable_SyscallErrors tests syscall error handling
func TestIsRetriable_SyscallErrors(t *testing.T) {
	classifier := &DefaultClassifier{
		SyscallErrors: []syscall.Errno{syscall.ECONNREFUSED, syscall.ECONNRESET},
	}

	t.Run("ECONNREFUSED - retriable", func(t *testing.T) {
		err := syscall.ECONNREFUSED
		if !classifier.IsRetriable(err) {
			t.Error("Should retry ECONNREFUSED")
		}
	})

	t.Run("ECONNRESET - retriable", func(t *testing.T) {
		err := syscall.ECONNRESET
		if !classifier.IsRetriable(err) {
			t.Error("Should retry ECONNRESET")
		}
	})
}

// TestDoWithData_ContextCanceled tests DoWithData with canceled context
func TestDoWithData_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := DoWithData(ctx, nil, func() (string, error) {
		return "result", nil
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}

// TestDoWithData_WithNilConfig tests DoWithData with nil config
func TestDoWithData_WithNilConfig(t *testing.T) {
	ctx := context.Background()

	result, err := DoWithData(ctx, nil, func() (int, error) {
		return 42, nil
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != 42 {
		t.Errorf("Expected result 42, got %d", result)
	}
}

// TestExponentialBackoff_EdgeCases tests edge cases for exponential backoff
func TestExponentialBackoff_EdgeCases(t *testing.T) {
	t.Run("negative attempt", func(t *testing.T) {
		b := &ExponentialBackoff{
			Initial: 100 * time.Millisecond,
		}
		// Should treat negative as 0
		if d := b.Duration(-1); d != 100*time.Millisecond {
			t.Errorf("Duration(-1) = %v, want 100ms", d)
		}
	})

	t.Run("jitter results in negative", func(t *testing.T) {
		b := &ExponentialBackoff{
			Initial: 1 * time.Nanosecond,
			Jitter:  10.0, // Large jitter
		}
		// Should not return negative duration
		d := b.Duration(0)
		if d < 0 {
			t.Errorf("Duration should not be negative, got %v", d)
		}
	})
}

// TestLinearBackoff_EdgeCases tests edge cases for linear backoff
func TestLinearBackoff_EdgeCases(t *testing.T) {
	t.Run("negative attempt", func(t *testing.T) {
		b := &LinearBackoff{
			Initial:   100 * time.Millisecond,
			Increment: 50 * time.Millisecond,
		}
		// Should treat negative as 0
		if d := b.Duration(-1); d != 100*time.Millisecond {
			t.Errorf("Duration(-1) = %v, want 100ms", d)
		}
	})
}
