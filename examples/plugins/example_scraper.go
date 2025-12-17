package main

import (
	"context"
	"fmt"

	"github.com/gongahkia/kite/internal/plugins"
	"github.com/gongahkia/kite/pkg/models"
)

// ExampleScraper is an example scraper plugin
type ExampleScraper struct {
	name        string
	version     string
	description string
	initialized bool
	running     bool
}

// New creates a new instance of the plugin
// This function is required and will be called by the plugin loader
func New() plugins.Plugin {
	return &ExampleScraper{
		name:        "example-scraper",
		version:     "1.0.0",
		description: "Example scraper plugin for demonstration",
	}
}

// Name returns the plugin name
func (s *ExampleScraper) Name() string {
	return s.name
}

// Version returns the plugin version
func (s *ExampleScraper) Version() string {
	return s.version
}

// Description returns the plugin description
func (s *ExampleScraper) Description() string {
	return s.description
}

// Init initializes the plugin
func (s *ExampleScraper) Init(config map[string]interface{}) error {
	fmt.Printf("Initializing %s v%s\n", s.name, s.version)

	// Process configuration
	if apiKey, ok := config["api_key"].(string); ok {
		fmt.Printf("Using API key: %s\n", apiKey)
	}

	s.initialized = true
	return nil
}

// Start starts the plugin
func (s *ExampleScraper) Start(ctx context.Context) error {
	if !s.initialized {
		return fmt.Errorf("plugin not initialized")
	}

	fmt.Printf("Starting %s\n", s.name)
	s.running = true
	return nil
}

// Stop stops the plugin
func (s *ExampleScraper) Stop(ctx context.Context) error {
	fmt.Printf("Stopping %s\n", s.name)
	s.running = false
	return nil
}

// Health checks if the plugin is healthy
func (s *ExampleScraper) Health() error {
	if !s.running {
		return fmt.Errorf("plugin not running")
	}
	return nil
}

// SearchCases searches for cases
func (s *ExampleScraper) SearchCases(ctx context.Context, query plugins.SearchQuery) ([]*models.Case, error) {
	fmt.Printf("Searching for cases: %s\n", query.Query)

	// Example: return mock cases
	cases := []*models.Case{
		{
			ID:           "example/2023/001",
			CaseName:     "Example Case 1",
			CaseNumber:   "[2023] EX 1",
			Court:        "Example Court",
			Jurisdiction: "Example Jurisdiction",
			Summary:      "This is an example case from the example scraper plugin",
		},
		{
			ID:           "example/2023/002",
			CaseName:     "Example Case 2",
			CaseNumber:   "[2023] EX 2",
			Court:        "Example Court",
			Jurisdiction: "Example Jurisdiction",
			Summary:      "Another example case",
		},
	}

	return cases, nil
}

// GetCaseByID retrieves a case by ID
func (s *ExampleScraper) GetCaseByID(ctx context.Context, id string) (*models.Case, error) {
	fmt.Printf("Getting case by ID: %s\n", id)

	return &models.Case{
		ID:           id,
		CaseName:     "Example Case",
		CaseNumber:   "[2023] EX 1",
		Court:        "Example Court",
		Jurisdiction: "Example Jurisdiction",
	}, nil
}

// GetCasesByDateRange retrieves cases by date range
func (s *ExampleScraper) GetCasesByDateRange(ctx context.Context, start, end string) ([]*models.Case, error) {
	fmt.Printf("Getting cases from %s to %s\n", start, end)
	return []*models.Case{}, nil
}

// IsAvailable checks if the scraper is available
func (s *ExampleScraper) IsAvailable(ctx context.Context) bool {
	return s.running
}

// GetJurisdiction returns the jurisdiction
func (s *ExampleScraper) GetJurisdiction() string {
	return "Example Jurisdiction"
}

// main is required for building as a plugin
func main() {}
