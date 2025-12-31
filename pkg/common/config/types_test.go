package config

import (
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/common/env"
)

// cfgWithGoVersion creates a valid MageConfig with the specified Go version.
// Used to test Go version validation edge cases.
func cfgWithGoVersion(v string) *MageConfig {
	return &MageConfig{
		Project: ProjectConfig{Name: "test-project", Version: "1.0.0"},
		Build:   BuildConfig{GoVersion: v},
	}
}

// cfgWithTimeout creates a valid MageConfig with the specified test timeout.
// Used to test timeout validation edge cases.
func cfgWithTimeout(t int) *MageConfig {
	return &MageConfig{
		Project: ProjectConfig{Name: "test-project", Version: "1.0.0"},
		Test:    TestConfig{Timeout: t},
	}
}

// TestConfigManagerValidate tests the Validate method of configManagerImpl.
// Validates that configuration validation catches all required fields and
// enforces constraints on optional fields.
func TestConfigManagerValidate(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name    string
		config  *MageConfig
		wantErr error
	}{
		// Nil config - critical failure case
		{
			name:    "nil config returns errConfigNil",
			config:  nil,
			wantErr: errConfigNil,
		},

		// Required field: Project.Name
		{
			name: "empty project name returns errProjectNameRequired",
			config: &MageConfig{
				Project: ProjectConfig{Version: "1.0.0"},
			},
			wantErr: errProjectNameRequired,
		},

		// Required field: Project.Version
		{
			name: "empty project version returns errProjectVersionReq",
			config: &MageConfig{
				Project: ProjectConfig{Name: "test"},
			},
			wantErr: errProjectVersionReq,
		},

		// Valid minimal config
		{
			name: "valid minimal config passes",
			config: &MageConfig{
				Project: ProjectConfig{Name: "test", Version: "1.0.0"},
			},
			wantErr: nil,
		},

		// Go version validation: too short (< 4 chars)
		{
			name:    "go version too short (1.2) returns errInvalidGoVersion",
			config:  cfgWithGoVersion("1.2"),
			wantErr: errInvalidGoVersion,
		},
		{
			name:    "go version too short (1) returns errInvalidGoVersion",
			config:  cfgWithGoVersion("1"),
			wantErr: errInvalidGoVersion,
		},
		{
			name:    "go version exactly 3 chars returns errInvalidGoVersion",
			config:  cfgWithGoVersion("1.x"),
			wantErr: errInvalidGoVersion,
		},

		// Go version validation: wrong prefix (must start with "1.")
		{
			name:    "go version wrong prefix (2.20) returns errInvalidGoVersion",
			config:  cfgWithGoVersion("2.20"),
			wantErr: errInvalidGoVersion,
		},
		{
			name:    "go version wrong prefix (go1.20) returns errInvalidGoVersion",
			config:  cfgWithGoVersion("go1.20"),
			wantErr: errInvalidGoVersion,
		},
		{
			name:    "go version wrong prefix (v1.20) returns errInvalidGoVersion",
			config:  cfgWithGoVersion("v1.20"),
			wantErr: errInvalidGoVersion,
		},
		{
			name:    "go version starts with 0. returns errInvalidGoVersion",
			config:  cfgWithGoVersion("0.20"),
			wantErr: errInvalidGoVersion,
		},

		// Go version validation: valid formats
		{
			name:    "go version valid short (1.20) passes",
			config:  cfgWithGoVersion("1.20"),
			wantErr: nil,
		},
		{
			name:    "go version valid with patch (1.24.1) passes",
			config:  cfgWithGoVersion("1.24.1"),
			wantErr: nil,
		},
		{
			name:    "go version valid long (1.24.10) passes",
			config:  cfgWithGoVersion("1.24.10"),
			wantErr: nil,
		},
		{
			name:    "go version empty (optional) passes",
			config:  cfgWithGoVersion(""),
			wantErr: nil,
		},

		// Test timeout validation: negative values
		{
			name:    "timeout negative (-1) returns errTestTimeoutNegative",
			config:  cfgWithTimeout(-1),
			wantErr: errTestTimeoutNegative,
		},
		{
			name:    "timeout negative (MinInt) returns errTestTimeoutNegative",
			config:  cfgWithTimeout(math.MinInt),
			wantErr: errTestTimeoutNegative,
		},

		// Test timeout validation: boundary values
		{
			name:    "timeout zero passes",
			config:  cfgWithTimeout(0),
			wantErr: nil,
		},
		{
			name:    "timeout positive (120) passes",
			config:  cfgWithTimeout(120),
			wantErr: nil,
		},
		{
			name:    "timeout MaxInt passes",
			config:  cfgWithTimeout(math.MaxInt),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.Validate(tt.config)

			if tt.wantErr != nil {
				require.Error(t, err, "expected error for test case: %s", tt.name)
				assert.ErrorIs(t, err, tt.wantErr,
					"expected error %v, got %v", tt.wantErr, err)
			} else {
				assert.NoError(t, err, "expected no error for test case: %s", tt.name)
			}
		})
	}
}

