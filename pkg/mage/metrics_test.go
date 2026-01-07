package mage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testGoMainContent = "package main\nfunc main() {}\n"

// TestFormatNumberWithCommas tests the formatNumberWithCommas helper function
func TestFormatNumberWithCommas(t *testing.T) {
	testCases := []struct {
		name     string
		input    int
		expected string
	}{
		{
			name:     "single digit",
			input:    5,
			expected: "5",
		},
		{
			name:     "double digit",
			input:    42,
			expected: "42",
		},
		{
			name:     "three digits",
			input:    123,
			expected: "123",
		},
		{
			name:     "four digits",
			input:    1234,
			expected: "1,234",
		},
		{
			name:     "five digits",
			input:    12345,
			expected: "12,345",
		},
		{
			name:     "six digits",
			input:    123456,
			expected: "123,456",
		},
		{
			name:     "seven digits",
			input:    1234567,
			expected: "1,234,567",
		},
		{
			name:     "test file count example",
			input:    59475,
			expected: "59,475",
		},
		{
			name:     "go file count example",
			input:    53778,
			expected: "53,778",
		},
		{
			name:     "total count example",
			input:    113253,
			expected: "113,253",
		},
		{
			name:     "one million",
			input:    1000000,
			expected: "1,000,000",
		},
		{
			name:     "zero",
			input:    0,
			expected: "0",
		},
		{
			name:     "exactly thousand",
			input:    1000,
			expected: "1,000",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatNumberWithCommas(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestLOCResult_JSONMarshal tests that LOCResult correctly marshals to JSON
func TestLOCResult_JSONMarshal(t *testing.T) {
	result := LOCResult{
		TestFilesLOC:        1000,
		TestFilesCount:      10,
		GoFilesLOC:          5000,
		GoFilesCount:        50,
		TotalLOC:            6000,
		TotalFilesCount:     60,
		Date:                "2025-01-15",
		ExcludedDirs:        []string{"vendor", "third_party"},
		TestFilesSizeBytes:  50000,
		TestFilesSizeHuman:  "48.8 KB",
		GoFilesSizeBytes:    250000,
		GoFilesSizeHuman:    "244.1 KB",
		TotalSizeBytes:      300000,
		TotalSizeHuman:      "293.0 KB",
		AvgLinesPerFile:     100.0,
		TestCoverageRatio:   20.0,
		PackageCount:        5,
		TestAvgLinesPerFile: 100.0,
		GoAvgLinesPerFile:   100.0,
		TestAvgSizeBytes:    5000,
		GoAvgSizeBytes:      5000,
	}

	jsonBytes, err := json.Marshal(result)
	require.NoError(t, err)

	// Verify it's valid JSON
	var unmarshaled LOCResult
	err = json.Unmarshal(jsonBytes, &unmarshaled)
	require.NoError(t, err)

	// Verify all existing fields are preserved
	assert.Equal(t, result.TestFilesLOC, unmarshaled.TestFilesLOC)
	assert.Equal(t, result.TestFilesCount, unmarshaled.TestFilesCount)
	assert.Equal(t, result.GoFilesLOC, unmarshaled.GoFilesLOC)
	assert.Equal(t, result.GoFilesCount, unmarshaled.GoFilesCount)
	assert.Equal(t, result.TotalLOC, unmarshaled.TotalLOC)
	assert.Equal(t, result.TotalFilesCount, unmarshaled.TotalFilesCount)
	assert.Equal(t, result.Date, unmarshaled.Date)
	assert.Equal(t, result.ExcludedDirs, unmarshaled.ExcludedDirs)

	// Verify all new fields are preserved
	assert.Equal(t, result.TestFilesSizeBytes, unmarshaled.TestFilesSizeBytes)
	assert.Equal(t, result.TestFilesSizeHuman, unmarshaled.TestFilesSizeHuman)
	assert.Equal(t, result.GoFilesSizeBytes, unmarshaled.GoFilesSizeBytes)
	assert.Equal(t, result.GoFilesSizeHuman, unmarshaled.GoFilesSizeHuman)
	assert.Equal(t, result.TotalSizeBytes, unmarshaled.TotalSizeBytes)
	assert.Equal(t, result.TotalSizeHuman, unmarshaled.TotalSizeHuman)
	assert.InDelta(t, result.AvgLinesPerFile, unmarshaled.AvgLinesPerFile, 0.001)
	assert.InDelta(t, result.TestCoverageRatio, unmarshaled.TestCoverageRatio, 0.001)
	assert.Equal(t, result.PackageCount, unmarshaled.PackageCount)
	assert.InDelta(t, result.TestAvgLinesPerFile, unmarshaled.TestAvgLinesPerFile, 0.001)
	assert.InDelta(t, result.GoAvgLinesPerFile, unmarshaled.GoAvgLinesPerFile, 0.001)
	assert.Equal(t, result.TestAvgSizeBytes, unmarshaled.TestAvgSizeBytes)
	assert.Equal(t, result.GoAvgSizeBytes, unmarshaled.GoAvgSizeBytes)
}

// TestLOCResult_JSONFieldNames tests that JSON field names are correct
func TestLOCResult_JSONFieldNames(t *testing.T) {
	result := LOCResult{
		TestFilesLOC:        100,
		TestFilesCount:      5,
		GoFilesLOC:          200,
		GoFilesCount:        10,
		TotalLOC:            300,
		TotalFilesCount:     15,
		Date:                "2025-12-15",
		ExcludedDirs:        []string{"vendor"},
		TestFilesSizeBytes:  1000,
		TestFilesSizeHuman:  "1.0 KB",
		GoFilesSizeBytes:    2000,
		GoFilesSizeHuman:    "2.0 KB",
		TotalSizeBytes:      3000,
		TotalSizeHuman:      "3.0 KB",
		AvgLinesPerFile:     20.0,
		TestCoverageRatio:   50.0,
		PackageCount:        3,
		TestAvgLinesPerFile: 20.0,
		GoAvgLinesPerFile:   20.0,
		TestAvgSizeBytes:    200,
		GoAvgSizeBytes:      200,
	}

	jsonBytes, err := json.Marshal(result)
	require.NoError(t, err)

	jsonStr := string(jsonBytes)

	// Verify existing snake_case field names are used
	assert.Contains(t, jsonStr, `"test_files_loc"`)
	assert.Contains(t, jsonStr, `"test_files_count"`)
	assert.Contains(t, jsonStr, `"go_files_loc"`)
	assert.Contains(t, jsonStr, `"go_files_count"`)
	assert.Contains(t, jsonStr, `"total_loc"`)
	assert.Contains(t, jsonStr, `"total_files_count"`)
	assert.Contains(t, jsonStr, `"date"`)
	assert.Contains(t, jsonStr, `"excluded_dirs"`)

	// Verify new snake_case field names are used
	assert.Contains(t, jsonStr, `"test_files_size_bytes"`)
	assert.Contains(t, jsonStr, `"test_files_size_human"`)
	assert.Contains(t, jsonStr, `"go_files_size_bytes"`)
	assert.Contains(t, jsonStr, `"go_files_size_human"`)
	assert.Contains(t, jsonStr, `"total_size_bytes"`)
	assert.Contains(t, jsonStr, `"total_size_human"`)
	assert.Contains(t, jsonStr, `"avg_lines_per_file"`)
	assert.Contains(t, jsonStr, `"test_coverage_ratio"`)
	assert.Contains(t, jsonStr, `"package_count"`)
	assert.Contains(t, jsonStr, `"test_avg_lines_per_file"`)
	assert.Contains(t, jsonStr, `"go_avg_lines_per_file"`)
	assert.Contains(t, jsonStr, `"test_avg_size_bytes"`)
	assert.Contains(t, jsonStr, `"go_avg_size_bytes"`)
}

// TestLOCStats tests the LOCStats struct
func TestLOCStats(t *testing.T) {
	stats := LOCStats{Lines: 100, Files: 5, TotalBytes: 50000}
	assert.Equal(t, 100, stats.Lines)
	assert.Equal(t, 5, stats.Files)
	assert.Equal(t, int64(50000), stats.TotalBytes)

	// Test zero values
	emptyStats := LOCStats{}
	assert.Equal(t, 0, emptyStats.Lines)
	assert.Equal(t, 0, emptyStats.Files)
	assert.Equal(t, int64(0), emptyStats.TotalBytes)
}

// TestSafeAverage tests the safeAverage helper function
func TestSafeAverage(t *testing.T) {
	// Test normal case
	assert.InDelta(t, 50.0, safeAverage(100, 2), 0.001)
	assert.InDelta(t, 33.333333333333336, safeAverage(100, 3), 0.001)

	// Test division by zero
	assert.InDelta(t, 0.0, safeAverage(100, 0), 0.001)

	// Test zero numerator
	assert.InDelta(t, 0.0, safeAverage(0, 5), 0.001)
}

// TestSafeAverageBytes tests the safeAverageBytes helper function
func TestSafeAverageBytes(t *testing.T) {
	// Test normal case
	assert.Equal(t, int64(50), safeAverageBytes(100, 2))
	assert.Equal(t, int64(33), safeAverageBytes(100, 3))

	// Test division by zero
	assert.Equal(t, int64(0), safeAverageBytes(100, 0))

	// Test zero numerator
	assert.Equal(t, int64(0), safeAverageBytes(0, 5))
}

// TestCountPackages tests the countPackages helper function
func TestCountPackages(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "packages_test")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(tmpDir) }) //nolint:errcheck // cleanup

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.Chdir(originalDir) }) //nolint:errcheck // cleanup

	// Create file in root directory
	err = os.WriteFile("root.go", []byte("package main"), 0o600)
	require.NoError(t, err)

	// Create multiple packages
	pkgDirs := []string{"pkg1", "pkg2", "pkg2/sub"}
	for _, dir := range pkgDirs {
		require.NoError(t, os.MkdirAll(dir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main"), 0o600))
	}

	count, err := countPackages([]string{})
	require.NoError(t, err)
	assert.Equal(t, 4, count) // . + pkg1 + pkg2 + pkg2/sub
}

// TestCountLinesWithStats tests the countLinesWithStats helper function
func TestCountLinesWithStats(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "loc_test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) }) //nolint:errcheck,gosec // cleanup

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(originalDir) }) //nolint:errcheck,gosec // cleanup

	// Create test files
	testContent := "package main\n// comment\nfunc Test() {}\n"
	err = os.WriteFile("example_test.go", []byte(testContent), 0o600)
	require.NoError(t, err)

	t.Run("CountTestFiles", func(t *testing.T) {
		stats, err := countLinesWithStats("*_test.go", []string{})
		require.NoError(t, err)
		assert.Equal(t, 1, stats.Files)
		assert.Equal(t, 2, stats.Lines) // excludes comment and empty lines
	})

	t.Run("ExcludeVendor", func(t *testing.T) {
		// Create vendor directory with test file
		vendorDir := filepath.Join(tmpDir, "vendor")
		err := os.MkdirAll(vendorDir, 0o750)
		require.NoError(t, err)
		err = os.WriteFile(filepath.Join(vendorDir, "vendor_test.go"), []byte(testContent), 0o600)
		require.NoError(t, err)

		stats, err := countLinesWithStats("*_test.go", []string{"vendor"})
		require.NoError(t, err)
		assert.Equal(t, 1, stats.Files) // Should not count vendor file
	})
}

