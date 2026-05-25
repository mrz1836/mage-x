package mage_test

import (
	"context"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/mage/runtimectx"
)

// TestSecureCommandRunner_RespectsCancellation proves the universal Layer 2
// plumbing: when the runtime root context is canceled, an in-flight
// subprocess started via SecureCommandRunner is terminated within the grace
// window — not after its full timeout.
//
// Deterministic guarantees:
//   - Uses runtimectx.Reset via t.Cleanup so global state never leaks.
//   - Asserts an UPPER bound (15s) — generous slack on slow CI; passing
//     within 6s in practice (SIGINT + 5s WaitDelay).
//   - Uses a real `sleep` subprocess so the chain (runner → exec →
//     CommandContext → kernel SIGINT) is end-to-end exercised.
func TestSecureCommandRunner_RespectsCancellation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix-only: relies on the POSIX `sleep` command")
	}
	if _, err := exec.LookPath("sleep"); err != nil {
		t.Skip("sleep not on PATH")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runtimectx.SetRoot(ctx)
	t.Cleanup(runtimectx.Reset)

	runner := mage.NewSecureCommandRunner()

	errCh := make(chan error, 1)
	startedAt := time.Now()
	go func() {
		// Sleep is much longer than gracefulCancelDelay (5s) — any return
		// before its natural end proves cancellation propagated.
		errCh <- runner.RunCmd("sleep", "60")
	}()

	// Give the subprocess a moment to start. Without this, cancel could
	// race the fork and Go's context-aware exec might never observe a
	// running process.
	time.Sleep(200 * time.Millisecond)

	cancel()

	select {
	case err := <-errCh:
		elapsed := time.Since(startedAt)
		// Generous upper bound: 200ms startup + 5s WaitDelay + slack.
		if elapsed > 15*time.Second {
			t.Fatalf("RunCmd took %v to honor cancel; expected <15s", elapsed)
		}
		// Canceled subprocess should produce an error (not nil) — the
		// shape varies (could be *exec.ExitError or context.Canceled),
		// what matters is propagation worked.
		if err == nil {
			t.Fatalf("RunCmd returned nil error after cancel; expected propagation failure")
		}
	case <-time.After(20 * time.Second):
		t.Fatalf("RunCmd never returned after cancel; runtimectx → runner → exec chain broken")
	}
}

// TestSecureCommandRunner_NoRegression confirms that a non-canceled run
// still completes normally — defends against accidentally killing healthy
// subprocesses.
func TestSecureCommandRunner_NoRegression(t *testing.T) {
	if _, err := exec.LookPath("true"); err != nil {
		t.Skip("true not on PATH")
	}

	t.Cleanup(runtimectx.Reset)
	runtimectx.Reset() // ensure default Background root

	runner := mage.NewSecureCommandRunner()
	if err := runner.RunCmd("true"); err != nil {
		t.Fatalf("RunCmd(true) failed unexpectedly: %v", err)
	}
}
