package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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
	require.NotNil(ts.T(), loader)
	require.NotNil(ts.T(), loader.fileOps)
	require.NotNil(ts.T(), loader.jsonOps)
	require.NotNil(ts.T(), loader.yamlOps)
}

// TestDefaultConfigLoader_LoadFrom tests loading from specific files
func (ts *DefaultConfigTestSuite) TestDefaultConfigLoader_LoadFrom() {
	loader := NewDefaultConfigLoader()

	ts.Run("load JSON file", func() {
		jsonFile := filepath.Join(ts.tempDir, "config.json")
		jsonContent := `{"name": "test", "value": 42}`
		err := os.WriteFile(jsonFile, []byte(jsonContent), 0o644)
		require.NoError(ts.T(), err)

		var result map[string]interface{}
		err = loader.LoadFrom(jsonFile, &result)
		require.NoError(ts.T(), err)
		require.Equal(ts.T(), "test", result["name"])
		require.Equal(ts.T(), float64(42), result["value"])
	})

	ts.Run("load YAML file", func() {
		yamlFile := filepath.Join(ts.tempDir, "config.yaml")
		yamlContent := `name: test
value: 42`
		err := os.WriteFile(yamlFile, []byte(yamlContent), 0o644)
		require.NoError(ts.T(), err)

		var result map[string]interface{}
		err = loader.LoadFrom(yamlFile, &result)
		require.NoError(ts.T(), err)
		require.Equal(ts.T(), "test", result["name"])
		require.Equal(ts.T(), 42, result["value"])
	})

	ts.Run("load YML file", func() {
		ymlFile := filepath.Join(ts.tempDir, "config.yml")
		ymlContent := `name: test
value: 42`
		err := os.WriteFile(ymlFile, []byte(ymlContent), 0o644)
		require.NoError(ts.T(), err)

		var result map[string]interface{}
		err = loader.LoadFrom(ymlFile, &result)
		require.NoError(ts.T(), err)
		require.Equal(ts.T(), "test", result["name"])
		require.Equal(ts.T(), 42, result["value"])
	})

	ts.Run("load file with no extension - tries YAML then JSON", func() {
		configFile := filepath.Join(ts.tempDir, "config")
		yamlContent := `name: test
value: 42`
		err := os.WriteFile(configFile, []byte(yamlContent), 0o644)
		require.NoError(ts.T(), err)

		var result map[string]interface{}
		err = loader.LoadFrom(configFile, &result)
		require.NoError(ts.T(), err)
		require.Equal(ts.T(), "test", result["name"])
	})

	ts.Run("file does not exist", func() {
		var result map[string]interface{}
		err := loader.LoadFrom("nonexistent.json", &result)
		require.Error(ts.T(), err)
		require.Contains(ts.T(), err.Error(), "does not exist")
	})
}

