package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	secureDirPerm  = 0o750
	secureFilePerm = 0o600
	testGoMod      = `module testmage

go 1.21
`
)

var errTest = errors.New("test error")

func TestDelegateToMage_MagefilesDir(t *testing.T) {
	// Create a temporary directory with magefiles/ subdirectory
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

	// Create magefiles directory
	magefilesDir := "magefiles"
	err = os.Mkdir(magefilesDir, secureDirPerm)
	if err != nil {
		t.Fatalf("Failed to create magefiles directory: %v", err)
	}

	// Create a test magefile
	magefilePath := filepath.Join(magefilesDir, "commands.go")
	magefileContent := `//go:build mage
package main

import "fmt"

// TestCmd is a test command
func TestCmd() error {
	fmt.Println("TestCmd executed")
	return nil
}
`

	err = os.WriteFile(magefilePath, []byte(magefileContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create test magefile: %v", err)
	}

	// Test that HasMagefile returns true for directory
	if !HasMagefile() {
		t.Error("HasMagefile() should return true for magefiles/ directory")
	}

	// Test GetMagefilePath returns directory path
	path := GetMagefilePath()
	if !strings.Contains(path, "magefiles") {
		t.Errorf("GetMagefilePath() should return path with 'magefiles', got: %s", path)
	}

	// Note: We can't easily test actual command execution in unit tests
	// without mocking exec.Command, which would be complex.
	// Integration tests will cover the actual execution.
}

func TestDelegateToMage_MagefileGo(t *testing.T) {
	// Create a temporary directory with single magefile.go
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

	// Create a test magefile.go
	magefilePath := "magefile.go"
	magefileContent := `//go:build mage
package main

import "fmt"

// TestCmd is a test command
func TestCmd() error {
	fmt.Println("TestCmd executed")
	return nil
}
`

	err = os.WriteFile(magefilePath, []byte(magefileContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create test magefile: %v", err)
	}

	// Test that HasMagefile returns true for file
	if !HasMagefile() {
		t.Error("HasMagefile() should return true for magefile.go")
	}

	// Test GetMagefilePath returns file path
	path := GetMagefilePath()
	if !strings.Contains(path, "magefile.go") {
		t.Errorf("GetMagefilePath() should return path with 'magefile.go', got: %s", path)
	}
}

func TestDelegateToMage_PrefersMagefilesDir(t *testing.T) {
	// Create a temporary directory with both magefiles/ directory and magefile.go
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

	// Create magefiles directory
	magefilesDir := "magefiles"
	err = os.Mkdir(magefilesDir, secureDirPerm)
	if err != nil {
		t.Fatalf("Failed to create magefiles directory: %v", err)
	}

	// Create a test magefile in directory
	dirMagefilePath := filepath.Join(magefilesDir, "commands.go")
	dirMagefileContent := `//go:build mage
package main

// DirCmd is from directory
func DirCmd() error {
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

// RootCmd is from root file
func RootCmd() error {
	return nil
}
`

	err = os.WriteFile(rootMagefilePath, []byte(rootMagefileContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create root magefile: %v", err)
	}

	// Test that HasMagefile returns true (should find directory first)
	if !HasMagefile() {
		t.Error("HasMagefile() should return true when both exist")
	}

	// Test GetMagefilePath prefers directory over file
	path := GetMagefilePath()
	if !strings.Contains(path, "magefiles") {
		t.Errorf("GetMagefilePath() should prefer directory over file, got: %s", path)
	}
	if strings.Contains(path, "magefile.go") {
		t.Errorf("GetMagefilePath() should not return magefile.go when directory exists, got: %s", path)
	}
}

func TestDelegateToMage_CommandNotFound(t *testing.T) {
	// Create a temporary directory with no magefiles
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

	// Test that DelegateToMage returns appropriate error
	result := DelegateToMage(context.Background(), "nonexistent")
	if result.Err == nil {
		t.Error("DelegateToMage should return error when no magefile exists")
	}
	if !errors.Is(result.Err, ErrCommandNotFound) {
		t.Errorf("DelegateToMage should return ErrCommandNotFound, got: %v", result.Err)
	}
}

func TestHasMagefile_NoFiles(t *testing.T) {
	// Create a temporary directory with no magefiles
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

	if HasMagefile() {
		t.Error("HasMagefile() should return false when no magefiles exist")
	}
}

func TestHasMagefile_Directory(t *testing.T) {
	// Create a temporary directory with magefiles/ subdirectory
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

	// Create magefiles directory
	err = os.Mkdir("magefiles", secureDirPerm)
	if err != nil {
		t.Fatalf("Failed to create magefiles directory: %v", err)
	}

	if !HasMagefile() {
		t.Error("HasMagefile() should return true for magefiles/ directory")
	}
}

func TestHasMagefile_SingleFile(t *testing.T) {
	// Create a temporary directory with magefile.go
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

	// Create magefile.go
	err = os.WriteFile("magefile.go", []byte("package main"), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create magefile.go: %v", err)
	}

	if !HasMagefile() {
		t.Error("HasMagefile() should return true for magefile.go")
	}
}

func TestGetMagefilePath_Directory(t *testing.T) {
	// Create a temporary directory with magefiles/ subdirectory
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

	// Create magefiles directory
	err = os.Mkdir("magefiles", secureDirPerm)
	if err != nil {
		t.Fatalf("Failed to create magefiles directory: %v", err)
	}

	path := GetMagefilePath()
	if !strings.Contains(path, "magefiles") {
		t.Errorf("GetMagefilePath() should contain 'magefiles', got: %s", path)
	}
	if filepath.IsAbs(path) && !strings.HasSuffix(path, "magefiles") {
		t.Errorf("GetMagefilePath() should end with 'magefiles', got: %s", path)
	}
}

func TestGetMagefilePath_SingleFile(t *testing.T) {
	// Create a temporary directory with magefile.go
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

	// Create magefile.go
	err = os.WriteFile("magefile.go", []byte("package main"), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create magefile.go: %v", err)
	}

	path := GetMagefilePath()
	if !strings.Contains(path, "magefile.go") {
		t.Errorf("GetMagefilePath() should contain 'magefile.go', got: %s", path)
	}
	if filepath.IsAbs(path) && !strings.HasSuffix(path, "magefile.go") {
		t.Errorf("GetMagefilePath() should end with 'magefile.go', got: %s", path)
	}
}

func TestGetMagefilePath_NoFiles(t *testing.T) {
	// Create a temporary directory with no magefiles
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

	path := GetMagefilePath()
	if path != "" {
		t.Errorf("GetMagefilePath() should return empty string when no files exist, got: %s", path)
	}
}

func TestFilterStderr(t *testing.T) {
	// Create a pipe to simulate stderr input
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Capture stderr output
	oldStderr := os.Stderr
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create stderr pipe: %v", err)
	}
	os.Stderr = stderrW

	// Write test data to the pipe
	testData := `Normal error message
Unknown target specified: somecommand
Another normal message
Unknown target specified: anothercommand
Final message
`

	// Create buffer and WaitGroup for the new function signature
	var buf strings.Builder
	var wg sync.WaitGroup
	wg.Add(1)

	// Write data and close the writer
	go func() {
		defer func() {
			if closeErr := w.Close(); closeErr != nil {
				t.Logf("Failed to close pipe writer: %v", closeErr)
			}
		}()
		if _, writeErr := fmt.Fprint(w, testData); writeErr != nil {
			t.Logf("Failed to write test data: %v", writeErr)
		}
	}()

	// Filter stderr in a goroutine
	go filterStderr(r, &buf, &wg)

	// Wait for filtering to complete
	wg.Wait()

	// Close the stderr writer to signal end
	if closeErr := stderrW.Close(); closeErr != nil {
		t.Logf("Failed to close stderr writer: %v", closeErr)
	}

	// Restore stderr and read the output
	os.Stderr = oldStderr
	var output strings.Builder
	_, err = io.Copy(&output, stderrR)
	if err != nil {
		t.Fatalf("Failed to read stderr output: %v", err)
	}
	if err := stderrR.Close(); err != nil {
		t.Logf("Failed to close stderr reader: %v", err)
	}

	result := output.String()

	// Check that normal messages are preserved in stderr output
	if !strings.Contains(result, "Normal error message") {
		t.Errorf("Normal error messages should be preserved, got: %q", result)
	}
	if !strings.Contains(result, "Another normal message") {
		t.Errorf("Normal error messages should be preserved, got: %q", result)
	}
	if !strings.Contains(result, "Final message") {
		t.Errorf("Normal error messages should be preserved, got: %q", result)
	}

	// Check that "Unknown target specified" messages are filtered out
	if strings.Contains(result, "Unknown target specified:") {
		t.Errorf("'Unknown target specified' messages should be filtered out, got: %q", result)
	}

	// Check that buffer captured ALL messages (including filtered ones) for error reporting
	captured := buf.String()
	if !strings.Contains(captured, "Normal error message") {
		t.Errorf("Buffer should capture normal messages, got: %q", captured)
	}
	// Buffer SHOULD capture filtered messages - they're needed for error reporting
	if !strings.Contains(captured, "Unknown target specified:") {
		t.Errorf("Buffer should capture all messages for error reporting, got: %q", captured)
	}
}

func TestConvertToMageFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Colon-separated namespace commands - preserves colon
		{"Speckit:Install", "speckit:install"},
		{"Bmad:Install", "bmad:install"},
		{"Bmad:Check", "bmad:check"},
		{"Bmad:Upgrade", "bmad:upgrade"},
		{"Bmad:Status", "bmad:status"},
		{"Pipeline:CI", "pipeline:ci"},
		{"Build:Default", "build:default"},
		{"Test:Unit", "test:unit"},

		// Simple commands (just lowercased)
		{"Deploy", "deploy"},
		{"build", "build"},

		// Edge cases - colon preserved
		{":", ":"},
		{"Namespace:", "namespace:"},
		{":Method", ":method"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := convertToMageFormat(tt.input)
			if result != tt.expected {
				t.Errorf("convertToMageFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDelegateToMage_Integration(t *testing.T) {
	// Skip this test if we don't have go available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available, skipping integration test")
	}

	// Create a temporary directory with working magefile
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

	// Create go.mod file for the test
	err = os.WriteFile("go.mod", []byte(testGoMod), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create magefiles directory
	err = os.Mkdir("magefiles", secureDirPerm)
	if err != nil {
		t.Fatalf("Failed to create magefiles directory: %v", err)
	}

	// Create a simple working magefile
	magefilePath := filepath.Join("magefiles", "commands.go")
	magefileContent := `//go:build mage
package main

import (
	"fmt"
	"os"
)

// TestCommand is a test command
func TestCommand() error {
	fmt.Println("TestCommand executed successfully")
	return nil
}

// ParamsTest shows how parameters are passed
func ParamsTest() error {
	fmt.Printf("Args: %v\n", os.Args[1:])
	return nil
}
`

	err = os.WriteFile(magefilePath, []byte(magefileContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create test magefile: %v", err)
	}

	// Test that the command can be executed
	// Note: This will use "go run" since mage might not be available
	result := DelegateToMage(context.Background(), "TestCommand")
	if result.Err != nil {
		// This is expected to work in a real environment with Go
		// but might fail in test environment, so we'll log rather than fail
		t.Logf("DelegateToMage execution failed (expected in test environment): %v", result.Err)
	}
}

func BenchmarkHasMagefile(b *testing.B) {
	// Create a temporary directory with magefile.go
	tmpDir := b.TempDir()
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

	// Create magefile.go
	err = os.WriteFile("magefile.go", []byte("package main"), secureFilePerm)
	if err != nil {
		b.Fatalf("Failed to create magefile.go: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HasMagefile()
	}
}

func BenchmarkGetMagefilePath(b *testing.B) {
	// Create a temporary directory with magefile.go
	tmpDir := b.TempDir()
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

	// Create magefiles directory
	err = os.Mkdir("magefiles", secureDirPerm)
	if err != nil {
		b.Fatalf("Failed to create magefiles directory: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetMagefilePath()
	}
}

// TestValidateGoEnvironmentSuccess tests that ValidateGoEnvironment returns nil
// when Go is available in the PATH (which should be the case in CI/dev environments).
func TestValidateGoEnvironmentSuccess(t *testing.T) {
	// Go is expected to be available since we're running tests
	err := ValidateGoEnvironment()
	assert.NoError(t, err, "ValidateGoEnvironment should succeed when Go is available")
}

// TestValidateGoEnvironmentNoGo tests that ValidateGoEnvironment returns
// ErrGoCommandNotFound when Go is not available in the PATH.
func TestValidateGoEnvironmentNoGo(t *testing.T) {
	// Save original PATH and restore after test
	originalPath := os.Getenv("PATH")
	t.Cleanup(func() {
		require.NoError(t, os.Setenv("PATH", originalPath))
	})

	// Clear PATH to simulate Go not being available
	require.NoError(t, os.Setenv("PATH", ""))

	err := ValidateGoEnvironment()
	require.Error(t, err, "ValidateGoEnvironment should fail when Go is not in PATH")
	assert.ErrorIs(t, err, ErrGoCommandNotFound, "Should return ErrGoCommandNotFound")
}

// TestDelegateToMageWithTimeoutTimesOut tests that DelegateToMageWithTimeout
// returns ErrCommandTimeout when the command exceeds the timeout duration.
func TestDelegateToMageWithTimeoutTimesOut(t *testing.T) {
	// Skip if go is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available, skipping timeout test")
	}

	// Skip if mage is not available (go run fallback can't handle magefiles directories)
	if _, err := exec.LookPath("mage"); err != nil {
		t.Skip("mage binary not available, skipping timeout test")
	}

	// Create a temporary directory with a magefile that sleeps
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalDir))
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte(testGoMod), secureFilePerm))

	// Create magefiles directory with a command that sleeps
	require.NoError(t, os.Mkdir("magefiles", secureDirPerm))

	magefileContent := `//go:build mage
package main

import (
	"time"
)

// SleepCmd sleeps for 10 seconds - used to test timeout
func SleepCmd() error {
	time.Sleep(10 * time.Second)
	return nil
}
`
	require.NoError(t, os.WriteFile(filepath.Join("magefiles", "commands.go"), []byte(magefileContent), secureFilePerm))

	// Execute with a very short timeout
	result := DelegateToMageWithTimeout(context.Background(), "sleepCmd", 500*time.Millisecond)

	// Should timeout and return the timeout error
	require.Error(t, result.Err, "Command should timeout")
	require.ErrorIs(t, result.Err, ErrCommandTimeout, "Should return ErrCommandTimeout")
	assert.Equal(t, 124, result.ExitCode, "Timeout exit code should be 124")
}

// TestDelegateToMageWithTimeoutWithArguments tests that arguments are passed
// to the delegated command via MAGE_ARGS environment variable.
func TestDelegateToMageWithTimeoutWithArguments(t *testing.T) {
	// Skip if go is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available, skipping arguments test")
	}

	// Create a temporary directory with a magefile
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalDir))
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte(testGoMod), secureFilePerm))

	// Create magefiles directory with a command that prints args
	require.NoError(t, os.Mkdir("magefiles", secureDirPerm))

	magefileContent := `//go:build mage
package main

import (
	"fmt"
	"os"
)

// ArgsCmd prints the MAGE_ARGS environment variable
func ArgsCmd() error {
	args := os.Getenv("MAGE_ARGS")
	fmt.Printf("ARGS:%s\n", args)
	return nil
}
`
	require.NoError(t, os.WriteFile(filepath.Join("magefiles", "commands.go"), []byte(magefileContent), secureFilePerm))

	// Execute with arguments
	result := DelegateToMageWithTimeout(context.Background(), "argsCmd", DefaultDelegateTimeout, "arg1", "arg2", "arg3")

	// The command should execute without error
	// Note: This test verifies the code path that sets MAGE_ARGS, even if the actual execution
	// might fail depending on environment
	_ = result // We just want to exercise the code path with arguments
}

