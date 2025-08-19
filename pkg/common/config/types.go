package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Static errors to comply with err113 linter
var (
	errConfigNil           = errors.New("config cannot be nil")
	errProjectNameRequired = errors.New("project name is required")
	errProjectVersionReq   = errors.New("project version is required")
	errInvalidGoVersion    = errors.New("invalid Go version format")
	errTestTimeoutNegative = errors.New("test timeout cannot be negative")
)

// MageConfig represents the main configuration structure
type MageConfig struct {
	Project   ProjectConfig   `yaml:"project" json:"project"`
	Build     BuildConfig     `yaml:"build" json:"build"`
	Test      TestConfig      `yaml:"test" json:"test"`
	Analytics AnalyticsConfig `yaml:"analytics" json:"analytics"`
	Security  SecurityConfig  `yaml:"security" json:"security"`
	Deploy    DeployConfig    `yaml:"deploy" json:"deploy"`
}

// ProjectConfig contains project-level configuration
type ProjectConfig struct {
	Name        string   `yaml:"name" json:"name"`
	Version     string   `yaml:"version" json:"version"`
	Description string   `yaml:"description" json:"description"`
	Authors     []string `yaml:"authors" json:"authors"`
	License     string   `yaml:"license" json:"license"`
	Homepage    string   `yaml:"homepage" json:"homepage"`
	Repository  string   `yaml:"repository" json:"repository"`
}

// BuildConfig contains build-related configuration
type BuildConfig struct {
	GoVersion  string   `yaml:"go_version" json:"go_version"`
	Platform   string   `yaml:"platform" json:"platform"`
	Tags       []string `yaml:"tags" json:"tags"`
	LDFlags    string   `yaml:"ldflags" json:"ldflags"`
	GCFlags    string   `yaml:"gcflags" json:"gcflags"`
	CGOEnabled bool     `yaml:"cgo_enabled" json:"cgo_enabled"`
	OutputDir  string   `yaml:"output_dir" json:"output_dir"`
	Binary     string   `yaml:"binary" json:"binary"`
}

// TestConfig contains test-related configuration
type TestConfig struct {
	Timeout    int      `yaml:"timeout" json:"timeout"`
	Coverage   bool     `yaml:"coverage" json:"coverage"`
	Verbose    bool     `yaml:"verbose" json:"verbose"`
	Race       bool     `yaml:"race" json:"race"`
	Parallel   int      `yaml:"parallel" json:"parallel"`
	Tags       []string `yaml:"tags" json:"tags"`
	OutputDir  string   `yaml:"output_dir" json:"output_dir"`
	BenchTime  string   `yaml:"bench_time" json:"bench_time"`
	MemProfile bool     `yaml:"mem_profile" json:"mem_profile"`
	CPUProfile bool     `yaml:"cpu_profile" json:"cpu_profile"`
}

// AnalyticsConfig contains analytics-related configuration
type AnalyticsConfig struct {
	Enabled       bool              `yaml:"enabled" json:"enabled"`
	SampleRate    float64           `yaml:"sample_rate" json:"sample_rate"`
	RetentionDays int               `yaml:"retention_days" json:"retention_days"`
	ExportFormats []string          `yaml:"export_formats" json:"export_formats"`
	Endpoints     map[string]string `yaml:"endpoints" json:"endpoints"`
	BatchSize     int               `yaml:"batch_size" json:"batch_size"`
	FlushInterval int               `yaml:"flush_interval" json:"flush_interval"`
}

// SecurityConfig contains security-related configuration
type SecurityConfig struct {
	EnableVulnCheck  bool     `yaml:"enable_vuln_check" json:"enable_vuln_check"`
	SkipVulnCheck    []string `yaml:"skip_vuln_check" json:"skip_vuln_check"`
	RequiredChecks   []string `yaml:"required_checks" json:"required_checks"`
	PolicyFile       string   `yaml:"policy_file" json:"policy_file"`
	EnableCodeScan   bool     `yaml:"enable_code_scan" json:"enable_code_scan"`
	EnableSecretScan bool     `yaml:"enable_secret_scan" json:"enable_secret_scan"`
}

