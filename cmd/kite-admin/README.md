# kite-admin CLI Tool

Administration CLI for Kite Legal Case Law Platform

## Installation

```bash
# Build from source
go build -o kite-admin ./cmd/kite-admin

# Install
go install github.com/gongahkia/kite/cmd/kite-admin@latest
```

## Usage

```bash
kite-admin [command] [subcommand] [flags]
```

### Global Flags

- `--config, -c`: Config file path (default: "configs/default.yaml")
- `--env, -e`: Environment (development/staging/production)
- `--verbose, -v`: Verbose output
- `--json, -j`: Output in JSON format

## Commands

### migrate - Database Migrations

Manage database schema migrations.

```bash
# Apply all pending migrations
kite-admin migrate up

# Apply specific number of migrations
kite-admin migrate up --steps 2

# Rollback migrations
kite-admin migrate down --steps 1

# Show migration status
kite-admin migrate status

# Show current schema version
kite-admin migrate version

# Create new migration
kite-admin migrate create add_new_field
```

### worker - Worker Management

Manage worker pool.

```bash
# Start workers
kite-admin worker start --workers 4 --queue-size 1000

# Stop workers
kite-admin worker stop --timeout 30s

# Show worker status
kite-admin worker status

# List all workers
kite-admin worker list

# Scale worker pool
kite-admin worker scale 8
```

### cache - Cache Management

Manage application cache.

```bash
# Flush entire cache
kite-admin cache flush

# Show cache statistics
kite-admin cache stats

# Clear cache by pattern
kite-admin cache clear --pattern "case:*"

# Warm up cache
kite-admin cache warm

# List cache keys
kite-admin cache keys --limit 50
```

### queue - Job Queue

Inspect and manage job queue.

```bash
# List jobs
kite-admin queue list --status pending --limit 20

# Show queue statistics
kite-admin queue stats

# Purge completed jobs
kite-admin queue purge --status completed --force

# Retry failed job
kite-admin queue retry job_12345

# Manage dead letter queue
kite-admin queue dlq list
kite-admin queue dlq retry-all
kite-admin queue dlq clear
```

### health - Health Checks

Check system health.

```bash
# Full health check
kite-admin health check

# Check API health
kite-admin health api

# Check database health
kite-admin health database

# Check cache health
kite-admin health cache

# Check queue health
kite-admin health queue
```

### config - Configuration

View and validate configuration.

```bash
# Show current configuration
kite-admin config show --format yaml

# Validate configuration
kite-admin config validate

# Show environment variables
kite-admin config env
```

### metrics - Metrics Query

Query system metrics.

```bash
# Run PromQL query
kite-admin metrics query 'http_requests_total'

# Show API metrics
kite-admin metrics api

# Show worker metrics
kite-admin metrics worker

# Show scraper metrics
kite-admin metrics scraper
```

### backup - Backup & Restore

Backup and restore database.

```bash
# Create backup
kite-admin backup create --compress --output backup.sql.gz

# List backups
kite-admin backup list

# Restore from backup
kite-admin backup restore backup.sql.gz --force

# Delete backup
kite-admin backup delete old-backup.sql.gz
```

## Examples

### Daily Operations

```bash
# Morning health check
kite-admin health check

# Check worker status
kite-admin worker status

# View cache statistics
kite-admin cache stats

# Check queue backlog
kite-admin queue stats
```

### Database Maintenance

```bash
# Apply pending migrations
kite-admin migrate up

# Create database backup
kite-admin backup create --compress

# Verify migration status
kite-admin migrate status
```

### Performance Tuning

```bash
# Check API metrics
kite-admin metrics api

# Monitor worker performance
kite-admin worker list

# Clear old cache entries
kite-admin cache clear --pattern "old:*"

# Warm up cache
kite-admin cache warm
```

### Troubleshooting

```bash
# Check all health endpoints
kite-admin health check --verbose

# View worker details
kite-admin worker list

# Check dead letter queue
kite-admin queue dlq list

# Retry failed jobs
kite-admin queue dlq retry-all
```

### Configuration Management

```bash
# Validate configuration before deployment
kite-admin config validate

# View current configuration
kite-admin config show --format json

# Check environment variables
kite-admin config env
```

## JSON Output

Most commands support JSON output for scripting:

```bash
# Get worker status as JSON
kite-admin worker status --json

# Get cache stats as JSON
kite-admin cache stats --json

# Get health check as JSON
kite-admin health check --json
```

Example JSON output:

```json
{
  "status": "healthy",
  "version": "4.0.0",
  "uptime_seconds": 86400,
  "checks": {
    "api": "healthy",
    "database": "healthy",
    "cache": "healthy",
    "queue": "healthy",
    "workers": "healthy"
  }
}
```

## Automation

Use kite-admin in scripts and cron jobs:

```bash
#!/bin/bash

# Daily backup script
kite-admin backup create --compress

# Health check script
if ! kite-admin health check --json | jq -e '.status == "healthy"'; then
  echo "System unhealthy!" | mail -s "Kite Health Alert" admin@example.com
fi

# Cache maintenance
kite-admin cache clear --pattern "expired:*"
kite-admin cache warm
```

## Best Practices

1. **Always validate** configuration before deployment
2. **Create backups** before migrations or major changes
3. **Monitor metrics** regularly for performance issues
4. **Check health** before and after deployments
5. **Use --verbose** flag when troubleshooting
6. **Use --json** output for automation
7. **Set proper timeouts** for long-running operations
8. **Review queue stats** to prevent backlogs
9. **Warm cache** after deployments
10. **Test restore** procedures regularly

## Development

### Building

```bash
# Build with version info
go build -ldflags "\
  -X main.version=v4.0.0 \
  -X main.commit=$(git rev-parse --short HEAD) \
  -X main.buildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  -o kite-admin ./cmd/kite-admin
```

### Testing

```bash
# Run with verbose output
kite-admin --verbose config validate

# Test against different environments
kite-admin --env staging health check
kite-admin --env production health check
```

## Support

- Documentation: https://docs.kite.example.com/admin-cli
- Issues: https://github.com/gongahkia/kite/issues
- Email: support@kite.example.com
