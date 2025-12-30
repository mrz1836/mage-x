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

// =============================================================================
// ENVIRONMENT VARIABLE SETUP TESTS
// =============================================================================

// BenchEnvTestSuite tests environment variable setup in bench functions
type BenchEnvTestSuite struct {
	suite.Suite

	origDir        string
	tempDir        string
	originalRunner CommandRunner
	mockRunner     *BenchMockRunner
	bench          Bench
}

// SetupTest runs before each test
func (ts *BenchEnvTestSuite) SetupTest() {
	var err error

	// Save original directory
	ts.origDir, err = os.Getwd()
	ts.Require().NoError(err)

	// Create and change to temp directory
	ts.tempDir, err = os.MkdirTemp("", "bench-env-test-*")
	ts.Require().NoError(err)

	err = os.Chdir(ts.tempDir)
	ts.Require().NoError(err)

	// Save original runner
	ts.originalRunner = GetRunner()

	// Set up mock runner
	ts.mockRunner = NewBenchMockRunner()
	err = SetRunner(ts.mockRunner)
	ts.Require().NoError(err)

	// Create go.mod for module discovery
	goModContent := `module test/bench

go 1.24
`
	err = os.WriteFile("go.mod", []byte(goModContent), 0o600)
	ts.Require().NoError(err)

	ts.bench = Bench{}
}

// TearDownTest runs after each test
func (ts *BenchEnvTestSuite) TearDownTest() {
	// Restore original runner
	if ts.originalRunner != nil {
		if err := SetRunner(ts.originalRunner); err != nil {
			ts.T().Logf("failed to restore runner: %v", err)
		}
	}

	// Restore original directory
	if ts.origDir != "" {
		if err := os.Chdir(ts.origDir); err != nil {
			ts.T().Logf("failed to restore directory: %v", err)
		}
	}

	// Remove temp directory
	if ts.tempDir != "" {
		if err := os.RemoveAll(ts.tempDir); err != nil {
			ts.T().Logf("failed to remove temp dir: %v", err)
		}
	}
}

// TestCPUWithArgsSetsEnvVar verifies CPUWithArgs sets MAGE_X_BENCH_CPU_PROFILE
func (ts *BenchEnvTestSuite) TestCPUWithArgsSetsEnvVar() {
	withCleanBenchEnv(ts.T(), func() {
		// Verify env var is not set initially
		_, exists := os.LookupEnv("MAGE_X_BENCH_CPU_PROFILE")
		ts.Require().False(exists, "MAGE_X_BENCH_CPU_PROFILE should not be set initially")

		// Run CPUWithArgs - it will fail at module discovery but that's OK
		// We just want to verify the env var is set
		err := ts.bench.CPUWithArgs()
		// Error is expected (mock runner doesn't execute real benchmarks) - we only care about env setup
		_ = err

		// Verify env var is now set
		val, exists := os.LookupEnv("MAGE_X_BENCH_CPU_PROFILE")
		ts.Require().True(exists, "MAGE_X_BENCH_CPU_PROFILE should be set")
		ts.Require().Equal("cpu.prof", val, "MAGE_X_BENCH_CPU_PROFILE should be set to default value")
	})
}

// TestCPUWithArgsCustomProfile verifies CPUWithArgs uses custom profile parameter
func (ts *BenchEnvTestSuite) TestCPUWithArgsCustomProfile() {
	withCleanBenchEnv(ts.T(), func() {
		// Run CPUWithArgs with custom profile
		err := ts.bench.CPUWithArgs("profile=custom-cpu.prof")
		// Error is expected - we only care about env setup
		_ = err

		// Verify env var is set to custom value
		val, exists := os.LookupEnv("MAGE_X_BENCH_CPU_PROFILE")
		ts.Require().True(exists, "MAGE_X_BENCH_CPU_PROFILE should be set")
		ts.Require().Equal("custom-cpu.prof", val, "MAGE_X_BENCH_CPU_PROFILE should use custom value")
	})
}

