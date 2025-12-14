// Package utils provides utility functions for mage tasks
package utils

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mrz1836/mage-x/pkg/common/fileops"
)

// Constants for common string values
const (
	TrueValue = "true"
)

// Static errors for utils operations
var (
	errInvalidPlatformFormat = errors.New("invalid platform format")
	errParseGoVersion        = errors.New("unable to parse go version")
	errParallelExecution     = errors.New("parallel execution failed")
)

// --- cmd.go ---

// RunCmd executes a command and returns its output
func RunCmd(name string, args ...string) error {
	cmd := exec.CommandContext(context.Background(), name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if IsVerbose() {
		Info("➤ %s %s", name, strings.Join(args, " "))
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed [%s %s]: %w", name, strings.Join(args, " "), err)
	}
	return nil
}

// RunCmdV executes a command with verbose output
func RunCmdV(name string, args ...string) error {
	Info("➤ %s %s", name, strings.Join(args, " "))
	return RunCmd(name, args...)
}

// RunCmdOutput executes a command and returns its output
func RunCmdOutput(name string, args ...string) (string, error) {
	cmd := exec.CommandContext(context.Background(), name, args...)
	cmd.Env = os.Environ()

	if IsVerbose() {
		Info("➤ %s %s", name, strings.Join(args, " "))
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Include command output in error for better diagnostics
		if trimmed := strings.TrimSpace(string(output)); trimmed != "" {
			return "", fmt.Errorf("command failed [%s %s]: %w\n%s", name, strings.Join(args, " "), err, trimmed)
		}
		return "", fmt.Errorf("command failed [%s %s]: %w", name, strings.Join(args, " "), err)
	}
	return string(output), nil
}

// RunCmdPipe executes commands in a pipeline
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

// --- env.go ---

// GetEnv returns an environment variable with a default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvBool returns a boolean environment variable
func GetEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == TrueValue || value == "1" || value == "yes"
}

// GetEnvInt returns an integer environment variable
func GetEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var result int
	if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
		return result
	}
	return defaultValue
}

// GetEnvClean retrieves an environment variable with inline comment stripping
// It removes anything after " #" and trims whitespace from the value
// This is useful for environment files that contain inline comments like:
// VARIABLE_NAME=value  # comment here
func GetEnvClean(key string) string {
	value := os.Getenv(key)
	if value == "" {
		return ""
	}

	// Find inline comment marker (space followed by #)
	if idx := strings.Index(value, " #"); idx >= 0 {
		value = value[:idx]
	}

	// Trim any leading/trailing whitespace
	return strings.TrimSpace(value)
}

// IsVerbose checks if verbose mode is enabled
func IsVerbose() bool {
	return GetEnvBool("VERBOSE", false) || GetEnvBool("V", false)
}

// IsCI checks if running in CI environment
func IsCI() bool {
	ciVars := []string{
		"CI",
		"CONTINUOUS_INTEGRATION",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"TRAVIS",
		"CIRCLECI",
		"JENKINS_URL",
		"CODEBUILD_BUILD_ID",
	}

	for _, v := range ciVars {
		if os.Getenv(v) != "" {
			return true
		}
	}
	return false
}

// --- fs.go ---

// FileExists checks if a file exists
func FileExists(path string) bool {
	return fileops.GetDefault().File.Exists(path)
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	return fileops.GetDefault().File.Exists(path) && fileops.GetDefault().File.IsDir(path)
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	return fileops.GetDefault().File.MkdirAll(path, 0o755)
}

// CleanDir removes and recreates a directory
func CleanDir(path string) error {
	if err := fileops.GetDefault().File.RemoveAll(path); err != nil {
		// Ignore "not exists" errors, similar to original behavior
		if !os.IsNotExist(err) {
			return err
		}
	}
	return fileops.GetDefault().File.MkdirAll(path, 0o755)
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	return fileops.GetDefault().File.Copy(src, dst)
}

// FindFiles finds files matching a pattern
func FindFiles(_, pattern string) ([]string, error) {
	return findFiles(".", pattern)
}

