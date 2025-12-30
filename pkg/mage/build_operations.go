// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"errors"

	"github.com/mrz1836/mage-x/pkg/common/providers"
)

// Static errors for Build operations
var errBuildOperationsNil = errors.New("build operations cannot be nil")

// BuildOperations abstracts Build operations for testability.
// This interface allows mocking of build-related operations in tests.
type BuildOperations interface {
	// DeterminePackagePath determines the package path to build.
	// It returns the package path and any error that occurred.
	DeterminePackagePath(cfg *Config, outputPath string, requireMain bool) (string, error)
}

// DefaultBuildOperations provides real Build implementations.
type DefaultBuildOperations struct{}

// DeterminePackagePath delegates to Build.determinePackagePath.
func (DefaultBuildOperations) DeterminePackagePath(cfg *Config, outputPath string, requireMain bool) (string, error) {
	return Build{}.determinePackagePath(cfg, outputPath, requireMain)
}

// Package-level provider (follows GetRunner pattern)
//
//nolint:gochecknoglobals // intentional package-level provider following GetRunner pattern
var buildOpsProvider = providers.NewPackageProvider(func() BuildOperations {
	return DefaultBuildOperations{}
})

// GetBuildOperations returns the current BuildOperations implementation.
func GetBuildOperations() BuildOperations {
	return buildOpsProvider.Get()
}

// SetBuildOperations allows setting a custom implementation (for testing).
func SetBuildOperations(ops BuildOperations) error {
	if ops == nil {
		return errBuildOperationsNil
	}
	buildOpsProvider.Set(ops)
	return nil
}

// ResetBuildOperations restores the default implementation.
func ResetBuildOperations() {
	buildOpsProvider.Reset()
}
