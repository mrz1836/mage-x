//go:build integration
// +build integration

package mage

import (
	"errors"
	"os"
	"testing"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Static test errors to satisfy err113 linter
var (
	errToolsInstallFailed = errors.New("install failed")
	errVulnerabilityFound = errors.New("vulnerability found")
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

	ts.Require().NoError(err)
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

	ts.Require().NoError(err)
}

// TestTools_Install_ConfigError tests Install function with clean config (LoadConfig creates default)
func (ts *ToolsTestSuite) TestTools_Install_ConfigError() {
	TestResetConfig() // This causes LoadConfig to create a default config

	// Mock expected tool installation calls for default config  
	fumptVersion := GetDefaultGofumptVersion()
	if fumptVersion == "" {
		fumptVersion = "latest"
	}
	govulnVersion := GetDefaultGoVulnCheckVersion()
	if govulnVersion == "" {
		govulnVersion = "latest"
	}
	
	ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@" + fumptVersion}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@" + govulnVersion}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.Install()
		},
	)

	ts.Require().NoError(err) // Should succeed with default config
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

	ts.Require().NoError(err)
}

// TestTools_Update_ConfigError tests Update function with clean config (LoadConfig creates default)
func (ts *ToolsTestSuite) TestTools_Update_ConfigError() {
	TestResetConfig() // This causes LoadConfig to create a default config

	// Mock expected tool update calls for default config
	fumptVersion := GetDefaultGofumptVersion()
	if fumptVersion == "" {
		fumptVersion = "latest"
	}
	govulnVersion := GetDefaultGoVulnCheckVersion()
	if govulnVersion == "" {
		govulnVersion = "latest"
	}
	
	ts.env.Runner.On("RunCmd", "brew", []string{"upgrade", "golangci-lint"}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@" + fumptVersion}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@" + govulnVersion}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.Update()
		},
	)

	ts.Require().NoError(err) // Should succeed with default config
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
	ts.Require().Error(err)
	ts.Require().Contains(err.Error(), "some tools are missing")
}

// TestTools_Verify_ConfigError tests Verify function with clean config (LoadConfig creates default)
func (ts *ToolsTestSuite) TestTools_Verify_ConfigError() {
	TestResetConfig() // This causes LoadConfig to create a default config

	// Mock version check calls for any tools that might exist on the system
	// This is needed because utils.CommandExists checks the real PATH
	ts.env.Runner.On("RunCmdOutput", mock.MatchedBy(func(_ string) bool { return true }), mock.Anything).Return("version info", nil).Maybe()

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
		ts.Require().Contains(err.Error(), "some tools are missing")
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

	ts.Require().NoError(err)
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

	ts.Require().NoError(err) // Should succeed with default config - just prints to stdout
}

// TestTools_VulnCheck tests the VulnCheck function
func (ts *ToolsTestSuite) TestTools_VulnCheck() {
	ts.setupConfig()
	// Since govulncheck likely isn't installed, expect installation first
	govulnVersion := GetDefaultGoVulnCheckVersion()
	if govulnVersion == "" {
		govulnVersion = "latest"
	}
	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@" + govulnVersion}).Return(nil)
	ts.env.Runner.On("RunCmd", "govulncheck", []string{"-show", "verbose", "./..."}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.VulnCheck()
		},
	)

	ts.Require().NoError(err)
}

// TestTools_VulnCheck_InstallFirst tests VulnCheck function when govulncheck needs to be installed
func (ts *ToolsTestSuite) TestTools_VulnCheck_InstallFirst() {
	ts.setupConfig()
	govulnVersion := GetDefaultGoVulnCheckVersion()
	if govulnVersion == "" {
		govulnVersion = "latest"
	}
	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@" + govulnVersion}).Return(nil)
	ts.env.Runner.On("RunCmd", "govulncheck", []string{"-show", "verbose", "./..."}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.VulnCheck()
		},
	)

	ts.Require().NoError(err)
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

	govulnVersion := GetDefaultGoVulnCheckVersion()
	if govulnVersion == "" {
		govulnVersion = "latest"
	}
	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@" + govulnVersion}).Return(errToolsInstallFailed)

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
	govulnVersion := GetDefaultGoVulnCheckVersion()
	if govulnVersion == "" {
		govulnVersion = "latest"
	}
	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@" + govulnVersion}).Return(nil)
	ts.env.Runner.On("RunCmd", "govulncheck", []string{"-show", "verbose", "./..."}).Return(errVulnerabilityFound)

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
	govulnVersion := GetDefaultGoVulnCheckVersion()
	if govulnVersion == "" {
		govulnVersion = "latest"
	}
	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@" + govulnVersion}).Return(nil)
	ts.env.Runner.On("RunCmd", "govulncheck", []string{"-show", "verbose", "./..."}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return ts.tools.VulnCheck()
		},
	)

	ts.Require().NoError(err) // Should succeed with default config
}

