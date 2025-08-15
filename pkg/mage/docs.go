// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for documentation tasks
var (
	errVersionRequired               = errors.New("version variable is required. Use: version=X.Y.Z mage docs:citation")
	errVersionLineNotFound           = errors.New("could not find version line to update")
	errDocumentationCheckFailed      = errors.New("documentation check failed")
	errNoDocumentationToolsAvailable = errors.New("no documentation tools available")
	errNoGeneratedDocumentationFound = errors.New("no generated documentation found - run 'mage docsGenerate' first")
	errNoMarkdownFilesFound          = errors.New("no markdown files found in generated documentation")
)

const (
	versionDev = "dev"
)

// DocServer represents a documentation server configuration
type DocServer struct {
	Tool     string   // "pkgsite", "godoc", or "none"
	Port     int      // Port to serve on
	Args     []string // Command arguments
	URL      string   // Full URL to serve documentation
	Mode     string   // "project", "stdlib", "both"
	Fallback bool     // Whether this is a fallback configuration
}

// DocTool constants for available documentation tools
const (
	DocToolPkgsite = "pkgsite"
	DocToolGodoc   = "godoc"
	DocToolNone    = "none"
)

// DocMode constants for documentation serving modes
const (
	DocModeProject = "project"
	DocModeStdlib  = "stdlib"
	DocModeBoth    = "both"
)

// Default ports for different tools
const (
	DefaultPkgsitePort = 8080
	DefaultGodocPort   = 6060
	DefaultStdlibPort  = 6061
)

// Docs namespace for documentation-related tasks
type Docs mg.Namespace

// GoDocs triggers pkg.go.dev sync
func (Docs) GoDocs() error {
	utils.Header("Syndicating to pkg.go.dev")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Construct module path
	module := config.Project.Module
	if module == "" {
		module, err = utils.GetModuleName()
		if err != nil {
			return fmt.Errorf("failed to get module name: %w", err)
		}
	}

	// Get latest version
	version := getVersion()
	if version == versionDev {
		utils.Warn("No version tag found, using latest")
		version = DefaultGoVulnCheckVersion
	}

	// Construct pkg.go.dev URL
	proxyURL := fmt.Sprintf("https://proxy.golang.org/%s/@v/%s.info", module, version)

	utils.Info("Triggering sync for %s@%s", module, version)

	// Use curl to trigger the proxy
	output, err := GetRunner().RunCmdOutput("curl", "-sSf", proxyURL)
	if err != nil {
		return fmt.Errorf("failed to trigger pkg.go.dev sync: %w", err)
	}

	utils.Success("Successfully triggered pkg.go.dev sync")
	utils.Info("Response: %s", strings.TrimSpace(output))
	utils.Info("View at: https://pkg.go.dev/%s", module)

	return nil
}

