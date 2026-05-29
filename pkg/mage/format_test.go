//go:build integration
// +build integration

package mage

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
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
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
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
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
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
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
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
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
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
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
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
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
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
		// Stage real YAML files; discovery now uses a native filesystem walk.
		ts.env.CreateFile("config.yml", "a: 1\n")
		ts.env.CreateFile("data.yaml", "b: 2\n")
		// No .github/.yamlfmt present, so yamlfmt receives the explicit file list in
		// lexical walk order (config.yml then data.yaml) without -conf.
		ts.env.Runner.On("RunCmd", "yamlfmt", []string{"config.yml", "data.yaml"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.format.YAML()
			},
		)

		ts.Require().NoError(err)
		ts.env.Runner.AssertExpectations(ts.T())
	})
}

// TestFormatYaml tests the Yaml method (alias)
func (ts *FormatTestSuite) TestFormatYaml() {
	ts.Run("successful Yaml formatting (alias)", func() {
		ts.env.CreateFile("config.yml", "a: 1\n")
		ts.env.CreateFile("data.yaml", "b: 2\n")
		ts.env.Runner.On("RunCmd", "yamlfmt", []string{"config.yml", "data.yaml"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.format.Yaml()
			},
		)

		ts.Require().NoError(err)
		ts.env.Runner.AssertExpectations(ts.T())
	})
}

// TestFormatJSON tests the JSON method
func (ts *FormatTestSuite) TestFormatJSON() {
	ts.Run("successful JSON formatting", func() {
		// JSON formatting uses native Go and a native walk: no external commands run.
		ts.env.CreateFile("package.json", `{"name":"test","value":123}`)
		ts.env.CreateFile("config.json", `{"a":1}`)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.format.JSON()
			},
		)

		ts.Require().NoError(err)
		// Native formatting rewrites the files with indentation.
		ts.Require().Contains(ts.env.ReadFile("package.json"), "\n    ")
	})
}

