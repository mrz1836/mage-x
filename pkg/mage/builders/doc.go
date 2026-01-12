// Package builders provides command builders for constructing arguments
// to external tools like golangci-lint and go test.
//
// # Overview
//
// Builders encapsulate the logic for constructing command-line arguments
// based on configuration, ensuring consistent tool invocation across
// the codebase.
//
// # Lint Command Builder
//
// Build golangci-lint arguments:
//
//	builder := builders.NewLintCommandBuilder(config)
//	args := builder.BuildGolangciArgs(module, builders.LintOptions{
//	    Fix:    true,
//	    Format: "colored-line-number",
//	})
//	// args: ["run", "--config", ".golangci.yml", "--fix", ...]
//
// # Test Command Builder
//
// Build go test arguments:
//
//	builder := builders.NewTestCommandBuilder(config)
//	args := builder.BuildTestArgs(builders.TestOptions{
//	    Race:    true,
//	    Verbose: true,
//	    Cover:   true,
//	})
//	// args: ["-race", "-v", "-cover", ...]
//
// # Configuration Integration
//
// Builders use the Config interface to access settings:
//
//	type Config interface {
//	    GetLint() LintConfig
//	    GetBuild() BuildConfig
//	}
//
// # Lint Options
//
// Available lint options:
//
//	type LintOptions struct {
//	    Fix      bool   // Enable auto-fix
//	    Format   string // Output format
//	    Output   string // Output file path
//	    NoConfig bool   // Skip config file lookup
//	}
//
// # Config File Discovery
//
// The lint builder automatically searches for configuration files:
//
//   - .golangci.yml
//   - .golangci.yaml
//   - .golangci.toml
//   - .golangci.json
package builders
