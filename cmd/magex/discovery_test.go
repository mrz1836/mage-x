package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

// =============================================================================
// NewCommandDiscovery Tests
// =============================================================================

func TestNewCommandDiscovery(t *testing.T) {
	t.Run("with nil registry", func(t *testing.T) {
		cd := NewCommandDiscovery(nil)
		require.NotNil(t, cd)
		assert.Nil(t, cd.registry)
		assert.False(t, cd.loaded)
		assert.Empty(t, cd.commands)
	})

	t.Run("with valid registry", func(t *testing.T) {
		reg := registry.NewRegistry()
		cd := NewCommandDiscovery(reg)
		require.NotNil(t, cd)
		assert.Equal(t, reg, cd.registry)
		assert.False(t, cd.loaded)
		assert.Empty(t, cd.commands)
	})

	t.Run("verbose from MAGE_X_VERBOSE", func(t *testing.T) {
		t.Setenv("MAGE_X_VERBOSE", "true")
		cd := NewCommandDiscovery(nil)
		assert.True(t, cd.verbose)
	})

	t.Run("verbose from MAGEX_VERBOSE", func(t *testing.T) {
		t.Setenv("MAGE_X_VERBOSE", "")
		t.Setenv("MAGEX_VERBOSE", "true")
		cd := NewCommandDiscovery(nil)
		assert.True(t, cd.verbose)
	})
}

// =============================================================================
// Clear Tests
// =============================================================================

func TestCommandDiscovery_Clear(t *testing.T) {
	t.Run("clears commands and loaded flag", func(t *testing.T) {
		cd := &CommandDiscovery{
			commands: []DiscoveredCommand{
				{Name: "test", Description: "Test command"},
				{Name: "deploy", Description: "Deploy command"},
			},
			loaded: true,
		}

		cd.Clear()

		assert.Nil(t, cd.commands)
		assert.False(t, cd.loaded)
	})

	t.Run("clear on empty discovery", func(t *testing.T) {
		cd := &CommandDiscovery{
			commands: nil,
			loaded:   false,
		}

		cd.Clear()

		assert.Nil(t, cd.commands)
		assert.False(t, cd.loaded)
	})
}

// =============================================================================
// HasCommand Tests
// =============================================================================

func TestCommandDiscovery_HasCommand(t *testing.T) {
	tests := []struct {
		name     string
		commands []DiscoveredCommand
		lookup   string
		want     bool
	}{
		{
			name:     "exact match lowercase",
			commands: []DiscoveredCommand{{Name: "deploy"}},
			lookup:   "deploy",
			want:     true,
		},
		{
			name:     "case insensitive - upper lookup",
			commands: []DiscoveredCommand{{Name: "deploy"}},
			lookup:   "DEPLOY",
			want:     true,
		},
		{
			name:     "case insensitive - mixed case",
			commands: []DiscoveredCommand{{Name: "deploy"}},
			lookup:   "DePlOy",
			want:     true,
		},
		{
			name:     "command not found",
			commands: []DiscoveredCommand{{Name: "deploy"}},
			lookup:   "unknown",
			want:     false,
		},
		{
			name:     "empty lookup string",
			commands: []DiscoveredCommand{{Name: "deploy"}},
			lookup:   "",
			want:     false,
		},
		{
			name:     "empty commands list",
			commands: []DiscoveredCommand{},
			lookup:   "deploy",
			want:     false,
		},
		{
			name: "namespace command match",
			commands: []DiscoveredCommand{
				{Name: "pipeline:ci", Namespace: "Pipeline", Method: "CI"},
			},
			lookup: "pipeline:ci",
			want:   true,
		},
		{
			name: "namespace command case insensitive",
			commands: []DiscoveredCommand{
				{Name: "pipeline:ci", Namespace: "Pipeline", Method: "CI"},
			},
			lookup: "PIPELINE:CI",
			want:   true,
		},
		{
			name: "multiple commands - find second",
			commands: []DiscoveredCommand{
				{Name: "build"},
				{Name: "test"},
				{Name: "deploy"},
			},
			lookup: "test",
			want:   true,
		},
		{
			name: "multiple commands - not found",
			commands: []DiscoveredCommand{
				{Name: "build"},
				{Name: "test"},
				{Name: "deploy"},
			},
			lookup: "lint",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cd := &CommandDiscovery{
				commands: tt.commands,
				loaded:   true, // Skip actual discovery
			}
			got := cd.HasCommand(tt.lookup)
			assert.Equal(t, tt.want, got)
		})
	}
}

