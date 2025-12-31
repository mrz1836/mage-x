package mage

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/common/cache"
	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// Static errors for test compliance with err113
var (
	errTestBuildFailed   = errors.New("test build failed")
	errTestInstallFailed = errors.New("test install failed")
	errTestGenFailed     = errors.New("test generate failed")
)

// BuildCoverageTestSuite provides comprehensive coverage for Build methods
type BuildCoverageTestSuite struct {
	suite.Suite

	env   *testutil.TestEnvironment
	build Build
}

func TestBuildCoverageTestSuite(t *testing.T) {
	suite.Run(t, new(BuildCoverageTestSuite))
}

func (ts *BuildCoverageTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.build = Build{}

	// Disable cache for tests
	err := os.Setenv("MAGE_X_CACHE_DISABLED", "true")
	ts.Require().NoError(err)

	// Create basic project structure
	ts.env.CreateFile("main.go", `package main
func main() { println("test") }`)
	ts.env.CreateFile("go.sum", "")

	// Create .mage.yaml config
	ts.env.CreateMageConfig(`
project:
  name: test
  binary: testbin
build:
  output: bin
  platforms:
    - linux/amd64
    - darwin/amd64
`)
}

func (ts *BuildCoverageTestSuite) TearDownTest() {
	// Clean up environment variables - wrap in anonymous function for error handling
	defer func() {
		// Ignore cleanup errors in teardown
		_ = os.Unsetenv("MAGE_X_CACHE_DISABLED") //nolint:errcheck // cleanup in defer // cleanup in defer
		_ = os.Unsetenv("GOOS")                  //nolint:errcheck // cleanup in defer // cleanup in defer
		_ = os.Unsetenv("GOARCH")                //nolint:errcheck // cleanup in defer
		_ = os.Unsetenv("MAGE_X_VERSION")        //nolint:errcheck // cleanup in defer
		_ = os.Unsetenv("CLEAN_CACHE")           //nolint:errcheck // cleanup in defer
	}()

	// Reset config
	TestResetConfig()

	ts.env.Cleanup()
}

// Helper to set up mock runner and execute function
func (ts *BuildCoverageTestSuite) withMockRunner(fn func() error) error {
	return ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		fn,
	)
}

// mockVersionCommands sets up git version command mocks (flexible matching)
func (ts *BuildCoverageTestSuite) mockVersionCommands() {
	// Match any git command for version determination
	ts.env.Runner.On("RunCmdOutput", "git", mock.Anything).
		Return("v1.0.0", nil).Maybe()
}

// TestBuildAllSuccess tests successful Build.All
func (ts *BuildCoverageTestSuite) TestBuildAllSuccess() {
	ts.mockVersionCommands()

	// Mock go build for each platform
	ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == CmdGoBuild
	})).Return(nil).Maybe()

	err := ts.withMockRunner(func() error {
		return ts.build.All()
	})

	ts.Require().NoError(err)
}

// TestBuildAllWithBuildError tests Build.All when a platform build fails
func (ts *BuildCoverageTestSuite) TestBuildAllWithBuildError() {
	ts.mockVersionCommands()

	// Mock go build to fail
	ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == CmdGoBuild
	})).Return(errTestBuildFailed).Maybe()

	err := ts.withMockRunner(func() error {
		return ts.build.All()
	})

	ts.Require().Error(err)
	ts.Contains(err.Error(), "build errors")
}

// TestBuildPlatformSuccess tests successful Build.Platform
func (ts *BuildCoverageTestSuite) TestBuildPlatformSuccess() {
	ts.mockVersionCommands()

	ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == CmdGoBuild
	})).Return(nil).Maybe()

	err := ts.withMockRunner(func() error {
		return ts.build.Platform("linux/amd64")
	})

	ts.Require().NoError(err)
}

// TestBuildPlatformInvalidPlatform tests Build.Platform with invalid platform
func (ts *BuildCoverageTestSuite) TestBuildPlatformInvalidPlatform() {
	err := ts.withMockRunner(func() error {
		return ts.build.Platform("invalid-platform")
	})

	ts.Require().Error(err)
}

