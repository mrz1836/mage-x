package mage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateDefaultConfig tests the createDefaultConfig function
func TestCreateDefaultConfig(t *testing.T) {
	config := createDefaultConfig()

	require.NotNil(t, config)

	// Verify project defaults
	assert.Equal(t, "1.0", config.Version)
	assert.Equal(t, "my-project", config.Project.Name)
	assert.Equal(t, "A Go project built with MAGE-X", config.Project.Description)
	assert.Equal(t, "github.com/username/my-project", config.Project.Module)
	assert.Equal(t, ".", config.Project.MainPkg)
	assert.Equal(t, "my-project", config.Project.BinaryName)
	assert.Equal(t, "v1.0.0", config.Project.Version)
	assert.Equal(t, "MIT", config.Project.License)
	assert.Len(t, config.Project.Authors, 1)

	// Verify build defaults
	assert.Equal(t, ".", config.Build.Dir)
	assert.Equal(t, "bin", config.Build.OutputDir)
	assert.Equal(t, "-s -w", config.Build.LDFlags)
	assert.False(t, config.Build.CGO)
	assert.Equal(t, 4, config.Build.Parallel)
	assert.Len(t, config.Build.Platforms, 3)

	// Verify test defaults
	assert.Equal(t, "10m", config.Test.Timeout)
	assert.True(t, config.Test.Parallel)
	assert.False(t, config.Test.Race)
	assert.True(t, config.Test.Cover)
	assert.Equal(t, "atomic", config.Test.CoverMode)

	// Verify lint defaults
	assert.True(t, config.Lint.Enabled)
	assert.Equal(t, ".golangci.json", config.Lint.ConfigFile)
	assert.Equal(t, "5m", config.Lint.Timeout)
	assert.Equal(t, 4, config.Lint.Parallel)

	// Verify release defaults
	assert.True(t, config.Release.Enabled)
	assert.Equal(t, "stable", config.Release.Channel)
	assert.False(t, config.Release.Prerelease)
	assert.False(t, config.Release.Draft)
	assert.True(t, config.Release.Changelog.Enabled)

	// Verify CI defaults
	assert.True(t, config.CI.Enabled)
	assert.Equal(t, "github", config.CI.Provider)
	assert.Len(t, config.CI.Matrix.GoVersions, 2)
	assert.Len(t, config.CI.Matrix.Platforms, 3)
}

// TestCreateLibraryTemplate tests the createLibraryTemplate function
func TestCreateLibraryTemplate(t *testing.T) {
	config := createLibraryTemplate()

	require.NotNil(t, config)
	assert.Equal(t, "A Go library built with MAGE-X", config.Project.Description)
	assert.True(t, config.Test.Benchmarks)
	assert.Len(t, config.Build.Platforms, 3)
}

// TestCreateCLITemplate tests the createCLITemplate function
func TestCreateCLITemplate(t *testing.T) {
	config := createCLITemplate()

	require.NotNil(t, config)
	assert.Equal(t, "A CLI application built with MAGE-X", config.Project.Description)
	assert.Len(t, config.Build.Platforms, 5) // More platforms for CLI
	assert.Contains(t, config.Release.Assets, "completions/*")
	assert.Contains(t, config.Release.Assets, "docs/*")
}

// TestCreateWebAPITemplate tests the createWebAPITemplate function
func TestCreateWebAPITemplate(t *testing.T) {
	config := createWebAPITemplate()

	require.NotNil(t, config)
	assert.Equal(t, "A web API built with MAGE-X", config.Project.Description)
	assert.Contains(t, config.Build.Tags, "netgo")
	assert.Contains(t, config.Test.Tags, "integration")
}

// TestCreateMicroserviceTemplate tests the createMicroserviceTemplate function
func TestCreateMicroserviceTemplate(t *testing.T) {
	config := createMicroserviceTemplate()

	require.NotNil(t, config)
	assert.Equal(t, "A microservice built with MAGE-X", config.Project.Description)
	assert.Contains(t, config.Build.Tags, "netgo")
	assert.Contains(t, config.Test.Tags, "integration")
	assert.Contains(t, config.Test.Tags, "e2e")
}

