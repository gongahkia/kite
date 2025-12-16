# Kite API Documentation

Comprehensive API documentation for Kite v4 Legal Case Law Platform.

## Table of Contents

- [Overview](#overview)
- [Authentication](#authentication)
- [Rate Limiting](#rate-limiting)
- [REST API](#rest-api)
- [gRPC API](#grpc-api)
- [Error Handling](#error-handling)
- [Pagination](#pagination)
- [Filtering & Searching](#filtering--searching)
- [Examples](#examples)

## Overview

Kite provides two primary API interfaces:

1. **REST API** - HTTP/JSON API for web clients and general integration
2. **gRPC API** - High-performance binary protocol for service-to-service communication

Base URLs:
- REST API: `https://api.kite.example.com/api/v1`
- gRPC API: `grpc://api.kite.example.com:50051`

## Authentication

Kite supports three authentication methods:

### 1. API Key Authentication

Include your API key in the `X-API-Key` header:

```bash
curl -H "X-API-Key: your_api_key_here" \
  https://api.kite.example.com/api/v1/cases
```

### 2. JWT Token Authentication

Obtain a JWT token via the login endpoint:

```bash
# Login
curl -X POST https://api.kite.example.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass"}'

# Response
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "expires_in": 3600
}

# Use token
curl -H "Authorization: Bearer eyJhbGc..." \
  https://api.kite.example.com/api/v1/cases
```

### 3. Role-Based Access Control (RBAC)

Users are assigned roles with specific permissions:

- **admin**: Full access to all resources
- **user**: Read/write access to cases and searches
- **readonly**: Read-only access

## Rate Limiting

Default rate limits per authentication method:

| Auth Method | Requests/Minute | Requests/Hour |
|-------------|----------------|---------------|
| API Key     | 60             | 3,000         |
| JWT Token   | 120            | 6,000         |
| Unauthenticated | 10         | 100           |

Rate limit headers are included in all responses:

```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1640000000
```

## REST API

### Cases

#### List Cases

```http
GET /api/v1/cases
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| limit | integer | Number of results (default: 50, max: 1000) |
| offset | integer | Pagination offset (default: 0) |
| jurisdiction | string | Filter by jurisdiction |
| court | string | Filter by court |
| start_date | string | Filter by decision date (ISO 8601) |
| end_date | string | Filter by decision date (ISO 8601) |

**Example:**

```bash
curl "https://api.kite.example.com/api/v1/cases?jurisdiction=Australia&limit=10"
```

**Response:**

```json
{
  "data": [
    {
      "id": "cth/HCA/2023/15",
      "case_name": "Smith v Jones",
      "case_number": "[2023] HCA 15",
      "court": "High Court of Australia",
      "jurisdiction": "Australia",
      "decision_date": "2023-06-15T00:00:00Z",
      "judges": ["Justice Smith", "Justice Brown"],
      "summary": "Case summary...",
      "url": "https://www.austlii.edu.au/cth/HCA/2023/15.html",
      "source_database": "AustLII",
      "legal_concepts": ["Contract Law", "Breach of Contract"]
    }
  ],
  "pagination": {
    "total": 150,
    "limit": 10,
    "offset": 0,
    "has_more": true
  }
}
```

#### Get Case by ID

```http
GET /api/v1/cases/{case_id}
```

**Example:**

```bash
curl "https://api.kite.example.com/api/v1/cases/cth%2FHCA%2F2023%2F15"
```

#### Create Case

```http
POST /api/v1/cases
```

**Request Body:**

```json
{
  "case_name": "Doe v Roe",
  "case_number": "[2023] NSWCA 45",
  "court": "NSW Court of Appeal",
  "jurisdiction": "Australia",
  "decision_date": "2023-07-20",
  "judges": ["Justice Williams"],
  "summary": "Appeal regarding contract interpretation",
  "full_text": "Full judgment text..."
}
```

#### Update Case

```http
PUT /api/v1/cases/{case_id}
```

#### Delete Case

```http
DELETE /api/v1/cases/{case_id}
```

### Search

#### Search Cases

```http
POST /api/v1/search
```

**Request Body:**

```json
{
  "query": "contract breach damages",
  "filters": {
    "jurisdiction": "Australia",
    "start_date": "2020-01-01",
    "end_date": "2023-12-31",
    "courts": ["High Court of Australia"],
    "legal_concepts": ["Contract Law"]
  },
  "facets": ["jurisdiction", "court", "year"],
  "limit": 20,
  "offset": 0
}
```

**Response:**

```json
{
  "results": [
    {
      "id": "cth/HCA/2023/15",
      "case_name": "Smith v Jones",
      "relevance_score": 0.95,
      "highlights": {
        "summary": "...breach of <em>contract</em> and resulting <em>damages</em>..."
      }
    }
  ],
  "facets": {
    "jurisdiction": {
      "Australia": 45,
      "United Kingdom": 12
    },
    "court": {
      "High Court of Australia": 15,
      "Federal Court of Australia": 30
    }
  },
  "total": 57,
  "took_ms": 42
}
```

#### Get Suggestions

```http
GET /api/v1/search/suggestions?q={query}
```

**Example:**

```bash
curl "https://api.kite.example.com/api/v1/search/suggestions?q=contract"
```

**Response:**

```json
{
  "suggestions": [
    {
      "text": "contract law",
      "type": "legal_concept",
      "count": 1234
    },
    {
      "text": "Smith v Contract Ltd",
      "type": "case_name",
      "count": 1
    }
  ]
}
```

### Citations

#### Get Citation Network

```http
GET /api/v1/citations/{case_id}/network
```

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| depth | integer | Network depth (default: 1, max: 3) |
| direction | string | "citing", "cited", or "both" (default: "both") |

**Response:**

```json
{
  "nodes": [
    {
      "id": "cth/HCA/2023/15",
      "case_name": "Smith v Jones",
      "influence_score": 0.85
    }
  ],
  "edges": [
    {
      "source": "cth/HCA/2023/15",
      "target": "cth/HCA/2020/5",
      "type": "cites"
    }
  ]
}
```

### Validation

#### Validate Case

```http
POST /api/v1/validation/case
```

**Request Body:**

```json
{
  "case_name": "Test Case",
  "case_number": "[2023] TEST 1",
  "court": "Test Court",
  "jurisdiction": "Test"
}
```

**Response:**

```json
{
  "valid": false,
  "errors": [
    {
      "stage": "structural",
      "field": "decision_date",
      "message": "decision_date is required"
    }
  ],
  "quality_score": 0.65,
  "completeness": 0.70
}
```

### Batch Operations

#### Create Batch Job

```http
POST /api/v1/batch/jobs
```

**Request Body:**

```json
{
  "type": "export",
  "input": {
    "format": "csv",
    "filters": {
      "jurisdiction": "Australia"
    }
  }
}
```

**Response:**

```json
{
  "id": "batch_1640000000000",
  "type": "export",
  "status": "pending",
  "created_at": "2023-12-16T10:00:00Z"
}
```

#### Get Batch Job Status

```http
GET /api/v1/batch/jobs/{job_id}
```

**Response:**

```json
{
  "id": "batch_1640000000000",
  "type": "export",
  "status": "completed",
  "progress": {
    "total": 1000,
    "processed": 1000,
    "succeeded": 998,
    "failed": 2,
    "percent": 100.0
  },
  "output": {
    "download_url": "https://api.kite.example.com/downloads/export_123.csv"
  },
  "created_at": "2023-12-16T10:00:00Z",
  "completed_at": "2023-12-16T10:05:32Z"
}
```

### Export

#### Export Cases

```http
POST /api/v1/export
```

**Request Body:**

```json
{
  "format": "json|jsonlines|csv|xml|bibtex|markdown|text",
  "filters": {
    "jurisdiction": "Australia",
    "start_date": "2023-01-01"
  },
  "options": {
    "compress": true,
    "include_full_text": false
  }
}
```

**Response:** File download with appropriate Content-Type

## gRPC API

See [GRPC_API.md](GRPC_API.md) for detailed gRPC documentation.

**Quick Example:**

```protobuf
service SearchService {
  rpc SearchCases(SearchRequest) returns (SearchResponse);
  rpc StreamCases(StreamRequest) returns (stream Case);
}
```

## Error Handling

All errors follow a consistent format:

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "Case not found",
    "details": {
      "case_id": "invalid_id"
    },
    "request_id": "req_abc123"
  }
}
```

**Common Error Codes:**

| Code | HTTP Status | Description |
|------|-------------|-------------|
| INVALID_REQUEST | 400 | Invalid request parameters |
| UNAUTHORIZED | 401 | Authentication required |
| FORBIDDEN | 403 | Insufficient permissions |
| NOT_FOUND | 404 | Resource not found |
| RATE_LIMITED | 429 | Rate limit exceeded |
| INTERNAL_ERROR | 500 | Internal server error |

## Pagination

Two pagination styles are supported:

### Offset-Based (REST)

```http
GET /api/v1/cases?limit=50&offset=100
```

### Cursor-Based (for large datasets)

```http
GET /api/v1/cases?limit=50&cursor=eyJpZCI6MTIzfQ==
```

**Response includes next cursor:**

```json
{
  "data": [...],
  "pagination": {
    "next_cursor": "eyJpZCI6MTczfQ==",
    "has_more": true
  }
}
```

## Filtering & Searching

### Field Filters

```json
{
  "filters": {
    "jurisdiction": "Australia",
    "court": "High Court of Australia",
    "court_level": "supreme",
    "start_date": "2020-01-01",
    "end_date": "2023-12-31",
    "judges": ["Justice Smith"],
    "legal_concepts": ["Contract Law"],
    "min_quality_score": 0.8
  }
}
```

### Full-Text Search

```json
{
  "query": "breach of contract",
  "fields": ["case_name", "summary", "full_text"],
  "operator": "AND"
}
```

### Advanced Search DSL

```json
{
  "query": {
    "bool": {
      "must": [
        {"match": {"full_text": "contract"}},
        {"term": {"jurisdiction": "Australia"}}
      ],
      "should": [
        {"match": {"legal_concepts": "Contract Law"}}
      ],
      "filter": [
        {"range": {"decision_date": {"gte": "2020-01-01"}}}
      ]
    }
  }
}
```

## Examples

### Complete Workflow

```bash
# 1. Authenticate
TOKEN=$(curl -X POST https://api.kite.example.com/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass"}' \
  | jq -r '.access_token')

# 2. Search for cases
curl -X POST https://api.kite.example.com/api/v1/search \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "contract breach",
    "filters": {"jurisdiction": "Australia"},
    "limit": 10
  }' | jq '.'

# 3. Get specific case
curl -H "Authorization: Bearer $TOKEN" \
  "https://api.kite.example.com/api/v1/cases/cth%2FHCA%2F2023%2F15" \
  | jq '.'

# 4. Get citation network
curl -H "Authorization: Bearer $TOKEN" \
  "https://api.kite.example.com/api/v1/citations/cth%2FHCA%2F2023%2F15/network?depth=2" \
  | jq '.'

# 5. Export results
curl -X POST https://api.kite.example.com/api/v1/export \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "format": "csv",
    "filters": {"jurisdiction": "Australia"}
  }' -o cases.csv
```

### Python Client Example

```python
import requests

class KiteClient:
    def __init__(self, api_key):
        self.base_url = "https://api.kite.example.com/api/v1"
        self.headers = {"X-API-Key": api_key}

    def search_cases(self, query, **filters):
        response = requests.post(
            f"{self.base_url}/search",
            headers=self.headers,
            json={"query": query, "filters": filters}
        )
        return response.json()

    def get_case(self, case_id):
        response = requests.get(
            f"{self.base_url}/cases/{case_id}",
            headers=self.headers
        )
        return response.json()

# Usage
client = KiteClient("your_api_key")
results = client.search_cases("contract law", jurisdiction="Australia")
```

### JavaScript/Node.js Example

```javascript
const axios = require('axios');

class KiteClient {
  constructor(apiKey) {
    this.baseURL = 'https://api.kite.example.com/api/v1';
    this.headers = { 'X-API-Key': apiKey };
  }

  async searchCases(query, filters = {}) {
    const response = await axios.post(
      `${this.baseURL}/search`,
      { query, filters },
      { headers: this.headers }
    );
    return response.data;
  }

  async getCase(caseId) {
    const response = await axios.get(
      `${this.baseURL}/cases/${encodeURIComponent(caseId)}`,
      { headers: this.headers }
    );
    return response.data;
  }
}

// Usage
const client = new KiteClient('your_api_key');
const results = await client.searchCases('contract law', {
  jurisdiction: 'Australia'
});
```

## Webhooks

Subscribe to events via webhooks:

```http
POST /api/v1/webhooks
```

**Request:**

```json
{
  "url": "https://your-app.com/webhook",
  "events": ["case.created", "case.updated", "scrape.completed"],
  "secret": "your_webhook_secret"
}
```

**Webhook Payload:**

```json
{
  "event": "case.created",
  "timestamp": "2023-12-16T10:00:00Z",
  "data": {
    "case_id": "cth/HCA/2023/15",
    "case_name": "Smith v Jones"
  },
  "signature": "sha256=abc123..."
}
```

## SDK Libraries

Official SDKs:
- **Go**: `go get github.com/gongahkia/kite/pkg/client`
- **Python**: `pip install kite-legal`
- **JavaScript**: `npm install @kite/client`
- **Java**: Maven/Gradle via Maven Central

## Support

- Documentation: https://docs.kite.example.com
- API Status: https://status.kite.example.com
- GitHub Issues: https://github.com/gongahkia/kite/issues
- Email: support@kite.example.com
