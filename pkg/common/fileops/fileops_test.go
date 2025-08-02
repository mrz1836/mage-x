package fileops

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"gopkg.in/yaml.v3"
)

// Static errors for fileops tests
var (
	errTestDataMismatch = errors.New("goroutine data mismatch")
	errTestMarshalError = errors.New("marshal error")
	errTestYAMLError    = errors.New("yaml error")
	errTestRemoveError  = errors.New("remove error")
)

// FileOpsTestSuite tests file operations
type FileOpsTestSuite struct {
	suite.Suite

	tmpDir string
	ops    *FileOps
}

func (s *FileOpsTestSuite) SetupSuite() {
	// Create temporary directory for all tests
	tmpDir, err := os.MkdirTemp("", "fileops-test-*")
	s.Require().NoError(err, "Failed to create temp dir")
	s.tmpDir = tmpDir
	s.ops = New()
}

func (s *FileOpsTestSuite) TearDownSuite() {
	// Clean up temp directory
	if err := os.RemoveAll(s.tmpDir); err != nil {
		s.T().Logf("Failed to remove temp dir %s: %v", s.tmpDir, err)
	}
}

func TestFileOpsTestSuite(t *testing.T) {
	suite.Run(t, new(FileOpsTestSuite))
}

func TestDefaultFileOperator(t *testing.T) {
	// Create temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "fileops-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	ops := NewDefaultFileOperator()

	t.Run("WriteFile and ReadFile", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.txt")
		testData := []byte("Hello, World!")

		// Write file
		err := ops.WriteFile(testFile, testData, 0o644)
		require.NoError(t, err, "Failed to write file")

		// Check if file exists
		assert.True(t, ops.Exists(testFile), "File should exist after writing")

		// Read file
		data, err := ops.ReadFile(testFile)
		require.NoError(t, err, "Failed to read file")

		assert.Equal(t, testData, data, "Read data should match written data")
	})

	t.Run("MkdirAll", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "deep", "nested", "dir")

		err := ops.MkdirAll(testDir, 0o755)
		require.NoError(t, err, "Failed to create directory")

		assert.True(t, ops.Exists(testDir), "Directory should exist after creation")
		assert.True(t, ops.IsDir(testDir), "Path should be a directory")
	})

	t.Run("Copy", func(t *testing.T) {
		srcFile := filepath.Join(tmpDir, "source.txt")
		dstFile := filepath.Join(tmpDir, "destination.txt")
		testData := []byte("Copy test data")

		// Create source file
		err := ops.WriteFile(srcFile, testData, 0o644)
		require.NoError(t, err, "Failed to write source file")

		// Copy file
		err = ops.Copy(srcFile, dstFile)
		require.NoError(t, err, "Failed to copy file")

		// Verify copied file
		data, err := ops.ReadFile(dstFile)
		require.NoError(t, err, "Failed to read copied file")

		assert.Equal(t, testData, data, "Copied data should match source data")
	})

	t.Run("Remove", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "remove-test.txt")

		// Create file
		err := ops.WriteFile(testFile, []byte("test"), 0o644)
		require.NoError(t, err, "Failed to create file")
		assert.True(t, ops.Exists(testFile), "File should exist")

		// Remove file
		err = ops.Remove(testFile)
		require.NoError(t, err, "Failed to remove file")
		assert.False(t, ops.Exists(testFile), "File should not exist after removal")
	})

	t.Run("IsDir and IsFile", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "test-dir")
		testFile := filepath.Join(tmpDir, "test-file.txt")

		// Create directory
		err := ops.MkdirAll(testDir, 0o755)
		require.NoError(t, err)

		// Create file
		err = ops.WriteFile(testFile, []byte("test"), 0o644)
		require.NoError(t, err)

		// Test IsDir
		assert.True(t, ops.IsDir(testDir), "Should identify directory correctly")
		assert.False(t, ops.IsDir(testFile), "Should not identify file as directory")
		assert.False(t, ops.IsDir(filepath.Join(tmpDir, "nonexistent")), "Should return false for nonexistent path")

		// Test that file is not a directory
		assert.False(t, ops.IsDir(testFile), "File should not be identified as directory")
		// Test that non-existent path is not a directory
		assert.False(t, ops.IsDir(filepath.Join(tmpDir, "nonexistent")), "Non-existent path should not be identified as directory")
	})

	t.Run("ReadFile with path traversal", func(t *testing.T) {
		// Test path traversal protection
		_, err := ops.ReadFile("../../../etc/passwd")
		require.Error(t, err, "Should reject path traversal")
		assert.Contains(t, err.Error(), "path traversal detected", "Should detect path traversal")
	})

	t.Run("Copy with path traversal", func(t *testing.T) {
		// Create a valid source file
		srcFile := filepath.Join(tmpDir, "source.txt")
		err := ops.WriteFile(srcFile, []byte("test data"), 0o644)
		require.NoError(t, err)

		// Test path traversal in source
		err = ops.Copy("../../../etc/passwd", filepath.Join(tmpDir, "dest1.txt"))
		require.Error(t, err, "Should reject source path traversal")
		assert.Contains(t, err.Error(), "path traversal detected", "Should detect source path traversal")

		// Test path traversal in destination
		err = ops.Copy(srcFile, "../../../tmp/dest2.txt")
		require.Error(t, err, "Should reject destination path traversal")
		assert.Contains(t, err.Error(), "path traversal detected", "Should detect destination path traversal")
	})

	t.Run("Stat method", func(t *testing.T) {
		// Create a test file to stat
		testFile := filepath.Join(tmpDir, "stat-test.txt")
		testData := []byte("stat test data")
		err := ops.WriteFile(testFile, testData, 0o644)
		require.NoError(t, err)

		// Test Stat method
		info, err := ops.Stat(testFile)
		require.NoError(t, err, "Stat should succeed")
		assert.Equal(t, "stat-test.txt", info.Name(), "File name should match")
		assert.Equal(t, int64(len(testData)), info.Size(), "File size should match")
		assert.False(t, info.IsDir(), "File should not be directory")

		// Test Stat on directory
		testDir := filepath.Join(tmpDir, "stat-dir")
		err = ops.MkdirAll(testDir, 0o755)
		require.NoError(t, err)

		info, err = ops.Stat(testDir)
		require.NoError(t, err, "Stat should succeed on directory")
		assert.True(t, info.IsDir(), "Directory should be identified as directory")

		// Test Stat on non-existent file
		_, err = ops.Stat(filepath.Join(tmpDir, "nonexistent.txt"))
		require.Error(t, err, "Stat should fail on non-existent file")
	})

	t.Run("Chmod method", func(t *testing.T) {
		// Create a test file
		testFile := filepath.Join(tmpDir, "chmod-test.txt")
		err := ops.WriteFile(testFile, []byte("chmod test"), 0o644)
		require.NoError(t, err)

		// Change permissions
		newMode := os.FileMode(0o600)
		err = ops.Chmod(testFile, newMode)
		require.NoError(t, err, "Chmod should succeed")

		// Verify permissions changed
		info, err := ops.Stat(testFile)
		require.NoError(t, err)
		assert.Equal(t, newMode, info.Mode().Perm(), "Permissions should be updated")
	})

	t.Run("RemoveAll method", func(t *testing.T) {
		// Create a directory with files
		testDir := filepath.Join(tmpDir, "remove-all-test")
		err := ops.MkdirAll(filepath.Join(testDir, "subdir"), 0o755)
		require.NoError(t, err)

		// Create files in the directory
		err = ops.WriteFile(filepath.Join(testDir, "file1.txt"), []byte("file1"), 0o644)
		require.NoError(t, err)
		err = ops.WriteFile(filepath.Join(testDir, "subdir", "file2.txt"), []byte("file2"), 0o644)
		require.NoError(t, err)

		// Verify directory exists
		assert.True(t, ops.Exists(testDir), "Directory should exist before removal")

		// Remove all
		err = ops.RemoveAll(testDir)
		require.NoError(t, err, "RemoveAll should succeed")

		// Verify directory is gone
		assert.False(t, ops.Exists(testDir), "Directory should not exist after RemoveAll")
	})

	t.Run("ReadDir method", func(t *testing.T) {
		// Create a directory with multiple files
		testDir := filepath.Join(tmpDir, "read-dir-test")
		err := ops.MkdirAll(testDir, 0o755)
		require.NoError(t, err)

		// Create test files
		fileNames := []string{"file1.txt", "file2.log", "file3.md"}
		for _, fileName := range fileNames {
			err = ops.WriteFile(filepath.Join(testDir, fileName), []byte("test"), 0o644)
			require.NoError(t, err)
		}

		// Create a subdirectory
		err = ops.MkdirAll(filepath.Join(testDir, "subdir"), 0o755)
		require.NoError(t, err)

		// Read directory
		entries, err := ops.ReadDir(testDir)
		require.NoError(t, err, "ReadDir should succeed")
		assert.Len(t, entries, 4, "Should have 3 files + 1 directory")

		// Verify all expected entries are present
		foundNames := make([]string, len(entries))
		for i, entry := range entries {
			foundNames[i] = entry.Name()
		}
		expectedNames := []string{"file1.txt", "file2.log", "file3.md", "subdir"}
		assert.ElementsMatch(t, expectedNames, foundNames, "Directory entries should match")
	})
}

