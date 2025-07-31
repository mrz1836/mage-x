package testhelpers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTempWorkspace(t *testing.T) {
	t.Run("with name", func(t *testing.T) {
		tw := NewTempWorkspace(t, "test-workspace")
		require.NotNil(t, tw)
		require.DirExists(t, tw.Root())
		require.Contains(t, tw.Root(), "mage-test-workspace-")
	})

	t.Run("without name", func(t *testing.T) {
		tw := NewTempWorkspace(t, "")
		require.NotNil(t, tw)
		require.DirExists(t, tw.Root())
		require.Contains(t, tw.Root(), "mage-workspace-")
	})
}

func TestTempWorkspace_Dir(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	// Create new directory
	dir1 := tw.Dir("subdir1")
	require.DirExists(t, dir1)
	require.Equal(t, filepath.Join(tw.Root(), "subdir1"), dir1)

	// Get existing directory
	dir2 := tw.Dir("subdir1")
	require.Equal(t, dir1, dir2)

	// Create nested directory
	dir3 := tw.Dir("nested/deep/dir")
	require.DirExists(t, dir3)
}

func TestTempWorkspace_Path(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	// Single part
	path1 := tw.Path("file.txt")
	require.Equal(t, filepath.Join(tw.Root(), "file.txt"), path1)

	// Multiple parts
	path2 := tw.Path("dir", "subdir", "file.txt")
	require.Equal(t, filepath.Join(tw.Root(), "dir", "subdir", "file.txt"), path2)

	// No parts
	path3 := tw.Path()
	require.Equal(t, tw.Root(), path3)
}

func TestTempWorkspace_WriteFile(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	// Write simple file
	path1 := tw.WriteFile("test.txt", []byte("content"))
	require.FileExists(t, path1)
	content, err := os.ReadFile(path1) // #nosec G304 -- controlled test workspace path
	require.NoError(t, err)
	require.Equal(t, "content", string(content))

	// Write file in nested directory
	path2 := tw.WriteFile("nested/dir/file.txt", []byte("nested"))
	require.FileExists(t, path2)
	require.DirExists(t, filepath.Dir(path2))
}

func TestTempWorkspace_WriteTextFile(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	path := tw.WriteTextFile("text.txt", "text content")
	require.FileExists(t, path)
	content := tw.ReadTextFile("text.txt")
	require.Equal(t, "text content", content)
}

func TestTempWorkspace_ReadFile(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	// Write then read
	tw.WriteFile("read.txt", []byte("read me"))
	content := tw.ReadFile("read.txt")
	require.Equal(t, "read me", string(content))
}

func TestTempWorkspace_ReadTextFile(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	tw.WriteTextFile("text.txt", "text content")
	content := tw.ReadTextFile("text.txt")
	require.Equal(t, "text content", content)
}

func TestTempWorkspace_CopyFile(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	// Create source file
	tw.WriteTextFile("source.txt", "source content")

	// Copy file
	tw.CopyFile("source.txt", "dest.txt")

	// Verify destination
	require.True(t, tw.Exists("dest.txt"))
	content := tw.ReadTextFile("dest.txt")
	require.Equal(t, "source content", content)

	// Copy to nested directory
	tw.CopyFile("source.txt", "nested/copy.txt")
	require.True(t, tw.Exists("nested/copy.txt"))
}

func TestTempWorkspace_CopyDir(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	// Create source directory structure
	tw.WriteTextFile("src/file1.txt", "file1")
	tw.WriteTextFile("src/file2.txt", "file2")
	tw.WriteTextFile("src/sub/file3.txt", "file3")

	// Copy directory
	tw.CopyDir("src", "dst")

	// Verify structure
	require.True(t, tw.IsDir("dst"))
	require.Equal(t, "file1", tw.ReadTextFile("dst/file1.txt"))
	require.Equal(t, "file2", tw.ReadTextFile("dst/file2.txt"))
	require.Equal(t, "file3", tw.ReadTextFile("dst/sub/file3.txt"))
}

func TestTempWorkspace_Move(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	// Create and move file
	tw.WriteTextFile("old.txt", "content")
	tw.Move("old.txt", "new.txt")

	require.False(t, tw.Exists("old.txt"))
	require.True(t, tw.Exists("new.txt"))
	require.Equal(t, "content", tw.ReadTextFile("new.txt"))

	// Move to nested directory
	tw.Move("new.txt", "nested/moved.txt")
	require.False(t, tw.Exists("new.txt"))
	require.True(t, tw.Exists("nested/moved.txt"))
}

func TestTempWorkspace_Remove(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	// Remove file
	tw.WriteTextFile("remove.txt", "content")
	require.True(t, tw.Exists("remove.txt"))
	tw.Remove("remove.txt")
	require.False(t, tw.Exists("remove.txt"))

	// Remove directory
	tw.WriteTextFile("dir/file.txt", "content")
	require.True(t, tw.IsDir("dir"))
	tw.Remove("dir")
	require.False(t, tw.Exists("dir"))
}

