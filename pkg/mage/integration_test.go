//go:build integration
// +build integration

package mage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/testhelpers"
)

// IntegrationTestSuite provides common infrastructure for integration tests
type IntegrationTestSuite struct {
	testhelpers.BaseSuite
}

// SetupSuite configures the integration test environment
func (its *IntegrationTestSuite) SetupSuite() {
	// Configure BaseSuite options for integration tests
	its.Options = testhelpers.BaseSuiteOptions{
		CreateTempDir:   true,
		ChangeToTempDir: true,
		CreateGoModule:  false, // We'll create our own test projects
		PreserveEnv:     true,
		DisableCache:    true,
		SetupGitRepo:    false,
	}

	// Call parent setup
	its.BaseSuite.SetupSuite()
}

// createTestProject creates a standardized test project with go.mod and main.go
func (its *IntegrationTestSuite) createTestProject(name, module string) {
	its.T().Helper()

	projectDir := filepath.Join(its.TmpDir, name)
	its.RequireNoError(os.MkdirAll(projectDir, 0o750))
	its.RequireNoError(os.Chdir(projectDir))

	// Create go.mod
	goMod := `module ` + module + `

go 1.24
`
	its.RequireNoError(os.WriteFile("go.mod", []byte(goMod), 0o600))
}

// setupGoProject creates a complete Go project with main.go
func (its *IntegrationTestSuite) setupGoProject(mainContent string) {
	its.T().Helper()

	if mainContent == "" {
		mainContent = `package main

import "fmt"

func main() {
	fmt.Println("Hello, Integration Test!")
}
`
	}

	its.RequireNoError(os.WriteFile("main.go", []byte(mainContent), 0o600))
}

// setupMageConfig creates and sets a mage configuration for testing
func (its *IntegrationTestSuite) setupMageConfig(name, binary, module string) {
	its.T().Helper()

	config := &Config{
		Project: ProjectConfig{
			Name:   name,
			Binary: binary,
			Module: module,
		},
		Build: BuildConfig{
			Output:   "bin",
			TrimPath: true,
		},
	}

	TestSetConfig(config)
}

// TestBuildIntegration tests the full build pipeline
func (its *IntegrationTestSuite) TestBuildIntegration() {
	if testing.Short() {
		its.T().Skip("Skipping integration test in short mode")
	}

	// Create test project
	its.createTestProject("test-project", "test-project")
	its.setupGoProject("")
	its.setupMageConfig("test-project", "test-app", "test-project")

	// Test build
	build := Build{}
	err := build.Default()
	its.AssertNoError(err)

	// Verify binary exists
	binaryPath := filepath.Join("bin", "test-app")
	if _, statErr := os.Stat(binaryPath); os.IsNotExist(statErr) {
		its.T().Errorf("Expected binary at %s, but it doesn't exist", binaryPath)
	}

	// Test clean
	err = build.Clean()
	its.AssertNoError(err)

	// Verify binary is removed
	if _, err := os.Stat(binaryPath); !os.IsNotExist(err) {
		its.T().Errorf("Expected binary to be removed after clean")
	}
}

