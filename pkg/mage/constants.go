package mage

import (
	"os"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// Command names
const (
	// Go toolchain commands
	CmdGo         = "go"
	CmdGoBuild    = "build"
	CmdGoTest     = "test"
	CmdGoMod      = "mod"
	CmdGoGenerate = "generate"
	CmdGoInstall  = "install"
	CmdGoGet      = "get"
	CmdGoList     = "list"
	CmdGoVet      = "vet"

	// External tools
	CmdGit          = "git"
	CmdDocker       = "docker"
	CmdGolangciLint = "golangci-lint"
	CmdGofumpt      = "gofumpt"
	CmdYamlfmt      = "yamlfmt"
	LintTool        = "golangci-lint" // Default lint tool for compatibility
	CmdGoVulnCheck  = "govulncheck"
	CmdMockgen      = "mockgen"
	CmdSwag         = "swag"

	// Shell commands
	CmdFind  = "find"
	CmdWC    = "wc"
	CmdRM    = "rm"
	CmdMkdir = "mkdir"
)

// Common arguments and flags
const (
	// Go build flags
	FlagOutput    = "-o"
	FlagVerbose   = "-v"
	FlagTags      = "-tags"
	FlagLDFlags   = "-ldflags"
	FlagTrimPath  = "-trimpath"
	FlagRace      = "-race"
	FlagCover     = "-cover"
	FlagCoverMode = "-covermode"
	FlagCoverPkg  = "-coverpkg"
	FlagTimeout   = "-timeout"
	FlagShort     = "-short"
	FlagParallel  = "-parallel"
	FlagBench     = "-bench"
	FlagRun       = "-run"
	FlagCount     = "-count"
	FlagCPU       = "-cpu"

	// Git flags
	FlagTags2    = "--tags"
	FlagAbbrev   = "--abbrev=0"
	FlagNoVerify = "--no-verify"

	// Common values
	ArgAll     = "./..."
	ArgCurrent = "."
	ArgNone    = "none"
)

// File and directory names
const (
	FileGoMod        = "go.mod"
	FileGoSum        = "go.sum"
	FileMageYAML     = ".mage.yaml"
	FileMageYML      = ".mage.yml"
	FileGitignore    = ".gitignore"
	FileVersion      = "VERSION"
	FileCoverageOut  = "coverage.out"
	FileCoverageHTML = "coverage.html"

	DirBin      = "bin"
	DirBuild    = "build"
	DirDist     = "dist"
	DirVendor   = "vendor"
	DirTestdata = "testdata"
	DirCmd      = "cmd"
	DirPkg      = "pkg"
	DirInternal = "internal"
)

// Environment variables
const (
	// MAGE-X specific environment variable prefix
	EnvPrefix = "MAGE_X_"

	// System environment variables (do not prefix these)
	EnvGOOS        = "GOOS"
	EnvGOARCH      = "GOARCH"
	EnvGOPATH      = "GOPATH"
	EnvGOBIN       = "GOBIN"
	EnvCGOEnabled  = "CGO_ENABLED"
	EnvPath        = "PATH"
	EnvDebug       = "DEBUG"
	EnvVerbose     = "VERBOSE"
	EnvDryRun      = "DRY_RUN"
	EnvNoColor     = "NO_COLOR"
	EnvForceColor  = "FORCE_COLOR"
	EnvCI          = "CI"
	EnvGitHubToken = "GITHUB_TOKEN" //nolint:gosec // G101: This is an environment variable name, not a hardcoded credential
)

// Platform constants
const (
	OSLinux   = "linux"
	OSDarwin  = "darwin"
	OSWindows = "windows"

	ArchAMD64 = "amd64"
	ArchARM64 = "arm64"
	Arch386   = "386"
	ArchARM   = "arm"
)

// Time and duration constants
const (
	DefaultTimeout   = "10m"
	DefaultBenchTime = "10s"
	ShortTimeout     = "1m"
	LongTimeout      = "30m"
)

// Coverage modes
const (
	CoverModeSet    = "set"
	CoverModeCount  = "count"
	CoverModeAtomic = "atomic"
)

// YAML formatting constants
const (
	// MaxYAMLLineLength is the maximum safe line length for YAML files
	// Set to 60KB to stay safely under bufio.Scanner's 64KB token limit
	MaxYAMLLineLength = 60000

	// EnvYAMLValidation is the environment variable to disable YAML validation
	// Set to "false" to skip pre-validation checks
	EnvYAMLValidation = "MAGE_X_YAML_VALIDATION"
)

// GetMageXEnv returns the value of a MAGE-X environment variable with the proper prefix
func GetMageXEnv(suffix string) string {
	return os.Getenv(EnvPrefix + suffix)
}

// getToolVersionOrWarn returns tool version from environment or warns if not found
func getToolVersionOrWarn(envVar, legacyEnvVar, toolName string) string {
	// Check primary environment variable
	if value := os.Getenv(envVar); value != "" {
		return value
	}

	// Check legacy environment variable for backward compatibility
	if legacyEnvVar != "" {
		if value := os.Getenv(legacyEnvVar); value != "" {
			return value
		}
	}

	// Provide fallback defaults for tools - all use latest to avoid hardcoded versions
	var fallback string
	switch toolName {
	case CmdGolangciLint:
		fallback = VersionLatest // Use latest if env vars not loaded
	case CmdGofumpt:
		fallback = VersionLatest
	case CmdYamlfmt:
		fallback = VersionLatest // Use latest if env vars not loaded
	case CmdGoVulnCheck:
		fallback = VersionLatest // Use latest if env vars not loaded
	case CmdMockgen:
		fallback = VersionLatest
	case CmdSwag:
		fallback = VersionLatest
	case "go":
		fallback = VersionLatest // Use latest if env vars not loaded
	default:
		fallback = VersionLatest
	}

	// Warn if not found but provide fallback
	utils.Warn("Tool version for %s not found in environment variables (%s)", toolName, envVar)
	utils.Warn("Consider sourcing .github/.env.base: source .github/.env.base")
	utils.Warn("Using fallback version: %s", fallback)
	return fallback
}

// GetDefaultGolangciLintVersion returns the default golangci-lint version from env or fallback
func GetDefaultGolangciLintVersion() string { return getToolVersionFromRegistry(CmdGolangciLint) }

// GetDefaultGofumptVersion returns the default gofumpt version from env or fallback
func GetDefaultGofumptVersion() string { return getToolVersionFromRegistry(CmdGofumpt) }

// GetDefaultYamlfmtVersion returns the default yamlfmt version from env or fallback
func GetDefaultYamlfmtVersion() string { return getToolVersionFromRegistry(CmdYamlfmt) }

// GetDefaultGoVulnCheckVersion returns the default govulncheck version from env or fallback
func GetDefaultGoVulnCheckVersion() string { return getToolVersionFromRegistry(CmdGoVulnCheck) }

// GetDefaultMockgenVersion returns the default mockgen version from env or fallback
func GetDefaultMockgenVersion() string { return getToolVersionFromRegistry(CmdMockgen) }

// GetDefaultSwagVersion returns the default swag version from env or fallback
func GetDefaultSwagVersion() string { return getToolVersionFromRegistry(CmdSwag) }

// GetDefaultGoVersion returns the default Go version from env or fallback
func GetDefaultGoVersion() string {
	goVersion := getToolVersionOrWarn("MAGE_X_GO_VERSION", "GO_PRIMARY_VERSION", "go")
	// Clean up the version to remove any .x suffix for actual usage
	if goVersion != "" && len(goVersion) > 2 && goVersion[len(goVersion)-2:] == ".x" {
		return goVersion[:len(goVersion)-2]
	}
	return goVersion
}

// GetSecondaryGoVersion returns the secondary Go version from env or fallback
func GetSecondaryGoVersion() string {
	secondaryVersion := getToolVersionOrWarn("MAGE_X_GO_SECONDARY_VERSION", "GO_SECONDARY_VERSION", "go")
	// Clean up the version to remove any .x suffix for actual usage
	if secondaryVersion != "" && len(secondaryVersion) > 2 && secondaryVersion[len(secondaryVersion)-2:] == ".x" {
		return secondaryVersion[:len(secondaryVersion)-2]
	}
	return secondaryVersion
}

// Version constants for consistency
const (
	VersionLatest   = "latest"
	VersionAtLatest = "@latest"
)

// toolVersionConfig holds configuration for tool version lookup
type toolVersionConfig struct {
	mageXEnvVar  string // Primary env var (e.g., "MAGE_X_GOLANGCI_LINT_VERSION")
	legacyEnvVar string // Legacy env var (e.g., "GOLANGCI_LINT_VERSION")
	toolName     string // Tool command name (e.g., "golangci-lint")
}

// toolVersionRegistry maps tool names to their version configuration
//
//nolint:gochecknoglobals // Required for package-level configuration registry
var toolVersionRegistry = map[string]toolVersionConfig{
	CmdGolangciLint: {
		mageXEnvVar:  "MAGE_X_GOLANGCI_LINT_VERSION",
		legacyEnvVar: "GOLANGCI_LINT_VERSION",
		toolName:     CmdGolangciLint,
	},
	CmdGofumpt: {
		mageXEnvVar:  "MAGE_X_GOFUMPT_VERSION",
		legacyEnvVar: "GOFUMPT_VERSION",
		toolName:     CmdGofumpt,
	},
	CmdYamlfmt: {
		mageXEnvVar:  "MAGE_X_YAMLFMT_VERSION",
		legacyEnvVar: "YAMLFMT_VERSION",
		toolName:     CmdYamlfmt,
	},
	CmdGoVulnCheck: {
		mageXEnvVar:  "MAGE_X_GOVULNCHECK_VERSION",
		legacyEnvVar: "GOVULNCHECK_VERSION",
		toolName:     CmdGoVulnCheck,
	},
	CmdMockgen: {
		mageXEnvVar:  "MAGE_X_MOCKGEN_VERSION",
		legacyEnvVar: "MOCKGEN_VERSION",
		toolName:     CmdMockgen,
	},
	CmdSwag: {
		mageXEnvVar:  "MAGE_X_SWAG_VERSION",
		legacyEnvVar: "SWAG_VERSION",
		toolName:     CmdSwag,
	},
}

// getToolVersionFromRegistry returns the version for a tool using the registry.
// This is the internal unified function used by GetDefault* functions.
func getToolVersionFromRegistry(toolName string) string {
	cfg, ok := toolVersionRegistry[toolName]
	if !ok {
		utils.Warn("Unknown tool: %s, using 'latest'", toolName)
		return VersionLatest
	}
	return getToolVersionOrWarn(cfg.mageXEnvVar, cfg.legacyEnvVar, cfg.toolName)
}

// Error messages
const (
	ErrNoGoMod      = "no go.mod file found"
	ErrNoGitRepo    = "not a git repository"
	ErrNoVersion    = "no version information available"
	ErrBuildFailed  = "build failed"
	ErrTestFailed   = "tests failed"
	ErrLintFailed   = "linting failed"
	ErrToolNotFound = "tool not found"
)

// Success messages
const (
	MsgBuildSuccess   = "Build completed successfully"
	MsgTestSuccess    = "All tests passed"
	MsgLintSuccess    = "No linting issues found"
	MsgInstallSuccess = "Installation completed successfully"
	MsgCleanSuccess   = "Clean completed successfully"
)

// Info messages
const (
	MsgBuildingApp  = "Building application"
	MsgRunningTests = "Running tests"
	MsgRunningLint  = "Running linter"
	MsgInstalling   = "Installing"
	MsgCleaning     = "Cleaning build artifacts"
	MsgGenerating   = "Running code generation"
)

// Emoji constants for user-friendly output
const (
	EmojiBuild   = "🔨"
	EmojiTest    = "🧪"
	EmojiLint    = "🔍"
	EmojiSuccess = "✅"
	EmojiError   = "❌"
	EmojiWarning = "⚠️"
	EmojiInfo    = "ℹ️"
	EmojiRocket  = "🚀"
	EmojiPackage = "📦"
	EmojiClean   = "🧹"
	EmojiTarget  = "🎯"
	EmojiClock   = "⏱️"
	EmojiShield  = "🛡️"
	EmojiChart   = "📊"
	EmojiBook    = "📚"
	EmojiGear    = "⚙️"
	EmojiRefresh = "🔄"
)

// Format strings
const (
	FmtPlatform    = "%s/%s"
	FmtBuildTag    = "Building for %s"
	FmtTestPackage = "Testing %s"
	FmtInstallTool = "Installing %s@%s"
	FmtVersion     = "Version: %s"
	FmtDuration    = "Duration: %s"
	FmtCoverage    = "Coverage: %.1f%%"
)

// Speckit command names
const (
	CmdUV      = "uv"
	CmdUVX     = "uvx"
	CmdSpecify = "specify"
)

// Speckit default configuration values
const (
	DefaultSpeckitConstitutionPath = ".specify/memory/constitution.md"
	DefaultSpeckitVersionFile      = ".specify/version.txt"
	DefaultSpeckitBackupDir        = ".specify/backups"
	DefaultSpeckitBackupsToKeep    = 5
	DefaultSpeckitCLIName          = "specify-cli"
	DefaultSpeckitGitHubRepo       = "git+https://github.com/github/spec-kit.git"
	DefaultSpeckitAIProvider       = "claude"
)

// BMAD command names
const (
	CmdNpm  = "npm"
	CmdNpx  = "npx"
	CmdBmad = "bmad-method"
)

// BMAD default configuration values
const (
	DefaultBmadProjectDir  = "_bmad"
	DefaultBmadVersionTag  = "@alpha"
	DefaultBmadPackageName = "bmad-method"
)
