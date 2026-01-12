// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/magefile/mage/mg"
	"golang.org/x/sync/errgroup"

	"github.com/mrz1836/mage-x/pkg/common/env"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors to satisfy err113 linter
var (
	errGoVetFailed        = errors.New("go vet failed")
	errStrictChecksFailed = errors.New("strict checks failed")
	// ErrVetPanic indicates a panic occurred during vet operation.
	// Use errors.Is(err, ErrVetPanic) to check for this condition.
	ErrVetPanic = errors.New("panic during vet operation")
)

// ShadowToolManager manages shadow tool installation and ensures thread-safety
type ShadowToolManager interface {
	EnsureShadowTool() error
}

// shadowToolManager implements ShadowToolManager with thread-safe shadow tool installation
type shadowToolManager struct {
	mutex sync.Mutex
}

// NewShadowToolManager creates a new shadow tool manager
func NewShadowToolManager() ShadowToolManager {
	return &shadowToolManager{}
}

// GetDefaultShadowToolManager returns a default shadow tool manager instance
func GetDefaultShadowToolManager() ShadowToolManager {
	return NewShadowToolManager()
}

// Vet namespace for go vet tasks
type Vet mg.Namespace

// Default runs go vet on module packages only
func (Vet) Default() error {
	utils.Header("Running go vet")

	// Get module name
	module, err := utils.GetModuleName()
	if err != nil {
		return fmt.Errorf("failed to get module name: %w", err)
	}

	utils.Info("Checking module: %s", module)

	// Get all packages in this module
	output, err := GetRunner().RunCmdOutput("go", "list", "./...")
	if err != nil {
		return fmt.Errorf("failed to list packages: %w", err)
	}

	packages := strings.Split(strings.TrimSpace(output), "\n")
	modulePackages := make([]string, 0, len(packages)) // Pre-allocate for better performance

	// Filter to only module packages
	for _, pkg := range packages {
		if strings.HasPrefix(pkg, module) {
			modulePackages = append(modulePackages, pkg)
		}
	}

	if len(modulePackages) == 0 {
		utils.Warn("No packages found in module")
		return nil
	}

	utils.Info("Vetting %d packages", len(modulePackages))

	// Build vet arguments
	args := []string{"vet"}

	// Add verbose flag if requested
	if verbose := env.GetString("VERBOSE", ""); verbose == trueValue {
		args = append(args, "-v")
	}

	// Add build tags if specified
	if tags := env.GetString("MAGE_X_BUILD_TAGS", ""); tags != "" {
		args = append(args, "-tags", tags)
	}

	// Add all module packages
	args = append(args, modulePackages...)

	start := time.Now()
	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("go vet found issues: %w", err)
	}

	utils.Success("go vet passed in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// All runs go vet on all packages (including dependencies)
func (Vet) All() error {
	utils.Header("Running go vet (All Packages)")

	args := []string{"vet"}

	// Add verbose flag if requested
	if verbose := env.GetString("VERBOSE", ""); verbose == trueValue {
		args = append(args, "-v")
	}

	// Add build tags if specified
	if tags := env.GetString("MAGE_X_BUILD_TAGS", ""); tags != "" {
		args = append(args, "-tags", tags)
	}

	args = append(args, "./...")

	start := time.Now()
	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("go vet found issues: %w", err)
	}

	utils.Success("go vet passed in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// Parallel runs go vet in parallel (faster for large repos)
// Uses errgroup for proper goroutine lifecycle management and error handling
func (Vet) Parallel() error {
	utils.Header("Running go vet in Parallel")

	// Get module name
	module, err := utils.GetModuleName()
	if err != nil {
		return fmt.Errorf("failed to get module name: %w", err)
	}

	// Get all packages in this module
	output, err := GetRunner().RunCmdOutput("go", "list", "./...")
	if err != nil {
		return fmt.Errorf("failed to list packages: %w", err)
	}

	packages := strings.Split(strings.TrimSpace(output), "\n")
	modulePackages := make([]string, 0, len(packages)) // Pre-allocate for better performance

	// Filter to only module packages
	for _, pkg := range packages {
		if strings.HasPrefix(pkg, module) {
			modulePackages = append(modulePackages, pkg)
		}
	}

	if len(modulePackages) == 0 {
		utils.Warn("No packages found in module")
		return nil
	}

	// Get CPU count for parallel execution
	parallel := getCPUCount()
	utils.Info("Vetting %d packages with %d workers", len(modulePackages), parallel)

	start := time.Now()

	// Use errgroup for proper goroutine lifecycle management
	g, ctx := errgroup.WithContext(context.Background())
	g.SetLimit(parallel)

	// Mutex to protect vetErrors slice
	var mu sync.Mutex
	var vetErrors []error

	// Vet packages in parallel using errgroup
	for _, pkg := range modulePackages {
		// Capture loop variable
		g.Go(func() error {
			defer func() {
				if p := recover(); p != nil {
					mu.Lock()
					defer mu.Unlock()
					vetErrors = append(vetErrors, fmt.Errorf("%w: package %s: %v", ErrVetPanic, pkg, p))
				}
			}()

			// Check for context cancellation
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			args := []string{"vet"}

			// Add build tags if specified
			if tags := env.GetString("MAGE_X_BUILD_TAGS", ""); tags != "" {
				args = append(args, "-tags", tags)
			}

			args = append(args, pkg)

			if runErr := GetRunner().RunCmd("go", args...); runErr != nil {
				mu.Lock()
				defer mu.Unlock()
				vetErrors = append(vetErrors, fmt.Errorf("vet failed for %s: %w", pkg, runErr))
			}
			return nil // Don't return error to allow all packages to be vetted
		})
	}

	// Wait for all goroutines to complete
	// Errors are collected in vetErrors, goroutines return nil to allow all packages to complete
	//nolint:errcheck,gosec // Errors collected in vetErrors slice, not returned from Wait
	g.Wait()

	if len(vetErrors) > 0 {
		utils.Error("go vet found issues in %d packages:", len(vetErrors))
		for _, vetErr := range vetErrors {
			fmt.Printf("  - %v\n", vetErr)
		}
		return errGoVetFailed
	}

	utils.Success("go vet passed in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// Shadow checks for shadowed variables
func (Vet) Shadow() error {
	utils.Header("Checking for Shadowed Variables")

	// Check if shadow tool exists, and install if needed
	if err := GetDefaultShadowToolManager().EnsureShadowTool(); err != nil {
		return fmt.Errorf("failed to ensure shadow tool is available: %w", err)
	}

	// Determine shadow command path
	shadowCmd := "shadow"
	if gopath, err := GetRunner().RunCmdOutput("go", "env", "GOPATH"); err == nil {
		shadowPath := strings.TrimSpace(gopath) + "/bin/shadow"
		if utils.FileExists(shadowPath) {
			shadowCmd = shadowPath
		}
	}

	// Build shadow command args - exclude test files to reduce noise
	args := []string{"-test=false", "./..."}

	// Add build tags if specified
	if tags := env.GetString("MAGE_X_BUILD_TAGS", ""); tags != "" {
		args = append([]string{"-tags", tags}, args...)
	}

	// Run shadow directly
	utils.Info("Running shadow check...")
	if err := GetRunner().RunCmd(shadowCmd, args...); err != nil {
		return fmt.Errorf("shadow check found issues: %w", err)
	}

	utils.Success("No shadowed variables found")
	return nil
}

// EnsureShadowTool ensures the shadow tool is installed and available
func (stm *shadowToolManager) EnsureShadowTool() error {
	// Use mutex to prevent concurrent installations
	stm.mutex.Lock()
	defer stm.mutex.Unlock()

	// Try using go env GOPATH to check if shadow is in GOPATH/bin first
	if gopath, err := GetRunner().RunCmdOutput("go", "env", "GOPATH"); err == nil {
		shadowPath := strings.TrimSpace(gopath) + "/bin/shadow"
		if utils.FileExists(shadowPath) {
			return nil // Tool is available in GOPATH/bin
		}
	}

	// Check if shadow tool is already available by trying to run it
	if err := GetRunner().RunCmd("which", "shadow"); err == nil {
		return nil // Tool is already available
	}

	// Install shadow tool
	utils.Info("Installing shadow analyzer...")
	return GetRunner().RunCmd("go", "install", "golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest")
}

// Strict runs go vet with additional strict checks
func (Vet) Strict() error {
	utils.Header("Running Strict go vet")

	// Run multiple vet-like tools
	checks := []struct {
		name string
		fn   func() error
	}{
		{"Standard vet", Vet{}.Default},
		{"Shadow check", Vet{}.Shadow},
		{"Staticcheck", runStaticcheck},
		{"Ineffassign", runIneffassign},
		{"Misspell", runMisspell},
	}

	failed := 0
	for _, check := range checks {
		utils.Info("\nRunning %s...", check.name)
		if err := check.fn(); err != nil {
			utils.Error("Failed: %v", err)
			failed++
		}
	}

	if failed > 0 {
		return fmt.Errorf("%w: %d", errStrictChecksFailed, failed)
	}

	utils.Success("All strict checks passed!")
	return nil
}

// Helper functions

// runStaticcheck runs the staticcheck tool
func runStaticcheck() error {
	// Ensure tool is installed using consolidated helper
	if err := installTool(ToolDefinition{
		Name:   "staticcheck",
		Module: "honnef.co/go/tools/cmd/staticcheck",
		Check:  "staticcheck",
	}); err != nil {
		return fmt.Errorf("failed to install staticcheck tool: %w", err)
	}

	version := getLinterVersion("staticcheck")
	utils.Info("Running staticcheck %s...", version)
	if err := GetRunner().RunCmd("staticcheck", "./..."); err != nil {
		return fmt.Errorf("staticcheck found issues: %w", err)
	}

	utils.Success("staticcheck %s passed", version)
	return nil
}

// runIneffassign checks for ineffectual assignments
func runIneffassign() error {
	// Ensure tool is installed using consolidated helper
	if err := installTool(ToolDefinition{
		Name:   "ineffassign",
		Module: "github.com/gordonklaus/ineffassign",
		Check:  "ineffassign",
	}); err != nil {
		return fmt.Errorf("failed to install ineffassign tool: %w", err)
	}

	version := getLinterVersion("ineffassign")
	utils.Info("Running ineffassign %s...", version)
	if err := GetRunner().RunCmd("ineffassign", "./..."); err != nil {
		return fmt.Errorf("ineffassign found issues: %w", err)
	}

	utils.Success("ineffassign %s passed", version)
	return nil
}

// runMisspell checks for misspelled words
func runMisspell() error {
	// Ensure tool is installed using consolidated helper
	if err := installTool(ToolDefinition{
		Name:   "misspell",
		Module: "github.com/client9/misspell/cmd/misspell",
		Check:  "misspell",
	}); err != nil {
		return fmt.Errorf("failed to install misspell tool: %w", err)
	}

	version := getLinterVersion("misspell")
	utils.Info("Running misspell %s...", version)
	if err := GetRunner().RunCmd("misspell", "-w", "."); err != nil {
		return fmt.Errorf("misspell found issues: %w", err)
	}

	utils.Success("misspell %s passed", version)
	return nil
}
