// Package exec provides unified command execution with security, retry, and timeout support.
//
// The package uses a decorator pattern for composability:
//
//	base := exec.NewBase()
//	secured := exec.WithValidation(base)
//	retrying := exec.WithRetry(secured, retry.DefaultClassifier)
//	final := exec.WithTimeout(retrying, 5*time.Minute)
//
// For convenience, use the Builder:
//
//	executor := exec.NewBuilder().
//		WithValidation().
//		WithRetry(retry.DefaultClassifier).
//		WithTimeout(5*time.Minute).
//		Build()
package exec

import (
	"context"
	"io"
)

// Executor defines the interface for command execution.
// All implementations should be safe for concurrent use.
type Executor interface {
	// Execute runs a command with the given arguments
	Execute(ctx context.Context, name string, args ...string) error

	// ExecuteOutput runs a command and returns its combined output
	ExecuteOutput(ctx context.Context, name string, args ...string) (string, error)
}

// ExecutorWithEnv extends Executor with custom environment support
type ExecutorWithEnv interface {
	Executor

	// ExecuteWithEnv runs a command with additional environment variables
	ExecuteWithEnv(ctx context.Context, env []string, name string, args ...string) error
}

// ExecutorWithDir extends Executor with working directory support
type ExecutorWithDir interface {
	Executor

	// ExecuteInDir runs a command in the specified directory
	ExecuteInDir(ctx context.Context, dir, name string, args ...string) error

	// ExecuteOutputInDir runs a command in the specified directory and returns output
	ExecuteOutputInDir(ctx context.Context, dir, name string, args ...string) (string, error)
}

// StreamingExecutor extends Executor with output streaming support
type StreamingExecutor interface {
	Executor

	// ExecuteStreaming runs a command with custom stdout/stderr
	ExecuteStreaming(ctx context.Context, stdout, stderr io.Writer, name string, args ...string) error
}

// FullExecutor combines all executor interfaces
type FullExecutor interface {
	Executor
	ExecutorWithEnv
	ExecutorWithDir
	StreamingExecutor
}
