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

// PacLIIScraper implements scraping for PacLII (Pacific Islands Legal Information Institute)
type PacLIIScraper struct {
	*scraper.BaseScraper
	baseURL string
	client  *http.Client
}

// NewPacLIIScraper creates a new PacLII scraper
func NewPacLIIScraper() *PacLIIScraper {
	baseURL := "http://www.paclii.org"
	base := scraper.NewBaseScraper(
		"PacLII",
		"Pacific Islands",
		baseURL,
		12,
	)

	return &PacLIIScraper{
		BaseScraper: base,
		baseURL:     baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchCases searches for cases matching the query
func (ps *PacLIIScraper) SearchCases(ctx context.Context, query scraper.SearchQuery) ([]*models.Case, error) {
	searchURL, err := ps.buildSearchURL(query)
	if err != nil {
		return nil, errors.ParsingError("failed to build search URL", err)
	}

	allowed, err := ps.BaseScraper.client.CheckRobots(ctx, "/cgi-bin/sinosrch.cgi")
	if err != nil || !allowed {
		return nil, errors.ErrRobotsDisallowed
	}

	if err := ps.BaseScraper.client.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.RateLimitError("rate limit exceeded")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, errors.NetworkError("failed to create request", err)
	}

	req.Header.Set("User-Agent", ps.BaseScraper.client.userAgent)

	resp, err := ps.client.Do(req)
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
	doc.Find("li").Each(func(i int, s *goquery.Selection) {
		if query.Limit > 0 && len(cases) >= query.Limit {
			return
		}
		if s.Find("a").Length() > 0 {
			caseData := ps.extractCaseFromSearchResult(s)
			if caseData != nil {
				cases = append(cases, caseData)
			}
		}
	})

	return cases, nil
}

// GetCaseByID retrieves a specific case by its ID
func (ps *PacLIIScraper) GetCaseByID(ctx context.Context, caseID string) (*models.Case, error) {
	caseURL := ps.buildCaseURL(caseID)

	allowed, err := ps.BaseScraper.client.CheckRobots(ctx, "/")
	if err != nil || !allowed {
		return nil, errors.ErrRobotsDisallowed
	}

	if err := ps.BaseScraper.client.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.RateLimitError("rate limit exceeded")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", caseURL, nil)
	if err != nil {
		return nil, errors.NetworkError("failed to create request", err)
	}

	req.Header.Set("User-Agent", ps.BaseScraper.client.userAgent)

	resp, err := ps.client.Do(req)
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

	return ps.extractCaseDetails(doc, caseID, caseURL)
}

// GetCasesByDateRange retrieves cases within a date range
func (ps *PacLIIScraper) GetCasesByDateRange(ctx context.Context, startDate, endDate time.Time, limit int) ([]*models.Case, error) {
	query := scraper.SearchQuery{
		StartDate: &startDate,
		EndDate:   &endDate,
		Limit:     limit,
	}
	return ps.SearchCases(ctx, query)
}

// IsAvailable checks if PacLII is available
func (ps *PacLIIScraper) IsAvailable(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", ps.baseURL, nil)
	if err != nil {
		return false
	}
	resp, err := ps.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// buildSearchURL builds the search URL with query parameters
func (ps *PacLIIScraper) buildSearchURL(query scraper.SearchQuery) (string, error) {
	params := url.Values{}
	if query.Query != "" {
		params.Set("query", query.Query)
	}
	if query.StartDate != nil && query.EndDate != nil {
		params.Set("datefrom", query.StartDate.Format("2006-01-02"))
		params.Set("dateto", query.EndDate.Format("2006-01-02"))
	}
	params.Set("method", "boolean")
	params.Set("results", "50")

	searchURL := fmt.Sprintf("%s/cgi-bin/sinosrch.cgi?%s", ps.baseURL, params.Encode())
	return searchURL, nil
}

// buildCaseURL builds a case URL from a case ID
func (ps *PacLIIScraper) buildCaseURL(caseID string) string {
	caseID = strings.TrimSpace(caseID)
	if strings.HasPrefix(caseID, "http") {
		return caseID
	}
	if !strings.HasPrefix(caseID, "/") {
		return fmt.Sprintf("%s/%s", ps.baseURL, caseID)
	}
	return ps.baseURL + caseID
}

// extractCaseFromSearchResult extracts case data from a search result item
func (ps *PacLIIScraper) extractCaseFromSearchResult(s *goquery.Selection) *models.Case {
	c := models.NewCase()

	titleLink := s.Find("a").First()
	caseName := titleLink.Text()
	c.CaseName = strings.TrimSpace(caseName)

	caseURL, exists := titleLink.Attr("href")
	if exists {
		if !strings.HasPrefix(caseURL, "http") {
			caseURL = ps.baseURL + caseURL
		}
		c.URL = caseURL

		// Extract jurisdiction from URL path
		if strings.Contains(caseURL, "/fj/") {
			c.Jurisdiction = "Fiji"
		} else if strings.Contains(caseURL, "/ws/") {
			c.Jurisdiction = "Samoa"
		} else if strings.Contains(caseURL, "/to/") {
			c.Jurisdiction = "Tonga"
		} else if strings.Contains(caseURL, "/vu/") {
			c.Jurisdiction = "Vanuatu"
		} else if strings.Contains(caseURL, "/pg/") {
			c.Jurisdiction = "Papua New Guinea"
		} else {
			c.Jurisdiction = "Pacific Islands"
		}

		// Extract ID from URL
		if strings.Contains(caseURL, "/cases/") {
			parts := strings.Split(caseURL, "/cases/")
			if len(parts) >= 2 {
				id := strings.TrimSuffix(parts[1], ".html")
				c.ID = id
			}
		}
	}

	text := s.Text()
	if idx := strings.Index(text, "["); idx != -1 {
		if endIdx := strings.Index(text[idx:], "]"); endIdx != -1 {
			citation := text[idx : idx+endIdx+1]
			c.CaseNumber = strings.TrimSpace(citation)
		}
	}

	c.SourceDatabase = "PacLII"
	c.ScrapedAt = time.Now()
	c.LastUpdated = time.Now()
	c.Language = "en"
	c.Status = models.CaseStatusActive

	return c
}

// extractCaseDetails extracts detailed case information from a case page
func (ps *PacLIIScraper) extractCaseDetails(doc *goquery.Document, caseID, caseURL string) (*models.Case, error) {
	c := models.NewCase()
	c.ID = caseID
	c.URL = caseURL

	caseName := doc.Find("h1").First().Text()
	if caseName == "" {
		caseName = doc.Find("title").First().Text()
	}
	c.CaseName = strings.TrimSpace(caseName)

	citation := doc.Find("center").First().Text()
	c.CaseNumber = strings.TrimSpace(citation)

	// Determine jurisdiction from URL
	if strings.Contains(caseURL, "/fj/") {
		c.Jurisdiction = "Fiji"
		c.Court = "Fiji Courts"
	} else if strings.Contains(caseURL, "/ws/") {
		c.Jurisdiction = "Samoa"
		c.Court = "Samoa Courts"
	} else if strings.Contains(caseURL, "/to/") {
		c.Jurisdiction = "Tonga"
		c.Court = "Tonga Courts"
	} else if strings.Contains(caseURL, "/vu/") {
		c.Jurisdiction = "Vanuatu"
		c.Court = "Vanuatu Courts"
	} else if strings.Contains(caseURL, "/pg/") {
		c.Jurisdiction = "Papua New Guinea"
		c.Court = "Papua New Guinea Courts"
	} else {
		c.Jurisdiction = "Pacific Islands"
	}

	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "Date:") || strings.Contains(text, "Judgment date:") {
			dateStr := strings.TrimSpace(strings.ReplaceAll(text, "Date:", ""))
			dateStr = strings.ReplaceAll(dateStr, "Judgment date:", "")
			dateStr = strings.TrimSpace(dateStr)

			formats := []string{"2 January 2006", "02 January 2006", "2006-01-02"}
			for _, format := range formats {
				if date, err := time.Parse(format, dateStr); err == nil {
					c.DecisionDate = &date
					break
				}
			}
		}
	})

	fullText := doc.Find("body").Text()
	c.FullText = strings.TrimSpace(fullText)

	c.SourceDatabase = "PacLII"
	c.ScrapedAt = time.Now()
	c.LastUpdated = time.Now()
	c.Language = "en"
	c.Status = models.CaseStatusActive

	return c, nil
}
