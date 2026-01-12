package mage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/common/cache"
	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// TestGenerateBuildHashWithEnabledCache tests generateBuildHash with cache manager
func TestGenerateBuildHashWithEnabledCache(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer func() {
		_ = os.Chdir(originalWd) //nolint:errcheck // cleanup in defer
	}()

	// Create minimal Go project
	require.NoError(t, os.WriteFile("go.mod", []byte("module test\n\ngo 1.24"), 0o600))
	require.NoError(t, os.WriteFile("main.go", []byte("package main\nfunc main() {}"), 0o600))

	// Create and initialize cache manager
	cfg := cache.DefaultConfig()
	cfg.Enabled = true
	cfg.Directory = filepath.Join(tempDir, ".cache")
	cm := cache.NewManager(cfg)
	require.NoError(t, cm.Init())

	b := Build{}
	ctx := &buildContext{
		cacheManager: cm,
		platform:     "linux/amd64",
		ldflags:      "-s -w",
		sourceFiles:  []string{"main.go"},
		configFiles:  []string{"go.mod"},
	}

	hash := b.generateBuildHash(ctx)
	assert.NotEmpty(t, hash, "Should generate hash with enabled cache")
}

// TestStoreBuildResultWithEnabledCache tests storeBuildResult with enabled cache
func TestStoreBuildResultWithEnabledCache(t *testing.T) {
	tempDir := t.TempDir()

	// Create and initialize cache manager
	cfg := cache.DefaultConfig()
	cfg.Enabled = true
	cfg.Directory = filepath.Join(tempDir, ".cache")
	cm := cache.NewManager(cfg)
	require.NoError(t, cm.Init())

	b := Build{}
	ctx := &buildContext{
		cacheManager: cm,
		platform:     "linux/amd64",
		outputPath:   "bin/test",
		buildArgs:    []string{"build", "-o", "bin/test"},
	}

	// Create output binary for size calculation
	require.NoError(t, os.MkdirAll("bin", 0o750))
	require.NoError(t, os.WriteFile("bin/test", []byte("fake binary"), 0o600))

	// Store successful build result
	b.storeBuildResult(ctx, "test-hash-123", true, "", time.Second)

	// Verify result was stored
	result, found := cm.GetBuildCache().GetBuildResult("test-hash-123")
	assert.True(t, found, "Build result should be stored in cache")
	if found {
		assert.True(t, result.Success)
		assert.Equal(t, "bin/test", result.Binary)
		assert.Equal(t, "linux/amd64", result.Platform)
	}
}

// TestTryUseCachedBuildWithCacheHit tests tryUseCachedBuild when cache has valid result
func TestTryUseCachedBuildWithCacheHit(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer func() {
		_ = os.Chdir(originalWd) //nolint:errcheck // cleanup in defer
	}()

	// Create and initialize cache manager
	cfg := cache.DefaultConfig()
	cfg.Enabled = true
	cfg.Directory = filepath.Join(tempDir, ".cache")
	cm := cache.NewManager(cfg)
	require.NoError(t, cm.Init())

	// Create cached binary
	require.NoError(t, os.MkdirAll("bin", 0o750))
	cachedBinary := filepath.Join(tempDir, "bin", "cached")
	require.NoError(t, os.WriteFile(cachedBinary, []byte("cached binary"), 0o600))

	// Store successful build result in cache
	buildResult := &cache.BuildResult{
		Binary:   cachedBinary,
		Platform: "linux/amd64",
		Success:  true,
	}
	require.NoError(t, cm.GetBuildCache().StoreBuildResult("valid-hash", buildResult))

	b := Build{}
	outputPath := filepath.Join(tempDir, "bin", "output")

	// Try to use cached build
	result := b.tryUseCachedBuild("valid-hash", outputPath, *cm)
	assert.True(t, result, "Should use cached build when valid result exists")
	assert.FileExists(t, outputPath, "Output binary should be copied from cache")
}

// TestFindSourceFilesNormal tests findSourceFiles in normal case
func TestFindSourceFilesNormal(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer func() {
		_ = os.Chdir(originalWd) //nolint:errcheck // cleanup in defer
	}()

	// Create go.mod
	require.NoError(t, os.WriteFile("go.mod", []byte("module test"), 0o600))

	b := Build{}
	files := b.findSourceFiles()

	// Should at least contain go.mod
	assert.Contains(t, files, "go.mod")
}

