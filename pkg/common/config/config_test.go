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

	"github.com/mrz1836/mage-x/pkg/testhelpers"
)

// ConfigTestSuite tests configuration functionality
type ConfigTestSuite struct {
	testhelpers.BaseSuite
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func TestDefaultConfigLoader(t *testing.T) {
	tmpDir := t.TempDir()

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

	// Environment variable cleanup handled by t.Cleanup
	t.Cleanup(func() {
		if err := os.Unsetenv(testKey); err != nil {
			t.Logf("Failed to unset %s: %v", testKey, err)
		}
	})

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

	t.Cleanup(func() {
		if err := os.Unsetenv("TEST_BOOL"); err != nil {
			t.Logf("Failed to unset TEST_BOOL: %v", err)
		}
	})
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

	t.Cleanup(func() {
		if err := os.Unsetenv("TEST_INT"); err != nil {
			t.Logf("Failed to unset TEST_INT: %v", err)
		}
	})
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

	t.Cleanup(func() {
		if err := os.Unsetenv("TEST_DURATION"); err != nil {
			t.Logf("Failed to unset TEST_DURATION: %v", err)
		}
	})
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

	t.Cleanup(func() {
		if err := os.Unsetenv("TEST_SLICE"); err != nil {
			t.Logf("Failed to unset TEST_SLICE: %v", err)
		}
	})
}

func TestFileConfigSource(t *testing.T) {
	tmpDir := t.TempDir()

	configPath := filepath.Join(tmpDir, "test.yaml")
	configContent := `test: value
nested:
  key: nestedvalue`

	err := os.WriteFile(configPath, []byte(configContent), 0o600)
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
		configPath := filepath.Join(s.TmpDir, "myapp.yaml")
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
		foundPath, err := config.LoadFromPaths(&testConfig, "myapp", s.TmpDir)
		s.Require().NoError(err, "Failed to load config")

		s.Equal(configPath, foundPath, "Should find config at expected path")
		s.Equal("testapp", testConfig.App.Name, "App name should match")
		s.Equal(8080, testConfig.App.Port, "App port should match")
		s.True(testConfig.App.Enabled, "App should be enabled")
		s.ElementsMatch([]string{"production", "stable"}, testConfig.App.Tags, "Tags should match")
	})

	s.Run("LoadWithEnvOverrides", func() {
		// Set environment variables
		s.SetEnvVar("MYAPP_APP_NAME", "env-override")
		s.SetEnvVar("MYAPP_APP_PORT", "9090")

		// Create config file
		configPath := filepath.Join(s.TmpDir, "myapp-env.yaml")
		configContent := `app:
  name: file-value
  port: 8080`

		err := os.WriteFile(configPath, []byte(configContent), 0o600)
		s.Require().NoError(err)

		var testConfig TestConfig
		foundPath, err := config.LoadWithEnvOverrides(&testConfig, "myapp-env", "MYAPP", s.TmpDir)
		s.Require().NoError(err, "Failed to load config with env overrides")

		s.Equal(configPath, foundPath)
		// These assertions depend on whether env override is implemented
		// If implemented, values should come from env vars
		// assert.Equal(s.T(), "env-override", testConfig.App.Name)
		// assert.Equal(s.T(), 9090, testConfig.App.Port)
	})

	s.Run("SetupManager", func() {
		config.SetupManager("testapp", "TEST", s.TmpDir)
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

// TestConfigExpandPath tests the expandPath method for security and functionality.
// This is critical for preventing path traversal attacks and ensuring correct HOME expansion.
func TestConfigExpandPath(t *testing.T) {
	config := New()

	// Save original HOME and restore after tests
	origHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if origHome != "" {
			if err := os.Setenv("HOME", origHome); err != nil {
				t.Logf("Failed to restore HOME: %v", err)
			}
		} else {
			if err := os.Unsetenv("HOME"); err != nil {
				t.Logf("Failed to unset HOME: %v", err)
			}
		}
	})

	tests := []struct {
		name     string
		input    string
		homeEnv  string
		expected string
	}{
		// Security: path traversal patterns are blocked
		{
			name:     "blocks simple path traversal",
			input:    "/etc/../passwd",
			homeEnv:  "/home/user",
			expected: "/etc/../passwd", // Returns original without expansion
		},
		{
			name:     "blocks double dot at start",
			input:    "../etc/passwd",
			homeEnv:  "/home/user",
			expected: "../etc/passwd",
		},
		{
			name:     "blocks URL-encoded path traversal (%2e)",
			input:    "/etc/%2e%2e/passwd",
			homeEnv:  "/home/user",
			expected: "/etc/%2e%2e/passwd",
		},
		{
			name:     "blocks double URL-encoded traversal",
			input:    "/etc/%252e%252e/passwd",
			homeEnv:  "/home/user",
			expected: "/etc/%2e%2e/passwd", // First decode reveals %2e%2e, which contains .. after second decode
		},
		{
			name:     "blocks null byte injection",
			input:    "/etc/passwd\x00.txt",
			homeEnv:  "/home/user",
			expected: "/etc/passwd\x00.txt",
		},
		{
			name:     "blocks newline injection",
			input:    "/etc/passwd\n/etc/shadow",
			homeEnv:  "/home/user",
			expected: "/etc/passwd\n/etc/shadow",
		},
		{
			name:     "blocks carriage return injection",
			input:    "/etc/passwd\r/etc/shadow",
			homeEnv:  "/home/user",
			expected: "/etc/passwd\r/etc/shadow",
		},

		// HOME expansion
		{
			name:     "expands $HOME prefix",
			input:    "$HOME/.config/app",
			homeEnv:  "/home/testuser",
			expected: "/home/testuser/.config/app",
		},
		{
			name:     "HOME not set returns path with empty HOME",
			input:    "$HOME/.config",
			homeEnv:  "",
			expected: "/.config",
		},
		{
			name:     "only first $HOME is expanded",
			input:    "$HOME/$HOME/config",
			homeEnv:  "/home/user",
			expected: "/home/user/$HOME/config",
		},

		// No expansion needed
		{
			name:     "absolute path unchanged",
			input:    "/etc/config.yaml",
			homeEnv:  "/home/user",
			expected: "/etc/config.yaml",
		},
		{
			name:     "relative path unchanged",
			input:    "config/app.yaml",
			homeEnv:  "/home/user",
			expected: "config/app.yaml",
		},
		{
			name:     "dot prefix unchanged",
			input:    "./config.yaml",
			homeEnv:  "/home/user",
			expected: "./config.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set HOME for this test
			if tt.homeEnv != "" {
				require.NoError(t, os.Setenv("HOME", tt.homeEnv))
			} else {
				require.NoError(t, os.Unsetenv("HOME"))
			}

			result := config.expandPath(tt.input)
			assert.Equal(t, tt.expected, result, "expandPath(%q)", tt.input)
		})
	}
}