// DeployConfig contains deployment-related configuration
type DeployConfig struct {
	Strategy    string            `yaml:"strategy" json:"strategy"`
	Environment string            `yaml:"environment" json:"environment"`
	Variables   map[string]string `yaml:"variables" json:"variables"`
	Hooks       DeployHooks       `yaml:"hooks" json:"hooks"`
	Rollback    RollbackConfig    `yaml:"rollback" json:"rollback"`
}

// DeployHooks contains deployment hook configuration
type DeployHooks struct {
	PreDeploy  []string `yaml:"pre_deploy" json:"pre_deploy"`
	PostDeploy []string `yaml:"post_deploy" json:"post_deploy"`
	OnFailure  []string `yaml:"on_failure" json:"on_failure"`
	OnSuccess  []string `yaml:"on_success" json:"on_success"`
}

// RollbackConfig contains rollback configuration
type RollbackConfig struct {
	Enabled          bool   `yaml:"enabled" json:"enabled"`
	MaxVersions      int    `yaml:"max_versions" json:"max_versions"`
	AutoRollback     bool   `yaml:"auto_rollback" json:"auto_rollback"`
	HealthCheckURL   string `yaml:"health_check_url" json:"health_check_url"`
	HealthCheckRetry int    `yaml:"health_check_retry" json:"health_check_retry"`
}

// MageConfigManager provides configuration management functionality for MageConfig
type MageConfigManager interface {
	Load() (*MageConfig, error)
	LoadFromPath(path string) (*MageConfig, error)
	Save(config *MageConfig, path string) error
	Validate(config *MageConfig) error
	GetDefaults() *MageConfig
	Merge(configs ...*MageConfig) *MageConfig
}

// MageLoader provides a simple interface for loading mage configuration
type MageLoader interface {
	Load() (*MageConfig, error)
	LoadFromPath(path string) (*MageConfig, error)
	Save(config *MageConfig) error
	Validate(config *MageConfig) error
	GetDefaults() *MageConfig
	Merge(configs ...*MageConfig) *MageConfig
}

// NewManager creates a new MageConfigManager instance
func NewManager() MageConfigManager {
	return &configManagerImpl{}
}

// NewLoader creates a new MageLoader instance
func NewLoader() MageLoader {
	return &loaderImpl{}
}

// cleanEnvValue removes inline comments and trims whitespace from environment variable values
func cleanEnvValue(value string) string {
	if value == "" {
		return ""
	}

	// Find inline comment marker (space followed by #)
	if idx := strings.Index(value, " #"); idx >= 0 {
		value = value[:idx]
	}

	// Trim any leading/trailing whitespace
	return strings.TrimSpace(value)
}

// getDefaultGoVersion returns the default Go version from environment or fallback
func getDefaultGoVersion() string {
	// Check primary environment variable
	if value := cleanEnvValue(os.Getenv("MAGE_X_GO_VERSION")); value != "" {
		// Clean up the version to remove any .x suffix for actual usage
		if len(value) > 2 && value[len(value)-2:] == ".x" {
			return value[:len(value)-2]
		}
		return value
	}

	// Check legacy environment variable for backward compatibility
	if value := os.Getenv("GO_PRIMARY_VERSION"); value != "" {
		// Clean up the version to remove any .x suffix for actual usage
		if len(value) > 2 && value[len(value)-2:] == ".x" {
			return value[:len(value)-2]
		}
		return value
	}

	// Fallback if environment is not set
	return "1.24"
}

// configManagerImpl implements ConfigManager
type configManagerImpl struct{}

// Load loads the default configuration
func (m *configManagerImpl) Load() (*MageConfig, error) {
	return m.GetDefaults(), nil
}

// LoadFromPath loads configuration from a specific path
func (m *configManagerImpl) LoadFromPath(path string) (*MageConfig, error) {
	loader := NewDefaultConfigLoader()
	config := &MageConfig{}

	if err := loader.LoadFrom(path, config); err != nil {
		return nil, err
	}

	return config, nil
}

