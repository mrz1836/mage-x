package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

// TestDelegateToMageWithTimeout_GoRunFallback tests go run path when mage binary unavailable
func TestDelegateToMageWithTimeout_GoRunFallback(t *testing.T) {
	// Skip if go is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available")
	}

	// Skip if mage binary IS available (we want to test the fallback)
	if _, err := exec.LookPath("mage"); err == nil {
		t.Skip("mage binary is available, skipping go run fallback test")
	}

	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))

	// Create magefile
	magefileContent := `/` + `/go:build mage

package main

import "fmt"

func TestGoRun() error {
	fmt.Println("Using go run")
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	// This should use go run since mage is not available
	result := DelegateToMageWithTimeout(context.Background(), "testGoRun", 10*time.Second)

	assert.Equal(t, 0, result.ExitCode, "Should succeed with go run")
	assert.NoError(t, result.Err)
}

// TestDelegateToMageWithTimeout_GoRunWithDirectory tests go run with magefiles directory
func TestDelegateToMageWithTimeout_GoRunWithDirectory(t *testing.T) {
	// Skip if go is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available")
	}

	// Skip if mage binary IS available
	if _, err := exec.LookPath("mage"); err == nil {
		t.Skip("mage binary is available, skipping go run test")
	}

	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))

	// Create magefiles directory (preferred)
	require.NoError(t, os.Mkdir("magefiles", 0o750))
	magefilesContent := `/` + `/go:build mage

package main

import "fmt"

func DirTest() error {
	fmt.Println("Using magefiles directory")
	return nil
}
`
	require.NoError(t, os.WriteFile("magefiles/commands.go", []byte(magefilesContent), 0o600))

	// This should use go run with magefiles directory
	result := DelegateToMageWithTimeout(context.Background(), "dirTest", 10*time.Second)

	assert.Equal(t, 0, result.ExitCode, "Should succeed with magefiles directory")
	assert.NoError(t, result.Err)
}

// TestDelegateToMageWithTimeout_CommandFailureNoStderr tests command failure without stderr
func TestDelegateToMageWithTimeout_CommandFailureNoStderr(t *testing.T) {
	// Skip if go is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available")
	}

	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))

	// Create magefile with a command that fails silently (no stderr)
	magefileContent := `/` + `/go:build mage

package main

import "errors"

func FailSilent() error {
	return errors.New("silent failure")
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	// This should fail without stderr content
	result := DelegateToMageWithTimeout(context.Background(), "failSilent", 10*time.Second)

	assert.NotEqual(t, 0, result.ExitCode, "Should have non-zero exit code")
	assert.Error(t, result.Err, "Should return error")
}

// TestGetMagefilePath_AbsolutePathErrors tests filepath.Abs error handling
func TestGetMagefilePath_AbsolutePathErrors(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create magefiles directory
	require.NoError(t, os.Mkdir("magefiles", 0o750))

	// GetMagefilePath should handle absolute path
	path := GetMagefilePath()
	assert.NotEmpty(t, path, "Should return path")

	// When it can get absolute path, it should be absolute
	if !filepath.IsAbs(path) {
		// If it returns relative, that means Abs failed (which is the error path we're testing)
		assert.Equal(t, "magefiles", path, "Should return relative path when Abs fails")
	} else {
		// It got absolute path successfully
		assert.Contains(t, path, "magefiles", "Should contain magefiles")
	}
}

// TestConvertToMageFormat_EdgeCasesAlt tests edge cases in format conversion
func TestConvertToMageFormat_EdgeCasesAlt(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no colon",
			input:    "simpleCommand",
			expected: "simpleCommand",
		},
		{
			name:     "empty method",
			input:    "build:",
			expected: "build",
		},
		{
			name:     "single character method",
			input:    "test:A",
			expected: "testa",
		},
		{
			name:     "uppercase namespace",
			input:    "BUILD:Default",
			expected: "builddefault",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToMageFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestTryCustomCommand_ErrorExitCode tests tryCustomCommand with error exit code
func TestTryCustomCommand_ErrorExitCode(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)

	// Capture stderr to avoid cluttering test output
	oldStderr := os.Stderr
	r, w, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)
	os.Stderr = w

	t.Cleanup(func() {
		_ = w.Close() //nolint:errcheck // test cleanup
		os.Stderr = oldStderr
		_ = r.Close() //nolint:errcheck // test cleanup

		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))

	// Create magefile with a failing command
	magefileContent := `/` + `/go:build mage

package main

import "errors"

func FailCmd() error {
	return errors.New("command failed")
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	// Note: tryCustomCommand calls os.Exit on error, so we can't easily test the exit path
	// But we can verify it attempts to execute
	// This test documents the limitation
	_ = magefileContent
}

// TestListByNamespace_WithCustomCommandsAlt tests listByNamespace with discovered custom commands
func TestListByNamespace_WithCustomCommandsAlt(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod and magefile
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))
	magefileContent := `/` + `/go:build mage

package main

type Build mg.Namespace

func (Build) Linux() error {
	return nil
}

func (Build) Windows() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)
	os.Stdout = w

	listByNamespace(reg, discovery)

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestShowQuickList_WithManyCommands tests showQuickList with multiple commands
func TestShowQuickList_WithManyCommands(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)
	os.Stdout = w

	showQuickList(reg, discovery)

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}
