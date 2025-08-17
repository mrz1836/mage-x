// Package mage provides enterprise configuration management
package mage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/utils"
	"gopkg.in/yaml.v3"
)

const (
	responseYes = "yes"
)

// Static errors for err113 compliance
var (
	ErrConfigValidationFailed       = errors.New("configuration validation failed")
	ErrInputEnvRequired             = errors.New("INPUT environment variable is required")
	ErrImportedConfigInvalid        = errors.New("imported configuration is invalid")
	ErrEnterpriseConfigNotFoundInit = errors.New("enterprise configuration not found. Run 'magex enterprise-config:init' first")
	ErrInvalidSelection             = errors.New("invalid selection")
	ErrUnknownSection               = errors.New("unknown section")
	ErrUnsupportedConfigFormat      = errors.New("unsupported format")
)

// EnterpriseConfigNamespace namespace for enterprise configuration management
type EnterpriseConfigNamespace mg.Namespace

// Init initializes enterprise configuration
func (EnterpriseConfigNamespace) Init() error {
	utils.Header("🏢 Enterprise Configuration Initialization")

	// Create enterprise configuration directory
	if err := os.MkdirAll(enterpriseDir, 0o750); err != nil {
		return fmt.Errorf("failed to create enterprise directory: %w", err)
	}

	// Initialize enterprise configuration
	config := NewEnterpriseConfiguration()

	// Run interactive configuration wizard
	if utils.GetEnv("INTERACTIVE", approvalTrue) == approvalTrue {
		runEnterpriseConfigWizard(config)
	}

	// Save enterprise configuration
	if err := saveEnterpriseConfiguration(config); err != nil {
		return fmt.Errorf("failed to save enterprise configuration: %w", err)
	}

	// Initialize additional configurations
	if err := initializeSubConfigurations(config); err != nil {
		return fmt.Errorf("failed to initialize sub-configurations: %w", err)
	}

	utils.Success("✅ Enterprise configuration initialized successfully")
	utils.Info("📁 Configuration directory: %s", enterpriseDir)

	return nil
}

// Validate validates enterprise configuration
func (EnterpriseConfigNamespace) Validate() error {
	utils.Header("🔍 Enterprise Configuration Validation")

	// Load enterprise configuration
	config, err := loadEnterpriseConfiguration()
	if err != nil {
		return fmt.Errorf("failed to load enterprise configuration: %w", err)
	}

	// Validate configuration
	validator := NewConfigurationValidator()
	results := validator.Validate(config)

	// Display validation results
	displayValidationResults(results)

	// Check for errors
	if results.HasErrors() {
		return fmt.Errorf("%w with %d errors", ErrConfigValidationFailed, len(results.Errors))
	}

	utils.Success("✅ Enterprise configuration is valid")
	return nil
}

// Update updates enterprise configuration
func (EnterpriseConfigNamespace) Update() error {
	utils.Header("🔄 Enterprise Configuration Update")

	// Load current configuration
	config, err := loadEnterpriseConfiguration()
	if err != nil {
		return fmt.Errorf("failed to load current configuration: %w", err)
	}

	// Get update section
	section := utils.GetEnv("SECTION", "")
	if section == "" {
		return runInteractiveConfigUpdate(config)
	}

	// Update specific section
	return updateConfigurationSection(config, section)
}

// Export exports enterprise configuration
func (EnterpriseConfigNamespace) Export() error {
	utils.Header("📤 Enterprise Configuration Export")

	// Load enterprise configuration
	config, err := loadEnterpriseConfiguration()
	if err != nil {
		return fmt.Errorf("failed to load enterprise configuration: %w", err)
	}

	// Get export parameters
	format := utils.GetEnv("FORMAT", "yaml")
	outputFile := utils.GetEnv("OUTPUT", "enterprise-config.yaml")

	// Export configuration
	if err := exportConfiguration(config, format, outputFile); err != nil {
		return fmt.Errorf("failed to export configuration: %w", err)
	}

	utils.Success("✅ Enterprise configuration exported to: %s", outputFile)
	return nil
}

// Import imports enterprise configuration
func (EnterpriseConfigNamespace) Import() error {
	utils.Header("📥 Enterprise Configuration Import")

	// Get import parameters
	inputFile := utils.GetEnv("INPUT", "")
	if inputFile == "" {
		return ErrInputEnvRequired
	}

	// Import configuration
	config, err := importConfiguration(inputFile)
	if err != nil {
		return fmt.Errorf("failed to import configuration: %w", err)
	}

	// Validate imported configuration
	validator := NewConfigurationValidator()
	results := validator.Validate(config)

	if results.HasErrors() {
		return fmt.Errorf("%w: %d errors", ErrImportedConfigInvalid, len(results.Errors))
	}

	// Save imported configuration
	if err := saveEnterpriseConfiguration(config); err != nil {
		return fmt.Errorf("failed to save imported configuration: %w", err)
	}

	utils.Success("✅ Enterprise configuration imported successfully")
	return nil
}

// Schema generates configuration schema
func (EnterpriseConfigNamespace) Schema() error {
	utils.Header("📋 Enterprise Configuration Schema")

	// Generate JSON schema
	schema := generateEnterpriseConfigurationSchema()

	// Get output format
	format := utils.GetEnv("FORMAT", "json")
	outputFile := utils.GetEnv("OUTPUT", "enterprise-config-schema.json")

	// Save schema
	var data []byte
	var err error

	switch format {
	case "yaml":
		data, err = yaml.Marshal(schema)
	case "json":
		data, err = json.MarshalIndent(schema, "", "  ")
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedConfigFormat, format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal schema: %w", err)
	}

	fileOps := fileops.New()
	if err := fileOps.File.WriteFile(outputFile, data, 0o644); err != nil {
		return fmt.Errorf("failed to write schema: %w", err)
	}

	utils.Success("✅ Configuration schema generated: %s", outputFile)
	return nil
}

// Enhanced Enterprise Configuration Types

// ECConfigMetadata contains metadata for enterprise configuration
type ECConfigMetadata struct {
	Version     string            `yaml:"version" json:"version"`
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description" json:"description"`
	CreatedAt   time.Time         `yaml:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `yaml:"updated_at" json:"updated_at"`
	Tags        []string          `yaml:"tags" json:"tags"`
	Labels      map[string]string `yaml:"labels" json:"labels"`
}

// EnterpriseConfiguration represents complete enterprise settings
type EnterpriseConfiguration struct {
	Metadata      ECConfigMetadata    `yaml:"metadata" json:"metadata"`
	Organization  OrganizationConfig  `yaml:"organization" json:"organization"`
	Security      SecurityConfig      `yaml:"security" json:"security"`
	Integrations  IntegrationsConfig  `yaml:"integrations" json:"integrations"`
	Workflows     WorkflowsConfig     `yaml:"workflows" json:"workflows"`
	Analytics     AnalyticsConfig     `yaml:"analytics" json:"analytics"`
	Audit         ECECAuditConfig     `yaml:"audit" json:"audit"`
	CLI           CLIConfig           `yaml:"cli" json:"cli"`
	Repositories  RepositoriesConfig  `yaml:"repositories" json:"repositories"`
	Compliance    ComplianceConfig    `yaml:"compliance" json:"compliance"`
	Monitoring    ECMonitoringConfig  `yaml:"monitoring" json:"monitoring"`
	Deployment    DeploymentConfig    `yaml:"deployment" json:"deployment"`
	Backup        BackupConfig        `yaml:"backup" json:"backup"`
	Notifications NotificationsConfig `yaml:"notifications" json:"notifications"`
}

// ECECConfigMetadata contains enhanced configuration metadata
type ECECConfigMetadata struct {
	Version     string            `yaml:"version" json:"version"`
	CreatedAt   time.Time         `yaml:"created_at" json:"created_at"`
	UpdatedAt   time.Time         `yaml:"updated_at" json:"updated_at"`
	CreatedBy   string            `yaml:"created_by" json:"created_by"`
	UpdatedBy   string            `yaml:"updated_by" json:"updated_by"`
	Description string            `yaml:"description" json:"description"`
	Tags        []string          `yaml:"tags" json:"tags"`
	Labels      map[string]string `yaml:"labels" json:"labels"`
}

// OrganizationConfig holds organization-specific settings
type OrganizationConfig struct {
	Name        string            `yaml:"name" json:"name"`
	Domain      string            `yaml:"domain" json:"domain"`
	Region      string            `yaml:"region" json:"region"`
	Timezone    string            `yaml:"timezone" json:"timezone"`
	Language    string            `yaml:"language" json:"language"`
	Currency    string            `yaml:"currency" json:"currency"`
	Metadata    map[string]string `yaml:"metadata" json:"metadata"`
	Contacts    []ContactInfo     `yaml:"contacts" json:"contacts"`
	Departments []Department      `yaml:"departments" json:"departments"`
}

// ContactInfo represents contact information
type ContactInfo struct {
	Name  string `yaml:"name" json:"name"`
	Email string `yaml:"email" json:"email"`
	Role  string `yaml:"role" json:"role"`
	Phone string `yaml:"phone" json:"phone"`
}

// Department represents an organizational department
type Department struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description" json:"description"`
	Manager     string   `yaml:"manager" json:"manager"`
	Members     []string `yaml:"members" json:"members"`
}