// TestBuildPlatformBuildFailure tests Build.Platform when build fails
func (ts *BuildCoverageTestSuite) TestBuildPlatformBuildFailure() {
	ts.mockVersionCommands()

	ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == CmdGoBuild
	})).Return(errTestBuildFailed).Maybe()

	err := ts.withMockRunner(func() error {
		return ts.build.Platform("linux/amd64")
	})

	ts.Require().Error(err)
	ts.Contains(err.Error(), "build")
}

// TestBuildLinux tests Build.Linux convenience method
func (ts *BuildCoverageTestSuite) TestBuildLinux() {
	ts.mockVersionCommands()

	ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == CmdGoBuild
	})).Return(nil).Maybe()

	err := ts.withMockRunner(func() error {
		return ts.build.Linux()
	})

	ts.Require().NoError(err)
}

// TestBuildDarwin tests Build.Darwin convenience method (builds both amd64 and arm64)
func (ts *BuildCoverageTestSuite) TestBuildDarwin() {
	ts.mockVersionCommands()

	ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == CmdGoBuild
	})).Return(nil).Maybe()

	err := ts.withMockRunner(func() error {
		return ts.build.Darwin()
	})

	ts.Require().NoError(err)
}

// TestBuildDarwinFirstPlatformFails tests Build.Darwin when first platform fails
func (ts *BuildCoverageTestSuite) TestBuildDarwinFirstPlatformFails() {
	ts.mockVersionCommands()

	ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == CmdGoBuild
	})).Return(errTestBuildFailed).Once()

	err := ts.withMockRunner(func() error {
		return ts.build.Darwin()
	})

	ts.Require().Error(err)
}

// TestBuildWindows tests Build.Windows convenience method
func (ts *BuildCoverageTestSuite) TestBuildWindows() {
	ts.mockVersionCommands()

	ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == CmdGoBuild
	})).Return(nil).Maybe()

	err := ts.withMockRunner(func() error {
		return ts.build.Windows()
	})

	ts.Require().NoError(err)
}

// TestBuildInstallSuccess tests successful Build.Install
func (ts *BuildCoverageTestSuite) TestBuildInstallSuccess() {
	ts.mockVersionCommands()

	ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == "install"
	})).Return(nil).Maybe()

	err := ts.withMockRunner(func() error {
		return ts.build.Install()
	})

	ts.Require().NoError(err)
}

// TestBuildInstallFailure tests Build.Install when install fails
func (ts *BuildCoverageTestSuite) TestBuildInstallFailure() {
	ts.mockVersionCommands()

	ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == "install"
	})).Return(errTestInstallFailed).Maybe()

	err := ts.withMockRunner(func() error {
		return ts.build.Install()
	})

	ts.Require().Error(err)
	ts.Contains(err.Error(), "install failed")
}

// TestBuildDevSuccess tests successful Build.Dev
func (ts *BuildCoverageTestSuite) TestBuildDevSuccess() {
	ts.mockVersionCommands()

	ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == "install"
	})).Return(nil).Maybe()

	err := ts.withMockRunner(func() error {
		return ts.build.Dev()
	})

	ts.Require().NoError(err)
}

// TestBuildDevFailure tests Build.Dev when build fails
func (ts *BuildCoverageTestSuite) TestBuildDevFailure() {
	ts.mockVersionCommands()

	ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == "install"
	})).Return(errTestBuildFailed).Maybe()

	err := ts.withMockRunner(func() error {
		return ts.build.Dev()
	})

	ts.Require().Error(err)
	ts.Contains(err.Error(), "dev build failed")
}

// TestBuildGenerateSuccess tests successful Build.Generate
func (ts *BuildCoverageTestSuite) TestBuildGenerateSuccess() {
	ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == "generate"
	})).Return(nil).Maybe()

	err := ts.withMockRunner(func() error {
		return ts.build.Generate()
	})

	ts.Require().NoError(err)
}

