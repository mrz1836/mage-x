// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/common/fileops"
	"github.com/mrz1836/go-mage/pkg/utils"
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

	fmt.Println(`🎯 MAGE-X: Write Once, Mage Everywhere

MAGE-X is a comprehensive Go build automation toolkit that provides
enterprise-grade development tools with a friendly user experience.

Quick Start:
  mage build              # Build your project
  mage test               # Run tests
  mage lint               # Run linter
  mage release:stable     # Create stable release
  mage interactive        # Start interactive mode

Available Commands:
  mage help:commands      # List all available commands
  mage help:examples      # Show usage examples
  mage help:getting-started # Getting started guide
  mage help:completions   # Generate shell completions

For detailed help on any command:
  mage help:command COMMAND_NAME

For interactive help:
  mage interactive:help

Documentation:
  https://github.com/mrz1836/go-mage`)

	return nil
}

// Commands lists all available commands with descriptions
func (Help) Commands() error {
	utils.Header("📋 Available Commands")

	commands := getAllCommands()

	// Group by namespace
	namespaces := make(map[string][]HelpCommand)
	for _, cmd := range commands {
		if cmd.Namespace == "" {
			cmd.Namespace = "core"
		}
		namespaces[cmd.Namespace] = append(namespaces[cmd.Namespace], cmd)
	}

	// Sort namespaces
	var sortedNamespaces []string
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

		for _, cmd := range nsCommands {
			fmt.Fprintf(w, "  %s\t%s\n", cmd.Name, cmd.Description)
		}

		w.Flush()
	}

	fmt.Println("\nUsage:")
	fmt.Println("  mage COMMAND [OPTIONS]")
	fmt.Println("  mage NAMESPACE:COMMAND [OPTIONS]")
	fmt.Println("  mage help:command COMMAND_NAME")

	return nil
}

// Command shows detailed help for a specific command
func (Help) Command() error {
	commandName := utils.GetEnv("COMMAND", "")
	if commandName == "" {
		return fmt.Errorf("COMMAND environment variable is required. Usage: COMMAND=<name> mage help:command")
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

	if len(cmd.Options) > 0 {
		fmt.Println("\nOptions:")
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		for _, opt := range cmd.Options {
			required := ""
			if opt.Required {
				required = " (required)"
			}
			defaultVal := ""
			if opt.Default != "" {
				defaultVal = fmt.Sprintf(" [default: %s]", opt.Default)
			}
			fmt.Fprintf(w, "  %s\t%s%s%s\n", opt.Name, opt.Description, required, defaultVal)
		}
		w.Flush()
	}

	if len(cmd.Examples) > 0 {
		fmt.Println("\nExamples:")
		for _, example := range cmd.Examples {
			fmt.Printf("  %s\n", example)
		}
	}

	if len(cmd.SeeAlso) > 0 {
		fmt.Println("\nSee Also:")
		for _, related := range cmd.SeeAlso {
			fmt.Printf("  mage help:command %s\n", related)
		}
	}

	return nil
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
				"mage init:webapi --name=myapi",
				"mage yaml:init  # Create mage.yaml configuration",
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
				"mage release:stable",
				"mage release:beta",
				"mage release:edge",
				"VERSION=v1.2.3 mage release:stable",
				"mage releases:status  # Show release status",
			},
		},
		{
			Category: "Interactive Mode",
			Examples: []string{
				"mage interactive  # Start interactive mode",
				"mage interactive:wizard  # Start guided wizard",
				"mage recipes:list  # List available recipes",
				"RECIPE=fresh-start mage recipes:run",
			},
		},
		{
			Category: "Version Management",
			Examples: []string{
				"mage version:show",
				"mage version:check  # Check for updates",
				"mage version:update  # Update to latest",
				"BUMP=minor mage version:bump",
			},
		},
		{
			Category: "Configuration",
			Examples: []string{
				"mage yaml:init  # Create mage.yaml",
				"mage yaml:validate  # Validate configuration",
				"mage yaml:show  # Show current configuration",
				"PROJECT_TYPE=cli mage yaml:template",
			},
		},
	}

	for _, category := range examples {
		fmt.Printf("\n%s:\n", category.Category)
		for _, example := range category.Examples {
			fmt.Printf("  %s\n", example)
		}
	}

	fmt.Println("\nTips:")
	fmt.Println("  • Use environment variables to pass parameters")
	fmt.Println("  • Add VERBOSE=true for detailed output")
	fmt.Println("  • Use interactive mode for guided assistance")
	fmt.Println("  • Check mage.yaml for project-specific configuration")

	return nil
}

