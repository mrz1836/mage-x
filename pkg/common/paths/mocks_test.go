package paths

import (
	"errors"
	"io/fs"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test-specific errors for mock testing (err113 compliance)
var (
	errMockAbs       = errors.New("abs failed")
	errMockRel       = errors.New("rel failed")
	errMockWalk      = errors.New("walk failed")
	errMockList      = errors.New("list failed")
	errMockGlob      = errors.New("glob failed")
	errMockValidate  = errors.New("validation failed")
	errMockInvalid   = errors.New("invalid")
	errMockCreate    = errors.New("create failed")
	errMockMkdir     = errors.New("mkdir failed")
	errMockMkdirAll  = errors.New("mkdirall failed")
	errMockRemove    = errors.New("remove failed")
	errMockRemoveAll = errors.New("removeall failed")
	errMockCopy      = errors.New("copy failed")
	errMockMove      = errors.New("move failed")
	errMockReadlink  = errors.New("readlink failed")
	errMockSymlink   = errors.New("symlink failed")
)

// ============================================================================
// MockPathBuilder Tests
// ============================================================================

func TestNewMockPathBuilder(t *testing.T) {
	mock := NewMockPathBuilder("/test/path")

	require.NotNil(t, mock)
	assert.Equal(t, "/test/path", mock.String())
	assert.Equal(t, "/test/path", mock.Original())
}

func TestMockPathBuilder_SetAndGetMockData(t *testing.T) {
	mock := NewMockPathBuilder("/test")

	// Set mock data
	mock.SetMockData("test_key", "test_value")

	// Get mock data
	value, exists := mock.GetMockData("test_key")
	require.True(t, exists)
	assert.Equal(t, "test_value", value)

	// Non-existent key
	_, exists = mock.GetMockData("non_existent")
	assert.False(t, exists)
}

func TestMockPathBuilder_Join(t *testing.T) {
	mock := NewMockPathBuilder("/base")

	result := mock.Join("sub", "path")

	assert.Equal(t, "/base/sub/path", result.String())
	assert.Equal(t, 1, mock.GetCallCount("Join"))
}

func TestMockPathBuilder_Dir(t *testing.T) {
	mock := NewMockPathBuilder("/test/file.txt")

	result := mock.Dir()

	assert.NotNil(t, result)
	assert.Equal(t, 1, mock.GetCallCount("Dir"))
}

func TestMockPathBuilder_Base(t *testing.T) {
	mock := NewMockPathBuilder("/test/file.txt")

	// Default behavior
	assert.Equal(t, "mock_base", mock.Base())

	// With mock data
	mock.SetMockData("base", "custom_base")
	assert.Equal(t, "custom_base", mock.Base())
	assert.Equal(t, 2, mock.GetCallCount("Base"))
}

func TestMockPathBuilder_Ext(t *testing.T) {
	mock := NewMockPathBuilder("/test/file.txt")

	// Default behavior
	assert.Equal(t, ".mock", mock.Ext())

	// With mock data
	mock.SetMockData("ext", ".custom")
	assert.Equal(t, ".custom", mock.Ext())
	assert.Equal(t, 2, mock.GetCallCount("Ext"))
}

func TestMockPathBuilder_Clean(t *testing.T) {
	mock := NewMockPathBuilder("/test/./path")

	result := mock.Clean()

	assert.NotNil(t, result)
	assert.Equal(t, 1, mock.GetCallCount("Clean"))
}

func TestMockPathBuilder_Abs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := NewMockPathBuilder("relative/path")

		result, err := mock.Abs()

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, mock.GetCallCount("Abs"))
	})

	t.Run("error", func(t *testing.T) {
		mock := NewMockPathBuilder("relative/path")
		mock.SetMockData("abs_error", errMockAbs)

		result, err := mock.Abs()

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, "abs failed", err.Error())
	})
}

func TestMockPathBuilder_Append(t *testing.T) {
	mock := NewMockPathBuilder("/test")

	result := mock.Append("_suffix")

	assert.Equal(t, "/test_suffix", result.String())
	assert.Equal(t, 1, mock.GetCallCount("Append"))
}

func TestMockPathBuilder_Prepend(t *testing.T) {
	mock := NewMockPathBuilder("/test")

	result := mock.Prepend("prefix_")

	assert.Equal(t, "prefix_/test", result.String())
	assert.Equal(t, 1, mock.GetCallCount("Prepend"))
}

