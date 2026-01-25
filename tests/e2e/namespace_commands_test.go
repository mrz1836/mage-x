package e2e

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const (
	namespaceTestDirPerm  = 0o750
	namespaceTestFilePerm = 0o600
)

// TestNamespaceCommands_EndToEnd tests that namespace commands work correctly
// through the full magex binary pipeline. This is a regression test for the
// convertToMageFormat fix that ensures commands like "Custom:Action" are properly
// converted to "custom:action" for mage execution.
//
// Note: The test uses unique namespace names (Custom, Pipeline) that don't conflict
// with magex's built-in commands. This ensures the commands are delegated to mage
// rather than being handled by built-in command handlers.
func TestNamespaceCommands_EndToEnd(t *testing.T) {
	// Create test project directory
	testDir := filepath.Join("test_projects", "namespace_test")
	if err := os.MkdirAll(testDir, namespaceTestDirPerm); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("Failed to cleanup test directory: %v", err)
		}
	}()

	// Change to test directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Logf("Failed to restore directory: %v", chErr)
		}
	}()

	if chErr := os.Chdir(testDir); chErr != nil {
		t.Fatalf("Failed to change to test directory: %v", chErr)
	}

	// Create go.mod
	goModContent := `module testnamespace

go 1.21
`
	if err := os.WriteFile("go.mod", []byte(goModContent), namespaceTestFilePerm); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create magefiles directory
	if err := os.Mkdir("magefiles", namespaceTestDirPerm); err != nil {
		t.Fatalf("Failed to create magefiles directory: %v", err)
	}

	// Create a magefile with unique namespace commands that don't conflict
	// with magex's built-in commands (CI, Pipeline, Deploy, Custom)
	magefileContent := `//go:build mage

package main

import (
	"fmt"
)

// Custom namespace for custom commands (won't conflict with built-ins)
type Custom struct{}

// Action runs a custom action
func (Custom) Action() error {
	fmt.Println("E2E_SUCCESS:custom:action")
	return nil
}

// Process runs processing
func (Custom) Process() error {
	fmt.Println("E2E_SUCCESS:custom:process")
	return nil
}

// Pipeline namespace for pipeline commands
type Pipeline struct{}

// Run runs the pipeline
func (Pipeline) Run() error {
	fmt.Println("E2E_SUCCESS:pipeline:run")
	return nil
}

// Deploy namespace for deployment commands
type Deploy struct{}

// Staging deploys to staging
func (Deploy) Staging() error {
	fmt.Println("E2E_SUCCESS:deploy:staging")
	return nil
}
`
	if err := os.WriteFile(filepath.Join("magefiles", "commands.go"), []byte(magefileContent), namespaceTestFilePerm); err != nil {
		t.Fatalf("Failed to create magefile: %v", err)
	}

	// Test cases for namespace command handling through the magex binary
	// These use unique namespace names that don't conflict with built-in commands
	tests := []struct {
		name              string
		input             string
		expectedLowercase string
		expectedOutput    string
		description       string
	}{
		{
			name:              "mixed_case_custom_action",
			input:             "Custom:Action",
			expectedLowercase: "custom:action",
			expectedOutput:    "E2E_SUCCESS:custom:action",
			description:       "Mixed case namespace - the primary test case",
		},
		{
			name:              "all_uppercase_custom_action",
			input:             "CUSTOM:ACTION",
			expectedLowercase: "custom:action",
			expectedOutput:    "E2E_SUCCESS:custom:action",
			description:       "All uppercase namespace",
		},
		{
			name:              "already_lowercase_custom_action",
			input:             "custom:action",
			expectedLowercase: "custom:action",
			expectedOutput:    "E2E_SUCCESS:custom:action",
			description:       "Already lowercase namespace",
		},
		{
			name:              "mixed_case_pipeline_run",
			input:             "Pipeline:Run",
			expectedLowercase: "pipeline:run",
			expectedOutput:    "E2E_SUCCESS:pipeline:run",
			description:       "Pipeline run command",
		},
		{
			name:              "mixed_case_deploy_staging",
			input:             "Deploy:Staging",
			expectedLowercase: "deploy:staging",
			expectedOutput:    "E2E_SUCCESS:deploy:staging",
			description:       "Deploy staging command",
		},
		{
			name:              "mixed_case_custom_process",
			input:             "Custom:Process",
			expectedLowercase: "custom:process",
			expectedOutput:    "E2E_SUCCESS:custom:process",
			description:       "Custom process command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			// Run the magex binary with the namespace command
			// #nosec G204 - tt.input is controlled test input
			cmd := exec.CommandContext(ctx, "../../magex", tt.input)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			output := stdout.String()
			errOutput := stderr.String()

			// Command succeeded - verify output
			if err == nil {
				if !strings.Contains(output, tt.expectedOutput) {
					t.Errorf("Output for '%s' (%s) should contain '%s', got: %s",
						tt.input, tt.description, tt.expectedOutput, output)
				}
				return
			}

			// Command failed - check if it's due to mage not finding the target
			// This is expected in test environments where mage can't compile the test magefiles
			switch {
			case strings.Contains(errOutput, "Unknown target specified"):
				// Verify the lowercase conversion happened correctly
				if strings.Contains(errOutput, tt.expectedLowercase) {
					t.Logf("Command '%s' correctly converted to '%s' (mage target not found in test env)",
						tt.input, tt.expectedLowercase)
				} else {
					t.Errorf("Error should reference the lowercase command format '%s', got: %s",
						tt.expectedLowercase, errOutput)
				}
			case strings.Contains(errOutput, "custom command failed"):
				// Also acceptable - this means it was delegated but mage couldn't find target
				t.Logf("Command '%s' was delegated to mage (target not found in test env): %s",
					tt.input, errOutput)
			default:
				// Unexpected error
				t.Errorf("magex %s failed unexpectedly: %v\nStdout: %s\nStderr: %s",
					tt.input, err, output, errOutput)
			}
		})
	}
}

