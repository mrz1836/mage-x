// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for help operations
var (
	errHelpCommandRequired = errors.New("command parameter is required. Usage: magex help:command command=<name>")
	errUnsupportedShell    = errors.New("unsupported shell (supported: bash, zsh, fish)")
	errCommandNotFound     = errors.New("command not found")
)

// Help namespace for help system and documentation
type Help mg.Namespace

// HelpCommand represents a command with help information
type HelpCommand struct {
	Name        string
	Namespace   string
	Description string
	Usage       string
	Examples    []string
	Options     []HelpOption
	SeeAlso     []string
}

// HelpOption represents a command option
type HelpOption struct {
	Name        string
	Description string
	Default     string
	Required    bool
}

// Default shows general help
func (Help) Default() error {
	utils.Header("📖 MAGE-X Help System")

	fmt.Printf(`🎯 MAGE-X: Write Once, Mage Everywhere

MAGE-X is a comprehensive Go build automation toolkit that provides
enterprise-grade development tools with a friendly user experience.

Quick Start:
  magex build              # Build your project
  magex test               # Run tests
  magex bench              # Run benchmarks
  magex lint               # Run linter
  magex release            # Create a release
  magex help               # Show this help

Available Commands:
  magex help:commands       # List all available commands
  magex help:examples       # Show usage examples
  magex help:gettingstarted # Getting started guide
  magex help:completions    # Generate shell completions

For detailed help on any command:
  magex help:command COMMAND_NAME

Documentation:
  https://github.com/mrz1836/mage-x
`)

	return nil
}

// Commands lists all available commands with descriptions
func (Help) Commands() error {
	utils.Header("📋 Available Commands")

	commands := getAllCommands()

	// Group by namespace
	namespaces := make(map[string][]HelpCommand)
	for i := range commands {
		if commands[i].Namespace == "" {
			commands[i].Namespace = "core"
		}
		namespaces[commands[i].Namespace] = append(namespaces[commands[i].Namespace], commands[i])
	}

	// Sort namespaces
	sortedNamespaces := make([]string, 0, len(namespaces))
	for ns := range namespaces {
		sortedNamespaces = append(sortedNamespaces, ns)
	}
	sort.Strings(sortedNamespaces)

	// Display commands by namespace
	for _, ns := range sortedNamespaces {
		fmt.Printf("\n%s Commands:\n", strings.ToUpper(ns[:1])+ns[1:])

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

		// Sort commands within namespace
		nsCommands := namespaces[ns]
		sort.Slice(nsCommands, func(i, j int) bool {
			return nsCommands[i].Name < nsCommands[j].Name
		})

		for i := range nsCommands {
			if _, err := fmt.Fprintf(w, "  %s\t%s\n", nsCommands[i].Name, nsCommands[i].Description); err != nil {
				return fmt.Errorf("failed to write command help: %w", err)
			}
		}

		if err := w.Flush(); err != nil {
			return fmt.Errorf("failed to flush help output: %w", err)
		}
	}

	fmt.Printf("\nUsage:\n")
	fmt.Printf("  magex COMMAND [OPTIONS]\n")
	fmt.Printf("  magex help:command COMMAND_NAME\n")

	return nil
}

