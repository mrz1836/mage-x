package fileops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestDefaultFileOperator_Extended(t *testing.T) {
	// Create temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "fileops-extended-test-*")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	ops := NewDefaultFileOperator()

	t.Run("RemoveAll", func(t *testing.T) {
		// Create nested directory structure
		nestedDir := filepath.Join(tmpDir, "nested", "deep", "dir")
		err := ops.MkdirAll(nestedDir, 0o755)
		require.NoError(t, err)

		// Create some files
		file1 := filepath.Join(tmpDir, "nested", "file1.txt")
		file2 := filepath.Join(nestedDir, "file2.txt")
		err = ops.WriteFile(file1, []byte("content1"), 0o644)
		require.NoError(t, err)
		err = ops.WriteFile(file2, []byte("content2"), 0o644)
		require.NoError(t, err)

		// Remove all
		err = ops.RemoveAll(filepath.Join(tmpDir, "nested"))
		require.NoError(t, err)

		// Verify removed
		assert.False(t, ops.Exists(filepath.Join(tmpDir, "nested")))
	})

	t.Run("Stat", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "stat_test.txt")
		err := ops.WriteFile(testFile, []byte("stat test"), 0o644)
		require.NoError(t, err)

		info, err := ops.Stat(testFile)
		require.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "stat_test.txt", info.Name())
		assert.False(t, info.IsDir())
	})

	t.Run("Chmod", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "chmod_test.txt")
		err := ops.WriteFile(testFile, []byte("chmod test"), 0o644)
		require.NoError(t, err)

		// Change permissions
		err = ops.Chmod(testFile, 0o600)
		require.NoError(t, err)

		// Verify permissions changed
		info, err := ops.Stat(testFile)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm())
	})

	t.Run("ReadDir", func(t *testing.T) {
		// Create test directory with files
		testDir := filepath.Join(tmpDir, "readdir_test")
		err := ops.MkdirAll(testDir, 0o755)
		require.NoError(t, err)

		// Create some files
		for i := 0; i < 3; i++ {
			filename := filepath.Join(testDir, string(rune('a'+i))+".txt")
			writeErr := ops.WriteFile(filename, []byte("content"), 0o644)
			require.NoError(t, writeErr)
		}

		// Read directory
		entries, err := ops.ReadDir(testDir)
		require.NoError(t, err)
		assert.Len(t, entries, 3)
	})
}

func TestDefaultJSONOperator_Extended(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "json-extended-test-*")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	fileOp := NewDefaultFileOperator()
	jsonOp := NewDefaultJSONOperator(fileOp)

	t.Run("WriteJSONIndent", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "indent.json")
		data := map[string]interface{}{
			"name":  "test",
			"value": 123,
			"nested": map[string]string{
				"key": "value",
			},
		}

		err := jsonOp.WriteJSONIndent(testFile, data, "", "  ")
		require.NoError(t, err)

		// Verify the file was written with proper indentation
		content, err := fileOp.ReadFile(testFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), "  \"name\"")
		assert.Contains(t, string(content), "    \"key\"")
	})
}

func TestDefaultYAMLOperator_Extended(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "yaml-extended-test-*")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	fileOp := NewDefaultFileOperator()
	yamlOp := NewDefaultYAMLOperator(fileOp)

	t.Run("Unmarshal", func(t *testing.T) {
		yamlData := `
name: test
value: 123
nested:
  key: value
`
		var result map[string]interface{}
		err := yamlOp.Unmarshal([]byte(yamlData), &result)
		require.NoError(t, err)
		assert.Equal(t, "test", result["name"])
		assert.Equal(t, 123, result["value"])
	})
}

