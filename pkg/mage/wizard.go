// Package mage provides interactive wizard capabilities for enterprise configuration
package mage

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/utils"
)

// Wizard namespace for interactive configuration wizards
type Wizard mg.Namespace

// Setup runs the complete enterprise setup wizard
func (Wizard) Setup() error {
	utils.Header("🧙 MAGE-X Enterprise Setup Wizard")

	wizard := NewEnterpriseWizard()
	return wizard.Run()
}

// Project runs the project configuration wizard
func (Wizard) Project() error {
	utils.Header("📁 Project Configuration Wizard")

	wizard := NewProjectWizard()
	return wizard.Run()
}

// Integration runs the integration setup wizard
func (Wizard) Integration() error {
	utils.Header("🔌 Integration Setup Wizard")

	wizard := NewIntegrationWizard()
	return wizard.Run()
}

// Security runs the security configuration wizard
func (Wizard) Security() error {
	utils.Header("🔒 Security Configuration Wizard")

	wizard := NewSecurityWizard()
	return wizard.Run()
}

// Workflow runs the workflow configuration wizard
func (Wizard) Workflow() error {
	utils.Header("🔄 Workflow Configuration Wizard")

	wizard := NewWorkflowWizard()
	return wizard.Run()
}

// Deployment runs the deployment configuration wizard
func (Wizard) Deployment() error {
	utils.Header("🚀 Deployment Configuration Wizard")

	wizard := NewDeploymentWizard()
	return wizard.Run()
}

// Wizard Types and Interfaces

// InteractiveWizard defines the interface for interactive setup wizards
type InteractiveWizard interface {
	Run() error
	GetName() string
	GetDescription() string
}

// WizardStep defines the interface for individual wizard steps
type WizardStep interface {
	Execute(context *WizardContext) error
	GetName() string
	GetDescription() string
	IsRequired() bool
	GetDependencies() []string
}

// WizardContext holds the state and data for wizard execution
type WizardContext struct {
	Data      map[string]interface{}
	Scanner   *bufio.Scanner
	Responses map[string]string
	Config    *EnterpriseConfiguration
	Errors    []error
}

// BaseWizard provides common wizard functionality
type BaseWizard struct {
	Name        string
	Description string
	Steps       []WizardStep
	Context     *WizardContext
}

// Enterprise Setup Wizard

// EnterpriseWizard handles enterprise configuration setup
type EnterpriseWizard struct {
	BaseWizard
}

// NewEnterpriseWizard creates a new enterprise configuration wizard
func NewEnterpriseWizard() *EnterpriseWizard {
	wizard := &EnterpriseWizard{
		BaseWizard: BaseWizard{
			Name:        "Enterprise Setup",
			Description: "Complete enterprise configuration setup",
			Context: &WizardContext{
				Data:      make(map[string]interface{}),
				Scanner:   bufio.NewScanner(os.Stdin),
				Responses: make(map[string]string),
				Config:    NewEnterpriseConfiguration(),
				Errors:    []error{},
			},
		},
	}

	wizard.Steps = []WizardStep{
		&WelcomeStep{},
		&OrganizationStep{},
		&SecurityStep{},
		&IntegrationsStep{},
		&WorkflowsStep{},
		&AnalyticsStep{},
		&DeploymentStep{},
		&ComplianceStep{},
		&SummaryStep{},
		&FinalizationStep{},
	}

	return wizard
}

// Run executes the enterprise wizard setup process
func (w *EnterpriseWizard) Run() error {
	utils.Info("🚀 Starting Enterprise Setup Wizard")
	utils.Info("📋 This wizard will guide you through setting up MAGE-X for enterprise use")
	utils.Info("")

	// Execute steps in order
	for i, step := range w.Steps {
		utils.Info("📍 Step %d/%d: %s", i+1, len(w.Steps), step.GetName())

		if err := step.Execute(w.Context); err != nil {
			w.Context.Errors = append(w.Context.Errors, err)
			utils.Error("❌ Step failed: %v", err)

			if step.IsRequired() {
				return fmt.Errorf("required step failed: %w", err)
			}

			if !w.askToContinue() {
				return fmt.Errorf("wizard canceled by user")
			}
		} else {
			utils.Success("✅ %s completed", step.GetName())
		}

		utils.Info("")
	}

	// Display summary
	if len(w.Context.Errors) > 0 {
		utils.Warn("⚠️  Wizard completed with %d errors", len(w.Context.Errors))
	} else {
		utils.Success("🎉 Enterprise setup completed successfully!")
	}

	return nil
}

// GetName returns the wizard name
func (w *EnterpriseWizard) GetName() string {
	return w.Name
}

// GetDescription returns the wizard description
func (w *EnterpriseWizard) GetDescription() string {
	return w.Description
}

func (w *EnterpriseWizard) askToContinue() bool {
	fmt.Print("Continue with the next step? (y/N): ")
	w.Context.Scanner.Scan()
	response := strings.ToLower(strings.TrimSpace(w.Context.Scanner.Text()))
	return response == "y" || response == "yes"
}

// Wizard Steps

// WelcomeStep handles the welcome/introduction step
type WelcomeStep struct{}

// Execute runs the welcome step
func (s *WelcomeStep) Execute(ctx *WizardContext) error {
	utils.Info("👋 Welcome to MAGE-X Enterprise Setup!")
	utils.Info("")
	utils.Info("This wizard will help you configure MAGE-X for enterprise use.")
	utils.Info("You can press Enter to use default values or type 'skip' to skip optional steps.")
	utils.Info("")

	fmt.Print("Press Enter to continue...")
	ctx.Scanner.Scan()

	return nil
}

// GetName returns the step name
func (s *WelcomeStep) GetName() string {
	return "Welcome"
}

// GetDescription returns the step description
func (s *WelcomeStep) GetDescription() string {
	return "Welcome and introduction"
}

// IsRequired indicates if this step is required
func (s *WelcomeStep) IsRequired() bool {
	return false
}

// GetDependencies returns step dependencies
func (s *WelcomeStep) GetDependencies() []string {
	return []string{}
}

// OrganizationStep handles organization configuration
type OrganizationStep struct{}

