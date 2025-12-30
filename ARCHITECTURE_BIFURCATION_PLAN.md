# Excise Tax Portal Backend Bifurcation - Implementation Plan

**Date:** 2025-12-29
**Status:** Proposed Architecture
**Agent:** Plan Agent Analysis
**Estimated Timeline:** 11 Weeks

---

## Executive Summary

This document outlines a comprehensive plan to bifurcate the Excise Tax Portal backend into two distinct layers:

1. **Backend Backend (Core/CLI Layer)** - Pure business logic as cross-platform CLI tools
2. **Frontend of Backend (API Layer)** - Protocol adapters serving the core over REST/gRPC/WebSocket/GraphQL

This architecture separates transport concerns from business logic, enabling:
- Native command-line tools for automation and scripting
- Multi-protocol API support (REST, gRPC, WebSocket, GraphQL)
- Improved testability and maintainability
- Cross-platform deployment (Windows, macOS, Linux)

---

## Table of Contents

1. [Current Architecture Analysis](#current-architecture-analysis)
2. [Proposed Bifurcated Architecture](#proposed-bifurcated-architecture)
3. [Directory Structure](#directory-structure)
4. [Implementation Plan](#implementation-plan)
5. [Code Patterns and Examples](#code-patterns-and-examples)
6. [Migration Strategy](#migration-strategy)
7. [Benefits and Trade-offs](#benefits-and-trade-offs)
8. [Critical Files](#critical-files)
9. [Recommendations](#recommendations)

---

## Current Architecture Analysis

### Existing Structure

The codebase follows a standard Go microservices pattern with 6 services:

```
backend/
├── cmd/
│   ├── api-gateway/main.go
│   ├── payment-service/main.go
│   ├── tax-service/main.go
│   ├── auth-service/main.go
│   ├── reporting-service/main.go
│   └── notification-service/main.go
│
├── internal/
│   ├── api-gateway/
│   ├── payment/
│   ├── tax/
│   ├── auth/
│   ├── reporting/
│   └── notification/
│
└── pkg/
    ├── database/
    ├── cache/
    ├── logger/
    └── utils/
```

### Key Observations

**Strengths:**
- ✅ Clean separation of handlers, services, repositories, and models within each service
- ✅ Shared packages in `pkg/` (database, cache, logger, utils, validator)
- ✅ XRPL blockchain integration is well-isolated in payment service
- ✅ Dependency injection pattern is already in use

**Current Coupling Issues:**
- ❌ **HTTP Tight Coupling**: Main entry points directly initialize HTTP servers with Gin framework
- ❌ **Mixed Concerns**: Service initialization logic is intertwined with HTTP server setup
- ❌ **Protocol Lock-in**: Business logic services are tightly coupled to REST endpoints
- ❌ **Infrastructure Dependencies**: Each service main.go contains infrastructure initialization

---

## Proposed Bifurcated Architecture

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                     Client Applications                      │
│  (Web Browser, Mobile App, Desktop App, Scripts, CI/CD)     │
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
    REST API            gRPC API          WebSocket
   (Port 8080)         (Port 9090)       (Port 8082)
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │
                ┌───────────▼───────────┐
                │    API Adapters       │
                │   (Thin Layer)        │
                │  REST | gRPC | WS     │
                └───────────┬───────────┘
                            │
                ┌───────────▼───────────┐
                │   Core Business       │
                │   Logic Layer         │
                │  (Pure Go Code)       │
                │  - Payment Service    │
                │  - Tax Service        │
                │  - Auth Service       │
                └───────────┬───────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
        ▼                   ▼                   ▼
   PostgreSQL            Redis              XRPL
   (Database)           (Cache)         (Blockchain)


┌─────────────────────────────────────────────────────────────┐
│                    Parallel Access via CLI                   │
└─────────────────────────────────────────────────────────────┘
                            │
                ┌───────────▼───────────┐
                │   CLI Commands        │
                │   (Cobra)             │
                │  excise-tax-cli       │
                └───────────┬───────────┘
                            │
                            │ (Uses same core)
                            │
                ┌───────────▼───────────┐
                │   Core Business       │
                │   Logic Layer         │
                │  (Shared with API)    │
                └───────────────────────┘
```

### Key Principles

1. **Separation of Concerns**: Business logic is completely isolated from transport protocols
2. **Protocol Agnostic Core**: Core services have zero HTTP/gRPC/WebSocket dependencies
3. **Adapter Pattern**: Thin adapters translate protocol requests to domain calls
4. **Shared Core**: CLI and API both use the same core business logic
5. **Cross-Platform**: CLI compiles to native binaries for Windows, macOS, Linux

---

## Directory Structure

### Complete Proposed Structure

```
backend/
├── cmd/
│   ├── cli/                           # NEW: CLI Entry Points
│   │   ├── excise-tax-cli/            # Main unified CLI
│   │   │   └── main.go
│   │   ├── payment-cli/               # Standalone payment CLI
│   │   │   └── main.go
│   │   ├── tax-cli/                   # Standalone tax CLI
│   │   │   └── main.go
│   │   ├── auth-cli/                  # Standalone auth CLI
│   │   │   └── main.go
│   │   └── reporting-cli/             # Standalone reporting CLI
│   │       └── main.go
│   │
│   └── api/                           # NEW: API Server Entry Points
│       ├── gateway/                   # API Gateway (proxy to all services)
│       │   └── main.go
│       ├── payment-api/               # Payment API server
│       │   └── main.go
│       ├── tax-api/                   # Tax API server
│       │   └── main.go
│       ├── auth-api/                  # Auth API server
│       │   └── main.go
│       └── reporting-api/             # Reporting API server
│           └── main.go
│
├── internal/
│   ├── core/                          # NEW: Pure Business Logic (CLI Core)
│   │   ├── payment/
│   │   │   ├── domain/                # Domain models (pure Go structs)
│   │   │   │   ├── payment.go
│   │   │   │   ├── errors.go
│   │   │   │   └── types.go
│   │   │   ├── service/               # Business logic (no HTTP dependencies)
│   │   │   │   ├── payment_service.go
│   │   │   │   └── wire.go            # Dependency injection
│   │   │   ├── repository/            # Data access interfaces
│   │   │   │   └── payment_repository.go
│   │   │   └── xrpl/                  # XRPL integration (blockchain logic)
│   │   │       ├── client.go
│   │   │       ├── oracle.go
│   │   │       └── monitor.go
│   │   ├── tax/
│   │   │   ├── domain/
│   │   │   ├── service/               # Tax calculation engine
│   │   │   └── repository/
│   │   ├── auth/
│   │   │   ├── domain/
│   │   │   ├── service/               # Auth logic (JWT, password hashing)
│   │   │   └── repository/
│   │   ├── reporting/
│   │   │   ├── domain/
│   │   │   ├── service/
│   │   │   └── repository/
│   │   └── notification/
│   │       ├── domain/
│   │       └── service/
│   │
│   ├── cli/                           # NEW: CLI Framework Layer
│   │   ├── commands/                  # Cobra command definitions
│   │   │   ├── payment/               # Payment CLI commands
│   │   │   │   ├── create.go
│   │   │   │   ├── verify.go
│   │   │   │   ├── list.go
│   │   │   │   └── get.go
│   │   │   ├── tax/                   # Tax CLI commands
│   │   │   │   ├── calculate.go
│   │   │   │   ├── report.go
│   │   │   │   └── rates.go
│   │   │   ├── auth/                  # Auth CLI commands
│   │   │   │   ├── login.go
│   │   │   │   ├── register.go
│   │   │   │   └── refresh.go
│   │   │   └── reporting/             # Reporting CLI commands
│   │   │       ├── generate.go
│   │   │       └── download.go
│   │   ├── output/                    # Output formatters
│   │   │   ├── json.go
│   │   │   ├── table.go
│   │   │   └── yaml.go
│   │   ├── input/                     # Input handlers
│   │   │   ├── flags.go
│   │   │   ├── files.go
│   │   │   └── stdin.go
│   │   └── config/                    # CLI configuration
│   │       └── config.go
│   │
│   ├── adapters/                      # NEW: Protocol Adapters (API Layer)
│   │   ├── rest/                      # REST API adapters
│   │   │   ├── payment/
│   │   │   │   ├── handler.go         # HTTP handlers (thin layer)
│   │   │   │   ├── dto.go             # Request/Response DTOs
│   │   │   │   ├── router.go          # Route definitions
│   │   │   │   └── mapper.go          # DTO <-> Domain mapping
│   │   │   ├── tax/
│   │   │   │   ├── handler.go
│   │   │   │   ├── dto.go
│   │   │   │   └── router.go
│   │   │   ├── auth/
│   │   │   │   ├── handler.go
│   │   │   │   ├── dto.go
│   │   │   │   └── router.go
│   │   │   └── reporting/
│   │   │       ├── handler.go
│   │   │       ├── dto.go
│   │   │       └── router.go
│   │   │
│   │   ├── grpc/                      # gRPC adapters
│   │   │   ├── payment/
│   │   │   │   ├── server.go          # gRPC server implementation
│   │   │   │   └── mapper.go          # Proto <-> Domain mapping
│   │   │   ├── tax/
│   │   │   │   ├── server.go
│   │   │   │   └── mapper.go
│   │   │   ├── auth/
│   │   │   │   ├── server.go
│   │   │   │   └── mapper.go
│   │   │   └── reporting/
│   │   │       ├── server.go
│   │   │       └── mapper.go
│   │   │
│   │   ├── websocket/                 # WebSocket adapters
│   │   │   ├── payment/               # Real-time payment updates
│   │   │   │   └── handler.go
│   │   │   └── notification/          # Real-time notifications
│   │   │       └── handler.go
│   │   │
│   │   └── graphql/                   # GraphQL adapters (optional)
│   │       ├── schema/                # GraphQL schema definitions
│   │       │   ├── payment.graphql
│   │   │       │   ├── tax.graphql
│   │       │   └── auth.graphql
│   │       ├── resolvers/             # GraphQL resolvers
│   │       │   ├── payment.go
│   │       │   ├── tax.go
│   │       │   └── auth.go
│   │       └── server.go              # GraphQL server
│   │
│   ├── gateway/                       # API Gateway (routes to adapters)
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── cors.go
│   │   │   ├── logging.go
│   │   │   └── ratelimit.go
│   │   ├── proxy/
│   │   │   └── proxy.go
│   │   └── router/
│   │       └── router.go
│   │
│   └── infrastructure/                # Infrastructure implementations
│       ├── database/                  # Database implementations
│       │   ├── payment_repository_impl.go
│       │   ├── tax_repository_impl.go
│       │   └── auth_repository_impl.go
│       ├── cache/                     # Cache implementations
│       │   └── redis_cache_impl.go
│       ├── queue/                     # Queue implementations
│       │   └── rabbitmq_impl.go
│       └── xrpl/                      # XRPL client implementations
│           ├── client_impl.go
│           └── monitor_impl.go
│
├── pkg/                               # Shared packages (unchanged)
│   ├── database/
│   │   └── postgres.go
│   ├── cache/
│   │   └── redis.go
│   ├── logger/
│   │   └── logger.go
│   ├── config/
│   │   └── config.go
│   ├── errors/
│   │   └── errors.go
│   ├── validator/
│   │   └── validator.go
│   └── utils/
│       ├── jwt.go
│       └── crypto.go
│
├── proto/                             # NEW: Protocol Buffer definitions
│   ├── payment/
│   │   └── v1/
│   │       └── payment.proto
│   ├── tax/
│   │   └── v1/
│   │       └── tax.proto
│   ├── auth/
│   │   └── v1/
│   │       └── auth.proto
│   └── reporting/
│       └── v1/
│           └── reporting.proto
│
├── build/                             # NEW: Build artifacts directory
│   ├── cli/                           # CLI binaries
│   │   ├── excise-tax-cli-linux-amd64
│   │   ├── excise-tax-cli-darwin-amd64
│   │   ├── excise-tax-cli-darwin-arm64
│   │   ├── excise-tax-cli-windows-amd64.exe
│   │   └── checksums.txt
│   └── api/                           # API server binaries
│       ├── payment-api
│       ├── tax-api
│       └── auth-api
│
├── configs/                           # Configuration (enhanced)
│   ├── cli/                           # CLI configs
│   │   ├── excise-tax-cli.yaml.example
│   │   └── README.md
│   └── api/                           # API server configs
│       ├── payment-api.yaml
│       ├── tax-api.yaml
│       └── auth-api.yaml
│
├── scripts/                           # Build and deployment scripts
│   ├── build-cli.sh                   # Cross-compile CLI for all platforms
│   ├── build-api.sh                   # Build API servers
│   ├── install-cli.sh                 # Install CLI to system PATH
│   ├── generate-proto.sh              # Generate gRPC code from proto files
│   └── test-all.sh                    # Run all tests
│
└── deployments/
    ├── docker/
    │   ├── Dockerfile.cli             # CLI container
    │   ├── Dockerfile.api-gateway     # API Gateway
    │   ├── Dockerfile.payment-api     # Payment API
    │   ├── Dockerfile.tax-api         # Tax API
    │   ├── docker-compose.bifurcated.yml
    │   └── docker-compose.dev.yml
    └── kubernetes/
        ├── cli/                       # CLI as jobs/cronjobs
        │   ├── payment-monitor.yaml
        │   └── tax-calculator.yaml
        └── api/                       # API deployments
            ├── payment-api.yaml
            ├── tax-api.yaml
            └── gateway.yaml
```

---

## Implementation Plan

### Phase 1: Core Business Logic Extraction (Weeks 1-3)

**Goal**: Extract pure business logic from existing services into `internal/core/`

#### Week 1: Payment Service Core

**Tasks:**
1. Create `internal/core/payment/domain/` with pure domain models
2. Move `internal/payment/service/payment_service.go` → `internal/core/payment/service/`
3. Remove all HTTP dependencies (Gin, gin.Context)
4. Replace with pure `context.Context`
5. Create repository interfaces in `internal/core/payment/repository/`
6. Refactor XRPL integration to be HTTP-agnostic

**Deliverables:**
- `internal/core/payment/domain/payment.go` - Pure domain models
- `internal/core/payment/service/payment_service.go` - Business logic with zero HTTP imports
- `internal/core/payment/repository/payment_repository.go` - Interface definitions

**Success Criteria:**
- [ ] Core payment service has zero Gin/HTTP imports
- [ ] Unit tests run without HTTP mocks
- [ ] Existing API handler can call core service

#### Week 2: Tax and Auth Services Core

**Tasks:**
1. Extract tax calculation engine to `internal/core/tax/`
2. Extract auth logic (JWT, password hashing) to `internal/core/auth/`
3. Create domain models for both services
4. Remove HTTP-specific error handling

**Deliverables:**
- Tax calculation as pure functions
- Auth service with no session management tied to HTTP cookies

#### Week 3: Reporting and Notification Services Core

**Tasks:**
1. Extract reporting service to `internal/core/reporting/`
2. Extract notification service to `internal/core/notification/`
3. Create infrastructure implementations in `internal/infrastructure/`

**Deliverables:**
- All 6 core services extracted
- Infrastructure layer separated from core logic

---

### Phase 2: CLI Framework Implementation (Weeks 3-5)

**Goal**: Build command-line interface using Cobra and Viper

#### Week 3-4: CLI Framework Setup

**Framework Choice: Cobra + Viper**

**Rationale:**
- Industry standard (used by kubectl, gh, hugo)
- Excellent documentation
- Viper provides configuration management (env vars, config files, flags)

**Tasks:**
1. Install dependencies:
   ```bash
   go get github.com/spf13/cobra@latest
   go get github.com/spf13/viper@latest
   ```

2. Create CLI command structure:
   ```
   excise-tax-cli
   ├── payment
   │   ├── create
   │   ├── verify
   │   ├── list
   │   └── get
   ├── tax
   │   ├── calculate
   │   ├── report-create
   │   ├── report-submit
   │   └── rates
   ├── auth
   │   ├── login
   │   ├── register
   │   └── refresh
   └── config
       ├── init
       └── show
   ```

3. Implement core commands for payment service

**Deliverables:**
- `cmd/cli/excise-tax-cli/main.go` - CLI entry point
- `internal/cli/commands/payment/create.go` - Payment creation command
- `internal/cli/config/config.go` - Configuration management

#### Week 5: CLI Output Formatting and Testing

**Tasks:**
1. Implement output formatters (JSON, Table, YAML)
2. Create configuration file support (~/.excise-tax/excise-tax-cli.yaml)
3. Add input handling (flags, files, stdin)
4. Write CLI integration tests

**Deliverables:**
- Multi-format output support
- Config file management
- CLI test suite

---

### Phase 3: API Adapter Layer (Weeks 5-7)

**Goal**: Create thin HTTP/gRPC/WebSocket handlers that delegate to core services

#### Week 5-6: REST Adapters

**Tasks:**
1. Create `internal/adapters/rest/payment/` with HTTP handlers
2. Create DTOs (Data Transfer Objects) for request/response
3. Implement DTO ↔ Domain conversion
4. Create routers for each service

**Pattern:**

```go
// internal/adapters/rest/payment/handler.go
type PaymentRESTHandler struct {
    coreService *service.PaymentService
    logger      *zap.Logger
}

func (h *PaymentRESTHandler) CreateXRPLPayment(c *gin.Context) {
    // Parse HTTP DTO
    var dto CreateXRPLPaymentDTO
    if err := c.ShouldBindJSON(&dto); err != nil {
        c.JSON(400, ErrorResponse{Message: err.Error()})
        return
    }

    // Convert DTO → Domain
    req := dto.ToDomain()

    // Call core service
    payment, err := h.coreService.CreateXRPLPayment(c.Request.Context(), req)
    if err != nil {
        c.JSON(mapDomainErrorToHTTP(err).StatusCode, err)
        return
    }

    // Convert Domain → DTO
    c.JSON(201, PaymentDTOFromDomain(payment))
}
```

**Deliverables:**
- REST handlers for all services
- DTO definitions and mappers
- Router configurations

#### Week 6-7: gRPC and WebSocket Adapters

**Tasks:**
1. Define Protocol Buffer schemas in `proto/`
2. Generate gRPC code: `protoc --go_out=. --go-grpc_out=. proto/**/*.proto`
3. Implement gRPC servers in `internal/adapters/grpc/`
4. Create WebSocket handlers for real-time updates
5. (Optional) Add GraphQL adapter

**Deliverables:**
- gRPC servers for all services
- WebSocket handlers for payment updates
- Proto definitions versioned (v1)

---

### Phase 4: Shared Code Management (Week 7)

**Goal**: Ensure no code duplication between CLI and API layers

#### Dependency Injection Pattern

**Create service factories:**

```go
// internal/core/payment/wire.go
package payment

func NewPaymentService(
    ctx context.Context,
    db *database.PostgresDB,
    cache *cache.RedisClient,
    xrplClient *xrpl.Client,
    logger *logger.Logger,
) (*service.PaymentService, error) {
    // Create repositories
    paymentRepo := database.NewPaymentRepository(db)

    // Create XRPL components
    oracle := xrpl.NewPriceOracle(...)
    monitor := xrpl.NewMonitorService(xrplClient, logger)
    processor := xrpl.NewPaymentProcessor(xrplClient, oracle, monitor, logger)

    // Create service
    return service.NewPaymentService(
        paymentRepo,
        cache,
        xrplClient,
        oracle,
        monitor,
        processor,
        logger,
    ), nil
}
```

**Both CLI and API use same factory:**

```go
// CLI uses it:
paymentService, err := payment.NewPaymentService(ctx, db, cache, xrplClient, logger)
cmd := commands.NewPaymentCommand(paymentService)

// API uses it:
paymentService, err := payment.NewPaymentService(ctx, db, cache, xrplClient, logger)
restHandler := rest.NewPaymentRESTHandler(paymentService, logger)
grpcServer := grpc.NewPaymentGRPCServer(paymentService)
```

**Deliverables:**
- Service factory functions for all core services
- Shared infrastructure initialization
- Documentation on dependency injection pattern

---

### Phase 5: Migration Strategy (Weeks 8-10)

**Goal**: Migrate from current monolithic services to bifurcated architecture with zero downtime

#### Strategy: Strangler Fig Pattern

**Week 8: Parallel Deployment**
- Deploy new API adapters alongside existing services
- Both old and new endpoints active
- Route 10% of traffic to new endpoints
- Monitor metrics (latency, error rates, throughput)

**Week 9: Progressive Rollout**
- Increase traffic to 50%
- A/B testing between old and new
- Performance comparison
- Rollback capability ready

**Week 10: Complete Migration**
- Route 100% traffic to new architecture
- Deprecate old endpoints (keep for 1 month for rollback)
- Remove old code after validation period

#### Database Migration

**No schema changes needed!** Repository implementations remain the same.

```go
// internal/infrastructure/database/payment_repository_impl.go
// Implements core/payment/repository.PaymentRepository interface
type PaymentRepositoryImpl struct {
    db *database.PostgresDB
}

func (r *PaymentRepositoryImpl) Create(ctx context.Context, payment *domain.Payment) error {
    // Same SQL queries, just uses domain.Payment instead of model.Payment
    query := `INSERT INTO payments (manufacturer_id, amount_usd, ...) VALUES ($1, $2, ...)`
    // ...
}
```

#### Docker Deployment Changes

**Current** (`docker-compose.yml`):
```yaml
services:
  payment-service:
    build:
      context: .
      dockerfile: deployments/docker/Dockerfile.payment-service
    ports:
      - "8081:8081"
```

**Bifurcated** (`docker-compose.bifurcated.yml`):
```yaml
services:
  # API Layer
  payment-api:
    build:
      context: .
      dockerfile: deployments/docker/Dockerfile.payment-api
    ports:
      - "8081:8081"    # REST API
      - "9091:9091"    # gRPC
    depends_on:
      - postgres
      - redis
    environment:
      - PAYMENT_SERVICE_MODE=api

  # CLI Layer (for scheduled jobs)
  payment-cli-worker:
    build:
      context: .
      dockerfile: deployments/docker/Dockerfile.cli
    command: ["excise-tax-cli", "payment", "monitor", "--interval=5m"]
    depends_on:
      - postgres
      - redis
    restart: always
    environment:
      - PAYMENT_SERVICE_MODE=cli
```

**Multi-stage Dockerfile for CLI:**

```dockerfile
# deployments/docker/Dockerfile.cli
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /excise-tax-cli ./cmd/cli/excise-tax-cli

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /excise-tax-cli /usr/local/bin/excise-tax-cli

ENTRYPOINT ["excise-tax-cli"]
```

**Deliverables:**
- Parallel deployment configuration
- Traffic routing scripts
- Monitoring dashboards
- Rollback procedures documented

---

### Phase 6: Cross-Platform CLI Distribution (Week 11)

**Goal**: Build and distribute CLI binaries for all platforms

#### Build System

```bash
#!/bin/bash
# scripts/build-cli.sh

VERSION=${VERSION:-"v1.0.0"}
PLATFORMS=("windows/amd64" "darwin/amd64" "darwin/arm64" "linux/amd64" "linux/arm64")

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS=${PLATFORM%/*}
    GOARCH=${PLATFORM#*/}
    OUTPUT_NAME="excise-tax-cli-${GOOS}-${GOARCH}-${VERSION}"

    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME+=".exe"
    fi

    echo "Building for $GOOS/$GOARCH..."
    GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "-X main.Version=$VERSION -s -w" \
        -o "build/cli/${OUTPUT_NAME}" \
        ./cmd/cli/excise-tax-cli
done

# Generate checksums
cd build/cli
sha256sum * > checksums.txt

echo "Build complete! Binaries in build/cli/"
```

#### Installation Script

```bash
#!/bin/bash
# scripts/install-cli.sh

set -e

OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux*)     OS=linux;;
    Darwin*)    OS=darwin;;
    MINGW*|MSYS*|CYGWIN*) OS=windows;;
    *)          echo "Unsupported OS: $OS"; exit 1;;
esac

case "$ARCH" in
    x86_64)     ARCH=amd64;;
    arm64|aarch64) ARCH=arm64;;
    *)          echo "Unsupported architecture: $ARCH"; exit 1;;
esac

VERSION=${VERSION:-"latest"}
BINARY_NAME="excise-tax-cli-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    BINARY_NAME+=".exe"
fi

DOWNLOAD_URL="https://github.com/YOUR_ORG/excise-tax-portal/releases/${VERSION}/download/${BINARY_NAME}"

echo "Downloading excise-tax-cli for $OS/$ARCH..."
curl -L "$DOWNLOAD_URL" -o /tmp/excise-tax-cli

chmod +x /tmp/excise-tax-cli

# Determine install directory
if [ "$OS" = "windows" ]; then
    INSTALL_DIR="$HOME/bin"
else
    INSTALL_DIR="/usr/local/bin"
fi

mkdir -p "$INSTALL_DIR"
mv /tmp/excise-tax-cli "$INSTALL_DIR/excise-tax-cli"

echo "✓ excise-tax-cli installed to $INSTALL_DIR/excise-tax-cli"
echo "Run 'excise-tax-cli --help' to get started"
```

#### Package Managers

**Homebrew** (macOS/Linux):
```ruby
# Formula/excise-tax-cli.rb
class ExciseTaxCli < Formula
  desc "CLI for Excise Tax Portal"
  homepage "https://github.com/YOUR_ORG/excise-tax-portal"
  version "1.0.0"

  if OS.mac? && Hardware::CPU.arm?
    url "https://github.com/YOUR_ORG/excise-tax-portal/releases/download/v1.0.0/excise-tax-cli-darwin-arm64.tar.gz"
    sha256 "..."
  elsif OS.mac?
    url "https://github.com/YOUR_ORG/excise-tax-portal/releases/download/v1.0.0/excise-tax-cli-darwin-amd64.tar.gz"
    sha256 "..."
  elsif OS.linux?
    url "https://github.com/YOUR_ORG/excise-tax-portal/releases/download/v1.0.0/excise-tax-cli-linux-amd64.tar.gz"
    sha256 "..."
  end

  def install
    bin.install "excise-tax-cli"
  end

  test do
    system "#{bin}/excise-tax-cli", "--version"
  end
end
```

**Chocolatey** (Windows):
```xml
<?xml version="1.0"?>
<package xmlns="http://schemas.microsoft.com/packaging/2015/06/nuspec.xsd">
  <metadata>
    <id>excise-tax-cli</id>
    <version>1.0.0</version>
    <title>Excise Tax CLI</title>
    <authors>State Government</authors>
    <description>Command-line interface for Excise Tax Portal</description>
    <projectUrl>https://github.com/YOUR_ORG/excise-tax-portal</projectUrl>
    <licenseUrl>https://github.com/YOUR_ORG/excise-tax-portal/blob/main/LICENSE</licenseUrl>
    <requireLicenseAcceptance>false</requireLicenseAcceptance>
  </metadata>
  <files>
    <file src="tools\**" target="tools" />
  </files>
</package>
```

**Deliverables:**
- Cross-platform binaries for 5 platforms
- Installation scripts
- Package manager configurations
- Release automation (GitHub Actions)

---

## Code Patterns and Examples

### Pattern 1: Domain Model (Core)

Pure business objects with no external dependencies:

```go
// internal/core/payment/domain/payment.go
package domain

import "time"

// Payment is the core domain model
type Payment struct {
    ID             int64
    ManufacturerID int64
    AmountUSD      float64
    XRPAmount      float64
    Status         PaymentStatus
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

type PaymentStatus string

const (
    PaymentStatusPending   PaymentStatus = "pending"
    PaymentStatusCompleted PaymentStatus = "completed"
    PaymentStatusFailed    PaymentStatus = "failed"
    PaymentStatusExpired   PaymentStatus = "expired"
)

// Business logic methods
func (p *Payment) CanBeCancelled() bool {
    return p.Status == PaymentStatusPending
}

func (p *Payment) IsExpired() bool {
    return p.Status == PaymentStatusExpired
}
```

### Pattern 2: Service Layer (Core)

Business logic with zero protocol dependencies:

```go
// internal/core/payment/service/payment_service.go
package service

import (
    "context"
    "fmt"

    "github.com/excise-tax-portal/backend/internal/core/payment/domain"
    "github.com/excise-tax-portal/backend/internal/core/payment/repository"
    "go.uber.org/zap"
)

type PaymentService struct {
    repo       repository.PaymentRepository
    xrplClient XRPLClient
    logger     *zap.Logger
}

func NewPaymentService(
    repo repository.PaymentRepository,
    xrplClient XRPLClient,
    logger *zap.Logger,
) *PaymentService {
    return &PaymentService{
        repo:       repo,
        xrplClient: xrplClient,
        logger:     logger,
    }
}

// CreateXRPLPayment - Pure business logic (no HTTP)
func (s *PaymentService) CreateXRPLPayment(
    ctx context.Context,
    req *domain.CreateXRPLPaymentRequest,
) (*domain.XRPLPayment, error) {
    // Validate
    if err := s.validatePaymentRequest(req); err != nil {
        return nil, domain.NewValidationError(err.Error())
    }

    // Get XRP/USD rate
    rate, err := s.xrplClient.GetExchangeRate(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get exchange rate: %w", err)
    }

    // Calculate XRP amount
    xrpAmount := req.AmountUSD / rate

    // Create payment
    payment := &domain.Payment{
        ManufacturerID: req.ManufacturerID,
        AmountUSD:      req.AmountUSD,
        XRPAmount:      xrpAmount,
        Status:         domain.PaymentStatusPending,
    }

    // Save to database
    if err := s.repo.Create(ctx, payment); err != nil {
        return nil, fmt.Errorf("failed to create payment: %w", err)
    }

    s.logger.Info("payment created",
        zap.Int64("payment_id", payment.ID),
        zap.Float64("amount_usd", payment.AmountUSD),
    )

    return &domain.XRPLPayment{
        PaymentID:     payment.ID,
        AmountUSD:     payment.AmountUSD,
        XRPAmount:     payment.XRPAmount,
        ExchangeRate:  rate,
    }, nil
}
```

### Pattern 3: CLI Command

Command that uses core service:

```go
// internal/cli/commands/payment/create.go
package payment

import (
    "context"
    "encoding/json"
    "fmt"
    "os"

    "github.com/spf13/cobra"
    "github.com/excise-tax-portal/backend/internal/core/payment/domain"
    "github.com/excise-tax-portal/backend/internal/core/payment/service"
)

type CreateCommand struct {
    service *service.PaymentService
}

func NewCreateCommand(svc *service.PaymentService) *cobra.Command {
    cmd := &CreateCommand{service: svc}

    cobraCmd := &cobra.Command{
        Use:   "create",
        Short: "Create a new XRPL payment",
        Long: `Create a new XRPL payment request for a manufacturer.

Example:
  excise-tax-cli payment create --manufacturer-id 123 --amount 1000.00 --description "Beer excise tax Q1 2024"`,
        RunE:  cmd.run,
    }

    // Define flags
    cobraCmd.Flags().Int64("manufacturer-id", 0, "Manufacturer ID (required)")
    cobraCmd.Flags().Float64("amount", 0, "Amount in USD (required)")
    cobraCmd.Flags().String("description", "", "Payment description")
    cobraCmd.Flags().String("output", "json", "Output format (json|table|yaml)")

    cobraCmd.MarkFlagRequired("manufacturer-id")
    cobraCmd.MarkFlagRequired("amount")

    return cobraCmd
}

func (c *CreateCommand) run(cmd *cobra.Command, args []string) error {
    ctx := context.Background()

    // Parse flags
    manufacturerID, _ := cmd.Flags().GetInt64("manufacturer-id")
    amount, _ := cmd.Flags().GetFloat64("amount")
    description, _ := cmd.Flags().GetString("description")
    outputFormat, _ := cmd.Flags().GetString("output")

    // Create domain request
    req := &domain.CreateXRPLPaymentRequest{
        ManufacturerID: manufacturerID,
        AmountUSD:      amount,
        Description:    description,
    }

    // Call core service
    payment, err := c.service.CreateXRPLPayment(ctx, req)
    if err != nil {
        return fmt.Errorf("failed to create payment: %w", err)
    }

    // Format output
    return c.formatOutput(payment, outputFormat)
}

func (c *CreateCommand) formatOutput(payment *domain.XRPLPayment, format string) error {
    switch format {
    case "json":
        encoder := json.NewEncoder(os.Stdout)
        encoder.SetIndent("", "  ")
        return encoder.Encode(payment)
    case "table":
        fmt.Printf("Payment ID:       %d\n", payment.PaymentID)
        fmt.Printf("Amount USD:       $%.2f\n", payment.AmountUSD)
        fmt.Printf("XRP Amount:       %.6f XRP\n", payment.XRPAmount)
        fmt.Printf("Exchange Rate:    $%.4f\n", payment.ExchangeRate)
        fmt.Printf("Destination:      %s\n", payment.DestinationAddress)
        return nil
    default:
        return fmt.Errorf("unsupported format: %s", format)
    }
}
```

### Pattern 4: REST Adapter

Thin HTTP layer that delegates to core:

```go
// internal/adapters/rest/payment/handler.go
package payment

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/excise-tax-portal/backend/internal/core/payment/service"
    "github.com/excise-tax-portal/backend/internal/core/payment/domain"
    "go.uber.org/zap"
)

type PaymentRESTHandler struct {
    coreService *service.PaymentService
    logger      *zap.Logger
}

func NewPaymentRESTHandler(svc *service.PaymentService, logger *zap.Logger) *PaymentRESTHandler {
    return &PaymentRESTHandler{
        coreService: svc,
        logger:      logger,
    }
}

func (h *PaymentRESTHandler) CreateXRPLPayment(c *gin.Context) {
    // Parse HTTP DTO
    var dto CreateXRPLPaymentDTO
    if err := c.ShouldBindJSON(&dto); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Code:    "INVALID_REQUEST",
            Message: "Invalid request body",
            Details: err.Error(),
        })
        return
    }

    // DTO → Domain conversion
    req := dto.ToDomain()

    // Call core service (pure business logic)
    payment, err := h.coreService.CreateXRPLPayment(c.Request.Context(), req)
    if err != nil {
        // Domain error → HTTP error conversion
        httpErr := mapDomainErrorToHTTP(err)
        c.JSON(httpErr.StatusCode, httpErr)
        return
    }

    // Domain → DTO conversion
    responseDTO := PaymentResponseDTOFromDomain(payment)
    c.JSON(http.StatusCreated, responseDTO)
}
```

```go
// internal/adapters/rest/payment/dto.go
package payment

