package testhelpers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// RunCommand runs a command and returns output
func RunCommand(t *testing.T, name string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.CommandContext(context.Background(), name, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%w\nstderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// RunCommandWithInput runs a command with stdin input
func RunCommandWithInput(t *testing.T, input, name string, args ...string) (string, error) {
	t.Helper()

	cmd := exec.CommandContext(context.Background(), name, args...)
	cmd.Stdin = strings.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%w\nstderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// RequireCommand skips the test if a command is not available
func RequireCommand(t *testing.T, name string) {
	t.Helper()

	_, err := exec.LookPath(name)
	if err != nil {
		t.Skipf("Required command not found: %s", name)
	}
}

// RequireEnv skips the test if an environment variable is not set
func RequireEnv(t *testing.T, key string) string {
	t.Helper()

	value := os.Getenv(key)
	if value == "" {
		t.Skipf("Required environment variable not set: %s", key)
	}

	return value
}

// RequireNetwork skips the test if network is not available
func RequireNetwork(t *testing.T) {
	t.Helper()

	// Try to resolve a well-known domain
	cmd := exec.CommandContext(context.Background(), "ping", "-c", "1", "-W", "1", "google.com")
	if err := cmd.Run(); err != nil {
		t.Skip("Network not available")
	}
}

// RequireDocker skips the test if Docker is not available
func RequireDocker(t *testing.T) {
	t.Helper()

	RequireCommand(t, "docker")

	// Check if Docker daemon is running
	cmd := exec.CommandContext(context.Background(), "docker", "info")
	if err := cmd.Run(); err != nil {
		t.Skip("Docker daemon not running")
	}
}

// RequireGit skips the test if Git is not available
func RequireGit(t *testing.T) {
	t.Helper()
	RequireCommand(t, "git")
}

// AssertContains asserts that a string contains a substring
func AssertContains(t *testing.T, str, substr string) {
	t.Helper()

	if !strings.Contains(str, substr) {
		t.Errorf("Expected string to contain %q, but it didn't:\n%s", substr, str)
	}
}

// AssertNotContains asserts that a string does not contain a substring
func AssertNotContains(t *testing.T, str, substr string) {
	t.Helper()

	if strings.Contains(str, substr) {
		t.Errorf("Expected string to not contain %q, but it did:\n%s", substr, str)
	}
}

// AssertEquals asserts that two values are equal
func AssertEquals(t *testing.T, expected, actual interface{}) {
	t.Helper()

	if expected != actual {
		t.Errorf("Expected %v, got %v", expected, actual)
	}
}

// AssertNotEquals asserts that two values are not equal
func AssertNotEquals(t *testing.T, expected, actual interface{}) {
	t.Helper()

	if expected == actual {
		t.Errorf("Expected values to be different, but both were %v", actual)
	}
}

// AssertTrue asserts that a value is true
func AssertTrue(t *testing.T, value bool, msg ...string) {
	t.Helper()

	if !value {
		message := "Expected true, got false"
		if len(msg) > 0 {
			message = msg[0]
		}
		t.Error(message)
	}
}

// AssertFalse asserts that a value is false
func AssertFalse(t *testing.T, value bool, msg ...string) {
	t.Helper()

	if value {
		message := "Expected false, got true"
		if len(msg) > 0 {
			message = msg[0]
		}
		t.Error(message)
	}
}

// AssertNil asserts that a value is nil
func AssertNil(t *testing.T, value interface{}) {
	t.Helper()

	if value != nil {
		t.Errorf("Expected nil, got %v", value)
	}
}

// AssertNotNil asserts that a value is not nil
func AssertNotNil(t *testing.T, value interface{}) {
	t.Helper()

	if value == nil {
		t.Error("Expected non-nil value, got nil")
	}
}

// EventuallyTrue waits for a condition to become true
func EventuallyTrue(t *testing.T, fn func() bool, timeout time.Duration, msg string) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	interval := timeout / 100
	if interval < 10*time.Millisecond {
		interval = 10 * time.Millisecond
	}

	for time.Now().Before(deadline) {
		if fn() {
			return
		}
		time.Sleep(interval)
	}

	t.Errorf("Condition never became true: %s", msg)
}

// EventuallyEquals waits for a value to equal expected
func EventuallyEquals(t *testing.T, fn func() interface{}, expected interface{}, timeout time.Duration) {
	t.Helper()

	EventuallyTrue(t, func() bool {
		return fn() == expected
	}, timeout, fmt.Sprintf("Value never became %v", expected))
}

// Retry retries a function until it succeeds or times out
func Retry(t *testing.T, fn func() error, attempts int, delay time.Duration) error {
	t.Helper()

	var lastErr error
	for i := 0; i < attempts; i++ {
		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err
		if i < attempts-1 {
			time.Sleep(delay)
		}
	}

	return fmt.Errorf("failed after %d attempts: %w", attempts, lastErr)
}

// SkipIfShort skips the test if -short flag is set
func SkipIfShort(t *testing.T) {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}
}

// SkipIfCI skips the test if running in CI
func SkipIfCI(t *testing.T) {
	t.Helper()

	ciVars := []string{"CI", "CONTINUOUS_INTEGRATION", "GITHUB_ACTIONS", "JENKINS", "TRAVIS"}
	for _, v := range ciVars {
		if os.Getenv(v) != "" {
			t.Skip("Skipping test in CI environment")
		}
	}
}

// RunParallel runs a test in parallel
func RunParallel(t *testing.T) {
	t.Helper()
	t.Parallel()
}

// Benchmark provides a simple benchmark helper
func Benchmark(t *testing.T, name string, fn func()) {
	t.Helper()

	start := time.Now()
	fn()
	duration := time.Since(start)

	t.Logf("Benchmark %s: %v", name, duration)
}

// MeasureTime measures execution time
func MeasureTime(t *testing.T, name string) func() {
	t.Helper()

	start := time.Now()
	return func() {
		t.Logf("%s took %v", name, time.Since(start))
	}
}

// CaptureLog captures log output during a function
func CaptureLog(t *testing.T, fn func()) string {
	t.Helper()

	// This is simplified - in real implementation you'd capture actual log output
	var buf bytes.Buffer

	// Save original stderr
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stderr = w

	// Run function
	fn()

	// Restore stderr
	if err := w.Close(); err != nil {
		t.Logf("Warning: failed to close pipe writer: %v", err)
	}
	os.Stderr = origStderr

	// Read captured output
	if _, err := buf.ReadFrom(r); err != nil {
		t.Logf("Warning: failed to read from pipe: %v", err)
	}

	return buf.String()
}

// Golden manages golden files for testing
type Golden struct {
	t      *testing.T
	dir    string
	update bool
}

// NewGolden creates a new golden file manager
func NewGolden(t *testing.T, dir string) *Golden {
	t.Helper()

	if dir == "" {
		dir = "testdata/golden"
	}

	update := os.Getenv("UPDATE_GOLDEN") == "true"

	return &Golden{
		t:      t,
		dir:    dir,
		update: update,
	}
}

// Check checks output against a golden file
func (g *Golden) Check(name string, actual []byte) {
	g.t.Helper()

	// Validate name to prevent directory traversal
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		g.t.Fatalf("Invalid golden file name: %s", name)
	}

	goldenPath := filepath.Join(g.dir, name+".golden")

	if g.update {
		// Update golden file
		dir := filepath.Dir(goldenPath)
		if err := os.MkdirAll(dir, 0o750); err != nil {
			g.t.Fatalf("Failed to create golden directory: %v", err)
		}

		if err := os.WriteFile(goldenPath, actual, 0o600); err != nil {
			g.t.Fatalf("Failed to update golden file: %v", err)
		}

		g.t.Logf("Updated golden file: %s", goldenPath)
		return
	}

	// Read golden file
	expected, err := os.ReadFile(goldenPath) // #nosec G304 -- controlled test golden file path
	if err != nil {
		g.t.Fatalf("Failed to read golden file: %v", err)
	}

	// Compare
	if !bytes.Equal(expected, actual) {
		g.t.Errorf("Output doesn't match golden file %s\nExpected:\n%s\nActual:\n%s",
			goldenPath, expected, actual)
	}
}

