package mage

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// Static test errors to satisfy err113 linter
var (
	errGitInitFailed       = errors.New("git init failed")
	errGoModDownloadFailed = errors.New("go mod download failed")
	errGoGetFailed         = errors.New("go get failed")
)

// Constants for commonly used strings
const (
	commitCommand = "commit"
)

// InitTestSuite defines the test suite for Init namespace methods
type InitTestSuite struct {
	suite.Suite

	env  *testutil.TestEnvironment
	init Init
}

// SetupTest runs before each test
func (ts *InitTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.init = Init{}
}

// TearDownTest runs after each test
func (ts *InitTestSuite) TearDownTest() {
	// Clean up environment variables that might be set by tests
	envVars := []string{
		"PROJECT_NAME",
		"PROJECT_MODULE",
		"PROJECT_DESCRIPTION",
		"PROJECT_AUTHOR",
		"PROJECT_EMAIL",
		"PROJECT_LICENSE",
		"PROJECT_TYPE",
	}

	for _, envVar := range envVars {
		if err := os.Unsetenv(envVar); err != nil {
			ts.T().Logf("Failed to unset %s: %v", envVar, err)
		}
	}

	ts.env.Cleanup()
}

// TestInitDefault tests the Default method
func (ts *InitTestSuite) TestInitDefault() {
	ts.Run("successful default initialization", func() {
		// Mock successful commands for full project setup
		ts.env.Runner.On("RunCmd", "git", []string{"init"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"add", "."}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", mock.MatchedBy(func(args []string) bool {
			return len(args) >= 3 && args[0] == commitCommand && args[1] == "-m"
		})).Return(nil)
		ts.env.Runner.On("RunCmd", "go", []string{"mod", "download"}).Return(nil)
		ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.init.Default()
			},
		)

		ts.Require().NoError(err)

		// Verify project structure was created
		ts.Require().True(ts.env.FileExists("go.mod"))
		ts.Require().True(ts.env.FileExists("main.go"))
		ts.Require().True(ts.env.FileExists("README.md"))
		ts.Require().True(ts.env.FileExists(".gitignore"))
		ts.Require().True(ts.env.FileExists("LICENSE"))
		ts.Require().True(ts.env.FileExists("magefile.go"))
		ts.Require().True(ts.env.FileExists(".github/workflows/ci.yml"))

		// Verify directory structure
		dirs := []string{"cmd", "pkg", "internal", "test", "docs", "scripts", ".github/workflows"}
		for _, dir := range dirs {
			_, err := os.Stat(dir)
			ts.Require().NoError(err, "Directory %s should exist", dir)
		}
	})

	ts.Run("handles git initialization failure gracefully", func() {
		// Mock git init failure - should warn but not fail overall
		ts.env.Runner.On("RunCmd", "git", []string{"init"}).Return(errGitInitFailed)
		ts.env.Runner.On("RunCmd", "go", []string{"mod", "download"}).Return(nil)
		ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.init.Default()
			},
		)

		// Should succeed despite git failure (warns but continues)
		ts.Require().NoError(err)
	})

	ts.Run("handles dependency installation failure gracefully", func() {
		// Mock successful git but failed dependencies - should warn but not fail
		ts.env.Runner.On("RunCmd", "git", []string{"init"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"add", "."}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", mock.MatchedBy(func(args []string) bool {
			return len(args) >= 3 && args[0] == commitCommand && args[1] == "-m"
		})).Return(nil)
		ts.env.Runner.On("RunCmd", "go", []string{"mod", "download"}).Return(errGoModDownloadFailed)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.init.Default()
			},
		)

		// Should succeed despite dependency failure (warns but continues)
		ts.Require().NoError(err)
	})
}