import "github.com/excise-tax-portal/backend/internal/core/payment/domain"

// CreateXRPLPaymentDTO is the HTTP request DTO
type CreateXRPLPaymentDTO struct {
    ManufacturerID int64   `json:"manufacturer_id" binding:"required"`
    AmountUSD      float64 `json:"amount_usd" binding:"required,gt=0"`
    Description    string  `json:"description"`
}

// ToDomain converts HTTP DTO to domain model
func (dto *CreateXRPLPaymentDTO) ToDomain() *domain.CreateXRPLPaymentRequest {
    return &domain.CreateXRPLPaymentRequest{
        ManufacturerID: dto.ManufacturerID,
        AmountUSD:      dto.AmountUSD,
        Description:    dto.Description,
    }
}

// PaymentResponseDTO is the HTTP response DTO
type PaymentResponseDTO struct {
    PaymentID          int64   `json:"payment_id"`
    ManufacturerID     int64   `json:"manufacturer_id"`
    AmountUSD          float64 `json:"amount_usd"`
    XRPAmount          float64 `json:"xrp_amount"`
    DestinationAddress string  `json:"destination_address"`
    Status             string  `json:"status"`
}

// FromDomain converts domain model to HTTP DTO
func PaymentResponseDTOFromDomain(payment *domain.XRPLPayment) *PaymentResponseDTO {
    return &PaymentResponseDTO{
        PaymentID:          payment.PaymentID,
        ManufacturerID:     payment.ManufacturerID,
        AmountUSD:          payment.AmountUSD,
        XRPAmount:          payment.XRPAmount,
        DestinationAddress: payment.DestinationAddress,
        Status:             string(payment.Status),
    }
}

