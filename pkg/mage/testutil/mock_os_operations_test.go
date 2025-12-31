package testutil

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockOSOperationsGetenv tests Getenv method
func TestMockOSOperationsGetenv(t *testing.T) {
	t.Run("returns configured env value", func(t *testing.T) {
		mock := NewMockOSOperations()
		mock.On("Getenv", "HOME").Return("/home/user")

		value := mock.Getenv("HOME")

		assert.Equal(t, "/home/user", value)
		mock.AssertExpectations(t)
	})

	t.Run("returns empty for unset env", func(t *testing.T) {
		mock := NewMockOSOperations()
		mock.On("Getenv", "UNSET_VAR").Return("")

		value := mock.Getenv("UNSET_VAR")

		assert.Empty(t, value)
		mock.AssertExpectations(t)
	})
}

// TestMockOSOperationsRemove tests Remove method
func TestMockOSOperationsRemove(t *testing.T) {
	t.Run("returns nil on success", func(t *testing.T) {
		mock := NewMockOSOperations()
		mock.On("Remove", "/tmp/test.txt").Return(nil)

		err := mock.Remove("/tmp/test.txt")

		require.NoError(t, err)
		mock.AssertExpectations(t)
	})

	t.Run("returns error on failure", func(t *testing.T) {
		mock := NewMockOSOperations()
		mock.On("Remove", "/protected/file").Return(assert.AnError)

		err := mock.Remove("/protected/file")

		require.ErrorIs(t, err, assert.AnError)
		mock.AssertExpectations(t)
	})
}

// TestMockOSOperationsTempDir tests TempDir method
func TestMockOSOperationsTempDir(t *testing.T) {
	t.Run("returns configured temp dir", func(t *testing.T) {
		mock := NewMockOSOperations()
		mock.On("TempDir").Return("/tmp")

		dir := mock.TempDir()

		assert.Equal(t, "/tmp", dir)
		mock.AssertExpectations(t)
	})
}

// TestMockOSOperationsFileExists tests FileExists method
func TestMockOSOperationsFileExists(t *testing.T) {
	t.Run("returns true when file exists", func(t *testing.T) {
		mock := NewMockOSOperations()
		mock.On("FileExists", "/path/to/file").Return(true)

		exists := mock.FileExists("/path/to/file")

		assert.True(t, exists)
		mock.AssertExpectations(t)
	})

	t.Run("returns false when file does not exist", func(t *testing.T) {
		mock := NewMockOSOperations()
		mock.On("FileExists", "/nonexistent").Return(false)

		exists := mock.FileExists("/nonexistent")

		assert.False(t, exists)
		mock.AssertExpectations(t)
	})
}

// TestMockOSOperationsWriteFile tests WriteFile method
func TestMockOSOperationsWriteFile(t *testing.T) {
	t.Run("returns nil on success", func(t *testing.T) {
		mock := NewMockOSOperations()
		mock.On("WriteFile", "/tmp/test.txt", []byte("content"), os.FileMode(0o600)).Return(nil)

		err := mock.WriteFile("/tmp/test.txt", []byte("content"), 0o600)

		require.NoError(t, err)
		mock.AssertExpectations(t)
	})

	t.Run("returns error on failure", func(t *testing.T) {
		mock := NewMockOSOperations()
		mock.On("WriteFile", "/tmp/fail.txt", []byte("data"), os.FileMode(0o600)).Return(assert.AnError)

		err := mock.WriteFile("/tmp/fail.txt", []byte("data"), 0o600)

		require.ErrorIs(t, err, assert.AnError)
		mock.AssertExpectations(t)
	})
}

// TestMockOSOperationsSymlink tests Symlink method
func TestMockOSOperationsSymlink(t *testing.T) {
	t.Run("returns nil on success", func(t *testing.T) {
		mock := NewMockOSOperations()
		mock.On("Symlink", "/actual/path", "/link/path").Return(nil)

		err := mock.Symlink("/actual/path", "/link/path")

		require.NoError(t, err)
		mock.AssertExpectations(t)
	})

	t.Run("returns error on failure", func(t *testing.T) {
		mock := NewMockOSOperations()
		mock.On("Symlink", "/actual", "/link").Return(assert.AnError)

		err := mock.Symlink("/actual", "/link")

		require.ErrorIs(t, err, assert.AnError)
		mock.AssertExpectations(t)
	})
}

// TestMockOSOperationsReadlink tests Readlink method
func TestMockOSOperationsReadlink(t *testing.T) {
	t.Run("returns target on success", func(t *testing.T) {
		mock := NewMockOSOperations()
		mock.On("Readlink", "/link/path").Return("/actual/path", nil)

		target, err := mock.Readlink("/link/path")

		require.NoError(t, err)
		assert.Equal(t, "/actual/path", target)
		mock.AssertExpectations(t)
	})

	t.Run("returns error for non-link", func(t *testing.T) {
		mock := NewMockOSOperations()
		mock.On("Readlink", "/regular/file").Return("", assert.AnError)

		target, err := mock.Readlink("/regular/file")

		require.ErrorIs(t, err, assert.AnError)
		assert.Empty(t, target)
		mock.AssertExpectations(t)
	})
}

