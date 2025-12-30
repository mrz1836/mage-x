package mage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockBuildOperations provides test mocking for Build operations.
// It implements the BuildOperations interface.
// Note: This is defined in the mage package to avoid import cycles,
// since BuildOperations uses *Config which is defined in this package.
type MockBuildOperations struct {
	mock.Mock
}

// DeterminePackagePath determines the package path to build.
func (m *MockBuildOperations) DeterminePackagePath(cfg *Config, outputPath string, requireMain bool) (string, error) {
	args := m.Called(cfg, outputPath, requireMain)
	return args.String(0), args.Error(1)
}

// NewMockBuildOperations creates a new mock Build operations instance.
func NewMockBuildOperations() *MockBuildOperations {
	return &MockBuildOperations{}
}

// TestBuildOperationsInterface tests the BuildOperations interface implementation
func TestBuildOperationsInterface(t *testing.T) {
	t.Run("DefaultBuildOperations implements BuildOperations", func(t *testing.T) {
		var _ BuildOperations = DefaultBuildOperations{}
	})

	t.Run("MockBuildOperations implements BuildOperations", func(t *testing.T) {
		var _ BuildOperations = &MockBuildOperations{}
	})
}

// TestBuildOperationsProviderFunctions tests the provider functions
func TestBuildOperationsProviderFunctions(t *testing.T) {
	t.Run("GetBuildOperations returns default implementation", func(t *testing.T) {
		// Reset first to ensure clean state
		ResetBuildOperations()

		ops := GetBuildOperations()
		require.NotNil(t, ops)
		_, ok := ops.(DefaultBuildOperations)
		assert.True(t, ok, "Expected DefaultBuildOperations")
	})

	t.Run("SetBuildOperations rejects nil", func(t *testing.T) {
		err := SetBuildOperations(nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("SetBuildOperations accepts custom implementation", func(t *testing.T) {
		original := GetBuildOperations()
		defer func() {
			require.NoError(t, SetBuildOperations(original))
		}()

		mockOps := NewMockBuildOperations()
		err := SetBuildOperations(mockOps)
		require.NoError(t, err)

		retrieved := GetBuildOperations()
		assert.Equal(t, mockOps, retrieved)
	})

	t.Run("ResetBuildOperations restores default", func(t *testing.T) {
		mockOps := NewMockBuildOperations()
		require.NoError(t, SetBuildOperations(mockOps))

		ResetBuildOperations()

		ops := GetBuildOperations()
		_, ok := ops.(DefaultBuildOperations)
		assert.True(t, ok, "Expected DefaultBuildOperations after reset")
	})
}

// TestMockBuildOperations tests the mock implementation
func TestMockBuildOperations(t *testing.T) {
	t.Run("DeterminePackagePath returns mocked values", func(t *testing.T) {
		mockOps := NewMockBuildOperations()
		mockOps.On("DeterminePackagePath", mock.Anything, "/path/to/output", true).
			Return("./cmd/app", nil)

		path, err := mockOps.DeterminePackagePath(&Config{}, "/path/to/output", true)
		require.NoError(t, err)
		assert.Equal(t, "./cmd/app", path)

		mockOps.AssertExpectations(t)
	})

	t.Run("DeterminePackagePath returns error when configured", func(t *testing.T) {
		mockOps := NewMockBuildOperations()
		mockOps.On("DeterminePackagePath", mock.Anything, mock.Anything, mock.Anything).
			Return("", assert.AnError)

		path, err := mockOps.DeterminePackagePath(&Config{}, "/path", false)
		require.Error(t, err)
		assert.Empty(t, path)

		mockOps.AssertExpectations(t)
	})
}
