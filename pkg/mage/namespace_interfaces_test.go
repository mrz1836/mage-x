// Package mage provides namespace interfaces for mage build operations
package mage

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// NamespaceInterfacesTestSuite provides a test suite for namespace interfaces
type NamespaceInterfacesTestSuite struct {
	suite.Suite

	originalProvider NamespaceRegistryProvider
}

// SetupTest saves the original provider before each test
func (suite *NamespaceInterfacesTestSuite) SetupTest() {
	// Create a new provider that wraps the current registry
	suite.originalProvider = NewDefaultNamespaceRegistryProvider()
}

// TearDownTest restores the original provider after each test
func (suite *NamespaceInterfacesTestSuite) TearDownTest() {
	if suite.originalProvider != nil {
		SetNamespaceRegistryProvider(suite.originalProvider)
	}
}

// TestDefaultNamespaceRegistryProvider tests the default provider implementation
func (suite *NamespaceInterfacesTestSuite) TestDefaultNamespaceRegistryProvider() {
	provider := NewDefaultNamespaceRegistryProvider()
	suite.Require().NotNil(provider)

	// First call should create the registry
	registry1 := provider.GetNamespaceRegistry()
	suite.Require().NotNil(registry1)

	// Second call should return the same instance (singleton)
	registry2 := provider.GetNamespaceRegistry()
	suite.Same(registry1, registry2)
}

// TestGetNamespaceRegistry tests the global getter function
func (suite *NamespaceInterfacesTestSuite) TestGetNamespaceRegistry() {
	// Get the registry multiple times
	registry1 := GetNamespaceRegistry()
	registry2 := GetNamespaceRegistry()

	suite.Require().NotNil(registry1)
	suite.Require().NotNil(registry2)
	suite.Same(registry1, registry2, "GetNamespaceRegistry should return the same instance")
}

// TestSetNamespaceRegistryProvider tests setting a custom provider
func (suite *NamespaceInterfacesTestSuite) TestSetNamespaceRegistryProvider() {
	// Create a mock provider
	mockProvider := &MockNamespaceRegistryProvider{
		registry: &DefaultNamespaceRegistry{},
	}

	// Set the mock provider
	SetNamespaceRegistryProvider(mockProvider)

	// Verify the mock provider is used
	registry := GetNamespaceRegistry()
	suite.Same(mockProvider.registry, registry)
	suite.True(mockProvider.getCalled, "GetNamespaceRegistry should have been called on the mock provider")
}

// TestSetNamespaceRegistryLegacy tests the legacy setter function
func (suite *NamespaceInterfacesTestSuite) TestSetNamespaceRegistryLegacy() {
	// Create a custom registry
	customRegistry := &DefaultNamespaceRegistry{}

	// Set it using the legacy function
	SetNamespaceRegistry(customRegistry)

	// Verify it's returned by the getter
	registry := GetNamespaceRegistry()
	suite.Same(customRegistry, registry)
}

// TestNamespaceRegistryMethods tests that all namespace methods work
func (suite *NamespaceInterfacesTestSuite) TestNamespaceRegistryMethods() {
	registry := GetNamespaceRegistry()
	suite.Require().NotNil(registry)

	// Test all namespace getters - they should not panic
	suite.NotPanics(func() { registry.Build() })
	suite.NotPanics(func() { registry.Test() })
	suite.NotPanics(func() { registry.Lint() })
	suite.NotPanics(func() { registry.Format() })
	suite.NotPanics(func() { registry.Deps() })
	suite.NotPanics(func() { registry.Git() })
	suite.NotPanics(func() { registry.Release() })
	suite.NotPanics(func() { registry.Docs() })
	suite.NotPanics(func() { registry.Deploy() })
	suite.NotPanics(func() { registry.Tools() })
	suite.NotPanics(func() { registry.Security() })
	suite.NotPanics(func() { registry.Generate() })
	suite.NotPanics(func() { registry.CLI() })
	suite.NotPanics(func() { registry.Update() })
	suite.NotPanics(func() { registry.Mod() })
	suite.NotPanics(func() { registry.Recipes() })
	suite.NotPanics(func() { registry.Metrics() })
	suite.NotPanics(func() { registry.Workflow() })
}

// TestConcurrentAccess tests thread-safety of the provider pattern
func (suite *NamespaceInterfacesTestSuite) TestConcurrentAccess() {
	const numGoroutines = 100
	registries := make([]*DefaultNamespaceRegistry, numGoroutines)
	var wg sync.WaitGroup

	// Launch multiple goroutines that all try to get the registry
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			registries[index] = GetNamespaceRegistry()
		}(i)
	}

	wg.Wait()

	// All should return the same instance
	firstRegistry := registries[0]
	suite.Require().NotNil(firstRegistry)

	for i := 1; i < numGoroutines; i++ {
		suite.Same(firstRegistry, registries[i], "All goroutines should get the same registry instance")
	}
}

