package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, ".mage-cache", config.Directory)
	assert.Equal(t, int64(5*1024*1024*1024), config.MaxSize) // 5GB
	assert.Equal(t, 7*24*time.Hour, config.TTL)              // 7 days
	assert.True(t, config.Compression)

	// Check strategies
	require.NotNil(t, config.Strategies)
	assert.Equal(t, "file_hash", config.Strategies.Build)
	assert.Equal(t, "content_hash", config.Strategies.Test)
	assert.Equal(t, "file_hash+config_hash", config.Strategies.Lint)
	assert.Equal(t, "version_hash", config.Strategies.Deps)
}

func TestNewManager(t *testing.T) {
	t.Run("with default config", func(t *testing.T) {
		manager := NewManager(nil)

		assert.NotNil(t, manager)
		assert.NotNil(t, manager.config)
		assert.NotNil(t, manager.buildCache)
		assert.NotNil(t, manager.fileOps)
		assert.Equal(t, ".mage-cache", manager.config.Directory)
	})

	t.Run("with custom config", func(t *testing.T) {
		tempDir := t.TempDir()
		config := &Config{
			Enabled:   true,
			Directory: tempDir,
			MaxSize:   1024 * 1024, // 1MB
			TTL:       time.Hour,
		}

		manager := NewManager(config)

		assert.NotNil(t, manager)
		assert.Equal(t, config, manager.config)
		assert.Equal(t, tempDir, manager.cacheDir)
	})

	t.Run("with relative path", func(t *testing.T) {
		config := &Config{
			Enabled:   true,
			Directory: "test-cache",
		}

		manager := NewManager(config)

		// Cache dir should be converted to absolute path
		assert.True(t, filepath.IsAbs(manager.cacheDir))
		assert.Contains(t, manager.cacheDir, "test-cache")
	})
}

func TestManager_Init(t *testing.T) {
	t.Run("enabled manager", func(t *testing.T) {
		tempDir := t.TempDir()
		config := &Config{
			Enabled:   true,
			Directory: tempDir,
			MaxSize:   1024 * 1024,
			TTL:       time.Hour,
		}

		manager := NewManager(config)
		err := manager.Init()

		require.NoError(t, err)

		// Check that cache directory was created
		assert.DirExists(t, tempDir)

		// Check that build cache subdirectories exist
		subdirs := []string{"builds", "tests", "lint", "deps", "tools", "meta"}
		for _, subdir := range subdirs {
			assert.DirExists(t, filepath.Join(tempDir, subdir))
		}
	})

	t.Run("disabled manager", func(t *testing.T) {
		tempDir := t.TempDir()
		config := &Config{
			Enabled:   false,
			Directory: tempDir,
		}

		manager := NewManager(config)
		err := manager.Init()

		assert.NoError(t, err)
		// No directories should be created when disabled
	})

	t.Run("invalid directory", func(t *testing.T) {
		config := &Config{
			Enabled:   true,
			Directory: "/invalid/readonly/path",
		}

		manager := NewManager(config)
		err := manager.Init()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create cache directory")
	})
}

func TestManager_GetBuildCache(t *testing.T) {
	tempDir := t.TempDir()
	config := &Config{
		Enabled:   true,
		Directory: tempDir,
	}

	manager := NewManager(config)
	buildCache := manager.GetBuildCache()

	assert.NotNil(t, buildCache)
	assert.Equal(t, manager.buildCache, buildCache)
}

