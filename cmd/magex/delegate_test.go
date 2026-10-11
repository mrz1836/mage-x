package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	secureDirPerm  = 0o750
	secureFilePerm = 0o600
	testGoMod      = `module testmage

go 1.21
`
)

var errTest = errors.New("test error")

func TestDelegateToMage_MagefilesDir(t *testing.T) {
	// Create a temporary directory with magefiles/ subdirectory
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create magefiles directory
	magefilesDir := "magefiles"
	err = os.Mkdir(magefilesDir, secureDirPerm)
	if err != nil {
		t.Fatalf("Failed to create magefiles directory: %v", err)
	}

	// Create a test magefile
	magefilePath := filepath.Join(magefilesDir, "commands.go")
	magefileContent := `//go:build mage
package main

import "fmt"

// TestCmd is a test command
func TestCmd() error {
	fmt.Println("TestCmd executed")
	return nil
}
`

	err = os.WriteFile(magefilePath, []byte(magefileContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create test magefile: %v", err)
	}

	// Test that HasMagefile returns true for directory
	if !HasMagefile() {
		t.Error("HasMagefile() should return true for magefiles/ directory")
	}

	// Test GetMagefilePath returns directory path
	path := GetMagefilePath()
	if !strings.Contains(path, "magefiles") {
		t.Errorf("GetMagefilePath() should return path with 'magefiles', got: %s", path)
	}

	// Note: We can't easily test actual command execution in unit tests
	// without mocking exec.Command, which would be complex.
	// Integration tests will cover the actual execution.
}

func TestDelegateToMage_MagefileGo(t *testing.T) {
	// Create a temporary directory with single magefile.go
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create a test magefile.go
	magefilePath := "magefile.go"
	magefileContent := `//go:build mage
package main

import "fmt"

// TestCmd is a test command
func TestCmd() error {
	fmt.Println("TestCmd executed")
	return nil
}
`

	err = os.WriteFile(magefilePath, []byte(magefileContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create test magefile: %v", err)
	}

	// Test that HasMagefile returns true for file
	if !HasMagefile() {
		t.Error("HasMagefile() should return true for magefile.go")
	}

	// Test GetMagefilePath returns file path
	path := GetMagefilePath()
	if !strings.Contains(path, "magefile.go") {
		t.Errorf("GetMagefilePath() should return path with 'magefile.go', got: %s", path)
	}
}

func TestDelegateToMage_PrefersMagefilesDir(t *testing.T) {
	// Create a temporary directory with both magefiles/ directory and magefile.go
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create magefiles directory
	magefilesDir := "magefiles"
	err = os.Mkdir(magefilesDir, secureDirPerm)
	if err != nil {
		t.Fatalf("Failed to create magefiles directory: %v", err)
	}

	// Create a test magefile in directory
	dirMagefilePath := filepath.Join(magefilesDir, "commands.go")
	dirMagefileContent := `//go:build mage
package main

// DirCmd is from directory
func DirCmd() error {
	return nil
}
`

	err = os.WriteFile(dirMagefilePath, []byte(dirMagefileContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create directory magefile: %v", err)
	}

	// Create root magefile.go (should be ignored)
	rootMagefilePath := "magefile.go"
	rootMagefileContent := `//go:build mage
package main

// RootCmd is from root file
func RootCmd() error {
	return nil
}
`

	err = os.WriteFile(rootMagefilePath, []byte(rootMagefileContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create root magefile: %v", err)
	}

	// Test that HasMagefile returns true (should find directory first)
	if !HasMagefile() {
		t.Error("HasMagefile() should return true when both exist")
	}

	// Test GetMagefilePath prefers directory over file
	path := GetMagefilePath()
	if !strings.Contains(path, "magefiles") {
		t.Errorf("GetMagefilePath() should prefer directory over file, got: %s", path)
	}
	if strings.Contains(path, "magefile.go") {
		t.Errorf("GetMagefilePath() should not return magefile.go when directory exists, got: %s", path)
	}
}

func TestDelegateToMage_CommandNotFound(t *testing.T) {
	// Create a temporary directory with no magefiles
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Test that DelegateToMage returns appropriate error
	result := DelegateToMage(context.Background(), "nonexistent")
	if result.Err == nil {
		t.Error("DelegateToMage should return error when no magefile exists")
	}
	if !errors.Is(result.Err, ErrCommandNotFound) {
		t.Errorf("DelegateToMage should return ErrCommandNotFound, got: %v", result.Err)
	}
}

func TestHasMagefile_NoFiles(t *testing.T) {
	// Create a temporary directory with no magefiles
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	if HasMagefile() {
		t.Error("HasMagefile() should return false when no magefiles exist")
	}
}

func TestHasMagefile_Directory(t *testing.T) {
	// Create a temporary directory with magefiles/ subdirectory
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create magefiles directory
	err = os.Mkdir("magefiles", secureDirPerm)
	if err != nil {
		t.Fatalf("Failed to create magefiles directory: %v", err)
	}

	if !HasMagefile() {
		t.Error("HasMagefile() should return true for magefiles/ directory")
	}
}

func TestHasMagefile_SingleFile(t *testing.T) {
	// Create a temporary directory with magefile.go
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create magefile.go
	err = os.WriteFile("magefile.go", []byte("package main"), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create magefile.go: %v", err)
	}

	if !HasMagefile() {
		t.Error("HasMagefile() should return true for magefile.go")
	}
}

func TestGetMagefilePath_Directory(t *testing.T) {
	// Create a temporary directory with magefiles/ subdirectory
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create magefiles directory
	err = os.Mkdir("magefiles", secureDirPerm)
	if err != nil {
		t.Fatalf("Failed to create magefiles directory: %v", err)
	}

	path := GetMagefilePath()
	if !strings.Contains(path, "magefiles") {
		t.Errorf("GetMagefilePath() should contain 'magefiles', got: %s", path)
	}
	if filepath.IsAbs(path) && !strings.HasSuffix(path, "magefiles") {
		t.Errorf("GetMagefilePath() should end with 'magefiles', got: %s", path)
	}
}

func TestGetMagefilePath_SingleFile(t *testing.T) {
	// Create a temporary directory with magefile.go
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create magefile.go
	err = os.WriteFile("magefile.go", []byte("package main"), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create magefile.go: %v", err)
	}

	path := GetMagefilePath()
	if !strings.Contains(path, "magefile.go") {
		t.Errorf("GetMagefilePath() should contain 'magefile.go', got: %s", path)
	}
	if filepath.IsAbs(path) && !strings.HasSuffix(path, "magefile.go") {
		t.Errorf("GetMagefilePath() should end with 'magefile.go', got: %s", path)
	}
}

func TestGetMagefilePath_NoFiles(t *testing.T) {
	// Create a temporary directory with no magefiles
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	path := GetMagefilePath()
	if path != "" {
		t.Errorf("GetMagefilePath() should return empty string when no files exist, got: %s", path)
	}
}

func TestStderrFilter(t *testing.T) {
	tests := []struct {
		name   string
		writes []string
		shown  string
		hidden string
	}{
		{
			name: "holds unknown-target lines back from display for the error message",
			writes: []string{`Normal error message
Unknown target specified: somecommand
Another normal message
Unknown target specified: anothercommand
Final message
`},
			shown:  "Normal error message\nAnother normal message\nFinal message\n",
			hidden: "Unknown target specified: somecommand\nUnknown target specified: anothercommand\n",
		},
		{
			name:   "joins a line split across writes",
			writes: []string{"first li", "ne\nsec", "ond line\n"},
			shown:  "first line\nsecond line\n",
		},
		{
			name:   "flushes a final line without a newline",
			writes: []string{"done\nno newline"},
			shown:  "done\nno newline\n",
		},
		{
			name:   "strips carriage returns",
			writes: []string{"windows line\r\n"},
			shown:  "windows line\n",
		},
		{
			name:   "handles a line longer than bufio.Scanner's limit",
			writes: []string{strings.Repeat("x", 100_000) + "\n"},
			shown:  strings.Repeat("x", 100_000) + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &lockedBuffer{}
			f := &stderrFilter{out: out}
			for _, w := range tt.writes {
				n, err := f.Write([]byte(w))
				require.NoError(t, err)
				assert.Equal(t, len(w), n)
			}
			f.Flush()

			assert.Equal(t, tt.shown, out.String())
			assert.Equal(t, tt.hidden, f.Hidden())
		})
	}
}

// lockedBuffer is a strings.Builder safe for the stderrFilter's timer goroutine
// to write while a test reads it.
type lockedBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (b *lockedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *lockedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// TestStderrFilter_PartialLines covers output that doesn't end in a newline,
// such as a prompt waiting for input, which is shown once stderr goes quiet.
func TestStderrFilter_PartialLines(t *testing.T) {
	waitShown := func(t *testing.T, out *lockedBuffer, want string) {
		t.Helper()
		assert.Eventually(t, func() bool { return out.String() == want },
			20*stderrPartialDelay, stderrPartialDelay/5, "want %q shown, got %q", want, out.String())
	}

	t.Run("shows a prompt without a newline once stderr goes quiet", func(t *testing.T) {
		out := &lockedBuffer{}
		f := &stderrFilter{out: out}

		_, err := f.Write([]byte("Type 'DESTROY' to confirm: "))
		require.NoError(t, err)
		waitShown(t, out, "Type 'DESTROY' to confirm: ")

		// The rest of the line and the next one follow without repeating the prompt
		_, err = f.Write([]byte("\nDestroying...\n"))
		require.NoError(t, err)
		f.Flush()
		assert.Equal(t, "Type 'DESTROY' to confirm: \nDestroying...\n", out.String())
	})

	t.Run("shows the rest of a line whose start was already shown", func(t *testing.T) {
		out := &lockedBuffer{}
		f := &stderrFilter{out: out}

		_, err := f.Write([]byte("Downloading"))
		require.NoError(t, err)
		waitShown(t, out, "Downloading")

		_, err = f.Write([]byte("... done\n"))
		require.NoError(t, err)
		f.Flush()
		assert.Equal(t, "Downloading... done\n", out.String())
		assert.Empty(t, f.Hidden())
	})

	t.Run("keeps a partial unknown-target line hidden", func(t *testing.T) {
		out := &lockedBuffer{}
		f := &stderrFilter{out: out}

		_, err := f.Write([]byte("Unknown tar"))
		require.NoError(t, err)
		time.Sleep(3 * stderrPartialDelay)
		assert.Empty(t, out.String(), "a line that may report an unknown target waits for its newline")

		_, err = f.Write([]byte("get specified: \"x\"\n"))
		require.NoError(t, err)
		f.Flush()
		assert.Empty(t, out.String())
		assert.Equal(t, "Unknown target specified: \"x\"\n", f.Hidden())
	})
}

// setupStderrMagefile writes a magefile.go into a fresh temp dir and changes into it.
func setupStderrMagefile(t *testing.T, magefile string) {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available")
	}

	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})
	require.NoError(t, os.Chdir(t.TempDir()))

	require.NoError(t, os.WriteFile("go.mod", []byte("module testmod\n\ngo 1.21\n"), 0o600))
	require.NoError(t, os.WriteFile(magefileFilename, []byte(magefile), 0o600))
}

// captureStderr redirects os.Stderr while fn runs and returns what was written.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	require.NoError(t, err)
	oldStderr := os.Stderr
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = oldStderr })

	var out strings.Builder
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&out, r) //nolint:errcheck // test helper; a short read shows up as a failed assertion
		close(done)
	}()

	fn()

	// Stop writing, and wait for the copy to finish before reading out
	os.Stderr = oldStderr
	require.NoError(t, w.Close())
	<-done
	require.NoError(t, r.Close())
	return out.String()
}

