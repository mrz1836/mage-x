package paths

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPackageLevelBuilderFunctions tests the package-level builder functions
func TestPackageLevelBuilderFunctions(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0o600)
	require.NoError(t, err)

	// Create a test directory
	testDir := filepath.Join(tmpDir, "subdir")
	err = os.MkdirAll(testDir, 0o750)
	require.NoError(t, err)

	t.Run("Build", func(t *testing.T) {
		result := Build("foo", "bar", "baz")
		assert.NotNil(t, result)
		assert.Contains(t, result.String(), "bar")
	})

	t.Run("File", func(t *testing.T) {
		result := File(testFile)
		assert.NotNil(t, result)
		assert.Equal(t, testFile, result.String())
	})

	t.Run("Dir", func(t *testing.T) {
		result := Dir(tmpDir)
		assert.NotNil(t, result)
		assert.Equal(t, tmpDir, result.String())
	})

	t.Run("Current", func(t *testing.T) {
		result := Current()
		assert.NotNil(t, result)
		assert.True(t, result.Exists())
	})

	t.Run("Home", func(t *testing.T) {
		result := Home()
		assert.NotNil(t, result)
	})

	t.Run("Root", func(t *testing.T) {
		result := Root()
		assert.NotNil(t, result)
	})

	t.Run("Clean", func(t *testing.T) {
		result := Clean(filepath.Join(tmpDir, "..", filepath.Base(tmpDir)))
		assert.NotNil(t, result)
	})

	t.Run("Abs", func(t *testing.T) {
		result, err := Abs(".")
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("Base", func(t *testing.T) {
		result := Base(testFile)
		assert.Equal(t, "test.txt", result)
	})

	t.Run("Ext", func(t *testing.T) {
		result := Ext(testFile)
		assert.Equal(t, ".txt", result)
	})

	t.Run("DirOf", func(t *testing.T) {
		result := DirOf(testFile)
		assert.NotNil(t, result)
		assert.Equal(t, tmpDir, result.String())
	})

	t.Run("JoinPaths", func(t *testing.T) {
		result := JoinPaths("foo", "bar", "baz")
		assert.NotNil(t, result)
	})

	t.Run("PathExists", func(t *testing.T) {
		assert.True(t, PathExists(testFile))
		assert.False(t, PathExists(filepath.Join(tmpDir, "nonexistent")))
	})

	t.Run("PathIsDir", func(t *testing.T) {
		assert.True(t, PathIsDir(testDir))
		assert.False(t, PathIsDir(testFile))
	})

	t.Run("PathIsFile", func(t *testing.T) {
		assert.True(t, PathIsFile(testFile))
		assert.False(t, PathIsFile(testDir))
	})

	t.Run("PathSize", func(t *testing.T) {
		size := PathSize(testFile)
		assert.Equal(t, int64(4), size) // "test" is 4 bytes
	})

	t.Run("PathModTime", func(t *testing.T) {
		modTime := PathModTime(testFile)
		assert.False(t, modTime.IsZero())
	})

	t.Run("CreateFile", func(t *testing.T) {
		newFile := filepath.Join(tmpDir, "created.txt")
		err := CreateFile(newFile)
		require.NoError(t, err)
		assert.True(t, PathExists(newFile))
	})

	t.Run("CreateDirectory", func(t *testing.T) {
		newDir := filepath.Join(tmpDir, "newdir")
		err := CreateDirectory(newDir)
		require.NoError(t, err)
		assert.True(t, PathIsDir(newDir))
	})

	t.Run("CopyPath", func(t *testing.T) {
		src := testFile
		dst := filepath.Join(tmpDir, "copy.txt")
		err := CopyPath(src, dst)
		require.NoError(t, err)
		assert.True(t, PathExists(dst))
	})

	t.Run("MovePath", func(t *testing.T) {
		src := filepath.Join(tmpDir, "tomove.txt")
		err := os.WriteFile(src, []byte("move me"), 0o600)
		require.NoError(t, err)

		dst := filepath.Join(tmpDir, "moved.txt")
		err = MovePath(src, dst)
		require.NoError(t, err)
		assert.True(t, PathExists(dst))
		assert.False(t, PathExists(src))
	})

	t.Run("RemovePath", func(t *testing.T) {
		toRemove := filepath.Join(tmpDir, "toremove.txt")
		err := os.WriteFile(toRemove, []byte("remove me"), 0o600)
		require.NoError(t, err)

		err = RemovePath(toRemove)
		require.NoError(t, err)
		assert.False(t, PathExists(toRemove))
	})

	t.Run("RemoveAll", func(t *testing.T) {
		toRemoveDir := filepath.Join(tmpDir, "toremovedir")
		err := os.MkdirAll(filepath.Join(toRemoveDir, "sub"), 0o750)
		require.NoError(t, err)

		err = RemoveAll(toRemoveDir)
		require.NoError(t, err)
		assert.False(t, PathExists(toRemoveDir))
	})

	t.Run("FindFiles", func(t *testing.T) {
		files, err := FindFiles(tmpDir, "*.txt")
		require.NoError(t, err)
		assert.NotEmpty(t, files)
	})

	t.Run("FindDirectories", func(t *testing.T) {
		dirs, err := FindDirectories(tmpDir, "*")
		require.NoError(t, err)
		// May be empty if no subdirs match
		_ = dirs
	})

	t.Run("SafePath", func(t *testing.T) {
		result, err := SafePath(testFile)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("SecurePath", func(t *testing.T) {
		result, err := SecurePath(testFile, tmpDir)
		require.NoError(t, err)
		assert.NotNil(t, result)
	})
}

// TestPackageLevelMatcherFunctions tests the package-level matcher functions
func TestPackageLevelMatcherFunctions(t *testing.T) {
	t.Run("MatchPattern", func(t *testing.T) {
		assert.True(t, MatchPattern("*.txt", "file.txt"))
		assert.False(t, MatchPattern("*.txt", "file.go"))
	})

	t.Run("MatchAnyPattern", func(t *testing.T) {
		assert.True(t, MatchAnyPattern("file.txt", "*.txt", "*.go"))
		assert.False(t, MatchAnyPattern("file.md", "*.txt", "*.go"))
	})

	t.Run("CreateMatcher", func(t *testing.T) {
		matcher, err := CreateMatcher("*.txt", "*.go")
		require.NoError(t, err)
		assert.NotNil(t, matcher)

		// Match returns a single bool value
		matched := matcher.Match("file.txt")
		assert.True(t, matched)
	})
}

// TestPackageLevelValidatorFunctions tests the package-level validator functions
func TestPackageLevelValidatorFunctions(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0o600)
	require.NoError(t, err)

	t.Run("ValidatePath", func(t *testing.T) {
		errors := ValidatePath(testFile)
		assert.Empty(t, errors)
	})

	t.Run("ValidatePathExists", func(t *testing.T) {
		err := ValidatePathExists(testFile)
		require.NoError(t, err)

		err = ValidatePathExists(filepath.Join(tmpDir, "nonexistent"))
		assert.Error(t, err)
	})

	t.Run("ValidatePathReadable", func(t *testing.T) {
		err := ValidatePathReadable(testFile)
		assert.NoError(t, err)
	})

	t.Run("ValidatePathWritable", func(t *testing.T) {
		err := ValidatePathWritable(testFile)
		assert.NoError(t, err)
	})

	t.Run("CreateValidator", func(t *testing.T) {
		v := CreateValidator()
		assert.NotNil(t, v)
	})

	t.Run("CreateStrictValidator", func(t *testing.T) {
		v := CreateStrictValidator()
		assert.NotNil(t, v)
	})
}

// TestPackageLevelSetFunctions tests the package-level set functions
func TestPackageLevelSetFunctions(t *testing.T) {
	t.Run("CreateSet", func(t *testing.T) {
		set := CreateSet("path1", "path2", "path3")
		assert.Equal(t, 3, set.Size())
	})

	t.Run("CreateSetFromBuilders", func(t *testing.T) {
		b1 := NewPathBuilder("path1")
		b2 := NewPathBuilder("path2")
		set := CreateSetFromBuilders(b1, b2)
		assert.Equal(t, 2, set.Size())
	})

	t.Run("UnionAll", func(t *testing.T) {
		set1 := CreateSet("a", "b")
		set2 := CreateSet("c", "d")
		set3 := CreateSet("e", "f")
		union := UnionAll(set1, set2, set3)
		assert.Equal(t, 6, union.Size())
	})

	t.Run("IntersectionAll", func(t *testing.T) {
		set1 := CreateSet("a", "b", "c")
		set2 := CreateSet("b", "c", "d")
		set3 := CreateSet("c", "d", "e")
		intersection := IntersectionAll(set1, set2, set3)
		assert.Equal(t, 1, intersection.Size())
		assert.True(t, intersection.Contains("c"))
	})
}

// TestSetDefaultFunctions tests the SetDefault* functions
func TestSetDefaultFunctions(t *testing.T) {
	t.Run("SetDefaultBuilder", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		newBuilder := NewPathBuilder("/tmp")
		SetDefaultBuilder(newBuilder)

		current := GetDefaultBuilder()
		assert.Equal(t, newBuilder.String(), current.String())
	})

	t.Run("SetDefaultMatcher", func(t *testing.T) {
		original := GetDefaultMatcher()
		defer SetDefaultMatcher(original)

		newMatcher := NewPathMatcher()
		SetDefaultMatcher(newMatcher)

		current := GetDefaultMatcher()
		assert.Same(t, newMatcher, current)
	})

	t.Run("SetDefaultValidator", func(t *testing.T) {
		original := GetDefaultValidator()
		defer SetDefaultValidator(original)

		newValidator := NewPathValidator()
		SetDefaultValidator(newValidator)

		current := GetDefaultValidator()
		assert.Same(t, newValidator, current)
	})

	t.Run("SetDefaultSet", func(t *testing.T) {
		original := GetDefaultSet()
		defer SetDefaultSet(original)

		newSet := NewPathSet()
		SetDefaultSet(newSet)

		current := GetDefaultSet()
		assert.Same(t, newSet, current)
	})

	t.Run("SetDefaultCache", func(t *testing.T) {
		original := GetDefaultCache()
		defer SetDefaultCache(original)

		newCache := NewPathCache()
		SetDefaultCache(newCache)

		current := GetDefaultCache()
		assert.Same(t, newCache, current)
	})

	t.Run("SetDefaultWatcher", func(t *testing.T) {
		original := GetDefaultWatcher()
		defer func() {
			if original != nil {
				SetDefaultWatcher(original)
			}
		}()

		newWatcher := NewPathWatcher()
		SetDefaultWatcher(newWatcher)

		current := GetDefaultWatcher()
		assert.Same(t, newWatcher, current)
	})
}

// TestPathUtilityFunctions tests additional path utility functions
func TestPathUtilityFunctions(t *testing.T) {
	t.Run("MapPaths", func(t *testing.T) {
		paths := []string{"file1.txt", "file2.txt"}
		// MapPaths takes func(PathBuilder) PathBuilder and returns []PathBuilder
		mapped := MapPaths(paths, func(pb PathBuilder) PathBuilder {
			return NewPathBuilder("prefix_" + pb.String())
		})
		require.Len(t, mapped, 2)
		assert.Equal(t, "prefix_file1.txt", mapped[0].String())
		assert.Equal(t, "prefix_file2.txt", mapped[1].String())
	})

	t.Run("SortPaths", func(t *testing.T) {
		// SortPaths takes []PathBuilder not []string
		paths := []PathBuilder{
			NewPathBuilder("c.txt"),
			NewPathBuilder("a.txt"),
			NewPathBuilder("b.txt"),
		}
		sorted := SortPaths(paths)
		require.Len(t, sorted, 3)
		assert.Equal(t, "a.txt", sorted[0].String())
		assert.Equal(t, "b.txt", sorted[1].String())
		assert.Equal(t, "c.txt", sorted[2].String())
	})

	t.Run("AnalyzePath", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "analyze.txt")
		err := os.WriteFile(testFile, []byte("test content"), 0o600)
		require.NoError(t, err)

		// AnalyzePath returns (*PathInfo, error)
		analysis, err := AnalyzePath(testFile)
		require.NoError(t, err)
		require.NotNil(t, analysis)
		assert.True(t, analysis.IsFile)
		assert.False(t, analysis.IsDir)
	})

	t.Run("ProcessPaths", func(t *testing.T) {
		paths := []string{"a", "b", "c"}
		var processed []string
		// ProcessPaths expects func(PathBuilder) error
		err := ProcessPaths(paths, func(pb PathBuilder) error {
			processed = append(processed, pb.String()+"_processed")
			return nil
		})
		require.NoError(t, err)
		assert.Len(t, processed, 3)
	})

	t.Run("FilterPaths", func(t *testing.T) {
		paths := []string{"file.txt", "file.go", "file.md"}
		// FilterPaths expects func(PathBuilder) bool and returns []PathBuilder
		filtered := FilterPaths(paths, func(pb PathBuilder) bool {
			return pb.Ext() == ".txt"
		})
		require.Len(t, filtered, 1)
		assert.Equal(t, "file.txt", filtered[0].String())
	})
}

