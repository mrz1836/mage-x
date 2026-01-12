package paths

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLazyPathBuilder tests the lazyPathBuilder methods that delegate to GetDefaultBuilder
func TestLazyPathBuilder(t *testing.T) {
	t.Run("Join", func(t *testing.T) {
		result := DefaultBuilder.Join("foo", "bar", "baz")
		assert.NotNil(t, result)
		assert.Contains(t, result.String(), "bar")
	})

	t.Run("Dir", func(t *testing.T) {
		// Set default builder to a specific path first
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		SetDefaultBuilder(NewPathBuilder("/path/to/file.txt"))
		result := DefaultBuilder.Dir()
		assert.NotNil(t, result)
		assert.Equal(t, "/path/to", result.String())
	})

	t.Run("Base", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		SetDefaultBuilder(NewPathBuilder("/path/to/file.txt"))
		result := DefaultBuilder.Base()
		assert.Equal(t, "file.txt", result)
	})

	t.Run("Ext", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		SetDefaultBuilder(NewPathBuilder("/path/to/file.txt"))
		result := DefaultBuilder.Ext()
		assert.Equal(t, ".txt", result)
	})

	t.Run("Clean", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		SetDefaultBuilder(NewPathBuilder("/path/to/../file.txt"))
		result := DefaultBuilder.Clean()
		assert.NotNil(t, result)
		assert.Equal(t, "/path/file.txt", result.String())
	})

	t.Run("Abs", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		SetDefaultBuilder(NewPathBuilder("relative/path.txt"))
		result, err := DefaultBuilder.Abs()
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, filepath.IsAbs(result.String()))
	})

	t.Run("Append", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		SetDefaultBuilder(NewPathBuilder("/path/to/file"))
		result := DefaultBuilder.Append(".txt")
		assert.NotNil(t, result)
		assert.Equal(t, "/path/to/file.txt", result.String())
	})

	t.Run("Prepend", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		SetDefaultBuilder(NewPathBuilder("/path/to/file.txt"))
		result := DefaultBuilder.Prepend("prefix_")
		assert.NotNil(t, result)
		assert.Contains(t, result.String(), "prefix_")
	})

	t.Run("WithExt", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		SetDefaultBuilder(NewPathBuilder("/path/to/file.txt"))
		result := DefaultBuilder.WithExt(".md")
		assert.NotNil(t, result)
		assert.Equal(t, "/path/to/file.md", result.String())
	})

	t.Run("WithoutExt", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		SetDefaultBuilder(NewPathBuilder("/path/to/file.txt"))
		result := DefaultBuilder.WithoutExt()
		assert.NotNil(t, result)
		assert.Equal(t, "/path/to/file", result.String())
	})

	t.Run("WithName", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		SetDefaultBuilder(NewPathBuilder("/path/to/file.txt"))
		result := DefaultBuilder.WithName("newfile.txt")
		assert.NotNil(t, result)
		assert.Equal(t, "/path/to/newfile.txt", result.String())
	})
}

// TestLazyPathBuilderRel tests Rel and RelTo methods
func TestLazyPathBuilderRel(t *testing.T) {
	t.Run("Rel", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		SetDefaultBuilder(NewPathBuilder("/path/to/file.txt"))
		result, err := DefaultBuilder.Rel("/path")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "to/file.txt", result.String())
	})

	t.Run("RelTo", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		SetDefaultBuilder(NewPathBuilder("/path/to/file.txt"))
		target := NewPathBuilder("/path")
		result, err := DefaultBuilder.RelTo(target)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "to/file.txt", result.String())
	})
}

