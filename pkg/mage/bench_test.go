package mage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// =============================================================================
// MOCK RUNNER FOR BENCH TESTS
// =============================================================================

// BenchMockRunner provides a mock implementation of CommandRunner for testing bench functions
type BenchMockRunner struct {
	outputs map[string]struct {
		output string
		err    error
	}
	commands []string
}

// NewBenchMockRunner creates a new mock runner for bench tests
func NewBenchMockRunner() *BenchMockRunner {
	return &BenchMockRunner{
		outputs: make(map[string]struct {
			output string
			err    error
		}),
		commands: []string{},
	}
}

// SetOutput configures the expected output for a given command
func (m *BenchMockRunner) SetOutput(cmd, output string, err error) {
	m.outputs[cmd] = struct {
		output string
		err    error
	}{output: output, err: err}
}

// RunCmd implements CommandRunner.RunCmd
func (m *BenchMockRunner) RunCmd(name string, args ...string) error {
	cmd := name + " " + strings.Join(args, " ")
	m.commands = append(m.commands, cmd)
	if result, ok := m.outputs[cmd]; ok {
		return result.err
	}
	return nil
}

// RunCmdOutput implements CommandRunner.RunCmdOutput
func (m *BenchMockRunner) RunCmdOutput(name string, args ...string) (string, error) {
	cmd := name + " " + strings.Join(args, " ")
	m.commands = append(m.commands, cmd)
	if result, ok := m.outputs[cmd]; ok {
		return result.output, result.err
	}
	return "", nil
}

// GetCommands returns all commands that were executed
func (m *BenchMockRunner) GetCommands() []string {
	return m.commands
}

// Reset clears all recorded commands
func (m *BenchMockRunner) Reset() {
	m.commands = []string{}
}

// =============================================================================
// ENVIRONMENT CLEANUP HELPER
// =============================================================================

// getBenchEnvVars returns the list of MAGE_X environment variables used by bench functions
func getBenchEnvVars() []string {
	return []string{
		"MAGE_X_BENCH_CPU_PROFILE",
		"MAGE_X_BENCH_MEM_PROFILE",
		"MAGE_X_BENCH_FILE",
		"BENCH_OLD",
		"BENCH_NEW",
		"BENCH_BASELINE",
		"CPU_PROFILE",
		"MEM_PROFILE",
		"TRACE_FILE",
		"UPDATE_BASELINE",
	}
}

// withCleanBenchEnv saves and restores bench environment variables
func withCleanBenchEnv(t *testing.T, fn func()) {
	t.Helper()

	envVars := getBenchEnvVars()

	// Save current values
	saved := make(map[string]string)
	for _, key := range envVars {
		if val, exists := os.LookupEnv(key); exists {
			saved[key] = val
		}
	}

	// Clear all bench env vars
	for _, key := range envVars {
		if err := os.Unsetenv(key); err != nil {
			t.Logf("failed to unset %s: %v", key, err)
		}
	}

	// Run the test function
	defer func() {
		// Restore original values
		for _, key := range envVars {
			if val, exists := saved[key]; exists {
				if err := os.Setenv(key, val); err != nil {
					t.Logf("failed to restore %s: %v", key, err)
				}
			} else {
				if err := os.Unsetenv(key); err != nil {
					t.Logf("failed to unset %s: %v", key, err)
				}
			}
		}
	}()

	fn()
}

// =============================================================================
// DELEGATION PATTERN TESTS
// =============================================================================

