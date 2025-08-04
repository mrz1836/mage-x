package mage

import (
	"testing"

	"github.com/mrz1836/mage-x/pkg/common/cache"
	"github.com/stretchr/testify/suite"
)

// MockCacheManagerProvider is a mock implementation of CacheManagerProvider for testing
type MockCacheManagerProvider struct {
	manager *cache.Manager
}

// GetCacheManager returns the mock cache manager
func (m *MockCacheManagerProvider) GetCacheManager() *cache.Manager {
	return m.manager
}

// BuildCacheTestSuite tests the cache manager provider functionality
type BuildCacheTestSuite struct {
	suite.Suite

	originalProvider CacheManagerProvider
}

// SetupTest saves the original provider and sets up a clean state
func (s *BuildCacheTestSuite) SetupTest() {
	// Save the original provider by getting it first
	s.originalProvider = &DefaultCacheManagerProvider{}
}

// TearDownTest restores the original provider
func (s *BuildCacheTestSuite) TearDownTest() {
	// Reset to default provider
	SetCacheManagerProvider(NewDefaultCacheManagerProvider())
}

// TestDefaultCacheManagerProvider tests the default cache manager provider
func (s *BuildCacheTestSuite) TestDefaultCacheManagerProvider() {
	provider := NewDefaultCacheManagerProvider()
	s.Require().NotNil(provider)

	// Test thread-safe singleton behavior
	manager1 := provider.GetCacheManager()
	manager2 := provider.GetCacheManager()

	s.NotNil(manager1)
	s.NotNil(manager2)
	s.Same(manager1, manager2, "Should return same instance (singleton)")
}

// TestCacheManagerProviderInterface tests the CacheManagerProvider interface
func (s *BuildCacheTestSuite) TestCacheManagerProviderInterface() {
	// Test that DefaultCacheManagerProvider implements the interface
	var provider CacheManagerProvider = NewDefaultCacheManagerProvider()
	s.Require().NotNil(provider)

	manager := provider.GetCacheManager()
	s.NotNil(manager)
}

// TestSetCacheManagerProvider tests setting a custom cache manager provider
func (s *BuildCacheTestSuite) TestSetCacheManagerProvider() {
	// Create a mock cache manager
	config := cache.DefaultConfig()
	config.Enabled = false // Disable for testing
	mockManager := cache.NewManager(config)

	// Create mock provider
	mockProvider := &MockCacheManagerProvider{
		manager: mockManager,
	}

	// Set the mock provider
	SetCacheManagerProvider(mockProvider)

	// Test that getCacheManager now returns the mock manager
	result := getCacheManager()
	s.Same(mockManager, result, "Should return the mock manager")
}

// TestBackwardCompatibility tests that getCacheManager maintains backward compatibility
func (s *BuildCacheTestSuite) TestBackwardCompatibility() {
	// Test default behavior
	manager1 := getCacheManager()
	manager2 := getCacheManager()

	s.NotNil(manager1)
	s.NotNil(manager2)
	s.Same(manager1, manager2, "Should maintain singleton behavior")
}

// TestThreadSafety tests thread-safe access to the cache manager
func (s *BuildCacheTestSuite) TestThreadSafety() {
	const numGoroutines = 10
	managers := make([]*cache.Manager, numGoroutines)
	done := make(chan struct{}, numGoroutines)

	// Launch multiple goroutines to access the cache manager
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer func() { done <- struct{}{} }()
			managers[index] = getCacheManager()
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all goroutines got the same instance
	for i := 1; i < numGoroutines; i++ {
		s.Same(managers[0], managers[i], "All goroutines should get the same instance")
	}
}

// TestProviderReset tests that setting a new provider works correctly
func (s *BuildCacheTestSuite) TestProviderReset() {
	// Get initial manager
	initialManager := getCacheManager()
	s.NotNil(initialManager)

	// Create and set a new provider
	config := cache.DefaultConfig()
	config.Enabled = false
	newManager := cache.NewManager(config)
	mockProvider := &MockCacheManagerProvider{manager: newManager}

	SetCacheManagerProvider(mockProvider)

	// Get manager again - should be different
	newResult := getCacheManager()
	s.Same(newManager, newResult, "Should return the new manager")
	s.NotSame(initialManager, newResult, "Should not return the initial manager")
}

// TestBuildContextWithCacheManager tests that getCacheManager uses the injected provider
func (s *BuildCacheTestSuite) TestBuildContextWithCacheManager() {
	// Create a mock cache manager
	config := cache.DefaultConfig()
	config.Enabled = false
	mockManager := cache.NewManager(config)
	mockProvider := &MockCacheManagerProvider{manager: mockManager}

	// Set the mock provider
	SetCacheManagerProvider(mockProvider)

	// Test that getCacheManager returns the mock manager
	result := getCacheManager()
	s.Same(mockManager, result, "getCacheManager should return the injected cache manager")

	// Test that multiple calls return the same instance
	result2 := getCacheManager()
	s.Same(mockManager, result2, "getCacheManager should consistently return the same instance")
	s.Same(result, result2, "Multiple calls should return identical instances")
}

// TestBuildNamespace runs the BuildCacheTestSuite
func TestBuildCache(t *testing.T) {
	suite.Run(t, new(BuildCacheTestSuite))
}