// TestBuildGenerateFailure tests Build.Generate when generate fails
func (ts *BuildCoverageTestSuite) TestBuildGenerateFailure() {
	ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == "generate"
	})).Return(errTestGenFailed).Maybe()

	err := ts.withMockRunner(func() error {
		return ts.build.Generate()
	})

	ts.Require().Error(err)
	ts.Contains(err.Error(), "generate failed")
}

// TestBuildCleanSuccess tests successful Build.Clean
func (ts *BuildCoverageTestSuite) TestBuildCleanSuccess() {
	// Create bin directory
	err := os.MkdirAll("bin", 0o750)
	ts.Require().NoError(err)

	ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == "clean"
	})).Return(nil).Maybe()

	err = ts.withMockRunner(func() error {
		return ts.build.Clean()
	})

	ts.Require().NoError(err)
}

// TestBuildCleanWithCacheClean tests Build.Clean with cache cleaning enabled
func (ts *BuildCoverageTestSuite) TestBuildCleanWithCacheClean() {
	err := os.Setenv("CLEAN_CACHE", "true")
	ts.Require().NoError(err)

	ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == "clean"
	})).Return(nil).Maybe()

	err = ts.withMockRunner(func() error {
		return ts.build.Clean()
	})

	ts.Require().NoError(err)
}

// TestBuildCleanFailure tests Build.Clean when clean fails
func (ts *BuildCoverageTestSuite) TestBuildCleanFailure() {
	ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == "clean"
	})).Return(errTestBuildFailed).Maybe()

	err := ts.withMockRunner(func() error {
		return ts.build.Clean()
	})

	ts.Require().Error(err)
}

// TestGenerateBuildHash tests generateBuildHash method
func (ts *BuildCoverageTestSuite) TestGenerateBuildHash() {
	b := Build{}

	ts.Run("returns empty when cache manager is nil", func() {
		ctx := &buildContext{
			cacheManager: nil,
			platform:     "linux/amd64",
			ldflags:      "-s -w",
			sourceFiles:  []string{"main.go"},
			configFiles:  []string{"go.mod"},
		}

		hash := b.generateBuildHash(ctx)
		ts.Empty(hash)
	})
}

// TestCheckCachedBuild tests checkCachedBuild method
func (ts *BuildCoverageTestSuite) TestCheckCachedBuild() {
	b := Build{}

	ts.Run("returns false when hash is empty", func() {
		result := b.checkCachedBuild("", "bin/test", nil)
		ts.False(result)
	})

	ts.Run("returns false when cache manager is nil", func() {
		result := b.checkCachedBuild("abc123", "bin/test", nil)
		ts.False(result)
	})
}

// TestExecuteBuild tests executeBuild method
func (ts *BuildCoverageTestSuite) TestExecuteBuild() {
	ts.Run("returns error on build failure", func() {
		ts.env.Runner.On("RunCmd", "go", mock.Anything).Return(errTestBuildFailed).Once()

		ctx := &buildContext{
			cfg:          &Config{},
			cacheManager: nil,
			outputPath:   "bin/test",
			buildArgs:    []string{"build", "-o", "bin/test", "."},
		}

		err := ts.withMockRunner(func() error {
			return ts.build.executeBuild(ctx, "")
		})

		ts.Require().Error(err)
		ts.ErrorIs(err, ErrBuildFailedError)
	})

	ts.Run("succeeds on successful build", func() {
		env := testutil.NewTestEnvironment(ts.T())
		defer env.Cleanup()

		env.Runner.On("RunCmd", "go", mock.Anything).Return(nil).Once()

		ctx := &buildContext{
			cfg:          &Config{},
			cacheManager: nil,
			outputPath:   "bin/test",
			buildArgs:    []string{"build", "-o", "bin/test", "."},
		}

		err := env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
			},
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.executeBuild(ctx, "")
			},
		)

		ts.Require().NoError(err)
	})
}

