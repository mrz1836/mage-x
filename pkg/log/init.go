package log

import (
	"os"
	"sync"
)

// initOnce ensures initialization happens only once
//
//nolint:gochecknoglobals // Required for singleton initialization
var initOnce sync.Once

// ensureInitialized ensures the loggers are initialized.
// This is called automatically by Default() and Structured() if needed.
func ensureInitialized() {
	initOnce.Do(func() {
		Initialize()
	})
}

// Initialize initializes or reinitializes the loggers with default settings.
// This is useful for testing or resetting state.
func Initialize() {
	cliAdapter := NewCLIAdapter()
	cliAdapter.SetOutput(os.Stdout)
	SetDefault(cliAdapter)

	structuredAdapter := NewStructuredAdapter()
	structuredAdapter.SetOutput(os.Stderr)
	SetStructured(structuredAdapter)

	// Set default level based on environment
	if os.Getenv("DEBUG") != "" || os.Getenv("VERBOSE") != "" {
		SetLevel(LevelDebug)
	}
}