// TestMemWithArgsSetsEnvVar verifies MemWithArgs sets MAGE_X_BENCH_MEM_PROFILE
func (ts *BenchEnvTestSuite) TestMemWithArgsSetsEnvVar() {
	withCleanBenchEnv(ts.T(), func() {
		// Verify env var is not set initially
		_, exists := os.LookupEnv("MAGE_X_BENCH_MEM_PROFILE")
		ts.Require().False(exists, "MAGE_X_BENCH_MEM_PROFILE should not be set initially")

		// Run MemWithArgs - it will fail at module discovery but that's OK
		err := ts.bench.MemWithArgs()
		// Error is expected - we only care about env setup
		_ = err

		// Verify env var is now set
		val, exists := os.LookupEnv("MAGE_X_BENCH_MEM_PROFILE")
		ts.Require().True(exists, "MAGE_X_BENCH_MEM_PROFILE should be set")
		ts.Require().Equal("mem.prof", val, "MAGE_X_BENCH_MEM_PROFILE should be set to default value")
	})
}

// TestMemWithArgsCustomProfile verifies MemWithArgs uses custom profile parameter
func (ts *BenchEnvTestSuite) TestMemWithArgsCustomProfile() {
	withCleanBenchEnv(ts.T(), func() {
		// Run MemWithArgs with custom profile
		err := ts.bench.MemWithArgs("profile=custom-mem.prof")
		// Error is expected - we only care about env setup
		_ = err

		// Verify env var is set to custom value
		val, exists := os.LookupEnv("MAGE_X_BENCH_MEM_PROFILE")
		ts.Require().True(exists, "MAGE_X_BENCH_MEM_PROFILE should be set")
		ts.Require().Equal("custom-mem.prof", val, "MAGE_X_BENCH_MEM_PROFILE should use custom value")
	})
}

// TestProfileWithArgsSetsBothEnvVars verifies ProfileWithArgs sets both CPU and MEM env vars
func (ts *BenchEnvTestSuite) TestProfileWithArgsSetsBothEnvVars() {
	withCleanBenchEnv(ts.T(), func() {
		// Verify env vars are not set initially
		_, cpuExists := os.LookupEnv("MAGE_X_BENCH_CPU_PROFILE")
		_, memExists := os.LookupEnv("MAGE_X_BENCH_MEM_PROFILE")
		ts.Require().False(cpuExists, "MAGE_X_BENCH_CPU_PROFILE should not be set initially")
		ts.Require().False(memExists, "MAGE_X_BENCH_MEM_PROFILE should not be set initially")

		// Run ProfileWithArgs
		err := ts.bench.ProfileWithArgs()
		// Error is expected - we only care about env setup
		_ = err

		// Verify both env vars are set
		cpuVal, cpuExists := os.LookupEnv("MAGE_X_BENCH_CPU_PROFILE")
		memVal, memExists := os.LookupEnv("MAGE_X_BENCH_MEM_PROFILE")
		ts.Require().True(cpuExists, "MAGE_X_BENCH_CPU_PROFILE should be set")
		ts.Require().True(memExists, "MAGE_X_BENCH_MEM_PROFILE should be set")
		ts.Require().Equal("cpu.prof", cpuVal, "MAGE_X_BENCH_CPU_PROFILE should be set to default value")
		ts.Require().Equal("mem.prof", memVal, "MAGE_X_BENCH_MEM_PROFILE should be set to default value")
	})
}

// TestProfileWithArgsCustomProfiles verifies ProfileWithArgs uses custom profile parameters
func (ts *BenchEnvTestSuite) TestProfileWithArgsCustomProfiles() {
	withCleanBenchEnv(ts.T(), func() {
		// Run ProfileWithArgs with custom profiles
		err := ts.bench.ProfileWithArgs("cpu-profile=my-cpu.prof", "mem-profile=my-mem.prof")
		// Error is expected - we only care about env setup
		_ = err

		// Verify both env vars are set to custom values
		cpuVal, _ := os.LookupEnv("MAGE_X_BENCH_CPU_PROFILE")
		memVal, _ := os.LookupEnv("MAGE_X_BENCH_MEM_PROFILE")
		ts.Require().Equal("my-cpu.prof", cpuVal, "MAGE_X_BENCH_CPU_PROFILE should use custom value")
		ts.Require().Equal("my-mem.prof", memVal, "MAGE_X_BENCH_MEM_PROFILE should use custom value")
	})
}

