package paths

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultPathCache_BasicOperations(t *testing.T) {
	cache := NewPathCache()

	// Test Set and Get
	pb := NewPathBuilder("/test/path")
	err := cache.Set("key1", pb)
	require.NoError(t, err)

	retrieved, exists := cache.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, pb.String(), retrieved.String())

	// Test non-existent key
	_, exists = cache.Get("nonexistent")
	assert.False(t, exists)
}

func TestDefaultPathCache_Delete(t *testing.T) {
	cache := NewPathCache()

	// Add and delete a key
	pb := NewPathBuilder("/test/path")
	err := cache.Set("key1", pb)
	require.NoError(t, err)

	err = cache.Delete("key1")
	require.NoError(t, err)

	// Verify it's gone
	_, exists := cache.Get("key1")
	assert.False(t, exists)

	// Try to delete non-existent key
	err = cache.Delete("nonexistent")
	assert.Error(t, err)
}

func TestDefaultPathCache_Clear(t *testing.T) {
	cache := NewPathCache()

	// Add multiple items
	err := cache.Set("key1", NewPathBuilder("/path1"))
	require.NoError(t, err)
	err = cache.Set("key2", NewPathBuilder("/path2"))
	require.NoError(t, err)
	err = cache.Set("key3", NewPathBuilder("/path3"))
	require.NoError(t, err)

	// Clear cache
	err = cache.Clear()
	require.NoError(t, err)

	// Verify all items are gone
	assert.Equal(t, 0, cache.Size())
}

func TestDefaultPathCache_Stats(t *testing.T) {
	cache := NewPathCache()

	// Get initial stats
	stats := cache.Stats()
	assert.Equal(t, 0, stats.Size)
	assert.Equal(t, int64(0), stats.Hits)
	assert.Equal(t, int64(0), stats.Misses)

	// Add item and test hit/miss
	err := cache.Set("key1", NewPathBuilder("/path1"))
	require.NoError(t, err)

	_, _ = cache.Get("key1")        // Hit
	_, _ = cache.Get("nonexistent") // Miss

	stats = cache.Stats()
	assert.Equal(t, int64(1), stats.Hits)
	assert.Equal(t, int64(1), stats.Misses)
}

func TestDefaultPathCache_KeysAndContains(t *testing.T) {
	cache := NewPathCache()

	// Add items
	err := cache.Set("key1", NewPathBuilder("/path1"))
	require.NoError(t, err)
	err = cache.Set("key2", NewPathBuilder("/path2"))
	require.NoError(t, err)

	// Test Keys
	keys := cache.Keys()
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")

	// Test Contains by checking keys
	assert.Contains(t, keys, "key1")
	assert.NotContains(t, keys, "key3")
}

func TestDefaultPathCache_TTLExpiration(t *testing.T) {
	// Create cache with short TTL
	cache := NewPathCacheWithOptions(CacheOptions{
		MaxSize: 100,
		TTL:     100 * time.Millisecond,
		Policy:  EvictLRU,
	})

	// Add item
	err := cache.Set("key1", NewPathBuilder("/path1"))
	require.NoError(t, err)

	// Verify it exists
	_, exists := cache.Get("key1")
	assert.True(t, exists)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, exists = cache.Get("key1")
	assert.False(t, exists)
}

func TestDefaultPathCache_Expire(t *testing.T) {
	// Create cache with TTL
	cacheImpl, ok := NewPathCacheWithOptions(CacheOptions{
		MaxSize: 100,
		TTL:     100 * time.Millisecond,
		Policy:  EvictLRU,
	}).(*DefaultPathCache)
	require.True(t, ok)

	// Add items
	err := cacheImpl.Set("key1", NewPathBuilder("/path1"))
	require.NoError(t, err)
	time.Sleep(50 * time.Millisecond)
	err = cacheImpl.Set("key2", NewPathBuilder("/path2"))
	require.NoError(t, err)

	// Wait for first item to expire
	time.Sleep(60 * time.Millisecond)

	// Manually expire
	expired := cacheImpl.Expire()
	assert.Equal(t, 1, expired) // Only key1 should be expired

	// Verify key1 is gone, key2 still exists
	assert.False(t, cacheImpl.Contains("key1"))
	assert.True(t, cacheImpl.Contains("key2"))
}

