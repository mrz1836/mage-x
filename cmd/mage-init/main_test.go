package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type MageInitTestSuite struct {
	suite.Suite

	tempDir string
}

func (suite *MageInitTestSuite) SetupTest() {
	tempDir := suite.T().TempDir()
	suite.tempDir = tempDir
}

func TestMageInitTestSuite(t *testing.T) {
	suite.Run(t, new(MageInitTestSuite))
}

// Test validateOptions function
func (suite *MageInitTestSuite) TestValidateOptionsWithDefaults() {
	opts := &InitOptions{
		ProjectPath: suite.tempDir,
		Template:    "basic",
		GoVersion:   getDefaultGoVersion(),
	}

	err := validateOptions(opts)
	suite.Require().NoError(err)
	suite.NotEmpty(opts.ProjectName)
	suite.NotEmpty(opts.ModulePath)
}

func (suite *MageInitTestSuite) TestValidateOptionsWithCustomValues() {
	opts := &InitOptions{
		ProjectName: "test-project",
		ProjectPath: suite.tempDir,
		ModulePath:  "github.com/test/test-project",
		Template:    "advanced",
		GoVersion:   getDefaultGoVersion(),
	}

	err := validateOptions(opts)
	suite.Require().NoError(err)
	suite.Equal("test-project", opts.ProjectName)
	suite.Equal("github.com/test/test-project", opts.ModulePath)
}

func (suite *MageInitTestSuite) TestValidateOptionsInvalidTemplate() {
	opts := &InitOptions{
		ProjectName: "test-project",
		ProjectPath: suite.tempDir,
		Template:    "invalid-template",
		GoVersion:   getDefaultGoVersion(),
	}

	err := validateOptions(opts)
	suite.Require().Error(err)
	suite.Require().ErrorIs(err, ErrInvalidTemplate)
	suite.Contains(err.Error(), "invalid-template")
}

func (suite *MageInitTestSuite) TestValidateOptionsCurrentDirectory() {
	originalDir, err := os.Getwd()
	suite.Require().NoError(err)
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			// Log the error but don't fail the test cleanup
			suite.T().Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(suite.tempDir)
	suite.Require().NoError(err)

	opts := &InitOptions{
		ProjectPath: ".",
		Template:    "basic",
		GoVersion:   getDefaultGoVersion(),
	}

	err = validateOptions(opts)
	suite.Require().NoError(err)
	suite.Equal(filepath.Base(suite.tempDir), opts.ProjectName)
}

// Test ensureProjectDirectory function
func (suite *MageInitTestSuite) TestEnsureProjectDirectoryNew() {
	projectPath := filepath.Join(suite.tempDir, "new-project")

	err := ensureProjectDirectory(projectPath, false)
	suite.Require().NoError(err)

	// Directory should exist
	stat, err := os.Stat(projectPath)
	suite.Require().NoError(err)
	suite.True(stat.IsDir())
}

func (suite *MageInitTestSuite) TestEnsureProjectDirectoryEmptyExisting() {
	projectPath := filepath.Join(suite.tempDir, "empty-project")
	err := os.Mkdir(projectPath, 0o750) // #nosec G301 -- test directory permissions
	suite.Require().NoError(err)

	err = ensureProjectDirectory(projectPath, false)
	suite.Require().NoError(err)
}

func (suite *MageInitTestSuite) TestEnsureProjectDirectoryNonEmptyWithoutForce() {
	projectPath := filepath.Join(suite.tempDir, "non-empty-project")
	err := os.Mkdir(projectPath, 0o750) // #nosec G301 -- test directory permissions
	suite.Require().NoError(err)

	// Create a file in the directory
	testFile := filepath.Join(projectPath, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0o600) // #nosec G306 -- test file permissions
	suite.Require().NoError(err)

	err = ensureProjectDirectory(projectPath, false)
	suite.Require().Error(err)
	suite.ErrorIs(err, ErrDirectoryNotEmpty)
}

func (suite *MageInitTestSuite) TestEnsureProjectDirectoryNonEmptyWithForce() {
	projectPath := filepath.Join(suite.tempDir, "non-empty-project-force")
	err := os.Mkdir(projectPath, 0o750) // #nosec G301 -- test directory permissions
	suite.Require().NoError(err)

	// Create a file in the directory
	testFile := filepath.Join(projectPath, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0o600) // #nosec G306 -- test file permissions
	suite.Require().NoError(err)

	err = ensureProjectDirectory(projectPath, true)
	suite.Require().NoError(err)
}

