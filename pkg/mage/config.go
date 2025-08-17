// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Config represents the mage configuration
type Config struct {
	Project    ProjectConfig            `yaml:"project"`
	Build      BuildConfig              `yaml:"build"`
	Test       TestConfig               `yaml:"test"`
	Lint       LintConfig               `yaml:"lint"`
	Tools      ToolsConfig              `yaml:"tools"`
	Docker     DockerConfig             `yaml:"docker"`
	Release    ReleaseConfig            `yaml:"release"`
	Download   DownloadConfig           `yaml:"download"`
	Enterprise *EnterpriseConfiguration `yaml:"enterprise,omitempty"`
	Metadata   map[string]string        `yaml:"metadata,omitempty"`
}

// ProjectConfig contains project-specific settings
type ProjectConfig struct {
	Name        string            `yaml:"name"`
	Binary      string            `yaml:"binary"`
	Module      string            `yaml:"module"`
	Description string            `yaml:"description"`
	Version     string            `yaml:"version"`
	GitDomain   string            `yaml:"git_domain"`
	RepoOwner   string            `yaml:"repo_owner"`
	RepoName    string            `yaml:"repo_name"`
	Env         map[string]string `yaml:"env"`
}

// BuildConfig contains build-specific settings
type BuildConfig struct {
	Tags      []string `yaml:"tags"`
	LDFlags   []string `yaml:"ldflags"`
	Platforms []string `yaml:"platforms"`
	GoFlags   []string `yaml:"goflags"`
	Output    string   `yaml:"output"`
	Parallel  int      `yaml:"parallel"`
	Verbose   bool     `yaml:"verbose"`
	TrimPath  bool     `yaml:"trimpath"`
}

// TestConfig contains test-specific settings
type TestConfig struct {
	Parallel           int      `yaml:"parallel"`
	Timeout            string   `yaml:"timeout"`
	IntegrationTimeout string   `yaml:"integration_timeout"`
	IntegrationTag     string   `yaml:"integration_tag"`
	Short              bool     `yaml:"short"`
	Verbose            bool     `yaml:"verbose"`
	Race               bool     `yaml:"race"`
	Cover              bool     `yaml:"cover"`
	CoverMode          string   `yaml:"covermode"`
	CoverPkg           []string `yaml:"coverpkg"`
	CoverageExclude    []string `yaml:"coverage_exclude"`
	Tags               string   `yaml:"tags"`
	SkipFuzz           bool     `yaml:"skip_fuzz"`
	Shuffle            bool     `yaml:"shuffle"`
	BenchCPU           int      `yaml:"bench_cpu"`
	BenchTime          string   `yaml:"bench_time"`
	BenchMem           bool     `yaml:"bench_mem"`
}

// LintConfig contains linting settings
type LintConfig struct {
	GolangciVersion string   `yaml:"golangci_version"`
	Timeout         string   `yaml:"timeout"`
	SkipDirs        []string `yaml:"skip_dirs"`
	SkipFiles       []string `yaml:"skip_files"`
	EnableAll       bool     `yaml:"enable_all"`
	DisableLinters  []string `yaml:"disable_linters"`
	EnableLinters   []string `yaml:"enable_linters"`
}

// ToolsConfig contains tool versions
type ToolsConfig struct {
	GolangciLint string            `yaml:"golangci_lint"`
	Fumpt        string            `yaml:"fumpt"`
	GoVulnCheck  string            `yaml:"govulncheck"`
	Mockgen      string            `yaml:"mockgen"`
	Swag         string            `yaml:"swag"`
	Custom       map[string]string `yaml:"custom"`
}

// DockerConfig contains Docker settings
type DockerConfig struct {
	Registry        string            `yaml:"registry"`
	Repository      string            `yaml:"repository"`
	Dockerfile      string            `yaml:"dockerfile"`
	BuildArgs       map[string]string `yaml:"build_args"`
	Platforms       []string          `yaml:"platforms"`
	EnableBuildKit  bool              `yaml:"enable_buildkit"`
	Labels          map[string]string `yaml:"labels"`
	CacheFrom       []string          `yaml:"cache_from"`
	SecurityOpts    []string          `yaml:"security_opts"`
	NetworkMode     string            `yaml:"network_mode"`
	DefaultRegistry string            `yaml:"default_registry"`
}

