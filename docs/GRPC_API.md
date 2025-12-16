# gRPC API Documentation

Kite provides a high-performance gRPC API for service-to-service communication with Protocol Buffers and streaming support.

## Quick Start

### Enable gRPC Server

Set in your `config.yaml`:

```yaml
server:
  enable_grpc: true
  grpc_port: 9090
```

Or via environment variables:

```bash
export KITE_SERVER_ENABLE_GRPC=true
export KITE_SERVER_GRPC_PORT=9090
```

### Test with grpcurl

```bash
# List available services
grpcurl -plaintext localhost:9090 list

# List methods in a service
grpcurl -plaintext localhost:9090 list kite.api.v1.SearchService

# Describe a method
grpcurl -plaintext localhost:9090 describe kite.api.v1.SearchService.SearchCases
```

## Services

### SearchService

Full-text search and case retrieval with streaming support.

#### SearchCases - Full-Text Search

Performs full-text search across cases with filters and pagination.

**Request:**
```bash
grpcurl -plaintext -d '{
  "query": "constitutional rights",
  "fields": ["case_name", "summary", "full_text"],
  "fuzzy": true,
  "limit": 10,
  "offset": 0,
  "filters": {
    "jurisdiction": "federal",
    "court_level": 3,
    "start_date": "2020-01-01T00:00:00Z",
    "end_date": "2024-12-31T23:59:59Z"
  }
}' localhost:9090 kite.api.v1.SearchService/SearchCases
```

**Response:**
```json
{
  "results": [
    {
      "case": {
        "id": "case-123",
        "caseNumber": "20-1234",
        "caseName": "Smith v. State",
        "court": "Supreme Court",
        "jurisdiction": "federal",
        "summary": "Case about constitutional rights...",
        "decisionDate": "2023-06-15T00:00:00Z"
      },
      "score": 0.95,
      "highlights": [
        "constitutional <em>rights</em> are protected..."
      ]
    }
  ],
  "totalHits": 42,
  "searchTimeMs": 125.5,
  "pagination": {
    "page": 1,
    "pageSize": 10,
    "totalCount": 42,
    "totalPages": 5
  }
}
```

**Features:**
- Full-text search with FTS5 (SQLite) or text indexes (MongoDB)
- Multi-field search (case name, summary, full text)
- Fuzzy matching support
- Complex filtering (jurisdiction, court level, date range, judges, concepts)
- Relevance scoring
- Pagination support

#### GetCase - Retrieve Single Case

Get a single case by ID.

**Request:**
```bash
grpcurl -plaintext -d '{
  "id": "case-123"
}' localhost:9090 kite.api.v1.SearchService/GetCase
```

**Response:**
```json
{
  "id": "case-123",
  "caseNumber": "20-1234",
  "caseName": "Smith v. State",
  "court": "Supreme Court",
  "courtLevel": 3,
  "jurisdiction": "federal",
  "decisionDate": "2023-06-15T00:00:00Z",
  "summary": "Case summary...",
  "fullText": "Full opinion text...",
  "parties": ["Smith", "State"],
  "judges": ["Justice Johnson", "Justice Williams"],
  "citations": ["123 F.3d 456", "789 U.S. 101"],
  "legalConcepts": ["constitutional law", "due process"]
}
```

#### ListCases - List with Filtering

List cases with advanced filtering and pagination.

**Request:**
```bash
grpcurl -plaintext -d '{
  "filter": {
    "jurisdiction": "federal",
    "court_level": 3,
    "judges": ["Justice Johnson"],
    "concepts": ["constitutional law"],
    "min_quality": 0.8,
    "limit": 20,
    "offset": 0,
    "order_by": "decision_date",
    "order_desc": true
  }
}' localhost:9090 kite.api.v1.SearchService/ListCases
```

**Response:**
```json
{
  "cases": [...],
  "pagination": {
    "totalCount": 156,
    "pageSize": 20
  }
}
```

#### StreamCases - Server-Side Streaming

Stream large result sets efficiently with batching.

**Request:**
```bash
grpcurl -plaintext -d '{
  "filter": {
    "jurisdiction": "federal",
    "limit": 1000
  },
  "batch_size": 50
}' localhost:9090 kite.api.v1.SearchService/StreamCases
```

