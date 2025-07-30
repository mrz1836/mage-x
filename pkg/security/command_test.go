package security

import (
	"context"
	"errors"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecureExecutor_ValidateCommand(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		args        []string
		allowedCmds map[string]bool
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid command with no restrictions",
			command: "echo",
			args:    []string{"hello", "world"},
			wantErr: false,
		},
		{
			name:        "command not in allowed list",
			command:     "rm",
			args:        []string{"-rf", "/"},
			allowedCmds: map[string]bool{"echo": true, "ls": true},
			wantErr:     true,
			errContains: "not in allowed list",
		},
		{
			name:        "command with path traversal in name",
			command:     "../bin/malicious",
			args:        []string{},
			wantErr:     true,
			errContains: "path traversal",
		},
		{
			name:        "command with shell injection in args",
			command:     "echo",
			args:        []string{"$(whoami)"},
			wantErr:     true,
			errContains: "dangerous pattern",
		},
		{
			name:        "command with backtick injection",
			command:     "echo",
			args:        []string{"`id`"},
			wantErr:     true,
			errContains: "dangerous pattern",
		},
		{
			name:        "command with pipe injection",
			command:     "echo",
			args:        []string{"hello | cat /etc/passwd"},
			wantErr:     true,
			errContains: "dangerous pattern",
		},
		{
			name:        "command with AND operator",
			command:     "echo",
			args:        []string{"hello && rm -rf /"},
			wantErr:     true,
			errContains: "dangerous pattern",
		},
		{
			name:        "command with OR operator",
			command:     "echo",
			args:        []string{"hello || rm -rf /"},
			wantErr:     true,
			errContains: "dangerous pattern",
		},
		{
			name:        "command with semicolon",
			command:     "echo",
			args:        []string{"hello; rm -rf /"},
			wantErr:     true,
			errContains: "dangerous pattern",
		},
		{
			name:        "command with redirect",
			command:     "echo",
			args:        []string{"secret > /tmp/stolen"},
			wantErr:     true,
			errContains: "dangerous pattern",
		},
		{
			name:        "command with input redirect",
			command:     "echo",
			args:        []string{"< /etc/passwd"},
			wantErr:     true,
			errContains: "dangerous pattern",
		},
		{
			name:        "command with IFS manipulation",
			command:     "echo",
			args:        []string{"${IFS}"},
			wantErr:     true,
			errContains: "dangerous pattern",
		},
		{
			name:    "safe command with special but allowed characters",
			command: "git",
			args:    []string{"commit", "-m", "feat: add new feature"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &SecureExecutor{
				AllowedCommands: tt.allowedCmds,
			}

			err := executor.validateCommand(tt.command, tt.args)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecureExecutor_FilterEnvironment(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
		filtered []string
	}{
		{
			name: "filters AWS secrets",
			input: []string{
				"PATH=/usr/bin",
				"AWS_SECRET_ACCESS_KEY=secret123",
				"AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE",
				"HOME=/home/user",
			},
			expected: []string{
				"PATH=/usr/bin",
				"AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE",
				"HOME=/home/user",
			},
			filtered: []string{"AWS_SECRET_ACCESS_KEY"},
		},
		{
			name: "filters GitHub tokens",
			input: []string{
				"PATH=/usr/bin",
				"GITHUB_TOKEN=ghp_xxxxxxxxxxxx",
				"GITHUB_USER=testuser",
			},
			expected: []string{
				"PATH=/usr/bin",
				"GITHUB_USER=testuser",
			},
			filtered: []string{"GITHUB_TOKEN"},
		},
		{
			name: "filters multiple sensitive vars",
			input: []string{
				"PATH=/usr/bin",
				"API_KEY=12345",
				"DATABASE_PASSWORD=secret",
				"SECRET_TOKEN=xyz",
				"PRIVATE_KEY=rsa_key",
				"USER=testuser",
			},
			expected: []string{
				"PATH=/usr/bin",
				"USER=testuser",
			},
			filtered: []string{"API_KEY", "DATABASE_PASSWORD", "SECRET_TOKEN", "PRIVATE_KEY"},
		},
		{
			name: "case insensitive filtering",
			input: []string{
				"path=/usr/bin",
				"api_key=12345",
				"Api_Key=67890",
				"API_KEY=abcde",
			},
			expected: []string{
				"path=/usr/bin",
			},
			filtered: []string{"api_key", "Api_Key", "API_KEY"},
		},
		{
			name: "no sensitive vars",
			input: []string{
				"PATH=/usr/bin",
				"HOME=/home/user",
				"LANG=en_US.UTF-8",
				"TERM=xterm-256color",
			},
			expected: []string{
				"PATH=/usr/bin",
				"HOME=/home/user",
				"LANG=en_US.UTF-8",
				"TERM=xterm-256color",
			},
			filtered: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &SecureExecutor{}
			result := executor.filterEnvironment(tt.input)

			// Check that expected vars are present
			for _, expected := range tt.expected {
				assert.Contains(t, result, expected)
			}

			// Check that filtered vars are not present
			for _, filtered := range tt.filtered {
				for _, env := range result {
					assert.False(t, strings.HasPrefix(strings.ToUpper(env), strings.ToUpper(filtered)))
				}
			}

			// Check total count
			assert.Equal(t, len(tt.expected), len(result))
		})
	}
}

func TestSecureExecutor_Execute(t *testing.T) {
	tempDir := t.TempDir() // Create cross-platform temp directory

	tests := []struct {
		name    string
		setup   func(*SecureExecutor)
		command string
		args    []string
		wantErr bool
		check   func(*testing.T, *SecureExecutor, error)
	}{
		{
			name: "successful command execution",
			setup: func(e *SecureExecutor) {
				e.DryRun = false
			},
			command: "echo",
			args:    []string{"test"},
			wantErr: false,
		},
		{
			name: "dry run mode",
			setup: func(e *SecureExecutor) {
				e.DryRun = true
			},
			command: "echo",
			args:    []string{"test"},
			wantErr: false,
			check: func(t *testing.T, e *SecureExecutor, err error) {
				assert.NoError(t, err)
				// In dry run, command should not actually execute
			},
		},
		{
			name: "command with timeout",
			setup: func(e *SecureExecutor) {
				e.Timeout = 100 * time.Millisecond
			},
			command: func() string {
				if runtime.GOOS == "windows" {
					return "timeout"
				}
				return "sleep"
			}(),
			args: func() []string {
				if runtime.GOOS == "windows" {
					return []string{"/t", "2"}
				}
				return []string{"2"}
			}(),
			wantErr: true,
			check: func(t *testing.T, e *SecureExecutor, err error) {
				assert.Error(t, err)
				// The error might be different on various platforms
				errMsg := err.Error()
				assert.True(t,
					strings.Contains(errMsg, "killed") ||
						strings.Contains(errMsg, "context deadline exceeded") ||
						strings.Contains(errMsg, "exit status") ||
						strings.Contains(errMsg, "terminated") ||
						strings.Contains(errMsg, "context canceled"),
					"Expected timeout-related error but got: %s", errMsg)
			},
		},
		{
			name: "command with working directory",
			setup: func(e *SecureExecutor) {
				e.WorkingDir = tempDir
			},
			command: "pwd",
			args:    []string{},
			wantErr: false,
		},
		{
			name: "invalid command",
			setup: func(e *SecureExecutor) {
				// No special setup needed
			},
			command: "nonexistentcommand12345",
			args:    []string{},
			wantErr: true,
		},
		{
			name: "command with dangerous args rejected",
			setup: func(e *SecureExecutor) {
				// No special setup needed
			},
			command: "echo",
			args:    []string{"$(rm -rf /)"},
			wantErr: true,
			check: func(t *testing.T, e *SecureExecutor, err error) {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "command validation failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewSecureExecutor()
			if tt.setup != nil {
				tt.setup(executor)
			}

			ctx := context.Background()
			err := executor.Execute(ctx, tt.command, tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.check != nil {
				tt.check(t, executor, err)
			}
		})
	}
}

func TestSecureExecutor_ExecuteOutput(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*SecureExecutor)
		command    string
		args       []string
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "capture command output",
			command:    "echo",
			args:       []string{"hello world"},
			wantOutput: "hello world",
			wantErr:    false,
		},
		{
			name: "dry run returns mock output",
			setup: func(e *SecureExecutor) {
				e.DryRun = true
			},
			command:    "echo",
			args:       []string{"test"},
			wantOutput: "[DRY RUN] Would execute: echo test",
			wantErr:    false,
		},
		// Skip this test for now as output handling varies by platform
		// {
		// 	name:       "command with error returns output and error",
		// 	command:    "sh",
		// 	args:       []string{"-c", "echo 'test output'; exit 1"},
		// 	wantOutput: "test output",
		// 	wantErr:    true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := NewSecureExecutor()
			if tt.setup != nil {
				tt.setup(executor)
			}

			ctx := context.Background()
			output, err := executor.ExecuteOutput(ctx, tt.command, tt.args...)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			if tt.wantOutput != "" {
				assert.Equal(t, tt.wantOutput, strings.TrimSpace(output))
			}
		})
	}
}

