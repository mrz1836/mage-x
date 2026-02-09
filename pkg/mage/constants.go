package mage

import "os"

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

// Boolean string values for environment variable comparisons
const (
	trueValue  = "true"
	falseValue = "false"
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
	FileMageYAMLAlt  = "mage.yaml"
	FileMageYMLAlt   = "mage.yml"
	FileGitignore    = ".gitignore"
	FileVersion      = "VERSION"
	FileCoverageOut  = "coverage.out"
	FileCoverageHTML = "coverage.html"

	// Goreleaser configuration files
	FileGoreleaserYML     = ".goreleaser.yml"
	FileGoreleaserYAML    = ".goreleaser.yaml"
	FileGoreleaserYMLAlt  = "goreleaser.yml"
	FileGoreleaserYAMLAlt = "goreleaser.yaml"

	// Golangci-lint configuration files
	FileGolangciJSON    = ".golangci.json"
	FileGolangciYML     = ".golangci.yml"
	FileGolangciYAML    = ".golangci.yaml"
	FileGolangciYMLAlt  = "golangci.yml"
	FileGolangciYAMLAlt = "golangci.yaml"

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
	EnvGoVersion   = "GO_VERSION"
	EnvCGOEnabled  = "CGO_ENABLED"
	EnvPath        = "PATH"
	EnvDebug       = "DEBUG"
	EnvVerbose     = "VERBOSE"
	EnvDryRun      = "DRY_RUN"
	EnvNoColor     = "NO_COLOR"
	EnvForceColor  = "FORCE_COLOR"
	EnvCI          = "CI"
	EnvGitHubToken = "GITHUB_TOKEN" //nolint:gosec // G101: This is an environment variable name, not a hardcoded credential

	// MAGE-specific environment variables (legacy, without prefix)
	EnvMageVerbose      = "MAGE_VERBOSE"
	EnvMageAuditEnabled = "MAGE_AUDIT_ENABLED"
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

// CI platform identifiers (used by CIDetector.Platform())
const (
	CIPlatformGitHub    = "github"
	CIPlatformGitLab    = "gitlab"
	CIPlatformCircleCI  = "circleci"
	CIPlatformTravis    = "travis"
	CIPlatformJenkins   = "jenkins"
	CIPlatformAzure     = "azure"
	CIPlatformBuildkite = "buildkite"
	CIPlatformDrone     = "drone"
	CIPlatformCodeBuild = "codebuild"
	CIPlatformTeamCity  = "teamcity"
	CIPlatformBitbucket = "bitbucket"
	CIPlatformGeneric   = "generic"
	CIPlatformLocal     = "local"
)

// Time and duration constants
const (
	DefaultTimeout   = "10m"
	DefaultBenchTime = "10s"
	ShortTimeout     = "1m"
	LongTimeout      = "30m"

	// Fuzz test timing constants
	// DefaultFuzzBaselineOverheadPerSeed is the estimated time per seed during baseline gathering.
	// Go's fuzz tests run all seeds before the -fuzztime timer starts.
	DefaultFuzzBaselineOverheadPerSeed = "500ms"
	// DefaultFuzzBaselineBuffer is extra buffer time added to fuzz test timeouts
	// to account for test setup, teardown, compilation overhead, and variance in seed execution time.
	// Increased from 1m to 90s to account for fuzz test binary compilation on CI systems.
	DefaultFuzzBaselineBuffer = "90s"
	// DefaultFuzzMinTimeout is the minimum timeout for fuzz tests.
	// Increased from 1m to 90s to account for compilation overhead.
	DefaultFuzzMinTimeout = "90s"
	// DefaultFuzzWarmupTimeout is the timeout for pre-compiling fuzz test binaries
	// with coverage instrumentation to warm the Go build cache before running individual tests.
	DefaultFuzzWarmupTimeout = "5m"
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

// Configuration file list functions
// These return slices of configuration file names in search order.
// Dot-prefixed files take precedence over non-prefixed ones.

// MageConfigFiles returns mage configuration file names in search order.
// The order is important: dot-prefixed files take precedence.
func MageConfigFiles() []string {
	return []string{FileMageYAML, FileMageYML, FileMageYAMLAlt, FileMageYMLAlt}
}

// GoreleaserConfigFiles returns goreleaser configuration file names in search order.
func GoreleaserConfigFiles() []string {
	return []string{FileGoreleaserYML, FileGoreleaserYAML, FileGoreleaserYMLAlt, FileGoreleaserYAMLAlt}
}

// GolangciLintConfigFiles returns golangci-lint configuration file names in search order.
func GolangciLintConfigFiles() []string {
	return []string{FileGolangciJSON, FileGolangciYML, FileGolangciYAML, FileGolangciYMLAlt, FileGolangciYAMLAlt}
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
	DefaultBmadVersionTag  = ""
	DefaultBmadPackageName = "bmad-method"
)

// Agent OS command names
const (
	CmdCurl = "curl"
	CmdBash = "bash"
)

// Agent OS default configuration values
const (
	DefaultAgentOSBaseDir    = "agent-os"
	DefaultAgentOSHomeDir    = "agent-os" // Relative to home directory
	DefaultAgentOSConfigFile = "config.yml"
	DefaultAgentOSGitHubRepo = "buildermethods/agent-os"
	DefaultAgentOSBranch     = "main"
	DefaultAgentOSSpecsDir   = "agent-os/specs"
	DefaultAgentOSProductDir = "agent-os/product"
)
