package compliance

import (
	"fmt"
	"strings"
	"time"

	"github.com/gongahkia/kite/pkg/models"
)

// AttributionHandler manages attribution requirements for scraped data
type AttributionHandler struct {
	policyManager *PolicyManager
}

// NewAttributionHandler creates a new attribution handler
func NewAttributionHandler(pm *PolicyManager) *AttributionHandler {
	return &AttributionHandler{
		policyManager: pm,
	}
}

// AddAttribution adds attribution metadata to a case
func (ah *AttributionHandler) AddAttribution(c *models.Case) error {
	if c.SourceDatabase == "" {
		return nil // No source database specified
	}

	// Get policy for the source
	policy, ok := ah.policyManager.GetPolicy(c.SourceDatabase)
	if !ok {
		// No specific policy, add generic attribution
		ah.addGenericAttribution(c)
		return nil
	}

	// Check if attribution is required
	if !policy.RequiresAttribution {
		return nil
	}

	// Add attribution text
	if c.Metadata == nil {
		c.Metadata = make(map[string]interface{})
	}

	c.Metadata["attribution"] = policy.AttributionText
	c.Metadata["source_url"] = policy.BaseURL
	c.Metadata["terms_of_service"] = policy.TermsOfServiceURL
	c.Metadata["scraped_at"] = time.Now().Format(time.RFC3339)

	// Add attribution footer to full text if available
	if c.FullText != "" && policy.AttributionText != "" {
		c.FullText = ah.appendAttributionToText(c.FullText, policy.AttributionText)
	}

	return nil
}

// addGenericAttribution adds generic attribution when no specific policy exists
func (ah *AttributionHandler) addGenericAttribution(c *models.Case) {
	if c.Metadata == nil {
		c.Metadata = make(map[string]interface{})
	}

	attribution := fmt.Sprintf("Data sourced from %s", c.SourceDatabase)
	if c.URL != "" {
		attribution += fmt.Sprintf(" (%s)", c.URL)
	}

	c.Metadata["attribution"] = attribution
	c.Metadata["scraped_at"] = time.Now().Format(time.RFC3339)
}

// appendAttributionToText appends attribution text to the full text
func (ah *AttributionHandler) appendAttributionToText(fullText, attribution string) string {
	separator := "\n\n" + strings.Repeat("-", 80) + "\n"
	footer := fmt.Sprintf("%sATTRIBUTION:\n%s\n%s", separator, attribution, separator)
	return fullText + footer
}

// GetAttributionRequirements returns attribution requirements for a source
func (ah *AttributionHandler) GetAttributionRequirements(sourceName string) *AttributionRequirements {
	policy, ok := ah.policyManager.GetPolicy(sourceName)
	if !ok {
		return &AttributionRequirements{
			SourceName:  sourceName,
			Required:    false,
			Text:        "",
			Format:      AttributionFormatNone,
			Placement:   AttributionPlacementNone,
		}
	}

	return &AttributionRequirements{
		SourceName:  sourceName,
		Required:    policy.RequiresAttribution,
		Text:        policy.AttributionText,
		Format:      AttributionFormatText,
		Placement:   AttributionPlacementFooter,
		SourceURL:   policy.BaseURL,
		TermsURL:    policy.TermsOfServiceURL,
	}
}

// AttributionRequirements defines the attribution requirements for a source
type AttributionRequirements struct {
	SourceName  string             `json:"source_name"`
	Required    bool               `json:"required"`
	Text        string             `json:"text"`
	Format      AttributionFormat  `json:"format"`
	Placement   AttributionPlacement `json:"placement"`
	SourceURL   string             `json:"source_url,omitempty"`
	TermsURL    string             `json:"terms_url,omitempty"`
}

// AttributionFormat represents the format of attribution
type AttributionFormat string

const (
	AttributionFormatNone     AttributionFormat = "none"
	AttributionFormatText     AttributionFormat = "text"
	AttributionFormatHTML     AttributionFormat = "html"
	AttributionFormatMarkdown AttributionFormat = "markdown"
	AttributionFormatJSON     AttributionFormat = "json"
)

// AttributionPlacement represents where attribution should be placed
type AttributionPlacement string

const (
	AttributionPlacementNone   AttributionPlacement = "none"
	AttributionPlacementHeader AttributionPlacement = "header"
	AttributionPlacementFooter AttributionPlacement = "footer"
	AttributionPlacementInline AttributionPlacement = "inline"
	AttributionPlacementMetadata AttributionPlacement = "metadata"
)