func TestMockPathBuilder_WithExt(t *testing.T) {
	mock := NewMockPathBuilder("/test/file")

	result := mock.WithExt(".txt")

	assert.Equal(t, "/test/file.txt", result.String())
	assert.Equal(t, 1, mock.GetCallCount("WithExt"))
}

func TestMockPathBuilder_WithoutExt(t *testing.T) {
	mock := NewMockPathBuilder("/test/file.txt")

	result := mock.WithoutExt()

	assert.NotNil(t, result)
	assert.Equal(t, 1, mock.GetCallCount("WithoutExt"))
}

func TestMockPathBuilder_WithName(t *testing.T) {
	mock := NewMockPathBuilder("/test/old.txt")

	result := mock.WithName("new.txt")

	assert.Contains(t, result.String(), "new.txt")
	assert.Equal(t, 1, mock.GetCallCount("WithName"))
}

func TestMockPathBuilder_Rel(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := NewMockPathBuilder("/test/sub/path")

		result, err := mock.Rel("/test")

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, mock.GetCallCount("Rel"))
	})

	t.Run("error", func(t *testing.T) {
		mock := NewMockPathBuilder("/test/path")
		mock.SetMockData("rel_error", errMockRel)

		result, err := mock.Rel("/base")

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestMockPathBuilder_RelTo(t *testing.T) {
	mock := NewMockPathBuilder("/test/sub/path")
	target := NewMockPathBuilder("/test")

	result, err := mock.RelTo(target)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, mock.GetCallCount("RelTo"))
}

func TestMockPathBuilder_IsAbs(t *testing.T) {
	t.Run("default_absolute", func(t *testing.T) {
		mock := NewMockPathBuilder("/absolute/path")
		assert.True(t, mock.IsAbs())
	})

	t.Run("default_relative", func(t *testing.T) {
		mock := NewMockPathBuilder("relative/path")
		assert.False(t, mock.IsAbs())
	})

	t.Run("with_mock_data", func(t *testing.T) {
		mock := NewMockPathBuilder("/path")
		mock.SetMockData("is_abs", false)
		assert.False(t, mock.IsAbs())
	})
}

func TestMockPathBuilder_IsDir(t *testing.T) {
	mock := NewMockPathBuilder("/test/dir")

	// Default is false
	assert.False(t, mock.IsDir())

	// With mock data
	mock.SetMockData("is_dir", true)
	assert.True(t, mock.IsDir())
	assert.Equal(t, 2, mock.GetCallCount("IsDir"))
}

func TestMockPathBuilder_IsFile(t *testing.T) {
	mock := NewMockPathBuilder("/test/file.txt")

	// Default is true
	assert.True(t, mock.IsFile())

	// With mock data
	mock.SetMockData("is_file", false)
	assert.False(t, mock.IsFile())
	assert.Equal(t, 2, mock.GetCallCount("IsFile"))
}

func TestMockPathBuilder_Exists(t *testing.T) {
	mock := NewMockPathBuilder("/test/path")

	// Default is false
	assert.False(t, mock.Exists())

	// With mock data
	mock.SetMockData("exists", true)
	assert.True(t, mock.Exists())
	assert.Equal(t, 2, mock.GetCallCount("Exists"))
}

func TestMockPathBuilder_Size(t *testing.T) {
	mock := NewMockPathBuilder("/test/file")

	// Default
	assert.Equal(t, int64(1024), mock.Size())

	// With mock data
	mock.SetMockData("size", int64(2048))
	assert.Equal(t, int64(2048), mock.Size())
	assert.Equal(t, 2, mock.GetCallCount("Size"))
}

func TestMockPathBuilder_ModTime(t *testing.T) {
	mock := NewMockPathBuilder("/test/file")

	// Default returns current time (approximately)
	result := mock.ModTime()
	assert.WithinDuration(t, time.Now(), result, time.Second)

	// With mock data
	customTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.SetMockData("mod_time", customTime)
	assert.Equal(t, customTime, mock.ModTime())
	assert.Equal(t, 2, mock.GetCallCount("ModTime"))
}

func TestMockPathBuilder_Mode(t *testing.T) {
	mock := NewMockPathBuilder("/test/file")

	// Default
	assert.Equal(t, fs.FileMode(0o644), mock.Mode())

	// With mock data
	mock.SetMockData("mode", fs.FileMode(0o755))
	assert.Equal(t, fs.FileMode(0o755), mock.Mode())
	assert.Equal(t, 2, mock.GetCallCount("Mode"))
}

func TestMockPathBuilder_Walk(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := NewMockPathBuilder("/test")
		callCount := 0

		err := mock.Walk(func(_ PathBuilder, _ fs.FileInfo, _ error) error {
			callCount++
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, 1, callCount)
		assert.Equal(t, 1, mock.GetCallCount("Walk"))
	})

	t.Run("error", func(t *testing.T) {
		mock := NewMockPathBuilder("/test")
		mock.SetMockData("walk_error", errMockWalk)

		err := mock.Walk(func(_ PathBuilder, _ fs.FileInfo, _ error) error {
			return nil
		})

		require.Error(t, err)
		assert.Equal(t, "walk failed", err.Error())
	})
}

