package mage

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func TestConfigProvider(t *testing.T) {
	t.Run("default provider loads config", func(t *testing.T) {
		provider := NewDefaultConfigProvider()
		cfg, err := provider.GetConfig()
		require.NoError(t, err)
		require.NotNil(t, cfg)

		// Should return same config on second call
		cfg2, err := provider.GetConfig()
		require.NoError(t, err)
		assert.Same(t, cfg, cfg2)
	})

	t.Run("set config overrides loading", func(t *testing.T) {
		provider := NewDefaultConfigProvider()
		testConfig := &Config{
			Project: ProjectConfig{
				Name: "test-project",
			},
		}

		provider.SetConfig(testConfig)
		cfg, err := provider.GetConfig()
		require.NoError(t, err)
		assert.Equal(t, "test-project", cfg.Project.Name)
	})

	t.Run("reset config clears state", func(t *testing.T) {
		provider := NewDefaultConfigProvider()
		testConfig := &Config{
			Project: ProjectConfig{
				Name: "test-project",
			},
		}

		provider.SetConfig(testConfig)
		provider.ResetConfig()

		cfg, err := provider.GetConfig()
		require.NoError(t, err)
		// After reset, should get default config
		assert.NotEqual(t, "test-project", cfg.Project.Name)
	})

	t.Run("mock provider returns configured values", func(t *testing.T) {
		mock := &MockConfigProvider{
			Config: &Config{
				Project: ProjectConfig{
					Name: "mock-project",
				},
			},
		}

		cfg, err := mock.GetConfig()
		require.NoError(t, err)
		assert.Equal(t, "mock-project", cfg.Project.Name)
	})

	t.Run("mock provider returns errors", func(t *testing.T) {
		mock := &MockConfigProvider{
			Err: ErrMockConfigFailure,
		}

		cfg, err := mock.GetConfig()
		require.Error(t, err)
		assert.Nil(t, cfg)
		assert.Equal(t, ErrMockConfigFailure, err)
	})

	t.Run("set and get config provider", func(t *testing.T) {
		// Save original provider
		originalProvider := GetConfigProvider()
		defer SetConfigProvider(originalProvider)

		// Set mock provider
		mock := &MockConfigProvider{
			Config: &Config{
				Project: ProjectConfig{
					Name: "provider-test",
				},
			},
		}
		SetConfigProvider(mock)

		// GetConfig should use the mock provider
		cfg, err := GetConfig()
		require.NoError(t, err)
		assert.Equal(t, "provider-test", cfg.Project.Name)
	})
}

func TestTestHelpers(t *testing.T) {
	t.Run("TestSetConfig sets config through provider", func(t *testing.T) {
		// Save original state
		originalProvider := GetConfigProvider()
		defer SetConfigProvider(originalProvider)
		defer TestResetConfig()

		testConfig := &Config{
			Project: ProjectConfig{
				Name: "test-helper-project",
			},
		}

		TestSetConfig(testConfig)

		cfg, err := GetConfig()
		require.NoError(t, err)
		assert.Equal(t, "test-helper-project", cfg.Project.Name)
	})

	t.Run("TestResetConfig resets provider", func(t *testing.T) {
		// Save original state
		originalProvider := GetConfigProvider()
		defer SetConfigProvider(originalProvider)

		TestSetConfig(&Config{
			Project: ProjectConfig{
				Name: "to-be-reset",
			},
		})

		TestResetConfig()

		cfg, err := GetConfig()
		require.NoError(t, err)
		assert.NotEqual(t, "to-be-reset", cfg.Project.Name)
	})
}

func TestPackageProviderRegistryProvider(t *testing.T) {
	t.Run("default package provider registry provider creates registry", func(t *testing.T) {
		provider := NewDefaultPackageProviderRegistryProvider()
		registry := provider.GetRegistry()
		require.NotNil(t, registry)

		// Should return same registry on second call
		registry2 := provider.GetRegistry()
		assert.Same(t, registry, registry2)
	})

	t.Run("set registry overrides default", func(t *testing.T) {
		provider := NewDefaultPackageProviderRegistryProvider()
		customRegistry := NewProviderRegistry()

		provider.SetRegistry(customRegistry)
		registry := provider.GetRegistry()
		assert.Same(t, customRegistry, registry)
	})

	t.Run("reset registry creates new instance", func(t *testing.T) {
		provider := NewDefaultPackageProviderRegistryProvider()
		originalRegistry := provider.GetRegistry()

		provider.ResetRegistry()
		newRegistry := provider.GetRegistry()
		assert.NotSame(t, originalRegistry, newRegistry)
	})

	t.Run("package provider registry provider is thread-safe", func(t *testing.T) {
		provider := NewDefaultPackageProviderRegistryProvider()
		const numGoroutines = 100
		var wg sync.WaitGroup
		registries := make([]*ProviderRegistry, numGoroutines)

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				defer wg.Done()
				registries[index] = provider.GetRegistry()
			}(i)
		}
		wg.Wait()

		// All goroutines should get the same registry instance
		for i := 1; i < numGoroutines; i++ {
			assert.Same(t, registries[0], registries[i], "Registry instance should be same across goroutines")
		}
	})

	t.Run("concurrent set and get operations are thread-safe", func(t *testing.T) {
		provider := NewDefaultPackageProviderRegistryProvider()
		const numGoroutines = 50
		var wg sync.WaitGroup
		customRegistries := make([]*ProviderRegistry, numGoroutines)

		// Create custom registries
		for i := 0; i < numGoroutines; i++ {
			customRegistries[i] = NewProviderRegistry()
		}

		wg.Add(numGoroutines * 2) // Double for set and get operations

		// Concurrent set operations
		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				defer wg.Done()
				provider.SetRegistry(customRegistries[index])
			}(i)
		}

		// Concurrent get operations
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				registry := provider.GetRegistry()
				assert.NotNil(t, registry)
			}()
		}

		wg.Wait()

		// Final registry should be one of the custom registries
		finalRegistry := provider.GetRegistry()
		found := false
		for _, customRegistry := range customRegistries {
			if finalRegistry == customRegistry {
				found = true
				break
			}
		}
		assert.True(t, found, "Final registry should be one of the custom registries")
	})
}

