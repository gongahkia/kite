package cache

import (
	"fmt"
	"time"
)

// NewCache creates a cache based on configuration
func NewCache(config *Config) (Cache, error) {
	switch config.Type {
	case "memory", "":
		return NewMemoryCache(config), nil

	case "redis":
		// This would need Redis config from environment or config file
		redisConfig := &RedisConfig{
			Addr:   "localhost:6379",
			Prefix: "kite:",
			TTL:    config.TTL,
		}
		return NewRedisCache(redisConfig)

	case "multilevel":
		// Create L1 (memory) and L2 (Redis)
		l1 := NewMemoryCache(&Config{
			MaxKeys: 1000,
			TTL:     5 * time.Minute,
		})

		redisConfig := &RedisConfig{
			Addr:   "localhost:6379",
			Prefix: "kite:",
			TTL:    config.TTL,
		}

		l2, err := NewRedisCache(redisConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create Redis cache: %w", err)
		}

		return NewMultiLevelCache(l1, l2), nil

	default:
		return nil, fmt.Errorf("unknown cache type: %s", config.Type)
	}
}

// DefaultCache creates a cache with default configuration
func DefaultCache() Cache {
	return NewMemoryCache(&Config{
		MaxKeys: 10000,
		TTL:     10 * time.Minute,
	})
}

// CacheKey generates a cache key with prefix
func CacheKey(prefix, id string) string {
	return prefix + ":" + id
}

// CacheKeys generates multiple cache keys
func CacheKeys(prefix string, ids []string) []string {
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = CacheKey(prefix, id)
	}
	return keys
}
