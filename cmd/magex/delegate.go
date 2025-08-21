package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

var (
	// ErrCommandNotFound is returned when command is not found and no magefile.go exists
	ErrCommandNotFound = errors.New("command not found and no magefile.go exists")
	// ErrGoCommandNotFound is returned when go command is not available for delegation
	ErrGoCommandNotFound = errors.New("go command not found - required for custom magefile.go commands")
)

// DelegateToMage executes a custom command using mage or go run
// This provides seamless execution of user-defined commands without plugin compilation
func DelegateToMage(command string, args ...string) error {
	// Check if magefile.go exists
	magefilePath := "magefile.go"
	if _, err := os.Stat(magefilePath); os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrCommandNotFound, command)
	}

	var cmd *exec.Cmd
	ctx := context.Background()

	// Try to use mage first if available
	if magePath, err := exec.LookPath("mage"); err == nil {
		// Use mage binary
		cmdArgs := []string{command}
		cmdArgs = append(cmdArgs, args...)
		// #nosec G204 -- This is necessary for dynamic command execution with user-defined commands
		cmd = exec.CommandContext(ctx, magePath, cmdArgs...)
	} else {
		// Fallback to go run with mage tags
		cmdArgs := []string{"run", "-tags=mage", magefilePath, command}
		cmdArgs = append(cmdArgs, args...)
		// #nosec G204 -- This is necessary for dynamic command execution with user-defined commands
		cmd = exec.CommandContext(ctx, "go", cmdArgs...)
	}

	// Set up environment
	cmd.Env = os.Environ()
	if dir, err := os.Getwd(); err == nil {
		cmd.Dir = dir
	}

	// Connect stdio for real-time output
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute the command
	if err := cmd.Run(); err != nil {
		// Extract exit code for proper error propagation
		exitError := &exec.ExitError{}
		if errors.As(err, &exitError) {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		return fmt.Errorf("failed to execute custom command '%s': %w", command, err)
	}

	return nil
}

// HasMagefile checks if magefile.go exists in the current directory
func HasMagefile() bool {
	_, err := os.Stat("magefile.go")
	return err == nil
}

// GetMagefilePath returns the path to magefile.go if it exists
func GetMagefilePath() string {
	if HasMagefile() {
		if abs, err := filepath.Abs("magefile.go"); err == nil {
			return abs
		}
		return "magefile.go"
	}
	return ""
}

// ValidateGoEnvironment checks if Go is available for delegation
func ValidateGoEnvironment() error {
	_, err := exec.LookPath("go")
	if err != nil {
		return fmt.Errorf("%w", ErrGoCommandNotFound)
	}
	return nil
}
