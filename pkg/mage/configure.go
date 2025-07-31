// Package mage provides configuration management commands
package mage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/common/fileops"
	"github.com/mrz1836/go-mage/pkg/utils"
	"gopkg.in/yaml.v3"
)

const (
	errUnexpectedNewline = "unexpected newline"
)

// Configure namespace for configuration management
type Configure mg.Namespace

// Init initializes a new mage configuration
func (Configure) Init() error {
	utils.Header("🔧 Initialize Mage Configuration")

	// Check if config already exists
	configFiles := []string{".mage.yaml", ".mage.yml", "mage.yaml", "mage.yml"}
	for _, cf := range configFiles {
		if _, err := os.Stat(cf); err == nil {
			return fmt.Errorf("configuration file %s already exists", cf)
		}
	}

	// Create default configuration
	config := defaultConfig()

	// Save configuration
	if err := SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	utils.Success("✅ Configuration initialized: %s", getConfigFilePath())
	return nil
}

// Show displays the current configuration
func (Configure) Show() error {
	utils.Header("📋 Current Configuration")

	config, err := GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Display basic configuration
	utils.Info("📁 Project: %s", config.Project.Name)
	utils.Info("🔧 Binary: %s", config.Project.Binary)
	utils.Info("📦 Module: %s", config.Project.Module)
	utils.Info("🌍 Git Domain: %s", config.Project.GitDomain)

	if config.Project.RepoOwner != "" {
		utils.Info("👤 Repository: %s/%s", config.Project.RepoOwner, config.Project.RepoName)
	}

	// Display build configuration
	utils.Info("\n🏗️  Build Configuration:")
	utils.Info("  Output: %s", config.Build.Output)
	utils.Info("  Parallel: %d", config.Build.Parallel)
	utils.Info("  Platforms: %s", strings.Join(config.Build.Platforms, ", "))
	utils.Info("  Trim Path: %v", config.Build.TrimPath)

	// Display test configuration
	utils.Info("\n🧪 Test Configuration:")
	utils.Info("  Parallel: %v", config.Test.Parallel)
	utils.Info("  Timeout: %s", config.Test.Timeout)
	utils.Info("  Cover Mode: %s", config.Test.CoverMode)
	utils.Info("  Race Detection: %v", config.Test.Race)

	// Display enterprise configuration if available
	if config.Enterprise != nil {
		utils.Info("\n🏢 Enterprise Configuration:")
		utils.Info("  Organization: %s", config.Enterprise.Organization.Name)
		utils.Info("  Domain: %s", config.Enterprise.Organization.Domain)
		utils.Info("  Security Level: %s", config.Enterprise.Security.Level)
		utils.Info("  Analytics: %v", config.Enterprise.Analytics.Enabled)
		utils.Info("  Integrations: %d configured", len(config.Enterprise.Integrations.Providers))
	}

	return nil
}

// Update updates configuration values interactively
func (Configure) Update() error {
	utils.Header("🔄 Update Configuration")

	config, err := GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Interactive update wizard
	wizard := &ConfigurationWizard{
		Config: config,
	}

	if err := wizard.Run(); err != nil {
		return fmt.Errorf("failed to run configuration wizard: %w", err)
	}

	utils.Success("✅ Configuration updated successfully")
	return nil
}

// Enterprise initializes enterprise configuration
func (Configure) Enterprise() error {
	utils.Header("🏢 Enterprise Configuration Setup")

	// Check if enterprise config already exists
	if HasEnterpriseConfig() {
		utils.Info("Enterprise configuration already exists")

		// Ask if user wants to update
		fmt.Print("Would you like to update the existing configuration? (y/N): ")
		var response string
		if _, err := fmt.Scanln(&response); err != nil && err.Error() != errUnexpectedNewline {
			return fmt.Errorf("failed to read response: %w", err)
		}

		if !strings.EqualFold(response, "y") && !strings.EqualFold(response, "yes") {
			return nil
		}
	}

	// Run enterprise setup wizard
	return SetupEnterpriseConfig()
}

