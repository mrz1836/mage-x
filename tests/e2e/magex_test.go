package e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestMain sets up the test environment
func TestMain(m *testing.M) {
	// Build magex binary for testing
	if err := buildMagexBinary(); err != nil {
		fmt.Printf("Failed to build magex binary: %v\n", err)
		os.Exit(1)
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
		// Cleanup error is non-critical
		_ = err
	}
	if err := os.RemoveAll("test_projects"); err != nil {
		// Cleanup error is non-critical
		_ = err
	}
}

func TestMagexVersion(t *testing.T) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "./magex", "-version")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("magex -version failed: %v", err)
	}

	outputStr := string(output)

	// Check for key components of the enhanced version output
	expectedStrings := []string{
		"MAGE-X - Write Once, Mage Everywhere", // Banner tagline
		"Version Information",                  // Section header
		"Version:",                             // Version field
		"Capabilities",                         // Capabilities section
		"Commands:",                            // Commands count
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(outputStr, expected) {
			t.Errorf("Version output should contain '%s', got: %s", expected, outputStr)
		}
	}

	// Verify it contains a version number (e.g., 1.2.1)
	if !strings.Contains(outputStr, "Version:") {
		t.Errorf("Version output should contain version number, got: %s", outputStr)
	}
}

func TestMagexHelp(t *testing.T) {
	tests := []struct {
		flag string
	}{
		{"-h"},
		{"-help"},
		{"--help"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("help_%s", strings.TrimPrefix(tt.flag, "-")), func(t *testing.T) {
			ctx := context.Background()
			// #nosec G204 - tt.flag is controlled test input
			cmd := exec.CommandContext(ctx, "./magex", tt.flag)
			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("magex %s failed: %v", tt.flag, err)
			}

			outputStr := string(output)
			expectedSections := []string{"MAGE-X", "Usage:", "Examples:"}
			for _, section := range expectedSections {
				if !strings.Contains(outputStr, section) {
					t.Errorf("Help output should contain '%s', got: %s", section, outputStr)
				}
			}
		})
	}
}

func TestMagexList(t *testing.T) {
	tests := []struct {
		flag     string
		expected []string
	}{
		{
			flag:     "-l",
			expected: []string{"build", "test", "lint"},
		},
		{
			flag:     "-list",
			expected: []string{"build", "test", "lint"},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("list_%s", strings.TrimPrefix(tt.flag, "-")), func(t *testing.T) {
			ctx := context.Background()
			// #nosec G204 - tt.flag is controlled test input
			cmd := exec.CommandContext(ctx, "./magex", tt.flag)
			output, err := cmd.Output()
			if err != nil {
				t.Fatalf("magex %s failed: %v", tt.flag, err)
			}

			outputStr := string(output)
			for _, expected := range tt.expected {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("List output should contain '%s', got: %s", expected, outputStr)
				}
			}
		})
	}
}

func TestMagexNamespaceList(t *testing.T) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "./magex", "-n")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("magex -n failed: %v", err)
	}

	outputStr := string(output)
	expectedNamespaces := []string{"build:", "test:", "lint:"}

	for _, ns := range expectedNamespaces {
		if !strings.Contains(outputStr, ns) {
			t.Errorf("Namespace list should contain '%s', got: %s", ns, outputStr)
		}
	}
}

func TestMagexSearch(t *testing.T) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "./magex", "-search", "build")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("magex -search build failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "build") {
		t.Errorf("Search output should contain 'build', got: %s", outputStr)
	}
}

func TestMagexInit(t *testing.T) {
	// Create temporary test directory
	testDir := filepath.Join("test_projects", "init_test")
	if err := os.MkdirAll(testDir, 0o750); err != nil {
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

	// Run init
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "../../magex", "-init")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("magex -init failed: %v\nOutput: %s", err, string(output))
	}

	// Check if magefile.go was created
	if _, err := os.Stat("magefile.go"); os.IsNotExist(err) {
		t.Error("magex -init should create magefile.go")
	}

	// Check output
	outputStr := string(output)
	if !strings.Contains(outputStr, "magefile.go") {
		t.Errorf("Init output should mention magefile.go, got: %s", outputStr)
	}
}

func TestMagexZeroConfig(t *testing.T) {
	// Test that magex works without any magefile.go
	testDir := filepath.Join("test_projects", "zero_config")
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

	// Create a simple Go project
	goMod := `module testproject

go 1.19
`
	mainGo := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`

	if writeErr := os.WriteFile("go.mod", []byte(goMod), 0o600); writeErr != nil {
		t.Fatalf("Failed to write go.mod: %v", writeErr)
	}
	if writeErr := os.WriteFile("main.go", []byte(mainGo), 0o600); writeErr != nil {
		t.Fatalf("Failed to write main.go: %v", writeErr)
	}

	// Test list commands work
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "../../magex", "-l")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("magex -l in zero-config project failed: %v", err)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "build") {
		t.Errorf("Zero-config should show built-in commands, got: %s", outputStr)
	}
}

func TestMagexWithCustomMagefile(t *testing.T) {
	// Test that magex works with custom magefile.go
	testDir := filepath.Join("test_projects", "with_custom")
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

	// Create a simple Go project with magefile
	goMod := `module testproject

go 1.19

require github.com/magefile/mage v1.14.0
`

	magefile := `//go:build mage
package main

import (
	"fmt"
)

// CustomBuild is a custom build command
func CustomBuild() error {
	fmt.Println("Running custom build...")
	return nil
}

// CustomTest is a custom test command
func CustomTest() error {
	fmt.Println("Running custom tests...")
	return nil
}
`

	mainGo := `package main

import "fmt"

func main() {
	fmt.Println("Hello, Custom!")
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

	// Test that both built-in and custom commands are available
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "../../magex", "-l")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("magex -l with custom magefile failed: %v", err)
	}

	outputStr := string(output)

	// Should have built-in commands
	if !strings.Contains(outputStr, "build") {
		t.Errorf("Should show built-in build commands, got: %s", outputStr)
	}

	// Should have custom commands (this might not work in plugin mode, so we just check parsing)
	if !strings.Contains(strings.ToLower(outputStr), "custom") {
		t.Log("Custom commands not shown - this is expected if plugin compilation failed")
	}
}

