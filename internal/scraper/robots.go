package scraper

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// RobotsCache caches robots.txt files and checks permissions
type RobotsCache struct {
	cache map[string]*RobotsTxt
	mu    sync.RWMutex
	ttl   time.Duration
}

// RobotsTxt represents a parsed robots.txt file
type RobotsTxt struct {
	baseURL       string
	rules         map[string]*RobotRules // user-agent -> rules
	defaultRules  *RobotRules
	crawlDelay    time.Duration
	fetchedAt     time.Time
	expiresAt     time.Time
}

// RobotRules represents rules for a specific user-agent
type RobotRules struct {
	disallowed []string
	allowed    []string
	crawlDelay time.Duration
}

// NewRobotsCache creates a new RobotsCache
func NewRobotsCache() *RobotsCache {
	return &RobotsCache{
		cache: make(map[string]*RobotsTxt),
		ttl:   24 * time.Hour, // Cache for 24 hours
	}
}

// SetTTL sets the cache TTL
func (rc *RobotsCache) SetTTL(ttl time.Duration) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.ttl = ttl
}

// IsAllowed checks if the given path is allowed for the user-agent
func (rc *RobotsCache) IsAllowed(ctx context.Context, baseURL, path, userAgent string) (bool, error) {
	// Get or fetch robots.txt
	robots, err := rc.getRobotsTxt(ctx, baseURL)
	if err != nil {
		// If robots.txt cannot be fetched, allow access by default
		return true, nil
	}

	// Check if expired
	if time.Now().After(robots.expiresAt) {
		rc.mu.Lock()
		delete(rc.cache, baseURL)
		rc.mu.Unlock()

		// Re-fetch
		robots, err = rc.getRobotsTxt(ctx, baseURL)
		if err != nil {
			return true, nil
		}
	}

	return robots.IsAllowed(path, userAgent), nil
}

// GetCrawlDelay returns the crawl delay for the given user-agent
func (rc *RobotsCache) GetCrawlDelay(ctx context.Context, baseURL, userAgent string) (time.Duration, error) {
	robots, err := rc.getRobotsTxt(ctx, baseURL)
	if err != nil {
		return 0, err
	}

	return robots.GetCrawlDelay(userAgent), nil
}

// getRobotsTxt gets robots.txt from cache or fetches it
func (rc *RobotsCache) getRobotsTxt(ctx context.Context, baseURL string) (*RobotsTxt, error) {
	// Check cache
	rc.mu.RLock()
	robots, ok := rc.cache[baseURL]
	rc.mu.RUnlock()

	if ok && time.Now().Before(robots.expiresAt) {
		return robots, nil
	}

	// Fetch robots.txt
	robots, err := rc.fetchRobotsTxt(ctx, baseURL)
	if err != nil {
		return nil, err
	}

	// Cache it
	rc.mu.Lock()
	rc.cache[baseURL] = robots
	rc.mu.Unlock()

	return robots, nil
}

// fetchRobotsTxt fetches and parses robots.txt from the given base URL
func (rc *RobotsCache) fetchRobotsTxt(ctx context.Context, baseURL string) (*RobotsTxt, error) {
	// Parse base URL
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}

	// Construct robots.txt URL
	robotsURL := fmt.Sprintf("%s://%s/robots.txt", u.Scheme, u.Host)

	// Fetch robots.txt
	req, err := http.NewRequestWithContext(ctx, "GET", robotsURL, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// If not found, return empty rules (allow all)
	if resp.StatusCode == 404 {
		return &RobotsTxt{
			baseURL:      baseURL,
			rules:        make(map[string]*RobotRules),
			defaultRules: &RobotRules{},
			fetchedAt:    time.Now(),
			expiresAt:    time.Now().Add(rc.ttl),
		}, nil
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("failed to fetch robots.txt: status %d", resp.StatusCode)
	}

	// Parse robots.txt
	return parseRobotsTxt(baseURL, resp.Body, rc.ttl)
}

