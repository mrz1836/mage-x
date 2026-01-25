package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

// TestDiscover_WithMultipleCommands tests Discover with multiple commands.
func TestDiscover_WithMultipleCommands(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldDir))
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))

	// Create magefile with multiple commands
	magefileContent := `//go:build mage

package main

// TestCmd is a test command
func TestCmd() error {
	return nil
}

// Deploy deploys the app
func Deploy() error {
	return nil
}

// Build builds the project
func Build() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	err = discovery.Discover()
	require.NoError(t, err)

	// Should have discovered commands
	assert.NotEmpty(t, discovery.commands)
}

// TestGetCommandsForHelp_EmptyCommands tests GetCommandsForHelp with no commands.
func TestGetCommandsForHelp_EmptyCommands(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldDir))
	})

	require.NoError(t, os.Chdir(tmpDir))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// No magefile exists
	helpLines := discovery.GetCommandsForHelp()

	// Should return nil or empty
	assert.Empty(t, helpLines)
}

// TestGetCommandsForHelp_WithDescriptions tests GetCommandsForHelp with various descriptions.
func TestGetCommandsForHelp_WithDescriptions(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldDir))
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))

	// Create magefile with commands with and without descriptions
	magefileContent := `//go:build mage

package main

// TestCmd is a test command with description
func TestCmd() error {
	return nil
}

// AnotherCmd has no comment
func AnotherCmd() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	helpLines := discovery.GetCommandsForHelp()

	// Should have help lines
	assert.GreaterOrEqual(t, len(helpLines), 1)

	// Check that output contains expected format
	found := false
	for _, line := range helpLines {
		if len(line) > 0 {
			found = true
			break
		}
	}
	assert.True(t, found, "Should have at least one non-empty help line")
}

// TestHasMagefile_Variations tests HasMagefile with different setups.
func TestHasMagefile_Variations(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(t *testing.T)
		expected bool
	}{
		{
			name: "magefiles directory exists",
			setup: func(t *testing.T) {
				require.NoError(t, os.Mkdir("magefiles", 0o750))
			},
			expected: true,
		},
		{
			name: "magefile.go exists",
			setup: func(t *testing.T) {
				require.NoError(t, os.WriteFile("magefile.go", []byte("package main"), 0o600))
			},
			expected: true,
		},
		{
			name: "both exist - prefers magefiles",
			setup: func(t *testing.T) {
				require.NoError(t, os.Mkdir("magefiles", 0o750))
				require.NoError(t, os.WriteFile("magefile.go", []byte("package main"), 0o600))
			},
			expected: true,
		},
		{
			name: "neither exists",
			setup: func(t *testing.T) {
				// Don't create anything
			},
			expected: false,
		},
		{
			name: "magefiles is a file not directory",
			setup: func(t *testing.T) {
				require.NoError(t, os.WriteFile("magefiles", []byte("not a dir"), 0o600))
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			oldDir, err := os.Getwd()
			require.NoError(t, err)
			t.Cleanup(func() {
				require.NoError(t, os.Chdir(oldDir))
			})

			require.NoError(t, os.Chdir(tmpDir))
			tt.setup(t)

			result := HasMagefile()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestConvertToMageFormat_SpecialCases tests convertToMageFormat with special cases.
func TestConvertToMageFormat_SpecialCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "colon at end",
			input:    "build:",
			expected: "build:",
		},
		{
			name:     "colon at start",
			input:    ":build",
			expected: ":build",
		},
		{
			name:     "only colon",
			input:    ":",
			expected: ":",
		},
		{
			name:     "spaces around colon",
			input:    "build : test",
			expected: "build : test", // Spaces and colon preserved
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToMageFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCleanCache_WithMageXDir tests cleanCache with .mage-x directory.
func TestCleanCache_WithMageXDir(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldDir))
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create .mage-x directory
	require.NoError(t, os.Mkdir(".mage-x", 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(".mage-x", "test.txt"), []byte("test"), 0o600))

	// Run cleanCache
	cleanCache()

	// Directory should be removed
	_, err = os.Stat(".mage-x")
	assert.True(t, os.IsNotExist(err), ".mage-x should be removed")
}

// TestCompileForMage_SuccessPath tests the success path of compileForMage.
func TestCompileForMage_SuccessPath(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldDir))
	})

	require.NoError(t, os.Chdir(tmpDir))

	outputFile := "test_magefile.go"

	// Call compileForMage
	compileForMage(outputFile)

	// File should exist
	_, err = os.Stat(outputFile)
	require.NoError(t, err, "output file should be created")

	// File should contain expected content
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	assert.Contains(t, string(content), "//go:build mage")
	assert.Contains(t, string(content), "package main")
	assert.Contains(t, string(content), "github.com/mrz1836/mage-x/pkg/mage")
}
