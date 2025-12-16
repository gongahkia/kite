package citation

import (
	"fmt"
	"strings"

	"github.com/gongahkia/kite/pkg/models"
)

// Normalizer normalizes citations to a standard format
type Normalizer struct {
	reporterAbbreviations map[string]string
	courtAbbreviations    map[string]string
}

// NewNormalizer creates a new citation normalizer
func NewNormalizer() *Normalizer {
	return &Normalizer{
		reporterAbbreviations: buildReporterAbbreviations(),
		courtAbbreviations:    buildCourtAbbreviations(),
	}
}

// Normalize normalizes a citation to a standard format
func (n *Normalizer) Normalize(citation *models.Citation) *models.Citation {
	if citation == nil {
		return nil
	}

	// Create a copy to avoid modifying the original
	normalized := &models.Citation{
		Format:       citation.Format,
		RawCitation:  citation.RawCitation,
		Volume:       citation.Volume,
		Reporter:     citation.Reporter,
		Page:         citation.Page,
		Year:         citation.Year,
		Court:        citation.Court,
		CaseNumber:   citation.CaseNumber,
		Country:      citation.Country,
		CitingCaseID: citation.CitingCaseID,
		CitedCaseID:  citation.CitedCaseID,
	}

	// Normalize components
	normalized.Reporter = n.normalizeReporter(citation.Reporter)
	normalized.Court = n.normalizeCourt(citation.Court)
	normalized.Year = n.normalizeYear(citation.Year)
	normalized.CaseNumber = n.normalizeCaseNumber(citation.CaseNumber)

	// Generate normalized citation string
	normalized.NormalizedCitation = n.generateNormalizedCitation(normalized)
	normalized.IsNormalized = true

	return normalized
}

// NormalizeBatch normalizes a batch of citations
func (n *Normalizer) NormalizeBatch(citations []*models.Citation) []*models.Citation {
	normalized := make([]*models.Citation, len(citations))
	for i, citation := range citations {
		normalized[i] = n.Normalize(citation)
	}
	return normalized
}

// normalizeReporter normalizes reporter abbreviations
func (n *Normalizer) normalizeReporter(reporter string) string {
	if reporter == "" {
		return ""
	}

	// Trim whitespace and convert to lowercase for lookup
	key := strings.TrimSpace(strings.ToLower(reporter))

	// Check if we have a standard abbreviation
	if standard, exists := n.reporterAbbreviations[key]; exists {
		return standard
	}

	// Remove extra whitespace
	reporter = strings.Join(strings.Fields(reporter), " ")

	return reporter
}

// normalizeCourt normalizes court abbreviations
func (n *Normalizer) normalizeCourt(court string) string {
	if court == "" {
		return ""
	}

	// Trim whitespace and convert to uppercase
	court = strings.ToUpper(strings.TrimSpace(court))

	// Check if we have a standard abbreviation
	if standard, exists := n.courtAbbreviations[court]; exists {
		return standard
	}

	return court
}

// normalizeYear normalizes year format
func (n *Normalizer) normalizeYear(year string) string {
	if year == "" {
		return ""
	}

	// Extract just the 4-digit year
	year = strings.TrimSpace(year)

	// Remove any non-numeric characters
	var digits strings.Builder
	for _, ch := range year {
		if ch >= '0' && ch <= '9' {
			digits.WriteRune(ch)
		}
	}

	yearStr := digits.String()

	// Return first 4 digits if we have them
	if len(yearStr) >= 4 {
		return yearStr[:4]
	}

	return yearStr
}

// normalizeCaseNumber normalizes case numbers
func (n *Normalizer) normalizeCaseNumber(caseNumber string) string {
	if caseNumber == "" {
		return ""
	}

	// Trim whitespace and remove leading zeros
	caseNumber = strings.TrimSpace(caseNumber)
	caseNumber = strings.TrimLeft(caseNumber, "0")

	// If all zeros, return "0"
	if caseNumber == "" {
		return "0"
	}

	return caseNumber
}

// generateNormalizedCitation generates a normalized citation string
func (n *Normalizer) generateNormalizedCitation(citation *models.Citation) string {
	switch citation.Format {
	case models.CitationFormatBluebook:
		return n.generateBluebookCitation(citation)
	case models.CitationFormatNeutral:
		return n.generateNeutralCitation(citation)
	case models.CitationFormatECLI:
		return n.generateECLICitation(citation)
	case models.CitationFormatCanadian:
		return n.generateCanadianCitation(citation)
	case models.CitationFormatUKCitation:
		return n.generateUKCitation(citation)
	case models.CitationFormatIrishCitation:
		return n.generateIrishCitation(citation)
	case models.CitationFormatAustralian:
		return n.generateAustralianCitation(citation)
	default:
		return citation.RawCitation
	}
}

