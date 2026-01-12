// Package cache provides build caching capabilities for mage-x,
// enabling faster incremental builds by reusing previous build artifacts.
//
// # Overview
//
// The cache system stores build artifacts keyed by a hash of:
//   - Source file contents
//   - Build configuration (ldflags, tags, etc.)
//   - Target platform (GOOS/GOARCH)
//
// # Manager
//
// The cache Manager coordinates all caching operations:
//
//	config := cache.DefaultConfig()
//	manager := cache.NewManager(config)
//	if err := manager.Init(); err != nil {
//	    log.Printf("Cache initialization failed: %v", err)
//	}
//
// # Hash Generation
//
// Build hashes are generated to identify unique build configurations:
//
//	hash, err := manager.GenerateBuildHash(platform, ldflags, sourceFiles, configFiles)
//	if cached := manager.GetCached(hash); cached != nil {
//	    // Use cached artifact
//	}
//
// # Configuration
//
// Cache behavior is controlled via environment variables:
//
//   - MAGE_X_CACHE_DISABLED=true: Disable caching entirely
//
// # Thread Safety
//
// The cache Manager is thread-safe and can be accessed concurrently from
// multiple goroutines during parallel builds.
//
// # Storage
//
// Cached artifacts are stored in a platform-specific cache directory,
// typically under ~/.cache/mage-x or equivalent.
package cache
