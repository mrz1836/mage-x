// Package registry provides the command registration system for mage-x,
// managing the collection of available build commands and their metadata.
//
// # Global Registry
//
// Access the global command registry:
//
//	reg := registry.Global()
//	cmd := reg.GetCommand("build:default")
//
// # Command Registration
//
// Commands are registered with full metadata:
//
//	reg.Register(&registry.Command{
//	    Name:        "build:default",
//	    Description: "Build the application",
//	    Category:    "build",
//	    Aliases:     []string{"build", "b"},
//	    Examples:    []string{"magex build:default"},
//	    Handler:     Build{}.Default,
//	})
//
// # Command Discovery
//
// Find commands by various criteria:
//
//	// Get all commands in a category
//	buildCmds := reg.GetByCategory("build")
//
//	// Search by prefix
//	matches := reg.Search("build:")
//
//	// Get by alias
//	cmd := reg.GetByAlias("b")
//
// # Thread Safety
//
// The global registry uses sync.Once for initialization and is safe for
// concurrent read access. Command registration should be done at startup.
//
// # Integration with embed
//
// The registry works with the embed package which provides declarative
// command registration using a builder pattern.
package registry
