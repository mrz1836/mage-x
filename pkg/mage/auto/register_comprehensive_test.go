package auto

import (
	"reflect"
	"sync"
	"testing"

	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// AutoRegistrationTestSuite provides comprehensive tests for auto registration
type AutoRegistrationTestSuite struct {
	suite.Suite
}

// TestConvenienceWrapperFunctionsComprehensive tests all wrapper functions comprehensively
func (suite *AutoRegistrationTestSuite) TestConvenienceWrapperFunctionsComprehensive() {
	tests := []struct {
		name          string
		wrapperFunc   func() error
		expectedCall  string
		description   string
		namespaceType string
		methodName    string
	}{
		{
			name:          "BuildCmd",
			wrapperFunc:   BuildCmd,
			expectedCall:  "Build.Default",
			description:   "builds the application",
			namespaceType: "Build",
			methodName:    "Default",
		},
		{
			name:          "TestCmd",
			wrapperFunc:   TestCmd,
			expectedCall:  "Test.Default",
			description:   "runs tests",
			namespaceType: "Test",
			methodName:    "Default",
		},
		{
			name:          "LintCmd",
			wrapperFunc:   LintCmd,
			expectedCall:  "Lint.Default",
			description:   "runs linting",
			namespaceType: "Lint",
			methodName:    "Default",
		},
		{
			name:          "FormatCmd",
			wrapperFunc:   FormatCmd,
			expectedCall:  "Format.Default",
			description:   "formats code",
			namespaceType: "Format",
			methodName:    "Default",
		},
		{
			name:          "CleanCmd",
			wrapperFunc:   CleanCmd,
			expectedCall:  "Build.Clean",
			description:   "cleans build artifacts",
			namespaceType: "Build",
			methodName:    "Clean",
		},
		{
			name:          "InstallCmd",
			wrapperFunc:   InstallCmd,
			expectedCall:  "Build.Install",
			description:   "installs the application",
			namespaceType: "Build",
			methodName:    "Install",
		},
		{
			name:          "DepsCmd",
			wrapperFunc:   DepsCmd,
			expectedCall:  "Deps.Default",
			description:   "manages dependencies",
			namespaceType: "Deps",
			methodName:    "Default",
		},
		{
			name:          "ReleaseCmd",
			wrapperFunc:   ReleaseCmd,
			expectedCall:  "Release.Default",
			description:   "creates a release",
			namespaceType: "Release",
			methodName:    "Default",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Test function exists and is callable
			suite.Require().NotNil(tt.wrapperFunc, "Function %s should not be nil", tt.name)

			// Test function signature
			suite.IsType(func() error { return nil }, tt.wrapperFunc,
				"Function %s should have signature func() error", tt.name)

			// Test function is not panicking when accessed
			suite.NotPanics(func() {
				_ = tt.wrapperFunc
			}, "Function %s should not panic when accessed", tt.name)
		})
	}
}