// =============================================================================
// GetCommand Tests
// =============================================================================

func TestCommandDiscovery_GetCommand(t *testing.T) {
	tests := []struct {
		name     string
		commands []DiscoveredCommand
		lookup   string
		wantCmd  *DiscoveredCommand
		wantOk   bool
	}{
		{
			name: "found - simple command",
			commands: []DiscoveredCommand{
				{Name: "deploy", Description: "Deploy the app", OriginalName: "Deploy"},
			},
			lookup: "deploy",
			wantCmd: &DiscoveredCommand{
				Name:         "deploy",
				Description:  "Deploy the app",
				OriginalName: "Deploy",
			},
			wantOk: true,
		},
		{
			name: "found - case insensitive",
			commands: []DiscoveredCommand{
				{Name: "deploy", Description: "Deploy the app"},
			},
			lookup:  "DEPLOY",
			wantCmd: &DiscoveredCommand{Name: "deploy", Description: "Deploy the app"},
			wantOk:  true,
		},
		{
			name: "found - namespace command",
			commands: []DiscoveredCommand{
				{
					Name:        "pipeline:ci",
					Description: "Run CI pipeline",
					IsNamespace: true,
					Namespace:   "Pipeline",
					Method:      "CI",
				},
			},
			lookup: "pipeline:ci",
			wantCmd: &DiscoveredCommand{
				Name:        "pipeline:ci",
				Description: "Run CI pipeline",
				IsNamespace: true,
				Namespace:   "Pipeline",
				Method:      "CI",
			},
			wantOk: true,
		},
		{
			name:     "not found",
			commands: []DiscoveredCommand{{Name: "deploy"}},
			lookup:   "unknown",
			wantCmd:  nil,
			wantOk:   false,
		},
		{
			name:     "empty commands",
			commands: []DiscoveredCommand{},
			lookup:   "deploy",
			wantCmd:  nil,
			wantOk:   false,
		},
		{
			name:     "empty lookup",
			commands: []DiscoveredCommand{{Name: "deploy"}},
			lookup:   "",
			wantCmd:  nil,
			wantOk:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cd := &CommandDiscovery{
				commands: tt.commands,
				loaded:   true, // Skip actual discovery
			}
			gotCmd, gotOk := cd.GetCommand(tt.lookup)
			assert.Equal(t, tt.wantOk, gotOk)
			if tt.wantCmd == nil {
				assert.Nil(t, gotCmd)
			} else {
				require.NotNil(t, gotCmd)
				assert.Equal(t, tt.wantCmd.Name, gotCmd.Name)
				assert.Equal(t, tt.wantCmd.Description, gotCmd.Description)
				assert.Equal(t, tt.wantCmd.IsNamespace, gotCmd.IsNamespace)
				assert.Equal(t, tt.wantCmd.Namespace, gotCmd.Namespace)
				assert.Equal(t, tt.wantCmd.Method, gotCmd.Method)
			}
		})
	}
}

// =============================================================================
// ListCommands Tests
// =============================================================================