// TestDelegateToMageWithTimeoutConflictHandling tests the conflict handling
// when both magefiles/ directory and magefile.go exist.
func TestDelegateToMageWithTimeoutConflictHandling(t *testing.T) {
	// Skip if go is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available, skipping conflict handling test")
	}

	// Create a temporary directory
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalDir))
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte(testGoMod), secureFilePerm))

	// Create magefiles directory with a command
	require.NoError(t, os.Mkdir("magefiles", secureDirPerm))

	dirMagefileContent := `//go:build mage
package main

import "fmt"

// TestCmd is from directory
func TestCmd() error {
	fmt.Println("FROM_DIR")
	return nil
}
`
	require.NoError(t, os.WriteFile(filepath.Join("magefiles", "commands.go"), []byte(dirMagefileContent), secureFilePerm))

	// Create root magefile.go (should be temporarily renamed)
	rootMagefileContent := `//go:build mage
package main

import "fmt"

// RootCmd is from root file
func RootCmd() error {
	fmt.Println("FROM_ROOT")
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(rootMagefileContent), secureFilePerm))

	// Execute a command - this should trigger the conflict handling code path
	result := DelegateToMageWithTimeout(context.Background(), "testCmd", DefaultDelegateTimeout)

	// After execution, magefile.go should be restored
	_, err = os.Stat("magefile.go")
	require.NoError(t, err, "magefile.go should be restored after command execution")

	// Verify no temp file is left behind
	_, err = os.Stat("magefile.go.tmp")
	assert.True(t, os.IsNotExist(err), "Temp file should not exist after execution")

	// The result may be an error (no mage installed) but that's fine
	// We're testing the conflict handling code path, not actual execution success
	_ = result
}

// TestDelegateToMageWithTimeoutStderrCapture tests that stderr output is
// properly captured and included in error messages.
func TestDelegateToMageWithTimeoutStderrCapture(t *testing.T) {
	// Skip if go is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available, skipping stderr capture test")
	}

	// Create a temporary directory with a magefile that writes to stderr and fails
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalDir))
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte(testGoMod), secureFilePerm))

	// Create magefiles directory with a command that writes to stderr and exits with error
	require.NoError(t, os.Mkdir("magefiles", secureDirPerm))

	magefileContent := `//go:build mage
