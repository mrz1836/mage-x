// Package fileops provides file operation utilities with mockable interfaces
package fileops

import (
	"io/fs"
	"os"
)

// FileOperator provides an interface for file system operations
type FileOperator interface {
	// ReadFile reads the entire file and returns its contents
	ReadFile(path string) ([]byte, error)

	// WriteFile writes data to a file with the specified permissions
	WriteFile(path string, data []byte, perm os.FileMode) error

	// Exists checks if a file or directory exists
	Exists(path string) bool

	// IsDir checks if the path is a directory
	IsDir(path string) bool

	// MkdirAll creates a directory and all necessary parents
	MkdirAll(path string, perm os.FileMode) error

	// Remove deletes a file or empty directory
	Remove(path string) error

	// RemoveAll deletes a path and any children it contains
	RemoveAll(path string) error

	// Stat returns file info
	Stat(path string) (fs.FileInfo, error)

	// Chmod changes file permissions
	Chmod(path string, mode os.FileMode) error

	// Copy copies a file from src to dst
	Copy(src, dst string) error

	// ReadDir reads the directory and returns directory entries
	ReadDir(path string) ([]fs.DirEntry, error)
}

// JSONOperator provides an interface for JSON operations
type JSONOperator interface {
	// Marshal converts a value to JSON
	Marshal(v interface{}) ([]byte, error)

	// MarshalIndent converts a value to formatted JSON
	MarshalIndent(v interface{}, prefix, indent string) ([]byte, error)

	// Unmarshal parses JSON data into a value
	Unmarshal(data []byte, v interface{}) error

	// WriteJSON writes a value as JSON to a file
	WriteJSON(path string, v interface{}) error

	// WriteJSONIndent writes a value as formatted JSON to a file
	WriteJSONIndent(path string, v interface{}, prefix, indent string) error

	// ReadJSON reads JSON from a file into a value
	ReadJSON(path string, v interface{}) error
}

// YAMLOperator provides an interface for YAML operations
type YAMLOperator interface {
	// Marshal converts a value to YAML
	Marshal(v interface{}) ([]byte, error)

	// Unmarshal parses YAML data into a value
	Unmarshal(data []byte, v interface{}) error

	// WriteYAML writes a value as YAML to a file
	WriteYAML(path string, v interface{}) error

	// ReadYAML reads YAML from a file into a value
	ReadYAML(path string, v interface{}) error
}

// SafeFileOperator provides atomic file operations
type SafeFileOperator interface {
	FileOperator

	// WriteFileAtomic writes a file atomically using a temporary file
	WriteFileAtomic(path string, data []byte, perm os.FileMode) error

	// WriteFileWithBackup writes a file and keeps a backup of the original
	WriteFileWithBackup(path string, data []byte, perm os.FileMode) error
}

// Options for file operations
type Options struct {
	// CreateDirs creates parent directories if they don't exist
	CreateDirs bool

	// Backup creates a backup before overwriting
	Backup bool

	// Atomic uses atomic write operations
	Atomic bool

	// Permissions for new files (defaults to 0644)
	FileMode os.FileMode

	// Permissions for new directories (defaults to 0755)
	DirMode os.FileMode
}
