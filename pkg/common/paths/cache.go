// Package paths provides advanced path caching capabilities
package paths

import (
	"container/list"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Error definitions for cache operations
var (
	ErrKeyNotFound           = errors.New("key not found")
	ErrCacheKeyEmpty         = errors.New("cache key cannot be empty")
	ErrCacheKeyTooLong       = errors.New("cache key too long")
	ErrKeyNotFoundForRefresh = errors.New("key not found for refresh")
	ErrMockError             = errors.New("mock error")
	ErrMockValidation        = errors.New("mock validation error")
	ErrMockRefresh           = errors.New("mock refresh error")
	ErrMockRefreshAll        = errors.New("mock refresh all error")
)

// DefaultPathCache implements the PathCache interface with LRU/TTL eviction
type DefaultPathCache struct {
	mu           sync.RWMutex
	items        map[string]*cacheItem
	lruList      *list.List
	maxSize      int
	ttl          time.Duration
	policy       EvictionPolicy
	stats        CacheStats
	cleanupTimer *time.Timer
}

type cacheItem struct {
	key         string
	value       PathBuilder
	element     *list.Element
	createdAt   time.Time
	accessAt    time.Time
	accessCount int64
}

// NewPathCache creates a new path cache with default settings
func NewPathCache() PathCache {
	return NewPathCacheWithOptions(CacheOptions{
		MaxSize: 1000,
		TTL:     5 * time.Minute,
		Policy:  EvictLRU,
	})
}

// NewPathCacheWithOptions creates a path cache with custom options
func NewPathCacheWithOptions(options CacheOptions) PathCache {
	cache := &DefaultPathCache{
		items:   make(map[string]*cacheItem),
		lruList: list.New(),
		maxSize: options.MaxSize,
		ttl:     options.TTL,
		policy:  options.Policy,
		stats: CacheStats{
			MaxSize: options.MaxSize,
			TTL:     options.TTL,
		},
	}

	// Start cleanup timer for TTL expiration
	if cache.ttl > 0 {
		cache.startCleanupTimer()
	}

	return cache
}

// Get retrieves a value from the cache
func (c *DefaultPathCache) Get(key string) (PathBuilder, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	if !exists {
		c.stats.Misses++
		return nil, false
	}

	// Check TTL expiration
	if c.ttl > 0 && time.Since(item.createdAt) > c.ttl {
		c.removeItem(item)
		c.stats.Misses++
		return nil, false
	}

	// Update access information for LRU/LFU
	item.accessAt = time.Now()
	item.accessCount++

	// Move to front for LRU
	if c.policy == EvictLRU {
		c.lruList.MoveToFront(item.element)
	}

	c.stats.Hits++
	return item.value, true
}

// Set stores a value in the cache
func (c *DefaultPathCache) Set(key string, path PathBuilder) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if item already exists
	if item, exists := c.items[key]; exists {
		item.value = path
		item.createdAt = time.Now()
		item.accessAt = time.Now()
		item.accessCount++

		if c.policy == EvictLRU {
			c.lruList.MoveToFront(item.element)
		}
		return nil
	}

	// Create new item
	now := time.Now()
	item := &cacheItem{
		key:         key,
		value:       path,
		createdAt:   now,
		accessAt:    now,
		accessCount: 1,
	}

	// Add to LRU list
	item.element = c.lruList.PushFront(item)
	c.items[key] = item
	c.stats.Size++

	// Evict if necessary
	if c.stats.Size > c.maxSize {
		c.evict()
	}

	return nil
}

// Delete removes a key from the cache
func (c *DefaultPathCache) Delete(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, exists := c.items[key]; exists {
		c.removeItem(item)
		return nil
	}

	return fmt.Errorf("%w: %s", ErrKeyNotFound, key)
}

// Clear removes all items from the cache
func (c *DefaultPathCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
	c.lruList.Init()
	c.stats.Size = 0
	return nil
}

// Stats returns cache statistics
func (c *DefaultPathCache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.stats
}

// Keys returns all cache keys
func (c *DefaultPathCache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}

	return keys
}

// Size returns the current cache size
func (c *DefaultPathCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.stats.Size
}

// Contains checks if a key exists in the cache
func (c *DefaultPathCache) Contains(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, exists := c.items[key]
	return exists
}

// Expire removes expired entries based on TTL
func (c *DefaultPathCache) Expire() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.ttl <= 0 {
		return 0
	}

	expired := 0
	now := time.Now()

	for _, item := range c.items {
		if now.Sub(item.createdAt) > c.ttl {
			c.removeItem(item)
			expired++
		}
	}

	return expired
}

// SetMaxSize sets the maximum cache size
func (c *DefaultPathCache) SetMaxSize(size int) PathCache {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.maxSize = size
	c.stats.MaxSize = size

	// Evict if necessary
	for c.stats.Size > c.maxSize {
		c.evict()
	}

	return c
}

