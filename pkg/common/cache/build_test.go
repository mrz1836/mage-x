package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuildCache(t *testing.T) {
	tempDir := t.TempDir()
	cache := NewBuildCache(tempDir)

	assert.NotNil(t, cache)
	assert.Equal(t, tempDir, cache.cacheDir)
	assert.NotNil(t, cache.fileOps)
	assert.True(t, cache.enabled)
	assert.Equal(t, int64(5*1024*1024*1024), cache.maxSize) // 5GB default
	assert.Equal(t, 7*24*time.Hour, cache.ttl)              // 7 days default
}

func TestBuildCache_SetOptions(t *testing.T) {
	cache := NewBuildCache(t.TempDir())

	cache.SetOptions(false, 1024*1024, time.Hour)

	assert.False(t, cache.enabled)
	assert.Equal(t, int64(1024*1024), cache.maxSize)
	assert.Equal(t, time.Hour, cache.ttl)
}

func TestBuildCache_Init(t *testing.T) {
	t.Run("enabled cache", func(t *testing.T) {
		tempDir := t.TempDir()
		cache := NewBuildCache(tempDir)

		err := cache.Init()
		require.NoError(t, err)

		// Check that all required subdirectories were created
		expectedDirs := []string{"builds", "tests", "lint", "deps", "tools", "meta"}
		for _, dir := range expectedDirs {
			dirPath := filepath.Join(tempDir, dir)
			assert.DirExists(t, dirPath, "Directory %s should exist", dir)
		}
	})

	t.Run("disabled cache", func(t *testing.T) {
		tempDir := t.TempDir()
		cache := NewBuildCache(tempDir)
		cache.SetOptions(false, 0, 0)

		err := cache.Init()
		require.NoError(t, err)

		// No directories should be created when disabled
		expectedDirs := []string{"builds", "tests", "lint", "deps", "tools", "meta"}
		for _, dir := range expectedDirs {
			dirPath := filepath.Join(tempDir, dir)
			assert.NoDirExists(t, dirPath, "Directory %s should not exist when disabled", dir)
		}
	})

	t.Run("invalid directory", func(t *testing.T) {
		cache := NewBuildCache("/invalid/readonly/path")

		err := cache.Init()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create cache directory")
	})
}

func TestBuildCache_BuildResult(t *testing.T) {
	tempDir := t.TempDir()
	cache := NewBuildCache(tempDir)
	err := cache.Init()
	require.NoError(t, err)

	// Test storing and retrieving build result
	hash := "test-build-hash"
	buildResult := &BuildResult{
		Binary:      "/path/to/binary",
		Platform:    "linux/amd64",
		BuildFlags:  []string{"-ldflags=-s -w"},
		Environment: map[string]string{"GOOS": "linux"},
		Success:     true,
		Metrics: BuildMetrics{
			CompileTime: time.Second,
			LinkTime:    500 * time.Millisecond,
			BinarySize:  1024 * 1024,
			SourceFiles: 10,
		},
	}

	t.Run("store and get build result", func(t *testing.T) {
		// Create a binary file for the test
		binaryPath := filepath.Join(tempDir, "test-binary")
		err = os.WriteFile(binaryPath, []byte("binary content"), 0o600)
		require.NoError(t, err)

		buildResult.Binary = binaryPath

		// Store result
		err = cache.StoreBuildResult(hash, buildResult)
		require.NoError(t, err)

		// Retrieve result
		retrieved, found := cache.GetBuildResult(hash)
		require.True(t, found)
		require.NotNil(t, retrieved)

		assert.Equal(t, hash, retrieved.Hash)
		assert.Equal(t, buildResult.Binary, retrieved.Binary)
		assert.Equal(t, buildResult.Platform, retrieved.Platform)
		assert.Equal(t, buildResult.BuildFlags, retrieved.BuildFlags)
		assert.Equal(t, buildResult.Environment, retrieved.Environment)
		assert.Equal(t, buildResult.Success, retrieved.Success)
		assert.Equal(t, buildResult.Metrics, retrieved.Metrics)
		assert.False(t, retrieved.Timestamp.IsZero())
	})

	t.Run("get non-existent build result", func(t *testing.T) {
		_, found := cache.GetBuildResult("non-existent-hash")
		assert.False(t, found)
	})

	t.Run("disabled cache", func(t *testing.T) {
		cache.SetOptions(false, 0, 0)

		// Store should succeed but do nothing
		err = cache.StoreBuildResult("disabled-hash", buildResult)
		require.NoError(t, err)

		// Get should return false
		_, found := cache.GetBuildResult("disabled-hash")
		assert.False(t, found)
	})

	t.Run("expired cache entry", func(t *testing.T) {
		cache.SetOptions(true, 1024*1024, time.Nanosecond) // Very short TTL

		err = cache.StoreBuildResult("expired-hash", buildResult)
		require.NoError(t, err)

		time.Sleep(time.Millisecond) // Wait for expiration

		_, found := cache.GetBuildResult("expired-hash")
		assert.False(t, found) // Should be expired and removed
	})
}