// TestLazyPathBuilderInfo tests info methods
func TestLazyPathBuilderInfo(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0o600)
	require.NoError(t, err)

	original := GetDefaultBuilder()
	defer SetDefaultBuilder(original)

	SetDefaultBuilder(NewPathBuilder(testFile))

	t.Run("String", func(t *testing.T) {
		result := DefaultBuilder.String()
		assert.Equal(t, testFile, result)
	})

	t.Run("Original", func(t *testing.T) {
		result := DefaultBuilder.Original()
		assert.Equal(t, testFile, result)
	})

	t.Run("IsAbs", func(t *testing.T) {
		result := DefaultBuilder.IsAbs()
		assert.True(t, result)
	})

	t.Run("IsDir", func(t *testing.T) {
		result := DefaultBuilder.IsDir()
		assert.False(t, result)
	})

	t.Run("IsFile", func(t *testing.T) {
		result := DefaultBuilder.IsFile()
		assert.True(t, result)
	})

	t.Run("Exists", func(t *testing.T) {
		result := DefaultBuilder.Exists()
		assert.True(t, result)
	})

	t.Run("Size", func(t *testing.T) {
		result := DefaultBuilder.Size()
		assert.Equal(t, int64(12), result) // "test content" is 12 bytes
	})

	t.Run("ModTime", func(t *testing.T) {
		result := DefaultBuilder.ModTime()
		assert.False(t, result.IsZero())
	})

	t.Run("Mode", func(t *testing.T) {
		result := DefaultBuilder.Mode()
		assert.NotZero(t, result)
	})
}

// TestLazyPathBuilderOperations tests file operations
func TestLazyPathBuilderOperations(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Walk", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		SetDefaultBuilder(NewPathBuilder(tmpDir))
		var paths []string
		err := DefaultBuilder.Walk(func(path PathBuilder, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			paths = append(paths, path.String())
			return nil
		})
		require.NoError(t, err)
		assert.NotEmpty(t, paths)
	})

	t.Run("List", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		// Create a subdirectory and file
		subDir := filepath.Join(tmpDir, "subdir")
		err := os.MkdirAll(subDir, 0o750)
		require.NoError(t, err)
		testFile := filepath.Join(tmpDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test"), 0o600)
		require.NoError(t, err)

		SetDefaultBuilder(NewPathBuilder(tmpDir))
		result, err := DefaultBuilder.List()
		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("ListFiles", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		SetDefaultBuilder(NewPathBuilder(tmpDir))
		result, err := DefaultBuilder.ListFiles()
		require.NoError(t, err)
		// May be empty or not depending on previous tests
		_ = result
	})

	t.Run("ListDirs", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		SetDefaultBuilder(NewPathBuilder(tmpDir))
		result, err := DefaultBuilder.ListDirs()
		require.NoError(t, err)
		// May be empty or not depending on previous tests
		_ = result
	})

	t.Run("Glob", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		SetDefaultBuilder(NewPathBuilder(tmpDir))
		result, err := DefaultBuilder.Glob("*.txt")
		require.NoError(t, err)
		// May be empty or not depending on previous tests
		_ = result
	})
}

// TestLazyPathBuilderValidation tests validation methods
func TestLazyPathBuilderValidation(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("test"), 0o600)
	require.NoError(t, err)

	original := GetDefaultBuilder()
	defer SetDefaultBuilder(original)

	SetDefaultBuilder(NewPathBuilder(testFile))

	t.Run("Validate", func(t *testing.T) {
		err := DefaultBuilder.Validate()
		assert.NoError(t, err)
	})

	t.Run("IsValid", func(t *testing.T) {
		result := DefaultBuilder.IsValid()
		assert.True(t, result)
	})

	t.Run("IsEmpty", func(t *testing.T) {
		result := DefaultBuilder.IsEmpty()
		assert.False(t, result)
	})

	t.Run("IsSafe", func(t *testing.T) {
		result := DefaultBuilder.IsSafe()
		assert.True(t, result)
	})
}

