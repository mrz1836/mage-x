package fileops

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCopySourceNotExist tests that Copy fails gracefully when the source doesn't exist.
func TestCopySourceNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultFileOperator()

	src := filepath.Join(tmpDir, "nonexistent.txt")
	dst := filepath.Join(tmpDir, "dest.txt")

	err := op.Copy(src, dst)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open source file")
}

// TestCopyDestinationInvalidPath tests Copy when destination path is invalid.
func TestCopyDestinationInvalidPath(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultFileOperator()

	// Create source file
	src := filepath.Join(tmpDir, "source.txt")
	require.NoError(t, os.WriteFile(src, []byte("test data"), 0o644)) //nolint:gosec // G306: Test file - intentional permissions

	// Try to copy to non-existent directory
	dst := filepath.Join(tmpDir, "nonexistent", "subdir", "dest.txt")

	err := op.Copy(src, dst)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create destination file")
}

// TestCopyPreservesPermissionsExtended tests that Copy properly preserves source permissions
// with additional permission modes.
func TestCopyPreservesPermissionsExtended(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission tests on Windows")
	}

	tmpDir := t.TempDir()
	op := NewDefaultFileOperator()

	tests := []struct {
		name string
		perm os.FileMode
	}{
		{"ReadOnly", 0o444},
		{"OwnerOnly", 0o600},
		{"OwnerReadWrite", 0o644},
		{"Executable", 0o755},
		{"AllReadWrite", 0o666},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := filepath.Join(tmpDir, tt.name+"_src.txt")
			dst := filepath.Join(tmpDir, tt.name+"_dst.txt")

			// Create source with specific permissions
			require.NoError(t, os.WriteFile(src, []byte("test"), 0o644)) //nolint:gosec // G306: Test file - intentional permissions
			require.NoError(t, os.Chmod(src, tt.perm))

			// Copy
			err := op.Copy(src, dst)
			require.NoError(t, err)

			// Verify permissions
			info, err := os.Stat(dst)
			require.NoError(t, err)
			assert.Equal(t, tt.perm, info.Mode().Perm(), "destination should have same permissions as source")
		})
	}
}

// TestCopyLargeFile tests Copy with a large file to ensure io.Copy handles it correctly.
func TestCopyLargeFile(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultFileOperator()

	src := filepath.Join(tmpDir, "large_src.txt")
	dst := filepath.Join(tmpDir, "large_dst.txt")

	// Create 10MB file
	size := 10 * 1024 * 1024
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	require.NoError(t, os.WriteFile(src, data, 0o644)) //nolint:gosec // G306: Test file - intentional permissions

	// Copy
	err := op.Copy(src, dst)
	require.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(dst) //nolint:gosec // G304: Test file - intentional variable path
	require.NoError(t, err)
	assert.Equal(t, data, content)
}

// TestCopyEmptyFile tests Copy with an empty file.
func TestCopyEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultFileOperator()

	src := filepath.Join(tmpDir, "empty_src.txt")
	dst := filepath.Join(tmpDir, "empty_dst.txt")

	// Create empty source
	require.NoError(t, os.WriteFile(src, []byte{}, 0o644)) //nolint:gosec // G306: Test file - intentional permissions

	// Copy
	err := op.Copy(src, dst)
	require.NoError(t, err)

	// Verify destination is empty
	info, err := os.Stat(dst)
	require.NoError(t, err)
	assert.Equal(t, int64(0), info.Size())
}

// TestCopyOverwriteExisting tests that Copy properly overwrites an existing destination.
func TestCopyOverwriteExisting(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultFileOperator()

	src := filepath.Join(tmpDir, "src.txt")
	dst := filepath.Join(tmpDir, "dst.txt")

	srcData := []byte("source content")
	dstData := []byte("original destination content that is longer")

	// Create source and destination
	require.NoError(t, os.WriteFile(src, srcData, 0o644)) //nolint:gosec // G306: Test file - intentional permissions
	require.NoError(t, os.WriteFile(dst, dstData, 0o644)) //nolint:gosec // G306: Test file - intentional permissions

	// Copy (should overwrite)
	err := op.Copy(src, dst)
	require.NoError(t, err)

	// Verify destination has source content
	content, err := os.ReadFile(dst) //nolint:gosec // G304: Test file - intentional variable path
	require.NoError(t, err)
	assert.Equal(t, srcData, content)
}