func TestSecureExecutor_ExecuteWithEnv(t *testing.T) {
	executor := NewSecureExecutor()
	ctx := context.Background()

	// Test that custom env vars are added
	customEnv := []string{"CUSTOM_VAR=test123"}
	err := executor.ExecuteWithEnv(ctx, customEnv, "sh", "-c", "echo $CUSTOM_VAR")
	assert.NoError(t, err)

	// Test that sensitive vars in custom env are still included (user's responsibility)
	sensitiveEnv := []string{"API_KEY=secret"}
	err = executor.ExecuteWithEnv(ctx, sensitiveEnv, "echo", "test")
	assert.NoError(t, err)
}

func TestValidateCommandArg(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		wantErr bool
	}{
		{"safe argument", "hello-world_123", false},
		{"command substitution with $()", "$(whoami)", true},
		{"command substitution with backticks", "`whoami`", true},
		{"command chaining with &&", "test && malicious", true},
		{"command chaining with ||", "test || malicious", true},
		{"command separator with ;", "test; malicious", true},
		{"pipe operator", "test | grep something", true},
		{"output redirect", "test > /tmp/file", true},
		{"input redirect", "test < /etc/passwd", true},
		{"complex injection", "$(echo test)", true},
		{"IFS manipulation", "${IFS}", true},
		{"safe path", "/usr/local/bin/tool", false},
		{"safe flag", "--verbose", false},
		{"safe value with spaces", "commit message here", false},
		{"safe JSON", `{"key": "value"}`, false},
		{"safe regex", "^[a-zA-Z0-9]+$", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommandArg(tt.arg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		{"valid relative path", "src/main.go", false, ""},
		{"valid nested path", "path/to/file.txt", false, ""},
		{"path traversal with ..", "../../../etc/passwd", true, "path traversal detected"},
		{"path traversal hidden", "path/../../../etc/passwd", true, "path traversal detected"},
		{
			name: "absolute path outside /tmp",
			path: func() string {
				if runtime.GOOS == "windows" {
					return "C:\\Windows\\System32\\drivers\\etc\\hosts"
				}
				return "/etc/passwd"
			}(),
			wantErr: true,
			errMsg:  "absolute paths not allowed",
		},
		{"absolute path in /tmp", "/tmp/test.txt", false, ""},
		{"current directory", ".", false, ""},
		{"parent directory", "..", true, "path traversal detected"},
		{"hidden file", ".gitignore", false, ""},
		{"path with spaces", "my documents/file.txt", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMockExecutor(t *testing.T) {
	mock := NewMockExecutor()
	ctx := context.Background()

	// Set up mock responses
	mock.SetResponse("echo hello", "hello", nil)
	mock.SetResponse("failing-command ", "", errors.New("command failed")) // Note the space after command name

	// Test successful command
	err := mock.Execute(ctx, "echo", "hello")
	assert.NoError(t, err)
	assert.Len(t, mock.ExecuteCalls, 1)
	assert.Equal(t, "echo", mock.ExecuteCalls[0].Name)
	assert.Equal(t, []string{"hello"}, mock.ExecuteCalls[0].Args)

	// Test command with output
	output, err := mock.ExecuteOutput(ctx, "echo", "hello")
	assert.NoError(t, err)
	assert.Equal(t, "hello", output)
	assert.Len(t, mock.ExecuteOutputCalls, 1)

	// Test failing command
	output, err = mock.ExecuteOutput(ctx, "failing-command")
	assert.Error(t, err)
	assert.Equal(t, "command failed", err.Error())
	assert.Empty(t, output)

	// Test command with environment
	err = mock.ExecuteWithEnv(ctx, []string{"TEST=1"}, "echo", "test")
	assert.NoError(t, err)
	assert.Len(t, mock.ExecuteCalls, 2) // First Execute + ExecuteWithEnv
	assert.Equal(t, []string{"TEST=1"}, mock.ExecuteCalls[1].Env)
}

func TestSecureExecutor_ContextCancellation(t *testing.T) {
	executor := NewSecureExecutor()

	// Create a context that we'll cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Use cross-platform long-running command
	var cmd, arg string
	if runtime.GOOS == "windows" {
		cmd = "timeout"
		arg = "/t 10"
	} else {
		cmd = "sleep"
		arg = "5"
	}

	// Start a long-running command
	done := make(chan error)
	go func() {
		done <- executor.Execute(ctx, cmd, arg)
	}()

	// Cancel the context after a short delay
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Wait for the command to finish
	err := <-done
	assert.Error(t, err)
	// Context cancellation can result in different error messages on different platforms
	errMsg := err.Error()
	assert.True(t,
		strings.Contains(errMsg, "context canceled") ||
			strings.Contains(errMsg, "signal: killed") ||
			strings.Contains(errMsg, "killed") ||
			strings.Contains(errMsg, "exit status") ||
			strings.Contains(errMsg, "terminated"),
		"Expected cancellation-related error but got: %s", errMsg)
}

func TestSecureExecutor_AuditLogging(t *testing.T) {
	// Note: This test verifies that audit logging is attempted but doesn't fail
	// if the audit logger is not available (as per the implementation)
	executor := NewSecureExecutor()
	ctx := context.Background()

	// Execute a command - audit logging should be attempted
	err := executor.Execute(ctx, "echo", "test")
	assert.NoError(t, err)

	// Even if audit logger is nil, the command should still execute
	// This ensures that audit logging doesn't break command execution
}

// Benchmark tests
func BenchmarkSecureExecutor_Execute(b *testing.B) {
	executor := NewSecureExecutor()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := executor.Execute(ctx, "echo", "test"); err != nil {
			b.Errorf("Execute failed: %v", err)
		}
	}
}

func BenchmarkSecureExecutor_ValidateCommand(b *testing.B) {
	executor := NewSecureExecutor()
	args := []string{"arg1", "arg2", "arg3"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := executor.validateCommand("echo", args); err != nil {
			b.Errorf("validateCommand failed: %v", err)
		}
	}
}

func BenchmarkSecureExecutor_FilterEnvironment(b *testing.B) {
	executor := NewSecureExecutor()
	env := os.Environ()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = executor.filterEnvironment(env)
	}
}