func TestBuildCache_TestResult(t *testing.T) {
	tempDir := t.TempDir()
	cache := NewBuildCache(tempDir)
	err := cache.Init()
	require.NoError(t, err)

	hash := "test-test-hash"
	testResult := &TestResult{
		Package:  "./pkg/test",
		Success:  true,
		Output:   "PASS\nok  \t./pkg/test\t1.234s",
		Coverage: 85.5,
		Duration: time.Second,
		Metrics: TestMetrics{
			TestCount:   10,
			PassCount:   8,
			FailCount:   1,
			SkipCount:   1,
			CompileTime: 500 * time.Millisecond,
			ExecuteTime: 700 * time.Millisecond,
		},
	}

	t.Run("store and get test result", func(t *testing.T) {
		err = cache.StoreTestResult(hash, testResult)
		require.NoError(t, err)

		retrieved, found := cache.GetTestResult(hash)
		require.True(t, found)
		require.NotNil(t, retrieved)

		assert.Equal(t, hash, retrieved.Hash)
		assert.Equal(t, testResult.Package, retrieved.Package)
		assert.Equal(t, testResult.Success, retrieved.Success)
		assert.Equal(t, testResult.Output, retrieved.Output)
		assert.InEpsilon(t, testResult.Coverage, retrieved.Coverage, 0.001)
		assert.Equal(t, testResult.Duration, retrieved.Duration)
		assert.Equal(t, testResult.Metrics, retrieved.Metrics)
	})

	t.Run("get non-existent test result", func(t *testing.T) {
		_, found := cache.GetTestResult("non-existent-test-hash")
		assert.False(t, found)
	})

	t.Run("disabled cache", func(t *testing.T) {
		cache.SetOptions(false, 0, 0)

		err = cache.StoreTestResult("disabled-test-hash", testResult)
		require.NoError(t, err)

		_, found := cache.GetTestResult("disabled-test-hash")
		assert.False(t, found)
	})
}

func TestBuildCache_LintResult(t *testing.T) {
	tempDir := t.TempDir()
	cache := NewBuildCache(tempDir)
	err := cache.Init()
	require.NoError(t, err)

	hash := "test-lint-hash"
	lintResult := &LintResult{
		Files:   []string{"main.go", "helper.go"},
		Success: false,
		Issues: []LintIssue{
			{
				File:     "main.go",
				Line:     10,
				Column:   5,
				Rule:     "gofmt",
				Severity: "error",
				Message:  "File is not gofmt-ed",
			},
			{
				File:     "helper.go",
				Line:     25,
				Column:   12,
				Rule:     "unused",
				Severity: "warning",
				Message:  "Variable 'x' is not used",
			},
		},
		Duration: 800 * time.Millisecond,
	}

	t.Run("store and get lint result", func(t *testing.T) {
		err = cache.StoreLintResult(hash, lintResult)
		require.NoError(t, err)

		retrieved, found := cache.GetLintResult(hash)
		require.True(t, found)
		require.NotNil(t, retrieved)

		assert.Equal(t, hash, retrieved.Hash)
		assert.Equal(t, lintResult.Files, retrieved.Files)
		assert.Equal(t, lintResult.Success, retrieved.Success)
		assert.Equal(t, lintResult.Issues, retrieved.Issues)
		assert.Equal(t, lintResult.Duration, retrieved.Duration)
	})

	t.Run("get non-existent lint result", func(t *testing.T) {
		_, found := cache.GetLintResult("non-existent-lint-hash")
		assert.False(t, found)
	})
}

