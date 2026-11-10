package embed

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVersionBumpHelpDocumentsFFOnly checks that version:bump's help names
// ff-only in its description, usage, examples, and detailed help.
func TestVersionBumpHelpDocumentsFFOnly(t *testing.T) {
	t.Parallel()

	mockReg := NewMockRegistry()
	registerVersionCommands(mockReg.Registry)

	bump, exists := mockReg.GetRegisteredCommands()["version:bump"]
	require.True(t, exists, "version:bump should be registered")

	assert.Contains(t, bump.Description, "ff-only")
	assert.Contains(t, bump.Usage, "[ff-only]")
	assert.Contains(t, bump.Examples, "magex version:bump bump=patch branch=master push ff-only")
	assert.Contains(t, bump.LongDescription, "ff-only")
}
