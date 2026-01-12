// Package fileops provides secure file operation utilities for mage-x,
// with consistent permission handling and safety checks.
//
// # Permissions
//
// Predefined permission constants for secure file operations:
//
//	fileops.PermFile           // 0644 - Standard file permissions
//	fileops.PermFileExecutable // 0755 - Executable file permissions
//	fileops.PermDir            // 0755 - Directory permissions
//	fileops.PermDirRestricted  // 0700 - Restricted directory permissions
//
// # File Operations
//
// Safe file operations with proper error handling:
//
//	// Write with atomic operations
//	err := fileops.WriteFile(path, data, fileops.PermFile)
//
//	// Copy with permission preservation
//	err := fileops.CopyFile(src, dst)
//
//	// Ensure directory exists
//	err := fileops.EnsureDir(path)
//
// # Security
//
// All file operations include:
//
//   - Path validation to prevent traversal attacks
//   - Consistent permission enforcement
//   - Atomic write operations where possible
//
// # Integration
//
// This package is used throughout mage-x for consistent file handling,
// ensuring security best practices are followed for all file operations.
package fileops
