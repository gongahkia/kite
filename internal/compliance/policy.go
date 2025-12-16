package compliance

import (
	"sync"
	"time"
)

// PolicyManager manages scraping policies and terms of service for different sources
type PolicyManager struct {
	policies map[string]*ScrapingPolicy
	mu       sync.RWMutex
}

// ScrapingPolicy represents the scraping policy for a legal database
type ScrapingPolicy struct {
	SourceName        string          `yaml:"source_name" json:"source_name"`
	BaseURL           string          `yaml:"base_url" json:"base_url"`
	AllowScraping     bool            `yaml:"allow_scraping" json:"allow_scraping"`
	RequiresAttribution bool          `yaml:"requires_attribution" json:"requires_attribution"`
	AttributionText   string          `yaml:"attribution_text,omitempty" json:"attribution_text,omitempty"`
	CommercialUse     CommercialUsePolicy `yaml:"commercial_use" json:"commercial_use"`
	RateLimit         int             `yaml:"rate_limit" json:"rate_limit"` // requests per minute
	CrawlDelay        time.Duration   `yaml:"crawl_delay" json:"crawl_delay"` // delay between requests
	TermsOfServiceURL string          `yaml:"terms_of_service_url,omitempty" json:"terms_of_service_url,omitempty"`
	LastChecked       time.Time       `yaml:"last_checked" json:"last_checked"`
	Restrictions      []string        `yaml:"restrictions,omitempty" json:"restrictions,omitempty"`
	BulkDownload      bool            `yaml:"bulk_download" json:"bulk_download"`
	APIAvailable      bool            `yaml:"api_available" json:"api_available"`
	ContactEmail      string          `yaml:"contact_email,omitempty" json:"contact_email,omitempty"`
}

// CommercialUsePolicy represents the commercial use policy
type CommercialUsePolicy string

const (
	CommercialUseAllowed    CommercialUsePolicy = "allowed"
	CommercialUseForbidden  CommercialUsePolicy = "forbidden"
	CommercialUseRestricted CommercialUsePolicy = "restricted"
	CommercialUseUnknown    CommercialUsePolicy = "unknown"
)

// NewPolicyManager creates a new policy manager
func NewPolicyManager() *PolicyManager {
	pm := &PolicyManager{
		policies: make(map[string]*ScrapingPolicy),
	}
	pm.initializeDefaults()
	return pm
}

