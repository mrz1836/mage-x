//nolint:gosec // G304: Test files intentionally read dynamic paths
package fileops

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// WriteFileAtomic Error Path Tests
// =============================================================================

// TestWriteFileAtomic_Success tests the happy path of WriteFileAtomic.
func TestWriteFileAtomic_Success(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "atomic.txt")
	testData := []byte("atomic write test")

	op := NewDefaultSafeFileOperator()
	err := op.WriteFileAtomic(testPath, testData, 0o644)

	require.NoError(t, err)

	// Verify file contents
	content, err := os.ReadFile(testPath)
	require.NoError(t, err)
	assert.Equal(t, testData, content)

	// Verify permissions
	info, err := os.Stat(testPath)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o644), info.Mode().Perm())
}

// TestWriteFileAtomic_CreatesParentDirs tests WriteFileAtomic doesn't create parent dirs.
func TestWriteFileAtomic_NonexistentDir(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "nonexistent", "atomic.txt")

	op := NewDefaultSafeFileOperator()
	err := op.WriteFileAtomic(testPath, []byte("test"), 0o644)

	// Should fail because parent directory doesn't exist
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create temp file")
}

// TestWriteFileAtomic_EmptyData tests writing empty data.
func TestWriteFileAtomic_EmptyData(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "empty.txt")

	op := NewDefaultSafeFileOperator()
	err := op.WriteFileAtomic(testPath, []byte{}, 0o644)

	require.NoError(t, err)

	content, err := os.ReadFile(testPath)
	require.NoError(t, err)
	assert.Empty(t, content)
}

// TestWriteFileAtomic_OverwriteExisting tests overwriting an existing file.
func TestWriteFileAtomic_OverwriteExisting(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "overwrite.txt")

	// Create existing file
	require.NoError(t, os.WriteFile(testPath, []byte("original"), 0o600))

	op := NewDefaultSafeFileOperator()
	newData := []byte("new content")
	err := op.WriteFileAtomic(testPath, newData, 0o644)

	require.NoError(t, err)

	content, err := os.ReadFile(testPath)
	require.NoError(t, err)
	assert.Equal(t, newData, content)
}

// TestWriteFileAtomic_LargeData tests writing large data.
func TestWriteFileAtomic_LargeData(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "large.txt")

	// Create 1MB of data
	largeData := bytes.Repeat([]byte("x"), 1024*1024)

	op := NewDefaultSafeFileOperator()
	err := op.WriteFileAtomic(testPath, largeData, 0o644)

	require.NoError(t, err)

	content, err := os.ReadFile(testPath)
	require.NoError(t, err)
	assert.Equal(t, largeData, content)
}

// =============================================================================
// Copy Error Path Tests
// =============================================================================

// TestCopy_PathTraversal tests that path traversal is detected in Copy.
func TestCopy_PathTraversal(t *testing.T) {
	tmpDir := t.TempDir()

	op := NewDefaultFileOperator()

	t.Run("traversal in source", func(t *testing.T) {
		err := op.Copy("../../../etc/passwd", filepath.Join(tmpDir, "dest.txt"))
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrPathTraversalDetected)
	})

	t.Run("traversal in destination", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "source.txt")
		require.NoError(t, os.WriteFile(testFile, []byte("test"), 0o644))

		err := op.Copy(testFile, "../../../tmp/dest.txt")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrPathTraversalDetected)
	})
}

