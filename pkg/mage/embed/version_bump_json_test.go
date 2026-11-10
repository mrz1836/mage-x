package embed

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegisterNamespaceCommandsLongDescription checks that a command's
// LongDesc becomes the detailed description magex -h <command> shows.
func TestRegisterNamespaceCommandsLongDescription(t *testing.T) {
	t.Parallel()

	mockReg := NewMockRegistry()
	commands := []CommandDef{
		{Method: "detailed", Desc: "Detailed command", LongDesc: "What the command does, in full."},
		{Method: "brief", Desc: "Brief command"},
	}
	bindings := map[string]MethodBinding{
		"detailed": {NoArgs: func() error { return nil }},
		"brief":    {NoArgs: func() error { return nil }},
	}

	registerNamespaceCommands(mockReg.Registry, "ns", "Cat", commands, bindings)

	registered := mockReg.GetRegisteredCommands()
	require.Contains(t, registered, "ns:detailed")
	require.Contains(t, registered, "ns:brief")
	assert.Equal(t, "What the command does, in full.", registered["ns:detailed"].LongDescription)
	assert.Empty(t, registered["ns:brief"].LongDescription)
}

// TestVersionBumpHelpDocumentsJSON checks that version:bump's help names json
// and every field of the JSON line it prints.
func TestVersionBumpHelpDocumentsJSON(t *testing.T) {
	t.Parallel()

	mockReg := NewMockRegistry()
	registerVersionCommands(mockReg.Registry)

	bump, exists := mockReg.GetRegisteredCommands()["version:bump"]
	require.True(t, exists, "version:bump should be registered")

	assert.Contains(t, bump.Description, "json")
	assert.Contains(t, bump.Usage, "[json]")
	assert.Contains(t, bump.Examples, "magex version:bump bump=patch branch=master push json")
	for _, field := range []string{
		"from", "to", "bump", "expect", "dryRun", "tag", "commit", "pushed",
		"startBranch", "branch", "warnings", "error",
	} {
		assert.Containsf(t, bump.LongDescription, field, "the help doesn't name the %s field", field)
	}
}
