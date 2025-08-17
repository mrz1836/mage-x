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

// Release channels
const (
	ChannelStable  = "stable"
	ChannelBeta    = "beta"
	ChannelEdge    = "edge"
	ChannelNightly = "nightly"
)

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

	// Provide fallback defaults for tools (matching test expectations)
	var fallback string
	switch toolName {
	case "golangci-lint":
		fallback = "v2.4.0" // Match .env.base current version
	case "gofumpt":
		fallback = VersionLatest // Match test expectation
	case "govulncheck":
		fallback = "v1.1.4" // Match specific test expectation
	case "mockgen":
		fallback = VersionLatest // Default fallback
	case "swag":
		fallback = VersionLatest // Default fallback
	case "go":
		fallback = "1.24"
	default:
		fallback = VersionLatest
	}

	// Warn if not found but provide fallback
	utils.Warn("Tool version for %s not found in environment variables (%s)", toolName, envVar)
	utils.Warn("Consider sourcing .github/.env.base: source .github/.env.base")
	utils.Warn("Using fallback version: %s", fallback)
	return fallback
}

// Tool version getters (environment-aware functions, not globals)
func GetDefaultGolangciLintVersion() string {
	return getToolVersionOrWarn("MAGE_X_GOLANGCI_LINT_VERSION", "GOLANGCI_LINT_VERSION", "golangci-lint")
}

func GetDefaultGofumptVersion() string {
	return getToolVersionOrWarn("MAGE_X_GOFUMPT_VERSION", "GOFUMPT_VERSION", "gofumpt")
}

func GetDefaultGoVulnCheckVersion() string {
	return getToolVersionOrWarn("MAGE_X_GOVULNCHECK_VERSION", "GOVULNCHECK_VERSION", "govulncheck")
}

func GetDefaultMockgenVersion() string {
	return getToolVersionOrWarn("MAGE_X_MOCKGEN_VERSION", "MOCKGEN_VERSION", "mockgen")
}

func GetDefaultSwagVersion() string {
	return getToolVersionOrWarn("MAGE_X_SWAG_VERSION", "SWAG_VERSION", "swag")
}

func GetDefaultGoVersion() string {
	version := getToolVersionOrWarn("MAGE_X_GO_VERSION", "GO_PRIMARY_VERSION", "go")
	// Clean up the version to remove any .x suffix for actual usage
	if version != "" && len(version) > 2 && version[len(version)-2:] == ".x" {
		return version[:len(version)-2]
	}
	return version
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
