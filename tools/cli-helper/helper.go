// Package clihelper provides utilities for command-line applications
package clihelper

import (
	"fmt"
	"strings"
)

// FormatCommand formats a command for display
func FormatCommand(cmd string, args []string) string {
	if len(args) == 0 {
		return cmd
	}
	return fmt.Sprintf("%s %s", cmd, strings.Join(args, " "))
}

// ParseArgs parses command-line arguments
func ParseArgs(args []string) (string, []string) {
	if len(args) == 0 {
		return "", nil
	}
	return args[0], args[1:]
}