// TestCreateToolTemplate tests the createToolTemplate function
func TestCreateToolTemplate(t *testing.T) {
	config := createToolTemplate()

	require.NotNil(t, config)
	assert.Equal(t, "A developer tool built with MAGE-X", config.Project.Description)
	assert.Len(t, config.Build.Platforms, 5) // More platforms for tools
	assert.Contains(t, config.Release.Assets, "LICENSE")
	assert.Contains(t, config.Release.Assets, "README.md")
}

// TestValidateConfigSuccess tests validateConfig with valid configuration
func TestValidateConfigSuccess(t *testing.T) {
	tests := []struct {
		name   string
		config *YamlConfig
	}{
		{
			name: "valid full config",
			config: &YamlConfig{
				Project: ProjectYamlConfig{
					Name:   "test-project",
					Module: "github.com/test/project",
				},
				Build: BuildYamlConfig{
					OutputDir: "bin",
					Platforms: []string{"linux/amd64", "darwin/amd64"},
				},
			},
		},
		{
			name: "valid minimal config",
			config: &YamlConfig{
				Project: ProjectYamlConfig{
					Name:   "a",
					Module: "b",
				},
				Build: BuildYamlConfig{
					OutputDir: "dist",
				},
			},
		},
		{
			name:   "valid default config",
			config: createDefaultConfig(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			require.NoError(t, err)
		})
	}
}

// TestValidateConfigErrors tests validateConfig with invalid configurations
func TestValidateConfigErrors(t *testing.T) {
	tests := []struct {
		name      string
		config    *YamlConfig
		wantErr   error
		errSubstr string
	}{
		{
			name: "empty project name",
			config: &YamlConfig{
				Project: ProjectYamlConfig{
					Name:   "",
					Module: "github.com/test/project",
				},
				Build: BuildYamlConfig{
					OutputDir: "bin",
				},
			},
			wantErr: errProjectNameRequired,
		},
		{
			name: "empty module",
			config: &YamlConfig{
				Project: ProjectYamlConfig{
					Name:   "test-project",
					Module: "",
				},
				Build: BuildYamlConfig{
					OutputDir: "bin",
				},
			},
			wantErr: errProjectModuleRequired,
		},
		{
			name: "empty output dir",
			config: &YamlConfig{
				Project: ProjectYamlConfig{
					Name:   "test-project",
					Module: "github.com/test/project",
				},
				Build: BuildYamlConfig{
					OutputDir: "",
				},
			},
			wantErr: errBuildOutputRequired,
		},
		{
			name: "invalid platform format single part",
			config: &YamlConfig{
				Project: ProjectYamlConfig{
					Name:   "test-project",
					Module: "github.com/test/project",
				},
				Build: BuildYamlConfig{
					OutputDir: "bin",
					Platforms: []string{"linux"},
				},
			},
			errSubstr: "invalid platform format",
		},
		{
			name: "invalid platform format three parts",
			config: &YamlConfig{
				Project: ProjectYamlConfig{
					Name:   "test-project",
					Module: "github.com/test/project",
				},
				Build: BuildYamlConfig{
					OutputDir: "bin",
					Platforms: []string{"linux/amd64/extra"},
				},
			},
			errSubstr: "invalid platform format",
		},
		{
			name: "empty platform string",
			config: &YamlConfig{
				Project: ProjectYamlConfig{
					Name:   "test-project",
					Module: "github.com/test/project",
				},
				Build: BuildYamlConfig{
					OutputDir: "bin",
					Platforms: []string{""},
				},
			},
			errSubstr: "invalid platform format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			require.Error(t, err)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			}
			if tt.errSubstr != "" {
				assert.Contains(t, err.Error(), tt.errSubstr)
			}
		})
	}
}

