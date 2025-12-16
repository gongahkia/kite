package scraper

import (
	"context"
	"time"

	"github.com/gongahkia/kite/pkg/models"
)

// Scraper is the interface that all jurisdiction-specific scrapers must implement
type Scraper interface {
	// GetName returns the name of the scraper/database
	GetName() string

	// GetJurisdiction returns the jurisdiction covered by this scraper
	GetJurisdiction() string

	// SearchCases searches for cases matching the query
	SearchCases(ctx context.Context, query SearchQuery) ([]*models.Case, error)

	// GetCaseByID retrieves a specific case by its ID
	GetCaseByID(ctx context.Context, caseID string) (*models.Case, error)

	// GetCasesByDateRange retrieves cases within a date range
	GetCasesByDateRange(ctx context.Context, startDate, endDate time.Time, limit int) ([]*models.Case, error)

	// IsAvailable checks if the scraper's data source is available
	IsAvailable(ctx context.Context) bool

	// GetRateLimit returns the rate limit for this scraper (requests per minute)
	GetRateLimit() int

	// GetMetadata returns metadata about this scraper
	GetMetadata() ScraperMetadata
}

// SearchQuery represents a search query for cases
type SearchQuery struct {
	Query       string     `json:"query"`
	Jurisdiction string    `json:"jurisdiction,omitempty"`
	Court       string     `json:"court,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
	Judges      []string   `json:"judges,omitempty"`
	Concepts    []string   `json:"concepts,omitempty"`
	CourtLevel  *models.CourtLevel `json:"court_level,omitempty"`
}

// ScraperMetadata contains metadata about a scraper
type ScraperMetadata struct {
	Name          string              `json:"name"`
	Jurisdiction  string              `json:"jurisdiction"`
	Coverage      string              `json:"coverage"`
	BaseURL       string              `json:"base_url"`
	Status        ScraperStatus       `json:"status"`
	RateLimit     int                 `json:"rate_limit"`
	Features      []string            `json:"features"`
	LastChecked   time.Time           `json:"last_checked"`
}

// ScraperStatus represents the status of a scraper
type ScraperStatus string

const (
	ScraperStatusActive      ScraperStatus = "active"
	ScraperStatusDegraded    ScraperStatus = "degraded"
	ScraperStatusUnavailable ScraperStatus = "unavailable"
	ScraperStatusMaintenance ScraperStatus = "maintenance"
)

// BaseScraper provides common functionality for all scrapers
type BaseScraper struct {
	name         string
	jurisdiction string
	baseURL      string
	rateLimit    int
	client       *ScraperHTTPClient
	logger       interface{}
	metrics      interface{}
}

// NewBaseScraper creates a new BaseScraper
func NewBaseScraper(name, jurisdiction, baseURL string, rateLimit int) *BaseScraper {
	return &BaseScraper{
		name:         name,
		jurisdiction: jurisdiction,
		baseURL:      baseURL,
		rateLimit:    rateLimit,
		client:       NewScraperHTTPClient(baseURL, rateLimit),
	}
}

// GetName returns the scraper name
func (bs *BaseScraper) GetName() string {
	return bs.name
}

// GetJurisdiction returns the jurisdiction
func (bs *BaseScraper) GetJurisdiction() string {
	return bs.jurisdiction
}

// GetRateLimit returns the rate limit
func (bs *BaseScraper) GetRateLimit() int {
	return bs.rateLimit
}

// GetMetadata returns scraper metadata
func (bs *BaseScraper) GetMetadata() ScraperMetadata {
	return ScraperMetadata{
		Name:         bs.name,
		Jurisdiction: bs.jurisdiction,
		BaseURL:      bs.baseURL,
		RateLimit:    bs.rateLimit,
		Status:       ScraperStatusActive,
		LastChecked:  time.Now(),
	}
}

// ScraperHTTPClient is a specialized HTTP client for scraping
type ScraperHTTPClient struct {
	baseURL     string
	rateLimit   int
	rateLimiter *RateLimiter
	robotsCache *RobotsCache
	timeout     time.Duration
	userAgent   string
}

// NewScraperHTTPClient creates a new ScraperHTTPClient
func NewScraperHTTPClient(baseURL string, rateLimit int) *ScraperHTTPClient {
	return &ScraperHTTPClient{
		baseURL:     baseURL,
		rateLimit:   rateLimit,
		rateLimiter: NewRateLimiter(rateLimit),
		robotsCache: NewRobotsCache(),
		timeout:     30 * time.Second,
		userAgent:   "Kite/4.0 (Legal Research Bot; +https://github.com/gongahkia/kite)",
	}
}

// SetUserAgent sets the user agent for the client
func (sc *ScraperHTTPClient) SetUserAgent(userAgent string) {
	sc.userAgent = userAgent
}

// SetTimeout sets the request timeout
func (sc *ScraperHTTPClient) SetTimeout(timeout time.Duration) {
	sc.timeout = timeout
}

// CheckRobots checks if scraping is allowed by robots.txt
func (sc *ScraperHTTPClient) CheckRobots(ctx context.Context, path string) (bool, error) {
	return sc.robotsCache.IsAllowed(ctx, sc.baseURL, path, sc.userAgent)
}

// ScraperConfig represents configuration for a scraper
type ScraperConfig struct {
	Name             string        `yaml:"name"`
	Jurisdiction     string        `yaml:"jurisdiction"`
	BaseURL          string        `yaml:"base_url"`
	RateLimit        int           `yaml:"rate_limit"`
	Timeout          time.Duration `yaml:"timeout"`
	MaxRetries       int           `yaml:"max_retries"`
	RespectRobotsTxt bool          `yaml:"respect_robots_txt"`
	Enabled          bool          `yaml:"enabled"`
}

// ScraperRegistry manages all available scrapers
type ScraperRegistry struct {
	scrapers map[string]Scraper
}

// NewScraperRegistry creates a new ScraperRegistry
func NewScraperRegistry() *ScraperRegistry {
	return &ScraperRegistry{
		scrapers: make(map[string]Scraper),
	}
}

// Register registers a scraper
func (sr *ScraperRegistry) Register(name string, scraper Scraper) {
	sr.scrapers[name] = scraper
}

// Get retrieves a scraper by name
func (sr *ScraperRegistry) Get(name string) (Scraper, bool) {
	scraper, ok := sr.scrapers[name]
	return scraper, ok
}

// GetAll returns all registered scrapers
func (sr *ScraperRegistry) GetAll() map[string]Scraper {
	return sr.scrapers
}

// GetByJurisdiction returns all scrapers for a jurisdiction
func (sr *ScraperRegistry) GetByJurisdiction(jurisdiction string) []Scraper {
	var result []Scraper
	for _, scraper := range sr.scrapers {
		if scraper.GetJurisdiction() == jurisdiction {
			result = append(result, scraper)
		}
	}
	return result
}