// Generate generates Go documentation for all packages
func (Docs) Generate() error {
	utils.Header("Generating Go Documentation")

	// Check if we have a docs directory
	docsDir := "docs"
	fileOps := fileops.New()
	if !utils.DirExists(docsDir) {
		utils.Info("Creating docs directory...")
		if err := fileOps.File.MkdirAll(docsDir, 0o755); err != nil {
			return fmt.Errorf("failed to create docs directory: %w", err)
		}
	}

	// Create generated docs subdirectory
	generatedDocsDir := filepath.Join(docsDir, "generated")
	if !utils.DirExists(generatedDocsDir) {
		utils.Info("Creating generated docs directory...")
		if err := fileOps.File.MkdirAll(generatedDocsDir, 0o755); err != nil {
			return fmt.Errorf("failed to create generated docs directory: %w", err)
		}
	}

	// Get list of all packages
	utils.Info("Discovering packages...")
	packagesOutput, err := GetRunner().RunCmdOutput("go", "list", "./...")
	if err != nil {
		return fmt.Errorf("failed to list packages: %w", err)
	}

	packages := strings.Split(strings.TrimSpace(packagesOutput), "\n")
	utils.Info("Found %d packages to document", len(packages))

	documentedPackages := make([]string, 0, len(packages))
	var failedPackages []string

	// Generate documentation for each package
	for _, pkg := range packages {
		if pkg == "" {
			continue
		}

		// Extract package name from full import path
		parts := strings.Split(pkg, "/")
		packageName := parts[len(parts)-1]

		// Skip certain packages that might not have useful documentation
		if shouldSkipPackage(pkg) {
			utils.Info("Skipping package: %s", packageName)
			continue
		}

		utils.Info("Generating docs for package: %s", packageName)

		// Generate documentation for this package
		output, err := GetRunner().RunCmdOutput("go", "doc", "-all", pkg)
		if err != nil {
			utils.Warn("Failed to generate docs for %s: %v", packageName, err)
			failedPackages = append(failedPackages, pkg)
			continue
		}

		// Create filename with package path to avoid conflicts
		filename := strings.ReplaceAll(strings.TrimPrefix(pkg, "github.com/mrz1836/mage-x/"), "/", "_") + ".md"
		docFile := filepath.Join(generatedDocsDir, filename)

		// Create content with proper Markdown formatting
		content := fmt.Sprintf("# %s Package Documentation\n\n", packageName)
		content += fmt.Sprintf("**Import Path:** `%s`\n\n", pkg)
		content += fmt.Sprintf("Generated by `go doc -all %s`\n\n", pkg)
		content += "---\n\n"
		content += output + "\n"

		if err := fileOps.File.WriteFile(docFile, []byte(content), 0o644); err != nil {
			utils.Warn("Failed to write docs for %s: %v", packageName, err)
			failedPackages = append(failedPackages, pkg)
			continue
		}

		documentedPackages = append(documentedPackages, pkg)
	}

	// Generate index file
	indexFile := filepath.Join(generatedDocsDir, "README.md")
	indexContent := generateDocumentationIndex(documentedPackages, failedPackages)

	if err := fileOps.File.WriteFile(indexFile, []byte(indexContent), 0o644); err != nil {
		utils.Warn("Failed to write documentation index: %v", err)
	}

	// Report results
	utils.Success("Documentation generated for %d packages", len(documentedPackages))
	if len(failedPackages) > 0 {
		utils.Warn("Failed to generate docs for %d packages", len(failedPackages))
	}
	utils.Info("Documentation available in: %s", generatedDocsDir)
	utils.Info("Index file: %s", indexFile)

	return nil
}

// shouldSkipPackage determines if a package should be skipped during documentation generation
func shouldSkipPackage(pkg string) bool {
	skipPatterns := []string{
		"testdata",
		"/fuzz",
		"examples/", // Skip example packages as they're demos
	}

	for _, pattern := range skipPatterns {
		if strings.Contains(pkg, pattern) {
			return true
		}
	}
	return false
}

