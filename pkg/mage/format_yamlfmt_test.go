package mage

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/utils"
)

var (
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
	t.Helper()
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

// setupYAMLTestDir creates an isolated temp working directory containing the given files
// (keyed by relative path), disables YAML line-length validation, and restores the working
// directory and environment on cleanup. Format.YAML() now discovers files with a native
// filesystem walk, so tests stage real files instead of mocking the external `find`.
// stubFormatToolsInstalled marks the formatting tools (gofumpt, gci, goimports,
// yamlfmt) as already present on PATH for the duration of the test, restoring the
// real check on cleanup. Format tests inject a MockCommandRunner and assert only on
// the formatting commands; without this stub, a host missing one of these tools
// (notably yamlfmt in CI) makes installTool attempt a real "go install" through the
// mock runner, which fails the SecureCommandRunner type assertion. Stubbing keeps
// these unit tests host-independent. Other CommandExists callers are unaffected
// because this seam only gates installTool.
func stubFormatToolsInstalled(t *testing.T) {
	t.Helper()
	orig := commandExists
	t.Cleanup(func() { commandExists = orig })
	commandExists = func(cmd string) bool {
		switch cmd {
		case "gofumpt", "gci", "goimports", "yamlfmt":
			return true
		default:
			return orig(cmd)
		}
	}
}

func setupYAMLTestDir(t *testing.T, files map[string]string) {
	t.Helper()

	stubFormatToolsInstalled(t)

	origDir, err := os.Getwd()
	require.NoError(t, err)
	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() { _ = os.Chdir(origDir) }) //nolint:errcheck // test cleanup

	for path, content := range files {
		if dir := filepath.Dir(path); dir != "." {
			require.NoError(t, os.MkdirAll(dir, 0o750))
		}
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	}

	origValidation, hadValidation := os.LookupEnv("MAGE_X_YAML_VALIDATION")
	require.NoError(t, os.Setenv("MAGE_X_YAML_VALIDATION", "false"))
	t.Cleanup(func() {
		if hadValidation {
			_ = os.Setenv("MAGE_X_YAML_VALIDATION", origValidation) //nolint:errcheck // test cleanup
		} else {
			_ = os.Unsetenv("MAGE_X_YAML_VALIDATION") //nolint:errcheck // test cleanup
		}
	})
}

// TestYamlfmtMockScenarios tests yamlfmt invocation with a mock command runner. With
// native-walk discovery, Format.YAML() feeds yamlfmt the explicit selected-file list
// (never "."), so these tests stage real files and assert the yamlfmt argv. Walk yields
// files in lexical order.
func TestYamlfmtMockScenarios(t *testing.T) {
	t.Run("successful yamlfmt formatting", func(t *testing.T) {
		setupYAMLTestDir(t, map[string]string{
			"test.yml":    "a: 1\n",
			"config.yaml": "b: 2\n",
		})

		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		mockRunner := &MockCommandRunner{}
		// Lexical walk order: config.yaml then test.yml. No config file present.
		mockRunner.On("RunCmd", "yamlfmt", "config.yaml", "test.yml").Return(nil)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		err := Format{}.YAML()
		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})

	t.Run("yamlfmt with config file", func(t *testing.T) {
		setupYAMLTestDir(t, map[string]string{
			".github/.yamlfmt": "formatter:\n  type: basic\n",
			"test.yml":         "a: 1\n",
		})

		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		mockRunner := &MockCommandRunner{}
		mockRunner.On("RunCmd", "yamlfmt", "-conf", ".github/.yamlfmt", "test.yml").Return(nil)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		err := Format{}.YAML()
		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})

	t.Run("no YAML files found", func(t *testing.T) {
		setupYAMLTestDir(t, nil) // empty directory

		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// No expectations set: any yamlfmt call would fail the mock.
		mockRunner := &MockCommandRunner{}
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		err := Format{}.YAML()
		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})

	t.Run("yamlfmt execution fails", func(t *testing.T) {
		setupYAMLTestDir(t, map[string]string{"test.yml": "a: 1\n"})

		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		mockRunner := &MockCommandRunner{}
		// The batch run fails, then the per-file fallback retries the same file; the
		// expectation matches both invocations.
		mockRunner.On("RunCmd", "yamlfmt", "test.yml").Return(ErrYamlfmtExecutionFailed)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		err := Format{}.YAML()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "yamlfmt formatting failed")
		mockRunner.AssertExpectations(t)
	})

	t.Run("yamlfmt with config file execution fails", func(t *testing.T) {
		setupYAMLTestDir(t, map[string]string{
			".github/.yamlfmt": "formatter:\n  type: basic\n",
			"test.yml":         "a: 1\n",
		})

		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		mockRunner := &MockCommandRunner{}
		mockRunner.On("RunCmd", "yamlfmt", "-conf", ".github/.yamlfmt", "test.yml").Return(ErrYamlfmtConfigExecutionFailed)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		err := Format{}.YAML()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "yamlfmt formatting failed")
		mockRunner.AssertExpectations(t)
	})
}

