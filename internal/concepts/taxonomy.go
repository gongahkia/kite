package concepts

import (
	"strings"
	"sync"

	"github.com/gongahkia/kite/pkg/models"
)

// Taxonomy manages the hierarchical structure of legal concepts
type Taxonomy struct {
	concepts map[string]*models.LegalConcept
	keywords map[string][]string // keyword -> concept IDs
	mu       sync.RWMutex
}

// NewTaxonomy creates a new legal concept taxonomy
func NewTaxonomy() *Taxonomy {
	t := &Taxonomy{
		concepts: make(map[string]*models.LegalConcept),
		keywords: make(map[string][]string),
	}

	// Initialize with default concepts
	t.initializeDefaultConcepts()

	return t
}

// GetConcept retrieves a concept by ID
func (t *Taxonomy) GetConcept(id string) (*models.LegalConcept, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	concept, exists := t.concepts[id]
	return concept, exists
}

// GetConceptsByArea retrieves all concepts for a specific area of law
func (t *Taxonomy) GetConceptsByArea(area models.AreaOfLaw) []*models.LegalConcept {
	t.mu.RLock()
	defer t.mu.RUnlock()

	concepts := make([]*models.LegalConcept, 0)
	for _, concept := range t.concepts {
		if concept.Area == area {
			concepts = append(concepts, concept)
		}
	}

	return concepts
}

// SearchByKeyword finds concepts matching a keyword
func (t *Taxonomy) SearchByKeyword(keyword string) []*models.LegalConcept {
	t.mu.RLock()
	defer t.mu.RUnlock()

	keyword = strings.ToLower(keyword)
	conceptIDs, exists := t.keywords[keyword]
	if !exists {
		return nil
	}

	concepts := make([]*models.LegalConcept, 0, len(conceptIDs))
	for _, id := range conceptIDs {
		if concept, ok := t.concepts[id]; ok {
			concepts = append(concepts, concept)
		}
	}

	return concepts
}

// AddConcept adds a new concept to the taxonomy
func (t *Taxonomy) AddConcept(concept *models.LegalConcept) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.concepts[concept.ID] = concept

	// Index keywords
	for _, keyword := range concept.Keywords {
		keyword = strings.ToLower(keyword)
		t.keywords[keyword] = append(t.keywords[keyword], concept.ID)
	}
}

// GetAllConcepts returns all concepts in the taxonomy
func (t *Taxonomy) GetAllConcepts() []*models.LegalConcept {
	t.mu.RLock()
	defer t.mu.RUnlock()

	concepts := make([]*models.LegalConcept, 0, len(t.concepts))
	for _, concept := range t.concepts {
		concepts = append(concepts, concept)
	}

	return concepts
}

// GetStats returns statistics about the taxonomy
func (t *Taxonomy) GetStats() map[string]int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	stats := make(map[string]int)
	stats["total_concepts"] = len(t.concepts)
	stats["total_keywords"] = len(t.keywords)

	// Count by area
	areaCount := make(map[models.AreaOfLaw]int)
	for _, concept := range t.concepts {
		areaCount[concept.Area]++
	}

	for area, count := range areaCount {
		stats[string(area)] = count
	}

	return stats
}

