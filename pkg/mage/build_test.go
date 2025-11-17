//go:build integration
// +build integration

package mage

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// Static test errors to satisfy err113 linter
var (
	errBuildError    = errors.New("build error")
	errGitErrorBuild = errors.New("git error")
)

// BuildTestSuite defines the test suite for Build namespace methods
type BuildTestSuite struct {
	suite.Suite

	env   *testutil.TestEnvironment
	build Build
}

// SetupTest runs before each test
func (ts *BuildTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.build = Build{}
	// Disable cache for tests
	if err := os.Setenv("MAGE_X_CACHE_DISABLED", "true"); err != nil {
		ts.T().Fatalf("Failed to set MAGE_X_CACHE_DISABLED: %v", err)
	}
}

// TearDownTest runs after each test
func (ts *BuildTestSuite) TearDownTest() {
	// Clean up environment variables that might be set by tests
	if err := os.Unsetenv("GOOS"); err != nil {
		ts.T().Logf("Failed to unset GOOS: %v", err)
	}
	if err := os.Unsetenv("GOARCH"); err != nil {
		ts.T().Logf("Failed to unset GOARCH: %v", err)
	}
	if err := os.Unsetenv("CGO_ENABLED"); err != nil {
		ts.T().Logf("Failed to unset CGO_ENABLED: %v", err)
	}
	if err := os.Unsetenv("MAGE_X_CACHE_DISABLED"); err != nil {
		ts.T().Logf("Failed to unset MAGE_X_CACHE_DISABLED: %v", err)
	}

	// Reset global config
	TestResetConfig()

	ts.env.Cleanup()
}

// mockGitCommands adds mock expectations for git commands used by buildFlags
func (ts *BuildTestSuite) mockGitCommands() {
	ts.env.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).Return("v1.0.0", nil)
	ts.env.Runner.On("RunCmdOutput", "git", []string{"rev-parse", "--short", "HEAD"}).Return("abc1234", nil)
}

// mockBuildCommand mocks a go build command with flexible ldflags matching
func (ts *BuildTestSuite) mockBuildCommand(outputPath string) {
	ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		// Check for go build command with variable structure
		// Could be: ["build", "-ldflags", <flags>, "-o", <output>, "."]
		// Or: ["build", "-ldflags", <flags>, "-trimpath", "-o", <output>, "."]
		if len(args) < 6 || args[0] != CmdGoBuild || args[1] != "-ldflags" {
			return false
		}

		// Find the -o flag and check the output path
		for i := 3; i < len(args)-1; i++ {
			if args[i] == "-o" && i+1 < len(args) {
				return args[i+1] == outputPath
			}
		}
		return false
	})).Return(nil)
}

// TestBuildDefault tests the Default method
func (ts *BuildTestSuite) TestBuildDefault() {
	ts.Run("successful default build", func() {
		// Create basic project structure
		ts.env.CreateFile("main.go", `package main
func main() {
	println("Hello, World!")
}`)

		// Mock git commands and build command
		ts.mockGitCommands()
		ts.mockBuildCommand("bin/module")

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.Default()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("handles build failure", func() {
		// Create a fresh test environment for this sub-test
		env := testutil.NewTestEnvironment(ts.T())
		defer env.Cleanup()

		// Create project file
		env.CreateFile("main.go", `package main
func main() {
	println("Test app")
}`)
		env.CreateGoMod("test/module")

		// Mock git commands and failed build command
		env.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).Return("v1.0.0", nil)
		env.Runner.On("RunCmdOutput", "git", []string{"rev-parse", "--short", "HEAD"}).Return("abc1234", nil)
		env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
			if len(args) < 6 || args[0] != CmdGoBuild || args[1] != "-ldflags" {
				return false
			}
			for i := 3; i < len(args)-1; i++ {
				if args[i] == "-o" && i+1 < len(args) {
					return args[i+1] == "bin/module"
				}
			}
			return false
		})).Return(errBuildError)

		err := env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.Default()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "build failed")
	})
}

