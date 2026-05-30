//go:build integration
// +build integration

package mage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// AgentOSMainTestSuite defines the test suite for AgentOS main methods
type AgentOSMainTestSuite struct {
	suite.Suite

	env     *testutil.TestEnvironment
	agentos AgentOS
}

// SetupTest runs before each test
func (ts *AgentOSMainTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.T().Setenv("GOTELEMETRY", "off")
	ts.agentos = AgentOS{}
}

// TearDownTest runs after each test
func (ts *AgentOSMainTestSuite) TearDownTest() {
	TestResetConfig()
	ts.env.Cleanup()
}

// TestInstall_AlreadyInstalled tests Install when project directory exists
func (ts *AgentOSMainTestSuite) TestInstall_AlreadyInstalled() {
	// Create a mock project directory
	projectDir := filepath.Join(ts.env.TempDir, DefaultAgentOSBaseDir)
	err := os.MkdirAll(projectDir, 0o750)
	ts.Require().NoError(err)

	defer chdirForTest(ts.T(), ts.env.TempDir)()

	// Create minimal config
	configContent := `
agentos:
  base_dir: ` + DefaultAgentOSBaseDir + `
`
	err = os.WriteFile(".mage.yaml", []byte(configContent), 0o600)
	ts.Require().NoError(err)

	// Attempt install
	err = ts.agentos.Install()

	// Should return errAgentOSAlreadyInstalled
	ts.Require().Error(err)
	ts.ErrorIs(err, errAgentOSAlreadyInstalled)
}