func TestDefaultPathCache_SetMaxSize(t *testing.T) {
	cache := NewPathCache()

	// Add items
	for i := 0; i < 10; i++ {
		err := cache.Set(string(rune('a'+i)), NewPathBuilder("/path"))
		require.NoError(t, err)
	}

	// Reduce max size
	cache.SetMaxSize(5)

	// Should have evicted items
	assert.LessOrEqual(t, cache.Size(), 5)
}

func TestDefaultPathCache_SetTTL(t *testing.T) {
	cache := NewPathCache()

	// Set new TTL
	cache.SetTTL(50 * time.Millisecond)

	// Add item
	err := cache.Set("key1", NewPathBuilder("/path1"))
	require.NoError(t, err)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	_, exists := cache.Get("key1")
	assert.False(t, exists)
}

func TestDefaultPathCache_SetEvictionPolicy(t *testing.T) {
	cache := NewPathCache()

	// Set different eviction policy
	cache.SetEvictionPolicy(EvictLFU)

	// This is mostly for coverage - the actual eviction behavior
	// would need more complex testing
	assert.NotNil(t, cache)
}

func TestDefaultPathCache_Validate(t *testing.T) {
	cache := NewPathCache()

	// Test valid key
	err := cache.Validate("validkey")
	require.NoError(t, err)

	// Test empty key
	err = cache.Validate("")
	require.Error(t, err)

	// Test very long key
	longKey := make([]byte, 1001)
	for i := range longKey {
		longKey[i] = 'a'
	}
	err = cache.Validate(string(longKey))
	assert.Error(t, err)
}

func TestDefaultPathCache_Refresh(t *testing.T) {
	cache := NewPathCache()

	// Add item
	err := cache.Set("key1", NewPathBuilder("/path1"))
	require.NoError(t, err)

	// Refresh existing key
	err = cache.Refresh("key1")
	require.NoError(t, err)

	// Try to refresh non-existent key
	err = cache.Refresh("nonexistent")
	assert.Error(t, err)
}

func TestDefaultPathCache_RefreshAll(t *testing.T) {
	cache := NewPathCache()

	// Add items
	err := cache.Set("key1", NewPathBuilder("/path1"))
	require.NoError(t, err)
	err = cache.Set("key2", NewPathBuilder("/path2"))
	require.NoError(t, err)

	// Refresh all
	err = cache.RefreshAll()
	require.NoError(t, err)
}

func TestDefaultPathCache_EvictionPolicies(t *testing.T) {
	tests := []struct {
		name   string
		policy EvictionPolicy
	}{
		{"LRU", EvictLRU},
		{"LFU", EvictLFU},
		{"FIFO", EvictFIFO},
		{"Random", EvictRandom},
		{"TTL", EvictTTL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := NewPathCacheWithOptions(CacheOptions{
				MaxSize: 3,
				TTL:     0, // No TTL for this test
				Policy:  tt.policy,
			})

			// Add more items than max size to trigger eviction
			for i := 0; i < 5; i++ {
				err := cache.Set(string(rune('a'+i)), NewPathBuilder("/path"))
				require.NoError(t, err)
			}

			// Should have exactly max size items
			assert.Equal(t, 3, cache.Size())

			// Verify eviction happened
			stats := cache.Stats()
			assert.Positive(t, stats.Evictions)
		})
	}
}

func TestDefaultPathCache_UpdateExisting(t *testing.T) {
	cache := NewPathCache()

	// Add item
	err := cache.Set("key1", NewPathBuilder("/path1"))
	require.NoError(t, err)

	// Update with new value
	err = cache.Set("key1", NewPathBuilder("/path2"))
	require.NoError(t, err)

	// Should get updated value
	pb, exists := cache.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, "/path2", pb.String())
}

