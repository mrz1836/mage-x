//nolint:errcheck // Some error values intentionally discarded in tests
package exec

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"
)

func TestNewBase(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		opts    []Option
		wantDir string
		wantEnv []string
	}{
		{
			name:    "default",
			opts:    nil,
			wantDir: "",
			wantEnv: nil,
		},
		{
			name:    "with working dir",
			opts:    []Option{WithWorkingDir("/tmp")},
			wantDir: "/tmp",
			wantEnv: nil,
		},
		{
			name:    "with env",
			opts:    []Option{WithEnv([]string{"FOO=bar"})},
			wantDir: "",
			wantEnv: []string{"FOO=bar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			b := NewBase(tt.opts...)
			if b.WorkingDir != tt.wantDir {
				t.Errorf("WorkingDir = %q, want %q", b.WorkingDir, tt.wantDir)
			}
			if len(b.Env) != len(tt.wantEnv) {
				t.Errorf("Env len = %d, want %d", len(b.Env), len(tt.wantEnv))
			}
		})
	}
}

func TestBase_Execute(t *testing.T) {
	t.Parallel()

	b := NewBase()
	ctx := context.Background()

	// Test successful command
	err := b.Execute(ctx, "echo", "hello")
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}

	// Test failing command
	err = b.Execute(ctx, "false")
	if err == nil {
		t.Error("Execute() expected error for failing command")
	}
}

func TestBase_ExecuteOutput(t *testing.T) {
	t.Parallel()

	b := NewBase()
	ctx := context.Background()

	// Test successful command
	output, err := b.ExecuteOutput(ctx, "echo", "hello")
	if err != nil {
		t.Errorf("ExecuteOutput() error = %v, want nil", err)
	}
	if !strings.Contains(output, "hello") {
		t.Errorf("ExecuteOutput() output = %q, want to contain 'hello'", output)
	}

	// Test failing command
	_, err = b.ExecuteOutput(ctx, "false")
	if err == nil {
		t.Error("ExecuteOutput() expected error for failing command")
	}
}

func TestBase_ExecuteStreaming(t *testing.T) {
	t.Parallel()

	b := NewBase()
	ctx := context.Background()

	var stdout, stderr bytes.Buffer
	err := b.ExecuteStreaming(ctx, &stdout, &stderr, "echo", "streaming")
	if err != nil {
		t.Errorf("ExecuteStreaming() error = %v, want nil", err)
	}
	if !strings.Contains(stdout.String(), "streaming") {
		t.Errorf("ExecuteStreaming() stdout = %q, want to contain 'streaming'", stdout.String())
	}
}

func TestBase_ExecuteWithEnv(t *testing.T) {
	t.Parallel()

	b := NewBase()
	ctx := context.Background()

	// Test command with environment
	err := b.ExecuteWithEnv(ctx, []string{"TEST_VAR=test_value"}, "sh", "-c", "echo $TEST_VAR")
	if err != nil {
		t.Errorf("ExecuteWithEnv() error = %v, want nil", err)
	}
}

func TestBase_ExecuteInDir(t *testing.T) {
	t.Parallel()

	b := NewBase()
	ctx := context.Background()

	// Test command in /tmp directory
	err := b.ExecuteInDir(ctx, "/tmp", "pwd")
	if err != nil {
		t.Errorf("ExecuteInDir() error = %v, want nil", err)
	}
}

func TestBase_ExecuteOutputInDir(t *testing.T) {
	t.Parallel()

	b := NewBase()
	ctx := context.Background()

	// Test command in /tmp directory
	output, err := b.ExecuteOutputInDir(ctx, "/tmp", "pwd")
	if err != nil {
		t.Errorf("ExecuteOutputInDir() error = %v, want nil", err)
	}
	if !strings.Contains(output, "/tmp") && !strings.Contains(output, "/private/tmp") {
		t.Errorf("ExecuteOutputInDir() output = %q, want to contain '/tmp'", output)
	}
}

func TestBase_ContextCancellation(t *testing.T) {
	t.Parallel()

	b := NewBase()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	// This command should be interrupted
	err := b.Execute(ctx, "sleep", "10")
	if err == nil {
		t.Error("Execute() expected error for canceled context")
	}
}

func TestBase_Verbose(t *testing.T) {
	t.Parallel()

	var logged bool
	logger := func(format string, args ...interface{}) {
		logged = true
	}

	b := NewBase(
		WithVerbose(true),
		WithLogger(logger),
	)
	ctx := context.Background()

	_ = b.Execute(ctx, "echo", "test")

	if !logged {
		t.Error("Expected verbose logging to be called")
	}
}
