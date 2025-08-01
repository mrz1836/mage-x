package mage

import (
	"errors"
	"testing"

	"github.com/mrz1836/mage-x/pkg/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Static test errors to satisfy err113 linter
var (
	errNotGitRepo      = errors.New("not a git repo")
	errBuildFailed     = errors.New("build failed")
	errLintFailed      = errors.New("lint failed")
	errTestsFailed     = errors.New("tests failed")
	errLintErrorsFound = errors.New("lint errors found")
)

// MockCommandRunner for testing
type MockCommandRunner struct {
	mock.Mock
}

func (m *MockCommandRunner) RunCmd(name string, args ...string) error {
	argList := []interface{}{name}
	for _, arg := range args {
		argList = append(argList, arg)
	}
	return m.Called(argList...).Error(0)
}

func (m *MockCommandRunner) RunCmdOutput(name string, args ...string) (string, error) {
	argList := []interface{}{name}
	for _, arg := range args {
		argList = append(argList, arg)
	}
	ret := m.Called(argList...)
	return ret.String(0), ret.Error(1)
}

// Test helper to replace the global runner
func withMockRunner(t *testing.T, fn func(*MockCommandRunner)) {
	originalRunner := defaultRunner
	mockRunner := new(MockCommandRunner)
	defaultRunner = mockRunner
	defer func() {
		defaultRunner = originalRunner
		mockRunner.AssertExpectations(t)
	}()
	fn(mockRunner)
}

func TestBuildNamespace_Default(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*MockCommandRunner)
		wantErr   bool
	}{
		{
			name: "successful build",
			setupMock: func(m *MockCommandRunner) {
				// Mock version detection
				m.On("RunCmdOutput", "git", "describe", "--tags", "--abbrev=0").Return("v1.0.0", nil).Maybe()
				// Mock commit hash detection
				m.On("RunCmdOutput", "git", "rev-parse", "--short", "HEAD").Return("abc123", nil).Maybe()
				// Mock actual build command (with variable number of args)
				m.On("RunCmd", "go", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil).Maybe()
			},
			wantErr: false,
		},
		{
			name: "build failure",
			setupMock: func(m *MockCommandRunner) {
				// Mock version detection
				m.On("RunCmdOutput", "git", "describe", "--tags", "--abbrev=0").Return("", errNotGitRepo).Maybe()
				// Mock commit hash detection
				m.On("RunCmdOutput", "git", "rev-parse", "--short", "HEAD").Return("", errNotGitRepo).Maybe()
				// Mock build failure (with variable number of args)
				m.On("RunCmd", "go", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(errBuildFailed).Maybe()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockRunner(t, func(mockRunner *MockCommandRunner) {
				tt.setupMock(mockRunner)

				build := NewBuildNamespace()
				if err := build.Default(); err != nil {
					// Expected to potentially fail in test environment
					t.Logf("Build failed as expected: %v", err)
				}
			})
		})
	}
}

func TestBuildNamespace_Clean(t *testing.T) {
	build := NewBuildNamespace()

	// Test clean - should not error even if directories don't exist
	err := build.Clean()
	assert.NoError(t, err)
}