// Command shows detailed help for a specific command
func (Help) Command() error {
	commandName := utils.GetEnv("COMMAND", "")
	if commandName == "" {
		return errHelpCommandRequired
	}

	cmd, err := getCommandHelp(commandName)
	if err != nil {
		return err
	}

	utils.Header("📖 Command Help: " + cmd.Name)

	fmt.Printf("Description: %s\n", cmd.Description)

	if cmd.Namespace != "" {
		fmt.Printf("Namespace: %s\n", cmd.Namespace)
	}

	fmt.Printf("Usage: %s\n", cmd.Usage)

	if len(cmd.Options) == 0 {
		return nil
	}

	fmt.Printf("\nOptions:\n")
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	for _, opt := range cmd.Options {
		optionLine := formatOption(opt)
		if _, err := fmt.Fprintf(w, "  %s\t%s\n", opt.Name, optionLine); err != nil {
			return fmt.Errorf("failed to write option help: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush option help: %w", err)
	}

	if len(cmd.Examples) > 0 {
		fmt.Printf("\nExamples:\n")
		for _, example := range cmd.Examples {
			fmt.Printf("  %s\n", example)
		}
	}

	if len(cmd.SeeAlso) > 0 {
		fmt.Printf("\nSee Also:\n")
		for _, related := range cmd.SeeAlso {
			fmt.Printf("  magex helpCommand %s\n", related)
		}
	}

	return nil
}

// formatOption formats an option with its description, required marker, and default value
func formatOption(opt HelpOption) string {
	description := opt.Description

	if opt.Required {
		description += " (required)"
	}

	if opt.Default != "" {
		description += fmt.Sprintf(" [default: %s]", opt.Default)
	}

	return description
}

// Examples shows usage examples
func (Help) Examples() error {
	utils.Header("💡 Usage Examples")

	examples := []struct {
		Category string
		Examples []string
	}{
		{
			Category: "Project Setup",
			Examples: []string{
				"mage init:cli --name=myapp --module=github.com/user/myapp",
				"mage init:library --name=mylib",
				"mage configure:init  # Create mage.yaml configuration",
			},
		},
		{
			Category: "Building & Testing",
			Examples: []string{
				"mage build",
				"mage build:all  # Build for all platforms",
				"mage test",
				"mage test:race  # Run tests with race detector",
				"mage test:cover  # Run tests with coverage",
				"mage bench  # Run benchmarks",
			},
		},
		{
			Category: "Code Quality",
			Examples: []string{
				"mage lint",
				"mage lint:fix  # Run linter and fix issues",
				"mage format",
				"mage format:fumpt  # Use stricter formatting",
				"mage security:vulncheck  # Check for vulnerabilities",
			},
		},
		{
			Category: "Release Management",
			Examples: []string{
				"mage release:default",
				"mage release:test",
				"mage release:snapshot",
				"mage release:localinstall",
				"mage release godocs",
			},
		},
		{
			Category: "Interactive Mode",
			Examples: []string{
				"mage help  # Show help",
				"mage configure:update  # Start configuration wizard",
				"mage recipes:list  # List available recipes",
				"mage recipes:run recipe=fresh-start",
			},
		},
		{
			Category: "Version Management",
			Examples: []string{
				"mage version:show",
				"mage version:check  # Check for updates",
				"mage version:update  # Update to latest",
				"mage version:bump bump=minor push",
				"mage git:tag version=1.2.3",
				"mage git:commit message='fix: bug fix'",
			},
		},
		{
			Category: "Configuration",
			Examples: []string{
				"mage yaml:init  # Create mage.yaml",
				"mage yaml:validate  # Validate configuration",
				"mage yaml:show  # Show current configuration",
				"mage yaml:template type=cli",
			},
		},
	}

	for _, category := range examples {
		fmt.Printf("\n%s:\n", category.Category)
		for _, example := range category.Examples {
			fmt.Printf("  %s\n", example)
		}
	}

	fmt.Printf("\nTips:\n")
	fmt.Printf("  • Use 'param=value' format to pass parameters\n")
	fmt.Printf("  • Add VERBOSE=true for detailed output\n")
	fmt.Printf("  • Check mage.yaml for project-specific configuration\n")
	fmt.Printf("  • Use 'mage help' for beautiful command listing\n")

	return nil
}

// GettingStarted shows getting started guide
func (Help) GettingStarted() error {
	utils.Header("🚀 Getting Started with MAGE-X")

	utils.Info(`Welcome to MAGE-X! This guide will help you get started with the most
powerful Go build automation toolkit.

🎯 What is MAGE-X?

MAGE-X is a comprehensive build automation toolkit for Go projects that
provides enterprise-grade development tools with a friendly user experience.
It follows the philosophy of "Write Once, Mage Everywhere" - create your
build configuration once and use it across all your projects.

📦 Step 1: Installation

First, install MAGE-X in your Go project:

  go get github.com/mrz1836/mage-x
  go install github.com/magefile/mage@latest

🏗️ Step 2: Initialize Your Project

For a new project:
  mage init:cli --name=myapp --module=github.com/user/myapp

For an existing project:
  mage init:project

🔧 Step 3: Basic Configuration

Create a mage configuration:
  mage configure:init

Edit the configuration to match your project needs.

🏃 Step 4: Your First Build

Build your project:
  mage build

Run tests:
  mage test

🎭 Step 5: Interactive Mode

For a guided experience:
  mage configure:update

Or show help:
  mage help

📚 Step 6: Explore Recipes

Discover pre-built patterns:
  mage recipes:list

Run a recipe:
  mage recipes:run recipe=fresh-start

🚀 Step 7: Advanced Features

- Releases: mage release
- Code quality: mage lint
- Security scanning: mage toolsVulnCheck
- Check for updates: mage updateCheck

📖 Next Steps

1. Read the documentation: mage help:commands
2. Try the configuration wizard: mage configure:update
3. Set up CI/CD: mage recipes:run recipe=ci-setup
4. Show version: mage version:show

🆘 Getting Help

- mage help:command COMMAND_NAME
- mage help
- mage help:examples
- GitHub: https://github.com/mrz1836/mage-x

Happy coding with MAGE-X! 🎉`)

	return nil
}

// Completions generates shell completions
func (Help) Completions() error {
	shell := utils.GetEnv("SHELL", "bash")

	utils.Header("🔗 Shell Completions")

	switch shell {
	case "bash":
		return generateBashCompletions()
	case "zsh":
		return generateZshCompletions()
	case "fish":
		return generateFishCompletions()
	default:
		return fmt.Errorf("%w: %s", errUnsupportedShell, shell)
	}
}

// Topics shows help topics
func (Help) Topics() error {
	utils.Header("📚 Help Topics")

	topics := []struct {
		Name        string
		Description string
		Command     string
	}{
		{"getting-started", "Getting started guide", "mage help:gettingstarted"},
		{"commands", "List all commands", "mage help:commands"},
		{"examples", "Usage examples", "mage help:examples"},
		{"recipes", "Recipe system", "mage recipes:list"},
		{"configuration", "Configuration management", "mage configure:show"},
		{"version", "Version management", "mage version:show"},
		{"completions", "Shell completions", "mage help:completions"},
	}

	utils.Info("Available Help Topics:")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	for _, topic := range topics {
		if _, err := fmt.Fprintf(w, "  %s\t%s\t%s\n", topic.Name, topic.Description, topic.Command); err != nil {
			return fmt.Errorf("failed to write topic help: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush topic help: %w", err)
	}

	utils.Info("Usage:")
	utils.Info("  mage helpTOPIC")
	utils.Info("  mage help:command COMMAND_NAME")

	return nil
}

// Helper functions

// getAllCommands returns all available commands with help information
func getAllCommands() []HelpCommand {
	return []HelpCommand{
		// Core commands
		{
			Name:        "build",
			Namespace:   "build",
			Description: "Build the project",
			Usage:       "mage build [options]",
			Examples: []string{
				"mage build",
				"mage build:all",
				"mage build:clean",
			},
			Options: []HelpOption{
				{Name: "MAGE_X_BUILD_TAGS", Description: "Build tags", Default: ""},
				{Name: "GOOS", Description: "Target OS", Default: "current"},
				{Name: "GOARCH", Description: "Target architecture", Default: "current"},
			},
			SeeAlso: []string{"test", "lint", "release"},
		},

		{
			Name:        "test",
			Namespace:   "test",
			Description: "Run tests",
			Usage:       "mage test [options]",
			Examples: []string{
				"mage test",
				"mage test:race",
				"mage test:cover",
				"MAGE_X_VERBOSE=true mage test",
			},
			Options: []HelpOption{
				{Name: "MAGE_X_VERBOSE", Description: "Verbose output", Default: "false"},
				{Name: "MAGE_X_TEST_TIMEOUT", Description: "Test timeout", Default: "10m"},
				{Name: "MAGE_X_TEST_RACE", Description: "Enable race detector", Default: "false"},
			},
			SeeAlso: []string{"build", "lint", "bench"},
		},

		{
			Name:        "lint",
			Namespace:   "lint",
			Description: "Run linter",
			Usage:       "mage lint [options]",
			Examples: []string{
				"mage lint",
				"mage lint:fix",
				"MAGE_X_LINT_TIMEOUT=5m mage lint",
			},
			Options: []HelpOption{
				{Name: "LINT_FIX", Description: "Fix issues automatically", Default: "false"},
				{Name: "MAGE_X_LINT_TIMEOUT", Description: "Lint timeout", Default: "5m"},
			},
			SeeAlso: []string{"format", "test", "security"},
		},

		{
			Name:        "format",
			Namespace:   "format",
			Description: "Format code",
			Usage:       "mage format [options]",
			Examples: []string{
				"mage format:all",
				"mage lint:fumpt",
				"mage format:check",
			},
			SeeAlso: []string{"lint", "test"},
		},

		// Recipe commands
		{
			Name:        "recipes",
			Namespace:   "recipes",
			Description: "Recipe system for common patterns",
			Usage:       "mage recipes COMMAND",
			Examples: []string{
				"mage recipes:list",
				"mage recipes:show recipe=fresh-start",
				"mage recipes:run recipe=fresh-start",
				"mage recipes:search term=docker",
			},
			Options: []HelpOption{
				{Name: "RECIPE", Description: "Recipe name", Required: true},
				{Name: "TERM", Description: "Search term for recipes:search"},
			},
			SeeAlso: []string{"init", "interactive"},
		},

		// Release commands
		{
			Name:        "release",
			Namespace:   "release",
			Description: "Create releases",
			Usage:       "mage release [options]",
			Examples: []string{
				"mage release",
				"CHANNEL=beta mage release",
				"VERSION=v1.2.3 mage release",
			},
			Options: []HelpOption{
				{Name: "VERSION", Description: "Release version"},
				{Name: "GITHUB_TOKEN", Description: "GitHub token", Required: true},
			},
			SeeAlso: []string{"version", "releases"},
		},

		// Init commands
		{
			Name:        "init",
			Namespace:   "init",
			Description: "Initialize projects",
			Usage:       "mage init TYPE [options]",
			Examples: []string{
				"mage initCLI --name=myapp",
				"mage initLibrary",
				"mage initProject",
			},
			Options: []HelpOption{
				{Name: "PROJECT_NAME", Description: "Project name"},
				{Name: "PROJECT_MODULE", Description: "Module path"},
				{Name: "PROJECT_TYPE", Description: "Project type"},
			},
			SeeAlso: []string{"yaml", "recipes"},
		},

		// Configuration commands
		{
			Name:        "configure",
			Namespace:   "configure",
			Description: "Configuration management",
			Usage:       "mage configure COMMAND",
			Examples: []string{
				"mage configure:init",
				"mage configure:show",
				"mage configure:update",
			},
			SeeAlso: []string{"init", "help"},
		},

		// Version commands
		{
			Name:        "version",
			Namespace:   "version",
			Description: "Version management",
			Usage:       "mage version COMMAND",
			Examples: []string{
				"mage version:show",
				"mage version:check",
				"mage version:bump",
			},
			SeeAlso: []string{"release", "update"},
		},

		// Help commands
		{
			Name:        "help",
			Namespace:   "help",
			Description: "Help system",
			Usage:       "mage help COMMAND",
			Examples: []string{
				"mage help",
				"mage help:commands",
				"mage help:examples",
			},
			SeeAlso: []string{"interactive", "topics"},
		},

		// Audit commands
		{
			Name:        "audit",
			Namespace:   "audit",
			Description: "Security and compliance auditing",
			Usage:       "mage audit COMMAND",
			Examples: []string{
				"mage audit:show",
				"mage audit:stats",
				"mage audit:export",
				"mage audit:report",
			},
			SeeAlso: []string{"security", "tools"},
		},

		// Enterprise commands
		{
			Name:        "enterprise",
			Namespace:   "enterprise",
			Description: "Enterprise deployment and management",
			Usage:       "mage enterprise COMMAND",
			Examples: []string{
				"mage enterprise:init",
				"mage enterprise:deploy",
				"mage enterprise:status",
				"mage enterprise:backup",
			},
			SeeAlso: []string{"workflow", "integrations"},
		},

		// Workflow commands
		{
			Name:        "workflow",
			Namespace:   "workflow",
			Description: "Automated workflow management",
			Usage:       "mage workflow COMMAND",
			Examples: []string{
				"mage workflow:execute",
				"mage workflow:list",
				"mage workflow:create",
				"mage workflow:schedule",
			},
			SeeAlso: []string{"enterprise", "cli"},
		},

		// Install commands
		{
			Name:        "install",
			Namespace:   "install",
			Description: "Installation management",
			Usage:       "mage install COMMAND",
			Examples: []string{
				"mage install:tools",
				"mage install:binary",
				"mage install:all",
				"mage uninstall",
			},
			SeeAlso: []string{"tools", "update"},
		},

		// YAML commands
		{
			Name:        "yaml",
			Namespace:   "yaml",
			Description: "YAML configuration management",
			Usage:       "mage yaml COMMAND",
			Examples: []string{
				"mage yaml:init",
				"mage yaml:validate",
				"mage yaml:show",
				"mage yaml:format",
			},
			SeeAlso: []string{"configure", "format"},
		},

		// Dependencies commands
		{
			Name:        "deps",
			Namespace:   "deps",
			Description: "Dependency management",
			Usage:       "mage deps COMMAND",
			Examples: []string{
				"mage deps:update",
				"mage deps:update allow-major",
				"mage deps:tidy",
				"mage deps:download",
				"mage deps:audit",
			},
			SeeAlso: []string{"mod", "tools"},
		},

		// Module commands
		{
			Name:        "mod",
			Namespace:   "mod",
			Description: "Go module management",
			Usage:       "mage mod COMMAND",
			Examples: []string{
				"mage mod:tidy",
				"mage mod:update",
				"mage mod:verify",
				"mage mod:download",
			},
			SeeAlso: []string{"deps", "tools"},
		},

		// Documentation commands
		{
			Name:        "docs",
			Namespace:   "docs",
			Description: "Documentation generation and serving",
			Usage:       "mage docs COMMAND",
			Examples: []string{
				"mage docs:serve",
				"mage docs:generate",
				"mage docs:build",
				"mage docs:godocs",
			},
			SeeAlso: []string{"help", "build"},
		},

		// Git commands
		{
			Name:        "git",
			Namespace:   "git",
			Description: "Git operations and version control",
			Usage:       "mage git COMMAND",
			Examples: []string{
				"mage git:status",
				"mage git:commit message='fix: bug'",
				"mage git:tag version=1.2.3",
			},
			SeeAlso: []string{"version", "release"},
		},

		// Tools commands
		{
			Name:        "tools",
			Namespace:   "tools",
			Description: "Development tool management",
			Usage:       "mage tools COMMAND",
			Examples: []string{
				"mage tools:install",
				"mage tools:update",
				"mage tools:verify",
				"mage tools:vulncheck",
			},
			SeeAlso: []string{"deps", "install"},
		},

		// Benchmarking commands
		{
			Name:        "bench",
			Namespace:   "bench",
			Description: "Performance benchmarking and profiling",
			Usage:       "mage bench [options]",
			Examples: []string{
				"mage bench",
				"mage bench time=50ms",
				"mage bench:cpu time=30s",
				"mage bench:profile",
			},
			SeeAlso: []string{"test", "metrics"},
		},

		// Metrics commands
		{
			Name:        "metrics",
			Namespace:   "metrics",
			Description: "Code metrics and analysis",
			Usage:       "mage metrics COMMAND",
			Examples: []string{
				"mage metrics:loc",
				"mage metrics:coverage",
				"mage metrics:complexity",
			},
			SeeAlso: []string{"test", "bench"},
		},

		// Vet commands
		{
			Name:        "vet",
			Namespace:   "vet",
			Description: "Code vetting and static analysis",
			Usage:       "mage vet [options]",
			Examples: []string{
				"mage vet:default",
				"mage vet:all",
				"mage vet:shadow",
			},
			SeeAlso: []string{"lint", "test"},
		},

		// Update commands
		{
			Name:        "update",
			Namespace:   "update",
			Description: "Update management",
			Usage:       "mage update COMMAND",
			Examples: []string{
				"mage update:check",
				"mage update:install",
			},
			SeeAlso: []string{"version", "tools"},
		},

		// CLI commands
		{
			Name:        "cli",
			Namespace:   "cli",
			Description: "CLI operations and utilities",
			Usage:       "mage cli COMMAND",
			Examples: []string{
				"mage cli:help",
				"mage cli:version",
				"mage cli:completion",
			},
			SeeAlso: []string{"help", "wizard"},
		},

		// Wizard commands
		{
			Name:        "wizard",
			Namespace:   "wizard",
			Description: "Interactive setup wizards",
			Usage:       "mage wizard COMMAND",
			Examples: []string{
				"mage wizard:setup",
				"mage wizard:config",
				"mage wizard:project",
			},
			SeeAlso: []string{"init", "configure"},
		},

		// Generate commands
		{
			Name:        "generate",
			Namespace:   "generate",
			Description: "Code generation",
			Usage:       "mage generate [options]",
			Examples: []string{
				"mage generate:default",
				"mage generate:clean",
			},
			SeeAlso: []string{"build", "init"},
		},
	}
}

// getCommandHelp returns help information for a specific command
func getCommandHelp(name string) (HelpCommand, error) {
	commands := getAllCommands()

	for i := range commands {
		if commands[i].Name == name {
			return commands[i], nil
		}
	}

	return HelpCommand{}, fmt.Errorf("%w: %s", errCommandNotFound, name)
}

// generateBashCompletions generates bash completions
func generateBashCompletions() error {
	utils.Info("Generating bash completions...")

	script := `#!/bin/bash

_mage_completions() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # All available commands (300+ commands organized by namespace)
    opts="build:default build:all build:docker build:clean build:generate build:linux build:darwin build:windows build:install build:prebuild test:default test:full test:unit test:race test:cover test:bench test:benchshort test:fuzz test:fuzzshort test:integration test:short test:coverrace test:ci lint:default lint:all lint:fix lint:vet lint:fumpt lint:version lint:go lint:docker lint:yaml lint:markdown lint:shell lint:json lint:sql format:default format:all format:gofmt format:fumpt format:imports format:go format:yaml format:json format:markdown format:sql format:dockerfile format:shell format:fix format:check vet:default vet:all vet:parallel vet:shadow vet:strict release:default deps:default deps:update deps:tidy deps:download deps:outdated deps:audit deps:clean deps:graph deps:why deps:verify tools:default tools:install tools:update tools:check tools:vulncheck tools:verify install:default install:local install:binary install:tools install:go install:stdlib install:systemwide install:deps install:mage install:docker install:githooks install:ci install:certs install:package install:all uninstall mod:update mod:tidy mod:verify mod:download mod:clean mod:graph mod:why mod:vendor mod:init docs:default docs:generate docs:serve docs:build docs:check docs:examples docs:lint docs:spell docs:links docs:api docs:markdown docs:readme docs:changelog docs:clean docs:godocs git:status git:diff git:tag git:log git:branch git:pull git:commit git:init git:add git:clone git:push version:show version:check version:update version:bump version:changelog version:tag version:compare version:validate metrics:loc metrics:coverage metrics:complexity audit:show audit:stats audit:export audit:cleanup audit:enable audit:disable audit:report configure:init configure:show configure:update configure:enterprise configure:export configure:import configure:validate configure:schema generate:default generate:clean init:default init:project init:library init:cli init:webapi init:microservice init:tool init:upgrade init:templates init:config init:git init:mage init:ci init:docker init:docs init:license init:makefile init:editorconfig recipes:default recipes:list recipes:run recipes:show recipes:search recipes:create recipes:install recipes:update recipes:categories recipes:interactive update:check update:install update:auto update:history update:rollback help:default help:commands help:command help:examples help:gettingstarted help:completions help:topics enterprise:init enterprise:config enterprise:deploy enterprise:rollback enterprise:promote enterprise:status enterprise:backup enterprise:restore workflow:execute workflow:list workflow:status workflow:create workflow:validate workflow:schedule workflow:template workflow:history integrations:setup integrations:test integrations:sync integrations:notify integrations:status integrations:webhook integrations:export integrations:import bench:default bench:run bench:profile bench:compare bench:report bench:regression bench:memory bench:cpu yaml:init yaml:validate yaml:show yaml:template yaml:format yaml:check yaml:merge yaml:convert yaml:schema cli:default cli:help cli:version cli:completion cli:config cli:update cli:bulk cli:query cli:dashboard cli:batch cli:monitor cli:workspace cli:pipeline cli:compliance wizard:setup wizard:config wizard:project wizard:deploy wizard:troubleshoot"

    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    return 0
}

complete -F _mage_completions mage
`

	completionFile := filepath.Join(os.Getenv("HOME"), ".mage_completion.bash")

	fileOps := fileops.New()
	if err := fileOps.File.WriteFile(completionFile, []byte(script), 0o644); err != nil {
		return fmt.Errorf("failed to write bash completion: %w", err)
	}

	utils.Success("Bash completions generated: %s", completionFile)
	utils.Info("Add to your ~/.bashrc: source %s", completionFile)

	return nil
}

// generateZshCompletions generates zsh completions
func generateZshCompletions() error {
	utils.Info("Generating zsh completions...")

	script := `#compdef mage

_mage() {
    local context state line
    typeset -A opt_args

    _arguments -C \
        '1: :->commands' \
        '*: :->args' \
        && return 0

    case $state in
        commands)
            local commands=(
                'build:Build the project'
                'test:Run tests'
                'lint:Run linter'
                'format:Format code'
                'interactive:Start interactive mode'
                'recipes:Recipe system'
                'release:Create releases'
                'init:Initialize projects'
                'yaml:Configuration management'
                'version:Version management'
                'help:Help system'
            )
            _describe 'commands' commands
            ;;
        args)
            case $line[1] in
                build)
                    _arguments \
                        '--platform[Target platform]:platform:' \
                        '--tags[Build tags]:tags:'
                    ;;
                test)
                    _arguments \
                        '--verbose[Verbose output]' \
                        '--race[Race detector]' \
                        '--cover[Coverage]'
                    ;;
            esac
            ;;
    esac
}

_mage "$@"
`

	completionFile := filepath.Join(os.Getenv("HOME"), ".zsh", "completions", "_mage")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(completionFile), 0o750); err != nil {
		return fmt.Errorf("failed to create completion directory: %w", err)
	}

	fileOps := fileops.New()
	if err := fileOps.File.WriteFile(completionFile, []byte(script), 0o644); err != nil {
		return fmt.Errorf("failed to write zsh completion: %w", err)
	}

	utils.Success("Zsh completions generated: %s", completionFile)
	utils.Info("Add to your ~/.zshrc: fpath=(~/.zsh/completions $fpath)")

	return nil
}

// generateFishCompletions generates fish completions
func generateFishCompletions() error {
	utils.Info("Generating fish completions...")

	script := `# Fish completions for mage
complete -c mage -f

# Basic commands
complete -c mage -n '__fish_use_subcommand' -a 'build' -d 'Build the project'
complete -c mage -n '__fish_use_subcommand' -a 'test' -d 'Run tests'
complete -c mage -n '__fish_use_subcommand' -a 'lint' -d 'Run linter'
complete -c mage -n '__fish_use_subcommand' -a 'format' -d 'Format code'
complete -c mage -n '__fish_use_subcommand' -a 'interactive' -d 'Start interactive mode'
complete -c mage -n '__fish_use_subcommand' -a 'recipes' -d 'Recipe system'
complete -c mage -n '__fish_use_subcommand' -a 'release' -d 'Create releases'
complete -c mage -n '__fish_use_subcommand' -a 'init' -d 'Initialize projects'
complete -c mage -n '__fish_use_subcommand' -a 'yaml' -d 'Configuration management'
complete -c mage -n '__fish_use_subcommand' -a 'version' -d 'Version management'
complete -c mage -n '__fish_use_subcommand' -a 'help' -d 'Help system'

# Namespace commands
complete -c mage -n '__fish_seen_subcommand_from build' -a 'default all clean'
complete -c mage -n '__fish_seen_subcommand_from test' -a 'default unit race cover bench'
complete -c mage -n '__fish_seen_subcommand_from release' -a 'stable beta edge'
complete -c mage -n '__fish_seen_subcommand_from init' -a 'library cli webapi microservice tool'
`

	configDir := filepath.Join(os.Getenv("HOME"), ".config", "fish", "completions")
	completionFile := filepath.Join(configDir, "mage.fish")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		return fmt.Errorf("failed to create completion directory: %w", err)
	}

	fileOps := fileops.New()
	if err := fileOps.File.WriteFile(completionFile, []byte(script), 0o644); err != nil {
		return fmt.Errorf("failed to write fish completion: %w", err)
	}

	utils.Success("Fish completions generated: %s", completionFile)
	utils.Info("Completions will be loaded automatically")

	return nil
}
