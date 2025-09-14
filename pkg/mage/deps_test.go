//go:build integration
// +build integration

package mage

import (
	"errors"
	"testing"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Static test errors to satisfy err113 linter
var (
	errListFailed          = errors.New("list failed")
	errTidyFailed          = errors.New("tidy failed")
	errOutdatedCheckFailed = errors.New("outdated check failed")
)

// DepsTestSuite defines the test suite for deps functions
type DepsTestSuite struct {
	suite.Suite

	env  *testutil.TestEnvironment
	deps Deps
}

// SetupTest runs before each test
func (ts *DepsTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.deps = Deps{}
}

// TearDownTest runs after each test
func (ts *DepsTestSuite) TearDownTest() {
	ts.env.Cleanup()
}

// TestDeps_Default tests the Default function
func (ts *DepsTestSuite) TestDeps_Default() {
	ts.env.Builder.ExpectGoCommand("mod", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Default()
		},
	)

	ts.Require().NoError(err)
}

// TestDeps_Download tests the Download function
func (ts *DepsTestSuite) TestDeps_Download() {
	ts.env.Builder.ExpectGoCommand("mod", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Download()
		},
	)

	ts.Require().NoError(err)
}

// TestDeps_Download_Error tests Download function with error
func (ts *DepsTestSuite) TestDeps_Download_Error() {
	expectedError := require.New(ts.T())
	ts.env.Builder.ExpectFailure("download failed")

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Download()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "failed to download dependencies")
}

// TestDeps_Tidy tests the Tidy function
func (ts *DepsTestSuite) TestDeps_Tidy() {
	ts.env.Builder.ExpectGoCommand("mod", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Tidy()
		},
	)

	ts.Require().NoError(err)
}

// TestDeps_Tidy_Error tests Tidy function with error
func (ts *DepsTestSuite) TestDeps_Tidy_Error() {
	expectedError := require.New(ts.T())
	ts.env.Builder.ExpectFailure("tidy failed")

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Tidy()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "failed to tidy dependencies")
}

// TestDeps_Update tests the Update function
func (ts *DepsTestSuite) TestDeps_Update() {
	// Set up expectations for the Update function
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-f", "{{if not .Indirect}}{{.Path}}{{end}}", "all"}).Return("github.com/stretchr/testify\ngithub.com/pkg/errors", nil)

	// Mock current version queries (new requirement)
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "github.com/stretchr/testify"}).Return("github.com/stretchr/testify v1.8.0", nil)
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "github.com/pkg/errors"}).Return("github.com/pkg/errors v0.9.0", nil)

	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-versions", "github.com/stretchr/testify"}).Return("github.com/stretchr/testify v1.8.0 v1.8.1 v1.9.0", nil)
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-versions", "github.com/pkg/errors"}).Return("github.com/pkg/errors v0.9.0 v0.9.1", nil)
	ts.env.Runner.On("RunCmd", "go", []string{"get", "-u", "github.com/stretchr/testify@v1.9.0"}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"get", "-u", "github.com/pkg/errors@v0.9.1"}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Update()
		},
	)

	ts.Require().NoError(err)
}

// TestDeps_Update_ListError tests Update function with list error
func (ts *DepsTestSuite) TestDeps_Update_ListError() {
	expectedError := require.New(ts.T())
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-f", "{{if not .Indirect}}{{.Path}}{{end}}", "all"}).Return("", errListFailed)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Update()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "failed to list dependencies")
}

// TestDeps_Update_TidyError tests Update function with tidy error at the end
func (ts *DepsTestSuite) TestDeps_Update_TidyError() {
	expectedError := require.New(ts.T())
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-f", "{{if not .Indirect}}{{.Path}}{{end}}", "all"}).Return("github.com/stretchr/testify", nil)

	// Mock current version query (new requirement)
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "github.com/stretchr/testify"}).Return("github.com/stretchr/testify v1.8.0", nil)

	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-versions", "github.com/stretchr/testify"}).Return("github.com/stretchr/testify v1.8.0 v1.9.0", nil)
	ts.env.Runner.On("RunCmd", "go", []string{"get", "-u", "github.com/stretchr/testify@v1.9.0"}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(errTidyFailed)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Update()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "failed to tidy dependencies")
}