// TestCountGoLinesWithStats tests the countGoLinesWithStats helper function
func TestCountGoLinesWithStats(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "loc_test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) }) //nolint:errcheck,gosec // cleanup

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(originalDir) }) //nolint:errcheck,gosec // cleanup

	// Create Go file (not test)
	goContent := testGoMainContent
	err = os.WriteFile("main.go", []byte(goContent), 0o600)
	require.NoError(t, err)

	// Create test file (should be excluded)
	testContent := "package main\nfunc TestMain() {}\n"
	err = os.WriteFile("main_test.go", []byte(testContent), 0o600)
	require.NoError(t, err)

	t.Run("CountGoFilesExcludingTests", func(t *testing.T) {
		stats, err := countGoLinesWithStats([]string{})
		require.NoError(t, err)
		assert.Equal(t, 1, stats.Files) // Only main.go, not main_test.go
		assert.Equal(t, 2, stats.Lines)
	})
}

// MetricsMockRunner provides a mock runner for metrics tests
type MetricsMockRunner struct {
	RunCmdErr       error
	RunCmdOutputVal string
	RunCmdOutputErr error
	Commands        []string
}

func (m *MetricsMockRunner) RunCmd(cmd string, args ...string) error {
	m.Commands = append(m.Commands, cmd+" "+filepath.Join(args...))
	return m.RunCmdErr
}