// TestNewMockOSOperations tests constructor
func TestNewMockOSOperations(t *testing.T) {
	mock := NewMockOSOperations()

	require.NotNil(t, mock)
}

// TestNewMockOSBuilder tests builder constructor
func TestNewMockOSBuilder(t *testing.T) {
	ops, builder := NewMockOSBuilder()

	require.NotNil(t, ops)
	require.NotNil(t, builder)
}

// TestMockOSBuilderWithEnv tests WithEnv builder method
func TestMockOSBuilderWithEnv(t *testing.T) {
	ops, builder := NewMockOSBuilder()

	result := builder.WithEnv("HOME", "/home/test")

	require.Same(t, builder, result)

	value := ops.Getenv("HOME")
	assert.Equal(t, "/home/test", value)
}

// TestMockOSBuilderWithTempDir tests WithTempDir builder method
func TestMockOSBuilderWithTempDir(t *testing.T) {
	ops, builder := NewMockOSBuilder()

	result := builder.WithTempDir("/custom/tmp")

	require.Same(t, builder, result)

	dir := ops.TempDir()
	assert.Equal(t, "/custom/tmp", dir)
}

// TestMockOSBuilderWithFileExists tests WithFileExists builder method
func TestMockOSBuilderWithFileExists(t *testing.T) {
	ops, builder := NewMockOSBuilder()

	result := builder.WithFileExists("/path/to/file", true)

	require.Same(t, builder, result)

	exists := ops.FileExists("/path/to/file")
	assert.True(t, exists)
}

// TestMockOSBuilderWithRemoveSuccess tests WithRemoveSuccess builder method
func TestMockOSBuilderWithRemoveSuccess(t *testing.T) {
	ops, builder := NewMockOSBuilder()

	result := builder.WithRemoveSuccess("/tmp/file")

	require.Same(t, builder, result)

	err := ops.Remove("/tmp/file")
	require.NoError(t, err)
}

// TestMockOSBuilderWithRemoveError tests WithRemoveError builder method
func TestMockOSBuilderWithRemoveError(t *testing.T) {
	ops, builder := NewMockOSBuilder()

	result := builder.WithRemoveError("/tmp/file", assert.AnError)

	require.Same(t, builder, result)

	err := ops.Remove("/tmp/file")
	require.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}

// TestMockOSBuilderWithWriteFileSuccess tests WithWriteFileSuccess builder method
func TestMockOSBuilderWithWriteFileSuccess(t *testing.T) {
	ops, builder := NewMockOSBuilder()

	result := builder.WithWriteFileSuccess("/tmp/file.txt")

	require.Same(t, builder, result)

	err := ops.WriteFile("/tmp/file.txt", []byte("content"), 0o600)
	require.NoError(t, err)
}

// TestMockOSBuilderWithSymlinkSuccess tests WithSymlinkSuccess builder method
func TestMockOSBuilderWithSymlinkSuccess(t *testing.T) {
	ops, builder := NewMockOSBuilder()

	result := builder.WithSymlinkSuccess("/original", "/link")

	require.Same(t, builder, result)

	err := ops.Symlink("/original", "/link")
	require.NoError(t, err)
}

// TestMockOSBuilderWithReadlink tests WithReadlink builder method
func TestMockOSBuilderWithReadlink(t *testing.T) {
	ops, builder := NewMockOSBuilder()

	result := builder.WithReadlink("/link", "/target", nil)

	require.Same(t, builder, result)

	target, err := ops.Readlink("/link")
	require.NoError(t, err)
	assert.Equal(t, "/target", target)
}

// TestMockOSBuilderBuild tests Build method
func TestMockOSBuilderBuild(t *testing.T) {
	ops, builder := NewMockOSBuilder()
	builder.WithEnv("KEY", "value")

	built := builder.Build()

	require.Same(t, ops, built)
}

// TestMockOSBuilderChaining tests fluent chaining
func TestMockOSBuilderChaining(t *testing.T) {
	ops, builder := NewMockOSBuilder()

	built := builder.
		WithEnv("HOME", "/home/user").
		WithTempDir("/tmp").
		WithFileExists("/config", true).
		WithRemoveSuccess("/tmp/old").
		WithWriteFileSuccess("/tmp/new").
		WithSymlinkSuccess("/orig", "/link").
		WithReadlink("/link", "/orig", nil).
		Build()

	require.Same(t, ops, built)

	// Verify all expectations work
	assert.Equal(t, "/home/user", built.Getenv("HOME"))
	assert.Equal(t, "/tmp", built.TempDir())
	assert.True(t, built.FileExists("/config"))
	require.NoError(t, built.Remove("/tmp/old"))
	require.NoError(t, built.WriteFile("/tmp/new", []byte("data"), 0o600))
	require.NoError(t, built.Symlink("/orig", "/link"))

	target, err := built.Readlink("/link")
	require.NoError(t, err)
	assert.Equal(t, "/orig", target)
}
