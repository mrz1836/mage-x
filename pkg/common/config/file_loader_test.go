package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Test constants for reuse
const (
	fileLoaderTestJSON      = `{"name": "test", "value": 42}`
	fileLoaderTestYAML      = "name: test\nvalue: 42"
	fileLoaderMalformedJSON = `{"name": "test", value: 42}` // Missing quotes around value key
	fileLoaderMalformedYAML = "name: test\nvalue: [invalid" // Unclosed bracket
)

// testConfig is a struct for unmarshaling test config files
type testConfig struct {
	Name  string `json:"name" yaml:"name"`
	Value int    `json:"value" yaml:"value"`
}

// FileLoaderTestSuite defines the test suite for FileConfigLoader
type FileLoaderTestSuite struct {
	suite.Suite

	tempDir string
	loader  *FileConfigLoader
}

// SetupTest runs before each test
func (ts *FileLoaderTestSuite) SetupTest() {
	ts.tempDir = ts.T().TempDir()
	ts.loader = &FileConfigLoader{}
}

// createTestFile creates a test file with the given content and returns its absolute path
func (ts *FileLoaderTestSuite) createTestFile(name, content string) string {
	path := filepath.Join(ts.tempDir, name)
	err := os.WriteFile(path, []byte(content), 0o600)
	ts.Require().NoError(err, "failed to create test file")
	return path
}

// TestNewFileLoader tests creating a new file loader
func (ts *FileLoaderTestSuite) TestNewFileLoader() {
	ts.Run("with default path", func() {
		loader := NewFileLoader("/default/path")
		ts.Require().NotNil(loader)

		fileLoader, ok := loader.(*FileConfigLoader)
		ts.Require().True(ok)
		ts.Require().Equal("/default/path", fileLoader.defaultPath)
	})

	ts.Run("with empty default path", func() {
		loader := NewFileLoader("")
		ts.Require().NotNil(loader)

		fileLoader, ok := loader.(*FileConfigLoader)
		ts.Require().True(ok)
		ts.Require().Empty(fileLoader.defaultPath)
	})
}

// TestFileConfigLoader_Load tests the Load method with multiple paths
func (ts *FileLoaderTestSuite) TestFileConfigLoader_Load() {
	ts.Run("loads from first valid path", func() {
		jsonFile := ts.createTestFile("config.json", fileLoaderTestJSON)

		var result testConfig
		loadedPath, err := ts.loader.Load([]string{jsonFile}, &result)

		ts.Require().NoError(err)
		ts.Require().Equal(jsonFile, loadedPath) //nolint:testifylint // comparing file paths, not JSON content
		ts.Require().Equal("test", result.Name)
		ts.Require().Equal(42, result.Value)
	})

	ts.Run("falls back to second path when first fails", func() {
		nonExistent := filepath.Join(ts.tempDir, "nonexistent.json")
		yamlFile := ts.createTestFile("config.yaml", fileLoaderTestYAML)

		var result testConfig
		loadedPath, err := ts.loader.Load([]string{nonExistent, yamlFile}, &result)

		ts.Require().NoError(err)
		ts.Require().Equal(yamlFile, loadedPath) //nolint:testifylint // comparing file paths, not YAML content
		ts.Require().Equal("test", result.Name)
		ts.Require().Equal(42, result.Value)
	})

	ts.Run("prepends default path to search list", func() {
		jsonFile := ts.createTestFile("default.json", fileLoaderTestJSON)
		loader := &FileConfigLoader{defaultPath: jsonFile}

		var result testConfig
		loadedPath, err := loader.Load([]string{}, &result)

		ts.Require().NoError(err)
		ts.Require().Equal(jsonFile, loadedPath) //nolint:testifylint // comparing file paths, not JSON content
		ts.Require().Equal("test", result.Name)
	})

	ts.Run("returns error when no files found", func() {
		nonExistent1 := filepath.Join(ts.tempDir, "nonexistent1.json")
		nonExistent2 := filepath.Join(ts.tempDir, "nonexistent2.yaml")

		var result testConfig
		_, err := ts.loader.Load([]string{nonExistent1, nonExistent2}, &result)

		ts.Require().Error(err)
		ts.Require().ErrorIs(err, errNoConfigFileFound)
	})

	ts.Run("returns error with empty path list and no default", func() {
		var result testConfig
		_, err := ts.loader.Load([]string{}, &result)

		ts.Require().Error(err)
		ts.Require().ErrorIs(err, errNoConfigFileFound)
	})
}

// TestFileConfigLoader_LoadFrom_JSON tests loading JSON files
func (ts *FileLoaderTestSuite) TestFileConfigLoader_LoadFrom_JSON() {
	ts.Run("loads valid JSON file", func() {
		jsonFile := ts.createTestFile("config.json", fileLoaderTestJSON)

		var result testConfig
		err := ts.loader.LoadFrom(jsonFile, &result)

		ts.Require().NoError(err)
		ts.Require().Equal("test", result.Name)
		ts.Require().Equal(42, result.Value)
	})

	ts.Run("loads JSON with nested structure", func() {
		nestedJSON := `{"outer": {"inner": "value"}, "count": 10}`
		jsonFile := ts.createTestFile("nested.json", nestedJSON)

		var result map[string]interface{}
		err := ts.loader.LoadFrom(jsonFile, &result)

		ts.Require().NoError(err)
		ts.Require().NotNil(result["outer"])
	})

	ts.Run("loads empty JSON object", func() {
		jsonFile := ts.createTestFile("empty.json", "{}")

		var result map[string]interface{}
		err := ts.loader.LoadFrom(jsonFile, &result)

		ts.Require().NoError(err)
		ts.Require().Empty(result)
	})
}