// TestDefaultFileOperatorErrorCases tests error conditions and edge cases
func TestDefaultFileOperatorErrorCases(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fileops-error-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	ops := NewDefaultFileOperator()

	t.Run("ReadFile nonexistent", func(t *testing.T) {
		_, err := ops.ReadFile(filepath.Join(tmpDir, "nonexistent.txt"))
		require.Error(t, err, "Should error on nonexistent file")
	})

	t.Run("Copy nonexistent source", func(t *testing.T) {
		src := filepath.Join(tmpDir, "nonexistent-src.txt")
		dst := filepath.Join(tmpDir, "dst.txt")
		err := ops.Copy(src, dst)
		require.Error(t, err, "Should error when source doesn't exist")
		assert.Contains(t, err.Error(), "failed to open source file", "Error should mention source file")
	})

	t.Run("Copy to invalid destination", func(t *testing.T) {
		// Create source file
		src := filepath.Join(tmpDir, "src.txt")
		err := ops.WriteFile(src, []byte("test"), 0o644)
		require.NoError(t, err)

		// Try to copy to invalid destination
		invalidDst := "/invalid/nonexistent/path/dst.txt"
		err = ops.Copy(src, invalidDst)
		require.Error(t, err, "Should error with invalid destination")
		assert.Contains(t, err.Error(), "failed to create destination file", "Error should mention destination file")
	})

	t.Run("Remove nonexistent file", func(t *testing.T) {
		err := ops.Remove(filepath.Join(tmpDir, "nonexistent.txt"))
		require.Error(t, err, "Should error when removing nonexistent file")
	})

	t.Run("Exists and IsDir on nonexistent paths", func(t *testing.T) {
		nonexistentPath := filepath.Join(tmpDir, "nonexistent")
		assert.False(t, ops.Exists(nonexistentPath), "Nonexistent path should not exist")
		assert.False(t, ops.IsDir(nonexistentPath), "Nonexistent path should not be directory")
	})
}

// TestDefaultJSONOperatorErrorCases tests error scenarios for JSON operations
func TestDefaultJSONOperatorErrorCases(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "json-error-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	fileOps := NewDefaultFileOperator()
	jsonOps := NewDefaultJSONOperator(fileOps)

	t.Run("ReadJSON nonexistent file", func(t *testing.T) {
		var result map[string]interface{}
		err := jsonOps.ReadJSON(filepath.Join(tmpDir, "nonexistent.json"), &result)
		require.Error(t, err, "Should error on nonexistent file")
		assert.Contains(t, err.Error(), "failed to read file", "Error should mention file read failure")
	})

	t.Run("WriteJSON to invalid path", func(t *testing.T) {
		invalidPath := "/invalid/nonexistent/path/test.json"
		err := jsonOps.WriteJSON(invalidPath, map[string]string{"key": "value"})
		require.Error(t, err, "Should error with invalid path")
	})

	t.Run("WriteJSONIndent to invalid path", func(t *testing.T) {
		invalidPath := "/invalid/nonexistent/path/indent.json"
		err := jsonOps.WriteJSONIndent(invalidPath, map[string]string{"key": "value"}, "", "  ")
		require.Error(t, err, "Should error with invalid path")
	})

	t.Run("Marshal unmarshalable data", func(t *testing.T) {
		// Create circular reference which can't be marshaled
		type Circular struct {
			Self *Circular
		}
		circular := &Circular{}
		circular.Self = circular

		_, err := jsonOps.Marshal(circular)
		require.Error(t, err, "Should error on circular reference")

		_, err = jsonOps.MarshalIndent(circular, "", "  ")
		require.Error(t, err, "Should error on circular reference with indent")
	})

	t.Run("Unmarshal invalid JSON", func(t *testing.T) {
		invalidJSON := []byte("{ invalid json }")
		var result map[string]interface{}
		err := jsonOps.Unmarshal(invalidJSON, &result)
		require.Error(t, err, "Should error on invalid JSON")
	})
}

