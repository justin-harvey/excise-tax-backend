# API Gateway Service

The API Gateway is the central entry point for all client requests to the Excise Tax Portal backend. It handles authentication, rate limiting, routing, and request/response logging.

## Features

- **Authentication & Authorization**: JWT token validation with role-based access control
- **Rate Limiting**: Distributed rate limiting using Redis with sliding window algorithm
- **Request Routing**: Intelligent routing to backend microservices
- **CORS Handling**: Configurable cross-origin resource sharing
- **Request Logging**: Structured logging with request ID propagation
- **Health Checks**: Comprehensive health monitoring of all backend services
- **Metrics**: Prometheus-compatible metrics endpoint
- **Panic Recovery**: Graceful panic recovery with detailed logging
- **Connection Pooling**: Efficient HTTP connection pooling to backend services

## Architecture

```
Client Request
      |
      v
  API Gateway (Port 8080)
      |
      +--- Middleware Chain:
      |    1. Request ID Generation
      |    2. Panic Recovery
      |    3. CORS
      |    4. Request Logging
      |    5. Authentication (route-specific)
      |    6. Rate Limiting (route-specific)
      |
      +--- Route to Backend Service:
           - Auth Service (8084)
           - Payment Service (8081)
           - Tax Service (8082)
           - Reporting Service (8083)
```

## Endpoints

### Public Endpoints

- `GET /health` - Overall health check
- `GET /health/liveness` - Kubernetes liveness probe
- `GET /health/readiness` - Kubernetes readiness probe
- `GET /metrics` - Prometheus metrics

### API v1 Routes

#### Authentication (via Auth Service)
- `POST /api/v1/auth/login` - User login (anonymous)
- `POST /api/v1/auth/register` - User registration (anonymous)
- `POST /api/v1/auth/forgot-password` - Password reset request (anonymous)
- `POST /api/v1/auth/reset-password` - Password reset (anonymous)
- `POST /api/v1/auth/logout` - User logout (authenticated)
- `GET /api/v1/auth/me` - Get current user (authenticated)
- `PUT /api/v1/auth/password` - Change password (authenticated)
- `PUT /api/v1/auth/profile` - Update profile (authenticated)

#### Payments (via Payment Service)
- `POST /api/v1/payments` - Create payment (authenticated)
- `GET /api/v1/payments/:id` - Get payment details (authenticated)
- `GET /api/v1/payments` - List payments (authenticated)
- `PUT /api/v1/payments/:id/status` - Update payment status (authenticated)
- `POST /api/v1/payments/:id/refund` - Process refund (authenticated)

#### Tax (via Tax Service)
- `POST /api/v1/tax/calculate` - Calculate tax (authenticated)
- `POST /api/v1/tax/validate` - Validate tax filing (authenticated)
- `POST /api/v1/tax/filings` - Create tax filing (authenticated)
- `GET /api/v1/tax/filings/:id` - Get filing details (authenticated)
- `GET /api/v1/tax/filings` - List filings (authenticated)
- `PUT /api/v1/tax/filings/:id` - Update filing (authenticated)
- `POST /api/v1/tax/filings/:id/submit` - Submit filing (authenticated)
- `GET /api/v1/tax/rates` - Get tax rates (authenticated)
- `GET /api/v1/tax/rates/:commodity` - Get commodity rates (authenticated)

#### Reports (via Reporting Service)
- `POST /api/v1/reports/generate` - Generate report (authenticated)
- `GET /api/v1/reports/:id` - Get report details (authenticated)
- `GET /api/v1/reports` - List reports (authenticated)
- `GET /api/v1/reports/:id/download` - Download report (authenticated)
- `DELETE /api/v1/reports/:id` - Delete report (authenticated)

#### Admin (Admin Role Required)
- `GET /api/v1/admin/users` - List all users
- `GET /api/v1/admin/users/:id` - Get user details
- `PUT /api/v1/admin/users/:id` - Update user
- `DELETE /api/v1/admin/users/:id` - Delete user
- `PUT /api/v1/admin/users/:id/role` - Update user role
- `GET /api/v1/admin/config` - Get system configuration
- `PUT /api/v1/admin/config` - Update system configuration
- `GET /api/v1/admin/analytics` - Get analytics data
- `GET /api/v1/admin/audit-logs` - Get audit logs

## Configuration

The API Gateway can be configured via:
1. YAML configuration file (`configs/api-gateway.yaml`)
2. Environment variables (override YAML settings)

### Environment Variables

