# Kite Plugin Development Guide

Create custom plugins to extend Kite's functionality.

## Table of Contents

- [Overview](#overview)
- [Plugin Types](#plugin-types)
- [Creating a Plugin](#creating-a-plugin)
- [Building Plugins](#building-plugins)
- [Installing Plugins](#installing-plugins)
- [Plugin Lifecycle](#plugin-lifecycle)
- [Examples](#examples)
- [Best Practices](#best-practices)

## Overview

Kite's plugin system allows you to extend the platform with custom functionality:

- **Scraper Plugins**: Add support for new legal databases
- **Processor Plugins**: Process and transform case data
- **Validator Plugins**: Add custom validation rules
- **Exporter Plugins**: Export data in custom formats
- **Middleware Plugins**: Add custom HTTP middleware

Plugins are Go shared libraries (`.so` files) loaded at runtime.

## Plugin Types

### 1. Scraper Plugin

Add support for new legal databases or jurisdictions.

```go
type ScraperPlugin interface {
    Plugin
    SearchCases(ctx context.Context, query SearchQuery) ([]*models.Case, error)
    GetCaseByID(ctx context.Context, id string) (*models.Case, error)
    GetCasesByDateRange(ctx context.Context, start, end string) ([]*models.Case, error)
    IsAvailable(ctx context.Context) bool
    GetJurisdiction() string
}
```

### 2. Processor Plugin

Transform or enrich case data.

```go
type ProcessorPlugin interface {
    Plugin
    Process(ctx context.Context, c *models.Case) (*models.Case, error)
    CanProcess(c *models.Case) bool
    Priority() int
}
```

### 3. Validator Plugin

Add custom validation logic.

```go
type ValidatorPlugin interface {
    Plugin
    Validate(ctx context.Context, c *models.Case) ([]ValidationError, error)
    Severity() ValidationSeverity
}
```

### 4. Exporter Plugin

Export data in custom formats.

```go
type ExporterPlugin interface {
    Plugin
    Export(ctx context.Context, cases []*models.Case) ([]byte, error)
    FileExtension() string
    MIMEType() string
}
```

### 5. Middleware Plugin

Add custom HTTP middleware.

```go
type MiddlewarePlugin interface {
    Plugin
    Handler() func(next func(c interface{}) error) func(c interface{}) error
    Order() int
}
```

## Creating a Plugin

### Step 1: Implement the Plugin Interface

All plugins must implement the base `Plugin` interface:

```go
type Plugin interface {
    Name() string
    Version() string
    Description() string
    Init(config map[string]interface{}) error
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() error
}
```

### Step 2: Implement Type-Specific Interface

Implement the interface for your plugin type (e.g., `ScraperPlugin`).

### Step 3: Export a New Function

Your plugin must export a `New()` function:

```go
func New() plugins.Plugin {
    return &MyPlugin{}
}
```

### Example: Simple Scraper Plugin

```go
package main

import (
    "context"
    "github.com/gongahkia/kite/internal/plugins"
    "github.com/gongahkia/kite/pkg/models"
)

type MyScraper struct {
    name    string
    version string
    running bool
}

func New() plugins.Plugin {
    return &MyScraper{
        name:    "my-scraper",
        version: "1.0.0",
    }
}

func (s *MyScraper) Name() string {
    return s.name
}

func (s *MyScraper) Version() string {
    return s.version
}

func (s *MyScraper) Description() string {
    return "My custom scraper"
}

func (s *MyScraper) Init(config map[string]interface{}) error {
    // Initialize with config
    return nil
}

func (s *MyScraper) Start(ctx context.Context) error {
    s.running = true
    return nil
}

func (s *MyScraper) Stop(ctx context.Context) error {
    s.running = false
    return nil
}

func (s *MyScraper) Health() error {
    if !s.running {
        return fmt.Errorf("not running")
    }
    return nil
}

func (s *MyScraper) SearchCases(ctx context.Context, query plugins.SearchQuery) ([]*models.Case, error) {
    // Implement search logic
    return []*models.Case{}, nil
}

func (s *MyScraper) GetCaseByID(ctx context.Context, id string) (*models.Case, error) {
    // Implement retrieval logic
    return nil, nil
}

func (s *MyScraper) GetCasesByDateRange(ctx context.Context, start, end string) ([]*models.Case, error) {
    // Implement date range query
    return []*models.Case{}, nil
}

func (s *MyScraper) IsAvailable(ctx context.Context) bool {
    return s.running
}

func (s *MyScraper) GetJurisdiction() string {
    return "My Jurisdiction"
}

func main() {}
```

## Building Plugins

Plugins must be built as Go shared libraries:

```bash
# Build as plugin
go build -buildmode=plugin -o my-scraper.so my-scraper.go

# With version info
go build -buildmode=plugin \
  -ldflags "-X main.version=1.0.0" \
  -o my-scraper.so \
  my-scraper.go
```

### Build Script Example

```bash
#!/bin/bash

PLUGIN_NAME="my-scraper"
VERSION="1.0.0"

go build -buildmode=plugin \
  -ldflags "-X main.version=${VERSION}" \
  -o "${PLUGIN_NAME}.so" \
  "${PLUGIN_NAME}.go"

echo "Built ${PLUGIN_NAME}.so v${VERSION}"
```

## Installing Plugins

### 1. Copy to Plugin Directory

```bash
# Default plugin directory
cp my-scraper.so /path/to/kite/plugins/

# Custom plugin directory (set in config)
cp my-scraper.so /custom/plugin/dir/
```

### 2. Configure Plugin Directory

In `configs/default.yaml`:

```yaml
plugins:
  enabled: true
  directory: "/path/to/kite/plugins"
  auto_load: true
```

### 3. Restart Kite

Plugins are loaded at startup:

```bash
kite-api serve --config configs/default.yaml
```

## Plugin Lifecycle

```
┌─────────┐
│  Load   │  Plugin .so file loaded
└────┬────┘
     │
     v
┌─────────┐
│  New()  │  Plugin instance created
└────┬────┘
     │
     v
┌─────────┐
│  Init() │  Plugin initialized with config
└────┬────┘
     │
     v
┌─────────┐
│ Start() │  Plugin started and ready
└────┬────┘
     │
     v
┌─────────┐
│Running  │  Plugin handles requests
└────┬────┘
     │
     v
┌─────────┐
│ Stop()  │  Plugin gracefully stopped
└─────────┘
```

## Examples

### Processor Plugin

```go
package main

import (
    "context"
    "strings"
    "github.com/gongahkia/kite/internal/plugins"
    "github.com/gongahkia/kite/pkg/models"
)

type TitleCaseProcessor struct {
    name string
}

func New() plugins.Plugin {
    return &TitleCaseProcessor{
        name: "titlecase-processor",
    }
}

// Implement Plugin interface methods...

func (p *TitleCaseProcessor) Process(ctx context.Context, c *models.Case) (*models.Case, error) {
    // Convert case name to title case
    c.CaseName = strings.Title(strings.ToLower(c.CaseName))
    return c, nil
}

func (p *TitleCaseProcessor) CanProcess(c *models.Case) bool {
    return c.CaseName != ""
}

func (p *TitleCaseProcessor) Priority() int {
    return 10 // Lower priority
}

func main() {}
```

### Validator Plugin

```go
package main

import (
    "context"
    "time"
    "github.com/gongahkia/kite/internal/plugins"
    "github.com/gongahkia/kite/pkg/models"
)

type DateValidator struct {
    name string
}

func New() plugins.Plugin {
    return &DateValidator{
        name: "date-validator",
    }
}

// Implement Plugin interface methods...

func (v *DateValidator) Validate(ctx context.Context, c *models.Case) ([]plugins.ValidationError, error) {
    var errors []plugins.ValidationError

    // Check if decision date is in the future
    if c.DecisionDate != nil && c.DecisionDate.After(time.Now()) {
        errors = append(errors, plugins.ValidationError{
            Field:    "decision_date",
            Message:  "Decision date cannot be in the future",
            Severity: plugins.SeverityError,
        })
    }

    return errors, nil
}

func (v *DateValidator) Severity() plugins.ValidationSeverity {
    return plugins.SeverityError
}

func main() {}
```

### Exporter Plugin

```go
package main

import (
    "context"
    "encoding/json"
    "github.com/gongahkia/kite/internal/plugins"
    "github.com/gongahkia/kite/pkg/models"
)

type CustomJSONExporter struct {
    name string
}

func New() plugins.Plugin {
    return &CustomJSONExporter{
        name: "custom-json-exporter",
    }
}

// Implement Plugin interface methods...

func (e *CustomJSONExporter) Export(ctx context.Context, cases []*models.Case) ([]byte, error) {
    // Custom JSON format
    output := make(map[string]interface{})
    output["version"] = "1.0"
    output["cases"] = cases
    output["count"] = len(cases)

    return json.MarshalIndent(output, "", "  ")
}

func (e *CustomJSONExporter) FileExtension() string {
    return "json"
}

func (e *CustomJSONExporter) MIMEType() string {
    return "application/json"
}

func main() {}
```

## Best Practices

### 1. Error Handling

Always handle errors gracefully:

```go
func (s *MyScraper) SearchCases(ctx context.Context, query plugins.SearchQuery) ([]*models.Case, error) {
    // Check context
    if ctx.Err() != nil {
        return nil, ctx.Err()
    }

    // Validate input
    if query.Query == "" {
        return nil, fmt.Errorf("query cannot be empty")
    }

    // Handle errors from external services
    cases, err := s.fetchCases(query)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch cases: %w", err)
    }

    return cases, nil
}
```

### 2. Context Handling

Respect context cancellation:

```go
func (p *MyProcessor) Process(ctx context.Context, c *models.Case) (*models.Case, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Process with context awareness
    result := make(chan *models.Case)
    go func() {
        result <- p.doProcess(c)
    }()

    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    case processed := <-result:
        return processed, nil
    }
}
```

### 3. Configuration

Accept configuration in Init():

```go
func (s *MyScraper) Init(config map[string]interface{}) error {
    // Parse API key
    if apiKey, ok := config["api_key"].(string); ok {
        s.apiKey = apiKey
    } else {
        return fmt.Errorf("api_key is required")
    }

    // Parse optional timeout
    if timeout, ok := config["timeout"].(int); ok {
        s.timeout = time.Duration(timeout) * time.Second
    } else {
        s.timeout = 30 * time.Second // default
    }

    return nil
}
```

### 4. Logging

Use structured logging:

```go
import "github.com/rs/zerolog/log"

func (s *MyScraper) SearchCases(ctx context.Context, query plugins.SearchQuery) ([]*models.Case, error) {
    log.Info().
        Str("plugin", s.name).
        Str("query", query.Query).
        Msg("Searching cases")

    cases, err := s.fetchCases(query)
    if err != nil {
        log.Error().
            Err(err).
            Str("plugin", s.name).
            Msg("Search failed")
        return nil, err
    }

    log.Info().
        Str("plugin", s.name).
        Int("count", len(cases)).
        Msg("Search completed")

    return cases, nil
}
```

### 5. Resource Cleanup

Clean up resources in Stop():

```go
func (s *MyScraper) Stop(ctx context.Context) error {
    log.Info().Str("plugin", s.name).Msg("Stopping plugin")

    // Close HTTP client
    if s.client != nil {
        s.client.CloseIdleConnections()
    }

    // Close database connections
    if s.db != nil {
        s.db.Close()
    }

    s.running = false
    return nil
}
```

## Testing Plugins

### Unit Tests

```go
package main

import (
    "context"
    "testing"
)

func TestMyScraper_SearchCases(t *testing.T) {
    scraper := &MyScraper{
        name:    "test-scraper",
        running: true,
    }

    query := plugins.SearchQuery{
        Query: "test",
        Limit: 10,
    }

    cases, err := scraper.SearchCases(context.Background(), query)
    if err != nil {
        t.Fatalf("Search failed: %v", err)
    }

    if len(cases) == 0 {
        t.Error("Expected cases, got none")
    }
}
```

### Integration Tests

Test with actual Kite instance:

```bash
# Build plugin
go build -buildmode=plugin -o test-plugin.so test-plugin.go

# Copy to test plugin directory
cp test-plugin.so test/plugins/

# Run integration tests
go test -tags=integration ./test/plugins/
```

## Debugging Plugins

### Enable Verbose Logging

```bash
export LOG_LEVEL=debug
kite-api serve --config configs/default.yaml
```

### Check Plugin Loading

```bash
# View logs
journalctl -u kite-api -f | grep plugin

# Or with Docker
docker logs -f kite-api | grep plugin
```

### Common Issues

1. **Plugin not loading**: Check file permissions and `.so` architecture
2. **Symbol not found**: Ensure `New()` function is exported
3. **Version mismatch**: Rebuild plugin with same Go version as Kite
4. **Interface not implemented**: Verify all interface methods are implemented

## Distribution

### Packaging

```bash
# Create plugin package
tar -czf my-scraper-v1.0.0.tar.gz \
  my-scraper.so \
  README.md \
  LICENSE \
  config.example.yaml
```

### Installation Script

```bash
#!/bin/bash
# install.sh

PLUGIN_NAME="my-scraper"
PLUGIN_DIR="${PLUGIN_DIR:-/opt/kite/plugins}"

# Copy plugin
cp ${PLUGIN_NAME}.so ${PLUGIN_DIR}/

# Set permissions
chmod 644 ${PLUGIN_DIR}/${PLUGIN_NAME}.so

echo "Plugin installed to ${PLUGIN_DIR}/${PLUGIN_NAME}.so"
echo "Restart Kite to load the plugin"
```

## Support

- Plugin Examples: https://github.com/gongahkia/kite/tree/main/examples/plugins
- API Documentation: https://pkg.go.dev/github.com/gongahkia/kite/internal/plugins
- Issues: https://github.com/gongahkia/kite/issues
- Discord: https://discord.gg/kite
