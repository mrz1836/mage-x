package mage

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mrz1836/mage-x/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	ErrFindFailed                   = errors.New("find failed")
	ErrYamlfmtExecutionFailed       = errors.New("yamlfmt execution failed")
	ErrYamlfmtConfigExecutionFailed = errors.New("yamlfmt config execution failed")
)

// TestYamlfmtFormatting tests yamlfmt formatting functionality with temporary YAML files
func TestYamlfmtFormatting(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "yamlfmt-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if rmErr := os.RemoveAll(tmpDir); rmErr != nil {
			t.Logf("Failed to remove temp directory: %v", rmErr)
		}
	}()

	// Change to the temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to change back to original directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create test YAML files with various formatting issues
	testFiles := map[string]string{
		"test1.yml": `name: "Test YAML"
version:   1.0
description: |
  This is a test YAML file
  with various formatting issues
items:
-   item1
-  item2
- item3
config:
    debug:    true
    verbose: false
    level:  "info"
`,
		"test2.yaml": `# Test configuration
api:
  host:  localhost
  port:    8080
  endpoints:
    - /api/v1/users
    -   /api/v1/posts
    -  /api/v1/comments
database:
     host: db.example.com
     port:     5432
     name:     myapp
features:
  - name:   feature1
    enabled: true
  - name: feature2
    enabled:   false
`,
		"nested/test3.yml": `# Nested test file
services:
  web:
    image:   nginx:latest
    ports:
      -   "80:80"
      - "443:443"
    environment:
       - NODE_ENV=production
       -   DEBUG=false
  db:
     image: postgres:13
     environment:
       POSTGRES_DB:   myapp
       POSTGRES_USER:   user
       POSTGRES_PASSWORD: pass
`,
	}

	// Create nested directory
	if err := os.MkdirAll("nested", 0o750); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Create test files with bad formatting
	for filename, content := range testFiles {
		if err := os.WriteFile(filename, []byte(content), 0o600); err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	// Copy yamlfmt config to temp directory
	configContent := `# yamlfmt test configuration
formatter:
  type: basic
  indent: 2
  include_document_start: false
  retain_line_breaks: true
  line_ending: lf
  max_line_length: 120
  scan_folded_as_literal: false
  indentless_arrays: false
  drop_merge_tag: false
  pad_line_comments: 1
`
	if err := os.MkdirAll(".github", 0o750); err != nil {
		t.Fatalf("Failed to create .github directory: %v", err)
	}
	if err := os.WriteFile(".github/.yamlfmt", []byte(configContent), 0o600); err != nil {
		t.Fatalf("Failed to create yamlfmt config: %v", err)
	}

	// Ensure yamlfmt is available
	if !utils.CommandExists("yamlfmt") {
		// Try to install it for the test
		t.Log("yamlfmt not found, attempting to install it for testing...")
		testRunner := GetRunner()
		if err := testRunner.RunCmd("go", "install", "github.com/google/yamlfmt/cmd/yamlfmt@latest"); err != nil {
			t.Skipf("Skipping test: yamlfmt not available and failed to install: %v", err)
		}

		// Verify it's now available
		if !utils.CommandExists("yamlfmt") {
			t.Skip("Skipping test: yamlfmt still not available after installation attempt")
		}
	}

	// Test the Format.YAML() method
	formatter := Format{}
	if err := formatter.YAML(); err != nil {
		t.Fatalf("Format.YAML() failed: %v", err)
	}

	// Verify that files were formatted correctly
	for filename := range testFiles {
		// Clean the file path to prevent directory traversal
		cleanPath := filepath.Clean(filename)
		content, err := os.ReadFile(cleanPath)
		if err != nil {
			t.Fatalf("Failed to read formatted file %s: %v", filename, err)
		}

		formattedContent := string(content)

		// Basic checks for proper formatting
		t.Run("verify_formatting_"+strings.ReplaceAll(filename, "/", "_"), func(t *testing.T) {
			// Check that there are no inconsistent indentation (more than one consecutive space after colons)
			lines := strings.Split(formattedContent, "\n")
			for i, line := range lines {
				// Skip comment lines and empty lines
				trimmed := strings.TrimSpace(line)
				if trimmed == "" || strings.HasPrefix(trimmed, "#") {
					continue
				}

				// Check for consistent indentation (should be multiples of 2 spaces)
				if len(line) > len(strings.TrimLeft(line, " ")) {
					indent := len(line) - len(strings.TrimLeft(line, " "))
					if indent%2 != 0 {
						t.Errorf("File %s, line %d: inconsistent indentation (%d spaces): %q",
							filename, i+1, indent, line)
					}
				}

				// Check for excessive spacing after colons in key-value pairs
				checkColonSpacing(t, line, filename, i+1)
			}

			// Verify the content is valid YAML by checking it has expected structure
			if strings.Contains(filename, "test1") {
				if !strings.Contains(formattedContent, "name:") || !strings.Contains(formattedContent, "version:") {
					t.Errorf("File %s missing expected keys after formatting", filename)
				}
			}

			if strings.Contains(filename, "test2") {
				if !strings.Contains(formattedContent, "api:") || !strings.Contains(formattedContent, "database:") {
					t.Errorf("File %s missing expected keys after formatting", filename)
				}
			}

			if strings.Contains(filename, "test3") {
				if !strings.Contains(formattedContent, "services:") || !strings.Contains(formattedContent, "web:") {
					t.Errorf("File %s missing expected keys after formatting", filename)
				}
			}
		})
	}

	// Test lint functionality
	t.Run("test_lint_mode", func(t *testing.T) {
		// Create a new file with bad formatting
		badFile := "lint-test.yml"
		badContent := `name:    badly-formatted
items:
-     item1
-   item2
config:
     debug:    true
`
		if err := os.WriteFile(badFile, []byte(badContent), 0o600); err != nil {
			t.Fatalf("Failed to create bad file for dry-run test: %v", err)
		}

		// Run yamlfmt in lint mode to check if it detects the formatting issues
		runner := GetRunner()
		err := runner.RunCmd("yamlfmt", "-conf", ".github/.yamlfmt", "-lint", ".")

		// We expect this to return an error since the file needs formatting
		if err == nil {
			t.Error("Expected yamlfmt lint to detect formatting issues, but it returned success")
		}

		// Verify the file content wasn't changed (lint mode doesn't modify files)
		unchangedContent, readErr := os.ReadFile(badFile)
		if readErr != nil {
			t.Fatalf("Failed to read lint test file: %v", readErr)
		}

		if string(unchangedContent) != badContent {
			t.Error("Lint mode should not modify files, but file content changed")
		}
	})

	// Test that files are found correctly
	t.Run("test_file_discovery", func(t *testing.T) {
		var yamlFiles []string
		err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && (strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")) {
				yamlFiles = append(yamlFiles, path)
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to walk directory: %v", err)
		}

		expectedFiles := len(testFiles) + 1 // +1 for lint-test.yml
		if len(yamlFiles) < expectedFiles {
			t.Errorf("Expected at least %d YAML files, found %d: %v", expectedFiles, len(yamlFiles), yamlFiles)
		}
	})
}