- `PORT` - Server port (default: 8080)
- `REDIS_HOST` - Redis host (default: localhost)
- `REDIS_PORT` - Redis port (default: 6379)
- `REDIS_PASSWORD` - Redis password
- `AUTH_SERVICE_URL` - Auth service URL (default: http://localhost:8084)
- `PAYMENT_SERVICE_URL` - Payment service URL (default: http://localhost:8081)
- `TAX_SERVICE_URL` - Tax service URL (default: http://localhost:8082)
- `REPORTING_SERVICE_URL` - Reporting service URL (default: http://localhost:8083)
- `CONFIG_PATH` - Path to config file (default: configs/config.yaml)

## Rate Limiting

The API Gateway implements distributed rate limiting using Redis:

- **Algorithm**: Sliding window with token bucket
- **Default Limits**: 100 requests per 15 minutes per user
- **Burst Capacity**: 20 additional requests
- **Scope**: Per-user and per-IP rate limiting

Rate limit headers are included in responses:
- `X-RateLimit-Limit`: Maximum requests allowed
- `X-RateLimit-Remaining`: Remaining requests in current window
- `X-RateLimit-Reset`: Unix timestamp when limit resets

## Authentication

The gateway validates JWT tokens by forwarding them to the Auth Service. User claims are extracted and propagated to backend services via headers:

- `X-User-ID`: Authenticated user ID
- `X-User-Role`: User's primary role
- `X-Request-ID`: Unique request identifier

## Health Checks

The `/health` endpoint returns comprehensive health status:

```json
{
  "status": "healthy",
  "services": {
    "auth": "healthy",
    "payment": "healthy",
    "tax": "healthy",
    "reporting": "healthy"
  },
  "dependencies": {
    "redis": "healthy"
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

Status values: `healthy`, `degraded`, `unhealthy`

## Running the Service

### Prerequisites
- Go 1.21 or higher
- Redis server running
- Backend services running (auth, payment, tax, reporting)

### Development
```bash
# From the backend directory
cd cmd/api-gateway
go run main.go
```

### Production
```bash
# Build the binary
go build -o api-gateway cmd/api-gateway/main.go

# Run the binary
./api-gateway
```

### Docker
```bash
# Build Docker image
docker build -f deployments/docker/Dockerfile.api-gateway -t api-gateway:latest .

# Run container
docker run -p 8080:8080 \
  -e REDIS_HOST=redis \
  -e AUTH_SERVICE_URL=http://auth:8084 \
  api-gateway:latest
```

## Monitoring

### Prometheus Metrics

The `/metrics` endpoint exposes Prometheus-compatible metrics:

- HTTP request count by endpoint and status
- Request duration histogram
- Active connections
- Rate limit hits
- Backend service health status

### Logging

Structured logs include:
- Request ID for correlation
- User ID (if authenticated)
- HTTP method and path
- Response status code
- Request duration
- Client IP address
- Error details

## Security

- All requests are logged with request IDs for audit trails
- Rate limiting prevents abuse
- CORS is strictly enforced
- Authentication tokens are validated before routing
- Sensitive headers are not forwarded to backend services
- Panic recovery prevents service crashes

## Graceful Shutdown

The service handles SIGINT and SIGTERM signals for graceful shutdown:
1. Stop accepting new connections
2. Wait for active requests to complete (up to 30 seconds)
3. Close Redis connections
4. Flush logs

## Development

### Adding New Routes

1. Add route configuration in `router/router.go`
2. Apply appropriate middleware (auth, rate limiting)
3. Use `serviceProxy.ProxyRequest()` to forward to backend service

### Adding New Middleware

1. Create middleware in `middleware/` directory
2. Implement as `gin.HandlerFunc`
3. Add to middleware chain in `router/router.go`

### Testing

```bash
# Run tests
go test ./internal/api-gateway/...

# Run with coverage
go test -cover ./internal/api-gateway/...
```

## Troubleshooting

### Common Issues

1. **503 Service Unavailable**: Backend service is down or unreachable
2. **429 Too Many Requests**: Rate limit exceeded
3. **401 Unauthorized**: Invalid or expired JWT token
4. **502 Bad Gateway**: Network error communicating with backend

### Debug Mode

Set `logging.environment: development` in config for detailed debug logs.

## Performance Tuning

- Connection pooling configured for 100 max idle connections
- HTTP client timeout: 30 seconds
- Read/Write timeouts: 30 seconds
- Redis connection pooling enabled
- Health checks run every 30 seconds
