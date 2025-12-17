package repository

import (
	"context"

	"github.com/gongahkia/kite/internal/storage"
	"github.com/gongahkia/kite/pkg/models"
)

// CitationRepository provides business logic layer for citation operations
type CitationRepository struct {
	storage storage.Storage
}

// NewCitationRepository creates a new citation repository
func NewCitationRepository(store storage.Storage) *CitationRepository {
	return &CitationRepository{
		storage: store,
	}
}

// FindByID retrieves a citation by ID
func (r *CitationRepository) FindByID(ctx context.Context, id string) (*models.Citation, error) {
	return r.storage.GetCitation(ctx, id)
}

// Save creates a new citation
func (r *CitationRepository) Save(ctx context.Context, c *models.Citation) error {
	return r.storage.SaveCitation(ctx, c)
}

// FindByCaseID retrieves all citations for a specific case
func (r *CitationRepository) FindByCaseID(ctx context.Context, caseID string) ([]*models.Citation, error) {
	filter := storage.CitationFilter{
		CaseID: caseID,
	}
	return r.storage.ListCitations(ctx, filter)
}

// FindByType retrieves citations by type
func (r *CitationRepository) FindByType(ctx context.Context, citationType string) ([]*models.Citation, error) {
	filter := storage.CitationFilter{
		Type: citationType,
	}
	return r.storage.ListCitations(ctx, filter)
}
