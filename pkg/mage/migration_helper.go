// Package mage provides migration helpers for updating file operations
package mage

import (
	"github.com/mrz1836/go-mage/pkg/common/fileops"
)

// MigrationHelper provides utilities for migrating to fileops
type MigrationHelper struct {
	fileOps *fileops.FileOps
}

// NewMigrationHelper creates a new migration helper
func NewMigrationHelper() *MigrationHelper {
	return &MigrationHelper{
		fileOps: fileops.New(),
	}
}

// GetFileOps returns the centralized fileops instance
func (m *MigrationHelper) GetFileOps() *fileops.FileOps {
	return m.fileOps
}

// Global instance for convenience
var migrationHelper = NewMigrationHelper()

// GetFileOps returns the global fileops instance for migration
func GetFileOps() *fileops.FileOps {
	return migrationHelper.GetFileOps()
}
