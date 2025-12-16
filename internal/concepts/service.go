package concepts

import (
	"context"

	"github.com/gongahkia/kite/internal/storage"
	"github.com/gongahkia/kite/pkg/models"
)

// Service provides legal concept extraction and analysis services
type Service struct {
	taxonomy  *Taxonomy
	extractor *Extractor
	storage   storage.Storage
}

// NewService creates a new concept service
func NewService(store storage.Storage) *Service {
	taxonomy := NewTaxonomy()
	extractor := NewExtractor(taxonomy)

	return &Service{
		taxonomy:  taxonomy,
		extractor: extractor,
		storage:   store,
	}
}

// ExtractAndStoreConcepts extracts concepts from a case and stores them
func (s *Service) ExtractAndStoreConcepts(ctx context.Context, c *models.Case) ([]models.ConceptMatch, error) {
	// Extract concepts
	matches := s.extractor.ExtractConceptsFromCase(ctx, c)

	// Update case with extracted concepts
	c.LegalConcepts = make([]string, 0, len(matches))
	for _, match := range matches {
		c.LegalConcepts = append(c.LegalConcepts, match.Name)
	}

	// Store updated case
	if err := s.storage.UpdateCase(ctx, c.ID, c); err != nil {
		return nil, err
	}

	return matches, nil
}

// ExtractConcepts extracts concepts from text
func (s *Service) ExtractConcepts(ctx context.Context, text string) []models.ConceptMatch {
	return s.extractor.ExtractConcepts(ctx, text)
}

// GetTaxonomy returns the taxonomy
func (s *Service) GetTaxonomy() *Taxonomy {
	return s.taxonomy
}

// GetConcept retrieves a concept by ID
func (s *Service) GetConcept(id string) (*models.LegalConcept, bool) {
	return s.taxonomy.GetConcept(id)
}

// GetConceptsByArea retrieves concepts for a specific area of law
func (s *Service) GetConceptsByArea(area models.AreaOfLaw) []*models.LegalConcept {
	return s.taxonomy.GetConceptsByArea(area)
}

// SearchConcepts searches for concepts by keyword
func (s *Service) SearchConcepts(keyword string) []*models.LegalConcept {
	return s.taxonomy.SearchByKeyword(keyword)
}

// GetTaxonomyStats returns statistics about the taxonomy
func (s *Service) GetTaxonomyStats() map[string]int {
	return s.taxonomy.GetStats()
}

// AnalyzeConceptDistribution analyzes concept distribution across cases
func (s *Service) AnalyzeConceptDistribution(ctx context.Context) (map[models.AreaOfLaw]int, error) {
	// Get all cases
	cases, err := s.storage.ListCases(ctx, storage.CaseFilter{})
	if err != nil {
		return nil, err
	}

	distribution := make(map[models.AreaOfLaw]int)

	// Extract concepts from all cases
	for _, c := range cases {
		matches := s.extractor.ExtractConceptsFromCase(ctx, c)
		for _, match := range matches {
			distribution[match.Area]++
		}
	}

	return distribution, nil
}

// GetRelatedConcepts finds concepts related to a given concept
func (s *Service) GetRelatedConcepts(conceptID string) []*models.LegalConcept {
	concept, exists := s.taxonomy.GetConcept(conceptID)
	if !exists {
		return nil
	}

	// Get all concepts in the same area
	related := s.taxonomy.GetConceptsByArea(concept.Area)

	// Filter out the original concept
	filtered := make([]*models.LegalConcept, 0)
	for _, c := range related {
		if c.ID != conceptID {
			filtered = append(filtered, c)
		}
	}

	return filtered
}

// BatchExtractConcepts extracts concepts from multiple cases concurrently
func (s *Service) BatchExtractConcepts(ctx context.Context, cases []*models.Case) ([][]models.ConceptMatch, error) {
	texts := make([]string, len(cases))
	for i, c := range cases {
		texts[i] = c.FullText + " " + c.Summary
	}

	matches := s.extractor.ExtractConceptsConcurrent(ctx, texts)

	// Update cases with extracted concepts
	for i, caseMatches := range matches {
		cases[i].LegalConcepts = make([]string, 0, len(caseMatches))
		for _, match := range caseMatches {
			cases[i].LegalConcepts = append(cases[i].LegalConcepts, match.Name)
		}

		// Store updated case
		if err := s.storage.UpdateCase(ctx, cases[i].ID, cases[i]); err != nil {
			// Log error but continue
			continue
		}
	}

	return matches, nil
}

// SetMinConfidenceScore sets the minimum confidence threshold for extraction
func (s *Service) SetMinConfidenceScore(score float64) {
	s.extractor.SetMinScore(score)
}