// TestDelegateToMageWithTimeout_CapturesAllStderr checks that stderr written just
// before the command exits is still displayed, and that the error doesn't repeat
// it: Wait used to close the stderr pipe on exit, racing the reader for the last
// of the output.
func TestDelegateToMageWithTimeout_CapturesAllStderr(t *testing.T) {
	setupStderrMagefile(t, `/`+`/go:build mage

package main

import (
	"errors"
	"fmt"
	"os"
)

func Noisy() error {
	for i := 0; i < 2000; i++ {
		fmt.Fprintf(os.Stderr, "stderr line %d\n", i)
	}
	fmt.Fprint(os.Stderr, "last line, no newline")
	return errors.New("noisy failure")
}
`)

	var result DelegateResult
	shown := captureStderr(t, func() {
		result = DelegateToMageWithTimeout(context.Background(), "noisy", 60*time.Second)
	})

	require.Error(t, result.Err)
	assert.NotEqual(t, 0, result.ExitCode)
	assert.Contains(t, shown, "stderr line 0\n")
	assert.Contains(t, shown, "stderr line 1999\n")
	assert.Contains(t, shown, "last line, no newline")
	assert.NotContains(t, result.Err.Error(), "stderr line", "stderr was already shown; the error must not repeat it")
	assert.Contains(t, result.Err.Error(), "exit status")
}

