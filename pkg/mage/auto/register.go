// Package auto provides automatic registration of MAGE-X commands for backward compatibility
// This allows users to import MAGE-X and get all commands without the magex binary
package auto

import (
	"github.com/mrz1836/mage-x/pkg/mage"
)

// This file provides backward compatibility for MAGE-X 1.0 users
// who want to use imports instead of the magex binary

// All namespace types are re-exported for Mage visibility
type (
	Build     = mage.Build
	Test      = mage.Test
	Lint      = mage.Lint
	Format    = mage.Format
	Deps      = mage.Deps
	Git       = mage.Git
	Release   = mage.Release
	Docs      = mage.Docs
	Tools     = mage.Tools
	Generate  = mage.Generate
	Update    = mage.Update
	Mod       = mage.Mod
	Metrics   = mage.Metrics
	Bench     = mage.Bench
	Vet       = mage.Vet
	Configure = mage.Configure
	Help      = mage.Help
)

// Convenience wrapper functions for the most common commands
// These allow users to call commands without namespace syntax

// BuildCmd builds the application (alias for Build.Default)
func BuildCmd() error {
	var b mage.Build
	return b.Default()
}

// TestCmd runs tests (alias for Test.Default)
func TestCmd() error {
	var t mage.Test
	return t.Default()
}

// LintCmd runs linting (alias for Lint.Default)
func LintCmd() error {
	var l mage.Lint
	return l.Default()
}

// FormatCmd formats code (alias for Format.Default)
func FormatCmd() error {
	var f mage.Format
	return f.Default()
}

// CleanCmd cleans build artifacts (alias for Build.Clean)
func CleanCmd() error {
	var b mage.Build
	return b.Clean()
}

// InstallCmd installs the application (alias for Build.Install)
func InstallCmd() error {
	var b mage.Build
	return b.Install()
}

// DepsCmd manages dependencies (alias for Deps.Default)
func DepsCmd() error {
	var d mage.Deps
	return d.Default()
}

// ReleaseCmd creates a release (alias for Release.Default)
func ReleaseCmd() error {
	var r mage.Release
	return r.Default()
}

// Note: For the full zero-boilerplate experience, use the magex binary instead:
//   go install github.com/mrz1836/mage-x/cmd/magex@latest
//   magex update:install  # Get latest stable release with proper version info
//   magex build
//   magex test
//   etc.
