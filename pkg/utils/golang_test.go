package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoList(t *testing.T) {
	t.Run("lists packages in current module", func(t *testing.T) {
		// This should work in the test environment
		packages, err := GoList("-m")
		require.NoError(t, err)
		assert.NotEmpty(t, packages)
	})

	t.Run("handles invalid arguments", func(t *testing.T) {
		// Invalid list target should return error
		_, err := GoList("-invalid-flag-12345")
		assert.Error(t, err)
	})
}

func TestGetModuleName(t *testing.T) {
	t.Run("returns current module name", func(t *testing.T) {
		moduleName, err := GetModuleName()
		require.NoError(t, err)
		assert.NotEmpty(t, moduleName)
		assert.Contains(t, moduleName, "mage-x")
	})
}

func TestGetGoVersion(t *testing.T) {
	t.Run("returns valid Go version", func(t *testing.T) {
		version, err := GetGoVersion()
		require.NoError(t, err)
		assert.NotEmpty(t, version)
		// Version should start with a number
		assert.Regexp(t, `^\d+\.\d+`, version)
	})
}
