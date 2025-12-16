package jurisdiction

import (
	"regexp"
	"strings"
	"time"

	"github.com/gongahkia/kite/pkg/models"
)

// MetadataEnricher handles jurisdiction-specific metadata enrichment
type MetadataEnricher struct {
	hierarchy *CourtHierarchy
	rules     *JurisdictionRules
}

// NewMetadataEnricher creates a new metadata enricher
func NewMetadataEnricher() *MetadataEnricher {
	return &MetadataEnricher{
		hierarchy: NewCourtHierarchy(),
		rules:     NewJurisdictionRules(),
	}
}

// EnrichCase enriches a case with jurisdiction-specific metadata
func (me *MetadataEnricher) EnrichCase(c *models.Case) error {
	// Determine court level
	if c.Court != "" {
		c.CourtLevel = me.hierarchy.GetCourtLevel(c.Court)
	}

	// Determine court type
	courtType := me.hierarchy.GetCourtType(c.Court)
	c.Metadata = c.Metadata // Ensure metadata map is initialized

	// Extract case type
	caseType := me.DetermineCaseType(c)
	if c.Metadata == nil {
		c.Metadata = make(map[string]interface{})
	}
	c.Metadata["case_type"] = caseType
	c.Metadata["court_type"] = string(courtType)

	// Check if precedential
	isPrecedential := me.hierarchy.IsPrecedential(c.Court)
	c.Metadata["precedential"] = isPrecedential

	// Extract procedural metadata
	procedural := me.ExtractProceduralMetadata(c)
	for k, v := range procedural {
		c.Metadata[k] = v
	}

	// Apply jurisdiction-specific rules
	me.ApplyJurisdictionRules(c)

	return nil
}

// CaseType represents the type of legal case
type CaseType string

const (
	CaseTypeCivil       CaseType = "civil"
	CaseTypeCriminal    CaseType = "criminal"
	CaseTypeConstitutional CaseType = "constitutional"
	CaseTypeAdministrative CaseType = "administrative"
	CaseTypeTax         CaseType = "tax"
	CaseTypeFamily      CaseType = "family"
	CaseTypeBankruptcy  CaseType = "bankruptcy"
	CaseTypeEmployment  CaseType = "employment"
	CaseTypeIntellectualProperty CaseType = "intellectual_property"
	CaseTypeEnvironmental CaseType = "environmental"
	CaseTypeCommercial  CaseType = "commercial"
	CaseTypeTort        CaseType = "tort"
	CaseTypeContract    CaseType = "contract"
	CaseTypeProperty    CaseType = "property"
	CaseTypeOther       CaseType = "other"
)

