//go:build integration
// +build integration

package mage

import (
	"fmt"
	"os"
	"testing"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
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

		ts.Require().NoError(err)
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

		ts.Require().NoError(err)
	})
}

// TestFormatCheck tests the Check method
func (ts *FormatTestSuite) TestFormatCheck() {
	ts.Run("all files properly formatted", func() {
		// Mock all format checks passing
		ts.env.Runner.On("RunCmdOutput", "gofmt", []string{"-l", "."}).Return("", nil)
		ts.env.Runner.On("RunCmdOutput", "gofumpt", []string{"-l", "-extra", "."}).Return("", nil)
		ts.env.Runner.On("RunCmdOutput", "goimports", []string{"-l", "."}).Return("", nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Check()
			},
		)

		ts.Require().NoError(err)
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

		ts.Require().NoError(err)
	})
}

// TestFormatAll tests the All method
func (ts *FormatTestSuite) TestFormatAll() {
	ts.Run("successful all formatting", func() {
		// Mock the Default() method commands: Gofmt, Fumpt, Imports
		// Gofmt commands
		ts.env.Runner.On("RunCmdOutput", "gofmt", []string{"-l", "."}).Return("", nil)

		// Fumpt commands (ensureGofumpt + run gofumpt)
		ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "gofumpt", []string{"-w", "-extra", "."}).Return(nil)

		// Imports commands (ensureGoimports + run goimports)
		ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/tools/cmd/goimports@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "goimports", []string{"-w", "."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.All()
			},
		)

		ts.Require().NoError(err)
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

		ts.Require().NoError(err)
	})
}

