package security

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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
			path:    "wendang/wenjian.txt", // Chinese docs/file
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

		result := executor.filterEnvironment(env, "validator-test")

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
			err := executor.Execute(ctx, "echo", string(rune('0'+id))) //nolint:gosec // G115: bounded test value (0-9), no overflow risk
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
		"international_wenzi_arabic", // Chinese/Arabic text
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
	var result error
	for i := 0; i < b.N; i++ {
		result = ValidateCommandArg(arg)
	}
	_ = result // Prevent optimization
}

func BenchmarkValidateCommandArg_Dangerous(b *testing.B) {
	arg := "$(whoami) && rm -rf /"
	b.ResetTimer()
	var result error
	for i := 0; i < b.N; i++ {
		result = ValidateCommandArg(arg)
	}
	_ = result // Prevent optimization
}

func BenchmarkValidatePath_Safe(b *testing.B) {
	path := "path/to/safe/file.txt"
	b.ResetTimer()
	var result error
	for i := 0; i < b.N; i++ {
		result = ValidatePath(path)
	}
	_ = result // Prevent optimization
}

func BenchmarkValidatePath_Dangerous(b *testing.B) {
	path := "../../../etc/passwd"
	b.ResetTimer()
	var result error
	for i := 0; i < b.N; i++ {
		result = ValidatePath(path)
	}
	_ = result // Prevent optimization
}

// SecurityValidatorTestSuite provides comprehensive security validation testing
type SecurityValidatorTestSuite struct {
	suite.Suite
}

// TestValidateVersionComprehensive tests version validation with comprehensive coverage
func (suite *SecurityValidatorTestSuite) TestValidateVersionComprehensive() {
	tests := []struct {
		name          string
		version       string
		wantErr       bool
		errorContains string
	}{
		// Valid versions
		{"simple version", "1.0.0", false, ""},
		{"version with v prefix", "v1.0.0", false, ""},
		{"pre-release", "1.0.0-alpha", false, ""},
		{"pre-release with number", "1.0.0-beta.1", false, ""},
		{"build metadata", "1.0.0+build.123", false, ""},
		{"complex version", "1.2.3-rc.1+build.456", false, ""},
		{"high version numbers", "999.999.999", false, ""},

		// Path traversal attacks
		{"path traversal basic", "1.0.0/../etc/passwd", true, "path traversal"},
		{"path traversal windows", "1.0.0\\..\\windows\\system32", true, "path traversal"},
		{"double dots", "1..0..0", true, "path traversal"},
		{"relative path", "./1.0.0", true, "path traversal"},
		{"parent directory", "../1.0.0", true, "path traversal"},

		// Command injection attempts
		{"command substitution", "1.0.0$(whoami)", true, "command injection"},
		{"command substitution nested", "1.0.0$(echo $(id))", true, "command injection"},
		{"backtick execution", "1.0.0`whoami`", true, "command injection"},
		{"backtick nested", "1.0.0`echo `id``", true, "command injection"},
		{"shell variable", "1.0.0${PATH}", true, "command injection"},
		{"environment variable expansion", "1.0.0${IFS}rm${IFS}-rf", true, "command injection"},

		// Control character injection
		{"null byte", "1.0.0\000", true, "null byte"},
		{"newline injection", "1.0.0\nrm -rf /", true, "control character"},
		{"carriage return", "1.0.0\r\nmalicious", true, "control character"},
		{"tab character", "1.0.0\tmalicious", true, "control character"},
		{"escape sequences", "1.0.0\033[31mred\033[0m", true, "control character"},

		// Double v prefix
		{"double v prefix", "vv1.0.0", true, "double 'v' prefix"},
		{"v in version", "v1.v0.0", true, "invalid version format"},

		// Invalid UTF-8
		{"invalid utf8", string([]byte{0xff, 0xfe, 0xfd}), true, "invalid UTF-8"},

		// Malformed versions
		{"empty version", "", true, "invalid version format"},
		{"only v", "v", true, "invalid version format"},
		{"incomplete version", "1.0", true, "invalid version format"},
		{"non-numeric", "a.b.c", true, "invalid version format"},
		{"negative version", "-1.0.0", true, "invalid version format"},
		{"leading zeros", "01.02.03", false, ""},
		{"spaces", "1 . 0 . 0", true, "invalid version format"},

		// Edge cases for length
		{"very long version", strings.Repeat("1", 1000) + ".0.0", false, ""},
		{"unicode characters", "1.0.0-α", true, "invalid version format"},
		{"emoji", "1.0.0-🚀", true, "invalid version format"},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := ValidateVersion(tt.version)
			if tt.wantErr {
				suite.Require().Error(err, "Expected error for version: %s", tt.version)
				if tt.errorContains != "" {
					suite.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errorContains))
				}
			} else {
				suite.NoError(err, "Unexpected error for valid version: %s", tt.version)
			}
		})
	}
}