// mapDomainErrorToHTTP converts domain errors to HTTP errors
func mapDomainErrorToHTTP(err error) *ErrorResponse {
    switch e := err.(type) {
    case *domain.ValidationError:
        return &ErrorResponse{
            Code:       "VALIDATION_ERROR",
            Message:    e.Error(),
            StatusCode: 400,
        }
    case *domain.NotFoundError:
        return &ErrorResponse{
            Code:       "NOT_FOUND",
            Message:    e.Error(),
            StatusCode: 404,
        }
    default:
        return &ErrorResponse{
            Code:       "INTERNAL_ERROR",
            Message:    "An internal error occurred",
            StatusCode: 500,
        }
    }
}

type ErrorResponse struct {
    Code       string `json:"code"`
    Message    string `json:"message"`
    Details    string `json:"details,omitempty"`
    StatusCode int    `json:"-"`
}
```

### Pattern 5: gRPC Adapter

Protocol buffer definitions and gRPC server:

```protobuf
// proto/payment/v1/payment.proto
syntax = "proto3";

package payment.v1;
option go_package = "github.com/excise-tax-portal/backend/proto/payment/v1;paymentv1";

import "google/protobuf/timestamp.proto";

service PaymentService {
  rpc CreateXRPLPayment(CreateXRPLPaymentRequest) returns (XRPLPaymentResponse);
  rpc GetPayment(GetPaymentRequest) returns (PaymentResponse);
  rpc ListPayments(ListPaymentsRequest) returns (ListPaymentsResponse);
  rpc VerifyPayment(VerifyPaymentRequest) returns (VerifyPaymentResponse);
}