// TestWriteAndLoadConfig tests writeConfig and loadConfig roundtrip
func TestWriteAndLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "mage.yaml")

	// Create a config to write
	original := createDefaultConfig()
	original.Project.Name = "test-roundtrip"
	original.Project.Module = "github.com/test/roundtrip"
	original.Build.Platforms = []string{"linux/amd64", "darwin/arm64"}

	// Write the config
	err := writeConfig(original, configPath)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Load it back
	loaded, err := loadConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, loaded)

	// Verify fields match
	assert.Equal(t, original.Project.Name, loaded.Project.Name)
	assert.Equal(t, original.Project.Module, loaded.Project.Module)
	assert.Equal(t, original.Build.Platforms, loaded.Build.Platforms)
	assert.Equal(t, original.Version, loaded.Version)
}

// TestLoadConfigErrors tests loadConfig error cases
func TestLoadConfigErrors(t *testing.T) {
	t.Run("nonexistent file", func(t *testing.T) {
		_, err := loadConfig("/nonexistent/path/mage.yaml")
		require.Error(t, err)
	})

	t.Run("invalid yaml content", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.yaml")

		err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0o600)
		require.NoError(t, err)

		_, err = loadConfig(configPath)
		require.Error(t, err)
	})

	t.Run("empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "empty.yaml")

		err := os.WriteFile(configPath, []byte(""), 0o600)
		require.NoError(t, err)

		// Empty file should load but result in zero-value config
		config, err := loadConfig(configPath)
		require.NoError(t, err)
		assert.Empty(t, config.Project.Name)
	})
}

// TestPopulateFromProject tests populateFromProject function
func TestPopulateFromProject(t *testing.T) {
	t.Run("populates from module path", func(t *testing.T) {
		config := createDefaultConfig()

		// We can't easily mock getModuleName, but we can test the logic
		// by checking that the function doesn't panic and maintains defaults
		// when module info is not available
		populateFromProject(config)

		// Should not panic and should have some value
		assert.NotEmpty(t, config.Project.Name)
	})
}

