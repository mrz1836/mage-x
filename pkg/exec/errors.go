package exec

import (
	"fmt"
	"strings"
)

// CommandError formats a command failure error with consistent messaging.
// It wraps the underlying error with command context for debugging.
func CommandError(name string, args []string, err error) error {
	return fmt.Errorf("command failed [%s %s]: %w", name, strings.Join(args, " "), err)
}

// CommandErrorWithOutput formats a command failure error including command output.
// If output is empty or whitespace-only, it falls back to CommandError without output.
func CommandErrorWithOutput(name string, args []string, err error, output string) error {
	if trimmed := strings.TrimSpace(output); trimmed != "" {
		return fmt.Errorf("command failed [%s %s]: %w\n%s", name, strings.Join(args, " "), err, trimmed)
	}
	return CommandError(name, args, err)
}

// CommandErrorInDir formats a command failure error with directory context.
func CommandErrorInDir(name string, args []string, dir string, err error) error {
	return fmt.Errorf("command failed [%s %s] in %s: %w", name, strings.Join(args, " "), dir, err)
}

// CommandErrorInDirWithOutput formats a command failure error with directory and output.
// If output is empty or whitespace-only, it falls back to CommandErrorInDir without output.
func CommandErrorInDirWithOutput(name string, args []string, dir string, err error, output string) error {
	if trimmed := strings.TrimSpace(output); trimmed != "" {
		return fmt.Errorf("command failed [%s %s] in %s: %w\n%s", name, strings.Join(args, " "), dir, err, trimmed)
	}
	return CommandErrorInDir(name, args, dir, err)
}
