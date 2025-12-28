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
	"sync"
	"time"
)

const (
	magefileFilename = "magefile.go"
	// DefaultDelegateTimeout is the default timeout for delegated mage commands
	DefaultDelegateTimeout = 10 * time.Minute
)

var (
	// ErrCommandNotFound is returned when command is not found and no magefile.go exists
	ErrCommandNotFound = errors.New("command not found and no magefile.go exists")
	// ErrGoCommandNotFound is returned when go command is not available for delegation
	ErrGoCommandNotFound = errors.New("go command not found - required for custom magefile.go commands")
	// ErrCommandFailed is returned when a custom command execution fails
	ErrCommandFailed = errors.New("custom command failed")
	// ErrMagefileRestoreFailed is returned when the temporary magefile.go cannot be restored
	ErrMagefileRestoreFailed = errors.New("failed to restore magefile.go after command execution")
	// ErrCommandTimeout is returned when a delegated command times out
	ErrCommandTimeout = errors.New("command timed out")
)

// DelegateResult contains the result of a delegated mage command execution
type DelegateResult struct {
	// ExitCode is the exit code from the delegated command (0 for success)
	ExitCode int
	// Err contains any error that occurred during execution
	Err error
}

// convertToMageFormat converts colon-separated command names to mage's camelCase format
// e.g., "Speckit:Install" -> "speckitInstall", "Pipeline:CI" -> "pipelineCI"
func convertToMageFormat(command string) string {
	if !strings.Contains(command, ":") {
		return command // Already in simple format
	}

	parts := strings.SplitN(command, ":", 2)
	if len(parts) != 2 {
		return command
	}

	// Mage uses lowercase namespace + method (preserving method case)
	// e.g., Speckit:Install -> speckitInstall
	namespace := strings.ToLower(parts[0])
	method := parts[1]

	// First letter of method should be lowercase to form camelCase
	if len(method) > 0 {
		method = strings.ToLower(method[:1]) + method[1:]
	}

	return namespace + method
}

// DelegateToMage executes a custom command using mage or go run.
// Returns DelegateResult with the exit code and any error.
// This provides seamless execution of user-defined commands without plugin compilation.
func DelegateToMage(command string, args ...string) (result DelegateResult) {
	return DelegateToMageWithTimeout(command, DefaultDelegateTimeout, args...)
}

// DelegateToMageWithTimeout executes a custom command with a configurable timeout.
// Returns DelegateResult with the exit code and any error.
func DelegateToMageWithTimeout(command string, timeout time.Duration, args ...string) (result DelegateResult) {
	// Convert colon-separated format to mage's camelCase format
	mageCommand := convertToMageFormat(command)

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
			return DelegateResult{ExitCode: 1, Err: fmt.Errorf("%w: %s", ErrCommandNotFound, command)}
		}
		targetPath = magefilePath
		useDirectory = false
	}

	var cmd *exec.Cmd
	// Create context with timeout to prevent hanging commands
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

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
			return DelegateResult{ExitCode: 1, Err: fmt.Errorf("failed to temporarily rename magefile.go: %w", err)}
		}
		// Use named return to properly capture restore errors
		defer func() {
			if restoreErr := os.Rename(tempName, magefileFilename); restoreErr != nil {
				// Combine errors if command also failed
				if result.Err != nil {
					result.Err = fmt.Errorf("%w; additionally, %w: %w", result.Err, ErrMagefileRestoreFailed, restoreErr)
				} else {
					// Command succeeded but restore failed - this is still an error
					result.Err = fmt.Errorf("%w: %w", ErrMagefileRestoreFailed, restoreErr)
					result.ExitCode = 1
				}
			}
		}()
	}

	// Try to use mage first if available
	if magePath, err := exec.LookPath("mage"); err == nil {
		// Use mage binary - mage handles both directory and file automatically
		// NOTE: mage binary does NOT support command-line arguments for custom functions
		// Arguments must be passed via environment variables (MAGE_ARGS)
		cmdArgs := []string{mageCommand}
		// #nosec G204 -- This is necessary for dynamic command execution with user-defined commands
		cmd = exec.CommandContext(ctx, magePath, cmdArgs...)
	} else {
		// Fallback to go run with mage tags
		var cmdArgs []string
		if useDirectory {
			// For directory, we need to run from within the directory using go run .
			cmdArgs = []string{"run", "-tags=mage", ".", mageCommand}
			cmdArgs = append(cmdArgs, args...)
			// #nosec G204 -- This is necessary for dynamic command execution with user-defined commands
			cmd = exec.CommandContext(ctx, "go", cmdArgs...)
			// Set the working directory to the magefiles directory
			cmd.Dir = targetPath
		} else {
			// For single file, specify the file path
			cmdArgs = []string{"run", "-tags=mage", targetPath, mageCommand}
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

	// Create buffer and WaitGroup for stderr capture
	var stderrBuf strings.Builder
	var wg sync.WaitGroup
	wg.Add(1)

	// Create a pipe to capture stderr for filtering
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return DelegateResult{ExitCode: 1, Err: fmt.Errorf("failed to create stderr pipe: %w", err)}
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return DelegateResult{ExitCode: 1, Err: fmt.Errorf("failed to start custom command '%s': %w", command, err)}
	}

	// Filter stderr output in a goroutine, capturing content for error reporting
	go filterStderr(stderrPipe, &stderrBuf, &wg)

	// Wait for the command to finish
	waitErr := cmd.Wait()

	// Wait for stderr to be fully captured before returning
	wg.Wait()

	if waitErr != nil {
		// Extract actual exit code from the error
		exitCode := 1
		var exitErr *exec.ExitError
		if errors.As(waitErr, &exitErr) {
			exitCode = exitErr.ExitCode()
		}

		// Check for context deadline exceeded (timeout)
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return DelegateResult{
				ExitCode: 124, // Standard timeout exit code
				Err:      fmt.Errorf("%w: '%s' after %v", ErrCommandTimeout, command, timeout),
			}
		}

		// Include captured stderr in the error message if available
		if stderrContent := strings.TrimSpace(stderrBuf.String()); stderrContent != "" {
			return DelegateResult{
				ExitCode: exitCode,
				Err:      fmt.Errorf("%w '%s':\n%s", ErrCommandFailed, command, stderrContent),
			}
		}
		return DelegateResult{
			ExitCode: exitCode,
			Err:      fmt.Errorf("%w '%s': %w", ErrCommandFailed, command, waitErr),
		}
	}

	return DelegateResult{ExitCode: 0, Err: nil}
}

// filterStderr filters stderr output for real-time display while capturing all content for error reporting
// The "Unknown target specified" messages are filtered from display to avoid duplication when shown in error
func filterStderr(stderrPipe io.ReadCloser, buf *strings.Builder, wg *sync.WaitGroup) {
	defer wg.Done()
	defer func() {
		if err := stderrPipe.Close(); err != nil {
			// Log error but don't fail the operation
			fmt.Fprintf(os.Stderr, "Warning: failed to close stderr pipe: %v\n", err)
		}
	}()

	scanner := bufio.NewScanner(stderrPipe)
	for scanner.Scan() {
		line := scanner.Text()

		// Always capture to buffer for error message (we need the full context)
		buf.WriteString(line)
		buf.WriteString("\n")

		// Filter "Unknown target specified" from real-time display to avoid duplication
		// (it will still appear in the error message if command fails)
		if strings.Contains(line, "Unknown target specified:") {
			continue // Skip displaying this line
		}

		// Pass through all other stderr output to user in real-time
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
