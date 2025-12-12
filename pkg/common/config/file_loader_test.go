package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type FileLoaderTestConfig struct {
	Name    string `json:"name" yaml:"name"`
	Version int    `json:"version" yaml:"version"`
}

func TestNewFileLoader(t *testing.T) {
	loader := NewFileLoader("/default/path")
	require.NotNil(t, loader)
}

func TestFileConfigLoader_Load(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a dummy config file
	configData := FileLoaderTestConfig{Name: "test", Version: 1}
	jsonContent, _ := json.Marshal(configData)
	configPath := filepath.Join(tmpDir, "config.json")
	err := os.WriteFile(configPath, jsonContent, 0o600)
	require.NoError(t, err)

	t.Run("success with explicit path", func(t *testing.T) {
		loader := NewFileLoader("")
		var dest FileLoaderTestConfig
		path, err := loader.Load([]string{configPath}, &dest)
		assert.NoError(t, err)
		assert.Equal(t, configPath, path)
		assert.Equal(t, configData, dest)
	})

	t.Run("success with default path", func(t *testing.T) {
		loader := NewFileLoader(configPath)
		var dest FileLoaderTestConfig
		path, err := loader.Load([]string{}, &dest)
		assert.NoError(t, err)
		assert.Equal(t, configPath, path)
		assert.Equal(t, configData, dest)
	})

	t.Run("failure - no config found", func(t *testing.T) {
		loader := NewFileLoader("")
		var dest FileLoaderTestConfig
		path, err := loader.Load([]string{filepath.Join(tmpDir, "nonexistent.json")}, &dest)
		assert.Error(t, err)
		assert.Equal(t, "", path)
		assert.ErrorIs(t, err, errNoConfigFileFound)
	})
}

func TestFileConfigLoader_LoadFrom(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("load json", func(t *testing.T) {
		configData := FileLoaderTestConfig{Name: "json", Version: 1}
		jsonContent, _ := json.Marshal(configData)
		configPath := filepath.Join(tmpDir, "config.json")
		err := os.WriteFile(configPath, jsonContent, 0o600)
		require.NoError(t, err)

		loader := NewFileLoader("")
		var dest FileLoaderTestConfig
		err = loader.LoadFrom(configPath, &dest)
		assert.NoError(t, err)
		assert.Equal(t, configData, dest)
	})

	t.Run("load yaml", func(t *testing.T) {
		configData := FileLoaderTestConfig{Name: "yaml", Version: 2}
		yamlContent, _ := yaml.Marshal(configData)
		configPath := filepath.Join(tmpDir, "config.yaml")
		err := os.WriteFile(configPath, yamlContent, 0o600)
		require.NoError(t, err)

		loader := NewFileLoader("")
		var dest FileLoaderTestConfig
		err = loader.LoadFrom(configPath, &dest)
		assert.NoError(t, err)
		assert.Equal(t, configData, dest)
	})

	t.Run("load unknown extension - fallback to json", func(t *testing.T) {
		configData := FileLoaderTestConfig{Name: "fallback_json", Version: 3}
		jsonContent, _ := json.Marshal(configData)
		configPath := filepath.Join(tmpDir, "config.unknown")
		err := os.WriteFile(configPath, jsonContent, 0o600)
		require.NoError(t, err)

		loader := NewFileLoader("")
		var dest FileLoaderTestConfig
		err = loader.LoadFrom(configPath, &dest)
		assert.NoError(t, err)
		assert.Equal(t, configData, dest)
	})

	t.Run("load unknown extension - fallback to yaml", func(t *testing.T) {
		configData := FileLoaderTestConfig{Name: "fallback_yaml", Version: 4}
		yamlContent, _ := yaml.Marshal(configData)
		configPath := filepath.Join(tmpDir, "config.unknown_yaml")
		err := os.WriteFile(configPath, yamlContent, 0o600)
		require.NoError(t, err)

		loader := NewFileLoader("")
		var dest FileLoaderTestConfig
		err = loader.LoadFrom(configPath, &dest)
		assert.NoError(t, err)
		assert.Equal(t, configData, dest)
	})

	t.Run("error - not absolute path", func(t *testing.T) {
		loader := NewFileLoader("")
		var dest FileLoaderTestConfig
		err := loader.LoadFrom("relative/path/config.json", &dest)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errConfigPathNotAbs)
	})

	t.Run("error - file not found", func(t *testing.T) {
		loader := NewFileLoader("")
		var dest FileLoaderTestConfig
		absPath := filepath.Join(tmpDir, "nonexistent.json")
		err := loader.LoadFrom(absPath, &dest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read config file")
	})

	t.Run("error - unsupported format", func(t *testing.T) {
		configPath := filepath.Join(tmpDir, "config.txt")
		err := os.WriteFile(configPath, []byte("invalid content"), 0o600)
		require.NoError(t, err)

		loader := NewFileLoader("")
		var dest FileLoaderTestConfig
		err = loader.LoadFrom(configPath, &dest)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errUnsupportedFileExt)
	})
}

