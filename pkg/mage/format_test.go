package mage

import (
	"strings"
	"testing"

	"github.com/mrz1836/go-mage/pkg/mage/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// FormatTestSuite defines the test suite for Format functions
type FormatTestSuite struct {
	suite.Suite
	env    *testutil.TestEnvironment
	format Format
}

// SetupTest runs before each test
func (ts *FormatTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.format = Format{}
}

// TearDownTest runs after each test
func (ts *FormatTestSuite) TearDownTest() {
	ts.env.Cleanup()
}

// TestFormatGofmt tests the Gofmt method
func (ts *FormatTestSuite) TestFormatGofmt() {
	ts.Run("no files need formatting", func() {
		// Mock gofmt list check with no output (all files formatted)
		ts.env.Runner.On("RunCmdOutput", "gofmt", []string{"-l", "."}).Return("", nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Gofmt()
			},
		)

		require.NoError(ts.T(), err)
	})

	ts.Run("files need formatting", func() {
		// Mock gofmt list check with files that need formatting
		ts.env.Runner.On("RunCmdOutput", "gofmt", []string{"-l", "."}).Return("file1.go\nfile2.go", nil)
		ts.env.Runner.On("RunCmd", "gofmt", []string{"-w", "."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Gofmt()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestFormatCheck tests the Check method
func (ts *FormatTestSuite) TestFormatCheck() {
	ts.Run("all files properly formatted", func() {
		// Mock all format checks passing
		ts.env.Runner.On("RunCmdOutput", "gofmt", []string{"-l", "."}).Return("", nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Check()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestFormatInstallTools tests the InstallTools method
func (ts *FormatTestSuite) TestFormatInstallTools() {
	ts.Run("successful tool installation", func() {
		// Mock successful installation of both tools
		ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/tools/cmd/goimports@latest"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.InstallTools()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestFormatAll tests the All method
func (ts *FormatTestSuite) TestFormatAll() {
	ts.Run("successful all formatting", func() {
		// Mock successful echo command
		ts.env.Runner.On("RunCmd", "echo", []string{"Formatting all files"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.All()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestFormatGo tests the Go method
func (ts *FormatTestSuite) TestFormatGo() {
	ts.Run("successful Go formatting", func() {
		// Mock successful gofmt command
		ts.env.Runner.On("RunCmd", "gofmt", []string{"-w", "."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Go()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestFormatYAML tests the YAML method
func (ts *FormatTestSuite) TestFormatYAML() {
	ts.Run("successful YAML formatting", func() {
		// Mock successful echo command
		ts.env.Runner.On("RunCmd", "echo", []string{"Formatting YAML files"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.YAML()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestFormatYaml tests the Yaml method (alias)
func (ts *FormatTestSuite) TestFormatYaml() {
	ts.Run("successful Yaml formatting (alias)", func() {
		// Mock successful echo command (called through YAML method)
		ts.env.Runner.On("RunCmd", "echo", []string{"Formatting YAML files"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Yaml()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestFormatJSON tests the JSON method
func (ts *FormatTestSuite) TestFormatJSON() {
	ts.Run("successful JSON formatting", func() {
		// Mock successful echo command
		ts.env.Runner.On("RunCmd", "echo", []string{"Formatting JSON files"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.JSON()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestFormatJson tests the Json method (alias)
func (ts *FormatTestSuite) TestFormatJson() {
	ts.Run("successful Json formatting (alias)", func() {
		// Mock successful echo command (called through JSON method)
		ts.env.Runner.On("RunCmd", "echo", []string{"Formatting JSON files"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Json()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestFormatMarkdown tests the Markdown method
func (ts *FormatTestSuite) TestFormatMarkdown() {
	ts.Run("successful Markdown formatting", func() {
		// Mock successful echo command
		ts.env.Runner.On("RunCmd", "echo", []string{"Formatting Markdown files"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Markdown()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestFormatSQL tests the SQL method
func (ts *FormatTestSuite) TestFormatSQL() {
	ts.Run("successful SQL formatting", func() {
		// Mock successful echo command
		ts.env.Runner.On("RunCmd", "echo", []string{"Formatting SQL files"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.SQL()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestFormatDockerfile tests the Dockerfile method
func (ts *FormatTestSuite) TestFormatDockerfile() {
	ts.Run("successful Dockerfile formatting", func() {
		// Mock successful echo command
		ts.env.Runner.On("RunCmd", "echo", []string{"Formatting Dockerfile"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Dockerfile()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestFormatShell tests the Shell method
func (ts *FormatTestSuite) TestFormatShell() {
	ts.Run("successful Shell formatting", func() {
		// Mock successful echo command
		ts.env.Runner.On("RunCmd", "echo", []string{"Formatting shell scripts"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Shell()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestFormatFix tests the Fix method
func (ts *FormatTestSuite) TestFormatFix() {
	ts.Run("successful formatting fix", func() {
		// Mock successful echo command
		ts.env.Runner.On("RunCmd", "echo", []string{"Fixing formatting issues"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Fix()
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestFormatHelperFunctions tests the helper functions
func (ts *FormatTestSuite) TestFormatHelperFunctions() {
	ts.Run("findGoFiles function", func() {
		// Create some test Go files in temp directory
		ts.env.CreateFile("main.go", "package main")
		ts.env.CreateFile("util.go", "package util")
		ts.env.CreateFile("test.pb.go", "// protobuf generated")
		ts.env.CreateFile("vendor/dep.go", "package dep")

		files, err := findGoFiles()
		require.NoError(ts.T(), err)
		require.GreaterOrEqual(ts.T(), len(files), 2) // At least main.go and util.go

		// Check that .pb.go files are excluded
		for _, file := range files {
			require.False(ts.T(), strings.Contains(file, ".pb.go"))
		}
	})

	ts.Run("filterEmpty function", func() {
		input := []string{"", "file1.go", "", "file2.go", ""}
		result := filterEmpty(input)
		expected := []string{"file1.go", "file2.go"}
		require.Equal(ts.T(), expected, result)
	})
}

// TestFormatTestSuite runs the test suite
func TestFormatTestSuite(t *testing.T) {
	suite.Run(t, new(FormatTestSuite))
}