// TestUpdateFromEnv tests updateFromEnv function
func TestUpdateFromEnv(t *testing.T) {
	tests := []struct {
		name      string
		envVars   map[string]string
		checkFunc func(*testing.T, *YamlConfig)
	}{
		{
			name: "update project name",
			envVars: map[string]string{
				"MAGE_X_PROJECT_NAME": "env-project",
			},
			checkFunc: func(t *testing.T, config *YamlConfig) {
				assert.Equal(t, "env-project", config.Project.Name)
			},
		},
		{
			name: "update project description",
			envVars: map[string]string{
				"MAGE_X_PROJECT_DESCRIPTION": "A test description",
			},
			checkFunc: func(t *testing.T, config *YamlConfig) {
				assert.Equal(t, "A test description", config.Project.Description)
			},
		},
		{
			name: "update project version",
			envVars: map[string]string{
				"MAGE_X_PROJECT_VERSION": "v2.0.0",
			},
			checkFunc: func(t *testing.T, config *YamlConfig) {
				assert.Equal(t, "v2.0.0", config.Project.Version)
			},
		},
		{
			name: "update project license",
			envVars: map[string]string{
				"MAGE_X_PROJECT_LICENSE": "Apache-2.0",
			},
			checkFunc: func(t *testing.T, config *YamlConfig) {
				assert.Equal(t, "Apache-2.0", config.Project.License)
			},
		},
		{
			name: "update build ldflags",
			envVars: map[string]string{
				"MAGE_X_BUILD_LDFLAGS": "-X main.version=test",
			},
			checkFunc: func(t *testing.T, config *YamlConfig) {
				assert.Equal(t, "-X main.version=test", config.Build.LDFlags)
			},
		},
		{
			name: "update build platforms",
			envVars: map[string]string{
				"MAGE_X_BUILD_PLATFORMS": "linux/amd64,darwin/arm64",
			},
			checkFunc: func(t *testing.T, config *YamlConfig) {
				assert.Equal(t, []string{"linux/amd64", "darwin/arm64"}, config.Build.Platforms)
			},
		},
		{
			name: "update test timeout",
			envVars: map[string]string{
				"MAGE_X_TEST_TIMEOUT": "30m",
			},
			checkFunc: func(t *testing.T, config *YamlConfig) {
				assert.Equal(t, "30m", config.Test.Timeout)
			},
		},
		{
			name: "update test verbose",
			envVars: map[string]string{
				"TEST_VERBOSE": "true",
			},
			checkFunc: func(t *testing.T, config *YamlConfig) {
				assert.True(t, config.Test.Verbose)
			},
		},
		{
			name: "update test race",
			envVars: map[string]string{
				"TEST_RACE": "true",
			},
			checkFunc: func(t *testing.T, config *YamlConfig) {
				assert.True(t, config.Test.Race)
			},
		},
		{
			name: "update test cover",
			envVars: map[string]string{
				"TEST_COVER": "false",
			},
			checkFunc: func(t *testing.T, config *YamlConfig) {
				assert.False(t, config.Test.Cover)
			},
		},
		{
			name: "multiple env vars",
			envVars: map[string]string{
				"MAGE_X_PROJECT_NAME":    "multi-env",
				"MAGE_X_PROJECT_VERSION": "v3.0.0",
				"TEST_VERBOSE":           "true",
			},
			checkFunc: func(t *testing.T, config *YamlConfig) {
				assert.Equal(t, "multi-env", config.Project.Name)
				assert.Equal(t, "v3.0.0", config.Project.Version)
				assert.True(t, config.Test.Verbose)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars
			for k, v := range tt.envVars {
				require.NoError(t, os.Setenv(k, v))
			}
			// Cleanup
			t.Cleanup(func() {
				for k := range tt.envVars {
					require.NoError(t, os.Unsetenv(k))
				}
			})

			config := createDefaultConfig()
			updateFromEnv(config)

			tt.checkFunc(t, config)
		})
	}
}

// TestUpdateFromEnvNoChange tests updateFromEnv with empty env vars
func TestUpdateFromEnvNoChange(t *testing.T) {
	// Ensure relevant env vars are not set
	envVars := []string{
		"MAGE_X_PROJECT_NAME",
		"MAGE_X_PROJECT_DESCRIPTION",
		"MAGE_X_PROJECT_VERSION",
		"MAGE_X_PROJECT_LICENSE",
		"MAGE_X_BUILD_LDFLAGS",
		"MAGE_X_BUILD_PLATFORMS",
		"MAGE_X_TEST_TIMEOUT",
	}
	for _, k := range envVars {
		require.NoError(t, os.Unsetenv(k))
	}

	original := createDefaultConfig()
	config := createDefaultConfig()
	updateFromEnv(config)

	// Should match original defaults
	assert.Equal(t, original.Project.Name, config.Project.Name)
	assert.Equal(t, original.Project.Description, config.Project.Description)
	assert.Equal(t, original.Build.LDFlags, config.Build.LDFlags)
}

// TestDisplayConfig tests displayConfig doesn't panic
func TestDisplayConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *YamlConfig
	}{
		{
			name:   "default config",
			config: createDefaultConfig(),
		},
		{
			name: "minimal config",
			config: &YamlConfig{
				Project: ProjectYamlConfig{Name: "test"},
				Build:   BuildYamlConfig{OutputDir: "bin"},
			},
		},
		{
			name:   "library template",
			config: createLibraryTemplate(),
		},
		{
			name:   "cli template",
			config: createCLITemplate(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture output - displayConfig should not panic
			assert.NotPanics(t, func() {
				displayConfig(tt.config)
			})
		})
	}
}

