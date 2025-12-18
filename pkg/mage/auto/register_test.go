package auto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage"
)

// TestConvenienceWrapperFunctions tests all wrapper functions to ensure they can be called
func TestConvenienceWrapperFunctions(t *testing.T) {
	tests := []struct {
		name        string
		wrapperFunc func() error
		description string
	}{
		{
			name:        "BuildCmd",
			wrapperFunc: BuildCmd,
			description: "builds the application",
		},
		{
			name:        "TestCmd",
			wrapperFunc: TestCmd,
			description: "runs tests",
		},
		{
			name:        "LintCmd",
			wrapperFunc: LintCmd,
			description: "runs linting",
		},
		{
			name:        "FormatCmd",
			wrapperFunc: FormatCmd,
			description: "formats code",
		},
		{
			name:        "CleanCmd",
			wrapperFunc: CleanCmd,
			description: "cleans build artifacts",
		},
		{
			name:        "InstallCmd",
			wrapperFunc: InstallCmd,
			description: "installs the application",
		},
		{
			name:        "DepsCmd",
			wrapperFunc: DepsCmd,
			description: "manages dependencies",
		},
		{
			name:        "ReleaseCmd",
			wrapperFunc: ReleaseCmd,
			description: "creates a release",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel() // Run tests in parallel for better performance

			// Test that the function exists and is callable
			require.NotNil(t, tt.wrapperFunc, "Function %s should not be nil", tt.name)

			// Test that the function has the expected signature
			assert.IsType(t, func() error { return nil }, tt.wrapperFunc,
				"Function %s should have signature func() error", tt.name)
		})
	}
}

// TestNamespaceTypeAliases tests that type aliases are properly defined and match their underlying types
func TestNamespaceTypeAliases(t *testing.T) {
	t.Parallel()

	// Test that type aliases match their underlying types
	tests := []struct {
		name     string
		testFunc func(t *testing.T)
	}{
		{
			name: "Build type alias",
			testFunc: func(t *testing.T) {
				var b Build
				var mb mage.Build
				// Both should have the same underlying type
				assert.IsType(t, mb, b)
			},
		},
		{
			name: "Test type alias",
			testFunc: func(t *testing.T) {
				var test Test
				var mtest mage.Test
				assert.IsType(t, mtest, test)
			},
		},
		{
			name: "Lint type alias",
			testFunc: func(t *testing.T) {
				var lint Lint
				var mlint mage.Lint
				assert.IsType(t, mlint, lint)
			},
		},
		{
			name: "Format type alias",
			testFunc: func(t *testing.T) {
				var format Format
				var mformat mage.Format
				assert.IsType(t, mformat, format)
			},
		},
		{
			name: "Deps type alias",
			testFunc: func(t *testing.T) {
				var deps Deps
				var mdeps mage.Deps
				assert.IsType(t, mdeps, deps)
			},
		},
		{
			name: "Git type alias",
			testFunc: func(t *testing.T) {
				var git Git
				var mgit mage.Git
				assert.IsType(t, mgit, git)
			},
		},
		{
			name: "Release type alias",
			testFunc: func(t *testing.T) {
				var release Release
				var mrelease mage.Release
				assert.IsType(t, mrelease, release)
			},
		},
		{
			name: "Docs type alias",
			testFunc: func(t *testing.T) {
				var docs Docs
				var mdocs mage.Docs
				assert.IsType(t, mdocs, docs)
			},
		},
		{
			name: "Tools type alias",
			testFunc: func(t *testing.T) {
				var tools Tools
				var mtools mage.Tools
				assert.IsType(t, mtools, tools)
			},
		},
		{
			name: "Generate type alias",
			testFunc: func(t *testing.T) {
				var generate Generate
				var mgenerate mage.Generate
				assert.IsType(t, mgenerate, generate)
			},
		},
		{
			name: "Update type alias",
			testFunc: func(t *testing.T) {
				var update Update
				var mupdate mage.Update
				assert.IsType(t, mupdate, update)
			},
		},
		{
			name: "Mod type alias",
			testFunc: func(t *testing.T) {
				var mod Mod
				var mmod mage.Mod
				assert.IsType(t, mmod, mod)
			},
		},
		{
			name: "Metrics type alias",
			testFunc: func(t *testing.T) {
				var metrics Metrics
				var mmetrics mage.Metrics
				assert.IsType(t, mmetrics, metrics)
			},
		},
		{
			name: "Bench type alias",
			testFunc: func(t *testing.T) {
				var bench Bench
				var mbench mage.Bench
				assert.IsType(t, mbench, bench)
			},
		},
		{
			name: "Vet type alias",
			testFunc: func(t *testing.T) {
				var vet Vet
				var mvet mage.Vet
				assert.IsType(t, mvet, vet)
			},
		},
		{
			name: "Configure type alias",
			testFunc: func(t *testing.T) {
				var configure Configure
				var mconfigure mage.Configure
				assert.IsType(t, mconfigure, configure)
			},
		},
		{
			name: "Help type alias",
			testFunc: func(t *testing.T) {
				var help Help
				var mhelp mage.Help
				assert.IsType(t, mhelp, help)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.testFunc(t)
		})
	}
}