func TestTempWorkspace_Exists(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	require.False(t, tw.Exists("nonexistent.txt"))

	tw.WriteTextFile("exists.txt", "content")
	require.True(t, tw.Exists("exists.txt"))

	tw.Dir("exists_dir")
	require.True(t, tw.Exists("exists_dir"))
}

func TestTempWorkspace_IsDir(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	tw.Dir("directory")
	require.True(t, tw.IsDir("directory"))

	tw.WriteTextFile("file.txt", "content")
	require.False(t, tw.IsDir("file.txt"))

	require.False(t, tw.IsDir("nonexistent"))
}

func TestTempWorkspace_IsFile(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	tw.WriteTextFile("file.txt", "content")
	require.True(t, tw.IsFile("file.txt"))

	tw.Dir("directory")
	require.False(t, tw.IsFile("directory"))

	require.False(t, tw.IsFile("nonexistent"))
}

func TestTempWorkspace_Chmod(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	tw.WriteTextFile("file.txt", "content")
	tw.Chmod("file.txt", 0o755)

	info, err := os.Stat(tw.Path("file.txt"))
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o755), info.Mode().Perm())
}

func TestTempWorkspace_Symlink(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	// Create target file
	tw.WriteTextFile("target.txt", "target content")

	// Create symlink
	tw.Symlink("target.txt", "link.txt")

	// Verify symlink
	info, err := os.Lstat(tw.Path("link.txt"))
	require.NoError(t, err)
	require.NotEqual(t, 0, info.Mode()&os.ModeSymlink)

	// Read through symlink
	content := tw.ReadTextFile("link.txt")
	require.Equal(t, "target content", content)
}

func TestTempWorkspace_ListFiles(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	// Create files and directories
	tw.WriteTextFile("dir/file1.txt", "content1")
	tw.WriteTextFile("dir/file2.txt", "content2")
	tw.Dir("dir/subdir")

	files := tw.ListFiles("dir")
	require.Len(t, files, 2)
	require.Contains(t, files, "file1.txt")
	require.Contains(t, files, "file2.txt")
}

func TestTempWorkspace_ListDirs(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	// Create directories and files
	tw.Dir("parent/dir1")
	tw.Dir("parent/dir2")
	tw.WriteTextFile("parent/file.txt", "content")

	dirs := tw.ListDirs("parent")
	require.Len(t, dirs, 2)
	require.Contains(t, dirs, "dir1")
	require.Contains(t, dirs, "dir2")
}

func TestTempWorkspace_Walk(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	// Create structure
	tw.WriteTextFile("file1.txt", "content1")
	tw.WriteTextFile("dir/file2.txt", "content2")
	tw.WriteTextFile("dir/sub/file3.txt", "content3")

	var paths []string
	tw.Walk(func(path string, info os.FileInfo, err error) error {
		require.NoError(t, err)
		if !info.IsDir() {
			relPath, relErr := filepath.Rel(tw.Root(), path)
			require.NoError(t, relErr)
			paths = append(paths, relPath)
		}
		return nil
	})

	require.Len(t, paths, 3)
	require.Contains(t, paths, "file1.txt")
	require.Contains(t, paths, filepath.Join("dir", "file2.txt"))
	require.Contains(t, paths, filepath.Join("dir", "sub", "file3.txt"))
}

func TestTempWorkspace_AddCleanup(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	var cleanupCalled bool
	tw.AddCleanup(func() {
		cleanupCalled = true
	})

	tw.Cleanup()
	require.True(t, cleanupCalled)
}

func TestTempWorkspace_Cleanup(t *testing.T) {
	tw := NewTempWorkspace(t, "test")
	root := tw.Root()

	// Create some files
	tw.WriteTextFile("file.txt", "content")

	// Cleanup
	tw.Cleanup()

	// Verify workspace is removed
	_, err := os.Stat(root)
	require.True(t, os.IsNotExist(err))
}

func TestTempWorkspace_AssertExists(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	tw.WriteTextFile("exists.txt", "content")
	tw.AssertExists("exists.txt")

	// Test negative case with mock
	t.Run("negative", func(t *testing.T) {
		mockT := &testing.T{}
		tw.t = mockT
		tw.AssertExists("nonexistent.txt")
		tw.t = t // Restore
	})
}

func TestTempWorkspace_AssertNotExists(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	tw.AssertNotExists("nonexistent.txt")

	// Test negative case
	t.Run("negative", func(t *testing.T) {
		tw.WriteTextFile("exists.txt", "content")
		mockT := &testing.T{}
		tw.t = mockT
		tw.AssertNotExists("exists.txt")
		tw.t = t // Restore
	})
}