// Execute runs the organization configuration step
func (s *OrganizationStep) Execute(ctx *WizardContext) error {
	utils.Info("🏢 Organization Configuration")

	// Organization Name
	orgName := s.promptString(ctx, "Organization Name", "MAGE-X Organization", validateNonEmpty)
	ctx.Config.Organization.Name = orgName

	// Domain
	domain := s.promptString(ctx, "Organization Domain", "example.com", validateDomain)
	ctx.Config.Organization.Domain = domain

	// Region
	region := s.promptChoice(ctx, "Region", []string{
		"us-east-1", "us-west-2", "eu-west-1", "eu-central-1",
		"ap-southeast-1", "ap-northeast-1", "other",
	}, "us-east-1")
	ctx.Config.Organization.Region = region

	// Timezone
	timezone := s.promptString(ctx, "Timezone", "UTC", validateTimezone)
	ctx.Config.Organization.Timezone = timezone

	// Language
	language := s.promptChoice(ctx, "Language", []string{
		"en", "es", "fr", "de", "ja", "zh", "other",
	}, "en")
	ctx.Config.Organization.Language = language

	return nil
}

// GetName returns the step name
func (s *OrganizationStep) GetName() string {
	return "Organization Configuration"
}

// GetDescription returns the step description
func (s *OrganizationStep) GetDescription() string {
	return "Configure organization settings"
}

// IsRequired indicates if this step is required
func (s *OrganizationStep) IsRequired() bool {
	return true
}

// GetDependencies returns step dependencies
func (s *OrganizationStep) GetDependencies() []string {
	return []string{}
}

func (s *OrganizationStep) promptString(ctx *WizardContext, prompt, defaultValue string, validator func(string) error) string {
	for {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
		ctx.Scanner.Scan()
		input := strings.TrimSpace(ctx.Scanner.Text())

		if input == "" {
			input = defaultValue
		}

		if input == "skip" {
			return defaultValue
		}

		if validator != nil {
			if err := validator(input); err != nil {
				utils.Error("Invalid input: %v", err)
				continue
			}
		}

		return input
	}
}

func (s *OrganizationStep) promptChoice(ctx *WizardContext, prompt string, choices []string, defaultValue string) string {
	fmt.Printf("%s:\n", prompt)
	for i, choice := range choices {
		marker := " "
		if choice == defaultValue {
			marker = "*"
		}
		fmt.Printf("  %s %d. %s\n", marker, i+1, choice)
	}

	for {
		fmt.Printf("Select option (1-%d) [%s]: ", len(choices), defaultValue)
		ctx.Scanner.Scan()
		input := strings.TrimSpace(ctx.Scanner.Text())

		if input == "" {
			return defaultValue
		}

		if input == "skip" {
			return defaultValue
		}

		if idx, err := strconv.Atoi(input); err == nil && idx >= 1 && idx <= len(choices) {
			return choices[idx-1]
		}

		utils.Error("Invalid selection. Please choose 1-%d.", len(choices))
	}
}

// SecurityStep handles security configuration
type SecurityStep struct{}

// Execute runs the security configuration step
func (s *SecurityStep) Execute(ctx *WizardContext) error {
	utils.Info("🔒 Security Configuration")

	// Security Level
	level := s.promptChoice(ctx, "Security Level", []string{
		"minimal", "standard", "high", "critical",
	}, "standard")
	ctx.Config.Security.Level = level

	// MFA
	mfa := s.promptBool(ctx, "Enable Multi-Factor Authentication", false)
	ctx.Config.Security.Authentication.MFA.Enabled = mfa

	if mfa {
		methods := s.promptMultiChoice(ctx, "MFA Methods", []string{
			"totp", "sms", "email", "hardware",
		}, []string{"totp"})
		ctx.Config.Security.Authentication.MFA.Methods = methods
	}

	// Encryption
	encryption := s.promptChoice(ctx, "Encryption Algorithm", []string{
		"AES-256", "AES-128", "ChaCha20-Poly1305",
	}, "AES-256")
	ctx.Config.Security.Encryption.Algorithm = encryption

	// Scanning
	scanning := s.promptBool(ctx, "Enable Security Scanning", true)
	ctx.Config.Security.Scanning.Enabled = scanning

	if scanning {
		tools := s.promptMultiChoice(ctx, "Security Scanning Tools", []string{
			"gosec", "govulncheck", "nancy", "snyk", "semgrep",
		}, []string{"gosec", "govulncheck"})
		ctx.Config.Security.Scanning.Tools = tools

		frequency := s.promptChoice(ctx, "Scanning Frequency", []string{
			"daily", "weekly", "monthly", "on-demand",
		}, "daily")
		ctx.Config.Security.Scanning.Frequency = frequency
	}

	return nil
}

// GetName returns the step name
func (s *SecurityStep) GetName() string {
	return "Security Configuration"
}

// GetDescription returns the step description
func (s *SecurityStep) GetDescription() string {
	return "Configure security settings"
}

// IsRequired indicates if this step is required
func (s *SecurityStep) IsRequired() bool {
	return true
}

// GetDependencies returns step dependencies
func (s *SecurityStep) GetDependencies() []string {
	return []string{}
}

func (s *SecurityStep) promptChoice(ctx *WizardContext, prompt string, choices []string, defaultValue string) string {
	fmt.Printf("%s:\n", prompt)
	for i, choice := range choices {
		marker := " "
		if choice == defaultValue {
			marker = "*"
		}
		fmt.Printf("  %s %d. %s\n", marker, i+1, choice)
	}

	for {
		fmt.Printf("Select option (1-%d) [%s]: ", len(choices), defaultValue)
		ctx.Scanner.Scan()
		input := strings.TrimSpace(ctx.Scanner.Text())

		if input == "" {
			return defaultValue
		}

		if input == "skip" {
			return defaultValue
		}

		if idx, err := strconv.Atoi(input); err == nil && idx >= 1 && idx <= len(choices) {
			return choices[idx-1]
		}

		utils.Error("Invalid selection. Please choose 1-%d.", len(choices))
	}
}

func (s *SecurityStep) promptBool(ctx *WizardContext, prompt string, defaultValue bool) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}

	for {
		fmt.Printf("%s (y/n) [%s]: ", prompt, defaultStr)
		ctx.Scanner.Scan()
		input := strings.ToLower(strings.TrimSpace(ctx.Scanner.Text()))

		if input == "" {
			return defaultValue
		}

		if input == "skip" {
			return defaultValue
		}

		if input == "y" || input == "yes" {
			return true
		}

		if input == "n" || input == "no" {
			return false
		}

		utils.Error("Please enter y or n")
	}
}