// initializeDefaults initializes default policies for major legal databases
func (pm *PolicyManager) initializeDefaults() {
	// CourtListener (US)
	pm.RegisterPolicy(&ScrapingPolicy{
		SourceName:        "CourtListener",
		BaseURL:           "https://www.courtlistener.com",
		AllowScraping:     true,
		RequiresAttribution: true,
		AttributionText:   "Data sourced from CourtListener, a project of Free Law Project (https://www.courtlistener.com)",
		CommercialUse:     CommercialUseAllowed,
		RateLimit:         60, // 60 requests per minute
		CrawlDelay:        1 * time.Second,
		TermsOfServiceURL: "https://www.courtlistener.com/terms/",
		LastChecked:       time.Now(),
		BulkDownload:      true,
		APIAvailable:      true,
		ContactEmail:      "info@free.law",
		Restrictions:      []string{"Must credit Free Law Project"},
	})

	// CanLII (Canada)
	pm.RegisterPolicy(&ScrapingPolicy{
		SourceName:        "CanLII",
		BaseURL:           "https://www.canlii.org",
		AllowScraping:     true,
		RequiresAttribution: true,
		AttributionText:   "Data sourced from CanLII (https://www.canlii.org)",
		CommercialUse:     CommercialUseAllowed,
		RateLimit:         30,
		CrawlDelay:        2 * time.Second,
		TermsOfServiceURL: "https://www.canlii.org/en/info/terms.html",
		LastChecked:       time.Now(),
		BulkDownload:      false,
		APIAvailable:      true,
		ContactEmail:      "info@canlii.org",
		Restrictions:      []string{"Must include citation", "Non-commercial preferred"},
	})

	// BAILII (UK/Ireland)
	pm.RegisterPolicy(&ScrapingPolicy{
		SourceName:        "BAILII",
		BaseURL:           "https://www.bailii.org",
		AllowScraping:     true,
		RequiresAttribution: true,
		AttributionText:   "Data sourced from BAILII (British and Irish Legal Information Institute) (https://www.bailii.org)",
		CommercialUse:     CommercialUseRestricted,
		RateLimit:         12,
		CrawlDelay:        5 * time.Second,
		TermsOfServiceURL: "https://www.bailii.org/bailii/legal_policy.html",
		LastChecked:       time.Now(),
		BulkDownload:      false,
		APIAvailable:      false,
		Restrictions:      []string{"Must credit BAILII", "Commercial use requires permission"},
	})

	// AustLII (Australia)
	pm.RegisterPolicy(&ScrapingPolicy{
		SourceName:        "AustLII",
		BaseURL:           "https://www.austlii.edu.au",
		AllowScraping:     true,
		RequiresAttribution: true,
		AttributionText:   "Data sourced from AustLII (Australasian Legal Information Institute) (https://www.austlii.edu.au)",
		CommercialUse:     CommercialUseRestricted,
		RateLimit:         12,
		CrawlDelay:        5 * time.Second,
		TermsOfServiceURL: "https://www.austlii.edu.au/austlii/legal_policy.html",
		LastChecked:       time.Now(),
		BulkDownload:      false,
		APIAvailable:      false,
		Restrictions:      []string{"Must credit AustLII", "Non-commercial preferred"},
	})

	// HKLII (Hong Kong)
	pm.RegisterPolicy(&ScrapingPolicy{
		SourceName:        "HKLII",
		BaseURL:           "https://www.hklii.hk",
		AllowScraping:     true,
		RequiresAttribution: true,
		AttributionText:   "Data sourced from HKLII (Hong Kong Legal Information Institute) (https://www.hklii.hk)",
		CommercialUse:     CommercialUseRestricted,
		RateLimit:         10,
		CrawlDelay:        6 * time.Second,
		TermsOfServiceURL: "https://www.hklii.hk/legal_policy.html",
		LastChecked:       time.Now(),
		BulkDownload:      false,
		APIAvailable:      false,
		Restrictions:      []string{"Must credit HKLII", "Respect robots.txt"},
	})

	// Indian Kanoon (India)
	pm.RegisterPolicy(&ScrapingPolicy{
		SourceName:        "IndianKanoon",
		BaseURL:           "https://indiankanoon.org",
		AllowScraping:     true,
		RequiresAttribution: true,
		AttributionText:   "Data sourced from Indian Kanoon (https://indiankanoon.org)",
		CommercialUse:     CommercialUseRestricted,
		RateLimit:         10,
		CrawlDelay:        6 * time.Second,
		TermsOfServiceURL: "https://indiankanoon.org/terms.html",
		LastChecked:       time.Now(),
		BulkDownload:      false,
		APIAvailable:      false,
		ContactEmail:      "contact@indiankanoon.org",
		Restrictions:      []string{"Moderate use only", "Must credit Indian Kanoon"},
	})

	// Add more jurisdictions as needed
}

// RegisterPolicy registers a scraping policy
func (pm *PolicyManager) RegisterPolicy(policy *ScrapingPolicy) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.policies[policy.SourceName] = policy
}

// GetPolicy retrieves the policy for a source
func (pm *PolicyManager) GetPolicy(sourceName string) (*ScrapingPolicy, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	policy, ok := pm.policies[sourceName]
	return policy, ok
}

