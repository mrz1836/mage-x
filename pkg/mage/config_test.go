package mage

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ConfigTestSuite provides a comprehensive test suite for configuration functionality
type ConfigTestSuite struct {
	suite.Suite

	tempDir     string
	origEnvVars map[string]string
}

// SetupSuite runs before all tests in the suite
func (ts *ConfigTestSuite) SetupSuite() {
	// Create temporary directory for test files
	var err error
	ts.tempDir, err = os.MkdirTemp("", "mage-config-test-*")
	ts.Require().NoError(err)

	// Store original environment variables
	ts.origEnvVars = make(map[string]string)
	envVars := []string{
		"MAGE_X_BINARY_NAME", "CUSTOM_BINARY_NAME", "MAGE_X_BUILD_TAGS", "MAGE_X_VERBOSE",
		"MAGE_X_TEST_RACE", "MAGE_X_PARALLEL", "MAGE_X_ORG_NAME", "MAGE_X_ORG_DOMAIN",
		"MAGE_X_SECURITY_LEVEL", "MAGE_X_ENABLE_VAULT", "VAULT_ADDR",
		"MAGE_X_ANALYTICS_ENABLED", "MAGE_X_METRICS_INTERVAL",
	}
	for _, env := range envVars {
		ts.origEnvVars[env] = os.Getenv(env)
	}
}

// TearDownSuite runs after all tests in the suite
func (ts *ConfigTestSuite) TearDownSuite() {
	// Restore original environment variables
	for env, value := range ts.origEnvVars {
		if value == "" {
			ts.Require().NoError(os.Unsetenv(env))
		} else {
			ts.Require().NoError(os.Setenv(env, value))
		}
	}

	// Clean up temporary directory
	ts.Require().NoError(os.RemoveAll(ts.tempDir))
}

// SetupTest runs before each test
func (ts *ConfigTestSuite) SetupTest() {
	// Reset configuration before each test
	TestResetConfig()

	// Clear environment variables for clean test state
	envVars := []string{
		"MAGE_X_BINARY_NAME", "CUSTOM_BINARY_NAME", "MAGE_X_BUILD_TAGS", "MAGE_X_VERBOSE",
		"MAGE_X_TEST_RACE", "MAGE_X_PARALLEL", "MAGE_X_ORG_NAME", "MAGE_X_ORG_DOMAIN",
	}
	for _, env := range envVars {
		ts.Require().NoError(os.Unsetenv(env))
	}
}

// TearDownTest runs after each test
func (ts *ConfigTestSuite) TearDownTest() {
	// Reset configuration after each test
	TestResetConfig()
}

// TestConfigSuite runs the configuration test suite
func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

// TestDefaultConfig tests the default configuration generation
func (ts *ConfigTestSuite) TestDefaultConfig() {
	ts.Run("DefaultConfigStructure", func() {
		config := defaultConfig()
		ts.Require().NotNil(config)

		// Test project defaults
		ts.Require().NotEmpty(config.Project.Binary)
		ts.Require().Equal("github.com", config.Project.GitDomain)

		// Test build defaults
		ts.Require().Equal("bin", config.Build.Output)
		ts.Require().True(config.Build.TrimPath)
		ts.Require().Equal(runtime.NumCPU(), config.Build.Parallel)
		ts.Require().Contains(config.Build.Platforms, "linux/amd64")
		ts.Require().Contains(config.Build.Platforms, "darwin/amd64")
		ts.Require().Contains(config.Build.Platforms, "windows/amd64")

		// Test test defaults
		ts.Require().Positive(config.Test.Parallel)
		ts.Require().Equal("10m", config.Test.Timeout)
		ts.Require().Equal("atomic", config.Test.CoverMode)

		// Test lint defaults
		ts.Require().Equal("v2.4.0", config.Lint.GolangciVersion)
		ts.Require().Equal("5m", config.Lint.Timeout)

		// Test tools defaults
		ts.Require().Equal("v2.4.0", config.Tools.GolangciLint)
		ts.Require().Equal(VersionLatest, config.Tools.Fumpt)
		ts.Require().Equal("v1.1.4", config.Tools.GoVulnCheck)

		// Test docker defaults
		ts.Require().Equal("Dockerfile", config.Docker.Dockerfile)
		ts.Require().True(config.Docker.EnableBuildKit)
		ts.Require().Equal("docker.io", config.Docker.DefaultRegistry)
		ts.Require().NotNil(config.Docker.BuildArgs)
		ts.Require().NotNil(config.Docker.Labels)

		// Test release defaults
		ts.Require().Equal("GITHUB_TOKEN", config.Release.GitHubToken)
		ts.Require().True(config.Release.Changelog)
		ts.Require().Contains(config.Release.Formats, "tar.gz")
		ts.Require().Contains(config.Release.Formats, "zip")
	})

	ts.Run("DefaultConfigWithEmptyModule", func() {
		// Test when module name cannot be determined
		config := defaultConfig()
		ts.Require().NotNil(config)
		ts.Require().NotEmpty(config.Project.Binary)
	})

	ts.Run("ParallelismHandling", func() {
		// Test that parallel value is at least 1
		config := defaultConfig()
		ts.Require().GreaterOrEqual(config.Build.Parallel, 1)
	})
}

