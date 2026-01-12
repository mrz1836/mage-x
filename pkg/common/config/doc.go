// Package config provides a flexible configuration management system with
// multi-source loading, validation, and change watching capabilities.
//
// # Architecture
//
// The configuration system is built around these core interfaces:
//
//   - Source: Represents a configuration source (file, environment, etc.)
//   - Manager: Coordinates loading from multiple sources
//   - Validator: Validates configuration data
//
// # Multi-Source Loading
//
// Configuration can be loaded from multiple sources with priority-based selection:
//
//	manager := config.NewDefaultConfigManager()
//	manager.AddSource(fileSource)     // Priority 100
//	manager.AddSource(envSource)      // Priority 200 (higher = loaded first)
//	err := manager.LoadConfig(&cfg)
//
// # Configuration Watching
//
// The manager supports watching for configuration changes:
//
//	err := manager.Watch(func(newConfig interface{}) {
//	    // Handle configuration update
//	})
//	defer manager.StopWatching()
//
// # Validation
//
// Configuration can be validated using the Validator interface:
//
//	validator := config.NewBasicValidator()
//	manager.SetValidator(validator)
//
// # Thread Safety
//
// The DefaultConfigManager is thread-safe and can be used concurrently from
// multiple goroutines. All operations are protected by appropriate mutex locks.
//
// # Environment Variables
//
// The EnvProvider and TypedEnvProvider interfaces support loading configuration
// from environment variables with type conversion:
//
//   - GetBool: Parse boolean values
//   - GetInt: Parse integer values
//   - GetDuration: Parse duration strings
//   - GetStringSlice: Parse comma-separated strings
package config
