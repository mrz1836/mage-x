// Package mage provides comprehensive interactive test coverage for wizard functionality.
// This file contains tests for wizard steps that require input simulation.
package mage

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

// WizardInteractiveTestSuite provides testing for interactive wizard functionality
type WizardInteractiveTestSuite struct {
	suite.Suite

	tempDir string
}

// SetupSuite runs once before all tests
func (s *WizardInteractiveTestSuite) SetupSuite() {
	var err error
	s.tempDir, err = os.MkdirTemp("", "wizard_interactive_test")
	s.Require().NoError(err)
}

// TearDownSuite runs once after all tests
func (s *WizardInteractiveTestSuite) TearDownSuite() {
	if s.tempDir != "" {
		if err := os.RemoveAll(s.tempDir); err != nil {
			// Log error but don't fail test since this is cleanup
			_ = err
		}
	}
}

// createTestContext creates a wizard context with simulated input
func (s *WizardInteractiveTestSuite) createTestContext(inputs ...string) *WizardContext {
	// Create a buffer with all inputs separated by newlines
	var buffer bytes.Buffer
	for _, input := range inputs {
		buffer.WriteString(input + "\n")
	}

	scanner := bufio.NewScanner(&buffer)

	return &WizardContext{
		Data:      make(map[string]interface{}),
		Scanner:   scanner,
		Responses: make(map[string]string),
		Config:    NewEnterpriseConfiguration(),
		Errors:    []error{},
	}
}

// Test Welcome Step
func (s *WizardInteractiveTestSuite) TestWelcomeStep() {
	s.Run("executes successfully", func() {
		step := &WelcomeStep{}
		ctx := s.createTestContext("") // Press Enter to continue

		err := step.Execute(ctx)
		s.Require().NoError(err)
	})
}

// Test Organization Step
func (s *WizardInteractiveTestSuite) TestOrganizationStep() {
	s.Run("executes with valid inputs", func() {
		step := &OrganizationStep{}
		ctx := s.createTestContext(
			"Test Organization", // Organization Name
			"test.example.com",  // Domain
			"2",                 // Region (us-west-2)
			"America/New_York",  // Timezone
			"1",                 // Language (en)
		)

		err := step.Execute(ctx)
		s.Require().NoError(err)
		s.Equal("Test Organization", ctx.Config.Organization.Name)
		s.Equal("test.example.com", ctx.Config.Organization.Domain)
		s.Equal("us-west-2", ctx.Config.Organization.Region)
		s.Equal("America/New_York", ctx.Config.Organization.Timezone)
		s.Equal("en", ctx.Config.Organization.Language)
	})

	s.Run("uses defaults for empty inputs", func() {
		step := &OrganizationStep{}
		ctx := s.createTestContext("", "", "", "", "") // All defaults

		err := step.Execute(ctx)
		s.Require().NoError(err)
		s.Equal("MAGE-X Organization", ctx.Config.Organization.Name)
		s.Equal("example.com", ctx.Config.Organization.Domain)
		s.Equal("us-east-1", ctx.Config.Organization.Region)
		s.Equal("UTC", ctx.Config.Organization.Timezone)
		s.Equal("en", ctx.Config.Organization.Language)
	})

	s.Run("handles skip values", func() {
		step := &OrganizationStep{}
		ctx := s.createTestContext("skip", "skip", "skip", "skip", "skip")

		err := step.Execute(ctx)
		s.Require().NoError(err)
		s.Equal("MAGE-X Organization", ctx.Config.Organization.Name)
		s.Equal("example.com", ctx.Config.Organization.Domain)
	})
}

