// mage-init is a command-line tool for initializing new mage projects
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mrz1836/mage-x/pkg/common/config"
	"github.com/mrz1836/mage-x/pkg/common/env"
	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/utils"
)

var (
	// ErrInvalidTemplate is returned when an invalid template is specified
	ErrInvalidTemplate = errors.New("invalid template")
	// ErrDirectoryNotEmpty is returned when trying to initialize in a non-empty directory without force
	ErrDirectoryNotEmpty = errors.New("directory is not empty")
)

const (
	version = "1.0.0"
	banner  = `
███╗   ███╗ █████╗  ██████╗ ███████╗      ██╗  ██╗
████╗ ████║██╔══██╗██╔════╝ ██╔════╝      ╚██╗██╔╝
██╔████╔██║███████║██║  ███╗█████╗  █████╗ ╚███╔╝
██║╚██╔╝██║██╔══██║██║   ██║██╔══╝  ╚════╝ ██╔██╗
██║ ╚═╝ ██║██║  ██║╚██████╔╝███████╗      ██╔╝ ██╗
╚═╝     ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝      ╚═╝  ╚═╝

🪄 MAGE-X Project Initialization Tool v%s
   Write Once, Mage Everywhere
`
)

// ProjectTemplate represents a project template configuration
type ProjectTemplate struct {
	Name         string
	Description  string
	Files        map[string]string
	Directories  []string
	Dependencies []string
}

// InitOptions holds configuration for project initialization
type InitOptions struct {
	ProjectName string
	ProjectPath string
	Template    string
	GoVersion   string
	ModulePath  string
	Force       bool
	Verbose     bool
	Interactive bool
}

// getDefaultGoVersion returns the default Go version from environment or fallback
func getDefaultGoVersion() string {
	if version := mage.GetDefaultGoVersion(); version != "" {
		return version
	}
	// Fallback if environment is not set
	return "1.24"
}

func main() {
	fmt.Printf(banner, version)

	var opts InitOptions
	flag.StringVar(&opts.ProjectName, "name", "", "Project name")
	flag.StringVar(&opts.ProjectPath, "path", ".", "Project path (default: current directory)")
	flag.StringVar(&opts.Template, "template", "basic", "Project template (basic, advanced, enterprise, cli, web, microservice)")
	flag.StringVar(&opts.GoVersion, "go-version", getDefaultGoVersion(), "Go version to use")
	flag.StringVar(&opts.ModulePath, "module", "", "Go module path (e.g., github.com/user/project)")
	flag.BoolVar(&opts.Force, "force", false, "Force initialization even if directory is not empty")
	flag.BoolVar(&opts.Verbose, "verbose", false, "Verbose output")
	flag.BoolVar(&opts.Interactive, "interactive", false, "Interactive mode")

	showHelp := flag.Bool("help", false, "Show help")
	showVersion := flag.Bool("version", false, "Show version")
	listTemplates := flag.Bool("list-templates", false, "List available templates")

	flag.Parse()

	// Load environment variables from .env files (early startup hook)
	// This loads .github/.env.base and other env files before tool version checks
	if err := env.LoadStartupEnv(); err != nil && opts.Verbose {
		fmt.Printf("Debug: Could not load env files: %v\n", err)
	}

	if *showHelp {
		showHelpMessage()
		return
	}

	if *showVersion {
		fmt.Printf("mage-init version %s\n", version)
		return
	}

	if *listTemplates {
		listAvailableTemplates()
		return
	}

	// Validate options
	if err := validateOptions(&opts); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Interactive mode
	if opts.Interactive {
		if err := runInteractiveMode(&opts); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Interactive mode failed: %v\n", err)
			os.Exit(1)
		}
	}

	// Initialize project
	if err := initializeProject(&opts); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Project initialization failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✅ Successfully initialized magex project '%s'\n", opts.ProjectName)
	fmt.Printf("📁 Project location: %s\n", opts.ProjectPath)
	fmt.Printf("🚀 Next steps:\n")
	fmt.Printf("   cd %s\n", opts.ProjectPath)
	fmt.Printf("   go run magefile.go\n")
}

