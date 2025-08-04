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
	errHelpCommandRequired = errors.New("COMMAND environment variable is required. Usage: COMMAND=<name> mage helpCommand")
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

	utils.Info(`🎯 MAGE-X: Write Once, Mage Everywhere

MAGE-X is a comprehensive Go build automation toolkit that provides
enterprise-grade development tools with a friendly user experience.

Quick Start:
  mage build              # Build your project
  mage test               # Run tests
  mage lint               # Run linter
  mage release           # Create a release
  mage interactive        # Start interactive mode

Available Commands:
  mage helpCommands      # List all available commands
  mage helpExamples      # Show usage examples
  mage helpGettingStarted # Getting started guide
  mage helpCompletions   # Generate shell completions

For detailed help on any command:
  mage helpCommand COMMAND_NAME

For interactive help:
  mage help

Documentation:
  https://github.com/mrz1836/mage-x`)

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
		utils.Info("\n%s Commands:", strings.ToUpper(ns[:1])+ns[1:])

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

	utils.Info("\nUsage:")
	utils.Info("  mage COMMAND [OPTIONS]")
	utils.Info("  mage COMMAND [OPTIONS]")
	utils.Info("  mage helpCommand COMMAND_NAME")

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

	utils.Info("Description: %s", cmd.Description)

	if cmd.Namespace != "" {
		utils.Info("Namespace: %s", cmd.Namespace)
	}

	utils.Info("Usage: %s", cmd.Usage)

	if len(cmd.Options) == 0 {
		return nil
	}

	utils.Info("\nOptions:")
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
		utils.Info("\nExamples:")
		for _, example := range cmd.Examples {
			utils.Info("  %s", example)
		}
	}

	if len(cmd.SeeAlso) > 0 {
		utils.Info("\nSee Also:")
		for _, related := range cmd.SeeAlso {
			utils.Info("  mage helpCommand %s", related)
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
				"mage help  # Show help",
				"mage configureUpdate  # Start configuration wizard",
				"mage recipesList  # List available recipes",
				"RECIPE=fresh-start mage recipesRun",
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
		utils.Info("\n%s:", category.Category)
		for _, example := range category.Examples {
			utils.Info("  %s", example)
		}
	}

	utils.Info("\nTips:")
	utils.Info("  • Use environment variables to pass parameters")
	utils.Info("  • Add VERBOSE=true for detailed output")
	utils.Info("  • Use interactive mode for guided assistance")
	utils.Info("  • Check mage.yaml for project-specific configuration")

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
  mage initCLI --name=myapp --module=github.com/user/myapp

For an existing project:
  mage initProject

🔧 Step 3: Basic Configuration

Create a mage configuration:
  mage configureInit

Edit the configuration to match your project needs.

🏃 Step 4: Your First Build

Build your project:
  mage build

Run tests:
  mage test

🎭 Step 5: Interactive Mode

For a guided experience:
  mage configureUpdate

Or show help:
  mage help

📚 Step 6: Explore Recipes

Discover pre-built patterns:
  mage recipesList

Run a recipe:
  RECIPE=fresh-start mage recipesRun

🚀 Step 7: Advanced Features

- Releases: mage release
- Code quality: mage lint
- Security scanning: mage toolsVulnCheck
- Check for updates: mage updateCheck

📖 Next Steps

1. Read the documentation: mage helpCommands
2. Try the configuration wizard: mage configureUpdate
3. Set up CI/CD: RECIPE=ci-setup mage recipesRun
4. Show version: mage versionShow

🆘 Getting Help

- mage helpCommand COMMAND_NAME
- mage help
- mage helpExamples
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
		{"getting-started", "Getting started guide", "mage helpGettingStarted"},
		{"commands", "List all commands", "mage helpCommands"},
		{"examples", "Usage examples", "mage helpExamples"},
		{"recipes", "Recipe system", "mage recipesList"},
		{"configuration", "Configuration management", "mage configureShow"},
		{"version", "Version management", "mage versionShow"},
		{"completions", "Shell completions", "mage helpCompletions"},
	}

	utils.Info("\nAvailable Help Topics:")

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	for _, topic := range topics {
		if _, err := fmt.Fprintf(w, "  %s\t%s\t%s\n", topic.Name, topic.Description, topic.Command); err != nil {
			return fmt.Errorf("failed to write topic help: %w", err)
		}
	}
	if err := w.Flush(); err != nil {
		return fmt.Errorf("failed to flush topic help: %w", err)
	}

	utils.Info("\nUsage:")
	utils.Info("  mage helpTOPIC")
	utils.Info("  mage helpCommand COMMAND_NAME")

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
				"mage buildAll",
				"mage buildClean",
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
				"mage testRace",
				"mage testCover",
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
				"mage lintFix",
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
				"mage formatAll",
				"mage lintFumpt",
				"mage formatCheck",
			},
			SeeAlso: []string{"lint", "test"},
		},

		// Interactive commands
		{
			Name:        "help",
			Namespace:   "help",
			Description: "Show help",
			Usage:       "mage help",
			Examples: []string{
				"mage help",
				"mage helpCommands",
				"mage helpExamples",
			},
			SeeAlso: []string{"help", "recipes"},
		},

		// Recipe commands
		{
			Name:        "recipes",
			Namespace:   "recipes",
			Description: "Recipe system for common patterns",
			Usage:       "mage recipes COMMAND",
			Examples: []string{
				"mage recipesList",
				"mage recipesRun",
				"RECIPE=fresh-start mage recipesRun",
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
				"mage configureInit",
				"mage configureShow",
				"mage configureUpdate",
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
				"mage versionShow",
				"mage versionCheck",
				"mage versionBump",
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
				"mage helpCommands",
				"mage helpExamples",
			},
			SeeAlso: []string{"interactive", "topics"},
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

    # All available commands
    opts="build buildAll buildClean buildDocker buildGenerate test testFull testUnit testRace testCover testBench testBenchShort testFuzz testFuzzShort testIntegration testShort testCoverRace lint lintAll lintFix lintVet lintFumpt lintVersion formatAll formatCheck vetDefault vetAll release deps depsUpdate depsTidy depsDownload depsOutdated depsAudit tools toolsInstall toolsUpdate toolsCheck toolsVulnCheck install installTools installBinary installStdlib uninstall mod modUpdate modTidy modVerify modDownload docs docsGenerate docsServe docsBuild docsCheck git gitStatus gitCommit gitTag gitPush version versionShow versionBump versionCheck metrics metricsLOC metricsCoverage metricsComplexity audit auditShow configure configureInit configureShow configureUpdate generate generateDefault generateClean init initProject initCLI initLibrary recipes recipesList recipesRun update updateCheck updateSelf help helpCommands helpExamples helpGettingStarted helpCompletions"

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
