package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SecurityTestSuite contains all security-focused tests for the paths package
type SecurityTestSuite struct {
	tempDir string
}

// setupSecurityTests creates a temporary directory with controlled structure
func setupSecurityTests(t *testing.T) *SecurityTestSuite {
	tempDir := t.TempDir()

	// Create test structure with various security scenarios
	err := os.MkdirAll(filepath.Join(tempDir, "safe"), 0o750)
	require.NoError(t, err)

	err = os.MkdirAll(filepath.Join(tempDir, "restricted"), 0o000) // No permissions
	require.NoError(t, err)

	// Create test files
	err = os.WriteFile(filepath.Join(tempDir, "safe", "test.txt"), []byte("safe content"), 0o600)
	require.NoError(t, err)

	return &SecurityTestSuite{tempDir: tempDir}
}

// TestPathTraversalPrevention tests various path traversal attack vectors
func TestPathTraversalPrevention(t *testing.T) {
	_ = setupSecurityTests(t) // Set up test environment

	tests := []struct {
		name     string
		path     string
		wantSafe bool
		desc     string
	}{
		{
			name:     "basic_dotdot",
			path:     "../../../etc/passwd",
			wantSafe: false,
			desc:     "Classic path traversal with ../../../",
		},
		{
			name:     "url_encoded_dotdot",
			path:     "%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
			wantSafe: false,
			desc:     "URL encoded path traversal",
		},
		{
			name:     "double_encoded_dotdot",
			path:     "%252e%252e%252f%252e%252e%252f%252e%252e%252fetc%252fpasswd",
			wantSafe: false,
			desc:     "Double URL encoded path traversal",
		},
		{
			name:     "unicode_dotdot",
			path:     "\u002e\u002e\u002f\u002e\u002e\u002f\u002e\u002e\u002fetc\u002fpasswd",
			wantSafe: false,
			desc:     "Unicode encoded path traversal",
		},
		{
			name:     "mixed_slashes",
			path:     "..\\..\\..\\etc\\passwd",
			wantSafe: false,
			desc:     "Path traversal with backslashes",
		},
		{
			name:     "null_byte_injection",
			path:     "safe.txt\x00../../../etc/passwd",
			wantSafe: false,
			desc:     "Null byte injection attack",
		},
		{
			name:     "control_chars",
			path:     "safe.txt\n../../../etc/passwd",
			wantSafe: false,
			desc:     "Control character injection",
		},
		{
			name:     "windows_unc_path",
			path:     "\\\\server\\share\\file",
			wantSafe: false,
			desc:     "Windows UNC path injection",
		},
		{
			name:     "windows_drive_path",
			path:     "C:\\Windows\\System32\\config\\SAM",
			wantSafe: false,
			desc:     "Windows absolute drive path",
		},
		{
			name:     "overlong_utf8",
			path:     "/safe/\xc0\xaf\xc0\xaf\xc0\xaf",
			wantSafe: false,
			desc:     "Overlong UTF-8 encoding attack",
		},
		{
			name:     "safe_path",
			path:     "safe/test.txt",
			wantSafe: true,
			desc:     "Legitimate safe path",
		},
		{
			name:     "dotfile",
			path:     ".hidden",
			wantSafe: true,
			desc:     "Hidden files should be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPathBuilder(tt.path)
			isSafe := pb.IsSafe()

			if tt.wantSafe {
				assert.True(t, isSafe, "Path should be safe: %s (%s)", tt.path, tt.desc)
			} else {
				assert.False(t, isSafe, "Path should NOT be safe: %s (%s)", tt.path, tt.desc)
			}

			// Additional validation using the secure validator
			validator := NewPathValidator().RequireSecure()
			errors := validator.ValidatePath(pb)

			if tt.wantSafe {
				assert.Empty(t, errors, "Safe path should pass validation: %s", tt.path)
			} else {
				// For unsafe paths, we expect validation errors
				assert.NotEmpty(t, errors, "Unsafe path should fail validation: %s", tt.path)
			}
		})
	}
}