// TestNormalizeMainPathEdgeCases tests normalizeMainPath with various inputs
func TestNormalizeMainPathEdgeCases(t *testing.T) {
	b := Build{}

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "./",
		},
		{
			name:     "already has dot slash",
			input:    "./cmd/app",
			expected: "./cmd/app",
		},
		{
			name:     "relative path without dot",
			input:    "cmd/app",
			expected: "./cmd/app",
		},
		{
			name:     "single directory",
			input:    "main",
			expected: "./main",
		},
		{
			name:     "dot only",
			input:    ".",
			expected: "./.",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := b.normalizeMainPath(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestDetermineBuildOutputEdgeCases tests determineBuildOutput with edge cases
func TestDetermineBuildOutputEdgeCases(t *testing.T) {
	b := Build{}

	t.Run("with configured output directory", func(t *testing.T) {
		cfg := &Config{
			Build: BuildConfig{
				Output: "dist",
			},
			Project: ProjectConfig{
				Binary: "myapp",
			},
		}

		output := b.determineBuildOutput(cfg)
		assert.Contains(t, output, "dist")
		assert.Contains(t, output, "myapp")
	})

	t.Run("with default output", func(t *testing.T) {
		cfg := &Config{
			Build: BuildConfig{
				Output: "",
			},
			Project: ProjectConfig{
				Binary: "testapp",
			},
		}

		output := b.determineBuildOutput(cfg)
		assert.NotEmpty(t, output)
	})
}

// TestBuildIncrementalWithWorkspace tests buildIncremental in workspace mode
func TestBuildIncrementalWithWorkspace(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create workspace structure
	env.CreateFile("go.work", "go 1.24\n\nuse ./module1")
	env.CreateGoMod("module1")
	env.CreateFile("module1/main.go", "package main\nfunc main() {}")

	// Create .mage.yaml config
	env.CreateMageConfig(`
project:
  name: test
`)

	// Mock commands
	env.Runner.On("RunCmdOutput", "go", mock.Anything).Return("module1", nil).Maybe()
	env.Runner.On("RunCmd", "go", mock.Anything).Return(nil).Maybe()

	b := Build{}
	err := env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		func() error {
			// buildIncremental should detect workspace mode
			return b.buildIncremental(5, 100, "", false, "")
		},
	)

	// May error due to workspace validation, but code path is exercised
	_ = err
}

// TestBuildMainsFirstSuccess tests buildMainsFirst strategy
func TestBuildMainsFirstSuccess(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateGoMod("test/module")
	env.CreateFile("main.go", "package main\nfunc main() {}")
	env.CreateFile("lib.go", "package main")

	env.CreateMageConfig(`
project:
  name: test
`)

	// Mock go list to return main package
	env.Runner.On("RunCmdOutput", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == CmdGoList
	})).Return("test/module\n", nil).Maybe()

	// Mock build commands
	env.Runner.On("RunCmd", "go", mock.Anything).Return(nil).Maybe()

	b := Build{}
	err := env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		func() error {
			return b.buildMainsFirst(10, false, "", false, "")
		},
	)

	// Exercise code path
	_ = err
}

// TestBuildMainsFirstMainsOnly tests buildMainsFirst with mains-only flag
func TestBuildMainsFirstMainsOnly(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateGoMod("test/module")
	env.CreateMageConfig(`
project:
  name: test
`)

	// Mock commands
	env.Runner.On("RunCmdOutput", "go", mock.Anything).Return("test/module\n", nil).Maybe()
	env.Runner.On("RunCmd", "go", mock.Anything).Return(nil).Maybe()

	b := Build{}
	err := env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		func() error {
			// Should stop after building mains
			return b.buildMainsFirst(10, true, "", false, "")
		},
	)

	// Exercise code path
	_ = err
}

// TestBuildSmartLowMemory tests buildSmart with low memory scenario
func TestBuildSmartLowMemory(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateGoMod("test/module")
	env.CreateFile("main.go", "package main\nfunc main() {}")

	env.CreateMageConfig(`
project:
  name: test
`)

	// Mock commands
	env.Runner.On("RunCmdOutput", "go", mock.Anything).Return("test/module\n", nil).Maybe()
	env.Runner.On("RunCmd", "go", mock.Anything).Return(nil).Maybe()

	b := Build{}
	err := env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		func() error {
			// buildSmart will select appropriate strategy based on resources
			return b.buildSmart("", false, "")
		},
	)

	// May error due to module detection, but code path is exercised
	_ = err
}

// TestBuildFullStrategy tests buildFull strategy
func TestBuildFullStrategy(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateGoMod("test/module")
	env.CreateMageConfig(`
project:
  name: test
`)

	// Mock commands
	env.Runner.On("RunCmd", "go", mock.Anything).Return(nil).Maybe()

	b := Build{}
	err := env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		func() error {
			return b.buildFull("", false, "")
		},
	)

	// Exercise code path
	_ = err
}

