package mage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFormatNumberWithCommas tests the formatNumberWithCommas helper function
func TestFormatNumberWithCommas(t *testing.T) {
	testCases := []struct {
		name     string
		input    int
		expected string
	}{
		{
			name:     "single digit",
			input:    5,
			expected: "5",
		},
		{
			name:     "double digit",
			input:    42,
			expected: "42",
		},
		{
			name:     "three digits",
			input:    123,
			expected: "123",
		},
		{
			name:     "four digits",
			input:    1234,
			expected: "1,234",
		},
		{
			name:     "five digits",
			input:    12345,
			expected: "12,345",
		},
		{
			name:     "six digits",
			input:    123456,
			expected: "123,456",
		},
		{
			name:     "seven digits",
			input:    1234567,
			expected: "1,234,567",
		},
		{
			name:     "test file count example",
			input:    59475,
			expected: "59,475",
		},
		{
			name:     "go file count example",
			input:    53778,
			expected: "53,778",
		},
		{
			name:     "total count example",
			input:    113253,
			expected: "113,253",
		},
		{
			name:     "one million",
			input:    1000000,
			expected: "1,000,000",
		},
		{
			name:     "zero",
			input:    0,
			expected: "0",
		},
		{
			name:     "exactly thousand",
			input:    1000,
			expected: "1,000",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatNumberWithCommas(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestLOCResult_JSONMarshal tests that LOCResult correctly marshals to JSON
func TestLOCResult_JSONMarshal(t *testing.T) {
	result := LOCResult{
		TestFilesLOC:    1000,
		TestFilesCount:  10,
		GoFilesLOC:      5000,
		GoFilesCount:    50,
		TotalLOC:        6000,
		TotalFilesCount: 60,
		Date:            "2025-01-15",
		ExcludedDirs:    []string{"vendor", "third_party"},
	}

	jsonBytes, err := json.Marshal(result)
	require.NoError(t, err)

	// Verify it's valid JSON
	var unmarshaled LOCResult
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	// Verify all fields are preserved
	assert.Equal(t, result.TestFilesLOC, unmarshaled.TestFilesLOC)
	assert.Equal(t, result.TestFilesCount, unmarshaled.TestFilesCount)
	assert.Equal(t, result.GoFilesLOC, unmarshaled.GoFilesLOC)
	assert.Equal(t, result.GoFilesCount, unmarshaled.GoFilesCount)
	assert.Equal(t, result.TotalLOC, unmarshaled.TotalLOC)
	assert.Equal(t, result.TotalFilesCount, unmarshaled.TotalFilesCount)
	assert.Equal(t, result.Date, unmarshaled.Date)
	assert.Equal(t, result.ExcludedDirs, unmarshaled.ExcludedDirs)
}

// TestLOCResult_JSONFieldNames tests that JSON field names are correct
func TestLOCResult_JSONFieldNames(t *testing.T) {
	result := LOCResult{
		TestFilesLOC:    100,
		TestFilesCount:  5,
		GoFilesLOC:      200,
		GoFilesCount:    10,
		TotalLOC:        300,
		TotalFilesCount: 15,
		Date:            "2025-12-15",
		ExcludedDirs:    []string{"vendor"},
	}

	jsonBytes, err := json.Marshal(result)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)

	// Verify snake_case field names are used
	assert.Contains(t, jsonStr, `"test_files_loc"`)
	assert.Contains(t, jsonStr, `"test_files_count"`)
	assert.Contains(t, jsonStr, `"go_files_loc"`)
	assert.Contains(t, jsonStr, `"go_files_count"`)
	assert.Contains(t, jsonStr, `"total_loc"`)
	assert.Contains(t, jsonStr, `"total_files_count"`)
	assert.Contains(t, jsonStr, `"date"`)
	assert.Contains(t, jsonStr, `"excluded_dirs"`)
}

// TestLOCStats tests the LOCStats struct
func TestLOCStats(t *testing.T) {
	stats := LOCStats{Lines: 100, Files: 5}
	assert.Equal(t, 100, stats.Lines)
	assert.Equal(t, 5, stats.Files)

	// Test zero values
	emptyStats := LOCStats{}
	assert.Equal(t, 0, emptyStats.Lines)
	assert.Equal(t, 0, emptyStats.Files)
}

// TestCountLinesWithStats tests the countLinesWithStats helper function
func TestCountLinesWithStats(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "loc_test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) }) //nolint:errcheck,gosec // cleanup

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(originalDir) }) //nolint:errcheck,gosec // cleanup

	// Create test files
	testContent := "package main\n// comment\nfunc Test() {}\n"
	err = os.WriteFile("example_test.go", []byte(testContent), 0o600)
	require.NoError(t, err)

	t.Run("CountTestFiles", func(t *testing.T) {
		stats, err := countLinesWithStats("*_test.go", []string{})
		require.NoError(t, err)
		assert.Equal(t, 1, stats.Files)
		assert.Equal(t, 2, stats.Lines) // excludes comment and empty lines
	})

	t.Run("ExcludeVendor", func(t *testing.T) {
		// Create vendor directory with test file
		vendorDir := filepath.Join(tmpDir, "vendor")
		err := os.MkdirAll(vendorDir, 0o750)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(vendorDir, "vendor_test.go"), []byte(testContent), 0o600)
		require.NoError(t, err)

		stats, err := countLinesWithStats("*_test.go", []string{"vendor"})
		require.NoError(t, err)
		assert.Equal(t, 1, stats.Files) // Should not count vendor file
	})
}

// TestCountGoLinesWithStats tests the countGoLinesWithStats helper function
func TestCountGoLinesWithStats(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "loc_test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) }) //nolint:errcheck,gosec // cleanup

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(originalDir) }) //nolint:errcheck,gosec // cleanup

	// Create Go file (not test)
	goContent := "package main\nfunc main() {}\n"
	err = os.WriteFile("main.go", []byte(goContent), 0o600)
	require.NoError(t, err)

	// Create test file (should be excluded)
	testContent := "package main\nfunc TestMain() {}\n"
	err = os.WriteFile("main_test.go", []byte(testContent), 0o600)
	require.NoError(t, err)

	t.Run("CountGoFilesExcludingTests", func(t *testing.T) {
		stats, err := countGoLinesWithStats([]string{})
		require.NoError(t, err)
		assert.Equal(t, 1, stats.Files) // Only main.go, not main_test.go
		assert.Equal(t, 2, stats.Lines)
	})
}
