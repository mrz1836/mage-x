// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"errors"
	"fmt"
	"strings"

	"github.com/magefile/mage/mg"

	"github.com/mrz1836/mage-x/pkg/common/env"
	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors to satisfy err113 linter
var (
	errMageYamlExists            = errors.New("mage.yaml already exists")
	errProjectNameRequired       = errors.New("project.name is required")
	errProjectModuleRequired     = errors.New("project.module is required")
	errBuildOutputRequired       = errors.New("build.output_dir is required")
	errYamlInvalidPlatformFormat = errors.New("invalid platform format")
)

// Yaml namespace for mage.yaml configuration management
type Yaml mg.Namespace

// YamlConfig represents the complete mage.yaml configuration
type YamlConfig struct {
	Version string            `yaml:"version"`
	Project ProjectYamlConfig `yaml:"project"`
	Build   BuildYamlConfig   `yaml:"build"`
	Test    TestYamlConfig    `yaml:"test"`
	Lint    LintYamlConfig    `yaml:"lint"`
	Release ReleaseYamlConfig `yaml:"release"`
	CI      CIYamlConfig      `yaml:"ci"`
}

// ProjectYamlConfig contains project-specific configuration
type ProjectYamlConfig struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Module      string   `yaml:"module"`
	MainPkg     string   `yaml:"main_pkg"`
	BinaryName  string   `yaml:"binary_name"`
	Version     string   `yaml:"version"`
	Authors     []string `yaml:"authors"`
	License     string   `yaml:"license"`
	Homepage    string   `yaml:"homepage"`
	Repository  string   `yaml:"repository"`
}

// BuildYamlConfig contains build configuration
type BuildYamlConfig struct {
	Dir         string            `yaml:"dir"`
	OutputDir   string            `yaml:"output_dir"`
	LDFlags     string            `yaml:"ldflags"`
	Tags        []string          `yaml:"tags"`
	Platforms   []string          `yaml:"platforms"`
	CGO         bool              `yaml:"cgo"`
	Parallel    int               `yaml:"parallel"`
	Env         map[string]string `yaml:"env"`
	BeforeBuild []string          `yaml:"before_build"`
	AfterBuild  []string          `yaml:"after_build"`
}

// TestYamlConfig contains test configuration
type TestYamlConfig struct {
	Timeout    string   `yaml:"timeout"`
	Parallel   bool     `yaml:"parallel"`
	Race       bool     `yaml:"race"`
	Cover      bool     `yaml:"cover"`
	CoverMode  string   `yaml:"cover_mode"`
	CoverPkg   []string `yaml:"cover_pkg"`
	Tags       []string `yaml:"tags"`
	Verbose    bool     `yaml:"verbose"`
	Short      bool     `yaml:"short"`
	SkipFuzz   bool     `yaml:"skip_fuzz"`
	Benchmarks bool     `yaml:"benchmarks"`
}

// LintYamlConfig contains lint configuration
type LintYamlConfig struct {
	Enabled    bool              `yaml:"enabled"`
	ConfigFile string            `yaml:"config_file"`
	Timeout    string            `yaml:"timeout"`
	Parallel   int               `yaml:"parallel"`
	Format     string            `yaml:"format"`
	Exclude    []string          `yaml:"exclude"`
	Enable     []string          `yaml:"enable"`
	Disable    []string          `yaml:"disable"`
	Settings   map[string]string `yaml:"settings"`
}

// ReleaseYamlConfig contains release configuration
type ReleaseYamlConfig struct {
	Enabled       bool                `yaml:"enabled"`
	Channel       string              `yaml:"channel"`
	Prerelease    bool                `yaml:"prerelease"`
	Draft         bool                `yaml:"draft"`
	Assets        []string            `yaml:"assets"`
	Platforms     []string            `yaml:"platforms"`
	BeforeRelease []string            `yaml:"before_release"`
	AfterRelease  []string            `yaml:"after_release"`
	Changelog     ChangelogConfig     `yaml:"changelog"`
	GitHub        GitHubReleaseConfig `yaml:"github"`
}

// ChangelogConfig contains changelog configuration
type ChangelogConfig struct {
	Enabled bool   `yaml:"enabled"`
	Format  string `yaml:"format"`
	File    string `yaml:"file"`
}

// GitHubReleaseConfig contains GitHub release configuration
type GitHubReleaseConfig struct {
	Owner      string `yaml:"owner"`
	Repository string `yaml:"repository"`
	Token      string `yaml:"token"`
}

