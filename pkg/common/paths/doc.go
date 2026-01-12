// Package paths provides advanced path building and manipulation utilities
// with mockable interfaces for testing and dependency injection.
//
// # Core Components
//
// The package provides several key interfaces:
//
//   - PathBuilder: Constructs and manipulates file paths
//   - PathMatcher: Pattern matching for file paths (glob, regex)
//   - PathValidator: Validates paths for security and correctness
//   - PathSet: Manages collections of paths efficiently
//   - PathWatcher: Monitors file system changes
//   - PathCache: Caches path resolution results
//
// # Default Instances
//
// Thread-safe singleton access to default implementations:
//
//	builder := paths.GetDefaultBuilder()
//	matcher := paths.GetDefaultMatcher()
//	validator := paths.GetDefaultValidator()
//
// # Path Building
//
// Construct paths safely with the builder:
//
//	path := paths.GetDefaultBuilder().
//	    Join("src", "pkg", "main.go").
//	    Abs()
//
// # Pattern Matching
//
// Match paths against patterns:
//
//	matcher := paths.GetDefaultMatcher()
//	matches := matcher.Match("**/*.go", files)
//	matched := matcher.MatchOne("*.go", "main.go")
//
// # Security Validation
//
// Validate paths to prevent directory traversal attacks:
//
//	validator := paths.GetDefaultValidator()
//	if err := validator.Validate(userPath); err != nil {
//	    // Handle invalid path
//	}
//
// # Thread Safety
//
// All default instances use sync.Once for thread-safe initialization.
// The package is safe for concurrent use from multiple goroutines.
//
// # Testing
//
// The mockable interfaces support dependency injection in tests:
//
//	mockBuilder := mocks.NewMockPathBuilder(ctrl)
//	// Configure mock expectations
package paths
