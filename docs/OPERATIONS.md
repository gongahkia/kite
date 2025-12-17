# Kite Operations Guide

Complete guide for deploying, operating, and maintaining Kite v4 in production.

## Table of Contents

- [Deployment](#deployment)
- [Configuration](#configuration)
- [Monitoring](#monitoring)
- [Backup & Recovery](#backup--recovery)
- [Performance Tuning](#performance-tuning)
- [Troubleshooting](#troubleshooting)
- [Security](#security)
- [Upgrades](#upgrades)

## Deployment

### Quick Deployment Options

#### 1. Docker Compose (Development/Small Scale)

```bash
# Clone repository
git clone https://github.com/gongahkia/kite.git
cd kite

# Configure environment
cp .env.example .env
# Edit .env with your settings

# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f kite-api
```

#### 2. Kubernetes (Production/Large Scale)

```bash
# Create namespace
kubectl create namespace kite

# Apply configurations
kubectl apply -f deployment/k8s/namespace.yaml
kubectl apply -f deployment/k8s/configmap.yaml
kubectl apply -f deployment/k8s/secret.yaml
kubectl apply -f deployment/k8s/deployment.yaml
kubectl apply -f deployment/k8s/service.yaml
kubectl apply -f deployment/k8s/hpa.yaml

# Check deployment
kubectl get pods -n kite
kubectl get svc -n kite
```

#### 3. Helm Chart (Recommended for Kubernetes)

```bash
# Add Helm repository
helm repo add kite https://charts.kite.example.com
helm repo update

# Install with custom values
helm install kite kite/kite \
  --namespace kite \
  --create-namespace \
  --values production-values.yaml

# Upgrade
helm upgrade kite kite/kite \
  --namespace kite \
  --values production-values.yaml

# Uninstall
helm uninstall kite --namespace kite
```

### Production Deployment Architecture

```
┌─────────────────────────────────────────────┐
│              Load Balancer                   │
│         (NGINX/HAProxy/AWS ALB)              │
└──────────────┬──────────────────────────────┘
               │
       ┌───────┴────────┐
       │                │
┌──────▼──────┐  ┌─────▼──────┐
│  API Server │  │ API Server │  (N replicas)
│   Pod 1     │  │   Pod 2    │
└─────────────┘  └────────────┘
       │                │
       └────────┬───────┘
                │
       ┌────────▼─────────┐
       │  Worker Pool     │  (Auto-scaled)
       │  (4-16 workers)  │
       └────────┬─────────┘
                │
    ┌───────────┼───────────┐
    │           │           │
┌───▼───┐  ┌───▼────┐  ┌──▼─────┐
│Postgres│  │ Redis  │  │  NATS  │
│ (HA)   │  │(Cluster│  │(Cluster│
└────────┘  └────────┘  └────────┘
```

### System Requirements

#### Minimum (Development)
- **CPU**: 2 cores
- **RAM**: 4 GB
- **Disk**: 20 GB SSD
- **Network**: 10 Mbps

#### Recommended (Production)
- **API Server**:
  - CPU: 4-8 cores
  - RAM: 8-16 GB
  - Disk: 50 GB SSD

- **Worker Node**:
  - CPU: 2-4 cores
  - RAM: 4-8 GB
  - Disk: 20 GB SSD

- **Database (PostgreSQL)**:
  - CPU: 8+ cores
  - RAM: 16-32 GB
  - Disk: 500 GB SSD (with high IOPS)

- **Cache (Redis)**:
  - CPU: 2-4 cores
  - RAM: 8-16 GB
  - Disk: 20 GB SSD

## Configuration

### Environment Variables

```bash
# Server Configuration
SERVER_HTTP_PORT=8080
SERVER_GRPC_PORT=50051
SERVER_ENABLE_GRPC=true

# Database
DATABASE_DRIVER=postgres
DATABASE_URL=postgres://user:pass@localhost:5432/kite?sslmode=require
DATABASE_MAX_CONNECTIONS=25
DATABASE_IDLE_CONNECTIONS=5

# Cache
CACHE_DRIVER=redis
CACHE_URL=redis://localhost:6379
CACHE_MAX_KEYS=10000
CACHE_DEFAULT_TTL=3600

# Queue
QUEUE_DRIVER=nats
QUEUE_URL=nats://localhost:4222
QUEUE_MAX_PENDING=10000

# Workers
WORKER_COUNT=4
WORKER_QUEUE_SIZE=1000
WORKER_BATCH_SIZE=100

# Logging
LOG_LEVEL=info
LOG_FORMAT=json
LOG_OUTPUT=stdout

# Metrics
METRICS_ENABLED=true
METRICS_PORT=9090

# Security
JWT_SECRET=your-secret-key-here
API_KEY_HEADER=X-API-Key
CORS_ALLOWED_ORIGINS=https://app.example.com

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=60
```

### Configuration File

```yaml
# config/production.yaml
server:
  http_port: 8080
  grpc_port: 50051
  enable_grpc: true
  shutdown_timeout: 30s
  read_timeout: 30s
  write_timeout: 30s

database:
  driver: postgres
  url: ${DATABASE_URL}
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: 5m
  auto_migrate: false

cache:
  driver: redis
  url: ${REDIS_URL}
  max_keys: 10000
  default_ttl: 1h
  multilevel:
    enabled: true
    l1_ttl: 5m
    l2_ttl: 1h

queue:
  driver: nats
  url: ${NATS_URL}
  max_pending: 10000
  ack_wait: 30s
  max_deliver: 3

workers:
  count: 4
  queue_size: 1000
  batch_size: 100
  shutdown_timeout: 30s

observability:
  logging:
    level: info
    format: json
    output: stdout

  metrics:
    enabled: true
    port: 9090
    path: /metrics

  tracing:
    enabled: true
    endpoint: http://jaeger:14268/api/traces
    sample_rate: 0.1

security:
  jwt:
    secret: ${JWT_SECRET}
    expiration: 1h
    refresh_expiration: 168h

  cors:
    allowed_origins:
      - https://app.example.com
    allowed_methods:
      - GET
      - POST
      - PUT
      - DELETE
    allowed_headers:
      - Authorization
      - Content-Type

  rate_limiting:
    enabled: true
    default_limit: 60
    burst: 10
    cleanup_interval: 1m
```

## Monitoring

### Health Checks

```bash
# Liveness probe
curl http://localhost:8080/health

# Readiness probe
curl http://localhost:8080/ready

# Response
{
  "status": "healthy",
  "version": "4.0.0",
  "uptime": 3600,
  "checks": {
    "database": "healthy",
    "cache": "healthy",
    "queue": "healthy"
  }
}
```

### Metrics

Prometheus metrics available at `/metrics`:

```prometheus
# HTTP Requests
http_requests_total{method="GET",path="/api/v1/cases",status="200"}
http_request_duration_seconds{method="GET",path="/api/v1/cases"}

# Database
database_connections_active
database_connections_idle
database_query_duration_seconds

# Workers
worker_jobs_processed_total{job_type="scrape",status="success"}
worker_queue_size
worker_active_count

# Cache
cache_hits_total
cache_misses_total
cache_keys_count

# Scraping
scraper_requests_total{jurisdiction="Australia",status="success"}
scraper_duration_seconds{jurisdiction="Australia"}
```

### Grafana Dashboards

Import pre-built dashboards:

```bash
# API Performance Dashboard
curl -o dashboards/api-performance.json \
  https://grafana.com/api/dashboards/12345/revisions/1/download

# Worker Health Dashboard
curl -o dashboards/worker-health.json \
  https://grafana.com/api/dashboards/12346/revisions/1/download
```

### Logging

Structured JSON logs with correlation IDs:

```json
{
  "timestamp": "2023-12-16T10:00:00Z",
  "level": "info",
  "message": "Case created successfully",
  "request_id": "req_abc123",
  "user_id": "user_456",
  "case_id": "cth/HCA/2023/15",
  "duration_ms": 42,
  "context": {
    "handler": "CreateCase",
    "method": "POST",
    "path": "/api/v1/cases"
  }
}
```

### Alerting Rules

```yaml
# prometheus-alerts.yaml
groups:
  - name: kite
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.05
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: High error rate detected
          description: Error rate is {{ $value }} (threshold: 0.05)

      - alert: WorkerQueueBacklog
        expr: worker_queue_size > 5000
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: Worker queue backlog
          description: Queue size is {{ $value }} (threshold: 5000)

      - alert: DatabaseConnectionsHigh
        expr: database_connections_active / database_connections_max > 0.8
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: Database connections near limit
          description: Using {{ $value }}% of max connections
```

## Backup & Recovery

### Database Backup

```bash
# Automated daily backup
cat > /etc/cron.d/kite-backup <<EOF
0 2 * * * postgres pg_dump kite | gzip > /backups/kite-\$(date +\%Y\%m\%d).sql.gz
EOF

# Manual backup
pg_dump -h localhost -U kite kite > kite-backup.sql

# Restore
psql -h localhost -U kite kite < kite-backup.sql
```

### Configuration Backup

```bash
# Backup configuration
kubectl get configmap kite-config -n kite -o yaml > config-backup.yaml
kubectl get secret kite-secrets -n kite -o yaml > secrets-backup.yaml

# Restore
kubectl apply -f config-backup.yaml
kubectl apply -f secrets-backup.yaml
```

### Disaster Recovery

1. **Database**: Use PostgreSQL streaming replication or managed DB service
2. **Cache**: Redis cluster with persistence enabled
3. **Queue**: NATS JetStream with replication
4. **Files**: S3-compatible storage with versioning

## Performance Tuning

### Database Optimization

```sql
-- Create indexes for common queries
CREATE INDEX idx_cases_jurisdiction ON cases(jurisdiction);
CREATE INDEX idx_cases_decision_date ON cases(decision_date);
CREATE INDEX idx_cases_court ON cases(court);
CREATE INDEX idx_citations_source ON citations(source_case_id);

-- Full-text search index
CREATE INDEX idx_cases_fulltext ON cases USING GIN(to_tsvector('english', full_text));

-- Analyze tables
ANALYZE cases;
ANALYZE citations;
```

### Connection Pooling

```yaml
database:
  max_open_conns: 25  # Max concurrent connections
  max_idle_conns: 5   # Idle connections in pool
  conn_max_lifetime: 5m  # Connection reuse time
```

### Cache Configuration

```yaml
cache:
  multilevel:
    enabled: true
    l1_ttl: 5m      # In-memory cache (fast)
    l2_ttl: 1h      # Redis cache (persistent)
  max_keys: 10000   # LRU eviction threshold
```

### Worker Tuning

```yaml
workers:
  count: 8          # Increase for more throughput
  queue_size: 2000  # Buffer more jobs
  batch_size: 200   # Process more per batch
```

## Troubleshooting

### Common Issues

#### 1. High Memory Usage

```bash
# Check memory usage
kubectl top pods -n kite

# Analyze heap dump (Go)
curl http://localhost:6060/debug/pprof/heap > heap.prof
go tool pprof heap.prof

# Solutions:
# - Reduce cache max_keys
# - Decrease worker batch_size
# - Enable memory limits in Kubernetes
```

#### 2. Slow API Responses

```bash
# Check metrics
curl http://localhost:9090/metrics | grep http_request_duration

# Enable profiling
curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
go tool pprof cpu.prof

# Solutions:
# - Add database indexes
# - Enable caching
# - Increase connection pool size
```

#### 3. Worker Queue Backlog

```bash
# Check queue size
curl http://localhost:9090/metrics | grep worker_queue_size

# Solutions:
# - Increase worker count
# - Optimize job processing
# - Add more worker nodes
```

### Debug Mode

```bash
# Enable debug logging
export LOG_LEVEL=debug

# Enable profiling
export ENABLE_PPROF=true
export PPROF_PORT=6060

# View CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile
```

### Logs Analysis

```bash
# Find errors
kubectl logs -n kite deployment/kite-api | grep '"level":"error"'

# Filter by request ID
kubectl logs -n kite deployment/kite-api | grep 'req_abc123'

# Tail logs
kubectl logs -f -n kite deployment/kite-api
```

## Security

### TLS/SSL Configuration

```yaml
server:
  tls:
    enabled: true
    cert_file: /etc/kite/tls/server.crt
    key_file: /etc/kite/tls/server.key
    min_version: "1.2"
```

### Secret Management

```bash
# Kubernetes secrets
kubectl create secret generic kite-secrets \
  --from-literal=jwt-secret=$(openssl rand -base64 32) \
  --from-literal=db-password=$(openssl rand -base64 24) \
  -n kite

# Vault integration
vault kv put secret/kite \
  jwt_secret=$(openssl rand -base64 32) \
  db_password=$(openssl rand -base64 24)
```

### Network Security

```yaml
# NetworkPolicy (Kubernetes)
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: kite-api
spec:
  podSelector:
    matchLabels:
      app: kite-api
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: ingress-nginx
      ports:
        - protocol: TCP
          port: 8080
```

## Upgrades

### Zero-Downtime Upgrade

```bash
# 1. Backup database
pg_dump kite > pre-upgrade-backup.sql

# 2. Run migrations
./kite-admin migrate up

# 3. Rolling update (Kubernetes)
kubectl set image deployment/kite-api \
  kite-api=kite:v4.1.0 \
  -n kite

# 4. Monitor rollout
kubectl rollout status deployment/kite-api -n kite

# 5. Rollback if needed
kubectl rollout undo deployment/kite-api -n kite
```

### Version Compatibility

- **Minor versions** (4.0 → 4.1): Zero-downtime upgrade
- **Major versions** (4.0 → 5.0): Requires maintenance window
- **Database migrations**: Automatic on startup (configurable)

## Best Practices

1. **Always** use TLS in production
2. **Enable** authentication and RBAC
3. **Monitor** key metrics (error rate, latency, queue size)
4. **Set up** alerts for critical failures
5. **Backup** database daily
6. **Test** disaster recovery procedures
7. **Keep** secrets in vault, not config files
8. **Use** resource limits in Kubernetes
9. **Enable** rate limiting
10. **Review** logs regularly for security issues

## Support

- **Documentation**: https://docs.kite.example.com
- **Status Page**: https://status.kite.example.com
- **GitHub Issues**: https://github.com/gongahkia/kite/issues
- **Email**: ops@kite.example.com
- **Emergency**: support@kite.example.com