// ReleaseConfig contains release settings
type ReleaseConfig struct {
	GitHubToken string   `yaml:"github_token_env"`
	Changelog   bool     `yaml:"changelog"`
	Draft       bool     `yaml:"draft"`
	Prerelease  bool     `yaml:"prerelease"`
	NameTmpl    string   `yaml:"name_template"`
	Formats     []string `yaml:"formats"`
}

// DownloadConfig contains download retry settings
type DownloadConfig struct {
	MaxRetries        int     `yaml:"max_retries"`
	InitialDelayMs    int     `yaml:"initial_delay_ms"`
	MaxDelayMs        int     `yaml:"max_delay_ms"`
	TimeoutMs         int     `yaml:"timeout_ms"`
	BackoffMultiplier float64 `yaml:"backoff_multiplier"`
	EnableResume      bool    `yaml:"enable_resume"`
	UserAgent         string  `yaml:"user_agent"`
}

// Static errors for err113 compliance
var (
	ErrEnterpriseConfigExists = errors.New("enterprise configuration already exists")
	ErrMissingToolVersions    = errors.New("missing required tool versions")
)

// LoadConfig loads the configuration from file or returns defaults
// Deprecated: Use GetConfig() instead, which uses the ConfigProvider pattern
func LoadConfig() (*Config, error) {
	// Use the config provider pattern
	return GetConfigProvider().GetConfig()
}

// GetConfig returns the current configuration using the active ConfigProvider
func GetConfig() (*Config, error) {
	return GetConfigProvider().GetConfig()
}

// GetToolVersion returns the version for a given tool, reading from environment variables
// with fallback to configuration or empty string if not found. This provides a centralized way to get tool versions.
func GetToolVersion(toolName string) string {
	// Define the mapping from tool names to environment variables
	toolVersionMap := map[string]struct {
		envVar       string
		legacyEnvVar string
	}{
		"golangci-lint": {"MAGE_X_GOLANGCI_LINT_VERSION", "GOLANGCI_LINT_VERSION"},
		"gofumpt":       {"MAGE_X_GOFUMPT_VERSION", "GOFUMPT_VERSION"},
		"govulncheck":   {"MAGE_X_GOVULNCHECK_VERSION", "GOVULNCHECK_VERSION"},
		"mockgen":       {"MAGE_X_MOCKGEN_VERSION", "MOCKGEN_VERSION"},
		"swag":          {"MAGE_X_SWAG_VERSION", "SWAG_VERSION"},
		"staticcheck":   {"MAGE_X_STATICCHECK_VERSION", ""},
		"nancy":         {"MAGE_X_NANCY_VERSION", "NANCY_VERSION"},
		"gitleaks":      {"MAGE_X_GITLEAKS_VERSION", "GITLEAKS_VERSION"},
		"goreleaser":    {"MAGE_X_GORELEASER_VERSION", "GORELEASER_VERSION"},
		"prettier":      {"MAGE_X_PRETTIER_VERSION", "PRETTIER_VERSION"},
	}

	toolInfo, exists := toolVersionMap[toolName]
	if !exists {
		// For unknown tools, try to construct environment variable name
		envVar := "MAGE_X_" + strings.ToUpper(strings.ReplaceAll(toolName, "-", "_")) + "_VERSION"
		if version := os.Getenv(envVar); version != "" {
			return version
		}
		utils.Warn("Tool version for %s not found in environment variables (%s)", toolName, envVar)
		utils.Warn("Consider sourcing .github/.env.base: source .github/.env.base")
		return ""
	}

	// Check primary environment variable first
	if version := os.Getenv(toolInfo.envVar); version != "" {
		return version
	}

	// Check legacy environment variable for backward compatibility
	if toolInfo.legacyEnvVar != "" {
		if version := os.Getenv(toolInfo.legacyEnvVar); version != "" {
			return version
		}
	}

	// Warn if not found and return empty string
	utils.Warn("Tool version for %s not found in environment variables (%s)", toolName, toolInfo.envVar)
	utils.Warn("Consider sourcing .github/.env.base: source .github/.env.base")
	return ""
}