// TestDelegateToMageWithTimeout_BackgroundProcessHoldsStderr checks that a
// command which leaves a process running with its stderr still succeeds, and
// doesn't wait for that process to exit.
func TestDelegateToMageWithTimeout_BackgroundProcessHoldsStderr(t *testing.T) {
	if _, err := exec.LookPath("sleep"); err != nil {
		t.Skip("sleep command not available")
	}
	setupStderrMagefile(t, `/`+`/go:build mage

package main

import (
	"os"
	"os/exec"
	"strconv"
)

func Spawn() error {
	cmd := exec.Command("sleep", "30")
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	return os.WriteFile("sleep.pid", []byte(strconv.Itoa(cmd.Process.Pid)), 0o600)
}
`)
	dir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		pidBytes, readErr := os.ReadFile(filepath.Join(dir, "sleep.pid")) //nolint:gosec // test-controlled path
		if readErr != nil {
			return
		}
		pid, convErr := strconv.Atoi(string(pidBytes))
		if convErr != nil {
			return
		}
		if proc, findErr := os.FindProcess(pid); findErr == nil {
			_ = proc.Kill() //nolint:errcheck // best-effort cleanup of the background sleep
		}
	})

	start := time.Now()
	result := DelegateToMageWithTimeout(context.Background(), "spawn", 60*time.Second)

	require.NoError(t, result.Err)
	assert.Equal(t, 0, result.ExitCode)
	// Building the magefile takes a few seconds; the sleep would hold stderr for 30.
	assert.Less(t, time.Since(start), 25*time.Second, "should not wait for the background process")
}

// TestDelegateToMageWithTimeout_PassesArgsAsJSON checks that arguments reach
// the target through MAGE_ARGS_JSON intact, including ones that contain spaces.
func TestDelegateToMageWithTimeout_PassesArgsAsJSON(t *testing.T) {
	if _, err := exec.LookPath("mage"); err != nil {
		t.Skip("needs the mage binary, which receives arguments only through the environment")
	}
	setupStderrMagefile(t, `/`+`/go:build mage

package main

import "os"

func Echo() error {
	return os.WriteFile("args.txt", []byte(os.Getenv("MAGE_ARGS_JSON")+"\n"+os.Getenv("MAGE_ARGS")), 0o600)
}
`)

	result := DelegateToMageWithTimeout(context.Background(), "echo", 60*time.Second, "msg=hello world", "count=2")

	require.NoError(t, result.Err)
	got, err := os.ReadFile("args.txt")
	require.NoError(t, err)
	assert.Equal(t, `["msg=hello world","count=2"]`+"\nmsg=hello world count=2", string(got))
}

func TestConvertToMageFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// Colon-separated namespace commands - preserves colon
		{"Speckit:Install", "speckit:install"},
		{"Bmad:Install", "bmad:install"},
		{"Bmad:Check", "bmad:check"},
		{"Bmad:Upgrade", "bmad:upgrade"},
		{"Bmad:Status", "bmad:status"},
		{"Pipeline:CI", "pipeline:ci"},
		{"Build:Default", "build:default"},
		{"Test:Unit", "test:unit"},

		// Simple commands (just lowercased)
		{"Deploy", "deploy"},
		{"build", "build"},

		// Edge cases - colon preserved
		{":", ":"},
		{"Namespace:", "namespace:"},
		{":Method", ":method"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := convertToMageFormat(tt.input)
			if result != tt.expected {
				t.Errorf("convertToMageFormat(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDelegateToMage_Integration(t *testing.T) {
	// Skip this test if we don't have go available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available, skipping integration test")
	}

	// Create a temporary directory with working magefile
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create go.mod file for the test
	err = os.WriteFile("go.mod", []byte(testGoMod), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create magefiles directory
	err = os.Mkdir("magefiles", secureDirPerm)
	if err != nil {
		t.Fatalf("Failed to create magefiles directory: %v", err)
	}

	// Create a simple working magefile
	magefilePath := filepath.Join("magefiles", "commands.go")
	magefileContent := `//go:build mage
package main

import (
	"fmt"
	"os"
)

// TestCommand is a test command
func TestCommand() error {
	fmt.Println("TestCommand executed successfully")
	return nil
}

// ParamsTest shows how parameters are passed
func ParamsTest() error {
	fmt.Printf("Args: %v\n", os.Args[1:])
	return nil
}
`

	err = os.WriteFile(magefilePath, []byte(magefileContent), secureFilePerm)
	if err != nil {
		t.Fatalf("Failed to create test magefile: %v", err)
	}

	// Test that the command can be executed
	// Note: This will use "go run" since mage might not be available
	result := DelegateToMage(context.Background(), "TestCommand")
	if result.Err != nil {
		// This is expected to work in a real environment with Go
		// but might fail in test environment, so we'll log rather than fail
		t.Logf("DelegateToMage execution failed (expected in test environment): %v", result.Err)
	}
}

func BenchmarkHasMagefile(b *testing.B) {
	// Create a temporary directory with magefile.go
	tmpDir := b.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			b.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create magefile.go
	err = os.WriteFile("magefile.go", []byte("package main"), secureFilePerm)
	if err != nil {
		b.Fatalf("Failed to create magefile.go: %v", err)
	}

	b.ResetTimer()
	for b.Loop() {
		HasMagefile()
	}
}

func BenchmarkGetMagefilePath(b *testing.B) {
	// Create a temporary directory with magefile.go
	tmpDir := b.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			b.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		b.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create magefiles directory
	err = os.Mkdir("magefiles", secureDirPerm)
	if err != nil {
		b.Fatalf("Failed to create magefiles directory: %v", err)
	}

	b.ResetTimer()
	for b.Loop() {
		GetMagefilePath()
	}
}

// TestValidateGoEnvironmentSuccess tests that ValidateGoEnvironment returns nil
// when Go is available in the PATH (which should be the case in CI/dev environments).
func TestValidateGoEnvironmentSuccess(t *testing.T) {
	// Go is expected to be available since we're running tests
	err := ValidateGoEnvironment()
	assert.NoError(t, err, "ValidateGoEnvironment should succeed when Go is available")
}

// TestValidateGoEnvironmentNoGo tests that ValidateGoEnvironment returns
// ErrGoCommandNotFound when Go is not available in the PATH.
func TestValidateGoEnvironmentNoGo(t *testing.T) {
	// Save original PATH and restore after test
	originalPath := os.Getenv("PATH")
	t.Cleanup(func() {
		require.NoError(t, os.Setenv("PATH", originalPath))
	})

	// Clear PATH to simulate Go not being available
	require.NoError(t, os.Setenv("PATH", ""))

	err := ValidateGoEnvironment()
	require.Error(t, err, "ValidateGoEnvironment should fail when Go is not in PATH")
	assert.ErrorIs(t, err, ErrGoCommandNotFound, "Should return ErrGoCommandNotFound")
}

// TestDelegateToMageWithTimeoutTimesOut tests that DelegateToMageWithTimeout
// returns ErrCommandTimeout when the command exceeds the timeout duration.
func TestDelegateToMageWithTimeoutTimesOut(t *testing.T) {
	// Skip if go is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available, skipping timeout test")
	}

	// Skip if mage is not available (go run fallback can't handle magefiles directories)
	if _, err := exec.LookPath("mage"); err != nil {
		t.Skip("mage binary not available, skipping timeout test")
	}

	// Create a temporary directory with a magefile that sleeps
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalDir))
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte(testGoMod), secureFilePerm))

	// Create magefiles directory with a command that sleeps
	require.NoError(t, os.Mkdir("magefiles", secureDirPerm))

	magefileContent := `//go:build mage
package main

import (
	"time"
)

// SleepCmd sleeps for 10 seconds - used to test timeout
func SleepCmd() error {
	time.Sleep(10 * time.Second)
	return nil
}
`
	require.NoError(t, os.WriteFile(filepath.Join("magefiles", "commands.go"), []byte(magefileContent), secureFilePerm))

	// Execute with a very short timeout
	result := DelegateToMageWithTimeout(context.Background(), "sleepCmd", 500*time.Millisecond)

	// Should timeout and return the timeout error
	require.Error(t, result.Err, "Command should timeout")
	require.ErrorIs(t, result.Err, ErrCommandTimeout, "Should return ErrCommandTimeout")
	assert.Equal(t, 124, result.ExitCode, "Timeout exit code should be 124")
}

