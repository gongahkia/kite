package models

import "time"

// CitationFormat represents the citation format type
type CitationFormat string

const (
	CitationFormatBluebook      CitationFormat = "bluebook"
	CitationFormatNeutral       CitationFormat = "neutral"
	CitationFormatJurisdiction  CitationFormat = "jurisdiction_specific"
	CitationFormatECLI          CitationFormat = "ecli"
	CitationFormatOther         CitationFormat = "other"
)

// Citation represents a legal citation
type Citation struct {
	// Core Citation Information
	RawCitation     string          `json:"raw_citation" validate:"required"`
	NormalizedCitation string       `json:"normalized_citation,omitempty"`
	Format          CitationFormat  `json:"format" validate:"required"`

	// Case Information
	CaseID          string          `json:"case_id,omitempty"`
	CaseName        string          `json:"case_name,omitempty"`
	CaseYear        int             `json:"case_year,omitempty" validate:"omitempty,min=1600,max=2100"`

	// Court Information
	Court           string          `json:"court,omitempty"`
	Reporter        string          `json:"reporter,omitempty"`
	Volume          string          `json:"volume,omitempty"`
	Page            string          `json:"page,omitempty"`

	// Context
	CitingCaseID    string          `json:"citing_case_id,omitempty"`
	Context         string          `json:"context,omitempty"`
	Paragraph       string          `json:"paragraph,omitempty"`
	TreatmentType   string          `json:"treatment_type,omitempty"` // followed, distinguished, overruled, etc.

	// Validation
	IsValid         bool            `json:"is_valid"`
	ValidationError string          `json:"validation_error,omitempty"`

	// Metadata
	ExtractedAt     time.Time       `json:"extracted_at" validate:"required"`
	Confidence      float64         `json:"confidence" validate:"min=0,max=1"`
}

// CitationNetwork represents a network of case citations
type CitationNetwork struct {
	Nodes []CitationNode `json:"nodes"`
	Edges []CitationEdge `json:"edges"`
}

// CitationNode represents a case in the citation network
type CitationNode struct {
	CaseID           string  `json:"case_id" validate:"required"`
	CaseName         string  `json:"case_name" validate:"required"`
	AuthorityScore   float64 `json:"authority_score"`   // PageRank-style score
	HubScore         float64 `json:"hub_score"`         // How often this case cites others
	InDegree         int     `json:"in_degree"`         // Number of incoming citations
	OutDegree        int     `json:"out_degree"`        // Number of outgoing citations
}

// CitationEdge represents a citation relationship
type CitationEdge struct {
	FromCaseID    string  `json:"from_case_id" validate:"required"`
	ToCaseID      string  `json:"to_case_id" validate:"required"`
	TreatmentType string  `json:"treatment_type,omitempty"`
	Weight        float64 `json:"weight"`
}

// NewCitation creates a new Citation
func NewCitation(rawCitation string, format CitationFormat) *Citation {
	return &Citation{
		RawCitation: rawCitation,
		Format:      format,
		IsValid:     false,
		ExtractedAt: time.Now(),
		Confidence:  0.0,
	}
}

// Normalize normalizes the citation to a standard format
func (c *Citation) Normalize() string {
	// TODO: Implement citation normalization logic
	if c.NormalizedCitation == "" {
		c.NormalizedCitation = c.RawCitation
	}
	return c.NormalizedCitation
}

// Validate validates the citation format and content
func (c *Citation) Validate() error {
	// TODO: Implement citation validation logic
	c.IsValid = true
	return nil
}

// SetConfidence sets the confidence score for the citation
func (c *Citation) SetConfidence(score float64) {
	if score < 0 {
		c.Confidence = 0
	} else if score > 1 {
		c.Confidence = 1
	} else {
		c.Confidence = score
	}
}