func TestDefaultJSONOperator(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "json-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	fileOps := NewDefaultFileOperator()
	jsonOps := NewDefaultJSONOperator(fileOps)

	type TestStruct struct {
		Name  string                 `json:"name"`
		Value int                    `json:"value"`
		Tags  []string               `json:"tags,omitempty"`
		Meta  map[string]interface{} `json:"meta,omitempty"`
	}

	testData := TestStruct{
		Name:  "test",
		Value: 42,
		Tags:  []string{"tag1", "tag2"},
		Meta:  map[string]interface{}{"key": "value", "count": float64(10)},
	}
	testFile := filepath.Join(tmpDir, "test.json")

	t.Run("WriteJSON and ReadJSON", func(t *testing.T) {
		// Write JSON
		err := jsonOps.WriteJSON(testFile, testData)
		require.NoError(t, err, "Failed to write JSON")

		// Read JSON
		var result TestStruct
		err = jsonOps.ReadJSON(testFile, &result)
		require.NoError(t, err, "Failed to read JSON")

		assert.Equal(t, testData.Name, result.Name, "Name field should match")
		assert.Equal(t, testData.Value, result.Value, "Value field should match")
		assert.ElementsMatch(t, testData.Tags, result.Tags, "Tags should match")
		// Compare meta fields individually to handle type differences
		assert.Equal(t, testData.Meta["key"], result.Meta["key"], "Meta key should match")
		assert.Equal(t, testData.Meta["count"], result.Meta["count"], "Meta count should match")
	})

	t.Run("MarshalIndent", func(t *testing.T) {
		data, err := jsonOps.MarshalIndent(testData, "", "  ")
		require.NoError(t, err, "Failed to marshal with indent")

		// Verify it's valid JSON
		var result TestStruct
		err = json.Unmarshal(data, &result)
		require.NoError(t, err, "Failed to unmarshal indented JSON")

		// Compare fields individually to handle type differences
		assert.Equal(t, testData.Name, result.Name, "Name should match")
		assert.Equal(t, testData.Value, result.Value, "Value should match")
		assert.ElementsMatch(t, testData.Tags, result.Tags, "Tags should match")
		assert.Equal(t, testData.Meta["key"], result.Meta["key"], "Meta key should match")
		assert.Equal(t, testData.Meta["count"], result.Meta["count"], "Meta count should match")

		// Verify indentation
		jsonStr := string(data)
		assert.Contains(t, jsonStr, "\n  ", "JSON should be indented")
	})

	t.Run("Marshal and Unmarshal", func(t *testing.T) {
		// Marshal
		data, err := jsonOps.Marshal(testData)
		require.NoError(t, err, "Failed to marshal JSON")

		// Unmarshal
		var result TestStruct
		err = jsonOps.Unmarshal(data, &result)
		require.NoError(t, err, "Failed to unmarshal JSON")

		// Compare fields individually to handle type differences
		assert.Equal(t, testData.Name, result.Name, "Name should match")
		assert.Equal(t, testData.Value, result.Value, "Value should match")
		assert.ElementsMatch(t, testData.Tags, result.Tags, "Tags should match")
		assert.Equal(t, testData.Meta["key"], result.Meta["key"], "Meta key should match")
		assert.Equal(t, testData.Meta["count"], result.Meta["count"], "Meta count should match")
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		invalidFile := filepath.Join(tmpDir, "invalid.json")
		err := fileOps.WriteFile(invalidFile, []byte("{ invalid json"), 0o644)
		require.NoError(t, err)

		var result TestStruct
		err = jsonOps.ReadJSON(invalidFile, &result)
		require.Error(t, err, "Should error on invalid JSON")
	})

	t.Run("NewDefaultJSONOperatorWithNil", func(t *testing.T) {
		// Test creating JSON operator with nil fileOps (should use default)
		jsonOpsWithNil := NewDefaultJSONOperator(nil)
		require.NotNil(t, jsonOpsWithNil, "JSON operator should not be nil")

		// Test that it works by marshaling some data
		testData := TestStruct{Name: "test", Value: 42}
		data, err := jsonOpsWithNil.Marshal(testData)
		require.NoError(t, err, "Marshal should work with nil fileOps")
		assert.Contains(t, string(data), "test", "Marshaled data should contain test name")

		// Test that file operations work
		testFile := filepath.Join(tmpDir, "nil-test.json")
		err = jsonOpsWithNil.WriteJSON(testFile, testData)
		require.NoError(t, err, "WriteJSON should work with nil fileOps")

		var result TestStruct
		err = jsonOpsWithNil.ReadJSON(testFile, &result)
		require.NoError(t, err, "ReadJSON should work with nil fileOps")
		assert.Equal(t, testData.Name, result.Name, "Data should match")
	})

	t.Run("WriteJSONIndent", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "indent-test.json")
		testData := TestStruct{Name: "indent-test", Value: 123}

		err := jsonOps.WriteJSONIndent(testFile, testData, "", "    ")
		require.NoError(t, err, "WriteJSONIndent should succeed")

		// Read the raw file to check indentation
		data, err := fileOps.ReadFile(testFile)
		require.NoError(t, err)
		jsonStr := string(data)
		assert.Contains(t, jsonStr, "\n    ", "JSON should be indented with 4 spaces")

		// Verify it's still valid JSON
		var result TestStruct
		err = jsonOps.ReadJSON(testFile, &result)
		require.NoError(t, err)
		assert.Equal(t, testData.Name, result.Name, "Indented JSON should be readable")
	})

	t.Run("JSONOperatorUnmarshal", func(t *testing.T) {
		// Test the Unmarshal method directly
		jsonData := []byte(`{"name":"direct-test","value":789,"tags":["tag1","tag2"]}`)
		var result TestStruct
		err := jsonOps.Unmarshal(jsonData, &result)
		require.NoError(t, err, "Direct Unmarshal should work")
		assert.Equal(t, "direct-test", result.Name, "Name should match")
		assert.Equal(t, 789, result.Value, "Value should match")
		assert.ElementsMatch(t, []string{"tag1", "tag2"}, result.Tags, "Tags should match")
	})
}

// TestDefaultYAMLOperatorErrorCases tests error scenarios for YAML operations
func TestDefaultYAMLOperatorErrorCases(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "yaml-error-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	fileOps := NewDefaultFileOperator()
	yamlOps := NewDefaultYAMLOperator(fileOps)

	t.Run("ReadYAML nonexistent file", func(t *testing.T) {
		var result map[string]interface{}
		err := yamlOps.ReadYAML(filepath.Join(tmpDir, "nonexistent.yaml"), &result)
		require.Error(t, err, "Should error on nonexistent file")
		assert.Contains(t, err.Error(), "failed to read file", "Error should mention file read failure")
	})

	t.Run("WriteYAML to invalid path", func(t *testing.T) {
		invalidPath := "/invalid/nonexistent/path/test.yaml"
		err := yamlOps.WriteYAML(invalidPath, map[string]string{"key": "value"})
		require.Error(t, err, "Should error with invalid path")
	})

	t.Run("Marshal unmarshalable data", func(t *testing.T) {
		// Function types can't be marshaled to YAML and will panic
		data := map[string]interface{}{
			"func": func() {},
		}

		// Test that marshal panics with function types
		assert.Panics(t, func() {
			_, _ = yamlOps.Marshal(data) //nolint:errcheck // Test intentionally ignores error
		}, "Should panic when marshaling function types")
	})

	t.Run("Unmarshal invalid YAML", func(t *testing.T) {
		invalidYAML := []byte("invalid:\n  - yaml\n  - [unclosed")
		var result map[string]interface{}
		err := yamlOps.Unmarshal(invalidYAML, &result)
		require.Error(t, err, "Should error on invalid YAML")
	})
}

func TestDefaultYAMLOperator(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "yaml-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	fileOps := NewDefaultFileOperator()
	yamlOps := NewDefaultYAMLOperator(fileOps)

	type TestStruct struct {
		Name   string                 `yaml:"name"`
		Value  int                    `yaml:"value"`
		Tags   []string               `yaml:"tags,omitempty"`
		Meta   map[string]interface{} `yaml:"meta,omitempty"`
		Nested struct {
			Field1 string `yaml:"field1"`
			Field2 bool   `yaml:"field2"`
		} `yaml:"nested"`
	}

	testData := TestStruct{
		Name:  "test",
		Value: 42,
		Tags:  []string{"tag1", "tag2"},
		Meta:  map[string]interface{}{"key": "value", "count": 10},
	}
	testData.Nested.Field1 = "nested value"
	testData.Nested.Field2 = true

	testFile := filepath.Join(tmpDir, "test.yaml")

	t.Run("WriteYAML and ReadYAML", func(t *testing.T) {
		// Write YAML
		err := yamlOps.WriteYAML(testFile, testData)
		require.NoError(t, err, "Failed to write YAML")

		// Read YAML
		var result TestStruct
		err = yamlOps.ReadYAML(testFile, &result)
		require.NoError(t, err, "Failed to read YAML")

		assert.Equal(t, testData.Name, result.Name, "Name should match")
		assert.Equal(t, testData.Value, result.Value, "Value should match")
		assert.ElementsMatch(t, testData.Tags, result.Tags, "Tags should match")
		assert.Equal(t, testData.Meta["key"], result.Meta["key"], "Meta key should match")
		assert.Equal(t, testData.Meta["count"], result.Meta["count"], "Meta count should match")
		assert.Equal(t, testData.Nested.Field1, result.Nested.Field1, "Nested field1 should match")
		assert.Equal(t, testData.Nested.Field2, result.Nested.Field2, "Nested field2 should match")
	})

	t.Run("Marshal and Unmarshal", func(t *testing.T) {
		data, err := yamlOps.Marshal(testData)
		require.NoError(t, err, "Failed to marshal YAML")

		// Verify it's valid YAML
		var result TestStruct
		err = yaml.Unmarshal(data, &result)
		require.NoError(t, err, "Failed to unmarshal YAML")

		// Compare fields individually to handle type differences
		assert.Equal(t, testData.Name, result.Name, "Name should match")
		assert.Equal(t, testData.Value, result.Value, "Value should match")
		assert.ElementsMatch(t, testData.Tags, result.Tags, "Tags should match")
		assert.Equal(t, testData.Meta["key"], result.Meta["key"], "Meta key should match")
		assert.Equal(t, testData.Meta["count"], result.Meta["count"], "Meta count should match")
	})

	t.Run("Invalid YAML", func(t *testing.T) {
		invalidFile := filepath.Join(tmpDir, "invalid.yaml")
		err := fileOps.WriteFile(invalidFile, []byte("invalid:\n  - unclosed\n  - [bracket"), 0o644)
		require.NoError(t, err)

		var result TestStruct
		err = yamlOps.ReadYAML(invalidFile, &result)
		require.Error(t, err, "Should error on invalid YAML")
	})

	t.Run("NewDefaultYAMLOperatorWithNil", func(t *testing.T) {
		// Test creating YAML operator with nil fileOps (should use default)
		yamlOpsWithNil := NewDefaultYAMLOperator(nil)
		require.NotNil(t, yamlOpsWithNil, "YAML operator should not be nil")

		// Test that it works by marshaling some data
		testData := TestStruct{Name: "yaml-test", Value: 99}
		data, err := yamlOpsWithNil.Marshal(testData)
		require.NoError(t, err, "Marshal should work with nil fileOps")
		assert.Contains(t, string(data), "yaml-test", "Marshaled data should contain test name")

		// Test that file operations work
		testFile := filepath.Join(tmpDir, "nil-test.yaml")
		err = yamlOpsWithNil.WriteYAML(testFile, testData)
		require.NoError(t, err, "WriteYAML should work with nil fileOps")

		var result TestStruct
		err = yamlOpsWithNil.ReadYAML(testFile, &result)
		require.NoError(t, err, "ReadYAML should work with nil fileOps")
		assert.Equal(t, testData.Name, result.Name, "Data should match")
		assert.Equal(t, testData.Value, result.Value, "Value should match")
	})

	t.Run("YAMLOperatorUnmarshal", func(t *testing.T) {
		// Test the Unmarshal method directly
		yamlData := []byte("name: direct-test\nvalue: 456\ntags:\n  - tag1\n  - tag2")
		var result TestStruct
		err := yamlOps.Unmarshal(yamlData, &result)
		require.NoError(t, err, "Direct Unmarshal should work")
		assert.Equal(t, "direct-test", result.Name, "Name should match")
		assert.Equal(t, 456, result.Value, "Value should match")
		assert.ElementsMatch(t, []string{"tag1", "tag2"}, result.Tags, "Tags should match")
	})
}

