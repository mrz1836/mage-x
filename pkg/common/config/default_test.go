package config

import (
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Test constants to satisfy goconst linter
const (
	testYAMLContent = `name: test
value: 42`
)

// DefaultConfigTestSuite defines the test suite for default config functions
type DefaultConfigTestSuite struct {
	suite.Suite

	tempDir string
}

// SetupTest runs before each test
func (ts *DefaultConfigTestSuite) SetupTest() {
	ts.tempDir = ts.T().TempDir()
}

// TestNewDefaultConfigLoader tests creating a new default config loader
func (ts *DefaultConfigTestSuite) TestNewDefaultConfigLoader() {
	loader := NewDefaultConfigLoader()
	ts.Require().NotNil(loader)
	ts.Require().NotNil(loader.fileOps)
	ts.Require().NotNil(loader.jsonOps)
	ts.Require().NotNil(loader.yamlOps)
}

// TestDefaultConfigLoader_LoadFrom tests loading from specific files
func (ts *DefaultConfigTestSuite) TestDefaultConfigLoader_LoadFrom() {
	loader := NewDefaultConfigLoader()

	ts.Run("load JSON file", func() {
		jsonFile := filepath.Join(ts.tempDir, "config.json")
		jsonContent := `{"name": "test", "value": 42}`
		err := os.WriteFile(jsonFile, []byte(jsonContent), 0o600)
		ts.Require().NoError(err)

		var result map[string]interface{}
		err = loader.LoadFrom(jsonFile, &result)
		ts.Require().NoError(err)
		ts.Require().Equal("test", result["name"])
		ts.Require().InDelta(float64(42), result["value"], 0.001)
	})

	ts.Run("load YAML file", func() {
		yamlFile := filepath.Join(ts.tempDir, "config.yaml")
		err := os.WriteFile(yamlFile, []byte(testYAMLContent), 0o600)
		ts.Require().NoError(err)

		var result map[string]interface{}
		err = loader.LoadFrom(yamlFile, &result)
		ts.Require().NoError(err)
		ts.Require().Equal("test", result["name"])
		ts.Require().Equal(42, result["value"])
	})

	ts.Run("load YML file", func() {
		ymlFile := filepath.Join(ts.tempDir, "config.yml")
		err := os.WriteFile(ymlFile, []byte(testYAMLContent), 0o600)
		ts.Require().NoError(err)

		var result map[string]interface{}
		err = loader.LoadFrom(ymlFile, &result)
		ts.Require().NoError(err)
		ts.Require().Equal("test", result["name"])
		ts.Require().Equal(42, result["value"])
	})

	ts.Run("load file with no extension - tries YAML then JSON", func() {
		configFile := filepath.Join(ts.tempDir, "config")
		err := os.WriteFile(configFile, []byte(testYAMLContent), 0o600)
		ts.Require().NoError(err)

		var result map[string]interface{}
		err = loader.LoadFrom(configFile, &result)
		ts.Require().NoError(err)
		ts.Require().Equal("test", result["name"])
	})

	ts.Run("file does not exist", func() {
		var result map[string]interface{}
		err := loader.LoadFrom("nonexistent.json", &result)
		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "does not exist")
	})
}

// TestDefaultConfigLoader_Load tests loading from multiple paths with fallback
func (ts *DefaultConfigTestSuite) TestDefaultConfigLoader_Load() {
	loader := NewDefaultConfigLoader()

	ts.Run("load from first available path", func() {
		configFile1 := filepath.Join(ts.tempDir, "config.json")
		configContent1 := `{"name": "first", "value": 1}`
		err := os.WriteFile(configFile1, []byte(configContent1), 0o600)
		ts.Require().NoError(err)

		configFile2 := filepath.Join(ts.tempDir, "config2.json")
		configContent2 := `{"name": "second", "value": 2}`
		err = os.WriteFile(configFile2, []byte(configContent2), 0o600)
		ts.Require().NoError(err)

		paths := []string{
			filepath.Join(ts.tempDir, "nonexistent.json"),
			configFile1,
			configFile2,
		}

		var result map[string]interface{}
		loadedPath, err := loader.Load(paths, &result)
		ts.Require().NoError(err)
		ts.Require().Equal(configFile1, loadedPath)
		ts.Require().Equal("first", result["name"])
	})

	ts.Run("no valid configuration file found", func() {
		paths := []string{
			filepath.Join(ts.tempDir, "nonexistent1.json"),
			filepath.Join(ts.tempDir, "nonexistent2.json"),
		}

		var result map[string]interface{}
		_, err := loader.Load(paths, &result)
		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "no valid configuration file found")
	})
}

