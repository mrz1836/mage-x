package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	magefileFilename = "magefile.go"
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
	// Check for magefiles/ directory first (preferred by standard mage)
	magefilesDir := "magefiles"
	var targetPath string
	var useDirectory bool

	if info, err := os.Stat(magefilesDir); err == nil && info.IsDir() {
		targetPath = magefilesDir
		useDirectory = true
	} else {
		// Fallback to root magefile.go
		magefilePath := magefileFilename
		if _, err := os.Stat(magefilePath); os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrCommandNotFound, command)
		}
		targetPath = magefilePath
		useDirectory = false
	}

	var cmd *exec.Cmd
	ctx := context.Background()

	// Check if we have both magefiles/ directory and magefile.go (conflict situation)
	hasRootMagefile := false
	if _, err := os.Stat(magefileFilename); err == nil {
		hasRootMagefile = true
	}
	hasConflict := useDirectory && hasRootMagefile

	// Handle conflict by temporarily renaming magefile.go
	tempName := ""
	if hasConflict {
		tempName = magefileFilename + ".tmp"
		if err := os.Rename(magefileFilename, tempName); err != nil {
			return fmt.Errorf("failed to temporarily rename magefile.go: %w", err)
		}
		defer func() {
			if err := os.Rename(tempName, magefileFilename); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to restore magefile.go: %v\n", err)
			}
		}()
	}

	// Try to use mage first if available
	if magePath, err := exec.LookPath("mage"); err == nil {
		// Use mage binary - mage handles both directory and file automatically
		// For mage, we only pass the command name, arguments are provided via environment
		cmdArgs := []string{command}
		// #nosec G204 -- This is necessary for dynamic command execution with user-defined commands
		cmd = exec.CommandContext(ctx, magePath, cmdArgs...)
	} else {
		// Fallback to go run with mage tags
		var cmdArgs []string
		if useDirectory {
			// For directory, we need to run from within the directory using go run .
			cmdArgs = []string{"run", "-tags=mage", ".", command}
			cmdArgs = append(cmdArgs, args...)
			// #nosec G204 -- This is necessary for dynamic command execution with user-defined commands
			cmd = exec.CommandContext(ctx, "go", cmdArgs...)
			// Set the working directory to the magefiles directory
			cmd.Dir = targetPath
		} else {
			// For single file, specify the file path
			cmdArgs = []string{"run", "-tags=mage", targetPath, command}
			cmdArgs = append(cmdArgs, args...)
			// #nosec G204 -- This is necessary for dynamic command execution with user-defined commands
			cmd = exec.CommandContext(ctx, "go", cmdArgs...)
		}
	}

	// Set up environment
	cmd.Env = os.Environ()

	// Make arguments available via environment variable for magefile functions to access
	if len(args) > 0 {
		cmd.Env = append(cmd.Env, "MAGE_ARGS="+strings.Join(args, " "))
	}

	// Set working directory if not already set (directory case sets it specifically)
	if cmd.Dir == "" {
		if dir, err := os.Getwd(); err == nil {
			cmd.Dir = dir
		}
	}

	// Set up stdio with error filtering for stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	// Create a pipe to capture stderr for filtering
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start custom command '%s': %w", command, err)
	}

	// Filter stderr output in a goroutine
	go filterStderr(stderrPipe)

	// Wait for the command to finish
	if err := cmd.Wait(); err != nil {
		// Return the error instead of calling os.Exit
		// This allows the caller to decide how to handle the exit status
		return fmt.Errorf("failed to execute custom command '%s': %w", command, err)
	}

	return nil
}

// filterStderr filters stderr output to suppress spurious "Unknown target specified" messages
func filterStderr(stderrPipe io.ReadCloser) {
	defer func() {
		if err := stderrPipe.Close(); err != nil {
			// Log error but don't fail the operation
			fmt.Fprintf(os.Stderr, "Warning: failed to close stderr pipe: %v\n", err)
		}
	}()

	scanner := bufio.NewScanner(stderrPipe)
	for scanner.Scan() {
		line := scanner.Text()

		// Filter out spurious "Unknown target specified" warnings
		if strings.Contains(line, "Unknown target specified:") {
			continue // Skip this line
		}

		// Pass through all other stderr output
		fmt.Fprintln(os.Stderr, line)
	}
}

// HasMagefile checks if magefiles/ directory or magefile.go exists in the current directory
func HasMagefile() bool {
	// Check for magefiles/ directory first
	if info, err := os.Stat("magefiles"); err == nil && info.IsDir() {
		return true
	}
	// Fallback to magefile.go
	_, err := os.Stat("magefile.go")
	return err == nil
}

// GetMagefilePath returns the path to magefiles/ directory or magefile.go if it exists
func GetMagefilePath() string {
	// Check for magefiles/ directory first
	if info, err := os.Stat("magefiles"); err == nil && info.IsDir() {
		if abs, err := filepath.Abs("magefiles"); err == nil {
			return abs
		}
		return "magefiles"
	}
	// Fallback to magefile.go
	if _, err := os.Stat(magefileFilename); err == nil {
		if abs, err := filepath.Abs("magefile.go"); err == nil {
			return abs
		}
		return magefileFilename
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