// TestFormatYAML tests the YAML method
func (ts *FormatTestSuite) TestFormatYAML() {
	ts.Run("successful YAML formatting", func() {
		// Mock finding YAML files
		ts.env.Runner.On("RunCmdOutput", "find", []string{".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*"}).Return("config.yml\ndata.yaml", nil)
		// Mock prettier formatting
		ts.env.Runner.On("RunCmd", "prettier", []string{"--write", "**/*.{yml,yaml}"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.YAML()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestFormatYaml tests the Yaml method (alias)
func (ts *FormatTestSuite) TestFormatYaml() {
	ts.Run("successful Yaml formatting (alias)", func() {
		// Mock finding YAML files (called through YAML method)
		ts.env.Runner.On("RunCmdOutput", "find", []string{".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*"}).Return("config.yml\ndata.yaml", nil)
		// Mock prettier formatting
		ts.env.Runner.On("RunCmd", "prettier", []string{"--write", "**/*.{yml,yaml}"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Yaml()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestFormatJSON tests the JSON method
func (ts *FormatTestSuite) TestFormatJSON() {
	ts.Run("successful JSON formatting", func() {
		// Mock finding JSON files
		ts.env.Runner.On("RunCmdOutput", "find", []string{".", "-name", "*.json", "-not", "-path", "./vendor/*"}).Return("package.json\nconfig.json", nil)
		// Mock python3 formatting for each file (fallback when prettier not available)
		ts.env.Runner.On("RunCmd", "python3", []string{"-m", "json.tool", "package.json", "package.json.tmp"}).Return(nil)
		ts.env.Runner.On("RunCmd", "python3", []string{"-m", "json.tool", "config.json", "config.json.tmp"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.JSON()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestFormatJson tests the Json method (alias)
func (ts *FormatTestSuite) TestFormatJson() {
	ts.Run("successful Json formatting (alias)", func() {
		// Mock finding JSON files (called through JSON method)
		ts.env.Runner.On("RunCmdOutput", "find", []string{".", "-name", "*.json", "-not", "-path", "./vendor/*"}).Return("package.json\nconfig.json", nil)
		// Mock python3 formatting for each file (fallback when prettier not available)
		ts.env.Runner.On("RunCmd", "python3", []string{"-m", "json.tool", "package.json", "package.json.tmp"}).Return(nil)
		ts.env.Runner.On("RunCmd", "python3", []string{"-m", "json.tool", "config.json", "config.json.tmp"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.JSON()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestFormatMarkdown tests the Markdown method
func (ts *FormatTestSuite) TestFormatMarkdown() {
	ts.Run("successful Markdown formatting", func() {
		// Mock finding Markdown files
		ts.env.Runner.On("RunCmdOutput", "find", []string{".", "-name", "*.md", "-not", "-path", "./vendor/*"}).Return("README.md\nDOCS.md", nil)
		// Mock prettier formatting
		ts.env.Runner.On("RunCmd", "prettier", []string{"--write", "**/*.md"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Markdown()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestFormatSQL tests the SQL method
func (ts *FormatTestSuite) TestFormatSQL() {
	ts.Run("successful SQL formatting", func() {
		// Mock finding SQL files
		ts.env.Runner.On("RunCmdOutput", "find", []string{".", "-name", "*.sql"}).Return("schema.sql\nqueries.sql", nil)
		// Mock sqlfluff formatting
		ts.env.Runner.On("RunCmd", "sqlfluff", []string{"format", "."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.SQL()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestFormatDockerfile tests the Dockerfile method
func (ts *FormatTestSuite) TestFormatDockerfile() {
	ts.Run("successful Dockerfile formatting", func() {
		// Mock dockerfile_lint command (if available)
		ts.env.Runner.On("RunCmd", "dockerfile_lint", []string{"Dockerfile"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Dockerfile()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestFormatShell tests the Shell method
func (ts *FormatTestSuite) TestFormatShell() {
	ts.Run("successful Shell formatting", func() {
		// Mock finding shell script files
		ts.env.Runner.On("RunCmdOutput", "find", []string{".", "-name", "*.sh", "-o", "-name", "*.bash"}).Return("script.sh\nbuild.bash", nil)
		// Mock shfmt formatting
		ts.env.Runner.On("RunCmd", "shfmt", []string{"-i", "2", "-w", "."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Shell()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestFormatFix tests the Fix method
func (ts *FormatTestSuite) TestFormatFix() {
	ts.Run("successful formatting fix", func() {
		// Mock all formatter commands called by Fix method
		// Gofmt commands
		ts.env.Runner.On("RunCmdOutput", "gofmt", []string{"-l", "."}).Return("", nil)

		// Fumpt commands (ensureGofumpt + run gofumpt)
		ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "gofumpt", []string{"-w", "-extra", "."}).Return(nil)

		// Imports commands (ensureGoimports + run goimports)
		ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/tools/cmd/goimports@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "goimports", []string{"-w", "."}).Return(nil)

		// JSON formatting commands
		ts.env.Runner.On("RunCmdOutput", "find", []string{".", "-name", "*.json", "-not", "-path", "./vendor/*"}).Return("package.json", nil)
		ts.env.Runner.On("RunCmd", "python3", []string{"-m", "json.tool", "package.json", "package.json.tmp"}).Return(nil)

		// YAML formatting commands
		ts.env.Runner.On("RunCmdOutput", "find", []string{".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*"}).Return("config.yml", nil)
		ts.env.Runner.On("RunCmd", "prettier", []string{"--write", "**/*.{yml,yaml}"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Fix()
			},
		)

		ts.Require().NoError(err)
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
		ts.Require().NoError(err)
		ts.Require().GreaterOrEqual(len(files), 2) // At least main.go and util.go

		// Check that .pb.go files are excluded
		for _, file := range files {
			ts.Require().NotContains(file, ".pb.go")
		}
	})

	ts.Run("filterEmpty function", func() {
		input := []string{"", "file1.go", "", "file2.go", ""}
		result := filterEmpty(input)
		expected := []string{"file1.go", "file2.go"}
		ts.Require().Equal(expected, result)
	})
}

// TestFormatGofmtErrorScenarios tests error handling in Gofmt method
func (ts *FormatTestSuite) TestFormatGofmtErrorScenarios() {
	ts.Run("gofmt check command fails", func() {
		ts.env.Runner.On("RunCmdOutput", "gofmt", []string{"-l", "."}).Return("", fmt.Errorf("gofmt command failed"))

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Gofmt()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "gofmt check failed")
	})

	ts.Run("gofmt format command fails", func() {
		ts.env.Runner.On("RunCmdOutput", "gofmt", []string{"-l", "."}).Return("file1.go", nil)
		ts.env.Runner.On("RunCmd", "gofmt", []string{"-w", "."}).Return(fmt.Errorf("formatting failed"))

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Gofmt()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "gofmt failed")
	})
}

// TestFormatFumptScenarios tests various Fumpt scenarios
func (ts *FormatTestSuite) TestFormatFumptScenarios() {
	ts.Run("gofumpt not installed, installation succeeds", func() {
		// Mock gofumpt installation and execution
		ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "gofumpt", []string{"-w", "-extra", "."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Fumpt()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("gofumpt execution fails", func() {
		// Mock successful installation but failed execution
		ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "gofumpt", []string{"-w", "-extra", "."}).Return(fmt.Errorf("gofumpt failed"))

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Fumpt()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "gofumpt failed")
	})
}

// TestFormatImportsScenarios tests various Imports scenarios
func (ts *FormatTestSuite) TestFormatImportsScenarios() {
	ts.Run("goimports not installed, installation succeeds", func() {
		ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/tools/cmd/goimports@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "goimports", []string{"-w", "."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Imports()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("goimports execution fails", func() {
		ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/tools/cmd/goimports@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "goimports", []string{"-w", "."}).Return(fmt.Errorf("goimports failed"))

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Imports()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "goimports failed")
	})
}

// TestFormatCheckWithFormatIssues tests Check method with various formatting issues
func (ts *FormatTestSuite) TestFormatCheckWithFormatIssues() {
	ts.Run("gofmt issues detected", func() {
		ts.env.Runner.On("RunCmdOutput", "gofmt", []string{"-l", "."}).Return("file1.go\nfile2.go", nil)
		ts.env.Runner.On("RunCmdOutput", "gofumpt", []string{"-l", "-extra", "."}).Return("", nil)
		ts.env.Runner.On("RunCmdOutput", "goimports", []string{"-l", "."}).Return("", nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Check()
			},
		)

		ts.Require().Error(err)
		ts.Require().Equal(ErrCodeNotFormatted, err)
	})

	ts.Run("gofumpt issues detected", func() {
		ts.env.Runner.On("RunCmdOutput", "gofmt", []string{"-l", "."}).Return("", nil)
		ts.env.Runner.On("RunCmdOutput", "gofumpt", []string{"-l", "-extra", "."}).Return("file1.go", nil)
		ts.env.Runner.On("RunCmdOutput", "goimports", []string{"-l", "."}).Return("", nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Check()
			},
		)

		ts.Require().Error(err)
		ts.Require().Equal(ErrCodeNotFormatted, err)
	})

	ts.Run("goimports issues detected", func() {
		ts.env.Runner.On("RunCmdOutput", "gofmt", []string{"-l", "."}).Return("", nil)
		ts.env.Runner.On("RunCmdOutput", "gofumpt", []string{"-l", "-extra", "."}).Return("", nil)
		ts.env.Runner.On("RunCmdOutput", "goimports", []string{"-l", "."}).Return("file1.go", nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Check()
			},
		)

		ts.Require().Error(err)
		ts.Require().Equal(ErrCodeNotFormatted, err)
	})

	ts.Run("multiple tools have issues", func() {
		ts.env.Runner.On("RunCmdOutput", "gofmt", []string{"-l", "."}).Return("file1.go", nil)
		ts.env.Runner.On("RunCmdOutput", "gofumpt", []string{"-l", "-extra", "."}).Return("file2.go", nil)
		ts.env.Runner.On("RunCmdOutput", "goimports", []string{"-l", "."}).Return("file3.go", nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Check()
			},
		)

		ts.Require().Error(err)
		ts.Require().Equal(ErrCodeNotFormatted, err)
	})
}

// TestFormatInstallToolsErrorScenarios tests error scenarios in InstallTools
func (ts *FormatTestSuite) TestFormatInstallToolsErrorScenarios() {
	ts.Run("gofumpt installation fails", func() {
		ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@latest"}).Return(fmt.Errorf("network error"))

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.InstallTools()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "failed to install gofumpt")
	})

	ts.Run("goimports installation fails", func() {
		// Mock gofumpt succeeding, goimports failing
		ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/tools/cmd/goimports@latest"}).Return(fmt.Errorf("network error"))

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.InstallTools()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "failed to install goimports")
	})
}

// TestFormatFileTypeScenarios tests formatting of different file types
func (ts *FormatTestSuite) TestFormatFileTypeScenarios() {
	ts.Run("YAML formatting with no files", func() {
		ts.env.Runner.On("RunCmdOutput", "find", []string{".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*"}).Return("", nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.YAML()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("JSON formatting with prettier", func() {
		ts.env.Runner.On("RunCmdOutput", "find", []string{".", "-name", "*.json", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*"}).Return("package.json\nconfig.json", nil)
		ts.env.Runner.On("RunCmd", "prettier", []string{"--write", "--ignore-path", ".github/.prettierignore", "package.json"}).Return(nil)
		ts.env.Runner.On("RunCmd", "prettier", []string{"--write", "--ignore-path", ".github/.prettierignore", "config.json"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.JSON()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("Markdown formatting with prettier", func() {
		ts.env.Runner.On("RunCmdOutput", "find", []string{".", "-name", "*.md", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*"}).Return("README.md", nil)
		ts.env.Runner.On("RunCmd", "prettier", []string{"--write", "--ignore-path", ".github/.prettierignore", "**/*.md"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Markdown()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("Shell formatting with shfmt", func() {
		ts.env.Runner.On("RunCmdOutput", "find", []string{".", "-name", "*.sh", "-o", "-name", "*.bash"}).Return("build.sh\ntest.bash", nil)
		ts.env.Runner.On("RunCmd", "shfmt", []string{"-i", "2", "-w", "."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Shell()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("SQL formatting with sqlfluff", func() {
		ts.env.Runner.On("RunCmdOutput", "find", []string{".", "-name", "*.sql"}).Return("schema.sql", nil)
		ts.env.Runner.On("RunCmd", "sqlfluff", []string{"format", "."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.SQL()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestFormatDefaultPartialFailures tests Default method with some formatters failing
func (ts *FormatTestSuite) TestFormatDefaultPartialFailures() {
	ts.Run("gofmt succeeds, fumpt fails, imports succeeds", func() {
		// Mock gofmt success
		ts.env.Runner.On("RunCmdOutput", "gofmt", []string{"-l", "."}).Return("", nil)

		// Mock fumpt failure (installation fails)
		ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@latest"}).Return(fmt.Errorf("network error"))

		// Mock imports success
		ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/tools/cmd/goimports@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "goimports", []string{"-w", "."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Default()
			},
		)

		// Default should continue even if some formatters fail
		ts.Require().NoError(err)
	})
}

// TestFormatWithEnvironmentVariables tests formatting with custom environment variables
func (ts *FormatTestSuite) TestFormatWithEnvironmentVariables() {
	ts.Run("custom exclude paths from environment", func() {
		// Set custom exclude paths
		originalEnv := os.Getenv("MAGE_X_FORMAT_EXCLUDE_PATHS")
		defer func() {
			if originalEnv != "" {
				os.Setenv("MAGE_X_FORMAT_EXCLUDE_PATHS", originalEnv)
			} else {
				os.Unsetenv("MAGE_X_FORMAT_EXCLUDE_PATHS")
			}
		}()
		os.Setenv("MAGE_X_FORMAT_EXCLUDE_PATHS", "build,dist,tmp")

		// Mock find command with custom exclude paths
		ts.env.Runner.On("RunCmdOutput", "find", []string{".", "-name", "*.json", "-not", "-path", "./build/*", "-not", "-path", "./*build*/*", "-not", "-path", "./dist/*", "-not", "-path", "./*dist*/*", "-not", "-path", "./tmp/*", "-not", "-path", "./*tmp*/*"}).Return("package.json", nil)
		ts.env.Runner.On("RunCmd", "prettier", []string{"--write", "--ignore-path", ".github/.prettierignore", "package.json"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.JSON()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestFormatFixMethod tests the Fix method comprehensive functionality
func (ts *FormatTestSuite) TestFormatFixMethod() {
	ts.Run("fix all formatting issues", func() {
		// Mock all formatter commands called by Fix method
		// Gofmt commands
		ts.env.Runner.On("RunCmdOutput", "gofmt", []string{"-l", "."}).Return("file1.go", nil)
		ts.env.Runner.On("RunCmd", "gofmt", []string{"-w", "."}).Return(nil)

		// Fumpt commands (ensureGofumpt + run gofumpt)
		ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "gofumpt", []string{"-w", "-extra", "."}).Return(nil)

		// Imports commands (ensureGoimports + run goimports)
		ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/tools/cmd/goimports@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "goimports", []string{"-w", "."}).Return(nil)

		// JSON formatting commands
		ts.env.Runner.On("RunCmdOutput", "find", []string{".", "-name", "*.json", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*"}).Return("package.json", nil)
		ts.env.Runner.On("RunCmd", "prettier", []string{"--write", "--ignore-path", ".github/.prettierignore", "package.json"}).Return(nil)

		// YAML formatting commands
		ts.env.Runner.On("RunCmdOutput", "find", []string{".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*"}).Return("config.yml", nil)
		ts.env.Runner.On("RunCmd", "prettier", []string{"--write", "--ignore-path", ".github/.prettierignore", "**/*.{yml,yaml}"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Fix()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("fix continues despite individual formatter failures", func() {
		// Mock some formatters failing
		ts.env.Runner.On("RunCmdOutput", "gofmt", []string{"-l", "."}).Return("", fmt.Errorf("gofmt failed"))
		ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@latest"}).Return(fmt.Errorf("fumpt install failed"))
		ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/tools/cmd/goimports@latest"}).Return(fmt.Errorf("goimports install failed"))

		// JSON and YAML should still be attempted
		ts.env.Runner.On("RunCmdOutput", "find", []string{".", "-name", "*.json", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*"}).Return("", nil)                         // No JSON files
		ts.env.Runner.On("RunCmdOutput", "find", []string{".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*"}).Return("", nil) // No YAML files

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				return ts.format.Fix()
			},
		)

		// Fix should complete successfully even with failures
		ts.Require().NoError(err)
	})
}

// TestFormatTestSuite runs the test suite
func TestFormatTestSuite(t *testing.T) {
	suite.Run(t, new(FormatTestSuite))
}
