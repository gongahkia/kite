package concepts

import (
	"context"
	"strings"
	"sync"

	"github.com/gongahkia/kite/pkg/models"
)

// Extractor extracts legal concepts from case text
type Extractor struct {
	taxonomy *Taxonomy
	minScore float64 // Minimum confidence score to include concept
}

// NewExtractor creates a new concept extractor
func NewExtractor(taxonomy *Taxonomy) *Extractor {
	return &Extractor{
		taxonomy: taxonomy,
		minScore: 0.3, // Default minimum confidence threshold
	}
}

// SetMinScore sets the minimum confidence score threshold
func (e *Extractor) SetMinScore(score float64) {
	e.minScore = score
}

// ExtractConcepts extracts legal concepts from text
func (e *Extractor) ExtractConcepts(ctx context.Context, text string) []models.ConceptMatch {
	// Normalize text
	text = strings.ToLower(text)
	words := tokenize(text)

	matches := make([]models.ConceptMatch, 0)
	seen := make(map[string]bool)

	// Check all concepts in taxonomy
	for _, concept := range e.taxonomy.GetAllConcepts() {
		if ctx.Err() != nil {
			break // Context cancelled
		}

		// Skip if already matched
		if seen[concept.ID] {
			continue
		}

		// Calculate match score
		score := e.calculateMatchScore(concept, text, words)

		if score >= e.minScore {
			matches = append(matches, models.ConceptMatch{
				ConceptID:  concept.ID,
				Name:       concept.Name,
				Area:       concept.Area,
				Confidence: score,
			})
			seen[concept.ID] = true
		}
	}

	// Sort by confidence (descending)
	sortByConfidence(matches)

	return matches
}

// ExtractConceptsFromCase extracts concepts from a case
func (e *Extractor) ExtractConceptsFromCase(ctx context.Context, c *models.Case) []models.ConceptMatch {
	// Combine all text fields
	text := strings.Join([]string{
		c.CaseName,
		c.Summary,
		c.FullText,
		strings.Join(c.KeyIssues, " "),
	}, " ")

	return e.ExtractConcepts(ctx, text)
}

// ExtractConceptsConcurrent extracts concepts from multiple texts concurrently
func (e *Extractor) ExtractConceptsConcurrent(ctx context.Context, texts []string) [][]models.ConceptMatch {
	results := make([][]models.ConceptMatch, len(texts))
	var wg sync.WaitGroup

	// Process texts concurrently
	for i, text := range texts {
		wg.Add(1)
		go func(index int, txt string) {
			defer wg.Done()
			results[index] = e.ExtractConcepts(ctx, txt)
		}(i, text)
	}

	wg.Wait()
	return results
}

// calculateMatchScore calculates confidence score for a concept match
func (e *Extractor) calculateMatchScore(concept *models.LegalConcept, text string, words []string) float64 {
	if len(concept.Keywords) == 0 {
		return 0.0
	}

	totalWeight := 0.0
	matchedWeight := 0.0
	wordSet := toSet(words)

	for _, keyword := range concept.Keywords {
		keyword = strings.ToLower(keyword)
		keywordWords := tokenize(keyword)

		// Calculate weight based on keyword length (longer = more specific = higher weight)
		weight := float64(len(keywordWords))

		totalWeight += weight

		// Check for exact phrase match (highest confidence)
		if strings.Contains(text, keyword) {
			matchedWeight += weight * 1.5 // Bonus for exact phrase match
		} else {
			// Check for partial match (all words present)
			allWordsPresent := true
			for _, kw := range keywordWords {
				if !wordSet[kw] {
					allWordsPresent = false
					break
				}
			}
			if allWordsPresent {
				matchedWeight += weight
			}
		}
	}

	if totalWeight == 0 {
		return 0.0
	}

	// Calculate base score
	score := matchedWeight / totalWeight

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	// Adjust score based on concept importance
	importanceBoost := float64(concept.Importance) / 10.0 * 0.1
	score = score*(1.0-importanceBoost) + importanceBoost

	return score
}

// tokenize splits text into words
func tokenize(text string) []string {
	// Simple tokenization (could be improved with proper NLP)
	text = strings.ToLower(text)

	// Replace punctuation with spaces
	replacer := strings.NewReplacer(
		".", " ",
		",", " ",
		":", " ",
		";", " ",
		"!", " ",
		"?", " ",
		"(", " ",
		")", " ",
		"[", " ",
		"]", " ",
		"{", " ",
		"}", " ",
		"\"", " ",
		"'", " ",
		"\n", " ",
		"\t", " ",
	)
	text = replacer.Replace(text)

	// Split on whitespace
	words := strings.Fields(text)

	return words
}

// toSet converts a slice to a set (map)
func toSet(words []string) map[string]bool {
	set := make(map[string]bool)
	for _, word := range words {
		set[strings.ToLower(word)] = true
	}
	return set
}

// sortByConfidence sorts concept matches by confidence (descending)
func sortByConfidence(matches []models.ConceptMatch) {
	// Simple bubble sort (for small arrays this is fine)
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].Confidence > matches[i].Confidence {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}
}

// GetAreaDistribution calculates distribution of concepts across areas of law
func (e *Extractor) GetAreaDistribution(matches []models.ConceptMatch) map[models.AreaOfLaw]int {
	distribution := make(map[models.AreaOfLaw]int)

	for _, match := range matches {
		distribution[match.Area]++
	}

	return distribution
}

// GetTopConcepts returns the top N concepts by confidence
func (e *Extractor) GetTopConcepts(matches []models.ConceptMatch, n int) []models.ConceptMatch {
	if n > len(matches) {
		n = len(matches)
	}
	return matches[:n]
}