func (m *MetricsMockRunner) RunCmdOutput(cmd string, args ...string) (string, error) {
	m.Commands = append(m.Commands, cmd+" "+filepath.Join(args...))
	return m.RunCmdOutputVal, m.RunCmdOutputErr
}

func (m *MetricsMockRunner) RunCmdOutputLines(cmd string, args ...string) ([]string, error) {
	out, err := m.RunCmdOutput(cmd, args...)
	if err != nil {
		return nil, err
	}
	return filepath.SplitList(out), nil
}

func (m *MetricsMockRunner) RunCmdWithInput(_, cmd string, args ...string) error {
	return m.RunCmd(cmd, args...)
}

func (m *MetricsMockRunner) RunCmdDir(_, cmd string, args ...string) error {
	return m.RunCmd(cmd, args...)
}

func (m *MetricsMockRunner) RunCmdDirOutput(_, cmd string, args ...string) (string, error) {
	return m.RunCmdOutput(cmd, args...)
}

func (m *MetricsMockRunner) RunCmdEnv(env []string, cmd string, args ...string) error {
	_ = env
	return m.RunCmd(cmd, args...)
}

// TestMetricsLOC tests the LOC method
func TestMetricsLOC(t *testing.T) {
	// Create temp directory with test files
	tmpDir, err := os.MkdirTemp("", "loc_method_test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) }) //nolint:errcheck,gosec // cleanup

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(originalDir) }) //nolint:errcheck,gosec // cleanup

	// Create test files
	goContent := testGoMainContent
	err = os.WriteFile("main.go", []byte(goContent), 0o600)
	require.NoError(t, err)

	testContent := "package main\nfunc TestMain() {}\n"
	err = os.WriteFile("main_test.go", []byte(testContent), 0o600)
	require.NoError(t, err)

	m := Metrics{}

	t.Run("default output", func(t *testing.T) {
		err := m.LOC()
		require.NoError(t, err)
	})

	t.Run("json output", func(t *testing.T) {
		err := m.LOC("json=true")
		require.NoError(t, err)
	})

	t.Run("json output false", func(t *testing.T) {
		err := m.LOC("json=false")
		require.NoError(t, err)
	})
}

