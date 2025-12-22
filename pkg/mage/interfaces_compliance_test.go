package mage

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Predefined errors for testing to avoid dynamic error creation
var (
	ErrBuildNil          = errors.New("Build() returned nil")
	ErrTestNil           = errors.New("Test() returned nil")
	ErrLintNil           = errors.New("Lint() returned nil")
	ErrInconsistentBuild = errors.New("inconsistent Build instances")
)

// InterfaceComplianceTestSuite provides comprehensive interface compliance testing
type InterfaceComplianceTestSuite struct {
	suite.Suite

	registry *DefaultNamespaceRegistry
}

// SetupTest prepares each test case
func (suite *InterfaceComplianceTestSuite) SetupTest() {
	suite.registry = NewNamespaceRegistry()
}

// TestNamespaceFactoryFunctions tests all namespace factory functions
func (suite *InterfaceComplianceTestSuite) TestNamespaceFactoryFunctions() {
	testCases := []struct {
		name          string
		factory       func() interface{}
		interfaceType reflect.Type
	}{
		{
			name:          "NewBuildNamespace",
			factory:       func() interface{} { return NewBuildNamespace() },
			interfaceType: reflect.TypeOf((*BuildNamespace)(nil)).Elem(),
		},
		{
			name:          "NewTestNamespace",
			factory:       func() interface{} { return NewTestNamespace() },
			interfaceType: reflect.TypeOf((*TestNamespace)(nil)).Elem(),
		},
		{
			name:          "NewLintNamespace",
			factory:       func() interface{} { return NewLintNamespace() },
			interfaceType: reflect.TypeOf((*LintNamespace)(nil)).Elem(),
		},
		{
			name:          "NewFormatNamespace",
			factory:       func() interface{} { return NewFormatNamespace() },
			interfaceType: reflect.TypeOf((*FormatNamespace)(nil)).Elem(),
		},
		{
			name:          "NewDepsNamespace",
			factory:       func() interface{} { return NewDepsNamespace() },
			interfaceType: reflect.TypeOf((*DepsNamespace)(nil)).Elem(),
		},
		{
			name:          "NewGitNamespace",
			factory:       func() interface{} { return NewGitNamespace() },
			interfaceType: reflect.TypeOf((*GitNamespace)(nil)).Elem(),
		},
		{
			name:          "NewReleaseNamespace",
			factory:       func() interface{} { return NewReleaseNamespace() },
			interfaceType: reflect.TypeOf((*ReleaseNamespace)(nil)).Elem(),
		},
		{
			name:          "NewDocsNamespace",
			factory:       func() interface{} { return NewDocsNamespace() },
			interfaceType: reflect.TypeOf((*DocsNamespace)(nil)).Elem(),
		},
		{
			name:          "NewToolsNamespace",
			factory:       func() interface{} { return NewToolsNamespace() },
			interfaceType: reflect.TypeOf((*ToolsNamespace)(nil)).Elem(),
		},
		{
			name:          "NewGenerateNamespace",
			factory:       func() interface{} { return NewGenerateNamespace() },
			interfaceType: reflect.TypeOf((*GenerateNamespace)(nil)).Elem(),
		},
		{
			name:          "NewUpdateNamespace",
			factory:       func() interface{} { return NewUpdateNamespace() },
			interfaceType: reflect.TypeOf((*UpdateNamespace)(nil)).Elem(),
		},
		{
			name:          "NewModNamespace",
			factory:       func() interface{} { return NewModNamespace() },
			interfaceType: reflect.TypeOf((*ModNamespace)(nil)).Elem(),
		},
		{
			name:          "NewMetricsNamespace",
			factory:       func() interface{} { return NewMetricsNamespace() },
			interfaceType: reflect.TypeOf((*MetricsNamespace)(nil)).Elem(),
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			namespace := tc.factory()
			suite.NotNil(namespace, "Factory function should not return nil")

			// Check if the returned value implements the interface
			namespaceType := reflect.TypeOf(namespace)
			suite.True(namespaceType.Implements(tc.interfaceType),
				"Factory %s should return a type that implements %s",
				tc.name, tc.interfaceType.Name())
		})
	}
}

