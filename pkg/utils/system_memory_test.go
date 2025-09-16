package utils

import (
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetSystemMemoryInfo tests the memory information retrieval
func TestGetSystemMemoryInfo(t *testing.T) {
	t.Run("gets memory info", func(t *testing.T) {
		info, err := GetSystemMemoryInfo()

		// Should work on all supported platforms
		require.NoError(t, err)
		require.NotNil(t, info)

		// Basic sanity checks
		assert.Positive(t, info.TotalMemory, "Total memory should be greater than 0")
		assert.Positive(t, info.AvailableMemory, "Available memory should be greater than 0")

		// Total memory should be greater than or equal to available
		assert.GreaterOrEqual(t, info.TotalMemory, info.AvailableMemory,
			"Total memory should be greater than or equal to available memory")

		// Used memory should make sense
		if info.UsedMemory > 0 {
			assert.LessOrEqual(t, info.UsedMemory, info.TotalMemory,
				"Used memory should not exceed total memory")
		}
	})

	t.Run("platform-specific", func(t *testing.T) {
		info, err := GetSystemMemoryInfo()
		require.NoError(t, err)

		switch runtime.GOOS {
		case osLinux:
			// On Linux, we should have parsed /proc/meminfo
			assert.Greater(t, info.TotalMemory, uint64(1024*1024), // At least 1MB
				"Linux system should report reasonable memory")

		case osDarwin:
			// On macOS, we should have parsed sysctl and vm_stat
			assert.Greater(t, info.TotalMemory, uint64(1024*1024*1024), // At least 1GB
				"macOS system should report reasonable memory")

		case osWindows:
			// On Windows, we should have parsed wmic output
			assert.Greater(t, info.TotalMemory, uint64(1024*1024*1024), // At least 1GB
				"Windows system should report reasonable memory")

		default:
			// Fallback should still provide some estimate
			assert.Positive(t, info.TotalMemory,
				"Fallback should provide non-zero memory estimate")
		}
	})
}

// TestGetAvailableMemory tests the convenience function.
// Note: Memory values may vary between calls due to dynamic system state.
func TestGetAvailableMemory(t *testing.T) {
	t.Run("returns available memory", func(t *testing.T) {
		available := GetAvailableMemory()

		// Should return a reasonable value
		assert.Greater(t, available, uint64(1024*1024), // At least 1MB
			"Available memory should be at least 1MB")

		// Should not be unreasonably large (more than 1TB)
		assert.Less(t, available, uint64(1024*1024*1024*1024),
			"Available memory should be less than 1TB")
	})

	t.Run("consistent with GetSystemMemoryInfo", func(t *testing.T) {
		available := GetAvailableMemory()
		info, err := GetSystemMemoryInfo()

		if err == nil {
			// Allow for reasonable variance between calls since memory is dynamic.
			// Use 1% of available memory or 50MB, whichever is larger, as tolerance.
			tolerance := available / 100  // 1%
			if tolerance < 50*1024*1024 { // minimum 50MB
				tolerance = 50 * 1024 * 1024
			}

			assert.InDelta(t, float64(info.AvailableMemory), float64(available), float64(tolerance),
				"GetAvailableMemory should be within reasonable range of GetSystemMemoryInfo (tolerance: %s)",
				FormatMemory(tolerance))
		} else {
			// If GetSystemMemoryInfo fails, GetAvailableMemory should return default
			assert.Equal(t, uint64(2*1024*1024*1024), available,
				"Should return 2GB default when detection fails")
		}
	})
}

// TestEstimatePackageBuildMemory tests memory estimation
func TestEstimatePackageBuildMemory(t *testing.T) {
	tests := []struct {
		name         string
		packageCount int
		minExpected  uint64
		maxExpected  uint64
	}{
		{
			name:         "single package",
			packageCount: 1,
			minExpected:  500 * 1024 * 1024, // 500MB base + 50MB
			maxExpected:  600 * 1024 * 1024, // Should be around 550MB
		},
		{
			name:         "small project",
			packageCount: 10,
			minExpected:  900 * 1024 * 1024,  // 500MB + 10*50MB
			maxExpected:  1100 * 1024 * 1024, // Should be around 1000MB
		},
		{
			name:         "medium project",
			packageCount: 50,
			minExpected:  2900 * 1024 * 1024, // 500MB + 50*50MB
			maxExpected:  3100 * 1024 * 1024, // Should be around 3000MB
		},
		{
			name:         "large project",
			packageCount: 200,
			minExpected:  8000 * 1024 * 1024, // Should cap at 8GB
			maxExpected:  8000 * 1024 * 1024, // Exactly 8GB due to cap
		},
		{
			name:         "very large project",
			packageCount: 500,
			minExpected:  8000 * 1024 * 1024, // Should cap at 8GB
			maxExpected:  8000 * 1024 * 1024, // Exactly 8GB due to cap
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			estimate := EstimatePackageBuildMemory(tt.packageCount)
			assert.GreaterOrEqual(t, estimate, tt.minExpected,
				"Estimate should be at least %s", FormatMemory(tt.minExpected))
			assert.LessOrEqual(t, estimate, tt.maxExpected,
				"Estimate should be at most %s", FormatMemory(tt.maxExpected))
		})
	}
}

