// Package utils provides general-purpose utility functions for the mage-x build system.
//
// This package includes utilities for:
//   - File downloads with retry logic and checksum verification
//   - HTTP client management with connection pooling
//   - Logging with emoji support and color detection
//   - Command execution helpers
//   - File and directory operations
//   - Progress indicators and spinners
//
// # Download Operations
//
// The download utilities support:
//
//   - Automatic retry with exponential backoff
//   - Resume capability for interrupted downloads
//   - SHA256 checksum verification
//   - Progress callbacks for UI feedback
//
// Example:
//
//	config := utils.DefaultDownloadConfig()
//	config.ChecksumSHA256 = "abc123..."
//	err := utils.DownloadWithRetry(ctx, url, destPath, config)
//
// # HTTP Client
//
// A shared HTTP client with connection pooling is provided for improved performance:
//
//	client := utils.DefaultHTTPClient()
//
// # Logging
//
// The package provides a CLI-optimized logger with:
//
//   - Colored output (auto-detected, respects NO_COLOR)
//   - Emoji prefixes for status messages
//   - Debug, Info, Warn, Error log levels
//
// # Security
//
// Script downloads can be secured with checksum enforcement:
//
//	export MAGE_X_REQUIRE_CHECKSUMS=true
//
// When enabled, script downloads without checksums will fail, preventing
// potential supply chain attacks in CI/CD environments.
package utils
