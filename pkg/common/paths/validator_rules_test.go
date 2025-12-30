package paths

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Rule Management Tests
// =============================================================================

// TestValidator_AddRule tests adding validation rules to the validator.
// Validates that rules can be added and nil rules are rejected.
func TestValidator_AddRule(t *testing.T) {
	t.Run("add valid rule", func(t *testing.T) {
		v := NewPathValidator()
		err := v.AddRule(&AbsolutePathRule{})
		require.NoError(t, err)

		rules := v.Rules()
		assert.Len(t, rules, 1)
		assert.Equal(t, "absolute-path", rules[0].Name())
	})

	t.Run("add nil rule returns error", func(t *testing.T) {
		v := NewPathValidator()
		err := v.AddRule(nil)
		require.ErrorIs(t, err, ErrRuleCannotBeNil)
	})

	t.Run("add multiple rules", func(t *testing.T) {
		v := NewPathValidator()
		require.NoError(t, v.AddRule(&AbsolutePathRule{}))
		require.NoError(t, v.AddRule(&ExistsRule{}))
		require.NoError(t, v.AddRule(&ReadableRule{}))

		rules := v.Rules()
		assert.Len(t, rules, 3)
	})
}

// TestValidator_RemoveRule tests removing validation rules from the validator.
// Rules are identified by name for removal.
func TestValidator_RemoveRule(t *testing.T) {
	t.Run("remove existing rule", func(t *testing.T) {
		v := NewPathValidator()
		require.NoError(t, v.AddRule(&AbsolutePathRule{}))
		require.NoError(t, v.AddRule(&ExistsRule{}))

		err := v.RemoveRule("absolute-path")
		require.NoError(t, err)

		rules := v.Rules()
		assert.Len(t, rules, 1)
		assert.Equal(t, "exists", rules[0].Name())
	})

	t.Run("remove non-existent rule returns error", func(t *testing.T) {
		v := NewPathValidator()
		require.NoError(t, v.AddRule(&AbsolutePathRule{}))

		err := v.RemoveRule("non-existent")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrRuleNotFound)
	})

	t.Run("remove from empty validator returns error", func(t *testing.T) {
		v := NewPathValidator()
		err := v.RemoveRule("any-rule")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrRuleNotFound)
	})
}

// TestValidator_ClearRules tests clearing all validation rules.
func TestValidator_ClearRules(t *testing.T) {
	v := NewPathValidator()
	require.NoError(t, v.AddRule(&AbsolutePathRule{}))
	require.NoError(t, v.AddRule(&ExistsRule{}))
	require.NoError(t, v.AddRule(&ReadableRule{}))

	err := v.ClearRules()
	require.NoError(t, err)

	rules := v.Rules()
	assert.Empty(t, rules)
}

// TestValidator_Validate tests the Validate method with multiple rules.
// Validates that all rules are checked and errors are accumulated.
func TestValidator_Validate(t *testing.T) {
	t.Run("path passes all rules", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")
		require.NoError(t, os.WriteFile(testFile, []byte("test"), 0o600))

		v := NewPathValidator()
		v.RequireExists()
		v.RequireFile()

		errors := v.Validate(testFile)
		assert.Empty(t, errors)
	})

	t.Run("path fails single rule", func(t *testing.T) {
		v := NewPathValidator()
		v.RequireAbsolute()

		errors := v.Validate("relative/path")
		assert.Len(t, errors, 1)
		assert.Equal(t, "absolute-path", errors[0].Rule)
	})

	t.Run("path fails multiple rules", func(t *testing.T) {
		v := NewPathValidator()
		v.RequireAbsolute()
		v.RequireExists()

		errors := v.Validate("relative/nonexistent")
		assert.Len(t, errors, 2)
	})

	t.Run("IsValid returns true when all rules pass", func(t *testing.T) {
		v := NewPathValidator()
		v.RequireRelative()

		assert.True(t, v.IsValid("relative/path"))
	})

	t.Run("IsValid returns false when any rule fails", func(t *testing.T) {
		v := NewPathValidator()
		v.RequireAbsolute()

		assert.False(t, v.IsValid("relative/path"))
	})
}

// =============================================================================
// AbsolutePathRule Tests
// =============================================================================

// TestAbsolutePathRule tests that paths must be absolute.
// Absolute paths start with / on Unix or a drive letter on Windows.
func TestAbsolutePathRule(t *testing.T) {
	rule := &AbsolutePathRule{}

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		{"absolute unix path", "/home/user/file.txt", false},
		{"absolute with multiple dirs", "/a/b/c/d/e", false},
		{"root path", "/", false},
		{"relative path", "relative/path", true},
		{"current directory", ".", true},
		{"parent directory", "..", true},
		{"relative with dot", "./file.txt", true},
		{"relative with parent", "../file.txt", true},
		{"empty path", "", true},
		{"just filename", "file.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.path)
			if tt.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrPathMustBeAbsolute)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("rule name and description", func(t *testing.T) {
		assert.Equal(t, "absolute-path", rule.Name())
		assert.NotEmpty(t, rule.Description())
	})

	t.Run("ValidatePath with PathBuilder", func(t *testing.T) {
		pb := NewPathBuilder("/absolute/path")
		err := rule.ValidatePath(pb)
		require.NoError(t, err)

		pb2 := NewPathBuilder("relative/path")
		err2 := rule.ValidatePath(pb2)
		require.Error(t, err2)
	})
}

