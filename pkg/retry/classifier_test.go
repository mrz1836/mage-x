package retry

import (
	"context"
	"errors"
	"io"
	"net"
	"syscall"
	"testing"
)

// Test error types for classification tests
var (
	errConnectionRefused  = errors.New("connection refused")
	errConnectionReset    = errors.New("connection reset by peer")
	errIOTimeout          = errors.New("i/o timeout")
	errNoSuchHost         = errors.New("no such host")
	errNetworkUnreachable = errors.New("network is unreachable")
	errHostUnreachable    = errors.New("host is unreachable")
	errTLSHandshake       = errors.New("tls handshake timeout")
	errDialTCP            = errors.New("dial tcp: connection failed")
	errGoDownloading      = errors.New("go: downloading module failed")
	errGoModule           = errors.New("go: module verification failed")
	errVerifyingModule    = errors.New("verifying module checksum failed")
	errHTTP500            = errors.New("status 500: internal server error")
	errHTTP502            = errors.New("status 502: bad gateway")
	errPermanent          = errors.New("permanent failure: invalid configuration")
	errNotFound           = errors.New("file not found")
)

func TestDefaultClassifier_NetworkPatterns(t *testing.T) {
	classifier := NewNetworkClassifier()

	tests := []struct {
		name      string
		err       error
		retriable bool
	}{
		{"nil error", nil, false},
		{"connection refused", errConnectionRefused, true},
		{"connection reset", errConnectionReset, true},
		{"i/o timeout", errIOTimeout, true},
		{"no such host", errNoSuchHost, true},
		{"network unreachable", errNetworkUnreachable, true},
		{"host unreachable", errHostUnreachable, true},
		{"tls handshake timeout", errTLSHandshake, true},
		{"dial tcp", errDialTCP, true},
		{"http 500", errHTTP500, true},
		{"http 502", errHTTP502, true},
		{"permanent error", errPermanent, false},
		{"file not found", errNotFound, false},
		{"context canceled", context.Canceled, false},
		{"context deadline exceeded", context.DeadlineExceeded, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.IsRetriable(tt.err)
			if result != tt.retriable {
				t.Errorf("IsRetriable(%v) = %v, want %v", tt.err, result, tt.retriable)
			}
		})
	}
}

func TestDefaultClassifier_GoModulePatterns(t *testing.T) {
	classifier := NewCommandClassifier()

	tests := []struct {
		name      string
		err       error
		retriable bool
	}{
		{"go downloading", errGoDownloading, true},
		{"go module", errGoModule, true},
		{"verifying module", errVerifyingModule, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.IsRetriable(tt.err)
			if result != tt.retriable {
				t.Errorf("IsRetriable(%v) = %v, want %v", tt.err, result, tt.retriable)
			}
		})
	}
}

func TestDefaultClassifier_SyscallErrors(t *testing.T) {
	classifier := NewNetworkClassifier()

	tests := []struct {
		name      string
		err       error
		retriable bool
	}{
		{"ECONNRESET", syscall.ECONNRESET, true},
		{"ECONNREFUSED", syscall.ECONNREFUSED, true},
		{"ETIMEDOUT", syscall.ETIMEDOUT, true},
		{"EHOSTUNREACH", syscall.EHOSTUNREACH, true},
		{"ENETUNREACH", syscall.ENETUNREACH, true},
		{"ENOENT", syscall.ENOENT, false},
		{"EPERM", syscall.EPERM, false},
		{"EACCES", syscall.EACCES, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.IsRetriable(tt.err)
			if result != tt.retriable {
				t.Errorf("IsRetriable(%v) = %v, want %v", tt.err, result, tt.retriable)
			}
		})
	}
}

func TestDefaultClassifier_IOErrors(t *testing.T) {
	classifier := NewNetworkClassifier()

	tests := []struct {
		name      string
		err       error
		retriable bool
	}{
		{"io.EOF", io.EOF, true},
		{"io.ErrUnexpectedEOF", io.ErrUnexpectedEOF, true},
		{"io.ErrClosedPipe", io.ErrClosedPipe, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.IsRetriable(tt.err)
			if result != tt.retriable {
				t.Errorf("IsRetriable(%v) = %v, want %v", tt.err, result, tt.retriable)
			}
		})
	}
}

func TestDefaultClassifier_ContextErrors(t *testing.T) {
	classifier := NewNetworkClassifier()

	errOpFailed := errors.New("operation failed") //nolint:err113 // Test error

	tests := []struct {
		name      string
		err       error
		retriable bool
	}{
		{"context.Canceled", context.Canceled, false},
		{"context.DeadlineExceeded", context.DeadlineExceeded, false},
		{"wrapped context.Canceled", errors.Join(errOpFailed, context.Canceled), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.IsRetriable(tt.err)
			if result != tt.retriable {
				t.Errorf("IsRetriable(%v) = %v, want %v", tt.err, result, tt.retriable)
			}
		})
	}
}