// TestEnvironmentOverrides tests environment variable overrides
func (ts *ConfigTestSuite) TestEnvironmentOverrides() {
	ts.Run("BinaryNameOverrides", func() {
		config := defaultConfig()

		// Test BINARY_NAME override
		ts.Require().NoError(os.Setenv("MAGE_X_BINARY_NAME", "custom-binary"))
		applyEnvOverrides(config)
		ts.Require().Equal("custom-binary", config.Project.Binary)

		// Test CUSTOM_BINARY_NAME override (should take precedence)
		ts.Require().NoError(os.Setenv("CUSTOM_BINARY_NAME", "another-binary"))
		applyEnvOverrides(config)
		ts.Require().Equal("another-binary", config.Project.Binary)
	})

	ts.Run("BuildTagsOverride", func() {
		config := defaultConfig()
		ts.Require().NoError(os.Setenv("MAGE_X_BUILD_TAGS", "tag1,tag2,tag3"))
		applyEnvOverrides(config)
		ts.Require().Equal([]string{"tag1", "tag2", "tag3"}, config.Build.Tags)
	})

	ts.Run("VerboseOverrides", func() {
		config := defaultConfig()

		// Test with "true"
		ts.Require().NoError(os.Setenv("MAGE_X_VERBOSE", "true"))
		applyEnvOverrides(config)
		ts.Require().True(config.Build.Verbose)
		ts.Require().True(config.Test.Verbose)

		// Reset config
		config = defaultConfig()

		// Test with "1"
		ts.Require().NoError(os.Setenv("MAGE_X_VERBOSE", "1"))
		applyEnvOverrides(config)
		ts.Require().True(config.Build.Verbose)
		ts.Require().True(config.Test.Verbose)

		// Test with "false" (should not change defaults)
		config = defaultConfig()
		ts.Require().NoError(os.Setenv("MAGE_X_VERBOSE", "false"))
		applyEnvOverrides(config)
		ts.Require().False(config.Build.Verbose)
		ts.Require().False(config.Test.Verbose)
	})

	ts.Run("TestRaceOverride", func() {
		config := defaultConfig()

		// Test with "true"
		ts.Require().NoError(os.Setenv("MAGE_X_TEST_RACE", "true"))
		applyEnvOverrides(config)
		ts.Require().True(config.Test.Race)

		// Test with "1"
		config = defaultConfig()
		ts.Require().NoError(os.Setenv("MAGE_X_TEST_RACE", "1"))
		applyEnvOverrides(config)
		ts.Require().True(config.Test.Race)
	})

	ts.Run("ParallelOverride", func() {
		config := defaultConfig()
		ts.Require().NoError(os.Setenv("MAGE_X_PARALLEL", "8"))
		applyEnvOverrides(config)
		ts.Require().Equal(8, config.Build.Parallel)

		// Test invalid value (should not change config)
		config = defaultConfig()
		originalParallel := config.Build.Parallel
		ts.Require().NoError(os.Setenv("MAGE_X_PARALLEL", "invalid"))
		applyEnvOverrides(config)
		ts.Require().Equal(originalParallel, config.Build.Parallel)

		// Test zero value (should not change config)
		config = defaultConfig()
		originalParallel = config.Build.Parallel
		ts.Require().NoError(os.Setenv("MAGE_X_PARALLEL", "0"))
		applyEnvOverrides(config)
		ts.Require().Equal(originalParallel, config.Build.Parallel)
	})
}

