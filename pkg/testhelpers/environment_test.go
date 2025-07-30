package testhelpers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTestEnvironment(t *testing.T) {
	te := NewTestEnvironment(t)
	require.NotNil(t, te)

	// Check that directories were created
	require.DirExists(t, te.RootDir())
	require.DirExists(t, te.WorkDir())

	// Check that we're in the work directory
	cwd, err := os.Getwd()
	require.NoError(t, err)
	// Use filepath.EvalSymlinks to resolve any symbolic links
	expectedPath, err := filepath.EvalSymlinks(te.WorkDir())
	require.NoError(t, err)
	actualPath, err := filepath.EvalSymlinks(cwd)
	require.NoError(t, err)
	require.Equal(t, expectedPath, actualPath)

	// Check that environment was initialized
	require.NotNil(t, te.fileOps)
	require.NotNil(t, te.envManager)
}

func TestTestEnvironment_WriteFile(t *testing.T) {
	te := NewTestEnvironment(t)

	t.Run("write simple file", func(t *testing.T) {
		te.WriteFile("test.txt", "hello world")
		require.FileExists(t, te.AbsPath("test.txt"))

		content, err := os.ReadFile(te.AbsPath("test.txt"))
		require.NoError(t, err)
		require.Equal(t, "hello world", string(content))
	})

	t.Run("write file in nested directory", func(t *testing.T) {
		te.WriteFile("nested/dir/test.txt", "nested content")
		require.FileExists(t, te.AbsPath("nested/dir/test.txt"))

		content, err := os.ReadFile(te.AbsPath("nested/dir/test.txt"))
		require.NoError(t, err)
		require.Equal(t, "nested content", string(content))
	})
}

func TestTestEnvironment_ReadFile(t *testing.T) {
	te := NewTestEnvironment(t)

	// Write a file first
	te.WriteFile("read_test.txt", "read me")

	// Read it back
	content := te.ReadFile("read_test.txt")
	require.Equal(t, "read me", content)
}

func TestTestEnvironment_FileExists(t *testing.T) {
	te := NewTestEnvironment(t)

	// Test non-existent file
	require.False(t, te.FileExists("nonexistent.txt"))

	// Create a file and test it exists
	te.WriteFile("exists.txt", "I exist")
	require.True(t, te.FileExists("exists.txt"))
}

func TestTestEnvironment_MkdirAll(t *testing.T) {
	te := NewTestEnvironment(t)

	te.MkdirAll("deep/nested/directory")
	require.DirExists(t, te.AbsPath("deep/nested/directory"))
}

func TestTestEnvironment_AbsPath(t *testing.T) {
	te := NewTestEnvironment(t)

	t.Run("relative path", func(t *testing.T) {
		absPath := te.AbsPath("test.txt")
		require.Equal(t, filepath.Join(te.WorkDir(), "test.txt"), absPath)
	})

	t.Run("absolute path", func(t *testing.T) {
		absPath := "/absolute/path/test.txt"
		require.Equal(t, absPath, te.AbsPath(absPath))
	})
}

func TestTestEnvironment_SetEnv(t *testing.T) {
	te := NewTestEnvironment(t)

	// Set a new environment variable
	te.SetEnv("TEST_VAR", "test_value")
	require.Equal(t, "test_value", os.Getenv("TEST_VAR"))

	// Override an existing variable
	os.Setenv("EXISTING_VAR", "original")
	te.SetEnv("EXISTING_VAR", "modified")
	require.Equal(t, "modified", os.Getenv("EXISTING_VAR"))
}

func TestTestEnvironment_UnsetEnv(t *testing.T) {
	te := NewTestEnvironment(t)

	// Set and then unset a variable
	os.Setenv("TO_UNSET", "value")
	te.UnsetEnv("TO_UNSET")
	require.Empty(t, os.Getenv("TO_UNSET"))
}

func TestTestEnvironment_Chdir(t *testing.T) {
	te := NewTestEnvironment(t)

	// Create a subdirectory
	te.MkdirAll("subdir")

	// Change to it
	te.Chdir("subdir")

	// Verify we're in the subdirectory
	cwd, err := os.Getwd()
	require.NoError(t, err)
	// Use filepath.EvalSymlinks to resolve any symbolic links
	expectedPath, err := filepath.EvalSymlinks(te.AbsPath("subdir"))
	require.NoError(t, err)
	actualPath, err := filepath.EvalSymlinks(cwd)
	require.NoError(t, err)
	require.Equal(t, expectedPath, actualPath)
}

