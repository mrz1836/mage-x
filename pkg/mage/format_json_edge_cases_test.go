package mage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFormatJSONEdgeCases tests edge cases for JSON formatting
func TestFormatJSONEdgeCases(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "json-edge-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }() //nolint:errcheck // cleanup in defer

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }() //nolint:errcheck // cleanup in defer
	require.NoError(t, os.Chdir(tmpDir))

	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
		description string
	}{
		{
			name:        "unicode characters",
			input:       `{"emoji":"😀","chinese":"你好","japanese":"こんにちは"}`,                                      //nolint:gosmopolitan // testing unicode support
			expected:    "{\n    \"emoji\": \"😀\",\n    \"chinese\": \"你好\",\n    \"japanese\": \"こんにちは\"\n}\n", //nolint:gosmopolitan // testing unicode support
			description: "JSON with unicode characters should be preserved",
		},
		{
			name:        "escaped characters",
			input:       `{"newline":"line1\nline2","tab":"before\tafter","quote":"say \"hello\"","backslash":"path\\to\\file"}`,
			expected:    "{\n    \"newline\": \"line1\\nline2\",\n    \"tab\": \"before\\tafter\",\n    \"quote\": \"say \\\"hello\\\"\",\n    \"backslash\": \"path\\\\to\\\\file\"\n}\n",
			description: "Escaped characters should be preserved",
		},
		{
			name:        "null values",
			input:       `{"value":null,"array":[null,1,null],"nested":{"prop":null}}`,
			expected:    "{\n    \"value\": null,\n    \"array\": [\n        null,\n        1,\n        null\n    ],\n    \"nested\": {\n        \"prop\": null\n    }\n}\n",
			description: "Null values should be preserved",
		},
		{
			name:        "deep nesting",
			input:       `{"a":{"b":{"c":{"d":{"e":{"f":"deep"}}}}}}`,
			expected:    "{\n    \"a\": {\n        \"b\": {\n            \"c\": {\n                \"d\": {\n                    \"e\": {\n                        \"f\": \"deep\"\n                    }\n                }\n            }\n        }\n    }\n}\n",
			description: "Deep nesting should be formatted correctly",
		},
		{
			name:        "mixed array types",
			input:       `[1,"string",true,null,{"key":"value"},[1,2,3]]`,
			expected:    "[\n    1,\n    \"string\",\n    true,\n    null,\n    {\n        \"key\": \"value\"\n    },\n    [\n        1,\n        2,\n        3\n    ]\n]\n",
			description: "Arrays with mixed types should be formatted correctly",
		},
		{
			name:        "large numbers",
			input:       `{"small":1,"large":123456789,"float":3.14159265359,"scientific":1.23e-10}`,
			expected:    "{\n    \"small\": 1,\n    \"large\": 123456789,\n    \"float\": 3.14159265359,\n    \"scientific\": 1.23e-10\n}\n",
			description: "Various number formats should be preserved",
		},
		{
			name:        "empty structures",
			input:       `{"empty_object":{},"empty_array":[],"empty_string":""}`,
			expected:    "{\n    \"empty_object\": {},\n    \"empty_array\": [],\n    \"empty_string\": \"\"\n}\n",
			description: "Empty objects and arrays should be formatted correctly",
		},
		{
			name:        "boolean values",
			input:       `{"true_val":true,"false_val":false,"mixed":[true,false,true]}`,
			expected:    "{\n    \"true_val\": true,\n    \"false_val\": false,\n    \"mixed\": [\n        true,\n        false,\n        true\n    ]\n}\n",
			description: "Boolean values should be preserved",
		},
		{
			name:        "trailing comma invalid",
			input:       `{"name":"test","value":123,}`,
			shouldError: true,
			description: "JSON with trailing comma should fail",
		},
		{
			name:        "duplicate keys",
			input:       `{"name":"first","name":"second"}`,
			expected:    "{\n    \"name\": \"first\",\n    \"name\": \"second\"\n}\n",
			description: "Duplicate keys should be preserved as written",
		},
		{
			name:        "html characters not escaped",
			input:       `{"postCreateCommand":"magex lint && magex vet && magex test","htmlChars":"<script>alert('test')</script>","ampersand":"A & B"}`,
			expected:    "{\n    \"postCreateCommand\": \"magex lint && magex vet && magex test\",\n    \"htmlChars\": \"<script>alert('test')</script>\",\n    \"ampersand\": \"A & B\"\n}\n",
			description: "HTML characters like &, <, > should not be escaped to unicode",
		},
		{
			name:        "very long string",
			input:       fmt.Sprintf(`{"long_string":"%s"}`, strings.Repeat("a", 1000)),
			expected:    fmt.Sprintf("{\n    \"long_string\": \"%s\"\n}\n", strings.Repeat("a", 1000)),
			description: "Very long strings should be handled correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, "test.json")
			err := os.WriteFile(testFile, []byte(tt.input), 0o600)
			require.NoError(t, err)

			_, err = formatJSONFileNative(testFile)

			if tt.shouldError {
				require.Error(t, err, tt.description)
				return
			}

			require.NoError(t, err, tt.description)

			result, err := os.ReadFile(testFile) //nolint:gosec // test file path
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(result), tt.description)

			// Verify result is valid JSON
			var jsonData any
			err = json.Unmarshal(result, &jsonData)
			require.NoError(t, err, "Result should be valid JSON")
		})
	}
}

