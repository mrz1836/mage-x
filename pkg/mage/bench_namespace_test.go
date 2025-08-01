package mage

import (
	"os"
	"testing"

	"github.com/mrz1836/go-mage/pkg/mage/testutil"
	"github.com/stretchr/testify/suite"
)

// BenchTestSuite defines the test suite for Bench namespace methods
type BenchTestSuite struct {
	suite.Suite

	env   *testutil.TestEnvironment
	bench Bench
}

// SetupTest runs before each test
func (ts *BenchTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.bench = Bench{}
}

// TearDownTest runs after each test
func (ts *BenchTestSuite) TearDownTest() {
	// Clean up environment variables that might be set by tests
	if err := os.Unsetenv("BENCH_TIME"); err != nil {
		ts.T().Logf("Failed to unset BENCH_TIME: %v", err)
	}
	if err := os.Unsetenv("BENCH_COUNT"); err != nil {
		ts.T().Logf("Failed to unset BENCH_COUNT: %v", err)
	}
	if err := os.Unsetenv("BENCH_CPU_PROFILE"); err != nil {
		ts.T().Logf("Failed to unset BENCH_CPU_PROFILE: %v", err)
	}
	if err := os.Unsetenv("BENCH_MEM_PROFILE"); err != nil {
		ts.T().Logf("Failed to unset BENCH_MEM_PROFILE: %v", err)
	}
	if err := os.Unsetenv("BENCH_FILE"); err != nil {
		ts.T().Logf("Failed to unset BENCH_FILE: %v", err)
	}
	if err := os.Unsetenv("BENCH_OLD"); err != nil {
		ts.T().Logf("Failed to unset BENCH_OLD: %v", err)
	}
	if err := os.Unsetenv("BENCH_NEW"); err != nil {
		ts.T().Logf("Failed to unset BENCH_NEW: %v", err)
	}
	if err := os.Unsetenv("CPU_PROFILE"); err != nil {
		ts.T().Logf("Failed to unset CPU_PROFILE: %v", err)
	}
	if err := os.Unsetenv("MEM_PROFILE"); err != nil {
		ts.T().Logf("Failed to unset MEM_PROFILE: %v", err)
	}
	if err := os.Unsetenv("TRACE_FILE"); err != nil {
		ts.T().Logf("Failed to unset TRACE_FILE: %v", err)
	}
	if err := os.Unsetenv("UPDATE_BASELINE"); err != nil {
		ts.T().Logf("Failed to unset UPDATE_BASELINE: %v", err)
	}
	if err := os.Unsetenv("BENCH_BASELINE"); err != nil {
		ts.T().Logf("Failed to unset BENCH_BASELINE: %v", err)
	}

	ts.env.Cleanup()
}