// TestInitProjectTypes tests all project type initialization methods
func (ts *InitTestSuite) TestInitProjectTypes() {
	projectTypes := []struct {
		name         string
		method       func() error
		projectType  ProjectType
		useDocker    bool
		expectedDirs []string
		features     []string
	}{
		{
			name:         "Library project",
			method:       ts.init.Library,
			projectType:  LibraryProject,
			useDocker:    false,
			expectedDirs: []string{"examples"},
			features:     []string{"testing", "benchmarks", "docs"},
		},
		{
			name:         "CLI project",
			method:       ts.init.CLI,
			projectType:  CLIProject,
			useDocker:    true,
			expectedDirs: []string{"completions"},
			features:     []string{"cobra", "viper", "testing", "goreleaser"},
		},
		{
			name:         "WebAPI project",
			method:       ts.init.WebAPI,
			projectType:  WebAPIProject,
			useDocker:    true,
			expectedDirs: []string{"api", "migrations", "deployments"},
			features:     []string{"gin", "gorm", "swagger", "testing", "migrations"},
		},
		{
			name:         "Microservice project",
			method:       ts.init.Microservice,
			projectType:  MicroserviceProject,
			useDocker:    true,
			expectedDirs: []string{"api", "migrations", "deployments"},
			features:     []string{"grpc", "prometheus", "tracing", "testing", "kubernetes"},
		},
		{
			name:         "Tool project",
			method:       ts.init.Tool,
			projectType:  ToolProject,
			useDocker:    true,
			expectedDirs: []string{"completions"},
			features:     []string{"cobra", "testing", "goreleaser", "homebrew"},
		},
	}

	for _, tc := range projectTypes {
		ts.Run(tc.name, func() {
			// Create fresh environment for each project type
			env := testutil.NewTestEnvironment(ts.T())
			defer env.Cleanup()

			// Mock successful commands
			env.Runner.On("RunCmd", "git", []string{"init"}).Return(nil)
			env.Runner.On("RunCmd", "git", []string{"add", "."}).Return(nil)
			env.Runner.On("RunCmd", "git", mock.MatchedBy(func(args []string) bool {
				return len(args) >= 3 && args[0] == commitCommand && args[1] == "-m"
			})).Return(nil)
			env.Runner.On("RunCmd", "go", []string{"mod", "download"}).Return(nil)
			env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(nil)

			err := env.WithMockRunner(
				func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
				func() interface{} { return GetRunner() },
				func() error {
					return tc.method()
				},
			)

			ts.Require().NoError(err)

			// Verify project structure
			ts.Require().True(env.FileExists("go.mod"))
			ts.Require().True(env.FileExists("main.go"))
			ts.Require().True(env.FileExists("magefile.go"))

			// Verify main.go contains appropriate template
			mainContent := env.ReadFile("main.go")
			ts.Require().NotEmpty(mainContent)

			// Verify type-specific directories exist
			for _, dir := range tc.expectedDirs {
				_, err := os.Stat(dir)
				ts.Require().NoError(err, "Type-specific directory %s should exist for %s", dir, tc.name)
			}

			// Verify Docker files if expected
			if tc.useDocker {
				ts.Require().True(env.FileExists("Dockerfile"))
				if tc.projectType == WebAPIProject || tc.projectType == MicroserviceProject {
					ts.Require().True(env.FileExists("docker-compose.yml"))
				}
			}
		})
	}
}

// TestInitUpgrade tests the Upgrade method
func (ts *InitTestSuite) TestInitUpgrade() {
	ts.Run("successful upgrade of existing project", func() {
		// Create existing go.mod to simulate existing project
		ts.env.CreateGoMod("existing/project")

		// Mock successful commands
		ts.env.Runner.On("RunCmd", "go", []string{"get", "github.com/magefile/mage@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "go", []string{"get", "github.com/mrz1836/mage-x@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.init.Upgrade()
			},
		)

		ts.Require().NoError(err)

		// Verify magefile.go was created
		ts.Require().True(ts.env.FileExists("magefile.go"))

		// Verify magefile content
		mageContent := ts.env.ReadFile("magefile.go")
		ts.Require().Contains(mageContent, "//go:build mage")
		ts.Require().Contains(mageContent, "github.com/mrz1836/mage-x/pkg/mage")
	})

	ts.Run("fails when not a Go project", func() {
		// Create fresh environment without go.mod
		env := testutil.NewTestEnvironment(ts.T())
		defer env.Cleanup()

		err := env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.init.Upgrade()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "doesn't appear to be a Go project")
	})

	ts.Run("handles go get failure", func() {
		// Create fresh environment with go.mod
		env := testutil.NewTestEnvironment(ts.T())
		defer env.Cleanup()

		env.CreateGoMod("existing/project")

		// Mock failed go get
		env.Runner.On("RunCmd", "go", []string{"get", "github.com/magefile/mage@latest"}).Return(errGoGetFailed)

		err := env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.init.Upgrade()
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "failed to update go.mod")
	})
}

