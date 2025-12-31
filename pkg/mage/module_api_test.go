package mage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFindAllModulesAPI tests the public FindAllModules function
func TestFindAllModulesAPI(t *testing.T) {
	// Save and restore current directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalDir))
	})

	// Create a temporary test directory structure
	tmpDir := t.TempDir()

	// Create a simple module structure
	modPath := filepath.Join(tmpDir, "go.mod")
	err = os.WriteFile(modPath, []byte("module github.com/test/api\n"), 0o600)
	require.NoError(t, err)

	// Change to the temp directory
	require.NoError(t, os.Chdir(tmpDir))

	// Test the public API
	modules, err := FindAllModules()
	require.NoError(t, err)
	require.Len(t, modules, 1)
	assert.Equal(t, "github.com/test/api", modules[0].Module)
}

// TestRunCommandInModuleAPI tests the public RunCommandInModule function
func TestRunCommandInModuleAPI(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a simple module
	modPath := filepath.Join(tmpDir, "go.mod")
	err := os.WriteFile(modPath, []byte("module github.com/test/cmd\n"), 0o600)
	require.NoError(t, err)

	module := Module{
		Path:     tmpDir,
		Module:   "github.com/test/cmd",
		Relative: ".",
		IsRoot:   true,
		Name:     "cmd",
	}

	// Run a simple command that should succeed
	err = RunCommandInModule(module, "echo", "hello")
	require.NoError(t, err)
}

// TestRunCommandInModuleAPIError tests RunCommandInModule with a failing command
func TestRunCommandInModuleAPIError(t *testing.T) {
	// Save and restore runner to ensure clean state
	originalRunner := GetRunner()
	t.Cleanup(func() {
		_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
	})

	// Reset to default runner to avoid interference from other tests
	err := SetRunner(NewSecureCommandRunner())
	require.NoError(t, err)

	tmpDir := t.TempDir()

	module := Module{
		Path:     tmpDir,
		Module:   "github.com/test/fail",
		Relative: ".",
		IsRoot:   true,
		Name:     "fail",
	}

	// Run a command that doesn't exist, which reliably fails
	err = RunCommandInModule(module, "this-command-does-not-exist-intentionally")
	require.Error(t, err)
}

// TestFormatModuleErrorsAPI tests the public FormatModuleErrors function
func TestFormatModuleErrorsAPI(t *testing.T) {
	t.Run("empty errors", func(t *testing.T) {
		err := FormatModuleErrors([]ModuleError{})
		require.NoError(t, err)
	})

	t.Run("nil errors", func(t *testing.T) {
		err := FormatModuleErrors(nil)
		require.NoError(t, err)
	})

	t.Run("single error", func(t *testing.T) {
		errors := []ModuleError{
			{
				Module: Module{
					Path:     "/test/path",
					Module:   "github.com/test/single",
					Relative: ".",
					IsRoot:   true,
					Name:     "single",
				},
				Error: assert.AnError,
			},
		}
		err := FormatModuleErrors(errors)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "main module")
		assert.Contains(t, err.Error(), "1 module")
	})

	t.Run("multiple errors", func(t *testing.T) {
		errors := []ModuleError{
			{
				Module: Module{
					Path:     "/test/path1",
					Module:   "github.com/test/mod1",
					Relative: ".",
					IsRoot:   true,
					Name:     "mod1",
				},
				Error: assert.AnError,
			},
			{
				Module: Module{
					Path:     "/test/path2",
					Module:   "github.com/test/mod2",
					Relative: "submod",
					IsRoot:   false,
					Name:     "mod2",
				},
				Error: assert.AnError,
			},
		}
		err := FormatModuleErrors(errors)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "2 module")
		assert.Contains(t, err.Error(), "main module")
		assert.Contains(t, err.Error(), "submod")
	})
}

