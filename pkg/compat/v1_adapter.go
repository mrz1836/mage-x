// Package compat provides backward compatibility for legacy mage configurations
package compat

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mrz1836/go-mage/pkg/common/config"
	"github.com/mrz1836/go-mage/pkg/common/env"
	"github.com/mrz1836/go-mage/pkg/common/errors"
	"github.com/mrz1836/go-mage/pkg/common/fileops"
	"github.com/mrz1836/go-mage/pkg/mage"
	"github.com/mrz1836/go-mage/pkg/utils"
)

// V1Adapter provides backward compatibility for v1 mage configurations
type V1Adapter struct {
	config config.ConfigLoader
	env    env.Environment
	files  fileops.FileOperator
	runner mage.CommandRunner

	// Legacy support flags
	legacyMode       bool
	deprecationWarns bool
	strictMode       bool
}

// NewV1Adapter creates a new v1 compatibility adapter
func NewV1Adapter() *V1Adapter {
	return &V1Adapter{
		config:           config.NewFileLoader(".mage.yaml"),
		env:              env.NewOSEnvironment(),
		files:            fileops.New().File,
		runner:           mage.GetRunner(),
		deprecationWarns: true,
		legacyMode:       detectLegacyMode(),
	}
}

// LegacyConfig represents the old configuration format
type LegacyConfig struct {
	// Old flat structure
	ProjectName    string            `yaml:"project_name"`
	ProjectVersion string            `yaml:"project_version"`
	BuildOutput    string            `yaml:"build_output"`
	BuildFlags     []string          `yaml:"build_flags"`
	TestTimeout    string            `yaml:"test_timeout"`
	TestCoverage   bool              `yaml:"test_coverage"`
	LintTools      []string          `yaml:"lint_tools"`
	DocOutput      string            `yaml:"doc_output"`
	Env            map[string]string `yaml:"env"`

	// Old nested structure (v1.x)
	Build struct {
		Output string   `yaml:"output"`
		Flags  []string `yaml:"flags"`
	} `yaml:"build"`

	Test struct {
		Timeout  string `yaml:"timeout"`
		Coverage bool   `yaml:"coverage"`
	} `yaml:"test"`
}

// LoadLegacyConfig loads and converts legacy configuration
func (a *V1Adapter) LoadLegacyConfig() (*mage.Config, error) {
	// Check for legacy config files
	legacyFiles := []string{
		".mage.yml",
		"mage.yaml",
		"mage.yml",
		"Magefile.yml",
		".magefile.yaml",
	}

	var legacyPath string
	for _, file := range legacyFiles {
		if a.files.Exists(file) {
			legacyPath = file
			break
		}
	}

	if legacyPath == "" {
		// No legacy config found, use new system
		return nil, nil
	}

	if a.deprecationWarns {
		utils.Warn("Legacy configuration file '%s' detected. Please migrate to '.mage.yaml' format.", legacyPath)
		utils.Info("Run 'mage migrate:config' to automatically convert your configuration.")
	}

	// Load legacy config
	legacy := &LegacyConfig{}
	if err := a.loadYAML(legacyPath, legacy); err != nil {
		return nil, errors.Wrap(err, "failed to load legacy config")
	}

	// Convert to new format
	return a.convertLegacyConfig(legacy), nil
}