// Save saves configuration to a path
func (m *configManagerImpl) Save(config *MageConfig, path string) error {
	loader := NewDefaultConfigLoader()

	// Detect format from file extension
	ext := strings.ToLower(filepath.Ext(path))
	var format string
	switch ext {
	case ".json":
		format = "json"
	case ".yaml", ".yml":
		format = string(FormatYAML)
	default:
		format = string(FormatYAML)
	}

	return loader.Save(path, config, format)
}

// Validate validates a configuration
func (m *configManagerImpl) Validate(config *MageConfig) error {
	if config == nil {
		return errConfigNil
	}

	// Validate project fields
	if config.Project.Name == "" {
		return errProjectNameRequired
	}
	if config.Project.Version == "" {
		return errProjectVersionReq
	}

	// Validate build fields
	if config.Build.GoVersion != "" {
		// Simple Go version validation - should be like "1.24", "1.20", etc.
		if len(config.Build.GoVersion) < 4 || config.Build.GoVersion[:2] != "1." {
			return fmt.Errorf("%w: %s", errInvalidGoVersion, config.Build.GoVersion)
		}
	}

	// Validate test fields
	if config.Test.Timeout < 0 {
		return fmt.Errorf("%w: %d", errTestTimeoutNegative, config.Test.Timeout)
	}

	return nil
}

// GetDefaults returns default configuration
func (m *configManagerImpl) GetDefaults() *MageConfig {
	return &MageConfig{
		Project: ProjectConfig{
			Name:    "mage-project",
			Version: "1.0.0",
		},
		Build: BuildConfig{
			GoVersion:  getDefaultGoVersion(),
			Platform:   "linux/amd64",
			CGOEnabled: false,
		},
		Test: TestConfig{
			Timeout:  120,
			Coverage: true,
			Parallel: 4,
		},
		Analytics: AnalyticsConfig{
			Enabled:    false,
			SampleRate: 0.1,
		},
	}
}

// Merge merges multiple configurations
func (m *configManagerImpl) Merge(configs ...*MageConfig) *MageConfig {
	if len(configs) == 0 {
		return m.GetDefaults()
	}

	result := m.initializeMergeResult(configs[0])

	for i := 1; i < len(configs); i++ {
		m.mergeConfig(result, configs[i])
	}

	return result
}

// initializeMergeResult creates a deep copy of the base configuration
func (m *configManagerImpl) initializeMergeResult(base *MageConfig) *MageConfig {
	result := &MageConfig{}
	*result = *base

	m.deepCopySlices(result, base)
	return result
}

// deepCopySlices creates deep copies of slices to avoid shared references
func (m *configManagerImpl) deepCopySlices(result, base *MageConfig) {
	if base.Project.Authors != nil {
		result.Project.Authors = make([]string, len(base.Project.Authors))
		copy(result.Project.Authors, base.Project.Authors)
	}
	if base.Build.Tags != nil {
		result.Build.Tags = make([]string, len(base.Build.Tags))
		copy(result.Build.Tags, base.Build.Tags)
	}
	if base.Test.Tags != nil {
		result.Test.Tags = make([]string, len(base.Test.Tags))
		copy(result.Test.Tags, base.Test.Tags)
	}
}

// mergeConfig merges a single override configuration into the result
func (m *configManagerImpl) mergeConfig(result, override *MageConfig) {
	m.mergeProjectConfig(&result.Project, &override.Project)
	m.mergeBuildConfig(&result.Build, &override.Build)
	m.mergeTestConfig(&result.Test, &override.Test)
	m.mergeAnalyticsConfig(&result.Analytics, &override.Analytics)
}

// mergeProjectConfig merges project configuration fields
func (m *configManagerImpl) mergeProjectConfig(result, override *ProjectConfig) {
	if override.Name != "" {
		result.Name = override.Name
	}
	if override.Version != "" {
		result.Version = override.Version
	}
	if override.Description != "" {
		result.Description = override.Description
	}
	if override.License != "" {
		result.License = override.License
	}
	if override.Homepage != "" {
		result.Homepage = override.Homepage
	}
	if override.Repository != "" {
		result.Repository = override.Repository
	}
	if override.Authors != nil {
		result.Authors = make([]string, len(override.Authors))
		copy(result.Authors, override.Authors)
	}
}

