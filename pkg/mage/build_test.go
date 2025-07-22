package mage

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mrz1836/go-mage/pkg/mage/testutil"
	"github.com/mrz1836/go-mage/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Import the new testutil package for consistent mocking
// MockRunner is now imported from testutil package

// BuildTestSuite defines the test suite for build functions
type BuildTestSuite struct {
	suite.Suite
	env   *testutil.TestEnvironment
	build Build
}

// SetupTest runs before each test
func (suite *BuildTestSuite) SetupTest() {
	// Create test environment
	suite.env = testutil.NewTestEnvironment(suite.T())

	// Create test project structure
	suite.createTestProject()

	// Initialize build
	suite.build = Build{}

	// Reset configuration
	cfg = nil
	
	// Reset cache manager to ensure clean state for each test
	cacheManager = nil
}

// TearDownTest runs after each test
func (suite *BuildTestSuite) TearDownTest() {
	// Clean up test environment
	suite.env.Cleanup()

	// Assert all mock expectations were met
	suite.env.Runner.AssertExpectations(suite.T())
}

// createTestProject creates a minimal test project structure
func (suite *BuildTestSuite) createTestProject() {
	// Create project structure
	suite.env.CreateProjectStructure()

	// Create go.mod
	suite.env.CreateGoMod("github.com/test/project")

	// Create .mage.yaml
	mageYaml := `project:
  name: testproject
  binary: testapp
  version: v1.0.0

build:
  output: bin
  trimpath: true
  platforms:
    - linux/amd64
    - darwin/amd64
    - windows/amd64
`
	suite.env.CreateMageConfig(mageYaml)
}

// TestBuildDefault tests the default build function
func (suite *BuildTestSuite) TestBuildDefault() {
	tests := []struct {
		name        string
		platform    string
		setupMock   func()
		expectError bool
		errorMsg    string
	}{
		{
			name:     "successful build on linux",
			platform: "linux",
			setupMock: func() {
				// Mock git commands for version info
				suite.env.Runner.On("RunCmdOutput", "git", mock.MatchedBy(func(args []string) bool {
					return len(args) > 0 && (args[0] == "describe" || args[0] == "rev-parse")
				})).Return("v1.0.0", nil).Maybe()

				// Mock successful go build command
				suite.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
					return args[0] == "build" && containsString(args, "-o")
				})).Return(nil)
			},
			expectError: false,
		},
		{
			name:     "successful build on windows",
			platform: "windows",
			setupMock: func() {
				// Mock git commands for version info
				suite.env.Runner.On("RunCmdOutput", "git", mock.MatchedBy(func(args []string) bool {
					return len(args) > 0 && (args[0] == "describe" || args[0] == "rev-parse")
				})).Return("v1.0.0", nil).Maybe()

				suite.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
					return args[0] == "build" && containsString(args, "bin/testapp.exe")
				})).Return(nil)
			},
			expectError: false,
		},
		{
			name:     "build failure",
			platform: "linux",
			setupMock: func() {
				// Reset all expectations for failure test
				suite.env.Runner.ExpectedCalls = nil
				
				// Mock git commands for version info
				suite.env.Runner.On("RunCmdOutput", "git", mock.MatchedBy(func(args []string) bool {
					return len(args) > 0 && (args[0] == "describe" || args[0] == "rev-parse")
				})).Return("v1.0.0", nil).Maybe()

				// Mock failed go build command - be more specific to avoid conflicts
				suite.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
					return len(args) > 0 && args[0] == "build" && containsString(args, "-o") && containsString(args, "bin/testapp")
				})).Return(fmt.Errorf("compilation error")).Once()
			},
			expectError: true,
			errorMsg:    "build failed",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Set platform via environment variable instead of runtime.GOOS
			if tt.platform == "windows" {
				os.Setenv("GOOS", "windows")
				defer func() { os.Unsetenv("GOOS") }()
			}

			// Setup mock expectations
			if tt.setupMock != nil {
				tt.setupMock()
			}

			// Execute build with mocked runner
			err := suite.executeWithMockRunner(func() error {
				return suite.build.Default()
			})

			// Assert results
			if tt.expectError {
				suite.Error(err)
				if tt.errorMsg != "" {
					suite.Contains(err.Error(), tt.errorMsg)
				}
			} else {
				suite.NoError(err)
			}
		})
	}
}

