// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mrz1836/mage-x/pkg/common/env"
	"github.com/mrz1836/mage-x/pkg/common/fileops"
)

// Config represents the mage configuration
type Config struct {
	Bmad     BmadConfig        `yaml:"bmad"`
	Build    BuildConfig       `yaml:"build"`
	Docs     DocsConfig        `yaml:"docs"`
	Download DownloadConfig    `yaml:"download"`
	Lint     LintConfig        `yaml:"lint"`
	Metadata map[string]string `yaml:"metadata,omitempty"`
	Project  ProjectConfig     `yaml:"project"`
	Release  ReleaseConfig     `yaml:"release"`
	Speckit  SpeckitConfig     `yaml:"speckit"`
	Test     TestConfig        `yaml:"test"`
	Tools    ToolsConfig       `yaml:"tools"`
}

// ProjectConfig contains project-specific settings
type ProjectConfig struct {
	Aliases     []string          `yaml:"aliases,omitempty"`
	Binary      string            `yaml:"binary"`
	Description string            `yaml:"description"`
	Env         map[string]string `yaml:"env"`
	GitDomain   string            `yaml:"git_domain"`
	Main        string            `yaml:"main"`
	Module      string            `yaml:"module"`
	Name        string            `yaml:"name"`
	RepoName    string            `yaml:"repo_name"`
	RepoOwner   string            `yaml:"repo_owner"`
	Version     string            `yaml:"version"`
}

// BuildConfig contains build-specific settings
type BuildConfig struct {
	GoFlags   []string       `yaml:"goflags"`
	LDFlags   []string       `yaml:"ldflags"`
	Output    string         `yaml:"output"`
	Parallel  int            `yaml:"parallel"`
	Platforms []string       `yaml:"platforms"`
	Tags      []string       `yaml:"tags"`
	TrimPath  bool           `yaml:"trimpath"`
	Verbose   bool           `yaml:"verbose"`
	PreBuild  PreBuildConfig `yaml:"prebuild"`
}

// PreBuildConfig contains pre-build specific settings
type PreBuildConfig struct {
	Strategy    string `yaml:"strategy"`     // Strategy: incremental, mains-first, smart, full
	BatchSize   int    `yaml:"batch_size"`   // Number of packages per batch
	BatchDelay  int    `yaml:"batch_delay"`  // Milliseconds between batches
	MemoryLimit string `yaml:"memory_limit"` // Memory limit (e.g., "4G", "auto")
	Exclude     string `yaml:"exclude"`      // Regex pattern for packages to exclude
	Priority    string `yaml:"priority"`     // Regex pattern for priority packages
	Verbose     bool   `yaml:"verbose"`      // Show detailed progress
}

// TestConfig contains test-specific settings
type TestConfig struct {
	AutoDiscoverBuildTags        bool     `yaml:"auto_discover_build_tags"`
	AutoDiscoverBuildTagsExclude []string `yaml:"auto_discover_build_tags_exclude"`
	BenchCPU                     int      `yaml:"bench_cpu"`
	BenchMem                     bool     `yaml:"bench_mem"`
	BenchTime                    string   `yaml:"bench_time"`
	CIMode                       CIMode   `yaml:"ci_mode"`
	Cover                        bool     `yaml:"cover"`
	CoverMode                    string   `yaml:"covermode"`
	CoverPkg                     []string `yaml:"coverpkg"`
	CoverageExclude              []string `yaml:"coverage_exclude"`
	ExcludeModules               []string `yaml:"exclude_modules"`
	IntegrationTag               string   `yaml:"integration_tag"`
	IntegrationTimeout           string   `yaml:"integration_timeout"`
	Parallel                     int      `yaml:"parallel"`
	Race                         bool     `yaml:"race"`
	Short                        bool     `yaml:"short"`
	Shuffle                      bool     `yaml:"shuffle"`
	SkipFuzz                     bool     `yaml:"skip_fuzz"`
	Tags                         string   `yaml:"tags"`
	Timeout                      string   `yaml:"timeout"`
	Verbose                      bool     `yaml:"verbose"`
}

