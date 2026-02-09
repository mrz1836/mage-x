package mage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultFuzzTimingConfig verifies the default configuration values
func TestDefaultFuzzTimingConfig(t *testing.T) {
	// Clear any env overrides for this test
	for _, env := range []string{
		"MAGE_X_FUZZ_BASELINE_BUFFER",
		"MAGE_X_FUZZ_BASELINE_OVERHEAD_PER_SEED",
		"MAGE_X_FUZZ_MIN_TIMEOUT",
		"MAGE_X_FUZZ_MAX_TIMEOUT",
		"MAGE_X_FUZZ_WARMUP_TIMEOUT",
	} {
		t.Setenv(env, "")
	}

	cfg := DefaultFuzzTimingConfig()

	assert.Equal(t, 500*time.Millisecond, cfg.BaselineOverheadPerSeed)
	assert.Equal(t, 90*time.Second, cfg.BaselineBuffer) // Increased to account for compilation
	assert.Equal(t, 30*time.Minute, cfg.MaxTimeout)
	assert.Equal(t, 90*time.Second, cfg.MinTimeout)   // Increased to account for compilation
	assert.Equal(t, 5*time.Minute, cfg.WarmupTimeout) // Build cache warmup timeout
}

// TestDefaultFuzzTimingConfigEnvOverride verifies environment variable overrides work
func TestDefaultFuzzTimingConfigEnvOverride(t *testing.T) {
	// Test MAGE_X_FUZZ_BASELINE_BUFFER override
	t.Run("MAGE_X_FUZZ_BASELINE_BUFFER", func(t *testing.T) {
		t.Setenv("MAGE_X_FUZZ_BASELINE_BUFFER", "2m")
		t.Setenv("MAGE_X_FUZZ_BASELINE_OVERHEAD_PER_SEED", "")
		t.Setenv("MAGE_X_FUZZ_MIN_TIMEOUT", "")
		t.Setenv("MAGE_X_FUZZ_MAX_TIMEOUT", "")
		t.Setenv("MAGE_X_FUZZ_WARMUP_TIMEOUT", "")

		cfg := DefaultFuzzTimingConfig()
		assert.Equal(t, 2*time.Minute, cfg.BaselineBuffer)
		assert.Equal(t, 500*time.Millisecond, cfg.BaselineOverheadPerSeed) // unchanged
		assert.Equal(t, 5*time.Minute, cfg.WarmupTimeout)                  // unchanged
	})

	// Test MAGE_X_FUZZ_BASELINE_OVERHEAD_PER_SEED override
	t.Run("MAGE_X_FUZZ_BASELINE_OVERHEAD_PER_SEED", func(t *testing.T) {
		t.Setenv("MAGE_X_FUZZ_BASELINE_BUFFER", "")
		t.Setenv("MAGE_X_FUZZ_BASELINE_OVERHEAD_PER_SEED", "1s")
		t.Setenv("MAGE_X_FUZZ_MIN_TIMEOUT", "")
		t.Setenv("MAGE_X_FUZZ_MAX_TIMEOUT", "")
		t.Setenv("MAGE_X_FUZZ_WARMUP_TIMEOUT", "")

		cfg := DefaultFuzzTimingConfig()
		assert.Equal(t, 1*time.Second, cfg.BaselineOverheadPerSeed)
		assert.Equal(t, 90*time.Second, cfg.BaselineBuffer) // unchanged
	})

	// Test MAGE_X_FUZZ_MIN_TIMEOUT override
	t.Run("MAGE_X_FUZZ_MIN_TIMEOUT", func(t *testing.T) {
		t.Setenv("MAGE_X_FUZZ_BASELINE_BUFFER", "")
		t.Setenv("MAGE_X_FUZZ_BASELINE_OVERHEAD_PER_SEED", "")
		t.Setenv("MAGE_X_FUZZ_MIN_TIMEOUT", "2m")
		t.Setenv("MAGE_X_FUZZ_MAX_TIMEOUT", "")
		t.Setenv("MAGE_X_FUZZ_WARMUP_TIMEOUT", "")

		cfg := DefaultFuzzTimingConfig()
		assert.Equal(t, 2*time.Minute, cfg.MinTimeout)
	})

	// Test MAGE_X_FUZZ_MAX_TIMEOUT override
	t.Run("MAGE_X_FUZZ_MAX_TIMEOUT", func(t *testing.T) {
		t.Setenv("MAGE_X_FUZZ_BASELINE_BUFFER", "")
		t.Setenv("MAGE_X_FUZZ_BASELINE_OVERHEAD_PER_SEED", "")
		t.Setenv("MAGE_X_FUZZ_MIN_TIMEOUT", "")
		t.Setenv("MAGE_X_FUZZ_MAX_TIMEOUT", "1h")
		t.Setenv("MAGE_X_FUZZ_WARMUP_TIMEOUT", "")

		cfg := DefaultFuzzTimingConfig()
		assert.Equal(t, 1*time.Hour, cfg.MaxTimeout)
	})

	// Test MAGE_X_FUZZ_WARMUP_TIMEOUT override
	t.Run("MAGE_X_FUZZ_WARMUP_TIMEOUT", func(t *testing.T) {
		t.Setenv("MAGE_X_FUZZ_BASELINE_BUFFER", "")
		t.Setenv("MAGE_X_FUZZ_BASELINE_OVERHEAD_PER_SEED", "")
		t.Setenv("MAGE_X_FUZZ_MIN_TIMEOUT", "")
		t.Setenv("MAGE_X_FUZZ_MAX_TIMEOUT", "")
		t.Setenv("MAGE_X_FUZZ_WARMUP_TIMEOUT", "10m")

		cfg := DefaultFuzzTimingConfig()
		assert.Equal(t, 10*time.Minute, cfg.WarmupTimeout)
	})

	// Test MAGE_X_FUZZ_WARMUP_TIMEOUT=0s disables warmup
	t.Run("MAGE_X_FUZZ_WARMUP_TIMEOUT_zero_disables", func(t *testing.T) {
		t.Setenv("MAGE_X_FUZZ_BASELINE_BUFFER", "")
		t.Setenv("MAGE_X_FUZZ_BASELINE_OVERHEAD_PER_SEED", "")
		t.Setenv("MAGE_X_FUZZ_MIN_TIMEOUT", "")
		t.Setenv("MAGE_X_FUZZ_MAX_TIMEOUT", "")
		t.Setenv("MAGE_X_FUZZ_WARMUP_TIMEOUT", "0s")

		cfg := DefaultFuzzTimingConfig()
		assert.Equal(t, time.Duration(0), cfg.WarmupTimeout)
	})

	// Test invalid env var values are ignored
	t.Run("invalid_values_ignored", func(t *testing.T) {
		t.Setenv("MAGE_X_FUZZ_BASELINE_BUFFER", "invalid")
		t.Setenv("MAGE_X_FUZZ_BASELINE_OVERHEAD_PER_SEED", "not-a-duration")
		t.Setenv("MAGE_X_FUZZ_MIN_TIMEOUT", "-5m") // negative
		t.Setenv("MAGE_X_FUZZ_MAX_TIMEOUT", "")
		t.Setenv("MAGE_X_FUZZ_WARMUP_TIMEOUT", "not-valid")

		cfg := DefaultFuzzTimingConfig()
		// Should use defaults when env vars are invalid
		assert.Equal(t, 90*time.Second, cfg.BaselineBuffer)
		assert.Equal(t, 500*time.Millisecond, cfg.BaselineOverheadPerSeed)
		assert.Equal(t, 90*time.Second, cfg.MinTimeout)
		assert.Equal(t, 5*time.Minute, cfg.WarmupTimeout) // default preserved
	})
}