func TestMockPathBuilder_List(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := NewMockPathBuilder("/test")

		result, err := mock.List()

		require.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, 1, mock.GetCallCount("List"))
	})

	t.Run("error", func(t *testing.T) {
		mock := NewMockPathBuilder("/test")
		mock.SetMockData("list_error", errMockList)

		result, err := mock.List()

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestMockPathBuilder_ListFiles(t *testing.T) {
	mock := NewMockPathBuilder("/test")

	result, err := mock.ListFiles()

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, mock.GetCallCount("ListFiles"))
}

func TestMockPathBuilder_ListDirs(t *testing.T) {
	mock := NewMockPathBuilder("/test")

	result, err := mock.ListDirs()

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, mock.GetCallCount("ListDirs"))
}

func TestMockPathBuilder_Glob(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := NewMockPathBuilder("/test")

		result, err := mock.Glob("*.txt")

		require.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, 1, mock.GetCallCount("Glob"))
	})

	t.Run("error", func(t *testing.T) {
		mock := NewMockPathBuilder("/test")
		mock.SetMockData("glob_error", errMockGlob)

		result, err := mock.Glob("*.txt")

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestMockPathBuilder_Validate(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := NewMockPathBuilder("/test")

		err := mock.Validate()

		require.NoError(t, err)
		assert.Equal(t, 1, mock.GetCallCount("Validate"))
	})

	t.Run("error", func(t *testing.T) {
		mock := NewMockPathBuilder("/test")
		mock.SetMockData("validate_error", errMockValidate)

		err := mock.Validate()

		require.Error(t, err)
	})
}

func TestMockPathBuilder_IsValid(t *testing.T) {
	mock := NewMockPathBuilder("/test")

	assert.True(t, mock.IsValid())
	assert.Equal(t, 1, mock.GetCallCount("IsValid"))

	// With validation error
	mock.SetMockData("validate_error", errMockInvalid)
	assert.False(t, mock.IsValid())
}

func TestMockPathBuilder_IsEmpty(t *testing.T) {
	t.Run("not_empty", func(t *testing.T) {
		mock := NewMockPathBuilder("/test")
		assert.False(t, mock.IsEmpty())
	})

	t.Run("empty", func(t *testing.T) {
		mock := NewMockPathBuilder("")
		assert.True(t, mock.IsEmpty())
	})
}

func TestMockPathBuilder_IsSafe(t *testing.T) {
	mock := NewMockPathBuilder("/test")

	// Default is true
	assert.True(t, mock.IsSafe())

	// With mock data
	mock.SetMockData("is_safe", false)
	assert.False(t, mock.IsSafe())
	assert.Equal(t, 2, mock.GetCallCount("IsSafe"))
}

func TestMockPathBuilder_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := NewMockPathBuilder("/test/file")

		err := mock.Create()

		require.NoError(t, err)
		assert.Equal(t, 1, mock.GetCallCount("Create"))
	})

	t.Run("error", func(t *testing.T) {
		mock := NewMockPathBuilder("/test/file")
		mock.SetMockData("create_error", errMockCreate)

		err := mock.Create()

		require.Error(t, err)
	})
}

func TestMockPathBuilder_CreateDir(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := NewMockPathBuilder("/test/dir")

		err := mock.CreateDir()

		require.NoError(t, err)
		assert.Equal(t, 1, mock.GetCallCount("CreateDir"))
	})

	t.Run("error", func(t *testing.T) {
		mock := NewMockPathBuilder("/test/dir")
		mock.SetMockData("create_dir_error", errMockMkdir)

		err := mock.CreateDir()

		require.Error(t, err)
	})
}

func TestMockPathBuilder_CreateDirAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := NewMockPathBuilder("/test/deep/dir")

		err := mock.CreateDirAll()

		require.NoError(t, err)
		assert.Equal(t, 1, mock.GetCallCount("CreateDirAll"))
	})

	t.Run("error", func(t *testing.T) {
		mock := NewMockPathBuilder("/test/dir")
		mock.SetMockData("create_dir_all_error", errMockMkdirAll)

		err := mock.CreateDirAll()

		require.Error(t, err)
	})
}

