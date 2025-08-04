package builders

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// LintOptions contains options for lint commands
type LintOptions struct {
	Fix      bool
	Format   string
	Output   string
	NoConfig bool
}

// LintCommandBuilder builds lint-related commands
type LintCommandBuilder struct {
	config *mage.Config
}

// NewLintCommandBuilder creates a new lint command builder
func NewLintCommandBuilder(config *mage.Config) *LintCommandBuilder {
	return &LintCommandBuilder{config: config}
}

// BuildGolangciArgs builds arguments for golangci-lint
func (b *LintCommandBuilder) BuildGolangciArgs(module mage.Module, options LintOptions) []string {
	args := []string{"run"}

	// Check for config file
	if !options.NoConfig {
		configPath := b.findLintConfig(module)
		if configPath != "" {
			args = append(args, "--config", configPath)
		}
	}

	// Add timeout
	if b.config.Lint.Timeout != "" {
		args = append(args, "--timeout", b.config.Lint.Timeout)
	}

	// Add verbose flag
	if b.config.Build.Verbose {
		args = append(args, "--verbose")
	}

	// Add fix flag
	if options.Fix {
		args = append(args, "--fix")
	}

	// Add output format
	if options.Format != "" {
		args = append(args, "--out-format", options.Format)
	}

	// Add output file
	if options.Output != "" {
		args = append(args, "--out", options.Output)
	}

	// Add concurrency based on build parallel setting
	if b.config.Build.Parallel > 0 {
		args = append(args, "--concurrency", fmt.Sprintf("%d", b.config.Build.Parallel))
	}

	// Add build tags
	if len(b.config.Build.Tags) > 0 {
		args = append(args, "--build-tags", strings.Join(b.config.Build.Tags, ","))
	}

	// Always lint all packages in the module
	args = append(args, "./...")

	return args
}

// BuildVetArgs builds arguments for go vet
func (b *LintCommandBuilder) BuildVetArgs() []string {
	args := []string{"vet"}

	// Add build tags
	if len(b.config.Build.Tags) > 0 {
		args = append(args, "-tags", strings.Join(b.config.Build.Tags, ","))
	}

	// Add verbose output if requested
	if b.config.Build.Verbose {
		args = append(args, "-v")
	}

	// Vet all packages
	args = append(args, "./...")

	return args
}

// BuildStaticcheckArgs builds arguments for staticcheck
func (b *LintCommandBuilder) BuildStaticcheckArgs() []string {
	args := []string{}

	// Add format flag
	args = append(args, "-f", "text")

	// Add build tags
	if len(b.config.Build.Tags) > 0 {
		args = append(args, "-tags", strings.Join(b.config.Build.Tags, ","))
	}

	// Check all packages
	args = append(args, "./...")

	return args
}

// BuildGofmtArgs builds arguments for gofmt
func (b *LintCommandBuilder) BuildGofmtArgs(checkOnly bool) []string {
	args := []string{}

	if checkOnly {
		// List files that need formatting
		args = append(args, "-l")
	} else {
		// Write formatted files
		args = append(args, "-w")
	}

	// Check current directory
	args = append(args, ".")

	return args
}

// BuildGofumptArgs builds arguments for gofumpt
func (b *LintCommandBuilder) BuildGofumptArgs(extra bool) []string {
	args := []string{"-w"}

	if extra {
		args = append(args, "-extra")
	}

	// Format current directory
	args = append(args, ".")

	return args
}

// BuildGoimportsArgs builds arguments for goimports
func (b *LintCommandBuilder) BuildGoimportsArgs() []string {
	args := []string{"-w"}

	// Add local prefix based on module name
	if b.config.Project.Module != "" {
		args = append(args, "-local", b.config.Project.Module)
	}

	// Format current directory
	args = append(args, ".")

	return args
}

// findLintConfig finds the appropriate lint config file
func (b *LintCommandBuilder) findLintConfig(module mage.Module) string {
	// Check for config file in module directory
	configPath := filepath.Join(module.Path, ".golangci.json")
	if utils.FileExists(configPath) {
		return configPath
	}

	// Check for config file in root directory
	rootConfig := ".golangci.json"
	if utils.FileExists(rootConfig) {
		// Use absolute path to root config
		absPath, err := filepath.Abs(rootConfig)
		if err != nil {
			return rootConfig
		}
		return absPath
	}

	// Check for YAML config as fallback
	configPath = filepath.Join(module.Path, ".golangci.yml")
	if utils.FileExists(configPath) {
		return configPath
	}

	rootConfig = ".golangci.yml"
	if utils.FileExists(rootConfig) {
		absPath, err := filepath.Abs(rootConfig)
		if err != nil {
			return rootConfig
		}
		return absPath
	}

	return ""
}