// TestFormatJSONFileSystemEdgeCases tests file system related edge cases
func TestFormatJSONFileSystemEdgeCases(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "json-fs-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }() //nolint:errcheck // cleanup in defer

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := formatJSONFileNative(filepath.Join(tmpDir, "nonexistent.json"))
		require.Error(t, err, "Should fail for nonexistent file")
	})

	t.Run("directory instead of file", func(t *testing.T) {
		dirPath := filepath.Join(tmpDir, "directory.json")
		err := os.Mkdir(dirPath, 0o750)
		require.NoError(t, err)

		_, err = formatJSONFileNative(dirPath)
		require.ErrorIs(t, err, ErrNotRegularFile, "Should fail when trying to format a directory")
	})

	t.Run("read-only file", func(t *testing.T) {
		readOnlyFile := filepath.Join(tmpDir, "readonly.json")
		err := os.WriteFile(readOnlyFile, []byte(`{"test": "value"}`), 0o444) //nolint:gosec // intentionally testing read-only permission
		require.NoError(t, err)

		// Replacing a read-only file is platform-dependent (Windows refuses), but the
		// function must never panic, and on success the file must stay read-only.
		assert.NotPanics(t, func() {
			_, err = formatJSONFileNative(readOnlyFile)
		}, "Should not panic for read-only file")
		if err != nil {
			t.Logf("Read-only file was not replaced on this platform: %v", err)
			return
		}

		content, err := os.ReadFile(readOnlyFile) //nolint:gosec // test file path
		require.NoError(t, err)
		want := "{\n    \"test\": \"value\"\n}\n"
		assert.Equal(t, want, string(content))

		info, err := os.Stat(readOnlyFile)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0o444), info.Mode().Perm(), "Read-only permissions should be preserved")
	})

	t.Run("symlink to json file", func(t *testing.T) {
		originalFile := filepath.Join(tmpDir, "original.json")
		symlinkFile := filepath.Join(tmpDir, "symlink.json")

		err := os.WriteFile(originalFile, []byte(`{"name":"test"}`), 0o600)
		require.NoError(t, err)

		err = os.Symlink(originalFile, symlinkFile)
		require.NoError(t, err)

		changed, err := formatJSONFileNative(symlinkFile)
		require.NoError(t, err)
		assert.True(t, changed)

		// The link itself survives and its target holds the formatted content.
		info, err := os.Lstat(symlinkFile)
		require.NoError(t, err)
		assert.NotZero(t, info.Mode()&os.ModeSymlink, "Symlink should not be replaced by a regular file")

		content, err := os.ReadFile(originalFile) //nolint:gosec // test file path
		require.NoError(t, err)
		want := "{\n    \"name\": \"test\"\n}\n"
		assert.Equal(t, want, string(content))
	})
}