// TestCheck_NoBase tests Check when base is not installed
func (ts *AgentOSMainTestSuite) TestCheck_NoBase() {
	// Create config pointing to non-existent base
	configContent := `
agentos:
  home_dir: nonexistent-home
`
	configPath := filepath.Join(ts.env.TempDir, ".mage.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	ts.Require().NoError(err)

	defer chdirForTest(ts.T(), ts.env.TempDir)()

	// Override HOME to temp dir
	ts.T().Setenv("HOME", ts.env.TempDir)

	// Run check - should detect base not installed
	err = ts.agentos.Check()

	// Should return errAgentOSNotInstalled
	ts.Require().Error(err)
	ts.ErrorIs(err, errAgentOSNotInstalled)
}

// TestCheck_BaseInstalled tests Check with base installed
func (ts *AgentOSMainTestSuite) TestCheck_BaseInstalled() {
	// Create base directory and config
	homeDir := filepath.Join(ts.env.TempDir, ".agent-os")
	err := os.MkdirAll(homeDir, 0o750)
	ts.Require().NoError(err)

	configFile := filepath.Join(homeDir, DefaultAgentOSConfigFile)
	configContent := "version: v1.0.0\n"
	err = os.WriteFile(configFile, []byte(configContent), 0o600)
	ts.Require().NoError(err)

	// Create project directory
	projectDir := filepath.Join(ts.env.TempDir, DefaultAgentOSBaseDir)
	specsDir := filepath.Join(projectDir, "specs")
	err = os.MkdirAll(specsDir, 0o750)
	ts.Require().NoError(err)

	// Create mage config
	mageConfig := `
agentos:
  home_dir: .agent-os
  base_dir: ` + DefaultAgentOSBaseDir + `
  claude_code_commands: false
  use_claude_code_subagents: false
`
	err = os.WriteFile(filepath.Join(ts.env.TempDir, ".mage.yaml"), []byte(mageConfig), 0o600)
	ts.Require().NoError(err)

	defer chdirForTest(ts.T(), ts.env.TempDir)()

	// Override HOME
	ts.T().Setenv("HOME", ts.env.TempDir)

	// Run check - should succeed
	err = ts.agentos.Check()

	// Should succeed (base and project exist)
	ts.Require().NoError(err)
}

// TestUpgrade_NoBase tests Upgrade when base is not installed
func (ts *AgentOSMainTestSuite) TestUpgrade_NoBase() {
	// Create config pointing to non-existent base
	configContent := `
agentos:
  home_dir: nonexistent-upgrade-home
`
	configPath := filepath.Join(ts.env.TempDir, ".mage.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0o600)
	ts.Require().NoError(err)

	defer chdirForTest(ts.T(), ts.env.TempDir)()

	// Override HOME
	ts.T().Setenv("HOME", ts.env.TempDir)

	// Attempt upgrade
	err = ts.agentos.Upgrade()

	// Should return errAgentOSNotInstalled
	ts.Require().Error(err)
	ts.ErrorIs(err, errAgentOSNotInstalled)
}

// TestAgentOSPathHelpers tests path helper functions with edge cases
func (ts *AgentOSMainTestSuite) TestAgentOSPathHelpers() {
	ts.Run("home path with long custom directory", func() {
		config := &Config{
			AgentOS: AgentOSConfig{
				HomeDir: "very/long/custom/path/to/agent-os/installation",
			},
		}

		path := getAgentOSHomePath(config)
		ts.Contains(path, "very/long/custom/path/to/agent-os/installation")
	})

	ts.Run("project dir with relative path", func() {
		config := &Config{
			AgentOS: AgentOSConfig{
				BaseDir: "./relative/path/to/project",
			},
		}

		path := getAgentOSProjectDir(config)
		ts.Equal("./relative/path/to/project", path)
	})

	ts.Run("project dir with tilde expansion", func() {
		config := &Config{
			AgentOS: AgentOSConfig{
				BaseDir: "~/agent-os-custom",
			},
		}

		path := getAgentOSProjectDir(config)
		ts.Equal("~/agent-os-custom", path)
	})
}

// TestAgentOSConfigVariations tests various configuration scenarios
func (ts *AgentOSMainTestSuite) TestAgentOSConfigVariations() {
	tests := []struct {
		name              string
		profile           string
		claudeCodeCmds    bool
		agentOSCmds       bool
		subagents         bool
		standardsAsSkills bool
		expectedArgsCount int
	}{
		{
			name:              "minimal config",
			profile:           "",
			claudeCodeCmds:    true,
			agentOSCmds:       false,
			subagents:         true,
			standardsAsSkills: false,
			expectedArgsCount: 0,
		},
		{
			name:              "all features enabled",
			profile:           "advanced",
			claudeCodeCmds:    true,
			agentOSCmds:       true,
			subagents:         true,
			standardsAsSkills: true,
			expectedArgsCount: 4,
		},
		{
			name:              "all features disabled",
			profile:           "",
			claudeCodeCmds:    false,
			agentOSCmds:       false,
			subagents:         false,
			standardsAsSkills: false,
			expectedArgsCount: 2,
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			config := &Config{
				AgentOS: AgentOSConfig{
					Profile:                tt.profile,
					ClaudeCodeCommands:     tt.claudeCodeCmds,
					AgentOSCommands:        tt.agentOSCmds,
					UseClaudeCodeSubagents: tt.subagents,
					StandardsAsSkills:      tt.standardsAsSkills,
				},
			}

			args := buildAgentOSInstallArgs(config)

			// Verify arg count matches expectations
			if tt.profile != "" && tt.profile != "default" {
				ts.GreaterOrEqual(len(args), 2, "should have profile args")
			}

			if !tt.claudeCodeCmds {
				ts.Contains(args, "--no-claude-code-commands")
			}

			if tt.agentOSCmds {
				ts.Contains(args, "--agent-os-commands")
			}

			if !tt.subagents {
				ts.Contains(args, "--no-subagents")
			}

			if tt.standardsAsSkills {
				ts.Contains(args, "--standards-as-skills")
			}
		})
	}
}

// TestAgentOSVersionParsing tests version string parsing edge cases
func (ts *AgentOSMainTestSuite) TestAgentOSVersionParsing() {
	tests := []struct {
		name        string
		yamlContent string
		wantVersion string
		wantError   bool
	}{
		{
			name:        "semantic version",
			yamlContent: "version: v1.2.3\n",
			wantVersion: "v1.2.3",
			wantError:   false,
		},
		{
			name:        "pre-release version",
			yamlContent: "version: v2.0.0-alpha.1\n",
			wantVersion: "v2.0.0-alpha.1",
			wantError:   false,
		},
		{
			name:        "build metadata",
			yamlContent: "version: v1.0.0+build.123\n",
			wantVersion: "v1.0.0+build.123",
			wantError:   false,
		},
		{
			name:        "version with extra spaces",
			yamlContent: "version:    v1.0.0   \n",
			wantVersion: "v1.0.0",
			wantError:   false,
		},
		{
			name:        "multiline yaml",
			yamlContent: "name: agent-os\nversion: v3.0.0\nauthor: test\n",
			wantVersion: "v3.0.0",
			wantError:   false,
		},
		{
			name:        "invalid yaml",
			yamlContent: "invalid: yaml: content: [unclosed\n",
			wantVersion: "",
			wantError:   true,
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			// Create temp home with config
			homeDir := filepath.Join(ts.env.TempDir, ".agent-os-"+tt.name)
			err := os.MkdirAll(homeDir, 0o750)
			ts.Require().NoError(err)

			configFile := filepath.Join(homeDir, DefaultAgentOSConfigFile)
			err = os.WriteFile(configFile, []byte(tt.yamlContent), 0o600)
			ts.Require().NoError(err)

			config := &Config{
				AgentOS: AgentOSConfig{
					HomeDir: filepath.Base(homeDir),
				},
			}

			ts.T().Setenv("HOME", ts.env.TempDir)

			version, err := getAgentOSVersion(config)

			if tt.wantError {
				ts.Require().Error(err)
			} else {
				ts.Require().NoError(err)
				ts.Equal(tt.wantVersion, version)
			}
		})
	}
}

// TestAgentOSMainTestSuite runs the test suite
func TestAgentOSMainTestSuite(t *testing.T) {
	suite.Run(t, new(AgentOSMainTestSuite))
}
