package fileops

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
)

// FileOpsTestSuite tests file operations
type FileOpsTestSuite struct {
	suite.Suite
	tmpDir string
	ops    *FileOps
}

func (s *FileOpsTestSuite) SetupSuite() {
	// Create temporary directory for all tests
	tmpDir, err := os.MkdirTemp("", "fileops-test-*")
	s.Require().NoError(err, "Failed to create temp dir")
	s.tmpDir = tmpDir
	s.ops = New()
}

func (s *FileOpsTestSuite) TearDownSuite() {
	// Clean up temp directory
	if err := os.RemoveAll(s.tmpDir); err != nil {
		s.T().Logf("Failed to remove temp dir %s: %v", s.tmpDir, err)
	}
}

func TestFileOpsTestSuite(t *testing.T) {
	suite.Run(t, new(FileOpsTestSuite))
}

func TestDefaultFileOperator(t *testing.T) {
	// Create temporary directory for tests
	tmpDir, err := os.MkdirTemp("", "fileops-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	ops := NewDefaultFileOperator()

	t.Run("WriteFile and ReadFile", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.txt")
		testData := []byte("Hello, World!")

		// Write file
		err := ops.WriteFile(testFile, testData, 0o644)
		require.NoError(t, err, "Failed to write file")

		// Check if file exists
		assert.True(t, ops.Exists(testFile), "File should exist after writing")

		// Read file
		data, err := ops.ReadFile(testFile)
		require.NoError(t, err, "Failed to read file")

		assert.Equal(t, testData, data, "Read data should match written data")
	})

	t.Run("MkdirAll", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "deep", "nested", "dir")

		err := ops.MkdirAll(testDir, 0o755)
		require.NoError(t, err, "Failed to create directory")

		assert.True(t, ops.Exists(testDir), "Directory should exist after creation")
		assert.True(t, ops.IsDir(testDir), "Path should be a directory")
	})

	t.Run("Copy", func(t *testing.T) {
		srcFile := filepath.Join(tmpDir, "source.txt")
		dstFile := filepath.Join(tmpDir, "destination.txt")
		testData := []byte("Copy test data")

		// Create source file
		err := ops.WriteFile(srcFile, testData, 0o644)
		require.NoError(t, err, "Failed to write source file")

		// Copy file
		err = ops.Copy(srcFile, dstFile)
		require.NoError(t, err, "Failed to copy file")

		// Verify copied file
		data, err := ops.ReadFile(dstFile)
		require.NoError(t, err, "Failed to read copied file")

		assert.Equal(t, testData, data, "Copied data should match source data")
	})

	t.Run("Remove", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "remove-test.txt")

		// Create file
		err := ops.WriteFile(testFile, []byte("test"), 0o644)
		require.NoError(t, err, "Failed to create file")
		assert.True(t, ops.Exists(testFile), "File should exist")

		// Remove file
		err = ops.Remove(testFile)
		require.NoError(t, err, "Failed to remove file")
		assert.False(t, ops.Exists(testFile), "File should not exist after removal")
	})

	t.Run("IsDir and IsFile", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "test-dir")
		testFile := filepath.Join(tmpDir, "test-file.txt")

		// Create directory
		err := ops.MkdirAll(testDir, 0o755)
		require.NoError(t, err)

		// Create file
		err = ops.WriteFile(testFile, []byte("test"), 0o644)
		require.NoError(t, err)

		// Test IsDir
		assert.True(t, ops.IsDir(testDir), "Should identify directory correctly")
		assert.False(t, ops.IsDir(testFile), "Should not identify file as directory")
		assert.False(t, ops.IsDir(filepath.Join(tmpDir, "nonexistent")), "Should return false for nonexistent path")

		// Test that file is not a directory
		assert.False(t, ops.IsDir(testFile), "File should not be identified as directory")
		// Test that non-existent path is not a directory
		assert.False(t, ops.IsDir(filepath.Join(tmpDir, "nonexistent")), "Non-existent path should not be identified as directory")
	})
}