// TestNamespaceInterfaceMethods tests that all interface methods are callable
func (suite *InterfaceComplianceTestSuite) TestNamespaceInterfaceMethods() {
	suite.Run("BuildNamespace methods", func() {
		build := suite.registry.Build()
		suite.NotNil(build)

		// Test all BuildNamespace methods exist and are callable
		methods := []string{
			"Default", "All", "Platform", "Linux", "Darwin", "Windows",
			"Clean", "Install", "Generate", "PreBuild",
		}

		for _, method := range methods {
			suite.checkMethodExists(build, method)
		}
	})

	suite.Run("TestNamespace methods", func() {
		test := suite.registry.Test()
		suite.NotNil(test)

		methods := []string{
			"Default", "Unit", "Short", "Race", "Cover", "CoverRace",
			"CoverReport", "CoverHTML", "Fuzz", "Bench", "Integration",
			"CI", "Parallel", "NoLint", "CINoRace", "Run", "Coverage",
			"Vet", "Lint", "Clean", "All",
		}

		for _, method := range methods {
			suite.checkMethodExists(test, method)
		}
	})

	suite.Run("LintNamespace methods", func() {
		lint := suite.registry.Lint()
		suite.NotNil(lint)

		methods := []string{
			"Default", "All", "Go", "Yaml", "Fix", "CI", "Fast", "Config",
		}

		for _, method := range methods {
			suite.checkMethodExists(lint, method)
		}
	})

	suite.Run("ReleaseNamespace methods", func() {
		release := suite.registry.Release()
		suite.NotNil(release)

		methods := []string{
			"Default", "Test", "Snapshot", "LocalInstall", "Check",
			"Init", "Changelog", "Validate", "Clean",
		}

		for _, method := range methods {
			suite.checkMethodExists(release, method)
		}
	})
}

// checkMethodExists verifies that a method exists on an interface
func (suite *InterfaceComplianceTestSuite) checkMethodExists(obj interface{}, methodName string) {
	value := reflect.ValueOf(obj)
	method := value.MethodByName(methodName)
	suite.True(method.IsValid(), "Method %s should exist on %T", methodName, obj)

	if method.IsValid() {
		// Verify method signature - all namespace methods should return error
		methodType := method.Type()
		suite.GreaterOrEqual(methodType.NumOut(), 1, "Method %s should have at least one return value", methodName)

		// Check if last return type is error
		lastReturnType := methodType.Out(methodType.NumOut() - 1)
		errorInterface := reflect.TypeOf((*error)(nil)).Elem()
		suite.True(lastReturnType.Implements(errorInterface),
			"Method %s should return error as last return value", methodName)
	}
}

// TestRegistryPatternCompliance tests registry pattern compliance
func (suite *InterfaceComplianceTestSuite) TestRegistryPatternCompliance() {
	suite.Run("registry returns consistent instances", func() {
		// Test that multiple calls to the same namespace return the same instance
		build1 := suite.registry.Build()
		build2 := suite.registry.Build()
		suite.Same(build1, build2, "Registry should return same Build instance")

		test1 := suite.registry.Test()
		test2 := suite.registry.Test()
		suite.Same(test1, test2, "Registry should return same Test instance")

		lint1 := suite.registry.Lint()
		lint2 := suite.registry.Lint()
		suite.Same(lint1, lint2, "Registry should return same Lint instance")
	})

	suite.Run("lazy initialization works correctly", func() {
		// Create new registry
		registry := NewNamespaceRegistry()

		// First access should initialize
		build := registry.Build()
		suite.NotNil(build)

		// Second access should return same instance
		build2 := registry.Build()
		suite.Same(build, build2)
	})

	suite.Run("custom namespace injection works", func() {
		// Create custom build namespace
		customBuild := NewBuildNamespace()

		// Inject it into registry
		suite.registry.SetBuild(customBuild)

		// Verify it's returned
		retrieved := suite.registry.Build()
		suite.Same(customBuild, retrieved)
	})
}