func TestMockPathBuilder_Remove(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := NewMockPathBuilder("/test/file")

		err := mock.Remove()

		require.NoError(t, err)
		assert.Equal(t, 1, mock.GetCallCount("Remove"))
	})

	t.Run("error", func(t *testing.T) {
		mock := NewMockPathBuilder("/test/file")
		mock.SetMockData("remove_error", errMockRemove)

		err := mock.Remove()

		require.Error(t, err)
	})
}

func TestMockPathBuilder_RemoveAll(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := NewMockPathBuilder("/test/dir")

		err := mock.RemoveAll()

		require.NoError(t, err)
		assert.Equal(t, 1, mock.GetCallCount("RemoveAll"))
	})

	t.Run("error", func(t *testing.T) {
		mock := NewMockPathBuilder("/test/dir")
		mock.SetMockData("remove_all_error", errMockRemoveAll)

		err := mock.RemoveAll()

		require.Error(t, err)
	})
}

func TestMockPathBuilder_Copy(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := NewMockPathBuilder("/src/file")
		dest := NewMockPathBuilder("/dest/file")

		err := mock.Copy(dest)

		require.NoError(t, err)
		assert.Equal(t, 1, mock.GetCallCount("Copy"))
	})

	t.Run("error", func(t *testing.T) {
		mock := NewMockPathBuilder("/src/file")
		mock.SetMockData("copy_error", errMockCopy)

		err := mock.Copy(NewMockPathBuilder("/dest"))

		require.Error(t, err)
	})
}

func TestMockPathBuilder_Move(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := NewMockPathBuilder("/src/file")
		dest := NewMockPathBuilder("/dest/file")

		err := mock.Move(dest)

		require.NoError(t, err)
		assert.Equal(t, 1, mock.GetCallCount("Move"))
	})

	t.Run("error", func(t *testing.T) {
		mock := NewMockPathBuilder("/src/file")
		mock.SetMockData("move_error", errMockMove)

		err := mock.Move(NewMockPathBuilder("/dest"))

		require.Error(t, err)
	})
}

func TestMockPathBuilder_Readlink(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := NewMockPathBuilder("/test/link")

		result, err := mock.Readlink()

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, mock.GetCallCount("Readlink"))
	})

	t.Run("error", func(t *testing.T) {
		mock := NewMockPathBuilder("/test/link")
		mock.SetMockData("readlink_error", errMockReadlink)

		result, err := mock.Readlink()

		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestMockPathBuilder_Symlink(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mock := NewMockPathBuilder("/test/link")
		target := NewMockPathBuilder("/test/target")

		err := mock.Symlink(target)

		require.NoError(t, err)
		assert.Equal(t, 1, mock.GetCallCount("Symlink"))
	})

	t.Run("error", func(t *testing.T) {
		mock := NewMockPathBuilder("/test/link")
		mock.SetMockData("symlink_error", errMockSymlink)

		err := mock.Symlink(NewMockPathBuilder("/target"))

		require.Error(t, err)
	})
}

func TestMockPathBuilder_Match(t *testing.T) {
	mock := NewMockPathBuilder("/test/file.txt")

	// Default is false
	assert.False(t, mock.Match("*.txt"))

	// With mock data
	mock.SetMockData("match_result", true)
	assert.True(t, mock.Match("*.txt"))
	assert.Equal(t, 2, mock.GetCallCount("Match"))
}

func TestMockPathBuilder_Contains(t *testing.T) {
	mock := NewMockPathBuilder("/test/path")

	assert.True(t, mock.Contains("test"))
	assert.False(t, mock.Contains(""))
	assert.Equal(t, 2, mock.GetCallCount("Contains"))
}

func TestMockPathBuilder_HasPrefix(t *testing.T) {
	mock := NewMockPathBuilder("/test/path")

	assert.True(t, mock.HasPrefix("/test"))
	assert.False(t, mock.HasPrefix(""))
	assert.Equal(t, 2, mock.GetCallCount("HasPrefix"))
}

func TestMockPathBuilder_HasSuffix(t *testing.T) {
	mock := NewMockPathBuilder("/test/path.txt")

	assert.True(t, mock.HasSuffix(".txt"))
	assert.False(t, mock.HasSuffix(""))
	assert.Equal(t, 2, mock.GetCallCount("HasSuffix"))
}

