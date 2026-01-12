package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRun_Version tests run with --version flag
func TestRun_Version(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	exitCode := run([]string{"magex", "--version"})

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, 0, exitCode, "Should exit with 0")
}

// TestRun_Help tests run with --help flag
func TestRun_Help(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	exitCode := run([]string{"magex", "--help"})

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, 0, exitCode, "Should exit with 0")
}

// TestRun_HelpShortFlag tests run with -h flag
func TestRun_HelpShortFlag(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	exitCode := run([]string{"magex", "-h"})

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, 0, exitCode, "Should exit with 0")
}

// TestRun_HelpCommand tests run with help command
func TestRun_HelpCommand(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	exitCode := run([]string{"magex", "help", "build"})

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, 0, exitCode, "Should exit with 0")
}

// TestRun_Init tests run with -init flag
func TestRun_Init(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)
	os.Stdout = w

	exitCode := run([]string{"magex", "-init"})

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, 0, exitCode, "Should exit with 0")
	_, statErr := os.Stat("magefile.go")
	assert.NoError(t, statErr, "magefile.go should be created")
}

// TestRun_InitAlreadyExists tests run with -init when magefile exists
func TestRun_InitAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create existing magefile
	require.NoError(t, os.WriteFile("magefile.go", []byte("package main"), 0o600))

	// Capture stderr
	oldStderr := os.Stderr
	r, w, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)
	os.Stderr = w

	exitCode := run([]string{"magex", "-init"})

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stderr = oldStderr
	_ = r.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, 1, exitCode, "Should exit with 1")
}

// TestRun_Clean tests run with -clean flag
func TestRun_Clean(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create cache directories
	require.NoError(t, os.Mkdir(".mage", 0o750))
	require.NoError(t, os.Mkdir(".mage-x", 0o750))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)
	os.Stdout = w

	exitCode := run([]string{"magex", "-clean"})

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, 0, exitCode, "Should exit with 0")
}

// TestRun_List tests run with -l flag
func TestRun_List(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	exitCode := run([]string{"magex", "-l"})

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, 0, exitCode, "Should exit with 0")
}

// TestRun_ListLong tests run with --list flag
func TestRun_ListLong(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	exitCode := run([]string{"magex", "--list"})

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, 0, exitCode, "Should exit with 0")
}

// TestRun_Namespace tests run with -n flag
func TestRun_Namespace(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	exitCode := run([]string{"magex", "-n"})

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, 0, exitCode, "Should exit with 0")
}

// TestRun_Search tests run with -search flag
func TestRun_Search(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	exitCode := run([]string{"magex", "-search", "build"})

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, 0, exitCode, "Should exit with 0")
}

// TestRun_Compile tests run with -compile flag
func TestRun_Compile(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)
	os.Stdout = w

	exitCode := run([]string{"magex", "-compile", "output.go"})

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, 0, exitCode, "Should exit with 0")
	_, statErr := os.Stat("output.go")
	assert.NoError(t, statErr, "output.go should be created")
}

// TestRun_NoCommand tests run with no command (shows banner and quick list)
func TestRun_NoCommand(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	exitCode := run([]string{"magex"})

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, 0, exitCode, "Should exit with 0")
}

// TestRun_UnknownCommand tests run with unknown command
func TestRun_UnknownCommand(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	exitCode := run([]string{"magex", "nonexistentcommand12345"})

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stderr = oldStderr
	_ = r.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, 1, exitCode, "Should exit with 1")
}

// TestRun_VerboseFlag tests run with -v flag
func TestRun_VerboseFlag(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
		_ = os.Unsetenv("MAGEX_VERBOSE")  //nolint:errcheck // test cleanup
		_ = os.Unsetenv("MAGE_X_VERBOSE") //nolint:errcheck // test cleanup
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))

	// Create magefile with a test command
	magefileContent := `//go:build mage

package main

func Test() error {
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)
	os.Stdout = w

	// Run with -v flag and list commands
	exitCode := run([]string{"magex", "-v", "-l"})

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, 0, exitCode, "Should exit with 0")
	// Verify environment variables were set
	assert.Equal(t, "true", os.Getenv("MAGEX_VERBOSE"))
	assert.Equal(t, "1", os.Getenv("MAGE_X_VERBOSE"))
}

// TestRun_DebugFlag tests run with -debug flag
func TestRun_DebugFlag(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
		_ = os.Unsetenv("MAGEX_DEBUG")  //nolint:errcheck // test cleanup
		_ = os.Unsetenv("MAGE_X_DEBUG") //nolint:errcheck // test cleanup
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)
	os.Stdout = w

	// Run with -debug flag and list commands
	exitCode := run([]string{"magex", "-debug", "-l"})

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, 0, exitCode, "Should exit with 0")
	// Verify environment variables were set
	assert.Equal(t, "true", os.Getenv("MAGEX_DEBUG"))
	assert.Equal(t, "1", os.Getenv("MAGE_X_DEBUG"))
}

// TestRun_ListWithNamespace tests run with -l and -n flags
func TestRun_ListWithNamespace(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	exitCode := run([]string{"magex", "-l", "-n"})

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, 0, exitCode, "Should exit with 0")
}

// TestTryCustomCommand_NoMagefile tests tryCustomCommand with no magefile
func TestTryCustomCommand_NoMagefile(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// No magefile exists
	exitCode, cmdErr := tryCustomCommand("test", []string{}, NewCommandDiscovery(nil))
	assert.Equal(t, 0, exitCode, "Should return 0 when no magefile")
	assert.NoError(t, cmdErr, "Should return no error when no magefile")
}

// TestTryCustomCommand_WithMagefile tests tryCustomCommand with a magefile
func TestTryCustomCommand_WithMagefile(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))

	// Create magefile
	magefileContent := `//go:build mage

package main

import "fmt"

func Test() error {
	fmt.Println("Test command executed")
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(magefileContent), 0o600))

	// Capture stdout
	oldStdout := os.Stdout
	r, w, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)
	os.Stdout = w

	exitCode, cmdErr := tryCustomCommand("test", []string{}, NewCommandDiscovery(nil))

	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	assert.Equal(t, 0, exitCode, "Should return 0 on success")
	assert.NoError(t, cmdErr, "Should return no error on success")
}