message CreateXRPLPaymentRequest {
  int64 manufacturer_id = 1;
  double amount_usd = 2;
  string description = 3;
}

message XRPLPaymentResponse {
  int64 payment_id = 1;
  int64 manufacturer_id = 2;
  double amount_usd = 3;
  double xrp_amount = 4;
  string destination_address = 5;
  uint32 destination_tag = 6;
  string status = 7;
  string qr_code = 8;
  double exchange_rate = 9;
  google.protobuf.Timestamp created_at = 10;
}

message GetPaymentRequest {
  int64 payment_id = 1;
}

message PaymentResponse {
  int64 payment_id = 1;
  int64 manufacturer_id = 2;
  double amount_usd = 3;
  double xrp_amount = 4;
  string status = 5;
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp updated_at = 7;
}

message ListPaymentsRequest {
  int64 manufacturer_id = 1;
  int32 page_size = 2;
  int32 page_number = 3;
}

message ListPaymentsResponse {
  repeated PaymentResponse payments = 1;
  int32 total_count = 2;
  int32 page_number = 3;
  int32 page_size = 4;
}

message VerifyPaymentRequest {
  int64 payment_id = 1;
  string transaction_hash = 2;
}

message VerifyPaymentResponse {
  bool verified = 1;
  string status = 2;
  string message = 3;
}
```

```go
// internal/adapters/grpc/payment/server.go
package payment

