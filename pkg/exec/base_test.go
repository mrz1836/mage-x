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

func TestBase_logVerbose(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		verbose    bool
		logger     func(string, ...interface{})
		wantLogged bool
	}{
		{
			name:       "verbose enabled with logger logs message",
			verbose:    true,
			logger:     func(string, ...interface{}) {},
			wantLogged: true,
		},
		{
			name:       "verbose disabled does not log",
			verbose:    false,
			logger:     func(string, ...interface{}) {},
			wantLogged: false,
		},
		{
			name:       "verbose enabled with nil logger does not panic",
			verbose:    true,
			logger:     nil,
			wantLogged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var logged bool
			var loggedMsg string
			var wrappedLogger func(string, ...interface{})
			if tt.logger != nil {
				wrappedLogger = func(format string, args ...interface{}) {
					logged = true
					loggedMsg = format
				}
			}

			b := &Base{
				Verbose: tt.verbose,
				logger:  wrappedLogger,
			}

			b.logVerbose("test message: %s", "arg")

			if logged != tt.wantLogged {
				t.Errorf("logged = %v, want %v", logged, tt.wantLogged)
			}
			if tt.wantLogged && !strings.Contains(loggedMsg, "test message") {
				t.Errorf("loggedMsg = %q, want to contain 'test message'", loggedMsg)
			}
		})
	}
}

func TestBase_checkDryRun(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		dryRun     bool
		logger     func(string, ...interface{})
		wantReturn bool
		wantLogged bool
	}{
		{
			name:       "dryRun enabled with logger returns true and logs",
			dryRun:     true,
			logger:     func(string, ...interface{}) {},
			wantReturn: true,
			wantLogged: true,
		},
		{
			name:       "dryRun disabled returns false",
			dryRun:     false,
			logger:     func(string, ...interface{}) {},
			wantReturn: false,
			wantLogged: false,
		},
		{
			name:       "dryRun enabled with nil logger returns true without panic",
			dryRun:     true,
			logger:     nil,
			wantReturn: true,
			wantLogged: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var logged bool
			var wrappedLogger func(string, ...interface{})
			if tt.logger != nil {
				wrappedLogger = func(format string, args ...interface{}) {
					logged = true
				}
			}

			b := &Base{
				DryRun: tt.dryRun,
				logger: wrappedLogger,
			}

			result := b.checkDryRun("[DRY RUN] test: %s", "arg")

			if result != tt.wantReturn {
				t.Errorf("checkDryRun() = %v, want %v", result, tt.wantReturn)
			}
			if logged != tt.wantLogged {
				t.Errorf("logged = %v, want %v", logged, tt.wantLogged)
			}
		})
	}
}

func TestBase_DryRunSkipsExecution(t *testing.T) {
	t.Parallel()

	var logMessages []string
	logger := func(format string, args ...interface{}) {
		logMessages = append(logMessages, format)
	}

	b := NewBase(
		WithDryRun(true),
		WithVerbose(true),
		WithLogger(logger),
	)
	ctx := context.Background()

	// DryRun should return early without executing
	output, err := b.ExecuteOutput(ctx, "echo", "should_not_run")
	if err != nil {
		t.Errorf("ExecuteOutput() with DryRun error = %v, want nil", err)
	}
	if output != "" {
		t.Errorf("ExecuteOutput() with DryRun output = %q, want empty", output)
	}

	// Should have logged verbose and dry-run messages
	if len(logMessages) != 2 {
		t.Errorf("Expected 2 log messages, got %d: %v", len(logMessages), logMessages)
	}
}
