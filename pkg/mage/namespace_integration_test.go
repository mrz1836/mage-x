package mage

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// NamespaceIntegrationTestSuite provides integration tests for namespace interactions
type NamespaceIntegrationTestSuite struct {
	suite.Suite

	registry         *DefaultNamespaceRegistry
	originalProvider NamespaceRegistryProvider
}

// SetupTest creates a clean registry for each test
func (suite *NamespaceIntegrationTestSuite) SetupTest() {
	suite.originalProvider = NewDefaultNamespaceRegistryProvider()
	suite.registry = NewNamespaceRegistry()
	SetNamespaceRegistryProvider(&MockNamespaceRegistryProvider{registry: suite.registry})
}

// TearDownTest restores the original provider
func (suite *NamespaceIntegrationTestSuite) TearDownTest() {
	if suite.originalProvider != nil {
		SetNamespaceRegistryProvider(suite.originalProvider)
	}
}

// TestNamespaceDiscovery tests namespace discovery and registration
func (suite *NamespaceIntegrationTestSuite) TestNamespaceDiscovery() {
	registry := GetNamespaceRegistry()
	suite.Require().NotNil(registry)

	// Test discovery of all registered namespaces
	namespaces := []struct {
		name   string
		getter func() interface{}
		notNil bool
	}{
		{"Build", func() interface{} { return registry.Build() }, true},
		{"Test", func() interface{} { return registry.Test() }, true},
		{"Lint", func() interface{} { return registry.Lint() }, true},
		{"Format", func() interface{} { return registry.Format() }, true},
		{"Deps", func() interface{} { return registry.Deps() }, true},
		{"Git", func() interface{} { return registry.Git() }, true},
		{"Release", func() interface{} { return registry.Release() }, true},
		{"Docs", func() interface{} { return registry.Docs() }, true},
		{"Deploy", func() interface{} { return registry.Deploy() }, false}, // Not implemented
		{"Tools", func() interface{} { return registry.Tools() }, true},
		{"Security", func() interface{} { return registry.Security() }, false}, // Disabled
		{"Generate", func() interface{} { return registry.Generate() }, true},
		{"CLI", func() interface{} { return registry.CLI() }, true},
		{"Update", func() interface{} { return registry.Update() }, true},
		{"Mod", func() interface{} { return registry.Mod() }, true},
		{"Recipes", func() interface{} { return registry.Recipes() }, true},
		{"Metrics", func() interface{} { return registry.Metrics() }, true},
		{"Workflow", func() interface{} { return registry.Workflow() }, true},
	}

	for _, ns := range namespaces {
		suite.Run(ns.name, func() {
			result := ns.getter()
			if ns.notNil {
				suite.NotNil(result, "Namespace %s should not be nil", ns.name)
			} else {
				suite.Nil(result, "Namespace %s should be nil", ns.name)
			}
		})
	}
}

// TestCommandRouting tests command routing between namespaces
func (suite *NamespaceIntegrationTestSuite) TestCommandRouting() {
	registry := GetNamespaceRegistry()
	suite.Require().NotNil(registry)

	// Test that commands can be accessed through different namespaces
	build := registry.Build()
	test := registry.Test()
	lint := registry.Lint()
	format := registry.Format()

	suite.NotNil(build)
	suite.NotNil(test)
	suite.NotNil(lint)
	suite.NotNil(format)

	// Verify each namespace has its expected methods
	suite.NotNil(build.Default, "Build namespace should have Default method")
	suite.NotNil(build.Clean, "Build namespace should have Clean method")
	suite.NotNil(build.Install, "Build namespace should have Install method")

	suite.NotNil(test.Default, "Test namespace should have Default method")
	suite.NotNil(test.Unit, "Test namespace should have Unit method")
	suite.NotNil(test.Race, "Test namespace should have Race method")

	suite.NotNil(lint.Default, "Lint namespace should have Default method")
	suite.NotNil(lint.All, "Lint namespace should have All method")

	suite.NotNil(format.Default, "Format namespace should have Default method")
	suite.NotNil(format.Check, "Format namespace should have Check method")
}