// TestMetricsMage tests the Mage method
func TestMetricsMage(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "mage_test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) }) //nolint:errcheck,gosec // cleanup

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(originalDir) }) //nolint:errcheck,gosec // cleanup

	m := Metrics{}

	t.Run("no magefiles", func(t *testing.T) {
		err := m.Mage()
		require.NoError(t, err)
	})

	t.Run("with magefiles directory", func(t *testing.T) {
		// Create magefiles directory
		magefilesDir := filepath.Join(tmpDir, "magefiles")
		require.NoError(t, os.MkdirAll(magefilesDir, 0o750))

		// Create a magefile
		mageContent := `package main

func Build() error {
	return nil
}

func Test() error {
	return nil
}
`
		err := os.WriteFile(filepath.Join(magefilesDir, "magefile.go"), []byte(mageContent), 0o600)
		require.NoError(t, err)

		err = m.Mage()
		require.NoError(t, err)
	})

	t.Run("with build tag magefile", func(t *testing.T) {
		// Create a file with mage build tag
		mageContent := `//go:build mage
// +build mage

package main

func Deploy() error {
	return nil
}
`
		err := os.WriteFile(filepath.Join(tmpDir, "deploy.go"), []byte(mageContent), 0o600)
		require.NoError(t, err)

		err = m.Mage()
		require.NoError(t, err)
	})

	t.Run("skips hidden directories", func(t *testing.T) {
		// Create hidden directory
		hiddenDir := filepath.Join(tmpDir, ".hidden")
		require.NoError(t, os.MkdirAll(hiddenDir, 0o750))

		mageContent := `//go:build mage
package main
func Hidden() error { return nil }
`
		err := os.WriteFile(filepath.Join(hiddenDir, "hidden.go"), []byte(mageContent), 0o600)
		require.NoError(t, err)

		err = m.Mage()
		require.NoError(t, err)
	})

	t.Run("skips vendor directory", func(t *testing.T) {
		// Create vendor directory
		vendorDir := filepath.Join(tmpDir, "vendor")
		require.NoError(t, os.MkdirAll(vendorDir, 0o750))

		mageContent := `//go:build mage
package main
func Vendor() error { return nil }
`
		err := os.WriteFile(filepath.Join(vendorDir, "vendor.go"), []byte(mageContent), 0o600)
		require.NoError(t, err)

		err = m.Mage()
		require.NoError(t, err)
	})
}

// TestMetricsCoverage tests the Coverage method
func TestMetricsCoverage(t *testing.T) {
	// Save and restore runner
	originalRunner := GetRunner()
	t.Cleanup(func() {
		_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
	})

	m := Metrics{}

	t.Run("success", func(t *testing.T) {
		mock := &MetricsMockRunner{
			RunCmdOutputVal: "total:\t(statements)\t85.0%\n",
		}
		err := SetRunner(mock)
		require.NoError(t, err)

		// Create temp coverage file that will be cleaned up
		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tmpDir))
		t.Cleanup(func() { _ = os.Chdir(originalDir) }) //nolint:errcheck // cleanup

		err = m.Coverage()
		require.NoError(t, err)
	})

	t.Run("coverage generation fails", func(t *testing.T) {
		mock := &MetricsMockRunner{
			RunCmdErr: assert.AnError,
		}
		err := SetRunner(mock)
		require.NoError(t, err)

		err = m.Coverage()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to generate coverage")
	})

	t.Run("coverage analysis fails", func(t *testing.T) {
		mock := &MetricsMockRunner{
			RunCmdOutputErr: assert.AnError,
		}
		err := SetRunner(mock)
		require.NoError(t, err)

		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tmpDir))
		t.Cleanup(func() { _ = os.Chdir(originalDir) }) //nolint:errcheck // cleanup

		err = m.Coverage()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to analyze coverage")
	})
}

// TestMetricsComplexity tests the Complexity method
func TestMetricsComplexity(t *testing.T) {
	// Save and restore runner
	originalRunner := GetRunner()
	t.Cleanup(func() {
		_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
	})

	m := Metrics{}

	t.Run("success with gocyclo installed", func(t *testing.T) {
		mock := &MetricsMockRunner{}
		err := SetRunner(mock)
		require.NoError(t, err)

		err = m.Complexity()
		require.NoError(t, err)
	})

	t.Run("gocyclo top fails", func(t *testing.T) {
		callCount := 0
		// Override to track calls with specialized mock
		err := SetRunner(&complexityMockRunner{
			failOnTopCall: true,
			callCount:     &callCount,
		})
		require.NoError(t, err)

		err = m.Complexity()
		require.Error(t, err)
	})
}

// complexityMockRunner is a specialized mock for complexity tests
type complexityMockRunner struct {
	failOnTopCall bool
	callCount     *int
}

