package paths

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test error for batch operations
var errBatchStop = errors.New("batch stop")

// =============================================================================
// Path Factory Tests
// =============================================================================

// TestConvenience_PathFactories tests File, Dir, Current, Home, and Root functions.
// These factory functions create PathBuilder instances for common path types.
func TestConvenience_PathFactories(t *testing.T) {
	t.Run("File creates PathBuilder", func(t *testing.T) {
		pb := File("/path/to/file.txt")
		assert.Equal(t, "/path/to/file.txt", pb.String())
	})

	t.Run("Dir creates PathBuilder", func(t *testing.T) {
		pb := Dir("/path/to/directory")
		assert.Equal(t, "/path/to/directory", pb.String())
	})

	t.Run("Current returns dot path", func(t *testing.T) {
		pb := Current()
		assert.Equal(t, ".", pb.String())
	})

	t.Run("Home returns tilde path", func(t *testing.T) {
		pb := Home()
		assert.Equal(t, "~", pb.String())
	})

	t.Run("Root returns root path", func(t *testing.T) {
		pb := Root()
		assert.Equal(t, "/", pb.String())
	})
}

// =============================================================================
// Path Manipulation Tests
// =============================================================================

// TestConvenience_PathManipulation tests Clean, Abs, Base, Ext, DirOf, and JoinPaths.
// These functions provide path manipulation without creating a PathBuilder first.
func TestConvenience_PathManipulation(t *testing.T) {
	t.Run("Clean normalizes path", func(t *testing.T) {
		pb := Clean("/path/to/../file.txt")
		assert.Equal(t, "/path/file.txt", pb.String())
	})

	t.Run("Clean handles double slashes", func(t *testing.T) {
		pb := Clean("/path//to///file.txt")
		assert.Equal(t, "/path/to/file.txt", pb.String())
	})

	t.Run("Abs returns absolute path", func(t *testing.T) {
		pb, err := Abs("relative/path.txt")
		require.NoError(t, err)
		assert.True(t, filepath.IsAbs(pb.String()), "Result should be absolute")
	})

	t.Run("Base returns filename", func(t *testing.T) {
		base := Base("/path/to/file.txt")
		assert.Equal(t, "file.txt", base)
	})

	t.Run("Base for directory", func(t *testing.T) {
		base := Base("/path/to/directory")
		assert.Equal(t, "directory", base)
	})

	t.Run("Ext returns extension", func(t *testing.T) {
		ext := Ext("/path/to/file.txt")
		assert.Equal(t, ".txt", ext)
	})

	t.Run("Ext for no extension", func(t *testing.T) {
		ext := Ext("/path/to/Makefile")
		assert.Empty(t, ext)
	})

	t.Run("DirOf returns directory", func(t *testing.T) {
		dir := DirOf("/path/to/file.txt")
		assert.Equal(t, "/path/to", dir.String())
	})

	t.Run("JoinPaths joins elements", func(t *testing.T) {
		joined := JoinPaths("path", "to", "file.txt")
		expected := filepath.Join("path", "to", "file.txt")
		assert.Equal(t, expected, joined.String())
	})
}

// =============================================================================
// Matcher Convenience Functions Tests
// =============================================================================

