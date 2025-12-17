package repository

import (
	"context"

	"github.com/gongahkia/kite/internal/storage"
	"github.com/gongahkia/kite/pkg/models"
)

// JudgeRepository provides business logic layer for judge operations
type JudgeRepository struct {
	storage storage.Storage
}

// NewJudgeRepository creates a new judge repository
func NewJudgeRepository(store storage.Storage) *JudgeRepository {
	return &JudgeRepository{
		storage: store,
	}
}

// FindByID retrieves a judge by ID
func (r *JudgeRepository) FindByID(ctx context.Context, id string) (*models.Judge, error) {
	return r.storage.GetJudge(ctx, id)
}

// Save creates or updates a judge
func (r *JudgeRepository) Save(ctx context.Context, j *models.Judge) error {
	return r.storage.SaveJudge(ctx, j)
}

// FindByCourt retrieves all judges for a specific court
func (r *JudgeRepository) FindByCourt(ctx context.Context, court string, limit, offset int) ([]*models.Judge, error) {
	filter := storage.JudgeFilter{
		Court:  court,
		Limit:  limit,
		Offset: offset,
	}
	return r.storage.ListJudges(ctx, filter)
}

// FindByName searches for judges by name
func (r *JudgeRepository) FindByName(ctx context.Context, name string) ([]*models.Judge, error) {
	filter := storage.JudgeFilter{
		Name: name,
	}
	return r.storage.ListJudges(ctx, filter)
}
