// Package embed provides registration of all built-in MAGE-X commands,
// forming the heart of the zero-boilerplate solution.
//
// # Overview
//
// This package registers all 174 pre-built commands with the global registry,
// making them immediately available without manual registration. Commands are
// defined declaratively using data tables and registered at startup.
//
// # Command Registration
//
// Commands are registered using a declarative pattern:
//
//	registerNamespaceCommands(reg, "build", "build", getBuildCommands(), buildBindings)
//
// # Command Definitions
//
// Commands are defined using CommandDef:
//
//	type CommandDef struct {
//	    Method   string   // e.g., "default", "all", "linux"
//	    Desc     string   // Short description
//	    Aliases  []string // Optional aliases
//	    Usage    string   // Optional usage pattern
//	    Examples []string // Optional examples
//	}
//
// # Method Bindings
//
// Methods are bound to actual functions:
//
//	type MethodBinding struct {
//	    NoArgs   func() error          // For commands without arguments
//	    WithArgs func(...string) error // For commands accepting arguments
//	}
//
// # Registered Namespaces
//
// The package registers commands for all namespaces:
//
//   - build: Build commands (default, all, linux, darwin, windows, clean)
//   - test: Test commands (default, unit, full, coverage, race, fuzz)
//   - lint: Lint commands (default, fix, check, format)
//   - format: Format commands (default, check, imports)
//   - deps: Dependency commands (default, update, tidy, vendor)
//   - git: Git commands (status, commit, push, tag)
//   - release: Release commands (default, major, minor, patch)
//   - And many more...
//
// # Integration
//
// The embed package is automatically imported by the magex binary,
// ensuring all commands are available at startup.
package embed