func TestDefaultClassifier_WrappedErrors(t *testing.T) {
	classifier := NewNetworkClassifier()

	// Static errors for test cases
	errFailedToConnect := errors.New("failed to connect") //nolint:err113 // Test error
	errReadFailed := errors.New("read failed")            //nolint:err113 // Test error
	errOperationFailed := errors.New("operation failed")  //nolint:err113 // Test error

	tests := []struct {
		name      string
		err       error
		retriable bool
	}{
		{
			name:      "wrapped connection refused",
			err:       errors.Join(errFailedToConnect, errConnectionRefused),
			retriable: true,
		},
		{
			name:      "wrapped syscall ECONNRESET",
			err:       errors.Join(errReadFailed, syscall.ECONNRESET),
			retriable: true,
		},
		{
			name:      "wrapped permanent error",
			err:       errors.Join(errOperationFailed, errPermanent),
			retriable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.IsRetriable(tt.err)
			if result != tt.retriable {
				t.Errorf("IsRetriable(%v) = %v, want %v", tt.err, result, tt.retriable)
			}
		})
	}
}

func TestDefaultClassifier_CaseInsensitive(t *testing.T) {
	classifier := NewNetworkClassifier()

	// Static errors for case sensitivity tests
	errUppercase := errors.New("CONNECTION REFUSED")              //nolint:err113 // Test error
	errMixedCase := errors.New("Connection Refused")              //nolint:err113 // Test error
	errUpperTimeout := errors.New("TIMEOUT waiting for response") //nolint:err113 // Test error

	tests := []struct {
		name      string
		err       error
		retriable bool
	}{
		{"uppercase CONNECTION REFUSED", errUppercase, true},
		{"mixed case Connection Refused", errMixedCase, true},
		{"uppercase TIMEOUT", errUpperTimeout, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.IsRetriable(tt.err)
			if result != tt.retriable {
				t.Errorf("IsRetriable(%v) = %v, want %v", tt.err, result, tt.retriable)
			}
		})
	}
}

func TestClassifierFunc(t *testing.T) {
	// Test that ClassifierFunc properly wraps a function
	alwaysTrue := ClassifierFunc(func(_ error) bool { return true })
	alwaysFalse := ClassifierFunc(func(_ error) bool { return false })

	if !alwaysTrue.IsRetriable(errPermanent) {
		t.Error("ClassifierFunc(true) should return true")
	}

	if alwaysFalse.IsRetriable(errConnectionRefused) {
		t.Error("ClassifierFunc(false) should return false")
	}
}

func TestAlwaysRetry(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		retriable bool
	}{
		{"nil error", nil, false},
		{"permanent error", errPermanent, true},
		{"connection refused", errConnectionRefused, true},
		{"context.Canceled", context.Canceled, false}, // AlwaysRetry should still respect context cancellation
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AlwaysRetry.IsRetriable(tt.err)
			if result != tt.retriable {
				t.Errorf("AlwaysRetry.IsRetriable(%v) = %v, want %v", tt.err, result, tt.retriable)
			}
		})
	}
}

func TestNeverRetry(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"nil error", nil},
		{"permanent error", errPermanent},
		{"connection refused", errConnectionRefused},
		{"timeout", errIOTimeout},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if NeverRetry.IsRetriable(tt.err) {
				t.Errorf("NeverRetry.IsRetriable(%v) = true, want false", tt.err)
			}
		})
	}
}

func TestNewNetworkClassifier(t *testing.T) {
	c := NewNetworkClassifier()

	if len(c.Patterns) == 0 {
		t.Error("NewNetworkClassifier should have patterns")
	}

	if !c.CheckNetErrors {
		t.Error("NewNetworkClassifier should check net errors")
	}

	if !c.CheckDNSErrors {
		t.Error("NewNetworkClassifier should check DNS errors")
	}

	if !c.CheckIOErrors {
		t.Error("NewNetworkClassifier should check IO errors")
	}

	if !c.CheckHTTPStatus {
		t.Error("NewNetworkClassifier should check HTTP status")
	}
}

func TestNewCommandClassifier(t *testing.T) {
	c := NewCommandClassifier()

	// Should have both network and Go module patterns
	hasNetworkPattern := false
	hasGoPattern := false

	for _, p := range c.Patterns {
		if p == "connection refused" {
			hasNetworkPattern = true
		}
		if p == "go: downloading" {
			hasGoPattern = true
		}
	}

	if !hasNetworkPattern {
		t.Error("NewCommandClassifier should have network patterns")
	}

	if !hasGoPattern {
		t.Error("NewCommandClassifier should have Go module patterns")
	}

	if !c.CheckExitCodes {
		t.Error("NewCommandClassifier should check exit codes")
	}

	if len(c.RetriableExitCodes) == 0 {
		t.Error("NewCommandClassifier should have retriable exit codes")
	}
}

// mockNetError implements net.Error for testing
type mockNetError struct {
	timeout   bool
	temporary bool
}

func (e *mockNetError) Error() string   { return "mock net error" }
func (e *mockNetError) Timeout() bool   { return e.timeout }
func (e *mockNetError) Temporary() bool { return e.temporary }

