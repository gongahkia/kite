package jurisdictions

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gongahkia/kite/internal/scraper"
	"github.com/gongahkia/kite/pkg/errors"
	"github.com/gongahkia/kite/pkg/models"
)

// SingaporeLawWatchScraper implements scraping for Singapore Law Watch
type SingaporeLawWatchScraper struct {
	*scraper.BaseScraper
	baseURL string
	client  *http.Client
}

// NewSingaporeLawWatchScraper creates a new Singapore Law Watch scraper
func NewSingaporeLawWatchScraper() *SingaporeLawWatchScraper {
	baseURL := "https://www.lawnet.sg"
	base := scraper.NewBaseScraper(
		"SingaporeLawWatch",
		"Singapore",
		baseURL,
		8, // Conservative rate limit for commercial site
	)

	return &SingaporeLawWatchScraper{
		BaseScraper: base,
		baseURL:     baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SearchCases searches for cases matching the query
func (sls *SingaporeLawWatchScraper) SearchCases(ctx context.Context, query scraper.SearchQuery) ([]*models.Case, error) {
	// Note: Singapore Law Watch often requires authentication
	// This is a basic implementation that may need to be enhanced

	allowed, err := sls.BaseScraper.client.CheckRobots(ctx, "/")
	if err != nil || !allowed {
		return nil, errors.ErrRobotsDisallowed
	}

	if err := sls.BaseScraper.client.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.RateLimitError("rate limit exceeded")
	}

	// For now, return empty results with a note
	// Real implementation would require authentication and proper API access
	return []*models.Case{}, nil
}

// GetCaseByID retrieves a specific case by its ID
func (sls *SingaporeLawWatchScraper) GetCaseByID(ctx context.Context, caseID string) (*models.Case, error) {
	// Singapore cases use neutral citations like [2023] SGCA 15
	caseURL := sls.buildCaseURL(caseID)

	allowed, err := sls.BaseScraper.client.CheckRobots(ctx, "/")
	if err != nil || !allowed {
		return nil, errors.ErrRobotsDisallowed
	}

	if err := sls.BaseScraper.client.rateLimiter.Wait(ctx); err != nil {
		return nil, errors.RateLimitError("rate limit exceeded")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", caseURL, nil)
	if err != nil {
		return nil, errors.NetworkError("failed to create request", err)
	}

	req.Header.Set("User-Agent", sls.BaseScraper.client.userAgent)

	resp, err := sls.client.Do(req)
	if err != nil {
		return nil, errors.NetworkError("failed to fetch case", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.ErrNotFound
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, errors.ValidationError("authentication required for Singapore Law Watch")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.NetworkError(fmt.Sprintf("unexpected status code: %d", resp.StatusCode), nil)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, errors.ParsingError("failed to parse HTML", err)
	}

	return sls.extractCaseDetails(doc, caseID, caseURL)
}

// GetCasesByDateRange retrieves cases within a date range
func (sls *SingaporeLawWatchScraper) GetCasesByDateRange(ctx context.Context, startDate, endDate time.Time, limit int) ([]*models.Case, error) {
	query := scraper.SearchQuery{
		StartDate: &startDate,
		EndDate:   &endDate,
		Limit:     limit,
	}
	return sls.SearchCases(ctx, query)
}

// IsAvailable checks if Singapore Law Watch is available
func (sls *SingaporeLawWatchScraper) IsAvailable(ctx context.Context) bool {
	req, err := http.NewRequestWithContext(ctx, "HEAD", sls.baseURL, nil)
	if err != nil {
		return false
	}
	resp, err := sls.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// buildCaseURL builds a case URL from a case citation
func (sls *SingaporeLawWatchScraper) buildCaseURL(caseID string) string {
	caseID = strings.TrimSpace(caseID)
	if strings.HasPrefix(caseID, "http") {
		return caseID
	}
	// Singapore neutral citations: [2023] SGCA 15
	return fmt.Sprintf("%s/case/%s", sls.baseURL, caseID)
}

// extractCaseDetails extracts detailed case information from a case page
func (sls *SingaporeLawWatchScraper) extractCaseDetails(doc *goquery.Document, caseID, caseURL string) (*models.Case, error) {
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
	c.CaseNumber = caseID

	// Determine court from citation
	if strings.Contains(caseID, "SGCA") {
		c.Court = "Court of Appeal of Singapore"
	} else if strings.Contains(caseID, "SGHC") {
		c.Court = "High Court of Singapore"
	} else if strings.Contains(caseID, "SGDC") {
		c.Court = "District Court of Singapore"
	} else if strings.Contains(caseID, "SGMC") {
		c.Court = "Magistrate's Court of Singapore"
	}

	// Extract date and other metadata
	doc.Find("p, div").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "Date:") || strings.Contains(text, "Judgment Date:") {
			dateStr := strings.TrimSpace(strings.ReplaceAll(text, "Date:", ""))
			dateStr = strings.ReplaceAll(dateStr, "Judgment Date:", "")
			dateStr = strings.TrimSpace(dateStr)

			formats := []string{"2 January 2006", "02 January 2006", "2006-01-02", "02/01/2006"}
			for _, format := range formats {
				if date, err := time.Parse(format, dateStr); err == nil {
					c.DecisionDate = &date
					break
				}
			}
		}

		if strings.Contains(text, "Coram:") || strings.Contains(text, "Before:") {
			parts := strings.Split(text, ":")
			if len(parts) > 1 {
				judges := strings.FieldsFunc(parts[1], func(r rune) bool {
					return r == ',' || r == ';'
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
	fullText := doc.Find("div.judgment-text").Text()
	if fullText == "" {
		fullText = doc.Find("body").Text()
	}
	c.FullText = strings.TrimSpace(fullText)

	c.Jurisdiction = "Singapore"
	c.SourceDatabase = "SingaporeLawWatch"
	c.ScrapedAt = time.Now()
	c.LastUpdated = time.Now()
	c.Language = "en"
	c.Status = models.CaseStatusActive

	return c, nil
}