func TestFileOps_Extended(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fileops-facade-test-*")
	require.NoError(t, err)
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	ops := New()

	t.Run("WriteJSONSafe", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "subdir", "safe.json")
		data := map[string]string{
			"key": "value",
		}

		err := ops.WriteJSONSafe(testFile, data)
		require.NoError(t, err)

		// Verify file was written
		assert.True(t, ops.File.Exists(testFile))

		// Verify content
		var loaded map[string]string
		err = ops.JSON.ReadJSON(testFile, &loaded)
		require.NoError(t, err)
		assert.Equal(t, data, loaded)
	})

	t.Run("WriteYAMLSafe", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "subdir2", "safe.yaml")
		data := map[string]string{
			"key": "value",
		}

		err := ops.WriteYAMLSafe(testFile, data)
		require.NoError(t, err)

		// Verify file was written
		assert.True(t, ops.File.Exists(testFile))

		// Verify content
		var loaded map[string]string
		err = ops.YAML.ReadYAML(testFile, &loaded)
		require.NoError(t, err)
		assert.Equal(t, data, loaded)
	})

	t.Run("LoadConfig", func(t *testing.T) {
		// Create test config files
		configYAML := filepath.Join(tmpDir, "config.yaml")
		configJSON := filepath.Join(tmpDir, "config.json")
		configUnknown := filepath.Join(tmpDir, "config.conf")

		config := map[string]string{"key": "yaml-value"}
		err := ops.YAML.WriteYAML(configYAML, config)
		require.NoError(t, err)

		config = map[string]string{"key": "json-value"}
		err = ops.JSON.WriteJSON(configJSON, config)
		require.NoError(t, err)

		// Write YAML content to unknown extension file
		unknownConfig := map[string]string{"key": "unknown-value"}
		yamlData, err := yaml.Marshal(unknownConfig)
		require.NoError(t, err)
		err = ops.File.WriteFile(configUnknown, yamlData, 0o644)
		require.NoError(t, err)

		// Test loading YAML
		var loaded map[string]string
		path, err := ops.LoadConfig([]string{configYAML}, &loaded)
		require.NoError(t, err)
		assert.Equal(t, configYAML, path) //nolint:testifylint // path comparison, not content comparison
		assert.Equal(t, "yaml-value", loaded["key"])

		// Test loading JSON
		loaded = map[string]string{}
		path, err = ops.LoadConfig([]string{configJSON}, &loaded)
		require.NoError(t, err)
		assert.Equal(t, configJSON, path) //nolint:testifylint // path comparison, not content comparison
		assert.Equal(t, "json-value", loaded["key"])

		// Test loading unknown extension (tries YAML first)
		loaded = map[string]string{}
		path, err = ops.LoadConfig([]string{configUnknown}, &loaded)
		require.NoError(t, err)
		assert.Equal(t, configUnknown, path)
		assert.Equal(t, "unknown-value", loaded["key"])

		// Test non-existent files
		_, err = ops.LoadConfig([]string{"/nonexistent1", "/nonexistent2"}, &loaded)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no valid config file found")
	})

	t.Run("CopyFile", func(t *testing.T) {
		srcFile := filepath.Join(tmpDir, "source.txt")
		dstFile := filepath.Join(tmpDir, "destination.txt")

		// Create source file
		err := ops.File.WriteFile(srcFile, []byte("test content"), 0o644)
		require.NoError(t, err)

		// Copy file
		err = ops.CopyFile(srcFile, dstFile)
		require.NoError(t, err)

		// Verify destination exists and has same content
		assert.True(t, ops.File.Exists(dstFile))
		content, err := ops.File.ReadFile(dstFile)
		require.NoError(t, err)
		assert.Equal(t, "test content", string(content))

		// Test copying non-existent file
		err = ops.CopyFile("/nonexistent", dstFile)
		require.Error(t, err)
	})

	t.Run("BackupFile", func(t *testing.T) {
		originalFile := filepath.Join(tmpDir, "original.txt")

		// Create original file
		err := ops.File.WriteFile(originalFile, []byte("original content"), 0o644)
		require.NoError(t, err)

		// Create backup
		err = ops.BackupFile(originalFile)
		require.NoError(t, err)

		// Backup should exist at originalFile.bak
		backupPath := originalFile + ".bak"
		assert.True(t, ops.File.Exists(backupPath))

		// Verify backup content
		content, err := ops.File.ReadFile(backupPath)
		require.NoError(t, err)
		assert.Equal(t, "original content", string(content))

		// Test backup of non-existent file
		err = ops.BackupFile("/nonexistent")
		require.Error(t, err)
	})

	t.Run("CleanupBackups", func(t *testing.T) {
		// Create multiple backup files
		for i := 0; i < 5; i++ {
			filename := fmt.Sprintf("backup-%d.bak", i)
			filePath := filepath.Join(tmpDir, filename)
			err := ops.File.WriteFile(filePath, []byte("backup"), 0o644)
			require.NoError(t, err)
		}

		// Clean up backup files matching pattern
		err := ops.CleanupBackups(tmpDir, "*.bak")
		require.NoError(t, err)

		// Verify all backup files are removed
		entries, err := ops.File.ReadDir(tmpDir)
		require.NoError(t, err)

		backupCount := 0
		for _, entry := range entries {
			if strings.HasSuffix(entry.Name(), ".bak") {
				backupCount++
			}
		}
		assert.Equal(t, 0, backupCount)
	})
}

func TestNewFileOperator(t *testing.T) {
	op := NewFileOperator()
	assert.NotNil(t, op)
	assert.IsType(t, &DefaultFileOperator{}, op)
}

func TestNewWithOptions(t *testing.T) {
	fileOp := NewDefaultFileOperator()
	jsonOp := NewDefaultJSONOperator(fileOp)
	yamlOp := NewDefaultYAMLOperator(fileOp)
	safeOp := NewDefaultSafeFileOperator()

	ops := NewWithOptions(fileOp, jsonOp, yamlOp, safeOp)
	assert.NotNil(t, ops)
	assert.Equal(t, fileOp, ops.File)
	assert.Equal(t, jsonOp, ops.JSON)
	assert.Equal(t, yamlOp, ops.YAML)
	assert.Equal(t, safeOp, ops.Safe)
}

func TestFileOps_ErrorHandling(t *testing.T) {
	ops := New()

	t.Run("WriteJSONSafe with invalid data", func(t *testing.T) {
		// Channel cannot be marshaled to JSON
		invalidData := make(chan int)
		err := ops.WriteJSONSafe("/tmp/test.json", invalidData)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal JSON")
	})

	t.Run("WriteYAMLSafe with invalid data", func(t *testing.T) {
		// Create data that will fail during write (invalid path)
		validData := map[string]string{"key": "value"}
		err := ops.WriteYAMLSafe("/invalid\x00path/test.yaml", validData)
		require.Error(t, err)
	})

	t.Run("LoadConfig with invalid JSON/YAML", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "invalid-config-*")
		require.NoError(t, err)
		defer func() {
			if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
				t.Logf("Failed to remove temp dir %s: %v", tmpDir, removeErr)
			}
		}()

		// Create invalid JSON file
		invalidFile := filepath.Join(tmpDir, "invalid.json")
		err = ops.File.WriteFile(invalidFile, []byte("{invalid json}"), 0o644)
		require.NoError(t, err)

		var result map[string]string
		_, err = ops.LoadConfig([]string{invalidFile}, &result)
		require.Error(t, err)
	})
}