// SetTTL sets the time-to-live for cache entries
func (c *DefaultPathCache) SetTTL(ttl time.Duration) PathCache {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.ttl = ttl
	c.stats.TTL = ttl

	// Restart cleanup timer if needed
	if c.cleanupTimer != nil {
		c.cleanupTimer.Stop()
	}

	if ttl > 0 {
		c.startCleanupTimer()
	}

	return c
}

// SetEvictionPolicy sets the eviction policy
func (c *DefaultPathCache) SetEvictionPolicy(policy EvictionPolicy) PathCache {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.policy = policy
	return c
}

// Validate validates a cache key
func (c *DefaultPathCache) Validate(key string) error {
	if key == "" {
		return ErrCacheKeyEmpty
	}

	if len(key) > 1000 {
		return fmt.Errorf("%w: %d characters", ErrCacheKeyTooLong, len(key))
	}

	return nil
}

// Refresh updates a cache entry by re-validating it
func (c *DefaultPathCache) Refresh(key string) error {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("%w: %s", ErrKeyNotFoundForRefresh, key)
	}

	// Update access time
	c.mu.Lock()
	defer c.mu.Unlock()

	item.accessAt = time.Now()
	item.accessCount++

	return nil
}

// RefreshAll refreshes all cache entries
func (c *DefaultPathCache) RefreshAll() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for _, item := range c.items {
		item.accessAt = now
		item.accessCount++
	}

	return nil
}

// evict removes items based on the eviction policy
func (c *DefaultPathCache) evict() {
	if c.stats.Size <= c.maxSize {
		return
	}

	itemToEvict := c.selectItemToEvict()
	if itemToEvict != nil {
		c.removeItem(itemToEvict)
		c.stats.Evictions++
	}
}

// selectItemToEvict selects the item to evict based on the policy
func (c *DefaultPathCache) selectItemToEvict() *cacheItem {
	switch c.policy {
	case EvictLRU:
		return c.selectLRUItem()
	case EvictLFU:
		return c.selectLFUItem()
	case EvictFIFO:
		return c.selectFIFOItem()
	case EvictRandom:
		return c.selectRandomItem()
	case EvictTTL:
		return c.selectTTLItem()
	default:
		return nil
	}
}

// selectLRUItem selects the least recently used item
func (c *DefaultPathCache) selectLRUItem() *cacheItem {
	if element := c.lruList.Back(); element != nil {
		if item, ok := element.Value.(*cacheItem); ok {
			return item
		}
	}
	return nil
}

// selectLFUItem selects the least frequently used item
func (c *DefaultPathCache) selectLFUItem() *cacheItem {
	var itemToEvict *cacheItem
	var minCount int64 = -1

	for _, item := range c.items {
		if minCount == -1 || item.accessCount < minCount {
			minCount = item.accessCount
			itemToEvict = item
		}
	}

	return itemToEvict
}

// selectFIFOItem selects the first in, first out item (oldest created)
func (c *DefaultPathCache) selectFIFOItem() *cacheItem {
	var itemToEvict *cacheItem
	var oldestTime time.Time

	for _, item := range c.items {
		if oldestTime.IsZero() || item.createdAt.Before(oldestTime) {
			oldestTime = item.createdAt
			itemToEvict = item
		}
	}

	return itemToEvict
}

// selectRandomItem selects a random item (first one encountered)
func (c *DefaultPathCache) selectRandomItem() *cacheItem {
	for _, item := range c.items {
		return item
	}
	return nil
}

// selectTTLItem selects the item with the oldest access time
func (c *DefaultPathCache) selectTTLItem() *cacheItem {
	var itemToEvict *cacheItem
	var oldestAccess time.Time

	for _, item := range c.items {
		if oldestAccess.IsZero() || item.accessAt.Before(oldestAccess) {
			oldestAccess = item.accessAt
			itemToEvict = item
		}
	}

	return itemToEvict
}

// removeItem removes an item from both the map and LRU list
func (c *DefaultPathCache) removeItem(item *cacheItem) {
	delete(c.items, item.key)
	if item.element != nil {
		c.lruList.Remove(item.element)
	}
	c.stats.Size--
}

// startCleanupTimer starts a periodic cleanup timer for TTL expiration
func (c *DefaultPathCache) startCleanupTimer() {
	cleanupInterval := c.ttl / 4 // Clean up every quarter of TTL period
	if cleanupInterval < time.Minute {
		cleanupInterval = time.Minute
	}

	c.cleanupTimer = time.AfterFunc(cleanupInterval, func() {
		c.Expire()
		c.startCleanupTimer() // Restart timer
	})
}

// CacheOptions holds configuration options for path cache
type CacheOptions struct {
	MaxSize int
	TTL     time.Duration
	Policy  EvictionPolicy
}

// SpecializedPathCaches provides common cache use cases

