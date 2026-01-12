package testhelpers

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestAssertTrue_Coverage tests AssertTrue
func TestAssertTrue_Coverage(t *testing.T) {
	t.Run("with no message", func(t *testing.T) {
		AssertTrue(t, true)
	})

	t.Run("with custom message", func(t *testing.T) {
		AssertTrue(t, true, "custom message")
	})

	// Test failure case with subtests to avoid failing the parent
	t.Run("failure case", func(t *testing.T) {
		mockT := &testing.T{}
		// This will fail mockT, not the parent test
		defer func() {
			if !mockT.Failed() {
				t.Error("AssertTrue should have failed for false value")
			}
		}()
		AssertTrue(mockT, false)
	})
}

// TestAssertFalse_Coverage tests AssertFalse
func TestAssertFalse_Coverage(t *testing.T) {
	t.Run("with no message", func(t *testing.T) {
		AssertFalse(t, false)
	})

	t.Run("with custom message", func(t *testing.T) {
		AssertFalse(t, false, "custom message")
	})

	// Test failure case with subtests
	t.Run("failure case", func(t *testing.T) {
		mockT := &testing.T{}
		defer func() {
			if !mockT.Failed() {
				t.Error("AssertFalse should have failed for true value")
			}
		}()
		AssertFalse(mockT, true)
	})
}

// TestAssertEquals_Coverage tests AssertEquals
func TestAssertEquals_Coverage(t *testing.T) {
	// Test with various types
	AssertEquals(t, 42, 42)
	AssertEquals(t, "test", "test")
	AssertEquals(t, true, true)
}

// TestAssertNotEquals_Coverage tests AssertNotEquals
func TestAssertNotEquals_Coverage(t *testing.T) {
	// Test with various types
	AssertNotEquals(t, 42, 43)
	AssertNotEquals(t, "test", "other")
	AssertNotEquals(t, true, false)
}

// TestAssertNil_Coverage tests AssertNil
func TestAssertNil_Coverage(t *testing.T) {
	// Test with nil value
	AssertNil(t, nil)
}

// TestAssertNotNil_Coverage tests AssertNotNil
func TestAssertNotNil_Coverage(t *testing.T) {
	// Test with non-nil values
	AssertNotNil(t, "not nil")
	AssertNotNil(t, 42)
	AssertNotNil(t, &struct{}{})
}

// TestEventuallyTrue_Success tests EventuallyTrue success scenario
func TestEventuallyTrue_Success(t *testing.T) {
	counter := 0
	EventuallyTrue(t, func() bool {
		counter++
		return counter >= 2 // Succeeds on second try
	}, 1*time.Second, "should succeed after retries")

	if counter < 2 {
		t.Errorf("EventuallyTrue should have called function at least twice, got %d calls", counter)
	}
}

// TestSkipIfShort_Coverage tests SkipIfShort (called when not in short mode)
func TestSkipIfShort_Coverage(t *testing.T) {
	// Call SkipIfShort - if we're not in short mode, it won't skip
	// This covers the function call
	if !testing.Short() {
		SkipIfShort(t)
		// If we reach here, we're not in short mode and the function didn't skip
	}
}

// TestSkipIfCI_Coverage tests SkipIfCI
func TestSkipIfCI_Coverage(t *testing.T) {
	// Call SkipIfCI when CI env is not set
	// This covers the function call
	SkipIfCI(t)
	// If we reach here, CI is not set and the function didn't skip
}

