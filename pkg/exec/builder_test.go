package exec

import (
	"context"
	"testing"
	"time"
)

func TestBuilder(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("basic builder", func(t *testing.T) {
		t.Parallel()
		executor := NewBuilder().Build()

		err := executor.Execute(ctx, "echo", "hello")
		if err != nil {
			t.Errorf("Execute() error = %v, want nil", err)
		}
	})

	t.Run("builder with validation", func(t *testing.T) {
		t.Parallel()
		executor := NewBuilder().
			WithValidation(WithAllowedCommands([]string{"echo"})).
			Build()

		// Allowed command
		err := executor.Execute(ctx, "echo", "hello")
		if err != nil {
			t.Errorf("Execute() error = %v, want nil", err)
		}

		// Disallowed command
		err = executor.Execute(ctx, "ls")
		if err == nil {
			t.Error("Execute() expected error for disallowed command")
		}
	})

	t.Run("builder with timeout", func(t *testing.T) {
		t.Parallel()
		executor := NewBuilder().
			WithTimeout(5 * time.Second).
			Build()

		err := executor.Execute(ctx, "echo", "hello")
		if err != nil {
			t.Errorf("Execute() error = %v, want nil", err)
		}
	})

	t.Run("builder with adaptive timeout", func(t *testing.T) {
		t.Parallel()
		executor := NewBuilder().
			WithAdaptiveTimeout().
			Build()

		err := executor.Execute(ctx, "echo", "hello")
		if err != nil {
			t.Errorf("Execute() error = %v, want nil", err)
		}
	})

	t.Run("builder with retry", func(t *testing.T) {
		t.Parallel()
		executor := NewBuilder().
			WithRetry(3).
			Build()

		err := executor.Execute(ctx, "echo", "hello")
		if err != nil {
			t.Errorf("Execute() error = %v, want nil", err)
		}
	})

	t.Run("builder with env filtering", func(t *testing.T) {
		t.Parallel()
		executor := NewBuilder().
			WithEnvFiltering().
			Build()

		err := executor.Execute(ctx, "echo", "hello")
		if err != nil {
			t.Errorf("Execute() error = %v, want nil", err)
		}
	})

	t.Run("builder with all features", func(t *testing.T) {
		t.Parallel()
		executor := NewBuilder().
			WithWorkingDirectory("/tmp").
			WithVerbose(false).
			WithValidation().
			WithEnvFiltering().
			WithRetry(2).
			WithAdaptiveTimeout().
			Build()

		err := executor.Execute(ctx, "pwd")
		if err != nil {
			t.Errorf("Execute() error = %v, want nil", err)
		}
	})
}

func TestSecure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	executor := Secure()

	// Valid command
	err := executor.Execute(ctx, "echo", "hello")
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	// Dangerous argument should be rejected
	err = executor.Execute(ctx, "echo", "$(whoami)")
	if err == nil {
		t.Error("Execute() expected error for dangerous pattern")
	}
}

func TestSecureWithRetry(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	executor := SecureWithRetry(3)

	err := executor.Execute(ctx, "echo", "hello")
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
}

func TestSimple(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	executor := Simple()

	err := executor.Execute(ctx, "echo", "hello")
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
}