// SecurityConfig contains security-related settings
type SecurityConfig struct {
	Level          string               `yaml:"level" json:"level"`
	Encryption     EncryptionConfig     `yaml:"encryption" json:"encryption"`
	Authentication AuthenticationConfig `yaml:"authentication" json:"authentication"`
	Authorization  AuthorizationConfig  `yaml:"authorization" json:"authorization"`
	Compliance     ComplianceSettings   `yaml:"compliance" json:"compliance"`
	Scanning       ScanningConfig       `yaml:"scanning" json:"scanning"`
	Policies       []ECSecurityPolicy   `yaml:"policies" json:"policies"`
	Secrets        SecretsConfig        `yaml:"secrets" json:"secrets"`
}

// EncryptionConfig holds encryption settings
type EncryptionConfig struct {
	Algorithm string            `yaml:"algorithm" json:"algorithm"`
	KeySize   int               `yaml:"key_size" json:"key_size"`
	Settings  map[string]string `yaml:"settings" json:"settings"`
}

// AuthenticationConfig contains authentication settings
type AuthenticationConfig struct {
	Methods     []string          `yaml:"methods" json:"methods"`
	TokenExpiry string            `yaml:"token_expiry" json:"token_expiry"`
	MFA         MFAConfig         `yaml:"mfa" json:"mfa"`
	Settings    map[string]string `yaml:"settings" json:"settings"`
}

// MFAConfig holds multi-factor authentication settings
type MFAConfig struct {
	Enabled  bool     `yaml:"enabled" json:"enabled"`
	Methods  []string `yaml:"methods" json:"methods"`
	Required bool     `yaml:"required" json:"required"`
}

// AuthorizationConfig contains authorization settings
type AuthorizationConfig struct {
	Model    string            `yaml:"model" json:"model"`
	Rules    []AuthRule        `yaml:"rules" json:"rules"`
	Settings map[string]string `yaml:"settings" json:"settings"`
}

// AuthRule represents an authorization rule
type AuthRule struct {
	Resource   string   `yaml:"resource" json:"resource"`
	Actions    []string `yaml:"actions" json:"actions"`
	Principals []string `yaml:"principals" json:"principals"`
	Effect     string   `yaml:"effect" json:"effect"`
}

// ComplianceSettings holds compliance-related configuration
type ComplianceSettings struct {
	Standards []string          `yaml:"standards" json:"standards"`
	Auditing  bool              `yaml:"auditing" json:"auditing"`
	Reporting bool              `yaml:"reporting" json:"reporting"`
	Settings  map[string]string `yaml:"settings" json:"settings"`
}

// ScanningConfig contains security scanning settings
type ScanningConfig struct {
	Enabled    bool              `yaml:"enabled" json:"enabled"`
	Frequency  string            `yaml:"frequency" json:"frequency"`
	Tools      []string          `yaml:"tools" json:"tools"`
	Thresholds ScanThresholds    `yaml:"thresholds" json:"thresholds"`
	Settings   map[string]string `yaml:"settings" json:"settings"`
}

// ScanThresholds defines security scan thresholds
type ScanThresholds struct {
	Critical int `yaml:"critical" json:"critical"`
	High     int `yaml:"high" json:"high"`
	Medium   int `yaml:"medium" json:"medium"`
	Low      int `yaml:"low" json:"low"`
}

// ECSecurityPolicy defines enterprise security policies
type ECSecurityPolicy struct {
	ID          string            `yaml:"id" json:"id"`
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description" json:"description"`
	Type        string            `yaml:"type" json:"type"`
	Rules       []string          `yaml:"rules" json:"rules"`
	Severity    string            `yaml:"severity" json:"severity"`
	Enabled     bool              `yaml:"enabled" json:"enabled"`
	Settings    map[string]string `yaml:"settings" json:"settings"`
}

// SecretsConfig contains secrets management settings
type SecretsConfig struct {
	Provider string            `yaml:"provider" json:"provider"`
	Vault    string            `yaml:"vault" json:"vault"`
	Rotation string            `yaml:"rotation" json:"rotation"`
	Settings map[string]string `yaml:"settings" json:"settings"`
}

// IntegrationsConfig contains integration configuration settings
type IntegrationsConfig struct {
	Enabled      bool                           `yaml:"enabled" json:"enabled"`
	Providers    map[string]IntegrationProvider `yaml:"providers" json:"providers"`
	Webhooks     WebhooksConfig                 `yaml:"webhooks" json:"webhooks"`
	SyncSettings SyncSettings                   `yaml:"sync_settings" json:"sync_settings"`
}

// IntegrationProvider represents a third-party integration provider
type IntegrationProvider struct {
	Type        string            `yaml:"type" json:"type"`
	Enabled     bool              `yaml:"enabled" json:"enabled"`
	Settings    map[string]string `yaml:"settings" json:"settings"`
	Credentials map[string]string `yaml:"credentials" json:"credentials"`
	Endpoints   map[string]string `yaml:"endpoints" json:"endpoints"`
}

// WebhooksConfig contains webhook configuration settings
type WebhooksConfig struct {
	Enabled   bool                       `yaml:"enabled" json:"enabled"`
	Endpoints map[string]WebhookEndpoint `yaml:"endpoints" json:"endpoints"`
	Security  WebhookSecurity            `yaml:"security" json:"security"`
}

// WebhookEndpoint represents a webhook endpoint configuration
type WebhookEndpoint struct {
	URL     string            `yaml:"url" json:"url"`
	Events  []string          `yaml:"events" json:"events"`
	Headers map[string]string `yaml:"headers" json:"headers"`
	Timeout string            `yaml:"timeout" json:"timeout"`
}

// WebhookSecurity contains webhook security settings
type WebhookSecurity struct {
	SigningKey string   `yaml:"signing_key" json:"signing_key"`
	Algorithm  string   `yaml:"algorithm" json:"algorithm"`
	Headers    []string `yaml:"headers" json:"headers"`
}

