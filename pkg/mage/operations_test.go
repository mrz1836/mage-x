//go:build integration
// +build integration

package mage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock runner for testing operations
type MockRunner struct {
	commands    [][]string
	shouldError bool
}

func (m *MockRunner) RunCmd(cmd string, args ...string) error {
	fullCmd := append([]string{cmd}, args...)
	m.commands = append(m.commands, fullCmd)
	if m.shouldError {
		return assert.AnError
	}
	return nil
}

func (m *MockRunner) RunCmdOutput(cmd string, args ...string) (string, error) {
	fullCmd := append([]string{cmd}, args...)
	m.commands = append(m.commands, fullCmd)
	if m.shouldError {
		return "", assert.AnError
	}
	return "mock output", nil
}

func (m *MockRunner) GetLastCommand() []string {
	if len(m.commands) == 0 {
		return nil
	}
	return m.commands[len(m.commands)-1]
}

func (m *MockRunner) GetAllCommands() [][]string {
	return m.commands
}

func (m *MockRunner) Reset() {
	m.commands = nil
	m.shouldError = false
}

// OperationsTestHelper provides test utilities for operations testing
type OperationsTestHelper struct {
	originalRunner CommandRunner
	mockRunner     *MockRunner
}

// NewOperationsTestHelper creates a new operations test helper
func NewOperationsTestHelper() *OperationsTestHelper {
	return &OperationsTestHelper{}
}

// SetupMockRunner sets up the mock runner for testing
func (h *OperationsTestHelper) SetupMockRunner(tb testing.TB) {
	h.mockRunner = &MockRunner{}
	h.originalRunner = GetRunner()
	require.NoError(tb, SetRunner(h.mockRunner))
}

// TeardownMockRunner restores the original runner
func (h *OperationsTestHelper) TeardownMockRunner() {
	if h.originalRunner != nil {
		_ = SetRunner(h.originalRunner) //nolint:errcheck // Test cleanup - error not critical
	}
}

// GetMockRunner returns the current mock runner
func (h *OperationsTestHelper) GetMockRunner() *MockRunner {
	return h.mockRunner
}

func TestCheck_Operations(t *testing.T) {
	helper := NewOperationsTestHelper()
	helper.SetupMockRunner(t)
	defer helper.TeardownMockRunner()

	check := Check{}

	t.Run("All", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := check.All()
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.Equal(t, "echo", lastCmd[0])
		assert.Contains(t, lastCmd[1], "Running all checks")
	})

	t.Run("Format", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := check.Format()
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.Equal(t, "gofmt", lastCmd[0])
		assert.Equal(t, "-l", lastCmd[1])
		assert.Equal(t, ".", lastCmd[2])
	})

	t.Run("Security", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := check.Security()
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.Equal(t, "gosec", lastCmd[0])
		assert.Equal(t, "./...", lastCmd[1])
	})

	t.Run("Tidy", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := check.Tidy()
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.Equal(t, "go", lastCmd[0])
		assert.Equal(t, "mod", lastCmd[1])
		assert.Equal(t, "tidy", lastCmd[2])
	})

	t.Run("Error handling", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		mockRunner.shouldError = true
		err := check.All()
		assert.Error(t, err)
	})
}

func TestCI_Operations(t *testing.T) {
	helper := NewOperationsTestHelper()
	helper.SetupMockRunner(t)
	defer helper.TeardownMockRunner()

	ci := CI{}

	t.Run("Run", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := ci.Run("test")
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		require.Len(t, lastCmd, 3) // ["echo", "Running CI job:", "test"]
		assert.Equal(t, "echo", lastCmd[0])
		assert.Equal(t, "Running CI job:", lastCmd[1])
		assert.Equal(t, "test", lastCmd[2])
	})

	t.Run("Validate", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := ci.Validate()
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		require.Len(t, lastCmd, 2) // ["echo", "Validating CI configuration"]
		assert.Equal(t, "echo", lastCmd[0])
		assert.Equal(t, "Validating CI configuration", lastCmd[1])
	})

	t.Run("Status", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := ci.Status("main")
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		require.Len(t, lastCmd, 3) // ["echo", "Checking CI status for branch:", "main"]
		assert.Equal(t, "echo", lastCmd[0])
		assert.Equal(t, "Checking CI status for branch:", lastCmd[1])
		assert.Equal(t, "main", lastCmd[2])
	})

	t.Run("Cache", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := ci.Cache("clear")
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		require.Len(t, lastCmd, 3) // ["echo", "Managing CI cache:", "clear"]
		assert.Equal(t, "echo", lastCmd[0])
		assert.Equal(t, "Managing CI cache:", lastCmd[1])
		assert.Equal(t, "clear", lastCmd[2])
	})
}

