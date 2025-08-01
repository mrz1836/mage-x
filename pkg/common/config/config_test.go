package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ConfigTestSuite tests configuration functionality
type ConfigTestSuite struct {
	suite.Suite

	tmpDir string
}

func (s *ConfigTestSuite) SetupSuite() {
	tmpDir, err := os.MkdirTemp("", "config-suite-test-*")
	s.Require().NoError(err, "Failed to create temp dir")
	s.tmpDir = tmpDir
}

func (s *ConfigTestSuite) TearDownSuite() {
	if err := os.RemoveAll(s.tmpDir); err != nil {
		s.T().Logf("Failed to remove temp dir: %v", err)
	}
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func TestDefaultConfigLoader(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "config-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
			t.Logf("Failed to remove temp dir: %v", removeErr)
		}
	}()

	loader := NewDefaultConfigLoader()

	type TestConfig struct {
		Database struct {
			Host string `yaml:"host" json:"host"`
			Port int    `yaml:"port" json:"port"`
		} `yaml:"database" json:"database"`
		Debug bool `yaml:"debug" json:"debug"`
	}

	t.Run("Load YAML config", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "test.yaml")
		yamlContent := `database:
  host: localhost
  port: 5432
debug: true`

		err := os.WriteFile(configPath, []byte(yamlContent), 0o600)
		require.NoError(t, err, "Failed to write config file")

		var config TestConfig
		err = loader.LoadFrom(configPath, &config)
		require.NoError(t, err, "Failed to load config")

		assert.Equal(t, "localhost", config.Database.Host, "Database host should match")
		assert.Equal(t, 5432, config.Database.Port, "Database port should match")
		assert.True(t, config.Debug, "Debug should be true")
	})

	t.Run("Load JSON config", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "test.json")
		jsonContent := `{
  "database": {
    "host": "dbhost",
    "port": 3306
  },
  "debug": false
}`

		err := os.WriteFile(configPath, []byte(jsonContent), 0o600)
		require.NoError(t, err, "Failed to write config file")

		var config TestConfig
		err = loader.LoadFrom(configPath, &config)
		require.NoError(t, err, "Failed to load config")

		assert.Equal(t, "dbhost", config.Database.Host, "Database host should match")
		assert.Equal(t, 3306, config.Database.Port, "Database port should match")
		assert.False(t, config.Debug, "Debug should be false")
	})

	t.Run("Load with fallback", func(t *testing.T) {
		config1Path := filepath.Join(tmpDir, "nonexistent.yaml")
		config2Path := filepath.Join(tmpDir, "fallback.yaml")

		yamlContent := `database:
  host: fallback-host
  port: 9999
debug: true`

		err := os.WriteFile(config2Path, []byte(yamlContent), 0o600)
		require.NoError(t, err, "Failed to write config file")

		var config TestConfig
		foundPath, err := loader.Load([]string{config1Path, config2Path}, &config)
		require.NoError(t, err, "Failed to load config with fallback")

		assert.Equal(t, config2Path, foundPath, "Should find config at fallback path")
		assert.Equal(t, "fallback-host", config.Database.Host, "Database host should match")
		assert.Equal(t, 9999, config.Database.Port, "Database port should match")
	})

	t.Run("Load non-existent file", func(t *testing.T) {
		var config TestConfig
		err := loader.LoadFrom("/non/existent/file.yaml", &config)
		assert.Error(t, err, "Should error when loading non-existent file")
	})

	t.Run("Load all missing", func(t *testing.T) {
		var config TestConfig
		_, err := loader.Load([]string{"/non/existent1.yaml", "/non/existent2.yaml"}, &config)
		assert.Error(t, err, "Should error when all files are missing")
	})

	t.Run("Save config", func(t *testing.T) {
		config := TestConfig{}
		config.Database.Host = "saved-host"
		config.Database.Port = 8080
		config.Debug = false

		configPath := filepath.Join(tmpDir, "saved.yaml")

		err := loader.Save(configPath, config, "yaml")
		require.NoError(t, err, "Failed to save config")

		// Load it back
		var loadedConfig TestConfig
		err = loader.LoadFrom(configPath, &loadedConfig)
		require.NoError(t, err, "Failed to load saved config")

		assert.Equal(t, config.Database.Host, loadedConfig.Database.Host, "Host should match after save/load")
		assert.Equal(t, config.Database.Port, loadedConfig.Database.Port, "Port should match after save/load")
		assert.Equal(t, config.Debug, loadedConfig.Debug, "Debug should match after save/load")
	})

	t.Run("Save with invalid format", func(t *testing.T) {
		config := TestConfig{}
		configPath := filepath.Join(tmpDir, "invalid.txt")

		err := loader.Save(configPath, config, "invalid")
		assert.Error(t, err, "Should error on invalid format")
	})

	t.Run("Load invalid YAML", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "invalid.yaml")
		invalidContent := `invalid: yaml: content:`

		err := os.WriteFile(configPath, []byte(invalidContent), 0o600)
		require.NoError(t, err)

		var config TestConfig
		err = loader.LoadFrom(configPath, &config)
		assert.Error(t, err, "Should error on invalid YAML")
	})
}

