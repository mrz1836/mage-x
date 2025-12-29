package exec

import (
	"bytes"
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

func TestBuilder_FullExecutorChain(t *testing.T) {
	t.Parallel()

	t.Run("Build returns FullExecutor", func(t *testing.T) {
		t.Parallel()
		// Verify Build() returns FullExecutor interface
		executor := NewBuilder().
			WithValidation().
			WithEnvFiltering().
			WithRetry(2).
			WithAdaptiveTimeout().
			Build()

		if executor == nil {
			t.Error("Build() returned nil")
		}
	})

	t.Run("Secure returns FullExecutor", func(t *testing.T) {
		t.Parallel()
		executor := Secure()
		if executor == nil {
			t.Error("Secure() returned nil")
		}
	})

	t.Run("SecureWithRetry returns FullExecutor", func(t *testing.T) {
		t.Parallel()
		executor := SecureWithRetry(3)
		if executor == nil {
			t.Error("SecureWithRetry() returned nil")
		}
	})
}

func TestBuilder_AllMethodsWork(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Create a full chain with all decorators
	executor := NewBuilder().
		WithValidation().
		WithEnvFiltering().
		WithRetry(1).
		WithTimeout(30 * time.Second).
		Build()

	t.Run("Execute works through chain", func(t *testing.T) {
		t.Parallel()
		err := executor.Execute(ctx, "echo", "hello")
		if err != nil {
			t.Errorf("Execute() error = %v", err)
		}
	})

	t.Run("ExecuteOutput works through chain", func(t *testing.T) {
		t.Parallel()
		output, err := executor.ExecuteOutput(ctx, "echo", "test output")
		if err != nil {
			t.Errorf("ExecuteOutput() error = %v", err)
		}
		if output == "" {
			t.Error("ExecuteOutput() returned empty output")
		}
	})

	t.Run("ExecuteInDir works through chain", func(t *testing.T) {
		t.Parallel()
		err := executor.ExecuteInDir(ctx, "/tmp", "pwd")
		if err != nil {
			t.Errorf("ExecuteInDir() error = %v", err)
		}
	})

	t.Run("ExecuteOutputInDir works through chain", func(t *testing.T) {
		t.Parallel()
		output, err := executor.ExecuteOutputInDir(ctx, "/tmp", "pwd")
		if err != nil {
			t.Errorf("ExecuteOutputInDir() error = %v", err)
		}
		if output == "" {
			t.Error("ExecuteOutputInDir() returned empty output")
		}
	})

	t.Run("ExecuteWithEnv works through chain", func(t *testing.T) {
		t.Parallel()
		err := executor.ExecuteWithEnv(ctx, []string{"TEST_VAR=value"}, "echo", "hello")
		if err != nil {
			t.Errorf("ExecuteWithEnv() error = %v", err)
		}
	})

	t.Run("ExecuteStreaming works through chain", func(t *testing.T) {
		t.Parallel()
		var stdout, stderr bytes.Buffer
		err := executor.ExecuteStreaming(ctx, &stdout, &stderr, "echo", "streaming test")
		if err != nil {
			t.Errorf("ExecuteStreaming() error = %v", err)
		}
		if stdout.Len() == 0 {
			t.Error("ExecuteStreaming() wrote nothing to stdout")
		}
	})
}

func TestBuilder_ValidationAppliesToAllMethods(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Create executor with validation
	executor := NewBuilder().
		WithValidation().
		Build()

	dangerousArg := "$(whoami)"

	t.Run("Execute rejects dangerous args", func(t *testing.T) {
		t.Parallel()
		err := executor.Execute(ctx, "echo", dangerousArg)
		if err == nil {
			t.Error("Execute() should reject dangerous argument")
		}
	})

	t.Run("ExecuteOutput rejects dangerous args", func(t *testing.T) {
		t.Parallel()
		_, err := executor.ExecuteOutput(ctx, "echo", dangerousArg)
		if err == nil {
			t.Error("ExecuteOutput() should reject dangerous argument")
		}
	})

	t.Run("ExecuteInDir rejects dangerous args", func(t *testing.T) {
		t.Parallel()
		err := executor.ExecuteInDir(ctx, "/tmp", "echo", dangerousArg)
		if err == nil {
			t.Error("ExecuteInDir() should reject dangerous argument")
		}
	})

	t.Run("ExecuteOutputInDir rejects dangerous args", func(t *testing.T) {
		t.Parallel()
		_, err := executor.ExecuteOutputInDir(ctx, "/tmp", "echo", dangerousArg)
		if err == nil {
			t.Error("ExecuteOutputInDir() should reject dangerous argument")
		}
	})

	t.Run("ExecuteWithEnv rejects dangerous args", func(t *testing.T) {
		t.Parallel()
		err := executor.ExecuteWithEnv(ctx, nil, "echo", dangerousArg)
		if err == nil {
			t.Error("ExecuteWithEnv() should reject dangerous argument")
		}
	})

	t.Run("ExecuteStreaming rejects dangerous args", func(t *testing.T) {
		t.Parallel()
		var stdout, stderr bytes.Buffer
		err := executor.ExecuteStreaming(ctx, &stdout, &stderr, "echo", dangerousArg)
		if err == nil {
			t.Error("ExecuteStreaming() should reject dangerous argument")
		}
	})
}