// TestCrossNamespaceDependencies tests dependencies between namespaces
func (suite *NamespaceIntegrationTestSuite) TestCrossNamespaceDependencies() {
	registry := GetNamespaceRegistry()
	suite.Require().NotNil(registry)

	// Get multiple namespaces that might have dependencies
	build := registry.Build()
	test := registry.Test()
	lint := registry.Lint()
	format := registry.Format()
	deps := registry.Deps()

	// Verify all namespaces are available
	suite.NotNil(build)
	suite.NotNil(test)
	suite.NotNil(lint)
	suite.NotNil(format)
	suite.NotNil(deps)

	// Test that namespaces are independent instances
	suite.NotSame(build, test, "Build and Test should be different instances")
	suite.NotSame(test, lint, "Test and Lint should be different instances")
	suite.NotSame(lint, format, "Lint and Format should be different instances")

	// Test that they implement different interfaces
	suite.Implements((*BuildNamespace)(nil), build)
	suite.Implements((*TestNamespace)(nil), test)
	suite.Implements((*LintNamespace)(nil), lint)
	suite.Implements((*FormatNamespace)(nil), format)
	suite.Implements((*DepsNamespace)(nil), deps)
}

// TestNamespaceCustomization tests custom namespace implementations
func (suite *NamespaceIntegrationTestSuite) TestNamespaceCustomization() {
	registry := GetNamespaceRegistry()
	suite.Require().NotNil(registry)

	// Create custom namespace implementations
	customBuild := NewBuildNamespace()
	customTest := NewTestNamespace()

	// Set custom implementations
	registry.SetBuild(customBuild)
	registry.SetTest(customTest)

	// Verify custom implementations are returned
	suite.Same(customBuild, registry.Build(), "Custom build namespace should be returned")
	suite.Same(customTest, registry.Test(), "Custom test namespace should be returned")

	// Test that other namespaces are still default
	lint := registry.Lint()
	suite.NotNil(lint)
	suite.IsType(&lintNamespaceWrapper{}, lint, "Lint should still be default wrapper")
}

// TestNamespaceMethodsAccessibility tests that all namespace methods are accessible
func (suite *NamespaceIntegrationTestSuite) TestNamespaceMethodsAccessibility() {
	registry := GetNamespaceRegistry()
	suite.Require().NotNil(registry)

	// Test Build namespace methods
	build := registry.Build()
	suite.Require().NotNil(build)
	suite.NotPanics(func() { _ = build.Default }, "Build.Default should be accessible")
	suite.NotPanics(func() { _ = build.All }, "Build.All should be accessible")
	suite.NotPanics(func() { _ = build.Clean }, "Build.Clean should be accessible")
	suite.NotPanics(func() { _ = build.Install }, "Build.Install should be accessible")
	suite.NotPanics(func() { _ = build.Generate }, "Build.Generate should be accessible")

	// Test Test namespace methods
	test := registry.Test()
	suite.Require().NotNil(test)
	suite.NotPanics(func() { _ = test.Default }, "Test.Default should be accessible")
	suite.NotPanics(func() { _ = test.Unit }, "Test.Unit should be accessible")
	suite.NotPanics(func() { _ = test.Race }, "Test.Race should be accessible")
	suite.NotPanics(func() { _ = test.Cover }, "Test.Cover should be accessible")
	suite.NotPanics(func() { _ = test.Bench }, "Test.Bench should be accessible")

	// Test Lint namespace methods
	lint := registry.Lint()
	suite.Require().NotNil(lint)
	suite.NotPanics(func() { _ = lint.Default }, "Lint.Default should be accessible")
	suite.NotPanics(func() { _ = lint.All }, "Lint.All should be accessible")
	suite.NotPanics(func() { _ = lint.Go }, "Lint.Go should be accessible")
	suite.NotPanics(func() { _ = lint.Fix }, "Lint.Fix should be accessible")

	// Test Format namespace methods
	format := registry.Format()
	suite.Require().NotNil(format)
	suite.NotPanics(func() { _ = format.Default }, "Format.Default should be accessible")
	suite.NotPanics(func() { _ = format.Check }, "Format.Check should be accessible")
	suite.NotPanics(func() { _ = format.Go }, "Format.Go should be accessible")
	suite.NotPanics(func() { _ = format.All }, "Format.All should be accessible")
}

