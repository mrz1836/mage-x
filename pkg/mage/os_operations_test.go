package mage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultOSOperations_Getenv tests environment variable retrieval
func TestDefaultOSOperations_Getenv(t *testing.T) {
	ops := DefaultOSOperations{}

	// Set a test environment variable
	testKey := "MAGE_X_TEST_VAR"
	testValue := "test_value"
	require.NoError(t, os.Setenv(testKey, testValue))
	defer func() {
		require.NoError(t, os.Unsetenv(testKey))
	}()

	// Test retrieval
	result := ops.Getenv(testKey)
	assert.Equal(t, testValue, result)

	// Test non-existent variable
	result = ops.Getenv("NONEXISTENT_VAR_12345")
	assert.Empty(t, result)
}

// TestDefaultOSOperations_TempDir tests temp directory retrieval
func TestDefaultOSOperations_TempDir(t *testing.T) {
	ops := DefaultOSOperations{}

	tempDir := ops.TempDir()
	assert.NotEmpty(t, tempDir, "temp dir should not be empty")

	// Verify it's a valid directory
	info, err := os.Stat(tempDir)
	require.NoError(t, err, "temp dir should exist")
	assert.True(t, info.IsDir(), "should be a directory")
}

// TestDefaultOSOperations_Remove tests file removal
func TestDefaultOSOperations_Remove(t *testing.T) {
	ops := DefaultOSOperations{}

	// Create a temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_file.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0o600)
	require.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(testFile)
	require.NoError(t, err)

	// Remove the file
	err = ops.Remove(testFile)
	require.NoError(t, err)

	// Verify file is gone
	_, err = os.Stat(testFile)
	assert.True(t, os.IsNotExist(err), "file should be removed")
}

// TestDefaultOSOperations_Remove_NonExistent tests removing non-existent file
func TestDefaultOSOperations_Remove_NonExistent(t *testing.T) {
	ops := DefaultOSOperations{}

	err := ops.Remove("/nonexistent/path/to/file.txt")
	assert.Error(t, err, "should error when removing non-existent file")
}

// TestDefaultOSOperations_FileExists tests file existence checking
func TestDefaultOSOperations_FileExists(t *testing.T) {
	ops := DefaultOSOperations{}

	// Create a temporary file
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_file.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0o600)
	require.NoError(t, err)

	// Test existing file
	assert.True(t, ops.FileExists(testFile), "file should exist")

	// Test non-existent file
	assert.False(t, ops.FileExists(filepath.Join(tempDir, "nonexistent.txt")), "file should not exist")

	// Test directory (behavior depends on utils.FileExists implementation)
	// Note: utils.FileExists may return true for directories as well
	_ = ops.FileExists(tempDir) // Just verify it doesn't panic
}

// TestDefaultOSOperations_WriteFile tests file writing
func TestDefaultOSOperations_WriteFile(t *testing.T) {
	ops := DefaultOSOperations{}

	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test_file.txt")
	testContent := []byte("test content line 1\ntest content line 2")

	// Write file
	err := ops.WriteFile(testFile, testContent, 0o600)
	require.NoError(t, err)

	// Verify content
	content, err := os.ReadFile(testFile) // #nosec G304 -- This is a test file reading from a temp directory
	require.NoError(t, err)
	assert.Equal(t, testContent, content)

	// Verify permissions
	info, err := os.Stat(testFile)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
}

// TestDefaultOSOperations_WriteFile_InvalidPath tests writing to invalid path
func TestDefaultOSOperations_WriteFile_InvalidPath(t *testing.T) {
	ops := DefaultOSOperations{}

	// Try to write to a non-existent directory
	err := ops.WriteFile("/nonexistent/dir/file.txt", []byte("test"), 0o600)
	assert.Error(t, err, "should error when writing to invalid path")
}

// TestDefaultOSOperations_Symlink tests symlink creation
func TestDefaultOSOperations_Symlink(t *testing.T) {
	ops := DefaultOSOperations{}

	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "target.txt")
	symlinkFile := filepath.Join(tempDir, "link.txt")

	// Create target file
	err := os.WriteFile(targetFile, []byte("target content"), 0o600)
	require.NoError(t, err)

	// Create symlink
	err = ops.Symlink(targetFile, symlinkFile)
	require.NoError(t, err)

	// Verify symlink exists
	info, err := os.Lstat(symlinkFile)
	require.NoError(t, err)
	assert.NotEqual(t, 0, info.Mode()&os.ModeSymlink, "should be a symlink")
}