// TestFormatMemory tests memory formatting
func TestFormatMemory(t *testing.T) {
	tests := []struct {
		name     string
		bytes    uint64
		expected string
	}{
		{
			name:     "bytes",
			bytes:    500,
			expected: "500 B",
		},
		{
			name:     "kilobytes",
			bytes:    1536,
			expected: "1.50 KB",
		},
		{
			name:     "megabytes",
			bytes:    5 * 1024 * 1024,
			expected: "5.00 MB",
		},
		{
			name:     "gigabytes",
			bytes:    8 * 1024 * 1024 * 1024,
			expected: "8.00 GB",
		},
		{
			name:     "fractional GB",
			bytes:    uint64(1.5 * 1024 * 1024 * 1024),
			expected: "1.50 GB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatMemory(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMemoryDetectionFallback tests the fallback mechanism
func TestMemoryDetectionFallback(t *testing.T) {
	t.Run("fallback provides reasonable defaults", func(t *testing.T) {
		// Call the default function directly
		info, err := getDefaultMemoryInfo()
		require.NoError(t, err)
		require.NotNil(t, info)

		// Should provide non-zero values
		assert.Positive(t, info.TotalMemory, "Fallback should provide total memory")
		assert.Positive(t, info.AvailableMemory, "Fallback should provide available memory")

		// Values should be reasonable
		assert.GreaterOrEqual(t, info.TotalMemory, info.AvailableMemory)
		assert.Equal(t, info.AvailableMemory, info.FreeMemory)
	})
}

// Benchmark memory detection performance
func BenchmarkGetSystemMemoryInfo(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		info, err := GetSystemMemoryInfo()
		if err != nil {
			b.Fatal(err)
		}
		if info.TotalMemory == 0 {
			b.Fatal("Got zero total memory")
		}
	}
}

// BenchmarkGetAvailableMemory benchmarks the convenience function
func BenchmarkGetAvailableMemory(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		available := GetAvailableMemory()
		if available == 0 {
			b.Fatal("Got zero available memory")
		}
	}
}

// TestPlatformSpecificParsing tests platform-specific parsing logic
func TestPlatformSpecificParsing(t *testing.T) {
	t.Run("linux meminfo parsing", func(t *testing.T) {
		if runtime.GOOS != "linux" {
			t.Skip("Linux-specific test")
		}

		info, err := getLinuxMemoryInfo()
		require.NoError(t, err)

		// On Linux, all values should be populated
		assert.Positive(t, info.TotalMemory)
		assert.Positive(t, info.FreeMemory)
		// AvailableMemory might be 0 on older kernels, but FreeMemory should be used as fallback
		assert.Positive(t, info.AvailableMemory)
	})

	t.Run("darwin vm_stat parsing", func(t *testing.T) {
		if runtime.GOOS != "darwin" {
			t.Skip("macOS-specific test")
		}

		info, err := getDarwinMemoryInfo()
		require.NoError(t, err)

		// On macOS, we should get reasonable values
		assert.Greater(t, info.TotalMemory, uint64(1024*1024*1024)) // At least 1GB
		assert.Positive(t, info.AvailableMemory)
		assert.LessOrEqual(t, info.AvailableMemory, info.TotalMemory)
	})

	t.Run("windows wmic parsing", func(t *testing.T) {
		if runtime.GOOS != "windows" {
			t.Skip("Windows-specific test")
		}

		info, err := getWindowsMemoryInfo()
		require.NoError(t, err)

		// On Windows, we should get reasonable values
		assert.Greater(t, info.TotalMemory, uint64(1024*1024*1024)) // At least 1GB
		assert.Positive(t, info.AvailableMemory)
		assert.LessOrEqual(t, info.AvailableMemory, info.TotalMemory)
	})
}

// TestMemoryParsingEdgeCases tests edge cases in memory parsing
func TestMemoryParsingEdgeCases(t *testing.T) {
	t.Run("handles parsing errors gracefully", func(t *testing.T) {
		// Even if parsing fails, we should get fallback values
		available := GetAvailableMemory()
		assert.Positive(t, available, "Should return non-zero even on error")
	})

	t.Run("memory formatting edge cases", func(t *testing.T) {
		// Test zero
		assert.Equal(t, "0 B", FormatMemory(0))

		// Test very large values
		largeValue := uint64(1024 * 1024 * 1024 * 1024) // 1TB
		result := FormatMemory(largeValue)
		assert.True(t, strings.Contains(result, "GB") || strings.Contains(result, "TB"))
	})
}