// TestDefaultConfigLoader_Save tests saving configuration to files
func (ts *DefaultConfigTestSuite) TestDefaultConfigLoader_Save() {
	loader := NewDefaultConfigLoader()
	data := map[string]interface{}{
		"name":  "test",
		"value": 42,
	}

	ts.Run("save as JSON", func() {
		jsonFile := filepath.Join(ts.tempDir, "output.json")
		err := loader.Save(jsonFile, data, "json")
		ts.Require().NoError(err)

		// Verify file exists and content is correct
		ts.Require().FileExists(jsonFile)
		// Validate path is within temp directory (security check)
		cleanPath := filepath.Clean(jsonFile)
		ts.Require().True(strings.HasPrefix(cleanPath, ts.tempDir), "path should be within temp directory")
		content, err := os.ReadFile(cleanPath)
		ts.Require().NoError(err)
		ts.Require().Contains(string(content), `"name": "test"`)
		ts.Require().Contains(string(content), `"value": 42`)
	})

	ts.Run("save as YAML", func() {
		yamlFile := filepath.Join(ts.tempDir, "output.yaml")
		err := loader.Save(yamlFile, data, "yaml")
		ts.Require().NoError(err)

		// Verify file exists
		ts.Require().FileExists(yamlFile)
		// Validate path is within temp directory (security check)
		cleanPath := filepath.Clean(yamlFile)
		ts.Require().True(strings.HasPrefix(cleanPath, ts.tempDir), "path should be within temp directory")
		content, err := os.ReadFile(cleanPath)
		ts.Require().NoError(err)
		ts.Require().Contains(string(content), "name: test")
		ts.Require().Contains(string(content), "value: 42")
	})

	ts.Run("save as YML", func() {
		ymlFile := filepath.Join(ts.tempDir, "output.yml")
		err := loader.Save(ymlFile, data, "yml")
		ts.Require().NoError(err)

		ts.Require().FileExists(ymlFile)
	})

	ts.Run("save in nested directory", func() {
		nestedFile := filepath.Join(ts.tempDir, "nested", "dir", "config.json")
		err := loader.Save(nestedFile, data, "json")
		ts.Require().NoError(err)

		ts.Require().FileExists(nestedFile)
	})

	ts.Run("unsupported format", func() {
		xmlFile := filepath.Join(ts.tempDir, "output.xml")
		err := loader.Save(xmlFile, data, "xml")
		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "unsupported format")
	})
}

// TestDefaultConfigLoader_Validate tests configuration validation
func (ts *DefaultConfigTestSuite) TestDefaultConfigLoader_Validate() {
	loader := NewDefaultConfigLoader()

	ts.Run("valid data", func() {
		data := map[string]interface{}{"key": "value"}
		err := loader.Validate(data)
		ts.Require().NoError(err)
	})

	ts.Run("nil data", func() {
		err := loader.Validate(nil)
		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "configuration data is nil")
	})
}

// TestDefaultConfigLoader_GetSupportedFormats tests getting supported formats
func (ts *DefaultConfigTestSuite) TestDefaultConfigLoader_GetSupportedFormats() {
	loader := NewDefaultConfigLoader()
	formats := loader.GetSupportedFormats()
	ts.Require().Contains(formats, "yaml")
	ts.Require().Contains(formats, "yml")
	ts.Require().Contains(formats, "json")
	ts.Require().Len(formats, 3)
}

// TestNewDefaultEnvProvider tests creating a new environment provider
func (ts *DefaultConfigTestSuite) TestNewDefaultEnvProvider() {
	provider := NewDefaultEnvProvider()
	ts.Require().NotNil(provider)
}

