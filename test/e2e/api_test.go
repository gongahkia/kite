package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// Base URL for the API server - set via environment variable or use default
	baseURL = "http://localhost:8080"
)

// TestHealthEndpoint verifies the health check endpoint returns 200 OK
func TestHealthEndpoint(t *testing.T) {
	resp, err := http.Get(baseURL + "/health")
	require.NoError(t, err, "Failed to call health endpoint")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Health check should return 200 OK")

	var health map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&health)
	require.NoError(t, err, "Failed to decode health response")

	assert.Equal(t, "healthy", health["status"], "Status should be healthy")
}

// TestReadinessEndpoint verifies the readiness check endpoint
func TestReadinessEndpoint(t *testing.T) {
	resp, err := http.Get(baseURL + "/ready")
	require.NoError(t, err, "Failed to call readiness endpoint")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Readiness check should return 200 OK")
}

// TestMetricsEndpoint verifies Prometheus metrics are exposed
func TestMetricsEndpoint(t *testing.T) {
	// Metrics are typically on a different port (9091)
	resp, err := http.Get("http://localhost:9091/metrics")
	require.NoError(t, err, "Failed to call metrics endpoint")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Metrics endpoint should return 200 OK")
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/plain", "Metrics should be in Prometheus format")
}

// TestListJurisdictions verifies the jurisdictions listing endpoint
func TestListJurisdictions(t *testing.T) {
	resp, err := http.Get(baseURL + "/api/v1/jurisdictions")
	require.NoError(t, err, "Failed to call jurisdictions endpoint")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Should return 200 OK")

	var jurisdictions []string
	err = json.NewDecoder(resp.Body).Decode(&jurisdictions)
	require.NoError(t, err, "Failed to decode jurisdictions response")

	assert.NotEmpty(t, jurisdictions, "Should return at least one jurisdiction")
}

// TestSearchCasesEndpoint verifies the case search functionality
func TestSearchCasesEndpoint(t *testing.T) {
	t.Skip("Requires populated database or test fixtures")

	resp, err := http.Get(baseURL + "/api/v1/cases?query=contract&jurisdiction=US_FEDERAL&limit=10")
	require.NoError(t, err, "Failed to call search endpoint")
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Search should return 200 OK")

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Failed to decode search response")

	cases, ok := result["cases"].([]interface{})
	require.True(t, ok, "Response should contain cases array")
	assert.LessOrEqual(t, len(cases), 10, "Should respect limit parameter")
}

// TestGetCaseByID verifies retrieving a specific case by ID
func TestGetCaseByID(t *testing.T) {
	t.Skip("Requires valid case ID in database")

	caseID := "test-case-id-123"
	resp, err := http.Get(baseURL + "/api/v1/cases/" + caseID)
	require.NoError(t, err, "Failed to call get case endpoint")
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		t.Skip("Case not found - test data needs to be set up")
	}

	assert.Equal(t, http.StatusOK, resp.StatusCode, "Should return 200 OK for valid case")

	var caseData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&caseData)
	require.NoError(t, err, "Failed to decode case response")

	assert.Equal(t, caseID, caseData["id"], "Case ID should match requested ID")
}

// TestSubmitScrapingJob verifies job submission endpoint
func TestSubmitScrapingJob(t *testing.T) {
	t.Skip("Requires authentication and job processing infrastructure")

	// This test would submit a job and verify it was created
	// Implementation depends on authentication mechanism
}

// TestJobStatusPolling verifies we can poll for job status
func TestJobStatusPolling(t *testing.T) {
	t.Skip("Requires valid job ID from submission")

	jobID := "test-job-id-456"
	
	// Poll for up to 30 seconds
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Job did not complete within timeout")
		case <-ticker.C:
			resp, err := http.Get(baseURL + "/api/v1/jobs/" + jobID)
			require.NoError(t, err)
			defer resp.Body.Close()

			var job map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&job)
			require.NoError(t, err)

			status := job["status"].(string)
			if status == "completed" || status == "failed" {
				t.Logf("Job finished with status: %s", status)
				return
			}
		}
	}
}

// TestCORSHeaders verifies CORS headers are set correctly
func TestCORSHeaders(t *testing.T) {
	req, err := http.NewRequest("OPTIONS", baseURL+"/api/v1/jurisdictions", nil)
	require.NoError(t, err)
	
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode, "OPTIONS request should return 204")
	assert.NotEmpty(t, resp.Header.Get("Access-Control-Allow-Origin"), "CORS headers should be present")
}

// TestRateLimiting verifies rate limiting is enforced
func TestRateLimiting(t *testing.T) {
	t.Skip("Rate limiting configuration may vary by environment")

	// Send many rapid requests to trigger rate limit
	const requestCount = 100
	statusCodes := make(map[int]int)

	for i := 0; i < requestCount; i++ {
		resp, err := http.Get(baseURL + "/api/v1/jurisdictions")
		if err != nil {
			continue
		}
		statusCodes[resp.StatusCode]++
		resp.Body.Close()
	}

	// Should see some 429 (Too Many Requests) responses
	assert.Greater(t, statusCodes[http.StatusTooManyRequests], 0, 
		"Rate limiting should trigger 429 responses")
}