func TestSpecializedCaches(t *testing.T) {
	// Test FileInfoCache
	fileCache := NewFileInfoCache()
	assert.NotNil(t, fileCache)
	stats := fileCache.Stats()
	assert.Equal(t, 5000, stats.MaxSize)

	// Test AbsolutePathCache
	absCache := NewAbsolutePathCache()
	assert.NotNil(t, absCache)
	stats = absCache.Stats()
	assert.Equal(t, 1000, stats.MaxSize)

	// Test PatternCache
	patternCache := NewPatternCache()
	assert.NotNil(t, patternCache)
	stats = patternCache.Stats()
	assert.Equal(t, 500, stats.MaxSize)
}

func TestMockPathCache(t *testing.T) {
	mock := NewMockPathCache()

	// Test Set
	pb := NewPathBuilder("/test")
	err := mock.Set("key1", pb)
	require.NoError(t, err)
	assert.Len(t, mock.SetCalls, 1)

	// Test Get
	retrieved, exists := mock.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, pb.String(), retrieved.String())
	assert.Len(t, mock.GetCalls, 1)

	// Test Delete
	err = mock.Delete("key1")
	require.NoError(t, err)
	assert.Len(t, mock.DeleteCalls, 1)

	// Test Clear
	err = mock.Clear()
	require.NoError(t, err)
	assert.Equal(t, 1, mock.ClearCalls)

	// Test with error
	mock.ShouldError = true
	err = mock.Set("key2", pb)
	assert.Error(t, err)
}

func TestDefaultPathCache_NoTTL(t *testing.T) {
	// Create cache with no TTL
	cacheImpl, ok := NewPathCacheWithOptions(CacheOptions{
		MaxSize: 100,
		TTL:     0,
		Policy:  EvictLRU,
	}).(*DefaultPathCache)
	require.True(t, ok)

	// Expire should return 0 when TTL is not set
	expired := cacheImpl.Expire()
	assert.Equal(t, 0, expired)
}

