package scraper

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter implements rate limiting using token bucket algorithm
type RateLimiter struct {
	limiter *rate.Limiter
	mu      sync.Mutex
}

// NewRateLimiter creates a new RateLimiter
// requestsPerMinute: number of requests allowed per minute
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	// Convert requests per minute to requests per second
	requestsPerSecond := float64(requestsPerMinute) / 60.0

	// Create limiter with burst capacity equal to rate per second (allow small bursts)
	burst := int(requestsPerSecond) + 1
	if burst < 1 {
		burst = 1
	}

	return &RateLimiter{
		limiter: rate.NewLimiter(rate.Limit(requestsPerSecond), burst),
	}
}

// Wait blocks until request is allowed under rate limit
func (rl *RateLimiter) Wait(ctx context.Context) error {
	return rl.limiter.Wait(ctx)
}

// Allow returns true if request is allowed immediately
func (rl *RateLimiter) Allow() bool {
	return rl.limiter.Allow()
}

// Reserve reserves a request and returns a Reservation
func (rl *RateLimiter) Reserve() *rate.Reservation {
	return rl.limiter.Reserve()
}

// SetLimit updates the rate limit
func (rl *RateLimiter) SetLimit(requestsPerMinute int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	requestsPerSecond := float64(requestsPerMinute) / 60.0
	burst := int(requestsPerSecond) + 1
	if burst < 1 {
		burst = 1
	}

	rl.limiter.SetLimit(rate.Limit(requestsPerSecond))
	rl.limiter.SetBurst(burst)
}

// MultiRateLimiter manages multiple rate limiters for different sources
type MultiRateLimiter struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
}

// NewMultiRateLimiter creates a new MultiRateLimiter
func NewMultiRateLimiter() *MultiRateLimiter {
	return &MultiRateLimiter{
		limiters: make(map[string]*RateLimiter),
	}
}

// AddLimiter adds a rate limiter for a specific source
func (mrl *MultiRateLimiter) AddLimiter(source string, requestsPerMinute int) {
	mrl.mu.Lock()
	defer mrl.mu.Unlock()

	mrl.limiters[source] = NewRateLimiter(requestsPerMinute)
}

// Wait blocks until request is allowed for the given source
func (mrl *MultiRateLimiter) Wait(ctx context.Context, source string) error {
	mrl.mu.RLock()
	limiter, ok := mrl.limiters[source]
	mrl.mu.RUnlock()

	if !ok {
		// No limiter for this source, allow immediately
		return nil
	}

	return limiter.Wait(ctx)
}

// Allow returns true if request is allowed immediately for the given source
func (mrl *MultiRateLimiter) Allow(source string) bool {
	mrl.mu.RLock()
	limiter, ok := mrl.limiters[source]
	mrl.mu.RUnlock()

	if !ok {
		// No limiter for this source, allow immediately
		return true
	}

	return limiter.Allow()
}

// UpdateLimit updates the rate limit for a source
func (mrl *MultiRateLimiter) UpdateLimit(source string, requestsPerMinute int) {
	mrl.mu.RLock()
	limiter, ok := mrl.limiters[source]
	mrl.mu.RUnlock()

	if ok {
		limiter.SetLimit(requestsPerMinute)
	} else {
		mrl.AddLimiter(source, requestsPerMinute)
	}
}

// DomainRateLimiter manages rate limits per domain
type DomainRateLimiter struct {
	limiters map[string]*RateLimiter
	mu       sync.RWMutex
	defaultLimit int
}

// NewDomainRateLimiter creates a new DomainRateLimiter
func NewDomainRateLimiter(defaultRequestsPerMinute int) *DomainRateLimiter {
	return &DomainRateLimiter{
		limiters:     make(map[string]*RateLimiter),
		defaultLimit: defaultRequestsPerMinute,
	}
}

// Wait blocks until request is allowed for the given domain
func (drl *DomainRateLimiter) Wait(ctx context.Context, domain string) error {
	limiter := drl.getLimiter(domain)
	return limiter.Wait(ctx)
}

// Allow returns true if request is allowed immediately for the given domain
func (drl *DomainRateLimiter) Allow(domain string) bool {
	limiter := drl.getLimiter(domain)
	return limiter.Allow()
}

// SetDomainLimit sets a specific rate limit for a domain
func (drl *DomainRateLimiter) SetDomainLimit(domain string, requestsPerMinute int) {
	drl.mu.Lock()
	defer drl.mu.Unlock()

	if limiter, ok := drl.limiters[domain]; ok {
		limiter.SetLimit(requestsPerMinute)
	} else {
		drl.limiters[domain] = NewRateLimiter(requestsPerMinute)
	}
}

// getLimiter gets or creates a limiter for a domain
func (drl *DomainRateLimiter) getLimiter(domain string) *RateLimiter {
	drl.mu.RLock()
	limiter, ok := drl.limiters[domain]
	drl.mu.RUnlock()

	if ok {
		return limiter
	}

	drl.mu.Lock()
	defer drl.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, ok := drl.limiters[domain]; ok {
		return limiter
	}

	limiter = NewRateLimiter(drl.defaultLimit)
	drl.limiters[domain] = limiter
	return limiter
}

// ClearDomain removes the rate limiter for a domain
func (drl *DomainRateLimiter) ClearDomain(domain string) {
	drl.mu.Lock()
	defer drl.mu.Unlock()

	delete(drl.limiters, domain)
}

// GetStats returns statistics about rate limiting
type RateLimitStats struct {
	Domain           string    `json:"domain"`
	RequestsAllowed  int64     `json:"requests_allowed"`
	RequestsDenied   int64     `json:"requests_denied"`
	LastRequest      time.Time `json:"last_request"`
	CurrentRate      float64   `json:"current_rate"`
}