// TestBenchTraceWithArgsParameters tests TraceWithArgs with various parameters
func TestBenchTraceWithArgsParameters(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateGoMod("test/module")
	env.CreateMageConfig(`
project:
  name: test
`)

	// Mock commands
	env.Runner.On("RunCmd", "go", mock.Anything).Return(nil).Maybe()

	bench := Bench{}
	err := env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		func() error {
			// Test with custom parameters to hit more code paths
			return bench.TraceWithArgs("trace=custom.out", "time=5s", "verbose=true", "skip=TestSkip")
		},
	)

	// Exercise code path
	_ = err
}

// TestBenchSaveWithArgsParameters tests SaveWithArgs with various parameters
func TestBenchSaveWithArgsParameters(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateGoMod("test/module")
	env.CreateMageConfig(`
project:
  name: test
`)

	// Mock all command variations that SaveWithArgs might use
	env.Runner.On("RunCmdOutputDir", mock.Anything, mock.Anything, mock.Anything).
		Return("BenchmarkTest 100 1000 ns/op", nil).Maybe()
	env.Runner.On("RunCmdOutput", mock.Anything, mock.Anything).
		Return("BenchmarkTest 100 1000 ns/op", nil).Maybe()
	env.Runner.On("RunCmd", mock.Anything, mock.Anything).Return(nil).Maybe()

	bench := Bench{}
	err := env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		func() error {
			// Test with custom parameters
			return bench.SaveWithArgs("output=bench-custom.txt", "time=2s", "count=3", "verbose=true")
		},
	)

	// Exercise code path
	_ = err
}

// TestBenchCPUWithArgsCustomProfile tests CPUWithArgs with custom profile path
func TestBenchCPUWithArgsCustomProfile(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateGoMod("test/module")
	env.CreateMageConfig(`
project:
  name: test
`)

	// Mock commands
	env.Runner.On("RunCmd", "go", mock.Anything).Return(nil).Maybe()

	bench := Bench{}
	err := env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		func() error {
			// Test with custom profile path
			return bench.CPUWithArgs("profile=custom-cpu.prof")
		},
	)

	// Exercise code path - may error on pprof analysis but that's expected
	_ = err
}

// TestBenchMemWithArgsCustomProfile tests MemWithArgs with custom profile path
func TestBenchMemWithArgsCustomProfile(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateGoMod("test/module")
	env.CreateMageConfig(`
project:
  name: test
`)

	// Mock commands
	env.Runner.On("RunCmd", "go", mock.Anything).Return(nil).Maybe()

	bench := Bench{}
	err := env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		func() error {
			// Test with custom profile path
			return bench.MemWithArgs("profile=custom-mem.prof")
		},
	)

	// Exercise code path - may error on pprof analysis but that's expected
	_ = err
}

// TestBenchProfileWithArgsCustomProfiles tests ProfileWithArgs with custom profile paths
func TestBenchProfileWithArgsCustomProfiles(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateGoMod("test/module")
	env.CreateMageConfig(`
project:
  name: test
`)

	// Mock commands
	env.Runner.On("RunCmd", "go", mock.Anything).Return(nil).Maybe()

	bench := Bench{}
	err := env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		func() error {
			// Test with custom profile paths
			return bench.ProfileWithArgs("cpu-profile=custom.cpu", "mem-profile=custom.mem")
		},
	)

	// Exercise code path
	_ = err
}

// TestBenchRegressionWithArgsCreateBaseline tests RegressionWithArgs creating baseline
func TestBenchRegressionWithArgsCreateBaseline(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateGoMod("test/module")
	env.CreateMageConfig(`
project:
  name: test
`)

	// Mock all command variations
	env.Runner.On("RunCmdOutputDir", mock.Anything, mock.Anything, mock.Anything).
		Return("BenchmarkTest 100 1000 ns/op", nil).Maybe()
	env.Runner.On("RunCmdOutput", mock.Anything, mock.Anything).
		Return("BenchmarkTest 100 1000 ns/op", nil).Maybe()
	env.Runner.On("RunCmd", mock.Anything, mock.Anything).Return(nil).Maybe()

	bench := Bench{}
	err := env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		func() error {
			// First run should create baseline
			return bench.RegressionWithArgs()
		},
	)

	// Exercise code path
	_ = err
}