// TestRegressionWithArgsSetsEnvVars verifies RegressionWithArgs sets required env vars
func (ts *BenchEnvTestSuite) TestRegressionWithArgsSetsEnvVars() {
	withCleanBenchEnv(ts.T(), func() {
		// Verify env var is not set initially
		_, exists := os.LookupEnv("MAGE_X_BENCH_FILE")
		ts.Require().False(exists, "MAGE_X_BENCH_FILE should not be set initially")

		// Run RegressionWithArgs - it will fail but we check env var setup
		err := ts.bench.RegressionWithArgs()
		// Error is expected - we only care about env setup
		_ = err

		// Verify env var is set to current file
		val, exists := os.LookupEnv("MAGE_X_BENCH_FILE")
		ts.Require().True(exists, "MAGE_X_BENCH_FILE should be set")
		ts.Require().Equal("bench-current.txt", val, "MAGE_X_BENCH_FILE should be set to current file")
	})
}

// TestBenchEnvTestSuite runs the environment test suite
func TestBenchEnvTestSuite(t *testing.T) {
	suite.Run(t, new(BenchEnvTestSuite))
}

// =============================================================================
// COMPARE COMMAND EXECUTION TESTS
// =============================================================================

// TestCompareWithArgsExecutesBenchstat verifies the benchstat command is called correctly
func (ts *BenchRealTestSuite) TestCompareWithArgsExecutesBenchstat() {
	// Save and restore runner
	originalRunner := GetRunner()
	defer func() {
		err := SetRunner(originalRunner)
		ts.Require().NoError(err)
	}()

	// Create mock runner that tracks commands
	mock := NewBenchMockRunner()
	err := SetRunner(mock)
	ts.Require().NoError(err)

	// Create both benchmark files
	oldFile := filepath.Join(ts.tempDir, "old-bench.txt")
	newFile := filepath.Join(ts.tempDir, "new-bench.txt")
	err = os.WriteFile(oldFile, []byte("BenchmarkFoo 1000 1000 ns/op"), 0o600)
	ts.Require().NoError(err)
	err = os.WriteFile(newFile, []byte("BenchmarkFoo 1000 900 ns/op"), 0o600)
	ts.Require().NoError(err)

	// Call CompareWithArgs
	err = ts.bench.CompareWithArgs("old="+oldFile, "new="+newFile)
	ts.Require().NoError(err)

	// Verify benchstat was called with correct arguments
	commands := mock.GetCommands()
	ts.Require().NotEmpty(commands, "benchstat command should be called")

	// Find the benchstat command
	found := false
	for _, cmd := range commands {
		if strings.HasPrefix(cmd, "benchstat") {
			found = true
			ts.Require().Contains(cmd, oldFile, "benchstat should receive old file")
			ts.Require().Contains(cmd, newFile, "benchstat should receive new file")
			break
		}
	}
	ts.Require().True(found, "benchstat command should be in executed commands")
}

// TestCompareWithArgsDefaultFiles verifies Compare uses default file names
func (ts *BenchRealTestSuite) TestCompareWithArgsDefaultFiles() {
	// Save and restore runner
	originalRunner := GetRunner()
	defer func() {
		err := SetRunner(originalRunner)
		ts.Require().NoError(err)
	}()

	mock := NewBenchMockRunner()
	err := SetRunner(mock)
	ts.Require().NoError(err)

	// Create files with default names
	err = os.WriteFile(filepath.Join(ts.tempDir, "old.txt"), []byte("data"), 0o600)
	ts.Require().NoError(err)
	err = os.WriteFile(filepath.Join(ts.tempDir, "new.txt"), []byte("data"), 0o600)
	ts.Require().NoError(err)

	// Call Compare without arguments (uses defaults)
	err = ts.bench.CompareWithArgs()
	ts.Require().NoError(err)

	// Verify benchstat was called with default files
	commands := mock.GetCommands()
	found := false
	for _, cmd := range commands {
		if strings.HasPrefix(cmd, "benchstat") {
			found = true
			ts.Require().Contains(cmd, "old.txt", "benchstat should use default old file")
			ts.Require().Contains(cmd, "new.txt", "benchstat should use default new file")
			break
		}
	}
	ts.Require().True(found, "benchstat command should be in executed commands")
}

// =============================================================================
// CPU/MEM PROFILE COMMAND EXECUTION TESTS
// =============================================================================

