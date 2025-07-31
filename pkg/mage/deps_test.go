package mage

import (
	"errors"
	"testing"

	"github.com/mrz1836/go-mage/pkg/mage/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
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
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-f", "{{if not .Indirect}}{{.Path}}{{end}}", "all"}).Return("", errors.New("list failed"))

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
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-versions", "github.com/stretchr/testify"}).Return("github.com/stretchr/testify v1.8.0 v1.9.0", nil)
	ts.env.Runner.On("RunCmd", "go", []string{"get", "-u", "github.com/stretchr/testify@v1.9.0"}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(errors.New("tidy failed"))

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
	expectedError.Contains(err.Error(), "failed to tidy after updates")
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
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-u", "-f", "{{if and (not .Indirect) .Update}}{{.Path}} {{.Version}} -> {{.Update.Version}}{{end}}", "all"}).Return("", nil)

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
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-u", "-f", "{{if and (not .Indirect) .Update}}{{.Path}} {{.Version}} -> {{.Update.Version}}{{end}}", "all"}).Return("github.com/pkg/errors v0.9.0 -> v0.9.1\ngithub.com/stretchr/testify v1.8.0 -> v1.9.0", nil)

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
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-u", "-f", "{{if and (not .Indirect) .Update}}{{.Path}} {{.Version}} -> {{.Update.Version}}{{end}}", "all"}).Return("", errors.New("outdated check failed"))

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

// TestDeps_Audit tests the Audit function
func (ts *DepsTestSuite) TestDeps_Audit() {
	ts.env.Builder.ExpectGoCommand("list", nil)

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

// TestDepsTestSuite runs the test suite
func TestDepsTestSuite(t *testing.T) {
	suite.Run(t, new(DepsTestSuite))
}
