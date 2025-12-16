package storage

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gongahkia/kite/pkg/errors"
	"github.com/gongahkia/kite/pkg/models"
)

// MemoryStorage is an in-memory implementation of the Storage interface
type MemoryStorage struct {
	cases     map[string]*models.Case
	judges    map[string]*models.Judge
	citations map[string]*models.Citation
	mu        sync.RWMutex
}

// NewMemoryStorage creates a new MemoryStorage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		cases:     make(map[string]*models.Case),
		judges:    make(map[string]*models.Judge),
		citations: make(map[string]*models.Citation),
	}
}

// SaveCase saves a case
func (ms *MemoryStorage) SaveCase(ctx context.Context, c *models.Case) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.cases[c.ID]; exists {
		return errors.StorageError("case already exists", errors.ErrAlreadyExists)
	}

	ms.cases[c.ID] = c
	return nil
}

// GetCase retrieves a case by ID
func (ms *MemoryStorage) GetCase(ctx context.Context, id string) (*models.Case, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	c, ok := ms.cases[id]
	if !ok {
		return nil, errors.StorageError("case not found", errors.ErrNotFound)
	}

	return c, nil
}

// UpdateCase updates a case
func (ms *MemoryStorage) UpdateCase(ctx context.Context, c *models.Case) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.cases[c.ID]; !exists {
		return errors.StorageError("case not found", errors.ErrNotFound)
	}

	c.LastUpdated = time.Now()
	ms.cases[c.ID] = c
	return nil
}

// DeleteCase deletes a case
func (ms *MemoryStorage) DeleteCase(ctx context.Context, id string) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.cases[id]; !exists {
		return errors.StorageError("case not found", errors.ErrNotFound)
	}

	delete(ms.cases, id)
	return nil
}

// ListCases lists cases with filters
func (ms *MemoryStorage) ListCases(ctx context.Context, filter CaseFilter) ([]*models.Case, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var results []*models.Case

	for _, c := range ms.cases {
		if ms.matchesFilter(c, filter) {
			results = append(results, c)
		}
	}

	// Apply limit and offset
	start := filter.Offset
	if start > len(results) {
		start = len(results)
	}

	end := start + filter.Limit
	if filter.Limit == 0 || end > len(results) {
		end = len(results)
	}

	return results[start:end], nil
}

// CountCases counts cases matching the filter
func (ms *MemoryStorage) CountCases(ctx context.Context, filter CaseFilter) (int64, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	count := int64(0)
	for _, c := range ms.cases {
		if ms.matchesFilter(c, filter) {
			count++
		}
	}

	return count, nil
}