// SyncSettings contains synchronization configuration
type SyncSettings struct {
	Frequency string            `yaml:"frequency" json:"frequency"`
	Batch     int               `yaml:"batch" json:"batch"`
	Retry     RetryConfig       `yaml:"retry" json:"retry"`
	Settings  map[string]string `yaml:"settings" json:"settings"`
}

// RetryConfig contains retry mechanism settings
type RetryConfig struct {
	MaxAttempts int    `yaml:"max_attempts" json:"max_attempts"`
	Backoff     string `yaml:"backoff" json:"backoff"`
	Timeout     string `yaml:"timeout" json:"timeout"`
}

// WorkflowsConfig contains workflow configuration settings
type WorkflowsConfig struct {
	Enabled   bool                              `yaml:"enabled" json:"enabled"`
	Directory string                            `yaml:"directory" json:"directory"`
	Templates map[string]ConfigWorkflowTemplate `yaml:"templates" json:"templates"`
	Scheduler SchedulerConfig                   `yaml:"scheduler" json:"scheduler"`
	Execution ExecutionConfig                   `yaml:"execution" json:"execution"`
}

// ConfigWorkflowTemplate represents a workflow template configuration
type ConfigWorkflowTemplate struct {
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description" json:"description"`
	Steps       []WorkflowStep    `yaml:"steps" json:"steps"`
	Variables   map[string]string `yaml:"variables" json:"variables"`
}

// SchedulerConfig contains scheduler configuration settings
type SchedulerConfig struct {
	Enabled  bool              `yaml:"enabled" json:"enabled"`
	Engine   string            `yaml:"engine" json:"engine"`
	Settings map[string]string `yaml:"settings" json:"settings"`
}

// ExecutionConfig contains execution configuration settings
type ExecutionConfig struct {
	Timeout     string            `yaml:"timeout" json:"timeout"`
	Parallelism int               `yaml:"parallelism" json:"parallelism"`
	Retry       RetryConfig       `yaml:"retry" json:"retry"`
	Settings    map[string]string `yaml:"settings" json:"settings"`
}

// AnalyticsConfig contains analytics configuration settings
type AnalyticsConfig struct {
	Enabled    bool                       `yaml:"enabled" json:"enabled"`
	Collectors map[string]CollectorConfig `yaml:"collectors" json:"collectors"`
	Storage    StorageConfig              `yaml:"storage" json:"storage"`
	Reporting  ReportingConfig            `yaml:"reporting" json:"reporting"`
}

// CollectorConfig contains collector configuration settings
type CollectorConfig struct {
	Type     string            `yaml:"type" json:"type"`
	Enabled  bool              `yaml:"enabled" json:"enabled"`
	Settings map[string]string `yaml:"settings" json:"settings"`
}

// StorageConfig contains storage configuration settings
type StorageConfig struct {
	Type       string            `yaml:"type" json:"type"`
	Retention  string            `yaml:"retention" json:"retention"`
	Encryption bool              `yaml:"encryption" json:"encryption"`
	Settings   map[string]string `yaml:"settings" json:"settings"`
}

// ReportingConfig contains reporting configuration settings
type ReportingConfig struct {
	Enabled   bool              `yaml:"enabled" json:"enabled"`
	Frequency string            `yaml:"frequency" json:"frequency"`
	Formats   []string          `yaml:"formats" json:"formats"`
	Settings  map[string]string `yaml:"settings" json:"settings"`
}

// ECECAuditConfig contains enterprise audit configuration settings
type ECECAuditConfig struct {
	Enabled    bool              `yaml:"enabled" json:"enabled"`
	Level      string            `yaml:"level" json:"level"`
	Storage    StorageConfig     `yaml:"storage" json:"storage"`
	Retention  string            `yaml:"retention" json:"retention"`
	Encryption bool              `yaml:"encryption" json:"encryption"`
	Settings   map[string]string `yaml:"settings" json:"settings"`
}

// CLIConfig contains CLI configuration settings
type CLIConfig struct {
	Interactive bool              `yaml:"interactive" json:"interactive"`
	Colors      bool              `yaml:"colors" json:"colors"`
	Verbose     bool              `yaml:"verbose" json:"verbose"`
	Timeout     string            `yaml:"timeout" json:"timeout"`
	Aliases     map[string]string `yaml:"aliases" json:"aliases"`
	Completion  CompletionConfig  `yaml:"completion" json:"completion"`
	Dashboard   ECDashboardConfig `yaml:"dashboard" json:"dashboard"`
}

// CompletionConfig contains shell completion configuration settings
type CompletionConfig struct {
	Enabled  bool              `yaml:"enabled" json:"enabled"`
	Shell    string            `yaml:"shell" json:"shell"`
	Settings map[string]string `yaml:"settings" json:"settings"`
}

// ECDashboardConfig contains enterprise dashboard configuration settings
type ECDashboardConfig struct {
	Enabled  bool              `yaml:"enabled" json:"enabled"`
	Refresh  string            `yaml:"refresh" json:"refresh"`
	Widgets  []string          `yaml:"widgets" json:"widgets"`
	Settings map[string]string `yaml:"settings" json:"settings"`
}

// RepositoriesConfig contains repository management configuration settings
type RepositoriesConfig struct {
	Discovery  DiscoveryConfig                 `yaml:"discovery" json:"discovery"`
	Management ManagementConfig                `yaml:"management" json:"management"`
	Templates  map[string]ECRepositoryTemplate `yaml:"templates" json:"templates"`
}

// DiscoveryConfig contains repository discovery configuration settings
type DiscoveryConfig struct {
	Enabled  bool              `yaml:"enabled" json:"enabled"`
	Sources  []string          `yaml:"sources" json:"sources"`
	Filters  []string          `yaml:"filters" json:"filters"`
	Settings map[string]string `yaml:"settings" json:"settings"`
}

// ManagementConfig contains repository management configuration settings
type ManagementConfig struct {
	Sync     SyncSettings      `yaml:"sync" json:"sync"`
	Backup   BackupConfig      `yaml:"backup" json:"backup"`
	Settings map[string]string `yaml:"settings" json:"settings"`
}

// ECRepositoryTemplate contains enterprise repository template settings
type ECRepositoryTemplate struct {
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description" json:"description"`
	Language    string            `yaml:"language" json:"language"`
	Framework   string            `yaml:"framework" json:"framework"`
	Settings    map[string]string `yaml:"settings" json:"settings"`
}

// ComplianceConfig contains compliance configuration settings
type ComplianceConfig struct {
	Enabled    bool                          `yaml:"enabled" json:"enabled"`
	Standards  []string                      `yaml:"standards" json:"standards"`
	Policies   map[string]ECCompliancePolicy `yaml:"policies" json:"policies"`
	Reporting  ReportingConfig               `yaml:"reporting" json:"reporting"`
	Validation ValidationConfig              `yaml:"validation" json:"validation"`
}

// ECCompliancePolicy contains enterprise compliance policy settings
type ECCompliancePolicy struct {
	Name        string             `yaml:"name" json:"name"`
	Description string             `yaml:"description" json:"description"`
	Rules       []ECComplianceRule `yaml:"rules" json:"rules"`
	Settings    map[string]string  `yaml:"settings" json:"settings"`
}

// ECComplianceRule contains enterprise compliance rule settings
type ECComplianceRule struct {
	ID          string            `yaml:"id" json:"id"`
	Description string            `yaml:"description" json:"description"`
	Severity    string            `yaml:"severity" json:"severity"`
	Settings    map[string]string `yaml:"settings" json:"settings"`
}

// ValidationConfig contains validation configuration settings
type ValidationConfig struct {
	Enabled  bool              `yaml:"enabled" json:"enabled"`
	Rules    []string          `yaml:"rules" json:"rules"`
	Settings map[string]string `yaml:"settings" json:"settings"`
}

