package benchmarks

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"testing"
	"time"

	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

// TestMain sets up benchmarking environment
func TestMain(m *testing.M) {
	// Build magex binary for benchmarking
	if err := buildMagexBinary(); err != nil {
		fmt.Printf("Failed to build magex binary: %v\n", err)
		os.Exit(1)
	}

	// Run benchmarks
	code := m.Run()

	// Cleanup
	cleanup()

	os.Exit(code)
}

func buildMagexBinary() error {
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "go", "build", "-o", "magex", "../../cmd/magex")
	cmd.Env = append(os.Environ(), "GO111MODULE=on")
	return cmd.Run()
}

func cleanup() {
	if err := os.Remove("magex"); err != nil {
		// Cleanup error is non-critical in tests
		_ = err
	}
}

// BenchmarkRegistryOperations benchmarks core registry operations
func BenchmarkRegistryOperations(b *testing.B) {
	b.Run("Register", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r := registry.NewRegistry()
			cmd, err := registry.NewCommand(fmt.Sprintf("cmd%d", i)).
				WithDescription("Test command").
				WithFunc(func() error { return nil }).
				WithCategory("Benchmark").
				Build()
			if err != nil {
				b.Fatalf("Failed to build command: %v", err)
			}
			if err := r.Register(cmd); err != nil {
				b.Fatalf("Failed to register command: %v", err)
			}
		}
	})

	b.Run("Get", func(b *testing.B) {
		r := registry.NewRegistry()
		// Pre-populate with commands
		for i := 0; i < 100; i++ {
			cmd, err := registry.NewCommand(fmt.Sprintf("cmd%d", i)).
				WithDescription("Test command").
				WithFunc(func() error { return nil }).
				WithCategory("Benchmark").
				Build()
			if err != nil {
				b.Fatalf("Failed to build command: %v", err)
			}
			r.MustRegister(cmd)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.Get(fmt.Sprintf("cmd%d", i%100))
		}
	})

	b.Run("List", func(b *testing.B) {
		r := registry.NewRegistry()
		// Pre-populate with commands
		for i := 0; i < 100; i++ {
			cmd, err := registry.NewCommand(fmt.Sprintf("cmd%d", i)).
				WithDescription("Test command").
				WithFunc(func() error { return nil }).
				WithCategory("Benchmark").
				Build()
			if err != nil {
				b.Fatalf("Failed to build command: %v", err)
			}
			r.MustRegister(cmd)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.List()
		}
	})

	b.Run("Search", func(b *testing.B) {
		r := registry.NewRegistry()
		// Pre-populate with commands
		for i := 0; i < 100; i++ {
			cmd, err := registry.NewCommand(fmt.Sprintf("cmd%d", i)).
				WithDescription("Test command for searching").
				WithFunc(func() error { return nil }).
				WithCategory("Benchmark").
				Build()
			if err != nil {
				b.Fatalf("Failed to build command: %v", err)
			}
			r.MustRegister(cmd)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.Search("test")
		}
	})

	b.Run("Execute", func(b *testing.B) {
		r := registry.NewRegistry()
		cmd, err := registry.NewCommand("benchmark").
			WithDescription("Benchmark command").
			WithFunc(func() error { return nil }).
			WithCategory("Benchmark").
			Build()
		if err != nil {
			b.Fatalf("Failed to build command: %v", err)
		}
		r.MustRegister(cmd)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := r.Execute("benchmark"); err != nil {
				b.Fatalf("Failed to execute command: %v", err)
			}
		}
	})

	b.Run("ExecuteWithArgs", func(b *testing.B) {
		r := registry.NewRegistry()
		cmd, err := registry.NewCommand("benchmarkargs").
			WithDescription("Benchmark command with args").
			WithArgsFunc(func(args ...string) error { return nil }).
			WithCategory("Benchmark").
			Build()
		if err != nil {
			b.Fatalf("Failed to build command: %v", err)
		}
		r.MustRegister(cmd)

		args := []string{"arg1", "arg2", "arg3"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := r.Execute("benchmarkargs", args...); err != nil {
				b.Fatalf("Failed to execute command with args: %v", err)
			}
		}
	})
}

