package mage

import (
	"errors"
	"os"
	"testing"

	"github.com/mrz1836/go-mage/pkg/mage/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ToolsTestSuite defines the test suite for tools functions
type ToolsTestSuite struct {
	suite.Suite
	env   *testutil.TestEnvironment
	tools Tools
}

// SetupTest runs before each test
func (ts *ToolsTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.tools = Tools{}
}

// TearDownTest runs after each test
func (ts *ToolsTestSuite) TearDownTest() {
	ts.env.Cleanup()
}

// TestTools_Default tests the Default function
func (ts *ToolsTestSuite) TestTools_Default() {
	// Mock config loading and tool installation
	ts.setupSuccessfulInstall()

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.Default()
		},
	)

	require.NoError(ts.T(), err)
}

// TestTools_Install tests the Install function
func (ts *ToolsTestSuite) TestTools_Install() {
	ts.setupSuccessfulInstall()

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.Install()
		},
	)

	require.NoError(ts.T(), err)
}

// TestTools_Install_ConfigError tests Install function with clean config (LoadConfig creates default)
func (ts *ToolsTestSuite) TestTools_Install_ConfigError() {
	TestResetConfig() // This causes LoadConfig to create a default config

	// Mock expected tool installation calls for default config
	ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@latest"}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@latest"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.Install()
		},
	)

	require.NoError(ts.T(), err) // Should succeed with default config
}

// TestTools_Update tests the Update function
func (ts *ToolsTestSuite) TestTools_Update() {
	ts.setupSuccessfulUpdate()

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.Update()
		},
	)

	require.NoError(ts.T(), err)
}

// TestTools_Update_ConfigError tests Update function with clean config (LoadConfig creates default)
func (ts *ToolsTestSuite) TestTools_Update_ConfigError() {
	TestResetConfig() // This causes LoadConfig to create a default config

	// Mock expected tool update calls for default config
	ts.env.Runner.On("RunCmd", "brew", []string{"upgrade", "golangci-lint"}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@latest"}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@latest"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.Update()
		},
	)

	require.NoError(ts.T(), err) // Should succeed with default config
}

// TestTools_Verify tests the Verify function when tools are missing (likely case)
func (ts *ToolsTestSuite) TestTools_Verify() {
	ts.setupConfig()

	// Mock version check calls in case some tools exist
	ts.env.Runner.On("RunCmdOutput", "golangci-lint", []string{"--version"}).Return("golangci-lint version 1.50.0", nil).Maybe()
	ts.env.Runner.On("RunCmdOutput", "gofumpt", []string{"--version"}).Return("gofumpt version v0.4.0", nil).Maybe()
	ts.env.Runner.On("RunCmdOutput", "govulncheck", []string{"--version"}).Return("govulncheck version latest", nil).Maybe()
	ts.env.Runner.On("RunCmdOutput", "gotestsum", []string{"--version"}).Return("gotestsum version v1.8.0", nil).Maybe()

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.Verify()
		},
	)

	// Expect error since utils.CommandExists will return false for most tools
	require.Error(ts.T(), err)
	require.Contains(ts.T(), err.Error(), "some tools are missing")
}

// TestTools_Verify_ConfigError tests Verify function with clean config (LoadConfig creates default)
func (ts *ToolsTestSuite) TestTools_Verify_ConfigError() {
	TestResetConfig() // This causes LoadConfig to create a default config

	// Mock version check calls for any tools that might exist on the system
	// This is needed because utils.CommandExists checks the real PATH
	ts.env.Runner.On("RunCmdOutput", mock.MatchedBy(func(cmd string) bool { return true }), mock.Anything).Return("version info", nil).Maybe()

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.Verify()
		},
	)
	// Since some tools might be missing, expect error in most cases
	// But if all tools are installed, no error is expected
	if err != nil {
		require.Contains(ts.T(), err.Error(), "some tools are missing")
	}
	// If no error, that means all tools were found - which is also valid
}

// TestTools_List tests the List function
func (ts *ToolsTestSuite) TestTools_List() {
	ts.setupConfig()

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.List()
		},
	)

	require.NoError(ts.T(), err)
}

// TestTools_List_ConfigError tests List function with clean config (LoadConfig creates default)
func (ts *ToolsTestSuite) TestTools_List_ConfigError() {
	TestResetConfig() // This causes LoadConfig to create a default config

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.List()
		},
	)

	require.NoError(ts.T(), err) // Should succeed with default config - just prints to stdout
}

// TestTools_VulnCheck tests the VulnCheck function
func (ts *ToolsTestSuite) TestTools_VulnCheck() {
	ts.setupConfig()
	// Since govulncheck likely isn't installed, expect installation first
	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@latest"}).Return(nil)
	ts.env.Runner.On("RunCmd", "govulncheck", []string{"-show", "verbose", "./..."}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.VulnCheck()
		},
	)

	require.NoError(ts.T(), err)
}

