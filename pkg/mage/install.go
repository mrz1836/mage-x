// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magefile/mage/mg"

	"github.com/mrz1836/mage-x/pkg/common/env"
	mageErrors "github.com/mrz1836/mage-x/pkg/common/errors"
	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for install operations
var (
	errInstallationVerificationFailed = errors.New("installation verification failed")
	errSystemWideNotSupportedWindows  = errors.New("system-wide installation not supported on Windows")
)

const (
	defaultBinaryName = "app"
	windowsExeExt     = ".exe"
)

// Install namespace for installation tasks
type Install mg.Namespace

// Local installs the application locally (alias for Default for compatibility)
func (Install) Local() error {
	return Install{}.Default()
}

// Uninstall removes the installed application
func (Install) Uninstall() error {
	utils.Header("Uninstalling Application")

	ops := GetOSOperations()
	goOps := GetGoOperations()

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Get binary name
	binaryName := config.Project.Binary
	if binaryName == "" {
		// Try to get from module name
		if module, moduleErr := goOps.GetModuleName(); moduleErr == nil {
			parts := strings.Split(module, "/")
			binaryName = parts[len(parts)-1]
		} else {
			binaryName = defaultBinaryName
		}
	}

	// Get GOPATH
	gopath := ops.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(ops.Getenv("HOME"), "go")
	}

	// Determine install path
	installPath := filepath.Join(gopath, "bin", binaryName)
	if runtime.GOOS == OSWindows && !strings.HasSuffix(installPath, ".exe") {
		installPath += windowsExeExt
	}

	// Remove the binary
	if err := ops.Remove(installPath); err != nil {
		if os.IsNotExist(err) {
			utils.Warn("Binary not found at %s", installPath)
			return nil
		}
		return mageErrors.WrapError(err, "failed to remove binary")
	}

	utils.Success("Uninstalled %s from %s", binaryName, installPath)
	return nil
}

// Default installs the application binary
func (Install) Default() error {
	utils.Header("Installing Application")

	ops := GetOSOperations()
	goOps := GetGoOperations()

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Get binary name
	binaryName := config.Project.Binary
	if binaryName == "" {
		// Try to get from module name
		if module, moduleErr := goOps.GetModuleName(); moduleErr == nil {
			parts := strings.Split(module, "/")
			binaryName = parts[len(parts)-1]
		} else {
			binaryName = defaultBinaryName
		}
	}

	// Get GOPATH
	gopath := ops.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(ops.Getenv("HOME"), "go")
	}

	// Determine install path
	installPath := filepath.Join(gopath, "bin", binaryName)
	if runtime.GOOS == OSWindows && !strings.HasSuffix(installPath, ".exe") {
		installPath += windowsExeExt
	}

	utils.Info("Installing to: %s", installPath)

	// Determine package path using Build logic
	buildOps := GetBuildOperations()
	packagePath, err := buildOps.DeterminePackagePath(config, installPath, true)
	if err != nil {
		return mageErrors.WrapError(err, "failed to determine package path")
	}

	// Build with installation flags
	args := []string{"build", "-o", installPath}

	// Add build flags using the shared helper (handles template expansion)
	args = append(args, buildFlags(config)...)

	// Add main package
	args = append(args, packagePath)

	// Build and install
	if err := GetRunner().RunCmd("go", args...); err != nil {
		return mageErrors.WrapError(err, "installation failed")
	}

	// Verify installation
	if ops.FileExists(installPath) {
		utils.Success("Successfully installed %s", binaryName)
		utils.Info("Binary location: %s", installPath)

		// Create symlink aliases from configuration
		for _, alias := range config.Project.Aliases {
			createSymlinkAlias(gopath, installPath, alias)
		}

		// Check if GOPATH/bin is in PATH
		if !isInPath(filepath.Join(gopath, "bin")) {
			utils.Warn("Note: %s/bin is not in your PATH", gopath)
			utils.Info("Add it with: export PATH=$PATH:%s/bin", gopath)
		}
	} else {
		return errInstallationVerificationFailed
	}

	return nil
}

