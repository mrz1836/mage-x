package config

import (
	"fmt"
	"path/filepath"
	"strings"
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

// Loader provides a simple interface for loading configuration
type Loader interface {
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

// NewLoader creates a new Loader instance
func NewLoader() Loader {
	return &loaderImpl{}
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
	format := "yaml"
	switch ext {
	case ".json":
		format = "json"
	case ".yaml", ".yml":
		format = "yaml"
	default:
		format = "yaml"
	}

	return loader.Save(path, config, format)
}

// Validate validates a configuration
func (m *configManagerImpl) Validate(config *MageConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Validate project fields
	if config.Project.Name == "" {
		return fmt.Errorf("project name is required")
	}
	if config.Project.Version == "" {
		return fmt.Errorf("project version is required")
	}

	// Validate build fields
	if config.Build.GoVersion != "" {
		// Simple Go version validation - should be like "1.24", "1.20", etc.
		if len(config.Build.GoVersion) < 4 || config.Build.GoVersion[:2] != "1." {
			return fmt.Errorf("invalid Go version format: %s", config.Build.GoVersion)
		}
	}

	// Validate test fields
	if config.Test.Timeout < 0 {
		return fmt.Errorf("test timeout cannot be negative: %d", config.Test.Timeout)
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
			GoVersion:  "1.24",
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

	// Start with a copy of the first config
	result := &MageConfig{}
	*result = *configs[0]

	// Deep copy slices and maps to avoid shared references
	if configs[0].Project.Authors != nil {
		result.Project.Authors = make([]string, len(configs[0].Project.Authors))
		copy(result.Project.Authors, configs[0].Project.Authors)
	}
	if configs[0].Build.Tags != nil {
		result.Build.Tags = make([]string, len(configs[0].Build.Tags))
		copy(result.Build.Tags, configs[0].Build.Tags)
	}
	if configs[0].Test.Tags != nil {
		result.Test.Tags = make([]string, len(configs[0].Test.Tags))
		copy(result.Test.Tags, configs[0].Test.Tags)
	}

	// Merge subsequent configs
	for i := 1; i < len(configs); i++ {
		override := configs[i]

		// Project fields
		if override.Project.Name != "" {
			result.Project.Name = override.Project.Name
		}
		if override.Project.Version != "" {
			result.Project.Version = override.Project.Version
		}
		if override.Project.Description != "" {
			result.Project.Description = override.Project.Description
		}
		if override.Project.License != "" {
			result.Project.License = override.Project.License
		}
		if override.Project.Homepage != "" {
			result.Project.Homepage = override.Project.Homepage
		}
		if override.Project.Repository != "" {
			result.Project.Repository = override.Project.Repository
		}
		if override.Project.Authors != nil {
			result.Project.Authors = make([]string, len(override.Project.Authors))
			copy(result.Project.Authors, override.Project.Authors)
		}

		// Build fields
		if override.Build.GoVersion != "" {
			result.Build.GoVersion = override.Build.GoVersion
		}
		if override.Build.Platform != "" {
			result.Build.Platform = override.Build.Platform
		}
		if override.Build.LDFlags != "" {
			result.Build.LDFlags = override.Build.LDFlags
		}
		if override.Build.GCFlags != "" {
			result.Build.GCFlags = override.Build.GCFlags
		}
		if override.Build.OutputDir != "" {
			result.Build.OutputDir = override.Build.OutputDir
		}
		if override.Build.Binary != "" {
			result.Build.Binary = override.Build.Binary
		}
		if override.Build.Tags != nil {
			result.Build.Tags = make([]string, len(override.Build.Tags))
			copy(result.Build.Tags, override.Build.Tags)
		}
		// CGOEnabled is a bool, so we need to check if it should be overridden
		// For simplicity, we'll always override if the override config has any build settings
		if override.Build.GoVersion != "" || override.Build.Platform != "" || override.Build.Tags != nil {
			result.Build.CGOEnabled = override.Build.CGOEnabled
		}

		// Test fields
		if override.Test.Timeout != 0 {
			result.Test.Timeout = override.Test.Timeout
		}
		if override.Test.Parallel != 0 {
			result.Test.Parallel = override.Test.Parallel
		}
		if override.Test.OutputDir != "" {
			result.Test.OutputDir = override.Test.OutputDir
		}
		if override.Test.BenchTime != "" {
			result.Test.BenchTime = override.Test.BenchTime
		}
		if override.Test.Tags != nil {
			result.Test.Tags = make([]string, len(override.Test.Tags))
			copy(result.Test.Tags, override.Test.Tags)
		}
		// Boolean fields - override if any test config is provided
		if override.Test.Timeout != 0 || override.Test.Parallel != 0 || override.Test.Tags != nil {
			result.Test.Coverage = override.Test.Coverage
			result.Test.Verbose = override.Test.Verbose
			result.Test.Race = override.Test.Race
			result.Test.MemProfile = override.Test.MemProfile
			result.Test.CPUProfile = override.Test.CPUProfile
		}

		// Analytics fields (simple override for now)
		if override.Analytics.SampleRate != 0 || override.Analytics.RetentionDays != 0 {
			result.Analytics = override.Analytics
		}
	}

	return result
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