// Test createDirectories function
func (suite *MageInitTestSuite) TestCreateDirectories() {
	projectPath := filepath.Join(suite.tempDir, "test-dirs")
	err := os.Mkdir(projectPath, 0o750) // #nosec G301 -- test directory permissions
	suite.Require().NoError(err)

	directories := []string{"cmd", "pkg", "internal", "docs"}

	err = createDirectories(projectPath, directories, false)
	suite.Require().NoError(err)

	// Verify all directories were created
	for _, dir := range directories {
		dirPath := filepath.Join(projectPath, dir)
		stat, err := os.Stat(dirPath)
		suite.Require().NoError(err)
		suite.True(stat.IsDir())
	}
}

func (suite *MageInitTestSuite) TestCreateDirectoriesVerbose() {
	projectPath := filepath.Join(suite.tempDir, "test-dirs-verbose")
	err := os.Mkdir(projectPath, 0o750) // #nosec G301 -- test directory permissions
	suite.Require().NoError(err)

	directories := []string{"cmd", "pkg"}

	err = createDirectories(projectPath, directories, true)
	suite.Require().NoError(err)
}

// Test createFiles function
func (suite *MageInitTestSuite) TestCreateFiles() {
	projectPath := filepath.Join(suite.tempDir, "test-files")
	err := os.Mkdir(projectPath, 0o750) // #nosec G301 -- test directory permissions
	suite.Require().NoError(err)

	opts := &InitOptions{
		ProjectName: "test-project",
		ModulePath:  "github.com/test/test-project",
		GoVersion:   getDefaultGoVersion(),
	}

	files := map[string]string{
		"test.txt":    "Hello {{.ProjectName}}",
		"config.yaml": "module: {{.ModulePath}}",
		"version.go":  "const version = \"1.0.0\"",
	}

	// Create subdirectory first for subdir test
	subDir := filepath.Join(projectPath, "subdir")
	err = os.MkdirAll(subDir, 0o750) // #nosec G301 -- test directory permissions
	suite.Require().NoError(err)

	files["subdir/sub.go"] = "package main"

	err = createFiles(projectPath, files, opts, false)
	suite.Require().NoError(err)

	// Verify files were created with template processing
	testFile := filepath.Join(projectPath, "test.txt")
	content, err := os.ReadFile(testFile) // #nosec G304 -- controlled test file path
	suite.Require().NoError(err)
	suite.Equal("Hello test-project", string(content))

	configFile := filepath.Join(projectPath, "config.yaml")
	content, err = os.ReadFile(configFile) // #nosec G304 -- controlled test file path
	suite.Require().NoError(err)
	suite.Contains(string(content), "github.com/test/test-project")

	// Verify subdirectory file
	subFile := filepath.Join(projectPath, "subdir", "sub.go")
	_, err = os.Stat(subFile)
	suite.Require().NoError(err)
}

func (suite *MageInitTestSuite) TestCreateFilesVerbose() {
	projectPath := filepath.Join(suite.tempDir, "test-files-verbose")
	err := os.Mkdir(projectPath, 0o750) // #nosec G301 -- test directory permissions
	suite.Require().NoError(err)

	opts := &InitOptions{
		ProjectName: "test-project",
		ModulePath:  "github.com/test/test-project",
		GoVersion:   getDefaultGoVersion(),
	}

	files := map[string]string{
		"main.go": "package main",
	}

	err = createFiles(projectPath, files, opts, true)
	suite.Require().NoError(err)
}

// Test processTemplate function
func (suite *MageInitTestSuite) TestProcessTemplate() {
	opts := &InitOptions{
		ProjectName: "my-project",
		ModulePath:  "github.com/user/my-project",
		GoVersion:   getDefaultGoVersion(),
	}

	template := `Project: {{.ProjectName}}
Module: {{.ModulePath}}
Go Version: {{.GoVersion}}
Mixed: {{.ProjectName}}/{{.GoVersion}}`

	result := processTemplate(template, opts)

	expected := `Project: my-project
Module: github.com/user/my-project
Go Version: ` + getDefaultGoVersion() + `
Mixed: my-project/` + getDefaultGoVersion()

	suite.Equal(expected, result)
}

func (suite *MageInitTestSuite) TestProcessTemplateNoReplacements() {
	opts := &InitOptions{
		ProjectName: "test",
		ModulePath:  "test",
		GoVersion:   getDefaultGoVersion(),
	}

	template := "No templates here"
	result := processTemplate(template, opts)
	suite.Equal(template, result)
}