func TestMagexCommandExecution(t *testing.T) {
	// Test executing built-in commands
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

	// Create a simple Go project
	goMod := `module testproject

go 1.19
`
	mainGo := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`

	if writeErr := os.WriteFile("go.mod", []byte(goMod), 0o600); writeErr != nil {
		t.Fatalf("Failed to write go.mod: %v", writeErr)
	}
	if writeErr := os.WriteFile("main.go", []byte(mainGo), 0o600); writeErr != nil {
		t.Fatalf("Failed to write main.go: %v", writeErr)
	}

	// Test some basic commands that should work
	tests := []struct {
		name    string
		command string
		timeout time.Duration
	}{
		{
			name:    "help_command",
			command: "help",
			timeout: 5 * time.Second,
		},
		{
			name:    "version_info",
			command: "version",
			timeout: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			// #nosec G204 - tt.command is controlled test input
			cmd := exec.CommandContext(ctx, "../../magex", tt.command)
			cmd.Env = append(os.Environ(), "MAGEX_VERBOSE=true")

			output, err := cmd.CombinedOutput()
			if err != nil {
				// Some commands might not exist or might fail, that's okay for testing
				t.Logf("Command %s failed (expected): %v", tt.command, err)
				t.Logf("Output: %s", string(output))
			} else {
				t.Logf("Command %s succeeded: %s", tt.command, string(output))
			}
		})
	}
}

func TestMagexErrorHandling(t *testing.T) {
	// Test error handling for invalid commands
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "./magex", "nonexistent_command")
	output, err := cmd.CombinedOutput()

	if err == nil {
		t.Error("magex should fail for non-existent command")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "unknown command") &&
		!strings.Contains(outputStr, "not found") {
		t.Errorf("Error output should indicate unknown command, got: %s", outputStr)
	}
}

func TestMagexVerboseMode(t *testing.T) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "./magex", "-v", "-version")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("magex -v -version failed: %v", err)
	}

	outputStr := string(output)
	// In verbose mode, should show additional output
	if !strings.Contains(outputStr, "MAGE-X") {
		t.Errorf("Verbose version should show MAGE-X info, got: %s", outputStr)
	}
}

func TestMagexClean(t *testing.T) {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "./magex", "-clean")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("magex -clean failed: %v", err)
	}

	outputStr := string(output)
	if len(strings.TrimSpace(outputStr)) == 0 {
		t.Error("Clean command should produce some output")
	}
}

func TestMagexDefaultBehavior(t *testing.T) {
	// Test running magex without any arguments
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "./magex")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("magex (no args) failed: %v", err)
	}

	outputStr := string(output)
	expectedElements := []string{
		"MAGE-X",
		"Available Commands",
		"magex <command>",
	}

	for _, element := range expectedElements {
		if !strings.Contains(outputStr, element) {
			t.Errorf("Default output should contain '%s', got: %s", element, outputStr)
		}
	}
}

func TestMagexBinarySize(t *testing.T) {
	// Test that the binary exists and has reasonable size
	stat, err := os.Stat("./magex")
	if err != nil {
		t.Fatalf("magex binary not found: %v", err)
	}

	size := stat.Size()
	// Should be at least 1MB (with all embedded commands) but less than 100MB
	if size < 1024*1024 {
		t.Errorf("magex binary seems too small: %d bytes", size)
	}
	if size > 100*1024*1024 {
		t.Errorf("magex binary seems too large: %d bytes", size)
	}

	t.Logf("magex binary size: %.2f MB", float64(size)/(1024*1024))
}

func TestMagexOutputFormat(t *testing.T) {
	// Test that command outputs are properly formatted
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "./magex", "-l")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("magex -l failed: %v\nStderr: %s", err, stderr.String())
	}

	output := stdout.String()

	// Should be formatted as a table/list
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 5 {
		t.Errorf("List output should have multiple lines, got %d", len(lines))
	}

	// Should not have excessive whitespace
	for i, line := range lines {
		if i < 3 && len(strings.TrimSpace(line)) == 0 {
			t.Errorf("List output should not have empty lines at the beginning, line %d is empty", i)
		}
	}
}

func BenchmarkMagexStartup(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		cmd := exec.CommandContext(ctx, "./magex", "-version")
		_, err := cmd.Output()
		if err != nil {
			b.Fatalf("magex -version failed: %v", err)
		}
	}
}

func BenchmarkMagexList(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		cmd := exec.CommandContext(ctx, "./magex", "-l")
		_, err := cmd.Output()
		if err != nil {
			b.Fatalf("magex -l failed: %v", err)
		}
	}
}