// CheckString checks a string against a golden file
func (g *Golden) CheckString(name, actual string) {
	g.Check(name, []byte(actual))
}

// DataProvider provides test data
type DataProvider struct {
	t *testing.T
}

// NewDataProvider creates a new data provider
func NewDataProvider(t *testing.T) *DataProvider {
	return &DataProvider{t: t}
}

// Strings returns test strings
func (dp *DataProvider) Strings() []string {
	return []string{
		"",
		"hello",
		"hello world",
		"special chars: !@#$%^&*()",
		"unicode: 你好世界 🌍",
		"multiline\nstring\nwith\nbreaks",
		strings.Repeat("long ", 100),
	}
}

// Ints returns test integers
func (dp *DataProvider) Ints() []int {
	return []int{0, 1, -1, 42, 100, -100, 1000000}
}

// Bools returns test booleans
func (dp *DataProvider) Bools() []bool {
	return []bool{true, false}
}

// Errors returns test errors
func (dp *DataProvider) Errors() []error {
	return []error{
		nil,
		fmt.Errorf("simple error"),
		fmt.Errorf("error with %s", "formatting"),
		fmt.Errorf("wrapped: %w", fmt.Errorf("inner error")),
	}
}

// TestCase represents a parameterized test case
type TestCase struct {
	Name  string
	Input interface{}
	Want  interface{}
	Error error
}

// RunTestCases runs parameterized test cases
func RunTestCases(t *testing.T, cases []TestCase, fn func(tc TestCase)) {
	t.Helper()

	for _, tc := range cases {
		t.Run(tc.Name, func(_ *testing.T) {
			fn(tc)
		})
	}
}
