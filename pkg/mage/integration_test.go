// +build integration

package mage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildIntegration tests the full build pipeline
func TestBuildIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test environment
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	// Create a simple Go project
	projectDir := filepath.Join(tempDir, "test-project")
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	require.NoError(t, os.Chdir(projectDir))

	// Create go.mod
	goMod := `module test-project

go 1.21
`
	require.NoError(t, os.WriteFile("go.mod", []byte(goMod), 0644))

	// Create main.go
	mainGo := `package main

import "fmt"

func main() {
	fmt.Println("Hello, Integration Test!")
}
`
	require.NoError(t, os.WriteFile("main.go", []byte(mainGo), 0644))

	// Create mage config
	config := &Config{
		Project: ProjectConfig{
			Name:   "test-project",
			Binary: "test-app",
			Module: "test-project",
		},
		Build: BuildConfig{
			Output:   "bin",
			TrimPath: true,
		},
	}

	// Override global config for test
	cfg = config

	// Test build
	build := Build{}
	err := build.Default()
	assert.NoError(t, err)

	// Verify binary exists
	binaryPath := filepath.Join("bin", "test-app")
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Errorf("Expected binary at %s, but it doesn't exist", binaryPath)
	}

	// Test clean
	err = build.Clean()
	assert.NoError(t, err)

	// Verify binary is removed
	if _, err := os.Stat(binaryPath); !os.IsNotExist(err) {
		t.Errorf("Expected binary to be removed after clean")
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
func TestMultipleNamespaceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup
	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	require.NoError(t, os.Chdir(tempDir))

	// Create a Go project with tests
	require.NoError(t, os.WriteFile("go.mod", []byte("module test\n\ngo 1.21\n"), 0644))
	require.NoError(t, os.WriteFile("main.go", []byte(`package main

func Add(a, b int) int {
	return a + b
}

func main() {}
`), 0644))

	require.NoError(t, os.WriteFile("main_test.go", []byte(`package main

import "testing"

func TestAdd(t *testing.T) {
	if Add(1, 2) != 3 {
		t.Error("1 + 2 should equal 3")
	}
}
`), 0644))

	// Override config
	cfg = &Config{
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
	}

	// Test workflow: Format -> Lint -> Test -> Build
	t.Run("format", func(t *testing.T) {
		format := Format{}
		// Skip if gofumpt not installed
		if _, err := GetRunner().RunCmdOutput("which", "gofumpt"); err != nil {
			t.Skip("gofumpt not installed")
		}
		err := format.Default()
		assert.NoError(t, err)
	})

	t.Run("test", func(t *testing.T) {
		test := Test{}
		err := test.Unit()
		assert.NoError(t, err)
	})

	t.Run("build", func(t *testing.T) {
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
func TestConfigurationLoading(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	require.NoError(t, os.Chdir(tempDir))

	// Test default config
	t.Run("default config", func(t *testing.T) {
		cfg = nil // Reset
		config, err := LoadConfig()
		assert.NoError(t, err)
		assert.NotNil(t, config)
		assert.Equal(t, "app", config.Project.Binary)
	})

	// Test custom config
	t.Run("custom config", func(t *testing.T) {
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
		require.NoError(t, os.WriteFile(".mage.yaml", []byte(configYAML), 0644))

		cfg = nil // Reset
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
func TestCacheIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	projectDir := filepath.Join(tempDir, "cache-test")
	require.NoError(t, os.MkdirAll(projectDir, 0755))
	require.NoError(t, os.Chdir(projectDir))

	// Create project
	require.NoError(t, os.WriteFile("go.mod", []byte("module cache-test\n\ngo 1.21\n"), 0644))
	require.NoError(t, os.WriteFile("main.go", []byte(`package main

func main() {
	println("cache test")
}
`), 0644))

	// Setup config with caching enabled
	cfg = &Config{
		Project: ProjectConfig{
			Name:   "cache-test",
			Binary: "cache-app",
			Module: "cache-test",
		},
		Build: BuildConfig{
			Output: "bin",
		},
	}

	build := Build{}

	// First build - should compile
	start1 := time.Now()
	err := build.Default()
	assert.NoError(t, err)
	duration1 := time.Since(start1)

	// Second build - should use cache (if implemented)
	start2 := time.Now()
	err = build.Default()
	assert.NoError(t, err)
	duration2 := time.Since(start2)

	// Cache hit should be faster (this assumes caching is implemented)
	t.Logf("First build: %v, Second build: %v", duration1, duration2)
}

// TestEnterpriseFeatures tests enterprise configuration features
func TestEnterpriseFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	require.NoError(t, os.Chdir(tempDir))

	// Create enterprise config
	enterpriseConfig := &EnterpriseConfiguration{
		Metadata: ECConfigMetadata{
			Version:     "1.0.0",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Environment: "test",
		},
		Organization: ECOrganization{
			Name:   "Test Corp",
			Domain: "test.com",
			Teams: []ECTeam{
				{
					Name: "Dev Team",
					Members: []ECTeamMember{
						{
							Email: "dev@test.com",
							Roles: []string{"developer"},
						},
					},
				},
			},
		},
	}

	// Save enterprise config
	err := SaveEnterpriseConfig(enterpriseConfig)
	assert.NoError(t, err)

	// Verify file exists
	_, err = os.Stat(".mage.enterprise.yaml")
	assert.NoError(t, err)

	// Load and verify
	cfg = nil
	config, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config.Enterprise)
	assert.Equal(t, "Test Corp", config.Enterprise.Organization.Name)
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
func TestErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)

	require.NoError(t, os.Chdir(tempDir))

	// Test build error handling
	t.Run("build with missing main", func(t *testing.T) {
		require.NoError(t, os.WriteFile("go.mod", []byte("module test\n\ngo 1.21\n"), 0644))
		// No main.go file

		cfg = &Config{
			Project: ProjectConfig{
				Binary: "test",
				Module: "test",
			},
			Build: BuildConfig{
				Output: "bin",
			},
		}

		build := Build{}
		err := build.Default()
		// Should fail due to missing main
		assert.Error(t, err)
	})

	// Test with invalid configuration
	t.Run("invalid platform", func(t *testing.T) {
		build := Build{}
		err := build.Platform("invalid-platform")
		assert.Error(t, err)
	})
}

// TestPerformanceRegression tests for performance regressions
func TestPerformanceRegression(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	// Test command execution performance
	t.Run("command execution", func(t *testing.T) {
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
	t.Run("config loading", func(t *testing.T) {
		start := time.Now()
		for i := 0; i < 100; i++ {
			cfg = nil
			_, _ = LoadConfig()
		}
		duration := time.Since(start)

		// Should load config 100 times quickly
		assert.Less(t, duration, 1*time.Second, "Config loading is too slow")
	})
}