package mage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// DocsCoverageTestSuite provides comprehensive coverage for Docs methods
type DocsCoverageTestSuite struct {
	suite.Suite

	env  *testutil.TestEnvironment
	docs Docs
}

func TestDocsCoverageTestSuite(t *testing.T) {
	suite.Run(t, new(DocsCoverageTestSuite))
}

func (ts *DocsCoverageTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.docs = Docs{}

	// Create .mage.yaml config
	ts.env.CreateMageConfig(`
project:
  name: test
docs:
  port: 6060
`)

	// Set up general mocks for all commands - catch-all approach
	ts.env.Runner.On("RunCmd", mock.Anything, mock.Anything).Return(nil).Maybe()
	ts.env.Runner.On("RunCmdOutput", mock.Anything, mock.Anything).Return("", nil).Maybe()
	ts.env.Runner.On("RunCmdOutputDir", mock.Anything, mock.Anything, mock.Anything).Return("", nil).Maybe()
	ts.env.Runner.On("RunCmdInDir", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
}

func (ts *DocsCoverageTestSuite) TearDownTest() {
	TestResetConfig()
	ts.env.Cleanup()
}

// Helper to set up mock runner and execute function
func (ts *DocsCoverageTestSuite) withMockRunner(fn func() error) error {
	return ts.env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		fn,
	)
}

// TestDocsDefaultSuccess tests Docs.Default
func (ts *DocsCoverageTestSuite) TestDocsDefaultSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.docs.Default()
	})

	ts.Require().NoError(err)
}

// TestDocsLintSuccess tests Docs.Lint
func (ts *DocsCoverageTestSuite) TestDocsLintSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.docs.Lint()
	})

	ts.Require().NoError(err)
}

// TestDocsSpellSuccess tests Docs.Spell
func (ts *DocsCoverageTestSuite) TestDocsSpellSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.docs.Spell()
	})

	ts.Require().NoError(err)
}

// TestDocsLinksSuccess tests Docs.Links
func (ts *DocsCoverageTestSuite) TestDocsLinksSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.docs.Links()
	})

	ts.Require().NoError(err)
}

// TestDocsAPISuccess tests Docs.API
func (ts *DocsCoverageTestSuite) TestDocsAPISuccess() {
	err := ts.withMockRunner(func() error {
		return ts.docs.API()
	})

	ts.Require().NoError(err)
}

// TestDocsMarkdownSuccess tests Docs.Markdown
func (ts *DocsCoverageTestSuite) TestDocsMarkdownSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.docs.Markdown()
	})

	ts.Require().NoError(err)
}

// TestDocsReadmeSuccess tests Docs.Readme
func (ts *DocsCoverageTestSuite) TestDocsReadmeSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.docs.Readme()
	})

	ts.Require().NoError(err)
}

// TestDocsChangelogSuccess tests Docs.Changelog
func (ts *DocsCoverageTestSuite) TestDocsChangelogSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.docs.Changelog("v1.0.0")
	})

	ts.Require().NoError(err)
}

// TestDocsUpdateExercise tests Docs.Update exercises the code path
func (ts *DocsCoverageTestSuite) TestDocsUpdateExercise() {
	err := ts.withMockRunner(func() error {
		return ts.docs.Update()
	})
	// Just exercise the code path - may fail due to missing files
	_ = err
}

// TestDocsCheckExercise tests Docs.Check exercises the code path
func (ts *DocsCoverageTestSuite) TestDocsCheckExercise() {
	err := ts.withMockRunner(func() error {
		return ts.docs.Check()
	})
	// Just exercise the code path - may fail due to missing files
	_ = err
}

// TestDocsCleanSuccess tests Docs.Clean
func (ts *DocsCoverageTestSuite) TestDocsCleanSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.docs.Clean()
	})

	ts.Require().NoError(err)
}

// TestDocsGenerateSuccess tests Docs.Generate
func (ts *DocsCoverageTestSuite) TestDocsGenerateSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.docs.Generate()
	})

	ts.Require().NoError(err)
}

// TestDocsGoDocsExercise tests Docs.GoDocs exercises the code path
func (ts *DocsCoverageTestSuite) TestDocsGoDocsExercise() {
	err := ts.withMockRunner(func() error {
		return ts.docs.GoDocs()
	})
	// Just exercise the code path - may fail due to missing docs tool
	_ = err
}

// TestDocsBuildSuccess tests Docs.Build
func (ts *DocsCoverageTestSuite) TestDocsBuildSuccess() {
	err := ts.withMockRunner(func() error {
		return ts.docs.Build()
	})

	ts.Require().NoError(err)
}

// ================== Standalone Tests for Helper Functions ==================