// TestDeps_UpdateWithArgs_AllowMajor tests UpdateWithArgs with allow-major=true
func (ts *DepsTestSuite) TestDeps_UpdateWithArgs_AllowMajor() {
	// Mock the list of direct dependencies command
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-f", "{{if not .Indirect}}{{.Path}}{{end}}", "all"}).Return("github.com/pkg/errors\ngithub.com/stretchr/testify", nil)

	// Mock current version queries
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "github.com/pkg/errors"}).Return("github.com/pkg/errors v0.8.1", nil)
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "github.com/stretchr/testify"}).Return("github.com/stretchr/testify v1.8.4", nil)

	// Mock versions command to return newer major versions
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-versions", "github.com/pkg/errors"}).Return("github.com/pkg/errors v0.8.0 v0.8.1 v0.9.0 v0.9.1 v2.0.0", nil)
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-versions", "github.com/stretchr/testify"}).Return("github.com/stretchr/testify v1.8.0 v1.8.4 v1.9.0", nil)

	// Mock update commands - should update to latest including major version
	ts.env.Runner.On("RunCmd", "go", []string{"get", "-u", "github.com/pkg/errors@v2.0.0"}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"get", "-u", "github.com/stretchr/testify@v1.9.0"}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.UpdateWithArgs("allow-major")
		},
	)

	ts.Require().NoError(err)
}

// TestDeps_UpdateWithArgs_NoMajor tests UpdateWithArgs without allow-major (default)
func (ts *DepsTestSuite) TestDeps_UpdateWithArgs_NoMajor() {
	// Mock the list of direct dependencies command
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-f", "{{if not .Indirect}}{{.Path}}{{end}}", "all"}).Return("github.com/pkg/errors\ngithub.com/stretchr/testify", nil)

	// Mock current version queries
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "github.com/pkg/errors"}).Return("github.com/pkg/errors v0.8.1", nil)
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "github.com/stretchr/testify"}).Return("github.com/stretchr/testify v1.8.4", nil)

	// Mock versions command to return newer major versions
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-versions", "github.com/pkg/errors"}).Return("github.com/pkg/errors v0.8.0 v0.8.1 v0.9.0 v0.9.1 v2.0.0", nil)
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-versions", "github.com/stretchr/testify"}).Return("github.com/stretchr/testify v1.8.0 v1.8.4 v1.9.0", nil)

	// Mock update commands - should skip major version update, only update minor
	ts.env.Runner.On("RunCmd", "go", []string{"get", "-u", "github.com/stretchr/testify@v1.9.0"}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.UpdateWithArgs()
		},
	)

	ts.Require().NoError(err)
}

// TestDeps_UpdateWithArgs_AllowMajorFalse tests UpdateWithArgs with allow-major=false explicitly
func (ts *DepsTestSuite) TestDeps_UpdateWithArgs_AllowMajorFalse() {
	// Mock the list of direct dependencies command
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-f", "{{if not .Indirect}}{{.Path}}{{end}}", "all"}).Return("github.com/pkg/errors", nil)

	// Mock current version queries
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "github.com/pkg/errors"}).Return("github.com/pkg/errors v0.8.1", nil)

	// Mock versions command to return newer major versions
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-versions", "github.com/pkg/errors"}).Return("github.com/pkg/errors v0.8.0 v0.8.1 v0.9.0 v0.9.1 v2.0.0", nil)

	// Mock tidy command (no updates should happen)
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.UpdateWithArgs("allow-major=false")
		},
	)

	ts.Require().NoError(err)
}

