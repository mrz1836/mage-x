// magex is a drop-in replacement for mage that includes all MAGE-X commands built-in
// Users can run MAGE-X commands without any magefile.go, achieving true "Write Once, Mage Everywhere"
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/mrz1836/mage-x/pkg/mage/embed"
	"github.com/mrz1836/mage-x/pkg/mage/registry"
	"github.com/mrz1836/mage-x/pkg/utils"
)

const (
	version = "1.0.0"
	banner  = `
███╗   ███╗ █████╗  ██████╗ ███████╗      ██╗  ██╗
████╗ ████║██╔══██╗██╔════╝ ██╔════╝      ╚██╗██╔╝
██╔████╔██║███████║██║  ███╗█████╗  █████╗ ╚███╔╝
██║╚██╔╝██║██╔══██║██║   ██║██╔══╝  ╚════╝ ██╔██╗
██║ ╚═╝ ██║██║  ██║╚██████╔╝███████╗      ██╔╝ ██╗
╚═╝     ╚═╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝      ╚═╝  ╚═╝
   🪄 MAGE-X - Write Once, Mage Everywhere
`
)

// ErrMagefileExists is returned when trying to initialize a magefile that already exists
var (
	ErrMagefileExists = errors.New("magefile.go already exists")
)

// Flags holds all command line flags
type Flags struct {
	List      *bool
	ListLong  *bool
	Help      *bool
	HelpLong  *bool
	Version   *bool
	Verbose   *bool
	Compile   *string
	Init      *bool
	Clean     *bool
	Debug     *bool
	Namespace *bool
	Search    *string
	Timeout   *string
	Force     *bool
}

// initFlags initializes all command line flags
func initFlags() *Flags {
	return &Flags{
		List:      flag.Bool("l", false, "list available commands"),
		ListLong:  flag.Bool("list", false, "list available commands (verbose)"),
		Help:      flag.Bool("h", false, "show help"),
		HelpLong:  flag.Bool("help", false, "show help"),
		Version:   flag.Bool("version", false, "show version"),
		Verbose:   flag.Bool("v", false, "verbose output"),
		Compile:   flag.String("compile", "", "compile a magefile for use with mage"),
		Init:      flag.Bool("init", false, "initialize a new magefile with MAGE-X imports"),
		Clean:     flag.Bool("clean", false, "clean MAGE-X cache and temporary files"),
		Debug:     flag.Bool("debug", false, "enable debug output"),
		Namespace: flag.Bool("n", false, "show commands organized by namespace"),
		Search:    flag.String("search", "", "search for commands"),
		Timeout:   flag.String("t", "", "timeout for command execution"),
		Force:     flag.Bool("f", false, "force operation"),
	}
}