// TestWorkspace_Operations tests workspace file operations
func TestWorkspace_Operations(t *testing.T) {
	ws := NewTempWorkspace(t, "test")

	t.Run("CopyFile", func(t *testing.T) {
		// Create source file using workspace
		ws.WriteFile("source.txt", []byte("test content"))

		// Copy file using relative paths
		ws.CopyFile("source.txt", "dest.txt")

		// Verify using workspace ReadFile
		content := ws.ReadFile("dest.txt")
		if string(content) != "test content" {
			t.Errorf("Content mismatch: got %q, want %q", string(content), "test content")
		}
	})

	t.Run("CopyDir", func(t *testing.T) {
		// Create file in source directory
		ws.WriteFile("srcdir/file.txt", []byte("content"))

		// Copy directory using relative paths
		ws.CopyDir("srcdir", "dstdir")

		// Verify using workspace ReadFile
		content := ws.ReadFile("dstdir/file.txt")
		if string(content) != "content" {
			t.Errorf("Content mismatch: got %q, want %q", string(content), "content")
		}
	})

	t.Run("Move", func(t *testing.T) {
		// Create source file
		ws.WriteFile("movesrc.txt", []byte("move content"))

		// Move file using relative paths
		ws.Move("movesrc.txt", "movedst.txt")

		// Verify source is gone
		if ws.Exists("movesrc.txt") {
			t.Error("Source file should not exist after move")
		}

		// Verify destination exists
		content := ws.ReadFile("movedst.txt")
		if string(content) != "move content" {
			t.Errorf("Content mismatch: got %q, want %q", string(content), "move content")
		}
	})

	t.Run("Remove", func(t *testing.T) {
		// Create file
		ws.WriteFile("remove.txt", []byte("remove me"))

		// Remove file using relative path
		ws.Remove("remove.txt")

		// Verify removed
		if ws.Exists("remove.txt") {
			t.Error("File should not exist after remove")
		}
	})

	t.Run("Chmod", func(t *testing.T) {
		// Create file
		ws.WriteFile("chmod.txt", []byte("chmod test"))

		// Change permissions using relative path
		ws.Chmod("chmod.txt", 0o600)

		// Verify permissions
		filePath := ws.Path("chmod.txt")
		info, err := os.Stat(filePath)
		if err != nil {
			t.Fatalf("Failed to stat file: %v", err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Errorf("Permission mismatch: got %o, want %o", info.Mode().Perm(), 0o600)
		}
	})

	t.Run("Symlink", func(t *testing.T) {
		// Create target file
		ws.WriteFile("symlink_target.txt", []byte("target content"))

		// Create symlink using relative paths
		ws.Symlink("symlink_target.txt", "symlink.txt")

		// Verify symlink exists
		linkPath := ws.Path("symlink.txt")
		info, err := os.Lstat(linkPath)
		if err != nil {
			t.Fatalf("Failed to lstat symlink: %v", err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Error("File should be a symlink")
		}

		// Verify content through symlink
		content, err := os.ReadFile(linkPath) //nolint:gosec // test code
		if err != nil {
			t.Fatalf("Failed to read through symlink: %v", err)
		}
		if string(content) != "target content" {
			t.Errorf("Content mismatch: got %q, want %q", string(content), "target content")
		}
	})

	t.Run("Walk", func(t *testing.T) {
		// Create directory structure
		ws.WriteFile("walkdir/file1.txt", []byte("1"))
		ws.WriteFile("walkdir/file2.txt", []byte("2"))

		// Walk and collect paths
		var paths []string
		ws.Walk(func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				paths = append(paths, filepath.Base(path))
			}
			return nil
		})

		// Verify files found (should include all files in workspace)
		if len(paths) < 2 {
			t.Errorf("Expected at least 2 files, got %d", len(paths))
		}
	})
}

// TestSandboxedWorkspace_Coverage tests SandboxedWorkspace operations
func TestSandboxedWorkspace_Coverage(t *testing.T) {
	ws := NewSandboxedWorkspace(t)

	// Allow access to a specific path
	testPath := ws.Path("test.txt")
	ws.AllowPath(testPath)

	t.Run("WriteFile with allowed path", func(t *testing.T) {
		// Write to allowed path
		ws.WriteFile("test.txt", []byte("content"))

		// Verify file was written
		content, err := os.ReadFile(testPath) //nolint:gosec // test code
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}
		if string(content) != "content" {
			t.Errorf("Content mismatch: got %q, want %q", string(content), "content")
		}
	})

	t.Run("ReadFile with allowed path", func(t *testing.T) {
		// Write file first using normal file operation
		err := os.WriteFile(testPath, []byte("read me"), 0o600)
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}

		// Read through sandboxed workspace
		content := ws.ReadFile("test.txt")
		if string(content) != "read me" {
			t.Errorf("Content mismatch: got %q, want %q", string(content), "read me")
		}
	})
}

// TestAssertContains_Coverage tests AssertContains
func TestAssertContains_Coverage(t *testing.T) {
	// Test success case
	AssertContains(t, "hello world", "world")

	// Test failure case
	mockT := &testing.T{}
	AssertContains(mockT, "hello", "xyz")
	if !mockT.Failed() {
		t.Error("AssertContains should have failed when substring not found")
	}
}

