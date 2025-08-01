package mage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