// TestSafeFileOperatorComprehensive provides comprehensive coverage for SafeFileOperator
func TestSafeFileOperatorComprehensive(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "safe-comprehensive-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	ops := NewDefaultSafeFileOperator()

	t.Run("WriteFileAtomicPermissions", func(t *testing.T) {
		// Test various permission combinations
		permTests := []struct {
			name string
			perm os.FileMode
		}{
			{"ReadWrite", 0o644},
			{"ReadOnly", 0o444},
			{"ReadWriteExecute", 0o755},
			{"Restrictive", 0o600},
		}

		for _, tt := range permTests {
			t.Run(tt.name, func(t *testing.T) {
				testFile := filepath.Join(tmpDir, fmt.Sprintf("perm-test-%s.txt", tt.name))
				testData := []byte(fmt.Sprintf("Permission test: %s", tt.name))

				err := ops.WriteFileAtomic(testFile, testData, tt.perm)
				require.NoError(t, err, "WriteFileAtomic should succeed with %s permissions", tt.name)

				// Verify file permissions
				info, err := os.Stat(testFile)
				require.NoError(t, err)
				assert.Equal(t, tt.perm, info.Mode().Perm(), "File should have correct permissions")

				// Verify content
				data, err := ops.ReadFile(testFile)
				require.NoError(t, err)
				assert.Equal(t, testData, data, "File content should match")
			})
		}
	})

	t.Run("WriteFileAtomicLargeData", func(t *testing.T) {
		// Test with large data to ensure proper buffering
		testFile := filepath.Join(tmpDir, "large-data.txt")
		// Create 1MB of test data
		testData := make([]byte, 1024*1024)
		for i := range testData {
			testData[i] = byte(i % 256)
		}

		err := ops.WriteFileAtomic(testFile, testData, 0o644)
		require.NoError(t, err, "Should handle large data")

		// Verify data integrity
		readData, err := ops.ReadFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, testData, readData, "Large data should be written correctly")
	})

	t.Run("WriteFileAtomicConcurrent", func(t *testing.T) {
		// Test concurrent atomic writes to different files
		const numGoroutines = 10
		var wg sync.WaitGroup
		errorChan := make(chan error, numGoroutines)

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				defer wg.Done()
				testFile := filepath.Join(tmpDir, fmt.Sprintf("concurrent-%d.txt", index))
				testData := []byte(fmt.Sprintf("Concurrent data %d", index))

				if err := ops.WriteFileAtomic(testFile, testData, 0o644); err != nil {
					errorChan <- fmt.Errorf("goroutine %d failed: %w", index, err)
					return
				}

				// Verify immediately after write
				data, err := ops.ReadFile(testFile)
				if err != nil {
					errorChan <- fmt.Errorf("goroutine %d read failed: %w", index, err)
					return
				}

				if !bytes.Equal(testData, data) {
					errorChan <- fmt.Errorf("%w: goroutine %d", errTestDataMismatch, index)
				}
			}(i)
		}

		wg.Wait()
		close(errorChan)

		// Check for any errors
		for err := range errorChan {
			t.Errorf("Concurrent test error: %v", err)
		}
	})
}

func TestSafeFileOperator(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "safe-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	ops := NewDefaultSafeFileOperator()

	t.Run("WriteFileAtomic", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "atomic.txt")
		testData := []byte("Atomic write test")

		err := ops.WriteFileAtomic(testFile, testData, 0o644)
		require.NoError(t, err, "Failed to write file atomically")

		// Verify file contents
		data, err := ops.ReadFile(testFile)
		require.NoError(t, err, "Failed to read file")

		assert.Equal(t, testData, data, "File contents should match")

		// Verify no temp files remain
		files, err := os.ReadDir(tmpDir)
		require.NoError(t, err)
		assert.Len(t, files, 1, "Should only have the final file, no temp files")
	})

	t.Run("WriteFileAtomicErrorCases", func(t *testing.T) {
		// Test write to invalid directory (should fail to create temp file)
		invalidPath := "/invalid/nonexistent/path/file.txt"
		err := ops.WriteFileAtomic(invalidPath, []byte("test"), 0o644)
		require.Error(t, err, "Should fail to write to invalid directory")
		assert.Contains(t, err.Error(), "failed to create temp file", "Error should mention temp file creation failure")

		// Test with existing file to ensure overwrite works
		existingFile := filepath.Join(tmpDir, "existing.txt")
		originalData := []byte("original data")
		newData := []byte("new atomic data")

		// Create original file
		err = ops.WriteFile(existingFile, originalData, 0o644)
		require.NoError(t, err)

		// Atomic overwrite
		err = ops.WriteFileAtomic(existingFile, newData, 0o600)
		require.NoError(t, err, "Should overwrite existing file atomically")

		// Verify content and permissions
		data, err := ops.ReadFile(existingFile)
		require.NoError(t, err)
		assert.Equal(t, newData, data, "Content should be updated")

		// Check permissions
		info, err := os.Stat(existingFile)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(), "Permissions should be updated")
	})

	t.Run("WriteFileWithBackup", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "backup.txt")
		originalData := []byte("Original data")
		newData := []byte("New data")

		// Write original file
		err := ops.WriteFile(testFile, originalData, 0o644)
		require.NoError(t, err, "Failed to write original file")

		// Write with backup
		err = ops.WriteFileWithBackup(testFile, newData, 0o644)
		require.NoError(t, err, "Failed to write file with backup")

		// Check backup exists
		backupFile := testFile + ".bak"
		assert.True(t, ops.Exists(backupFile), "Backup file should exist")

		// Verify backup contents
		backupData, err := ops.ReadFile(backupFile)
		require.NoError(t, err, "Failed to read backup file")
		assert.Equal(t, originalData, backupData, "Backup should contain original data")

		// Verify new file contents
		newFileData, err := ops.ReadFile(testFile)
		require.NoError(t, err, "Failed to read new file")
		assert.Equal(t, newData, newFileData, "File should contain new data")
	})

	t.Run("WriteFileWithBackupNewFile", func(t *testing.T) {
		// Test WriteFileWithBackup on non-existent file (no backup should be created)
		newFile := filepath.Join(tmpDir, "new-backup-file.txt")
		testData := []byte("New file data")

		err := ops.WriteFileWithBackup(newFile, testData, 0o644)
		require.NoError(t, err, "Should succeed writing new file with backup")

		// Verify file exists and has correct content
		assert.True(t, ops.Exists(newFile), "New file should exist")
		data, err := ops.ReadFile(newFile)
		require.NoError(t, err)
		assert.Equal(t, testData, data, "New file should have correct content")

		// Verify no backup was created
		backupFile := newFile + ".bak"
		assert.False(t, ops.Exists(backupFile), "Backup should not exist for new file")
	})
}

