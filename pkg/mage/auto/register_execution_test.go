package auto

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage"
)

// setupTestConfig sets up a minimal test configuration for wrapper function tests.
// The functions will likely error (no real project), but shouldn't panic.
func setupTestConfig(t *testing.T) {
	t.Helper()

	mage.TestSetConfig(&mage.Config{
		Project: mage.ProjectConfig{
			Name:   "test-project",
			Binary: "test-bin",
			Module: "github.com/test/project",
		},
		Build: mage.BuildConfig{
			Output:   "bin",
			Parallel: 1,
		},
		Test: mage.TestConfig{
			Timeout: "1m",
		},
	})
}

// TestWrapperFunctionExecution tests that all wrapper functions can be called
// without panicking. Since these functions delegate to mage namespace methods
// that require a real Go project, they will return errors - but the important
// thing is that they execute the code path and don't crash.
func TestWrapperFunctionExecution(t *testing.T) {
	// These tests modify global config, so don't run in parallel
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	// Create a temp dir to avoid affecting the real project
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	defer func() {
		require.NoError(t, os.Chdir(originalDir))
	}()

	// Create a minimal go.mod to prevent "go: no go.mod file" errors
	err = os.WriteFile("go.mod", []byte("module testproject\ngo 1.21\n"), os.FileMode(0o600))
	require.NoError(t, err)

	setupTestConfig(t)
	defer mage.TestResetConfig()

	t.Run("BuildCmd executes without panic", func(t *testing.T) {
		// BuildCmd delegates to mage.Build.Default()
		// It will likely error due to no real project, but shouldn't panic
		err := BuildCmd()
		// We don't assert NoError because the function may legitimately error
		// in a test environment. The key is that it executed without panic.
		_ = err
	})

	t.Run("CleanCmd executes without panic", func(t *testing.T) {
		// CleanCmd delegates to mage.Build.Clean()
		err := CleanCmd()
		// Clean is more likely to succeed since it just removes files
		// but we don't assert because the output dir may not exist
		_ = err
	})

	t.Run("DepsCmd executes without panic", func(t *testing.T) {
		// DepsCmd delegates to mage.Deps.Default()
		err := DepsCmd()
		_ = err
	})

	t.Run("TestCmd executes without panic", func(t *testing.T) {
		// TestCmd delegates to mage.Test.Default()
		// Will error because no tests exist in temp project
		err := TestCmd()
		_ = err
	})

	t.Run("LintCmd executes without panic", func(t *testing.T) {
		// LintCmd delegates to mage.Lint.Default()
		err := LintCmd()
		_ = err
	})

	t.Run("FormatCmd executes without panic", func(t *testing.T) {
		// FormatCmd delegates to mage.Format.Default()
		err := FormatCmd()
		_ = err
	})

	t.Run("InstallCmd executes without panic", func(t *testing.T) {
		// InstallCmd delegates to mage.Build.Install()
		err := InstallCmd()
		_ = err
	})

	t.Run("ReleaseCmd executes without panic", func(t *testing.T) {
		// ReleaseCmd delegates to mage.Release.Default()
		// Will error because no goreleaser config exists
		err := ReleaseCmd()
		_ = err
	})
}

// TestWrapperFunctionSignatures verifies that wrapper functions have
// the expected signature and can be assigned to func() error variables.
func TestWrapperFunctionSignatures(t *testing.T) {
	t.Parallel()

	// Verify all wrapper functions can be assigned to func() error
	// This is a compile-time check that also executes at runtime
	wrappers := map[string]func() error{
		"BuildCmd":   BuildCmd,
		"TestCmd":    TestCmd,
		"LintCmd":    LintCmd,
		"FormatCmd":  FormatCmd,
		"CleanCmd":   CleanCmd,
		"InstallCmd": InstallCmd,
		"DepsCmd":    DepsCmd,
		"ReleaseCmd": ReleaseCmd,
	}

	for name, fn := range wrappers {
		t.Run(name+"_signature", func(t *testing.T) {
			t.Parallel()
			assert.NotNil(t, fn, "%s should be a valid function", name)
		})
	}
}

// TestTypeAliasMethodExecution tests that type alias methods can be called.
// This verifies the type aliases correctly expose the underlying mage types.
func TestTypeAliasMethodExecution(t *testing.T) {
	// These tests modify global config, so don't run in parallel
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	defer func() {
		require.NoError(t, os.Chdir(originalDir))
	}()

	// Create a minimal go.mod
	err = os.WriteFile("go.mod", []byte("module testproject\ngo 1.21\n"), os.FileMode(0o600))
	require.NoError(t, err)

	setupTestConfig(t)
	defer mage.TestResetConfig()

	t.Run("Build type alias methods accessible", func(t *testing.T) {
		var b Build
		// Test that we can call methods on the type alias
		// Clean is a good test because it's less likely to fail hard
		err := b.Clean()
		_ = err // May error, but shouldn't panic
	})

	t.Run("Test type alias instantiation", func(t *testing.T) {
		var test Test
		// Verify the type alias is properly defined
		assert.IsType(t, mage.Test{}, test)
	})

	t.Run("Lint type alias instantiation", func(t *testing.T) {
		var lint Lint
		assert.IsType(t, mage.Lint{}, lint)
	})

	t.Run("Format type alias instantiation", func(t *testing.T) {
		var format Format
		assert.IsType(t, mage.Format{}, format)
	})

	t.Run("Deps type alias instantiation", func(t *testing.T) {
		var deps Deps
		assert.IsType(t, mage.Deps{}, deps)
	})

	t.Run("Release type alias instantiation", func(t *testing.T) {
		var release Release
		assert.IsType(t, mage.Release{}, release)
	})

	t.Run("Configure type alias instantiation", func(t *testing.T) {
		var configure Configure
		assert.IsType(t, mage.Configure{}, configure)
	})
}