// TestIsMajorVersionUpdate tests the isMajorVersionUpdate helper function
func (ts *DepsTestSuite) TestIsMajorVersionUpdate() {
	tests := []struct {
		name     string
		current  string
		latest   string
		expected bool
	}{
		{"v1 to v2", "v1.0.0", "v2.0.0", true},
		{"v1.2.3 to v2.0.0", "v1.2.3", "v2.0.0", true},
		{"v0.8.1 to v2.0.0", "v0.8.1", "v2.0.0", true},
		{"v1.8.4 to v1.9.0", "v1.8.4", "v1.9.0", false},
		{"v2.0.0 to v2.1.0", "v2.0.0", "v2.1.0", false},
		{"v1.0.0 to v1.0.1", "v1.0.0", "v1.0.1", false},
		{"same version", "v1.0.0", "v1.0.0", false},
		{"v2 to v3", "v2.1.0", "v3.0.0", true},
		{"no v prefix", "1.0.0", "2.0.0", true},
		{"mixed prefix", "v1.0.0", "2.0.0", true},
		{"complex version", "v1.0.0-alpha", "v2.0.0-beta", true},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			result := isMajorVersionUpdate(test.current, test.latest)
			ts.Equal(test.expected, result, "Expected %v for %s -> %s", test.expected, test.current, test.latest)
		})
	}
}

// TestExtractMajorVersion tests the extractMajorVersion helper function
func (ts *DepsTestSuite) TestExtractMajorVersion() {
	tests := []struct {
		name     string
		version  string
		expected int
	}{
		{"simple major", "1", 1},
		{"major with suffix", "2.0.0", 2},
		{"major with alpha", "3-alpha", 3},
		{"major with beta", "1-beta.1", 1},
		{"zero major", "0", 0},
		{"empty string", "", 0},
		{"no digits", "alpha", 0},
		{"large number", "123", 123},
		{"mixed", "45abc", 45},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			result := extractMajorVersion(test.version)
			ts.Equal(test.expected, result, "Expected %d for %s", test.expected, test.version)
		})
	}
}

// TestDeps_Clean tests the Clean function
func (ts *DepsTestSuite) TestDeps_Clean() {
	ts.env.Builder.ExpectGoCommand("clean", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Clean()
		},
	)

	ts.Require().NoError(err)
}

// TestDeps_Clean_Error tests Clean function with error
func (ts *DepsTestSuite) TestDeps_Clean_Error() {
	expectedError := require.New(ts.T())
	ts.env.Builder.ExpectFailure("clean failed")

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Clean()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "failed to clean module cache")
}

// TestDeps_Graph tests the Graph function
func (ts *DepsTestSuite) TestDeps_Graph() {
	ts.env.Builder.ExpectGoCommand("mod", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Graph()
		},
	)

	ts.Require().NoError(err)
}

// TestDeps_Why tests the Why function
func (ts *DepsTestSuite) TestDeps_Why() {
	ts.env.Builder.ExpectGoCommand("mod", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Why("github.com/pkg/errors")
		},
	)

	ts.Require().NoError(err)
}

// TestDeps_Why_EmptyDep tests Why function with empty dependency
func (ts *DepsTestSuite) TestDeps_Why_EmptyDep() {
	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Why("")
		},
	)

	ts.Require().Error(err)
	ts.Require().Contains(err.Error(), "dependency name required")
}

// TestDeps_Verify tests the Verify function
func (ts *DepsTestSuite) TestDeps_Verify() {
	ts.env.Builder.ExpectGoCommand("mod", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Verify()
		},
	)

	ts.Require().NoError(err)
}

// TestDeps_Verify_Error tests Verify function with error
func (ts *DepsTestSuite) TestDeps_Verify_Error() {
	expectedError := require.New(ts.T())
	ts.env.Builder.ExpectFailure("verify failed")

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Verify()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "dependency verification failed")
}

// TestDeps_List tests the List function
func (ts *DepsTestSuite) TestDeps_List() {
	ts.env.Runner.On("RunCmd", "go", []string{"list", "-m", "-f", "{{if not .Indirect}}{{.Path}} {{.Version}}{{end}}", "all"}).Return(nil)
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-f", "{{if .Indirect}}1{{end}}", "all"}).Return("1\n1\n1", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.List()
		},
	)

	ts.Require().NoError(err)
}

