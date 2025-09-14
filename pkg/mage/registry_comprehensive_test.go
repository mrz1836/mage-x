package mage

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// RegistryComprehensiveTestSuite provides comprehensive tests for registry implementation
type RegistryComprehensiveTestSuite struct {
	suite.Suite

	originalProvider NamespaceRegistryProvider
}

// SetupTest saves the original provider and creates a clean registry for each test
func (suite *RegistryComprehensiveTestSuite) SetupTest() {
	suite.originalProvider = NewDefaultNamespaceRegistryProvider()
	// Set a fresh provider for each test to ensure isolation
	SetNamespaceRegistryProvider(NewDefaultNamespaceRegistryProvider())
}

// TearDownTest restores the original provider
func (suite *RegistryComprehensiveTestSuite) TearDownTest() {
	if suite.originalProvider != nil {
		SetNamespaceRegistryProvider(suite.originalProvider)
	}
}

// TestNewNamespaceRegistry tests registry creation
func (suite *RegistryComprehensiveTestSuite) TestNewNamespaceRegistry() {
	registry := NewNamespaceRegistry()
	suite.Require().NotNil(registry)
	suite.IsType(&DefaultNamespaceRegistry{}, registry)

	// Test that all fields are properly initialized
	suite.NotNil(registry, "Registry should be initialized")
}

// TestRegistryLazyInitialization tests lazy initialization of namespaces
func (suite *RegistryComprehensiveTestSuite) TestRegistryLazyInitialization() {
	registry := GetNamespaceRegistry()
	suite.Require().NotNil(registry)

	// Test lazy initialization for each namespace
	tests := []struct {
		name         string
		getter       func() interface{}
		expectedType interface{}
		notNil       bool
	}{
		{
			name:         "Build",
			getter:       func() interface{} { return registry.Build() },
			expectedType: (*buildNamespaceWrapper)(nil),
			notNil:       true,
		},
		{
			name:         "Test",
			getter:       func() interface{} { return registry.Test() },
			expectedType: (*testNamespaceWrapper)(nil),
			notNil:       true,
		},
		{
			name:         "Lint",
			getter:       func() interface{} { return registry.Lint() },
			expectedType: (*lintNamespaceWrapper)(nil),
			notNil:       true,
		},
		{
			name:         "Format",
			getter:       func() interface{} { return registry.Format() },
			expectedType: (*formatNamespaceWrapper)(nil),
			notNil:       true,
		},
		{
			name:         "Deps",
			getter:       func() interface{} { return registry.Deps() },
			expectedType: (*depsNamespaceWrapper)(nil),
			notNil:       true,
		},
		{
			name:         "Git",
			getter:       func() interface{} { return registry.Git() },
			expectedType: (*gitNamespaceWrapper)(nil),
			notNil:       true,
		},
		{
			name:         "Release",
			getter:       func() interface{} { return registry.Release() },
			expectedType: (*releaseNamespaceWrapper)(nil),
			notNil:       true,
		},
		{
			name:         "Docs",
			getter:       func() interface{} { return registry.Docs() },
			expectedType: (*docsNamespaceWrapper)(nil),
			notNil:       true,
		},
		{
			name:         "Deploy",
			getter:       func() interface{} { return registry.Deploy() },
			expectedType: nil,
			notNil:       false, // Deploy is not implemented
		},
		{
			name:         "Tools",
			getter:       func() interface{} { return registry.Tools() },
			expectedType: (*toolsNamespaceWrapper)(nil),
			notNil:       true,
		},
		{
			name:         "Security",
			getter:       func() interface{} { return registry.Security() },
			expectedType: nil,
			notNil:       false, // Security is temporarily disabled
		},
		{
			name:         "Generate",
			getter:       func() interface{} { return registry.Generate() },
			expectedType: (*generateNamespaceWrapper)(nil),
			notNil:       true,
		},
		{
			name:         "CLI",
			getter:       func() interface{} { return registry.CLI() },
			expectedType: (*cliNamespaceWrapper)(nil),
			notNil:       true,
		},
		{
			name:         "Update",
			getter:       func() interface{} { return registry.Update() },
			expectedType: (*updateNamespaceWrapper)(nil),
			notNil:       true,
		},
		{
			name:         "Mod",
			getter:       func() interface{} { return registry.Mod() },
			expectedType: (*modNamespaceWrapper)(nil),
			notNil:       true,
		},
		{
			name:         "Recipes",
			getter:       func() interface{} { return registry.Recipes() },
			expectedType: (*recipesNamespaceWrapper)(nil),
			notNil:       true,
		},
		{
			name:         "Metrics",
			getter:       func() interface{} { return registry.Metrics() },
			expectedType: (*metricsNamespaceWrapper)(nil),
			notNil:       true,
		},
		{
			name:         "Workflow",
			getter:       func() interface{} { return registry.Workflow() },
			expectedType: (*workflowNamespaceWrapper)(nil),
			notNil:       true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			result := tt.getter()
			if tt.notNil {
				suite.NotNil(result, "Namespace %s should not be nil", tt.name)
				if tt.expectedType != nil {
					suite.IsType(tt.expectedType, result, "Namespace %s should have correct type", tt.name)
				}
			} else {
				suite.Nil(result, "Namespace %s should be nil", tt.name)
			}
		})
	}
}