// mergeBuildConfig merges build configuration fields
func (m *configManagerImpl) mergeBuildConfig(result, override *BuildConfig) {
	if override.GoVersion != "" {
		result.GoVersion = override.GoVersion
	}
	if override.Platform != "" {
		result.Platform = override.Platform
	}
	if override.LDFlags != "" {
		result.LDFlags = override.LDFlags
	}
	if override.GCFlags != "" {
		result.GCFlags = override.GCFlags
	}
	if override.OutputDir != "" {
		result.OutputDir = override.OutputDir
	}
	if override.Binary != "" {
		result.Binary = override.Binary
	}
	if override.Tags != nil {
		result.Tags = make([]string, len(override.Tags))
		copy(result.Tags, override.Tags)
	}

	// CGOEnabled is a bool, override if any build settings are provided
	if m.hasBuildOverrides(override) {
		result.CGOEnabled = override.CGOEnabled
	}
}

// hasBuildOverrides checks if the override config has any build settings
func (m *configManagerImpl) hasBuildOverrides(override *BuildConfig) bool {
	return override.GoVersion != "" || override.Platform != "" || override.Tags != nil
}

// mergeTestConfig merges test configuration fields
func (m *configManagerImpl) mergeTestConfig(result, override *TestConfig) {
	if override.Timeout != 0 {
		result.Timeout = override.Timeout
	}
	if override.Parallel != 0 {
		result.Parallel = override.Parallel
	}
	if override.OutputDir != "" {
		result.OutputDir = override.OutputDir
	}
	if override.BenchTime != "" {
		result.BenchTime = override.BenchTime
	}
	if override.Tags != nil {
		result.Tags = make([]string, len(override.Tags))
		copy(result.Tags, override.Tags)
	}

	// Boolean fields - override if any test config is provided
	if m.hasTestOverrides(override) {
		result.Coverage = override.Coverage
		result.Verbose = override.Verbose
		result.Race = override.Race
		result.MemProfile = override.MemProfile
		result.CPUProfile = override.CPUProfile
	}
}

// hasTestOverrides checks if the override config has any test settings
func (m *configManagerImpl) hasTestOverrides(override *TestConfig) bool {
	return override.Timeout != 0 || override.Parallel != 0 || override.Tags != nil
}

// mergeAnalyticsConfig merges analytics configuration fields
func (m *configManagerImpl) mergeAnalyticsConfig(result, override *AnalyticsConfig) {
	// Simple override for analytics configuration
	if override.SampleRate != 0 || override.RetentionDays != 0 {
		*result = *override
	}
}

// loaderImpl implements Loader
type loaderImpl struct {
	manager MageConfigManager
}

func (l *loaderImpl) Load() (*MageConfig, error) {
	if l.manager == nil {
		l.manager = NewManager()
	}
	return l.manager.Load()
}

func (l *loaderImpl) LoadFromPath(path string) (*MageConfig, error) {
	if l.manager == nil {
		l.manager = NewManager()
	}
	return l.manager.LoadFromPath(path)
}

func (l *loaderImpl) Save(config *MageConfig) error {
	if l.manager == nil {
		l.manager = NewManager()
	}
	return l.manager.Save(config, "mage.yaml")
}

func (l *loaderImpl) Validate(config *MageConfig) error {
	if l.manager == nil {
		l.manager = NewManager()
	}
	return l.manager.Validate(config)
}

func (l *loaderImpl) GetDefaults() *MageConfig {
	if l.manager == nil {
		l.manager = NewManager()
	}
	return l.manager.GetDefaults()
}

func (l *loaderImpl) Merge(configs ...*MageConfig) *MageConfig {
	if l.manager == nil {
		l.manager = NewManager()
	}
	return l.manager.Merge(configs...)
}
