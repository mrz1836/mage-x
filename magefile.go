//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"

	"github.com/mrz1836/go-mage/pkg/mage"
	"github.com/mrz1836/go-mage/pkg/utils"
)

// Re-export all namespaces from mage package
type (
	Build   = mage.Build
	Test    = mage.Test
	Lint    = mage.Lint
	Tools   = mage.Tools
	Deps    = mage.Deps
	Mod     = mage.Mod
	Docs    = mage.Docs
	Git     = mage.Git
	Release = mage.Release
	Metrics = mage.Metrics
	Version = mage.Version
	Install = mage.Install
)

// Default target
var Default = BuildDefault

// BuildDefault is the default build target
func BuildDefault() error {
	var b Build
	return b.Default()
}

// TestDefault runs the default test suite
func TestDefault() error {
	var t Test
	return t.Default()
}

// LintDefault runs the default linter
func LintDefault() error {
	var l Lint
	return l.Default()
}

// TestUnit runs unit tests without linting
func TestUnit() error {
	var t Test
	return t.Unit()
}

// TestRace runs tests with race detector
func TestRace() error {
	var t Test
	return t.Race()
}

// InstallStdlib installs Go standard library (for cross-compilation)
func InstallStdlib() error {
	utils.Header("Installing Go Standard Library")

	cfg, err := mage.GetConfig()
	if err != nil {
		return err
	}

	// Install standard library for all configured platforms
	for _, platform := range cfg.Build.Platforms {
		p, err := utils.ParsePlatform(platform)
		if err != nil {
			return err
		}

		utils.Info("Installing stdlib for %s/%s...", p.OS, p.Arch)

		// Set environment for the target platform
		os.Setenv("GOOS", p.OS)
		os.Setenv("GOARCH", p.Arch)
		defer func() {
			os.Unsetenv("GOOS")
			os.Unsetenv("GOARCH")
		}()

		// Install the standard library
		if err := mage.GetRunner().RunCmd("go", "install", "-a", "std"); err != nil {
			return fmt.Errorf("failed to install stdlib for %s: %w", platform, err)
		}
	}

	utils.Success("Standard library installed for all platforms")
	return nil
}

// Uninstall removes the installed binary
func Uninstall() error {
	var i Install
	return i.Uninstall()
}

// ReleaseDefault creates a new release (default)
func ReleaseDefault() error {
	var r Release
	return r.Default()
}

// DepsUpdate updates all dependencies (equivalent to "make update")
func DepsUpdate() error {
	var d Deps
	return d.Update()
}

// DepsTidy cleans up go.mod and go.sum
func DepsTidy() error {
	var d Deps
	return d.Tidy()
}

// DepsDownload downloads all dependencies
func DepsDownload() error {
	var d Deps
	return d.Download()
}

// DepsOutdated shows outdated dependencies
func DepsOutdated() error {
	var d Deps
	return d.Outdated()
}

// DepsAudit audits dependencies for vulnerabilities
func DepsAudit() error {
	var d Deps
	return d.Audit()
}

// ToolsUpdate updates all development tools
func ToolsUpdate() error {
	var t Tools
	return t.Update()
}

// ToolsInstall installs all required development tools
func ToolsInstall() error {
	var t Tools
	return t.Install()
}

// ToolsCheck checks if all required tools are available
func ToolsCheck() error {
	var t Tools
	return t.Check()
}

// ModUpdate updates go.mod file
func ModUpdate() error {
	var m Mod
	return m.Update()
}

// ModTidy tidies the go.mod file
func ModTidy() error {
	var m Mod
	return m.Tidy()
}

// ModVerify verifies module checksums
func ModVerify() error {
	var m Mod
	return m.Verify()
}

// ModDownload downloads modules
func ModDownload() error {
	var m Mod
	return m.Download()
}

// DocsGenerate generates documentation
func DocsGenerate() error {
	var d Docs
	return d.Generate()
}

// DocsServe serves documentation locally
func DocsServe() error {
	var d Docs
	return d.Serve()
}

// DocsBuild builds static documentation
func DocsBuild() error {
	var d Docs
	return d.Build()
}

// DocsCheck validates documentation
func DocsCheck() error {
	var d Docs
	return d.Check()
}

// GitStatus shows git repository status
func GitStatus() error {
	var g Git
	return g.Status()
}

// GitCommit commits changes (interactive)
func GitCommit() error {
	var g Git
	return g.Commit()
}

// GitTag creates and pushes a new tag
func GitTag() error {
	var g Git
	return g.Tag()
}

// GitPush pushes changes to remote
func GitPush() error {
	var g Git
	return g.Push("origin", "main")
}

// VersionShow displays current version information
func VersionShow() error {
	var v Version
	return v.Show()
}

// VersionBump bumps the version (interactive)
func VersionBump() error {
	var v Version
	return v.Bump()
}

// VersionCheck checks version information
func VersionCheck() error {
	var v Version
	return v.Check()
}

// BuildDocker builds Docker containers
func BuildDocker() error {
	var b Build
	return b.Docker()
}

// BuildClean cleans build artifacts
func BuildClean() error {
	var b Build
	return b.Clean()
}

// BuildGenerate generates code before building
func BuildGenerate() error {
	var b Build
	return b.Generate()
}

// TestCover runs tests with coverage analysis
func TestCover() error {
	var t Test
	return t.Cover()
}

// TestBench runs benchmark tests
func TestBench() error {
	var t Test
	return t.Bench()
}

// TestFuzz runs fuzz tests
func TestFuzz() error {
	var t Test
	return t.Fuzz()
}

// TestIntegration runs integration tests
func TestIntegration() error {
	var t Test
	return t.Integration()
}

// LintAll runs all linting checks
func LintAll() error {
	var l Lint
	return l.All()
}

// LintFix automatically fixes linting issues
func LintFix() error {
	var l Lint
	return l.Fix()
}

// MetricsLOC analyzes lines of code
func MetricsLOC() error {
	var m Metrics
	return m.LOC()
}

// MetricsCoverage generates coverage reports
func MetricsCoverage() error {
	var m Metrics
	return m.Coverage()
}

// MetricsComplexity analyzes code complexity
func MetricsComplexity() error {
	var m Metrics
	return m.Complexity()
}

// InstallTools installs all development tools
func InstallTools() error {
	var i Install
	return i.Tools()
}

// InstallBinary installs the project binary
func InstallBinary() error {
	var i Install
	return i.Binary()
}

// TestShort runs short tests (excludes integration tests) 
func TestShort() error {
	var t Test
	return t.Short()
}

// TestCoverRace runs tests with both coverage and race detector
func TestCoverRace() error {
	var t Test
	return t.CoverRace()
}

// LintVersion shows the linter version information
func LintVersion() error {
	var l Lint
	return l.Version()
}

// BuildAll builds for all configured platforms
func BuildAll() error {
	var b Build
	return b.All()
}