// TestYamlInitSuccess tests Yaml.Init with successful creation
func TestYamlInitSuccess(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldWd))
	})

	// Run Init
	y := Yaml{}
	err = y.Init()
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(filepath.Join(tmpDir, "mage.yaml"))
	require.NoError(t, err)
}

// TestYamlInitAlreadyExists tests Yaml.Init when file already exists
func TestYamlInitAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Create existing mage.yaml
	existingPath := filepath.Join(tmpDir, "mage.yaml")
	err := os.WriteFile(existingPath, []byte("version: 1.0"), 0o600)
	require.NoError(t, err)

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldWd))
	})

	// Run Init - should fail
	y := Yaml{}
	err = y.Init()
	require.Error(t, err)
	assert.ErrorIs(t, err, errMageYamlExists)
}

// TestYamlValidateSuccess tests Yaml.Validate with valid config
func TestYamlValidateSuccess(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid mage.yaml
	config := createDefaultConfig()
	configPath := filepath.Join(tmpDir, "mage.yaml")
	err := writeConfig(config, configPath)
	require.NoError(t, err)

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldWd))
	})

	// Run Validate
	y := Yaml{}
	err = y.Validate()
	require.NoError(t, err)
}

// TestYamlValidateNoFile tests Yaml.Validate when file doesn't exist
func TestYamlValidateNoFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory (no mage.yaml)
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldWd))
	})

	// Run Validate - should fail
	y := Yaml{}
	err = y.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load configuration")
}

// TestYamlValidateInvalidConfig tests Yaml.Validate with invalid config
func TestYamlValidateInvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()

	// Create invalid mage.yaml (missing required fields)
	configPath := filepath.Join(tmpDir, "mage.yaml")
	err := os.WriteFile(configPath, []byte(`
version: "1.0"
project:
  name: ""
  module: ""
build:
  output_dir: ""
`), 0o600)
	require.NoError(t, err)

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldWd))
	})

	// Run Validate - should fail
	y := Yaml{}
	err = y.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

// TestYamlShowSuccess tests Yaml.Show with valid config
func TestYamlShowSuccess(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid mage.yaml
	config := createDefaultConfig()
	configPath := filepath.Join(tmpDir, "mage.yaml")
	err := writeConfig(config, configPath)
	require.NoError(t, err)

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldWd))
	})

	// Run Show
	y := Yaml{}
	err = y.Show()
	require.NoError(t, err)
}

// TestYamlShowNoFile tests Yaml.Show when file doesn't exist
func TestYamlShowNoFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory (no mage.yaml)
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldWd))
	})

	// Run Show - should fail
	y := Yaml{}
	err = y.Show()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load configuration")
}

// TestYamlUpdateSuccess tests Yaml.Update with valid config
func TestYamlUpdateSuccess(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid mage.yaml
	config := createDefaultConfig()
	configPath := filepath.Join(tmpDir, "mage.yaml")
	err := writeConfig(config, configPath)
	require.NoError(t, err)

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldWd))
	})

	// Set env var to update - use description since project name gets overwritten
	// by populateFromProject which runs after updateFromEnv
	require.NoError(t, os.Setenv("MAGE_X_PROJECT_DESCRIPTION", "Updated description"))
	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv("MAGE_X_PROJECT_DESCRIPTION"))
	})

	// Run Update
	y := Yaml{}
	err = y.Update()
	require.NoError(t, err)

	// Verify update was saved
	updated, err := loadConfig(configPath)
	require.NoError(t, err)
	assert.Equal(t, "Updated description", updated.Project.Description)
}

// TestYamlUpdateNoFile tests Yaml.Update when file doesn't exist
func TestYamlUpdateNoFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory (no mage.yaml)
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldWd))
	})

	// Run Update - should fail
	y := Yaml{}
	err = y.Update()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load configuration")
}

