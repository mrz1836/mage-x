// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/mrz1836/go-mage/pkg/common/fileops"
	"gopkg.in/yaml.v3"
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
	Parallel  bool     `yaml:"parallel"`
	Timeout   string   `yaml:"timeout"`
	Short     bool     `yaml:"short"`
	Verbose   bool     `yaml:"verbose"`
	Race      bool     `yaml:"race"`
	Cover     bool     `yaml:"cover"`
	CoverMode string   `yaml:"covermode"`
	CoverPkg  []string `yaml:"coverpkg"`
	Tags      []string `yaml:"tags"`
	SkipFuzz  bool     `yaml:"skip_fuzz"`
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

var (
	// configOnce ensures configuration is loaded only once
	configOnce sync.Once //nolint:gochecknoglobals // Required for singleton initialization
	// loadedConfig holds the loaded configuration
	loadedConfig *Config //nolint:gochecknoglobals // Required for singleton configuration
	// errLoadConfig stores any error encountered during config loading
	errLoadConfig error
	// cfg is a backward compatibility variable for tests
	cfg *Config //nolint:gochecknoglobals // Backward compatibility for tests
)

// LoadConfig loads the configuration from file or returns defaults
func LoadConfig() (*Config, error) {
	// Check if cfg is already set by tests
	if cfg != nil {
		return cfg, nil
	}

	configOnce.Do(func() {
		loadedConfig = defaultConfig()
		configFile := ".mage.yaml"

		// Try multiple config file names
		configFiles := []string{".mage.yaml", ".mage.yml", "mage.yaml", "mage.yml"}

		for _, cf := range configFiles {
			if _, err := os.Stat(cf); err == nil {
				configFile = cf
				break
			}
		}

		// Check for enterprise configuration file
		enterpriseConfigFile := ".mage.enterprise.yaml"
		fileOps := fileops.New()
		if fileOps.File.Exists(enterpriseConfigFile) {
			// Load enterprise configuration
			var enterpriseConfig EnterpriseConfiguration
			if err := fileOps.YAML.ReadYAML(enterpriseConfigFile, &enterpriseConfig); err == nil {
				loadedConfig.Enterprise = &enterpriseConfig
			}
		}

		if !fileOps.File.Exists(configFile) {
			// Config file doesn't exist, use defaults
			return
		}

		data, err := fileOps.File.ReadFile(configFile)
		if err != nil {
			errLoadConfig = fmt.Errorf("failed to read config: %w", err)
			return
		}

		if err := yaml.Unmarshal(data, loadedConfig); err != nil {
			errLoadConfig = fmt.Errorf("failed to parse config: %w", err)
			return
		}

		// Apply environment variable overrides
		applyEnvOverrides(loadedConfig)

		// TODO: Add enterprise configuration validation when EnterpriseConfiguration type is defined
	})
	cfg = loadedConfig // Keep backward compatibility variable in sync
	return loadedConfig, errLoadConfig
}

// GetConfig is an alias for LoadConfig for backward compatibility
func GetConfig() (*Config, error) {
	return LoadConfig()
}

// TestResetConfig resets the config for testing purposes only
// This should only be used in tests
func TestResetConfig() {
	configOnce = sync.Once{}
	loadedConfig = nil
	errLoadConfig = nil
	cfg = nil
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
	if binary == "." || binary == "" {
		binary = "app"
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
			Parallel:  true,
			Timeout:   "10m",
			CoverMode: "atomic",
		},
		Lint: LintConfig{
			GolangciVersion: "v2.3.0",
			Timeout:         "5m",
		},
		Tools: ToolsConfig{
			GolangciLint: "v2.3.0",
			Fumpt:        "latest",
			GoVulnCheck:  "latest",
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
	}
}

// applyEnvOverrides applies environment variable overrides to config
func applyEnvOverrides(c *Config) {
	// Binary name override
	if v := os.Getenv("BINARY_NAME"); v != "" {
		c.Project.Binary = v
	}
	if v := os.Getenv("CUSTOM_BINARY_NAME"); v != "" {
		c.Project.Binary = v
	}

	// Build tags override
	if v := os.Getenv("GO_BUILD_TAGS"); v != "" {
		c.Build.Tags = strings.Split(v, ",")
	}

	// Verbose override
	if v := os.Getenv("VERBOSE"); v == "true" || v == "1" {
		c.Build.Verbose = true
		c.Test.Verbose = true
	}

	// Test race override
	if v := os.Getenv("TEST_RACE"); v == "true" || v == "1" {
		c.Test.Race = true
	}

	// Parallel override
	if v := os.Getenv("PARALLEL"); v != "" {
		var parallel int
		if _, err := fmt.Sscanf(v, "%d", &parallel); err == nil && parallel > 0 {
			c.Build.Parallel = parallel
		}
	}

	// Enterprise overrides
	if c.Enterprise != nil {
		applyEnterpriseEnvOverrides(c.Enterprise)
	}
}

// applyEnterpriseEnvOverrides applies environment variable overrides to enterprise config
func applyEnterpriseEnvOverrides(cfg *EnterpriseConfiguration) {
	// Organization overrides
	if v := os.Getenv("MAGE_ORG_NAME"); v != "" {
		cfg.Organization.Name = v
	}
	if v := os.Getenv("MAGE_ORG_DOMAIN"); v != "" {
		cfg.Organization.Domain = v
	}

	// TODO: Add security configuration when Security field is added to Config
	_ = os.Getenv("MAGE_SECURITY_LEVEL") // placeholder for security level
	// TODO: Add vault integration when Security field is added to Config
	_ = os.Getenv("MAGE_ENABLE_VAULT") // placeholder for vault enabled
	_ = os.Getenv("VAULT_ADDR")        // placeholder for vault address

	// TODO: Add analytics configuration when Analytics field is added to Config
	_ = os.Getenv("MAGE_ANALYTICS_ENABLED") // placeholder for analytics enabled
	_ = os.Getenv("MAGE_METRICS_INTERVAL")  // placeholder for metrics interval
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
		return fmt.Errorf("enterprise configuration already exists")
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
func (c *Config) Save(filename string) error {
	return SaveConfig(c)
}