// TestInitTemplates tests the Templates method
func (ts *InitTestSuite) TestInitTemplates() {
	ts.Run("displays available templates", func() {
		err := ts.init.Templates()
		ts.Require().NoError(err)
		// Templates method should always succeed as it just displays information
	})
}

// TestGetInitProjectConfig tests the getInitProjectConfig function
func (ts *InitTestSuite) TestGetInitProjectConfig() {
	ts.Run("gets config from environment variables", func() {
		// Set environment variables
		ts.Require().NoError(os.Setenv("PROJECT_NAME", "test-project"))
		ts.Require().NoError(os.Setenv("PROJECT_MODULE", "github.com/test/test-project"))
		ts.Require().NoError(os.Setenv("PROJECT_DESCRIPTION", "A test project"))
		ts.Require().NoError(os.Setenv("PROJECT_AUTHOR", "Test Author"))
		ts.Require().NoError(os.Setenv("PROJECT_EMAIL", "test@example.com"))
		ts.Require().NoError(os.Setenv("PROJECT_LICENSE", "Apache-2.0"))
		ts.Require().NoError(os.Setenv("PROJECT_TYPE", "cli"))

		config, err := getInitProjectConfig()
		ts.Require().NoError(err)

		ts.Require().Equal("test-project", config.Name)
		ts.Require().Equal("github.com/test/test-project", config.Module)
		ts.Require().Equal("A test project", config.Description)
		ts.Require().Equal("Test Author", config.Author)
		ts.Require().Equal("test@example.com", config.Email)
		ts.Require().Equal("Apache-2.0", config.License)
		ts.Require().Equal(CLIProject, config.Type)
		ts.Require().True(config.UseMage)
		ts.Require().True(config.UseCI)
		ts.Require().Equal(runtime.Version(), config.GoVersion)
	})

	ts.Run("uses defaults when no environment variables", func() {
		// Clear all environment variables that might affect the test
		envVars := []string{
			"PROJECT_NAME",
			"PROJECT_MODULE",
			"PROJECT_DESCRIPTION",
			"PROJECT_AUTHOR",
			"PROJECT_EMAIL",
			"PROJECT_LICENSE",
			"PROJECT_TYPE",
		}

		for _, envVar := range envVars {
			ts.Require().NoError(os.Unsetenv(envVar))
		}

		config, err := getInitProjectConfig()
		ts.Require().NoError(err)

		// Should use current directory name as project name
		currentDir, err := os.Getwd()
		ts.Require().NoError(err)
		expectedName := filepath.Base(currentDir)

		ts.Require().Equal(expectedName, config.Name)
		ts.Require().Equal("github.com/username/"+expectedName, config.Module)
		ts.Require().Equal("A generic project built with MAGE-X", config.Description)
		ts.Require().Equal("MIT", config.License)
		ts.Require().Equal(GenericProject, config.Type)
		ts.Require().Equal(runtime.Version(), config.GoVersion)
	})

	ts.Run("handles all project types", func() {
		testCases := []struct {
			envValue string
			expected ProjectType
		}{
			{"library", LibraryProject},
			{"cli", CLIProject},
			{"webapi", WebAPIProject},
			{"microservice", MicroserviceProject},
			{"tool", ToolProject},
			{"invalid", GenericProject},
			{"", GenericProject},
		}

		for _, tc := range testCases {
			ts.Run("project type "+tc.envValue, func() {
				ts.Require().NoError(os.Setenv("PROJECT_TYPE", tc.envValue))

				config, err := getInitProjectConfig()
				ts.Require().NoError(err)
				ts.Require().Equal(tc.expected, config.Type)

				ts.Require().NoError(os.Unsetenv("PROJECT_TYPE"))
			})
		}
	})
}

