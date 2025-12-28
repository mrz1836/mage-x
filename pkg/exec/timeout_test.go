package exec

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestTimeoutExecutor(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("command completes within timeout", func(t *testing.T) {
		t.Parallel()
		base := NewBase()
		te := NewTimeoutExecutor(base, WithDefaultTimeout(5*time.Second))

		err := te.Execute(ctx, "echo", "hello")
		if err != nil {
			t.Errorf("Execute() error = %v, want nil", err)
		}
	})

	t.Run("command exceeds timeout", func(t *testing.T) {
		t.Parallel()
		base := NewBase()
		te := NewTimeoutExecutor(base, WithDefaultTimeout(10*time.Millisecond))

		err := te.Execute(ctx, "sleep", "1")
		if err == nil {
			t.Error("Execute() expected timeout error")
		}
		// The command is killed when context times out, resulting in "signal: killed"
		// We check for either DeadlineExceeded or killed signal
		errStr := err.Error()
		if !errors.Is(err, context.DeadlineExceeded) &&
			!strings.Contains(errStr, "killed") &&
			!strings.Contains(errStr, "deadline") {
			t.Errorf("Execute() error = %v, want timeout-related error", err)
		}
	})

	t.Run("output completes within timeout", func(t *testing.T) {
		t.Parallel()
		base := NewBase()
		te := NewTimeoutExecutor(base, WithDefaultTimeout(5*time.Second))

		output, err := te.ExecuteOutput(ctx, "echo", "hello")
		if err != nil {
			t.Errorf("ExecuteOutput() error = %v, want nil", err)
		}
		if output == "" {
			t.Error("ExecuteOutput() output is empty")
		}
	})
}

func TestAdaptiveTimeoutResolver(t *testing.T) {
	t.Parallel()

	resolver := NewAdaptiveTimeoutResolver()

	tests := []struct {
		name    string
		cmdName string
		args    []string
		wantGte time.Duration
		wantLte time.Duration
	}{
		{
			name:    "go test",
			cmdName: "go",
			args:    []string{"test", "./..."},
			wantGte: 10 * time.Minute,
			wantLte: 10 * time.Minute,
		},
		{
			name:    "go build",
			cmdName: "go",
			args:    []string{"build", "-o", "main"},
			wantGte: 3 * time.Minute,
			wantLte: 3 * time.Minute,
		},
		{
			name:    "golangci-lint",
			cmdName: "golangci-lint",
			args:    []string{"run"},
			wantGte: 20 * time.Minute,
			wantLte: 20 * time.Minute,
		},
		{
			name:    "unknown command",
			cmdName: "unknown",
			args:    []string{},
			wantGte: 30 * time.Second,
			wantLte: 30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := resolver.GetTimeout(tt.cmdName, tt.args)
			if got < tt.wantGte || got > tt.wantLte {
				t.Errorf("GetTimeout(%s, %v) = %v, want between %v and %v",
					tt.cmdName, tt.args, got, tt.wantGte, tt.wantLte)
			}
		})
	}
}

func TestTimeoutExecutor_WithResolver(t *testing.T) {
	t.Parallel()

	base := NewBase()
	resolver := NewAdaptiveTimeoutResolver()
	te := NewTimeoutExecutor(base, WithTimeoutResolver(resolver))

	ctx := context.Background()
	err := te.Execute(ctx, "echo", "hello")
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
}

func TestTimeoutExecutor_ContextAlreadyHasDeadline(t *testing.T) {
	t.Parallel()

	base := NewBase()
	te := NewTimeoutExecutor(base, WithDefaultTimeout(5*time.Minute))

	// Create context with shorter deadline
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := te.Execute(ctx, "sleep", "1")
	if err == nil {
		t.Error("Execute() expected timeout error from context deadline")
	}
}

func TestTimeoutResolverFunc(t *testing.T) {
	t.Parallel()

	resolver := TimeoutResolverFunc(func(name string, args []string) time.Duration {
		if name == "fast" {
			return 1 * time.Second
		}
		return 5 * time.Minute
	})

	if got := resolver.GetTimeout("fast", nil); got != 1*time.Second {
		t.Errorf("GetTimeout(fast) = %v, want 1s", got)
	}
	if got := resolver.GetTimeout("slow", nil); got != 5*time.Minute {
		t.Errorf("GetTimeout(slow) = %v, want 5m", got)
	}
}