// validateOptions validates the initialization options
func validateOptions(opts *InitOptions) error {
	if opts.ProjectName == "" {
		if opts.ProjectPath == "." {
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}
			opts.ProjectName = filepath.Base(cwd)
		} else {
			opts.ProjectName = filepath.Base(opts.ProjectPath)
		}
	}

	if opts.ModulePath == "" {
		opts.ModulePath = opts.ProjectName
	}

	// Validate template
	templates := getAvailableTemplates()
	if _, exists := templates[opts.Template]; !exists {
		return fmt.Errorf("%w '%s'. Available: %s", ErrInvalidTemplate, opts.Template, strings.Join(getTemplateNames(), ", "))
	}

	return nil
}

// runInteractiveMode runs the interactive project initialization
func runInteractiveMode(opts *InitOptions) error {
	utils.Info("🎯 Interactive Project Setup")
	utils.Info("=============================")

	// Project name
	if opts.ProjectName == "" {
		fmt.Print("Enter project name: ")
		var name string
		if _, err := fmt.Scanln(&name); err != nil {
			return fmt.Errorf("failed to read project name: %w", err)
		}
		opts.ProjectName = name
	}

	// Module path
	fmt.Printf("Enter Go module path [%s]: ", opts.ModulePath)
	var module string
	if _, err := fmt.Scanln(&module); err != nil {
		return fmt.Errorf("failed to read module path: %w", err)
	}
	if module != "" {
		opts.ModulePath = module
	}

	// Template selection
	utils.Info("Available templates:")
	templates := getAvailableTemplates()
	for name, template := range templates {
		fmt.Printf("  %s - %s\n", name, template.Description)
	}
	fmt.Printf("Select template [%s]: ", opts.Template)
	var template string
	if _, err := fmt.Scanln(&template); err != nil {
		return fmt.Errorf("failed to read template: %w", err)
	}
	if template != "" {
		opts.Template = template
	}

	// Go version
	fmt.Printf("Go version [%s]: ", opts.GoVersion)
	var goVersion string
	if _, err := fmt.Scanln(&goVersion); err != nil {
		return fmt.Errorf("failed to read Go version: %w", err)
	}
	if goVersion != "" {
		opts.GoVersion = goVersion
	}

	return nil
}

// initializeProject creates the project structure and files
func initializeProject(opts *InitOptions) error {
	if opts.Verbose {
		fmt.Printf("🔧 Initializing project '%s' with template '%s'\n", opts.ProjectName, opts.Template)
	}

	// Create project directory
	projectPath, err := filepath.Abs(opts.ProjectPath)
	if err != nil {
		return fmt.Errorf("failed to resolve project path: %w", err)
	}

	if err := ensureProjectDirectory(projectPath, opts.Force); err != nil {
		return err
	}

	// Get template
	template := getAvailableTemplates()[opts.Template]

	// Create directory structure
	if err := createDirectories(projectPath, template.Directories, opts.Verbose); err != nil {
		return err
	}

	// Create files
	if err := createFiles(projectPath, template.Files, opts, opts.Verbose); err != nil {
		return err
	}

	// Initialize go.mod
	if err := initializeGoMod(projectPath, opts); err != nil {
		return err
	}

	// Create mage configuration
	if err := createMageConfig(projectPath, opts); err != nil {
		return err
	}

	return nil
}

// ensureProjectDirectory ensures the project directory exists and is valid for initialization
func ensureProjectDirectory(projectPath string, force bool) error {
	fileOps := fileops.New()

	if !fileOps.File.Exists(projectPath) {
		return fileOps.File.MkdirAll(projectPath, 0o755)
	}

	// Check if directory is empty
	entries, err := os.ReadDir(projectPath)
	if err != nil {
		return fmt.Errorf("failed to read project directory: %w", err)
	}

	if len(entries) > 0 && !force {
		return fmt.Errorf("%w '%s'. Use --force to initialize anyway", ErrDirectoryNotEmpty, projectPath)
	}

	return nil
}