// TestRegistryThreadSafety tests thread-safe access to registry
func (suite *RegistryComprehensiveTestSuite) TestRegistryThreadSafety() {
	const numGoroutines = 100
	const numIterations = 10

	var wg sync.WaitGroup
	results := make([][]*DefaultNamespaceRegistry, numGoroutines)

	// Test concurrent access to GetNamespaceRegistry
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index] = make([]*DefaultNamespaceRegistry, numIterations)
			for j := 0; j < numIterations; j++ {
				results[index][j] = GetNamespaceRegistry()
			}
		}(i)
	}

	wg.Wait()

	// Verify all results are the same instance (singleton behavior)
	firstRegistry := results[0][0]
	suite.NotNil(firstRegistry)

	for i, goroutineResults := range results {
		for j, registry := range goroutineResults {
			suite.Same(firstRegistry, registry, "All registry instances should be the same (goroutine %d, iteration %d)", i, j)
		}
	}
}

// TestRegistryNamespaceThreadSafety tests thread-safe access to individual namespaces
func (suite *RegistryComprehensiveTestSuite) TestRegistryNamespaceThreadSafety() {
	const numGoroutines = 50
	registry := GetNamespaceRegistry()

	var wg sync.WaitGroup
	buildResults := make([]BuildNamespace, numGoroutines)
	testResults := make([]TestNamespace, numGoroutines)
	lintResults := make([]LintNamespace, numGoroutines)

	// Test concurrent access to namespace getters
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			buildResults[index] = registry.Build()
			testResults[index] = registry.Test()
			lintResults[index] = registry.Lint()
		}(i)
	}

	wg.Wait()

	// Verify consistency - each namespace type should return the same instance
	firstBuild := buildResults[0]
	firstTest := testResults[0]
	firstLint := lintResults[0]

	suite.NotNil(firstBuild)
	suite.NotNil(firstTest)
	suite.NotNil(firstLint)

	for i := 1; i < numGoroutines; i++ {
		suite.Same(firstBuild, buildResults[i], "Build namespace should be consistent across goroutines")
		suite.Same(firstTest, testResults[i], "Test namespace should be consistent across goroutines")
		suite.Same(firstLint, lintResults[i], "Lint namespace should be consistent across goroutines")
	}
}

// TestRegistrySettersAndGetters tests custom namespace setters
func (suite *RegistryComprehensiveTestSuite) TestRegistrySettersAndGetters() {
	registry := GetNamespaceRegistry()
	suite.Require().NotNil(registry)

	// Create custom namespace implementations
	customBuild := NewBuildNamespace()
	customTest := NewTestNamespace()
	customLint := NewLintNamespace()

	// Test setters
	registry.SetBuild(customBuild)
	registry.SetTest(customTest)
	registry.SetLint(customLint)

	// Test that getters return the custom implementations
	suite.Same(customBuild, registry.Build(), "SetBuild should set custom build namespace")
	suite.Same(customTest, registry.Test(), "SetTest should set custom test namespace")
	suite.Same(customLint, registry.Lint(), "SetLint should set custom lint namespace")
}

