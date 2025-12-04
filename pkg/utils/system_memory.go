// Package utils provides system memory detection utilities
package utils

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// Error variables for memory operations
var (
	ErrParseMemoryOutput = errors.New("failed to parse total memory from wmic output")
)

// OS constants
const (
	osLinux   = "linux"
	osDarwin  = "darwin"
	osWindows = "windows"
	trueValue = "true"
)

// MemoryInfo contains system memory information
type MemoryInfo struct {
	TotalMemory     uint64 // Total system memory in bytes
	AvailableMemory uint64 // Available memory in bytes
	UsedMemory      uint64 // Used memory in bytes
	FreeMemory      uint64 // Free memory in bytes
}

// GetSystemMemoryInfo returns comprehensive memory information
func GetSystemMemoryInfo() (*MemoryInfo, error) {
	switch runtime.GOOS {
	case osLinux:
		return getLinuxMemoryInfo()
	case osDarwin:
		return getDarwinMemoryInfo()
	case osWindows:
		return getWindowsMemoryInfo()
	default:
		return getDefaultMemoryInfo()
	}
}

// GetAvailableMemory returns the available memory in bytes
func GetAvailableMemory() uint64 {
	info, err := GetSystemMemoryInfo()
	if err != nil {
		// Return a conservative default if we can't detect memory
		return 2 * 1024 * 1024 * 1024 // 2GB default
	}
	return info.AvailableMemory
}

// getLinuxMemoryInfo reads memory info from /proc/meminfo
func getLinuxMemoryInfo() (*MemoryInfo, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return nil, fmt.Errorf("failed to open /proc/meminfo: %w", err)
	}
	defer func() {
		//nolint:errcheck,gosec // Read-only file, close error not critical
		file.Close()
	}()

	info := &MemoryInfo{}
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}

		// Values in /proc/meminfo are in KB
		switch key {
		case "MemTotal":
			info.TotalMemory = value * 1024
		case "MemAvailable":
			info.AvailableMemory = value * 1024
		case "MemFree":
			info.FreeMemory = value * 1024
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read /proc/meminfo: %w", err)
	}

	// Calculate used memory
	info.UsedMemory = info.TotalMemory - info.AvailableMemory

	// If MemAvailable is not present (older kernels), estimate it
	if info.AvailableMemory == 0 && info.FreeMemory > 0 {
		info.AvailableMemory = info.FreeMemory
	}

	return info, nil
}

// getDarwinMemoryInfo uses vm_stat to get memory info on macOS
func getDarwinMemoryInfo() (*MemoryInfo, error) {
	// Get total memory using sysctl
	totalOutput, err := RunCmdOutput("sysctl", "-n", "hw.memsize")
	if err != nil {
		return nil, fmt.Errorf("failed to get total memory: %w", err)
	}

	totalMemory, err := strconv.ParseUint(strings.TrimSpace(totalOutput), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("failed to parse total memory: %w", err)
	}

	// Get memory statistics using vm_stat
	vmStatOutput, err := RunCmdOutput("vm_stat")
	if err != nil {
		// Fallback to a simpler approach when vm_stat fails
		// This provides conservative estimates rather than failing entirely
		//nolint:nilerr // Intentional fallback behavior
		return &MemoryInfo{
			TotalMemory:     totalMemory,
			AvailableMemory: totalMemory / 2, // Conservative estimate
			UsedMemory:      totalMemory / 2,
			FreeMemory:      totalMemory / 2,
		}, nil
	}

	// Parse vm_stat output
	pageSize, freePages, inactivePages := parseVmStatOutput(vmStatOutput)

	// Calculate available memory (free + inactive pages)
	availableMemory := (freePages + inactivePages) * pageSize

	// Ensure available memory doesn't exceed total
	if availableMemory > totalMemory {
		availableMemory = totalMemory * 8 / 10 // Use 80% as available
	}

	return &MemoryInfo{
		TotalMemory:     totalMemory,
		AvailableMemory: availableMemory,
		UsedMemory:      totalMemory - availableMemory,
		FreeMemory:      freePages * pageSize,
	}, nil
}