func main() {
	// Initialize flags
	flags := initFlags()

	// Custom usage function
	flag.Usage = showUsage

	// Parse command line arguments
	flag.Parse()

	// Set environment variables based on flags (ignore errors as these are non-critical)
	if *flags.Verbose {
		if err := os.Setenv("MAGEX_VERBOSE", "true"); err != nil {
			fmt.Printf("Warning: Could not set MAGEX_VERBOSE: %v\n", err)
		}
		if err := os.Setenv("MAGE_X_VERBOSE", "1"); err != nil {
			fmt.Printf("Warning: Could not set MAGE_X_VERBOSE: %v\n", err)
		}
	}
	if *flags.Debug {
		if err := os.Setenv("MAGEX_DEBUG", "true"); err != nil {
			fmt.Printf("Warning: Could not set MAGEX_DEBUG: %v\n", err)
		}
		if err := os.Setenv("MAGE_X_DEBUG", "1"); err != nil {
			fmt.Printf("Warning: Could not set MAGE_X_DEBUG: %v\n", err)
		}
	}

	// Initialize the command registry early (needed for help)
	reg := registry.Global()
	embed.RegisterAll(reg)

	// Load user's magefile if it exists
	loader := registry.NewLoader(reg)
	if err := loader.LoadUserMagefile("."); err != nil {
		if *flags.Verbose {
			_, err = fmt.Fprintf(os.Stderr, "Warning: failed to load user magefile: %v\n", err)
			if err != nil {
				return
			}
		}
	}

	// Get arguments early for help processing
	args := flag.Args()

	// Handle special flags
	if *flags.Version {
		showVersion()
		return
	}

	// Handle help - support both general and command-specific help
	if *flags.Help || *flags.HelpLong {
		if len(args) > 0 {
			// Command-specific help: magex -h build
			showUnifiedHelp(args[0])
		} else {
			// General help: magex -h
			showUnifiedHelp("")
		}
		return
	}

	// Handle help command: magex help [command]
	if len(args) > 0 && args[0] == "help" {
		if len(args) > 1 {
			showUnifiedHelp(args[1])
		} else {
			showUnifiedHelp("")
		}
		return
	}

	if *flags.Init {
		if err := initMagefile(); err != nil {
			_, err = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			if err != nil {
				return
			}
			os.Exit(1)
		}
		return
	}

	if *flags.Clean {
		cleanCache()
		return
	}

	// Handle list commands
	if *flags.List || *flags.ListLong {
		if *flags.Namespace {
			listByNamespace(reg)
		} else {
			listCommands(reg, *flags.ListLong)
		}
		return
	}

	// Handle search
	if *flags.Search != "" {
		searchCommands(reg, *flags.Search)
		return
	}

	// Handle namespace listing
	if *flags.Namespace {
		listByNamespace(reg)
		return
	}

	// Handle compilation request
	if *flags.Compile != "" {
		compileForMage(*flags.Compile)
		return
	}

	// Process command execution (args already defined above)
	if len(args) == 0 {
		// No command specified, show available commands
		fmt.Print(banner)
		utils.Println("\n📋 Available Commands (run 'magex -l' for full list):")
		showQuickList(reg)
		utils.Println("\n💡 Run 'magex <command>' to execute a command")
		utils.Println("   Run 'magex -h' for help")
		return
	}

	// Execute the command
	command := args[0]
	commandArgs := args[1:]

	// Convert mage-style namespace:method to our format
	command = normalizeCommandName(command)

	// Show banner in verbose mode
	if *flags.Verbose {
		fmt.Print(banner)
	}

	// Execute the command
	if err := reg.Execute(command, commandArgs...); err != nil {
		_, err = fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
		if err != nil {
			return
		}
		os.Exit(1)
	}
}

// showUsage displays custom usage information
func showUsage() {
	// Use the unified help system for consistent output
	// Registry should already be initialized by main()
	showUnifiedHelp("")
}

// showVersion displays version information
func showVersion() {
	fmt.Printf("MAGE-X version %s\n", version)
	utils.Println("Built-in commands from all MAGE-X namespaces")
	utils.Println("Compatible with Mage build tool")
}

// showUnifiedHelp displays the comprehensive unified help system
func showUnifiedHelp(command string) {
	reg := registry.Global()
	embed.RegisterAll(reg)

	if command == "" {
		showGeneralHelp(reg)
	} else {
		showCommandHelp(reg, command)
	}
}

// showGeneralHelp displays the main help with all commands categorized
func showGeneralHelp(reg *registry.Registry) {
	// Show banner
	fmt.Print(banner)
	utils.Println("\n📚 Universal Build Automation for Go")

	// Show usage
	showUsageSection()

	// Show quick start
	showQuickStartSection()

	// Show options
	showOptionsSection()

	// Show categorized commands
	showCategorizedCommands(reg)

	// Show tips and footer
	showTipsAndFooter()
}

// showUsageSection displays usage information
func showUsageSection() {
	fmt.Printf("\n📋 Usage: magex [options] [command] [arguments...]\n")
	fmt.Printf("\nMAGE-X is a drop-in replacement for Mage with 215+ built-in commands.\n")
	fmt.Printf("Zero configuration needed - works immediately in any Go project!\n")
}

// showQuickStartSection displays quick start commands
func showQuickStartSection() {
	fmt.Printf("\n🎯 Quick Start:\n")
	fmt.Printf("  magex build      # Build your project\n")
	fmt.Printf("  magex test       # Run tests\n")
	fmt.Printf("  magex lint       # Check code quality\n")
	fmt.Printf("  magex release    # Create a release\n")
}

