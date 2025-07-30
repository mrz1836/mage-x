package fileops

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DefaultFileOperator implements FileOperator using standard library
type DefaultFileOperator struct{}

// NewDefaultFileOperator creates a new default file operator
func NewDefaultFileOperator() *DefaultFileOperator {
	return &DefaultFileOperator{}
}

// ReadFile reads the entire file and returns its contents
func (d *DefaultFileOperator) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile writes data to a file with the specified permissions
func (d *DefaultFileOperator) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// Exists checks if a file or directory exists
func (d *DefaultFileOperator) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if the path is a directory
func (d *DefaultFileOperator) IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// MkdirAll creates a directory and all necessary parents
func (d *DefaultFileOperator) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Remove deletes a file or empty directory
func (d *DefaultFileOperator) Remove(path string) error {
	return os.Remove(path)
}

// RemoveAll deletes a path and any children it contains
func (d *DefaultFileOperator) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Stat returns file info
func (d *DefaultFileOperator) Stat(path string) (fs.FileInfo, error) {
	return os.Stat(path)
}

// Chmod changes file permissions
func (d *DefaultFileOperator) Chmod(path string, mode os.FileMode) error {
	return os.Chmod(path, mode)
}

// Copy copies a file from src to dst
func (d *DefaultFileOperator) Copy(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() {
		if closeErr := sourceFile.Close(); closeErr != nil {
			// Log error but don't override the main operation error
			fmt.Fprintf(os.Stderr, "Warning: failed to close source file: %v\n", closeErr)
		}
	}()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() {
		if closeErr := destFile.Close(); closeErr != nil {
			// Log error but don't override the main operation error
			fmt.Fprintf(os.Stderr, "Warning: failed to close destination file: %v\n", closeErr)
		}
	}()

	if _, copyErr := io.Copy(destFile, sourceFile); copyErr != nil {
		return fmt.Errorf("failed to copy file: %w", copyErr)
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

// ReadDir reads the directory and returns directory entries
func (d *DefaultFileOperator) ReadDir(path string) ([]fs.DirEntry, error) {
	return os.ReadDir(path)
}

// DefaultJSONOperator implements JSONOperator using encoding/json
type DefaultJSONOperator struct {
	fileOps FileOperator
}

// NewDefaultJSONOperator creates a new default JSON operator
func NewDefaultJSONOperator(fileOps FileOperator) *DefaultJSONOperator {
	if fileOps == nil {
		fileOps = NewDefaultFileOperator()
	}
	return &DefaultJSONOperator{fileOps: fileOps}
}

// Marshal converts a value to JSON
func (d *DefaultJSONOperator) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// MarshalIndent converts a value to formatted JSON
func (d *DefaultJSONOperator) MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

// Unmarshal parses JSON data into a value
func (d *DefaultJSONOperator) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// WriteJSON writes a value as JSON to a file
func (d *DefaultJSONOperator) WriteJSON(path string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return d.fileOps.WriteFile(path, data, 0o644)
}

// WriteJSONIndent writes a value as formatted JSON to a file
func (d *DefaultJSONOperator) WriteJSONIndent(path string, v interface{}, prefix, indent string) error {
	data, err := json.MarshalIndent(v, prefix, indent)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return d.fileOps.WriteFile(path, data, 0o644)
}

// ReadJSON reads JSON from a file into a value
func (d *DefaultJSONOperator) ReadJSON(path string, v interface{}) error {
	data, err := d.fileOps.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if unmarshalErr := json.Unmarshal(data, v); unmarshalErr != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", unmarshalErr)
	}

	return nil
}

// DefaultYAMLOperator implements YAMLOperator using gopkg.in/yaml.v3
type DefaultYAMLOperator struct {
	fileOps FileOperator
}

// NewDefaultYAMLOperator creates a new default YAML operator
func NewDefaultYAMLOperator(fileOps FileOperator) *DefaultYAMLOperator {
	if fileOps == nil {
		fileOps = NewDefaultFileOperator()
	}
	return &DefaultYAMLOperator{fileOps: fileOps}
}

// Marshal converts a value to YAML
func (d *DefaultYAMLOperator) Marshal(v interface{}) ([]byte, error) {
	return yaml.Marshal(v)
}

// Unmarshal parses YAML data into a value
func (d *DefaultYAMLOperator) Unmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

// WriteYAML writes a value as YAML to a file
func (d *DefaultYAMLOperator) WriteYAML(path string, v interface{}) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return d.fileOps.WriteFile(path, data, 0o644)
}

// ReadYAML reads YAML from a file into a value
func (d *DefaultYAMLOperator) ReadYAML(path string, v interface{}) error {
	data, err := d.fileOps.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	if unmarshalErr := yaml.Unmarshal(data, v); unmarshalErr != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", unmarshalErr)
	}

	return nil
}

// DefaultSafeFileOperator implements SafeFileOperator with atomic operations
type DefaultSafeFileOperator struct {
	*DefaultFileOperator
}

// NewDefaultSafeFileOperator creates a new safe file operator
func NewDefaultSafeFileOperator() *DefaultSafeFileOperator {
	return &DefaultSafeFileOperator{
		DefaultFileOperator: NewDefaultFileOperator(),
	}
}

// WriteFileAtomic writes a file atomically using a temporary file
func (d *DefaultSafeFileOperator) WriteFileAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)

	// Create temp file in the same directory
	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Clean up temp file on error
	defer func() {
		if err != nil {
			if removeErr := os.Remove(tmpPath); removeErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to remove temp file %s: %v\n", tmpPath, removeErr)
			}
		}
	}()

	// Write data to temp file
	if _, err = tmpFile.Write(data); err != nil {
		if closeErr := tmpFile.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close temp file: %v\n", closeErr)
		}
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err = tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Set permissions
	if err = os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Atomic rename
	if err = os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// WriteFileWithBackup writes a file and keeps a backup of the original
func (d *DefaultSafeFileOperator) WriteFileWithBackup(path string, data []byte, perm os.FileMode) error {
	// Check if file exists
	if d.Exists(path) {
		// Create backup
		backupPath := path + ".bak"
		if err := d.Copy(path, backupPath); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Write the file
	return d.WriteFile(path, data, perm)
}