// TestDefaultEnvProvider_BasicOperations tests basic env operations
func (ts *DefaultConfigTestSuite) TestDefaultEnvProvider_BasicOperations() {
	provider := NewDefaultEnvProvider()
	testKey := "TEST_CONFIG_KEY"
	testValue := "test_value"

	// Clean up
	defer func() {
		if err := os.Unsetenv(testKey); err != nil {
			ts.T().Logf("Failed to unset %s: %v", testKey, err)
		}
	}()

	ts.Run("Set and Get", func() {
		err := provider.Set(testKey, testValue)
		ts.Require().NoError(err)

		value := provider.Get(testKey)
		ts.Require().Equal(testValue, value)
	})

	ts.Run("GetWithDefault - existing", func() {
		err := provider.Set(testKey, testValue)
		ts.Require().NoError(err)

		value := provider.GetWithDefault(testKey, "default")
		ts.Require().Equal(testValue, value)
	})

	ts.Run("GetWithDefault - missing", func() {
		if err := os.Unsetenv(testKey); err != nil {
			ts.T().Logf("Failed to unset %s: %v", testKey, err)
		}
		value := provider.GetWithDefault(testKey, "default")
		ts.Require().Equal("default", value)
	})

	ts.Run("LookupEnv - existing", func() {
		err := provider.Set(testKey, testValue)
		ts.Require().NoError(err)

		value, found := provider.LookupEnv(testKey)
		ts.Require().True(found)
		ts.Require().Equal(testValue, value)
	})

	ts.Run("LookupEnv - missing", func() {
		if err := os.Unsetenv(testKey); err != nil {
			ts.T().Logf("Failed to unset %s: %v", testKey, err)
		}
		value, found := provider.LookupEnv(testKey)
		ts.Require().False(found)
		ts.Require().Empty(value)
	})

	ts.Run("Unset", func() {
		err := provider.Set(testKey, testValue)
		ts.Require().NoError(err)

		err = provider.Unset(testKey)
		ts.Require().NoError(err)

		value := provider.Get(testKey)
		ts.Require().Empty(value)
	})
}

// TestDefaultEnvProvider_GetAll tests getting all environment variables
func (ts *DefaultConfigTestSuite) TestDefaultEnvProvider_GetAll() {
	provider := NewDefaultEnvProvider()
	testKey := "TEST_CONFIG_ALL_KEY"
	testValue := "test_all_value"

	// Clean up
	defer func() {
		if err := os.Unsetenv(testKey); err != nil {
			ts.T().Logf("Failed to unset %s: %v", testKey, err)
		}
	}()

	err := provider.Set(testKey, testValue)
	ts.Require().NoError(err)

	allVars := provider.GetAll()
	ts.Require().Contains(allVars, testKey)
	ts.Require().Equal(testValue, allVars[testKey])
}

// TestDefaultEnvProvider_GetBool tests getting boolean environment variables
func (ts *DefaultConfigTestSuite) TestDefaultEnvProvider_GetBool() {
	provider := NewDefaultEnvProvider()
	testKey := "TEST_CONFIG_BOOL"

	// Clean up
	defer func() {
		if err := os.Unsetenv(testKey); err != nil {
			ts.T().Logf("Failed to unset %s: %v", testKey, err)
		}
	}()

	testCases := []struct {
		name         string
		value        string
		defaultValue bool
		expected     bool
	}{
		{"true string", "true", false, true},
		{"True string", "True", false, true},
		{"1 string", "1", false, true},
		{"yes string", "yes", false, true},
		{"on string", "on", false, true},
		{"enabled string", "enabled", false, true},
		{"false string", "false", true, false},
		{"0 string", "0", true, false},
		{"no string", "no", true, false},
		{"off string", "off", true, false},
		{"disabled string", "disabled", true, false},
		{"invalid string", "invalid", true, true},
		{"empty string", "", false, false},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			if tc.value != "" {
				if err := os.Setenv(testKey, tc.value); err != nil {
					ts.T().Fatalf("Failed to set %s: %v", testKey, err)
				}
			} else {
				if err := os.Unsetenv(testKey); err != nil {
					ts.T().Logf("Failed to unset %s: %v", testKey, err)
				}
			}

			result := provider.GetBool(testKey, tc.defaultValue)
			ts.Require().Equal(tc.expected, result)
		})
	}
}

