package models

import (
	"time"
)

// CaseStatus represents the status of a case
type CaseStatus string

const (
	CaseStatusPending    CaseStatus = "pending"
	CaseStatusActive     CaseStatus = "active"
	CaseStatusClosed     CaseStatus = "closed"
	CaseStatusAppealed   CaseStatus = "appealed"
	CaseStatusOverturned CaseStatus = "overturned"
)

// CourtLevel represents the hierarchical level of a court
type CourtLevel int

const (
	CourtLevelSupreme CourtLevel = iota + 1 // 1 - Supreme/Constitutional Court
	CourtLevelAppellate                      // 2 - Court of Appeal
	CourtLevelHigher                         // 3 - High/Regional Court
	CourtLevelDistrict                       // 4 - District/County Court
	CourtLevelLocal                          // 5 - Local/Magistrate Court
)

// CourtType represents the type of court
type CourtType string

const (
	CourtTypeCivil        CourtType = "civil"
	CourtTypeCriminal     CourtType = "criminal"
	CourtTypeConstitutional CourtType = "constitutional"
	CourtTypeAdministrative CourtType = "administrative"
	CourtTypeFamily       CourtType = "family"
	CourtTypeCommercial   CourtType = "commercial"
	CourtTypeMixed        CourtType = "mixed"
)

// Case represents a legal case with all associated metadata
type Case struct {
	// Identifiers
	ID           string    `json:"id" validate:"required"`
	CaseNumber   string    `json:"case_number" validate:"required"`
	CaseName     string    `json:"case_name" validate:"required"`
	AlternateNames []string `json:"alternate_names,omitempty"`

	// Temporal Information
	FilingDate   *time.Time `json:"filing_date,omitempty"`
	DecisionDate *time.Time `json:"decision_date,omitempty" validate:"required"`
	HearingDate  *time.Time `json:"hearing_date,omitempty"`

	// Court Information
	Court           string      `json:"court" validate:"required"`
	CourtLevel      CourtLevel  `json:"court_level" validate:"required,min=1,max=5"`
	CourtType       CourtType   `json:"court_type" validate:"required"`
	Jurisdiction    string      `json:"jurisdiction" validate:"required"`

	// Parties
	Parties     []Party   `json:"parties,omitempty"`
	Appellant   string    `json:"appellant,omitempty"`
	Respondent  string    `json:"respondent,omitempty"`

	// Judges
	Judges      []string  `json:"judges,omitempty"`
	ChiefJudge  string    `json:"chief_judge,omitempty"`

	// Content
	Summary     string    `json:"summary,omitempty"`
	Headnotes   string    `json:"headnotes,omitempty"`
	FullText    string    `json:"full_text,omitempty"`
	Language    string    `json:"language" validate:"required"`

	// Citations
	Citations       []Citation  `json:"citations,omitempty"`
	CitedBy         []string    `json:"cited_by,omitempty"`
	Precedent       []string    `json:"precedent,omitempty"`

	// Legal Concepts
	LegalConcepts   []string    `json:"legal_concepts,omitempty"`
	AreasOfLaw      []string    `json:"areas_of_law,omitempty"`
	Keywords        []string    `json:"keywords,omitempty"`

	// Case Outcome
	Status          CaseStatus  `json:"status" validate:"required"`
	Outcome         string      `json:"outcome,omitempty"`
	Disposition     string      `json:"disposition,omitempty"`

	// Source Information
	URL             string      `json:"url" validate:"required,url"`
	SourceDatabase  string      `json:"source_database" validate:"required"`
	ScrapedAt       time.Time   `json:"scraped_at" validate:"required"`
	LastUpdated     time.Time   `json:"last_updated" validate:"required"`

	// Metadata
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
	QualityScore    float64     `json:"quality_score" validate:"min=0,max=1"`

	// Document Information
	DocumentType    string      `json:"document_type,omitempty"`
	ECLI            string      `json:"ecli,omitempty"` // European Case Law Identifier
	Docket          string      `json:"docket,omitempty"`
}

// Party represents a party involved in a legal case
type Party struct {
	Name     string   `json:"name" validate:"required"`
	Role     string   `json:"role" validate:"required"` // plaintiff, defendant, appellant, respondent, etc.
	Type     string   `json:"type,omitempty"`           // individual, corporation, government, etc.
	Lawyers  []string `json:"lawyers,omitempty"`
}

// NewCase creates a new Case with defaults
func NewCase() *Case {
	now := time.Now()
	return &Case{
		Status:       CaseStatusPending,
		Language:     "en",
		ScrapedAt:    now,
		LastUpdated:  now,
		QualityScore: 0.0,
		Metadata:     make(map[string]interface{}),
	}
}

// IsComplete checks if the case has all required fields
func (c *Case) IsComplete() bool {
	return c.ID != "" &&
		c.CaseNumber != "" &&
		c.CaseName != "" &&
		c.DecisionDate != nil &&
		c.Court != "" &&
		c.Jurisdiction != "" &&
		c.URL != ""
}

// AddCitation adds a citation to the case
func (c *Case) AddCitation(citation Citation) {
	c.Citations = append(c.Citations, citation)
}

// AddLegalConcept adds a legal concept to the case
func (c *Case) AddLegalConcept(concept string) {
	c.LegalConcepts = append(c.LegalConcepts, concept)
}

// SetQualityScore sets the quality score based on completeness and validation
func (c *Case) SetQualityScore(score float64) {
	if score < 0 {
		c.QualityScore = 0
	} else if score > 1 {
		c.QualityScore = 1
	} else {
		c.QualityScore = score
	}
}
