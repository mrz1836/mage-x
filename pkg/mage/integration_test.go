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
func (its *IntegrationTestSuite) createTestProject(name, module string) string {
	its.T().Helper()

	projectDir := filepath.Join(its.TmpDir, name)
	its.RequireNoError(os.MkdirAll(projectDir, 0o755))
	its.RequireNoError(os.Chdir(projectDir))

	// Create go.mod
	goMod := `module ` + module + `

go 1.24
`
	its.RequireNoError(os.WriteFile("go.mod", []byte(goMod), 0o644))

	return projectDir
}

// setupGoProject creates a complete Go project with main.go
func (its *IntegrationTestSuite) setupGoProject(dir, module, mainContent string) {
	its.T().Helper()

	if mainContent == "" {
		mainContent = `package main

import "fmt"

func main() {
	fmt.Println("Hello, Integration Test!")
}
`
	}

	its.RequireNoError(os.WriteFile("main.go", []byte(mainContent), 0o644))
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
	its.setupGoProject("", "test-project", "")
	its.setupMageConfig("test-project", "test-app", "test-project")

	// Test build
	build := Build{}
	err := build.Default()
	its.AssertNoError(err)

	// Verify binary exists
	binaryPath := filepath.Join("bin", "test-app")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
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
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
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
	its.RequireNoError(os.WriteFile("main.go", []byte(mainGo), 0o644))

	// Create test file
	testGo := `package main

import "testing"

func TestAdd(t *testing.T) {
	if Add(1, 2) != 3 {
		t.Error("1 + 2 should equal 3")
	}
}
`
	its.RequireNoError(os.WriteFile("main_test.go", []byte(testGo), 0o644))

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
			Parallel: true,
			Cover:    true,
		},
		Lint: LintConfig{
			Timeout: "1m",
		},
	})

	// Test workflow: Format -> Test -> Build
	its.T().Run("format", func(t *testing.T) {
		format := Format{}
		// Skip if gofumpt not installed
		if _, err := GetRunner().RunCmdOutput("which", "gofumpt"); err != nil {
			t.Skip("gofumpt not installed")
		}
		err := format.Default()
		assert.NoError(t, err)
	})

	its.T().Run("test", func(t *testing.T) {
		test := Test{}
		err := test.Unit()
		assert.NoError(t, err)
	})

	its.T().Run("build", func(t *testing.T) {
		build := Build{}
		err := build.Default()
		assert.NoError(t, err)
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
	its.T().Run("default config", func(t *testing.T) {
		// Create a completely isolated directory with no go.mod
		isolatedDir := filepath.Join(its.TmpDir, "isolated")
		its.RequireNoError(os.MkdirAll(isolatedDir, 0o755))
		its.RequireNoError(os.Chdir(isolatedDir))

		TestResetConfig() // Reset
		config, err := LoadConfig()
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "app", config.Project.Binary)
	})

	// Test custom config
	its.T().Run("custom config", func(t *testing.T) {
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
		its.RequireNoError(os.WriteFile(".mage.yaml", []byte(configYAML), 0o644))

		TestResetConfig() // Reset
		config, err := LoadConfig()
		assert.NoError(t, err)
		assert.Equal(t, "my-app", config.Project.Name)
		assert.Equal(t, "myapp", config.Project.Binary)
		assert.Equal(t, "dist", config.Build.Output)
		assert.True(t, config.Test.Cover)
		assert.True(t, config.Test.Race)
	})
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
	its.RequireNoError(os.WriteFile("main.go", []byte(mainGo), 0o644))

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

// TestEnterpriseFeatures tests enterprise configuration features
func (its *IntegrationTestSuite) TestEnterpriseFeatures() {
	if testing.Short() {
		its.T().Skip("Skipping integration test in short mode")
	}

	// Create enterprise config
	enterpriseConfig := &EnterpriseConfiguration{
		Metadata: ECConfigMetadata{
			Version:   "1.0.0",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Organization: OrganizationConfig{
			Name:   "Test Corp",
			Domain: "test.com",
		},
	}

	// Save enterprise config
	err := SaveEnterpriseConfig(enterpriseConfig)
	its.AssertNoError(err)

	// Verify file exists
	_, err = os.Stat(".mage.enterprise.yaml")
	its.AssertNoError(err)

	// Load and verify
	TestResetConfig()
	config, err := LoadConfig()
	its.AssertNoError(err)
	its.NotNil(config.Enterprise)
	its.Equal("Test Corp", config.Enterprise.Organization.Name)
}

// TestWorkflowExecution tests workflow execution
func TestWorkflowExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	workflow := Workflow{}

	// Test listing workflows
	t.Run("list workflows", func(t *testing.T) {
		// This should work even without workflows defined
		err := workflow.List()
		// Should not error, just show empty or default list
		assert.NoError(t, err)
	})
}

// TestErrorHandling tests error handling across namespaces
func (its *IntegrationTestSuite) TestErrorHandling() {
	if testing.Short() {
		its.T().Skip("Skipping integration test in short mode")
	}

	// Test build error handling
	its.T().Run("build with missing main", func(t *testing.T) {
		// Create go.mod but no main.go file
		its.RequireNoError(os.WriteFile("go.mod", []byte("module test\n\ngo 1.24\n"), 0o644))

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
		assert.Error(t, err)
	})

	// Test with invalid configuration
	its.T().Run("invalid platform", func(t *testing.T) {
		build := Build{}
		err := build.Platform("invalid-platform")
		assert.Error(t, err)
	})
}

// TestPerformanceRegression tests for performance regressions
func (its *IntegrationTestSuite) TestPerformanceRegression() {
	if testing.Short() {
		its.T().Skip("Skipping performance test in short mode")
	}

	// Test command execution performance
	its.T().Run("command execution", func(t *testing.T) {
		runner := NewSecureCommandRunner()

		start := time.Now()
		for i := 0; i < 100; i++ {
			_ = runner.RunCmd("echo", "test")
		}
		duration := time.Since(start)

		// Should complete 100 echo commands in reasonable time
		assert.Less(t, duration, 5*time.Second, "Command execution is too slow")
	})

	// Test configuration loading performance
	its.T().Run("config loading", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < 100; i++ {
			TestResetConfig()
			_, _ = LoadConfig()
		}
		duration := time.Since(start)

		// Should load config 100 times quickly
		assert.Less(t, duration, 1*time.Second, "Config loading is too slow")
	})
}

// TestIntegrationSuite runs the integration test suite
func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