// DetermineCaseType classifies the type of case based on various indicators
func (me *MetadataEnricher) DetermineCaseType(c *models.Case) CaseType {
	// Combine text sources for analysis
	textToAnalyze := strings.ToLower(c.CaseName + " " + c.Summary + " " + c.FullText)

	// Criminal case indicators
	criminalPatterns := []string{
		"criminal", "prosecution", "defendant", "guilty", "plea", "sentence",
		"indictment", "conviction", "acquittal", "murder", "assault", "robbery",
		"theft", "fraud", "drug", "narcotics", "felony", "misdemeanor",
		"bail", "parole", "probation", "r v ", " v regina", " v the queen",
	}
	if containsAny(textToAnalyze, criminalPatterns) {
		return CaseTypeCriminal
	}

	// Constitutional case indicators
	constitutionalPatterns := []string{
		"constitutional", "bill of rights", "charter", "fundamental rights",
		"separation of powers", "judicial review", "unconstitutional",
		"constitution", "amendment",
	}
	if containsAny(textToAnalyze, constitutionalPatterns) {
		return CaseTypeConstitutional
	}

	// Administrative case indicators
	administrativePatterns := []string{
		"administrative", "regulatory", "agency", "commission",
		"tribunal", "board of", "minister", "secretary of state",
		"immigration", "asylum", "license", "permit",
	}
	if containsAny(textToAnalyze, administrativePatterns) {
		return CaseTypeAdministrative
	}

	// Tax case indicators
	taxPatterns := []string{
		"tax", "revenue", "inland revenue", "irs", "vat", "duty",
		"customs", "excise", "taxation",
	}
	if containsAny(textToAnalyze, taxPatterns) {
		return CaseTypeTax
	}

	// Family case indicators
	familyPatterns := []string{
		"divorce", "custody", "child support", "alimony", "adoption",
		"domestic violence", "family law", "matrimonial", "spousal",
		"parental rights", "guardianship",
	}
	if containsAny(textToAnalyze, familyPatterns) {
		return CaseTypeFamily
	}

	// Bankruptcy case indicators
	bankruptcyPatterns := []string{
		"bankruptcy", "insolvency", "liquidation", "receivership",
		"creditor", "debtor", "chapter 7", "chapter 11",
	}
	if containsAny(textToAnalyze, bankruptcyPatterns) {
		return CaseTypeBankruptcy
	}

	// Employment case indicators
	employmentPatterns := []string{
		"employment", "unfair dismissal", "discrimination", "wrongful termination",
		"labor", "workplace", "employee", "employer", "wages", "hours",
		"harassment", "equal pay",
	}
	if containsAny(textToAnalyze, employmentPatterns) {
		return CaseTypeEmployment
	}

	// Intellectual Property case indicators
	ipPatterns := []string{
		"patent", "trademark", "copyright", "trade secret",
		"intellectual property", "infringement", "licensing",
	}
	if containsAny(textToAnalyze, ipPatterns) {
		return CaseTypeIntellectualProperty
	}

	// Environmental case indicators
	environmentalPatterns := []string{
		"environmental", "pollution", "clean air", "clean water",
		"endangered species", "environmental impact", "epa",
	}
	if containsAny(textToAnalyze, environmentalPatterns) {
		return CaseTypeEnvironmental
	}

	// Commercial case indicators
	commercialPatterns := []string{
		"commercial", "business", "corporate", "shareholder",
		"merger", "acquisition", "securities", "stock",
	}
	if containsAny(textToAnalyze, commercialPatterns) {
		return CaseTypeCommercial
	}

	// Tort case indicators
	tortPatterns := []string{
		"negligence", "damages", "tort", "personal injury",
		"malpractice", "defamation", "nuisance", "trespass",
	}
	if containsAny(textToAnalyze, tortPatterns) {
		return CaseTypeTort
	}

	// Contract case indicators
	contractPatterns := []string{
		"contract", "breach", "agreement", "warranty",
		"consideration", "offer", "acceptance",
	}
	if containsAny(textToAnalyze, contractPatterns) {
		return CaseTypeContract
	}

	// Property case indicators
	propertyPatterns := []string{
		"property", "real estate", "landlord", "tenant",
		"lease", "easement", "foreclosure", "title",
	}
	if containsAny(textToAnalyze, propertyPatterns) {
		return CaseTypeProperty
	}

	// Default to civil if no specific type identified
	return CaseTypeCivil
}

// ExtractProceduralMetadata extracts procedural information from the case
func (me *MetadataEnricher) ExtractProceduralMetadata(c *models.Case) map[string]interface{} {
	metadata := make(map[string]interface{})

	text := strings.ToLower(c.Summary + " " + c.FullText)

	// Determine if it's an appeal
	appealPatterns := []string{"appeal", "appellant", "appellee", "petition for review"}
	metadata["is_appeal"] = containsAny(text, appealPatterns)

	// Determine procedural stage
	if strings.Contains(text, "motion to dismiss") || strings.Contains(text, "dismissal") {
		metadata["procedural_stage"] = "motion_to_dismiss"
	} else if strings.Contains(text, "summary judgment") {
		metadata["procedural_stage"] = "summary_judgment"
	} else if strings.Contains(text, "trial") {
		metadata["procedural_stage"] = "trial"
	} else if strings.Contains(text, "sentencing") {
		metadata["procedural_stage"] = "sentencing"
	} else if containsAny(text, appealPatterns) {
		metadata["procedural_stage"] = "appeal"
	}

	// Extract decision type
	if strings.Contains(text, "affirmed") {
		metadata["decision"] = "affirmed"
	} else if strings.Contains(text, "reversed") {
		metadata["decision"] = "reversed"
	} else if strings.Contains(text, "remanded") {
		metadata["decision"] = "remanded"
	} else if strings.Contains(text, "dismissed") {
		metadata["decision"] = "dismissed"
	} else if strings.Contains(text, "granted") {
		metadata["decision"] = "granted"
	} else if strings.Contains(text, "denied") {
		metadata["decision"] = "denied"
	}

	// Check for unanimous decision
	if strings.Contains(text, "unanimous") {
		metadata["unanimous"] = true
	}

	// Check for dissenting opinions
	if strings.Contains(text, "dissent") || strings.Contains(text, "dissenting") {
		metadata["has_dissent"] = true
	}

	// Check for concurring opinions
	if strings.Contains(text, "concur") || strings.Contains(text, "concurring") {
		metadata["has_concurrence"] = true
	}

	return metadata
}