// TestNamespaceInterfaceCompliance tests interface compliance across all namespaces
func (suite *NamespaceIntegrationTestSuite) TestNamespaceInterfaceCompliance() {
	registry := GetNamespaceRegistry()
	suite.Require().NotNil(registry)

	tests := []struct {
		name          string
		namespace     interface{}
		interfaceType interface{}
	}{
		{"Build", registry.Build(), (*BuildNamespace)(nil)},
		{"Test", registry.Test(), (*TestNamespace)(nil)},
		{"Lint", registry.Lint(), (*LintNamespace)(nil)},
		{"Format", registry.Format(), (*FormatNamespace)(nil)},
		{"Deps", registry.Deps(), (*DepsNamespace)(nil)},
		{"Git", registry.Git(), (*GitNamespace)(nil)},
		{"Release", registry.Release(), (*ReleaseNamespace)(nil)},
		{"Docs", registry.Docs(), (*DocsNamespace)(nil)},
		{"Tools", registry.Tools(), (*ToolsNamespace)(nil)},
		{"Generate", registry.Generate(), (*GenerateNamespace)(nil)},
		{"CLI", registry.CLI(), (*CLINamespace)(nil)},
		{"Update", registry.Update(), (*UpdateNamespace)(nil)},
		{"Mod", registry.Mod(), (*ModNamespace)(nil)},
		{"Recipes", registry.Recipes(), (*RecipesNamespace)(nil)},
		{"Metrics", registry.Metrics(), (*MetricsNamespace)(nil)},
		{"Workflow", registry.Workflow(), (*WorkflowNamespace)(nil)},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.namespace != nil {
				suite.Implements(tt.interfaceType, tt.namespace, "Namespace %s should implement its interface", tt.name)
			}
		})
	}
}

// TestNamespaceErrorHandling tests error handling in namespace operations
func (suite *NamespaceIntegrationTestSuite) TestNamespaceErrorHandling() {
	registry := GetNamespaceRegistry()
	suite.Require().NotNil(registry)

	// Test that nil namespaces are handled correctly
	registry.SetBuild(nil)
	build := registry.Build()
	suite.Nil(build, "Setting nil build namespace should return nil")

	// Test that setting nil doesn't affect other namespaces
	test := registry.Test()
	suite.NotNil(test, "Other namespaces should not be affected by nil setting")

	// Reset build namespace
	registry.SetBuild(NewBuildNamespace())
	build = registry.Build()
	suite.NotNil(build, "Build namespace should be restored")
}