// Test initializeGoMod function
func (suite *MageInitTestSuite) TestInitializeGoMod() {
	projectPath := filepath.Join(suite.tempDir, "test-gomod")
	err := os.Mkdir(projectPath, 0o750) // #nosec G301 -- test directory permissions
	suite.Require().NoError(err)

	opts := &InitOptions{
		ModulePath: "github.com/test/test-project",
		GoVersion:  getDefaultGoVersion(),
	}

	err = initializeGoMod(projectPath, opts)
	suite.Require().NoError(err)

	// Verify go.mod was created
	goModPath := filepath.Join(projectPath, "go.mod")
	content, err := os.ReadFile(goModPath) // #nosec G304 -- controlled test file path
	suite.Require().NoError(err)

	contentStr := string(content)
	suite.Contains(contentStr, "module github.com/test/test-project")
	suite.Contains(contentStr, "go "+getDefaultGoVersion())
	suite.Contains(contentStr, "github.com/magefile/mage")
}

// Test createMageConfig function
func (suite *MageInitTestSuite) TestCreateMageConfig() {
	projectPath := filepath.Join(suite.tempDir, "test-mage-config")
	err := os.Mkdir(projectPath, 0o750) // #nosec G301 -- test directory permissions
	suite.Require().NoError(err)

	opts := &InitOptions{
		ProjectName: "test-project",
		GoVersion:   getDefaultGoVersion(),
	}

	err = createMageConfig(projectPath, opts)
	suite.Require().NoError(err)

	// Verify mage.yaml was created
	configPath := filepath.Join(projectPath, "mage.yaml")
	_, err = os.Stat(configPath)
	suite.Require().NoError(err)

	// Read and verify content
	content, err := os.ReadFile(configPath) // #nosec G304 -- controlled test file path
	suite.Require().NoError(err)
	suite.Contains(string(content), "test-project")
}

// Test getTemplateNames function
func (suite *MageInitTestSuite) TestGetTemplateNames() {
	names := getTemplateNames()
	suite.NotEmpty(names)

	// Should contain expected template names
	expectedTemplates := []string{"basic", "advanced", "cli", "web", "microservice"}
	for _, expected := range expectedTemplates {
		suite.Contains(names, expected)
	}
}

// Test getAvailableTemplates function
func (suite *MageInitTestSuite) TestGetAvailableTemplates() {
	templates := getAvailableTemplates()
	suite.NotEmpty(templates)

	// Test specific templates
	basic, exists := templates["basic"]
	suite.True(exists)
	suite.Equal("basic", basic.Name)
	suite.NotEmpty(basic.Description)
	suite.NotEmpty(basic.Files)
	suite.NotEmpty(basic.Directories)

	advanced, exists := templates["advanced"]
	suite.True(exists)
	suite.Equal("advanced", advanced.Name)
	suite.NotEmpty(advanced.Dependencies)
}

// Test template file generation functions
func (suite *MageInitTestSuite) TestGetBasicTemplateFiles() {
	files := getBasicTemplateFiles()
	suite.NotEmpty(files)

	expectedFiles := []string{"magefile.go", "README.md", ".gitignore", "cmd/main.go"}
	for _, file := range expectedFiles {
		content, exists := files[file]
		suite.True(exists, "File %s should exist", file)
		suite.NotEmpty(content, "File %s should have content", file)
	}
}

func (suite *MageInitTestSuite) TestGetBasicMagefileContent() {
	content := getBasicMagefileContent()
	suite.NotEmpty(content)
	suite.Contains(content, "//go:build mage")
	suite.Contains(content, "var Default = Build")
	suite.Contains(content, "func Build() error")
	suite.Contains(content, "func Test() error")
	suite.Contains(content, "{{.ProjectName}}")
}

func (suite *MageInitTestSuite) TestGetBasicReadmeContent() {
	content := getBasicReadmeContent()
	suite.NotEmpty(content)
	suite.Contains(content, "# {{.ProjectName}}")
	suite.Contains(content, "Go {{.GoVersion}}")
	suite.Contains(content, "magex build")
	suite.Contains(content, "magex test")
}

func (suite *MageInitTestSuite) TestGetBasicGitignoreContent() {
	content := getBasicGitignoreContent()
	suite.NotEmpty(content)
	suite.Contains(content, "*.exe")
	suite.Contains(content, "bin/")
	suite.Contains(content, ".DS_Store")
	suite.Contains(content, "go.work")
}

func (suite *MageInitTestSuite) TestGetBasicMainContent() {
	content := getBasicMainContent()
	suite.NotEmpty(content)
	suite.Contains(content, "package main")
	suite.Contains(content, "func main()")
	suite.Contains(content, "{{.ProjectName}}")
}