// TestDefaultEnvProvider_GetInt tests getting integer environment variables
func (ts *DefaultConfigTestSuite) TestDefaultEnvProvider_GetInt() {
	provider := NewDefaultEnvProvider()
	testKey := "TEST_CONFIG_INT"

	// Clean up
	defer func() {
		if err := os.Unsetenv(testKey); err != nil {
			ts.T().Logf("Failed to unset %s: %v", testKey, err)
		}
	}()

	testCases := []struct {
		name         string
		value        string
		defaultValue int
		expected     int
	}{
		{"valid positive int", "42", 0, 42},
		{"valid negative int", "-10", 0, -10},
		{"valid zero", "0", 5, 0},
		{"invalid int", "not_a_number", 100, 100},
		{"empty string", "", 50, 50},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			if tc.value != "" {
				if err := os.Setenv(testKey, tc.value); err != nil {
					ts.T().Fatalf("Failed to set %s: %v", testKey, err)
				}
			} else {
				if err := os.Unsetenv(testKey); err != nil {
					ts.T().Logf("Failed to unset %s: %v", testKey, err)
				}
			}

			result := provider.GetInt(testKey, tc.defaultValue)
			ts.Require().Equal(tc.expected, result)
		})
	}
}

// TestDefaultEnvProvider_GetInt64 tests getting int64 environment variables
func (ts *DefaultConfigTestSuite) TestDefaultEnvProvider_GetInt64() {
	provider := NewDefaultEnvProvider()
	testKey := "TEST_CONFIG_INT64"

	// Clean up
	defer func() {
		if err := os.Unsetenv(testKey); err != nil {
			ts.T().Logf("Failed to unset %s: %v", testKey, err)
		}
	}()

	testCases := []struct {
		name         string
		value        string
		defaultValue int64
		expected     int64
	}{
		{"valid int64", "9223372036854775807", 0, 9223372036854775807},
		{"valid negative int64", "-9223372036854775808", 0, -9223372036854775808},
		{"invalid int64", "not_a_number", 100, 100},
		{"empty string", "", 50, 50},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			if tc.value != "" {
				if err := os.Setenv(testKey, tc.value); err != nil {
					ts.T().Fatalf("Failed to set %s: %v", testKey, err)
				}
			} else {
				if err := os.Unsetenv(testKey); err != nil {
					ts.T().Logf("Failed to unset %s: %v", testKey, err)
				}
			}

			result := provider.GetInt64(testKey, tc.defaultValue)
			ts.Require().Equal(tc.expected, result)
		})
	}
}

// TestDefaultEnvProvider_GetFloat64 tests getting float64 environment variables
func (ts *DefaultConfigTestSuite) TestDefaultEnvProvider_GetFloat64() {
	provider := NewDefaultEnvProvider()
	testKey := "TEST_CONFIG_FLOAT64"

	// Clean up
	defer func() {
		if err := os.Unsetenv(testKey); err != nil {
			ts.T().Logf("Failed to unset %s: %v", testKey, err)
		}
	}()

	testCases := []struct {
		name         string
		value        string
		defaultValue float64
		expected     float64
	}{
		{"valid float", "3.14159", 0.0, 3.14159},
		{"valid negative float", "-2.71828", 0.0, -2.71828},
		{"valid integer as float", "42", 0.0, 42.0},
		{"invalid float", "not_a_number", 1.0, 1.0},
		{"empty string", "", 2.5, 2.5},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			if tc.value != "" {
				if err := os.Setenv(testKey, tc.value); err != nil {
					ts.T().Fatalf("Failed to set %s: %v", testKey, err)
				}
			} else {
				if err := os.Unsetenv(testKey); err != nil {
					ts.T().Logf("Failed to unset %s: %v", testKey, err)
				}
			}

			result := provider.GetFloat64(testKey, tc.defaultValue)
			ts.Require().InDelta(tc.expected, result, 0.00001)
		})
	}
}

// TestDefaultEnvProvider_GetDuration tests getting duration environment variables
func (ts *DefaultConfigTestSuite) TestDefaultEnvProvider_GetDuration() {
	provider := NewDefaultEnvProvider()
	testKey := "TEST_CONFIG_DURATION"

	// Clean up
	defer func() {
		if err := os.Unsetenv(testKey); err != nil {
			ts.T().Logf("Failed to unset %s: %v", testKey, err)
		}
	}()

	testCases := []struct {
		name         string
		value        string
		defaultValue time.Duration
		expected     time.Duration
	}{
		{"valid duration seconds", "30s", 0, 30 * time.Second},
		{"valid duration minutes", "5m", 0, 5 * time.Minute},
		{"valid duration hours", "2h", 0, 2 * time.Hour},
		{"valid duration mixed", "1h30m", 0, time.Hour + 30*time.Minute},
		{"invalid duration", "not_a_duration", time.Minute, time.Minute},
		{"empty string", "", time.Hour, time.Hour},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			if tc.value != "" {
				if err := os.Setenv(testKey, tc.value); err != nil {
					ts.T().Fatalf("Failed to set %s: %v", testKey, err)
				}
			} else {
				if err := os.Unsetenv(testKey); err != nil {
					ts.T().Logf("Failed to unset %s: %v", testKey, err)
				}
			}

			result := provider.GetDuration(testKey, tc.defaultValue)
			ts.Require().Equal(tc.expected, result)
		})
	}
}

