package citation

import (
	"context"

	"github.com/gongahkia/kite/internal/storage"
	"github.com/gongahkia/kite/pkg/models"
)

// Service provides citation extraction and analysis services
type Service struct {
	extractor *Extractor
	normalizer *Normalizer
	analyzer  *NetworkAnalyzer
	storage   storage.Storage
}

// NewService creates a new citation service
func NewService(store storage.Storage) *Service {
	return &Service{
		extractor:  NewExtractor(),
		normalizer: NewNormalizer(),
		analyzer:   NewNetworkAnalyzer(),
		storage:    store,
	}
}

// ExtractAndStoreCitations extracts citations from a case and stores them
func (s *Service) ExtractAndStoreCitations(ctx context.Context, c *models.Case) ([]*models.Citation, error) {
	// Extract citations
	citations := s.extractor.ExtractCitationsFromCase(c)

	// Normalize citations
	normalizedCitations := s.normalizer.NormalizeBatch(citations)

	// Store citations
	for _, citation := range normalizedCitations {
		if err := s.storage.CreateCitation(ctx, citation); err != nil {
			// Log error but continue with other citations
			continue
		}
	}

	return normalizedCitations, nil
}

// ExtractCitationsFromText extracts citations from raw text
func (s *Service) ExtractCitationsFromText(text string) []*models.Citation {
	citations := s.extractor.ExtractCitations(text)
	return s.normalizer.NormalizeBatch(citations)
}

// BuildCitationNetwork builds a citation network from stored cases
func (s *Service) BuildCitationNetwork(ctx context.Context) (*models.CitationNetwork, error) {
	// Get all cases
	cases, err := s.storage.ListCases(ctx, storage.CaseFilter{})
	if err != nil {
		return nil, err
	}

	// Get all citations
	citations, err := s.storage.ListCitations(ctx, storage.CitationFilter{})
	if err != nil {
		return nil, err
	}

	// Build network
	network := s.analyzer.BuildNetwork(cases, citations)

	return network, nil
}

// GetMostCitedCases returns the most cited cases
func (s *Service) GetMostCitedCases(ctx context.Context, limit int) ([]*models.CitationNode, error) {
	// Build network
	if _, err := s.BuildCitationNetwork(ctx); err != nil {
		return nil, err
	}

	// Get most cited
	return s.analyzer.GetMostCitedCases(limit), nil
}

// GetCitationChain finds the citation chain between two cases
func (s *Service) GetCitationChain(ctx context.Context, fromCaseID, toCaseID string) ([]string, error) {
	// Build network
	if _, err := s.BuildCitationNetwork(ctx); err != nil {
		return nil, err
	}

	// Find chain
	return s.analyzer.GetCitationChain(fromCaseID, toCaseID), nil
}

// GetNetworkStatistics returns citation network statistics
func (s *Service) GetNetworkStatistics(ctx context.Context) (map[string]interface{}, error) {
	// Build network
	if _, err := s.BuildCitationNetwork(ctx); err != nil {
		return nil, err
	}

	return s.analyzer.GetNetworkStatistics(), nil
}

// ValidateCitation validates a citation format
func (s *Service) ValidateCitation(citation *models.Citation) bool {
	return s.extractor.validateCitation(citation)
}

// NormalizeCitation normalizes a single citation
func (s *Service) NormalizeCitation(citation *models.Citation) *models.Citation {
	return s.normalizer.Normalize(citation)
}
