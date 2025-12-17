# Kite Monitoring Stack

Complete monitoring and alerting setup for Kite using Prometheus, Grafana, and AlertManager.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Components](#components)
- [Configuration](#configuration)
- [Dashboards](#dashboards)
- [Alerts](#alerts)
- [Troubleshooting](#troubleshooting)

## Overview

The Kite monitoring stack provides:

- **Prometheus**: Metrics collection and storage
- **Grafana**: Visualization dashboards
- **AlertManager**: Alert routing and notifications
- **Node Exporter**: System metrics (optional)

## Quick Start

### Docker Compose

```bash
# Start the complete monitoring stack
docker-compose -f deployment/monitoring/docker-compose.yaml up -d

# Access services
# Prometheus: http://localhost:9090
# Grafana: http://localhost:3000 (admin/admin)
# AlertManager: http://localhost:9093
```

### Kubernetes

```bash
# Install Prometheus Operator (if not already installed)
kubectl create -f https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/bundle.yaml

# Deploy Kite monitoring
kubectl apply -f deployment/monitoring/k8s/

# Port forward to access Grafana
kubectl port-forward -n monitoring svc/grafana 3000:80
```

## Components

### Prometheus

Prometheus scrapes metrics from Kite API and worker instances.

**Configuration**: `deployment/prometheus/prometheus.yaml`

**Scrape Targets**:
- Kite API: `:9090/metrics`
- Kite Workers: `:9091/metrics`
- Node Exporter: `:9100/metrics` (system metrics)

**Alert Rules**: `deployment/prometheus/rules.yaml`

### Grafana

Pre-configured dashboards for monitoring Kite.

**Dashboards**:
- API Performance (`deployment/grafana/dashboards/api-performance.json`)
- Worker Health (`deployment/grafana/dashboards/worker-health.json`)
- Scraper Metrics (`deployment/grafana/dashboards/scraper-metrics.json`)
- System Overview (`deployment/grafana/dashboards/system-overview.json`)

**Default Credentials**: admin/admin (change on first login)

### AlertManager

Routes alerts to appropriate channels (email, Slack, PagerDuty).

**Configuration**: `deployment/alertmanager/config.yaml`

**Notification Channels**:
- Email (SMTP)
- Slack webhooks
- PagerDuty integration
- Custom webhooks

## Configuration

### Environment Variables

Set these environment variables for AlertManager:

```bash
# SMTP Configuration
export SMTP_PASSWORD="your-smtp-password"

# Slack Configuration
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

# PagerDuty Configuration
export PAGERDUTY_SERVICE_KEY="your-pagerduty-service-key"
export PAGERDUTY_DATABASE_KEY="your-database-team-key"

# Environment
export ENVIRONMENT="production"  # or "staging", "development"
```

### Prometheus Configuration

Edit `deployment/prometheus/prometheus.yaml`:

```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'kite-api'
    static_configs:
      - targets: ['kite-api:9090']

  - job_name: 'kite-worker'
    static_configs:
      - targets: ['kite-worker:9091']

  - job_name: 'node-exporter'
    static_configs:
      - targets: ['node-exporter:9100']
```

### Grafana Datasource

Grafana is pre-configured to use Prometheus as a datasource:

```yaml
apiVersion: 1
datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
```

## Dashboards

### 1. API Performance Dashboard

Monitors API health and performance:

- Request rate (total and by endpoint)
- Error rate and HTTP status codes
- Latency percentiles (P50, P95, P99)
- Active connections
- Request/response sizes
- Top slowest endpoints

**Alerts**:
- High error rate (>5%)
- High latency (>1s)

### 2. Worker Health Dashboard

Monitors worker pool and job processing:

- Worker utilization
- Active vs total workers
- Queue size and backlog
- Job processing rate (success/failed)
- Job success rate
- Job processing by type
- Job duration percentiles
- Dead letter queue size

**Alerts**:
- High queue backlog (>5000 jobs)
- All workers down
- High job failure rate (>10%)

### 3. Scraper Metrics Dashboard

Monitors web scraping operations:

- Success rate by jurisdiction
- Request rate by status
- Scraper latency
- Cases scraped (total and by jurisdiction)
- Active scrapers
- Rate limit status
- Robots.txt cache hits
- Scraper failures

**Alerts**:
- High scraper failure rate (>30%)
- Scraper blocked (rate limit/IP ban)
- Slow scraper performance (>30s)

### 4. System Overview Dashboard

High-level system health:

- System health status
- Total requests (24h)
- Total cases in database
- Active jobs
- Request rate and errors
- Memory usage
- Database connections
- Cache hit rate
- Goroutines
- CPU usage
- Data quality score
- Recent alerts

## Alerts

### Alert Severity Levels

- **critical**: Immediate action required (pages on-call)
- **warning**: Requires attention (Slack/email)
- **info**: Informational only

### Alert Groups

#### API Alerts
- `HighAPIErrorRate`: >5% error rate for 5 minutes
- `HighAPILatency`: P95 latency >1s for 10 minutes
- `APIEndpointDown`: API instance down for 1 minute
- `HighRequestRate`: Unusual spike in traffic

#### Worker Alerts
- `WorkerQueueBacklog`: >5000 jobs in queue for 10 minutes
- `WorkerQueueCritical`: >10000 jobs in queue for 5 minutes
- `LowWorkerUtilization`: <20% worker usage for 30 minutes
- `HighJobFailureRate`: >10% job failures for 10 minutes
- `AllWorkersDown`: No active workers for 5 minutes

#### Scraper Alerts
- `HighScraperFailureRate`: >30% failures for 15 minutes
- `ScraperBlocked`: Scraper blocked by site
- `SlowScraperPerformance`: P95 duration >30s for 20 minutes

#### Database Alerts
- `HighDatabaseConnections`: >80% of max connections
- `SlowDatabaseQueries`: P95 query time >5s
- `DatabaseConnectionFailures`: Connection errors detected

#### Cache Alerts
- `LowCacheHitRate`: <50% hit rate for 30 minutes
- `CacheMemoryHigh`: >90% of max keys

#### System Alerts
- `HighMemoryUsage`: >90% memory usage
- `HighCPUUsage`: >80% CPU usage for 10 minutes
- `DiskSpaceLow`: <10% disk space remaining
- `ContainerRestarting`: Pod restarting frequently

#### Business Alerts
- `NoRecentCaseIngestion`: No cases scraped in 1 hour
- `LowDataQuality`: Avg quality score <0.6 for 2 hours
- `HighDuplicateRate`: >20% duplicates for 2 hours

### Alert Routing

```
critical alerts → PagerDuty + Slack (#kite-critical)
warning alerts → Slack (#kite-alerts) + Email
info alerts → Slack (#kite-info)

By component:
- api → team-api (#kite-api)
- worker → team-workers (#kite-workers)
- scraper → team-scraper (#kite-scraper)
- database → team-database (#kite-database)
- business → team-product (#kite-product-metrics)
```

### Alert Inhibition

Some alerts suppress others to reduce noise:

- `APIEndpointDown` suppresses API latency/error alerts
- `AllWorkersDown` suppresses queue backlog alerts
- `DatabaseConnectionFailures` suppresses slow query alerts
- Higher severity suppresses lower severity for same alert

## Metrics Reference

### HTTP Metrics

```
http_requests_total - Total HTTP requests
http_request_duration_seconds - Request duration histogram
http_request_size_bytes - Request size histogram
http_response_size_bytes - Response size histogram
http_active_connections - Current active connections
```

### Worker Metrics

```
worker_total_count - Total worker count
worker_active_count - Active workers
worker_queue_size - Jobs in queue
worker_jobs_processed_total - Total jobs processed
worker_job_duration_seconds - Job duration histogram
```

### Scraper Metrics

```
scraper_requests_total - Scraper requests by status
scraper_duration_seconds - Scraper request duration
scraper_cases_scraped_total - Total cases scraped
scraper_rate_limit_remaining - Rate limit tokens remaining
scraper_robots_cache_hits - Robots.txt cache hits
scraper_robots_cache_misses - Robots.txt cache misses
```

### Database Metrics

```
database_connections_active - Active DB connections
database_connections_idle - Idle DB connections
database_connections_max - Max DB connections
database_query_duration_seconds - Query duration
database_errors_total - Database errors
```

### Cache Metrics

```
cache_hits_total - Cache hits
cache_misses_total - Cache misses
cache_keys_count - Current keys in cache
cache_evictions_total - Cache evictions
```

### Business Metrics

```
cases_total - Total cases in database
cases_created_total - Cases created counter
duplicates_detected_total - Duplicates detected
case_quality_score - Data quality score gauge
```

## Troubleshooting

### Prometheus Not Scraping Targets

```bash
# Check Prometheus targets status
curl http://localhost:9090/api/v1/targets

# Check if metrics endpoint is accessible
curl http://kite-api:9090/metrics
```

### Grafana Dashboards Not Loading

```bash
# Check Grafana logs
docker logs kite-grafana

# Verify Prometheus datasource
curl http://localhost:3000/api/datasources
```

### Alerts Not Firing

```bash
# Check AlertManager status
curl http://localhost:9093/api/v2/status

# Check alert rules in Prometheus
curl http://localhost:9090/api/v1/rules

# Test alert route
amtool alert add alertname="test" --alertmanager.url=http://localhost:9093
```

### Missing Metrics

```bash
# Check if Kite API is exposing metrics
curl http://localhost:8080/metrics

# Check Prometheus scrape errors
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | select(.health != "up")'
```

## Best Practices

1. **Set up proper retention**: Configure Prometheus retention based on storage
2. **Enable authentication**: Secure Grafana and Prometheus in production
3. **Configure backups**: Backup Grafana dashboards and Prometheus data
4. **Test alerts**: Regularly test alert routing to ensure notifications work
5. **Monitor the monitors**: Set up alerts for Prometheus/Grafana downtime
6. **Tune scrape intervals**: Balance between data resolution and resource usage
7. **Use recording rules**: Pre-compute expensive queries for dashboards
8. **Set up remote storage**: Use Thanos or Cortex for long-term storage

## Support

- Documentation: https://docs.kite.example.com/monitoring
- Runbooks: https://docs.kite.example.com/runbooks
- Issues: https://github.com/gongahkia/kite/issues