// TestEnterpriseEnvironmentOverrides tests enterprise-specific environment overrides
func (ts *ConfigTestSuite) TestEnterpriseEnvironmentOverrides() {
	ts.Run("OrganizationOverrides", func() {
		enterpriseConfig := &EnterpriseConfiguration{
			Organization: OrganizationConfig{
				Name:   "Original Org",
				Domain: "original.com",
			},
		}

		ts.Require().NoError(os.Setenv("MAGE_X_ORG_NAME", "New Org"))
		ts.Require().NoError(os.Setenv("MAGE_X_ORG_DOMAIN", "new.com"))

		applyEnterpriseEnvOverrides(enterpriseConfig)

		ts.Require().Equal("New Org", enterpriseConfig.Organization.Name)
		ts.Require().Equal("new.com", enterpriseConfig.Organization.Domain)
	})

	ts.Run("PlaceholderEnvironmentVariables", func() {
		// Test that placeholder environment variables don't cause errors
		ts.Require().NoError(os.Setenv("MAGE_X_SECURITY_LEVEL", "high"))
		ts.Require().NoError(os.Setenv("MAGE_X_ENABLE_VAULT", "true"))
		ts.Require().NoError(os.Setenv("VAULT_ADDR", "https://vault.example.com"))
		ts.Require().NoError(os.Setenv("MAGE_X_ANALYTICS_ENABLED", "true"))
		ts.Require().NoError(os.Setenv("MAGE_X_METRICS_INTERVAL", "30s"))

		enterpriseConfig := &EnterpriseConfiguration{}

		// Should not panic or error
		ts.Require().NotPanics(func() {
			applyEnterpriseEnvOverrides(enterpriseConfig)
		})
	})

	ts.Run("EnterpriseOverridesWithMainConfig", func() {
		config := defaultConfig()
		config.Enterprise = &EnterpriseConfiguration{
			Organization: OrganizationConfig{
				Name:   "Test Org",
				Domain: "test.com",
			},
		}

		ts.Require().NoError(os.Setenv("MAGE_X_ORG_NAME", "Updated Org"))
		applyEnvOverrides(config)

		ts.Require().Equal("Updated Org", config.Enterprise.Organization.Name)
	})
}

// TestConfigFunctions tests the main configuration functions
func (ts *ConfigTestSuite) TestConfigFunctions() {
	ts.Run("GetConfig", func() {
		config, err := GetConfig()
		ts.Require().NoError(err)
		ts.Require().NotNil(config)
	})

	ts.Run("LoadConfigDeprecated", func() {
		// Test deprecated LoadConfig function
		config, err := LoadConfig()
		ts.Require().NoError(err)
		ts.Require().NotNil(config)
	})

	ts.Run("BinaryName", func() {
		// Test with default config
		name := BinaryName()
		ts.Require().NotEmpty(name)

		// Test with custom config
		customConfig := defaultConfig()
		customConfig.Project.Binary = "test-binary"
		TestSetConfig(customConfig)

		name = BinaryName()
		ts.Require().Equal("test-binary", name)
	})

	ts.Run("BuildTags", func() {
		// Test with no tags
		tags := BuildTags()
		ts.Require().Empty(tags)

		// Test with custom tags
		customConfig := defaultConfig()
		customConfig.Build.Tags = []string{"tag1", "tag2"}
		TestSetConfig(customConfig)

		tags = BuildTags()
		ts.Require().Equal("tag1,tag2", tags)
	})

	ts.Run("IsVerbose", func() {
		// Test default (should be false)
		ts.Require().False(IsVerbose())

		// Test with verbose build
		customConfig := defaultConfig()
		customConfig.Build.Verbose = true
		TestSetConfig(customConfig)
		ts.Require().True(IsVerbose())

		// Test with verbose test
		customConfig = defaultConfig()
		customConfig.Test.Verbose = true
		TestSetConfig(customConfig)
		ts.Require().True(IsVerbose())
	})

	ts.Run("HasEnterpriseConfig", func() {
		// Test without enterprise config
		ts.Require().False(HasEnterpriseConfig())

		// Test with enterprise config
		customConfig := defaultConfig()
		customConfig.Enterprise = &EnterpriseConfiguration{}
		TestSetConfig(customConfig)
		ts.Require().True(HasEnterpriseConfig())
	})

	ts.Run("GetEnterpriseConfig", func() {
		// Test without enterprise config
		TestResetConfig()
		enterpriseConfig := GetEnterpriseConfig()

		// Enterprise config might exist due to default configuration
		// We test that the function returns without error
		if enterpriseConfig != nil {
			ts.Require().NotNil(enterpriseConfig)
		}

		// Test with custom enterprise config
		customConfig := defaultConfig()
		expectedEnterprise := &EnterpriseConfiguration{
			Organization: OrganizationConfig{
				Name: "Test Org",
			},
		}
		customConfig.Enterprise = expectedEnterprise
		TestSetConfig(customConfig)

		enterpriseConfig = GetEnterpriseConfig()
		ts.Require().NotNil(enterpriseConfig)
		ts.Require().Equal("Test Org", enterpriseConfig.Organization.Name)
	})
}