// TestSymlinkAttackPrevention tests protection against symlink-based attacks
func TestSymlinkAttackPrevention(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Symlink tests not reliable on Windows")
	}

	suite := setupSecurityTests(t)

	// Create a symlink that points outside the safe area
	unsafePath := "/etc/passwd"
	symlinkPath := filepath.Join(suite.tempDir, "safe", "evil_symlink")

	err := os.Symlink(unsafePath, symlinkPath)
	require.NoError(t, err, "Failed to create test symlink")

	// Test symlink detection
	pb := NewPathBuilder(symlinkPath)

	// The path should be detected as unsafe due to symlink
	assert.False(t, pb.IsSafe(), "Symlink pointing outside safe area should be unsafe")

	// Test Readlink functionality
	target, err := pb.Readlink()
	require.NoError(t, err, "Readlink should work")
	assert.Equal(t, unsafePath, target.String(), "Readlink should return correct target")

	// Test that following symlinks is controlled
	options := PathOptions{
		FollowSymlinks:     false,
		RestrictToBasePath: suite.tempDir,
		AllowUnsafePaths:   false,
	}

	pbRestricted := NewPathBuilderWithOptions(symlinkPath, options)
	assert.False(t, pbRestricted.IsSafe(), "Restricted builder should reject unsafe symlinks")
}

// TestIsSymlinkUnsafe_Comprehensive tests isSymlinkUnsafe for all branches
func TestIsSymlinkUnsafe_Comprehensive(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Symlink tests not reliable on Windows")
	}

	tempDir := t.TempDir()

	t.Run("path does not exist", func(t *testing.T) {
		// Non-existent path should return false (not unsafe)
		nonExistent := filepath.Join(tempDir, "does_not_exist")
		pb := NewPathBuilder(nonExistent)
		// isSymlinkUnsafe is a private method, so we test via IsSafe behavior
		assert.True(t, pb.IsSafe(), "Non-existent path should be safe from symlink perspective")
	})

	t.Run("regular file is not symlink unsafe", func(t *testing.T) {
		regularFile := filepath.Join(tempDir, "regular_file.txt")
		err := os.WriteFile(regularFile, []byte("content"), 0o600)
		require.NoError(t, err)

		pb := NewPathBuilder(regularFile)
		assert.True(t, pb.IsSafe(), "Regular file should be safe")
	})

	t.Run("symlink with FollowSymlinks disabled", func(t *testing.T) {
		// Create a regular file as target
		targetFile := filepath.Join(tempDir, "target_file.txt")
		err := os.WriteFile(targetFile, []byte("content"), 0o600)
		require.NoError(t, err)

		// Create symlink to the target
		symlinkPath := filepath.Join(tempDir, "safe_symlink")
		err = os.Symlink(targetFile, symlinkPath)
		require.NoError(t, err)

		// With FollowSymlinks disabled, any symlink is unsafe
		options := PathOptions{
			FollowSymlinks: false,
		}
		pb := NewPathBuilderWithOptions(symlinkPath, options)
		assert.False(t, pb.IsSafe(), "Symlink should be unsafe when FollowSymlinks is disabled")
	})

	t.Run("symlink with RestrictToBasePath and absolute target inside base", func(t *testing.T) {
		subDir := filepath.Join(tempDir, "base")
		err := os.MkdirAll(subDir, 0o750)
		require.NoError(t, err)

		// Target file inside base
		targetFile := filepath.Join(subDir, "target.txt")
		err = os.WriteFile(targetFile, []byte("content"), 0o600)
		require.NoError(t, err)

		// Symlink inside base pointing to target inside base (absolute path)
		symlinkPath := filepath.Join(subDir, "symlink")
		err = os.Symlink(targetFile, symlinkPath)
		require.NoError(t, err)

		options := PathOptions{
			FollowSymlinks:     true,
			RestrictToBasePath: subDir,
		}
		pb := NewPathBuilderWithOptions(symlinkPath, options)
		assert.True(t, pb.IsSafe(), "Symlink to target within base should be safe")
	})

	t.Run("symlink with RestrictToBasePath and absolute target outside base", func(t *testing.T) {
		subDir := filepath.Join(tempDir, "restricted_base")
		err := os.MkdirAll(subDir, 0o750)
		require.NoError(t, err)

		// Symlink pointing to /etc/passwd (outside base)
		symlinkPath := filepath.Join(subDir, "evil_symlink")
		err = os.Symlink("/etc/passwd", symlinkPath)
		require.NoError(t, err)

		options := PathOptions{
			FollowSymlinks:     true,
			RestrictToBasePath: subDir,
		}
		pb := NewPathBuilderWithOptions(symlinkPath, options)
		assert.False(t, pb.IsSafe(), "Symlink to target outside base should be unsafe")
	})

	t.Run("symlink with RestrictToBasePath and relative target escaping base", func(t *testing.T) {
		subDir := filepath.Join(tempDir, "restricted_base2")
		err := os.MkdirAll(subDir, 0o750)
		require.NoError(t, err)

		// Create a file outside the base
		outsideFile := filepath.Join(tempDir, "outside.txt")
		err = os.WriteFile(outsideFile, []byte("content"), 0o600)
		require.NoError(t, err)

		// Symlink with relative path escaping base
		symlinkPath := filepath.Join(subDir, "relative_escape")
		err = os.Symlink("../outside.txt", symlinkPath)
		require.NoError(t, err)

		options := PathOptions{
			FollowSymlinks:     true,
			RestrictToBasePath: subDir,
		}
		pb := NewPathBuilderWithOptions(symlinkPath, options)
		assert.False(t, pb.IsSafe(), "Symlink with relative path escaping base should be unsafe")
	})

	t.Run("symlink with RestrictToBasePath and relative target staying in base", func(t *testing.T) {
		subDir := filepath.Join(tempDir, "restricted_base3")
		nestedDir := filepath.Join(subDir, "nested")
		err := os.MkdirAll(nestedDir, 0o750)
		require.NoError(t, err)

		// Create a file in the base
		targetFile := filepath.Join(subDir, "target.txt")
		err = os.WriteFile(targetFile, []byte("content"), 0o600)
		require.NoError(t, err)

		// Symlink from nested dir with relative path to target in base
		symlinkPath := filepath.Join(nestedDir, "relative_safe")
		err = os.Symlink("../target.txt", symlinkPath)
		require.NoError(t, err)

		options := PathOptions{
			FollowSymlinks:     true,
			RestrictToBasePath: subDir,
		}
		pb := NewPathBuilderWithOptions(symlinkPath, options)
		assert.True(t, pb.IsSafe(), "Symlink with relative path staying in base should be safe")
	})
}

