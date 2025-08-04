package builders_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/builders"
)

// mockLintConfig implements the LintConfig interface for testing
type mockLintConfig struct {
	timeout string
}

func (c mockLintConfig) GetTimeout() string {
	return c.timeout
}

// mockBuildConfig implements the BuildConfig interface for testing
type mockBuildConfig struct {
	verbose  bool
	parallel int
	tags     []string
}

func (c mockBuildConfig) GetVerbose() bool {
	return c.verbose
}

func (c mockBuildConfig) GetParallel() int {
	return c.parallel
}

func (c mockBuildConfig) GetTags() []string {
	return c.tags
}

// mockConfig implements the Config interface for testing
type mockConfig struct {
	lint  builders.LintConfig
	build builders.BuildConfig
}

func (c mockConfig) GetLint() builders.LintConfig {
	return c.lint
}

func (c mockConfig) GetBuild() builders.BuildConfig {
	return c.build
}

// mockModule implements the Module interface for testing
type mockModule struct {
	path string
}

func (m mockModule) GetPath() string {
	return m.path
}

// LintCommandBuilderTestSuite tests the LintCommandBuilder
type LintCommandBuilderTestSuite struct {
	suite.Suite

	builder *builders.LintCommandBuilder
	config  mockConfig
	module  mockModule
	tempDir string
}

func (ts *LintCommandBuilderTestSuite) SetupTest() {
	// Create temp directory for test files
	tempDir, err := os.MkdirTemp("", "lint_builder_test")
	ts.Require().NoError(err)
	ts.tempDir = tempDir

	// Setup mock configuration
	ts.config = mockConfig{
		lint: mockLintConfig{
			timeout: "5m",
		},
		build: mockBuildConfig{
			verbose:  true,
			parallel: 4,
			tags:     []string{"integration", "e2e"},
		},
	}

	ts.module = mockModule{
		path: tempDir,
	}

	ts.builder = builders.NewLintCommandBuilder(ts.config)
}

func (ts *LintCommandBuilderTestSuite) TearDownTest() {
	if ts.tempDir != "" {
		err := os.RemoveAll(ts.tempDir)
		ts.Require().NoError(err)
	}
}

func (ts *LintCommandBuilderTestSuite) TestNewLintCommandBuilder() {
	builder := builders.NewLintCommandBuilder(ts.config)
	ts.Require().NotNil(builder)
}

func (ts *LintCommandBuilderTestSuite) TestBuildGolangciArgsDefault() {
	options := builders.LintOptions{}
	args := ts.builder.BuildGolangciArgs(ts.module, options)

	expected := []string{
		"run",
		"--timeout", "5m",
		"--verbose",
		"--concurrency", "4",
		"--build-tags", "integration,e2e",
		"./...",
	}

	ts.Require().Equal(expected, args)
}

func (ts *LintCommandBuilderTestSuite) TestBuildGolangciArgsWithConfig() {
	// Create a .golangci.json file
	configFile := filepath.Join(ts.tempDir, ".golangci.json")
	err := os.WriteFile(configFile, []byte(`{"linters": {}}`), 0o600)
	ts.Require().NoError(err)

	options := builders.LintOptions{}
	args := ts.builder.BuildGolangciArgs(ts.module, options)

	ts.Require().Contains(args, "--config")
	ts.Require().Contains(args, configFile)
}

func (ts *LintCommandBuilderTestSuite) TestBuildGolangciArgsWithRootConfig() {
	// Save current directory
	origDir, err := os.Getwd()
	ts.Require().NoError(err)
	defer func() {
		chdirErr := os.Chdir(origDir)
		ts.Require().NoError(chdirErr)
	}()

	// Change to temp directory
	err = os.Chdir(ts.tempDir)
	ts.Require().NoError(err)

	// Create root .golangci.json file
	err = os.WriteFile(".golangci.json", []byte(`{"linters": {}}`), 0o600)
	ts.Require().NoError(err)

	// Create module in subdirectory
	moduleDir := filepath.Join(ts.tempDir, "submodule")
	err = os.MkdirAll(moduleDir, 0o750)
	ts.Require().NoError(err)

	module := mockModule{path: moduleDir}
	options := builders.LintOptions{}
	args := ts.builder.BuildGolangciArgs(module, options)

	ts.Require().Contains(args, "--config")
	// Check that some config path is present (could be absolute path)
	configIndex := -1
	for i, arg := range args {
		if arg == "--config" && i+1 < len(args) {
			configIndex = i + 1
			break
		}
	}
	ts.Require().NotEqual(-1, configIndex)
	ts.Require().Contains(args[configIndex], ".golangci.json")
}

func (ts *LintCommandBuilderTestSuite) TestBuildGolangciArgsNoConfig() {
	options := builders.LintOptions{NoConfig: true}
	args := ts.builder.BuildGolangciArgs(ts.module, options)

	ts.Require().NotContains(args, "--config")
}

func (ts *LintCommandBuilderTestSuite) TestBuildGolangciArgsWithOptions() {
	options := builders.LintOptions{
		Fix:    true,
		Format: "json",
		Output: "report.json",
	}
	args := ts.builder.BuildGolangciArgs(ts.module, options)

	ts.Require().Contains(args, "--fix")
	ts.Require().Contains(args, "--out-format")
	ts.Require().Contains(args, "json")
	ts.Require().Contains(args, "--out")
	ts.Require().Contains(args, "report.json")
}