// TestConcurrentProviderSwitch tests behavior when switching providers
// Note: Provider switching itself is not thread-safe by design, so we test
// sequential switching followed by concurrent reads
func (suite *NamespaceInterfacesTestSuite) TestConcurrentProviderSwitch() {
	const numReaders = 50
	var wg sync.WaitGroup

	// Create multiple providers
	providers := make([]*MockNamespaceRegistryProvider, 3)
	for i := range providers {
		providers[i] = &MockNamespaceRegistryProvider{
			registry: &DefaultNamespaceRegistry{},
		}
	}

	// Test sequential provider switching
	for i, provider := range providers {
		SetNamespaceRegistryProvider(provider)
		registry := GetNamespaceRegistry()
		suite.Same(provider.registry, registry, "Provider %d should be active", i)
		suite.True(provider.getCalled, "Provider %d should have been called", i)
	}

	// Now test concurrent reads after setting a provider
	// Set the last provider
	SetNamespaceRegistryProvider(providers[len(providers)-1])

	results := make([]*DefaultNamespaceRegistry, numReaders)

	// Launch goroutines that only read the registry
	for i := 0; i < numReaders; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			// Only perform reads - no provider switching
			results[index] = GetNamespaceRegistry()
		}(i)
	}

	wg.Wait()

	// Verify all results are valid and the same
	expectedRegistry := providers[len(providers)-1].registry
	for i, result := range results {
		suite.NotNil(result, "Result %d should not be nil", i)
		suite.Same(expectedRegistry, result, "All readers should get the same registry instance")
	}
}

// TestProviderInterface tests that the provider interface is properly implemented
func (suite *NamespaceInterfacesTestSuite) TestProviderInterface() {
	// Verify DefaultNamespaceRegistryProvider implements NamespaceRegistryProvider
	var provider NamespaceRegistryProvider = NewDefaultNamespaceRegistryProvider()
	suite.Require().NotNil(provider)

	registry := provider.GetNamespaceRegistry()
	suite.NotNil(registry)
}

// MockNamespaceRegistryProvider is a mock implementation for testing
type MockNamespaceRegistryProvider struct {
	registry  *DefaultNamespaceRegistry
	getCalled bool
	mu        sync.RWMutex
}

// GetNamespaceRegistry returns the mock registry and marks that it was called
func (m *MockNamespaceRegistryProvider) GetNamespaceRegistry() *DefaultNamespaceRegistry {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getCalled = true
	return m.registry
}

// TestRunNamespaceInterfacesTestSuite runs the test suite
func TestRunNamespaceInterfacesTestSuite(t *testing.T) {
	suite.Run(t, new(NamespaceInterfacesTestSuite))
}

// TestDefaultNamespaceRegistryProviderStandalone tests the provider without the suite
func TestDefaultNamespaceRegistryProviderStandalone(t *testing.T) {
	provider := NewDefaultNamespaceRegistryProvider()
	require.NotNil(t, provider)

	// Test singleton behavior
	registry1 := provider.GetNamespaceRegistry()
	registry2 := provider.GetNamespaceRegistry()

	require.NotNil(t, registry1)
	require.NotNil(t, registry2)
	assert.Same(t, registry1, registry2, "Should return the same instance")
}

// TestNamespaceRegistryProviderInterface ensures interface compliance
func TestNamespaceRegistryProviderInterface(t *testing.T) {
	// This will cause a compile error if DefaultNamespaceRegistryProvider doesn't implement NamespaceRegistryProvider
	var _ NamespaceRegistryProvider = (*DefaultNamespaceRegistryProvider)(nil)
	var _ NamespaceRegistryProvider = (*MockNamespaceRegistryProvider)(nil)
}

// BenchmarkGetNamespaceRegistry benchmarks the registry getter performance
func BenchmarkGetNamespaceRegistry(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		registry := GetNamespaceRegistry()
		if registry == nil {
			b.Fatal("Registry should not be nil")
		}
	}
}

// BenchmarkGetNamespaceRegistryConcurrent benchmarks concurrent access
func BenchmarkGetNamespaceRegistryConcurrent(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			registry := GetNamespaceRegistry()
			if registry == nil {
				b.Fatal("Registry should not be nil")
			}
		}
	})
}