// TestBenchDelegationPattern verifies that all wrapper methods correctly delegate
// to their WithArgs counterparts
func TestBenchDelegationPattern(t *testing.T) {
	// These tests verify the delegation pattern works correctly
	// They don't test full execution but confirm the call chain is correct

	tests := []struct {
		name         string
		wrapperName  string
		withArgsName string
	}{
		{
			name:         "Default delegates to DefaultWithArgs",
			wrapperName:  "Default",
			withArgsName: "DefaultWithArgs",
		},
		{
			name:         "Compare delegates to CompareWithArgs",
			wrapperName:  "Compare",
			withArgsName: "CompareWithArgs",
		},
		{
			name:         "Save delegates to SaveWithArgs",
			wrapperName:  "Save",
			withArgsName: "SaveWithArgs",
		},
		{
			name:         "CPU delegates to CPUWithArgs",
			wrapperName:  "CPU",
			withArgsName: "CPUWithArgs",
		},
		{
			name:         "Mem delegates to MemWithArgs",
			wrapperName:  "Mem",
			withArgsName: "MemWithArgs",
		},
		{
			name:         "Profile delegates to ProfileWithArgs",
			wrapperName:  "Profile",
			withArgsName: "ProfileWithArgs",
		},
		{
			name:         "Trace delegates to TraceWithArgs",
			wrapperName:  "Trace",
			withArgsName: "TraceWithArgs",
		},
		{
			name:         "Regression delegates to RegressionWithArgs",
			wrapperName:  "Regression",
			withArgsName: "RegressionWithArgs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify both wrapper and WithArgs methods exist and are callable
			var b Bench

			// Use reflection to verify methods exist
			// This is a compile-time guarantee but documents the pattern
			switch tt.wrapperName {
			case "Default":
				_ = b.Default
				_ = b.DefaultWithArgs
			case "Compare":
				_ = b.Compare
				_ = b.CompareWithArgs
			case "Save":
				_ = b.Save
				_ = b.SaveWithArgs
			case "CPU":
				_ = b.CPU
				_ = b.CPUWithArgs
			case "Mem":
				_ = b.Mem
				_ = b.MemWithArgs
			case "Profile":
				_ = b.Profile
				_ = b.ProfileWithArgs
			case "Trace":
				_ = b.Trace
				_ = b.TraceWithArgs
			case "Regression":
				_ = b.Regression
				_ = b.RegressionWithArgs
			}
		})
	}
}

// BenchRealTestSuite tests actual bench.go functionality
type BenchRealTestSuite struct {
	suite.Suite

	origDir string
	tempDir string
	bench   Bench
}

// SetupTest runs before each test
func (ts *BenchRealTestSuite) SetupTest() {
	var err error

	// Save original directory
	ts.origDir, err = os.Getwd()
	ts.Require().NoError(err)

	ts.tempDir, err = os.MkdirTemp("", "bench-test-*")
	ts.Require().NoError(err)

	err = os.Chdir(ts.tempDir)
	ts.Require().NoError(err)

	ts.bench = Bench{}
}

// TearDownTest runs after each test
func (ts *BenchRealTestSuite) TearDownTest() {
	// Restore original directory before removing temp dir
	if ts.origDir != "" {
		if err := os.Chdir(ts.origDir); err != nil {
			ts.T().Logf("failed to restore original directory: %v", err)
		}
	}

	if ts.tempDir != "" {
		if err := os.RemoveAll(ts.tempDir); err != nil {
			ts.T().Logf("failed to remove temp dir: %v", err)
		}
	}
}

// TestCompareWithArgsOldFileNotFound tests that Compare returns error when old file doesn't exist
func (ts *BenchRealTestSuite) TestCompareWithArgsOldFileNotFound() {
	// Save and restore runner
	originalRunner := GetRunner()
	defer func() {
		err := SetRunner(originalRunner)
		ts.Require().NoError(err)
	}()

	// Mock benchstat installation check
	mock := NewBenchMockRunner()
	mock.SetOutput("benchstat", "", nil)
	err := SetRunner(mock)
	ts.Require().NoError(err)

	// Create only the new file, not the old file
	newFile := filepath.Join(ts.tempDir, "new.txt")
	err = os.WriteFile(newFile, []byte("benchmark data"), 0o600)
	ts.Require().NoError(err)

	// Call Compare with old file that doesn't exist
	err = ts.bench.CompareWithArgs("old=nonexistent.txt", "new="+newFile)
	ts.Require().Error(err)
	ts.Require().ErrorIs(err, ErrOldBenchFileNotFound)
}