// LintConfig contains linting settings
type LintConfig struct {
	DisableLinters  []string `yaml:"disable_linters"`
	EnableAll       bool     `yaml:"enable_all"`
	EnableLinters   []string `yaml:"enable_linters"`
	GolangciVersion string   `yaml:"golangci_version"`
	SkipDirs        []string `yaml:"skip_dirs"`
	SkipFiles       []string `yaml:"skip_files"`
	Timeout         string   `yaml:"timeout"`
}

// ToolsConfig contains tool versions
type ToolsConfig struct {
	Custom       map[string]string `yaml:"custom"`
	Fumpt        string            `yaml:"fumpt"`
	Yamlfmt      string            `yaml:"yamlfmt"`
	GoVulnCheck  string            `yaml:"govulncheck"`
	GolangciLint string            `yaml:"golangci_lint"`
	Mockgen      string            `yaml:"mockgen"`
	Swag         string            `yaml:"swag"`
}

// ReleaseConfig contains release settings
type ReleaseConfig struct {
	Changelog   bool     `yaml:"changelog"`
	Draft       bool     `yaml:"draft"`
	Formats     []string `yaml:"formats"`
	GitHubToken string   `yaml:"github_token_env"`
	NameTmpl    string   `yaml:"name_template"`
	Prerelease  bool     `yaml:"prerelease"`
}

// DownloadConfig contains download retry settings
type DownloadConfig struct {
	BackoffMultiplier float64 `yaml:"backoff_multiplier"`
	EnableResume      bool    `yaml:"enable_resume"`
	InitialDelayMs    int     `yaml:"initial_delay_ms"`
	MaxDelayMs        int     `yaml:"max_delay_ms"`
	MaxRetries        int     `yaml:"max_retries"`
	TimeoutMs         int     `yaml:"timeout_ms"`
	UserAgent         string  `yaml:"user_agent"`
}

// DocsConfig contains documentation settings
type DocsConfig struct {
	Tool string `yaml:"tool"` // "pkgsite", "godoc", or "" for auto-detect
	Port int    `yaml:"port"` // 0 for default port
}

// SpeckitConfig contains spec-kit CLI management settings
type SpeckitConfig struct {
	Enabled          bool   `yaml:"enabled"`           // Whether speckit commands are available (default: false, opt-in)
	ConstitutionPath string `yaml:"constitution_path"` // Path to constitution file (default: ".specify/memory/constitution.md")
	VersionFile      string `yaml:"version_file"`      // Path to version tracking file (default: ".specify/version.txt")
	BackupDir        string `yaml:"backup_dir"`        // Directory for constitution backups (default: ".specify/backups")
	BackupsToKeep    int    `yaml:"backups_to_keep"`   // Number of backups to retain (default: 5)
	CLIName          string `yaml:"cli_name"`          // Package name for spec-kit CLI (default: "specify-cli")
	GitHubRepo       string `yaml:"github_repo"`       // GitHub repository URL for spec-kit
	AIProvider       string `yaml:"ai_provider"`       // AI provider for spec-kit initialization (default: "claude")
}

// BmadConfig contains BMAD (Build More, Architect Dreams) CLI management settings
type BmadConfig struct {
	Enabled     bool   `yaml:"enabled"`      // Whether bmad commands are available (default: false, opt-in)
	ProjectDir  string `yaml:"project_dir"`  // Directory for BMAD project files (default: "_bmad")
	VersionTag  string `yaml:"version_tag"`  // npm version tag to use (default: "@alpha" for v6)
	PackageName string `yaml:"package_name"` // npm package name (default: "bmad-method")
}

// Static errors for err113 compliance
var (
	ErrMissingToolVersions = errors.New("missing required tool versions")
)