// ECMonitoringConfig contains enterprise monitoring configuration settings
type ECMonitoringConfig struct {
	Enabled    bool                     `yaml:"enabled" json:"enabled"`
	Metrics    ECMetricsConfig          `yaml:"metrics" json:"metrics"`
	Alerting   AlertingConfig           `yaml:"alerting" json:"alerting"`
	Dashboards map[string]DashboardSpec `yaml:"dashboards" json:"dashboards"`
}

// ECMetricsConfig contains enterprise metrics configuration settings
type ECMetricsConfig struct {
	Enabled    bool              `yaml:"enabled" json:"enabled"`
	Collectors []string          `yaml:"collectors" json:"collectors"`
	Interval   string            `yaml:"interval" json:"interval"`
	Settings   map[string]string `yaml:"settings" json:"settings"`
}

// AlertingConfig contains alerting configuration settings
type AlertingConfig struct {
	Enabled  bool              `yaml:"enabled" json:"enabled"`
	Rules    []ECAlertRule     `yaml:"rules" json:"rules"`
	Channels []string          `yaml:"channels" json:"channels"`
	Settings map[string]string `yaml:"settings" json:"settings"`
}

// ECAlertRule contains enterprise alert rule settings
type ECAlertRule struct {
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description" json:"description"`
	Condition   string            `yaml:"condition" json:"condition"`
	Severity    string            `yaml:"severity" json:"severity"`
	Settings    map[string]string `yaml:"settings" json:"settings"`
}

// DashboardSpec contains dashboard specification settings
type DashboardSpec struct {
	Title       string            `yaml:"title" json:"title"`
	Description string            `yaml:"description" json:"description"`
	Panels      []DashboardPanel  `yaml:"panels" json:"panels"`
	Settings    map[string]string `yaml:"settings" json:"settings"`
}

// DashboardPanel contains dashboard panel settings
type DashboardPanel struct {
	Title    string            `yaml:"title" json:"title"`
	Type     string            `yaml:"type" json:"type"`
	Query    string            `yaml:"query" json:"query"`
	Settings map[string]string `yaml:"settings" json:"settings"`
}

// DeploymentConfig contains deployment configuration settings
type DeploymentConfig struct {
	Enabled      bool                   `yaml:"enabled" json:"enabled"`
	Environments map[string]Environment `yaml:"environments" json:"environments"`
	Pipeline     PipelineConfig         `yaml:"pipeline" json:"pipeline"`
	Approval     ApprovalConfig         `yaml:"approval" json:"approval"`
}

// Environment contains environment configuration settings
type Environment struct {
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description" json:"description"`
	Type        string            `yaml:"type" json:"type"`
	Settings    map[string]string `yaml:"settings" json:"settings"`
}

// PipelineConfig contains pipeline configuration settings
type PipelineConfig struct {
	Stages   []PipelineStage   `yaml:"stages" json:"stages"`
	Triggers []PipelineTrigger `yaml:"triggers" json:"triggers"`
	Settings map[string]string `yaml:"settings" json:"settings"`
}

// PipelineStage contains pipeline stage settings
type PipelineStage struct {
	Name         string            `yaml:"name" json:"name"`
	Description  string            `yaml:"description" json:"description"`
	Commands     []string          `yaml:"commands" json:"commands"`
	Dependencies []string          `yaml:"dependencies" json:"dependencies"`
	Settings     map[string]string `yaml:"settings" json:"settings"`
}

// PipelineTrigger contains pipeline trigger settings
type PipelineTrigger struct {
	Type       string            `yaml:"type" json:"type"`
	Conditions []string          `yaml:"conditions" json:"conditions"`
	Settings   map[string]string `yaml:"settings" json:"settings"`
}

// ApprovalConfig contains approval configuration settings
type ApprovalConfig struct {
	Enabled  bool              `yaml:"enabled" json:"enabled"`
	Required []string          `yaml:"required" json:"required"`
	Settings map[string]string `yaml:"settings" json:"settings"`
}

// BackupConfig contains backup configuration settings
type BackupConfig struct {
	Enabled    bool              `yaml:"enabled" json:"enabled"`
	Schedule   string            `yaml:"schedule" json:"schedule"`
	Retention  string            `yaml:"retention" json:"retention"`
	Storage    StorageConfig     `yaml:"storage" json:"storage"`
	Encryption bool              `yaml:"encryption" json:"encryption"`
	Settings   map[string]string `yaml:"settings" json:"settings"`
}

// NotificationsConfig contains notifications configuration settings
type NotificationsConfig struct {
	Enabled  bool                           `yaml:"enabled" json:"enabled"`
	Channels map[string]NotificationChannel `yaml:"channels" json:"channels"`
	Rules    []NotificationRule             `yaml:"rules" json:"rules"`
	Settings map[string]string              `yaml:"settings" json:"settings"`
}

// NotificationChannel contains notification channel settings
type NotificationChannel struct {
	Type     string            `yaml:"type" json:"type"`
	Enabled  bool              `yaml:"enabled" json:"enabled"`
	Settings map[string]string `yaml:"settings" json:"settings"`
}

// NotificationRule contains notification rule settings
type NotificationRule struct {
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description" json:"description"`
	Conditions  []string          `yaml:"conditions" json:"conditions"`
	Channels    []string          `yaml:"channels" json:"channels"`
	Settings    map[string]string `yaml:"settings" json:"settings"`
}

// Validation Types

// ValidationResults contains validation results
type ValidationResults struct {
	Valid    bool                `json:"valid" yaml:"valid"`
	Errors   []ValidationError   `json:"errors" yaml:"errors"`
	Warnings []ValidationWarning `json:"warnings" yaml:"warnings"`
	Info     []ValidationInfo    `json:"info" yaml:"info"`
}

// ValidationError contains validation error details
type ValidationError struct {
	Field   string `json:"field" yaml:"field"`
	Message string `json:"message" yaml:"message"`
	Code    string `json:"code" yaml:"code"`
}

// ValidationWarning contains validation warning details
type ValidationWarning struct {
	Field   string `json:"field" yaml:"field"`
	Message string `json:"message" yaml:"message"`
	Code    string `json:"code" yaml:"code"`
}

// ValidationInfo contains validation information details
type ValidationInfo struct {
	Field   string `json:"field" yaml:"field"`
	Message string `json:"message" yaml:"message"`
	Code    string `json:"code" yaml:"code"`
}

// Configuration Validator

// ConfigurationValidator validates enterprise configuration
type ConfigurationValidator struct {
	rules []ECValidationRule
}

// ECValidationRule defines the interface for enterprise configuration validation rules
type ECValidationRule interface {
	Validate(config *EnterpriseConfiguration) []ValidationError
}

// Implementation functions

