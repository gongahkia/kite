package search

import (
	"fmt"
	"strings"
	"time"

	"github.com/gongahkia/kite/pkg/models"
)

// QueryType represents the type of search query
type QueryType string

const (
	QueryTypeFullText QueryType = "fulltext"
	QueryTypeExact    QueryType = "exact"
	QueryTypeFuzzy    QueryType = "fuzzy"
	QueryTypeRegex    QueryType = "regex"
)

// Query represents a structured search query
type Query struct {
	Type    QueryType
	Text    string
	Fields  []string
	Filters *Filters
	Sort    *SortOptions
	Page    *Pagination
	Facets  []string
}

// Filters represents search filters
type Filters struct {
	IDs          []string
	Jurisdiction *string
	Court        *string
	CourtLevel   *models.CourtLevel
	Status       *models.CaseStatus
	StartDate    *time.Time
	EndDate      *time.Time
	Judges       []string
	Parties      []string
	Concepts     []string
	MinQuality   *float64
	HasPDF       *bool
}

// SortOptions represents sorting options
type SortOptions struct {
	Field string
	Desc  bool
}

// Pagination represents pagination options
type Pagination struct {
	Limit  int
	Offset int
	Cursor *string
}

// QueryBuilder builds search queries using fluent API
type QueryBuilder struct {
	query *Query
}

// NewQuery creates a new query builder
func NewQuery() *QueryBuilder {
	return &QueryBuilder{
		query: &Query{
			Type:    QueryTypeFullText,
			Fields:  []string{},
			Filters: &Filters{},
			Sort:    &SortOptions{},
			Page: &Pagination{
				Limit:  10,
				Offset: 0,
			},
		},
	}
}

// FullText sets full-text search query
func (qb *QueryBuilder) FullText(text string) *QueryBuilder {
	qb.query.Type = QueryTypeFullText
	qb.query.Text = text
	return qb
}

// Exact sets exact match query
func (qb *QueryBuilder) Exact(text string) *QueryBuilder {
	qb.query.Type = QueryTypeExact
	qb.query.Text = text
	return qb
}

// Fuzzy sets fuzzy match query
func (qb *QueryBuilder) Fuzzy(text string) *QueryBuilder {
	qb.query.Type = QueryTypeFuzzy
	qb.query.Text = text
	return qb
}

// Regex sets regex match query
func (qb *QueryBuilder) Regex(pattern string) *QueryBuilder {
	qb.query.Type = QueryTypeRegex
	qb.query.Text = pattern
	return qb
}

// InFields specifies which fields to search
func (qb *QueryBuilder) InFields(fields ...string) *QueryBuilder {
	qb.query.Fields = fields
	return qb
}

// FilterByID filters by case IDs
func (qb *QueryBuilder) FilterByID(ids ...string) *QueryBuilder {
	qb.query.Filters.IDs = ids
	return qb
}

// FilterByJurisdiction filters by jurisdiction
func (qb *QueryBuilder) FilterByJurisdiction(jurisdiction string) *QueryBuilder {
	qb.query.Filters.Jurisdiction = &jurisdiction
	return qb
}

// FilterByCourt filters by court
func (qb *QueryBuilder) FilterByCourt(court string) *QueryBuilder {
	qb.query.Filters.Court = &court
	return qb
}

// FilterByCourtLevel filters by court level
func (qb *QueryBuilder) FilterByCourtLevel(level models.CourtLevel) *QueryBuilder {
	qb.query.Filters.CourtLevel = &level
	return qb
}

// FilterByStatus filters by case status
func (qb *QueryBuilder) FilterByStatus(status models.CaseStatus) *QueryBuilder {
	qb.query.Filters.Status = &status
	return qb
}

// FilterByDateRange filters by decision date range
func (qb *QueryBuilder) FilterByDateRange(start, end time.Time) *QueryBuilder {
	qb.query.Filters.StartDate = &start
	qb.query.Filters.EndDate = &end
	return qb
}

// FilterByJudges filters by judges
func (qb *QueryBuilder) FilterByJudges(judges ...string) *QueryBuilder {
	qb.query.Filters.Judges = judges
	return qb
}

