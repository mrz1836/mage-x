package builders_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/builders"
)

// mockTestConfig implements the TestConfig interface for testing
type mockTestConfig struct {
	timeout            string
	integrationTimeout string
	integrationTag     string
	coverMode          string
	parallel           int
	tags               string
	shuffle            bool
	benchCPU           int
	benchTime          string
	benchMem           bool
	coverageExclude    []string
}

func (c mockTestConfig) GetTimeout() string {
	return c.timeout
}

func (c mockTestConfig) GetIntegrationTimeout() string {
	return c.integrationTimeout
}

func (c mockTestConfig) GetIntegrationTag() string {
	return c.integrationTag
}

func (c mockTestConfig) GetCoverMode() string {
	return c.coverMode
}

func (c mockTestConfig) GetParallel() int {
	return c.parallel
}

func (c mockTestConfig) GetTags() string {
	return c.tags
}

func (c mockTestConfig) GetShuffle() bool {
	return c.shuffle
}

func (c mockTestConfig) GetBenchCPU() int {
	return c.benchCPU
}

func (c mockTestConfig) GetBenchTime() string {
	return c.benchTime
}

func (c mockTestConfig) GetBenchMem() bool {
	return c.benchMem
}

func (c mockTestConfig) GetCoverageExclude() []string {
	return c.coverageExclude
}

// mockTestBuildConfig implements the TestBuildConfig interface for testing
type mockTestBuildConfig struct {
	verbose bool
}

func (c mockTestBuildConfig) GetVerbose() bool {
	return c.verbose
}

// mockTestConfigProvider implements the TestConfigProvider interface for testing
type mockTestConfigProvider struct {
	test  builders.TestConfig
	build builders.TestBuildConfig
}

func (c mockTestConfigProvider) GetTest() builders.TestConfig {
	return c.test
}

func (c mockTestConfigProvider) GetBuild() builders.TestBuildConfig {
	return c.build
}

// TestCommandBuilderTestSuite tests the TestCommandBuilder
type TestCommandBuilderTestSuite struct {
	suite.Suite

	builder *builders.TestCommandBuilder
	config  mockTestConfigProvider
}

func (ts *TestCommandBuilderTestSuite) SetupTest() {
	// Setup mock configuration
	ts.config = mockTestConfigProvider{
		test: mockTestConfig{
			timeout:            "10m",
			integrationTimeout: "30m",
			integrationTag:     "integration",
			coverMode:          "atomic",
			parallel:           4,
			tags:               "unit,integration",
			shuffle:            true,
			benchCPU:           2,
			benchTime:          "10s",
			benchMem:           true,
			coverageExclude:    []string{"excluded"},
		},
		build: mockTestBuildConfig{
			verbose: true,
		},
	}

	ts.builder = builders.NewTestCommandBuilder(ts.config)
}

func (ts *TestCommandBuilderTestSuite) TestNewTestCommandBuilder() {
	builder := builders.NewTestCommandBuilder(ts.config)
	ts.Require().NotNil(builder)
}

func (ts *TestCommandBuilderTestSuite) TestBuildUnitTestArgsDefault() {
	options := builders.TestOptions{}
	args := ts.builder.BuildUnitTestArgs(options)

	expected := []string{
		"test",
		"-v",
		"-timeout", "10m",
		"-parallel", "4",
		"-tags", "unit,integration",
		"-shuffle=on",
		"./...",
	}

	ts.Require().Equal(expected, args)
}

func (ts *TestCommandBuilderTestSuite) TestBuildUnitTestArgsWithShort() {
	options := builders.TestOptions{Short: true}
	args := ts.builder.BuildUnitTestArgs(options)

	ts.Require().Contains(args, "-short")
}

func (ts *TestCommandBuilderTestSuite) TestBuildUnitTestArgsWithPattern() {
	options := builders.TestOptions{Pattern: "./pkg/..."}
	args := ts.builder.BuildUnitTestArgs(options)

	ts.Require().Contains(args, "./pkg/...")
	ts.Require().NotContains(args, "./...")
}

func (ts *TestCommandBuilderTestSuite) TestBuildUnitTestArgsWithRace() {
	// Only test race detection on non-Windows systems
	if runtime.GOOS != "windows" {
		options := builders.TestOptions{Race: true}
		args := ts.builder.BuildUnitTestArgs(options)

		ts.Require().Contains(args, "-race")
	}
}