// TestValidateFilenameComprehensive tests filename validation thoroughly
func (suite *SecurityValidatorTestSuite) TestValidateFilenameComprehensive() {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		// Valid filenames
		{"simple file", "file.txt", false},
		{"with dash", "my-file.txt", false},
		{"with underscore", "my_file.txt", false},
		{"with dots", "file.tar.gz", false},
		{"numeric", "123.txt", false},
		{"mixed case", "MyFile.TXT", false},

		// Invalid filenames
		{"empty filename", "", true},
		{"directory reference", "..", true},
		{"current directory", ".", true},
		{"null byte", "file\000.txt", true},
		{"leading space", " file.txt", true},
		{"trailing space", "file.txt ", true},
		{"path separator unix", "dir/file.txt", true},
		{"path separator windows", "dir\\file.txt", true},
		{"with spaces", "my file.txt", true},
		{"with special chars", "file@home.txt", true},
		{"with pipe", "file|cmd.txt", true},
		{"with redirect", "file>output.txt", true},
		{"with semicolon", "file;cmd.txt", true},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := ValidateFilename(tt.filename)
			if tt.wantErr {
				suite.Require().Error(err)
			} else {
				suite.NoError(err)
			}
		})
	}
}

// TestValidateURLComprehensive tests URL validation with security focus
func (suite *SecurityValidatorTestSuite) TestValidateURLComprehensive() {
	tests := []struct {
		name          string
		url           string
		wantErr       bool
		errorContains string
	}{
		// Valid URLs
		{"http URL", "http://example.com", false, ""},
		{"https URL", "https://example.com", false, ""},
		{"with path", "https://example.com/path", false, ""},
		{"with query", "https://example.com?query=value", false, ""},
		{"with port", "https://example.com:8080", false, ""},
		{"with auth", "https://user:pass@example.com", false, ""},

		// Invalid protocols - security risk
		{"javascript protocol", "javascript:alert('xss')", true, "protocol"},
		{"data protocol", "data:text/html,<script>alert('xss')</script>", true, "suspicious"},
		{"file protocol", "file:///etc/passwd", true, "protocol"},
		{"ftp protocol", "ftp://example.com", true, "protocol"},
		{"vbscript protocol", "vbscript:msgbox('xss')", true, "suspicious"},

		// XSS attempts
		{"script tag", "https://example.com/<script>alert('xss')</script>", true, "suspicious"},
		{"script tag encoded", "https://example.com/%3Cscript%3E", true, "suspicious"},
		{"onerror attribute", "https://example.com/path?onerror=alert('xss')", true, "suspicious"},
		{"onload attribute", "https://example.com/path?onload=alert('xss')", true, "suspicious"},

		// Control characters
		{"null byte", "https://example.com\000", true, "null byte"},
		{"newline", "https://example.com\nmalicious", true, "control character"},
		{"carriage return", "https://example.com\r\nmalicious", true, "control character"},

		// Whitespace issues
		{"leading space", " https://example.com", true, "whitespace"},
		{"trailing space", "https://example.com ", true, "whitespace"},

		// Empty and malformed
		{"empty URL", "", true, "empty"},
		{"just protocol", "https://", false, ""},
		{"missing protocol", "example.com", true, "protocol"},
		{"relative URL", "//example.com", true, "protocol"},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := ValidateURL(tt.url)
			if tt.wantErr {
				suite.Require().Error(err)
				if tt.errorContains != "" {
					suite.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.errorContains))
				}
			} else {
				suite.NoError(err)
			}
		})
	}
}

// TestConcurrentValidation tests thread safety of validation functions
func (suite *SecurityValidatorTestSuite) TestConcurrentValidation() {
	const numGoroutines = 100
	const iterations = 10

	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines*iterations*4)

	testCases := []struct {
		name      string
		validator func() error
	}{
		{"ValidateVersion", func() error { return ValidateVersion("v1.0.0") }},
		{"ValidateGitRef", func() error { return ValidateGitRef("main") }},
		{"ValidateFilename", func() error { return ValidateFilename("test.txt") }},
		{"ValidateURL", func() error { return ValidateURL("https://example.com") }},
	}

	for _, tc := range testCases {
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(validator func() error) {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					if err := validator(); err != nil {
						errorChan <- err
					}
				}
			}(tc.validator)
		}
	}

	wg.Wait()
	close(errorChan)

	// Collect any errors
	errors := make([]error, 0, numGoroutines*iterations*4)
	for err := range errorChan {
		errors = append(errors, err)
	}

	suite.Empty(errors, "Validation functions should not produce errors under concurrent access")
}