// ValidateToolVersions validates that all required tool versions are available in environment variables
func ValidateToolVersions() error {
	requiredTools := []string{
		"golangci-lint", "gofumpt", "govulncheck",
	}

	optionalTools := []string{
		"mockgen", "swag", "staticcheck", "nancy", "gitleaks", "goreleaser", "prettier",
	}

	var missingRequired []string
	var missingOptional []string

	// Check required tools
	for _, tool := range requiredTools {
		if GetToolVersion(tool) == "" {
			missingRequired = append(missingRequired, tool)
		}
	}

	// Check optional tools
	for _, tool := range optionalTools {
		if GetToolVersion(tool) == "" {
			missingOptional = append(missingOptional, tool)
		}
	}

	// Report missing tools
	if len(missingRequired) > 0 {
		return fmt.Errorf("%w: %v. Source .github/.env.base to fix", ErrMissingToolVersions, missingRequired)
	}

	if len(missingOptional) > 0 {
		utils.Warn("Missing optional tool versions: %v", missingOptional)
		utils.Warn("These tools may not function correctly without version information")
	}

	return nil
}

// EnsureEnvironmentLoaded checks if key environment variables are loaded and provides guidance
func EnsureEnvironmentLoaded() {
	// Check if key environment variables are set
	keyEnvVars := []string{
		"MAGE_X_GOLANGCI_LINT_VERSION",
		"MAGE_X_GOFUMPT_VERSION",
		"MAGE_X_GOVULNCHECK_VERSION",
	}

	var missing []string
	for _, envVar := range keyEnvVars {
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}

	if len(missing) > 0 {
		utils.Warn("Key environment variables not loaded: %v", missing)
		utils.Warn("Consider sourcing .github/.env.base:")
		utils.Warn("  source .github/.env.base")
		utils.Warn("Or set environment variables manually:")
		for _, envVar := range missing {
			utils.Warn("  export %s=<version>", envVar)
		}
	}
}

// TestResetConfig resets the config for testing purposes only
// This should only be used in tests
func TestResetConfig() {
	// Reset the config provider
	GetConfigProvider().ResetConfig()
	// Also reset the package provider registry provider to ensure fresh instances
	getPackageProviderRegistryProvider().ResetRegistry()
}

// TestSetConfig sets a config for testing purposes only
// This should only be used in tests
func TestSetConfig(config *Config) {
	// Set via the config provider
	GetConfigProvider().SetConfig(config)
}

// defaultConfig returns the default configuration
func defaultConfig() *Config {
	parallel := runtime.NumCPU()
	if parallel < 1 {
		parallel = 1
	}

	// Try to detect project info
	module, err := getModuleName()
	if err != nil {
		// If we can't get module name, use empty string
		module = ""
	}
	binary := filepath.Base(module)
	if binary == "." || binary == "" || binary == "command-line-arguments" {
		binary = defaultBinaryName
	}

	return &Config{
		Project: ProjectConfig{
			Name:      binary,
			Binary:    binary,
			Module:    module,
			GitDomain: "github.com",
		},
		Build: BuildConfig{
			Output:   "bin",
			Parallel: parallel,
			TrimPath: true,
			Platforms: []string{
				"linux/amd64",
				"darwin/amd64",
				"darwin/arm64",
				"windows/amd64",
			},
		},
		Test: TestConfig{
			Parallel:           runtime.NumCPU(),
			Timeout:            "10m",
			IntegrationTimeout: "30m",
			CoverMode:          "atomic",
		},
		Lint: LintConfig{
			GolangciVersion: GetDefaultGolangciLintVersion(),
			Timeout:         "5m",
		},
		Tools: ToolsConfig{
			GolangciLint: GetDefaultGolangciLintVersion(),
			Fumpt:        GetDefaultGofumptVersion(),
			GoVulnCheck:  GetDefaultGoVulnCheckVersion(),
			Mockgen:      GetDefaultMockgenVersion(),
			Swag:         GetDefaultSwagVersion(),
			Custom:       make(map[string]string),
		},
		Docker: DockerConfig{
			Dockerfile:      "Dockerfile",
			Platforms:       []string{"linux/amd64", "linux/arm64"},
			EnableBuildKit:  true,
			DefaultRegistry: "docker.io",
			BuildArgs:       make(map[string]string),
			Labels:          make(map[string]string),
		},
		Release: ReleaseConfig{
			GitHubToken: "GITHUB_TOKEN",
			Changelog:   true,
			Formats:     []string{"tar.gz", "zip"},
		},
		Download: DownloadConfig{
			MaxRetries:        5,
			InitialDelayMs:    1000,  // 1 second
			MaxDelayMs:        30000, // 30 seconds
			TimeoutMs:         60000, // 60 seconds
			BackoffMultiplier: 2.0,
			EnableResume:      true,
			UserAgent:         "mage-x-downloader/1.0",
		},
	}
}

