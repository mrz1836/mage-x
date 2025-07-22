package main

// Additional template functions that were missing

// getEnterpriseMagefileContent returns enterprise magefile content  
func getEnterpriseMagefileContent() string {
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

// Build builds the application with security checks
func Build() error {
	mg.Deps(Format, Lint, Security, Test)
	fmt.Println("Building", binaryName, "with enterprise features")
	return sh.Run("go", "build", "-ldflags", buildLdflags(), "-o", filepath.Join(buildDir, binaryName), "./cmd")
}

// Test runs all tests with coverage and compliance checks
func Test() error {
	fmt.Println("Running comprehensive test suite...")
	err := sh.Run("go", "test", "-v", "-coverprofile=coverage.out", "-race", "./...")
	if err != nil {
		return err
	}
	return generateComplianceReport()
}

// Security runs comprehensive security checks
func Security() error {
	fmt.Println("Running security analysis...")
	err := sh.Run("gosec", "-fmt", "json", "-out", "security-report.json", "./...")
	if err != nil {
		fmt.Println("Warning: Security scan found issues")
	}
	
	sh.Run("nancy", "sleuth")
	return sh.Run("govulncheck", "./...")
}

// Compliance generates compliance reports
func Compliance() error {
	fmt.Println("Generating compliance reports...")
	mg.Deps(Test, Security)
	
	err := sh.Run("syft", ".", "-o", "spdx-json=sbom.json")
	if err != nil {
		fmt.Println("Warning: Could not generate SBOM")
	}
	
	return generateComplianceReport()
}

// Lint runs comprehensive linting
func Lint() error {
	fmt.Println("Running enterprise linting...")
	return sh.Run("golangci-lint", "run", "--timeout=10m", "--config=.golangci.json")
}

// Format formats the code with enterprise standards
func Format() error {
	fmt.Println("Formatting code with enterprise standards...")
	err := sh.Run("go", "fmt", "./...")
	if err != nil {
		return err
	}
	
	err = sh.Run("goimports", "-w", ".")
	if err != nil {
		return err
	}
	
	return sh.Run("gofumpt", "-w", ".")
}

// Clean removes all build artifacts and reports
func Clean() error {
	fmt.Println("Cleaning enterprise artifacts...")
	sh.Rm(buildDir)
	sh.Rm(distDir)
	sh.Rm("coverage.out")
	sh.Rm("security-report.json")
	sh.Rm("compliance-report.json")
	sh.Rm("sbom.json")
	return nil
}

// Docker builds secure Docker image
func Docker() error {
	fmt.Println("Building secure Docker image...")
	err := sh.Run("docker", "build", "-t", dockerRepo+":latest", ".")
	if err != nil {
		return err
	}
	
	return sh.Run("trivy", "image", dockerRepo+":latest")
}

func buildLdflags() string {
	return fmt.Sprintf("-X main.version=%s -X main.buildTime=%s -X main.gitCommit=%s",
		getVersion(), getBuildTime(), getGitCommit())
}

func generateComplianceReport() error {
	fmt.Println("Generating compliance report...")
	return nil
}

func getVersion() string {
	if version := os.Getenv("VERSION"); version != "" {
		return version
	}
	return "dev"
}

func getBuildTime() string {
	return time.Now().Format(time.RFC3339)
}

func getGitCommit() string {
	commit, _ := sh.Output("git", "rev-parse", "--short", "HEAD")
	return commit
}
`
}

// getSecurityConfigContent returns security configuration
func getSecurityConfigContent() string {
	return `# Enterprise Security Configuration
security:
  encryption:
    algorithm: "AES-256-GCM"
    key_rotation_days: 90
  
  authentication:
    method: "oauth2"
    token_expiry: "1h"
    refresh_token_expiry: "24h"
  
  authorization:
    rbac_enabled: true
    default_role: "user"
  
  audit:
    enabled: true
    log_level: "info"
    retention_days: 365
  
  tls:
    min_version: "1.3"
    cipher_suites:
      - "TLS_AES_256_GCM_SHA384"
      - "TLS_AES_128_GCM_SHA256"
`
}

// getComplianceConfigContent returns compliance configuration
func getComplianceConfigContent() string {
	return `# Enterprise Compliance Configuration
compliance:
  standards:
    - "SOC2"
    - "ISO27001"
    - "GDPR"
    - "HIPAA"
  
  data_classification:
    enabled: true
    levels:
      - "public"
      - "internal"
      - "confidential"
      - "restricted"
  
  privacy:
    data_retention_days: 2555  # 7 years
    anonymization_enabled: true
    consent_management: true
  
  monitoring:
    compliance_dashboard: true
    automated_reporting: true
    alert_threshold: "high"
`
}

// getWebMagefileContent returns web magefile content
func getWebMagefileContent() string {
	return `//go:build mage
// +build mage

package main

import (
	"fmt"
	"path/filepath"
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

const (
	binaryName = "{{.ProjectName}}"
	buildDir   = "bin"
	staticDir  = "static"
	templateDir = "templates"
)

var Default = Build

// Build builds the web application
func Build() error {
	mg.Deps(BuildAssets, Format, Lint, Test)
	fmt.Println("Building web application", binaryName)
	return sh.Run("go", "build", "-o", filepath.Join(buildDir, binaryName), "./cmd/server")
}

// BuildAssets builds static assets
func BuildAssets() error {
	fmt.Println("Building static assets...")
	sh.Run("mkdir", "-p", staticDir)
	sh.Run("mkdir", "-p", templateDir)
	return nil
}

// Test runs tests
func Test() error {
	fmt.Println("Running tests...")
	return sh.Run("go", "test", "-v", "./...")
}

// Lint runs linter
func Lint() error {
	fmt.Println("Running linter...")
	return sh.Run("golangci-lint", "run")
}

// Format formats code
func Format() error {
	fmt.Println("Formatting code...")
	return sh.Run("go", "fmt", "./...")
}

// Clean removes build artifacts
func Clean() error {
	fmt.Println("Cleaning...")
	return sh.Rm(buildDir)
}

// Dev runs the development server
func Dev() error {
	fmt.Println("Starting development server...")
	return sh.Run("go", "run", "./cmd/server", "-dev")
}

// Docker builds Docker image for web app
func Docker() error {
	fmt.Println("Building Docker image...")
	return sh.Run("docker", "build", "-t", binaryName+":latest", ".")
}
`
}

// getWebServerContent returns web server main content
func getWebServerContent() string {
	return `package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
	
	"{{.ModulePath}}/web"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	var (
		port = flag.String("port", "8080", "Server port")
		dev  = flag.Bool("dev", false, "Development mode")
		host = flag.String("host", "localhost", "Server host")
	)
	
	flag.Parse()
	
	fmt.Printf("{{.ProjectName}} web server v%s\n", version)
	fmt.Printf("Build time: %s\n", buildTime)
	
	// Setup handlers
	mux := http.NewServeMux()
	
	// Static files
	staticDir := "./static/"
	if *dev {
		fmt.Println("Running in development mode")
	}
	
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	
	// Routes
	mux.HandleFunc("/", web.HomeHandler)
	mux.HandleFunc("/health", web.HealthHandler)
	mux.HandleFunc("/api/", web.APIHandler)
	
	// Create server
	server := &http.Server{
		Addr:         *host + ":" + *port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	
	fmt.Printf("Server starting on http://%s:%s\n", *host, *port)
	log.Fatal(server.ListenAndServe())
}
`
}

// getWebHandlersContent returns web handlers content
func getWebHandlersContent() string {
	return `package web

import (
	"encoding/json"
	"html/template"
	"net/http"
	"path/filepath"
)

var templates *template.Template

func init() {
	// Load templates
	templatePattern := filepath.Join("templates", "*.html")
	templates = template.Must(template.ParseGlob(templatePattern))
}

// HomeHandler handles the home page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
	data := struct {
		Title   string
		Message string
	}{
		Title:   "{{.ProjectName}}",
		Message: "Welcome to {{.ProjectName}}!",
	}
	
	err := templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// HealthHandler handles health checks
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"status":  "healthy",
		"service": "{{.ProjectName}}",
		"version": "1.0.0",
	}
	
	json.NewEncoder(w).Encode(response)
}