func TestFileConfigLoader_Save(t *testing.T) {
	tmpDir := t.TempDir()
	loader := NewFileLoader("")

	t.Run("save json", func(t *testing.T) {
		configData := FileLoaderTestConfig{Name: "save_json", Version: 1}
		configPath := filepath.Join(tmpDir, "saved.json")

		err := loader.Save(configPath, configData, "json")
		assert.NoError(t, err)

		content, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var dest FileLoaderTestConfig
		err = json.Unmarshal(content, &dest)
		require.NoError(t, err)
		assert.Equal(t, configData, dest)
	})

	t.Run("save yaml", func(t *testing.T) {
		configData := FileLoaderTestConfig{Name: "save_yaml", Version: 2}
		configPath := filepath.Join(tmpDir, "saved.yaml")

		err := loader.Save(configPath, configData, "yaml")
		assert.NoError(t, err)

		content, err := os.ReadFile(configPath)
		require.NoError(t, err)

		var dest FileLoaderTestConfig
		err = yaml.Unmarshal(content, &dest)
		require.NoError(t, err)
		assert.Equal(t, configData, dest)
	})

	t.Run("save unsupported format", func(t *testing.T) {
		configData := FileLoaderTestConfig{Name: "fail", Version: 0}
		configPath := filepath.Join(tmpDir, "fail.txt")

		err := loader.Save(configPath, configData, "txt")
		assert.Error(t, err)
		assert.ErrorIs(t, err, errUnsupportedFormat)
	})

	t.Run("save create directory failure", func(t *testing.T) {
		// Create a file where we expect a directory
		pathAsFile := filepath.Join(tmpDir, "file_as_dir")
		err := os.WriteFile(pathAsFile, []byte("data"), 0o600)
		require.NoError(t, err)

		configData := FileLoaderTestConfig{Name: "fail_dir", Version: 0}
		configPath := filepath.Join(pathAsFile, "config.json")

		err = loader.Save(configPath, configData, "json")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create directory")
	})

    t.Run("fail to marshal", func(t *testing.T) {
        // channel is not marshally-able
        data := map[string]interface{}{
            "foo": make(chan int),
        }
        err := loader.Save(filepath.Join(tmpDir, "whatever.json"), data, "json")
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "failed to marshal data")
    })
}

func TestFileConfigLoader_Validate(t *testing.T) {
	loader := NewFileLoader("")

	t.Run("valid data", func(t *testing.T) {
		data := FileLoaderTestConfig{Name: "valid", Version: 1}
		err := loader.Validate(data)
		assert.NoError(t, err)
	})

	t.Run("nil data", func(t *testing.T) {
		err := loader.Validate(nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errConfigDataNil)
	})
}

func TestFileConfigLoader_GetSupportedFormats(t *testing.T) {
	loader := NewFileLoader("")
	formats := loader.GetSupportedFormats()
	assert.Contains(t, formats, "json")
	assert.Contains(t, formats, "yaml")
	assert.Contains(t, formats, "yml")
}