// ApplyJurisdictionRules applies jurisdiction-specific rules to the case
func (me *MetadataEnricher) ApplyJurisdictionRules(c *models.Case) {
	rules := me.rules.GetRulesForJurisdiction(c.Jurisdiction)
	if rules == nil {
		return
	}

	// Apply citation format validation
	if rules.CitationPattern != "" {
		matched, _ := regexp.MatchString(rules.CitationPattern, c.CaseNumber)
		if c.Metadata == nil {
			c.Metadata = make(map[string]interface{})
		}
		c.Metadata["citation_valid"] = matched
	}

	// Apply date format rules
	if c.DecisionDate != nil && rules.MinDate != nil {
		if c.DecisionDate.Before(*rules.MinDate) {
			c.Metadata["date_warning"] = "date before minimum expected date"
		}
	}

	// Apply court-specific rules
	if rules.RequiredFields != nil {
		missing := me.checkRequiredFields(c, rules.RequiredFields)
		if len(missing) > 0 {
			c.Metadata["missing_fields"] = missing
		}
	}
}

// checkRequiredFields checks if required fields are present
func (me *MetadataEnricher) checkRequiredFields(c *models.Case, required []string) []string {
	var missing []string

	for _, field := range required {
		switch field {
		case "case_number":
			if c.CaseNumber == "" {
				missing = append(missing, field)
			}
		case "decision_date":
			if c.DecisionDate == nil {
				missing = append(missing, field)
			}
		case "court":
			if c.Court == "" {
				missing = append(missing, field)
			}
		case "judges":
			if len(c.Judges) == 0 {
				missing = append(missing, field)
			}
		}
	}

	return missing
}

// containsAny checks if text contains any of the patterns
func containsAny(text string, patterns []string) bool {
	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

// JurisdictionRules contains rules for a specific jurisdiction
type JurisdictionRules struct {
	rules map[string]*RuleSet
}

// RuleSet contains validation and processing rules for a jurisdiction
type RuleSet struct {
	Jurisdiction    string    `yaml:"jurisdiction"`
	CitationPattern string    `yaml:"citation_pattern"`
	DateFormat      string    `yaml:"date_format"`
	MinDate         *time.Time `yaml:"min_date,omitempty"`
	MaxDate         *time.Time `yaml:"max_date,omitempty"`
	RequiredFields  []string  `yaml:"required_fields,omitempty"`
	CourtAbbreviations map[string]string `yaml:"court_abbreviations,omitempty"`
}

// NewJurisdictionRules creates a new jurisdiction rules system
func NewJurisdictionRules() *JurisdictionRules {
	jr := &JurisdictionRules{
		rules: make(map[string]*RuleSet),
	}
	jr.initializeDefaults()
	return jr
}

// initializeDefaults initializes default rules for major jurisdictions
func (jr *JurisdictionRules) initializeDefaults() {
	// United States
	jr.rules["United States"] = &RuleSet{
		Jurisdiction:    "United States",
		CitationPattern: `\d+\s+[A-Z]\.?\s*\d+`,
		DateFormat:      "2006-01-02",
		RequiredFields:  []string{"case_number", "court"},
	}

	// United Kingdom
	jr.rules["United Kingdom"] = &RuleSet{
		Jurisdiction:    "United Kingdom",
		CitationPattern: `\[\d{4}\]\s+[A-Z]+\s+\d+`,
		DateFormat:      "2006-01-02",
		RequiredFields:  []string{"case_number", "court"},
	}

	// Canada
	jr.rules["Canada"] = &RuleSet{
		Jurisdiction:    "Canada",
		CitationPattern: `\d{4}\s+[A-Z]+\s+\d+`,
		DateFormat:      "2006-01-02",
		RequiredFields:  []string{"case_number", "court"},
	}

	// Australia
	jr.rules["Australia"] = &RuleSet{
		Jurisdiction:    "Australia",
		CitationPattern: `\[\d{4}\]\s+[A-Z]+\s+\d+`,
		DateFormat:      "2006-01-02",
		RequiredFields:  []string{"case_number", "court"},
	}

	// India
	jr.rules["India"] = &RuleSet{
		Jurisdiction:    "India",
		DateFormat:      "02.01.2006",
		RequiredFields:  []string{"case_number"},
	}
}

// GetRulesForJurisdiction retrieves rules for a specific jurisdiction
func (jr *JurisdictionRules) GetRulesForJurisdiction(jurisdiction string) *RuleSet {
	return jr.rules[jurisdiction]
}

// AddRules adds custom rules for a jurisdiction
func (jr *JurisdictionRules) AddRules(rules *RuleSet) {
	jr.rules[rules.Jurisdiction] = rules
}
