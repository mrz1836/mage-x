// Package env provides environment variable handling utilities for mage-x,
// including loading from .env files and type-safe value retrieval.
//
// # Loading Environment Files
//
// Load environment variables from .env files at startup:
//
//	if err := env.LoadStartupEnv(); err != nil {
//	    log.Printf("Could not load env files: %v", err)
//	}
//
// # Retrieving Values
//
// Type-safe accessors for environment variables:
//
//	debug := env.GetBool("DEBUG", false)
//	port := env.GetInt("PORT", 8080)
//	timeout := env.GetDuration("TIMEOUT", 30*time.Second)
//	tags := env.GetString("BUILD_TAGS", "")
//
// # Environment Variables
//
// Common mage-x environment variables:
//
//   - MAGE_X_BUILD_TAGS: Build tags to pass to go build
//   - MAGE_X_CACHE_DISABLED: Disable build caching
//   - MAGE_X_REQUIRE_CHECKSUMS: Require checksums for script downloads
//   - DEBUG: Enable debug mode (disables binary stripping)
//   - VERBOSE: Enable verbose output
//
// # .env File Format
//
// The .env file format supports:
//
//	KEY=value
//	QUOTED="value with spaces"
//	# Comments are ignored
//
// Files are searched in order: .env, .env.local
package env