// CheckCompliance checks if a scraping operation complies with the policy
func (pm *PolicyManager) CheckCompliance(sourceName string, requestsPerMinute int) (bool, []string) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	policy, ok := pm.policies[sourceName]
	if !ok {
		return false, []string{"No policy found for source: " + sourceName}
	}

	var violations []string

	// Check if scraping is allowed
	if !policy.AllowScraping {
		violations = append(violations, "Scraping is not allowed for this source")
		return false, violations
	}

	// Check rate limit
	if requestsPerMinute > policy.RateLimit {
		violations = append(violations, "Rate limit exceeded")
	}

	return len(violations) == 0, violations
}

// GetAttributionText returns the attribution text for a source
func (pm *PolicyManager) GetAttributionText(sourceName string) string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	policy, ok := pm.policies[sourceName]
	if !ok || !policy.RequiresAttribution {
		return ""
	}

	return policy.AttributionText
}

// IsCommercialUseAllowed checks if commercial use is allowed
func (pm *PolicyManager) IsCommercialUseAllowed(sourceName string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	policy, ok := pm.policies[sourceName]
	if !ok {
		return false
	}

	return policy.CommercialUse == CommercialUseAllowed
}

// GetCrawlDelay returns the recommended crawl delay for a source
func (pm *PolicyManager) GetCrawlDelay(sourceName string) time.Duration {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	policy, ok := pm.policies[sourceName]
	if !ok {
		return 5 * time.Second // Default delay
	}

	return policy.CrawlDelay
}

// GetAllPolicies returns all registered policies
func (pm *PolicyManager) GetAllPolicies() []*ScrapingPolicy {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	policies := make([]*ScrapingPolicy, 0, len(pm.policies))
	for _, policy := range pm.policies {
		policies = append(policies, policy)
	}

	return policies
}

// UpdatePolicy updates an existing policy
func (pm *PolicyManager) UpdatePolicy(policy *ScrapingPolicy) {
	pm.mu.Lock()
	defer pm.mu.Unlock()
	policy.LastChecked = time.Now()
	pm.policies[policy.SourceName] = policy
}

// PolicyViolation represents a policy violation event
type PolicyViolation struct {
	SourceName  string    `json:"source_name"`
	ViolationType string  `json:"violation_type"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
	Severity    string    `json:"severity"` // "low", "medium", "high", "critical"
}

// ViolationTracker tracks policy violations
type ViolationTracker struct {
	violations []PolicyViolation
	mu         sync.RWMutex
}

// NewViolationTracker creates a new violation tracker
func NewViolationTracker() *ViolationTracker {
	return &ViolationTracker{
		violations: make([]PolicyViolation, 0),
	}
}

// RecordViolation records a policy violation
func (vt *ViolationTracker) RecordViolation(violation PolicyViolation) {
	vt.mu.Lock()
	defer vt.mu.Unlock()
	violation.Timestamp = time.Now()
	vt.violations = append(vt.violations, violation)
}

// GetViolations retrieves all violations for a source
func (vt *ViolationTracker) GetViolations(sourceName string) []PolicyViolation {
	vt.mu.RLock()
	defer vt.mu.RUnlock()

	var result []PolicyViolation
	for _, v := range vt.violations {
		if v.SourceName == sourceName {
			result = append(result, v)
		}
	}

	return result
}

// GetRecentViolations retrieves violations within a time period
func (vt *ViolationTracker) GetRecentViolations(since time.Duration) []PolicyViolation {
	vt.mu.RLock()
	defer vt.mu.RUnlock()

	cutoff := time.Now().Add(-since)
	var result []PolicyViolation
	for _, v := range vt.violations {
		if v.Timestamp.After(cutoff) {
			result = append(result, v)
		}
	}

	return result
}

// ClearViolations clears all violations for a source
func (vt *ViolationTracker) ClearViolations(sourceName string) {
	vt.mu.Lock()
	defer vt.mu.Unlock()

	newViolations := make([]PolicyViolation, 0)
	for _, v := range vt.violations {
		if v.SourceName != sourceName {
			newViolations = append(newViolations, v)
		}
	}

	vt.violations = newViolations
}