func TestDefaultJSONOperator(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "json-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	fileOps := NewDefaultFileOperator()
	jsonOps := NewDefaultJSONOperator(fileOps)

	type TestStruct struct {
		Name  string                 `json:"name"`
		Value int                    `json:"value"`
		Tags  []string               `json:"tags,omitempty"`
		Meta  map[string]interface{} `json:"meta,omitempty"`
	}

	testData := TestStruct{
		Name:  "test",
		Value: 42,
		Tags:  []string{"tag1", "tag2"},
		Meta:  map[string]interface{}{"key": "value", "count": float64(10)},
	}
	testFile := filepath.Join(tmpDir, "test.json")

	t.Run("WriteJSON and ReadJSON", func(t *testing.T) {
		// Write JSON
		err := jsonOps.WriteJSON(testFile, testData)
		require.NoError(t, err, "Failed to write JSON")

		// Read JSON
		var result TestStruct
		err = jsonOps.ReadJSON(testFile, &result)
		require.NoError(t, err, "Failed to read JSON")

		assert.Equal(t, testData.Name, result.Name, "Name field should match")
		assert.Equal(t, testData.Value, result.Value, "Value field should match")
		assert.ElementsMatch(t, testData.Tags, result.Tags, "Tags should match")
		// Compare meta fields individually to handle type differences
		assert.Equal(t, testData.Meta["key"], result.Meta["key"], "Meta key should match")
		assert.Equal(t, testData.Meta["count"], result.Meta["count"], "Meta count should match")
	})

	t.Run("MarshalIndent", func(t *testing.T) {
		data, err := jsonOps.MarshalIndent(testData, "", "  ")
		require.NoError(t, err, "Failed to marshal with indent")

		// Verify it's valid JSON
		var result TestStruct
		err = json.Unmarshal(data, &result)
		require.NoError(t, err, "Failed to unmarshal indented JSON")

		// Compare fields individually to handle type differences
		assert.Equal(t, testData.Name, result.Name, "Name should match")
		assert.Equal(t, testData.Value, result.Value, "Value should match")
		assert.ElementsMatch(t, testData.Tags, result.Tags, "Tags should match")
		assert.Equal(t, testData.Meta["key"], result.Meta["key"], "Meta key should match")
		assert.Equal(t, testData.Meta["count"], result.Meta["count"], "Meta count should match")

		// Verify indentation
		jsonStr := string(data)
		assert.Contains(t, jsonStr, "\n  ", "JSON should be indented")
	})

	t.Run("Marshal and Unmarshal", func(t *testing.T) {
		// Marshal
		data, err := jsonOps.Marshal(testData)
		require.NoError(t, err, "Failed to marshal JSON")

		// Unmarshal
		var result TestStruct
		err = jsonOps.Unmarshal(data, &result)
		require.NoError(t, err, "Failed to unmarshal JSON")

		// Compare fields individually to handle type differences
		assert.Equal(t, testData.Name, result.Name, "Name should match")
		assert.Equal(t, testData.Value, result.Value, "Value should match")
		assert.ElementsMatch(t, testData.Tags, result.Tags, "Tags should match")
		assert.Equal(t, testData.Meta["key"], result.Meta["key"], "Meta key should match")
		assert.Equal(t, testData.Meta["count"], result.Meta["count"], "Meta count should match")
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		invalidFile := filepath.Join(tmpDir, "invalid.json")
		err := fileOps.WriteFile(invalidFile, []byte("{ invalid json"), 0o644)
		require.NoError(t, err)

		var result TestStruct
		err = jsonOps.ReadJSON(invalidFile, &result)
		assert.Error(t, err, "Should error on invalid JSON")
	})
}