// Test Security Step
func (s *WizardInteractiveTestSuite) TestSecurityStep() {
	s.Run("executes with MFA enabled", func() {
		step := &SecurityStep{}
		ctx := s.createTestContext(
			"2",   // Security Level (standard)
			"y",   // Enable MFA
			"1,2", // MFA Methods (totp, sms)
			"1",   // Encryption Algorithm (AES-256)
			"y",   // Enable Security Scanning
			"1,2", // Scanning Tools (gosec, govulncheck)
			"1",   // Scanning Frequency (daily)
		)

		err := step.Execute(ctx)
		s.Require().NoError(err)
		s.Equal("standard", ctx.Config.Security.Level)
		s.True(ctx.Config.Security.Authentication.MFA.Enabled)
		s.Contains(ctx.Config.Security.Authentication.MFA.Methods, "totp")
		s.Contains(ctx.Config.Security.Authentication.MFA.Methods, "sms")
		s.Equal("AES-256", ctx.Config.Security.Encryption.Algorithm)
		s.True(ctx.Config.Security.Scanning.Enabled)
		s.Contains(ctx.Config.Security.Scanning.Tools, "gosec")
		s.Contains(ctx.Config.Security.Scanning.Tools, "govulncheck")
		s.Equal("daily", ctx.Config.Security.Scanning.Frequency)
	})

	s.Run("executes with MFA disabled", func() {
		step := &SecurityStep{}
		ctx := s.createTestContext(
			"3", // Security Level (high)
			"n", // Disable MFA
			"2", // Encryption Algorithm (AES-128)
			"n", // Disable Security Scanning
		)

		err := step.Execute(ctx)
		s.Require().NoError(err)
		s.Equal("high", ctx.Config.Security.Level)
		s.False(ctx.Config.Security.Authentication.MFA.Enabled)
		s.Equal("AES-128", ctx.Config.Security.Encryption.Algorithm)
		s.False(ctx.Config.Security.Scanning.Enabled)
	})
}

// Test Integrations Step
func (s *WizardInteractiveTestSuite) TestIntegrationsStep() {
	s.Run("executes with integrations enabled", func() {
		step := &IntegrationsStep{}
		ctx := s.createTestContext(
			"y", // Enable Integrations
			"y", // Enable slack
			"y", // Enable github
			"n", // Disable gitlab
			"n", // Disable jira
			"n", // Disable jenkins
			"y", // Enable docker
			"n", // Disable kubernetes
			"n", // Disable prometheus
			"n", // Disable grafana
			"n", // Disable aws
			"n", // Disable azure
			"n", // Disable gcp
			"y", // Enable Webhooks
			"2", // Sync Frequency (hourly)
		)

		err := step.Execute(ctx)
		s.Require().NoError(err)
		s.True(ctx.Config.Integrations.Enabled)
		s.True(ctx.Config.Integrations.Webhooks.Enabled)
		s.Equal("hourly", ctx.Config.Integrations.SyncSettings.Frequency)

		// Check enabled integrations
		s.Contains(ctx.Config.Integrations.Providers, "slack")
		s.Contains(ctx.Config.Integrations.Providers, "github")
		s.Contains(ctx.Config.Integrations.Providers, "docker")

		// Check integration types
		slackProvider := ctx.Config.Integrations.Providers["slack"]
		s.Equal("communication", slackProvider.Type)
		s.True(slackProvider.Enabled)
	})

	s.Run("executes with integrations disabled", func() {
		step := &IntegrationsStep{}
		ctx := s.createTestContext("n") // Disable Integrations

		err := step.Execute(ctx)
		s.Require().NoError(err)
		s.False(ctx.Config.Integrations.Enabled)
	})
}

