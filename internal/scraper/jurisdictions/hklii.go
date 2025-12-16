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

// HKLIIScraper implements scraping for HKLII (Hong Kong Legal Information Institute)
type HKLIIScraper struct {
	*scraper.BaseScraper
	baseURL string
	client  *http.Client
}

// NewHKLIIScraper creates a new HKLII scraper
func NewHKLIIScraper() *HKLIIScraper {
	baseURL := "https://www.hklii.hk"
	base := scraper.NewBaseScraper(
		"HKLII",
		"Hong Kong",
		baseURL,
		10, // 10 requests per minute to be conservative
	)

	return &HKLIIScraper{
		BaseScraper: base,
		baseURL:     baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchCases searches for cases matching the query
func (hs *HKLIIScraper) SearchCases(ctx context.Context, query scraper.SearchQuery) ([]*models.Case, error) {
	// Build search URL
	searchURL, err := hs.buildSearchURL(query)
	if err != nil {
		return nil, errors.ParsingError("failed to build search URL", err)
	}

	// Check robots.txt
	allowed, err := hs.BaseScraper.client.CheckRobots(ctx, "/cgi-bin/sinosrch.cgi")
	if err != nil || !allowed {
		return nil, errors.ErrRobotsDisallowed
	}

	// Wait for rate limit
	if err := hs.BaseScraper.client.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.RateLimitError("rate limit exceeded")
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, errors.NetworkError("failed to create request", err)
	}

	req.Header.Set("User-Agent", hs.BaseScraper.client.userAgent)

	resp, err := hs.client.Do(req)
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

		if s.Find("a").Length() > 0 {
			caseData := hs.extractCaseFromSearchResult(s)
			if caseData != nil {
				cases = append(cases, caseData)
			}
		}
	})

	return cases, nil
}

// GetCaseByID retrieves a specific case by its ID
func (hs *HKLIIScraper) GetCaseByID(ctx context.Context, caseID string) (*models.Case, error) {
	caseURL := hs.buildCaseURL(caseID)

	// Check robots.txt
	allowed, err := hs.BaseScraper.client.CheckRobots(ctx, "/")
	if err != nil || !allowed {
		return nil, errors.ErrRobotsDisallowed
	}

	// Wait for rate limit
	if err := hs.BaseScraper.client.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.RateLimitError("rate limit exceeded")
	}

	// Make HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", caseURL, nil)
	if err != nil {
		return nil, errors.NetworkError("failed to create request", err)
	}

	req.Header.Set("User-Agent", hs.BaseScraper.client.userAgent)

	resp, err := hs.client.Do(req)
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
	return hs.extractCaseDetails(doc, caseID, caseURL)
}

// GetCasesByDateRange retrieves cases within a date range
func (hs *HKLIIScraper) GetCasesByDateRange(ctx context.Context, startDate, endDate time.Time, limit int) ([]*models.Case, error) {
	query := scraper.SearchQuery{
		StartDate: &startDate,
		EndDate:   &endDate,
		Limit:     limit,
	}

	return hs.SearchCases(ctx, query)
}

