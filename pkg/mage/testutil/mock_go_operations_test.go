package testutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockGoOperationsGetModuleName tests GetModuleName method
func TestMockGoOperationsGetModuleName(t *testing.T) {
	t.Run("returns configured module name", func(t *testing.T) {
		mock := NewMockGoOperations()
		mock.On("GetModuleName").Return("github.com/test/module", nil)

		name, err := mock.GetModuleName()

		require.NoError(t, err)
		assert.Equal(t, "github.com/test/module", name)
		mock.AssertExpectations(t)
	})

	t.Run("returns configured error", func(t *testing.T) {
		mock := NewMockGoOperations()
		mock.On("GetModuleName").Return("", assert.AnError)

		name, err := mock.GetModuleName()

		require.ErrorIs(t, err, assert.AnError)
		assert.Empty(t, name)
		mock.AssertExpectations(t)
	})
}

// TestMockGoOperationsGetVersion tests GetVersion method
func TestMockGoOperationsGetVersion(t *testing.T) {
	t.Run("returns configured version", func(t *testing.T) {
		mock := NewMockGoOperations()
		mock.On("GetVersion").Return("v1.2.3")

		version := mock.GetVersion()

		assert.Equal(t, "v1.2.3", version)
		mock.AssertExpectations(t)
	})

	t.Run("returns dev version", func(t *testing.T) {
		mock := NewMockGoOperations()
		mock.On("GetVersion").Return("dev")

		version := mock.GetVersion()

		assert.Equal(t, "dev", version)
		mock.AssertExpectations(t)
	})
}

// TestMockGoOperationsGetGoVulnCheckVersion tests GetGoVulnCheckVersion method
func TestMockGoOperationsGetGoVulnCheckVersion(t *testing.T) {
	t.Run("returns configured vulncheck version", func(t *testing.T) {
		mock := NewMockGoOperations()
		mock.On("GetGoVulnCheckVersion").Return("v1.0.0")

		version := mock.GetGoVulnCheckVersion()

		assert.Equal(t, "v1.0.0", version)
		mock.AssertExpectations(t)
	})
}

// TestNewMockGoOperations tests constructor
func TestNewMockGoOperations(t *testing.T) {
	mock := NewMockGoOperations()

	require.NotNil(t, mock)
}

// TestNewMockGoBuilder tests builder constructor
func TestNewMockGoBuilder(t *testing.T) {
	ops, builder := NewMockGoBuilder()

	require.NotNil(t, ops)
	require.NotNil(t, builder)
}

// TestMockGoBuilderWithModuleName tests WithModuleName builder method
func TestMockGoBuilderWithModuleName(t *testing.T) {
	ops, builder := NewMockGoBuilder()

	result := builder.WithModuleName("github.com/test/project", nil)

	require.Same(t, builder, result)

	name, err := ops.GetModuleName()
	require.NoError(t, err)
	assert.Equal(t, "github.com/test/project", name)
}

// TestMockGoBuilderWithVersion tests WithVersion builder method
func TestMockGoBuilderWithVersion(t *testing.T) {
	ops, builder := NewMockGoBuilder()

	result := builder.WithVersion("v2.0.0")

	require.Same(t, builder, result)

	version := ops.GetVersion()
	assert.Equal(t, "v2.0.0", version)
}

// TestMockGoBuilderWithGoVulnCheckVersion tests WithGoVulnCheckVersion builder method
func TestMockGoBuilderWithGoVulnCheckVersion(t *testing.T) {
	ops, builder := NewMockGoBuilder()

	result := builder.WithGoVulnCheckVersion("v1.1.0")

	require.Same(t, builder, result)

	version := ops.GetGoVulnCheckVersion()
	assert.Equal(t, "v1.1.0", version)
}

// TestMockGoBuilderBuild tests Build method
func TestMockGoBuilderBuild(t *testing.T) {
	ops, builder := NewMockGoBuilder()
	builder.WithModuleName("test/module", nil).WithVersion("v1.0.0")

	built := builder.Build()

	require.Same(t, ops, built)
}

// TestMockGoBuilderChaining tests fluent chaining
func TestMockGoBuilderChaining(t *testing.T) {
	ops, builder := NewMockGoBuilder()

	built := builder.
		WithModuleName("github.com/example/project", nil).
		WithVersion("v1.0.0").
		WithGoVulnCheckVersion("v0.1.0").
		Build()

	require.Same(t, ops, built)

	// Verify all expectations work
	name, err := built.GetModuleName()
	require.NoError(t, err)
	assert.Equal(t, "github.com/example/project", name)

	assert.Equal(t, "v1.0.0", built.GetVersion())
	assert.Equal(t, "v0.1.0", built.GetGoVulnCheckVersion())
}
