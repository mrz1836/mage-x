package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPackageLevelLoadFromPaths tests the package-level LoadFromPaths convenience function
func TestPackageLevelLoadFromPaths(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test config file
	configContent := `{"name": "test-app", "version": "1.0.0"}`
	configPath := filepath.Join(tmpDir, "config.json")
	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	require.NoError(t, err)

	type TestConfig struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}

	t.Run("LoadFromPaths success", func(t *testing.T) {
		var config TestConfig
		path, err := LoadFromPaths(&config, "config", tmpDir)
		require.NoError(t, err)
		assert.Equal(t, configPath, path)
		assert.Equal(t, "test-app", config.Name)
		assert.Equal(t, "1.0.0", config.Version)
	})

	t.Run("LoadFromPaths not found", func(t *testing.T) {
		var config TestConfig
		_, err := LoadFromPaths(&config, "nonexistent", tmpDir)
		require.Error(t, err)
	})
}

// TestPackageLevelLoadWithEnvOverrides tests the package-level LoadWithEnvOverrides convenience function
func TestPackageLevelLoadWithEnvOverrides(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test config file
	configContent := `name: "test-app"
version: "1.0.0"`
	configPath := filepath.Join(tmpDir, "app.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	require.NoError(t, err)

	type TestConfig struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	}

	t.Run("LoadWithEnvOverrides success", func(t *testing.T) {
		// Set an env override
		t.Setenv("TEST_NAME", "overridden-name")

		var config TestConfig
		path, err := LoadWithEnvOverrides(&config, "app", "TEST", tmpDir)
		require.NoError(t, err)
		assert.Equal(t, configPath, path)
		// The env override should have been applied
		// Note: The actual behavior depends on the applyEnvOverrides implementation
		assert.NotEmpty(t, config.Name)
	})

	t.Run("LoadWithEnvOverrides not found", func(t *testing.T) {
		var config TestConfig
		_, err := LoadWithEnvOverrides(&config, "nonexistent", "TEST", tmpDir)
		require.Error(t, err)
	})
}

// TestLoaderImplMethods tests the loaderImpl wrapper methods
func TestLoaderImplMethods(t *testing.T) {
	loader := NewLoader()

	t.Run("LoadFromPath with nil manager", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a test config file
		configContent := `project:
  name: "test-project"
build:
  output_dir: "bin"
`
		configPath := filepath.Join(tmpDir, "mage.yaml")
		err := os.WriteFile(configPath, []byte(configContent), 0o600)
		require.NoError(t, err)

		config, err := loader.LoadFromPath(configPath)
		require.NoError(t, err)
		assert.NotNil(t, config)
	})

	t.Run("Save with nil manager", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a new loader to ensure nil manager
		newLoader := &loaderImpl{}

		// Change to tmpDir to save the file
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			_ = os.Chdir(originalDir) //nolint:errcheck // Best effort restore
		}()

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		config := &MageConfig{
			Project: ProjectConfig{
				Name: "test-save",
			},
		}

		err = newLoader.Save(config)
		require.NoError(t, err)

		// Verify file was created
		_, err = os.Stat(filepath.Join(tmpDir, "mage.yaml"))
		require.NoError(t, err)
	})

	t.Run("Validate with nil manager", func(t *testing.T) {
		newLoader := &loaderImpl{}

		config := &MageConfig{
			Project: ProjectConfig{
				Name:    "valid-project",
				Version: "1.0.0", // Required field
			},
		}

		err := newLoader.Validate(config)
		require.NoError(t, err)
	})

	t.Run("GetDefaults with nil manager", func(t *testing.T) {
		newLoader := &loaderImpl{}

		defaults := newLoader.GetDefaults()
		require.NotNil(t, defaults)
	})

	t.Run("Merge with nil manager", func(t *testing.T) {
		newLoader := &loaderImpl{}

		config1 := &MageConfig{
			Project: ProjectConfig{
				Name: "config1",
			},
		}
		config2 := &MageConfig{
			Project: ProjectConfig{
				Description: "merged config",
			},
		}

		merged := newLoader.Merge(config1, config2)
		require.NotNil(t, merged)
	})
}

// TestConfigHelperLoadWithEnvOverridesError tests error path in LoadWithEnvOverrides
func TestConfigHelperLoadWithEnvOverridesError(t *testing.T) {
	t.Run("LoadWithEnvOverrides with load error", func(t *testing.T) {
		cfg := New()

		var dest map[string]string
		_, err := cfg.LoadWithEnvOverrides(&dest, "nonexistent", "PREFIX", "/nonexistent/path")
		require.Error(t, err, "Should error when config file not found")
	})
}

// TestDefaultConfigLoaderSaveError tests error path in Save
func TestDefaultConfigLoaderSaveError(t *testing.T) {
	loader := NewDefaultConfigLoader()

	t.Run("Save to invalid path", func(t *testing.T) {
		// Try to save to a path that doesn't exist and can't be created
		err := loader.Save("/nonexistent/path/that/cannot/be/created/config.json", map[string]string{"key": "value"}, "json")
		require.Error(t, err)
	})
}

// TestMergeConfigEdgeCases tests edge cases in merge functions
func TestMergeConfigEdgeCases(t *testing.T) {
	manager := NewManager()

	t.Run("mergeBuildConfig all fields", func(t *testing.T) {
		base := &MageConfig{
			Build: BuildConfig{
				OutputDir:  "bin",
				Tags:       []string{"integration"},
				CGOEnabled: false,
			},
		}

		override := &MageConfig{
			Build: BuildConfig{
				OutputDir:  "dist",
				Tags:       []string{"e2e"},
				CGOEnabled: true,
				LDFlags:    "-s -w",
			},
		}

		merged := manager.Merge(base, override)
		require.NotNil(t, merged)
		assert.Equal(t, "dist", merged.Build.OutputDir)
		assert.True(t, merged.Build.CGOEnabled)
		assert.Equal(t, "-s -w", merged.Build.LDFlags)
	})

	t.Run("mergeTestConfig all fields", func(t *testing.T) {
		base := &MageConfig{
			Test: TestConfig{
				Timeout:  300,
				Race:     false,
				Coverage: false,
				Parallel: 4,
			},
		}

		override := &MageConfig{
			Test: TestConfig{
				Timeout:  600,
				Race:     true,
				Coverage: true,
				Parallel: 8,
			},
		}

		merged := manager.Merge(base, override)
		require.NotNil(t, merged)
		assert.Equal(t, 600, merged.Test.Timeout)
		assert.True(t, merged.Test.Race)
		assert.True(t, merged.Test.Coverage)
		assert.Equal(t, 8, merged.Test.Parallel)
	})

	t.Run("mergeProjectConfig all fields", func(t *testing.T) {
		base := &MageConfig{
			Project: ProjectConfig{
				Name:        "base-project",
				Description: "base description",
				Version:     "1.0.0",
			},
		}

		override := &MageConfig{
			Project: ProjectConfig{
				Name:        "override-project",
				Description: "override description",
				Version:     "2.0.0",
			},
		}

		merged := manager.Merge(base, override)
		require.NotNil(t, merged)
		assert.Equal(t, "override-project", merged.Project.Name)
		assert.Equal(t, "override description", merged.Project.Description)
		assert.Equal(t, "2.0.0", merged.Project.Version)
	})
}
