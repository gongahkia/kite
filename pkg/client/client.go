// Package client provides a Go client library for the Kite Legal Case Law API
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gongahkia/kite/pkg/models"
)

// Client represents a Kite API client
type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
	userAgent  string
}

// Config holds client configuration
type Config struct {
	BaseURL    string
	APIKey     string
	Timeout    time.Duration
	UserAgent  string
}

// NewClient creates a new Kite API client with default settings
func NewClient(baseURL, apiKey string) *Client {
	return NewClientWithConfig(Config{
		BaseURL:   baseURL,
		APIKey:    apiKey,
		Timeout:   30 * time.Second,
		UserAgent: "kite-go-client/4.0.0",
	})
}

// NewClientWithConfig creates a new client with custom configuration
func NewClientWithConfig(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.UserAgent == "" {
		cfg.UserAgent = "kite-go-client/4.0.0"
	}

	return &Client{
		baseURL:   cfg.BaseURL,
		apiKey:    cfg.APIKey,
		userAgent: cfg.UserAgent,
		httpClient: &http.Client{
			Timeout: cfg.Timeout,
		},
	}
}

// SearchCasesParams holds parameters for case search
type SearchCasesParams struct {
	Query        string
	Jurisdiction string
	Court        string
	Limit        int
	Offset       int
}

// SearchCases searches for cases matching the given parameters
func (c *Client) SearchCases(ctx context.Context, params SearchCasesParams) ([]models.Case, error) {
	endpoint := "/api/v1/cases"
	
	// Build query parameters
	queryParams := url.Values{}
	queryParams.Set("query", params.Query)
	if params.Jurisdiction != "" {
		queryParams.Set("jurisdiction", params.Jurisdiction)
	}
	if params.Court != "" {
		queryParams.Set("court", params.Court)
	}
	if params.Limit > 0 {
		queryParams.Set("limit", fmt.Sprintf("%d", params.Limit))
	}
	if params.Offset > 0 {
		queryParams.Set("offset", fmt.Sprintf("%d", params.Offset))
	}

	fullURL := fmt.Sprintf("%s%s?%s", c.baseURL, endpoint, queryParams.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var result struct {
		Cases []models.Case `json:"cases"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Cases, nil
}

// GetCaseByID retrieves a specific case by its ID
func (c *Client) GetCaseByID(ctx context.Context, id string) (*models.Case, error) {
	endpoint := fmt.Sprintf("/api/v1/cases/%s", id)
	fullURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("case not found: %s", id)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var caseData models.Case
	if err := json.NewDecoder(resp.Body).Decode(&caseData); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &caseData, nil
}

// SubmitJobRequest holds parameters for submitting a scraping job
type SubmitJobRequest struct {
	Jurisdiction string `json:"jurisdiction"`
	Query        string `json:"query,omitempty"`
	Priority     string `json:"priority,omitempty"`
}

// JobResponse represents a job submission response
type JobResponse struct {
	JobID     string    `json:"job_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// SubmitScrapingJob submits a new scraping job to the API
func (c *Client) SubmitScrapingJob(ctx context.Context, req SubmitJobRequest) (*JobResponse, error) {
	endpoint := "/api/v1/jobs"
	fullURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, fullURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var jobResp JobResponse
	if err := json.NewDecoder(resp.Body).Decode(&jobResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &jobResp, nil
}

// GetJobStatus retrieves the status of a scraping job
func (c *Client) GetJobStatus(ctx context.Context, jobID string) (*JobResponse, error) {
	endpoint := fmt.Sprintf("/api/v1/jobs/%s", jobID)
	fullURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var jobResp JobResponse
	if err := json.NewDecoder(resp.Body).Decode(&jobResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &jobResp, nil
}

// ListJurisdictions retrieves the list of supported jurisdictions
func (c *Client) ListJurisdictions(ctx context.Context) ([]string, error) {
	endpoint := "/api/v1/jurisdictions"
	fullURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
	}

	var jurisdictions []string
	if err := json.NewDecoder(resp.Body).Decode(&jurisdictions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return jurisdictions, nil
}

// HealthCheck checks if the API is healthy
func (c *Client) HealthCheck(ctx context.Context) error {
	endpoint := "/health"
	fullURL := fmt.Sprintf("%s%s", c.baseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

// setHeaders sets common headers for all requests
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("User-Agent", c.userAgent)
	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	}
}

// handleErrorResponse processes error responses from the API
func (c *Client) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var errResp struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(body, &errResp); err != nil {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	if errResp.Message != "" {
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Message)
	}

	if errResp.Error != "" {
		return fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Error)
	}

	return fmt.Errorf("request failed with status %d", resp.StatusCode)
}