func TestDefaultClassifier_NetError(t *testing.T) {
	classifier := NewNetworkClassifier()

	tests := []struct {
		name      string
		err       net.Error
		retriable bool
	}{
		{"timeout error", &mockNetError{timeout: true}, true},
		{"non-timeout error", &mockNetError{timeout: false}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.IsRetriable(tt.err)
			if result != tt.retriable {
				t.Errorf("IsRetriable(%v) = %v, want %v", tt.err, result, tt.retriable)
			}
		})
	}
}

func TestDefaultClassifier_ExitError(t *testing.T) {
	classifier := NewCommandClassifier()

	// Test that the classifier is properly configured for exit codes
	// Note: We can't easily create a real exec.ExitError without executing a command
	// but we can verify the classifier configuration
	t.Run("classifier configured for exit codes", func(t *testing.T) {
		if !classifier.CheckExitCodes {
			t.Error("CommandClassifier should check exit codes")
		}
		if len(classifier.RetriableExitCodes) == 0 {
			t.Error("CommandClassifier should have retriable exit codes")
		}

		// Verify exit code 124 (timeout) is retriable
		found124 := false
		for _, code := range classifier.RetriableExitCodes {
			if code == 124 {
				found124 = true
				break
			}
		}
		if !found124 {
			t.Error("CommandClassifier should include exit code 124 (timeout)")
		}
	})
}

func TestDefaultClassifier_DNSError(t *testing.T) {
	classifier := NewNetworkClassifier()

	tests := []struct {
		name      string
		dnsErr    *net.DNSError
		retriable bool
	}{
		{
			name:      "temporary DNS failure",
			dnsErr:    &net.DNSError{Err: "temporary failure", Name: "example.com"},
			retriable: true,
		},
		{
			name:      "no such host (matches pattern - retriable)",
			dnsErr:    &net.DNSError{Err: "no such host", Name: "invalid.test"},
			retriable: true, // "no such host" is in NetworkPatterns, so it matches
		},
		{
			name:      "server misbehaving (retriable)",
			dnsErr:    &net.DNSError{Err: "server misbehaving", Name: "example.com"},
			retriable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifier.IsRetriable(tt.dnsErr)
			if result != tt.retriable {
				t.Errorf("IsRetriable(%v) = %v, want %v", tt.dnsErr, result, tt.retriable)
			}
		})
	}

	// Test DNS error specifically without pattern matching
	classifierNoPat := &DefaultClassifier{
		Patterns:       []string{}, // No patterns
		CheckDNSErrors: true,
	}

	// DNS error with "server misbehaving" should be retriable via DNS check
	dnsRetriable := &net.DNSError{Err: "server misbehaving", Name: "example.com"}
	if !classifierNoPat.IsRetriable(dnsRetriable) {
		t.Error("DNS error with misbehaving should be retriable")
	}

	// DNS error with "no such host" should NOT be retriable without pattern matching
	dnsNoHost := &net.DNSError{Err: "no such host", Name: "invalid.test"}
	if classifierNoPat.IsRetriable(dnsNoHost) {
		t.Error("DNS error with 'no such host' should NOT be retriable without pattern")
	}
}

func TestDefaultClassifier_OpError(t *testing.T) {
	classifier := NewNetworkClassifier()

	// Test OpError is retriable
	opErr := &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: errors.New("connection refused"), //nolint:err113 // Test error
	}

	if !classifier.IsRetriable(opErr) {
		t.Error("OpError should be retriable")
	}

	// Verify CheckOpErrors flag works - OpError still matches via patterns
	classifierNoOp := &DefaultClassifier{
		Patterns:      NetworkPatterns,
		CheckOpErrors: false,
	}
	// OpError is still retriable due to "connection refused" pattern matching
	if !classifierNoOp.IsRetriable(opErr) {
		t.Error("OpError should still match pattern 'connection refused'")
	}
}

func TestDefaultClassifier_EmptyPatterns(t *testing.T) {
	classifier := &DefaultClassifier{
		Patterns: []string{}, // Empty patterns
	}

	// Should not match any patterns
	if classifier.IsRetriable(errConnectionRefused) {
		t.Error("Empty patterns should not match anything")
	}
}

func TestDefaultClassifier_HTTPStatus(t *testing.T) {
	classifierWithHTTP := &DefaultClassifier{
		CheckHTTPStatus: true,
	}
	classifierWithoutHTTP := &DefaultClassifier{
		CheckHTTPStatus: false,
	}

	errStatus503 := errors.New("status 503: service unavailable") //nolint:err113 // Test error

	// With HTTP check enabled
	if !classifierWithHTTP.IsRetriable(errStatus503) {
		t.Error("HTTP 5xx status should be retriable when CheckHTTPStatus is true")
	}

	// Without HTTP check enabled
	if classifierWithoutHTTP.IsRetriable(errStatus503) {
		t.Error("HTTP status should not be checked when CheckHTTPStatus is false")
	}
}