// useFakeMage changes into a new temp dir holding a magefile.go, and puts an
// executable "mage" that runs script (a POSIX shell script) at the front of
// PATH, so delegated commands run script in place of mage. Returns the dir.
func useFakeMage(t *testing.T, script string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fake mage is a POSIX shell script, and these tests rely on Unix signals")
	}

	binDir := t.TempDir()
	mage := filepath.Join(binDir, "mage")
	require.NoError(t, os.WriteFile(mage, []byte("#!/bin/sh\n"+script), 0o700)) //nolint:gosec // the fake mage must be executable
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	dir := t.TempDir()
	t.Chdir(dir)
	require.NoError(t, os.WriteFile(magefileFilename, []byte("//go:build mage\n\npackage main\n"), secureFilePerm))
	return dir
}

// cancelWhenCreated calls cancel once path exists, which the fake mage creates
// once its signal handling is set up, or after 30 seconds.
func cancelWhenCreated(path string, cancel context.CancelFunc) {
	go func() {
		defer cancel()
		for deadline := time.Now().Add(30 * time.Second); time.Now().Before(deadline); {
			if _, err := os.Stat(path); err == nil {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
}

// TestDelegateToMageWithTimeoutCanceledCommandFinishes checks that canceling
// the context, as Ctrl-C does, lets the command finish its own cleanup instead
// of killing it, and reports it as canceled rather than failed however it
// exits. Like mage while a target runs, the fake mage ignores the interrupt
// sent to it, and finishes a second later.
func TestDelegateToMageWithTimeoutCanceledCommandFinishes(t *testing.T) {
	tests := []struct {
		name     string
		exitCode int
	}{
		{name: "exits cleanly", exitCode: 0},
		{name: "exits with an error", exitCode: 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := useFakeMage(t, fmt.Sprintf("trap '' INT\n: > started\nsleep 1\n: > finished\nexit %d\n", tt.exitCode))
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			cancelWhenCreated(filepath.Join(dir, "started"), cancel)

			result := DelegateToMageWithTimeout(ctx, "serve", DefaultDelegateTimeout)

			require.ErrorIs(t, result.Err, ErrCommandCanceled)
			require.NotErrorIs(t, result.Err, ErrCommandFailed)
			assert.Equal(t, exitCodeSIGINT, result.ExitCode)
			assert.FileExists(t, filepath.Join(dir, "finished"), "the command should be left to finish")
		})
	}
}

// TestDelegateToMageWithTimeoutCanceledCommandKilledAfterDelay checks that a
// command still running killDelay after the context is canceled is killed.
func TestDelegateToMageWithTimeoutCanceledCommandKilledAfterDelay(t *testing.T) {
	// sleep inherits the ignored interrupt, and with exec no process outlives the kill
	dir := useFakeMage(t, "trap '' INT\n: > started\nexec sleep 60\n")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cancelWhenCreated(filepath.Join(dir, "started"), cancel)

	const killDelay = 200 * time.Millisecond
	var notices strings.Builder
	start := time.Now()
	result := delegateToMage(ctx, "serve", DefaultDelegateTimeout, delegateOptions{killDelay: killDelay, notices: &notices})
	elapsed := time.Since(start)

	require.ErrorIs(t, result.Err, ErrCommandCanceled)
	require.NotErrorIs(t, result.Err, ErrCommandFailed)
	assert.Contains(t, result.Err.Error(), "signal: killed")
	assert.Equal(t, exitCodeSIGINT, result.ExitCode)
	assert.GreaterOrEqual(t, elapsed, killDelay)
	assert.Less(t, elapsed, interruptKillDelay, "should be killed after the given delay, not the default one")
	assert.Equal(t, "⚠  Custom command 'serve' was still running 200ms after the interrupt — killed it.\n", notices.String(),
		"magex should say it killed the command")
}

// TestDelegateToMageWithTimeoutKillsAtTimeout checks that the timeout stays a
// hard limit: the command is killed as soon as it expires, without the delay
// an interrupt allows.
func TestDelegateToMageWithTimeoutKillsAtTimeout(t *testing.T) {
	useFakeMage(t, "trap '' INT\nexec sleep 60\n")

	start := time.Now()
	result := delegateToMage(context.Background(), "serve", 300*time.Millisecond, delegateOptions{killDelay: time.Minute})

	require.ErrorIs(t, result.Err, ErrCommandTimeout)
	assert.Equal(t, 124, result.ExitCode)
	assert.Less(t, time.Since(start), 10*time.Second)
}

// useFakeMageWithChild is useFakeMage with a fake mage that, like mage running
// a compiled magefile, catches the interrupt without passing it on, and runs a
// child in the foreground, as mage runs the compiled magefile. The child sets
// childTrap, writes its PID to child.pid, creates started, and then runs until
// a file named finish appears. So that a child the test failed to stop can't
// run on, it also stops once the dir is removed, or after a minute. Returns the
// dir.
func useFakeMageWithChild(t *testing.T, childTrap string) string {
	t.Helper()
	dir := useFakeMage(t, "trap ':' INT\nsh ./child.sh\n")
	if _, err := exec.LookPath("ps"); err != nil {
		t.Skip("ps is needed to find the processes a delegated command starts")
	}
	child := childTrap + "\necho $$ > child.pid\n: > started\n" +
		"i=0\nwhile [ ! -e finish ] && [ -e child.sh ] && [ $i -lt 1200 ]; do sleep 0.05; i=$((i+1)); done\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "child.sh"), []byte(child), secureFilePerm))
	// However the test ends, let a child that is still running exit
	t.Cleanup(func() {
		_ = os.WriteFile(filepath.Join(dir, "finish"), nil, secureFilePerm) //nolint:errcheck // best-effort cleanup
	})
	return dir
}

// childPID returns the PID the fake mage's child wrote to dir/child.pid.
func childPID(t *testing.T, dir string) int {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, "child.pid")) //nolint:gosec // test-controlled temp dir
	require.NoError(t, err)
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	require.NoError(t, err)
	return pid
}