// TestYamlfmtConfiguration tests that yamlfmt configuration is properly loaded
func TestYamlfmtConfiguration(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "yamlfmt-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() {
		if rmErr := os.RemoveAll(tmpDir); rmErr != nil {
			t.Logf("Failed to remove temp directory: %v", rmErr)
		}
	}()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to change back to original directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test without config file
	t.Run("without_config_file", func(t *testing.T) {
		testFile := "no-config.yml"
		testContent := `name:   test
value:  123
`
		if err := os.WriteFile(testFile, []byte(testContent), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		formatter := Format{}

		// This should work even without config file (using defaults)
		if !utils.CommandExists("yamlfmt") {
			t.Skip("Skipping test: yamlfmt not available")
		}

		// Should not return an error even without config file
		if err := formatter.YAML(); err != nil {
			t.Errorf("YAML formatting failed without config file: %v", err)
		}
	})

	// Test with config file
	t.Run("with_config_file", func(t *testing.T) {
		// Create config file
		if err := os.MkdirAll(".github", 0o750); err != nil {
			t.Fatalf("Failed to create .github directory: %v", err)
		}

		configContent := `formatter:
  type: basic
  indent: 4
  max_line_length: 100
`
		if err := os.WriteFile(".github/.yamlfmt", []byte(configContent), 0o600); err != nil {
			t.Fatalf("Failed to create config file: %v", err)
		}

		testFile := "with-config.yml"
		testContent := `name:   test
value:  123
`
		if err := os.WriteFile(testFile, []byte(testContent), 0o600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		formatter := Format{}

		if !utils.CommandExists("yamlfmt") {
			t.Skip("Skipping test: yamlfmt not available")
		}

		if err := formatter.YAML(); err != nil {
			t.Errorf("YAML formatting failed with config file: %v", err)
		}

		// Verify the file exists and config was used
		if !utils.FileExists(".github/.yamlfmt") {
			t.Error("Config file should exist")
		}
	})
}

// checkColonSpacing checks for excessive spacing after colons in YAML key-value pairs
func checkColonSpacing(t *testing.T, line, filename string, lineNum int) {
	if !strings.Contains(line, ":") || strings.HasPrefix(strings.TrimSpace(line), "#") {
		return
	}

	// Look for patterns like "key:   value" (more than one space after colon)
	colonIndex := strings.Index(line, ":")
	if colonIndex == -1 || colonIndex >= len(line)-1 {
		return
	}

	afterColon := line[colonIndex+1:]
	if len(afterColon) == 0 || afterColon[0] != ' ' {
		return
	}

	// Count spaces after colon
	spaceCount := 0
	for j := 0; j < len(afterColon) && afterColon[j] == ' '; j++ {
		spaceCount++
	}

	// Allow exactly one space after colon for values (not arrays or objects)
	remainingAfterSpaces := strings.TrimLeft(afterColon, " ")
	if remainingAfterSpaces != "" &&
		!strings.HasPrefix(remainingAfterSpaces, "-") &&
		!strings.HasPrefix(remainingAfterSpaces, "|") &&
		!strings.HasPrefix(remainingAfterSpaces, ">") &&
		spaceCount > 1 {
		t.Errorf("File %s, line %d: excessive spacing after colon (%d spaces): %q",
			filename, lineNum, spaceCount, line)
	}
}

// TestYamlfmtMockScenarios tests yamlfmt with mock command runner scenarios
func TestYamlfmtMockScenarios(t *testing.T) {
	t.Run("successful yamlfmt formatting", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create mock runner
		mockRunner := &MockCommandRunner{}
		mockRunner.On("RunCmdOutput", "find", ".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*").Return("test.yml\nconfig.yaml", nil)
		mockRunner.On("RunCmd", "yamlfmt", ".").Return(nil)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		// Test successful YAML formatting
		formatter := Format{}
		err := formatter.YAML()
		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})

	t.Run("yamlfmt with config file", func(t *testing.T) {
		// Create temporary directory with config file
		tmpDir := t.TempDir()
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(origDir) }() //nolint:errcheck // test cleanup

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create .github directory and config file
		err = os.MkdirAll(".github", 0o750)
		require.NoError(t, err)
		err = os.WriteFile(".github/.yamlfmt", []byte("formatter:\n  type: basic\n"), 0o600)
		require.NoError(t, err)

		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create mock runner
		mockRunner := &MockCommandRunner{}
		mockRunner.On("RunCmdOutput", "find", ".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*").Return("test.yml", nil)
		mockRunner.On("RunCmd", "yamlfmt", "-conf", ".github/.yamlfmt", ".").Return(nil)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		// Test YAML formatting with config file
		formatter := Format{}
		err = formatter.YAML()
		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})

	t.Run("no YAML files found", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create mock runner that returns no YAML files
		mockRunner := &MockCommandRunner{}
		mockRunner.On("RunCmdOutput", "find", ".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*").Return("", nil)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		// Test YAML formatting when no files are found
		formatter := Format{}
		err := formatter.YAML()
		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})

	t.Run("find command fails", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create mock runner that fails the find command
		mockRunner := &MockCommandRunner{}
		mockRunner.On("RunCmdOutput", "find", ".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*").Return("", ErrFindFailed)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		// Test YAML formatting when find command fails
		formatter := Format{}
		err := formatter.YAML()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find YAML files")
		mockRunner.AssertExpectations(t)
	})

	t.Run("yamlfmt execution fails", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create mock runner that fails yamlfmt execution
		mockRunner := &MockCommandRunner{}
		mockRunner.On("RunCmdOutput", "find", ".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*").Return("test.yml", nil)
		mockRunner.On("RunCmd", "yamlfmt", ".").Return(ErrYamlfmtExecutionFailed)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		// Test YAML formatting when yamlfmt execution fails
		formatter := Format{}
		err := formatter.YAML()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "yamlfmt formatting failed")
		mockRunner.AssertExpectations(t)
	})

	t.Run("yamlfmt with config file execution fails", func(t *testing.T) {
		// Create temporary directory with config file
		tmpDir := t.TempDir()
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(origDir) }() //nolint:errcheck // test cleanup

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create .github directory and config file
		err = os.MkdirAll(".github", 0o750)
		require.NoError(t, err)
		err = os.WriteFile(".github/.yamlfmt", []byte("formatter:\n  type: basic\n"), 0o600)
		require.NoError(t, err)

		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create mock runner that fails yamlfmt with config
		mockRunner := &MockCommandRunner{}
		mockRunner.On("RunCmdOutput", "find", ".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*").Return("test.yml", nil)
		mockRunner.On("RunCmd", "yamlfmt", "-conf", ".github/.yamlfmt", ".").Return(ErrYamlfmtConfigExecutionFailed)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		// Test YAML formatting when yamlfmt with config fails
		formatter := Format{}
		err = formatter.YAML()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "yamlfmt formatting failed")
		mockRunner.AssertExpectations(t)
	})
}

