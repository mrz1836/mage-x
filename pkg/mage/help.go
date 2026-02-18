// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/magefile/mage/mg"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/mrz1836/mage-x/pkg/common/env"
	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/mage/registry"
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

MAGE-X is a comprehensive Go build tooling suite that provides
powerful development tools with a friendly user experience.

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

	// Get the global registry
	reg := registry.Global()

	// Use the registry's categorized commands for more efficient display
	categorized := reg.CategorizedCommands()
	categoryOrder := reg.CategoryOrder()
	metadata := reg.Metadata()

	totalCommands := metadata.TotalCommands
	fmt.Printf("\n(%d commands available)\n", totalCommands)

	// Display commands by category
	for _, category := range categoryOrder {
		commands, exists := categorized[category]
		if !exists || len(commands) == 0 {
			continue
		}

		// Display category header
		categoryInfo := metadata.CategoryInfo[category]
		categoryName := category
		if categoryInfo.Name != "" {
			categoryName = categoryInfo.Name
		}

		fmt.Printf("\n%s Commands:\n", categoryName)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

		// Commands are already sorted by the registry
		for _, cmd := range commands {
			description := cmd.Description
			if description == "" {
				description = "No description available"
			}

			// Use the command's full name
			cmdName := cmd.FullName()
			if len(cmd.Aliases) > 0 {
				cmdName = cmd.Aliases[0] // Use primary alias if available
			}

			if _, err := fmt.Fprintf(w, "  %s\t%s\n", cmdName, description); err != nil {
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
	fmt.Printf("  magex -h NAMESPACE  # Show all commands in a namespace\n")

	return nil
}

// Command shows detailed help for a specific command or namespace
func (Help) Command() error {
	commandName := env.GetString("COMMAND", "")
	if commandName == "" {
		return errHelpCommandRequired
	}

	// Get the global registry for namespace checking
	reg := registry.Global()

	// First check if this is a namespace request
	namespaces := reg.Namespaces()
	for _, namespace := range namespaces {
		if strings.EqualFold(namespace, commandName) {
			return showHelpNamespace(reg, namespace)
		}
	}

	// If not a namespace, try to get specific command help
	cmd, err := getCommandHelp(commandName)
	if err != nil {
		return fmt.Errorf("failed to get command help for %s: %w", commandName, err)
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
				"mage configure:update  # Update configuration settings",
			},
		},
		{
			Category: "Version Management",
			Examples: []string{
				"mage version:show",
				"mage version:modules  # List sub-modules",
				"mage version:check  # Check for updates",
				"mage version:update  # Update to latest",
				"mage version:bump bump=minor push",
				"mage version:bump module=models bump=patch  # Bump sub-module",
				"mage version:bump module=all bump=minor  # Bump all sub-modules",
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
powerful Go build tooling suite.

🎯 What is MAGE-X?

MAGE-X is a comprehensive build tooling suite for Go projects that
provides powerful development tools with a friendly user experience.
It follows the philosophy of "Write Once, Mage Everywhere" - create your
build configuration once and use it across all your projects.

📦 Step 1: Installation

First, install MAGE-X in your Go project:

  go get github.com/mrz1836/mage-x
  go install github.com/magefile/mage@latest

🏗️ Step 2: Your First Build

Build your project:
  mage build

Run tests:
  mage test

🎭 Step 5: Interactive Mode

For a guided experience:
  mage configure:update

Or show help:
  mage help

🚀 Step 6: Advanced Features

- Releases: mage release
- Code quality: mage lint
- Security scanning: mage toolsVulnCheck
- Check for updates: mage updateCheck

📖 Next Steps

1. Read the documentation: mage help:commands
2. Try the configuration: mage configure:update
3. Set up CI/CD: mage test:ci && mage lint:ci
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
	shell := env.GetString("SHELL", "bash")

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

// showHelpNamespace displays help for all commands in a namespace (for Help namespace)
func showHelpNamespace(reg *registry.Registry, namespace string) error {
	commands := reg.ListByNamespace(namespace)
	if len(commands) == 0 {
		fmt.Printf("❌ No commands found in namespace '%s'\n", namespace)
		return nil
	}

	// Show namespace help header
	utils.Header(fmt.Sprintf("📖 Namespace Help: %s", namespace))

	// Get namespace info if available
	metadata := reg.Metadata()
	titleCaser := cases.Title(language.English)
	categoryInfo, hasInfo := metadata.CategoryInfo[titleCaser.String(namespace)]
	if hasInfo && categoryInfo.Description != "" {
		fmt.Printf("📝 Description: %s\n", categoryInfo.Description)
	}

	fmt.Printf("\n🔧 Available Commands in %s namespace (%d commands):\n\n", namespace, len(commands))

	// Display commands in the namespace
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	for _, cmd := range commands {
		fullName := cmd.FullName()
		description := cmd.Description
		if description == "" {
			description = "No description available"
		}

		// Truncate long descriptions
		if len(description) > 70 {
			description = description[:67] + "..."
		}

		if _, err := fmt.Fprintf(w, "  %s\t%s\n", fullName, description); err != nil {
			return fmt.Errorf("failed to write namespace help: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush namespace help: %w", err)
	}

	// Show usage examples
	fmt.Printf("\n💡 Usage Examples:\n")
	for i, cmd := range commands {
		if i >= 3 { // Show max 3 examples
			fmt.Printf("  ... and %d more commands\n", len(commands)-i)
			break
		}
		fmt.Printf("  magex %s\n", cmd.FullName())
	}

	// Show general help hint
	fmt.Printf("\n🔍 For detailed help on any command:\n")
	fmt.Printf("  magex help:command command=%s:<method>\n", namespace)
	fmt.Printf("  Example: magex help:command command=%s:%s\n", namespace, commands[0].Method)

	return nil
}

// getAllCommands returns all available commands with help information from the registry
func getAllCommands() []HelpCommand {
	// Get the global registry and ensure commands are registered
	reg := registry.Global()

	// Get all commands from the registry
	commands := reg.List()

	// Convert registry commands to HelpCommand format
	helpCommands := make([]HelpCommand, 0, len(commands))
	for _, cmd := range commands {
		helpCmd := HelpCommand{
			Name:        cmd.FullName(),
			Namespace:   cmd.Namespace,
			Description: cmd.Description,
			Usage:       cmd.Usage,
			Examples:    cmd.Examples,
			SeeAlso:     cmd.SeeAlso,
		}

		// Convert options if available
		if len(cmd.Options) > 0 {
			helpCmd.Options = make([]HelpOption, len(cmd.Options))
			for i, opt := range cmd.Options {
				helpCmd.Options[i] = HelpOption{
					Name:        opt.Name,
					Description: opt.Description,
					Default:     opt.Default,
					Required:    opt.Required,
				}
			}
		}

		helpCommands = append(helpCommands, helpCmd)
	}

	return helpCommands
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
    opts="build:default build:all build:clean build:generate build:linux build:darwin build:windows build:install build:prebuild test:default test:full test:unit test:race test:cover test:bench test:benchshort test:fuzz test:fuzzshort test:integration test:short test:coverrace test:ci lint:default lint:all lint:fix lint:vet lint:fumpt lint:version lint:go lint:yaml lint:json format:default format:all format:gofmt format:fumpt format:imports format:go format:yaml format:json format:fix format:check vet:default vet:all vet:parallel vet:shadow vet:strict release:default deps:default deps:update deps:tidy deps:download deps:outdated deps:clean deps:graph deps:why deps:verify tools:default tools:install tools:update tools:check tools:vulncheck tools:verify install:default install:local install:binary install:tools install:go install:stdlib install:systemwide install:deps install:mage install:githooks install:ci install:certs install:package install:all uninstall mod:update mod:tidy mod:verify mod:download mod:clean mod:graph mod:why mod:vendor mod:init docs:default docs:generate docs:serve docs:build docs:check docs:examples docs:lint docs:spell docs:links docs:api docs:markdown docs:readme docs:changelog docs:clean docs:godocs git:status git:diff git:tag git:log git:branch git:pull git:commit git:init git:add git:clone git:push version:show version:modules version:check version:update version:bump version:changelog version:tag version:compare version:validate metrics:loc metrics:coverage metrics:complexity configure:init configure:show configure:update configure:export configure:import configure:validate configure:schema generate:default generate:clean update:check update:install help:default help:commands help:command help:examples help:gettingstarted help:completions help:topics bench:default bench:run bench:profile bench:compare bench:report bench:regression bench:memory bench:cpu yaml:init yaml:validate yaml:show yaml:template yaml:format yaml:check yaml:merge yaml:convert yaml:schema"

    COMPREPLY=( $(compgen -W "${opts}" -- ${cur}) )
    return 0
}

complete -F _mage_completions mage
`

	completionFile := filepath.Join(os.Getenv("HOME"), ".mage_completion.bash")

	fileOps := fileops.New()
	if err := fileOps.File.WriteFile(completionFile, []byte(script), fileops.PermFile); err != nil {
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
                'release:Create releases'
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
	if err := os.MkdirAll(filepath.Dir(completionFile), fileops.PermDirSensitive); err != nil { // #nosec G703 -- completionFile path is from HOME env + fixed subpath
		return fmt.Errorf("failed to create completion directory: %w", err)
	}

	fileOps := fileops.New()
	if err := fileOps.File.WriteFile(completionFile, []byte(script), fileops.PermFile); err != nil {
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
complete -c mage -n '__fish_use_subcommand' -a 'release' -d 'Create releases'
complete -c mage -n '__fish_use_subcommand' -a 'yaml' -d 'Configuration management'
complete -c mage -n '__fish_use_subcommand' -a 'version' -d 'Version management'
complete -c mage -n '__fish_use_subcommand' -a 'help' -d 'Help system'

# Namespace commands
complete -c mage -n '__fish_seen_subcommand_from build' -a 'default all clean'
complete -c mage -n '__fish_seen_subcommand_from test' -a 'default unit race cover bench'
complete -c mage -n '__fish_seen_subcommand_from release' -a 'stable beta edge'
`

	configDir := filepath.Join(os.Getenv("HOME"), ".config", "fish", "completions")
	completionFile := filepath.Join(configDir, "mage.fish")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, fileops.PermDirSensitive); err != nil { // #nosec G703 -- configDir is from HOME env + fixed subpath
		return fmt.Errorf("failed to create completion directory: %w", err)
	}

	fileOps := fileops.New()
	if err := fileOps.File.WriteFile(completionFile, []byte(script), fileops.PermFile); err != nil {
		return fmt.Errorf("failed to write fish completion: %w", err)
	}

	utils.Success("Fish completions generated: %s", completionFile)
	utils.Info("Completions will be loaded automatically")

	return nil
}