// TestSymlinkRaceCondition tests for TOCTOU (Time-of-Check-Time-of-Use) vulnerabilities
func TestSymlinkRaceCondition(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Symlink tests not reliable on Windows")
	}

	suite := setupSecurityTests(t)

	// Create a legitimate file first
	filePath := filepath.Join(suite.tempDir, "safe", "race_test")
	err := os.WriteFile(filePath, []byte("legitimate content"), 0o600)
	require.NoError(t, err)

	pb := NewPathBuilder(filePath)

	// Verify it's initially safe
	assert.True(t, pb.IsSafe(), "Legitimate file should be safe")
	assert.True(t, pb.Exists(), "File should exist")

	// Simulate the file being replaced with a symlink (race condition)
	err = os.Remove(filePath)
	require.NoError(t, err)

	err = os.Symlink("/etc/passwd", filePath)
	require.NoError(t, err)

	// The PathBuilder should detect the change
	assert.True(t, pb.Exists(), "Path should still exist (as symlink)")
	assert.False(t, pb.IsSafe(), "Path should now be unsafe due to symlink")
}

// TestPermissionBoundaryValidation tests file permission and access controls
func TestPermissionBoundaryValidation(t *testing.T) {
	suite := setupSecurityTests(t)

	tests := []struct {
		name        string
		path        string
		permissions os.FileMode
		wantRead    bool
		wantWrite   bool
		wantExec    bool
	}{
		{
			name:        "read_only_file",
			path:        "readonly.txt",
			permissions: 0o444,
			wantRead:    true,
			wantWrite:   false,
			wantExec:    false,
		},
		{
			name:        "write_only_file",
			path:        "writeonly.txt",
			permissions: 0o222,
			wantRead:    false,
			wantWrite:   true,
			wantExec:    false,
		},
		{
			name:        "executable_file",
			path:        "executable.sh",
			permissions: 0o755,
			wantRead:    true,
			wantWrite:   true,
			wantExec:    true,
		},
		{
			name:        "no_permissions",
			path:        "noaccess.txt",
			permissions: 0o000,
			wantRead:    false,
			wantWrite:   false,
			wantExec:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(suite.tempDir, "safe", tt.path)

			// Create the file with specific permissions
			err := os.WriteFile(filePath, []byte("test content"), tt.permissions)
			require.NoError(t, err)

			pb := NewPathBuilder(filePath)

			// Test read validation
			readValidator := NewPathValidator().RequireReadable()
			readErrors := readValidator.ValidatePath(pb)
			if tt.wantRead {
				assert.Empty(t, readErrors, "File should be readable: %s", tt.path)
			} else {
				assert.NotEmpty(t, readErrors, "File should NOT be readable: %s", tt.path)
			}

			// Test write validation
			writeValidator := NewPathValidator().RequireWritable()
			writeErrors := writeValidator.ValidatePath(pb)
			if tt.wantWrite {
				assert.Empty(t, writeErrors, "File should be writable: %s", tt.path)
			} else {
				assert.NotEmpty(t, writeErrors, "File should NOT be writable: %s", tt.path)
			}

			// Test execute validation
			execValidator := NewPathValidator().RequireExecutable()
			execErrors := execValidator.ValidatePath(pb)
			if tt.wantExec {
				assert.Empty(t, execErrors, "File should be executable: %s", tt.path)
			} else {
				assert.NotEmpty(t, execErrors, "File should NOT be executable: %s", tt.path)
			}
		})
	}
}

