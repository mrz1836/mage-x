package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"sync/atomic"
	"syscall"
	"testing"
	"time"
)

// waitFor blocks until ch is closed or the deadline elapses. Returns true if
// ch closed first, false on timeout. Used in place of arbitrary time.Sleep
// to keep tests deterministic.
func waitFor(t *testing.T, ch <-chan struct{}, d time.Duration) bool {
	t.Helper()
	select {
	case <-ch:
		return true
	case <-time.After(d):
		return false
	}
}

// TestInstallSignalHandler_FirstSIGINT verifies that the first SIGINT cancels
// the context, writes the banner exactly once, and classifies the signal as
// SIGINT — without any real signal delivery.
func TestInstallSignalHandler_FirstSIGINT(t *testing.T) {
	sigCh := make(chan os.Signal, 2)
	var cancelledBy atomic.Int32
	var stderr bytes.Buffer

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exitCalled := make(chan int, 1)
	exit := func(code int) { exitCalled <- code }

	done := installSignalHandler(sigCh, &cancelledBy, cancel, exit, &stderr)

	sigCh <- os.Interrupt

	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("context was not canceled within 2s of SIGINT")
	}

	// Close the channel so the handler goroutine returns deterministically.
	close(sigCh)
	if !waitFor(t, done, 2*time.Second) {
		t.Fatal("signal-handler goroutine did not exit after channel closed")
	}

	if got := cancelledBy.Load(); got != cancelledBySIGINT {
		t.Fatalf("cancelledBy = %d, want %d (SIGINT)", got, cancelledBySIGINT)
	}
	if !bytes.Contains(stderr.Bytes(), []byte("Canceled by user")) {
		t.Fatalf("stderr missing banner: %q", stderr.String())
	}
	select {
	case code := <-exitCalled:
		t.Fatalf("exit was called with %d on single SIGINT; should only fire on second signal", code)
	default:
	}
}

// TestInstallSignalHandler_SIGTERM verifies CI-style cancellation paths.
func TestInstallSignalHandler_SIGTERM(t *testing.T) {
	sigCh := make(chan os.Signal, 2)
	var cancelledBy atomic.Int32
	var stderr bytes.Buffer

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := installSignalHandler(sigCh, &cancelledBy, cancel, func(int) {}, &stderr)

	sigCh <- syscall.SIGTERM

	select {
	case <-ctx.Done():
	case <-time.After(2 * time.Second):
		t.Fatal("context was not canceled within 2s of SIGTERM")
	}

	close(sigCh)
	if !waitFor(t, done, 2*time.Second) {
		t.Fatal("signal-handler goroutine did not exit")
	}

	if got := cancelledBy.Load(); got != cancelledBySIGTERM {
		t.Fatalf("cancelledBy = %d, want %d (SIGTERM)", got, cancelledBySIGTERM)
	}
}

// TestInstallSignalHandler_SecondSignalForceExits verifies that a second
// signal calls exit(130) — the user's force-quit escape hatch.
func TestInstallSignalHandler_SecondSignalForceExits(t *testing.T) {
	sigCh := make(chan os.Signal, 2)
	var cancelledBy atomic.Int32
	var stderr bytes.Buffer

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	exitCalled := make(chan int, 1)
	exit := func(code int) { exitCalled <- code }

	done := installSignalHandler(sigCh, &cancelledBy, cancel, exit, &stderr)

	// Send two signals in a row. The buffer of size 2 + the goroutine's
	// two reads guarantee both are delivered without dropping.
	sigCh <- os.Interrupt
	sigCh <- os.Interrupt

	select {
	case code := <-exitCalled:
		if code != exitCodeSIGINT {
			t.Fatalf("exit called with %d, want %d", code, exitCodeSIGINT)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("exit was not called within 2s of second SIGINT")
	}

	if !bytes.Contains(stderr.Bytes(), []byte("Force exit")) {
		t.Fatalf("stderr missing 'Force exit' marker: %q", stderr.String())
	}

	close(sigCh)
	_ = waitFor(t, done, time.Second) // goroutine may already have exited via exit()
	_ = ctx                           // keep ctx referenced
}

// TestInstallSignalHandler_BannerPrintedExactlyOnce guards against double-banner
// races when signals arrive in quick succession.
func TestInstallSignalHandler_BannerPrintedExactlyOnce(t *testing.T) {
	sigCh := make(chan os.Signal, 2)
	var cancelledBy atomic.Int32
	var stderr bytes.Buffer

	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	exitCalled := make(chan int, 1)
	done := installSignalHandler(sigCh, &cancelledBy, cancel, func(code int) { exitCalled <- code }, &stderr)

	sigCh <- os.Interrupt
	sigCh <- os.Interrupt

	// Wait for force-exit to fire so we know both signals have been processed.
	select {
	case <-exitCalled:
	case <-time.After(2 * time.Second):
		t.Fatal("exit not called; handler may be stuck")
	}

	close(sigCh)
	_ = waitFor(t, done, time.Second)

	count := bytes.Count(stderr.Bytes(), []byte("Canceled by user"))
	if count != 1 {
		t.Fatalf("banner printed %d times, want exactly 1; full stderr: %q", count, stderr.String())
	}
}

// TestInstallSignalHandler_ChannelCloseExitsCleanly ensures the handler does
// not leak its goroutine when sigCh closes before any signal is delivered
// (mirrors orderly shutdown via signal.Stop).
func TestInstallSignalHandler_ChannelCloseExitsCleanly(t *testing.T) {
	sigCh := make(chan os.Signal, 2)
	var cancelledBy atomic.Int32
	var stderr bytes.Buffer

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := installSignalHandler(sigCh, &cancelledBy, cancel, func(int) {
		t.Errorf("exit should not be called on clean shutdown")
	}, &stderr)

	close(sigCh)
	if !waitFor(t, done, 2*time.Second) {
		t.Fatal("handler did not exit after sigCh closed")
	}

	if got := cancelledBy.Load(); got != cancelledByNone {
		t.Fatalf("cancelledBy = %d, want %d (none) — handler should not classify on clean close", got, cancelledByNone)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr should be empty on clean close, got: %q", stderr.String())
	}
	if err := ctx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("unexpected ctx error: %v", err)
	}
}