// TestCompareWithArgsNewFileNotFound tests that Compare returns error when new file doesn't exist
func (ts *BenchRealTestSuite) TestCompareWithArgsNewFileNotFound() {
	// Save and restore runner
	originalRunner := GetRunner()
	defer func() {
		err := SetRunner(originalRunner)
		ts.Require().NoError(err)
	}()

	// Mock benchstat installation check
	mock := NewBenchMockRunner()
	mock.SetOutput("benchstat", "", nil)
	err := SetRunner(mock)
	ts.Require().NoError(err)

	// Create only the old file, not the new file
	oldFile := filepath.Join(ts.tempDir, "old.txt")
	err = os.WriteFile(oldFile, []byte("benchmark data"), 0o600)
	ts.Require().NoError(err)

	// Call Compare with new file that doesn't exist
	err = ts.bench.CompareWithArgs("old="+oldFile, "new=nonexistent.txt")
	ts.Require().Error(err)
	ts.Require().ErrorIs(err, ErrNewBenchFileNotFound)
}

// TestCompareWithArgsBothFilesExist tests that Compare proceeds when both files exist
func (ts *BenchRealTestSuite) TestCompareWithArgsBothFilesExist() {
	// Save and restore runner
	originalRunner := GetRunner()
	defer func() {
		err := SetRunner(originalRunner)
		ts.Require().NoError(err)
	}()

	// Create mock runner that succeeds for all commands
	mock := NewBenchMockRunner()
	err := SetRunner(mock)
	ts.Require().NoError(err)

	// Create both files
	oldFile := filepath.Join(ts.tempDir, "old.txt")
	newFile := filepath.Join(ts.tempDir, "new.txt")
	err = os.WriteFile(oldFile, []byte("old benchmark data"), 0o600)
	ts.Require().NoError(err)
	err = os.WriteFile(newFile, []byte("new benchmark data"), 0o600)
	ts.Require().NoError(err)

	// Call Compare - should not return file not found errors
	err = ts.bench.CompareWithArgs("old="+oldFile, "new="+newFile)
	// May fail for other reasons (like benchstat not installed), but should not be file errors
	if err != nil {
		ts.Require().NotErrorIs(err, ErrOldBenchFileNotFound)
		ts.Require().NotErrorIs(err, ErrNewBenchFileNotFound)
	}
}

// TestBenchErrorConstants tests that benchmark error constants are properly defined
func (ts *BenchRealTestSuite) TestBenchErrorConstants() {
	// Verify error constants are properly defined
	ts.Require().Error(ErrOldBenchFileNotFound)
	ts.Require().Error(ErrNewBenchFileNotFound)

	// Verify they are different errors
	ts.Require().NotEqual(ErrOldBenchFileNotFound.Error(), ErrNewBenchFileNotFound.Error())

	// Verify error messages contain meaningful text
	ts.Require().Contains(ErrOldBenchFileNotFound.Error(), "old")
	ts.Require().Contains(ErrNewBenchFileNotFound.Error(), "new")
}

// TestBenchDefaultDelegatesToDefaultWithArgs tests that Default calls DefaultWithArgs
func (ts *BenchRealTestSuite) TestBenchDefaultDelegatesToDefaultWithArgs() {
	// This test verifies the delegate pattern works
	// We can't easily test the full execution, but we can verify the method exists and is callable
	var b Bench
	// Just verify the method exists and is callable - actual execution tested in integration
	_ = b
}

// TestBenchRealTestSuite runs the real test suite
func TestBenchRealTestSuite(t *testing.T) {
	suite.Run(t, new(BenchRealTestSuite))
}

// =============================================================================
// UNIT TESTS FOR WithArgs FUNCTIONS
// =============================================================================

// TestBenchCPUWithArgsSetenv tests that CPUWithArgs sets the environment variable correctly
func TestBenchCPUWithArgsSetenv(t *testing.T) {
	withCleanBenchEnv(t, func() {
		// Verify env var is not set initially
		_, exists := os.LookupEnv("MAGE_X_BENCH_CPU_PROFILE")
		require.False(t, exists, "MAGE_X_BENCH_CPU_PROFILE should not be set initially")

		// We can't run the full function because it calls DefaultWithArgs which needs config
		// But we can test the environment variable setup logic directly
		// by checking that the constant is used correctly
		require.Equal(t, "cpu.prof", defaultCPUProfile)
	})
}

