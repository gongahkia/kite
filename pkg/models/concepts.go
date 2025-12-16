package models

// LegalConcept represents a legal concept or doctrine
type LegalConcept struct {
	// Identifiers
	ID          string   `json:"id" validate:"required"`
	Name        string   `json:"name" validate:"required"`
	Category    string   `json:"category" validate:"required"`
	Subcategory string   `json:"subcategory,omitempty"`

	// Hierarchy
	ParentConcept string   `json:"parent_concept,omitempty"`
	ChildConcepts []string `json:"child_concepts,omitempty"`

	// Keywords
	Keywords    []string `json:"keywords" validate:"required"`
	Synonyms    []string `json:"synonyms,omitempty"`
	Aliases     []string `json:"aliases,omitempty"`

	// Metadata
	Description string   `json:"description,omitempty"`
	Jurisdiction string  `json:"jurisdiction,omitempty"` // Some concepts are jurisdiction-specific
	AreaOfLaw   string   `json:"area_of_law,omitempty"`
}

// ConceptMatch represents a matched legal concept in a case
type ConceptMatch struct {
	Concept     LegalConcept `json:"concept"`
	Confidence  float64      `json:"confidence" validate:"min=0,max=1"`
	Occurrences int          `json:"occurrences"`
	Context     []string     `json:"context,omitempty"`
}

// AreaOfLaw represents a legal area/domain
type AreaOfLaw string

const (
	AreaOfLawConstitutional AreaOfLaw = "constitutional"
	AreaOfLawCriminal       AreaOfLaw = "criminal"
	AreaOfLawCivil          AreaOfLaw = "civil"
	AreaOfLawContract       AreaOfLaw = "contract"
	AreaOfLawTort           AreaOfLaw = "tort"
	AreaOfLawProperty       AreaOfLaw = "property"
	AreaOfLawFamily         AreaOfLaw = "family"
	AreaOfLawCorporate      AreaOfLaw = "corporate"
	AreaOfLawEmployment     AreaOfLaw = "employment"
	AreaOfLawTax            AreaOfLaw = "tax"
	AreaOfLawIntellectualProperty AreaOfLaw = "intellectual_property"
	AreaOfLawEnvironmental  AreaOfLaw = "environmental"
	AreaOfLawAdministrative AreaOfLaw = "administrative"
	AreaOfLawInternational  AreaOfLaw = "international"
	AreaOfLawHumanRights    AreaOfLaw = "human_rights"
)

// Taxonomy represents the hierarchical structure of legal concepts
type Taxonomy struct {
	Version  string          `json:"version"`
	Concepts []LegalConcept  `json:"concepts"`
	Index    map[string]int  `json:"-"` // For fast lookup
}

// NewTaxonomy creates a new legal concept taxonomy
func NewTaxonomy(version string) *Taxonomy {
	return &Taxonomy{
		Version:  version,
		Concepts: make([]LegalConcept, 0),
		Index:    make(map[string]int),
	}
}

// AddConcept adds a concept to the taxonomy
func (t *Taxonomy) AddConcept(concept LegalConcept) {
	t.Index[concept.ID] = len(t.Concepts)
	t.Concepts = append(t.Concepts, concept)
}

// GetConcept retrieves a concept by ID
func (t *Taxonomy) GetConcept(id string) *LegalConcept {
	if idx, exists := t.Index[id]; exists {
		return &t.Concepts[idx]
	}
	return nil
}

// FindConceptsByKeyword finds concepts that match a keyword
func (t *Taxonomy) FindConceptsByKeyword(keyword string) []LegalConcept {
	matches := make([]LegalConcept, 0)
	for _, concept := range t.Concepts {
		for _, kw := range concept.Keywords {
			if kw == keyword {
				matches = append(matches, concept)
				break
			}
		}
	}
	return matches
}
