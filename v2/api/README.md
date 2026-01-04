# XRPL API Documentation

RESTful API for interacting with the XRP Ledger (XRPL).

## Features

- **Account Operations**: Query balances, account info, and transaction history
- **Transaction Operations**: Get transaction details and verify status
- **Health Checks**: Readiness and liveness probes for Kubernetes
- **Rate Limiting**: Per-IP rate limiting to prevent abuse
- **OpenAPI 3.0**: Complete API specification with Swagger UI
- **Graceful Shutdown**: Properly handles SIGINT/SIGTERM signals

## Quick Start

### Using Docker (Recommended)

```bash
# Development mode
docker-compose --profile dev up

# Production mode
docker-compose --profile prod up
```

The API will be available at:
- API: http://localhost:8080
- Swagger UI: http://localhost:8081/swagger

### Local Development

```bash
# Build the API
cd cmd/api
go build -o api

# Run with default settings
./api

# Run with custom settings
./api -port=9000 -env=development -log-level=debug
```

## Configuration

Configuration is loaded in the following priority:
1. Command-line flags
2. Environment variables
3. Default values

### Command-Line Flags

```
-port int          API server port (default: 8080)
-env string        Environment: development|production (default: development)
-xrpl-url string   XRPL WebSocket URL (default: testnet)
-log-level string  Log level: debug|info|warn|error (default: info)
-log-format string Log format: json|text (default: json)
-rate-limit bool   Enable rate limiting (default: true)
-rate-limit-rps float  Requests per second (default: 10)
-rate-limit-burst int  Burst size (default: 20)
```

### Environment Variables

```bash
API_PORT=8080
ENV=development
XRPL_URL=wss://s.altnet.rippletest.net:51233
LOG_LEVEL=debug
```

## API Endpoints

### Health Checks

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Basic health check |
| `/health/ready` | GET | Readiness probe (checks XRPL connection) |
| `/health/live` | GET | Liveness probe |
| `/info` | GET | API version and capabilities |

### Account Operations

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/xrpl/accounts/{address}/balance` | GET | Get account balance |
| `/xrpl/accounts/{address}/info` | GET | Get detailed account info |
| `/xrpl/accounts/{address}/transactions` | GET | Get transaction history |

Query parameters for transactions:
- `limit`: Number of transactions (1-100, default: 10)

### Transaction Operations

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/xrpl/transactions/{hash}` | GET | Get transaction details |
| `/xrpl/transactions/{hash}/status` | GET | Verify transaction status |

## Examples

### Get Account Balance

```bash
curl http://localhost:8080/xrpl/accounts/rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY/balance
```

Response:
```json
{
  "account": "rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY",
  "balance_xrp": "777.495588",
  "balance_drops": 777495588
}
```

### Get Account Info

```bash
curl http://localhost:8080/xrpl/accounts/rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY/info
```

### Get Transaction History

```bash
curl "http://localhost:8080/xrpl/accounts/rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY/transactions?limit=5"
```

### Get Transaction Details

```bash
curl http://localhost:8080/xrpl/transactions/<hash>
```

### Check Transaction Status

```bash
curl http://localhost:8080/xrpl/transactions/<hash>/status
```

## Error Handling

All errors follow RFC 7807 Problem Details format:

```json
{
  "error": {
    "type": "https://api.example.com/errors/404",
    "title": "Not Found",
    "status": 404,
    "detail": "The requested resource could not be found",
    "instance": "/xrpl/accounts/rNotFound/balance"
  }
}
```

### Common HTTP Status Codes

- `200 OK`: Success
- `400 Bad Request`: Invalid input (e.g., malformed address)
- `404 Not Found`: Resource not found (e.g., account doesn't exist)
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error
- `503 Service Unavailable`: Service not ready (e.g., XRPL disconnected)

## Rate Limiting

The API implements per-IP rate limiting:
- Default: 10 requests per second
- Burst: 20 requests
- Configurable via flags or environment variables

When rate limited, you'll receive a `429 Too Many Requests` response.

## Testing with Swagger UI

1. Start the services:
   ```bash
   docker-compose --profile dev up
   ```

2. Open Swagger UI:
   ```
   http://localhost:8081/swagger
   ```

3. Try out the endpoints interactively

## Development

### Project Structure

```
v2/
├── cmd/
│   └── api/
│       └── main.go              # API entry point
├── internal/
│   ├── api/
│   │   ├── config.go            # Configuration
│   │   ├── server.go            # Server setup
│   │   ├── routes.go            # Route definitions
│   │   ├── handlers.go          # Request handlers
│   │   ├── middleware.go        # Middleware chain
│   │   ├── errors.go            # Error handling
│   │   └── helpers.go           # Helper functions
│   └── validator/
│       └── validator.go         # Input validation
├── api/
│   ├── openapi.yaml             # OpenAPI specification
│   └── README.md                # This file
└── pkg/
    ├── xrpl/                    # XRPL client library
    ├── logger/                  # Logging
    └── config/                  # Config loading
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with race detector
go test -race ./...

# Run with coverage
go test -cover ./...
```

### Building for Production

```bash
# Build optimized binary
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags='-s -w' -o api cmd/api/main.go

# Or use Docker
docker build -t xrpl-api:latest -f Dockerfile .
```

## Deployment

### Docker Compose

See `docker-compose.yml` for complete configuration.

### Kubernetes

Example deployment configuration:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: xrpl-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: xrpl-api
  template:
    metadata:
      labels:
        app: xrpl-api
    spec:
      containers:
      - name: api
        image: xrpl-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENV
          value: "production"
        - name: XRPL_URL
          valueFrom:
            secretKeyRef:
              name: xrpl-config
              key: url
        livenessProbe:
          httpGet:
            path: /health/live
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health/ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
```

## Monitoring

### Health Checks

- **Liveness**: `/health/live` - Always returns 200 if process is running
- **Readiness**: `/health/ready` - Returns 200 only if XRPL connection is healthy

### Logs

Structured JSON logs include:
- Request ID
- HTTP method and path
- Response status and duration
- Error details (when applicable)

Example log entry:
```json
{
  "time": "2026-01-03T10:00:00Z",
  "level": "INFO",
  "msg": "http request",
  "request_id": "1704272400000000000",
  "method": "GET",
  "path": "/xrpl/accounts/rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY/balance",
  "remote_addr": "172.17.0.1:54321",
  "status": 200,
  "duration": "45.2ms"
}
```

## Troubleshooting

### API won't start

Check that:
1. Port 8080 is not already in use
2. XRPL URL is correct and accessible
3. Configuration is valid

### Connection to XRPL fails

- Verify XRPL URL is correct
- Check network connectivity
- Try testnet: `wss://s.altnet.rippletest.net:51233`

### Rate limiting too strict

Adjust rate limit settings:
```bash
./api -rate-limit-rps=100 -rate-limit-burst=200
```

## Support

For issues and questions:
- Open an issue on GitHub
- Check the OpenAPI specification: `/api/openapi.yaml`
- Review the CLI documentation for equivalent commands

## License

MIT License - see LICENSE file for details