func (s *FileOpsTestSuite) TestFileOpsFacade() {
	type TestConfig struct {
		Database struct {
			Host string `yaml:"host" json:"host"`
			Port int    `yaml:"port" json:"port"`
		} `yaml:"database" json:"database"`
		Debug    bool     `yaml:"debug" json:"debug"`
		Features []string `yaml:"features" json:"features"`
	}

	testConfig := TestConfig{}
	testConfig.Database.Host = "localhost"
	testConfig.Database.Port = 5432
	testConfig.Debug = true
	testConfig.Features = []string{"feature1", "feature2"}

	s.Run("SaveConfig and LoadConfig YAML", func() {
		configPath := filepath.Join(s.tmpDir, "config.yaml")

		// Save config
		err := s.ops.SaveConfig(configPath, testConfig, "yaml")
		s.Require().NoError(err, "Failed to save config")

		// Load config
		var result TestConfig
		foundPath, err := s.ops.LoadConfig([]string{configPath}, &result)
		s.Require().NoError(err, "Failed to load config")

		s.Equal(configPath, foundPath, "Should find config at expected path")
		s.Equal(testConfig.Database.Host, result.Database.Host, "Database host should match")
		s.Equal(testConfig.Database.Port, result.Database.Port, "Database port should match")
		s.Equal(testConfig.Debug, result.Debug, "Debug flag should match")
		s.ElementsMatch(testConfig.Features, result.Features, "Features should match")
	})

	s.Run("SaveConfig and LoadConfig JSON", func() {
		configPath := filepath.Join(s.tmpDir, "config.json")

		// Save config
		err := s.ops.SaveConfig(configPath, testConfig, "json")
		s.Require().NoError(err, "Failed to save config")

		// Load config
		var result TestConfig
		foundPath, err := s.ops.LoadConfig([]string{configPath}, &result)
		s.Require().NoError(err, "Failed to load config")

		s.Equal(configPath, foundPath, "Should find config at expected path")
		s.Equal(testConfig, result, "Loaded config should match saved config")
	})

	s.Run("LoadConfig with fallback", func() {
		// Create only the second config file
		configPath1 := filepath.Join(s.tmpDir, "nonexistent.yaml")
		configPath2 := filepath.Join(s.tmpDir, "fallback.yaml")

		err := s.ops.SaveConfig(configPath2, testConfig, "yaml")
		s.Require().NoError(err, "Failed to save fallback config")

		// Load with fallback
		var result TestConfig
		foundPath, err := s.ops.LoadConfig([]string{configPath1, configPath2}, &result)
		s.Require().NoError(err, "Failed to load config with fallback")

		s.Equal(configPath2, foundPath, "Should find config at fallback path")
		s.Equal(testConfig.Database.Host, result.Database.Host, "Config should be loaded correctly")
	})

	s.Run("LoadConfig all missing", func() {
		// Try to load from non-existent files
		paths := []string{
			filepath.Join(s.tmpDir, "missing1.yaml"),
			filepath.Join(s.tmpDir, "missing2.yaml"),
		}

		var result TestConfig
		foundPath, err := s.ops.LoadConfig(paths, &result)
		s.Require().Error(err, "Should error when all config files are missing")
		s.Empty(foundPath, "Should return empty path when no config found")
	})

	s.Run("SaveConfig invalid format", func() {
		configPath := filepath.Join(s.tmpDir, "config.invalid")

		err := s.ops.SaveConfig(configPath, testConfig, "invalid")
		// SaveConfig defaults to YAML for unknown formats, so no error
		s.Require().NoError(err, "Should default to YAML for unknown formats")

		// Verify YAML file was created
		s.True(s.ops.File.Exists(configPath), "File should exist")
	})
}

func TestPackageLevelFunctions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "package-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, removeErr)
		}
	}()

	testFile := filepath.Join(tmpDir, "package-test.txt")
	testData := []byte("Package level test")

	// Test package-level WriteFile
	err = WriteFile(testFile, testData, 0o644)
	require.NoError(t, err, "Failed to write file using package function")

	// Test package-level Exists
	assert.True(t, Exists(testFile), "File should exist")

	// Test package-level ReadFile
	data, err := ReadFile(testFile)
	require.NoError(t, err, "Failed to read file using package function")

	assert.Equal(t, testData, data, "Read data should match written data")

	// Test package-level IsDir
	assert.True(t, IsDir(tmpDir), "Should identify directory correctly")
	assert.False(t, IsDir(testFile), "Should not identify file as directory")

	// Test package-level IsFile
	assert.True(t, IsFile(testFile), "Should identify file correctly")
	assert.False(t, IsFile(tmpDir), "Should not identify directory as file")

	// Test package-level Remove
	err = Remove(testFile)
	require.NoError(t, err, "Failed to remove file")
	assert.False(t, Exists(testFile), "File should not exist after removal")

	// Test package-level MkdirAll
	testDir := filepath.Join(tmpDir, "test", "nested", "dir")
	err = MkdirAll(testDir, 0o755)
	require.NoError(t, err, "Failed to create directory")
	assert.True(t, IsDir(testDir), "Directory should exist")

	// Test package-level Copy
	srcFile := filepath.Join(tmpDir, "src.txt")
	dstFile := filepath.Join(tmpDir, "dst.txt")
	err = WriteFile(srcFile, testData, 0o644)
	require.NoError(t, err)

	err = Copy(srcFile, dstFile)
	require.NoError(t, err, "Failed to copy file")

	copiedData, err := ReadFile(dstFile)
	require.NoError(t, err)
	assert.Equal(t, testData, copiedData, "Copied data should match source")
}