// BenchmarkCommandOperations benchmarks command-specific operations
func BenchmarkCommandOperations(b *testing.B) {
	b.Run("CommandBuild", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cmd, err := registry.NewCommand(fmt.Sprintf("cmd%d", i)).
				WithDescription("Benchmark command").
				WithFunc(func() error { return nil }).
				WithCategory("Benchmark").
				Build()
			if err != nil {
				b.Fatalf("Failed to build command: %v", err)
			}
			_ = cmd // Use the command variable
		}
	})

	b.Run("CommandValidate", func(b *testing.B) {
		cmd, err := registry.NewCommand("validate").
			WithDescription("Validation benchmark").
			WithFunc(func() error { return nil }).
			Build()
		if err != nil {
			b.Fatalf("Failed to build command: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := cmd.Validate(); err != nil {
				b.Fatalf("Command validation failed: %v", err)
			}
		}
	})

	b.Run("CommandFullName", func(b *testing.B) {
		cmd, err := registry.NewNamespaceCommand("build", "linux").
			WithDescription("Full name benchmark").
			WithFunc(func() error { return nil }).
			Build()
		if err != nil {
			b.Fatalf("Failed to build namespace command: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			cmd.FullName()
		}
	})

	b.Run("CommandExecute", func(b *testing.B) {
		cmd, err := registry.NewCommand("execute").
			WithDescription("Execute benchmark").
			WithFunc(func() error { return nil }).
			Build()
		if err != nil {
			b.Fatalf("Failed to build execute command: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := cmd.Execute(); err != nil {
				b.Fatalf("Failed to execute command: %v", err)
			}
		}
	})

	b.Run("CommandExecuteWithArgs", func(b *testing.B) {
		cmd, err := registry.NewCommand("executeargs").
			WithDescription("Execute with args benchmark").
			WithArgsFunc(func(args ...string) error { return nil }).
			Build()
		if err != nil {
			b.Fatalf("Failed to build executeargs command: %v", err)
		}

		args := []string{"arg1", "arg2"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := cmd.Execute(args...); err != nil {
				b.Fatalf("Failed to execute command with args: %v", err)
			}
		}
	})
}

// BenchmarkMagexBinaryPerformance benchmarks actual magex binary performance
func BenchmarkMagexBinaryPerformance(b *testing.B) {
	b.Run("Startup", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			cmd := exec.CommandContext(ctx, "./magex", "-version")
			_, err := cmd.Output()
			if err != nil {
				b.Fatalf("magex -version failed: %v", err)
			}
		}
	})

	b.Run("ListCommands", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			cmd := exec.CommandContext(ctx, "./magex", "-l")
			_, err := cmd.Output()
			if err != nil {
				b.Fatalf("magex -l failed: %v", err)
			}
		}
	})

	b.Run("SearchCommands", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			cmd := exec.CommandContext(ctx, "./magex", "-search", "build")
			_, err := cmd.Output()
			if err != nil {
				b.Fatalf("magex -search failed: %v", err)
			}
		}
	})

	b.Run("Help", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			cmd := exec.CommandContext(ctx, "./magex", "-h")
			_, err := cmd.Output()
			if err != nil {
				b.Fatalf("magex -h failed: %v", err)
			}
		}
	})

	b.Run("NamespaceList", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ctx := context.Background()
			cmd := exec.CommandContext(ctx, "./magex", "-n")
			_, err := cmd.Output()
			if err != nil {
				b.Fatalf("magex -n failed: %v", err)
			}
		}
	})
}

// BenchmarkMemoryUsage benchmarks memory usage patterns
func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("RegistryMemory", func(b *testing.B) {
		var m1, m2 runtime.MemStats
		runtime.GC()
		runtime.ReadMemStats(&m1)

		// Create registry with many commands
		r := registry.NewRegistry()
		for i := 0; i < 1000; i++ {
			cmd, err := registry.NewCommand(fmt.Sprintf("cmd%d", i)).
				WithDescription(fmt.Sprintf("Command %d description", i)).
				WithFunc(func() error { return nil }).
				WithCategory("Memory").
				Build()
			if err != nil {
				b.Fatalf("Failed to build command: %v", err)
			}
			r.MustRegister(cmd)
		}

		runtime.GC()
		runtime.ReadMemStats(&m2)

		b.ReportMetric(float64(m2.Alloc-m1.Alloc), "bytes/registry")
		b.ReportMetric(float64(m2.Alloc-m1.Alloc)/1000, "bytes/command")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.List()
		}
	})

	b.Run("CommandMemory", func(b *testing.B) {
		var m1, m2 runtime.MemStats

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			runtime.GC()
			runtime.ReadMemStats(&m1)

			// Create and execute command
			cmd, err := registry.NewCommand("memtest").
				WithDescription("Memory test command").
				WithFunc(func() error { return nil }).
				WithCategory("Memory").
				Build()
			if err != nil {
				b.Fatalf("Failed to build memtest command: %v", err)
			}
			if err := cmd.Execute(); err != nil {
				b.Fatalf("Failed to execute memtest command: %v", err)
			}

			runtime.ReadMemStats(&m2)
			if i == 0 {
				b.ReportMetric(float64(m2.Alloc-m1.Alloc), "bytes/command-exec")
			}
		}
	})
}