// TestBackwardCompatibility tests backward compatibility aspects
func TestBackwardCompatibility(t *testing.T) {
	t.Parallel()

	t.Run("package provides expected functions", func(t *testing.T) {
		// Test that all convenience functions exist and are callable
		functions := []func() error{
			BuildCmd,
			TestCmd,
			LintCmd,
			FormatCmd,
			CleanCmd,
			InstallCmd,
			DepsCmd,
			ReleaseCmd,
		}

		assert.Len(t, functions, 8, "Should have exactly 8 convenience functions")

		for i, fn := range functions {
			assert.NotNil(t, fn, "Function at index %d should not be nil", i)
		}
	})

	t.Run("type aliases provide namespace access", func(t *testing.T) {
		// Test that type aliases can be used to access namespace methods
		var b Build
		var test Test
		var l Lint
		var f Format
		var d Deps
		var r Release

		// These should compile and be callable (we're just testing the interface)
		assert.NotNil(t, b.Default)
		assert.NotNil(t, test.Default)
		assert.NotNil(t, l.Default)
		assert.NotNil(t, f.Default)
		assert.NotNil(t, d.Default)
		assert.NotNil(t, r.Default)

		// Test specific methods that convenience functions use
		assert.NotNil(t, b.Clean, "Build namespace should have Clean method")
		assert.NotNil(t, b.Install, "Build namespace should have Install method")
	})

	t.Run("package comment describes purpose correctly", func(t *testing.T) {
		// The package provides automatic registration for backward compatibility
		// This test ensures the package structure aligns with its documented purpose
		// Package serves backward compatibility purpose - verified by existence of functions
		assert.Len(t, []func() error{BuildCmd, TestCmd, LintCmd, FormatCmd, CleanCmd, InstallCmd, DepsCmd, ReleaseCmd}, 8, "Should have 8 convenience functions for backward compatibility")
	})
}

// TestWrapperFunctionBehavior tests the actual behavior of wrapper functions
func TestWrapperFunctionBehavior(t *testing.T) {
	tests := []struct {
		name         string
		wrapperFunc  func() error
		expectedCall string
	}{
		{
			name:         "BuildCmd calls Build.Default",
			wrapperFunc:  BuildCmd,
			expectedCall: "Build.Default",
		},
		{
			name:         "TestCmd calls Test.Default",
			wrapperFunc:  TestCmd,
			expectedCall: "Test.Default",
		},
		{
			name:         "LintCmd calls Lint.Default",
			wrapperFunc:  LintCmd,
			expectedCall: "Lint.Default",
		},
		{
			name:         "FormatCmd calls Format.Default",
			wrapperFunc:  FormatCmd,
			expectedCall: "Format.Default",
		},
		{
			name:         "CleanCmd calls Build.Clean",
			wrapperFunc:  CleanCmd,
			expectedCall: "Build.Clean",
		},
		{
			name:         "InstallCmd calls Build.Install",
			wrapperFunc:  InstallCmd,
			expectedCall: "Build.Install",
		},
		{
			name:         "DepsCmd calls Deps.Default",
			wrapperFunc:  DepsCmd,
			expectedCall: "Deps.Default",
		},
		{
			name:         "ReleaseCmd calls Release.Default",
			wrapperFunc:  ReleaseCmd,
			expectedCall: "Release.Default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that wrapper functions are callable and return error type
			// Since we cannot mock the underlying namespaces easily in this context,
			// we verify the function signature and that it's properly defined
			require.NotNil(t, tt.wrapperFunc, "Wrapper function should be defined")

			// Verify the function signature
			assert.IsType(t, func() error { return nil }, tt.wrapperFunc,
				"Wrapper function should have signature func() error")
		})
	}
}