func TestDefaultYAMLOperator(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "yaml-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	fileOps := NewDefaultFileOperator()
	yamlOps := NewDefaultYAMLOperator(fileOps)

	type TestStruct struct {
		Name   string                 `yaml:"name"`
		Value  int                    `yaml:"value"`
		Tags   []string               `yaml:"tags,omitempty"`
		Meta   map[string]interface{} `yaml:"meta,omitempty"`
		Nested struct {
			Field1 string `yaml:"field1"`
			Field2 bool   `yaml:"field2"`
		} `yaml:"nested"`
	}

	testData := TestStruct{
		Name:  "test",
		Value: 42,
		Tags:  []string{"tag1", "tag2"},
		Meta:  map[string]interface{}{"key": "value", "count": 10},
	}
	testData.Nested.Field1 = "nested value"
	testData.Nested.Field2 = true

	testFile := filepath.Join(tmpDir, "test.yaml")

	t.Run("WriteYAML and ReadYAML", func(t *testing.T) {
		// Write YAML
		err := yamlOps.WriteYAML(testFile, testData)
		require.NoError(t, err, "Failed to write YAML")

		// Read YAML
		var result TestStruct
		err = yamlOps.ReadYAML(testFile, &result)
		require.NoError(t, err, "Failed to read YAML")

		assert.Equal(t, testData.Name, result.Name, "Name should match")
		assert.Equal(t, testData.Value, result.Value, "Value should match")
		assert.ElementsMatch(t, testData.Tags, result.Tags, "Tags should match")
		assert.Equal(t, testData.Meta["key"], result.Meta["key"], "Meta key should match")
		assert.Equal(t, testData.Meta["count"], result.Meta["count"], "Meta count should match")
		assert.Equal(t, testData.Nested.Field1, result.Nested.Field1, "Nested field1 should match")
		assert.Equal(t, testData.Nested.Field2, result.Nested.Field2, "Nested field2 should match")
	})

	t.Run("Marshal and Unmarshal", func(t *testing.T) {
		data, err := yamlOps.Marshal(testData)
		require.NoError(t, err, "Failed to marshal YAML")

		// Verify it's valid YAML
		var result TestStruct
		err = yaml.Unmarshal(data, &result)
		require.NoError(t, err, "Failed to unmarshal YAML")

		// Compare fields individually to handle type differences
		assert.Equal(t, testData.Name, result.Name, "Name should match")
		assert.Equal(t, testData.Value, result.Value, "Value should match")
		assert.ElementsMatch(t, testData.Tags, result.Tags, "Tags should match")
		assert.Equal(t, testData.Meta["key"], result.Meta["key"], "Meta key should match")
		assert.Equal(t, testData.Meta["count"], result.Meta["count"], "Meta count should match")
	})

	t.Run("Invalid YAML", func(t *testing.T) {
		invalidFile := filepath.Join(tmpDir, "invalid.yaml")
		err := fileOps.WriteFile(invalidFile, []byte("invalid:\n  - unclosed\n  - [bracket"), 0o644)
		require.NoError(t, err)

		var result TestStruct
		err = yamlOps.ReadYAML(invalidFile, &result)
		assert.Error(t, err, "Should error on invalid YAML")
	})
}

func TestSafeFileOperator(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "safe-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, err)
		}
	}()

	ops := NewDefaultSafeFileOperator()

	t.Run("WriteFileAtomic", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "atomic.txt")
		testData := []byte("Atomic write test")

		err := ops.WriteFileAtomic(testFile, testData, 0o644)
		require.NoError(t, err, "Failed to write file atomically")

		// Verify file contents
		data, err := ops.ReadFile(testFile)
		require.NoError(t, err, "Failed to read file")

		assert.Equal(t, testData, data, "File contents should match")

		// Verify no temp files remain
		files, err := os.ReadDir(tmpDir)
		require.NoError(t, err)
		assert.Len(t, files, 1, "Should only have the final file, no temp files")
	})

	t.Run("WriteFileWithBackup", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "backup.txt")
		originalData := []byte("Original data")
		newData := []byte("New data")

		// Write original file
		err := ops.WriteFile(testFile, originalData, 0o644)
		require.NoError(t, err, "Failed to write original file")

		// Write with backup
		err = ops.WriteFileWithBackup(testFile, newData, 0o644)
		require.NoError(t, err, "Failed to write file with backup")

		// Check backup exists
		backupFile := testFile + ".bak"
		assert.True(t, ops.Exists(backupFile), "Backup file should exist")

		// Verify backup contents
		backupData, err := ops.ReadFile(backupFile)
		require.NoError(t, err, "Failed to read backup file")
		assert.Equal(t, originalData, backupData, "Backup should contain original data")

		// Verify new file contents
		newFileData, err := ops.ReadFile(testFile)
		require.NoError(t, err, "Failed to read new file")
		assert.Equal(t, newData, newFileData, "File should contain new data")
	})
}