// Export exports configuration to different formats
func (Configure) Export() error {
	utils.Header("📤 Export Configuration")

	config, err := GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	format := utils.GetEnv("FORMAT", "yaml")
	output := utils.GetEnv("OUTPUT", "")

	var data []byte
	var ext string

	switch format {
	case "yaml", "yml":
		data, err = yaml.Marshal(config)
		ext = ".yaml"
	case "json":
		data, err = marshalJSON(config)
		ext = ".json"
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	if output == "" {
		// Print to stdout
		utils.Info("%s", string(data))
	} else {
		// Save to file
		if !strings.HasSuffix(output, ext) {
			output += ext
		}

		fileOps := fileops.New()
		if err := fileOps.File.WriteFile(output, data, 0o644); err != nil {
			return fmt.Errorf("failed to write configuration: %w", err)
		}

		utils.Success("✅ Configuration exported to: %s", output)
	}

	return nil
}

// Import imports configuration from file
func (Configure) Import() error {
	utils.Header("📥 Import Configuration")

	importFile := utils.GetEnv("FILE", "")
	if importFile == "" {
		return fmt.Errorf("FILE environment variable is required")
	}

	if _, err := os.Stat(importFile); os.IsNotExist(err) {
		return fmt.Errorf("import file not found: %s", importFile)
	}

	// Read import file
	fileOps := fileops.New()
	data, err := fileOps.File.ReadFile(importFile)
	if err != nil {
		return fmt.Errorf("failed to read import file: %w", err)
	}

	// Parse configuration
	var cfg Config
	ext := filepath.Ext(importFile)

	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &cfg)
	case ".json":
		err = unmarshalJSON(data, &cfg)
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}

	if err != nil {
		return fmt.Errorf("failed to parse import file: %w", err)
	}

	// Validate configuration
	if err := validateConfiguration(&cfg); err != nil {
		return fmt.Errorf("imported configuration is invalid: %w", err)
	}

	// Save configuration
	if err := SaveConfig(&cfg); err != nil {
		return fmt.Errorf("failed to save imported configuration: %w", err)
	}

	utils.Success("✅ Configuration imported successfully")
	return nil
}

// Validate validates the current configuration
func (Configure) Validate() error {
	utils.Header("🔍 Validate Configuration")

	config, err := GetConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := validateConfiguration(config); err != nil {
		utils.Error("❌ Configuration validation failed: %v", err)
		return err
	}

	utils.Success("✅ Configuration is valid")
	return nil
}

// Schema generates JSON schema for configuration
func (Configure) Schema() error {
	utils.Header("📋 Configuration Schema")

	output := utils.GetEnv("OUTPUT", "")

	schema := generateConfigurationSchema()

	if output == "" {
		// Print to stdout
		utils.Info("%s", schema)
	} else {
		fileOps := fileops.New()
		if err := fileOps.File.WriteFile(output, []byte(schema), 0o644); err != nil {
			return fmt.Errorf("failed to write schema: %w", err)
		}

		utils.Success("✅ Configuration schema generated: %s", output)
	}

	return nil
}

// Supporting types and functions

// ConfigurationWizard provides interactive configuration updates
type ConfigurationWizard struct {
	Config *Config
}

// Run executes the configuration wizard
func (w *ConfigurationWizard) Run() error {
	utils.Info("🧙 Configuration Update Wizard")

	// Update project configuration
	if err := w.updateProjectConfig(); err != nil {
		return err
	}

	// Update build configuration
	if err := w.updateBuildConfig(); err != nil {
		return err
	}

	// Update test configuration
	if err := w.updateTestConfig(); err != nil {
		return err
	}

	// Save updated configuration
	return SaveConfig(w.Config)
}

func (w *ConfigurationWizard) updateProjectConfig() error {
	utils.Info("📁 Project Configuration")

	fmt.Printf("Project Name [%s]: ", w.Config.Project.Name)
	var name string
	if _, err := fmt.Scanln(&name); err != nil && err.Error() != errUnexpectedNewline {
		// Handle scan errors other than empty input
		utils.Error("Error reading input: %v", err)
	}
	if name != "" {
		w.Config.Project.Name = name
	}

	fmt.Printf("Binary Name [%s]: ", w.Config.Project.Binary)
	var binary string
	if _, err := fmt.Scanln(&binary); err != nil && err.Error() != errUnexpectedNewline {
		// Handle scan errors other than empty input
		utils.Error("Error reading input: %v", err)
	}
	if binary != "" {
		w.Config.Project.Binary = binary
	}

	fmt.Printf("Module Path [%s]: ", w.Config.Project.Module)
	var module string
	if _, err := fmt.Scanln(&module); err != nil && err.Error() != errUnexpectedNewline {
		// Handle scan errors other than empty input
		utils.Error("Error reading input: %v", err)
	}
	if module != "" {
		w.Config.Project.Module = module
	}

	return nil
}

