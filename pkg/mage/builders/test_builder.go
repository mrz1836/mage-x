package builders

import (
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/mrz1836/mage-x/pkg/common/env"
)

// getMageXEnv returns the value of a MAGE-X environment variable with the proper prefix
func getMageXEnv(suffix string) string {
	return os.Getenv("MAGE_X_" + suffix)
}

// TestConfig interface provides access to test configuration
type TestConfig interface {
	GetTimeout() string
	GetIntegrationTimeout() string
	GetIntegrationTag() string
	GetCoverMode() string
	GetParallel() int
	GetTags() string
	GetShuffle() bool
	GetBenchCPU() int
	GetBenchTime() string
	GetBenchMem() bool
	GetCoverageExclude() []string
}

// TestBuildConfig interface provides access to build configuration for tests
type TestBuildConfig interface {
	GetVerbose() bool
}

// TestConfigProvider interface provides access to configuration needed by test builders
type TestConfigProvider interface {
	GetTest() TestConfig
	GetBuild() TestBuildConfig
}

// TestOptions contains options for test commands
type TestOptions struct {
	Coverage     bool
	Race         bool
	Short        bool
	Integration  bool
	CoverageFile string
	CoverageMode string
	Pattern      string
}

// TestCommandBuilder builds test-related commands
type TestCommandBuilder struct {
	config TestConfigProvider
}

// NewTestCommandBuilder creates a new test command builder
func NewTestCommandBuilder(config TestConfigProvider) *TestCommandBuilder {
	return &TestCommandBuilder{config: config}
}

// BuildUnitTestArgs builds arguments for unit tests
func (b *TestCommandBuilder) BuildUnitTestArgs(options TestOptions) []string {
	args := b.buildBaseTestArgs(options)

	// Add short flag for unit tests
	if options.Short {
		args = append(args, "-short")
	}

	// Add pattern or default to all packages
	if options.Pattern != "" {
		args = append(args, options.Pattern)
	} else {
		args = append(args, "./...")
	}

	return args
}

// BuildIntegrationTestArgs builds arguments for integration tests
func (b *TestCommandBuilder) BuildIntegrationTestArgs(options TestOptions) []string {
	args := b.buildBaseTestArgs(options)

	// Integration tests typically run longer
	args = append(args, "-timeout", b.config.GetTest().GetIntegrationTimeout())

	// Add integration tag if configured
	if tag := b.config.GetTest().GetIntegrationTag(); tag != "" {
		args = append(args, "-tags", tag)
	}

	// Add pattern or default to all packages
	if options.Pattern != "" {
		args = append(args, options.Pattern)
	} else {
		args = append(args, "./...")
	}

	return args
}

// BuildCoverageArgs builds arguments for coverage tests
func (b *TestCommandBuilder) BuildCoverageArgs(options TestOptions) []string {
	args := b.buildBaseTestArgs(options)

	// Coverage specific flags
	args = append(args, "-cover")

	if options.CoverageFile != "" {
		args = append(args, "-coverprofile="+options.CoverageFile)
	}

	if options.CoverageMode != "" {
		args = append(args, "-covermode="+options.CoverageMode)
	} else if mode := b.config.GetTest().GetCoverMode(); mode != "" {
		args = append(args, "-covermode="+mode)
	}

	// Add pattern or default to all packages
	if options.Pattern != "" {
		args = append(args, options.Pattern)
	} else {
		args = append(args, "./...")
	}

	return args
}

// BuildBenchmarkArgs builds arguments for benchmarks
func (b *TestCommandBuilder) BuildBenchmarkArgs(pattern string) []string {
	args := []string{"test", "-bench", pattern}

	// Run tests in benchmark mode
	args = append(args, "-run", "^$")

	// Add CPU count if configured
	if cpu := b.config.GetTest().GetBenchCPU(); cpu > 0 {
		args = append(args, "-cpu", strconv.Itoa(cpu))
	}

	// Add benchmark time if configured
	if benchTime := b.config.GetTest().GetBenchTime(); benchTime != "" {
		args = append(args, "-benchtime", benchTime)
	}

	// Add benchmark memory flag if configured
	if b.config.GetTest().GetBenchMem() {
		args = append(args, "-benchmem")
	}

	return args
}

// buildBaseTestArgs builds common test arguments
func (b *TestCommandBuilder) buildBaseTestArgs(options TestOptions) []string {
	args := []string{"test"}

	// Add verbose flag
	if b.config.GetBuild().GetVerbose() || env.GetString("MAGE_X_VERBOSE", "") == "true" || env.GetString("MAGE_X_TEST_VERBOSE", "") == "true" {
		args = append(args, "-v")
	}

	// Add timeout
	if timeout := b.config.GetTest().GetTimeout(); timeout != "" {
		args = append(args, "-timeout", timeout)
	}

	// Add race detection
	if options.Race && runtime.GOOS != "windows" {
		args = append(args, "-race")
	}

	// Add parallel flag
	if parallel := b.config.GetTest().GetParallel(); parallel > 0 {
		args = append(args, "-parallel", strconv.Itoa(parallel))
	}

	// Add custom test flags from environment
	if customFlags := getMageXEnv("TEST_FLAGS"); customFlags != "" {
		args = append(args, strings.Fields(customFlags)...)
	}

	// Add build tags
	if tags := b.config.GetTest().GetTags(); tags != "" {
		args = append(args, "-tags", tags)
	}

	// Add test shuffle
	if b.config.GetTest().GetShuffle() && !options.Coverage {
		args = append(args, "-shuffle=on")
	}

	return args
}

// GetCoverageExcludePackages returns packages to exclude from coverage
func (b *TestCommandBuilder) GetCoverageExcludePackages() []string {
	excludes := []string{
		"testdata",
		"vendor",
		"examples",
		"cmd",
		"tools",
		"mocks",
		"testutil",
	}

	// Add custom excludes from config
	if exclude := b.config.GetTest().GetCoverageExclude(); len(exclude) > 0 {
		excludes = append(excludes, exclude...)
	}

	return excludes
}