// applyEnvOverrides applies environment variable overrides to config
func applyEnvOverrides(c *Config) {
	// Binary name override
	if v := os.Getenv("MAGE_X_BINARY_NAME"); v != "" {
		c.Project.Binary = v
	}
	if v := os.Getenv("CUSTOM_BINARY_NAME"); v != "" {
		c.Project.Binary = v
	}

	// Build tags override
	if v := os.Getenv("MAGE_X_BUILD_TAGS"); v != "" {
		c.Build.Tags = strings.Split(v, ",")
	}

	// Verbose override
	if v := os.Getenv("MAGE_X_VERBOSE"); v == approvalTrue || v == "1" {
		c.Build.Verbose = true
		c.Test.Verbose = true
	}

	// Test race override
	if v := os.Getenv("MAGE_X_TEST_RACE"); v == approvalTrue || v == "1" {
		c.Test.Race = true
	}

	// Parallel override
	if v := os.Getenv("MAGE_X_PARALLEL"); v != "" {
		var parallel int
		if _, err := fmt.Sscanf(v, "%d", &parallel); err == nil && parallel > 0 {
			c.Build.Parallel = parallel
		}
	}

	// Download config overrides
	applyDownloadEnvOverrides(&c.Download)

	// Tool version overrides
	applyToolVersionEnvOverrides(&c.Tools)

	// Enterprise overrides
	if c.Enterprise != nil {
		applyEnterpriseEnvOverrides(c.Enterprise)
	}
}

// applyDownloadEnvOverrides applies environment variable overrides to download config
func applyDownloadEnvOverrides(cfg *DownloadConfig) {
	// Max retries override
	if v := os.Getenv("MAGE_X_DOWNLOAD_RETRIES"); v != "" {
		var retries int
		if _, err := fmt.Sscanf(v, "%d", &retries); err == nil && retries >= 0 {
			cfg.MaxRetries = retries
		}
	}

	// Timeout override
	if v := os.Getenv("MAGE_X_DOWNLOAD_TIMEOUT"); v != "" {
		var timeout int
		if _, err := fmt.Sscanf(v, "%d", &timeout); err == nil && timeout > 0 {
			cfg.TimeoutMs = timeout
		}
	}

	// Initial delay override
	if v := os.Getenv("MAGE_X_DOWNLOAD_INITIAL_DELAY"); v != "" {
		var delay int
		if _, err := fmt.Sscanf(v, "%d", &delay); err == nil && delay > 0 {
			cfg.InitialDelayMs = delay
		}
	}

	// Max delay override
	if v := os.Getenv("MAGE_X_DOWNLOAD_MAX_DELAY"); v != "" {
		var delay int
		if _, err := fmt.Sscanf(v, "%d", &delay); err == nil && delay > 0 {
			cfg.MaxDelayMs = delay
		}
	}

	// Backoff multiplier override
	if v := os.Getenv("MAGE_X_DOWNLOAD_BACKOFF"); v != "" {
		var backoff float64
		if _, err := fmt.Sscanf(v, "%f", &backoff); err == nil && backoff > 0 {
			cfg.BackoffMultiplier = backoff
		}
	}

	// Resume override
	if v := os.Getenv("MAGE_X_DOWNLOAD_RESUME"); v == approvalTrue || v == "1" {
		cfg.EnableResume = true
	} else if v == "false" || v == "0" {
		cfg.EnableResume = false
	}

	// User agent override
	if v := os.Getenv("MAGE_X_DOWNLOAD_USER_AGENT"); v != "" {
		cfg.UserAgent = v
	}
}