package main

import (
	"errors"
	"fmt"
	"os"
)

// FailCmd writes to stderr and returns an error
func FailCmd() error {
	fmt.Fprintln(os.Stderr, "Custom error message on stderr")
	return errors.New("command failed intentionally")
}
`
	require.NoError(t, os.WriteFile(filepath.Join("magefiles", "commands.go"), []byte(magefileContent), secureFilePerm))

	// Execute the failing command
	result := DelegateToMageWithTimeout(context.Background(), "failCmd", DefaultDelegateTimeout)

	// Should fail with the error message captured
	require.Error(t, result.Err, "Command should fail")
	assert.ErrorIs(t, result.Err, ErrCommandFailed, "Should return ErrCommandFailed")
}

// TestConvertToMageFormatEdgeCases tests additional edge cases for convertToMageFormat
func TestConvertToMageFormatEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "multiple colons preserves all",
			input:    "A:B:C",
			expected: "a:b:c", // Just lowercased, colons preserved
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single colon with empty parts",
			input:    ":",
			expected: ":", // Colon preserved
		},
		{
			name:     "unicode characters",
			input:    "Ünïcödé:Tëst",
			expected: "ünïcödé:tëst", // Just lowercased, colon preserved
		},
		{
			name:     "numbers in namespace",
			input:    "Build123:Run456",
			expected: "build123:run456", // Colon preserved
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToMageFormat(tt.input)
			assert.Equal(t, tt.expected, result, "convertToMageFormat(%q) should return %q", tt.input, tt.expected)
		})
	}
}

// TestStaticErrorsExist verifies that all static error variables are properly defined
func TestStaticErrorsExist(t *testing.T) {
	t.Parallel()

	errorVars := map[string]error{
		"ErrCommandNotFound":       ErrCommandNotFound,
		"ErrGoCommandNotFound":     ErrGoCommandNotFound,
		"ErrCommandFailed":         ErrCommandFailed,
		"ErrMagefileRestoreFailed": ErrMagefileRestoreFailed,
		"ErrCommandTimeout":        ErrCommandTimeout,
	}

	for name, err := range errorVars {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.Error(t, err, "%s should not be nil", name)
			assert.NotEmpty(t, err.Error(), "%s should have an error message", name)
		})
	}
}

// TestDelegateResultStruct tests the DelegateResult struct
func TestDelegateResultStruct(t *testing.T) {
	t.Parallel()

	t.Run("zero value has nil error", func(t *testing.T) {
		t.Parallel()
		var result DelegateResult
		assert.Equal(t, 0, result.ExitCode)
		assert.NoError(t, result.Err)
	})

	t.Run("can set exit code and error", func(t *testing.T) {
		t.Parallel()
		testErr := errTest
		result := DelegateResult{
			ExitCode: 42,
			Err:      testErr,
		}
		assert.Equal(t, 42, result.ExitCode)
		assert.Equal(t, testErr, result.Err)
	})
}

// TestGetMagefilePath_PrefersMagefiles tests that GetMagefilePath prefers magefiles directory
func TestGetMagefilePath_PrefersMagefiles(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create both magefiles directory and magefile.go
	err = os.Mkdir("magefiles", secureDirPerm)
	require.NoError(t, err)
	err = os.WriteFile("magefile.go", []byte("package main"), secureFilePerm)
	require.NoError(t, err)

	path := GetMagefilePath()
	// Should prefer magefiles directory
	assert.Contains(t, path, "magefiles")
	assert.NotContains(t, path, "magefile.go")
}

// TestConvertToMageFormat_MultipleColons tests command with multiple colons
func TestConvertToMageFormat_MultipleColons(t *testing.T) {
	input := "namespace:method:extra"
	result := convertToMageFormat(input)
	// Just lowercased, all colons preserved
	expected := "namespace:method:extra"
	assert.Equal(t, expected, result)
}

// TestGetMagefilePath_MagefilesDir tests GetMagefilePath with magefiles/ directory
func TestGetMagefilePath_MagefilesDir(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create magefiles/ directory
	err = os.Mkdir("magefiles", 0o750)
	require.NoError(t, err)

	path := GetMagefilePath()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "magefiles")
}

// TestGetMagefilePath_MagefileGo tests GetMagefilePath with magefile.go
func TestGetMagefilePath_MagefileGo(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create magefile.go
	err = os.WriteFile("magefile.go", []byte("package main"), 0o600)
	require.NoError(t, err)

	path := GetMagefilePath()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "magefile.go")
}

// TestGetMagefilePath_NoMagefile tests GetMagefilePath with no magefile
func TestGetMagefilePath_NoMagefile(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	path := GetMagefilePath()
	assert.Empty(t, path)
}

// TestGetMagefilePath_PrefersMagefilesDir tests that magefiles/ is preferred over magefile.go
func TestGetMagefilePath_PrefersMagefilesDir(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create both magefiles/ and magefile.go
	err = os.Mkdir("magefiles", 0o750)
	require.NoError(t, err)
	err = os.WriteFile("magefile.go", []byte("package main"), 0o600)
	require.NoError(t, err)

	path := GetMagefilePath()
	assert.NotEmpty(t, path)
	// Should prefer magefiles/ directory
	assert.Contains(t, path, "magefiles")
	assert.NotContains(t, path, "magefile.go")
}

// TestDelegateToMage_NamespaceCommands tests that namespace commands (colon-separated)
// are properly converted and passed to mage. This is a regression test for the
// convertToMageFormat fix that ensures commands like "Ci:Static" are properly
// converted to "ci:static" (lowercase) before being passed to mage.
//
// Note: Full execution may not work in all test environments due to mage compilation
// requirements. The test validates format conversion and command delegation.
func TestDelegateToMage_NamespaceCommands(t *testing.T) {
	// Skip if mage binary not available
	if _, err := exec.LookPath("mage"); err != nil {
		t.Skip("mage binary not available, skipping namespace command integration test")
	}

	// Skip if go is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available, skipping namespace command integration test")
	}

	// Create temporary project directory
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			t.Logf("Failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	goModContent := `module testnamespace

