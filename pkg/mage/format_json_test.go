package mage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

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
			name:     "nested object keeps key order",
			input:    `{"user":{"name":"john","age":30},"active":true}`,
			expected: "{\n    \"user\": {\n        \"name\": \"john\",\n        \"age\": 30\n    },\n    \"active\": true\n}\n",
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
			_, err = formatJSONFileNative(testFile)

			if tt.shouldError {
				require.Error(t, err, "Expected formatting to fail")
				return
			}

			require.NoError(t, err, "Expected formatting to succeed")

			// Check the result
			result, err := os.ReadFile(testFile) //nolint:gosec // test file path
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(result))

			// Verify the result is valid JSON
			var jsonData any
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
		"package.json":     "{\n    \"name\": \"test\",\n    \"version\": \"1.0.0\",\n    \"dependencies\": {\n        \"lodash\": \"^4.17.21\"\n    }\n}\n",
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
	changed, err := formatJSONFileNative(testFile)
	require.NoError(t, err)
	require.True(t, changed)

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

// TestFormatJSONBytes verifies that formatting changes only insignificant whitespace.
func TestFormatJSONBytes(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "keeps key order",
			input:    `{"name":"x","version":"1.0.0","dependencies":{"b":"1","a":"2"}}`,
			expected: "{\n    \"name\": \"x\",\n    \"version\": \"1.0.0\",\n    \"dependencies\": {\n        \"b\": \"1\",\n        \"a\": \"2\"\n    }\n}\n",
		},
		{
			name:     "keeps number literals",
			input:    `{"big":12345678901234567890,"float":1.0,"exp":1e3,"neg":-0.5}`,
			expected: "{\n    \"big\": 12345678901234567890,\n    \"float\": 1.0,\n    \"exp\": 1e3,\n    \"neg\": -0.5\n}\n",
		},
		{
			name:     "keeps string escapes",
			input:    `{"unicode":"caf\u00e9","slash":"a\/b","html":"<a&b>"}`,
			expected: "{\n    \"unicode\": \"caf\\u00e9\",\n    \"slash\": \"a\\/b\",\n    \"html\": \"<a&b>\"\n}\n",
		},
		{
			name:     "compacts empty containers",
			input:    `{"a": { }, "b": [ ]}`,
			expected: "{\n    \"a\": {},\n    \"b\": []\n}\n",
		},
		{
			name:     "normalizes surrounding whitespace",
			input:    "\n\t {\"a\":1} \r\n\n",
			expected: "{\n    \"a\": 1\n}\n",
		},
		{
			name:     "reindents two-space documents",
			input:    "{\n  \"a\": [\n    1\n  ]\n}\n",
			expected: "{\n    \"a\": [\n        1\n    ]\n}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := formatJSONBytes([]byte(tt.input))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(got))

			again, err := formatJSONBytes(got)
			require.NoError(t, err)
			assert.Equal(t, string(got), string(again), "formatting must be idempotent")
		})
	}

	t.Run("rejects invalid documents", func(t *testing.T) {
		for _, input := range []string{"", "   ", `{"a":}`, `{"a":1}{"b":2}`, "{\"a\":1} // comment"} {
			_, err := formatJSONBytes([]byte(input))
			assert.Error(t, err, "input %q", input)
		}
	})
}

// TestFormatJSONFileNativePreservesPermissions verifies that rewriting a file keeps its
// permission bits instead of tightening them.
func TestFormatJSONFileNativePreservesPermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix permission bits are not supported on Windows")
	}

	for _, perm := range []os.FileMode{0o644, 0o640, 0o600, 0o444} {
		t.Run(perm.String(), func(t *testing.T) {
			testFile := filepath.Join(t.TempDir(), "test.json")
			require.NoError(t, os.WriteFile(testFile, []byte(`{"b":1,"a":2}`), 0o600))
			require.NoError(t, os.Chmod(testFile, perm))

			changed, err := formatJSONFileNative(testFile)
			require.NoError(t, err)
			require.True(t, changed)

			info, err := os.Stat(testFile)
			require.NoError(t, err)
			assert.Equal(t, perm, info.Mode().Perm(), "permissions must be preserved")
			assert.Equal(t, "{\n    \"b\": 1,\n    \"a\": 2\n}\n", readTestFile(t, testFile))

			_, err = os.Stat(testFile + ".tmp")
			assert.True(t, os.IsNotExist(err), "temporary file must be removed")
		})
	}
}

// TestFormatJSONFileNativeLeavesFormattedFileUntouched verifies that an already formatted
// file is not rewritten.
func TestFormatJSONFileNativeLeavesFormattedFileUntouched(t *testing.T) {
	testFile := filepath.Join(t.TempDir(), "test.json")
	formatted := "{\n    \"b\": 1,\n    \"a\": 2\n}\n"
	require.NoError(t, os.WriteFile(testFile, []byte(formatted), 0o600))

	past := time.Now().Add(-time.Hour).Truncate(time.Second)
	require.NoError(t, os.Chtimes(testFile, past, past))

	changed, err := formatJSONFileNative(testFile)
	require.NoError(t, err)
	assert.False(t, changed)

	info, err := os.Stat(testFile)
	require.NoError(t, err)
	assert.True(t, info.ModTime().Equal(past), "an already formatted file must not be rewritten")
	assert.Equal(t, formatted, readTestFile(t, testFile))
}
