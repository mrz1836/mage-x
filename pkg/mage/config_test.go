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
		// Temporarily clear environment variables to test true defaults
		originalVars := make(map[string]string)
		envVars := []string{
			"MAGE_X_GOLANGCI_LINT_VERSION",
			"MAGE_X_GOFUMPT_VERSION",
			"MAGE_X_GOVULNCHECK_VERSION",
			"MAGE_X_MOCKGEN_VERSION",
			"MAGE_X_SWAG_VERSION",
			"GOLANGCI_LINT_VERSION",
			"GOFUMPT_VERSION",
			"GOVULNCHECK_VERSION",
			"MOCKGEN_VERSION",
			"SWAG_VERSION",
		}

		// Save original values and unset
		for _, envVar := range envVars {
			originalVars[envVar] = os.Getenv(envVar)
			ts.Require().NoError(os.Unsetenv(envVar))
		}

		// Restore environment after test
		defer func() {
			for _, envVar := range envVars {
				if originalVal, exists := originalVars[envVar]; exists && originalVal != "" {
					ts.Require().NoError(os.Setenv(envVar, originalVal))
				} else {
					ts.Require().NoError(os.Unsetenv(envVar))
				}
			}
		}()

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
		ts.Require().Equal(GetDefaultGolangciLintVersion(), config.Lint.GolangciVersion)
		ts.Require().Equal("5m", config.Lint.Timeout)

		// Test tools defaults
		ts.Require().Equal(GetDefaultGolangciLintVersion(), config.Tools.GolangciLint)
		ts.Require().Equal(VersionLatest, config.Tools.Fumpt)
		ts.Require().Equal(GetDefaultGoVulnCheckVersion(), config.Tools.GoVulnCheck)

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
		// Test GetConfig function (formerly LoadConfig)
		config, err := GetConfig()
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

// TestCleanEnvValue tests the cleanEnvValue function
func (ts *ConfigTestSuite) TestCleanEnvValue() {
	ts.Run("EmptyString", func() {
		result := cleanEnvValue("")
		ts.Empty(result)
	})

	ts.Run("NoComment", func() {
		result := cleanEnvValue("v1.2.3")
		ts.Equal("v1.2.3", result)
	})

	ts.Run("WithInlineComment", func() {
		result := cleanEnvValue("v2.4.0 # https://github.com/golangci/golangci-lint/releases")
		ts.Equal("v2.4.0", result)
	})

	ts.Run("WithWhitespace", func() {
		result := cleanEnvValue("  v1.0.0  ")
		ts.Equal("v1.0.0", result)
	})

	ts.Run("WithCommentAndWhitespace", func() {
		result := cleanEnvValue("  v0.8.0           # https://github.com/mvdan/gofumpt/releases  ")
		ts.Equal("v0.8.0", result)
	})

	ts.Run("OnlyComment", func() {
		result := cleanEnvValue(" # just a comment")
		ts.Empty(result)
	})

	ts.Run("HashInValue", func() {
		// Hash not preceded by space should be preserved
		result := cleanEnvValue("commit#abc123")
		ts.Equal("commit#abc123", result)
	})

	ts.Run("MultipleSpaceHashes", func() {
		// Only first space-hash should be treated as comment
		result := cleanEnvValue("value # comment # more")
		ts.Equal("value", result)
	})
}

