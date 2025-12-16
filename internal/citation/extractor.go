package citation

import (
	"regexp"
	"strings"

	"github.com/gongahkia/kite/pkg/models"
)

// Extractor extracts legal citations from text
type Extractor struct {
	patterns map[models.CitationFormat]*regexp.Regexp
}

// NewExtractor creates a new citation extractor
func NewExtractor() *Extractor {
	return &Extractor{
		patterns: compileCitationPatterns(),
	}
}

// ExtractCitations extracts all citations from the given text
func (e *Extractor) ExtractCitations(text string) []*models.Citation {
	citations := make([]*models.Citation, 0)
	seen := make(map[string]bool)

	// Extract citations for each format
	for format, pattern := range e.patterns {
		matches := pattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 0 {
				rawCitation := strings.TrimSpace(match[0])

				// Skip if already seen (deduplication)
				if seen[rawCitation] {
					continue
				}
				seen[rawCitation] = true

				citation := &models.Citation{
					Format:       format,
					RawCitation:  rawCitation,
					IsNormalized: false,
				}

				// Parse the citation to extract components
				e.parseCitation(citation, match)

				// Validate the citation
				if e.validateCitation(citation) {
					citations = append(citations, citation)
				}
			}
		}
	}

	return citations
}

// ExtractCitationsFromCase extracts citations from a case and updates the case
func (e *Extractor) ExtractCitationsFromCase(c *models.Case) []*models.Citation {
	// Combine all text fields to search for citations
	text := c.FullText + " " + c.Summary

	citations := e.ExtractCitations(text)

	// Link citations to the case
	for _, citation := range citations {
		citation.CitingCaseID = c.ID
	}

	return citations
}

// parseCitation parses the raw citation to extract components
func (e *Extractor) parseCitation(citation *models.Citation, match []string) {
	switch citation.Format {
	case models.CitationFormatBluebook:
		e.parseBluebook(citation, match)
	case models.CitationFormatNeutral:
		e.parseNeutral(citation, match)
	case models.CitationFormatECLI:
		e.parseECLI(citation, match)
	case models.CitationFormatCanadian:
		e.parseCanadian(citation, match)
	case models.CitationFormatUKCitation:
		e.parseUKCitation(citation, match)
	case models.CitationFormatIrishCitation:
		e.parseIrishCitation(citation, match)
	case models.CitationFormatAustralian:
		e.parseAustralian(citation, match)
	case models.CitationFormatOther:
		// Generic parsing
		citation.Year = extractYear(citation.RawCitation)
	}
}

// parseBluebook parses Bluebook-style citations
// Format: Volume Reporter Page (Court Year)
// Example: 410 U.S. 113 (1973)
func (e *Extractor) parseBluebook(citation *models.Citation, match []string) {
	if len(match) >= 5 {
		citation.Volume = strings.TrimSpace(match[1])
		citation.Reporter = strings.TrimSpace(match[2])
		citation.Page = strings.TrimSpace(match[3])
		citation.Year = strings.TrimSpace(match[4])
	}
}

// parseNeutral parses neutral citations
// Format: [Year] Court Number
// Example: [2023] UKSC 15
func (e *Extractor) parseNeutral(citation *models.Citation, match []string) {
	if len(match) >= 3 {
		citation.Year = strings.TrimSpace(match[1])
		citation.Court = strings.TrimSpace(match[2])
		if len(match) >= 4 {
			citation.CaseNumber = strings.TrimSpace(match[3])
		}
	}
}

// parseECLI parses ECLI citations
// Format: ECLI:Country:Court:Year:Number
// Example: ECLI:EU:C:2023:123
func (e *Extractor) parseECLI(citation *models.Citation, match []string) {
	parts := strings.Split(citation.RawCitation, ":")
	if len(parts) >= 5 {
		citation.Country = parts[1]
		citation.Court = parts[2]
		citation.Year = parts[3]
		citation.CaseNumber = parts[4]
	}
}

// parseCanadian parses Canadian neutral citations
// Format: Year CourtCode Number
// Example: 2023 SCC 15
func (e *Extractor) parseCanadian(citation *models.Citation, match []string) {
	if len(match) >= 3 {
		citation.Year = strings.TrimSpace(match[1])
		citation.Court = strings.TrimSpace(match[2])
		if len(match) >= 4 {
			citation.CaseNumber = strings.TrimSpace(match[3])
		}
	}
}