// =============================================================================
// RelativePathRule Tests
// =============================================================================

// TestRelativePathRule tests that paths must be relative.
// Relative paths do not start with / or a drive letter.
func TestRelativePathRule(t *testing.T) {
	rule := &RelativePathRule{}

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		{"relative path", "relative/path", false},
		{"current directory", ".", false},
		{"parent directory", "..", false},
		{"relative with dot", "./file.txt", false},
		{"relative with parent", "../file.txt", false},
		{"just filename", "file.txt", false},
		{"empty path", "", false},
		{"absolute unix path", "/home/user/file.txt", true},
		{"root path", "/", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.path)
			if tt.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrPathMustBeRelative)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("rule name and description", func(t *testing.T) {
		assert.Equal(t, "relative-path", rule.Name())
		assert.NotEmpty(t, rule.Description())
	})
}

// =============================================================================
// ExistsRule Tests
// =============================================================================

// TestExistsRule tests that paths must exist on the filesystem.
func TestExistsRule(t *testing.T) {
	rule := &ExistsRule{}
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "exists.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0o600))

	// Create test directory
	testSubDir := filepath.Join(tmpDir, "subdir")
	require.NoError(t, os.Mkdir(testSubDir, 0o700))

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		{"existing file", testFile, false},
		{"existing directory", testSubDir, false},
		{"existing temp dir", tmpDir, false},
		{"non-existent file", filepath.Join(tmpDir, "nonexistent.txt"), true},
		{"non-existent directory", filepath.Join(tmpDir, "nonexistent_dir"), true},
		{"empty path", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.path)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("rule name and description", func(t *testing.T) {
		assert.Equal(t, "exists", rule.Name())
		assert.NotEmpty(t, rule.Description())
	})

	t.Run("ValidatePath with PathBuilder", func(t *testing.T) {
		pb := NewPathBuilder(testFile)
		err := rule.ValidatePath(pb)
		require.NoError(t, err)
	})
}

// =============================================================================
// NotExistsRule Tests
// =============================================================================

// TestNotExistsRule tests that paths must not exist on the filesystem.
func TestNotExistsRule(t *testing.T) {
	rule := &NotExistsRule{}
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "exists.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0o600))

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		{"non-existent file", filepath.Join(tmpDir, "nonexistent.txt"), false},
		{"non-existent directory", filepath.Join(tmpDir, "nonexistent_dir"), false},
		{"existing file", testFile, true},
		{"existing directory", tmpDir, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.path)
			if tt.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrPathAlreadyExists)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("rule name and description", func(t *testing.T) {
		assert.Equal(t, "not-exists", rule.Name())
		assert.NotEmpty(t, rule.Description())
	})
}

// =============================================================================
// ReadableRule Tests
// =============================================================================

// TestReadableRule tests that paths must be readable.
func TestReadableRule(t *testing.T) {
	rule := &ReadableRule{}
	tmpDir := t.TempDir()

	// Create readable file
	readableFile := filepath.Join(tmpDir, "readable.txt")
	require.NoError(t, os.WriteFile(readableFile, []byte("test"), 0o600))

	t.Run("readable file passes", func(t *testing.T) {
		err := rule.Validate(readableFile)
		require.NoError(t, err)
	})

	t.Run("readable directory passes", func(t *testing.T) {
		err := rule.Validate(tmpDir)
		require.NoError(t, err)
	})

	t.Run("non-existent file fails", func(t *testing.T) {
		err := rule.Validate(filepath.Join(tmpDir, "nonexistent.txt"))
		require.Error(t, err)
	})

	t.Run("path traversal detected with unresolved dots", func(t *testing.T) {
		// The implementation checks for ".." after filepath.Clean
		// Use a path that keeps ".." after cleaning (relative path)
		err := rule.Validate("../../../etc/passwd")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrPathTraversalDetected)
	})

	t.Run("rule name and description", func(t *testing.T) {
		assert.Equal(t, "readable", rule.Name())
		assert.NotEmpty(t, rule.Description())
	})
}

// =============================================================================
// WritableRule Tests
// =============================================================================