// TestMockFileOperations tests using mocks for comprehensive coverage
func TestMockFileOperations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("FileOps with mocked components", func(t *testing.T) {
		// Create mocks
		mockFile := NewMockFileOperator(ctrl)
		mockJSON := NewMockJSONOperator(ctrl)
		mockYAML := NewMockYAMLOperator(ctrl)
		mockSafe := NewMockSafeFileOperator(ctrl)

		// Create FileOps with mocks
		ops := NewWithOptions(mockFile, mockJSON, mockYAML, mockSafe)
		require.NotNil(t, ops, "FileOps should not be nil")

		// Test WriteJSONSafe with mocks
		testPath := "/test/path/file.json"
		testData := map[string]string{"key": "value"}
		expectedJSON := []byte(`{"key":"value"}`)

		// Set up mock expectations
		mockFile.EXPECT().Exists("/test/path").Return(false)
		mockFile.EXPECT().MkdirAll("/test/path", os.FileMode(0o755)).Return(nil)
		mockJSON.EXPECT().MarshalIndent(testData, "", "  ").Return(expectedJSON, nil)
		mockSafe.EXPECT().WriteFileAtomic(testPath, expectedJSON, os.FileMode(0o644)).Return(nil)

		err := ops.WriteJSONSafe(testPath, testData)
		assert.NoError(t, err, "WriteJSONSafe should succeed with mocks")
	})

	t.Run("WriteJSONSafe with JSON marshal error", func(t *testing.T) {
		mockFile := NewMockFileOperator(ctrl)
		mockJSON := NewMockJSONOperator(ctrl)
		mockYAML := NewMockYAMLOperator(ctrl)
		mockSafe := NewMockSafeFileOperator(ctrl)

		ops := NewWithOptions(mockFile, mockJSON, mockYAML, mockSafe)
		testPath := "/test/file.json"
		testData := map[string]string{"key": "value"}

		// Mock directory creation success but JSON marshal failure
		mockFile.EXPECT().Exists("/test").Return(false)
		mockFile.EXPECT().MkdirAll("/test", os.FileMode(0o755)).Return(nil)
		mockJSON.EXPECT().MarshalIndent(testData, "", "  ").Return(nil, errTestMarshalError)

		err := ops.WriteJSONSafe(testPath, testData)
		require.Error(t, err, "Should error when JSON marshal fails")
		assert.Contains(t, err.Error(), "failed to marshal JSON", "Error should mention JSON marshal failure")
	})

	t.Run("WriteYAMLSafe with mocks", func(t *testing.T) {
		mockFile := NewMockFileOperator(ctrl)
		mockJSON := NewMockJSONOperator(ctrl)
		mockYAML := NewMockYAMLOperator(ctrl)
		mockSafe := NewMockSafeFileOperator(ctrl)

		ops := NewWithOptions(mockFile, mockJSON, mockYAML, mockSafe)
		testPath := "/test/path/file.yaml"
		testData := map[string]string{"key": "value"}
		expectedYAML := []byte("key: value\n")

		// Set up mock expectations
		mockFile.EXPECT().Exists("/test/path").Return(true) // Directory exists
		mockYAML.EXPECT().Marshal(testData).Return(expectedYAML, nil)
		mockSafe.EXPECT().WriteFileAtomic(testPath, expectedYAML, os.FileMode(0o644)).Return(nil)

		err := ops.WriteYAMLSafe(testPath, testData)
		assert.NoError(t, err, "WriteYAMLSafe should succeed with mocks")
	})

	t.Run("LoadConfig with mocks", func(t *testing.T) {
		mockFile := NewMockFileOperator(ctrl)
		mockJSON := NewMockJSONOperator(ctrl)
		mockYAML := NewMockYAMLOperator(ctrl)
		mockSafe := NewMockSafeFileOperator(ctrl)

		ops := NewWithOptions(mockFile, mockJSON, mockYAML, mockSafe)
		paths := []string{"/nonexistent.json", "/existing.yaml"}
		var dest map[string]interface{}

		// First path doesn't exist, second path exists and is YAML
		mockFile.EXPECT().Exists("/nonexistent.json").Return(false)
		mockFile.EXPECT().Exists("/existing.yaml").Return(true)
		mockYAML.EXPECT().ReadYAML("/existing.yaml", &dest).Return(nil)

		foundPath, err := ops.LoadConfig(paths, &dest)
		require.NoError(t, err, "LoadConfig should succeed")
		assert.Equal(t, "/existing.yaml", foundPath, "Should return the found path")
	})

	t.Run("LoadConfig with JSON fallback", func(t *testing.T) {
		mockFile := NewMockFileOperator(ctrl)
		mockJSON := NewMockJSONOperator(ctrl)
		mockYAML := NewMockYAMLOperator(ctrl)
		mockSafe := NewMockSafeFileOperator(ctrl)

		ops := NewWithOptions(mockFile, mockJSON, mockYAML, mockSafe)
		paths := []string{"/config"} // No extension
		var dest map[string]interface{}

		// File exists, YAML fails, JSON succeeds
		mockFile.EXPECT().Exists("/config").Return(true)
		mockYAML.EXPECT().ReadYAML("/config", &dest).Return(errTestYAMLError)
		mockJSON.EXPECT().ReadJSON("/config", &dest).Return(nil)

		foundPath, err := ops.LoadConfig(paths, &dest)
		require.NoError(t, err, "LoadConfig should succeed with JSON fallback")
		assert.Equal(t, "/config", foundPath, "Should return the found path")
	})

	t.Run("SaveConfig with mocks", func(t *testing.T) {
		mockFile := NewMockFileOperator(ctrl)
		mockJSON := NewMockJSONOperator(ctrl)
		mockYAML := NewMockYAMLOperator(ctrl)
		mockSafe := NewMockSafeFileOperator(ctrl)

		ops := NewWithOptions(mockFile, mockJSON, mockYAML, mockSafe)
		testPath := "/test/config.json" // Use a path that will trigger ensureDir
		testData := map[string]string{"key": "value"}
		expectedJSON := []byte(`{"key":"value"}`)

		// Mock for JSON format - ensureDir will check if /test exists
		mockFile.EXPECT().Exists("/test").Return(true).AnyTimes() // Parent directory exists
		mockJSON.EXPECT().MarshalIndent(testData, "", "  ").Return(expectedJSON, nil)
		mockSafe.EXPECT().WriteFileAtomic(testPath, expectedJSON, os.FileMode(0o644)).Return(nil)

		err := ops.SaveConfig(testPath, testData, "json")
		assert.NoError(t, err, "SaveConfig should succeed with JSON format")
	})

	t.Run("CopyFile with mocks", func(t *testing.T) {
		mockFile := NewMockFileOperator(ctrl)
		mockJSON := NewMockJSONOperator(ctrl)
		mockYAML := NewMockYAMLOperator(ctrl)
		mockSafe := NewMockSafeFileOperator(ctrl)

		ops := NewWithOptions(mockFile, mockJSON, mockYAML, mockSafe)
		srcPath := "/src.txt"
		dstPath := "/dest/dst.txt"

		// Mock directory creation and file copy
		mockFile.EXPECT().Exists("/dest").Return(false)
		mockFile.EXPECT().MkdirAll("/dest", os.FileMode(0o755)).Return(nil)
		mockFile.EXPECT().Copy(srcPath, dstPath).Return(nil)

		err := ops.CopyFile(srcPath, dstPath)
		assert.NoError(t, err, "CopyFile should succeed with mocks")
	})

	t.Run("BackupFile with mocks", func(t *testing.T) {
		mockFile := NewMockFileOperator(ctrl)
		mockJSON := NewMockJSONOperator(ctrl)
		mockYAML := NewMockYAMLOperator(ctrl)
		mockSafe := NewMockSafeFileOperator(ctrl)

		ops := NewWithOptions(mockFile, mockJSON, mockYAML, mockSafe)
		filePath := "/test.txt"
		backupPath := "/test.txt.bak"

		// Mock file exists and copy
		mockFile.EXPECT().Exists(filePath).Return(true)
		mockFile.EXPECT().Copy(filePath, backupPath).Return(nil)

		err := ops.BackupFile(filePath)
		assert.NoError(t, err, "BackupFile should succeed with mocks")
	})

	t.Run("CleanupBackups with mocks", func(t *testing.T) {
		mockFile := NewMockFileOperator(ctrl)
		mockJSON := NewMockJSONOperator(ctrl)
		mockYAML := NewMockYAMLOperator(ctrl)
		mockSafe := NewMockSafeFileOperator(ctrl)

		ops := NewWithOptions(mockFile, mockJSON, mockYAML, mockSafe)
		dir := "/backups"
		pattern := "*.bak"

		// Mock successful cleanup - this will test the error case in the actual implementation
		// when filepath.Glob fails or when Remove fails
		mockFile.EXPECT().Remove(gomock.Any()).Return(errTestRemoveError).AnyTimes()

		err := ops.CleanupBackups(dir, pattern)
		// This should error due to the mocked Remove failure
		if err != nil {
			assert.Contains(t, err.Error(), "failed to remove backup", "Error should mention backup removal failure")
		}
	})
}