// showOptionsSection displays command line options
func showOptionsSection() {
	fmt.Printf("\n⚡ Common Options:\n")
	fmt.Printf("  -h, --help       Show this comprehensive help\n")
	fmt.Printf("  -l, --list       List all commands (simple format)\n")
	fmt.Printf("  -n, --namespace  Show commands by namespace\n")
	fmt.Printf("  -v, --verbose    Verbose output\n")
	fmt.Printf("  --version        Show version information\n")
	fmt.Printf("  -search <term>   Search for specific commands\n")
	fmt.Printf("  -init            Create a magefile with MAGE-X imports\n")
	fmt.Printf("  -clean           Clean MAGE-X cache and temporary files\n")
	fmt.Printf("  -debug           Enable debug output\n")
}

// showCategorizedCommands displays all commands organized by category
func showCategorizedCommands(reg *registry.Registry) {
	categorized := reg.CategorizedCommands()
	categoryOrder := reg.CategoryOrder()
	metadata := reg.Metadata()

	totalCommands := metadata.TotalCommands
	fmt.Printf("\n📋 Available Commands (%d total):\n", totalCommands)

	for _, category := range categoryOrder {
		commands, exists := categorized[category]
		if !exists || len(commands) == 0 {
			continue
		}

		categoryInfo := metadata.CategoryInfo[category]
		if categoryInfo.Name == "" {
			categoryInfo.Name = cases.Title(language.English).String(category)
		}
		if categoryInfo.Icon == "" {
			categoryInfo.Icon = "📋"
		}

		fmt.Printf("\n%s %s:\n", categoryInfo.Icon, categoryInfo.Name)

		for _, cmd := range commands {
			// Show command name and description
			cmdName := cmd.FullName()
			if len(cmd.Aliases) > 0 {
				cmdName = cmd.Aliases[0] // Use primary alias if available
			}

			description := cmd.Description
			if description == "" {
				description = "No description available"
			}

			// Truncate long descriptions
			if len(description) > 60 {
				description = description[:57] + "..."
			}

			// Show deprecated warning
			if cmd.Deprecated != "" {
				fmt.Printf("  %-20s ⚠️  DEPRECATED: %s\n", cmdName, cmd.Deprecated)
			} else {
				fmt.Printf("  %-20s %s\n", cmdName, description)
			}
		}
	}
}

// showTipsAndFooter displays tips and documentation links
func showTipsAndFooter() {
	fmt.Printf("\n💡 Tips:\n")
	fmt.Printf("  • Use 'magex -h <command>' for detailed command help\n")
	fmt.Printf("  • Use 'magex -n' to see commands organized by namespace\n")
	fmt.Printf("  • Use 'magex -search <term>' to find specific commands\n")
	fmt.Printf("  • Add VERBOSE=true for detailed output\n")
	fmt.Printf("  • Create magefile.go to add custom commands\n")

	fmt.Printf("\n📖 More Help:\n")
	fmt.Printf("  • Documentation: https://github.com/mrz1836/mage-x\n")
	fmt.Printf("  • Examples: magex -search example\n")
	fmt.Printf("  • Configuration: magex configure:show\n")
}