func TestBuildCache_DependencyResult(t *testing.T) {
	tempDir := t.TempDir()
	cache := NewBuildCache(tempDir)
	err := cache.Init()
	require.NoError(t, err)

	hash := "test-deps-hash"
	depsResult := &DependencyResult{
		ModFile: "go.mod",
		SumFile: "go.sum",
		Modules: []ModuleInfo{
			{
				Path:    "github.com/example/lib",
				Version: "v1.2.3",
				Hash:    "abc123",
				Time:    time.Now(),
			},
		},
		Success:     true,
		Environment: map[string]string{"GOPROXY": "https://proxy.golang.org"},
	}

	t.Run("store and get dependency result", func(t *testing.T) {
		err = cache.StoreDependencyResult(hash, depsResult)
		require.NoError(t, err)

		retrieved, found := cache.GetDependencyResult(hash)
		require.True(t, found)
		require.NotNil(t, retrieved)

		assert.Equal(t, hash, retrieved.Hash)
		assert.Equal(t, depsResult.ModFile, retrieved.ModFile)
		assert.Equal(t, depsResult.SumFile, retrieved.SumFile)
		// Compare modules but skip Time field due to precision issues
		require.Len(t, retrieved.Modules, len(depsResult.Modules))
		assert.Equal(t, depsResult.Modules[0].Path, retrieved.Modules[0].Path)
		assert.Equal(t, depsResult.Modules[0].Version, retrieved.Modules[0].Version)
		assert.Equal(t, depsResult.Modules[0].Hash, retrieved.Modules[0].Hash)
		assert.Equal(t, depsResult.Success, retrieved.Success)
		assert.Equal(t, depsResult.Environment, retrieved.Environment)
	})

	t.Run("get non-existent dependency result", func(t *testing.T) {
		_, found := cache.GetDependencyResult("non-existent-deps-hash")
		assert.False(t, found)
	})
}