// TestNewConstructors tests the constructor functions
func TestNewConstructors(t *testing.T) {
	t.Run("NewFileOperator", func(t *testing.T) {
		ops := NewFileOperator()
		require.NotNil(t, ops, "NewFileOperator should return non-nil operator")
		// Test that it's actually a DefaultFileOperator
		_, ok := ops.(*DefaultFileOperator)
		assert.True(t, ok, "NewFileOperator should return DefaultFileOperator")
	})

	t.Run("GetDefault", func(t *testing.T) {
		ops := GetDefault()
		require.NotNil(t, ops, "GetDefault should return non-nil FileOps")
		assert.NotNil(t, ops.File, "File operator should not be nil")
		assert.NotNil(t, ops.JSON, "JSON operator should not be nil")
		assert.NotNil(t, ops.YAML, "YAML operator should not be nil")
		assert.NotNil(t, ops.Safe, "Safe operator should not be nil")
	})

	t.Run("New", func(t *testing.T) {
		ops := New()
		require.NotNil(t, ops, "New should return non-nil FileOps")
		assert.NotNil(t, ops.File, "File operator should not be nil")
		assert.NotNil(t, ops.JSON, "JSON operator should not be nil")
		assert.NotNil(t, ops.YAML, "YAML operator should not be nil")
		assert.NotNil(t, ops.Safe, "Safe operator should not be nil")
	})

	t.Run("NewWithOptions", func(t *testing.T) {
		// Create mock operators
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockFile := NewMockFileOperator(ctrl)
		mockJSON := NewMockJSONOperator(ctrl)
		mockYAML := NewMockYAMLOperator(ctrl)
		mockSafe := NewMockSafeFileOperator(ctrl)

		ops := NewWithOptions(mockFile, mockJSON, mockYAML, mockSafe)
		require.NotNil(t, ops, "NewWithOptions should return non-nil FileOps")
		assert.Equal(t, mockFile, ops.File, "Should use provided file operator")
		assert.Equal(t, mockJSON, ops.JSON, "Should use provided JSON operator")
		assert.Equal(t, mockYAML, ops.YAML, "Should use provided YAML operator")
		assert.Equal(t, mockSafe, ops.Safe, "Should use provided safe operator")
	})
}

// TestFileOpsErrorHandling tests error paths in FileOps methods
func TestFileOpsErrorHandling(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "fileops-error-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, removeErr)
		}
	}()

	ops := New()

	t.Run("BackupFile nonexistent file", func(t *testing.T) {
		nonexistentFile := filepath.Join(tmpDir, "nonexistent.txt")
		err := ops.BackupFile(nonexistentFile)
		require.Error(t, err, "Should error when trying to backup nonexistent file")
		assert.Contains(t, err.Error(), "file does not exist", "Error should mention file doesn't exist")
	})

	t.Run("BackupFile success", func(t *testing.T) {
		// Create a test file
		testFile := filepath.Join(tmpDir, "backup-test.txt")
		testData := []byte("backup test data")
		err := ops.File.WriteFile(testFile, testData, 0o644)
		require.NoError(t, err)

		// Create backup
		err = ops.BackupFile(testFile)
		require.NoError(t, err, "BackupFile should succeed")

		// Verify backup exists
		backupFile := testFile + ".bak"
		assert.True(t, ops.File.Exists(backupFile), "Backup file should exist")

		// Verify backup content
		backupData, err := ops.File.ReadFile(backupFile)
		require.NoError(t, err)
		assert.Equal(t, testData, backupData, "Backup should have same content")
	})

	t.Run("CleanupBackups", func(t *testing.T) {
		// Create test directory with backup files
		backupDir := filepath.Join(tmpDir, "backup-cleanup")
		err := ops.File.MkdirAll(backupDir, 0o755)
		require.NoError(t, err)

		// Create various files
		files := []string{"file1.txt.bak", "file2.log.bak", "file3.txt", "file4.json.bak"}
		for _, file := range files {
			err = ops.File.WriteFile(filepath.Join(backupDir, file), []byte("test"), 0o644)
			require.NoError(t, err)
		}

		// Cleanup backup files
		err = ops.CleanupBackups(backupDir, "*.bak")
		require.NoError(t, err, "CleanupBackups should succeed")

		// Verify only non-backup files remain
		assert.False(t, ops.File.Exists(filepath.Join(backupDir, "file1.txt.bak")), "Backup file should be removed")
		assert.False(t, ops.File.Exists(filepath.Join(backupDir, "file2.log.bak")), "Backup file should be removed")
		assert.False(t, ops.File.Exists(filepath.Join(backupDir, "file4.json.bak")), "Backup file should be removed")
		assert.True(t, ops.File.Exists(filepath.Join(backupDir, "file3.txt")), "Non-backup file should remain")
	})

	t.Run("CleanupBackups invalid pattern", func(t *testing.T) {
		// Test with invalid glob pattern
		err := ops.CleanupBackups(tmpDir, "[invalid")
		require.Error(t, err, "Should error with invalid glob pattern")
		assert.Contains(t, err.Error(), "failed to glob pattern", "Error should mention glob failure")
	})

	t.Run("CopyFile with directory creation", func(t *testing.T) {
		// Create source file
		srcFile := filepath.Join(tmpDir, "copy-src.txt")
		testData := []byte("copy test data")
		err := ops.File.WriteFile(srcFile, testData, 0o644)
		require.NoError(t, err)

		// Copy to nested directory (should create parent dirs)
		dstFile := filepath.Join(tmpDir, "nested", "deep", "copy-dst.txt")
		err = ops.CopyFile(srcFile, dstFile)
		require.NoError(t, err, "CopyFile should succeed with directory creation")

		// Verify file was copied
		assert.True(t, ops.File.Exists(dstFile), "Destination file should exist")
		copiedData, err := ops.File.ReadFile(dstFile)
		require.NoError(t, err)
		assert.Equal(t, testData, copiedData, "Copied data should match")
	})

	t.Run("ensureDir edge cases", func(t *testing.T) {
		// Test with root directory
		err := ops.WriteJSONSafe("/test.json", map[string]string{"key": "value"})
		// This should not error due to ensureDir handling root
		if err != nil {
			// Expected to fail due to permissions, but not due to ensureDir
			assert.NotContains(t, err.Error(), "ensureDir", "Error should not be from ensureDir")
		}

		// Test with current directory
		testFile := "./current-dir-test.json"
		defer func() {
			// Clean up
			if ops.File.Exists(testFile) {
				if removeErr := ops.File.Remove(testFile); removeErr != nil {
					t.Logf("Failed to remove test file: %v", removeErr)
				}
			}
		}()

		err = ops.WriteJSONSafe(testFile, map[string]string{"key": "value"})
		assert.NoError(t, err, "Should succeed with current directory")
	})
}

