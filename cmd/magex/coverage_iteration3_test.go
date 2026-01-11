package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

// TestInitMagefile_Success tests successful magefile creation
func TestInitMagefile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Call initMagefile - should succeed
	err = initMagefile()
	require.NoError(t, err, "initMagefile should succeed")

	// Verify file was created
	_, err = os.Stat("magefile.go")
	require.NoError(t, err, "magefile.go should exist")

	// Verify content
	content, err := os.ReadFile("magefile.go")
	require.NoError(t, err)
	assert.Contains(t, string(content), "//go:build mage")
	assert.Contains(t, string(content), "github.com/mrz1836/mage-x/pkg/mage/auto")
}

// TestInitMagefile_FileExists tests error when magefile already exists
func TestInitMagefile_FileExists(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create existing magefile
	require.NoError(t, os.WriteFile("magefile.go", []byte("existing"), 0o600))

	// Call initMagefile - should fail
	err = initMagefile()
	require.Error(t, err, "initMagefile should fail when file exists")
	assert.ErrorIs(t, err, ErrMagefileExists)
}

// TestInitMagefile_WriteError tests error during file write
func TestInitMagefile_WriteError(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create magefile.go as a directory to cause it to be detected as existing
	require.NoError(t, os.Mkdir("magefile.go", 0o750))

	// Call initMagefile - should fail because file "exists" (as a directory)
	err = initMagefile()
	require.Error(t, err, "initMagefile should fail when path exists as directory")
	assert.ErrorIs(t, err, ErrMagefileExists)
}

// TestSearchCommands_WithCategoryInfo tests searchCommands with category metadata
func TestSearchCommands_WithCategoryInfo(t *testing.T) {
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
		WithCategory("testing").
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

	// Search for "build" - should show categorized results
	searchCommands(reg, discovery, "build")

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestSearchCommands_EmptyDescription tests searchCommands with commands that have no description
func TestSearchCommands_EmptyDescription(t *testing.T) {
	reg := registry.NewRegistry()

	// Add command with no description
	nodescCmd, err := registry.NewNamespaceCommand("build", "nodesc").
		WithCategory("build").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(nodescCmd))

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

	// Search for "nodesc" - should handle empty description
	searchCommands(reg, discovery, "nodesc")

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestSearchCommands_WithCustomMatchesAlt tests searchCommands showing custom commands
func TestSearchCommands_WithCustomMatchesAlt(t *testing.T) {
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
	magefileContent := `//go:build mage //nolint:goconst // test data

package main

// MySpecialCommand does something special
func MySpecialCommand() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Search for "special" - should find custom command
	searchCommands(reg, discovery, "special")

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestSearchCommands_CustomCommandNoDescription tests custom commands without descriptions
func TestSearchCommands_CustomCommandNoDescription(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod and magefile with custom command (no doc comment)
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))
	magefileContent := `//go:build mage //nolint:goconst // test data

package main

func NoDescCommand() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Search for "nodesc" - should find custom command with default description
	searchCommands(reg, discovery, "nodesc")

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestCleanCache_WithGlobError tests cleanCache when glob pattern is invalid
// This is difficult to trigger, but we can test with edge cases
func TestCleanCache_WithMultipleDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create multiple cache directories with files
	require.NoError(t, os.Mkdir(".mage", 0o750))
	require.NoError(t, os.WriteFile(".mage/test.txt", []byte("test"), 0o600))

	require.NoError(t, os.Mkdir(".mage-x", 0o750))
	require.NoError(t, os.WriteFile(".mage-x/test.txt", []byte("test"), 0o600))

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

// TestListCommandsVerbose_WithFmtError tests listCommandsVerbose error handling
// This is difficult to test without mocking, but we can test various edge cases
func TestListCommandsVerbose_MultipleCustomCommands(t *testing.T) {
	reg := registry.NewRegistry()

	// Add built-in commands
	buildCmd, err := registry.NewNamespaceCommand("build", "default").
		WithDescription("Build").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(buildCmd))

	// Create multiple custom commands with varying descriptions
	customCommands := []DiscoveredCommand{
		{Name: "deploy", OriginalName: "Deploy", Description: "Deploy app"},
		{Name: "migrate", OriginalName: "Migrate", Description: ""},
		{Name: "seed", OriginalName: "Seed", Description: "Seed database"},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	listCommandsVerbose(reg.List(), customCommands)

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestShowCategorizedCommands_EmptyCategoryInfoAlt tests when category metadata is missing
func TestShowCategorizedCommands_EmptyCategoryInfoAlt(t *testing.T) {
	reg := registry.NewRegistry()

	// Add command with category that has no metadata
	cmd, err := registry.NewNamespaceCommand("custom", "test").
		WithCategory("uncategorized").
		WithDescription("Test command").
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

	discovery := NewCommandDiscovery(reg)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// This exercises searchCommands which calls category info logic
	searchCommands(reg, discovery, "test")

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestCompileForMage_WritesSuccessfully tests compileForMage success path
func TestCompileForMage_WritesSuccessfully(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	outputFile := "compiled-magefile.go"

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	compileForMage(outputFile)

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	// Verify file was created
	_, err = os.Stat(outputFile)
	require.NoError(t, err, "Output file should exist")

	// Verify content
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "//go:build mage")
	assert.Contains(t, string(content), "auto-generated")
	assert.Contains(t, string(content), "github.com/mrz1836/mage-x/pkg/mage")
}

// TestGetMagefilePath_PreferMagefiles tests that magefiles directory is preferred
func TestGetMagefilePath_PreferMagefilesOverFile(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create both magefiles directory and magefile.go
	require.NoError(t, os.Mkdir("magefiles", 0o750))
	require.NoError(t, os.WriteFile("magefile.go", []byte("package main"), 0o600))

	// GetMagefilePath should prefer magefiles directory
	path := GetMagefilePath()
	assert.Contains(t, path, "magefiles")
	assert.NotContains(t, path, "magefile.go")
}

// TestDelegateToMageWithTimeout_WithMagefileError tests error when no magefile exists
func TestDelegateToMageWithTimeout_NoMagefile(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Don't create any magefile - should return error
	result := DelegateToMageWithTimeout("test", DefaultDelegateTimeout)

	assert.Equal(t, 1, result.ExitCode)
	require.Error(t, result.Err)
	assert.ErrorIs(t, result.Err, ErrCommandNotFound)
}

// TestShowQuickList_WithCustomCommands tests showQuickList with custom commands
func TestShowQuickList_WithCustomCommands(t *testing.T) {
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
	magefileContent := `//go:build mage //nolint:goconst // test data

package main

func CustomDeploy() error {
	return nil
}

func CustomMigrate() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()

	// Add some built-in commands
	buildCmd, err := registry.NewNamespaceCommand("build", "default").
		WithDescription("Build").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(buildCmd))

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
}

// TestListCommands_SimpleMode tests listCommands in non-verbose mode
func TestListCommands_SimpleMode(t *testing.T) {
	reg := registry.NewRegistry()

	// Add commands
	buildCmd, err := registry.NewNamespaceCommand("build", "default").
		WithDescription("Build the project").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(buildCmd))

	testCmd, err := registry.NewNamespaceCommand("test", "unit").
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

	// Call in simple mode (verbose=false)
	listCommands(reg, discovery, false)

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}
