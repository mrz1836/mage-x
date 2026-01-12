package main

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

// TestTryCustomCommand_NoMagefileAlt tests tryCustomCommand when no magefile exists
func TestTryCustomCommand_NoMagefileAlt(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// No magefile exists
	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	exitCode, err := tryCustomCommand(context.Background(), "somecommand", []string{}, discovery)
	assert.Equal(t, 0, exitCode, "Should return 0 when no magefile exists")
	assert.NoError(t, err, "Should return no error when no magefile exists")
}

// TestTryCustomCommand_SuccessfulExecution tests tryCustomCommand with successful execution
func TestTryCustomCommand_SuccessfulExecution(t *testing.T) {
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

	// Create magefile with a simple command that succeeds
	magefileContent := `/` + `/go:build mage

package main

import "fmt"

func Success() error {
	fmt.Println("Success!")
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// This should execute successfully
	exitCode, err := tryCustomCommand(context.Background(), "success", []string{}, discovery)
	assert.Equal(t, 0, exitCode, "Should return 0 when command executes successfully")
	assert.NoError(t, err, "Should return no error when command executes successfully")
}

// TestTryCustomCommand_WithDiscoveredCommandAlt tests tryCustomCommand using discovered command
func TestTryCustomCommand_WithDiscoveredCommandAlt(t *testing.T) {
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

	// Create magefile with CamelCase command
	magefileContent := `/` + `/go:build mage

package main

import "fmt"

func BuildProject() error {
	fmt.Println("Building...")
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// Trigger discovery
	_ = discovery.HasCommand("buildproject")

	// This should use the discovered command's OriginalName
	exitCode, err := tryCustomCommand(context.Background(), "buildproject", []string{}, discovery)
	assert.Equal(t, 0, exitCode, "Should return 0 when using discovered command")
	assert.NoError(t, err, "Should return no error when using discovered command")
}
