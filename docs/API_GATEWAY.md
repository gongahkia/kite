# Kite API Gateway Documentation

## Overview

The Kite API Gateway provides a complete RESTful API with authentication, rate limiting, and OpenAPI documentation.

## Features

### 1. OpenAPI/Swagger Documentation

Interactive API documentation is available at:
- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **Docs redirect**: `http://localhost:8080/docs`
- **OpenAPI JSON**: `http://localhost:8080/swagger/doc.json`

#### Generating Swagger Documentation

```bash
# Install swag CLI tool
go install github.com/swaggo/swag/cmd/swag@latest

# Generate documentation
make swagger

# Or manually:
swag init -g internal/api/swagger.go -o docs --parseDependency --parseInternal
```

The documentation is automatically served when you run the API server.

### 2. Authentication

The API supports two authentication methods:

#### JWT Authentication

**Login to get a token:**

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "your-password"
  }'
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2025-12-17T10:00:00Z",
  "user_id": "admin",
  "client_id": "client_admin",
  "roles": ["user", "read", "admin", "write"]
}
```

**Use the token:**
```bash
curl -X GET http://localhost:8080/api/v1/cases \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

**Refresh token:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Authorization: Bearer YOUR_CURRENT_TOKEN"
```

**Validate token:**
```bash
curl -X GET http://localhost:8080/api/v1/auth/validate \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

#### API Key Authentication

**Generate an API key (requires admin role):**
```bash
curl -X POST http://localhost:8080/api/v1/auth/api-key \
  -H "Authorization: Bearer YOUR_ADMIN_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "client_id": "my-app",
    "description": "API key for my application"
  }'
```

**Use API key:**
```bash
# Method 1: X-API-Key header
curl -X GET http://localhost:8080/api/v1/cases \
  -H "X-API-Key: YOUR_API_KEY"

# Method 2: Authorization header
curl -X GET http://localhost:8080/api/v1/cases \
  -H "Authorization: ApiKey YOUR_API_KEY"
```

### 3. Rate Limiting

The API implements multiple levels of rate limiting:

#### Global Rate Limit
- **Limit**: 100 requests/second per IP
- **Burst**: 200 requests
- **Headers**:
  - `X-RateLimit-Limit`: Maximum requests allowed
  - `X-RateLimit-Remaining`: Remaining requests in current window
  - `Retry-After`: Seconds to wait before retrying (if limited)

#### Endpoint-Specific Rate Limits

Different endpoints have different rate limits:

| Endpoint | RPS | Burst |
|----------|-----|-------|
| `/api/v1/cases` | 10 | 20 |
| `/api/v1/judges` | 15 | 30 |
| `/api/v1/citations` | 20 | 40 |
| `/api/v1/search` | 5 | 10 |
| Default | 10 | 20 |

Rate limiting is applied per client ID (if authenticated) or per IP address (if unauthenticated).

**429 Too Many Requests Response:**
```json
{
  "error": "Rate limit exceeded for this endpoint",
  "limit": 10,
  "endpoint": "/api/v1/search",
  "message": "Too many requests to this endpoint, please slow down"
}
```

### 4. Middleware

The API uses a middleware chain pattern:

1. **Request ID**: Adds unique ID to each request
2. **Logger**: Logs all requests with timing
3. **CORS**: Handles Cross-Origin Resource Sharing
4. **Recovery**: Recovers from panics gracefully
5. **Metrics**: Records Prometheus metrics
6. **Global Rate Limit**: 100 req/s per IP
7. **Authentication**: JWT or API Key (optional by default)
8. **Endpoint Rate Limit**: Per-endpoint rate limiting

### 5. Role-Based Access Control

JWT tokens can include roles for fine-grained access control:

```go
// Example: Require admin role for certain endpoints
router.Post("/admin/settings",
  middleware.RequireRoles("admin"),
  handler.UpdateSettings
)
```

**Roles returned in JWT:**
- `user`: Basic user access
- `read`: Read-only access
- `write`: Write access
- `admin`: Administrative access

### 6. Security Best Practices

#### Configuration

Add to `configs/default.yaml`:

```yaml
security:
  jwt_secret: "your-secret-key-min-32-chars-long"  # CHANGE IN PRODUCTION
  jwt_expiration: "24h"
  api_keys:
    "key1": "client1"
    "key2": "client2"
```

Or use environment variables:

```bash
export KITE_SECURITY_JWT_SECRET="your-secret-key"
export KITE_SECURITY_JWT_EXPIRATION="24h"
```

#### Production Recommendations