// TestDefaultEnvProvider_GetStringSlice tests getting string slice environment variables
func (ts *DefaultConfigTestSuite) TestDefaultEnvProvider_GetStringSlice() {
	provider := NewDefaultEnvProvider()
	testKey := "TEST_CONFIG_STRING_SLICE"

	// Clean up
	defer func() {
		if err := os.Unsetenv(testKey); err != nil {
			ts.T().Logf("Failed to unset %s: %v", testKey, err)
		}
	}()

	testCases := []struct {
		name         string
		value        string
		defaultValue []string
		expected     []string
	}{
		{"single value", "item1", []string{"default"}, []string{"item1"}},
		{"multiple values", "item1,item2,item3", []string{"default"}, []string{"item1", "item2", "item3"}},
		{"values with spaces", "item1, item2 , item3", []string{"default"}, []string{"item1", "item2", "item3"}},
		{"empty items filtered", "item1,,item2,", []string{"default"}, []string{"item1", "item2"}},
		{"all empty items", ",,", []string{"default"}, []string{"default"}},
		{"empty string", "", []string{"default"}, []string{"default"}},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			if tc.value != "" {
				if err := os.Setenv(testKey, tc.value); err != nil {
					ts.T().Fatalf("Failed to set %s: %v", testKey, err)
				}
			} else {
				if err := os.Unsetenv(testKey); err != nil {
					ts.T().Logf("Failed to unset %s: %v", testKey, err)
				}
			}

			result := provider.GetStringSlice(testKey, tc.defaultValue)
			ts.Require().Equal(tc.expected, result)
		})
	}
}

// TestNewFileConfigSource tests creating a new file config source
func (ts *DefaultConfigTestSuite) TestNewFileConfigSource() {
	path := filepath.Join(ts.tempDir, "config.json")
	source := NewFileConfigSource(path, FormatJSON, 100)

	ts.Require().NotNil(source)
	ts.Require().Equal(path, source.path)
	ts.Require().Equal(FormatJSON, source.format)
	ts.Require().Equal(100, source.priority)
	ts.Require().NotNil(source.loader)
}

// TestFileConfigSource_Methods tests file config source methods
func (ts *DefaultConfigTestSuite) TestFileConfigSource_Methods() {
	path := filepath.Join(ts.tempDir, "config.json")
	source := NewFileConfigSource(path, FormatJSON, 100)

	ts.Run("Name", func() {
		name := source.Name()
		ts.Require().Equal("file:"+path, name)
	})

	ts.Run("Priority", func() {
		priority := source.Priority()
		ts.Require().Equal(100, priority)
	})

	ts.Run("IsAvailable - file does not exist", func() {
		available := source.IsAvailable()
		ts.Require().False(available)
	})

	ts.Run("IsAvailable - file exists", func() {
		// Create the file
		err := os.WriteFile(path, []byte(`{"test": true}`), 0o600)
		ts.Require().NoError(err)

		available := source.IsAvailable()
		ts.Require().True(available)
	})

	ts.Run("Load - successful", func() {
		// Create a valid JSON file
		jsonContent := `{"name": "test", "value": 42}`
		err := os.WriteFile(path, []byte(jsonContent), 0o600)
		ts.Require().NoError(err)

		var result map[string]interface{}
		err = source.Load(&result)
		ts.Require().NoError(err)
		ts.Require().Equal("test", result["name"])
		ts.Require().InDelta(float64(42), result["value"], 0.001)
	})
}

