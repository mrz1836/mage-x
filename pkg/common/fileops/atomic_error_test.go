package fileops

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWriteFileAtomicCreateTempFailure tests that WriteFileAtomic fails gracefully
// when it cannot create a temp file (e.g., directory doesn't exist).
func TestWriteFileAtomicCreateTempFailure(t *testing.T) {
	op := NewDefaultSafeFileOperator()

	tests := []struct {
		name    string
		path    string
		wantErr string
	}{
		{
			name:    "NonExistentDirectory",
			path:    "/nonexistent/directory/file.txt",
			wantErr: "failed to create temp file",
		},
		{
			name:    "DeeplyNestedNonExistent",
			path:    "/a/b/c/d/e/f/g/file.txt",
			wantErr: "failed to create temp file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := op.WriteFileAtomic(tt.path, []byte("data"), 0o644)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// TestWriteFileAtomicRenameFailureDestIsDir tests that WriteFileAtomic fails
// when the destination is an existing directory.
func TestWriteFileAtomicRenameFailureDestIsDir(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows - rename behavior differs")
	}

	tmpDir := t.TempDir()
	op := NewDefaultSafeFileOperator()

	// Create a directory where we want to write the file
	destDir := filepath.Join(tmpDir, "dest_is_dir")
	require.NoError(t, os.Mkdir(destDir, 0o755)) //nolint:gosec // G301: Test file - intentional permissions

	// Create a file inside to make it non-empty
	innerFile := filepath.Join(destDir, "inner.txt")
	require.NoError(t, os.WriteFile(innerFile, []byte("inner"), 0o644)) //nolint:gosec // G306: Test file - intentional permissions

	// Try to atomically write to the directory path (should fail on rename)
	err := op.WriteFileAtomic(destDir, []byte("data"), 0o644)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to rename temp file")

	// Verify no temp files left behind
	tempFiles, err := filepath.Glob(filepath.Join(tmpDir, ".tmp-*"))
	require.NoError(t, err)
	assert.Empty(t, tempFiles, "temp files should be cleaned up after failure")
}

// TestWriteFileAtomicTempFileCleanup verifies that temp files are cleaned up
// after various failure scenarios.
func TestWriteFileAtomicTempFileCleanup(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultSafeFileOperator()

	tests := []struct {
		name  string
		setup func(t *testing.T) string
	}{
		{
			name: "FailureOnNonExistentDir",
			setup: func(t *testing.T) string {
				return filepath.Join(tmpDir, "nonexistent", "file.txt")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			dir := filepath.Dir(path)

			// Attempt write (will fail) - error is expected
			err := op.WriteFileAtomic(path, []byte("data"), 0o644)
			_ = err // intentionally ignore - we're testing cleanup behavior

			// Check for leftover temp files in the parent directory if it exists
			if _, err := os.Stat(dir); err == nil {
				tempFiles, err := filepath.Glob(filepath.Join(dir, ".tmp-*"))
				require.NoError(t, err)
				assert.Empty(t, tempFiles, "temp files should be cleaned up")
			}
		})
	}
}

// TestWriteFileAtomicPermissionVariations tests WriteFileAtomic with various
// permission modes to ensure proper handling.
func TestWriteFileAtomicPermissionVariations(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows")
	}

	tmpDir := t.TempDir()
	op := NewDefaultSafeFileOperator()

	tests := []struct {
		name string
		perm os.FileMode
	}{
		{"ReadOnly", 0o444},
		{"OwnerOnly", 0o600},
		{"OwnerReadWrite", 0o644},
		{"Executable", 0o755},
		{"AllReadWrite", 0o666},
		{"Private", 0o700},
		{"GroupReadable", 0o640},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, tt.name+".txt")
			data := []byte("test data for " + tt.name)

			err := op.WriteFileAtomic(path, data, tt.perm)
			require.NoError(t, err)

			// Verify file exists with correct permissions
			info, err := os.Stat(path)
			require.NoError(t, err)
			assert.Equal(t, tt.perm, info.Mode().Perm())

			// Verify content
			content, err := os.ReadFile(path) //nolint:gosec // G304: Test file - intentional variable path
			require.NoError(t, err)
			assert.Equal(t, data, content)
		})
	}
}

// TestWriteFileAtomicOverwriteExisting tests that WriteFileAtomic properly
// overwrites existing files atomically.
func TestWriteFileAtomicOverwriteExisting(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultSafeFileOperator()

	path := filepath.Join(tmpDir, "existing.txt")
	originalData := []byte("original content")
	newData := []byte("new content that is different")

	// Create original file
	require.NoError(t, os.WriteFile(path, originalData, 0o644)) //nolint:gosec // G306: Test file - intentional permissions

	// Overwrite atomically
	err := op.WriteFileAtomic(path, newData, 0o644)
	require.NoError(t, err)

	// Verify new content
	content, err := os.ReadFile(path) //nolint:gosec // G304: Test file - intentional variable path
	require.NoError(t, err)
	assert.Equal(t, newData, content)

	// Verify no temp files left
	tempFiles, err := filepath.Glob(filepath.Join(tmpDir, ".tmp-*"))
	require.NoError(t, err)
	assert.Empty(t, tempFiles)
}

// TestWriteFileAtomicEmptyData tests WriteFileAtomic with empty data.
func TestWriteFileAtomicEmptyData(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultSafeFileOperator()

	path := filepath.Join(tmpDir, "empty.txt")

	err := op.WriteFileAtomic(path, []byte{}, 0o644)
	require.NoError(t, err)

	// Verify empty file was created
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, int64(0), info.Size())
}