// TestDeps_Outdated tests the Outdated function
func (ts *DepsTestSuite) TestDeps_Outdated() {
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-u", "-f", "{{if and (not .Indirect) .Update}}{{.Path}} {{.Version}} → {{.Update.Version}}{{end}}", "all"}).Return("", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Outdated()
		},
	)

	ts.Require().NoError(err)
}

// TestDeps_Outdated_WithUpdates tests Outdated function with available updates
func (ts *DepsTestSuite) TestDeps_Outdated_WithUpdates() {
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-u", "-f", "{{if and (not .Indirect) .Update}}{{.Path}} {{.Version}} → {{.Update.Version}}{{end}}", "all"}).Return("github.com/pkg/errors v0.9.0 → v0.9.1\ngithub.com/stretchr/testify v1.8.0 → v1.9.0", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Outdated()
		},
	)

	ts.Require().NoError(err)
}

// TestDeps_Outdated_Error tests Outdated function with error
func (ts *DepsTestSuite) TestDeps_Outdated_Error() {
	expectedError := require.New(ts.T())
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-u", "-f", "{{if and (not .Indirect) .Update}}{{.Path}} {{.Version}} → {{.Update.Version}}{{end}}", "all"}).Return("", errOutdatedCheckFailed)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Outdated()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "failed to check outdated dependencies")
}

// TestDeps_Vendor tests the Vendor function
func (ts *DepsTestSuite) TestDeps_Vendor() {
	ts.env.Builder.ExpectGoCommand("mod", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Vendor()
		},
	)

	ts.Require().NoError(err)
}

// TestDeps_Vendor_Error tests Vendor function with error
func (ts *DepsTestSuite) TestDeps_Vendor_Error() {
	expectedError := require.New(ts.T())
	ts.env.Builder.ExpectFailure("vendor failed")

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Vendor()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "failed to vendor dependencies")
}

// TestDeps_Init tests the Init function
func (ts *DepsTestSuite) TestDeps_Init() {
	// Clean up and recreate environment without go.mod for this specific test
	ts.env.Cleanup()
	ts.env = testutil.NewTestEnvironment(ts.T())

	ts.env.Builder.ExpectGoCommand("mod", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Init("github.com/example/project")
		},
	)

	ts.Require().NoError(err)
}

// TestDeps_Init_EmptyModule tests Init function with empty module name
func (ts *DepsTestSuite) TestDeps_Init_EmptyModule() {
	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Init("")
		},
	)

	ts.Require().Error(err)
	ts.Require().Contains(err.Error(), "module name required")
}

// TestDeps_Init_GoModExists tests Init function when go.mod already exists
func (ts *DepsTestSuite) TestDeps_Init_GoModExists() {
	// go.mod is created in SetupTest, so it already exists
	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Init("github.com/example/project")
		},
	)

	ts.Require().Error(err)
	ts.Require().Contains(err.Error(), "go.mod already exists")
}

// TestDeps_Init_Error tests Init function with command error
func (ts *DepsTestSuite) TestDeps_Init_Error() {
	// Remove go.mod for this test
	ts.env.Cleanup()
	ts.env = testutil.NewTestEnvironment(ts.T())

	expectedError := require.New(ts.T())
	ts.env.Builder.ExpectFailure("init failed")

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Init("github.com/example/project")
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "failed to initialize module")
}

// TestDeps_Licenses tests the Licenses function
func (ts *DepsTestSuite) TestDeps_Licenses() {
	ts.env.Runner.On("RunCmd", "echo", []string{"Checking dependency licenses"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Licenses()
		},
	)

	ts.Require().NoError(err)
}

// TestDeps_Check tests the Check function
func (ts *DepsTestSuite) TestDeps_Check() {
	ts.env.Builder.ExpectGoCommand("list", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Check()
		},
	)

	ts.Require().NoError(err)
}