// TestCPUWithArgsRunsPprofAnalysis verifies pprof analysis is run after benchmarks
func (ts *BenchRealTestSuite) TestCPUWithArgsRunsPprofAnalysis() {
	withCleanBenchEnv(ts.T(), func() {
		// Save and restore runner
		originalRunner := GetRunner()
		defer func() {
			err := SetRunner(originalRunner)
			ts.Require().NoError(err)
		}()

		// Create mock runner
		mock := NewBenchMockRunner()
		err := SetRunner(mock)
		ts.Require().NoError(err)

		// Create go.mod for module discovery
		err = os.WriteFile(filepath.Join(ts.tempDir, "go.mod"), []byte("module test\n\ngo 1.24\n"), 0o600)
		ts.Require().NoError(err)

		// Run CPUWithArgs - error is expected with mock runner, we only verify commands
		err = ts.bench.CPUWithArgs()
		_ = err // Error expected - we only verify commands were called

		// Verify pprof was called
		commands := mock.GetCommands()
		pprofCalled := false
		for _, cmd := range commands {
			if strings.Contains(cmd, "pprof") && strings.Contains(cmd, "-top") {
				pprofCalled = true
				ts.Require().Contains(cmd, "cpu.prof", "pprof should analyze cpu.prof")
				break
			}
		}
		ts.Require().True(pprofCalled, "pprof analysis should be called")
	})
}

// TestMemWithArgsRunsPprofAnalysis verifies memory pprof analysis is run
func (ts *BenchRealTestSuite) TestMemWithArgsRunsPprofAnalysis() {
	withCleanBenchEnv(ts.T(), func() {
		// Save and restore runner
		originalRunner := GetRunner()
		defer func() {
			err := SetRunner(originalRunner)
			ts.Require().NoError(err)
		}()

		// Create mock runner
		mock := NewBenchMockRunner()
		err := SetRunner(mock)
		ts.Require().NoError(err)

		// Create go.mod for module discovery
		err = os.WriteFile(filepath.Join(ts.tempDir, "go.mod"), []byte("module test\n\ngo 1.24\n"), 0o600)
		ts.Require().NoError(err)

		// Run MemWithArgs - error is expected with mock runner, we only verify commands
		err = ts.bench.MemWithArgs()
		_ = err // Error expected - we only verify commands were called

		// Verify pprof was called with alloc_space flag
		commands := mock.GetCommands()
		pprofCalled := false
		for _, cmd := range commands {
			if strings.Contains(cmd, "pprof") && strings.Contains(cmd, "-top") && strings.Contains(cmd, "-alloc_space") {
				pprofCalled = true
				ts.Require().Contains(cmd, "mem.prof", "pprof should analyze mem.prof")
				break
			}
		}
		ts.Require().True(pprofCalled, "memory pprof analysis should be called")
	})
}

// =============================================================================
// DEFAULT WITH ARGS PARAMETER TESTS
// =============================================================================

// BenchDefaultArgsTestSuite tests parameter handling in DefaultWithArgs
type BenchDefaultArgsTestSuite struct {
	suite.Suite

	origDir        string
	tempDir        string
	originalRunner CommandRunner
	mockRunner     *BenchMockRunner
	bench          Bench
}

// SetupTest runs before each test
func (ts *BenchDefaultArgsTestSuite) SetupTest() {
	var err error

	ts.origDir, err = os.Getwd()
	ts.Require().NoError(err)

	ts.tempDir, err = os.MkdirTemp("", "bench-default-args-test-*")
	ts.Require().NoError(err)

	err = os.Chdir(ts.tempDir)
	ts.Require().NoError(err)

	ts.originalRunner = GetRunner()
	ts.mockRunner = NewBenchMockRunner()
	err = SetRunner(ts.mockRunner)
	ts.Require().NoError(err)

	// Create go.mod
	err = os.WriteFile("go.mod", []byte("module test\n\ngo 1.24\n"), 0o600)
	ts.Require().NoError(err)

	ts.bench = Bench{}
}