func TestMockPathBuilder_Clone(t *testing.T) {
	mock := NewMockPathBuilder("/test/path")

	clone := mock.Clone()

	assert.NotSame(t, mock, clone)
	assert.Equal(t, mock.String(), clone.String())
	assert.Equal(t, 1, mock.GetCallCount("Clone"))
}

func TestMockPathBuilder_GetCallCount(t *testing.T) {
	mock := NewMockPathBuilder("/test")

	// No calls yet
	assert.Equal(t, 0, mock.GetCallCount("Base"))

	// After calls
	_ = mock.Base()
	_ = mock.Base()
	assert.Equal(t, 2, mock.GetCallCount("Base"))
}

// ============================================================================
// MockPathMatcher Tests
// ============================================================================

func TestNewMockPathMatcher(t *testing.T) {
	mock := NewMockPathMatcher()

	require.NotNil(t, mock)
	assert.Empty(t, mock.Patterns())
}

func TestMockPathMatcher_Match(t *testing.T) {
	mock := NewMockPathMatcher()

	// Default is false
	assert.False(t, mock.Match("/test/file.txt"))

	// With mock data
	mock.mockData["match_result"] = true
	assert.True(t, mock.Match("/test/file.txt"))
	assert.Equal(t, 2, mock.GetCallCount("Match"))
}

func TestMockPathMatcher_MatchPath(t *testing.T) {
	mock := NewMockPathMatcher()
	path := NewMockPathBuilder("/test/file.txt")

	result := mock.MatchPath(path)

	assert.False(t, result)
	assert.Equal(t, 1, mock.GetCallCount("MatchPath"))
}

func TestMockPathMatcher_Compile(t *testing.T) {
	mock := NewMockPathMatcher()

	err := mock.Compile("*.txt")

	require.NoError(t, err)
	assert.Equal(t, []string{"*.txt"}, mock.Patterns())
	assert.Equal(t, 1, mock.GetCallCount("Compile"))
}

func TestMockPathMatcher_Pattern(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		mock := NewMockPathMatcher()
		assert.Empty(t, mock.Pattern())
	})

	t.Run("with_pattern", func(t *testing.T) {
		mock := NewMockPathMatcher()
		require.NoError(t, mock.Compile("*.go"))
		assert.Equal(t, "*.go", mock.Pattern())
	})
}

func TestMockPathMatcher_AddPattern(t *testing.T) {
	mock := NewMockPathMatcher()

	err := mock.AddPattern("*.txt")
	require.NoError(t, err)

	err = mock.AddPattern("*.go")
	require.NoError(t, err)

	assert.Len(t, mock.Patterns(), 2)
	assert.Equal(t, 2, mock.GetCallCount("AddPattern"))
}

func TestMockPathMatcher_RemovePattern(t *testing.T) {
	mock := NewMockPathMatcher()
	require.NoError(t, mock.AddPattern("*.txt"))
	require.NoError(t, mock.AddPattern("*.go"))

	err := mock.RemovePattern("*.txt")

	require.NoError(t, err)
	assert.Equal(t, []string{"*.go"}, mock.Patterns())
	assert.Equal(t, 1, mock.GetCallCount("RemovePattern"))
}

func TestMockPathMatcher_ClearPatterns(t *testing.T) {
	mock := NewMockPathMatcher()
	require.NoError(t, mock.AddPattern("*.txt"))
	require.NoError(t, mock.AddPattern("*.go"))

	err := mock.ClearPatterns()

	require.NoError(t, err)
	assert.Empty(t, mock.Patterns())
	assert.Equal(t, 1, mock.GetCallCount("ClearPatterns"))
}

func TestMockPathMatcher_SetCaseSensitive(t *testing.T) {
	mock := NewMockPathMatcher()

	result := mock.SetCaseSensitive(true)

	assert.Same(t, mock, result)
	assert.Equal(t, 1, mock.GetCallCount("SetCaseSensitive"))
}

func TestMockPathMatcher_SetRecursive(t *testing.T) {
	mock := NewMockPathMatcher()

	result := mock.SetRecursive(true)

	assert.Same(t, mock, result)
	assert.Equal(t, 1, mock.GetCallCount("SetRecursive"))
}

func TestMockPathMatcher_SetMaxDepth(t *testing.T) {
	mock := NewMockPathMatcher()

	result := mock.SetMaxDepth(5)

	assert.Same(t, mock, result)
	assert.Equal(t, 1, mock.GetCallCount("SetMaxDepth"))
}