// CIYamlConfig contains CI/CD configuration
type CIYamlConfig struct {
	Enabled   bool     `yaml:"enabled"`
	Provider  string   `yaml:"provider"`
	Workflows []string `yaml:"workflows"`
	Matrix    CIMatrix `yaml:"matrix"`
}

// CIMatrix contains CI matrix configuration
type CIMatrix struct {
	GoVersions []string `yaml:"go_versions"`
	Platforms  []string `yaml:"platforms"`
}

// Init creates a new mage.yaml configuration file
func (Yaml) Init() error {
	utils.Header("📝 Initializing mage.yaml Configuration")

	// Check if mage.yaml already exists
	if utils.FileExists("mage.yaml") {
		return errMageYamlExists
	}

	// Create default configuration
	config := createDefaultConfig()

	// Try to populate from existing project
	populateFromProject(config)

	// Write configuration
	if err := writeConfig(config, "mage.yaml"); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	utils.Success("Created mage.yaml configuration")
	utils.Info("Edit mage.yaml to customize your project settings")

	return nil
}

// Validate validates the mage.yaml configuration
func (Yaml) Validate() error {
	utils.Header("✅ Validating mage.yaml Configuration")

	config, err := loadConfig("mage.yaml")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("configuration validation failed: %w", err)
	}

	utils.Success("Configuration is valid")
	return nil
}

// Show displays the current configuration
func (Yaml) Show() error {
	utils.Header("📋 Current mage.yaml Configuration")

	config, err := loadConfig("mage.yaml")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Display configuration
	displayConfig(config)

	return nil
}

// Update updates the mage.yaml configuration
func (Yaml) Update() error {
	utils.Header("🔄 Updating mage.yaml Configuration")

	// Load existing configuration
	config, err := loadConfig("mage.yaml")
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Update from environment variables
	updateFromEnv(config)

	// Update from project changes
	populateFromProject(config)

	// Write updated configuration
	if err := writeConfig(config, "mage.yaml"); err != nil {
		return fmt.Errorf("failed to write configuration: %w", err)
	}

	utils.Success("Updated mage.yaml configuration")
	return nil
}

// Template creates a configuration template for different project types
func (Yaml) Template() error {
	utils.Header("📄 Creating Configuration Template")

	projectType := env.GetString("MAGE_X_PROJECT_TYPE", "library")

	var config *YamlConfig

	switch projectType {
	case "library":
		config = createLibraryTemplate()
	case "cli":
		config = createCLITemplate()
	case "webapi":
		config = createWebAPITemplate()
	case "microservice":
		config = createMicroserviceTemplate()
	case "tool":
		config = createToolTemplate()
	default:
		config = createDefaultConfig()
	}

	// Write template
	filename := fmt.Sprintf("mage.%s.yaml", projectType)
	if err := writeConfig(config, filename); err != nil {
		return fmt.Errorf("failed to write template: %w", err)
	}

	utils.Success("Created configuration template: %s", filename)
	utils.Info("Copy to mage.yaml and customize for your project")

	return nil
}

// Helper functions

// createDefaultConfig creates a default configuration
func createDefaultConfig() *YamlConfig {
	return &YamlConfig{
		Version: "1.0",
		Project: ProjectYamlConfig{
			Name:        "my-project",
			Description: "A Go project built with MAGE-X",
			Module:      "github.com/username/my-project",
			MainPkg:     ".",
			BinaryName:  "my-project",
			Version:     "v1.0.0",
			Authors:     []string{"Your Name <your.email@example.com>"},
			License:     "MIT",
		},
		Build: BuildYamlConfig{
			Dir:       ".",
			OutputDir: "bin",
			LDFlags:   "-s -w",
			Tags:      []string{},
			Platforms: []string{"linux/amd64", "darwin/amd64", "windows/amd64"},
			CGO:       false,
			Parallel:  4,
			Env:       map[string]string{},
		},
		Test: TestYamlConfig{
			Timeout:   "10m",
			Parallel:  true,
			Race:      false,
			Cover:     true,
			CoverMode: "atomic",
			CoverPkg:  []string{},
			Tags:      []string{},
			Verbose:   false,
			Short:     false,
			SkipFuzz:  false,
		},
		Lint: LintYamlConfig{
			Enabled:    true,
			ConfigFile: ".golangci.json",
			Timeout:    "5m",
			Parallel:   4,
			Format:     "colored-line-number",
			Exclude:    []string{},
			Enable:     []string{},
			Disable:    []string{},
			Settings:   map[string]string{},
		},
		Release: ReleaseYamlConfig{
			Enabled:    true,
			Channel:    "stable",
			Prerelease: false,
			Draft:      false,
			Assets:     []string{},
			Platforms:  []string{"linux/amd64", "darwin/amd64", "windows/amd64"},
			Changelog: ChangelogConfig{
				Enabled: true,
				Format:  "markdown",
				File:    "CHANGELOG.md",
			},
			GitHub: GitHubReleaseConfig{
				Owner:      "username",
				Repository: "repository",
				Token:      "GITHUB_TOKEN",
			},
		},
		CI: CIYamlConfig{
			Enabled:   true,
			Provider:  "github",
			Workflows: []string{"ci.yml"},
			Matrix: CIMatrix{
				GoVersions: []string{GetDefaultGoVersion(), GetSecondaryGoVersion()},
				Platforms:  []string{"ubuntu-latest", "macos-latest", "windows-latest"},
			},
		},
	}
}