// TestConvenience_MatcherFunctions tests MatchPattern, MatchAnyPattern, and CreateMatcher.
func TestConvenience_MatcherFunctions(t *testing.T) {
	t.Run("MatchPattern matches", func(t *testing.T) {
		assert.True(t, MatchPattern("*.go", "main.go"))
		assert.False(t, MatchPattern("*.go", "main.txt"))
	})

	t.Run("MatchAnyPattern matches any", func(t *testing.T) {
		assert.True(t, MatchAnyPattern("main.go", "*.go", "*.txt"))
		assert.True(t, MatchAnyPattern("readme.txt", "*.go", "*.txt"))
		assert.False(t, MatchAnyPattern("config.json", "*.go", "*.txt"))
	})

	t.Run("CreateMatcher with valid patterns", func(t *testing.T) {
		matcher, err := CreateMatcher("*.go", "*.txt")
		require.NoError(t, err)
		require.NotNil(t, matcher)

		assert.True(t, matcher.Match("main.go"))
		assert.True(t, matcher.Match("readme.txt"))
		assert.False(t, matcher.Match("config.json"))
	})

	t.Run("CreateMatcher with empty pattern returns error", func(t *testing.T) {
		matcher, err := CreateMatcher("*.go", "")
		require.Error(t, err)
		assert.Nil(t, matcher)
	})

	t.Run("CreateMatcher with invalid regex returns error", func(t *testing.T) {
		matcher, err := CreateMatcher("^[invalid")
		require.Error(t, err)
		assert.Nil(t, matcher)
	})

	t.Run("CreateMatcher with no patterns", func(t *testing.T) {
		matcher, err := CreateMatcher()
		require.NoError(t, err)
		require.NotNil(t, matcher)

		// Empty matcher matches nothing
		assert.False(t, matcher.Match("any.file"))
	})
}

// =============================================================================
// Validator Convenience Functions Tests
// =============================================================================

// TestConvenience_ValidatorFunctions tests ValidatePath, ValidatePathExists,
// ValidatePathReadable, ValidatePathWritable, CreateValidator, CreateStrictValidator.
func TestConvenience_ValidatorFunctions(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(testFile, []byte("test"), 0o600))

	t.Run("ValidatePath with no rules passes", func(t *testing.T) {
		// ValidatePath creates an empty validator, so it always passes
		errors := ValidatePath("any/path")
		assert.Empty(t, errors)
	})

	t.Run("ValidatePathExists for existing file", func(t *testing.T) {
		err := ValidatePathExists(testFile)
		require.NoError(t, err)
	})

	t.Run("ValidatePathExists for non-existent file", func(t *testing.T) {
		err := ValidatePathExists(filepath.Join(tmpDir, "nonexistent"))
		require.Error(t, err)
	})

	t.Run("ValidatePathReadable for readable file", func(t *testing.T) {
		err := ValidatePathReadable(testFile)
		require.NoError(t, err)
	})

	t.Run("ValidatePathReadable for non-existent file", func(t *testing.T) {
		err := ValidatePathReadable(filepath.Join(tmpDir, "nonexistent"))
		require.Error(t, err)
	})

	t.Run("ValidatePathWritable for writable directory", func(t *testing.T) {
		err := ValidatePathWritable(tmpDir)
		require.NoError(t, err)
	})

	t.Run("CreateValidator creates empty validator", func(t *testing.T) {
		v := CreateValidator()
		require.NotNil(t, v)

		// Empty validator has no rules
		errors := v.Validate("any/path")
		assert.Empty(t, errors)
	})

	t.Run("CreateStrictValidator has predefined rules", func(t *testing.T) {
		v := CreateStrictValidator()
		require.NotNil(t, v)

		// Should fail for relative path (RequireAbsolute)
		errors := v.Validate("relative/path")
		assert.NotEmpty(t, errors, "Strict validator should fail for relative path")
	})

	t.Run("CreateStrictValidator passes for valid absolute existing file", func(t *testing.T) {
		v := CreateStrictValidator()

		// Create an absolute, existing, readable file
		absFile, err := filepath.Abs(testFile)
		require.NoError(t, err)

		errors := v.Validate(absFile)
		assert.Empty(t, errors, "Strict validator should pass for absolute existing readable file")
	})
}

// =============================================================================
// Set Convenience Functions Tests
// =============================================================================

