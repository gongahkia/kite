package export

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gongahkia/kite/pkg/models"
)

// ExportFormat represents different export formats
type ExportFormat string

const (
	FormatJSON      ExportFormat = "json"
	FormatJSONLines ExportFormat = "jsonlines"
	FormatCSV       ExportFormat = "csv"
	FormatXML       ExportFormat = "xml"
	FormatBibTeX    ExportFormat = "bibtex"
	FormatMarkdown  ExportFormat = "markdown"
	FormatPlainText ExportFormat = "text"
)

// Exporter handles exporting cases in different formats
type Exporter struct {
	format ExportFormat
	writer io.Writer
}

// NewExporter creates a new exporter
func NewExporter(format ExportFormat, writer io.Writer) *Exporter {
	return &Exporter{
		format: format,
		writer: writer,
	}
}

// Export exports a slice of cases
func (e *Exporter) Export(cases []*models.Case) error {
	switch e.format {
	case FormatJSON:
		return e.exportJSON(cases)
	case FormatJSONLines:
		return e.exportJSONLines(cases)
	case FormatCSV:
		return e.exportCSV(cases)
	case FormatXML:
		return e.exportXML(cases)
	case FormatBibTeX:
		return e.exportBibTeX(cases)
	case FormatMarkdown:
		return e.exportMarkdown(cases)
	case FormatPlainText:
		return e.exportPlainText(cases)
	default:
		return fmt.Errorf("unsupported export format: %s", e.format)
	}
}

// exportJSON exports cases as pretty-printed JSON array
func (e *Exporter) exportJSON(cases []*models.Case) error {
	encoder := json.NewEncoder(e.writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cases)
}

// exportJSONLines exports cases as newline-delimited JSON (one case per line)
func (e *Exporter) exportJSONLines(cases []*models.Case) error {
	encoder := json.NewEncoder(e.writer)
	for _, c := range cases {
		if err := encoder.Encode(c); err != nil {
			return err
		}
	}
	return nil
}

