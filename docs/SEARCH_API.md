# Search & Query API Documentation

Kite provides a powerful Search & Query API with full-text search, faceted filtering, query suggestions, and relevance scoring.

## Quick Start

### Basic Search

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "constitutional rights",
    "limit": 10
  }'
```

### Advanced Search with Filters

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "due process",
    "query_type": "fulltext",
    "fields": ["case_name", "summary", "full_text"],
    "jurisdiction": "federal",
    "court_level": 3,
    "start_date": "2020-01-01T00:00:00Z",
    "end_date": "2024-12-31T23:59:59Z",
    "concepts": ["constitutional law"],
    "min_quality": 0.7,
    "sort_by": "relevance",
    "sort_desc": true,
    "limit": 20,
    "facets": ["jurisdiction", "court", "year", "concepts"]
  }'
```

## Search Endpoint

### POST /api/v1/search

Execute a search query with full-text search, filtering, and faceting.

**Request Body:**

```json
{
  "query": "search text",
  "query_type": "fulltext",
  "fields": ["case_name", "summary", "full_text"],
  "jurisdiction": "federal",
  "court": "Supreme Court",
  "court_level": 3,
  "start_date": "2020-01-01T00:00:00Z",
  "end_date": "2024-12-31T23:59:59Z",
  "judges": ["Justice Johnson", "Justice Williams"],
  "parties": ["Smith", "State"],
  "concepts": ["constitutional law", "due process"],
  "min_quality": 0.7,
  "sort_by": "relevance",
  "sort_desc": true,
  "limit": 20,
  "offset": 0,
  "facets": ["jurisdiction", "court", "year", "concepts"]
}
```

**Parameters:**

| Field | Type | Description | Default |
|-------|------|-------------|---------|
| `query` | string | Search query text | Required |
| `query_type` | string | Query type: `fulltext`, `exact`, `fuzzy`, `regex` | `fulltext` |
| `fields` | []string | Fields to search: `case_name`, `summary`, `full_text` | All fields |
| `jurisdiction` | string | Filter by jurisdiction | - |
| `court` | string | Filter by court name | - |
| `court_level` | int | Filter by court level (1=Trial, 2=Appellate, 3=Supreme) | - |
| `start_date` | string | Start date (RFC3339 format) | - |
| `end_date` | string | End date (RFC3339 format) | - |
| `judges` | []string | Filter by judges | - |
| `parties` | []string | Filter by parties | - |
| `concepts` | []string | Filter by legal concepts | - |
| `min_quality` | float | Minimum quality score (0-1) | - |
| `sort_by` | string | Sort field: `relevance`, `decision_date`, `quality_score` | `relevance` |
| `sort_desc` | bool | Sort descending | `true` |
| `limit` | int | Number of results (1-1000) | 10 |
| `offset` | int | Result offset for pagination | 0 |
| `facets` | []string | Facet fields: `jurisdiction`, `court`, `court_level`, `year`, `concepts` | - |

**Response:**

```json
{
  "results": [
    {
      "case": {
        "id": "case-123",
        "case_number": "20-1234",
        "case_name": "Smith v. State",
        "court": "Supreme Court",
        "jurisdiction": "federal",
        "decision_date": "2023-06-15T00:00:00Z",
        "summary": "Case about constitutional rights...",
        "legal_concepts": ["constitutional law", "due process"]
      },
      "score": 8.5,
      "highlights": [
        "...constitutional <em>rights</em> are protected...",
        "The court found that <em>due process</em> was violated..."
      ]
    }
  ],
  "total_hits": 42,
  "search_time_ms": 125.5,
  "facets": {
    "jurisdiction": [
      {"value": "federal", "count": 30},
      {"value": "state", "count": 12}
    ],
    "court": [
      {"value": "Supreme Court", "count": 25},
      {"value": "Court of Appeal", "count": 17}
    ],
    "year": [
      {"value": "2023", "count": 18},
      {"value": "2022", "count": 15},
      {"value": "2021", "count": 9}
    ],
    "concepts": [
      {"value": "constitutional law", "count": 28},
      {"value": "due process", "count": 14}
    ]
  }
}
```

