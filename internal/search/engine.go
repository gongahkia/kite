package search

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gongahkia/kite/internal/observability"
	"github.com/gongahkia/kite/internal/storage"
	"github.com/gongahkia/kite/pkg/models"
)

// SearchEngine provides advanced search capabilities
type SearchEngine struct {
	storage storage.Storage
	logger  *observability.Logger
	metrics *observability.Metrics
}

// SearchResult represents a single search result
type SearchResult struct {
	Case       *models.Case
	Score      float64
	Highlights []string
	Explain    string
}

// SearchResponse represents search results with metadata
type SearchResponse struct {
	Results    []*SearchResult
	TotalHits  int
	SearchTime time.Duration
	Facets     map[string]*Facet
	Cursor     *string
}

// Facet represents a faceted search result
type Facet struct {
	Field  string
	Values []*FacetValue
}

// FacetValue represents a value in a facet
type FacetValue struct {
	Value string
	Count int
}

// NewSearchEngine creates a new search engine
func NewSearchEngine(storage storage.Storage, logger *observability.Logger, metrics *observability.Metrics) *SearchEngine {
	return &SearchEngine{
		storage: storage,
		logger:  logger,
		metrics: metrics,
	}
}

// Search executes a search query
func (se *SearchEngine) Search(ctx context.Context, query *Query) (*SearchResponse, error) {
	start := time.Now()

	// Validate query
	if err := query.Validate(); err != nil {
		return nil, fmt.Errorf("invalid query: %w", err)
	}

	se.logger.WithField("query", query.String()).Info("Executing search")

	// Convert query to storage query
	storageQuery := se.convertToStorageQuery(query)

	// Execute search
	cases, err := se.storage.SearchCases(ctx, storageQuery)
	if err != nil {
		se.logger.WithField("error", err).Error("Search failed")
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Build results
	results := make([]*SearchResult, len(cases))
	for i, c := range cases {
		results[i] = &SearchResult{
			Case:       c,
			Score:      se.calculateRelevanceScore(c, query),
			Highlights: se.extractHighlights(c, query),
		}
	}

	// Sort by score if not already sorted
	if query.Sort.Field == "relevance" || query.Sort.Field == "" {
		sort.Slice(results, func(i, j int) bool {
			if query.Sort.Desc {
				return results[i].Score > results[j].Score
			}
			return results[i].Score < results[j].Score
		})
	}

	// Calculate facets if requested
	facets := make(map[string]*Facet)
	if len(query.Facets) > 0 {
		facets = se.calculateFacets(cases, query.Facets)
	}

	searchTime := time.Since(start)

	// Record metrics
	se.metrics.RecordSearchQuery(searchTime, len(results))

	response := &SearchResponse{
		Results:    results,
		TotalHits:  len(results),
		SearchTime: searchTime,
		Facets:     facets,
	}

	se.logger.WithFields(map[string]interface{}{
		"total_hits":  len(results),
		"search_time": searchTime.Milliseconds(),
	}).Info("Search completed")

	return response, nil
}

// convertToStorageQuery converts a search query to storage query
func (se *SearchEngine) convertToStorageQuery(query *Query) storage.SearchQuery {
	sq := storage.SearchQuery{
		Query:  query.Text,
		Fields: query.Fields,
		Limit:  query.Page.Limit,
		Offset: query.Page.Offset,
	}

	// Set fuzzy based on query type
	sq.Fuzzy = query.Type == QueryTypeFuzzy

	// Convert filters
	if query.Filters != nil {
		sq.Filters = storage.CaseFilter{
			IDs:          query.Filters.IDs,
			Jurisdiction: ptrToString(query.Filters.Jurisdiction),
			Court:        ptrToString(query.Filters.Court),
			Judges:       query.Filters.Judges,
			Concepts:     query.Filters.Concepts,
			MinQuality:   ptrToFloat(query.Filters.MinQuality),
			Limit:        query.Page.Limit,
			Offset:       query.Page.Offset,
		}

		if query.Filters.CourtLevel != nil {
			sq.Filters.CourtLevel = query.Filters.CourtLevel
		}

		if query.Filters.Status != nil {
			sq.Filters.Status = *query.Filters.Status
		}

		if query.Filters.StartDate != nil {
			sq.Filters.StartDate = query.Filters.StartDate
		}

		if query.Filters.EndDate != nil {
			sq.Filters.EndDate = query.Filters.EndDate
		}

		// Set sort options
		if query.Sort != nil && query.Sort.Field != "" {
			sq.Filters.OrderBy = query.Sort.Field
			sq.Filters.OrderDesc = query.Sort.Desc
		}
	}

	return sq
}

// calculateRelevanceScore calculates relevance score for a case
func (se *SearchEngine) calculateRelevanceScore(c *models.Case, query *Query) float64 {
	if query.Text == "" {
		return 1.0
	}

	score := 0.0
	queryTerms := strings.Fields(strings.ToLower(query.Text))

	// Check case name (weight: 3.0)
	for _, term := range queryTerms {
		if strings.Contains(strings.ToLower(c.CaseName), term) {
			score += 3.0
		}
	}

	// Check summary (weight: 2.0)
	for _, term := range queryTerms {
		if strings.Contains(strings.ToLower(c.Summary), term) {
			score += 2.0
		}
	}

	// Check full text (weight: 1.0)
	for _, term := range queryTerms {
		if strings.Contains(strings.ToLower(c.FullText), term) {
			score += 1.0
		}
	}

	// Check legal concepts (weight: 2.5)
	for _, term := range queryTerms {
		for _, concept := range c.LegalConcepts {
			if strings.Contains(strings.ToLower(concept), term) {
				score += 2.5
			}
		}
	}

	// Normalize by query length
	if len(queryTerms) > 0 {
		score = score / float64(len(queryTerms))
	}

	// Boost by quality score
	if c.QualityScore != nil {
		score *= (1.0 + *c.QualityScore)
	}

	return score
}

// extractHighlights extracts highlighted snippets from the case
func (se *SearchEngine) extractHighlights(c *models.Case, query *Query) []string {
	if query.Text == "" {
		return []string{}
	}

	highlights := []string{}
	queryTerms := strings.Fields(strings.ToLower(query.Text))

	// Highlight in case name
	if containsAny(strings.ToLower(c.CaseName), queryTerms) {
		highlights = append(highlights, se.highlightText(c.CaseName, queryTerms, 100))
	}

	// Highlight in summary
	if containsAny(strings.ToLower(c.Summary), queryTerms) {
		highlights = append(highlights, se.highlightText(c.Summary, queryTerms, 200))
	}

	// Limit to top 5 highlights
	if len(highlights) > 5 {
		highlights = highlights[:5]
	}

	return highlights
}

// highlightText highlights query terms in text
func (se *SearchEngine) highlightText(text string, terms []string, maxLen int) string {
	lowerText := strings.ToLower(text)

	// Find first occurrence of any term
	firstIdx := len(text)
	for _, term := range terms {
		idx := strings.Index(lowerText, term)
		if idx >= 0 && idx < firstIdx {
			firstIdx = idx
		}
	}

	if firstIdx == len(text) {
		// No match found, return beginning
		if len(text) > maxLen {
			return text[:maxLen] + "..."
		}
		return text
	}

	// Extract context around the match
	start := firstIdx - 50
	if start < 0 {
		start = 0
	}

	end := firstIdx + maxLen
	if end > len(text) {
		end = len(text)
	}

	snippet := text[start:end]

	// Add ellipsis
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(text) {
		snippet = snippet + "..."
	}

	// Highlight terms
	for _, term := range terms {
		snippet = strings.ReplaceAll(snippet, term, "<em>"+term+"</em>")
		snippet = strings.ReplaceAll(snippet, strings.Title(term), "<em>"+strings.Title(term)+"</em>")
	}

	return snippet
}

// calculateFacets calculates faceted search results
func (se *SearchEngine) calculateFacets(cases []*models.Case, facetFields []string) map[string]*Facet {
	facets := make(map[string]*Facet)

	for _, field := range facetFields {
		switch field {
		case "jurisdiction":
			facets["jurisdiction"] = se.calculateJurisdictionFacet(cases)
		case "court":
			facets["court"] = se.calculateCourtFacet(cases)
		case "court_level":
			facets["court_level"] = se.calculateCourtLevelFacet(cases)
		case "year":
			facets["year"] = se.calculateYearFacet(cases)
		case "concepts":
			facets["concepts"] = se.calculateConceptsFacet(cases)
		}
	}

	return facets
}

// calculateJurisdictionFacet calculates jurisdiction facet
func (se *SearchEngine) calculateJurisdictionFacet(cases []*models.Case) *Facet {
	counts := make(map[string]int)

	for _, c := range cases {
		counts[c.Jurisdiction]++
	}

	return buildFacet("jurisdiction", counts)
}

// calculateCourtFacet calculates court facet
func (se *SearchEngine) calculateCourtFacet(cases []*models.Case) *Facet {
	counts := make(map[string]int)

	for _, c := range cases {
		counts[c.Court]++
	}

	return buildFacet("court", counts)
}

// calculateCourtLevelFacet calculates court level facet
func (se *SearchEngine) calculateCourtLevelFacet(cases []*models.Case) *Facet {
	counts := make(map[string]int)

	for _, c := range cases {
		level := fmt.Sprintf("%d", c.CourtLevel)
		counts[level]++
	}

	return buildFacet("court_level", counts)
}

// calculateYearFacet calculates year facet
func (se *SearchEngine) calculateYearFacet(cases []*models.Case) *Facet {
	counts := make(map[string]int)

	for _, c := range cases {
		if c.DecisionDate != nil {
			year := fmt.Sprintf("%d", c.DecisionDate.Year())
			counts[year]++
		}
	}

	return buildFacet("year", counts)
}

// calculateConceptsFacet calculates legal concepts facet
func (se *SearchEngine) calculateConceptsFacet(cases []*models.Case) *Facet {
	counts := make(map[string]int)

	for _, c := range cases {
		for _, concept := range c.LegalConcepts {
			counts[concept]++
		}
	}

	return buildFacet("concepts", counts)
}

// buildFacet builds a facet from counts
func buildFacet(field string, counts map[string]int) *Facet {
	values := make([]*FacetValue, 0, len(counts))

	for value, count := range counts {
		values = append(values, &FacetValue{
			Value: value,
			Count: count,
		})
	}

	// Sort by count descending
	sort.Slice(values, func(i, j int) bool {
		return values[i].Count > values[j].Count
	})

	// Limit to top 20
	if len(values) > 20 {
		values = values[:20]
	}

	return &Facet{
		Field:  field,
		Values: values,
	}
}

// Helper functions

func ptrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func ptrToFloat(f *float64) float64 {
	if f == nil {
		return 0
	}
	return *f
}

func containsAny(text string, terms []string) bool {
	for _, term := range terms {
		if strings.Contains(text, term) {
			return true
		}
	}
	return false
}