// TestValidationPerformance tests performance characteristics
func (suite *SecurityValidatorTestSuite) TestValidationPerformance() {
	// Test that validation functions complete within reasonable time
	testCases := []struct {
		name      string
		validator func() error
	}{
		{"ValidateVersion long", func() error { return ValidateVersion(strings.Repeat("1", 10000) + ".0.0") }},
		{"ValidateGitRef long", func() error { return ValidateGitRef(strings.Repeat("a", 1000)) }},
		{"ValidateFilename long", func() error { return ValidateFilename(strings.Repeat("a", 255) + ".txt") }},
		{"ValidateURL long", func() error { return ValidateURL("https://" + strings.Repeat("a", 1000) + ".com") }},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Function should complete (may error, but shouldn't hang)
			done := make(chan bool, 1)
			go func() {
				if err := tc.validator(); err != nil {
					// Expected for timeout validation tests - these should error
					suite.T().Logf("Validator returned error as expected: %v", err)
				}
				done <- true
			}()

			select {
			case <-done:
				// Success - function completed
			case <-time.After(5 * time.Second):
				suite.Fail("Validation function took too long to complete")
			}
		})
	}
}

// TestRunSecurityValidatorTestSuite runs the comprehensive security test suite
func TestRunSecurityValidatorTestSuite(t *testing.T) {
	suite.Run(t, new(SecurityValidatorTestSuite))
}

// TestValidationEdgeCases tests additional edge cases
func TestValidationEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("ValidateEmail comprehensive", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name    string
			email   string
			wantErr bool
		}{
			{"valid simple", "user@example.com", false},
			{"valid with subdomain", "user@mail.example.com", false},
			{"valid with plus", "user+tag@example.com", false},
			{"valid with numbers", "user123@example123.com", false},
			{"empty email", "", true},
			{"no at symbol", "userexample.com", true},
			{"multiple at symbols", "user@@example.com", true},
			{"no domain", "user@", true},
			{"no user", "@example.com", true},
			{"no dot in domain", "user@example", true},
			{"domain ends with dot", "user@example.", true},
			{"domain starts with dot", "user@.example.com", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateEmail(tt.email)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("ValidatePort comprehensive", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name    string
			port    int
			wantErr bool
		}{
			{"valid port 80", 80, false},
			{"valid port 443", 443, false},
			{"valid port 8080", 8080, false},
			{"valid port 65535", 65535, false},
			{"valid port 1", 1, false},
			{"invalid port 0", 0, true},
			{"invalid port -1", -1, true},
			{"invalid port 65536", 65536, true},
			{"invalid port large", 999999, true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidatePort(tt.port)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("ValidateEnvVar comprehensive", func(t *testing.T) {
		t.Parallel()
		tests := []struct {
			name    string
			envVar  string
			wantErr bool
		}{
			{"valid PATH", "PATH", false},
			{"valid HOME", "HOME", false},
			{"valid with underscore", "MY_VAR", false},
			{"valid with number", "VAR123", false},
			{"valid starts with underscore", "_VAR", false},
			{"empty name", "", true},
			{"starts with number", "123VAR", true},
			{"contains dash", "MY-VAR", true},
			{"contains space", "MY VAR", true},
			{"contains special char", "MY@VAR", true},
			{"starts with dot", ".VAR", true},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := ValidateEnvVar(tt.envVar)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})
}

// Benchmark tests for performance validation
func BenchmarkSecurityValidation(b *testing.B) {
	b.Run("ValidateVersionSafe", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if err := ValidateVersion("v1.2.3-alpha.1+build.123"); err != nil {
					// This is a benchmark - errors are not expected for valid versions
					b.Errorf("Unexpected error validating version: %v", err)
				}
			}
		})
	})

	b.Run("ValidateVersionDangerous", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if err := ValidateVersion("1.0.0$(whoami)"); err == nil {
					b.Error("Expected validation error for dangerous version string")
				}
			}
		})
	})

	b.Run("ValidateFilenameSafe", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if err := ValidateFilename("document.tar.gz"); err != nil {
					// Expected - this is a benchmark, just consume the error
					_ = err
				}
			}
		})
	})

	b.Run("ValidateURLSafe", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if err := ValidateURL("https://api.example.com/v1/endpoint?param=value"); err != nil {
					// Expected - this is a benchmark, just consume the error
					_ = err
				}
			}
		})
	})
}