func (s *SecurityStep) promptMultiChoice(ctx *WizardContext, prompt string, choices []string, defaultValues []string) []string {
	fmt.Printf("%s (select multiple by entering numbers separated by commas):\n", prompt)
	for i, choice := range choices {
		marker := " "
		for _, def := range defaultValues {
			if choice == def {
				marker = "*"
				break
			}
		}
		fmt.Printf("  %s %d. %s\n", marker, i+1, choice)
	}

	defaultStr := ""
	if len(defaultValues) > 0 {
		defaultStr = strings.Join(defaultValues, ", ")
	}

	for {
		fmt.Printf("Select options (e.g., 1,3,5) [%s]: ", defaultStr)
		ctx.Scanner.Scan()
		input := strings.TrimSpace(ctx.Scanner.Text())

		if input == "" {
			return defaultValues
		}

		if input == "skip" {
			return defaultValues
		}

		var selected []string
		for _, part := range strings.Split(input, ",") {
			part = strings.TrimSpace(part)
			if idx, err := strconv.Atoi(part); err == nil && idx >= 1 && idx <= len(choices) {
				selected = append(selected, choices[idx-1])
			}
		}

		if len(selected) > 0 {
			return selected
		}

		utils.Error("Invalid selection. Please enter valid numbers separated by commas.")
	}
}

// IntegrationsStep handles third-party integrations configuration
type IntegrationsStep struct{}

// Execute runs the integrations configuration step
func (s *IntegrationsStep) Execute(ctx *WizardContext) error {
	utils.Info("🔌 Integration Configuration")

	// Enable integrations
	enabled := s.promptBool(ctx, "Enable Integrations", true)
	ctx.Config.Integrations.Enabled = enabled

	if !enabled {
		return nil
	}

	// Available integrations
	integrations := map[string]string{
		"slack":      "Slack messaging",
		"github":     "GitHub source control",
		"gitlab":     "GitLab source control",
		"jira":       "Jira issue tracking",
		"jenkins":    "Jenkins CI/CD",
		"docker":     "Docker containers",
		"kubernetes": "Kubernetes orchestration",
		"prometheus": "Prometheus monitoring",
		"grafana":    "Grafana dashboards",
		"aws":        "Amazon Web Services",
		"azure":      "Microsoft Azure",
		"gcp":        "Google Cloud Platform",
	}

	ctx.Config.Integrations.Providers = make(map[string]IntegrationProvider)

	for name, description := range integrations {
		enable := s.promptBool(ctx, fmt.Sprintf("Enable %s (%s)", name, description), false)
		if enable {
			provider := IntegrationProvider{
				Type:        s.getIntegrationType(name),
				Enabled:     true,
				Settings:    make(map[string]string),
				Credentials: make(map[string]string),
				Endpoints:   make(map[string]string),
			}
			ctx.Config.Integrations.Providers[name] = provider
		}
	}

	// Webhooks
	webhooks := s.promptBool(ctx, "Enable Webhooks", false)
	ctx.Config.Integrations.Webhooks.Enabled = webhooks

	// Sync settings
	frequency := s.promptChoice(ctx, "Integration Sync Frequency", []string{
		"realtime", "hourly", "daily", "weekly", "manual",
	}, "hourly")
	ctx.Config.Integrations.SyncSettings.Frequency = frequency

	return nil
}

// GetName returns the step name
func (s *IntegrationsStep) GetName() string {
	return "Integration Configuration"
}

// GetDescription returns the step description
func (s *IntegrationsStep) GetDescription() string {
	return "Configure enterprise integrations"
}

// IsRequired indicates if this step is required
func (s *IntegrationsStep) IsRequired() bool {
	return false
}

// GetDependencies returns step dependencies
func (s *IntegrationsStep) GetDependencies() []string {
	return []string{}
}

func (s *IntegrationsStep) promptBool(ctx *WizardContext, prompt string, defaultValue bool) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}

	for {
		fmt.Printf("%s (y/n) [%s]: ", prompt, defaultStr)
		ctx.Scanner.Scan()
		input := strings.ToLower(strings.TrimSpace(ctx.Scanner.Text()))

		if input == "" {
			return defaultValue
		}

		if input == "skip" {
			return defaultValue
		}

		if input == "y" || input == "yes" {
			return true
		}

		if input == "n" || input == "no" {
			return false
		}

		utils.Error("Please enter y or n")
	}
}

func (s *IntegrationsStep) promptChoice(ctx *WizardContext, prompt string, choices []string, defaultValue string) string {
	fmt.Printf("%s:\n", prompt)
	for i, choice := range choices {
		marker := " "
		if choice == defaultValue {
			marker = "*"
		}
		fmt.Printf("  %s %d. %s\n", marker, i+1, choice)
	}

	for {
		fmt.Printf("Select option (1-%d) [%s]: ", len(choices), defaultValue)
		ctx.Scanner.Scan()
		input := strings.TrimSpace(ctx.Scanner.Text())

		if input == "" {
			return defaultValue
		}

		if input == "skip" {
			return defaultValue
		}

		if idx, err := strconv.Atoi(input); err == nil && idx >= 1 && idx <= len(choices) {
			return choices[idx-1]
		}

		utils.Error("Invalid selection. Please choose 1-%d.", len(choices))
	}
}

func (s *IntegrationsStep) getIntegrationType(name string) string {
	types := map[string]string{
		"slack":      "communication",
		"github":     "source_control",
		"gitlab":     "source_control",
		"jira":       "issue_tracking",
		"jenkins":    "ci_cd",
		"docker":     "containerization",
		"kubernetes": "orchestration",
		"prometheus": "monitoring",
		"grafana":    "monitoring",
		"aws":        "cloud_provider",
		"azure":      "cloud_provider",
		"gcp":        "cloud_provider",
	}

	if t, exists := types[name]; exists {
		return t
	}
	return "other"
}

// WorkflowsStep handles workflow configuration
type WorkflowsStep struct{}