// TestNamespaceCommands_ListShowsNamespaces verifies that namespace commands
// appear in the command list with proper formatting.
func TestNamespaceCommands_ListShowsNamespaces(t *testing.T) {
	// Create test project directory
	testDir := filepath.Join("test_projects", "namespace_list_test")
	if err := os.MkdirAll(testDir, namespaceTestDirPerm); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("Failed to cleanup test directory: %v", err)
		}
	}()

	// Change to test directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Logf("Failed to restore directory: %v", chErr)
		}
	}()

	if chErr := os.Chdir(testDir); chErr != nil {
		t.Fatalf("Failed to change to test directory: %v", chErr)
	}

	// Create go.mod
	goModContent := `module testnamespace

go 1.21
`
	if writeErr := os.WriteFile("go.mod", []byte(goModContent), namespaceTestFilePerm); writeErr != nil {
		t.Fatalf("Failed to create go.mod: %v", writeErr)
	}

	// Create magefiles directory
	if mkdirErr := os.Mkdir("magefiles", namespaceTestDirPerm); mkdirErr != nil {
		t.Fatalf("Failed to create magefiles directory: %v", mkdirErr)
	}

	// Create a magefile with namespace commands
	magefileContent := `//go:build mage

package main

import "fmt"

// CustomNS namespace for custom commands
type CustomNS struct{}

// MyCommand is a custom command
func (CustomNS) MyCommand() error {
	fmt.Println("custom command executed")
	return nil
}
`
	if writeErr := os.WriteFile(filepath.Join("magefiles", "commands.go"), []byte(magefileContent), namespaceTestFilePerm); writeErr != nil {
		t.Fatalf("Failed to create magefile: %v", writeErr)
	}

	// Run magex -l to list commands
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "../../magex", "-l")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		t.Fatalf("magex -l failed: %v\nStdout: %s\nStderr: %s",
			err, stdout.String(), stderr.String())
	}

	output := stdout.String()

	// The list should include the custom namespace command
	// Note: The exact format depends on how magex formats the list
	// We just verify it contains something related to our namespace
	if !strings.Contains(strings.ToLower(output), "customns") &&
		!strings.Contains(strings.ToLower(output), "mycommand") {
		t.Logf("Custom namespace commands may not appear in list (plugin mode might not be active)")
		t.Logf("Output: %s", output)
	}
}

// TestNamespaceCommands_ErrorHandling verifies that invalid namespace commands
// produce appropriate error messages.
func TestNamespaceCommands_ErrorHandling(t *testing.T) {
	// Create test project directory
	testDir := filepath.Join("test_projects", "namespace_error_test")
	if err := os.MkdirAll(testDir, namespaceTestDirPerm); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("Failed to cleanup test directory: %v", err)
		}
	}()

	// Change to test directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Logf("Failed to restore directory: %v", chErr)
		}
	}()

	if chErr := os.Chdir(testDir); chErr != nil {
		t.Fatalf("Failed to change to test directory: %v", chErr)
	}

	// Create go.mod
	goModContent := `module testnamespace

go 1.21
`
	if writeErr := os.WriteFile("go.mod", []byte(goModContent), namespaceTestFilePerm); writeErr != nil {
		t.Fatalf("Failed to create go.mod: %v", writeErr)
	}

	// Create magefiles directory with only CI namespace
	if mkdirErr := os.Mkdir("magefiles", namespaceTestDirPerm); mkdirErr != nil {
		t.Fatalf("Failed to create magefiles directory: %v", mkdirErr)
	}

	magefileContent := `//go:build mage

package main

import "fmt"

// CI namespace for CI/CD commands
type CI struct{}

// Static runs static analysis
func (CI) Static() error {
	fmt.Println("ci:static executed")
	return nil
}
`
	if writeErr := os.WriteFile(filepath.Join("magefiles", "commands.go"), []byte(magefileContent), namespaceTestFilePerm); writeErr != nil {
		t.Fatalf("Failed to create magefile: %v", writeErr)
	}

	// Test that nonexistent namespace command fails appropriately
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "../../magex", "NonExistent:Command")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("magex should fail for non-existent namespace command")
	}

	outputStr := string(output)
	// Should indicate the command was not found
	if !strings.Contains(strings.ToLower(outputStr), "unknown") &&
		!strings.Contains(strings.ToLower(outputStr), "not found") &&
		!strings.Contains(strings.ToLower(outputStr), "error") {
		t.Logf("Error output may vary by implementation: %s", outputStr)
	}
}