// createLibraryTemplate creates a library project template
func createLibraryTemplate() *YamlConfig {
	config := createDefaultConfig()
	config.Project.Description = "A Go library built with MAGE-X"
	config.Build.Platforms = []string{"linux/amd64", "darwin/amd64", "windows/amd64"}
	config.Test.Benchmarks = true
	return config
}

// createCLITemplate creates a CLI project template
func createCLITemplate() *YamlConfig {
	config := createDefaultConfig()
	config.Project.Description = "A CLI application built with MAGE-X"
	config.Build.Platforms = []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64", "windows/amd64"}
	config.Release.Assets = []string{"completions/*", "docs/*"}
	return config
}

// createWebAPITemplate creates a web API project template
func createWebAPITemplate() *YamlConfig {
	config := createDefaultConfig()
	config.Project.Description = "A web API built with MAGE-X"
	config.Build.Tags = []string{"netgo"}
	config.Test.Tags = []string{"integration"}
	return config
}

// createMicroserviceTemplate creates a microservice project template
func createMicroserviceTemplate() *YamlConfig {
	config := createDefaultConfig()
	config.Project.Description = "A microservice built with MAGE-X"
	config.Build.Tags = []string{"netgo"}
	config.Test.Tags = []string{"integration", "e2e"}
	return config
}

// createToolTemplate creates a tool project template
func createToolTemplate() *YamlConfig {
	config := createDefaultConfig()
	config.Project.Description = "A developer tool built with MAGE-X"
	config.Build.Platforms = []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64", "windows/amd64"}
	config.Release.Assets = []string{"LICENSE", "README.md"}
	return config
}

// populateFromProject populates configuration from existing project
func populateFromProject(config *YamlConfig) {
	// Get module name
	if module, err := getModuleName(); err == nil {
		config.Project.Module = module

		// Extract project name from module
		parts := strings.Split(module, "/")
		if len(parts) > 0 {
			config.Project.Name = parts[len(parts)-1]
			config.Project.BinaryName = parts[len(parts)-1]
		}

		// Extract GitHub info
		if len(parts) >= 3 && parts[0] == "github.com" {
			config.Release.GitHub.Owner = parts[1]
			config.Release.GitHub.Repository = parts[2]
		}
	}

	// Get version
	if projectVersion := getVersion(); projectVersion != versionDev {
		config.Project.Version = projectVersion
	}

	// Check for existing files
	if utils.FileExists("README.md") {
		// Could parse README for description
		// README parsing for project description is reserved for future implementation.
		// When implemented, this will extract project description from README content.
		_ = "README.md" // Future feature placeholder
	}

	if utils.FileExists("LICENSE") {
		// Could parse LICENSE for license type
		// LICENSE parsing for license detection is reserved for future implementation.
		// When implemented, this will detect license type from LICENSE file content.
		_ = "LICENSE" // Future feature placeholder
	}

	if utils.FileExists(".github/workflows") {
		config.CI.Enabled = true
	}
}