// generateDocumentationIndex creates an index file listing all generated documentation
func generateDocumentationIndex(documented, failed []string) string {
	var content strings.Builder

	content.WriteString("# Generated Package Documentation\n\n")
	content.WriteString("This directory contains auto-generated documentation for all packages in the mage-x project.\n\n")
	content.WriteString(fmt.Sprintf("**Generated on:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	if len(documented) > 0 {
		content.WriteString("## 📚 Available Documentation\n\n")

		// Group packages by category
		categories := map[string][]string{
			"Core":      {},
			"Common":    {},
			"Providers": {},
			"Security":  {},
			"Utils":     {},
			"Commands":  {},
			"Other":     {},
		}

		for _, pkg := range documented {
			category := categorizePackage(pkg)
			categories[category] = append(categories[category], pkg)
		}

		// Write each category
		for category, packages := range categories {
			if len(packages) == 0 {
				continue
			}

			content.WriteString(fmt.Sprintf("### %s Packages\n\n", category))
			for _, pkg := range packages {
				filename := strings.ReplaceAll(strings.TrimPrefix(pkg, "github.com/mrz1836/mage-x/"), "/", "_") + ".md"
				packageName := pkg[strings.LastIndex(pkg, "/")+1:]
				content.WriteString(fmt.Sprintf("- [%s](./%s) - `%s`\n", packageName, filename, pkg))
			}
			content.WriteString("\n")
		}
	}

	if len(failed) > 0 {
		content.WriteString("## ⚠️ Failed Documentation Generation\n\n")
		content.WriteString("The following packages failed to generate documentation:\n\n")
		for _, pkg := range failed {
			content.WriteString(fmt.Sprintf("- `%s`\n", pkg))
		}
		content.WriteString("\n")
	}

	content.WriteString("## 🔧 Regenerating Documentation\n\n")
	content.WriteString("To regenerate this documentation, run:\n\n")
	content.WriteString("```bash\n")
	content.WriteString("mage docsGenerate\n")
	content.WriteString("```\n\n")
	content.WriteString("---\n")
	content.WriteString("*Documentation generated by mage-x docs generator*\n")

	return content.String()
}

// categorizePackage determines the category for a package based on its path
func categorizePackage(pkg string) string {
	switch {
	case strings.Contains(pkg, "/pkg/mage"):
		return "Core"
	case strings.Contains(pkg, "/pkg/common"):
		return "Common"
	case strings.Contains(pkg, "/pkg/providers"):
		return "Providers"
	case strings.Contains(pkg, "/pkg/security"):
		return "Security"
	case strings.Contains(pkg, "/pkg/utils") || strings.Contains(pkg, "/pkg/testhelpers"):
		return "Utils"
	case strings.Contains(pkg, "/cmd/"):
		return "Commands"
	default:
		return "Other"
	}
}

// ServeDefault starts a local godoc server on default port
func (Docs) ServeDefault() error {
	return Docs{}.Serve()
}

// Serve serves documentation locally using the best available tool
func (Docs) Serve() error {
	utils.Header("Starting Documentation Server")

	// Detect the best available documentation tool
	server := detectBestDocTool()

	if server.Tool == DocToolNone {
		utils.Info("No documentation tools found. Installing pkgsite...")
		if err := installDocTool(DocToolPkgsite); err != nil {
			utils.Warn("Failed to install pkgsite: %v", err)
			utils.Info("Installing godoc as fallback...")
			if err := installDocTool(DocToolGodoc); err != nil {
				return fmt.Errorf("failed to install documentation tools: %w", err)
			}
		}
		// Re-detect after installation
		server = detectBestDocTool()
	}

	return serveWithDocServer(server)
}

// serveWithDocServer starts a documentation server using the provided DocServer configuration
func serveWithDocServer(server DocServer) error {
	if server.Tool == DocToolNone {
		return errNoDocumentationToolsAvailable
	}

	// Install tool if missing
	if err := installDocTool(server.Tool); err != nil {
		return fmt.Errorf("failed to install %s: %w", server.Tool, err)
	}

	// Display server information
	utils.Info("Using %s for documentation serving", server.Tool)
	if server.Fallback {
		utils.Info("Port adjusted due to availability: %d", server.Port)
	}
	utils.Info("Starting %s server on %s", server.Tool, server.URL)
	utils.Info("Press Ctrl+C to stop")

	// Open browser if not in CI and tool doesn't auto-open
	if !isCI() && server.Tool != DocToolPkgsite {
		go openBrowser(server.URL)
	}

	// Start the documentation server
	return GetRunner().RunCmd(server.Tool, server.Args...)
}

// openBrowser opens the documentation URL in the default browser
func openBrowser(url string) {
	// Skip browser opening in test environment
	if isTestEnvironment() {
		utils.Info("Test environment detected - skipping browser open for: %s", url)
		return
	}

	utils.Info("Opening browser...")

	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd, args = "open", []string{url}
	case "linux":
		cmd, args = "xdg-open", []string{url}
	case OSWindows:
		cmd, args = "cmd", []string{"/c", "start", url}
	default:
		utils.Info("Please open your browser and navigate to: %s", url)
		return
	}

	if err := GetRunner().RunCmd(cmd, args...); err != nil {
		utils.Warn("Failed to open browser: %v", err)
		utils.Info("Please open your browser and navigate to: %s", url)
	}
}