// Execute runs the workflows configuration step
func (s *WorkflowsStep) Execute(ctx *WizardContext) error {
	utils.Info("🔄 Workflow Configuration")

	// Enable workflows
	enabled := s.promptBool(ctx, "Enable Workflows", true)
	ctx.Config.Workflows.Enabled = enabled

	if !enabled {
		return nil
	}

	// Workflow directory
	directory := s.promptString(ctx, "Workflow Directory", ".mage/workflows", nil)
	ctx.Config.Workflows.Directory = directory

	// Scheduler
	scheduler := s.promptBool(ctx, "Enable Workflow Scheduler", false)
	ctx.Config.Workflows.Scheduler.Enabled = scheduler

	if scheduler {
		engine := s.promptChoice(ctx, "Scheduler Engine", []string{
			"cron", "systemd", "kubernetes", "custom",
		}, "cron")
		ctx.Config.Workflows.Scheduler.Engine = engine
	}

	// Execution settings
	timeout := s.promptString(ctx, "Default Workflow Timeout", "30m", validateDuration)
	ctx.Config.Workflows.Execution.Timeout = timeout

	parallelism := s.promptInt(ctx, "Max Parallel Workflows", 5, 1, 50)
	ctx.Config.Workflows.Execution.Parallelism = parallelism

	// Retry settings
	maxRetries := s.promptInt(ctx, "Max Retry Attempts", 3, 0, 10)
	ctx.Config.Workflows.Execution.Retry.MaxAttempts = maxRetries

	return nil
}

// GetName returns the step name
func (s *WorkflowsStep) GetName() string {
	return "Workflow Configuration"
}

// GetDescription returns the step description
func (s *WorkflowsStep) GetDescription() string {
	return "Configure workflow settings"
}

// IsRequired indicates if this step is required
func (s *WorkflowsStep) IsRequired() bool {
	return false
}

// GetDependencies returns step dependencies
func (s *WorkflowsStep) GetDependencies() []string {
	return []string{}
}

func (s *WorkflowsStep) promptBool(ctx *WizardContext, prompt string, defaultValue bool) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}

	for {
		fmt.Printf("%s (y/n) [%s]: ", prompt, defaultStr)
		ctx.Scanner.Scan()
		input := strings.ToLower(strings.TrimSpace(ctx.Scanner.Text()))

		if input == "" {
			return defaultValue
		}

		if input == "skip" {
			return defaultValue
		}

		if input == "y" || input == "yes" {
			return true
		}

		if input == "n" || input == "no" {
			return false
		}

		utils.Error("Please enter y or n")
	}
}

func (s *WorkflowsStep) promptString(ctx *WizardContext, prompt, defaultValue string, validator func(string) error) string {
	for {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
		ctx.Scanner.Scan()
		input := strings.TrimSpace(ctx.Scanner.Text())

		if input == "" {
			input = defaultValue
		}

		if input == "skip" {
			return defaultValue
		}

		if validator != nil {
			if err := validator(input); err != nil {
				utils.Error("Invalid input: %v", err)
				continue
			}
		}

		return input
	}
}

func (s *WorkflowsStep) promptChoice(ctx *WizardContext, prompt string, choices []string, defaultValue string) string {
	fmt.Printf("%s:\n", prompt)
	for i, choice := range choices {
		marker := " "
		if choice == defaultValue {
			marker = "*"
		}
		fmt.Printf("  %s %d. %s\n", marker, i+1, choice)
	}

	for {
		fmt.Printf("Select option (1-%d) [%s]: ", len(choices), defaultValue)
		ctx.Scanner.Scan()
		input := strings.TrimSpace(ctx.Scanner.Text())

		if input == "" {
			return defaultValue
		}

		if input == "skip" {
			return defaultValue
		}

		if idx, err := strconv.Atoi(input); err == nil && idx >= 1 && idx <= len(choices) {
			return choices[idx-1]
		}

		utils.Error("Invalid selection. Please choose 1-%d.", len(choices))
	}
}

func (s *WorkflowsStep) promptInt(ctx *WizardContext, prompt string, defaultValue, minValue, maxValue int) int {
	for {
		fmt.Printf("%s [%d]: ", prompt, defaultValue)
		ctx.Scanner.Scan()
		input := strings.TrimSpace(ctx.Scanner.Text())

		if input == "" {
			return defaultValue
		}

		if input == "skip" {
			return defaultValue
		}

		if value, err := strconv.Atoi(input); err == nil {
			if value >= minValue && value <= maxValue {
				return value
			}
			utils.Error("Value must be between %d and %d", minValue, maxValue)
		} else {
			utils.Error("Please enter a valid number")
		}
	}
}

// AnalyticsStep handles analytics and monitoring configuration
type AnalyticsStep struct{}

// Execute runs the analytics configuration step
func (s *AnalyticsStep) Execute(ctx *WizardContext) error {
	utils.Info("📊 Analytics Configuration")

	// Enable analytics
	enabled := s.promptBool(ctx, "Enable Analytics", true)
	ctx.Config.Analytics.Enabled = enabled

	if !enabled {
		return nil
	}

	// Collectors
	collectors := map[string]string{
		"performance": "Performance metrics",
		"build":       "Build metrics",
		"test":        "Test metrics",
		"security":    "Security metrics",
		"usage":       "Usage metrics",
	}

	ctx.Config.Analytics.Collectors = make(map[string]CollectorConfig)

	for name, description := range collectors {
		enable := s.promptBool(ctx, fmt.Sprintf("Enable %s (%s)", name, description), true)
		if enable {
			collector := CollectorConfig{
				Type:     name,
				Enabled:  true,
				Settings: make(map[string]string),
			}
			ctx.Config.Analytics.Collectors[name] = collector
		}
	}

	// Storage
	storageType := s.promptChoice(ctx, "Storage Type", []string{
		"file", "sqlite", "postgresql", "mysql", "mongodb",
	}, "file")
	ctx.Config.Analytics.Storage.Type = storageType

	retention := s.promptString(ctx, "Data Retention Period", "90d", validateDuration)
	ctx.Config.Analytics.Storage.Retention = retention

	// Reporting
	reporting := s.promptBool(ctx, "Enable Reporting", true)
	ctx.Config.Analytics.Reporting.Enabled = reporting

	if reporting {
		frequency := s.promptChoice(ctx, "Report Frequency", []string{
			"daily", "weekly", "monthly", "quarterly",
		}, "weekly")
		ctx.Config.Analytics.Reporting.Frequency = frequency

		formats := s.promptMultiChoice(ctx, "Report Formats", []string{
			"json", "html", "pdf", "csv", "xml",
		}, []string{"json", "html"})
		ctx.Config.Analytics.Reporting.Formats = formats
	}

	return nil
}

// GetName returns the step name
func (s *AnalyticsStep) GetName() string {
	return "Analytics Configuration"
}