// TestBuildPlatform tests platform-specific builds
func (suite *BuildTestSuite) TestBuildPlatform() {
	testCases := []struct {
		name         string
		platform     string
		expectedOS   string
		expectedArch string
		expectBinary string
	}{
		{
			name:         "linux amd64",
			platform:     "linux/amd64",
			expectedOS:   "linux",
			expectedArch: "amd64",
			expectBinary: "testapp-linux-amd64",
		},
		{
			name:         "darwin arm64",
			platform:     "darwin/arm64",
			expectedOS:   "darwin",
			expectedArch: "arm64",
			expectBinary: "testapp-darwin-arm64",
		},
		{
			name:         "windows amd64",
			platform:     "windows/amd64",
			expectedOS:   "windows",
			expectedArch: "amd64",
			expectBinary: "testapp-windows-amd64.exe",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Reset expectations and calls for clean test
			suite.env.Runner.ExpectedCalls = nil
			suite.env.Runner.Calls = nil
			
			// Create a spy to capture environment variables
			var capturedEnv []string
			var buildCalls []*mock.Call

			// Mock git commands for version info
			suite.env.Runner.On("RunCmdOutput", "git", mock.MatchedBy(func(args []string) bool {
				return len(args) > 0 && (args[0] == "describe" || args[0] == "rev-parse")
			})).Return("v1.0.0", nil).Maybe()

			// Mock successful build - be specific to only match go build commands
			suite.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
				return len(args) > 0 && args[0] == "build"
			})).Run(func(args mock.Arguments) {
				// Capture environment when build command is called
				capturedEnv = os.Environ()
				// Track this specific call
				buildCalls = append(buildCalls, &mock.Call{})
			}).Return(nil).Once()

			// Execute build
			err := suite.executeWithMockRunner(func() error {
				return suite.build.Platform(tc.platform)
			})

			// Assert no error
			suite.NoError(err)

			// Verify environment variables were set correctly
			suite.Contains(getEnvValue(capturedEnv, "GOOS"), tc.expectedOS)
			suite.Contains(getEnvValue(capturedEnv, "GOARCH"), tc.expectedArch)

			// Verify exactly one build call was made
			suite.Require().Len(buildCalls, 1)

			// Find the build command call and verify binary name
			var buildArgs []string
			for _, call := range suite.env.Runner.Calls {
				if call.Method == "RunCmd" && len(call.Arguments) > 1 {
					if cmd, ok := call.Arguments[0].(string); ok && cmd == "go" {
						if args, ok := call.Arguments[1].([]string); ok && len(args) > 0 && args[0] == "build" {
							buildArgs = args
							break
						}
					}
				}
			}
			suite.Require().NotEmpty(buildArgs, "No go build command found")
			suite.Contains(strings.Join(buildArgs, " "), tc.expectBinary)
		})
	}
}

// TestBuildAll tests building for all platforms
func (suite *BuildTestSuite) TestBuildAll() {
	// Set up mocks for build operations - allow any number of calls
	suite.env.StandardMocks().ForBuild()

	// Mock successful builds for all platforms - allow any number since we might have git calls too
	suite.env.Runner.On("RunCmd", "go", mock.Anything).Return(nil).Maybe()

	// Execute build all
	err := suite.env.WithMockRunner(
		func(r interface{}) { SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return suite.build.All()
		},
	)

	// Assert success
	suite.NoError(err)
}

// TestBuildClean tests the clean function
func (suite *BuildTestSuite) TestBuildClean() {
	// Create some build artifacts
	testFiles := []string{
		"bin/testapp",
		"bin/testapp-linux-amd64",
		"coverage.txt",
	}

	for _, file := range testFiles {
		dir := filepath.Dir(file)
		if dir != "." {
			os.MkdirAll(dir, 0755)
		}
		os.WriteFile(file, []byte("test"), 0644)
	}

	// Mock go clean command
	suite.env.Runner.On("RunCmd", "go", []string{"clean", "-testcache"}).Return(nil)

	// Execute clean
	err := suite.executeWithMockRunner(func() error {
		return suite.build.Clean()
	})

	// Assert success
	suite.NoError(err)

	// Verify bin directory was cleaned
	_, err = os.Stat("bin/testapp")
	suite.True(os.IsNotExist(err))
}

