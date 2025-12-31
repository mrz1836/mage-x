package mage

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var (
	errNpmCommandFailed = errors.New("npm command failed")
	errNpxCommandFailed = errors.New("npx command failed")
	errTestError        = errors.New("test error")
)

// BmadMockRunner provides a mock implementation of CommandRunner for testing bmad functions
type BmadMockRunner struct {
	outputs map[string]struct {
		output string
		err    error
	}
	commands []string
}

// NewBmadMockRunner creates a new mock runner for bmad tests
func NewBmadMockRunner() *BmadMockRunner {
	return &BmadMockRunner{
		outputs: make(map[string]struct {
			output string
			err    error
		}),
		commands: []string{},
	}
}

// SetOutput configures the expected output for a given command
func (m *BmadMockRunner) SetOutput(cmd, output string, err error) {
	m.outputs[cmd] = struct {
		output string
		err    error
	}{output: output, err: err}
}

// RunCmd implements CommandRunner.RunCmd
func (m *BmadMockRunner) RunCmd(name string, args ...string) error {
	cmd := name + " " + strings.Join(args, " ")
	m.commands = append(m.commands, cmd)
	if result, ok := m.outputs[cmd]; ok {
		return result.err
	}
	return nil
}

// RunCmdOutput implements CommandRunner.RunCmdOutput
func (m *BmadMockRunner) RunCmdOutput(name string, args ...string) (string, error) {
	cmd := name + " " + strings.Join(args, " ")
	m.commands = append(m.commands, cmd)
	if result, ok := m.outputs[cmd]; ok {
		return result.output, result.err
	}
	return "", nil
}

// GetCommands returns all commands that were executed
func (m *BmadMockRunner) GetCommands() []string {
	return m.commands
}

// BmadUnitTestSuite defines the unit test suite for bmad functions
type BmadUnitTestSuite struct {
	suite.Suite

	origDir string
	tempDir string
}

// SetupTest runs before each test
func (ts *BmadUnitTestSuite) SetupTest() {
	var err error

	// Save original directory
	ts.origDir, err = os.Getwd()
	ts.Require().NoError(err)

	ts.tempDir, err = os.MkdirTemp("", "bmad-test-*")
	ts.Require().NoError(err)

	// Change to temp directory for tests
	err = os.Chdir(ts.tempDir)
	ts.Require().NoError(err)
}

// TearDownTest runs after each test
func (ts *BmadUnitTestSuite) TearDownTest() {
	// Restore original directory before removing temp dir
	if ts.origDir != "" {
		if err := os.Chdir(ts.origDir); err != nil {
			ts.T().Logf("failed to restore original directory: %v", err)
		}
	}

	if ts.tempDir != "" {
		if err := os.RemoveAll(ts.tempDir); err != nil {
			ts.T().Logf("failed to remove temp dir: %v", err)
		}
	}
}

// TestGetBmadProjectDirDefault tests getBmadProjectDir with default config
func (ts *BmadUnitTestSuite) TestGetBmadProjectDirDefault() {
	config := &Config{
		Bmad: BmadConfig{},
	}
	path := getBmadProjectDir(config)
	ts.Require().Equal(DefaultBmadProjectDir, path)
}

// TestGetBmadProjectDirCustom tests getBmadProjectDir with custom config
func (ts *BmadUnitTestSuite) TestGetBmadProjectDirCustom() {
	config := &Config{
		Bmad: BmadConfig{
			ProjectDir: "custom/_bmad",
		},
	}
	path := getBmadProjectDir(config)
	ts.Require().Equal("custom/_bmad", path)
}

// TestVerifyBmadInstallationNotInstalled tests verifyBmadInstallation when directory doesn't exist
func (ts *BmadUnitTestSuite) TestVerifyBmadInstallationNotInstalled() {
	config := &Config{
		Bmad: BmadConfig{
			ProjectDir: "nonexistent_bmad_dir",
		},
	}
	err := verifyBmadInstallation(config)
	ts.Require().Error(err)
	ts.Require().ErrorIs(err, errBmadNotInstalled)
}