// TestBuilderModeFunctions tests the Mode function in PathBuilder
func TestBuilderModeFunctions(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "mode_test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0o600)
	require.NoError(t, err)

	builder := NewPathBuilder(testFile)
	mode := builder.Mode()
	assert.NotZero(t, mode)
}

// TestSymlinkFunction tests the Symlink function
func TestSymlinkFunction(t *testing.T) {
	tmpDir := t.TempDir()

	// Create the target file that the symlink will point to
	target := filepath.Join(tmpDir, "target.txt")
	err := os.WriteFile(target, []byte("target content"), 0o600)
	require.NoError(t, err)

	// The symlink location
	link := filepath.Join(tmpDir, "link.txt")

	// Symlink(target) creates symlink at pb.path pointing to target
	// So builder is the link location, and argument is what it points to
	builder := NewPathBuilder(link)
	err = builder.Symlink(NewPathBuilder(target))
	require.NoError(t, err)

	// Verify symlink was created
	linkInfo, err := os.Lstat(link)
	require.NoError(t, err)
	assert.NotEqual(t, 0, linkInfo.Mode()&os.ModeSymlink)
}

// TestCopyDirFunction tests the Copy function with directories
func TestCopyDirFunction(t *testing.T) {
	tmpDir := t.TempDir()

	// Create source directory with files
	srcDir := filepath.Join(tmpDir, "srcdir")
	err := os.MkdirAll(filepath.Join(srcDir, "subdir"), 0o750)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("content2"), 0o600)
	require.NoError(t, err)

	// Copy directory - Copy expects PathBuilder not string
	dstDir := filepath.Join(tmpDir, "dstdir")
	builder := NewPathBuilder(srcDir)
	err = builder.Copy(NewPathBuilder(dstDir))
	require.NoError(t, err)

	// Verify copy
	assert.True(t, PathExists(filepath.Join(dstDir, "file1.txt")))
	assert.True(t, PathExists(filepath.Join(dstDir, "subdir", "file2.txt")))
}