func TestPackageProviderRegistryProviderIntegration(t *testing.T) {
	t.Run("package level functions work with custom provider", func(t *testing.T) {
		// Save original state
		originalProvider := getPackageProviderRegistryProvider()
		defer func() {
			// Restore original provider
			SetPackageProviderRegistryProvider(originalProvider)
		}()

		// Create custom provider with mock config
		customProvider := NewDefaultPackageProviderRegistryProvider()
		customRegistry := NewProviderRegistry()
		mockProvider := &MockConfigProvider{
			Config: &Config{
				Project: ProjectConfig{
					Name: "custom-provider-test",
				},
			},
		}
		customRegistry.SetProvider(mockProvider)
		customProvider.SetRegistry(customRegistry)

		// Set custom provider
		SetPackageProviderRegistryProvider(customProvider)

		// Test that package-level functions use the custom provider
		cfg, err := GetConfig()
		require.NoError(t, err)
		assert.Equal(t, "custom-provider-test", cfg.Project.Name)
	})

	t.Run("test reset config works with new provider pattern", func(t *testing.T) {
		// Save original state
		originalProvider := getPackageProviderRegistryProvider()
		defer func() {
			// Restore original provider
			SetPackageProviderRegistryProvider(originalProvider)
		}()

		// Set a test config
		TestSetConfig(&Config{
			Project: ProjectConfig{
				Name: "test-before-reset",
			},
		})

		cfg, err := GetConfig()
		require.NoError(t, err)
		assert.Equal(t, "test-before-reset", cfg.Project.Name)

		// Reset config
		TestResetConfig()

		// Should get default config after reset
		cfg, err = GetConfig()
		require.NoError(t, err)
		assert.NotEqual(t, "test-before-reset", cfg.Project.Name)
	})
}

// TestSuite for comprehensive testing
type ConfigProviderTestSuite struct {
	suite.Suite

	originalProvider PackageProviderRegistryProvider
}

func (suite *ConfigProviderTestSuite) SetupTest() {
	// Save the original provider
	suite.originalProvider = getPackageProviderRegistryProvider()
}

func (suite *ConfigProviderTestSuite) TearDownTest() {
	// Restore the original provider
	SetPackageProviderRegistryProvider(suite.originalProvider)
}

func (suite *ConfigProviderTestSuite) TestProviderIsolation() {
	// Test that different provider instances are isolated
	provider1 := NewDefaultPackageProviderRegistryProvider()
	provider2 := NewDefaultPackageProviderRegistryProvider()

	registry1 := provider1.GetRegistry()
	registry2 := provider2.GetRegistry()

	suite.NotSame(registry1, registry2, "Different providers should have different registries")

	// Set different configs
	mock1 := &MockConfigProvider{
		Config: &Config{Project: ProjectConfig{Name: "provider1"}},
	}
	mock2 := &MockConfigProvider{
		Config: &Config{Project: ProjectConfig{Name: "provider2"}},
	}

	registry1.SetProvider(mock1)
	registry2.SetProvider(mock2)

	cfg1, err := registry1.GetProvider().GetConfig()
	suite.Require().NoError(err)
	suite.Equal("provider1", cfg1.Project.Name)

	cfg2, err := registry2.GetProvider().GetConfig()
	suite.Require().NoError(err)
	suite.Equal("provider2", cfg2.Project.Name)
}

func (suite *ConfigProviderTestSuite) TestConcurrentProviderAccess() {
	// Test concurrent access to the package-level provider
	const numGoroutines = 100
	var wg sync.WaitGroup
	results := make([]PackageProviderRegistryProvider, numGoroutines)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			defer wg.Done()
			// Small delay to increase chance of race conditions
			time.Sleep(time.Microsecond * 10)
			results[index] = getPackageProviderRegistryProvider()
		}(i)
	}
	wg.Wait()

	// All results should be the same provider instance
	for i := 1; i < numGoroutines; i++ {
		suite.Same(results[0], results[i], "All goroutines should get the same provider instance")
	}
}

func (suite *ConfigProviderTestSuite) TestProviderRegistryProviderBackwardCompatibility() {
	// Test that the new pattern maintains backward compatibility

	// Test setting a custom config provider
	mockProvider := &MockConfigProvider{
		Config: &Config{
			Project: ProjectConfig{
				Name: "backward-compatibility-test",
			},
		},
	}

	SetConfigProvider(mockProvider)

	// Should be able to get the config using the old API
	cfg, err := GetConfig()
	suite.Require().NoError(err)
	suite.Equal("backward-compatibility-test", cfg.Project.Name)

	// Should be able to get the provider using the old API
	provider := GetConfigProvider()
	suite.Same(mockProvider, provider)
}

func TestConfigProviderTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigProviderTestSuite))
}