// NewEnterpriseConfiguration creates a new enterprise configuration with default values
func NewEnterpriseConfiguration() *EnterpriseConfiguration {
	now := time.Now()
	return &EnterpriseConfiguration{
		Metadata: ECConfigMetadata{
			Version:     "1.0.0",
			CreatedAt:   now,
			UpdatedAt:   now,
			Description: "MAGE-X Enterprise Configuration",
			Tags:        []string{"enterprise", "mage-x"},
			Labels:      make(map[string]string),
		},
		Organization: OrganizationConfig{
			Name:     "MAGE-X Organization",
			Domain:   "example.com",
			Region:   "us-east-1",
			Timezone: "UTC",
			Language: "en",
			Currency: "USD",
			Metadata: make(map[string]string),
		},
		Security: SecurityConfig{
			Level: "standard",
			Encryption: EncryptionConfig{
				Algorithm: "AES-256",
				KeySize:   256,
				Settings:  make(map[string]string),
			},
			Authentication: AuthenticationConfig{
				Methods:     []string{"token", "oauth"},
				TokenExpiry: "24h",
				MFA: MFAConfig{
					Enabled:  false,
					Methods:  []string{"totp"},
					Required: false,
				},
				Settings: make(map[string]string),
			},
			Scanning: ScanningConfig{
				Enabled:   true,
				Frequency: "daily",
				Tools:     []string{"gosec", "govulncheck"},
				Thresholds: ScanThresholds{
					Critical: 0,
					High:     5,
					Medium:   10,
					Low:      20,
				},
				Settings: make(map[string]string),
			},
			Secrets: SecretsConfig{
				Provider: "env",
				Rotation: "manual",
				Settings: make(map[string]string),
			},
		},
		Integrations: IntegrationsConfig{
			Enabled:   true,
			Providers: make(map[string]IntegrationProvider),
			Webhooks: WebhooksConfig{
				Enabled:   false,
				Endpoints: make(map[string]WebhookEndpoint),
				Security: WebhookSecurity{
					Algorithm: "HMAC-SHA256",
					Headers:   []string{"X-Hub-Signature-256"},
				},
			},
			SyncSettings: SyncSettings{
				Frequency: "hourly",
				Batch:     100,
				Retry: RetryConfig{
					MaxAttempts: 3,
					Backoff:     "exponential",
					Timeout:     "30s",
				},
				Settings: make(map[string]string),
			},
		},
		Workflows: WorkflowsConfig{
			Enabled:   true,
			Directory: ".mage/workflows",
			Templates: make(map[string]ConfigWorkflowTemplate),
			Scheduler: SchedulerConfig{
				Enabled:  false,
				Engine:   "cron",
				Settings: make(map[string]string),
			},
			Execution: ExecutionConfig{
				Timeout:     "30m",
				Parallelism: 5,
				Retry: RetryConfig{
					MaxAttempts: 3,
					Backoff:     "exponential",
					Timeout:     "5m",
				},
				Settings: make(map[string]string),
			},
		},
		Analytics: AnalyticsConfig{
			Enabled:    true,
			Collectors: make(map[string]CollectorConfig),
			Storage: StorageConfig{
				Type:       "file",
				Retention:  "90d",
				Encryption: false,
				Settings:   make(map[string]string),
			},
			Reporting: ReportingConfig{
				Enabled:   true,
				Frequency: "weekly",
				Formats:   []string{"json", "html"},
				Settings:  make(map[string]string),
			},
		},
		Audit: ECECAuditConfig{
			Enabled:    true,
			Level:      "info",
			Retention:  "1y",
			Encryption: true,
			Storage: StorageConfig{
				Type:       "sqlite",
				Retention:  "1y",
				Encryption: true,
				Settings:   make(map[string]string),
			},
			Settings: make(map[string]string),
		},
		CLI: CLIConfig{
			Interactive: true,
			Colors:      true,
			Verbose:     false,
			Timeout:     "30m",
			Aliases:     make(map[string]string),
			Completion: CompletionConfig{
				Enabled:  true,
				Shell:    "bash",
				Settings: make(map[string]string),
			},
			Dashboard: ECDashboardConfig{
				Enabled:  true,
				Refresh:  "30s",
				Widgets:  []string{"overview", "metrics", "alerts"},
				Settings: make(map[string]string),
			},
		},
		Repositories: RepositoriesConfig{
			Discovery: DiscoveryConfig{
				Enabled:  true,
				Sources:  []string{"github", "gitlab"},
				Filters:  []string{"go", "active"},
				Settings: make(map[string]string),
			},
			Management: ManagementConfig{
				Sync: SyncSettings{
					Frequency: "daily",
					Batch:     50,
					Retry: RetryConfig{
						MaxAttempts: 3,
						Backoff:     "exponential",
						Timeout:     "10m",
					},
					Settings: make(map[string]string),
				},
				Settings: make(map[string]string),
			},
			Templates: make(map[string]ECRepositoryTemplate),
		},
		Compliance: ComplianceConfig{
			Enabled:   true,
			Standards: []string{"SOC2", "ISO27001"},
			Policies:  make(map[string]ECCompliancePolicy),
			Reporting: ReportingConfig{
				Enabled:   true,
				Frequency: "monthly",
				Formats:   []string{"json", "pdf"},
				Settings:  make(map[string]string),
			},
			Validation: ValidationConfig{
				Enabled:  true,
				Rules:    []string{"security", "quality", "documentation"},
				Settings: make(map[string]string),
			},
		},
		Monitoring: ECMonitoringConfig{
			Enabled: true,
			Metrics: ECMetricsConfig{
				Enabled:    true,
				Collectors: []string{"performance", "build", "test"},
				Interval:   "1m",
				Settings:   make(map[string]string),
			},
			Alerting: AlertingConfig{
				Enabled:  true,
				Rules:    []ECAlertRule{},
				Channels: []string{"email", "slack"},
				Settings: make(map[string]string),
			},
			Dashboards: make(map[string]DashboardSpec),
		},
		Deployment: DeploymentConfig{
			Enabled:      true,
			Environments: make(map[string]Environment),
			Pipeline: PipelineConfig{
				Stages:   []PipelineStage{},
				Triggers: []PipelineTrigger{},
				Settings: make(map[string]string),
			},
			Approval: ApprovalConfig{
				Enabled:  false,
				Required: []string{},
				Settings: make(map[string]string),
			},
		},
		Backup: BackupConfig{
			Enabled:    true,
			Schedule:   "daily",
			Retention:  "30d",
			Encryption: true,
			Storage: StorageConfig{
				Type:       "file",
				Retention:  "30d",
				Encryption: true,
				Settings:   make(map[string]string),
			},
			Settings: make(map[string]string),
		},
		Notifications: NotificationsConfig{
			Enabled:  true,
			Channels: make(map[string]NotificationChannel),
			Rules:    []NotificationRule{},
			Settings: make(map[string]string),
		},
	}
}

func runEnterpriseConfigWizard(config *EnterpriseConfiguration) {
	utils.Info("🧙 Starting Enterprise Configuration Wizard")

	scanner := bufio.NewScanner(os.Stdin)
	wizard := &configWizard{scanner: scanner, config: config}

	wizard.configureOrganization()
	wizard.configureSecurity()
	wizard.configureIntegrations()
	wizard.configureAnalytics()
	wizard.configureAudit()

	utils.Success("✅ Configuration wizard completed")
}

// configWizard handles the interactive configuration process
type configWizard struct {
	scanner *bufio.Scanner
	config  *EnterpriseConfiguration
}

// configureOrganization handles organization configuration
func (w *configWizard) configureOrganization() {
	utils.Info("📋 Organization Configuration")

	w.promptStringConfig("Organization Name [MAGE-X Organization]: ", func(value string) {
		w.config.Organization.Name = value
	})

	w.promptStringConfig("Organization Domain [example.com]: ", func(value string) {
		w.config.Organization.Domain = value
	})

	w.promptStringConfig("Region [us-east-1]: ", func(value string) {
		w.config.Organization.Region = value
	})
}

// configureSecurity handles security configuration
func (w *configWizard) configureSecurity() {
	utils.Info("🔒 Security Configuration")

	w.promptStringConfig("Security Level (minimal/standard/high/critical) [standard]: ", func(value string) {
		w.config.Security.Level = value
	})

	w.promptBoolConfig("Enable MFA (y/n) [n]: ", func(enabled bool) {
		w.config.Security.Authentication.MFA.Enabled = enabled
	})
}

