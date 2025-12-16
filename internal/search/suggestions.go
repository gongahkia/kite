package search

import (
	"context"
	"sort"
	"strings"

	"github.com/gongahkia/kite/internal/storage"
	"github.com/gongahkia/kite/pkg/models"
)

// Suggestion represents a query suggestion
type Suggestion struct {
	Text  string
	Score float64
	Type  string // "case_name", "judge", "concept", "court"
}

// SuggestionEngine provides query suggestions and autocomplete
type SuggestionEngine struct {
	storage storage.Storage
}

// NewSuggestionEngine creates a new suggestion engine
func NewSuggestionEngine(storage storage.Storage) *SuggestionEngine {
	return &SuggestionEngine{
		storage: storage,
	}
}

// Suggest provides query suggestions based on partial input
func (se *SuggestionEngine) Suggest(ctx context.Context, partial string, limit int) ([]*Suggestion, error) {
	if len(partial) < 2 {
		return []*Suggestion{}, nil
	}

	suggestions := make([]*Suggestion, 0)

	// Get case name suggestions
	caseNameSuggestions, err := se.suggestCaseNames(ctx, partial, limit)
	if err == nil {
		suggestions = append(suggestions, caseNameSuggestions...)
	}

	// Get judge suggestions
	judgeSuggestions, err := se.suggestJudges(ctx, partial, limit)
	if err == nil {
		suggestions = append(suggestions, judgeSuggestions...)
	}

	// Get court suggestions
	courtSuggestions, err := se.suggestCourts(ctx, partial, limit)
	if err == nil {
		suggestions = append(suggestions, courtSuggestions...)
	}

	// Get concept suggestions
	conceptSuggestions := se.suggestConcepts(partial, limit)
	suggestions = append(suggestions, conceptSuggestions...)

	// Sort by score
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Score > suggestions[j].Score
	})

	// Limit results
	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	return suggestions, nil
}

// suggestCaseNames suggests case names
func (se *SuggestionEngine) suggestCaseNames(ctx context.Context, partial string, limit int) ([]*Suggestion, error) {
	// Simple prefix search for case names
	filter := storage.CaseFilter{
		Limit: limit * 2,
	}

	cases, err := se.storage.ListCases(ctx, filter)
	if err != nil {
		return nil, err
	}

	suggestions := make([]*Suggestion, 0)
	partialLower := strings.ToLower(partial)

	for _, c := range cases {
		if strings.Contains(strings.ToLower(c.CaseName), partialLower) {
			score := calculateSimilarity(partialLower, strings.ToLower(c.CaseName))
			suggestions = append(suggestions, &Suggestion{
				Text:  c.CaseName,
				Score: score * 2.0, // Boost case names
				Type:  "case_name",
			})
		}
	}

	return suggestions, nil
}

// suggestJudges suggests judge names
func (se *SuggestionEngine) suggestJudges(ctx context.Context, partial string, limit int) ([]*Suggestion, error) {
	// Get all judges
	judges, err := se.storage.ListJudges(ctx, storage.JudgeFilter{Limit: limit * 5})
	if err != nil {
		return nil, err
	}

	suggestions := make([]*Suggestion, 0)
	partialLower := strings.ToLower(partial)

	for _, j := range judges {
		if strings.Contains(strings.ToLower(j.Name), partialLower) {
			score := calculateSimilarity(partialLower, strings.ToLower(j.Name))
			suggestions = append(suggestions, &Suggestion{
				Text:  j.Name,
				Score: score * 1.5, // Boost judges
				Type:  "judge",
			})
		}
	}

	return suggestions, nil
}

// suggestCourts suggests court names
func (se *SuggestionEngine) suggestCourts(ctx context.Context, partial string, limit int) ([]*Suggestion, error) {
	// Common courts (this could be cached)
	commonCourts := []string{
		"Supreme Court",
		"Court of Appeal",
		"High Court",
		"District Court",
		"Magistrates Court",
		"Federal Court",
		"Circuit Court",
		"Crown Court",
		"County Court",
	}

	suggestions := make([]*Suggestion, 0)
	partialLower := strings.ToLower(partial)

	for _, court := range commonCourts {
		if strings.Contains(strings.ToLower(court), partialLower) {
			score := calculateSimilarity(partialLower, strings.ToLower(court))
			suggestions = append(suggestions, &Suggestion{
				Text:  court,
				Score: score,
				Type:  "court",
			})
		}
	}

	return suggestions, nil
}

// suggestConcepts suggests legal concepts
func (se *SuggestionEngine) suggestConcepts(partial string, limit int) []*Suggestion {
	// Common legal concepts (this should ideally come from a taxonomy)
	commonConcepts := []string{
		"constitutional law",
		"contract law",
		"criminal law",
		"tort law",
		"property law",
		"family law",
		"employment law",
		"intellectual property",
		"administrative law",
		"tax law",
		"due process",
		"judicial review",
		"habeas corpus",
		"negligence",
		"breach of contract",
		"damages",
		"injunction",
		"precedent",
		"stare decisis",
		"jurisdiction",
	}

	suggestions := make([]*Suggestion, 0)
	partialLower := strings.ToLower(partial)

	for _, concept := range commonConcepts {
		if strings.Contains(concept, partialLower) {
			score := calculateSimilarity(partialLower, concept)
			suggestions = append(suggestions, &Suggestion{
				Text:  concept,
				Score: score,
				Type:  "concept",
			})
		}
	}

	return suggestions
}

