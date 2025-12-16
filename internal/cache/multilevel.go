package cache

import (
	"context"
	"time"
)

// MultiLevelCache implements a multi-tier caching system (L1: Memory, L2: Redis)
type MultiLevelCache struct {
	l1 Cache // Fast in-memory cache
	l2 Cache // Persistent Redis cache
}

// NewMultiLevelCache creates a new multi-level cache
func NewMultiLevelCache(l1, l2 Cache) *MultiLevelCache {
	return &MultiLevelCache{
		l1: l1,
		l2: l2,
	}
}

// Get retrieves a value, checking L1 first, then L2
func (mc *MultiLevelCache) Get(ctx context.Context, key string) (interface{}, error) {
	// Try L1 (memory) first
	value, err := mc.l1.Get(ctx, key)
	if err == nil {
		return value, nil
	}

	// Try L2 (Redis)
	value, err = mc.l2.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	// Populate L1 with L2 value
	_ = mc.l1.Set(ctx, key, value, 5*time.Minute) // Shorter TTL for L1

	return value, nil
}

// Set stores a value in both L1 and L2
func (mc *MultiLevelCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Write to both caches
	if err := mc.l2.Set(ctx, key, value, ttl); err != nil {
		return err
	}

	// Use shorter TTL for L1
	l1TTL := ttl / 2
	if l1TTL < time.Minute {
		l1TTL = time.Minute
	}

	return mc.l1.Set(ctx, key, value, l1TTL)
}

// Delete removes a value from both caches
func (mc *MultiLevelCache) Delete(ctx context.Context, key string) error {
	// Delete from both
	_ = mc.l1.Delete(ctx, key)
	return mc.l2.Delete(ctx, key)
}

// Exists checks if a key exists in either cache
func (mc *MultiLevelCache) Exists(ctx context.Context, key string) (bool, error) {
	// Check L1 first
	exists, err := mc.l1.Exists(ctx, key)
	if err == nil && exists {
		return true, nil
	}

	// Check L2
	return mc.l2.Exists(ctx, key)
}

// Clear clears both caches
func (mc *MultiLevelCache) Clear(ctx context.Context) error {
	_ = mc.l1.Clear(ctx)
	return mc.l2.Clear(ctx)
}

// GetMulti retrieves multiple values
func (mc *MultiLevelCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	missingKeys := make([]string, 0)

	// Try L1 first
	l1Results, _ := mc.l1.GetMulti(ctx, keys)
	for key, value := range l1Results {
		result[key] = value
	}

	// Find missing keys
	for _, key := range keys {
		if _, found := result[key]; !found {
			missingKeys = append(missingKeys, key)
		}
	}

	// Try L2 for missing keys
	if len(missingKeys) > 0 {
		l2Results, _ := mc.l2.GetMulti(ctx, missingKeys)

		// Populate L1 with L2 results
		if len(l2Results) > 0 {
			_ = mc.l1.SetMulti(ctx, l2Results, 5*time.Minute)
		}

		// Add L2 results to final result
		for key, value := range l2Results {
			result[key] = value
		}
	}

	return result, nil
}

// SetMulti stores multiple values in both caches
func (mc *MultiLevelCache) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	// Write to L2
	if err := mc.l2.SetMulti(ctx, items, ttl); err != nil {
		return err
	}

	// Write to L1 with shorter TTL
	l1TTL := ttl / 2
	if l1TTL < time.Minute {
		l1TTL = time.Minute
	}

	return mc.l1.SetMulti(ctx, items, l1TTL)
}

// DeleteMulti removes multiple values from both caches
func (mc *MultiLevelCache) DeleteMulti(ctx context.Context, keys []string) error {
	_ = mc.l1.DeleteMulti(ctx, keys)
	return mc.l2.DeleteMulti(ctx, keys)
}

// Close closes both caches
func (mc *MultiLevelCache) Close() error {
	_ = mc.l1.Close()
	return mc.l2.Close()
}

// Stats returns combined statistics
func (mc *MultiLevelCache) Stats(ctx context.Context) (*Stats, error) {
	l1Stats, _ := mc.l1.Stats(ctx)
	l2Stats, _ := mc.l2.Stats(ctx)

	combined := &Stats{
		Hits:      l1Stats.Hits + l2Stats.Hits,
		Misses:    l1Stats.Misses + l2Stats.Misses,
		Keys:      l2Stats.Keys, // L2 is authoritative
		Evictions: l1Stats.Evictions + l2Stats.Evictions,
	}

	total := combined.Hits + combined.Misses
	if total > 0 {
		combined.HitRate = float64(combined.Hits) / float64(total)
	}

	return combined, nil
}
