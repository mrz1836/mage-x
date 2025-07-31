// Package testutil provides testing utilities and helpers for mage operations.
package testutil

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mrz1836/go-mage/pkg/common/fileops"
)

// TestingInterface covers common functionality of *testing.T and *testing.B
type TestingInterface interface {
	TempDir() string
	Helper()
	Fatalf(format string, args ...interface{})
}

// TestEnvironment provides a clean test environment with temporary directories
type TestEnvironment struct {
	TempDir string
	OrigDir string
	Runner  *MockRunner
	Builder *MockBuilder
	t       TestingInterface
}

// NewTestEnvironment creates a new isolated test environment
func NewTestEnvironment(t TestingInterface) *TestEnvironment {
	// Get original directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Create temp directory
	tempDir := t.TempDir()

	// Change to temp directory
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Set up mock runner
	runner, builder := NewMockRunner()

	return &TestEnvironment{
		TempDir: tempDir,
		OrigDir: origDir,
		Runner:  runner,
		Builder: builder,
		t:       t,
	}
}

// Cleanup restores the original environment
func (env *TestEnvironment) Cleanup() {
	if err := os.Chdir(env.OrigDir); err != nil {
		// Log error but don't fail cleanup
		fmt.Fprintf(os.Stderr, "failed to restore original directory %s: %v\n", env.OrigDir, err)
	}
}

// RunnerSetter is a function type for setting command runners
type RunnerSetter func(runner interface{}) error

// RunnerGetter is a function type for getting command runners
type RunnerGetter func() interface{}

// WithMockRunner executes a function with the mock runner active
// The setter and getter functions should be passed from the calling package
func (env *TestEnvironment) WithMockRunner(setter RunnerSetter, getter RunnerGetter, fn func() error) error {
	// Save original runner
	originalRunner := getter()
	defer func() {
		if err := setter(originalRunner); err != nil {
			// Log error but don't fail cleanup
			fmt.Fprintf(os.Stderr, "failed to restore original runner: %v\n", err)
		}
	}()

	// Set mock runner
	if err := setter(env.Runner); err != nil {
		return fmt.Errorf("failed to set mock runner: %w", err)
	}

	// Execute function
	return fn()
}

// CreateGoMod creates a go.mod file in the test environment
func (env *TestEnvironment) CreateGoMod(moduleName string) {
	content := `module ` + moduleName + `

go 1.24

require (
	github.com/magefile/mage v1.15.0
	github.com/stretchr/testify v1.10.0
)
`
	fileOps := fileops.New()
	err := fileOps.File.WriteFile("go.mod", []byte(content), 0o644)
	if err != nil {
		env.t.Fatalf("Error: %v", err)
	}
}

// CreateMageConfig creates a .mage.yaml configuration file
func (env *TestEnvironment) CreateMageConfig(config string) {
	fileOps := fileops.New()
	err := fileOps.File.WriteFile(".mage.yaml", []byte(config), 0o644)
	if err != nil {
		env.t.Fatalf("Error: %v", err)
	}
}

// CreateProjectStructure creates a basic project structure
func (env *TestEnvironment) CreateProjectStructure() {
	dirs := []string{
		"cmd/app",
		"pkg/utils",
		"bin",
		"docs",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0o750)
		if err != nil {
			env.t.Fatalf("Error: %v", err)
		}
	}

	// Create main.go
	mainGo := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`
	fileOps := fileops.New()
	err := fileOps.File.WriteFile("cmd/app/main.go", []byte(mainGo), 0o644)
	if err != nil {
		env.t.Fatalf("Error: %v", err)
	}
}

// CreateFile creates a file with the given content
func (env *TestEnvironment) CreateFile(path, content string) {
	dir := filepath.Dir(path)
	if dir != "." {
		err := os.MkdirAll(dir, 0o750)
		if err != nil {
			env.t.Fatalf("Error: %v", err)
		}
	}

	fileOps := fileops.New()
	err := fileOps.File.WriteFile(path, []byte(content), 0o644)
	if err != nil {
		env.t.Fatalf("Error: %v", err)
	}
}

// FileExists checks if a file exists
func (env *TestEnvironment) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ReadFile reads a file and returns its content
func (env *TestEnvironment) ReadFile(path string) string {
	fileOps := fileops.New()
	content, err := fileOps.File.ReadFile(path)
	if err != nil {
		env.t.Fatalf("Error: %v", err)
	}
	return string(content)
}

// StandardMocks sets up common mock expectations for standard operations
type StandardMocks struct {
	env *TestEnvironment
}

// StandardMocks creates standard mock configurations
func (env *TestEnvironment) StandardMocks() *StandardMocks {
	return &StandardMocks{env: env}
}

// ForBuild sets up mocks for build operations
func (sm *StandardMocks) ForBuild() *StandardMocks {
	sm.env.Builder.
		ExpectVersion("v1.0.0").
		ExpectGoCommand("build", nil).
		ExpectGoCommand("clean", nil)
	return sm
}

// ForTest sets up mocks for test operations
func (sm *StandardMocks) ForTest() *StandardMocks {
	sm.env.Builder.
		ExpectGoCommand("test", nil).
		ExpectGoCommand("vet", nil)
	return sm
}

// ForLint sets up mocks for linting operations
func (sm *StandardMocks) ForLint() *StandardMocks {
	sm.env.Builder.
		ExpectAnyCommand(nil) // golangci-lint and related commands
	return sm
}

// ForGit sets up mocks for git operations
func (sm *StandardMocks) ForGit() *StandardMocks {
	sm.env.Builder.
		ExpectGitCommand("status", "", nil).
		ExpectGitCommand("add", "", nil).
		ExpectGitCommand("commit", "", nil).
		ExpectGitCommand("push", "", nil).
		ExpectGitCommand("pull", "", nil)
	return sm
}

// ForAll sets up mocks for all common operations
func (sm *StandardMocks) ForAll() *StandardMocks {
	return sm.ForBuild().ForTest().ForLint().ForGit()
}