func (ts *TestCommandBuilderTestSuite) TestBuildIntegrationTestArgs() {
	options := builders.TestOptions{}
	args := ts.builder.BuildIntegrationTestArgs(options)

	// Integration tests should have both timeouts (base test timeout first, then integration timeout)
	ts.Require().Contains(args, "test")
	ts.Require().Contains(args, "-v")
	ts.Require().Contains(args, "-timeout")
	ts.Require().Contains(args, "30m") // Integration timeout
	ts.Require().Contains(args, "-parallel")
	ts.Require().Contains(args, "4")
	ts.Require().Contains(args, "-tags")
	ts.Require().Contains(args, "unit,integration")
	ts.Require().Contains(args, "-shuffle=on")
	ts.Require().Contains(args, "-tags")
	ts.Require().Contains(args, "integration")
	ts.Require().Contains(args, "./...")
}

func (ts *TestCommandBuilderTestSuite) TestBuildIntegrationTestArgsWithPattern() {
	options := builders.TestOptions{Pattern: "./integration/..."}
	args := ts.builder.BuildIntegrationTestArgs(options)

	ts.Require().Contains(args, "./integration/...")
	ts.Require().NotContains(args, "./...")
}

func (ts *TestCommandBuilderTestSuite) TestBuildCoverageArgs() {
	options := builders.TestOptions{Coverage: true}
	args := ts.builder.BuildCoverageArgs(options)

	// Coverage tests should not have shuffle enabled (shuffle is disabled when options.Coverage is true)
	expected := []string{
		"test",
		"-v",
		"-timeout", "10m",
		"-parallel", "4",
		"-tags", "unit,integration",
		"-cover",
		"-covermode=atomic",
		"./...",
	}

	ts.Require().Equal(expected, args)
}

func (ts *TestCommandBuilderTestSuite) TestBuildCoverageArgsWithCoverageFile() {
	options := builders.TestOptions{
		CoverageFile: "coverage.out",
	}
	args := ts.builder.BuildCoverageArgs(options)

	ts.Require().Contains(args, "-coverprofile=coverage.out")
}

func (ts *TestCommandBuilderTestSuite) TestBuildCoverageArgsWithCoverageMode() {
	options := builders.TestOptions{
		CoverageMode: "count",
	}
	args := ts.builder.BuildCoverageArgs(options)

	ts.Require().Contains(args, "-covermode=count")
}

func (ts *TestCommandBuilderTestSuite) TestBuildCoverageArgsShuffleDisabled() {
	options := builders.TestOptions{Coverage: true}
	args := ts.builder.BuildCoverageArgs(options)

	// Shuffle should be disabled for coverage tests
	ts.Require().NotContains(args, "-shuffle=on")
}

func (ts *TestCommandBuilderTestSuite) TestBuildBenchmarkArgs() {
	pattern := "BenchmarkExample"
	args := ts.builder.BuildBenchmarkArgs(pattern)

	expected := []string{
		"test",
		"-bench", "BenchmarkExample",
		"-run", "^$",
		"-cpu", "2",
		"-benchtime", "10s",
		"-benchmem",
	}

	ts.Require().Equal(expected, args)
}

func (ts *TestCommandBuilderTestSuite) TestBuildBenchmarkArgsMinimal() {
	config := mockTestConfigProvider{
		test: mockTestConfig{
			benchCPU:  0,
			benchTime: "",
			benchMem:  false,
		},
	}

	builder := builders.NewTestCommandBuilder(config)
	pattern := "BenchmarkExample"
	args := builder.BuildBenchmarkArgs(pattern)

	expected := []string{
		"test",
		"-bench", "BenchmarkExample",
		"-run", "^$",
	}

	ts.Require().Equal(expected, args)
}

func (ts *TestCommandBuilderTestSuite) TestGetCoverageExcludePackages() {
	excludes := ts.builder.GetCoverageExcludePackages()

	// Check default excludes
	ts.Require().Contains(excludes, "testdata")
	ts.Require().Contains(excludes, "vendor")
	ts.Require().Contains(excludes, "examples")
	ts.Require().Contains(excludes, "cmd")
	ts.Require().Contains(excludes, "tools")
	ts.Require().Contains(excludes, "mocks")
	ts.Require().Contains(excludes, "testutil")

	// Check custom exclude from config
	ts.Require().Contains(excludes, "excluded")
}

