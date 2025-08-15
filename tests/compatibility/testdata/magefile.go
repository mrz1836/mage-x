//go:build mage
// +build mage

package main

import (
	// Import MAGE-X for automatic command registration
	_ "github.com/mrz1836/mage-x/pkg/mage/auto"
)

// Custom commands can be added here alongside MAGE-X built-ins

// Deploy is a custom deployment command
func Deploy() error {
	// Your custom deployment logic here
	return nil
}

// The MAGE-X auto import above provides all standard commands:
// - build, test, lint, format, clean
// - release, docker, deps, tools
// - and 80+ more commands across 24 namespaces
//
// Run 'magex -l' to see all available commands