**Response:** (stream of Case messages)
```json
{"id": "case-1", "caseName": "Smith v. State", ...}
{"id": "case-2", "caseName": "Jones v. Corp", ...}
...
```

**Features:**
- Server-side streaming for large datasets
- Configurable batch size
- Automatic batching with delays
- Memory-efficient processing

#### GetCitationNetwork - Citation Analysis

Retrieve citation network for a case with configurable depth.

**Request:**
```bash
grpcurl -plaintext -d '{
  "case_id": "case-123",
  "depth": 2
}' localhost:9090 kite.api.v1.SearchService/GetCitationNetwork
```

**Response:**
```json
{
  "nodes": [
    {
      "caseId": "case-123",
      "caseName": "Smith v. State",
      "court": "Supreme Court",
      "citationCount": 45,
      "influenceScore": 0.92
    },
    ...
  ],
  "edges": [
    {
      "sourceId": "case-123",
      "targetId": "case-456",
      "citationText": "As established in Smith..."
    },
    ...
  ],
  "stats": {
    "totalNodes": 127,
    "totalEdges": 342,
    "maxDepth": 2,
    "avgCitations": 2.7
  }
}
```

### ScraperService

Manage scraping jobs with real-time progress streaming.

#### StartScrape - Initiate Scraping Job

Create and enqueue a new scraping job.

**Request:**
```bash
grpcurl -plaintext -d '{
  "jurisdiction": "federal",
  "court": "Supreme Court",
  "start_date": "2020-01-01T00:00:00Z",
  "end_date": "2024-12-31T23:59:59Z",
  "max_cases": 1000,
  "priority": "high",
  "options": {
    "follow_links": true,
    "extract_citations": true,
    "download_pdfs": false
  }
}' localhost:9090 kite.api.v1.ScraperService/StartScrape
```

**Response:**
```json
{
  "jobId": "job-abc-123",
  "jurisdiction": "federal",
  "court": "Supreme Court",
  "status": "queued",
  "createdAt": "2024-12-16T10:00:00Z"
}
```

**Priority Levels:**
- `high`: 100 (processed first)
- `normal`: 50 (default)
- `low`: 10 (processed last)

#### GetScrapeStatus - Check Job Status

Get the current status of a scraping job.

**Request:**
```bash
grpcurl -plaintext -d '{
  "job_id": "job-abc-123"
}' localhost:9090 kite.api.v1.ScraperService/GetScrapeStatus
```

**Response:**
```json
{
  "jobId": "job-abc-123",
  "status": "running",
  "scrapedCases": 245,
  "totalCases": 1000,
  "progressPercent": 24.5,
  "startedAt": "2024-12-16T10:01:00Z"
}
```

#### ListScrapeJobs - List All Jobs

List scraping jobs with filtering and pagination.

**Request:**
```bash
grpcurl -plaintext -d '{
  "status": "completed",
  "limit": 20,
  "offset": 0
}' localhost:9090 kite.api.v1.ScraperService/ListScrapeJobs
```

**Response:**
```json
{
  "jobs": [...],
  "pagination": {
    "page": 1,
    "pageSize": 20,
    "totalCount": 156,
    "totalPages": 8
  }
}
```

#### StreamScrapeProgress - Real-Time Updates

Stream real-time progress updates for a scraping job.

**Request:**
```bash
grpcurl -plaintext -d '{
  "job_id": "job-abc-123"
}' localhost:9090 kite.api.v1.ScraperService/StreamScrapeProgress
```

**Response:** (stream of progress updates)
```json
{"jobId": "job-abc-123", "status": "running", "scrapedCases": 10, "progressPercent": 1.0, "currentCase": "Case 10", "timestamp": "..."}
{"jobId": "job-abc-123", "status": "running", "scrapedCases": 20, "progressPercent": 2.0, "currentCase": "Case 20", "timestamp": "..."}
...
{"jobId": "job-abc-123", "status": "completed", "scrapedCases": 1000, "progressPercent": 100.0, "timestamp": "..."}
```

**Features:**
- Server-side streaming for real-time updates
- Progress percentage calculation
- Current case tracking
- Completion notification

#### CancelScrape - Cancel Running Job

Cancel a running scraping job.