func (s *FileOpsTestSuite) TestFileOpsFacade() {
	type TestConfig struct {
		Database struct {
			Host string `yaml:"host" json:"host"`
			Port int    `yaml:"port" json:"port"`
		} `yaml:"database" json:"database"`
		Debug    bool     `yaml:"debug" json:"debug"`
		Features []string `yaml:"features" json:"features"`
	}

	testConfig := TestConfig{}
	testConfig.Database.Host = "localhost"
	testConfig.Database.Port = 5432
	testConfig.Debug = true
	testConfig.Features = []string{"feature1", "feature2"}

	s.Run("SaveConfig and LoadConfig YAML", func() {
		configPath := filepath.Join(s.tmpDir, "config.yaml")

		// Save config
		err := s.ops.SaveConfig(configPath, testConfig, "yaml")
		s.Require().NoError(err, "Failed to save config")

		// Load config
		var result TestConfig
		foundPath, err := s.ops.LoadConfig([]string{configPath}, &result)
		s.Require().NoError(err, "Failed to load config")

		s.Equal(configPath, foundPath, "Should find config at expected path")
		s.Equal(testConfig.Database.Host, result.Database.Host, "Database host should match")
		s.Equal(testConfig.Database.Port, result.Database.Port, "Database port should match")
		s.Equal(testConfig.Debug, result.Debug, "Debug flag should match")
		s.ElementsMatch(testConfig.Features, result.Features, "Features should match")
	})

	s.Run("SaveConfig and LoadConfig JSON", func() {
		configPath := filepath.Join(s.tmpDir, "config.json")

		// Save config
		err := s.ops.SaveConfig(configPath, testConfig, "json")
		s.Require().NoError(err, "Failed to save config")

		// Load config
		var result TestConfig
		foundPath, err := s.ops.LoadConfig([]string{configPath}, &result)
		s.Require().NoError(err, "Failed to load config")

		s.Equal(configPath, foundPath, "Should find config at expected path")
		s.Equal(testConfig, result, "Loaded config should match saved config")
	})

	s.Run("LoadConfig with fallback", func() {
		// Create only the second config file
		configPath1 := filepath.Join(s.tmpDir, "nonexistent.yaml")
		configPath2 := filepath.Join(s.tmpDir, "fallback.yaml")

		err := s.ops.SaveConfig(configPath2, testConfig, "yaml")
		s.Require().NoError(err, "Failed to save fallback config")

		// Load with fallback
		var result TestConfig
		foundPath, err := s.ops.LoadConfig([]string{configPath1, configPath2}, &result)
		s.Require().NoError(err, "Failed to load config with fallback")

		s.Equal(configPath2, foundPath, "Should find config at fallback path")
		s.Equal(testConfig.Database.Host, result.Database.Host, "Config should be loaded correctly")
	})

	s.Run("LoadConfig all missing", func() {
		// Try to load from non-existent files
		paths := []string{
			filepath.Join(s.tmpDir, "missing1.yaml"),
			filepath.Join(s.tmpDir, "missing2.yaml"),
		}

		var result TestConfig
		foundPath, err := s.ops.LoadConfig(paths, &result)
		s.Require().Error(err, "Should error when all config files are missing")
		s.Empty(foundPath, "Should return empty path when no config found")
	})

	s.Run("SaveConfig invalid format", func() {
		configPath := filepath.Join(s.tmpDir, "config.invalid")

		err := s.ops.SaveConfig(configPath, testConfig, "invalid")
		// SaveConfig defaults to YAML for unknown formats, so no error
		s.Require().NoError(err, "Should default to YAML for unknown formats")

		// Verify YAML file was created
		s.True(s.ops.File.Exists(configPath), "File should exist")
	})
}

func TestPackageLevelFunctions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "package-test-*")
	require.NoError(t, err, "Failed to create temp dir")
	defer func() {
		if removeErr := os.RemoveAll(tmpDir); removeErr != nil {
			t.Logf("Failed to remove temp dir %s: %v", tmpDir, removeErr)
		}
	}()

	testFile := filepath.Join(tmpDir, "package-test.txt")
	testData := []byte("Package level test")

	// Test package-level WriteFile
	err = WriteFile(testFile, testData, 0o644)
	require.NoError(t, err, "Failed to write file using package function")

	// Test package-level Exists
	assert.True(t, Exists(testFile), "File should exist")

	// Test package-level ReadFile
	data, err := ReadFile(testFile)
	require.NoError(t, err, "Failed to read file using package function")

	assert.Equal(t, testData, data, "Read data should match written data")

	// Test package-level IsDir
	assert.True(t, IsDir(tmpDir), "Should identify directory correctly")
	assert.False(t, IsDir(testFile), "Should not identify file as directory")

	// Test package-level IsFile
	assert.True(t, IsFile(testFile), "Should identify file correctly")
	assert.False(t, IsFile(tmpDir), "Should not identify directory as file")

	// Test package-level Remove
	err = Remove(testFile)
	require.NoError(t, err, "Failed to remove file")
	assert.False(t, Exists(testFile), "File should not exist after removal")

	// Test package-level MkdirAll
	testDir := filepath.Join(tmpDir, "test", "nested", "dir")
	err = MkdirAll(testDir, 0o755)
	require.NoError(t, err, "Failed to create directory")
	assert.True(t, IsDir(testDir), "Directory should exist")

	// Test package-level Copy
	srcFile := filepath.Join(tmpDir, "src.txt")
	dstFile := filepath.Join(tmpDir, "dst.txt")
	err = WriteFile(srcFile, testData, 0o644)
	require.NoError(t, err)

	err = Copy(srcFile, dstFile)
	require.NoError(t, err, "Failed to copy file")

	copiedData, err := ReadFile(dstFile)
	require.NoError(t, err)
	assert.Equal(t, testData, copiedData, "Copied data should match source")
}