// TestCompareVersions tests the compareVersions function
func (ts *DepsTestSuite) TestCompareVersions() {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		// Equal versions
		{"same version", "v1.2.3", "v1.2.3", 0},
		{"same without v prefix", "1.2.3", "1.2.3", 0},
		{"same mixed prefix", "v1.2.3", "1.2.3", 0},

		// v1 > v2 (should return 1)
		{"major version greater", "v2.0.0", "v1.0.0", 1},
		{"minor version greater", "v1.2.0", "v1.1.0", 1},
		{"patch version greater", "v1.0.2", "v1.0.1", 1},
		{"stable vs prerelease", "v1.2.3", "v1.2.3-alpha", 1},

		// v1 < v2 (should return -1)
		{"major version less", "v1.0.0", "v2.0.0", -1},
		{"minor version less", "v1.1.0", "v1.2.0", -1},
		{"patch version less", "v1.0.1", "v1.0.2", -1},
		{"prerelease vs stable", "v1.2.3-alpha", "v1.2.3", -1},

		// Pre-release comparisons
		{"prerelease lexicographic", "v1.2.3-beta", "v1.2.3-alpha", 1},
		{"prerelease with timestamp", "v1.1.16-0.20250601040535-ed473510065e", "v1.1.15", 1},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			result := compareVersions(test.v1, test.v2)
			ts.Equal(test.expected, result, "Expected %d for %s vs %s", test.expected, test.v1, test.v2)
		})
	}
}

// TestIsVersionNewer tests the isVersionNewer function
func (ts *DepsTestSuite) TestIsVersionNewer() {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected bool
	}{
		{"newer major", "v2.0.0", "v1.0.0", true},
		{"newer minor", "v1.2.0", "v1.1.0", true},
		{"newer patch", "v1.0.2", "v1.0.1", true},
		{"stable newer than prerelease", "v1.2.3", "v1.2.3-alpha", true},
		{"prerelease newer than stable (user case)", "v1.1.16-0.20250601040535-ed473510065e", "v1.1.15", true},
		{"older version", "v1.0.0", "v2.0.0", false},
		{"same version", "v1.2.3", "v1.2.3", false},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			result := isVersionNewer(test.v1, test.v2)
			ts.Equal(test.expected, result, "Expected %t for %s > %s", test.expected, test.v1, test.v2)
		})
	}
}

// TestParseVersion tests the parseVersion function
func (ts *DepsTestSuite) TestParseVersion() {
	tests := []struct {
		name               string
		version            string
		expectedNumbers    [3]int
		expectedPrerelease string
	}{
		{"simple version", "1.2.3", [3]int{1, 2, 3}, ""},
		{"with prerelease", "1.2.3-alpha", [3]int{1, 2, 3}, "alpha"},
		{"with timestamp prerelease", "1.1.16-0.20250601040535-ed473510065e", [3]int{1, 1, 16}, "0.20250601040535-ed473510065e"},
		{"only major", "2", [3]int{2, 0, 0}, ""},
		{"major minor", "1.5", [3]int{1, 5, 0}, ""},
		{"beta prerelease", "2.0.0-beta.1", [3]int{2, 0, 0}, "beta.1"},
	}

	for _, test := range tests {
		ts.Run(test.name, func() {
			result := parseVersion(test.version)
			ts.Equal(test.expectedNumbers, result.numbers, "Numbers mismatch for %s", test.version)
			ts.Equal(test.expectedPrerelease, result.prerelease, "Prerelease mismatch for %s", test.version)
		})
	}
}

// TestDeps_UpdateWithArgs_PreReleaseProtection tests that pre-release versions are not downgraded
func (ts *DepsTestSuite) TestDeps_UpdateWithArgs_PreReleaseProtection() {
	// Mock the list of direct dependencies command
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-f", "{{if not .Indirect}}{{.Path}}{{end}}", "all"}).Return("github.com/uber/athenadriver", nil)

	// Mock current version query - returns a pre-release version
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "github.com/uber/athenadriver"}).Return("github.com/uber/athenadriver v1.1.16-0.20250601040535-ed473510065e", nil)

	// Mock versions command to return stable versions only (as would happen in real life)
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-versions", "github.com/uber/athenadriver"}).Return("github.com/uber/athenadriver v1.1.13 v1.1.14 v1.1.15", nil)

	// Should only run tidy, NO update should happen because pre-release is newer than stable
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.UpdateWithArgs()
		},
	)

	ts.Require().NoError(err)

	// Verify no update command was called (the mock would fail if it was called)
	ts.env.Runner.AssertNotCalled(ts.T(), "RunCmd", "go", []string{"get", "-u", "github.com/uber/athenadriver@v1.1.15"})
}