// TestFileConfigLoader_LoadFrom_YAML tests loading YAML files
func (ts *FileLoaderTestSuite) TestFileConfigLoader_LoadFrom_YAML() {
	ts.Run("loads valid YAML file with .yaml extension", func() {
		yamlFile := ts.createTestFile("config.yaml", fileLoaderTestYAML)

		var result testConfig
		err := ts.loader.LoadFrom(yamlFile, &result)

		ts.Require().NoError(err)
		ts.Require().Equal("test", result.Name)
		ts.Require().Equal(42, result.Value)
	})

	ts.Run("loads valid YAML file with .yml extension", func() {
		ymlFile := ts.createTestFile("config.yml", fileLoaderTestYAML)

		var result testConfig
		err := ts.loader.LoadFrom(ymlFile, &result)

		ts.Require().NoError(err)
		ts.Require().Equal("test", result.Name)
		ts.Require().Equal(42, result.Value)
	})

	ts.Run("loads YAML with nested structure", func() {
		nestedYAML := `outer:
  inner: value
count: 10`
		yamlFile := ts.createTestFile("nested.yaml", nestedYAML)

		var result map[string]interface{}
		err := ts.loader.LoadFrom(yamlFile, &result)

		ts.Require().NoError(err)
		ts.Require().NotNil(result["outer"])
	})

	ts.Run("loads empty YAML file", func() {
		yamlFile := ts.createTestFile("empty.yaml", "")

		var result map[string]interface{}
		err := ts.loader.LoadFrom(yamlFile, &result)

		ts.Require().NoError(err)
		// Empty YAML unmarshals to nil map
	})
}

// TestFileConfigLoader_LoadFrom_AutoDetect tests format auto-detection
func (ts *FileLoaderTestSuite) TestFileConfigLoader_LoadFrom_AutoDetect() {
	ts.Run("auto-detects JSON content with unknown extension", func() {
		jsonFile := ts.createTestFile("config.cfg", fileLoaderTestJSON)

		var result testConfig
		err := ts.loader.LoadFrom(jsonFile, &result)

		ts.Require().NoError(err)
		ts.Require().Equal("test", result.Name)
		ts.Require().Equal(42, result.Value)
	})

	ts.Run("auto-detects YAML content with unknown extension", func() {
		yamlFile := ts.createTestFile("config.conf", fileLoaderTestYAML)

		var result testConfig
		err := ts.loader.LoadFrom(yamlFile, &result)

		ts.Require().NoError(err)
		ts.Require().Equal("test", result.Name)
		ts.Require().Equal(42, result.Value)
	})

	ts.Run("auto-detects JSON content with no extension", func() {
		jsonFile := ts.createTestFile("config", fileLoaderTestJSON)

		var result testConfig
		err := ts.loader.LoadFrom(jsonFile, &result)

		ts.Require().NoError(err)
		ts.Require().Equal("test", result.Name)
	})
}

// TestFileConfigLoader_LoadFrom_Errors tests error conditions
func (ts *FileLoaderTestSuite) TestFileConfigLoader_LoadFrom_Errors() {
	ts.Run("returns error for relative path", func() {
		var result testConfig
		err := ts.loader.LoadFrom("relative/path.json", &result)

		ts.Require().Error(err)
		ts.Require().ErrorIs(err, errConfigPathNotAbs)
	})

	ts.Run("returns error for non-existent file", func() {
		nonExistent := filepath.Join(ts.tempDir, "nonexistent.json")

		var result testConfig
		err := ts.loader.LoadFrom(nonExistent, &result)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "failed to read config file")
	})

	ts.Run("returns error for unreadable content with unknown extension", func() {
		// Create file with content that's neither valid JSON nor YAML
		badFile := ts.createTestFile("config.xyz", "<<<not valid json or yaml>>>")

		var result testConfig
		err := ts.loader.LoadFrom(badFile, &result)

		ts.Require().Error(err)
		ts.Require().ErrorIs(err, errUnsupportedFileExt)
	})

	ts.Run("returns error for malformed JSON", func() {
		jsonFile := ts.createTestFile("malformed.json", fileLoaderMalformedJSON)

		var result testConfig
		err := ts.loader.LoadFrom(jsonFile, &result)

		ts.Require().Error(err)
		// JSON unmarshal error
	})

	ts.Run("case insensitive extension handling", func() {
		jsonFile := ts.createTestFile("config.JSON", fileLoaderTestJSON)

		var result testConfig
		err := ts.loader.LoadFrom(jsonFile, &result)

		ts.Require().NoError(err)
		ts.Require().Equal("test", result.Name)
	})
}