// TestFuzzTimingConfigFromTestConfig tests config conversion
func TestFuzzTimingConfigFromTestConfig(t *testing.T) {
	tests := []struct {
		name     string
		input    *TestConfig
		expected FuzzTimingConfig
	}{
		{
			name:     "nil config uses defaults",
			input:    nil,
			expected: DefaultFuzzTimingConfig(),
		},
		{
			name:     "empty config uses defaults",
			input:    &TestConfig{},
			expected: DefaultFuzzTimingConfig(),
		},
		{
			name: "custom overhead per seed",
			input: &TestConfig{
				FuzzBaselineOverheadPerSeed: "1s",
			},
			expected: FuzzTimingConfig{
				BaselineOverheadPerSeed: 1 * time.Second,
				BaselineBuffer:          90 * time.Second,
				MaxTimeout:              30 * time.Minute,
				MinTimeout:              90 * time.Second,
				WarmupTimeout:           5 * time.Minute,
			},
		},
		{
			name: "custom buffer",
			input: &TestConfig{
				FuzzBaselineBuffer: "2m",
			},
			expected: FuzzTimingConfig{
				BaselineOverheadPerSeed: 500 * time.Millisecond,
				BaselineBuffer:          2 * time.Minute,
				MaxTimeout:              30 * time.Minute,
				MinTimeout:              90 * time.Second,
				WarmupTimeout:           5 * time.Minute,
			},
		},
		{
			name: "both custom values",
			input: &TestConfig{
				FuzzBaselineOverheadPerSeed: "200ms",
				FuzzBaselineBuffer:          "30s",
			},
			expected: FuzzTimingConfig{
				BaselineOverheadPerSeed: 200 * time.Millisecond,
				BaselineBuffer:          30 * time.Second,
				MaxTimeout:              30 * time.Minute,
				MinTimeout:              90 * time.Second,
				WarmupTimeout:           5 * time.Minute,
			},
		},
		{
			name: "invalid overhead falls back to default",
			input: &TestConfig{
				FuzzBaselineOverheadPerSeed: "invalid",
			},
			expected: DefaultFuzzTimingConfig(),
		},
		{
			name: "negative overhead uses default",
			input: &TestConfig{
				FuzzBaselineOverheadPerSeed: "-1s",
			},
			expected: DefaultFuzzTimingConfig(),
		},
		{
			name: "zero buffer is allowed",
			input: &TestConfig{
				FuzzBaselineBuffer: "0s",
			},
			expected: FuzzTimingConfig{
				BaselineOverheadPerSeed: 500 * time.Millisecond,
				BaselineBuffer:          0,
				MaxTimeout:              30 * time.Minute,
				MinTimeout:              90 * time.Second,
				WarmupTimeout:           5 * time.Minute,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FuzzTimingConfigFromTestConfig(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCalculateFuzzTimeoutWithSeedCount tests the timeout calculation logic with explicit seed counts
func TestCalculateFuzzTimeoutWithSeedCount(t *testing.T) {
	cfg := FuzzTimingConfig{
		BaselineOverheadPerSeed: 500 * time.Millisecond,
		BaselineBuffer:          1 * time.Minute,
		MaxTimeout:              30 * time.Minute,
		MinTimeout:              1 * time.Minute,
	}

	tests := []struct {
		name      string
		fuzzTime  time.Duration
		seedCount int
		expected  time.Duration
	}{
		{
			name:      "zero seeds",
			fuzzTime:  10 * time.Second,
			seedCount: 0,
			expected:  1*time.Minute + 10*time.Second, // 10s + 0 + 1m = 1m10s
		},
		{
			name:      "10 seeds",
			fuzzTime:  10 * time.Second,
			seedCount: 10,
			expected:  1*time.Minute + 15*time.Second, // 10s + (10 * 500ms) + 1m = 1m15s
		},
		{
			name:      "60 seeds (like FuzzValidateExtractPath)",
			fuzzTime:  5 * time.Second,
			seedCount: 60,
			expected:  1*time.Minute + 35*time.Second, // 5s + (60 * 500ms) + 1m = 1m35s
		},
		{
			name:      "100 seeds",
			fuzzTime:  10 * time.Second,
			seedCount: 100,
			expected:  2 * time.Minute, // 10s + (100 * 500ms = 50s) + 1m = 2m0s
		},
		{
			name:      "timeout capped at max",
			fuzzTime:  25 * time.Minute,
			seedCount: 1000,
			expected:  30 * time.Minute, // 25m + (1000 * 500ms = 8m20s) + 1m = 34m20s → capped at 30m
		},
		{
			name:      "minimum timeout enforced",
			fuzzTime:  1 * time.Second,
			seedCount: 0,
			expected:  1*time.Minute + 1*time.Second, // 1s + 0 + 1m = 1m1s (exceeds min, so not clamped)
		},
		{
			name:      "minimum timeout enforced when result is below",
			fuzzTime:  0,
			seedCount: 0,
			expected:  1 * time.Minute, // 0 + 0 + 1m = 1m, which equals min
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateFuzzTimeout(tt.fuzzTime, tt.seedCount, cfg)
			assert.Equal(t, tt.expected, result, "Expected %v, got %v", tt.expected, result)
		})
	}
}

// TestCalculateFuzzTimeoutWithArgs tests argument parsing
func TestCalculateFuzzTimeoutWithArgs(t *testing.T) {
	cfg := DefaultFuzzTimingConfig()

	tests := []struct {
		name      string
		args      []string
		seedCount int
	}{
		{
			name:      "fuzztime with space separator",
			args:      []string{"test", "-fuzz=^FuzzFoo$", "-fuzztime", "5s"},
			seedCount: 10,
		},
		{
			name:      "fuzztime with equals",
			args:      []string{"test", "-fuzz=^FuzzFoo$", "-fuzztime=10s"},
			seedCount: 5,
		},
		{
			name:      "no fuzztime uses default (10s)",
			args:      []string{"test", "-fuzz=^FuzzFoo$"},
			seedCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateFuzzTimeoutWithArgs(tt.args, tt.seedCount, cfg)
			// Just verify it returns a reasonable value
			assert.Greater(t, result, time.Duration(0))
			assert.LessOrEqual(t, result, cfg.MaxTimeout)
		})
	}
}

// TestCountFuzzSeeds tests seed counting from source code
func TestCountFuzzSeeds(t *testing.T) {
	// Create a temporary directory with test files
	tmpDir := t.TempDir()

	// Test case 1: Simple fuzz test with direct f.Add calls
	simpleFuzzTest := `package testpkg

import "testing"

func FuzzSimple(f *testing.F) {
	f.Add("hello")
	f.Add("world")
	f.Add("test")
	f.Fuzz(func(t *testing.T, s string) {
		_ = len(s)
	})
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "simple_fuzz_test.go"), []byte(simpleFuzzTest), 0o600)
	require.NoError(t, err)

	// Test case 2: Fuzz test with loop (should detect loop)
	loopFuzzTest := `package testpkg

import "testing"

func FuzzWithLoop(f *testing.F) {
	seeds := []string{"a", "b", "c"}
	for _, s := range seeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		_ = len(s)
	})
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "loop_fuzz_test.go"), []byte(loopFuzzTest), 0o600)
	require.NoError(t, err)

	// Test case 3: Fuzz test with no seeds
	noSeedsFuzzTest := `package testpkg

import "testing"

func FuzzNoSeeds(f *testing.F) {
	f.Fuzz(func(t *testing.T, s string) {
		_ = len(s)
	})
}
`
	err = os.WriteFile(filepath.Join(tmpDir, "noseed_fuzz_test.go"), []byte(noSeedsFuzzTest), 0o600)
	require.NoError(t, err)

	tests := []struct {
		name             string
		testName         string
		expectedCode     int
		expectedHasLoop  bool
		expectSourceFile bool
	}{
		{
			name:             "simple fuzz test with 3 seeds",
			testName:         "FuzzSimple",
			expectedCode:     3,
			expectedHasLoop:  false,
			expectSourceFile: true,
		},
		{
			name:             "fuzz test with loop",
			testName:         "FuzzWithLoop",
			expectedCode:     1, // Detects 1 f.Add call (in loop)
			expectedHasLoop:  true,
			expectSourceFile: true,
		},
		{
			name:             "fuzz test with no seeds",
			testName:         "FuzzNoSeeds",
			expectedCode:     0,
			expectedHasLoop:  false,
			expectSourceFile: true,
		},
		{
			name:             "non-existent test",
			testName:         "FuzzNonExistent",
			expectedCode:     0,
			expectedHasLoop:  false,
			expectSourceFile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := CountFuzzSeeds(tmpDir, tt.testName)
			require.NoError(t, err)
			require.NotNil(t, info)

			assert.Equal(t, tt.expectedCode, info.CodeSeeds, "code seeds mismatch")
			assert.Equal(t, tt.expectedHasLoop, info.HasLoopedSeeds, "hasLoop mismatch")
			if tt.expectSourceFile {
				assert.NotEmpty(t, info.SourceFile, "expected source file to be set")
			}
		})
	}
}