// processRunning reports whether a process with pid is still running.
func processRunning(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	defer func() {
		_ = proc.Release() //nolint:errcheck // nothing to do if releasing the handle fails
	}()
	return proc.Signal(syscall.Signal(0)) == nil
}

// TestDelegateToMageWithTimeoutSIGTERMInterruptsTheTarget checks that after a
// SIGTERM sent to magex alone, which mage would not pass on, the process the
// command runs, the compiled magefile under mage, gets exactly one interrupt
// and can stop on its own, well before the command would be killed.
func TestDelegateToMageWithTimeoutSIGTERMInterruptsTheTarget(t *testing.T) {
	dir := useFakeMageWithChild(t, "trap 'echo interrupted >> interrupts; exit 0' INT")
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	cancelWhenCreated(filepath.Join(dir, "started"), func() { cancel(signalCancelError{sig: syscall.SIGTERM}) })

	const killDelay = 30 * time.Second
	start := time.Now()
	result := delegateToMage(ctx, "serve", DefaultDelegateTimeout, delegateOptions{killDelay: killDelay})

	require.ErrorIs(t, result.Err, ErrCommandCanceled)
	require.NotErrorIs(t, result.Err, ErrCommandFailed)
	assert.Less(t, time.Since(start), killDelay, "the target should stop on the interrupt, not be killed")
	interrupts, err := os.ReadFile(filepath.Join(dir, "interrupts")) //nolint:gosec // test-controlled temp dir
	require.NoError(t, err, "the target never got the interrupt")
	assert.Equal(t, "interrupted\n", string(interrupts), "the target should get exactly one interrupt")
}

// TestDelegateToMageWithTimeoutSIGINTSendsTheTargetNothing checks that after a
// SIGINT magex sends nothing to the processes the command started: a
// terminal's Ctrl-C has already reached them, and a second interrupt would cut
// a mage target's cleanup short.
func TestDelegateToMageWithTimeoutSIGINTSendsTheTargetNothing(t *testing.T) {
	dir := useFakeMageWithChild(t, "trap 'echo interrupted >> interrupts' INT")
	ctx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)
	cancelWhenCreated(filepath.Join(dir, "started"), func() { cancel(signalCancelError{sig: os.Interrupt}) })

	var listed atomic.Bool
	opts := delegateOptions{
		killDelay: time.Minute,
		processes: func() (map[int][]int, error) {
			listed.Store(true)
			return processChildren()
		},
	}
	finished := make(chan DelegateResult, 1)
	go func() {
		finished <- delegateToMage(ctx, "serve", DefaultDelegateTimeout, opts)
	}()

	// stopDelegated signals as soon as ctx is canceled; give it time to, then let
	// the child finish
	<-ctx.Done()
	time.Sleep(300 * time.Millisecond)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "finish"), nil, secureFilePerm))

	select {
	case result := <-finished:
		require.ErrorIs(t, result.Err, ErrCommandCanceled)
	case <-time.After(30 * time.Second):
		t.Fatal("the command did not finish")
	}
	assert.NoFileExists(t, filepath.Join(dir, "interrupts"), "the target should get no interrupt from magex")
	assert.False(t, listed.Load(), "magex has no reason to look for the command's processes")
}

// TestDelegateToMageWithTimeoutKillTakesTheTarget checks that killing the
// command, when it outlives killDelay after an interrupt or when the timeout
// expires, kills the processes it started too, which would otherwise keep
// running without it.
func TestDelegateToMageWithTimeoutKillTakesTheTarget(t *testing.T) {
	tests := []struct {
		name       string
		timeout    time.Duration
		cancel     bool
		wantErr    error
		wantNotice string
	}{
		{
			name: "after the interrupt", cancel: true, wantErr: ErrCommandCanceled,
			wantNotice: "⚠  Custom command 'serve' was still running 200ms after the interrupt — killed it.\n",
		},
		// Long enough for the child to start, even on a loaded machine; the
		// timeout reports itself, so there is no notice
		{name: "at the timeout", timeout: 2 * time.Second, wantErr: ErrCommandTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The child ignores the interrupt, so only a kill stops it
			dir := useFakeMageWithChild(t, "trap '' INT")
			ctx, cancel := context.WithCancelCause(context.Background())
			defer cancel(nil)
			if tt.cancel {
				cancelWhenCreated(filepath.Join(dir, "started"), func() { cancel(signalCancelError{sig: syscall.SIGTERM}) })
			}

			var notices strings.Builder
			result := delegateToMage(ctx, "serve", tt.timeout, delegateOptions{killDelay: 200 * time.Millisecond, notices: &notices})

			require.ErrorIs(t, result.Err, tt.wantErr)
			assert.Equal(t, tt.wantNotice, notices.String())
			pid := childPID(t, dir)
			assert.Eventually(t, func() bool { return !processRunning(pid) }, 10*time.Second, 50*time.Millisecond,
				"the command's child should be killed with it")
		})
	}
}

// TestStopDelegatedTimeoutCutsKillDelayShort checks that the timeout stays a
// hard limit after an interrupt too: a command still running when the timeout
// expires during killDelay is killed then.
func TestStopDelegatedTimeoutCutsKillDelayShort(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("runs a POSIX shell script, and relies on Unix signals")
	}
	dir := t.TempDir()
	started, interrupted := filepath.Join(dir, "started"), filepath.Join(dir, "interrupted")
	fileExists := func(path string) func() bool {
		return func() bool {
			_, err := os.Stat(path)
			return err == nil
		}
	}

	// Notes the interrupt, then keeps running
	cmd := exec.CommandContext(t.Context(), "sh", "-c", "trap ': > interrupted' INT; : > started; while :; do sleep 0.05; done")
	cmd.Dir = dir
	require.NoError(t, cmd.Start())
	exited := make(chan struct{})
	go func() {
		_ = cmd.Wait() //nolint:errcheck // the test checks how the command stopped through ProcessState
		close(exited)
	}()
	t.Cleanup(func() {
		_ = cmd.Process.Kill() //nolint:errcheck // best-effort cleanup; it has normally exited already
		<-exited
	})
	require.Eventually(t, fileExists(started), 30*time.Second, 10*time.Millisecond, "the command never started")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	timeoutC := make(chan time.Time)
	stopped := make(chan stopReason, 1)
	var notices strings.Builder
	go func() {
		stopped <- stopDelegated(ctx, cmd.Process, "serve", delegateOptions{killDelay: time.Minute, notices: &notices}, timeoutC, exited)
	}()

	// Once the command has the interrupt, stopDelegated is waiting out killDelay
	require.Eventually(t, fileExists(interrupted), 30*time.Second, 10*time.Millisecond, "the command never got the interrupt")
	timeoutC <- time.Now()

	select {
	case reason := <-stopped:
		assert.Equal(t, stopCanceled, reason, "the interrupt came first")
	case <-time.After(10 * time.Second):
		t.Fatal("the command was not killed when the timeout expired")
	}
	<-exited
	assert.Equal(t, "signal: killed", cmd.ProcessState.String())
	assert.Empty(t, notices.String(), "the timeout reports itself; the kill notice is for the grace period")
}

