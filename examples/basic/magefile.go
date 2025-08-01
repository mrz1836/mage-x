//go:build mage
// +build mage

package main

import (
	// Import all tasks from MAGE-X
	"github.com/mrz1836/mage-x/pkg/mage"
)

// Re-export Build type
type Build = mage.Build

// Default target when running "mage" without arguments
func Default() error {
	var b Build
	return b.Default()
}

// This is the simplest possible magefile.go
// It imports all tasks from MAGE-X and sets Build as the default

// Run "mage -l" to see all available tasks:
// - build, build:all, build:linux, build:darwin, build:windows, etc.
// - test, test:unit, test:race, test:cover, etc.
// - lint, lint:fix, lint:fmt, etc.
// - deps, deps:update, deps:tidy, etc.
// - tools:install, tools:verify, etc.
