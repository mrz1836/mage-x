package exec_test

import (
	"context"
	"errors"
	"io"
	"os"
	osexec "os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/mrz1836/mage-x/pkg/exec"
)

// trapSIGINTSource is a tiny Go program that traps SIGINT and exits with code
// 7. Compiled into a temp binary per test and invoked through the executor.
// If the executor sends SIGKILL on context cancel (the pre-fix behavior),
// the child cannot trap and exits with a SIGKILL exit code instead.
const trapSIGINTSource = `package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	// Print a ready marker so the test knows the child has installed its
	// signal handler before we cancel. Stdout is buffered, so flush via
	// File.Sync after writing.
	if _, err := os.Stdout.WriteString("READY\n"); err != nil {
		os.Exit(99)
	}
	_ = os.Stdout.Sync()

	select {
	case <-ch:
		os.Exit(7)
	case <-time.After(30 * time.Second):
		os.Exit(98) // safety net: never seen in passing tests
	}
}
`

// buildTrapBinary compiles trapSIGINTSource into t.TempDir() and returns the
// path. Skips on Windows (different signal semantics) and when `go` is not
// available (offline minimal envs).
func buildTrapBinary(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("SIGINT graceful-kill test is Unix-only (Windows has no SIGINT semantics)")
	}
	if _, err := osexec.LookPath("go"); err != nil {
		t.Skip("go toolchain not on PATH; cannot build trap binary")
	}

	dir := t.TempDir()
	src := filepath.Join(dir, "trap.go")
	if err := os.WriteFile(src, []byte(trapSIGINTSource), 0o600); err != nil {
		t.Fatalf("write trap source: %v", err)
	}
	bin := filepath.Join(dir, "trap")
	buildCtx, buildCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer buildCancel()
	build := osexec.CommandContext(buildCtx, "go", "build", "-o", bin, src) //nolint:gosec // test-controlled paths
	build.Env = append(os.Environ(), "GOTOOLCHAIN=local")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build trap: %v\n%s", err, out)
	}
	return bin
}

// TestBase_GracefulCancel_DeliversSIGINT proves that canceling the executor's
// context delivers SIGINT to the child (not SIGKILL) — the child catches it
// and exits with code 7. If WaitDelay had been left at zero or the cancel hook
// was missing, the child would be SIGKILL'd and exit with a signal-killed
// code instead.
func TestBase_GracefulCancel_DeliversSIGINT(t *testing.T) {
	bin := buildTrapBinary(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Capture stdout so we can synchronize on the "READY" marker.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer func() { _ = r.Close() }() //nolint:errcheck // test cleanup

	base := exec.NewBase()
	errCh := make(chan error, 1)
	go func() {
		errCh <- base.ExecuteStreaming(ctx, w, os.Stderr, bin)
		_ = w.Close() //nolint:errcheck // test cleanup
	}()

	// Wait until the child prints READY — guarantees it has installed its
	// signal handler. Without this, cancel could race the child's startup
	// and SIGINT could arrive before signal.Notify is wired up.
	buf := make([]byte, 64)
	readyCh := make(chan struct{})
	go func() {
		// Read until we see "READY" or EOF.
		acc := make([]byte, 0, 64)
		for {
			n, rerr := r.Read(buf)
			if n > 0 {
				acc = append(acc, buf[:n]...)
				if len(acc) >= len("READY") {
					close(readyCh)
					return
				}
			}
			if rerr != nil {
				return
			}
		}
	}()

	select {
	case <-readyCh:
	case <-time.After(10 * time.Second):
		t.Fatal("child never printed READY; signal-handler install failed")
	}

	cancel()

	// gracefulCancelDelay is 5s; allow generous slack for slow CI.
	select {
	case err := <-errCh:
		var exitErr *osexec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("expected *exec.ExitError, got %T: %v", err, err)
		}
		if exitErr.ExitCode() != 7 {
			t.Fatalf("child exit code = %d, want 7 (proves SIGINT was delivered, not SIGKILL); state=%v",
				exitErr.ExitCode(), exitErr.ProcessState)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("executor did not return within 15s of cancel; cmd.WaitDelay may be missing")
	}
}

// TestBase_GracefulCancel_KillsHungChild proves the WaitDelay safety net: a
// child that ignores SIGINT is still terminated within WaitDelay (not
// indefinitely). This ensures Ctrl+C cannot leave magex stuck.
func TestBase_GracefulCancel_KillsHungChild(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix-only: relies on POSIX signal semantics")
	}
	if _, err := osexec.LookPath("sh"); err != nil {
		t.Skip("sh not on PATH")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	base := exec.NewBase()
	errCh := make(chan error, 1)
	go func() {
		// Shell traps SIGINT and ignores it; sleeps forever. Only WaitDelay's
		// SIGKILL escalation can terminate this. Stream to io.Discard so the
		// child's inherited FDs don't outlive the test goroutine.
		errCh <- base.ExecuteStreaming(ctx, io.Discard, io.Discard, "sh", "-c",
			`trap "" INT TERM; sleep 60`)
	}()

	// Give the shell ~250ms to install its trap before canceling.
	time.Sleep(250 * time.Millisecond)
	cancel()

	// gracefulCancelDelay (5s) + slack for SIGKILL delivery and process reaping.
	select {
	case <-errCh:
		// Any return is fine — what matters is that the child died within the bound.
	case <-time.After(10 * time.Second):
		t.Fatal("hung child outlived WaitDelay; SIGKILL escalation failed")
	}
}
