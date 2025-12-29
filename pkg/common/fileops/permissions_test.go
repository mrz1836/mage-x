package fileops

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissionConstants(t *testing.T) {
	t.Run("PermFile equals 0o644", func(t *testing.T) {
		assert.Equal(t, PermFile, os.FileMode(0o644))
		// Verify meaning: owner rw, group r, other r
		assert.True(t, IsReadable(PermFile))
		assert.True(t, IsWritable(PermFile))
		assert.False(t, IsExecutable(PermFile))
		assert.True(t, HasGroupAccess(PermFile))
		assert.True(t, HasOtherAccess(PermFile))
	})

	t.Run("PermFileSensitive equals 0o600", func(t *testing.T) {
		assert.Equal(t, PermFileSensitive, os.FileMode(0o600))
		// Verify meaning: owner rw only
		assert.True(t, IsReadable(PermFileSensitive))
		assert.True(t, IsWritable(PermFileSensitive))
		assert.False(t, IsExecutable(PermFileSensitive))
		assert.True(t, IsOwnerOnly(PermFileSensitive))
	})

	t.Run("PermFileExecutable equals 0o755", func(t *testing.T) {
		assert.Equal(t, PermFileExecutable, os.FileMode(0o755))
		// Verify meaning: owner rwx, group rx, other rx
		assert.True(t, IsReadable(PermFileExecutable))
		assert.True(t, IsWritable(PermFileExecutable))
		assert.True(t, IsExecutable(PermFileExecutable))
		assert.True(t, HasGroupAccess(PermFileExecutable))
		assert.True(t, HasOtherAccess(PermFileExecutable))
	})

	t.Run("PermFileExecutablePrivate equals 0o700", func(t *testing.T) {
		assert.Equal(t, PermFileExecutablePrivate, os.FileMode(0o700))
		// Verify meaning: owner only, executable
		assert.True(t, IsReadable(PermFileExecutablePrivate))
		assert.True(t, IsWritable(PermFileExecutablePrivate))
		assert.True(t, IsExecutable(PermFileExecutablePrivate))
		assert.True(t, IsOwnerOnly(PermFileExecutablePrivate))
	})

	t.Run("PermDir equals 0o755", func(t *testing.T) {
		assert.Equal(t, PermDir, os.FileMode(0o755))
		// Same as executable - for directories, x means traverse
		assert.True(t, IsExecutable(PermDir))
		assert.True(t, HasGroupAccess(PermDir))
		assert.True(t, HasOtherAccess(PermDir))
	})

	t.Run("PermDirSensitive equals 0o750", func(t *testing.T) {
		assert.Equal(t, PermDirSensitive, os.FileMode(0o750))
		// Verify meaning: owner rwx, group rx, other none
		assert.True(t, IsReadable(PermDirSensitive))
		assert.True(t, IsWritable(PermDirSensitive))
		assert.True(t, IsExecutable(PermDirSensitive))
		assert.True(t, HasGroupAccess(PermDirSensitive))
		assert.False(t, HasOtherAccess(PermDirSensitive))
	})

	t.Run("PermDirPrivate equals 0o700", func(t *testing.T) {
		assert.Equal(t, PermDirPrivate, os.FileMode(0o700))
		// Verify meaning: owner only
		assert.True(t, IsReadable(PermDirPrivate))
		assert.True(t, IsWritable(PermDirPrivate))
		assert.True(t, IsExecutable(PermDirPrivate))
		assert.True(t, IsOwnerOnly(PermDirPrivate))
	})
}

