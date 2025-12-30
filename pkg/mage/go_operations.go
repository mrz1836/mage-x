// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"errors"

	"github.com/mrz1836/mage-x/pkg/common/providers"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for Go operations
var errGoOperationsNil = errors.New("go operations cannot be nil")

// GoOperations abstracts Go toolchain operations for testability.
// This interface allows mocking of Go-related operations in unit tests.
type GoOperations interface {
	// GetModuleName returns the current Go module name from go.mod
	GetModuleName() (string, error)
	// GetVersion returns the current project version
	GetVersion() string
	// GetGoVulnCheckVersion returns the govulncheck tool version
	GetGoVulnCheckVersion() string
}

// DefaultGoOperations provides real Go toolchain implementations.
// This is the production implementation used when not testing.
type DefaultGoOperations struct{}

// GetModuleName returns the current Go module name from go.mod.
func (DefaultGoOperations) GetModuleName() (string, error) {
	return utils.GetModuleName()
}

// GetVersion returns the current project version.
func (DefaultGoOperations) GetVersion() string {
	return getVersion()
}

// GetGoVulnCheckVersion returns the govulncheck tool version.
func (DefaultGoOperations) GetGoVulnCheckVersion() string {
	return GetDefaultGoVulnCheckVersion()
}

// Package-level provider (follows GetRunner pattern)
//
//nolint:gochecknoglobals // intentional package-level provider following GetRunner pattern
var goOpsProvider = providers.NewPackageProvider(func() GoOperations {
	return DefaultGoOperations{}
})

// GetGoOperations returns the current Go operations implementation.
// In production, this returns DefaultGoOperations.
// In tests, this can be replaced with a mock implementation.
func GetGoOperations() GoOperations {
	return goOpsProvider.Get()
}

// SetGoOperations allows setting a custom implementation (for testing).
// Returns an error if ops is nil.
func SetGoOperations(ops GoOperations) error {
	if ops == nil {
		return errGoOperationsNil
	}
	goOpsProvider.Set(ops)
	return nil
}

// ResetGoOperations resets the Go operations provider to its default state.
// This is primarily useful for testing cleanup.
func ResetGoOperations() {
	goOpsProvider.Reset()
}