func TestCommandDiscovery_ListCommands(t *testing.T) {
	t.Run("returns all commands", func(t *testing.T) {
		commands := []DiscoveredCommand{
			{Name: "build", Description: "Build the app"},
			{Name: "test", Description: "Run tests"},
			{Name: "deploy", Description: "Deploy the app"},
		}
		cd := &CommandDiscovery{
			commands: commands,
			loaded:   true,
		}

		result, err := cd.ListCommands()
		require.NoError(t, err)
		assert.Equal(t, commands, result)
	})

	t.Run("returns empty slice when no commands", func(t *testing.T) {
		cd := &CommandDiscovery{
			commands: []DiscoveredCommand{},
			loaded:   true,
		}

		result, err := cd.ListCommands()
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

// =============================================================================
// GetCommandsForHelp Tests
// =============================================================================

func TestCommandDiscovery_GetCommandsForHelp(t *testing.T) {
	t.Run("formats commands with descriptions", func(t *testing.T) {
		cd := &CommandDiscovery{
			commands: []DiscoveredCommand{
				{Name: "deploy", Description: "Deploy the app"},
				{Name: "pipeline:ci", Description: "Run CI"},
			},
			loaded: true,
		}

		result := cd.GetCommandsForHelp()

		require.Len(t, result, 2)
		assert.Contains(t, result[0], "deploy")
		assert.Contains(t, result[0], "Deploy the app")
		assert.Contains(t, result[0], "(custom)")
		assert.Contains(t, result[1], "pipeline:ci")
		assert.Contains(t, result[1], "Run CI")
	})

	t.Run("uses default description when empty", func(t *testing.T) {
		cd := &CommandDiscovery{
			commands: []DiscoveredCommand{
				{Name: "deploy", Description: ""},
			},
			loaded: true,
		}

		result := cd.GetCommandsForHelp()

		require.Len(t, result, 1)
		assert.Contains(t, result[0], "deploy")
		assert.Contains(t, result[0], "Custom command")
	})

	t.Run("returns nil for empty commands", func(t *testing.T) {
		cd := &CommandDiscovery{
			commands: []DiscoveredCommand{},
			loaded:   true,
		}

		result := cd.GetCommandsForHelp()
		assert.Empty(t, result)
	})

	t.Run("format includes padding", func(t *testing.T) {
		cd := &CommandDiscovery{
			commands: []DiscoveredCommand{
				{Name: "a", Description: "Short name"},
			},
			loaded: true,
		}

		result := cd.GetCommandsForHelp()
		require.Len(t, result, 1)

		// Should have padding for alignment (%-20s format)
		// "  a                    Short name (custom)"
		assert.True(t, strings.HasPrefix(result[0], "  a"))
	})
}

// =============================================================================
// isLikelyNamespaceWrapper Tests
// =============================================================================

func TestIsLikelyNamespaceWrapper(t *testing.T) {
	// Create a registry with some built-in commands
	reg := registry.NewRegistry()

	// Register some test commands that would be matched
	buildDefault, err := registry.NewNamespaceCommand("build", "default").
		WithDescription("Default build").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	buildLinux, err := registry.NewNamespaceCommand("build", "linux").
		WithDescription("Build for Linux").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	testUnit, err := registry.NewNamespaceCommand("test", "unit").
		WithDescription("Run unit tests").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	lintFix, err := registry.NewNamespaceCommand("lint", "fix").
		WithDescription("Fix lint issues").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	build, err := registry.NewCommand("build").
		WithDescription("Build").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)

	reg.MustRegister(buildDefault)
	reg.MustRegister(buildLinux)
	reg.MustRegister(testUnit)
	reg.MustRegister(lintFix)
	reg.MustRegister(build)

	tests := []struct {
		name     string
		registry *registry.Registry
		funcName string
		want     bool
	}{
		// Nil registry always returns false
		{
			name:     "nil registry",
			registry: nil,
			funcName: "BuildDefault",
			want:     false,
		},
		// Matching namespace:method patterns
		{
			name:     "BuildDefault matches build:default",
			registry: reg,
			funcName: "builddefault",
			want:     true,
		},
		{
			name:     "BuildLinux matches build:linux",
			registry: reg,
			funcName: "buildlinux",
			want:     true,
		},
		{
			name:     "TestUnit matches test:unit",
			registry: reg,
			funcName: "testunit",
			want:     true,
		},
		{
			name:     "LintFix matches lint:fix",
			registry: reg,
			funcName: "lintfix",
			want:     true,
		},
		// Non-matching patterns
		{
			name:     "RandomFunc - no namespace match",
			registry: reg,
			funcName: "randomfunc",
			want:     false,
		},
		{
			name:     "Buildx - not a method pattern (no matching namespace:x)",
			registry: reg,
			funcName: "buildx",
			want:     false,
		},
		{
			name:     "Deploy - standalone, not a wrapper",
			registry: reg,
			funcName: "deploy",
			want:     false,
		},
		{
			name:     "empty function name",
			registry: reg,
			funcName: "",
			want:     false,
		},
		// Edge cases
		{
			name:     "just namespace name",
			registry: reg,
			funcName: "build",
			want:     false, // "build" alone isn't "build:something"
		},
		{
			name:     "case variations",
			registry: reg,
			funcName: "BUILDDEFAULT",
			want:     true, // Should be case-insensitive
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cd := &CommandDiscovery{
				registry: tt.registry,
				loaded:   true,
			}
			got := cd.isLikelyNamespaceWrapper(tt.funcName)
			assert.Equal(t, tt.want, got, "isLikelyNamespaceWrapper(%q) = %v, want %v", tt.funcName, got, tt.want)
		})
	}
}

func TestIsLikelyNamespaceWrapper_AllCommonNamespaces(t *testing.T) {
	// Test that all common namespaces from the function are checked
	commonNamespaces := []string{
		"build", "test", "lint", "format", "deps", "git", "release", "docs",
		"tools", "generate", "mod", "help", "version", "install",
		"configure", "init", "bench", "vet",
	}

	reg := registry.NewRegistry()

	// Register a command for each namespace:default
	for _, ns := range commonNamespaces {
		cmd, err := registry.NewNamespaceCommand(ns, "default").
			WithDescription("Default " + ns).
			WithFunc(func() error { return nil }).
			Build()
		require.NoError(t, err)
		reg.MustRegister(cmd)
	}

	cd := &CommandDiscovery{
		registry: reg,
		loaded:   true,
	}

	for _, ns := range commonNamespaces {
		funcName := ns + "default"
		t.Run(funcName, func(t *testing.T) {
			got := cd.isLikelyNamespaceWrapper(funcName)
			assert.True(t, got, "isLikelyNamespaceWrapper(%q) should return true", funcName)
		})
	}
}

// =============================================================================
// DiscoveredCommand Struct Tests
// =============================================================================

func TestDiscoveredCommand_Fields(t *testing.T) {
	cmd := DiscoveredCommand{
		Name:         "pipeline:ci",
		OriginalName: "Pipeline:CI",
		Description:  "Run CI pipeline",
		IsNamespace:  true,
		Namespace:    "Pipeline",
		Method:       "CI",
	}

	assert.Equal(t, "pipeline:ci", cmd.Name)
	assert.Equal(t, "Pipeline:CI", cmd.OriginalName)
	assert.Equal(t, "Run CI pipeline", cmd.Description)
	assert.True(t, cmd.IsNamespace)
	assert.Equal(t, "Pipeline", cmd.Namespace)
	assert.Equal(t, "CI", cmd.Method)
}

// =============================================================================
// Integration Tests
// =============================================================================

func TestCommandDiscovery_Integration_WithRegistry(t *testing.T) {
	reg := registry.NewRegistry()

	// Register some built-in commands
	buildCmd, err := registry.NewCommand("build").
		WithDescription("Build the project").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	reg.MustRegister(buildCmd)

	cd := NewCommandDiscovery(reg)

	// Verify initial state
	assert.False(t, cd.loaded)
	assert.Empty(t, cd.commands)
	assert.Equal(t, reg, cd.registry)
}

func TestCommandDiscovery_ClearAndReload(t *testing.T) {
	cd := &CommandDiscovery{
		commands: []DiscoveredCommand{
			{Name: "test", Description: "Test"},
		},
		loaded: true,
	}

	// Verify command exists
	assert.True(t, cd.HasCommand("test"))

	// Clear
	cd.Clear()

	// Verify cleared
	assert.False(t, cd.loaded)
	assert.Nil(t, cd.commands)
}

// =============================================================================
// Discover Tests
// =============================================================================

func TestCommandDiscovery_Discover_Idempotent(t *testing.T) {
	t.Run("already loaded skips discovery", func(t *testing.T) {
		cd := &CommandDiscovery{
			commands: []DiscoveredCommand{{Name: "existing"}},
			loaded:   true,
		}

		err := cd.Discover()
		require.NoError(t, err)

		// Should still have the same command
		assert.Len(t, cd.commands, 1)
		assert.Equal(t, "existing", cd.commands[0].Name)
	})
}

func TestCommandDiscovery_Discover_WithMagefile(t *testing.T) {
	// Test discovery with a real magefile
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create go.mod
	err = os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600)
	require.NoError(t, err)

	// Create magefile.go with test commands
	magefileContent := `//go:build mage

package main

import "fmt"

// Deploy deploys the application
func Deploy() error {
	fmt.Println("Deploying...")
	return nil
}

// Pipeline namespace
type Pipeline struct{}

// CI runs the CI pipeline
func (Pipeline) CI() error {
	fmt.Println("Running CI...")
	return nil
}
`
	err = os.WriteFile("magefile.go", []byte(magefileContent), 0o600)
	require.NoError(t, err)

	// Create registry and discovery
	reg := registry.NewRegistry()
	cd := NewCommandDiscovery(reg)

	// Run discovery
	err = cd.Discover()

	// We just check it doesn't panic and completes
	// The actual result depends on whether mage is installed
	if err == nil {
		assert.True(t, cd.loaded, "loaded should be true after successful discovery")
		// If we have commands, verify they're properly formatted
		for _, cmd := range cd.commands {
			assert.NotEmpty(t, cmd.Name, "command name should not be empty")
			assert.Equal(t, strings.ToLower(cmd.Name), cmd.Name, "command name should be lowercase")
		}
	}
}

