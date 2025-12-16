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

// BAILIIScraper implements scraping for BAILII (British and Irish Legal Information Institute)
type BAILIIScraper struct {
	*scraper.BaseScraper
	baseURL string
	client  *http.Client
}

// NewBAILIIScraper creates a new BAILII scraper
func NewBAILIIScraper() *BAILIIScraper {
	baseURL := "https://www.bailii.org"
	base := scraper.NewBaseScraper(
		"BAILII",
		"United Kingdom & Ireland",
		baseURL,
		12, // 12 requests per minute to be conservative
	)

	return &BAILIIScraper{
		BaseScraper: base,
		baseURL:     baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchCases searches for cases matching the query
func (bs *BAILIIScraper) SearchCases(ctx context.Context, query scraper.SearchQuery) ([]*models.Case, error) {
	// Build search URL
	searchURL, err := bs.buildSearchURL(query)
	if err != nil {
		return nil, errors.ParsingError("failed to build search URL", err)
	}

	// Check robots.txt
	allowed, err := bs.BaseScraper.client.CheckRobots(ctx, "/form/search_multidatabase.html")
	if err != nil || !allowed {
		return nil, errors.ErrRobotsDisallowed
	}

	// Wait for rate limit
	if err := bs.BaseScraper.client.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.RateLimitError("rate limit exceeded")
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, errors.NetworkError("failed to create request", err)
	}

	req.Header.Set("User-Agent", bs.BaseScraper.client.userAgent)

	resp, err := bs.client.Do(req)
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

	doc.Find("li.resultItem").Each(func(i int, s *goquery.Selection) {
		if query.Limit > 0 && len(cases) >= query.Limit {
			return
		}

		caseData := bs.extractCaseFromSearchResult(s)
		if caseData != nil {
			cases = append(cases, caseData)
		}
	})

	return cases, nil
}

// GetCaseByID retrieves a specific case by its ID (e.g., "UKSC/2023/15")
func (bs *BAILIIScraper) GetCaseByID(ctx context.Context, caseID string) (*models.Case, error) {
	// BAILII case IDs are typically in format: UKSC/2023/15
	// Build case URL from ID
	caseURL := bs.buildCaseURL(caseID)

	// Check robots.txt
	allowed, err := bs.BaseScraper.client.CheckRobots(ctx, "/")
	if err != nil || !allowed {
		return nil, errors.ErrRobotsDisallowed
	}

	// Wait for rate limit
	if err := bs.BaseScraper.client.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.RateLimitError("rate limit exceeded")
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", caseURL, nil)
	if err != nil {
		return nil, errors.NetworkError("failed to create request", err)
	}

	req.Header.Set("User-Agent", bs.BaseScraper.client.userAgent)

	resp, err := bs.client.Do(req)
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
	return bs.extractCaseDetails(doc, caseID, caseURL)
}

// GetCasesByDateRange retrieves cases within a date range
func (bs *BAILIIScraper) GetCasesByDateRange(ctx context.Context, startDate, endDate time.Time, limit int) ([]*models.Case, error) {
	query := scraper.SearchQuery{
		StartDate: &startDate,
		EndDate:   &endDate,
		Limit:     limit,
	}

	return bs.SearchCases(ctx, query)
}

// IsAvailable checks if BAILII is available
func (bs *BAILIIScraper) IsAvailable(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", bs.baseURL, nil)
	if err != nil {
		return false
	}

	resp, err := bs.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// buildSearchURL builds the search URL with query parameters
func (bs *BAILIIScraper) buildSearchURL(query scraper.SearchQuery) (string, error) {
	params := url.Values{}

	if query.Query != "" {
		params.Set("query", query.Query)
	}

	if query.Court != "" {
		params.Set("court", query.Court)
	}

	if query.StartDate != nil && query.EndDate != nil {
		// BAILII uses a date range format
		dateRange := fmt.Sprintf("%s to %s",
			query.StartDate.Format("02/01/2006"),
			query.EndDate.Format("02/01/2006"))
		params.Set("date", dateRange)
	}

	// Set default database to search all UK and Irish databases
	params.Set("database", "all")
	params.Set("method", "boolean")

	searchURL := fmt.Sprintf("%s/cgi-bin/lucy_search_1.cgi?%s", bs.baseURL, params.Encode())
	return searchURL, nil
}

// buildCaseURL builds a case URL from a case citation (e.g., "UKSC/2023/15")
func (bs *BAILIIScraper) buildCaseURL(caseID string) string {
	// BAILII URLs are structured as: /jurisdiction/court/year/number.html
	// Example: /uk/cases/UKSC/2023/15.html

	// Clean the caseID
	caseID = strings.TrimSpace(caseID)

	// If caseID already looks like a path, use it
	if strings.Contains(caseID, "/") {
		// Extract components
		parts := strings.Split(caseID, "/")
		if len(parts) >= 3 {
			// Format: court/year/number
			court := strings.ToUpper(parts[0])
			year := parts[1]
			number := parts[2]

			// Determine jurisdiction from court code
			jurisdiction := "uk"
			if strings.Contains(court, "IE") || strings.Contains(court, "IESC") {
				jurisdiction = "ie"
			}

			return fmt.Sprintf("%s/%s/cases/%s/%s/%s.html",
				bs.baseURL, jurisdiction, court, year, number)
		}
	}

	// Default: assume UK Supreme Court format
	return fmt.Sprintf("%s/uk/cases/%s.html", bs.baseURL, caseID)
}

// extractCaseFromSearchResult extracts case data from a search result item
func (bs *BAILIIScraper) extractCaseFromSearchResult(s *goquery.Selection) *models.Case {
	c := models.NewCase()

	// Extract case name and URL
	titleLink := s.Find("a.resultTitle")
	caseName := titleLink.Text()
	c.CaseName = strings.TrimSpace(caseName)

	// Extract case URL
	caseURL, exists := titleLink.Attr("href")
	if exists {
		if !strings.HasPrefix(caseURL, "http") {
			caseURL = bs.baseURL + caseURL
		}
		c.URL = caseURL

		// Extract ID from URL
		// Example: /uk/cases/UKSC/2023/15.html -> UKSC/2023/15
		if strings.Contains(caseURL, "/cases/") {
			parts := strings.Split(caseURL, "/cases/")
			if len(parts) >= 2 {
				id := strings.TrimSuffix(parts[1], ".html")
				c.ID = id
			}
		}
	}

	// Extract citation
	citation := s.Find(".resultCitation").Text()
	c.CaseNumber = strings.TrimSpace(citation)

	// Extract court from metadata
	court := s.Find(".resultCourt").Text()
	c.Court = strings.TrimSpace(court)

	// Extract date from metadata
	dateStr := s.Find(".resultDate").Text()
	if dateStr != "" {
		// Try to parse various date formats used by BAILII
		formats := []string{
			"02 January 2006",
			"2 January 2006",
			"02/01/2006",
			"2006-01-02",
		}
		for _, format := range formats {
			if date, err := time.Parse(format, strings.TrimSpace(dateStr)); err == nil {
				c.DecisionDate = &date
				break
			}
		}
	}

	// Extract snippet/summary
	snippet := s.Find(".resultSnippet").Text()
	c.Summary = strings.TrimSpace(snippet)

	// Determine jurisdiction from URL or court
	if strings.Contains(c.URL, "/ie/") || strings.Contains(c.Court, "Irish") {
		c.Jurisdiction = "Ireland"
	} else {
		c.Jurisdiction = "United Kingdom"
	}

	// Set metadata
	c.SourceDatabase = "BAILII"
	c.ScrapedAt = time.Now()
	c.LastUpdated = time.Now()
	c.Language = "en"
	c.Status = models.CaseStatusActive

	return c
}

// extractCaseDetails extracts detailed case information from a case page
func (bs *BAILIIScraper) extractCaseDetails(doc *goquery.Document, caseID, caseURL string) (*models.Case, error) {
	c := models.NewCase()

	c.ID = caseID
	c.URL = caseURL

	// Extract case name - BAILII uses various selectors
	caseName := doc.Find("h1.case-title").First().Text()
	if caseName == "" {
		caseName = doc.Find("blockquote > p > b").First().Text()
	}
	if caseName == "" {
		caseName = doc.Find("h2").First().Text()
	}
	c.CaseName = strings.TrimSpace(caseName)

	// Extract neutral citation
	citation := doc.Find(".citation").Text()
	if citation == "" {
		// Try alternative selectors
		citation = doc.Find("p:contains('[')").First().Text()
	}
	c.CaseNumber = strings.TrimSpace(citation)

	// Extract court
	court := doc.Find(".court").Text()
	if court == "" {
		// Extract from citation/URL
		if strings.Contains(caseID, "UKSC") {
			court = "UK Supreme Court"
		} else if strings.Contains(caseID, "EWCA") {
			court = "Court of Appeal (England & Wales)"
		} else if strings.Contains(caseID, "EWHC") {
			court = "High Court (England & Wales)"
		}
	}
	c.Court = strings.TrimSpace(court)

	// Extract date
	dateStr := doc.Find(".judgment-date").Text()
	if dateStr == "" {
		// Try to find date in various formats
		doc.Find("p").Each(func(i int, s *goquery.Selection) {
			text := s.Text()
			if strings.Contains(text, "Date:") || strings.Contains(text, "Judgment Date:") {
				dateStr = text
				return
			}
		})
	}

	if dateStr != "" {
		// Clean up date string
		dateStr = strings.TrimSpace(dateStr)
		dateStr = strings.TrimPrefix(dateStr, "Date:")
		dateStr = strings.TrimPrefix(dateStr, "Judgment Date:")
		dateStr = strings.TrimSpace(dateStr)

		// Try to parse various date formats
		formats := []string{
			"02 January 2006",
			"2 January 2006",
			"02/01/2006",
			"2006-01-02",
			"January 2, 2006",
		}
		for _, format := range formats {
			if date, err := time.Parse(format, dateStr); err == nil {
				c.DecisionDate = &date
				break
			}
		}
	}

	// Extract judges
	doc.Find(".judge").Each(func(i int, s *goquery.Selection) {
		judge := strings.TrimSpace(s.Text())
		if judge != "" {
			c.Judges = append(c.Judges, judge)
		}
	})

	// If no judges found with .judge selector, try alternative
	if len(c.Judges) == 0 {
		doc.Find("p:contains('Before:')").Each(func(i int, s *goquery.Selection) {
			text := s.Text()
			// Extract judge names after "Before:"
			if strings.Contains(text, "Before:") {
				parts := strings.Split(text, "Before:")
				if len(parts) > 1 {
					judgeText := parts[1]
					// Split by common separators
					judges := strings.Split(judgeText, ",")
					for _, judge := range judges {
						judge = strings.TrimSpace(judge)
						if judge != "" {
							c.Judges = append(c.Judges, judge)
						}
					}
				}
			}
		})
	}

	// Extract docket/case number
	docket := doc.Find(".docket-number").Text()
	if docket == "" {
		doc.Find("p").Each(func(i int, s *goquery.Selection) {
			text := s.Text()
			if strings.Contains(text, "Case No:") || strings.Contains(text, "Case Number:") {
				docket = text
				return
			}
		})
	}
	c.Docket = strings.TrimSpace(docket)

	// Extract full judgment text
	fullText := doc.Find(".judgment-body").Text()
	if fullText == "" {
		// Try alternative selectors
		fullText = doc.Find("ol[type='1']").Text()
		if fullText == "" {
			fullText = doc.Find("blockquote").Text()
		}
	}
	c.FullText = strings.TrimSpace(fullText)

	// Determine jurisdiction from URL
	if strings.Contains(caseURL, "/ie/") {
		c.Jurisdiction = "Ireland"
	} else {
		c.Jurisdiction = "United Kingdom"
	}

	// Set metadata
	c.SourceDatabase = "BAILII"
	c.ScrapedAt = time.Now()
	c.LastUpdated = time.Now()
	c.Language = "en"
	c.Status = models.CaseStatusActive

	return c, nil
}
