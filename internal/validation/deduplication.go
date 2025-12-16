package validation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gongahkia/kite/pkg/models"
)

// DuplicationValidator detects duplicate cases
type DuplicationValidator struct {
	cache map[string]string // hash -> case_id
	mu    sync.RWMutex
}

func NewDuplicationValidator() *DuplicationValidator {
	return &DuplicationValidator{
		cache: make(map[string]string),
	}
}

func (v *DuplicationValidator) Name() string {
	return "duplication"
}

func (v *DuplicationValidator) Stage() ValidationStage {
	return StageDuplication
}

func (v *DuplicationValidator) Validate(ctx context.Context, c *models.Case) (*ValidationResult, error) {
	start := time.Now()

	result := &ValidationResult{
		Valid:    true,
		Stage:    StageDuplication,
		Score:    1.0,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}

	// Generate hashes for duplicate detection
	hashes := v.generateHashes(c)

	// Check for duplicates
	v.mu.RLock()
	for hashType, hash := range hashes {
		if existingID, exists := v.cache[hash]; exists && existingID != c.ID {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   hashType,
				Message: fmt.Sprintf("Potential duplicate detected (matches case: %s)", existingID),
				Code:    "POTENTIAL_DUPLICATE",
			})
			result.Score -= 0.2
		}
	}
	v.mu.RUnlock()

	// Store hashes for future checks
	v.mu.Lock()
	for _, hash := range hashes {
		v.cache[hash] = c.ID
	}
	v.mu.Unlock()

	if result.Score < 0 {
		result.Score = 0
	}

	result.Duration = time.Since(start)

	return result, nil
}

// generateHashes generates multiple hashes for duplicate detection
func (v *DuplicationValidator) generateHashes(c *models.Case) map[string]string {
	hashes := make(map[string]string)

	// Hash 1: Exact case number match
	if c.CaseNumber != "" {
		hashes["case_number"] = hashString(c.CaseNumber)
	}

	// Hash 2: Case name similarity (normalized)
	if c.CaseName != "" {
		normalized := normalizeText(c.CaseName)
		hashes["case_name"] = hashString(normalized)
	}

	// Hash 3: Court + case number combination
	if c.Court != "" && c.CaseNumber != "" {
		combined := c.Court + "|" + c.CaseNumber
		hashes["court_case_number"] = hashString(combined)
	}

	// Hash 4: Content fingerprint (first 500 chars of summary)
	if c.Summary != "" && len(c.Summary) > 100 {
		fingerprint := c.Summary
		if len(fingerprint) > 500 {
			fingerprint = fingerprint[:500]
		}
		hashes["content_fingerprint"] = hashString(fingerprint)
	}

	// Hash 5: Structural fingerprint
	structural := fmt.Sprintf("%s|%s|%s|%d",
		c.Jurisdiction,
		c.Court,
		c.CaseName,
		c.CourtLevel,
	)
	hashes["structural"] = hashString(structural)

	return hashes
}

// ClearCache clears the duplication cache
func (v *DuplicationValidator) ClearCache() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.cache = make(map[string]string)
}

// GetCacheSize returns the current cache size
func (v *DuplicationValidator) GetCacheSize() int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return len(v.cache)
}

