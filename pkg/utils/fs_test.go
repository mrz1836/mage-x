package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	t.Parallel()

	t.Run("returns true for existing file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")

		if err := os.WriteFile(testFile, []byte("content"), 0o600); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		if !FileExists(testFile) {
			t.Error("FileExists() returned false for existing file")
		}
	})

	t.Run("returns false for non-existent file", func(t *testing.T) {
		t.Parallel()
		if FileExists("/nonexistent/path/file.txt") {
			t.Error("FileExists() returned true for non-existent file")
		}
	})

	t.Run("returns true for directory", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		// FileExists uses Exists which returns true for directories too
		if !FileExists(tmpDir) {
			t.Error("FileExists() returned false for existing directory")
		}
	})
}

func TestDirExists(t *testing.T) {
	t.Parallel()

	t.Run("returns true for existing directory", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		if !DirExists(tmpDir) {
			t.Error("DirExists() returned false for existing directory")
		}
	})

	t.Run("returns false for non-existent directory", func(t *testing.T) {
		t.Parallel()
		if DirExists("/nonexistent/path/dir") {
			t.Error("DirExists() returned true for non-existent directory")
		}
	})

	t.Run("returns false for file", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "test.txt")

		if err := os.WriteFile(testFile, []byte("content"), 0o600); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		if DirExists(testFile) {
			t.Error("DirExists() returned true for a file")
		}
	})
}

func TestEnsureDir(t *testing.T) {
	t.Parallel()

	t.Run("creates nested directory", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		nestedDir := filepath.Join(tmpDir, "a", "b", "c")

		if err := EnsureDir(nestedDir); err != nil {
			t.Fatalf("EnsureDir() error = %v", err)
		}

		if !DirExists(nestedDir) {
			t.Error("EnsureDir() did not create nested directory")
		}
	})

	t.Run("succeeds for existing directory", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()

		// Call EnsureDir on already existing directory
		if err := EnsureDir(tmpDir); err != nil {
			t.Errorf("EnsureDir() error = %v for existing directory", err)
		}
	})
}

func TestCleanDir(t *testing.T) {
	t.Parallel()

	t.Run("cleans existing directory with files", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		targetDir := filepath.Join(tmpDir, "target")

		// Create directory with files
		if err := os.MkdirAll(targetDir, 0o750); err != nil {
			t.Fatalf("failed to create target dir: %v", err)
		}
		testFile := filepath.Join(targetDir, "test.txt")
		if err := os.WriteFile(testFile, []byte("content"), 0o600); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		// Clean the directory
		if err := CleanDir(targetDir); err != nil {
			t.Fatalf("CleanDir() error = %v", err)
		}

		// Directory should exist but be empty
		if !DirExists(targetDir) {
			t.Error("CleanDir() should have recreated directory")
		}

		entries, err := os.ReadDir(targetDir)
		if err != nil {
			t.Fatalf("failed to read cleaned dir: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("CleanDir() should result in empty directory, got %d entries", len(entries))
		}
	})

	t.Run("creates directory if not exists", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		targetDir := filepath.Join(tmpDir, "newdir")

		if err := CleanDir(targetDir); err != nil {
			t.Fatalf("CleanDir() error = %v for non-existent directory", err)
		}

		if !DirExists(targetDir) {
			t.Error("CleanDir() should have created directory")
		}
	})
}