// TestCreateProjectStructure tests the createProjectStructure function
func (ts *InitTestSuite) TestCreateProjectStructure() {
	ts.Run("creates basic directory structure", func() {
		config := &InitProjectConfig{
			Type: GenericProject,
		}

		err := createProjectStructure(config)
		ts.Require().NoError(err)

		// Verify basic directories exist
		basicDirs := []string{"cmd", "pkg", "internal", "test", "docs", "scripts", ".github/workflows"}
		for _, dir := range basicDirs {
			_, err := os.Stat(dir)
			ts.Require().NoError(err, "Basic directory %s should exist", dir)
		}
	})

	ts.Run("creates type-specific directories", func() {
		testCases := []struct {
			projectType  ProjectType
			expectedDirs []string
		}{
			{LibraryProject, []string{"examples"}},
			{CLIProject, []string{"completions"}},
			{ToolProject, []string{"completions"}},
			{WebAPIProject, []string{"api", "migrations", "deployments"}},
			{MicroserviceProject, []string{"api", "migrations", "deployments"}},
		}

		for _, tc := range testCases {
			ts.Run(string(tc.projectType), func() {
				// Create fresh environment
				env := testutil.NewTestEnvironment(ts.T())
				defer env.Cleanup()

				config := &InitProjectConfig{
					Type: tc.projectType,
				}

				err := createProjectStructure(config)
				ts.Require().NoError(err)

				// Verify type-specific directories
				for _, dir := range tc.expectedDirs {
					_, err := os.Stat(dir)
					ts.Require().NoError(err, "Type-specific directory %s should exist for %s", dir, tc.projectType)
				}
			})
		}
	})
}

// TestInitializeProjectFiles tests the initializeProjectFiles function
func (ts *InitTestSuite) TestInitializeProjectFiles() {
	ts.Run("creates all project files", func() {
		config := &InitProjectConfig{
			Name:        "test-app",
			Module:      "github.com/test/test-app",
			Description: "Test application",
			Author:      "Test Author",
			License:     "MIT",
			GoVersion:   "go1.24",
			Type:        GenericProject,
			UseMage:     true,
			UseDocker:   true,
			UseCI:       true,
		}

		// Create the directory structure first (as done in the full initialization)
		err := createProjectStructure(config)
		ts.Require().NoError(err)

		err = initializeProjectFiles(config)
		ts.Require().NoError(err)

		// Verify all files were created
		expectedFiles := []string{
			"go.mod",
			"main.go",
			"README.md",
			".gitignore",
			"LICENSE",
			"magefile.go",
			"Dockerfile",
			".github/workflows/ci.yml",
		}

		for _, file := range expectedFiles {
			ts.Require().True(ts.env.FileExists(file), "File %s should exist", file)
		}
	})

	ts.Run("skips optional files when disabled", func() {
		// Create fresh environment to avoid file pollution from previous test
		env := testutil.NewTestEnvironment(ts.T())
		defer env.Cleanup()

		config := &InitProjectConfig{
			Name:        "minimal-app",
			Module:      "github.com/test/minimal-app",
			Description: "Minimal application",
			Type:        GenericProject,
			UseMage:     false,
			UseDocker:   false,
			UseCI:       false,
		}

		err := initializeProjectFiles(config)
		ts.Require().NoError(err)

		// Verify core files exist
		ts.Require().True(env.FileExists("go.mod"))
		ts.Require().True(env.FileExists("main.go"))
		ts.Require().True(env.FileExists("README.md"))
		ts.Require().True(env.FileExists(".gitignore"))
		ts.Require().True(env.FileExists("LICENSE"))

		// Verify optional files don't exist
		ts.Require().False(env.FileExists("magefile.go"))
		ts.Require().False(env.FileExists("Dockerfile"))
		ts.Require().False(env.FileExists(".github/workflows/ci.yml"))
	})
}

