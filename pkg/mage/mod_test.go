package mage

import (
	"errors"
	"os"
	"testing"

	"github.com/mrz1836/go-mage/pkg/mage/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ModTestSuite defines the test suite for mod functions
type ModTestSuite struct {
	suite.Suite
	env *testutil.TestEnvironment
	mod Mod
}

// SetupTest runs before each test
func (ts *ModTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.mod = Mod{}
}

// TearDownTest runs after each test
func (ts *ModTestSuite) TearDownTest() {
	ts.env.Cleanup()
}

// TestMod_Download tests the Download function
func (ts *ModTestSuite) TestMod_Download() {
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "download"}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "verify"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Download()
		},
	)

	require.NoError(ts.T(), err)
}

// TestMod_Download_DownloadError tests Download function with download error
func (ts *ModTestSuite) TestMod_Download_DownloadError() {
	expectedError := require.New(ts.T())
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "download"}).Return(errors.New("download failed"))

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Download()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "failed to download dependencies")
}

// TestMod_Download_VerifyFails tests Download function with verify warning
func (ts *ModTestSuite) TestMod_Download_VerifyFails() {
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "download"}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "verify"}).Return(errors.New("verify failed"))

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Download()
		},
	)

	require.NoError(ts.T(), err)
}

// TestMod_Tidy tests the Tidy function
func (ts *ModTestSuite) TestMod_Tidy() {
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(nil)
	ts.env.Runner.On("RunCmdOutput", "git", []string{"status", "--porcelain", "go.mod", "go.sum"}).Return("", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Tidy()
		},
	)

	require.NoError(ts.T(), err)
}

// TestMod_Tidy_WithChanges tests Tidy function when files are changed
func (ts *ModTestSuite) TestMod_Tidy_WithChanges() {
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(nil)
	ts.env.Runner.On("RunCmdOutput", "git", []string{"status", "--porcelain", "go.mod", "go.sum"}).Return(" M go.mod\n M go.sum", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Tidy()
		},
	)

	require.NoError(ts.T(), err)
}

// TestMod_Tidy_Error tests Tidy function with error
func (ts *ModTestSuite) TestMod_Tidy_Error() {
	expectedError := require.New(ts.T())
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(errors.New("tidy failed"))

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Tidy()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "go mod tidy failed")
}

// TestMod_Update tests the Update function
func (ts *ModTestSuite) TestMod_Update() {
	listOutput := "github.com/stretchr/testify v1.8.0 [v1.9.0]\ngithub.com/pkg/errors v0.9.0"
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-u", "-m", "all"}).Return(listOutput, nil)
	ts.env.Runner.On("RunCmd", "go", []string{"get", "-u", "./..."}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Update()
		},
	)

	require.NoError(ts.T(), err)
}

// TestMod_Update_NoUpdates tests Update function when no updates are available
func (ts *ModTestSuite) TestMod_Update_NoUpdates() {
	listOutput := "github.com/stretchr/testify v1.9.0\ngithub.com/pkg/errors v0.9.1"
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-u", "-m", "all"}).Return(listOutput, nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Update()
		},
	)

	require.NoError(ts.T(), err)
}

// TestMod_Update_ListError tests Update function with list error
func (ts *ModTestSuite) TestMod_Update_ListError() {
	expectedError := require.New(ts.T())
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-u", "-m", "all"}).Return("", errors.New("list failed"))

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Update()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "failed to list dependencies")
}

// TestMod_Update_GetError tests Update function with get error
func (ts *ModTestSuite) TestMod_Update_GetError() {
	expectedError := require.New(ts.T())
	listOutput := "github.com/stretchr/testify v1.8.0 [v1.9.0]"
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-u", "-m", "all"}).Return(listOutput, nil)
	ts.env.Runner.On("RunCmd", "go", []string{"get", "-u", "./..."}).Return(errors.New("get failed"))

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Update()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "failed to update dependencies")
}

// TestMod_Update_TidyError tests Update function with tidy error
func (ts *ModTestSuite) TestMod_Update_TidyError() {
	expectedError := require.New(ts.T())
	listOutput := "github.com/stretchr/testify v1.8.0 [v1.9.0]"
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-u", "-m", "all"}).Return(listOutput, nil)
	ts.env.Runner.On("RunCmd", "go", []string{"get", "-u", "./..."}).Return(nil)
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(errors.New("tidy failed"))

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Update()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "go mod tidy failed")
}