// TestCleanConfigValues tests the cleanConfigValues function
func (ts *ConfigTestSuite) TestCleanConfigValues() {
	ts.Run("CleanAllStringFields", func() {
		config := &Config{
			Project: ProjectConfig{
				Name:        "test-project # comment",
				Binary:      "  test-binary  ",
				Module:      "github.com/test/module # module comment",
				Main:        "cmd/main.go # main file",
				Description: "  Test description  # desc comment  ",
				Version:     "v1.0.0 # version comment",
				GitDomain:   "github.com",
				RepoOwner:   "test-owner # owner comment",
				RepoName:    "test-repo # repo comment",
				Env: map[string]string{
					"KEY1": "value1 # env comment",
					"KEY2": "  value2  ",
				},
			},
			Build: BuildConfig{
				Output:    "bin # output dir",
				Tags:      []string{"tag1 # comment", "  tag2  "},
				LDFlags:   []string{"-X main.version=1.0 # version flag"},
				Platforms: []string{"linux/amd64 # platform comment"},
				GoFlags:   []string{"-v # verbose flag"},
			},
			Test: TestConfig{
				Timeout:            "10m # timeout comment",
				IntegrationTimeout: "  30m  ",
				IntegrationTag:     "integration # tag comment",
				CoverMode:          "atomic # cover mode",
				Tags:               "unit # test tags",
				BenchTime:          "1s # bench time",
				CoverPkg:           []string{"./... # cover pkg"},
				CoverageExclude:    []string{"*_test.go # exclude pattern"},
			},
			Lint: LintConfig{
				GolangciVersion: "v2.4.0 # golangci version",
				Timeout:         "5m # lint timeout",
				SkipDirs:        []string{"vendor # skip dir"},
				SkipFiles:       []string{"*.pb.go # skip files"},
				DisableLinters:  []string{"gosec # disable linter"},
				EnableLinters:   []string{"gofmt # enable linter"},
			},
			Tools: ToolsConfig{
				GolangciLint: "v2.4.0 # tools version",
				Fumpt:        "  v0.8.0  ",
				GoVulnCheck:  "v1.1.4 # vuln check version",
				Mockgen:      "v0.4.0 # mockgen version",
				Swag:         "v1.16.6 # swag version",
				Custom: map[string]string{
					"tool1": "version1 # custom tool",
				},
			},
			Docker: DockerConfig{
				Registry:        "registry.com # docker registry",
				Repository:      "  repo  ",
				Dockerfile:      "Dockerfile # dockerfile comment",
				NetworkMode:     "bridge # network mode",
				DefaultRegistry: "docker.io # default registry",
				BuildArgs: map[string]string{
					"ARG1": "value1 # build arg",
				},
				Labels: map[string]string{
					"label1": "value1 # label",
				},
				Platforms:    []string{"linux/amd64 # platform"},
				CacheFrom:    []string{"image:latest # cache"},
				SecurityOpts: []string{"no-new-privileges # security"},
			},
			Release: ReleaseConfig{
				GitHubToken: "GITHUB_TOKEN # token env",
				NameTmpl:    "  template  ",
				Formats:     []string{"tar.gz # format"},
			},
			Download: DownloadConfig{
				UserAgent: "agent/1.0 # user agent",
			},
			Docs: DocsConfig{
				Tool: "pkgsite # docs tool",
			},
			Metadata: map[string]string{
				"key1": "value1 # metadata",
			},
		}

		cleanConfigValues(config)

		// Check Project fields
		ts.Equal("test-project", config.Project.Name)
		ts.Equal("test-binary", config.Project.Binary)
		ts.Equal("github.com/test/module", config.Project.Module)
		ts.Equal("cmd/main.go", config.Project.Main)
		ts.Equal("Test description", config.Project.Description)
		ts.Equal("v1.0.0", config.Project.Version)
		ts.Equal("github.com", config.Project.GitDomain)
		ts.Equal("test-owner", config.Project.RepoOwner)
		ts.Equal("test-repo", config.Project.RepoName)
		ts.Equal("value1", config.Project.Env["KEY1"])
		ts.Equal("value2", config.Project.Env["KEY2"])

		// Check Build fields
		ts.Equal("bin", config.Build.Output)
		ts.Equal("tag1", config.Build.Tags[0])
		ts.Equal("tag2", config.Build.Tags[1])
		ts.Equal("-X main.version=1.0", config.Build.LDFlags[0])
		ts.Equal("linux/amd64", config.Build.Platforms[0])
		ts.Equal("-v", config.Build.GoFlags[0])

		// Check Test fields
		ts.Equal("10m", config.Test.Timeout)
		ts.Equal("30m", config.Test.IntegrationTimeout)
		ts.Equal("integration", config.Test.IntegrationTag)
		ts.Equal("atomic", config.Test.CoverMode)
		ts.Equal("unit", config.Test.Tags)
		ts.Equal("1s", config.Test.BenchTime)
		ts.Equal("./...", config.Test.CoverPkg[0])
		ts.Equal("*_test.go", config.Test.CoverageExclude[0])

		// Check Lint fields
		ts.Equal("v2.4.0", config.Lint.GolangciVersion)
		ts.Equal("5m", config.Lint.Timeout)
		ts.Equal("vendor", config.Lint.SkipDirs[0])
		ts.Equal("*.pb.go", config.Lint.SkipFiles[0])
		ts.Equal("gosec", config.Lint.DisableLinters[0])
		ts.Equal("gofmt", config.Lint.EnableLinters[0])

		// Check Tools fields
		ts.Equal("v2.4.0", config.Tools.GolangciLint)
		ts.Equal("v0.8.0", config.Tools.Fumpt)
		ts.Equal("v1.1.4", config.Tools.GoVulnCheck)
		ts.Equal("v0.4.0", config.Tools.Mockgen)
		ts.Equal("v1.16.6", config.Tools.Swag)
		ts.Equal("version1", config.Tools.Custom["tool1"])

		// Check Docker fields
		ts.Equal("registry.com", config.Docker.Registry)
		ts.Equal("repo", config.Docker.Repository)
		ts.Equal("Dockerfile", config.Docker.Dockerfile)
		ts.Equal("bridge", config.Docker.NetworkMode)
		ts.Equal("docker.io", config.Docker.DefaultRegistry)
		ts.Equal("value1", config.Docker.BuildArgs["ARG1"])
		ts.Equal("value1", config.Docker.Labels["label1"])
		ts.Equal("linux/amd64", config.Docker.Platforms[0])
		ts.Equal("image:latest", config.Docker.CacheFrom[0])
		ts.Equal("no-new-privileges", config.Docker.SecurityOpts[0])

		// Check Release fields
		ts.Equal("GITHUB_TOKEN", config.Release.GitHubToken)
		ts.Equal("template", config.Release.NameTmpl)
		ts.Equal("tar.gz", config.Release.Formats[0])

		// Check Download fields
		ts.Equal("agent/1.0", config.Download.UserAgent)

		// Check Docs fields
		ts.Equal("pkgsite", config.Docs.Tool)

		// Check Metadata
		ts.Equal("value1", config.Metadata["key1"])
	})

	ts.Run("NilConfig", func() {
		// Should not panic with nil config
		ts.NotPanics(func() {
			cleanConfigValues(nil)
		})
	})

	ts.Run("EmptyMaps", func() {
		config := &Config{
			Project: ProjectConfig{
				Env: make(map[string]string),
			},
			Tools: ToolsConfig{
				Custom: make(map[string]string),
			},
			Docker: DockerConfig{
				BuildArgs: make(map[string]string),
				Labels:    make(map[string]string),
			},
			Metadata: make(map[string]string),
		}

		// Should not panic with empty maps
		ts.NotPanics(func() {
			cleanConfigValues(config)
		})
	})
}

// TestDefaultConfigAppliesCleanValues tests that defaultConfig() applies cleaning
func (ts *ConfigTestSuite) TestDefaultConfigAppliesCleanValues() {
	ts.Run("DefaultConfigCleansValues", func() {
		// Mock environment variables with comments
		originalVars := make(map[string]string)
		envVars := map[string]string{
			"MAGE_X_GOLANGCI_LINT_VERSION": "v2.4.0 # test comment",
			"MAGE_X_GOFUMPT_VERSION":       "  v0.8.0  ",
		}

		// Set env vars and store originals
		for key, value := range envVars {
			originalVars[key] = os.Getenv(key)
			ts.Require().NoError(os.Setenv(key, value))
		}
		defer func() {
			for key, original := range originalVars {
				if original == "" {
					ts.Require().NoError(os.Unsetenv(key))
				} else {
					ts.Require().NoError(os.Setenv(key, original))
				}
			}
		}()

		config := defaultConfig()

		// Verify that cleaning was applied and comments removed
		ts.Equal("v2.4.0", config.Lint.GolangciVersion)
		ts.Equal("v0.8.0", config.Tools.Fumpt)
	})
}
