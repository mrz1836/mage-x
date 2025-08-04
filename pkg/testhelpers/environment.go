// Package testhelpers provides utilities for testing mage tasks
package testhelpers

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mrz1836/mage-x/pkg/common/env"
	"github.com/mrz1836/mage-x/pkg/common/fileops"
)

// TestEnvironment provides an isolated environment for testing
type TestEnvironment struct {
	t          *testing.T
	rootDir    string
	workDir    string
	origDir    string
	origEnv    map[string]string
	fileOps    fileops.FileOperator
	envManager env.Environment
	cleanup    []func()
}

// NewTestEnvironment creates a new test environment
func NewTestEnvironment(t *testing.T) *TestEnvironment {
	t.Helper()

	// Save original working directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Create temporary root directory
	rootDir, err := os.MkdirTemp("", "mage-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create work directory within root
	workDir := filepath.Join(rootDir, "work")
	if err := os.MkdirAll(workDir, 0o750); err != nil {
		if rmErr := os.RemoveAll(rootDir); rmErr != nil {
			t.Logf("Warning: failed to cleanup root dir after work dir creation failed: %v", rmErr)
		}
		t.Fatalf("Failed to create work dir: %v", err)
	}

	// Save original environment
	origEnv := make(map[string]string)
	for _, e := range os.Environ() {
		parts := strings.SplitN(e, "=", 2)
		if len(parts) == 2 {
			origEnv[parts[0]] = parts[1]
		}
	}

	te := &TestEnvironment{
		t:          t,
		rootDir:    rootDir,
		workDir:    workDir,
		origDir:    origDir,
		origEnv:    origEnv,
		fileOps:    fileops.NewFileOperator(),
		envManager: env.NewEnvironment(),
		cleanup:    []func(){},
	}

	// Change to work directory
	if err := os.Chdir(workDir); err != nil {
		te.Cleanup()
		t.Fatalf("Failed to change to work dir: %v", err)
	}

	// Register cleanup
	t.Cleanup(te.Cleanup)

	return te
}

// RootDir returns the root directory of the test environment
func (te *TestEnvironment) RootDir() string {
	return te.rootDir
}

// WorkDir returns the working directory of the test environment
func (te *TestEnvironment) WorkDir() string {
	return te.workDir
}

// WriteFile writes a file in the test environment
func (te *TestEnvironment) WriteFile(path, content string) {
	te.t.Helper()

	fullPath := te.AbsPath(path)
	dir := filepath.Dir(fullPath)

	if err := os.MkdirAll(dir, 0o750); err != nil {
		te.t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	if err := os.WriteFile(fullPath, []byte(content), 0o600); err != nil {
		te.t.Fatalf("Failed to write file %s: %v", path, err)
	}
}

// ReadFile reads a file from the test environment
func (te *TestEnvironment) ReadFile(path string) string {
	te.t.Helper()

	fullPath := te.AbsPath(path)
	// Clean path for security
	cleanPath := filepath.Clean(fullPath)
	data, err := os.ReadFile(cleanPath)
	if err != nil {
		te.t.Fatalf("Failed to read file %s: %v", path, err)
	}

	return string(data)
}

// FileExists checks if a file exists in the test environment
func (te *TestEnvironment) FileExists(path string) bool {
	fullPath := te.AbsPath(path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// MkdirAll creates directories in the test environment
func (te *TestEnvironment) MkdirAll(path string) {
	te.t.Helper()

	fullPath := te.AbsPath(path)
	if err := os.MkdirAll(fullPath, 0o750); err != nil {
		te.t.Fatalf("Failed to create directories %s: %v", path, err)
	}
}

// AbsPath returns the absolute path within the test environment
func (te *TestEnvironment) AbsPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(te.workDir, path)
}

// SetEnv sets an environment variable for the test
func (te *TestEnvironment) SetEnv(key, value string) {
	te.t.Helper()

	// Save original value if not already saved
	if _, saved := te.origEnv[key]; !saved {
		te.origEnv[key] = os.Getenv(key)
	}

	if err := os.Setenv(key, value); err != nil {
		te.t.Fatalf("Failed to set env var %s: %v", key, err)
	}
}

// UnsetEnv unsets an environment variable for the test
func (te *TestEnvironment) UnsetEnv(key string) {
	te.t.Helper()

	// Save original value if not already saved
	if _, saved := te.origEnv[key]; !saved {
		te.origEnv[key] = os.Getenv(key)
	}

	if err := os.Unsetenv(key); err != nil {
		te.t.Fatalf("Failed to unset env var %s: %v", key, err)
	}
}

// Chdir changes the working directory within the test environment
func (te *TestEnvironment) Chdir(path string) {
	te.t.Helper()

	fullPath := te.AbsPath(path)
	if err := os.Chdir(fullPath); err != nil {
		te.t.Fatalf("Failed to change directory to %s: %v", path, err)
	}
}

// Run executes a function in the test environment
func (te *TestEnvironment) Run(fn func()) {
	fn()
}

// RunWithError executes a function that returns an error
func (te *TestEnvironment) RunWithError(fn func() error) error {
	return fn()
}

// CaptureOutput captures stdout during function execution
func (te *TestEnvironment) CaptureOutput(fn func()) string {
	te.t.Helper()

	// Save original stdout
	origStdout := os.Stdout
	defer func() { os.Stdout = origStdout }()

	// Create pipe
	r, w, err := os.Pipe()
	if err != nil {
		te.t.Fatalf("Failed to create pipe: %v", err)
	}

	// Set stdout to pipe writer
	os.Stdout = w

	// Run function
	fn()

	// Close writer and restore stdout
	if closeErr := w.Close(); closeErr != nil {
		te.t.Logf("Warning: failed to close pipe writer: %v", closeErr)
	}
	os.Stdout = origStdout

	// Read output
	output, err := io.ReadAll(r)
	if err != nil {
		te.t.Fatalf("Failed to read output: %v", err)
	}

	return string(output)
}

// CaptureError captures both stdout and stderr during function execution
func (te *TestEnvironment) CaptureError(fn func()) (stdout, stderr string) {
	te.t.Helper()

	// Save original stdout and stderr
	origStdout := os.Stdout
	origStderr := os.Stderr
	defer func() {
		os.Stdout = origStdout
		os.Stderr = origStderr
	}()

	// Create pipes
	rOut, wOut, err := os.Pipe()
	if err != nil {
		te.t.Fatalf("Failed to create stdout pipe: %v", err)
	}

	rErr, wErr, pipeErr := os.Pipe()
	if pipeErr != nil {
		te.t.Fatalf("Failed to create stderr pipe: %v", pipeErr)
	}

	// Set stdout and stderr to pipe writers
	os.Stdout = wOut
	os.Stderr = wErr

	// Run function
	fn()

	// Close writers and restore
	if err := wOut.Close(); err != nil {
		te.t.Logf("Warning: failed to close stdout pipe writer: %v", err)
	}
	if err := wErr.Close(); err != nil {
		te.t.Logf("Warning: failed to close stderr pipe writer: %v", err)
	}
	os.Stdout = origStdout
	os.Stderr = origStderr

	// Read outputs
	stdoutBytes, readErr := io.ReadAll(rOut)
	if readErr != nil {
		te.t.Fatalf("Failed to read stdout: %v", readErr)
	}

	stderrBytes, readErr := io.ReadAll(rErr)
	if readErr != nil {
		te.t.Fatalf("Failed to read stderr: %v", readErr)
	}

	return string(stdoutBytes), string(stderrBytes)
}

// AssertFileContains asserts that a file contains expected content
func (te *TestEnvironment) AssertFileContains(path, expected string) {
	te.t.Helper()

	content := te.ReadFile(path)
	if !strings.Contains(content, expected) {
		te.t.Errorf("File %s does not contain expected content.\nExpected: %s\nActual: %s",
			path, expected, content)
	}
}

// AssertFileNotContains asserts that a file does not contain content
func (te *TestEnvironment) AssertFileNotContains(path, unexpected string) {
	te.t.Helper()

	content := te.ReadFile(path)
	if strings.Contains(content, unexpected) {
		te.t.Errorf("File %s contains unexpected content: %s", path, unexpected)
	}
}

// AssertFileExists asserts that a file exists
func (te *TestEnvironment) AssertFileExists(path string) {
	te.t.Helper()

	if !te.FileExists(path) {
		te.t.Errorf("Expected file %s to exist, but it doesn't", path)
	}
}

// AssertFileNotExists asserts that a file does not exist
func (te *TestEnvironment) AssertFileNotExists(path string) {
	te.t.Helper()

	if te.FileExists(path) {
		te.t.Errorf("Expected file %s to not exist, but it does", path)
	}
}

// AssertDirExists asserts that a directory exists
func (te *TestEnvironment) AssertDirExists(path string) {
	te.t.Helper()

	fullPath := te.AbsPath(path)
	info, err := os.Stat(fullPath)
	if err != nil {
		te.t.Errorf("Expected directory %s to exist, but it doesn't", path)
		return
	}

	if !info.IsDir() {
		te.t.Errorf("Expected %s to be a directory, but it's a file", path)
	}
}

// AssertNoError asserts that an error is nil
func (te *TestEnvironment) AssertNoError(err error) {
	te.t.Helper()

	if err != nil {
		te.t.Errorf("Expected no error, but got: %v", err)
	}
}

// AssertError asserts that an error is not nil
func (te *TestEnvironment) AssertError(err error) {
	te.t.Helper()

	if err == nil {
		te.t.Error("Expected an error, but got nil")
	}
}

// AssertErrorContains asserts that an error contains expected text
func (te *TestEnvironment) AssertErrorContains(err error, expected string) {
	te.t.Helper()

	if err == nil {
		te.t.Errorf("Expected error containing '%s', but got nil", expected)
		return
	}

	if !strings.Contains(err.Error(), expected) {
		te.t.Errorf("Expected error to contain '%s', but got: %v", expected, err)
	}
}

// AddCleanup adds a cleanup function to be called during cleanup
func (te *TestEnvironment) AddCleanup(fn func()) {
	te.cleanup = append(te.cleanup, fn)
}

// Cleanup cleans up the test environment
func (te *TestEnvironment) Cleanup() {
	// Run cleanup functions in reverse order
	for i := len(te.cleanup) - 1; i >= 0; i-- {
		te.cleanup[i]()
	}

	// Restore original directory
	if err := os.Chdir(te.origDir); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to restore original directory: %v\n", err)
	}

	// Restore original environment
	for key, value := range te.origEnv {
		if value == "" {
			if err := os.Unsetenv(key); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to unset env var %s: %v\n", key, err)
			}
		} else {
			if err := os.Setenv(key, value); err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to restore env var %s: %v\n", key, err)
			}
		}
	}

	// Remove temporary directory
	if err := os.RemoveAll(te.rootDir); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Warning: failed to remove temp directory %s: %v\n", te.rootDir, err)
	}
}