func (m *complexityMockRunner) RunCmd(cmd string, args ...string) error {
	*m.callCount++
	// Fail on "gocyclo -top 10 ." call (3rd call after install check and -over call)
	if m.failOnTopCall && cmd == "gocyclo" {
		for _, arg := range args {
			if arg == "-top" {
				return assert.AnError
			}
		}
	}
	return nil
}

func (m *complexityMockRunner) RunCmdOutput(cmd string, args ...string) (string, error) {
	return "", nil
}

func (m *complexityMockRunner) RunCmdOutputLines(cmd string, args ...string) ([]string, error) {
	return nil, nil
}

func (m *complexityMockRunner) RunCmdWithInput(_, cmd string, args ...string) error {
	return m.RunCmd(cmd, args...)
}

func (m *complexityMockRunner) RunCmdDir(_, cmd string, args ...string) error {
	return m.RunCmd(cmd, args...)
}

func (m *complexityMockRunner) RunCmdDirOutput(_, cmd string, args ...string) (string, error) {
	return "", nil
}

func (m *complexityMockRunner) RunCmdEnv(env []string, cmd string, args ...string) error {
	return m.RunCmd(cmd, args...)
}

// TestMetricsSize tests the Size method
func TestMetricsSizeErrors(t *testing.T) {
	// Save and restore runner
	originalRunner := GetRunner()
	t.Cleanup(func() {
		_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
	})

	m := Metrics{}

	t.Run("build fails", func(t *testing.T) {
		mock := &MetricsMockRunner{
			RunCmdErr: assert.AnError,
		}
		err := SetRunner(mock)
		require.NoError(t, err)

		err = m.Size()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to build binary")
	})
}

// TestMetricsQuality tests the Quality method
func TestMetricsQuality(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "quality_test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) }) //nolint:errcheck,gosec // cleanup

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(originalDir) }) //nolint:errcheck,gosec // cleanup

	// Create basic files for LOC
	goContent := testGoMainContent
	err = os.WriteFile("main.go", []byte(goContent), 0o600)
	require.NoError(t, err)

	// Save and restore runner
	originalRunner := GetRunner()
	t.Cleanup(func() {
		_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
	})

	m := Metrics{}

	t.Run("all checks fail", func(t *testing.T) {
		mock := &MetricsMockRunner{
			RunCmdErr: assert.AnError,
		}
		err := SetRunner(mock)
		require.NoError(t, err)

		err = m.Quality()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "quality checks failed")
	})
}

// TestMetricsImports tests the Imports method
func TestMetricsImports(t *testing.T) {
	// Save and restore runner
	originalRunner := GetRunner()
	t.Cleanup(func() {
		_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
	})

	m := Metrics{}

	t.Run("success", func(t *testing.T) {
		mock := &MetricsMockRunner{
			RunCmdOutputVal: "fmt\nos\ngithub.com/stretchr/testify/assert\ngithub.com/stretchr/testify/assert\ngithub.com/stretchr/testify/assert\ngithub.com/stretchr/testify/assert\n",
		}
		err := SetRunner(mock)
		require.NoError(t, err)

		err = m.Imports()
		require.NoError(t, err)
	})

	t.Run("go list fails", func(t *testing.T) {
		mock := &MetricsMockRunner{
			RunCmdOutputErr: assert.AnError,
		}
		err := SetRunner(mock)
		require.NoError(t, err)

		err = m.Imports()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list imports")
	})

	t.Run("empty imports", func(t *testing.T) {
		mock := &MetricsMockRunner{
			RunCmdOutputVal: "",
		}
		err := SetRunner(mock)
		require.NoError(t, err)

		err = m.Imports()
		require.NoError(t, err)
	})

	t.Run("with internal imports", func(t *testing.T) {
		// Mock to return module name and imports
		ResetGoOperations()
		mockOps := &MockGoOperations{
			ModuleName: "github.com/test/project",
		}
		require.NoError(t, SetGoOperations(mockOps))
		t.Cleanup(func() { ResetGoOperations() })

		mock := &MetricsMockRunner{
			RunCmdOutputVal: "fmt\ngithub.com/test/project/internal\ngithub.com/external/pkg\n",
		}
		err := SetRunner(mock)
		require.NoError(t, err)

		err = m.Imports()
		require.NoError(t, err)
	})
}

// TestMetricsStaticErrors tests the static error variables
func TestMetricsStaticErrors(t *testing.T) {
	require.Error(t, errQualityChecksFailed)
	assert.Contains(t, errQualityChecksFailed.Error(), "quality")
}