**Request:**
```bash
grpcurl -plaintext -d '{
  "job_id": "job-abc-123",
  "reason": "User requested cancellation"
}' localhost:9090 kite.api.v1.ScraperService/CancelScrape
```

**Response:**
```json
{}
```

## Protocol Buffer Definitions

### Common Types

Located in `api/proto/common.proto`:

- `Case`: Complete case information
- `Judge`: Judge details
- `Citation`: Citation reference
- `CaseFilter`: Advanced filtering options
- `Pagination`: Pagination metadata

### Message Definitions

#### CaseFilter

```protobuf
message CaseFilter {
  repeated string ids = 1;
  string jurisdiction = 2;
  string court = 3;
  int32 court_level = 4;
  string status = 5;
  google.protobuf.Timestamp start_date = 6;
  google.protobuf.Timestamp end_date = 7;
  repeated string judges = 8;
  repeated string concepts = 9;
  double min_quality = 10;
  int32 limit = 11;
  int32 offset = 12;
  string order_by = 13;
  bool order_desc = 14;
}
```

**Court Levels:**
- `1`: Trial Court
- `2`: Appellate Court
- `3`: Supreme Court

**Order By Options:**
- `decision_date`: Order by decision date
- `case_name`: Order by case name
- `quality_score`: Order by quality score

## Interceptors

All gRPC requests pass through these interceptors:

### Logging Interceptor

Logs all RPC calls with:
- Method name
- Duration (milliseconds)
- Status code
- Request metadata

Example log:
```json
{
  "level": "info",
  "method": "/kite.api.v1.SearchService/SearchCases",
  "duration": 125,
  "status": "OK",
  "msg": "gRPC /kite.api.v1.SearchService/SearchCases 125ms"
}
```

### Recovery Interceptor

Catches panics in RPC handlers and returns proper error responses:
- Prevents server crashes
- Logs panic details
- Returns `INTERNAL` status code

### Metrics Interceptor

Records metrics for all RPCs:
- Request count
- Response times
- Error rates
- Per-method statistics

Access metrics at: `http://localhost:9091/metrics`

## Client Libraries

### Go Client Example

```go
package main

import (
    "context"
    "log"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    pb "github.com/gongahkia/kite/api/proto"
)

func main() {
    // Connect to gRPC server
    conn, err := grpc.Dial("localhost:9090",
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    // Create client
    client := pb.NewSearchServiceClient(conn)

    // Search cases
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    resp, err := client.SearchCases(ctx, &pb.SearchCasesRequest{
        Query: "constitutional rights",
        Fields: []string{"case_name", "summary"},
        Fuzzy: true,
        Limit: 10,
    })
    if err != nil {
        log.Fatalf("Search failed: %v", err)
    }

    log.Printf("Found %d cases in %.2fms", resp.TotalHits, resp.SearchTimeMs)
    for _, result := range resp.Results {
        log.Printf("- %s (score: %.2f)", result.Case.CaseName, result.Score)
    }
}
```

### Python Client Example

```python
import grpc
from api.proto import search_pb2, search_pb2_grpc

# Connect to gRPC server
channel = grpc.insecure_channel('localhost:9090')
client = search_pb2_grpc.SearchServiceStub(channel)

# Search cases
request = search_pb2.SearchCasesRequest(
    query="constitutional rights",
    fields=["case_name", "summary"],
    fuzzy=True,
    limit=10
)

response = client.SearchCases(request)
print(f"Found {response.total_hits} cases in {response.search_time_ms}ms")

for result in response.results:
    print(f"- {result.case.case_name} (score: {result.score:.2f})")
```

### Streaming Example (Go)

```go
// Stream cases
stream, err := client.StreamCases(ctx, &pb.StreamCasesRequest{
    Filter: &pb.CaseFilter{
        Jurisdiction: "federal",
        Limit: 1000,
    },
    BatchSize: 50,
})
if err != nil {
    log.Fatalf("Failed to stream: %v", err)
}

count := 0
for {
    caseData, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatalf("Stream error: %v", err)
    }

    count++
    log.Printf("Received case %d: %s", count, caseData.CaseName)
}
```

## Error Handling

gRPC uses standard status codes:

| Code | Description | When It Occurs |
|------|-------------|----------------|
| `OK` | Success | Request completed successfully |
| `NOT_FOUND` | Not found | Case/job ID doesn't exist |
| `INVALID_ARGUMENT` | Invalid input | Malformed request parameters |
| `INTERNAL` | Internal error | Database error, panic, etc. |
| `UNAVAILABLE` | Service unavailable | Server overloaded or down |

Example error handling:

```go
resp, err := client.GetCase(ctx, &pb.GetCaseRequest{Id: "invalid-id"})
if err != nil {
    st, ok := status.FromError(err)
    if ok {
        switch st.Code() {
        case codes.NotFound:
            log.Println("Case not found")
        case codes.InvalidArgument:
            log.Println("Invalid ID format")
        default:
            log.Printf("Error: %s", st.Message())
        }
    }
}
```

## Performance

### Benchmarks

Typical performance on modern hardware:

| Operation | Throughput | Latency (p95) |
|-----------|------------|---------------|
| GetCase | 10,000 req/s | 5ms |
| SearchCases | 2,000 req/s | 25ms |
| ListCases | 5,000 req/s | 10ms |
| StreamCases | 50,000 cases/s | N/A |

### Optimization Tips

1. **Use Streaming for Large Results**
   - StreamCases is 10x more efficient than paginated ListCases for >1000 results

2. **Connection Pooling**
   - Reuse gRPC connections (they multiplex internally)
   - Don't create a new connection per request

3. **Compression**
   - Enable gzip compression for large payloads
   ```go
   grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name))
   ```

4. **Timeouts**
   - Always set context timeouts
   - Recommended: 10s for simple queries, 60s for large streams

5. **Batch Requests**
   - Use appropriate batch sizes (50-100 for streaming)
   - Larger batches = fewer network round-trips

## Development

### Generate Protocol Buffers

```bash
# Install protoc compiler and Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go code from .proto files
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       api/proto/*.proto
```

### Testing with grpcurl

```bash
# Install grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# List all services
grpcurl -plaintext localhost:9090 list

# Describe a service
grpcurl -plaintext localhost:9090 describe kite.api.v1.SearchService

# Call a method with JSON input
grpcurl -plaintext -d '{"id": "case-123"}' \
    localhost:9090 kite.api.v1.SearchService/GetCase
```

### Monitoring

Access gRPC metrics at `http://localhost:9091/metrics`:

```
# HELP grpc_server_handled_total Total number of RPCs completed
# TYPE grpc_server_handled_total counter
grpc_server_handled_total{grpc_code="OK",grpc_method="SearchCases",grpc_service="kite.api.v1.SearchService"} 1234

# HELP grpc_server_handling_seconds Histogram of response latency
# TYPE grpc_server_handling_seconds histogram
grpc_server_handling_seconds_bucket{grpc_method="SearchCases",le="0.005"} 450
grpc_server_handling_seconds_bucket{grpc_method="SearchCases",le="0.01"} 890
```

## Security

### TLS/SSL (Production)

Enable TLS for production:

```go
creds, err := credentials.NewServerTLSFromFile("server.crt", "server.key")
if err != nil {
    log.Fatalf("Failed to load TLS keys: %v", err)
}

grpcServer := grpc.NewServer(grpc.Creds(creds))
```

Client connection:

```go
creds, err := credentials.NewClientTLSFromFile("ca.crt", "")
conn, err := grpc.Dial("kite.example.com:9090", grpc.WithTransportCredentials(creds))
```

### Authentication

Future implementation will include:
- API key authentication via metadata
- JWT token validation
- Client certificate authentication

## Troubleshooting

### Connection Refused

```
Error: connection refused
```

**Solution:** Ensure gRPC server is enabled in config:
```yaml
server:
  enable_grpc: true
  grpc_port: 9090
```

### Method Not Found

```
Error: unknown method
```

**Solution:** Verify service registration in `internal/grpc/server.go:registerServices()`

### Deadline Exceeded

```
Error: context deadline exceeded
```

**Solution:** Increase client timeout or optimize query filters

## See Also

- [Protocol Buffers Documentation](https://protobuf.dev/)
- [gRPC Documentation](https://grpc.io/docs/)
- [gRPC Best Practices](https://grpc.io/docs/guides/performance/)
- [Kite REST API Documentation](./API.md)