// TestWritableRule tests that paths must be writable.
func TestWritableRule(t *testing.T) {
	rule := &WritableRule{}
	tmpDir := t.TempDir()

	// Create writable file
	writableFile := filepath.Join(tmpDir, "writable.txt")
	require.NoError(t, os.WriteFile(writableFile, []byte("test"), 0o600))

	t.Run("writable file passes", func(t *testing.T) {
		err := rule.Validate(writableFile)
		require.NoError(t, err)
	})

	t.Run("writable directory passes", func(t *testing.T) {
		err := rule.Validate(tmpDir)
		require.NoError(t, err)
	})

	t.Run("non-existent file in writable directory passes", func(t *testing.T) {
		// Parent directory is writable, so creating new file should be possible
		err := rule.Validate(filepath.Join(tmpDir, "newfile.txt"))
		require.NoError(t, err)
	})

	t.Run("path traversal detected with unresolved dots", func(t *testing.T) {
		// The implementation checks for ".." after filepath.Clean
		// Use a path that keeps ".." after cleaning (relative path)
		err := rule.Validate("../../../etc/passwd")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrPathTraversalDetected)
	})

	t.Run("rule name and description", func(t *testing.T) {
		assert.Equal(t, "writable", rule.Name())
		assert.NotEmpty(t, rule.Description())
	})
}

// =============================================================================
// ExecutableRule Tests
// =============================================================================

// TestExecutableRule tests that paths must be executable.
func TestExecutableRule(t *testing.T) {
	rule := &ExecutableRule{}
	tmpDir := t.TempDir()

	// Create executable file
	execFile := filepath.Join(tmpDir, "executable")
	//nolint:gosec // G306: intentionally creating executable file for testing
	require.NoError(t, os.WriteFile(execFile, []byte("#!/bin/sh\necho test"), 0o700))

	// Create non-executable file
	nonExecFile := filepath.Join(tmpDir, "non_executable.txt")
	require.NoError(t, os.WriteFile(nonExecFile, []byte("test"), 0o600))

	t.Run("executable file passes", func(t *testing.T) {
		err := rule.Validate(execFile)
		require.NoError(t, err)
	})

	t.Run("non-executable file fails", func(t *testing.T) {
		err := rule.Validate(nonExecFile)
		require.Error(t, err)
	})

	t.Run("directory is executable", func(t *testing.T) {
		// Directories typically have execute permission (for traversal)
		err := rule.Validate(tmpDir)
		require.NoError(t, err)
	})

	t.Run("non-existent file fails", func(t *testing.T) {
		err := rule.Validate(filepath.Join(tmpDir, "nonexistent"))
		require.Error(t, err)
	})

	t.Run("rule name and description", func(t *testing.T) {
		assert.Equal(t, "executable", rule.Name())
		assert.NotEmpty(t, rule.Description())
	})
}

// =============================================================================
// DirectoryRule Tests
// =============================================================================

// TestDirectoryRule tests that paths must be directories.
func TestDirectoryRule(t *testing.T) {
	rule := &DirectoryRule{}
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "file.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0o600))

	// Create test subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	require.NoError(t, os.Mkdir(subDir, 0o700))

	t.Run("directory passes", func(t *testing.T) {
		err := rule.Validate(tmpDir)
		require.NoError(t, err)
	})

	t.Run("subdirectory passes", func(t *testing.T) {
		err := rule.Validate(subDir)
		require.NoError(t, err)
	})

	t.Run("file fails", func(t *testing.T) {
		err := rule.Validate(testFile)
		require.Error(t, err)
	})

	t.Run("non-existent path fails", func(t *testing.T) {
		err := rule.Validate(filepath.Join(tmpDir, "nonexistent"))
		require.Error(t, err)
	})

	t.Run("rule name and description", func(t *testing.T) {
		assert.Equal(t, "directory", rule.Name())
		assert.NotEmpty(t, rule.Description())
	})

	t.Run("ValidatePath with PathBuilder", func(t *testing.T) {
		pb := NewPathBuilder(tmpDir)
		err := rule.ValidatePath(pb)
		require.NoError(t, err)

		pb2 := NewPathBuilder(testFile)
		err2 := rule.ValidatePath(pb2)
		require.Error(t, err2)
	})
}

// =============================================================================
// FileRule Tests
// =============================================================================

// TestFileRule tests that paths must be regular files.
func TestFileRule(t *testing.T) {
	rule := &FileRule{}
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "file.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0o600))

	t.Run("file passes", func(t *testing.T) {
		err := rule.Validate(testFile)
		require.NoError(t, err)
	})

	t.Run("directory fails", func(t *testing.T) {
		err := rule.Validate(tmpDir)
		require.Error(t, err)
	})

	t.Run("non-existent path fails", func(t *testing.T) {
		err := rule.Validate(filepath.Join(tmpDir, "nonexistent.txt"))
		require.Error(t, err)
	})

	t.Run("rule name and description", func(t *testing.T) {
		assert.Equal(t, "file", rule.Name())
		assert.NotEmpty(t, rule.Description())
	})

	t.Run("ValidatePath with PathBuilder", func(t *testing.T) {
		pb := NewPathBuilder(testFile)
		err := rule.ValidatePath(pb)
		require.NoError(t, err)

		pb2 := NewPathBuilder(tmpDir)
		err2 := rule.ValidatePath(pb2)
		require.Error(t, err2)
	})
}