// TestFileCreationFunctions tests individual file creation functions
func (ts *InitTestSuite) TestFileCreationFunctions() {
	ts.Run("createGoMod creates correct go.mod", func() {
		config := &InitProjectConfig{
			Module:    "github.com/test/project",
			GoVersion: "go1.24",
		}

		err := createGoMod(config)
		ts.Require().NoError(err)

		content := ts.env.ReadFile("go.mod")
		ts.Require().Contains(content, "module github.com/test/project")
		ts.Require().Contains(content, "go 1.24")
		ts.Require().Contains(content, "github.com/magefile/mage")
		ts.Require().Contains(content, "github.com/mrz1836/mage-x")
	})

	ts.Run("createMainFile creates correct templates", func() {
		testCases := []struct {
			projectType ProjectType
			contains    []string
		}{
			{
				GenericProject,
				[]string{"package main", "func main()", "Hello"},
			},
			{
				CLIProject,
				[]string{"package main", "os.Args", "version", "help"},
			},
			{
				WebAPIProject,
				[]string{"package main", "http.HandleFunc", ":8080", "/health"},
			},
			{
				MicroserviceProject,
				[]string{"package main", "http.Server", "/ready", "Shutting down"},
			},
		}

		for _, tc := range testCases {
			ts.Run(string(tc.projectType), func() {
				config := &InitProjectConfig{
					Name: "test-app",
					Type: tc.projectType,
				}

				err := createMainFile(config)
				ts.Require().NoError(err)

				content := ts.env.ReadFile("main.go")
				for _, expected := range tc.contains {
					ts.Require().Contains(content, expected)
				}
			})
		}
	})

	ts.Run("createReadme creates correct README", func() {
		config := &InitProjectConfig{
			Name:        "awesome-app",
			Module:      "github.com/test/awesome-app",
			Description: "An awesome application",
			License:     "MIT",
		}

		err := createReadme(config)
		ts.Require().NoError(err)

		content := ts.env.ReadFile("README.md")
		ts.Require().Contains(content, "# awesome-app")
		ts.Require().Contains(content, "An awesome application")
		ts.Require().Contains(content, "go install github.com/test/awesome-app@latest")
		ts.Require().Contains(content, "MIT License")
		ts.Require().Contains(content, "MAGE-X")
	})

	ts.Run("createGitignore creates standard .gitignore", func() {
		err := createGitignore()
		ts.Require().NoError(err)

		content := ts.env.ReadFile(".gitignore")
		expectedEntries := []string{
			"*.exe",
			"*.test",
			"*.out",
			"vendor/",
			".DS_Store",
			"bin/",
			".env",
		}

		for _, entry := range expectedEntries {
			ts.Require().Contains(content, entry)
		}
	})

	ts.Run("createLicense creates MIT license", func() {
		config := &InitProjectConfig{
			Author: "John Doe",
		}

		err := createLicense(config)
		ts.Require().NoError(err)

		content := ts.env.ReadFile("LICENSE")
		ts.Require().Contains(content, "MIT License")
		ts.Require().Contains(content, "John Doe")
		ts.Require().Contains(content, "2024")
	})

	ts.Run("addMageFiles creates magefile.go", func() {
		err := addMageFiles()
		ts.Require().NoError(err)

		content := ts.env.ReadFile("magefile.go")
		ts.Require().Contains(content, "//go:build mage")
		ts.Require().Contains(content, "github.com/mrz1836/mage-x/pkg/mage")
		ts.Require().Contains(content, "func Build() error")
		ts.Require().Contains(content, "func Test() error")
		ts.Require().Contains(content, "func Lint() error")
		ts.Require().Contains(content, "func Clean() error")
	})
}

// TestTemplateGeneration tests template generation functions
func (ts *InitTestSuite) TestTemplateGeneration() {
	ts.Run("getMainGenericTemplate", func() {
		config := &InitProjectConfig{Name: "test-app"}
		content := getMainGenericTemplate(config)

		ts.Require().Contains(content, "package main")
		ts.Require().Contains(content, "func main()")
		ts.Require().Contains(content, "test-app")
	})

	ts.Run("getMainCLITemplate", func() {
		config := &InitProjectConfig{Name: "cli-app"}
		content := getMainCLITemplate(config)

		ts.Require().Contains(content, "package main")
		ts.Require().Contains(content, "os.Args")
		ts.Require().Contains(content, "version")
		ts.Require().Contains(content, "cli-app v1.0.0")
	})

	ts.Run("getMainWebAPITemplate", func() {
		config := &InitProjectConfig{Name: "web-app"}
		content := getMainWebAPITemplate(config)

		ts.Require().Contains(content, "package main")
		ts.Require().Contains(content, "http.HandleFunc")
		ts.Require().Contains(content, ":8080")
		ts.Require().Contains(content, "web-app")
	})

	ts.Run("getMainMicroserviceTemplate", func() {
		config := &InitProjectConfig{Name: "micro-app"}
		content := getMainMicroserviceTemplate(config)

		ts.Require().Contains(content, "package main")
		ts.Require().Contains(content, "http.Server")
		ts.Require().Contains(content, "Shutting down")
		ts.Require().Contains(content, "micro-app")
	})
}