// TestBuildAll tests the All method
func (ts *BuildTestSuite) TestBuildAll() {
	ts.Run("builds for all configured platforms", func() {
		// Create configuration with multiple platforms
		ts.env.CreateFile(".mage.yaml", `
project:
  binary: multi-app
build:
  platforms: [linux/amd64, darwin/amd64, windows/amd64]
  output: dist
`)

		// Create project file
		ts.env.CreateFile("main.go", `package main
func main() {
	println("Multi-platform app")
}`)

		// Mock git commands and multi-platform build commands
		ts.mockGitCommands()
		ts.mockBuildCommand("dist/multi-app-linux-amd64")
		ts.mockBuildCommand("dist/multi-app-darwin-amd64")
		ts.mockBuildCommand("dist/multi-app-windows-amd64.exe")

		TestResetConfig() // Reset global config
		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.All()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("handles no platforms configured", func() {
		// Create minimal configuration
		ts.env.CreateFile(".mage.yaml", `
project:
  binary: no-platform-app
`)

		// Create project file
		ts.env.CreateFile("main.go", `package main
func main() {
	println("No platform app")
}`)

		// Mock git commands and build commands for default platforms
		ts.mockGitCommands()
		// When no platforms are configured, it builds for default platforms
		ts.mockBuildCommand("bin/no-platform-app-windows-amd64.exe")
		ts.mockBuildCommand("bin/no-platform-app-darwin-arm64")
		ts.mockBuildCommand("bin/no-platform-app-darwin-amd64")
		ts.mockBuildCommand("bin/no-platform-app-linux-amd64")

		TestResetConfig() // Reset global config
		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.All()
			},
		)

		ts.Require().NoError(err) // Should handle gracefully
	})
}

// TestBuildPlatform tests the Platform method
func (ts *BuildTestSuite) TestBuildPlatform() {
	ts.Run("builds for specific platform", func() {
		// Create project file
		ts.env.CreateFile("main.go", `package main
func main() {
	println("Platform app")
}`)

		// Mock git commands used by buildFlags
		ts.mockGitCommands()

		// Mock go build command for linux/amd64
		ts.mockBuildCommand("bin/module-linux-amd64")

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.Platform("linux/amd64")
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("handles invalid platform format", func() {
		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.Platform("invalid-platform")
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "invalid platform format")
	})

	ts.Run("builds for windows platform with .exe extension", func() {
		// Create project file
		ts.env.CreateFile("main.go", `package main
func main() {
	println("Windows app")
}`)

		// Mock git commands used by buildFlags
		ts.mockGitCommands()

		// Mock go build command for windows
		ts.mockBuildCommand("bin/module-windows-amd64.exe")

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.Platform("windows/amd64")
			},
		)

		ts.Require().NoError(err)
	})
}