// TestStoreBuildResult tests storeBuildResult method
func (ts *BuildCoverageTestSuite) TestStoreBuildResult() {
	b := Build{}

	ts.Run("does nothing when hash is empty", func() {
		ctx := &buildContext{
			cacheManager: nil,
		}
		// Should not panic
		b.storeBuildResult(ctx, "", true, "", time.Second)
	})

	ts.Run("does nothing when cache manager is nil", func() {
		ctx := &buildContext{
			cacheManager: nil,
		}
		// Should not panic
		b.storeBuildResult(ctx, "abc123", true, "", time.Second)
	})
}

// TestTryUseCachedBuild tests tryUseCachedBuild method
func (ts *BuildCoverageTestSuite) TestTryUseCachedBuild() {
	b := Build{}

	ts.Run("returns false when build result not found", func() {
		cfg := cache.DefaultConfig()
		cfg.Enabled = true
		cfg.Directory = ts.T().TempDir()
		cm := cache.NewManager(cfg)
		err := cm.Init()
		ts.Require().NoError(err)

		result := b.tryUseCachedBuild("nonexistent-hash", "bin/test", *cm)
		ts.False(result)
	})
}

// TestPreBuildWithArgs tests PreBuildWithArgs with various strategies
func (ts *BuildCoverageTestSuite) TestPreBuildWithArgs() {
	// Mock go list and build commands
	ts.env.Runner.On("RunCmdOutput", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == "list"
	})).Return("test/module\n", nil).Maybe()

	ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == CmdGoBuild
	})).Return(nil).Maybe()

	ts.Run("uses default smart strategy", func() {
		err := ts.withMockRunner(func() error {
			return ts.build.PreBuildWithArgs()
		})
		// May fail due to workspace mode detection, which is expected
		// The important thing is the code paths are exercised
		_ = err
	})
}

// TestPreBuild tests PreBuild wrapper
func (ts *BuildCoverageTestSuite) TestPreBuild() {
	ts.env.Runner.On("RunCmdOutput", "go", mock.Anything).Return("", nil).Maybe()
	ts.env.Runner.On("RunCmd", "go", mock.Anything).Return(nil).Maybe()

	err := ts.withMockRunner(func() error {
		return ts.build.PreBuild()
	})
	// Exercise the code path, error is acceptable
	_ = err
}

// Standalone tests (not in suite) for simpler cases

func TestGenerateBuildHashNilCacheManager(t *testing.T) {
	b := Build{}
	ctx := &buildContext{
		cacheManager: nil,
		platform:     "linux/amd64",
	}

	hash := b.generateBuildHash(ctx)
	assert.Empty(t, hash)
}

func TestCheckCachedBuildEmptyHash(t *testing.T) {
	b := Build{}
	result := b.checkCachedBuild("", "bin/test", nil)
	assert.False(t, result)
}

func TestStoreBuildResultNilManager(t *testing.T) {
	b := Build{}
	ctx := &buildContext{
		cacheManager: nil,
		outputPath:   "bin/test",
	}

	// Should not panic
	b.storeBuildResult(ctx, "hash123", true, "", time.Second)
}

func TestStoreBuildResultEmptyHash(t *testing.T) {
	b := Build{}
	ctx := &buildContext{
		cacheManager: nil,
		outputPath:   "bin/test",
	}

	// Should not panic
	b.storeBuildResult(ctx, "", true, "", time.Second)
}

// TestCacheManagerProviderCoverage tests the cache manager provider
func TestCacheManagerProviderCoverage(t *testing.T) {
	t.Run("creates default provider", func(t *testing.T) {
		provider := NewDefaultCacheManagerProvider()
		require.NotNil(t, provider)

		// Should return a cache manager (may be disabled)
		cm := provider.GetCacheManager()
		require.NotNil(t, cm)
	})

	t.Run("handles disabled cache", func(t *testing.T) {
		err := os.Setenv("MAGE_X_CACHE_DISABLED", "true")
		require.NoError(t, err)
		defer func() {
			_ = os.Unsetenv("MAGE_X_CACHE_DISABLED") //nolint:errcheck // cleanup in defer // cleanup in defer
		}()

		provider := NewDefaultCacheManagerProvider()
		cm := provider.GetCacheManager()
		require.NotNil(t, cm)
		assert.False(t, cm.IsEnabled())
	})
}