func TestMockPathMatcher_MatchAny(t *testing.T) {
	mock := NewMockPathMatcher()

	result := mock.MatchAny("/path1", "/path2")

	assert.False(t, result)
	assert.Equal(t, 1, mock.GetCallCount("MatchAny"))
}

func TestMockPathMatcher_MatchAll(t *testing.T) {
	mock := NewMockPathMatcher()

	result := mock.MatchAll("/path1", "/path2")

	assert.False(t, result)
	assert.Equal(t, 1, mock.GetCallCount("MatchAll"))
}

func TestMockPathMatcher_Filter(t *testing.T) {
	mock := NewMockPathMatcher()

	result := mock.Filter([]string{"/path1", "/path2"})

	assert.Empty(t, result)
	assert.Equal(t, 1, mock.GetCallCount("Filter"))
}

func TestMockPathMatcher_FilterPaths(t *testing.T) {
	mock := NewMockPathMatcher()
	paths := []PathBuilder{
		NewMockPathBuilder("/path1"),
		NewMockPathBuilder("/path2"),
	}

	result := mock.FilterPaths(paths)

	assert.Empty(t, result)
	assert.Equal(t, 1, mock.GetCallCount("FilterPaths"))
}

// ============================================================================
// MockPathValidator Tests
// ============================================================================

func TestNewMockPathValidator(t *testing.T) {
	mock := NewMockPathValidator()

	require.NotNil(t, mock)
	assert.Empty(t, mock.Rules())
}

func TestMockPathValidator_AddRule(t *testing.T) {
	mock := NewMockPathValidator()
	rule := &AbsolutePathRule{}

	err := mock.AddRule(rule)

	require.NoError(t, err)
	assert.Len(t, mock.Rules(), 1)
	assert.Equal(t, 1, mock.GetCallCount("AddRule"))
}

func TestMockPathValidator_RemoveRule(t *testing.T) {
	mock := NewMockPathValidator()

	err := mock.RemoveRule("test_rule")

	require.NoError(t, err)
	assert.Equal(t, 1, mock.GetCallCount("RemoveRule"))
}

func TestMockPathValidator_ClearRules(t *testing.T) {
	mock := NewMockPathValidator()
	require.NoError(t, mock.AddRule(&AbsolutePathRule{}))
	require.NoError(t, mock.AddRule(&RelativePathRule{}))

	err := mock.ClearRules()

	require.NoError(t, err)
	assert.Empty(t, mock.Rules())
	assert.Equal(t, 1, mock.GetCallCount("ClearRules"))
}

func TestMockPathValidator_Validate(t *testing.T) {
	t.Run("no_errors", func(t *testing.T) {
		mock := NewMockPathValidator()

		errors := mock.Validate("/test/path")

		assert.Empty(t, errors)
		assert.Equal(t, 1, mock.GetCallCount("Validate"))
	})

	t.Run("with_errors", func(t *testing.T) {
		mock := NewMockPathValidator()
		mock.mockData["validation_errors"] = []ValidationError{
			{Rule: "test", Message: "test error"},
		}

		errors := mock.Validate("/test/path")

		assert.Len(t, errors, 1)
	})
}

func TestMockPathValidator_ValidatePath(t *testing.T) {
	mock := NewMockPathValidator()
	path := NewMockPathBuilder("/test/path")

	errors := mock.ValidatePath(path)

	assert.Empty(t, errors)
	assert.Equal(t, 1, mock.GetCallCount("ValidatePath"))
}

func TestMockPathValidator_IsValid(t *testing.T) {
	mock := NewMockPathValidator()

	assert.True(t, mock.IsValid("/test/path"))
	assert.Equal(t, 1, mock.GetCallCount("IsValid"))
}

func TestMockPathValidator_IsValidPath(t *testing.T) {
	mock := NewMockPathValidator()
	path := NewMockPathBuilder("/test/path")

	assert.True(t, mock.IsValidPath(path))
	assert.Equal(t, 1, mock.GetCallCount("IsValidPath"))
}

func TestMockPathValidator_RequireAbsolute(t *testing.T) {
	mock := NewMockPathValidator()

	result := mock.RequireAbsolute()

	assert.Same(t, mock, result)
	assert.Equal(t, 1, mock.GetCallCount("RequireAbsolute"))
}

func TestMockPathValidator_RequireRelative(t *testing.T) {
	mock := NewMockPathValidator()

	result := mock.RequireRelative()

	assert.Same(t, mock, result)
	assert.Equal(t, 1, mock.GetCallCount("RequireRelative"))
}

