# V2 Platform - Excise Tax Backend

Lightweight XRPL client and services following Unix philosophy.

## Quick Start

### 1. Build CLI

```bash
cd v2
go mod download
go build -o xrpl-cli ./cmd/xrpl-cli
```

### 2. Run CLI

```bash
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

## Testing

### Local

```bash
# Run all tests
go test ./...

# With race detector
go test -race ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Docker

```bash
# Run tests in container
docker compose run --rm test

# Run specific package
docker compose run --rm test go test -v ./internal/xrpl

# View coverage
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

## Help

```bash
./xrpl-cli --help
```

## Requirements

- Go 1.25+
- Docker & Docker Compose (optional)

## Next Steps

See [MIGRATION_PLAN.md](MIGRATION_PLAN.md) for migration roadmap.