// showCommandHelp displays detailed help for a specific command
func showCommandHelp(reg *registry.Registry, commandName string) {
	cmd, exists := reg.Get(commandName)
	if !exists {
		// Try to find similar commands
		suggestions := reg.Search(commandName)
		fmt.Printf("❌ Unknown command '%s'\n", commandName)
		if len(suggestions) > 0 {
			fmt.Printf("\n🔍 Did you mean:\n")
			for i, suggestion := range suggestions {
				if i >= 5 {
					break // Limit suggestions
				}
				fmt.Printf("  • %s - %s\n", suggestion.FullName(), suggestion.Description)
			}
		}
		return
	}

	// Show detailed command information
	fmt.Printf("\n📖 Command Help: %s\n", cmd.FullName())
	fmt.Printf("\n%s\n", strings.Repeat("=", 50))

	// Description
	if cmd.Description != "" {
		fmt.Printf("\n📝 Description:\n  %s\n", cmd.Description)
	}

	if cmd.LongDescription != "" {
		fmt.Printf("\n📚 Detailed Description:\n  %s\n", cmd.LongDescription)
	}

	// Usage
	usage := cmd.Usage
	if usage == "" {
		usage = fmt.Sprintf("magex %s [options]", cmd.FullName())
	}
	fmt.Printf("\n🔧 Usage:\n  %s\n", usage)

	// Category and namespace
	if cmd.Category != "" {
		fmt.Printf("\n📂 Category: %s\n", cmd.Category)
	}
	if cmd.Namespace != "" {
		fmt.Printf("🏷️  Namespace: %s\n", cmd.Namespace)
	}

	// Aliases
	if len(cmd.Aliases) > 0 {
		fmt.Printf("\n🔗 Aliases:\n")
		for _, alias := range cmd.Aliases {
			fmt.Printf("  • %s\n", alias)
		}
	}

	// Options
	if len(cmd.Options) > 0 {
		fmt.Printf("\n⚙️  Options:\n")
		for _, opt := range cmd.Options {
			requiredText := ""
			if opt.Required {
				requiredText = " (required)"
			}
			defaultText := ""
			if opt.Default != "" {
				defaultText = fmt.Sprintf(" [default: %s]", opt.Default)
			}
			fmt.Printf("  %-20s %s%s%s\n", opt.Name, opt.Description, requiredText, defaultText)
		}
	}

	// Examples
	if len(cmd.Examples) > 0 {
		fmt.Printf("\n💡 Examples:\n")
		for _, example := range cmd.Examples {
			fmt.Printf("  %s\n", example)
		}
	}

	// Dependencies
	if len(cmd.Dependencies) > 0 {
		fmt.Printf("\n🔗 Dependencies:\n")
		for _, dep := range cmd.Dependencies {
			fmt.Printf("  • %s\n", dep)
		}
	}

	// See also
	if len(cmd.SeeAlso) > 0 {
		fmt.Printf("\n🔍 See Also:\n")
		for _, related := range cmd.SeeAlso {
			fmt.Printf("  • magex %s\n", related)
		}
	}

	// Tags
	if len(cmd.Tags) > 0 {
		fmt.Printf("\n🏷️  Tags: %s\n", strings.Join(cmd.Tags, ", "))
	}

	// Version information
	if cmd.Since != "" {
		fmt.Printf("\n📅 Since: MAGE-X %s\n", cmd.Since)
	}

	// Deprecation warning
	if cmd.Deprecated != "" {
		fmt.Printf("\n⚠️  WARNING: This command is deprecated\n")
		fmt.Printf("   %s\n", cmd.Deprecated)
	}
}

// listCommands displays all available commands
func listCommands(reg *registry.Registry, verbose bool) {
	commands := reg.List()

	if len(commands) == 0 {
		utils.Println("No commands available")
		return
	}

	fmt.Printf("🎯 Available Commands (%d total):\n", len(commands))

	if verbose {
		listCommandsVerbose(commands)
	} else {
		listCommandsSimple(commands)
	}

	utils.Println("\n💡 Run 'magex <command>' to execute a command")
}

// listCommandsVerbose displays commands with descriptions
func listCommandsVerbose(commands []*registry.Command) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, cmd := range commands {
		desc := cmd.Description
		if desc == "" {
			desc = "No description available"
		}
		if cmd.Deprecated != "" {
			desc = fmt.Sprintf("⚠️  DEPRECATED: %s", cmd.Deprecated)
		}
		if _, err := fmt.Fprintf(w, "  %s\t%s\n", cmd.FullName(), desc); err != nil {
			// Print error is non-critical, continue
			_ = err
		}
	}
	if err := w.Flush(); err != nil {
		// Flush error is non-critical, continue
		_ = err
	}
}

// listCommandsSimple displays commands in a simple grid format
func listCommandsSimple(commands []*registry.Command) {
	for i, cmd := range commands {
		fmt.Printf("  %-25s", cmd.FullName())
		if (i+1)%3 == 0 {
			utils.Println("")
		}
	}
	if len(commands)%3 != 0 {
		utils.Println("")
	}
}

