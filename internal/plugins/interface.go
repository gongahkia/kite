package plugins

import (
	"context"

	"github.com/gongahkia/kite/pkg/models"
)

// Plugin is the base interface that all plugins must implement
type Plugin interface {
	// Name returns the plugin name
	Name() string

	// Version returns the plugin version
	Version() string

	// Description returns a description of the plugin
	Description() string

	// Init initializes the plugin with configuration
	Init(config map[string]interface{}) error

	// Start starts the plugin
	Start(ctx context.Context) error

	// Stop stops the plugin gracefully
	Stop(ctx context.Context) error

	// Health checks if the plugin is healthy
	Health() error
}

// ScraperPlugin extends Plugin for custom scrapers
type ScraperPlugin interface {
	Plugin

	// SearchCases searches for cases matching the query
	SearchCases(ctx context.Context, query SearchQuery) ([]*models.Case, error)

	// GetCaseByID retrieves a case by its ID
	GetCaseByID(ctx context.Context, id string) (*models.Case, error)

	// GetCasesByDateRange retrieves cases within a date range
	GetCasesByDateRange(ctx context.Context, start, end string) ([]*models.Case, error)

	// IsAvailable checks if the scraper source is available
	IsAvailable(ctx context.Context) bool

	// GetJurisdiction returns the jurisdiction this scraper handles
	GetJurisdiction() string
}

// ProcessorPlugin extends Plugin for data processing
type ProcessorPlugin interface {
	Plugin

	// Process processes a case and returns the modified case
	Process(ctx context.Context, c *models.Case) (*models.Case, error)

	// CanProcess checks if this processor can handle the case
	CanProcess(c *models.Case) bool

	// Priority returns the processing priority (higher = earlier)
	Priority() int
}

// ValidatorPlugin extends Plugin for custom validation
type ValidatorPlugin interface {
	Plugin

	// Validate validates a case and returns validation errors
	Validate(ctx context.Context, c *models.Case) ([]ValidationError, error)

	// Severity returns the severity level of this validator
	Severity() ValidationSeverity
}

// ExporterPlugin extends Plugin for custom export formats
type ExporterPlugin interface {
	Plugin

	// Export exports cases to the plugin's format
	Export(ctx context.Context, cases []*models.Case) ([]byte, error)

	// FileExtension returns the file extension for exported files
	FileExtension() string

	// MIMEType returns the MIME type for exported files
	MIMEType() string
}

// MiddlewarePlugin extends Plugin for HTTP middleware
type MiddlewarePlugin interface {
	Plugin

	// Handler returns the HTTP middleware handler
	Handler() func(next func(c interface{}) error) func(c interface{}) error

	// Order returns the middleware execution order (lower = earlier)
	Order() int
}

// SearchQuery represents a search query
type SearchQuery struct {
	Query        string
	Jurisdiction string
	Court        string
	StartDate    string
	EndDate      string
	Limit        int
	Offset       int
}

// ValidationError represents a validation error
type ValidationError struct {
	Field    string
	Message  string
	Severity ValidationSeverity
}

// ValidationSeverity represents the severity of a validation error
type ValidationSeverity string

const (
	SeverityError   ValidationSeverity = "error"
	SeverityWarning ValidationSeverity = "warning"
	SeverityInfo    ValidationSeverity = "info"
)

// Metadata contains plugin metadata
type Metadata struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	License     string                 `json:"license"`
	Homepage    string                 `json:"homepage"`
	Type        PluginType             `json:"type"`
	Config      map[string]interface{} `json:"config"`
}

// PluginType represents the type of plugin
type PluginType string

const (
	TypeScraper    PluginType = "scraper"
	TypeProcessor  PluginType = "processor"
	TypeValidator  PluginType = "validator"
	TypeExporter   PluginType = "exporter"
	TypeMiddleware PluginType = "middleware"
)