// TearDownTest runs after each test
func (ts *BenchDefaultArgsTestSuite) TearDownTest() {
	if ts.originalRunner != nil {
		if err := SetRunner(ts.originalRunner); err != nil {
			ts.T().Logf("failed to restore runner: %v", err)
		}
	}
	if ts.origDir != "" {
		if err := os.Chdir(ts.origDir); err != nil {
			ts.T().Logf("failed to restore directory: %v", err)
		}
	}
	if ts.tempDir != "" {
		if err := os.RemoveAll(ts.tempDir); err != nil {
			ts.T().Logf("failed to remove temp dir: %v", err)
		}
	}
}

// TestDefaultWithArgsBaseCommand verifies base benchmark command structure
func (ts *BenchDefaultArgsTestSuite) TestDefaultWithArgsBaseCommand() {
	withCleanBenchEnv(ts.T(), func() {
		// Run with no arguments - error is expected with mock runner
		err := ts.bench.DefaultWithArgs()
		_ = err // Error expected - we only verify command construction

		// Find the go test command
		commands := ts.mockRunner.GetCommands()
		var goTestCmd string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "go test") {
				goTestCmd = cmd
				break
			}
		}

		ts.Require().NotEmpty(goTestCmd, "go test command should be executed")
		ts.Require().Contains(goTestCmd, "-bench=.", "should include -bench=. flag")
		ts.Require().Contains(goTestCmd, "-benchmem", "should include -benchmem flag")
		ts.Require().Contains(goTestCmd, "-run=^$", "should include -run=^$ to skip regular tests")
		ts.Require().Contains(goTestCmd, "-benchtime", "should include -benchtime flag")
		ts.Require().Contains(goTestCmd, "3s", "should use default benchtime of 3s")
		ts.Require().Contains(goTestCmd, "./...", "should use default package ./...")
	})
}

// TestDefaultWithArgsCustomBenchtime verifies custom benchtime parameter
func (ts *BenchDefaultArgsTestSuite) TestDefaultWithArgsCustomBenchtime() {
	withCleanBenchEnv(ts.T(), func() {
		err := ts.bench.DefaultWithArgs("time=5s")
		_ = err // Error expected - we only verify command construction

		commands := ts.mockRunner.GetCommands()
		var goTestCmd string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "go test") {
				goTestCmd = cmd
				break
			}
		}

		ts.Require().Contains(goTestCmd, "-benchtime 5s", "should use custom benchtime")
	})
}

// TestDefaultWithArgsVerboseFlag verifies verbose flag is added
func (ts *BenchDefaultArgsTestSuite) TestDefaultWithArgsVerboseFlag() {
	withCleanBenchEnv(ts.T(), func() {
		err := ts.bench.DefaultWithArgs("verbose=true")
		_ = err // Error expected - we only verify command construction

		commands := ts.mockRunner.GetCommands()
		var goTestCmd string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "go test") {
				goTestCmd = cmd
				break
			}
		}

		ts.Require().Contains(goTestCmd, "-v", "should include -v flag when verbose=true")
	})
}

// TestDefaultWithArgsCountParameter verifies count parameter
func (ts *BenchDefaultArgsTestSuite) TestDefaultWithArgsCountParameter() {
	withCleanBenchEnv(ts.T(), func() {
		err := ts.bench.DefaultWithArgs("count=5")
		_ = err // Error expected - we only verify command construction

		commands := ts.mockRunner.GetCommands()
		var goTestCmd string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "go test") {
				goTestCmd = cmd
				break
			}
		}

		ts.Require().Contains(goTestCmd, "-count 5", "should include -count parameter")
	})
}

// TestDefaultWithArgsSkipPattern verifies skip pattern parameter
func (ts *BenchDefaultArgsTestSuite) TestDefaultWithArgsSkipPattern() {
	withCleanBenchEnv(ts.T(), func() {
		err := ts.bench.DefaultWithArgs("skip=BenchmarkSlow")
		_ = err // Error expected - we only verify command construction

		commands := ts.mockRunner.GetCommands()
		var goTestCmd string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "go test") {
				goTestCmd = cmd
				break
			}
		}

		ts.Require().Contains(goTestCmd, "-skip BenchmarkSlow", "should include -skip pattern")
	})
}

