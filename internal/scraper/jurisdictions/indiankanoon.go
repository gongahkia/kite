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

// IndianKanoonScraper implements scraping for Indian Kanoon (India)
type IndianKanoonScraper struct {
	*scraper.BaseScraper
	baseURL string
	client  *http.Client
}

// NewIndianKanoonScraper creates a new Indian Kanoon scraper
func NewIndianKanoonScraper() *IndianKanoonScraper {
	baseURL := "https://indiankanoon.org"
	base := scraper.NewBaseScraper(
		"IndianKanoon",
		"India",
		baseURL,
		10, // Conservative rate limit
	)

	return &IndianKanoonScraper{
		BaseScraper: base,
		baseURL:     baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchCases searches for cases matching the query
func (iks *IndianKanoonScraper) SearchCases(ctx context.Context, query scraper.SearchQuery) ([]*models.Case, error) {
	searchURL, err := iks.buildSearchURL(query)
	if err != nil {
		return nil, errors.ParsingError("failed to build search URL", err)
	}

	allowed, err := iks.BaseScraper.client.CheckRobots(ctx, "/search/")
	if err != nil || !allowed {
		return nil, errors.ErrRobotsDisallowed
	}

	if err := iks.BaseScraper.client.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.RateLimitError("rate limit exceeded")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, errors.NetworkError("failed to create request", err)
	}

	req.Header.Set("User-Agent", iks.BaseScraper.client.userAgent)

	resp, err := iks.client.Do(req)
	if err != nil {
		return nil, errors.NetworkError("failed to fetch search results", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NetworkError(fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, errors.ParsingError("failed to parse HTML", err)
	}

	cases := make([]*models.Case, 0)
	doc.Find("div.result").Each(func(i int, s *goquery.Selection) {
		if query.Limit > 0 && len(cases) >= query.Limit {
			return
		}
		caseData := iks.extractCaseFromSearchResult(s)
		if caseData != nil {
			cases = append(cases, caseData)
		}
	})

	return cases, nil
}

// GetCaseByID retrieves a specific case by its ID
func (iks *IndianKanoonScraper) GetCaseByID(ctx context.Context, caseID string) (*models.Case, error) {
	caseURL := iks.buildCaseURL(caseID)

	allowed, err := iks.BaseScraper.client.CheckRobots(ctx, "/doc/")
	if err != nil || !allowed {
		return nil, errors.ErrRobotsDisallowed
	}

	if err := iks.BaseScraper.client.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.RateLimitError("rate limit exceeded")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", caseURL, nil)
	if err != nil {
		return nil, errors.NetworkError("failed to create request", err)
	}

	req.Header.Set("User-Agent", iks.BaseScraper.client.userAgent)

	resp, err := iks.client.Do(req)
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

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, errors.ParsingError("failed to parse HTML", err)
	}

	return iks.extractCaseDetails(doc, caseID, caseURL)
}

// GetCasesByDateRange retrieves cases within a date range
func (iks *IndianKanoonScraper) GetCasesByDateRange(ctx context.Context, startDate, endDate time.Time, limit int) ([]*models.Case, error) {
	query := scraper.SearchQuery{
		StartDate: &startDate,
		EndDate:   &endDate,
		Limit:     limit,
	}
	return iks.SearchCases(ctx, query)
}

// IsAvailable checks if Indian Kanoon is available
func (iks *IndianKanoonScraper) IsAvailable(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", iks.baseURL, nil)
	if err != nil {
		return false
	}
	resp, err := iks.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// buildSearchURL builds the search URL with query parameters
func (iks *IndianKanoonScraper) buildSearchURL(query scraper.SearchQuery) (string, error) {
	params := url.Values{}

	if query.Query != "" {
		params.Set("formInput", query.Query)
	}

	if query.Court != "" {
		params.Set("court", query.Court)
	}

	if query.StartDate != nil {
		params.Set("fromYear", fmt.Sprintf("%d", query.StartDate.Year()))
	}

	if query.EndDate != nil {
		params.Set("toYear", fmt.Sprintf("%d", query.EndDate.Year()))
	}

	searchURL := fmt.Sprintf("%s/search/?%s", iks.baseURL, params.Encode())
	return searchURL, nil
}

// buildCaseURL builds a case URL from a case ID
func (iks *IndianKanoonScraper) buildCaseURL(caseID string) string {
	caseID = strings.TrimSpace(caseID)
	if strings.HasPrefix(caseID, "http") {
		return caseID
	}
	// Indian Kanoon uses numeric doc IDs
	return fmt.Sprintf("%s/doc/%s/", iks.baseURL, caseID)
}

// extractCaseFromSearchResult extracts case data from a search result item
func (iks *IndianKanoonScraper) extractCaseFromSearchResult(s *goquery.Selection) *models.Case {
	c := models.NewCase()

	titleLink := s.Find("div.result_title a").First()
	caseName := titleLink.Text()
	c.CaseName = strings.TrimSpace(caseName)

	caseURL, exists := titleLink.Attr("href")
	if exists {
		if !strings.HasPrefix(caseURL, "http") {
			caseURL = iks.baseURL + caseURL
		}
		c.URL = caseURL

		// Extract ID from URL: /doc/123456/ -> 123456
		if strings.Contains(caseURL, "/doc/") {
			parts := strings.Split(caseURL, "/doc/")
			if len(parts) >= 2 {
				id := strings.Trim(parts[1], "/")
				c.ID = id
			}
		}
	}

	// Extract court and date from metadata
	metadata := s.Find("div.doc_cite").Text()
	if metadata != "" {
		// Try to extract court
		if strings.Contains(metadata, "Supreme Court") {
			c.Court = "Supreme Court of India"
		} else if strings.Contains(metadata, "High Court") {
			// Extract specific high court
			c.Court = strings.TrimSpace(metadata)
		}
	}

	// Extract snippet
	snippet := s.Find("div.result_highlight").Text()
	c.Summary = strings.TrimSpace(snippet)

	c.Jurisdiction = "India"
	c.SourceDatabase = "IndianKanoon"
	c.ScrapedAt = time.Now()
	c.LastUpdated = time.Now()
	c.Language = "en"
	c.Status = models.CaseStatusActive

	return c
}

// extractCaseDetails extracts detailed case information from a case page
func (iks *IndianKanoonScraper) extractCaseDetails(doc *goquery.Document, caseID, caseURL string) (*models.Case, error) {
	c := models.NewCase()
	c.ID = caseID
	c.URL = caseURL

	// Extract case name from title
	caseName := doc.Find("h1.doc_heading").First().Text()
	if caseName == "" {
		caseName = doc.Find("title").First().Text()
	}
	c.CaseName = strings.TrimSpace(caseName)

	// Extract citation
	citation := doc.Find("div.doc_cite").First().Text()
	c.CaseNumber = strings.TrimSpace(citation)

	// Extract court from citation
	if strings.Contains(citation, "Supreme Court") {
		c.Court = "Supreme Court of India"
	} else if strings.Contains(citation, "Delhi High Court") {
		c.Court = "Delhi High Court"
	} else if strings.Contains(citation, "Bombay High Court") {
		c.Court = "Bombay High Court"
	} else if strings.Contains(citation, "High Court") {
		c.Court = strings.TrimSpace(citation)
	}

	// Extract date
	doc.Find("p, div").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "Decided On:") || strings.Contains(text, "Date:") {
			dateStr := text
			dateStr = strings.ReplaceAll(dateStr, "Decided On:", "")
			dateStr = strings.ReplaceAll(dateStr, "Date:", "")
			dateStr = strings.TrimSpace(dateStr)

			formats := []string{
				"02.01.2006",
				"02-01-2006",
				"2006-01-02",
				"2 January 2006",
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
		if strings.Contains(text, "Bench:") || strings.Contains(text, "Judge:") {
			parts := strings.Split(text, ":")
			if len(parts) > 1 {
				judges := strings.FieldsFunc(parts[1], func(r rune) bool {
					return r == ',' || r == '&'
				})
				for _, judge := range judges {
					judge = strings.TrimSpace(judge)
					if judge != "" {
						c.Judges = append(c.Judges, judge)
					}
				}
			}
		}
	})

	// Extract full judgment text
	fullText := doc.Find("div.judgments").Text()
	if fullText == "" {
		fullText = doc.Find("div.doc_content").Text()
	}
	c.FullText = strings.TrimSpace(fullText)

	c.Jurisdiction = "India"
	c.SourceDatabase = "IndianKanoon"
	c.ScrapedAt = time.Now()
	c.LastUpdated = time.Now()
	c.Language = "en"
	c.Status = models.CaseStatusActive

	return c, nil
}
