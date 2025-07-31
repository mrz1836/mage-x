package mage

import (
	"strings"
	"testing"

	"github.com/mrz1836/go-mage/pkg/mage/testutil"
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
	ts.Require().Equal("go", CmdGo)
	ts.Require().Equal("build", CmdGoBuild)
	ts.Require().Equal("test", CmdGoTest)
	ts.Require().Equal("mod", CmdGoMod)
	ts.Require().Equal("generate", CmdGoGenerate)
	ts.Require().Equal("install", CmdGoInstall)
	ts.Require().Equal("get", CmdGoGet)
	ts.Require().Equal("list", CmdGoList)
	ts.Require().Equal("vet", CmdGoVet)

	// Ensure none are empty
	goCommands := []string{
		CmdGo, CmdGoBuild, CmdGoTest, CmdGoMod, CmdGoGenerate,
		CmdGoInstall, CmdGoGet, CmdGoList, CmdGoVet,
	}
	for _, cmd := range goCommands {
		ts.Require().NotEmpty(cmd, "Go command should not be empty")
	}
}

// TestExternalToolCommands tests external tool command constants
func (ts *ConstantsTestSuite) TestExternalToolCommands() {
	// Verify external tool commands
	ts.Require().Equal("git", CmdGit)
	ts.Require().Equal("docker", CmdDocker)
	ts.Require().Equal("golangci-lint", CmdGolangciLint)
	ts.Require().Equal("gofumpt", CmdGofumpt)
	ts.Require().Equal("golangci-lint", LintTool) // Should match CmdGolangciLint
	ts.Require().Equal("govulncheck", CmdGoVulnCheck)
	ts.Require().Equal("mockgen", CmdMockgen)
	ts.Require().Equal("swag", CmdSwag)

	// Verify LintTool matches CmdGolangciLint
	ts.Require().Equal(CmdGolangciLint, LintTool)

	// Ensure none are empty
	externalCommands := []string{
		CmdGit, CmdDocker, CmdGolangciLint, CmdGofumpt,
		LintTool, CmdGoVulnCheck, CmdMockgen, CmdSwag,
	}
	for _, cmd := range externalCommands {
		ts.Require().NotEmpty(cmd, "External command should not be empty")
	}
}

// TestShellCommands tests shell command constants
func (ts *ConstantsTestSuite) TestShellCommands() {
	// Verify shell commands
	ts.Require().Equal("find", CmdFind)
	ts.Require().Equal("wc", CmdWC)
	ts.Require().Equal("rm", CmdRM)
	ts.Require().Equal("mkdir", CmdMkdir)

	// Ensure none are empty
	shellCommands := []string{CmdFind, CmdWC, CmdRM, CmdMkdir}
	for _, cmd := range shellCommands {
		ts.Require().NotEmpty(cmd, "Shell command should not be empty")
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
		ts.Require().True(strings.HasPrefix(flag, "-"),
			"Go flag %s (%s) should start with hyphen", name, flag)
		ts.Require().NotEmpty(flag, "Go flag %s should not be empty", name)
	}

	// Test specific flag values
	ts.Require().Equal("-o", FlagOutput)
	ts.Require().Equal("-v", FlagVerbose)
	ts.Require().Equal("-race", FlagRace)
	ts.Require().Equal("-cover", FlagCover)
}

// TestGitFlags tests Git flag constants
func (ts *ConstantsTestSuite) TestGitFlags() {
	// Verify Git flags start with double hyphen
	gitFlags := []string{FlagTags2, FlagAbbrev, FlagNoVerify}
	for _, flag := range gitFlags {
		ts.Require().True(strings.HasPrefix(flag, "--"),
			"Git flag (%s) should start with double hyphen", flag)
		ts.Require().NotEmpty(flag, "Git flag should not be empty")
	}

	// Test specific values
	ts.Require().Equal("--tags", FlagTags2)
	ts.Require().Equal("--abbrev=0", FlagAbbrev)
	ts.Require().Equal("--no-verify", FlagNoVerify)
}