func TestBuildCache_GenerateHash(t *testing.T) {
	cache := NewBuildCache(t.TempDir())

	t.Run("same inputs produce same hash", func(t *testing.T) {
		hash1 := cache.GenerateHash("input1", "input2", "input3")
		hash2 := cache.GenerateHash("input1", "input2", "input3")

		assert.Equal(t, hash1, hash2)
		assert.Len(t, hash1, 16) // Should be first 16 chars of sha256
	})

	t.Run("different inputs produce different hashes", func(t *testing.T) {
		hash1 := cache.GenerateHash("input1", "input2")
		hash2 := cache.GenerateHash("input1", "input3")

		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("empty inputs", func(t *testing.T) {
		hash := cache.GenerateHash()
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 16)
	})

	t.Run("order matters", func(t *testing.T) {
		hash1 := cache.GenerateHash("a", "b", "c")
		hash2 := cache.GenerateHash("c", "b", "a")

		assert.NotEqual(t, hash1, hash2)
	})
}

func TestBuildCache_GenerateFileHash(t *testing.T) {
	tempDir := t.TempDir()
	cache := NewBuildCache(tempDir)

	// Create test files
	testFile1 := filepath.Join(tempDir, "test1.txt")
	testFile2 := filepath.Join(tempDir, "test2.txt")
	largeFile := filepath.Join(tempDir, "large.txt")

	err := os.WriteFile(testFile1, []byte("content1"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(testFile2, []byte("content2"), 0o600)
	require.NoError(t, err)

	// Create a large file (> 1MB) to test size limit
	largeContent := make([]byte, 2*1024*1024) // 2MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}
	err = os.WriteFile(largeFile, largeContent, 0o600)
	require.NoError(t, err)

	t.Run("hash existing files", func(t *testing.T) {
		hash1, err := cache.GenerateFileHash([]string{testFile1, testFile2})
		require.NoError(t, err)
		assert.NotEmpty(t, hash1)
		assert.Len(t, hash1, 16)

		// Same files should produce same hash
		hash2, err := cache.GenerateFileHash([]string{testFile1, testFile2})
		require.NoError(t, err)
		assert.Equal(t, hash1, hash2)

		// Different order should produce different hash
		hash3, err := cache.GenerateFileHash([]string{testFile2, testFile1})
		require.NoError(t, err)
		assert.NotEqual(t, hash1, hash3)
	})

	t.Run("hash with non-existent files", func(t *testing.T) {
		hash, err := cache.GenerateFileHash([]string{testFile1, "non-existent.txt", testFile2})
		require.NoError(t, err)
		assert.NotEmpty(t, hash) // Should still work, just skip non-existent files
	})

	t.Run("hash large files", func(t *testing.T) {
		hash, err := cache.GenerateFileHash([]string{largeFile})
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		// Large files should not include content in hash (only metadata)
	})

	t.Run("empty file list", func(t *testing.T) {
		hash, err := cache.GenerateFileHash([]string{})
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
	})

	t.Run("file modification affects hash", func(t *testing.T) {
		hash1, err := cache.GenerateFileHash([]string{testFile1})
		require.NoError(t, err)

		// Modify file content
		time.Sleep(time.Millisecond * 10) // Ensure different timestamp
		err = os.WriteFile(testFile1, []byte("modified content"), 0o600)
		require.NoError(t, err)

		hash2, err := cache.GenerateFileHash([]string{testFile1})
		require.NoError(t, err)

		assert.NotEqual(t, hash1, hash2)
	})
}

func TestBuildCache_GetStats(t *testing.T) {
	tempDir := t.TempDir()
	cache := NewBuildCache(tempDir)
	err := cache.Init()
	require.NoError(t, err)

	t.Run("enabled cache", func(t *testing.T) {
		// Add some test cache entries
		testResult := &BuildResult{
			Binary:   "/path/to/binary",
			Platform: "linux/amd64",
			Success:  true,
		}

		for i := 0; i < 3; i++ {
			hash := fmt.Sprintf("test-hash-%d", i)
			err = cache.StoreBuildResult(hash, testResult)
			require.NoError(t, err)
		}

		stats, err := cache.GetStats()
		require.NoError(t, err)
		require.NotNil(t, stats)

		assert.Positive(t, stats.TotalSize)
		assert.Equal(t, 3, stats.EntryCount)
		assert.False(t, stats.LastCleanup.IsZero())
	})

	t.Run("disabled cache", func(t *testing.T) {
		cache.SetOptions(false, 0, 0)

		stats, err := cache.GetStats()
		require.NoError(t, err)
		require.NotNil(t, stats)

		assert.Equal(t, int64(0), stats.TotalSize)
		assert.Equal(t, 0, stats.EntryCount)
	})
}

func TestBuildCache_Cleanup(t *testing.T) {
	tempDir := t.TempDir()
	cache := NewBuildCache(tempDir)
	cache.SetOptions(true, 1024*1024, 100*time.Millisecond) // Short TTL for testing
	err := cache.Init()
	require.NoError(t, err)

	t.Run("cleanup expired entries", func(t *testing.T) {
		// Create some cache entries
		testResult := &BuildResult{
			Binary:   "/path/to/binary",
			Platform: "linux/amd64",
			Success:  true,
		}

		for i := 0; i < 3; i++ {
			hash := fmt.Sprintf("cleanup-test-%d", i)
			err = cache.StoreBuildResult(hash, testResult)
			require.NoError(t, err)
		}

		// Wait for entries to expire
		time.Sleep(150 * time.Millisecond)

		// Cleanup should remove expired entries
		err = cache.Cleanup()
		require.NoError(t, err)

		// Verify entries were removed
		for i := 0; i < 3; i++ {
			hash := fmt.Sprintf("cleanup-test-%d", i)
			_, found := cache.GetBuildResult(hash)
			assert.False(t, found)
		}
	})

	t.Run("disabled cache", func(t *testing.T) {
		cache.SetOptions(false, 0, 0)
		err = cache.Cleanup()
		require.NoError(t, err)
	})
}

func TestBuildCache_Clear(t *testing.T) {
	tempDir := t.TempDir()
	cache := NewBuildCache(tempDir)
	err := cache.Init()
	require.NoError(t, err)

	t.Run("clear all cache", func(t *testing.T) {
		// Create a binary file for the test
		binaryPath := filepath.Join(tempDir, "clear-test-binary")
		err = os.WriteFile(binaryPath, []byte("binary content"), 0o600)
		require.NoError(t, err)

		// Add some test data
		testResult := &BuildResult{
			Binary:   binaryPath,
			Platform: "linux/amd64",
			Success:  true,
		}

		err = cache.StoreBuildResult("clear-test", testResult)
		require.NoError(t, err)

		// Verify it exists
		_, found := cache.GetBuildResult("clear-test")
		assert.True(t, found)

		// Clear cache
		err = cache.Clear()
		require.NoError(t, err)

		// After clear, the cache directory is gone, so need to reinit to use it
		err = cache.Init()
		require.NoError(t, err)

		// Verify cache entry is gone
		_, found = cache.GetBuildResult("clear-test")
		assert.False(t, found)
	})

	t.Run("disabled cache", func(t *testing.T) {
		cache := NewBuildCache(t.TempDir())
		cache.SetOptions(false, 0, 0)

		err = cache.Clear()
		assert.NoError(t, err)
	})
}

func TestBuildCache_IsEnabled(t *testing.T) {
	cache := NewBuildCache(t.TempDir())

	assert.True(t, cache.IsEnabled()) // Default is enabled

	cache.SetOptions(false, 0, 0)
	assert.False(t, cache.IsEnabled())

	cache.SetOptions(true, 0, 0)
	assert.True(t, cache.IsEnabled())
}

func TestBuildCache_GetCacheDir(t *testing.T) {
	tempDir := t.TempDir()
	cache := NewBuildCache(tempDir)

	assert.Equal(t, tempDir, cache.GetCacheDir())
}

func TestBuildCache_removeCacheEntry(t *testing.T) {
	tempDir := t.TempDir()
	cache := NewBuildCache(tempDir)
	err := cache.Init()
	require.NoError(t, err)

	// Create a test file
	testFile := filepath.Join(tempDir, "builds", "test.json")
	err = os.WriteFile(testFile, []byte(`{"test": "data"}`), 0o600)
	require.NoError(t, err)

	// Verify file exists
	assert.FileExists(t, testFile)

	// Remove it using the private method
	err = cache.removeCacheEntry(testFile)
	require.NoError(t, err)

	// Verify file is gone
	assert.NoFileExists(t, testFile)
}

func TestBuildCache_CacheEntryExpiration(t *testing.T) {
	tempDir := t.TempDir()
	cache := NewBuildCache(tempDir)
	err := cache.Init()
	require.NoError(t, err)

	// Create a build result with a binary file
	binaryPath := filepath.Join(tempDir, "test-binary")
	err = os.WriteFile(binaryPath, []byte("binary content"), 0o600)
	require.NoError(t, err)

	buildResult := &BuildResult{
		Binary:   binaryPath,
		Platform: "linux/amd64",
		Success:  true,
	}

	t.Run("valid binary file", func(t *testing.T) {
		err = cache.StoreBuildResult("binary-test", buildResult)
		require.NoError(t, err)

		// Should retrieve successfully
		retrieved, found := cache.GetBuildResult("binary-test")
		assert.True(t, found)
		assert.NotNil(t, retrieved)
	})

	t.Run("missing binary file", func(t *testing.T) {
		// Remove the binary file
		err = os.Remove(binaryPath)
		require.NoError(t, err)

		// Should not find the cache entry (binary missing)
		_, found := cache.GetBuildResult("binary-test")
		assert.False(t, found)
	})
}

func TestBuildCache_JSONMarshaling(t *testing.T) {
	// Test that all cache result types can be properly marshaled/unmarshaled

	t.Run("BuildResult JSON", func(t *testing.T) {
		result := &BuildResult{
			Hash:        "test-hash",
			Binary:      "/path/to/binary",
			Platform:    "linux/amd64",
			BuildFlags:  []string{"-ldflags=-s -w"},
			Environment: map[string]string{"GOOS": "linux"},
			Timestamp:   time.Now(),
			Success:     true,
			Metrics: BuildMetrics{
				CompileTime: time.Second,
				LinkTime:    500 * time.Millisecond,
				BinarySize:  1024 * 1024,
				SourceFiles: 10,
			},
		}

		data, err := json.Marshal(result)
		require.NoError(t, err)

		var unmarshaled BuildResult
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, result.Hash, unmarshaled.Hash)
		assert.Equal(t, result.Binary, unmarshaled.Binary)
		assert.Equal(t, result.Platform, unmarshaled.Platform)
		assert.Equal(t, result.BuildFlags, unmarshaled.BuildFlags)
	})

	t.Run("TestResult JSON", func(t *testing.T) {
		result := &TestResult{
			Hash:      "test-hash",
			Package:   "./pkg/test",
			Success:   true,
			Output:    "PASS",
			Coverage:  85.5,
			Duration:  time.Second,
			Timestamp: time.Now(),
			Metrics: TestMetrics{
				TestCount:   10,
				PassCount:   8,
				FailCount:   1,
				SkipCount:   1,
				CompileTime: 500 * time.Millisecond,
				ExecuteTime: 700 * time.Millisecond,
			},
		}

		data, err := json.Marshal(result)
		require.NoError(t, err)

		var unmarshaled TestResult
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, result.Hash, unmarshaled.Hash)
		assert.Equal(t, result.Package, unmarshaled.Package)
		assert.InEpsilon(t, result.Coverage, unmarshaled.Coverage, 0.001)
		// Skip timestamp comparison due to JSON precision issues
	})

	t.Run("LintResult JSON", func(t *testing.T) {
		result := &LintResult{
			Hash:    "lint-hash",
			Files:   []string{"main.go", "helper.go"},
			Success: false,
			Issues: []LintIssue{
				{
					File:     "main.go",
					Line:     10,
					Column:   5,
					Rule:     "gofmt",
					Severity: "error",
					Message:  "File is not gofmt-ed",
				},
			},
			Duration:  800 * time.Millisecond,
			Timestamp: time.Now(),
		}

		data, err := json.Marshal(result)
		require.NoError(t, err)

		var unmarshaled LintResult
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, result.Hash, unmarshaled.Hash)
		assert.Equal(t, result.Files, unmarshaled.Files)
		assert.Equal(t, result.Issues, unmarshaled.Issues)
	})
}

