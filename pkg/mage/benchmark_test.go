package mage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mrz1836/go-mage/pkg/common/fileops"
)

// BenchmarkCommandExecution benchmarks secure command execution
func BenchmarkCommandExecution(b *testing.B) {
	runner := NewSecureCommandRunner()

	b.Run("echo", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := runner.RunCmd("echo", "test"); err != nil {
				b.Logf("RunCmd error (expected in benchmark): %v", err)
			}
		}
	})

	b.Run("echo_with_output", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := runner.RunCmdOutput("echo", "test"); err != nil {
				b.Logf("RunCmdOutput error (expected in benchmark): %v", err)
			}
		}
	})

	b.Run("true_command", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if err := runner.RunCmd("true"); err != nil {
				b.Logf("RunCmd error (expected in benchmark): %v", err)
			}
		}
	})
}

// BenchmarkConfigOperations benchmarks configuration operations
func BenchmarkConfigOperations(b *testing.B) {
	// Setup
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			b.Logf("Failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change directory: %v", err)
	}

	// Create a config file
	configYAML := `
project:
  name: bench-test
  binary: bench
  module: github.com/test/bench

build:
  output: bin
  platforms:
    - linux/amd64
    - darwin/amd64
    - windows/amd64

test:
  parallel: true
  cover: true
  race: false

lint:
  timeout: 5m
`
	if err := os.WriteFile(".mage.yaml", []byte(configYAML), 0o600); err != nil {
		b.Fatalf("Failed to write config file: %v", err)
	}

	b.Run("LoadConfig", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Reset config for benchmark
			_, _ = LoadConfig() //nolint:errcheck // Benchmark ignores errors
		}
	})

	b.Run("SaveConfig", func(b *testing.B) {
		config, _ := LoadConfig() //nolint:errcheck // Benchmark ignores errors
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = SaveConfig(config) //nolint:errcheck // Benchmark ignores errors
		}
	})

	b.Run("GetConfig", func(b *testing.B) {
		// Pre-load config
		testCfg, _ := LoadConfig() //nolint:errcheck // Benchmark ignores errors
		_ = testCfg
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = GetConfig() //nolint:errcheck // Benchmark ignores errors
		}
	})
}

// BenchmarkNamespaceCreation benchmarks namespace factory functions
func BenchmarkNamespaceCreation(b *testing.B) {
	benchmarks := []struct {
		name    string
		factory func() interface{}
	}{
		{"Build", func() interface{} { return NewBuildNamespace() }},
		{"Test", func() interface{} { return NewTestNamespace() }},
		{"Lint", func() interface{} { return NewLintNamespace() }},
		{"Format", func() interface{} { return NewFormatNamespace() }},
		{"Deps", func() interface{} { return NewDepsNamespace() }},
		{"Git", func() interface{} { return NewGitNamespace() }},
		{"Tools", func() interface{} { return NewToolsNamespace() }},
		{"Generate", func() interface{} { return NewGenerateNamespace() }},
		{"CLI", func() interface{} { return NewCLINamespace() }},
		{"Update", func() interface{} { return NewUpdateNamespace() }},
		{"Mod", func() interface{} { return NewModNamespace() }},
		{"Metrics", func() interface{} { return NewMetricsNamespace() }},
		{"Workflow", func() interface{} { return NewWorkflowNamespace() }},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = bm.factory()
			}
		})
	}
}

// BenchmarkFileOperations benchmarks file operations
func BenchmarkFileOperations(b *testing.B) {
	tempDir := b.TempDir()
	fileOps := fileops.New()

	testData := createBenchmarkTestData()
	runWriteBenchmarks(b, tempDir, fileOps, testData)
	runReadBenchmarks(b, tempDir, fileOps, testData)
}

// benchmarkTestData holds test data for benchmarks
type benchmarkTestData struct {
	small  []byte
	medium []byte
	large  []byte
}

// createBenchmarkTestData creates test data of different sizes
func createBenchmarkTestData() benchmarkTestData {
	smallData := []byte("small test data")
	mediumData := make([]byte, 1024)     // 1KB
	largeData := make([]byte, 1024*1024) // 1MB

	for i := range mediumData {
		mediumData[i] = byte(i % 256)
	}
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	return benchmarkTestData{
		small:  smallData,
		medium: mediumData,
		large:  largeData,
	}
}