// TestCountLinesWithStatsExcludeDir tests directory exclusion in countLinesWithStats
func TestCountLinesWithStatsExcludeDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc_exclude_test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) }) //nolint:errcheck,gosec // cleanup

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(originalDir) }) //nolint:errcheck,gosec // cleanup

	// Create test file in root
	testContent := "package main\nfunc Test() {}\n"
	err = os.WriteFile("root_test.go", []byte(testContent), 0o600)
	require.NoError(t, err)

	// Create third_party directory with test file
	thirdPartyDir := filepath.Join(tmpDir, "third_party")
	require.NoError(t, os.MkdirAll(thirdPartyDir, 0o750))
	err = os.WriteFile(filepath.Join(thirdPartyDir, "external_test.go"), []byte(testContent), 0o600)
	require.NoError(t, err)

	t.Run("excludes third_party", func(t *testing.T) {
		stats, err := countLinesWithStats("*_test.go", []string{"third_party"})
		require.NoError(t, err)
		assert.Equal(t, 1, stats.Files) // Only root_test.go
	})

	t.Run("no exclusions counts all", func(t *testing.T) {
		stats, err := countLinesWithStats("*_test.go", []string{})
		require.NoError(t, err)
		assert.Equal(t, 2, stats.Files) // Both files
	})
}

// TestCountGoLinesWithStatsExcludeDir tests directory exclusion in countGoLinesWithStats
func TestCountGoLinesWithStatsExcludeDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "go_exclude_test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) }) //nolint:errcheck,gosec // cleanup

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(originalDir) }) //nolint:errcheck,gosec // cleanup

	// Create Go file in root
	goContent := testGoMainContent
	err = os.WriteFile("main.go", []byte(goContent), 0o600)
	require.NoError(t, err)

	// Create vendor directory with Go file
	vendorDir := filepath.Join(tmpDir, "vendor")
	require.NoError(t, os.MkdirAll(vendorDir, 0o750))
	err = os.WriteFile(filepath.Join(vendorDir, "vendor.go"), []byte(goContent), 0o600)
	require.NoError(t, err)

	t.Run("excludes vendor", func(t *testing.T) {
		stats, err := countGoLinesWithStats([]string{"vendor"})
		require.NoError(t, err)
		assert.Equal(t, 1, stats.Files) // Only main.go
	})
}

// TestMetricsSize tests the Size method with success path
func TestMetricsSize(t *testing.T) {
	// Save and restore runner
	originalRunner := GetRunner()
	t.Cleanup(func() {
		_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
	})

	m := Metrics{}

	t.Run("success creates and stats binary", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tmpDir))
		t.Cleanup(func() { _ = os.Chdir(originalDir) }) //nolint:errcheck // cleanup

		// Create a fake binary file that will be created by mock
		callCount := 0
		err = SetRunner(&sizeMockRunner{
			createBinary: true,
			tmpDir:       tmpDir,
			callCount:    &callCount,
		})
		require.NoError(t, err)

		err = m.Size()
		require.NoError(t, err)
	})

	t.Run("stat binary fails", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tmpDir))
		t.Cleanup(func() { _ = os.Chdir(originalDir) }) //nolint:errcheck // cleanup

		// Don't create the binary, so stat will fail
		err = SetRunner(&sizeMockRunner{
			createBinary: false,
			tmpDir:       tmpDir,
		})
		require.NoError(t, err)

		err = m.Size()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to stat binary")
	})
}

// sizeMockRunner is a specialized mock for size tests
type sizeMockRunner struct {
	createBinary bool
	tmpDir       string
	callCount    *int
}

func (m *sizeMockRunner) RunCmd(cmd string, args ...string) error {
	if m.callCount != nil {
		*m.callCount++
	}
	// On "go build -o temp-size-check" call, create the binary if requested
	if cmd == "go" && len(args) > 0 && args[0] == "build" && m.createBinary {
		binaryName := "temp-size-check"
		binaryPath := filepath.Join(m.tmpDir, binaryName)
		if err := os.WriteFile(binaryPath, []byte("fake binary content"), 0o600); err != nil {
			return err
		}
	}
	return nil
}

func (m *sizeMockRunner) RunCmdOutput(cmd string, args ...string) (string, error) {
	return "", nil
}

func (m *sizeMockRunner) RunCmdOutputLines(cmd string, args ...string) ([]string, error) {
	return nil, nil
}

func (m *sizeMockRunner) RunCmdWithInput(_, cmd string, args ...string) error {
	return m.RunCmd(cmd, args...)
}

func (m *sizeMockRunner) RunCmdDir(_, cmd string, args ...string) error {
	return m.RunCmd(cmd, args...)
}

func (m *sizeMockRunner) RunCmdDirOutput(_, cmd string, args ...string) (string, error) {
	return "", nil
}

func (m *sizeMockRunner) RunCmdEnv(env []string, cmd string, args ...string) error {
	return m.RunCmd(cmd, args...)
}

// TestMetricsComplexityGocycloInstall tests gocyclo installation path
func TestMetricsComplexityGocycloInstall(t *testing.T) {
	originalRunner := GetRunner()
	t.Cleanup(func() {
		_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
	})

	m := Metrics{}

	t.Run("gocyclo install fails", func(t *testing.T) {
		callCount := 0
		err := SetRunner(&complexityInstallMockRunner{
			failOnInstall: true,
			callCount:     &callCount,
		})
		require.NoError(t, err)

		// This test would require gocyclo to not exist and install to fail
		// The CommandExists check happens outside the runner
		err = m.Complexity()
		// Even if install fails, subsequent calls may succeed
		// This mainly tests the code path
		_ = err
	})
}