// Test Workflows Step
func (s *WizardInteractiveTestSuite) TestWorkflowsStep() {
	s.Run("executes with workflows enabled", func() {
		step := &WorkflowsStep{}
		ctx := s.createTestContext(
			"y",                // Enable Workflows
			"custom/workflows", // Workflow Directory
			"y",                // Enable Scheduler
			"2",                // Scheduler Engine (systemd)
			"45m",              // Workflow Timeout
			"10",               // Max Parallel Workflows
			"5",                // Max Retry Attempts
		)

		err := step.Execute(ctx)
		s.Require().NoError(err)
		s.True(ctx.Config.Workflows.Enabled)
		s.Equal("custom/workflows", ctx.Config.Workflows.Directory)
		s.True(ctx.Config.Workflows.Scheduler.Enabled)
		s.Equal("systemd", ctx.Config.Workflows.Scheduler.Engine)
		s.Equal("45m", ctx.Config.Workflows.Execution.Timeout)
		s.Equal(10, ctx.Config.Workflows.Execution.Parallelism)
		s.Equal(5, ctx.Config.Workflows.Execution.Retry.MaxAttempts)
	})

	s.Run("executes with workflows disabled", func() {
		step := &WorkflowsStep{}
		ctx := s.createTestContext("n") // Disable Workflows

		err := step.Execute(ctx)
		s.Require().NoError(err)
		s.False(ctx.Config.Workflows.Enabled)
	})
}

// Test Analytics Step
func (s *WizardInteractiveTestSuite) TestAnalyticsStep() {
	s.Run("executes with analytics enabled", func() {
		step := &AnalyticsStep{}
		ctx := s.createTestContext(
			"y",     // Enable Analytics
			"y",     // Enable collector 1
			"y",     // Enable collector 2
			"y",     // Enable collector 3
			"y",     // Enable collector 4
			"y",     // Enable collector 5
			"3",     // Storage Type (postgresql)
			"4320h", // Data Retention Period (180 days in hours)
			"y",     // Enable Reporting
			"2",     // Report Frequency (weekly)
			"1,2",   // Report Formats (json, html)
		)

		err := step.Execute(ctx)
		s.Require().NoError(err)
		s.True(ctx.Config.Analytics.Enabled)
		s.Equal("postgresql", ctx.Config.Analytics.Storage.Type)
		s.Equal("4320h", ctx.Config.Analytics.Storage.Retention)
		s.True(ctx.Config.Analytics.Reporting.Enabled)
		s.Equal("weekly", ctx.Config.Analytics.Reporting.Frequency)
		s.Contains(ctx.Config.Analytics.Reporting.Formats, "json")
		s.Contains(ctx.Config.Analytics.Reporting.Formats, "html")

		// Check all collectors are enabled
		s.Contains(ctx.Config.Analytics.Collectors, "performance")
		s.Contains(ctx.Config.Analytics.Collectors, "build")
		s.Contains(ctx.Config.Analytics.Collectors, "test")
		s.Contains(ctx.Config.Analytics.Collectors, "security")
		s.Contains(ctx.Config.Analytics.Collectors, "usage")
		s.Len(ctx.Config.Analytics.Collectors, 5)
	})

	s.Run("executes with analytics disabled", func() {
		step := &AnalyticsStep{}
		ctx := s.createTestContext("n") // Disable Analytics

		err := step.Execute(ctx)
		s.Require().NoError(err)
		s.False(ctx.Config.Analytics.Enabled)
	})
}

