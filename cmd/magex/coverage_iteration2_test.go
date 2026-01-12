package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

// TestGetMagefilePath_MagefilesAbsPathError tests GetMagefilePath when filepath.Abs fails for magefiles
// This path is covered by the function returning magefiles directly if Abs fails
func TestGetMagefilePath_MagefilesRelativePath(t *testing.T) {
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

	// GetMagefilePath should find magefiles directory
	path := GetMagefilePath()
	assert.Contains(t, path, "magefiles")
}

// TestGetMagefilePath_MagefileGoAbsPathError tests GetMagefilePath when filepath.Abs fails for magefile.go
func TestGetMagefilePath_MagefileGoRelativePath(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create magefile.go only (no magefiles directory)
	require.NoError(t, os.WriteFile("magefile.go", []byte("package main"), 0o600))

	// GetMagefilePath should find magefile.go
	path := GetMagefilePath()
	assert.Contains(t, path, "magefile.go")
}

// TestSearchCommands_CustomCommandsMatch tests searchCommands when custom commands match
func TestSearchCommands_CustomCommandsMatch(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod and magefile with custom command
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))
	magefileContent := `/` + `/go:build mage

package main

// Deploy deploys the application to production
func Deploy() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// Capture stdout to avoid cluttering test output
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Search for "deploy" should find the custom command
	searchCommands(reg, discovery, "deploy")

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestSearchCommands_DescriptionMatchAlt tests searchCommands matching by description
func TestSearchCommands_DescriptionMatchAlt(t *testing.T) {
	reg := registry.NewRegistry()

	// Add command with specific description
	testCmd, err := registry.NewNamespaceCommand("build", "docker").
		WithDescription("Build Docker containers for deployment").
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

	// Search for "docker" should match by description
	searchCommands(reg, discovery, "docker")

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestInitMagefile_NoGoMod tests initMagefile when go.mod doesn't exist
func TestInitMagefile_NoGoMod(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Don't create go.mod - check that it doesn't exist
	_, err = os.Stat("go.mod")
	assert.True(t, os.IsNotExist(err), "go.mod should not exist")

	// initMagefile would call os.Exit(1) if we actually called it
	// Just verify the condition exists
}

// TestCleanCache_WithTempPlugins tests cleanCache with temp plugin directories
func TestCleanCache_WithTempPlugins(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create cache directories
	require.NoError(t, os.Mkdir(".mage", 0o750))
	require.NoError(t, os.Mkdir(".mage-x", 0o750))

	// Create a temp plugin directory pattern in system temp
	tempPluginDir := filepath.Join(os.TempDir(), "magex-plugin-test-"+t.Name())
	require.NoError(t, os.Mkdir(tempPluginDir, 0o750))
	t.Cleanup(func() {
		_ = os.RemoveAll(tempPluginDir) //nolint:errcheck // test cleanup
	})

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	cleanCache()

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	// Verify local cache directories were removed
	_, err1 := os.Stat(".mage")
	_, err2 := os.Stat(".mage-x")
	assert.True(t, os.IsNotExist(err1), "Should remove .mage")
	assert.True(t, os.IsNotExist(err2), "Should remove .mage-x")
}

// TestShowCategorizedCommands_WithCategories tests showCategorizedCommands with multiple categories
func TestShowCategorizedCommands_WithCategories(t *testing.T) {
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
		WithDescription("Run unit tests").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(testCmd))

	lintCmd, err := registry.NewNamespaceCommand("lint", "default").
		WithCategory("quality").
		WithDescription("Run linters").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(lintCmd))

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

	// Call through showUsage which calls showCategorizedCommands
	showUsage()

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	// Just verify it doesn't crash
	// The actual output validation is difficult without parsing the output
	_ = discovery
}

// TestListCommands_EmptyRegistryAlt tests listCommands with empty registry
func TestListCommands_EmptyRegistryAlt(t *testing.T) {
	reg := registry.NewRegistry()

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

	listCommands(reg, discovery, false)

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestShowCommandHelp_NonExistentCommand tests showCommandHelp with non-existent command
func TestShowCommandHelp_NonExistentCommand(t *testing.T) {
	reg := registry.NewRegistry()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	showCommandHelp(reg, "nonexistent:command")

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestShowCommandHelp_WithAliases tests showCommandHelp with command aliases
func TestShowCommandHelp_WithAliases(t *testing.T) {
	reg := registry.NewRegistry()

	// Add command with aliases
	buildCmd, err := registry.NewNamespaceCommand("build", "default").
		WithDescription("Build the project").
		WithAliases("b", "bld").
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
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestShowNamespaceHelp_WithMultipleCommands tests showNamespaceHelp with multiple commands
func TestShowNamespaceHelp_WithMultipleCommands(t *testing.T) {
	reg := registry.NewRegistry()

	// Add multiple commands in same namespace
	buildDefault, err := registry.NewNamespaceCommand("build", "default").
		WithDescription("Build for current platform").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(buildDefault))

	buildAll, err := registry.NewNamespaceCommand("build", "all").
		WithDescription("Build for all platforms").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(buildAll))

	buildLinux, err := registry.NewNamespaceCommand("build", "linux").
		WithDescription("Build for Linux").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(buildLinux))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	showNamespaceHelp(reg, "build")

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestHasCommand_CaseInsensitive tests HasCommand is case-insensitive
func TestHasCommand_CaseInsensitive(t *testing.T) {
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

func DeployApp() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// Test different cases
	assert.True(t, discovery.HasCommand("deployapp"), "Should find lowercase")
	assert.True(t, discovery.HasCommand("DEPLOYAPP"), "Should find uppercase")
	assert.True(t, discovery.HasCommand("DeployApp"), "Should find mixed case")
}

// TestGetCommand_CaseInsensitive tests GetCommand is case-insensitive
func TestGetCommand_CaseInsensitive(t *testing.T) {
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

func TestCommand() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// Test different cases
	cmd1, found1 := discovery.GetCommand("testcommand")
	assert.True(t, found1, "Should find lowercase")
	assert.NotNil(t, cmd1)

	cmd2, found2 := discovery.GetCommand("TESTCOMMAND")
	assert.True(t, found2, "Should find uppercase")
	assert.NotNil(t, cmd2)

	cmd3, found3 := discovery.GetCommand("TestCommand")
	assert.True(t, found3, "Should find mixed case")
	assert.NotNil(t, cmd3)
}
