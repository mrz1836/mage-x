// Package fileops provides file system operations and permission constants.
package fileops

import "os"

// File permission constants with clear semantic meaning.
// These follow the principle of least privilege and should be used
// instead of magic octal numbers throughout the codebase.
//
// Permission meanings (octal notation):
//   - First digit: owner permissions
//   - Second digit: group permissions
//   - Third digit: other (world) permissions
//
// Each digit is the sum of: read (4), write (2), execute (1)
const (
	// PermFile is the default permission for regular files.
	// Owner can read/write, group and others can read.
	// Use for: documentation, generated files, non-sensitive output.
	PermFile os.FileMode = 0o644

	// PermFileSensitive is for configuration files and credentials.
	// Only the owner can read and write.
	// Use for: .env files, API keys, certificates, coverage reports.
	PermFileSensitive os.FileMode = 0o600

	// PermFileExecutable is for executable scripts and binaries.
	// Owner can read/write/execute, group and others can read/execute.
	// Use for: shell scripts, built binaries, public git hooks.
	PermFileExecutable os.FileMode = 0o755

	// PermFileExecutablePrivate is for owner-only executable scripts.
	// Only the owner can read, write, and execute.
	// Use for: private git hooks, sensitive scripts.
	PermFileExecutablePrivate os.FileMode = 0o700

	// PermDir is the default permission for directories.
	// Owner can read/write/traverse, group and others can read/traverse.
	// Use for: documentation directories, output directories.
	PermDir os.FileMode = 0o755

	// PermDirSensitive is for directories with restricted access.
	// Owner can do everything, group can read/traverse, others have no access.
	// Use for: build directories, cache directories, temp directories.
	PermDirSensitive os.FileMode = 0o750

	// PermDirPrivate is for strictly private directories.
	// Only the owner can access.
	// Use for: credential storage, private caches.
	PermDirPrivate os.FileMode = 0o700
)

// IsOwnerOnly returns true if the permission mode restricts access to only the owner.
// This is useful for validating that sensitive files have appropriate permissions.
func IsOwnerOnly(mode os.FileMode) bool {
	// Extract just the permission bits (lower 9 bits)
	perm := mode.Perm()
	// Check that group and others have no permissions
	return perm&0o077 == 0
}

// HasGroupAccess returns true if the permission mode allows group access.
func HasGroupAccess(mode os.FileMode) bool {
	perm := mode.Perm()
	// Check if group has any permissions
	return perm&0o070 != 0
}

// HasOtherAccess returns true if the permission mode allows world access.
func HasOtherAccess(mode os.FileMode) bool {
	perm := mode.Perm()
	// Check if others have any permissions
	return perm&0o007 != 0
}

// IsExecutable returns true if the permission mode has the owner execute bit set.
func IsExecutable(mode os.FileMode) bool {
	return mode.Perm()&0o100 != 0
}

// IsWritable returns true if the permission mode has the owner write bit set.
func IsWritable(mode os.FileMode) bool {
	return mode.Perm()&0o200 != 0
}

// IsReadable returns true if the permission mode has the owner read bit set.
func IsReadable(mode os.FileMode) bool {
	return mode.Perm()&0o400 != 0
}
