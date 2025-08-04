package builders

import (
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/utils"
)

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
	config *mage.Config
}

// NewTestCommandBuilder creates a new test command builder
func NewTestCommandBuilder(config *mage.Config) *TestCommandBuilder {
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
	args = append(args, "-timeout", b.config.Test.IntegrationTimeout)

	// Add integration tag if configured
	if b.config.Test.IntegrationTag != "" {
		args = append(args, "-tags", b.config.Test.IntegrationTag)
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
	} else if b.config.Test.CoverageMode != "" {
		args = append(args, "-covermode="+b.config.Test.CoverageMode)
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
	if b.config.Test.BenchCPU > 0 {
		args = append(args, "-cpu", strconv.Itoa(b.config.Test.BenchCPU))
	}

	// Add benchmark time if configured
	if b.config.Test.BenchTime != "" {
		args = append(args, "-benchtime", b.config.Test.BenchTime)
	}

	// Add benchmark memory flag if configured
	if b.config.Test.BenchMem {
		args = append(args, "-benchmem")
	}

	return args
}

// buildBaseTestArgs builds common test arguments
func (b *TestCommandBuilder) buildBaseTestArgs(options TestOptions) []string {
	args := []string{"test"}

	// Add verbose flag
	if b.config.Build.Verbose || utils.GetEnv("VERBOSE", "") == "true" || utils.GetEnv("TEST_VERBOSE", "") == "true" {
		args = append(args, "-v")
	}

	// Add timeout
	if b.config.Test.Timeout != "" {
		args = append(args, "-timeout", b.config.Test.Timeout)
	}

	// Add race detection
	if options.Race && runtime.GOOS != "windows" {
		args = append(args, "-race")
	}

	// Add parallel flag
	if b.config.Test.Parallel > 0 {
		args = append(args, "-parallel", strconv.Itoa(b.config.Test.Parallel))
	}

	// Add custom test flags from environment
	if customFlags := os.Getenv("TEST_FLAGS"); customFlags != "" {
		args = append(args, strings.Fields(customFlags)...)
	}

	// Add build tags
	if b.config.Test.Tags != "" {
		args = append(args, "-tags", b.config.Test.Tags)
	}

	// Add test shuffle
	if b.config.Test.Shuffle && !options.Coverage {
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
	if len(b.config.Test.CoverageExclude) > 0 {
		excludes = append(excludes, b.config.Test.CoverageExclude...)
	}

	return excludes
}
