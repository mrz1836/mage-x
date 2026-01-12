// Package auto provides automatic registration of MAGE-X commands for
// backward compatibility with Mage import-based usage patterns.
//
// # Overview
//
// This package allows users to import MAGE-X and get all commands available
// without requiring the magex binary. It re-exports all namespace types
// making them visible to Mage's reflection-based command discovery.
//
// # Usage
//
// Import the package in your magefile.go:
//
//	//go:build mage
//
//	package main
//
//	import (
//	    _ "github.com/mrz1836/mage-x/pkg/mage/auto"
//	)
//
// # Re-exported Types
//
// All MAGE-X namespace types are available:
//
//	auto.Build     // Build commands
//	auto.Test      // Test commands
//	auto.Lint      // Lint commands
//	auto.Format    // Format commands
//	auto.Deps      // Dependency commands
//	auto.Git       // Git commands
//	auto.Release   // Release commands
//	auto.Docs      // Documentation commands
//	// ... and more
//
// # Convenience Functions
//
// Simple wrapper functions for common commands:
//
//	auto.BuildCmd()   // Build the application
//	auto.TestCmd()    // Run tests
//	auto.LintCmd()    // Run linting
//	auto.FormatCmd()  // Format code
//	auto.CleanCmd()   // Clean build artifacts
//
// # Backward Compatibility
//
// This package is designed for MAGE-X 1.0 users who prefer the import-based
// approach over using the magex binary directly.
package auto