// TestConfigManagerMerge tests the Merge method of configManagerImpl.
// Validates merge order sensitivity, deep copy behavior, and override logic.
func TestConfigManagerMerge(t *testing.T) {
	manager := NewManager()

	t.Run("empty input returns defaults", func(t *testing.T) {
		result := manager.Merge()

		defaults := manager.GetDefaults()
		assert.Equal(t, defaults.Project.Name, result.Project.Name)
		assert.Equal(t, defaults.Project.Version, result.Project.Version)
		assert.Equal(t, defaults.Build.Platform, result.Build.Platform)
	})

	t.Run("single config returns copy not same pointer", func(t *testing.T) {
		original := &MageConfig{
			Project: ProjectConfig{Name: "original", Version: "1.0.0"},
		}

		result := manager.Merge(original)

		assert.NotSame(t, original, result, "Merge should return new pointer")
		assert.Equal(t, original.Project.Name, result.Project.Name)

		// Modifying result should not affect original
		result.Project.Name = "modified"
		assert.Equal(t, "original", original.Project.Name,
			"modifying result should not affect original")
	})

	t.Run("two configs later overrides earlier", func(t *testing.T) {
		base := &MageConfig{
			Project: ProjectConfig{Name: "base", Version: "1.0.0", Description: "base desc"},
			Build:   BuildConfig{GoVersion: "1.20", Platform: "linux/amd64"},
		}
		override := &MageConfig{
			Project: ProjectConfig{Name: "override", Version: "2.0.0"},
			Build:   BuildConfig{GoVersion: "1.24"},
		}

		result := manager.Merge(base, override)

		assert.Equal(t, "override", result.Project.Name, "name should be overridden")
		assert.Equal(t, "2.0.0", result.Project.Version, "version should be overridden")
		assert.Equal(t, "base desc", result.Project.Description,
			"description should be preserved from base")
		assert.Equal(t, "1.24", result.Build.GoVersion, "GoVersion should be overridden")
		assert.Equal(t, "linux/amd64", result.Build.Platform,
			"Platform should be preserved from base")
	})

	t.Run("three configs cascading override", func(t *testing.T) {
		config1 := &MageConfig{
			Project: ProjectConfig{Name: "first", Version: "1.0.0"},
		}
		config2 := &MageConfig{
			Project: ProjectConfig{Name: "second"},
		}
		config3 := &MageConfig{
			Project: ProjectConfig{Name: "third"},
		}

		result := manager.Merge(config1, config2, config3)

		assert.Equal(t, "third", result.Project.Name, "last non-empty value should win")
		assert.Equal(t, "1.0.0", result.Project.Version, "first version should persist")
	})

	t.Run("slice fields are deep-copied", func(t *testing.T) {
		original := &MageConfig{
			Project: ProjectConfig{
				Name:    "test",
				Version: "1.0.0",
				Authors: []string{"author1", "author2"},
			},
			Build: BuildConfig{
				GoVersion: "1.24",
				Tags:      []string{"tag1", "tag2"},
			},
			Test: TestConfig{
				Tags: []string{"test-tag1"},
			},
		}

		result := manager.Merge(original)

		// Verify slices are equal
		assert.Equal(t, original.Project.Authors, result.Project.Authors)
		assert.Equal(t, original.Build.Tags, result.Build.Tags)
		assert.Equal(t, original.Test.Tags, result.Test.Tags)

		// Verify slices are not shared references
		if len(result.Project.Authors) > 0 {
			result.Project.Authors[0] = "modified"
			assert.Equal(t, "author1", original.Project.Authors[0],
				"modifying result slice should not affect original")
		}
	})

	t.Run("empty override preserves base values", func(t *testing.T) {
		base := &MageConfig{
			Project: ProjectConfig{Name: "base", Version: "1.0.0"},
			Build:   BuildConfig{GoVersion: "1.24", Platform: "linux/amd64"},
		}
		empty := &MageConfig{}

		result := manager.Merge(base, empty)

		assert.Equal(t, "base", result.Project.Name)
		assert.Equal(t, "1.0.0", result.Project.Version)
		assert.Equal(t, "1.24", result.Build.GoVersion)
	})

	t.Run("boolean fields only override when other fields present", func(t *testing.T) {
		base := &MageConfig{
			Project: ProjectConfig{Name: "test", Version: "1.0.0"},
			Build: BuildConfig{
				GoVersion:  "1.24",
				CGOEnabled: true,
			},
		}
		// Override with only boolean change (no other build fields)
		boolOnlyOverride := &MageConfig{
			Build: BuildConfig{
				CGOEnabled: false,
			},
		}

		result := manager.Merge(base, boolOnlyOverride)

		// CGOEnabled should NOT be overridden because hasBuildOverrides returns false
		// (no GoVersion, Platform, or Tags set in override)
		assert.True(t, result.Build.CGOEnabled,
			"boolean should not override when no other build fields present")

		// Now test with other fields present
		withOtherFields := &MageConfig{
			Build: BuildConfig{
				GoVersion:  "1.20", // This triggers hasBuildOverrides = true
				CGOEnabled: false,
			},
		}

		result2 := manager.Merge(base, withOtherFields)
		assert.False(t, result2.Build.CGOEnabled,
			"boolean should override when other build fields present")
	})

	t.Run("analytics config replaces entirely", func(t *testing.T) {
		base := &MageConfig{
			Project: ProjectConfig{Name: "test", Version: "1.0.0"},
			Analytics: AnalyticsConfig{
				Enabled:       true,
				SampleRate:    0.5,
				RetentionDays: 30,
				BatchSize:     100,
			},
		}
		override := &MageConfig{
			Analytics: AnalyticsConfig{
				SampleRate: 0.8, // Only set sample rate
			},
		}

		result := manager.Merge(base, override)

		// Analytics should be fully replaced, not merged field-by-field
		assert.InDelta(t, 0.8, result.Analytics.SampleRate, 0.001)
		// Other fields should be zero values from override, not preserved from base
		assert.False(t, result.Analytics.Enabled,
			"Enabled should be zero value from override")
		assert.Equal(t, 0, result.Analytics.RetentionDays,
			"RetentionDays should be zero value from override")
	})

	t.Run("project authors slice override", func(t *testing.T) {
		base := &MageConfig{
			Project: ProjectConfig{
				Name:    "test",
				Version: "1.0.0",
				Authors: []string{"author1", "author2"},
			},
		}
		override := &MageConfig{
			Project: ProjectConfig{
				Authors: []string{"new-author"},
			},
		}

		result := manager.Merge(base, override)

		assert.Equal(t, []string{"new-author"}, result.Project.Authors,
			"authors should be replaced, not appended")
	})
}

