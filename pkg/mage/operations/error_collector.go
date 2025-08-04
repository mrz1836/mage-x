package operations

import (
	"fmt"
	"strings"

	"github.com/mrz1836/mage-x/pkg/mage"
)

// ModuleErrorCollector collects and formats errors from module operations
type ModuleErrorCollector struct {
	errors []mage.ModuleError
}

// NewModuleErrorCollector creates a new error collector
func NewModuleErrorCollector() *ModuleErrorCollector {
	return &ModuleErrorCollector{
		errors: make([]mage.ModuleError, 0),
	}
}

// Add adds an error for a specific module
func (c *ModuleErrorCollector) Add(module mage.Module, err error) {
	if err != nil {
		c.errors = append(c.errors, mage.ModuleError{
			Module: module,
			Error:  err,
		})
	}
}

// HasErrors returns true if any errors were collected
func (c *ModuleErrorCollector) HasErrors() bool {
	return len(c.errors) > 0
}

// Count returns the number of errors collected
func (c *ModuleErrorCollector) Count() int {
	return len(c.errors)
}

// Errors returns all collected errors
func (c *ModuleErrorCollector) Errors() []mage.ModuleError {
	return c.errors
}

// Report formats all errors into a single error message
func (c *ModuleErrorCollector) Report() error {
	if !c.HasErrors() {
		return nil
	}

	if len(c.errors) == 1 {
		return c.errors[0].Error
	}

	// Format multiple errors
	var errMsgs []string
	for _, modErr := range c.errors {
		location := modErr.Module.Relative
		if location == "" {
			location = modErr.Module.Path
		}
		errMsgs = append(errMsgs, fmt.Sprintf("  - %s: %v", location, modErr.Error))
	}

	return fmt.Errorf("errors in %d modules:\n%s", len(c.errors), strings.Join(errMsgs, "\n"))
}

// Summary returns a summary of errors by module
func (c *ModuleErrorCollector) Summary() string {
	if !c.HasErrors() {
		return "No errors"
	}

	modules := make([]string, 0, len(c.errors))
	for _, modErr := range c.errors {
		if modErr.Module.IsRoot {
			modules = append(modules, "root")
		} else {
			modules = append(modules, modErr.Module.Relative)
		}
	}

	return fmt.Sprintf("Errors in: %s", strings.Join(modules, ", "))
}