// Test Deployment Step
func (s *WizardInteractiveTestSuite) TestDeploymentStep() {
	s.Run("executes with deployment enabled", func() {
		step := &DeploymentStep{}
		ctx := s.createTestContext(
			"y",                   // Enable Deployment Management
			"y",                   // Setup development environment
			"Dev Environment",     // Development description
			"1",                   // Development type (local)
			"y",                   // Setup staging environment
			"Staging Environment", // Staging description
			"2",                   // Staging type (cloud)
			"y",                   // Setup production environment
			"Prod Environment",    // Production description
			"3",                   // Production type (container)
			"y",                   // Enable Deployment Pipeline
			"y",                   // Include build stage
			"y",                   // Include test stage
			"y",                   // Include security stage
			"y",                   // Include deploy stage
			"y",                   // Include verify stage
			"y",                   // Enable Deployment Approval
			"3",                   // Environments requiring approval (production)
		)

		err := step.Execute(ctx)
		s.Require().NoError(err)
		s.True(ctx.Config.Deployment.Enabled)
		s.True(ctx.Config.Deployment.Approval.Enabled)
		s.Contains(ctx.Config.Deployment.Approval.Required, "production")

		// Check environments
		s.Contains(ctx.Config.Deployment.Environments, "development")
		s.Contains(ctx.Config.Deployment.Environments, "staging")
		s.Contains(ctx.Config.Deployment.Environments, "production")

		devEnv := ctx.Config.Deployment.Environments["development"]
		s.Equal("local", devEnv.Type)
		s.Equal("Dev Environment", devEnv.Description)

		// Check pipeline stages
		s.Len(ctx.Config.Deployment.Pipeline.Stages, 5)
		stageNames := make([]string, len(ctx.Config.Deployment.Pipeline.Stages))
		for i, stage := range ctx.Config.Deployment.Pipeline.Stages {
			stageNames[i] = stage.Name
		}
		s.Contains(stageNames, "build")
		s.Contains(stageNames, "test")
		s.Contains(stageNames, "security")
		s.Contains(stageNames, "deploy")
		s.Contains(stageNames, "verify")
	})

	s.Run("executes with deployment disabled", func() {
		step := &DeploymentStep{}
		ctx := s.createTestContext("n") // Disable Deployment

		err := step.Execute(ctx)
		s.Require().NoError(err)
		s.False(ctx.Config.Deployment.Enabled)
	})
}

// Test Compliance Step
func (s *WizardInteractiveTestSuite) TestComplianceStep() {
	s.Run("executes with compliance enabled", func() {
		step := &ComplianceStep{}
		ctx := s.createTestContext(
			"y",     // Enable Compliance Management
			"1,2",   // Compliance Standards (SOC2, ISO27001)
			"y",     // Enable Compliance Validation
			"1,2,3", // Validation Rules (security, quality, documentation)
			"y",     // Enable Compliance Reporting
			"3",     // Report Frequency (quarterly)
			"1,3",   // Report Formats (json, pdf)
		)

		err := step.Execute(ctx)
		s.Require().NoError(err)
		s.True(ctx.Config.Compliance.Enabled)
		s.Contains(ctx.Config.Compliance.Standards, "SOC2")
		s.Contains(ctx.Config.Compliance.Standards, "ISO27001")
		s.True(ctx.Config.Compliance.Validation.Enabled)
		s.Contains(ctx.Config.Compliance.Validation.Rules, "security")
		s.Contains(ctx.Config.Compliance.Validation.Rules, "quality")
		s.Contains(ctx.Config.Compliance.Validation.Rules, "documentation")
		s.True(ctx.Config.Compliance.Reporting.Enabled)
		s.Equal("quarterly", ctx.Config.Compliance.Reporting.Frequency)
		s.Contains(ctx.Config.Compliance.Reporting.Formats, "json")
		s.Contains(ctx.Config.Compliance.Reporting.Formats, "pdf")
	})

	s.Run("executes with compliance disabled", func() {
		step := &ComplianceStep{}
		ctx := s.createTestContext("n") // Disable Compliance

		err := step.Execute(ctx)
		s.Require().NoError(err)
		s.False(ctx.Config.Compliance.Enabled)
	})
}

// Test Summary Step
func (s *WizardInteractiveTestSuite) TestSummaryStep() {
	s.Run("executes and accepts configuration", func() {
		step := &SummaryStep{}
		config := NewEnterpriseConfiguration()
		config.Organization.Name = "Test Organization"
		config.Security.Level = "high"
		config.Integrations.Enabled = true

		ctx := s.createTestContext("y") // Accept configuration
		ctx.Config = config

		err := step.Execute(ctx)
		s.Require().NoError(err)
	})

	s.Run("executes and rejects configuration", func() {
		step := &SummaryStep{}
		ctx := s.createTestContext("n") // Reject configuration

		err := step.Execute(ctx)
		s.Require().Error(err)
		s.Equal(errConfigRejected, err)
	})
}