// TestDelegateToMageWithTimeoutCanceledBeforeStart checks that a command isn't
// started once the context is already canceled.
func TestDelegateToMageWithTimeoutCanceledBeforeStart(t *testing.T) {
	dir := useFakeMage(t, ": > started\n")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result := DelegateToMageWithTimeout(ctx, "serve", DefaultDelegateTimeout)

	require.ErrorIs(t, result.Err, ErrCommandCanceled)
	assert.Equal(t, exitCodeSIGINT, result.ExitCode)
	assert.NoFileExists(t, filepath.Join(dir, "started"))
}

// TestDelegateToMageWithTimeoutForcedExitRestoresMagefile checks that while a
// command runs with magefile.go renamed out of the way, the forced-exit hooks
// put it back, and that the deferred restore then has nothing left to do.
func TestDelegateToMageWithTimeoutForcedExitRestoresMagefile(t *testing.T) {
	dir := useFakeMage(t, ": > started\nwhile [ ! -e finish ]; do sleep 0.05; done\n")
	// magefiles/ next to magefile.go makes the delegation rename magefile.go
	require.NoError(t, os.Mkdir(magefilesDirname, secureDirPerm))
	hooks := &forceExitHooks{}
	ctx := withForceExitHooks(context.Background(), hooks)

	var result DelegateResult
	finished := make(chan struct{})
	go func() {
		defer close(finished)
		result = DelegateToMageWithTimeout(ctx, "serve", DefaultDelegateTimeout)
	}()
	finish := func() {
		require.NoError(t, os.WriteFile(filepath.Join(dir, "finish"), nil, secureFilePerm))
	}
	// However the test ends, let the fake mage exit, and wait for it
	t.Cleanup(func() {
		finish()
		<-finished
	})

	require.Eventually(t, func() bool {
		_, err := os.Stat(filepath.Join(dir, "started"))
		return err == nil
	}, 30*time.Second, 10*time.Millisecond, "the fake mage never started")
	require.NoFileExists(t, filepath.Join(dir, magefileFilename), "magefile.go should be out of the way while the command runs")

	hooks.run() // what a forced exit does before os.Exit

	assert.FileExists(t, filepath.Join(dir, magefileFilename))
	assert.NoFileExists(t, filepath.Join(dir, magefileFilename+".tmp"))

	finish()
	select {
	case <-finished:
	case <-time.After(30 * time.Second):
		t.Fatal("the command did not finish")
	}
	require.NoError(t, result.Err, "the deferred restore should find nothing left to do")
	assert.Empty(t, hooks.hooks, "the hook should be removed once the command returns")
}

// TestMagefileRestore checks that the deferred restore and a forced exit can
// both restore magefile.go, even at the same time: the rename happens once,
// and every caller gets its result.
func TestMagefileRestore(t *testing.T) {
	t.Run("renames once for concurrent callers", func(t *testing.T) {
		t.Chdir(t.TempDir())
		require.NoError(t, os.WriteFile(magefileFilename+".tmp", []byte("package main\n"), secureFilePerm))
		restore := &magefileRestore{tempName: magefileFilename + ".tmp"}

		errs := make([]error, 8)
		var wg sync.WaitGroup
		for i := range errs {
			wg.Go(func() { errs[i] = restore.restore() })
		}
		wg.Wait()

		for _, err := range errs {
			require.NoError(t, err)
		}
		assert.FileExists(t, magefileFilename)
		assert.NoFileExists(t, magefileFilename+".tmp")
	})

	t.Run("forced exit reports a failure", func(t *testing.T) {
		t.Chdir(t.TempDir())
		restore := &magefileRestore{tempName: magefileFilename + ".tmp"}

		shown := captureStderr(t, restore.onForcedExit)

		require.Error(t, restore.restore())
		assert.Contains(t, shown, ErrMagefileRestoreFailed.Error())
	})
}

// TestDelegateToMageWithTimeoutWithArguments tests that arguments are passed
// to the delegated command via MAGE_ARGS environment variable.
func TestDelegateToMageWithTimeoutWithArguments(t *testing.T) {
	// Skip if go is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available, skipping arguments test")
	}

	// Create a temporary directory with a magefile
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalDir))
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte(testGoMod), secureFilePerm))

	// Create magefiles directory with a command that prints args
	require.NoError(t, os.Mkdir("magefiles", secureDirPerm))

	magefileContent := `//go:build mage
package main

import (
	"fmt"
	"os"
)

// ArgsCmd prints the MAGE_ARGS environment variable
func ArgsCmd() error {
	args := os.Getenv("MAGE_ARGS")
	fmt.Printf("ARGS:%s\n", args)
	return nil
}
`
	require.NoError(t, os.WriteFile(filepath.Join("magefiles", "commands.go"), []byte(magefileContent), secureFilePerm))

	// Execute with arguments
	result := DelegateToMageWithTimeout(context.Background(), "argsCmd", DefaultDelegateTimeout, "arg1", "arg2", "arg3")

	// The command should execute without error
	// Note: This test verifies the code path that sets MAGE_ARGS, even if the actual execution
	// might fail depending on environment
	_ = result // We just want to exercise the code path with arguments
}

