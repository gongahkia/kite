package repository

import (
	"context"

	"github.com/gongahkia/kite/internal/storage"
	"github.com/gongahkia/kite/pkg/models"
)

// CaseRepository provides business logic layer for case operations
type CaseRepository struct {
	storage storage.Storage
}

// NewCaseRepository creates a new case repository
func NewCaseRepository(store storage.Storage) *CaseRepository {
	return &CaseRepository{
		storage: store,
	}
}

// FindByID retrieves a case by ID with business logic
func (r *CaseRepository) FindByID(ctx context.Context, id string) (*models.Case, error) {
	return r.storage.GetCase(ctx, id)
}

// Search performs a case search with business logic and filtering
func (r *CaseRepository) Search(ctx context.Context, query string, opts SearchOptions) ([]*models.Case, error) {
	searchQuery := storage.SearchQuery{
		Query:        query,
		Jurisdiction: opts.Jurisdiction,
		Court:        opts.Court,
		DateFrom:     opts.DateFrom,
		DateTo:       opts.DateTo,
		Limit:        opts.Limit,
		Offset:       opts.Offset,
	}

	return r.storage.SearchCases(ctx, searchQuery)
}

// Save creates or updates a case
func (r *CaseRepository) Save(ctx context.Context, c *models.Case) error {
	// Perform validation or business logic here
	return r.storage.SaveCase(ctx, c)
}

// Delete removes a case
func (r *CaseRepository) Delete(ctx context.Context, id string) error {
	return r.storage.DeleteCase(ctx, id)
}

// ListByJurisdiction retrieves cases for a specific jurisdiction
func (r *CaseRepository) ListByJurisdiction(ctx context.Context, jurisdiction string, limit, offset int) ([]*models.Case, error) {
	filter := storage.CaseFilter{
		Jurisdiction: jurisdiction,
		Limit:        limit,
		Offset:       offset,
	}
	return r.storage.ListCases(ctx, filter)
}

// SearchOptions holds options for case search
type SearchOptions struct {
	Jurisdiction string
	Court        string
	DateFrom     *string
	DateTo       *string
	Limit        int
	Offset       int
}
