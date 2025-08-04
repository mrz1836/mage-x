package operations

import (
	"fmt"
	"time"

	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// OperationContext provides consistent operation handling with config loading,
// timing, and reporting
type OperationContext struct {
	config    *mage.Config
	startTime time.Time
	header    string
}

// NewOperation creates a new operation context with config loading and header display
func NewOperation(header string) (*OperationContext, error) {
	// Display header
	utils.Header(header)

	// Load config once
	config, err := mage.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	return &OperationContext{
		config:    config,
		startTime: time.Now(),
		header:    header,
	}, nil
}

// Config returns the loaded configuration
func (o *OperationContext) Config() *mage.Config {
	return o.config
}

// Complete reports operation completion with timing
func (o *OperationContext) Complete(err error) error {
	duration := time.Since(o.startTime)

	if err != nil {
		utils.Error("%s failed in %s: %v", o.header, utils.FormatDuration(duration), err)
		return err
	}

	utils.Success("%s completed successfully in %s", o.header, utils.FormatDuration(duration))
	return nil
}

// CompleteWithoutTiming reports operation completion without timing info
func (o *OperationContext) CompleteWithoutTiming(err error) error {
	if err != nil {
		utils.Error("%s failed: %v", o.header, err)
		return err
	}

	utils.Success("%s completed successfully", o.header)
	return nil
}

// Info logs an info message within the operation context
func (o *OperationContext) Info(format string, args ...interface{}) {
	utils.Info(format, args...)
}

// Warn logs a warning message within the operation context
func (o *OperationContext) Warn(format string, args ...interface{}) {
	utils.Warn(format, args...)
}

// Error logs an error message within the operation context
func (o *OperationContext) Error(format string, args ...interface{}) {
	utils.Error(format, args...)
}

// Success logs a success message within the operation context
func (o *OperationContext) Success(format string, args ...interface{}) {
	utils.Success(format, args...)
}