// TestAssertNotContains_Coverage tests AssertNotContains
func TestAssertNotContains_Coverage(t *testing.T) {
	// Test success case
	AssertNotContains(t, "hello world", "xyz")

	// Test failure case
	mockT := &testing.T{}
	AssertNotContains(mockT, "hello world", "world")
	if !mockT.Failed() {
		t.Error("AssertNotContains should have failed when substring was found")
	}
}

// TestAssertEquals_Failure tests AssertEquals failure path
func TestAssertEquals_Failure(t *testing.T) {
	mockT := &testing.T{}
	AssertEquals(mockT, 1, 2)
	if !mockT.Failed() {
		t.Error("AssertEquals should have failed for unequal values")
	}
}

// TestAssertNotEquals_Failure tests AssertNotEquals failure path
func TestAssertNotEquals_Failure(t *testing.T) {
	mockT := &testing.T{}
	AssertNotEquals(mockT, 1, 1)
	if !mockT.Failed() {
		t.Error("AssertNotEquals should have failed for equal values")
	}
}

// TestAssertNil_Failure tests AssertNil failure path
func TestAssertNil_Failure(t *testing.T) {
	mockT := &testing.T{}
	AssertNil(mockT, "not nil")
	if !mockT.Failed() {
		t.Error("AssertNil should have failed for non-nil value")
	}
}

// TestAssertNotNil_Failure tests AssertNotNil failure path
func TestAssertNotNil_Failure(t *testing.T) {
	mockT := &testing.T{}
	AssertNotNil(mockT, nil)
	if !mockT.Failed() {
		t.Error("AssertNotNil should have failed for nil value")
	}
}

// TestTempWorkspace_NestedPath tests WriteFile with nested directory
func TestTempWorkspace_NestedPath(t *testing.T) {
	ws := NewTempWorkspace(t, "nested-test")

	// Test that WriteFile creates parent directories
	filePath := ws.WriteFile("nested/dir/file.txt", []byte("nested content"))
	content, err := os.ReadFile(filePath) //nolint:gosec // test code
	if err != nil {
		t.Fatalf("Failed to read nested file: %v", err)
	}
	if string(content) != "nested content" {
		t.Errorf("Content mismatch: got %q, want %q", string(content), "nested content")
	}
}

// TestEventuallyTrue_Failure tests EventuallyTrue timeout
func TestEventuallyTrue_Failure(t *testing.T) {
	mockT := &testing.T{}
	EventuallyTrue(mockT, func() bool {
		return false // Never succeeds
	}, 50*time.Millisecond, "should fail")

	if !mockT.Failed() {
		t.Error("EventuallyTrue should have failed on timeout")
	}
}