// TestNamespaceConcurrency tests concurrent access to namespaces
func (suite *NamespaceIntegrationTestSuite) TestNamespaceConcurrency() {
	const numGoroutines = 100
	const numOperations = 10

	registry := GetNamespaceRegistry()
	suite.Require().NotNil(registry)

	var wg sync.WaitGroup
	results := make([][]interface{}, numGoroutines)

	// Test concurrent access to multiple namespaces
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index] = make([]interface{}, numOperations)
			for j := 0; j < numOperations; j++ {
				// Rotate through different namespaces
				switch j % 4 {
				case 0:
					results[index][j] = registry.Build()
				case 1:
					results[index][j] = registry.Test()
				case 2:
					results[index][j] = registry.Lint()
				case 3:
					results[index][j] = registry.Format()
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify consistency of results
	firstBuild := results[0][0]
	firstTest := results[0][1]
	firstLint := results[0][2]
	firstFormat := results[0][3]

	for i, goroutineResults := range results {
		for j, result := range goroutineResults {
			switch j % 4 {
			case 0:
				suite.Same(firstBuild, result, "Build namespace should be consistent (goroutine %d, op %d)", i, j)
			case 1:
				suite.Same(firstTest, result, "Test namespace should be consistent (goroutine %d, op %d)", i, j)
			case 2:
				suite.Same(firstLint, result, "Lint namespace should be consistent (goroutine %d, op %d)", i, j)
			case 3:
				suite.Same(firstFormat, result, "Format namespace should be consistent (goroutine %d, op %d)", i, j)
			}
		}
	}
}

// TestNamespaceLifecycle tests namespace lifecycle management
func (suite *NamespaceIntegrationTestSuite) TestNamespaceLifecycle() {
	registry := GetNamespaceRegistry()
	suite.Require().NotNil(registry)

	// Test initial state
	build1 := registry.Build()
	suite.NotNil(build1)

	// Test that subsequent calls return same instance
	build2 := registry.Build()
	suite.Same(build1, build2, "Subsequent calls should return same instance")

	// Test custom namespace setting
	customBuild := NewBuildNamespace()
	registry.SetBuild(customBuild)
	build3 := registry.Build()
	suite.Same(customBuild, build3, "Custom namespace should be returned after setting")
	suite.NotSame(build1, build3, "Custom namespace should be different from original")

	// Test that other namespaces are unaffected
	test := registry.Test()
	suite.NotNil(test)
	suite.NotSame(customBuild, test, "Other namespaces should be unaffected")
}

// TestNamespaceProviderIntegration tests integration with provider pattern
func (suite *NamespaceIntegrationTestSuite) TestNamespaceProviderIntegration() {
	// Create custom registry and provider
	customRegistry := NewNamespaceRegistry()
	customProvider := &MockNamespaceRegistryProvider{registry: customRegistry}

	// Set custom provider
	SetNamespaceRegistryProvider(customProvider)

	// Verify custom registry is used
	actualRegistry := GetNamespaceRegistry()
	suite.Same(customRegistry, actualRegistry, "Custom provider should provide custom registry")

	// Test that namespace operations work with custom registry
	build := actualRegistry.Build()
	suite.NotNil(build)
	suite.Implements((*BuildNamespace)(nil), build)

	// Verify provider was called
	suite.True(customProvider.getCalled, "Provider should have been called")
}

// TestNamespaceContextIntegration tests context integration across namespaces
func (suite *NamespaceIntegrationTestSuite) TestNamespaceContextIntegration() {
	registry := GetNamespaceRegistry()
	suite.Require().NotNil(registry)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test that namespaces can be accessed within context
	done := make(chan bool, 1)
	go func() {
		defer func() { done <- true }()

		// Access multiple namespaces within the context
		build := registry.Build()
		test := registry.Test()
		lint := registry.Lint()

		suite.NotNil(build)
		suite.NotNil(test)
		suite.NotNil(lint)
	}()

	select {
	case <-done:
		// Success
	case <-ctx.Done():
		suite.Fail("Context timeout - namespace access took too long")
	}
}

// TestRunNamespaceIntegrationTestSuite runs the namespace integration test suite
func TestRunNamespaceIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(NamespaceIntegrationTestSuite))
}

// TestNamespaceCommandChaining tests chaining commands across namespaces
func TestNamespaceCommandChaining(t *testing.T) {
	t.Parallel()

	registry := GetNamespaceRegistry()
	require.NotNil(t, registry)

	// Test that different namespace commands can be accessed in sequence
	// This simulates a workflow where multiple namespace operations might be chained
	build := registry.Build()
	test := registry.Test()
	lint := registry.Lint()
	format := registry.Format()

	assert.NotNil(t, build)
	assert.NotNil(t, test)
	assert.NotNil(t, lint)
	assert.NotNil(t, format)

	// Verify method accessibility in chain
	assert.NotNil(t, build.Default)
	assert.NotNil(t, test.Default)
	assert.NotNil(t, lint.Default)
	assert.NotNil(t, format.Default)

	// Test that each namespace maintains its identity
	assert.IsType(t, &buildNamespaceWrapper{}, build)
	assert.IsType(t, &testNamespaceWrapper{}, test)
	assert.IsType(t, &lintNamespaceWrapper{}, lint)
	assert.IsType(t, &formatNamespaceWrapper{}, format)
}

// TestNamespaceRegistryReset tests resetting registry state
func TestNamespaceRegistryReset(t *testing.T) {
	t.Parallel()

	// Save original provider
	originalProvider := NewDefaultNamespaceRegistryProvider()
	defer SetNamespaceRegistryProvider(originalProvider)

	// Create test registry
	testRegistry := NewNamespaceRegistry()
	SetNamespaceRegistryProvider(&MockNamespaceRegistryProvider{registry: testRegistry})

	// Get initial namespaces
	build1 := testRegistry.Build()
	test1 := testRegistry.Test()

	assert.NotNil(t, build1)
	assert.NotNil(t, test1)

	// Set custom namespaces
	customBuild := NewBuildNamespace()
	customTest := NewTestNamespace()
	testRegistry.SetBuild(customBuild)
	testRegistry.SetTest(customTest)

	// Verify custom namespaces are returned
	build2 := testRegistry.Build()
	test2 := testRegistry.Test()

	assert.Same(t, customBuild, build2)
	assert.Same(t, customTest, test2)
	assert.NotSame(t, build1, build2)
	assert.NotSame(t, test1, test2)
}

