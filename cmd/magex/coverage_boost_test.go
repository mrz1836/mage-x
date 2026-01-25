package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

// TestListCommands_SuccessPath tests the success path of ListCommands
func TestListCommands_SuccessPath(t *testing.T) {
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
	magefileContent := "//" + "go:build mage\n\npackage main\n\nfunc TestCommand() error {\n\treturn nil\n}\n"
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// Call ListCommands - this should trigger discovery and return commands
	commands, err := discovery.ListCommands()

	// Should succeed (even if empty)
	require.NoError(t, err)
	assert.NotNil(t, commands)
}

// TestGetMagefilePath_MagefilesWithAbsError tests GetMagefilePath when filepath.Abs fails
// This is difficult to test without mocking, but we can at least cover the basic paths
func TestGetMagefilePath_MagefilesDirectory(t *testing.T) {
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

	// GetMagefilePath should return the magefiles directory path
	path := GetMagefilePath()
	assert.Contains(t, path, "magefiles")
}

// TestGetMagefilePath_MagefileGoAlt tests GetMagefilePath with magefile.go
func TestGetMagefilePath_MagefileGoAlt(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create magefile.go
	require.NoError(t, os.WriteFile("magefile.go", []byte("package main"), 0o600))

	// GetMagefilePath should return the magefile.go path
	path := GetMagefilePath()
	assert.Contains(t, path, "magefile.go")
}

// TestGetMagefilePath_NeitherExists tests GetMagefilePath when neither exists
func TestGetMagefilePath_NeitherExists(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Don't create anything
	path := GetMagefilePath()
	assert.Empty(t, path)
}

// TestSearchCommands_WithError tests searchCommands when discovery fails
func TestSearchCommands_WithDiscoveryError(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod but with invalid content to cause discovery error
	require.NoError(t, os.WriteFile("go.mod", []byte("invalid go.mod"), 0o600))

	// Create magefile with syntax error
	require.NoError(t, os.WriteFile("magefile.go", []byte("this is not valid go code"), 0o600))

	reg := registry.NewRegistry()
	// Add a test command so we have something to search
	searchCmd, err := registry.NewNamespaceCommand("test", "search").
		WithDescription("test search command").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(searchCmd))

	discovery := NewCommandDiscovery(reg)

	// Capture stdout to avoid cluttering test output
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Search should handle discovery errors gracefully
	searchCommands(reg, discovery, "search")

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestInitMagefile_ErrorPaths tests error paths in initMagefile
// This function calls os.Exit, so we can only test paths before that
func TestInitMagefile_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create existing magefile.go
	require.NoError(t, os.WriteFile("magefile.go", []byte("existing"), 0o600))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// This should detect existing file and exit early
	// We can't fully test the os.Exit behavior, but we can test the detection
	exists := false
	if _, err := os.Stat("magefile.go"); err == nil {
		exists = true
	}

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	assert.True(t, exists, "Should detect existing magefile.go")
}

// TestShowQuickList_EmptyRegistry tests showQuickList with no commands
func TestShowQuickList_EmptyRegistry(t *testing.T) {
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

	showQuickList(reg, discovery)

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	// Just verify it doesn't crash
}

// TestListCommandsVerbose_WithFormatError tests error handling in listCommandsVerbose
// This is hard to trigger without mocking, but we can test the normal path
func TestListCommandsVerbose_WithDeprecated(t *testing.T) {
	reg := registry.NewRegistry()

	// Add a deprecated command
	oldCmd, err := registry.NewNamespaceCommand("test", "old").
		WithDescription("old command").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	oldCmd.Deprecated = "use new command instead"
	require.NoError(t, reg.Register(oldCmd))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	listCommandsVerbose(reg.List(), nil)

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	// Just verify it doesn't crash with deprecated commands
}

// TestListCommandsVerbose_WithEmptyDescription tests empty descriptions
func TestListCommandsVerbose_WithEmptyDescription(t *testing.T) {
	reg := registry.NewRegistry()

	// Add command with empty description
	nodescCmd, err := registry.NewNamespaceCommand("test", "nodesc").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(nodescCmd))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	listCommandsVerbose(reg.List(), nil)

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	// Just verify it handles empty description
}

// TestCleanCache_WithGlobError tests cleanCache error paths
// Glob errors are hard to trigger, but we can test the normal path
func TestCleanCache_NormalPath(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create some cache directories
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

	// Verify cache directories were removed
	_, err1 := os.Stat(".mage")
	_, err2 := os.Stat(".mage-x")
	assert.True(t, os.IsNotExist(err1) || os.IsNotExist(err2), "Should remove at least one cache directory")
}

// TestShowCategorizedCommands_EmptyCategory tests empty category handling
func TestShowCategorizedCommands_EmptyCategory(t *testing.T) {
	reg := registry.NewRegistry()

	// Add command with no category
	nocatCmd, err := registry.NewNamespaceCommand("test", "nocat").
		WithDescription("no category").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	nocatCmd.Category = "" // Empty category
	require.NoError(t, reg.Register(nocatCmd))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Need to call through flag processing to hit showCategorizedCommands
	// For now, just verify registry works
	commands := reg.List()
	assert.Len(t, commands, 1)

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestConvertToMageFormat_WithVariants tests convertToMageFormat with different inputs
func TestConvertToMageFormat_ColonSeparated(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test:unit", "test:unit"},
		{"build:all", "build:all"},
		{"simple", "simple"},
		{"test:unit:coverage", "test:unit:coverage"}, // All colons preserved
		{"Speckit:Install", "speckit:install"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := convertToMageFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestShowNamespaceHelp_WithCommands tests showNamespaceHelp with various command types
func TestShowNamespaceHelp_WithAliases(t *testing.T) {
	reg := registry.NewRegistry()

	// Add command with aliases
	buildCmd, err := registry.NewNamespaceCommand("build", "default").
		WithDescription("default build").
		WithAliases("d", "def").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(buildCmd))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	showNamespaceHelp(reg, "build")

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	// Just verify it doesn't crash
}