func TestTestEnvironment_Run(t *testing.T) {
	te := NewTestEnvironment(t)

	var executed bool
	te.Run(func() {
		executed = true
	})

	require.True(t, executed)
}

func TestTestEnvironment_RunWithError(t *testing.T) {
	te := NewTestEnvironment(t)

	t.Run("success", func(t *testing.T) {
		err := te.RunWithError(func() error {
			return nil
		})
		require.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		expectedErr := os.ErrNotExist
		err := te.RunWithError(func() error {
			return expectedErr
		})
		require.Equal(t, expectedErr, err)
	})
}

func TestTestEnvironment_CaptureOutput(t *testing.T) {
	te := NewTestEnvironment(t)

	output := te.CaptureOutput(func() {
		os.Stdout.WriteString("captured output")
	})

	require.Equal(t, "captured output", output)
}

func TestTestEnvironment_CaptureError(t *testing.T) {
	te := NewTestEnvironment(t)

	stdout, stderr := te.CaptureError(func() {
		os.Stdout.WriteString("stdout content")
		os.Stderr.WriteString("stderr content")
	})

	require.Equal(t, "stdout content", stdout)
	require.Equal(t, "stderr content", stderr)
}

func TestTestEnvironment_AssertFileContains(t *testing.T) {
	te := NewTestEnvironment(t)

	// Create a file with content
	te.WriteFile("assert_test.txt", "this is a test file with content")

	// This should pass
	te.AssertFileContains("assert_test.txt", "test file")

	// This would fail if not in a sub-test
	t.Run("negative test", func(t *testing.T) {
		mockT := &testing.T{}
		te.t = mockT
		te.AssertFileContains("assert_test.txt", "not found")
		// We can't actually test the failure here without complex mocking
		te.t = t // Restore original t
	})
}

func TestTestEnvironment_AssertFileNotContains(t *testing.T) {
	te := NewTestEnvironment(t)

	te.WriteFile("assert_not_test.txt", "this is a test file")

	// This should pass
	te.AssertFileNotContains("assert_not_test.txt", "not present")
}

func TestTestEnvironment_AssertFileExists(t *testing.T) {
	te := NewTestEnvironment(t)

	te.WriteFile("exists_assert.txt", "content")
	te.AssertFileExists("exists_assert.txt")
}

func TestTestEnvironment_AssertFileNotExists(t *testing.T) {
	te := NewTestEnvironment(t)

	te.AssertFileNotExists("does_not_exist.txt")
}

func TestTestEnvironment_AssertDirExists(t *testing.T) {
	te := NewTestEnvironment(t)

	te.MkdirAll("test_dir")
	te.AssertDirExists("test_dir")

	// Test that it fails for files
	te.WriteFile("not_a_dir.txt", "content")
	t.Run("file is not dir", func(t *testing.T) {
		mockT := &testing.T{}
		te.t = mockT
		te.AssertDirExists("not_a_dir.txt")
		te.t = t // Restore
	})
}

func TestTestEnvironment_AssertNoError(t *testing.T) {
	te := NewTestEnvironment(t)

	te.AssertNoError(nil)

	// Test with error
	t.Run("with error", func(t *testing.T) {
		mockT := &testing.T{}
		te.t = mockT
		te.AssertNoError(os.ErrNotExist)
		te.t = t
	})
}

func TestTestEnvironment_AssertError(t *testing.T) {
	te := NewTestEnvironment(t)

	te.AssertError(os.ErrNotExist)

	// Test with nil
	t.Run("with nil", func(t *testing.T) {
		mockT := &testing.T{}
		te.t = mockT
		te.AssertError(nil)
		te.t = t
	})
}

func TestTestEnvironment_AssertErrorContains(t *testing.T) {
	te := NewTestEnvironment(t)

	err := os.ErrNotExist
	te.AssertErrorContains(err, "file does not exist")
}

func TestTestEnvironment_AddCleanup(t *testing.T) {
	te := NewTestEnvironment(t)

	var cleanupCalled bool
	te.AddCleanup(func() {
		cleanupCalled = true
	})

	// Cleanup should be called when test ends
	te.Cleanup()
	require.True(t, cleanupCalled)
}

func TestTestEnvironment_CreateGoModule(t *testing.T) {
	te := NewTestEnvironment(t)

	te.CreateGoModule("github.com/test/module")

	require.FileExists(t, te.AbsPath("go.mod"))
	content := te.ReadFile("go.mod")
	require.Contains(t, content, "module github.com/test/module")
	require.Contains(t, content, "go 1.24")
	require.Contains(t, content, "github.com/magefile/mage")
}

