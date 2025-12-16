package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/gongahkia/kite/internal/observability"
	"github.com/gongahkia/kite/internal/search"
	"github.com/gongahkia/kite/internal/storage"
	"github.com/gongahkia/kite/pkg/models"
)

// SearchHandler handles search requests
type SearchHandler struct {
	engine      *search.SearchEngine
	suggestions *search.SuggestionEngine
	logger      *observability.Logger
	metrics     *observability.Metrics
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(storage storage.Storage, logger *observability.Logger, metrics *observability.Metrics) *SearchHandler {
	return &SearchHandler{
		engine:      search.NewSearchEngine(storage, logger, metrics),
		suggestions: search.NewSuggestionEngine(storage),
		logger:      logger,
		metrics:     metrics,
	}
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query        string   `json:"query"`
	QueryType    string   `json:"query_type,omitempty"`
	Fields       []string `json:"fields,omitempty"`
	Jurisdiction string   `json:"jurisdiction,omitempty"`
	Court        string   `json:"court,omitempty"`
	CourtLevel   int      `json:"court_level,omitempty"`
	StartDate    string   `json:"start_date,omitempty"`
	EndDate      string   `json:"end_date,omitempty"`
	Judges       []string `json:"judges,omitempty"`
	Parties      []string `json:"parties,omitempty"`
	Concepts     []string `json:"concepts,omitempty"`
	MinQuality   float64  `json:"min_quality,omitempty"`
	SortBy       string   `json:"sort_by,omitempty"`
	SortDesc     bool     `json:"sort_desc,omitempty"`
	Limit        int      `json:"limit,omitempty"`
	Offset       int      `json:"offset,omitempty"`
	Facets       []string `json:"facets,omitempty"`
}

// SearchResponse represents a search response
type SearchResponse struct {
	Results    []SearchResult        `json:"results"`
	TotalHits  int                   `json:"total_hits"`
	SearchTime float64               `json:"search_time_ms"`
	Facets     map[string][]FacetVal `json:"facets,omitempty"`
}

// SearchResult represents a single search result
type SearchResult struct {
	Case       *models.Case `json:"case"`
	Score      float64      `json:"score"`
	Highlights []string     `json:"highlights,omitempty"`
}

// FacetVal represents a facet value
type FacetVal struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// Search handles POST /api/v1/search
func (h *SearchHandler) Search(c *fiber.Ctx) error {
	var req SearchRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Build query
	qb := search.NewQuery()

	// Set query type and text
	switch req.QueryType {
	case "exact":
		qb.Exact(req.Query)
	case "fuzzy":
		qb.Fuzzy(req.Query)
	case "regex":
		qb.Regex(req.Query)
	default:
		qb.FullText(req.Query)
	}

	// Set fields
	if len(req.Fields) > 0 {
		qb.InFields(req.Fields...)
	}

	// Set filters
	if req.Jurisdiction != "" {
		qb.FilterByJurisdiction(req.Jurisdiction)
	}

	if req.Court != "" {
		qb.FilterByCourt(req.Court)
	}

	if req.CourtLevel > 0 {
		qb.FilterByCourtLevel(models.CourtLevel(req.CourtLevel))
	}

	if req.StartDate != "" && req.EndDate != "" {
		start, err := time.Parse(time.RFC3339, req.StartDate)
		if err == nil {
			end, err := time.Parse(time.RFC3339, req.EndDate)
			if err == nil {
				qb.FilterByDateRange(start, end)
			}
		}
	}

	if len(req.Judges) > 0 {
		qb.FilterByJudges(req.Judges...)
	}

	if len(req.Parties) > 0 {
		qb.FilterByParties(req.Parties...)
	}

	if len(req.Concepts) > 0 {
		qb.FilterByConcepts(req.Concepts...)
	}

	if req.MinQuality > 0 {
		qb.FilterByMinQuality(req.MinQuality)
	}

	// Set sorting
	if req.SortBy != "" {
		qb.SortBy(req.SortBy, req.SortDesc)
	} else {
		qb.SortByRelevance()
	}

	// Set pagination
	limit := req.Limit
	if limit == 0 {
		limit = 10
	}
	qb.Limit(limit)

	if req.Offset > 0 {
		qb.Offset(req.Offset)
	}

	// Set facets
	if len(req.Facets) > 0 {
		qb.WithFacets(req.Facets...)
	}

	query := qb.Build()

	// Execute search
	results, err := h.engine.Search(c.Context(), query)
	if err != nil {
		h.logger.WithField("error", err).Error("Search failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Search failed",
		})
	}

	// Convert results
	searchResults := make([]SearchResult, len(results.Results))
	for i, r := range results.Results {
		searchResults[i] = SearchResult{
			Case:       r.Case,
			Score:      r.Score,
			Highlights: r.Highlights,
		}
	}

	// Convert facets
	facets := make(map[string][]FacetVal)
	for field, facet := range results.Facets {
		values := make([]FacetVal, len(facet.Values))
		for i, v := range facet.Values {
			values[i] = FacetVal{
				Value: v.Value,
				Count: v.Count,
			}
		}
		facets[field] = values
	}

	return c.JSON(SearchResponse{
		Results:    searchResults,
		TotalHits:  results.TotalHits,
		SearchTime: float64(results.SearchTime.Milliseconds()),
		Facets:     facets,
	})
}

// Suggest handles GET /api/v1/search/suggest?q=...
func (h *SearchHandler) Suggest(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query parameter 'q' is required",
		})
	}

	limit := c.QueryInt("limit", 10)

	suggestions, err := h.suggestions.Suggest(c.Context(), query, limit)
	if err != nil {
		h.logger.WithField("error", err).Error("Suggestion failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Suggestion failed",
		})
	}

	return c.JSON(fiber.Map{
		"suggestions": suggestions,
	})
}

// Autocomplete handles GET /api/v1/search/autocomplete?q=...
func (h *SearchHandler) Autocomplete(c *fiber.Ctx) error {
	return h.Suggest(c) // Same implementation for now
}
