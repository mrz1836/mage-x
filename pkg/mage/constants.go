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
	FileDockerfile   = "Dockerfile"
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
func GetDefaultGolangciLintVersion() string {
	return getToolVersionOrWarn("MAGE_X_GOLANGCI_LINT_VERSION", "GOLANGCI_LINT_VERSION", CmdGolangciLint)
}

// GetDefaultGofumptVersion returns the default gofumpt version from env or fallback
func GetDefaultGofumptVersion() string {
	return getToolVersionOrWarn("MAGE_X_GOFUMPT_VERSION", "GOFUMPT_VERSION", CmdGofumpt)
}

// GetDefaultGoVulnCheckVersion returns the default govulncheck version from env or fallback
func GetDefaultGoVulnCheckVersion() string {
	return getToolVersionOrWarn("MAGE_X_GOVULNCHECK_VERSION", "GOVULNCHECK_VERSION", CmdGoVulnCheck)
}

// GetDefaultMockgenVersion returns the default mockgen version from env or fallback
func GetDefaultMockgenVersion() string {
	return getToolVersionOrWarn("MAGE_X_MOCKGEN_VERSION", "MOCKGEN_VERSION", CmdMockgen)
}

// GetDefaultSwagVersion returns the default swag version from env or fallback
func GetDefaultSwagVersion() string {
	return getToolVersionOrWarn("MAGE_X_SWAG_VERSION", "SWAG_VERSION", CmdSwag)
}

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