## Query Types

### Full-Text Search (default)

Uses storage backend's full-text search capabilities (FTS5 for SQLite, text indexes for MongoDB).

```json
{
  "query": "negligence liability",
  "query_type": "fulltext"
}
```

**Features:**
- Stemming and relevance ranking
- Multi-field search
- Weighted scoring (case name > concepts > summary > full text)

### Exact Match

Matches exact phrases.

```json
{
  "query": "strict liability",
  "query_type": "exact"
}
```

### Fuzzy Search

Tolerates typos and variations.

```json
{
  "query": "neglegence",
  "query_type": "fuzzy"
}
```

### Regex Search

Powerful pattern matching.

```json
{
  "query": "\\d+ U\\.S\\. \\d+",
  "query_type": "regex"
}
```

## Relevance Scoring

Results are scored based on multiple factors:

| Factor | Weight | Description |
|--------|--------|-------------|
| Case Name Match | 3.0 | Query terms in case name |
| Summary Match | 2.0 | Query terms in summary |
| Full Text Match | 1.0 | Query terms in full text |
| Legal Concepts Match | 2.5 | Query terms in legal concepts |
| Quality Score | Multiplier | Boosts high-quality cases |

**Example Scoring:**
```
Query: "constitutional rights"
Case Name: "Smith v. State (Constitutional Rights)" → +3.0
Summary: "Case about constitutional rights..." → +2.0
Concepts: ["constitutional law"] → +2.5
Quality: 0.95 → ×1.95

Final Score: (3.0 + 2.0 + 2.5) × 1.95 = 14.625
```

## Faceted Search

Request facets to get aggregate counts by field.

**Available Facets:**
- `jurisdiction`: Group by jurisdiction
- `court`: Group by court
- `court_level`: Group by court level
- `year`: Group by decision year
- `concepts`: Group by legal concepts

**Example:**

```bash
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "negligence",
    "facets": ["jurisdiction", "year", "concepts"]
  }'
```

**Response:**

```json
{
  "facets": {
    "jurisdiction": [
      {"value": "federal", "count": 150},
      {"value": "california", "count": 89},
      {"value": "new_york", "count": 67}
    ],
    "year": [
      {"value": "2023", "count": 45},
      {"value": "2022", "count": 38},
      {"value": "2021", "count": 32}
    ],
    "concepts": [
      {"value": "tort law", "count": 180},
      {"value": "negligence", "count": 156},
      {"value": "damages", "count": 98}
    ]
  }
}
```

## Query Suggestions

### GET /api/v1/search/suggest

Get query suggestions based on partial input.

**Parameters:**
- `q` (required): Partial query text
- `limit` (optional): Number of suggestions (default: 10)

**Example:**

```bash
curl "http://localhost:8080/api/v1/search/suggest?q=constitu&limit=5"
```

**Response:**

```json
{
  "suggestions": [
    {
      "text": "constitutional law",
      "score": 0.95,
      "type": "concept"
    },
    {
      "text": "Smith v. State (Constitutional Rights)",
      "score": 0.87,
      "type": "case_name"
    },
    {
      "text": "Constitutional Court",
      "score": 0.72,
      "type": "court"
    },
    {
      "text": "Justice Constitutional",
      "score": 0.65,
      "type": "judge"
    }
  ]
}
```

**Suggestion Types:**
- `case_name`: Case names
- `judge`: Judge names
- `court`: Court names
- `concept`: Legal concepts

## Autocomplete

### GET /api/v1/search/autocomplete

Autocomplete query as user types (same as `/suggest`).

**Example:**

```bash
curl "http://localhost:8080/api/v1/search/autocomplete?q=neg"
```