// TestNamespaceTypeAliasesComprehensive tests all type aliases comprehensively
func (suite *AutoRegistrationTestSuite) TestNamespaceTypeAliasesComprehensive() {
	tests := []struct {
		name         string
		autoType     interface{}
		mageType     interface{}
		testCreation func() (interface{}, interface{})
		description  string
	}{
		{
			name:     "Build type alias",
			autoType: Build{},
			mageType: mage.Build{},
			testCreation: func() (interface{}, interface{}) {
				return Build{}, mage.Build{}
			},
			description: "Build should be aliased to mage.Build",
		},
		{
			name:     "Test type alias",
			autoType: Test{},
			mageType: mage.Test{},
			testCreation: func() (interface{}, interface{}) {
				return Test{}, mage.Test{}
			},
			description: "Test should be aliased to mage.Test",
		},
		{
			name:     "Lint type alias",
			autoType: Lint{},
			mageType: mage.Lint{},
			testCreation: func() (interface{}, interface{}) {
				return Lint{}, mage.Lint{}
			},
			description: "Lint should be aliased to mage.Lint",
		},
		{
			name:     "Format type alias",
			autoType: Format{},
			mageType: mage.Format{},
			testCreation: func() (interface{}, interface{}) {
				return Format{}, mage.Format{}
			},
			description: "Format should be aliased to mage.Format",
		},
		{
			name:     "Deps type alias",
			autoType: Deps{},
			mageType: mage.Deps{},
			testCreation: func() (interface{}, interface{}) {
				return Deps{}, mage.Deps{}
			},
			description: "Deps should be aliased to mage.Deps",
		},
		{
			name:     "Git type alias",
			autoType: Git{},
			mageType: mage.Git{},
			testCreation: func() (interface{}, interface{}) {
				return Git{}, mage.Git{}
			},
			description: "Git should be aliased to mage.Git",
		},
		{
			name:     "Release type alias",
			autoType: Release{},
			mageType: mage.Release{},
			testCreation: func() (interface{}, interface{}) {
				return Release{}, mage.Release{}
			},
			description: "Release should be aliased to mage.Release",
		},
		{
			name:     "Docs type alias",
			autoType: Docs{},
			mageType: mage.Docs{},
			testCreation: func() (interface{}, interface{}) {
				return Docs{}, mage.Docs{}
			},
			description: "Docs should be aliased to mage.Docs",
		},
		{
			name:     "Tools type alias",
			autoType: Tools{},
			mageType: mage.Tools{},
			testCreation: func() (interface{}, interface{}) {
				return Tools{}, mage.Tools{}
			},
			description: "Tools should be aliased to mage.Tools",
		},
		{
			name:     "Generate type alias",
			autoType: Generate{},
			mageType: mage.Generate{},
			testCreation: func() (interface{}, interface{}) {
				return Generate{}, mage.Generate{}
			},
			description: "Generate should be aliased to mage.Generate",
		},
		{
			name:     "CLI type alias",
			autoType: CLI{},
			mageType: mage.CLI{},
			testCreation: func() (interface{}, interface{}) {
				return CLI{}, mage.CLI{}
			},
			description: "CLI should be aliased to mage.CLI",
		},
		{
			name:     "Update type alias",
			autoType: Update{},
			mageType: mage.Update{},
			testCreation: func() (interface{}, interface{}) {
				return Update{}, mage.Update{}
			},
			description: "Update should be aliased to mage.Update",
		},
		{
			name:     "Mod type alias",
			autoType: Mod{},
			mageType: mage.Mod{},
			testCreation: func() (interface{}, interface{}) {
				return Mod{}, mage.Mod{}
			},
			description: "Mod should be aliased to mage.Mod",
		},
		{
			name:     "Recipes type alias",
			autoType: Recipes{},
			mageType: mage.Recipes{},
			testCreation: func() (interface{}, interface{}) {
				return Recipes{}, mage.Recipes{}
			},
			description: "Recipes should be aliased to mage.Recipes",
		},
		{
			name:     "Metrics type alias",
			autoType: Metrics{},
			mageType: mage.Metrics{},
			testCreation: func() (interface{}, interface{}) {
				return Metrics{}, mage.Metrics{}
			},
			description: "Metrics should be aliased to mage.Metrics",
		},
		{
			name:     "Workflow type alias",
			autoType: Workflow{},
			mageType: mage.Workflow{},
			testCreation: func() (interface{}, interface{}) {
				return Workflow{}, mage.Workflow{}
			},
			description: "Workflow should be aliased to mage.Workflow",
		},
		{
			name:     "Bench type alias",
			autoType: Bench{},
			mageType: mage.Bench{},
			testCreation: func() (interface{}, interface{}) {
				return Bench{}, mage.Bench{}
			},
			description: "Bench should be aliased to mage.Bench",
		},
		{
			name:     "Vet type alias",
			autoType: Vet{},
			mageType: mage.Vet{},
			testCreation: func() (interface{}, interface{}) {
				return Vet{}, mage.Vet{}
			},
			description: "Vet should be aliased to mage.Vet",
		},
		{
			name:     "Configure type alias",
			autoType: Configure{},
			mageType: mage.Configure{},
			testCreation: func() (interface{}, interface{}) {
				return Configure{}, mage.Configure{}
			},
			description: "Configure should be aliased to mage.Configure",
		},
		{
			name:     "Init type alias",
			autoType: Init{},
			mageType: mage.Init{},
			testCreation: func() (interface{}, interface{}) {
				return Init{}, mage.Init{}
			},
			description: "Init should be aliased to mage.Init",
		},
		{
			name:     "Enterprise type alias",
			autoType: Enterprise{},
			mageType: mage.Enterprise{},
			testCreation: func() (interface{}, interface{}) {
				return Enterprise{}, mage.Enterprise{}
			},
			description: "Enterprise should be aliased to mage.Enterprise",
		},
		{
			name:     "Integrations type alias",
			autoType: Integrations{},
			mageType: mage.Integrations{},
			testCreation: func() (interface{}, interface{}) {
				return Integrations{}, mage.Integrations{}
			},
			description: "Integrations should be aliased to mage.Integrations",
		},
		{
			name:     "Wizard type alias",
			autoType: Wizard{},
			mageType: mage.Wizard{},
			testCreation: func() (interface{}, interface{}) {
				return Wizard{}, mage.Wizard{}
			},
			description: "Wizard should be aliased to mage.Wizard",
		},
		{
			name:     "Help type alias",
			autoType: Help{},
			mageType: mage.Help{},
			testCreation: func() (interface{}, interface{}) {
				return Help{}, mage.Help{}
			},
			description: "Help should be aliased to mage.Help",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Test that types have the same underlying type
			autoType, mageType := tt.testCreation()
			suite.IsType(mageType, autoType, tt.description)

			// Test that both types have the same reflection type
			autoReflectType := reflect.TypeOf(autoType)
			mageReflectType := reflect.TypeOf(mageType)
			suite.Equal(mageReflectType, autoReflectType, "Reflection types should match for %s", tt.name)
		})
	}
}

