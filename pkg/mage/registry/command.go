// Package registry provides a command registry system for MAGE-X
// This enables both library and binary modes to share command definitions
package registry

import (
	"errors"
	"fmt"
	"strings"
)

// Static errors
var (
	ErrInvalidCommand = errors.New("command must have either Name or Namespace+Method")
	ErrNoFunction     = errors.New("command has no executable function")
)

// CommandFunc is the function signature for all commands
type CommandFunc func() error

// CommandWithArgsFunc is the function signature for commands that accept arguments
type CommandWithArgsFunc func(args ...string) error

// Command represents a single executable command in MAGE-X
type Command struct {
	// Name is the full command name (e.g., "build:linux")
	Name string

	// Namespace is the command namespace (e.g., "build")
	Namespace string

	// Method is the method name within the namespace (e.g., "linux")
	Method string

	// Description provides help text for the command
	Description string

	// LongDescription provides detailed help text for the command
	LongDescription string

	// Usage shows the command usage pattern
	Usage string

	// Examples provides usage examples
	Examples []string

	// Options describes available options/environment variables
	Options []CommandOption

	// SeeAlso lists related commands
	SeeAlso []string

	// Tags for categorizing and searching commands
	Tags []string

	// Aliases are alternative names for the command
	Aliases []string

	// Func is the actual function to execute
	Func CommandFunc

	// FuncWithArgs is the function that accepts arguments (if applicable)
	FuncWithArgs CommandWithArgsFunc

	// Hidden indicates if this command should be hidden from help
	Hidden bool

	// Deprecated marks a command as deprecated with a replacement suggestion
	Deprecated string

	// Dependencies are commands that must run before this one
	Dependencies []string

	// Category groups related commands in help output
	Category string

	// Since indicates the MAGE-X version that introduced this command
	Since string

	// Icon is the emoji icon for this command
	Icon string
}

// FullName returns the complete command name including namespace
func (c *Command) FullName() string {
	if c.Namespace != "" && c.Method != "" {
		return fmt.Sprintf("%s:%s", strings.ToLower(c.Namespace), strings.ToLower(c.Method))
	}
	return strings.ToLower(c.Name)
}

// IsNamespace returns true if this is a namespace default command
func (c *Command) IsNamespace() bool {
	return c.Method == "default" || c.Method == ""
}

// Validate ensures the command is properly configured
func (c *Command) Validate() error {
	if c.Name == "" && (c.Namespace == "" || c.Method == "") {
		return ErrInvalidCommand
	}
	if c.Func == nil && c.FuncWithArgs == nil {
		return fmt.Errorf("%w: %s", ErrNoFunction, c.FullName())
	}
	return nil
}

// Execute runs the command with optional arguments
func (c *Command) Execute(args ...string) error {
	// Debug logging
	fmt.Printf("🐛 DEBUG [Command.Execute]: Command %s, args: %v, FuncWithArgs: %v, Func: %v\n",
		c.FullName(), args, c.FuncWithArgs != nil, c.Func != nil)

	// Check if deprecated
	if c.Deprecated != "" {
		fmt.Printf("⚠️  Warning: '%s' is deprecated. %s\n", c.FullName(), c.Deprecated)
	}

	// If arguments are provided and FuncWithArgs exists, use it
	if len(args) > 0 && c.FuncWithArgs != nil {
		fmt.Printf("🐛 DEBUG [Command.Execute]: Using FuncWithArgs with %d args\n", len(args))
		return c.FuncWithArgs(args...)
	}

	// If no arguments provided, prefer Func over FuncWithArgs
	if len(args) == 0 && c.Func != nil {
		return c.Func()
	}

	// If we have args but no FuncWithArgs, fallback to Func if it exists
	if len(args) > 0 && c.FuncWithArgs == nil && c.Func != nil {
		return c.Func()
	}

	// If we have FuncWithArgs but no args were provided and no Func, use FuncWithArgs
	if c.FuncWithArgs != nil {
		return c.FuncWithArgs(args...)
	}

	return fmt.Errorf("%w: %s", ErrNoFunction, c.FullName())
}

// CommandOption represents a command option or environment variable
type CommandOption struct {
	// Name is the option name (e.g., "VERBOSE", "--timeout")
	Name string

	// Description explains what the option does
	Description string

	// Default is the default value
	Default string

	// Required indicates if this option is required
	Required bool

	// Type indicates the option type (bool, string, int, duration)
	Type string
}