// =============================================================================
// ExtensionRule Tests
// =============================================================================

// TestExtensionRule tests that paths must have one of the allowed extensions.
// Extensions are case-insensitive and can be specified with or without leading dot.
func TestExtensionRule(t *testing.T) {
	tests := []struct {
		name       string
		extensions []string
		path       string
		wantError  bool
	}{
		// Basic extension matching
		{"single extension match", []string{".txt"}, "file.txt", false},
		{"single extension no match", []string{".txt"}, "file.go", true},
		{"multiple extensions match first", []string{".txt", ".md", ".go"}, "file.txt", false},
		{"multiple extensions match second", []string{".txt", ".md", ".go"}, "file.md", false},
		{"multiple extensions no match", []string{".txt", ".md"}, "file.go", true},

		// Case insensitivity
		{"case insensitive upper", []string{".txt"}, "file.TXT", false},
		{"case insensitive mixed", []string{".txt"}, "file.Txt", false},
		{"case insensitive lower", []string{".TXT"}, "file.txt", false},

		// Unicode case folding (Turkish İ is distinct from ASCII i)
		{"Turkish İ exact match", []string{".İ"}, "file.İ", false},
		{"Turkish İ not equal to ASCII i", []string{".i"}, "file.İ", true},
		{"ASCII i not equal to Turkish İ", []string{".İ"}, "file.i", true},

		// Without leading dot
		{"without dot match", []string{"txt"}, "file.txt", false},
		{"without dot no match", []string{"txt"}, "file.md", true},
		{"mixed with and without dot", []string{".txt", "md"}, "file.md", false},

		// Edge cases
		{"no extension in path", []string{".txt"}, "Makefile", true},
		{"double extension", []string{".gz"}, "archive.tar.gz", false},
		{"hidden file with extension", []string{".txt"}, ".hidden.txt", false},
		{"path with directory", []string{".go"}, "/path/to/file.go", false},
		{"empty extension list", []string{}, "file.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &ExtensionRule{Extensions: tt.extensions}
			err := rule.Validate(tt.path)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("rule name and description", func(t *testing.T) {
		rule := &ExtensionRule{Extensions: []string{".txt"}}
		assert.Equal(t, "extension", rule.Name())
		assert.NotEmpty(t, rule.Description())
	})
}

// =============================================================================
// MaxLengthRule Tests
// =============================================================================

// TestMaxLengthRule tests that paths must not exceed a maximum length.
func TestMaxLengthRule(t *testing.T) {
	tests := []struct {
		name      string
		maxLen    int
		path      string
		wantError bool
	}{
		{"under limit", 100, "/short/path.txt", false},
		{"at exact limit", 10, "/path.txt", false},
		{"over limit", 10, "/very/long/path.txt", true},
		{"zero length path", 100, "", false},
		{"max length zero", 0, "a", true},
		{"max length one", 1, "a", false},
		{"unicode characters", 100, "/путь/файл.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := &MaxLengthRule{MaxLength: tt.maxLen}
			err := rule.Validate(tt.path)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("rule name and description", func(t *testing.T) {
		rule := &MaxLengthRule{MaxLength: 100}
		assert.Equal(t, "max-length", rule.Name())
		assert.NotEmpty(t, rule.Description())
	})

	t.Run("ValidatePath uses Original path", func(t *testing.T) {
		// This tests that the Original() path is used for length validation
		// to prevent bypasses through path cleaning
		rule := &MaxLengthRule{MaxLength: 20}

		// PathBuilder with long original path
		pb := NewPathBuilder("/a/./b/../c/d/e/f/g/h/i/j/k/l")
		err := rule.ValidatePath(pb)
		// The original path is longer than cleaned, so should fail
		require.Error(t, err)
	})
}

// =============================================================================
// PatternRule Tests
// =============================================================================

// TestPatternRule tests regex pattern matching for required and forbidden patterns.
func TestPatternRule(t *testing.T) {
	t.Run("required pattern matches", func(t *testing.T) {
		rule := &PatternRule{Pattern: `\.go$`, Required: true}
		err := rule.Validate("main.go")
		require.NoError(t, err)
	})

	t.Run("required pattern does not match", func(t *testing.T) {
		rule := &PatternRule{Pattern: `\.go$`, Required: true}
		err := rule.Validate("main.txt")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrPatternMismatch)
	})

	t.Run("forbidden pattern does not match", func(t *testing.T) {
		rule := &PatternRule{Pattern: `\.exe$`, Required: false}
		err := rule.Validate("program.go")
		require.NoError(t, err)
	})

	t.Run("forbidden pattern matches", func(t *testing.T) {
		rule := &PatternRule{Pattern: `\.exe$`, Required: false}
		err := rule.Validate("program.exe")
		require.Error(t, err)
		require.ErrorIs(t, err, ErrForbiddenPattern)
	})

	t.Run("invalid regex returns error", func(t *testing.T) {
		rule := &PatternRule{Pattern: `[invalid`, Required: true}
		err := rule.Validate("any.path")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid pattern")
	})

	t.Run("complex pattern matching", func(t *testing.T) {
		rule := &PatternRule{Pattern: `^(src|lib)/.*\.(go|rs)$`, Required: true}

		require.NoError(t, rule.Validate("src/main.go"))
		require.NoError(t, rule.Validate("lib/utils.rs"))
		require.Error(t, rule.Validate("test/main.go"))
		require.Error(t, rule.Validate("src/main.txt"))
	})

	t.Run("rule name for required pattern", func(t *testing.T) {
		rule := &PatternRule{Pattern: `.*`, Required: true}
		assert.Equal(t, "require-pattern", rule.Name())
	})

	t.Run("rule name for forbidden pattern", func(t *testing.T) {
		rule := &PatternRule{Pattern: `.*`, Required: false}
		assert.Equal(t, "forbid-pattern", rule.Name())
	})
}

// =============================================================================
// PathTraversalRule Tests
// =============================================================================

// TestPathTraversalRule tests detection of path traversal attacks.
// Path traversal allows accessing files outside intended directories.
func TestPathTraversalRule(t *testing.T) {
	rule := &PathTraversalRule{}

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		// Safe paths
		{"normal path", "/home/user/file.txt", false},
		{"relative path", "src/main.go", false},
		{"path with single dot", "./file.txt", false},
		{"path ending with dot", "file.", false},

		// Basic path traversal
		{"double dot in path", "../etc/passwd", true},
		{"double dot in middle", "/home/../etc/passwd", true},
		{"multiple traversals", "../../../../../../etc/passwd", true},
		{"windows style traversal", "..\\..\\windows\\system32", true},

		// URL encoded traversal
		{"url encoded dots", "%2e%2e/etc/passwd", true},
		{"double url encoded", "%252e%252e/etc/passwd", true},

		// Unicode encoded
		{"unicode dots", "\u002e\u002e/etc/passwd", true},

		// Hex encoded
		{"hex encoded dots", "\x2e\x2e/etc/passwd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.path)
			if tt.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrPathTraversalDetected)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("rule name and description", func(t *testing.T) {
		assert.Equal(t, "no-path-traversal", rule.Name())
		assert.NotEmpty(t, rule.Description())
	})
}

// =============================================================================
// NullByteRule Tests
// =============================================================================

// TestNullByteRule tests detection of null bytes in paths.
// Null bytes can be used to truncate strings in some systems.
func TestNullByteRule(t *testing.T) {
	rule := &NullByteRule{}

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		// Safe paths
		{"normal path", "/home/user/file.txt", false},
		{"path with spaces", "/home/user/my file.txt", false},
		{"unicode path", "/home/пользователь/файл.txt", false},

		// Null byte attacks
		{"null byte at end", "file.txt\x00.jpg", true},
		{"null byte in middle", "file\x00malicious.txt", true},
		{"null byte at start", "\x00/etc/passwd", true},
		{"url encoded null", "file%00.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.path)
			if tt.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrPathContainsNullByte)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("rule name and description", func(t *testing.T) {
		assert.Equal(t, "no-null-bytes", rule.Name())
		assert.NotEmpty(t, rule.Description())
	})
}

// =============================================================================
// ControlCharacterRule Tests
// =============================================================================

// TestControlCharacterRule tests detection of control characters in paths.
// Control characters (0x00-0x1F except tab) can cause unexpected behavior.
func TestControlCharacterRule(t *testing.T) {
	rule := &ControlCharacterRule{}

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		// Safe paths
		{"normal path", "/home/user/file.txt", false},
		{"path with tab", "/home/user\tfile.txt", false}, // Tab is allowed
		{"path with spaces", "/home/user/my file.txt", false},

		// Control characters
		{"bell character", "file\x07.txt", true},
		{"backspace", "file\x08.txt", true},
		{"carriage return", "file\x0d.txt", true},
		{"line feed", "file\x0a.txt", true},
		{"escape", "file\x1b.txt", true},
		{"form feed", "file\x0c.txt", true},
		{"vertical tab", "file\x0b.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.path)
			if tt.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrPathContainsControlChar)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("rule name and description", func(t *testing.T) {
		assert.Equal(t, "no-control-chars", rule.Name())
		assert.NotEmpty(t, rule.Description())
	})
}

// =============================================================================
// WindowsReservedRule Tests
// =============================================================================

// TestWindowsReservedRule tests detection of Windows reserved device names.
// These names cannot be used as file names on Windows.
func TestWindowsReservedRule(t *testing.T) {
	rule := &WindowsReservedRule{}

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		// Safe paths
		{"normal file", "file.txt", false},
		{"normal path", "/path/to/file.txt", false},
		{"file starting with reserved", "CONFILE.txt", false},
		{"file ending with reserved", "fileCON.txt", false},
		{"file containing reserved", "myconfile.txt", false},

		// Reserved names
		{"CON", "CON", true},
		{"PRN", "PRN", true},
		{"AUX", "AUX", true},
		{"NUL", "NUL", true},
		{"COM1", "COM1", true},
		{"COM9", "COM9", true},
		{"LPT1", "LPT1", true},
		{"LPT9", "LPT9", true},
		{"CONIN$", "CONIN$", true},
		{"CONOUT$", "CONOUT$", true},

		// Case variations
		{"lowercase con", "con", true},
		{"mixed case Con", "Con", true},
		{"lowercase com1", "com1", true},

		// With extensions
		{"CON with extension", "CON.txt", true},
		{"PRN with extension", "PRN.doc", true},
		{"COM1 with extension", "COM1.sys", true},

		// In path
		{"reserved in path", "/path/to/CON", true},
		{"reserved in path with ext", "/path/to/NUL.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.path)
			if tt.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrWindowsReservedName)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("rule name and description", func(t *testing.T) {
		assert.Equal(t, "no-windows-reserved", rule.Name())
		assert.NotEmpty(t, rule.Description())
	})
}

// =============================================================================
// UNCPathRule Tests
// =============================================================================

// TestUNCPathRule tests detection of UNC (Universal Naming Convention) paths.
// UNC paths are Windows network paths starting with \\.
func TestUNCPathRule(t *testing.T) {
	rule := &UNCPathRule{}

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		// Safe paths
		{"normal unix path", "/home/user/file.txt", false},
		{"normal windows path", "C:\\Users\\file.txt", false},
		{"relative path", "path\\to\\file.txt", false},
		{"single backslash", "\\path\\file.txt", false},

		// UNC paths
		{"basic UNC path", "\\\\server\\share", true},
		{"UNC with file", "\\\\server\\share\\file.txt", true},
		{"UNC with IP", "\\\\192.168.1.1\\share", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.path)
			if tt.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrUNCPathNotAllowed)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("rule name and description", func(t *testing.T) {
		assert.Equal(t, "no-unc-paths", rule.Name())
		assert.NotEmpty(t, rule.Description())
	})
}

// =============================================================================
// DrivePathRule Tests
// =============================================================================

// TestDrivePathRule tests detection of Windows drive paths.
// Drive paths start with a letter followed by a colon (e.g., C:).
func TestDrivePathRule(t *testing.T) {
	rule := &DrivePathRule{}

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		// Safe paths
		{"normal unix path", "/home/user/file.txt", false},
		{"relative path", "path/to/file.txt", false},
		{"path with colon in name", "/path/file:alternate", false},
		{"empty path", "", false},
		{"single char", "a", false},

		// Drive paths
		{"C drive", "C:\\Users\\file.txt", true},
		{"D drive", "D:\\data\\file.txt", true},
		{"lowercase drive", "c:\\users\\file.txt", true},
		{"drive only", "C:", true},
		{"drive with forward slash", "C:/Users/file.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.path)
			if tt.wantError {
				require.Error(t, err)
				require.ErrorIs(t, err, ErrDrivePathNotAllowed)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("rule name and description", func(t *testing.T) {
		assert.Equal(t, "no-drive-paths", rule.Name())
		assert.NotEmpty(t, rule.Description())
	})
}

// =============================================================================
// ValidUTF8Rule Tests
// =============================================================================

// TestValidUTF8Rule tests detection of invalid UTF-8 byte sequences.
// Invalid UTF-8 can be used to bypass security filters.
func TestValidUTF8Rule(t *testing.T) {
	rule := &ValidUTF8Rule{}

	tests := []struct {
		name      string
		path      string
		wantError bool
	}{
		// Valid UTF-8
		{"ascii path", "/home/user/file.txt", false},
		{"unicode path", "/home/user/file.txt", false},
		{"emoji in path", "/home/user/file.txt", false},
		{"cyrillic chars", "/home/user/file.txt", false},
		{"mixed unicode", "/home/user/file.txt", false},

		// Invalid UTF-8 bytes
		{"0xff byte", "/home/\xff/file.txt", true},
		{"0xfe byte", "/home/\xfe/file.txt", true},

		// Overlong sequences
		{"0xc0 byte (overlong)", "/home/\xc0\xaf/file.txt", true},
		{"0xc1 byte (overlong)", "/home/\xc1\xaf/file.txt", true},

		// Invalid continuation bytes
		{"invalid continuation at start", "\x80path/file.txt", true},
		{"continuation after ascii", "a\x80path/file.txt", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rule.Validate(tt.path)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}

	t.Run("rule name and description", func(t *testing.T) {
		assert.Equal(t, "valid-utf8", rule.Name())
		assert.NotEmpty(t, rule.Description())
	})
}

// =============================================================================
// RequireSecure Tests
// =============================================================================

// TestValidator_RequireSecure tests the RequireSecure convenience method.
// This adds all security-related validation rules at once.
func TestValidator_RequireSecure(t *testing.T) {
	v := NewPathValidator()
	v.RequireSecure()

	rules := v.Rules()
	// Should have 7 security rules
	assert.Len(t, rules, 7)

	// Verify all security rules are present
	ruleNames := make([]string, len(rules))
	for i, r := range rules {
		ruleNames[i] = r.Name()
	}

	assert.Contains(t, ruleNames, "no-path-traversal")
	assert.Contains(t, ruleNames, "no-null-bytes")
	assert.Contains(t, ruleNames, "no-control-chars")
	assert.Contains(t, ruleNames, "no-windows-reserved")
	assert.Contains(t, ruleNames, "no-unc-paths")
	assert.Contains(t, ruleNames, "no-drive-paths")
	assert.Contains(t, ruleNames, "valid-utf8")

	// Test that a malicious path fails
	errs := v.Validate("../../../etc/passwd")
	assert.NotEmpty(t, errs)

	// Test that a safe path passes
	errs = v.Validate("safe/path/file.txt")
	assert.Empty(t, errs)
}

// =============================================================================
// Validator Chaining Tests
// =============================================================================

// TestValidator_Chaining tests that validator methods can be chained.
func TestValidator_Chaining(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0o600))

	v := NewPathValidator()
	v.RequireExists().RequireFile().RequireExtension(".go", ".rs")

	rules := v.Rules()
	assert.Len(t, rules, 3)

	// Valid file passes all rules
	errs := v.Validate(testFile)
	assert.Empty(t, errs)

	// Directory fails file rule
	errs = v.Validate(tmpDir)
	assert.NotEmpty(t, errs)

	// Wrong extension fails
	txtFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(txtFile, []byte("test"), 0o600))
	errs = v.Validate(txtFile)
	assert.NotEmpty(t, errs)
}

// =============================================================================
// Package-Level Convenience Functions Tests
// =============================================================================

// TestValidator_PackageFunctions tests package-level validation convenience functions.
func TestValidator_PackageFunctions(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0o600))

	t.Run("ValidateExists", func(t *testing.T) {
		require.NoError(t, ValidateExists(testFile))
		require.Error(t, ValidateExists(filepath.Join(tmpDir, "nonexistent")))
	})

	t.Run("ValidateReadable", func(t *testing.T) {
		require.NoError(t, ValidateReadable(testFile))
		require.Error(t, ValidateReadable(filepath.Join(tmpDir, "nonexistent")))
	})

	t.Run("ValidateWritable", func(t *testing.T) {
		require.NoError(t, ValidateWritable(testFile))
	})

	t.Run("ValidateExtension", func(t *testing.T) {
		require.NoError(t, ValidateExtension(testFile, ".txt", ".md"))
		require.Error(t, ValidateExtension(testFile, ".go", ".rs"))
	})

	t.Run("Validate with validator", func(t *testing.T) {
		v := NewPathValidator()
		v.RequireExists()

		errs := Validate(testFile, v)
		assert.Empty(t, errs)

		errs = Validate(filepath.Join(tmpDir, "nonexistent"), v)
		assert.NotEmpty(t, errs)
	})

	t.Run("IsValid with validator", func(t *testing.T) {
		v := NewPathValidator()
		v.RequireExists()

		assert.True(t, IsValid(testFile, v))
		assert.False(t, IsValid(filepath.Join(tmpDir, "nonexistent"), v))
	})
}