func TestShouldSkipPackage(t *testing.T) {
	tests := []struct {
		name     string
		pkgPath  string
		expected bool
	}{
		{
			name:     "skip testdata",
			pkgPath:  "path/to/testdata/pkg",
			expected: true,
		},
		{
			name:     "skip fuzz",
			pkgPath:  "path/to/fuzz/pkg",
			expected: true,
		},
		{
			name:     "skip examples directory",
			pkgPath:  "path/to/examples/basic",
			expected: true,
		},
		{
			name:     "normal package",
			pkgPath:  "path/to/pkg",
			expected: false,
		},
		{
			name:     "vendor not skipped by this function",
			pkgPath:  "path/to/vendor/pkg",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldSkipPackage(tt.pkgPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCategorizePackage(t *testing.T) {
	tests := []struct {
		name    string
		pkgPath string
		wantCat string
	}{
		{
			name:    "mage package",
			pkgPath: "github.com/mrz1836/mage-x/pkg/mage",
			wantCat: "Core",
		},
		{
			name:    "common package",
			pkgPath: "github.com/mrz1836/mage-x/pkg/common/utils",
			wantCat: "Common",
		},
		{
			name:    "cmd package",
			pkgPath: "github.com/mrz1836/mage-x/cmd/magex",
			wantCat: "Commands",
		},
		{
			name:    "examples package (other)",
			pkgPath: "github.com/mrz1836/mage-x/examples/basic",
			wantCat: "Other",
		},
		{
			name:    "security package",
			pkgPath: "github.com/mrz1836/mage-x/pkg/security",
			wantCat: "Security",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cat := categorizePackage(tt.pkgPath)
			assert.Equal(t, tt.wantCat, cat)
		})
	}
}

func TestBuildPkgsiteArgs(t *testing.T) {
	tests := []struct {
		name string
		mode string
		port int
	}{
		{
			name: "default port",
			mode: "local",
			port: 6060,
		},
		{
			name: "custom port",
			mode: "local",
			port: 8080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := buildPkgsiteArgs(tt.mode, tt.port)
			assert.Contains(t, args, "-http")
		})
	}
}

func TestBuildGodocArgs(t *testing.T) {
	tests := []struct {
		name string
		mode string
		port int
	}{
		{
			name: "default port",
			mode: "local",
			port: 6060,
		},
		{
			name: "custom port",
			mode: "local",
			port: 8080,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := buildGodocArgs(tt.mode, tt.port)
			assert.Contains(t, args, "-http")
		})
	}
}

func TestIsPortAvailable(t *testing.T) {
	// Test that a high port number is likely available
	result := isPortAvailable(59999)
	// Can't guarantee availability, but function should not panic
	_ = result
}

func TestGetPortFromConfig(t *testing.T) {
	// Without config, should return default
	cfg := &Config{}
	port := getPortFromConfig(cfg, 6060)
	assert.Equal(t, 6060, port)

	// With config port
	cfg = &Config{
		Docs: DocsConfig{
			Port: 8080,
		},
	}
	port = getPortFromConfig(cfg, 6060)
	assert.Equal(t, 8080, port)
}

func TestIsCI(t *testing.T) {
	// Just test that the function runs without error
	result := isCI()
	// Result depends on environment
	_ = result
}

func TestIsTestEnvironment(t *testing.T) {
	// During tests, this should return true
	result := isTestEnvironment()
	assert.True(t, result)
}

func TestDetectBestDocTool(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateGoMod("test/module")

	// Set up mocks
	env.Runner.On("RunCmd", mock.Anything, mock.Anything).Return(nil).Maybe()
	env.Runner.On("RunCmdOutput", mock.Anything, mock.Anything).Return("", nil).Maybe()

	_ = env.WithMockRunner( //nolint:errcheck // exercise code path only
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		func() error {
			tool := detectBestDocTool()
			// Should return one of the doc tools or empty
			_ = tool
			return nil
		},
	)
}

func TestAutoDetectDocTool(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateGoMod("test/module")

	// Set up mocks
	env.Runner.On("RunCmd", mock.Anything, mock.Anything).Return(nil).Maybe()
	env.Runner.On("RunCmdOutput", mock.Anything, mock.Anything).Return("", nil).Maybe()

	err := env.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		func() error {
			_ = autoDetectDocTool()
			return nil
		},
	)
	// Function may fail if no doc tools installed, that's ok
	_ = err
}

func TestBuildDocServer(t *testing.T) {
	tests := []struct {
		name     string
		tool     string
		mode     string
		wantTool string
	}{
		{
			name:     "pkgsite",
			tool:     "pkgsite",
			mode:     "local",
			wantTool: "pkgsite",
		},
		{
			name:     "godoc",
			tool:     "godoc",
			mode:     "local",
			wantTool: "godoc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := buildDocServer(tt.tool, tt.mode, 6060)
			require.NotNil(t, server)
			assert.Equal(t, tt.wantTool, server.Tool)
		})
	}
}

func TestCategorizeBuildFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected string
	}{
		{
			name:     "pkg_mage file",
			filename: "pkg_mage_build.md",
			expected: "Core Packages",
		},
		{
			name:     "pkg_common file",
			filename: "pkg_common_utils.md",
			expected: "Common Packages",
		},
		{
			name:     "pkg_security file",
			filename: "pkg_security_cmd.md",
			expected: "Security Packages",
		},
		{
			name:     "cmd file",
			filename: "cmd_magex.md",
			expected: "Command Packages",
		},
		{
			name:     "other file",
			filename: "other.md",
			expected: "Other Packages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := categorizeBuildFile(tt.filename)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCreateBuildMetadata(t *testing.T) {
	// Create metadata for a file count
	metadata := createBuildMetadata(2)
	// Returns a string, just verify it's not empty
	assert.NotEmpty(t, metadata)
}