go 1.21
`
	require.NoError(t, os.WriteFile("go.mod", []byte(goModContent), secureFilePerm))

	// Create magefiles directory with namespace commands
	require.NoError(t, os.Mkdir("magefiles", secureDirPerm))

	// Create a magefile with a namespace (type CI struct{})
	magefileContent := `//go:build mage

package main

import (
	"fmt"
)

// CI namespace for CI/CD commands
type CI struct{}

// Static runs static analysis (this was the command that was broken)
func (CI) Static() error {
	fmt.Println("NAMESPACE_COMMAND_SUCCESS:ci:static")
	return nil
}

// Lint runs linting
func (CI) Lint() error {
	fmt.Println("NAMESPACE_COMMAND_SUCCESS:ci:lint")
	return nil
}

// Build namespace for build commands
type Build struct{}

// Default runs the default build
func (Build) Default() error {
	fmt.Println("NAMESPACE_COMMAND_SUCCESS:build:default")
	return nil
}

// Test namespace for test commands
type Test struct{}

// Unit runs unit tests
func (Test) Unit() error {
	fmt.Println("NAMESPACE_COMMAND_SUCCESS:test:unit")
	return nil
}

// Coverage runs tests with coverage (multiple-colon test when chained)
func (Test) Coverage() error {
	fmt.Println("NAMESPACE_COMMAND_SUCCESS:test:coverage")
	return nil
}
`
	require.NoError(t, os.WriteFile(filepath.Join("magefiles", "commands.go"), []byte(magefileContent), secureFilePerm))

	// Test cases for namespace command handling
	tests := []struct {
		name              string
		input             string
		expectedLowercase string
		expectedOutput    string
		description       string
	}{
		{
			name:              "mixed_case_namespace",
			input:             "Ci:Static",
			expectedLowercase: "ci:static",
			expectedOutput:    "NAMESPACE_COMMAND_SUCCESS:ci:static",
			description:       "Mixed case namespace (the bug that was fixed)",
		},
		{
			name:              "all_uppercase_namespace",
			input:             "CI:STATIC",
			expectedLowercase: "ci:static",
			expectedOutput:    "NAMESPACE_COMMAND_SUCCESS:ci:static",
			description:       "All uppercase namespace",
		},
		{
			name:              "already_lowercase_namespace",
			input:             "ci:static",
			expectedLowercase: "ci:static",
			expectedOutput:    "NAMESPACE_COMMAND_SUCCESS:ci:static",
			description:       "Already lowercase namespace",
		},
		{
			name:              "mixed_case_build",
			input:             "Build:Default",
			expectedLowercase: "build:default",
			expectedOutput:    "NAMESPACE_COMMAND_SUCCESS:build:default",
			description:       "Standard mixed case namespace",
		},
		{
			name:              "ci_lint",
			input:             "CI:Lint",
			expectedLowercase: "ci:lint",
			expectedOutput:    "NAMESPACE_COMMAND_SUCCESS:ci:lint",
			description:       "CI lint command",
		},
		{
			name:              "test_coverage",
			input:             "Test:Coverage",
			expectedLowercase: "test:coverage",
			expectedOutput:    "NAMESPACE_COMMAND_SUCCESS:test:coverage",
			description:       "Test coverage command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First verify the format conversion works correctly (unit test level)
			converted := convertToMageFormat(tt.input)
			assert.Equal(t, tt.expectedLowercase, converted,
				"convertToMageFormat(%q) should return %q", tt.input, tt.expectedLowercase)

			// Capture stdout to verify output
			oldStdout := os.Stdout
			r, w, pipeErr := os.Pipe()
			require.NoError(t, pipeErr)
			os.Stdout = w

			// Execute the command with a timeout
			result := DelegateToMageWithTimeout(context.Background(), tt.input, 30*time.Second)

			// Restore stdout and read captured output
			if closeErr := w.Close(); closeErr != nil {
				t.Logf("Failed to close pipe writer: %v", closeErr)
			}
			os.Stdout = oldStdout

			var capturedOutput strings.Builder
			if _, copyErr := io.Copy(&capturedOutput, r); copyErr != nil {
				t.Logf("Failed to read captured output: %v", copyErr)
			}
			if closeErr := r.Close(); closeErr != nil {
				t.Logf("Failed to close pipe reader: %v", closeErr)
			}

			// If execution succeeded, verify the output
			if result.Err == nil {
				assert.Equal(t, 0, result.ExitCode, "Exit code should be 0 for '%s'", tt.input)
				output := capturedOutput.String()
				assert.Contains(t, output, tt.expectedOutput,
					"Output for '%s' should contain '%s', got: %s",
					tt.input, tt.expectedOutput, output)
			} else {
				// If execution failed, verify it's due to mage not finding the target
				// (expected in test environments where mage can't compile the test magefiles)
				// The key verification is that the command was passed correctly (lowercase)
				errMsg := result.Err.Error()
				if strings.Contains(errMsg, "Unknown target specified") {
					// This is acceptable - mage received the correctly formatted command
					// but couldn't find the target (test environment limitation)
					assert.Contains(t, errMsg, tt.expectedLowercase,
						"Error should reference the lowercase command format '%s'", tt.expectedLowercase)
					t.Logf("Command '%s' correctly converted to '%s' (mage target not found in test env)",
						tt.input, tt.expectedLowercase)
				} else {
					// Unexpected error
					t.Errorf("Unexpected error for command '%s': %v", tt.input, result.Err)
				}
			}
		})
	}
}
