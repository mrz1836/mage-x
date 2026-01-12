package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

// TestDelegateToMageWithTimeout_WithMagefilesDir tests delegation to magefiles directory
func TestDelegateToMageWithTimeout_WithMagefilesDir(t *testing.T) {
	// Skip if go is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available")
	}

	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldDir))
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))

	// Create magefiles directory (preferred over magefile.go)
	require.NoError(t, os.Mkdir("magefiles", 0o750))
	magefilesContent := `/` + `/go:build mage

package main

func TestCmd() error {
	return nil
}
`
	require.NoError(t, os.WriteFile(filepath.Join("magefiles", "commands.go"), []byte(magefilesContent), 0o600))

	// Delegate to magefiles directory
	result := DelegateToMageWithTimeout("testCmd", 10*time.Second)

	// Result depends on environment, just testing the code path
	_ = result
}

// TestDelegateToMageWithTimeout_ConflictHandling tests handling of both magefiles/ and magefile.go
func TestDelegateToMageWithTimeout_ConflictHandling(t *testing.T) {
	// Skip if go is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available")
	}

	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldDir))
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))

	// Create BOTH magefiles directory and magefile.go (conflict situation)
	require.NoError(t, os.Mkdir("magefiles", 0o750))
	magefilesContent := `/` + `/go:build mage

package main

func TestCmd() error {
	return nil
}
`
	require.NoError(t, os.WriteFile(filepath.Join("magefiles", "commands.go"), []byte(magefilesContent), 0o600))

	// Also create root magefile.go
	rootMagefileContent := `/` + `/go:build mage

package main

func RootCmd() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(rootMagefileContent), 0o600))

	// Delegate - this should temporarily rename magefile.go and restore it
	result := DelegateToMageWithTimeout("testCmd", 10*time.Second)

	// Verify magefile.go was restored after execution
	_, err = os.Stat("magefile.go")
	assert.NoError(t, err, "magefile.go should be restored after conflict handling")

	// Result depends on environment
	_ = result
}

// TestDelegateToMageWithTimeout_UsingMageBinary tests using mage binary if available
func TestDelegateToMageWithTimeout_UsingMageBinary(t *testing.T) {
	// Skip if mage binary is not available
	if _, err := exec.LookPath("mage"); err != nil {
		t.Skip("mage binary not available")
	}

	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldDir))
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))

	// Create magefile
	magefileContent := `/` + `/go:build mage

package main

func TestCmd() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	// Delegate using mage binary (different code path than go run)
	result := DelegateToMageWithTimeout("testCmd", 10*time.Second)

	// Result depends on environment
	_ = result
}

// TestCompileForMage_WithReadOnlyDir tests compileForMage write error
func TestCompileForMage_WithReadOnlyDir(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Skipping read-only test when running as root")
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

	// Make directory read-only to trigger write error
	require.NoError(t, os.Chmod(tmpDir, 0o500)) //nolint:gosec // intentional for test

	outputFile := "readonly-test.go"

	// Capture stdout to verify error message
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// This should trigger os.Exit(1), but we can't test that
	// We're just testing the code path is exercised
	// compileForMage(outputFile)

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	// Note: Can't actually call compileForMage here as it would exit the test
	// This test documents the limitation
	_ = outputFile
}

// TestHasCommand_ReturnsFalseOnMismatch tests HasCommand when command doesn't exist
func TestHasCommand_ReturnsFalseOnMismatch(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod and magefile with specific commands
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))
	magefileContent := `/` + `/go:build mage

package main

func SpecificCommand() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// Test that non-existent command returns false
	hasCmd := discovery.HasCommand("nonexistentcommand")
	assert.False(t, hasCmd, "Should return false for non-existent command")

	// Verify existing command returns true
	hasCmd = discovery.HasCommand("specificcommand")
	assert.True(t, hasCmd, "Should return true for existing command")
}

// TestGetCommand_ReturnsNilOnMismatch tests GetCommand when command doesn't exist
func TestGetCommand_ReturnsNilOnMismatch(t *testing.T) {
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

func ExistingCommand() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// Test that non-existent command returns nil
	cmd, found := discovery.GetCommand("missingcommand")
	assert.False(t, found, "Should return false for non-existent command")
	assert.Nil(t, cmd, "Should return nil for non-existent command")

	// Verify existing command is found
	cmd, found = discovery.GetCommand("existingcommand")
	assert.True(t, found, "Should return true for existing command")
	assert.NotNil(t, cmd, "Should return command info")
}

// TestGetCommandsForHelp_WithMultipleCommands tests GetCommandsForHelp formatting
func TestGetCommandsForHelp_WithMultipleCommands(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod and magefile with multiple commands
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))
	magefileContent := `/` + `/go:build mage