// configureIntegrations handles integrations configuration
func (w *configWizard) configureIntegrations() {
	utils.Info("🔌 Integrations Configuration")

	w.promptBoolConfig("Enable Slack integration (y/n) [n]: ", func(enabled bool) {
		if enabled {
			w.config.Integrations.Providers["slack"] = w.createIntegrationProvider("communication")
		}
	})

	w.promptBoolConfig("Enable GitHub integration (y/n) [n]: ", func(enabled bool) {
		if enabled {
			w.config.Integrations.Providers["github"] = w.createIntegrationProvider("source_control")
		}
	})
}

// configureAnalytics handles analytics configuration
func (w *configWizard) configureAnalytics() {
	utils.Info("📊 Analytics Configuration")

	fmt.Print("Enable Analytics (y/n) [y]: ")
	if w.scanner.Scan() {
		if response := strings.TrimSpace(w.scanner.Text()); response == "n" || response == "no" {
			w.config.Analytics.Enabled = false
		}
	}
}

// configureAudit handles audit configuration
func (w *configWizard) configureAudit() {
	utils.Info("📝 Audit Configuration")

	w.promptStringConfig("Audit Level (debug/info/warn/error) [info]: ", func(value string) {
		w.config.Audit.Level = value
	})
}

// promptStringConfig prompts for a string configuration value
func (w *configWizard) promptStringConfig(prompt string, setter func(string)) {
	fmt.Print(prompt)
	if w.scanner.Scan() {
		if value := strings.TrimSpace(w.scanner.Text()); value != "" {
			setter(value)
		}
	}
}

// promptBoolConfig prompts for a boolean configuration value
func (w *configWizard) promptBoolConfig(prompt string, setter func(bool)) {
	fmt.Print(prompt)
	if w.scanner.Scan() {
		response := strings.TrimSpace(w.scanner.Text())
		if response == "y" || response == responseYes {
			setter(true)
		}
	}
}

// createIntegrationProvider creates a new integration provider
func (w *configWizard) createIntegrationProvider(providerType string) IntegrationProvider {
	return IntegrationProvider{
		Type:        providerType,
		Enabled:     true,
		Settings:    make(map[string]string),
		Credentials: make(map[string]string),
		Endpoints:   make(map[string]string),
	}
}

func saveEnterpriseConfiguration(config *EnterpriseConfiguration) error {
	// Update metadata
	config.Metadata.UpdatedAt = time.Now()

	// Save main configuration
	configPath := filepath.Join(".mage", "enterprise", "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o750); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	fileOps := fileops.New()
	return fileOps.File.WriteFile(configPath, data, 0o644)
}

func loadEnterpriseConfiguration() (*EnterpriseConfiguration, error) {
	configPath := filepath.Join(".mage", "enterprise", "config.yaml")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, ErrEnterpriseConfigNotFoundInit
	}

	fileOps := fileops.New()
	data, err := fileOps.File.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config EnterpriseConfiguration
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

func initializeSubConfigurations(config *EnterpriseConfiguration) error {
	// Initialize repository configuration
	if err := initializeRepositoryConfiguration(); err != nil {
		return err
	}

	// Initialize workflow templates
	if err := initializeWorkflowTemplates(); err != nil {
		return err
	}

	// Initialize integration configurations
	return initializeIntegrationConfigurations(config)
}

func initializeRepositoryConfiguration() error {
	repoConfig := RepositoryConfig{
		Version:      "1.0.0",
		Repositories: []Repository{},
		Groups:       []Group{},
		Settings: Settings{
			MaxConcurrency: 5,
			Timeout:        "30m",
			DefaultBranch:  "main",
		},
	}

	configPath := filepath.Join(".mage", "repositories.json")
	data, err := json.MarshalIndent(repoConfig, "", "  ")
	if err != nil {
		return err
	}

	fileOps := fileops.New()
	return fileOps.File.WriteFile(configPath, data, 0o644)
}

func initializeWorkflowTemplates() error {
	workflowsDir := filepath.Join(".mage", "workflows")
	if err := os.MkdirAll(workflowsDir, 0o750); err != nil {
		return err
	}

	// Create basic workflow templates
	templates := []struct {
		name     string
		workflow WorkflowDefinition
	}{
		{
			name:     "ci",
			workflow: createCIWorkflowTemplate(),
		},
		{
			name:     "deploy",
			workflow: createDeploymentWorkflowTemplate(),
		},
		{
			name:     "security",
			workflow: createSecurityWorkflowTemplate(),
		},
	}

	for i := range templates {
		template := &templates[i]
		templatePath := filepath.Join(workflowsDir, fmt.Sprintf("%s.json", template.name))
		data, err := json.MarshalIndent(template.workflow, "", "  ")
		if err != nil {
			return err
		}

		fileOps := fileops.New()
		if err := fileOps.File.WriteFile(templatePath, data, 0o644); err != nil {
			return err
		}
	}

	return nil
}

func createCIWorkflowTemplate() WorkflowDefinition {
	return WorkflowDefinition{
		Name:        "ci",
		Description: "Continuous Integration workflow",
		Version:     "1.0.0",
		Steps: []WorkflowStep{
			{
				Name:    "checkout",
				Type:    "shell",
				Command: "git",
				Args:    []string{"pull", "origin", "main"},
			},
			{
				Name:    "dependencies",
				Type:    "shell",
				Command: "go",
				Args:    []string{"mod", "download"},
			},
			{
				Name:    "build",
				Type:    "shell",
				Command: "go",
				Args:    []string{"build", "./..."},
			},
			{
				Name:    "test",
				Type:    "shell",
				Command: "go",
				Args:    []string{"test", "-v", "./..."},
			},
			{
				Name:    "lint",
				Type:    "shell",
				Command: "golangci-lint",
				Args:    []string{"run"},
			},
		},
		Variables: make(map[string]interface{}),
		Settings: WorkflowSettings{
			Timeout:          "20m",
			MaxRetries:       3,
			FailureStrategy:  "stop",
			NotificationMode: "on_failure",
			Environment:      make(map[string]string),
		},
		Triggers:    []WorkflowTrigger{},
		LastUpdated: time.Now(),
	}
}

func createDeploymentWorkflowTemplate() WorkflowDefinition {
	return WorkflowDefinition{
		Name:        "deploy",
		Description: "Deployment workflow",
		Version:     "1.0.0",
		Steps: []WorkflowStep{
			{
				Name:    "build",
				Type:    "shell",
				Command: "go",
				Args:    []string{"build", "-o", "app", "./cmd/app"},
			},
			{
				Name:    "test",
				Type:    "shell",
				Command: "go",
				Args:    []string{"test", "./..."},
			},
			{
				Name:    "security-scan",
				Type:    "shell",
				Command: "gosec",
				Args:    []string{"./..."},
			},
			{
				Name:    "deploy",
				Type:    "shell",
				Command: "kubectl",
				Args:    []string{"apply", "-f", "deployment.yaml"},
			},
			{
				Name:    "health-check",
				Type:    "http",
				Command: "curl",
				Args:    []string{"-f", "http://localhost:8080/health"},
			},
		},
		Variables: make(map[string]interface{}),
		Settings: WorkflowSettings{
			Timeout:          "30m",
			MaxRetries:       2,
			FailureStrategy:  "rollback",
			NotificationMode: "always",
			Environment:      make(map[string]string),
		},
		Triggers:    []WorkflowTrigger{},
		LastUpdated: time.Now(),
	}
}

func createSecurityWorkflowTemplate() WorkflowDefinition {
	return WorkflowDefinition{
		Name:        "security",
		Description: "Security scanning workflow",
		Version:     "1.0.0",
		Steps: []WorkflowStep{
			{
				Name:    "vulnerability-scan",
				Type:    "shell",
				Command: "govulncheck",
				Args:    []string{"./..."},
			},
			{
				Name:    "secret-scan",
				Type:    "shell",
				Command: "gitleaks",
				Args:    []string{"detect", "--source", ".", "--verbose"},
			},
			{
				Name:    "security-lint",
				Type:    "shell",
				Command: "gosec",
				Args:    []string{"-severity", "medium", "./..."},
			},
			{
				Name:    "dependency-check",
				Type:    "shell",
				Command: "nancy",
				Args:    []string{"sleuth"},
			},
		},
		Variables: make(map[string]interface{}),
		Settings: WorkflowSettings{
			Timeout:          "15m",
			MaxRetries:       2,
			FailureStrategy:  "stop",
			NotificationMode: "on_failure",
			Environment:      make(map[string]string),
		},
		Triggers:    []WorkflowTrigger{},
		LastUpdated: time.Now(),
	}
}

func initializeIntegrationConfigurations(config *EnterpriseConfiguration) error {
	integrationsDir := filepath.Join(".mage", "integrations")
	if err := os.MkdirAll(integrationsDir, 0o750); err != nil {
		return err
	}

	// Create example integration configurations
	for name, provider := range config.Integrations.Providers {
		// Convert Settings from map[string]string to map[string]interface{}
		settings := make(map[string]interface{})
		for k, v := range provider.Settings {
			settings[k] = v
		}

		integrationConfig := IntegrationConfig{
			Name:        name,
			Type:        provider.Type,
			Enabled:     provider.Enabled,
			Settings:    settings,
			Credentials: provider.Credentials,
			Endpoints:   provider.Endpoints,
			Webhooks:    []WebhookConfig{},
			Created:     time.Now(),
			Updated:     time.Now(),
		}

		configPath := filepath.Join(integrationsDir, fmt.Sprintf("%s.json", name))
		data, err := json.MarshalIndent(integrationConfig, "", "  ")
		if err != nil {
			return err
		}

		fileOps := fileops.New()
		if err := fileOps.File.WriteFile(configPath, data, 0o644); err != nil {
			return err
		}
	}

	return nil
}

// NewConfigurationValidator creates a new configuration validator with default rules
func NewConfigurationValidator() *ConfigurationValidator {
	return &ConfigurationValidator{
		rules: []ECValidationRule{
			&ECOrganizationValidationRule{},
			&ECSecurityValidationRule{},
			&ECIntegrationsValidationRule{},
			&ECWorkflowsValidationRule{},
		},
	}
}

// Validate validates an enterprise configuration using all registered rules
func (v *ConfigurationValidator) Validate(config *EnterpriseConfiguration) *ValidationResults {
	results := &ValidationResults{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
		Info:     []ValidationInfo{},
	}

	// Run all validation rules
	for _, rule := range v.rules {
		errors := rule.Validate(config)
		results.Errors = append(results.Errors, errors...)
	}

	// Set overall validity
	results.Valid = len(results.Errors) == 0

	return results
}

// HasErrors returns true if there are any validation errors
func (vr *ValidationResults) HasErrors() bool {
	return len(vr.Errors) > 0
}

// Validation Rules

// ECOrganizationValidationRule validates organization configuration settings
type ECOrganizationValidationRule struct{}

// Validate validates organization configuration settings
func (r *ECOrganizationValidationRule) Validate(config *EnterpriseConfiguration) []ValidationError {
	var errors []ValidationError

	if config.Organization.Name == "" {
		errors = append(errors, ValidationError{
			Field:   "organization.name",
			Message: "Organization name is required",
			Code:    "ORG_NAME_REQUIRED",
		})
	}

	if config.Organization.Domain == "" {
		errors = append(errors, ValidationError{
			Field:   "organization.domain",
			Message: "Organization domain is required",
			Code:    "ORG_DOMAIN_REQUIRED",
		})
	}

	return errors
}

// ECSecurityValidationRule validates security configuration settings
type ECSecurityValidationRule struct{}

// Validate validates security configuration settings
func (r *ECSecurityValidationRule) Validate(config *EnterpriseConfiguration) []ValidationError {
	var errors []ValidationError

	validLevels := []string{"minimal", "standard", "high", "critical"}
	if !contains(validLevels, config.Security.Level) {
		errors = append(errors, ValidationError{
			Field:   "security.level",
			Message: "Invalid security level",
			Code:    "SECURITY_LEVEL_INVALID",
		})
	}

	return errors
}

// ECIntegrationsValidationRule validates integrations configuration settings
type ECIntegrationsValidationRule struct{}

// Validate validates integrations configuration settings
func (r *ECIntegrationsValidationRule) Validate(config *EnterpriseConfiguration) []ValidationError {
	var errors []ValidationError

	// Validate each integration provider
	for name, provider := range config.Integrations.Providers {
		if provider.Type == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("integrations.providers.%s.type", name),
				Message: "Integration provider type is required",
				Code:    "INTEGRATION_TYPE_REQUIRED",
			})
		}
	}

	return errors
}