// TestDefaultConfigLoader_Load tests loading from multiple paths with fallback
func (ts *DefaultConfigTestSuite) TestDefaultConfigLoader_Load() {
	loader := NewDefaultConfigLoader()

	ts.Run("load from first available path", func() {
		jsonFile := filepath.Join(ts.tempDir, "config.json")
		jsonContent := `{"name": "first", "value": 1}`
		err := os.WriteFile(jsonFile, []byte(jsonContent), 0o644)
		require.NoError(ts.T(), err)

		jsonFile2 := filepath.Join(ts.tempDir, "config2.json")
		jsonContent2 := `{"name": "second", "value": 2}`
		err = os.WriteFile(jsonFile2, []byte(jsonContent2), 0o644)
		require.NoError(ts.T(), err)

		paths := []string{
			filepath.Join(ts.tempDir, "nonexistent.json"),
			jsonFile,
			jsonFile2,
		}

		var result map[string]interface{}
		loadedPath, err := loader.Load(paths, &result)
		require.NoError(ts.T(), err)
		require.Equal(ts.T(), jsonFile, loadedPath)
		require.Equal(ts.T(), "first", result["name"])
	})

	ts.Run("no valid configuration file found", func() {
		paths := []string{
			filepath.Join(ts.tempDir, "nonexistent1.json"),
			filepath.Join(ts.tempDir, "nonexistent2.json"),
		}

		var result map[string]interface{}
		_, err := loader.Load(paths, &result)
		require.Error(ts.T(), err)
		require.Contains(ts.T(), err.Error(), "no valid configuration file found")
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
		require.NoError(ts.T(), err)

		// Verify file exists and content is correct
		require.FileExists(ts.T(), jsonFile)
		content, err := os.ReadFile(jsonFile)
		require.NoError(ts.T(), err)
		require.Contains(ts.T(), string(content), `"name": "test"`)
		require.Contains(ts.T(), string(content), `"value": 42`)
	})

	ts.Run("save as YAML", func() {
		yamlFile := filepath.Join(ts.tempDir, "output.yaml")
		err := loader.Save(yamlFile, data, "yaml")
		require.NoError(ts.T(), err)

		// Verify file exists
		require.FileExists(ts.T(), yamlFile)
		content, err := os.ReadFile(yamlFile)
		require.NoError(ts.T(), err)
		require.Contains(ts.T(), string(content), "name: test")
		require.Contains(ts.T(), string(content), "value: 42")
	})

	ts.Run("save as YML", func() {
		ymlFile := filepath.Join(ts.tempDir, "output.yml")
		err := loader.Save(ymlFile, data, "yml")
		require.NoError(ts.T(), err)

		require.FileExists(ts.T(), ymlFile)
	})

	ts.Run("save in nested directory", func() {
		nestedFile := filepath.Join(ts.tempDir, "nested", "dir", "config.json")
		err := loader.Save(nestedFile, data, "json")
		require.NoError(ts.T(), err)

		require.FileExists(ts.T(), nestedFile)
	})

	ts.Run("unsupported format", func() {
		xmlFile := filepath.Join(ts.tempDir, "output.xml")
		err := loader.Save(xmlFile, data, "xml")
		require.Error(ts.T(), err)
		require.Contains(ts.T(), err.Error(), "unsupported format")
	})
}

// TestDefaultConfigLoader_Validate tests configuration validation
func (ts *DefaultConfigTestSuite) TestDefaultConfigLoader_Validate() {
	loader := NewDefaultConfigLoader()

	ts.Run("valid data", func() {
		data := map[string]interface{}{"key": "value"}
		err := loader.Validate(data)
		require.NoError(ts.T(), err)
	})

	ts.Run("nil data", func() {
		err := loader.Validate(nil)
		require.Error(ts.T(), err)
		require.Contains(ts.T(), err.Error(), "configuration data is nil")
	})
}

// TestDefaultConfigLoader_GetSupportedFormats tests getting supported formats
func (ts *DefaultConfigTestSuite) TestDefaultConfigLoader_GetSupportedFormats() {
	loader := NewDefaultConfigLoader()
	formats := loader.GetSupportedFormats()
	require.Contains(ts.T(), formats, "yaml")
	require.Contains(ts.T(), formats, "yml")
	require.Contains(ts.T(), formats, "json")
	require.Len(ts.T(), formats, 3)
}

// TestNewDefaultEnvProvider tests creating a new environment provider
func (ts *DefaultConfigTestSuite) TestNewDefaultEnvProvider() {
	provider := NewDefaultEnvProvider()
	require.NotNil(ts.T(), provider)
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
		require.NoError(ts.T(), err)

		value := provider.Get(testKey)
		require.Equal(ts.T(), testValue, value)
	})

	ts.Run("GetWithDefault - existing", func() {
		err := provider.Set(testKey, testValue)
		require.NoError(ts.T(), err)

		value := provider.GetWithDefault(testKey, "default")
		require.Equal(ts.T(), testValue, value)
	})

	ts.Run("GetWithDefault - missing", func() {
		if err := os.Unsetenv(testKey); err != nil {
			ts.T().Logf("Failed to unset %s: %v", testKey, err)
		}
		value := provider.GetWithDefault(testKey, "default")
		require.Equal(ts.T(), "default", value)
	})

	ts.Run("LookupEnv - existing", func() {
		err := provider.Set(testKey, testValue)
		require.NoError(ts.T(), err)

		value, found := provider.LookupEnv(testKey)
		require.True(ts.T(), found)
		require.Equal(ts.T(), testValue, value)
	})

	ts.Run("LookupEnv - missing", func() {
		if err := os.Unsetenv(testKey); err != nil {
			ts.T().Logf("Failed to unset %s: %v", testKey, err)
		}
		value, found := provider.LookupEnv(testKey)
		require.False(ts.T(), found)
		require.Empty(ts.T(), value)
	})

	ts.Run("Unset", func() {
		err := provider.Set(testKey, testValue)
		require.NoError(ts.T(), err)

		err = provider.Unset(testKey)
		require.NoError(ts.T(), err)

		value := provider.Get(testKey)
		require.Empty(ts.T(), value)
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
	require.NoError(ts.T(), err)

	allVars := provider.GetAll()
	require.Contains(ts.T(), allVars, testKey)
	require.Equal(ts.T(), testValue, allVars[testKey])
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
			require.Equal(ts.T(), tc.expected, result)
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
			require.Equal(ts.T(), tc.expected, result)
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
			require.Equal(ts.T(), tc.expected, result)
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
			require.InDelta(ts.T(), tc.expected, result, 0.00001)
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
			require.Equal(ts.T(), tc.expected, result)
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
			require.Equal(ts.T(), tc.expected, result)
		})
	}
}