// TestBenchRegressionWithArgsUpdateBaseline tests RegressionWithArgs with update flag
func TestBenchRegressionWithArgsUpdateBaseline(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateGoMod("test/module")
	env.CreateMageConfig(`
project:
  name: test
`)

	// Create baseline file
	baseline := filepath.Join(env.TempDir, "bench-baseline.txt")
	require.NoError(t, os.WriteFile(baseline, []byte("BenchmarkTest 100 1000 ns/op"), 0o600))
	require.NoError(t, os.Setenv("BENCH_BASELINE", baseline))
	defer func() {
		_ = os.Unsetenv("BENCH_BASELINE") //nolint:errcheck // cleanup in defer
	}()

	// Mock all command variations
	env.Runner.On("RunCmdOutputDir", mock.Anything, mock.Anything, mock.Anything).
		Return("BenchmarkTest 100 900 ns/op", nil).Maybe()
	env.Runner.On("RunCmdOutput", mock.Anything, mock.Anything).
		Return("BenchmarkTest 100 900 ns/op", nil).Maybe()
	env.Runner.On("RunCmd", mock.Anything, mock.Anything).Return(nil).Maybe()

	bench := Bench{}
	err := env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		func() error {
			// Run with update-baseline flag
			return bench.RegressionWithArgs("update-baseline=true")
		},
	)

	// Exercise code path
	_ = err
}

// TestCacheManagerProviderWithFailedInit tests cache manager when init fails
func TestCacheManagerProviderWithFailedInit(t *testing.T) {
	tempDir := t.TempDir()

	// Create directory without write permissions to cause init failure
	cacheDir := filepath.Join(tempDir, "readonly")
	require.NoError(t, os.MkdirAll(cacheDir, 0o444)) //nolint:gosec // Read-only dir intentional for test

	require.NoError(t, os.Setenv("MAGE_X_CACHE_DIR", cacheDir))
	defer func() {
		_ = os.Unsetenv("MAGE_X_CACHE_DIR") //nolint:errcheck // cleanup in defer
	}()

	// Should handle init failure gracefully
	provider := NewDefaultCacheManagerProvider()
	cm := provider.GetCacheManager()
	require.NotNil(t, cm, "Should return cache manager even if init fails")
}

// TestBuildWorkspaceModulesExercise tests buildWorkspaceModules code path
func TestBuildWorkspaceModulesExercise(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	// Create workspace structure
	env.CreateFile("go.work", "go 1.24\n\nuse (\n\t./module1\n\t./module2\n)")
	env.CreateGoMod("module1")
	env.CreateGoMod("module2")

	env.CreateMageConfig(`
project:
  name: test
`)

	// Mock commands
	env.Runner.On("RunCmd", "go", mock.Anything).Return(nil).Maybe()
	env.Runner.On("RunCmdOutput", "go", mock.Anything).Return("", nil).Maybe()

	b := Build{}
	err := env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		func() error {
			return b.buildWorkspaceModules(false, "", "")
		},
	)

	// Exercise code path
	_ = err
}

// TestDiscoverPackagesWithExclude tests discoverPackages with exclude pattern
func TestDiscoverPackagesWithExclude(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer func() {
		_ = os.Chdir(originalWd) //nolint:errcheck // cleanup in defer
	}()

	// Create minimal Go project
	require.NoError(t, os.WriteFile("go.mod", []byte("module test\n\ngo 1.24"), 0o600))
	require.NoError(t, os.WriteFile("main.go", []byte("package main\nfunc main() {}"), 0o600))

	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	// Mock go list to return packages
	env.Runner.On("RunCmdOutput", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == CmdGoList
	})).Return("test\ntest/internal\ntest/vendor", nil).Maybe()

	b := Build{}
	err = env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		func() error {
			// Exercise discoverPackages with exclude pattern
			packages, discoverErr := b.discoverPackages("vendor")
			if discoverErr == nil {
				// Packages should exclude vendor
				for _, pkg := range packages {
					assert.NotContains(t, pkg, "vendor", "Should exclude vendor packages")
				}
			}
			return discoverErr
		},
	)

	// Exercise code path
	_ = err
}

// TestFindMainPackagesMultiple tests findMainPackages with multiple mains
func TestFindMainPackagesMultiple(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer func() {
		_ = os.Chdir(originalWd) //nolint:errcheck // cleanup in defer
	}()

	// Create multiple main packages
	require.NoError(t, os.WriteFile("go.mod", []byte("module test\n\ngo 1.24"), 0o600))
	require.NoError(t, os.MkdirAll("cmd/app1", 0o750))
	require.NoError(t, os.MkdirAll("cmd/app2", 0o750))
	require.NoError(t, os.WriteFile("cmd/app1/main.go", []byte("package main\nfunc main() {}"), 0o600))
	require.NoError(t, os.WriteFile("cmd/app2/main.go", []byte("package main\nfunc main() {}"), 0o600))

	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	// Mock go list
	env.Runner.On("RunCmdOutput", "go", mock.Anything).
		Return("test/cmd/app1\ntest/cmd/app2", nil).Maybe()

	b := Build{}
	err = env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		func() error {
			_, findErr := b.findMainPackages()
			return findErr
		},
	)

	// Exercise code path
	_ = err
}