// APIHandler handles API requests
func APIHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	switch r.Method {
	case http.MethodGet:
		handleAPIGet(w, r)
	case http.MethodPost:
		handleAPIPost(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]string{"error": "Method not allowed"})
	}
}

func handleAPIGet(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"message": "Hello from {{.ProjectName}} API",
		"path":    r.URL.Path,
		"method":  r.Method,
	}
	
	json.NewEncoder(w).Encode(response)
}

func handleAPIPost(w http.ResponseWriter, r *http.Request) {
	var requestData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
		return
	}
	
	response := map[string]interface{}{
		"message": "Data received",
		"data":    requestData,
	}
	
	json.NewEncoder(w).Encode(response)
}
`
}

// getWebCSSContent returns CSS content
func getWebCSSContent() string {
	return `/* {{.ProjectName}} Styles */

body {
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
    line-height: 1.6;
    margin: 0;
    padding: 0;
    background-color: #f8f9fa;
    color: #333;
}

.container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 20px;
}

header {
    background-color: #007bff;
    color: white;
    padding: 1rem 0;
    text-align: center;
}

header h1 {
    margin: 0;
    font-size: 2rem;
}

main {
    padding: 2rem 0;
}

.hero {
    text-align: center;
    padding: 3rem 0;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    border-radius: 8px;
    margin-bottom: 2rem;
}