// TestMod_Clean tests the Clean function with FORCE environment variable
func (ts *ModTestSuite) TestMod_Clean() {
	require.NoError(ts.T(), os.Setenv("FORCE", "true"))
	defer func() { _ = os.Unsetenv("FORCE") }()

	ts.env.Runner.On("RunCmd", "go", []string{"clean", "-modcache"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Clean()
		},
	)

	require.NoError(ts.T(), err)
}

// TestMod_Clean_NoForce tests Clean function without FORCE environment variable
func (ts *ModTestSuite) TestMod_Clean_NoForce() {
	_ = os.Unsetenv("FORCE")

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Clean()
		},
	)

	require.Error(ts.T(), err)
	require.Contains(ts.T(), err.Error(), "operation canceled")
}

// TestMod_Clean_Error tests Clean function with command error
func (ts *ModTestSuite) TestMod_Clean_Error() {
	expectedError := require.New(ts.T())
	require.NoError(ts.T(), os.Setenv("FORCE", "true"))
	defer func() { _ = os.Unsetenv("FORCE") }()

	ts.env.Runner.On("RunCmd", "go", []string{"clean", "-modcache"}).Return(errors.New("clean failed"))

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Clean()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "failed to clean module cache")
}

// TestMod_Graph tests the Graph function
func (ts *ModTestSuite) TestMod_Graph() {
	graphOutput := "test/module github.com/stretchr/testify@v1.9.0\ntest/module github.com/pkg/errors@v0.9.1"
	ts.env.Runner.On("RunCmdOutput", "go", []string{"mod", "graph"}).Return(graphOutput, nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Graph()
		},
	)

	require.NoError(ts.T(), err)
}

// TestMod_Graph_WithGraphFile tests Graph function with GRAPH_FILE environment variable
func (ts *ModTestSuite) TestMod_Graph_WithGraphFile() {
	graphOutput := "test/module github.com/stretchr/testify@v1.9.0"
	require.NoError(ts.T(), os.Setenv("GRAPH_FILE", "test-graph.txt"))
	defer func() { _ = os.Unsetenv("GRAPH_FILE") }()

	ts.env.Runner.On("RunCmdOutput", "go", []string{"mod", "graph"}).Return(graphOutput, nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Graph()
		},
	)

	require.NoError(ts.T(), err)
}

// TestMod_Graph_Error tests Graph function with error
func (ts *ModTestSuite) TestMod_Graph_Error() {
	expectedError := require.New(ts.T())
	ts.env.Runner.On("RunCmdOutput", "go", []string{"mod", "graph"}).Return("", errors.New("graph failed"))

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Graph()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "failed to generate dependency graph")
}

// TestMod_Why tests the Why function with MODULE environment variable
func (ts *ModTestSuite) TestMod_Why() {
	require.NoError(ts.T(), os.Setenv("MODULE", "github.com/pkg/errors"))
	defer func() { _ = os.Unsetenv("MODULE") }()

	whyOutput := "# github.com/pkg/errors\ntest/module\ngithub.com/pkg/errors"
	ts.env.Runner.On("RunCmdOutput", "go", []string{"mod", "why", "github.com/pkg/errors"}).Return(whyOutput, nil)
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-f", "{{.Require}}", "all"}).Return("github.com/pkg/errors", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Why()
		},
	)

	require.NoError(ts.T(), err)
}

// TestMod_Why_NoModule tests Why function without MODULE environment variable
func (ts *ModTestSuite) TestMod_Why_NoModule() {
	_ = os.Unsetenv("MODULE")

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Why()
		},
	)

	require.Error(ts.T(), err)
	require.Contains(ts.T(), err.Error(), "MODULE environment variable is required")
}