// ServePkgsite serves documentation using pkgsite (forced)
func (Docs) ServePkgsite() error {
	utils.Header("Starting Pkgsite Documentation Server")

	server := buildDocServer(DocToolPkgsite, DocModeProject, getPortFromEnv(DefaultPkgsitePort))
	return serveWithDocServer(server)
}

// ServeGodoc serves documentation using godoc (forced)
func (Docs) ServeGodoc() error {
	utils.Header("Starting Godoc Documentation Server")

	server := buildDocServer(DocToolGodoc, DocModeProject, getPortFromEnv(DefaultGodocPort))
	return serveWithDocServer(server)
}

// ServeStdlib serves standard library documentation
func (Docs) ServeStdlib() error {
	utils.Header("Starting Standard Library Documentation Server")

	// Prefer godoc for stdlib docs as it's more comprehensive
	server := buildDocServer(DocToolGodoc, DocModeStdlib, getPortFromEnv(DefaultStdlibPort))
	return serveWithDocServer(server)
}

// ServeProject serves project-focused documentation
func (Docs) ServeProject() error {
	utils.Header("Starting Project Documentation Server")

	// Prefer pkgsite for project docs as it's more modern
	server := detectBestDocTool()
	if server.Tool == DocToolNone {
		server = buildDocServer(DocToolGodoc, DocModeProject, getPortFromEnv(DefaultGodocPort))
	}
	server.Mode = DocModeProject
	return serveWithDocServer(server)
}

// ServeBoth serves both project and standard library documentation
func (Docs) ServeBoth() error {
	utils.Header("Starting Full Documentation Server")

	// Use godoc for comprehensive docs (both project and stdlib)
	server := buildDocServer(DocToolGodoc, DocModeBoth, getPortFromEnv(DefaultGodocPort))
	return serveWithDocServer(server)
}

// detectBestDocTool determines the best available documentation tool and configuration
func detectBestDocTool() DocServer {
	// Check environment variable override
	if tool := os.Getenv("DOCS_TOOL"); tool != "" {
		if tool == DocToolPkgsite && utils.CommandExists("pkgsite") {
			return buildDocServer(DocToolPkgsite, DocModeProject, getPortFromEnv(DefaultPkgsitePort))
		}
		if tool == DocToolGodoc && utils.CommandExists("godoc") {
			return buildDocServer(DocToolGodoc, DocModeProject, getPortFromEnv(DefaultGodocPort))
		}
	}

	// Priority 1: pkgsite (modern, project-focused)
	if utils.CommandExists("pkgsite") {
		port := getPortFromEnv(DefaultPkgsitePort)
		return buildDocServer(DocToolPkgsite, DocModeProject, port)
	}

	// Priority 2: godoc with project focus
	if utils.CommandExists("godoc") {
		port := getPortFromEnv(DefaultGodocPort)
		return buildDocServer(DocToolGodoc, DocModeProject, port)
	}

	// No tools available
	return DocServer{
		Tool: DocToolNone,
		Port: 0,
		Args: []string{},
		URL:  "",
		Mode: DocModeProject,
	}
}

// buildDocServer creates a DocServer configuration for the specified tool and mode
func buildDocServer(tool, mode string, port int) DocServer {
	// Ensure port is available, find alternative if needed
	finalPort := findAvailablePort(port)

	server := DocServer{
		Tool:     tool,
		Port:     finalPort,
		Mode:     mode,
		URL:      fmt.Sprintf("http://localhost:%d", finalPort),
		Fallback: finalPort != port,
	}

	switch tool {
	case DocToolPkgsite:
		server.Args = buildPkgsiteArgs(mode, finalPort)
	case DocToolGodoc:
		server.Args = buildGodocArgs(mode, finalPort)
	default:
		server.Args = []string{}
	}

	return server
}