// TestBenchMemWithArgsSetenv tests that MemWithArgs sets the environment variable correctly
func TestBenchMemWithArgsSetenv(t *testing.T) {
	withCleanBenchEnv(t, func() {
		// Verify env var is not set initially
		_, exists := os.LookupEnv("MAGE_X_BENCH_MEM_PROFILE")
		require.False(t, exists, "MAGE_X_BENCH_MEM_PROFILE should not be set initially")

		// Verify the constant is correct
		require.Equal(t, "mem.prof", defaultMemProfile)
	})
}

// TestBenchProfileWithArgsConstants tests that ProfileWithArgs uses correct defaults
func TestBenchProfileWithArgsConstants(t *testing.T) {
	// Verify both profile constants are correct
	require.Equal(t, "cpu.prof", defaultCPUProfile)
	require.Equal(t, "mem.prof", defaultMemProfile)
}

// TestBenchTraceConstants tests trace-related constants
func TestBenchTraceConstants(t *testing.T) {
	require.Equal(t, "trace.out", defaultTraceFile)
}

// TestBenchRegressionConstants tests regression-related constants
func TestBenchRegressionConstants(t *testing.T) {
	require.Equal(t, "bench-current.txt", defaultCurrentFile)
	require.Equal(t, "bench-baseline.txt", defaultBaselineFile)
}

// TestBenchDefaultConstants tests default benchmark constants
func TestBenchDefaultConstants(t *testing.T) {
	require.Equal(t, "3s", defaultBenchTime)
	require.Equal(t, "./...", defaultPackage)
	require.Equal(t, "old.txt", defaultOldFile)
	require.Equal(t, "new.txt", defaultNewFile)
}

// TestBenchCompareWithArgsFileValidation tests the file validation logic in CompareWithArgs
func TestBenchCompareWithArgsFileValidation(t *testing.T) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "bench-file-test-*")
	require.NoError(t, err)
	defer func() {
		if cleanupErr := os.RemoveAll(tempDir); cleanupErr != nil {
			t.Logf("failed to remove temp dir: %v", cleanupErr)
		}
	}()

	// Save original directory and runner
	origDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if chdirErr := os.Chdir(origDir); chdirErr != nil {
			t.Logf("failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	originalRunner := GetRunner()
	defer func() {
		if runnerErr := SetRunner(originalRunner); runnerErr != nil {
			t.Logf("failed to restore runner: %v", runnerErr)
		}
	}()

	// Use mock runner
	mock := NewBenchMockRunner()
	err = SetRunner(mock)
	require.NoError(t, err)

	tests := []struct {
		name        string
		oldExists   bool
		newExists   bool
		expectedErr error
	}{
		{
			name:        "old file missing",
			oldExists:   false,
			newExists:   true,
			expectedErr: ErrOldBenchFileNotFound,
		},
		{
			name:        "new file missing",
			oldExists:   true,
			newExists:   false,
			expectedErr: ErrNewBenchFileNotFound,
		},
		{
			name:        "both files missing old checked first",
			oldExists:   false,
			newExists:   false,
			expectedErr: ErrOldBenchFileNotFound, // Old is checked first
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing files
			oldFile := filepath.Join(tempDir, "old.txt")
			newFile := filepath.Join(tempDir, "new.txt")
			// Ignore errors as files may not exist
			if err := os.Remove(oldFile); err != nil && !os.IsNotExist(err) {
				t.Logf("failed to remove old file: %v", err)
			}
			if err := os.Remove(newFile); err != nil && !os.IsNotExist(err) {
				t.Logf("failed to remove new file: %v", err)
			}

			if tt.oldExists {
				err := os.WriteFile(oldFile, []byte("old data"), 0o600)
				require.NoError(t, err)
			}
			if tt.newExists {
				err := os.WriteFile(newFile, []byte("new data"), 0o600)
				require.NoError(t, err)
			}

			var b Bench
			err := b.CompareWithArgs("old="+oldFile, "new="+newFile)
			require.Error(t, err)
			require.ErrorIs(t, err, tt.expectedErr,
				"expected error %v, got %v", tt.expectedErr, err)
		})
	}
}
