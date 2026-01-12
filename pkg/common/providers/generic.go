// Package providers contains generic provider implementations for thread-safe singleton patterns.
package providers

import (
	"sync"
	"sync/atomic"
)

// Provider provides a generic thread-safe singleton pattern for any type T.
// It uses sync.Once to ensure thread-safe lazy initialization with a factory function.
type Provider[T any] struct {
	once     sync.Once
	mu       sync.RWMutex
	instance T
	factory  func() T
	wasSet   atomic.Bool // Tracks if instance was manually set via Set()
}

// NewProvider creates a new generic provider with the given factory function.
// The factory function will be called exactly once when Get() is first invoked.
func NewProvider[T any](factory func() T) *Provider[T] {
	return &Provider[T]{
		factory: factory,
	}
}

// Get returns the singleton instance, creating it on first call using the factory function.
// Subsequent calls return the same instance. This method is thread-safe.
// sync.Once provides memory ordering guarantees, making the RWMutex unnecessary.
func (p *Provider[T]) Get() T {
	// If instance was manually set, return it directly
	if p.wasSet.Load() {
		return p.instance
	}
	// Otherwise, lazily initialize with factory
	p.once.Do(func() {
		p.instance = p.factory()
	})
	return p.instance
}

// Reset clears the singleton instance and resets the sync.Once.
// This is primarily useful for testing scenarios where you need to reinitialize the provider.
// This method is now thread-safe.
func (p *Provider[T]) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.once = sync.Once{}
	var zero T
	p.instance = zero
	p.wasSet.Store(false)
}

// Set replaces the singleton instance with the provided value.
// This bypasses the factory function and is primarily useful for dependency injection in tests.
// This method is now thread-safe.
func (p *Provider[T]) Set(instance T) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.instance = instance
	p.wasSet.Store(true)
}

// PackageProvider provides a package-level generic singleton provider with thread-safe management.
// It combines the Provider[T] pattern with additional thread-safe management for package-level access.
type PackageProvider[T any] struct {
	mu       sync.RWMutex
	provider *Provider[T]
	once     sync.Once
	factory  func() T
}

// NewPackageProvider creates a new package-level provider with the given factory function.
func NewPackageProvider[T any](factory func() T) *PackageProvider[T] {
	return &PackageProvider[T]{
		factory: factory,
	}
}

// Get returns the singleton instance with thread-safe lazy initialization of the provider itself.
func (pp *PackageProvider[T]) Get() T {
	pp.mu.RLock()
	if pp.provider != nil {
		provider := pp.provider
		pp.mu.RUnlock()
		return provider.Get()
	}
	pp.mu.RUnlock()

	pp.once.Do(func() {
		pp.mu.Lock()
		defer pp.mu.Unlock()
		if pp.provider == nil {
			pp.provider = NewProvider(pp.factory)
		}
	})

	pp.mu.RLock()
	defer pp.mu.RUnlock()
	return pp.provider.Get()
}

// SetProvider sets a custom provider instance (primarily for testing and dependency injection).
func (pp *PackageProvider[T]) SetProvider(provider *Provider[T]) {
	pp.mu.Lock()
	defer pp.mu.Unlock()
	pp.provider = provider
	// Don't reset once - we're explicitly setting the provider
}

// Set replaces the singleton instance with the provided value (primarily for testing).
func (pp *PackageProvider[T]) Set(instance T) {
	pp.mu.RLock()
	if pp.provider != nil {
		pp.mu.RUnlock()
		pp.provider.Set(instance)
		return
	}
	pp.mu.RUnlock()

	// Create provider if it doesn't exist
	pp.mu.Lock()
	defer pp.mu.Unlock()
	if pp.provider == nil {
		pp.provider = NewProvider(pp.factory)
	}
	pp.provider.Set(instance)
}

// Reset resets the provider and its instance (primarily for testing).
func (pp *PackageProvider[T]) Reset() {
	pp.mu.Lock()
	defer pp.mu.Unlock()
	pp.once = sync.Once{}
	pp.provider = nil
}