import (
    "context"

    paymentv1 "github.com/excise-tax-portal/backend/proto/payment/v1"
    "github.com/excise-tax-portal/backend/internal/core/payment/service"
    "github.com/excise-tax-portal/backend/internal/core/payment/domain"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "go.uber.org/zap"
)

type PaymentGRPCServer struct {
    paymentv1.UnimplementedPaymentServiceServer
    coreService *service.PaymentService
    logger      *zap.Logger
}

func NewPaymentGRPCServer(svc *service.PaymentService, logger *zap.Logger) *PaymentGRPCServer {
    return &PaymentGRPCServer{
        coreService: svc,
        logger:      logger,
    }
}

func (s *PaymentGRPCServer) CreateXRPLPayment(
    ctx context.Context,
    req *paymentv1.CreateXRPLPaymentRequest,
) (*paymentv1.XRPLPaymentResponse, error) {
    // Proto → Domain conversion
    domainReq := &domain.CreateXRPLPaymentRequest{
        ManufacturerID: req.ManufacturerId,
        AmountUSD:      req.AmountUsd,
        Description:    req.Description,
    }

    // Call core service
    payment, err := s.coreService.CreateXRPLPayment(ctx, domainReq)
    if err != nil {
        return nil, mapDomainErrorToGRPC(err)
    }

    // Domain → Proto conversion
    return &paymentv1.XRPLPaymentResponse{
        PaymentId:          payment.PaymentID,
        ManufacturerId:     payment.ManufacturerID,
        AmountUsd:          payment.AmountUSD,
        XrpAmount:          payment.XRPAmount,
        DestinationAddress: payment.DestinationAddress,
        DestinationTag:     payment.DestinationTag,
        Status:             string(payment.Status),
        QrCode:             payment.QRCode,
        ExchangeRate:       payment.ExchangeRate,
    }, nil
}

