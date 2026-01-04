# V2 Platform - Excise Tax Backend

Lightweight XRPL client and tax calculation services following Unix philosophy and microservice patterns.

## 🚀 Quick Start

### XRPL CLI Tool

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

# Subscribe to real-time transactions
./xrpl-cli subscribe rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY

# Use mainnet with JSON output
./xrpl-cli -network mainnet -json balance rPEPPER7kfTD9w2To4CQk6UCfuHM9c6GDY
```

### Tax CLI Tool

```bash
# Build tax calculator CLI
go build -o tax-cli ./cmd/tax-cli

# Calculate taxes from production data
./tax-cli calculate cmd/tax-cli/example_production.json

# List all available tax rates
./tax-cli rates

# Get specific tax rate
./tax-cli rate beer barrel

# Validate production data
./tax-cli validate cmd/tax-cli/example_production.json

# Use JSON output for scripting
./tax-cli calculate production.json --json
./tax-cli rates --json
```

## 📚 Reusable Libraries

### XRPL Library (`pkg/xrpl`)

WebSocket client library for XRPL interaction with clean functional options API:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/maxfelker/excise-tax-backend/v2/pkg/xrpl"
)

func main() {
    // Create client with options
    client := xrpl.New("wss://s1.ripple.com", 
        xrpl.WithTimeout(10*time.Second),
        xrpl.WithLogger(slog.Default()),
    )
    
    // Connect and query
    if err := client.Connect(context.Background()); err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    account, err := client.GetAccountInfo(ctx, "rN7n7otQDd6FczFgLdSqtcsAUxDkw6fzRH")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Balance: %s XRP\n", account.Balance)
}
```

### Tax Library (`pkg/tax`)

Excise tax calculation library for alcoholic beverages:

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/maxfelker/excise-tax-backend/v2/pkg/tax"
)

func main() {
    // Create calculator with default federal rates
    calc, err := tax.New()
    if err != nil {
        log.Fatal(err)
    }

    // Production data
    data := &tax.ProductionData{
        Items: []tax.ProductionItem{
            {
                ProductType: tax.ProductTypeBeer,
                ProductName: "IPA",
                UnitType:    tax.UnitTypeBarrel,
                Quantity:    100,
                ABV:         6.5,
            },
        },
    }

    // Calculate taxes
    result, err := calc.Calculate(context.Background(), data, "federal", time.Now())
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Total tax: $%.2f\n", result.TotalTaxAmount) // $1800.00
}
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