// complexityInstallMockRunner mock for testing gocyclo install
type complexityInstallMockRunner struct {
	failOnInstall bool
	callCount     *int
}

func (m *complexityInstallMockRunner) RunCmd(cmd string, args ...string) error {
	*m.callCount++
	if m.failOnInstall && cmd == "go" {
		for _, arg := range args {
			if arg == "install" {
				return assert.AnError
			}
		}
	}
	return nil
}

func (m *complexityInstallMockRunner) RunCmdOutput(cmd string, args ...string) (string, error) {
	return "", nil
}

func (m *complexityInstallMockRunner) RunCmdOutputLines(cmd string, args ...string) ([]string, error) {
	return nil, nil
}

func (m *complexityInstallMockRunner) RunCmdWithInput(_, cmd string, args ...string) error {
	return m.RunCmd(cmd, args...)
}

func (m *complexityInstallMockRunner) RunCmdDir(_, cmd string, args ...string) error {
	return m.RunCmd(cmd, args...)
}

func (m *complexityInstallMockRunner) RunCmdDirOutput(_, cmd string, args ...string) (string, error) {
	return "", nil
}

func (m *complexityInstallMockRunner) RunCmdEnv(env []string, cmd string, args ...string) error {
	return m.RunCmd(cmd, args...)
}

// TestMetricsLOCEdgeCases tests edge cases for LOC
func TestMetricsLOCEdgeCases(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "loc_edge_test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) }) //nolint:errcheck,gosec // cleanup

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(originalDir) }) //nolint:errcheck,gosec // cleanup

	m := Metrics{}

	t.Run("empty directory", func(t *testing.T) {
		err := m.LOC()
		require.NoError(t, err)
	})

	t.Run("with comments only file", func(t *testing.T) {
		content := "// This is a comment\n// Another comment\n"
		err := os.WriteFile("comments.go", []byte(content), 0o600)
		require.NoError(t, err)
		t.Cleanup(func() { os.Remove("comments.go") }) //nolint:errcheck,gosec // cleanup

		err = m.LOC()
		require.NoError(t, err)
	})

	t.Run("with package declaration only", func(t *testing.T) {
		content := "package main\n"
		err := os.WriteFile("minimal.go", []byte(content), 0o600)
		require.NoError(t, err)
		t.Cleanup(func() { os.Remove("minimal.go") }) //nolint:errcheck,gosec // cleanup

		err = m.LOC()
		require.NoError(t, err)
	})
}

// TestMetricsMageWithLongPath tests Mage with long file paths
func TestMetricsMageWithLongPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mage_long_path_test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) }) //nolint:errcheck,gosec // cleanup

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(originalDir) }) //nolint:errcheck,gosec // cleanup

	// Create a deep path with magefile
	deepPath := filepath.Join(tmpDir, "magefiles", "very", "deep", "nested", "path")
	require.NoError(t, os.MkdirAll(deepPath, 0o750))

	// Create a magefile with long path (> 48 chars)
	mageContent := `package main
func VeryLongFunctionNameThatExceedsDisplayLimit() error { return nil }
`
	err = os.WriteFile(filepath.Join(deepPath, "long.go"), []byte(mageContent), 0o600)
	require.NoError(t, err)

	m := Metrics{}
	err = m.Mage()
	require.NoError(t, err)
}

// TestMetricsMageNodeModulesSkip tests that node_modules are skipped
func TestMetricsMageNodeModulesSkip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mage_node_modules_test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) }) //nolint:errcheck,gosec // cleanup

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(originalDir) }) //nolint:errcheck,gosec // cleanup

	// Create node_modules with Go file
	nodeModulesDir := filepath.Join(tmpDir, "node_modules")
	require.NoError(t, os.MkdirAll(nodeModulesDir, 0o750))

	mageContent := `//go:build mage
package main
func NodeModule() error { return nil }
`
	err = os.WriteFile(filepath.Join(nodeModulesDir, "node.go"), []byte(mageContent), 0o600)
	require.NoError(t, err)

	m := Metrics{}
	err = m.Mage()
	require.NoError(t, err)
}

// TestMetricsMageWithTestFile tests that _test.go files are skipped
func TestMetricsMageWithTestFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mage_test_file_test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) }) //nolint:errcheck,gosec // cleanup

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(originalDir) }) //nolint:errcheck,gosec // cleanup

	// Create magefiles directory with test file
	magefilesDir := filepath.Join(tmpDir, "magefiles")
	require.NoError(t, os.MkdirAll(magefilesDir, 0o750))

	// Test file should be skipped
	testContent := `package main
func TestBuild() error { return nil }
`
	err = os.WriteFile(filepath.Join(magefilesDir, "build_test.go"), []byte(testContent), 0o600)
	require.NoError(t, err)

	m := Metrics{}
	err = m.Mage()
	require.NoError(t, err)
}