// TestCopyToReadOnlyDirectory tests Copy when destination directory is read-only.
func TestCopyToReadOnlyDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows - read-only directory behavior differs")
	}
	if os.Getuid() == 0 {
		t.Skip("Skipping as root - root can write to read-only directories")
	}

	tmpDir := t.TempDir()
	op := NewDefaultFileOperator()

	// Create source
	src := filepath.Join(tmpDir, "src.txt")
	require.NoError(t, os.WriteFile(src, []byte("data"), 0o644)) //nolint:gosec // G306: Test file - intentional permissions //nolint:gosec // G306: Test file - intentional permissions

	// Create read-only directory
	readOnlyDir := filepath.Join(tmpDir, "readonly")
	require.NoError(t, os.Mkdir(readOnlyDir, 0o555)) //nolint:gosec // G301: Test file - intentional permissions
	defer func() {
		// Restore permissions for cleanup - error ignored in defer
		if err := os.Chmod(readOnlyDir, 0o755); err != nil { //nolint:gosec // G302: Test file - intentional permissions
			t.Logf("cleanup chmod failed: %v", err)
		}
	}()

	dst := filepath.Join(readOnlyDir, "dst.txt")
	err := op.Copy(src, dst)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create destination file")
}

// TestCopySpecialCharactersInPath tests Copy with paths containing special characters.
func TestCopySpecialCharactersInPath(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultFileOperator()

	tests := []struct {
		name     string
		filename string
	}{
		{"SpacesInName", "file with spaces.txt"},
		{"Dashes", "file-with-dashes.txt"},
		{"Underscores", "file_with_underscores.txt"},
		{"Unicode", "文件.txt"}, //nolint:gosmopolitan // Test file - intentional Unicode
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := filepath.Join(tmpDir, "src_"+tt.filename)
			dst := filepath.Join(tmpDir, "dst_"+tt.filename)
			data := []byte("test data for " + tt.name)

			require.NoError(t, os.WriteFile(src, data, 0o644)) //nolint:gosec // G306: Test file - intentional permissions //nolint:gosec // G306: Test file - intentional permissions

			err := op.Copy(src, dst)
			require.NoError(t, err)

			content, err := os.ReadFile(dst) //nolint:gosec // G304: Test file - intentional variable path
			require.NoError(t, err)
			assert.Equal(t, data, content)
		})
	}
}

// TestCopyPathTraversalSource tests that Copy rejects path traversal in source.
func TestCopyPathTraversalSource(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultFileOperator()

	dst := filepath.Join(tmpDir, "dst.txt")

	tests := []struct {
		name string
		src  string
	}{
		{"BasicTraversal", "../../../etc/passwd"},
		{"MixedTraversal", "foo/../../../bar"},
		{"DoubleEncoded", "..%2F..%2Fetc/passwd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := op.Copy(tt.src, dst)
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrPathTraversalDetected)
		})
	}
}

// TestCopyPathTraversalDestination tests that Copy rejects path traversal in destination.
func TestCopyPathTraversalDestination(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultFileOperator()

	src := filepath.Join(tmpDir, "src.txt")
	require.NoError(t, os.WriteFile(src, []byte("data"), 0o644)) //nolint:gosec // G306: Test file - intentional permissions

	tests := []struct {
		name string
		dst  string
	}{
		{"BasicTraversal", "../../../tmp/evil.txt"},
		{"MixedTraversal", "foo/../../../bar.txt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := op.Copy(src, tt.dst)
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrPathTraversalDetected)
		})
	}
}

// TestCopyConcurrent tests that concurrent copy operations work correctly.
func TestCopyConcurrent(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultFileOperator()

	const numCopiers = 10

	// Create source files
	for i := 0; i < numCopiers; i++ {
		src := filepath.Join(tmpDir, "src"+string(rune('0'+i))+".txt")
		data := []byte("content for file " + string(rune('0'+i)))
		require.NoError(t, os.WriteFile(src, data, 0o644)) //nolint:gosec // G306: Test file - intentional permissions
	}

	done := make(chan error, numCopiers)

	// Run concurrent copies
	for i := 0; i < numCopiers; i++ {
		go func(idx int) {
			src := filepath.Join(tmpDir, "src"+string(rune('0'+idx))+".txt")
			dst := filepath.Join(tmpDir, "dst"+string(rune('0'+idx))+".txt")
			done <- op.Copy(src, dst)
		}(i)
	}

	// Collect results
	for i := 0; i < numCopiers; i++ {
		err := <-done
		require.NoError(t, err)
	}

	// Verify all destinations exist
	for i := 0; i < numCopiers; i++ {
		dstPath := filepath.Join(tmpDir, "dst"+string(rune('0'+i))+".txt")
		_, err := os.Stat(dstPath)
		require.NoError(t, err)
	}
}