func (ts *LintCommandBuilderTestSuite) TestBuildGolangciArgsEmptyConfig() {
	// Test with empty configuration values
	config := mockConfig{
		lint: mockLintConfig{
			timeout: "",
		},
		build: mockBuildConfig{
			verbose:  false,
			parallel: 0,
			tags:     []string{},
		},
	}

	builder := builders.NewLintCommandBuilder(config)
	options := builders.LintOptions{}
	args := builder.BuildGolangciArgs(ts.module, options)

	expected := []string{
		"run",
		"./...",
	}

	ts.Require().Equal(expected, args)
}

func (ts *LintCommandBuilderTestSuite) TestBuildVetArgs() {
	args := ts.builder.BuildVetArgs()

	expected := []string{
		"vet",
		"-tags", "integration,e2e",
		"-v",
		"./...",
	}

	ts.Require().Equal(expected, args)
}

func (ts *LintCommandBuilderTestSuite) TestBuildVetArgsMinimal() {
	config := mockConfig{
		build: mockBuildConfig{
			verbose: false,
			tags:    []string{},
		},
	}

	builder := builders.NewLintCommandBuilder(config)
	args := builder.BuildVetArgs()

	expected := []string{
		"vet",
		"./...",
	}

	ts.Require().Equal(expected, args)
}

func (ts *LintCommandBuilderTestSuite) TestBuildStaticcheckArgs() {
	args := ts.builder.BuildStaticcheckArgs()

	expected := []string{
		"-f", "text",
		"-tags", "integration,e2e",
		"./...",
	}

	ts.Require().Equal(expected, args)
}

func (ts *LintCommandBuilderTestSuite) TestBuildStaticcheckArgsNoTags() {
	config := mockConfig{
		build: mockBuildConfig{
			tags: []string{},
		},
	}

	builder := builders.NewLintCommandBuilder(config)
	args := builder.BuildStaticcheckArgs()

	expected := []string{
		"-f", "text",
		"./...",
	}

	ts.Require().Equal(expected, args)
}

func (ts *LintCommandBuilderTestSuite) TestBuildGofmtArgsCheckOnly() {
	args := ts.builder.BuildGofmtArgs(true)

	expected := []string{
		"-l",
		".",
	}

	ts.Require().Equal(expected, args)
}

func (ts *LintCommandBuilderTestSuite) TestBuildGofmtArgsWrite() {
	args := ts.builder.BuildGofmtArgs(false)

	expected := []string{
		"-w",
		".",
	}

	ts.Require().Equal(expected, args)
}

func (ts *LintCommandBuilderTestSuite) TestBuildGofumptArgs() {
	args := ts.builder.BuildGofumptArgs(false)

	expected := []string{
		"-w",
		".",
	}

	ts.Require().Equal(expected, args)
}

func (ts *LintCommandBuilderTestSuite) TestBuildGofumptArgsWithExtra() {
	args := ts.builder.BuildGofumptArgs(true)

	expected := []string{
		"-w",
		"-extra",
		".",
	}

	ts.Require().Equal(expected, args)
}

func (ts *LintCommandBuilderTestSuite) TestBuildGoimportsArgs() {
	args := ts.builder.BuildGoimportsArgs()

	expected := []string{
		"-w",
		".",
	}

	ts.Require().Equal(expected, args)
}

func (ts *LintCommandBuilderTestSuite) TestFindLintConfigYAML() {
	// Create a .golangci.yml file
	configFile := filepath.Join(ts.tempDir, ".golangci.yml")
	err := os.WriteFile(configFile, []byte(`linters: {}`), 0o600)
	ts.Require().NoError(err)

	options := builders.LintOptions{}
	args := ts.builder.BuildGolangciArgs(ts.module, options)

	ts.Require().Contains(args, "--config")
	ts.Require().Contains(args, configFile)
}

func (ts *LintCommandBuilderTestSuite) TestFindLintConfigRootYAML() {
	// Save current directory
	origDir, err := os.Getwd()
	ts.Require().NoError(err)
	defer func() {
		chdirErr := os.Chdir(origDir)
		ts.Require().NoError(chdirErr)
	}()

	// Change to temp directory
	err = os.Chdir(ts.tempDir)
	ts.Require().NoError(err)

	// Create root .golangci.yml file
	err = os.WriteFile(".golangci.yml", []byte(`linters: {}`), 0o600)
	ts.Require().NoError(err)

	// Create module in subdirectory
	moduleDir := filepath.Join(ts.tempDir, "submodule")
	err = os.MkdirAll(moduleDir, 0o750)
	ts.Require().NoError(err)

	module := mockModule{path: moduleDir}
	options := builders.LintOptions{}
	args := ts.builder.BuildGolangciArgs(module, options)

	ts.Require().Contains(args, "--config")
	// Check that some config path is present (could be absolute path)
	configIndex := -1
	for i, arg := range args {
		if arg == "--config" && i+1 < len(args) {
			configIndex = i + 1
			break
		}
	}
	ts.Require().NotEqual(-1, configIndex)
	ts.Require().Contains(args[configIndex], ".golangci.yml")
}

func TestLintCommandBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(LintCommandBuilderTestSuite))
}