// =============================================================================
// Constants and Helper Method Tests
// =============================================================================

// TestWindowsReservedDeviceNames_Constant verifies the package-level constant
// contains all expected Windows reserved device names.
func TestWindowsReservedDeviceNames_Constant(t *testing.T) {
	expectedNames := []string{
		"CON", "PRN", "AUX", "NUL",
		"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
		"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
		"CONIN$", "CONOUT$",
	}

	assert.Len(t, windowsReservedDeviceNames, len(expectedNames),
		"windowsReservedDeviceNames should contain exactly %d names", len(expectedNames))

	// Verify all expected names are present
	nameSet := make(map[string]bool)
	for _, name := range windowsReservedDeviceNames {
		nameSet[name] = true
	}

	for _, expected := range expectedNames {
		assert.True(t, nameSet[expected],
			"windowsReservedDeviceNames should contain %q", expected)
	}
}

// TestAddBuiltInRule_Helper tests the addBuiltInRule helper method.
func TestAddBuiltInRule_Helper(t *testing.T) {
	t.Run("adds rule successfully", func(t *testing.T) {
		v := NewPathValidator()
		result := v.addBuiltInRule(&AbsolutePathRule{})

		// Should return the validator for chaining
		assert.Same(t, v, result)

		// Rule should be added
		rules := v.Rules()
		assert.Len(t, rules, 1)
		assert.Equal(t, "absolute-path", rules[0].Name())
	})

	t.Run("adds multiple rules sequentially", func(t *testing.T) {
		v := NewPathValidator()
		v.addBuiltInRule(&AbsolutePathRule{})
		v.addBuiltInRule(&ExistsRule{})
		v.addBuiltInRule(&ReadableRule{})

		rules := v.Rules()
		assert.Len(t, rules, 3)
		assert.Equal(t, "absolute-path", rules[0].Name())
		assert.Equal(t, "exists", rules[1].Name())
		assert.Equal(t, "readable", rules[2].Name())
	})
}