```json
{
  "suggestions": [
    {"text": "negligence", "score": 0.95, "type": "concept"},
    {"text": "negligent homicide", "score": 0.82, "type": "concept"},
    {"text": "negotiation", "score": 0.65, "type": "concept"}
  ]
}
```

## Pagination

### Offset-Based Pagination

Use `limit` and `offset` for traditional pagination.

**Page 1:**
```json
{"limit": 20, "offset": 0}
```

**Page 2:**
```json
{"limit": 20, "offset": 20}
```

**Page 3:**
```json
{"limit": 20, "offset": 40}
```

### Cursor-Based Pagination (Planned)

For large result sets, cursor-based pagination is more efficient.

```json
{
  "limit": 20,
  "cursor": "eyJpZCI6ImNhc2UtMTIzIiwic2NvcmUiOjguNX0="
}
```

## Sorting

### Sort by Relevance (Default)

Sorts by calculated relevance score.

```json
{
  "sort_by": "relevance",
  "sort_desc": true
}
```

### Sort by Date

Most recent first:

```json
{
  "sort_by": "decision_date",
  "sort_desc": true
}
```

Oldest first:

```json
{
  "sort_by": "decision_date",
  "sort_desc": false
}
```

### Sort by Quality

Highest quality first:

```json
{
  "sort_by": "quality_score",
  "sort_desc": true
}
```

## Advanced Filtering

### Date Range Filtering

```json
{
  "start_date": "2020-01-01T00:00:00Z",
  "end_date": "2020-12-31T23:59:59Z"
}
```

### Multi-Field Filtering

Combine multiple filters for precise results:

```json
{
  "jurisdiction": "federal",
  "court_level": 3,
  "judges": ["Justice Johnson"],
  "concepts": ["constitutional law", "due process"],
  "min_quality": 0.8
}
```

### Boolean Logic (Implicit AND)

All filters are combined with AND logic:

```json
{
  "jurisdiction": "federal",
  "concepts": ["constitutional law", "due process"]
}
```

Returns cases that are:
- In federal jurisdiction AND
- Tagged with both "constitutional law" AND "due process"

## Search in Specific Fields

Restrict search to specific fields for better precision.

### Search Case Names Only

```json
{
  "query": "Smith v. State",
  "fields": ["case_name"]
}
```

### Search Summaries and Full Text

```json
{
  "query": "negligence liability",
  "fields": ["summary", "full_text"]
}
```

### Search All Fields (Default)

```json
{
  "query": "constitutional rights"
}
```

Searches: case_name, summary, full_text, legal_concepts

## Highlights

Results include highlighted snippets showing query matches.

**Input:**
```json
{"query": "due process"}
```

**Output:**
```json
{
  "highlights": [
    "The court found that <em>due process</em> was violated when...",
    "...constitutional protections including <em>due</em> <em>process</em> rights..."
  ]
}
```

**Features:**
- Context-aware snippets (200 characters)
- HTML `<em>` tags for highlighting
- Up to 5 highlights per result
- Ellipsis for truncated text

## Performance

### Response Times

| Operation | Typical Response Time | Notes |
|-----------|----------------------|-------|
| Simple Search | 10-50ms | Full-text search with 10 results |
| Advanced Search | 50-200ms | Multiple filters + facets |
| Faceted Search | 100-300ms | 5 facets across 10,000 cases |
| Suggestions | 5-20ms | Autocomplete query |

### Optimization Tips

1. **Limit Result Count**
   - Use smaller `limit` values (10-50)
   - Reduce facet counts

2. **Use Specific Fields**
   - Searching specific fields is faster
   - Avoid searching `full_text` when possible

3. **Filter Early**
   - Apply filters to reduce result set
   - Use jurisdiction and court filters

4. **Minimize Facets**
   - Request only needed facets
   - Facets add 20-50ms per facet field

## Query DSL Examples

### Constitutional Law Cases from Supreme Court

