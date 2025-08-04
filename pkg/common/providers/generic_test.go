package providers

import (
	"sync"
	"testing"
	"time"
)

// Test constants
const (
	testValue      = "test-value"
	packageValue   = "package-value"
	benchmarkValue = "benchmark-value"
)

// TestProvider tests the basic Provider functionality
func TestProvider(t *testing.T) {
	callCount := 0
	factory := func() string {
		callCount++
		return testValue
	}

	p := NewProvider(factory)

	// First call should invoke factory
	result1 := p.Get()
	if result1 != testValue {
		t.Errorf("Expected 'test-value', got '%s'", result1)
	}
	if callCount != 1 {
		t.Errorf("Expected factory to be called once, got %d calls", callCount)
	}

	// Second call should return same instance without calling factory
	result2 := p.Get()
	if result2 != testValue {
		t.Errorf("Expected 'test-value', got '%s'", result2)
	}
	if callCount != 1 {
		t.Errorf("Expected factory to still be called once, got %d calls", callCount)
	}
}

// TestProviderConcurrency tests thread safety of Provider
func TestProviderConcurrency(t *testing.T) {
	callCount := 0
	var mu sync.Mutex
	factory := func() int {
		mu.Lock()
		defer mu.Unlock()
		callCount++
		return callCount
	}

	p := NewProvider(factory)

	// Start multiple goroutines
	const numGoroutines = 10
	results := make(chan int, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			results <- p.Get()
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		result := <-results
		if result != 1 {
			t.Errorf("Expected 1, got %d", result)
		}
	}

	// Factory should have been called exactly once
	mu.Lock()
	finalCallCount := callCount
	mu.Unlock()

	if finalCallCount != 1 {
		t.Errorf("Expected factory to be called once, got %d calls", finalCallCount)
	}
}

// TestProviderReset tests the Reset functionality
func TestProviderReset(t *testing.T) {
	callCount := 0
	factory := func() string {
		callCount++
		return testValue
	}

	p := NewProvider(factory)

	// Get initial value
	result1 := p.Get()
	if result1 != testValue {
		t.Errorf("Expected 'test-value', got '%s'", result1)
	}
	if callCount != 1 {
		t.Errorf("Expected factory to be called once, got %d calls", callCount)
	}

	// Reset and get again
	p.Reset()
	result2 := p.Get()
	if result2 != testValue {
		t.Errorf("Expected 'test-value', got '%s'", result2)
	}
	if callCount != 2 {
		t.Errorf("Expected factory to be called twice after reset, got %d calls", callCount)
	}
}

// TestProviderSet tests the Set functionality
func TestProviderSet(t *testing.T) {
	callCount := 0
	factory := func() string {
		callCount++
		return "factory-value"
	}

	p := NewProvider(factory)

	// Set custom value
	p.Set("custom-value")

	// Get should return custom value without calling factory
	result := p.Get()
	if result != "custom-value" {
		t.Errorf("Expected 'custom-value', got '%s'", result)
	}
	if callCount != 0 {
		t.Errorf("Expected factory to not be called, got %d calls", callCount)
	}
}

// TestPackageProvider tests the basic PackageProvider functionality
func TestPackageProvider(t *testing.T) {
	callCount := 0
	factory := func() string {
		callCount++
		return packageValue
	}

	pp := NewPackageProvider(factory)

	// First call should invoke factory
	result1 := pp.Get()
	if result1 != packageValue {
		t.Errorf("Expected 'package-value', got '%s'", result1)
	}
	if callCount != 1 {
		t.Errorf("Expected factory to be called once, got %d calls", callCount)
	}

	// Second call should return same instance
	result2 := pp.Get()
	if result2 != packageValue {
		t.Errorf("Expected 'package-value', got '%s'", result2)
	}
	if callCount != 1 {
		t.Errorf("Expected factory to still be called once, got %d calls", callCount)
	}
}

