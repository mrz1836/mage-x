// Package utils provides utility functions for mage tasks
package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mrz1836/go-mage/pkg/common/fileops"
)

// --- cmd.go ---

// RunCmd executes a command and returns its output
func RunCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if IsVerbose() {
		Info("➤ %s %s", name, strings.Join(args, " "))
	}

	return cmd.Run()
}

// RunCmdV executes a command with verbose output
func RunCmdV(name string, args ...string) error {
	Info("➤ %s %s", name, strings.Join(args, " "))
	return RunCmd(name, args...)
}

// RunCmdOutput executes a command and returns its output
func RunCmdOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Env = os.Environ()

	if IsVerbose() {
		Info("➤ %s %s", name, strings.Join(args, " "))
	}

	output, err := cmd.CombinedOutput()
	return string(output), err
}

// RunCmdPipe executes commands in a pipeline
func RunCmdPipe(cmds ...*exec.Cmd) error {
	for i, cmd := range cmds {
		if i > 0 {
			cmds[i].Stdin, _ = cmds[i-1].StdoutPipe()
		}
		if i < len(cmds)-1 {
			cmd.Stdout = nil
		} else {
			cmd.Stdout = os.Stdout
		}
		cmd.Stderr = os.Stderr
	}

	for _, cmd := range cmds {
		if err := cmd.Start(); err != nil {
			return err
		}
	}

	for _, cmd := range cmds {
		if err := cmd.Wait(); err != nil {
			return err
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
	return value == "true" || value == "1" || value == "yes"
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

var fileOps = fileops.New()

// FileExists checks if a file exists
func FileExists(path string) bool {
	return fileOps.File.Exists(path)
}

// DirExists checks if a directory exists
func DirExists(path string) bool {
	return fileOps.File.Exists(path) && fileOps.File.IsDir(path)
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	return fileOps.File.MkdirAll(path, 0o755)
}

// CleanDir removes and recreates a directory
func CleanDir(path string) error {
	if err := fileOps.File.RemoveAll(path); err != nil {
		// Ignore "not exists" errors, similar to original behavior
		if !os.IsNotExist(err) {
			return err
		}
	}
	return fileOps.File.MkdirAll(path, 0o755)
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	return fileOps.File.Copy(src, dst)
}

// FindFiles finds files matching a pattern
func FindFiles(dir, pattern string) ([]string, error) {
	return findFiles(".", pattern)
}

func findFiles(root, pattern string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			return err
		}

		// Use fileops to check if it's a file (not directory)
		if matched && fileOps.File.Exists(path) && !fileOps.File.IsDir(path) {
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
		return Platform{}, fmt.Errorf("invalid platform format: %s", s)
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
func GetShell() (string, []string) {
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
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// GetGoVersion returns the Go version
func GetGoVersion() (string, error) {
	output, err := RunCmdOutput("go", "version")
	if err != nil {
		return "", err
	}

	// Parse version from output like "go version go1.24.0 darwin/amd64"
	parts := strings.Fields(output)
	if len(parts) >= 3 {
		return strings.TrimPrefix(parts[2], "go"), nil
	}

	return "", fmt.Errorf("unable to parse go version from: %s", output)
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
		return fmt.Errorf("parallel execution failed: %v", errs)
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

// Note: Print functions have been moved to logger.go
// Use the logger package functions instead:
// - Header(text)
// - Success(format, args...)
// - Error(format, args...)
// - Info(format, args...)
// - Warn(format, args...)
