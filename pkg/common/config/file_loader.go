package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/mrz1836/mage-x/pkg/common/fileops"
)

// Static errors to comply with err113 linter
var (
	errNoConfigFileFound  = errors.New("no configuration file found in paths")
	errConfigPathNotAbs   = errors.New("config file path must be absolute")
	errUnsupportedFileExt = errors.New("unsupported file format")
	errUnsupportedFormat  = errors.New("unsupported format")
	errConfigDataNil      = errors.New("configuration data is nil")
)

// FileConfigLoader implements ConfigLoader for file-based configurations
type FileConfigLoader struct {
	defaultPath string
}

// NewFileLoader creates a new file-based configuration loader
func NewFileLoader(defaultPath string) Loader {
	return &FileConfigLoader{
		defaultPath: defaultPath,
	}
}

// Load loads configuration from multiple file paths with fallback
func (f *FileConfigLoader) Load(paths []string, dest interface{}) (string, error) {
	// Add default path if provided
	if f.defaultPath != "" {
		paths = append([]string{f.defaultPath}, paths...)
	}

	for _, path := range paths {
		if err := f.LoadFrom(path, dest); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("%w: %v", errNoConfigFileFound, paths)
}

// LoadFrom loads configuration from a specific file
func (f *FileConfigLoader) LoadFrom(path string, dest interface{}) error {
	// Clean and validate the path
	cleanPath := filepath.Clean(path)
	if !filepath.IsAbs(cleanPath) {
		return fmt.Errorf("%w: %s", errConfigPathNotAbs, path)
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return json.Unmarshal(data, dest)
	case ".yaml", ".yml":
		return yaml.Unmarshal(data, dest)
	default:
		// Try to detect format from content
		if err := json.Unmarshal(data, dest); err == nil {
			return nil
		}
		if err := yaml.Unmarshal(data, dest); err == nil {
			return nil
		}
		return fmt.Errorf("%w: %s", errUnsupportedFileExt, ext)
	}
}

// Save saves configuration to a file in the specified format
func (f *FileConfigLoader) Save(path string, data interface{}, format string) error {
	var content []byte
	var err error

	switch strings.ToLower(format) {
	case "json":
		content, err = json.MarshalIndent(data, "", "  ")
	case "yaml", "yml":
		content, err = yaml.Marshal(data)
	default:
		return fmt.Errorf("%w: %s", errUnsupportedFormat, format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, fileops.PermDirSensitive); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return os.WriteFile(path, content, fileops.PermFileSensitive)
}

// Validate validates configuration data
func (f *FileConfigLoader) Validate(data interface{}) error {
	// Basic validation - ensure data is not nil
	if data == nil {
		return errConfigDataNil
	}

	// Additional validation can be added here based on specific requirements
	return nil
}

// GetSupportedFormats returns list of supported file formats
func (f *FileConfigLoader) GetSupportedFormats() []string {
	return []string{"json", "yaml", "yml"}
}
