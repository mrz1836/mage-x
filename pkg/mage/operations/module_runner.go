package operations

import (
	"fmt"
	"time"

	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// ModuleOperation defines the interface for operations that run on modules
type ModuleOperation interface {
	Name() string
	Execute(module mage.Module, config *mage.Config) error
}

// ModuleRunner provides centralized module discovery and operation execution
type ModuleRunner struct {
	config    *mage.Config
	operation ModuleOperation
}

// NewModuleRunner creates a new module runner
func NewModuleRunner(config *mage.Config, operation ModuleOperation) *ModuleRunner {
	return &ModuleRunner{
		config:    config,
		operation: operation,
	}
}

// RunForAllModules discovers modules and runs the operation on each
func (r *ModuleRunner) RunForAllModules() error {
	// Discover all modules
	modules, err := mage.FindAllModules()
	if err != nil {
		return fmt.Errorf("failed to find modules: %w", err)
	}

	if len(modules) == 0 {
		utils.Warn("No Go modules found")
		return nil
	}

	// Show modules found
	if len(modules) > 1 {
		utils.Info("Found %d Go modules", len(modules))
	}

	totalStart := time.Now()
	var moduleErrors []mage.ModuleError

	// Run operation for each module
	for _, module := range modules {
		r.displayModuleHeader(module)

		moduleStart := time.Now()

		// Execute the operation
		err := r.operation.Execute(module, r.config)

		if err != nil {
			moduleErrors = append(moduleErrors, mage.ModuleError{Module: module, Error: err})
			utils.Error("%s failed for %s in %s", r.operation.Name(), module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		} else {
			utils.Success("%s passed for %s in %s", r.operation.Name(), module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		}
	}

	// Report overall results
	if len(moduleErrors) > 0 {
		utils.Error("\n%s failed in %d/%d modules", r.operation.Name(), len(moduleErrors), len(modules))
		return mage.FormatModuleErrors(moduleErrors)
	}

	utils.Success("\n%s completed successfully for all modules in %s", r.operation.Name(), utils.FormatDuration(time.Since(totalStart)))
	return nil
}

// displayModuleHeader shows module information
func (r *ModuleRunner) displayModuleHeader(module mage.Module) {
	if module.IsRoot {
		utils.Info("\n📦 %s module: %s", r.operation.Name(), module.Name)
	} else {
		utils.Info("\n📦 %s module: %s (%s)", r.operation.Name(), module.Name, module.Relative)
	}
}
