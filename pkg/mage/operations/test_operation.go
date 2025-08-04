package operations

import (
	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/mage/builders"
)

// TestOperation implements ModuleOperation for running tests
type TestOperation struct {
	builder *builders.TestCommandBuilder
	options builders.TestOptions
}

// NewTestOperation creates a new test operation
func NewTestOperation(config *mage.Config, options builders.TestOptions) *TestOperation {
	return &TestOperation{
		builder: builders.NewTestCommandBuilder(config),
		options: options,
	}
}

// Name returns the operation name
func (t *TestOperation) Name() string {
	if t.options.Coverage {
		return "Coverage tests"
	}
	if t.options.Integration {
		return "Integration tests"
	}
	return "Unit tests"
}

// Execute runs tests for a module
func (t *TestOperation) Execute(module mage.Module, config *mage.Config) error {
	var args []string

	if t.options.Coverage {
		args = t.builder.BuildCoverageArgs(t.options)
	} else if t.options.Integration {
		args = t.builder.BuildIntegrationTestArgs(t.options)
	} else {
		args = t.builder.BuildUnitTestArgs(t.options)
	}

	// Run tests in module directory
	return mage.RunCommandInModule(module, "go", args...)
}