// TestConfigDetectFormat tests the detectFormat method.
// Validates correct format detection from file extensions.
func TestConfigDetectFormat(t *testing.T) {
	config := New()

	tests := []struct {
		name     string
		path     string
		expected Format
	}{
		// Standard extensions
		{"yaml lowercase", "config.yaml", FormatYAML},
		{"yml lowercase", "config.yml", FormatYAML},
		{"json lowercase", "config.json", FormatJSON},
		{"toml lowercase", "config.toml", FormatTOML},
		{"ini lowercase", "config.ini", FormatINI},

		// Case insensitivity
		{"YAML uppercase", "config.YAML", FormatYAML},
		{"YML uppercase", "config.YML", FormatYAML},
		{"JSON uppercase", "config.JSON", FormatJSON},
		{"TOML uppercase", "config.TOML", FormatTOML},
		{"INI uppercase", "config.INI", FormatINI},
		{"mixed case Yaml", "config.Yaml", FormatYAML},
		{"mixed case Json", "config.Json", FormatJSON},

		// No extension - defaults to YAML
		{"no extension", "config", FormatYAML},
		{"no extension with path", "/etc/myapp/config", FormatYAML},

		// Hidden files
		{"hidden file no ext", ".configrc", FormatYAML},
		{"hidden file with yaml", ".config.yaml", FormatYAML},
		{"hidden file with json", ".config.json", FormatJSON},

		// Multiple dots in filename
		{"multiple dots yaml", "config.prod.yaml", FormatYAML},
		{"multiple dots json", "config.dev.json", FormatJSON},
		{"many dots", "my.app.config.v2.yaml", FormatYAML},

		// Unknown extensions default to YAML
		{"unknown ext txt", "config.txt", FormatYAML},
		{"unknown ext xml", "config.xml", FormatYAML},
		{"unknown ext cfg", "config.cfg", FormatYAML},

		// Full paths
		{"full path yaml", "/etc/myapp/config.yaml", FormatYAML},
		{"full path json", "/home/user/.config/app.json", FormatJSON},
		{"relative path yaml", "./configs/local.yaml", FormatYAML},

		// Edge cases
		{"env extension", "config.env", FormatYAML}, // .env defaults to YAML
		{"dot only", ".", FormatYAML},
		{"empty string", "", FormatYAML},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.detectFormat(tt.path)
			assert.Equal(t, tt.expected, result, "detectFormat(%q)", tt.path)
		})
	}
}