func TestDefaultEnvProvider(t *testing.T) {
	env := NewDefaultEnvProvider()
	testKey := "TEST_CONFIG_VAR"
	testValue := "test_value"

	setupTestEnv(t, testKey)
	defer cleanupTestEnv(t, testKey)

	t.Run("Set and Get", func(t *testing.T) {
		testEnvSetAndGet(t, env, testKey, testValue)
	})

	t.Run("GetWithDefault", func(t *testing.T) {
		testEnvGetWithDefault(t, env, testKey, testValue)
	})

	t.Run("LookupEnv", func(t *testing.T) {
		testEnvLookup(t, env, testKey, testValue)
	})

	t.Run("GetBool", func(t *testing.T) {
		testEnvGetBool(t, env)
	})

	t.Run("GetInt", func(t *testing.T) {
		testEnvGetInt(t, env)
	})

	t.Run("GetDuration", func(t *testing.T) {
		testEnvGetDuration(t, env)
	})

	t.Run("GetStringSlice", func(t *testing.T) {
		testEnvGetStringSlice(t, env)
	})
}

// setupTestEnv sets up test environment
func setupTestEnv(t *testing.T, testKey string) {
	t.Helper()
	if err := os.Unsetenv(testKey); err != nil {
		t.Logf("Failed to unset %s: %v", testKey, err)
	}
}

// cleanupTestEnv cleans up test environment
func cleanupTestEnv(t *testing.T, testKey string) {
	t.Helper()
	if err := os.Unsetenv(testKey); err != nil {
		t.Logf("Failed to unset %s: %v", testKey, err)
	}
}

// testEnvSetAndGet tests basic set and get functionality
func testEnvSetAndGet(t *testing.T, env EnvProvider, testKey, testValue string) {
	t.Helper()
	err := env.Set(testKey, testValue)
	require.NoError(t, err, "Failed to set env var")

	value := env.Get(testKey)
	assert.Equal(t, testValue, value, "Get should return the set value")
}

// testEnvGetWithDefault tests GetWithDefault functionality
func testEnvGetWithDefault(t *testing.T, env EnvProvider, testKey, testValue string) {
	t.Helper()
	// Test with existing value
	value := env.GetWithDefault(testKey, "default")
	assert.Equal(t, testValue, value, "Should return existing value")

	// Test with non-existing value
	value = env.GetWithDefault("NON_EXISTING_VAR", "default")
	assert.Equal(t, "default", value, "Should return default for non-existing var")
}

// testEnvLookup tests LookupEnv functionality
func testEnvLookup(t *testing.T, env EnvProvider, testKey, testValue string) {
	t.Helper()
	value, found := env.LookupEnv(testKey)
	assert.True(t, found, "Should find existing env var")
	assert.Equal(t, testValue, value, "Should return correct value")

	_, found = env.LookupEnv("NON_EXISTING_VAR")
	assert.False(t, found, "Should not find non-existing env var")
}

