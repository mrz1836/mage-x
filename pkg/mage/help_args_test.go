package mage_test

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/mage/embed"
	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

// ensureCommandsRegistered loads the full command catalog into the global
// registry exactly once. The package-level tests need this because they assert
// behavior of the help system, which reads from the global registry.
//

var ensureCommandsRegistered = sync.OnceFunc(func() {
	reg := registry.Global()
	if reg.IsRegistered() {
		return
	}
	embed.RegisterAll(reg)
	reg.SetRegistered(true)
})

// captureStdout redirects os.Stdout for the duration of fn and returns whatever
// fn wrote. Used to assert the JSON payload printed by help:command and
// help:commands without polluting test output.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	defer func() { os.Stdout = origStdout }()

	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r) //nolint:errcheck // Test helper
		close(done)
	}()

	fn()

	require.NoError(t, w.Close())
	<-done
	return buf.String()
}

// firstRegisteredCommand picks a deterministic, fully-qualified command name
// (e.g. "test:run") from the global registry so the help tests don't hard-code
// a name that might be renamed. Avoids namespace names by requiring a ':'.
func firstRegisteredCommand(t *testing.T) string {
	t.Helper()
	commands := registry.Global().List()
	require.NotEmpty(t, commands, "expected at least one registered command")
	for _, c := range commands {
		full := c.FullName()
		if strings.Contains(full, ":") {
			return full
		}
	}
	t.Fatal("no namespaced command (with ':') found in registry")
	return ""
}

// TestHelpCommand_ParamSyntax exercises the bug-fix path: `magex help:command
// command=<name>` (and the env-var fallback for back-compat) should both resolve
// the target command. Previously the param was silently dropped.
func TestHelpCommand_ParamSyntax(t *testing.T) {
	ensureCommandsRegistered()
	target := firstRegisteredCommand(t)

	t.Run("missing both env and param errors with usage hint", func(t *testing.T) {
		t.Setenv("COMMAND", "")
		err := mage.Help{}.Command()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "required")
		assert.Contains(t, err.Error(), "command=")
	})

	t.Run("param resolves command", func(t *testing.T) {
		t.Setenv("COMMAND", "")
		err := mage.Help{}.Command("command=" + target)
		require.NoError(t, err)
	})

	t.Run("env var still works (back-compat)", func(t *testing.T) {
		t.Setenv("COMMAND", target)
		err := mage.Help{}.Command()
		require.NoError(t, err)
	})

	t.Run("param takes precedence over env", func(t *testing.T) {
		t.Setenv("COMMAND", "nonexistent-command-xyz")
		err := mage.Help{}.Command("command=" + target)
		require.NoError(t, err)
	})

	t.Run("nonexistent command in param yields not-found error", func(t *testing.T) {
		t.Setenv("COMMAND", "")
		err := mage.Help{}.Command("command=this-command-does-not-exist-zzz")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "this-command-does-not-exist-zzz")
	})
}

// TestHelpCommand_JSONOutput verifies that `magex help:command command=<name>
// json=true` emits a parseable JSON document matching the HelpCommand schema.
// This is the contract agents rely on for command discovery.
func TestHelpCommand_JSONOutput(t *testing.T) {
	ensureCommandsRegistered()
	target := firstRegisteredCommand(t)
	t.Setenv("COMMAND", "")

	out := captureStdout(t, func() {
		err := mage.Help{}.Command("command="+target, "json=true")
		require.NoError(t, err)
	})

	out = strings.TrimSpace(out)
	require.NotEmpty(t, out)

	var got mage.HelpCommand
	require.NoError(t, json.Unmarshal([]byte(out), &got), "JSON must be parseable")
	assert.Equal(t, target, got.Name, "JSON payload must report the requested command")
}

// TestHelpCommand_JSONOutputForTestRun pins the test:run JSON contract since
// agents will key on this command's Options/Examples to figure out how to call it.
func TestHelpCommand_JSONOutputForTestRun(t *testing.T) {
	ensureCommandsRegistered()
	t.Setenv("COMMAND", "")

	out := captureStdout(t, func() {
		err := mage.Help{}.Command("command=test:run", "json=true")
		require.NoError(t, err)
	})

	var got mage.HelpCommand
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(out)), &got))

	assert.Equal(t, "test:run", got.Name)
	assert.Contains(t, got.Aliases, "test:specific", "test:specific must be registered as an alias of test:run")

	// Options must include both name and pkg so agents can discover the params.
	optNames := make(map[string]mage.HelpOption, len(got.Options))
	for _, o := range got.Options {
		optNames[o.Name] = o
	}
	require.Contains(t, optNames, "name")
	require.Contains(t, optNames, "pkg")
	assert.Equal(t, "./...", optNames["pkg"].Default)

	require.NotEmpty(t, got.Examples)
	assert.Contains(t, got.Usage, "name=")
	assert.Contains(t, got.Usage, "pkg=")
}

// TestHelpCommands_JSONOutput verifies the list endpoint returns a JSON catalog
// that parses cleanly and includes at least the test:run command.
func TestHelpCommands_JSONOutput(t *testing.T) {
	ensureCommandsRegistered()

	out := captureStdout(t, func() {
		err := mage.Help{}.Commands("json=true")
		require.NoError(t, err)
	})

	var got mage.HelpCommandList
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(out)), &got))

	assert.Equal(t, len(got.Commands), got.Total, "Total must match Commands length")
	assert.NotEmpty(t, got.Commands)

	seen := false
	for _, c := range got.Commands {
		if c.Name == "test:run" {
			seen = true
			break
		}
	}
	assert.True(t, seen, "test:run must appear in the catalog")
}

// TestHelpCommand_JSONForNamespace verifies that requesting a namespace name
// with json=true returns the full set of commands in that namespace.
func TestHelpCommand_JSONForNamespace(t *testing.T) {
	ensureCommandsRegistered()
	t.Setenv("COMMAND", "")

	out := captureStdout(t, func() {
		err := mage.Help{}.Command("command=test", "json=true")
		require.NoError(t, err)
	})

	var got mage.HelpCommandList
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(out)), &got))
	assert.NotEmpty(t, got.Commands, "namespace JSON must list child commands")
	for _, c := range got.Commands {
		assert.Equal(t, "test", c.Namespace)
	}
}

// TestHelpCommand_TextPathStillWorks confirms the non-JSON rendering path is
// untouched — important because the bug-fix changed the signature.
func TestHelpCommand_TextPathStillWorks(t *testing.T) {
	ensureCommandsRegistered()
	target := firstRegisteredCommand(t)
	t.Setenv("COMMAND", "")

	// Just make sure it doesn't error; we don't pin the exact text output.
	err := mage.Help{}.Command("command=" + target)
	require.NoError(t, err)
}