// runWriteBenchmarks runs all write benchmarks
func runWriteBenchmarks(b *testing.B, tempDir string, fileOps *fileops.FileOps, data benchmarkTestData) {
	b.Run("WriteSmallFile", func(b *testing.B) {
		benchmarkWriteFile(b, tempDir, fileOps, "small.txt", data.small)
	})

	b.Run("WriteMediumFile", func(b *testing.B) {
		benchmarkWriteFile(b, tempDir, fileOps, "medium.txt", data.medium)
	})

	b.Run("WriteLargeFile", func(b *testing.B) {
		benchmarkWriteFile(b, tempDir, fileOps, "large.txt", data.large)
	})
}

// benchmarkWriteFile benchmarks writing a file of specific size
func benchmarkWriteFile(b *testing.B, tempDir string, fileOps *fileops.FileOps, filename string, data []byte) {
	for i := 0; i < b.N; i++ {
		path := filepath.Join(tempDir, filename)
		if err := fileOps.File.WriteFile(path, data, 0o644); err != nil {
			b.Logf("WriteFile error: %v", err)
		}
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			b.Logf("Remove error: %v", err)
		}
	}
}

// runReadBenchmarks runs all read benchmarks
func runReadBenchmarks(b *testing.B, tempDir string, fileOps *fileops.FileOps, data benchmarkTestData) {
	readPaths := setupReadFiles(b, tempDir, fileOps, data)

	b.Run("ReadSmallFile", func(b *testing.B) {
		benchmarkReadFile(b, fileOps, readPaths.small)
	})

	b.Run("ReadMediumFile", func(b *testing.B) {
		benchmarkReadFile(b, fileOps, readPaths.medium)
	})

	b.Run("ReadLargeFile", func(b *testing.B) {
		benchmarkReadFile(b, fileOps, readPaths.large)
	})
}

// readFilePaths holds paths to read test files
type readFilePaths struct {
	small  string
	medium string
	large  string
}

// setupReadFiles creates files for read benchmarks
func setupReadFiles(b *testing.B, tempDir string, fileOps *fileops.FileOps, data benchmarkTestData) readFilePaths {
	paths := readFilePaths{
		small:  filepath.Join(tempDir, "read_small.txt"),
		medium: filepath.Join(tempDir, "read_medium.txt"),
		large:  filepath.Join(tempDir, "read_large.txt"),
	}

	if err := fileOps.File.WriteFile(paths.small, data.small, 0o644); err != nil {
		b.Fatalf("Failed to write small file: %v", err)
	}
	if err := fileOps.File.WriteFile(paths.medium, data.medium, 0o644); err != nil {
		b.Fatalf("Failed to write medium file: %v", err)
	}
	if err := fileOps.File.WriteFile(paths.large, data.large, 0o644); err != nil {
		b.Fatalf("Failed to write large file: %v", err)
	}

	return paths
}

// benchmarkReadFile benchmarks reading a specific file
func benchmarkReadFile(b *testing.B, fileOps *fileops.FileOps, path string) {
	for i := 0; i < b.N; i++ {
		if _, err := fileOps.File.ReadFile(path); err != nil {
			b.Logf("ReadFile error: %v", err)
		}
	}
}

// BenchmarkStringOperations benchmarks string manipulation functions
func BenchmarkStringOperations(b *testing.B) {
	shortString := "hello"
	mediumString := "this is a medium length string for testing purposes"
	longString := ""
	for i := 0; i < 100; i++ {
		longString += "this is a very long string that will be used for benchmarking "
	}

	b.Run("truncateString_short", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = truncateString(shortString, 10)
		}
	})

	b.Run("truncateString_medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = truncateString(mediumString, 20)
		}
	})

	b.Run("truncateString_long", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = truncateString(longString, 50)
		}
	})

	b.Run("formatBytes_small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = formatBytes(512)
		}
	})

	b.Run("formatBytes_medium", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = formatBytes(1024 * 1024)
		}
	})

	b.Run("formatBytes_large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = formatBytes(1024 * 1024 * 1024)
		}
	})
}