// TestBuildPlatformSpecific tests platform-specific methods
func (ts *BuildTestSuite) TestBuildPlatformSpecific() {
	ts.Run("Linux method", func() {
		// Create project file
		ts.env.CreateFile("main.go", `package main
func main() {
	println("Linux app")
}`)

		// Mock git commands used by buildFlags
		ts.mockGitCommands()

		// Mock go build command for Linux
		ts.mockBuildCommand("bin/module-linux-amd64")

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.Linux()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("Darwin method", func() {
		// Create project file
		ts.env.CreateFile("main.go", `package main
func main() {
	println("Darwin app")
}`)

		// Mock git commands used by buildFlags
		ts.mockGitCommands()

		// Mock go build commands for Darwin (both amd64 and arm64)
		ts.mockBuildCommand("bin/module-darwin-amd64")
		ts.mockBuildCommand("bin/module-darwin-arm64")

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.Darwin()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("Windows method", func() {
		// Create project file
		ts.env.CreateFile("main.go", `package main
func main() {
	println("Windows app")
}`)

		// Mock git commands used by buildFlags
		ts.mockGitCommands()

		// Mock go build command for Windows
		ts.mockBuildCommand("bin/module-windows-amd64.exe")

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.Windows()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestBuildClean tests the Clean method
func (ts *BuildTestSuite) TestBuildClean() {
	ts.Run("cleans build artifacts", func() {
		// Create some build artifacts
		ts.env.CreateFile("bin/app", "fake binary")
		ts.env.CreateFile("dist/app-linux", "fake linux binary")
		ts.env.CreateFile("coverage.out", "fake coverage")

		// Mock clean commands
		ts.env.Runner.On("RunCmd", "go", []string{"clean", "-testcache"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.Clean()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestBuildInstall tests the Install method
func (ts *BuildTestSuite) TestBuildInstall() {
	ts.Run("installs binary to GOPATH/bin", func() {
		// Create project file
		ts.env.CreateFile("main.go", `package main
func main() {
	println("Install app")
}`)

		// Mock git commands needed for version info
		ts.mockGitCommands()

		// Mock go install command with flexible args matching
		ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
			return len(args) >= 2 && args[0] == "install" && args[len(args)-1] == "."
		})).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.Install()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestBuildDev tests the Dev method
func (ts *BuildTestSuite) TestBuildDev() {
	ts.Run("builds and installs with forced dev version", func() {
		// Create project file
		ts.env.CreateFile("main.go", `package main
func main() {
	println("Dev app")
}`)

		// Mock git commands needed for version info
		ts.mockGitCommands()

		// Mock go install command with flexible args matching
		ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
			return len(args) >= 2 && args[0] == "install" && args[len(args)-1] == "."
		})).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.Dev()
			},
		)

		ts.Require().NoError(err)

		// Verify that MAGE_X_VERSION environment variable was cleaned up
		ts.Require().Empty(os.Getenv("MAGE_X_VERSION"))
	})
}

// TestBuildGenerate tests the Generate method
func (ts *BuildTestSuite) TestBuildGenerate() {
	ts.Run("runs go generate", func() {
		// Create file with generate directive
		ts.env.CreateFile("generate.go", `package main
//go:generate echo "Generated code"
func main() {}`)

		// Mock go generate command
		ts.env.Runner.On("RunCmd", "go", []string{"generate", "./..."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.Generate()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestBuildPreBuild tests the PreBuild method
func (ts *BuildTestSuite) TestBuildPreBuild() {
	ts.Run("runs pre-build tasks without parallel flag", func() {
		// Mock go build with flexible args (no parallel flag expected)
		ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
			// Should be ["build", "./..."] or ["build", "-v", "./..."]
			if len(args) < 2 || args[0] != "build" || args[len(args)-1] != "./..." {
				return false
			}
			// Should NOT contain -p flag
			for i, arg := range args {
				if arg == "-p" && i+1 < len(args) {
					return false // Found -p flag, which shouldn't be there
				}
			}
			return true
		})).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.PreBuild()
			},
		)
		ts.Require().NoError(err)
	})
}

// TestBuildPreBuildWithArgs tests the PreBuildWithArgs method
func (ts *BuildTestSuite) TestBuildPreBuildWithArgs() {
	ts.Run("runs pre-build with parallel=2", func() {
		// Set up os.Args to simulate command line arguments
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()
		os.Args = []string{"magex", "build:prebuild", "parallel=2"}

		// Mock go build with -p 2 flag
		ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
			// Should contain ["build", "-p", "2", "./..."]
			expectedArgs := []string{"build", "-p", "2", "./..."}
			if len(args) != len(expectedArgs) {
				return false
			}
			for i, expected := range expectedArgs {
				if args[i] != expected {
					return false
				}
			}
			return true
		})).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.PreBuildWithArgs()
			},
		)
		ts.Require().NoError(err)
	})

	ts.Run("runs pre-build with p=4 (short form)", func() {
		// Set up os.Args to simulate command line arguments
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()
		os.Args = []string{"magex", "build:prebuild", "p=4"}

		// Mock go build with -p 4 flag
		ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
			// Should contain ["build", "-p", "4", "./..."]
			expectedArgs := []string{"build", "-p", "4", "./..."}
			if len(args) != len(expectedArgs) {
				return false
			}
			for i, expected := range expectedArgs {
				if args[i] != expected {
					return false
				}
			}
			return true
		})).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.PreBuildWithArgs()
			},
		)
		ts.Require().NoError(err)
	})

	ts.Run("runs pre-build without parallel flag when not specified", func() {
		// Set up os.Args to simulate command line arguments without parallel flag
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()
		os.Args = []string{"magex", "build:prebuild"}

		// Mock go build without -p flag
		ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
			// Should be ["build", "./..."] or ["build", "-v", "./..."]
			if len(args) < 2 || args[0] != "build" || args[len(args)-1] != "./..." {
				return false
			}
			// Should NOT contain -p flag
			for i, arg := range args {
				if arg == "-p" && i+1 < len(args) {
					return false // Found -p flag, which shouldn't be there
				}
			}
			return true
		})).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.PreBuildWithArgs()
			},
		)
		ts.Require().NoError(err)
	})

	ts.Run("runs pre-build with verbose and parallel flags", func() {
		// Set up os.Args and configuration
		originalArgs := os.Args
		defer func() { os.Args = originalArgs }()
		os.Args = []string{"magex", "build:prebuild", "parallel=1"}

		// Override config to enable verbose
		config := defaultConfig()
		config.Build.Verbose = true
		TestSetConfig(config)

		// Mock go build with -v and -p 1 flags
		ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
			// Should contain ["build", "-v", "-p", "1", "./..."]
			expectedArgs := []string{"build", "-v", "-p", "1", "./..."}
			if len(args) != len(expectedArgs) {
				return false
			}
			for i, expected := range expectedArgs {
				if args[i] != expected {
					return false
				}
			}
			return true
		})).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.PreBuildWithArgs()
			},
		)
		ts.Require().NoError(err)
	})
}