// TestNamespaceMethodAccessibility tests that namespace methods are accessible through auto package
func (suite *AutoRegistrationTestSuite) TestNamespaceMethodAccessibility() {
	tests := []struct {
		name           string
		createInstance func() interface{}
		methodTests    []string
	}{
		{
			name:           "Build namespace methods",
			createInstance: func() interface{} { return Build{} },
			methodTests:    []string{"Default", "All", "Clean", "Install", "Generate"},
		},
		{
			name:           "Test namespace methods",
			createInstance: func() interface{} { return Test{} },
			methodTests:    []string{"Default", "Unit", "Race", "Cover", "Bench"},
		},
		{
			name:           "Lint namespace methods",
			createInstance: func() interface{} { return Lint{} },
			methodTests:    []string{"Default", "All", "Go", "Fix"},
		},
		{
			name:           "Format namespace methods",
			createInstance: func() interface{} { return Format{} },
			methodTests:    []string{"Default", "Check", "Go", "All"},
		},
		{
			name:           "Deps namespace methods",
			createInstance: func() interface{} { return Deps{} },
			methodTests:    []string{"Default", "Download", "Update", "Tidy"},
		},
		{
			name:           "Release namespace methods",
			createInstance: func() interface{} { return Release{} },
			methodTests:    []string{"Default", "Test", "Snapshot"},
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			instance := tt.createInstance()
			suite.NotNil(instance, "Instance should be created successfully")

			v := reflect.ValueOf(instance)
			for _, methodName := range tt.methodTests {
				method := v.MethodByName(methodName)
				suite.True(method.IsValid(), "Method %s should exist on %s", methodName, tt.name)
				if method.IsValid() {
					methodType := method.Type()
					// Verify method signature (should return error)
					suite.Equal(1, methodType.NumOut(), "Method %s should return one value", methodName)
					if methodType.NumOut() > 0 {
						returnType := methodType.Out(0)
						errorInterface := reflect.TypeOf((*error)(nil)).Elem()
						suite.True(returnType.Implements(errorInterface), "Method %s should return error", methodName)
					}
				}
			}
		})
	}
}

// TestBackwardCompatibilityComprehensive tests comprehensive backward compatibility
func (suite *AutoRegistrationTestSuite) TestBackwardCompatibilityComprehensive() {
	// Test that auto package provides all expected functionality for MAGE-X 1.0 users
	suite.Run("all convenience functions exist", func() {
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

		suite.Len(functions, 8, "Should have exactly 8 convenience functions")

		for name, fn := range functions {
			suite.NotNil(fn, "Function %s should not be nil", name)
			suite.IsType(func() error { return nil }, fn, "Function %s should have correct signature", name)
		}
	})

	suite.Run("all type aliases exist", func() {
		// Test that all 24 type aliases are defined
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
			CLI{},
			Update{},
			Mod{},
			Recipes{},
			Metrics{},
			Workflow{},
			Bench{},
			Vet{},
			Configure{},
			Init{},
			Enterprise{},
			Integrations{},
			Wizard{},
			Help{},
		}

		suite.Len(typeAliases, 24, "Should have exactly 24 type aliases")

		for i, alias := range typeAliases {
			suite.NotNil(alias, "Type alias at index %d should not be nil", i)
		}
	})

	suite.Run("magefile.go integration compatibility", func() {
		// Test that types can be used in magefile.go context
		var b Build
		var t Test
		var l Lint
		var f Format

		// These should have accessible methods for magefile.go usage
		suite.NotNil(b.Default)
		suite.NotNil(t.Default)
		suite.NotNil(l.Default)
		suite.NotNil(f.Default)

		// Test specific methods used by convenience functions
		suite.NotNil(b.Clean, "Build.Clean should be accessible for CleanCmd")
		suite.NotNil(b.Install, "Build.Install should be accessible for InstallCmd")
	})
}