// calculateSimilarity calculates string similarity (simple version)
func calculateSimilarity(a, b string) float64 {
	// Simple prefix-based similarity
	if strings.HasPrefix(b, a) {
		return 1.0 / float64(len(b)-len(a)+1)
	}

	// Contains-based similarity
	if strings.Contains(b, a) {
		return 0.5 / float64(len(b))
	}

	return 0.1
}

// SearchHistory tracks search queries
type SearchHistory struct {
	UserID    string
	Query     string
	Timestamp int64
	Results   int
}

// HistoryManager manages search history
type HistoryManager struct {
	storage storage.Storage
}

// NewHistoryManager creates a new history manager
func NewHistoryManager(storage storage.Storage) *HistoryManager {
	return &HistoryManager{
		storage: storage,
	}
}

// RecordSearch records a search query
func (hm *HistoryManager) RecordSearch(ctx context.Context, userID, query string, resultCount int) error {
	// This would store search history in a separate table
	// For now, we'll just return nil
	// TODO: Implement search history storage
	return nil
}

// GetHistory retrieves search history for a user
func (hm *HistoryManager) GetHistory(ctx context.Context, userID string, limit int) ([]*SearchHistory, error) {
	// TODO: Implement search history retrieval
	return []*SearchHistory{}, nil
}

// GetPopularQueries returns popular search queries
func (hm *HistoryManager) GetPopularQueries(ctx context.Context, limit int) ([]string, error) {
	// TODO: Implement popular queries retrieval
	return []string{}, nil
}

// SpellChecker provides spell checking for queries
type SpellChecker struct {
	dictionary map[string]bool
}

// NewSpellChecker creates a new spell checker
func NewSpellChecker() *SpellChecker {
	return &SpellChecker{
		dictionary: make(map[string]bool),
	}
}

// LoadDictionary loads legal terms into the dictionary
func (sc *SpellChecker) LoadDictionary(terms []string) {
	for _, term := range terms {
		sc.dictionary[strings.ToLower(term)] = true
	}
}

// Check checks if a term is spelled correctly
func (sc *SpellChecker) Check(term string) bool {
	return sc.dictionary[strings.ToLower(term)]
}

// Suggest suggests corrections for a misspelled term
func (sc *SpellChecker) Suggest(term string) []string {
	if sc.Check(term) {
		return []string{term}
	}

	suggestions := make([]string, 0)
	termLower := strings.ToLower(term)

	// Find similar terms in dictionary
	for dictTerm := range sc.dictionary {
		if levenshteinDistance(termLower, dictTerm) <= 2 {
			suggestions = append(suggestions, dictTerm)
		}
	}

	return suggestions
}

// levenshteinDistance calculates Levenshtein distance between two strings
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}

	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}

			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

func min(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

// QueryExpander expands queries with synonyms
type QueryExpander struct {
	synonyms map[string][]string
}

// NewQueryExpander creates a new query expander
func NewQueryExpander() *QueryExpander {
	qe := &QueryExpander{
		synonyms: make(map[string][]string),
	}

	// Load default legal synonyms
	qe.loadDefaultSynonyms()

	return qe
}

// loadDefaultSynonyms loads default legal term synonyms
func (qe *QueryExpander) loadDefaultSynonyms() {
	qe.synonyms = map[string][]string{
		"contract":    {"agreement", "covenant", "accord"},
		"negligence":  {"carelessness", "neglect"},
		"damages":     {"compensation", "restitution", "reparation"},
		"plaintiff":   {"claimant", "complainant"},
		"defendant":   {"respondent", "accused"},
		"appeal":      {"review", "reconsideration"},
		"judgment":    {"decision", "ruling", "verdict"},
		"precedent":   {"case law", "authority"},
		"statute":     {"law", "act", "legislation"},
		"regulation":  {"rule", "ordinance"},
		"injunction":  {"restraining order", "prohibition"},
		"liability":   {"responsibility", "obligation"},
		"remedy":      {"relief", "redress"},
		"jurisdiction": {"authority", "power"},
	}
}

// Expand expands a query with synonyms
func (qe *QueryExpander) Expand(query string) []string {
	words := strings.Fields(strings.ToLower(query))
	expanded := make([]string, 0)

	for _, word := range words {
		expanded = append(expanded, word)

		if synonyms, ok := qe.synonyms[word]; ok {
			expanded = append(expanded, synonyms...)
		}
	}

	return expanded
}

// AddSynonym adds a synonym relationship
func (qe *QueryExpander) AddSynonym(term string, synonyms ...string) {
	qe.synonyms[strings.ToLower(term)] = synonyms
}
