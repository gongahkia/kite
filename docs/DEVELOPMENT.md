# Kite Developer Guide

Complete guide for developers contributing to or extending Kite v4.

## Table of Contents

- [Getting Started](#getting-started)
- [Architecture](#architecture)
- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Code Style](#code-style)
- [Testing](#testing)
- [Contributing](#contributing)
- [Plugin Development](#plugin-development)

## Getting Started

### Prerequisites

- **Go 1.22+** (with generics support)
- **Docker** and **Docker Compose** (for local services)
- **Make** (for build automation)
- **Git** (version control)

Optional:
- **PostgreSQL 15+** (or use Docker)
- **Redis 7+** (or use Docker)
- **NATS 2.9+** (or use Docker)

### Quick Start

```bash
# Clone repository
git clone https://github.com/gongahkia/kite.git
cd kite

# Install dependencies
go mod download

# Run tests
make test

# Build binaries
make build

# Run locally with Docker Compose
docker-compose up -d

# Run API server
./bin/kite-api serve --config configs/default.yaml

# Run workers
./bin/kite-worker start --workers 4
```

## Architecture

### High-Level Overview

```
┌─────────────────────────────────────────┐
│          API Gateway Layer               │
│  ┌──────┐ ┌──────┐ ┌────────┐           │
│  │ REST │ │ gRPC │ │WebSocket│           │
│  └──────┘ └──────┘ └────────┘           │
│         │        │         │             │
│    [Middleware Chain]                    │
│  Auth │ CORS │ RateLimit │ Logging      │
└─────────────┬───────────────────────────┘
              │
┌─────────────┴───────────────────────────┐
│          Service Layer                   │
│  ┌─────────┐ ┌─────────┐ ┌──────────┐  │
│  │Scraper  │ │ Search  │ │ Analysis │  │
│  │Service  │ │ Service │ │ Service  │  │
│  └─────────┘ └─────────┘ └──────────┘  │
│         │           │            │       │
│         └───────────┴────────────┘       │
│                     │                    │
│              [Event Bus]                 │
└─────────────────────┬───────────────────┘
                      │
┌─────────────────────┴───────────────────┐
│          Worker Layer                    │
│  ┌────────────────────────────────┐     │
│  │  Distributed Scraper Workers    │     │
│  │  [Worker Pool with Job Queue]   │     │
│  └────────────────────────────────┘     │
└─────────────────────┬───────────────────┘
                      │
┌─────────────────────┴───────────────────┐
│          Adapter Layer                   │
│  ┌─────────┐ ┌────────┐ ┌────────┐     │
│  │Storage  │ │ Cache  │ │ Queue  │     │
│  │Adapter  │ │Adapter │ │Adapter │     │
│  └─────────┘ └────────┘ └────────┘     │
└─────────────────────────────────────────┘
```

### Core Components

#### 1. API Layer (`internal/api/`)
- **Handlers**: HTTP request handlers
- **Middleware**: Request processing chain
- **Routes**: URL routing configuration
- **Validators**: Request validation

#### 2. Scraper Engine (`internal/scraper/`)
- **Base**: Common scraper interface
- **Jurisdictions**: Jurisdiction-specific scrapers
- **RateLimit**: Request throttling
- **Robots**: robots.txt compliance

#### 3. Worker System (`internal/worker/`)
- **Pool**: Goroutine worker pool
- **Queue**: Job queue abstraction
- **Job**: Job definitions and handlers

#### 4. Storage Layer (`internal/storage/`)
- **Interface**: Storage adapter interface
- **Implementations**: PostgreSQL, SQLite, MongoDB, In-Memory

#### 5. Search Engine (`internal/search/`)
- **Engine**: Full-text search
- **Query**: Query DSL
- **Suggestions**: Autocomplete and spell-check

## Development Setup

### Local Development with Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: kite_dev
      POSTGRES_USER: kite
      POSTGRES_PASSWORD: kite_dev_pass
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  nats:
    image: nats:2.9-alpine
    ports:
      - "4222:4222"
      - "8222:8222"

volumes:
  postgres_data:
```

```bash
# Start services
docker-compose up -d

# Run migrations
make migrate

# Run API server
make run-api

# Run workers
make run-worker
```

### Configuration

Configuration is managed via YAML files and environment variables:

```yaml
# configs/development.yaml
server:
  http_port: 8080
  grpc_port: 50051
  enable_grpc: true

database:
  driver: postgres
  url: postgres://kite:kite_dev_pass@localhost:5432/kite_dev?sslmode=disable

cache:
  driver: redis
  url: redis://localhost:6379

queue:
  driver: nats
  url: nats://localhost:4222

workers:
  count: 4
  queue_size: 1000

logging:
  level: debug
  format: json

metrics:
  enabled: true
  port: 9090
```

Environment variables override config:

```bash
export DATABASE_URL="postgres://..."
export REDIS_URL="redis://..."
export LOG_LEVEL="debug"
```

## Project Structure

```
kite/
├── cmd/                    # Application entry points
│   ├── kite-api/          # API server
│   ├── kite-worker/       # Worker process
│   └── kite-admin/        # Admin CLI
├── internal/              # Private application code
│   ├── api/               # HTTP API
│   │   ├── handlers/      # Request handlers
│   │   ├── middleware/    # Middleware
│   │   └── routes.go      # Route definitions
│   ├── grpc/              # gRPC API
│   ├── scraper/           # Scraping engine
│   │   ├── base.go        # Base scraper
│   │   └── jurisdictions/ # Jurisdiction scrapers
│   ├── worker/            # Worker pool
│   ├── queue/             # Job queue
│   ├── storage/           # Storage adapters
│   ├── cache/             # Caching layer
│   ├── search/            # Search engine
│   ├── citation/          # Citation analysis
│   ├── concepts/          # Legal concept extraction
│   ├── validation/        # Data validation
│   ├── jurisdiction/      # Jurisdiction metadata
│   ├── compliance/        # Ethical scraping
│   ├── batch/             # Batch processing
│   ├── export/            # Data export
│   ├── events/            # Event system
│   └── observability/     # Metrics, logging
├── pkg/                   # Public library code
│   ├── models/            # Data models
│   ├── client/            # Go client library
│   └── errors/            # Error types
├── api/                   # API definitions
│   ├── proto/             # Protocol Buffers
│   └── openapi.yaml       # OpenAPI spec
├── configs/               # Configuration files
├── deployment/            # Deployment manifests
│   ├── docker/
│   ├── k8s/
│   └── helm/
├── docs/                  # Documentation
├── test/                  # Tests
│   ├── integration/
│   └── e2e/
├── .github/               # GitHub workflows
├── Dockerfile
├── docker-compose.yaml
├── Makefile
└── go.mod
```

## Code Style

### Go Conventions

Follow standard Go conventions:

```go
// Package documentation
package scraper

// Exported types with documentation
type Scraper interface {
    // GetName returns the scraper name
    GetName() string

    // SearchCases searches for cases
    SearchCases(ctx context.Context, query SearchQuery) ([]*models.Case, error)
}

// Unexported types
type baseScraper struct {
    name string
    url  string
}

// Constructor functions
func NewBaseScraper(name, url string) *baseScraper {
    return &baseScraper{
        name: name,
        url:  url,
    }
}
```

### Error Handling

Use typed errors from `pkg/errors`:

```go
import "github.com/gongahkia/kite/pkg/errors"

func fetchCase(id string) (*models.Case, error) {
    if id == "" {
        return nil, errors.ValidationError("case ID is required")
    }

    c, err := storage.Get(id)
    if err != nil {
        return nil, errors.StorageError("failed to fetch case", err)
    }

    if c == nil {
        return nil, errors.ErrNotFound
    }

    return c, nil
}
```

### Context Usage

Always pass context as first parameter:

```go
func ProcessCase(ctx context.Context, c *models.Case) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }

    // Process with context
    result, err := enricher.Enrich(ctx, c)
    if err != nil {
        return err
    }

    return nil
}
```

### Concurrency

Use goroutines with proper error handling:

```go
func ProcessBatch(ctx context.Context, cases []*models.Case) error {
    g, ctx := errgroup.WithContext(ctx)

    for _, c := range cases {
        c := c // Capture loop variable
        g.Go(func() error {
            return ProcessCase(ctx, c)
        })
    }

    return g.Wait()
}
```

## Testing

### Unit Tests

```go
func TestScraperSearch(t *testing.T) {
    scraper := NewTestScraper()

    tests := []struct {
        name    string
        query   SearchQuery
        want    int
        wantErr bool
    }{
        {
            name:  "valid query",
            query: SearchQuery{Query: "test", Limit: 10},
            want:  5,
        },
        {
            name:    "empty query",
            query:   SearchQuery{},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cases, err := scraper.SearchCases(context.Background(), tt.query)

            if tt.wantErr {
                assert.Error(t, err)
                return
            }

            assert.NoError(t, err)
            assert.Len(t, cases, tt.want)
        })
    }
}
```

### Integration Tests

```go
// +build integration

func TestPostgresStorage(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()

    storage := NewPostgresStorage(db)

    // Test CRUD operations
    c := models.NewCase()
    c.CaseName = "Test Case"

    err := storage.Create(c)
    require.NoError(t, err)

    retrieved, err := storage.Get(c.ID)
    require.NoError(t, err)
    assert.Equal(t, c.CaseName, retrieved.CaseName)
}
```

### Running Tests

```bash
# Unit tests
make test

# Integration tests (requires Docker)
make test-integration

# Specific package
go test ./internal/scraper/...

# With coverage
go test -cover ./...

# Race detection
go test -race ./...

# Benchmark
go test -bench=. ./internal/search/
```

## Contributing

### Workflow

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** changes (`git commit -m 'feat: add amazing feature'`)
4. **Push** to branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Commit Messages

Follow Conventional Commits:

```
feat: add new jurisdiction scraper for Singapore
fix: resolve rate limiting issue in BAILII scraper
docs: update API documentation
test: add tests for citation extraction
refactor: simplify search query builder
perf: optimize database queries
chore: update dependencies
```

### Pull Request Checklist

- [ ] Tests pass (`make test`)
- [ ] Linter passes (`make lint`)
- [ ] Documentation updated
- [ ] Changelog updated (if applicable)
- [ ] Code reviewed
- [ ] No merge conflicts

## Plugin Development

### Creating a Custom Scraper

```go
package myscrapers

import (
    "context"
    "github.com/gongahkia/kite/internal/scraper"
    "github.com/gongahkia/kite/pkg/models"
)

type MyCustomScraper struct {
    *scraper.BaseScraper
}

func NewMyCustomScraper() *MyCustomScraper {
    return &MyCustomScraper{
        BaseScraper: scraper.NewBaseScraper(
            "MyCustomScraper",
            "My Jurisdiction",
            "https://example.com",
            30, // rate limit
        ),
    }
}

func (s *MyCustomScraper) SearchCases(ctx context.Context, query scraper.SearchQuery) ([]*models.Case, error) {
    // Implementation
    return nil, nil
}

// Implement other interface methods...
```

### Registering Plugin

```go
func init() {
    registry := scraper.NewScraperRegistry()
    registry.Register("my-custom", NewMyCustomScraper())
}
```

## Best Practices

1. **Use interfaces** for abstraction and testability
2. **Handle errors** explicitly, don't ignore them
3. **Use context** for cancellation and timeouts
4. **Write tests** for all public APIs
5. **Document** exported functions and types
6. **Follow Go idioms** (effective Go)
7. **Keep functions small** and focused
8. **Avoid premature optimization**
9. **Use structured logging** with context
10. **Profile** before optimizing

## Resources

- [Effective Go](https://golang.org/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Project Layout](https://github.com/golang-standards/project-layout)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)

## Getting Help

- **GitHub Issues**: Report bugs or request features
- **Discussions**: Ask questions and share ideas
- **Discord**: Join our community chat
- **Email**: dev@kite.example.com