func (w *ConfigurationWizard) updateBuildConfig() error {
	utils.Info("🏗️  Build Configuration")

	fmt.Printf("Output Directory [%s]: ", w.Config.Build.Output)
	var output string
	if _, err := fmt.Scanln(&output); err != nil && err.Error() != errUnexpectedNewline {
		// Handle scan errors other than empty input
		utils.Error("Error reading input: %v", err)
	}
	if output != "" {
		w.Config.Build.Output = output
	}

	fmt.Printf("Parallel Jobs [%d]: ", w.Config.Build.Parallel)
	var parallel int
	if _, err := fmt.Scanf("%d", &parallel); err != nil {
		// Invalid input, keep default
		parallel = 0
	}
	if parallel > 0 {
		w.Config.Build.Parallel = parallel
	}

	return nil
}

func (w *ConfigurationWizard) updateTestConfig() error {
	utils.Info("🧪 Test Configuration")

	fmt.Printf("Test Timeout [%s]: ", w.Config.Test.Timeout)
	var timeout string
	if _, err := fmt.Scanln(&timeout); err != nil && err.Error() != errUnexpectedNewline {
		// Handle scan errors other than empty input
		utils.Error("Error reading input: %v", err)
	}
	if timeout != "" {
		w.Config.Test.Timeout = timeout
	}

	fmt.Printf("Enable Race Detection [%v]: ", w.Config.Test.Race)
	var race string
	if _, err := fmt.Scanln(&race); err != nil && err.Error() != errUnexpectedNewline {
		// Handle scan errors other than empty input
		utils.Error("Error reading input: %v", err)
	}
	if race != "" {
		w.Config.Test.Race = strings.EqualFold(race, "true") || race == "y"
	}

	return nil
}

func validateConfiguration(cfg *Config) error {
	if cfg == nil {
		// If no config provided, try to load default config
		var err error
		cfg, err = GetConfig()
		if err != nil {
			return fmt.Errorf("failed to load configuration: %w", err)
		}
	}

	if cfg.Project.Name == "" {
		return fmt.Errorf("project name is required")
	}

	if cfg.Project.Binary == "" {
		return fmt.Errorf("binary name is required")
	}

	if cfg.Project.Module == "" {
		return fmt.Errorf("module path is required")
	}

	// Validate enterprise configuration if present
	if cfg.Enterprise != nil {
		return validateEnterpriseConfiguration(*cfg.Enterprise)
	}

	return nil
}

// validateEnterpriseConfiguration validates enterprise configuration settings
func validateEnterpriseConfiguration(cfg EnterpriseConfiguration) error {
	if cfg.Organization.Name == "" {
		return fmt.Errorf("enterprise organization name is required")
	}
	if cfg.Organization.Domain == "" {
		return fmt.Errorf("enterprise organization domain is required")
	}
	return nil
}

func generateConfigurationSchema() string {
	// Generate JSON schema for configuration
	return `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "MAGE-X Configuration",
  "type": "object",
  "properties": {
    "project": {
      "type": "object",
      "properties": {
        "name": {"type": "string"},
        "binary": {"type": "string"},
        "module": {"type": "string"},
        "description": {"type": "string"},
        "version": {"type": "string"},
        "git_domain": {"type": "string"},
        "repo_owner": {"type": "string"},
        "repo_name": {"type": "string"},
        "env": {"type": "object"}
      },
      "required": ["name", "binary", "module"]
    },
    "build": {
      "type": "object",
      "properties": {
        "tags": {"type": "array", "items": {"type": "string"}},
        "ldflags": {"type": "array", "items": {"type": "string"}},
        "platforms": {"type": "array", "items": {"type": "string"}},
        "goflags": {"type": "array", "items": {"type": "string"}},
        "output": {"type": "string"},
        "parallel": {"type": "integer"},
        "verbose": {"type": "boolean"},
        "trimpath": {"type": "boolean"}
      }
    },
    "test": {
      "type": "object",
      "properties": {
        "parallel": {"type": "boolean"},
        "timeout": {"type": "string"},
        "short": {"type": "boolean"},
        "verbose": {"type": "boolean"},
        "race": {"type": "boolean"},
        "cover": {"type": "boolean"},
        "covermode": {"type": "string"},
        "coverpkg": {"type": "array", "items": {"type": "string"}},
        "tags": {"type": "array", "items": {"type": "string"}},
        "skip_fuzz": {"type": "boolean"}
      }
    },
    "enterprise": {
      "type": "object",
      "description": "Enterprise configuration for advanced features"
    }
  },
  "required": ["project", "build", "test"]
}`
}

func marshalJSON(v interface{}) ([]byte, error) {
	// Use yaml package to marshal to JSON-compatible format
	return yaml.Marshal(v)
}

func unmarshalJSON(data []byte, v interface{}) error {
	// Use yaml package to unmarshal JSON
	return yaml.Unmarshal(data, v)
}