// TestDeps_UpdateWithArgs_StableOnly tests that stable-only forces downgrade from pre-release
func (ts *DepsTestSuite) TestDeps_UpdateWithArgs_StableOnly() {
	// Mock the list of direct dependencies command
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-f", "{{if not .Indirect}}{{.Path}}{{end}}", "all"}).Return("github.com/uber/athenadriver", nil)

	// Mock current version query - returns a pre-release version
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "github.com/uber/athenadriver"}).Return("github.com/uber/athenadriver v1.1.16-0.20250601040535-ed473510065e", nil)

	// Mock versions command to return stable versions only
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-versions", "github.com/uber/athenadriver"}).Return("github.com/uber/athenadriver v1.1.13 v1.1.14 v1.1.15", nil)

	// Should downgrade to stable version when stable-only is enabled
	ts.env.Runner.On("RunCmd", "go", []string{"get", "-u", "github.com/uber/athenadriver@v1.1.15"}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.UpdateWithArgs("stable-only")
		},
	)

	ts.Require().NoError(err)
}

// TestDeps_UpdateWithArgs_PreReleaseToNewer tests normal update when newer stable version is available
func (ts *DepsTestSuite) TestDeps_UpdateWithArgs_PreReleaseToNewer() {
	// Mock the list of direct dependencies command
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-f", "{{if not .Indirect}}{{.Path}}{{end}}", "all"}).Return("github.com/example/lib", nil)

	// Mock current version query - returns a pre-release version
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "github.com/example/lib"}).Return("github.com/example/lib v1.1.16-0.20250601040535-ed473510065e", nil)

	// Mock versions command to return newer stable version
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-versions", "github.com/example/lib"}).Return("github.com/example/lib v1.1.13 v1.1.14 v1.1.15 v1.2.0", nil)

	// Should update to newer stable version (v1.2.0 is newer than the pre-release v1.1.16)
	ts.env.Runner.On("RunCmd", "go", []string{"get", "-u", "github.com/example/lib@v1.2.0"}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.UpdateWithArgs()
		},
	)

	ts.Require().NoError(err)
}

// TestDeps_Audit tests the Audit function
func (ts *DepsTestSuite) TestDeps_Audit() {
	// Mock config loading
	cfg := Config{
		Tools: ToolsConfig{
			GoVulnCheck: "latest",
		},
	}
	TestSetConfig(&cfg)

	// Mock command exists check (govulncheck not installed)
	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@latest"}).Return(nil)
	ts.env.Runner.On("RunCmd", "govulncheck", []string{"-show", "verbose", "./..."}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Audit()
		},
	)

	ts.Require().NoError(err)
}

// TestDeps_Audit_Error tests Audit function with govulncheck failure
func (ts *DepsTestSuite) TestDeps_Audit_Error() {
	// Mock config loading
	cfg := Config{
		Tools: ToolsConfig{
			GoVulnCheck: "latest",
		},
	}
	TestSetConfig(&cfg)

	expectedError := require.New(ts.T())

	// Mock govulncheck failure
	ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/vuln/cmd/govulncheck@latest"}).Return(nil)
	ts.env.Runner.On("RunCmd", "govulncheck", []string{"-show", "verbose", "./..."}).Return(errors.New("vulnerability check failed"))

	err := ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
		},
		func() interface{} { return GetRunner() },
		func() error {
			return ts.deps.Audit()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "vulnerability check failed")
}

// TestDepsTestSuite runs the test suite
func TestDepsTestSuite(t *testing.T) {
	suite.Run(t, new(DepsTestSuite))
}
