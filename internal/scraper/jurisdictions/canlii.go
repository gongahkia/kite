package jurisdictions

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gongahkia/kite/internal/scraper"
	"github.com/gongahkia/kite/pkg/errors"
	"github.com/gongahkia/kite/pkg/models"
)

// CanLIIScraper implements scraping for CanLII (Canadian Legal Information Institute)
type CanLIIScraper struct {
	*scraper.BaseScraper
	baseURL string
	client  *http.Client
}

// NewCanLIIScraper creates a new CanLII scraper
func NewCanLIIScraper() *CanLIIScraper {
	baseURL := "https://www.canlii.org"
	base := scraper.NewBaseScraper(
		"CanLII",
		"Canada",
		baseURL,
		15, // 15 requests per minute to be conservative
	)

	return &CanLIIScraper{
		BaseScraper: base,
		baseURL:     baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchCases searches for cases matching the query
func (cs *CanLIIScraper) SearchCases(ctx context.Context, query scraper.SearchQuery) ([]*models.Case, error) {
	// Build search URL
	searchURL, err := cs.buildSearchURL(query)
	if err != nil {
		return nil, errors.ParsingError("failed to build search URL", err)
	}

	// Check robots.txt
	allowed, err := cs.BaseScraper.client.CheckRobots(ctx, "/en/search/")
	if err != nil || !allowed {
		return nil, errors.ErrRobotsDisallowed
	}

	// Wait for rate limit
	if err := cs.BaseScraper.client.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.RateLimitError("rate limit exceeded")
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, errors.NetworkError("failed to create request", err)
	}

	req.Header.Set("User-Agent", cs.BaseScraper.client.userAgent)

	resp, err := cs.client.Do(req)
	if err != nil {
		return nil, errors.NetworkError("failed to fetch search results", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NetworkError(fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, errors.ParsingError("failed to parse HTML", err)
	}

	// Extract cases from search results
	cases := make([]*models.Case, 0)

	doc.Find(".result").Each(func(i int, s *goquery.Selection) {
		if query.Limit > 0 && len(cases) >= query.Limit {
			return
		}

		caseData := cs.extractCaseFromSearchResult(s)
		if caseData != nil {
			cases = append(cases, caseData)
		}
	})

	return cases, nil
}

// GetCaseByID retrieves a specific case by its ID (e.g., "2023scc15")
func (cs *CanLIIScraper) GetCaseByID(ctx context.Context, caseID string) (*models.Case, error) {
	// CanLII case IDs are typically in format: 2023scc15
	// Parse to determine court and build URL
	caseURL := cs.buildCaseURL(caseID)

	// Check robots.txt
	allowed, err := cs.BaseScraper.client.CheckRobots(ctx, "/en/ca/")
	if err != nil || !allowed {
		return nil, errors.ErrRobotsDisallowed
	}

	// Wait for rate limit
	if err := cs.BaseScraper.client.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.RateLimitError("rate limit exceeded")
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", caseURL, nil)
	if err != nil {
		return nil, errors.NetworkError("failed to create request", err)
	}

	req.Header.Set("User-Agent", cs.BaseScraper.client.userAgent)

	resp, err := cs.client.Do(req)
	if err != nil {
		return nil, errors.NetworkError("failed to fetch case", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.ErrNotFound
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NetworkError(fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, errors.ParsingError("failed to parse HTML", err)
	}

	// Extract case details
	return cs.extractCaseDetails(doc, caseID, caseURL)
}

// GetCasesByDateRange retrieves cases within a date range
func (cs *CanLIIScraper) GetCasesByDateRange(ctx context.Context, startDate, endDate time.Time, limit int) ([]*models.Case, error) {
	query := scraper.SearchQuery{
		StartDate: &startDate,
		EndDate:   &endDate,
		Limit:     limit,
	}

	return cs.SearchCases(ctx, query)
}

// IsAvailable checks if CanLII is available
func (cs *CanLIIScraper) IsAvailable(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", cs.baseURL, nil)
	if err != nil {
		return false
	}

	resp, err := cs.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// buildSearchURL builds the search URL with query parameters
func (cs *CanLIIScraper) buildSearchURL(query scraper.SearchQuery) (string, error) {
	params := url.Values{}

	if query.Query != "" {
		params.Set("search", query.Query)
	}

	if query.StartDate != nil && query.EndDate != nil {
		dateRange := fmt.Sprintf("%s:%s",
			query.StartDate.Format("2006-01-02"),
			query.EndDate.Format("2006-01-02"))
		params.Set("dateRange", dateRange)
	}

	searchURL := fmt.Sprintf("%s/en/search/?%s", cs.baseURL, params.Encode())
	return searchURL, nil
}

// buildCaseURL builds a case URL from a case citation (e.g., "2023scc15")
func (cs *CanLIIScraper) buildCaseURL(caseID string) string {
	// Simplified - in real implementation would parse the citation properly
	// Example: 2023scc15 -> /en/ca/scc/doc/2023/2023scc15/
	lowerID := strings.ToLower(caseID)

	// Extract year (first 4 digits)
	var year string
	for _, ch := range lowerID {
		if ch >= '0' && ch <= '9' {
			year += string(ch)
			if len(year) == 4 {
				break
			}
		}
	}

	// For now, assume SCC (Supreme Court of Canada) as example
	return fmt.Sprintf("%s/en/ca/scc/doc/%s/%s/", cs.baseURL, year, lowerID)
}

// extractCaseFromSearchResult extracts case data from a search result item
func (cs *CanLIIScraper) extractCaseFromSearchResult(s *goquery.Selection) *models.Case {
	c := models.NewCase()

	// Extract case name
	caseName := s.Find(".resultTitle a").Text()
	c.CaseName = strings.TrimSpace(caseName)

	// Extract case URL and ID
	caseURL, exists := s.Find(".resultTitle a").Attr("href")
	if exists {
		if !strings.HasPrefix(caseURL, "http") {
			caseURL = cs.baseURL + caseURL
		}
		c.URL = caseURL

		// Extract ID from URL
		parts := strings.Split(strings.Trim(caseURL, "/"), "/")
		if len(parts) > 0 {
			c.ID = parts[len(parts)-1]
		}
	}

	// Extract citation
	citation := s.Find(".resultCitation").Text()
	c.CaseNumber = strings.TrimSpace(citation)

	// Extract court and date from metadata
	metadata := s.Find(".resultMeta").Text()
	c.Summary = strings.TrimSpace(metadata)

	// Set basic metadata
	c.Jurisdiction = "Canada"
	c.SourceDatabase = "CanLII"
	c.ScrapedAt = time.Now()
	c.LastUpdated = time.Now()
	c.Language = "en"
	c.Status = models.CaseStatusActive

	return c
}

// extractCaseDetails extracts detailed case information from a case page
func (cs *CanLIIScraper) extractCaseDetails(doc *goquery.Document, caseID, caseURL string) (*models.Case, error) {
	c := models.NewCase()

	c.ID = caseID
	c.URL = caseURL

	// Extract case name
	caseName := doc.Find("h1.documentTitle").Text()
	c.CaseName = strings.TrimSpace(caseName)

	// Extract citation
	citation := doc.Find(".documentCitation").Text()
	c.CaseNumber = strings.TrimSpace(citation)

	// Extract court
	court := doc.Find(".court").Text()
	c.Court = strings.TrimSpace(court)

	// Extract date
	dateStr := doc.Find(".documentDate").Text()
	if dateStr != "" {
		// Try to parse various date formats
		formats := []string{"2006-01-02", "January 2, 2006", "02-01-2006"}
		for _, format := range formats {
			if date, err := time.Parse(format, strings.TrimSpace(dateStr)); err == nil {
				c.DecisionDate = &date
				break
			}
		}
	}

	// Extract docket number
	docket := doc.Find(".docketNumber").Text()
	c.Docket = strings.TrimSpace(docket)

	// Extract full text
	fullText := doc.Find(".documentContent").Text()
	c.FullText = strings.TrimSpace(fullText)

	// Extract judges
	doc.Find(".judge").Each(func(i int, s *goquery.Selection) {
		judge := strings.TrimSpace(s.Text())
		if judge != "" {
			c.Judges = append(c.Judges, judge)
		}
	})

	// Set metadata
	c.Jurisdiction = "Canada"
	c.SourceDatabase = "CanLII"
	c.ScrapedAt = time.Now()
	c.LastUpdated = time.Now()
	c.Language = "en"
	c.Status = models.CaseStatusActive

	return c, nil
}