// TestCommonArgs tests common argument constants
func (ts *ConstantsTestSuite) TestCommonArgs() {
	ts.Require().Equal("./...", ArgAll)
	ts.Require().Equal(".", ArgCurrent)
	ts.Require().Equal("none", ArgNone)

	// Ensure none are empty
	args := []string{ArgAll, ArgCurrent, ArgNone}
	for _, arg := range args {
		ts.Require().NotEmpty(arg, "Common arg should not be empty")
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
		ts.Require().NotEmpty(fileName, "File name %s should not be empty", name)
	}

	// Test specific file name values
	ts.Require().Equal("go.mod", FileGoMod)
	ts.Require().Equal("go.sum", FileGoSum)
	ts.Require().YAMLEq(".mage.yaml", FileMageYAML)
	ts.Require().YAMLEq(".mage.yml", FileMageYML)
	ts.Require().Equal("VERSION", FileVersion)

	// Test directory names
	dirNames := []string{DirBin, DirBuild, DirDist, DirVendor, DirTestdata, DirCmd, DirPkg, DirInternal}
	for _, dirName := range dirNames {
		ts.Require().NotEmpty(dirName, "Directory name should not be empty")
		ts.Require().False(strings.HasPrefix(dirName, "/"),
			"Directory name (%s) should be relative", dirName)
	}

	// Test specific directory values
	ts.Require().Equal("bin", DirBin)
	ts.Require().Equal("pkg", DirPkg)
	ts.Require().Equal("cmd", DirCmd)
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
		ts.Require().NotEmpty(envVar, "Environment variable %s should not be empty", name)
		ts.Require().Equal(strings.ToUpper(envVar), envVar,
			"Environment variable %s (%s) should be uppercase", name, envVar)
	}

	// Test specific values
	ts.Require().Equal("GOOS", EnvGOOS)
	ts.Require().Equal("GOARCH", EnvGOARCH)
	ts.Require().Equal("CGO_ENABLED", EnvCGOEnabled)
	ts.Require().Equal("GITHUB_TOKEN", EnvGitHubToken)
}

// TestPlatformConstants tests platform-related constants
func (ts *ConstantsTestSuite) TestPlatformConstants() {
	// Test OS constants
	operatingSystems := []string{OSLinux, OSDarwin, OSWindows}
	for _, os := range operatingSystems {
		ts.Require().NotEmpty(os, "OS constant should not be empty")
		ts.Require().Equal(strings.ToLower(os), os,
			"OS constant (%s) should be lowercase", os)
	}

	ts.Require().Equal("linux", OSLinux)
	ts.Require().Equal("darwin", OSDarwin)
	ts.Require().Equal("windows", OSWindows)

	// Test architecture constants
	architectures := []string{ArchAMD64, ArchARM64, Arch386, ArchARM}
	for _, arch := range architectures {
		ts.Require().NotEmpty(arch, "Architecture constant should not be empty")
	}

	ts.Require().Equal("amd64", ArchAMD64)
	ts.Require().Equal("arm64", ArchARM64)
	ts.Require().Equal("386", Arch386)
	ts.Require().Equal("arm", ArchARM)
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
		ts.Require().NotEmpty(timeout, "Timeout %s should not be empty", name)
		ts.Require().True(strings.HasSuffix(timeout, "m") || strings.HasSuffix(timeout, "s"),
			"Timeout %s (%s) should end with 'm' or 's'", name, timeout)
	}

	ts.Require().Equal("10m", DefaultTimeout)
	ts.Require().Equal("10s", DefaultBenchTime)
	ts.Require().Equal("1m", ShortTimeout)
	ts.Require().Equal("30m", LongTimeout)
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
		ts.Require().NotEmpty(mode, "Coverage mode should not be empty")
		ts.Require().True(validModes[mode],
			"Coverage mode (%s) should be valid Go coverage mode", mode)
	}

	ts.Require().Equal("set", CoverModeSet)
	ts.Require().Equal("count", CoverModeCount)
	ts.Require().Equal("atomic", CoverModeAtomic)
}

// TestReleaseChannels tests release channel constants
func (ts *ConstantsTestSuite) TestReleaseChannels() {
	channels := []string{ChannelStable, ChannelBeta, ChannelEdge, ChannelNightly}
	for _, channel := range channels {
		ts.Require().NotEmpty(channel, "Release channel should not be empty")
		ts.Require().Equal(strings.ToLower(channel), channel,
			"Release channel (%s) should be lowercase", channel)
	}

	ts.Require().Equal("stable", ChannelStable)
	ts.Require().Equal("beta", ChannelBeta)
	ts.Require().Equal("edge", ChannelEdge)
	ts.Require().Equal("nightly", ChannelNightly)
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
		ts.Require().NotEmpty(version, "Tool version %s should not be empty", name)
		if version != "latest" {
			ts.Require().True(strings.HasPrefix(version, "v"),
				"Tool version %s (%s) should start with 'v' or be 'latest'", name, version)
		}
	}

	ts.Require().Equal("latest", DefaultGoVulnCheckVersion)
	ts.Require().True(strings.HasPrefix(DefaultGolangciLintVersion, "v"))
}

