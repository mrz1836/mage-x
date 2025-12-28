package utils

import (
	"errors"
	"fmt"
	"strings"
)

// errParseGoVersion is returned when Go version cannot be parsed
var errParseGoVersion = errors.New("unable to parse go version")

// GoList runs go list and returns the output
func GoList(args ...string) ([]string, error) {
	cmdArgs := append([]string{"list"}, args...)
	output, err := RunCmdOutput("go", cmdArgs...)
	if err != nil {
		// Include command output in error for better diagnostics (e.g., dependency conflicts)
		if output = strings.TrimSpace(output); output != "" {
			return nil, fmt.Errorf("%w\n%s", err, output)
		}
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	var result []string
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			result = append(result, line)
		}
	}

	return result, nil
}

// GetModuleName returns the current module name from go.mod
func GetModuleName() (string, error) {
	output, err := RunCmdOutput("go", "list", "-m")
	if err != nil {
		return "", fmt.Errorf("failed to get module name: %w", err)
	}
	return strings.TrimSpace(output), nil
}

// GetModuleNameInDir returns the module name from go.mod in the specified directory.
// Uses the -C flag (Go 1.20+) to change directory before running the command.
func GetModuleNameInDir(dir string) (string, error) {
	output, err := RunCmdOutput("go", "-C", dir, "list", "-m")
	if err != nil {
		return "", fmt.Errorf("failed to get module name in %s: %w", dir, err)
	}
	return strings.TrimSpace(output), nil
}

// GetGoVersion returns the Go version
func GetGoVersion() (string, error) {
	output, err := RunCmdOutput("go", "version")
	if err != nil {
		return "", fmt.Errorf("failed to get Go version: %w", err)
	}

	// Parse version from output like "go version go1.24.0 darwin/amd64"
	parts := strings.Fields(output)
	if len(parts) >= 3 {
		return strings.TrimPrefix(parts[2], "go"), nil
	}

	return "", fmt.Errorf("%w from: %s", errParseGoVersion, output)
}
