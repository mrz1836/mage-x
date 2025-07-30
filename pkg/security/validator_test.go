package security

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateCommandArg_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		wantErr bool
		errMsg  string
	}{
		// Edge cases for command injection
		{
			name:    "nested command substitution",
			arg:     "$(echo $(whoami))",
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
		{
			name:    "escaped characters",
			arg:     "test\\$(whoami)",
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
		{
			name:    "unicode characters",
			arg:     "test™€£",
			wantErr: false,
		},
		{
			name:    "very long argument",
			arg:     string(make([]byte, 10000)),
			wantErr: false,
		},
		{
			name:    "empty string",
			arg:     "",
			wantErr: false,
		},
		{
			name:    "only spaces",
			arg:     "   ",
			wantErr: false,
		},
		{
			name:    "newline injection",
			arg:     "test\nmalicious command",
			wantErr: false, // Newlines alone aren't dangerous
		},
		{
			name:    "null byte injection",
			arg:     "test\x00malicious",
			wantErr: false, // Null bytes handled by OS
		},
		// SQL-like injection attempts
		{
			name:    "SQL comment style",
			arg:     "test--comment",
			wantErr: false, // Not dangerous for shell
		},
		{
			name:    "quotes within string",
			arg:     `test "quoted" value`,
			wantErr: false,
		},
		{
			name:    "single quotes",
			arg:     `test 'quoted' value`,
			wantErr: false,
		},
		// Environment variable references (safe)
		{
			name:    "environment variable reference",
			arg:     "$HOME",
			wantErr: false, // Simple env var reference is safe
		},
		{
			name:    "environment variable in braces",
			arg:     "${HOME}",
			wantErr: false, // Unless it's ${IFS}
		},
		// Special shell characters that are dangerous
		{
			name:    "background execution",
			arg:     "command &",
			wantErr: false, // & at end is less dangerous than &&
		},
		{
			name:    "glob patterns",
			arg:     "*.txt",
			wantErr: false, // Globs are generally safe
		},
		{
			name:    "brace expansion attempt",
			arg:     "{1..10}",
			wantErr: false, // Brace expansion is safe
		},
		// Real-world examples
		{
			name:    "git commit message",
			arg:     "feat: add support for $(variables)",
			wantErr: true, // Still contains $()
		},
		{
			name:    "file path with spaces",
			arg:     "/path/to/my documents/file.txt",
			wantErr: false,
		},
		{
			name:    "URL with parameters",
			arg:     "https://example.com?param=value&other=123",
			wantErr: false, // & in URL context is safe
		},
		{
			name:    "JSON string",
			arg:     `{"key": "value", "nested": {"item": 123}}`,
			wantErr: false,
		},
		{
			name:    "regular expression",
			arg:     `^[a-zA-Z0-9]+\.(jpg|png|gif)$`,
			wantErr: false,
		},
		// Combined attacks
		{
			name:    "multiple injection attempts",
			arg:     "$(whoami) && `id` || echo test; cat /etc/passwd | grep root > /tmp/out",
			wantErr: true,
			errMsg:  "dangerous pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCommandArg(tt.arg)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePath_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
		errMsg  string
	}{
		// Unicode and special characters
		{
			name:    "unicode in path",
			path:    "文档/文件.txt",
			wantErr: false,
		},
		{
			name:    "spaces and special chars",
			path:    "My Documents/Project (2023)/file [1].txt",
			wantErr: false,
		},
		// Sneaky path traversal attempts
		{
			name:    "encoded path traversal",
			path:    "path%2F..%2F..%2Fetc%2Fpasswd",
			wantErr: false, // Not decoded here
		},
		{
			name:    "double dots in filename",
			path:    "file..name.txt",
			wantErr: false, // Only directory traversal is dangerous
		},
		{
			name:    "Windows-style path",
			path:    "C:\\Users\\Documents\\file.txt",
			wantErr: true, // Absolute path
		},
		{
			name:    "Windows UNC path",
			path:    "\\\\server\\share\\file.txt",
			wantErr: true, // Absolute path
		},
		// Symbolic links (path validation doesn't resolve them)
		{
			name:    "potential symlink",
			path:    "link_to_etc",
			wantErr: false, // Can't detect symlinks without filesystem access
		},
		// Empty and edge cases
		{
			name:    "empty path",
			path:    "",
			wantErr: false, // Empty path is current directory
		},
		{
			name:    "single dot",
			path:    ".",
			wantErr: false,
		},
		{
			name:    "multiple slashes",
			path:    "path//to///file",
			wantErr: false, // Gets normalized
		},
		{
			name:    "trailing slash",
			path:    "path/to/dir/",
			wantErr: false,
		},
		// Complex traversal attempts
		{
			name:    "complex traversal",
			path:    "path/./../../etc/passwd",
			wantErr: true,
			errMsg:  "path traversal",
		},
		{
			name:    "hidden traversal",
			path:    "path/subdir/../../../etc/passwd",
			wantErr: true,
			errMsg:  "path traversal",
		},
		// Allowed absolute paths
		{
			name:    "tmp file",
			path:    "/tmp/build-12345/output.bin",
			wantErr: false,
		},
		{
			name:    "tmp with traversal",
			path:    "/tmp/../etc/passwd",
			wantErr: true, // Still has traversal
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePath(tt.path)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test for potential security bypasses
func TestSecurityBypasses(t *testing.T) {
	t.Run("command allowlist bypass attempts", func(t *testing.T) {
		executor := &SecureExecutor{
			AllowedCommands: map[string]bool{
				"echo": true,
				"ls":   true,
			},
		}

		tests := []struct {
			name    string
			command string
			wantErr bool
		}{
			{"allowed command", "echo", false},
			{"disallowed command", "rm", true},
			{"path to allowed command", "/bin/echo", true}, // Should fail, not in allowlist
			{"relative path", "./echo", true},              // Should fail
			{"command with path traversal", "../echo", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := executor.validateCommand(tt.command, []string{})
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("environment filtering bypass attempts", func(t *testing.T) {
		executor := &SecureExecutor{}

		// Try various ways to hide sensitive vars
		env := []string{
			"API_KEY=visible",        // Should be filtered
			"api_key=lowercase",      // Should be filtered
			"ApI_kEy=mixedcase",      // Should be filtered
			"XAPI_KEY=prefixed",      // Should pass
			"API_KEYX=suffixed",      // Should pass
			"API KEY=with space",     // Malformed, but should pass
			"=API_KEY",               // Malformed
			"API_KEY",                // Malformed (no =)
			"PRIVATE_KEY_FILE=/path", // Should be filtered (PRIVATE_KEY prefix)
			"MY_PRIVATE_KEY=value",   // Should pass (PRIVATE_KEY not at start)
		}

		result := executor.filterEnvironment(env)

		// Check filtered
		for _, r := range result {
			assert.NotContains(t, r, "API_KEY=visible")
			assert.NotContains(t, r, "api_key=lowercase")
			assert.NotContains(t, r, "ApI_kEy=mixedcase")
			assert.NotContains(t, r, "PRIVATE_KEY_FILE=/path")
		}

		// Check passed
		assert.Contains(t, result, "XAPI_KEY=prefixed")
		assert.Contains(t, result, "API_KEYX=suffixed")
		assert.Contains(t, result, "MY_PRIVATE_KEY=value")
	})
}

// Concurrency tests
func TestSecureExecutor_Concurrent(t *testing.T) {
	executor := NewSecureExecutor()
	ctx := context.Background()

	// Run multiple commands concurrently
	done := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			err := executor.Execute(ctx, "echo", string(rune('0'+id)))
			done <- err
		}(i)
	}

	// Collect results
	for i := 0; i < 10; i++ {
		err := <-done
		assert.NoError(t, err)
	}
}

// Fuzzing-style test with random inputs
func TestValidateCommandArg_RandomInputs(t *testing.T) {
	// Test with various random but safe inputs
	safeInputs := []string{
		"simple",
		"with-dash",
		"with_underscore",
		"with.dot",
		"with:colon",
		"with=equals",
		"with spaces in middle",
		"123numeric",
		"UPPERCASE",
		"CamelCase",
		"/absolute/path/to/file",
		"relative/path/to/file",
		"file.extension.tar.gz",
		"[bracketed]",
		"(parentheses)",
		"{braces}",
		"@special#chars%allowed",
		"international_文字_العربية",
	}

	for _, input := range safeInputs {
		t.Run("safe_input", func(t *testing.T) {
			err := ValidateCommandArg(input)
			assert.NoError(t, err, "Input %q should be safe", input)
		})
	}
}

// Performance test for validation
func BenchmarkValidateCommandArg_Safe(b *testing.B) {
	arg := "this is a safe argument with no dangerous patterns"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ValidateCommandArg(arg); err != nil {
			// Expected for dangerous inputs, ignore for benchmarking
		}
	}
}

func BenchmarkValidateCommandArg_Dangerous(b *testing.B) {
	arg := "$(whoami) && rm -rf /"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ValidateCommandArg(arg); err != nil {
			// Expected for dangerous inputs, ignore for benchmarking
		}
	}
}

func BenchmarkValidatePath_Safe(b *testing.B) {
	path := "path/to/safe/file.txt"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ValidatePath(path); err != nil {
			// Expected for dangerous paths, ignore for benchmarking
		}
	}
}

func BenchmarkValidatePath_Dangerous(b *testing.B) {
	path := "../../../etc/passwd"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := ValidatePath(path); err != nil {
			// Expected for dangerous paths, ignore for benchmarking
		}
	}
}