// CreateGoModule creates a go.mod file in the test environment
func (te *TestEnvironment) CreateGoModule(module string) {
	te.t.Helper()

	content := fmt.Sprintf(`module %s

go 1.24

require github.com/magefile/mage v1.15.0
`, module)

	te.WriteFile("go.mod", content)
}

// CreateMagefile creates a sample magefile in the test environment
func (te *TestEnvironment) CreateMagefile(content string) {
	te.t.Helper()

	if content == "" {
		content = `// +build mage

package main

import (
	"fmt"
	"github.com/magefile/mage/mg"
)

// Build builds the project
func Build() error {
	fmt.Println("Building...")
	return nil
}

// Test runs tests
func Test() error {
	mg.Deps(Build)
	fmt.Println("Testing...")
	return nil
}
`
	}

	te.WriteFile("magefile.go", content)
}

// CreateConfigFile creates a mage config file
func (te *TestEnvironment) CreateConfigFile(content string) {
	te.t.Helper()

	if content == "" {
		content = `project:
  name: test-project
  version: 1.0.0

build:
  output: bin/
  flags:
    - -v

test:
  coverage: true
`
	}

	te.WriteFile(".mage.yaml", content)
}

// SetupGitRepo initializes a git repository in the test environment
func (te *TestEnvironment) SetupGitRepo() {
	te.t.Helper()

	cmds := [][]string{
		{"git", "init"},
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "user.name", "Test User"},
	}

	for _, cmd := range cmds {
		ctx := context.Background()
		// #nosec G204 -- test helper with controlled input
		if err := exec.CommandContext(ctx, cmd[0], cmd[1:]...).Run(); err != nil {
			te.t.Fatalf("Failed to run %v: %v", cmd, err)
		}
	}
}

// GitAdd adds files to git
func (te *TestEnvironment) GitAdd(files ...string) {
	te.t.Helper()

	args := append([]string{"add"}, files...)
	ctx := context.Background()
	// #nosec G204 -- test helper with controlled input
	if err := exec.CommandContext(ctx, "git", args...).Run(); err != nil {
		te.t.Fatalf("Failed to git add: %v", err)
	}
}

// GitCommit creates a git commit
func (te *TestEnvironment) GitCommit(message string) {
	te.t.Helper()

	ctx := context.Background()
	if err := exec.CommandContext(ctx, "git", "commit", "-m", message).Run(); err != nil {
		te.t.Fatalf("Failed to git commit: %v", err)
	}
}