// GetDescription returns the step description
func (s *AnalyticsStep) GetDescription() string {
	return "Configure analytics and reporting"
}

// IsRequired indicates if this step is required
func (s *AnalyticsStep) IsRequired() bool {
	return false
}

// GetDependencies returns step dependencies
func (s *AnalyticsStep) GetDependencies() []string {
	return []string{}
}

func (s *AnalyticsStep) promptBool(ctx *WizardContext, prompt string, defaultValue bool) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}

	for {
		fmt.Printf("%s (y/n) [%s]: ", prompt, defaultStr)
		ctx.Scanner.Scan()
		input := strings.ToLower(strings.TrimSpace(ctx.Scanner.Text()))

		if input == "" {
			return defaultValue
		}

		if input == "skip" {
			return defaultValue
		}

		if input == "y" || input == "yes" {
			return true
		}

		if input == "n" || input == "no" {
			return false
		}

		utils.Error("Please enter y or n")
	}
}

func (s *AnalyticsStep) promptChoice(ctx *WizardContext, prompt string, choices []string, defaultValue string) string {
	fmt.Printf("%s:\n", prompt)
	for i, choice := range choices {
		marker := " "
		if choice == defaultValue {
			marker = "*"
		}
		fmt.Printf("  %s %d. %s\n", marker, i+1, choice)
	}

	for {
		fmt.Printf("Select option (1-%d) [%s]: ", len(choices), defaultValue)
		ctx.Scanner.Scan()
		input := strings.TrimSpace(ctx.Scanner.Text())

		if input == "" {
			return defaultValue
		}

		if input == "skip" {
			return defaultValue
		}

		if idx, err := strconv.Atoi(input); err == nil && idx >= 1 && idx <= len(choices) {
			return choices[idx-1]
		}

		utils.Error("Invalid selection. Please choose 1-%d.", len(choices))
	}
}

func (s *AnalyticsStep) promptString(ctx *WizardContext, prompt, defaultValue string, validator func(string) error) string {
	for {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
		ctx.Scanner.Scan()
		input := strings.TrimSpace(ctx.Scanner.Text())

		if input == "" {
			input = defaultValue
		}

		if input == "skip" {
			return defaultValue
		}

		if validator != nil {
			if err := validator(input); err != nil {
				utils.Error("Invalid input: %v", err)
				continue
			}
		}

		return input
	}
}

func (s *AnalyticsStep) promptMultiChoice(ctx *WizardContext, prompt string, choices []string, defaultValues []string) []string {
	fmt.Printf("%s (select multiple by entering numbers separated by commas):\n", prompt)
	for i, choice := range choices {
		marker := " "
		for _, def := range defaultValues {
			if choice == def {
				marker = "*"
				break
			}
		}
		fmt.Printf("  %s %d. %s\n", marker, i+1, choice)
	}

	defaultStr := ""
	if len(defaultValues) > 0 {
		defaultStr = strings.Join(defaultValues, ", ")
	}

	for {
		fmt.Printf("Select options (e.g., 1,3,5) [%s]: ", defaultStr)
		ctx.Scanner.Scan()
		input := strings.TrimSpace(ctx.Scanner.Text())

		if input == "" {
			return defaultValues
		}

		if input == "skip" {
			return defaultValues
		}

		var selected []string
		for _, part := range strings.Split(input, ",") {
			part = strings.TrimSpace(part)
			if idx, err := strconv.Atoi(part); err == nil && idx >= 1 && idx <= len(choices) {
				selected = append(selected, choices[idx-1])
			}
		}

		if len(selected) > 0 {
			return selected
		}

		utils.Error("Invalid selection. Please enter valid numbers separated by commas.")
	}
}

// DeploymentStep handles deployment configuration
type DeploymentStep struct{}

// Execute runs the deployment configuration step
func (s *DeploymentStep) Execute(ctx *WizardContext) error {
	utils.Info("🚀 Deployment Configuration")

	// Enable deployment
	enabled := s.promptBool(ctx, "Enable Deployment Management", true)
	ctx.Config.Deployment.Enabled = enabled

	if !enabled {
		return nil
	}

	// Environments
	environments := []string{"development", "staging", "production"}
	ctx.Config.Deployment.Environments = make(map[string]Environment)

	for _, env := range environments {
		setup := s.promptBool(ctx, fmt.Sprintf("Setup %s environment", env), true)
		if setup {
			description := s.promptString(ctx, fmt.Sprintf("%s environment description", env),
				fmt.Sprintf("%s environment", strings.ToUpper(env[:1])+env[1:]), nil)

			envType := s.promptChoice(ctx, fmt.Sprintf("%s environment type", env), []string{
				"local", "cloud", "container", "kubernetes", "serverless",
			}, "cloud")

			ctx.Config.Deployment.Environments[env] = Environment{
				Name:        env,
				Description: description,
				Type:        envType,
				Settings:    make(map[string]string),
			}
		}
	}

	// Pipeline
	pipeline := s.promptBool(ctx, "Enable Deployment Pipeline", true)
	if pipeline {
		stages := []string{"build", "test", "security", "deploy", "verify"}
		ctx.Config.Deployment.Pipeline.Stages = []PipelineStage{}

		for _, stage := range stages {
			include := s.promptBool(ctx, fmt.Sprintf("Include %s stage", stage), true)
			if include {
				ctx.Config.Deployment.Pipeline.Stages = append(ctx.Config.Deployment.Pipeline.Stages, PipelineStage{
					Name:        stage,
					Description: fmt.Sprintf("%s stage", strings.ToUpper(stage[:1])+stage[1:]),
					Commands:    []string{},
					Settings:    make(map[string]string),
				})
			}
		}
	}

	// Approval
	approval := s.promptBool(ctx, "Enable Deployment Approval", false)
	ctx.Config.Deployment.Approval.Enabled = approval

	if approval {
		required := s.promptMultiChoice(ctx, "Environments requiring approval", []string{
			"development", "staging", "production",
		}, []string{"production"})
		ctx.Config.Deployment.Approval.Required = required
	}

	return nil
}

// GetName returns the step name
func (s *DeploymentStep) GetName() string {
	return "Deployment Configuration"
}

// GetDescription returns the step description
func (s *DeploymentStep) GetDescription() string {
	return "Configure deployment settings"
}

// IsRequired indicates if this step is required
func (s *DeploymentStep) IsRequired() bool {
	return false
}