func TestIsOwnerOnly(t *testing.T) {
	tests := []struct {
		name     string
		mode     os.FileMode
		expected bool
	}{
		{"0o600 is owner only", 0o600, true},
		{"0o700 is owner only", 0o700, true},
		{"0o644 is not owner only", 0o644, false},
		{"0o755 is not owner only", 0o755, false},
		{"0o640 is not owner only", 0o640, false},
		{"0o604 is not owner only", 0o604, false},
		{"0o000 is owner only (no perms)", 0o000, true},
		{"0o400 is owner only (read only)", 0o400, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsOwnerOnly(tt.mode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasGroupAccess(t *testing.T) {
	tests := []struct {
		name     string
		mode     os.FileMode
		expected bool
	}{
		{"0o640 has group read", 0o640, true},
		{"0o650 has group read/execute", 0o650, true},
		{"0o670 has group rwx", 0o670, true},
		{"0o600 has no group access", 0o600, false},
		{"0o604 has no group access", 0o604, false},
		{"0o750 has group access", 0o750, true},
		{"0o755 has group access", 0o755, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasGroupAccess(tt.mode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHasOtherAccess(t *testing.T) {
	tests := []struct {
		name     string
		mode     os.FileMode
		expected bool
	}{
		{"0o644 has other read", 0o644, true},
		{"0o645 has other read/execute", 0o645, true},
		{"0o647 has other rwx", 0o647, true},
		{"0o640 has no other access", 0o640, false},
		{"0o750 has no other access", 0o750, false},
		{"0o600 has no other access", 0o600, false},
		{"0o755 has other access", 0o755, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasOtherAccess(tt.mode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsExecutable(t *testing.T) {
	tests := []struct {
		name     string
		mode     os.FileMode
		expected bool
	}{
		{"0o755 is executable", 0o755, true},
		{"0o700 is executable", 0o700, true},
		{"0o644 is not executable", 0o644, false},
		{"0o600 is not executable", 0o600, false},
		{"0o100 is executable (x only)", 0o100, true},
		{"0o111 is executable", 0o111, true},
		{"0o000 is not executable", 0o000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsExecutable(tt.mode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsWritable(t *testing.T) {
	tests := []struct {
		name     string
		mode     os.FileMode
		expected bool
	}{
		{"0o644 is writable", 0o644, true},
		{"0o600 is writable", 0o600, true},
		{"0o755 is writable", 0o755, true},
		{"0o444 is not writable", 0o444, false},
		{"0o555 is not writable", 0o555, false},
		{"0o200 is writable (w only)", 0o200, true},
		{"0o000 is not writable", 0o000, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsWritable(tt.mode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsReadable(t *testing.T) {
	tests := []struct {
		name     string
		mode     os.FileMode
		expected bool
	}{
		{"0o644 is readable", 0o644, true},
		{"0o600 is readable", 0o600, true},
		{"0o755 is readable", 0o755, true},
		{"0o400 is readable (r only)", 0o400, true},
		{"0o000 is not readable", 0o000, false},
		{"0o200 is not readable (w only)", 0o200, false},
		{"0o100 is not readable (x only)", 0o100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsReadable(tt.mode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPermissionsIntegration(t *testing.T) {
	// Test that permissions work correctly with actual file operations
	t.Run("create file with PermFileSensitive", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := tmpDir + "/sensitive.txt"

		err := os.WriteFile(testFile, []byte("secret"), PermFileSensitive)
		require.NoError(t, err)

		info, err := os.Stat(testFile)
		require.NoError(t, err)

		// Verify the permission was applied correctly
		assert.True(t, IsOwnerOnly(info.Mode()))
	})

	t.Run("create directory with PermDirSensitive", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDir := tmpDir + "/sensitive"

		err := os.Mkdir(testDir, PermDirSensitive)
		require.NoError(t, err)

		info, err := os.Stat(testDir)
		require.NoError(t, err)

		// Verify group access but no other access
		assert.True(t, HasGroupAccess(info.Mode()))
		assert.False(t, HasOtherAccess(info.Mode()))
	})

	t.Run("create file with PermFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := tmpDir + "/public.txt"

		err := os.WriteFile(testFile, []byte("public"), PermFile)
		require.NoError(t, err)

		info, err := os.Stat(testFile)
		require.NoError(t, err)

		// Verify world-readable
		assert.True(t, HasOtherAccess(info.Mode()))
		assert.False(t, IsExecutable(info.Mode()))
	})
}

func TestPermissionDocumentation(t *testing.T) {
	// These tests serve as documentation for permission usage
	t.Run("PermFileSensitive for credentials and config", func(t *testing.T) {
		// 0o600 = rw------- (owner read/write only)
		// Use for: .env files, API keys, certificates
		assert.Equal(t, PermFileSensitive, os.FileMode(0o600))
	})

	t.Run("PermFile for general output files", func(t *testing.T) {
		// 0o644 = rw-r--r-- (owner rw, world readable)
		// Use for: generated docs, reports, config examples
		assert.Equal(t, PermFile, os.FileMode(0o644))
	})

	t.Run("PermFileExecutable for scripts and binaries", func(t *testing.T) {
		// 0o755 = rwxr-xr-x (owner rwx, world read/execute)
		// Use for: git hooks, shell scripts, built binaries
		assert.Equal(t, PermFileExecutable, os.FileMode(0o755))
	})

	t.Run("PermDirSensitive for build and cache directories", func(t *testing.T) {
		// 0o750 = rwxr-x--- (owner rwx, group rx, no other)
		// Use for: .cache, .mage, build artifacts
		assert.Equal(t, PermDirSensitive, os.FileMode(0o750))
	})

	t.Run("PermDirPrivate for strictly private directories", func(t *testing.T) {
		// 0o700 = rwx------ (owner only)
		// Use for: credential stores, private temp directories
		assert.Equal(t, PermDirPrivate, os.FileMode(0o700))
	})
}