// TestConvenience_SetFunctions tests CreateSet, CreateSetFromBuilders, UnionAll, IntersectionAll.
func TestConvenience_SetFunctions(t *testing.T) {
	t.Run("CreateSet from paths", func(t *testing.T) {
		set := CreateSet("/a.txt", "/b.txt", "/c.txt")
		assert.Equal(t, 3, set.Size())
		assert.True(t, set.Contains("/a.txt"))
	})

	t.Run("CreateSet with no paths", func(t *testing.T) {
		set := CreateSet()
		assert.True(t, set.IsEmpty())
	})

	t.Run("CreateSetFromBuilders", func(t *testing.T) {
		pb1 := NewPathBuilder("/a.txt")
		pb2 := NewPathBuilder("/b.txt")
		set := CreateSetFromBuilders(pb1, pb2)

		assert.Equal(t, 2, set.Size())
		assert.True(t, set.Contains("/a.txt"))
		assert.True(t, set.Contains("/b.txt"))
	})

	t.Run("UnionAll multiple sets", func(t *testing.T) {
		set1 := CreateSet("/a.txt")
		set2 := CreateSet("/b.txt")
		set3 := CreateSet("/c.txt")

		union := UnionAll(set1, set2, set3)
		assert.Equal(t, 3, union.Size())
	})

	t.Run("IntersectionAll multiple sets", func(t *testing.T) {
		set1 := CreateSet("/a.txt", "/b.txt", "/c.txt")
		set2 := CreateSet("/b.txt", "/c.txt", "/d.txt")
		set3 := CreateSet("/c.txt", "/d.txt", "/e.txt")

		intersection := IntersectionAll(set1, set2, set3)
		assert.Equal(t, 1, intersection.Size())
		assert.True(t, intersection.Contains("/c.txt"))
	})
}

// =============================================================================
// File System Operation Tests
// =============================================================================

// TestConvenience_FileSystemOperations tests CreateFile, CreateDirectory, RemovePath, RemoveAll, CopyPath, MovePath.
func TestConvenience_FileSystemOperations(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("CreateFile creates file", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "created.txt")
		err := CreateFile(testFile)
		require.NoError(t, err)

		assert.True(t, PathExists(testFile))
		assert.True(t, PathIsFile(testFile))
	})

	t.Run("CreateDirectory creates directory", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "created_dir", "subdir")
		err := CreateDirectory(testDir)
		require.NoError(t, err)

		assert.True(t, PathExists(testDir))
		assert.True(t, PathIsDir(testDir))
	})

	t.Run("RemovePath removes file", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "to_remove.txt")
		require.NoError(t, CreateFile(testFile))

		err := RemovePath(testFile)
		require.NoError(t, err)
		assert.False(t, PathExists(testFile))
	})

	t.Run("RemoveAll removes directory tree", func(t *testing.T) {
		dirPath := filepath.Join(tmpDir, "dir_to_remove")
		subFile := filepath.Join(dirPath, "subdir", "file.txt")
		require.NoError(t, CreateDirectory(filepath.Join(dirPath, "subdir")))
		require.NoError(t, CreateFile(subFile))

		err := RemoveAll(dirPath)
		require.NoError(t, err)
		assert.False(t, PathExists(dirPath))
	})

	t.Run("CopyPath copies file", func(t *testing.T) {
		src := filepath.Join(tmpDir, "copy_src.txt")
		dst := filepath.Join(tmpDir, "copy_dst.txt")

		require.NoError(t, os.WriteFile(src, []byte("content"), 0o600))

		err := CopyPath(src, dst)
		require.NoError(t, err)

		assert.True(t, PathExists(dst))
		//nolint:gosec // G304: test file path is controlled
		content, err := os.ReadFile(dst)
		require.NoError(t, err)
		assert.Equal(t, "content", string(content))
	})

	t.Run("MovePath moves file", func(t *testing.T) {
		src := filepath.Join(tmpDir, "move_src.txt")
		dst := filepath.Join(tmpDir, "move_dst.txt")

		require.NoError(t, os.WriteFile(src, []byte("moved"), 0o600))

		err := MovePath(src, dst)
		require.NoError(t, err)

		assert.False(t, PathExists(src))
		assert.True(t, PathExists(dst))
	})
}

// =============================================================================
// Information Shortcuts Tests
// =============================================================================

