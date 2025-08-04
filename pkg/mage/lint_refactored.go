package mage

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/mage/builders"
	"github.com/mrz1836/mage-x/pkg/mage/operations"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for linting operations
var (
	errVetFailed     = errors.New("go vet found issues")
	errLintingFailed = errors.New("linting failed")
	errFixFailed     = errors.New("fix failed")
)

// Lint namespace for linting and formatting tasks
type Lint mg.Namespace

// Default runs the default linter (golangci-lint + go vet)
func (Lint) Default() error {
	ctx, err := operations.NewOperation("Running Default Linters")
	if err != nil {
		return err
	}

	// Display linter configuration info
	displayLinterConfig()

	// Ensure golangci-lint is installed
	ctx.Info("Checking golangci-lint installation...")
	if err := ensureGolangciLint(ctx.Config()); err != nil {
		return ctx.Complete(err)
	}

	// Create lint operation
	lintOp := &LintOperation{
		config:  ctx.Config(),
		builder: builders.NewLintCommandBuilder(ctx.Config()),
		runVet:  true,
	}

	// Run linters on all modules
	runner := operations.NewModuleRunner(ctx.Config(), lintOp)
	return ctx.Complete(runner.RunForAllModules())
}

// LintOperation implements ModuleOperation for linting
type LintOperation struct {
	config  *Config
	builder *builders.LintCommandBuilder
	runVet  bool
	fix     bool
	format  string
}

func (l *LintOperation) Name() string {
	if l.fix {
		return "Lint fix"
	}
	return "Linting"
}

func (l *LintOperation) Execute(module Module, config *Config) error {
	var hasError bool

	// Run golangci-lint
	utils.Info("Running golangci-lint...")
	args := l.builder.BuildGolangciArgs(module, builders.LintOptions{
		Fix:    l.fix,
		Format: l.format,
	})

	if err := RunCommandInModule(module, "golangci-lint", args...); err != nil {
		hasError = true
		if !l.fix {
			utils.Error("golangci-lint found issues")
		}
	}

	// Run go vet if requested
	if l.runVet && !l.fix {
		utils.Info("Running go vet...")
		vetArgs := l.builder.BuildVetArgs()
		if err := RunCommandInModule(module, "go", vetArgs...); err != nil {
			hasError = true
			utils.Error("go vet found issues")
		}
	}

	if hasError && !l.fix {
		return errLintingFailed
	}

	return nil
}

// Fix automatically fixes linting issues
func (Lint) Fix() error {
	ctx, err := operations.NewOperation("Running Linter with Auto-Fix")
	if err != nil {
		return err
	}

	// Ensure golangci-lint is installed
	if err := ensureGolangciLint(ctx.Config()); err != nil {
		return ctx.Complete(err)
	}

	// Create fix operation
	fixOp := &LintOperation{
		config:  ctx.Config(),
		builder: builders.NewLintCommandBuilder(ctx.Config()),
		fix:     true,
	}

	// Run fix on all modules
	runner := operations.NewModuleRunner(ctx.Config(), fixOp)
	if err := runner.RunForAllModules(); err != nil {
		ctx.Warn("Some fixes may have failed")
	}

	ctx.Success("Auto-fix completed. Review changes before committing.")
	return ctx.Complete(nil)
}

// Vet runs go vet static analysis
func (Lint) Vet() error {
	ctx, err := operations.NewOperation("Running go vet")
	if err != nil {
		return err
	}

	// Create vet-only operation
	vetOp := &VetOperation{
		builder: builders.NewLintCommandBuilder(ctx.Config()),
	}

	// Run vet on all modules
	runner := operations.NewModuleRunner(ctx.Config(), vetOp)
	return ctx.Complete(runner.RunForAllModules())
}

// VetOperation implements ModuleOperation for go vet
type VetOperation struct {
	builder *builders.LintCommandBuilder
}

func (v *VetOperation) Name() string {
	return "Go vet"
}

func (v *VetOperation) Execute(module Module, config *Config) error {
	args := v.builder.BuildVetArgs()
	return RunCommandInModule(module, "go", args...)
}

// All runs all available linters
func (Lint) All() error {
	ctx, err := operations.NewOperation("Running All Linters")
	if err != nil {
		return err
	}

	// Run default linters first
	if err := Lint{}.Default(); err != nil {
		return ctx.Complete(err)
	}

	// Run additional linters if available
	if utils.CommandExists("staticcheck") {
		ctx.Info("\nRunning staticcheck...")
		if err := Lint{}.Staticcheck(); err != nil {
			ctx.Warn("staticcheck found issues: %v", err)
		}
	}

	if utils.CommandExists("revive") {
		ctx.Info("\nRunning revive...")
		if err := Lint{}.Revive(); err != nil {
			ctx.Warn("revive found issues: %v", err)
		}
	}

	return ctx.Complete(nil)
}

// Staticcheck runs the staticcheck linter
func (Lint) Staticcheck() error {
	ctx, err := operations.NewOperation("Running staticcheck")
	if err != nil {
		return err
	}

	// Check if staticcheck is installed
	if !utils.CommandExists("staticcheck") {
		ctx.Info("Installing staticcheck...")
		if err := utils.RunCommand("go", "install", "honnef.co/go/tools/cmd/staticcheck@latest"); err != nil {
			return ctx.Complete(fmt.Errorf("failed to install staticcheck: %w", err))
		}
	}

	// Create staticcheck operation
	staticOp := &StaticcheckOperation{
		builder: builders.NewLintCommandBuilder(ctx.Config()),
	}

	// Run on all modules
	runner := operations.NewModuleRunner(ctx.Config(), staticOp)
	return ctx.Complete(runner.RunForAllModules())
}