// Go installs using go install with specific version
func (Install) Go() error {
	utils.Header("Installing with go install")

	ops := GetOSOperations()
	goOps := GetGoOperations()

	// Get module info
	module, err := goOps.GetModuleName()
	if err != nil {
		return mageErrors.WrapError(err, "failed to get module name")
	}

	// Get version
	installVersion := env.GetString("VERSION", goOps.GetVersion())
	if installVersion == versionDev || installVersion == "" {
		installVersion = goOps.GetGoVulnCheckVersion()
		if installVersion == "" {
			utils.Warn("GoVulnCheck version not available, using @latest")
			installVersion = VersionLatest
		}
	}

	// Build install command
	var pkg string
	if installVersion != VersionLatest && installVersion != "" {
		pkg = fmt.Sprintf("%s@%s", module, installVersion)
	} else {
		pkg = fmt.Sprintf("%s@latest", module)
	}

	utils.Info("Installing %s", pkg)

	args := []string{"install"}

	// Add build tags if specified
	if tags := ops.Getenv("MAGE_X_BUILD_TAGS"); tags != "" {
		args = append(args, "-tags", tags)
	}

	args = append(args, pkg)

	// Run go install
	if err := GetRunner().RunCmd("go", args...); err != nil {
		return mageErrors.WrapError(err, "go install failed")
	}

	utils.Success("Successfully installed via go install")
	return nil
}

// Stdlib installs the Go standard library for the host platform
func (Install) Stdlib() error {
	utils.Header("Installing Go Standard Library")

	utils.Info("Installing standard library packages...")

	if err := GetRunner().RunCmd("go", "install", "std"); err != nil {
		return mageErrors.WrapError(err, "failed to install standard library")
	}

	utils.Success("Go standard library installed")
	return nil
}