// getWindowsMemoryInfo uses wmic to get memory info on Windows
func getWindowsMemoryInfo() (*MemoryInfo, error) {
	// Get total physical memory
	totalOutput, err := RunCmdOutput("wmic", "computersystem", "get", "TotalPhysicalMemory", "/value")
	if err != nil {
		return nil, fmt.Errorf("failed to get total memory: %w", err)
	}

	var totalMemory uint64
	for _, line := range strings.Split(totalOutput, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "TotalPhysicalMemory=") {
			value := strings.TrimPrefix(line, "TotalPhysicalMemory=")
			if tm, parseErr := strconv.ParseUint(value, 10, 64); parseErr == nil {
				totalMemory = tm
				break
			}
		}
	}

	if totalMemory == 0 {
		return nil, ErrParseMemoryOutput
	}

	// Get available memory
	availOutput, err := RunCmdOutput("wmic", "OS", "get", "FreePhysicalMemory", "/value")
	if err != nil {
		// Fallback to conservative estimate when wmic fails for available memory
		// This provides usable estimates rather than failing entirely
		//nolint:nilerr // Intentional fallback behavior
		return &MemoryInfo{
			TotalMemory:     totalMemory,
			AvailableMemory: totalMemory / 2,
			UsedMemory:      totalMemory / 2,
			FreeMemory:      totalMemory / 2,
		}, nil
	}

	var freeMemoryKB uint64
	for _, line := range strings.Split(availOutput, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "FreePhysicalMemory=") {
			value := strings.TrimPrefix(line, "FreePhysicalMemory=")
			if fm, err := strconv.ParseUint(value, 10, 64); err == nil {
				freeMemoryKB = fm
				break
			}
		}
	}

	availableMemory := freeMemoryKB * 1024 // Convert KB to bytes

	return &MemoryInfo{
		TotalMemory:     totalMemory,
		AvailableMemory: availableMemory,
		UsedMemory:      totalMemory - availableMemory,
		FreeMemory:      availableMemory,
	}, nil
}

// getDefaultMemoryInfo returns a conservative default when OS-specific detection fails
func getDefaultMemoryInfo() (*MemoryInfo, error) {
	// Use runtime.MemStats as a fallback
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Conservative estimate based on Go's heap allocation
	// Assume the system has at least 4x the current heap size available
	estimatedTotal := m.Sys * 4
	estimatedAvailable := estimatedTotal / 2

	return &MemoryInfo{
		TotalMemory:     estimatedTotal,
		AvailableMemory: estimatedAvailable,
		UsedMemory:      estimatedTotal - estimatedAvailable,
		FreeMemory:      estimatedAvailable,
	}, nil
}

// EstimatePackageBuildMemory estimates memory needed to build N packages
func EstimatePackageBuildMemory(packageCount int) uint64 {
	// Heuristic: Base memory + per-package memory
	// These values are based on empirical observations
	const (
		baseMemoryMB    = 500  // Base Go compiler overhead
		perPackageMemMB = 50   // Average memory per package
		maxMemoryMB     = 8000 // Cap at 8GB
	)

	estimatedMB := baseMemoryMB + (packageCount * perPackageMemMB)
	if estimatedMB > maxMemoryMB {
		estimatedMB = maxMemoryMB
	}
	if estimatedMB < 0 {
		estimatedMB = baseMemoryMB
	}

	return uint64(estimatedMB) * 1024 * 1024
}

// parseVmStatOutput parses vm_stat output to extract page size and memory stats
func parseVmStatOutput(vmStatOutput string) (pageSize, freePages, inactivePages uint64) {
	pageSize = 4096 // Default page size on macOS
	freePages = 0
	inactivePages = 0

	scanner := bufio.NewScanner(strings.NewReader(vmStatOutput))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "page size of") {
			pageSize = extractPageSize(line, pageSize)
		} else if strings.Contains(line, "Pages free:") {
			freePages = extractPagesValue(line)
		} else if strings.Contains(line, "Pages inactive:") {
			inactivePages = extractPagesValue(line)
		}
	}
	return pageSize, freePages, inactivePages
}

// extractPageSize extracts page size from vm_stat line
func extractPageSize(line string, defaultSize uint64) uint64 {
	fields := strings.Fields(line)
	for i, field := range fields {
		if field == "of" && i+1 < len(fields) {
			if ps, err := strconv.ParseUint(fields[i+1], 10, 64); err == nil {
				return ps
			}
			break
		}
	}
	return defaultSize
}

// extractPagesValue extracts page count from vm_stat line
func extractPagesValue(line string) uint64 {
	fields := strings.Fields(line)
	if len(fields) >= 3 {
		value := strings.TrimSuffix(fields[2], ".")
		if pages, err := strconv.ParseUint(value, 10, 64); err == nil {
			return pages
		}
	}
	return 0
}

// FormatMemory formats memory size in human-readable format
func FormatMemory(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