// TestBenchDefault tests the Default method
func (ts *BenchTestSuite) TestBenchDefault() {
	ts.Run("successful benchmark execution", func() {
		// Mock successful go test benchmark command
		ts.env.Runner.On("RunCmd", "go", []string{"test", "-bench=.", "-benchmem", "-run=^$", "-benchtime", "10s", "./..."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.Default()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("benchmark with custom time", func() {
		// Set environment variable for custom bench time
		originalBenchTime := os.Getenv("BENCH_TIME")
		defer func() {
			if err := os.Setenv("BENCH_TIME", originalBenchTime); err != nil {
				ts.T().Logf("Failed to restore BENCH_TIME: %v", err)
			}
		}()
		if err := os.Setenv("BENCH_TIME", "5s"); err != nil {
			ts.T().Fatalf("Failed to set BENCH_TIME: %v", err)
		}

		// Mock successful go test benchmark command with custom time
		ts.env.Runner.On("RunCmd", "go", []string{"test", "-bench=.", "-benchmem", "-run=^$", "-benchtime", "5s", "./..."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.Default()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("benchmark with count", func() {
		// Set environment variable for count
		originalBenchCount := os.Getenv("BENCH_COUNT")
		defer func() {
			if err := os.Setenv("BENCH_COUNT", originalBenchCount); err != nil {
				ts.T().Logf("Failed to restore BENCH_COUNT: %v", err)
			}
		}()
		if err := os.Setenv("BENCH_COUNT", "3"); err != nil {
			ts.T().Fatalf("Failed to set BENCH_COUNT: %v", err)
		}

		// Mock successful go test benchmark command with count
		ts.env.Runner.On("RunCmd", "go", []string{"test", "-bench=.", "-benchmem", "-run=^$", "-benchtime", "10s", "-count", "3", "./..."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.Default()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("benchmark with CPU and memory profiling", func() {
		// Set environment variables for profiling
		originalCPUProfile := os.Getenv("BENCH_CPU_PROFILE")
		originalMemProfile := os.Getenv("BENCH_MEM_PROFILE")
		defer func() {
			if err := os.Setenv("BENCH_CPU_PROFILE", originalCPUProfile); err != nil {
				ts.T().Logf("Failed to restore BENCH_CPU_PROFILE: %v", err)
			}
			if err := os.Setenv("BENCH_MEM_PROFILE", originalMemProfile); err != nil {
				ts.T().Logf("Failed to restore BENCH_MEM_PROFILE: %v", err)
			}
		}()

		if err := os.Setenv("BENCH_CPU_PROFILE", "cpu.prof"); err != nil {
			ts.T().Fatalf("Failed to set BENCH_CPU_PROFILE: %v", err)
		}
		if err := os.Setenv("BENCH_MEM_PROFILE", "mem.prof"); err != nil {
			ts.T().Fatalf("Failed to set BENCH_MEM_PROFILE: %v", err)
		}

		// Mock successful go test benchmark command with profiling
		ts.env.Runner.On("RunCmd", "go", []string{"test", "-bench=.", "-benchmem", "-run=^$", "-benchtime", "10s", "-cpuprofile", "cpu.prof", "-memprofile", "mem.prof", "./..."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.Default()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestBenchCompare tests the Compare method
func (ts *BenchTestSuite) TestBenchCompare() {
	ts.Run("successful benchmark comparison", func() {
		// Create test benchmark files
		ts.env.CreateFile("old.txt", "BenchmarkTest 1000 1000000 ns/op")
		ts.env.CreateFile("new.txt", "BenchmarkTest 1200 800000 ns/op")

		// Mock benchstat installation (if needed) and comparison command
		ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/perf/cmd/benchstat@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "benchstat", []string{"old.txt", "new.txt"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.Compare()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("benchmark comparison with custom files", func() {
		// Set environment variables for custom files
		originalOld := os.Getenv("BENCH_OLD")
		originalNew := os.Getenv("BENCH_NEW")
		defer func() {
			if err := os.Setenv("BENCH_OLD", originalOld); err != nil {
				ts.T().Logf("Failed to restore BENCH_OLD: %v", err)
			}
			if err := os.Setenv("BENCH_NEW", originalNew); err != nil {
				ts.T().Logf("Failed to restore BENCH_NEW: %v", err)
			}
		}()

		if err := os.Setenv("BENCH_OLD", "baseline.txt"); err != nil {
			ts.T().Fatalf("Failed to set BENCH_OLD: %v", err)
		}
		if err := os.Setenv("BENCH_NEW", "current.txt"); err != nil {
			ts.T().Fatalf("Failed to set BENCH_NEW: %v", err)
		}

		// Create test benchmark files
		ts.env.CreateFile("baseline.txt", "BenchmarkTest 1000 1000000 ns/op")
		ts.env.CreateFile("current.txt", "BenchmarkTest 1200 800000 ns/op")

		// Mock benchstat installation (if needed) and comparison command
		ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/perf/cmd/benchstat@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "benchstat", []string{"baseline.txt", "current.txt"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.Compare()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("missing old benchmark file", func() {
		// Ensure old.txt doesn't exist
		if err := os.Remove("old.txt"); err != nil && !os.IsNotExist(err) {
			ts.Require().NoError(err)
		}
		if err := os.Remove("new.txt"); err != nil && !os.IsNotExist(err) {
			ts.Require().NoError(err)
		}

		// Mock benchstat installation but don't create old.txt file
		ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/perf/cmd/benchstat@latest"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.Compare()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "old benchmark file not found")
	})

	ts.Run("missing new benchmark file", func() {
		// Ensure new.txt doesn't exist, but old.txt does
		if err := os.Remove("new.txt"); err != nil && !os.IsNotExist(err) {
			ts.Require().NoError(err)
		}
		ts.env.CreateFile("old.txt", "BenchmarkTest 1000 1000000 ns/op")

		// Mock benchstat installation but don't create new.txt file
		ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/perf/cmd/benchstat@latest"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.Compare()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "new benchmark file not found")
	})
}

// TestBenchSave tests the Save method
func (ts *BenchTestSuite) TestBenchSave() {
	ts.Run("successful benchmark save with default filename", func() {
		// Mock successful go test benchmark command and output
		ts.env.Runner.On("RunCmdOutput", "go", []string{"test", "-bench=.", "-benchmem", "-run=^$", "-benchtime", "10s", "./..."}).Return("BenchmarkTest 1000 1000000 ns/op", nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.Save()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("benchmark save with custom filename", func() {
		// Set environment variable for custom output file
		originalBenchFile := os.Getenv("BENCH_FILE")
		defer func() {
			if err := os.Setenv("BENCH_FILE", originalBenchFile); err != nil {
				ts.T().Logf("Failed to restore BENCH_FILE: %v", err)
			}
		}()
		if err := os.Setenv("BENCH_FILE", "custom-bench.txt"); err != nil {
			ts.T().Fatalf("Failed to set BENCH_FILE: %v", err)
		}

		// Mock successful go test benchmark command and output
		ts.env.Runner.On("RunCmdOutput", "go", []string{"test", "-bench=.", "-benchmem", "-run=^$", "-benchtime", "10s", "./..."}).Return("BenchmarkTest 1000 1000000 ns/op", nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.Save()
			},
		)

		ts.Require().NoError(err)
		ts.Require().True(ts.env.FileExists("custom-bench.txt"))
	})

	ts.Run("benchmark save with directory creation", func() {
		// Set environment variable for output file in subdirectory
		originalBenchFile := os.Getenv("BENCH_FILE")
		defer func() {
			if err := os.Setenv("BENCH_FILE", originalBenchFile); err != nil {
				ts.T().Logf("Failed to restore BENCH_FILE: %v", err)
			}
		}()
		if err := os.Setenv("BENCH_FILE", "benchmarks/results.txt"); err != nil {
			ts.T().Fatalf("Failed to set BENCH_FILE: %v", err)
		}

		// Mock successful go test benchmark command and output
		ts.env.Runner.On("RunCmdOutput", "go", []string{"test", "-bench=.", "-benchmem", "-run=^$", "-benchtime", "10s", "./..."}).Return("BenchmarkTest 1000 1000000 ns/op", nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.Save()
			},
		)

		ts.Require().NoError(err)
		ts.Require().True(ts.env.FileExists("benchmarks"))
		ts.Require().True(ts.env.FileExists("benchmarks/results.txt"))
	})
}

// TestBenchCPU tests the CPU method
func (ts *BenchTestSuite) TestBenchCPU() {
	ts.Run("successful CPU profiling", func() {
		// Mock successful go test benchmark command with CPU profiling
		ts.env.Runner.On("RunCmd", "go", []string{"test", "-bench=.", "-benchmem", "-run=^$", "-benchtime", "10s", "-cpuprofile", "cpu.prof", "./..."}).Return(nil)
		// Mock CPU profile analysis
		ts.env.Runner.On("RunCmd", "go", []string{"tool", "pprof", "-top", "cpu.prof"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.CPU()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("CPU profiling with custom profile name", func() {
		// Set environment variable for custom CPU profile name
		originalCPUProfile := os.Getenv("CPU_PROFILE")
		defer func() {
			if err := os.Setenv("CPU_PROFILE", originalCPUProfile); err != nil {
				ts.T().Logf("Failed to restore CPU_PROFILE: %v", err)
			}
		}()
		if err := os.Setenv("CPU_PROFILE", "custom-cpu.prof"); err != nil {
			ts.T().Fatalf("Failed to set CPU_PROFILE: %v", err)
		}

		// Mock successful go test benchmark command with custom CPU profile
		ts.env.Runner.On("RunCmd", "go", []string{"test", "-bench=.", "-benchmem", "-run=^$", "-benchtime", "10s", "-cpuprofile", "custom-cpu.prof", "./..."}).Return(nil)
		// Mock CPU profile analysis
		ts.env.Runner.On("RunCmd", "go", []string{"tool", "pprof", "-top", "custom-cpu.prof"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.CPU()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestBenchMem tests the Mem method
func (ts *BenchTestSuite) TestBenchMem() {
	ts.Run("successful memory profiling", func() {
		// Mock successful go test benchmark command with memory profiling
		ts.env.Runner.On("RunCmd", "go", []string{"test", "-bench=.", "-benchmem", "-run=^$", "-benchtime", "10s", "-memprofile", "mem.prof", "./..."}).Return(nil)
		// Mock memory profile analysis
		ts.env.Runner.On("RunCmd", "go", []string{"tool", "pprof", "-top", "-alloc_space", "mem.prof"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.Mem()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("memory profiling with custom profile name", func() {
		// Set environment variable for custom memory profile name
		originalMemProfile := os.Getenv("MEM_PROFILE")
		defer func() {
			if err := os.Setenv("MEM_PROFILE", originalMemProfile); err != nil {
				ts.T().Logf("Failed to restore MEM_PROFILE: %v", err)
			}
		}()
		if err := os.Setenv("MEM_PROFILE", "custom-mem.prof"); err != nil {
			ts.T().Fatalf("Failed to set MEM_PROFILE: %v", err)
		}

		// Mock successful go test benchmark command with custom memory profile
		ts.env.Runner.On("RunCmd", "go", []string{"test", "-bench=.", "-benchmem", "-run=^$", "-benchtime", "10s", "-memprofile", "custom-mem.prof", "./..."}).Return(nil)
		// Mock memory profile analysis
		ts.env.Runner.On("RunCmd", "go", []string{"tool", "pprof", "-top", "-alloc_space", "custom-mem.prof"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.Mem()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestBenchProfile tests the Profile method
func (ts *BenchTestSuite) TestBenchProfile() {
	ts.Run("successful full profiling", func() {
		// Mock successful go test benchmark command with both CPU and memory profiling
		ts.env.Runner.On("RunCmd", "go", []string{"test", "-bench=.", "-benchmem", "-run=^$", "-benchtime", "10s", "-cpuprofile", "cpu.prof", "-memprofile", "mem.prof", "./..."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.Profile()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestBenchTrace tests the Trace method
func (ts *BenchTestSuite) TestBenchTrace() {
	ts.Run("successful execution tracing", func() {
		// Mock successful go test benchmark command with tracing
		ts.env.Runner.On("RunCmd", "go", []string{"test", "-bench=.", "-benchmem", "-run=^$", "-trace", "trace.out", "-benchtime", "10s", "./..."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.Trace()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("tracing with custom trace file", func() {
		// Set environment variable for custom trace file
		originalTraceFile := os.Getenv("TRACE_FILE")
		defer func() {
			if err := os.Setenv("TRACE_FILE", originalTraceFile); err != nil {
				ts.T().Logf("Failed to restore TRACE_FILE: %v", err)
			}
		}()
		if err := os.Setenv("TRACE_FILE", "custom-trace.out"); err != nil {
			ts.T().Fatalf("Failed to set TRACE_FILE: %v", err)
		}

		// Mock successful go test benchmark command with custom trace file
		ts.env.Runner.On("RunCmd", "go", []string{"test", "-bench=.", "-benchmem", "-run=^$", "-trace", "custom-trace.out", "-benchtime", "10s", "./..."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.Trace()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestBenchRegression tests the Regression method
func (ts *BenchTestSuite) TestBenchRegression() {
	ts.Run("regression check with new baseline creation", func() {
		// Mock successful benchmark save
		ts.env.Runner.On("RunCmdOutput", "go", []string{"test", "-bench=.", "-benchmem", "-run=^$", "-benchtime", "10s", "./..."}).Return("BenchmarkTest 1000 1000000 ns/op", nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.Regression()
			},
		)

		ts.Require().NoError(err)
		ts.Require().True(ts.env.FileExists("bench-baseline.txt"))
	})

	ts.Run("regression check with existing baseline", func() {
		// Create existing baseline
		ts.env.CreateFile("bench-baseline.txt", "BenchmarkTest 1000 1000000 ns/op")

		// Mock successful benchmark save and comparison
		ts.env.Runner.On("RunCmdOutput", "go", []string{"test", "-bench=.", "-benchmem", "-run=^$", "-benchtime", "10s", "./..."}).Return("BenchmarkTest 1200 800000 ns/op", nil)
		ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/perf/cmd/benchstat@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "benchstat", []string{"bench-baseline.txt", "bench-current.txt"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.Regression()
			},
		)

		ts.Require().NoError(err)
		ts.Require().True(ts.env.FileExists("bench-current.txt"))
	})

	ts.Run("regression check with baseline update", func() {
		// Set environment variable to update baseline
		originalUpdateBaseline := os.Getenv("UPDATE_BASELINE")
		defer func() {
			if err := os.Setenv("UPDATE_BASELINE", originalUpdateBaseline); err != nil {
				ts.T().Logf("Failed to restore UPDATE_BASELINE: %v", err)
			}
		}()
		if err := os.Setenv("UPDATE_BASELINE", "true"); err != nil {
			ts.T().Fatalf("Failed to set UPDATE_BASELINE: %v", err)
		}

		// Create existing baseline
		ts.env.CreateFile("bench-baseline.txt", "BenchmarkTest 1000 1000000 ns/op")

		// Mock successful benchmark save and comparison
		ts.env.Runner.On("RunCmdOutput", "go", []string{"test", "-bench=.", "-benchmem", "-run=^$", "-benchtime", "10s", "./..."}).Return("BenchmarkTest 1200 800000 ns/op", nil)
		ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/perf/cmd/benchstat@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "benchstat", []string{"bench-baseline.txt", "bench-current.txt"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.bench.Regression()
			},
		)

		ts.Require().NoError(err)
		// baseline should be updated with new content
		ts.Require().True(ts.env.FileExists("bench-baseline.txt"))
	})
}

// TestBenchTestSuite runs the test suite
func TestBenchTestSuite(t *testing.T) {
	suite.Run(t, new(BenchTestSuite))
}