func TestMonitor_Operations(t *testing.T) {
	helper := NewOperationsTestHelper()
	helper.SetupMockRunner(t)
	defer helper.TeardownMockRunner()

	monitor := Monitor{}

	t.Run("Health", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := monitor.Health("http://localhost:8080/health")
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		require.Len(t, lastCmd, 3) // ["echo", "Checking health of endpoint:", "http://localhost:8080/health"]
		assert.Equal(t, "echo", lastCmd[0])
		assert.Equal(t, "Checking health of endpoint:", lastCmd[1])
		assert.Equal(t, "http://localhost:8080/health", lastCmd[2])
	})

	t.Run("Metrics", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := monitor.Metrics("1h")
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		require.Len(t, lastCmd, 3) // ["echo", "Fetching metrics for time range:", "1h"]
		assert.Equal(t, "echo", lastCmd[0])
		assert.Equal(t, "Fetching metrics for time range:", lastCmd[1])
		assert.Equal(t, "1h", lastCmd[2])
	})

	t.Run("Logs", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := monitor.Logs("api-service")
		require.NoError(t, err)

		commands := mockRunner.GetAllCommands()
		require.Len(t, commands, 1)

		lastCmd := commands[0]
		require.Len(t, lastCmd, 3) // ["echo", "Viewing logs for service:", "api-service"]
		assert.Equal(t, "echo", lastCmd[0])
		assert.Equal(t, "Viewing logs for service:", lastCmd[1])
		assert.Equal(t, "api-service", lastCmd[2])
	})
}

func TestDatabase_Operations(t *testing.T) {
	helper := NewOperationsTestHelper()
	helper.SetupMockRunner(t)
	defer helper.TeardownMockRunner()

	db := Database{}

	t.Run("Migrate", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := db.Migrate("up")
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		require.Len(t, lastCmd, 3) // ["echo", "Running database migration:", "up"]
		assert.Equal(t, "echo", lastCmd[0])
		assert.Equal(t, "Running database migration:", lastCmd[1])
		assert.Equal(t, "up", lastCmd[2])
	})

	t.Run("Seed", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := db.Seed("test_data")
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		require.Len(t, lastCmd, 3) // ["echo", "Seeding database:", "test_data"]
		assert.Equal(t, "echo", lastCmd[0])
		assert.Equal(t, "Seeding database:", lastCmd[1])
		assert.Equal(t, "test_data", lastCmd[2])
	})

	t.Run("Reset", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := db.Reset()
		require.NoError(t, err)

		commands := mockRunner.GetAllCommands()
		require.GreaterOrEqual(t, len(commands), 1)

		// Should run multiple commands for reset
		assert.GreaterOrEqual(t, len(commands), 1)
	})

	t.Run("Backup", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := db.Backup("backup.sql")
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		require.Len(t, lastCmd, 3) // ["echo", "Creating database backup:", "backup.sql"]
		assert.Equal(t, "echo", lastCmd[0])
		assert.Equal(t, "Creating database backup:", lastCmd[1])
		assert.Equal(t, "backup.sql", lastCmd[2])
	})
}

func TestDeploy_Operations(t *testing.T) {
	helper := NewOperationsTestHelper()
	helper.SetupMockRunner(t)
	defer helper.TeardownMockRunner()

	deploy := Deploy{}

	t.Run("Staging", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := deploy.Staging()
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		require.Len(t, lastCmd, 2) // ["echo", "Deploying to staging"]
		assert.Equal(t, "echo", lastCmd[0])
		assert.Equal(t, "Deploying to staging", lastCmd[1])
	})

	t.Run("Production", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := deploy.Production()
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		require.Len(t, lastCmd, 2) // ["echo", "Deploying to production"]
		assert.Equal(t, "echo", lastCmd[0])
		assert.Equal(t, "Deploying to production", lastCmd[1])
	})

	t.Run("Status", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := deploy.Status("production")
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		require.Len(t, lastCmd, 3) // ["echo", "Checking deployment status for:", "production"]
		assert.Equal(t, "echo", lastCmd[0])
		assert.Equal(t, "Checking deployment status for:", lastCmd[1])
		assert.Equal(t, "production", lastCmd[2])
	})

	t.Run("Rollback", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := deploy.Rollback("production", "v1.0.0")
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.Equal(t, "echo", lastCmd[0])
		// Check the complete message in the second argument
		assert.Contains(t, lastCmd[1], "Rolling")
		assert.Contains(t, lastCmd[1], "back")
		assert.Contains(t, lastCmd[2], "production")
		assert.Contains(t, lastCmd[4], "v1.0.0")
	})
}