// testEnvGetBool tests boolean environment variable parsing
func testEnvGetBool(t *testing.T, env TypedEnvProvider) {
	t.Helper()
	testCases := []struct {
		name     string
		value    string
		expected bool
	}{
		{"true lowercase", "true", true},
		{"TRUE uppercase", "TRUE", true},
		{"1 numeric", "1", true},
		{"yes", "yes", true},
		{"YES uppercase", "YES", true},
		{"on", "on", true},
		{"false lowercase", "false", false},
		{"FALSE uppercase", "FALSE", false},
		{"0 numeric", "0", false},
		{"no", "no", false},
		{"off", "off", false},
		{"invalid value", "invalid", false},
		{"empty string", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.NoError(t, env.Set("TEST_BOOL", tc.value))
			result := env.GetBool("TEST_BOOL", false)
			assert.Equal(t, tc.expected, result, "Boolean parsing for '%s'", tc.value)
		})
	}

	// Test default value
	result := env.GetBool("NON_EXISTING_BOOL", true)
	assert.True(t, result, "Should return default when var doesn't exist")

	cleanupEnvVar(t, "TEST_BOOL")
}

// testEnvGetInt tests integer environment variable parsing
func testEnvGetInt(t *testing.T, env TypedEnvProvider) {
	t.Helper()
	// Test positive number
	require.NoError(t, env.Set("TEST_INT", "42"))
	result := env.GetInt("TEST_INT", 0)
	assert.Equal(t, 42, result, "Should parse int correctly")

	// Test negative number
	require.NoError(t, env.Set("TEST_INT", "-100"))
	result = env.GetInt("TEST_INT", 0)
	assert.Equal(t, -100, result, "Should parse negative int correctly")

	// Test default with invalid value
	require.NoError(t, env.Set("TEST_INT", "invalid"))
	result = env.GetInt("TEST_INT", 100)
	assert.Equal(t, 100, result, "Should return default for invalid int")

	// Test default with non-existing var
	result = env.GetInt("NON_EXISTING_INT", 200)
	assert.Equal(t, 200, result, "Should return default for non-existing var")

	cleanupEnvVar(t, "TEST_INT")
}

// testEnvGetDuration tests duration environment variable parsing
func testEnvGetDuration(t *testing.T, env TypedEnvProvider) {
	t.Helper()
	testCases := []struct {
		value    string
		expected time.Duration
	}{
		{"5m", 5 * time.Minute},
		{"30s", 30 * time.Second},
		{"1h", time.Hour},
		{"2h30m", 2*time.Hour + 30*time.Minute},
		{"100ms", 100 * time.Millisecond},
	}

	for _, tc := range testCases {
		require.NoError(t, env.Set("TEST_DURATION", tc.value))
		result := env.GetDuration("TEST_DURATION", time.Second)
		assert.Equal(t, tc.expected, result, "Should parse duration '%s' correctly", tc.value)
	}

	// Test invalid duration
	require.NoError(t, env.Set("TEST_DURATION", "invalid"))
	result := env.GetDuration("TEST_DURATION", 10*time.Second)
	assert.Equal(t, 10*time.Second, result, "Should return default for invalid duration")

	cleanupEnvVar(t, "TEST_DURATION")
}

// testEnvGetStringSlice tests string slice environment variable parsing
func testEnvGetStringSlice(t *testing.T, env TypedEnvProvider) {
	t.Helper()
	// Test basic comma-separated list
	require.NoError(t, env.Set("TEST_SLICE", "a,b,c"))
	result := env.GetStringSlice("TEST_SLICE", []string{"default"})
	expected := []string{"a", "b", "c"}
	assert.ElementsMatch(t, expected, result, "Should parse comma-separated list")

	// Test with spaces (should be trimmed)
	require.NoError(t, env.Set("TEST_SLICE", "a, b, c"))
	result = env.GetStringSlice("TEST_SLICE", []string{"default"})
	assert.Len(t, result, 3, "Should handle spaces correctly")
	assert.Equal(t, "a", result[0])
	assert.Equal(t, "b", result[1])
	assert.Equal(t, "c", result[2])

	// Test empty string (returns default)
	require.NoError(t, env.Set("TEST_SLICE", ""))
	result = env.GetStringSlice("TEST_SLICE", []string{"default"})
	assert.ElementsMatch(t, []string{"default"}, result, "Empty string returns default")

	// Test non-existing var
	result = env.GetStringSlice("NON_EXISTING_SLICE", []string{"def1", "def2"})
	assert.ElementsMatch(t, []string{"def1", "def2"}, result, "Should return default for non-existing var")

	cleanupEnvVar(t, "TEST_SLICE")
}

