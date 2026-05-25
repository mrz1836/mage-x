//go:build integration
// +build integration

package integration

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"
	"time"
)

// repoRootForSignalTest resolves the mage-x repository root from the test's
// working directory. Cached at package level would be cleaner, but a per-call
// resolution keeps each test self-contained.
func repoRootForSignalTest(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	// tests/integration → ../..
	return filepath.Clean(filepath.Join(cwd, "..", ".."))
}

// buildMagexForSignalTest compiles the magex binary into a temp dir for use
// in the signal tests. Skips on Windows (POSIX signal semantics required).
func buildMagexForSignalTest(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("Unix-only: relies on POSIX SIGINT/SIGTERM semantics")
	}

	dir := t.TempDir()
	bin := filepath.Join(dir, "magex")

	build := exec.Command("go", "build", "-o", bin, "./cmd/magex") //nolint:gosec // test paths
	build.Dir = repoRootForSignalTest(t)
	build.Env = append(os.Environ(), "GOTOOLCHAIN=local")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build magex: %v\n%s", err, out)
	}
	return bin
}

// newMagexCmd builds an *exec.Cmd that runs magex from the repo root, so the
// `lint` command has real Go modules to chew on and stays alive long enough
// for the test to send a signal.
func newMagexCmd(t *testing.T, bin string, args ...string) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(bin, args...) //nolint:gosec // test-controlled
	cmd.Dir = repoRootForSignalTest(t)
	return cmd
}

// waitOrTimeout runs cmd.Wait in a goroutine and returns when it completes
// or when the deadline elapses. Returns (exitCode, didTimeOut).
func waitOrTimeout(t *testing.T, cmd *exec.Cmd, deadline time.Duration) (int, bool) {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		if err == nil {
			return 0, false
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode(), false
		}
		t.Fatalf("unexpected Wait error: %v", err)
		return -1, false
	case <-time.After(deadline):
		// Force-kill so the test process doesn't leak the child.
		_ = cmd.Process.Kill()
		<-done
		return -1, true
	}
}

// TestMagex_SIGINTCleanExit_ExitCode130 proves end-to-end:
//   - magex lint runs and spawns subprocesses
//   - SIGINT cancels the whole process tree
//   - magex exits with code 130 (the conventional SIGINT shell code)
//   - the cancel banner appears in stderr exactly once
//
// This is the regression test that locks in the user's bug fix and runs in
// CI on every commit.
func TestMagex_SIGINTCleanExit_ExitCode130(t *testing.T) {
	bin := buildMagexForSignalTest(t)

	cmd := newMagexCmd(t, bin, "lint")
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer func() { _ = stderrR.Close() }()
	cmd.Stderr = stderrW
	cmd.Stdout = os.Stdout

	if err := cmd.Start(); err != nil {
		t.Fatalf("start magex: %v", err)
	}
	_ = stderrW.Close() // child holds the write end; parent only reads

	// Give magex time to enter the lint loop and spawn golangci-lint.
	time.Sleep(2 * time.Second)

	if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
		t.Fatalf("send SIGINT: %v", err)
	}

	// Drain stderr concurrently so the child's pipe doesn't block.
	stderrCh := make(chan []byte, 1)
	go func() {
		buf := make([]byte, 0, 4096)
		tmp := make([]byte, 1024)
		for {
			n, rerr := stderrR.Read(tmp)
			if n > 0 {
				buf = append(buf, tmp[:n]...)
			}
			if rerr != nil {
				stderrCh <- buf
				return
			}
		}
	}()

	// Generous bound: signal handler + WaitDelay (5s) + slack for CI.
	exitCode, timedOut := waitOrTimeout(t, cmd, 20*time.Second)
	if timedOut {
		t.Fatal("magex did not exit within 20s of SIGINT; cancellation broken")
	}

	if exitCode != 130 {
		t.Errorf("magex exit code = %d after SIGINT, want 130", exitCode)
	}

	stderr := string(<-stderrCh)
	if !strings.Contains(stderr, "Canceled by user") {
		t.Errorf("stderr missing cancel banner; got:\n%s", stderr)
	}
	if got := strings.Count(stderr, "Canceled by user"); got != 1 {
		t.Errorf("cancel banner printed %d times, want 1", got)
	}
}

// TestMagex_SIGTERMCleanExit_ExitCode143 proves the CI path: SIGTERM (which
// CI runners like GitHub Actions send when cancelling a job) is handled and
// produces exit code 143.
func TestMagex_SIGTERMCleanExit_ExitCode143(t *testing.T) {
	bin := buildMagexForSignalTest(t)

	cmd := newMagexCmd(t, bin, "lint")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("start magex: %v", err)
	}

	time.Sleep(2 * time.Second)

	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		t.Fatalf("send SIGTERM: %v", err)
	}

	exitCode, timedOut := waitOrTimeout(t, cmd, 20*time.Second)
	if timedOut {
		t.Fatal("magex did not exit within 20s of SIGTERM")
	}
	if exitCode != 143 {
		t.Errorf("magex exit code = %d after SIGTERM, want 143", exitCode)
	}
}

// TestMagex_DoubleSIGINT_ForceExit proves the second-signal escape hatch:
// if the first SIGINT's graceful path is somehow stuck, a second SIGINT
// triggers an immediate exit (130). User can never get "stuck" pressing Ctrl+C.
func TestMagex_DoubleSIGINT_ForceExit(t *testing.T) {
	bin := buildMagexForSignalTest(t)

	cmd := newMagexCmd(t, bin, "lint")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		t.Fatalf("start magex: %v", err)
	}

	time.Sleep(2 * time.Second)

	_ = cmd.Process.Signal(syscall.SIGINT)
	// Brief gap so the handler observes signal #1 before #2 arrives,
	// otherwise the buffered channel may swallow them as a pair on the
	// first read.
	time.Sleep(50 * time.Millisecond)
	_ = cmd.Process.Signal(syscall.SIGINT)

	exitCode, timedOut := waitOrTimeout(t, cmd, 10*time.Second)
	if timedOut {
		t.Fatal("magex did not exit after double SIGINT; force-exit broken")
	}
	if exitCode != 130 {
		t.Errorf("magex exit code = %d after double SIGINT, want 130", exitCode)
	}
}
