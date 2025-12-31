package mage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockGoOperations provides a mock implementation for testing
type MockGoOperations struct {
	ModuleName         string
	ModuleNameErr      error
	Version            string
	GoVulnCheckVersion string
}

// GetModuleName returns the mocked module name
func (m *MockGoOperations) GetModuleName() (string, error) {
	return m.ModuleName, m.ModuleNameErr
}

// GetVersion returns the mocked version
func (m *MockGoOperations) GetVersion() string {
	return m.Version
}

// GetGoVulnCheckVersion returns the mocked govulncheck version
func (m *MockGoOperations) GetGoVulnCheckVersion() string {
	return m.GoVulnCheckVersion
}

// TestDefaultGoOperationsGetModuleName tests DefaultGoOperations.GetModuleName
func TestDefaultGoOperationsGetModuleName(t *testing.T) {
	ops := DefaultGoOperations{}

	// Should not panic and return some result (depends on test environment)
	name, err := ops.GetModuleName()

	// In a valid Go module environment, it should succeed
	// but in test isolation it might fail - just verify no panic
	if err == nil {
		assert.NotEmpty(t, name)
	}
}

// TestDefaultGoOperationsGetVersion tests DefaultGoOperations.GetVersion
func TestDefaultGoOperationsGetVersion(t *testing.T) {
	ops := DefaultGoOperations{}

	// Should not panic and return some result
	version := ops.GetVersion()
	// Version can be empty, "dev", or a real version - just verify no panic
	_ = version
}

// TestDefaultGoOperationsGetGoVulnCheckVersion tests DefaultGoOperations.GetGoVulnCheckVersion
func TestDefaultGoOperationsGetGoVulnCheckVersion(t *testing.T) {
	ops := DefaultGoOperations{}

	// Should return a version string
	version := ops.GetGoVulnCheckVersion()
	// Should return the default version constant
	assert.NotEmpty(t, version)
}

// TestGetGoOperations tests GetGoOperations function
func TestGetGoOperations(t *testing.T) {
	// Reset to ensure clean state
	ResetGoOperations()
	t.Cleanup(func() {
		ResetGoOperations()
	})

	// Should return non-nil operations
	ops := GetGoOperations()
	require.NotNil(t, ops)

	// Should be the default implementation
	_, ok := ops.(DefaultGoOperations)
	assert.True(t, ok, "should return DefaultGoOperations by default")
}

// TestSetGoOperations tests SetGoOperations function
func TestSetGoOperations(t *testing.T) {
	// Reset to ensure clean state
	ResetGoOperations()
	t.Cleanup(func() {
		ResetGoOperations()
	})

	t.Run("set nil returns error", func(t *testing.T) {
		err := SetGoOperations(nil)
		require.ErrorIs(t, err, errGoOperationsNil)
	})

	t.Run("set valid mock", func(t *testing.T) {
		mock := &MockGoOperations{
			ModuleName:         "test/module",
			Version:            "v1.0.0",
			GoVulnCheckVersion: "v1.0.0",
		}

		err := SetGoOperations(mock)
		require.NoError(t, err)

		// Verify mock is used
		ops := GetGoOperations()
		require.Equal(t, mock, ops)

		name, err := ops.GetModuleName()
		require.NoError(t, err)
		assert.Equal(t, "test/module", name)

		assert.Equal(t, "v1.0.0", ops.GetVersion())
		assert.Equal(t, "v1.0.0", ops.GetGoVulnCheckVersion())
	})
}

// TestResetGoOperations tests ResetGoOperations function
func TestResetGoOperations(t *testing.T) {
	// First set a mock
	mock := &MockGoOperations{
		ModuleName: "mock/module",
		Version:    "v2.0.0",
	}
	err := SetGoOperations(mock)
	require.NoError(t, err)

	// Verify mock is set
	ops := GetGoOperations()
	assert.Equal(t, mock, ops)

	// Reset
	ResetGoOperations()

	// Should return default implementation again
	ops = GetGoOperations()
	_, ok := ops.(DefaultGoOperations)
	assert.True(t, ok, "should return DefaultGoOperations after reset")
}

// TestGoOperationsIntegration tests the full workflow
func TestGoOperationsIntegration(t *testing.T) {
	// Reset to ensure clean state
	ResetGoOperations()
	t.Cleanup(func() {
		ResetGoOperations()
	})

	// Test 1: Default operations
	ops := GetGoOperations()
	require.NotNil(t, ops)

	// Test 2: Set mock
	mock := &MockGoOperations{
		ModuleName:         "integration/test",
		ModuleNameErr:      nil,
		Version:            "v3.0.0",
		GoVulnCheckVersion: "v0.2.0",
	}
	err := SetGoOperations(mock)
	require.NoError(t, err)

	// Test 3: Verify mock is used
	ops = GetGoOperations()
	name, nameErr := ops.GetModuleName()
	require.NoError(t, nameErr)
	assert.Equal(t, "integration/test", name)
	assert.Equal(t, "v3.0.0", ops.GetVersion())
	assert.Equal(t, "v0.2.0", ops.GetGoVulnCheckVersion())

	// Test 4: Reset and verify default
	ResetGoOperations()
	ops = GetGoOperations()
	_, ok := ops.(DefaultGoOperations)
	assert.True(t, ok)
}

// TestMockGoOperationsWithError tests mock with error handling
func TestMockGoOperationsWithError(t *testing.T) {
	ResetGoOperations()
	t.Cleanup(func() {
		ResetGoOperations()
	})

	mock := &MockGoOperations{
		ModuleName:    "",
		ModuleNameErr: assert.AnError,
	}

	err := SetGoOperations(mock)
	require.NoError(t, err)

	ops := GetGoOperations()
	_, err = ops.GetModuleName()
	require.ErrorIs(t, err, assert.AnError)
}

// TestGoOperationsStaticError tests the static error constant
func TestGoOperationsStaticError(t *testing.T) {
	require.Error(t, errGoOperationsNil)
	require.NotEmpty(t, errGoOperationsNil.Error())
	assert.Contains(t, errGoOperationsNil.Error(), "nil")
}