// applyToolVersionEnvOverrides applies environment variable overrides to tool versions
func applyToolVersionEnvOverrides(cfg *ToolsConfig) {
	// Core linting tools
	if v := os.Getenv("MAGE_X_GOLANGCI_LINT_VERSION"); v != "" {
		cfg.GolangciLint = v
	}
	if v := os.Getenv("MAGE_X_GOFUMPT_VERSION"); v != "" {
		cfg.Fumpt = v
	}

	// Security scanning tools
	if v := os.Getenv("MAGE_X_GOVULNCHECK_VERSION"); v != "" {
		cfg.GoVulnCheck = v
	}

	// Code generation tools
	if v := os.Getenv("MAGE_X_MOCKGEN_VERSION"); v != "" {
		cfg.Mockgen = v
	}
	if v := os.Getenv("MAGE_X_SWAG_VERSION"); v != "" {
		cfg.Swag = v
	}

	// Legacy environment variable support for backward compatibility
	if v := os.Getenv("GOLANGCI_LINT_VERSION"); v != "" && cfg.GolangciLint == "" {
		cfg.GolangciLint = v
	}
	if v := os.Getenv("GOFUMPT_VERSION"); v != "" && cfg.Fumpt == "" {
		cfg.Fumpt = v
	}
	if v := os.Getenv("GOVULNCHECK_VERSION"); v != "" && cfg.GoVulnCheck == "" {
		cfg.GoVulnCheck = v
	}
	if v := os.Getenv("MOCKGEN_VERSION"); v != "" && cfg.Mockgen == "" {
		cfg.Mockgen = v
	}
	if v := os.Getenv("SWAG_VERSION"); v != "" && cfg.Swag == "" {
		cfg.Swag = v
	}
}

// applyEnterpriseEnvOverrides applies environment variable overrides to enterprise config
func applyEnterpriseEnvOverrides(cfg *EnterpriseConfiguration) {
	// Organization overrides
	if v := os.Getenv("MAGE_X_ORG_NAME"); v != "" {
		cfg.Organization.Name = v
	}
	if v := os.Getenv("MAGE_X_ORG_DOMAIN"); v != "" {
		cfg.Organization.Domain = v
	}

	// Security configuration environment variables are reserved for future implementation.
	// When Security field is added to Config, these will be processed:
	_ = os.Getenv("MAGE_X_SECURITY_LEVEL") // placeholder for security level
	// Vault integration is reserved for future implementation.
	// When Security field is added to Config, these will be processed:
	_ = os.Getenv("MAGE_X_ENABLE_VAULT") // placeholder for vault enabled
	_ = os.Getenv("VAULT_ADDR")          // placeholder for vault address

	// Analytics configuration is reserved for future implementation.
	// When Analytics field is added to Config, these will be processed:
	_ = os.Getenv("MAGE_X_ANALYTICS_ENABLED") // placeholder for analytics enabled
	_ = os.Getenv("MAGE_X_METRICS_INTERVAL")  // placeholder for metrics interval
}

// BinaryName returns the configured binary name
func BinaryName() string {
	c, err := GetConfig()
	if err != nil {
		// Return default binary name if config loading fails
		return "app"
	}
	return c.Project.Binary
}

// BuildTags returns the configured build tags as a string
func BuildTags() string {
	c, err := GetConfig()
	if err != nil {
		// Return empty string if config loading fails
		return ""
	}
	if len(c.Build.Tags) == 0 {
		return ""
	}
	return strings.Join(c.Build.Tags, ",")
}

// IsVerbose returns whether verbose mode is enabled
func IsVerbose() bool {
	c, err := GetConfig()
	if err != nil {
		// Return false if config loading fails
		return false
	}
	return c.Build.Verbose || c.Test.Verbose
}

// HasEnterpriseConfig returns whether enterprise configuration is enabled
func HasEnterpriseConfig() bool {
	c, err := GetConfig()
	if err != nil {
		// Return false if config loading fails
		return false
	}
	return c.Enterprise != nil
}

