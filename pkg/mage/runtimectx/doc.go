// Package runtimectx publishes a process-wide root context for magex.
//
// The binary entry point (cmd/magex/main.go) installs a signal-aware
// context via SetRoot before dispatching any command. Helpers throughout
// pkg/mage and pkg/exec read it via Context so that Ctrl+C / SIGTERM
// propagates to every subprocess and every iterating loop.
//
// Library consumers that never call SetRoot get context.Background, so
// behavior is unchanged from the pre-signal-aware era.
package runtimectx