// TestMod_Why_IndirectDependency tests Why function for indirect dependency
func (ts *ModTestSuite) TestMod_Why_IndirectDependency() {
	require.NoError(ts.T(), os.Setenv("MODULE", "github.com/indirect/dep"))
	defer func() { _ = os.Unsetenv("MODULE") }()

	whyOutput := "# github.com/indirect/dep\ntest/module\ngithub.com/some/other\ngithub.com/indirect/dep"
	ts.env.Runner.On("RunCmdOutput", "go", []string{"mod", "why", "github.com/indirect/dep"}).Return(whyOutput, nil)
	ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-m", "-f", "{{.Require}}", "all"}).Return("github.com/pkg/errors", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Why()
		},
	)

	require.NoError(ts.T(), err)
}

// TestMod_Why_Error tests Why function with error
func (ts *ModTestSuite) TestMod_Why_Error() {
	expectedError := require.New(ts.T())
	require.NoError(ts.T(), os.Setenv("MODULE", "github.com/pkg/errors"))
	defer func() { _ = os.Unsetenv("MODULE") }()

	ts.env.Runner.On("RunCmdOutput", "go", []string{"mod", "why", "github.com/pkg/errors"}).Return("", errors.New("why failed"))

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Why()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "failed to analyze module")
}

// TestMod_Vendor tests the Vendor function
func (ts *ModTestSuite) TestMod_Vendor() {
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "vendor"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Vendor()
		},
	)

	require.NoError(ts.T(), err)
}

// TestMod_Vendor_Error tests Vendor function with error
func (ts *ModTestSuite) TestMod_Vendor_Error() {
	expectedError := require.New(ts.T())
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "vendor"}).Return(errors.New("vendor failed"))

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Vendor()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "vendoring failed")
}

// TestMod_Init tests the Init function with MODULE environment variable
func (ts *ModTestSuite) TestMod_Init() {
	// Clean up and recreate environment without go.mod for this specific test
	ts.env.Cleanup()
	ts.env = testutil.NewTestEnvironment(ts.T())

	require.NoError(ts.T(), os.Setenv("MODULE", "github.com/example/project"))
	defer func() { _ = os.Unsetenv("MODULE") }()

	ts.env.Runner.On("RunCmd", "go", []string{"mod", "init", "github.com/example/project"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Init()
		},
	)

	require.NoError(ts.T(), err)
}

// TestMod_Init_WithGitRemote tests Init function inferring module name from git remote
func (ts *ModTestSuite) TestMod_Init_WithGitRemote() {
	// Clean up and recreate environment without go.mod for this specific test
	ts.env.Cleanup()
	ts.env = testutil.NewTestEnvironment(ts.T())

	_ = os.Unsetenv("MODULE")

	ts.env.Runner.On("RunCmdOutput", "git", []string{"remote", "get-url", "origin"}).Return("https://github.com/example/project.git", nil)
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "init", "github.com/example/project"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Init()
		},
	)

	require.NoError(ts.T(), err)
}

// TestMod_Init_GoModExists tests Init function when go.mod already exists
func (ts *ModTestSuite) TestMod_Init_GoModExists() {
	require.NoError(ts.T(), os.Setenv("MODULE", "github.com/example/project"))
	defer func() { _ = os.Unsetenv("MODULE") }()

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Init()
		},
	)

	require.Error(ts.T(), err)
	require.Contains(ts.T(), err.Error(), "go.mod already exists")
}

// TestMod_Init_NoModule tests Init function without MODULE or git remote
func (ts *ModTestSuite) TestMod_Init_NoModule() {
	// Clean up and recreate environment without go.mod for this specific test
	ts.env.Cleanup()
	ts.env = testutil.NewTestEnvironment(ts.T())

	_ = os.Unsetenv("MODULE")
	ts.env.Runner.On("RunCmdOutput", "git", []string{"remote", "get-url", "origin"}).Return("", errors.New("no remote"))

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Init()
		},
	)

	require.Error(ts.T(), err)
	require.Contains(ts.T(), err.Error(), "MODULE environment variable is required")
}

// TestMod_Init_InitError tests Init function with init command error
func (ts *ModTestSuite) TestMod_Init_InitError() {
	// Clean up and recreate environment without go.mod for this specific test
	ts.env.Cleanup()
	ts.env = testutil.NewTestEnvironment(ts.T())

	expectedError := require.New(ts.T())
	require.NoError(ts.T(), os.Setenv("MODULE", "github.com/example/project"))
	defer func() { _ = os.Unsetenv("MODULE") }()

	ts.env.Runner.On("RunCmd", "go", []string{"mod", "init", "github.com/example/project"}).Return(errors.New("init failed"))

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Init()
		},
	)

	expectedError.Error(err)
	expectedError.Contains(err.Error(), "module initialization failed")
}