func TestMockPathValidator_RequireExists(t *testing.T) {
	mock := NewMockPathValidator()

	result := mock.RequireExists()

	assert.Same(t, mock, result)
	assert.Equal(t, 1, mock.GetCallCount("RequireExists"))
}

func TestMockPathValidator_RequireNotExists(t *testing.T) {
	mock := NewMockPathValidator()

	result := mock.RequireNotExists()

	assert.Same(t, mock, result)
	assert.Equal(t, 1, mock.GetCallCount("RequireNotExists"))
}

func TestMockPathValidator_RequireReadable(t *testing.T) {
	mock := NewMockPathValidator()

	result := mock.RequireReadable()

	assert.Same(t, mock, result)
	assert.Equal(t, 1, mock.GetCallCount("RequireReadable"))
}

func TestMockPathValidator_RequireWritable(t *testing.T) {
	mock := NewMockPathValidator()

	result := mock.RequireWritable()

	assert.Same(t, mock, result)
	assert.Equal(t, 1, mock.GetCallCount("RequireWritable"))
}

func TestMockPathValidator_RequireExecutable(t *testing.T) {
	mock := NewMockPathValidator()

	result := mock.RequireExecutable()

	assert.Same(t, mock, result)
	assert.Equal(t, 1, mock.GetCallCount("RequireExecutable"))
}

func TestMockPathValidator_RequireDirectory(t *testing.T) {
	mock := NewMockPathValidator()

	result := mock.RequireDirectory()

	assert.Same(t, mock, result)
	assert.Equal(t, 1, mock.GetCallCount("RequireDirectory"))
}

func TestMockPathValidator_RequireFile(t *testing.T) {
	mock := NewMockPathValidator()

	result := mock.RequireFile()

	assert.Same(t, mock, result)
	assert.Equal(t, 1, mock.GetCallCount("RequireFile"))
}

func TestMockPathValidator_RequireExtension(t *testing.T) {
	mock := NewMockPathValidator()

	result := mock.RequireExtension(".txt", ".go")

	assert.Same(t, mock, result)
	assert.Equal(t, 1, mock.GetCallCount("RequireExtension"))
}

func TestMockPathValidator_RequireMaxLength(t *testing.T) {
	mock := NewMockPathValidator()

	result := mock.RequireMaxLength(255)

	assert.Same(t, mock, result)
	assert.Equal(t, 1, mock.GetCallCount("RequireMaxLength"))
}

func TestMockPathValidator_RequirePattern(t *testing.T) {
	mock := NewMockPathValidator()

	result := mock.RequirePattern("^/test/.*")

	assert.Same(t, mock, result)
	assert.Equal(t, 1, mock.GetCallCount("RequirePattern"))
}

func TestMockPathValidator_ForbidPattern(t *testing.T) {
	mock := NewMockPathValidator()

	result := mock.ForbidPattern(".*secret.*")

	assert.Same(t, mock, result)
	assert.Equal(t, 1, mock.GetCallCount("ForbidPattern"))
}

// ============================================================================
// MockPathSet Tests
// ============================================================================

func TestNewMockPathSet(t *testing.T) {
	mock := NewMockPathSet()

	require.NotNil(t, mock)
	assert.True(t, mock.IsEmpty())
}

func TestMockPathSet_Add(t *testing.T) {
	mock := NewMockPathSet()

	// First add returns true
	assert.True(t, mock.Add("/path1"))
	assert.Equal(t, 1, mock.Size())

	// Duplicate returns false
	assert.False(t, mock.Add("/path1"))
	assert.Equal(t, 1, mock.Size())

	assert.Equal(t, 2, mock.GetCallCount("Add"))
}

func TestMockPathSet_AddPath(t *testing.T) {
	mock := NewMockPathSet()
	path := NewMockPathBuilder("/test/path")

	result := mock.AddPath(path)

	assert.True(t, result)
	assert.Equal(t, 1, mock.GetCallCount("AddPath"))
}

func TestMockPathSet_Remove(t *testing.T) {
	mock := NewMockPathSet()
	mock.Add("/path1")

	// Remove existing returns true
	assert.True(t, mock.Remove("/path1"))
	assert.Equal(t, 0, mock.Size())

	// Remove non-existent returns false
	assert.False(t, mock.Remove("/path1"))
	assert.Equal(t, 2, mock.GetCallCount("Remove"))
}