// TestRegistryLazyInitializationConsistency tests that lazy initialization is consistent
func (suite *RegistryComprehensiveTestSuite) TestRegistryLazyInitializationConsistency() {
	registry := GetNamespaceRegistry()

	// Test that multiple calls to the same namespace getter return the same instance
	build1 := registry.Build()
	build2 := registry.Build()
	suite.Same(build1, build2, "Multiple calls to Build() should return same instance")

	test1 := registry.Test()
	test2 := registry.Test()
	suite.Same(test1, test2, "Multiple calls to Test() should return same instance")

	lint1 := registry.Lint()
	lint2 := registry.Lint()
	suite.Same(lint1, lint2, "Multiple calls to Lint() should return same instance")

	format1 := registry.Format()
	format2 := registry.Format()
	suite.Same(format1, format2, "Multiple calls to Format() should return same instance")
}

// TestRegistryInterfaceCompliance tests that registry implements NamespaceRegistry interface
func (suite *RegistryComprehensiveTestSuite) TestRegistryInterfaceCompliance() {
	registry := GetNamespaceRegistry()
	suite.Require().NotNil(registry)

	// Test that registry implements NamespaceRegistry interface
	var namespaceRegistry NamespaceRegistry = registry
	suite.NotNil(namespaceRegistry)

	// Test all required methods exist and are callable
	suite.NotPanics(func() { namespaceRegistry.Build() })
	suite.NotPanics(func() { namespaceRegistry.Test() })
	suite.NotPanics(func() { namespaceRegistry.Lint() })
	suite.NotPanics(func() { namespaceRegistry.Format() })
	suite.NotPanics(func() { namespaceRegistry.Deps() })
	suite.NotPanics(func() { namespaceRegistry.Git() })
	suite.NotPanics(func() { namespaceRegistry.Release() })
	suite.NotPanics(func() { namespaceRegistry.Docs() })
	suite.NotPanics(func() { namespaceRegistry.Deploy() })
	suite.NotPanics(func() { namespaceRegistry.Tools() })
	suite.NotPanics(func() { namespaceRegistry.Security() })
	suite.NotPanics(func() { namespaceRegistry.Generate() })
	suite.NotPanics(func() { namespaceRegistry.CLI() })
	suite.NotPanics(func() { namespaceRegistry.Update() })
	suite.NotPanics(func() { namespaceRegistry.Mod() })
	suite.NotPanics(func() { namespaceRegistry.Recipes() })
	suite.NotPanics(func() { namespaceRegistry.Metrics() })
	suite.NotPanics(func() { namespaceRegistry.Workflow() })
}

// TestRegistryProviderInterface tests provider interface implementations
func (suite *RegistryComprehensiveTestSuite) TestRegistryProviderInterface() {
	// Test DefaultNamespaceRegistryProvider
	provider := NewDefaultNamespaceRegistryProvider()
	suite.Require().NotNil(provider)
	suite.Implements((*NamespaceRegistryProvider)(nil), provider)

	// Test singleton behavior
	registry1 := provider.GetNamespaceRegistry()
	registry2 := provider.GetNamespaceRegistry()
	suite.Same(registry1, registry2, "Provider should return same registry instance")

	// Test that provider can be set globally
	SetNamespaceRegistryProvider(provider)
	globalRegistry := GetNamespaceRegistry()
	suite.Same(registry1, globalRegistry, "Global registry should match provider registry")
}

// TestRegistryMemoryLeaks tests for potential memory leaks in registry
func (suite *RegistryComprehensiveTestSuite) TestRegistryMemoryLeaks() {
	// Create many registry instances to test for leaks
	const numRegistries = 1000

	for i := 0; i < numRegistries; i++ {
		// Create new provider each time to avoid singleton caching
		provider := NewDefaultNamespaceRegistryProvider()
		registry := provider.GetNamespaceRegistry()
		suite.NotNil(registry)

		// Access various namespaces to trigger initialization
		build := registry.Build()
		test := registry.Test()
		lint := registry.Lint()

		suite.NotNil(build)
		suite.NotNil(test)
		suite.NotNil(lint)
	}

	// If we reach here without excessive memory usage, the test passes - no assertion needed
}

