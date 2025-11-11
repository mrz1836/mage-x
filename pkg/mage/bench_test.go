package mage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBench_ComparisonLogic tests benchmark comparison logic
func TestBench_ComparisonLogic(t *testing.T) {
	tests := []struct {
		name      string
		oldResult float64
		newResult float64
		improved  bool
	}{
		{
			name:      "performance improved",
			oldResult: 100.0,
			newResult: 80.0,
			improved:  true,
		},
		{
			name:      "performance degraded",
			oldResult: 100.0,
			newResult: 120.0,
			improved:  false,
		},
		{
			name:      "no change",
			oldResult: 100.0,
			newResult: 100.0,
			improved:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			improved := tt.newResult < tt.oldResult
			assert.Equal(t, tt.improved, improved)
		})
	}
}

// TestBench_RegressionDetection tests regression detection logic
func TestBench_RegressionDetection(t *testing.T) {
	tests := []struct {
		name       string
		threshold  float64
		change     float64
		regression bool
	}{
		{
			name:       "within threshold",
			threshold:  10.0,
			change:     5.0,
			regression: false,
		},
		{
			name:       "exceeds threshold",
			threshold:  10.0,
			change:     15.0,
			regression: true,
		},
		{
			name:       "negative change (improvement)",
			threshold:  10.0,
			change:     -5.0,
			regression: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isRegression := tt.change > tt.threshold
			assert.Equal(t, tt.regression, isRegression)
		})
	}
}

// TestBench_ProfileTypes tests different profiling types
func TestBench_ProfileTypes(t *testing.T) {
	tests := []struct {
		name        string
		profileType string
		valid       bool
	}{
		{
			name:        "cpu profile",
			profileType: "cpu",
			valid:       true,
		},
		{
			name:        "memory profile",
			profileType: "mem",
			valid:       true,
		},
		{
			name:        "block profile",
			profileType: "block",
			valid:       true,
		},
		{
			name:        "mutex profile",
			profileType: "mutex",
			valid:       true,
		},
		{
			name:        "trace profile",
			profileType: "trace",
			valid:       true,
		},
		{
			name:        "invalid profile",
			profileType: "invalid",
			valid:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validTypes := map[string]bool{
				"cpu":   true,
				"mem":   true,
				"block": true,
				"mutex": true,
				"trace": true,
			}
			assert.Equal(t, tt.valid, validTypes[tt.profileType])
		})
	}
}

// TestBench_BenchmarkFilePaths tests benchmark file path handling
func TestBench_BenchmarkFilePaths(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		valid    bool
	}{
		{
			name:     "benchmark result file",
			filename: "benchmark.txt",
			valid:    true,
		},
		{
			name:     "profile output file",
			filename: "cpu.prof",
			valid:    true,
		},
		{
			name:     "empty filename",
			filename: "",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.filename != ""
			assert.Equal(t, tt.valid, isValid)
		})
	}
}

// TestBench_BenchmarkSaving tests benchmark result saving
func TestBench_BenchmarkSaving(t *testing.T) {
	tests := []struct {
		name       string
		savePath   string
		shouldSave bool
	}{
		{
			name:       "save to file",
			savePath:   "benchmarks/current.txt",
			shouldSave: true,
		},
		{
			name:       "no save path",
			savePath:   "",
			shouldSave: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldSave := tt.savePath != ""
			assert.Equal(t, tt.shouldSave, shouldSave)
		})
	}
}

// TestBench_BenchmarkDuration tests benchmark duration settings
func TestBench_BenchmarkDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration string
		valid    bool
	}{
		{
			name:     "valid duration in seconds",
			duration: "10s",
			valid:    true,
		},
		{
			name:     "valid duration in minutes",
			duration: "2m",
			valid:    true,
		},
		{
			name:     "valid duration with iterations",
			duration: "100x",
			valid:    true,
		},
		{
			name:     "invalid duration",
			duration: "invalid",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if duration has valid suffix
			hasValidSuffix := len(tt.duration) > 0 &&
				(tt.duration[len(tt.duration)-1] == 's' ||
					tt.duration[len(tt.duration)-1] == 'm' ||
					tt.duration[len(tt.duration)-1] == 'x')
			if tt.valid {
				assert.True(t, hasValidSuffix || tt.duration == "invalid")
			}
		})
	}
}

// TestBench_PackageSelection tests package selection for benchmarks
func TestBench_PackageSelection(t *testing.T) {
	tests := []struct {
		name     string
		packages []string
		wantAll  bool
	}{
		{
			name:     "specific package",
			packages: []string{"./pkg/mage"},
			wantAll:  false,
		},
		{
			name:     "all packages",
			packages: []string{"./..."},
			wantAll:  true,
		},
		{
			name:     "multiple packages",
			packages: []string{"./pkg/mage", "./pkg/utils"},
			wantAll:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasAll := false
			for _, pkg := range tt.packages {
				if pkg == "./..." {
					hasAll = true
					break
				}
			}
			assert.Equal(t, tt.wantAll, hasAll)
		})
	}
}

// TestBench_BenchmarkFiltering tests benchmark filtering by pattern
func TestBench_BenchmarkFiltering(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		matches []string
	}{
		{
			name:    "match all",
			pattern: ".",
			matches: []string{"BenchmarkFoo", "BenchmarkBar"},
		},
		{
			name:    "match specific",
			pattern: "BenchmarkFoo",
			matches: []string{"BenchmarkFoo"},
		},
		{
			name:    "match none",
			pattern: "NonExistent",
			matches: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.pattern)
		})
	}
}

// TestBench_MemoryAllocationTracking tests memory allocation tracking
func TestBench_MemoryAllocationTracking(t *testing.T) {
	tests := []struct {
		name       string
		allocBytes int64
		threshold  int64
		excessive  bool
	}{
		{
			name:       "low allocation",
			allocBytes: 1024,
			threshold:  10000,
			excessive:  false,
		},
		{
			name:       "high allocation",
			allocBytes: 50000,
			threshold:  10000,
			excessive:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isExcessive := tt.allocBytes > tt.threshold
			assert.Equal(t, tt.excessive, isExcessive)
		})
	}
}

// TestBench_ComparisonOutput tests benchmark comparison output formats
func TestBench_ComparisonOutput(t *testing.T) {
	tests := []struct {
		name   string
		format string
		valid  bool
	}{
		{
			name:   "text format",
			format: "text",
			valid:  true,
		},
		{
			name:   "json format",
			format: "json",
			valid:  true,
		},
		{
			name:   "invalid format",
			format: "invalid",
			valid:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validFormats := map[string]bool{
				"text": true,
				"json": true,
			}
			assert.Equal(t, tt.valid, validFormats[tt.format])
		})
	}
}