// TestCrossPlatformSecurityEdgeCases tests platform-specific security issues
func TestCrossPlatformSecurityEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		platform string
		wantSafe bool
		desc     string
	}{
		// Windows-specific edge cases
		{
			name:     "windows_reserved_con",
			path:     "CON",
			platform: "windows",
			wantSafe: false,
			desc:     "Windows reserved device name CON",
		},
		{
			name:     "windows_reserved_prn",
			path:     "PRN.txt",
			platform: "windows",
			wantSafe: false,
			desc:     "Windows reserved device name PRN with extension",
		},
		{
			name:     "windows_reserved_nul",
			path:     "NUL",
			platform: "windows",
			wantSafe: false,
			desc:     "Windows reserved device name NUL",
		},
		{
			name:     "windows_alternate_data_stream",
			path:     "file.txt:hidden",
			platform: "windows",
			wantSafe: false,
			desc:     "Windows alternate data stream",
		},
		// Unix-specific edge cases
		{
			name:     "unix_proc_self_fd",
			path:     "/proc/self/fd/0",
			platform: "linux",
			wantSafe: false,
			desc:     "Unix proc filesystem access",
		},
		{
			name:     "unix_dev_random",
			path:     "/dev/random",
			platform: "linux",
			wantSafe: false,
			desc:     "Unix device file access",
		},
		// Universal edge cases
		{
			name:     "long_path_attack",
			path:     strings.Repeat("A", 5000),
			platform: "any",
			wantSafe: false,
			desc:     "Extremely long path name attack",
		},
		{
			name:     "unicode_confusables",
			path:     "file⁄txt", // Unicode fraction slash
			platform: "any",
			wantSafe: false,
			desc:     "Unicode confusable characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Skip platform-specific tests
			if tt.platform != "any" && tt.platform != runtime.GOOS {
				t.Skipf("Test %s is for platform %s, current platform is %s", tt.name, tt.platform, runtime.GOOS)
			}

			pb := NewPathBuilder(tt.path)
			isSafe := pb.IsSafe()

			if tt.wantSafe {
				assert.True(t, isSafe, "Path should be safe: %s (%s)", tt.path, tt.desc)
			} else {
				assert.False(t, isSafe, "Path should NOT be safe: %s (%s)", tt.path, tt.desc)
			}

			// Test with path length validator
			if len(tt.path) > 1000 {
				lengthValidator := NewPathValidator().RequireMaxLength(1000)
				errors := lengthValidator.ValidatePath(pb)
				assert.NotEmpty(t, errors, "Long path should fail length validation")
			}
		})
	}
}

// TestPathSanitization tests the effectiveness of path sanitization
func TestPathSanitization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		desc     string
	}{
		{
			name:     "basic_cleanup",
			input:    "./path/../to/file",
			expected: "to/file",
			desc:     "Basic path cleaning should remove ./ and ../",
		},
		{
			name:     "multiple_slashes",
			input:    "path//to///file",
			expected: "path/to/file",
			desc:     "Multiple slashes should be collapsed",
		},
		{
			name:     "trailing_slashes",
			input:    "path/to/dir/",
			expected: "path/to/dir",
			desc:     "Trailing slashes should be removed",
		},
		{
			name:     "current_dir_refs",
			input:    "./path/./to/./file",
			expected: "path/to/file",
			desc:     "Current directory references should be removed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPathBuilder(tt.input)
			cleaned := pb.Clean()

			assert.Equal(t, tt.expected, cleaned.String(), "Path sanitization failed: %s", tt.desc)

			// Ensure cleaned path is safe
			assert.True(t, cleaned.IsSafe(), "Cleaned path should be safe")
		})
	}
}