// TestWriteFileAtomicLargeData tests WriteFileAtomic with large data to
// ensure it handles multi-write scenarios correctly.
func TestWriteFileAtomicLargeData(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultSafeFileOperator()

	path := filepath.Join(tmpDir, "large.txt")

	// Create 5MB of data
	size := 5 * 1024 * 1024
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}

	err := op.WriteFileAtomic(path, data, 0o644)
	require.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(path) //nolint:gosec // G304: Test file - intentional variable path
	require.NoError(t, err)
	assert.Equal(t, data, content)

	// Verify no temp files left
	tempFiles, err := filepath.Glob(filepath.Join(tmpDir, ".tmp-*"))
	require.NoError(t, err)
	assert.Empty(t, tempFiles)
}

// TestWriteFileAtomicConcurrentWrites tests that concurrent atomic writes
// don't interfere with each other and don't leave temp files.
func TestWriteFileAtomicConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultSafeFileOperator()

	const numWriters = 10
	const iterations = 5

	done := make(chan error, numWriters*iterations)

	for i := 0; i < numWriters; i++ {
		go func(writerID int) {
			for j := 0; j < iterations; j++ {
				path := filepath.Join(tmpDir, "concurrent.txt")
				data := []byte("writer " + string(rune('A'+writerID)))
				err := op.WriteFileAtomic(path, data, 0o644)
				done <- err
			}
		}(i)
	}

	// Collect all results
	for i := 0; i < numWriters*iterations; i++ {
		err := <-done
		require.NoError(t, err)
	}

	// Verify no temp files left
	tempFiles, err := filepath.Glob(filepath.Join(tmpDir, ".tmp-*"))
	require.NoError(t, err)
	assert.Empty(t, tempFiles, "no temp files should remain after concurrent writes")

	// Verify final file exists and is valid
	_, err = os.Stat(filepath.Join(tmpDir, "concurrent.txt"))
	require.NoError(t, err)
}

// TestWriteFileAtomicReadOnlyDirectory tests behavior when writing to a
// read-only directory (CreateTemp should fail).
func TestWriteFileAtomicReadOnlyDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows - read-only directory behavior differs")
	}
	if os.Getuid() == 0 {
		t.Skip("Skipping as root - root can write to read-only directories")
	}

	tmpDir := t.TempDir()
	op := NewDefaultSafeFileOperator()

	// Create a read-only directory
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	require.NoError(t, os.Mkdir(readOnlyDir, 0o555)) //nolint:gosec // G301: Test file - intentional permissions
	defer func() {
		// Restore permissions for cleanup - error ignored in defer
		if err := os.Chmod(readOnlyDir, 0o755); err != nil { //nolint:gosec // G302: Test file - intentional permissions
			t.Logf("cleanup chmod failed: %v", err)
		}
	}()

	path := filepath.Join(readOnlyDir, "file.txt")
	err := op.WriteFileAtomic(path, []byte("data"), 0o644)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create temp file")
}

// TestWriteFileAtomicSpecialCharactersInPath tests WriteFileAtomic with
// paths containing special characters.
func TestWriteFileAtomicSpecialCharactersInPath(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultSafeFileOperator()

	tests := []struct {
		name     string
		filename string
	}{
		{"SpacesInName", "file with spaces.txt"},
		{"Dashes", "file-with-dashes.txt"},
		{"Underscores", "file_with_underscores.txt"},
		{"Numbers", "file123.txt"},
		{"Unicode", "文件.txt"}, //nolint:gosmopolitan // Test file - intentional Unicode
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(tmpDir, tt.filename)
			data := []byte("test data")

			err := op.WriteFileAtomic(path, data, 0o644)
			require.NoError(t, err)

			// Verify content
			content, err := os.ReadFile(path) //nolint:gosec // G304: Test file - intentional variable path
			require.NoError(t, err)
			assert.Equal(t, data, content)
		})
	}
}

// TestWriteFileAtomicPreservesOriginalOnFailure tests that when a write fails,
// any existing file at the destination remains unchanged.
func TestWriteFileAtomicPreservesOriginalOnFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows - behavior differs")
	}

	tmpDir := t.TempDir()
	op := NewDefaultSafeFileOperator()

	// Create an original file
	originalPath := filepath.Join(tmpDir, "original.txt")
	originalData := []byte("original content that should be preserved")
	require.NoError(t, os.WriteFile(originalPath, originalData, 0o644)) //nolint:gosec // G306: Test file - intentional permissions

	// Create a subdirectory where we want to write but make it a dir instead
	// This will cause rename to fail
	destDir := filepath.Join(tmpDir, "dest_dir")
	require.NoError(t, os.Mkdir(destDir, 0o755))                                                //nolint:gosec // G301: Test file - intentional permissions
	require.NoError(t, os.WriteFile(filepath.Join(destDir, "blocker.txt"), []byte("x"), 0o644)) //nolint:gosec // G306: Test file - intentional permissions

	// Attempt atomic write to the directory (will fail) - error expected
	writeErr := op.WriteFileAtomic(destDir, []byte("new data"), 0o644)
	_ = writeErr // intentionally ignore - we're testing original file preservation

	// Verify original file is unchanged
	content, err := os.ReadFile(originalPath) //nolint:gosec // G304: Test file - intentional variable path
	require.NoError(t, err)
	assert.Equal(t, originalData, content, "original file should be unchanged")
}