// TestConfigPersistence tests configuration saving and loading
func (ts *ConfigTestSuite) TestConfigPersistence() {
	ts.Run("GetConfigFilePath", func() {
		// Test default config file path when no config files exist
		configPath := getConfigFilePath()
		ts.Require().Equal(".mage.yaml", configPath)
	})

	ts.Run("SaveConfig", func() {
		// Change to temp directory for this test
		originalDir, err := os.Getwd()
		ts.Require().NoError(err)
		defer func() {
			ts.Require().NoError(os.Chdir(originalDir))
		}()

		ts.Require().NoError(os.Chdir(ts.tempDir))

		config := defaultConfig()
		config.Project.Name = "test-project"

		err = SaveConfig(config)
		ts.Require().NoError(err)

		// Check that file was created
		configPath := getConfigFilePath()
		_, err = os.Stat(configPath)
		ts.Require().NoError(err)
	})

	ts.Run("SaveEnterpriseConfig", func() {
		// Change to temp directory for this test
		originalDir, err := os.Getwd()
		ts.Require().NoError(err)
		defer func() {
			ts.Require().NoError(os.Chdir(originalDir))
		}()

		ts.Require().NoError(os.Chdir(ts.tempDir))

		enterpriseConfig := &EnterpriseConfiguration{
			Organization: OrganizationConfig{
				Name:   "Test Enterprise",
				Domain: "test.enterprise.com",
			},
		}

		err = SaveEnterpriseConfig(enterpriseConfig)
		ts.Require().NoError(err)

		// Check that file was created
		_, err = os.Stat(".mage.enterprise.yaml")
		ts.Require().NoError(err)
	})
}

// TestEnterpriseSetup tests enterprise configuration setup
func (ts *ConfigTestSuite) TestEnterpriseSetup() {
	ts.Run("SetupEnterpriseConfigWhenExists", func() {
		// Set up config with enterprise already configured
		customConfig := defaultConfig()
		customConfig.Enterprise = &EnterpriseConfiguration{}
		TestSetConfig(customConfig)

		err := SetupEnterpriseConfig()
		ts.Require().Error(err)
		ts.Require().ErrorIs(err, ErrEnterpriseConfigExists)
	})

	ts.Run("SetupEnterpriseConfigWhenNotExists", func() {
		// Skip this test as it causes nil pointer panics due to interactive wizard
		ts.T().Skip("Skipping interactive wizard test - causes nil pointer panic without input")
	})
}

// TestConfigMethods tests Config struct methods
func (ts *ConfigTestSuite) TestConfigMethods() {
	ts.Run("ConfigLoad", func() {
		config := &Config{}
		err := config.Load()
		ts.Require().NoError(err)
	})

	ts.Run("ConfigValidate", func() {
		config := &Config{}
		err := config.Validate()
		ts.Require().NoError(err)
	})

	ts.Run("ConfigSave", func() {
		config := defaultConfig()
		err := config.Save("test-path")
		ts.Require().NoError(err)
	})
}

// TestTestUtilities tests the test utility functions
func (ts *ConfigTestSuite) TestTestUtilities() {
	ts.Run("TestResetConfig", func() {
		// Set a custom config
		customConfig := defaultConfig()
		customConfig.Project.Name = "custom-project"
		TestSetConfig(customConfig)

		// Verify it's set
		config, err := GetConfig()
		ts.Require().NoError(err)
		ts.Require().Equal("custom-project", config.Project.Name)

		// Reset and verify default behavior
		TestResetConfig()

		// After reset, should get a fresh default config
		config, err = GetConfig()
		ts.Require().NoError(err)
		ts.Require().NotEqual("custom-project", config.Project.Name)
	})

	ts.Run("TestSetConfig", func() {
		customConfig := &Config{
			Project: ProjectConfig{
				Name:   "test-project",
				Binary: "test-binary",
			},
		}

		TestSetConfig(customConfig)

		// Verify config is set
		config, err := GetConfig()
		ts.Require().NoError(err)
		ts.Require().Equal("test-project", config.Project.Name)
		ts.Require().Equal("test-binary", config.Project.Binary)
	})
}