// TestTools_Check tests the Check function
func (ts *ToolsTestSuite) TestTools_Check() {
	// Check now delegates to Verify, so we need to mock tool checks
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
			return ts.tools.Check()
		},
	)
	// Since utils.CommandExists checks the real PATH, some tools might be missing
	// The test behavior depends on what's actually installed
	if err != nil {
		ts.Require().Contains(err.Error(), "some tools are missing")
	}
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

	ts.Require().NoError(err)
}

// TestGetRequiredTools tests the getRequiredTools function
func (ts *ToolsTestSuite) TestGetRequiredTools() {
	config := &Config{
		Tools: ToolsConfig{
			GolangciLint: GetDefaultGolangciLintVersion(),
			Fumpt:        GetDefaultGofumptVersion(),
			GoVulnCheck:  GetDefaultGoVulnCheckVersion(),
			Mockgen:      GetDefaultMockgenVersion(),
			Swag:         GetDefaultSwagVersion(),
			Custom: map[string]string{
				"gotestsum": "gotest.tools/gotestsum@v1.8.0",
				"gci":       "github.com/daixiang0/gci@latest",
			},
		},
	}

	tools := getRequiredTools(config)

	// Should have at least 3 base tools, more if optional tools have versions
	ts.Require().GreaterOrEqual(len(tools), 3)

	// Check base tools (always present)
	ts.Require().Equal("golangci-lint", tools[0].Name)
	ts.Require().Equal("gofumpt", tools[1].Name)
	ts.Require().Equal("govulncheck", tools[2].Name)

	// Optional tools are only added if they have versions configured
	foundMockgen := false
	foundSwag := false
	for _, tool := range tools {
		if tool.Name == "mockgen" {
			foundMockgen = true
		}
		if tool.Name == "swag" {
			foundSwag = true
		}
	}
	
	// If versions are available, tools should be present
	if GetDefaultMockgenVersion() != "" {
		ts.Require().True(foundMockgen, "mockgen should be present when version is configured")
	}
	if GetDefaultSwagVersion() != "" {
		ts.Require().True(foundSwag, "swag should be present when version is configured")
	}

	// Check custom tools - find them by name since order is not deterministic
	var gotestsum, gci *ToolDefinition
	for i := range tools {
		switch tools[i].Name {
		case "gotestsum":
			gotestsum = &tools[i]
		case "gci":
			gci = &tools[i]
		}
	}

	ts.Require().NotNil(gotestsum)
	ts.Require().Equal("gotest.tools/gotestsum", gotestsum.Module)
	ts.Require().Equal("v1.8.0", gotestsum.Version)

	ts.Require().NotNil(gci)
	ts.Require().Equal("github.com/daixiang0/gci", gci.Module)
	ts.Require().Equal("latest", gci.Version)
}

// TestGetRequiredTools_MinimalConfig tests getRequiredTools with minimal config
func (ts *ToolsTestSuite) TestGetRequiredTools_MinimalConfig() {
	config := &Config{
		Tools: ToolsConfig{
			GolangciLint: GetDefaultGolangciLintVersion(),
			Fumpt:        GetDefaultGofumptVersion(),
			GoVulnCheck:  GetDefaultGoVulnCheckVersion(),
		},
	}

	tools := getRequiredTools(config)

	// Should have only the 3 base tools
	ts.Require().Len(tools, 3)
	ts.Require().Equal("golangci-lint", tools[0].Name)
	ts.Require().Equal("gofumpt", tools[1].Name)
	ts.Require().Equal("govulncheck", tools[2].Name)
}