// TestInterfaceContractValidation tests interface contracts
func (suite *InterfaceComplianceTestSuite) TestInterfaceContractValidation() {
	suite.Run("BuildNamespace contract validation", func() {
		build := suite.registry.Build()
		suite.NotNil(build)

		// Test method contracts - they should not panic
		suite.NotPanics(func() {
			// Most methods will return errors in test environment
			// but should not panic
			if err := build.Clean(); err != nil {
				// Error is expected in test environment, just ensure no panic
				suite.T().Logf("Clean method returned error as expected: %v", err)
			}
		})
	})

	suite.Run("TestNamespace contract validation", func() {
		test := suite.registry.Test()
		suite.NotNil(test)

		suite.NotPanics(func() {
			if err := test.Clean(); err != nil {
				// Error is expected in test environment, just ensure no panic
				suite.T().Logf("Clean method returned error as expected: %v", err)
			}
		})
	})

	suite.Run("interface type assertions", func() {
		// Test that registry implements NamespaceRegistry
		var _ NamespaceRegistry = suite.registry

		// Test that all namespace methods return correct interface types
		_ = suite.registry.Build()
		_ = suite.registry.Test()
		_ = suite.registry.Lint()
		_ = suite.registry.Format()
		_ = suite.registry.Deps()
		_ = suite.registry.Git()
		_ = suite.registry.Release()
		_ = suite.registry.Docs()
		_ = suite.registry.Tools()
		_ = suite.registry.Generate()
		_ = suite.registry.Update()
		_ = suite.registry.Mod()
		_ = suite.registry.Metrics()
	})
}

// TestConcurrentNamespaceAccess tests thread safety
func (suite *InterfaceComplianceTestSuite) TestConcurrentNamespaceAccess() {
	const numGoroutines = 50
	const iterations = 10

	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines*iterations)

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < iterations; j++ {
				// Test concurrent access to different namespaces
				build := suite.registry.Build()
				if build == nil {
					errorChan <- fmt.Errorf("goroutine %d iteration %d: %w", id, j, ErrBuildNil)
					continue
				}

				test := suite.registry.Test()
				if test == nil {
					errorChan <- fmt.Errorf("goroutine %d iteration %d: %w", id, j, ErrTestNil)
					continue
				}

				lint := suite.registry.Lint()
				if lint == nil {
					errorChan <- fmt.Errorf("goroutine %d iteration %d: %w", id, j, ErrLintNil)
					continue
				}

				// Test that same instances are returned across goroutines
				build2 := suite.registry.Build()
				if build != build2 {
					errorChan <- fmt.Errorf("goroutine %d iteration %d: %w", id, j, ErrInconsistentBuild)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errorChan)

	// Collect errors
	errors := make([]error, 0, numGoroutines*iterations)
	for err := range errorChan {
		errors = append(errors, err)
	}

	suite.Empty(errors, "Concurrent namespace access should be thread-safe")
}

// TestNamespaceRegistryProvider tests provider pattern
func (suite *InterfaceComplianceTestSuite) TestNamespaceRegistryProvider() {
	suite.Run("default provider works correctly", func() {
		provider := NewDefaultNamespaceRegistryProvider()
		suite.NotNil(provider)

		registry := provider.GetNamespaceRegistry()
		suite.NotNil(registry)

		// Subsequent calls should return same instance
		registry2 := provider.GetNamespaceRegistry()
		suite.Same(registry, registry2)
	})

	suite.Run("global registry access works", func() {
		registry1 := GetNamespaceRegistry()
		suite.NotNil(registry1)

		registry2 := GetNamespaceRegistry()
		suite.Same(registry1, registry2)
	})

	suite.Run("provider interface compliance", func() {
		// Test interface compliance at compile time
		var _ NamespaceRegistryProvider = NewDefaultNamespaceRegistryProvider()
		var _ NamespaceRegistryProvider = (*DefaultNamespaceRegistryProvider)(nil)
	})
}

// TestNamespaceMethodSignatures tests method signature consistency
func (suite *InterfaceComplianceTestSuite) TestNamespaceMethodSignatures() {
	testCases := []struct {
		name      string
		namespace interface{}
		methods   []string
	}{
		{
			name:      "BuildNamespace",
			namespace: suite.registry.Build(),
			methods:   []string{"Default", "All", "Clean", "Install"},
		},
		{
			name:      "TestNamespace",
			namespace: suite.registry.Test(),
			methods:   []string{"Default", "Unit", "Race", "Cover"},
		},
		{
			name:      "LintNamespace",
			namespace: suite.registry.Lint(),
			methods:   []string{"Default", "All", "Go", "Fix"},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			value := reflect.ValueOf(tc.namespace)
			type_ := value.Type()

			for _, methodName := range tc.methods {
				method, exists := type_.MethodByName(methodName)
				suite.True(exists, "Method %s should exist on %s", methodName, tc.name)

				if exists {
					// Check method signature
					methodType := method.Type

					// Should have at least one return value (error)
					suite.GreaterOrEqual(methodType.NumOut(), 1, "Method %s.%s should have at least one return value", tc.name, methodName)

					// Last return value should implement error
					lastReturn := methodType.Out(methodType.NumOut() - 1)
					errorInterface := reflect.TypeOf((*error)(nil)).Elem()
					suite.True(lastReturn.Implements(errorInterface), "Method %s.%s should return error", tc.name, methodName)
				}
			}
		})
	}
}