// BenchmarkScalability tests how performance scales with number of commands
func BenchmarkScalability(b *testing.B) {
	sizes := []int{10, 100, 500, 1000, 5000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Registry%d", size), func(b *testing.B) {
			// Setup registry with 'size' commands
			r := registry.NewRegistry()
			for i := 0; i < size; i++ {
				cmd, err := registry.NewCommand(fmt.Sprintf("cmd%d", i)).
					WithDescription(fmt.Sprintf("Command %d", i)).
					WithFunc(func() error { return nil }).
					WithCategory(fmt.Sprintf("Cat%d", i%10)).
					Build()
				if err != nil {
					b.Fatalf("Failed to build command: %v", err)
				}
				r.MustRegister(cmd)
			}

			// Benchmark operations that should scale
			b.Run("Get", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					r.Get(fmt.Sprintf("cmd%d", i%size))
				}
			})

			b.Run("List", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					r.List()
				}
			})

			b.Run("Search", func(b *testing.B) {
				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					r.Search("cmd")
				}
			})
		})
	}
}

// BenchmarkConcurrency tests concurrent access performance
func BenchmarkConcurrency(b *testing.B) {
	r := registry.NewRegistry()

	// Pre-populate registry
	for i := 0; i < 100; i++ {
		cmd, err := registry.NewCommand(fmt.Sprintf("concurrent%d", i)).
			WithDescription("Concurrent test command").
			WithFunc(func() error { return nil }).
			WithCategory("Concurrency").
			Build()
		if err != nil {
			b.Fatalf("Failed to build concurrent command: %v", err)
		}
		r.MustRegister(cmd)
	}

	b.Run("ConcurrentGet", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				r.Get(fmt.Sprintf("concurrent%d", i%100))
				i++
			}
		})
	})

	b.Run("ConcurrentList", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				r.List()
			}
		})
	})

	b.Run("ConcurrentSearch", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				r.Search("concurrent")
			}
		})
	})

	b.Run("ConcurrentExecute", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				if err := r.Execute(fmt.Sprintf("concurrent%d", i%100)); err != nil {
					b.Fatalf("Failed to execute concurrent command: %v", err)
				}
				i++
			}
		})
	})
}

// BenchmarkStringOperations benchmarks string-heavy operations
func BenchmarkStringOperations(b *testing.B) {
	// Test string operations that are common in command processing
	commands := make([]*registry.Command, 100)
	for i := 0; i < 100; i++ {
		cmd, err := registry.NewNamespaceCommand("namespace", fmt.Sprintf("method%d", i)).
			WithDescription(fmt.Sprintf("This is a longer description for command %d that includes more text", i)).
			WithFunc(func() error { return nil }).
			WithCategory("StringOps").
			WithAliases(fmt.Sprintf("alias%d", i), fmt.Sprintf("short%d", i)).
			Build()
		if err != nil {
			b.Fatalf("Failed to build namespace command: %v", err)
		}
		commands[i] = cmd
	}

	b.Run("FullName", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			commands[i%100].FullName()
		}
	})

	b.Run("StringMatching", func(b *testing.B) {
		r := registry.NewRegistry()
		for _, cmd := range commands {
			r.MustRegister(cmd)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			r.Search("method")
		}
	})
}

// BenchmarkRealWorldScenarios benchmarks realistic usage patterns
func BenchmarkRealWorldScenarios(b *testing.B) {
	b.Run("TypicalWorkflow", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate typical magex usage: startup, list commands, execute
			r := registry.NewRegistry()

			// Register common commands (simulating embedded commands)
			commonCommands := []string{
				"build", "test", "lint", "clean", "deps", "release",
				"format", "vet", "mod", "generate", "docs", "help",
			}

			for _, name := range commonCommands {
				cmd, err := registry.NewCommand(name).
					WithDescription(fmt.Sprintf("%s command", name)).
					WithFunc(func() error { return nil }).
					WithCategory("Common").
					Build()
				if err != nil {
					b.Fatalf("Failed to build common command: %v", err)
				}
				r.MustRegister(cmd)
			}

			// Typical operations
			r.List()       // User lists commands
			r.Get("build") // User looks up specific command
			if err := r.Execute("help"); err != nil {
				// Help command execution error is non-critical in benchmark
				_ = err
			}
		}
	})

	b.Run("HeavyUsage", func(b *testing.B) {
		// Simulate heavy usage with many commands and frequent operations
		r := registry.NewRegistry()

		// Register many commands (like real MAGE-X)
		for i := 0; i < 200; i++ {
			cmd, err := registry.NewCommand(fmt.Sprintf("cmd%d", i)).
				WithDescription(fmt.Sprintf("Command %d for heavy usage testing", i)).
				WithFunc(func() error {
					// Simulate some work
					time.Sleep(1 * time.Microsecond)
					return nil
				}).
				WithCategory(fmt.Sprintf("Category%d", i%20)).
				Build()
			if err != nil {
				b.Fatalf("Failed to build command: %v", err)
			}
			r.MustRegister(cmd)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate heavy usage pattern
			r.List()        // List all commands
			r.Search("cmd") // Search commands
			if err := r.Execute(fmt.Sprintf("cmd%d", i%200)); err != nil {
				// Command execution error is non-critical in benchmark
				_ = err
			}
		}
	})
}