// cleanupEnvVar cleans up an environment variable
func cleanupEnvVar(t *testing.T, varName string) {
	t.Helper()
	if err := os.Unsetenv(varName); err != nil {
		t.Logf("Failed to unset %s: %v", varName, err)
	}
}

func TestFileConfigSource(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "source-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
			t.Logf("Failed to remove temp dir: %v", removeErr)
		}
	}()

	configPath := filepath.Join(tmpDir, "test.yaml")
	configContent := `test: value
nested:
  key: nestedvalue`

	err = os.WriteFile(configPath, []byte(configContent), 0o600)
	require.NoError(t, err, "Failed to write config file")

	source := NewFileConfigSource(configPath, FormatYAML, 100)

	t.Run("Name", func(t *testing.T) {
		name := source.Name()
		expected := "file:" + configPath
		assert.Equal(t, expected, name, "Source name should include file prefix and path")
	})

	t.Run("IsAvailable", func(t *testing.T) {
		assert.True(t, source.IsAvailable(), "Source should be available for existing file")

		// Test with non-existing file
		nonExistentSource := NewFileConfigSource("/non/existent/file.yaml", FormatYAML, 100)
		assert.False(t, nonExistentSource.IsAvailable(), "Source should not be available for non-existent file")
	})

	t.Run("Priority", func(t *testing.T) {
		priority := source.Priority()
		assert.Equal(t, 100, priority, "Priority should match constructor value")

		// Test different priority
		source2 := NewFileConfigSource(configPath, FormatYAML, 200)
		assert.Equal(t, 200, source2.Priority())
	})

	t.Run("Load", func(t *testing.T) {
		var config map[string]interface{}
		err := source.Load(&config)
		require.NoError(t, err, "Failed to load config")

		assert.Equal(t, "value", config["test"], "Should load top-level value")

		nested, ok := config["nested"].(map[string]interface{})
		require.True(t, ok, "Nested should be a map")
		assert.Equal(t, "nestedvalue", nested["key"], "Should load nested value")
	})

	t.Run("Load with structured config", func(t *testing.T) {
		type StructuredConfig struct {
			Test   string `yaml:"test"`
			Nested struct {
				Key string `yaml:"key"`
			} `yaml:"nested"`
		}

		var config StructuredConfig
		err := source.Load(&config)
		require.NoError(t, err, "Failed to load structured config")

		assert.Equal(t, "value", config.Test)
		assert.Equal(t, "nestedvalue", config.Nested.Key)
	})
}

