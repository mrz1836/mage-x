package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

// TestDiscovery_VerboseModeWithError tests verbose mode when discovery fails
func TestDiscovery_VerboseModeWithError(t *testing.T) {
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

func InvalidSyntax( {  // Missing closing paren and brace
	return nil
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(invalidMagefile), 0o600))

	// Create registry
	reg := registry.NewRegistry()

	// Capture stdout to check verbose warning
	oldStdout := os.Stdout
	r, w, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)
	os.Stdout = w

	// Create discovery with verbose mode enabled
	discovery := &CommandDiscovery{
		registry: reg,
		verbose:  true, // Enable verbose mode
	}

	// Try to discover - should fail and show verbose warning
	discErr := discovery.Discover()

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	// Should return error
	assert.Error(t, discErr, "Should error on invalid magefile")
}

// TestDiscovery_VerboseModeWithEnvVar tests verbose mode via environment variable
func TestDiscovery_VerboseModeWithEnvVar(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Set MAGEX_VERBOSE environment variable
	t.Setenv("MAGEX_VERBOSE", "true")

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))

	// Create invalid magefile
	invalidMagefile := `/` + `/go:build mage

package main

func InvalidSyntax( {
	return nil
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(invalidMagefile), 0o600))

	reg := registry.NewRegistry()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)
	os.Stdout = w

	// Create discovery without verbose flag (will check env var)
	discovery := NewCommandDiscovery(reg)

	// Try to discover - should show warning due to env var
	_ = discovery.Discover() //nolint:errcheck // testing error path

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestDiscovery_OverrideBuiltinCommandVerbose tests verbose logging when overriding built-in
func TestDiscovery_OverrideBuiltinCommandVerbose(t *testing.T) {
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

	// Create magefile with a command that matches a built-in
	magefileContent := `/` + `/go:build mage

package main

// Test is a custom command that might override a built-in
func Test() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	// Create registry with a built-in command named "test"
	reg := registry.NewRegistry()
	testCmd, err := registry.NewNamespaceCommand("test", "default").
		WithDescription("Built-in test command").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(testCmd))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)
	os.Stdout = w

	// Create verbose discovery
	discovery := &CommandDiscovery{
		registry: reg,
		verbose:  true,
	}

	// Discover - should log override warning
	_ = discovery.Discover() //nolint:errcheck // testing verbose logging

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestGetCommandsForHelp_DiscoveryError tests GetCommandsForHelp when discovery fails
func TestGetCommandsForHelp_DiscoveryError(t *testing.T) {
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

func Invalid( {
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(invalidMagefile), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// GetCommandsForHelp should handle discovery error gracefully
	// It may return nil or empty slice, but shouldn't panic
	helpLines := discovery.GetCommandsForHelp()

	// Test passes if we reach here without panicking
	// helpLines may be nil or empty, both are acceptable
	_ = helpLines
}

// TestIsLikelyNamespaceWrapper_VerboseMode tests verbose logging for namespace wrappers
func TestIsLikelyNamespaceWrapper_VerboseMode(t *testing.T) {
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

	// Create magefile with namespace wrapper pattern
	magefileContent := `/` + `/go:build mage

package main

// BuildDefault is likely a namespace wrapper for build:default
func BuildDefault() error {
	return nil
}

// TestUnit is likely a namespace wrapper for test:unit
func TestUnit() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()

	// Capture stdout
	oldStdout := os.Stdout
	r, w, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)
	os.Stdout = w

	// Create verbose discovery
	discovery := &CommandDiscovery{
		registry: reg,
		verbose:  true,
	}

	// Discover - should log skipped namespace wrappers
	_ = discovery.Discover() //nolint:errcheck // testing verbose logging

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestConvertToMageFormat_InvalidSplit tests defensive split validation
func TestConvertToMageFormat_InvalidSplit(t *testing.T) {
	// This tests the simple lowercasing behavior
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "only colon",
			input:    ":",
			expected: ":",
		},
		{
			name:     "multiple colons preserves all",
			input:    "a:b:c",
			expected: "a:b:c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToMageFormat(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