// TestCleanEnvValue tests the env.CleanValue helper function.
// This function removes inline comments and trims whitespace from env values.
func TestCleanEnvValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string returns empty",
			input:    "",
			expected: "",
		},
		{
			name:     "plain value unchanged",
			input:    "value",
			expected: "value",
		},
		{
			name:     "removes inline comment with space",
			input:    "value # comment",
			expected: "value",
		},
		{
			name:     "trims leading whitespace",
			input:    "  value",
			expected: "value",
		},
		{
			name:     "trims trailing whitespace",
			input:    "value  ",
			expected: "value",
		},
		{
			name:     "trims both whitespace",
			input:    "  value  ",
			expected: "value",
		},
		{
			name:     "handles whitespace and comment",
			input:    "  value # comment  ",
			expected: "value",
		},
		{
			name:     "hash without space not treated as comment",
			input:    "value#nocomment",
			expected: "value#nocomment",
		},
		{
			name:     "multiple spaces before hash",
			input:    "value   # comment",
			expected: "value",
		},
		{
			name:     "comment at start with space",
			input:    " # full comment",
			expected: "",
		},
		{
			name:     "preserves internal hash without space",
			input:    "value1#value2 # real comment",
			expected: "value1#value2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := env.CleanValue(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetDefaultGoVersion tests the getDefaultGoVersion function.
// Validates priority: MAGE_X_GO_VERSION > GO_PRIMARY_VERSION > fallback "1.24"
func TestGetDefaultGoVersion(t *testing.T) {
	// Save original env values and restore after tests
	origMageXGoVersion := os.Getenv("MAGE_X_GO_VERSION")
	origGoPrimaryVersion := os.Getenv("GO_PRIMARY_VERSION")

	t.Cleanup(func() {
		restoreEnv(t, "MAGE_X_GO_VERSION", origMageXGoVersion)
		restoreEnv(t, "GO_PRIMARY_VERSION", origGoPrimaryVersion)
	})

	clearEnvVars := func(t *testing.T) {
		t.Helper()
		require.NoError(t, os.Unsetenv("MAGE_X_GO_VERSION"))
		require.NoError(t, os.Unsetenv("GO_PRIMARY_VERSION"))
	}

	t.Run("returns fallback when no env vars set", func(t *testing.T) {
		clearEnvVars(t)

		result := getDefaultGoVersion()

		assert.Equal(t, "1.24", result, "should return fallback version")
	})

	t.Run("MAGE_X_GO_VERSION takes priority", func(t *testing.T) {
		clearEnvVars(t)
		require.NoError(t, os.Setenv("MAGE_X_GO_VERSION", "1.22"))
		require.NoError(t, os.Setenv("GO_PRIMARY_VERSION", "1.20"))

		result := getDefaultGoVersion()

		assert.Equal(t, "1.22", result, "MAGE_X_GO_VERSION should take priority")
	})

	t.Run("GO_PRIMARY_VERSION used when MAGE_X_GO_VERSION empty", func(t *testing.T) {
		clearEnvVars(t)
		require.NoError(t, os.Setenv("GO_PRIMARY_VERSION", "1.21"))

		result := getDefaultGoVersion()

		assert.Equal(t, "1.21", result)
	})

	t.Run("strips .x suffix from MAGE_X_GO_VERSION", func(t *testing.T) {
		clearEnvVars(t)
		require.NoError(t, os.Setenv("MAGE_X_GO_VERSION", "1.24.x"))

		result := getDefaultGoVersion()

		assert.Equal(t, "1.24", result, "should strip .x suffix")
	})

	t.Run("strips .x suffix from GO_PRIMARY_VERSION", func(t *testing.T) {
		clearEnvVars(t)
		require.NoError(t, os.Setenv("GO_PRIMARY_VERSION", "1.23.x"))

		result := getDefaultGoVersion()

		assert.Equal(t, "1.23", result, "should strip .x suffix")
	})

	t.Run("handles inline comment in MAGE_X_GO_VERSION", func(t *testing.T) {
		clearEnvVars(t)
		require.NoError(t, os.Setenv("MAGE_X_GO_VERSION", "1.24 # primary version"))

		result := getDefaultGoVersion()

		assert.Equal(t, "1.24", result, "should clean comment from value")
	})

	t.Run("short version without .x suffix unchanged", func(t *testing.T) {
		clearEnvVars(t)
		require.NoError(t, os.Setenv("MAGE_X_GO_VERSION", "1.24"))

		result := getDefaultGoVersion()

		assert.Equal(t, "1.24", result)
	})

	t.Run("version too short for .x stripping", func(t *testing.T) {
		clearEnvVars(t)
		require.NoError(t, os.Setenv("MAGE_X_GO_VERSION", "1x"))

		result := getDefaultGoVersion()

		// Length is 2, which is not > 2, so .x check doesn't apply
		assert.Equal(t, "1x", result)
	})
}

// restoreEnv is a helper to restore environment variable to original state
func restoreEnv(t *testing.T, key, value string) {
	t.Helper()
	if value == "" {
		if err := os.Unsetenv(key); err != nil {
			t.Logf("Failed to unset %s: %v", key, err)
		}
	} else {
		if err := os.Setenv(key, value); err != nil {
			t.Logf("Failed to set %s: %v", key, err)
		}
	}
}

// TestNewManager tests the NewManager factory function
func TestNewManager(t *testing.T) {
	t.Run("returns non-nil manager", func(t *testing.T) {
		manager := NewManager()
		assert.NotNil(t, manager)
	})

	t.Run("implements MageConfigManager interface", func(t *testing.T) {
		_ = NewManager()
	})
}

// TestNewLoader tests the NewLoader factory function
func TestNewLoader(t *testing.T) {
	t.Run("returns non-nil loader", func(t *testing.T) {
		loader := NewLoader()
		assert.NotNil(t, loader)
	})

	t.Run("implements MageLoader interface", func(t *testing.T) {
		_ = NewLoader()
	})
}

// TestLoaderImpl tests the loaderImpl methods
func TestLoaderImpl(t *testing.T) {
	loader := NewLoader()

	t.Run("Load returns defaults", func(t *testing.T) {
		config, err := loader.Load()

		require.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "mage-project", config.Project.Name)
	})

	t.Run("GetDefaults returns valid config", func(t *testing.T) {
		defaults := loader.GetDefaults()

		assert.NotNil(t, defaults)
		assert.NotEmpty(t, defaults.Project.Name)
		assert.NotEmpty(t, defaults.Project.Version)
	})

	t.Run("Validate delegates to manager", func(t *testing.T) {
		validConfig := &MageConfig{
			Project: ProjectConfig{Name: "test", Version: "1.0.0"},
		}

		err := loader.Validate(validConfig)
		require.NoError(t, err)

		invalidConfig := &MageConfig{}
		err = loader.Validate(invalidConfig)
		assert.Error(t, err)
	})

	t.Run("Merge delegates to manager", func(t *testing.T) {
		config1 := &MageConfig{
			Project: ProjectConfig{Name: "first", Version: "1.0.0"},
		}
		config2 := &MageConfig{
			Project: ProjectConfig{Name: "second"},
		}

		result := loader.Merge(config1, config2)

		assert.Equal(t, "second", result.Project.Name)
		assert.Equal(t, "1.0.0", result.Project.Version)
	})
}