// Tools installs development tools
func (Install) Tools() error {
	utils.Header("Installing Development Tools")

	tools := []struct {
		name string
		pkg  string
		desc string
	}{
		{"golangci-lint", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest", "Linter aggregator"},
		{"gofumpt", "mvdan.cc/gofumpt@latest", "Stricter gofmt"},
		{"yamlfmt", "github.com/google/yamlfmt/cmd/yamlfmt@" + GetDefaultYamlfmtVersion(), "YAML formatter"},
		{"godoc", "golang.org/x/tools/cmd/godoc@latest", "Documentation server"},
		{"goimports", "golang.org/x/tools/cmd/goimports@latest", "Import organizer"},
		{"govulncheck", "golang.org/x/vuln/cmd/govulncheck@latest", "Vulnerability scanner"},
		{"mockgen", "github.com/golang/mock/mockgen@latest", "Mock generator"},
		{"staticcheck", "honnef.co/go/tools/cmd/staticcheck@latest", "Static analyzer"},
		{"benchstat", "golang.org/x/perf/cmd/benchstat@latest", "Benchmark comparison"},
	}

	installed := 0

	for _, tool := range tools {
		if utils.CommandExists(tool.name) {
			utils.Info("✓ %s already installed", tool.name)
			continue
		}

		utils.Info("Installing %s - %s", tool.name, tool.desc)
		if err := GetRunner().RunCmd("go", "install", tool.pkg); err != nil {
			utils.Warn("Failed to install %s: %v", tool.name, err)
		} else {
			installed++
		}
	}

	if installed > 0 {
		utils.Success("Installed %d new tools", installed)
	} else {
		utils.Success("All tools already installed")
	}

	return nil
}

// SystemWide installs the binary system-wide (requires sudo)
func (Install) SystemWide() error {
	utils.Header("Installing System-Wide")

	if runtime.GOOS == OSWindows {
		return errSystemWideNotSupportedWindows
	}

	ops := GetOSOperations()
	goOps := GetGoOperations()

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Get binary name
	binaryName := config.Project.Binary
	if binaryName == "" {
		if module, moduleErr := goOps.GetModuleName(); moduleErr == nil {
			parts := strings.Split(module, "/")
			binaryName = parts[len(parts)-1]
		} else {
			binaryName = defaultBinaryName
		}
	}

	// Build binary first
	tempBinary := filepath.Join(ops.TempDir(), binaryName)

	utils.Info("Building binary...")

	// Determine package path using Build logic
	buildOps := GetBuildOperations()
	packagePath, err := buildOps.DeterminePackagePath(config, tempBinary, true)
	if err != nil {
		return mageErrors.WrapError(err, "failed to determine package path")
	}

	args := []string{"build", "-o", tempBinary}

	// Add build flags using the shared helper (handles template expansion)
	args = append(args, buildFlags(config)...)

	args = append(args, packagePath)

	if err := GetRunner().RunCmd("go", args...); err != nil {
		return mageErrors.WrapError(err, "build failed")
	}

	// Install to /usr/local/bin
	usrLocalBin := string(os.PathSeparator) + filepath.Join("usr", "local", "bin")
	installPath := filepath.Join(usrLocalBin, binaryName)

	utils.Info("Installing to %s (requires sudo)...", installPath)

	// Use sudo to copy
	if err := GetRunner().RunCmd("sudo", "cp", tempBinary, installPath); err != nil {
		return mageErrors.WrapError(err, "installation failed")
	}

	// Make executable
	if err := GetRunner().RunCmd("sudo", "chmod", "+x", installPath); err != nil {
		return mageErrors.WrapError(err, "failed to set permissions")
	}

	// Clean up temp file
	if err := ops.Remove(tempBinary); err != nil {
		// Log but don't fail - this is cleanup
		utils.Debug("Failed to remove temporary file %s: %v", tempBinary, err)
	}

	utils.Success("Successfully installed %s system-wide", binaryName)
	utils.Info("Binary location: %s", installPath)

	return nil
}

// Helper functions

// isInPath checks if a directory is in PATH
func isInPath(dir string) bool {
	ops := GetOSOperations()
	path := ops.Getenv("PATH")
	paths := strings.Split(path, string(os.PathListSeparator))

	for _, p := range paths {
		if p == dir {
			return true
		}
	}

	return false
}

// createSymlinkAlias creates a symlink alias for the installed binary
func createSymlinkAlias(gopath, installPath, aliasName string) {
	ops := GetOSOperations()

	// Determine alias path
	aliasPath := filepath.Join(gopath, "bin", aliasName)
	if runtime.GOOS == OSWindows && !strings.HasSuffix(aliasPath, ".exe") {
		aliasPath += windowsExeExt
	}

	// Check if alias already exists
	if ops.FileExists(aliasPath) {
		if checkExistingAlias(aliasPath, installPath, aliasName) {
			return
		}
		utils.Warn("File '%s' already exists, skipping alias creation", aliasPath)
		return
	}

	// Create symlink (or batch file on Windows)
	if runtime.GOOS == OSWindows {
		// On Windows, create a batch file wrapper instead of symlink
		createWindowsBatchWrapper(aliasPath, installPath, aliasName)
	} else {
		// Unix/Mac: create symlink
		if err := ops.Symlink(installPath, aliasPath); err != nil {
			utils.Warn("Failed to create alias '%s': %v", aliasName, err)
		} else {
			binaryName := filepath.Base(installPath)
			if runtime.GOOS == OSWindows {
				binaryName = strings.TrimSuffix(binaryName, ".exe")
			}
			utils.Success("Created alias '%s' → '%s'", aliasName, binaryName)
		}
	}
}

// checkExistingAlias checks if an existing alias already points to our binary
func checkExistingAlias(aliasPath, installPath, aliasName string) bool {
	ops := GetOSOperations()

	// If it's already a symlink to our binary, that's fine
	link, err := ops.Readlink(aliasPath)
	if err != nil {
		return false
	}

	absLink, err := filepath.Abs(link)
	if err != nil {
		return false
	}

	absInstall, err := filepath.Abs(installPath)
	if err != nil {
		return false
	}

	if absLink == absInstall {
		utils.Info("Alias '%s' already points to the binary", aliasName)
		return true
	}

	return false
}

// createWindowsBatchWrapper creates a batch file wrapper on Windows
func createWindowsBatchWrapper(aliasPath, installPath, aliasName string) {
	ops := GetOSOperations()

	// Remove .exe extension for batch file
	if strings.HasSuffix(aliasPath, ".exe") {
		aliasPath = strings.TrimSuffix(aliasPath, ".exe") + ".bat"
	}

	// Create batch file content that forwards all arguments
	batchContent := fmt.Sprintf("@echo off\n\"%s\" %%*\n", installPath)

	if err := ops.WriteFile(aliasPath, []byte(batchContent), fileops.PermFileSensitive); err != nil {
		utils.Warn("Failed to create alias '%s': %v", aliasName, err)
	} else {
		binaryName := filepath.Base(installPath)
		if runtime.GOOS == OSWindows {
			binaryName = strings.TrimSuffix(binaryName, ".exe")
		}
		utils.Success("Created alias '%s' → '%s'", aliasName, binaryName)
	}
}

// getVersion gets the current version

// Additional methods for Install namespace required by tests

// Binary installs binary
func (Install) Binary() error {
	utils.Header("Installing Binary")

	installer := Install{}

	// Use the default installation method
	if err := installer.Default(); err != nil {
		return mageErrors.WrapError(err, "binary installation failed")
	}

	return nil
}

// Deps installs dependencies
func (Install) Deps() error {
	runner := GetRunner()
	return runner.RunCmd("go", "mod", "download")
}

// Mage installs mage
func (Install) Mage() error {
	runner := GetRunner()
	return runner.RunCmd("go", "install", "github.com/magefile/mage@latest")
}

// Docker installs Docker requirements
func (Install) Docker() error {
	utils.Header("Installing Docker Requirements")

	// Check if Docker is already installed
	if utils.CommandExists("docker") {
		utils.Info("Docker already installed")
		return nil
	}

	// Provide installation instructions based on OS
	switch runtime.GOOS {
	case "darwin":
		utils.Info("Install Docker Desktop for Mac: https://docs.docker.com/docker-for-mac/install/")
		utils.Info("Or use Homebrew: brew install --cask docker")
	case "linux":
		utils.Info("Install Docker Engine: https://docs.docker.com/engine/install/")
	case OSWindows:
		utils.Info("Install Docker Desktop for Windows: https://docs.docker.com/docker-for-windows/install/")
	default:
		utils.Info("Please install Docker for your platform: https://docs.docker.com/get-docker/")
	}

	utils.Success("Docker installation guidance provided")
	return nil
}

// GitHooks installs git hooks
func (Install) GitHooks() error {
	utils.Header("Installing Git Hooks")

	// Check if we're in a git repository
	if !utils.FileExists(".git") {
		utils.Warn("Not a git repository, skipping git hooks installation")
		return nil
	}

	// Create hooks directory if it doesn't exist
	hooksDir := ".git/hooks"
	if err := utils.EnsureDir(hooksDir); err != nil {
		return mageErrors.WrapError(err, "failed to create hooks directory")
	}

	// Install pre-commit hook
	preCommitPath := filepath.Join(hooksDir, "pre-commit")
	preCommitContent := `#!/bin/sh
# Run linting before commit
if command -v golangci-lint >/dev/null 2>&1; then
    golangci-lint run
fi
`

	if err := os.WriteFile(preCommitPath, []byte(preCommitContent), fileops.PermFileExecutablePrivate); err != nil {
		return mageErrors.WrapError(err, "failed to write pre-commit hook")
	}

	utils.Success("Git hooks installed")
	utils.Info("Pre-commit hook will run linting before commits")
	return nil
}

// CI installs CI tools
func (Install) CI() error {
	utils.Header("Installing CI Tools")

	// Install common CI tools
	tools := []struct {
		name string
		pkg  string
		desc string
	}{
		{"golangci-lint", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest", "CI linter"},
		{"govulncheck", "golang.org/x/vuln/cmd/govulncheck@latest", "Vulnerability scanner"},
		{"staticcheck", "honnef.co/go/tools/cmd/staticcheck@latest", "Static analyzer"},
	}

	installed := 0
	for _, tool := range tools {
		if utils.CommandExists(tool.name) {
			utils.Info("✓ %s already installed", tool.name)
			continue
		}

		utils.Info("Installing %s - %s", tool.name, tool.desc)
		if err := GetRunner().RunCmd("go", "install", tool.pkg); err != nil {
			utils.Warn("Failed to install %s: %v", tool.name, err)
		} else {
			installed++
		}
	}

	if installed > 0 {
		utils.Success("Installed %d CI tools", installed)
	} else {
		utils.Success("All CI tools already installed")
	}
	return nil
}

// Certs installs certificates
func (Install) Certs() error {
	utils.Header("Installing Certificates")

	// Check if we need to install certificates
	if !utils.CommandExists("openssl") {
		utils.Warn("OpenSSL not found, cannot generate certificates")
		utils.Info("Install OpenSSL for certificate management")
		return nil
	}

	// Create certs directory
	certsDir := "certs"
	if err := utils.EnsureDir(certsDir); err != nil {
		return mageErrors.WrapError(err, "failed to create certs directory")
	}

	// Generate self-signed certificate for development
	certPath := filepath.Join(certsDir, "server.crt")
	keyPath := filepath.Join(certsDir, "server.key")

	if utils.FileExists(certPath) && utils.FileExists(keyPath) {
		utils.Info("Certificates already exist")
		return nil
	}

	utils.Info("Generating self-signed certificate for development...")
	args := []string{
		"req", "-x509", "-newkey", "rsa:4096", "-keyout", keyPath,
		"-out", certPath, "-days", "365", "-nodes",
		"-subj", "/C=US/ST=Development/L=Local/O=Dev/CN=localhost",
	}

	if err := GetRunner().RunCmd("openssl", args...); err != nil {
		return mageErrors.WrapError(err, "failed to generate certificates")
	}

	utils.Success("Development certificates generated")
	utils.Info("Certificate: %s", certPath)
	utils.Info("Private key: %s", keyPath)
	return nil
}

// Package installs package
func (Install) Package() error {
	utils.Header("Installing Package")

	// Install dependencies first
	utils.Info("Installing dependencies...")
	if err := GetRunner().RunCmd("go", "mod", "download"); err != nil {
		return mageErrors.WrapError(err, "failed to download dependencies")
	}

	// Verify dependencies
	if err := GetRunner().RunCmd("go", "mod", "verify"); err != nil {
		return mageErrors.WrapError(err, "dependency verification failed")
	}

	installer := Install{}

	// Install the package
	if err := installer.Default(); err != nil {
		return mageErrors.WrapError(err, "package installation failed")
	}

	utils.Success("Package installed successfully")
	return nil
}

// All installs all components
func (Install) All() error {
	utils.Header("Installing All Components")

	installer := Install{}

	// Install dependencies
	if err := installer.Deps(); err != nil {
		utils.Warn("Failed to install dependencies: %v", err)
	}

	// Install development tools
	if err := installer.Tools(); err != nil {
		utils.Warn("Failed to install tools: %v", err)
	}

	// Install CI tools
	if err := installer.CI(); err != nil {
		utils.Warn("Failed to install CI tools: %v", err)
	}

	// Install git hooks
	if err := installer.GitHooks(); err != nil {
		utils.Warn("Failed to install git hooks: %v", err)
	}

	// Install the main binary
	if err := installer.Default(); err != nil {
		return mageErrors.WrapError(err, "binary installation failed")
	}

	utils.Success("All components installed")
	return nil
}