// Test Finalization Step
func (s *WizardInteractiveTestSuite) TestFinalizationStep() {
	s.Run("executes successfully", func() {
		step := &FinalizationStep{}
		ctx := s.createTestContext() // No input needed

		// Mock file operations by changing to temp directory
		originalWd, err := os.Getwd()
		s.Require().NoError(err)
		err = os.Chdir(s.tempDir)
		s.Require().NoError(err)
		defer func() {
			chdirErr := os.Chdir(originalWd)
			s.Require().NoError(chdirErr)
		}()

		// Create necessary directory structure
		enterpriseDir := filepath.Join(s.tempDir, ".mage", "enterprise")
		err = os.MkdirAll(enterpriseDir, 0o750)
		s.Require().NoError(err)

		err = step.Execute(ctx)
		s.Require().NoError(err)

		// Verify timestamp was updated
		s.True(ctx.Config.Metadata.UpdatedAt.After(time.Time{}))
	})
}

// Test Prompt Functions with various inputs
func (s *WizardInteractiveTestSuite) TestPromptFunctions() {
	s.Run("promptString", func() {
		step := &OrganizationStep{}

		// Test with custom input
		ctx := s.createTestContext("Custom Value")
		result := step.promptString(ctx, "Test Prompt", "Default", validateNonEmpty)
		s.Equal("Custom Value", result)

		// Test with empty input (uses default)
		ctx = s.createTestContext("")
		result = step.promptString(ctx, "Test Prompt", "Default", validateNonEmpty)
		s.Equal("Default", result)

		// Test with skip
		ctx = s.createTestContext("skip")
		result = step.promptString(ctx, "Test Prompt", "Default", validateNonEmpty)
		s.Equal("Default", result)
	})

	s.Run("promptChoice", func() {
		step := &OrganizationStep{}
		choices := []string{"option1", "option2", "option3"}

		// Test with valid selection
		ctx := s.createTestContext("2")
		result := step.promptChoice(ctx, "Select option", choices, "option1")
		s.Equal("option2", result)

		// Test with empty input (uses default)
		ctx = s.createTestContext("")
		result = step.promptChoice(ctx, "Select option", choices, "option2")
		s.Equal("option2", result)

		// Test with skip
		ctx = s.createTestContext("skip")
		result = step.promptChoice(ctx, "Select option", choices, "option3")
		s.Equal("option3", result)
	})

	s.Run("promptBool", func() {
		step := &SecurityStep{}

		// Test true responses
		for _, input := range []string{"y", "yes"} {
			ctx := s.createTestContext(input)
			result := step.promptBool(ctx, "Enable feature", false)
			s.True(result, "Input: %s", input)
		}

		// Test false responses
		for _, input := range []string{"n", "no"} {
			ctx := s.createTestContext(input)
			result := step.promptBool(ctx, "Enable feature", true)
			s.False(result, "Input: %s", input)
		}

		// Test empty input (uses default)
		ctx := s.createTestContext("")
		result := step.promptBool(ctx, "Enable feature", true)
		s.True(result)

		// Test skip
		ctx = s.createTestContext("skip")
		result = step.promptBool(ctx, "Enable feature", false)
		s.False(result)
	})

	s.Run("promptMultiChoice", func() {
		step := &SecurityStep{}
		choices := []string{"option1", "option2", "option3", "option4"}
		defaults := []string{"option2"}

		// Test multiple selections
		ctx := s.createTestContext("1,3")
		result := step.promptMultiChoice(ctx, "Select options", choices, defaults)
		s.Equal([]string{"option1", "option3"}, result)

		// Test empty input (uses defaults)
		ctx = s.createTestContext("")
		result = step.promptMultiChoice(ctx, "Select options", choices, defaults)
		s.Equal(defaults, result)

		// Test skip
		ctx = s.createTestContext("skip")
		result = step.promptMultiChoice(ctx, "Select options", choices, defaults)
		s.Equal(defaults, result)

		// Test with spaces
		ctx = s.createTestContext("1, 2, 4")
		result = step.promptMultiChoice(ctx, "Select options", choices, []string{})
		s.Equal([]string{"option1", "option2", "option4"}, result)
	})

	s.Run("promptInt", func() {
		step := &WorkflowsStep{}

		// Test valid input
		ctx := s.createTestContext("7")
		result := step.promptInt(ctx, "Enter number", 5, 1, 10)
		s.Equal(7, result)

		// Test empty input (uses default)
		ctx = s.createTestContext("")
		result = step.promptInt(ctx, "Enter number", 5, 1, 10)
		s.Equal(5, result)

		// Test skip
		ctx = s.createTestContext("skip")
		result = step.promptInt(ctx, "Enter number", 8, 1, 10)
		s.Equal(8, result)
	})
}

