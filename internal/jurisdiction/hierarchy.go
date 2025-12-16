package jurisdiction

import (
	"strings"

	"github.com/gongahkia/kite/pkg/models"
)

// CourtHierarchy represents the hierarchical structure of courts
type CourtHierarchy struct {
	levels map[string]models.CourtLevel
	courts map[string]*CourtInfo
}

// CourtInfo contains metadata about a specific court
type CourtInfo struct {
	Name         string            `json:"name"`
	Abbreviation string            `json:"abbreviation"`
	Jurisdiction string            `json:"jurisdiction"`
	Level        models.CourtLevel `json:"level"`
	Type         CourtType         `json:"type"`
	ParentCourt  string            `json:"parent_court,omitempty"`
	Precedential bool              `json:"precedential"`
	Active       bool              `json:"active"`
}

// CourtType represents the type of court
type CourtType string

const (
	CourtTypeSupreme      CourtType = "supreme"
	CourtTypeAppellate    CourtType = "appellate"
	CourtTypeTrial        CourtType = "trial"
	CourtTypeSpecialized  CourtType = "specialized"
	CourtTypeAdministrative CourtType = "administrative"
	CourtTypeMilitary     CourtType = "military"
	CourtTypeInternational CourtType = "international"
)

// NewCourtHierarchy creates a new court hierarchy system
func NewCourtHierarchy() *CourtHierarchy {
	ch := &CourtHierarchy{
		levels: make(map[string]models.CourtLevel),
		courts: make(map[string]*CourtInfo),
	}
	ch.initializeDefaults()
	return ch
}