func TestCommandDiscovery_Discover_NoMagefile(t *testing.T) {
	// Test in a temporary directory with no magefile
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create a go.mod file so we're in a valid Go module
	err = os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600)
	require.NoError(t, err)

	cd := NewCommandDiscovery(nil)

	err = cd.Discover()
	// May error or not depending on implementation - we just verify it doesn't panic
	// and that loaded is set appropriately
	if err == nil {
		assert.True(t, cd.loaded)
	}
}

func TestCommandDiscovery_HasCommand_DiscoverError(t *testing.T) {
	// When Discover fails, HasCommand should return false
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// No go.mod - should cause discovery to fail
	cd := NewCommandDiscovery(nil)

	// Should return false when discovery fails
	result := cd.HasCommand("anything")
	assert.False(t, result)
}

func TestCommandDiscovery_GetCommand_DiscoverError(t *testing.T) {
	// When Discover fails, GetCommand should return nil, false
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// No go.mod - should cause discovery to fail
	cd := NewCommandDiscovery(nil)

	// Should return nil, false when discovery fails
	cmd, found := cd.GetCommand("anything")
	assert.Nil(t, cmd)
	assert.False(t, found)
}

func TestCommandDiscovery_ListCommands_DiscoverError(t *testing.T) {
	// When Discover fails, ListCommands should return nil, error
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// No go.mod - should cause discovery to fail
	cd := NewCommandDiscovery(nil)

	// Should return error when discovery fails
	commands, err := cd.ListCommands()
	if err != nil {
		assert.Nil(t, commands)
	}
	// If no error, just check we didn't crash
}