// updateFromEnv updates configuration from environment variables
func updateFromEnv(config *YamlConfig) {
	// Update from environment variables
	if name := env.GetString("MAGE_X_PROJECT_NAME", ""); name != "" {
		config.Project.Name = name
	}

	if desc := env.GetString("MAGE_X_PROJECT_DESCRIPTION", ""); desc != "" {
		config.Project.Description = desc
	}

	if envVersion := env.GetString("MAGE_X_PROJECT_VERSION", ""); envVersion != "" {
		config.Project.Version = envVersion
	}

	if license := env.GetString("MAGE_X_PROJECT_LICENSE", ""); license != "" {
		config.Project.License = license
	}

	// Build configuration
	if ldflags := env.GetString("MAGE_X_BUILD_LDFLAGS", ""); ldflags != "" {
		config.Build.LDFlags = ldflags
	}

	if platforms := env.GetString("MAGE_X_BUILD_PLATFORMS", ""); platforms != "" {
		config.Build.Platforms = strings.Split(platforms, ",")
	}

	// Test configuration
	if timeout := env.GetString("MAGE_X_TEST_TIMEOUT", ""); timeout != "" {
		config.Test.Timeout = timeout
	}

	config.Test.Verbose = env.GetBool("TEST_VERBOSE", config.Test.Verbose)
	config.Test.Race = env.GetBool("TEST_RACE", config.Test.Race)
	config.Test.Cover = env.GetBool("TEST_COVER", config.Test.Cover)
}

// loadConfig loads configuration from file
func loadConfig(filename string) (*YamlConfig, error) {
	fileOps := fileops.New()
	var config YamlConfig
	if err := fileOps.YAML.ReadYAML(filename, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// writeConfig writes configuration to file
func writeConfig(config *YamlConfig, filename string) error {
	fileOps := fileops.New()
	return fileOps.SaveConfig(filename, config, "yaml")
}

// validateConfig validates the configuration
func validateConfig(config *YamlConfig) error {
	if config.Project.Name == "" {
		return errProjectNameRequired
	}

	if config.Project.Module == "" {
		return errProjectModuleRequired
	}

	if config.Build.OutputDir == "" {
		return errBuildOutputRequired
	}

	// Validate platforms
	for _, platform := range config.Build.Platforms {
		parts := strings.Split(platform, "/")
		if len(parts) != 2 {
			return fmt.Errorf("%w: %s", errYamlInvalidPlatformFormat, platform)
		}
	}

	return nil
}

// displayConfig displays the configuration
func displayConfig(config *YamlConfig) {
	fmt.Printf("📦 Project: %s\n", config.Project.Name)
	fmt.Printf("📝 Description: %s\n", config.Project.Description)
	fmt.Printf("🏗️  Module: %s\n", config.Project.Module)
	fmt.Printf("📊 Version: %s\n", config.Project.Version)
	fmt.Printf("📄 License: %s\n", config.Project.License)

	fmt.Printf("\n🔨 Build Configuration:\n")
	fmt.Printf("  Output Directory: %s\n", config.Build.OutputDir)
	fmt.Printf("  Platforms: %s\n", strings.Join(config.Build.Platforms, ", "))
	fmt.Printf("  CGO Enabled: %t\n", config.Build.CGO)
	fmt.Printf("  Parallel Jobs: %d\n", config.Build.Parallel)

	fmt.Printf("\n🧪 Test Configuration:\n")
	fmt.Printf("  Timeout: %s\n", config.Test.Timeout)
	fmt.Printf("  Parallel: %t\n", config.Test.Parallel)
	fmt.Printf("  Race Detection: %t\n", config.Test.Race)
	fmt.Printf("  Coverage: %t\n", config.Test.Cover)

	fmt.Printf("\n🔍 Lint Configuration:\n")
	fmt.Printf("  Enabled: %t\n", config.Lint.Enabled)
	fmt.Printf("  Config File: %s\n", config.Lint.ConfigFile)
	fmt.Printf("  Timeout: %s\n", config.Lint.Timeout)

	fmt.Printf("\n🚀 Release Configuration:\n")
	fmt.Printf("  Enabled: %t\n", config.Release.Enabled)
	fmt.Printf("  Channel: %s\n", config.Release.Channel)
	fmt.Printf("  Prerelease: %t\n", config.Release.Prerelease)

	fmt.Printf("\n⚙️  CI Configuration:\n")
	fmt.Printf("  Enabled: %t\n", config.CI.Enabled)
	fmt.Printf("  Provider: %s\n", config.CI.Provider)
	fmt.Printf("  Go Versions: %s\n", strings.Join(config.CI.Matrix.GoVersions, ", "))
}