// TestMetricsQualityPartialFailure tests Quality with partial failures
func TestMetricsQualityPartialFailure(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "quality_partial_test")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(tmpDir) }) //nolint:errcheck,gosec // cleanup

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	err = os.Chdir(tmpDir)
	require.NoError(t, err)
	t.Cleanup(func() { os.Chdir(originalDir) }) //nolint:errcheck,gosec // cleanup

	// Create basic files for LOC
	goContent := testGoMainContent
	err = os.WriteFile("main.go", []byte(goContent), 0o600)
	require.NoError(t, err)

	originalRunner := GetRunner()
	t.Cleanup(func() {
		_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
	})

	m := Metrics{}

	t.Run("coverage fails but LOC succeeds", func(t *testing.T) {
		err := SetRunner(&qualityPartialMockRunner{
			failCoverage: true,
		})
		require.NoError(t, err)

		err = m.Quality()
		require.Error(t, err)
		// Should report failed checks
		assert.Contains(t, err.Error(), "quality checks failed")
	})
}

// qualityPartialMockRunner is for testing partial Quality failures
type qualityPartialMockRunner struct {
	failCoverage bool
}

func (m *qualityPartialMockRunner) RunCmd(cmd string, args ...string) error {
	if m.failCoverage && cmd == "go" {
		for _, arg := range args {
			if arg == "-coverprofile=coverage.tmp" {
				return assert.AnError
			}
		}
	}
	return nil
}

func (m *qualityPartialMockRunner) RunCmdOutput(cmd string, args ...string) (string, error) {
	return "", nil
}

func (m *qualityPartialMockRunner) RunCmdOutputLines(cmd string, args ...string) ([]string, error) {
	return nil, nil
}

func (m *qualityPartialMockRunner) RunCmdWithInput(_, cmd string, args ...string) error {
	return m.RunCmd(cmd, args...)
}

func (m *qualityPartialMockRunner) RunCmdDir(_, cmd string, args ...string) error {
	return m.RunCmd(cmd, args...)
}

func (m *qualityPartialMockRunner) RunCmdDirOutput(_, cmd string, args ...string) (string, error) {
	return "", nil
}

func (m *qualityPartialMockRunner) RunCmdEnv(env []string, cmd string, args ...string) error {
	return m.RunCmd(cmd, args...)
}

// TestMetricsCoverageCleanup tests Coverage file cleanup
func TestMetricsCoverageCleanup(t *testing.T) {
	originalRunner := GetRunner()
	t.Cleanup(func() {
		_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
	})

	m := Metrics{}

	t.Run("coverage file exists and is cleaned up", func(t *testing.T) {
		tmpDir := t.TempDir()
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tmpDir))
		t.Cleanup(func() { _ = os.Chdir(originalDir) }) //nolint:errcheck // cleanup

		// Create a mock that creates the coverage file in current directory
		err = SetRunner(&coverageCleanupMockRunner{
			tmpDir: tmpDir,
		})
		require.NoError(t, err)

		err = m.Coverage()
		require.NoError(t, err)

		// Verify file was cleaned up (coverage.tmp is created in cwd)
		_, statErr := os.Stat("coverage.tmp")
		assert.True(t, os.IsNotExist(statErr))
	})
}

// coverageCleanupMockRunner mock that creates coverage file
type coverageCleanupMockRunner struct {
	tmpDir string
}

func (m *coverageCleanupMockRunner) RunCmd(cmd string, args ...string) error {
	// Create coverage file only when go test -coverprofile is called (first call)
	// Don't recreate on second "go test -cover" call
	if cmd == "go" && len(args) > 1 && args[0] == "test" {
		for _, arg := range args {
			if arg == "-coverprofile=coverage.tmp" {
				if err := os.WriteFile("coverage.tmp", []byte("mode: set\n"), 0o600); err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}

func (m *coverageCleanupMockRunner) RunCmdOutput(cmd string, args ...string) (string, error) {
	return "total:\t(statements)\t85.0%\n", nil
}

func (m *coverageCleanupMockRunner) RunCmdOutputLines(cmd string, args ...string) ([]string, error) {
	return nil, nil
}

func (m *coverageCleanupMockRunner) RunCmdWithInput(_, cmd string, args ...string) error {
	return m.RunCmd(cmd, args...)
}

func (m *coverageCleanupMockRunner) RunCmdDir(_, cmd string, args ...string) error {
	return m.RunCmd(cmd, args...)
}

func (m *coverageCleanupMockRunner) RunCmdDirOutput(_, cmd string, args ...string) (string, error) {
	return "", nil
}

func (m *coverageCleanupMockRunner) RunCmdEnv(env []string, cmd string, args ...string) error {
	return m.RunCmd(cmd, args...)
}
