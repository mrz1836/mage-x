package main

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

// TestGetMagefilePath_NoMagefileAlt tests when no magefile exists
func TestGetMagefilePath_NoMagefileAlt(t *testing.T) {
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
	path := GetMagefilePath()
	assert.Empty(t, path, "Should return empty string when no magefile exists")
}

// TestDelegateToMageWithTimeout_WithArgs tests delegation with command arguments
func TestDelegateToMageWithTimeout_WithArgs(t *testing.T) {
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

	// Create magefile that uses MAGE_ARGS
	magefileContent := `//go:build mage

package main

import (
	"fmt"
	"os"
)

func TestArgs() error {
	args := os.Getenv("MAGE_ARGS")
	if args != "" {
		fmt.Printf("Args: %s\n", args)
	}
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	// Delegate with arguments
	result := DelegateToMageWithTimeout("testArgs", 10*time.Second, "arg1", "arg2")

	// Should complete successfully
	assert.Equal(t, 0, result.ExitCode, "Should succeed with args")
	assert.NoError(t, result.Err)
}

// TestDelegateToMageWithTimeout_ContextTimeout tests timeout handling
func TestDelegateToMageWithTimeout_ContextTimeout(t *testing.T) {
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

	// Create magefile with a slow command
	magefileContent := `//go:build mage

package main

import "time"

func SlowCmd() error {
	time.Sleep(30 * time.Second)
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	// Delegate with very short timeout
	result := DelegateToMageWithTimeout("slowCmd", 100*time.Millisecond)

	// Should timeout
	assert.Equal(t, 124, result.ExitCode, "Should return timeout exit code")
	assert.Error(t, result.Err, "Should return timeout error")
}

// TestConvertToMageFormat_ComplexCases tests complex format conversions
func TestConvertToMageFormat_ComplexCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple namespace:method",
			input:    "build:Linux",
			expected: "buildlinux",
		},
		{
			name:     "preserve method case after first letter",
			input:    "test:RunAll",
			expected: "testrunAll",
		},
		{
			name:     "already simple format",
			input:    "buildAll",
			expected: "buildAll",
		},
		{
			name:     "lowercase namespace",
			input:    "Build:Default",
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

// TestTryCustomCommand_WithArgs tests tryCustomCommand with command arguments
func TestTryCustomCommand_WithArgs(t *testing.T) {
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

	// Create magefile with a command that uses args
	magefileContent := `//go:build mage

package main

import (
	"fmt"
	"os"
)

func TestWithArgs() error {
	args := os.Getenv("MAGE_ARGS")
	if args != "" {
		fmt.Printf("Received: %s\n", args)
	}
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	reg := registry.NewRegistry()
	discovery := NewCommandDiscovery(reg)

	// This should execute successfully with args
	exitCode, err := tryCustomCommand("testwithargs", []string{"foo", "bar"}, discovery)
	assert.Equal(t, 0, exitCode, "Should return 0 when command executes with args")
	assert.NoError(t, err, "Should return no error when command executes with args")
}

// TestListCommandsVerbose_WithEmptyDescriptionAlt tests verbose listing with empty descriptions
func TestListCommandsVerbose_WithEmptyDescriptionAlt(t *testing.T) {
	reg := registry.NewRegistry()

	// Add command without description
	emptyDescCmd, err := registry.NewNamespaceCommand("test", "empty").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(emptyDescCmd))

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

// TestListCommandsSimple_WithCustomCommandsAlt tests simple listing with custom commands
func TestListCommandsSimple_WithCustomCommandsAlt(t *testing.T) {
	reg := registry.NewRegistry()

	// Add some built-in commands
	for i := 0; i < 5; i++ {
		subCmd := "cmd" + string(rune('a'+i))
		cmd, err := registry.NewNamespaceCommand("test", subCmd).
			WithDescription("Test command").
			WithFunc(func() error { return nil }).
			Build()
		require.NoError(t, err)
		require.NoError(t, reg.Register(cmd))
	}

	// Add custom commands
	customCmds := []DiscoveredCommand{
		{Name: "custom1", Description: "Custom 1"},
		{Name: "custom2", Description: "Custom 2"},
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	listCommandsSimple(reg.List(), customCmds)

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup
}

// TestCleanCache_WithNestedDirectories tests cleanCache with nested structure
func TestCleanCache_WithNestedDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create nested cache structure
	require.NoError(t, os.MkdirAll(".mage/subdir", 0o750))
	require.NoError(t, os.WriteFile(".mage/file.txt", []byte("test"), 0o600))
	require.NoError(t, os.WriteFile(".mage/subdir/file2.txt", []byte("test2"), 0o600))

	require.NoError(t, os.MkdirAll(".mage-x/another", 0o750))
	require.NoError(t, os.WriteFile(".mage-x/another/file3.txt", []byte("test3"), 0o600))

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
	assert.True(t, os.IsNotExist(err1), "Should remove .mage with nested content")
	assert.True(t, os.IsNotExist(err2), "Should remove .mage-x with nested content")
}

// TestShowCategorizedCommands_WithMixedCategories tests categorized display
func TestShowCategorizedCommands_WithMixedCategories(t *testing.T) {
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

	// Add commands in multiple categories
	buildCmd, err := registry.NewNamespaceCommand("build", "linux").
		WithCategory("build").
		WithDescription("Build for Linux").
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

	deployCmd, err := registry.NewNamespaceCommand("deploy", "staging").
		WithCategory("deploy").
		WithDescription("Deploy to staging").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	require.NoError(t, reg.Register(deployCmd))

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

// TestValidateGoEnvironment_Success tests ValidateGoEnvironment when go is available
func TestValidateGoEnvironment_Success(t *testing.T) {
	// Skip if go is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available")
	}

	err := ValidateGoEnvironment()
	assert.NoError(t, err, "Should succeed when go is available")
}