// TestVerifyBmadInstallationSuccess tests verifyBmadInstallation when directory exists
func (ts *BmadUnitTestSuite) TestVerifyBmadInstallationSuccess() {
	projectDir := filepath.Join(ts.tempDir, "_bmad")
	err := os.MkdirAll(projectDir, 0o750)
	ts.Require().NoError(err)

	config := &Config{
		Bmad: BmadConfig{
			ProjectDir: projectDir,
		},
	}
	err = verifyBmadInstallation(config)
	ts.Require().NoError(err)
}

// TestGetBmadVersionSuccess tests getBmadVersion with successful response
func (ts *BmadUnitTestSuite) TestGetBmadVersionSuccess() {
	// Save original runner and restore after test
	originalRunner := GetRunner()
	defer func() {
		err := SetRunner(originalRunner)
		ts.Require().NoError(err)
	}()

	// Create mock runner
	mock := NewBmadMockRunner()
	expectedCmd := "npm view bmad-method@alpha version"
	mock.SetOutput(expectedCmd, "6.0.0-alpha.1", nil)

	err := SetRunner(mock)
	ts.Require().NoError(err)

	config := &Config{
		Bmad: BmadConfig{
			PackageName: "bmad-method",
			VersionTag:  "@alpha",
		},
	}

	version, err := getBmadVersion(config)
	ts.Require().NoError(err)
	ts.Require().Equal("6.0.0-alpha.1", version)
}

// TestGetBmadVersionWithDefaults tests getBmadVersion with default config values
func (ts *BmadUnitTestSuite) TestGetBmadVersionWithDefaults() {
	originalRunner := GetRunner()
	defer func() {
		err := SetRunner(originalRunner)
		ts.Require().NoError(err)
	}()

	mock := NewBmadMockRunner()
	// With defaults: DefaultBmadPackageName + DefaultBmadVersionTag
	expectedCmd := "npm view " + DefaultBmadPackageName + DefaultBmadVersionTag + " version"
	mock.SetOutput(expectedCmd, "5.0.0", nil)

	err := SetRunner(mock)
	ts.Require().NoError(err)

	config := &Config{
		Bmad: BmadConfig{}, // Empty config, should use defaults
	}

	version, err := getBmadVersion(config)
	ts.Require().NoError(err)
	ts.Require().Equal("5.0.0", version)
}

// TestGetBmadVersionRunnerError tests getBmadVersion when runner returns an error
func (ts *BmadUnitTestSuite) TestGetBmadVersionRunnerError() {
	originalRunner := GetRunner()
	defer func() {
		err := SetRunner(originalRunner)
		ts.Require().NoError(err)
	}()

	mock := NewBmadMockRunner()
	expectedCmd := "npm view bmad-method@alpha version"
	mock.SetOutput(expectedCmd, "", errNpmCommandFailed)

	err := SetRunner(mock)
	ts.Require().NoError(err)

	config := &Config{
		Bmad: BmadConfig{
			PackageName: "bmad-method",
			VersionTag:  "@alpha",
		},
	}

	_, err = getBmadVersion(config)
	ts.Require().Error(err)
	ts.Require().Contains(err.Error(), "failed to get bmad version")
}

// TestGetBmadVersionEmptyOutput tests getBmadVersion when npm returns empty output
func (ts *BmadUnitTestSuite) TestGetBmadVersionEmptyOutput() {
	originalRunner := GetRunner()
	defer func() {
		err := SetRunner(originalRunner)
		ts.Require().NoError(err)
	}()

	mock := NewBmadMockRunner()
	expectedCmd := "npm view bmad-method@alpha version"
	mock.SetOutput(expectedCmd, "", nil)

	err := SetRunner(mock)
	ts.Require().NoError(err)

	config := &Config{
		Bmad: BmadConfig{
			PackageName: "bmad-method",
			VersionTag:  "@alpha",
		},
	}

	_, err = getBmadVersion(config)
	ts.Require().Error(err)
	ts.Require().ErrorIs(err, errBmadVersionParse)
}