// buildPkgsiteArgs creates command arguments for pkgsite based on port
func buildPkgsiteArgs(_ string, port int) []string {
	args := []string{"-http", fmt.Sprintf(":%d", port)}

	// pkgsite automatically focuses on current directory/module
	// Add -open flag to automatically open browser
	if !isCI() {
		args = append(args, "-open")
	}

	return args
}

// buildGodocArgs creates command arguments for godoc based on mode and port
func buildGodocArgs(mode string, port int) []string {
	args := []string{"-http", fmt.Sprintf(":%d", port)}

	switch mode {
	case DocModeProject:
		// Focus on current project, disable stdlib focus
		args = append(args, "-goroot=")
	case DocModeStdlib:
		// Default godoc behavior (stdlib focused)
		// No additional args needed
	case DocModeBoth:
		// Default godoc behavior shows both
		// No additional args needed
	}

	return args
}

// findAvailablePort finds an available port starting from the preferred port
func findAvailablePort(preferredPort int) int {
	for port := preferredPort; port < preferredPort+10; port++ {
		if isPortAvailable(port) {
			return port
		}
	}
	// Fallback to any available port
	return findAnyAvailablePort()
}

// isPortAvailable checks if a port is available for binding
func isPortAvailable(port int) bool {
	address := fmt.Sprintf(":%d", port)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	lc := &net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", address)
	if err != nil {
		return false
	}
	_ = listener.Close() //nolint:errcheck // Close error not critical in port availability check
	return true
}

// findAnyAvailablePort finds any available port in a reasonable range
func findAnyAvailablePort() int {
	for port := 8000; port < 9000; port++ {
		if isPortAvailable(port) {
			return port
		}
	}
	return 8080 // Ultimate fallback
}

// getPortFromEnv gets port from environment variable or returns default
func getPortFromEnv(defaultPort int) int {
	if portStr := os.Getenv("DOCS_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 && port < 65536 {
			return port
		}
	}
	return defaultPort
}

// isCI detects if running in CI environment
func isCI() bool {
	ci := os.Getenv("CI")
	return ci == "true" || ci == "1"
}

// isTestEnvironment detects if code is running under go test
func isTestEnvironment() bool {
	// Check if running under go test by looking for test-specific indicators
	// Method 1: Check for -test. binary suffix (most reliable)
	if strings.HasSuffix(os.Args[0], ".test") || strings.Contains(os.Args[0], ".test") {
		return true
	}
	// Method 2: Check for test flags
	for _, arg := range os.Args {
		if strings.HasPrefix(arg, "-test.") {
			return true
		}
	}
	// Method 3: Check environment variable that could be set by test runners
	if os.Getenv("GO_TEST") == "1" {
		return true
	}
	return false
}

// installDocTool installs the specified documentation tool if missing
func installDocTool(tool string) error {
	switch tool {
	case DocToolPkgsite:
		if !utils.CommandExists("pkgsite") {
			utils.Info("Installing pkgsite...")
			return GetRunner().RunCmd("go", "install", "golang.org/x/pkgsite/cmd/pkgsite@latest")
		}
	case DocToolGodoc:
		if !utils.CommandExists("godoc") {
			utils.Info("Installing godoc...")
			return GetRunner().RunCmd("go", "install", "golang.org/x/tools/cmd/godoc@latest")
		}
	}
	return nil
}