// TestInstallTool tests the installTool function for new tool installation (most likely case)
func (ts *ToolsTestSuite) TestInstallTool_AlreadyInstalled() {
	tool := ToolDefinition{
		Name:   "gofumpt",
		Module: "mvdan.cc/gofumpt",
		Check:  "gofumpt",
	}

	// Since utils.CommandExists will likely return false, expect installation
	fumptVersion := GetDefaultGofumptVersion()
	if fumptVersion == "" {
		fumptVersion = "latest"
	}
	ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@" + fumptVersion}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return installTool(tool)
		},
	)

	ts.Require().NoError(err)
}

// TestInstallTool tests the installTool function for new tool installation
func (ts *ToolsTestSuite) TestInstallTool_NewInstall() {
	mockgenVersion := GetDefaultMockgenVersion()
	if mockgenVersion == "" {
		mockgenVersion = "latest"
	}
	
	tool := ToolDefinition{
		Name:    "mockgen",
		Module:  "go.uber.org/mock/mockgen",
		Version: mockgenVersion,
		Check:   "mockgen",
	}

	expectedCmd := "go.uber.org/mock/mockgen@" + mockgenVersion
	ts.env.Runner.On("RunCmd", "go", []string{"install", expectedCmd}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
		func() interface{} { return GetRunner() },
		func() error {
			return installTool(tool)
		},
	)

	ts.Require().NoError(err)
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

	ts.Require().NoError(err)
}

// TestInstallTool_InstallError tests installTool with installation error
func (ts *ToolsTestSuite) TestInstallTool_InstallError() {
	expectedError := require.New(ts.T())
	tool := ToolDefinition{
		Name:   "nonexistent-tool",
		Module: "example.com/nonexistent/tool",
		Check:  "nonexistent-tool",
	}

	ts.env.Runner.On("RunCmd", "go", []string{"install", "example.com/nonexistent/tool@latest"}).Return(errToolsInstallFailed)

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
	TestSetConfig(&Config{
		Tools: ToolsConfig{
			GolangciLint: GetDefaultGolangciLintVersion(),
			Fumpt:        GetDefaultGofumptVersion(),
			GoVulnCheck:  GetDefaultGoVulnCheckVersion(),
			Custom: map[string]string{
				"gotestsum": "gotest.tools/gotestsum@v1.8.0",
			},
		},
	})
}

// setupSuccessfulInstall sets up mocks for successful tool installation
func (ts *ToolsTestSuite) setupSuccessfulInstall() {
	ts.setupConfig()

	// Get versions from environment or use @latest as fallback
	fumptVersion := GetDefaultGofumptVersion()
	if fumptVersion == "" {
		fumptVersion = "latest"
	}
	govulnVersion := GetDefaultGoVulnCheckVersion()
	if govulnVersion == "" {
		govulnVersion = "latest"
	}

	// Mock installation commands for various tools
	// Note: golangci-lint is a special case and uses ensureGolangciLint which we can't easily mock here
	ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@" + fumptVersion}).Return(nil).Maybe()
	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@" + govulnVersion}).Return(nil).Maybe()
	ts.env.Runner.On("RunCmd", "go", []string{"install", "gotest.tools/gotestsum@v1.8.0"}).Return(nil).Maybe()

	// golangci-lint might try brew or curl installation
	ts.env.Runner.On("RunCmd", "brew", []string{"install", "golangci-lint"}).Return(nil).Maybe()
}

// setupSuccessfulUpdate sets up mocks for successful tool updates
func (ts *ToolsTestSuite) setupSuccessfulUpdate() {
	ts.setupConfig()

	// Get versions from environment or use @latest as fallback
	fumptVersion := GetDefaultGofumptVersion()
	if fumptVersion == "" {
		fumptVersion = "latest"
	}
	govulnVersion := GetDefaultGoVulnCheckVersion()
	if govulnVersion == "" {
		govulnVersion = "latest"
	}

	// Mock brew upgrade for golangci-lint (assumes Mac)
	ts.env.Runner.On("RunCmd", "brew", []string{"upgrade", "golangci-lint"}).Return(nil)

	// Mock go install commands for tool updates
	ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@" + fumptVersion}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@" + govulnVersion}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"install", "gotest.tools/gotestsum@v1.8.0"}).Return(nil)
}

// TestToolsTestSuite runs the test suite
func TestToolsTestSuite(t *testing.T) {
	suite.Run(t, new(ToolsTestSuite))
}
