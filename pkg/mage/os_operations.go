// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"errors"
	"os"

	"github.com/mrz1836/mage-x/pkg/common/providers"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for OS operations
var errOSOperationsNil = errors.New("os operations cannot be nil")

// OSOperations abstracts OS-level operations for testability.
// This interface allows mocking of OS operations in unit tests.
type OSOperations interface {
	// Getenv retrieves the value of the environment variable named by the key
	Getenv(key string) string
	// Remove removes the named file or (empty) directory
	Remove(path string) error
	// TempDir returns the default directory to use for temporary files
	TempDir() string
	// FileExists checks if a file exists at the given path
	FileExists(path string) bool
	// WriteFile writes data to the named file, creating it if necessary
	WriteFile(path string, data []byte, perm os.FileMode) error
	// Symlink creates newname as a symbolic link to oldname
	Symlink(oldname, newname string) error
	// Readlink returns the destination of the named symbolic link
	Readlink(name string) (string, error)
}

// DefaultOSOperations provides real OS implementations.
// This is the production implementation used when not testing.
type DefaultOSOperations struct{}

// Getenv retrieves the value of the environment variable named by the key.
func (DefaultOSOperations) Getenv(key string) string {
	return os.Getenv(key)
}

// Remove removes the named file or (empty) directory.
func (DefaultOSOperations) Remove(path string) error {
	return os.Remove(path)
}

// TempDir returns the default directory to use for temporary files.
func (DefaultOSOperations) TempDir() string {
	return os.TempDir()
}

// FileExists checks if a file exists at the given path.
func (DefaultOSOperations) FileExists(path string) bool {
	return utils.FileExists(path)
}

// WriteFile writes data to the named file, creating it if necessary.
func (DefaultOSOperations) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// Symlink creates newname as a symbolic link to oldname.
func (DefaultOSOperations) Symlink(oldname, newname string) error {
	return os.Symlink(oldname, newname)
}

// Readlink returns the destination of the named symbolic link.
func (DefaultOSOperations) Readlink(name string) (string, error) {
	return os.Readlink(name)
}

// Package-level provider (follows GetRunner pattern)
//
//nolint:gochecknoglobals // intentional package-level provider following GetRunner pattern
var osOpsProvider = providers.NewPackageProvider(func() OSOperations {
	return DefaultOSOperations{}
})

// GetOSOperations returns the current OS operations implementation.
// In production, this returns DefaultOSOperations.
// In tests, this can be replaced with a mock implementation.
func GetOSOperations() OSOperations {
	return osOpsProvider.Get()
}

// SetOSOperations allows setting a custom implementation (for testing).
// Returns an error if ops is nil.
func SetOSOperations(ops OSOperations) error {
	if ops == nil {
		return errOSOperationsNil
	}
	osOpsProvider.Set(ops)
	return nil
}

// ResetOSOperations resets the OS operations provider to its default state.
// This is primarily useful for testing cleanup.
func ResetOSOperations() {
	osOpsProvider.Reset()
}