// TestYamlTemplateSuccess tests Yaml.Template for all project types
func TestYamlTemplateSuccess(t *testing.T) {
	projectTypes := []string{"library", "cli", "webapi", "microservice", "tool", "unknown"}

	for _, pt := range projectTypes {
		t.Run(pt, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Change to temp directory
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			require.NoError(t, os.Chdir(tmpDir))
			t.Cleanup(func() {
				require.NoError(t, os.Chdir(oldWd))
			})

			// Set project type
			require.NoError(t, os.Setenv("MAGE_X_PROJECT_TYPE", pt))
			t.Cleanup(func() {
				require.NoError(t, os.Unsetenv("MAGE_X_PROJECT_TYPE"))
			})

			// Run Template
			y := Yaml{}
			err = y.Template()
			require.NoError(t, err)

			// Verify file was created
			expectedFile := "mage." + pt + ".yaml"
			_, err = os.Stat(filepath.Join(tmpDir, expectedFile))
			require.NoError(t, err)
		})
	}
}

// TestYamlTemplateDefaultType tests Yaml.Template with default type
func TestYamlTemplateDefaultType(t *testing.T) {
	tmpDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldWd))
	})

	// Ensure env var is not set
	require.NoError(t, os.Unsetenv("MAGE_X_PROJECT_TYPE"))

	// Run Template - should use default "library"
	y := Yaml{}
	err = y.Template()
	require.NoError(t, err)

	// Verify library template was created
	_, err = os.Stat(filepath.Join(tmpDir, "mage.library.yaml"))
	require.NoError(t, err)
}

// TestYamlConfigStructTags tests that YAML struct tags are correct
func TestYamlConfigStructTags(t *testing.T) {
	config := createDefaultConfig()

	// Write and read back to verify struct tags work
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test.yaml")

	err := writeConfig(config, configPath)
	require.NoError(t, err)

	// Read file contents
	content, err := os.ReadFile(configPath) //nolint:gosec // test with controlled path
	require.NoError(t, err)

	// Verify expected YAML keys are present
	expectedKeys := []string{
		"version:",
		"project:",
		"name:",
		"description:",
		"module:",
		"build:",
		"output_dir:",
		"ldflags:",
		"platforms:",
		"test:",
		"timeout:",
		"parallel:",
		"race:",
		"cover:",
		"lint:",
		"enabled:",
		"config_file:",
		"release:",
		"channel:",
		"ci:",
		"provider:",
	}

	contentStr := string(content)
	for _, key := range expectedKeys {
		assert.Contains(t, contentStr, key,
			"expected key %q not found in YAML output", key)
	}
}

// TestYamlConfigPlatformValidation tests platform format validation
func TestYamlConfigPlatformValidation(t *testing.T) {
	validPlatforms := []string{
		"linux/amd64",
		"darwin/arm64",
		"windows/amd64",
		"linux/386",
		"freebsd/amd64",
	}

	// Note: validation only checks for exactly 2 parts after split by "/"
	// It doesn't validate that parts are non-empty (e.g., "/amd64" passes)
	invalidPlatforms := []string{
		"linux",             // 1 part
		"linux/amd64/extra", // 3 parts
		"",                  // 1 empty part
	}

	t.Run("valid platforms", func(t *testing.T) {
		for _, p := range validPlatforms {
			config := &YamlConfig{
				Project: ProjectYamlConfig{
					Name:   "test",
					Module: "test/module",
				},
				Build: BuildYamlConfig{
					OutputDir: "bin",
					Platforms: []string{p},
				},
			}
			err := validateConfig(config)
			assert.NoError(t, err, "platform %q should be valid", p)
		}
	})

	t.Run("invalid platforms", func(t *testing.T) {
		for _, p := range invalidPlatforms {
			config := &YamlConfig{
				Project: ProjectYamlConfig{
					Name:   "test",
					Module: "test/module",
				},
				Build: BuildYamlConfig{
					OutputDir: "bin",
					Platforms: []string{p},
				},
			}
			err := validateConfig(config)
			assert.Error(t, err, "platform %q should be invalid", p)
		}
	})
}