func TestCopyFile(t *testing.T) {
	t.Parallel()

	t.Run("copies file content", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		srcFile := filepath.Join(tmpDir, "src.txt")
		dstFile := filepath.Join(tmpDir, "dst.txt")
		content := []byte("test content for copy")

		if err := os.WriteFile(srcFile, content, 0o600); err != nil {
			t.Fatalf("failed to create source file: %v", err)
		}

		if err := CopyFile(srcFile, dstFile); err != nil {
			t.Fatalf("CopyFile() error = %v", err)
		}

		// Verify destination exists
		if !FileExists(dstFile) {
			t.Error("CopyFile() did not create destination file")
		}

		// Verify content matches
		dstContent, err := os.ReadFile(dstFile) // #nosec G304 -- test file in temp directory
		if err != nil {
			t.Fatalf("failed to read destination file: %v", err)
		}
		if string(dstContent) != string(content) {
			t.Errorf("CopyFile() content = %q, want %q", string(dstContent), string(content))
		}
	})

	t.Run("errors for non-existent source", func(t *testing.T) {
		t.Parallel()
		tmpDir := t.TempDir()
		srcFile := filepath.Join(tmpDir, "nonexistent.txt")
		dstFile := filepath.Join(tmpDir, "dst.txt")

		if err := CopyFile(srcFile, dstFile); err == nil {
			t.Error("CopyFile() expected error for non-existent source")
		}
	})
}

func TestFindFiles(t *testing.T) {
	// Note: These tests cannot run in parallel because they change the working directory

	t.Run("finds files matching pattern", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create test file structure
		testFiles := []string{
			filepath.Join(tmpDir, "file1.go"),
			filepath.Join(tmpDir, "file2.go"),
			filepath.Join(tmpDir, "file3.txt"),
			filepath.Join(tmpDir, "subdir", "file4.go"),
		}

		// Create subdirectory
		if err := os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0o750); err != nil {
			t.Fatalf("failed to create subdir: %v", err)
		}

		// Create files
		for _, f := range testFiles {
			if err := os.WriteFile(f, []byte("content"), 0o600); err != nil {
				t.Fatalf("failed to create %s: %v", f, err)
			}
		}

		// Change to tmpDir for test (findFiles uses "." as root)
		oldDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get current dir: %v", err)
		}
		if chErr := os.Chdir(tmpDir); chErr != nil {
			t.Fatalf("failed to change to temp dir: %v", chErr)
		}
		defer func() {
			_ = os.Chdir(oldDir) //nolint:errcheck // cleanup in defer, error is acceptable
		}()

		// Find .go files
		found, err := findFiles(".", "*.go")
		if err != nil {
			t.Fatalf("findFiles() error = %v", err)
		}

		// Should find 3 .go files
		if len(found) != 3 {
			t.Errorf("findFiles(*.go) found %d files, want 3: %v", len(found), found)
		}

		// Verify no .txt files included
		for _, f := range found {
			if filepath.Ext(f) != ".go" {
				t.Errorf("findFiles(*.go) included non-.go file: %s", f)
			}
		}
	})

	t.Run("returns empty for no matches", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create only .txt files
		testFile := filepath.Join(tmpDir, "file.txt")
		if err := os.WriteFile(testFile, []byte("content"), 0o600); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}

		oldDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("failed to get current dir: %v", err)
		}
		if chErr := os.Chdir(tmpDir); chErr != nil {
			t.Fatalf("failed to change to temp dir: %v", chErr)
		}
		defer func() {
			_ = os.Chdir(oldDir) //nolint:errcheck // cleanup in defer, error is acceptable
		}()

		found, err := findFiles(".", "*.go")
		if err != nil {
			t.Fatalf("findFiles() error = %v", err)
		}

		if len(found) != 0 {
			t.Errorf("findFiles() found %d files, want 0: %v", len(found), found)
		}
	})
}

func TestFindFilesWrapper(t *testing.T) {
	// Note: This test cannot run in parallel because it changes the working directory
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(testFile, []byte("content"), 0o600); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current dir: %v", err)
	}
	if chErr := os.Chdir(tmpDir); chErr != nil {
		t.Fatalf("failed to change to temp dir: %v", chErr)
	}
	defer func() {
		_ = os.Chdir(oldDir) //nolint:errcheck // cleanup in defer, error is acceptable
	}()

	// First argument is ignored per implementation
	found, err := FindFiles("ignored", "*.go")
	if err != nil {
		t.Fatalf("FindFiles() error = %v", err)
	}

	if len(found) != 1 {
		t.Errorf("FindFiles() found %d files, want 1", len(found))
	}
}