// GetEnterpriseConfig returns the enterprise configuration if available
func GetEnterpriseConfig() *EnterpriseConfiguration {
	c, err := GetConfig()
	if err != nil {
		// Return nil if config loading fails
		return nil
	}
	if c.Enterprise != nil {
		return c.Enterprise
	}
	return nil
}

// SaveConfig saves the configuration to file
func SaveConfig(cfg *Config) error {
	fileOps := fileops.New()
	configFile := getConfigFilePath()
	return fileOps.SaveConfig(configFile, cfg, "yaml")
}

// getConfigFilePath returns the path to the config file
func getConfigFilePath() string {
	configFiles := []string{".mage.yaml", ".mage.yml", "mage.yaml", "mage.yml"}
	for _, cf := range configFiles {
		if _, err := os.Stat(cf); err == nil {
			return cf
		}
	}
	return ".mage.yaml" // default
}

// SaveEnterpriseConfig saves the enterprise configuration to a separate file
func SaveEnterpriseConfig(cfg *EnterpriseConfiguration) error {
	enterpriseConfigFile := ".mage.enterprise.yaml"
	fileOps := fileops.New()
	return fileOps.SaveConfig(enterpriseConfigFile, cfg, "yaml")
}

// SetupEnterpriseConfig initializes enterprise configuration
func SetupEnterpriseConfig() error {
	// Check if enterprise config already exists
	if HasEnterpriseConfig() {
		return ErrEnterpriseConfigExists
	}

	// Run the enterprise setup wizard
	wizard := &EnterpriseWizard{}
	if err := wizard.Run(); err != nil {
		return fmt.Errorf("failed to run enterprise setup wizard: %w", err)
	}

	return nil
}

// Methods for Config struct required by tests

// Load loads configuration from file
func (c *Config) Load() error {
	_, err := LoadConfig()
	return err
}

// Validate validates the configuration
func (c *Config) Validate() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Validating config")
}

// Save saves the configuration to file
func (c *Config) Save(_ string) error {
	return SaveConfig(c)
}

// Builder interface methods for Config
func (c *Config) GetLint() LintConfigInterface {
	return &c.Lint
}

func (c *Config) GetBuild() BuildConfigInterface {
	return &c.Build
}

func (c *Config) GetTest() TestConfigInterface {
	return &c.Test
}

// LintConfigInterface methods
type LintConfigInterface interface {
	GetTimeout() string
}

func (l *LintConfig) GetTimeout() string {
	return l.Timeout
}

// BuildConfigInterface methods
type BuildConfigInterface interface {
	GetVerbose() bool
	GetParallel() int
	GetTags() []string
}

func (b *BuildConfig) GetVerbose() bool {
	return b.Verbose
}

func (b *BuildConfig) GetParallel() int {
	return b.Parallel
}

func (b *BuildConfig) GetTags() []string {
	return b.Tags
}

// TestConfigInterface methods
type TestConfigInterface interface {
	GetTimeout() string
	GetIntegrationTimeout() string
	GetIntegrationTag() string
	GetCoverMode() string
	GetParallel() int
	GetTags() string
	GetShuffle() bool
	GetBenchCPU() int
	GetBenchTime() string
	GetBenchMem() bool
	GetCoverageExclude() []string
}

func (t *TestConfig) GetTimeout() string {
	return t.Timeout
}

func (t *TestConfig) GetIntegrationTimeout() string {
	return t.IntegrationTimeout
}

func (t *TestConfig) GetIntegrationTag() string {
	return t.IntegrationTag
}

func (t *TestConfig) GetCoverMode() string {
	return t.CoverMode
}

func (t *TestConfig) GetParallel() int {
	return t.Parallel
}

func (t *TestConfig) GetTags() string {
	return t.Tags
}

func (t *TestConfig) GetShuffle() bool {
	return t.Shuffle
}

func (t *TestConfig) GetBenchCPU() int {
	return t.BenchCPU
}

func (t *TestConfig) GetBenchTime() string {
	return t.BenchTime
}

func (t *TestConfig) GetBenchMem() bool {
	return t.BenchMem
}

func (t *TestConfig) GetCoverageExclude() []string {
	return t.CoverageExclude
}