// initializeDefaultConcepts initializes the taxonomy with default legal concepts
func (t *Taxonomy) initializeDefaultConcepts() {
	// This is a curated set of common legal concepts across jurisdictions
	// In production, this could be loaded from a YAML file

	defaultConcepts := []*models.LegalConcept{
		// Constitutional Law
		{
			ID:          "const-01",
			Name:        "Freedom of Speech",
			Description: "The right to express opinions without government restraint",
			Area:        models.AreaConstitutional,
			Keywords:    []string{"freedom of speech", "first amendment", "expression", "free speech", "censorship"},
			Importance:  9,
		},
		{
			ID:          "const-02",
			Name:        "Due Process",
			Description: "Fair treatment through the normal judicial system",
			Area:        models.AreaConstitutional,
			Keywords:    []string{"due process", "procedural fairness", "natural justice", "fair hearing"},
			Importance:  10,
		},
		{
			ID:          "const-03",
			Name:        "Equal Protection",
			Description: "Equal treatment under the law",
			Area:        models.AreaConstitutional,
			Keywords:    []string{"equal protection", "discrimination", "equality", "disparate treatment"},
			Importance:  9,
		},

		// Criminal Law
		{
			ID:          "crim-01",
			Name:        "Mens Rea",
			Description: "Criminal intent or knowledge of wrongdoing",
			Area:        models.AreaCriminal,
			Keywords:    []string{"mens rea", "criminal intent", "guilty mind", "intention", "recklessness"},
			Importance:  10,
		},
		{
			ID:          "crim-02",
			Name:        "Actus Reus",
			Description: "The physical act of committing a crime",
			Area:        models.AreaCriminal,
			Keywords:    []string{"actus reus", "guilty act", "criminal act", "physical element"},
			Importance:  10,
		},
		{
			ID:          "crim-03",
			Name:        "Self-Defense",
			Description: "Legal justification for using force to protect oneself",
			Area:        models.AreaCriminal,
			Keywords:    []string{"self-defense", "self-defence", "defense of person", "justification"},
			Importance:  8,
		},
		{
			ID:          "crim-04",
			Name:        "Reasonable Doubt",
			Description: "Standard of proof in criminal cases",
			Area:        models.AreaCriminal,
			Keywords:    []string{"reasonable doubt", "beyond reasonable doubt", "burden of proof", "standard of proof"},
			Importance:  10,
		},

		// Contract Law
		{
			ID:          "cont-01",
			Name:        "Offer and Acceptance",
			Description: "Essential elements for contract formation",
			Area:        models.AreaContract,
			Keywords:    []string{"offer", "acceptance", "agreement", "meeting of minds", "consensus ad idem"},
			Importance:  10,
		},
		{
			ID:          "cont-02",
			Name:        "Consideration",
			Description: "Something of value exchanged between parties",
			Area:        models.AreaContract,
			Keywords:    []string{"consideration", "quid pro quo", "valuable consideration", "bargain"},
			Importance:  10,
		},
		{
			ID:          "cont-03",
			Name:        "Breach of Contract",
			Description: "Failure to perform contractual obligations",
			Area:        models.AreaContract,
			Keywords:    []string{"breach", "breach of contract", "material breach", "fundamental breach", "repudiation"},
			Importance:  9,
		},
		{
			ID:          "cont-04",
			Name:        "Damages",
			Description: "Monetary compensation for breach",
			Area:        models.AreaContract,
			Keywords:    []string{"damages", "compensation", "expectation damages", "reliance damages", "restitution"},
			Importance:  8,
		},

		// Tort Law
		{
			ID:          "tort-01",
			Name:        "Negligence",
			Description: "Failure to exercise reasonable care",
			Area:        models.AreaTort,
			Keywords:    []string{"negligence", "duty of care", "breach of duty", "reasonable person", "standard of care"},
			Importance:  10,
		},
		{
			ID:          "tort-02",
			Name:        "Causation",
			Description: "Link between conduct and harm",
			Area:        models.AreaTort,
			Keywords:    []string{"causation", "proximate cause", "but-for test", "cause in fact", "foreseeability"},
			Importance:  9,
		},
		{
			ID:          "tort-03",
			Name:        "Defamation",
			Description: "False statement harming reputation",
			Area:        models.AreaTort,
			Keywords:    []string{"defamation", "libel", "slander", "reputation", "false statement"},
			Importance:  7,
		},

		// Property Law
		{
			ID:          "prop-01",
			Name:        "Adverse Possession",
			Description: "Acquiring ownership through continuous possession",
			Area:        models.AreaProperty,
			Keywords:    []string{"adverse possession", "squatter's rights", "prescription", "possessory title"},
			Importance:  7,
		},
		{
			ID:          "prop-02",
			Name:        "Easement",
			Description: "Right to use another's property for specific purpose",
			Area:        models.AreaProperty,
			Keywords:    []string{"easement", "right of way", "servitude", "profit à prendre"},
			Importance:  6,
		},

		// Family Law
		{
			ID:          "fam-01",
			Name:        "Child Custody",
			Description: "Legal guardianship of a child",
			Area:        models.AreaFamily,
			Keywords:    []string{"custody", "child custody", "parental rights", "best interests of child"},
			Importance:  8,
		},
		{
			ID:          "fam-02",
			Name:        "Divorce",
			Description: "Legal dissolution of marriage",
			Area:        models.AreaFamily,
			Keywords:    []string{"divorce", "dissolution", "marriage breakdown", "separation"},
			Importance:  7,
		},

		// Administrative Law
		{
			ID:          "admin-01",
			Name:        "Judicial Review",
			Description: "Court review of administrative decisions",
			Area:        models.AreaAdministrative,
			Keywords:    []string{"judicial review", "administrative review", "ultra vires", "unreasonableness"},
			Importance:  8,
		},
		{
			ID:          "admin-02",
			Name:        "Procedural Fairness",
			Description: "Fair process in administrative decisions",
			Area:        models.AreaAdministrative,
			Keywords:    []string{"procedural fairness", "natural justice", "right to be heard", "bias"},
			Importance:  9,
		},

		// Labor/Employment Law
		{
			ID:          "labor-01",
			Name:        "Wrongful Dismissal",
			Description: "Termination without just cause",
			Area:        models.AreaLabor,
			Keywords:    []string{"wrongful dismissal", "unfair dismissal", "termination", "just cause"},
			Importance:  7,
		},
		{
			ID:          "labor-02",
			Name:        "Collective Bargaining",
			Description: "Negotiation between employer and union",
			Area:        models.AreaLabor,
			Keywords:    []string{"collective bargaining", "union", "labor agreement", "collective agreement"},
			Importance:  6,
		},

		// Evidence Law
		{
			ID:          "evid-01",
			Name:        "Hearsay",
			Description: "Out-of-court statement offered for truth",
			Area:        models.AreaEvidence,
			Keywords:    []string{"hearsay", "hearsay rule", "out of court statement", "second-hand evidence"},
			Importance:  8,
		},
		{
			ID:          "evid-02",
			Name:        "Privilege",
			Description: "Protection from disclosure of confidential communications",
			Area:        models.AreaEvidence,
			Keywords:    []string{"privilege", "attorney-client privilege", "solicitor-client privilege", "confidential"},
			Importance:  7,
		},

		// Intellectual Property
		{
			ID:          "ip-01",
			Name:        "Copyright Infringement",
			Description: "Unauthorized use of copyrighted work",
			Area:        models.AreaIntellectualProperty,
			Keywords:    []string{"copyright", "infringement", "fair use", "fair dealing", "reproduction"},
			Importance:  7,
		},
		{
			ID:          "ip-02",
			Name:        "Patent",
			Description: "Exclusive right to invention",
			Area:        models.AreaIntellectualProperty,
			Keywords:    []string{"patent", "invention", "novelty", "non-obviousness", "utility"},
			Importance:  6,
		},

		// Tax Law
		{
			ID:          "tax-01",
			Name:        "Tax Evasion",
			Description: "Illegal non-payment of taxes",
			Area:        models.AreaTax,
			Keywords:    []string{"tax evasion", "tax fraud", "tax avoidance", "evasion"},
			Importance:  7,
		},

		// Environmental Law
		{
			ID:          "env-01",
			Name:        "Environmental Impact Assessment",
			Description: "Evaluation of environmental effects",
			Area:        models.AreaEnvironmental,
			Keywords:    []string{"environmental impact", "assessment", "environmental assessment", "EIA"},
			Importance:  6,
		},

		// Human Rights
		{
			ID:          "hr-01",
			Name:        "Discrimination",
			Description: "Unfair treatment based on protected characteristics",
			Area:        models.AreaHumanRights,
			Keywords:    []string{"discrimination", "protected grounds", "equality", "human rights violation"},
			Importance:  9,
		},
		{
			ID:          "hr-02",
			Name:        "Freedom from Torture",
			Description: "Prohibition of cruel and unusual punishment",
			Area:        models.AreaHumanRights,
			Keywords:    []string{"torture", "cruel and unusual", "inhuman treatment", "degrading treatment"},
			Importance:  10,
		},

		// Corporate Law
		{
			ID:          "corp-01",
			Name:        "Fiduciary Duty",
			Description: "Obligation to act in best interest of another",
			Area:        models.AreaCorporate,
			Keywords:    []string{"fiduciary duty", "duty of loyalty", "duty of care", "fiduciary"},
			Importance:  8,
		},
		{
			ID:          "corp-02",
			Name:        "Piercing the Corporate Veil",
			Description: "Holding shareholders personally liable",
			Area:        models.AreaCorporate,
			Keywords:    []string{"piercing the veil", "corporate veil", "alter ego", "personal liability"},
			Importance:  7,
		},
	}

	// Add all default concepts
	for _, concept := range defaultConcepts {
		t.AddConcept(concept)
	}
}