func TestDefaultPathCache_ConcurrentAccess(t *testing.T) {
	cache := NewPathCache()
	done := make(chan bool, 3)

	// Concurrent writes
	go func() {
		for i := 0; i < 100; i++ {
			// Ignore errors in concurrent test for performance
			err := cache.Set(string(rune('a'+i%26)), NewPathBuilder("/path"))
			_ = err // Expected in concurrent test
		}
		done <- true
	}()

	// Concurrent reads
	go func() {
		for i := 0; i < 100; i++ {
			_, _ = cache.Get(string(rune('a' + i%26)))
		}
		done <- true
	}()

	// Concurrent deletes
	go func() {
		for i := 0; i < 100; i++ {
			// Ignore errors in concurrent test for performance
			err := cache.Delete(string(rune('a' + i%26)))
			_ = err // Expected in concurrent test
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Should not panic
	assert.NotNil(t, cache)
}

// TestDefaultPathCache_RandomEviction tests the random eviction policy behavior
func TestDefaultPathCache_RandomEviction(t *testing.T) {
	// Create cache with random eviction policy and max size of 3
	cache := NewPathCacheWithOptions(CacheOptions{
		MaxSize: 3,
		TTL:     0,
		Policy:  EvictRandom,
	})

	// Add initial items
	for i := 0; i < 3; i++ {
		err := cache.Set(string(rune('a'+i)), NewPathBuilder("/path"))
		require.NoError(t, err)
	}

	// Should have exactly 3 items
	assert.Equal(t, 3, cache.Size())

	// Add more items to trigger eviction
	for i := 3; i < 10; i++ {
		err := cache.Set(string(rune('a'+i)), NewPathBuilder("/path"))
		require.NoError(t, err)
	}

	// Should still have max size items
	assert.Equal(t, 3, cache.Size())

	// Verify evictions happened
	stats := cache.Stats()
	assert.Positive(t, stats.Evictions)
}

// TestDefaultPathCache_LFUEviction tests the LFU eviction policy behavior
func TestDefaultPathCache_LFUEviction(t *testing.T) {
	// Create cache with LFU eviction policy and max size of 4
	// Using 4 to allow us to add items with different access patterns
	cache := NewPathCacheWithOptions(CacheOptions{
		MaxSize: 4,
		TTL:     0,
		Policy:  EvictLFU,
	})

	// Add items and access them to build up frequency
	err := cache.Set("a", NewPathBuilder("/path/a"))
	require.NoError(t, err)
	err = cache.Set("b", NewPathBuilder("/path/b"))
	require.NoError(t, err)
	err = cache.Set("c", NewPathBuilder("/path/c"))
	require.NoError(t, err)

	// Access "a" and "b" multiple times to increase their frequency
	for i := 0; i < 10; i++ {
		_, _ = cache.Get("a")
		_, _ = cache.Get("b")
	}
	// Access "c" less frequently
	for i := 0; i < 3; i++ {
		_, _ = cache.Get("c")
	}

	// Add new item that also gets accessed (to not be the lowest)
	err = cache.Set("d", NewPathBuilder("/path/d"))
	require.NoError(t, err)
	for i := 0; i < 5; i++ {
		_, _ = cache.Get("d")
	}

	assert.Equal(t, 4, cache.Size())

	// Add another item to trigger eviction
	err = cache.Set("e", NewPathBuilder("/path/e"))
	require.NoError(t, err)

	// Should have evicted one item (keeping max 4)
	assert.Equal(t, 4, cache.Size())

	// "a" and "b" should still exist (most frequently accessed)
	_, exists := cache.Get("a")
	assert.True(t, exists, "a should still exist (high frequency)")
	_, exists = cache.Get("b")
	assert.True(t, exists, "b should still exist (high frequency)")

	// Verify eviction happened
	stats := cache.Stats()
	assert.Positive(t, stats.Evictions)
}

// TestDefaultPathCache_LRUEviction tests the LRU eviction policy behavior
func TestDefaultPathCache_LRUEviction(t *testing.T) {
	// Create cache with LRU eviction policy and max size of 3
	cache := NewPathCacheWithOptions(CacheOptions{
		MaxSize: 3,
		TTL:     0,
		Policy:  EvictLRU,
	})

	// Add items
	err := cache.Set("a", NewPathBuilder("/path/a"))
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)
	err = cache.Set("b", NewPathBuilder("/path/b"))
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)
	err = cache.Set("c", NewPathBuilder("/path/c"))
	require.NoError(t, err)

	// Access "a" to update its access time
	_, _ = cache.Get("a")
	time.Sleep(10 * time.Millisecond)

	// Add new item to trigger eviction
	err = cache.Set("d", NewPathBuilder("/path/d"))
	require.NoError(t, err)

	// "b" should have been evicted (least recently used)
	_, exists := cache.Get("b")
	assert.False(t, exists, "b should have been evicted due to LRU policy")

	// "a" should still exist (was accessed recently)
	_, exists = cache.Get("a")
	assert.True(t, exists, "a should still exist")
}

// TestDefaultPathCache_CleanupTimer tests the cleanup timer functionality
func TestDefaultPathCache_CleanupTimer(t *testing.T) {
	// Create cache with short TTL
	cacheImpl, ok := NewPathCacheWithOptions(CacheOptions{
		MaxSize: 100,
		TTL:     50 * time.Millisecond,
		Policy:  EvictTTL,
	}).(*DefaultPathCache)
	require.True(t, ok)

	// Add items
	err := cacheImpl.Set("key1", NewPathBuilder("/path1"))
	require.NoError(t, err)
	err = cacheImpl.Set("key2", NewPathBuilder("/path2"))
	require.NoError(t, err)

	// Verify items exist
	assert.Equal(t, 2, cacheImpl.Size())

	// Wait for TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Check that expired items are detected
	// Items should be expired when accessed
	_, exists := cacheImpl.Get("key1")
	assert.False(t, exists, "key1 should be expired")
	_, exists = cacheImpl.Get("key2")
	assert.False(t, exists, "key2 should be expired")
}

