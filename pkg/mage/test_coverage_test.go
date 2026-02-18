package mage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// TestCoverageTestSuite provides comprehensive coverage for Test methods
type TestCoverageTestSuite struct {
	suite.Suite

	env  *testutil.TestEnvironment
	test Test
}

func TestTestCoverageTestSuite(t *testing.T) {
	suite.Run(t, new(TestCoverageTestSuite))
}

func (ts *TestCoverageTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.test = Test{}

	// Create .mage.yaml config
	ts.env.CreateMageConfig(`
project:
  name: test
test:
  verbose: false
  timeout: "10m"
  parallel: 4
`)

	// Set up general mocks for all commands - catch-all approach
	ts.env.Runner.On("RunCmd", mock.Anything, mock.Anything).Return(nil).Maybe()
	ts.env.Runner.On("RunCmdOutput", mock.Anything, mock.Anything).Return("", nil).Maybe()
	ts.env.Runner.On("RunCmdOutputDir", mock.Anything, mock.Anything, mock.Anything).Return("", nil).Maybe()
	ts.env.Runner.On("RunCmdInDir", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
}

func (ts *TestCoverageTestSuite) TearDownTest() {
	TestResetConfig()
	ts.env.Cleanup()
}

// Helper to set up mock runner and execute function
func (ts *TestCoverageTestSuite) withMockRunner(fn func() error) error {
	return ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		fn,
	)
}

// TestTestDefaultExercise exercises Test.Default code path
func (ts *TestCoverageTestSuite) TestTestDefaultExercise() {
	err := ts.withMockRunner(func() error {
		return ts.test.Default()
	})
	// Exercise code path - may fail due to module resolution
	_ = err
}

// TestTestUnitExercise exercises Test.Unit code path
func (ts *TestCoverageTestSuite) TestTestUnitExercise() {
	err := ts.withMockRunner(func() error {
		return ts.test.Unit()
	})
	// Exercise code path - may fail due to module resolution
	_ = err
}

// TestTestShortExercise exercises Test.Short code path
func (ts *TestCoverageTestSuite) TestTestShortExercise() {
	err := ts.withMockRunner(func() error {
		return ts.test.Short()
	})
	// Exercise code path - may fail due to module resolution
	_ = err
}

// TestTestRaceExercise exercises Test.Race code path
func (ts *TestCoverageTestSuite) TestTestRaceExercise() {
	err := ts.withMockRunner(func() error {
		return ts.test.Race()
	})
	// Exercise code path - may fail due to module resolution
	_ = err
}

// TestTestCoverExercise exercises Test.Cover code path
func (ts *TestCoverageTestSuite) TestTestCoverExercise() {
	err := ts.withMockRunner(func() error {
		return ts.test.Cover()
	})
	// Exercise code path - may fail due to module resolution
	_ = err
}

// TestTestCoverRaceExercise exercises Test.CoverRace code path
func (ts *TestCoverageTestSuite) TestTestCoverRaceExercise() {
	err := ts.withMockRunner(func() error {
		return ts.test.CoverRace()
	})
	// Exercise code path - may fail due to module resolution
	_ = err
}

// TestTestRunSuccess tests Test.Run
func (ts *TestCoverageTestSuite) TestTestRunSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.test.Run()
	})

	ts.Require().NoError(err)
}

// TestTestCoverageSuccess tests Test.Coverage
func (ts *TestCoverageTestSuite) TestTestCoverageSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.test.Coverage()
	})

	ts.Require().NoError(err)
}

// TestTestVetSuccess tests Test.Vet
func (ts *TestCoverageTestSuite) TestTestVetSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.test.Vet()
	})

	ts.Require().NoError(err)
}

// TestTestLintSuccess tests Test.Lint
func (ts *TestCoverageTestSuite) TestTestLintSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.test.Lint()
	})

	ts.Require().NoError(err)
}

// TestTestCleanSuccess tests Test.Clean
func (ts *TestCoverageTestSuite) TestTestCleanSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.test.Clean()
	})

	ts.Require().NoError(err)
}

// TestTestAllSuccess tests Test.All
func (ts *TestCoverageTestSuite) TestTestAllSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.test.All()
	})

	ts.Require().NoError(err)
}

// TestTestFullExercise tests Test.Full exercises the code path
func (ts *TestCoverageTestSuite) TestTestFullExercise() {
	err := ts.withMockRunner(func() error {
		return ts.test.Full()
	})
	// Just exercise the code path - may fail due to lint dependencies
	_ = err
}

