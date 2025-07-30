package mage

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
	EnvGitHubToken = "GITHUB_TOKEN"
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

// Tool versions (defaults)
const (
	DefaultGolangciLintVersion = "v2.3.0"
	DefaultGofumptVersion      = "v0.5.0"
	DefaultGoVulnCheckVersion  = "latest"
	DefaultMockgenVersion      = "v1.6.0"
	DefaultSwagVersion         = "v1.16.2"
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
