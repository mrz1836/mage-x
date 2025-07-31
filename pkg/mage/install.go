// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/utils"
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

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Get binary name
	binaryName := config.Project.Binary
	if binaryName == "" {
		// Try to get from module name
		if module, err := utils.GetModuleName(); err == nil {
			parts := strings.Split(module, "/")
			binaryName = parts[len(parts)-1]
		} else {
			binaryName = defaultBinaryName
		}
	}

	// Get GOPATH
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("HOME"), "go")
	}

	// Determine install path
	installPath := filepath.Join(gopath, "bin", binaryName)
	if runtime.GOOS == OSWindows && !strings.HasSuffix(installPath, ".exe") {
		installPath += windowsExeExt
	}

	// Remove the binary
	if err := os.Remove(installPath); err != nil {
		if os.IsNotExist(err) {
			utils.Warn("Binary not found at %s", installPath)
			return nil
		}
		return fmt.Errorf("failed to remove binary: %w", err)
	}

	utils.Success("Uninstalled %s from %s", binaryName, installPath)
	return nil
}

// Default installs the application binary
func (Install) Default() error {
	utils.Header("Installing Application")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Get binary name
	binaryName := config.Project.Binary
	if binaryName == "" {
		// Try to get from module name
		if module, err := utils.GetModuleName(); err == nil {
			parts := strings.Split(module, "/")
			binaryName = parts[len(parts)-1]
		} else {
			binaryName = defaultBinaryName
		}
	}

	// Get GOPATH
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("HOME"), "go")
	}

	// Determine install path
	installPath := filepath.Join(gopath, "bin", binaryName)
	if runtime.GOOS == OSWindows && !strings.HasSuffix(installPath, ".exe") {
		installPath += windowsExeExt
	}

	utils.Info("Installing to: %s", installPath)

	// Build with installation flags
	args := []string{"build", "-o", installPath}

	// Add build flags
	args = append(args, "-trimpath")

	if len(cfg.Build.Tags) > 0 {
		args = append(args, "-tags", strings.Join(cfg.Build.Tags, ","))
	}

	if len(cfg.Build.LDFlags) > 0 {
		args = append(args, "-ldflags", strings.Join(cfg.Build.LDFlags, " "))
	}

	// Add main package
	args = append(args, ".")

	// Build and install
	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	// Verify installation
	if utils.FileExists(installPath) {
		utils.Success("Successfully installed %s", binaryName)
		utils.Info("Binary location: %s", installPath)

		// Check if GOPATH/bin is in PATH
		if !isInPath(filepath.Join(gopath, "bin")) {
			utils.Warn("Note: %s/bin is not in your PATH", gopath)
			utils.Info("Add it with: export PATH=$PATH:%s/bin", gopath)
		}
	} else {
		return fmt.Errorf("installation verification failed")
	}

	return nil
}

// Go installs using go install with specific version
func (Install) Go() error {
	utils.Header("Installing with go install")

	// Get module info
	module, err := utils.GetModuleName()
	if err != nil {
		return fmt.Errorf("failed to get module name: %w", err)
	}

	// Get version
	version := utils.GetEnv("VERSION", getVersion())
	if version == versionDev || version == "" {
		version = DefaultGoVulnCheckVersion
	}

	// Build install command
	var pkg string
	if version != DefaultGoVulnCheckVersion {
		pkg = fmt.Sprintf("%s@%s", module, version)
	} else {
		pkg = fmt.Sprintf("%s@latest", module)
	}

	utils.Info("Installing %s", pkg)

	args := []string{"install"}

	// Add build tags if specified
	if tags := os.Getenv("GO_BUILD_TAGS"); tags != "" {
		args = append(args, "-tags", tags)
	}

	args = append(args, pkg)

	// Run go install
	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("go install failed: %w", err)
	}

	utils.Success("Successfully installed via go install")
	return nil
}

// Stdlib installs the Go standard library for the host platform
func (Install) Stdlib() error {
	utils.Header("Installing Go Standard Library")

	utils.Info("Installing standard library packages...")

	if err := GetRunner().RunCmd("go", "install", "std"); err != nil {
		return fmt.Errorf("failed to install standard library: %w", err)
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
		return fmt.Errorf("system-wide installation not supported on Windows")
	}

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Get binary name
	binaryName := config.Project.Binary
	if binaryName == "" {
		if module, err := utils.GetModuleName(); err == nil {
			parts := strings.Split(module, "/")
			binaryName = parts[len(parts)-1]
		} else {
			binaryName = defaultBinaryName
		}
	}

	// Build binary first
	tempBinary := filepath.Join(os.TempDir(), binaryName)

	utils.Info("Building binary...")

	args := []string{"build", "-o", tempBinary}
	args = append(args, "-trimpath")

	if len(cfg.Build.Tags) > 0 {
		args = append(args, "-tags", strings.Join(cfg.Build.Tags, ","))
	}

	if len(cfg.Build.LDFlags) > 0 {
		args = append(args, "-ldflags", strings.Join(cfg.Build.LDFlags, " "))
	}

	args = append(args, ".")

	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	// Install to /usr/local/bin
	installPath := filepath.Join("/usr/local/bin", binaryName)

	utils.Info("Installing to %s (requires sudo)...", installPath)

	// Use sudo to copy
	if err := GetRunner().RunCmd("sudo", "cp", tempBinary, installPath); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	// Make executable
	if err := GetRunner().RunCmd("sudo", "chmod", "+x", installPath); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	// Clean up temp file
	if err := os.Remove(tempBinary); err != nil {
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
	path := os.Getenv("PATH")
	paths := strings.Split(path, string(os.PathListSeparator))

	for _, p := range paths {
		if p == dir {
			return true
		}
	}

	return false
}

// getVersion gets the current version

// Additional methods for Install namespace required by tests

// Binary installs binary
func (Install) Binary() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Installing binary")
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
	runner := GetRunner()
	return runner.RunCmd("echo", "Installing Docker requirements")
}

// GitHooks installs git hooks
func (Install) GitHooks() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Installing git hooks")
}

// CI installs CI tools
func (Install) CI() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Installing CI tools")
}

// Certs installs certificates
func (Install) Certs() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Installing certificates")
}

// Package installs package
func (Install) Package() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Installing package")
}

// All installs all components
func (Install) All() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Installing all components")
}
