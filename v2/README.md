# V2 Platform - Excise Tax Backend

Lightweight XRPL client and services following Unix philosophy.

## Quick Start

### CLI Tool

```bash
cd v2
go mod download
go build -o xrpl-cli ./cmd/xrpl-cli

# Get account balance
./xrpl-cli balance rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY

# Get account info
./xrpl-cli info rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY

# Get transaction details
./xrpl-cli tx C71F385124008A436842B56DEF8196B0621762FBD9464F2510EE9C3D1A3322DA

# Get transaction history
./xrpl-cli history rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY 10

# Use mainnet
./xrpl-cli -network mainnet balance rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY
```

### HTTP API

```bash
# Build and run locally
go build -o api ./cmd/api/main.go
./api -env=development -log-level=debug

# Or run with Docker + Swagger UI
docker-compose --profile dev up

# Test the API
curl http://localhost:8080/health
curl http://localhost:8080/xrpl/accounts/rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY/balance

# Open Swagger UI
open http://localhost:8081/swagger
```

#### API Endpoints

- **Health:** `GET /health`, `/health/ready`, `/health/live`
- **Accounts:** `GET /xrpl/accounts/{address}/balance`, `/info`, `/transactions`
- **Transactions:** `GET /xrpl/transactions/{hash}`, `/{hash}/status`

See [api/README.md](api/README.md) for complete API documentation.

## Testing

### Unit Tests

Run tests that don't require network connectivity:

```bash
# Run all unit tests
go test ./...

# With race detector
go test -race ./...

# With coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Expected output:** Tests complete in under 2 seconds. Current coverage: 26% overall (pkg/config: 78%, pkg/logger: 65%, internal/xrpl: 22%).

### Integration Tests

Run tests against real XRPL testnet (requires internet):

```bash
# Run integration tests
go test -tags=integration ./internal/xrpl -v

# With race detector (recommended)
go test -race -tags=integration ./internal/xrpl -timeout 120s

# Skip integration tests in short mode
go test -tags=integration -short ./internal/xrpl -v
```

**Expected output:** Tests complete in 3-5 seconds. Connects to `wss://s.altnet.rippletest.net:51233` and queries real accounts.

**Note:** Integration tests use the `-tags=integration` flag so they don't run with `go test ./...` by default.

### Docker Testing

```bash
# Unit tests
docker compose run --rm test

# Integration tests
docker compose run --rm test go test -tags=integration ./internal/xrpl -v

# Coverage report
docker compose run --rm test sh -c "go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out"
```

## Docker Development

### Development (hot reload)

```bash
docker compose --profile dev up
```

### Production

```bash
docker compose --profile release up
```

### Cleanup

```bash
# Stop all services
docker compose down

# Remove volumes
docker compose down -v
```

## Environment Setup

Create `.env` file:

```bash
POSTGRES_PASSWORD=your_password
JWT_SECRET=your_secret
SERVER_PORT=8080
LOG_LEVEL=debug
```

## Configuration

### CLI Flags

```bash
./xrpl-cli --help
```

### API Configuration

Set via command-line flags or environment variables:

```bash
# Flags
./api -port=8080 -env=development -xrpl-url=wss://s.altnet.rippletest.net:51233

# Environment variables
export API_PORT=8080
export ENV=development
export XRPL_URL=wss://s.altnet.rippletest.net:51233
export LOG_LEVEL=debug
```

## Requirements

- Go 1.25+
- Docker & Docker Compose (optional)

## Next Steps

See [MIGRATION_PLAN.md](MIGRATION_PLAN.md) for migration roadmap.