// TestBuildUtilityFunctions tests utility functions
func (ts *BuildTestSuite) TestBuildUtilityFunctions() {
	ts.Run("buildFlags generates correct flags", func() {
		config := &Config{
			Build: BuildConfig{
				Tags:     []string{"integration", "e2e"},
				LDFlags:  []string{"-X main.version=1.0.0", "-X main.build=123"},
				TrimPath: true,
				Verbose:  true,
			},
		}

		flags := buildFlags(config)

		ts.Require().Contains(flags, "-tags")
		ts.Require().Contains(flags, "integration,e2e")
		ts.Require().Contains(flags, "-ldflags")
		ts.Require().Contains(flags, "-trimpath")
		ts.Require().Contains(flags, "-v")
	})

	ts.Run("buildFlags with minimal config", func() {
		config := &Config{
			Build: BuildConfig{},
		}

		flags := buildFlags(config)

		ts.Require().Contains(flags, "-ldflags")
		ts.Require().NotContains(flags, "-tags")
		ts.Require().NotContains(flags, "-trimpath")
		ts.Require().NotContains(flags, "-v")
	})

	ts.Run("getCommit returns commit hash", func() {
		// Mock git command to return a commit hash
		originalRunner := GetRunner()
		mockRunner := &testutil.MockRunner{}
		mockRunner.On("RunCmdOutput", "git", []string{"rev-parse", "--short", "HEAD"}).Return("abc1234", nil)
		ts.Require().NoError(SetRunner(mockRunner))

		commit := getCommit()

		ts.Require().Equal("abc1234", commit)

		// Restore original runner
		ts.Require().NoError(SetRunner(originalRunner))
	})

	ts.Run("getCommit handles git error", func() {
		// Mock git command to return error
		originalRunner := GetRunner()
		mockRunner := &testutil.MockRunner{}
		mockRunner.On("RunCmdOutput", "git", []string{"rev-parse", "--short", "HEAD"}).Return("", errGitErrorBuild)
		ts.Require().NoError(SetRunner(mockRunner))

		commit := getCommit()

		ts.Require().Equal("unknown", commit)

		// Restore original runner
		ts.Require().NoError(SetRunner(originalRunner))
	})
}

// TestBuildStrategies tests the new build strategy methods
func (ts *BuildTestSuite) TestBuildStrategies() {
	ts.Run("incremental strategy", func() {
		// Mock package listing
		ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "./..."}).Return(
			"github.com/test/pkg1\ngithub.com/test/pkg2\ngithub.com/test/pkg3", nil)

		// Mock batch builds - expect 2 batches with batch size 2
		ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
			// First batch: ["build", "-p", "1", "github.com/test/pkg1", "github.com/test/pkg2"]
			return len(args) >= 4 && args[0] == "build" &&
				strings.Contains(strings.Join(args, " "), "github.com/test/pkg1")
		})).Return(nil).Once()

		ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
			// Second batch: ["build", "-p", "1", "github.com/test/pkg3"]
			return len(args) >= 3 && args[0] == "build" &&
				strings.Contains(strings.Join(args, " "), "github.com/test/pkg3")
		})).Return(nil).Once()

		err := utils.WithIsolatedRunner(
			ts.env.Runner,
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.buildIncremental(2, 0, "", false, "1")
			},
		)
		ts.Require().NoError(err)
	})

	ts.Run("mains-first strategy", func() {
		// Mock finding main packages
		ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-json", "./..."}).Return(
			`{"ImportPath":"github.com/test/cmd/app","Name":"main"}
{"ImportPath":"github.com/test/pkg1","Name":"pkg1"}`, nil)

		// Mock listing all packages
		ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "./..."}).Return(
			"github.com/test/cmd/app\ngithub.com/test/pkg1", nil)

		// Expect main package to be built first
		ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
			return len(args) >= 2 && args[0] == "build" &&
				strings.Contains(strings.Join(args, " "), "github.com/test/cmd/app")
		})).Return(nil).Once()

		// Then remaining packages
		ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
			return len(args) >= 2 && args[0] == "build" &&
				strings.Contains(strings.Join(args, " "), "github.com/test/pkg1")
		})).Return(nil).Once()

		err := utils.WithIsolatedRunner(
			ts.env.Runner,
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.buildMainsFirst(10, false, "", false, "")
			},
		)
		ts.Require().NoError(err)
	})

	ts.Run("smart strategy selects appropriate method", func() {
		// Mock package listing for smart strategy
		ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "./..."}).Return(
			"github.com/test/pkg1\ngithub.com/test/pkg2", nil)

		// Mock build command - smart strategy should choose based on memory
		ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
			return args[0] == "build"
		})).Return(nil)

		err := utils.WithIsolatedRunner(
			ts.env.Runner,
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup
			func() interface{} { return GetRunner() },
			func() error {
				// Smart strategy will select based on available memory
				return ts.build.buildSmart("", false, "")
			},
		)
		ts.Require().NoError(err)
	})

	ts.Run("full strategy uses traditional build", func() {
		// Mock traditional full build
		ts.env.Runner.On("RunCmd", "go", []string{"build", "-p", "4", "./..."}).Return(nil)

		err := utils.WithIsolatedRunner(
			ts.env.Runner,
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.buildFull("4", false)
			},
		)
		ts.Require().NoError(err)
	})
}

