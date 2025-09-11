package mage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFormatJSONFileNative tests the native JSON formatting function
func TestFormatJSONFileNative(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "json-format-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }() //nolint:errcheck // cleanup in defer // cleanup in defer

	// Change to temp directory for test
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }() //nolint:errcheck // cleanup in defer
	require.NoError(t, os.Chdir(tmpDir))

	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		{
			name:     "simple object",
			input:    `{"name":"test","value":123}`,
			expected: "{\n    \"name\": \"test\",\n    \"value\": 123\n}\n",
		},
		{
			name:     "nested object",
			input:    `{"user":{"name":"john","age":30},"active":true}`,
			expected: "{\n    \"active\": true,\n    \"user\": {\n        \"age\": 30,\n        \"name\": \"john\"\n    }\n}\n",
		},
		{
			name:     "array",
			input:    `[1,2,3,{"a":"b"}]`,
			expected: "[\n    1,\n    2,\n    3,\n    {\n        \"a\": \"b\"\n    }\n]\n",
		},
		{
			name:     "already formatted",
			input:    "{\n    \"name\": \"test\"\n}",
			expected: "{\n    \"name\": \"test\"\n}\n",
		},
		{
			name:        "invalid json",
			input:       `{"invalid": json}`,
			shouldError: true,
		},
		{
			name:        "empty file",
			input:       "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tmpDir, "test.json")
			err := os.WriteFile(testFile, []byte(tt.input), 0o600)
			require.NoError(t, err)

			// Test formatting
			success := formatJSONFileNative(testFile)

			if tt.shouldError {
				assert.False(t, success, "Expected formatting to fail")
				return
			}

			assert.True(t, success, "Expected formatting to succeed")

			// Check the result
			result, err := os.ReadFile(testFile) //nolint:gosec // test file path
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(result))

			// Verify the result is valid JSON
			var jsonData interface{}
			err = json.Unmarshal(result, &jsonData)
			require.NoError(t, err, "Result should be valid JSON")
		})
	}
}

// TestValidateJSONFile tests the JSON validation function
func TestValidateJSONFile(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "json-validate-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }() //nolint:errcheck // cleanup in defer // cleanup in defer

	tests := []struct {
		name        string
		content     string
		shouldError bool
	}{
		{
			name:    "valid simple object",
			content: `{"name": "test", "value": 123}`,
		},
		{
			name:    "valid array",
			content: `[1, 2, 3, {"a": "b"}]`,
		},
		{
			name:    "valid complex nested",
			content: `{"users": [{"name": "john", "settings": {"theme": "dark"}}]}`,
		},
		{
			name:        "invalid json syntax",
			content:     `{"name": "test", "value": }`,
			shouldError: true,
		},
		{
			name:        "invalid json missing quote",
			content:     `{name: "test"}`,
			shouldError: true,
		},
		{
			name:        "empty content",
			content:     "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tmpDir, "test.json")
			err := os.WriteFile(testFile, []byte(tt.content), 0o600)
			require.NoError(t, err)

			// Test validation
			err = validateJSONFile(testFile)

			if tt.shouldError {
				require.Error(t, err, "Expected validation to fail")
			} else {
				require.NoError(t, err, "Expected validation to succeed")
			}
		})
	}

	// Test file not found
	t.Run("file not found", func(t *testing.T) {
		err := validateJSONFile(filepath.Join(tmpDir, "nonexistent.json"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read file")
	})
}

// TestFormatJSONIntegration tests the full JSON formatting workflow
func TestFormatJSONIntegration(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "json-integration-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }() //nolint:errcheck // cleanup in defer // cleanup in defer

	// Change to temp directory for test
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }() //nolint:errcheck // cleanup in defer
	require.NoError(t, os.Chdir(tmpDir))

	// Create some JSON files
	files := map[string]string{
		"package.json":     `{"name":"test","version":"1.0.0","dependencies":{"lodash":"^4.17.21"}}`,
		"config.json":      `{"database":{"host":"localhost","port":5432},"debug":true}`,
		"nested/data.json": `[{"id":1,"name":"item1"},{"id":2,"name":"item2"}]`,
	}

	// Create nested directory
	err = os.MkdirAll("nested", 0o750)
	require.NoError(t, err)

	for filename, content := range files {
		writeErr := os.WriteFile(filename, []byte(content), 0o600)
		require.NoError(t, writeErr)
	}

	// Test formatting
	formatter := Format{}
	err = formatter.JSON()
	require.NoError(t, err)

	// Verify all files are properly formatted
	expectedResults := map[string]string{
		"package.json":     "{\n    \"dependencies\": {\n        \"lodash\": \"^4.17.21\"\n    },\n    \"name\": \"test\",\n    \"version\": \"1.0.0\"\n}\n",
		"config.json":      "{\n    \"database\": {\n        \"host\": \"localhost\",\n        \"port\": 5432\n    },\n    \"debug\": true\n}\n",
		"nested/data.json": "[\n    {\n        \"id\": 1,\n        \"name\": \"item1\"\n    },\n    {\n        \"id\": 2,\n        \"name\": \"item2\"\n    }\n]\n",
	}

	for filename, expected := range expectedResults {
		content, err := os.ReadFile(filename) //nolint:gosec // test file path
		require.NoError(t, err)
		assert.Equal(t, expected, string(content), "File %s should be properly formatted", filename)
	}
}

// TestAtomicJSONFormatting tests that JSON formatting is atomic (uses temp files)
func TestAtomicJSONFormatting(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "json-atomic-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }() //nolint:errcheck // cleanup in defer // cleanup in defer

	// Create test file
	testFile := filepath.Join(tmpDir, "test.json")
	originalContent := `{"name": "test"}`
	err = os.WriteFile(testFile, []byte(originalContent), 0o600)
	require.NoError(t, err)

	// Get original file stats
	originalStat, err := os.Stat(testFile)
	require.NoError(t, err)

	// Format the file
	success := formatJSONFileNative(testFile)
	require.True(t, success)

	// Check that temp file is cleaned up
	tempFile := testFile + ".tmp"
	_, err = os.Stat(tempFile)
	assert.True(t, os.IsNotExist(err), "Temp file should be cleaned up")

	// Check that original file still exists with new content
	newStat, err := os.Stat(testFile)
	require.NoError(t, err)
	assert.Equal(t, originalStat.Mode(), newStat.Mode(), "File permissions should be preserved")

	// Verify content is formatted
	content, err := os.ReadFile(testFile) //nolint:gosec // test file path
	require.NoError(t, err)
	expected := "{\n    \"name\": \"test\"\n}\n"
	assert.Equal(t, expected, string(content))
}
