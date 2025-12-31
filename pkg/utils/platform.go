package utils

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

// errInvalidPlatformFormat is returned when a platform string cannot be parsed
var errInvalidPlatformFormat = errors.New("invalid platform format")

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
	if p.OS == osWindows {
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
