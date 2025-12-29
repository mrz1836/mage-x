package exec

import (
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
)

// EnvFilteringExecutor wraps an executor with environment variable filtering
type EnvFilteringExecutor struct {
	wrapped FullExecutor

	// SensitivePrefixes are environment variable prefixes to filter out
	SensitivePrefixes []string

	// Whitelist maps command names to environment variables they're allowed to access
	// e.g. "goreleaser" -> ["GITHUB_TOKEN", "GITLAB_TOKEN"]
	Whitelist map[string][]string
}

// EnvFilterOption configures an EnvFilteringExecutor
type EnvFilterOption func(*EnvFilteringExecutor)

// WithSensitivePrefixes sets the sensitive prefixes to filter
func WithSensitivePrefixes(prefixes []string) EnvFilterOption {
	return func(e *EnvFilteringExecutor) {
		e.SensitivePrefixes = prefixes
	}
}

// WithEnvWhitelist sets the command-specific env whitelist
func WithEnvWhitelist(whitelist map[string][]string) EnvFilterOption {
	return func(e *EnvFilteringExecutor) {
		e.Whitelist = whitelist
	}
}

// DefaultSensitivePrefixes are common environment variable prefixes containing secrets
//
//nolint:gochecknoglobals // Package-level constant for convenience
var DefaultSensitivePrefixes = []string{
	"AWS_SECRET",
	"GITHUB_TOKEN",
	"GITLAB_TOKEN",
	"NPM_TOKEN",
	"DOCKER_PASSWORD",
	"DATABASE_PASSWORD",
	"API_KEY",
	"SECRET",
	"PRIVATE_KEY",
}

// DefaultEnvWhitelist contains common command-specific env whitelist entries
//
//nolint:gochecknoglobals // Package-level constant for convenience
var DefaultEnvWhitelist = map[string][]string{
	"goreleaser": {"GITHUB_TOKEN", "GITLAB_TOKEN", "GITEA_TOKEN"},
}

// NewEnvFilteringExecutor creates a new environment filtering executor
func NewEnvFilteringExecutor(wrapped FullExecutor, opts ...EnvFilterOption) *EnvFilteringExecutor {
	e := &EnvFilteringExecutor{
		wrapped:           wrapped,
		SensitivePrefixes: DefaultSensitivePrefixes,
		Whitelist:         DefaultEnvWhitelist,
	}
	for _, opt := range opts {
		opt(e)
	}
	return e
}

// Execute runs a command with filtered environment
func (e *EnvFilteringExecutor) Execute(ctx context.Context, name string, args ...string) error {
	// We need to intercept command creation to filter environment
	// Since we can't modify the wrapped executor's behavior directly,
	// we create a new command with filtered environment

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = e.FilterEnvironment(os.Environ(), name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ExecuteOutput runs a command with filtered environment and returns output
func (e *EnvFilteringExecutor) ExecuteOutput(ctx context.Context, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = e.FilterEnvironment(os.Environ(), name)

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// ExecuteWithEnv runs a command with additional environment variables (filtered)
func (e *EnvFilteringExecutor) ExecuteWithEnv(ctx context.Context, env []string, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)

	// Filter base environment, then add additional env
	baseEnv := e.FilterEnvironment(os.Environ(), name)
	cmd.Env = make([]string, 0, len(baseEnv)+len(env))
	cmd.Env = append(cmd.Env, baseEnv...)
	cmd.Env = append(cmd.Env, env...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ExecuteStreaming runs a command with filtered environment and custom output
func (e *EnvFilteringExecutor) ExecuteStreaming(ctx context.Context, stdout, stderr io.Writer, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = e.FilterEnvironment(os.Environ(), name)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	return cmd.Run()
}

// ExecuteInDir runs a command in the specified directory with filtered environment
func (e *EnvFilteringExecutor) ExecuteInDir(ctx context.Context, dir, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = e.FilterEnvironment(os.Environ(), name)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ExecuteOutputInDir runs a command in the specified directory with filtered environment and returns output
func (e *EnvFilteringExecutor) ExecuteOutputInDir(ctx context.Context, dir, name string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = e.FilterEnvironment(os.Environ(), name)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// FilterEnvironment removes sensitive environment variables
func (e *EnvFilteringExecutor) FilterEnvironment(env []string, commandName string) []string {
	filtered := make([]string, 0, len(env))
	for _, envVar := range env {
		keep := true

		// Extract the variable name part (before =)
		parts := strings.SplitN(envVar, "=", 2)
		if len(parts) < 2 {
			// Malformed env var (no =), include it as is
			filtered = append(filtered, envVar)
			continue
		}

		varName := strings.ToUpper(parts[0])

		// Check if this environment variable contains sensitive data
		for _, prefix := range e.SensitivePrefixes {
			if !strings.HasPrefix(varName, prefix) {
				continue
			}
			if len(varName) != len(prefix) && varName[len(prefix)] != '_' {
				continue
			}

			// Check if this variable is whitelisted for this command
			if whitelistedVars, ok := e.Whitelist[commandName]; ok {
				for _, whitelistedVar := range whitelistedVars {
					if varName == whitelistedVar {
						// This sensitive variable is whitelisted for this command
						keep = true
						goto nextVar
					}
				}
			}
			// Only filter if it's an exact match or followed by underscore
			keep = false
			break
		}
	nextVar:

		if keep {
			filtered = append(filtered, envVar)
		}
	}

	return filtered
}

// Ensure EnvFilteringExecutor implements FullExecutor
var _ FullExecutor = (*EnvFilteringExecutor)(nil)
