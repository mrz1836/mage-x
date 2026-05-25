package runtimectx_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/mrz1836/mage-x/pkg/mage/runtimectx"
)

// errSentinelForTest is a static, non-cancel error used to verify IsCanceled
// returns false for unrelated errors. Defined here (not inline) to satisfy err113.
var errSentinelForTest = errors.New("boom")

// resetAfter restores the default Background root when the test finishes,
// so global state does not bleed between tests (see Test sandbox hygiene).
func resetAfter(t *testing.T) {
	t.Helper()
	t.Cleanup(runtimectx.Reset)
}

func TestContext_DefaultIsBackground(t *testing.T) {
	resetAfter(t)
	runtimectx.Reset()

	ctx := runtimectx.Context()
	if ctx == nil {
		t.Fatal("Context() returned nil")
	}
	if err := ctx.Err(); err != nil {
		t.Fatalf("default Context().Err() = %v, want nil", err)
	}
}

func TestSetRoot_RoundTrip(t *testing.T) {
	resetAfter(t)

	parent, cancel := context.WithCancel(context.Background())
	defer cancel()

	runtimectx.SetRoot(parent)
	got := runtimectx.Context()
	if got != parent {
		t.Fatalf("Context() returned different context than SetRoot")
	}

	cancel()
	select {
	case <-got.Done():
		// expected
	case <-time.After(time.Second):
		t.Fatal("canceling parent did not propagate through runtimectx")
	}
}

func TestSetRoot_NilCoercesToBackground(t *testing.T) {
	resetAfter(t)

	runtimectx.SetRoot(nil) //nolint:staticcheck // exercising the nil-guard

	ctx := runtimectx.Context()
	if ctx == nil {
		t.Fatal("Context() returned nil after SetRoot(nil)")
	}
	if err := ctx.Err(); err != nil {
		t.Fatalf("Context().Err() = %v after SetRoot(nil), want nil", err)
	}
}

func TestReset_RestoresBackground(t *testing.T) {
	resetAfter(t)

	parent, cancel := context.WithCancel(context.Background())
	defer cancel()
	runtimectx.SetRoot(parent)
	cancel()
	if runtimectx.CheckCanceled() == nil {
		t.Fatal("expected CheckCanceled to report error before Reset")
	}

	runtimectx.Reset()
	if err := runtimectx.CheckCanceled(); err != nil {
		t.Fatalf("after Reset, CheckCanceled() = %v, want nil", err)
	}
}

func TestCheckCanceled(t *testing.T) {
	resetAfter(t)

	if err := runtimectx.CheckCanceled(); err != nil {
		t.Fatalf("baseline CheckCanceled() = %v, want nil", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	runtimectx.SetRoot(ctx)
	cancel()

	err := runtimectx.CheckCanceled()
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("CheckCanceled() = %v, want context.Canceled", err)
	}
}

func TestIsCanceled(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"canceled", context.Canceled, true},
		{"deadline", context.DeadlineExceeded, true},
		{"wrapped canceled", fmt.Errorf("op: %w", context.Canceled), true},
		{"wrapped deadline", fmt.Errorf("op: %w", context.DeadlineExceeded), true},
		{"plain", errSentinelForTest, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := runtimectx.IsCanceled(tc.err); got != tc.want {
				t.Fatalf("IsCanceled(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestConcurrent_SetAndGet(t *testing.T) {
	resetAfter(t)

	const goroutines = 100
	const iterations = 200

	var wg sync.WaitGroup
	wg.Add(goroutines * 3)

	for range goroutines {
		go func() {
			defer wg.Done()
			for range iterations {
				ctx, cancel := context.WithCancel(context.Background())
				runtimectx.SetRoot(ctx)
				cancel()
			}
		}()
		go func() {
			defer wg.Done()
			for range iterations {
				_ = runtimectx.Context()
			}
		}()
		go func() {
			defer wg.Done()
			for range iterations {
				_ = runtimectx.CheckCanceled() //nolint:errcheck // concurrency-stress only; result intentionally discarded
			}
		}()
	}
	wg.Wait()
}