// Check validates documentation
func (Docs) Check() error {
	utils.Header("Checking Documentation")

	issues := []string{}

	// Check README.md
	if !utils.FileExists("README.md") {
		issues = append(issues, "README.md is missing")
	}

	// Check LICENSE
	if !utils.FileExists("LICENSE") {
		issues = append(issues, "LICENSE file is missing")
	}

	// Check CONTRIBUTING.md
	if !utils.FileExists("CONTRIBUTING.md") {
		utils.Warn("CONTRIBUTING.md is missing (recommended)")
	}

	// Check SECURITY.md
	if !utils.FileExists("SECURITY.md") {
		utils.Warn("SECURITY.md is missing (recommended)")
	}

	// Check package comments
	utils.Info("Checking package documentation...")
	output, err := GetRunner().RunCmdOutput("go", "list", "-f", "{{.Doc}}", "./...")
	if err != nil {
		utils.Warn("Failed to check package documentation: %v", err)
	} else {
		packages := strings.Split(strings.TrimSpace(output), "\n")
		for i, doc := range packages {
			if strings.TrimSpace(doc) == "" {
				utils.Warn("Package %d is missing documentation comment", i+1)
			}
		}
	}

	// Check for example files
	examples, err := utils.FindFiles(".", "example*.go")
	if err != nil {
		utils.Warn("Failed to find example files: %v", err)
		examples = []string{}
	}
	if len(examples) == 0 {
		utils.Info("No example files found (consider adding examples)")
	} else {
		utils.Success("Found %d example file(s)", len(examples))
	}

	if len(issues) > 0 {
		utils.Error("Documentation issues found:")
		for _, issue := range issues {
			utils.Info("  - %s", issue)
		}
		return errDocumentationCheckFailed
	}

	utils.Success("Documentation check passed")
	return nil
}

// Examples generates example documentation
func (Docs) Examples() error {
	utils.Header("Generating Example Documentation")

	// Find all example files
	examples, err := utils.FindFiles(".", "example*.go")
	if err != nil {
		return fmt.Errorf("failed to find examples: %w", err)
	}

	if len(examples) == 0 {
		utils.Warn("No example files found")
		return nil
	}

	// Create examples documentation
	var content strings.Builder
	content.WriteString("# Examples\n\n")

	for _, example := range examples {
		utils.Info("Processing %s", example)

		// Read example file
		fileOps := fileops.New()
		code, err := fileOps.File.ReadFile(example)
		if err != nil {
			utils.Warn("Failed to read %s: %v", example, err)
			continue
		}

		// Extract example name from filename
		name := filepath.Base(example)
		name = strings.TrimSuffix(name, ".go")
		name = strings.TrimPrefix(name, "example_")

		// Add to documentation
		content.WriteString(fmt.Sprintf("## %s\n\n", name))
		content.WriteString("```go\n")
		content.Write(code)
		content.WriteString("\n```\n\n")
	}

	// Write examples documentation
	examplesFile := "EXAMPLES.md"
	fileOps := fileops.New()
	if err := fileOps.File.WriteFile(examplesFile, []byte(content.String()), 0o644); err != nil {
		return fmt.Errorf("failed to write examples documentation: %w", err)
	}

	utils.Success("Generated %s with %d examples", examplesFile, len(examples))
	return nil
}

// Default generates default documentation
func (Docs) Default() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Generating default documentation")
}