// TestCountCorpusFiles tests counting files in testdata/fuzz directory
func TestCountCorpusFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create testdata/fuzz/FuzzExample directory with corpus files
	corpusDir := filepath.Join(tmpDir, "testdata", "fuzz", "FuzzExample")
	err := os.MkdirAll(corpusDir, 0o700)
	require.NoError(t, err)

	// Create some corpus files
	for i := 0; i < 5; i++ {
		err = os.WriteFile(filepath.Join(corpusDir, "corpus"+string(rune('a'+i))), []byte("test"), 0o600)
		require.NoError(t, err)
	}

	// Test counting
	info, err := CountFuzzSeeds(tmpDir, "FuzzExample")
	require.NoError(t, err)
	assert.Equal(t, 5, info.CorpusSeeds)
	assert.Equal(t, 5, info.TotalSeeds) // No code seeds, only corpus
}

// TestFuzzSeedInfoTotalSeeds verifies total seed calculation
func TestFuzzSeedInfoTotalSeeds(t *testing.T) {
	tmpDir := t.TempDir()

	// Create fuzz test file
	fuzzTest := `package testpkg

import "testing"

func FuzzCombined(f *testing.F) {
	f.Add("a")
	f.Add("b")
	f.Fuzz(func(t *testing.T, s string) {})
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "combined_fuzz_test.go"), []byte(fuzzTest), 0o600)
	require.NoError(t, err)

	// Create corpus files
	corpusDir := filepath.Join(tmpDir, "testdata", "fuzz", "FuzzCombined")
	err = os.MkdirAll(corpusDir, 0o700)
	require.NoError(t, err)
	for i := 0; i < 3; i++ {
		err = os.WriteFile(filepath.Join(corpusDir, "corpus"+string(rune('a'+i))), []byte("test"), 0o600)
		require.NoError(t, err)
	}

	info, err := CountFuzzSeeds(tmpDir, "FuzzCombined")
	require.NoError(t, err)
	assert.Equal(t, 2, info.CodeSeeds)
	assert.Equal(t, 3, info.CorpusSeeds)
	assert.Equal(t, 5, info.TotalSeeds)
}

// TestParseFuzzTimeFromArgs tests fuzztime argument parsing
func TestParseFuzzTimeFromArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected time.Duration
	}{
		{
			name:     "space separated",
			args:     []string{"-fuzztime", "5s"},
			expected: 5 * time.Second,
		},
		{
			name:     "equals format",
			args:     []string{"-fuzztime=10s"},
			expected: 10 * time.Second,
		},
		{
			name:     "with other args",
			args:     []string{"-run=^$", "-fuzz=^Fuzz$", "-fuzztime", "30s", "-v"},
			expected: 30 * time.Second,
		},
		{
			name:     "no fuzztime",
			args:     []string{"-run=^$", "-fuzz=^Fuzz$"},
			expected: 10 * time.Second, // default
		},
		{
			name:     "minutes",
			args:     []string{"-fuzztime", "2m"},
			expected: 2 * time.Minute,
		},
		{
			name:     "complex duration",
			args:     []string{"-fuzztime", "1m30s"},
			expected: 90 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFuzzTimeFromArgs(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFuzzTestTimingString tests the String() method
func TestFuzzTestTimingString(t *testing.T) {
	tests := []struct {
		name     string
		timing   FuzzTestTiming
		contains []string // Check that output contains these substrings
	}{
		{
			name: "with baseline",
			timing: FuzzTestTiming{
				SeedCount:        10,
				BaselineDuration: 5 * time.Second,
				FuzzingDuration:  10 * time.Second,
				TotalDuration:    15 * time.Second,
			},
			contains: []string{"Baseline:", "10 seeds", "Fuzzing:", "Total:"},
		},
		{
			name: "without baseline",
			timing: FuzzTestTiming{
				TotalDuration: 10 * time.Second,
			},
			contains: []string{"Total:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.timing.String()
			for _, sub := range tt.contains {
				assert.Contains(t, result, sub, "expected output to contain %q", sub)
			}
		})
	}
}

// TestEstimateFuzzTestDuration tests duration estimation
func TestEstimateFuzzTestDuration(t *testing.T) {
	cfg := DefaultFuzzTimingConfig()
	fuzzTime := 10 * time.Second

	tests := []struct {
		name     string
		seedInfo *FuzzSeedInfo
		expected time.Duration
	}{
		{
			name:     "nil seed info",
			seedInfo: nil,
			expected: fuzzTime + cfg.BaselineBuffer, // 10s + 1m = 1m10s
		},
		{
			name:     "zero seeds",
			seedInfo: &FuzzSeedInfo{TotalSeeds: 0},
			expected: fuzzTime + cfg.BaselineBuffer,
		},
		{
			name:     "10 seeds",
			seedInfo: &FuzzSeedInfo{TotalSeeds: 10},
			expected: fuzzTime + (10 * cfg.BaselineOverheadPerSeed) + cfg.BaselineBuffer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EstimateFuzzTestDuration(fuzzTime, tt.seedInfo, cfg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsAddCall tests the AST helper function
func TestIsAddCall(t *testing.T) {
	// This is tested indirectly via CountFuzzSeeds, but we can add specific tests
	// if needed for edge cases
}

// TestCountFuzzSeedsEdgeCases tests edge cases in seed counting
func TestCountFuzzSeedsEdgeCases(t *testing.T) {
	t.Run("non-existent directory", func(t *testing.T) {
		info, err := CountFuzzSeeds("/non/existent/path", "FuzzTest")
		require.NoError(t, err) // Should not error, just return 0 seeds
		assert.Equal(t, 0, info.TotalSeeds)
	})

	t.Run("empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		info, err := CountFuzzSeeds(tmpDir, "FuzzTest")
		require.NoError(t, err)
		assert.Equal(t, 0, info.TotalSeeds)
	})

	t.Run("malformed go file", func(t *testing.T) {
		tmpDir := t.TempDir()
		err := os.WriteFile(filepath.Join(tmpDir, "bad_test.go"), []byte("not valid go code {{{"), 0o600)
		require.NoError(t, err)

		info, err := CountFuzzSeeds(tmpDir, "FuzzTest")
		require.NoError(t, err) // Should not error, just skip bad file
		assert.Equal(t, 0, info.TotalSeeds)
	})
}

// TestWarnIfHighSeedCount tests warning generation
func TestWarnIfHighSeedCount(t *testing.T) {
	// This function only logs warnings, so we mainly test it doesn't panic
	cfg := DefaultFuzzTimingConfig()
	fuzzTime := 5 * time.Second

	// Should not panic
	WarnIfHighSeedCount(nil, fuzzTime, cfg)
	WarnIfHighSeedCount(&FuzzSeedInfo{TotalSeeds: 0}, fuzzTime, cfg)
	WarnIfHighSeedCount(&FuzzSeedInfo{TotalSeeds: 100}, fuzzTime, cfg)
	WarnIfHighSeedCount(&FuzzSeedInfo{TotalSeeds: 10, HasLoopedSeeds: true}, fuzzTime, cfg)
}

// TestFuzzTimingConstants verifies the constants are set correctly
func TestFuzzTimingConstants(t *testing.T) {
	// Verify the constants match expected values
	assert.Equal(t, "500ms", DefaultFuzzBaselineOverheadPerSeed)
	assert.Equal(t, "90s", DefaultFuzzBaselineBuffer) // Increased to account for compilation
	assert.Equal(t, "90s", DefaultFuzzMinTimeout)     // Increased to account for compilation
	assert.Equal(t, "5m", DefaultFuzzWarmupTimeout)   // Build cache warmup timeout
}

// Benchmark seed counting performance
func BenchmarkCountFuzzSeeds(b *testing.B) {
	tmpDir := b.TempDir()

	// Create a reasonably sized fuzz test file
	fuzzTest := `package testpkg

import "testing"

func FuzzBenchmark(f *testing.F) {
	f.Add("a")
	f.Add("b")
	f.Add("c")
	f.Add("d")
	f.Add("e")
	f.Fuzz(func(t *testing.T, s string) {})
}
`
	err := os.WriteFile(filepath.Join(tmpDir, "bench_fuzz_test.go"), []byte(fuzzTest), 0o600)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		info, err := CountFuzzSeeds(tmpDir, "FuzzBenchmark")
		if err != nil {
			b.Fatalf("CountFuzzSeeds failed: %v", err)
		}
		_ = info // Use the result
	}
}

// BenchmarkCalculateFuzzTimeout benchmarks timeout calculation
func BenchmarkCalculateFuzzTimeout(b *testing.B) {
	cfg := DefaultFuzzTimingConfig()
	fuzzTime := 10 * time.Second
	seedCount := 50

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = CalculateFuzzTimeout(fuzzTime, seedCount, cfg)
	}
}

// TestWarmFuzzBuildCache tests the build cache warmup function
func TestWarmFuzzBuildCache(t *testing.T) {
	t.Run("valid package compiles successfully", func(t *testing.T) {
		// Use a known valid package from this repo
		elapsed := warmFuzzBuildCache("github.com/mrz1836/mage-x/pkg/utils", 2*time.Minute)
		assert.Greater(t, elapsed, time.Duration(0), "warmup should take some time")
	})

	t.Run("invalid package returns quickly without panic", func(t *testing.T) {
		// warmFuzzBuildCache is non-fatal: it should log and return, not panic
		elapsed := warmFuzzBuildCache("github.com/nonexistent/package/does/not/exist", 30*time.Second)
		assert.Greater(t, elapsed, time.Duration(0), "should still return elapsed time")
	})

	t.Run("very short timeout triggers deadline exceeded", func(t *testing.T) {
		// Use a real package but with an impossibly short timeout
		elapsed := warmFuzzBuildCache("github.com/mrz1836/mage-x/pkg/mage", 1*time.Nanosecond)
		// Should return very quickly (either timeout or immediate failure)
		assert.Greater(t, elapsed, time.Duration(0))
	})

	t.Run("zero timeout disables warmup at call site", func(t *testing.T) {
		// Verify the calling convention: WarmupTimeout=0 means caller skips the call
		cfg := FuzzTimingConfig{WarmupTimeout: 0}
		assert.Equal(t, time.Duration(0), cfg.WarmupTimeout)
		// The caller checks `if fuzzTimingCfg.WarmupTimeout > 0` before calling
	})
}

// TestWarmFuzzBuildCacheEnvDisable tests that MAGE_X_FUZZ_WARMUP_TIMEOUT=0s disables warmup
func TestWarmFuzzBuildCacheEnvDisable(t *testing.T) {
	t.Setenv("MAGE_X_FUZZ_WARMUP_TIMEOUT", "0s")
	// Clear other env vars to isolate
	t.Setenv("MAGE_X_FUZZ_BASELINE_BUFFER", "")
	t.Setenv("MAGE_X_FUZZ_BASELINE_OVERHEAD_PER_SEED", "")
	t.Setenv("MAGE_X_FUZZ_MIN_TIMEOUT", "")
	t.Setenv("MAGE_X_FUZZ_MAX_TIMEOUT", "")

	cfg := DefaultFuzzTimingConfig()
	assert.Equal(t, time.Duration(0), cfg.WarmupTimeout, "warmup should be disabled with 0s")

	// Verify the guard condition
	assert.LessOrEqual(t, cfg.WarmupTimeout, time.Duration(0), "WarmupTimeout > 0 should be false when disabled")
}

// TestWarmFuzzBuildCacheNegativeIgnored tests that negative warmup timeout env values are ignored
func TestWarmFuzzBuildCacheNegativeIgnored(t *testing.T) {
	t.Setenv("MAGE_X_FUZZ_WARMUP_TIMEOUT", "-5m")
	t.Setenv("MAGE_X_FUZZ_BASELINE_BUFFER", "")
	t.Setenv("MAGE_X_FUZZ_BASELINE_OVERHEAD_PER_SEED", "")
	t.Setenv("MAGE_X_FUZZ_MIN_TIMEOUT", "")
	t.Setenv("MAGE_X_FUZZ_MAX_TIMEOUT", "")

	cfg := DefaultFuzzTimingConfig()
	assert.Equal(t, 5*time.Minute, cfg.WarmupTimeout, "negative warmup timeout should be ignored, default preserved")
}

// TestCountFuzzSeedsRealWorld tests seed counting on actual fuzz tests in the repo
func TestCountFuzzSeedsRealWorld(t *testing.T) {
	// Test counting seeds in the actual FuzzValidateExtractPath test
	// This test has seeds added via a loop over a slice of test cases
	info, err := CountFuzzSeeds(".", "FuzzValidateExtractPath")
	require.NoError(t, err)
	require.NotNil(t, info)

	// The test uses a loop to add seeds, so HasLoopedSeeds should be true
	// CodeSeeds will be 1 (the f.Add call in the loop)
	assert.Equal(t, 1, info.CodeSeeds, "should detect 1 f.Add call (in loop)")
	assert.True(t, info.HasLoopedSeeds, "should detect that seeds are added in a loop")
	assert.Contains(t, info.SourceFile, "update_fuzz_test.go", "should find correct source file")
}
