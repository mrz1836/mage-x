// Package log provides logging infrastructure for the MAGE-X build system.
//
// # Logger Types
//
// Two logger implementations are available:
//   - CLILogger: Colored terminal output with spinners and progress
//   - StructuredLogger: Structured logging with fields and context
//
// # Usage
//
// Use package-level functions for convenience:
//
//	log.Info("Building %s", target)
//	log.Error("Build failed: %v", err)
//	log.Success("Build completed in %v", duration)
//
// Or create a logger with fields:
//
//	logger := log.Default().WithFields(log.Fields{
//	    "module": moduleName,
//	    "target": target,
//	})
//	logger.Info("Starting build")
//
// # Log Levels
//
// Supported levels: Debug, Info, Warn, Error, Fatal
//
// Set the log level programmatically:
//
//	log.SetLevel(log.LevelDebug)
//
// Or via environment variable:
//
//	MAGE_X_LOG_LEVEL=debug magex build
//
// # CLI Features
//
// The CLILogger provides additional features for terminal output:
//   - Colored output (respects NO_COLOR and CI environment variables)
//   - Progress spinners for long-running operations
//   - Success/Fail status indicators
//   - Header formatting for section organization
package log
