package exec

import (
	"fmt"
	"strings"
)

// CommandErrorOptions contains optional parameters for command error formatting.
type CommandErrorOptions struct {
	Dir    string // Working directory (empty if not applicable)
	Output string // Command output to include (empty if not applicable)
}

// NewCommandError creates a formatted error for a failed command execution.
// The opts parameter is optional and can be used to include directory and output info.
func NewCommandError(name string, args []string, err error, opts *CommandErrorOptions) error {
	argsStr := strings.Join(args, " ")

	var msg string
	if opts != nil && opts.Dir != "" {
		msg = fmt.Sprintf("command failed [%s %s] in %s", name, argsStr, opts.Dir)
	} else {
		msg = fmt.Sprintf("command failed [%s %s]", name, argsStr)
	}

	if opts != nil {
		if trimmed := strings.TrimSpace(opts.Output); trimmed != "" {
			return fmt.Errorf("%s: %w\n%s", msg, err, trimmed)
		}
	}
	return fmt.Errorf("%s: %w", msg, err)
}

// CommandError formats a command failure error with consistent messaging.
// It wraps the underlying error with command context for debugging.
func CommandError(name string, args []string, err error) error {
	return NewCommandError(name, args, err, nil)
}

// CommandErrorWithOutput formats a command failure error including command output.
// If output is empty or whitespace-only, it falls back to CommandError without output.
func CommandErrorWithOutput(name string, args []string, err error, output string) error {
	return NewCommandError(name, args, err, &CommandErrorOptions{Output: output})
}

// CommandErrorInDir formats a command failure error with directory context.
func CommandErrorInDir(name string, args []string, dir string, err error) error {
	return NewCommandError(name, args, err, &CommandErrorOptions{Dir: dir})
}

// CommandErrorInDirWithOutput formats a command failure error with directory and output.
// If output is empty or whitespace-only, it falls back to CommandErrorInDir without output.
func CommandErrorInDirWithOutput(name string, args []string, dir string, err error, output string) error {
	return NewCommandError(name, args, err, &CommandErrorOptions{Dir: dir, Output: output})
}