// Benchmark tests
func BenchmarkBuildCache_StoreBuildResult(b *testing.B) {
	tempDir := b.TempDir()
	cache := NewBuildCache(tempDir)
	err := cache.Init()
	if err != nil {
		b.Fatal(err)
	}

	buildResult := &BuildResult{
		Binary:      "/path/to/binary",
		Platform:    "linux/amd64",
		BuildFlags:  []string{"-ldflags=-s -w"},
		Environment: map[string]string{"GOOS": "linux"},
		Success:     true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash := fmt.Sprintf("bench-hash-%d", i)
		err := cache.StoreBuildResult(hash, buildResult)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBuildCache_GetBuildResult(b *testing.B) {
	tempDir := b.TempDir()
	cache := NewBuildCache(tempDir)
	err := cache.Init()
	if err != nil {
		b.Fatal(err)
	}

	// Create a dummy binary file
	binaryPath := filepath.Join(tempDir, "test-binary")
	err = os.WriteFile(binaryPath, []byte("binary content"), 0o600)
	if err != nil {
		b.Fatal(err)
	}

	// Pre-populate cache with fewer items for faster setup
	buildResult := &BuildResult{
		Binary:   binaryPath,
		Platform: "linux/amd64",
		Success:  true,
	}

	hashes := make([]string, 10)
	for i := 0; i < 10; i++ {
		hash := fmt.Sprintf("bench-get-hash-%d", i)
		hashes[i] = hash
		err := cache.StoreBuildResult(hash, buildResult)
		if err != nil {
			b.Fatalf("Failed to store result %d: %v", i, err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash := hashes[i%10]
		_, found := cache.GetBuildResult(hash)
		if !found {
			b.Fatalf("Expected to find cached result for hash: %s", hash)
		}
	}
}

func BenchmarkBuildCache_GenerateHash(b *testing.B) {
	cache := NewBuildCache(b.TempDir())
	inputs := []string{"input1", "input2", "input3", "input4", "input5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cache.GenerateHash(inputs...)
	}
}

func BenchmarkBuildCache_GenerateFileHash(b *testing.B) {
	tempDir := b.TempDir()
	cache := NewBuildCache(tempDir)

	// Create test files
	files := make([]string, 10)
	for i := 0; i < 10; i++ {
		file := filepath.Join(tempDir, fmt.Sprintf("test%d.go", i))
		err := os.WriteFile(file, []byte(fmt.Sprintf("package main\nfunc test%d() {}", i)), 0o600)
		if err != nil {
			b.Fatal(err)
		}
		files[i] = file
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := cache.GenerateFileHash(files)
		if err != nil {
			b.Fatal(err)
		}
	}
}
