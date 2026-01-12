package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

// TestDelegateToMageWithTimeout_ConflictRenameErrorAlt tests error during conflict handling
func TestDelegateToMageWithTimeout_ConflictRenameErrorAlt(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		// Restore permissions before cleanup
		_ = os.Chmod(tmpDir, 0o750) //nolint:errcheck,gosec // test cleanup
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))

	// Create magefiles directory
	require.NoError(t, os.Mkdir("magefiles", 0o750))
	magefilesContent := `/` + `/go:build mage

package main

func TestCmd() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefiles/commands.go", []byte(magefilesContent), 0o600))

	// Create root magefile.go (conflict)
	require.NoError(t, os.WriteFile("magefile.go", []byte("package main\n"), 0o600))

	// Make directory read-only to prevent rename
	require.NoError(t, os.Chmod(tmpDir, 0o500)) //nolint:gosec // intentional for test

	// This should fail when trying to rename magefile.go
	result := DelegateToMageWithTimeout(context.Background(), "testCmd", 5*time.Second)

	// Restore permissions
	_ = os.Chmod(tmpDir, 0o750) //nolint:errcheck,gosec // test cleanup

	assert.NotEqual(t, 0, result.ExitCode, "Should have non-zero exit code")
	assert.Error(t, result.Err, "Should return error when rename fails")
}

// TestListCommandsVerbose_WithDeprecatedCommands tests verbose listing with deprecated commands
func TestListCommandsVerbose_WithDeprecatedCommands(t *testing.T) {
	reg := registry.NewRegistry()

	// Add a deprecated command
	deprecatedCmd, err := registry.NewNamespaceCommand("old", "command").
		WithDescription("Old command (deprecated)").
		Deprecated("Use new:command instead").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(deprecatedCmd))

	// Add a normal command
	normalCmd, err := registry.NewNamespaceCommand("new", "command").
		WithDescription("New command").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(normalCmd))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Test with nil discovery (custom commands section)
	listCommandsVerbose(reg.List(), nil)

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestCleanCache_WithReadOnlyFile tests cleanCache with file that can't be removed
func TestCleanCache_WithReadOnlyFile(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		// Restore permissions before cleanup
		_ = os.Chmod(tmpDir, 0o750) //nolint:errcheck,gosec // test cleanup
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create .mage directory with a file inside
	require.NoError(t, os.Mkdir(".mage", 0o750))
	require.NoError(t, os.WriteFile(".mage/testfile.txt", []byte("test"), 0o600))

	// Make directory read-only to prevent file deletion
	require.NoError(t, os.Chmod(".mage", 0o500)) //nolint:gosec // intentional for test

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	cleanCache()

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	// Restore permissions for cleanup
	_ = os.Chmod(".mage", 0o750) //nolint:errcheck,gosec // test cleanup
}

// TestHasCommand_WithError tests HasCommand when discovery fails
func TestHasCommand_WithError(t *testing.T) {
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

	// Create invalid magefile that will cause parsing error
	invalidMagefile := `/` + `/go:build mage

package main

func InvalidSyntax( {  // Missing closing paren
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(invalidMagefile), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// This should handle the error gracefully
	result := discovery.HasCommand("somecommand")
	assert.False(t, result, "Should return false when discovery fails")
}

// TestGetCommand_WithError tests GetCommand when discovery fails
func TestGetCommand_WithError(t *testing.T) {
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

	// Create invalid magefile
	invalidMagefile := `/` + `/go:build mage

package main

func InvalidSyntax( {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(invalidMagefile), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// This should handle the error gracefully
	cmd, found := discovery.GetCommand("somecommand")
	assert.False(t, found, "Should return false when discovery fails")
	assert.Nil(t, cmd, "Should return nil when discovery fails")
}

// TestGetCommandsForHelp_WithNamespaces tests GetCommandsForHelp with namespace filtering
func TestGetCommandsForHelp_WithNamespaces(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod and magefile with namespace commands
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))
	magefileContent := `/` + `/go:build mage

package main

type Build mg.Namespace

// Default builds for current platform
func (Build) Default() error {
	return nil
}

// All builds for all platforms
func (Build) All() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// Trigger discovery
	_ = discovery.HasCommand("build:default")

	// Get help lines
	helpLines := discovery.GetCommandsForHelp()
	assert.NotEmpty(t, helpLines, "Should have help lines")
}

// TestShowCategorizedCommands_EmptyCategories tests showCategorizedCommands with no categories
func TestShowCategorizedCommands_EmptyCategories(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// No commands, no categories
	reg := registry.NewRegistry()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	showCategorizedCommands(reg)

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestInitMagefile_WritePermissionDenied tests initMagefile when write is denied
func TestInitMagefile_WritePermissionDenied(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping permission test when running as root")
	}

	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		// Restore permissions before cleanup
		_ = os.Chmod(tmpDir, 0o750) //nolint:errcheck,gosec // test cleanup
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))

	// Make directory read-only
	require.NoError(t, os.Chmod(tmpDir, 0o500)) //nolint:gosec // intentional for test

	// This should fail
	err = initMagefile()
	require.Error(t, err, "Should fail when directory is read-only")

	// Restore permissions
	_ = os.Chmod(tmpDir, 0o750) //nolint:errcheck,gosec // test cleanup
}
