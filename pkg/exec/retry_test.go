package exec

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mrz1836/mage-x/pkg/retry"
)

// Test error sentinels for err113 compliance
var (
	errTemporaryFailure  = errors.New("temporary failure")
	errPersistentFailure = errors.New("persistent failure")
	errKeepRetrying      = errors.New("keep retrying")
	errSpecificError     = errors.New("specific error")
	errAlwaysFail        = errors.New("always fail")
	errNonRetriable      = errors.New("non-retriable error")
)

// mockFullExecutor is a mock FullExecutor for testing
type mockFullExecutor struct {
	executeFunc          func(ctx context.Context, name string, args ...string) error
	executeOutputFunc    func(ctx context.Context, name string, args ...string) (string, error)
	executeWithEnvFunc   func(ctx context.Context, env []string, name string, args ...string) error
	executeInDirFunc     func(ctx context.Context, dir, name string, args ...string) error
	executeOutputInDir   func(ctx context.Context, dir, name string, args ...string) (string, error)
	executeStreamingFunc func(ctx context.Context, stdout, stderr io.Writer, name string, args ...string) error
}

func (m *mockFullExecutor) Execute(ctx context.Context, name string, args ...string) error {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, name, args...)
	}
	return nil
}

func (m *mockFullExecutor) ExecuteOutput(ctx context.Context, name string, args ...string) (string, error) {
	if m.executeOutputFunc != nil {
		return m.executeOutputFunc(ctx, name, args...)
	}
	return "", nil
}

func (m *mockFullExecutor) ExecuteWithEnv(ctx context.Context, env []string, name string, args ...string) error {
	if m.executeWithEnvFunc != nil {
		return m.executeWithEnvFunc(ctx, env, name, args...)
	}
	return nil
}

func (m *mockFullExecutor) ExecuteInDir(ctx context.Context, dir, name string, args ...string) error {
	if m.executeInDirFunc != nil {
		return m.executeInDirFunc(ctx, dir, name, args...)
	}
	return nil
}

func (m *mockFullExecutor) ExecuteOutputInDir(ctx context.Context, dir, name string, args ...string) (string, error) {
	if m.executeOutputInDir != nil {
		return m.executeOutputInDir(ctx, dir, name, args...)
	}
	return "", nil
}

func (m *mockFullExecutor) ExecuteStreaming(ctx context.Context, stdout, stderr io.Writer, name string, args ...string) error {
	if m.executeStreamingFunc != nil {
		return m.executeStreamingFunc(ctx, stdout, stderr, name, args...)
	}
	return nil
}

// Ensure mockFullExecutor implements FullExecutor
var _ FullExecutor = (*mockFullExecutor)(nil)