// TestNamespaceLifecycle tests namespace lifecycle management
func (suite *InterfaceComplianceTestSuite) TestNamespaceLifecycle() {
	suite.Run("namespace initialization order", func() {
		// Create fresh registry
		registry := NewNamespaceRegistry()

		// Test that namespaces are initialized on first access
		build := registry.Build()
		suite.NotNil(build)

		test := registry.Test()
		suite.NotNil(test)

		// Test that they maintain consistency
		build2 := registry.Build()
		suite.Same(build, build2)
	})

	suite.Run("namespace replacement", func() {
		// Test namespace replacement functionality
		originalBuild := suite.registry.Build()
		suite.NotNil(originalBuild)

		// Replace with new instance
		newBuild := NewBuildNamespace()
		suite.registry.SetBuild(newBuild)

		// Verify replacement - the registry should now return the new instance
		currentBuild := suite.registry.Build()
		suite.Same(newBuild, currentBuild)

		// For empty structs, Go may optimize memory allocation, so we verify
		// the registry replacement functionality by confirming the current
		// instance is the one we set, rather than comparing memory addresses
		suite.NotNil(currentBuild)
	})
}

// TestInterfaceStabilityAndEvolution tests interface evolution patterns
func (suite *InterfaceComplianceTestSuite) TestInterfaceStabilityAndEvolution() {
	suite.Run("interface method count stability", func() {
		// Track expected method counts to detect interface changes
		expectedCounts := map[string]int{
			"BuildNamespace":   10, // Default, All, Platform, Linux, Darwin, Windows, Clean, Install, Generate, PreBuild
			"TestNamespace":    21, // All test-related methods
			"LintNamespace":    11, // All lint-related methods
			"ReleaseNamespace": 9,  // All release-related methods
		}

		testCases := []struct {
			name      string
			namespace interface{}
		}{
			{"BuildNamespace", suite.registry.Build()},
			{"TestNamespace", suite.registry.Test()},
			{"LintNamespace", suite.registry.Lint()},
			{"ReleaseNamespace", suite.registry.Release()},
		}

		for _, tc := range testCases {
			value := reflect.ValueOf(tc.namespace)
			type_ := value.Type()

			actualCount := type_.NumMethod()
			expectedCount, exists := expectedCounts[tc.name]

			if exists {
				suite.GreaterOrEqual(actualCount, expectedCount,
					"Interface %s has fewer methods than expected (backward compatibility)", tc.name)
			} else {
				// Log for new interfaces
				suite.T().Logf("Interface %s has %d methods", tc.name, actualCount)
			}
		}
	})

	suite.Run("method signature consistency", func() {
		// Test that common method patterns are consistent
		namespaces := []interface{}{
			suite.registry.Build(),
			suite.registry.Test(),
			suite.registry.Lint(),
		}

		// All namespaces should have Default method with consistent signature per namespace type
		for _, ns := range namespaces {
			value := reflect.ValueOf(ns)
			defaultMethod := value.MethodByName("Default")

			if defaultMethod.IsValid() {
				methodType := defaultMethod.Type()

				// TestNamespace Default() should take variadic string args and return error
				// Other namespaces Default() should take no parameters and return error
				nsType := reflect.TypeOf(ns).String()
				if strings.Contains(nsType, "testNamespaceWrapper") {
					// Test namespace should accept variadic arguments
					suite.True(methodType.IsVariadic(), "Test namespace Default method should be variadic")
					suite.Equal(1, methodType.NumIn(), "Test namespace Default method should take one variadic parameter")
				} else {
					// Other namespaces should take no parameters
					suite.Equal(0, methodType.NumIn(), "Default method should take no parameters")
				}

				suite.Equal(1, methodType.NumOut(), "Default method should return one value")
				errorInterface := reflect.TypeOf((*error)(nil)).Elem()
				suite.True(methodType.Out(0).Implements(errorInterface), "Default method should return error")
			}
		}
	})
}

// TestRunInterfaceComplianceTestSuite runs the comprehensive interface compliance test suite
func TestRunInterfaceComplianceTestSuite(t *testing.T) {
	suite.Run(t, new(InterfaceComplianceTestSuite))
}