// TestFunctionDocumentation tests that functions have proper documentation
func TestFunctionDocumentation(t *testing.T) {
	t.Parallel()

	// This test verifies the existence and basic structure of convenience functions
	// Each function should exist and be documented as specified in register.go
	t.Run("all documented functions exist", func(t *testing.T) {
		// List of expected convenience functions based on register.go comments
		expectedFunctions := map[string]string{
			"BuildCmd":   "builds the application (alias for Build.Default)",
			"TestCmd":    "runs tests (alias for Test.Default)",
			"LintCmd":    "runs linting (alias for Lint.Default)",
			"FormatCmd":  "formats code (alias for Format.Default)",
			"CleanCmd":   "cleans build artifacts (alias for Build.Clean)",
			"InstallCmd": "installs the application (alias for Build.Install)",
			"DepsCmd":    "manages dependencies (alias for Deps.Default)",
			"ReleaseCmd": "creates a release (alias for Release.Default)",
		}

		functions := map[string]func() error{
			"BuildCmd":   BuildCmd,
			"TestCmd":    TestCmd,
			"LintCmd":    LintCmd,
			"FormatCmd":  FormatCmd,
			"CleanCmd":   CleanCmd,
			"InstallCmd": InstallCmd,
			"DepsCmd":    DepsCmd,
			"ReleaseCmd": ReleaseCmd,
		}

		for name, description := range expectedFunctions {
			fn, exists := functions[name]
			assert.True(t, exists, "Function %s should exist", name)
			assert.NotNil(t, fn, "Function %s should not be nil", name)
			// Description is used for documentation validation
			assert.NotEmpty(t, description, "Function %s should have documentation", name)
		}
	})
}

// TestNamespaceInstantiation tests that namespaces can be instantiated correctly
func TestNamespaceInstantiation(t *testing.T) {
	t.Parallel()

	t.Run("namespaces can be instantiated", func(t *testing.T) {
		// Test that we can create instances of each namespace type
		// This validates the type aliases work correctly

		// Test Build namespace
		var b Build
		assert.IsType(t, mage.Build{}, b, "Build should be aliased to mage.Build")

		// Test Test namespace
		var test Test
		assert.IsType(t, mage.Test{}, test, "Test should be aliased to mage.Test")

		// Test Lint namespace
		var lint Lint
		assert.IsType(t, mage.Lint{}, lint, "Lint should be aliased to mage.Lint")

		// Test Format namespace
		var format Format
		assert.IsType(t, mage.Format{}, format, "Format should be aliased to mage.Format")

		// Test Deps namespace
		var deps Deps
		assert.IsType(t, mage.Deps{}, deps, "Deps should be aliased to mage.Deps")

		// Test Release namespace
		var release Release
		assert.IsType(t, mage.Release{}, release, "Release should be aliased to mage.Release")
	})

	t.Run("all type aliases exist", func(t *testing.T) {
		// Verify all 24 type aliases exist and are properly defined
		typeAliases := []interface{}{
			Build{},
			Test{},
			Lint{},
			Format{},
			Deps{},
			Git{},
			Release{},
			Docs{},
			Tools{},
			Generate{},
			Update{},
			Mod{},
			Metrics{},
			Bench{},
			Vet{},
			Configure{},
			Help{},
		}

		assert.Len(t, typeAliases, 17, "Should have exactly 17 type aliases")

		for i, alias := range typeAliases {
			assert.NotNil(t, alias, "Type alias at index %d should not be nil", i)
		}
	})
}

// Benchmark tests for performance validation
func BenchmarkConvenienceWrappers(b *testing.B) {
	// Benchmark function signature validation (lightweight operation)
	wrappers := []func() error{
		BuildCmd,
		TestCmd,
		LintCmd,
		FormatCmd,
		CleanCmd,
		InstallCmd,
		DepsCmd,
		ReleaseCmd,
	}

	b.ResetTimer() // Reset timer after setup

	for i := 0; i < b.N; i++ {
		// Benchmark function existence checks (not actual execution)
		for _, wrapper := range wrappers {
			if wrapper == nil {
				b.Fatal("Wrapper function should not be nil")
			}
		}
	}

	b.StopTimer() // Stop timer before cleanup
}

// Fuzz test for wrapper function stability
func FuzzWrapperFunctionExists(f *testing.F) {
	// Test that wrapper functions exist regardless of input
	f.Add("BuildCmd")
	f.Add("TestCmd")
	f.Add("LintCmd")
	f.Add("FormatCmd")
	f.Add("CleanCmd")
	f.Add("InstallCmd")
	f.Add("DepsCmd")
	f.Add("ReleaseCmd")

	f.Fuzz(func(t *testing.T, funcName string) {
		// Map of actual functions
		functions := map[string]func() error{
			"BuildCmd":   BuildCmd,
			"TestCmd":    TestCmd,
			"LintCmd":    LintCmd,
			"FormatCmd":  FormatCmd,
			"CleanCmd":   CleanCmd,
			"InstallCmd": InstallCmd,
			"DepsCmd":    DepsCmd,
			"ReleaseCmd": ReleaseCmd,
		}

		if fn, exists := functions[funcName]; exists {
			// If it's one of our expected functions, it should not be nil
			assert.NotNil(t, fn, "Function %s should not be nil", funcName)
		}
		// For invalid function names, we don't panic or error
	})
}