// Test advanced template functions
func (suite *MageInitTestSuite) TestGetAdvancedTemplateFiles() {
	files := getAdvancedTemplateFiles()
	suite.NotEmpty(files)

	// Should include basic files plus additional ones
	suite.Contains(files, "magefile.go")
	suite.Contains(files, "Makefile")
}

func (suite *MageInitTestSuite) TestGetAdvancedMagefileContent() {
	content := getAdvancedMagefileContent()
	suite.NotEmpty(content)
	suite.Contains(content, "func TestCoverage() error")
	suite.Contains(content, "func Release() error")
	suite.Contains(content, "buildForPlatform")
}

// Test CLI template functions
func (suite *MageInitTestSuite) TestGetCLITemplateFiles() {
	files := getCLITemplateFiles()
	suite.NotEmpty(files)

	suite.Contains(files, "cmd/{{.ProjectName}}/main.go")
	suite.Contains(files, "pkg/cli/cli.go")
}

// Test web template functions
func (suite *MageInitTestSuite) TestGetWebTemplateFiles() {
	files := getWebTemplateFiles()
	suite.NotEmpty(files)

	suite.Contains(files, "cmd/server/main.go")
	suite.Contains(files, "web/handlers.go")
	suite.Contains(files, "static/style.css")
	suite.Contains(files, "templates/index.html")
}

// Test microservice template functions
func (suite *MageInitTestSuite) TestGetMicroserviceTemplateFiles() {
	files := getMicroserviceTemplateFiles()
	suite.NotEmpty(files)

	suite.Contains(files, "deployments/kubernetes.yaml")
	suite.Contains(files, "internal/service/service.go")
}

func (suite *MageInitTestSuite) TestGetMicroserviceMagefileContent() {
	content := getMicroserviceMagefileContent()
	suite.NotEmpty(content)
	suite.Contains(content, "func K8sDeploy() error")
	suite.Contains(content, "kubeNamespace")
}

func (suite *MageInitTestSuite) TestGetKubernetesContent() {
	content := getKubernetesContent()
	suite.NotEmpty(content)
	suite.Contains(content, "apiVersion: apps/v1")
	suite.Contains(content, "kind: Deployment")
	suite.Contains(content, "{{.ProjectName}}")
	suite.Contains(content, "livenessProbe:")
	suite.Contains(content, "readinessProbe:")
}

func (suite *MageInitTestSuite) TestGetServiceContent() {
	content := getServiceContent()
	suite.NotEmpty(content)
	suite.Contains(content, "package service")
	suite.Contains(content, "type Service struct")
	suite.Contains(content, "func NewService")
	suite.Contains(content, "/health")
	suite.Contains(content, "/ready")
}

func (suite *MageInitTestSuite) TestGetMakefileContent() {
	content := getMakefileContent()
	suite.NotEmpty(content)
	suite.Contains(content, ".PHONY:")
	suite.Contains(content, "BINARY_NAME={{.ProjectName}}")
	suite.Contains(content, "## Build the application")
	suite.Contains(content, "@mage build")
}

// Test full initialization workflow
func (suite *MageInitTestSuite) TestInitializeProjectBasic() {
	projectPath := filepath.Join(suite.tempDir, "full-init-test")

	opts := &InitOptions{
		ProjectName: "full-test",
		ProjectPath: projectPath,
		Template:    "basic",
		GoVersion:   getDefaultGoVersion(),
		ModulePath:  "github.com/test/full-test",
		Force:       false,
		Verbose:     false,
	}

	err := initializeProject(opts)
	suite.Require().NoError(err)

	// Verify project structure was created
	suite.DirExists(projectPath)
	suite.FileExists(filepath.Join(projectPath, "go.mod"))
	suite.FileExists(filepath.Join(projectPath, "mage.yaml"))
	suite.FileExists(filepath.Join(projectPath, "magefile.go"))
	suite.FileExists(filepath.Join(projectPath, "README.md"))
	suite.FileExists(filepath.Join(projectPath, ".gitignore"))
	suite.DirExists(filepath.Join(projectPath, "cmd"))
	suite.DirExists(filepath.Join(projectPath, "pkg"))

	// Verify content was processed
	goMod, err := os.ReadFile(filepath.Join(projectPath, "go.mod")) // #nosec G304 -- controlled test file path
	suite.Require().NoError(err)
	suite.Contains(string(goMod), "module github.com/test/full-test")

	readme, err := os.ReadFile(filepath.Join(projectPath, "README.md")) // #nosec G304 -- controlled test file path
	suite.Require().NoError(err)
	suite.Contains(string(readme), "# full-test")
}