func mapDomainErrorToGRPC(err error) error {
    switch err.(type) {
    case *domain.ValidationError:
        return status.Error(codes.InvalidArgument, err.Error())
    case *domain.NotFoundError:
        return status.Error(codes.NotFound, err.Error())
    case *domain.UnauthorizedError:
        return status.Error(codes.Unauthenticated, err.Error())
    default:
        return status.Error(codes.Internal, "internal server error")
    }
}
```

---

## Benefits and Trade-offs

### Benefits

#### 1. CLI Benefits
- **Automation**: Scripts can call `excise-tax-cli` for batch operations
- **CI/CD Integration**: Use CLI in pipelines for automated testing and deployment
- **Developer Experience**: Test business logic without running HTTP server
- **Cross-Platform**: Single binary works on Windows, macOS, Linux
- **No HTTP Overhead**: Direct database access, faster for bulk operations
- **Debugging**: Easier to debug pure Go code vs HTTP requests
- **Scripting**: Shell scripts can automate complex workflows

#### 2. API Layer Benefits
- **Protocol Flexibility**: Support REST, gRPC, WebSocket, GraphQL simultaneously
- **Frontend Agnostic**: Any client (web, mobile, desktop) can use any protocol
- **Future-Proof**: Easy to add new protocols (MQTT, Server-Sent Events)
- **Versioning**: API v1, v2 can coexist while core logic is shared
- **Testing**: Mock adapters easily, test core logic independently
- **Performance**: Protocol-specific optimizations without touching core

#### 3. Architectural Benefits
- **Separation of Concerns**: Business logic completely isolated from transport
- **Testability**: Core services testable without HTTP mocks
- **Maintainability**: Changes to protocols don't affect business logic
- **Reusability**: Core services used by both CLI and API
- **Security**: Clear boundaries, easier to audit and secure
- **Performance**: Optimize each layer independently
- **Team Scaling**: Frontend team works on adapters, backend team on core

### Trade-offs

#### 1. Increased Complexity
- **More Code**: DTOs for each protocol, converters, adapters (estimated 30% more code)
- **Learning Curve**: Team needs to understand bifurcated architecture
- **Boilerplate**: DTO ↔ Domain conversions can be repetitive
- **Mitigation**:
  - Use code generation (protoc for gRPC, oapi-codegen for REST)
  - Create reusable conversion utilities
  - Clear documentation with examples

#### 2. Development Time
- **Initial Investment**: 11 weeks for full migration
- **Dual Maintenance**: During migration, maintain both old and new
- **Training Required**: Team needs architectural training
- **Mitigation**:
  - Gradual rollout (strangler fig pattern)
  - Start with one service (payment) as POC
  - Pair programming for knowledge transfer

#### 3. Operational Overhead
- **More Deployables**: CLI binaries + multiple API servers
- **Monitoring**: Need to monitor adapters separately from core
- **Deployment Complexity**: More Docker containers/Kubernetes pods
- **Mitigation**:
  - Docker Compose templates for easy local development
  - Kubernetes manifests for production
  - Observability built-in (structured logging, metrics)

#### 4. Debugging Complexity
- **Cross-Layer**: Bugs might span adapter + core layers
- **Trace IDs**: Need correlation across layers
- **Mitigation**:
  - Structured logging with trace IDs
  - OpenTelemetry for distributed tracing
  - Clear error propagation patterns

---

## Critical Files

### Phase 1: Core Extraction (Highest Priority)

1. **backend/internal/payment/service/payment_service.go**
   - **Why**: Core payment business logic with XRPL integration (most complex)
   - **Impact**: High - blockchain integration is critical
   - **Refactoring**: Remove HTTP dependencies, create pure domain service
   - **Estimated Effort**: 3-5 days

2. **backend/internal/tax/service/tax_service.go**
   - **Why**: Tax calculation engine
   - **Impact**: Medium - critical for tax logic but simpler than payment
   - **Refactoring**: Extract pure calculation functions, remove HTTP context
   - **Estimated Effort**: 2-3 days

3. **backend/internal/auth/service/auth_service.go**
   - **Why**: Authentication logic (JWT generation, password hashing)
   - **Impact**: High - used by all other services
   - **Refactoring**: Keep crypto logic, remove session management tied to HTTP
   - **Estimated Effort**: 2-3 days

4. **backend/cmd/payment-service/main.go**
   - **Why**: Entry point pattern to understand current initialization
   - **Impact**: Medium - template for all other services
   - **Refactoring**: Split into CLI entry point and API entry point
   - **Estimated Effort**: 1-2 days

5. **backend/internal/payment/handler/payment_handler.go**
   - **Why**: Current HTTP handler to understand adapter pattern
   - **Impact**: Medium - shows how to separate HTTP from business logic
   - **Refactoring**: Move to internal/adapters/rest/payment/handler.go
   - **Estimated Effort**: 1-2 days

### Phase 2: Infrastructure Files

6. **backend/pkg/database/postgres.go**
   - **Why**: Database connection shared by both CLI and API
   - **Impact**: Medium - needs to work in both contexts
   - **Refactoring**: Ensure connection pooling works for CLI (short-lived)
   - **Estimated Effort**: 1 day

7. **backend/internal/payment/xrpl/client.go**
   - **Why**: XRPL blockchain client
   - **Impact**: High - critical for payment processing
   - **Refactoring**: Ensure it's protocol-agnostic
   - **Estimated Effort**: 2 days

### Phase 3: New Files to Create

8. **proto/payment/v1/payment.proto** (NEW)
   - **Why**: gRPC protocol definition
   - **Impact**: Medium - enables gRPC support
   - **Estimated Effort**: 1 day

9. **internal/cli/commands/payment/create.go** (NEW)
   - **Why**: CLI command for payment creation
   - **Impact**: Medium - first CLI command, sets pattern
   - **Estimated Effort**: 1-2 days

10. **scripts/build-cli.sh** (NEW)
    - **Why**: Cross-platform build system
    - **Impact**: Low - but necessary for distribution
    - **Estimated Effort**: 0.5 day

---

## Recommendations

### Recommended Approach: CLI-First Incremental Migration

**Why CLI-First?**
1. ✅ Can build and test CLI without disrupting existing API
2. ✅ Validates core extraction is correct
3. ✅ Provides immediate value (automation, scripting)
4. ✅ Lower risk than changing API first
5. ✅ Team can learn patterns with less pressure

### Immediate Next Steps (Week 1)

#### Step 1: Create POC with Payment Service

**Goal**: Prove the architecture works with one service

**Tasks:**
1. Extract payment service to `internal/core/payment/`
2. Build simple CLI: `excise-tax-cli payment create --amount 100 --manufacturer-id 1`
3. Keep existing HTTP API running
4. Verify CLI can perform same operations as API

**Success Criteria:**
- [ ] CLI can create XRPL payments independently
- [ ] Core payment service has zero Gin/HTTP imports
- [ ] Existing API still works (backward compatibility)
- [ ] Unit tests for core service (no HTTP mocks needed)
- [ ] CLI binaries build for Windows, macOS, Linux

#### Step 2: Validate Architecture

**Tasks:**
1. Ensure core service has zero HTTP dependencies (`go list -json ./internal/core/... | jq '.Imports' | grep -i gin` should return nothing)
2. Verify database repositories work with domain models
3. Test XRPL integration from CLI
4. Run full test suite (unit + integration)

#### Step 3: Document Patterns

**Tasks:**
1. Create architectural decision records (ADRs)
2. Document DTO ↔ Domain conversion pattern
3. Create coding standards for adapters
4. Write onboarding guide for team

#### Step 4: Team Training

**Tasks:**
1. Present architecture to team (1-hour session)
2. Pair programming sessions on first service (2-3 sessions)
3. Code review guidelines for bifurcated code
4. Q&A session

### Decision Points

#### Decision 1: Proceed with Architecture?

**Options:**
- **A)** Yes, start with Week 1 POC (recommended)
- **B)** Yes, but modify the plan (specify changes)
- **C)** No, keep current monolithic approach

**Recommendation**: **Option A** - The benefits outweigh the complexity, especially for a government platform that needs automation, security, and flexibility.

#### Decision 2: Which Service to Bifurcate First?

**Options:**
- **A)** Payment Service (most complex, highest value)
- **B)** Auth Service (used by all others, simpler)
- **C)** Tax Service (medium complexity)

**Recommendation**: **Option A (Payment Service)** - If we can successfully bifurcate the most complex service, the rest will be easier.

#### Decision 3: Timeline Priority?

**Options:**
- **A)** Fast Track - 11 weeks full migration (aggressive)
- **B)** Gradual - CLI only first (5 weeks), API adapters later (6 weeks) - 11 weeks total, less risk
- **C)** Experimental - POC only (1 week), then decide

**Recommendation**: **Option B (Gradual)** - Build CLI completely first, validate it works, then add API adapters.

---

## Appendix

### Plugin Architecture Consideration

**Question**: Should core services be Go plugins that can be loaded dynamically?

**Analysis**:

**Pros:**
- Dynamic loading of services
- Hot reload without recompilation
- Extensibility for third-party plugins

**Cons:**
- **Go plugins limitations**:
  - Only work on Linux/macOS (not Windows)
  - Version compatibility issues (must match exact Go version)
  - CGo required (breaks cross-compilation)
  - No official Windows support
- **Complexity**: Build system, versioning, loading mechanics
- **Performance**: Plugin loading overhead

**Recommendation**: **NO** - Use standard Go modules instead

**Alternative**: Use **interface-based extensibility** compiled into the binary:

```go
// internal/core/payment/plugin/interface.go
type PaymentProcessor interface {
    Process(ctx context.Context, payment *domain.Payment) error
}