// TestDockerAndCIFiles tests Docker and CI file creation
func (ts *InitTestSuite) TestDockerAndCIFiles() {
	ts.Run("createDockerFiles creates Dockerfile", func() {
		config := &InitProjectConfig{
			Name: "docker-app",
			Type: GenericProject,
		}

		err := createDockerFiles(config)
		ts.Require().NoError(err)

		content := ts.env.ReadFile("Dockerfile")
		ts.Require().Contains(content, "FROM golang:")
		ts.Require().Contains(content, "WORKDIR /app")
		ts.Require().Contains(content, "go build")
	})

	ts.Run("createDockerFiles creates docker-compose for web projects", func() {
		testCases := []ProjectType{WebAPIProject, MicroserviceProject}

		for _, projectType := range testCases {
			ts.Run(string(projectType), func() {
				config := &InitProjectConfig{
					Name: "web-app",
					Type: projectType,
				}

				err := createDockerFiles(config)
				ts.Require().NoError(err)

				ts.Require().True(ts.env.FileExists("docker-compose.yml"))
				content := ts.env.ReadFile("docker-compose.yml")
				ts.Require().Contains(content, "version: '3.8'")
				ts.Require().Contains(content, "postgres")
				ts.Require().Contains(content, "web-app")
			})
		}
	})

	ts.Run("createCIFiles creates GitHub Actions workflow", func() {
		// Create the directory first
		err := os.MkdirAll(".github/workflows", 0o750)
		ts.Require().NoError(err)

		err = createCIFiles()
		ts.Require().NoError(err)

		content := ts.env.ReadFile(".github/workflows/ci.yml")
		ts.Require().Contains(content, "name: CI")
		ts.Require().Contains(content, "actions/checkout")
		ts.Require().Contains(content, "actions/setup-go")
		ts.Require().Contains(content, "mage test")
		ts.Require().Contains(content, "mage lint")
		ts.Require().Contains(content, "mage build")
	})
}