// generateBluebookCitation generates a Bluebook-style citation
func (n *Normalizer) generateBluebookCitation(citation *models.Citation) string {
	if citation.Volume == "" || citation.Reporter == "" || citation.Page == "" {
		return citation.RawCitation
	}

	if citation.Year != "" {
		return fmt.Sprintf("%s %s %s (%s)",
			citation.Volume, citation.Reporter, citation.Page, citation.Year)
	}

	return fmt.Sprintf("%s %s %s",
		citation.Volume, citation.Reporter, citation.Page)
}

// generateNeutralCitation generates a neutral citation
func (n *Normalizer) generateNeutralCitation(citation *models.Citation) string {
	if citation.Year == "" || citation.Court == "" || citation.CaseNumber == "" {
		return citation.RawCitation
	}

	return fmt.Sprintf("[%s] %s %s", citation.Year, citation.Court, citation.CaseNumber)
}

// generateECLICitation generates an ECLI citation
func (n *Normalizer) generateECLICitation(citation *models.Citation) string {
	if citation.Country == "" || citation.Court == "" || citation.Year == "" || citation.CaseNumber == "" {
		return citation.RawCitation
	}

	return fmt.Sprintf("ECLI:%s:%s:%s:%s",
		citation.Country, citation.Court, citation.Year, citation.CaseNumber)
}

// generateCanadianCitation generates a Canadian neutral citation
func (n *Normalizer) generateCanadianCitation(citation *models.Citation) string {
	if citation.Year == "" || citation.Court == "" || citation.CaseNumber == "" {
		return citation.RawCitation
	}

	return fmt.Sprintf("%s %s %s", citation.Year, citation.Court, citation.CaseNumber)
}

// generateUKCitation generates a UK citation
func (n *Normalizer) generateUKCitation(citation *models.Citation) string {
	if citation.Year == "" || citation.Court == "" || citation.CaseNumber == "" {
		return citation.RawCitation
	}

	return fmt.Sprintf("[%s] %s %s", citation.Year, citation.Court, citation.CaseNumber)
}

// generateIrishCitation generates an Irish citation
func (n *Normalizer) generateIrishCitation(citation *models.Citation) string {
	if citation.Year == "" || citation.Court == "" || citation.CaseNumber == "" {
		return citation.RawCitation
	}

	return fmt.Sprintf("[%s] %s %s", citation.Year, citation.Court, citation.CaseNumber)
}

// generateAustralianCitation generates an Australian citation
func (n *Normalizer) generateAustralianCitation(citation *models.Citation) string {
	if citation.Year == "" || citation.Court == "" || citation.CaseNumber == "" {
		return citation.RawCitation
	}

	return fmt.Sprintf("[%s] %s %s", citation.Year, citation.Court, citation.CaseNumber)
}

// buildReporterAbbreviations builds a map of reporter abbreviations
func buildReporterAbbreviations() map[string]string {
	return map[string]string{
		// US Reporters
		"u.s.":    "U.S.",
		"us":      "U.S.",
		"s.ct.":   "S. Ct.",
		"s. ct.":  "S. Ct.",
		"sct":     "S. Ct.",
		"f.2d":    "F.2d",
		"f2d":     "F.2d",
		"f.3d":    "F.3d",
		"f3d":     "F.3d",
		"f.supp":  "F. Supp.",
		"fsupp":   "F. Supp.",
		"f.supp.2d": "F. Supp. 2d",
		"fsupp2d": "F. Supp. 2d",

		// UK Reporters
		"ac":      "A.C.",
		"a.c.":    "A.C.",
		"ch":      "Ch.",
		"ch.":     "Ch.",
		"qb":      "Q.B.",
		"q.b.":    "Q.B.",
		"kb":      "K.B.",
		"k.b.":    "K.B.",
		"wlr":     "W.L.R.",
		"w.l.r.":  "W.L.R.",

		// Canadian Reporters
		"scr":     "S.C.R.",
		"s.c.r.":  "S.C.R.",
		"fc":      "F.C.",
		"f.c.":    "F.C.",
		"drs":     "D.R.S.",
		"d.r.s.":  "D.R.S.",
	}
}

// buildCourtAbbreviations builds a map of court abbreviations
func buildCourtAbbreviations() map[string]string {
	return map[string]string{
		// US Courts
		"SCOTUS":        "SCOTUS",
		"USSC":          "SCOTUS",
		"US SUPREME COURT": "SCOTUS",

		// UK Courts
		"UKSC":          "UKSC",
		"UK SUPREME COURT": "UKSC",
		"UKPC":          "UKPC",
		"EWCA":          "EWCA",
		"EWHC":          "EWHC",
		"EWFC":          "EWFC",

		// Canadian Courts
		"SCC":           "SCC",
		"SUPREME COURT OF CANADA": "SCC",
		"FCA":           "FCA",
		"FC":            "FC",

		// Irish Courts
		"IESC":          "IESC",
		"IECA":          "IECA",
		"IEHC":          "IEHC",

		// Australian Courts
		"HCA":           "HCA",
		"FCAFC":         "FCAFC",
		"FCA":           "FCA",
	}
}