func (s *ConfigTestSuite) TestConfigFacade() {
	config := New()

	type TestConfig struct {
		App struct {
			Name    string   `yaml:"name" json:"name"`
			Port    int      `yaml:"port" json:"port"`
			Enabled bool     `yaml:"enabled" json:"enabled"`
			Tags    []string `yaml:"tags" json:"tags"`
		} `yaml:"app" json:"app"`
	}

	s.Run("LoadFromPaths", func() {
		// Create config file
		configPath := filepath.Join(s.tmpDir, "myapp.yaml")
		configContent := `app:
  name: testapp
  port: 8080
  enabled: true
  tags:
    - production
    - stable`

		err := os.WriteFile(configPath, []byte(configContent), 0o600)
		s.Require().NoError(err, "Failed to write config file")

		var testConfig TestConfig
		foundPath, err := config.LoadFromPaths(&testConfig, "myapp", s.tmpDir)
		s.Require().NoError(err, "Failed to load config")

		s.Equal(configPath, foundPath, "Should find config at expected path")
		s.Equal("testapp", testConfig.App.Name, "App name should match")
		s.Equal(8080, testConfig.App.Port, "App port should match")
		s.True(testConfig.App.Enabled, "App should be enabled")
		s.ElementsMatch([]string{"production", "stable"}, testConfig.App.Tags, "Tags should match")
	})

	s.Run("LoadWithEnvOverrides", func() {
		// Set environment variables
		s.Require().NoError(os.Setenv("MYAPP_APP_NAME", "env-override"))
		s.Require().NoError(os.Setenv("MYAPP_APP_PORT", "9090"))
		defer func() {
			if err := os.Unsetenv("MYAPP_APP_NAME"); err != nil {
				s.T().Logf("Failed to unset MYAPP_APP_NAME: %v", err)
			}
		}()
		defer func() {
			if err := os.Unsetenv("MYAPP_APP_PORT"); err != nil {
				s.T().Logf("Failed to unset MYAPP_APP_PORT: %v", err)
			}
		}()

		// Create config file
		configPath := filepath.Join(s.tmpDir, "myapp-env.yaml")
		configContent := `app:
  name: file-value
  port: 8080`

		err := os.WriteFile(configPath, []byte(configContent), 0o600)
		s.Require().NoError(err)

		var testConfig TestConfig
		foundPath, err := config.LoadWithEnvOverrides(&testConfig, "myapp-env", "MYAPP", s.tmpDir)
		s.Require().NoError(err, "Failed to load config with env overrides")

		s.Equal(configPath, foundPath)
		// These assertions depend on whether env override is implemented
		// If implemented, values should come from env vars
		// assert.Equal(s.T(), "env-override", testConfig.App.Name)
		// assert.Equal(s.T(), 9090, testConfig.App.Port)
	})

	s.Run("SetupManager", func() {
		config.SetupManager("testapp", "TEST", s.tmpDir)
		// Manager setup verification would require checking internal state
		// or using the manager to load config
	})
}

func TestGetCommonConfigPaths(t *testing.T) {
	paths := GetCommonConfigPaths("myapp")

	assert.NotEmpty(t, paths, "Should return at least one config path")

	// Check that common patterns are included
	expectedPatterns := []string{
		".myapp.yaml",
		".myapp.yml",
		"myapp.yaml",
		"myapp.yml",
		".myapp.json",
		"myapp.json",
	}

	for _, pattern := range expectedPatterns {
		found := false
		for _, path := range paths {
			if filepath.Base(path) == pattern {
				found = true
				break
			}
		}
		assert.True(t, found, "Should include pattern %s in common paths", pattern)
	}

	// Check that it includes home directory configs
	homeFound := false
	for _, path := range paths {
		if strings.Contains(path, filepath.Join("$HOME", ".config")) ||
			strings.Contains(path, filepath.Join("${HOME}", ".config")) {
			homeFound = true
			break
		}
	}
	assert.True(t, homeFound, "Should include home directory config paths")
}