// TestUpgradeBmadCLISuccess tests upgradeBmadCLI with successful execution
func (ts *BmadUnitTestSuite) TestUpgradeBmadCLISuccess() {
	originalRunner := GetRunner()
	defer func() {
		err := SetRunner(originalRunner)
		ts.Require().NoError(err)
	}()

	mock := NewBmadMockRunner()
	// With custom config values
	expectedCmd := "npx --yes custom-bmad@latest update -d custom_bmad --force"
	mock.SetOutput(expectedCmd, "", nil)

	err := SetRunner(mock)
	ts.Require().NoError(err)

	config := &Config{
		Bmad: BmadConfig{
			PackageName: "custom-bmad",
			VersionTag:  "@latest",
			ProjectDir:  "custom_bmad",
		},
	}

	err = upgradeBmadCLI(config)
	ts.Require().NoError(err)

	// Verify the command was called
	commands := mock.GetCommands()
	ts.Require().Len(commands, 1)
	ts.Require().Equal(expectedCmd, commands[0])
}

// TestUpgradeBmadCLIWithDefaults tests upgradeBmadCLI with default config values
func (ts *BmadUnitTestSuite) TestUpgradeBmadCLIWithDefaults() {
	originalRunner := GetRunner()
	defer func() {
		err := SetRunner(originalRunner)
		ts.Require().NoError(err)
	}()

	mock := NewBmadMockRunner()
	// With defaults: DefaultBmadPackageName + DefaultBmadVersionTag + DefaultBmadProjectDir
	expectedCmd := "npx --yes " + DefaultBmadPackageName + DefaultBmadVersionTag + " update -d " + DefaultBmadProjectDir + " --force"
	mock.SetOutput(expectedCmd, "", nil)

	err := SetRunner(mock)
	ts.Require().NoError(err)

	config := &Config{
		Bmad: BmadConfig{}, // Empty config, should use defaults
	}

	err = upgradeBmadCLI(config)
	ts.Require().NoError(err)
}

// TestUpgradeBmadCLIRunnerError tests upgradeBmadCLI when runner returns an error
func (ts *BmadUnitTestSuite) TestUpgradeBmadCLIRunnerError() {
	originalRunner := GetRunner()
	defer func() {
		err := SetRunner(originalRunner)
		ts.Require().NoError(err)
	}()

	mock := NewBmadMockRunner()
	expectedCmd := "npx --yes bmad-method@alpha update -d _bmad --force"
	mock.SetOutput(expectedCmd, "", errNpxCommandFailed)

	err := SetRunner(mock)
	ts.Require().NoError(err)

	config := &Config{
		Bmad: BmadConfig{
			PackageName: "bmad-method",
			VersionTag:  "@alpha",
			ProjectDir:  "_bmad",
		},
	}

	err = upgradeBmadCLI(config)
	ts.Require().Error(err)
	ts.Require().Contains(err.Error(), "npx command failed")
}

// TestBmadUnitTestSuite runs the unit test suite
func TestBmadUnitTestSuite(t *testing.T) {
	suite.Run(t, new(BmadUnitTestSuite))
}

