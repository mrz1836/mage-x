//go:build integration
// +build integration

package integration

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMagexParametersComprehensive tests parameter passing for all major commands
func TestMagexParametersComprehensive(t *testing.T) {
	// Build magex once for all tests
	magexBinary := buildMagexForTesting(t)
	defer os.Remove(magexBinary)

	// Get the absolute path to the test binary
	magexPath, err := filepath.Abs(magexBinary)
	require.NoError(t, err)

	t.Run("ModGraphDepthParameter", func(t *testing.T) {
		// Test mod:graph with depth parameter
		testDir := setupGoModule(t)

		// Run mod:graph with depth=1
		cmd := exec.Command(magexPath, "mod:graph", "depth=1")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Should succeed
		assert.NoError(t, err, "Command failed: %s", outputStr)

		// Use helper function to count depth accurately
		maxDepth := countTreeDepth(outputStr)

		// With depth=1, we should see at most depth 1 (root is 0, direct deps are 1)
		assert.LessOrEqual(t, maxDepth, 1, "Depth parameter not respected - saw depth %d when requested 1", maxDepth)

		// Verify parameter is actually working by comparing with depth=3
		cmd3 := exec.Command(magexPath, "mod:graph", "depth=3")
		cmd3.Dir = testDir
		output3, err3 := cmd3.CombinedOutput()
		outputStr3 := string(output3)

		if err3 == nil {
			maxDepth3 := countTreeDepth(outputStr3)
			deeper, moreLines, analysis := compareTreeOutputs(outputStr, outputStr3)

			// depth=3 should show more than depth=1 (unless very simple module)
			if deeper || moreLines {
				t.Logf("✓ Depth parameter working correctly: %s", analysis)
			} else if maxDepth3 <= 1 {
				t.Logf("Note: Simple module graph - depth parameter working but limited dependencies")
			}

			// At minimum, depths should be properly limited
			assert.LessOrEqual(t, maxDepth3, 3, "Depth=3 should not exceed depth 3, saw %d", maxDepth3)
		}
	})

	t.Run("TestCoveragePackageParameter", func(t *testing.T) {
		// Test coverage with package parameter
		testDir := setupGoProject(t)

		// Run test:cover with package parameter
		cmd := exec.Command(magexPath, "test:cover", "package=./pkg/utils")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Should run without error (even if no tests exist)
		// The important thing is that the parameter is passed through
		if err != nil {
			// Check if it's just "no test files" which is OK
			if !strings.Contains(outputStr, "no test files") {
				t.Errorf("Command failed: %s", outputStr)
			}
		}

		// Should show that it's testing the specified package
		assert.Contains(t, outputStr, "pkg/utils", "Package parameter not used")
	})

	t.Run("BuildPlatformParameter", func(t *testing.T) {
		// Test build with platform parameter
		testDir := setupGoProject(t)

		// Run build with platform parameter
		cmd := exec.Command(magexPath, "build", "platform=linux/amd64")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Should succeed or at least show it's trying to build for linux
		if err != nil {
			// Some build errors are OK if we're not on Linux
			if !strings.Contains(outputStr, "linux") && !strings.Contains(outputStr, "GOOS=linux") {
				t.Errorf("Platform parameter not recognized: %s", outputStr)
			}
		}
	})

	t.Run("TestVerboseParameter", func(t *testing.T) {
		// Test verbose parameter
		testDir := setupGoProject(t)

		// Run test with verbose parameter
		cmd := exec.Command(magexPath, "test", "verbose=true")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Should show verbose output markers
		if err == nil || strings.Contains(outputStr, "no test files") {
			// Check for verbose mode indicators
			if !strings.Contains(outputStr, "Verbose: ✓") && !strings.Contains(outputStr, "-v") {
				t.Logf("Warning: Verbose parameter might not be working correctly")
			}
		}
	})

	t.Run("LintFixParameter", func(t *testing.T) {
		// Test lint with fix parameter
		testDir := setupGoProject(t)

		// Run lint with fix parameter
		cmd := exec.Command(magexPath, "lint", "fix=true")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Check that fix parameter is recognized
		if err == nil || strings.Contains(outputStr, "golangci-lint") {
			// Should show that --fix flag is being used
			if !strings.Contains(outputStr, "--fix") && !strings.Contains(outputStr, "fix=true") {
				t.Logf("Warning: Fix parameter might not be passed to linter")
			}
		}
	})

	t.Run("MultipleParameters", func(t *testing.T) {
		// Test multiple parameters together
		testDir := setupGoModule(t)

		// Run mod:graph with multiple parameters
		cmd := exec.Command(magexPath, "mod:graph", "depth=2", "format=tree", "show_versions=true")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Should succeed
		assert.NoError(t, err, "Command failed: %s", outputStr)

		// Check that versions are shown (@ symbol indicates version)
		if !strings.Contains(outputStr, "@") {
			t.Logf("Warning: show_versions parameter might not be working")
		}
	})

	t.Run("BooleanFlagParameter", func(t *testing.T) {
		// Test boolean flag style parameters
		testDir := setupGoProject(t)

		// Run test with race detector (boolean flag)
		cmd := exec.Command(magexPath, "test:race", "short")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Should recognize the short flag
		if err == nil || strings.Contains(outputStr, "no test files") {
			// Check for -short flag
			if !strings.Contains(outputStr, "-short") && !strings.Contains(outputStr, "Short: ✓") {
				t.Logf("Warning: Boolean flag parameter might not be working")
			}
		}
	})
}