// TestFormatJSONIdempotent tests that formatting is idempotent
func TestFormatJSONIdempotent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "json-idempotent-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }() //nolint:errcheck // cleanup in defer

	testFile := filepath.Join(tmpDir, "test.json")
	originalJSON := `{"name":"test","value":123,"nested":{"key":"value"}}`

	err = os.WriteFile(testFile, []byte(originalJSON), 0o600)
	require.NoError(t, err)

	// Format the file the first time
	changed, err := formatJSONFileNative(testFile)
	require.NoError(t, err)
	require.True(t, changed)

	firstFormat, err := os.ReadFile(testFile) //nolint:gosec // test file path
	require.NoError(t, err)

	// Format the file a second time; it is already formatted, so it is not rewritten
	changed, err = formatJSONFileNative(testFile)
	require.NoError(t, err)
	require.False(t, changed)

	secondFormat, err := os.ReadFile(testFile) //nolint:gosec // test file path
	require.NoError(t, err)

	// Both formats should be identical
	assert.Equal(t, string(firstFormat), string(secondFormat), "Formatting should be idempotent")

	// Format a third time to be sure
	changed, err = formatJSONFileNative(testFile)
	require.NoError(t, err)
	require.False(t, changed)

	thirdFormat, err := os.ReadFile(testFile) //nolint:gosec // test file path
	require.NoError(t, err)

	assert.Equal(t, string(firstFormat), string(thirdFormat), "Formatting should remain idempotent")
}

// TestFormatJSONErrorRecovery tests error recovery scenarios
func TestFormatJSONErrorRecovery(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "json-error-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }() //nolint:errcheck // cleanup in defer

	t.Run("temp file cleanup on error", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.json")
		err := os.WriteFile(testFile, []byte(`{"valid": "json"}`), 0o600)
		require.NoError(t, err)

		// Create a directory with the temp file name to force os.Rename to fail
		tmpFile := testFile + ".tmp"
		err = os.Mkdir(tmpFile, 0o750)
		require.NoError(t, err)

		_, err = formatJSONFileNative(testFile)
		require.Error(t, err, "Should fail when temp file exists as directory")

		// Cleanup the directory we created
		_ = os.RemoveAll(tmpFile) //nolint:errcheck // cleanup in defer

		// Verify temp file is cleaned up and original file is unchanged
		_, err = os.Stat(tmpFile)
		assert.True(t, os.IsNotExist(err), "Temp file should be cleaned up")

		content, err := os.ReadFile(testFile) //nolint:gosec // test file path
		require.NoError(t, err)
		assert.JSONEq(t, `{"valid": "json"}`, string(content), "Original file should be unchanged")
	})
}

// TestFormatJSONLineEndings tests different line ending formats
func TestFormatJSONLineEndings(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "json-line-endings-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }() //nolint:errcheck // cleanup in defer

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "unix line endings",
			input:    "{\n  \"name\": \"test\"\n}",
			expected: "{\n    \"name\": \"test\"\n}\n",
		},
		{
			name:     "windows line endings",
			input:    "{\r\n  \"name\": \"test\"\r\n}",
			expected: "{\n    \"name\": \"test\"\n}\n",
		},
		{
			name:     "mixed line endings",
			input:    "{\r\n  \"name\": \"test\"\n}",
			expected: "{\n    \"name\": \"test\"\n}\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, "test.json")
			err := os.WriteFile(testFile, []byte(tt.input), 0o600)
			require.NoError(t, err)

			_, err = formatJSONFileNative(testFile)
			require.NoError(t, err)

			result, err := os.ReadFile(testFile) //nolint:gosec // test file path
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(result))
		})
	}
}