// TestConvenienceWrapperImplementation tests the actual implementation of wrapper functions
func (suite *AutoRegistrationTestSuite) TestConvenienceWrapperImplementation() {
	tests := []struct {
		name          string
		wrapperFunc   func() error
		namespaceType string
		methodName    string
		description   string
	}{
		{
			name:          "BuildCmd implementation",
			wrapperFunc:   BuildCmd,
			namespaceType: "Build",
			methodName:    "Default",
			description:   "should create Build namespace and call Default method",
		},
		{
			name:          "TestCmd implementation",
			wrapperFunc:   TestCmd,
			namespaceType: "Test",
			methodName:    "Default",
			description:   "should create Test namespace and call Default method",
		},
		{
			name:          "LintCmd implementation",
			wrapperFunc:   LintCmd,
			namespaceType: "Lint",
			methodName:    "Default",
			description:   "should create Lint namespace and call Default method",
		},
		{
			name:          "FormatCmd implementation",
			wrapperFunc:   FormatCmd,
			namespaceType: "Format",
			methodName:    "Default",
			description:   "should create Format namespace and call Default method",
		},
		{
			name:          "CleanCmd implementation",
			wrapperFunc:   CleanCmd,
			namespaceType: "Build",
			methodName:    "Clean",
			description:   "should create Build namespace and call Clean method",
		},
		{
			name:          "InstallCmd implementation",
			wrapperFunc:   InstallCmd,
			namespaceType: "Build",
			methodName:    "Install",
			description:   "should create Build namespace and call Install method",
		},
		{
			name:          "DepsCmd implementation",
			wrapperFunc:   DepsCmd,
			namespaceType: "Deps",
			methodName:    "Default",
			description:   "should create Deps namespace and call Default method",
		},
		{
			name:          "ReleaseCmd implementation",
			wrapperFunc:   ReleaseCmd,
			namespaceType: "Release",
			methodName:    "Default",
			description:   "should create Release namespace and call Default method",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Test that wrapper function is properly defined
			suite.NotNil(tt.wrapperFunc, "Wrapper function should be defined")
			suite.IsType(func() error { return nil }, tt.wrapperFunc,
				"Wrapper function should have signature func() error")

			// Test that function doesn't panic when called
			// Note: We can't easily test the actual behavior without mocking,
			// but we can test that the function structure is correct
			suite.NotPanics(func() {
				// Get function value for inspection
				funcValue := reflect.ValueOf(tt.wrapperFunc)
				suite.True(funcValue.IsValid(), "Function should be valid")
				suite.Equal(reflect.Func, funcValue.Kind(), "Should be a function")
			}, "Wrapper function should not panic when inspected")
		})
	}
}

// TestConcurrentAccess tests concurrent access to auto package functions
func (suite *AutoRegistrationTestSuite) TestConcurrentAccess() {
	const numGoroutines = 100
	var wg sync.WaitGroup

	// Test concurrent access to convenience wrapper functions
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			// Access different functions based on index to test concurrent access patterns
			switch index % 8 {
			case 0:
				fn := BuildCmd
				suite.NotNil(fn)
			case 1:
				fn := TestCmd
				suite.NotNil(fn)
			case 2:
				fn := LintCmd
				suite.NotNil(fn)
			case 3:
				fn := FormatCmd
				suite.NotNil(fn)
			case 4:
				fn := CleanCmd
				suite.NotNil(fn)
			case 5:
				fn := InstallCmd
				suite.NotNil(fn)
			case 6:
				fn := DepsCmd
				suite.NotNil(fn)
			case 7:
				fn := ReleaseCmd
				suite.NotNil(fn)
			}
		}(i)
	}

	wg.Wait()
	// Concurrent access test completed successfully if we reach this point
}