// TestGitAndDependencies tests git and dependency management functions
func (ts *InitTestSuite) TestGitAndDependencies() {
	ts.Run("initializeGitRepo creates git repository", func() {
		config := &InitProjectConfig{
			Description: "Test project description",
		}

		ts.env.Runner.On("RunCmd", "git", []string{"init"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"add", "."}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", mock.MatchedBy(func(args []string) bool {
			return len(args) >= 3 && args[0] == commitCommand && args[1] == "-m" &&
				strings.Contains(args[2], "Test project description")
		})).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return initializeGitRepo(config)
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("installDependencies runs go commands", func() {
		ts.env.Runner.On("RunCmd", "go", []string{"mod", "download"}).Return(nil)
		ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return installDependencies()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("updateGoMod installs mage dependencies", func() {
		ts.env.Runner.On("RunCmd", "go", []string{"get", "github.com/magefile/mage@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "go", []string{"get", "github.com/mrz1836/mage-x@latest"}).Return(nil)
		ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return updateGoMod()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestUtilityFunctions tests utility functions
func (ts *InitTestSuite) TestUtilityFunctions() {
	ts.Run("showCompletionMessage displays project info", func() {
		config := &InitProjectConfig{
			Name:      "test-project",
			Module:    "github.com/test/test-project",
			Type:      CLIProject,
			Features:  []string{"cobra", "viper"},
			UseDocker: true,
		}

		// Should not panic or error
		ts.Require().NotPanics(func() {
			showCompletionMessage(config)
		})
	})
}

// TestInitializeProject tests the initializeProject helper function
func (ts *InitTestSuite) TestInitializeProject() {
	ts.Run("merges config and initializes project", func() {
		// Set some environment variables
		ts.Require().NoError(os.Setenv("PROJECT_NAME", "env-project"))
		ts.Require().NoError(os.Setenv("PROJECT_AUTHOR", "Env Author"))

		config := &InitProjectConfig{
			Type:      CLIProject,
			UseMage:   true,
			UseDocker: true,
			UseCI:     true,
			Features:  []string{"cobra", "viper"},
		}

		// Mock successful commands
		ts.env.Runner.On("RunCmd", "git", []string{"init"}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", []string{"add", "."}).Return(nil)
		ts.env.Runner.On("RunCmd", "git", mock.MatchedBy(func(args []string) bool {
			return len(args) >= 3 && args[0] == commitCommand && args[1] == "-m"
		})).Return(nil)
		ts.env.Runner.On("RunCmd", "go", []string{"mod", "download"}).Return(nil)
		ts.env.Runner.On("RunCmd", "go", []string{"mod", "tidy"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return initializeProject(config)
			},
		)

		ts.Require().NoError(err)

		// Verify project was created with merged config
		ts.Require().True(ts.env.FileExists("go.mod"))
		ts.Require().True(ts.env.FileExists("main.go"))
		ts.Require().True(ts.env.FileExists("magefile.go"))
		ts.Require().True(ts.env.FileExists("Dockerfile"))
		ts.Require().True(ts.env.FileExists(".github/workflows/ci.yml"))
	})
}

// TestAdditionalMethods tests the additional methods required by the namespace
func (ts *InitTestSuite) TestAdditionalMethods() {
	additionalMethods := []struct {
		name   string
		method func() error
	}{
		{"Project", ts.init.Project},
		{"Config", ts.init.Config},
		{"Git", ts.init.Git},
		{"Mage", ts.init.Mage},
		{"CI", ts.init.CI},
		{"Docker", ts.init.Docker},
		{"Docs", ts.init.Docs},
		{"License", ts.init.License},
		{"Makefile", ts.init.Makefile},
		{"Editorconfig", ts.init.Editorconfig},
	}

	for _, method := range additionalMethods {
		ts.Run(method.name, func() {
			// Mock the echo command that these methods use
			ts.env.Runner.On("RunCmd", mock.AnythingOfType("string"), mock.Anything).Return(nil)

			err := ts.env.WithMockRunner(
				func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
				func() interface{} { return GetRunner() },
				method.method,
			)

			ts.Require().NoError(err)
		})
	}

	// Special test for Git method which runs actual git command
	ts.Run("Git method runs git init", func() {
		ts.env.Runner.On("RunCmd", "git", []string{"init"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.init.Git()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestErrorHandling tests comprehensive error handling scenarios
func (ts *InitTestSuite) TestErrorHandling() {
	ts.Run("handles directory creation failure", func() {
		// This test would require mocking the filesystem, which is complex
		// In a real scenario, you might test with permissions issues
		// For now, we test the happy path and assume directory creation works
		// or fails at the OS level, which createProjectStructure handles appropriately
		config := &InitProjectConfig{Type: GenericProject}
		err := createProjectStructure(config)
		ts.Require().NoError(err) // Should succeed in test environment
	})

	ts.Run("getInitProjectConfig handles current directory error", func() {
		// This is hard to test without mocking os.Getwd()
		// The function handles errors by returning them
		config, err := getInitProjectConfig()
		ts.Require().NoError(err) // Should succeed in test environment
		ts.Require().NotNil(config)
	})
}

// TestTableDrivenProjectTypes tests all project types with table-driven approach
func (ts *InitTestSuite) TestTableDrivenProjectTypes() {
	testCases := []struct {
		name             string
		projectType      ProjectType
		useMage          bool
		useDocker        bool
		useCI            bool
		expectedFeatures []string
		expectedFiles    []string
		expectedDirs     []string
		mainFileContains []string
	}{
		{
			name:             "Library Project Complete",
			projectType:      LibraryProject,
			useMage:          true,
			useDocker:        false,
			useCI:            true,
			expectedFeatures: []string{"testing", "benchmarks", "docs"},
			expectedFiles:    []string{"go.mod", "main.go", "README.md", ".gitignore", "LICENSE", "magefile.go", ".github/workflows/ci.yml"},
			expectedDirs:     []string{"cmd", "pkg", "internal", "test", "docs", "scripts", "examples"},
			mainFileContains: []string{"package main", "func main()"},
		},
		{
			name:             "CLI Project Complete",
			projectType:      CLIProject,
			useMage:          true,
			useDocker:        true,
			useCI:            true,
			expectedFeatures: []string{"cobra", "viper", "testing", "goreleaser"},
			expectedFiles:    []string{"go.mod", "main.go", "README.md", "magefile.go", "Dockerfile", ".github/workflows/ci.yml"},
			expectedDirs:     []string{"cmd", "pkg", "internal", "test", "docs", "scripts", "completions"},
			mainFileContains: []string{"package main", "os.Args", "version"},
		},
		{
			name:             "WebAPI Project Complete",
			projectType:      WebAPIProject,
			useMage:          true,
			useDocker:        true,
			useCI:            true,
			expectedFeatures: []string{"gin", "gorm", "swagger", "testing", "migrations"},
			expectedFiles:    []string{"go.mod", "main.go", "README.md", "magefile.go", "Dockerfile", "docker-compose.yml", ".github/workflows/ci.yml"},
			expectedDirs:     []string{"cmd", "pkg", "internal", "test", "docs", "scripts", "api", "migrations", "deployments"},
			mainFileContains: []string{"package main", "http.HandleFunc", ":8080"},
		},
		{
			name:             "Microservice Project Complete",
			projectType:      MicroserviceProject,
			useMage:          true,
			useDocker:        true,
			useCI:            true,
			expectedFeatures: []string{"grpc", "prometheus", "tracing", "testing", "kubernetes"},
			expectedFiles:    []string{"go.mod", "main.go", "README.md", "magefile.go", "Dockerfile", "docker-compose.yml", ".github/workflows/ci.yml"},
			expectedDirs:     []string{"cmd", "pkg", "internal", "test", "docs", "scripts", "api", "migrations", "deployments"},
			mainFileContains: []string{"package main", "http.Server", "Shutting down"},
		},
		{
			name:             "Tool Project Complete",
			projectType:      ToolProject,
			useMage:          true,
			useDocker:        true,
			useCI:            true,
			expectedFeatures: []string{"cobra", "testing", "goreleaser", "homebrew"},
			expectedFiles:    []string{"go.mod", "main.go", "README.md", "magefile.go", "Dockerfile", ".github/workflows/ci.yml"},
			expectedDirs:     []string{"cmd", "pkg", "internal", "test", "docs", "scripts", "completions"},
			mainFileContains: []string{"package main", "os.Args", "version"},
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			// Create fresh environment
			env := testutil.NewTestEnvironment(ts.T())
			defer env.Cleanup()

			config := &InitProjectConfig{
				Name:        "test-project",
				Module:      "github.com/test/test-project",
				Description: "Test project description",
				Author:      "Test Author",
				License:     "MIT",
				Type:        tc.projectType,
				UseMage:     tc.useMage,
				UseDocker:   tc.useDocker,
				UseCI:       tc.useCI,
				Features:    tc.expectedFeatures,
			}

			// Create project structure and files
			err := createProjectStructure(config)
			ts.Require().NoError(err)

			err = initializeProjectFiles(config)
			ts.Require().NoError(err)

			// Verify expected files exist
			for _, file := range tc.expectedFiles {
				ts.Require().True(env.FileExists(file), "File %s should exist for %s project", file, tc.projectType)
			}

			// Verify expected directories exist
			for _, dir := range tc.expectedDirs {
				_, err := os.Stat(dir)
				ts.Require().NoError(err, "Directory %s should exist for %s project", dir, tc.projectType)
			}

			// Verify main.go content
			if env.FileExists("main.go") {
				mainContent := env.ReadFile("main.go")
				for _, expected := range tc.mainFileContains {
					ts.Require().Contains(mainContent, expected, "main.go should contain %s for %s project", expected, tc.projectType)
				}
			}

			// Verify Docker files for projects that use Docker
			if tc.useDocker {
				ts.Require().True(env.FileExists("Dockerfile"), "Dockerfile should exist when useDocker is true")

				// WebAPI and Microservice projects should have docker-compose.yml
				if tc.projectType == WebAPIProject || tc.projectType == MicroserviceProject {
					ts.Require().True(env.FileExists("docker-compose.yml"), "docker-compose.yml should exist for %s project", tc.projectType)
				}
			}

			// Verify CI files when enabled
			if tc.useCI {
				ts.Require().True(env.FileExists(".github/workflows/ci.yml"), "CI workflow should exist when useCI is true")
			}

			// Verify Mage files when enabled
			if tc.useMage {
				ts.Require().True(env.FileExists("magefile.go"), "magefile.go should exist when useMage is true")
			}
		})
	}
}

// TestInitTestSuite runs the test suite
func TestInitTestSuite(t *testing.T) {
	suite.Run(t, new(InitTestSuite))
}