// TestDirectoryEscapePrevention tests prevention of escaping from restricted directories
func TestDirectoryEscapePrevention(t *testing.T) {
	suite := setupSecurityTests(t)

	// Set up restricted path builder
	basePath := filepath.Join(suite.tempDir, "safe")
	options := PathOptions{
		RestrictToBasePath: basePath,
		AllowUnsafePaths:   false,
		MaxPathLength:      1024,
	}

	tests := []struct {
		name        string
		path        string
		shouldAllow bool
		desc        string
	}{
		{
			name:        "within_bounds",
			path:        "test.txt",
			shouldAllow: true,
			desc:        "Path within restricted directory should be allowed",
		},
		{
			name:        "escape_attempt_relative",
			path:        "../../../etc/passwd",
			shouldAllow: false,
			desc:        "Relative path escape should be blocked",
		},
		{
			name:        "escape_attempt_absolute",
			path:        "/etc/passwd",
			shouldAllow: false,
			desc:        "Absolute path outside base should be blocked",
		},
		{
			name:        "subdir_access",
			path:        "subdir/file.txt",
			shouldAllow: true,
			desc:        "Subdirectory access should be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fullPath string
			if filepath.IsAbs(tt.path) {
				// For absolute paths, use them directly
				fullPath = tt.path
			} else {
				// For relative paths, join with base path
				fullPath = filepath.Join(basePath, tt.path)
			}
			pb := NewPathBuilderWithOptions(fullPath, options)

			if tt.shouldAllow {
				assert.True(t, pb.IsSafe(), "Path should be allowed: %s (%s)", tt.path, tt.desc)
			} else {
				assert.False(t, pb.IsSafe(), "Path should be blocked: %s (%s)", tt.path, tt.desc)
			}
		})
	}
}

// BenchmarkSecurityValidation benchmarks the performance of security validation
func BenchmarkSecurityValidation(b *testing.B) {
	tests := []struct {
		name string
		path string
	}{
		{"safe_path", "safe/path/to/file.txt"},
		{"traversal_path", "../../../etc/passwd"},
		{"long_path", strings.Repeat("long/", 100) + "file.txt"},
		{"complex_path", "./path/../to/./file//name.txt"},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			pb := NewPathBuilder(tt.path)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = pb.IsSafe()
			}
		})
	}

	b.Run("validator_performance", func(b *testing.B) {
		validator := NewPathValidator()
		validator.RequireAbsolute()
		validator.RequireReadable()
		validator.RequireMaxLength(1024)

		paths := []string{
			"/safe/path/file.txt",
			"../unsafe/path",
			"/very/long/path/" + strings.Repeat("segment/", 50) + "file.txt",
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			for _, path := range paths {
				_ = validator.Validate(path)
			}
		}
	})
}

// TestErrorHandlingAndEdgeCases tests error conditions and edge cases
func TestErrorHandlingAndEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expectError bool
		desc        string
	}{
		{
			name:        "empty_path",
			path:        "",
			expectError: false, // Empty path becomes "."
			desc:        "Empty path should be handled gracefully",
		},
		{
			name:        "just_dots",
			path:        "...",
			expectError: false,
			desc:        "Path with just dots should be handled",
		},
		{
			name:        "invalid_utf8",
			path:        "file\xff\xfe.txt",
			expectError: true,
			desc:        "Invalid UTF-8 should be detected",
		},
		{
			name:        "embedded_null",
			path:        "file\x00name.txt",
			expectError: true,
			desc:        "Embedded null bytes should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPathBuilder(tt.path)

			if tt.expectError {
				assert.False(t, pb.IsSafe(), "Path should be unsafe: %s (%s)", tt.path, tt.desc)

				validator := NewPathValidator().RequireSecure()
				errors := validator.ValidatePath(pb)
				assert.NotEmpty(t, errors, "Should have validation errors for: %s", tt.desc)
			} else {
				// Path might still be unsafe for other reasons, but shouldn't error
				validator := NewPathValidator().RequireSecure()
				errors := validator.ValidatePath(pb)
				// We don't assert on errors here since some paths might be unsafe for other reasons
				_ = errors
			}
		})
	}
}