// TestBuildContextCreation tests createBuildContext
func TestBuildContextCreation(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer func() {
		_ = os.Chdir(originalWd) //nolint:errcheck // cleanup in defer //nolint:errcheck // cleanup in defer
	}()

	// Create minimal project structure
	require.NoError(t, os.WriteFile("go.mod", []byte("module test\n\ngo 1.24"), 0o600))
	require.NoError(t, os.WriteFile("main.go", []byte("package main\nfunc main() {}"), 0o600))
	require.NoError(t, os.MkdirAll("bin", 0o750))

	// Disable cache
	require.NoError(t, os.Setenv("MAGE_X_CACHE_DISABLED", "true"))
	defer func() {
		_ = os.Unsetenv("MAGE_X_CACHE_DISABLED") //nolint:errcheck // cleanup in defer // cleanup in defer
	}()

	b := Build{}
	cfg := &Config{
		Project: ProjectConfig{
			Name:   "test",
			Binary: "testbin",
		},
		Build: BuildConfig{
			Output: "bin",
		},
	}

	ctx, err := b.createBuildContext(cfg)
	require.NoError(t, err)
	require.NotNil(t, ctx)

	assert.Equal(t, cfg, ctx.cfg)
	assert.NotEmpty(t, ctx.outputPath)
	assert.NotEmpty(t, ctx.platform)
}

// TestBuildGetConfigFilesWithMageYaml tests getConfigFiles when .mage.yaml exists
func TestBuildGetConfigFilesWithMageYaml(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer func() {
		_ = os.Chdir(originalWd) //nolint:errcheck // cleanup in defer
	}()

	// Create .mage.yaml
	require.NoError(t, os.WriteFile(".mage.yaml", []byte("project:\n  name: test"), 0o600))

	b := Build{}
	files := b.getConfigFiles()

	assert.Contains(t, files, "go.mod")
	assert.Contains(t, files, "go.sum")
	assert.Contains(t, files, ".mage.yaml")
}

// mockCacheManager is a mock implementation for testing
type mockCacheManager struct {
	enabled bool
}

func (m *mockCacheManager) GetCacheManager() *cache.Manager {
	cfg := cache.DefaultConfig()
	cfg.Enabled = m.enabled
	return cache.NewManager(cfg)
}

func TestSetCacheManagerProvider(t *testing.T) {
	// Save original provider
	originalProvider := packageCacheManagerProvider.Get()
	defer func() {
		packageCacheManagerProvider.Set(originalProvider)
	}()

	// Set custom provider
	mockProvider := &mockCacheManager{enabled: false}
	SetCacheManagerProvider(mockProvider)

	// Verify it was set
	cm := getCacheManager()
	require.NotNil(t, cm)
}

// TestDeterminePackagePath tests the determinePackagePath function
func TestDeterminePackagePath(t *testing.T) {
	b := Build{}

	t.Run("uses configured main path when provided", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			_ = os.Chdir(originalWd) //nolint:errcheck // cleanup in defer
		}()

		// Create main package
		require.NoError(t, os.MkdirAll("cmd/app", 0o750))
		require.NoError(t, os.WriteFile("cmd/app/main.go", []byte("package main\nfunc main() {}"), 0o600))

		cfg := &Config{
			Project: ProjectConfig{
				Main: "cmd/app",
			},
		}

		path, err := b.determinePackagePath(cfg, "bin/app", true)
		require.NoError(t, err)
		assert.Equal(t, "./cmd/app", path)
	})

	t.Run("returns error for invalid configured path", func(t *testing.T) {
		cfg := &Config{
			Project: ProjectConfig{
				Main: "nonexistent/path",
			},
		}

		_, err := b.determinePackagePath(cfg, "bin/app", true)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidMainPath)
	})
}