// listByNamespace displays commands organized by namespace
func listByNamespace(reg *registry.Registry) {
	utils.Println("Available commands with namespaces:")

	namespaces := reg.Namespaces()
	for _, ns := range namespaces {
		commands := reg.ListByNamespace(ns)
		if len(commands) == 0 {
			continue
		}

		for _, cmd := range commands {
			method := cmd.Method
			if method == "default" || method == "Default" {
				fmt.Printf("  %s:\n", ns)
			} else {
				fmt.Printf("  %s:%s\n", ns, method)
			}
		}
	}
}

// searchCommands searches for commands matching a query with enhanced results
func searchCommands(reg *registry.Registry, query string) {
	matches := reg.Search(query)

	if len(matches) == 0 {
		// Try fuzzy search for similar commands
		allCommands := reg.List()
		var fuzzyMatches []*registry.Command
		for _, cmd := range allCommands {
			if fuzzyMatch(cmd.FullName(), query) {
				fuzzyMatches = append(fuzzyMatches, cmd)
			}
		}

		fmt.Printf("❌ No exact commands found matching '%s'\n", query)
		if len(fuzzyMatches) > 0 {
			fmt.Printf("\n🔍 Did you mean:\n")
			for i, cmd := range fuzzyMatches {
				if i >= 5 {
					break
				}
				fmt.Printf("  • %s - %s\n", cmd.FullName(), cmd.Description)
			}
		}
		return
	}

	fmt.Printf("\n🔍 Search Results for '%s' (%d found):\n", query, len(matches))
	fmt.Printf("%s\n", strings.Repeat("-", 50))

	// Group by category for better organization
	categorized := make(map[string][]*registry.Command)
	for _, cmd := range matches {
		category := cmd.Category
		if category == "" {
			category = "other"
		}
		categorized[category] = append(categorized[category], cmd)
	}

	metadata := reg.Metadata()
	for _, category := range reg.CategoryOrder() {
		commands, exists := categorized[category]
		if !exists || len(commands) == 0 {
			continue
		}

		categoryInfo := metadata.CategoryInfo[category]
		if categoryInfo.Name == "" {
			categoryInfo.Name = cases.Title(language.English).String(category)
		}
		if categoryInfo.Icon == "" {
			categoryInfo.Icon = "📋"
		}

		fmt.Printf("\n%s %s:\n", categoryInfo.Icon, categoryInfo.Name)

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		for _, cmd := range commands {
			desc := cmd.Description
			if desc == "" {
				desc = "No description available"
			}
			// Highlight the matched term
			highlightedName := highlightMatch(cmd.FullName(), query)
			highlightedDesc := highlightMatch(desc, query)
			if _, err := fmt.Fprintf(w, "  %s\t%s\n", highlightedName, highlightedDesc); err != nil {
				// Print error is non-critical, continue
				_ = err
			}
		}
		if err := w.Flush(); err != nil {
			// Flush error is non-critical, continue
			_ = err
		}
	}

	fmt.Printf("\n💡 Tips:\n")
	fmt.Printf("  • Use 'magex -h <command>' for detailed help on any command\n")
	fmt.Printf("  • Use 'magex -n' to browse commands by namespace\n")
	fmt.Printf("  • Try broader search terms for more results\n")
}

// fuzzyMatch performs simple fuzzy matching
func fuzzyMatch(text, pattern string) bool {
	text = strings.ToLower(text)
	pattern = strings.ToLower(pattern)

	// Check for partial matches, transpositions, etc.
	if strings.Contains(text, pattern) {
		return true
	}

	// Simple edit distance check
	return editDistance(text, pattern) <= 2 && len(pattern) > 2
}

// editDistance calculates simple edit distance
func editDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	if s1[0] == s2[0] {
		return editDistance(s1[1:], s2[1:])
	}

	insert := 1 + editDistance(s1, s2[1:])
	deleteOp := 1 + editDistance(s1[1:], s2)
	replace := 1 + editDistance(s1[1:], s2[1:])

	minVal := insert
	if deleteOp < minVal {
		minVal = deleteOp
	}
	if replace < minVal {
		minVal = replace
	}
	return minVal
}

// highlightMatch highlights the matched term (simple approach for CLI)
func highlightMatch(text, pattern string) string {
	pattern = strings.ToLower(pattern)
	lowerText := strings.ToLower(text)
	if idx := strings.Index(lowerText, pattern); idx >= 0 {
		return text[:idx] + "[" + text[idx:idx+len(pattern)] + "]" + text[idx+len(pattern):]
	}
	return text
}