// FilterByParties filters by parties
func (qb *QueryBuilder) FilterByParties(parties ...string) *QueryBuilder {
	qb.query.Filters.Parties = parties
	return qb
}

// FilterByConcepts filters by legal concepts
func (qb *QueryBuilder) FilterByConcepts(concepts ...string) *QueryBuilder {
	qb.query.Filters.Concepts = concepts
	return qb
}

// FilterByMinQuality filters by minimum quality score
func (qb *QueryBuilder) FilterByMinQuality(score float64) *QueryBuilder {
	qb.query.Filters.MinQuality = &score
	return qb
}

// FilterByHasPDF filters cases with/without PDF
func (qb *QueryBuilder) FilterByHasPDF(hasPDF bool) *QueryBuilder {
	qb.query.Filters.HasPDF = &hasPDF
	return qb
}

// SortBy sets sorting options
func (qb *QueryBuilder) SortBy(field string, desc bool) *QueryBuilder {
	qb.query.Sort.Field = field
	qb.query.Sort.Desc = desc
	return qb
}

// SortByRelevance sorts by relevance (default for full-text search)
func (qb *QueryBuilder) SortByRelevance() *QueryBuilder {
	qb.query.Sort.Field = "relevance"
	qb.query.Sort.Desc = true
	return qb
}

// SortByDate sorts by decision date
func (qb *QueryBuilder) SortByDate(desc bool) *QueryBuilder {
	qb.query.Sort.Field = "decision_date"
	qb.query.Sort.Desc = desc
	return qb
}

// SortByQuality sorts by quality score
func (qb *QueryBuilder) SortByQuality(desc bool) *QueryBuilder {
	qb.query.Sort.Field = "quality_score"
	qb.query.Sort.Desc = desc
	return qb
}

// Limit sets result limit
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.query.Page.Limit = limit
	return qb
}

// Offset sets result offset
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.query.Page.Offset = offset
	return qb
}

// WithCursor sets cursor for pagination
func (qb *QueryBuilder) WithCursor(cursor string) *QueryBuilder {
	qb.query.Page.Cursor = &cursor
	return qb
}

// WithFacets requests faceted search results
func (qb *QueryBuilder) WithFacets(facets ...string) *QueryBuilder {
	qb.query.Facets = facets
	return qb
}

// Build returns the constructed query
func (qb *QueryBuilder) Build() *Query {
	return qb.query
}

// String returns a human-readable representation of the query
func (q *Query) String() string {
	var parts []string

	parts = append(parts, fmt.Sprintf("Type: %s", q.Type))

	if q.Text != "" {
		parts = append(parts, fmt.Sprintf("Text: %q", q.Text))
	}

	if len(q.Fields) > 0 {
		parts = append(parts, fmt.Sprintf("Fields: %v", q.Fields))
	}

	if q.Filters != nil {
		if q.Filters.Jurisdiction != nil {
			parts = append(parts, fmt.Sprintf("Jurisdiction: %s", *q.Filters.Jurisdiction))
		}
		if q.Filters.Court != nil {
			parts = append(parts, fmt.Sprintf("Court: %s", *q.Filters.Court))
		}
		if len(q.Filters.Judges) > 0 {
			parts = append(parts, fmt.Sprintf("Judges: %v", q.Filters.Judges))
		}
	}

	if q.Sort != nil && q.Sort.Field != "" {
		order := "ASC"
		if q.Sort.Desc {
			order = "DESC"
		}
		parts = append(parts, fmt.Sprintf("Sort: %s %s", q.Sort.Field, order))
	}

	if q.Page != nil {
		parts = append(parts, fmt.Sprintf("Limit: %d, Offset: %d", q.Page.Limit, q.Page.Offset))
	}

	return strings.Join(parts, " | ")
}

// Validate validates the query
func (q *Query) Validate() error {
	if q.Text == "" && len(q.Filters.IDs) == 0 {
		return fmt.Errorf("query must have either text or ID filters")
	}

	if q.Page != nil {
		if q.Page.Limit < 1 || q.Page.Limit > 1000 {
			return fmt.Errorf("limit must be between 1 and 1000")
		}
		if q.Page.Offset < 0 {
			return fmt.Errorf("offset must be non-negative")
		}
	}

	return nil
}
