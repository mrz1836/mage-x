// Package utils provides utility functions for mage tasks
package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mrz1836/mage-x/pkg/common/env"
	pkgexec "github.com/mrz1836/mage-x/pkg/exec"
)

// DefaultExecutor is the package-level executor using pkg/exec.
// This provides integration with the unified command execution package.
// Can be overridden in tests using SetExecutor.
var DefaultExecutor pkgexec.Executor = pkgexec.Simple() //nolint:gochecknoglobals // Package singleton for backward compatibility

// SetExecutor allows overriding the default executor for testing.
func SetExecutor(e pkgexec.Executor) {
	DefaultExecutor = e
}

// ResetExecutor restores the default executor.
func ResetExecutor() {
	DefaultExecutor = pkgexec.Simple()
}

// RunCmd executes a command and returns its output
// Uses the unified pkg/exec package under the hood
func RunCmd(name string, args ...string) error {
	if env.IsVerbose() {
		Info("➤ %s %s", name, strings.Join(args, " "))
	}

	if err := DefaultExecutor.Execute(context.Background(), name, args...); err != nil {
		return pkgexec.CommandError(name, args, err)
	}
	return nil
}

// RunCmdV executes a command with verbose output
func RunCmdV(name string, args ...string) error {
	Info("➤ %s %s", name, strings.Join(args, " "))
	return RunCmd(name, args...)
}

// RunCmdOutput executes a command and returns its output
// Uses the unified pkg/exec package under the hood
func RunCmdOutput(name string, args ...string) (string, error) {
	if env.IsVerbose() {
		Info("➤ %s %s", name, strings.Join(args, " "))
	}

	output, err := DefaultExecutor.ExecuteOutput(context.Background(), name, args...)
	if err != nil {
		return "", pkgexec.CommandErrorWithOutput(name, args, err, output)
	}
	return output, nil
}

// RunCmdPipe executes commands in a pipeline
// Note: This uses direct exec.Cmd as pipelines require special handling
func RunCmdPipe(cmds ...*exec.Cmd) error {
	for i, cmd := range cmds {
		if i > 0 {
			pipe, err := cmds[i-1].StdoutPipe()
			if err != nil {
				return fmt.Errorf("failed to create stdout pipe: %w", err)
			}
			cmds[i].Stdin = pipe
		}
		if i < len(cmds)-1 {
			cmd.Stdout = nil
		} else {
			cmd.Stdout = os.Stdout
		}
		cmd.Stderr = os.Stderr
	}

	for i, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start pipeline command %d [%s]: %w", i+1, cmd.Path, err)
		}
	}

	for i, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("pipeline command %d [%s] failed: %w", i+1, cmd.Path, err)
		}
	}

	return nil
}

// CommandExists checks if a command exists in PATH
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

// RunCmdSecure executes a command with security validation and environment filtering
// This is a convenience function for secure command execution
func RunCmdSecure(name string, args ...string) error {
	executor := pkgexec.Secure()
	if env.IsVerbose() {
		Info("➤ [secure] %s %s", name, strings.Join(args, " "))
	}

	if err := executor.Execute(context.Background(), name, args...); err != nil {
		return pkgexec.CommandError(name, args, err)
	}
	return nil
}

// RunCmdWithRetry executes a command with retry support for transient failures
func RunCmdWithRetry(maxRetries int, name string, args ...string) error {
	executor := pkgexec.SecureWithRetry(maxRetries)
	if env.IsVerbose() {
		Info("➤ [retryable:%d] %s %s", maxRetries, name, strings.Join(args, " "))
	}

	if err := executor.Execute(context.Background(), name, args...); err != nil {
		return pkgexec.CommandError(name, args, err)
	}
	return nil
}