// TestMod_Verify tests the Verify function
func (ts *ModTestSuite) TestMod_Verify() {
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "verify"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Verify()
		},
	)

	require.NoError(ts.T(), err)
}

// TestMod_Edit tests the Edit function
func (ts *ModTestSuite) TestMod_Edit() {
	ts.env.Runner.On("RunCmd", "go", []string{"mod", "edit", "-require", "github.com/pkg/errors@v0.9.1"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Edit("-require", "github.com/pkg/errors@v0.9.1")
		},
	)

	require.NoError(ts.T(), err)
}

// TestMod_Get tests the Get function
func (ts *ModTestSuite) TestMod_Get() {
	ts.env.Runner.On("RunCmd", "go", []string{"get", "github.com/pkg/errors@v0.9.1", "github.com/stretchr/testify@v1.9.0"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.Get("github.com/pkg/errors@v0.9.1", "github.com/stretchr/testify@v1.9.0")
		},
	)

	require.NoError(ts.T(), err)
}

// TestMod_List tests the List function
func (ts *ModTestSuite) TestMod_List() {
	ts.env.Runner.On("RunCmd", "go", []string{"list", "-m", "all"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.List()
		},
	)

	require.NoError(ts.T(), err)
}

// TestMod_List_WithPattern tests the List function with patterns
func (ts *ModTestSuite) TestMod_List_WithPattern() {
	ts.env.Runner.On("RunCmd", "go", []string{"list", "-m", "github.com/pkg/*"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.mod.List("github.com/pkg/*")
		},
	)

	require.NoError(ts.T(), err)
}

// Helper function tests

// TestGitURLToModulePath tests the gitURLToModulePath helper function
func (ts *ModTestSuite) TestGitURLToModulePath() {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTTPS URL",
			input:    "https://github.com/user/repo.git",
			expected: "github.com/user/repo",
		},
		{
			name:     "SSH URL",
			input:    "git@github.com:user/repo.git",
			expected: "github.com/user/repo",
		},
		{
			name:     "SSH with ssh:// prefix",
			input:    "ssh://git@github.com/user/repo.git",
			expected: "github.com/user/repo",
		},
		{
			name:     "HTTP URL",
			input:    "http://github.com/user/repo.git",
			expected: "github.com/user/repo",
		},
		{
			name:     "URL without .git suffix",
			input:    "https://github.com/user/repo",
			expected: "github.com/user/repo",
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			result := gitURLToModulePath(tc.input)
			require.Equal(ts.T(), tc.expected, result)
		})
	}
}

// TestGetModCache tests the getModCache helper function
func (ts *ModTestSuite) TestGetModCache() {
	originalGomodcache := os.Getenv("GOMODCACHE")
	originalGopath := os.Getenv("GOPATH")
	defer func() {
		if originalGomodcache != "" {
			os.Setenv("GOMODCACHE", originalGomodcache)
		} else {
			_ = os.Unsetenv("GOMODCACHE")
		}
		if originalGopath != "" {
			os.Setenv("GOPATH", originalGopath)
		} else {
			_ = os.Unsetenv("GOPATH")
		}
	}()

	ts.Run("with GOMODCACHE", func() {
		os.Setenv("GOMODCACHE", "/custom/modcache")
		_ = os.Unsetenv("GOPATH")

		result := getModCache()
		require.Equal(ts.T(), "/custom/modcache", result)
	})

	ts.Run("with GOPATH", func() {
		_ = os.Unsetenv("GOMODCACHE")
		os.Setenv("GOPATH", "/custom/gopath")

		result := getModCache()
		require.Equal(ts.T(), "/custom/gopath/pkg/mod", result)
	})

	ts.Run("default location", func() {
		_ = os.Unsetenv("GOMODCACHE")
		_ = os.Unsetenv("GOPATH")

		result := getModCache()
		require.Contains(ts.T(), result, "go/pkg/mod")
	})
}

// TestModTestSuite runs the test suite
func TestModTestSuite(t *testing.T) {
	suite.Run(t, new(ModTestSuite))
}