func TestMockPathSet_RemovePath(t *testing.T) {
	mock := NewMockPathSet()
	path := NewMockPathBuilder("/test/path")
	mock.AddPath(path)

	result := mock.RemovePath(path)

	assert.True(t, result)
	assert.Equal(t, 1, mock.GetCallCount("RemovePath"))
}

func TestMockPathSet_Contains(t *testing.T) {
	mock := NewMockPathSet()
	mock.Add("/path1")

	assert.True(t, mock.Contains("/path1"))
	assert.False(t, mock.Contains("/path2"))
	assert.Equal(t, 2, mock.GetCallCount("Contains"))
}

func TestMockPathSet_ContainsPath(t *testing.T) {
	mock := NewMockPathSet()
	path := NewMockPathBuilder("/test/path")
	mock.AddPath(path)

	assert.True(t, mock.ContainsPath(path))
	assert.Equal(t, 1, mock.GetCallCount("ContainsPath"))
}

func TestMockPathSet_Clear(t *testing.T) {
	mock := NewMockPathSet()
	mock.Add("/path1")
	mock.Add("/path2")

	err := mock.Clear()

	require.NoError(t, err)
	assert.True(t, mock.IsEmpty())
	assert.Equal(t, 1, mock.GetCallCount("Clear"))
}

func TestMockPathSet_Size(t *testing.T) {
	mock := NewMockPathSet()

	assert.Equal(t, 0, mock.Size())

	mock.Add("/path1")
	mock.Add("/path2")
	assert.Equal(t, 2, mock.Size())
}

func TestMockPathSet_IsEmpty(t *testing.T) {
	mock := NewMockPathSet()

	assert.True(t, mock.IsEmpty())

	mock.Add("/path")
	assert.False(t, mock.IsEmpty())
}

func TestMockPathSet_Paths(t *testing.T) {
	mock := NewMockPathSet()
	mock.Add("/path1")
	mock.Add("/path2")

	paths := mock.Paths()

	assert.Len(t, paths, 2)
	assert.Equal(t, 1, mock.GetCallCount("Paths"))
}

func TestMockPathSet_PathBuilders(t *testing.T) {
	mock := NewMockPathSet()
	mock.Add("/path1")
	mock.Add("/path2")

	builders := mock.PathBuilders()

	assert.Len(t, builders, 2)
	assert.Equal(t, 1, mock.GetCallCount("PathBuilders"))
}

func TestMockPathSet_Union(t *testing.T) {
	mock := NewMockPathSet()
	other := NewMockPathSet()

	result := mock.Union(other)

	assert.NotNil(t, result)
	assert.Equal(t, 1, mock.GetCallCount("Union"))
}

func TestMockPathSet_Intersection(t *testing.T) {
	mock := NewMockPathSet()
	other := NewMockPathSet()

	result := mock.Intersection(other)

	assert.NotNil(t, result)
	assert.Equal(t, 1, mock.GetCallCount("Intersection"))
}

func TestMockPathSet_Difference(t *testing.T) {
	mock := NewMockPathSet()
	other := NewMockPathSet()

	result := mock.Difference(other)

	assert.NotNil(t, result)
	assert.Equal(t, 1, mock.GetCallCount("Difference"))
}

func TestMockPathSet_SymmetricDifference(t *testing.T) {
	mock := NewMockPathSet()
	other := NewMockPathSet()

	result := mock.SymmetricDifference(other)

	assert.NotNil(t, result)
	assert.Equal(t, 1, mock.GetCallCount("SymmetricDifference"))
}

func TestMockPathSet_Filter(t *testing.T) {
	mock := NewMockPathSet()

	result := mock.Filter(func(s string) bool { return true })

	assert.NotNil(t, result)
	assert.Equal(t, 1, mock.GetCallCount("Filter"))
}

func TestMockPathSet_FilterPaths(t *testing.T) {
	mock := NewMockPathSet()

	result := mock.FilterPaths(func(p PathBuilder) bool { return true })

	assert.NotNil(t, result)
	assert.Equal(t, 1, mock.GetCallCount("FilterPaths"))
}

func TestMockPathSet_ForEach(t *testing.T) {
	mock := NewMockPathSet()

	err := mock.ForEach(func(s string) error { return nil })

	require.NoError(t, err)
	assert.Equal(t, 1, mock.GetCallCount("ForEach"))
}

func TestMockPathSet_ForEachPath(t *testing.T) {
	mock := NewMockPathSet()

	err := mock.ForEachPath(func(p PathBuilder) error { return nil })

	require.NoError(t, err)
	assert.Equal(t, 1, mock.GetCallCount("ForEachPath"))
}