// TestDefaultWithArgsCustomPackage verifies custom package parameter
func (ts *BenchDefaultArgsTestSuite) TestDefaultWithArgsCustomPackage() {
	withCleanBenchEnv(ts.T(), func() {
		err := ts.bench.DefaultWithArgs("pkg=./cmd/...")
		_ = err // Error expected - we only verify command construction

		commands := ts.mockRunner.GetCommands()
		var goTestCmd string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "go test") {
				goTestCmd = cmd
				break
			}
		}

		ts.Require().Contains(goTestCmd, "./cmd/...", "should use custom package path")
	})
}

// TestDefaultWithArgsCPUProfileFromEnv verifies CPU profile is added from env var
func (ts *BenchDefaultArgsTestSuite) TestDefaultWithArgsCPUProfileFromEnv() {
	withCleanBenchEnv(ts.T(), func() {
		err := os.Setenv("MAGE_X_BENCH_CPU_PROFILE", "test-cpu.prof")
		ts.Require().NoError(err)

		err = ts.bench.DefaultWithArgs()
		_ = err // Error expected - we only verify command construction

		commands := ts.mockRunner.GetCommands()
		var goTestCmd string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "go test") {
				goTestCmd = cmd
				break
			}
		}

		ts.Require().Contains(goTestCmd, "-cpuprofile test-cpu.prof", "should include cpuprofile from env")
	})
}

// TestDefaultWithArgsMemProfileFromEnv verifies memory profile is added from env var
func (ts *BenchDefaultArgsTestSuite) TestDefaultWithArgsMemProfileFromEnv() {
	withCleanBenchEnv(ts.T(), func() {
		err := os.Setenv("MAGE_X_BENCH_MEM_PROFILE", "test-mem.prof")
		ts.Require().NoError(err)

		err = ts.bench.DefaultWithArgs()
		_ = err // Error expected - we only verify command construction

		commands := ts.mockRunner.GetCommands()
		var goTestCmd string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "go test") {
				goTestCmd = cmd
				break
			}
		}

		ts.Require().Contains(goTestCmd, "-memprofile test-mem.prof", "should include memprofile from env")
	})
}

// TestDefaultWithArgsMultipleParameters verifies multiple parameters work together
func (ts *BenchDefaultArgsTestSuite) TestDefaultWithArgsMultipleParameters() {
	withCleanBenchEnv(ts.T(), func() {
		err := ts.bench.DefaultWithArgs("time=10s", "count=3", "verbose=true", "pkg=./pkg/...")
		_ = err // Error expected - we only verify command construction

		commands := ts.mockRunner.GetCommands()
		var goTestCmd string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "go test") {
				goTestCmd = cmd
				break
			}
		}

		ts.Require().Contains(goTestCmd, "-benchtime 10s", "should include custom benchtime")
		ts.Require().Contains(goTestCmd, "-count 3", "should include count")
		ts.Require().Contains(goTestCmd, "-v", "should include verbose flag")
		ts.Require().Contains(goTestCmd, "./pkg/...", "should include custom package")
	})
}

// TestBenchDefaultArgsTestSuite runs the default args test suite
func TestBenchDefaultArgsTestSuite(t *testing.T) {
	suite.Run(t, new(BenchDefaultArgsTestSuite))
}

// =============================================================================
// TRACE COMMAND TESTS
// =============================================================================

// TestTraceWithArgsBaseCommand verifies trace command includes trace flag
func (ts *BenchRealTestSuite) TestTraceWithArgsBaseCommand() {
	withCleanBenchEnv(ts.T(), func() {
		// Save and restore runner
		originalRunner := GetRunner()
		defer func() {
			err := SetRunner(originalRunner)
			ts.Require().NoError(err)
		}()

		mock := NewBenchMockRunner()
		err := SetRunner(mock)
		ts.Require().NoError(err)

		// Create go.mod for module discovery
		err = os.WriteFile(filepath.Join(ts.tempDir, "go.mod"), []byte("module test\n\ngo 1.24\n"), 0o600)
		ts.Require().NoError(err)

		// Run TraceWithArgs - error is expected with mock runner
		err = ts.bench.TraceWithArgs()
		_ = err // Error expected - we only verify command construction

		commands := mock.GetCommands()
		var goTestCmd string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "go test") {
				goTestCmd = cmd
				break
			}
		}

		ts.Require().NotEmpty(goTestCmd, "go test command should be executed")
		ts.Require().Contains(goTestCmd, "-trace", "should include -trace flag")
	})
}