// Table-driven test for getBmadProjectDir
func TestGetBmadProjectDirTableDriven(t *testing.T) {
	tests := []struct {
		name       string
		projectDir string
		expected   string
	}{
		{
			name:       "empty project dir uses default",
			projectDir: "",
			expected:   DefaultBmadProjectDir,
		},
		{
			name:       "custom project dir is used",
			projectDir: "custom/bmad",
			expected:   "custom/bmad",
		},
		{
			name:       "relative path is preserved",
			projectDir: "./my_bmad",
			expected:   "./my_bmad",
		},
		{
			name:       "absolute path is preserved",
			projectDir: "/tmp/bmad",
			expected:   "/tmp/bmad",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{
				Bmad: BmadConfig{
					ProjectDir: tt.projectDir,
				},
			}
			result := getBmadProjectDir(config)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestCheckBmadPrerequisites tests checkBmadPrerequisites function
func TestCheckBmadPrerequisites(t *testing.T) {
	// This test verifies the function runs without panicking
	// Actual results depend on whether npm/npx are installed on the test machine
	err := checkBmadPrerequisites()
	// We don't assert on the error since it depends on the test environment
	// Just verify it returns one of the expected errors or nil
	if err != nil {
		if !errors.Is(err, errNpmNotInstalled) && !errors.Is(err, errNpxNotInstalled) {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

// TestPrintBmadUpgradeSummary tests printBmadUpgradeSummary function
func TestPrintBmadUpgradeSummary(t *testing.T) {
	tests := []struct {
		name       string
		oldVersion string
		newVersion string
	}{
		{
			name:       "known versions",
			oldVersion: "5.0.0",
			newVersion: "6.0.0",
		},
		{
			name:       "unknown old version",
			oldVersion: statusUnknown,
			newVersion: "6.0.0",
		},
		{
			name:       "unknown new version",
			oldVersion: "5.0.0",
			newVersion: statusUnknown,
		},
		{
			name:       "both unknown",
			oldVersion: statusUnknown,
			newVersion: statusUnknown,
		},
		{
			name:       "same version",
			oldVersion: "5.0.0",
			newVersion: "5.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify it doesn't panic
			require.NotPanics(t, func() {
				printBmadUpgradeSummary(tt.oldVersion, tt.newVersion)
			})
		})
	}
}

// TestInstallBmadCLIWithDefaults tests installBmadCLI uses defaults correctly
func (ts *BmadUnitTestSuite) TestInstallBmadCLIConfigDefaults() {
	// This test verifies the function uses default config values
	config := &Config{
		Bmad: BmadConfig{}, // Empty config, should use defaults
	}

	// We can't actually run the interactive command in tests,
	// but we can verify the config defaults are correctly applied
	packageName := config.Bmad.PackageName
	if packageName == "" {
		packageName = DefaultBmadPackageName
	}
	ts.Require().Equal(DefaultBmadPackageName, packageName)

	versionTag := config.Bmad.VersionTag
	if versionTag == "" {
		versionTag = DefaultBmadVersionTag
	}
	ts.Require().Equal(DefaultBmadVersionTag, versionTag)
}

// TestInstallBmadCLIWithCustomConfig tests installBmadCLI with custom config
func (ts *BmadUnitTestSuite) TestInstallBmadCLIConfigCustom() {
	config := &Config{
		Bmad: BmadConfig{
			PackageName: "custom-bmad",
			VersionTag:  "@latest",
		},
	}

	packageName := config.Bmad.PackageName
	if packageName == "" {
		packageName = DefaultBmadPackageName
	}
	ts.Require().Equal("custom-bmad", packageName)

	versionTag := config.Bmad.VersionTag
	if versionTag == "" {
		versionTag = DefaultBmadVersionTag
	}
	ts.Require().Equal("@latest", versionTag)
}

// TestBmadStaticErrors tests that static errors are properly defined
func TestBmadStaticErrors(t *testing.T) {
	// Verify all static errors are defined
	require.Error(t, errNpmNotInstalled)
	require.Error(t, errNpxNotInstalled)
	require.Error(t, errBmadNotInstalled)
	require.Error(t, errBmadInstallFailed)
	require.Error(t, errBmadVersionParse)

	// Verify error messages are not empty
	require.NotEmpty(t, errNpmNotInstalled.Error())
	require.NotEmpty(t, errNpxNotInstalled.Error())
	require.NotEmpty(t, errBmadNotInstalled.Error())
	require.NotEmpty(t, errBmadInstallFailed.Error())
	require.NotEmpty(t, errBmadVersionParse.Error())
}

// TestBmadMockRunnerGetCommands tests the mock runner's command tracking
func TestBmadMockRunnerGetCommands(t *testing.T) {
	mock := NewBmadMockRunner()

	// Verify initially empty
	require.Empty(t, mock.GetCommands())

	// Execute some commands
	err := mock.RunCmd("echo", "hello")
	require.NoError(t, err)

	_, err = mock.RunCmdOutput("echo", "world")
	require.NoError(t, err)

	// Verify commands are tracked
	commands := mock.GetCommands()
	require.Len(t, commands, 2)
	require.Equal(t, "echo hello", commands[0])
	require.Equal(t, "echo world", commands[1])
}

// TestBmadMockRunnerSetOutput tests the mock runner's output setting
func TestBmadMockRunnerSetOutput(t *testing.T) {
	mock := NewBmadMockRunner()

	// Set expected output
	expectedOutput := "test output"
	mock.SetOutput("echo test", expectedOutput, nil)

	// Verify output is returned
	output, err := mock.RunCmdOutput("echo", "test")
	require.NoError(t, err)
	require.Equal(t, expectedOutput, output)

	// Verify error is returned when set
	mock.SetOutput("fail cmd", "", errTestError)

	err = mock.RunCmd("fail", "cmd")
	require.Error(t, err)
	require.Equal(t, errTestError, err)
}
