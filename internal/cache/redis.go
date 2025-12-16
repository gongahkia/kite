package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache is a Redis-based cache implementation
type RedisCache struct {
	client *redis.Client
	prefix string
	ttl    time.Duration
}

// RedisConfig holds Redis cache configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Prefix   string
	TTL      time.Duration
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(config *RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
		prefix: config.Prefix,
		ttl:    config.TTL,
	}, nil
}

// Get retrieves a value from Redis
func (rc *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	fullKey := rc.prefix + key

	data, err := rc.client.Get(ctx, fullKey).Bytes()
	if err == redis.Nil {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, &CacheError{Op: "get", Key: key, Err: err}
	}

	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return nil, &CacheError{Op: "unmarshal", Key: key, Err: err}
	}

	return value, nil
}

// Set stores a value in Redis with TTL
func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := rc.prefix + key

	data, err := json.Marshal(value)
	if err != nil {
		return &CacheError{Op: "marshal", Key: key, Err: err}
	}

	if ttl == 0 {
		ttl = rc.ttl
	}

	if err := rc.client.Set(ctx, fullKey, data, ttl).Err(); err != nil {
		return &CacheError{Op: "set", Key: key, Err: err}
	}

	return nil
}

// Delete removes a value from Redis
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := rc.prefix + key

	if err := rc.client.Del(ctx, fullKey).Err(); err != nil {
		return &CacheError{Op: "delete", Key: key, Err: err}
	}

	return nil
}

// Exists checks if a key exists in Redis
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := rc.prefix + key

	count, err := rc.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, &CacheError{Op: "exists", Key: key, Err: err}
	}

	return count > 0, nil
}

// Clear clears all cached data with the prefix
func (rc *RedisCache) Clear(ctx context.Context) error {
	pattern := rc.prefix + "*"

	iter := rc.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := rc.client.Del(ctx, iter.Val()).Err(); err != nil {
			return &CacheError{Op: "clear", Err: err}
		}
	}

	if err := iter.Err(); err != nil {
		return &CacheError{Op: "clear", Err: err}
	}

	return nil
}

// GetMulti retrieves multiple values from Redis
func (rc *RedisCache) GetMulti(ctx context.Context, keys []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// Build full keys
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = rc.prefix + key
	}

	// Use pipeline for efficiency
	pipe := rc.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(fullKeys))

	for i, fullKey := range fullKeys {
		cmds[i] = pipe.Get(ctx, fullKey)
	}

	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		return nil, &CacheError{Op: "getmulti", Err: err}
	}

	// Process results
	for i, cmd := range cmds {
		data, err := cmd.Bytes()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			continue
		}

		var value interface{}
		if err := json.Unmarshal(data, &value); err != nil {
			continue
		}

		result[keys[i]] = value
	}

	return result, nil
}

// SetMulti stores multiple values in Redis
func (rc *RedisCache) SetMulti(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if ttl == 0 {
		ttl = rc.ttl
	}

	pipe := rc.client.Pipeline()

	for key, value := range items {
		fullKey := rc.prefix + key

		data, err := json.Marshal(value)
		if err != nil {
			return &CacheError{Op: "marshal", Key: key, Err: err}
		}

		pipe.Set(ctx, fullKey, data, ttl)
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return &CacheError{Op: "setmulti", Err: err}
	}

	return nil
}

// DeleteMulti removes multiple values from Redis
func (rc *RedisCache) DeleteMulti(ctx context.Context, keys []string) error {
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = rc.prefix + key
	}

	if err := rc.client.Del(ctx, fullKeys...).Err(); err != nil {
		return &CacheError{Op: "deletemulti", Err: err}
	}

	return nil
}

// Close closes the Redis connection
func (rc *RedisCache) Close() error {
	return rc.client.Close()
}

// Stats returns Redis cache statistics
func (rc *RedisCache) Stats(ctx context.Context) (*Stats, error) {
	// Get keyspace info
	info, err := rc.client.Info(ctx, "stats", "keyspace").Result()
	if err != nil {
		return nil, &CacheError{Op: "stats", Err: err}
	}

	// Parse stats (simplified)
	stats := &Stats{}

	// Count keys with prefix
	pattern := rc.prefix + "*"
	iter := rc.client.Scan(ctx, 0, pattern, 0).Iterator()
	keyCount := int64(0)
	for iter.Next(ctx) {
		keyCount++
	}
	stats.Keys = keyCount

	// Note: Redis doesn't provide hit/miss stats per prefix
	// You would need Redis module or separate tracking

	return stats, nil
}