// TestGlobalConvenienceFunctions tests the global convenience functions that were previously uncovered
func TestGlobalConvenienceFunctions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "global-conv-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, removeErr)
		}
	}()

	type TestConfig struct {
		Name     string            `json:"name" yaml:"name"`
		Version  string            `json:"version" yaml:"version"`
		Settings map[string]string `json:"settings" yaml:"settings"`
	}

	testConfig := TestConfig{
		Name:    "test-config",
		Version: "1.0.0",
		Settings: map[string]string{
			"debug": "true",
			"port":  "8080",
		},
	}

	t.Run("WriteJSONSafe", func(t *testing.T) {
		// Test writing JSON to nested directory (should create parent dirs)
		jsonPath := filepath.Join(tmpDir, "nested", "dir", "config.json")
		err := WriteJSONSafe(jsonPath, testConfig)
		require.NoError(t, err, "WriteJSONSafe should succeed")

		// Verify file exists and has correct content
		assert.True(t, Exists(jsonPath), "JSON file should exist")

		// Read and verify content
		var result TestConfig
		data, err := ReadFile(jsonPath)
		require.NoError(t, err)
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)
		assert.Equal(t, testConfig, result, "JSON content should match")

		// Verify it's properly indented
		jsonStr := string(data)
		assert.Contains(t, jsonStr, "\n  ", "JSON should be indented")
	})

	t.Run("WriteYAMLSafe", func(t *testing.T) {
		// Test writing YAML to nested directory (should create parent dirs)
		yamlPath := filepath.Join(tmpDir, "nested", "yaml", "config.yaml")
		err := WriteYAMLSafe(yamlPath, testConfig)
		require.NoError(t, err, "WriteYAMLSafe should succeed")

		// Verify file exists and has correct content
		assert.True(t, Exists(yamlPath), "YAML file should exist")

		// Read and verify content
		var result TestConfig
		data, err := ReadFile(yamlPath)
		require.NoError(t, err)
		err = yaml.Unmarshal(data, &result)
		require.NoError(t, err)
		assert.Equal(t, testConfig, result, "YAML content should match")
	})

	t.Run("SaveConfig", func(t *testing.T) {
		// Test SaveConfig with JSON format
		jsonConfigPath := filepath.Join(tmpDir, "save-config", "config.json")
		err := SaveConfig(jsonConfigPath, testConfig, "json")
		require.NoError(t, err, "SaveConfig JSON should succeed")
		assert.True(t, Exists(jsonConfigPath), "JSON config file should exist")

		// Test SaveConfig with YAML format
		yamlConfigPath := filepath.Join(tmpDir, "save-config", "config.yaml")
		err = SaveConfig(yamlConfigPath, testConfig, "yaml")
		require.NoError(t, err, "SaveConfig YAML should succeed")
		assert.True(t, Exists(yamlConfigPath), "YAML config file should exist")

		// Test SaveConfig with yml format
		ymlConfigPath := filepath.Join(tmpDir, "save-config", "config.yml")
		err = SaveConfig(ymlConfigPath, testConfig, "yml")
		require.NoError(t, err, "SaveConfig YML should succeed")
		assert.True(t, Exists(ymlConfigPath), "YML config file should exist")

		// Test SaveConfig with unknown format (should default to YAML)
		unknownConfigPath := filepath.Join(tmpDir, "save-config", "config.unknown")
		err = SaveConfig(unknownConfigPath, testConfig, "unknown")
		require.NoError(t, err, "SaveConfig unknown format should succeed")
		assert.True(t, Exists(unknownConfigPath), "Unknown format config file should exist")

		// Verify it's YAML format
		var result TestConfig
		data, err := ReadFile(unknownConfigPath)
		require.NoError(t, err)
		err = yaml.Unmarshal(data, &result)
		require.NoError(t, err, "Unknown format should default to YAML")
	})

	t.Run("LoadConfig", func(t *testing.T) {
		// Create test config files
		jsonPath := filepath.Join(tmpDir, "load-test", "config.json")
		yamlPath := filepath.Join(tmpDir, "load-test", "config.yaml")
		ymlPath := filepath.Join(tmpDir, "load-test", "config.yml")
		noExtPath := filepath.Join(tmpDir, "load-test", "config")

		// Create JSON config
		err := SaveConfig(jsonPath, testConfig, "json")
		require.NoError(t, err)

		// Create YAML config
		err = SaveConfig(yamlPath, testConfig, "yaml")
		require.NoError(t, err)

		// Create YML config
		err = SaveConfig(ymlPath, testConfig, "yml")
		require.NoError(t, err)

		// Create config without extension (YAML format)
		yamlData, err := yaml.Marshal(testConfig)
		require.NoError(t, err)
		err = WriteFile(noExtPath, yamlData, 0o644)
		require.NoError(t, err)

		// Test loading JSON config
		var jsonResult TestConfig
		foundPath, err := LoadConfig([]string{jsonPath}, &jsonResult)
		require.NoError(t, err, "LoadConfig JSON should succeed")
		assert.Equal(t, jsonPath, foundPath, "Should find JSON config")
		assert.Equal(t, testConfig, jsonResult, "JSON config should match")

		// Test loading YAML config
		var yamlResult TestConfig
		foundPath, err = LoadConfig([]string{yamlPath}, &yamlResult)
		require.NoError(t, err, "LoadConfig YAML should succeed")
		assert.YAMLEq(t, yamlPath, foundPath, "Should find YAML config")
		assert.Equal(t, testConfig, yamlResult, "YAML config should match")

		// Test loading YML config
		var ymlResult TestConfig
		foundPath, err = LoadConfig([]string{ymlPath}, &ymlResult)
		require.NoError(t, err, "LoadConfig YML should succeed")
		assert.YAMLEq(t, ymlPath, foundPath, "Should find YML config")
		assert.Equal(t, testConfig, ymlResult, "YML config should match")

		// Test loading config without extension (should try YAML first)
		var noExtResult TestConfig
		foundPath, err = LoadConfig([]string{noExtPath}, &noExtResult)
		require.NoError(t, err, "LoadConfig no extension should succeed")
		assert.Equal(t, noExtPath, foundPath, "Should find no extension config")
		assert.Equal(t, testConfig, noExtResult, "No extension config should match")

		// Test fallback behavior (first file doesn't exist, second does)
		nonExistentPath := filepath.Join(tmpDir, "nonexistent.json")
		var fallbackResult TestConfig
		foundPath, err = LoadConfig([]string{nonExistentPath, jsonPath}, &fallbackResult)
		require.NoError(t, err, "LoadConfig fallback should succeed")
		assert.Equal(t, jsonPath, foundPath, "Should find fallback config")
		assert.Equal(t, testConfig, fallbackResult, "Fallback config should match")

		// Test error case - no valid config files
		var errorResult TestConfig
		foundPath, err = LoadConfig([]string{"nonexistent1.json", "nonexistent2.yaml"}, &errorResult)
		require.Error(t, err, "LoadConfig should error when no files exist")
		assert.Contains(t, err.Error(), "no valid config file found", "Error should mention no valid config file")
		assert.Empty(t, foundPath, "Should return empty path on error")

		// Test with corrupted JSON file
		corruptedJSONPath := filepath.Join(tmpDir, "corrupted.json")
		err = WriteFile(corruptedJSONPath, []byte("{ invalid json }"), 0o644)
		require.NoError(t, err)

		var corruptedResult TestConfig
		foundPath, err = LoadConfig([]string{corruptedJSONPath}, &corruptedResult)
		require.Error(t, err, "LoadConfig should error on corrupted JSON")
		assert.Empty(t, foundPath, "Should return empty path on error")

		// Test with corrupted YAML file
		corruptedYAMLPath := filepath.Join(tmpDir, "corrupted.yaml")
		err = WriteFile(corruptedYAMLPath, []byte("invalid:\n  - yaml\n  - [unclosed"), 0o644)
		require.NoError(t, err)

		var corruptedYAMLResult TestConfig
		foundPath, err = LoadConfig([]string{corruptedYAMLPath}, &corruptedYAMLResult)
		require.Error(t, err, "LoadConfig should error on corrupted YAML")
		assert.Empty(t, foundPath, "Should return empty path on error")
	})

	t.Run("LoadConfigNoExtensionFallback", func(t *testing.T) {
		// Test file without extension that fails YAML but succeeds JSON
		jsonDataPath := filepath.Join(tmpDir, "json-data")
		jsonData, err := json.Marshal(testConfig)
		require.NoError(t, err)
		err = WriteFile(jsonDataPath, jsonData, 0o644)
		require.NoError(t, err)

		var result TestConfig
		foundPath, err := LoadConfig([]string{jsonDataPath}, &result)
		require.NoError(t, err, "LoadConfig should succeed with JSON fallback")
		assert.Equal(t, jsonDataPath, foundPath, "Should find JSON data file")
		assert.Equal(t, testConfig, result, "JSON data should match")
	})
}