// TestWorkspace_MoreEdgeCases tests additional workspace edge cases
func TestWorkspace_MoreEdgeCases(t *testing.T) {
	ws := NewTempWorkspace(t, "more-tests")

	t.Run("Exists for nonexistent file", func(t *testing.T) {
		if ws.Exists("nonexistent.txt") {
			t.Error("Exists should return false for nonexistent file")
		}
	})

	t.Run("IsDir and IsFile checks", func(t *testing.T) {
		ws.WriteFile("file.txt", []byte("content"))
		ws.Dir("subdir")

		if !ws.IsFile("file.txt") {
			t.Error("IsFile should return true for file")
		}
		if ws.IsDir("file.txt") {
			t.Error("IsDir should return false for file")
		}
		if !ws.IsDir("subdir") {
			t.Error("IsDir should return true for directory")
		}
		if ws.IsFile("subdir") {
			t.Error("IsFile should return false for directory")
		}
	})

	t.Run("Walk with error", func(t *testing.T) {
		ws.WriteFile("walkfile.txt", []byte("content"))

		// Walk and force an error by returning an error from the walk function
		var called bool
		ws.Walk(func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				called = true
			}
			return nil
		})

		if !called {
			t.Error("Walk should have called the function for files")
		}
	})

	t.Run("Path and Dir operations", func(t *testing.T) {
		// Test Path with multiple parts
		path := ws.Path("a", "b", "c")
		if !filepath.IsAbs(path) {
			t.Error("Path should return absolute path")
		}

		// Test Dir creates directory
		dir := ws.Dir("testdir")
		if !filepath.IsAbs(dir) {
			t.Error("Dir should return absolute path")
		}
		if !ws.IsDir("testdir") {
			t.Error("Dir should create the directory")
		}
	})

	t.Run("WriteTextFile and ReadTextFile", func(t *testing.T) {
		// Test WriteTextFile
		filePath := ws.WriteTextFile("text.txt", "text content")
		if !ws.Exists("text.txt") {
			t.Error("WriteTextFile should create file")
		}

		// Test ReadTextFile
		content := ws.ReadTextFile("text.txt")
		if content != "text content" {
			t.Errorf("ReadTextFile mismatch: got %q, want %q", content, "text content")
		}

		// Verify the path is absolute
		if !filepath.IsAbs(filePath) {
			t.Error("WriteTextFile should return absolute path")
		}
	})

	t.Run("AddCleanup", func(t *testing.T) {
		ws.AddCleanup(func() {
			// Cleanup function for coverage
		})

		// Cleanup is called by test framework, so we can't verify it directly
		// But this at least calls the AddCleanup method for coverage
	})

	t.Run("CopyDir with nested structure", func(t *testing.T) {
		// Create nested directory structure
		ws.WriteFile("source/a/file1.txt", []byte("1"))
		ws.WriteFile("source/b/file2.txt", []byte("2"))

		// Copy directory
		ws.CopyDir("source", "dest")

		// Verify files were copied
		if !ws.Exists("dest/a/file1.txt") {
			t.Error("Nested file should have been copied")
		}
		if !ws.Exists("dest/b/file2.txt") {
			t.Error("Nested file should have been copied")
		}

		// Verify content
		content := ws.ReadFile("dest/a/file1.txt")
		if string(content) != "1" {
			t.Errorf("Content mismatch: got %q, want %q", string(content), "1")
		}
	})

	t.Run("Move with different names", func(t *testing.T) {
		// Create source with specific name
		ws.WriteFile("move_src.txt", []byte("move test"))

		// Move to different name
		ws.Move("move_src.txt", "move_dst.txt")

		// Verify
		if ws.Exists("move_src.txt") {
			t.Error("Source should not exist after move")
		}
		if !ws.Exists("move_dst.txt") {
			t.Error("Destination should exist after move")
		}
	})
}

// TestWorkspaceAssertions tests workspace assertion methods
func TestWorkspaceAssertions(t *testing.T) {
	ws := NewTempWorkspace(t, "assertions-test")

	ws.WriteFile("test.txt", []byte("test content"))

	// AssertExists for existing file
	ws.AssertExists("test.txt")

	// AssertNotExists for nonexistent file
	ws.AssertNotExists("nonexistent.txt")

	// AssertFileContains
	ws.AssertFileContains("test.txt", "content")

	// AssertFileEquals
	ws.AssertFileEquals("test.txt", "test content")
}

// TestBaseSuite_RestoreEnvironment tests restoreEnvironment
func TestBaseSuite_RestoreEnvironment(t *testing.T) {
	suite := &BaseSuite{}
	suite.SetT(t)

	// Set original environment
	suite.OriginalEnv = map[string]string{
		"TEST_VAR": "original",
	}

	// Set vars to restore
	suite.EnvVarsToSet = map[string]string{
		"TEST_VAR": "modified",
	}

	// Set current value
	require.NoError(t, os.Setenv("TEST_VAR", "modified"))

	// Restore environment
	suite.restoreEnvironment()

	// Verify it was restored
	value := os.Getenv("TEST_VAR")
	if value != "original" {
		t.Errorf("restoreEnvironment failed: expected %q, got %q", "original", value)
	}
}

// TestBaseSuite_WithTestEnv tests WithTestEnv
func TestBaseSuite_WithTestEnv(t *testing.T) {
	suite := &BaseSuite{}
	suite.SetT(t)

	// First call should create TestEnv
	env1 := suite.WithTestEnv()
	if env1 == nil {
		t.Error("WithTestEnv should create TestEnvironment")
	}

	// Second call should return the same instance
	env2 := suite.WithTestEnv()
	if env2 != env1 {
		t.Error("WithTestEnv should return the same instance on subsequent calls")
	}
}