// TestNamespaceMethodSignatureConsistency tests method signature consistency
func TestNamespaceMethodSignatureConsistency(t *testing.T) {
	t.Parallel()

	registry := GetNamespaceRegistry()
	require.NotNil(t, registry)

	// Test that Default methods have consistent signatures across namespaces
	build := registry.Build()
	test := registry.Test()
	lint := registry.Lint()
	format := registry.Format()

	require.NotNil(t, build)
	require.NotNil(t, test)
	require.NotNil(t, lint)
	require.NotNil(t, format)

	// Verify Default method signatures (func(...string) error for test namespace, func() error for others)
	assert.IsType(t, func() error { return nil }, build.Default)
	assert.IsType(t, func(args ...string) error { return nil }, test.Default)
	assert.IsType(t, func() error { return nil }, lint.Default)
	assert.IsType(t, func() error { return nil }, format.Default)

	// Test other common method signatures
	assert.IsType(t, func() error { return nil }, build.Clean)
	assert.IsType(t, func(args ...string) error { return nil }, test.Unit)
	assert.IsType(t, func() error { return nil }, lint.All)
	assert.IsType(t, func() error { return nil }, format.Check)
}

// BenchmarkNamespaceIntegration benchmarks integrated namespace operations
func BenchmarkNamespaceIntegration(b *testing.B) {
	registry := GetNamespaceRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate accessing multiple namespaces in sequence
		build := registry.Build()
		test := registry.Test()
		lint := registry.Lint()
		format := registry.Format()

		if build == nil || test == nil || lint == nil || format == nil {
			b.Fatal("Namespace should not be nil")
		}
	}
	b.StopTimer()
}

// BenchmarkNamespaceIntegrationConcurrent benchmarks concurrent namespace access
func BenchmarkNamespaceIntegrationConcurrent(b *testing.B) {
	registry := GetNamespaceRegistry()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Access different namespaces concurrently
			build := registry.Build()
			test := registry.Test()
			lint := registry.Lint()

			if build == nil || test == nil || lint == nil {
				b.Fatal("Namespace should not be nil")
			}
		}
	})
	b.StopTimer()
}

// FuzzNamespaceIntegration fuzz tests namespace integration
func FuzzNamespaceIntegration(f *testing.F) {
	// Add seed values for different namespace access patterns
	f.Add(0, 1, 2, 3)     // Build, Test, Lint, Format
	f.Add(4, 5, 6, 7)     // Deps, Git, Release, Docs
	f.Add(8, 9, 10, 11)   // Tools, Generate, CLI, Update
	f.Add(12, 13, 14, 15) // Mod, Recipes, Metrics, Workflow

	f.Fuzz(func(t *testing.T, ns1, ns2, ns3, ns4 int) {
		registry := GetNamespaceRegistry()
		require.NotNil(t, registry)

		// Map indices to namespace getters
		getters := []func() interface{}{
			func() interface{} { return registry.Build() },
			func() interface{} { return registry.Test() },
			func() interface{} { return registry.Lint() },
			func() interface{} { return registry.Format() },
			func() interface{} { return registry.Deps() },
			func() interface{} { return registry.Git() },
			func() interface{} { return registry.Release() },
			func() interface{} { return registry.Docs() },
			func() interface{} { return registry.Tools() },
			func() interface{} { return registry.Generate() },
			func() interface{} { return registry.CLI() },
			func() interface{} { return registry.Update() },
			func() interface{} { return registry.Mod() },
			func() interface{} { return registry.Recipes() },
			func() interface{} { return registry.Metrics() },
			func() interface{} { return registry.Workflow() },
		}

		// Access namespaces based on fuzz input
		indices := []int{ns1, ns2, ns3, ns4}
		for _, idx := range indices {
			if idx >= 0 && idx < len(getters) {
				result := getters[idx]()
				// Some namespaces might be nil (Security, Deploy) - that's expected
				_ = result
			}
		}
	})
}