// TestConvenience_InformationShortcuts tests PathExists, PathIsDir, PathIsFile, PathSize, PathModTime.
func TestConvenience_InformationShortcuts(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "info_test.txt")
	testContent := "test content"
	require.NoError(t, os.WriteFile(testFile, []byte(testContent), 0o600))

	t.Run("PathExists for existing file", func(t *testing.T) {
		assert.True(t, PathExists(testFile))
	})

	t.Run("PathExists for non-existent file", func(t *testing.T) {
		assert.False(t, PathExists(filepath.Join(tmpDir, "nonexistent")))
	})

	t.Run("PathIsDir for directory", func(t *testing.T) {
		assert.True(t, PathIsDir(tmpDir))
	})

	t.Run("PathIsDir for file returns false", func(t *testing.T) {
		assert.False(t, PathIsDir(testFile))
	})

	t.Run("PathIsFile for file", func(t *testing.T) {
		assert.True(t, PathIsFile(testFile))
	})

	t.Run("PathIsFile for directory returns false", func(t *testing.T) {
		assert.False(t, PathIsFile(tmpDir))
	})

	t.Run("PathSize returns correct size", func(t *testing.T) {
		size := PathSize(testFile)
		assert.Equal(t, int64(len(testContent)), size)
	})

	t.Run("PathModTime returns valid time", func(t *testing.T) {
		modTime := PathModTime(testFile)
		assert.False(t, modTime.IsZero(), "ModTime should not be zero")
	})
}

// =============================================================================
// Find Operations Tests
// =============================================================================

// TestConvenience_FindOperations tests FindFiles and FindDirectories.
func TestConvenience_FindOperations(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test structure
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file1.go"), []byte(""), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file2.go"), []byte(""), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "file3.txt"), []byte(""), 0o600))
	require.NoError(t, os.Mkdir(filepath.Join(tmpDir, "subdir1"), 0o700))
	require.NoError(t, os.Mkdir(filepath.Join(tmpDir, "subdir2"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "subdir1", "nested.go"), []byte(""), 0o600))

	t.Run("FindFiles with pattern", func(t *testing.T) {
		files, err := FindFiles(tmpDir, "*.go")
		require.NoError(t, err)
		// Should find file1.go, file2.go, and subdir1/nested.go
		assert.GreaterOrEqual(t, len(files), 2, "Should find at least 2 .go files")
	})

	t.Run("FindFiles with non-directory returns error", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "file1.go")
		_, err := FindFiles(testFile, "*.go")
		require.Error(t, err)
	})

	t.Run("FindDirectories with pattern", func(t *testing.T) {
		dirs, err := FindDirectories(tmpDir, "subdir*")
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(dirs), 2, "Should find at least 2 subdirs")
	})

	t.Run("FindDirectories with non-directory returns error", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "file1.go")
		_, err := FindDirectories(testFile, "*")
		require.Error(t, err)
	})
}

// =============================================================================
// Safe/Secure Path Tests
// =============================================================================

// TestConvenience_SafeSecurePath tests SafePath and SecurePath functions.
func TestConvenience_SafeSecurePath(t *testing.T) {
	t.Run("SafePath with valid path", func(t *testing.T) {
		pb, err := SafePath("/safe/path.txt")
		require.NoError(t, err)
		assert.NotNil(t, pb)
	})

	t.Run("SecurePath within base path", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Create the path within the base directory
		securePath := filepath.Join(tmpDir, "subdir", "file.txt")
		pb, err := SecurePath(securePath, tmpDir)
		require.NoError(t, err)
		assert.NotNil(t, pb)
	})

	t.Run("SecurePath with traversal attempt", func(t *testing.T) {
		tmpDir := t.TempDir()
		_, err := SecurePath("../../../etc/passwd", tmpDir)
		require.Error(t, err, "SecurePath should reject path traversal")
	})
}

// =============================================================================
// Batch Operations Tests
// =============================================================================

