package mage

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/mrz1836/mage-x/pkg/mage/runtimectx"
)

// countingRunner is a minimal CommandRunner that counts invocations without
// spawning any subprocess. Used to prove that runCommandInModuleWithRunner
// bails out on a canceled runtime context BEFORE it reaches the runner —
// which is what makes magex exit cleanly on Ctrl+C instead of grinding
// through every remaining module.
type countingRunner struct {
	calls atomic.Int64
}

func (r *countingRunner) RunCmd(_ string, _ ...string) error {
	r.calls.Add(1)
	return nil
}

func (r *countingRunner) RunCmdOutput(_ string, _ ...string) (string, error) {
	r.calls.Add(1)
	return "", nil
}

func (r *countingRunner) RunCmdInDir(_, _ string, _ ...string) error {
	r.calls.Add(1)
	return nil
}

func (r *countingRunner) RunCmdOutputInDir(_, _ string, _ ...string) (string, error) {
	r.calls.Add(1)
	return "", nil
}

// TestRunCommandInModuleWithRunner_BailsOnCancel proves the Layer 3a
// chokepoint guard: a canceled root context causes the helper to return
// immediately without invoking the runner.
func TestRunCommandInModuleWithRunner_BailsOnCancel(t *testing.T) {
	t.Cleanup(runtimectx.Reset)

	ctx, cancel := context.WithCancel(context.Background())
	runtimectx.SetRoot(ctx)
	cancel() // pre-canceled

	runner := &countingRunner{}
	module := ModuleInfo{Path: "/tmp/x", Relative: "x"}

	err := runCommandInModuleWithRunner(module, runner, "echo", "hi")

	if err == nil {
		t.Fatal("expected cancellation error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error does not wrap context.Canceled: %v", err)
	}
	if !strings.Contains(err.Error(), "canceled") {
		t.Fatalf("error message missing 'cancelled' marker: %v", err)
	}
	if got := runner.calls.Load(); got != 0 {
		t.Fatalf("runner was invoked %d times after cancel; expected 0", got)
	}
}

// TestRunCommandInModuleOutputWithRunner_BailsOnCancel is the same check for
// the output-returning variant.
func TestRunCommandInModuleOutputWithRunner_BailsOnCancel(t *testing.T) {
	t.Cleanup(runtimectx.Reset)

	ctx, cancel := context.WithCancel(context.Background())
	runtimectx.SetRoot(ctx)
	cancel()

	runner := &countingRunner{}
	module := ModuleInfo{Path: "/tmp/x", Relative: "x"}

	out, err := runCommandInModuleOutputWithRunner(module, runner, "echo", "hi")

	if out != "" {
		t.Fatalf("expected empty output on cancel, got %q", out)
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error does not wrap context.Canceled: %v", err)
	}
	if got := runner.calls.Load(); got != 0 {
		t.Fatalf("runner was invoked %d times after cancel; expected 0", got)
	}
}

// TestForEachModule_BailsOnCancelMidLoop proves the Layer 3b loop guard: a
// cancellation that arrives PARTWAY through a multi-module operation stops
// iteration on the next module instead of grinding through them all. This is
// the exact bug the user reported in the second screenshot.
func TestForEachModule_BailsOnCancelMidLoop(t *testing.T) {
	t.Cleanup(runtimectx.Reset)

	ctx, cancel := context.WithCancel(context.Background())
	runtimectx.SetRoot(ctx)

	modules := []ModuleInfo{
		{Path: "/tmp/a", Relative: "a"},
		{Path: "/tmp/b", Relative: "b"},
		{Path: "/tmp/c", Relative: "c"},
		{Path: "/tmp/d", Relative: "d"},
	}

	var processed atomic.Int64
	action := func(_ ModuleInfo) error {
		n := processed.Add(1)
		if n == 1 {
			// Simulate Ctrl+C arriving while module #1 is running.
			cancel()
		}
		return nil
	}

	err := forEachModule(modules, ModuleIteratorOptions{Operation: "test", Verb: "tested"}, action)

	if err == nil {
		t.Fatal("expected cancellation error after mid-loop cancel, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error does not wrap context.Canceled: %v", err)
	}
	if got := processed.Load(); got >= int64(len(modules)) {
		t.Fatalf("forEachModule processed %d modules after cancel; expected <%d", got, len(modules))
	}
	if got := processed.Load(); got < 1 {
		t.Fatalf("forEachModule processed %d modules; expected at least 1 (the one that triggered the cancel)", got)
	}
}

// TestForEachModule_NoRegression confirms a clean (uncancelled) iteration
// still processes every module.
func TestForEachModule_NoRegression(t *testing.T) {
	t.Cleanup(runtimectx.Reset)
	runtimectx.Reset()

	modules := []ModuleInfo{
		{Path: "/tmp/a", Relative: "a"},
		{Path: "/tmp/b", Relative: "b"},
		{Path: "/tmp/c", Relative: "c"},
	}

	var processed atomic.Int64
	action := func(_ ModuleInfo) error {
		processed.Add(1)
		return nil
	}

	if err := forEachModule(modules, ModuleIteratorOptions{Operation: "test", Verb: "tested"}, action); err != nil {
		t.Fatalf("forEachModule returned unexpected error: %v", err)
	}
	if got := processed.Load(); got != int64(len(modules)) {
		t.Fatalf("forEachModule processed %d/%d modules", got, len(modules))
	}
}