// TestSortModulesByDependencyAPI tests the public SortModulesByDependency function
func TestSortModulesByDependencyAPI(t *testing.T) {
	t.Run("empty modules", func(t *testing.T) {
		sorted, err := SortModulesByDependency([]Module{})
		require.NoError(t, err)
		assert.Empty(t, sorted)
	})

	t.Run("single module", func(t *testing.T) {
		modules := []Module{
			{
				Path:     "/test/path",
				Module:   "github.com/test/single",
				Relative: ".",
				IsRoot:   true,
				Name:     "single",
			},
		}
		sorted, err := SortModulesByDependency(modules)
		require.NoError(t, err)
		assert.Len(t, sorted, 1)
	})

	t.Run("multiple modules without dependencies", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create modules without dependencies
		mod1Path := filepath.Join(tmpDir, "mod1")
		mod2Path := filepath.Join(tmpDir, "mod2")

		require.NoError(t, os.MkdirAll(mod1Path, 0o750))
		require.NoError(t, os.MkdirAll(mod2Path, 0o750))

		require.NoError(t, os.WriteFile(filepath.Join(mod1Path, "go.mod"),
			[]byte("module github.com/test/mod1\n"), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(mod2Path, "go.mod"),
			[]byte("module github.com/test/mod2\n"), 0o600))

		modules := []Module{
			{Path: mod1Path, Module: "github.com/test/mod1", Relative: "mod1", IsRoot: false, Name: "mod1"},
			{Path: mod2Path, Module: "github.com/test/mod2", Relative: ".", IsRoot: true, Name: "mod2"},
		}

		sorted, err := SortModulesByDependency(modules)
		require.NoError(t, err)
		assert.Len(t, sorted, 2)
		// Root should come first
		assert.True(t, sorted[0].IsRoot)
	})
}

// TestParseModuleDependenciesAPI tests the public ParseModuleDependencies function
func TestParseModuleDependenciesAPI(t *testing.T) {
	t.Run("no dependencies", func(t *testing.T) {
		tmpDir := t.TempDir()
		modPath := filepath.Join(tmpDir, "go.mod")
		require.NoError(t, os.WriteFile(modPath, []byte("module github.com/test/nodeps\n"), 0o600))

		module := Module{
			Path:   tmpDir,
			Module: "github.com/test/nodeps",
		}

		deps, err := ParseModuleDependencies(module, map[string]bool{})
		require.NoError(t, err)
		assert.Empty(t, deps)
	})

	t.Run("with local replace directive", func(t *testing.T) {
		tmpDir := t.TempDir()
		modPath := filepath.Join(tmpDir, "go.mod")
		content := `module github.com/test/withdeps

require github.com/test/dep v0.0.0

replace github.com/test/dep => ../dep
`
		require.NoError(t, os.WriteFile(modPath, []byte(content), 0o600))

		module := Module{
			Path:   tmpDir,
			Module: "github.com/test/withdeps",
		}

		allModules := map[string]bool{
			"github.com/test/dep":      true,
			"github.com/test/withdeps": true,
		}

		deps, err := ParseModuleDependencies(module, allModules)
		require.NoError(t, err)
		assert.Contains(t, deps, "github.com/test/dep")
	})

	t.Run("nonexistent go.mod", func(t *testing.T) {
		module := Module{
			Path:   "/nonexistent/path",
			Module: "github.com/test/missing",
		}

		_, err := ParseModuleDependencies(module, map[string]bool{})
		require.Error(t, err)
	})
}

// TestDisplayModuleSummaryAPI tests the public DisplayModuleSummary function
func TestDisplayModuleSummaryAPI(t *testing.T) {
	t.Run("single module no deps", func(t *testing.T) {
		modules := []Module{
			{
				Path:     "/test/path",
				Module:   "github.com/test/single",
				Relative: ".",
				IsRoot:   true,
				Name:     "single",
			},
		}

		// Should not panic
		assert.NotPanics(t, func() {
			DisplayModuleSummary(modules, map[string][]string{})
		})
	})

	t.Run("multiple modules with deps", func(t *testing.T) {
		modules := []Module{
			{Path: "/test/path1", Module: "github.com/test/mod1", Relative: ".", IsRoot: true, Name: "mod1"},
			{Path: "/test/path2", Module: "github.com/test/mod2", Relative: "submod", IsRoot: false, Name: "mod2"},
		}

		moduleDeps := map[string][]string{
			"github.com/test/mod2": {"github.com/test/mod1"},
		}

		assert.NotPanics(t, func() {
			DisplayModuleSummary(modules, moduleDeps)
		})
	})
}

// TestModuleInfoGetPath tests the GetPath method
func TestModuleInfoGetPath(t *testing.T) {
	m := ModuleInfo{
		Path:     "/test/module/path",
		Module:   "github.com/test/module",
		Relative: ".",
		IsRoot:   true,
		Name:     "module",
	}

	assert.Equal(t, "/test/module/path", m.GetPath())
}