// initializeDefaults initializes default court hierarchies for major jurisdictions
func (ch *CourtHierarchy) initializeDefaults() {
	// United States
	ch.registerCourt(&CourtInfo{
		Name:         "Supreme Court of the United States",
		Abbreviation: "SCOTUS",
		Jurisdiction: "United States",
		Level:        models.CourtLevelSupreme,
		Type:         CourtTypeSupreme,
		Precedential: true,
		Active:       true,
	})
	ch.registerCourt(&CourtInfo{
		Name:         "U.S. Court of Appeals",
		Abbreviation: "Circuit",
		Jurisdiction: "United States",
		Level:        models.CourtLevelAppellate,
		Type:         CourtTypeAppellate,
		ParentCourt:  "SCOTUS",
		Precedential: true,
		Active:       true,
	})
	ch.registerCourt(&CourtInfo{
		Name:         "U.S. District Court",
		Abbreviation: "District",
		Jurisdiction: "United States",
		Level:        models.CourtLevelTrial,
		Type:         CourtTypeTrial,
		ParentCourt:  "Circuit",
		Precedential: false,
		Active:       true,
	})

	// United Kingdom
	ch.registerCourt(&CourtInfo{
		Name:         "UK Supreme Court",
		Abbreviation: "UKSC",
		Jurisdiction: "United Kingdom",
		Level:        models.CourtLevelSupreme,
		Type:         CourtTypeSupreme,
		Precedential: true,
		Active:       true,
	})
	ch.registerCourt(&CourtInfo{
		Name:         "Court of Appeal (England & Wales)",
		Abbreviation: "EWCA",
		Jurisdiction: "United Kingdom",
		Level:        models.CourtLevelAppellate,
		Type:         CourtTypeAppellate,
		ParentCourt:  "UKSC",
		Precedential: true,
		Active:       true,
	})
	ch.registerCourt(&CourtInfo{
		Name:         "High Court (England & Wales)",
		Abbreviation: "EWHC",
		Jurisdiction: "United Kingdom",
		Level:        models.CourtLevelTrial,
		Type:         CourtTypeTrial,
		ParentCourt:  "EWCA",
		Precedential: false,
		Active:       true,
	})

	// Canada
	ch.registerCourt(&CourtInfo{
		Name:         "Supreme Court of Canada",
		Abbreviation: "SCC",
		Jurisdiction: "Canada",
		Level:        models.CourtLevelSupreme,
		Type:         CourtTypeSupreme,
		Precedential: true,
		Active:       true,
	})
	ch.registerCourt(&CourtInfo{
		Name:         "Federal Court of Appeal",
		Abbreviation: "FCA",
		Jurisdiction: "Canada",
		Level:        models.CourtLevelAppellate,
		Type:         CourtTypeAppellate,
		ParentCourt:  "SCC",
		Precedential: true,
		Active:       true,
	})

	// Australia
	ch.registerCourt(&CourtInfo{
		Name:         "High Court of Australia",
		Abbreviation: "HCA",
		Jurisdiction: "Australia",
		Level:        models.CourtLevelSupreme,
		Type:         CourtTypeSupreme,
		Precedential: true,
		Active:       true,
	})
	ch.registerCourt(&CourtInfo{
		Name:         "Federal Court of Australia",
		Abbreviation: "FCA",
		Jurisdiction: "Australia",
		Level:        models.CourtLevelAppellate,
		Type:         CourtTypeAppellate,
		ParentCourt:  "HCA",
		Precedential: true,
		Active:       true,
	})

	// Hong Kong
	ch.registerCourt(&CourtInfo{
		Name:         "Court of Final Appeal",
		Abbreviation: "HKCFA",
		Jurisdiction: "Hong Kong",
		Level:        models.CourtLevelSupreme,
		Type:         CourtTypeSupreme,
		Precedential: true,
		Active:       true,
	})
	ch.registerCourt(&CourtInfo{
		Name:         "Court of Appeal",
		Abbreviation: "HKCA",
		Jurisdiction: "Hong Kong",
		Level:        models.CourtLevelAppellate,
		Type:         CourtTypeAppellate,
		ParentCourt:  "HKCFA",
		Precedential: true,
		Active:       true,
	})

	// India
	ch.registerCourt(&CourtInfo{
		Name:         "Supreme Court of India",
		Abbreviation: "SCI",
		Jurisdiction: "India",
		Level:        models.CourtLevelSupreme,
		Type:         CourtTypeSupreme,
		Precedential: true,
		Active:       true,
	})
	ch.registerCourt(&CourtInfo{
		Name:         "High Court",
		Abbreviation: "HC",
		Jurisdiction: "India",
		Level:        models.CourtLevelAppellate,
		Type:         CourtTypeAppellate,
		ParentCourt:  "SCI",
		Precedential: true,
		Active:       true,
	})

	// New Zealand
	ch.registerCourt(&CourtInfo{
		Name:         "Supreme Court of New Zealand",
		Abbreviation: "NZSC",
		Jurisdiction: "New Zealand",
		Level:        models.CourtLevelSupreme,
		Type:         CourtTypeSupreme,
		Precedential: true,
		Active:       true,
	})
	ch.registerCourt(&CourtInfo{
		Name:         "Court of Appeal of New Zealand",
		Abbreviation: "NZCA",
		Jurisdiction: "New Zealand",
		Level:        models.CourtLevelAppellate,
		Type:         CourtTypeAppellate,
		ParentCourt:  "NZSC",
		Precedential: true,
		Active:       true,
	})

	// South Africa
	ch.registerCourt(&CourtInfo{
		Name:         "Constitutional Court of South Africa",
		Abbreviation: "ZACC",
		Jurisdiction: "South Africa",
		Level:        models.CourtLevelSupreme,
		Type:         CourtTypeSupreme,
		Precedential: true,
		Active:       true,
	})
	ch.registerCourt(&CourtInfo{
		Name:         "Supreme Court of Appeal of South Africa",
		Abbreviation: "ZASCA",
		Jurisdiction: "South Africa",
		Level:        models.CourtLevelAppellate,
		Type:         CourtTypeAppellate,
		ParentCourt:  "ZACC",
		Precedential: true,
		Active:       true,
	})

	// Singapore
	ch.registerCourt(&CourtInfo{
		Name:         "Court of Appeal of Singapore",
		Abbreviation: "SGCA",
		Jurisdiction: "Singapore",
		Level:        models.CourtLevelSupreme,
		Type:         CourtTypeSupreme,
		Precedential: true,
		Active:       true,
	})
	ch.registerCourt(&CourtInfo{
		Name:         "High Court of Singapore",
		Abbreviation: "SGHC",
		Jurisdiction: "Singapore",
		Level:        models.CourtLevelTrial,
		Type:         CourtTypeTrial,
		ParentCourt:  "SGCA",
		Precedential: false,
		Active:       true,
	})
}

// registerCourt registers a court in the hierarchy
func (ch *CourtHierarchy) registerCourt(info *CourtInfo) {
	ch.courts[info.Abbreviation] = info
	ch.courts[info.Name] = info
	ch.levels[info.Name] = info.Level
}

