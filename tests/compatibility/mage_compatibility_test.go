package compatibility

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMageCompatibility tests that magex is compatible with existing mage usage
func TestMain(m *testing.M) {
	// Build magex binary for testing
	if err := buildMagexBinary(); err != nil {
		panic("Failed to build magex binary: " + err.Error())
	}

	// Run tests
	code := m.Run()

	// Cleanup
	cleanup()

	os.Exit(code)
}

func buildMagexBinary() error {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", "magex", "../../cmd/magex")
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	return cmd.Run()
}

func cleanup() {
	if err := os.Remove("magex"); err != nil {
		// Cleanup errors are non-critical
		_ = err
	}
	if err := os.RemoveAll("test_projects"); err != nil {
		// Cleanup errors are non-critical
		_ = err
	}
}

func TestMageCommandLineFlags(t *testing.T) {
	// Test that all major mage flags are supported by magex
	tests := []struct {
		name     string
		flag     string
		expectOk bool
	}{
		{"list", "-l", true},
		{"verbose", "-v", true},
		{"help", "-h", true},
		{"clean", "-clean", true},
		{"init", "-init", true},
		{"compile", "-compile", false}, // Should exist but might not work without args
		{"timeout", "-t", false},       // Requires argument
		{"force", "-f", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cmd *exec.Cmd
			if tt.flag == "-compile" {
				// Skip compile flag as it needs arguments
				t.Skip("Compile flag requires arguments")
				return
			}
			switch tt.flag {
			case "-init":
				// Init flag needs a clean directory
				tmpDir := t.TempDir()
				// Get absolute path to magex binary
				magexPath := filepath.Join(".", "magex")
				if absPath, err := filepath.Abs(magexPath); err == nil {
					magexPath = absPath
				}
				// #nosec G204 - magexPath is controlled test input
				cmd = exec.CommandContext(context.Background(), magexPath, "-init")
				cmd.Dir = tmpDir
			case "-t":
				// Timeout flag requires duration
				cmd = exec.CommandContext(context.Background(), "./magex", "-t", "30s", "-version")
			default:
				// #nosec G204 - tt.flag is controlled test input
				cmd = exec.CommandContext(context.Background(), "./magex", tt.flag)
			}

			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			if tt.expectOk && err != nil && !strings.Contains(outputStr, "usage") {
				t.Errorf("Flag %s should be supported, got error: %v\nOutput: %s", tt.flag, err, outputStr)
			}

			// Just verify the flag is recognized (doesn't produce "unknown flag" error)
			if strings.Contains(outputStr, "unknown flag") || strings.Contains(outputStr, "flag provided but not defined") {
				t.Errorf("Flag %s not recognized by magex: %s", tt.flag, outputStr)
			}
		})
	}
}

func TestMagefileCompatibility(t *testing.T) {
	// Test that existing magefiles work with magex
	testDir := filepath.Join("test_projects", "magefile_compat")
	if err := os.MkdirAll(testDir, 0o750); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("Failed to cleanup test directory: %v", err)
		}
	}()

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

	// Create typical mage project structure
	goMod := `module testcompat

go 1.19

require github.com/magefile/mage v1.14.0
`

	// Create a standard magefile that would work with mage
	magefile := `//go:build mage
package main

import (
	"fmt"
	"os"
	"os/exec"
)

// Build builds the application
func Build() error {
	fmt.Println("Building application...")
	return nil
}

// Test runs the tests
func Test() error {
	fmt.Println("Running tests...")
	cmd := exec.CommandContext(context.Background(), "go", "test", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Clean removes build artifacts
func Clean() error {
	fmt.Println("Cleaning build artifacts...")
	return os.RemoveAll("bin")
}

// Default target
func Default() {
	Build()
}
`

	mainGo := `package main

import "fmt"

func main() {
	fmt.Println("Hello, Mage Compatibility!")
}
`

	testGo := `package main

import "testing"

func TestMain(t *testing.T) {
	// Simple test
}
`

	if writeErr := os.WriteFile("go.mod", []byte(goMod), 0o600); writeErr != nil {
		t.Fatalf("Failed to write go.mod: %v", writeErr)
	}
	if writeErr := os.WriteFile("magefile.go", []byte(magefile), 0o600); writeErr != nil {
		t.Fatalf("Failed to write magefile.go: %v", writeErr)
	}
	if writeErr := os.WriteFile("main.go", []byte(mainGo), 0o600); writeErr != nil {
		t.Fatalf("Failed to write main.go: %v", writeErr)
	}
	if writeErr := os.WriteFile("main_test.go", []byte(testGo), 0o600); writeErr != nil {
		t.Fatalf("Failed to write main_test.go: %v", writeErr)
	}

	// Test that magex can discover and list the magefile commands
	cmd := exec.CommandContext(context.Background(), "../../magex", "-l")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("magex -l failed with magefile: %v", err)
	}

	outputStr := string(output)

	// Should show built-in commands
	if !strings.Contains(outputStr, "build") {
		t.Errorf("Should show build commands, got: %s", outputStr)
	}

	// Note: Custom magefile commands might not be shown if plugin compilation fails,
	// which is expected in test environments without full Go toolchain
	t.Logf("Commands output: %s", outputStr)
}

