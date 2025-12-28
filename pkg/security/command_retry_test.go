package security

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"syscall"
	"testing"
	"time"
)

// Static error definitions for err113 compliance
var (
	errConnectionRefused       = errors.New("connection refused")
	errConnectionReset         = errors.New("connection reset by peer")
	errIOTimeout               = errors.New("i/o timeout")
	errNoSuchHost              = errors.New("no such host: example.com")
	errNetworkUnreachable      = errors.New("network is unreachable")
	errHostUnreachable         = errors.New("host is unreachable")
	errContextDeadlineExceeded = errors.New("context deadline exceeded")
	errUnexpectedEOF           = errors.New("unexpected EOF")
	errTLSHandshakeTimeout     = errors.New("tls handshake timeout")
	errDialTCP                 = errors.New("dial tcp: connection failed")
	errGoDownloading           = errors.New("go: downloading module")
	errGoModule                = errors.New("go: module not found")
	errVerifyingModule         = errors.New("verifying module checksum")
	errGettingRequirements     = errors.New("getting requirements failed")
	errSumdbVerification       = errors.New("sumdb verification failed")
	errFileNotFound            = errors.New("file not found")
	errPermissionDenied        = errors.New("permission denied")
	errInvalidArgument         = errors.New("invalid argument")
	errNetworkTimeout          = errors.New("network timeout")
)

func TestSecureExecutor_ExecuteWithRetry_Success(t *testing.T) {
	executor := NewSecureExecutor()
	executor.DryRun = true // Use dry run to avoid actually executing commands

	ctx := context.Background()
	err := executor.ExecuteWithRetry(ctx, 3, 100*time.Millisecond, "echo", "test")
	if err != nil {
		t.Errorf("ExecuteWithRetry should succeed in dry run mode, got: %v", err)
	}
}

func TestSecureExecutor_ExecuteOutputWithRetry_Success(t *testing.T) {
	executor := NewSecureExecutor()
	executor.DryRun = true

	ctx := context.Background()
	output, err := executor.ExecuteOutputWithRetry(ctx, 3, 100*time.Millisecond, "echo", "test")
	if err != nil {
		t.Errorf("ExecuteOutputWithRetry should succeed in dry run mode, got: %v", err)
	}

	if !strings.Contains(output, "Would execute: echo test") {
		t.Errorf("Expected dry run output, got: %s", output)
	}
}

func TestIsRetriableCommandError(t *testing.T) {
	testCases := []struct {
		name      string
		err       error
		retriable bool
	}{
		{
			name:      "nil error",
			err:       nil,
			retriable: false,
		},
		{
			name:      "connection refused",
			err:       errConnectionRefused,
			retriable: true,
		},
		{
			name:      "connection reset",
			err:       errConnectionReset,
			retriable: true,
		},
		{
			name:      "timeout error",
			err:       errIOTimeout,
			retriable: true,
		},
		{
			name:      "no such host",
			err:       errNoSuchHost,
			retriable: true,
		},
		{
			name:      "network unreachable",
			err:       errNetworkUnreachable,
			retriable: true,
		},
		{
			name:      "host unreachable",
			err:       errHostUnreachable,
			retriable: true,
		},
		{
			name:      "context deadline exceeded",
			err:       errContextDeadlineExceeded,
			retriable: true,
		},
		{
			name:      "unexpected eof",
			err:       errUnexpectedEOF,
			retriable: true,
		},
		{
			name:      "tls handshake timeout",
			err:       errTLSHandshakeTimeout,
			retriable: true,
		},
		{
			name:      "dial tcp error",
			err:       errDialTCP,
			retriable: true,
		},
		{
			name:      "go downloading error",
			err:       errGoDownloading,
			retriable: true,
		},
		{
			name:      "go module error",
			err:       errGoModule,
			retriable: true,
		},
		{
			name:      "verifying module error",
			err:       errVerifyingModule,
			retriable: true,
		},
		{
			name:      "getting requirements error",
			err:       errGettingRequirements,
			retriable: true,
		},
		{
			name:      "sumdb verification error",
			err:       errSumdbVerification,
			retriable: true,
		},
		{
			name:      "file not found error",
			err:       errFileNotFound,
			retriable: false,
		},
		{
			name:      "permission denied error",
			err:       errPermissionDenied,
			retriable: false,
		},
		{
			name:      "invalid argument error",
			err:       errInvalidArgument,
			retriable: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isRetriableCommandError(tc.err)
			if result != tc.retriable {
				t.Errorf("Expected retriable=%v for error %q, got %v", tc.retriable, tc.err, result)
			}
		})
	}
}

func TestIsRetriableCommandError_NetworkErrors(t *testing.T) {
	// Test with actual network error types
	// Test a network timeout error using a static error
	timeoutErr := errNetworkTimeout
	if !isRetriableCommandError(timeoutErr) {
		t.Error("Network timeout error should be retriable")
	}

	mockDNSError := &net.DNSError{
		Err:  "no such host",
		Name: "example.com",
	}

	if !isRetriableCommandError(mockDNSError) {
		t.Error("Temporary DNS error should be retriable")
	}

	mockOpError := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: errConnectionRefused,
	}

	if !isRetriableCommandError(mockOpError) {
		t.Error("Connection op error should be retriable")
	}
}

func TestIsRetriableCommandError_SyscallErrors(t *testing.T) {
	// Test syscall errors that should be retriable
	retriableSyscallErrors := []syscall.Errno{
		syscall.ECONNRESET,
		syscall.ECONNREFUSED,
		syscall.ETIMEDOUT,
		syscall.EHOSTUNREACH,
		syscall.ENETUNREACH,
	}

	for _, sysErr := range retriableSyscallErrors {
		// Wrap syscall error in a way that errors.As can detect it
		wrappedErr := fmt.Errorf("wrapped syscall error: %w", sysErr)
		if !isRetriableCommandError(wrappedErr) {
			t.Errorf("Syscall error %v (%s) should be retriable", sysErr, sysErr.Error())
		}
	}

	// Test syscall errors that should not be retriable
	nonRetriableSyscallErrors := []syscall.Errno{
		syscall.ENOENT, // No such file or directory
		syscall.EACCES, // Permission denied
		syscall.EINVAL, // Invalid argument
	}

	for _, sysErr := range nonRetriableSyscallErrors {
		// Wrap syscall error in a way that errors.As can detect it
		wrappedErr := fmt.Errorf("wrapped syscall error: %w", sysErr)
		if isRetriableCommandError(wrappedErr) {
			t.Errorf("Syscall error %v should not be retriable", sysErr)
		}
	}
}

// NOTE: TestIsRetriableCommandError_ExitErrors was removed as the retry logic
// has been migrated to pkg/retry and pkg/exec. Tests for retry behavior are in:
// - pkg/retry/retry_test.go
// - pkg/retry/classifier_test.go

// NOTE: TestSecureExecutor_RetryBackoff was removed - retry backoff is now tested in pkg/retry/backoff_test.go

// NOTE: TestSecureExecutor_RetryContextCancellation was removed - context cancellation is tested in pkg/retry/retry_test.go

// NOTE: TestSecureExecutor_NonRetriableError was removed - non-retriable error behavior is tested in pkg/retry/classifier_test.go

// NOTE: BenchmarkSecureExecutor_ExecuteWithRetry_WithFailures was removed - retry benchmarks are in pkg/retry/retry_test.go