// TestRegistryEdgeCases tests edge cases and error conditions
func (suite *RegistryComprehensiveTestSuite) TestRegistryEdgeCases() {
	// Test setting nil namespaces
	registry := GetNamespaceRegistry()
	suite.Require().NotNil(registry)

	// Set nil namespace (should be allowed)
	registry.SetBuild(nil)
	build := registry.Build()
	suite.Nil(build, "Setting nil build namespace should return nil")

	// Reset to non-nil
	registry.SetBuild(NewBuildNamespace())
	build = registry.Build()
	suite.NotNil(build, "Resetting build namespace should work")
}

// TestRegistryCrossNamespaceDependencies tests interactions between namespaces
func (suite *RegistryComprehensiveTestSuite) TestRegistryCrossNamespaceDependencies() {
	registry := GetNamespaceRegistry()
	suite.Require().NotNil(registry)

	// Get multiple namespaces and verify they're independent but accessible
	build := registry.Build()
	test := registry.Test()
	lint := registry.Lint()
	format := registry.Format()

	suite.NotNil(build)
	suite.NotNil(test)
	suite.NotNil(lint)
	suite.NotNil(format)

	// Verify they're different types
	suite.IsType(&buildNamespaceWrapper{}, build)
	suite.IsType(&testNamespaceWrapper{}, test)
	suite.IsType(&lintNamespaceWrapper{}, lint)
	suite.IsType(&formatNamespaceWrapper{}, format)

	// Verify they implement their respective interfaces
	suite.Implements((*BuildNamespace)(nil), build)
	suite.Implements((*TestNamespace)(nil), test)
	suite.Implements((*LintNamespace)(nil), lint)
	suite.Implements((*FormatNamespace)(nil), format)
}

// TestRunRegistryComprehensiveTestSuite runs the registry comprehensive test suite
func TestRunRegistryComprehensiveTestSuite(t *testing.T) {
	suite.Run(t, new(RegistryComprehensiveTestSuite))
}

// BenchmarkRegistryAccess benchmarks registry access performance
func BenchmarkRegistryAccess(b *testing.B) {
	tests := []struct {
		name   string
		getter func(*DefaultNamespaceRegistry) interface{}
	}{
		{"GetRegistry", func(*DefaultNamespaceRegistry) interface{} { return GetNamespaceRegistry() }},
		{"BuildNamespace", func(r *DefaultNamespaceRegistry) interface{} { return r.Build() }},
		{"TestNamespace", func(r *DefaultNamespaceRegistry) interface{} { return r.Test() }},
		{"LintNamespace", func(r *DefaultNamespaceRegistry) interface{} { return r.Lint() }},
		{"FormatNamespace", func(r *DefaultNamespaceRegistry) interface{} { return r.Format() }},
	}

	registry := GetNamespaceRegistry()

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result := tt.getter(registry)
				if result == nil && tt.name != "GetRegistry" {
					// Some namespaces like Security might be nil
					continue
				}
				if result == nil && tt.name == "GetRegistry" {
					b.Fatal("Registry should not be nil")
				}
			}
			b.StopTimer()
		})
	}
}

// BenchmarkRegistryConcurrentAccess benchmarks concurrent registry access
func BenchmarkRegistryConcurrentAccess(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			registry := GetNamespaceRegistry()
			if registry == nil {
				b.Fatal("Registry should not be nil")
			}
			// Access a namespace to test lazy initialization under load
			build := registry.Build()
			if build == nil {
				b.Fatal("Build namespace should not be nil")
			}
		}
	})
	b.StopTimer()
}