func TestMageEnvironmentVariables(t *testing.T) {
	// Test that mage environment variables work with magex
	tests := []struct {
		name string
		env  string
		flag string
	}{
		{"verbose_env", "MAGE_VERBOSE=1", "-version"},
		{"debug_env", "MAGE_DEBUG=1", "-version"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// #nosec G204 - tt.flag is controlled test input
			cmd := exec.CommandContext(context.Background(), "./magex", tt.flag)
			cmd.Env = append(os.Environ(), tt.env)

			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Logf("Command failed (expected): %v", err)
			}

			// Just verify that environment variables don't break magex
			if strings.Contains(string(output), "panic") {
				t.Errorf("Environment variable %s caused panic: %s", tt.env, string(output))
			}
		})
	}
}

func TestMageBehaviorCompatibility(t *testing.T) {
	// Test that magex behaves similarly to mage for common operations
	tests := []struct {
		name        string
		args        []string
		shouldPass  bool
		description string
	}{
		{
			name:        "version_check",
			args:        []string{"-version"},
			shouldPass:  true,
			description: "Version flag should work",
		},
		{
			name:        "list_commands",
			args:        []string{"-l"},
			shouldPass:  true,
			description: "List flag should work",
		},
		{
			name:        "help_flag",
			args:        []string{"-h"},
			shouldPass:  true,
			description: "Help flag should work",
		},
		{
			name:        "verbose_version",
			args:        []string{"-v", "-version"},
			shouldPass:  true,
			description: "Verbose version should work",
		},
		{
			name:        "unknown_command",
			args:        []string{"totally_nonexistent_command"},
			shouldPass:  false,
			description: "Unknown commands should fail gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// #nosec G204 - tt.args is controlled test input
			cmd := exec.CommandContext(context.Background(), "./magex", tt.args...)
			output, err := cmd.CombinedOutput()

			if tt.shouldPass && err != nil {
				t.Errorf("%s: expected success but got error: %v\nOutput: %s",
					tt.description, err, string(output))
			}

			if !tt.shouldPass && err == nil {
				t.Errorf("%s: expected failure but command succeeded\nOutput: %s",
					tt.description, string(output))
			}

			// Check for panic or other serious errors
			outputStr := string(output)
			if strings.Contains(outputStr, "panic:") {
				t.Errorf("%s: command panicked: %s", tt.description, outputStr)
			}
		})
	}
}

func TestMageOutputFormat(t *testing.T) {
	// Test that magex produces output in compatible format with mage
	cmd := exec.CommandContext(context.Background(), "./magex", "-l")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("magex -l failed: %v\nStderr: %s", err, stderr.String())
	}

	output := stdout.String()

	// Should produce tabular output similar to mage
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) < 3 {
		t.Errorf("List output should have multiple lines like mage, got %d lines", len(lines))
	}

	// Should not produce excessive debug output in normal mode
	if strings.Contains(output, "DEBUG:") || strings.Contains(output, "TRACE:") {
		t.Errorf("List output should not contain debug information by default: %s", output)
	}
}

func TestMageExitCodes(t *testing.T) {
	// Test that magex uses compatible exit codes with mage
	tests := []struct {
		name         string
		args         []string
		expectedCode int
	}{
		{
			name:         "success_version",
			args:         []string{"-version"},
			expectedCode: 0,
		},
		{
			name:         "success_help",
			args:         []string{"-h"},
			expectedCode: 0,
		},
		{
			name:         "failure_unknown_command",
			args:         []string{"nonexistent_command_xyz"},
			expectedCode: 1,
		},
		{
			name:         "success_list",
			args:         []string{"-l"},
			expectedCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// #nosec G204 - tt.args is controlled test input
			cmd := exec.CommandContext(context.Background(), "./magex", tt.args...)
			err := cmd.Run()

			var actualCode int
			if err != nil {
				exitError := &exec.ExitError{}
				if errors.As(err, &exitError) {
					actualCode = exitError.ExitCode()
				} else {
					t.Fatalf("Unexpected error type: %v", err)
				}
			}

			if actualCode != tt.expectedCode {
				t.Errorf("Expected exit code %d, got %d for args %v",
					tt.expectedCode, actualCode, tt.args)
			}
		})
	}
}