func TestTempWorkspace_AssertFileContains(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	tw.WriteTextFile("file.txt", "this is a test file")
	tw.AssertFileContains("file.txt", "test file")

	// Test negative case
	t.Run("negative", func(t *testing.T) {
		mockT := &testing.T{}
		tw.t = mockT
		tw.AssertFileContains("file.txt", "not found")
		tw.t = t // Restore
	})
}

func TestTempWorkspace_AssertFileEquals(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	tw.WriteTextFile("file.txt", "exact content")
	tw.AssertFileEquals("file.txt", "exact content")

	// Test negative case
	t.Run("negative", func(t *testing.T) {
		mockT := &testing.T{}
		tw.t = mockT
		tw.AssertFileEquals("file.txt", "different content")
		tw.t = t // Restore
	})
}

func TestWorkspaceBuilder(t *testing.T) {
	wb := NewWorkspaceBuilder(t)
	require.NotNil(t, wb)

	// Build workspace with files and directories
	ws := wb.
		WithFile("file1.txt", "content1").
		WithDir("mydir").
		WithFile("mydir/file2.txt", "content2").
		WithGoModule("github.com/test/module").
		WithMagefile().
		Build()

	require.NotNil(t, ws)
	require.True(t, ws.Exists("file1.txt"))
	require.True(t, ws.IsDir("mydir"))
	require.True(t, ws.Exists("mydir/file2.txt"))
	require.True(t, ws.Exists("go.mod"))
	require.True(t, ws.Exists("magefile.go"))

	// Verify content
	require.Equal(t, "content1", ws.ReadTextFile("file1.txt"))
	require.Contains(t, ws.ReadTextFile("go.mod"), "module github.com/test/module")
	require.Contains(t, ws.ReadTextFile("magefile.go"), "func Build() error")
}

func TestSandboxedWorkspace(t *testing.T) {
	sw := NewSandboxedWorkspace(t)
	require.NotNil(t, sw)

	// Write within workspace should work
	path := sw.WriteTextFile("allowed.txt", "content")
	require.FileExists(t, path)

	// Read within workspace should work
	content := string(sw.ReadFile("allowed.txt"))
	require.Equal(t, "content", content)

	// Test allowed external path
	tempFile, err := os.CreateTemp("", "external-*.txt")
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Remove(tempFile.Name()))
	}()

	sw.AllowPath(filepath.Dir(tempFile.Name()))

	// Note: The current implementation doesn't actually restrict external paths
	// This is just testing the AllowPath method works
	require.Contains(t, sw.allowedPaths, filepath.Dir(tempFile.Name()))
}

func TestSandboxedWorkspace_isAllowed(t *testing.T) {
	sw := NewSandboxedWorkspace(t)

	// Workspace paths should be allowed
	require.True(t, sw.isAllowed(sw.Path("test.txt")))

	// External paths should not be allowed by default
	require.False(t, sw.isAllowed("/external/path"))

	// Allowed external paths should be allowed
	sw.AllowPath("/allowed/path")
	require.True(t, sw.isAllowed("/allowed/path/file.txt"))
}

func TestTempWorkspace_NestedOperations(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	// Create complex nested structure
	tw.WriteTextFile("root.txt", "root")
	tw.WriteTextFile("a/b/c/deep.txt", "deep content")
	tw.WriteTextFile("a/b/mid.txt", "mid content")
	tw.WriteTextFile("a/shallow.txt", "shallow content")

	// Copy nested directory
	tw.CopyDir("a", "a_copy")
	require.Equal(t, "deep content", tw.ReadTextFile("a_copy/b/c/deep.txt"))
	require.Equal(t, "mid content", tw.ReadTextFile("a_copy/b/mid.txt"))
	require.Equal(t, "shallow content", tw.ReadTextFile("a_copy/shallow.txt"))

	// Move nested directory
	tw.Move("a_copy/b", "moved_b")
	require.False(t, tw.Exists("a_copy/b"))
	require.True(t, tw.Exists("moved_b/c/deep.txt"))
}

func TestTempWorkspace_EdgeCases(t *testing.T) {
	tw := NewTempWorkspace(t, "test")

	// Empty file
	tw.WriteFile("empty.txt", []byte{})
	require.True(t, tw.Exists("empty.txt"))
	require.Empty(t, tw.ReadFile("empty.txt"))

	// File with special characters
	tw.WriteTextFile("special-!@#$%.txt", "special")
	require.True(t, tw.Exists("special-!@#$%.txt"))

	// Very long path
	longPath := "very/deep/nested/directory/structure/that/goes/on/and/on/file.txt"
	tw.WriteTextFile(longPath, "deep")
	require.True(t, tw.Exists(longPath))
}