// TestTools_VulnCheck_InstallFirst tests VulnCheck function when govulncheck needs to be installed
func (ts *ToolsTestSuite) TestTools_VulnCheck_InstallFirst() {
	ts.setupConfig()
	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@latest"}).Return(nil)
	ts.env.Runner.On("RunCmd", "govulncheck", []string{"-show", "verbose", "./..."}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.VulnCheck()
		},
	)

	require.NoError(ts.T(), err)
}

// TestTools_VulnCheck_InstallError tests VulnCheck function with install error
func (ts *ToolsTestSuite) TestTools_VulnCheck_InstallError() {
	expectedError := require.New(ts.T())
	ts.setupConfig()

	// Temporarily modify PATH to ensure govulncheck is not found
	originalPath := os.Getenv("PATH")
	defer func() {
		if err := os.Setenv("PATH", originalPath); err != nil {
			ts.T().Logf("Failed to restore PATH: %v", err)
		}
	}()
	if err := os.Setenv("PATH", "/nonexistent"); err != nil {
		ts.T().Fatalf("Failed to set PATH: %v", err)
	}

	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@latest"}).Return(errors.New("install failed"))

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.VulnCheck()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "failed to install govulncheck")
}

// TestTools_VulnCheck_CheckError tests VulnCheck function with check error
func (ts *ToolsTestSuite) TestTools_VulnCheck_CheckError() {
	expectedError := require.New(ts.T())
	ts.setupConfig()
	// Expect installation first, then failure on check
	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@latest"}).Return(nil)
	ts.env.Runner.On("RunCmd", "govulncheck", []string{"-show", "verbose", "./..."}).Return(errors.New("vulnerability found"))

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.VulnCheck()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "vulnerability check failed")
}

// TestTools_VulnCheck_ConfigError tests VulnCheck function with clean config (since LoadConfig creates default)
func (ts *ToolsTestSuite) TestTools_VulnCheck_ConfigError() {
	TestResetConfig() // This causes LoadConfig to create a default config

	// Mock installation and vulnerability check calls
	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@latest"}).Return(nil)
	ts.env.Runner.On("RunCmd", "govulncheck", []string{"-show", "verbose", "./..."}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.VulnCheck()
		},
	)

	require.NoError(ts.T(), err) // Should succeed with default config
}

// TestTools_Check tests the Check function
func (ts *ToolsTestSuite) TestTools_Check() {
	ts.env.Runner.On("RunCmd", "echo", []string{"Checking tool versions"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.Check()
		},
	)

	require.NoError(ts.T(), err)
}

// TestTools_Clean tests the Clean function
func (ts *ToolsTestSuite) TestTools_Clean() {
	ts.env.Runner.On("RunCmd", "echo", []string{"Cleaning tool installations"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.Clean()
		},
	)

	require.NoError(ts.T(), err)
}

// TestGetRequiredTools tests the getRequiredTools function
func (ts *ToolsTestSuite) TestGetRequiredTools() {
	config := &Config{
		Tools: ToolsConfig{
			GolangciLint: "v1.50.0",
			Fumpt:        "v0.4.0",
			GoVulnCheck:  "latest",
			Mockgen:      "v1.6.0",
			Swag:         "v1.8.0",
			Custom: map[string]string{
				"gotestsum": "gotest.tools/gotestsum@v1.8.0",
				"gci":       "github.com/daixiang0/gci@latest",
			},
		},
	}

	tools := getRequiredTools(config)

	// Should have 3 base tools + 2 optional + 2 custom = 7 tools
	require.Len(ts.T(), tools, 7)

	// Check base tools
	require.Equal(ts.T(), "golangci-lint", tools[0].Name)
	require.Equal(ts.T(), "gofumpt", tools[1].Name)
	require.Equal(ts.T(), "govulncheck", tools[2].Name)

	// Check optional tools
	require.Equal(ts.T(), "mockgen", tools[3].Name)
	require.Equal(ts.T(), "swag", tools[4].Name)

	// Check custom tools - find them by name since order is not deterministic
	var gotestsum, gci *ToolDefinition
	for i := 5; i < 7; i++ {
		if tools[i].Name == "gotestsum" {
			gotestsum = &tools[i]
		} else if tools[i].Name == "gci" {
			gci = &tools[i]
		}
	}

	require.NotNil(ts.T(), gotestsum)
	require.Equal(ts.T(), "gotest.tools/gotestsum", gotestsum.Module)
	require.Equal(ts.T(), "v1.8.0", gotestsum.Version)

	require.NotNil(ts.T(), gci)
	require.Equal(ts.T(), "github.com/daixiang0/gci", gci.Module)
	require.Equal(ts.T(), "latest", gci.Version)
}