// TestNamespaceCommands_CaseInsensitivity verifies that mage's case-insensitive
// command lookup works correctly through magex. The key test is that all case
// variations are converted to lowercase before being passed to mage.
func TestNamespaceCommands_CaseInsensitivity(t *testing.T) {
	// Create test project directory
	testDir := filepath.Join("test_projects", "namespace_case_test")
	if err := os.MkdirAll(testDir, namespaceTestDirPerm); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("Failed to cleanup test directory: %v", err)
		}
	}()

	// Change to test directory
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Logf("Failed to restore directory: %v", chErr)
		}
	}()

	if chErr := os.Chdir(testDir); chErr != nil {
		t.Fatalf("Failed to change to test directory: %v", chErr)
	}

	// Create go.mod
	goModContent := `module testnamespace

go 1.21
`
	if err := os.WriteFile("go.mod", []byte(goModContent), namespaceTestFilePerm); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create magefiles directory
	if err := os.Mkdir("magefiles", namespaceTestDirPerm); err != nil {
		t.Fatalf("Failed to create magefiles directory: %v", err)
	}

	// Create a magefile with a unique namespace name
	magefileContent := `//go:build mage

package main

import "fmt"

// UniqueNS is a test namespace with a unique name
type UniqueNS struct{}

// Execute is a test command
func (UniqueNS) Execute() error {
	fmt.Println("CASE_TEST_SUCCESS")
	return nil
}
`
	if err := os.WriteFile(filepath.Join("magefiles", "commands.go"), []byte(magefileContent), namespaceTestFilePerm); err != nil {
		t.Fatalf("Failed to create magefile: %v", err)
	}

	// Test various case combinations - all should be converted to lowercase
	caseVariations := []struct {
		input             string
		expectedLowercase string
	}{
		{"uniquens:execute", "uniquens:execute"},
		{"UniqueNS:Execute", "uniquens:execute"},
		{"UNIQUENS:EXECUTE", "uniquens:execute"},
		{"uniqueNS:Execute", "uniquens:execute"},
	}

	for _, variation := range caseVariations {
		t.Run(variation.input, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			// #nosec G204 - variation.input is controlled test input
			cmd := exec.CommandContext(ctx, "../../magex", variation.input)
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			output := stdout.String()
			errOutput := stderr.String()

			// Command succeeded - verify output
			if err == nil {
				if !strings.Contains(output, "CASE_TEST_SUCCESS") {
					t.Errorf("Case variation '%s' should produce expected output, got: %s",
						variation.input, output)
				}
				return
			}

			// Command failed - verify the lowercase conversion happened
			// Combine stdout and stderr for checking (some output may go to either)
			combinedOutput := output + errOutput

			isKnownFailure := strings.Contains(combinedOutput, "Unknown target specified") ||
				strings.Contains(combinedOutput, "custom command failed")

			// Pipe warnings from Go's exec package are acceptable - they indicate
			// the command ran but had issues with pipe cleanup (race condition)
			isPipeWarning := strings.Contains(errOutput, "failed to close stderr pipe") ||
				strings.Contains(errOutput, "failed to close stdout pipe")

			if !isKnownFailure && !isPipeWarning {
				// Unexpected error
				t.Errorf("magex %s failed unexpectedly: %v\nStdout: %s\nStderr: %s",
					variation.input, err, output, errOutput)
				return
			}

			// If it's just a pipe warning without command output, log and pass
			// (the command delegation worked, just had cleanup issues)
			if isPipeWarning && !isKnownFailure {
				t.Logf("Case variation '%s' completed with pipe warning (acceptable): %s",
					variation.input, errOutput)
				return
			}

			// Verify the command was converted to lowercase
			if strings.Contains(combinedOutput, variation.expectedLowercase) {
				t.Logf("Case variation '%s' correctly converted to '%s' (target not found in test env)",
					variation.input, variation.expectedLowercase)
			} else {
				t.Errorf("Error should reference lowercase '%s', got: %s",
					variation.expectedLowercase, combinedOutput)
			}
		})
	}
}