// CommandMetadata provides additional information about commands
type CommandMetadata struct {
	// TotalCommands is the total number of registered commands
	TotalCommands int

	// Namespaces is the list of all namespaces
	Namespaces []string

	// Categories maps category names to command counts
	Categories map[string]int

	// Version is the MAGE-X version
	Version string

	// CategoryInfo provides category metadata
	CategoryInfo map[string]CategoryInfo
}

// CategoryInfo provides metadata about a command category
type CategoryInfo struct {
	// Name is the category name
	Name string

	// Description describes the category
	Description string

	// Icon is the emoji icon for the category
	Icon string

	// Order is the display order
	Order int
}

// CommandBuilder provides a fluent interface for building commands
type CommandBuilder struct {
	cmd *Command
}

// NewCommand creates a new command builder
func NewCommand(name string) *CommandBuilder {
	return &CommandBuilder{
		cmd: &Command{
			Name: name,
		},
	}
}

// NewNamespaceCommand creates a new namespace-based command
func NewNamespaceCommand(namespace, method string) *CommandBuilder {
	return &CommandBuilder{
		cmd: &Command{
			Namespace: namespace,
			Method:    method,
		},
	}
}

// WithDescription sets the command description
func (b *CommandBuilder) WithDescription(desc string) *CommandBuilder {
	b.cmd.Description = desc
	return b
}

// WithFunc sets the command function
func (b *CommandBuilder) WithFunc(fn CommandFunc) *CommandBuilder {
	b.cmd.Func = fn
	return b
}

// WithArgsFunc sets the command function that accepts arguments
func (b *CommandBuilder) WithArgsFunc(fn CommandWithArgsFunc) *CommandBuilder {
	b.cmd.FuncWithArgs = fn
	return b
}

// WithAliases sets command aliases
func (b *CommandBuilder) WithAliases(aliases ...string) *CommandBuilder {
	b.cmd.Aliases = aliases
	return b
}

// WithCategory sets the command category
func (b *CommandBuilder) WithCategory(category string) *CommandBuilder {
	b.cmd.Category = category
	return b
}

// WithLongDescription sets the detailed command description
func (b *CommandBuilder) WithLongDescription(desc string) *CommandBuilder {
	b.cmd.LongDescription = desc
	return b
}

// WithUsage sets the command usage pattern
func (b *CommandBuilder) WithUsage(usage string) *CommandBuilder {
	b.cmd.Usage = usage
	return b
}

// WithExamples sets command examples
func (b *CommandBuilder) WithExamples(examples ...string) *CommandBuilder {
	b.cmd.Examples = examples
	return b
}

// WithOptions sets command options
func (b *CommandBuilder) WithOptions(options ...CommandOption) *CommandBuilder {
	b.cmd.Options = options
	return b
}

// WithSeeAlso sets related commands
func (b *CommandBuilder) WithSeeAlso(commands ...string) *CommandBuilder {
	b.cmd.SeeAlso = commands
	return b
}

// WithTags sets command tags
func (b *CommandBuilder) WithTags(tags ...string) *CommandBuilder {
	b.cmd.Tags = tags
	return b
}

// WithIcon sets the command icon
func (b *CommandBuilder) WithIcon(icon string) *CommandBuilder {
	b.cmd.Icon = icon
	return b
}

// WithDependencies sets command dependencies
func (b *CommandBuilder) WithDependencies(deps ...string) *CommandBuilder {
	b.cmd.Dependencies = deps
	return b
}

// Hidden marks the command as hidden
func (b *CommandBuilder) Hidden() *CommandBuilder {
	b.cmd.Hidden = true
	return b
}

// Deprecated marks the command as deprecated
func (b *CommandBuilder) Deprecated(message string) *CommandBuilder {
	b.cmd.Deprecated = message
	return b
}

// Since sets the version when the command was introduced
func (b *CommandBuilder) Since(version string) *CommandBuilder {
	b.cmd.Since = version
	return b
}

// Build validates and returns the command
func (b *CommandBuilder) Build() (*Command, error) {
	if err := b.cmd.Validate(); err != nil {
		return nil, err
	}
	return b.cmd, nil
}

// MustBuild validates and returns the command, panicking on error
func (b *CommandBuilder) MustBuild() *Command {
	cmd, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build command: %v", err))
	}
	return cmd
}
