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

// CourtListenerScraper implements scraping for CourtListener (US Federal & State Courts)
type CourtListenerScraper struct {
	*scraper.BaseScraper
	baseURL string
	client  *http.Client
}

// NewCourtListenerScraper creates a new CourtListener scraper
func NewCourtListenerScraper() *CourtListenerScraper {
	baseURL := "https://www.courtlistener.com"
	base := scraper.NewBaseScraper(
		"CourtListener",
		"United States",
		baseURL,
		20, // 20 requests per minute
	)

	return &CourtListenerScraper{
		BaseScraper: base,
		baseURL:     baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchCases searches for cases matching the query
func (cls *CourtListenerScraper) SearchCases(ctx context.Context, query scraper.SearchQuery) ([]*models.Case, error) {
	// Build search URL
	searchURL, err := cls.buildSearchURL(query)
	if err != nil {
		return nil, errors.ParsingError("failed to build search URL", err)
	}

	// Check robots.txt
	allowed, err := cls.BaseScraper.client.CheckRobots(ctx, "/")
	if err != nil || !allowed {
		return nil, errors.ErrRobotsDisallowed
	}

	// Wait for rate limit
	if err := cls.BaseScraper.client.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.RateLimitError("rate limit exceeded")
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, errors.NetworkError("failed to create request", err)
	}

	req.Header.Set("User-Agent", cls.BaseScraper.client.userAgent)

	resp, err := cls.client.Do(req)
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

	doc.Find("article.search-document").Each(func(i int, s *goquery.Selection) {
		if query.Limit > 0 && len(cases) >= query.Limit {
			return
		}

		caseData := cls.extractCaseFromSearchResult(s)
		if caseData != nil {
			cases = append(cases, caseData)
		}
	})

	return cases, nil
}

// GetCaseByID retrieves a specific case by its ID
func (cls *CourtListenerScraper) GetCaseByID(ctx context.Context, caseID string) (*models.Case, error) {
	// CourtListener uses opinion IDs in URLs
	caseURL := fmt.Sprintf("%s/opinion/%s/", cls.baseURL, caseID)

	// Check robots.txt
	allowed, err := cls.BaseScraper.client.CheckRobots(ctx, "/opinion/")
	if err != nil || !allowed {
		return nil, errors.ErrRobotsDisallowed
	}

	// Wait for rate limit
	if err := cls.BaseScraper.client.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.RateLimitError("rate limit exceeded")
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", caseURL, nil)
	if err != nil {
		return nil, errors.NetworkError("failed to create request", err)
	}

	req.Header.Set("User-Agent", cls.BaseScraper.client.userAgent)

	resp, err := cls.client.Do(req)
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
	return cls.extractCaseDetails(doc, caseID, caseURL)
}

// GetCasesByDateRange retrieves cases within a date range
func (cls *CourtListenerScraper) GetCasesByDateRange(ctx context.Context, startDate, endDate time.Time, limit int) ([]*models.Case, error) {
	query := scraper.SearchQuery{
		StartDate: &startDate,
		EndDate:   &endDate,
		Limit:     limit,
	}

	return cls.SearchCases(ctx, query)
}

// IsAvailable checks if CourtListener is available
func (cls *CourtListenerScraper) IsAvailable(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", cls.baseURL, nil)
	if err != nil {
		return false
	}

	resp, err := cls.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// buildSearchURL builds the search URL with query parameters
func (cls *CourtListenerScraper) buildSearchURL(query scraper.SearchQuery) (string, error) {
	params := url.Values{}

	if query.Query != "" {
		params.Set("q", query.Query)
	}

	if query.Court != "" {
		params.Set("court", query.Court)
	}

	if query.StartDate != nil {
		params.Set("filed_after", query.StartDate.Format("2006-01-02"))
	}

	if query.EndDate != nil {
		params.Set("filed_before", query.EndDate.Format("2006-01-02"))
	}

	params.Set("type", "o") // o = opinions
	params.Set("order_by", "score desc")

	searchURL := fmt.Sprintf("%s/?%s", cls.baseURL, params.Encode())
	return searchURL, nil
}

// extractCaseFromSearchResult extracts case data from a search result item
func (cls *CourtListenerScraper) extractCaseFromSearchResult(s *goquery.Selection) *models.Case {
	c := models.NewCase()

	// Extract case name
	caseName := s.Find("h3.bottom a").Text()
	c.CaseName = strings.TrimSpace(caseName)

	// Extract case URL and ID
	caseURL, exists := s.Find("h3.bottom a").Attr("href")
	if exists {
		c.URL = cls.baseURL + caseURL
		// Extract ID from URL (e.g., /opinion/123456/case-name/)
		parts := strings.Split(strings.Trim(caseURL, "/"), "/")
		if len(parts) >= 2 {
			c.ID = parts[1]
		}
	}

	// Extract court
	court := s.Find(".meta-data-header").Text()
	c.Court = strings.TrimSpace(court)

	// Extract date
	dateStr := s.Find(".meta-data-header time").AttrOr("datetime", "")
	if dateStr != "" {
		if date, err := time.Parse("2006-01-02", dateStr); err == nil {
			c.DecisionDate = &date
		}
	}

	// Extract snippet/summary
	snippet := s.Find(".snippet").Text()
	c.Summary = strings.TrimSpace(snippet)

	// Set metadata
	c.Jurisdiction = "United States"
	c.SourceDatabase = "CourtListener"
	c.ScrapedAt = time.Now()
	c.LastUpdated = time.Now()
	c.Language = "en"
	c.Status = models.CaseStatusActive

	return c
}

// extractCaseDetails extracts detailed case information from a case page
func (cls *CourtListenerScraper) extractCaseDetails(doc *goquery.Document, caseID, caseURL string) (*models.Case, error) {
	c := models.NewCase()

	c.ID = caseID
	c.URL = caseURL

	// Extract case name
	caseName := doc.Find("h1.text-center").Text()
	c.CaseName = strings.TrimSpace(caseName)

	// Extract court
	court := doc.Find(".meta-data-header a").First().Text()
	c.Court = strings.TrimSpace(court)

	// Extract date
	dateStr := doc.Find("time").AttrOr("datetime", "")
	if dateStr != "" {
		if date, err := time.Parse("2006-01-02", dateStr); err == nil {
			c.DecisionDate = &date
		}
	}

	// Extract judges
	doc.Find(".author a").Each(func(i int, s *goquery.Selection) {
		judge := strings.TrimSpace(s.Text())
		if judge != "" {
			c.Judges = append(c.Judges, judge)
		}
	})

	// Extract docket number
	docket := doc.Find(".docket-number").Text()
	c.Docket = strings.TrimSpace(docket)

	// Extract case text
	fullText := doc.Find("#opinion-content").Text()
	c.FullText = strings.TrimSpace(fullText)

	// Set metadata
	c.Jurisdiction = "United States"
	c.SourceDatabase = "CourtListener"
	c.ScrapedAt = time.Now()
	c.LastUpdated = time.Now()
	c.Language = "en"
	c.Status = models.CaseStatusActive

	return c, nil
}