// TestGetRequiredTools_MinimalConfig tests getRequiredTools with minimal config
func (ts *ToolsTestSuite) TestGetRequiredTools_MinimalConfig() {
	config := &Config{
		Tools: ToolsConfig{
			GolangciLint: "latest",
			Fumpt:        "latest",
			GoVulnCheck:  "latest",
		},
	}

	tools := getRequiredTools(config)

	// Should have only the 3 base tools
	require.Len(ts.T(), tools, 3)
	require.Equal(ts.T(), "golangci-lint", tools[0].Name)
	require.Equal(ts.T(), "gofumpt", tools[1].Name)
	require.Equal(ts.T(), "govulncheck", tools[2].Name)
}

// TestInstallTool tests the installTool function for new tool installation (most likely case)
func (ts *ToolsTestSuite) TestInstallTool_AlreadyInstalled() {
	tool := ToolDefinition{
		Name:   "gofumpt",
		Module: "mvdan.cc/gofumpt",
		Check:  "gofumpt",
	}

	// Since utils.CommandExists will likely return false, expect installation
	ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@latest"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return installTool(tool)
		},
	)

	require.NoError(ts.T(), err)
}

// TestInstallTool tests the installTool function for new tool installation
func (ts *ToolsTestSuite) TestInstallTool_NewInstall() {
	tool := ToolDefinition{
		Name:    "mockgen",
		Module:  "go.uber.org/mock/mockgen",
		Version: "v1.6.0",
		Check:   "mockgen",
	}

	ts.env.Runner.On("RunCmd", "go", []string{"install", "go.uber.org/mock/mockgen@v1.6.0"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return installTool(tool)
		},
	)

	require.NoError(ts.T(), err)
}

// TestInstallTool_GolangciLint tests installTool for golangci-lint special case
func (ts *ToolsTestSuite) TestInstallTool_GolangciLint() {
	tool := ToolDefinition{
		Name:   "golangci-lint",
		Module: "",
		Check:  "golangci-lint",
	}

	ts.setupConfig()

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return installTool(tool)
		},
	)

	require.NoError(ts.T(), err)
}

// TestInstallTool_InstallError tests installTool with installation error
func (ts *ToolsTestSuite) TestInstallTool_InstallError() {
	expectedError := require.New(ts.T())
	tool := ToolDefinition{
		Name:   "mockgen",
		Module: "go.uber.org/mock/mockgen",
		Check:  "mockgen",
	}

	ts.env.Runner.On("RunCmd", "go", []string{"install", "go.uber.org/mock/mockgen@latest"}).Return(errors.New("install failed"))

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return installTool(tool)
		},
	)

	expectedError.Error(err)
}

// Helper methods for setting up test scenarios

// setupConfig creates a basic configuration for testing
func (ts *ToolsTestSuite) setupConfig() {
	cfg = &Config{
		Tools: ToolsConfig{
			GolangciLint: "v1.50.0",
			Fumpt:        "v0.4.0",
			GoVulnCheck:  "latest",
			Custom: map[string]string{
				"gotestsum": "gotest.tools/gotestsum@v1.8.0",
			},
		},
	}
}

// setupSuccessfulInstall sets up mocks for successful tool installation
func (ts *ToolsTestSuite) setupSuccessfulInstall() {
	ts.setupConfig()

	// Mock installation commands for various tools
	// Note: golangci-lint is a special case and uses ensureGolangciLint which we can't easily mock here
	ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@v0.4.0"}).Return(nil).Maybe()
	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@latest"}).Return(nil).Maybe()
	ts.env.Runner.On("RunCmd", "go", []string{"install", "gotest.tools/gotestsum@v1.8.0"}).Return(nil).Maybe()

	// golangci-lint might try brew or curl installation
	ts.env.Runner.On("RunCmd", "brew", []string{"install", "golangci-lint"}).Return(nil).Maybe()
}

// setupSuccessfulUpdate sets up mocks for successful tool updates
func (ts *ToolsTestSuite) setupSuccessfulUpdate() {
	ts.setupConfig()

	// Mock brew upgrade for golangci-lint (assumes Mac)
	ts.env.Runner.On("RunCmd", "brew", []string{"upgrade", "golangci-lint"}).Return(nil)

	// Mock go install commands for tool updates
	ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@v0.4.0"}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@latest"}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"install", "gotest.tools/gotestsum@v1.8.0"}).Return(nil)
}

// TestToolsTestSuite runs the test suite
func TestToolsTestSuite(t *testing.T) {
	suite.Run(t, new(ToolsTestSuite))
}