// GetDependencies returns step dependencies
func (s *DeploymentStep) GetDependencies() []string {
	return []string{}
}

func (s *DeploymentStep) promptBool(ctx *WizardContext, prompt string, defaultValue bool) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}

	for {
		fmt.Printf("%s (y/n) [%s]: ", prompt, defaultStr)
		ctx.Scanner.Scan()
		input := strings.ToLower(strings.TrimSpace(ctx.Scanner.Text()))

		if input == "" {
			return defaultValue
		}

		if input == "skip" {
			return defaultValue
		}

		if input == "y" || input == "yes" {
			return true
		}

		if input == "n" || input == "no" {
			return false
		}

		utils.Error("Please enter y or n")
	}
}

func (s *DeploymentStep) promptString(ctx *WizardContext, prompt, defaultValue string, validator func(string) error) string {
	for {
		fmt.Printf("%s [%s]: ", prompt, defaultValue)
		ctx.Scanner.Scan()
		input := strings.TrimSpace(ctx.Scanner.Text())

		if input == "" {
			input = defaultValue
		}

		if input == "skip" {
			return defaultValue
		}

		if validator != nil {
			if err := validator(input); err != nil {
				utils.Error("Invalid input: %v", err)
				continue
			}
		}

		return input
	}
}

func (s *DeploymentStep) promptChoice(ctx *WizardContext, prompt string, choices []string, defaultValue string) string {
	fmt.Printf("%s:\n", prompt)
	for i, choice := range choices {
		marker := " "
		if choice == defaultValue {
			marker = "*"
		}
		fmt.Printf("  %s %d. %s\n", marker, i+1, choice)
	}

	for {
		fmt.Printf("Select option (1-%d) [%s]: ", len(choices), defaultValue)
		ctx.Scanner.Scan()
		input := strings.TrimSpace(ctx.Scanner.Text())

		if input == "" {
			return defaultValue
		}

		if input == "skip" {
			return defaultValue
		}

		if idx, err := strconv.Atoi(input); err == nil && idx >= 1 && idx <= len(choices) {
			return choices[idx-1]
		}

		utils.Error("Invalid selection. Please choose 1-%d.", len(choices))
	}
}

func (s *DeploymentStep) promptMultiChoice(ctx *WizardContext, prompt string, choices []string, defaultValues []string) []string {
	fmt.Printf("%s (select multiple by entering numbers separated by commas):\n", prompt)
	for i, choice := range choices {
		marker := " "
		for _, def := range defaultValues {
			if choice == def {
				marker = "*"
				break
			}
		}
		fmt.Printf("  %s %d. %s\n", marker, i+1, choice)
	}

	defaultStr := ""
	if len(defaultValues) > 0 {
		defaultStr = strings.Join(defaultValues, ", ")
	}

	for {
		fmt.Printf("Select options (e.g., 1,3,5) [%s]: ", defaultStr)
		ctx.Scanner.Scan()
		input := strings.TrimSpace(ctx.Scanner.Text())

		if input == "" {
			return defaultValues
		}

		if input == "skip" {
			return defaultValues
		}

		var selected []string
		for _, part := range strings.Split(input, ",") {
			part = strings.TrimSpace(part)
			if idx, err := strconv.Atoi(part); err == nil && idx >= 1 && idx <= len(choices) {
				selected = append(selected, choices[idx-1])
			}
		}

		if len(selected) > 0 {
			return selected
		}

		utils.Error("Invalid selection. Please enter valid numbers separated by commas.")
	}
}

// ComplianceStep handles compliance configuration
type ComplianceStep struct{}

// Execute runs the compliance configuration step
func (s *ComplianceStep) Execute(ctx *WizardContext) error {
	utils.Info("⚖️ Compliance Configuration")

	// Enable compliance
	enabled := s.promptBool(ctx, "Enable Compliance Management", true)
	ctx.Config.Compliance.Enabled = enabled

	if !enabled {
		return nil
	}

	// Standards
	standards := s.promptMultiChoice(ctx, "Compliance Standards", []string{
		"SOC2", "ISO27001", "PCI-DSS", "HIPAA", "GDPR", "CCPA", "FedRAMP",
	}, []string{"SOC2", "ISO27001"})
	ctx.Config.Compliance.Standards = standards

	// Validation
	validation := s.promptBool(ctx, "Enable Compliance Validation", true)
	ctx.Config.Compliance.Validation.Enabled = validation

	if validation {
		rules := s.promptMultiChoice(ctx, "Validation Rules", []string{
			"security", "quality", "documentation", "testing", "licensing",
		}, []string{"security", "quality", "documentation"})
		ctx.Config.Compliance.Validation.Rules = rules
	}

	// Reporting
	reporting := s.promptBool(ctx, "Enable Compliance Reporting", true)
	ctx.Config.Compliance.Reporting.Enabled = reporting

	if reporting {
		frequency := s.promptChoice(ctx, "Compliance Report Frequency", []string{
			"weekly", "monthly", "quarterly", "annually",
		}, "monthly")
		ctx.Config.Compliance.Reporting.Frequency = frequency

		formats := s.promptMultiChoice(ctx, "Report Formats", []string{
			"json", "html", "pdf", "docx", "xlsx",
		}, []string{"json", "pdf"})
		ctx.Config.Compliance.Reporting.Formats = formats
	}

	return nil
}

// GetName returns the step name
func (s *ComplianceStep) GetName() string {
	return "Compliance Configuration"
}

// GetDescription returns the step description
func (s *ComplianceStep) GetDescription() string {
	return "Configure compliance and governance"
}

// IsRequired indicates if this step is required
func (s *ComplianceStep) IsRequired() bool {
	return false
}

// GetDependencies returns step dependencies
func (s *ComplianceStep) GetDependencies() []string {
	return []string{}
}

func (s *ComplianceStep) promptBool(ctx *WizardContext, prompt string, defaultValue bool) bool {
	defaultStr := "n"
	if defaultValue {
		defaultStr = "y"
	}

	for {
		fmt.Printf("%s (y/n) [%s]: ", prompt, defaultStr)
		ctx.Scanner.Scan()
		input := strings.ToLower(strings.TrimSpace(ctx.Scanner.Text()))

		if input == "" {
			return defaultValue
		}

		if input == "skip" {
			return defaultValue
		}

		if input == "y" || input == "yes" {
			return true
		}

		if input == "n" || input == "no" {
			return false
		}

		utils.Error("Please enter y or n")
	}
}