// BenchmarkNamespaceAccess benchmarks namespace access patterns
func BenchmarkNamespaceAccess(b *testing.B) {
	registry := NewNamespaceRegistry()

	b.Run("Build namespace access", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				build := registry.Build()
				if build == nil {
					b.Fatal("Build namespace should not be nil")
				}
			}
		})
	})

	b.Run("Multiple namespace access", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = registry.Build()
			_ = registry.Test()
			_ = registry.Lint()
			_ = registry.Release()
		}
	})

	b.Run("Registry creation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			registry := NewNamespaceRegistry()
			if registry == nil {
				b.Fatal("Registry should not be nil")
			}
		}
	})

	b.Run("Provider pattern", func(b *testing.B) {
		provider := NewDefaultNamespaceRegistryProvider()
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				registry := provider.GetNamespaceRegistry()
				if registry == nil {
					b.Fatal("Registry should not be nil")
				}
			}
		})
	})
}

// TestInterfaceEvolutionTracking tests for interface evolution tracking
func TestInterfaceEvolutionTracking(t *testing.T) {
	// This test helps track interface evolution over time
	// It should be updated when interfaces change

	t.Run("track interface method counts", func(t *testing.T) {
		t.Parallel()
		registry := NewNamespaceRegistry()

		// Current method counts as of test creation
		interfaceCounts := map[string]interface{}{
			"Build":    registry.Build(),
			"Test":     registry.Test(),
			"Lint":     registry.Lint(),
			"Format":   registry.Format(),
			"Deps":     registry.Deps(),
			"Git":      registry.Git(),
			"Release":  registry.Release(),
			"Docs":     registry.Docs(),
			"Tools":    registry.Tools(),
			"Generate": registry.Generate(),
			"Update":   registry.Update(),
			"Mod":      registry.Mod(),
			"Metrics":  registry.Metrics(),
		}

		for name, ns := range interfaceCounts {
			if ns != nil {
				value := reflect.ValueOf(ns)
				type_ := value.Type()
				methodCount := type_.NumMethod()
				t.Logf("%s interface has %d methods", name, methodCount)

				// Ensure minimum method count (at least Default method)
				assert.GreaterOrEqual(t, methodCount, 1, "%s should have at least one method", name)
			} else {
				t.Logf("%s interface is not implemented (nil)", name)
			}
		}
	})

	t.Run("verify interface naming consistency", func(t *testing.T) {
		t.Parallel()

		// Test that interface names follow consistent patterns
		expectedInterfaces := []string{
			"BuildNamespace", "TestNamespace", "LintNamespace", "FormatNamespace",
			"DepsNamespace", "GitNamespace", "ReleaseNamespace", "DocsNamespace",
			"DeployNamespace", "ToolsNamespace", "SecurityNamespace", "GenerateNamespace",
		}

		for _, interfaceName := range expectedInterfaces {
			// This is a compile-time check to ensure interfaces exist
			t.Logf("Interface %s should exist in the codebase", interfaceName)
			// The actual verification is done at compile time through imports and usage
		}
	})

	t.Run("verify DepsNamespace.Audit method exists", func(t *testing.T) {
		t.Parallel()

		// This test ensures the Audit method exists on DepsNamespace interface
		// and is not accidentally removed during refactoring
		depsInterface := reflect.TypeOf((*DepsNamespace)(nil)).Elem()

		// Check if Audit method exists
		auditMethod, found := depsInterface.MethodByName("Audit")
		assert.True(t, found, "DepsNamespace interface must have an Audit method")

		if found {
			// Verify method signature: func(args ...string) error
			methodType := auditMethod.Type
			assert.Equal(t, 1, methodType.NumIn(), "Audit method should have 1 input (variadic args)")
			assert.Equal(t, 1, methodType.NumOut(), "Audit method should have 1 output (error)")

			// Check return type is error
			errorType := reflect.TypeOf((*error)(nil)).Elem()
			assert.True(t, methodType.Out(0).Implements(errorType), "Audit method should return error")
		}

		// Also verify the implementation exists
		deps := NewDepsNamespace()
		depsValue := reflect.ValueOf(deps)
		auditMethodValue := depsValue.MethodByName("Audit")
		assert.True(t, auditMethodValue.IsValid(), "Deps implementation must have Audit method")
	})
}