// TestMockPathCacheFunctions tests the MockPathCache wrapper functions
func TestMockPathCacheFunctions(t *testing.T) {
	mock := NewMockPathCache()
	require.NotNil(t, mock)

	t.Run("Set and Get", func(t *testing.T) {
		// MockPathCache.Set expects PathBuilder not string
		err := mock.Set("key1", NewPathBuilder("/path/to/value1"))
		require.NoError(t, err)
		value, found := mock.Get("key1")
		assert.True(t, found)
		assert.Equal(t, "/path/to/value1", value.String())
	})

	t.Run("Delete", func(t *testing.T) {
		err := mock.Set("todelete", NewPathBuilder("/tmp/value"))
		require.NoError(t, err)
		err = mock.Delete("todelete")
		require.NoError(t, err)
		_, found := mock.Get("todelete")
		assert.False(t, found)
	})

	t.Run("Clear", func(t *testing.T) {
		err := mock.Set("key1", NewPathBuilder("/tmp/v1"))
		require.NoError(t, err)
		err = mock.Clear()
		require.NoError(t, err)
		assert.Equal(t, 0, mock.Size())
	})

	t.Run("Stats", func(t *testing.T) {
		stats := mock.Stats()
		_ = stats // CacheStats is a value type
	})

	t.Run("Keys", func(t *testing.T) {
		m := NewMockPathCache()
		err := m.Set("key1", NewPathBuilder("/tmp/v1"))
		require.NoError(t, err)
		keys := m.Keys()
		assert.Contains(t, keys, "key1")
	})

	t.Run("Size", func(t *testing.T) {
		m := NewMockPathCache()
		err := m.Set("key1", NewPathBuilder("/tmp/v1"))
		require.NoError(t, err)
		assert.Equal(t, 1, m.Size())
	})

	t.Run("Contains", func(t *testing.T) {
		m := NewMockPathCache()
		err := m.Set("key1", NewPathBuilder("/tmp/v1"))
		require.NoError(t, err)
		assert.True(t, m.Contains("key1"))
	})

	t.Run("Expire", func(t *testing.T) {
		m := NewMockPathCache()
		err := m.Set("key1", NewPathBuilder("/tmp/v1"))
		require.NoError(t, err)
		// MockPathCache.Expire returns int not error
		expired := m.Expire()
		assert.Equal(t, 0, expired)
	})

	t.Run("SetMaxSize", func(t *testing.T) {
		m := NewMockPathCache()
		m.SetMaxSize(100)
	})

	t.Run("SetTTL", func(t *testing.T) {
		m := NewMockPathCache()
		m.SetTTL(60)
	})

	t.Run("SetEvictionPolicy", func(t *testing.T) {
		m := NewMockPathCache()
		m.SetEvictionPolicy(EvictLRU)
	})

	t.Run("Validate", func(t *testing.T) {
		m := NewMockPathCache()
		// Validate returns error not bool
		err := m.Validate("key1")
		assert.NoError(t, err)
	})

	t.Run("Refresh", func(t *testing.T) {
		m := NewMockPathCache()
		err := m.Set("key1", NewPathBuilder("/tmp/v1"))
		require.NoError(t, err)
		// Refresh returns error not bool
		err = m.Refresh("key1")
		assert.NoError(t, err)
	})

	t.Run("RefreshAll", func(t *testing.T) {
		m := NewMockPathCache()
		err := m.Set("key1", NewPathBuilder("/tmp/v1"))
		require.NoError(t, err)
		// RefreshAll returns error not int
		err = m.RefreshAll()
		assert.NoError(t, err)
	})
}

// TestGetDefaultWatcher tests the GetDefaultWatcher function
func TestGetDefaultWatcher(t *testing.T) {
	watcher := GetDefaultWatcher()
	assert.NotNil(t, watcher)
}

// TestDefaultOptionsManagement tests the default options functions
func TestDefaultOptionsManagement(t *testing.T) {
	t.Run("GetDefaultOptions", func(t *testing.T) {
		options := GetDefaultOptions()
		// PathOptions is a value type, not a pointer
		_ = options
	})

	t.Run("SetDefaultOptions", func(t *testing.T) {
		original := GetDefaultOptions()
		defer SetDefaultOptions(original)

		newOptions := PathOptions{
			CreateMode:    0o755,
			CreateParents: true,
			BufferSize:    16384,
		}
		SetDefaultOptions(newOptions)

		current := GetDefaultOptions()
		assert.Equal(t, newOptions.CreateParents, current.CreateParents)
		assert.Equal(t, newOptions.BufferSize, current.BufferSize)
	})
}