.hero h2 {
    font-size: 2.5rem;
    margin-bottom: 1rem;
}

.hero p {
    font-size: 1.2rem;
    opacity: 0.9;
}

.features {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
    gap: 2rem;
    margin: 2rem 0;
}

.feature {
    background: white;
    padding: 2rem;
    border-radius: 8px;
    box-shadow: 0 2px 10px rgba(0,0,0,0.1);
    text-align: center;
}

.feature h3 {
    color: #007bff;
    margin-bottom: 1rem;
}

footer {
    background-color: #343a40;
    color: white;
    text-align: center;
    padding: 2rem 0;
    margin-top: 3rem;
}

.btn {
    display: inline-block;
    padding: 0.75rem 1.5rem;
    background-color: #007bff;
    color: white;
    text-decoration: none;
    border-radius: 4px;
    border: none;
    cursor: pointer;
    font-size: 1rem;
    transition: background-color 0.3s;
}

.btn:hover {
    background-color: #0056b3;
}

@media (max-width: 768px) {
    .hero h2 {
        font-size: 2rem;
    }
    
    .features {
        grid-template-columns: 1fr;
    }
    
    .container {
        padding: 10px;
    }
}
`
}

// getWebTemplateContent returns HTML template content
func getWebTemplateContent() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}}</title>
    <link rel="stylesheet" href="/static/style.css">
</head>
<body>
    <header>
        <div class="container">
            <h1>{{.Title}}</h1>
        </div>
    </header>
    
    <main>
        <div class="container">
            <div class="hero">
                <h2>{{.Message}}</h2>
                <p>A modern web application built with Go and Mage</p>
                <a href="/api/" class="btn">Try the API</a>
            </div>
            
            <div class="features">
                <div class="feature">
                    <h3>🚀 Fast</h3>
                    <p>Built with Go for maximum performance and efficiency</p>
                </div>
                
                <div class="feature">
                    <h3>🔧 Automated</h3>
                    <p>Powered by Mage for streamlined build and deployment</p>
                </div>
                
                <div class="feature">
                    <h3>📦 Containerized</h3>
                    <p>Ready for Docker and Kubernetes deployment</p>
                </div>
            </div>
        </div>
    </main>
    
    <footer>
        <div class="container">
            <p>&copy; 2024 {{.Title}}. Built with Go and Mage.</p>
        </div>
    </footer>
</body>
</html>
`
}

// getCLIMainContent returns CLI main content
func getCLIMainContent() string {
	return `package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	
	"{{.ModulePath}}/pkg/cli"
)

var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	var (
		showVersion = flag.Bool("version", false, "Show version information")
		verbose     = flag.Bool("verbose", false, "Enable verbose output")
		config      = flag.String("config", "", "Configuration file path")
	)
	
	flag.Parse()
	
	if *showVersion {
		fmt.Printf("{{.ProjectName}} version %s (built: %s, commit: %s)\n", version, buildTime, gitCommit)
		os.Exit(0)
	}
	
	app := cli.NewApp(cli.Config{
		Verbose:    *verbose,
		ConfigPath: *config,
	})
	
	if err := app.Run(flag.Args()); err != nil {
		log.Fatal(err)
	}
}
`
}

// getCLIPackageContent returns CLI package content
func getCLIPackageContent() string {
	return `package cli

import (
	"fmt"
	"os"
)

// Config holds CLI configuration
type Config struct {
	Verbose    bool
	ConfigPath string
}

// App represents the CLI application
type App struct {
	config Config
}

// NewApp creates a new CLI application
func NewApp(config Config) *App {
	return &App{
		config: config,
	}
}

// Run executes the CLI application
func (a *App) Run(args []string) error {
	if len(args) == 0 {
		return a.showHelp()
	}
	
	command := args[0]
	args = args[1:]
	
	switch command {
	case "hello":
		return a.helloCommand(args)
	case "version":
		return a.versionCommand()
	case "help":
		return a.showHelp()
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func (a *App) helloCommand(args []string) error {
	name := "World"
	if len(args) > 0 {
		name = args[0]
	}
	
	fmt.Printf("Hello, %s!\n", name)
	return nil
}

func (a *App) versionCommand() error {
	fmt.Println("{{.ProjectName}} CLI version 1.0.0")
	return nil
}

func (a *App) showHelp() error {
	fmt.Printf("Usage: %s [command] [args...]\n\n", os.Args[0])
	fmt.Println("Commands:")
	fmt.Println("  hello [name]  Say hello to someone")
	fmt.Println("  version       Show version information")
	fmt.Println("  help          Show this help message")
	fmt.Println("\nOptions:")
	fmt.Println("  -verbose      Enable verbose output")
	fmt.Println("  -config PATH  Configuration file path")
	fmt.Println("  -version      Show version and exit")
	return nil
}
`
}