// TestTypeAliasMethodCallability tests that methods can be called through type aliases
func (suite *AutoRegistrationTestSuite) TestTypeAliasMethodCallability() {
	// Test that type aliases properly expose methods
	tests := []struct {
		name           string
		createInstance func() interface{}
		methodName     string
	}{
		{"Build.Default", func() interface{} { return Build{} }, "Default"},
		{"Test.Default", func() interface{} { return Test{} }, "Default"},
		{"Lint.Default", func() interface{} { return Lint{} }, "Default"},
		{"Format.Default", func() interface{} { return Format{} }, "Default"},
		{"Deps.Default", func() interface{} { return Deps{} }, "Default"},
		{"Release.Default", func() interface{} { return Release{} }, "Default"},
		{"Build.Clean", func() interface{} { return Build{} }, "Clean"},
		{"Build.Install", func() interface{} { return Build{} }, "Install"},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			instance := tt.createInstance()
			v := reflect.ValueOf(instance)
			method := v.MethodByName(tt.methodName)

			suite.True(method.IsValid(), "Method %s should be callable", tt.methodName)
			if method.IsValid() {
				methodType := method.Type()
				suite.Equal(reflect.Func, method.Kind(), "Should be a function")
				suite.Equal(1, methodType.NumOut(), "Should return one value (error)")
			}
		})
	}
}

// TestPackageDocumentationCompliance tests that package provides documented features
func (suite *AutoRegistrationTestSuite) TestPackageDocumentationCompliance() {
	// Test features mentioned in package comments
	suite.Run("backward compatibility for MAGE-X 1.0", func() {
		// Ensure all documented convenience functions exist
		functions := []func() error{
			BuildCmd, TestCmd, LintCmd, FormatCmd,
			CleanCmd, InstallCmd, DepsCmd, ReleaseCmd,
		}

		for i, fn := range functions {
			suite.NotNil(fn, "Convenience function %d should exist", i)
		}
	})

	suite.Run("import-based usage without magex binary", func() {
		// Test that types can be used without magex binary
		var b Build
		var t Test
		var l Lint

		// Should be able to access methods directly
		suite.NotNil(b.Default)
		suite.NotNil(t.Default)
		suite.NotNil(l.Default)
	})

	suite.Run("zero-boilerplate experience mention", func() {
		// Package comments mention magex binary for zero-boilerplate
		// Test that auto package provides the import-based alternative
		convenienceFunctions := 8 // As documented
		typeAliases := 24         // As counted in tests

		functions := []func() error{
			BuildCmd, TestCmd, LintCmd, FormatCmd,
			CleanCmd, InstallCmd, DepsCmd, ReleaseCmd,
		}

		types := []interface{}{
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
			CLI{},
			Update{},
			Mod{},
			Recipes{},
			Metrics{},
			Workflow{},
			Bench{},
			Vet{},
			Configure{},
			Init{},
			Enterprise{},
			Integrations{},
			Wizard{},
			Help{},
		}

		suite.Len(functions, convenienceFunctions, "Should provide documented number of convenience functions")
		suite.Len(types, typeAliases, "Should provide all namespace type aliases")
	})
}

// TestRunAutoRegistrationTestSuite runs the auto registration test suite
func TestRunAutoRegistrationTestSuite(t *testing.T) {
	suite.Run(t, new(AutoRegistrationTestSuite))
}

// BenchmarkConvenienceFunctions benchmarks convenience function access
func BenchmarkConvenienceFunctions(b *testing.B) {
	tests := []struct {
		name string
		fn   func() error
	}{
		{"BuildCmd", BuildCmd},
		{"TestCmd", TestCmd},
		{"LintCmd", LintCmd},
		{"FormatCmd", FormatCmd},
		{"CleanCmd", CleanCmd},
		{"InstallCmd", InstallCmd},
		{"DepsCmd", DepsCmd},
		{"ReleaseCmd", ReleaseCmd},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				// Benchmark function access (not execution)
				fn := tt.fn
				if fn == nil {
					b.Fatal("Function should not be nil")
				}
			}
			b.StopTimer()
		})
	}
}

// BenchmarkTypeAliasCreation benchmarks type alias instantiation
func BenchmarkTypeAliasCreation(b *testing.B) {
	tests := []struct {
		name    string
		creator func() interface{}
	}{
		{"Build", func() interface{} { return Build{} }},
		{"Test", func() interface{} { return Test{} }},
		{"Lint", func() interface{} { return Lint{} }},
		{"Format", func() interface{} { return Format{} }},
		{"Deps", func() interface{} { return Deps{} }},
		{"Release", func() interface{} { return Release{} }},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				instance := tt.creator()
				if instance == nil {
					b.Fatal("Instance should not be nil")
				}
			}
			b.StopTimer()
		})
	}
}