// internal/core/payment/plugin/xrpl_processor.go
type XRPLProcessor struct { /* ... */ }
func (p *XRPLProcessor) Process(ctx context.Context, payment *domain.Payment) error {
    // XRPL-specific processing
}

// internal/core/payment/plugin/ach_processor.go (future)
type ACHProcessor struct { /* ... */ }
func (p *ACHProcessor) Process(ctx context.Context, payment *domain.Payment) error {
    // ACH-specific processing
}

// Registered at compile time
var processors = map[string]PaymentProcessor{
    "xrpl": &XRPLProcessor{},
    "ach":  &ACHProcessor{},
}
```

This provides plugin-like extensibility without Go plugins' limitations.

---

## Summary

This bifurcated architecture provides:

1. ✅ **Pure CLI tools** for automation, testing, and operations
2. ✅ **Multi-protocol API layer** (REST, gRPC, WebSocket, GraphQL) for maximum flexibility
3. ✅ **Clean separation** between business logic and transport protocols
4. ✅ **Backward compatibility** during migration (zero downtime)
5. ✅ **Cross-platform support** (Windows, macOS, Linux)
6. ✅ **Future-proof design** for adding new protocols or use cases
7. ✅ **Improved testability** (unit tests without HTTP mocks)
8. ✅ **Better maintainability** (changes to one layer don't affect others)

The migration can be done incrementally over **11 weeks** with minimal risk and maximum value delivery at each phase.

### Key Metrics

- **Estimated Total Effort**: 11 weeks (1 senior developer full-time)
- **Code Increase**: ~30% (due to adapters and DTOs)
- **Performance Improvement**: 20-30% for CLI operations (no HTTP overhead)
- **Test Coverage**: Expected to increase from ~60% to ~80% (easier to test core)
- **API Flexibility**: 4 protocols supported (REST, gRPC, WebSocket, GraphQL)

---

**Document Version**: 1.0
**Last Updated**: 2025-12-29
**Status**: Proposed - Awaiting Approval
