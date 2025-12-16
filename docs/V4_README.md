# Kite v4.0.0 - Golang Implementation

Welcome to **Kite v4**, a complete reimagining of the legal case law scraping platform as a production-ready, API-first backend service written in Go.

## What's New in v4?

Kite v4 represents a significant evolution from the Python library (v1/v2) to a distributed, scalable backend service:

### Architecture Highlights

- **API-First Design**: RESTful HTTP API with OpenAPI documentation
- **Distributed Architecture**: Job queue with worker pools for concurrent scraping
- **Production-Ready**: Comprehensive observability, graceful shutdown, health checks
- **Type-Safe**: Go's strong typing prevents runtime errors
- **High Performance**: Goroutines enable efficient concurrent operations
- **Extensible**: Middleware chains, storage adapters, queue adapters

### Core Features Implemented

✅ **Data Models**
- Case, Judge, Citation, Legal Concept models
- Comprehensive validation with go-playground/validator
- Quality scoring and completeness checking

✅ **Configuration Management**
- Viper-based configuration with environment variable support
- Multiple environments (dev, staging, prod)
- Sensible defaults

✅ **Observability**
- Structured logging with zerolog
- Prometheus metrics for all operations
- Request tracing with request IDs
- Performance profiling support

✅ **Scraping Infrastructure**
- Base scraper interface for jurisdiction-specific implementations
- Token bucket rate limiting per domain
- Robots.txt compliance with caching
- Retry logic with exponential backoff

✅ **Job Queue & Workers**
- Queue abstraction (currently in-memory, extensible to NATS/Redis)
- Worker pool with configurable concurrency
- Job prioritization and retry logic
- Dead letter queue for failed jobs

✅ **Storage Layer**
- Storage adapter interface (currently in-memory, extensible to PostgreSQL/MongoDB)
- CRUD operations for all models
- Search and filtering capabilities
- Deduplication support

✅ **HTTP API**
- Fiber-based REST API
- CORS, request logging, recovery middleware
- Metrics collection per endpoint
- JSON error responses
- Health and readiness checks

## Quick Start

### Prerequisites

- Go 1.22+ (install from https://go.dev/dl/)
- Docker & Docker Compose (optional, for containerized deployment)

### Running Locally

1. **Clone the repository**
   ```bash
   git clone https://github.com/gongahkia/kite
   cd kite
   ```

2. **Install dependencies**
   ```bash
   make install
   ```

3. **Run the API server**
   ```bash
   make run
   ```

4. **Access the API**
   - API: http://localhost:8080
   - Health: http://localhost:8080/health
   - Metrics: http://localhost:9091/metrics

### Using Docker

```bash
# Build and run with docker-compose
docker-compose up -d

# Or build and run manually
make docker-build
make docker-run
```

### With Monitoring Stack

```bash
# Run with Prometheus and Grafana
docker-compose --profile monitoring up -d

# Access Grafana at http://localhost:3000
# Default credentials: admin/admin
```

## API Endpoints

### Cases

- `GET /api/v1/cases` - List cases
- `GET /api/v1/cases/:id` - Get case by ID
- `POST /api/v1/cases` - Create case
- `PUT /api/v1/cases/:id` - Update case
- `DELETE /api/v1/cases/:id` - Delete case
- `POST /api/v1/cases/search` - Search cases

### Judges

- `GET /api/v1/judges` - List judges
- `GET /api/v1/judges/:id` - Get judge by ID
- `POST /api/v1/judges` - Create judge
- `PUT /api/v1/judges/:id` - Update judge

### Citations

- `GET /api/v1/citations` - List citations
- `GET /api/v1/citations/:id` - Get citation by ID
- `POST /api/v1/citations` - Create citation

### Stats

- `GET /api/v1/stats` - Get system statistics
- `GET /api/v1/stats/storage` - Get storage statistics

### Health

- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /metrics` - Prometheus metrics

## Configuration

Configuration can be provided via:
1. YAML files in `configs/` directory
2. Environment variables (prefixed with `KITE_`)

Example environment variables:
```bash
export KITE_SERVER_PORT=8080
export KITE_DATABASE_DRIVER=memory
export KITE_OBSERVABILITY_LOG_LEVEL=debug
export KITE_WORKER_COUNT=8
```

See `configs/default.yaml` for all available options.

## Project Structure

```
kite/
├── cmd/
│   ├── kite-api/          # API server entry point
│   ├── kite-worker/       # Worker process (future)
│   └── kite-admin/        # Admin CLI (future)
├── internal/              # Private application code
│   ├── api/               # HTTP API layer
│   ├── scraper/           # Scraping engine
│   ├── worker/            # Worker pool
│   ├── queue/             # Job queue
│   ├── storage/           # Storage adapters
│   ├── config/            # Configuration
│   └── observability/     # Logging & metrics
├── pkg/                   # Public library code
│   ├── models/            # Data models
│   ├── errors/            # Error types
│   └── validation/        # Validation logic
├── configs/               # Configuration files
├── deployment/            # Deployment manifests
│   ├── docker/
│   ├── k8s/
│   └── prometheus/
└── docs/                  # Documentation
```

## Development

### Building

```bash
make build          # Build binary
make test           # Run tests
make lint           # Run linters
make fmt            # Format code
```

### Hot Reload

```bash
# Install air for hot reload
go install github.com/air-verse/air@latest

# Start development server
make dev
```

## Roadmap

### Immediate Next Steps

- [ ] Implement jurisdiction-specific scrapers (CourtListener, CanLII, BAILII, etc.)
- [ ] Add citation extraction system
- [ ] Implement legal concept taxonomy and extraction
- [ ] Add PostgreSQL storage adapter
- [ ] Add Redis/NATS queue adapter
- [ ] Implement gRPC API
- [ ] Add authentication middleware (JWT, API keys)

### Phase 2

- [ ] WebSocket streaming API
- [ ] GraphQL API (optional)
- [ ] Batch processing and export
- [ ] Cache layer (Redis)
- [ ] Advanced search with full-text indexing

### Phase 3

- [ ] Kubernetes deployment manifests
- [ ] CI/CD pipeline
- [ ] Grafana dashboards
- [ ] Performance testing
- [ ] Load testing

## Comparison with v2 (Python)

| Feature | v2 (Python) | v4 (Go) |
|---------|-------------|---------|
| Architecture | Library | API Service |
| Concurrency | Threading/AsyncIO | Goroutines |
| Type Safety | Dynamic | Static |
| Deployment | pip install | Single binary |
| API | No | RESTful + future gRPC |
| Distributed | No | Yes (job queue) |
| Performance | Good | Excellent |
| Memory Usage | Higher | Lower |
| Startup Time | Slower | Fast (<1s) |

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

## License

[LICENSE](../LICENSE)

## Support

For issues and questions, please [open an issue](https://github.com/gongahkia/kite/issues).

---

**Note**: This is the initial v4.0.0 release with core infrastructure. Jurisdiction-specific scrapers and advanced features will be added incrementally.