// StaticcheckOperation implements ModuleOperation for staticcheck
type StaticcheckOperation struct {
	builder *builders.LintCommandBuilder
}

func (s *StaticcheckOperation) Name() string {
	return "Staticcheck"
}

func (s *StaticcheckOperation) Execute(module Module, config *Config) error {
	args := s.builder.BuildStaticcheckArgs()
	return RunCommandInModule(module, "staticcheck", args...)
}

// Revive runs the revive linter
func (Lint) Revive() error {
	ctx, err := operations.NewOperation("Running revive")
	if err != nil {
		return err
	}

	// Check if revive is installed
	if !utils.CommandExists("revive") {
		ctx.Info("Installing revive...")
		if err := utils.RunCommand("go", "install", "github.com/mgechev/revive@latest"); err != nil {
			return ctx.Complete(fmt.Errorf("failed to install revive: %w", err))
		}
	}

	// Create revive operation
	reviveOp := &ReviveOperation{}

	// Run on all modules
	runner := operations.NewModuleRunner(ctx.Config(), reviveOp)
	return ctx.Complete(runner.RunForAllModules())
}

// ReviveOperation implements ModuleOperation for revive
type ReviveOperation struct{}

func (r *ReviveOperation) Name() string {
	return "Revive"
}

func (r *ReviveOperation) Execute(module Module, config *Config) error {
	// Check for revive config
	configPath := ".revive.toml"
	args := []string{}
	
	if utils.FileExists(configPath) {
		args = append(args, "-config", configPath)
	}
	
	args = append(args, "./...")
	
	return RunCommandInModule(module, "revive", args...)
}

// Version shows golangci-lint version information
func (Lint) Version() error {
	ctx, err := operations.NewOperation("Linter Version Information")
	if err != nil {
		return err
	}

	// Check golangci-lint version
	if utils.CommandExists("golangci-lint") {
		output, err := utils.RunCommandOutput("golangci-lint", "version")
		if err == nil {
			ctx.Info("golangci-lint version:\n%s", output)
		}
	} else {
		ctx.Warn("golangci-lint is not installed")
	}

	// Check go version
	output, err := utils.RunCommandOutput("go", "version")
	if err == nil {
		ctx.Info("\nGo version: %s", strings.TrimSpace(output))
	}

	// Check other linters if available
	if utils.CommandExists("staticcheck") {
		output, err := utils.RunCommandOutput("staticcheck", "-version")
		if err == nil {
			ctx.Info("\nstaticcheck version: %s", strings.TrimSpace(output))
		}
	}

	if utils.CommandExists("revive") {
		ctx.Info("\nrevive is installed")
	}

	return ctx.Complete(nil)
}

// Fumpt runs gofumpt for stricter formatting
func (Lint) Fumpt() error {
	ctx, err := operations.NewOperation("Running gofumpt")
	if err != nil {
		return err
	}

	// Ensure gofumpt is installed
	if !utils.CommandExists("gofumpt") {
		ctx.Info("Installing gofumpt...")
		if err := utils.RunCommand("go", "install", "mvdan.cc/gofumpt@latest"); err != nil {
			return ctx.Complete(fmt.Errorf("failed to install gofumpt: %w", err))
		}
	}

	// Create fumpt operation
	fumptOp := &FumptOperation{
		builder: builders.NewLintCommandBuilder(ctx.Config()),
	}

	// Run on all modules
	runner := operations.NewModuleRunner(ctx.Config(), fumptOp)
	return ctx.Complete(runner.RunForAllModules())
}

// FumptOperation implements ModuleOperation for gofumpt
type FumptOperation struct {
	builder *builders.LintCommandBuilder
}

func (f *FumptOperation) Name() string {
	return "Gofumpt formatting"
}

func (f *FumptOperation) Execute(module Module, config *Config) error {
	args := f.builder.BuildGofumptArgs(true) // extra rules enabled
	return RunCommandInModule(module, "gofumpt", args...)
}

// Helper functions

func displayLinterConfig() {
	utils.Info("Linter Configuration:")
	utils.Info("  - Config: .golangci.json")
	utils.Info("  - Parallel: %d CPUs", runtime.NumCPU())
	
	if verbose := os.Getenv("VERBOSE"); verbose == "true" {
		utils.Info("  - Verbose: enabled")
	}
}

func ensureGolangciLint(config *Config) error {
	if utils.CommandExists("golangci-lint") {
		return nil
	}

	utils.Info("golangci-lint not found. Installing...")
	
	// Try to install using go install
	version := config.Lint.GolangciVersion
	if version == "" {
		version = "latest"
	}
	
	installCmd := fmt.Sprintf("github.com/golangci/golangci-lint/cmd/golangci-lint@%s", version)
	if err := utils.RunCommand("go", "install", installCmd); err != nil {
		return fmt.Errorf("failed to install golangci-lint: %w", err)
	}
	
	utils.Success("golangci-lint installed successfully")
	return nil
}