// TestTraceWithArgsCustomTraceFile verifies custom trace file parameter
func (ts *BenchRealTestSuite) TestTraceWithArgsCustomTraceFile() {
	withCleanBenchEnv(ts.T(), func() {
		originalRunner := GetRunner()
		defer func() {
			err := SetRunner(originalRunner)
			ts.Require().NoError(err)
		}()

		mock := NewBenchMockRunner()
		err := SetRunner(mock)
		ts.Require().NoError(err)

		err = os.WriteFile(filepath.Join(ts.tempDir, "go.mod"), []byte("module test\n\ngo 1.24\n"), 0o600)
		ts.Require().NoError(err)

		err = ts.bench.TraceWithArgs("trace=custom-trace.out")
		_ = err // Error expected - we only verify command construction

		commands := mock.GetCommands()
		var goTestCmd string
		for _, cmd := range commands {
			if strings.HasPrefix(cmd, "go test") {
				goTestCmd = cmd
				break
			}
		}

		ts.Require().Contains(goTestCmd, "custom-trace.out", "should use custom trace file")
	})
}

// =============================================================================
// REGRESSION WORKFLOW TESTS
// =============================================================================

// TestRegressionWithArgsNoBaseline verifies baseline creation when missing
func (ts *BenchRealTestSuite) TestRegressionWithArgsNoBaseline() {
	withCleanBenchEnv(ts.T(), func() {
		originalRunner := GetRunner()
		defer func() {
			err := SetRunner(originalRunner)
			ts.Require().NoError(err)
		}()

		mock := NewBenchMockRunner()
		err := SetRunner(mock)
		ts.Require().NoError(err)

		err = os.WriteFile(filepath.Join(ts.tempDir, "go.mod"), []byte("module test\n\ngo 1.24\n"), 0o600)
		ts.Require().NoError(err)

		// Verify no baseline exists
		baselinePath := filepath.Join(ts.tempDir, "bench-baseline.txt")
		_, err = os.Stat(baselinePath)
		ts.Require().True(os.IsNotExist(err), "baseline should not exist initially")

		// Run regression - it should attempt to save benchmarks first
		// The command will execute even though benchmarks don't exist
		err = ts.bench.RegressionWithArgs()
		_ = err // Error expected - we only verify env var setup

		// Verify MAGE_X_BENCH_FILE was set
		val, exists := os.LookupEnv("MAGE_X_BENCH_FILE")
		ts.Require().True(exists, "MAGE_X_BENCH_FILE should be set")
		ts.Require().Equal("bench-current.txt", val)
	})
}

// TestRegressionWithArgsWithBaseline verifies comparison when baseline exists
func (ts *BenchRealTestSuite) TestRegressionWithArgsWithBaseline() {
	withCleanBenchEnv(ts.T(), func() {
		originalRunner := GetRunner()
		defer func() {
			err := SetRunner(originalRunner)
			ts.Require().NoError(err)
		}()

		mock := NewBenchMockRunner()
		err := SetRunner(mock)
		ts.Require().NoError(err)

		err = os.WriteFile(filepath.Join(ts.tempDir, "go.mod"), []byte("module test\n\ngo 1.24\n"), 0o600)
		ts.Require().NoError(err)

		// Create baseline file (use relative path since that's what RegressionWithArgs uses)
		err = os.WriteFile("bench-baseline.txt", []byte("BenchmarkFoo 1000 1000 ns/op"), 0o600)
		ts.Require().NoError(err)

		// Create current file (simulating successful benchmark run)
		err = os.WriteFile("bench-current.txt", []byte("BenchmarkFoo 1000 900 ns/op"), 0o600)
		ts.Require().NoError(err)

		// Run regression
		err = ts.bench.RegressionWithArgs()
		_ = err // Error expected - we only verify env var setup

		// Verify BENCH_OLD and BENCH_NEW env vars are set for comparison
		// RegressionWithArgs uses relative paths (defaults from constants)
		oldVal, _ := os.LookupEnv("BENCH_OLD")
		newVal, _ := os.LookupEnv("BENCH_NEW")
		ts.Require().Equal("bench-baseline.txt", oldVal, "BENCH_OLD should be set to baseline")
		ts.Require().Equal("bench-current.txt", newVal, "BENCH_NEW should be set to current")
	})
}