// TestNewFileConfigSource tests creating a new file config source
func (ts *DefaultConfigTestSuite) TestNewFileConfigSource() {
	path := filepath.Join(ts.tempDir, "config.json")
	source := NewFileConfigSource(path, FormatJSON, 100)

	require.NotNil(ts.T(), source)
	require.Equal(ts.T(), path, source.path)
	require.Equal(ts.T(), FormatJSON, source.format)
	require.Equal(ts.T(), 100, source.priority)
	require.NotNil(ts.T(), source.loader)
}

// TestFileConfigSource_Methods tests file config source methods
func (ts *DefaultConfigTestSuite) TestFileConfigSource_Methods() {
	path := filepath.Join(ts.tempDir, "config.json")
	source := NewFileConfigSource(path, FormatJSON, 100)

	ts.Run("Name", func() {
		name := source.Name()
		require.Equal(ts.T(), "file:"+path, name)
	})

	ts.Run("Priority", func() {
		priority := source.Priority()
		require.Equal(ts.T(), 100, priority)
	})

	ts.Run("IsAvailable - file does not exist", func() {
		available := source.IsAvailable()
		require.False(ts.T(), available)
	})

	ts.Run("IsAvailable - file exists", func() {
		// Create the file
		err := os.WriteFile(path, []byte(`{"test": true}`), 0o644)
		require.NoError(ts.T(), err)

		available := source.IsAvailable()
		require.True(ts.T(), available)
	})

	ts.Run("Load - successful", func() {
		// Create a valid JSON file
		jsonContent := `{"name": "test", "value": 42}`
		err := os.WriteFile(path, []byte(jsonContent), 0o644)
		require.NoError(ts.T(), err)

		var result map[string]interface{}
		err = source.Load(&result)
		require.NoError(ts.T(), err)
		require.Equal(ts.T(), "test", result["name"])
		require.Equal(ts.T(), float64(42), result["value"])
	})
}

// TestNewEnvConfigSource tests creating a new environment config source
func (ts *DefaultConfigTestSuite) TestNewEnvConfigSource() {
	source := NewEnvConfigSource("APP_", 200)

	require.NotNil(ts.T(), source)
	require.Equal(ts.T(), "APP_", source.prefix)
	require.Equal(ts.T(), 200, source.priority)
	require.NotNil(ts.T(), source.envOps)
}

// TestEnvConfigSource_Methods tests environment config source methods
func (ts *DefaultConfigTestSuite) TestEnvConfigSource_Methods() {
	source := NewEnvConfigSource("APP_", 200)

	ts.Run("Name", func() {
		name := source.Name()
		require.Equal(ts.T(), "env:APP_", name)
	})

	ts.Run("Priority", func() {
		priority := source.Priority()
		require.Equal(ts.T(), 200, priority)
	})

	ts.Run("IsAvailable", func() {
		available := source.IsAvailable()
		require.True(ts.T(), available)
	})

	ts.Run("Load - not implemented", func() {
		var result map[string]interface{}
		err := source.Load(&result)
		require.Error(ts.T(), err)
		require.Contains(ts.T(), err.Error(), "not implemented yet")
	})
}

// TestDefaultConfigTestSuite runs the test suite
func TestDefaultConfigTestSuite(t *testing.T) {
	suite.Run(t, new(DefaultConfigTestSuite))
}