// convertLegacyConfig converts legacy config to new format
func (a *V1Adapter) convertLegacyConfig(legacy *LegacyConfig) *mage.Config {
	cfg := &mage.Config{
		Project: mage.ProjectConfig{
			Name:    a.coalesce(legacy.ProjectName, "unnamed"),
			Version: a.coalesce(legacy.ProjectVersion, "0.0.0"),
			Binary:  a.inferBinaryName(legacy.ProjectName),
		},
		Build: mage.BuildConfig{
			Output:   a.coalesce(legacy.BuildOutput, legacy.Build.Output, "dist"),
			Tags:     []string{},
			LDFlags:  a.extractLDFlags(legacy.BuildFlags),
			Verbose:  a.env.GetBool("VERBOSE", false),
			TrimPath: true,
		},
		Test: mage.TestConfig{
			Timeout:   a.coalesce(legacy.TestTimeout, legacy.Test.Timeout, "10m"),
			Cover:     legacy.TestCoverage || legacy.Test.Coverage,
			CoverMode: "atomic",
			Verbose:   a.env.GetBool("TEST_VERBOSE", false),
			Race:      true,
			Parallel:  true,
		},
		Lint: mage.LintConfig{
			GolangciVersion: "latest",
			Timeout:         "5m",
			SkipDirs:        []string{"vendor", ".git"},
			EnableAll:       false,
		},
		Docker: mage.DockerConfig{
			Registry:   a.env.GetWithDefault("DOCKER_REGISTRY", ""),
			Repository: a.inferDockerRepo(legacy.ProjectName),
			Dockerfile: "Dockerfile",
			BuildArgs:  make(map[string]string),
		},
		// Docs config removed as DocsConfig type doesn't exist in current version
		// Env field removed as it doesn't exist in current Config struct
	}

	// Add compatibility metadata
	cfg.Metadata = map[string]string{
		"migrated_from": "v1",
		"legacy_config": "true",
	}

	return cfg
}

// MigrateConfig migrates legacy configuration to new format
func (a *V1Adapter) MigrateConfig() error {
	utils.Header("Migrating Configuration")

	// Load legacy config
	cfg, err := a.LoadLegacyConfig()
	if err != nil {
		return err
	}

	if cfg == nil {
		utils.Info("No legacy configuration found.")
		return nil
	}

	// Check if new config already exists
	if a.files.Exists(".mage.yaml") {
		if !a.confirmOverwrite() {
			utils.Info("Migration canceled.")
			return nil
		}

		// Backup existing config
		backupPath := fmt.Sprintf(".mage.yaml.backup.%d", os.Getpid())
		if err := a.files.Copy(".mage.yaml", backupPath); err != nil {
			return errors.Wrap(err, "failed to backup existing config")
		}
		utils.Info("Existing config backed up to: %s", backupPath)
	}

	// Save new config
	if err := a.saveConfig(cfg); err != nil {
		return errors.Wrap(err, "failed to save migrated config")
	}

	utils.Success("Configuration migrated successfully to .mage.yaml")

	// Show migration summary
	a.showMigrationSummary(cfg)

	return nil
}

// WrapLegacyTarget wraps a legacy mage target for compatibility
func (a *V1Adapter) WrapLegacyTarget(name string, fn func() error) func() error {
	return func() error {
		if a.legacyMode && a.deprecationWarns {
			utils.Warn("Legacy target '%s' is deprecated. Please update to use new interfaces.", name)
		}

		// Set up legacy environment
		if err := a.setupLegacyEnvironment(); err != nil {
			return err
		}

		// Run legacy target
		return fn()
	}
}

// CheckLegacyTargets checks for legacy target definitions
func (a *V1Adapter) CheckLegacyTargets() []string {
	var legacyTargets []string

	// Check for old-style magefiles
	patterns := []string{
		"mage.go",
		"magefile.go",
		"Magefile.go",
		"mage_*.go",
	}

	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			if a.isLegacyMagefile(match) {
				legacyTargets = append(legacyTargets, match)
			}
		}
	}

	return legacyTargets
}

// MigrateLegacyTargets migrates legacy targets to new format
func (a *V1Adapter) MigrateLegacyTargets() error {
	utils.Header("Migrating Legacy Targets")

	targets := a.CheckLegacyTargets()
	if len(targets) == 0 {
		utils.Info("No legacy targets found.")
		return nil
	}

	utils.Info("Found %d legacy target file(s):", len(targets))
	for _, target := range targets {
		fmt.Printf("  - %s\n", target)
	}

	// Create migration plan
	plan := a.createMigrationPlan(targets)

	// Show migration plan
	a.showMigrationPlan(plan)

	if !a.confirmMigration() {
		utils.Info("Migration canceled.")
		return nil
	}

	// Execute migration
	for _, step := range plan {
		if err := a.executeMigrationStep(step); err != nil {
			return errors.Wrapf(err, "failed to migrate %s", step.Source)
		}
	}

	utils.Success("Legacy targets migrated successfully!")
	return nil
}

