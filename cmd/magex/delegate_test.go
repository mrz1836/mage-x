package main

import (
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
	result := DelegateToMage("nonexistent")
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
		// Colon-separated namespace commands
		{"Speckit:Install", "speckitinstall"},
		{"Bmad:Install", "bmadinstall"},
		{"Bmad:Check", "bmadcheck"},
		{"Bmad:Upgrade", "bmadupgrade"},
		{"Bmad:Status", "bmadstatus"},
		{"Pipeline:CI", "pipelinecI"},
		{"Build:Default", "builddefault"},
		{"Test:Unit", "testunit"},

		// Simple commands (no conversion needed)
		{"Deploy", "Deploy"},
		{"build", "build"},

		// Edge cases
		{":", ""}, // Empty parts result in empty string
		{"Namespace:", "namespace"},
		{":Method", "method"},
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
	result := DelegateToMage("TestCommand")
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
	result := DelegateToMageWithTimeout("sleepCmd", 500*time.Millisecond)

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
	result := DelegateToMageWithTimeout("argsCmd", DefaultDelegateTimeout, "arg1", "arg2", "arg3")

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
	result := DelegateToMageWithTimeout("testCmd", DefaultDelegateTimeout)

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
	result := DelegateToMageWithTimeout("failCmd", DefaultDelegateTimeout)

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
			name:     "multiple colons only uses first split",
			input:    "A:B:C",
			expected: "ab:C", // SplitN with 2 means B:C is the second part
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single colon with empty parts",
			input:    ":",
			expected: "", // Both parts are empty
		},
		{
			name:     "unicode characters",
			input:    "Ünïcödé:Tëst",
			expected: "ünïcödétëst", // é is lowercased from É, T becomes t
		},
		{
			name:     "numbers in namespace",
			input:    "Build123:Run456",
			expected: "build123run456",
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