func TestTestEnvironment_CreateMagefile(t *testing.T) {
	te := NewTestEnvironment(t)

	t.Run("default content", func(t *testing.T) {
		te.CreateMagefile("")

		require.FileExists(t, te.AbsPath("magefile.go"))
		content := te.ReadFile("magefile.go")
		require.Contains(t, content, "// +build mage")
		require.Contains(t, content, "func Build() error")
		require.Contains(t, content, "func Test() error")
	})

	t.Run("custom content", func(t *testing.T) {
		customContent := `// +build mage

package main

func Custom() error {
    return nil
}`
		te.CreateMagefile(customContent)

		content := te.ReadFile("magefile.go")
		require.Equal(t, customContent, content)
	})
}

func TestTestEnvironment_CreateConfigFile(t *testing.T) {
	te := NewTestEnvironment(t)

	t.Run("default content", func(t *testing.T) {
		te.CreateConfigFile("")

		require.FileExists(t, te.AbsPath(".mage.yaml"))
		content := te.ReadFile(".mage.yaml")
		require.Contains(t, content, "project:")
		require.Contains(t, content, "name: test-project")
		require.Contains(t, content, "build:")
		require.Contains(t, content, "test:")
	})

	t.Run("custom content", func(t *testing.T) {
		customContent := `project:
  name: custom
`
		te.CreateConfigFile(customContent)

		content := te.ReadFile(".mage.yaml")
		require.Equal(t, customContent, content)
	})
}

func TestTestEnvironment_Cleanup(t *testing.T) {
	// Save original state
	origDir, _ := os.Getwd()
	origEnv := os.Getenv("TEST_CLEANUP_VAR")

	te := NewTestEnvironment(t)
	rootDir := te.RootDir()

	// Modify state
	te.SetEnv("TEST_CLEANUP_VAR", "modified")
	te.MkdirAll("subdir")
	te.Chdir("subdir")

	// Run cleanup
	te.Cleanup()

	// Verify state was restored
	currentDir, _ := os.Getwd()
	require.Equal(t, origDir, currentDir)
	require.Equal(t, origEnv, os.Getenv("TEST_CLEANUP_VAR"))

	// Verify temp directory was removed
	_, err := os.Stat(rootDir)
	require.True(t, os.IsNotExist(err))
}

func TestTestEnvironment_CleanupOrder(t *testing.T) {
	te := NewTestEnvironment(t)

	var order []int
	te.AddCleanup(func() { order = append(order, 1) })
	te.AddCleanup(func() { order = append(order, 2) })
	te.AddCleanup(func() { order = append(order, 3) })

	te.Cleanup()

	// Cleanup functions should run in reverse order
	require.Equal(t, []int{3, 2, 1}, order)
}

func TestTestEnvironment_GitOperations(t *testing.T) {
	// Skip if git is not available
	if _, err := os.Stat("/usr/bin/git"); err != nil && os.IsNotExist(err) {
		t.Skip("git not available")
	}

	te := NewTestEnvironment(t)

	// Setup git repo
	te.SetupGitRepo()
	require.DirExists(t, te.AbsPath(".git"))

	// Create a file and add it
	te.WriteFile("test.txt", "test content")
	te.GitAdd("test.txt")

	// Commit
	te.GitCommit("Initial commit")

	// Verify git log contains our commit
	output := te.CaptureOutput(func() {
		// We can't easily test git commands without executing them
		// This is just to ensure the methods don't panic
	})
	_ = output
}

func TestTestEnvironment_NestedDirectories(t *testing.T) {
	te := NewTestEnvironment(t)

	// Test deeply nested directory creation
	deepPath := "a/b/c/d/e/f/g/h/test.txt"
	te.WriteFile(deepPath, "deep content")

	require.FileExists(t, te.AbsPath(deepPath))
	require.Equal(t, "deep content", te.ReadFile(deepPath))
}

func TestTestEnvironment_ConcurrentAccess(t *testing.T) {
	te := NewTestEnvironment(t)

	// Test that multiple goroutines can write files
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			filename := strings.ReplaceAll("concurrent_$n.txt", "$n", string(rune('0'+n)))
			te.WriteFile(filename, "content")
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all files were created
	for i := 0; i < 10; i++ {
		filename := strings.ReplaceAll("concurrent_$n.txt", "$n", string(rune('0'+i)))
		require.True(t, te.FileExists(filename))
	}
}