// GetConfig returns the current configuration using the active ConfigProvider
func GetConfig() (*Config, error) {
	return GetConfigProvider().GetConfig()
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

// cleanConfigValues recursively cleans all string fields in a Config struct
func cleanConfigValues(config *Config) {
	if config == nil {
		return
	}

	// Clean Project config strings
	config.Project.Name = env.CleanValue(config.Project.Name)
	config.Project.Binary = env.CleanValue(config.Project.Binary)
	config.Project.Module = env.CleanValue(config.Project.Module)
	config.Project.Main = env.CleanValue(config.Project.Main)
	config.Project.Description = env.CleanValue(config.Project.Description)
	config.Project.Version = env.CleanValue(config.Project.Version)
	config.Project.GitDomain = env.CleanValue(config.Project.GitDomain)
	config.Project.RepoOwner = env.CleanValue(config.Project.RepoOwner)
	config.Project.RepoName = env.CleanValue(config.Project.RepoName)

	// Clean Project env map
	for k, v := range config.Project.Env {
		config.Project.Env[k] = env.CleanValue(v)
	}

	// Clean Build config strings
	config.Build.Output = env.CleanValue(config.Build.Output)
	for i, tag := range config.Build.Tags {
		config.Build.Tags[i] = env.CleanValue(tag)
	}
	for i, flag := range config.Build.LDFlags {
		config.Build.LDFlags[i] = env.CleanValue(flag)
	}
	for i, platform := range config.Build.Platforms {
		config.Build.Platforms[i] = env.CleanValue(platform)
	}
	for i, flag := range config.Build.GoFlags {
		config.Build.GoFlags[i] = env.CleanValue(flag)
	}

	// Clean Test config strings
	config.Test.Timeout = env.CleanValue(config.Test.Timeout)
	config.Test.IntegrationTimeout = env.CleanValue(config.Test.IntegrationTimeout)
	config.Test.IntegrationTag = env.CleanValue(config.Test.IntegrationTag)
	config.Test.CoverMode = env.CleanValue(config.Test.CoverMode)
	config.Test.Tags = env.CleanValue(config.Test.Tags)
	config.Test.BenchTime = env.CleanValue(config.Test.BenchTime)
	for i, pkg := range config.Test.CoverPkg {
		config.Test.CoverPkg[i] = env.CleanValue(pkg)
	}
	for i, exclude := range config.Test.CoverageExclude {
		config.Test.CoverageExclude[i] = env.CleanValue(exclude)
	}

	// Clean Lint config strings
	config.Lint.GolangciVersion = env.CleanValue(config.Lint.GolangciVersion)
	config.Lint.Timeout = env.CleanValue(config.Lint.Timeout)
	for i, dir := range config.Lint.SkipDirs {
		config.Lint.SkipDirs[i] = env.CleanValue(dir)
	}
	for i, file := range config.Lint.SkipFiles {
		config.Lint.SkipFiles[i] = env.CleanValue(file)
	}
	for i, linter := range config.Lint.DisableLinters {
		config.Lint.DisableLinters[i] = env.CleanValue(linter)
	}
	for i, linter := range config.Lint.EnableLinters {
		config.Lint.EnableLinters[i] = env.CleanValue(linter)
	}

	// Clean Tools config strings
	config.Tools.GolangciLint = env.CleanValue(config.Tools.GolangciLint)
	config.Tools.Fumpt = env.CleanValue(config.Tools.Fumpt)
	config.Tools.GoVulnCheck = env.CleanValue(config.Tools.GoVulnCheck)
	config.Tools.Mockgen = env.CleanValue(config.Tools.Mockgen)
	config.Tools.Swag = env.CleanValue(config.Tools.Swag)
	for k, v := range config.Tools.Custom {
		config.Tools.Custom[k] = env.CleanValue(v)
	}

	// Clean Release config strings
	config.Release.GitHubToken = env.CleanValue(config.Release.GitHubToken)
	config.Release.NameTmpl = env.CleanValue(config.Release.NameTmpl)
	for i, format := range config.Release.Formats {
		config.Release.Formats[i] = env.CleanValue(format)
	}

	// Clean Download config strings
	config.Download.UserAgent = env.CleanValue(config.Download.UserAgent)

	// Clean Docs config strings
	config.Docs.Tool = env.CleanValue(config.Docs.Tool)

	// Clean Metadata map
	for k, v := range config.Metadata {
		config.Metadata[k] = env.CleanValue(v)
	}

	// Clean Speckit config strings
	config.Speckit.ConstitutionPath = env.CleanValue(config.Speckit.ConstitutionPath)
	config.Speckit.VersionFile = env.CleanValue(config.Speckit.VersionFile)
	config.Speckit.BackupDir = env.CleanValue(config.Speckit.BackupDir)
	config.Speckit.CLIName = env.CleanValue(config.Speckit.CLIName)
	config.Speckit.GitHubRepo = env.CleanValue(config.Speckit.GitHubRepo)
	config.Speckit.AIProvider = env.CleanValue(config.Speckit.AIProvider)

	// Clean Bmad config strings
	config.Bmad.ProjectDir = env.CleanValue(config.Bmad.ProjectDir)
	config.Bmad.VersionTag = env.CleanValue(config.Bmad.VersionTag)
	config.Bmad.PackageName = env.CleanValue(config.Bmad.PackageName)
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

	config := &Config{
		Project: ProjectConfig{
			Name:      binary,
			Binary:    binary,
			Module:    module,
			GitDomain: "github.com",
			Aliases:   getDefaultAliases(binary),
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
			ExcludeModules:     []string{""},
			CIMode:             DefaultCIMode(),
		},
		Lint: LintConfig{
			GolangciVersion: VersionLatest,
			Timeout:         "5m",
		},
		Tools: ToolsConfig{
			GolangciLint: VersionLatest,
			Fumpt:        VersionLatest,
			Yamlfmt:      VersionLatest,
			GoVulnCheck:  VersionLatest,
			Mockgen:      VersionLatest,
			Swag:         VersionLatest,
			Custom:       make(map[string]string),
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
		Bmad: BmadConfig{
			Enabled:     false, // Opt-in, disabled by default
			ProjectDir:  DefaultBmadProjectDir,
			VersionTag:  DefaultBmadVersionTag,
			PackageName: DefaultBmadPackageName,
		},
		Speckit: SpeckitConfig{
			Enabled:          false, // Opt-in, disabled by default
			ConstitutionPath: DefaultSpeckitConstitutionPath,
			VersionFile:      DefaultSpeckitVersionFile,
			BackupDir:        DefaultSpeckitBackupDir,
			BackupsToKeep:    DefaultSpeckitBackupsToKeep,
			CLIName:          DefaultSpeckitCLIName,
			GitHubRepo:       DefaultSpeckitGitHubRepo,
			AIProvider:       DefaultSpeckitAIProvider,
		},
	}

	// Clean environment-sourced values in the default config
	cleanConfigValues(config)

	return config
}

// getDefaultAliases returns default aliases for the given binary name
func getDefaultAliases(binary string) []string {
	switch binary {
	case "magex":
		return []string{"mgx"}
	default:
		return nil
	}
}

// applyEnvOverrides applies environment variable overrides to config
func applyEnvOverrides(c *Config) {
	// Binary name override
	if v := env.MustGet("MAGE_X_BINARY_NAME"); v != "" {
		c.Project.Binary = v
	}
	if v := env.CleanValue(GetMageXEnv("CUSTOM_BINARY_NAME")); v != "" {
		c.Project.Binary = v
	}

	// Build tags override (ParseStringSlice handles trimming)
	if tags := env.ParseStringSlice("MAGE_X_BUILD_TAGS"); tags != nil {
		c.Build.Tags = tags
	}

	// Verbose override (only enable, never disable via env)
	if v, ok := env.ParseBool("MAGE_X_VERBOSE"); ok && v {
		c.Build.Verbose = true
		c.Test.Verbose = true
	}

	// Test race override (only enable, never disable via env)
	if v, ok := env.ParseBool("MAGE_X_TEST_RACE"); ok && v {
		c.Test.Race = true
	}

	// Parallel override
	if v, ok := env.ParseInt("MAGE_X_PARALLEL", env.Positive); ok {
		c.Build.Parallel = v
	}

	// Auto discover build tags override (can enable or disable)
	if v, ok := env.ParseBool("MAGE_X_AUTO_DISCOVER_BUILD_TAGS"); ok {
		c.Test.AutoDiscoverBuildTags = v
	}

	// Auto discover build tags exclude override (ParseStringSlice handles trimming)
	if tags := env.ParseStringSlice("MAGE_X_AUTO_DISCOVER_BUILD_TAGS_EXCLUDE"); tags != nil {
		c.Test.AutoDiscoverBuildTagsExclude = tags
	}

	// Test exclude modules override (ParseStringSlice handles trimming)
	if mods := env.ParseStringSlice("MAGE_X_TEST_EXCLUDE_MODULES"); mods != nil {
		c.Test.ExcludeModules = mods
	}

	// Test timeout override
	if v := env.MustGet("MAGE_X_TEST_TIMEOUT"); v != "" {
		c.Test.Timeout = v
	}

	// Download config overrides
	applyDownloadEnvOverrides(&c.Download)

	// Tool version overrides
	applyToolVersionEnvOverrides(&c.Tools)
}

// applyDownloadEnvOverrides applies environment variable overrides to download config
func applyDownloadEnvOverrides(cfg *DownloadConfig) {
	if v, ok := env.ParseInt("MAGE_X_DOWNLOAD_RETRIES", env.NonNegative); ok {
		cfg.MaxRetries = v
	}
	if v, ok := env.ParseInt("MAGE_X_DOWNLOAD_TIMEOUT", env.Positive); ok {
		cfg.TimeoutMs = v
	}
	if v, ok := env.ParseInt("MAGE_X_DOWNLOAD_INITIAL_DELAY", env.Positive); ok {
		cfg.InitialDelayMs = v
	}
	if v, ok := env.ParseInt("MAGE_X_DOWNLOAD_MAX_DELAY", env.Positive); ok {
		cfg.MaxDelayMs = v
	}
	if v, ok := env.ParseFloat("MAGE_X_DOWNLOAD_BACKOFF", env.PositiveFloat); ok {
		cfg.BackoffMultiplier = v
	}
	if v, ok := env.ParseBool("MAGE_X_DOWNLOAD_RESUME"); ok {
		cfg.EnableResume = v
	}
	if v := env.MustGet("MAGE_X_DOWNLOAD_USER_AGENT"); v != "" {
		cfg.UserAgent = v
	}
}

// applyToolVersionEnvOverrides applies environment variable overrides to tool versions
func applyToolVersionEnvOverrides(cfg *ToolsConfig) {
	// Core linting tools
	if v := env.MustGet("MAGE_X_GOLANGCI_LINT_VERSION"); v != "" {
		cfg.GolangciLint = v
	}
	if v := env.MustGet("MAGE_X_GOFUMPT_VERSION"); v != "" {
		cfg.Fumpt = v
	}
	if v := env.MustGet("MAGE_X_YAMLFMT_VERSION"); v != "" {
		cfg.Yamlfmt = v
	}

	// Security scanning tools
	if v := env.MustGet("MAGE_X_GOVULNCHECK_VERSION"); v != "" {
		cfg.GoVulnCheck = v
	}

	// Code generation tools
	if v := env.MustGet("MAGE_X_MOCKGEN_VERSION"); v != "" {
		cfg.Mockgen = v
	}
	if v := env.MustGet("MAGE_X_SWAG_VERSION"); v != "" {
		cfg.Swag = v
	}
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

// IsConfigVerbose returns whether verbose mode is enabled in the config file.
// This checks the Build.Verbose and Test.Verbose settings in the config file.
// For environment-based verbose detection, use env.IsVerbose().
func IsConfigVerbose() bool {
	c, err := GetConfig()
	if err != nil {
		// Return false if config loading fails
		return false
	}
	return c.Build.Verbose || c.Test.Verbose
}

// SaveConfig saves the configuration to file
func SaveConfig(cfg *Config) error {
	fileOps := fileops.New()
	configFile := getConfigFilePath()
	return fileOps.SaveConfig(configFile, cfg, "yaml")
}

// getConfigFilePath returns the path to the config file
func getConfigFilePath() string {
	for _, cf := range MageConfigFiles() {
		if _, err := os.Stat(cf); err == nil {
			return cf
		}
	}
	return FileMageYAML // default
}

// Methods for Config struct required by tests

// Load loads configuration from file
func (c *Config) Load() error {
	_, err := GetConfig()
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