// TestErrorMessages tests error message constants
func (ts *ConstantsTestSuite) TestErrorMessages() {
	errorMessages := []string{
		ErrNoGoMod, ErrNoGitRepo, ErrNoVersion,
		ErrBuildFailed, ErrTestFailed, ErrLintFailed, ErrToolNotFound,
	}

	for _, msg := range errorMessages {
		ts.Require().NotEmpty(msg, "Error message should not be empty")
		ts.Require().False(strings.HasSuffix(msg, "."),
			"Error message (%s) should not end with period", msg)
	}

	ts.Require().Equal("no go.mod file found", ErrNoGoMod)
	ts.Require().Equal("not a git repository", ErrNoGitRepo)
	ts.Require().Equal("build failed", ErrBuildFailed)
}

// TestSuccessMessages tests success message constants
func (ts *ConstantsTestSuite) TestSuccessMessages() {
	successMessages := []string{
		MsgBuildSuccess, MsgTestSuccess, MsgLintSuccess,
		MsgInstallSuccess, MsgCleanSuccess,
	}

	for _, msg := range successMessages {
		ts.Require().NotEmpty(msg, "Success message should not be empty")
		ts.Require().True(strings.Contains(strings.ToLower(msg), "success") ||
			strings.Contains(strings.ToLower(msg), "passed") ||
			strings.Contains(strings.ToLower(msg), "completed") ||
			strings.Contains(strings.ToLower(msg), "no") && strings.Contains(strings.ToLower(msg), "issues"),
			"Success message (%s) should indicate success", msg)
	}

	ts.Require().Contains(MsgBuildSuccess, "successfully")
	ts.Require().Contains(MsgTestSuccess, "passed")
}

// TestInfoMessages tests info message constants
func (ts *ConstantsTestSuite) TestInfoMessages() {
	infoMessages := []string{
		MsgBuildingApp, MsgRunningTests, MsgRunningLint,
		MsgInstalling, MsgCleaning, MsgGenerating,
	}

	for _, msg := range infoMessages {
		ts.Require().NotEmpty(msg, "Info message should not be empty")
		// Info messages should be present tense action words
		ts.Require().True(strings.Contains(strings.ToLower(msg), "running") ||
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
		ts.Require().NotEmpty(emoji, "Emoji %s should not be empty", name)
		// Each emoji should be non-empty and contain valid Unicode
		ts.Require().GreaterOrEqual(len([]rune(emoji)), 1,
			"Emoji %s (%s) should contain at least one Unicode character", name, emoji)
	}

	// Test specific emoji values
	ts.Require().Equal("🔨", EmojiBuild)
	ts.Require().Equal("🧪", EmojiTest)
	ts.Require().Equal("✅", EmojiSuccess)
	ts.Require().Equal("❌", EmojiError)
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
		ts.Require().NotEmpty(format, "Format string %s should not be empty", name)
		ts.Require().Contains(format, "%",
			"Format string %s (%s) should contain format specifier", name, format)
	}

	// Test specific format string values
	ts.Require().Equal("%s/%s", FmtPlatform)
	ts.Require().Equal("Building for %s", FmtBuildTag)
	ts.Require().Equal("Version: %s", FmtVersion)
	ts.Require().Equal("Coverage: %.1f%%", FmtCoverage)

	// Verify format strings have correct number of placeholders
	ts.Require().Equal(2, strings.Count(FmtPlatform, "%s"))
	ts.Require().Equal(1, strings.Count(FmtBuildTag, "%s"))
	ts.Require().Equal(2, strings.Count(FmtInstallTool, "%s"))
}

// TestConstantsTestSuite runs the test suite
func TestConstantsTestSuite(t *testing.T) {
	suite.Run(t, new(ConstantsTestSuite))
}