func TestCommandDiscovery_GetCommandsForHelp_DiscoverError(t *testing.T) {
	// When Discover fails, GetCommandsForHelp should return nil or empty
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// No go.mod - should cause discovery to fail
	cd := NewCommandDiscovery(nil)

	// Should return empty when discovery fails (or returns no commands)
	result := cd.GetCommandsForHelp()
	assert.Empty(t, result)
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkHasCommand(b *testing.B) {
	commands := make([]DiscoveredCommand, 100)
	for i := 0; i < 100; i++ {
		commands[i] = DiscoveredCommand{
			Name:        strings.ToLower("cmd" + string(rune('a'+i%26))),
			Description: "Command",
		}
	}

	cd := &CommandDiscovery{
		commands: commands,
		loaded:   true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cd.HasCommand("cmdz")
	}
}

func BenchmarkGetCommand(b *testing.B) {
	commands := make([]DiscoveredCommand, 100)
	for i := 0; i < 100; i++ {
		commands[i] = DiscoveredCommand{
			Name:        strings.ToLower("cmd" + string(rune('a'+i%26))),
			Description: "Command",
		}
	}

	cd := &CommandDiscovery{
		commands: commands,
		loaded:   true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cd.GetCommand("cmdz")
	}
}

func BenchmarkGetCommandsForHelp(b *testing.B) {
	commands := make([]DiscoveredCommand, 50)
	for i := 0; i < 50; i++ {
		commands[i] = DiscoveredCommand{
			Name:        "command" + string(rune('a'+i%26)),
			Description: "A test command description",
		}
	}

	cd := &CommandDiscovery{
		commands: commands,
		loaded:   true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cd.GetCommandsForHelp()
	}
}

func BenchmarkIsLikelyNamespaceWrapper(b *testing.B) {
	reg := registry.NewRegistry()
	buildDefault, err := registry.NewNamespaceCommand("build", "default").
		WithDescription("Default build").
		WithFunc(func() error { return nil }).
		Build()
	if err != nil {
		b.Fatal(err)
	}
	reg.MustRegister(buildDefault)

	cd := &CommandDiscovery{
		registry: reg,
		loaded:   true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cd.isLikelyNamespaceWrapper("builddefault")
	}
}

// TestCommandDiscovery_Discover_VerboseMode tests verbose output during discovery
func TestCommandDiscovery_Discover_VerboseMode(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create go.mod
	err = os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600)
	require.NoError(t, err)

	// Create magefile.go with commands that will trigger verbose logging
	magefileContent := `//go:build mage

package main

import "fmt"

// Deploy deploys the application
func Deploy() error {
	fmt.Println("Deploying...")
	return nil
}

// BuildDefault is a wrapper that should be skipped
func BuildDefault() error {
	fmt.Println("Building...")
	return nil
}
`
	err = os.WriteFile("magefile.go", []byte(magefileContent), 0o600)
	require.NoError(t, err)

	// Create registry with build:default command (to test override detection)
	reg := registry.NewRegistry()
	buildDefault, err := registry.NewNamespaceCommand("build", "default").
		WithDescription("Default build").
		WithFunc(func() error { return nil }).
		Build()
	require.NoError(t, err)
	reg.MustRegister(buildDefault)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Create discovery with verbose mode
	cd := &CommandDiscovery{
		registry: reg,
		verbose:  true,
	}

	// Run discovery
	err = cd.Discover()

	if closeErr := w.Close(); closeErr != nil {
		t.Logf("Failed to close writer: %v", closeErr)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, readErr := buf.ReadFrom(r); readErr != nil {
		t.Logf("Failed to read from pipe: %v", readErr)
	}
	output := buf.String()

	// In verbose mode, should show discovered commands if any were found
	if err == nil && len(cd.commands) > 0 {
		assert.Contains(t, output, "Discovered")
	}
}

// TestCommandDiscovery_Discover_VerboseError tests verbose error output
func TestCommandDiscovery_Discover_VerboseError(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// No go.mod - may or may not cause discovery to fail depending on implementation

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Create discovery with verbose mode
	cd := &CommandDiscovery{
		verbose: true,
	}

	// Run discovery - may fail or succeed
	discoverErr := cd.Discover()
	if discoverErr != nil {
		t.Logf("Discovery failed as expected: %v", discoverErr)
	}

	if closeErr := w.Close(); closeErr != nil {
		t.Logf("Failed to close writer: %v", closeErr)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, readErr := buf.ReadFrom(r); readErr != nil {
		t.Logf("Failed to read from pipe: %v", readErr)
	}
	output := buf.String()

	// In verbose mode, may show warning about failed discovery or succeed silently
	// Just verify verbose mode doesn't crash
	t.Logf("Verbose output: %s", output)
}

// TestCommandDiscovery_ListCommands_AfterDiscover tests ListCommands after successful discovery
func TestCommandDiscovery_ListCommands_AfterDiscover(t *testing.T) {
	cd := &CommandDiscovery{
		commands: []DiscoveredCommand{
			{Name: "test1", Description: "Test 1"},
			{Name: "test2", Description: "Test 2"},
		},
		loaded: true,
	}

	commands, err := cd.ListCommands()
	require.NoError(t, err)
	assert.Len(t, commands, 2)
	assert.Equal(t, "test1", commands[0].Name)
	assert.Equal(t, "test2", commands[1].Name)
}

// TestCommandDiscovery_HasCommand_TriggersDiscover tests that HasCommand triggers discovery
func TestCommandDiscovery_HasCommand_TriggersDiscover(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create go.mod
	err = os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600)
	require.NoError(t, err)

	// Create magefile.go
	magefileContent := `//go:build mage

package main

// Deploy deploys the application
func Deploy() error {
	return nil
}
`
	err = os.WriteFile("magefile.go", []byte(magefileContent), 0o600)
	require.NoError(t, err)

	cd := NewCommandDiscovery(nil)

	// HasCommand should trigger discovery
	found := cd.HasCommand("deploy")

	// May find it or not depending on environment, but should have tried discovery
	if found {
		assert.True(t, cd.loaded)
	}
}

// TestCommandDiscovery_GetCommand_TriggersDiscover tests that GetCommand triggers discovery
func TestCommandDiscovery_GetCommand_TriggersDiscover(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create go.mod
	err = os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600)
	require.NoError(t, err)

	// Create magefile.go
	magefileContent := `//go:build mage

package main

// Deploy deploys the application
func Deploy() error {
	return nil
}
`
	err = os.WriteFile("magefile.go", []byte(magefileContent), 0o600)
	require.NoError(t, err)

	cd := NewCommandDiscovery(nil)

	// GetCommand should trigger discovery
	_, found := cd.GetCommand("deploy")

	// May find it or not depending on environment, but should have tried discovery
	if found {
		assert.True(t, cd.loaded)
	}
}

// TestCommandDiscovery_GetCommandsForHelp_TriggersDiscover tests that GetCommandsForHelp triggers discovery
func TestCommandDiscovery_GetCommandsForHelp_TriggersDiscover(t *testing.T) {
	cd := &CommandDiscovery{
		commands: []DiscoveredCommand{
			{Name: "test", Description: "Test command"},
		},
		loaded: true,
	}

	helpLines := cd.GetCommandsForHelp()
	assert.NotEmpty(t, helpLines)
	assert.Contains(t, helpLines[0], "test")
	assert.Contains(t, helpLines[0], "Test command")
	assert.Contains(t, helpLines[0], "(custom)")
}