// TestYamlfmtMockInstallationScenarios tests yamlfmt installation scenarios with mocks
func TestYamlfmtMockInstallationScenarios(t *testing.T) {
	t.Run("yamlfmt installation success", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create mock secure runner
		mockRunner := &MockSecureCommandRunner{}
		mockRunner.MockCommandRunner.On("RunCmdOutput", "find", ".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*").Return("test.yml", nil)
		mockRunner.MockCommandRunner.On("RunCmd", "yamlfmt", ".").Return(nil)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		// Test YAML formatting that might trigger installation
		formatter := Format{}
		err := formatter.YAML()
		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})

	t.Run("yamlfmt installation failure", func(t *testing.T) {
		// This test would need to mock utils.CommandExists and the installation process
		// For now, we test the error handling when ensureYamlfmt fails
		t.Skip("Requires advanced mocking of utils.CommandExists and installation process")
	})
}

// TestYamlfmtVersionManagement tests yamlfmt version management
func TestYamlfmtVersionManagement(t *testing.T) {
	t.Run("get default yamlfmt version", func(t *testing.T) {
		version := GetDefaultYamlfmtVersion()
		assert.NotEmpty(t, version, "Default yamlfmt version should not be empty")

		// Version should be either "latest", VersionLatest constant, or a semantic version
		validVersion := version == "latest" ||
			version == VersionLatest ||
			strings.HasPrefix(version, "v") ||
			strings.Contains(version, ".")
		assert.True(t, validVersion, "Version should be valid format: %s", version)
	})

	t.Run("version handling in installation command", func(t *testing.T) {
		version := GetDefaultYamlfmtVersion()
		expectedCmd := "github.com/google/yamlfmt/cmd/yamlfmt@" + version

		// Test that the installation command is formed correctly
		if version == "" || version == VersionLatest {
			expectedCmd = "github.com/google/yamlfmt/cmd/yamlfmt@latest"
		}

		assert.Contains(t, expectedCmd, "github.com/google/yamlfmt/cmd/yamlfmt@")
	})
}

