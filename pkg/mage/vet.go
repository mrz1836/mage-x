// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"fmt"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/utils"
)

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
	modulePackages := []string{}

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
	if verbose := utils.GetEnv("VERBOSE", ""); verbose == "true" {
		args = append(args, "-v")
	}

	// Add build tags if specified
	if tags := utils.GetEnv("GO_BUILD_TAGS", ""); tags != "" {
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
	if verbose := utils.GetEnv("VERBOSE", ""); verbose == "true" {
		args = append(args, "-v")
	}

	// Add build tags if specified
	if tags := utils.GetEnv("GO_BUILD_TAGS", ""); tags != "" {
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
	modulePackages := []string{}

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

	// Create error channel
	errChan := make(chan error, len(modulePackages))
	semaphore := make(chan struct{}, parallel)

	start := time.Now()

	// Vet packages in parallel
	for _, pkg := range modulePackages {
		go func(p string) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			args := []string{"vet"}

			// Add build tags if specified
			if tags := utils.GetEnv("GO_BUILD_TAGS", ""); tags != "" {
				args = append(args, "-tags", tags)
			}

			args = append(args, p)

			if err := GetRunner().RunCmd("go", args...); err != nil {
				errChan <- fmt.Errorf("vet failed for %s: %w", p, err)
			} else {
				errChan <- nil
			}
		}(pkg)
	}

	// Wait for all goroutines to complete
	var errors []error
	for i := 0; i < len(modulePackages); i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		utils.Error("go vet found issues in %d packages:", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
		return fmt.Errorf("go vet failed")
	}

	utils.Success("go vet passed in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// Shadow checks for shadowed variables
func (Vet) Shadow() error {
	utils.Header("Checking for Shadowed Variables")

	// Check if shadow analyzer is available

	// Try to run vet with shadow
	args := []string{"vet", "-vettool=$(which shadow)"}

	// Add build tags if specified
	if tags := utils.GetEnv("GO_BUILD_TAGS", ""); tags != "" {
		args = append(args, "-tags", tags)
	}

	args = append(args, "./...")

	// First, try with the shadow tool
	if err := GetRunner().RunCmd("go", args...); err != nil {
		// If that fails, try installing and using it
		utils.Info("Installing shadow analyzer...")
		if err := GetRunner().RunCmd("go", "install", "golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest"); err != nil {
			return fmt.Errorf("failed to install shadow analyzer: %w", err)
		}

		// Try again with shadow
		utils.Info("Running shadow check...")
		if err := GetRunner().RunCmd("shadow", "./..."); err != nil {
			return fmt.Errorf("shadow check found issues: %w", err)
		}
	}

	utils.Success("No shadowed variables found")
	return nil
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
		return fmt.Errorf("%d strict checks failed", failed)
	}

	utils.Success("\nAll strict checks passed!")
	return nil
}

// Helper functions

// runStaticcheck runs the staticcheck tool
func runStaticcheck() error {
	// Check if staticcheck is installed
	if !utils.CommandExists("staticcheck") {
		utils.Info("Installing staticcheck...")
		if err := GetRunner().RunCmd("go", "install", "honnef.co/go/tools/cmd/staticcheck@latest"); err != nil {
			return fmt.Errorf("failed to install staticcheck: %w", err)
		}
	}

	// Run staticcheck
	if err := GetRunner().RunCmd("staticcheck", "./..."); err != nil {
		return fmt.Errorf("staticcheck found issues: %w", err)
	}

	return nil
}

// runIneffassign checks for ineffectual assignments
func runIneffassign() error {
	// Check if ineffassign is installed
	if !utils.CommandExists("ineffassign") {
		utils.Info("Installing ineffassign...")
		if err := GetRunner().RunCmd("go", "install", "github.com/gordonklaus/ineffassign@latest"); err != nil {
			return fmt.Errorf("failed to install ineffassign: %w", err)
		}
	}

	// Run ineffassign
	if err := GetRunner().RunCmd("ineffassign", "./..."); err != nil {
		return fmt.Errorf("ineffassign found issues: %w", err)
	}

	return nil
}

// runMisspell checks for misspelled words
func runMisspell() error {
	// Check if misspell is installed
	if !utils.CommandExists("misspell") {
		utils.Info("Installing misspell...")
		if err := GetRunner().RunCmd("go", "install", "github.com/client9/misspell/cmd/misspell@latest"); err != nil {
			return fmt.Errorf("failed to install misspell: %w", err)
		}
	}

	// Run misspell
	if err := GetRunner().RunCmd("misspell", "-w", "."); err != nil {
		return fmt.Errorf("misspell found issues: %w", err)
	}

	return nil
}