func (s *ComplianceStep) promptChoice(ctx *WizardContext, prompt string, choices []string, defaultValue string) string {
	fmt.Printf("%s:\n", prompt)
	for i, choice := range choices {
		marker := " "
		if choice == defaultValue {
			marker = "*"
		}
		fmt.Printf("  %s %d. %s\n", marker, i+1, choice)
	}

	for {
		fmt.Printf("Select option (1-%d) [%s]: ", len(choices), defaultValue)
		ctx.Scanner.Scan()
		input := strings.TrimSpace(ctx.Scanner.Text())

		if input == "" {
			return defaultValue
		}

		if input == "skip" {
			return defaultValue
		}

		if idx, err := strconv.Atoi(input); err == nil && idx >= 1 && idx <= len(choices) {
			return choices[idx-1]
		}

		utils.Error("Invalid selection. Please choose 1-%d.", len(choices))
	}
}

func (s *ComplianceStep) promptMultiChoice(ctx *WizardContext, prompt string, choices []string, defaultValues []string) []string {
	fmt.Printf("%s (select multiple by entering numbers separated by commas):\n", prompt)
	for i, choice := range choices {
		marker := " "
		for _, def := range defaultValues {
			if choice == def {
				marker = "*"
				break
			}
		}
		fmt.Printf("  %s %d. %s\n", marker, i+1, choice)
	}

	defaultStr := ""
	if len(defaultValues) > 0 {
		defaultStr = strings.Join(defaultValues, ", ")
	}

	for {
		fmt.Printf("Select options (e.g., 1,3,5) [%s]: ", defaultStr)
		ctx.Scanner.Scan()
		input := strings.TrimSpace(ctx.Scanner.Text())

		if input == "" {
			return defaultValues
		}

		if input == "skip" {
			return defaultValues
		}

		var selected []string
		for _, part := range strings.Split(input, ",") {
			part = strings.TrimSpace(part)
			if idx, err := strconv.Atoi(part); err == nil && idx >= 1 && idx <= len(choices) {
				selected = append(selected, choices[idx-1])
			}
		}

		if len(selected) > 0 {
			return selected
		}

		utils.Error("Invalid selection. Please enter valid numbers separated by commas.")
	}
}

// SummaryStep provides configuration summary and confirmation
type SummaryStep struct{}

// Execute runs the summary and confirmation step
func (s *SummaryStep) Execute(ctx *WizardContext) error {
	utils.Info("📋 Configuration Summary")

	utils.Info("")
	utils.Info("🏢 Organization:")
	fmt.Printf("  Name: %s\n", ctx.Config.Organization.Name)
	fmt.Printf("  Domain: %s\n", ctx.Config.Organization.Domain)
	fmt.Printf("  Region: %s\n", ctx.Config.Organization.Region)

	utils.Info("")
	utils.Info("🔒 Security:")
	fmt.Printf("  Level: %s\n", ctx.Config.Security.Level)
	fmt.Printf("  MFA Enabled: %t\n", ctx.Config.Security.Authentication.MFA.Enabled)
	fmt.Printf("  Scanning Enabled: %t\n", ctx.Config.Security.Scanning.Enabled)

	utils.Info("")
	utils.Info("🔌 Integrations:")
	fmt.Printf("  Enabled: %t\n", ctx.Config.Integrations.Enabled)
	if ctx.Config.Integrations.Enabled {
		fmt.Printf("  Configured: %d\n", len(ctx.Config.Integrations.Providers))
		for name := range ctx.Config.Integrations.Providers {
			fmt.Printf("    - %s\n", name)
		}
	}

	utils.Info("")
	utils.Info("🔄 Workflows:")
	fmt.Printf("  Enabled: %t\n", ctx.Config.Workflows.Enabled)
	if ctx.Config.Workflows.Enabled {
		fmt.Printf("  Directory: %s\n", ctx.Config.Workflows.Directory)
		fmt.Printf("  Scheduler: %t\n", ctx.Config.Workflows.Scheduler.Enabled)
	}

	utils.Info("")
	utils.Info("📊 Analytics:")
	fmt.Printf("  Enabled: %t\n", ctx.Config.Analytics.Enabled)
	if ctx.Config.Analytics.Enabled {
		fmt.Printf("  Collectors: %d\n", len(ctx.Config.Analytics.Collectors))
		fmt.Printf("  Reporting: %t\n", ctx.Config.Analytics.Reporting.Enabled)
	}

	utils.Info("")
	utils.Info("🚀 Deployment:")
	fmt.Printf("  Enabled: %t\n", ctx.Config.Deployment.Enabled)
	if ctx.Config.Deployment.Enabled {
		fmt.Printf("  Environments: %d\n", len(ctx.Config.Deployment.Environments))
		fmt.Printf("  Approval Required: %t\n", ctx.Config.Deployment.Approval.Enabled)
	}

	utils.Info("")
	utils.Info("⚖️ Compliance:")
	fmt.Printf("  Enabled: %t\n", ctx.Config.Compliance.Enabled)
	if ctx.Config.Compliance.Enabled {
		fmt.Printf("  Standards: %s\n", strings.Join(ctx.Config.Compliance.Standards, ", "))
		fmt.Printf("  Validation: %t\n", ctx.Config.Compliance.Validation.Enabled)
	}

	utils.Info("")
	fmt.Print("Does this configuration look correct? (y/n): ")
	ctx.Scanner.Scan()
	input := strings.ToLower(strings.TrimSpace(ctx.Scanner.Text()))

	if input != "y" && input != "yes" {
		return fmt.Errorf("configuration rejected by user")
	}

	return nil
}

// GetName returns the step name
func (s *SummaryStep) GetName() string {
	return "Configuration Summary"
}

// GetDescription returns the step description
func (s *SummaryStep) GetDescription() string {
	return "Review and confirm configuration"
}

// IsRequired indicates if this step is required
func (s *SummaryStep) IsRequired() bool {
	return true
}

// GetDependencies returns step dependencies
func (s *SummaryStep) GetDependencies() []string {
	return []string{}
}

// FinalizationStep handles the final step of the wizard process
type FinalizationStep struct{}