// showQuickList shows a quick list of common commands
func showQuickList(reg *registry.Registry) {
	// Show the most common/useful commands
	commonCommands := []string{
		"build", "test", "lint", "format", "clean",
		"deps", "release", "docker", "help",
	}

	utils.Println("")
	for _, name := range commonCommands {
		if cmd, exists := reg.Get(name); exists {
			fmt.Printf("  %-15s - %s\n", name, truncate(cmd.Description, 50))
		}
	}
}

// initMagefile creates a new magefile with MAGE-X imports
func initMagefile() error {
	content := `//go:build mage
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
`

	filename := "magefile.go"

	// Check if file already exists
	if _, err := os.Stat(filename); err == nil {
		fmt.Printf("❌ Error: %s already exists\n", filename)
		utils.Println("💡 Tip: Remove or rename the existing file first")
		return fmt.Errorf("%w: %s", ErrMagefileExists, filename)
	}

	// Write the file
	if err := os.WriteFile(filename, []byte(content), 0o600); err != nil {
		fmt.Printf("❌ Error creating magefile: %v\n", err)
		return fmt.Errorf("failed to create magefile: %w", err)
	}

	fmt.Printf("✅ Created %s with MAGE-X imports\n", filename)
	utils.Println("🚀 You can now:")
	utils.Println("   • Run 'magex -l' to see all available commands")
	utils.Println("   • Add custom commands to magefile.go")
	utils.Println("   • Run 'magex build' to build your project")
	return nil
}

// cleanCache cleans MAGE-X cache and temporary files
func cleanCache() {
	utils.Header("Cleaning MAGE-X Cache")

	// Clean directories
	dirs := []string{
		".mage",
		".mage-x",
		filepath.Join(os.TempDir(), "magex-plugin-*"),
	}

	for _, dir := range dirs {
		matches, err := filepath.Glob(dir)
		if err != nil {
			fmt.Printf("⚠️  Failed to glob pattern %s: %v\n", dir, err)
			continue
		}
		for _, match := range matches {
			if err := os.RemoveAll(match); err != nil {
				fmt.Printf("⚠️  Failed to remove %s: %v\n", match, err)
			} else {
				fmt.Printf("✅ Removed %s\n", match)
			}
		}
	}

	utils.Success("Cache cleaned")
}

// compileForMage generates a traditional magefile from MAGE-X commands
func compileForMage(output string) {
	fmt.Printf("🔧 Compiling MAGE-X commands to %s...\n", output)

	// This would generate a magefile.go with all MAGE-X commands
	// as wrapper functions that can be used with standard mage

	// For now, we'll create a simple example
	content := `//go:build mage
// +build mage

package main

// This file was auto-generated by MAGE-X
// It provides all MAGE-X commands for use with standard mage

import "github.com/mrz1836/mage-x/pkg/mage"

// Build commands
func Build() error { var b mage.Build; return b.Default() }
func BuildAll() error { var b mage.Build; return b.All() }
func BuildLinux() error { var b mage.Build; return b.Linux() }
func BuildDarwin() error { var b mage.Build; return b.Darwin() }
func BuildWindows() error { var b mage.Build; return b.Windows() }

// Test commands
func Test() error { var t mage.Test; return t.Default() }
func TestUnit() error { var t mage.Test; return t.Unit() }
func TestRace() error { var t mage.Test; return t.Race() }
func TestCover() error { var t mage.Test; return t.Cover() }

// Add more commands as needed...
`

	if err := os.WriteFile(output, []byte(content), 0o600); err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Generated %s\n", output)
	utils.Println("📝 You can now use this file with standard mage")
}

// Helper functions

// normalizeCommandName converts various command formats to our standard
func normalizeCommandName(name string) string {
	// Convert namespace:method to namespace:method (already our format)
	// Convert namespace.method to namespace:method
	// Convert namespace-method to namespace:method

	name = strings.ReplaceAll(name, ".", ":")
	name = strings.ReplaceAll(name, "-", ":")

	return strings.ToLower(name)
}

// truncate truncates a string to a maximum length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