func findFiles(root, pattern string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, _ os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			return err
		}

		// Use fileops to check if it's a file (not directory)
		if matched && fileops.GetDefault().File.Exists(path) && !fileops.GetDefault().File.IsDir(path) {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// --- platform.go ---

// Platform represents a target platform
type Platform struct {
	OS   string
	Arch string
}

// String returns the platform as OS/Arch
func (p Platform) String() string {
	return fmt.Sprintf("%s/%s", p.OS, p.Arch)
}

// GetCurrentPlatform returns the current platform
func GetCurrentPlatform() Platform {
	return Platform{
		OS:   runtime.GOOS,
		Arch: runtime.GOARCH,
	}
}

// ParsePlatform parses a platform string (e.g., "linux/amd64")
func ParsePlatform(s string) (Platform, error) {
	parts := strings.Split(s, "/")
	if len(parts) != 2 {
		return Platform{}, fmt.Errorf("%w: %s", errInvalidPlatformFormat, s)
	}
	// Validate that both OS and Arch are non-empty
	if parts[0] == "" || parts[1] == "" {
		return Platform{}, fmt.Errorf("%w: %s", errInvalidPlatformFormat, s)
	}
	return Platform{OS: parts[0], Arch: parts[1]}, nil
}

// GetBinaryExt returns the binary extension for a platform
func GetBinaryExt(p Platform) string {
	if p.OS == "windows" {
		return ".exe"
	}
	return ""
}

// IsWindows checks if the current platform is Windows
func IsWindows() bool {
	return runtime.GOOS == "windows"
}

// IsMac checks if the current platform is macOS
func IsMac() bool {
	return runtime.GOOS == "darwin"
}

// IsLinux checks if the current platform is Linux
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// GetShell returns the appropriate shell command
func GetShell() (shell string, args []string) {
	if IsWindows() {
		return "cmd", []string{"/c"}
	}
	return "sh", []string{"-c"}
}

// --- Additional utilities ---

// CommandExists checks if a command exists in PATH
func CommandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

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

// Parallel runs functions in parallel
func Parallel(fns ...func() error) error {
	type result struct {
		err error
	}

	ch := make(chan result, len(fns))

	for _, fn := range fns {
		go func(f func() error) {
			ch <- result{err: f()}
		}(fn)
	}

	var errs []error
	for range fns {
		res := <-ch
		if res.err != nil {
			errs = append(errs, res.err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("%w: %v", errParallelExecution, errs)
	}

	return nil
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	return formatDuration(d) // Use the shared implementation from logger.go
}

// FormatBytes formats bytes in a human-readable way
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// PromptForInput prompts the user for input and returns the response
func PromptForInput(prompt string) (string, error) {
	if prompt != "" {
		fmt.Printf("%s: ", prompt)
	}

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}
		return "", nil // EOF
	}

	return strings.TrimSpace(scanner.Text()), nil
}

// CheckFileLineLength checks if any line in a file exceeds maxLen bytes
// Returns: (hasLongLines, lineNumber, lineLength, error)
// Uses bufio.Reader with a large buffer to avoid token size limits
func CheckFileLineLength(path string, maxLen int) (bool, int, int, error) {
	file, err := os.Open(path) // #nosec G304 -- path is validated before use
	if err != nil {
		return false, 0, 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() {
		_ = file.Close() //nolint:errcheck // Best-effort close, errors not actionable in validation context
	}()

	// Use a reader with a large buffer (128KB) to handle lines larger than default
	reader := bufio.NewReaderSize(file, 128*1024)
	lineNum := 0
	maxLineLen := 0

	for {
		lineNum++
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			// If we get "token too long" or similar, report what we know
			return true, lineNum, maxLineLen, nil
		}

		lineLen := len(line)

		// If isPrefix is true, the line is longer than our buffer
		// Keep reading until we get the full line
		for isPrefix {
			var part []byte
			part, isPrefix, err = reader.ReadLine()
			if err != nil {
				break
			}
			lineLen += len(part)
		}

		if lineLen > maxLineLen {
			maxLineLen = lineLen
		}

		// Early return if we find a line exceeding the limit
		if lineLen > maxLen {
			return true, lineNum, lineLen, nil
		}
	}

	return false, 0, maxLineLen, nil
}

// Note: Print functions have been moved to logger.go
// Use the logger package functions instead:
// - Header(text)
// - Success(format, args...)
// - Error(format, args...)
// - Info(format, args...)
// - Warn(format, args...)