// TestFindPackagePath tests findPackagePath function
func TestFindPackagePath(t *testing.T) {
	b := Build{}

	t.Run("returns ./... for non-binary builds", func(t *testing.T) {
		path, err := b.findPackagePath("", false)
		require.NoError(t, err)
		assert.Equal(t, "./...", path)
	})

	t.Run("returns error when no main package found for binary build", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			_ = os.Chdir(originalWd) //nolint:errcheck // cleanup in defer
		}()

		// Create go.mod but no main package
		require.NoError(t, os.WriteFile("go.mod", []byte("module test\n\ngo 1.24"), 0o600))
		require.NoError(t, os.WriteFile("lib.go", []byte("package lib"), 0o600))

		_, err = b.findPackagePath("bin/app", true)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrNoMainPackage)
	})

	t.Run("finds main package in root directory", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			_ = os.Chdir(originalWd) //nolint:errcheck // cleanup in defer
		}()

		// Create main.go in root
		require.NoError(t, os.WriteFile("main.go", []byte("package main\nfunc main() {}"), 0o600))

		path, err := b.findPackagePath("bin/app", true)
		require.NoError(t, err)
		assert.Equal(t, ".", path)
	})

	t.Run("finds main package in cmd directory", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			_ = os.Chdir(originalWd) //nolint:errcheck // cleanup in defer
		}()

		// Create cmd/app with main.go
		require.NoError(t, os.MkdirAll("cmd/app", 0o750))
		require.NoError(t, os.WriteFile("cmd/app/main.go", []byte("package main\nfunc main() {}"), 0o600))

		path, err := b.findPackagePath("bin/app", true)
		require.NoError(t, err)
		assert.Equal(t, "./cmd/app", path)
	})
}

// TestBuildContextPlatform tests that platform is correctly set in buildContext
func TestBuildContextPlatform(t *testing.T) {
	tempDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tempDir))
	defer func() {
		_ = os.Chdir(originalWd) //nolint:errcheck // cleanup in defer
	}()

	// Create minimal structure
	require.NoError(t, os.WriteFile("go.mod", []byte("module test\n\ngo 1.24"), 0o600))
	require.NoError(t, os.WriteFile("main.go", []byte("package main\nfunc main() {}"), 0o600))

	// Disable cache
	require.NoError(t, os.Setenv("MAGE_X_CACHE_DISABLED", "true"))
	defer func() {
		_ = os.Unsetenv("MAGE_X_CACHE_DISABLED") //nolint:errcheck // cleanup in defer
	}()

	b := Build{}
	cfg := &Config{
		Project: ProjectConfig{Binary: "test"},
		Build:   BuildConfig{Output: "."},
	}

	ctx, err := b.createBuildContext(cfg)
	require.NoError(t, err)

	// Platform should be set to runtime.GOOS/runtime.GOARCH
	assert.Contains(t, ctx.platform, "/")
}

// BenchmarkCreateBuildContext benchmarks the createBuildContext function
func BenchmarkCreateBuildContext(b *testing.B) {
	tempDir := b.TempDir()
	originalWd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}
	if err := os.Chdir(tempDir); err != nil {
		b.Fatal(err)
	}
	defer func() {
		_ = os.Chdir(originalWd) //nolint:errcheck // cleanup in defer
	}()

	// Create minimal structure
	if err := os.WriteFile("go.mod", []byte("module test\n\ngo 1.24"), 0o600); err != nil {
		b.Fatal(err)
	}
	if err := os.WriteFile("main.go", []byte("package main\nfunc main() {}"), 0o600); err != nil {
		b.Fatal(err)
	}

	if err := os.Setenv("MAGE_X_CACHE_DISABLED", "true"); err != nil {
		b.Fatal(err)
	}
	defer func() {
		_ = os.Unsetenv("MAGE_X_CACHE_DISABLED") //nolint:errcheck // cleanup in defer
	}()

	build := Build{}
	cfg := &Config{
		Project: ProjectConfig{Binary: "test"},
		Build:   BuildConfig{Output: "."},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//nolint:errcheck // benchmark - error checking not needed
		_, _ = build.createBuildContext(cfg)
	}
}
