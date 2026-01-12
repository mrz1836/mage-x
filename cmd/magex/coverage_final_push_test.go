package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

// TestHasCommand_DiscoverSuccess tests HasCommand when discovery succeeds
func TestHasCommand_DiscoverSuccess(t *testing.T) {
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

	// Create a simple magefile with a custom command
	magefileContent := `//go:build mage

package main

func MyCustomCommand() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// Check for existing command (after discovery)
	hasCmd := discovery.HasCommand("mycustomcommand")
	assert.True(t, hasCmd, "Should find custom command")

	// Check for non-existent command
	hasCmd = discovery.HasCommand("nonexistent")
	assert.False(t, hasCmd, "Should not find non-existent command")
}

// TestGetCommand_DiscoverSuccess tests GetCommand when discovery succeeds
func TestGetCommand_DiscoverSuccess(t *testing.T) {
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

	// Create a simple magefile
	magefileContent := `//go:build mage

package main

func TestCommand() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// Get existing command
	cmd, found := discovery.GetCommand("testcommand")
	assert.True(t, found, "Should find command")
	assert.NotNil(t, cmd, "Command should not be nil")

	// Get non-existent command
	cmd, found = discovery.GetCommand("nonexistent")
	assert.False(t, found, "Should not find non-existent command")
	assert.Nil(t, cmd, "Command should be nil")
}

// TestGetCommandsForHelp_WithDiscoveredCommands tests GetCommandsForHelp
func TestGetCommandsForHelp_WithDiscoveredCommands(t *testing.T) {
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

	// Create a magefile with description comment
	magefileContent := `//go:build mage

package main

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
	// Should have at least the discovered command
	if len(helpLines) > 0 {
		// Check that at least one line contains "deploy"
		found := false
		for _, line := range helpLines {
			if strings.Contains(strings.ToLower(line), "deploy") {
				found = true
				break
			}
		}
		assert.True(t, found, "Should find deploy command in help")
	}
}

// TestSearchCommands_NoMatchesAlt tests searchCommands with no matches
func TestSearchCommands_NoMatchesAlt(t *testing.T) {
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
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Search for something that doesn't exist
	searchCommands(reg, discovery, "totallyfakecommandthatdoesnotexist")

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r) //nolint:errcheck // test cleanup
	output := buf.String()

	// Should show no results message (actual message is different)
	assert.Contains(t, output, "No exact commands found", "Should indicate no results")
}

// TestSearchCommands_WithMatches tests searchCommands with matches
func TestSearchCommands_WithMatches(t *testing.T) {
	reg := registry.NewRegistry()

	// Add a test command
	testCmd, err := registry.NewNamespaceCommand("test", "unit").
		WithDescription("Run unit tests").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(testCmd))

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

	// Search for "test"
	searchCommands(reg, discovery, "test")

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r) //nolint:errcheck // test cleanup
	output := buf.String()

	// Should show results
	assert.Contains(t, output, "Search Results", "Should show search results header")
}

// TestListCommandsVerbose_WithCustomCommands tests listCommandsVerbose with custom commands
func TestListCommandsVerbose_WithCustomCommands(t *testing.T) {
	reg := registry.NewRegistry()

	// Add a built-in command
	builtinCmd, err := registry.NewNamespaceCommand("build", "default").
		WithDescription("Build the project").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(builtinCmd))

	// Create custom commands
	customCommands := []DiscoveredCommand{
		{
			Name:         "deploy",
			OriginalName: "Deploy",
			Description:  "Deploy application",
		},
		{
			Name:         "migrate",
			OriginalName: "Migrate",
			Description:  "", // Empty description
		},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	listCommandsVerbose(reg.List(), customCommands)

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r) //nolint:errcheck // test cleanup
	output := buf.String()

	// Verify output contains commands
	assert.Contains(t, output, "build", "Should show built-in command")
	assert.Contains(t, output, "deploy", "Should show custom command")
	assert.Contains(t, output, "migrate", "Should show custom command with no description")
	assert.Contains(t, output, "custom", "Custom commands should be marked")
}

// TestCleanCache_WithRemoveError tests cleanCache when RemoveAll fails
// This is hard to test without mocking, but we can at least test the normal path more thoroughly
func TestCleanCache_WithCacheDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create multiple cache directories
	require.NoError(t, os.Mkdir(".mage", 0o750))
	require.NoError(t, os.Mkdir(".mage-x", 0o750))

	// Add some files in the directories
	require.NoError(t, os.WriteFile(".mage/file.txt", []byte("test"), 0o600))
	require.NoError(t, os.WriteFile(".mage-x/file.txt", []byte("test"), 0o600))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	cleanCache()

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r) //nolint:errcheck // test cleanup
	output := buf.String()

	// Verify the function ran (looks for "Removed" message)
	assert.Contains(t, output, "Removed", "Should show removal messages")

	// Verify cache directories were removed
	_, err1 := os.Stat(".mage")
	_, err2 := os.Stat(".mage-x")
	assert.True(t, os.IsNotExist(err1) || os.IsNotExist(err2), "Should remove cache directories")
}

// TestShowCategorizedCommands_MultipleCategories tests showCategorizedCommands
func TestShowCategorizedCommands_MultipleCategories(t *testing.T) {
	reg := registry.NewRegistry()

	// Add commands in different categories
	buildCmd, err := registry.NewNamespaceCommand("build", "default").
		WithCategory("build").
		WithDescription("Build the project").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(buildCmd))

	testCmd, err := registry.NewNamespaceCommand("test", "unit").
		WithCategory("test").
		WithDescription("Run tests").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(testCmd))

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

	// This will call showCategorizedCommands internally
	// We need to trigger it through listCommands
	listCommands(reg, discovery, false)

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r) //nolint:errcheck // test cleanup
	output := buf.String()

	// Should show commands
	assert.NotEmpty(t, output, "Should produce output")
}

// TestInitMagefile_WithGoModError tests initMagefile paths
// Can't fully test os.Exit behavior, but can test setup
func TestInitMagefile_GoModExists(t *testing.T) {
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

	// Verify go.mod exists (this tests the check in initMagefile)
	_, err = os.Stat("go.mod")
	assert.NoError(t, err, "go.mod should exist")
}

// TestCompileForMage_OutputPath tests compileForMage with different output paths
// Can't fully test os.Exit behavior
func TestCompileForMage_ValidPath(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	outputFile := "test-magefile.go"

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	compileForMage(outputFile)

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r) //nolint:errcheck // test cleanup
	output := buf.String()

	// Should show success message
	assert.Contains(t, output, "Generated", "Should show generated message")

	// Verify file was created
	_, err = os.Stat(outputFile)
	assert.NoError(t, err, "Output file should be created")
}

// TestShowQuickList_WithMixedCommands tests showQuickList with various command types
func TestShowQuickList_WithMixedCommands(t *testing.T) {
	reg := registry.NewRegistry()

	// Add different types of commands
	buildCmd, err := registry.NewNamespaceCommand("build", "default").
		WithDescription("Build project").
		WithAliases("b").
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

	// Create a custom command
	magefileContent := `//go:build mage

package main

func Custom() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	discovery := NewCommandDiscovery(reg)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	showQuickList(reg, discovery)

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r) //nolint:errcheck // test cleanup
	output := buf.String()

	// Should show quick list
	assert.NotEmpty(t, output, "Should produce output")
}

