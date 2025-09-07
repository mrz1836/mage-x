package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

// DiscoveredCommand represents a custom command found in magefile.go
type DiscoveredCommand struct {
	Name        string // Full command name (e.g., "deploy" or "pipeline:ci")
	Description string
	IsNamespace bool
	Namespace   string
	Method      string
}

// CommandDiscovery handles discovery and caching of custom commands
type CommandDiscovery struct {
	commands []DiscoveredCommand
	loaded   bool
	verbose  bool
	registry *registry.Registry
}

// NewCommandDiscovery creates a new command discovery instance
func NewCommandDiscovery(reg *registry.Registry) *CommandDiscovery {
	return &CommandDiscovery{
		verbose:  os.Getenv("MAGE_X_VERBOSE") == trueValue || os.Getenv("MAGEX_VERBOSE") == trueValue,
		registry: reg,
	}
}

// Discover finds all custom commands in magefile.go
func (d *CommandDiscovery) Discover() error {
	if d.loaded {
		return nil // Already discovered
	}

	// Use the simplified loader to discover commands
	loader := registry.NewLoader(nil)
	commands, err := loader.DiscoverUserCommands(".")
	if err != nil {
		// In verbose mode or when debugging, show the actual error
		if d.verbose || os.Getenv("MAGEX_VERBOSE") == trueValue {
			fmt.Printf("Warning: Failed to discover custom commands: %v\n", err)
		}
		return fmt.Errorf("failed to discover commands: %w", err)
	}

	// Convert to our format
	d.commands = make([]DiscoveredCommand, 0, len(commands))
	for _, cmd := range commands {
		discovered := DiscoveredCommand{
			Description: cmd.Description,
			IsNamespace: cmd.IsNamespace,
			Namespace:   cmd.Namespace,
			Method:      cmd.Method,
		}

		// Set the command name based on type
		if cmd.IsNamespace && cmd.Method != "" {
			// Namespace method: Pipeline.CI -> pipeline:ci
			discovered.Name = strings.ToLower(cmd.Namespace) + ":" + strings.ToLower(cmd.Method)
		} else {
			// Simple function: Deploy -> deploy
			discovered.Name = strings.ToLower(cmd.Name)
		}

		// Don't skip commands that exist as built-in - custom commands should override built-ins
		// This allows users to customize behavior of standard commands like build, test, etc.
		if d.registry == nil {
			d.commands = append(d.commands, discovered)
			continue
		}

		// Still log if we're overriding a built-in command, but allow it
		if _, exists := d.registry.Get(discovered.Name); exists && d.verbose {
			fmt.Printf("Custom command '%s' will override built-in command\n", discovered.Name)
		}

		// Also skip functions that look like namespace wrappers (e.g., BuildDefault -> build:default)
		if d.isLikelyNamespaceWrapper(discovered.Name) {
			if d.verbose {
				fmt.Printf("Skipping '%s' - appears to be a namespace wrapper function\n", discovered.Name)
			}
			continue
		}

		d.commands = append(d.commands, discovered)
	}

	d.loaded = true

	if d.verbose && len(d.commands) > 0 {
		fmt.Printf("Discovered %d custom commands:\n", len(d.commands))
		for _, cmd := range d.commands {
			fmt.Printf("  • %s - %s\n", cmd.Name, cmd.Description)
		}
	}

	return nil
}

// HasCommand checks if a command exists in the discovered commands
func (d *CommandDiscovery) HasCommand(name string) bool {
	if err := d.Discover(); err != nil {
		return false
	}

	name = strings.ToLower(name)
	for _, cmd := range d.commands {
		if cmd.Name == name {
			return true
		}
	}
	return false
}

// GetCommand returns information about a discovered command
func (d *CommandDiscovery) GetCommand(name string) (*DiscoveredCommand, bool) {
	if err := d.Discover(); err != nil {
		return nil, false
	}

	name = strings.ToLower(name)
	for _, cmd := range d.commands {
		if cmd.Name == name {
			return &cmd, true
		}
	}
	return nil, false
}

// ListCommands returns all discovered commands
func (d *CommandDiscovery) ListCommands() ([]DiscoveredCommand, error) {
	if err := d.Discover(); err != nil {
		return nil, err
	}
	return d.commands, nil
}

// GetCommandsForHelp returns commands formatted for help display
func (d *CommandDiscovery) GetCommandsForHelp() []string {
	if err := d.Discover(); err != nil {
		return nil
	}

	helpLines := make([]string, 0, len(d.commands))
	for _, cmd := range d.commands {
		desc := cmd.Description
		if desc == "" {
			desc = "Custom command"
		}
		helpLines = append(helpLines, fmt.Sprintf("  %-20s %s (custom)", cmd.Name, desc))
	}
	return helpLines
}

// isLikelyNamespaceWrapper checks if a function name looks like a namespace wrapper
// e.g., BuildDefault, TestUnit, LintFix correspond to build:default, test:unit, lint:fix
func (d *CommandDiscovery) isLikelyNamespaceWrapper(funcName string) bool {
	if d.registry == nil {
		return false
	}

	// Check common patterns: NamespaceMethod -> namespace:method
	lower := strings.ToLower(funcName)

	// Try to split CamelCase into namespace:method patterns
	commonNamespaces := []string{
		"build", "test", "lint", "format", "deps", "git", "release", "docs",
		"tools", "generate", "mod", "audit", "help", "version", "install",
		"enterprise", "configure", "init", "workflow", "bench", "vet",
	}

	for _, ns := range commonNamespaces {
		if strings.HasPrefix(lower, ns) && len(lower) > len(ns) {
			// Extract the method part
			method := lower[len(ns):]

			// Check if namespace:method exists as built-in
			candidate := ns + ":" + method
			if _, exists := d.registry.Get(candidate); exists {
				return true
			}

			// Also check just the namespace (for Default methods)
			if _, exists := d.registry.Get(ns); exists && method == "default" {
				return true
			}
		}
	}

	return false
}

// Clear clears the discovery cache (useful for testing)
func (d *CommandDiscovery) Clear() {
	d.commands = nil
	d.loaded = false
}