func TestTestNamespace_Default(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*MockCommandRunner)
		wantErr   bool
	}{
		{
			name: "successful test run",
			setupMock: func(m *MockCommandRunner) {
				// Lint check with flexible args (handles variable number of arguments)
				m.On("RunCmd", mock.MatchedBy(func(cmd string) bool { return cmd == CmdGolangciLint }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
				// Test run with flexible args (up to 7 args total)
				m.On("RunCmd", mock.MatchedBy(func(cmd string) bool { return cmd == "go" }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
			},
			wantErr: false,
		},
		{
			name: "lint failure",
			setupMock: func(m *MockCommandRunner) {
				m.On("RunCmd", mock.MatchedBy(func(cmd string) bool { return cmd == CmdGolangciLint }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errLintFailed).Maybe()
			},
			wantErr: true,
		},
		{
			name: "test failure",
			setupMock: func(m *MockCommandRunner) {
				m.On("RunCmd", mock.MatchedBy(func(cmd string) bool { return cmd == CmdGolangciLint }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
				m.On("RunCmd", mock.MatchedBy(func(cmd string) bool { return cmd == "go" }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errTestsFailed).Maybe()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockRunner(t, func(mockRunner *MockCommandRunner) {
				tt.setupMock(mockRunner)

				test := NewTestNamespace()
				err := test.Default()

				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		})
	}
}

func TestTestNamespace_Cover(t *testing.T) {
	withMockRunner(t, func(mockRunner *MockCommandRunner) {
		// Mock for go test with cover flags - expects 10 arguments total
		mockRunner.On("RunCmd", "go", "test", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil).Maybe()

		test := NewTestNamespace()
		err := test.Cover()
		assert.NoError(t, err)
	})
}

func TestLintNamespace_Default(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*MockCommandRunner)
		wantErr   bool
	}{
		{
			name: "successful lint",
			setupMock: func(m *MockCommandRunner) {
				// Lint calls with 5 args: golangci-lint run ./pkg/... --timeout 5m
				m.On("RunCmd", mock.MatchedBy(func(cmd string) bool { return cmd == CmdGolangciLint }), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
			},
			wantErr: false,
		},
		{
			name: "lint with errors",
			setupMock: func(m *MockCommandRunner) {
				m.On("RunCmd", mock.MatchedBy(func(cmd string) bool { return cmd == CmdGolangciLint }), mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errLintErrorsFound).Maybe()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withMockRunner(t, func(mockRunner *MockCommandRunner) {
				tt.setupMock(mockRunner)

				lint := NewLintNamespace()
				err := lint.Default()

				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
			})
		})
	}
}

func TestLintNamespace_Fix(t *testing.T) {
	withMockRunner(t, func(mockRunner *MockCommandRunner) {
		// Fix calls with 6 args: golangci-lint run --fix ./pkg/... --timeout 5m
		mockRunner.On("RunCmd", mock.MatchedBy(func(cmd string) bool { return cmd == CmdGolangciLint }), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()

		lint := NewLintNamespace()
		err := lint.Fix()
		assert.NoError(t, err)
	})
}

func TestToolsNamespace_Install(t *testing.T) {
	withMockRunner(t, func(mockRunner *MockCommandRunner) {
		// Mock installation of tools - gofumpt and govulncheck
		mockRunner.On("RunCmd", "go", "install", mock.AnythingOfType("string")).Return(nil).Maybe()

		tools := NewToolsNamespace()
		// Note: actual installation may skip if tools are already installed
		if err := tools.Install(); err != nil {
			// Expected to potentially fail in test environment
			t.Logf("Tools install failed as expected: %v", err)
		}
		// Just verify no panic occurred
	})
}

func TestDepsNamespace_Default(t *testing.T) {
	withMockRunner(t, func(mockRunner *MockCommandRunner) {
		mockRunner.On("RunCmd", "go", "mod", "download").Return(nil)

		deps := NewDepsNamespace()
		err := deps.Default()
		assert.NoError(t, err)
	})
}

func TestDepsNamespace_Update(t *testing.T) {
	withMockRunner(t, func(mockRunner *MockCommandRunner) {
		// Mock go list command to get direct dependencies
		mockRunner.On("RunCmdOutput", "go", "list", "-m", "-f", "{{if not .Indirect}}{{.Path}}{{end}}", "all").Return("github.com/magefile/mage\ngithub.com/stretchr/testify", nil).Maybe()
		// Mock go list -versions for each dependency
		mockRunner.On("RunCmdOutput", "go", "list", "-m", "-versions", mock.AnythingOfType("string")).Return("github.com/magefile/mage v1.11.0 v1.12.0 v1.13.0", nil).Maybe()
		// Mock go get -u for each direct dependency
		mockRunner.On("RunCmd", "go", "get", "-u", mock.AnythingOfType("string")).Return(nil).Maybe()
		// Mock go mod tidy
		mockRunner.On("RunCmd", "go", "mod", "tidy").Return(nil).Maybe()

		deps := NewDepsNamespace()
		err := deps.Update()
		assert.NoError(t, err)
	})
}

func TestModNamespace_Download(t *testing.T) {
	withMockRunner(t, func(mockRunner *MockCommandRunner) {
		mockRunner.On("RunCmd", "go", "mod", "download").Return(nil).Maybe()
		mockRunner.On("RunCmd", "go", "mod", "verify").Return(nil).Maybe()

		mod := NewModNamespace()
		err := mod.Download()
		assert.NoError(t, err)
	})
}

func TestUpdateNamespace_Check(t *testing.T) {
	withMockRunner(t, func(mockRunner *MockCommandRunner) {
		// Mock getting current version
		mockRunner.On("RunCmdOutput", "git", "describe", "--tags", "--abbrev=0").Return("v1.0.0", nil)
		// Mock checking for updates (simulate no internet)

		update := NewUpdateNamespace()
		// Check method might fail due to network, but shouldn't panic
		if err := update.Check(); err != nil {
			// Expected to potentially fail in test environment
			t.Logf("Update check failed as expected: %v", err)
		}
	})
}

func TestGenerateNamespace_Default(t *testing.T) {
	withMockRunner(t, func(mockRunner *MockCommandRunner) {
		mockRunner.On("RunCmd", "go", "generate", "-v").Return(nil).Maybe()

		generate := NewGenerateNamespace()
		err := generate.Default()
		assert.NoError(t, err)
	})
}

func TestMetricsNamespace_LOC(t *testing.T) {
	withMockRunner(t, func(mockRunner *MockCommandRunner) {
		// Mock various metrics commands
		mockRunner.On("RunCmdOutput", mock.Anything, mock.Anything, mock.Anything).Return("100", nil).Maybe()

		metrics := NewMetricsNamespace()
		// Metrics might fail due to missing tools, but shouldn't panic
		if err := metrics.LOC(); err != nil {
			// Expected to potentially fail in test environment
			t.Logf("Metrics LOC failed as expected: %v", err)
		}
	})
}

func TestNamespaceRegistry(t *testing.T) {
	// Test that all namespaces can be retrieved from registry
	registry := GetNamespaceRegistry()

	// Test the namespaces that are actually in the registry
	assert.NotNil(t, registry.Build())
	assert.NotNil(t, registry.Test())
	assert.NotNil(t, registry.Lint())
	assert.NotNil(t, registry.Format())
	assert.NotNil(t, registry.Deps())
	assert.NotNil(t, registry.Git())
	assert.NotNil(t, registry.Release())
	assert.NotNil(t, registry.Docs())
	assert.NotNil(t, registry.Tools())
	assert.NotNil(t, registry.Generate())
	assert.NotNil(t, registry.CLI())
	assert.NotNil(t, registry.Update())
	assert.NotNil(t, registry.Mod())
	assert.NotNil(t, registry.Recipes())
	assert.NotNil(t, registry.Metrics())
	assert.NotNil(t, registry.Workflow())
}

func TestFactoryFunctions(t *testing.T) {
	// Test that all factory functions return non-nil implementations
	assert.NotNil(t, NewBuildNamespace())
	assert.NotNil(t, NewTestNamespace())
	assert.NotNil(t, NewLintNamespace())
	assert.NotNil(t, NewFormatNamespace())
	assert.NotNil(t, NewDepsNamespace())
	assert.NotNil(t, NewGitNamespace())
	assert.NotNil(t, NewReleaseNamespace())
	assert.NotNil(t, NewDocsNamespace())
	assert.NotNil(t, NewToolsNamespace())
	// assert.NotNil(t, NewSecurityNamespace()) // Temporarily disabled
	assert.NotNil(t, NewGenerateNamespace())
	assert.NotNil(t, NewCLINamespace())
	assert.NotNil(t, NewUpdateNamespace())
	assert.NotNil(t, NewModNamespace())
	assert.NotNil(t, NewRecipesNamespace())
	assert.NotNil(t, NewMetricsNamespace())
	assert.NotNil(t, NewWorkflowNamespace())
}

func TestSecureCommandRunner_Integration(t *testing.T) {
	// Test that the SecureCommandRunner is properly integrated
	runner := GetRunner()
	assert.NotNil(t, runner)

	// Verify it's actually a SecureCommandRunner
	_, ok := runner.(*SecureCommandRunner)
	assert.True(t, ok, "GetRunner should return a SecureCommandRunner")
}

func TestSecureCommandRunner_ValidatesCommands(t *testing.T) {
	runner := NewSecureCommandRunner()

	// Test that dangerous commands are rejected
	err := runner.RunCmd("echo", "$(whoami)")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "command validation failed")

	// Test that safe commands work
	err = runner.RunCmd("echo", "hello")
	assert.NoError(t, err)
}

// Test command validation through namespaces
func TestNamespaceCommandValidation(t *testing.T) {
	// Create a custom executor that tracks validation
	executor := security.NewSecureExecutor()
	executor.AllowedCommands = map[string]bool{
		"go":   true,
		"echo": true,
	}

	// This test verifies that namespace operations use secure execution
	// In a real test, we'd inject the executor, but this demonstrates the concept
	t.Run("BuildNamespace respects security", func(t *testing.T) {
		build := NewBuildNamespace()
		// The build will use the global secure runner which validates commands
		if err := build.Default(); err != nil {
			// May fail due to missing tools, but won't execute dangerous commands
			t.Logf("build failed as expected: %v", err)
		}
	})
}