// TestPackageProviderConcurrency tests thread safety of PackageProvider
func TestPackageProviderConcurrency(t *testing.T) {
	callCount := 0
	var mu sync.Mutex
	factory := func() int {
		mu.Lock()
		defer mu.Unlock()
		callCount++
		return callCount
	}

	pp := NewPackageProvider(factory)

	// Start multiple goroutines
	const numGoroutines = 20
	results := make(chan int, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			results <- pp.Get()
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		result := <-results
		if result != 1 {
			t.Errorf("Expected 1, got %d", result)
		}
	}

	// Factory should have been called exactly once
	mu.Lock()
	finalCallCount := callCount
	mu.Unlock()

	if finalCallCount != 1 {
		t.Errorf("Expected factory to be called once, got %d calls", finalCallCount)
	}
}

// TestPackageProviderSet tests the Set functionality
func TestPackageProviderSet(t *testing.T) {
	callCount := 0
	factory := func() string {
		callCount++
		return "factory-value"
	}

	pp := NewPackageProvider(factory)

	// Set custom value
	pp.Set("custom-value")

	// Get should return custom value
	result := pp.Get()
	if result != "custom-value" {
		t.Errorf("Expected 'custom-value', got '%s'", result)
	}
	if callCount != 0 {
		t.Errorf("Expected factory to not be called, got %d calls", callCount)
	}
}

// TestPackageProviderSetProvider tests the SetProvider functionality
func TestPackageProviderSetProvider(t *testing.T) {
	callCount1 := 0
	factory1 := func() string {
		callCount1++
		return "factory1-value"
	}

	callCount2 := 0
	factory2 := func() string {
		callCount2++
		return "factory2-value"
	}

	pp := NewPackageProvider(factory1)
	customProvider := NewProvider(factory2)

	// Set custom provider
	pp.SetProvider(customProvider)

	// Get should use custom provider
	result := pp.Get()
	if result != "factory2-value" {
		t.Errorf("Expected 'factory2-value', got '%s'", result)
	}
	if callCount1 != 0 {
		t.Errorf("Expected original factory to not be called, got %d calls", callCount1)
	}
	if callCount2 != 1 {
		t.Errorf("Expected custom factory to be called once, got %d calls", callCount2)
	}
}

// TestPackageProviderReset tests the Reset functionality
func TestPackageProviderReset(t *testing.T) {
	callCount := 0
	factory := func() string {
		callCount++
		return packageValue
	}

	pp := NewPackageProvider(factory)

	// Get initial value
	result1 := pp.Get()
	if result1 != packageValue {
		t.Errorf("Expected 'package-value', got '%s'", result1)
	}
	if callCount != 1 {
		t.Errorf("Expected factory to be called once, got %d calls", callCount)
	}

	// Reset and get again
	pp.Reset()
	result2 := pp.Get()
	if result2 != packageValue {
		t.Errorf("Expected 'package-value', got '%s'", result2)
	}
	if callCount != 2 {
		t.Errorf("Expected factory to be called twice after reset, got %d calls", callCount)
	}
}

// TestProviderWithStruct tests using Provider with struct types
func TestProviderWithStruct(t *testing.T) {
	type Config struct {
		Name    string
		Timeout time.Duration
	}

	callCount := 0
	factory := func() *Config {
		callCount++
		return &Config{
			Name:    "test-config",
			Timeout: 30 * time.Second,
		}
	}

	p := NewProvider(factory)

	// First call
	config1 := p.Get()
	if config1.Name != "test-config" {
		t.Errorf("Expected name 'test-config', got '%s'", config1.Name)
	}
	if config1.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", config1.Timeout)
	}
	if callCount != 1 {
		t.Errorf("Expected factory to be called once, got %d calls", callCount)
	}

	// Second call should return same instance
	config2 := p.Get()
	if config1 != config2 {
		t.Error("Expected same instance to be returned")
	}
	if callCount != 1 {
		t.Errorf("Expected factory to still be called once, got %d calls", callCount)
	}
}

// BenchmarkProvider benchmarks the Provider performance
func BenchmarkProvider(b *testing.B) {
	factory := func() string {
		return benchmarkValue
	}

	p := NewProvider(factory)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = p.Get()
		}
	})
}

// BenchmarkPackageProvider benchmarks the PackageProvider performance
func BenchmarkPackageProvider(b *testing.B) {
	factory := func() string {
		return benchmarkValue
	}

	pp := NewPackageProvider(factory)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = pp.Get()
		}
	})
}