// TestNewEnvConfigSource tests creating a new environment config source
func (ts *DefaultConfigTestSuite) TestNewEnvConfigSource() {
	source := NewEnvConfigSource("APP_", 200)

	ts.Require().NotNil(source)
	ts.Require().Equal("APP_", source.prefix)
	ts.Require().Equal(200, source.priority)
	ts.Require().NotNil(source.envOps)
}

// TestEnvConfigSource_Methods tests environment config source methods
func (ts *DefaultConfigTestSuite) TestEnvConfigSource_Methods() {
	source := NewEnvConfigSource("APP_", 200)

	ts.Run("Name", func() {
		name := source.Name()
		ts.Require().Equal("env:APP_", name)
	})

	ts.Run("Priority", func() {
		priority := source.Priority()
		ts.Require().Equal(200, priority)
	})

	ts.Run("IsAvailable", func() {
		available := source.IsAvailable()
		ts.Require().True(available)
	})

	ts.Run("Load - not implemented", func() {
		var result map[string]interface{}
		err := source.Load(&result)
		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "not implemented yet")
	})
}

// TestDefaultConfigTestSuite runs the test suite
func TestDefaultConfigTestSuite(t *testing.T) {
	suite.Run(t, new(DefaultConfigTestSuite))
}

// TestDefaultEnvProvider_GetInt_EdgeCases tests integer parsing edge cases
// including overflow, boundary values, and unusual input formats.
func TestDefaultEnvProvider_GetInt_EdgeCases(t *testing.T) {
	provider := NewDefaultEnvProvider()
	testKey := "TEST_INT_EDGE"

	t.Cleanup(func() {
		if err := os.Unsetenv(testKey); err != nil {
			t.Logf("Failed to unset %s: %v", testKey, err)
		}
	})

	tests := []struct {
		name     string
		value    string
		defVal   int
		expected int
	}{
		// Overflow cases - should return default
		{
			name:     "overflow large positive",
			value:    "99999999999999999999",
			defVal:   42,
			expected: 42,
		},
		{
			name:     "overflow large negative",
			value:    "-99999999999999999999",
			defVal:   42,
			expected: 42,
		},

		// Boundary values - should parse correctly
		{
			name:     "max int32",
			value:    "2147483647",
			defVal:   0,
			expected: 2147483647,
		},
		{
			name:     "min int32",
			value:    "-2147483648",
			defVal:   0,
			expected: -2147483648,
		},

		// Invalid formats - should return default
		{
			name:     "float string",
			value:    "3.14",
			defVal:   99,
			expected: 99,
		},
		{
			name:     "scientific notation",
			value:    "1e10",
			defVal:   77,
			expected: 77,
		},
		{
			name:     "hex notation",
			value:    "0xFF",
			defVal:   88,
			expected: 88,
		},
		{
			name:     "whitespace around number",
			value:    "  42  ",
			defVal:   0,
			expected: 0, // strconv.Atoi doesn't trim whitespace
		},
		{
			name:     "number with suffix",
			value:    "42abc",
			defVal:   0,
			expected: 0,
		},
		{
			name:     "plus sign prefix",
			value:    "+42",
			defVal:   0,
			expected: 42, // strconv.Atoi handles +
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := os.Setenv(testKey, tt.value); err != nil {
				t.Fatalf("Failed to set env: %v", err)
			}

			result := provider.GetInt(testKey, tt.defVal)
			if result != tt.expected {
				t.Errorf("GetInt(%q, %d) = %d, want %d", tt.value, tt.defVal, result, tt.expected)
			}
		})
	}
}