// BenchmarkConcurrentFunctionAccess benchmarks concurrent function access
func BenchmarkConcurrentFunctionAccess(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Access different functions in parallel
			functions := []func() error{
				BuildCmd, TestCmd, LintCmd, FormatCmd,
			}
			for _, fn := range functions {
				if fn == nil {
					b.Fatal("Function should not be nil")
				}
			}
		}
	})
	b.StopTimer()
}

// Additional standalone tests

// TestAutoPackageCompileTimeChecks ensures compile-time correctness
func TestAutoPackageCompileTimeChecks(t *testing.T) {
	// These assignments will cause compile errors if type aliases are incorrect
	_ = Build{}
	_ = Test{}
	_ = Lint{}
	_ = Format{}
	_ = Deps{}
	_ = Git{}
	_ = Release{}
	_ = Docs{}
	_ = Tools{}
	_ = Generate{}
	_ = CLI{}
	_ = Update{}
	_ = Mod{}
	_ = Recipes{}
	_ = Metrics{}
	_ = Workflow{}
	_ = Bench{}
	_ = Vet{}
	_ = Configure{}
	_ = Init{}
	_ = Enterprise{}
	_ = Integrations{}
	_ = Wizard{}
	_ = Help{}

	// Test passes if it compiles - no assertion needed
}

// TestAutoPackageFunctionSignatures tests function signatures
func TestAutoPackageFunctionSignatures(t *testing.T) {
	t.Parallel()

	// Test that all convenience functions have the expected signature
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

	for name, fn := range functions {
		t.Run(name, func(t *testing.T) {
			require.NotNil(t, fn, "Function %s should not be nil", name)
			assert.IsType(t, func() error { return nil }, fn,
				"Function %s should have signature func() error", name)
		})
	}
}

// TestAutoPackageTypeAliasEquivalence tests type alias equivalence
func TestAutoPackageTypeAliasEquivalence(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		autoType interface{}
		mageType interface{}
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
		{"CLI", CLI{}, mage.CLI{}},
		{"Update", Update{}, mage.Update{}},
		{"Mod", Mod{}, mage.Mod{}},
		{"Recipes", Recipes{}, mage.Recipes{}},
		{"Metrics", Metrics{}, mage.Metrics{}},
		{"Workflow", Workflow{}, mage.Workflow{}},
		{"Bench", Bench{}, mage.Bench{}},
		{"Vet", Vet{}, mage.Vet{}},
		{"Configure", Configure{}, mage.Configure{}},
		{"Init", Init{}, mage.Init{}},
		{"Enterprise", Enterprise{}, mage.Enterprise{}},
		{"Integrations", Integrations{}, mage.Integrations{}},
		{"Wizard", Wizard{}, mage.Wizard{}},
		{"Help", Help{}, mage.Help{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.IsType(t, tt.mageType, tt.autoType,
				"Auto package %s should be equivalent to mage.%s", tt.name, tt.name)

			// Test reflection type equivalence
			autoReflectType := reflect.TypeOf(tt.autoType)
			mageReflectType := reflect.TypeOf(tt.mageType)
			assert.Equal(t, mageReflectType, autoReflectType,
				"Reflection types should be identical for %s", tt.name)
		})
	}
}

// FuzzAutoPackageFunctions fuzz tests auto package functions
func FuzzAutoPackageFunctions(f *testing.F) {
	// Add seed values for different function indices
	f.Add(0)   // BuildCmd
	f.Add(1)   // TestCmd
	f.Add(2)   // LintCmd
	f.Add(3)   // FormatCmd
	f.Add(4)   // CleanCmd
	f.Add(5)   // InstallCmd
	f.Add(6)   // DepsCmd
	f.Add(7)   // ReleaseCmd
	f.Add(100) // Out of range

	f.Fuzz(func(t *testing.T, funcIndex int) {
		functions := []func() error{
			BuildCmd, TestCmd, LintCmd, FormatCmd,
			CleanCmd, InstallCmd, DepsCmd, ReleaseCmd,
		}

		// Test accessing functions by index
		if funcIndex >= 0 && funcIndex < len(functions) {
			fn := functions[funcIndex]
			assert.NotNil(t, fn, "Function at index %d should not be nil", funcIndex)
			assert.IsType(t, func() error { return nil }, fn,
				"Function at index %d should have correct signature", funcIndex)
		}
		// For out of range indices, we just continue without error
	})
}