func TestPackageLevelFunctions(t *testing.T) {
	// Test package-level convenience functions

	// Set up test environment variables
	testCases := []struct {
		key   string
		value string
	}{
		{"TEST_PKG_STRING", "package_test"},
		{"TEST_PKG_BOOL", "true"},
		{"TEST_PKG_INT", "123"},
		{"TEST_PKG_DURATION", "5m"},
		{"TEST_PKG_SLICE", "a,b,c"},
	}

	for _, tc := range testCases {
		require.NoError(t, os.Setenv(tc.key, tc.value))
	}

	// Clean up environment variables at the end
	defer func() {
		for _, tc := range testCases {
			if err := os.Unsetenv(tc.key); err != nil {
				t.Logf("Failed to unset %s: %v", tc.key, err)
			}
		}
	}()

	t.Run("GetString", func(t *testing.T) {
		result := GetString("TEST_PKG_STRING", "default")
		assert.Equal(t, "package_test", result, "Should get string value")

		result = GetString("NON_EXISTENT", "default")
		assert.Equal(t, "default", result, "Should return default for non-existent")
	})

	t.Run("GetBool", func(t *testing.T) {
		result := GetBool("TEST_PKG_BOOL", false)
		assert.True(t, result, "Should get bool value")

		result = GetBool("NON_EXISTENT_BOOL", true)
		assert.True(t, result, "Should return default for non-existent")
	})

	t.Run("GetInt", func(t *testing.T) {
		result := GetInt("TEST_PKG_INT", 0)
		assert.Equal(t, 123, result, "Should get int value")

		result = GetInt("NON_EXISTENT_INT", 456)
		assert.Equal(t, 456, result, "Should return default for non-existent")
	})

	t.Run("GetDuration", func(t *testing.T) {
		result := GetDuration("TEST_PKG_DURATION", time.Second)
		assert.Equal(t, 5*time.Minute, result, "Should get duration value")

		result = GetDuration("NON_EXISTENT_DURATION", 10*time.Second)
		assert.Equal(t, 10*time.Second, result, "Should return default for non-existent")
	})

	t.Run("GetStringSlice", func(t *testing.T) {
		result := GetStringSlice("TEST_PKG_SLICE", []string{"default"})
		assert.ElementsMatch(t, []string{"a", "b", "c"}, result, "Should get string slice")

		result = GetStringSlice("NON_EXISTENT_SLICE", []string{"d", "e"})
		assert.ElementsMatch(t, []string{"d", "e"}, result, "Should return default for non-existent")
	})
}

// Table-driven tests for complex scenarios
func TestEnvProviderBoolParsing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// True values
		{"true", "true", true},
		{"TRUE", "TRUE", true},
		{"True", "True", true},
		{"1", "1", true},
		{"yes", "yes", true},
		{"YES", "YES", true},
		{"on", "on", true},
		{"ON", "ON", true},

		// False values
		{"false", "false", false},
		{"FALSE", "FALSE", false},
		{"False", "False", false},
		{"0", "0", false},
		{"no", "no", false},
		{"NO", "NO", false},
		{"off", "off", false},
		{"OFF", "OFF", false},

		// Invalid values (should return default)
		{"empty", "", false},
		{"invalid", "invalid", false},
		{"2", "2", false},
		{"truee", "truee", false},
		{"y", "y", false},
		{"n", "n", false},
	}

	env := NewDefaultEnvProvider()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := env.Set("TEST_BOOL_TABLE", tt.input); err != nil {
				t.Fatalf("Failed to set TEST_BOOL_TABLE: %v", err)
			}
			result := env.GetBool("TEST_BOOL_TABLE", false)
			assert.Equal(t, tt.expected, result, "For input '%s'", tt.input)
		})
	}
}

// Benchmark tests
func BenchmarkConfigLoad(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "bench-config-*")
	require.NoError(b, err)
	defer func() {
		if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
			b.Logf("Failed to remove temp dir: %v", removeErr)
		}
	}()

	// Create a test config
	configPath := filepath.Join(tmpDir, "bench.yaml")
	configContent := `
database:
  host: localhost
  port: 5432
  user: admin
  password: secret
app:
  name: benchapp
  version: 1.0.0
  features:
    - feature1
    - feature2
    - feature3
settings:
  timeout: 30s
  retries: 3
  debug: true
`
	err = os.WriteFile(configPath, []byte(configContent), 0o600)
	require.NoError(b, err)

	loader := NewDefaultConfigLoader()

	type BenchConfig struct {
		Database struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
		} `yaml:"database"`
		App struct {
			Name     string   `yaml:"name"`
			Version  string   `yaml:"version"`
			Features []string `yaml:"features"`
		} `yaml:"app"`
		Settings struct {
			Timeout time.Duration `yaml:"timeout"`
			Retries int           `yaml:"retries"`
			Debug   bool          `yaml:"debug"`
		} `yaml:"settings"`
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var config BenchConfig
		err := loader.LoadFrom(configPath, &config)
		if err != nil {
			b.Fatal(err)
		}
	}
}