// TestBuildFlags tests build flag generation
func (suite *BuildTestSuite) TestBuildFlags() {
	testCases := []struct {
		name     string
		config   BuildConfig
		expected []string
	}{
		{
			name: "with tags",
			config: BuildConfig{
				Tags: []string{"prod", "feature1"},
			},
			expected: []string{"-tags", "prod,feature1"},
		},
		{
			name: "with custom ldflags",
			config: BuildConfig{
				LDFlags: []string{"-X main.version=1.0.0", "-s -w"},
			},
			expected: []string{"-ldflags", "-X main.version=1.0.0 -s -w"},
		},
		{
			name: "with trimpath",
			config: BuildConfig{
				TrimPath: true,
			},
			expected: []string{"-trimpath"},
		},
		{
			name:     "debug mode",
			config:   BuildConfig{},
			expected: []string{"-ldflags"}, // Should not contain -s -w
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Override config
			cfg = &Config{
				Build: tc.config,
			}

			// Get build flags
			flags := buildFlags(cfg)

			// Assert expected flags are present
			for _, expected := range tc.expected {
				suite.Contains(flags, expected)
			}

			// Special case: debug mode should not strip
			if tc.name == "debug mode" {
				os.Setenv("DEBUG", "true")
				defer os.Unsetenv("DEBUG")

				flags = buildFlags(cfg)
				flagStr := strings.Join(flags, " ")
				suite.NotContains(flagStr, "-s")
				suite.NotContains(flagStr, "-w")
			}
		})
	}
}

// TestVersionAndCommit tests version and commit detection
func (suite *BuildTestSuite) TestVersionAndCommit() {
	suite.Run("version from git tag", func() {
		// Mock git command
		suite.env.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).
			Return("v1.2.3\n", nil)

		version := suite.executeWithMockOutput(getVersion)
		suite.Equal("v1.2.3", version)
	})

	suite.Run("version from config", func() {
		// Reset all expectations
		suite.env.Runner.ExpectedCalls = nil
		
		// Mock git command failure
		suite.env.Runner.On("RunCmdOutput", "git", mock.Anything).
			Return("", fmt.Errorf("not a git repo"))

		// Set config version
		cfg = &Config{
			Project: ProjectConfig{Version: "v2.0.0"},
		}

		version := suite.executeWithMockOutput(getVersion)
		suite.Equal("v2.0.0", version)
	})

	suite.Run("commit hash", func() {
		// Reset all expectations
		suite.env.Runner.ExpectedCalls = nil
		
		// Mock git command
		suite.env.Runner.On("RunCmdOutput", "git", []string{"rev-parse", "--short", "HEAD"}).
			Return("abc123\n", nil)

		commit := suite.executeWithMockOutput(getCommit)
		suite.Equal("abc123", commit)
	})
}

// Helper functions

// executeWithMockRunner temporarily replaces the command runner with a mock
// Deprecated: Use env.WithMockRunner instead
func (suite *BuildTestSuite) executeWithMockRunner(fn func() error) error {
	return suite.env.WithMockRunner(
		func(r interface{}) { SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		fn,
	)
}

// executeWithMockOutput executes a function that returns a string
// Deprecated: Use env.WithMockRunner with appropriate return handling
func (suite *BuildTestSuite) executeWithMockOutput(fn func() string) string {
	var result string
	suite.env.WithMockRunner(
		func(r interface{}) { SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			result = fn()
			return nil
		},
	)
	return result
}

// containsString checks if a slice contains a value
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getEnvValue gets a value from environment slice
func getEnvValue(env []string, key string) string {
	prefix := key + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return strings.TrimPrefix(e, prefix)
		}
	}
	return ""
}

// getOriginalOS returns the original OS
func (suite *BuildTestSuite) getOriginalOS() string {
	return runtime.GOOS
}

// TestBuildTestSuite runs the build test suite
func TestBuildTestSuite(t *testing.T) {
	suite.Run(t, new(BuildTestSuite))
}

// Additional table-driven tests for specific functions

func TestParsePlatform(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantOS   string
		wantArch string
		wantErr  bool
	}{
		{
			name:     "valid linux/amd64",
			input:    "linux/amd64",
			wantOS:   "linux",
			wantArch: "amd64",
			wantErr:  false,
		},
		{
			name:     "valid darwin/arm64",
			input:    "darwin/arm64",
			wantOS:   "darwin",
			wantArch: "arm64",
			wantErr:  false,
		},
		{
			name:    "invalid format",
			input:   "linux-amd64",
			wantErr: true,
		},
		{
			name:    "missing arch",
			input:   "linux",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			platform, err := utils.ParsePlatform(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantOS, platform.OS)
				assert.Equal(t, tt.wantArch, platform.Arch)
			}
		})
	}
}