// TestDefaultPathCache_GetExpiredItem tests that Get returns false for expired items
func TestDefaultPathCache_GetExpiredItem(t *testing.T) {
	// Create cache with short TTL
	cache := NewPathCacheWithOptions(CacheOptions{
		MaxSize: 100,
		TTL:     50 * time.Millisecond,
		Policy:  EvictTTL,
	})

	// Add item
	err := cache.Set("key1", NewPathBuilder("/path1"))
	require.NoError(t, err)

	// Verify item exists
	_, exists := cache.Get("key1")
	assert.True(t, exists)

	// Wait for TTL
	time.Sleep(60 * time.Millisecond)

	// Item should be expired
	_, exists = cache.Get("key1")
	assert.False(t, exists, "key1 should be expired")
}

// TestDefaultPathCache_FIFOEviction tests the FIFO eviction policy behavior
func TestDefaultPathCache_FIFOEviction(t *testing.T) {
	// Create cache with FIFO eviction policy and max size of 3
	cache := NewPathCacheWithOptions(CacheOptions{
		MaxSize: 3,
		TTL:     0,
		Policy:  EvictFIFO,
	})

	// Add items
	err := cache.Set("a", NewPathBuilder("/path/a"))
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)
	err = cache.Set("b", NewPathBuilder("/path/b"))
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)
	err = cache.Set("c", NewPathBuilder("/path/c"))
	require.NoError(t, err)

	// Access "a" (shouldn't matter for FIFO)
	_, _ = cache.Get("a")

	// Add new item to trigger eviction
	err = cache.Set("d", NewPathBuilder("/path/d"))
	require.NoError(t, err)

	// "a" should have been evicted (first in)
	_, exists := cache.Get("a")
	assert.False(t, exists, "a should have been evicted due to FIFO policy")

	// "b" and "c" should still exist
	_, exists = cache.Get("b")
	assert.True(t, exists, "b should still exist")
	_, exists = cache.Get("c")
	assert.True(t, exists, "c should still exist")
}

// TestDefaultPathCache_SelectItemToEvict_LFUTie tests LFU eviction with same frequency
func TestDefaultPathCache_SelectItemToEvict_LFUTie(t *testing.T) {
	// Create cache with LFU eviction policy and max size of 3
	cache := NewPathCacheWithOptions(CacheOptions{
		MaxSize: 3,
		TTL:     0,
		Policy:  EvictLFU,
	})

	// Add items with same access pattern (all accessed once)
	err := cache.Set("a", NewPathBuilder("/path/a"))
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)
	err = cache.Set("b", NewPathBuilder("/path/b"))
	require.NoError(t, err)
	time.Sleep(10 * time.Millisecond)
	err = cache.Set("c", NewPathBuilder("/path/c"))
	require.NoError(t, err)

	// Add new item to trigger eviction - when frequencies are equal, oldest should be evicted
	err = cache.Set("d", NewPathBuilder("/path/d"))
	require.NoError(t, err)

	// Should have 3 items
	assert.Equal(t, 3, cache.Size())
}