1. **JWT Secret**: Use a strong, randomly generated secret (min 32 characters)
2. **HTTPS**: Always use HTTPS in production
3. **API Keys**: Store securely, rotate regularly
4. **Rate Limits**: Adjust based on your traffic patterns
5. **CORS**: Configure allowed origins appropriately
6. **Logging**: Monitor authentication failures

### 7. Error Responses

All errors follow a consistent format:

```json
{
  "error": "Error message",
  "request_id": "20251216153045-abc123",
  "path": "/api/v1/cases"
}
```

**Common Status Codes:**
- `400 Bad Request`: Invalid request body or parameters
- `401 Unauthorized`: Missing or invalid authentication
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

### 8. API Endpoints

#### Health & Monitoring
- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /metrics` - Prometheus metrics

#### Authentication
- `POST /api/v1/auth/login` - Login with username/password
- `POST /api/v1/auth/refresh` - Refresh JWT token
- `GET /api/v1/auth/validate` - Validate token
- `POST /api/v1/auth/api-key` - Generate API key (admin only)

#### Cases
- `GET /api/v1/cases` - List cases
- `GET /api/v1/cases/:id` - Get case by ID
- `POST /api/v1/cases` - Create case
- `PUT /api/v1/cases/:id` - Update case
- `DELETE /api/v1/cases/:id` - Delete case
- `POST /api/v1/cases/search` - Search cases

#### Judges
- `GET /api/v1/judges` - List judges
- `GET /api/v1/judges/:id` - Get judge by ID
- `POST /api/v1/judges` - Create judge
- `PUT /api/v1/judges/:id` - Update judge

#### Citations
- `GET /api/v1/citations` - List citations
- `GET /api/v1/citations/:id` - Get citation by ID
- `POST /api/v1/citations` - Create citation

#### Statistics
- `GET /api/v1/stats` - Get system statistics
- `GET /api/v1/stats/storage` - Get storage statistics

## Testing

### cURL Examples

```bash
# Login
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"pass"}' | jq -r '.token')

# List cases with authentication
curl -X GET http://localhost:8080/api/v1/cases \
  -H "Authorization: Bearer $TOKEN"

# Create a case
curl -X POST http://localhost:8080/api/v1/cases \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "case_name": "Smith v. Jones",
    "court": "Supreme Court",
    "date": "2025-01-15",
    "jurisdiction": "US"
  }'
```

### HTTPie Examples

```bash
# Login
http POST :8080/api/v1/auth/login username=admin password=pass

# Get cases
http GET :8080/api/v1/cases "Authorization:Bearer $TOKEN"

# Create case
http POST :8080/api/v1/cases \
  "Authorization:Bearer $TOKEN" \
  case_name="Smith v. Jones" \
  court="Supreme Court"
```

## Monitoring

### Prometheus Metrics

The API exposes Prometheus metrics at `/metrics`:

- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - HTTP request duration
- `http_requests_in_flight` - Current in-flight requests
- Rate limit hits, auth successes/failures, etc.

### Logs

Structured JSON logs include:
- `request_id`: Unique request identifier
- `method`: HTTP method
- `path`: Request path
- `status`: Response status code
- `duration`: Request duration in milliseconds
- `ip`: Client IP address
- Authentication events, rate limit events

## Customization

### Custom Rate Limits

```go
// In routes.go
config := &middleware.EndpointRateLimitConfig{
  Limits: map[string]*middleware.RateLimitConfig{
    "/api/v1/my-endpoint": {RPS: 5, Burst: 10},
  },
  UseClientID: true,
}
api.Use(middleware.EndpointRateLimit(config, logger))
```

### Custom Authentication

```go
// Skip auth for specific routes
authConfig.Skipper = func(c *fiber.Ctx) bool {
  return c.Path() == "/api/v1/public"
}
```

### Custom Middleware

```go
// Add custom middleware to the chain
app.Use(func(c *fiber.Ctx) error {
  // Your custom logic
  return c.Next()
})
```

## Troubleshooting

### "Missing API key" error
- Ensure you're sending the `X-API-Key` header or `Authorization: ApiKey <key>` header

### "Invalid or expired token" error
- Token may have expired (default 24h)
- Use `/api/v1/auth/refresh` to get a new token
- Verify JWT secret matches between config and runtime

### "Rate limit exceeded" error
- Wait for the time specified in `Retry-After` header
- Consider increasing rate limits for your use case
- Use authentication to get client-based rate limiting (higher limits)

### Swagger docs not showing
- Run `make swagger` to generate documentation
- Ensure `docs/` directory exists
- Check that swag is installed: `go install github.com/swaggo/swag/cmd/swag@latest`

## Next Steps

- See [gRPC API](./GRPC.md) for gRPC endpoints
- See [Search API](./SEARCH.md) for advanced search capabilities
- See [Deployment](./DEPLOYMENT.md) for production deployment