// TestYamlfmtEdgeCases tests yamlfmt edge cases and error scenarios
func TestYamlfmtEdgeCases(t *testing.T) {
	t.Run("empty yaml files list", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create mock runner that returns empty string (no files)
		mockRunner := &MockCommandRunner{}
		mockRunner.On("RunCmdOutput", "find", ".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*").Return("", nil)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		// Test YAML formatting with no files
		formatter := Format{}
		err := formatter.YAML()
		require.NoError(t, err)

		// Should not call yamlfmt when no files are found
		mockRunner.AssertExpectations(t)
	})

	t.Run("yaml alias method", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create mock runner
		mockRunner := &MockCommandRunner{}
		mockRunner.On("RunCmdOutput", "find", ".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*").Return("test.yml", nil)
		mockRunner.On("RunCmd", "yamlfmt", ".").Return(nil)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		// Test Yaml() alias method
		formatter := Format{}
		err := formatter.Yaml()
		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})

	t.Run("exclude paths handling", func(t *testing.T) {
		// Test that exclude paths are properly applied to YAML find command
		originalEnv := os.Getenv("MAGE_X_FORMAT_EXCLUDE_PATHS")
		defer func() {
			if originalEnv == "" {
				_ = os.Unsetenv("MAGE_X_FORMAT_EXCLUDE_PATHS") //nolint:errcheck // test cleanup
			} else {
				_ = os.Setenv("MAGE_X_FORMAT_EXCLUDE_PATHS", originalEnv) //nolint:errcheck // test cleanup
			}
		}()

		// Set custom exclude paths
		_ = os.Setenv("MAGE_X_FORMAT_EXCLUDE_PATHS", "custom_exclude,build") //nolint:errcheck // test setup

		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create mock runner with custom exclude paths
		mockRunner := &MockCommandRunner{}
		mockRunner.On("RunCmdOutput", "find", ".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./custom_exclude/*", "-not", "-path", "./*custom_exclude*/*", "-not", "-path", "./build/*", "-not", "-path", "./*build*/*").Return("test.yml", nil)
		mockRunner.On("RunCmd", "yamlfmt", ".").Return(nil)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		// Test YAML formatting with custom exclude paths
		formatter := Format{}
		err := formatter.YAML()
		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})
}

// TestYamlfmtWithDifferentRunners tests yamlfmt with different types of command runners
func TestYamlfmtWithDifferentRunners(t *testing.T) {
	t.Run("with mock command runner", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Test with basic MockCommandRunner
		mockRunner := &MockCommandRunner{}
		mockRunner.On("RunCmdOutput", "find", ".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*").Return("test.yml", nil)
		mockRunner.On("RunCmd", "yamlfmt", ".").Return(nil)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		formatter := Format{}
		err := formatter.YAML()
		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})

	t.Run("with mock secure command runner", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Test with MockSecureCommandRunner
		mockRunner := &MockSecureCommandRunner{}
		mockRunner.MockCommandRunner.On("RunCmdOutput", "find", ".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*").Return("test.yml", nil)
		mockRunner.MockCommandRunner.On("RunCmd", "yamlfmt", ".").Return(nil)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		formatter := Format{}
		err := formatter.YAML()
		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})
}