```json
{
  "query": "constitutional interpretation",
  "court_level": 3,
  "concepts": ["constitutional law"],
  "sort_by": "decision_date",
  "sort_desc": true,
  "limit": 25
}
```

### Recent Employment Law Cases

```json
{
  "query": "employment discrimination",
  "concepts": ["employment law"],
  "start_date": "2023-01-01T00:00:00Z",
  "min_quality": 0.7,
  "facets": ["court", "year"]
}
```

### High-Quality Tort Cases

```json
{
  "query": "negligence damages",
  "concepts": ["tort law"],
  "min_quality": 0.85,
  "sort_by": "quality_score",
  "sort_desc": true
}
```

### Cases by Specific Judge

```json
{
  "judges": ["Justice Johnson"],
  "sort_by": "decision_date",
  "sort_desc": true,
  "limit": 50
}
```

## Metrics

Search queries are instrumented with Prometheus metrics.

### Available Metrics

```
# Search queries
kite_search_queries_total{query_type="fulltext"} 1234

# Search duration
kite_search_duration_seconds{query_type="fulltext",quantile="0.5"} 0.025
kite_search_duration_seconds{query_type="fulltext",quantile="0.95"} 0.125

# Result counts
kite_search_results_count{query_type="fulltext",quantile="0.5"} 15
kite_search_results_count{query_type="fulltext",quantile="0.95"} 87
```

Access metrics at: `http://localhost:9091/metrics`

## Error Handling

### Invalid Query

```json
{
  "error": "query must have either text or ID filters"
}
```

### Limit Out of Range

```json
{
  "error": "limit must be between 1 and 1000"
}
```

### Search Failed

```json
{
  "error": "Search failed"
}
```

## Integration Examples

### cURL

```bash
# Basic search
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -d '{"query": "negligence", "limit": 10}'

# With auth
curl -X POST http://localhost:8080/api/v1/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"query": "constitutional rights", "limit": 20}'

# Autocomplete
curl "http://localhost:8080/api/v1/search/autocomplete?q=const"
```

### JavaScript (fetch)

```javascript
async function search(query) {
  const response = await fetch('http://localhost:8080/api/v1/search', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      query,
      limit: 20,
      facets: ['jurisdiction', 'year', 'concepts']
    })
  });

  const data = await response.json();
  return data;
}

// Usage
const results = await search('constitutional rights');
console.log(`Found ${results.total_hits} cases in ${results.search_time_ms}ms`);
```

### Python (requests)

```python
import requests

def search_cases(query, **filters):
    url = 'http://localhost:8080/api/v1/search'
    payload = {
        'query': query,
        'limit': 20,
        **filters
    }

    response = requests.post(url, json=payload)
    return response.json()

# Usage
results = search_cases(
    'due process',
    jurisdiction='federal',
    court_level=3,
    start_date='2020-01-01T00:00:00Z',
    facets=['jurisdiction', 'year']
)

print(f"Found {results['total_hits']} cases")
for result in results['results']:
    print(f"- {result['case']['case_name']} (score: {result['score']:.2f})")
```

## Search Features Roadmap

### Implemented
- ✅ Full-text search with FTS5/text indexes
- ✅ Multi-field search
- ✅ Advanced filtering (jurisdiction, court, date, judges, concepts)
- ✅ Relevance scoring
- ✅ Pagination (offset-based)
- ✅ Faceted search
- ✅ Query suggestions
- ✅ Autocomplete
- ✅ Highlights

### Planned
- ⏳ Query expansion with synonyms
- ⏳ Spell checking and correction
- ⏳ Search history tracking
- ⏳ Saved searches
- ⏳ Boolean operators (AND, OR, NOT)
- ⏳ Proximity search
- ⏳ More Like This
- ⏳ Export search results

## See Also

- [API Documentation](./API.md)
- [gRPC API Documentation](./GRPC_API.md)
- [Storage Adapters](../internal/storage/)
- [Prometheus Metrics](./OPERATIONS.md)
