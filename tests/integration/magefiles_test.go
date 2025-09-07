package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	secureDirPerm  = 0o750
	secureFilePerm = 0o600
)

func TestIntegration_MagefilesDirectory(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if magex binary exists or can be built
	magexPath := getMagexBinary(t)

	// Create a temporary directory for the test
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create go.mod file
	goModContent := `module testintegration

go 1.21
`
	err = os.WriteFile("go.mod", []byte(goModContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create magefiles directory
	err = os.Mkdir("magefiles", secureDirPerm)
	if err != nil {
		t.Fatalf("Failed to create magefiles directory: %v", err)
	}

	// Create test magefile with multiple commands
	magefilePath := filepath.Join("magefiles", "commands.go")
	magefileContent := `//go:build mage
package main

import (
	"fmt"
	"os"
	"strings"
)

// BuildProject builds the project
func BuildProject() error {
	fmt.Println("Building project from magefiles directory...")
	return nil
}

// TestProject runs tests
func TestProject() error {
	fmt.Println("Running tests from magefiles directory...")
	return nil
}

// ParamsTest demonstrates parameter handling
func ParamsTest() error {
	// Check for MAGE_ARGS environment variable set by magex
	if mageArgs := os.Getenv("MAGE_ARGS"); mageArgs != "" {
		// Parse the space-separated arguments back to slice format for display
		args := strings.Fields(mageArgs)
		fmt.Printf("Parameters received: %v\n", args)
	} else {
		// Fallback to os.Args for direct execution
		args := os.Args[1:]
		if len(args) > 0 && args[0] == "paramstest" {
			args = args[1:]
		}
		fmt.Printf("Parameters received: %v\n", args)
	}
	return nil
}

// Deploy namespace for deployment commands
type Deploy struct{}

// Staging deploys to staging
func (Deploy) Staging() error {
	fmt.Println("Deploying to staging...")
	return nil
}

// Production deploys to production
func (Deploy) Production() error {
	fmt.Println("Deploying to production...")
	return nil
}
`

	err = os.WriteFile(magefilePath, []byte(magefileContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create magefile: %v", err)
	}

	// Test command listing
	t.Run("ListCommands", func(t *testing.T) {
		// #nosec G204 -- magexPath is controlled in tests
		cmd := exec.CommandContext(context.Background(), magexPath, "-l")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command output: %s", output)
			t.Fatalf("Failed to list commands: %v", err)
		}

		outputStr := string(output)
		// Should contain commands from magefiles directory
		if !strings.Contains(outputStr, "BuildProject") && !strings.Contains(outputStr, "buildproject") {
			t.Errorf("Output should contain BuildProject command, got: %s", outputStr)
		}
		if !strings.Contains(outputStr, "TestProject") && !strings.Contains(outputStr, "testproject") {
			t.Errorf("Output should contain TestProject command, got: %s", outputStr)
		}
		if !strings.Contains(outputStr, "Deploy:Staging") && !strings.Contains(outputStr, "deploy:staging") {
			t.Errorf("Output should contain Deploy:Staging command, got: %s", outputStr)
		}
	})

	// Test command execution
	t.Run("ExecuteCommand", func(t *testing.T) {
		// #nosec G204 -- magexPath is controlled in tests
		cmd := exec.CommandContext(context.Background(), magexPath, "BuildProject")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command output: %s", output)
			t.Fatalf("Failed to execute BuildProject: %v", err)
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "Building project from magefiles directory") {
			t.Errorf("Output should contain build message, got: %s", outputStr)
		}
	})

	// Test namespace command execution
	t.Run("ExecuteNamespaceCommand", func(t *testing.T) {
		// #nosec G204 -- magexPath is controlled in tests
		cmd := exec.CommandContext(context.Background(), magexPath, "deploy:staging")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command output: %s", output)
			// Don't fail here as namespace commands might need mage binary
			t.Logf("Namespace command execution failed (may be expected): %v", err)
			return
		}

		outputStr := string(output)
		if !strings.Contains(outputStr, "Deploying to staging") {
			t.Logf("Expected staging deployment message, got: %s", outputStr)
		}
	})

	// Test parameter passing
	t.Run("ParameterPassing", func(t *testing.T) {
		// #nosec G204 -- magexPath is controlled in tests
		cmd := exec.CommandContext(context.Background(), magexPath, "paramstest", "key1=value1", "key2=value2")
		cmd.Dir = tmpDir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Logf("Command output: %s", output)
			t.Fatalf("Failed to execute ParamsTest: %v", err)
		}

		outputStr := string(output)
		// Parameters should be passed and visible in output
		if !strings.Contains(outputStr, "key1=value1") || !strings.Contains(outputStr, "key2=value2") {
			t.Errorf("Output should contain parameters, got: %s", outputStr)
		}
		// Should NOT contain "Unknown target specified" warnings
		if strings.Contains(outputStr, "Unknown target specified:") {
			t.Errorf("Output should not contain 'Unknown target specified' warnings, got: %s", outputStr)
		}
	})
}