func TestClean_Operations(t *testing.T) {
	helper := NewOperationsTestHelper()
	helper.SetupMockRunner(t)
	defer helper.TeardownMockRunner()

	clean := Clean{}

	t.Run("All", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := clean.All()
		require.NoError(t, err)

		commands := mockRunner.GetAllCommands()
		require.GreaterOrEqual(t, len(commands), 1)

		// Should run multiple cleanup commands
		assert.GreaterOrEqual(t, len(commands), 1)
	})

	t.Run("Build", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := clean.Build()
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.Equal(t, "go", lastCmd[0])
		assert.Equal(t, "clean", lastCmd[1])
	})

	t.Run("Cache", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := clean.Cache()
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.Equal(t, "go", lastCmd[0])
		assert.Equal(t, "clean", lastCmd[1])
		assert.Equal(t, "-cache", lastCmd[2])
	})

	t.Run("Dependencies", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := clean.Dependencies()
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.Equal(t, "go", lastCmd[0])
		assert.Equal(t, "clean", lastCmd[1])
		assert.Equal(t, "-modcache", lastCmd[2])
	})
}

func TestRun_Operations(t *testing.T) {
	helper := NewOperationsTestHelper()
	helper.SetupMockRunner(t)
	defer helper.TeardownMockRunner()

	run := Run{}

	t.Run("Dev", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := run.Dev()
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.Equal(t, "echo", lastCmd[0])
		assert.Contains(t, lastCmd[1], "Running")
		assert.Contains(t, lastCmd[1], "dev")
	})

	t.Run("Watch", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := run.Watch()
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.Equal(t, "echo", lastCmd[0])
		assert.Contains(t, lastCmd[1], "Running")
		assert.Contains(t, lastCmd[1], "watch")
	})

	t.Run("Debug", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := run.Debug()
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.Equal(t, "echo", lastCmd[0])
		assert.Contains(t, lastCmd[1], "Running")
		assert.Contains(t, lastCmd[1], "debug")
	})
}

func TestServe_Operations(t *testing.T) {
	helper := NewOperationsTestHelper()
	helper.SetupMockRunner(t)
	defer helper.TeardownMockRunner()

	serve := Serve{}

	t.Run("HTTP", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := serve.HTTP()
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.Equal(t, "echo", lastCmd[0])
		assert.Contains(t, lastCmd[1], "Serving")
		assert.Contains(t, lastCmd[1], "HTTP")
	})

	t.Run("HTTPS", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := serve.HTTPS()
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.Equal(t, "echo", lastCmd[0])
		assert.Contains(t, lastCmd[1], "Serving")
		assert.Contains(t, lastCmd[1], "HTTPS")
		// Should have HTTPS configuration
		assert.GreaterOrEqual(t, len(lastCmd), 2)
	})

	t.Run("Static", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := serve.Static()
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.GreaterOrEqual(t, len(lastCmd), 1)
	})
}

func TestDockerOps_Operations(t *testing.T) {
	helper := NewOperationsTestHelper()
	helper.SetupMockRunner(t)
	defer helper.TeardownMockRunner()

	docker := DockerOps{}

	t.Run("Build", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := docker.Build("test-app")
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.Equal(t, "docker", lastCmd[0])
		assert.Equal(t, "build", lastCmd[1])
		assert.Equal(t, "-t", lastCmd[2])
		assert.Equal(t, "test-app", lastCmd[3])
		assert.Equal(t, ".", lastCmd[4])
	})

	t.Run("Run", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := docker.Run("test-app")
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.Equal(t, "docker", lastCmd[0])
		assert.Equal(t, "run", lastCmd[1])
		assert.Equal(t, "test-app", lastCmd[2])
	})

	t.Run("Push", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := docker.Push("test-app")
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.Equal(t, "docker", lastCmd[0])
		assert.Equal(t, "push", lastCmd[1])
		assert.Equal(t, "test-app", lastCmd[2])
	})

	t.Run("Stop", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := docker.Stop("test-app")
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.Equal(t, "docker", lastCmd[0])
		assert.Equal(t, "stop", lastCmd[1])
		assert.Equal(t, "test-app", lastCmd[2])
	})

	t.Run("Clean", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := docker.Clean()
		require.NoError(t, err)

		commands := mockRunner.GetAllCommands()
		require.GreaterOrEqual(t, len(commands), 1)

		// Should run docker cleanup commands
		for _, cmd := range commands {
			assert.Equal(t, "docker", cmd[0])
		}
	})
}