func TestRetryingExecutor_Execute(t *testing.T) {
	t.Parallel()

	t.Run("succeeds on first attempt", func(t *testing.T) {
		t.Parallel()
		var callCount int32
		mock := &mockFullExecutor{
			executeFunc: func(_ context.Context, _ string, _ ...string) error {
				atomic.AddInt32(&callCount, 1)
				return nil
			},
		}

		executor := NewRetryingExecutor(mock, WithMaxRetries(3))
		err := executor.Execute(context.Background(), "echo", "hello")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if atomic.LoadInt32(&callCount) != 1 {
			t.Errorf("expected 1 call, got %d", callCount)
		}
	})

	t.Run("retries on retriable error and succeeds", func(t *testing.T) {
		t.Parallel()
		var callCount int32
		mock := &mockFullExecutor{
			executeFunc: func(_ context.Context, _ string, _ ...string) error {
				count := atomic.AddInt32(&callCount, 1)
				if count < 3 {
					return errTemporaryFailure
				}
				return nil
			},
		}

		// Use AlwaysRetry classifier to ensure retries happen
		executor := NewRetryingExecutor(mock,
			WithMaxRetries(5),
			WithClassifier(retry.AlwaysRetry),
			WithBackoff(&retry.ConstantBackoff{Delay: time.Millisecond}),
		)
		err := executor.Execute(context.Background(), "echo", "hello")
		if err != nil {
			t.Errorf("expected no error after retries, got %v", err)
		}
		if atomic.LoadInt32(&callCount) != 3 {
			t.Errorf("expected 3 calls, got %d", callCount)
		}
	})

	t.Run("fails after max retries", func(t *testing.T) {
		t.Parallel()
		var callCount int32
		expectedErr := errPersistentFailure
		mock := &mockFullExecutor{
			executeFunc: func(_ context.Context, _ string, _ ...string) error {
				atomic.AddInt32(&callCount, 1)
				return expectedErr
			},
		}

		executor := NewRetryingExecutor(mock,
			WithMaxRetries(2),
			WithClassifier(retry.AlwaysRetry),
			WithBackoff(&retry.ConstantBackoff{Delay: time.Millisecond}),
		)
		err := executor.Execute(context.Background(), "echo", "hello")

		if err == nil {
			t.Error("expected error after max retries")
		}
		// MaxRetries=2 means 3 total attempts (initial + 2 retries)
		if atomic.LoadInt32(&callCount) != 3 {
			t.Errorf("expected 3 calls, got %d", callCount)
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		t.Parallel()
		var callCount int32
		mock := &mockFullExecutor{
			executeFunc: func(_ context.Context, _ string, _ ...string) error {
				atomic.AddInt32(&callCount, 1)
				return errKeepRetrying
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		executor := NewRetryingExecutor(mock,
			WithMaxRetries(100),
			WithClassifier(retry.AlwaysRetry),
			WithBackoff(&retry.ConstantBackoff{Delay: 10 * time.Millisecond}),
		)

		// Cancel after a short delay
		go func() {
			time.Sleep(25 * time.Millisecond)
			cancel()
		}()

		err := executor.Execute(ctx, "echo", "hello")

		if err == nil {
			t.Error("expected error from context cancellation")
		}
		// Should have made at least 1 call but less than 100
		count := atomic.LoadInt32(&callCount)
		if count < 1 || count >= 10 {
			t.Errorf("expected between 1 and 10 calls before cancellation, got %d", count)
		}
	})
}

func TestRetryingExecutor_ExecuteOutput(t *testing.T) {
	t.Parallel()

	t.Run("returns output on success", func(t *testing.T) {
		t.Parallel()
		mock := &mockFullExecutor{
			executeOutputFunc: func(_ context.Context, _ string, _ ...string) (string, error) {
				return "test output", nil
			},
		}

		executor := NewRetryingExecutor(mock, WithMaxRetries(3))
		output, err := executor.ExecuteOutput(context.Background(), "echo", "hello")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if output != "test output" {
			t.Errorf("expected 'test output', got '%s'", output)
		}
	})

	t.Run("retries and returns output", func(t *testing.T) {
		t.Parallel()
		var callCount int32
		mock := &mockFullExecutor{
			executeOutputFunc: func(_ context.Context, _ string, _ ...string) (string, error) {
				count := atomic.AddInt32(&callCount, 1)
				if count < 2 {
					return "", errTemporaryFailure
				}
				return "success after retry", nil
			},
		}

		executor := NewRetryingExecutor(mock,
			WithMaxRetries(3),
			WithClassifier(retry.AlwaysRetry),
			WithBackoff(&retry.ConstantBackoff{Delay: time.Millisecond}),
		)
		output, err := executor.ExecuteOutput(context.Background(), "echo", "hello")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if output != "success after retry" {
			t.Errorf("expected 'success after retry', got '%s'", output)
		}
		if atomic.LoadInt32(&callCount) != 2 {
			t.Errorf("expected 2 calls, got %d", callCount)
		}
	})
}

func TestRetryingExecutor_ExecuteInDir(t *testing.T) {
	t.Parallel()

	t.Run("passes directory to wrapped executor", func(t *testing.T) {
		t.Parallel()
		var capturedDir string
		mock := &mockFullExecutor{
			executeInDirFunc: func(_ context.Context, dir, _ string, _ ...string) error {
				capturedDir = dir
				return nil
			},
		}

		executor := NewRetryingExecutor(mock, WithMaxRetries(3))
		err := executor.ExecuteInDir(context.Background(), "/test/dir", "echo", "hello")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if capturedDir != "/test/dir" {
			t.Errorf("expected '/test/dir', got '%s'", capturedDir)
		}
	})

	t.Run("retries on failure in directory", func(t *testing.T) {
		t.Parallel()
		var callCount int32
		mock := &mockFullExecutor{
			executeInDirFunc: func(_ context.Context, _, _ string, _ ...string) error {
				count := atomic.AddInt32(&callCount, 1)
				if count < 2 {
					return errTemporaryFailure
				}
				return nil
			},
		}

		executor := NewRetryingExecutor(mock,
			WithMaxRetries(3),
			WithClassifier(retry.AlwaysRetry),
			WithBackoff(&retry.ConstantBackoff{Delay: time.Millisecond}),
		)
		err := executor.ExecuteInDir(context.Background(), "/test/dir", "echo", "hello")
		if err != nil {
			t.Errorf("expected no error after retry, got %v", err)
		}
		if atomic.LoadInt32(&callCount) != 2 {
			t.Errorf("expected 2 calls, got %d", callCount)
		}
	})
}

func TestRetryingExecutor_ExecuteWithEnv(t *testing.T) {
	t.Parallel()

	t.Run("passes environment to wrapped executor", func(t *testing.T) {
		t.Parallel()
		var capturedEnv []string
		mock := &mockFullExecutor{
			executeWithEnvFunc: func(_ context.Context, env []string, _ string, _ ...string) error {
				capturedEnv = env
				return nil
			},
		}

		executor := NewRetryingExecutor(mock, WithMaxRetries(3))
		env := []string{"FOO=bar", "BAZ=qux"}
		err := executor.ExecuteWithEnv(context.Background(), env, "echo", "hello")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(capturedEnv) != 2 || capturedEnv[0] != "FOO=bar" {
			t.Errorf("expected env to be passed, got %v", capturedEnv)
		}
	})

	t.Run("retries with environment", func(t *testing.T) {
		t.Parallel()
		var callCount int32
		mock := &mockFullExecutor{
			executeWithEnvFunc: func(_ context.Context, _ []string, _ string, _ ...string) error {
				count := atomic.AddInt32(&callCount, 1)
				if count < 2 {
					return errTemporaryFailure
				}
				return nil
			},
		}

		executor := NewRetryingExecutor(mock,
			WithMaxRetries(3),
			WithClassifier(retry.AlwaysRetry),
			WithBackoff(&retry.ConstantBackoff{Delay: time.Millisecond}),
		)
		err := executor.ExecuteWithEnv(context.Background(), []string{"FOO=bar"}, "echo", "hello")
		if err != nil {
			t.Errorf("expected no error after retry, got %v", err)
		}
		if atomic.LoadInt32(&callCount) != 2 {
			t.Errorf("expected 2 calls, got %d", callCount)
		}
	})
}

func TestRetryingExecutor_ExecuteOutputInDir(t *testing.T) {
	t.Parallel()

	t.Run("returns output from directory execution", func(t *testing.T) {
		t.Parallel()
		var capturedDir string
		mock := &mockFullExecutor{
			executeOutputInDir: func(_ context.Context, dir, _ string, _ ...string) (string, error) {
				capturedDir = dir
				return "output from " + dir, nil
			},
		}

		executor := NewRetryingExecutor(mock, WithMaxRetries(3))
		output, err := executor.ExecuteOutputInDir(context.Background(), "/test/dir", "echo", "hello")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if output != "output from /test/dir" {
			t.Errorf("expected 'output from /test/dir', got '%s'", output)
		}
		if capturedDir != "/test/dir" {
			t.Errorf("expected '/test/dir', got '%s'", capturedDir)
		}
	})
}

func TestRetryingExecutor_ExecuteStreaming(t *testing.T) {
	t.Parallel()

	t.Run("passes stdout/stderr to wrapped executor", func(t *testing.T) {
		t.Parallel()
		var capturedStdout, capturedStderr io.Writer
		mock := &mockFullExecutor{
			executeStreamingFunc: func(_ context.Context, stdout, stderr io.Writer, _ string, _ ...string) error {
				capturedStdout = stdout
				capturedStderr = stderr
				return nil
			},
		}

		executor := NewRetryingExecutor(mock, WithMaxRetries(3))
		stdout := &bytes.Buffer{}
		stderr := &bytes.Buffer{}
		err := executor.ExecuteStreaming(context.Background(), stdout, stderr, "echo", "hello")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if capturedStdout != stdout || capturedStderr != stderr {
			t.Error("stdout/stderr not passed correctly")
		}
	})

	t.Run("retries streaming execution", func(t *testing.T) {
		t.Parallel()
		var callCount int32
		mock := &mockFullExecutor{
			executeStreamingFunc: func(_ context.Context, _, _ io.Writer, _ string, _ ...string) error {
				count := atomic.AddInt32(&callCount, 1)
				if count < 2 {
					return errTemporaryFailure
				}
				return nil
			},
		}

		executor := NewRetryingExecutor(mock,
			WithMaxRetries(3),
			WithClassifier(retry.AlwaysRetry),
			WithBackoff(&retry.ConstantBackoff{Delay: time.Millisecond}),
		)
		err := executor.ExecuteStreaming(context.Background(), &bytes.Buffer{}, &bytes.Buffer{}, "echo", "hello")
		if err != nil {
			t.Errorf("expected no error after retry, got %v", err)
		}
		if atomic.LoadInt32(&callCount) != 2 {
			t.Errorf("expected 2 calls, got %d", callCount)
		}
	})
}

func TestRetryingExecutor_OnRetry(t *testing.T) {
	t.Parallel()

	t.Run("calls OnRetry callback before each retry", func(t *testing.T) {
		t.Parallel()
		var callCount int32
		var retryAttempts []int
		mock := &mockFullExecutor{
			executeFunc: func(_ context.Context, _ string, _ ...string) error {
				count := atomic.AddInt32(&callCount, 1)
				if count < 4 {
					return errTemporaryFailure
				}
				return nil
			},
		}

		executor := NewRetryingExecutor(mock,
			WithMaxRetries(5),
			WithClassifier(retry.AlwaysRetry),
			WithBackoff(&retry.ConstantBackoff{Delay: time.Millisecond}),
			WithOnRetry(func(attempt int, _ error, _ time.Duration) {
				retryAttempts = append(retryAttempts, attempt)
			}),
		)
		err := executor.Execute(context.Background(), "echo", "hello")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		// Should have 3 retries (attempts 1, 2, 3 after initial attempt 0)
		if len(retryAttempts) != 3 {
			t.Errorf("expected 3 retry callbacks, got %d: %v", len(retryAttempts), retryAttempts)
		}
	})

	t.Run("OnRetry receives error from previous attempt", func(t *testing.T) {
		t.Parallel()
		expectedErr := errSpecificError
		var receivedErr error
		mock := &mockFullExecutor{
			executeFunc: func(_ context.Context, _ string, _ ...string) error {
				return expectedErr
			},
		}

		executor := NewRetryingExecutor(mock,
			WithMaxRetries(1),
			WithClassifier(retry.AlwaysRetry),
			WithBackoff(&retry.ConstantBackoff{Delay: time.Millisecond}),
			WithOnRetry(func(_ int, err error, _ time.Duration) {
				receivedErr = err
			}),
		)
		_ = executor.Execute(context.Background(), "echo", "hello") //nolint:errcheck // intentionally ignoring error to test OnRetry callback

		if receivedErr == nil || receivedErr.Error() != expectedErr.Error() {
			t.Errorf("expected error '%v' in OnRetry, got '%v'", expectedErr, receivedErr)
		}
	})
}

func TestRetryingExecutor_MaxRetries(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name          string
		maxRetries    int
		expectedCalls int32
	}{
		{"zero retries means 1 attempt", 0, 1},
		{"one retry means 2 attempts", 1, 2},
		{"three retries means 4 attempts", 3, 4},
	}

	for _, tc := range testCases {
		// capture
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var callCount int32
			mock := &mockFullExecutor{
				executeFunc: func(_ context.Context, _ string, _ ...string) error {
					atomic.AddInt32(&callCount, 1)
					return errAlwaysFail
				},
			}

			executor := NewRetryingExecutor(mock,
				WithMaxRetries(tc.maxRetries),
				WithClassifier(retry.AlwaysRetry),
				WithBackoff(&retry.ConstantBackoff{Delay: time.Millisecond}),
			)
			_ = executor.Execute(context.Background(), "echo", "hello") //nolint:errcheck // intentionally ignoring error to count attempts

			if atomic.LoadInt32(&callCount) != tc.expectedCalls {
				t.Errorf("expected %d calls, got %d", tc.expectedCalls, callCount)
			}
		})
	}
}