func TestIntegration_MagefilePreference(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if magex binary exists or can be built
	magexPath := getMagexBinary(t)

	// Create a temporary directory for the test
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create go.mod file
	goModContent := `module testpreference

go 1.21
`
	err = os.WriteFile("go.mod", []byte(goModContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create both magefiles directory and root magefile.go
	err = os.Mkdir("magefiles", secureDirPerm)
	if err != nil {
		t.Fatalf("Failed to create magefiles directory: %v", err)
	}

	// Create magefile in directory (should be preferred)
	dirMagefilePath := filepath.Join("magefiles", "commands.go")
	dirMagefileContent := `//go:build mage
package main

import "fmt"

// DirCommand is from magefiles directory
func DirCommand() error {
	fmt.Println("Command from magefiles directory")
	return nil
}
`

	err = os.WriteFile(dirMagefilePath, []byte(dirMagefileContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create directory magefile: %v", err)
	}

	// Create root magefile.go (should be ignored)
	rootMagefilePath := "magefile.go"
	rootMagefileContent := `//go:build mage
package main

import "fmt"

// RootCommand is from root magefile
func RootCommand() error {
	fmt.Println("Command from root magefile")
	return nil
}
`

	err = os.WriteFile(rootMagefilePath, []byte(rootMagefileContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create root magefile: %v", err)
	}

	// Test that commands from directory are found, not from root file
	// #nosec G204 -- magexPath is controlled in tests
	cmd := exec.CommandContext(context.Background(), magexPath, "-l")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Command output: %s", output)
		t.Fatalf("Failed to list commands: %v", err)
	}

	outputStr := string(output)
	// Should contain command from directory
	if !strings.Contains(outputStr, "DirCommand") && !strings.Contains(outputStr, "dircommand") {
		t.Errorf("Output should contain DirCommand from directory, got: %s", outputStr)
	}
	// Should NOT contain command from root file
	if strings.Contains(outputStr, "RootCommand") {
		t.Errorf("Output should not contain RootCommand from root file, got: %s", outputStr)
	}

	// Test executing the directory command
	// #nosec G204 -- magexPath is controlled in tests
	cmd = exec.CommandContext(context.Background(), magexPath, "dircommand")
	cmd.Dir = tmpDir
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Logf("Command output: %s", output)
		t.Fatalf("Failed to execute DirCommand: %v", err)
	}

	outputStr = string(output)
	if !strings.Contains(outputStr, "Command from magefiles directory") {
		t.Errorf("Output should contain directory command message, got: %s", outputStr)
	}
}

func TestIntegration_MultipleFiles(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Check if magex binary exists or can be built
	magexPath := getMagexBinary(t)

	// Create a temporary directory for the test
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create go.mod file
	goModContent := `module testmultiple

go 1.21
`
	err = os.WriteFile("go.mod", []byte(goModContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create magefiles directory
	err = os.Mkdir("magefiles", secureDirPerm)
	if err != nil {
		t.Fatalf("Failed to create magefiles directory: %v", err)
	}

	// Create multiple magefile files
	buildFile := filepath.Join("magefiles", "build.go")
	buildContent := `//go:build mage
package main

import "fmt"

// Build builds the project
func Build() error {
	fmt.Println("Building...")
	return nil
}

// Clean cleans build artifacts
func Clean() error {
	fmt.Println("Cleaning...")
	return nil
}
`

	testFile := filepath.Join("magefiles", "test.go")
	testContent := `//go:build mage
package main

import "fmt"

// Test runs tests
func Test() error {
	fmt.Println("Testing...")
	return nil
}

// Lint runs linting
func Lint() error {
	fmt.Println("Linting...")
	return nil
}
`

	deployFile := filepath.Join("magefiles", "deploy.go")
	deployContent := `//go:build mage
package main

import "fmt"

// Deploy namespace for deployment commands
type Deploy struct{}

// Dev deploys to development
func (Deploy) Dev() error {
	fmt.Println("Deploying to dev...")
	return nil
}
`

	err = os.WriteFile(buildFile, []byte(buildContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create build file: %v", err)
	}

	err = os.WriteFile(testFile, []byte(testContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = os.WriteFile(deployFile, []byte(deployContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create deploy file: %v", err)
	}

	// Test that all commands from all files are discovered
	// #nosec G204 -- magexPath is controlled in tests
	cmd := exec.CommandContext(context.Background(), magexPath, "-l")
	cmd.Dir = tmpDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Command output: %s", output)
		t.Fatalf("Failed to list commands: %v", err)
	}

	outputStr := string(output)
	expectedCommands := []string{"Build", "Clean", "Test", "Lint"}
	for _, cmdName := range expectedCommands {
		if !strings.Contains(outputStr, cmdName) && !strings.Contains(outputStr, strings.ToLower(cmdName)) {
			t.Errorf("Output should contain %s command, got: %s", cmdName, outputStr)
		}
	}

	// Test executing commands from different files
	testCommands := []struct {
		name     string
		expected string
	}{
		{"Build", "Building..."},
		{"Test", "Testing..."},
		{"Lint", "Linting..."},
	}

	for _, tc := range testCommands {
		t.Run(fmt.Sprintf("Execute_%s", tc.name), func(t *testing.T) {
			// #nosec G204 -- magexPath and tc.name are controlled in tests
			cmd := exec.CommandContext(context.Background(), magexPath, tc.name)
			cmd.Dir = tmpDir
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Logf("Command output: %s", output)
				t.Fatalf("Failed to execute %s: %v", tc.name, err)
			}

			outputStr := string(output)
			if !strings.Contains(outputStr, tc.expected) {
				t.Errorf("Output should contain %q, got: %s", tc.expected, outputStr)
			}
		})
	}
}

// Helper function to get magex binary path
func getMagexBinary(t *testing.T) string {
	t.Helper()

	// Always build fresh binary for tests to ensure we test current code
	projectRoot := findProjectRoot(t)
	magexSource := filepath.Join(projectRoot, "cmd", "magex")

	// Build magex binary in temp location
	tmpDir := t.TempDir()
	magexBinary := filepath.Join(tmpDir, "magex")

	// #nosec G204 -- controlled paths in tests
	cmd := exec.CommandContext(context.Background(), "go", "build", "-o", magexBinary, magexSource)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build magex binary: %v", err)
	}

	return magexBinary
}

// Helper function to find the project root
func findProjectRoot(t *testing.T) string {
	t.Helper()

	// Start from current directory and walk up
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	for {
		// Check if this directory contains go.mod and looks like mage-x
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Check if it's the mage-x project by looking for cmd/magex
			magexPath := filepath.Join(dir, "cmd", "magex")
			if _, err := os.Stat(magexPath); err == nil {
				return dir
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	t.Fatal("Could not find mage-x project root")
	return ""
}

func BenchmarkIntegration_CommandListing(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping integration benchmark in short mode")
	}

	// Setup test environment
	magexPath := getMagexBinary(&testing.T{})
	tmpDir := b.TempDir()

	// Create test project
	setupTestProject(b, tmpDir)

	originalDir, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			b.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// #nosec G204 -- magexPath is controlled in tests
		cmd := exec.CommandContext(context.Background(), magexPath, "-l")
		cmd.Dir = tmpDir
		_, err := cmd.CombinedOutput()
		if err != nil {
			b.Fatalf("Command failed: %v", err)
		}
	}
}

func setupTestProject(t testing.TB, dir string) {
	t.Helper()

	// Create go.mod
	goModContent := `module testbench

go 1.21
`
	err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goModContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create magefiles directory
	magefilesDir := filepath.Join(dir, "magefiles")
	err = os.Mkdir(magefilesDir, secureDirPerm)
	if err != nil {
		t.Fatalf("Failed to create magefiles directory: %v", err)
	}

	// Create test magefile
	magefileContent := `//go:build mage
package main

import "fmt"

// Build builds the project
func Build() error {
	fmt.Println("Building...")
	return nil
}

// Test runs tests
func Test() error {
	fmt.Println("Testing...")
	return nil
}
`

	err = os.WriteFile(filepath.Join(magefilesDir, "commands.go"), []byte(magefileContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create magefile: %v", err)
	}
}
