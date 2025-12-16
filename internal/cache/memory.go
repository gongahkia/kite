package cache

import (
	"context"
	"sync"
	"time"
)

// cacheItem represents a cached item with expiration
type cacheItem struct {
	Value      interface{}
	Expiration int64
}

// isExpired checks if the item has expired
func (item *cacheItem) isExpired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

// MemoryCache is an in-memory cache implementation
type MemoryCache struct {
	items      map[string]*cacheItem
	mu         sync.RWMutex
	hits       int64
	misses     int64
	evictions  int64
	maxSize    int64
	maxKeys    int
	cleanupInterval time.Duration
	stopCleanup    chan bool
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache(config *Config) *MemoryCache {
	if config == nil {
		config = &Config{
			MaxKeys: 1000,
			TTL:     5 * time.Minute,
		}
	}

	mc := &MemoryCache{
		items:           make(map[string]*cacheItem),
		maxSize:         config.MaxSize,
		maxKeys:         config.MaxKeys,
		cleanupInterval: 1 * time.Minute,
		stopCleanup:     make(chan bool),
	}

	// Start cleanup goroutine
	go mc.cleanupExpired()

	return mc
}

// Get retrieves a value from the cache
func (mc *MemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	mc.mu.RLock()
	item, found := mc.items[key]
	mc.mu.RUnlock()

	if !found {
		mc.mu.Lock()
		mc.misses++
		mc.mu.Unlock()
		return nil, ErrCacheMiss
	}

	if item.isExpired() {
		mc.mu.Lock()
		delete(mc.items, key)
		mc.misses++
		mc.mu.Unlock()
		return nil, ErrCacheMiss
	}

	mc.mu.Lock()
	mc.hits++
	mc.mu.Unlock()

	return item.Value, nil
}

// Set stores a value in the cache with TTL
func (mc *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Check if we need to evict
	if mc.maxKeys > 0 && len(mc.items) >= mc.maxKeys {
		mc.evictOldest()
	}

	var expiration int64
	if ttl > 0 {
		expiration = time.Now().Add(ttl).UnixNano()
	}

	mc.items[key] = &cacheItem{
		Value:      value,
		Expiration: expiration,
	}

	return nil
}

// Delete removes a value from the cache
func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	delete(mc.items, key)
	return nil
}

// Exists checks if a key exists in the cache
func (mc *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	mc.mu.RLock()
	item, found := mc.items[key]
	mc.mu.RUnlock()

	if !found {
		return false, nil
	}

	if item.isExpired() {
		mc.mu.Lock()
		delete(mc.items, key)
		mc.mu.Unlock()
		return false, nil
	}

	return true, nil
}

// Clear clears all cached data
func (mc *MemoryCache) Clear(ctx context.Context) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.items = make(map[string]*cacheItem)
	return nil
}

// GetMulti retrieves multiple values
func (mc *MemoryCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, key := range keys {
		if value, err := mc.Get(ctx, key); err == nil {
			result[key] = value
		}
	}

	return result, nil
}

// SetMulti stores multiple values
func (mc *MemoryCache) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	for key, value := range items {
		if err := mc.Set(ctx, key, value, ttl); err != nil {
			return err
		}
	}
	return nil
}

// DeleteMulti removes multiple values
func (mc *MemoryCache) DeleteMulti(ctx context.Context, keys []string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	for _, key := range keys {
		delete(mc.items, key)
	}

	return nil
}

// Close closes the cache
func (mc *MemoryCache) Close() error {
	mc.stopCleanup <- true
	return nil
}

// Stats returns cache statistics
func (mc *MemoryCache) Stats(ctx context.Context) (*Stats, error) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	total := mc.hits + mc.misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(mc.hits) / float64(total)
	}

	return &Stats{
		Hits:      mc.hits,
		Misses:    mc.misses,
		Keys:      int64(len(mc.items)),
		Evictions: mc.evictions,
		HitRate:   hitRate,
	}, nil
}

// cleanupExpired removes expired items periodically
func (mc *MemoryCache) cleanupExpired() {
	ticker := time.NewTicker(mc.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mc.mu.Lock()
			for key, item := range mc.items {
				if item.isExpired() {
					delete(mc.items, key)
				}
			}
			mc.mu.Unlock()

		case <-mc.stopCleanup:
			return
		}
	}
}

// evictOldest evicts the oldest item (simple LRU approximation)
func (mc *MemoryCache) evictOldest() {
	// Simple eviction: remove first item
	for key := range mc.items {
		delete(mc.items, key)
		mc.evictions++
		return
	}
}
