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
			_ = runner.RunCmd("echo", "test")
		}
	})

	b.Run("echo_with_output", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = runner.RunCmdOutput("echo", "test")
		}
	})

	b.Run("true_command", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = runner.RunCmd("true")
		}
	})
}

// BenchmarkConfigOperations benchmarks configuration operations
func BenchmarkConfigOperations(b *testing.B) {
	// Setup
	tempDir := b.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

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
	os.WriteFile(".mage.yaml", []byte(configYAML), 0644)

	b.Run("LoadConfig", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			cfg = nil
			_, _ = LoadConfig()
		}
	})

	b.Run("SaveConfig", func(b *testing.B) {
		config, _ := LoadConfig()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = SaveConfig(config)
		}
	})

	b.Run("GetConfig", func(b *testing.B) {
		// Pre-load config
		cfg, _ = LoadConfig()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = GetConfig()
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

	// Create test data
	smallData := []byte("small test data")
	mediumData := make([]byte, 1024)     // 1KB
	largeData := make([]byte, 1024*1024) // 1MB

	for i := range mediumData {
		mediumData[i] = byte(i % 256)
	}
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	b.Run("WriteSmallFile", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			path := filepath.Join(tempDir, "small.txt")
			_ = fileOps.File.WriteFile(path, smallData, 0644)
			os.Remove(path)
		}
	})

	b.Run("WriteMediumFile", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			path := filepath.Join(tempDir, "medium.txt")
			_ = fileOps.File.WriteFile(path, mediumData, 0644)
			os.Remove(path)
		}
	})

	b.Run("WriteLargeFile", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			path := filepath.Join(tempDir, "large.txt")
			_ = fileOps.File.WriteFile(path, largeData, 0644)
			os.Remove(path)
		}
	})

	// Create files for reading
	smallPath := filepath.Join(tempDir, "read_small.txt")
	mediumPath := filepath.Join(tempDir, "read_medium.txt")
	largePath := filepath.Join(tempDir, "read_large.txt")

	fileOps.File.WriteFile(smallPath, smallData, 0644)
	fileOps.File.WriteFile(mediumPath, mediumData, 0644)
	fileOps.File.WriteFile(largePath, largeData, 0644)

	b.Run("ReadSmallFile", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = fileOps.File.ReadFile(smallPath)
		}
	})

	b.Run("ReadMediumFile", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = fileOps.File.ReadFile(mediumPath)
		}
	})

	b.Run("ReadLargeFile", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = fileOps.File.ReadFile(largePath)
		}
	})
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
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Create a simple project
	os.WriteFile("go.mod", []byte("module bench\n\ngo 1.21\n"), 0644)
	os.WriteFile("main.go", []byte("package main\nfunc main() {}\n"), 0644)

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
		config, _ := GetConfig()
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
				_, _, _ = parsePlatform(p)
			}
		}
	})

	b.Run("generateOutputPath", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, p := range platforms {
				goos, goarch, _ := parsePlatform(p)
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
		cfg, _ = LoadConfig()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_, _ = GetConfig()
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
			_ = newDashboardCommandHandler(dashboard)
		}
	})
}