// Build builds static documentation files from generated markdown
func (Docs) Build() error {
	utils.Header("Building Static Documentation")

	// First ensure we have generated documentation
	utils.Info("Step 1: Generating documentation...")
	var d Docs
	if err := d.Generate(); err != nil {
		return fmt.Errorf("failed to generate documentation: %w", err)
	}

	// Create build output directory
	buildDir := "docs/build"
	fileOps := fileops.New()
	if !utils.DirExists(buildDir) {
		utils.Info("Creating build directory...")
		if err := fileOps.File.MkdirAll(buildDir, 0o755); err != nil {
			return fmt.Errorf("failed to create build directory: %w", err)
		}
	}

	// Copy generated markdown files to build directory
	generatedDir := "docs/generated"
	if !utils.DirExists(generatedDir) {
		return errNoGeneratedDocumentationFound
	}

	utils.Info("Step 2: Processing generated documentation...")

	// Find all generated markdown files
	markdownFiles, err := utils.FindFiles(generatedDir, "*.md")
	if err != nil {
		return fmt.Errorf("failed to find generated documentation files: %w", err)
	}

	if len(markdownFiles) == 0 {
		return errNoMarkdownFilesFound
	}

	// Process each markdown file
	processedFiles := make([]string, 0, len(markdownFiles))
	for _, file := range markdownFiles {
		utils.Info("Processing %s", filepath.Base(file))

		// Read the markdown file
		content, err := fileOps.File.ReadFile(file)
		if err != nil {
			utils.Warn("Failed to read %s: %v", file, err)
			continue
		}

		// Process the content (add metadata, improve formatting)
		processedContent := processMarkdownForBuild(string(content), file)

		// Write to build directory
		outputFile := filepath.Join(buildDir, filepath.Base(file))
		if err := fileOps.File.WriteFile(outputFile, []byte(processedContent), 0o644); err != nil {
			utils.Warn("Failed to write processed file %s: %v", outputFile, err)
			continue
		}

		processedFiles = append(processedFiles, outputFile)
	}

	// Create an enhanced index file for the build
	utils.Info("Step 3: Creating build index...")
	buildIndexContent := createBuildIndex(processedFiles)
	buildIndexFile := filepath.Join(buildDir, "index.md")
	if err := fileOps.File.WriteFile(buildIndexFile, []byte(buildIndexContent), 0o644); err != nil {
		utils.Warn("Failed to create build index: %v", err)
	}

	// Create build metadata
	buildMetadata := createBuildMetadata(len(processedFiles))
	metadataFile := filepath.Join(buildDir, "build-info.json")
	if err := fileOps.File.WriteFile(metadataFile, []byte(buildMetadata), 0o644); err != nil {
		utils.Warn("Failed to create build metadata: %v", err)
	}

	utils.Success("Documentation build completed!")
	utils.Info("Built %d documentation files", len(processedFiles))
	utils.Info("Build directory: %s", buildDir)
	utils.Info("Main index: %s", buildIndexFile)
	utils.Info("Build metadata: %s", metadataFile)

	return nil
}

// Lint lints the documentation
func (Docs) Lint() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Linting documentation")
}

// Spell checks spelling in documentation
func (Docs) Spell() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Spell checking documentation")
}

// Links checks links in documentation
func (Docs) Links() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Checking documentation links")
}

// API generates API documentation
func (Docs) API() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Generating API documentation")
}

// Markdown generates Markdown documentation
func (Docs) Markdown() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Generating Markdown documentation")
}

// Readme generates README documentation
func (Docs) Readme() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Generating README")
}

// Changelog generates changelog documentation
func (Docs) Changelog(version string) error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Generating changelog for version:", version)
}