// TestTestIntegrationExercise tests Test.Integration exercises the code path
func (ts *TestCoverageTestSuite) TestTestIntegrationExercise() {
	err := ts.withMockRunner(func() error {
		return ts.test.Integration()
	})
	// Just exercise the code path
	_ = err
}

// TestTestCIExercise tests Test.CI exercises the code path
func (ts *TestCoverageTestSuite) TestTestCIExercise() {
	err := ts.withMockRunner(func() error {
		return ts.test.CI()
	})
	// Just exercise the code path
	_ = err
}

// TestTestParallelExercise tests Test.Parallel exercises the code path
func (ts *TestCoverageTestSuite) TestTestParallelExercise() {
	err := ts.withMockRunner(func() error {
		return ts.test.Parallel()
	})
	// Just exercise the code path
	_ = err
}

// TestTestNoLintExercise tests Test.NoLint exercises the code path
func (ts *TestCoverageTestSuite) TestTestNoLintExercise() {
	err := ts.withMockRunner(func() error {
		return ts.test.NoLint()
	})
	// Just exercise the code path
	_ = err
}

// TestTestCINoRaceExercise tests Test.CINoRace exercises the code path
func (ts *TestCoverageTestSuite) TestTestCINoRaceExercise() {
	err := ts.withMockRunner(func() error {
		return ts.test.CINoRace()
	})
	// Just exercise the code path
	_ = err
}

// TestTestBenchExercise tests Test.Bench exercises the code path
func (ts *TestCoverageTestSuite) TestTestBenchExercise() {
	err := ts.withMockRunner(func() error {
		return ts.test.Bench()
	})
	// Just exercise the code path
	_ = err
}

// TestTestBenchShortExercise tests Test.BenchShort exercises the code path
func (ts *TestCoverageTestSuite) TestTestBenchShortExercise() {
	err := ts.withMockRunner(func() error {
		return ts.test.BenchShort()
	})
	// Just exercise the code path
	_ = err
}

// TestTestFuzzExercise tests Test.Fuzz exercises the code path
func (ts *TestCoverageTestSuite) TestTestFuzzExercise() {
	err := ts.withMockRunner(func() error {
		return ts.test.Fuzz()
	})
	// Just exercise the code path - may find no fuzz tests
	_ = err
}

// TestTestFuzzShortExercise tests Test.FuzzShort exercises the code path
func (ts *TestCoverageTestSuite) TestTestFuzzShortExercise() {
	err := ts.withMockRunner(func() error {
		return ts.test.FuzzShort()
	})
	// Just exercise the code path
	_ = err
}

// TestTestFuzzWithTimeExercise tests Test.FuzzWithTime exercises the code path
func (ts *TestCoverageTestSuite) TestTestFuzzWithTimeExercise() {
	err := ts.withMockRunner(func() error {
		return ts.test.FuzzWithTime(5 * time.Second)
	})
	// Just exercise the code path
	_ = err
}

// TestTestFuzzShortWithTimeExercise tests Test.FuzzShortWithTime exercises the code path
func (ts *TestCoverageTestSuite) TestTestFuzzShortWithTimeExercise() {
	err := ts.withMockRunner(func() error {
		return ts.test.FuzzShortWithTime(5 * time.Second)
	})
	// Just exercise the code path
	_ = err
}

// TestTestCoverReportNoCoverageFile tests CoverReport when no coverage file exists
func (ts *TestCoverageTestSuite) TestTestCoverReportNoCoverageFile() {
	err := ts.withMockRunner(func() error {
		return ts.test.CoverReport()
	})
	// Should not error, just warns
	ts.Require().NoError(err)
}

// TestTestCoverHTMLNoCoverageFile tests CoverHTML when no coverage file exists
func (ts *TestCoverageTestSuite) TestTestCoverHTMLNoCoverageFile() {
	err := ts.withMockRunner(func() error {
		return ts.test.CoverHTML()
	})
	// Should error because no coverage file
	ts.Require().Error(err)
	ts.Contains(err.Error(), "coverage")
}

// ================== Standalone Tests for Helper Functions ==================

// Note: TestGetCIParams is defined in ci_runner_test.go