// GettingStarted shows getting started guide
func (Help) GettingStarted() error {
	utils.Header("🚀 Getting Started with MAGE-X")

	fmt.Println(`Welcome to MAGE-X! This guide will help you get started with the most
powerful Go build automation toolkit.

🎯 What is MAGE-X?

MAGE-X is a comprehensive build automation toolkit for Go projects that
provides enterprise-grade development tools with a friendly user experience.
It follows the philosophy of "Write Once, Mage Everywhere" - create your
build configuration once and use it across all your projects.

📦 Step 1: Installation

First, install MAGE-X in your Go project:

  go get github.com/mrz1836/go-mage
  go install github.com/magefile/mage@latest

🏗️ Step 2: Initialize Your Project

For a new project:
  mage init:cli --name=myapp --module=github.com/user/myapp

For an existing project:
  mage init:upgrade

🔧 Step 3: Basic Configuration

Create a mage.yaml configuration file:
  mage yaml:init

Edit the configuration to match your project needs.

🏃 Step 4: Your First Build

Build your project:
  mage build

Run tests:
  mage test

🎭 Step 5: Interactive Mode

For a guided experience:
  mage interactive

Or start with the wizard:
  mage interactive:wizard

📚 Step 6: Explore Recipes

Discover pre-built patterns:
  mage recipes:list

Run a recipe:
  RECIPE=fresh-start mage recipes:run

🚀 Step 7: Advanced Features

- Multi-channel releases: mage release:stable
- Code quality: mage lint
- Security scanning: mage security:vulncheck
- Auto-updates: mage version:check

📖 Next Steps

1. Read the documentation: mage help:commands
2. Try the interactive wizard: mage interactive:wizard
3. Set up CI/CD: RECIPE=ci-setup mage recipes:run
4. Configure releases: mage releases:status

🆘 Getting Help

- mage help:command COMMAND_NAME
- mage interactive:help
- mage help:examples
- GitHub: https://github.com/mrz1836/go-mage

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
		return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shell)
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
		{"getting-started", "Getting started guide", "mage help:getting-started"},
		{"commands", "List all commands", "mage help:commands"},
		{"examples", "Usage examples", "mage help:examples"},
		{"interactive", "Interactive mode help", "mage interactive:help"},
		{"recipes", "Recipe system", "mage recipes:list"},
		{"configuration", "Configuration with mage.yaml", "mage yaml:show"},
		{"releases", "Release management", "mage releases:status"},
		{"completions", "Shell completions", "mage help:completions"},
	}

	fmt.Println("\nAvailable Help Topics:")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	for _, topic := range topics {
		fmt.Fprintf(w, "  %s\t%s\t%s\n", topic.Name, topic.Description, topic.Command)
	}
	w.Flush()

	fmt.Println("\nUsage:")
	fmt.Println("  mage help:TOPIC")
	fmt.Println("  mage help:command COMMAND_NAME")

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
				{Name: "GO_BUILD_TAGS", Description: "Build tags", Default: ""},
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
				"VERBOSE=true mage test",
			},
			Options: []HelpOption{
				{Name: "VERBOSE", Description: "Verbose output", Default: "false"},
				{Name: "TEST_TIMEOUT", Description: "Test timeout", Default: "10m"},
				{Name: "TEST_RACE", Description: "Enable race detector", Default: "false"},
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
				"LINT_TIMEOUT=5m mage lint",
			},
			Options: []HelpOption{
				{Name: "LINT_FIX", Description: "Fix issues automatically", Default: "false"},
				{Name: "LINT_TIMEOUT", Description: "Lint timeout", Default: "5m"},
			},
			SeeAlso: []string{"format", "test", "security"},
		},

		{
			Name:        "format",
			Namespace:   "format",
			Description: "Format code",
			Usage:       "mage format [options]",
			Examples: []string{
				"mage format",
				"mage format:fumpt",
				"mage format:imports",
			},
			SeeAlso: []string{"lint", "test"},
		},

		// Interactive commands
		{
			Name:        "interactive",
			Namespace:   "interactive",
			Description: "Start interactive mode",
			Usage:       "mage interactive",
			Examples: []string{
				"mage interactive",
				"mage interactive:wizard",
				"mage interactive:help",
			},
			SeeAlso: []string{"help", "recipes"},
		},

		// Recipe commands
		{
			Name:        "recipes",
			Namespace:   "recipes",
			Description: "Recipe system for common patterns",
			Usage:       "mage recipes:COMMAND",
			Examples: []string{
				"mage recipes:list",
				"mage recipes:show fresh-start",
				"RECIPE=fresh-start mage recipes:run",
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
			Usage:       "mage release:CHANNEL",
			Examples: []string{
				"mage release:stable",
				"mage release:beta",
				"VERSION=v1.2.3 mage release:stable",
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
			Usage:       "mage init:TYPE [options]",
			Examples: []string{
				"mage init:cli --name=myapp",
				"mage init:library",
				"mage init:upgrade",
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
			Name:        "yaml",
			Namespace:   "yaml",
			Description: "Configuration management",
			Usage:       "mage yaml:COMMAND",
			Examples: []string{
				"mage yaml:init",
				"mage yaml:validate",
				"mage yaml:show",
			},
			SeeAlso: []string{"init", "help"},
		},

		// Version commands
		{
			Name:        "version",
			Namespace:   "version",
			Description: "Version management",
			Usage:       "mage version:COMMAND",
			Examples: []string{
				"mage version:show",
				"mage version:check",
				"mage version:update",
			},
			SeeAlso: []string{"release", "update"},
		},

		// Help commands
		{
			Name:        "help",
			Namespace:   "help",
			Description: "Help system",
			Usage:       "mage help:COMMAND",
			Examples: []string{
				"mage help",
				"mage help:commands",
				"mage help:examples",
			},
			SeeAlso: []string{"interactive", "topics"},
		},
	}
}

// getCommandHelp returns help information for a specific command
func getCommandHelp(name string) (HelpCommand, error) {
	commands := getAllCommands()

	for _, cmd := range commands {
		if cmd.Name == name || cmd.Namespace+":"+cmd.Name == name {
			return cmd, nil
		}
	}

	return HelpCommand{}, fmt.Errorf("command not found: %s", name)
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

    # Basic commands
    opts="build test lint format interactive recipes release init yaml version help"

    # Namespace commands
    if [[ ${cur} == *:* ]]; then
        case "${cur%:*}" in
            build)
                COMPREPLY=( $(compgen -W "default all clean" -- "${cur#*:}") )
                return 0
                ;;
            test)
                COMPREPLY=( $(compgen -W "default unit race cover bench" -- "${cur#*:}") )
                return 0
                ;;
            release)
                COMPREPLY=( $(compgen -W "stable beta edge" -- "${cur#*:}") )
                return 0
                ;;
            init)
                COMPREPLY=( $(compgen -W "library cli webapi microservice tool" -- "${cur#*:}") )
                return 0
                ;;
        esac
    fi

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
	if err := os.MkdirAll(filepath.Dir(completionFile), 0o755); err != nil {
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
	if err := os.MkdirAll(configDir, 0o755); err != nil {
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
