package mage

import (
	"strings"
	"testing"

	"github.com/mrz1836/go-mage/pkg/mage/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ConstantsTestSuite defines the test suite for constants
type ConstantsTestSuite struct {
	suite.Suite
	env *testutil.TestEnvironment
}

// SetupTest runs before each test
func (ts *ConstantsTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
}

// TearDownTest runs after each test
func (ts *ConstantsTestSuite) TearDownTest() {
	ts.env.Cleanup()
}

// TestGoToolchainCommands tests Go toolchain command constants
func (ts *ConstantsTestSuite) TestGoToolchainCommands() {
	// Verify Go toolchain commands
	require.Equal(ts.T(), "go", CmdGo)
	require.Equal(ts.T(), "build", CmdGoBuild)
	require.Equal(ts.T(), "test", CmdGoTest)
	require.Equal(ts.T(), "mod", CmdGoMod)
	require.Equal(ts.T(), "generate", CmdGoGenerate)
	require.Equal(ts.T(), "install", CmdGoInstall)
	require.Equal(ts.T(), "get", CmdGoGet)
	require.Equal(ts.T(), "list", CmdGoList)
	require.Equal(ts.T(), "vet", CmdGoVet)

	// Ensure none are empty
	goCommands := []string{
		CmdGo, CmdGoBuild, CmdGoTest, CmdGoMod, CmdGoGenerate,
		CmdGoInstall, CmdGoGet, CmdGoList, CmdGoVet,
	}
	for _, cmd := range goCommands {
		require.NotEmpty(ts.T(), cmd, "Go command should not be empty")
	}
}

// TestExternalToolCommands tests external tool command constants
func (ts *ConstantsTestSuite) TestExternalToolCommands() {
	// Verify external tool commands
	require.Equal(ts.T(), "git", CmdGit)
	require.Equal(ts.T(), "docker", CmdDocker)
	require.Equal(ts.T(), "golangci-lint", CmdGolangciLint)
	require.Equal(ts.T(), "gofumpt", CmdGofumpt)
	require.Equal(ts.T(), "golangci-lint", LintTool) // Should match CmdGolangciLint
	require.Equal(ts.T(), "govulncheck", CmdGoVulnCheck)
	require.Equal(ts.T(), "mockgen", CmdMockgen)
	require.Equal(ts.T(), "swag", CmdSwag)

	// Verify LintTool matches CmdGolangciLint
	require.Equal(ts.T(), CmdGolangciLint, LintTool)

	// Ensure none are empty
	externalCommands := []string{
		CmdGit, CmdDocker, CmdGolangciLint, CmdGofumpt,
		LintTool, CmdGoVulnCheck, CmdMockgen, CmdSwag,
	}
	for _, cmd := range externalCommands {
		require.NotEmpty(ts.T(), cmd, "External command should not be empty")
	}
}

// TestShellCommands tests shell command constants
func (ts *ConstantsTestSuite) TestShellCommands() {
	// Verify shell commands
	require.Equal(ts.T(), "find", CmdFind)
	require.Equal(ts.T(), "wc", CmdWC)
	require.Equal(ts.T(), "rm", CmdRM)
	require.Equal(ts.T(), "mkdir", CmdMkdir)

	// Ensure none are empty
	shellCommands := []string{CmdFind, CmdWC, CmdRM, CmdMkdir}
	for _, cmd := range shellCommands {
		require.NotEmpty(ts.T(), cmd, "Shell command should not be empty")
	}
}

// TestGoFlags tests Go build and test flag constants
func (ts *ConstantsTestSuite) TestGoFlags() {
	// Verify Go flags start with hyphen
	goFlags := map[string]string{
		"FlagOutput":    FlagOutput,
		"FlagVerbose":   FlagVerbose,
		"FlagTags":      FlagTags,
		"FlagLDFlags":   FlagLDFlags,
		"FlagTrimPath":  FlagTrimPath,
		"FlagRace":      FlagRace,
		"FlagCover":     FlagCover,
		"FlagCoverMode": FlagCoverMode,
		"FlagCoverPkg":  FlagCoverPkg,
		"FlagTimeout":   FlagTimeout,
		"FlagShort":     FlagShort,
		"FlagParallel":  FlagParallel,
		"FlagBench":     FlagBench,
		"FlagRun":       FlagRun,
		"FlagCount":     FlagCount,
		"FlagCPU":       FlagCPU,
	}

	for name, flag := range goFlags {
		require.True(ts.T(), strings.HasPrefix(flag, "-"),
			"Go flag %s (%s) should start with hyphen", name, flag)
		require.NotEmpty(ts.T(), flag, "Go flag %s should not be empty", name)
	}

	// Test specific flag values
	require.Equal(ts.T(), "-o", FlagOutput)
	require.Equal(ts.T(), "-v", FlagVerbose)
	require.Equal(ts.T(), "-race", FlagRace)
	require.Equal(ts.T(), "-cover", FlagCover)
}

// TestGitFlags tests Git flag constants
func (ts *ConstantsTestSuite) TestGitFlags() {
	// Verify Git flags start with double hyphen
	gitFlags := []string{FlagTags2, FlagAbbrev, FlagNoVerify}
	for _, flag := range gitFlags {
		require.True(ts.T(), strings.HasPrefix(flag, "--"),
			"Git flag (%s) should start with double hyphen", flag)
		require.NotEmpty(ts.T(), flag, "Git flag should not be empty")
	}

	// Test specific values
	require.Equal(ts.T(), "--tags", FlagTags2)
	require.Equal(ts.T(), "--abbrev=0", FlagAbbrev)
	require.Equal(ts.T(), "--no-verify", FlagNoVerify)
}

// TestCommonArgs tests common argument constants
func (ts *ConstantsTestSuite) TestCommonArgs() {
	require.Equal(ts.T(), "./...", ArgAll)
	require.Equal(ts.T(), ".", ArgCurrent)
	require.Equal(ts.T(), "none", ArgNone)

	// Ensure none are empty
	args := []string{ArgAll, ArgCurrent, ArgNone}
	for _, arg := range args {
		require.NotEmpty(ts.T(), arg, "Common arg should not be empty")
	}
}

// TestFileAndDirectoryNames tests file and directory name constants
func (ts *ConstantsTestSuite) TestFileAndDirectoryNames() {
	// Test file names
	fileNames := map[string]string{
		"FileGoMod":        FileGoMod,
		"FileGoSum":        FileGoSum,
		"FileMageYAML":     FileMageYAML,
		"FileMageYML":      FileMageYML,
		"FileGitignore":    FileGitignore,
		"FileDockerfile":   FileDockerfile,
		"FileVersion":      FileVersion,
		"FileCoverageOut":  FileCoverageOut,
		"FileCoverageHTML": FileCoverageHTML,
	}

	for name, fileName := range fileNames {
		require.NotEmpty(ts.T(), fileName, "File name %s should not be empty", name)
	}

	// Test specific file name values
	require.Equal(ts.T(), "go.mod", FileGoMod)
	require.Equal(ts.T(), "go.sum", FileGoSum)
	require.Equal(ts.T(), ".mage.yaml", FileMageYAML)
	require.Equal(ts.T(), ".mage.yml", FileMageYML)
	require.Equal(ts.T(), "VERSION", FileVersion)

	// Test directory names
	dirNames := []string{DirBin, DirBuild, DirDist, DirVendor, DirTestdata, DirCmd, DirPkg, DirInternal}
	for _, dirName := range dirNames {
		require.NotEmpty(ts.T(), dirName, "Directory name should not be empty")
		require.False(ts.T(), strings.HasPrefix(dirName, "/"),
			"Directory name (%s) should be relative", dirName)
	}

	// Test specific directory values
	require.Equal(ts.T(), "bin", DirBin)
	require.Equal(ts.T(), "pkg", DirPkg)
	require.Equal(ts.T(), "cmd", DirCmd)
}

// TestEnvironmentVariables tests environment variable constants
func (ts *ConstantsTestSuite) TestEnvironmentVariables() {
	envVars := map[string]string{
		"EnvGOOS":        EnvGOOS,
		"EnvGOARCH":      EnvGOARCH,
		"EnvGOPATH":      EnvGOPATH,
		"EnvGOBIN":       EnvGOBIN,
		"EnvCGOEnabled":  EnvCGOEnabled,
		"EnvPath":        EnvPath,
		"EnvDebug":       EnvDebug,
		"EnvVerbose":     EnvVerbose,
		"EnvDryRun":      EnvDryRun,
		"EnvNoColor":     EnvNoColor,
		"EnvForceColor":  EnvForceColor,
		"EnvCI":          EnvCI,
		"EnvGitHubToken": EnvGitHubToken,
	}

	for name, envVar := range envVars {
		require.NotEmpty(ts.T(), envVar, "Environment variable %s should not be empty", name)
		require.True(ts.T(), strings.ToUpper(envVar) == envVar,
			"Environment variable %s (%s) should be uppercase", name, envVar)
	}

	// Test specific values
	require.Equal(ts.T(), "GOOS", EnvGOOS)
	require.Equal(ts.T(), "GOARCH", EnvGOARCH)
	require.Equal(ts.T(), "CGO_ENABLED", EnvCGOEnabled)
	require.Equal(ts.T(), "GITHUB_TOKEN", EnvGitHubToken)
}

// TestPlatformConstants tests platform-related constants
func (ts *ConstantsTestSuite) TestPlatformConstants() {
	// Test OS constants
	operatingSystems := []string{OSLinux, OSDarwin, OSWindows}
	for _, os := range operatingSystems {
		require.NotEmpty(ts.T(), os, "OS constant should not be empty")
		require.True(ts.T(), strings.ToLower(os) == os,
			"OS constant (%s) should be lowercase", os)
	}

	require.Equal(ts.T(), "linux", OSLinux)
	require.Equal(ts.T(), "darwin", OSDarwin)
	require.Equal(ts.T(), "windows", OSWindows)

	// Test architecture constants
	architectures := []string{ArchAMD64, ArchARM64, Arch386, ArchARM}
	for _, arch := range architectures {
		require.NotEmpty(ts.T(), arch, "Architecture constant should not be empty")
	}

	require.Equal(ts.T(), "amd64", ArchAMD64)
	require.Equal(ts.T(), "arm64", ArchARM64)
	require.Equal(ts.T(), "386", Arch386)
	require.Equal(ts.T(), "arm", ArchARM)
}

// TestTimeoutConstants tests timeout and duration constants
func (ts *ConstantsTestSuite) TestTimeoutConstants() {
	timeouts := map[string]string{
		"DefaultTimeout":   DefaultTimeout,
		"DefaultBenchTime": DefaultBenchTime,
		"ShortTimeout":     ShortTimeout,
		"LongTimeout":      LongTimeout,
	}

	for name, timeout := range timeouts {
		require.NotEmpty(ts.T(), timeout, "Timeout %s should not be empty", name)
		require.True(ts.T(), strings.HasSuffix(timeout, "m") || strings.HasSuffix(timeout, "s"),
			"Timeout %s (%s) should end with 'm' or 's'", name, timeout)
	}

	require.Equal(ts.T(), "10m", DefaultTimeout)
	require.Equal(ts.T(), "10s", DefaultBenchTime)
	require.Equal(ts.T(), "1m", ShortTimeout)
	require.Equal(ts.T(), "30m", LongTimeout)
}

// TestCoverageModes tests coverage mode constants
func (ts *ConstantsTestSuite) TestCoverageModes() {
	coverageModes := []string{CoverModeSet, CoverModeCount, CoverModeAtomic}
	validModes := map[string]bool{
		"set":    true,
		"count":  true,
		"atomic": true,
	}

	for _, mode := range coverageModes {
		require.NotEmpty(ts.T(), mode, "Coverage mode should not be empty")
		require.True(ts.T(), validModes[mode],
			"Coverage mode (%s) should be valid Go coverage mode", mode)
	}

	require.Equal(ts.T(), "set", CoverModeSet)
	require.Equal(ts.T(), "count", CoverModeCount)
	require.Equal(ts.T(), "atomic", CoverModeAtomic)
}

// TestReleaseChannels tests release channel constants
func (ts *ConstantsTestSuite) TestReleaseChannels() {
	channels := []string{ChannelStable, ChannelBeta, ChannelEdge, ChannelNightly}
	for _, channel := range channels {
		require.NotEmpty(ts.T(), channel, "Release channel should not be empty")
		require.True(ts.T(), strings.ToLower(channel) == channel,
			"Release channel (%s) should be lowercase", channel)
	}

	require.Equal(ts.T(), "stable", ChannelStable)
	require.Equal(ts.T(), "beta", ChannelBeta)
	require.Equal(ts.T(), "edge", ChannelEdge)
	require.Equal(ts.T(), "nightly", ChannelNightly)
}

// TestToolVersions tests default tool version constants
func (ts *ConstantsTestSuite) TestToolVersions() {
	toolVersions := map[string]string{
		"DefaultGolangciLintVersion": DefaultGolangciLintVersion,
		"DefaultGofumptVersion":      DefaultGofumptVersion,
		"DefaultGoVulnCheckVersion":  DefaultGoVulnCheckVersion,
		"DefaultMockgenVersion":      DefaultMockgenVersion,
		"DefaultSwagVersion":         DefaultSwagVersion,
	}

	for name, version := range toolVersions {
		require.NotEmpty(ts.T(), version, "Tool version %s should not be empty", name)
		if version != "latest" {
			require.True(ts.T(), strings.HasPrefix(version, "v"),
				"Tool version %s (%s) should start with 'v' or be 'latest'", name, version)
		}
	}

	require.Equal(ts.T(), "latest", DefaultGoVulnCheckVersion)
	require.True(ts.T(), strings.HasPrefix(DefaultGolangciLintVersion, "v"))
}

// TestErrorMessages tests error message constants
func (ts *ConstantsTestSuite) TestErrorMessages() {
	errorMessages := []string{
		ErrNoGoMod, ErrNoGitRepo, ErrNoVersion,
		ErrBuildFailed, ErrTestFailed, ErrLintFailed, ErrToolNotFound,
	}

	for _, msg := range errorMessages {
		require.NotEmpty(ts.T(), msg, "Error message should not be empty")
		require.False(ts.T(), strings.HasSuffix(msg, "."),
			"Error message (%s) should not end with period", msg)
	}

	require.Equal(ts.T(), "no go.mod file found", ErrNoGoMod)
	require.Equal(ts.T(), "not a git repository", ErrNoGitRepo)
	require.Equal(ts.T(), "build failed", ErrBuildFailed)
}

// TestSuccessMessages tests success message constants
func (ts *ConstantsTestSuite) TestSuccessMessages() {
	successMessages := []string{
		MsgBuildSuccess, MsgTestSuccess, MsgLintSuccess,
		MsgInstallSuccess, MsgCleanSuccess,
	}

	for _, msg := range successMessages {
		require.NotEmpty(ts.T(), msg, "Success message should not be empty")
		require.True(ts.T(), strings.Contains(strings.ToLower(msg), "success") ||
			strings.Contains(strings.ToLower(msg), "passed") ||
			strings.Contains(strings.ToLower(msg), "completed") ||
			strings.Contains(strings.ToLower(msg), "no") && strings.Contains(strings.ToLower(msg), "issues"),
			"Success message (%s) should indicate success", msg)
	}

	require.Contains(ts.T(), MsgBuildSuccess, "successfully")
	require.Contains(ts.T(), MsgTestSuccess, "passed")
}

// TestInfoMessages tests info message constants
func (ts *ConstantsTestSuite) TestInfoMessages() {
	infoMessages := []string{
		MsgBuildingApp, MsgRunningTests, MsgRunningLint,
		MsgInstalling, MsgCleaning, MsgGenerating,
	}

	for _, msg := range infoMessages {
		require.NotEmpty(ts.T(), msg, "Info message should not be empty")
		// Info messages should be present tense action words
		require.True(ts.T(), strings.Contains(strings.ToLower(msg), "running") ||
			strings.Contains(strings.ToLower(msg), "building") ||
			strings.Contains(strings.ToLower(msg), "installing") ||
			strings.Contains(strings.ToLower(msg), "cleaning") ||
			strings.Contains(strings.ToLower(msg), "generating"),
			"Info message (%s) should be an action", msg)
	}
}

// TestEmojiConstants tests emoji constants
func (ts *ConstantsTestSuite) TestEmojiConstants() {
	emojis := map[string]string{
		"EmojiBuild":   EmojiBuild,
		"EmojiTest":    EmojiTest,
		"EmojiLint":    EmojiLint,
		"EmojiSuccess": EmojiSuccess,
		"EmojiError":   EmojiError,
		"EmojiWarning": EmojiWarning,
		"EmojiInfo":    EmojiInfo,
		"EmojiRocket":  EmojiRocket,
		"EmojiPackage": EmojiPackage,
		"EmojiClean":   EmojiClean,
		"EmojiTarget":  EmojiTarget,
		"EmojiClock":   EmojiClock,
		"EmojiShield":  EmojiShield,
		"EmojiChart":   EmojiChart,
		"EmojiBook":    EmojiBook,
		"EmojiGear":    EmojiGear,
		"EmojiRefresh": EmojiRefresh,
	}

	for name, emoji := range emojis {
		require.NotEmpty(ts.T(), emoji, "Emoji %s should not be empty", name)
		// Each emoji should be non-empty and contain valid Unicode
		require.True(ts.T(), len([]rune(emoji)) >= 1,
			"Emoji %s (%s) should contain at least one Unicode character", name, emoji)
	}

	// Test specific emoji values
	require.Equal(ts.T(), "🔨", EmojiBuild)
	require.Equal(ts.T(), "🧪", EmojiTest)
	require.Equal(ts.T(), "✅", EmojiSuccess)
	require.Equal(ts.T(), "❌", EmojiError)
}

// TestFormatStrings tests format string constants
func (ts *ConstantsTestSuite) TestFormatStrings() {
	formatStrings := map[string]string{
		"FmtPlatform":    FmtPlatform,
		"FmtBuildTag":    FmtBuildTag,
		"FmtTestPackage": FmtTestPackage,
		"FmtInstallTool": FmtInstallTool,
		"FmtVersion":     FmtVersion,
		"FmtDuration":    FmtDuration,
		"FmtCoverage":    FmtCoverage,
	}

	for name, format := range formatStrings {
		require.NotEmpty(ts.T(), format, "Format string %s should not be empty", name)
		require.Contains(ts.T(), format, "%",
			"Format string %s (%s) should contain format specifier", name, format)
	}

	// Test specific format string values
	require.Equal(ts.T(), "%s/%s", FmtPlatform)
	require.Equal(ts.T(), "Building for %s", FmtBuildTag)
	require.Equal(ts.T(), "Version: %s", FmtVersion)
	require.Equal(ts.T(), "Coverage: %.1f%%", FmtCoverage)

	// Verify format strings have correct number of placeholders
	require.Equal(ts.T(), 2, strings.Count(FmtPlatform, "%s"))
	require.Equal(ts.T(), 1, strings.Count(FmtBuildTag, "%s"))
	require.Equal(ts.T(), 2, strings.Count(FmtInstallTool, "%s"))
}

// TestConstantsTestSuite runs the test suite
func TestConstantsTestSuite(t *testing.T) {
	suite.Run(t, new(ConstantsTestSuite))
}