// TestDefaultEnvProvider_GetFloat64_EdgeCases tests float64 parsing edge cases
// including special values like infinity and NaN.
func TestDefaultEnvProvider_GetFloat64_EdgeCases(t *testing.T) {
	provider := NewDefaultEnvProvider()
	testKey := "TEST_FLOAT_EDGE"

	t.Cleanup(func() {
		if err := os.Unsetenv(testKey); err != nil {
			t.Logf("Failed to unset %s: %v", testKey, err)
		}
	})

	t.Run("positive infinity", func(t *testing.T) {
		require.NoError(t, os.Setenv(testKey, "Inf"))
		result := provider.GetFloat64(testKey, 0)
		if result != math.Inf(1) {
			t.Errorf("expected +Inf, got %v", result)
		}
	})

	t.Run("positive infinity with plus", func(t *testing.T) {
		require.NoError(t, os.Setenv(testKey, "+Inf"))
		result := provider.GetFloat64(testKey, 0)
		if result != math.Inf(1) {
			t.Errorf("expected +Inf, got %v", result)
		}
	})

	t.Run("negative infinity", func(t *testing.T) {
		require.NoError(t, os.Setenv(testKey, "-Inf"))
		result := provider.GetFloat64(testKey, 0)
		if result != math.Inf(-1) {
			t.Errorf("expected -Inf, got %v", result)
		}
	})

	t.Run("NaN returns default", func(t *testing.T) {
		require.NoError(t, os.Setenv(testKey, "NaN"))
		result := provider.GetFloat64(testKey, 42.0)
		// NaN parsing actually works in Go, but NaN != NaN
		// So we check if it's NaN
		if !math.IsNaN(result) {
			// If parsing succeeded, we should get NaN
			// If it returned default, that's also acceptable behavior
			if result != 42.0 {
				t.Errorf("expected NaN or default 42.0, got %v", result)
			}
		}
	})

	t.Run("scientific notation positive", func(t *testing.T) {
		require.NoError(t, os.Setenv(testKey, "1e10"))
		result := provider.GetFloat64(testKey, 0)
		if result != 1e10 {
			t.Errorf("expected 1e10, got %v", result)
		}
	})

	t.Run("scientific notation negative exponent", func(t *testing.T) {
		require.NoError(t, os.Setenv(testKey, "1e-5"))
		result := provider.GetFloat64(testKey, 0)
		if math.Abs(result-1e-5) > 1e-15 {
			t.Errorf("expected 1e-5, got %v", result)
		}
	})

	t.Run("very small number", func(t *testing.T) {
		require.NoError(t, os.Setenv(testKey, "0.000000001"))
		result := provider.GetFloat64(testKey, 0)
		if math.Abs(result-0.000000001) > 1e-15 {
			t.Errorf("expected 0.000000001, got %v", result)
		}
	})

	t.Run("max float64", func(t *testing.T) {
		require.NoError(t, os.Setenv(testKey, "1.7976931348623157e+308"))
		result := provider.GetFloat64(testKey, 0)
		if result != math.MaxFloat64 {
			t.Errorf("expected MaxFloat64, got %v", result)
		}
	})
}

// TestDefaultConfigLoader_LoadFrom_EdgeCases tests file loading edge cases
func TestDefaultConfigLoader_LoadFrom_EdgeCases(t *testing.T) {
	loader := NewDefaultConfigLoader()
	tmpDir := t.TempDir()

	t.Run("empty YAML file", func(t *testing.T) {
		emptyFile := filepath.Join(tmpDir, "empty.yaml")
		if err := os.WriteFile(emptyFile, []byte(""), 0o600); err != nil {
			t.Fatal(err)
		}

		var result map[string]interface{}
		err := loader.LoadFrom(emptyFile, &result)
		// Empty file should parse successfully to nil/empty map
		if err != nil {
			t.Errorf("empty YAML should parse, got error: %v", err)
		}
	})

	t.Run("YAML with only comments", func(t *testing.T) {
		commentFile := filepath.Join(tmpDir, "comments.yaml")
		content := `# This is a comment
# Another comment
# No actual data`
		if err := os.WriteFile(commentFile, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}

		var result map[string]interface{}
		err := loader.LoadFrom(commentFile, &result)
		// Comments-only file should parse successfully
		if err != nil {
			t.Errorf("comments-only YAML should parse, got error: %v", err)
		}
	})

	t.Run("empty JSON file", func(t *testing.T) {
		emptyFile := filepath.Join(tmpDir, "empty.json")
		if err := os.WriteFile(emptyFile, []byte(""), 0o600); err != nil {
			t.Fatal(err)
		}

		var result map[string]interface{}
		err := loader.LoadFrom(emptyFile, &result)
		// Empty JSON is invalid
		if err == nil {
			t.Error("empty JSON should fail to parse")
		}
	})

	t.Run("empty JSON object", func(t *testing.T) {
		emptyObjFile := filepath.Join(tmpDir, "empty_obj.json")
		if err := os.WriteFile(emptyObjFile, []byte("{}"), 0o600); err != nil {
			t.Fatal(err)
		}

		var result map[string]interface{}
		err := loader.LoadFrom(emptyObjFile, &result)
		if err != nil {
			t.Errorf("empty JSON object should parse, got error: %v", err)
		}
		if result == nil {
			t.Error("result should be initialized empty map, not nil")
		}
	})

	t.Run("unknown extension with JSON content auto-detects", func(t *testing.T) {
		unknownFile := filepath.Join(tmpDir, "config.conf")
		content := `{"name": "test", "value": 42}`
		if err := os.WriteFile(unknownFile, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}

		var result map[string]interface{}
		err := loader.LoadFrom(unknownFile, &result)
		if err != nil {
			t.Errorf("should auto-detect JSON, got error: %v", err)
		}
		if result["name"] != "test" {
			t.Errorf("expected name=test, got %v", result["name"])
		}
	})

	t.Run("unknown extension with YAML content auto-detects", func(t *testing.T) {
		unknownFile := filepath.Join(tmpDir, "config.cfg")
		content := `name: test
value: 42`
		if err := os.WriteFile(unknownFile, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}

		var result map[string]interface{}
		err := loader.LoadFrom(unknownFile, &result)
		if err != nil {
			t.Errorf("should auto-detect YAML, got error: %v", err)
		}
		if result["name"] != "test" {
			t.Errorf("expected name=test, got %v", result["name"])
		}
	})

	t.Run("unknown extension with invalid content fails", func(t *testing.T) {
		unknownFile := filepath.Join(tmpDir, "config.xxx")
		content := `not valid json or yaml: {{{`
		if err := os.WriteFile(unknownFile, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}

		var result map[string]interface{}
		err := loader.LoadFrom(unknownFile, &result)
		if err == nil {
			t.Error("invalid content should fail to parse")
		}
	})
}