// BenchmarkComparison provides baseline comparisons
func BenchmarkComparison(b *testing.B) {
	b.Run("DirectFunctionCall", func(b *testing.B) {
		fn := func() error { return nil }
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := fn(); err != nil {
				// Function call error is non-critical in benchmark
				_ = err
			}
		}
	})

	b.Run("RegistryFunctionCall", func(b *testing.B) {
		r := registry.NewRegistry()
		cmd, err := registry.NewCommand("direct").
			WithDescription("Direct comparison").
			WithFunc(func() error { return nil }).
			Build()
		if err != nil {
			b.Fatalf("Failed to build direct command: %v", err)
		}
		r.MustRegister(cmd)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := r.Execute("direct"); err != nil {
				// Direct execution error is non-critical in benchmark
				_ = err
			}
		}
	})

	b.Run("MapLookup", func(b *testing.B) {
		// Baseline: direct map lookup
		m := make(map[string]func() error)
		for i := 0; i < 100; i++ {
			m[fmt.Sprintf("func%d", i)] = func() error { return nil }
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if fn, ok := m[fmt.Sprintf("func%d", i%100)]; ok {
				if err := fn(); err != nil {
					// Function call error is non-critical in benchmark
					_ = err
				}
			}
		}
	})

	b.Run("RegistryLookup", func(b *testing.B) {
		// Registry lookup and execution
		r := registry.NewRegistry()
		for i := 0; i < 100; i++ {
			cmd, err := registry.NewCommand(fmt.Sprintf("func%d", i)).
				WithFunc(func() error { return nil }).
				Build()
			if err != nil {
				b.Fatalf("Failed to build func command: %v", err)
			}
			r.MustRegister(cmd)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := r.Execute(fmt.Sprintf("func%d", i%100)); err != nil {
				// Function execution error is non-critical in benchmark
				_ = err
			}
		}
	})
}

// Performance test helpers
func reportPerformanceMetrics(b *testing.B, _ string, duration time.Duration) {
	b.ReportMetric(float64(duration.Nanoseconds())/float64(b.N), "ns/op")
	opsPerSecond := float64(b.N) / duration.Seconds()
	b.ReportMetric(opsPerSecond, "ops/sec")
}

// BenchmarkWithMetrics demonstrates the use of reportPerformanceMetrics
func BenchmarkWithMetrics(b *testing.B) {
	start := time.Now()
	for i := 0; i < b.N; i++ {
		// Simple operation for demonstration
		_ = fmt.Sprintf("test-%d", i)
	}
	duration := time.Since(start)
	reportPerformanceMetrics(b, "string_formatting", duration)
}

// TestPerformanceRegression runs performance regression tests
func TestPerformanceRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance regression tests in short mode")
	}

	// Test that basic operations complete within reasonable time
	tests := []struct {
		name        string
		maxDuration time.Duration
		operation   func() error
	}{
		{
			name:        "RegistryCreation",
			maxDuration: 1 * time.Millisecond,
			operation: func() error {
				registry.NewRegistry()
				return nil
			},
		},
		{
			name:        "CommandRegistration",
			maxDuration: 100 * time.Microsecond,
			operation: func() error {
				r := registry.NewRegistry()
				cmd, err := registry.NewCommand("test").
					WithFunc(func() error { return nil }).
					Build()
				if err != nil {
					return err
				}
				return r.Register(cmd)
			},
		},
		{
			name:        "CommandExecution",
			maxDuration: 50 * time.Microsecond,
			operation: func() error {
				r := registry.NewRegistry()
				cmd, err := registry.NewCommand("test").
					WithFunc(func() error { return nil }).
					Build()
				if err != nil {
					return err
				}
				r.MustRegister(cmd)
				return r.Execute("test")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			err := tt.operation()
			duration := time.Since(start)

			if err != nil {
				t.Fatalf("Operation failed: %v", err)
			}

			if duration > tt.maxDuration {
				t.Errorf("Operation %s took %v, expected less than %v",
					tt.name, duration, tt.maxDuration)
			}

			t.Logf("Operation %s completed in %v", tt.name, duration)
		})
	}
}