func TestCommon_Operations(t *testing.T) {
	helper := NewOperationsTestHelper()
	helper.SetupMockRunner(t)
	defer helper.TeardownMockRunner()

	common := Common{}

	t.Run("Version", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := common.Version()
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.Equal(t, "echo", lastCmd[0])
		assert.Contains(t, lastCmd[1], "Getting")
		assert.Contains(t, lastCmd[1], "version")
	})

	t.Run("Duration", func(t *testing.T) {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := common.Duration()
		require.NoError(t, err)

		lastCmd := mockRunner.GetLastCommand()
		require.NotNil(t, lastCmd)
		assert.Equal(t, "echo", lastCmd[0])
		assert.Contains(t, lastCmd[1], "Getting")
		assert.Contains(t, lastCmd[1], "duration")
	})
}

// Test struct instantiation
func TestOperationStructs(t *testing.T) {
	t.Run("struct creation", func(t *testing.T) {
		check := Check{}
		ci := CI{}
		monitor := Monitor{}
		database := Database{}
		deploy := Deploy{}
		clean := Clean{}
		run := Run{}
		serve := Serve{}
		docker := DockerOps{}
		common := Common{}

		// Test that structs can be instantiated
		assert.NotNil(t, check)
		assert.NotNil(t, ci)
		assert.NotNil(t, monitor)
		assert.NotNil(t, database)
		assert.NotNil(t, deploy)
		assert.NotNil(t, clean)
		assert.NotNil(t, run)
		assert.NotNil(t, serve)
		assert.NotNil(t, docker)
		assert.NotNil(t, common)
	})

	t.Run("Docker alias", func(t *testing.T) {
		// Test that Docker is an alias for DockerOps
		var docker Docker
		var dockerOps DockerOps

		// Should be the same type
		assert.IsType(t, dockerOps, docker)
	})
}

// Benchmark tests for operations
func BenchmarkCheck_All(b *testing.B) {
	helper := NewOperationsTestHelper()
	helper.SetupMockRunner(b)
	defer helper.TeardownMockRunner()

	check := Check{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := check.All()
		_ = err // Ignore errors in benchmark - we're testing performance
	}
}

func BenchmarkCI_Test(b *testing.B) {
	helper := NewOperationsTestHelper()
	helper.SetupMockRunner(b)
	defer helper.TeardownMockRunner()

	ci := CI{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := ci.Run("test")
		_ = err // Ignore errors in benchmark - we're testing performance
	}
}

func BenchmarkDocker_Build(b *testing.B) {
	helper := NewOperationsTestHelper()
	helper.SetupMockRunner(b)
	defer helper.TeardownMockRunner()

	docker := DockerOps{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mockRunner := helper.GetMockRunner()
		mockRunner.Reset()
		err := docker.Build("test")
		_ = err // Ignore errors in benchmark - we're testing performance
	}
}

// Test error propagation
func TestOperations_ErrorHandling(t *testing.T) {
	helper := NewOperationsTestHelper()
	helper.SetupMockRunner(t)
	defer helper.TeardownMockRunner()

	mockRunner := helper.GetMockRunner()
	mockRunner.shouldError = true

	t.Run("Check operations error handling", func(t *testing.T) {
		check := Check{}
		require.Error(t, check.All())
		require.Error(t, check.Format())
		require.Error(t, check.Security())
		assert.Error(t, check.Tidy())
	})

	t.Run("CI operations error handling", func(t *testing.T) {
		ci := CI{}
		require.Error(t, ci.Run("test"))
		require.Error(t, ci.Validate())
		require.Error(t, ci.Status("main"))
		assert.Error(t, ci.Cache("clear"))
	})

	t.Run("Docker operations error handling", func(t *testing.T) {
		docker := DockerOps{}
		require.Error(t, docker.Build("test"))
		require.Error(t, docker.Run("test"))
		require.Error(t, docker.Push("test"))
		assert.Error(t, docker.Stop("test"))
	})
}