// TestSecureCommandExecution tests secure command execution
func TestSecureCommandExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	runner := NewSecureCommandRunner()

	tests := []struct {
		name      string
		cmd       string
		args      []string
		shouldErr bool
	}{
		{
			name:      "valid command",
			cmd:       "echo",
			args:      []string{"hello"},
			shouldErr: false,
		},
		{
			name:      "command injection attempt",
			cmd:       "echo",
			args:      []string{"hello; rm -rf /"},
			shouldErr: true,
		},
		{
			name:      "path traversal attempt",
			cmd:       "cat",
			args:      []string{"../../../etc/passwd"},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runner.RunCmd(tt.cmd, tt.args...)
			if tt.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestMultipleNamespaceIntegration tests multiple namespaces working together
func (its *IntegrationTestSuite) TestMultipleNamespaceIntegration() {
	if testing.Short() {
		its.T().Skip("Skipping integration test in short mode")
	}

	// Create test project with tests
	its.createTestProject("test", "test")

	// Create main.go with testable function
	mainGo := `package main

func Add(a, b int) int {
	return a + b
}

func main() {}
`
	its.RequireNoError(os.WriteFile("main.go", []byte(mainGo), 0o600))

	// Create test file
	testGo := `package main

import "testing"

func TestAdd(t *testing.T) {
	if Add(1, 2) != 3 {
		t.Error("1 + 2 should equal 3")
	}
}
`
	its.RequireNoError(os.WriteFile("main_test.go", []byte(testGo), 0o600))

	// Setup config
	TestSetConfig(&Config{
		Project: ProjectConfig{
			Name:   "test",
			Module: "test",
		},
		Build: BuildConfig{
			Output: "bin",
		},
		Test: TestConfig{
			Parallel: 1,
			Cover:    true,
		},
		Lint: LintConfig{
			Timeout: "1m",
		},
	})

	// Test workflow: Format -> Test -> Build
	its.Run("format", func() {
		format := Format{}
		// Skip if gofumpt not installed
		if _, err := GetRunner().RunCmdOutput("which", "gofumpt"); err != nil {
			its.T().Skip("gofumpt not installed")
		}
		err := format.Default()
		its.Require().NoError(err)
	})

	its.Run("test", func() {
		test := Test{}
		err := test.Unit()
		its.Require().NoError(err)
	})

	its.Run("build", func() {
		build := Build{}
		err := build.Default()
		its.Require().NoError(err)
	})
}

// TestConcurrentOperations tests concurrent namespace operations
func TestConcurrentOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Run multiple operations concurrently
	errChan := make(chan error, 3)

	go func() {
		metrics := Metrics{}
		errChan <- metrics.LOC()
	}()

	go func() {
		git := Git{}
		errChan <- git.Status()
	}()

	go func() {
		mod := Mod{}
		errChan <- mod.Tidy()
	}()

	// Wait for all operations
	for i := 0; i < 3; i++ {
		select {
		case err := <-errChan:
			// Some operations might fail if tools not installed
			if err != nil {
				t.Logf("Operation failed (might be expected): %v", err)
			}
		case <-ctx.Done():
			t.Fatal("Timeout waiting for operations")
		}
	}
}

// TestConfigurationLoading tests configuration loading and validation
func (its *IntegrationTestSuite) TestConfigurationLoading() {
	if testing.Short() {
		its.T().Skip("Skipping integration test in short mode")
	}

	// Test default config
	its.Run("default config", func() {
		// Create a completely isolated directory with no go.mod
		isolatedDir := filepath.Join(its.TmpDir, "isolated")
		its.RequireNoError(os.MkdirAll(isolatedDir, 0o750))
		its.RequireNoError(os.Chdir(isolatedDir))

		// Ensure ambient MAGE_X_* env overrides do not contaminate the
		// default-value assertions below (applyEnvOverrides reads these).
		restoreEnv := unsetEnvVars(
			its.T(),
			"MAGE_X_AUTO_DISCOVER_BUILD_TAGS",
			"MAGE_X_AUTO_DISCOVER_BUILD_TAGS_COMBINE",
			// The mage-x test runner exports MAGE_X_AUTO_DISCOVER_BUILD_TAGS_EXCLUDE
			// to its `go test` subprocess (so `magex test:race` runs these tests with
			// it set, e.g. "mage,windows,..."). It overrides both the default and the
			// YAML fixture, so clear it to assert the configured exclude list.
			"MAGE_X_AUTO_DISCOVER_BUILD_TAGS_EXCLUDE",
			// Binary-name overrides win over both defaults and the YAML fixture, so a
			// leaked value (e.g. config_test.go sets MAGE_X_CUSTOM_BINARY_NAME) would
			// override the asserted Project.Binary. Clear them for a clean baseline.
			"MAGE_X_BINARY_NAME",
			"MAGE_X_CUSTOM_BINARY_NAME",
		)
		defer restoreEnv()

		TestResetConfig() // Reset
		config, err := GetConfig()
		its.Require().NoError(err)
		its.Require().NotNil(config)
		its.Equal("app", config.Project.Binary)

		// Test-section defaults from defaultConfig() (config.go).
		// CombineBuildTags defaults to true; auto-discover defaults to off.
		its.True(config.Test.CombineBuildTags,
			"CombineBuildTags should default to true")
		its.False(config.Test.AutoDiscoverBuildTags,
			"AutoDiscoverBuildTags should default to false")
		its.Empty(config.Test.AutoDiscoverBuildTagsExclude,
			"AutoDiscoverBuildTagsExclude should default to empty")
	})

	// Test custom config
	its.Run("custom config", func() {
		// Guard against ambient overrides for the merge-preservation assertions.
		restoreEnv := unsetEnvVars(
			its.T(),
			"MAGE_X_AUTO_DISCOVER_BUILD_TAGS",
			"MAGE_X_AUTO_DISCOVER_BUILD_TAGS_COMBINE",
			// The mage-x test runner exports MAGE_X_AUTO_DISCOVER_BUILD_TAGS_EXCLUDE
			// to its `go test` subprocess (so `magex test:race` runs these tests with
			// it set, e.g. "mage,windows,..."). It overrides both the default and the
			// YAML fixture, so clear it to assert the configured exclude list.
			"MAGE_X_AUTO_DISCOVER_BUILD_TAGS_EXCLUDE",
			// Binary-name overrides win over both defaults and the YAML fixture, so a
			// leaked value (e.g. config_test.go sets MAGE_X_CUSTOM_BINARY_NAME) would
			// override the asserted Project.Binary. Clear them for a clean baseline.
			"MAGE_X_BINARY_NAME",
			"MAGE_X_CUSTOM_BINARY_NAME",
		)
		defer restoreEnv()

		configYAML := `
project:
  name: my-app
  binary: myapp
  module: github.com/example/myapp

build:
  output: dist
  platforms:
    - linux/amd64
    - darwin/arm64

test:
  cover: true
  race: true
`
		its.RequireNoError(os.WriteFile(".mage.yaml", []byte(configYAML), 0o600))

		TestResetConfig() // Reset
		config, err := GetConfig()
		its.Require().NoError(err)
		its.Equal("my-app", config.Project.Name)
		its.Equal("myapp", config.Project.Binary)
		its.Equal("dist", config.Build.Output)
		its.True(config.Test.Cover)
		its.True(config.Test.Race)

		// The loader unmarshals YAML into a defaultConfig() base
		// (config_provider.go), so fields omitted from the file keep their
		// defaults. The fixture sets no combine/auto-discover keys, so
		// CombineBuildTags must remain at its true default.
		its.True(config.Test.CombineBuildTags,
			"CombineBuildTags should be preserved from defaults when omitted from YAML")
		its.False(config.Test.AutoDiscoverBuildTags,
			"AutoDiscoverBuildTags should stay at its false default when omitted from YAML")
	})

	// Test that the new test-section build-tag keys are actually wired
	// through YAML unmarshalling (proves the assertions above are meaningful).
	its.Run("custom config build tag overrides", func() {
		restoreEnv := unsetEnvVars(
			its.T(),
			"MAGE_X_AUTO_DISCOVER_BUILD_TAGS",
			"MAGE_X_AUTO_DISCOVER_BUILD_TAGS_COMBINE",
			// The mage-x test runner exports MAGE_X_AUTO_DISCOVER_BUILD_TAGS_EXCLUDE
			// to its `go test` subprocess (so `magex test:race` runs these tests with
			// it set, e.g. "mage,windows,..."). It overrides both the default and the
			// YAML fixture, so clear it to assert the configured exclude list.
			"MAGE_X_AUTO_DISCOVER_BUILD_TAGS_EXCLUDE",
			// Binary-name overrides win over both defaults and the YAML fixture, so a
			// leaked value (e.g. config_test.go sets MAGE_X_CUSTOM_BINARY_NAME) would
			// override the asserted Project.Binary. Clear them for a clean baseline.
			"MAGE_X_BINARY_NAME",
			"MAGE_X_CUSTOM_BINARY_NAME",
		)
		defer restoreEnv()

		configYAML := `
project:
  name: my-app
  binary: myapp
  module: github.com/example/myapp

test:
  combine_build_tags: false
  auto_discover_build_tags: true
  auto_discover_build_tags_exclude:
    - integration
    - e2e
`
		its.RequireNoError(os.WriteFile(".mage.yaml", []byte(configYAML), 0o600))

		TestResetConfig() // Reset
		config, err := GetConfig()
		its.Require().NoError(err)
		its.False(config.Test.CombineBuildTags,
			"combine_build_tags: false in YAML must override the true default")
		its.True(config.Test.AutoDiscoverBuildTags,
			"auto_discover_build_tags: true must be read from YAML")
		its.Equal([]string{"integration", "e2e"}, config.Test.AutoDiscoverBuildTagsExclude,
			"auto_discover_build_tags_exclude must be read from YAML")
	})
}

// unsetEnvVars clears the given environment variables for the duration of a
// test and returns a function that restores their original values. It is used
// to keep config-default assertions hermetic regardless of the ambient
// environment (t.Setenv cannot unset a variable).
func unsetEnvVars(t *testing.T, keys ...string) func() {
	t.Helper()

	type saved struct {
		val string
		set bool
	}

	originals := make(map[string]saved, len(keys))
	for _, k := range keys {
		v, ok := os.LookupEnv(k)
		originals[k] = saved{val: v, set: ok}
		require.NoError(t, os.Unsetenv(k))
	}

	return func() {
		for k, s := range originals {
			if s.set {
				require.NoError(t, os.Setenv(k, s.val))
			} else {
				require.NoError(t, os.Unsetenv(k))
			}
		}
	}
}

// TestCacheIntegration tests build caching functionality
func (its *IntegrationTestSuite) TestCacheIntegration() {
	if testing.Short() {
		its.T().Skip("Skipping integration test in short mode")
	}

	// Create test project
	its.createTestProject("cache-test", "cache-test")

	// Create main.go
	mainGo := `package main

func main() {
	println("cache test")
}
`
	its.RequireNoError(os.WriteFile("main.go", []byte(mainGo), 0o600))

	// Setup config
	its.setupMageConfig("cache-test", "cache-app", "cache-test")

	build := Build{}

	// First build - should compile
	start1 := time.Now()
	err := build.Default()
	its.AssertNoError(err)
	duration1 := time.Since(start1)

	// Second build - should use cache (if implemented)
	start2 := time.Now()
	err = build.Default()
	its.AssertNoError(err)
	duration2 := time.Since(start2)

	// Cache hit should be faster (this assumes caching is implemented)
	its.T().Logf("First build: %v, Second build: %v", duration1, duration2)
}

// TestErrorHandling tests error handling across namespaces
func (its *IntegrationTestSuite) TestErrorHandling() {
	if testing.Short() {
		its.T().Skip("Skipping integration test in short mode")
	}

	// Test build error handling
	its.Run("build with missing main", func() {
		// Create go.mod but no main.go file
		its.RequireNoError(os.WriteFile("go.mod", []byte("module test\n\ngo 1.24\n"), 0o600))

		TestSetConfig(&Config{
			Project: ProjectConfig{
				Binary: "test",
				Module: "test",
			},
			Build: BuildConfig{
				Output: "bin",
			},
		})

		build := Build{}
		err := build.Default()
		// Should fail due to missing main
		its.Require().Error(err)
	})

	// Test with invalid configuration
	its.Run("invalid platform", func() {
		build := Build{}
		err := build.Platform("invalid-platform")
		its.Require().Error(err)
	})
}

// TestPerformanceRegression tests for performance regressions
func (its *IntegrationTestSuite) TestPerformanceRegression() {
	if testing.Short() {
		its.T().Skip("Skipping performance test in short mode")
	}

	const perfIterations = 100

	// Test command execution performance
	its.Run("command execution", func() {
		runner := NewSecureCommandRunner()

		its.Require().NoError(runner.RunCmd("echo", "test")) // warm up: discard one-time PATH lookup cost

		start := time.Now()
		for i := 0; i < perfIterations; i++ {
			its.Require().NoError(runner.RunCmd("echo", "test"))
		}
		duration := time.Since(start)

		// Coarse regression guard, not an SLA — see perfBudget for the rationale.
		budget := perfBudget("MAGE_X_PERF_CMD_EXEC_BUDGET", 15*time.Second)
		its.Less(duration, budget,
			"command execution too slow: %d runs took %s (%s/run), budget %s",
			perfIterations, duration, duration/time.Duration(perfIterations), budget)
	})

	// Test configuration loading performance
	its.Run("config loading", func() {
		TestResetConfig()
		_, err := GetConfig() // warm up: first load reads .mage.yaml and initializes providers
		its.Require().NoError(err)

		start := time.Now()
		for i := 0; i < perfIterations; i++ {
			TestResetConfig()
			_, err = GetConfig()
			its.Require().NoError(err)
		}
		duration := time.Since(start)

		// Coarse regression guard, not an SLA — see perfBudget for the rationale.
		budget := perfBudget("MAGE_X_PERF_CONFIG_LOAD_BUDGET", 10*time.Second)
		its.Less(duration, budget,
			"config loading too slow: %d loads took %s (%s/load), budget %s",
			perfIterations, duration, duration/time.Duration(perfIterations), budget)
	})
}

// perfBudget returns the wall-clock ceiling for a performance-regression guard.
// These guards catch order-of-magnitude regressions, not tight SLAs: wall-clock on
// shared CI runners is inherently noisy (CPU contention, cold caches, and very large
// environments that inflate config/env parsing), so the defaults are deliberately
// generous. A pathologically slow runner can raise an individual budget via env,
// e.g. MAGE_X_PERF_CONFIG_LOAD_BUDGET=30s. Invalid or non-positive overrides are
// ignored in favor of the fallback.
func perfBudget(envKey string, fallback time.Duration) time.Duration {
	if v := os.Getenv(envKey); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return fallback
}

// TestIntegrationSuite runs the integration test suite
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}

// TestPerfBudget verifies the env-tunable wall-clock budget used by the
// performance-regression guards: a valid override wins, while unset, malformed,
// and non-positive values fall back to the default.
func TestPerfBudget(t *testing.T) {
	const key = "MAGE_X_PERF_BUDGET_UNITTEST"
	const fallback = 7 * time.Second

	t.Run("falls back when unset", func(t *testing.T) {
		t.Setenv(key, "")
		assert.Equal(t, fallback, perfBudget(key, fallback))
	})

	t.Run("uses a valid override", func(t *testing.T) {
		t.Setenv(key, "12s")
		assert.Equal(t, 12*time.Second, perfBudget(key, fallback))
	})

	t.Run("ignores a malformed override", func(t *testing.T) {
		t.Setenv(key, "not-a-duration")
		assert.Equal(t, fallback, perfBudget(key, fallback))
	})

	t.Run("ignores a non-positive override", func(t *testing.T) {
		t.Setenv(key, "0s")
		assert.Equal(t, fallback, perfBudget(key, fallback))
	})
}