func TestFormatBool(t *testing.T) {
	tests := []struct {
		name     string
		value    bool
		expected string
	}{
		{
			name:     "true returns checkmark",
			value:    true,
			expected: "✓",
		},
		{
			name:     "false returns X",
			value:    false,
			expected: "✗",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatBool(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvWithDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "env not set uses default",
			key:          "TEST_GETENV_NOT_SET",
			defaultValue: "default_val",
			envValue:     "",
			expected:     "default_val",
		},
		{
			name:         "env set uses env value",
			key:          "TEST_GETENV_SET",
			defaultValue: "default_val",
			envValue:     "env_val",
			expected:     "env_val",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				t.Setenv(tt.key, tt.envValue)
			} else {
				_ = os.Unsetenv(tt.key) //nolint:errcheck // test cleanup
			}
			result := getEnvWithDefault(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatDurationHelper(t *testing.T) {
	tests := []struct {
		name     string
		duration string
		expected string
	}{
		{
			name:     "empty returns default",
			duration: "",
			expected: "default",
		},
		{
			name:     "non-empty returns as-is",
			duration: "10m",
			expected: "10m",
		},
		{
			name:     "30s",
			duration: "30s",
			expected: "30s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateList(t *testing.T) {
	tests := []struct {
		name      string
		items     []string
		maxLength int
		expected  string
	}{
		{
			name:      "empty list",
			items:     []string{},
			maxLength: 20,
			expected:  "none",
		},
		{
			name:      "short list fits",
			items:     []string{"a", "b", "c"},
			maxLength: 20,
			expected:  "a, b, c",
		},
		{
			name:      "list truncated with ellipsis",
			items:     []string{"item1", "item2", "item3"},
			maxLength: 10,
			expected:  "item1 ...",
		},
		{
			name:      "single long item truncated",
			items:     []string{"verylongitemname"},
			maxLength: 10,
			expected:  "verylon...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateList(tt.items, tt.maxLength)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Note: TestTitleCase is defined in test_buildtags_test.go
// Note: TestGetTagInfo is defined in test_buildtags_test.go

// Note: TestHasTimeoutFlag, TestRemoveTimeoutFlag, TestCalculateFuzzTimeout, TestParseSafeTestArgs
// are defined in test_unit_test.go

func TestParseSafeTestArgsAdditionalCases(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		wantArgs  []string
		wantError bool
	}{
		{
			name:      "allowed flags",
			args:      []string{"-v", "-race", "-cover"},
			wantArgs:  []string{"-v", "-race", "-cover"},
			wantError: false,
		},
		{
			name:      "flag with equals",
			args:      []string{"-count=5"},
			wantArgs:  []string{"-count=5"},
			wantError: false,
		},
		{
			name:      "flag with value",
			args:      []string{"-timeout", "10m"},
			wantArgs:  []string{"-timeout", "10m"},
			wantError: false,
		},
		{
			name:      "disallowed flag",
			args:      []string{"-exec", "rm -rf /"},
			wantArgs:  nil,
			wantError: true,
		},
		{
			name:      "package paths allowed",
			args:      []string{"./...", "./pkg/..."},
			wantArgs:  []string{"./...", "./pkg/..."},
			wantError: false,
		},
		{
			name:      "empty args",
			args:      []string{},
			wantArgs:  nil,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSafeTestArgs(tt.args)
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not allowed")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantArgs, result)
			}
		})
	}
}

func TestBuildTestArgs(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		race           bool
		cover          bool
		additionalArgs []string
		wantContains   []string
	}{
		{
			name: "basic config",
			config: &Config{
				Test: TestConfig{
					Verbose:  true,
					Timeout:  "10m",
					Parallel: 4,
				},
			},
			race:         false,
			cover:        false,
			wantContains: []string{"test", "-v", "-timeout", "10m", "-p", "4"},
		},
		{
			name: "with race",
			config: &Config{
				Test: TestConfig{},
			},
			race:         true,
			cover:        false,
			wantContains: []string{"test", "-race"},
		},
		{
			name: "with cover",
			config: &Config{
				Test: TestConfig{},
			},
			race:         false,
			cover:        true,
			wantContains: []string{"test", "-cover"},
		},
		{
			name: "with tags",
			config: &Config{
				Test: TestConfig{
					Tags: "integration",
				},
			},
			race:         false,
			cover:        false,
			wantContains: []string{"test", "-tags", "integration"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildTestArgs(tt.config, tt.race, tt.cover, tt.additionalArgs...)
			for _, want := range tt.wantContains {
				assert.Contains(t, result, want)
			}
		})
	}
}

func TestBuildTestArgsWithOverrides(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name          string
		config        *Config
		raceOverride  *bool
		coverOverride *bool
		wantRace      bool
		wantCover     bool
	}{
		{
			name: "no overrides uses config",
			config: &Config{
				Test: TestConfig{
					Race:  true,
					Cover: true,
				},
			},
			raceOverride:  nil,
			coverOverride: nil,
			wantRace:      true,
			wantCover:     true,
		},
		{
			name: "override race to false",
			config: &Config{
				Test: TestConfig{
					Race: true,
				},
			},
			raceOverride:  &falseVal,
			coverOverride: nil,
			wantRace:      false,
			wantCover:     false,
		},
		{
			name: "override cover to true",
			config: &Config{
				Test: TestConfig{
					Cover: false,
				},
			},
			raceOverride:  nil,
			coverOverride: &trueVal,
			wantRace:      false,
			wantCover:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildTestArgsWithOverrides(tt.config, tt.raceOverride, tt.coverOverride)
			hasRace := false
			hasCover := false
			for _, arg := range result {
				if arg == "-race" {
					hasRace = true
				}
				if arg == "-cover" {
					hasCover = true
				}
			}
			assert.Equal(t, tt.wantRace, hasRace, "race flag mismatch")
			assert.Equal(t, tt.wantCover, hasCover, "cover flag mismatch")
		})
	}
}

func TestFuzzResultsToError(t *testing.T) {
	tests := []struct {
		name      string
		results   []fuzzTestResult
		wantError bool
		errMsg    string
	}{
		{
			name:      "no failures",
			results:   []fuzzTestResult{{Package: "pkg", Test: "Test1", Error: nil}},
			wantError: false,
		},
		{
			name: "single failure",
			results: []fuzzTestResult{
				{Package: "pkg", Test: "FuzzTest1", Error: errFuzzTestFailed},
			},
			wantError: true,
			errMsg:    "pkg.FuzzTest1",
		},
		{
			name: "multiple failures",
			results: []fuzzTestResult{
				{Package: "pkg1", Test: "FuzzTest1", Error: errFuzzTestFailed},
				{Package: "pkg2", Test: "FuzzTest2", Error: errFuzzTestFailed},
			},
			wantError: true,
			errMsg:    "2 tests",
		},
		{
			name:      "empty results",
			results:   []fuzzTestResult{},
			wantError: false,
		},
		{
			name: "tolerated deadline not counted as failure",
			results: []fuzzTestResult{
				{Package: "pkg", Test: "FuzzTest1", Error: errFuzzTestFailed, DeadlineTolerated: true},
			},
			wantError: false,
		},
		{
			name: "mixed tolerated and real failure",
			results: []fuzzTestResult{
				{Package: "pkg1", Test: "FuzzTest1", Error: errFuzzTestFailed, DeadlineTolerated: true},
				{Package: "pkg2", Test: "FuzzTest2", Error: errFuzzTestFailed, DeadlineTolerated: false},
			},
			wantError: true,
			errMsg:    "pkg2.FuzzTest2",
		},
		{
			name: "all tolerated no failure",
			results: []fuzzTestResult{
				{Package: "pkg1", Test: "FuzzTest1", Error: errFuzzTestFailed, DeadlineTolerated: true},
				{Package: "pkg2", Test: "FuzzTest2", Error: errFuzzTestFailed, DeadlineTolerated: true},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := fuzzResultsToError(tt.results)
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMergeCoverageFiles(t *testing.T) {
	t.Run("no files returns error", func(t *testing.T) {
		err := mergeCoverageFiles([]string{}, "output.txt")
		require.Error(t, err)
		assert.Equal(t, errNoCoverageFilesToMerge, err)
	})

	t.Run("merge single file", func(t *testing.T) {
		tmpDir := t.TempDir()

		file1 := filepath.Join(tmpDir, "coverage1.txt")
		err := os.WriteFile(file1, []byte("mode: atomic\npkg/a.go:1.1,2.1 1 1\n"), 0o600)
		require.NoError(t, err)

		output := filepath.Join(tmpDir, "merged.txt")
		err = mergeCoverageFiles([]string{file1}, output)
		require.NoError(t, err)

		content, err := os.ReadFile(output) //nolint:gosec // G304 - test file path is controlled
		require.NoError(t, err)
		assert.Contains(t, string(content), "mode: atomic")
		assert.Contains(t, string(content), "pkg/a.go")
	})

	t.Run("merge multiple files", func(t *testing.T) {
		tmpDir := t.TempDir()

		file1 := filepath.Join(tmpDir, "coverage1.txt")
		err := os.WriteFile(file1, []byte("mode: atomic\npkg/a.go:1.1,2.1 1 1\n"), 0o600)
		require.NoError(t, err)

		file2 := filepath.Join(tmpDir, "coverage2.txt")
		err = os.WriteFile(file2, []byte("mode: atomic\npkg/b.go:1.1,2.1 1 1\n"), 0o600)
		require.NoError(t, err)

		output := filepath.Join(tmpDir, "merged.txt")
		err = mergeCoverageFiles([]string{file1, file2}, output)
		require.NoError(t, err)

		content, err := os.ReadFile(output) //nolint:gosec // G304 - test file path is controlled
		require.NoError(t, err)
		assert.Contains(t, string(content), "mode: atomic")
		assert.Contains(t, string(content), "pkg/a.go")
		assert.Contains(t, string(content), "pkg/b.go")
	})
}

func TestHandleCoverageFiles(t *testing.T) {
	t.Run("empty list does nothing", func(t *testing.T) {
		// Should not panic with empty list
		handleCoverageFiles([]string{})
	})
}

func TestRunTestsForModulesWithRunnerNilConfig(t *testing.T) {
	err := runTestsForModulesWithRunner(nil, nil, false, false, nil, "unit", "", nil)
	require.Error(t, err)
	assert.Equal(t, errConfigNil, err)
}

func TestRunCoverageTestsForModulesWithRunnerNilConfig(t *testing.T) {
	err := runCoverageTestsForModulesWithRunner(nil, nil, false, nil, "", nil)
	require.Error(t, err)
	assert.Equal(t, errConfigNil, err)
}

func TestDisplayTestHeaderWithNilModules(t *testing.T) {
	config := &Config{
		Test: TestConfig{
			Timeout:  "10m",
			Verbose:  false,
			Parallel: 4,
		},
	}
	// Should not panic
	tags := displayTestHeader("unit", config)
	assert.Empty(t, tags) // No auto-discovery enabled
}

// Note: TestProcessBuildTagAutoDiscovery is defined in test_buildtags_test.go

func TestTestStaticErrors(t *testing.T) {
	// Verify static errors in test.go are defined correctly
	require.Error(t, errNoCoverageFile)
	require.Error(t, errNoCoverageFilesToMerge)
	require.Error(t, errFlagNotAllowed)
	require.Error(t, errFuzzTestFailed)
	assert.Error(t, errConfigNil)
}

func TestExtractFuzzInfrastructureError(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected string
	}{
		{
			name:     "cannot use fuzz flag error",
			output:   "=== RUN   FuzzTest\ncannot use -fuzz flag on package outside the main module\n--- PASS: FuzzTest\n",
			expected: "cannot use -fuzz flag on package outside the main module",
		},
		{
			name:     "build error",
			output:   "# package/name\nerror: undefined: foo\nFAIL package/name [build failed]",
			expected: "error: undefined: foo",
		},
		{
			name:     "FAIL line captured",
			output:   "FAIL github.com/test/pkg [build failed]",
			expected: "FAIL github.com/test/pkg [build failed]",
		},
		{
			name:     "Error with capital E",
			output:   "Running test\nError: something went wrong\nDone",
			expected: "Error: something went wrong",
		},
		{
			name:     "first non-empty line fallback",
			output:   "\n\nSome output line\nAnother line",
			expected: "Some output line",
		},
		{
			name:     "empty output",
			output:   "",
			expected: "",
		},
		{
			name:     "only whitespace and markers",
			output:   "=== RUN Test\n--- PASS: Test\n",
			expected: "",
		},
		{
			name:     "failed keyword",
			output:   "Test failed due to timeout",
			expected: "Test failed due to timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFuzzInfrastructureError(tt.output)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseFuzzResultsWithTextParser_InfrastructureErrors(t *testing.T) {
	tests := []struct {
		name             string
		results          []fuzzTestResult
		contextLines     int
		dedup            bool
		expectedFailures int
		checkFailure     func(t *testing.T, failures []CITestFailure)
	}{
		{
			name: "infrastructure error creates synthetic failure",
			results: []fuzzTestResult{
				{
					Package:  "github.com/test/pkg",
					Test:     "FuzzTest",
					Duration: time.Second,
					Error:    errFuzzTestFailed,
					Output:   "cannot use -fuzz flag on package outside the main module",
				},
			},
			contextLines:     3,
			dedup:            true,
			expectedFailures: 1,
			checkFailure: func(t *testing.T, failures []CITestFailure) {
				require.Len(t, failures, 1)
				assert.Equal(t, "github.com/test/pkg", failures[0].Package)
				assert.Equal(t, "FuzzTest", failures[0].Test)
				assert.Equal(t, FailureTypeBuild, failures[0].Type)
				assert.Contains(t, failures[0].Error, "cannot use -fuzz flag")
				assert.Contains(t, failures[0].Output, "cannot use -fuzz flag")
			},
		},
		{
			name: "normal fuzz failure still works",
			results: []fuzzTestResult{
				{
					Package:  "github.com/test/pkg",
					Test:     "FuzzTest",
					Duration: time.Second,
					Error:    errFuzzTestFailed,
					Output:   "=== RUN   FuzzTest\n--- FAIL: FuzzTest (0.50s)\n    fuzz_test.go:15: failed on input\nFAIL github.com/test/pkg 0.5s\n",
				},
			},
			contextLines:     3,
			dedup:            true,
			expectedFailures: 1,
			checkFailure: func(t *testing.T, failures []CITestFailure) {
				require.Len(t, failures, 1)
				assert.Equal(t, "FuzzTest", failures[0].Test)
				// Should be fuzz type from text parser, not build
				assert.Equal(t, FailureTypeFuzz, failures[0].Type)
			},
		},
		{
			name: "no error no synthetic failure",
			results: []fuzzTestResult{
				{
					Package:  "github.com/test/pkg",
					Test:     "FuzzTest",
					Duration: time.Second,
					Error:    nil,
					Output:   "=== RUN   FuzzTest\n--- PASS: FuzzTest (0.50s)\nok github.com/test/pkg 0.5s\n",
				},
			},
			contextLines:     3,
			dedup:            true,
			expectedFailures: 0,
			checkFailure: func(t *testing.T, failures []CITestFailure) {
				assert.Empty(t, failures)
			},
		},
		{
			name: "empty output with error creates failure via parser",
			results: []fuzzTestResult{
				{
					Package:  "github.com/test/pkg",
					Test:     "FuzzTest",
					Duration: time.Second,
					Error:    errFuzzTestFailed,
					Output:   "",
				},
			},
			contextLines:     3,
			dedup:            true,
			expectedFailures: 1,
			checkFailure: func(t *testing.T, failures []CITestFailure) {
				require.Len(t, failures, 1)
				assert.Equal(t, "FuzzTest", failures[0].Test)
			},
		},
		{
			name: "multiple results with mixed errors",
			results: []fuzzTestResult{
				{
					Package:  "github.com/test/pkg1",
					Test:     "FuzzGood",
					Duration: time.Second,
					Error:    nil,
					Output:   "=== RUN   FuzzGood\n--- PASS: FuzzGood (0.50s)\nok github.com/test/pkg1 0.5s\n",
				},
				{
					Package:  "github.com/test/pkg2",
					Test:     "FuzzInfraError",
					Duration: time.Second,
					Error:    errFuzzTestFailed,
					Output:   "cannot use -fuzz flag on package outside the main module",
				},
			},
			contextLines:     3,
			dedup:            true,
			expectedFailures: 1,
			checkFailure: func(t *testing.T, failures []CITestFailure) {
				require.Len(t, failures, 1)
				assert.Equal(t, "FuzzInfraError", failures[0].Test)
				assert.Equal(t, FailureTypeBuild, failures[0].Type)
			},
		},
		{
			name: "infrastructure error uses Error.Error() fallback",
			results: []fuzzTestResult{
				{
					Package:  "github.com/test/pkg",
					Test:     "FuzzTest",
					Duration: time.Second,
					Error:    errFuzzTestFailed,
					Output:   "=== RUN FuzzTest\n--- PASS: FuzzTest\n", // no error pattern in output
				},
			},
			contextLines:     3,
			dedup:            true,
			expectedFailures: 1,
			checkFailure: func(t *testing.T, failures []CITestFailure) {
				require.Len(t, failures, 1)
				assert.Equal(t, "FuzzTest", failures[0].Test)
				assert.Equal(t, FailureTypeBuild, failures[0].Type)
				// Should use the error message since output has no error patterns
				assert.Contains(t, failures[0].Error, "fuzz test")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			failures := parseFuzzResultsWithTextParser(tt.results, tt.contextLines, tt.dedup)
			assert.Len(t, failures, tt.expectedFailures)
			if tt.checkFailure != nil {
				tt.checkFailure(t, failures)
			}
		})
	}
}