// TestRequireSecure_AddsAllSecurityRules verifies that RequireSecure
// adds all seven security validation rules in the expected order.
func TestRequireSecure_AddsAllSecurityRules(t *testing.T) {
	v := NewPathValidator()
	result := v.RequireSecure()

	// Should return the validator for chaining
	assert.Same(t, v, result)

	// Should have exactly 7 security rules
	rules := v.Rules()
	require.Len(t, rules, 7, "RequireSecure should add exactly 7 rules")

	// Verify rules are added in the expected order
	expectedRuleNames := []string{
		"no-path-traversal",
		"no-null-bytes",
		"no-control-chars",
		"no-windows-reserved",
		"no-unc-paths",
		"no-drive-paths",
		"valid-utf8",
	}

	for i, expectedName := range expectedRuleNames {
		assert.Equal(t, expectedName, rules[i].Name(),
			"Rule at position %d should be %q", i, expectedName)
	}
}

// TestBuilderMethods_AllReturnValidator verifies that all builder methods
// return the validator for chaining and properly add their rules.
func TestBuilderMethods_AllReturnValidator(t *testing.T) {
	testCases := []struct {
		name         string
		builderFunc  func(*DefaultPathValidator) PathValidator
		expectedRule string
	}{
		{"RequireAbsolute", func(v *DefaultPathValidator) PathValidator { return v.RequireAbsolute() }, "absolute-path"},
		{"RequireRelative", func(v *DefaultPathValidator) PathValidator { return v.RequireRelative() }, "relative-path"},
		{"RequireExists", func(v *DefaultPathValidator) PathValidator { return v.RequireExists() }, "exists"},
		{"RequireNotExists", func(v *DefaultPathValidator) PathValidator { return v.RequireNotExists() }, "not-exists"},
		{"RequireReadable", func(v *DefaultPathValidator) PathValidator { return v.RequireReadable() }, "readable"},
		{"RequireWritable", func(v *DefaultPathValidator) PathValidator { return v.RequireWritable() }, "writable"},
		{"RequireExecutable", func(v *DefaultPathValidator) PathValidator { return v.RequireExecutable() }, "executable"},
		{"RequireDirectory", func(v *DefaultPathValidator) PathValidator { return v.RequireDirectory() }, "directory"},
		{"RequireFile", func(v *DefaultPathValidator) PathValidator { return v.RequireFile() }, "file"},
		{"RequireExtension", func(v *DefaultPathValidator) PathValidator { return v.RequireExtension(".go", ".txt") }, "extension"},
		{"RequireMaxLength", func(v *DefaultPathValidator) PathValidator { return v.RequireMaxLength(255) }, "max-length"},
		{"RequirePattern", func(v *DefaultPathValidator) PathValidator { return v.RequirePattern(".*\\.go$") }, "require-pattern"},
		{"ForbidPattern", func(v *DefaultPathValidator) PathValidator { return v.ForbidPattern(".*\\.exe$") }, "forbid-pattern"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := NewPathValidator()
			result := tc.builderFunc(v)

			// Should return the validator for chaining
			assert.NotNil(t, result)

			// Rule should be added with expected name
			rules := v.Rules()
			require.Len(t, rules, 1, "%s should add exactly 1 rule", tc.name)
			assert.Equal(t, tc.expectedRule, rules[0].Name(),
				"%s should add rule with name %q", tc.name, tc.expectedRule)
		})
	}
}

// TestBuilderMethods_ChainCorrectly verifies that builder methods can be chained.
func TestBuilderMethods_ChainCorrectly(t *testing.T) {
	v := NewPathValidator()
	result := v.
		RequireAbsolute().
		RequireExists().
		RequireReadable().
		RequireFile().
		RequireExtension(".go", ".txt").
		RequireMaxLength(4096)

	// Should return the validator
	assert.NotNil(t, result)

	// All 6 rules should be added
	rules := v.Rules()
	assert.Len(t, rules, 6)

	// Verify rules are in order of addition
	assert.Equal(t, "absolute-path", rules[0].Name())
	assert.Equal(t, "exists", rules[1].Name())
	assert.Equal(t, "readable", rules[2].Name())
	assert.Equal(t, "file", rules[3].Name())
	assert.Equal(t, "extension", rules[4].Name())
	assert.Equal(t, "max-length", rules[5].Name())
}