// Helper methods

// detectLegacyMode detects if running in legacy mode
func detectLegacyMode() bool {
	// Check for legacy indicators
	indicators := []string{
		"mage.go",
		"magefile.go",
		".mage.yml",
		"mage.yaml",
	}

	fileOps := fileops.New().File
	for _, indicator := range indicators {
		if fileOps.Exists(indicator) {
			return true
		}
	}

	// Check environment
	if os.Getenv("MAGE_LEGACY_MODE") == "true" {
		return true
	}

	return false
}

// loadYAML loads YAML configuration
func (a *V1Adapter) loadYAML(path string, v interface{}) error {
	data, err := a.files.ReadFile(path)
	if err != nil {
		return err
	}

	// Use fileops YAML support
	yamlOps := fileops.New().YAML
	return yamlOps.Unmarshal(data, v)
}

// saveConfig saves configuration in new format
func (a *V1Adapter) saveConfig(cfg *mage.Config) error {
	yamlOps := fileops.New().YAML
	return yamlOps.WriteYAML(".mage.yaml", cfg)
}

// coalesce returns the first non-empty string
func (a *V1Adapter) coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// inferBinaryName infers binary name from project name
func (a *V1Adapter) inferBinaryName(projectName string) string {
	if projectName == "" {
		return "app"
	}

	// Clean project name for binary
	binary := strings.ToLower(projectName)
	binary = strings.ReplaceAll(binary, " ", "-")
	binary = strings.ReplaceAll(binary, "_", "-")

	return binary
}

// inferDockerRepo infers Docker repository from project name
func (a *V1Adapter) inferDockerRepo(projectName string) string {
	if projectName == "" {
		return "app"
	}

	return strings.ToLower(strings.ReplaceAll(projectName, " ", "-"))
}

// mergeStringSlices merges multiple string slices
func (a *V1Adapter) mergeStringSlices(slices ...[]string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, slice := range slices {
		for _, s := range slice {
			if !seen[s] {
				seen[s] = true
				result = append(result, s)
			}
		}
	}

	return result
}

// extractLDFlags extracts ldflags from build flags
func (a *V1Adapter) extractLDFlags(flags []string) []string {
	var ldflags []string

	for i, flag := range flags {
		if flag == "-ldflags" && i+1 < len(flags) {
			// Parse ldflags value
			ldflags = append(ldflags, strings.Split(flags[i+1], " ")...)
		}
	}

	return ldflags
}

// convertLintTools converts legacy lint tools to new format (deprecated)
// LintTool type no longer exists, now using golangci-lint configuration
func (a *V1Adapter) convertLintTools(tools []string) []string {
	// Convert legacy tool names to golangci-lint linter names
	linterMap := map[string]string{
		"golint":      "golint",
		"govet":       "govet",
		"gofmt":       "gofmt",
		"ineffassign": "ineffassign",
		"misspell":    "misspell",
		"gosec":       "gosec",
		"staticcheck": "staticcheck",
	}

	var result []string
	for _, tool := range tools {
		if linter, ok := linterMap[tool]; ok {
			result = append(result, linter)
		} else {
			// Add unknown tools as-is (might be valid golangci-lint linters)
			result = append(result, tool)
		}
	}

	return result
}

// mergeEnvMaps merges environment maps
func (a *V1Adapter) mergeEnvMaps(legacy map[string]string) map[string]string {
	result := make(map[string]string)

	// Start with legacy values
	for k, v := range legacy {
		result[k] = v
	}

	// Override with current environment
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 && strings.HasPrefix(parts[0], "MAGE_") {
			result[parts[0]] = parts[1]
		}
	}

	return result
}