// TestCopy_SourceNotExist tests copying from non-existent source.
func TestCopy_SourceNotExist(t *testing.T) {
	tmpDir := t.TempDir()

	op := NewDefaultFileOperator()
	err := op.Copy(
		filepath.Join(tmpDir, "nonexistent.txt"),
		filepath.Join(tmpDir, "dest.txt"),
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open source file")
}

// =============================================================================
// YAML Marshal Error Path Tests
// =============================================================================

// TestYAMLMarshal_ErrorHandling tests YAML marshaling edge cases.
func TestYAMLMarshal_ErrorHandling(t *testing.T) {
	op := NewDefaultYAMLOperator(nil)

	t.Run("marshal valid struct", func(t *testing.T) {
		data := struct {
			Name  string `yaml:"name"`
			Value int    `yaml:"value"`
		}{
			Name:  "test",
			Value: 42,
		}

		result, err := op.Marshal(data)
		require.NoError(t, err)
		assert.Contains(t, string(result), "name: test")
		assert.Contains(t, string(result), "value: 42")
	})

	t.Run("marshal map", func(t *testing.T) {
		data := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		}

		result, err := op.Marshal(data)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("marshal slice", func(t *testing.T) {
		data := []string{"a", "b", "c"}

		result, err := op.Marshal(data)
		require.NoError(t, err)
		assert.Contains(t, string(result), "- a")
		assert.Contains(t, string(result), "- b")
		assert.Contains(t, string(result), "- c")
	})

	t.Run("marshal nil", func(t *testing.T) {
		result, err := op.Marshal(nil)
		require.NoError(t, err)
		assert.Contains(t, string(result), "null")
	})
}

// TestWriteYAML_ErrorHandling tests WriteYAML error paths.
func TestWriteYAML_ErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()
	op := NewDefaultYAMLOperator(nil)

	t.Run("write to non-existent directory", func(t *testing.T) {
		err := op.WriteYAML(
			filepath.Join(tmpDir, "nonexistent", "file.yaml"),
			map[string]string{"key": "value"},
		)
		require.Error(t, err)
	})

	t.Run("write valid YAML", func(t *testing.T) {
		testPath := filepath.Join(tmpDir, "test.yaml")
		data := struct {
			Name string `yaml:"name"`
		}{Name: "test"}

		err := op.WriteYAML(testPath, data)
		require.NoError(t, err)

		// Verify file exists and has content
		content, err := os.ReadFile(testPath)
		require.NoError(t, err)
		assert.Contains(t, string(content), "name: test")
	})
}

// =============================================================================
// WriteFileWithBackup Tests
// =============================================================================

// TestWriteFileWithBackup_NewFile tests writing a new file (no backup needed).
func TestWriteFileWithBackup_NewFile(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "new.txt")

	op := NewDefaultSafeFileOperator()
	err := op.WriteFileWithBackup(testPath, []byte("new content"), 0o644)

	require.NoError(t, err)

	// Verify main file
	content, err := os.ReadFile(testPath)
	require.NoError(t, err)
	assert.Equal(t, []byte("new content"), content)

	// Backup should not exist for new file
	_, err = os.Stat(testPath + ".bak")
	assert.True(t, os.IsNotExist(err))
}

// TestWriteFileWithBackup_ExistingFile tests creating backup of existing file.
func TestWriteFileWithBackup_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "existing.txt")
	originalContent := []byte("original content")
	newContent := []byte("new content")

	// Create existing file
	require.NoError(t, os.WriteFile(testPath, originalContent, 0o600))

	op := NewDefaultSafeFileOperator()
	err := op.WriteFileWithBackup(testPath, newContent, 0o644)

	require.NoError(t, err)

	// Verify main file has new content
	content, err := os.ReadFile(testPath)
	require.NoError(t, err)
	assert.Equal(t, newContent, content)

	// Verify backup has original content
	backupContent, err := os.ReadFile(testPath + ".bak")
	require.NoError(t, err)
	assert.Equal(t, originalContent, backupContent)
}

// =============================================================================
// JSON Operator Tests
// =============================================================================

// TestJSONOperator_NilFileOps tests JSON operator with nil fileops defaults to DefaultFileOperator.
func TestJSONOperator_NilFileOps(t *testing.T) {
	op := NewDefaultJSONOperator(nil)
	require.NotNil(t, op)

	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test.json")

	data := map[string]string{"key": "value"}
	err := op.WriteJSON(testPath, data)
	require.NoError(t, err)

	var result map[string]string
	err = op.ReadJSON(testPath, &result)
	require.NoError(t, err)
	assert.Equal(t, data, result)
}

// TestJSONOperator_WriteJSONIndent tests indented JSON writing.
func TestJSONOperator_WriteJSONIndent(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "indented.json")

	op := NewDefaultJSONOperator(nil)
	data := map[string]string{"key": "value"}

	err := op.WriteJSONIndent(testPath, data, "", "  ")
	require.NoError(t, err)

	content, err := os.ReadFile(testPath)
	require.NoError(t, err)
	// Should be pretty-printed with newlines
	assert.Contains(t, string(content), "\n")
	assert.Contains(t, string(content), "key")
	assert.Contains(t, string(content), "value")
}

// =============================================================================
// YAML Operator Tests
// =============================================================================

// TestYAMLOperator_NilFileOps tests YAML operator with nil fileops defaults to DefaultFileOperator.
func TestYAMLOperator_NilFileOps(t *testing.T) {
	op := NewDefaultYAMLOperator(nil)
	require.NotNil(t, op)

	tmpDir := t.TempDir()
	testPath := filepath.Join(tmpDir, "test.yaml")

	data := map[string]string{"key": "value"}
	err := op.WriteYAML(testPath, data)
	require.NoError(t, err)

	var result map[string]string
	err = op.ReadYAML(testPath, &result)
	require.NoError(t, err)
	assert.Equal(t, data, result)
}