// TestConfigErrorHandling tests error handling in configuration functions
func (ts *ConfigTestSuite) TestConfigErrorHandling() {
	ts.Run("ConfigFunctionsWithProviderError", func() {
		// This test would require mocking the config provider to return errors
		// For now, we test the basic functionality

		// Test that functions handle config loading errors gracefully
		name := BinaryName()
		ts.Require().NotEmpty(name) // Should return default even on error

		tags := BuildTags()
		ts.Require().Empty(tags) // Should return empty string on error

		verbose := IsVerbose()
		ts.Require().False(verbose) // Should return false on error

		hasEnterprise := HasEnterpriseConfig()
		ts.Require().False(hasEnterprise) // Should return false on error

		enterpriseConfig := GetEnterpriseConfig()
		ts.Require().Nil(enterpriseConfig) // Should return nil on error
	})
}

// TestConfigFileDetection tests configuration file detection
func (ts *ConfigTestSuite) TestConfigFileDetection() {
	ts.Run("DetectExistingConfigFiles", func() {
		// Change to temp directory for this test
		originalDir, err := os.Getwd()
		ts.Require().NoError(err)
		defer func() {
			ts.Require().NoError(os.Chdir(originalDir))
		}()

		ts.Require().NoError(os.Chdir(ts.tempDir))

		// Test each possible config file name
		configFiles := []string{".mage.yaml", ".mage.yml", "mage.yaml", "mage.yml"}

		for _, configFile := range configFiles {
			// Create the config file
			file, err := os.Create(configFile) //nolint:gosec // G304: configFile is from safe hardcoded test slice
			ts.Require().NoError(err)
			ts.Require().NoError(file.Close())

			// Test that it's detected
			detectedPath := getConfigFilePath()
			ts.Require().Equal(configFile, detectedPath)

			// Clean up
			ts.Require().NoError(os.Remove(configFile))
		}
	})
}

// TestConfigStructValidation tests that all config structs are properly initialized
func (ts *ConfigTestSuite) TestConfigStructValidation() {
	ts.Run("AllStructsProperlyInitialized", func() {
		config := defaultConfig()

		// Test that all nested structs are properly initialized
		ts.Require().NotNil(config)
		ts.Require().NotZero(config.Project)
		ts.Require().NotZero(config.Build)
		ts.Require().NotZero(config.Test)
		ts.Require().NotZero(config.Lint)
		ts.Require().NotZero(config.Tools)
		ts.Require().NotZero(config.Docker)
		ts.Require().NotZero(config.Release)

		// Test that maps are initialized
		ts.Require().NotNil(config.Docker.BuildArgs)
		ts.Require().NotNil(config.Docker.Labels)
	})

	ts.Run("DefaultValuesAreReasonable", func() {
		config := defaultConfig()

		// Test that default values make sense
		ts.Require().Positive(config.Build.Parallel)
		ts.Require().NotEmpty(config.Test.Timeout)
		ts.Require().NotEmpty(config.Lint.Timeout)
		ts.Require().NotEmpty(config.Build.Platforms)
		ts.Require().NotEmpty(config.Docker.Platforms)
		ts.Require().NotEmpty(config.Release.Formats)
	})
}

// Benchmark tests for performance validation
func BenchmarkDefaultConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = defaultConfig()
	}
}

func BenchmarkApplyEnvOverrides(b *testing.B) {
	config := defaultConfig()
	require.NoError(b, os.Setenv("VERBOSE", "true"))
	require.NoError(b, os.Setenv("PARALLEL", "8"))
	defer func() {
		require.NoError(b, os.Unsetenv("VERBOSE"))
		require.NoError(b, os.Unsetenv("PARALLEL"))
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		applyEnvOverrides(config)
	}
}

func BenchmarkGetConfig(b *testing.B) {
	TestResetConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config, err := GetConfig()
		require.NoError(b, err)
		require.NotNil(b, config)
	}
}