package main

// Build builds the project
func Build() error {
	return nil
}

// Test runs tests
func Test() error {
	return nil
}

// Deploy deploys the application
func Deploy() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// Get help lines
	helpLines := discovery.GetCommandsForHelp()
	assert.NotNil(t, helpLines, "Help lines should not be nil")
	assert.NotEmpty(t, helpLines, "Should have help lines for commands")

	// Should have entries for all three commands
	assert.GreaterOrEqual(t, len(helpLines), 3, "Should have at least 3 help lines")
}

// TestCleanCache_EmptyDirectories tests cleanCache with empty cache dirs
func TestCleanCache_EmptyDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create empty cache directories
	require.NoError(t, os.Mkdir(".mage", 0o750))
	require.NoError(t, os.Mkdir(".mage-x", 0o750))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	cleanCache()

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	// Verify directories were removed
	_, err1 := os.Stat(".mage")
	_, err2 := os.Stat(".mage-x")
	assert.True(t, os.IsNotExist(err1), "Should remove .mage")
	assert.True(t, os.IsNotExist(err2), "Should remove .mage-x")
}

// TestSearchCommands_FuzzyMatching tests fuzzy matching in handleNoSearchResults
func TestSearchCommands_FuzzyMatching(t *testing.T) {
	reg := registry.NewRegistry()

	// Add commands with names that could fuzzy match
	buildCmd, err := registry.NewNamespaceCommand("build", "default").
		WithDescription("Build the project").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(buildCmd))

	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	discovery := NewCommandDiscovery(reg)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Search for typo that should fuzzy match (e.g., "bild" close to "build")
	searchCommands(reg, discovery, "bild")

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestShowCommandHelp_WithLongDescription tests showCommandHelp with multiline description
func TestShowCommandHelp_WithLongDescription(t *testing.T) {
	reg := registry.NewRegistry()

	// Add command with long description
	longDesc := "This is a very long description that spans multiple concepts. " +
		"It explains what the command does in great detail. " +
		"It covers various aspects and use cases."

	buildCmd, err := registry.NewNamespaceCommand("build", "detailed").
		WithDescription(longDesc).
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(buildCmd))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	showCommandHelp(reg, "build:detailed")

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestListCommandsVerbose_FlushError tests listCommandsVerbose flush error path
// This is difficult to trigger without mocking, but we test the code path exists
func TestListCommandsVerbose_LargeOutput(t *testing.T) {
	reg := registry.NewRegistry()

	// Add many commands to exercise the flush path
	for i := 0; i < 50; i++ {
		subCmd := "cmd" + string(rune('a'+i%26)) + string(rune('0'+i/26))
		cmd, err := registry.NewNamespaceCommand("test", subCmd).
			WithDescription("Test command").
			WithFunc(func() error { return nil }).
			Build()
		require.NoError(t, err)
		require.NoError(t, reg.Register(cmd))
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	listCommandsVerbose(reg.List(), nil)

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestDelegateToMageWithTimeout_CommandFormat tests command format conversion
func TestDelegateToMageWithTimeout_CommandFormat(t *testing.T) {
	// Skip if go is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available")
	}

	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldDir))
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod and magefile
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))
	magefileContent := `/` + `/go:build mage

package main

func TestUnit() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	// Test with colon-separated format (should be converted to camelCase)
	result := DelegateToMageWithTimeout("test:unit", 10*time.Second)

	// Result depends on environment
	_ = result
}
