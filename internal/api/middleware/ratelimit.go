package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gongahkia/kite/internal/observability"
	"golang.org/x/time/rate"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	// Requests per second
	RPS float64
	// Burst size
	Burst int
	// Key generator function (default: IP address)
	KeyGenerator func(*fiber.Ctx) string
	// Custom error handler
	ErrorHandler func(*fiber.Ctx) error
	// Storage for rate limiters (default: in-memory)
	Storage RateLimitStorage
}

// RateLimitStorage interface for storing rate limiters
type RateLimitStorage interface {
	Get(key string) *rate.Limiter
	Set(key string, limiter *rate.Limiter)
	Reset(key string)
	Clear()
}

// InMemoryRateLimitStorage is an in-memory implementation of RateLimitStorage
type InMemoryRateLimitStorage struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rps      float64
	burst    int
}

// NewInMemoryRateLimitStorage creates a new in-memory rate limit storage
func NewInMemoryRateLimitStorage(rps float64, burst int) *InMemoryRateLimitStorage {
	return &InMemoryRateLimitStorage{
		limiters: make(map[string]*rate.Limiter),
		rps:      rps,
		burst:    burst,
	}
}

// Get retrieves or creates a rate limiter for a key
func (s *InMemoryRateLimitStorage) Get(key string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()

	limiter, exists := s.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(s.rps), s.burst)
		s.limiters[key] = limiter
	}

	return limiter
}

// Set stores a rate limiter for a key
func (s *InMemoryRateLimitStorage) Set(key string, limiter *rate.Limiter) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.limiters[key] = limiter
}

// Reset removes a rate limiter for a key
func (s *InMemoryRateLimitStorage) Reset(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.limiters, key)
}

// Clear removes all rate limiters
func (s *InMemoryRateLimitStorage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.limiters = make(map[string]*rate.Limiter)
}

// DefaultRateLimitConfig returns a default rate limit configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		RPS:   10,
		Burst: 20,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		ErrorHandler: nil,
	}
}

// RateLimit creates a rate limiting middleware
func RateLimit(config *RateLimitConfig, logger *observability.Logger) fiber.Handler {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	if config.Storage == nil {
		config.Storage = NewInMemoryRateLimitStorage(config.RPS, config.Burst)
	}

	if config.KeyGenerator == nil {
		config.KeyGenerator = func(c *fiber.Ctx) string {
			return c.IP()
		}
	}

	return func(c *fiber.Ctx) error {
		// Generate key for this request
		key := config.KeyGenerator(c)

		// Get or create limiter for this key
		limiter := config.Storage.Get(key)

		// Check if request is allowed
		if !limiter.Allow() {
			logger.WithFields(map[string]interface{}{
				"key":    key,
				"path":   c.Path(),
				"method": c.Method(),
			}).Warn("Rate limit exceeded")

			// Set rate limit headers
			c.Set("X-RateLimit-Limit", fmt.Sprintf("%.0f", config.RPS))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("Retry-After", "1")

			if config.ErrorHandler != nil {
				return config.ErrorHandler(c)
			}

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":   "Rate limit exceeded",
				"limit":   config.RPS,
				"message": "Too many requests, please slow down",
			})
		}

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%.0f", config.RPS))

		return c.Next()
	}
}

// EndpointRateLimit creates a per-endpoint rate limiting middleware
type EndpointRateLimitConfig struct {
	// Limits per endpoint pattern (e.g., "/api/v1/cases": {RPS: 10, Burst: 20})
	Limits map[string]*RateLimitConfig
	// Default limit for endpoints not in the map
	DefaultLimit *RateLimitConfig
	// Whether to use IP-based or client-based limiting
	UseClientID bool
}

// DefaultEndpointRateLimitConfig returns a default endpoint rate limit configuration
func DefaultEndpointRateLimitConfig() *EndpointRateLimitConfig {
	return &EndpointRateLimitConfig{
		Limits: map[string]*RateLimitConfig{
			"/api/v1/cases":     {RPS: 10, Burst: 20},
			"/api/v1/judges":    {RPS: 15, Burst: 30},
			"/api/v1/citations": {RPS: 20, Burst: 40},
			"/api/v1/search":    {RPS: 5, Burst: 10},
		},
		DefaultLimit: &RateLimitConfig{
			RPS:   10,
			Burst: 20,
		},
		UseClientID: false,
	}
}

// EndpointRateLimit creates a per-endpoint rate limiting middleware
func EndpointRateLimit(config *EndpointRateLimitConfig, logger *observability.Logger) fiber.Handler {
	if config == nil {
		config = DefaultEndpointRateLimitConfig()
	}

	// Initialize storage for each endpoint
	storages := make(map[string]RateLimitStorage)
	for endpoint, limitConfig := range config.Limits {
		storages[endpoint] = NewInMemoryRateLimitStorage(limitConfig.RPS, limitConfig.Burst)
	}

	// Default storage
	var defaultStorage RateLimitStorage
	if config.DefaultLimit != nil {
		defaultStorage = NewInMemoryRateLimitStorage(config.DefaultLimit.RPS, config.DefaultLimit.Burst)
	}

	return func(c *fiber.Ctx) error {
		path := c.Path()

		// Find matching endpoint config
		var limitConfig *RateLimitConfig
		var storage RateLimitStorage

		if cfg, exists := config.Limits[path]; exists {
			limitConfig = cfg
			storage = storages[path]
		} else {
			limitConfig = config.DefaultLimit
			storage = defaultStorage
		}

		if limitConfig == nil {
			// No rate limiting for this endpoint
			return c.Next()
		}

		// Generate key
		key := c.IP()
		if config.UseClientID {
			if clientID, ok := c.Locals("client_id").(string); ok && clientID != "" {
				key = clientID
			}
		}

		// Get or create limiter
		limiter := storage.Get(key)

		// Check if request is allowed
		if !limiter.Allow() {
			logger.WithFields(map[string]interface{}{
				"key":      key,
				"path":     path,
				"method":   c.Method(),
				"limit":    limitConfig.RPS,
				"endpoint": path,
			}).Warn("Endpoint rate limit exceeded")

			c.Set("X-RateLimit-Limit", fmt.Sprintf("%.0f", limitConfig.RPS))
			c.Set("X-RateLimit-Remaining", "0")
			c.Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))
			c.Set("Retry-After", "1")

			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":    "Rate limit exceeded for this endpoint",
				"limit":    limitConfig.RPS,
				"endpoint": path,
				"message":  "Too many requests to this endpoint, please slow down",
			})
		}

		// Set rate limit headers
		c.Set("X-RateLimit-Limit", fmt.Sprintf("%.0f", limitConfig.RPS))

		return c.Next()
	}
}

// IPRateLimit creates a simple IP-based rate limiter
func IPRateLimit(rps float64, burst int, logger *observability.Logger) fiber.Handler {
	return RateLimit(&RateLimitConfig{
		RPS:   rps,
		Burst: burst,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}, logger)
}

// ClientRateLimit creates a client-ID-based rate limiter
func ClientRateLimit(rps float64, burst int, logger *observability.Logger) fiber.Handler {
	return RateLimit(&RateLimitConfig{
		RPS:   rps,
		Burst: burst,
		KeyGenerator: func(c *fiber.Ctx) string {
			if clientID, ok := c.Locals("client_id").(string); ok && clientID != "" {
				return clientID
			}
			return c.IP()
		},
	}, logger)
}