// IsAvailable checks if HKLII is available
func (hs *HKLIIScraper) IsAvailable(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", hs.baseURL, nil)
	if err != nil {
		return false
	}

	resp, err := hs.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// buildSearchURL builds the search URL with query parameters
func (hs *HKLIIScraper) buildSearchURL(query scraper.SearchQuery) (string, error) {
	params := url.Values{}

	if query.Query != "" {
		params.Set("query", query.Query)
	}

	if query.Court != "" {
		params.Set("court", query.Court)
	}

	if query.StartDate != nil && query.EndDate != nil {
		params.Set("datefrom", query.StartDate.Format("2006-01-02"))
		params.Set("dateto", query.EndDate.Format("2006-01-02"))
	}

	// Set default database to search all Hong Kong databases
	params.Set("mask", "hk")
	params.Set("method", "boolean")
	params.Set("results", "50")

	searchURL := fmt.Sprintf("%s/cgi-bin/sinosrch.cgi?%s", hs.baseURL, params.Encode())
	return searchURL, nil
}

// buildCaseURL builds a case URL from a case ID
func (hs *HKLIIScraper) buildCaseURL(caseID string) string {
	caseID = strings.TrimSpace(caseID)

	if strings.Contains(caseID, "/") {
		if !strings.HasPrefix(caseID, "http") {
			return fmt.Sprintf("%s%s", hs.baseURL, caseID)
		}
		return caseID
	}

	// Default format
	return fmt.Sprintf("%s/hk/cases/%s.html", hs.baseURL, caseID)
}

// extractCaseFromSearchResult extracts case data from a search result item
func (hs *HKLIIScraper) extractCaseFromSearchResult(s *goquery.Selection) *models.Case {
	c := models.NewCase()

	titleLink := s.Find("a").First()
	caseName := titleLink.Text()
	c.CaseName = strings.TrimSpace(caseName)

	caseURL, exists := titleLink.Attr("href")
	if exists {
		if !strings.HasPrefix(caseURL, "http") {
			caseURL = hs.baseURL + caseURL
		}
		c.URL = caseURL

		if strings.Contains(caseURL, "/cases/") {
			parts := strings.Split(caseURL, "/cases/")
			if len(parts) >= 2 {
				id := strings.TrimSuffix(parts[1], ".html")
				c.ID = id
			}
		}
	}

	text := s.Text()

	// Extract citation
	if idx := strings.Index(text, "["); idx != -1 {
		if endIdx := strings.Index(text[idx:], "]"); endIdx != -1 {
			citation := text[idx : idx+endIdx+1]
			c.CaseNumber = strings.TrimSpace(citation)
		}
	}

	// Extract court from URL
	if strings.Contains(c.URL, "/HKCFA/") {
		c.Court = "Court of Final Appeal"
	} else if strings.Contains(c.URL, "/HKCA/") {
		c.Court = "Court of Appeal"
	} else if strings.Contains(c.URL, "/HKCFI/") {
		c.Court = "Court of First Instance"
	} else if strings.Contains(c.URL, "/HKDC/") {
		c.Court = "District Court"
	} else if strings.Contains(c.URL, "/HKCU/") {
		c.Court = "Competition Tribunal"
	}

	c.Jurisdiction = "Hong Kong"
	c.SourceDatabase = "HKLII"
	c.ScrapedAt = time.Now()
	c.LastUpdated = time.Now()
	c.Language = "en"
	c.Status = models.CaseStatusActive

	return c
}

// extractCaseDetails extracts detailed case information from a case page
func (hs *HKLIIScraper) extractCaseDetails(doc *goquery.Document, caseID, caseURL string) (*models.Case, error) {
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
		citation = doc.Find("p").First().Text()
	}
	c.CaseNumber = strings.TrimSpace(citation)

	// Extract court from URL
	if strings.Contains(caseURL, "/HKCFA/") {
		c.Court = "Court of Final Appeal"
	} else if strings.Contains(caseURL, "/HKCA/") {
		c.Court = "Court of Appeal"
	} else if strings.Contains(caseURL, "/HKCFI/") {
		c.Court = "Court of First Instance"
	} else if strings.Contains(caseURL, "/HKDC/") {
		c.Court = "District Court"
	} else if strings.Contains(caseURL, "/HKCU/") {
		c.Court = "Competition Tribunal"
	}

	// Extract date
	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "Date:") || strings.Contains(text, "Judgment date:") {
			dateStr := text
			dateStr = strings.TrimSpace(dateStr)
			dateStr = strings.ReplaceAll(dateStr, "Date:", "")
			dateStr = strings.ReplaceAll(dateStr, "Judgment date:", "")
			dateStr = strings.TrimSpace(dateStr)

			formats := []string{
				"2 January 2006",
				"02 January 2006",
				"2006-01-02",
				"02/01/2006",
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
			parts := strings.Split(text, ":")
			if len(parts) > 1 {
				judgeText := parts[1]
				judges := strings.FieldsFunc(judgeText, func(r rune) bool {
					return r == ',' || r == '&'
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
		if strings.Contains(text, "Case No") || strings.Contains(text, "HCAL") {
			c.Docket = strings.TrimSpace(text)
			return
		}
	})

	// Extract full judgment text
	fullText := doc.Find("body").Text()
	c.FullText = strings.TrimSpace(fullText)

	c.Jurisdiction = "Hong Kong"
	c.SourceDatabase = "HKLII"
	c.ScrapedAt = time.Now()
	c.LastUpdated = time.Now()
	c.Language = "en"
	c.Status = models.CaseStatusActive

	return c, nil
}