func TestRetryingExecutor_Classifier(t *testing.T) {
	t.Parallel()

	t.Run("does not retry non-retriable errors", func(t *testing.T) {
		t.Parallel()
		var callCount int32
		mock := &mockFullExecutor{
			executeFunc: func(_ context.Context, _ string, _ ...string) error {
				atomic.AddInt32(&callCount, 1)
				return errNonRetriable
			},
		}

		// Use NeverRetry classifier
		executor := NewRetryingExecutor(mock,
			WithMaxRetries(5),
			WithClassifier(retry.NeverRetry),
			WithBackoff(&retry.ConstantBackoff{Delay: time.Millisecond}),
		)
		err := executor.Execute(context.Background(), "echo", "hello")

		if err == nil {
			t.Error("expected error")
		}
		// Should only make 1 call since error is not retriable
		if atomic.LoadInt32(&callCount) != 1 {
			t.Errorf("expected 1 call (no retries), got %d", callCount)
		}
	})

	t.Run("uses CommandClassifier for command errors", func(t *testing.T) {
		t.Parallel()
		// CommandClassifier should be available as a package-level variable
		if CommandClassifier == nil {
			t.Error("CommandClassifier should not be nil")
		}
	})
}

func TestRetryingExecutor_ImplementsFullExecutor(t *testing.T) {
	t.Parallel()

	// This test verifies at compile-time that RetryingExecutor implements FullExecutor
	// The assignment below would fail compilation if RetryingExecutor didn't implement FullExecutor
	var _ FullExecutor = NewRetryingExecutor(NewBase())
}

func TestNewRetryingExecutor_Defaults(t *testing.T) {
	t.Parallel()

	executor := NewRetryingExecutor(NewBase())

	if executor.MaxRetries != 3 {
		t.Errorf("expected default MaxRetries=3, got %d", executor.MaxRetries)
	}
	if executor.Classifier == nil {
		t.Error("expected default Classifier to be set")
	}
	if executor.Backoff == nil {
		t.Error("expected default Backoff to be set")
	}
}
