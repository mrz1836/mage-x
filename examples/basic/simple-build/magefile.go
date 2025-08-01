//go:build mage
// +build mage

package main

import (
	"fmt"

	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Build compiles the application using the new namespace interface
func Build() error {
	utils.Info("🔨 Building application...")
	build := mage.NewBuildNamespace()
	return build.Default()
}

// Test runs the test suite using the new namespace interface
func Test() error {
	utils.Info("🧪 Running tests...")
	test := mage.NewTestNamespace()
	return test.Unit()
}

// Lint runs code analysis using the new namespace interface
func Lint() error {
	utils.Info("🔍 Running linter...")
	lint := mage.NewLintNamespace()
	return lint.Default()
}

// Format formats the code using the new namespace interface
func Format() error {
	utils.Info("📝 Formatting code...")
	format := mage.NewFormatNamespace()
	return format.Default()
}

// Clean removes build artifacts using the new namespace interface
func Clean() error {
	utils.Info("🧹 Cleaning build artifacts...")
	build := mage.NewBuildNamespace()
	return build.Clean()
}

// All runs the complete build pipeline
func All() error {
	utils.Info("🚀 Running complete build pipeline...")

	// Format first
	if err := Format(); err != nil {
		return fmt.Errorf("formatting failed: %w", err)
	}

	// Then lint
	if err := Lint(); err != nil {
		return fmt.Errorf("linting failed: %w", err)
	}

	// Then test
	if err := Test(); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	// Finally build
	if err := Build(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	utils.Info("✅ All tasks completed successfully!")
	return nil
}

// Install builds and installs the binary
func Install() error {
	utils.Info("📦 Installing application...")
	build := mage.NewBuildNamespace()
	return build.Install()
}

// PreBuild runs pre-build tasks
func PreBuild() error {
	utils.Info("🔧 Running pre-build tasks...")
	build := mage.NewBuildNamespace()
	return build.PreBuild()
}