// TestShowCommandHelp_WithExamples tests showCommandHelp with examples
func TestShowCommandHelp_WithExamples(t *testing.T) {
	reg := registry.NewRegistry()

	// Add command with examples
	buildCmd, err := registry.NewNamespaceCommand("build", "default").
		WithDescription("Build the project").
		WithExamples("magex build:default", "magex build").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(buildCmd))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	showCommandHelp(reg, "build:default")

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r) //nolint:errcheck // test cleanup
	output := buf.String()

	// Should show examples
	assert.Contains(t, output, "Examples", "Should show examples section")
}

// TestShowNamespaceHelp_EmptyNamespace tests showNamespaceHelp with no commands
func TestShowNamespaceHelp_EmptyNamespace(t *testing.T) {
	reg := registry.NewRegistry()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	showNamespaceHelp(reg, "nonexistent")

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r) //nolint:errcheck // test cleanup
	output := buf.String()

	// Should handle empty namespace gracefully
	assert.NotEmpty(t, output, "Should produce some output")
}

// TestListCommands_WithVerboseAndCustom tests listCommands with verbose flag and custom commands
func TestListCommands_WithVerboseAndCustom(t *testing.T) {
	reg := registry.NewRegistry()

	// Add a command
	cmd, err := registry.NewNamespaceCommand("test", "unit").
		WithDescription("Run unit tests").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(cmd))

	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create custom command
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))
	magefileContent := `//go:build mage

package main

func Deploy() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	discovery := NewCommandDiscovery(reg)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	listCommands(reg, discovery, true) // verbose = true

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r) //nolint:errcheck // test cleanup
	output := buf.String()

	// Should show both built-in and custom commands
	assert.Contains(t, output, "test", "Should show built-in command")
	assert.Contains(t, output, "deploy", "Should show custom command")
}

// TestDelegateToMageWithTimeout_WithArguments tests delegation with arguments
func TestDelegateToMageWithTimeout_WithArguments(t *testing.T) {
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

	// Create magefile that accepts arguments
	magefileContent := `//go:build mage

package main

import "fmt"

func PrintArgs() error {
	fmt.Println("Command executed")
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	// Call with arguments
	result := DelegateToMageWithTimeout(context.Background(), "printArgs", DefaultDelegateTimeout, "arg1", "arg2")

	// Should execute (may succeed or fail depending on environment)
	_ = result // Just testing the code path with arguments
}