// setupLegacyEnvironment sets up environment for legacy targets
func (a *V1Adapter) setupLegacyEnvironment() error {
	// Load legacy config
	cfg, err := a.LoadLegacyConfig()
	if err != nil {
		return err
	}

	if cfg == nil {
		return nil
	}

	// Environment variables no longer part of Config struct
	// Legacy environment handling moved to ProjectConfig.Env

	// Set legacy compatibility flags
	a.env.Set("MAGE_LEGACY_MODE", "true")

	return nil
}

// isLegacyMagefile checks if a file contains legacy mage targets
func (a *V1Adapter) isLegacyMagefile(path string) bool {
	data, err := a.files.ReadFile(path)
	if err != nil {
		return false
	}

	content := string(data)

	// Check for legacy patterns
	legacyPatterns := []string{
		"//+build mage",
		"// +build mage",
		"package main",
		"mg.Deps(",
		"mg.SerialDeps(",
		"mg.CtxDeps(",
	}

	for _, pattern := range legacyPatterns {
		if strings.Contains(content, pattern) {
			return true
		}
	}

	return false
}

// MigrationStep represents a migration step
type MigrationStep struct {
	Source      string
	Target      string
	Type        string
	Description string
}

// createMigrationPlan creates a plan for migrating legacy targets
func (a *V1Adapter) createMigrationPlan(targets []string) []MigrationStep {
	var steps []MigrationStep

	for _, target := range targets {
		// Determine target namespace
		namespace := a.inferNamespace(target)

		steps = append(steps, MigrationStep{
			Source:      target,
			Target:      fmt.Sprintf("pkg/mage/%s.go", namespace),
			Type:        "namespace",
			Description: fmt.Sprintf("Convert to %s namespace", namespace),
		})
	}

	return steps
}

// inferNamespace infers the namespace from filename
func (a *V1Adapter) inferNamespace(filename string) string {
	base := filepath.Base(filename)
	base = strings.TrimSuffix(base, ".go")
	base = strings.TrimPrefix(base, "mage_")
	base = strings.TrimPrefix(base, "magefile_")

	// Common mappings
	mappings := map[string]string{
		"mage":     "build",
		"magefile": "build",
		"test":     "test",
		"lint":     "lint",
		"docker":   "docker",
		"docs":     "docs",
		"deps":     "deps",
		"release":  "releases",
	}

	if namespace, ok := mappings[base]; ok {
		return namespace
	}

	return base
}

// executeMigrationStep executes a single migration step
func (a *V1Adapter) executeMigrationStep(step MigrationStep) error {
	utils.Info("Migrating %s -> %s", step.Source, step.Target)

	// This would contain the actual migration logic
	// For now, we'll create a stub

	return nil
}

// UI helper methods

func (a *V1Adapter) confirmOverwrite() bool {
	fmt.Print("\n.mage.yaml already exists. Overwrite? [y/N] ")
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y"
}

func (a *V1Adapter) confirmMigration() bool {
	fmt.Print("\nProceed with migration? [y/N] ")
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y"
}

func (a *V1Adapter) showMigrationSummary(cfg *mage.Config) {
	fmt.Println("\nMigration Summary:")
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Project: %s (v%s)\n", cfg.Project.Name, cfg.Project.Version)
	fmt.Printf("Binary: %s\n", cfg.Project.Binary)
	fmt.Printf("Build Output: %s\n", cfg.Build.Output)
	fmt.Printf("Test Coverage: %v\n", cfg.Test.Cover)
	fmt.Printf("Lint EnableLinters: %d configured\n", len(cfg.Lint.EnableLinters))
	fmt.Println(strings.Repeat("-", 50))
}

func (a *V1Adapter) showMigrationPlan(steps []MigrationStep) {
	fmt.Println("\nMigration Plan:")
	fmt.Println(strings.Repeat("-", 50))
	for i, step := range steps {
		fmt.Printf("%d. %s\n", i+1, step.Description)
		fmt.Printf("   %s -> %s\n", step.Source, step.Target)
	}
	fmt.Println(strings.Repeat("-", 50))
}
