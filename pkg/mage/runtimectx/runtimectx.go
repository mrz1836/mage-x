package runtimectx

import (
	"context"
	"errors"
	"sync"
)

//nolint:gochecknoglobals // process-wide root context is the package's purpose
var (
	mu   sync.RWMutex
	root = context.Background()
)

// SetRoot publishes the process-wide root context. Call once from main()
// after installing the signal handler. A nil ctx is coerced to Background.
//
//nolint:contextcheck // the entire purpose of this function is to publish a brand-new root context
func SetRoot(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}
	mu.Lock()
	root = ctx
	mu.Unlock()
}

// Context returns the published root context, or context.Background if
// SetRoot was never called (library / test mode). Safe under concurrent use.
func Context() context.Context {
	mu.RLock()
	defer mu.RUnlock()
	return root
}

// CheckCanceled returns the root context's error (context.Canceled or
// context.DeadlineExceeded) if it has been canceled, otherwise nil.
// Loops call this at iteration boundaries to bail cleanly on Ctrl+C.
func CheckCanceled() error {
	return Context().Err()
}

// IsCanceled reports whether err is a context cancellation (Canceled or
// DeadlineExceeded), unwrapped. Use to identify cancel-origin errors
// returned from runners/loops so callers can distinguish user-canceled
// failures from real errors.
func IsCanceled(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

// Reset restores the default (Background) root. For test cleanup only.
func Reset() {
	SetRoot(context.Background())
}
