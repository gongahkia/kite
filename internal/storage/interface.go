package storage

import (
	"context"
	"time"

	"github.com/gongahkia/kite/pkg/models"
)

// Storage defines the interface for data storage implementations
type Storage interface {
	// Case operations
	SaveCase(ctx context.Context, c *models.Case) error
	GetCase(ctx context.Context, id string) (*models.Case, error)
	UpdateCase(ctx context.Context, c *models.Case) error
	DeleteCase(ctx context.Context, id string) error
	ListCases(ctx context.Context, filter CaseFilter) ([]*models.Case, error)
	CountCases(ctx context.Context, filter CaseFilter) (int64, error)

	// Judge operations
	SaveJudge(ctx context.Context, j *models.Judge) error
	GetJudge(ctx context.Context, id string) (*models.Judge, error)
	UpdateJudge(ctx context.Context, j *models.Judge) error
	ListJudges(ctx context.Context, filter JudgeFilter) ([]*models.Judge, error)

	// Citation operations
	SaveCitation(ctx context.Context, c *models.Citation) error
	GetCitation(ctx context.Context, id string) (*models.Citation, error)
	ListCitations(ctx context.Context, filter CitationFilter) ([]*models.Citation, error)

	// Search operations
	SearchCases(ctx context.Context, query SearchQuery) ([]*models.Case, error)

	// Transaction operations (optional, nil if not supported)
	BeginTx(ctx context.Context) (Transaction, error)

	// Health check
	Ping(ctx context.Context) error

	// Close connection
	Close() error
}

// Transaction represents a database transaction
type Transaction interface {
	// Execute operations within the transaction
	SaveCase(ctx context.Context, c *models.Case) error
	SaveJudge(ctx context.Context, j *models.Judge) error
	SaveCitation(ctx context.Context, c *models.Citation) error

	// Transaction control
	Commit() error
	Rollback() error
}

// CaseFilter represents filters for case queries
type CaseFilter struct {
	IDs          []string               `json:"ids,omitempty"`
	Jurisdiction string                 `json:"jurisdiction,omitempty"`
	Court        string                 `json:"court,omitempty"`
	CourtLevel   *models.CourtLevel     `json:"court_level,omitempty"`
	StartDate    *time.Time             `json:"start_date,omitempty"`
	EndDate      *time.Time             `json:"end_date,omitempty"`
	Status       models.CaseStatus      `json:"status,omitempty"`
	Judges       []string               `json:"judges,omitempty"`
	Concepts     []string               `json:"concepts,omitempty"`
	MinQuality   float64                `json:"min_quality,omitempty"`
	Limit        int                    `json:"limit,omitempty"`
	Offset       int                    `json:"offset,omitempty"`
	OrderBy      string                 `json:"order_by,omitempty"`
	OrderDesc    bool                   `json:"order_desc,omitempty"`
}

// JudgeFilter represents filters for judge queries
type JudgeFilter struct {
	Name         string     `json:"name,omitempty"`
	Court        string     `json:"court,omitempty"`
	Jurisdiction string     `json:"jurisdiction,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Offset       int        `json:"offset,omitempty"`
}

// CitationFilter represents filters for citation queries
type CitationFilter struct {
	CaseID       string     `json:"case_id,omitempty"`
	Format       string     `json:"format,omitempty"`
	Year         int        `json:"year,omitempty"`
	Valid        *bool      `json:"valid,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Offset       int        `json:"offset,omitempty"`
}

// SearchQuery represents a search query
type SearchQuery struct {
	Query        string     `json:"query"`
	Fields       []string   `json:"fields,omitempty"`
	Filters      CaseFilter `json:"filters,omitempty"`
	Fuzzy        bool       `json:"fuzzy,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Offset       int        `json:"offset,omitempty"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// StorageStats represents storage statistics
type StorageStats struct {
	TotalCases     int64     `json:"total_cases"`
	TotalJudges    int64     `json:"total_judges"`
	TotalCitations int64     `json:"total_citations"`
	StorageSize    int64     `json:"storage_size"`
	LastUpdated    time.Time `json:"last_updated"`
}