// ECWorkflowsValidationRule validates workflows configuration settings
type ECWorkflowsValidationRule struct{}

// Validate validates workflows configuration settings
func (r *ECWorkflowsValidationRule) Validate(config *EnterpriseConfiguration) []ValidationError {
	var errors []ValidationError

	if config.Workflows.Enabled && config.Workflows.Directory == "" {
		errors = append(errors, ValidationError{
			Field:   "workflows.directory",
			Message: "Workflows directory is required when workflows are enabled",
			Code:    "WORKFLOWS_DIRECTORY_REQUIRED",
		})
	}

	return errors
}

func displayValidationResults(results *ValidationResults) {
	if results.Valid {
		utils.Success("✅ Configuration validation passed")
	} else {
		utils.Error("❌ Configuration validation failed")
	}

	if len(results.Errors) > 0 {
		utils.Error("🚨 Validation Errors:")
		for _, err := range results.Errors {
			utils.Error("  - %s: %s (%s)", err.Field, err.Message, err.Code)
		}
	}

	if len(results.Warnings) > 0 {
		utils.Warn("⚠️  Validation Warnings:")
		for _, warn := range results.Warnings {
			utils.Warn("  - %s: %s (%s)", warn.Field, warn.Message, warn.Code)
		}
	}

	if len(results.Info) > 0 {
		utils.Info("ℹ️  Validation Info:")
		for _, info := range results.Info {
			utils.Info("  - %s: %s (%s)", info.Field, info.Message, info.Code)
		}
	}
}

func runInteractiveConfigUpdate(config *EnterpriseConfiguration) error {
	utils.Info("🔄 Interactive Configuration Update")

	scanner := bufio.NewScanner(os.Stdin)

	sections := []string{
		"organization",
		"security",
		"integrations",
		"workflows",
		"analytics",
		"audit",
		"cli",
		"repositories",
		"compliance",
		"monitoring",
		"deployment",
		"backup",
		"notifications",
	}

	utils.Info("Available sections:")
	for i, section := range sections {
		utils.Info("  %d. %s", i+1, section)
	}

	fmt.Print("Select section to update (1-13): ")
	if scanner.Scan() {
		if choice := strings.TrimSpace(scanner.Text()); choice != "" {
			if idx, err := strconv.Atoi(choice); err == nil && idx >= 1 && idx <= len(sections) {
				return updateConfigurationSection(config, sections[idx-1])
			}
		}
	}

	return ErrInvalidSelection
}