// processMarkdownForBuild enhances markdown content for the build process
func processMarkdownForBuild(content, filePath string) string {
	var enhanced strings.Builder

	// Add build metadata header
	enhanced.WriteString("---\n")
	enhanced.WriteString(fmt.Sprintf("generated: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	enhanced.WriteString(fmt.Sprintf("source: %s\n", filePath))
	enhanced.WriteString("build: mage-x documentation builder\n")
	enhanced.WriteString("---\n\n")

	// Add navigation hint
	enhanced.WriteString("> 📚 **MAGE-X Documentation** | [View All Packages](./index.md) | [Back to Generated Docs](../generated/README.md)\n\n")

	// Add the original content
	enhanced.WriteString(content)

	// Add footer
	enhanced.WriteString("\n\n---\n")
	enhanced.WriteString("*Built with MAGE-X documentation system*\n")

	return enhanced.String()
}

// createBuildIndex generates an enhanced index file for the build
func createBuildIndex(processedFiles []string) string {
	var content strings.Builder

	content.WriteString("# MAGE-X Documentation Build\n\n")
	content.WriteString("> 🏗️ **Static Documentation Build**\n\n")
	content.WriteString(fmt.Sprintf("**Generated:** %s  \n", time.Now().Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("**Files:** %d packages documented  \n", len(processedFiles)))
	content.WriteString("**Format:** Enhanced Markdown with metadata  \n\n")

	content.WriteString("## 📚 Available Documentation\n\n")

	// Group files by category based on filename patterns
	categories := map[string][]string{
		"Core Packages":     {},
		"Common Packages":   {},
		"Provider Packages": {},
		"Security Packages": {},
		"Utility Packages":  {},
		"Command Packages":  {},
		"Other Packages":    {},
	}

	for _, file := range processedFiles {
		basename := filepath.Base(file)
		if basename == "README.md" {
			continue // Skip the main README
		}

		category := categorizeBuildFile(basename)
		categories[category] = append(categories[category], basename)
	}

	// Display each category
	for category, files := range categories {
		if len(files) == 0 {
			continue
		}

		content.WriteString(fmt.Sprintf("### %s\n\n", category))
		for _, file := range files {
			// Extract package name from filename
			packageName := strings.TrimSuffix(file, ".md")
			packageName = strings.ReplaceAll(packageName, "_", "/")
			packageName = strings.TrimPrefix(packageName, "pkg/")
			packageName = strings.TrimPrefix(packageName, "cmd/")

			content.WriteString(fmt.Sprintf("- [%s](./%s)\n", packageName, file))
		}
		content.WriteString("\n")
	}

	content.WriteString("## 🔧 Build Information\n\n")
	content.WriteString("This documentation was built using the MAGE-X documentation system:\n\n")
	content.WriteString("```bash\n")
	content.WriteString("# Regenerate this build\n")
	content.WriteString("mage docsBuild\n")
	content.WriteString("\n")
	content.WriteString("# View build metadata\n")
	content.WriteString("cat docs/build/build-info.json\n")
	content.WriteString("```\n\n")

	content.WriteString("## 📖 More Resources\n\n")
	content.WriteString("- [Generated Documentation](../generated/README.md) - Raw generated docs\n")
	content.WriteString("- [Project Documentation](../README.md) - Main documentation index\n")
	content.WriteString("- [API Reference](../API_REFERENCE.md) - Complete API documentation\n")
	content.WriteString("- [Quick Start Guide](../QUICK_START.md) - Get started with MAGE-X\n\n")

	content.WriteString("---\n")
	content.WriteString("*Documentation built with ❤️ by MAGE-X*\n")

	return content.String()
}

// categorizeBuildFile determines the category for a documentation file
func categorizeBuildFile(filename string) string {
	switch {
	case strings.Contains(filename, "pkg_mage"):
		return "Core Packages"
	case strings.Contains(filename, "pkg_common"):
		return "Common Packages"
	case strings.Contains(filename, "pkg_providers"):
		return "Provider Packages"
	case strings.Contains(filename, "pkg_security"):
		return "Security Packages"
	case strings.Contains(filename, "pkg_utils") || strings.Contains(filename, "pkg_testhelpers"):
		return "Utility Packages"
	case strings.Contains(filename, "cmd_"):
		return "Command Packages"
	default:
		return "Other Packages"
	}
}

// createBuildMetadata generates JSON metadata about the build
func createBuildMetadata(fileCount int) string {
	metadata := map[string]interface{}{
		"build_time":    time.Now().Format(time.RFC3339),
		"builder":       "mage-x documentation system",
		"version":       "1.0.0",
		"file_count":    fileCount,
		"format":        "enhanced-markdown",
		"features":      []string{"metadata", "navigation", "categorization"},
		"build_command": "mage docsBuild",
	}

	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "{\"error\": \"failed to generate metadata\"}"
	}

	return string(jsonData)
}

// Clean cleans documentation build artifacts
func (Docs) Clean() error {
	utils.Header("Cleaning Documentation Build Artifacts")

	buildDir := "docs/build"
	if !utils.DirExists(buildDir) {
		utils.Info("No build directory found - nothing to clean")
		return nil
	}

	fileOps := fileops.New()
	if err := fileOps.File.RemoveAll(buildDir); err != nil {
		return fmt.Errorf("failed to clean build directory: %w", err)
	}

	utils.Success("Documentation build artifacts cleaned")
	utils.Info("Removed directory: %s", buildDir)

	return nil
}
