// Package mage provides the core command implementations for MAGE-X,
// a universal build automation tool for Go projects.
//
// # Namespaces
//
// Commands are organized into namespaces:
//   - Build: Compilation and packaging (Build.Default, Build.All, Build.Platform)
//   - Test: Testing and coverage (Test.Default, Test.Coverage, Test.Race)
//   - Lint: Code quality (Lint.Default, Lint.Fix, Lint.Vet)
//   - Format: Code formatting (Format.Default, Format.Check)
//   - Deps: Dependency management (Deps.Update, Deps.Tidy, Deps.Audit)
//   - Version: Version management (Version.Bump, Version.Tag)
//   - Docs: Documentation generation
//   - Release: Release automation
//
// # Configuration
//
// Commands read configuration from .mage.yaml with environment
// variable overrides via MAGE_X_* prefix. Example configuration:
//
//	project:
//	  name: myapp
//	  binary: myapp
//	build:
//	  output: ./bin
//	  platforms:
//	    - linux/amd64
//	    - darwin/arm64
//
// # Registry
//
// All commands are registered in the global registry and can be
// discovered and executed by the magex CLI:
//
//	magex build:default    # Run default build
//	magex test:coverage    # Run tests with coverage
//	magex lint:fix         # Run linter with auto-fix
//
// # Extending
//
// Custom commands can be added via magefile.go and will be
// automatically discovered and integrated with built-in commands.
package mage