// Test Enterprise Wizard Full Flow
func (s *WizardInteractiveTestSuite) TestEnterpriseWizardFullFlow() {
	s.Run("executes minimal wizard flow", func() {
		wizard := NewEnterpriseWizard()

		// Create inputs for a minimal successful flow
		inputs := []string{
			// WelcomeStep
			"", // Press Enter to continue

			// OrganizationStep
			"Test Enterprise",     // Organization Name
			"test.enterprise.com", // Domain
			"1",                   // Region (us-east-1)
			"",                    // Timezone (default)
			"1",                   // Language (en)

			// SecurityStep
			"2", // Security Level (standard)
			"n", // Disable MFA
			"1", // Encryption Algorithm (AES-256)
			"n", // Disable Security Scanning

			// IntegrationsStep
			"n", // Disable Integrations

			// WorkflowsStep
			"n", // Disable Workflows

			// AnalyticsStep
			"n", // Disable Analytics

			// DeploymentStep
			"n", // Disable Deployment

			// ComplianceStep
			"n", // Disable Compliance

			// SummaryStep
			"y", // Accept configuration
		}

		// Create input buffer
		var buffer bytes.Buffer
		for _, input := range inputs {
			buffer.WriteString(input + "\n")
		}
		wizard.Context.Scanner = bufio.NewScanner(&buffer)

		// Mock file operations
		originalWd, err := os.Getwd()
		s.Require().NoError(err)
		err = os.Chdir(s.tempDir)
		s.Require().NoError(err)
		defer func() {
			chdirErr := os.Chdir(originalWd)
			s.Require().NoError(chdirErr)
		}()

		enterpriseDir := filepath.Join(s.tempDir, ".mage", "enterprise")
		err = os.MkdirAll(enterpriseDir, 0o750)
		s.Require().NoError(err)

		err = wizard.Run()
		s.Require().NoError(err)

		// Verify configuration was set correctly
		s.Equal("Test Enterprise", wizard.Context.Config.Organization.Name)
		s.Equal("test.enterprise.com", wizard.Context.Config.Organization.Domain)
		s.Equal("standard", wizard.Context.Config.Security.Level)
		s.False(wizard.Context.Config.Security.Authentication.MFA.Enabled)
	})
}

// Test askToContinue method
func (s *WizardInteractiveTestSuite) TestAskToContinue() {
	wizard := NewEnterpriseWizard()

	tests := []struct {
		input    string
		expected bool
	}{
		{"y", true},
		{"yes", true},
		{"Y", true},
		{"YES", true},
		{"n", false},
		{"no", false},
		{"N", false},
		{"NO", false},
		{"maybe", false},
		{"", false},
	}

	for _, tt := range tests {
		s.Run(fmt.Sprintf("input_%s", tt.input), func() {
			var buffer bytes.Buffer
			buffer.WriteString(tt.input + "\n")
			wizard.Context.Scanner = bufio.NewScanner(&buffer)

			result := wizard.askToContinue()
			s.Equal(tt.expected, result)
		})
	}
}

// Test Runner
func TestWizardInteractiveSuite(t *testing.T) {
	suite.Run(t, new(WizardInteractiveTestSuite))
}