// GetCourtLevel determines the court level from court name or abbreviation
func (ch *CourtHierarchy) GetCourtLevel(courtName string) models.CourtLevel {
	// Try exact match first
	if level, ok := ch.levels[courtName]; ok {
		return level
	}

	// Try case-insensitive match
	courtNameLower := strings.ToLower(courtName)
	for name, level := range ch.levels {
		if strings.ToLower(name) == courtNameLower {
			return level
		}
	}

	// Try pattern matching for common court names
	if strings.Contains(courtNameLower, "supreme") || strings.Contains(courtNameLower, "final appeal") {
		return models.CourtLevelSupreme
	}
	if strings.Contains(courtNameLower, "appeal") || strings.Contains(courtNameLower, "appellate") {
		return models.CourtLevelAppellate
	}
	if strings.Contains(courtNameLower, "high court") && !strings.Contains(courtNameLower, "appeal") {
		return models.CourtLevelTrial
	}
	if strings.Contains(courtNameLower, "district") || strings.Contains(courtNameLower, "magistrate") {
		return models.CourtLevelLower
	}

	// Default to trial level
	return models.CourtLevelTrial
}

// GetCourtInfo retrieves detailed information about a court
func (ch *CourtHierarchy) GetCourtInfo(courtNameOrAbbr string) (*CourtInfo, bool) {
	info, ok := ch.courts[courtNameOrAbbr]
	if ok {
		return info, true
	}

	// Try case-insensitive match
	courtNameLower := strings.ToLower(courtNameOrAbbr)
	for name, info := range ch.courts {
		if strings.ToLower(name) == courtNameLower {
			return info, true
		}
	}

	return nil, false
}

// GetCourtType determines the court type from court name
func (ch *CourtHierarchy) GetCourtType(courtName string) CourtType {
	if info, ok := ch.GetCourtInfo(courtName); ok {
		return info.Type
	}

	// Pattern matching
	courtNameLower := strings.ToLower(courtName)
	if strings.Contains(courtNameLower, "supreme") || strings.Contains(courtNameLower, "final") {
		return CourtTypeSupreme
	}
	if strings.Contains(courtNameLower, "appeal") {
		return CourtTypeAppellate
	}
	if strings.Contains(courtNameLower, "military") || strings.Contains(courtNameLower, "court-martial") {
		return CourtTypeMilitary
	}
	if strings.Contains(courtNameLower, "administrative") {
		return CourtTypeAdministrative
	}
	if strings.Contains(courtNameLower, "tax") || strings.Contains(courtNameLower, "family") ||
	   strings.Contains(courtNameLower, "bankruptcy") {
		return CourtTypeSpecialized
	}

	return CourtTypeTrial
}

// IsPrecedential checks if a court's decisions are precedential
func (ch *CourtHierarchy) IsPrecedential(courtName string) bool {
	if info, ok := ch.GetCourtInfo(courtName); ok {
		return info.Precedential
	}

	// Default: appellate and supreme courts are precedential
	level := ch.GetCourtLevel(courtName)
	return level == models.CourtLevelSupreme || level == models.CourtLevelAppellate
}

// GetParentCourt returns the parent court in the hierarchy
func (ch *CourtHierarchy) GetParentCourt(courtName string) string {
	if info, ok := ch.GetCourtInfo(courtName); ok {
		return info.ParentCourt
	}
	return ""
}

// GetCourtsByJurisdiction returns all courts for a jurisdiction
func (ch *CourtHierarchy) GetCourtsByJurisdiction(jurisdiction string) []*CourtInfo {
	var courts []*CourtInfo
	seen := make(map[string]bool)

	for _, info := range ch.courts {
		if info.Jurisdiction == jurisdiction && !seen[info.Abbreviation] {
			courts = append(courts, info)
			seen[info.Abbreviation] = true
		}
	}

	return courts
}

// GetCourtsByLevel returns all courts at a specific level
func (ch *CourtHierarchy) GetCourtsByLevel(level models.CourtLevel) []*CourtInfo {
	var courts []*CourtInfo
	seen := make(map[string]bool)

	for _, info := range ch.courts {
		if info.Level == level && !seen[info.Abbreviation] {
			courts = append(courts, info)
			seen[info.Abbreviation] = true
		}
	}

	return courts
}