// TestFormatJSONConcurrent tests concurrent formatting of different files
func TestFormatJSONConcurrent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "json-concurrent-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }() //nolint:errcheck // cleanup in defer

	const numFiles = 10
	var wg sync.WaitGroup
	results := make([]error, numFiles)

	// Create multiple JSON files
	for i := range numFiles {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.json", i))
		jsonContent := fmt.Sprintf(`{"file":%d,"name":"test%d","data":[1,2,3]}`, i, i)
		err := os.WriteFile(testFile, []byte(jsonContent), 0o600)
		require.NoError(t, err)
	}

	// Format all files concurrently
	for i := range numFiles {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.json", index))
			_, results[index] = formatJSONFileNative(testFile)
		}(i)
	}

	wg.Wait()

	// Check that all formatting operations succeeded
	for i, err := range results {
		require.NoError(t, err, "File %d should have been formatted successfully", i)
	}

	// Verify all files are properly formatted
	for i := range numFiles {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.json", i))
		content, err := os.ReadFile(testFile) //nolint:gosec // test file path
		require.NoError(t, err)

		expected := fmt.Sprintf("{\n    \"file\": %d,\n    \"name\": \"test%d\",\n    \"data\": [\n        1,\n        2,\n        3\n    ]\n}\n", i, i)
		assert.Equal(t, expected, string(content), "File %d should be properly formatted", i)
	}
}

// BenchmarkFormatJSONFileNative benchmarks the JSON formatting performance
func BenchmarkFormatJSONFileNative(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "json-benchmark-*")
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }() //nolint:errcheck // cleanup in defer

	testCases := []struct {
		name string
		json string
	}{
		{
			name: "small",
			json: `{"name":"test","value":123}`,
		},
		{
			name: "medium",
			json: `{"users":[{"id":1,"name":"john","email":"john@example.com","profile":{"age":30,"city":"New York"}},{"id":2,"name":"jane","email":"jane@example.com","profile":{"age":25,"city":"Los Angeles"}}],"metadata":{"total":2,"page":1,"size":10}}`,
		},
		{
			name: "large",
			json: func() string {
				var items []string
				for i := 0; i < 1000; i++ {
					items = append(items, fmt.Sprintf(`{"id":%d,"name":"item%d","value":%d}`, i, i, i*10))
				}
				return fmt.Sprintf(`{"items":[%s],"count":%d}`, strings.Join(items, ","), len(items))
			}(),
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			testFile := filepath.Join(tmpDir, fmt.Sprintf("%s.json", tc.name))

			b.ResetTimer()
			for b.Loop() {
				// Write the unformatted JSON
				err := os.WriteFile(testFile, []byte(tc.json), 0o600)
				if err != nil {
					b.Fatal(err)
				}

				// Format the file
				if _, err = formatJSONFileNative(testFile); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkFormatJSONMemoryUsage benchmarks memory usage during formatting
func BenchmarkFormatJSONMemoryUsage(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "json-memory-benchmark-*")
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }() //nolint:errcheck // cleanup in defer

	// Create a large JSON file
	var items []string
	for i := 0; i < 10000; i++ {
		items = append(items, fmt.Sprintf(`{"id":%d,"name":"item%d","description":"%s"}`,
			i, i, strings.Repeat("test", 10)))
	}
	largeJSON := fmt.Sprintf(`{"items":[%s]}`, strings.Join(items, ","))

	testFile := filepath.Join(tmpDir, "large.json")

	b.ResetTimer()
	b.ReportAllocs()

	for b.Loop() {
		err := os.WriteFile(testFile, []byte(largeJSON), 0o600)
		if err != nil {
			b.Fatal(err)
		}

		if _, err = formatJSONFileNative(testFile); err != nil {
			b.Fatal(err)
		}
	}
}