// TestConvenience_BatchOperations tests ProcessPaths and FilterPaths.
func TestConvenience_BatchOperations(t *testing.T) {
	t.Run("ProcessPaths processes all paths", func(t *testing.T) {
		paths := []string{"/a.txt", "/b.txt", "/c.txt"}
		var processed []string

		err := ProcessPaths(paths, func(pb PathBuilder) error {
			processed = append(processed, pb.String())
			return nil
		})

		require.NoError(t, err)
		assert.Len(t, processed, 3)
	})

	t.Run("ProcessPaths stops on error", func(t *testing.T) {
		paths := []string{"/a.txt", "/b.txt", "/c.txt"}
		count := 0

		err := ProcessPaths(paths, func(pb PathBuilder) error {
			count++
			if count == 2 {
				return errBatchStop
			}
			return nil
		})

		require.Error(t, err)
		require.ErrorIs(t, err, errBatchStop)
		assert.Equal(t, 2, count)
	})

	t.Run("FilterPaths with predicate", func(t *testing.T) {
		paths := []string{"/file.go", "/file.txt", "/other.go", "/readme.md"}

		filtered := FilterPaths(paths, func(pb PathBuilder) bool {
			return pb.Ext() == ".go"
		})

		assert.Len(t, filtered, 2)
		assert.Equal(t, "/file.go", filtered[0].String())
		assert.Equal(t, "/other.go", filtered[1].String())
	})

	t.Run("FilterPaths with no matches", func(t *testing.T) {
		paths := []string{"/file.txt", "/readme.md"}

		filtered := FilterPaths(paths, func(pb PathBuilder) bool {
			return pb.Ext() == ".go"
		})

		assert.Empty(t, filtered)
	})
}

// =============================================================================
// Default Configuration Tests
// =============================================================================

// TestConvenience_DefaultConfiguration tests GetDefaultOptions and SetDefaultOptions.
func TestConvenience_DefaultConfiguration(t *testing.T) {
	t.Run("GetDefaultOptions returns valid options", func(t *testing.T) {
		opts := GetDefaultOptions()
		// Should have sensible defaults
		assert.True(t, opts.CreateParents, "CreateParents should default to true")
		assert.Positive(t, opts.BufferSize, "BufferSize should be positive")
	})

	t.Run("SetDefaultOptions changes options", func(t *testing.T) {
		// Save original options
		originalOpts := GetDefaultOptions()

		// Set new options
		newOpts := PathOptions{
			CreateMode:    0o644,
			CreateParents: false,
			BufferSize:    4096,
		}
		SetDefaultOptions(newOpts)

		// Verify change
		currentOpts := GetDefaultOptions()
		assert.Equal(t, newOpts.BufferSize, currentOpts.BufferSize)

		// Restore original options
		SetDefaultOptions(originalOpts)
	})
}

// =============================================================================
// Global Default Instance Tests
// =============================================================================

// TestConvenience_GlobalDefaults tests GetDefault* and SetDefault* functions.
func TestConvenience_GlobalDefaults(t *testing.T) {
	t.Run("GetDefaultBuilder returns builder", func(t *testing.T) {
		builder := GetDefaultBuilder()
		assert.NotNil(t, builder)
	})

	t.Run("GetDefaultMatcher returns matcher", func(t *testing.T) {
		matcher := GetDefaultMatcher()
		assert.NotNil(t, matcher)
	})

	t.Run("GetDefaultValidator returns validator", func(t *testing.T) {
		validator := GetDefaultValidator()
		assert.NotNil(t, validator)
	})

	t.Run("GetDefaultSet returns set", func(t *testing.T) {
		set := GetDefaultSet()
		assert.NotNil(t, set)
	})

	t.Run("GetDefaultCache returns cache", func(t *testing.T) {
		cache := GetDefaultCache()
		assert.NotNil(t, cache)
	})

	t.Run("SetDefaultBuilder with nil is ignored", func(t *testing.T) {
		original := GetDefaultBuilder()
		SetDefaultBuilder(nil)
		current := GetDefaultBuilder()
		assert.Equal(t, original.String(), current.String())
	})
}
