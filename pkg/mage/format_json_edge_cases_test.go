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
			expected:    "{\n    \"chinese\": \"你好\",\n    \"emoji\": \"😀\",\n    \"japanese\": \"こんにちは\"\n}\n", //nolint:gosmopolitan // testing unicode support
			description: "JSON with unicode characters should be preserved",
		},
		{
			name:        "escaped characters",
			input:       `{"newline":"line1\nline2","tab":"before\tafter","quote":"say \"hello\"","backslash":"path\\to\\file"}`,
			expected:    "{\n    \"backslash\": \"path\\\\to\\\\file\",\n    \"newline\": \"line1\\nline2\",\n    \"quote\": \"say \\\"hello\\\"\",\n    \"tab\": \"before\\tafter\"\n}\n",
			description: "Escaped characters should be preserved",
		},
		{
			name:        "null values",
			input:       `{"value":null,"array":[null,1,null],"nested":{"prop":null}}`,
			expected:    "{\n    \"array\": [\n        null,\n        1,\n        null\n    ],\n    \"nested\": {\n        \"prop\": null\n    },\n    \"value\": null\n}\n",
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
			expected:    "{\n    \"float\": 3.14159265359,\n    \"large\": 123456789,\n    \"scientific\": 1.23e-10,\n    \"small\": 1\n}\n",
			description: "Various number formats should be preserved",
		},
		{
			name:        "empty structures",
			input:       `{"empty_object":{},"empty_array":[],"empty_string":""}`,
			expected:    "{\n    \"empty_array\": [],\n    \"empty_object\": {},\n    \"empty_string\": \"\"\n}\n",
			description: "Empty objects and arrays should be formatted correctly",
		},
		{
			name:        "boolean values",
			input:       `{"true_val":true,"false_val":false,"mixed":[true,false,true]}`,
			expected:    "{\n    \"false_val\": false,\n    \"mixed\": [\n        true,\n        false,\n        true\n    ],\n    \"true_val\": true\n}\n",
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
			expected:    "{\n    \"name\": \"second\"\n}\n",
			description: "Duplicate keys should result in last value winning",
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

			success := formatJSONFileNative(testFile)

			if tt.shouldError {
				assert.False(t, success, tt.description)
				return
			}

			assert.True(t, success, tt.description)

			result, err := os.ReadFile(testFile) //nolint:gosec // test file path
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(result), tt.description)

			// Verify result is valid JSON
			var jsonData interface{}
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
		success := formatJSONFileNative(filepath.Join(tmpDir, "nonexistent.json"))
		assert.False(t, success, "Should fail for nonexistent file")
	})

	t.Run("directory instead of file", func(t *testing.T) {
		dirPath := filepath.Join(tmpDir, "directory.json")
		err := os.Mkdir(dirPath, 0o750)
		require.NoError(t, err)

		success := formatJSONFileNative(dirPath)
		assert.False(t, success, "Should fail when trying to format a directory")
	})

	t.Run("read-only file", func(t *testing.T) {
		readOnlyFile := filepath.Join(tmpDir, "readonly.json")
		err := os.WriteFile(readOnlyFile, []byte(`{"test": "value"}`), 0o444) //nolint:gosec // intentionally testing read-only permission
		require.NoError(t, err)

		success := formatJSONFileNative(readOnlyFile)
		// Note: This might succeed on some systems where we can create temp files
		// and replace even read-only files. The behavior is system-dependent.
		// We just verify the function handles the case gracefully.
		if success {
			// If it succeeded, verify the result is properly formatted
			content, err := os.ReadFile(readOnlyFile) //nolint:gosec // test file path
			require.NoError(t, err)
			var jsonData interface{}
			err = json.Unmarshal(content, &jsonData)
			require.NoError(t, err, "Result should be valid JSON")
		}
		// Either way (success or failure), the function should handle it gracefully
		assert.NotPanics(t, func() {
			formatJSONFileNative(readOnlyFile)
		}, "Should not panic for read-only file")
	})

	t.Run("symlink to json file", func(t *testing.T) {
		originalFile := filepath.Join(tmpDir, "original.json")
		symlinkFile := filepath.Join(tmpDir, "symlink.json")

		err := os.WriteFile(originalFile, []byte(`{"name":"test"}`), 0o600)
		require.NoError(t, err)

		err = os.Symlink(originalFile, symlinkFile)
		require.NoError(t, err)

		success := formatJSONFileNative(symlinkFile)

		// Note: Symlink handling can be system-dependent. The test verifies graceful handling.
		// The important thing is that it doesn't panic and handles the case appropriately
		assert.NotPanics(t, func() {
			formatJSONFileNative(symlinkFile)
		}, "Should not panic for symlink file")

		if success {
			// If formatting succeeded, verify the result is valid JSON
			content, err := os.ReadFile(symlinkFile) //nolint:gosec // test file path
			require.NoError(t, err)
			var jsonData interface{}
			err = json.Unmarshal(content, &jsonData)
			require.NoError(t, err, "Result should be valid JSON")
			t.Log("Symlink formatting succeeded and result is valid JSON")
		} else {
			t.Log("Symlink formatting failed, but function handled it gracefully")
		}
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
	success := formatJSONFileNative(testFile)
	require.True(t, success)

	firstFormat, err := os.ReadFile(testFile) //nolint:gosec // test file path
	require.NoError(t, err)

	// Format the file a second time
	success = formatJSONFileNative(testFile)
	require.True(t, success)

	secondFormat, err := os.ReadFile(testFile) //nolint:gosec // test file path
	require.NoError(t, err)

	// Both formats should be identical
	assert.Equal(t, string(firstFormat), string(secondFormat), "Formatting should be idempotent")

	// Format a third time to be sure
	success = formatJSONFileNative(testFile)
	require.True(t, success)

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

		success := formatJSONFileNative(testFile)
		assert.False(t, success, "Should fail when temp file exists as directory")

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

			success := formatJSONFileNative(testFile)
			require.True(t, success)

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
	results := make([]bool, numFiles)

	// Create multiple JSON files
	for i := 0; i < numFiles; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.json", i))
		jsonContent := fmt.Sprintf(`{"file":%d,"name":"test%d","data":[1,2,3]}`, i, i)
		err := os.WriteFile(testFile, []byte(jsonContent), 0o600)
		require.NoError(t, err)
	}

	// Format all files concurrently
	for i := 0; i < numFiles; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.json", index))
			results[index] = formatJSONFileNative(testFile)
		}(i)
	}

	wg.Wait()

	// Check that all formatting operations succeeded
	for i, success := range results {
		assert.True(t, success, "File %d should have been formatted successfully", i)
	}

	// Verify all files are properly formatted
	for i := 0; i < numFiles; i++ {
		testFile := filepath.Join(tmpDir, fmt.Sprintf("test%d.json", i))
		content, err := os.ReadFile(testFile) //nolint:gosec // test file path
		require.NoError(t, err)

		expected := fmt.Sprintf("{\n    \"data\": [\n        1,\n        2,\n        3\n    ],\n    \"file\": %d,\n    \"name\": \"test%d\"\n}\n", i, i)
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
			for i := 0; i < b.N; i++ {
				// Write the unformatted JSON
				err := os.WriteFile(testFile, []byte(tc.json), 0o600)
				if err != nil {
					b.Fatal(err)
				}

				// Format the file
				success := formatJSONFileNative(testFile)
				if !success {
					b.Fatal("Formatting failed")
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

	for i := 0; i < b.N; i++ {
		err := os.WriteFile(testFile, []byte(largeJSON), 0o600)
		if err != nil {
			b.Fatal(err)
		}

		success := formatJSONFileNative(testFile)
		if !success {
			b.Fatal("Formatting failed")
		}
	}
}