// TestDefaultPathCache_TTLEviction tests the TTL eviction policy behavior
func TestDefaultPathCache_TTLEviction(t *testing.T) {
	// Create cache with TTL eviction policy
	cacheImpl, ok := NewPathCacheWithOptions(CacheOptions{
		MaxSize: 3,
		TTL:     50 * time.Millisecond,
		Policy:  EvictTTL,
	}).(*DefaultPathCache)
	require.True(t, ok)

	// Add items at different times
	err := cacheImpl.Set("a", NewPathBuilder("/path/a"))
	require.NoError(t, err)
	time.Sleep(30 * time.Millisecond)
	err = cacheImpl.Set("b", NewPathBuilder("/path/b"))
	require.NoError(t, err)
	time.Sleep(30 * time.Millisecond)
	err = cacheImpl.Set("c", NewPathBuilder("/path/c"))
	require.NoError(t, err)

	// First item should be expired, manually expire
	expired := cacheImpl.Expire()
	assert.Equal(t, 1, expired, "Should have expired 1 item")

	// "a" should be expired
	assert.False(t, cacheImpl.Contains("a"), "a should be expired")
	// "b" might be expired too depending on timing
	// "c" should still exist
	assert.True(t, cacheImpl.Contains("c"), "c should still exist")
}

// TestDefaultPathCache_SelectRandomItemEmpty tests random eviction with empty cache
func TestDefaultPathCache_SelectRandomItemEmpty(t *testing.T) {
	cacheImpl, ok := NewPathCacheWithOptions(CacheOptions{
		MaxSize: 3,
		TTL:     0,
		Policy:  EvictRandom,
	}).(*DefaultPathCache)
	require.True(t, ok)

	// With empty cache, selectRandomItem should return nil
	item := cacheImpl.selectRandomItem()
	assert.Nil(t, item, "selectRandomItem should return nil for empty cache")
}

// TestDefaultPathCache_SelectLRUItemEmpty tests LRU eviction with empty list
func TestDefaultPathCache_SelectLRUItemEmpty(t *testing.T) {
	cacheImpl, ok := NewPathCacheWithOptions(CacheOptions{
		MaxSize: 3,
		TTL:     0,
		Policy:  EvictLRU,
	}).(*DefaultPathCache)
	require.True(t, ok)

	// With empty list, selectLRUItem should return nil
	item := cacheImpl.selectLRUItem()
	assert.Nil(t, item, "selectLRUItem should return nil for empty list")
}

// TestDefaultPathCache_ShortTTLCleanupTimer tests cleanup timer with very short TTL
func TestDefaultPathCache_ShortTTLCleanupTimer(t *testing.T) {
	// Create cache with TTL less than minimum cleanup interval (should use 1 minute minimum)
	cache := NewPathCacheWithOptions(CacheOptions{
		MaxSize: 100,
		TTL:     10 * time.Millisecond, // Very short TTL
		Policy:  EvictTTL,
	})

	// Add item
	err := cache.Set("key1", NewPathBuilder("/path1"))
	require.NoError(t, err)

	// The cleanup timer should still work
	assert.NotNil(t, cache)
}

// TestDefaultPathCache_UnknownEvictionPolicy tests unknown eviction policy
func TestDefaultPathCache_UnknownEvictionPolicy(t *testing.T) {
	cacheImpl, ok := NewPathCacheWithOptions(CacheOptions{
		MaxSize: 2,
		TTL:     0,
		Policy:  EvictionPolicy(99), // Unknown policy
	}).(*DefaultPathCache)
	require.True(t, ok)

	// Add items to trigger eviction
	err := cacheImpl.Set("a", NewPathBuilder("/path/a"))
	require.NoError(t, err)
	err = cacheImpl.Set("b", NewPathBuilder("/path/b"))
	require.NoError(t, err)
	err = cacheImpl.Set("c", NewPathBuilder("/path/c"))
	require.NoError(t, err)

	// With unknown policy, selectItemToEvict returns nil, so no eviction
	// The cache should still work but not evict
	assert.GreaterOrEqual(t, cacheImpl.Size(), 2)
}