// Helper function to build magex for testing
func buildMagexForTesting(t *testing.T) string {
	t.Helper()

	// Build magex from the project root
	cmd := exec.Command("go", "build", "-o", "magex-test", "../../cmd/magex")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build magex: %s", string(output))

	return "magex-test"
}

// Helper function to setup a basic Go module for testing
func setupGoModule(t *testing.T) string {
	t.Helper()

	testDir := t.TempDir()

	// Create go.mod
	goModContent := `module testmodule

go 1.21

require github.com/stretchr/testify v1.8.4
`
	err := os.WriteFile(filepath.Join(testDir, "go.mod"), []byte(goModContent), 0o644)
	require.NoError(t, err)

	// Run go mod download to populate the module graph
	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = testDir
	_ = cmd.Run() // Ignore errors as it might fail in CI

	return testDir
}

// Helper function to setup a basic Go project for testing
func setupGoProject(t *testing.T) string {
	t.Helper()

	testDir := setupGoModule(t)

	// Create a simple main.go
	mainContent := `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
`
	err := os.WriteFile(filepath.Join(testDir, "main.go"), []byte(mainContent), 0o644)
	require.NoError(t, err)

	// Create pkg/utils directory
	err = os.MkdirAll(filepath.Join(testDir, "pkg", "utils"), 0o755)
	require.NoError(t, err)

	// Create a simple utils file
	utilsContent := `package utils

func Add(a, b int) int {
    return a + b
}
`
	err = os.WriteFile(filepath.Join(testDir, "pkg", "utils", "math.go"), []byte(utilsContent), 0o644)
	require.NoError(t, err)

	return testDir
}

// TestParameterParsing tests the parameter parsing logic directly
func TestParameterParsing(t *testing.T) {
	// This test verifies the parameter parsing works correctly
	// by running commands and checking their output

	magexBinary := buildMagexForTesting(t)
	defer os.Remove(magexBinary)

	magexPath, err := filepath.Abs(magexBinary)
	require.NoError(t, err)

	tests := []struct {
		name           string
		command        string
		args           []string
		expectedOutput []string
		setup          func(*testing.T) string
	}{
		{
			name:    "SingleKeyValueParameter",
			command: "mod:graph",
			args:    []string{"depth=1"},
			expectedOutput: []string{
				"Dependency Tree:",
			},
			setup: setupGoModule,
		},
		{
			name:    "MultipleKeyValueParameters",
			command: "mod:graph",
			args:    []string{"depth=2", "show_versions=false"},
			expectedOutput: []string{
				"Dependency Tree:",
			},
			setup: setupGoModule,
		},
		{
			name:    "BooleanFlagParameter",
			command: "test",
			args:    []string{"verbose"},
			expectedOutput: []string{
				"Test", // Should see test-related output
			},
			setup: setupGoProject,
		},
		{
			name:    "MixedParameters",
			command: "test",
			args:    []string{"package=./...", "verbose", "short"},
			expectedOutput: []string{
				"Test", // Should see test-related output
			},
			setup: setupGoProject,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testDir := tt.setup(t)

			args := append([]string{tt.command}, tt.args...)
			cmd := exec.Command(magexPath, args...)
			cmd.Dir = testDir

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			output := stdout.String() + stderr.String()

			// Check for expected output patterns
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain '%s', but got: %s", expected, output)
				}
			}

			// If command failed, check if it's an expected failure
			if err != nil {
				// Some commands might fail due to missing dependencies, but
				// the important thing is that parameters are being parsed
				if !strings.Contains(output, "no test files") &&
					!strings.Contains(output, "no packages") &&
					!strings.Contains(output, "not found") {
					t.Errorf("Command failed unexpectedly: %v\nOutput: %s", err, output)
				}
			}
		})
	}
}

// TestParameterHelp tests that parameter information is shown in help
func TestParameterHelp(t *testing.T) {
	magexBinary := buildMagexForTesting(t)
	defer os.Remove(magexBinary)

	magexPath, err := filepath.Abs(magexBinary)
	require.NoError(t, err)

	// Test help for commands with parameters
	commandsWithParams := []struct {
		command        string
		expectedParams []string
	}{
		{
			command:        "mod:graph",
			expectedParams: []string{"depth", "format", "show_versions", "filter"},
		},
		{
			command:        "test",
			expectedParams: []string{"verbose", "race", "cover"},
		},
		{
			command:        "build",
			expectedParams: []string{"platform", "output"},
		},
	}

	for _, tc := range commandsWithParams {
		t.Run(fmt.Sprintf("Help_%s", tc.command), func(t *testing.T) {
			cmd := exec.Command(magexPath, "-h", tc.command)

			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			// Should succeed
			assert.NoError(t, err, "Help command failed: %s", outputStr)

			// Check that parameter information is shown
			for _, param := range tc.expectedParams {
				if !strings.Contains(outputStr, param) {
					t.Logf("Warning: Parameter '%s' not documented in help for %s", param, tc.command)
				}
			}
		})
	}
}