// TestFileConfigLoader_Save tests the Save method
func (ts *FileLoaderTestSuite) TestFileConfigLoader_Save() {
	testData := &testConfig{Name: "saved", Value: 100}

	ts.Run("saves as JSON format", func() {
		outputPath := filepath.Join(ts.tempDir, "output.json")

		err := ts.loader.Save(outputPath, testData, "json")
		ts.Require().NoError(err)

		// Verify file was created and content is valid
		content, readErr := os.ReadFile(outputPath) //nolint:gosec // G304: test file path is controlled
		ts.Require().NoError(readErr)
		ts.Require().Contains(string(content), "saved")
		ts.Require().Contains(string(content), "100")
	})

	ts.Run("saves as YAML format", func() {
		outputPath := filepath.Join(ts.tempDir, "output.yaml")

		err := ts.loader.Save(outputPath, testData, "yaml")
		ts.Require().NoError(err)

		// Verify file was created and content is valid
		content, readErr := os.ReadFile(outputPath) //nolint:gosec // G304: test file path is controlled
		ts.Require().NoError(readErr)
		ts.Require().Contains(string(content), "saved")
		ts.Require().Contains(string(content), "100")
	})

	ts.Run("saves as YML format alias", func() {
		outputPath := filepath.Join(ts.tempDir, "output.yml")

		err := ts.loader.Save(outputPath, testData, "yml")
		ts.Require().NoError(err)

		ts.Require().FileExists(outputPath)
	})

	ts.Run("creates directory if not exists", func() {
		nestedPath := filepath.Join(ts.tempDir, "nested", "dir", "config.json")

		err := ts.loader.Save(nestedPath, testData, "json")
		ts.Require().NoError(err)

		ts.Require().FileExists(nestedPath)
	})

	ts.Run("returns error for unsupported format", func() {
		outputPath := filepath.Join(ts.tempDir, "output.xml")

		err := ts.loader.Save(outputPath, testData, "xml")

		ts.Require().Error(err)
		ts.Require().ErrorIs(err, errUnsupportedFormat)
	})

	ts.Run("handles case insensitive format", func() {
		outputPath := filepath.Join(ts.tempDir, "output_upper.json")

		err := ts.loader.Save(outputPath, testData, "JSON")
		ts.Require().NoError(err)

		ts.Require().FileExists(outputPath)
	})

	ts.Run("saves complex nested structure", func() {
		complexData := map[string]interface{}{
			"nested": map[string]interface{}{
				"key": "value",
			},
			"array": []string{"a", "b", "c"},
		}
		outputPath := filepath.Join(ts.tempDir, "complex.json")

		err := ts.loader.Save(outputPath, complexData, "json")
		ts.Require().NoError(err)

		// Verify round-trip
		var loaded map[string]interface{}
		loadErr := ts.loader.LoadFrom(outputPath, &loaded)
		ts.Require().NoError(loadErr)
		ts.Require().NotNil(loaded["nested"])
	})

	ts.Run("returns error when marshaling fails for JSON", func() {
		// Channels cannot be marshaled to JSON
		unmarshalable := map[string]interface{}{
			"channel": make(chan int),
		}
		outputPath := filepath.Join(ts.tempDir, "fail.json")

		err := ts.loader.Save(outputPath, unmarshalable, "json")
		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "failed to marshal data")
	})

	// Note: YAML marshaling failures cause panic in the yaml.v3 library rather than
	// returning an error, so we cannot test that path without recovering from panic.
}

// TestFileConfigLoader_Validate tests the Validate method
func (ts *FileLoaderTestSuite) TestFileConfigLoader_Validate() {
	ts.Run("returns error for nil data", func() {
		err := ts.loader.Validate(nil)

		ts.Require().Error(err)
		ts.Require().ErrorIs(err, errConfigDataNil)
	})

	ts.Run("passes for non-nil struct", func() {
		data := &testConfig{Name: "test", Value: 42}

		err := ts.loader.Validate(data)
		ts.Require().NoError(err)
	})

	ts.Run("passes for non-nil map", func() {
		data := map[string]interface{}{"key": "value"}

		err := ts.loader.Validate(data)
		ts.Require().NoError(err)
	})

	ts.Run("passes for empty struct", func() {
		data := &testConfig{}

		err := ts.loader.Validate(data)
		ts.Require().NoError(err)
	})

	ts.Run("passes for empty map", func() {
		data := map[string]interface{}{}

		err := ts.loader.Validate(data)
		ts.Require().NoError(err)
	})
}

// TestFileConfigLoader_GetSupportedFormats tests GetSupportedFormats
func (ts *FileLoaderTestSuite) TestFileConfigLoader_GetSupportedFormats() {
	formats := ts.loader.GetSupportedFormats()

	ts.Require().Len(formats, 3)
	ts.Require().Contains(formats, "json")
	ts.Require().Contains(formats, "yaml")
	ts.Require().Contains(formats, "yml")
}

// TestFileLoaderTestSuite runs the test suite
func TestFileLoaderTestSuite(t *testing.T) {
	suite.Run(t, new(FileLoaderTestSuite))
}