// TestPackageDiscoveryUtilities tests package discovery and batching utilities
func (ts *BuildTestSuite) TestPackageDiscoveryUtilities() {
	ts.Run("discoverPackages lists all packages", func() {
		ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "./..."}).Return(
			"github.com/test/pkg1\ngithub.com/test/pkg2\ngithub.com/test/pkg3", nil)

		err := utils.WithIsolatedRunner(
			ts.env.Runner,
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup
			func() interface{} { return GetRunner() },
			func() error {
				packages, err := ts.build.discoverPackages("./...", "")
				ts.Require().NoError(err)
				ts.Require().Len(packages, 3)
				ts.Assert().Contains(packages, "github.com/test/pkg1")
				return nil
			},
		)
		ts.Require().NoError(err)
	})

	ts.Run("findMainPackages identifies main packages", func() {
		ts.env.Runner.On("RunCmdOutput", "go", []string{"list", "-json", "./..."}).Return(
			`{"ImportPath":"github.com/test/cmd/app","Name":"main"}
{"ImportPath":"github.com/test/pkg1","Name":"pkg1"}
{"ImportPath":"github.com/test/cmd/tool","Name":"main"}`, nil)

		err := utils.WithIsolatedRunner(
			ts.env.Runner,
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup
			func() interface{} { return GetRunner() },
			func() error {
				mainPkgs, err := ts.build.findMainPackages()
				ts.Require().NoError(err)
				ts.Require().Len(mainPkgs, 2)
				ts.Assert().Contains(mainPkgs, "github.com/test/cmd/app")
				ts.Assert().Contains(mainPkgs, "github.com/test/cmd/tool")
				return nil
			},
		)
		ts.Require().NoError(err)
	})

	ts.Run("splitIntoBatches divides packages correctly", func() {
		packages := []string{"pkg1", "pkg2", "pkg3", "pkg4", "pkg5"}

		// Test batch size 2
		batches := ts.build.splitIntoBatches(packages, 2)
		ts.Require().Len(batches, 3) // 5 packages / 2 per batch = 3 batches
		ts.Assert().Len(batches[0], 2)
		ts.Assert().Len(batches[1], 2)
		ts.Assert().Len(batches[2], 1) // Last batch has remainder

		// Test batch size 3
		batches = ts.build.splitIntoBatches(packages, 3)
		ts.Require().Len(batches, 2) // 5 packages / 3 per batch = 2 batches
		ts.Assert().Len(batches[0], 3)
		ts.Assert().Len(batches[1], 2)

		// Test invalid batch size (should default to 10)
		batches = ts.build.splitIntoBatches(packages, 0)
		ts.Require().Len(batches, 1) // All in one batch since we have fewer than 10
		ts.Assert().Len(batches[0], 5)
	})
}

// TestBuildTestSuite runs the test suite
func TestBuildTestSuite(t *testing.T) {
	suite.Run(t, new(BuildTestSuite))
}
