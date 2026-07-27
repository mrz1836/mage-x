package mage

import (
	"testing"
)

// setCommandExists replaces the commandExists seam for the duration of the test
// and restores it afterwards. Tests use it to pin whether an external tool looks
// installed, instead of depending on what the host happens to have - and, more
// importantly, instead of running tools that would go out to the network.
func setCommandExists(t *testing.T, stub func(name string) bool) {
	t.Helper()

	original := commandExists
	commandExists = stub
	t.Cleanup(func() { commandExists = original })
}

// setCommandsMissing marks the named commands as absent, leaving every other
// lookup answered by the real implementation.
func setCommandsMissing(t *testing.T, names ...string) {
	t.Helper()

	missing := make(map[string]bool, len(names))
	for _, name := range names {
		missing[name] = true
	}

	original := commandExists
	setCommandExists(t, func(name string) bool {
		if missing[name] {
			return false
		}
		return original(name)
	})
}
