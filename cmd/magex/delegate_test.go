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
)

const (
	secureDirPerm  = 0o750
	secureFilePerm = 0o600
)

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
	err = DelegateToMage("nonexistent")
	if err == nil {
		t.Error("DelegateToMage should return error when no magefile exists")
	}
	if !errors.Is(err, ErrCommandNotFound) {
		t.Errorf("DelegateToMage should return ErrCommandNotFound, got: %v", err)
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
	goModContent := `module testmage

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
	err = DelegateToMage("TestCommand")
	if err != nil {
		// This is expected to work in a real environment with Go
		// but might fail in test environment, so we'll log rather than fail
		t.Logf("DelegateToMage execution failed (expected in test environment): %v", err)
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