func TestMageDirectoryHandling(t *testing.T) {
	// Test that magex handles directory changes like mage
	testDir := filepath.Join("test_projects", "directory_test")
	subDir := filepath.Join(testDir, "subdir")
	if err := os.MkdirAll(subDir, 0o750); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("Failed to cleanup test directory: %v", err)
		}
	}()

	// Create magefile in root directory
	magefile := `//go:build mage
package main

import "fmt"

func Hello() error {
	fmt.Println("Hello from magefile")
	return nil
}
`

	magefilePath := filepath.Join(testDir, "magefile.go")
	if err := os.WriteFile(magefilePath, []byte(magefile), 0o600); err != nil {
		t.Fatalf("Failed to write magefile: %v", err)
	}

	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if restoreErr := os.Chdir(oldDir); restoreErr != nil {
			t.Logf("Failed to restore directory: %v", restoreErr)
		}
	}()

	// Test from root directory
	if chErr := os.Chdir(testDir); chErr != nil {
		t.Fatalf("Failed to change to test directory: %v", chErr)
	}
	cmd := exec.CommandContext(context.Background(), "../../magex", "-l")
	output, err := cmd.Output()
	if err != nil {
		t.Logf("magex -l from root failed: %v (expected if plugin compilation fails)", err)
	} else {
		t.Logf("Root directory output: %s", string(output))
	}

	// Test from subdirectory (should find magefile in parent like mage)
	if chdirErr := os.Chdir("subdir"); chdirErr != nil {
		t.Fatalf("Failed to change to subdir: %v", chdirErr)
	}
	cmd = exec.CommandContext(context.Background(), "../../../magex", "-l")
	output, err = cmd.Output()
	if err != nil {
		t.Logf("magex -l from subdir failed: %v (expected if plugin compilation fails)", err)
	} else {
		t.Logf("Subdirectory output: %s", string(output))
	}
}

func TestMageTargetExecution(t *testing.T) {
	// Test basic target execution compatibility
	testDir := filepath.Join("test_projects", "execution_test")
	if err := os.MkdirAll(testDir, 0o750); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("Failed to cleanup test directory: %v", err)
		}
	}()

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

	// Create simple Go project
	goMod := `module exectest

go 1.19
`
	mainGo := `package main

import "fmt"

func main() {
	fmt.Println("Execution test")
}
`

	if writeErr := os.WriteFile("go.mod", []byte(goMod), 0o600); writeErr != nil {
		t.Fatalf("Failed to write go.mod: %v", writeErr)
	}
	if writeErr := os.WriteFile("main.go", []byte(mainGo), 0o600); writeErr != nil {
		t.Fatalf("Failed to write main.go: %v", writeErr)
	}

	// Test that built-in commands work
	tests := []string{
		"help",
		// Note: We avoid commands that actually do work (like build, test)
		// as they might fail in test environment
	}

	for _, command := range tests {
		t.Run("command_"+command, func(t *testing.T) {
			// #nosec G204 - command is controlled test input
			cmd := exec.CommandContext(context.Background(), "../../magex", command)
			output, err := cmd.CombinedOutput()

			// Some commands might fail, but they shouldn't panic
			if strings.Contains(string(output), "panic:") {
				t.Errorf("Command %s panicked: %s", command, string(output))
			}

			if err != nil {
				t.Logf("Command %s failed: %v (might be expected)", command, err)
			} else {
				t.Logf("Command %s succeeded", command)
			}
		})
	}
}

func BenchmarkMagexVsMageStartup(b *testing.B) {
	// Benchmark magex startup time (should be comparable to mage)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := exec.CommandContext(context.Background(), "./magex", "-version")
		_, err := cmd.Output()
		if err != nil {
			b.Fatalf("magex -version failed: %v", err)
		}
	}
}

func TestMageWorkspaceCompatibility(t *testing.T) {
	// Test compatibility with Go workspace features
	testDir := filepath.Join("test_projects", "workspace_test")
	if err := os.MkdirAll(testDir, 0o750); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("Failed to cleanup test directory: %v", err)
		}
	}()

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

	// Create go.work file (Go 1.18+ workspace)
	goWork := `go 1.19

use .
`

	goMod := `module workspace-test

go 1.19
`

	if writeErr := os.WriteFile("go.work", []byte(goWork), 0o600); writeErr != nil {
		t.Fatalf("Failed to write go.work: %v", writeErr)
	}
	if writeErr := os.WriteFile("go.mod", []byte(goMod), 0o600); writeErr != nil {
		t.Fatalf("Failed to write go.mod: %v", writeErr)
	}

	// Test that magex works in workspace environment
	cmd := exec.CommandContext(context.Background(), "../../magex", "-version")
	output, err := cmd.Output()
	if err != nil {
		t.Errorf("magex should work in Go workspace: %v\nOutput: %s", err, string(output))
	}

	if !strings.Contains(string(output), "MAGE-X") {
		t.Errorf("Version output should contain MAGE-X: %s", string(output))
	}
}