// NewFileInfoCache creates a cache optimized for file info
func NewFileInfoCache() PathCache {
	return NewPathCacheWithOptions(CacheOptions{
		MaxSize: 5000,
		TTL:     2 * time.Minute,
		Policy:  EvictLRU,
	})
}

// NewAbsolutePathCache creates a cache for absolute path resolutions
func NewAbsolutePathCache() PathCache {
	return NewPathCacheWithOptions(CacheOptions{
		MaxSize: 1000,
		TTL:     10 * time.Minute,
		Policy:  EvictLRU,
	})
}

// NewPatternCache creates a cache for pattern matching results
func NewPatternCache() PathCache {
	return NewPathCacheWithOptions(CacheOptions{
		MaxSize: 500,
		TTL:     5 * time.Minute,
		Policy:  EvictLFU, // Pattern matches are likely to be reused
	})
}

// MockPathCache implements PathCache for testing
type MockPathCache struct {
	GetCalls      []string
	SetCalls      []MockSetCall
	DeleteCalls   []string
	ClearCalls    int
	StatsReturns  CacheStats
	KeysReturns   []string
	SizeReturns   int
	ContainsCalls []string
	ExpireCalls   int
	ShouldError   bool
	Storage       map[string]PathBuilder
}

// MockSetCall represents a call to the Set method in MockPathCache
type MockSetCall struct {
	Key   string
	Value PathBuilder
}

// NewMockPathCache creates a new mock path cache for testing
func NewMockPathCache() *MockPathCache {
	return &MockPathCache{
		GetCalls:    make([]string, 0),
		SetCalls:    make([]MockSetCall, 0),
		DeleteCalls: make([]string, 0),
		Storage:     make(map[string]PathBuilder),
	}
}

// Get retrieves a path from the mock cache
func (m *MockPathCache) Get(key string) (PathBuilder, bool) {
	m.GetCalls = append(m.GetCalls, key)
	if m.ShouldError {
		return nil, false
	}

	value, exists := m.Storage[key]
	return value, exists
}

// Set stores a path in the mock cache
func (m *MockPathCache) Set(key string, path PathBuilder) error {
	m.SetCalls = append(m.SetCalls, MockSetCall{Key: key, Value: path})
	if m.ShouldError {
		return ErrMockError
	}
	m.Storage[key] = path
	return nil
}

// Delete removes a path from the mock cache
func (m *MockPathCache) Delete(key string) error {
	m.DeleteCalls = append(m.DeleteCalls, key)
	if m.ShouldError {
		return ErrMockError
	}

	_, exists := m.Storage[key]
	if exists {
		delete(m.Storage, key)
		return nil
	}
	return fmt.Errorf("%w: %s", ErrKeyNotFound, key)
}

// Clear removes all paths from the mock cache
func (m *MockPathCache) Clear() error {
	m.ClearCalls++
	if m.ShouldError {
		return ErrMockError
	}
	m.Storage = make(map[string]PathBuilder)
	return nil
}

// Stats returns mock cache statistics
func (m *MockPathCache) Stats() CacheStats {
	return m.StatsReturns
}

// Keys returns all keys in the mock cache
func (m *MockPathCache) Keys() []string {
	if len(m.KeysReturns) > 0 {
		return m.KeysReturns
	}

	keys := make([]string, 0, len(m.Storage))
	for key := range m.Storage {
		keys = append(keys, key)
	}
	return keys
}

// Size returns the number of items in the mock cache
func (m *MockPathCache) Size() int {
	if m.SizeReturns > 0 {
		return m.SizeReturns
	}
	return len(m.Storage)
}

// Contains checks if a key exists in the mock cache
func (m *MockPathCache) Contains(key string) bool {
	m.ContainsCalls = append(m.ContainsCalls, key)
	_, exists := m.Storage[key]
	return exists
}

// Expire removes expired items from the mock cache
func (m *MockPathCache) Expire() int {
	m.ExpireCalls++
	return 0
}

// SetMaxSize sets the maximum size of the mock cache
func (m *MockPathCache) SetMaxSize(_ int) PathCache {
	return m
}

// SetTTL sets the time-to-live for the mock cache
func (m *MockPathCache) SetTTL(_ time.Duration) PathCache {
	return m
}

// SetEvictionPolicy sets the eviction policy for the mock cache
func (m *MockPathCache) SetEvictionPolicy(_ EvictionPolicy) PathCache {
	return m
}

// Validate validates a key in the mock cache
func (m *MockPathCache) Validate(_ string) error {
	if m.ShouldError {
		return ErrMockValidation
	}
	return nil
}

// Refresh refreshes a specific key in the mock cache
func (m *MockPathCache) Refresh(_ string) error {
	if m.ShouldError {
		return ErrMockRefresh
	}
	return nil
}

// RefreshAll refreshes all items in the mock cache
func (m *MockPathCache) RefreshAll() error {
	if m.ShouldError {
		return ErrMockRefreshAll
	}
	return nil
}