// parseUKCitation parses UK citations
// Format: [Year] Court Number
// Example: [2023] UKSC 15
func (e *Extractor) parseUKCitation(citation *models.Citation, match []string) {
	if len(match) >= 3 {
		citation.Year = strings.TrimSpace(match[1])
		citation.Court = strings.TrimSpace(match[2])
		if len(match) >= 4 {
			citation.CaseNumber = strings.TrimSpace(match[3])
		}
	}
}

// parseIrishCitation parses Irish citations
// Format: [Year] Court Number
// Example: [2023] IESC 15
func (e *Extractor) parseIrishCitation(citation *models.Citation, match []string) {
	if len(match) >= 3 {
		citation.Year = strings.TrimSpace(match[1])
		citation.Court = strings.TrimSpace(match[2])
		if len(match) >= 4 {
			citation.CaseNumber = strings.TrimSpace(match[3])
		}
	}
}

// parseAustralian parses Australian citations
// Format: [Year] Court Number
// Example: [2023] HCA 15
func (e *Extractor) parseAustralian(citation *models.Citation, match []string) {
	if len(match) >= 3 {
		citation.Year = strings.TrimSpace(match[1])
		citation.Court = strings.TrimSpace(match[2])
		if len(match) >= 4 {
			citation.CaseNumber = strings.TrimSpace(match[3])
		}
	}
}

// validateCitation validates that a citation has required fields
func (e *Extractor) validateCitation(citation *models.Citation) bool {
	// Must have raw citation
	if citation.RawCitation == "" {
		return false
	}

	// Must have at least a year or court
	if citation.Year == "" && citation.Court == "" {
		return false
	}

	return true
}

// compileCitationPatterns compiles regex patterns for different citation formats
func compileCitationPatterns() map[models.CitationFormat]*regexp.Regexp {
	patterns := make(map[models.CitationFormat]*regexp.Regexp)

	// Bluebook format: Volume Reporter Page (Court Year)
	// Example: 410 U.S. 113 (1973)
	patterns[models.CitationFormatBluebook] = regexp.MustCompile(
		`(\d+)\s+([A-Z][A-Za-z.\s]+)\s+(\d+)\s*\((?:[A-Z][A-Za-z.\s]*\s+)?(\d{4})\)`,
	)

	// Neutral citation: [Year] Court Number
	// Example: [2023] UKSC 15
	patterns[models.CitationFormatNeutral] = regexp.MustCompile(
		`\[(\d{4})\]\s+([A-Z]{2,})\s+(\d+)`,
	)

	// ECLI: ECLI:Country:Court:Year:Number
	// Example: ECLI:EU:C:2023:123
	patterns[models.CitationFormatECLI] = regexp.MustCompile(
		`ECLI:[A-Z]{2}:[A-Z]+:\d{4}:\d+`,
	)

	// Canadian neutral citation: Year CourtCode Number
	// Example: 2023 SCC 15
	patterns[models.CitationFormatCanadian] = regexp.MustCompile(
		`(\d{4})\s+(SCC|FCA|FC|ONCA|BCCA|ABCA|QCCA|SKCA|MBCA|NSCA|NBCA|PECA|NLCA|NWTCA|NUCA|YKCA)\s+(\d+)`,
	)

	// UK citation: [Year] Court Number
	// Example: [2023] UKSC 15, [2023] EWCA Civ 123
	patterns[models.CitationFormatUKCitation] = regexp.MustCompile(
		`\[(\d{4})\]\s+(UKSC|UKPC|EWCA|EWHC|EWFC|EWCOP)\s+(?:Civ|Crim|Admin|Ch|QB|Fam|Pat|Comm|TCC|Admlty)?\s*(\d+)`,
	)

	// Irish citation: [Year] Court Number
	// Example: [2023] IESC 15
	patterns[models.CitationFormatIrishCitation] = regexp.MustCompile(
		`\[(\d{4})\]\s+(IESC|IECA|IEHC|IEIC)\s+(\d+)`,
	)

	// Australian citation: [Year] Court Number
	// Example: [2023] HCA 15
	patterns[models.CitationFormatAustralian] = regexp.MustCompile(
		`\[(\d{4})\]\s+(HCA|FCAFC|FCA|NSWCA|VCA|QCA|WASCA|SASCFC|TASFC|ACTCA|NTCA)\s+(\d+)`,
	)

	return patterns
}

// extractYear is a helper function to extract a year from text
func extractYear(text string) string {
	yearPattern := regexp.MustCompile(`\b(19|20)\d{2}\b`)
	match := yearPattern.FindString(text)
	return match
}