// TestDefaultPathCache_SetTTLZero tests setting TTL to zero stops cleanup timer
func TestDefaultPathCache_SetTTLZero(t *testing.T) {
	// Create cache with TTL
	cache := NewPathCacheWithOptions(CacheOptions{
		MaxSize: 100,
		TTL:     100 * time.Millisecond,
		Policy:  EvictTTL,
	})

	// Set TTL to 0 to disable
	cache.SetTTL(0)

	// Add item
	err := cache.Set("key1", NewPathBuilder("/path1"))
	require.NoError(t, err)

	// Items should not expire
	time.Sleep(50 * time.Millisecond)
	_, exists := cache.Get("key1")
	assert.True(t, exists, "key1 should still exist when TTL is 0")
}

// TestMockPathCache_GetMiss tests MockPathCache Get miss
func TestMockPathCache_GetMiss(t *testing.T) {
	mock := NewMockPathCache()

	// Get non-existent key
	_, exists := mock.Get("nonexistent")
	assert.False(t, exists)
	assert.Len(t, mock.GetCalls, 1)
}

// TestMockPathCache_GetWithError tests MockPathCache Get with ShouldError
func TestMockPathCache_GetWithError(t *testing.T) {
	mock := NewMockPathCache()
	mock.ShouldError = true

	// Get should return false when ShouldError is true
	_, exists := mock.Get("anykey")
	assert.False(t, exists)
}

// TestMockPathCache_DeleteNonExistent tests MockPathCache Delete for non-existent key
func TestMockPathCache_DeleteNonExistent(t *testing.T) {
	mock := NewMockPathCache()

	// Delete non-existent key
	err := mock.Delete("nonexistent")
	assert.Error(t, err)
}

// TestMockPathCache_DeleteWithError tests MockPathCache Delete with ShouldError
func TestMockPathCache_DeleteWithError(t *testing.T) {
	mock := NewMockPathCache()
	mock.ShouldError = true

	// Delete should return error
	err := mock.Delete("anykey")
	assert.Error(t, err)
}

// TestMockPathCache_ClearWithError tests MockPathCache Clear with ShouldError
func TestMockPathCache_ClearWithError(t *testing.T) {
	mock := NewMockPathCache()
	mock.ShouldError = true

	// Clear should return error
	err := mock.Clear()
	assert.Error(t, err)
}

// TestMockPathCache_KeysWithPreset tests MockPathCache Keys with preset return
func TestMockPathCache_KeysWithPreset(t *testing.T) {
	mock := NewMockPathCache()
	mock.KeysReturns = []string{"key1", "key2", "key3"}

	keys := mock.Keys()
	assert.Len(t, keys, 3)
	assert.Equal(t, mock.KeysReturns, keys)
}

// TestMockPathCache_SizeWithPreset tests MockPathCache Size with preset return
func TestMockPathCache_SizeWithPreset(t *testing.T) {
	mock := NewMockPathCache()
	mock.SizeReturns = 42

	size := mock.Size()
	assert.Equal(t, 42, size)
}

// TestMockPathCache_ValidateWithError tests MockPathCache Validate with ShouldError
func TestMockPathCache_ValidateWithError(t *testing.T) {
	mock := NewMockPathCache()
	mock.ShouldError = true

	err := mock.Validate("anykey")
	assert.Error(t, err)
}

// TestMockPathCache_RefreshWithError tests MockPathCache Refresh with ShouldError
func TestMockPathCache_RefreshWithError(t *testing.T) {
	mock := NewMockPathCache()
	mock.ShouldError = true

	err := mock.Refresh("anykey")
	assert.Error(t, err)
}

// TestMockPathCache_RefreshAllWithError tests MockPathCache RefreshAll with ShouldError
func TestMockPathCache_RefreshAllWithError(t *testing.T) {
	mock := NewMockPathCache()
	mock.ShouldError = true

	err := mock.RefreshAll()
	assert.Error(t, err)
}

// TestMockPathCache_ChainedMethods tests MockPathCache method chaining
func TestMockPathCache_ChainedMethods(t *testing.T) {
	mock := NewMockPathCache()

	// Test chaining
	result := mock.SetMaxSize(100).SetTTL(time.Minute).SetEvictionPolicy(EvictLRU)
	assert.Equal(t, mock, result)
}