// exportCSV exports cases as CSV
func (e *Exporter) exportCSV(cases []*models.Case) error {
	writer := csv.NewWriter(e.writer)
	defer writer.Flush()

	// Write header
	header := []string{
		"ID",
		"CaseName",
		"CaseNumber",
		"Court",
		"Jurisdiction",
		"DecisionDate",
		"URL",
		"Judges",
		"Summary",
		"SourceDatabase",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write rows
	for _, c := range cases {
		var decisionDate string
		if c.DecisionDate != nil {
			decisionDate = c.DecisionDate.Format("2006-01-02")
		}

		row := []string{
			c.ID,
			c.CaseName,
			c.CaseNumber,
			c.Court,
			c.Jurisdiction,
			decisionDate,
			c.URL,
			strings.Join(c.Judges, "; "),
			c.Summary,
			c.SourceDatabase,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// exportXML exports cases as XML
func (e *Exporter) exportXML(cases []*models.Case) error {
	type CasesWrapper struct {
		XMLName xml.Name      `xml:"cases"`
		Cases   []*models.Case `xml:"case"`
	}

	wrapper := CasesWrapper{Cases: cases}
	encoder := xml.NewEncoder(e.writer)
	encoder.Indent("", "  ")

	// Write XML header
	if _, err := e.writer.Write([]byte(xml.Header)); err != nil {
		return err
	}

	return encoder.Encode(wrapper)
}

// exportBibTeX exports cases as BibTeX entries
func (e *Exporter) exportBibTeX(cases []*models.Case) error {
	for _, c := range cases {
		entry := formatBibTeXEntry(c)
		if _, err := e.writer.Write([]byte(entry + "\n\n")); err != nil {
			return err
		}
	}
	return nil
}

// formatBibTeXEntry formats a single case as a BibTeX entry
func formatBibTeXEntry(c *models.Case) string {
	var sb strings.Builder

	// Use case ID or generate one
	citeKey := c.ID
	if citeKey == "" {
		citeKey = strings.ReplaceAll(c.CaseNumber, " ", "_")
	}

	sb.WriteString(fmt.Sprintf("@misc{%s,\n", citeKey))
	sb.WriteString(fmt.Sprintf("  title = {{%s}},\n", c.CaseName))

	if c.Court != "" {
		sb.WriteString(fmt.Sprintf("  institution = {{%s}},\n", c.Court))
	}

	if c.DecisionDate != nil {
		sb.WriteString(fmt.Sprintf("  year = {%d},\n", c.DecisionDate.Year()))
		sb.WriteString(fmt.Sprintf("  month = {%s},\n", c.DecisionDate.Format("Jan")))
	}

	if c.CaseNumber != "" {
		sb.WriteString(fmt.Sprintf("  number = {{%s}},\n", c.CaseNumber))
	}

	if c.URL != "" {
		sb.WriteString(fmt.Sprintf("  url = {%s},\n", c.URL))
	}

	if c.SourceDatabase != "" {
		sb.WriteString(fmt.Sprintf("  note = {{Source: %s}},\n", c.SourceDatabase))
	}

	sb.WriteString("}")

	return sb.String()
}

// exportMarkdown exports cases as Markdown
func (e *Exporter) exportMarkdown(cases []*models.Case) error {
	for i, c := range cases {
		md := formatMarkdown(c)
		if _, err := e.writer.Write([]byte(md)); err != nil {
			return err
		}
		if i < len(cases)-1 {
			if _, err := e.writer.Write([]byte("\n---\n\n")); err != nil {
				return err
			}
		}
	}
	return nil
}

// formatMarkdown formats a single case as Markdown
func formatMarkdown(c *models.Case) string {
	var sb strings.Builder

	// Title
	sb.WriteString(fmt.Sprintf("# %s\n\n", c.CaseName))

	// Metadata table
	sb.WriteString("| Field | Value |\n")
	sb.WriteString("|-------|-------|\n")

	if c.CaseNumber != "" {
		sb.WriteString(fmt.Sprintf("| **Citation** | %s |\n", c.CaseNumber))
	}
	if c.Court != "" {
		sb.WriteString(fmt.Sprintf("| **Court** | %s |\n", c.Court))
	}
	if c.Jurisdiction != "" {
		sb.WriteString(fmt.Sprintf("| **Jurisdiction** | %s |\n", c.Jurisdiction))
	}
	if c.DecisionDate != nil {
		sb.WriteString(fmt.Sprintf("| **Date** | %s |\n", c.DecisionDate.Format("January 2, 2006")))
	}
	if len(c.Judges) > 0 {
		sb.WriteString(fmt.Sprintf("| **Judges** | %s |\n", strings.Join(c.Judges, ", ")))
	}
	if c.URL != "" {
		sb.WriteString(fmt.Sprintf("| **URL** | [Link](%s) |\n", c.URL))
	}
	if c.SourceDatabase != "" {
		sb.WriteString(fmt.Sprintf("| **Source** | %s |\n", c.SourceDatabase))
	}

	sb.WriteString("\n")

	// Summary
	if c.Summary != "" {
		sb.WriteString("## Summary\n\n")
		sb.WriteString(c.Summary)
		sb.WriteString("\n\n")
	}

	// Legal concepts
	if len(c.LegalConcepts) > 0 {
		sb.WriteString("## Legal Concepts\n\n")
		for _, concept := range c.LegalConcepts {
			sb.WriteString(fmt.Sprintf("- %s\n", concept))
		}
		sb.WriteString("\n")
	}

	// Full text (truncated in markdown for readability)
	if c.FullText != "" {
		sb.WriteString("## Full Text\n\n")
		// Truncate if too long
		maxLen := 1000
		if len(c.FullText) > maxLen {
			sb.WriteString(c.FullText[:maxLen])
			sb.WriteString("\n\n*[Truncated]*\n\n")
		} else {
			sb.WriteString(c.FullText)
			sb.WriteString("\n\n")
		}
	}

	return sb.String()
}

// exportPlainText exports cases as plain text
func (e *Exporter) exportPlainText(cases []*models.Case) error {
	for i, c := range cases {
		text := formatPlainText(c)
		if _, err := e.writer.Write([]byte(text)); err != nil {
			return err
		}
		if i < len(cases)-1 {
			if _, err := e.writer.Write([]byte("\n" + strings.Repeat("=", 80) + "\n\n")); err != nil {
				return err
			}
		}
	}
	return nil
}

// formatPlainText formats a single case as plain text
func formatPlainText(c *models.Case) string {
	var sb strings.Builder

	sb.WriteString(c.CaseName + "\n")
	sb.WriteString(strings.Repeat("-", len(c.CaseName)) + "\n\n")

	if c.CaseNumber != "" {
		sb.WriteString(fmt.Sprintf("Citation: %s\n", c.CaseNumber))
	}
	if c.Court != "" {
		sb.WriteString(fmt.Sprintf("Court: %s\n", c.Court))
	}
	if c.Jurisdiction != "" {
		sb.WriteString(fmt.Sprintf("Jurisdiction: %s\n", c.Jurisdiction))
	}
	if c.DecisionDate != nil {
		sb.WriteString(fmt.Sprintf("Date: %s\n", c.DecisionDate.Format("January 2, 2006")))
	}
	if len(c.Judges) > 0 {
		sb.WriteString(fmt.Sprintf("Judges: %s\n", strings.Join(c.Judges, ", ")))
	}
	if c.URL != "" {
		sb.WriteString(fmt.Sprintf("URL: %s\n", c.URL))
	}
	if c.SourceDatabase != "" {
		sb.WriteString(fmt.Sprintf("Source: %s\n", c.SourceDatabase))
	}

	if c.Summary != "" {
		sb.WriteString(fmt.Sprintf("\nSummary:\n%s\n", c.Summary))
	}

	if len(c.LegalConcepts) > 0 {
		sb.WriteString(fmt.Sprintf("\nLegal Concepts: %s\n", strings.Join(c.LegalConcepts, ", ")))
	}

	if c.FullText != "" {
		sb.WriteString(fmt.Sprintf("\nFull Text:\n%s\n", c.FullText))
	}

	return sb.String()
}

// ExportOptions holds options for export operations
type ExportOptions struct {
	Format      ExportFormat `json:"format"`
	Fields      []string     `json:"fields,omitempty"`      // For CSV: custom fields
	Compress    bool         `json:"compress"`              // Whether to gzip compress
	Pretty      bool         `json:"pretty"`                // For JSON: pretty print
	IncludeFullText bool     `json:"include_full_text"`    // Whether to include full case text
	DateFormat  string       `json:"date_format,omitempty"` // Custom date format
}

// DefaultExportOptions returns default export options
func DefaultExportOptions() *ExportOptions {
	return &ExportOptions{
		Format:          FormatJSON,
		Compress:        false,
		Pretty:          true,
		IncludeFullText: true,
		DateFormat:      time.RFC3339,
	}
}