// parseRobotsTxt parses a robots.txt file
func parseRobotsTxt(baseURL string, body interface{ Read([]byte) (int, error) }, ttl time.Duration) (*RobotsTxt, error) {
	robots := &RobotsTxt{
		baseURL:   baseURL,
		rules:     make(map[string]*RobotRules),
		fetchedAt: time.Now(),
		expiresAt: time.Now().Add(ttl),
	}

	scanner := bufio.NewScanner(body)
	var currentUserAgent string
	var currentRules *RobotRules

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split by first colon
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		field := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch field {
		case "user-agent":
			currentUserAgent = strings.ToLower(value)
			if _, ok := robots.rules[currentUserAgent]; !ok {
				robots.rules[currentUserAgent] = &RobotRules{
					disallowed: []string{},
					allowed:    []string{},
				}
			}
			currentRules = robots.rules[currentUserAgent]

		case "disallow":
			if currentRules != nil && value != "" {
				currentRules.disallowed = append(currentRules.disallowed, value)
			}

		case "allow":
			if currentRules != nil && value != "" {
				currentRules.allowed = append(currentRules.allowed, value)
			}

		case "crawl-delay":
			if currentRules != nil {
				var delay float64
				fmt.Sscanf(value, "%f", &delay)
				currentRules.crawlDelay = time.Duration(delay * float64(time.Second))
				robots.crawlDelay = currentRules.crawlDelay
			}
		}
	}

	// Set default rules
	if defaultRules, ok := robots.rules["*"]; ok {
		robots.defaultRules = defaultRules
	} else {
		robots.defaultRules = &RobotRules{}
	}

	return robots, scanner.Err()
}

// IsAllowed checks if a path is allowed for the user-agent
func (rt *RobotsTxt) IsAllowed(path, userAgent string) bool {
	userAgent = strings.ToLower(userAgent)

	// Get rules for this user-agent
	var rules *RobotRules
	if r, ok := rt.rules[userAgent]; ok {
		rules = r
	} else {
		// Try to find a matching wildcard
		for ua, r := range rt.rules {
			if ua == "*" || strings.Contains(userAgent, ua) {
				rules = r
				break
			}
		}
	}

	// Use default rules if no specific rules found
	if rules == nil {
		rules = rt.defaultRules
	}

	// Check allowed rules first (they take precedence)
	for _, pattern := range rules.allowed {
		if matchesPattern(path, pattern) {
			return true
		}
	}

	// Check disallowed rules
	for _, pattern := range rules.disallowed {
		if matchesPattern(path, pattern) {
			return false
		}
	}

	// Default to allowed
	return true
}

// GetCrawlDelay returns the crawl delay for the user-agent
func (rt *RobotsTxt) GetCrawlDelay(userAgent string) time.Duration {
	userAgent = strings.ToLower(userAgent)

	if rules, ok := rt.rules[userAgent]; ok && rules.crawlDelay > 0 {
		return rules.crawlDelay
	}

	// Return default crawl delay
	return rt.crawlDelay
}

// matchesPattern checks if a path matches a robots.txt pattern
func matchesPattern(path, pattern string) bool {
	// Handle exact match
	if path == pattern {
		return true
	}

	// Handle prefix match
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}

	// Handle suffix match
	if strings.HasPrefix(pattern, "*") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(path, suffix)
	}

	// Handle wildcard in middle
	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.HasPrefix(path, parts[0]) && strings.HasSuffix(path, parts[1])
		}
	}

	// No wildcard, check prefix
	return strings.HasPrefix(path, pattern)
}

// Clear clears the robots.txt cache
func (rc *RobotsCache) Clear() {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	rc.cache = make(map[string]*RobotsTxt)
}

// ClearDomain removes robots.txt for a specific domain from cache
func (rc *RobotsCache) ClearDomain(baseURL string) {
	rc.mu.Lock()
	defer rc.mu.Unlock()
	delete(rc.cache, baseURL)
}
