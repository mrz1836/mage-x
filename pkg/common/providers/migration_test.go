package providers

import (
	"testing"
	"time"

	"github.com/mrz1836/mage-x/pkg/common/cache"
)

// TestProviderMigrationCompatibility tests that the migrated providers maintain backward compatibility
func TestProviderMigrationCompatibility(t *testing.T) {
	// Test that we can create providers using the same patterns as before

	// Test cache manager provider pattern
	cacheManagerProvider := NewProvider(func() *cache.Manager {
		config := cache.DefaultConfig()
		return cache.NewManager(config)
	})

	// First call should create instance
	manager1 := cacheManagerProvider.Get()
	if manager1 == nil {
		t.Error("Expected cache manager to be created")
	}

	// Second call should return same instance
	manager2 := cacheManagerProvider.Get()
	if manager1 != manager2 {
		t.Error("Expected same cache manager instance")
	}
}

// TestPackageProviderMigrationPattern tests the package-level provider pattern
func TestPackageProviderMigrationPattern(t *testing.T) {
	type MockService struct {
		name string
		id   int
	}

	// Simulate the old pattern with package-level access
	packageProvider := NewPackageProvider(func() *MockService {
		return &MockService{name: "default", id: 1}
	})

	// Test getting service
	service1 := packageProvider.Get()
	if service1.name != "default" || service1.id != 1 {
		t.Errorf("Expected default service, got %+v", service1)
	}

	// Test setting custom service
	customService := &MockService{name: "custom", id: 2}
	packageProvider.Set(customService)

	service2 := packageProvider.Get()
	if service2 != customService {
		t.Error("Expected custom service to be returned")
	}
	if service2.name != "custom" || service2.id != 2 {
		t.Errorf("Expected custom service values, got %+v", service2)
	}

	// Test reset
	packageProvider.Reset()
	service3 := packageProvider.Get()
	if service3 == customService {
		t.Error("Expected new service instance after reset")
	}
	if service3.name != "default" || service3.id != 1 {
		t.Errorf("Expected default service after reset, got %+v", service3)
	}
}

// TestProviderFactoryFlexibility tests that providers work with different factory functions
func TestProviderFactoryFlexibility(t *testing.T) {
	// Test with simple value factory
	simpleProvider := NewProvider(func() string {
		return "simple-value"
	})

	if simpleProvider.Get() != "simple-value" {
		t.Error("Simple provider failed")
	}

	// Test with complex initialization factory
	complexProvider := NewProvider(func() map[string]interface{} {
		config := make(map[string]interface{})
		config["timeout"] = 30 * time.Second
		config["retries"] = 3
		config["enabled"] = true
		return config
	})

	complexResult := complexProvider.Get()
	if complexResult["timeout"] != 30*time.Second {
		t.Error("Complex provider initialization failed")
	}
	if complexResult["retries"] != 3 {
		t.Error("Complex provider initialization failed")
	}
	if complexResult["enabled"] != true {
		t.Error("Complex provider initialization failed")
	}
}

// TestProviderErrorHandling tests provider behavior with factory functions that might have issues
func TestProviderErrorHandling(t *testing.T) {
	// Test with factory that returns nil (valid for pointer types)
	nilProvider := NewProvider(func() *string {
		return nil
	})

	result := nilProvider.Get()
	if result != nil {
		t.Error("Expected nil result from nil factory")
	}

	// Test with factory that panics (should be handled by calling code)
	panicProvider := NewProvider(func() string {
		panic("test panic")
	})

	// This should panic when called, which is expected behavior
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic from panic factory")
		}
	}()

	_ = panicProvider.Get()
}

// BenchmarkProviderVsDirectAccess compares performance of Provider vs direct access
func BenchmarkProviderVsDirectAccess(b *testing.B) {
	// Setup direct access
	directValue := "benchmark-value"

	// Setup provider
	provider := NewProvider(func() string {
		return "benchmark-value"
	})

	// Pre-initialize provider to ensure fair comparison
	_ = provider.Get()

	b.Run("DirectAccess", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = directValue
		}
	})

	b.Run("ProviderAccess", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = provider.Get()
		}
	})
}

// BenchmarkPackageProviderOverhead compares PackageProvider vs regular Provider
func BenchmarkPackageProviderOverhead(b *testing.B) {
	regularProvider := NewProvider(func() string {
		return "benchmark-value"
	})

	packageProvider := NewPackageProvider(func() string {
		return "benchmark-value"
	})

	// Pre-initialize both
	_ = regularProvider.Get()
	_ = packageProvider.Get()

	b.Run("RegularProvider", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = regularProvider.Get()
		}
	})

	b.Run("PackageProvider", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = packageProvider.Get()
		}
	})
}