// hashString generates a SHA-256 hash of a string
func hashString(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// normalizeText normalizes text for similarity comparison
func normalizeText(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Remove extra whitespace
	text = strings.Join(strings.Fields(text), " ")

	// Remove common punctuation
	replacements := map[string]string{
		".": "",
		",": "",
		";": "",
		":": "",
		"(": "",
		")": "",
		"[": "",
		"]": "",
		"\"": "",
		"'": "",
	}

	for old, new := range replacements {
		text = strings.ReplaceAll(text, old, new)
	}

	return text
}

// DuplicateDetector provides advanced duplicate detection
type DuplicateDetector struct {
	validators map[string]*DuplicationValidator
	mu         sync.RWMutex
}

// NewDuplicateDetector creates a new duplicate detector
func NewDuplicateDetector() *DuplicateDetector {
	return &DuplicateDetector{
		validators: make(map[string]*DuplicationValidator),
	}
}

// DetectDuplicates detects duplicates across a batch of cases
func (dd *DuplicateDetector) DetectDuplicates(cases []*models.Case) map[string][]string {
	// Map of hash -> list of case IDs
	duplicates := make(map[string][]string)
	hashToCases := make(map[string][]string)

	for _, c := range cases {
		validator := NewDuplicationValidator()
		hashes := validator.generateHashes(c)

		for _, hash := range hashes {
			hashToCases[hash] = append(hashToCases[hash], c.ID)
		}
	}

	// Find duplicates (hashes with multiple case IDs)
	for hash, caseIDs := range hashToCases {
		if len(caseIDs) > 1 {
			duplicates[hash] = caseIDs
		}
	}

	return duplicates
}

// DuplicateGroup represents a group of duplicate cases
type DuplicateGroup struct {
	Hash       string
	CaseIDs    []string
	Similarity float64
	Type       string // "exact", "near", "structural"
}

// FindDuplicateGroups finds and groups duplicate cases
func (dd *DuplicateDetector) FindDuplicateGroups(cases []*models.Case) []*DuplicateGroup {
	groups := make([]*DuplicateGroup, 0)
	duplicates := dd.DetectDuplicates(cases)

	for hash, caseIDs := range duplicates {
		group := &DuplicateGroup{
			Hash:       hash,
			CaseIDs:    caseIDs,
			Similarity: 1.0, // Exact match for hash-based detection
			Type:       "exact",
		}
		groups = append(groups, group)
	}

	return groups
}

// MergeDuplicates suggests which cases to keep from duplicate groups
func (dd *DuplicateDetector) MergeDuplicates(cases []*models.Case, groups []*DuplicateGroup) map[string]string {
	// Returns map of duplicate_id -> primary_id
	mergeMap := make(map[string]string)

	for _, group := range groups {
		if len(group.CaseIDs) < 2 {
			continue
		}

		// Find the best case to keep (highest quality)
		var bestCase *models.Case
		bestScore := 0.0

		casesMap := make(map[string]*models.Case)
		for _, c := range cases {
			casesMap[c.ID] = c
		}

		for _, caseID := range group.CaseIDs {
			c, exists := casesMap[caseID]
			if !exists {
				continue
			}

			score := 0.0
			if c.QualityScore != nil {
				score = *c.QualityScore
			}

			if score > bestScore || bestCase == nil {
				bestScore = score
				bestCase = c
			}
		}

		if bestCase != nil {
			// Mark other cases as duplicates of the best one
			for _, caseID := range group.CaseIDs {
				if caseID != bestCase.ID {
					mergeMap[caseID] = bestCase.ID
				}
			}
		}
	}

	return mergeMap
}

// SimilarityChecker checks text similarity between cases
type SimilarityChecker struct{}

// NewSimilarityChecker creates a new similarity checker
func NewSimilarityChecker() *SimilarityChecker {
	return &SimilarityChecker{}
}

// CalculateSimilarity calculates similarity between two cases (0-1)
func (sc *SimilarityChecker) CalculateSimilarity(c1, c2 *models.Case) float64 {
	scores := make([]float64, 0)

	// Case number similarity
	if c1.CaseNumber != "" && c2.CaseNumber != "" {
		if c1.CaseNumber == c2.CaseNumber {
			scores = append(scores, 1.0)
		} else {
			scores = append(scores, 0.0)
		}
	}

	// Case name similarity (Jaccard)
	if c1.CaseName != "" && c2.CaseName != "" {
		sim := jaccardSimilarity(c1.CaseName, c2.CaseName)
		scores = append(scores, sim)
	}

	// Court similarity
	if c1.Court == c2.Court {
		scores = append(scores, 1.0)
	} else {
		scores = append(scores, 0.0)
	}

	// Jurisdiction similarity
	if c1.Jurisdiction == c2.Jurisdiction {
		scores = append(scores, 1.0)
	} else {
		scores = append(scores, 0.0)
	}

	// Date similarity (within 7 days = similar)
	if c1.DecisionDate != nil && c2.DecisionDate != nil {
		daysDiff := c1.DecisionDate.Sub(*c2.DecisionDate).Hours() / 24
		if daysDiff < 0 {
			daysDiff = -daysDiff
		}
		dateSim := 1.0 - (daysDiff / 365.0)
		if dateSim < 0 {
			dateSim = 0
		}
		scores = append(scores, dateSim)
	}

	// Summary similarity
	if c1.Summary != "" && c2.Summary != "" {
		sim := jaccardSimilarity(c1.Summary, c2.Summary)
		scores = append(scores, sim)
	}

	// Average all similarities
	if len(scores) == 0 {
		return 0.0
	}

	total := 0.0
	for _, s := range scores {
		total += s
	}

	return total / float64(len(scores))
}

// jaccardSimilarity calculates Jaccard similarity between two texts
func jaccardSimilarity(text1, text2 string) float64 {
	words1 := strings.Fields(strings.ToLower(text1))
	words2 := strings.Fields(strings.ToLower(text2))

	// Create sets
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, w := range words1 {
		set1[w] = true
	}

	for _, w := range words2 {
		set2[w] = true
	}

	// Calculate intersection
	intersection := 0
	for w := range set1 {
		if set2[w] {
			intersection++
		}
	}

	// Calculate union
	union := len(set1) + len(set2) - intersection

	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}