// TestDefaultConfigLoader_Save_Permissions tests file permission handling
func TestDefaultConfigLoader_Save_Permissions(t *testing.T) {
	loader := NewDefaultConfigLoader()
	tmpDir := t.TempDir()

	t.Run("creates nested directories", func(t *testing.T) {
		nestedPath := filepath.Join(tmpDir, "a", "b", "c", "config.json")
		data := map[string]interface{}{"test": true}

		err := loader.Save(nestedPath, data, "json")
		if err != nil {
			t.Fatalf("failed to save to nested path: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(nestedPath); os.IsNotExist(err) {
			t.Error("file should exist after save")
		}
	})

	t.Run("file has correct permissions", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "perms.json")
		data := map[string]interface{}{"test": true}

		err := loader.Save(configPath, data, "json")
		if err != nil {
			t.Fatalf("failed to save: %v", err)
		}

		info, err := os.Stat(configPath)
		if err != nil {
			t.Fatal(err)
		}

		// File should be 0600 (owner read/write only)
		perm := info.Mode().Perm()
		if perm != 0o600 {
			t.Errorf("expected permissions 0600, got %04o", perm)
		}
	})

	t.Run("directory has correct permissions", func(t *testing.T) {
		nestedPath := filepath.Join(tmpDir, "dir_perms", "config.json")
		data := map[string]interface{}{"test": true}

		err := loader.Save(nestedPath, data, "json")
		if err != nil {
			t.Fatalf("failed to save: %v", err)
		}

		dirInfo, err := os.Stat(filepath.Dir(nestedPath))
		if err != nil {
			t.Fatal(err)
		}

		// Directory should be 0755 (based on code: os.MkdirAll with 0o755)
		perm := dirInfo.Mode().Perm()
		if perm != 0o755 {
			t.Errorf("expected directory permissions 0755, got %04o", perm)
		}
	})
}

// TestEnvConfigSource_Load_Error verifies the error path for EnvConfigSource.Load
func TestEnvConfigSource_Load_Error(t *testing.T) {
	source := NewEnvConfigSource("TEST_PREFIX_", 100)

	var config map[string]interface{}
	err := source.Load(&config)

	if err == nil {
		t.Fatal("EnvConfigSource.Load should return error")
	}

	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("error should mention 'not implemented', got: %v", err)
	}
}

// TestFileConfigSource_Load_NonExistent tests loading from non-existent file
func TestFileConfigSource_Load_NonExistent(t *testing.T) {
	source := NewFileConfigSource("/non/existent/path/config.yaml", FormatYAML, 100)

	var config map[string]interface{}
	err := source.Load(&config)

	if err == nil {
		t.Error("loading non-existent file should fail")
	}
}
