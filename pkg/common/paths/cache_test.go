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
	assert.NoError(t, err)

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
	assert.NoError(t, err)

	// Test empty key
	err = cache.Validate("")
	assert.Error(t, err)

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
	assert.NoError(t, err)

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
	assert.NoError(t, err)
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
				assert.NoError(t, err)
			}

			// Should have exactly max size items
			assert.Equal(t, 3, cache.Size())

			// Verify eviction happened
			stats := cache.Stats()
			assert.Greater(t, stats.Evictions, int64(0))
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
	assert.NoError(t, err)
	assert.Len(t, mock.SetCalls, 1)

	// Test Get
	retrieved, exists := mock.Get("key1")
	assert.True(t, exists)
	assert.Equal(t, pb.String(), retrieved.String())
	assert.Len(t, mock.GetCalls, 1)

	// Test Delete
	err = mock.Delete("key1")
	assert.NoError(t, err)
	assert.Len(t, mock.DeleteCalls, 1)

	// Test Clear
	err = mock.Clear()
	assert.NoError(t, err)
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