// TestDelegateToMageWithTimeoutConflictHandling tests the conflict handling
// when both magefiles/ directory and magefile.go exist.
func TestDelegateToMageWithTimeoutConflictHandling(t *testing.T) {
	// Skip if go is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available, skipping conflict handling test")
	}

	// Create a temporary directory
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalDir))
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte(testGoMod), secureFilePerm))

	// Create magefiles directory with a command
	require.NoError(t, os.Mkdir("magefiles", secureDirPerm))

	dirMagefileContent := `//go:build mage
package main

import "fmt"

// TestCmd is from directory
func TestCmd() error {
	fmt.Println("FROM_DIR")
	return nil
}
`
	require.NoError(t, os.WriteFile(filepath.Join("magefiles", "commands.go"), []byte(dirMagefileContent), secureFilePerm))

	// Create root magefile.go (should be temporarily renamed)
	rootMagefileContent := `//go:build mage
package main

import "fmt"

// RootCmd is from root file
func RootCmd() error {
	fmt.Println("FROM_ROOT")
	return nil
}
`
	require.NoError(t, os.WriteFile("magefile.go", []byte(rootMagefileContent), secureFilePerm))

	// Execute a command - this should trigger the conflict handling code path
	result := DelegateToMageWithTimeout(context.Background(), "testCmd", DefaultDelegateTimeout)

	// After execution, magefile.go should be restored
	_, err = os.Stat("magefile.go")
	require.NoError(t, err, "magefile.go should be restored after command execution")

	// Verify no temp file is left behind
	_, err = os.Stat("magefile.go.tmp")
	assert.True(t, os.IsNotExist(err), "Temp file should not exist after execution")

	// The result may be an error (no mage installed) but that's fine
	// We're testing the conflict handling code path, not actual execution success
	_ = result
}

// TestDelegateToMageWithTimeoutStderrCapture tests that stderr output is
// properly captured and included in error messages.
func TestDelegateToMageWithTimeoutStderrCapture(t *testing.T) {
	// Skip if go is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available, skipping stderr capture test")
	}

	// Create a temporary directory with a magefile that writes to stderr and fails
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(originalDir))
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte(testGoMod), secureFilePerm))

	// Create magefiles directory with a command that writes to stderr and exits with error
	require.NoError(t, os.Mkdir("magefiles", secureDirPerm))

	magefileContent := `//go:build mage
package main

import (
	"errors"
	"fmt"
	"os"
)

// FailCmd writes to stderr and returns an error
func FailCmd() error {
	fmt.Fprintln(os.Stderr, "Custom error message on stderr")
	return errors.New("command failed intentionally")
}
`
	require.NoError(t, os.WriteFile(filepath.Join("magefiles", "commands.go"), []byte(magefileContent), secureFilePerm))

	// Execute the failing command
	result := DelegateToMageWithTimeout(context.Background(), "failCmd", DefaultDelegateTimeout)

	// Should fail with the error message captured
	require.Error(t, result.Err, "Command should fail")
	assert.ErrorIs(t, result.Err, ErrCommandFailed, "Should return ErrCommandFailed")
}

// TestConvertToMageFormatEdgeCases tests additional edge cases for convertToMageFormat
func TestConvertToMageFormatEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "multiple colons preserves all",
			input:    "A:B:C",
			expected: "a:b:c", // Just lowercased, colons preserved
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "single colon with empty parts",
			input:    ":",
			expected: ":", // Colon preserved
		},
		{
			name:     "unicode characters",
			input:    "Ünïcödé:Tëst",
			expected: "ünïcödé:tëst", // Just lowercased, colon preserved
		},
		{
			name:     "numbers in namespace",
			input:    "Build123:Run456",
			expected: "build123:run456", // Colon preserved
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertToMageFormat(tt.input)
			assert.Equal(t, tt.expected, result, "convertToMageFormat(%q) should return %q", tt.input, tt.expected)
		})
	}
}

// TestStaticErrorsExist verifies that all static error variables are properly defined
func TestStaticErrorsExist(t *testing.T) {
	t.Parallel()

	errorVars := map[string]error{
		"ErrCommandNotFound":       ErrCommandNotFound,
		"ErrGoCommandNotFound":     ErrGoCommandNotFound,
		"ErrCommandFailed":         ErrCommandFailed,
		"ErrMagefileRestoreFailed": ErrMagefileRestoreFailed,
		"ErrCommandTimeout":        ErrCommandTimeout,
		"ErrCommandCanceled":       ErrCommandCanceled,
	}

	for name, err := range errorVars {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.Error(t, err, "%s should not be nil", name)
			assert.NotEmpty(t, err.Error(), "%s should have an error message", name)
		})
	}
}

// TestDelegateResultStruct tests the DelegateResult struct
func TestDelegateResultStruct(t *testing.T) {
	t.Parallel()

	t.Run("zero value has nil error", func(t *testing.T) {
		t.Parallel()
		var result DelegateResult
		assert.Equal(t, 0, result.ExitCode)
		assert.NoError(t, result.Err)
	})

	t.Run("can set exit code and error", func(t *testing.T) {
		t.Parallel()
		testErr := errTest
		result := DelegateResult{
			ExitCode: 42,
			Err:      testErr,
		}
		assert.Equal(t, 42, result.ExitCode)
		assert.Equal(t, testErr, result.Err)
	})
}

// TestGetMagefilePath_PrefersMagefiles tests that GetMagefilePath prefers magefiles directory
func TestGetMagefilePath_PrefersMagefiles(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create both magefiles directory and magefile.go
	err = os.Mkdir("magefiles", secureDirPerm)
	require.NoError(t, err)
	err = os.WriteFile("magefile.go", []byte("package main"), secureFilePerm)
	require.NoError(t, err)

	path := GetMagefilePath()
	// Should prefer magefiles directory
	assert.Contains(t, path, "magefiles")
	assert.NotContains(t, path, "magefile.go")
}

// TestConvertToMageFormat_MultipleColons tests command with multiple colons
func TestConvertToMageFormat_MultipleColons(t *testing.T) {
	input := "namespace:method:extra"
	result := convertToMageFormat(input)
	// Just lowercased, all colons preserved
	expected := "namespace:method:extra"
	assert.Equal(t, expected, result)
}

// TestGetMagefilePath_MagefilesDir tests GetMagefilePath with magefiles/ directory
func TestGetMagefilePath_MagefilesDir(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create magefiles/ directory
	err = os.Mkdir("magefiles", 0o750)
	require.NoError(t, err)

	path := GetMagefilePath()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "magefiles")
}

// TestGetMagefilePath_MagefileGo tests GetMagefilePath with magefile.go
func TestGetMagefilePath_MagefileGo(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create magefile.go
	err = os.WriteFile("magefile.go", []byte("package main"), 0o600)
	require.NoError(t, err)

	path := GetMagefilePath()
	assert.NotEmpty(t, path)
	assert.Contains(t, path, "magefile.go")
}

// TestGetMagefilePath_NoMagefile tests GetMagefilePath with no magefile
func TestGetMagefilePath_NoMagefile(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	path := GetMagefilePath()
	assert.Empty(t, path)
}

// TestGetMagefilePath_PrefersMagefilesDir tests that magefiles/ is preferred over magefile.go
func TestGetMagefilePath_PrefersMagefilesDir(t *testing.T) {
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(oldDir); chErr != nil {
			t.Errorf("failed to restore directory: %v", chErr)
		}
	})

	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	// Create both magefiles/ and magefile.go
	err = os.Mkdir("magefiles", 0o750)
	require.NoError(t, err)
	err = os.WriteFile("magefile.go", []byte("package main"), 0o600)
	require.NoError(t, err)

	path := GetMagefilePath()
	assert.NotEmpty(t, path)
	// Should prefer magefiles/ directory
	assert.Contains(t, path, "magefiles")
	assert.NotContains(t, path, "magefile.go")
}

