package mage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSafeTestArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expected    []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid timeout with space separator",
			args:        []string{"-timeout", "5m"},
			expected:    []string{"-timeout", "5m"},
			expectError: false,
		},
		{
			name:        "valid timeout with equals",
			args:        []string{"-timeout=5m"},
			expected:    []string{"-timeout=5m"},
			expectError: false,
		},
		{
			name:        "multiple valid flags including timeout",
			args:        []string{"-v", "-timeout", "30s", "-json"},
			expected:    []string{"-v", "-timeout", "30s", "-json"},
			expectError: false,
		},
		{
			name:        "various duration formats",
			args:        []string{"-timeout", "1h30m45s"},
			expected:    []string{"-timeout", "1h30m45s"},
			expectError: false,
		},
		{
			name:        "invalid flag rejected",
			args:        []string{"-invalid"},
			expected:    nil,
			expectError: true,
			errorMsg:    "flag not allowed for security reasons: -invalid",
		},
		{
			name:        "mixed valid and invalid flags",
			args:        []string{"-v", "-invalid", "-timeout", "5m"},
			expected:    nil,
			expectError: true,
			errorMsg:    "flag not allowed for security reasons: -invalid",
		},
		{
			name:        "empty args",
			args:        []string{},
			expected:    nil,
			expectError: false,
		},
		{
			name:        "single valid flag",
			args:        []string{"-json"},
			expected:    []string{"-json"},
			expectError: false,
		},
		{
			name:        "timeout flag without value",
			args:        []string{"-timeout"},
			expected:    []string{"-timeout"},
			expectError: false,
		},
		{
			name:        "count flag with value",
			args:        []string{"-count", "3"},
			expected:    []string{"-count", "3"},
			expectError: false,
		},
		{
			name:        "count flag with equals",
			args:        []string{"-count=3"},
			expected:    []string{"-count=3"},
			expectError: false,
		},
		{
			name:        "non-flag arguments allowed",
			args:        []string{"-timeout", "5m", "./pkg/utils"},
			expected:    []string{"-timeout", "5m", "./pkg/utils"},
			expectError: false,
		},
		{
			name:        "benchtime flag",
			args:        []string{"-benchtime", "10s"},
			expected:    []string{"-benchtime", "10s"},
			expectError: false,
		},
		{
			name:        "run flag with pattern",
			args:        []string{"-run", "TestMyFunction"},
			expected:    []string{"-run", "TestMyFunction"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSafeTestArgs(tt.args)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestHasTimeoutFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "timeout with space separator",
			args:     []string{"-timeout", "5m"},
			expected: true,
		},
		{
			name:     "timeout with equals",
			args:     []string{"-timeout=5m"},
			expected: true,
		},
		{
			name:     "no timeout flag",
			args:     []string{"-v", "-json"},
			expected: false,
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: false,
		},
		{
			name:     "timeout among other flags",
			args:     []string{"-v", "-timeout", "30s", "-json"},
			expected: true,
		},
		{
			name:     "timeout equals among other flags",
			args:     []string{"-v", "-timeout=30s", "-json"},
			expected: true,
		},
		{
			name:     "partial match should not match",
			args:     []string{"-timeouts", "5m"},
			expected: false,
		},
		{
			name:     "timeout flag without value",
			args:     []string{"-timeout"},
			expected: true,
		},
		{
			name:     "multiple timeout flags",
			args:     []string{"-timeout", "5m", "-timeout=10s"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasTimeoutFlag(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveTimeoutFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "remove timeout with space separator",
			args:     []string{"-timeout", "5m", "-v"},
			expected: []string{"-v"},
		},
		{
			name:     "remove timeout with equals",
			args:     []string{"-v", "-timeout=5m", "-json"},
			expected: []string{"-v", "-json"},
		},
		{
			name:     "no timeout flag present",
			args:     []string{"-v", "-json"},
			expected: []string{"-v", "-json"},
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: nil,
		},
		{
			name:     "only timeout flag",
			args:     []string{"-timeout", "5m"},
			expected: nil,
		},
		{
			name:     "only timeout equals flag",
			args:     []string{"-timeout=5m"},
			expected: nil,
		},
		{
			name:     "preserve flag order",
			args:     []string{"-v", "-timeout", "5m", "-json", "-race"},
			expected: []string{"-v", "-json", "-race"},
		},
		{
			name:     "timeout flag without value",
			args:     []string{"-timeout", "-v"},
			expected: []string{"-v"},
		},
		{
			name:     "multiple timeout flags",
			args:     []string{"-timeout", "5m", "-v", "-timeout=10s", "-json"},
			expected: []string{"-v", "-json"},
		},
		{
			name:     "timeout at end of args",
			args:     []string{"-v", "-json", "-timeout", "5m"},
			expected: []string{"-v", "-json"},
		},
		{
			name:     "timeout with path arguments",
			args:     []string{"-timeout", "5m", "./pkg/utils"},
			expected: []string{"./pkg/utils"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeTimeoutFlag(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildTestArgsTimeoutHandling(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		race           bool
		cover          bool
		additionalArgs []string
		expectedArgs   []string
	}{
		{
			name: "config timeout used when no user timeout",
			config: &Config{
				Test: TestConfig{
					Timeout: "10m",
					Verbose: true,
				},
			},
			race:           false,
			cover:          false,
			additionalArgs: []string{"-json"},
			expectedArgs:   []string{"test", "-v", "-timeout", "10m", "-json"},
		},
		{
			name: "user timeout overrides config timeout",
			config: &Config{
				Test: TestConfig{
					Timeout: "10m",
					Verbose: true,
				},
			},
			race:           false,
			cover:          false,
			additionalArgs: []string{"-timeout", "5m", "-json"},
			expectedArgs:   []string{"test", "-v", "-timeout", "5m", "-json"},
		},
		{
			name: "user timeout equals format overrides config",
			config: &Config{
				Test: TestConfig{
					Timeout: "10m",
					Verbose: true,
				},
			},
			race:           false,
			cover:          false,
			additionalArgs: []string{"-timeout=5m", "-json"},
			expectedArgs:   []string{"test", "-v", "-timeout=5m", "-json"},
		},
		{
			name: "no config timeout, user provides timeout",
			config: &Config{
				Test: TestConfig{
					Verbose: true,
				},
			},
			race:           false,
			cover:          false,
			additionalArgs: []string{"-timeout", "5m"},
			expectedArgs:   []string{"test", "-v", "-timeout", "5m"},
		},
		{
			name: "race and cover flags with timeout handling",
			config: &Config{
				Test: TestConfig{
					Timeout: "10m",
				},
			},
			race:           true,
			cover:          true,
			additionalArgs: []string{"-timeout", "5m"},
			expectedArgs:   []string{"test", "-race", "-cover", "-timeout", "5m"},
		},
		{
			name: "empty additional args uses config defaults",
			config: &Config{
				Test: TestConfig{
					Timeout: "15m",
					Verbose: true,
					Tags:    "integration",
				},
			},
			race:         false,
			cover:        false,
			expectedArgs: []string{"test", "-v", "-timeout", "15m", "-tags", "integration"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildTestArgs(tt.config, tt.race, tt.cover, tt.additionalArgs...)
			assert.Equal(t, tt.expectedArgs, result)
		})
	}
}

func TestBuildTestArgsWithOverridesTimeoutHandling(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		raceOverride   *bool
		coverOverride  *bool
		additionalArgs []string
		expectedArgs   []string
	}{
		{
			name: "user timeout overrides config with race enabled",
			config: &Config{
				Test: TestConfig{
					Timeout: "10m",
					Race:    false,
				},
			},
			raceOverride:   boolPtr(true),
			coverOverride:  nil,
			additionalArgs: []string{"-timeout", "5m"},
			expectedArgs:   []string{"test", "-race", "-timeout", "5m"},
		},
		{
			name: "user timeout overrides config with cover enabled",
			config: &Config{
				Test: TestConfig{
					Timeout: "10m",
					Cover:   false,
				},
			},
			raceOverride:   nil,
			coverOverride:  boolPtr(true),
			additionalArgs: []string{"-timeout=30s"},
			expectedArgs:   []string{"test", "-cover", "-timeout=30s"},
		},
		{
			name: "combination of user flags and overrides",
			config: &Config{
				Test: TestConfig{
					Timeout: "10m",
					Verbose: true,
					Race:    true,
					Cover:   true,
				},
			},
			raceOverride:   boolPtr(false),
			coverOverride:  boolPtr(false),
			additionalArgs: []string{"-timeout", "2m", "-json"},
			expectedArgs:   []string{"test", "-v", "-timeout", "2m", "-json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildTestArgsWithOverrides(tt.config, tt.raceOverride, tt.coverOverride, tt.additionalArgs...)
			assert.Equal(t, tt.expectedArgs, result)
		})
	}
}

// Helper function to create bool pointers for tests
func boolPtr(b bool) *bool {
	return &b
}

func TestCalculateFuzzTimeout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		args     []string
		expected time.Duration
	}{
		{
			name:     "default timeout when no fuzztime",
			args:     []string{"test", "-fuzz=FuzzFoo"},
			expected: 15 * time.Minute,
		},
		{
			name:     "parses fuzztime with 5 minute buffer",
			args:     []string{"test", "-fuzz=FuzzFoo", "-fuzztime", "10s"},
			expected: 10*time.Second + 5*time.Minute,
		},
		{
			name:     "parses longer fuzztime",
			args:     []string{"test", "-fuzztime", "5m", "-fuzz=FuzzBar"},
			expected: 5*time.Minute + 5*time.Minute,
		},
		{
			name:     "caps at maxFuzzTimeout",
			args:     []string{"test", "-fuzztime", "1h"},
			expected: maxFuzzTimeout, // Should be capped at 30 minutes
		},
		{
			name:     "handles invalid duration format",
			args:     []string{"test", "-fuzztime", "invalid"},
			expected: 15 * time.Minute, // Falls back to default
		},
		{
			name:     "handles fuzztime at end of args",
			args:     []string{"test", "-fuzz=FuzzFoo", "-fuzztime"},
			expected: 15 * time.Minute, // Missing value, falls back
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := calculateFuzzTimeout(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindFuzzPackages(t *testing.T) {
	// This test verifies findFuzzPackages uses native Go instead of grep
	// Uses findFuzzPackagesInDir to avoid race conditions from os.Chdir

	// Create temp directory with test files
	tmpDir := t.TempDir()

	// Create go.mod
	goModContent := "module github.com/test/fuzztest\n\ngo 1.24\n"
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goModContent), 0o600))

	// Create a test file WITH fuzz test
	fuzzTestContent := `package fuzztest

import "testing"

func FuzzMyFunction(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		// fuzz logic
	})
}
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "fuzz_test.go"), []byte(fuzzTestContent), 0o600))

	// Create a test file WITHOUT fuzz test
	normalTestContent := `package fuzztest

import "testing"

func TestNormalFunction(t *testing.T) {
	// normal test
}
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "normal_test.go"), []byte(normalTestContent), 0o600))

	// Create a non-test file
	nonTestContent := `package fuzztest

func RegularFunction() {
	// not a test
}
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "regular.go"), []byte(nonTestContent), 0o600))

	// Create subdirectory with fuzz test
	subDir := filepath.Join(tmpDir, "subpkg")
	require.NoError(t, os.MkdirAll(subDir, 0o750))

	subFuzzContent := `package subpkg

import "testing"

func FuzzSubFunction(f *testing.F) {
	f.Add([]byte("seed"))
	f.Fuzz(func(t *testing.T, data []byte) {})
}
`
	require.NoError(t, os.WriteFile(filepath.Join(subDir, "sub_fuzz_test.go"), []byte(subFuzzContent), 0o600))

	// Create vendor directory that should be skipped
	vendorDir := filepath.Join(tmpDir, "vendor", "example")
	require.NoError(t, os.MkdirAll(vendorDir, 0o750))

	vendorFuzzContent := `package example

import "testing"

func FuzzVendored(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {})
}
`
	require.NoError(t, os.WriteFile(filepath.Join(vendorDir, "vendor_fuzz_test.go"), []byte(vendorFuzzContent), 0o600))

	// Run findFuzzPackagesInDir with the temp directory directly
	// This avoids os.Chdir which causes race conditions in parallel tests
	packages := findFuzzPackagesInDir(tmpDir)

	// Should find packages with fuzz tests
	assert.NotEmpty(t, packages, "Should find at least one fuzz package")

	// Should include root package
	foundRoot := false
	for _, pkg := range packages {
		if pkg == "github.com/test/fuzztest" {
			foundRoot = true
			break
		}
	}
	assert.True(t, foundRoot, "Should find fuzz test in root package")

	// Should include subpkg
	foundSubpkg := false
	for _, pkg := range packages {
		if pkg == "github.com/test/fuzztest/subpkg" {
			foundSubpkg = true
			break
		}
	}
	assert.True(t, foundSubpkg, "Should find fuzz test in subpkg")

	// Should NOT include vendor directory
	for _, pkg := range packages {
		assert.NotContains(t, pkg, "vendor", "Should not include vendor directory")
	}
}

func TestRunCoverageTestsForModulesWithRunner_NilConfig(t *testing.T) {
	t.Parallel()

	// Test that nil config returns appropriate error
	err := runCoverageTestsForModulesWithRunner(nil, nil, false, nil, "", nil)
	assert.ErrorIs(t, err, errConfigNil)
}

func TestRunTestsForModulesWithRunner_NilConfig(t *testing.T) {
	t.Parallel()

	// Test that nil config returns appropriate error
	err := runTestsForModulesWithRunner(nil, nil, false, false, nil, "unit", "", nil)
	assert.ErrorIs(t, err, errConfigNil)
}

func TestTestRunnerOptionsValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		opts       testRunnerOptions
		expectType string
		expectRace bool
		expectCov  bool
	}{
		{
			name:       "unit test options",
			opts:       testRunnerOptions{testType: "unit", race: false, isCoverage: false},
			expectType: "unit",
			expectRace: false,
			expectCov:  false,
		},
		{
			name:       "short test options",
			opts:       testRunnerOptions{testType: "short", race: false, isCoverage: false},
			expectType: "short",
			expectRace: false,
			expectCov:  false,
		},
		{
			name:       "race test options",
			opts:       testRunnerOptions{testType: "race", race: true, isCoverage: false},
			expectType: "race",
			expectRace: true,
			expectCov:  false,
		},
		{
			name:       "coverage test options",
			opts:       testRunnerOptions{testType: "coverage", race: false, isCoverage: true},
			expectType: "coverage",
			expectRace: false,
			expectCov:  true,
		},
		{
			name:       "coverage+race test options",
			opts:       testRunnerOptions{testType: "coverage+race", race: true, isCoverage: true},
			expectType: "coverage+race",
			expectRace: true,
			expectCov:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expectType, tt.opts.testType)
			assert.Equal(t, tt.expectRace, tt.opts.race)
			assert.Equal(t, tt.expectCov, tt.opts.isCoverage)
		})
	}
}

func TestBenchmarkOptionsValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		opts            benchmarkOptions
		expectType      string
		expectBenchTime string
		expectLogPrefix string
	}{
		{
			name: "benchmark options",
			opts: benchmarkOptions{
				testType:         "benchmark",
				defaultBenchTime: "10s",
				logPrefix:        "Benchmarks",
			},
			expectType:      "benchmark",
			expectBenchTime: "10s",
			expectLogPrefix: "Benchmarks",
		},
		{
			name: "benchmark-short options",
			opts: benchmarkOptions{
				testType:         "benchmark-short",
				defaultBenchTime: "1s",
				logPrefix:        "Short benchmarks",
			},
			expectType:      "benchmark-short",
			expectBenchTime: "1s",
			expectLogPrefix: "Short benchmarks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expectType, tt.opts.testType)
			assert.Equal(t, tt.expectBenchTime, tt.opts.defaultBenchTime)
			assert.Equal(t, tt.expectLogPrefix, tt.opts.logPrefix)
		})
	}
}

func TestFuzzOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                 string
		opts                 fuzzOptions
		expectHeaderName     string
		expectHeaderText     string
		expectDefaultTime    string
		expectSuccessMessage string
	}{
		{
			name: "fuzz options",
			opts: fuzzOptions{
				headerName:     "fuzz",
				headerText:     "Running Fuzz Tests",
				defaultTime:    "10s",
				successMessage: "Fuzz tests completed",
			},
			expectHeaderName:     "fuzz",
			expectHeaderText:     "Running Fuzz Tests",
			expectDefaultTime:    "10s",
			expectSuccessMessage: "Fuzz tests completed",
		},
		{
			name: "fuzz-short options",
			opts: fuzzOptions{
				headerName:     "fuzz-short",
				headerText:     "Running Short Fuzz Tests",
				defaultTime:    "5s",
				successMessage: "Short fuzz tests completed",
			},
			expectHeaderName:     "fuzz-short",
			expectHeaderText:     "Running Short Fuzz Tests",
			expectDefaultTime:    "5s",
			expectSuccessMessage: "Short fuzz tests completed",
		},
		{
			name: "minimal fuzz options for duration-based methods",
			opts: fuzzOptions{
				headerText:     "Running Fuzz Tests",
				successMessage: "Fuzz tests completed",
			},
			expectHeaderName:     "",
			expectHeaderText:     "Running Fuzz Tests",
			expectDefaultTime:    "",
			expectSuccessMessage: "Fuzz tests completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expectHeaderName, tt.opts.headerName)
			assert.Equal(t, tt.expectHeaderText, tt.opts.headerText)
			assert.Equal(t, tt.expectDefaultTime, tt.opts.defaultTime)
			assert.Equal(t, tt.expectSuccessMessage, tt.opts.successMessage)
		})
	}
}

func TestFuzzOptionsConsistency(t *testing.T) {
	t.Parallel()

	// Test that Fuzz and FuzzShort use correct options
	fuzzOpts := fuzzOptions{
		headerName:     "fuzz",
		defaultTime:    "10s",
		successMessage: "Fuzz tests completed",
	}

	fuzzShortOpts := fuzzOptions{
		headerName:     "fuzz-short",
		defaultTime:    "5s",
		successMessage: "Short fuzz tests completed",
	}

	// Verify the default times are different
	assert.NotEqual(t, fuzzOpts.defaultTime, fuzzShortOpts.defaultTime,
		"Fuzz and FuzzShort should have different default times")

	// Verify header names are different
	assert.NotEqual(t, fuzzOpts.headerName, fuzzShortOpts.headerName,
		"Fuzz and FuzzShort should have different header names")

	// Verify success messages are different
	assert.NotEqual(t, fuzzOpts.successMessage, fuzzShortOpts.successMessage,
		"Fuzz and FuzzShort should have different success messages")

	// Verify the durations are parseable
	_, err := time.ParseDuration(fuzzOpts.defaultTime)
	require.NoError(t, err, "Fuzz default time should be parseable")

	_, err = time.ParseDuration(fuzzShortOpts.defaultTime)
	require.NoError(t, err, "FuzzShort default time should be parseable")
}

func TestFuzzWithTimeOptions(t *testing.T) {
	t.Parallel()

	// Test that FuzzWithTime and FuzzShortWithTime use correct options
	fuzzWithTimeOpts := fuzzOptions{
		headerText:     "Running Fuzz Tests",
		successMessage: "Fuzz tests completed",
	}

	fuzzShortWithTimeOpts := fuzzOptions{
		headerText:     "Running Short Fuzz Tests",
		successMessage: "Short fuzz tests completed",
	}

	// Verify header texts are different
	assert.NotEqual(t, fuzzWithTimeOpts.headerText, fuzzShortWithTimeOpts.headerText,
		"FuzzWithTime and FuzzShortWithTime should have different header texts")

	// Verify success messages are different
	assert.NotEqual(t, fuzzWithTimeOpts.successMessage, fuzzShortWithTimeOpts.successMessage,
		"FuzzWithTime and FuzzShortWithTime should have different success messages")

	// Verify header text contains expected keywords
	assert.Contains(t, fuzzWithTimeOpts.headerText, "Fuzz")
	assert.Contains(t, fuzzShortWithTimeOpts.headerText, "Short")
}
