//go:build mage
// +build mage

package main

import (
	"fmt"
	"github.com/mrz1836/go-mage/pkg/mage"
)

// Build compiles the application using the new namespace interface
func Build() error {
	fmt.Println("🔨 Building application...")
	build := mage.NewBuildNamespace()
	return build.Default()
}

// Test runs the test suite using the new namespace interface
func Test() error {
	fmt.Println("🧪 Running tests...")
	test := mage.NewTestNamespace()
	return test.Unit()
}

// Lint runs code analysis using the new namespace interface
func Lint() error {
	fmt.Println("🔍 Running linter...")
	lint := mage.NewLintNamespace()
	return lint.Default()
}

// Format formats the code using the new namespace interface
func Format() error {
	fmt.Println("📝 Formatting code...")
	format := mage.NewFormatNamespace()
	return format.Default()
}

// Clean removes build artifacts using the new namespace interface
func Clean() error {
	fmt.Println("🧹 Cleaning build artifacts...")
	build := mage.NewBuildNamespace()
	return build.Clean()
}

// All runs the complete build pipeline
func All() error {
	fmt.Println("🚀 Running complete build pipeline...")
	
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
	
	fmt.Println("✅ All tasks completed successfully!")
	return nil
}

// Install builds and installs the binary
func Install() error {
	fmt.Println("📦 Installing application...")
	build := mage.NewBuildNamespace()
	return build.Install()
}

// PreBuild runs pre-build tasks
func PreBuild() error {
	fmt.Println("🔧 Running pre-build tasks...")
	build := mage.NewBuildNamespace()
	return build.PreBuild()
}