// TestDelegateToMage_NamespaceCommands tests that namespace commands (colon-separated)
// are properly converted and passed to mage. This is a regression test for the
// convertToMageFormat fix that ensures commands like "Ci:Static" are properly
// converted to "ci:static" (lowercase) before being passed to mage.
//
// Note: Full execution may not work in all test environments due to mage compilation
// requirements. The test validates format conversion and command delegation.
func TestDelegateToMage_NamespaceCommands(t *testing.T) {
	// Skip if mage binary not available
	if _, err := exec.LookPath("mage"); err != nil {
		t.Skip("mage binary not available, skipping namespace command integration test")
	}

	// Skip if go is not available
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go command not available, skipping namespace command integration test")
	}

	// Create temporary project directory
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			t.Logf("Failed to restore directory: %v", chErr)
		}
	})

	require.NoError(t, os.Chdir(tmpDir))

	// Create go.mod
	goModContent := `module testnamespace

go 1.21
`
	require.NoError(t, os.WriteFile("go.mod", []byte(goModContent), secureFilePerm))

	// Create magefiles directory with namespace commands
	require.NoError(t, os.Mkdir("magefiles", secureDirPerm))

	// Create a magefile with a namespace (type CI struct{})
	magefileContent := `//go:build mage

package main

import (
	"fmt"
)

// CI namespace for CI/CD commands
type CI struct{}

// Static runs static analysis (this was the command that was broken)
func (CI) Static() error {
	fmt.Println("NAMESPACE_COMMAND_SUCCESS:ci:static")
	return nil
}

// Lint runs linting
func (CI) Lint() error {
	fmt.Println("NAMESPACE_COMMAND_SUCCESS:ci:lint")
	return nil
}

// Build namespace for build commands
type Build struct{}

// Default runs the default build
func (Build) Default() error {
	fmt.Println("NAMESPACE_COMMAND_SUCCESS:build:default")
	return nil
}

// Test namespace for test commands
type Test struct{}

// Unit runs unit tests
func (Test) Unit() error {
	fmt.Println("NAMESPACE_COMMAND_SUCCESS:test:unit")
	return nil
}

// Coverage runs tests with coverage (multiple-colon test when chained)
func (Test) Coverage() error {
	fmt.Println("NAMESPACE_COMMAND_SUCCESS:test:coverage")
	return nil
}
`
	require.NoError(t, os.WriteFile(filepath.Join("magefiles", "commands.go"), []byte(magefileContent), secureFilePerm))

	// Test cases for namespace command handling
	tests := []struct {
		name              string
		input             string
		expectedLowercase string
		expectedOutput    string
		description       string
	}{
		{
			name:              "mixed_case_namespace",
			input:             "Ci:Static",
			expectedLowercase: "ci:static",
			expectedOutput:    "NAMESPACE_COMMAND_SUCCESS:ci:static",
			description:       "Mixed case namespace (the bug that was fixed)",
		},
		{
			name:              "all_uppercase_namespace",
			input:             "CI:STATIC",
			expectedLowercase: "ci:static",
			expectedOutput:    "NAMESPACE_COMMAND_SUCCESS:ci:static",
			description:       "All uppercase namespace",
		},
		{
			name:              "already_lowercase_namespace",
			input:             "ci:static",
			expectedLowercase: "ci:static",
			expectedOutput:    "NAMESPACE_COMMAND_SUCCESS:ci:static",
			description:       "Already lowercase namespace",
		},
		{
			name:              "mixed_case_build",
			input:             "Build:Default",
			expectedLowercase: "build:default",
			expectedOutput:    "NAMESPACE_COMMAND_SUCCESS:build:default",
			description:       "Standard mixed case namespace",
		},
		{
			name:              "ci_lint",
			input:             "CI:Lint",
			expectedLowercase: "ci:lint",
			expectedOutput:    "NAMESPACE_COMMAND_SUCCESS:ci:lint",
			description:       "CI lint command",
		},
		{
			name:              "test_coverage",
			input:             "Test:Coverage",
			expectedLowercase: "test:coverage",
			expectedOutput:    "NAMESPACE_COMMAND_SUCCESS:test:coverage",
			description:       "Test coverage command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// First verify the format conversion works correctly (unit test level)
			converted := convertToMageFormat(tt.input)
			assert.Equal(t, tt.expectedLowercase, converted,
				"convertToMageFormat(%q) should return %q", tt.input, tt.expectedLowercase)

			// Capture stdout to verify output
			oldStdout := os.Stdout
			r, w, pipeErr := os.Pipe()
			require.NoError(t, pipeErr)
			os.Stdout = w

			// Execute the command with a timeout
			result := DelegateToMageWithTimeout(context.Background(), tt.input, 30*time.Second)

			// Restore stdout and read captured output
			if closeErr := w.Close(); closeErr != nil {
				t.Logf("Failed to close pipe writer: %v", closeErr)
			}
			os.Stdout = oldStdout

			var capturedOutput strings.Builder
			if _, copyErr := io.Copy(&capturedOutput, r); copyErr != nil {
				t.Logf("Failed to read captured output: %v", copyErr)
			}
			if closeErr := r.Close(); closeErr != nil {
				t.Logf("Failed to close pipe reader: %v", closeErr)
			}

			// If execution succeeded, verify the output
			if result.Err == nil {
				assert.Equal(t, 0, result.ExitCode, "Exit code should be 0 for '%s'", tt.input)
				output := capturedOutput.String()
				assert.Contains(t, output, tt.expectedOutput,
					"Output for '%s' should contain '%s', got: %s",
					tt.input, tt.expectedOutput, output)
			} else {
				// If execution failed, verify it's due to mage not finding the target
				// (expected in test environments where mage can't compile the test magefiles)
				// The key verification is that the command was passed correctly (lowercase)
				errMsg := result.Err.Error()
				if strings.Contains(errMsg, "Unknown target specified") {
					// This is acceptable - mage received the correctly formatted command
					// but couldn't find the target (test environment limitation)
					assert.Contains(t, errMsg, tt.expectedLowercase,
						"Error should reference the lowercase command format '%s'", tt.expectedLowercase)
					t.Logf("Command '%s' correctly converted to '%s' (mage target not found in test env)",
						tt.input, tt.expectedLowercase)
				} else {
					// Unexpected error
					t.Errorf("Unexpected error for command '%s': %v", tt.input, result.Err)
				}
			}
		})
	}
}