// FormatAttribution formats attribution text according to the specified format
func (ah *AttributionHandler) FormatAttribution(sourceName string, format AttributionFormat) string {
	policy, ok := ah.policyManager.GetPolicy(sourceName)
	if !ok || !policy.RequiresAttribution {
		return ""
	}

	switch format {
	case AttributionFormatText:
		return policy.AttributionText

	case AttributionFormatHTML:
		return fmt.Sprintf(
			`<div class="attribution"><p>%s</p><p><a href="%s">%s</a></p></div>`,
			policy.AttributionText,
			policy.BaseURL,
			policy.BaseURL,
		)

	case AttributionFormatMarkdown:
		return fmt.Sprintf(
			"**Attribution:** %s\n\nSource: [%s](%s)",
			policy.AttributionText,
			policy.SourceName,
			policy.BaseURL,
		)

	case AttributionFormatJSON:
		return fmt.Sprintf(
			`{"attribution":"%s","source_url":"%s","terms_url":"%s"}`,
			policy.AttributionText,
			policy.BaseURL,
			policy.TermsOfServiceURL,
		)

	default:
		return policy.AttributionText
	}
}

// ValidateAttribution checks if proper attribution is present in a case
func (ah *AttributionHandler) ValidateAttribution(c *models.Case) (bool, []string) {
	var issues []string

	if c.SourceDatabase == "" {
		issues = append(issues, "No source database specified")
		return false, issues
	}

	policy, ok := ah.policyManager.GetPolicy(c.SourceDatabase)
	if !ok {
		// No specific policy, check for generic attribution
		if c.Metadata == nil || c.Metadata["attribution"] == nil {
			issues = append(issues, "Missing attribution metadata")
		}
		return len(issues) == 0, issues
	}

	if !policy.RequiresAttribution {
		return true, nil // Attribution not required
	}

	// Check metadata attribution
	if c.Metadata == nil || c.Metadata["attribution"] == nil {
		issues = append(issues, "Missing attribution in metadata")
	}

	// Check source URL
	if c.Metadata == nil || c.Metadata["source_url"] == nil {
		issues = append(issues, "Missing source URL in metadata")
	}

	return len(issues) == 0, issues
}

// GenerateCitation generates a proper legal citation with attribution
func (ah *AttributionHandler) GenerateCitation(c *models.Case, style CitationStyle) string {
	switch style {
	case CitationStyleBluebook:
		return ah.generateBluebookCitation(c)
	case CitationStyleAPA:
		return ah.generateAPACitation(c)
	case CitationStyleMLA:
		return ah.generateMLACitation(c)
	default:
		return ah.generatePlainCitation(c)
	}
}

// CitationStyle represents different citation formats
type CitationStyle string

const (
	CitationStylePlain    CitationStyle = "plain"
	CitationStyleBluebook CitationStyle = "bluebook"
	CitationStyleAPA      CitationStyle = "apa"
	CitationStyleMLA      CitationStyle = "mla"
)

// generateBluebookCitation generates a Bluebook-style citation
func (ah *AttributionHandler) generateBluebookCitation(c *models.Case) string {
	// Example: Case Name, Citation, Court Date
	citation := c.CaseName
	if c.CaseNumber != "" {
		citation += ", " + c.CaseNumber
	}
	if c.Court != "" {
		citation += " (" + c.Court
		if c.DecisionDate != nil {
			citation += " " + c.DecisionDate.Format("Jan. 2, 2006")
		}
		citation += ")"
	}

	// Add source attribution
	if attribution := ah.policyManager.GetAttributionText(c.SourceDatabase); attribution != "" {
		citation += ". " + attribution
	}

	return citation
}

// generateAPACitation generates an APA-style citation
func (ah *AttributionHandler) generateAPACitation(c *models.Case) string {
	// Example: Case Name, Citation (Court Year)
	citation := c.CaseName
	if c.CaseNumber != "" {
		citation += ", " + c.CaseNumber
	}
	if c.Court != "" && c.DecisionDate != nil {
		citation += fmt.Sprintf(" (%s %d)", c.Court, c.DecisionDate.Year())
	}

	if c.URL != "" {
		citation += ". Retrieved from " + c.URL
	}

	return citation
}

// generateMLACitation generates an MLA-style citation
func (ah *AttributionHandler) generateMLACitation(c *models.Case) string {
	// Example: Case Name. Citation. Court. Date.
	citation := c.CaseName + "."
	if c.CaseNumber != "" {
		citation += " " + c.CaseNumber + "."
	}
	if c.Court != "" {
		citation += " " + c.Court + "."
	}
	if c.DecisionDate != nil {
		citation += " " + c.DecisionDate.Format("2 Jan. 2006") + "."
	}

	if c.SourceDatabase != "" {
		citation += " " + c.SourceDatabase + "."
	}

	if c.URL != "" {
		citation += " Web."
	}

	return citation
}

// generatePlainCitation generates a plain text citation
func (ah *AttributionHandler) generatePlainCitation(c *models.Case) string {
	citation := c.CaseName
	if c.CaseNumber != "" {
		citation += " [" + c.CaseNumber + "]"
	}
	if c.Court != "" {
		citation += ", " + c.Court
	}
	if c.DecisionDate != nil {
		citation += ", " + c.DecisionDate.Format("2 January 2006")
	}
	if c.SourceDatabase != "" {
		citation += " (Source: " + c.SourceDatabase + ")"
	}
	return citation
}