// Execute performs the finalization step of the wizard
func (s *FinalizationStep) Execute(ctx *WizardContext) error {
	utils.Info("💾 Finalizing Configuration")

	// Update metadata
	ctx.Config.Metadata.UpdatedAt = time.Now()

	// Save configuration
	if err := saveEnterpriseConfiguration(ctx.Config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	// Initialize sub-configurations
	if err := initializeSubConfigurations(ctx.Config); err != nil {
		return fmt.Errorf("failed to initialize sub-configurations: %w", err)
	}

	utils.Success("✅ Configuration saved successfully")
	utils.Info("📁 Configuration directory: .mage/enterprise/")

	return nil
}

// GetName returns the name of the finalization step
func (s *FinalizationStep) GetName() string {
	return "Finalization"
}

// GetDescription returns the description of the finalization step
func (s *FinalizationStep) GetDescription() string {
	return "Save and finalize configuration"
}

// IsRequired returns whether the finalization step is required
func (s *FinalizationStep) IsRequired() bool {
	return true
}

// GetDependencies returns the dependencies for the finalization step
func (s *FinalizationStep) GetDependencies() []string {
	return []string{}
}

// Specialized Wizards

// ProjectWizard handles project-specific configuration wizards
type ProjectWizard struct {
	BaseWizard
}

// NewProjectWizard creates a new project configuration wizard
func NewProjectWizard() *ProjectWizard {
	return &ProjectWizard{
		BaseWizard: BaseWizard{
			Name:        "Project Configuration",
			Description: "Configure project-specific settings",
			Steps:       []WizardStep{
				// Project-specific steps would go here
			},
		},
	}
}

// Run executes the project configuration wizard
func (w *ProjectWizard) Run() error {
	utils.Info("📁 Project Configuration Wizard")
	// Implementation for project wizard
	return nil
}

// GetName returns the name of the project wizard
func (w *ProjectWizard) GetName() string {
	return w.Name
}

// GetDescription returns the description of the project wizard
func (w *ProjectWizard) GetDescription() string {
	return w.Description
}

// IntegrationWizard handles integration configuration wizards
type IntegrationWizard struct {
	BaseWizard
}

// NewIntegrationWizard creates a new integration configuration wizard
func NewIntegrationWizard() *IntegrationWizard {
	return &IntegrationWizard{
		BaseWizard: BaseWizard{
			Name:        "Integration Setup",
			Description: "Configure enterprise integrations",
			Steps:       []WizardStep{
				// Integration-specific steps would go here
			},
		},
	}
}

// Run executes the integration configuration wizard
func (w *IntegrationWizard) Run() error {
	utils.Info("🔌 Integration Setup Wizard")
	// Implementation for integration wizard
	return nil
}

// GetName returns the name of the integration wizard
func (w *IntegrationWizard) GetName() string {
	return w.Name
}

// GetDescription returns the description of the integration wizard
func (w *IntegrationWizard) GetDescription() string {
	return w.Description
}

// SecurityWizard handles security configuration wizards
type SecurityWizard struct {
	BaseWizard
}

// NewSecurityWizard creates a new security configuration wizard
func NewSecurityWizard() *SecurityWizard {
	return &SecurityWizard{
		BaseWizard: BaseWizard{
			Name:        "Security Configuration",
			Description: "Configure security settings",
			Steps:       []WizardStep{
				// Security-specific steps would go here
			},
		},
	}
}

// Run executes the security configuration wizard
func (w *SecurityWizard) Run() error {
	utils.Info("🔒 Security Configuration Wizard")
	// Implementation for security wizard
	return nil
}

// GetName returns the name of the security wizard
func (w *SecurityWizard) GetName() string {
	return w.Name
}

// GetDescription returns the description of the security wizard
func (w *SecurityWizard) GetDescription() string {
	return w.Description
}

// WorkflowWizard provides workflow configuration wizard functionality
type WorkflowWizard struct {
	BaseWizard
}

// NewWorkflowWizard creates a new WorkflowWizard instance
func NewWorkflowWizard() *WorkflowWizard {
	return &WorkflowWizard{
		BaseWizard: BaseWizard{
			Name:        "Workflow Configuration",
			Description: "Configure workflow settings",
			Steps:       []WizardStep{
				// Workflow-specific steps would go here
			},
		},
	}
}

// Run executes the workflow configuration wizard
func (w *WorkflowWizard) Run() error {
	utils.Info("🔄 Workflow Configuration Wizard")
	// Implementation for workflow wizard
	return nil
}

// GetName returns the name of the workflow wizard
func (w *WorkflowWizard) GetName() string {
	return w.Name
}

// GetDescription returns the description of the workflow wizard
func (w *WorkflowWizard) GetDescription() string {
	return w.Description
}

// DeploymentWizard provides deployment configuration wizard functionality
type DeploymentWizard struct {
	BaseWizard
}

// NewDeploymentWizard creates a new DeploymentWizard instance
func NewDeploymentWizard() *DeploymentWizard {
	return &DeploymentWizard{
		BaseWizard: BaseWizard{
			Name:        "Deployment Configuration",
			Description: "Configure deployment settings",
			Steps:       []WizardStep{
				// Deployment-specific steps would go here
			},
		},
	}
}

// Run executes the deployment configuration wizard
func (w *DeploymentWizard) Run() error {
	utils.Info("🚀 Deployment Configuration Wizard")
	// Implementation for deployment wizard
	return nil
}

// GetName returns the name of the deployment wizard
func (w *DeploymentWizard) GetName() string {
	return w.Name
}

// GetDescription returns the description of the deployment wizard
func (w *DeploymentWizard) GetDescription() string {
	return w.Description
}

// Validation functions

func validateNonEmpty(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("value cannot be empty")
	}
	return nil
}

func validateDomain(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("domain cannot be empty")
	}

	// Simple domain validation
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$`)
	if !domainRegex.MatchString(value) {
		return fmt.Errorf("invalid domain format")
	}

	return nil
}

func validateTimezone(value string) error {
	// Simple timezone validation
	validTimezones := []string{
		"UTC", "America/New_York", "America/Los_Angeles", "Europe/London",
		"Europe/Paris", "Asia/Tokyo", "Asia/Shanghai", "Australia/Sydney",
	}

	for _, tz := range validTimezones {
		if value == tz {
			return nil
		}
	}

	// Allow any timezone format
	return nil
}

func validateDuration(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("duration cannot be empty")
	}

	_, err := time.ParseDuration(value)
	if err != nil {
		return fmt.Errorf("invalid duration format (e.g., 30m, 1h, 2h30m)")
	}

	return nil
}