// TestFormatJson tests the Json method (alias)
func (ts *FormatTestSuite) TestFormatJson() {
	ts.Run("successful Json formatting (alias)", func() {
		ts.env.CreateFile("package.json", `{"name":"test","value":123}`)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
			func() error {
				return ts.format.JSON()
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

		// gci commands (ensureGci + run gci)
		ts.env.Runner.On("RunCmd", "go", []string{"install", "github.com/daixiang0/gci@latest"}).Return(nil).Maybe()
		ts.env.Runner.On("RunCmd", "gci", mock.Anything).Return(nil).Maybe()

		// JSON formatting now uses native Go and a native walk (no external commands).
		ts.env.CreateFile("package.json", `{"name":"test"}`)

		// YAML formatting discovers files via a native walk and feeds yamlfmt the
		// explicit list (no -conf, since .github/.yamlfmt is absent here).
		ts.env.CreateFile("config.yml", "a: 1\n")
		ts.env.Runner.On("RunCmd", "yamlfmt", []string{"config.yml"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() any { return GetRunner() },
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
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
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
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
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
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
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
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
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
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
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
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
			func() error {
				return ts.format.Imports()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "goimports failed")
	})
}

// TestFormatYamlfmtScenarios tests various yamlfmt scenarios through Format.YAML.
func (ts *FormatTestSuite) TestFormatYamlfmtScenarios() {
	ts.Run("yamlfmt formats discovered files", func() {
		ts.env.CreateFile("test.yml", "a: 1\n")
		ts.env.Runner.On("RunCmd", "yamlfmt", []string{"test.yml"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
			func() error {
				return ts.format.YAML()
			},
		)

		ts.Require().NoError(err)
		ts.env.Runner.AssertExpectations(ts.T())
	})

	ts.Run("yamlfmt execution fails", func() {
		ts.env.CreateFile("test.yml", "a: 1\n")
		// The batch run fails, then the per-file fallback retries the same file; the
		// expectation matches both invocations.
		ts.env.Runner.On("RunCmd", "yamlfmt", []string{"test.yml"}).Return(fmt.Errorf("yamlfmt failed"))

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
			func() error {
				return ts.format.YAML()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "yamlfmt formatting failed")
	})

	ts.Run("yamlfmt with config file", func() {
		ts.env.CreateFile(".github/.yamlfmt", "formatter:\n  type: basic\n")
		ts.env.CreateFile("test.yml", "a: 1\n")
		ts.env.Runner.On("RunCmd", "yamlfmt", []string{"-conf", ".github/.yamlfmt", "test.yml"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
			func() error {
				return ts.format.YAML()
			},
		)

		ts.Require().NoError(err)
		ts.env.Runner.AssertExpectations(ts.T())
	})

	ts.Run("yamlfmt no files found", func() {
		// Bare temp dir (only go.mod): the native walk finds no YAML files, so yamlfmt
		// is never invoked.
		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
			func() error {
				return ts.format.YAML()
			},
		)

		ts.Require().NoError(err) // Should succeed when no files found
	})

	ts.Run("yamlfmt installation fails", func() {
		// Mock installation failure
		ts.env.Runner.On("RunCmd", "go", []string{"install", "github.com/google/yamlfmt/cmd/yamlfmt@latest"}).Return(fmt.Errorf("network error"))

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
			func() error {
				// This would trigger installation which would fail
				return ensureYamlfmt()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "network error")
	})
}

// TestFormatCheckWithFormatIssues tests Check method with various formatting issues
func (ts *FormatTestSuite) TestFormatCheckWithFormatIssues() {
	ts.Run("gofmt issues detected", func() {
		ts.env.Runner.On("RunCmdOutput", "gofmt", []string{"-l", "."}).Return("file1.go\nfile2.go", nil)
		ts.env.Runner.On("RunCmdOutput", "gofumpt", []string{"-l", "-extra", "."}).Return("", nil)
		ts.env.Runner.On("RunCmdOutput", "goimports", []string{"-l", "."}).Return("", nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
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
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
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
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
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
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
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
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
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
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
			func() error {
				return ts.format.InstallTools()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "failed to install goimports")
	})

	ts.Run("yamlfmt installation fails", func() {
		// Mock gofumpt and goimports succeeding, yamlfmt failing
		ts.env.Runner.On("RunCmd", "go", []string{"install", "mvdan.cc/gofumpt@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "go", []string{"install", "golang.org/x/tools/cmd/goimports@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "go", []string{"install", "github.com/google/yamlfmt/cmd/yamlfmt@latest"}).Return(fmt.Errorf("network error"))

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
			func() error {
				return ts.format.InstallTools()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "failed to install yamlfmt")
	})
}

// TestFormatFileTypeScenarios tests formatting of different file types
func (ts *FormatTestSuite) TestFormatFileTypeScenarios() {
	ts.Run("YAML formatting with no files", func() {
		// Bare temp dir: native walk finds no YAML files.
		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
			func() error {
				return ts.format.YAML()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("JSON formatting with native Go", func() {
		// JSON formatting uses native Go via a native walk - no external commands.
		ts.env.CreateFile("package.json", `{"name":"test"}`)
		ts.env.CreateFile("config.json", `{"a":1}`)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
			func() error {
				return ts.format.JSON()
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
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
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

		// Stage a JSON file in an excluded directory and one outside it; the native
		// walk must skip the excluded directory entirely.
		ts.env.CreateFile("package.json", `{"name":"test"}`)
		ts.env.CreateFile("build/skip.json", `not valid json that would error if formatted`)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
			func() error {
				return ts.format.JSON()
			},
		)

		// Succeeds: the invalid JSON under build/ is excluded and never parsed.
		ts.Require().NoError(err)
		ts.Require().Equal(`not valid json that would error if formatted`, ts.env.ReadFile("build/skip.json"))
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

		// gci commands (ensureGci + run gci) - tolerated whether or not gci is on PATH.
		ts.env.Runner.On("RunCmd", "go", []string{"install", "github.com/daixiang0/gci@latest"}).Return(nil).Maybe()
		ts.env.Runner.On("RunCmd", "gci", mock.Anything).Return(nil).Maybe()

		// JSON formatting uses native Go via a native walk (no external commands).
		ts.env.CreateFile("package.json", `{"name":"test"}`)

		// YAML formatting discovers files via a native walk and feeds yamlfmt the
		// explicit list.
		ts.env.CreateFile("config.yml", "a: 1\n")
		ts.env.Runner.On("RunCmd", "yamlfmt", []string{"config.yml"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
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
		ts.env.Runner.On("RunCmd", "go", []string{"install", "github.com/daixiang0/gci@latest"}).Return(fmt.Errorf("gci install failed")).Maybe()
		ts.env.Runner.On("RunCmd", "gci", mock.Anything).Return(fmt.Errorf("gci failed")).Maybe()

		// JSON and YAML still run via native walk; the bare temp dir has no such files.

		err := ts.env.WithMockRunner(
			func(r any) error { return SetRunner(r.(CommandRunner)) },
			func() any { return GetRunner() },
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
