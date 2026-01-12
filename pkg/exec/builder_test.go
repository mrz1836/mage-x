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

func TestBuilder_WithBaseEnv(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("base env variables are set", func(t *testing.T) {
		t.Parallel()
		executor := NewBuilder().
			WithBaseEnv([]string{"TEST_VAR=test_value"}).
			Build()

		// Execute command that uses the environment variable
		output, err := executor.ExecuteOutput(ctx, "sh", "-c", "echo $TEST_VAR")
		if err != nil {
			t.Errorf("ExecuteOutput() error = %v", err)
		}
		if output == "" {
			t.Error("ExecuteOutput() should use base environment variable")
		}
	})

	t.Run("multiple base env variables", func(t *testing.T) {
		t.Parallel()
		executor := NewBuilder().
			WithBaseEnv([]string{"VAR1=value1", "VAR2=value2"}).
			Build()

		err := executor.Execute(ctx, "echo", "hello")
		if err != nil {
			t.Errorf("Execute() error = %v", err)
		}
	})
}

func TestBuilder_WithTimeoutResolver(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("custom timeout resolver", func(t *testing.T) {
		t.Parallel()
		customResolver := TimeoutResolverFunc(func(name string, args []string) time.Duration {
			return 10 * time.Second
		})

		executor := NewBuilder().
			WithTimeoutResolver(customResolver).
			Build()

		err := executor.Execute(ctx, "echo", "hello")
		if err != nil {
			t.Errorf("Execute() error = %v", err)
		}
	})

	t.Run("resolver allows fast command", func(t *testing.T) {
		t.Parallel()
		customResolver := TimeoutResolverFunc(func(name string, args []string) time.Duration {
			return 5 * time.Second
		})

		executor := NewBuilder().
			WithTimeoutResolver(customResolver).
			Build()

		err := executor.Execute(ctx, "echo", "fast command")
		if err != nil {
			t.Errorf("Execute() should succeed with adequate timeout: %v", err)
		}
	})
}

func TestBuilder_WithDryRun(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("dry run mode enabled", func(t *testing.T) {
		t.Parallel()
		executor := NewBuilder().
			WithDryRun(true).
			Build()

		// In dry run mode, commands should not actually execute
		err := executor.Execute(ctx, "echo", "hello")
		if err != nil {
			t.Errorf("Execute() in dry run should not error: %v", err)
		}
	})

	t.Run("dry run mode disabled", func(t *testing.T) {
		t.Parallel()
		executor := NewBuilder().
			WithDryRun(false).
			Build()

		err := executor.Execute(ctx, "echo", "hello")
		if err != nil {
			t.Errorf("Execute() error = %v", err)
		}
	})
}

// mockAuditLoggerBuilder is a simple mock for builder tests
type mockAuditLoggerBuilder struct {
	logged *bool
}

func (m *mockAuditLoggerBuilder) LogEvent(_ AuditEvent) error {
	*m.logged = true
	return nil
}

func TestBuilder_WithAuditLogging(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("audit logging with custom logger", func(t *testing.T) {
		t.Parallel()
		logged := false
		customLogger := &mockAuditLoggerBuilder{logged: &logged}

		executor := NewBuilder().
			WithAuditLogging(customLogger).
			Build()

		err := executor.Execute(ctx, "echo", "test")
		if err != nil {
			t.Errorf("Execute() error = %v", err)
		}
		if !logged {
			t.Error("Custom audit logger was not called")
		}
	})

	t.Run("audit logging with nil logger uses default", func(t *testing.T) {
		t.Parallel()
		executor := NewBuilder().
			WithAuditLogging(nil).
			Build()

		err := executor.Execute(ctx, "echo", "test")
		if err != nil {
			t.Errorf("Execute() error = %v", err)
		}
	})
}

// capturingAuditLogger captures the audit event for testing
type capturingAuditLogger struct {
	event *AuditEvent
}

func (c *capturingAuditLogger) LogEvent(event AuditEvent) error {
	*c.event = event
	return nil
}

func TestBuilder_WithAuditMetadata(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("single metadata entry", func(t *testing.T) {
		t.Parallel()
		var capturedEvent AuditEvent
		customLogger := &capturingAuditLogger{event: &capturedEvent}

		executor := NewBuilder().
			WithAuditLogging(customLogger).
			WithAuditMetadata("user", "testuser").
			Build()

		err := executor.Execute(ctx, "echo", "test")
		if err != nil {
			t.Errorf("Execute() error = %v", err)
		}

		if capturedEvent.Metadata["user"] != "testuser" {
			t.Errorf("Audit metadata 'user' = %v, want 'testuser'", capturedEvent.Metadata["user"])
		}
	})

	t.Run("multiple metadata entries", func(t *testing.T) {
		t.Parallel()
		var capturedEvent AuditEvent
		customLogger := &capturingAuditLogger{event: &capturedEvent}

		executor := NewBuilder().
			WithAuditLogging(customLogger).
			WithAuditMetadata("user", "testuser").
			WithAuditMetadata("session", "abc123").
			WithAuditMetadata("env", "test").
			Build()

		err := executor.Execute(ctx, "echo", "test")
		if err != nil {
			t.Errorf("Execute() error = %v", err)
		}

		if capturedEvent.Metadata["user"] != "testuser" {
			t.Errorf("Audit metadata 'user' = %v, want 'testuser'", capturedEvent.Metadata["user"])
		}
		if capturedEvent.Metadata["session"] != "abc123" {
			t.Errorf("Audit metadata 'session' = %v, want 'abc123'", capturedEvent.Metadata["session"])
		}
		if capturedEvent.Metadata["env"] != "test" {
			t.Errorf("Audit metadata 'env' = %v, want 'test'", capturedEvent.Metadata["env"])
		}
	})
}