// TestConfigManagerGetDefaults tests GetDefaults returns expected default values
func TestConfigManagerGetDefaults(t *testing.T) {
	manager := NewManager()
	defaults := manager.GetDefaults()

	t.Run("has required project fields", func(t *testing.T) {
		assert.Equal(t, "mage-project", defaults.Project.Name)
		assert.Equal(t, "1.0.0", defaults.Project.Version)
	})

	t.Run("has build defaults", func(t *testing.T) {
		assert.NotEmpty(t, defaults.Build.GoVersion)
		assert.Equal(t, "linux/amd64", defaults.Build.Platform)
		assert.False(t, defaults.Build.CGOEnabled)
	})

	t.Run("has test defaults", func(t *testing.T) {
		assert.Equal(t, 120, defaults.Test.Timeout)
		assert.True(t, defaults.Test.Coverage)
		assert.Equal(t, 4, defaults.Test.Parallel)
	})

	t.Run("has analytics defaults", func(t *testing.T) {
		assert.False(t, defaults.Analytics.Enabled)
		assert.InDelta(t, 0.1, defaults.Analytics.SampleRate, 0.001)
	})
}

// TestConfigManagerLoadFromPath tests LoadFromPath with error cases
func TestConfigManagerLoadFromPath(t *testing.T) {
	manager := NewManager()

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := manager.LoadFromPath("/nonexistent/path/config.yaml")
		require.Error(t, err)
	})

	t.Run("returns error for relative path", func(t *testing.T) {
		_, err := manager.LoadFromPath("relative/path.yaml")
		require.Error(t, err)
	})

	t.Run("loads valid config from absolute path", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := tmpDir + "/test-config.yaml"

		// Create a valid config file
		configContent := `project:
  name: test-project
  version: 1.0.0
build:
  go_version: "1.24"
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0o600))

		config, err := manager.LoadFromPath(configPath)
		require.NoError(t, err)
		assert.Equal(t, "test-project", config.Project.Name)
		assert.Equal(t, "1.0.0", config.Project.Version)
	})
}

// TestConfigManagerSave tests Save with various scenarios
func TestConfigManagerSave(t *testing.T) {
	manager := NewManager()

	t.Run("saves config as YAML", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := tmpDir + "/output.yaml"

		config := &MageConfig{
			Project: ProjectConfig{Name: "test", Version: "1.0.0"},
			Build:   BuildConfig{GoVersion: "1.24"},
		}

		err := manager.Save(config, configPath)
		require.NoError(t, err)

		// Verify file exists
		_, statErr := os.Stat(configPath)
		require.NoError(t, statErr)

		// Load and verify content
		loaded, loadErr := manager.LoadFromPath(configPath)
		require.NoError(t, loadErr)
		assert.Equal(t, "test", loaded.Project.Name)
	})

	t.Run("saves config as JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := tmpDir + "/output.json"

		config := &MageConfig{
			Project: ProjectConfig{Name: "json-test", Version: "2.0.0"},
		}

		err := manager.Save(config, configPath)
		require.NoError(t, err)

		// Verify file exists and is valid
		_, statErr := os.Stat(configPath)
		require.NoError(t, statErr)
	})

	t.Run("saves config with yml extension", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := tmpDir + "/output.yml"

		config := &MageConfig{
			Project: ProjectConfig{Name: "yml-test", Version: "1.0.0"},
		}

		err := manager.Save(config, configPath)
		require.NoError(t, err)

		_, statErr := os.Stat(configPath)
		require.NoError(t, statErr)
	})

	t.Run("defaults to YAML for unknown extension", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := tmpDir + "/output.conf"

		config := &MageConfig{
			Project: ProjectConfig{Name: "default-test", Version: "1.0.0"},
		}

		err := manager.Save(config, configPath)
		require.NoError(t, err)

		_, statErr := os.Stat(configPath)
		require.NoError(t, statErr)
	})
}

// TestMergeProjectConfigComprehensive tests all project config merge scenarios
func TestMergeProjectConfigComprehensive(t *testing.T) {
	manager := NewManager()

	t.Run("all project fields override", func(t *testing.T) {
		base := &MageConfig{
			Project: ProjectConfig{
				Name:        "base",
				Version:     "1.0.0",
				Description: "base desc",
				License:     "MIT",
				Homepage:    "https://base.com",
				Repository:  "https://github.com/base/repo",
				Authors:     []string{"author1"},
			},
		}
		override := &MageConfig{
			Project: ProjectConfig{
				Name:        "override",
				Version:     "2.0.0",
				Description: "override desc",
				License:     "Apache-2.0",
				Homepage:    "https://override.com",
				Repository:  "https://github.com/override/repo",
				Authors:     []string{"author2", "author3"},
			},
		}

		result := manager.Merge(base, override)

		assert.Equal(t, "override", result.Project.Name)
		assert.Equal(t, "2.0.0", result.Project.Version)
		assert.Equal(t, "override desc", result.Project.Description)
		assert.Equal(t, "Apache-2.0", result.Project.License)
		assert.Equal(t, "https://override.com", result.Project.Homepage)
		assert.Equal(t, "https://github.com/override/repo", result.Project.Repository)
		assert.Equal(t, []string{"author2", "author3"}, result.Project.Authors)
	})

	t.Run("partial project override preserves unset fields", func(t *testing.T) {
		base := &MageConfig{
			Project: ProjectConfig{
				Name:        "base",
				Version:     "1.0.0",
				Description: "base desc",
				License:     "MIT",
				Homepage:    "https://base.com",
			},
		}
		override := &MageConfig{
			Project: ProjectConfig{
				Name: "override",
				// Version empty - should preserve base
				// Description empty - should preserve base
			},
		}

		result := manager.Merge(base, override)

		assert.Equal(t, "override", result.Project.Name)
		assert.Equal(t, "1.0.0", result.Project.Version)
		assert.Equal(t, "base desc", result.Project.Description)
		assert.Equal(t, "MIT", result.Project.License)
	})
}

// TestMergeBuildConfigComprehensive tests all build config merge scenarios
func TestMergeBuildConfigComprehensive(t *testing.T) {
	manager := NewManager()

	t.Run("all build fields override", func(t *testing.T) {
		base := &MageConfig{
			Project: ProjectConfig{Name: "test", Version: "1.0.0"},
			Build: BuildConfig{
				GoVersion:  "1.20",
				Platform:   "linux/amd64",
				Tags:       []string{"tag1"},
				LDFlags:    "-s -w",
				GCFlags:    "-N -l",
				CGOEnabled: true,
				OutputDir:  "/base/output",
				Binary:     "base-binary",
			},
		}
		override := &MageConfig{
			Build: BuildConfig{
				GoVersion:  "1.24",
				Platform:   "darwin/arm64",
				Tags:       []string{"tag2", "tag3"},
				LDFlags:    "-X main.version=2.0",
				GCFlags:    "-m",
				CGOEnabled: false,
				OutputDir:  "/override/output",
				Binary:     "override-binary",
			},
		}

		result := manager.Merge(base, override)

		assert.Equal(t, "1.24", result.Build.GoVersion)
		assert.Equal(t, "darwin/arm64", result.Build.Platform)
		assert.Equal(t, []string{"tag2", "tag3"}, result.Build.Tags)
		assert.Equal(t, "-X main.version=2.0", result.Build.LDFlags)
		assert.Equal(t, "-m", result.Build.GCFlags)
		assert.False(t, result.Build.CGOEnabled)
		assert.Equal(t, "/override/output", result.Build.OutputDir)
		assert.Equal(t, "override-binary", result.Build.Binary)
	})

	t.Run("partial build override preserves unset fields", func(t *testing.T) {
		base := &MageConfig{
			Project: ProjectConfig{Name: "test", Version: "1.0.0"},
			Build: BuildConfig{
				GoVersion:  "1.20",
				Platform:   "linux/amd64",
				LDFlags:    "-s -w",
				CGOEnabled: true,
			},
		}
		override := &MageConfig{
			Build: BuildConfig{
				GoVersion: "1.24",
				// Platform empty - should preserve base
				// LDFlags empty - should preserve base
			},
		}

		result := manager.Merge(base, override)

		assert.Equal(t, "1.24", result.Build.GoVersion)
		assert.Equal(t, "linux/amd64", result.Build.Platform)
		assert.Equal(t, "-s -w", result.Build.LDFlags)
	})

	t.Run("tags nil does not override base tags", func(t *testing.T) {
		base := &MageConfig{
			Project: ProjectConfig{Name: "test", Version: "1.0.0"},
			Build: BuildConfig{
				GoVersion: "1.20",
				Tags:      []string{"original-tag"},
			},
		}
		override := &MageConfig{
			Build: BuildConfig{
				GoVersion: "1.24",
				Tags:      nil, // nil should not override
			},
		}

		result := manager.Merge(base, override)

		assert.Equal(t, []string{"original-tag"}, result.Build.Tags)
	})
}

// TestMergeTestConfigComprehensive tests all test config merge scenarios
func TestMergeTestConfigComprehensive(t *testing.T) {
	manager := NewManager()

	t.Run("all test fields override", func(t *testing.T) {
		base := &MageConfig{
			Project: ProjectConfig{Name: "test", Version: "1.0.0"},
			Test: TestConfig{
				Timeout:    60,
				Coverage:   false,
				Verbose:    false,
				Race:       false,
				Parallel:   2,
				Tags:       []string{"test-tag1"},
				OutputDir:  "/base/test-output",
				BenchTime:  "1s",
				MemProfile: false,
				CPUProfile: false,
			},
		}
		override := &MageConfig{
			Test: TestConfig{
				Timeout:    120,
				Coverage:   true,
				Verbose:    true,
				Race:       true,
				Parallel:   8,
				Tags:       []string{"test-tag2"},
				OutputDir:  "/override/test-output",
				BenchTime:  "5s",
				MemProfile: true,
				CPUProfile: true,
			},
		}

		result := manager.Merge(base, override)

		assert.Equal(t, 120, result.Test.Timeout)
		assert.True(t, result.Test.Coverage)
		assert.True(t, result.Test.Verbose)
		assert.True(t, result.Test.Race)
		assert.Equal(t, 8, result.Test.Parallel)
		assert.Equal(t, []string{"test-tag2"}, result.Test.Tags)
		assert.Equal(t, "/override/test-output", result.Test.OutputDir)
		assert.Equal(t, "5s", result.Test.BenchTime)
		assert.True(t, result.Test.MemProfile)
		assert.True(t, result.Test.CPUProfile)
	})

	t.Run("partial test override preserves unset fields", func(t *testing.T) {
		base := &MageConfig{
			Project: ProjectConfig{Name: "test", Version: "1.0.0"},
			Test: TestConfig{
				Timeout:   60,
				Coverage:  true,
				OutputDir: "/base/output",
				BenchTime: "2s",
			},
		}
		override := &MageConfig{
			Test: TestConfig{
				Timeout: 120,
				// Coverage has no override trigger (no Timeout, Parallel, or Tags in override)
				// But Timeout IS set, so hasTestOverrides returns true
			},
		}

		result := manager.Merge(base, override)

		assert.Equal(t, 120, result.Test.Timeout)
		assert.Equal(t, "/base/output", result.Test.OutputDir)
		assert.Equal(t, "2s", result.Test.BenchTime)
	})

	t.Run("test tags nil does not override base tags", func(t *testing.T) {
		base := &MageConfig{
			Project: ProjectConfig{Name: "test", Version: "1.0.0"},
			Test: TestConfig{
				Timeout: 60,
				Tags:    []string{"original-tag"},
			},
		}
		override := &MageConfig{
			Test: TestConfig{
				Timeout: 120,
				Tags:    nil, // nil should not override
			},
		}

		result := manager.Merge(base, override)

		assert.Equal(t, []string{"original-tag"}, result.Test.Tags)
	})

	t.Run("boolean fields override when test overrides present", func(t *testing.T) {
		base := &MageConfig{
			Project: ProjectConfig{Name: "test", Version: "1.0.0"},
			Test: TestConfig{
				Timeout:    60,
				Coverage:   true,
				Verbose:    true,
				Race:       true,
				MemProfile: true,
				CPUProfile: true,
			},
		}
		override := &MageConfig{
			Test: TestConfig{
				Timeout:    120, // This triggers hasTestOverrides = true
				Coverage:   false,
				Verbose:    false,
				Race:       false,
				MemProfile: false,
				CPUProfile: false,
			},
		}

		result := manager.Merge(base, override)

		assert.False(t, result.Test.Coverage)
		assert.False(t, result.Test.Verbose)
		assert.False(t, result.Test.Race)
		assert.False(t, result.Test.MemProfile)
		assert.False(t, result.Test.CPUProfile)
	})
}

// TestLoaderImplSave tests the loaderImpl.Save method
func TestLoaderImplSave(t *testing.T) {
	loader := NewLoader()

	t.Run("saves to default mage.yaml path", func(t *testing.T) {
		// Change to temp directory for this test
		originalDir, err := os.Getwd()
		require.NoError(t, err)

		tmpDir := t.TempDir()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() {
			require.NoError(t, os.Chdir(originalDir))
		}()

		config := &MageConfig{
			Project: ProjectConfig{Name: "save-test", Version: "1.0.0"},
		}

		err = loader.Save(config)
		require.NoError(t, err)

		// Verify file was created
		_, statErr := os.Stat("mage.yaml")
		require.NoError(t, statErr)
	})
}

// TestLoaderImplLoadFromPath tests the loaderImpl.LoadFromPath method
func TestLoaderImplLoadFromPath(t *testing.T) {
	loader := NewLoader()

	t.Run("returns error for invalid path", func(t *testing.T) {
		_, err := loader.LoadFromPath("/nonexistent/config.yaml")
		require.Error(t, err)
	})

	t.Run("loads valid config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := tmpDir + "/config.yaml"

		configContent := `project:
  name: loader-test
  version: 2.0.0
`
		require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0o600))

		config, err := loader.LoadFromPath(configPath)
		require.NoError(t, err)
		assert.Equal(t, "loader-test", config.Project.Name)
	})
}