func (suite *MageInitTestSuite) TestInitializeProjectAdvanced() {
	projectPath := filepath.Join(suite.tempDir, "advanced-init-test")

	opts := &InitOptions{
		ProjectName: "advanced-test",
		ProjectPath: projectPath,
		Template:    "advanced",
		GoVersion:   getDefaultGoVersion(),
		ModulePath:  "github.com/test/advanced-test",
		Force:       false,
		Verbose:     true,
	}

	err := initializeProject(opts)
	suite.Require().NoError(err)

	// Verify advanced template files were created
	suite.FileExists(filepath.Join(projectPath, "Makefile"))
	suite.DirExists(filepath.Join(projectPath, "deployments"))
	suite.DirExists(filepath.Join(projectPath, "tests"))
}

// Helper methods for the test suite
func (suite *MageInitTestSuite) FileExists(path string) {
	_, err := os.Stat(path)
	suite.NoError(err, "File should exist: %s", path)
}

func (suite *MageInitTestSuite) DirExists(path string) {
	stat, err := os.Stat(path)
	suite.Require().NoError(err, "Directory should exist: %s", path)
	suite.True(stat.IsDir(), "Path should be a directory: %s", path)
}

// Test error conditions
func (suite *MageInitTestSuite) TestInitializeProjectInvalidPath() {
	opts := &InitOptions{
		ProjectName: "test",
		ProjectPath: "/invalid/path/that/does/not/exist/and/cannot/be/created",
		Template:    "basic",
		GoVersion:   getDefaultGoVersion(),
		ModulePath:  "test",
	}

	err := initializeProject(opts)
	suite.Require().Error(err)
}

// Test edge cases
func (suite *MageInitTestSuite) TestProcessTemplateEdgeCases() {
	opts := &InitOptions{
		ProjectName: "",
		ModulePath:  "",
		GoVersion:   "",
	}

	template := "{{.ProjectName}} {{.ModulePath}} {{.GoVersion}}"
	result := processTemplate(template, opts)
	suite.Equal("  ", result)
}

func (suite *MageInitTestSuite) TestProcessTemplateSpecialCharacters() {
	opts := &InitOptions{
		ProjectName: "test-project_v2",
		ModulePath:  "github.com/user/test-project_v2",
		GoVersion:   getDefaultGoVersion(),
	}

	template := "Project: {{.ProjectName}}, Module: {{.ModulePath}}, Version: {{.GoVersion}}"
	result := processTemplate(template, opts)
	expected := "Project: test-project_v2, Module: github.com/user/test-project_v2, Version: " + getDefaultGoVersion()
	suite.Equal(expected, result)
}

// Test constants and error types
func (suite *MageInitTestSuite) TestConstants() {
	suite.Equal("1.0.0", version)
	suite.Contains(banner, "MAGE-X Project Initialization Tool")
	suite.Contains(banner, "%s") // Version placeholder
}

func (suite *MageInitTestSuite) TestErrorTypes() {
	suite.Equal("invalid template", ErrInvalidTemplate.Error())
	suite.Equal("directory is not empty", ErrDirectoryNotEmpty.Error())
}

// Test ProjectTemplate struct
func (suite *MageInitTestSuite) TestProjectTemplateStruct() {
	template := ProjectTemplate{
		Name:         "test",
		Description:  "Test template",
		Files:        map[string]string{"test.go": "package main"},
		Directories:  []string{"cmd", "pkg"},
		Dependencies: []string{},
	}

	suite.Equal("test", template.Name)
	suite.Equal("Test template", template.Description)
	suite.Contains(template.Files, "test.go")
	suite.Contains(template.Directories, "cmd")
	suite.Empty(template.Dependencies)
}

// Test InitOptions struct
func (suite *MageInitTestSuite) TestInitOptionsStruct() {
	opts := InitOptions{
		ProjectName: "test",
		ProjectPath: "/path/to/test",
		Template:    "basic",
		GoVersion:   getDefaultGoVersion(),
		ModulePath:  "github.com/user/test",
		Force:       true,
		Verbose:     true,
		Interactive: false,
	}

	suite.Equal("test", opts.ProjectName)
	suite.Equal("/path/to/test", opts.ProjectPath)
	suite.Equal("basic", opts.Template)
	suite.Equal(getDefaultGoVersion(), opts.GoVersion)
	suite.Equal("github.com/user/test", opts.ModulePath)
	suite.True(opts.Force)
	suite.True(opts.Verbose)
	suite.False(opts.Interactive)
}