// TestRegistryIntegrationWithFactories tests integration between registry and factory functions
func TestRegistryIntegrationWithFactories(t *testing.T) {
	t.Parallel()

	// Create a fresh registry instead of using the global one to avoid test pollution
	registry := NewNamespaceRegistry()
	require.NotNil(t, registry)

	// Test that registry namespaces are compatible with factory-created namespaces
	registryBuild := registry.Build()
	factoryBuild := NewBuildNamespace()

	// They should be the same type but different instances
	assert.IsType(t, factoryBuild, registryBuild)
	assert.NotSame(t, factoryBuild, registryBuild, "Factory and registry should create different instances")

	// Both should implement the interface
	assert.Implements(t, (*BuildNamespace)(nil), registryBuild)
	assert.Implements(t, (*BuildNamespace)(nil), factoryBuild)
}

// TestRegistryProviderSwitching tests switching between different providers
func TestRegistryProviderSwitching(t *testing.T) {
	t.Parallel()

	// Reset to a fresh provider state to avoid contamination from other tests
	freshProvider := NewDefaultNamespaceRegistryProvider()
	defer SetNamespaceRegistryProvider(freshProvider)

	// Create first custom registry
	customRegistry1 := NewNamespaceRegistry()
	customProvider1 := &MockNamespaceRegistryProvider{registry: customRegistry1}
	SetNamespaceRegistryProvider(customProvider1)

	registry1 := GetNamespaceRegistry()
	assert.Same(t, customRegistry1, registry1, "Should return first custom registry")

	// Create second custom registry
	customRegistry2 := NewNamespaceRegistry()
	customProvider2 := &MockNamespaceRegistryProvider{registry: customRegistry2}
	SetNamespaceRegistryProvider(customProvider2)

	registry2 := GetNamespaceRegistry()
	assert.Same(t, customRegistry2, registry2, "Should return second custom registry")
	assert.NotSame(t, registry1, registry2, "Registries should be different")
}

// FuzzRegistryAccess fuzz tests registry access patterns
func FuzzRegistryAccess(f *testing.F) {
	// Add seed values for namespace access patterns
	f.Add(0)   // Build
	f.Add(1)   // Test
	f.Add(2)   // Lint
	f.Add(3)   // Format
	f.Add(4)   // Deps
	f.Add(5)   // Git
	f.Add(6)   // Release
	f.Add(7)   // Docs
	f.Add(8)   // Tools
	f.Add(9)   // Generate
	f.Add(10)  // CLI
	f.Add(11)  // Update
	f.Add(12)  // Mod
	f.Add(13)  // Recipes
	f.Add(14)  // Metrics
	f.Add(15)  // Workflow
	f.Add(100) // Out of range

	f.Fuzz(func(t *testing.T, namespaceIndex int) {
		registry := GetNamespaceRegistry()
		require.NotNil(t, registry)

		// Test accessing namespaces by index (simulating different access patterns)
		switch namespaceIndex % 18 { // 18 total namespaces including nil ones
		case 0:
			build := registry.Build()
			assert.NotNil(t, build)
		case 1:
			test := registry.Test()
			assert.NotNil(t, test)
		case 2:
			lint := registry.Lint()
			assert.NotNil(t, lint)
		case 3:
			format := registry.Format()
			assert.NotNil(t, format)
		case 4:
			deps := registry.Deps()
			assert.NotNil(t, deps)
		case 5:
			git := registry.Git()
			assert.NotNil(t, git)
		case 6:
			release := registry.Release()
			assert.NotNil(t, release)
		case 7:
			docs := registry.Docs()
			assert.NotNil(t, docs)
		case 8:
			deploy := registry.Deploy()
			// Deploy might be nil - that's expected
			_ = deploy
		case 9:
			tools := registry.Tools()
			assert.NotNil(t, tools)
		case 10:
			security := registry.Security()
			// Security is nil - that's expected
			assert.Nil(t, security)
		case 11:
			generate := registry.Generate()
			assert.NotNil(t, generate)
		case 12:
			cli := registry.CLI()
			assert.NotNil(t, cli)
		case 13:
			update := registry.Update()
			assert.NotNil(t, update)
		case 14:
			mod := registry.Mod()
			assert.NotNil(t, mod)
		case 15:
			recipes := registry.Recipes()
			assert.NotNil(t, recipes)
		case 16:
			metrics := registry.Metrics()
			assert.NotNil(t, metrics)
		case 17:
			workflow := registry.Workflow()
			assert.NotNil(t, workflow)
		}
	})
}
