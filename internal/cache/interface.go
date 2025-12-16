package cache

import (
	"context"
	"time"
)

// Cache is the interface for cache implementations
type Cache interface {
	// Get retrieves a value from the cache
	Get(ctx context.Context, key string) (interface{}, error)

	// Set stores a value in the cache with TTL
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a value from the cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key string) (bool, error)

	// Clear clears all cached data
	Clear(ctx context.Context) error

	// GetMulti retrieves multiple values
	GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error)

	// SetMulti stores multiple values
	SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error

	// DeleteMulti removes multiple values
	DeleteMulti(ctx context.Context, keys []string) error

	// Close closes the cache connection
	Close() error

	// Stats returns cache statistics
	Stats(ctx context.Context) (*Stats, error)
}

// Stats represents cache statistics
type Stats struct {
	Hits        int64
	Misses      int64
	Keys        int64
	Size        int64
	Evictions   int64
	HitRate     float64
}

// CacheError represents a cache error
type CacheError struct {
	Op  string
	Key string
	Err error
}

func (e *CacheError) Error() string {
	if e.Key != "" {
		return "cache " + e.Op + " " + e.Key + ": " + e.Err.Error()
	}
	return "cache " + e.Op + ": " + e.Err.Error()
}

// ErrCacheMiss indicates a cache miss
var ErrCacheMiss = &CacheError{Op: "get", Err: nil}

// InvalidationStrategy defines cache invalidation strategy
type InvalidationStrategy string

const (
	InvalidateOnWrite InvalidationStrategy = "write"
	InvalidateOnTTL   InvalidationStrategy = "ttl"
	InvalidateManual  InvalidationStrategy = "manual"
)

// Config holds cache configuration
type Config struct {
	Type     string        // "memory", "redis", "multilevel"
	TTL      time.Duration // Default TTL
	MaxSize  int64         // Maximum cache size (bytes)
	MaxKeys  int           // Maximum number of keys
	Strategy InvalidationStrategy
}