// TestCopySourceIsDirectory tests that Copy fails when source is a directory.
func TestCopySourceIsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultFileOperator()

	src := filepath.Join(tmpDir, "srcdir")
	require.NoError(t, os.Mkdir(src, 0o755)) //nolint:gosec // G301: Test file - intentional permissions

	dst := filepath.Join(tmpDir, "dst.txt")

	err := op.Copy(src, dst)
	require.Error(t, err)
	// On macOS/Linux, opening a directory succeeds but reading fails with "is a directory"
	// On Windows, it may fail differently. The key is that it fails.
	assert.Contains(t, err.Error(), "is a directory")
}

// TestCopyDestinationIsExistingDirectory tests Copy when destination is an existing directory.
func TestCopyDestinationIsExistingDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultFileOperator()

	src := filepath.Join(tmpDir, "src.txt")
	require.NoError(t, os.WriteFile(src, []byte("data"), 0o644)) //nolint:gosec // G306: Test file - intentional permissions

	dst := filepath.Join(tmpDir, "dstdir")
	require.NoError(t, os.Mkdir(dst, 0o755)) //nolint:gosec // G301: Test file - intentional permissions

	err := op.Copy(src, dst)
	// Behavior varies by OS - may fail to create or truncate a directory
	// The important thing is it doesn't silently succeed
	if err == nil {
		// If it succeeded, the directory should be replaced with a file
		info, statErr := os.Stat(dst)
		require.NoError(t, statErr)
		assert.False(t, info.IsDir(), "destination should be a file, not a directory")
	}
}

// TestCopySymlinkSource tests Copy behavior with symlink as source.
func TestCopySymlinkSource(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping symlink test on Windows")
	}

	tmpDir := t.TempDir()
	op := NewDefaultFileOperator()

	// Create actual file
	actualFile := filepath.Join(tmpDir, "actual.txt")
	data := []byte("actual content")
	require.NoError(t, os.WriteFile(actualFile, data, 0o644)) //nolint:gosec // G306: Test file - intentional permissions

	// Create symlink to it
	symlink := filepath.Join(tmpDir, "symlink.txt")
	require.NoError(t, os.Symlink(actualFile, symlink))

	dst := filepath.Join(tmpDir, "dst.txt")

	// Copy from symlink
	err := op.Copy(symlink, dst)
	require.NoError(t, err)

	// Destination should have the content (follows symlink)
	content, err := os.ReadFile(dst) //nolint:gosec // G304: Test file - intentional variable path
	require.NoError(t, err)
	assert.Equal(t, data, content)

	// Destination should be a regular file, not a symlink
	info, err := os.Lstat(dst)
	require.NoError(t, err)
	assert.True(t, info.Mode().IsRegular())
}

// TestCopyBrokenSymlinkSource tests Copy behavior with broken symlink as source.
func TestCopyBrokenSymlinkSource(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping symlink test on Windows")
	}

	tmpDir := t.TempDir()
	op := NewDefaultFileOperator()

	// Create symlink to non-existent file
	symlink := filepath.Join(tmpDir, "broken_symlink.txt")
	require.NoError(t, os.Symlink("/nonexistent/file", symlink))

	dst := filepath.Join(tmpDir, "dst.txt")

	err := op.Copy(symlink, dst)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open source file")
}

// TestCopyBinaryData tests Copy with binary (non-text) data.
func TestCopyBinaryData(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultFileOperator()

	src := filepath.Join(tmpDir, "binary_src.bin")
	dst := filepath.Join(tmpDir, "binary_dst.bin")

	// Create binary data with all byte values
	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
	}
	require.NoError(t, os.WriteFile(src, data, 0o644)) //nolint:gosec // G306: Test file - intentional permissions

	err := op.Copy(src, dst)
	require.NoError(t, err)

	content, err := os.ReadFile(dst) //nolint:gosec // G304: Test file - intentional variable path
	require.NoError(t, err)
	assert.Equal(t, data, content)
}