// TestYamlfmtMockInstallationScenarios tests yamlfmt installation scenarios with mocks
func TestYamlfmtMockInstallationScenarios(t *testing.T) {
	t.Run("yamlfmt installation success", func(t *testing.T) {
		setupYAMLTestDir(t, map[string]string{"test.yml": "a: 1\n"})

		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create mock secure runner
		mockRunner := &MockSecureCommandRunner{}
		mockRunner.MockCommandRunner.On("RunCmd", "yamlfmt", "test.yml").Return(nil)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		// Test YAML formatting through the secure-runner wrapper
		err := Format{}.YAML()
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

// TestYamlfmtEdgeCases tests yamlfmt edge cases and exclusion handling.
func TestYamlfmtEdgeCases(t *testing.T) {
	t.Run("empty yaml files list", func(t *testing.T) {
		setupYAMLTestDir(t, nil) // empty directory

		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// No expectations: yamlfmt must not be called when no YAML files exist.
		mockRunner := &MockCommandRunner{}
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		err := Format{}.YAML()
		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})

	t.Run("yaml alias method", func(t *testing.T) {
		setupYAMLTestDir(t, map[string]string{"test.yml": "a: 1\n"})

		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		mockRunner := &MockCommandRunner{}
		mockRunner.On("RunCmd", "yamlfmt", "test.yml").Return(nil)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		// Test the Yaml() alias method.
		err := Format{}.Yaml()
		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})

	t.Run("exclude paths prune both .yml and .yaml", func(t *testing.T) {
		setupYAMLTestDir(t, map[string]string{
			"keep.yml":                "a: 1\n",
			"custom_exclude/skip.yml": "x: 1\n",
			"build/skip.yaml":         "y: 2\n",
		})

		origExclude, hadExclude := os.LookupEnv("MAGE_X_FORMAT_EXCLUDE_PATHS")
		require.NoError(t, os.Setenv("MAGE_X_FORMAT_EXCLUDE_PATHS", "custom_exclude,build"))
		t.Cleanup(func() {
			if hadExclude {
				_ = os.Setenv("MAGE_X_FORMAT_EXCLUDE_PATHS", origExclude) //nolint:errcheck // test cleanup
			} else {
				_ = os.Unsetenv("MAGE_X_FORMAT_EXCLUDE_PATHS") //nolint:errcheck // test cleanup
			}
		})

		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Only the un-excluded file reaches yamlfmt; the .yml AND .yaml under excluded
		// directories are pruned. This is the precedence bug the fix resolves — previously
		// excludes applied only to .yaml, never .yml.
		mockRunner := &MockCommandRunner{}
		mockRunner.On("RunCmd", "yamlfmt", "keep.yml").Return(nil)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		err := Format{}.YAML()
		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})
}

// TestYamlfmtWithDifferentRunners tests yamlfmt with different command-runner types.
func TestYamlfmtWithDifferentRunners(t *testing.T) {
	t.Run("with mock command runner", func(t *testing.T) {
		setupYAMLTestDir(t, map[string]string{"test.yml": "a: 1\n"})

		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Test with basic MockCommandRunner
		mockRunner := &MockCommandRunner{}
		mockRunner.On("RunCmd", "yamlfmt", "test.yml").Return(nil)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		err := Format{}.YAML()
		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})

	t.Run("with mock secure command runner", func(t *testing.T) {
		setupYAMLTestDir(t, map[string]string{"test.yml": "a: 1\n"})

		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Test with MockSecureCommandRunner
		mockRunner := &MockSecureCommandRunner{}
		mockRunner.MockCommandRunner.On("RunCmd", "yamlfmt", "test.yml").Return(nil)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		err := Format{}.YAML()
		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})
}