// TestLazyPathBuilderFileManagement tests file management methods
func TestLazyPathBuilderFileManagement(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Create", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		newFile := filepath.Join(tmpDir, "created.txt")
		SetDefaultBuilder(NewPathBuilder(newFile))
		err := DefaultBuilder.Create()
		require.NoError(t, err)
		assert.True(t, PathExists(newFile))
	})

	t.Run("CreateDir", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		newDir := filepath.Join(tmpDir, "newdir")
		SetDefaultBuilder(NewPathBuilder(newDir))
		err := DefaultBuilder.CreateDir()
		require.NoError(t, err)
		assert.True(t, PathIsDir(newDir))
	})

	t.Run("CreateDirAll", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		newDir := filepath.Join(tmpDir, "nested", "dir", "structure")
		SetDefaultBuilder(NewPathBuilder(newDir))
		err := DefaultBuilder.CreateDirAll()
		require.NoError(t, err)
		assert.True(t, PathIsDir(newDir))
	})

	t.Run("Remove", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		toRemove := filepath.Join(tmpDir, "toremove.txt")
		err := os.WriteFile(toRemove, []byte("remove me"), 0o600)
		require.NoError(t, err)

		SetDefaultBuilder(NewPathBuilder(toRemove))
		err = DefaultBuilder.Remove()
		require.NoError(t, err)
		assert.False(t, PathExists(toRemove))
	})

	t.Run("RemoveAll", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		toRemoveDir := filepath.Join(tmpDir, "toremovedir")
		err := os.MkdirAll(filepath.Join(toRemoveDir, "sub"), 0o750)
		require.NoError(t, err)

		SetDefaultBuilder(NewPathBuilder(toRemoveDir))
		err = DefaultBuilder.RemoveAll()
		require.NoError(t, err)
		assert.False(t, PathExists(toRemoveDir))
	})

	t.Run("Copy", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		src := filepath.Join(tmpDir, "src.txt")
		err := os.WriteFile(src, []byte("source content"), 0o600)
		require.NoError(t, err)

		dst := filepath.Join(tmpDir, "dst.txt")
		SetDefaultBuilder(NewPathBuilder(src))
		err = DefaultBuilder.Copy(NewPathBuilder(dst))
		require.NoError(t, err)
		assert.True(t, PathExists(dst))
	})

	t.Run("Move", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		src := filepath.Join(tmpDir, "tomove.txt")
		err := os.WriteFile(src, []byte("move me"), 0o600)
		require.NoError(t, err)

		dst := filepath.Join(tmpDir, "moved.txt")
		SetDefaultBuilder(NewPathBuilder(src))
		err = DefaultBuilder.Move(NewPathBuilder(dst))
		require.NoError(t, err)
		assert.True(t, PathExists(dst))
		assert.False(t, PathExists(src))
	})

	t.Run("Symlink", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		target := filepath.Join(tmpDir, "target.txt")
		err := os.WriteFile(target, []byte("target"), 0o600)
		require.NoError(t, err)

		link := filepath.Join(tmpDir, "link.txt")
		SetDefaultBuilder(NewPathBuilder(link))
		err = DefaultBuilder.Symlink(NewPathBuilder(target))
		require.NoError(t, err)

		linkInfo, err := os.Lstat(link)
		require.NoError(t, err)
		assert.NotEqual(t, 0, linkInfo.Mode()&os.ModeSymlink)
	})

	t.Run("Readlink", func(t *testing.T) {
		original := GetDefaultBuilder()
		defer SetDefaultBuilder(original)

		target := filepath.Join(tmpDir, "readlink_target.txt")
		err := os.WriteFile(target, []byte("target"), 0o600)
		require.NoError(t, err)

		link := filepath.Join(tmpDir, "readlink.txt")
		err = os.Symlink(target, link)
		require.NoError(t, err)

		SetDefaultBuilder(NewPathBuilder(link))
		result, err := DefaultBuilder.Readlink()
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, target, result.String())
	})
}

// TestLazyPathBuilderMatching tests pattern matching methods
func TestLazyPathBuilderMatching(t *testing.T) {
	original := GetDefaultBuilder()
	defer SetDefaultBuilder(original)

	SetDefaultBuilder(NewPathBuilder("/path/to/file.txt"))

	t.Run("Match", func(t *testing.T) {
		result := DefaultBuilder.Match("*.txt")
		assert.True(t, result)
	})

	t.Run("Contains", func(t *testing.T) {
		result := DefaultBuilder.Contains("to/file")
		assert.True(t, result)
	})

	t.Run("HasPrefix", func(t *testing.T) {
		result := DefaultBuilder.HasPrefix("/path")
		assert.True(t, result)
	})

	t.Run("HasSuffix", func(t *testing.T) {
		result := DefaultBuilder.HasSuffix(".txt")
		assert.True(t, result)
	})

	t.Run("Clone", func(t *testing.T) {
		result := DefaultBuilder.Clone()
		assert.NotNil(t, result)
		assert.Equal(t, "/path/to/file.txt", result.String())
	})
}
