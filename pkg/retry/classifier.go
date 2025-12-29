// Package retry provides unified retry logic with error classification and backoff strategies.
package retry

import (
	"context"
	"errors"
	"io"
	"net"
	"os/exec"
	"strings"
	"syscall"
)

// Classifier determines whether an error should trigger a retry.
type Classifier interface {
	IsRetriable(err error) bool
}

// ClassifierFunc is an adapter to allow ordinary functions to be used as Classifiers.
type ClassifierFunc func(err error) bool

// IsRetriable implements Classifier.
func (f ClassifierFunc) IsRetriable(err error) bool {
	return f(err)
}

// DefaultClassifier provides comprehensive error classification for network and command operations.
type DefaultClassifier struct {
	// Patterns are string patterns to match in error messages (case-insensitive)
	Patterns []string

	// SyscallErrors are syscall.Errno values that are considered retriable
	SyscallErrors []syscall.Errno

	// CheckNetErrors enables checking for net.Error types
	CheckNetErrors bool

	// CheckDNSErrors enables checking for net.DNSError types
	CheckDNSErrors bool

	// CheckOpErrors enables checking for net.OpError types
	CheckOpErrors bool

	// CheckIOErrors enables checking for io.EOF and io.ErrUnexpectedEOF
	CheckIOErrors bool

	// CheckExitCodes enables checking for exec.ExitError with specific exit codes
	CheckExitCodes bool

	// RetriableExitCodes are exit codes that are considered retriable
	RetriableExitCodes []int

	// CheckHTTPStatus enables checking for HTTP 5xx status codes in error messages
	CheckHTTPStatus bool
}

// IsRetriable implements Classifier.
func (c *DefaultClassifier) IsRetriable(err error) bool {
	if err == nil {
		return false
	}

	// Context errors are never retriable
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Check string patterns
	if c.matchesPatterns(err) {
		return true
	}

	// Check network error types
	if c.CheckNetErrors {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return true
		}
	}

	// Check DNS errors
	if c.CheckDNSErrors {
		var dnsErr *net.DNSError
		if errors.As(err, &dnsErr) {
			errLower := strings.ToLower(dnsErr.Err)
			// DNS errors are retriable unless they're permanent
			if !strings.Contains(errLower, "no such host") ||
				strings.Contains(errLower, "temporary failure") {
				return true
			}
		}
	}

	// Check operation errors
	if c.CheckOpErrors {
		var opErr *net.OpError
		if errors.As(err, &opErr) {
			return true
		}
	}

	// Check syscall errors
	if len(c.SyscallErrors) > 0 {
		var syscallErr syscall.Errno
		if errors.As(err, &syscallErr) {
			for _, retriable := range c.SyscallErrors {
				if syscallErr == retriable {
					return true
				}
			}
		}
	}

	// Check IO errors
	if c.CheckIOErrors {
		if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
			return true
		}
	}

	// Check exit codes
	if c.CheckExitCodes {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode := exitErr.ExitCode()
			for _, retriable := range c.RetriableExitCodes {
				if exitCode == retriable {
					return true
				}
			}
			// Also check if stderr contains retriable patterns
			if exitErr.Stderr != nil && c.matchesPatternsInString(string(exitErr.Stderr)) {
				return true
			}
		}
	}

	// Check HTTP status codes
	if c.CheckHTTPStatus {
		errStr := err.Error()
		if strings.Contains(errStr, "status 5") {
			return true
		}
	}

	return false
}

// matchesPatterns checks if the error message contains any of the configured patterns.
func (c *DefaultClassifier) matchesPatterns(err error) bool {
	return c.matchesPatternsInString(err.Error())
}

// matchesPatternsInString checks if a string contains any of the configured patterns.
func (c *DefaultClassifier) matchesPatternsInString(s string) bool {
	if len(c.Patterns) == 0 {
		return false
	}

	lower := strings.ToLower(s)
	for _, pattern := range c.Patterns {
		if strings.Contains(lower, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// NetworkPatterns are common network-related error patterns.
//
//nolint:gochecknoglobals // Package-level constants for reuse
var NetworkPatterns = []string{
	"connection refused",
	"connection reset",
	"connection timeout",
	"timeout",
	"temporary failure in name resolution",
	"no such host",
	"no route to host",
	"network is unreachable",
	"host is unreachable",
	"i/o timeout",
	"unexpected eof",
	"tls handshake timeout",
	"dial tcp",
	"proxyconnect tcp",
}

// GoModulePatterns are Go module/download specific error patterns.
//
//nolint:gochecknoglobals // Package-level constants for reuse
var GoModulePatterns = []string{
	"go: downloading",
	"go: module",
	"verifying module",
	"getting requirements",
	"sumdb verification",
}

// NetworkSyscallErrors are syscall errors that indicate network issues.
//
//nolint:gochecknoglobals // Package-level constants for reuse
var NetworkSyscallErrors = []syscall.Errno{
	syscall.ECONNRESET,
	syscall.ECONNREFUSED,
	syscall.ETIMEDOUT,
	syscall.EHOSTUNREACH,
	syscall.ENETUNREACH,
}

// NewNetworkClassifier creates a classifier optimized for network operations (HTTP, downloads).
func NewNetworkClassifier() *DefaultClassifier {
	return &DefaultClassifier{
		Patterns:        NetworkPatterns,
		SyscallErrors:   NetworkSyscallErrors,
		CheckNetErrors:  true,
		CheckDNSErrors:  true,
		CheckOpErrors:   true,
		CheckIOErrors:   true,
		CheckHTTPStatus: true,
	}
}

// NewCommandClassifier creates a classifier optimized for command execution.
func NewCommandClassifier() *DefaultClassifier {
	patterns := make([]string, 0, len(NetworkPatterns)+len(GoModulePatterns))
	patterns = append(patterns, NetworkPatterns...)
	patterns = append(patterns, GoModulePatterns...)

	return &DefaultClassifier{
		Patterns:           patterns,
		SyscallErrors:      NetworkSyscallErrors,
		CheckNetErrors:     true,
		CheckDNSErrors:     true,
		CheckOpErrors:      true,
		CheckExitCodes:     true,
		RetriableExitCodes: []int{1, 124}, // 1 = generic failure (check stderr), 124 = timeout
	}
}

// AlwaysRetry is a classifier that considers all errors retriable.
// Use with caution and always with a max retry limit.
//
//nolint:gochecknoglobals // Intentional package-level singleton
var AlwaysRetry Classifier = ClassifierFunc(func(err error) bool {
	return err != nil && !errors.Is(err, context.Canceled)
})

// NeverRetry is a classifier that never considers errors retriable.
//
//nolint:gochecknoglobals // Intentional package-level singleton
var NeverRetry Classifier = ClassifierFunc(func(_ error) bool {
	return false
})
