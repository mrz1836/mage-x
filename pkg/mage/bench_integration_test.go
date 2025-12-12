package mage_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// setupBenchTest creates a temp directory with a go.mod file and changes to it.
// It returns a cleanup function that should be deferred.
func setupBenchTest(t *testing.T) (string, func()) {
	tmpDir := t.TempDir()

	// Create go.mod
	err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module example.com/test"), 0o644)
	require.NoError(t, err)

	// Save original directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)

	// Change to temp dir
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	return tmpDir, func() {
		os.Chdir(originalDir)
	}
}

func TestBench_DefaultWithArgs(t *testing.T) {
	_, cleanup := setupBenchTest(t)
	defer cleanup()

	// Mock runner
	runner, builder := testutil.NewMockRunner()

	// Save original runner
	originalRunner := mage.GetRunner()
	defer mage.SetRunner(originalRunner)

	mage.SetRunner(runner)

	// Expectation: go test -bench=. -benchmem -run=^$ -benchtime 3s ./...
	builder.ExpectGoCommand("test", nil)

	b := mage.Bench{}
	err := b.DefaultWithArgs() // Uses default args

	assert.NoError(t, err)
	runner.AssertExpectations(t)
}

func TestBench_DefaultWithArgs_Custom(t *testing.T) {
	_, cleanup := setupBenchTest(t)
	defer cleanup()

	// Mock runner
	runner, builder := testutil.NewMockRunner()

	originalRunner := mage.GetRunner()
	defer mage.SetRunner(originalRunner)

	mage.SetRunner(runner)

	// We expect arguments to contain custom values
	// time=1s -> -benchtime 1s
	// count=2 -> -count 2
	// verbose=true -> -v
	builder.ExpectGoCommand("test", nil)

	b := mage.Bench{}
	err := b.DefaultWithArgs("time=1s", "count=2", "verbose=true")

	assert.NoError(t, err)
	runner.AssertExpectations(t)
}

func TestBench_SaveWithArgs(t *testing.T) {
	tmpDir, cleanup := setupBenchTest(t)
	defer cleanup()

	runner, _ := testutil.NewMockRunner()

	originalRunner := mage.GetRunner()
	defer mage.SetRunner(originalRunner)

	mage.SetRunner(runner)

	// Expect go test execution, returning some output
	runner.On("RunCmdOutput", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == "test"
	})).Return("BenchmarkResult\t100 ns/op", nil)

	outputFile := filepath.Join(tmpDir, "results.txt")
	b := mage.Bench{}
	err := b.SaveWithArgs("output="+outputFile)

	assert.NoError(t, err)
	runner.AssertExpectations(t)

	// Verify file content
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Contains(t, string(content), "BenchmarkResult")
}

func TestBench_CompareWithArgs(t *testing.T) {
	tmpDir, cleanup := setupBenchTest(t)
	defer cleanup()

	// Create dummy benchmark files
	oldFile := filepath.Join(tmpDir, "old.txt")
	newFile := filepath.Join(tmpDir, "new.txt")
	os.WriteFile(oldFile, []byte("old"), 0o644)
	os.WriteFile(newFile, []byte("new"), 0o644)

	runner, _ := testutil.NewMockRunner()

	originalRunner := mage.GetRunner()
	defer mage.SetRunner(originalRunner)

	mage.SetRunner(runner)

	// Expect benchstat installation if needed
	// The code checks if benchstat exists using utils.CommandExists.
	// Since we can't easily mock utils.CommandExists which uses exec.LookPath,
	// and likely it returns false in this sandbox or environment, we should expect the install command.
	// Or we can expect "RunCmd" with "go" "install" ... optionally.
	// But since the mock is strict, we should probably add it.

	// Note: We use .Maybe() because on some systems where benchstat IS installed, this won't be called.
	runner.On("RunCmd", "go", []string{"install", "golang.org/x/perf/cmd/benchstat@latest"}).Return(nil).Maybe()

	// Expect benchstat
	runner.On("RunCmd", "benchstat", []string{oldFile, newFile}).Return(nil)

	b := mage.Bench{}
	err := b.CompareWithArgs("old="+oldFile, "new="+newFile)

	assert.NoError(t, err)
	runner.AssertExpectations(t)
}

func TestBench_CPUWithArgs(t *testing.T) {
	_, cleanup := setupBenchTest(t)
	defer cleanup()

	runner, builder := testutil.NewMockRunner()

	originalRunner := mage.GetRunner()
	defer mage.SetRunner(originalRunner)

	mage.SetRunner(runner)

	// Expect go test
	builder.ExpectGoCommand("test", nil)

	// Expect pprof
	runner.On("RunCmd", "go", []string{"tool", "pprof", "-top", "cpu.prof"}).Return(nil)

	b := mage.Bench{}
	// Run
	err := b.CPUWithArgs()

	assert.NoError(t, err)
	runner.AssertExpectations(t)

	// Check env var
	assert.Equal(t, "cpu.prof", os.Getenv("MAGE_X_BENCH_CPU_PROFILE"))
	os.Unsetenv("MAGE_X_BENCH_CPU_PROFILE")
}

func TestBench_MemWithArgs(t *testing.T) {
	_, cleanup := setupBenchTest(t)
	defer cleanup()

	runner, builder := testutil.NewMockRunner()

	originalRunner := mage.GetRunner()
	defer mage.SetRunner(originalRunner)

	mage.SetRunner(runner)

	// Expect go test
	builder.ExpectGoCommand("test", nil)

	// Expect pprof
	runner.On("RunCmd", "go", []string{"tool", "pprof", "-top", "-alloc_space", "mem.prof"}).Return(nil)

	b := mage.Bench{}
	// Run
	err := b.MemWithArgs()

	assert.NoError(t, err)
	runner.AssertExpectations(t)

	// Check env var
	assert.Equal(t, "mem.prof", os.Getenv("MAGE_X_BENCH_MEM_PROFILE"))
	os.Unsetenv("MAGE_X_BENCH_MEM_PROFILE")
}

func TestBench_TraceWithArgs(t *testing.T) {
	_, cleanup := setupBenchTest(t)
	defer cleanup()

	runner, _ := testutil.NewMockRunner()

	originalRunner := mage.GetRunner()
	defer mage.SetRunner(originalRunner)

	mage.SetRunner(runner)

	runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		if len(args) == 0 || args[0] != "test" { return false }
		for i, arg := range args {
			if arg == "-trace" {
				// Verify the trace file argument follows
				if i+1 < len(args) {
					// We just verify it's there
					return true
				}
			}
		}
		return false
	})).Return(nil)

	b := mage.Bench{}
	err := b.TraceWithArgs("trace=trace.out")

	assert.NoError(t, err)
	runner.AssertExpectations(t)
}
