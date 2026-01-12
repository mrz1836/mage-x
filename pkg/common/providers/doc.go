// Package providers contains generic provider implementations for thread-safe
// singleton patterns with support for dependency injection and testing.
//
// # Provider[T]
//
// Generic thread-safe singleton pattern for any type:
//
//	// Create provider with factory function
//	provider := providers.NewProvider(func() *MyService {
//	    return &MyService{config: loadConfig()}
//	})
//
//	// Get singleton instance (thread-safe, lazy initialization)
//	service := provider.Get()
//
// # Dependency Injection
//
// Replace instances for testing:
//
//	// In test setup
//	provider.Set(mockService)
//	defer provider.Reset()
//
// # PackageProvider[T]
//
// Package-level singleton with additional thread-safe management:
//
//	var configProvider = providers.NewPackageProvider(func() *Config {
//	    return loadConfig()
//	})
//
//	func GetConfig() *Config {
//	    return configProvider.Get()
//	}
//
// # Thread Safety
//
// Both Provider and PackageProvider are fully thread-safe:
//
//   - Get() uses sync.Once for lazy initialization
//   - Set() and Reset() use mutex protection
//   - Safe for concurrent access from multiple goroutines
//
// # Testing Patterns
//
// Common testing patterns:
//
//	func TestWithMock(t *testing.T) {
//	    // Save and restore original
//	    original := provider.Get()
//	    defer provider.Set(original)
//
//	    // Use mock for test
//	    provider.Set(mock)
//	    // Run test...
//	}
package providers