// createDirectories creates the directory structure
func createDirectories(projectPath string, directories []string, verbose bool) error {
	fileOps := fileops.New()

	for _, dir := range directories {
		dirPath := filepath.Join(projectPath, dir)
		if verbose {
			fmt.Printf("📁 Creating directory: %s\n", dir)
		}
		if err := fileOps.File.MkdirAll(dirPath, 0o755); err != nil {
			return fmt.Errorf("failed to create directory '%s': %w", dir, err)
		}
	}
	return nil
}

// createFiles creates the project files from templates
func createFiles(projectPath string, files map[string]string, opts *InitOptions, verbose bool) error {
	fileOps := fileops.New()

	for filePath, content := range files {
		fullPath := filepath.Join(projectPath, filePath)

		if verbose {
			fmt.Printf("📄 Creating file: %s\n", filePath)
		}

		// Process template variables
		processedContent := processTemplate(content, opts)

		if err := fileOps.File.WriteFile(fullPath, []byte(processedContent), 0o644); err != nil {
			return fmt.Errorf("failed to create file '%s': %w", filePath, err)
		}
	}

	return nil
}

// processTemplate replaces template variables with actual values
func processTemplate(content string, opts *InitOptions) string {
	replacements := map[string]string{
		"{{.ProjectName}}": opts.ProjectName,
		"{{.ModulePath}}":  opts.ModulePath,
		"{{.GoVersion}}":   opts.GoVersion,
	}

	result := content
	for placeholder, value := range replacements {
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result
}

// initializeGoMod creates the go.mod file
func initializeGoMod(projectPath string, opts *InitOptions) error {
	fileOps := fileops.New()
	goModPath := filepath.Join(projectPath, "go.mod")

	content := fmt.Sprintf(`module %s

go %s

require (
	github.com/magefile/mage v1.15.0
)
`, opts.ModulePath, opts.GoVersion)

	return fileOps.File.WriteFile(goModPath, []byte(content), 0o644)
}

// createMageConfig creates the mage configuration file
func createMageConfig(projectPath string, opts *InitOptions) error {
	configPath := filepath.Join(projectPath, "mage.yaml")

	mageConfig := &config.MageConfig{
		Project: config.ProjectConfig{
			Name:        opts.ProjectName,
			Version:     "1.0.0",
			Description: fmt.Sprintf("Mage build automation for %s", opts.ProjectName),
		},
		Build: config.BuildConfig{
			GoVersion: opts.GoVersion,
			Platform:  "linux/amd64",
		},
		Test: config.TestConfig{
			Timeout:  120,
			Coverage: true,
			Parallel: 4,
		},
		Analytics: config.AnalyticsConfig{
			Enabled: false,
		},
	}

	fileOps := fileops.New()
	return fileOps.SaveConfig(configPath, mageConfig, "yaml")
}

// showHelpMessage displays the help message
func showHelpMessage() {
	utils.Info(`
mage-init - Initialize a new mage project

USAGE:
    mage-init [OPTIONS]

OPTIONS:
    -name string          Project name
    -path string          Project path (default: current directory)
    -template string      Project template (default: basic)
    -go-version string    Go version (default: 1.24)
    -module string        Go module path
    -force               Force initialization in non-empty directory
    -verbose             Verbose output
    -interactive         Interactive mode
    -help                Show this help message
    -version             Show version
    -list-templates      List available templates

TEMPLATES:
    basic         Simple project with basic mage targets
    advanced      Advanced project with comprehensive targets
    enterprise    Enterprise project with security and compliance
    cli           Command-line application project
    web           Web application project
    microservice  Microservice project

EXAMPLES:
    # Initialize in current directory
    mage-init -name my-project

    # Initialize with specific template
    mage-init -name my-project -template advanced

    # Initialize interactively
    mage-init -interactive

    # Initialize with custom module path
    mage-init -name my-project -module github.com/user/my-project`)
}

// listAvailableTemplates displays all available templates
func listAvailableTemplates() {
	utils.Info("📋 Available Project Templates")
	utils.Info("===============================")

	templates := getAvailableTemplates()
	for name, template := range templates {
		fmt.Printf("\n🎯 %s\n", name)
		fmt.Printf("   %s\n", template.Description)

		if len(template.Dependencies) > 0 {
			fmt.Printf("   Dependencies: %s\n", strings.Join(template.Dependencies, ", "))
		}
	}
}

// getTemplateNames returns a list of template names
func getTemplateNames() []string {
	templates := getAvailableTemplates()
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	return names
}

// getAvailableTemplates returns all available project templates
func getAvailableTemplates() map[string]ProjectTemplate {
	return map[string]ProjectTemplate{
		"basic": {
			Name:        "basic",
			Description: "Simple project with basic mage targets",
			Directories: []string{"cmd", "pkg", "scripts", "docs"},
			Files:       getBasicTemplateFiles(),
		},
		"advanced": {
			Name:         "advanced",
			Description:  "Advanced project with comprehensive build automation",
			Directories:  []string{"cmd", "pkg", "internal", "scripts", "docs", "tests", "deployments"},
			Files:        getAdvancedTemplateFiles(),
			Dependencies: []string{"docker", "kubernetes"},
		},
		"enterprise": {
			Name:         "enterprise",
			Description:  "Enterprise project with security and compliance features",
			Directories:  []string{"cmd", "pkg", "internal", "scripts", "docs", "tests", "deployments", "security", "compliance"},
			Files:        getEnterpriseTemplateFiles(),
			Dependencies: []string{"docker", "kubernetes", "vault", "prometheus"},
		},
		"cli": {
			Name:        "cli",
			Description: "Command-line application project",
			Directories: []string{"cmd", "pkg", "scripts", "docs"},
			Files:       getCLITemplateFiles(),
		},
		"web": {
			Name:         "web",
			Description:  "Web application project",
			Directories:  []string{"cmd", "pkg", "web", "static", "templates", "scripts", "docs"},
			Files:        getWebTemplateFiles(),
			Dependencies: []string{"docker"},
		},
		"microservice": {
			Name:         "microservice",
			Description:  "Microservice project with containerization",
			Directories:  []string{"cmd", "pkg", "internal", "scripts", "docs", "deployments"},
			Files:        getMicroserviceTemplateFiles(),
			Dependencies: []string{"docker", "kubernetes"},
		},
	}
}

// getBasicTemplateFiles returns files for the basic template
func getBasicTemplateFiles() map[string]string {
	return map[string]string{
		"magefile.go": getBasicMagefileContent(),
		"README.md":   getBasicReadmeContent(),
		".gitignore":  getBasicGitignoreContent(),
		"cmd/main.go": getBasicMainContent(),
	}
}

// getBasicMagefileContent returns the basic magefile content
func getBasicMagefileContent() string {
	return `//go:build mage
// +build mage

package main

import (
	"fmt"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Default target to run when none is specified
// If not set, running mage will list available targets
var Default = Build

// Build builds the application
func Build() error {
	fmt.Println("Building {{.ProjectName}}...")
	return sh.Run("go", "build", "-o", "bin/{{.ProjectName}}", "./cmd")
}

// Test runs the test suite
func Test() error {
	fmt.Println("Running tests...")
	return sh.Run("go", "test", "-v", "./...")
}

// Clean removes build artifacts
func Clean() error {
	fmt.Println("Cleaning build artifacts...")
	return sh.Rm("bin")
}

// Install installs the application
func Install() error {
	mg.Deps(Build)
	fmt.Println("Installing {{.ProjectName}}...")
	return sh.Run("go", "install", "./cmd")
}

// Lint runs the linter
func Lint() error {
	fmt.Println("Running linter...")
	return sh.Run("golangci-lint", "run")
}

// Format formats the code
func Format() error {
	fmt.Println("Formatting code...")
	return sh.Run("go", "fmt", "./...")
}
`
}

// getBasicReadmeContent returns the basic README content
func getBasicReadmeContent() string {
	return `# {{.ProjectName}}

A Go project built with Mage.

## Getting Started

### Prerequisites

- Go {{.GoVersion}} or later
- [Mage](https://magefile.org/) build tool

### Installation

` + "```bash" + `
go mod download
` + "```" + `

### Building

` + "```bash" + `
# Build the application
magex build

# Run tests
magex test

# Clean build artifacts
magex clean
` + "```" + `

### Available Mage Targets

- ` + "`build`" + ` - Build the application
- ` + "`test`" + ` - Run tests
- ` + "`clean`" + ` - Remove build artifacts
- ` + "`install`" + ` - Install the application
- ` + "`lint`" + ` - Run linter
- ` + "`format`" + ` - Format code

## Development

To see all available mage targets:

` + "```bash" + `
magex help  # Beautiful categorized list
magex -l    # Plain text list
` + "```" + `

To run a specific target:

` + "```bash" + `
magex <target>
` + "```" + `
`
}

// getBasicGitignoreContent returns the basic .gitignore content
func getBasicGitignoreContent() string {
	return `# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with go test -c
*.test

# Output of the go coverage tool
*.out

# Dependency directories (remove the comment below to include it)
# vendor/

# Go workspace file
go.work

# Build artifacts
bin/
dist/

# IDE files
.vscode/
.idea/
*.swp
*.swo
*~

# OS generated files
.DS_Store
.DS_Store?
._*
.Spotlight-V100
.Trashes
ehthumbs.db
Thumbs.db

# Temporary files
*.tmp
*.temp
`
}

// getBasicMainContent returns the basic main.go content
func getBasicMainContent() string {
	return `package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("Hello from {{.ProjectName}}!")

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	// TODO: Implement your application logic here
	fmt.Println("Application is running...")
	return nil
}
`
}

// getAdvancedTemplateFiles returns files for the advanced template
func getAdvancedTemplateFiles() map[string]string {
	files := getBasicTemplateFiles()
	files["magefile.go"] = getAdvancedMagefileContent()
	files["Dockerfile"] = getDockerfileContent()
	files["docker-compose.yml"] = getDockerComposeContent()
	files["Makefile"] = getMakefileContent()
	return files
}

// getEnterpriseTemplateFiles returns files for the enterprise template
func getEnterpriseTemplateFiles() map[string]string {
	files := getAdvancedTemplateFiles()
	files["magefile.go"] = getEnterpriseMagefileContent()
	files["security/security.yaml"] = getSecurityConfigContent()
	files["compliance/compliance.yaml"] = getComplianceConfigContent()
	return files
}

// getCLITemplateFiles returns files for the CLI template
func getCLITemplateFiles() map[string]string {
	files := getBasicTemplateFiles()
	files["cmd/{{.ProjectName}}/main.go"] = getCLIMainContent()
	files["pkg/cli/cli.go"] = getCLIPackageContent()
	return files
}

// getWebTemplateFiles returns files for the web template
func getWebTemplateFiles() map[string]string {
	files := getBasicTemplateFiles()
	files["magefile.go"] = getWebMagefileContent()
	files["cmd/server/main.go"] = getWebServerContent()
	files["web/handlers.go"] = getWebHandlersContent()
	files["static/style.css"] = getWebCSSContent()
	files["templates/index.html"] = getWebTemplateContent()
	return files
}

// getMicroserviceTemplateFiles returns files for the microservice template
func getMicroserviceTemplateFiles() map[string]string {
	files := getAdvancedTemplateFiles()
	files["magefile.go"] = getMicroserviceMagefileContent()
	files["deployments/kubernetes.yaml"] = getKubernetesContent()
	files["internal/service/service.go"] = getServiceContent()
	return files
}

// getMicroserviceMagefileContent returns microservice magefile content
func getMicroserviceMagefileContent() string {
	return `//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	binaryName = "{{.ProjectName}}"
	buildDir   = "bin"
	distDir    = "dist"
	dockerRepo = "{{.ProjectName}}"
	kubeNamespace = "{{.ProjectName}}"
)

var Default = Build

// Build builds the microservice
func Build() error {
	mg.Deps(Format, Lint, Test)
	fmt.Println("Building microservice", binaryName)
	return sh.Run("go", "build", "-ldflags", buildLdflags(), "-o", filepath.Join(buildDir, binaryName), "./cmd")
}

// Test runs all tests with coverage
func Test() error {
	fmt.Println("Running microservice tests...")
	return sh.Run("go", "test", "-v", "-coverprofile=coverage.out", "-race", "./...")
}

// Lint runs linter
func Lint() error {
	fmt.Println("Running linter...")
	return sh.Run("golangci-lint", "run", "--timeout=5m")
}

// Format formats code
func Format() error {
	fmt.Println("Formatting code...")
	return sh.Run("go", "fmt", "./...")
}

// Clean removes build artifacts
func Clean() error {
	fmt.Println("Cleaning...")
	if err := sh.Rm(buildDir); err != nil {
		fmt.Printf("Warning: failed to remove %s: %v\n", buildDir, err)
	}
	if err := sh.Rm(distDir); err != nil {
		fmt.Printf("Warning: failed to remove %s: %v\n", distDir, err)
	}
	if err := sh.Rm("coverage.out"); err != nil {
		fmt.Printf("Warning: failed to remove coverage.out: %v\n", err)
	}
	return nil
}

// Docker builds Docker image
func Docker() error {
	fmt.Println("Building Docker image...")
	return sh.Run("docker", "build", "-t", dockerRepo+":latest", ".")
}

// K8sDeploy deploys to Kubernetes
func K8sDeploy() error {
	mg.Deps(Docker)
	fmt.Println("Deploying to Kubernetes...")
	return sh.Run("kubectl", "apply", "-f", "deployments/", "-n", kubeNamespace)
}

// Dev runs in development mode
func Dev() error {
	fmt.Println("Running in development mode...")
	return sh.Run("go", "run", "./cmd", "-dev")
}

func buildLdflags() string {
	return fmt.Sprintf("-X main.version=%s -X main.buildTime=%s",
		getVersion(), getBuildTime())
}

func getVersion() string {
	if version := os.Getenv("MAGE_X_VERSION"); version != "" {
		return version
	}
	return "dev"
}

func getBuildTime() string {
	return time.Now().Format(time.RFC3339)
}
`
}

// getKubernetesContent returns Kubernetes deployment content
func getKubernetesContent() string {
	return `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{.ProjectName}}
  labels:
    app: {{.ProjectName}}
spec:
  replicas: 3
  selector:
    matchLabels:
      app: {{.ProjectName}}
  template:
    metadata:
      labels:
        app: {{.ProjectName}}
    spec:
      containers:
      - name: {{.ProjectName}}
        image: {{.ProjectName}}:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENV
          value: "production"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
---
apiVersion: v1
kind: Service
metadata:
  name: {{.ProjectName}}-service
spec:
  selector:
    app: {{.ProjectName}}
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP
`
}

// getServiceContent returns microservice service content
func getServiceContent() string {
	return `package service

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Service represents the microservice
type Service struct {
	server *http.Server
	config Config
}

// Config holds service configuration
type Config struct {
	Port    string
	Host    string
	Timeout time.Duration
}

// NewService creates a new service instance
func NewService(config Config) *Service {
	mux := http.NewServeMux()

	// Health endpoints
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/ready", readyHandler)

	// API endpoints
	mux.HandleFunc("/api/v1/", apiHandler)

	server := &http.Server{
		Addr:         config.Host + ":" + config.Port,
		Handler:      mux,
		ReadTimeout:  config.Timeout,
		WriteTimeout: config.Timeout,
		IdleTimeout:  config.Timeout * 2,
	}

	return &Service{
		server: server,
		config: config,
	}
}

// Start starts the service
func (s *Service) Start() error {
	log.Printf("Starting service on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop gracefully stops the service
func (s *Service) Stop(ctx context.Context) error {
	log.Println("Stopping service...")
	return s.server.Shutdown(ctx)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, ` + "`" + `{"status":"healthy","timestamp":"%s"}` + "`" + `, time.Now().Format(time.RFC3339))
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, ` + "`" + `{"status":"ready","timestamp":"%s"}` + "`" + `, time.Now().Format(time.RFC3339))
}

func apiHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, ` + "`" + `{"message":"Hello from {{.ProjectName}} microservice","path":"%s"}` + "`" + `, r.URL.Path)
	case http.MethodPost:
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, ` + "`" + `{"message":"Resource created","path":"%s"}` + "`" + `, r.URL.Path)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprintf(w, ` + "`" + `{"error":"Method not allowed"}` + "`" + `)
	}
}
`
}

// getAdvancedMagefileContent returns the advanced magefile content
func getAdvancedMagefileContent() string {
	return `//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	binaryName = "{{.ProjectName}}"
	buildDir   = "bin"
	distDir    = "dist"
	dockerRepo = "{{.ProjectName}}"
)

var Default = Build

// Build builds the application
func Build() error {
	mg.Deps(Format, Lint)
	fmt.Println("Building", binaryName)
	return sh.Run("go", "build", "-ldflags", buildLdflags(), "-o", filepath.Join(buildDir, binaryName), "./cmd")
}

// Test runs all tests with coverage
func Test() error {
	fmt.Println("Running tests with coverage...")
	return sh.Run("go", "test", "-v", "-coverprofile=coverage.out", "./...")
}

// TestCoverage shows test coverage report
func TestCoverage() error {
	mg.Deps(Test)
	return sh.Run("go", "tool", "cover", "-html=coverage.out")
}

// Lint runs golangci-lint
func Lint() error {
	fmt.Println("Running linter...")
	return sh.Run("golangci-lint", "run", "--timeout=5m")
}

// Format formats the code
func Format() error {
	fmt.Println("Formatting code...")
	err := sh.Run("go", "fmt", "./...")
	if err != nil {
		return err
	}
	return sh.Run("goimports", "-w", ".")
}

// Clean removes build artifacts
func Clean() error {
	fmt.Println("Cleaning...")
	if err := sh.Rm(buildDir); err != nil {
		fmt.Printf("Warning: failed to remove %s: %v\n", buildDir, err)
	}
	if err := sh.Rm(distDir); err != nil {
		fmt.Printf("Warning: failed to remove %s: %v\n", distDir, err)
	}
	if err := sh.Rm("coverage.out"); err != nil {
		fmt.Printf("Warning: failed to remove coverage.out: %v\n", err)
	}
	return nil
}

// Install installs the application
func Install() error {
	mg.Deps(Build)
	fmt.Println("Installing", binaryName)
	return sh.Run("go", "install", "./cmd")
}

// Docker builds Docker image
func Docker() error {
	fmt.Println("Building Docker image...")
	return sh.Run("docker", "build", "-t", dockerRepo+":latest", ".")
}

// DockerRun runs the Docker container
func DockerRun() error {
	mg.Deps(Docker)
	fmt.Println("Running Docker container...")
	return sh.Run("docker", "run", "--rm", "-p", "8080:8080", dockerRepo+":latest")
}

// Release builds release artifacts
func Release() error {
	mg.Deps(Clean, Test)
	fmt.Println("Building release artifacts...")

	platforms := []string{
		"linux/amd64",
		"linux/arm64",
		"darwin/amd64",
		"darwin/arm64",
		"windows/amd64",
	}

	for _, platform := range platforms {
		if err := buildForPlatform(platform); err != nil {
			return err
		}
	}

	return nil
}

// Dev runs the application in development mode
func Dev() error {
	fmt.Println("Running in development mode...")
	return sh.Run("go", "run", "./cmd", "-dev")
}

// Generate runs go generate
func Generate() error {
	fmt.Println("Running go generate...")
	return sh.Run("go", "generate", "./...")
}

// Vendor updates vendor directory
func Vendor() error {
	fmt.Println("Updating vendor directory...")
	return sh.Run("go", "mod", "vendor")
}

// ModTidy runs go mod tidy
func ModTidy() error {
	fmt.Println("Running go mod tidy...")
	return sh.Run("go", "mod", "tidy")
}

// Security runs security checks
func Security() error {
	fmt.Println("Running security checks...")
	return sh.Run("gosec", "./...")
}

func buildLdflags() string {
	return fmt.Sprintf("-X main.version=%s -X main.buildTime=%s",
		getVersion(), getBuildTime())
}

func buildForPlatform(platform string) error {
	parts := strings.Split(platform, "/")
	goos, goarch := parts[0], parts[1]

	output := filepath.Join(distDir, fmt.Sprintf("%s-%s-%s", binaryName, goos, goarch))
	if goos == "windows" {
		output += ".exe"
	}

	env := map[string]string{
		"GOOS":   goos,
		"GOARCH": goarch,
		"CGO_ENABLED": "0",
	}

	fmt.Printf("Building for %s/%s...\n", goos, goarch)
	return sh.RunWith(env, "go", "build", "-ldflags", buildLdflags(), "-o", output, "./cmd")
}

func getVersion() string {
	if version := os.Getenv("MAGE_X_VERSION"); version != "" {
		return version
	}
	return "dev"
}

func getBuildTime() string {
	return time.Now().Format(time.RFC3339)
}
`
}

// getDockerfileContent returns Dockerfile content
func getDockerfileContent() string {
	return `# Build stage
FROM golang:{{.GoVersion}}-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/{{.ProjectName}} ./cmd

# Final stage
FROM alpine:latest

WORKDIR /root/

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Copy the binary from builder stage
COPY --from=builder /app/bin/{{.ProjectName}} .

# Create non-root user
RUN addgroup -g 1001 -S {{.ProjectName}} && \
    adduser -u 1001 -S {{.ProjectName}} -G {{.ProjectName}}

USER {{.ProjectName}}

EXPOSE 8080

CMD ["./{{.ProjectName}}"]
`
}

// getDockerComposeContent returns docker-compose.yml content
func getDockerComposeContent() string {
	return `version: '3.8'

services:
  {{.ProjectName}}:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENV=docker
    volumes:
      - ./config:/app/config:ro
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  # Add other services as needed
  # redis:
  #   image: redis:7-alpine
  #   ports:
  #     - "6379:6379"
  #   restart: unless-stopped

  # postgres:
  #   image: postgres:15-alpine
  #   environment:
  #     POSTGRES_DB: {{.ProjectName}}
  #     POSTGRES_USER: user
  #     POSTGRES_PASSWORD: password
  #   ports:
  #     - "5432:5432"
  #   volumes:
  #     - postgres_data:/var/lib/postgresql/data
  #   restart: unless-stopped

# volumes:
#   postgres_data:
`
}

// getMakefileContent returns Makefile content
func getMakefileContent() string {
	return `.PHONY: build test clean install lint format docker help

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME={{.ProjectName}}
BUILD_DIR=bin
GO_VERSION={{.GoVersion}}

## Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mage build

## Run tests
test:
	@echo "Running tests..."
	@mage test

## Clean build artifacts
clean:
	@echo "Cleaning..."
	@mage clean

## Install the application
install:
	@echo "Installing $(BINARY_NAME)..."
	@mage install

## Run linter
lint:
	@echo "Running linter..."
	@mage lint

## Format code
format:
	@echo "Formatting code..."
	@mage format

## Build Docker image
docker:
	@echo "Building Docker image..."
	@mage docker

## Run in development mode
dev:
	@echo "Running in development mode..."
	@mage dev

## Show help
help:
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*##"; printf "\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
`
}