func TestManager_GenerateBuildHash(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	testFile1 := filepath.Join(tempDir, "test1.go")
	testFile2 := filepath.Join(tempDir, "test2.go")
	configFile := filepath.Join(tempDir, "config.yaml")

	err := os.WriteFile(testFile1, []byte("package main\nfunc main() {}"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(testFile2, []byte("package main\nvar x = 1"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(configFile, []byte("config: value"), 0o600)
	require.NoError(t, err)

	config := &Config{
		Enabled:   true,
		Directory: tempDir,
		Strategies: &Strategies{
			Build: "file_hash",
		},
	}

	manager := NewManager(config)
	err = manager.Init()
	require.NoError(t, err)

	t.Run("file_hash strategy", func(t *testing.T) {
		hash1, err := manager.GenerateBuildHash(
			"linux/amd64",
			"-ldflags=-s -w",
			[]string{testFile1, testFile2},
			[]string{configFile},
		)
		require.NoError(t, err)
		assert.NotEmpty(t, hash1)

		// Same inputs should produce same hash
		hash2, err := manager.GenerateBuildHash(
			"linux/amd64",
			"-ldflags=-s -w",
			[]string{testFile1, testFile2},
			[]string{configFile},
		)
		require.NoError(t, err)
		assert.Equal(t, hash1, hash2)

		// Different platform should produce different hash
		hash3, err := manager.GenerateBuildHash(
			"darwin/amd64",
			"-ldflags=-s -w",
			[]string{testFile1, testFile2},
			[]string{configFile},
		)
		require.NoError(t, err)
		assert.NotEqual(t, hash1, hash3)
	})

	t.Run("content_hash strategy", func(t *testing.T) {
		manager.config.Strategies.Build = "content_hash"

		hash1, err := manager.GenerateBuildHash(
			"linux/amd64",
			"-ldflags=-s -w",
			[]string{testFile1},
			[]string{},
		)
		require.NoError(t, err)
		assert.NotEmpty(t, hash1)

		// Modify file content
		err = os.WriteFile(testFile1, []byte("package main\nfunc main() { println(\"hello\") }"), 0o600)
		require.NoError(t, err)

		hash2, err := manager.GenerateBuildHash(
			"linux/amd64",
			"-ldflags=-s -w",
			[]string{testFile1},
			[]string{},
		)
		require.NoError(t, err)
		assert.NotEqual(t, hash1, hash2) // Content changed, hash should be different
	})

	t.Run("timestamp_hash strategy", func(t *testing.T) {
		manager.config.Strategies.Build = "timestamp_hash"

		hash1, err := manager.GenerateBuildHash(
			"linux/amd64",
			"-ldflags=-s -w",
			[]string{testFile1},
			[]string{},
		)
		require.NoError(t, err)
		assert.NotEmpty(t, hash1)

		// Touch file to change timestamp (wait longer to ensure timestamp change)
		time.Sleep(time.Second)
		currentTime := time.Now()
		err = os.Chtimes(testFile1, currentTime, currentTime)
		require.NoError(t, err)

		hash2, err := manager.GenerateBuildHash(
			"linux/amd64",
			"-ldflags=-s -w",
			[]string{testFile1},
			[]string{},
		)
		require.NoError(t, err)
		assert.NotEqual(t, hash1, hash2) // Timestamp changed, hash should be different
	})

	t.Run("unknown strategy defaults to file_hash", func(t *testing.T) {
		manager.config.Strategies.Build = "unknown_strategy"

		hash, err := manager.GenerateBuildHash(
			"linux/amd64",
			"-ldflags=-s -w",
			[]string{testFile1},
			[]string{},
		)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
	})
}

func TestManager_GenerateTestHash(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	testFile := filepath.Join(tempDir, "test.go")
	err := os.WriteFile(testFile, []byte("package main\nfunc TestExample(t *testing.T) {}"), 0o600)
	require.NoError(t, err)

	config := &Config{
		Enabled:   true,
		Directory: tempDir,
	}

	manager := NewManager(config)
	err = manager.Init()
	require.NoError(t, err)

	hash1, err := manager.GenerateTestHash(
		"./pkg/test",
		[]string{testFile},
		[]string{"-v", "-race"},
	)
	require.NoError(t, err)
	assert.NotEmpty(t, hash1)

	// Same inputs should produce same hash
	hash2, err := manager.GenerateTestHash(
		"./pkg/test",
		[]string{testFile},
		[]string{"-v", "-race"},
	)
	require.NoError(t, err)
	assert.Equal(t, hash1, hash2)

	// Different flags should produce different hash
	hash3, err := manager.GenerateTestHash(
		"./pkg/test",
		[]string{testFile},
		[]string{"-v"},
	)
	require.NoError(t, err)
	assert.NotEqual(t, hash1, hash3)
}

func TestManager_GenerateLintHash(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	sourceFile := filepath.Join(tempDir, "main.go")
	configFile := filepath.Join(tempDir, ".golangci.yml")

	err := os.WriteFile(sourceFile, []byte("package main\nfunc main() {}"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(configFile, []byte("linters:\n  enable:\n    - gofmt"), 0o600)
	require.NoError(t, err)

	config := &Config{
		Enabled:   true,
		Directory: tempDir,
	}

	manager := NewManager(config)
	err = manager.Init()
	require.NoError(t, err)

	hash1, err := manager.GenerateLintHash(
		[]string{sourceFile},
		[]string{configFile},
		"gofmt,govet",
	)
	require.NoError(t, err)
	assert.NotEmpty(t, hash1)

	// Same inputs should produce same hash
	hash2, err := manager.GenerateLintHash(
		[]string{sourceFile},
		[]string{configFile},
		"gofmt,govet",
	)
	require.NoError(t, err)
	assert.Equal(t, hash1, hash2)

	// Different lint config should produce different hash
	hash3, err := manager.GenerateLintHash(
		[]string{sourceFile},
		[]string{configFile},
		"gofmt,govet,ineffassign",
	)
	require.NoError(t, err)
	assert.NotEqual(t, hash1, hash3)
}

func TestManager_GenerateDependencyHash(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	modFile := filepath.Join(tempDir, "go.mod")
	sumFile := filepath.Join(tempDir, "go.sum")

	err := os.WriteFile(modFile, []byte("module test\ngo 1.19"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(sumFile, []byte("github.com/example/lib v1.0.0 h1:hash"), 0o600)
	require.NoError(t, err)

	config := &Config{
		Enabled:   true,
		Directory: tempDir,
	}

	manager := NewManager(config)
	err = manager.Init()
	require.NoError(t, err)

	env := map[string]string{
		"GOPROXY": "https://proxy.golang.org",
		"GOSUMDB": "sum.golang.org",
		"PATH":    "/usr/bin:/bin",
		"HOME":    "/home/user",
	}

	hash1, err := manager.GenerateDependencyHash(modFile, sumFile, env)
	require.NoError(t, err)
	assert.NotEmpty(t, hash1)

	// Same inputs should produce same hash (create new map to test stability)
	env1Copy := map[string]string{
		"GOPROXY": "https://proxy.golang.org",
		"GOSUMDB": "sum.golang.org",
		"PATH":    "/usr/bin:/bin",
		"HOME":    "/home/user",
	}
	hash2, err := manager.GenerateDependencyHash(modFile, sumFile, env1Copy)
	require.NoError(t, err)
	assert.Equal(t, hash1, hash2)

	// Different environment should produce different hash
	env2 := map[string]string{
		"GOPROXY": "direct",
		"GOSUMDB": "sum.golang.org",
		"PATH":    "/usr/bin:/bin",
		"HOME":    "/home/user",
	}
	hash3, err := manager.GenerateDependencyHash(modFile, sumFile, env2)
	require.NoError(t, err)
	assert.NotEqual(t, hash1, hash3)
}

func TestManager_Cleanup(t *testing.T) {
	t.Run("enabled manager", func(t *testing.T) {
		tempDir := t.TempDir()
		config := &Config{
			Enabled:   true,
			Directory: tempDir,
			MaxSize:   1024, // Small max size to trigger size-based cleanup
		}

		manager := NewManager(config)
		err := manager.Init()
		require.NoError(t, err)

		err = manager.Cleanup()
		assert.NoError(t, err)
	})

	t.Run("disabled manager", func(t *testing.T) {
		config := &Config{
			Enabled: false,
		}

		manager := NewManager(config)
		err := manager.Cleanup()
		assert.NoError(t, err)
	})
}

func TestManager_GetStats(t *testing.T) {
	t.Run("enabled manager", func(t *testing.T) {
		tempDir := t.TempDir()
		config := &Config{
			Enabled:   true,
			Directory: tempDir,
		}

		manager := NewManager(config)
		err := manager.Init()
		require.NoError(t, err)

		stats, err := manager.GetStats()
		require.NoError(t, err)
		assert.NotNil(t, stats)
	})

	t.Run("disabled manager", func(t *testing.T) {
		config := &Config{
			Enabled: false,
		}

		manager := NewManager(config)
		stats, err := manager.GetStats()
		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, int64(0), stats.TotalSize)
		assert.Equal(t, 0, stats.EntryCount)
	})
}

func TestManager_Clear(t *testing.T) {
	t.Run("enabled manager", func(t *testing.T) {
		tempDir := t.TempDir()
		config := &Config{
			Enabled:   true,
			Directory: tempDir,
		}

		manager := NewManager(config)
		err := manager.Init()
		require.NoError(t, err)

		// Create some test cache files
		testFile := filepath.Join(tempDir, "builds", "test.json")
		err = os.WriteFile(testFile, []byte(`{"test": "data"}`), 0o600)
		require.NoError(t, err)

		err = manager.Clear()
		require.NoError(t, err)

		// Cache directory should not exist after clear
		assert.NoFileExists(t, testFile)
	})

	t.Run("disabled manager", func(t *testing.T) {
		config := &Config{
			Enabled: false,
		}

		manager := NewManager(config)
		err := manager.Clear()
		require.NoError(t, err)
	})
}

func TestManager_IsEnabled(t *testing.T) {
	t.Run("enabled", func(t *testing.T) {
		config := &Config{Enabled: true}
		manager := NewManager(config)
		assert.True(t, manager.IsEnabled())
	})

	t.Run("disabled", func(t *testing.T) {
		config := &Config{Enabled: false}
		manager := NewManager(config)
		assert.False(t, manager.IsEnabled())
	})
}

func TestManager_GetCacheDir(t *testing.T) {
	tempDir := t.TempDir()
	config := &Config{
		Directory: tempDir,
	}

	manager := NewManager(config)
	assert.Equal(t, tempDir, manager.GetCacheDir())
}

func TestManager_WarmCache(t *testing.T) {
	tempDir := t.TempDir()
	config := &Config{
		Enabled:   true,
		Directory: tempDir,
	}

	manager := NewManager(config)
	err := manager.Init()
	require.NoError(t, err)

	operations := []string{"build", "test", "lint"}
	err = manager.WarmCache(operations)
	require.NoError(t, err)

	// Test with disabled manager
	config.Enabled = false
	manager = NewManager(config)
	err = manager.WarmCache(operations)
	require.NoError(t, err)
}

func TestManager_HashGenerationMethods(t *testing.T) {
	tempDir := t.TempDir()

	// Create test files
	sourceFile := filepath.Join(tempDir, "main.go")
	configFile := filepath.Join(tempDir, "config.json")

	err := os.WriteFile(sourceFile, []byte("package main\nfunc main() {}"), 0o600)
	require.NoError(t, err)
	err = os.WriteFile(configFile, []byte(`{"key": "value"}`), 0o600)
	require.NoError(t, err)

	config := &Config{
		Enabled:   true,
		Directory: tempDir,
	}

	manager := NewManager(config)
	err = manager.Init()
	require.NoError(t, err)

	sourceFiles := []string{sourceFile}
	configFiles := []string{configFile}
	platform := runtime.GOOS + "/" + runtime.GOARCH
	ldflags := "-s -w"

	t.Run("generateFileBasedHash", func(t *testing.T) {
		hash, err := manager.generateFileBasedHash(sourceFiles, configFiles, platform, ldflags)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 16) // Should be truncated to 16 chars by GenerateHash
	})

	t.Run("generateContentBasedHash", func(t *testing.T) {
		hash, err := manager.generateContentBasedHash(sourceFiles, configFiles, platform, ldflags)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 16)
	})

	t.Run("generateTimestampBasedHash", func(t *testing.T) {
		hash, err := manager.generateTimestampBasedHash(sourceFiles, configFiles, platform, ldflags)
		require.NoError(t, err)
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 16)
	})
}

func TestManager_performSizeBasedCleanup(t *testing.T) {
	tempDir := t.TempDir()
	config := &Config{
		Enabled:   true,
		Directory: tempDir,
		MaxSize:   1024,
	}

	manager := NewManager(config)
	err := manager.Init()
	require.NoError(t, err)

	// Create some cache files to exceed the size limit
	for i := 0; i < 5; i++ {
		testFile := filepath.Join(tempDir, "builds", fmt.Sprintf("test%d.json", i))
		err = os.WriteFile(testFile, make([]byte, 300), 0o600) // 300 bytes each
		require.NoError(t, err)
	}

	err = manager.performSizeBasedCleanup(500) // Free 500 bytes
	assert.NoError(t, err)
}

// Benchmark tests
func BenchmarkManager_GenerateBuildHash(b *testing.B) {
	tempDir := b.TempDir()

	// Create test files
	sourceFile := filepath.Join(tempDir, "main.go")
	err := os.WriteFile(sourceFile, []byte("package main\nfunc main() {}"), 0o600)
	if err != nil {
		b.Fatal(err)
	}

	config := &Config{
		Enabled:   true,
		Directory: tempDir,
		Strategies: &Strategies{
			Build: "file_hash",
		},
	}

	manager := NewManager(config)
	err = manager.Init()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GenerateBuildHash(
			"linux/amd64",
			"-ldflags=-s -w",
			[]string{sourceFile},
			[]string{},
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkManager_GenerateTestHash(b *testing.B) {
	tempDir := b.TempDir()

	// Create test files
	testFile := filepath.Join(tempDir, "test.go")
	err := os.WriteFile(testFile, []byte("package main\nfunc TestExample(t *testing.T) {}"), 0o600)
	if err != nil {
		b.Fatal(err)
	}

	manager := NewManager(&Config{
		Enabled:   true,
		Directory: tempDir,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GenerateTestHash(
			"./pkg/test",
			[]string{testFile},
			[]string{"-v"},
		)
		if err != nil {
			b.Fatal(err)
		}
	}
}