func updateConfigurationSection(config *EnterpriseConfiguration, section string) error {
	switch section {
	case "organization":
		return updateOrganizationSection(config)
	case "security":
		return updateSecuritySection(config)
	case "integrations":
		return updateIntegrationsSection(config)
	case "workflows":
		return updateWorkflowsSection(config)
	case "analytics":
		return updateAnalyticsSection(config)
	case "audit":
		return updateAuditSection(config)
	case "cli":
		return updateCLISection(config)
	case "repositories":
		return updateRepositoriesSection(config)
	case "compliance":
		return updateComplianceSection(config)
	case "monitoring":
		return updateMonitoringSection(config)
	case "deployment":
		return updateDeploymentSection(config)
	case "backup":
		return updateBackupSection(config)
	case "notifications":
		return updateNotificationsSection(config)
	default:
		return fmt.Errorf("%w: %s", ErrUnknownSection, section)
	}
}

func updateOrganizationSection(config *EnterpriseConfiguration) error {
	utils.Info("🏢 Updating Organization Configuration")

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("Organization Name [%s]: ", config.Organization.Name)
	if scanner.Scan() {
		if name := strings.TrimSpace(scanner.Text()); name != "" {
			config.Organization.Name = name
		}
	}

	fmt.Printf("Domain [%s]: ", config.Organization.Domain)
	if scanner.Scan() {
		if domain := strings.TrimSpace(scanner.Text()); domain != "" {
			config.Organization.Domain = domain
		}
	}

	fmt.Printf("Region [%s]: ", config.Organization.Region)
	if scanner.Scan() {
		if region := strings.TrimSpace(scanner.Text()); region != "" {
			config.Organization.Region = region
		}
	}

	// Save updated configuration
	return saveEnterpriseConfiguration(config)
}

func updateSecuritySection(config *EnterpriseConfiguration) error {
	utils.Info("🔒 Updating Security Configuration")

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("Security Level [%s]: ", config.Security.Level)
	if scanner.Scan() {
		if level := strings.TrimSpace(scanner.Text()); level != "" {
			config.Security.Level = level
		}
	}

	fmt.Printf("Enable MFA [%t]: ", config.Security.Authentication.MFA.Enabled)
	if scanner.Scan() {
		if mfa := strings.TrimSpace(scanner.Text()); mfa != "" {
			config.Security.Authentication.MFA.Enabled = mfa == approvalTrue || mfa == "y" || mfa == responseYes
		}
	}

	// Save updated configuration
	return saveEnterpriseConfiguration(config)
}

func updateIntegrationsSection(config *EnterpriseConfiguration) error {
	utils.Info("🔌 Updating Integrations Configuration")
	// Implementation for integrations section update
	return saveEnterpriseConfiguration(config)
}

func updateWorkflowsSection(config *EnterpriseConfiguration) error {
	utils.Info("🔄 Updating Workflows Configuration")
	// Implementation for workflows section update
	return saveEnterpriseConfiguration(config)
}

func updateAnalyticsSection(config *EnterpriseConfiguration) error {
	utils.Info("📊 Updating Analytics Configuration")
	// Implementation for analytics section update
	return saveEnterpriseConfiguration(config)
}

func updateAuditSection(config *EnterpriseConfiguration) error {
	utils.Info("📝 Updating Audit Configuration")
	// Implementation for audit section update
	return saveEnterpriseConfiguration(config)
}

func updateCLISection(config *EnterpriseConfiguration) error {
	utils.Info("💻 Updating CLI Configuration")
	// Implementation for CLI section update
	return saveEnterpriseConfiguration(config)
}

func updateRepositoriesSection(config *EnterpriseConfiguration) error {
	utils.Info("📁 Updating Repositories Configuration")
	// Implementation for repositories section update
	return saveEnterpriseConfiguration(config)
}

func updateComplianceSection(config *EnterpriseConfiguration) error {
	utils.Info("⚖️ Updating Compliance Configuration")
	// Implementation for compliance section update
	return saveEnterpriseConfiguration(config)
}

func updateMonitoringSection(config *EnterpriseConfiguration) error {
	utils.Info("📡 Updating Monitoring Configuration")
	// Implementation for monitoring section update
	return saveEnterpriseConfiguration(config)
}

func updateDeploymentSection(config *EnterpriseConfiguration) error {
	utils.Info("🚀 Updating Deployment Configuration")
	// Implementation for deployment section update
	return saveEnterpriseConfiguration(config)
}

func updateBackupSection(config *EnterpriseConfiguration) error {
	utils.Info("💾 Updating Backup Configuration")
	// Implementation for backup section update
	return saveEnterpriseConfiguration(config)
}

func updateNotificationsSection(config *EnterpriseConfiguration) error {
	utils.Info("📢 Updating Notifications Configuration")
	// Implementation for notifications section update
	return saveEnterpriseConfiguration(config)
}

func exportConfiguration(config *EnterpriseConfiguration, format, outputFile string) error {
	var data []byte
	var err error

	switch format {
	case "yaml":
		data, err = yaml.Marshal(config)
	case "json":
		data, err = json.MarshalIndent(config, "", "  ")
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedConfigFormat, format)
	}

	if err != nil {
		return err
	}

	return os.WriteFile(outputFile, data, 0o600)
}

func importConfiguration(inputFile string) (*EnterpriseConfiguration, error) {
	fileOps := fileops.New()
	data, err := fileOps.File.ReadFile(inputFile)
	if err != nil {
		return nil, err
	}

	var config EnterpriseConfiguration

	// Try YAML first, then JSON
	if err := yaml.Unmarshal(data, &config); err != nil {
		if err := json.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse configuration file: %w", err)
		}
	}

	return &config, nil
}

func generateEnterpriseConfigurationSchema() map[string]interface{} {
	schema := map[string]interface{}{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"title":   "MAGE-X Enterprise Configuration Schema",
		"type":    "object",
		"properties": map[string]interface{}{
			"metadata": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"version":     map[string]interface{}{"type": "string"},
					"created_at":  map[string]interface{}{"type": "string", "format": "date-time"},
					"updated_at":  map[string]interface{}{"type": "string", "format": "date-time"},
					"created_by":  map[string]interface{}{"type": "string"},
					"updated_by":  map[string]interface{}{"type": "string"},
					"description": map[string]interface{}{"type": "string"},
					"tags":        map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"labels":      map[string]interface{}{"type": "object", "additionalProperties": map[string]interface{}{"type": "string"}},
				},
				"required": []string{"version", "created_at", "updated_at"},
			},
			"organization": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name":     map[string]interface{}{"type": "string"},
					"domain":   map[string]interface{}{"type": "string"},
					"region":   map[string]interface{}{"type": "string"},
					"timezone": map[string]interface{}{"type": "string"},
					"language": map[string]interface{}{"type": "string"},
					"currency": map[string]interface{}{"type": "string"},
				},
				"required": []string{"name", "domain"},
			},
			"security": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"level": map[string]interface{}{
						"type": "string",
						"enum": []string{"minimal", "standard", "high", "critical"},
					},
					"encryption": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"algorithm": map[string]interface{}{"type": "string"},
							"key_size":  map[string]interface{}{"type": "integer"},
						},
					},
					"authentication": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"methods":      map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
							"token_expiry": map[string]interface{}{"type": "string"},
							"mfa": map[string]interface{}{
								"type": "object",
								"properties": map[string]interface{}{
									"enabled":  map[string]interface{}{"type": "boolean"},
									"methods":  map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
									"required": map[string]interface{}{"type": "boolean"},
								},
							},
						},
					},
				},
				"required": []string{"level"},
			},
		},
		"required": []string{"metadata", "organization", "security"},
	}

	return schema
}

// Utility functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