// BenchmarkBuildOperations benchmarks build-related operations
func BenchmarkBuildOperations(b *testing.B) {
	// Setup
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			b.Logf("Failed to restore working directory: %v", err)
		}
	}()
	if err := os.Chdir(tempDir); err != nil {
		b.Fatalf("Failed to change directory: %v", err)
	}

	// Create a simple project
	if err := os.WriteFile("go.mod", []byte("module bench\n\ngo 1.24\n"), 0o600); err != nil {
		b.Fatalf("Failed to write go.mod: %v", err)
	}
	if err := os.WriteFile("main.go", []byte("package main\nfunc main() {}\n"), 0o600); err != nil {
		b.Fatalf("Failed to write main.go: %v", err)
	}

	cfg = &Config{
		Project: ProjectConfig{
			Binary: "bench",
			Module: "bench",
		},
		Build: BuildConfig{
			Output: "bin",
		},
	}

	b.Run("buildFlags", func(b *testing.B) {
		config, _ := GetConfig() //nolint:errcheck // Benchmark ignores errors
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = buildFlags(config)
		}
	})

	b.Run("getVersion", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = getVersion()
		}
	})

	b.Run("getCommit", func(b *testing.B) {
		// This might fail if not in a git repo, but benchmark the attempt
		for i := 0; i < b.N; i++ {
			_ = getCommit()
		}
	})
}

// BenchmarkCLIOperations benchmarks CLI-related operations
func BenchmarkCLIOperations(b *testing.B) {
	// Create test data
	results := []BatchOperationResult{}
	for i := 0; i < 10; i++ {
		results = append(results, BatchOperationResult{
			Operation: BatchOperation{
				Name: "test-op",
			},
			Success:  i%2 == 0,
			Duration: time.Duration(i) * time.Millisecond,
		})
	}

	b.Run("calculateBatchStats", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = calculateBatchStats(results)
		}
	})

	b.Run("formatStatus", func(b *testing.B) {
		formatter := newBatchResultFormatter()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = formatter.formatStatus(true)
			_ = formatter.formatStatus(false)
		}
	})
}

// BenchmarkEnterpriseOperations benchmarks enterprise config operations
func BenchmarkEnterpriseOperations(b *testing.B) {
	config := NewEnterpriseConfiguration()

	b.Run("NewEnterpriseConfiguration", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = NewEnterpriseConfiguration()
		}
	})

	b.Run("ValidateConfiguration", func(b *testing.B) {
		validator := NewConfigurationValidator()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = validator.Validate(config)
		}
	})
}

// BenchmarkPlatformParsing benchmarks platform string parsing
func BenchmarkPlatformParsing(b *testing.B) {
	platforms := []string{
		"linux/amd64",
		"darwin/amd64",
		"darwin/arm64",
		"windows/amd64",
		"linux/arm64",
	}

	b.Run("parsePlatform", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, p := range platforms {
				_, _, _ = parsePlatform(p) //nolint:errcheck // Benchmark ignores errors
			}
		}
	})

	b.Run("generateOutputPath", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, p := range platforms {
				goos, goarch, _ := parsePlatform(p) //nolint:errcheck // Benchmark ignores errors
				_ = generateOutputPath("myapp", goos, goarch)
			}
		}
	})
}

// BenchmarkConcurrentOperations benchmarks concurrent execution patterns
func BenchmarkConcurrentOperations(b *testing.B) {
	b.Run("parallel_namespace_creation", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = NewBuildNamespace()
				_ = NewTestNamespace()
				_ = NewLintNamespace()
			}
		})
	})

	b.Run("parallel_config_access", func(b *testing.B) {
		// Pre-load config
		testCfg, _ := LoadConfig() //nolint:errcheck // Benchmark ignores errors
		_ = testCfg
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = GetConfig() //nolint:errcheck // Benchmark ignores errors
				_ = BinaryName()
				_ = IsVerbose()
			}
		})
	})
}

// BenchmarkMemoryAllocation benchmarks memory allocation patterns
func BenchmarkMemoryAllocation(b *testing.B) {
	b.Run("namespace_allocation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = NewBuildNamespace()
		}
	})

	b.Run("config_allocation", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = defaultConfig()
		}
	})

	b.Run("command_handler_allocation", func(b *testing.B) {
		dashboard := Dashboard{}
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = newDashboardCommandHandler(&dashboard)
		}
	})
}