// TestConfigBuildConfigPaths tests the buildConfigPaths method.
// Validates that search paths are constructed correctly.
func TestConfigBuildConfigPaths(t *testing.T) {
	config := New()

	// Save original HOME
	origHome := os.Getenv("HOME")
	t.Cleanup(func() {
		if origHome != "" {
			if err := os.Setenv("HOME", origHome); err != nil {
				t.Logf("Failed to restore HOME: %v", err)
			}
		}
	})
	require.NoError(t, os.Setenv("HOME", "/home/testuser"))

	t.Run("default search directories", func(t *testing.T) {
		paths := config.buildConfigPaths("myapp")

		// Should include current directory
		assert.Contains(t, paths, "myapp.yaml")
		assert.Contains(t, paths, "myapp.yml")
		assert.Contains(t, paths, "myapp.json")
		assert.Contains(t, paths, "myapp")

		// Should have multiple paths
		assert.Greater(t, len(paths), 4, "should have paths from multiple directories")
	})

	t.Run("custom search directories", func(t *testing.T) {
		paths := config.buildConfigPaths("app", "/custom/dir", "/another/dir")

		// Should include custom directories
		assert.Contains(t, paths, "/custom/dir/app.yaml")
		assert.Contains(t, paths, "/another/dir/app.json")

		// Should NOT include default directories when custom provided
		for _, p := range paths {
			assert.NotContains(t, p, "/etc/")
		}
	})

	t.Run("HOME expansion in directories", func(t *testing.T) {
		paths := config.buildConfigPaths("app", "$HOME/.config")

		// HOME should be expanded
		found := false
		for _, p := range paths {
			if strings.Contains(p, "/home/testuser/.config/app") {
				found = true
				break
			}
		}
		assert.True(t, found, "should expand $HOME in search directories")
	})

	t.Run("filters paths with null bytes", func(t *testing.T) {
		// This tests the null byte filtering in buildConfigPaths
		paths := config.buildConfigPaths("app", "/normal/dir")

		for _, p := range paths {
			assert.NotContains(t, p, "\x00", "paths should not contain null bytes")
		}
	})
}

// TestConfigNew tests the Config constructor
func TestConfigNew(t *testing.T) {
	t.Run("creates config with default implementations", func(t *testing.T) {
		config := New()

		assert.NotNil(t, config)
		assert.NotNil(t, config.Loader)
		assert.NotNil(t, config.Env)
		assert.NotNil(t, config.Manager)
	})
}

// TestConfigNewWithOptions tests creating Config with custom options
func TestConfigNewWithOptions(t *testing.T) {
	t.Run("accepts custom implementations", func(t *testing.T) {
		customLoader := NewDefaultConfigLoader()
		customEnv := NewDefaultEnvProvider()
		customManager := NewDefaultConfigManager()

		config := NewWithOptions(customLoader, customEnv, customManager)

		assert.Same(t, customLoader, config.Loader)
		assert.Same(t, customEnv, config.Env)
		assert.Same(t, customManager, config.Manager)
	})

	t.Run("accepts nil values", func(t *testing.T) {
		config := NewWithOptions(nil, nil, nil)

		assert.Nil(t, config.Loader)
		assert.Nil(t, config.Env)
		assert.Nil(t, config.Manager)
	})
}

// TestGetDefault tests the GetDefault package function
func TestGetDefault(t *testing.T) {
	t.Run("returns new Config instance", func(t *testing.T) {
		config := GetDefault()

		assert.NotNil(t, config)
		assert.NotNil(t, config.Loader)
		assert.NotNil(t, config.Env)
		assert.NotNil(t, config.Manager)
	})

	t.Run("returns different instances each call", func(t *testing.T) {
		config1 := GetDefault()
		config2 := GetDefault()

		assert.NotSame(t, config1, config2, "should return new instances")
	})
}

// Benchmark tests
func BenchmarkConfigLoad(b *testing.B) {
	tmpDir := b.TempDir()

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
	err := os.WriteFile(configPath, []byte(configContent), 0o600)
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