// TestDefaultOSOperations_Readlink tests reading symlink destination
func TestDefaultOSOperations_Readlink(t *testing.T) {
	ops := DefaultOSOperations{}

	tempDir := t.TempDir()
	targetFile := filepath.Join(tempDir, "target.txt")
	symlinkFile := filepath.Join(tempDir, "link.txt")

	// Create target file
	err := os.WriteFile(targetFile, []byte("target content"), 0o600)
	require.NoError(t, err)

	// Create symlink
	err = os.Symlink(targetFile, symlinkFile)
	require.NoError(t, err)

	// Read symlink
	destination, err := ops.Readlink(symlinkFile)
	require.NoError(t, err)
	assert.Equal(t, targetFile, destination)
}

// TestDefaultOSOperations_Readlink_NotSymlink tests reading non-symlink
func TestDefaultOSOperations_Readlink_NotSymlink(t *testing.T) {
	ops := DefaultOSOperations{}

	tempDir := t.TempDir()
	regularFile := filepath.Join(tempDir, "regular.txt")

	// Create regular file
	err := os.WriteFile(regularFile, []byte("content"), 0o600)
	require.NoError(t, err)

	// Try to read as symlink
	_, err = ops.Readlink(regularFile)
	assert.Error(t, err, "should error when reading non-symlink")
}

// TestGetOSOperations_Default tests getting the default OS operations
func TestGetOSOperations_Default(t *testing.T) {
	// Reset to ensure we have default
	ResetOSOperations()

	ops := GetOSOperations()
	assert.NotNil(t, ops)

	// Should return DefaultOSOperations
	_, ok := ops.(DefaultOSOperations)
	assert.True(t, ok, "should be DefaultOSOperations")
}

// TestSetOSOperations tests setting custom OS operations
func TestSetOSOperations(t *testing.T) {
	// Reset first
	ResetOSOperations()

	// Create a mock implementation
	mockOps := &mockOSOperations{}

	// Set it
	err := SetOSOperations(mockOps)
	require.NoError(t, err)

	// Verify it's set
	ops := GetOSOperations()
	assert.Equal(t, mockOps, ops)

	// Cleanup
	ResetOSOperations()
}

// TestSetOSOperations_Nil tests setting nil OS operations
func TestSetOSOperations_Nil(t *testing.T) {
	// Reset first
	ResetOSOperations()

	// Try to set nil
	err := SetOSOperations(nil)
	require.Error(t, err, "should error when setting nil")
	assert.Equal(t, errOSOperationsNil, err)

	// Verify default is still set
	ops := GetOSOperations()
	assert.NotNil(t, ops)
}

// TestOSOperationsReset tests resetting OS operations to default
func TestOSOperationsReset(t *testing.T) {
	// Set a mock
	mockOps := &mockOSOperations{}
	err := SetOSOperations(mockOps)
	require.NoError(t, err)

	// Verify mock is set
	ops := GetOSOperations()
	assert.Equal(t, mockOps, ops)

	// Reset
	ResetOSOperations()

	// Verify default is restored
	ops = GetOSOperations()
	_, ok := ops.(DefaultOSOperations)
	assert.True(t, ok, "should be reset to DefaultOSOperations")
}

// mockOSOperations is a mock implementation for testing
type mockOSOperations struct{}

func (m *mockOSOperations) Getenv(key string) string {
	return "mock_" + key
}

func (m *mockOSOperations) Remove(path string) error {
	return nil
}

func (m *mockOSOperations) TempDir() string {
	return "/tmp/mock"
}

func (m *mockOSOperations) FileExists(path string) bool {
	return true
}

func (m *mockOSOperations) WriteFile(path string, data []byte, perm os.FileMode) error {
	return nil
}

func (m *mockOSOperations) Symlink(oldname, newname string) error {
	return nil
}

func (m *mockOSOperations) Readlink(name string) (string, error) {
	return "mock_link", nil
}
