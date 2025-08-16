//go:build integration
// +build integration

package mage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
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
	if err := os.Unsetenv("DOCKER_BUILDKIT"); err != nil {
		ts.T().Logf("Failed to unset DOCKER_BUILDKIT: %v", err)
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

// TestBuildDocker tests the Docker method
func (ts *BuildTestSuite) TestBuildDocker() {
	ts.Run("builds docker image with default settings", func() {
		// Create Dockerfile
		ts.env.CreateFile("Dockerfile", `FROM golang:1.24-alpine
WORKDIR /app
COPY . .
RUN go build -o app .
CMD ["./app"]`)

		// Mock git commands needed for version info
		ts.mockGitCommands()

		// Mock docker build command with flexible matching
		ts.env.Runner.On("RunCmd", "docker", mock.MatchedBy(func(args []string) bool {
			return len(args) >= 4 && args[0] == "build" && args[1] == "-t" && args[len(args)-1] == "."
		})).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.Docker()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("handles missing Dockerfile", func() {
		// Ensure the Dockerfile doesn't exist
		if err := os.RemoveAll(filepath.Join(ts.env.TempDir, "Dockerfile")); err != nil {
			ts.T().Logf("Failed to remove Dockerfile: %v", err)
		}

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.build.Docker()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "dockerfile not found")
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
	ts.Run("runs pre-build tasks", func() {
		// Mock go mod tidy
		ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(nil)
		// Mock go build with flexible args
		ts.env.Runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
			return len(args) >= 2 && args[0] == "build" && args[len(args)-1] == "./..."
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

// TestBuildTestSuite runs the test suite
func TestBuildTestSuite(t *testing.T) {
	suite.Run(t, new(BuildTestSuite))
}