func (ts *TestCommandBuilderTestSuite) TestBuildBaseTestArgsWithEnvironmentVerbose() {
	// Test with MAGE_X_VERBOSE environment variable
	err := os.Setenv("MAGE_X_VERBOSE", "true")
	ts.Require().NoError(err)
	defer func() {
		err := os.Unsetenv("MAGE_X_VERBOSE")
		ts.Require().NoError(err)
	}()

	config := mockTestConfigProvider{
		test: mockTestConfig{
			timeout:  "5m",
			parallel: 2,
		},
		build: mockTestBuildConfig{
			verbose: false, // Config verbose is false, but env should override
		},
	}

	builder := builders.NewTestCommandBuilder(config)
	options := builders.TestOptions{}
	args := builder.BuildUnitTestArgs(options)

	ts.Require().Contains(args, "-v")
}

func (ts *TestCommandBuilderTestSuite) TestBuildBaseTestArgsWithTestVerbose() {
	// Test with MAGE_X_TEST_VERBOSE environment variable
	err := os.Setenv("MAGE_X_TEST_VERBOSE", "true")
	ts.Require().NoError(err)
	defer func() {
		err := os.Unsetenv("MAGE_X_TEST_VERBOSE")
		ts.Require().NoError(err)
	}()

	config := mockTestConfigProvider{
		test: mockTestConfig{
			timeout:  "5m",
			parallel: 2,
		},
		build: mockTestBuildConfig{
			verbose: false, // Config verbose is false, but env should override
		},
	}

	builder := builders.NewTestCommandBuilder(config)
	options := builders.TestOptions{}
	args := builder.BuildUnitTestArgs(options)

	ts.Require().Contains(args, "-v")
}

func (ts *TestCommandBuilderTestSuite) TestBuildBaseTestArgsWithTestFlags() {
	// Test with TEST_FLAGS environment variable
	err := os.Setenv("TEST_FLAGS", "-count=1 -failfast")
	ts.Require().NoError(err)
	defer func() {
		err := os.Unsetenv("TEST_FLAGS")
		ts.Require().NoError(err)
	}()

	options := builders.TestOptions{}
	args := ts.builder.BuildUnitTestArgs(options)

	ts.Require().Contains(args, "-count=1")
	ts.Require().Contains(args, "-failfast")
}

func (ts *TestCommandBuilderTestSuite) TestBuildBaseTestArgsEmptyConfig() {
	// Test with minimal configuration
	config := mockTestConfigProvider{
		test: mockTestConfig{
			timeout:  "",
			parallel: 0,
			tags:     "",
			shuffle:  false,
		},
		build: mockTestBuildConfig{
			verbose: false,
		},
	}

	builder := builders.NewTestCommandBuilder(config)
	options := builders.TestOptions{}
	args := builder.BuildUnitTestArgs(options)

	expected := []string{
		"test",
		"./...",
	}

	ts.Require().Equal(expected, args)
}

func (ts *TestCommandBuilderTestSuite) TestBuildIntegrationTestArgsNoTag() {
	// Test when integration tag is not configured
	config := mockTestConfigProvider{
		test: mockTestConfig{
			timeout:            "10m",
			integrationTimeout: "30m",
			integrationTag:     "", // No integration tag
			parallel:           2,
		},
		build: mockTestBuildConfig{
			verbose: false,
		},
	}

	builder := builders.NewTestCommandBuilder(config)
	options := builders.TestOptions{}
	args := builder.BuildIntegrationTestArgs(options)

	// Should not contain -tags integration since it's not configured
	tagCount := 0
	for _, arg := range args {
		if arg == "-tags" {
			tagCount++
		}
	}

	// Should have no -tags flags since integration tag is empty and no other tags configured
	ts.Require().Equal(0, tagCount)
}

func TestTestCommandBuilderTestSuite(t *testing.T) {
	suite.Run(t, new(TestCommandBuilderTestSuite))
}