// matchesFilter checks if a case matches the filter
func (ms *MemoryStorage) matchesFilter(c *models.Case, filter CaseFilter) bool {
	// Check IDs
	if len(filter.IDs) > 0 {
		found := false
		for _, id := range filter.IDs {
			if c.ID == id {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check jurisdiction
	if filter.Jurisdiction != "" && c.Jurisdiction != filter.Jurisdiction {
		return false
	}

	// Check court
	if filter.Court != "" && c.Court != filter.Court {
		return false
	}

	// Check court level
	if filter.CourtLevel != nil && c.CourtLevel != *filter.CourtLevel {
		return false
	}

	// Check date range
	if filter.StartDate != nil && c.DecisionDate != nil && c.DecisionDate.Before(*filter.StartDate) {
		return false
	}

	if filter.EndDate != nil && c.DecisionDate != nil && c.DecisionDate.After(*filter.EndDate) {
		return false
	}

	// Check status
	if filter.Status != "" && c.Status != filter.Status {
		return false
	}

	// Check judges
	if len(filter.Judges) > 0 {
		found := false
		for _, filterJudge := range filter.Judges {
			for _, caseJudge := range c.Judges {
				if caseJudge == filterJudge {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check concepts
	if len(filter.Concepts) > 0 {
		found := false
		for _, filterConcept := range filter.Concepts {
			for _, caseConcept := range c.LegalConcepts {
				if caseConcept == filterConcept {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check quality score
	if filter.MinQuality > 0 && c.QualityScore < filter.MinQuality {
		return false
	}

	return true
}

// SaveJudge saves a judge
func (ms *MemoryStorage) SaveJudge(ctx context.Context, j *models.Judge) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.judges[j.ID]; exists {
		return errors.StorageError("judge already exists", errors.ErrAlreadyExists)
	}

	ms.judges[j.ID] = j
	return nil
}

// GetJudge retrieves a judge by ID
func (ms *MemoryStorage) GetJudge(ctx context.Context, id string) (*models.Judge, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	j, ok := ms.judges[id]
	if !ok {
		return nil, errors.StorageError("judge not found", errors.ErrNotFound)
	}

	return j, nil
}

// UpdateJudge updates a judge
func (ms *MemoryStorage) UpdateJudge(ctx context.Context, j *models.Judge) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if _, exists := ms.judges[j.ID]; !exists {
		return errors.StorageError("judge not found", errors.ErrNotFound)
	}

	j.UpdatedAt = time.Now()
	ms.judges[j.ID] = j
	return nil
}

// ListJudges lists judges with filters
func (ms *MemoryStorage) ListJudges(ctx context.Context, filter JudgeFilter) ([]*models.Judge, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var results []*models.Judge

	for _, j := range ms.judges {
		match := true

		if filter.Name != "" && !strings.Contains(strings.ToLower(j.Name), strings.ToLower(filter.Name)) {
			match = false
		}

		if filter.Court != "" && j.Court != filter.Court {
			match = false
		}

		if filter.Jurisdiction != "" && j.Jurisdiction != filter.Jurisdiction {
			match = false
		}

		if match {
			results = append(results, j)
		}
	}

	// Apply limit and offset
	start := filter.Offset
	if start > len(results) {
		start = len(results)
	}

	end := start + filter.Limit
	if filter.Limit == 0 || end > len(results) {
		end = len(results)
	}

	return results[start:end], nil
}

// SaveCitation saves a citation
func (ms *MemoryStorage) SaveCitation(ctx context.Context, c *models.Citation) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Generate ID if not set
	id := c.CaseID + "-" + c.RawCitation
	ms.citations[id] = c
	return nil
}

// GetCitation retrieves a citation
func (ms *MemoryStorage) GetCitation(ctx context.Context, id string) (*models.Citation, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	c, ok := ms.citations[id]
	if !ok {
		return nil, errors.StorageError("citation not found", errors.ErrNotFound)
	}

	return c, nil
}

// ListCitations lists citations with filters
func (ms *MemoryStorage) ListCitations(ctx context.Context, filter CitationFilter) ([]*models.Citation, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var results []*models.Citation

	for _, c := range ms.citations {
		match := true

		if filter.CaseID != "" && c.CaseID != filter.CaseID {
			match = false
		}

		if filter.Format != "" && string(c.Format) != filter.Format {
			match = false
		}

		if filter.Year != 0 && c.CaseYear != filter.Year {
			match = false
		}

		if filter.Valid != nil && c.IsValid != *filter.Valid {
			match = false
		}

		if match {
			results = append(results, c)
		}
	}

	// Apply limit and offset
	start := filter.Offset
	if start > len(results) {
		start = len(results)
	}

	end := start + filter.Limit
	if filter.Limit == 0 || end > len(results) {
		end = len(results)
	}

	return results[start:end], nil
}

// SearchCases performs a simple search on cases
func (ms *MemoryStorage) SearchCases(ctx context.Context, query SearchQuery) ([]*models.Case, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	var results []*models.Case
	queryLower := strings.ToLower(query.Query)

	for _, c := range ms.cases {
		// Simple text search in case name and summary
		if strings.Contains(strings.ToLower(c.CaseName), queryLower) ||
			strings.Contains(strings.ToLower(c.Summary), queryLower) {

			// Also check against filters
			if ms.matchesFilter(c, query.Filters) {
				results = append(results, c)
			}
		}
	}

	// Apply limit and offset
	start := query.Offset
	if start > len(results) {
		start = len(results)
	}

	end := start + query.Limit
	if query.Limit == 0 || end > len(results) {
		end = len(results)
	}

	return results[start:end], nil
}

// Ping checks if the storage is available
func (ms *MemoryStorage) Ping(ctx context.Context) error {
	return nil
}

// Close closes the storage connection
func (ms *MemoryStorage) Close() error {
	return nil
}

// GetStats returns storage statistics
func (ms *MemoryStorage) GetStats() StorageStats {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return StorageStats{
		TotalCases:     int64(len(ms.cases)),
		TotalJudges:    int64(len(ms.judges)),
		TotalCitations: int64(len(ms.citations)),
		LastUpdated:    time.Now(),
	}
}

// Clear clears all data from storage
func (ms *MemoryStorage) Clear() {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.cases = make(map[string]*models.Case)
	ms.judges = make(map[string]*models.Judge)
	ms.citations = make(map[string]*models.Citation)
}
