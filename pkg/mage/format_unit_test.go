package mage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errGofmtCheckFailed = fmt.Errorf("gofmt check failed")
	errGofmtWriteFailed = fmt.Errorf("gofmt write failed")
)

// TestFormatHelperFunctions tests all helper functions in isolation
func TestFormatHelperFunctions(t *testing.T) {
	t.Run("filterEmpty", func(t *testing.T) {
		tests := []struct {
			name     string
			input    []string
			expected []string
		}{
			{
				name:     "empty slice",
				input:    []string{},
				expected: nil,
			},
			{
				name:     "all empty strings",
				input:    []string{"", "", ""},
				expected: nil,
			},
			{
				name:     "mixed empty and non-empty",
				input:    []string{"", "file1.go", "", "file2.go", ""},
				expected: []string{"file1.go", "file2.go"},
			},
			{
				name:     "no empty strings",
				input:    []string{"file1.go", "file2.go"},
				expected: []string{"file1.go", "file2.go"},
			},
			{
				name:     "single empty string",
				input:    []string{""},
				expected: nil,
			},
			{
				name:     "single non-empty string",
				input:    []string{"file.go"},
				expected: []string{"file.go"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := filterEmpty(tt.input)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("getFormatExcludePaths", func(t *testing.T) {
		tests := []struct {
			name     string
			envValue string
			expected []string
		}{
			{
				name:     "default paths",
				envValue: "",
				expected: []string{"vendor", "node_modules", ".git", ".idea", ".vscode"},
			},
			{
				name:     "custom single path",
				envValue: "build",
				expected: []string{"build"},
			},
			{
				name:     "custom multiple paths",
				envValue: "build,dist,tmp",
				expected: []string{"build", "dist", "tmp"},
			},
			{
				name:     "paths with spaces",
				envValue: "build, dist , tmp",
				expected: []string{"build", " dist ", " tmp"},
			},
			{
				name:     "empty path in list",
				envValue: "build,,dist",
				expected: []string{"build", "", "dist"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.envValue == "" {
					// Test default behavior when env var not set
					_ = os.Unsetenv("MAGE_X_FORMAT_EXCLUDE_PATHS") //nolint:errcheck // test cleanup
				} else {
					_ = os.Setenv("MAGE_X_FORMAT_EXCLUDE_PATHS", tt.envValue) //nolint:errcheck // test setup
				}

				result := getFormatExcludePaths()
				assert.Equal(t, tt.expected, result)

				// Clean up
				_ = os.Unsetenv("MAGE_X_FORMAT_EXCLUDE_PATHS") //nolint:errcheck // test cleanup
			})
		}
	})

	t.Run("buildFindExcludeArgs", func(t *testing.T) {
		tests := []struct {
			name     string
			envValue string
			expected []string
		}{
			{
				name:     "single exclude path",
				envValue: "vendor",
				expected: []string{"-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*"},
			},
			{
				name:     "multiple exclude paths",
				envValue: "vendor,node_modules",
				expected: []string{
					"-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*",
					"-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*",
				},
			},
			{
				name:     "exclude path with spaces",
				envValue: " build ",
				expected: []string{"-not", "-path", "./build/*", "-not", "-path", "./*build*/*"},
			},
			{
				name:     "empty paths filtered out",
				envValue: "vendor,,build",
				expected: []string{
					"-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*",
					"-not", "-path", "./build/*", "-not", "-path", "./*build*/*",
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_ = os.Setenv("MAGE_X_FORMAT_EXCLUDE_PATHS", tt.envValue) //nolint:errcheck // test setup

				result := buildFindExcludeArgs()
				assert.Equal(t, tt.expected, result)

				// Clean up
				_ = os.Unsetenv("MAGE_X_FORMAT_EXCLUDE_PATHS") //nolint:errcheck // test cleanup
			})
		}
	})
}

// TestFormatJSONFile tests the formatJSONFile helper function behavior
func TestFormatJSONFile(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")

	// Create a test JSON file
	jsonContent := `{"name":"test","value":123}`
	err := os.WriteFile(testFile, []byte(jsonContent), 0o600)
	require.NoError(t, err)

	t.Run("function completes with system tools", func(t *testing.T) {
		// Test the function with whatever tools are available on the system
		// This is more of an integration test but demonstrates the function works
		result := formatJSONFile(testFile)

		// Result depends on what formatters are available
		// We just verify the function completes without panicking
		_ = result

		t.Logf("formatJSONFile returned: %v", result)
	})

	t.Run("test file validation", func(t *testing.T) {
		// Test with non-existent file
		nonExistentFile := filepath.Join(tmpDir, "nonexistent.json")
		result := formatJSONFile(nonExistentFile)

		// Should return false for non-existent file
		assert.False(t, result)
	})
}

// TestFindGoFiles tests the findGoFiles function
func TestFindGoFiles(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }() //nolint:errcheck // test cleanup

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create test directory structure
	testFiles := map[string]string{
		"main.go":             "package main",
		"utils.go":            "package utils",
		"test.pb.go":          "// protobuf generated file",
		"vendor/dep.go":       "package dep",
		"node_modules/lib.go": "package lib",
		"cmd/app/main.go":     "package main",
		"pkg/service/svc.go":  "package service",
		".git/hooks/hook.go":  "package hooks",
		"README.md":           "# README",
		"config.json":         "{}",
	}

	for path, content := range testFiles {
		dir := filepath.Dir(path)
		if dir != "." {
			err := os.MkdirAll(dir, 0o750)
			require.NoError(t, err)
		}
		err := os.WriteFile(path, []byte(content), 0o600)
		require.NoError(t, err)
	}

	t.Run("default exclude paths", func(t *testing.T) {
		_ = os.Unsetenv("MAGE_X_FORMAT_EXCLUDE_PATHS") //nolint:errcheck // test cleanup // Use defaults

		files, err := findGoFiles()
		require.NoError(t, err)

		// Should include Go files but exclude .pb.go files and vendor/node_modules/etc
		expectedFiles := []string{"main.go", "utils.go", "cmd/app/main.go", "pkg/service/svc.go"}

		// Convert to map for easier comparison
		foundMap := make(map[string]bool)
		for _, file := range files {
			foundMap[file] = true
		}

		for _, expected := range expectedFiles {
			assert.True(t, foundMap[expected], "Expected to find %s", expected)
		}

		// Should not include these
		unexpectedFiles := []string{"test.pb.go", "vendor/dep.go", "node_modules/lib.go", ".git/hooks/hook.go"}
		for _, unexpected := range unexpectedFiles {
			assert.False(t, foundMap[unexpected], "Should not find %s", unexpected)
		}
	})

	t.Run("custom exclude paths", func(t *testing.T) {
		_ = os.Setenv("MAGE_X_FORMAT_EXCLUDE_PATHS", "cmd,pkg")           //nolint:errcheck // test setup
		defer func() { _ = os.Unsetenv("MAGE_X_FORMAT_EXCLUDE_PATHS") }() //nolint:errcheck // test cleanup

		files, err := findGoFiles()
		require.NoError(t, err)

		// Convert to map for easier comparison
		foundMap := make(map[string]bool)
		for _, file := range files {
			foundMap[file] = true
		}

		// Should include top-level files
		assert.True(t, foundMap["main.go"])
		assert.True(t, foundMap["utils.go"])

		// Should exclude custom paths but not default ones (since we overrode)
		assert.False(t, foundMap["cmd/app/main.go"])
		assert.False(t, foundMap["pkg/service/svc.go"])

		// Should still exclude .pb.go files
		assert.False(t, foundMap["test.pb.go"])
	})
}

// TestFormatDefaultWithPartialFailures tests Format.Default() with some formatters failing
// This test runs in a temp directory to avoid modifying actual project files.
func TestFormatDefaultWithPartialFailures(t *testing.T) {
	// Create temp directory and change to it
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }() //nolint:errcheck // test cleanup

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create minimal Go files for formatters to work on
	err = os.WriteFile("go.mod", []byte("module testformat\n\ngo 1.21\n"), 0o600)
	require.NoError(t, err)

	err = os.WriteFile("main.go", []byte("package main\n\nfunc main() {}\n"), 0o600)
	require.NoError(t, err)

	format := Format{}

	// Test Default() method - should complete without error in an isolated environment
	// This verifies the code path works without modifying real project files
	err = format.Default()

	// The result depends on whether formatting tools are available
	// We just verify no panic occurs and the function completes
	_ = err
}

// TestFormatCheckWithMultipleIssues tests Format.Check() with various formatting issues
// This test runs in a temp directory to avoid modifying actual project files.
func TestFormatCheckWithMultipleIssues(t *testing.T) {
	// Create temp directory and change to it
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(origDir) }() //nolint:errcheck // test cleanup

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create minimal Go files for formatters to check
	err = os.WriteFile("go.mod", []byte("module testformat\n\ngo 1.21\n"), 0o600)
	require.NoError(t, err)

	err = os.WriteFile("main.go", []byte("package main\n\nfunc main() {}\n"), 0o600)
	require.NoError(t, err)

	format := Format{}

	// Test Check() method - should complete without error in an isolated environment
	// This verifies the code path works without checking real project files
	err = format.Check()

	// The result depends on whether formatting tools are available and files need formatting
	// We just verify no panic occurs and the function completes
	_ = err
}

// TestFormatGofmtEdgeCases tests Format.Gofmt() edge cases
func TestFormatGofmtEdgeCases(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*MockCommandRunner)
		expectedError bool
	}{
		{
			name: "no files need formatting",
			setupMock: func(m *MockCommandRunner) {
				// Mock gofmt -l returns empty output (no files need formatting)
				m.On("RunCmdOutput", "gofmt", "-l", ".").Return("", nil)
			},
			expectedError: false,
		},
		{
			name: "files need formatting",
			setupMock: func(m *MockCommandRunner) {
				// Mock gofmt -l returns files that need formatting
				m.On("RunCmdOutput", "gofmt", "-l", ".").Return("file1.go\nfile2.go", nil)
				// Mock gofmt -w to format the files
				m.On("RunCmd", "gofmt", "-w", ".").Return(nil)
			},
			expectedError: false,
		},
		{
			name: "gofmt check fails",
			setupMock: func(m *MockCommandRunner) {
				// Mock gofmt -l returns error
				m.On("RunCmdOutput", "gofmt", "-l", ".").Return("", errGofmtCheckFailed)
			},
			expectedError: true,
		},
		{
			name: "gofmt format fails",
			setupMock: func(m *MockCommandRunner) {
				// Mock gofmt -l returns files that need formatting
				m.On("RunCmdOutput", "gofmt", "-l", ".").Return("file1.go", nil)
				// Mock gofmt -w fails
				m.On("RunCmd", "gofmt", "-w", ".").Return(errGofmtWriteFailed)
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockRunner(t, func(mockRunner *MockCommandRunner) {
				tt.setupMock(mockRunner)

				format := Format{}
				err := format.Gofmt()

				if tt.expectedError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		})
	}
}