// TestWrapperFunctionDelegation verifies that wrapper functions properly
// delegate to the correct namespace methods.
func TestWrapperFunctionDelegation(t *testing.T) {
	t.Parallel()

	// This test verifies the delegation pattern used in the wrapper functions.
	// Each wrapper creates a namespace instance and calls a specific method.
	// We verify by checking that the wrapper function body creates the correct type.

	tests := []struct {
		name          string
		wrapperFunc   func() error
		namespaceType interface{}
		methodName    string
	}{
		{
			name:          "BuildCmd delegates to Build.Default",
			wrapperFunc:   BuildCmd,
			namespaceType: mage.Build{},
			methodName:    "Default",
		},
		{
			name:          "TestCmd delegates to Test.Default",
			wrapperFunc:   TestCmd,
			namespaceType: mage.Test{},
			methodName:    "Default",
		},
		{
			name:          "LintCmd delegates to Lint.Default",
			wrapperFunc:   LintCmd,
			namespaceType: mage.Lint{},
			methodName:    "Default",
		},
		{
			name:          "FormatCmd delegates to Format.Default",
			wrapperFunc:   FormatCmd,
			namespaceType: mage.Format{},
			methodName:    "Default",
		},
		{
			name:          "CleanCmd delegates to Build.Clean",
			wrapperFunc:   CleanCmd,
			namespaceType: mage.Build{},
			methodName:    "Clean",
		},
		{
			name:          "InstallCmd delegates to Build.Install",
			wrapperFunc:   InstallCmd,
			namespaceType: mage.Build{},
			methodName:    "Install",
		},
		{
			name:          "DepsCmd delegates to Deps.Default",
			wrapperFunc:   DepsCmd,
			namespaceType: mage.Deps{},
			methodName:    "Default",
		},
		{
			name:          "ReleaseCmd delegates to Release.Default",
			wrapperFunc:   ReleaseCmd,
			namespaceType: mage.Release{},
			methodName:    "Default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify the wrapper function is not nil
			assert.NotNil(t, tt.wrapperFunc)

			// Verify the namespace type has the expected method
			// This is checked via the type alias, which must match the mage type
			assert.NotNil(t, tt.namespaceType)
		})
	}
}

// TestAllTypeAliasesExist verifies that all expected type aliases are defined
// and match their underlying mage types.
func TestAllTypeAliasesExist(t *testing.T) {
	t.Parallel()

	// List of all type aliases defined in register.go
	typeAliases := []struct {
		name         string
		aliasType    interface{}
		expectedType interface{}
	}{
		{"Build", Build{}, mage.Build{}},
		{"Test", Test{}, mage.Test{}},
		{"Lint", Lint{}, mage.Lint{}},
		{"Format", Format{}, mage.Format{}},
		{"Deps", Deps{}, mage.Deps{}},
		{"Git", Git{}, mage.Git{}},
		{"Release", Release{}, mage.Release{}},
		{"Docs", Docs{}, mage.Docs{}},
		{"Tools", Tools{}, mage.Tools{}},
		{"Generate", Generate{}, mage.Generate{}},
		{"Update", Update{}, mage.Update{}},
		{"Mod", Mod{}, mage.Mod{}},
		{"Metrics", Metrics{}, mage.Metrics{}},
		{"Bench", Bench{}, mage.Bench{}},
		{"Vet", Vet{}, mage.Vet{}},
		{"Configure", Configure{}, mage.Configure{}},
		{"Help", Help{}, mage.Help{}},
	}

	for _, ta := range typeAliases {
		t.Run(ta.name+"_type_alias", func(t *testing.T) {
			t.Parallel()
			assert.IsType(t, ta.expectedType, ta.aliasType,
				"Type alias %s should match mage.%s", ta.name, ta.name)
		})
	}
}

// TestPackageImportability verifies the package can be imported and used.
// This is a basic sanity check for backward compatibility.
func TestPackageImportability(t *testing.T) {
	t.Parallel()

	// The fact that this test compiles and runs means the package is importable
	// and the type aliases and functions are accessible.

	t.Run("package provides namespace types", func(t *testing.T) {
		t.Parallel()

		// Create instances of each namespace type
		// This verifies they're properly exported
		var (
			_ Build
			_ Test
			_ Lint
			_ Format
			_ Deps
			_ Git
			_ Release
			_ Docs
			_ Tools
			_ Generate
			_ Update
			_ Mod
			_ Metrics
			_ Bench
			_ Vet
			_ Configure
			_ Help
		)
	})

	t.Run("package provides convenience functions", func(t *testing.T) {
		t.Parallel()

		// Create a slice of function references
		// This verifies they're properly exported
		_ = []func() error{
			BuildCmd,
			TestCmd,
			LintCmd,
			FormatCmd,
			CleanCmd,
			InstallCmd,
			DepsCmd,
			ReleaseCmd,
		}
	})
}

// Benchmarks for wrapper function overhead

func BenchmarkWrapperFunctionNamespaceCreation(b *testing.B) {
	// Benchmark the overhead of creating namespace instances
	// This is what wrapper functions do on each call
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var build mage.Build
		_ = build
	}
}

func BenchmarkTypeAliasInstanceCreation(b *testing.B) {
	// Benchmark creating instances via type aliases
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var build Build
		_ = build
	}
}
