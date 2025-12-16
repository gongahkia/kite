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

// AustLIIScraper implements scraping for AustLII (Australasian Legal Information Institute)
type AustLIIScraper struct {
	*scraper.BaseScraper
	baseURL string
	client  *http.Client
}

// NewAustLIIScraper creates a new AustLII scraper
func NewAustLIIScraper() *AustLIIScraper {
	baseURL := "https://www.austlii.edu.au"
	base := scraper.NewBaseScraper(
		"AustLII",
		"Australia",
		baseURL,
		12, // 12 requests per minute to be conservative
	)

	return &AustLIIScraper{
		BaseScraper: base,
		baseURL:     baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchCases searches for cases matching the query
func (as *AustLIIScraper) SearchCases(ctx context.Context, query scraper.SearchQuery) ([]*models.Case, error) {
	// Build search URL
	searchURL, err := as.buildSearchURL(query)
	if err != nil {
		return nil, errors.ParsingError("failed to build search URL", err)
	}

	// Check robots.txt
	allowed, err := as.BaseScraper.client.CheckRobots(ctx, "/cgi-bin/sinosrch.cgi")
	if err != nil || !allowed {
		return nil, errors.ErrRobotsDisallowed
	}

	// Wait for rate limit
	if err := as.BaseScraper.client.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.RateLimitError("rate limit exceeded")
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, errors.NetworkError("failed to create request", err)
	}

	req.Header.Set("User-Agent", as.BaseScraper.client.userAgent)

	resp, err := as.client.Do(req)
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

	doc.Find("li").Each(func(i int, s *goquery.Selection) {
		if query.Limit > 0 && len(cases) >= query.Limit {
			return
		}

		// Check if this is a result item
		if s.Find("a").Length() > 0 {
			caseData := as.extractCaseFromSearchResult(s)
			if caseData != nil {
				cases = append(cases, caseData)
			}
		}
	})

	return cases, nil
}

// GetCaseByID retrieves a specific case by its ID
func (as *AustLIIScraper) GetCaseByID(ctx context.Context, caseID string) (*models.Case, error) {
	// AustLII case URLs are typically: /au/cases/cth/HCA/2023/15.html
	caseURL := as.buildCaseURL(caseID)

	// Check robots.txt
	allowed, err := as.BaseScraper.client.CheckRobots(ctx, "/")
	if err != nil || !allowed {
		return nil, errors.ErrRobotsDisallowed
	}

	// Wait for rate limit
	if err := as.BaseScraper.client.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.RateLimitError("rate limit exceeded")
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", caseURL, nil)
	if err != nil {
		return nil, errors.NetworkError("failed to create request", err)
	}

	req.Header.Set("User-Agent", as.BaseScraper.client.userAgent)

	resp, err := as.client.Do(req)
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
	return as.extractCaseDetails(doc, caseID, caseURL)
}

// GetCasesByDateRange retrieves cases within a date range
func (as *AustLIIScraper) GetCasesByDateRange(ctx context.Context, startDate, endDate time.Time, limit int) ([]*models.Case, error) {
	query := scraper.SearchQuery{
		StartDate: &startDate,
		EndDate:   &endDate,
		Limit:     limit,
	}

	return as.SearchCases(ctx, query)
}

// IsAvailable checks if AustLII is available
func (as *AustLIIScraper) IsAvailable(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", as.baseURL, nil)
	if err != nil {
		return false
	}

	resp, err := as.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// buildSearchURL builds the search URL with query parameters
func (as *AustLIIScraper) buildSearchURL(query scraper.SearchQuery) (string, error) {
	params := url.Values{}

	if query.Query != "" {
		params.Set("query", query.Query)
	}

	if query.Court != "" {
		params.Set("court", query.Court)
	}

	if query.StartDate != nil && query.EndDate != nil {
		// AustLII uses a date range format
		params.Set("datefrom", query.StartDate.Format("2006-01-02"))
		params.Set("dateto", query.EndDate.Format("2006-01-02"))
	}

	// Set default database to search all Australian databases
	params.Set("mask", "au")
	params.Set("method", "boolean")
	params.Set("results", "50")

	searchURL := fmt.Sprintf("%s/cgi-bin/sinosrch.cgi?%s", as.baseURL, params.Encode())
	return searchURL, nil
}

// buildCaseURL builds a case URL from a case ID
func (as *AustLIIScraper) buildCaseURL(caseID string) string {
	// Clean the caseID
	caseID = strings.TrimSpace(caseID)

	// If caseID already looks like a path, use it
	if strings.Contains(caseID, "/") {
		if !strings.HasPrefix(caseID, "http") {
			return fmt.Sprintf("%s%s", as.baseURL, caseID)
		}
		return caseID
	}

	// Default: assume High Court format
	return fmt.Sprintf("%s/au/cases/%s.html", as.baseURL, caseID)
}

// extractCaseFromSearchResult extracts case data from a search result item
func (as *AustLIIScraper) extractCaseFromSearchResult(s *goquery.Selection) *models.Case {
	c := models.NewCase()

	// Extract case name and URL from first link
	titleLink := s.Find("a").First()
	caseName := titleLink.Text()
	c.CaseName = strings.TrimSpace(caseName)

	// Extract case URL
	caseURL, exists := titleLink.Attr("href")
	if exists {
		if !strings.HasPrefix(caseURL, "http") {
			caseURL = as.baseURL + caseURL
		}
		c.URL = caseURL

		// Extract ID from URL
		// Example: /au/cases/cth/HCA/2023/15.html -> cth/HCA/2023/15
		if strings.Contains(caseURL, "/cases/") {
			parts := strings.Split(caseURL, "/cases/")
			if len(parts) >= 2 {
				id := strings.TrimSuffix(parts[1], ".html")
				c.ID = id
			}
		}
	}

	// Extract citation and metadata from text
	text := s.Text()

	// Try to extract citation (often in square brackets)
	if idx := strings.Index(text, "["); idx != -1 {
		if endIdx := strings.Index(text[idx:], "]"); endIdx != -1 {
			citation := text[idx : idx+endIdx+1]
			c.CaseNumber = strings.TrimSpace(citation)
		}
	}

	// Extract court information from URL or text
	if strings.Contains(c.URL, "/HCA/") {
		c.Court = "High Court of Australia"
	} else if strings.Contains(c.URL, "/FCA/") {
		c.Court = "Federal Court of Australia"
	} else if strings.Contains(c.URL, "/FCAFC/") {
		c.Court = "Federal Court of Australia (Full Court)"
	} else if strings.Contains(c.URL, "/NSWCA/") {
		c.Court = "New South Wales Court of Appeal"
	} else if strings.Contains(c.URL, "/NSWSC/") {
		c.Court = "Supreme Court of New South Wales"
	} else if strings.Contains(c.URL, "/VCA/") {
		c.Court = "Victorian Court of Appeal"
	} else if strings.Contains(c.URL, "/VSC/") {
		c.Court = "Supreme Court of Victoria"
	}

	// Set jurisdiction
	c.Jurisdiction = "Australia"

	// Set metadata
	c.SourceDatabase = "AustLII"
	c.ScrapedAt = time.Now()
	c.LastUpdated = time.Now()
	c.Language = "en"
	c.Status = models.CaseStatusActive

	return c
}

// extractCaseDetails extracts detailed case information from a case page
func (as *AustLIIScraper) extractCaseDetails(doc *goquery.Document, caseID, caseURL string) (*models.Case, error) {
	c := models.NewCase()

	c.ID = caseID
	c.URL = caseURL

	// Extract case name
	caseName := doc.Find("h1").First().Text()
	if caseName == "" {
		caseName = doc.Find("title").First().Text()
	}
	c.CaseName = strings.TrimSpace(caseName)

	// Extract neutral citation
	citation := doc.Find("center").First().Text()
	if citation == "" {
		// Try to find in title or first paragraph
		citation = doc.Find("p").First().Text()
	}
	c.CaseNumber = strings.TrimSpace(citation)

	// Extract court from URL
	if strings.Contains(caseURL, "/HCA/") {
		c.Court = "High Court of Australia"
	} else if strings.Contains(caseURL, "/FCA/") {
		c.Court = "Federal Court of Australia"
	} else if strings.Contains(caseURL, "/FCAFC/") {
		c.Court = "Federal Court of Australia (Full Court)"
	} else if strings.Contains(caseURL, "/NSWCA/") {
		c.Court = "New South Wales Court of Appeal"
	} else if strings.Contains(caseURL, "/NSWSC/") {
		c.Court = "Supreme Court of New South Wales"
	} else if strings.Contains(caseURL, "/VCA/") {
		c.Court = "Victorian Court of Appeal"
	} else if strings.Contains(caseURL, "/VSC/") {
		c.Court = "Supreme Court of Victoria"
	}

	// Extract date
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "Date:") || strings.Contains(text, "Hearing date:") || strings.Contains(text, "Judgment date:") {
			dateStr := text
			// Clean up date string
			dateStr = strings.TrimSpace(dateStr)
			dateStr = strings.ReplaceAll(dateStr, "Date:", "")
			dateStr = strings.ReplaceAll(dateStr, "Hearing date:", "")
			dateStr = strings.ReplaceAll(dateStr, "Judgment date:", "")
			dateStr = strings.TrimSpace(dateStr)

			// Try to parse various date formats
			formats := []string{
				"2 January 2006",
				"02 January 2006",
				"2006-01-02",
				"02/01/2006",
				"January 2, 2006",
			}
			for _, format := range formats {
				if date, err := time.Parse(format, dateStr); err == nil {
					c.DecisionDate = &date
					return
				}
			}
		}
	})

	// Extract judges
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "Before:") || strings.Contains(text, "Judges:") {
			// Extract judge names after "Before:" or "Judges:"
			parts := strings.Split(text, ":")
			if len(parts) > 1 {
				judgeText := parts[1]
				// Split by common separators
				judges := strings.FieldsFunc(judgeText, func(r rune) bool {
					return r == ',' || r == '&' || strings.ContainsRune(" and ", r)
				})
				for _, judge := range judges {
					judge = strings.TrimSpace(judge)
					if judge != "" && len(judge) > 2 {
						c.Judges = append(c.Judges, judge)
					}
				}
			}
		}
	})

	// Extract docket/case number
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "Case No") || strings.Contains(text, "Matter No") {
			c.Docket = strings.TrimSpace(text)
			return
		}
	})

	// Extract full judgment text
	fullText := doc.Find("body").Text()
	c.FullText = strings.TrimSpace(fullText)

	// Set jurisdiction
	c.Jurisdiction = "Australia"

	// Set metadata
	c.SourceDatabase = "AustLII"
	c.ScrapedAt = time.Now()
	c.LastUpdated = time.Now()
	c.Language = "en"
	c.Status = models.CaseStatusActive

	return c, nil
}
