package fileops

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Static errors
var (
	ErrNoValidConfigFileFound = errors.New("no valid config file found in paths")
	ErrFileDoesNotExist       = errors.New("file does not exist")
)

// FileOps provides a facade combining all file operations
type FileOps struct {
	File FileOperator
	JSON JSONOperator
	YAML YAMLOperator
	Safe SafeFileOperator
}

// NewFileOperator creates a new default file operator
func NewFileOperator() FileOperator {
	return NewDefaultFileOperator()
}

// New creates a new FileOps with default implementations
func New() *FileOps {
	fileOp := NewDefaultFileOperator()
	return &FileOps{
		File: fileOp,
		JSON: NewDefaultJSONOperator(fileOp),
		YAML: NewDefaultYAMLOperator(fileOp),
		Safe: NewDefaultSafeFileOperator(),
	}
}

// NewWithOptions creates a new FileOps with custom implementations
func NewWithOptions(file FileOperator, json JSONOperator, yaml YAMLOperator, safe SafeFileOperator) *FileOps {
	return &FileOps{
		File: file,
		JSON: json,
		YAML: yaml,
		Safe: safe,
	}
}

// Convenience methods that handle common patterns

// WriteJSONSafe writes JSON data atomically with directory creation
func (f *FileOps) WriteJSONSafe(path string, data interface{}) error {
	if err := f.ensureDir(path); err != nil {
		return err
	}

	jsonData, err := f.JSON.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return f.Safe.WriteFileAtomic(path, jsonData, PermFile)
}

// WriteYAMLSafe writes YAML data atomically with directory creation
func (f *FileOps) WriteYAMLSafe(path string, data interface{}) error {
	if err := f.ensureDir(path); err != nil {
		return err
	}

	yamlData, err := f.YAML.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return f.Safe.WriteFileAtomic(path, yamlData, PermFileSensitive)
}

// LoadConfig loads configuration from file with fallback to multiple paths and formats
func (f *FileOps) LoadConfig(paths []string, dest interface{}) (string, error) {
	for _, path := range paths {
		if !f.File.Exists(path) {
			continue
		}

		ext := filepath.Ext(path)
		var err error

		switch ext {
		case ".yaml", ".yml":
			err = f.YAML.ReadYAML(path, dest)
		case ".json":
			err = f.JSON.ReadJSON(path, dest)
		default:
			// Try YAML first, then JSON
			err = f.YAML.ReadYAML(path, dest)
			if err != nil {
				err = f.JSON.ReadJSON(path, dest)
			}
		}

		if err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("%w: %v", ErrNoValidConfigFileFound, paths)
}

// SaveConfig saves configuration to file in the specified format
func (f *FileOps) SaveConfig(path string, data interface{}, format string) error {
	if err := f.ensureDir(path); err != nil {
		return err
	}

	switch format {
	case "yaml", "yml":
		return f.WriteYAMLSafe(path, data)
	case "json":
		return f.WriteJSONSafe(path, data)
	default:
		// Default to YAML
		return f.WriteYAMLSafe(path, data)
	}
}

// CopyFile copies a file with error handling and parent directory creation
func (f *FileOps) CopyFile(src, dst string) error {
	if err := f.ensureDir(dst); err != nil {
		return err
	}

	return f.File.Copy(src, dst)
}

// BackupFile creates a backup of a file with timestamp
func (f *FileOps) BackupFile(path string) error {
	if !f.File.Exists(path) {
		return fmt.Errorf("%w: %s", ErrFileDoesNotExist, path)
	}

	backupPath := path + ".bak"
	return f.File.Copy(path, backupPath)
}

// CleanupBackups removes backup files matching a pattern
func (f *FileOps) CleanupBackups(dir, pattern string) error {
	matches, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return fmt.Errorf("failed to glob pattern: %w", err)
	}

	for _, match := range matches {
		if err := f.File.Remove(match); err != nil {
			return fmt.Errorf("failed to remove backup %s: %w", match, err)
		}
	}

	return nil
}

// ensureDir creates parent directory if it doesn't exist
func (f *FileOps) ensureDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "/" {
		return nil
	}

	if !f.File.Exists(dir) {
		return f.File.MkdirAll(dir, PermDir)
	}

	return nil
}

// GetDefault returns a new FileOps instance with default configuration.
func GetDefault() *FileOps {
	return New()
}

// Package-level convenience functions

// WriteJSONSafe writes JSON data atomically using the default instance
func WriteJSONSafe(path string, data interface{}) error {
	return GetDefault().WriteJSONSafe(path, data)
}

// WriteYAMLSafe writes YAML data atomically using the default instance
func WriteYAMLSafe(path string, data interface{}) error {
	return GetDefault().WriteYAMLSafe(path, data)
}

// LoadConfig loads configuration using the default instance
func LoadConfig(paths []string, dest interface{}) (string, error) {
	return GetDefault().LoadConfig(paths, dest)
}

// SaveConfig saves configuration using the default instance
func SaveConfig(path string, data interface{}, format string) error {
	return GetDefault().SaveConfig(path, data, format)
}

// Exists checks if a file exists using the default instance
func Exists(path string) bool {
	return GetDefault().File.Exists(path)
}

// ReadFile reads a file using the default instance
func ReadFile(path string) ([]byte, error) {
	return GetDefault().File.ReadFile(path)
}

// WriteFile writes a file using the default instance
func WriteFile(path string, data []byte, perm os.FileMode) error {
	return GetDefault().File.WriteFile(path, data, perm)
}

// Remove removes a file using the default instance
func Remove(path string) error {
	return GetDefault().File.Remove(path)
}

// IsDir checks if path is a directory using the default instance
func IsDir(path string) bool {
	return GetDefault().File.IsDir(path)
}

// IsFile checks if path is a file using the default instance
func IsFile(path string) bool {
	if GetDefault().File.Exists(path) && !GetDefault().File.IsDir(path) {
		return true
	}
	return false
}

// MkdirAll creates directories using the default instance
func MkdirAll(path string, perm os.FileMode) error {
	return GetDefault().File.MkdirAll(path, perm)
}

// Copy copies a file using the default instance
func Copy(src, dst string) error {
	return GetDefault().File.Copy(src, dst)
}
