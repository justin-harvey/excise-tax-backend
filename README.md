# Excise Tax Portal - Go Backend

A scalable microservices-based backend for the Excise Tax Portal, built with Go 1.21+ and designed for high-performance tax payment processing with XRPL blockchain integration.

## Architecture Overview

The backend is organized as a microservices architecture with the following services:

- **API Gateway** (Port 8080): Routes requests, handles authentication and CORS
- **Payment Service** (Port 8081): Processes payments via XRPL blockchain
- **Tax Service** (Port 8082): Tax calculation and validation logic
- **Reporting Service** (Port 8083): Generates tax reports and analytics
- **Notification Service** (Port 8084): Sends email and SMS notifications
- **Auth Service** (Port 8085): User authentication and authorization (JWT, OAuth)

## Project Structure

```
backend/
├── cmd/                        # Entry points for each service
│   ├── api-gateway/
│   ├── payment-service/
│   ├── tax-service/
│   ├── reporting-service/
│   ├── notification-service/
│   └── auth-service/
├── internal/                   # Private application code
│   ├── api-gateway/           # Gateway handlers, middleware, routing
│   ├── payment/               # Payment service with XRPL integration
│   ├── tax/                   # Tax calculation logic
│   ├── reporting/             # Report generation
│   ├── notification/          # Email/SMS service
│   └── auth/                  # Authentication logic
├── pkg/                       # Shared packages
│   ├── database/              # PostgreSQL connection pool
│   ├── logger/                # Structured logging
│   ├── config/                # Configuration management
│   ├── errors/                # Custom error types
│   ├── validator/             # Request validation
│   └── utils/                 # Helper functions
├── migrations/                # Database migrations
├── configs/                   # Configuration files
├── scripts/                   # Utility scripts
├── tests/                     # Integration and E2E tests
├── docs/                      # API documentation
├── go.mod                     # Go module definition
└── .env.example              # Environment variables template
```

## Prerequisites

- **Go 1.21+**: [Install Go](https://golang.org/doc/install)
- **PostgreSQL 14+**: For data persistence
- **Docker 24.0+**: For containerized deployment
- **golang-migrate**: For database migrations
- **golangci-lint**: For code linting

## Quick Start

### 1. Clone and Setup

```bash
cd backend

# Copy environment variables
cp .env.example .env

# Edit .env with your configuration
# Set database credentials, JWT secret, etc.
```

### 2. Install Dependencies

```bash
# Download Go dependencies
go mod download

# Install development tools
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

### 3. Setup Database

```bash
# Start PostgreSQL with Docker
docker compose up -d postgres

# Run migrations
migrate -path migrations -database "postgresql://postgres:<your-password>@localhost:5432/excise_tax_db?sslmode=disable" up
```

### 4. Run Services

Option A: Run individual services in separate terminals:

```bash
# Terminal 1 - API Gateway
go run cmd/api-gateway/main.go

# Terminal 2 - Payment Service
go run cmd/payment-service/main.go

# Terminal 3 - Tax Service
go run cmd/tax-service/main.go

# Terminal 4 - Reporting Service
go run cmd/reporting-service/main.go

# Terminal 5 - Notification Service
go run cmd/notification-service/main.go

# Terminal 6 - Auth Service
go run cmd/auth-service/main.go
```

Option B: Use Docker Compose:

```bash
docker compose --profile dev up
```

### 5. Verify Services

```bash
# Check API Gateway health
curl http://localhost:8080/health

# Check Docker services status
docker compose ps
```

## Development Workflow

### Building

```bash
# Build all services
go build -o bin/api-gateway ./cmd/api-gateway
go build -o bin/payment-service ./cmd/payment-service
go build -o bin/tax-service ./cmd/tax-service
go build -o bin/auth-service ./cmd/auth-service
go build -o bin/reporting-service ./cmd/reporting-service
go build -o bin/notification-service ./cmd/notification-service

# Or build with Docker
docker compose build
```

### Testing

```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run specific package tests
go test -v ./internal/payment/...
go test -v ./pkg/utils/...
```

### Code Quality

```bash
# Format code
go fmt ./...

# Run linter
golangci-lint run

# Run tests
go test -v ./...
```

### Database Migrations

```bash
# Create new migration
migrate create -ext sql -dir migrations -seq add_users_table

# Apply migrations
migrate -path migrations -database "postgresql://postgres:<password>@localhost:5432/excise_tax_db?sslmode=disable" up

# Rollback last migration
migrate -path migrations -database "postgresql://postgres:<password>@localhost:5432/excise_tax_db?sslmode=disable" down 1

# Check migration status
migrate -path migrations -database "postgresql://postgres:<password>@localhost:5432/excise_tax_db?sslmode=disable" version

# Force to specific version (use with caution)
migrate -path migrations -database "postgresql://postgres:<password>@localhost:5432/excise_tax_db?sslmode=disable" force <version>
```

## Configuration

Configuration can be provided via:

1. **YAML files** in `configs/` directory
2. **Environment variables** (prefixed with `APP_`)
3. **.env file** for local development

### Key Configuration Sections

#### Server
- `SERVER_HOST`: Server bind address (default: 0.0.0.0)
- `SERVER_PORT`: Server port (default: 8080)
- `ENV`: Environment (development/staging/production)

#### Database
- `DATABASE_HOST`: PostgreSQL host
- `DATABASE_PORT`: PostgreSQL port (default: 5432)
- `DATABASE_USER`: Database user
- `DATABASE_PASSWORD`: Database password
- `DATABASE_NAME`: Database name

#### XRPL
- `XRPL_NETWORK_URL`: XRPL network WebSocket URL
- `XRPL_WALLET_ADDRESS`: XRPL wallet address
- `XRPL_WALLET_SEED`: XRPL wallet seed (KEEP SECRET!)
- `XRPL_IS_TESTNET`: Use testnet (true/false)

## API Documentation

### API Gateway Endpoints

- `GET /health` - Health check
- `GET /health/liveness` - Liveness probe
- `GET /health/readiness` - Readiness probe

### Authentication Endpoints

- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout
- `POST /api/v1/auth/forgot-password` - Request password reset
- `POST /api/v1/auth/reset-password` - Reset password

### Payment Endpoints

- `POST /api/v1/payments` - Create payment
- `GET /api/v1/payments/:id` - Get payment details
- `GET /api/v1/payments` - List payments
- `POST /api/v1/payments/:id/confirm` - Confirm payment

### Tax Endpoints

- `POST /api/v1/tax/calculate` - Calculate tax
- `GET /api/v1/tax/rates` - Get tax rates
- `POST /api/v1/tax/reports` - Submit tax report
- `GET /api/v1/tax/reports/:id` - Get report details

For detailed API documentation, see [docs/swagger.yaml](docs/swagger.yaml)

## Docker Deployment

### Build Images

```bash
docker compose build
```

### Run with Docker Compose

```bash
# Start all services
docker compose --profile dev up

# View logs
docker compose logs -f

# Stop services
docker compose down
```

## Production Deployment

### Prerequisites
- Kubernetes cluster or Docker Swarm
- PostgreSQL cluster (managed or self-hosted)
- Load balancer (NGINX, HAProxy, or cloud LB)

### Deployment Steps

1. **Build production images**
```bash
docker compose build
```

2. **Push to container registry**
```bash
docker tag excise-tax/api-gateway your-registry/excise-tax/api-gateway:v1.0.0
docker push your-registry/excise-tax/api-gateway:v1.0.0
# Repeat for other services
```

3. **Deploy to Kubernetes**
```bash
kubectl apply -f infrastructure/kubernetes/
```

4. **Configure monitoring**
- Setup application logging
- Configure alerts for critical errors
- Monitor health check endpoints

## Security Considerations

1. **Secrets Management**
   - Never commit `.env` files
   - Use secrets manager (AWS Secrets Manager, HashiCorp Vault)
   - Rotate JWT secrets regularly

2. **XRPL Wallet Security**
   - Store wallet seeds in encrypted secrets
   - Use hardware security modules (HSM) in production
   - Implement multi-signature wallets for large transactions

3. **Database Security**
   - Use strong passwords
   - Enable SSL/TLS connections
   - Implement row-level security
   - Regular backups and disaster recovery plan

4. **API Security**
   - Rate limiting enabled by default
   - CORS properly configured
   - Input validation on all endpoints
   - SQL injection prevention via parameterized queries

## Monitoring and Observability

### Health Checks
All services expose:
- `/health` - Overall health
- `/health/liveness` - Kubernetes liveness probe
- `/health/readiness` - Kubernetes readiness probe

### Metrics
Health check metrics available at `/health`:
- Service status
- Database connectivity
- XRPL connection status
- Error rates

### Logging
Structured JSON logging with:
- Request ID tracking
- User ID context
- Trace ID for distributed tracing
- Error stack traces

## Performance Optimization

- **Connection Pooling**: Configured for PostgreSQL
- **Database Indexing**: Optimized indexes on frequent queries
- **Horizontal Scaling**: Stateless services support multiple replicas
- **XRPL Integration**: Async payment processing with webhooks
- **Efficient Queries**: Optimized SQL queries with proper indexes

## Troubleshooting

### Common Issues

**Database connection failed**
```bash
# Check database is running
docker ps | grep postgres

# Check connection
docker exec -it excise-tax-postgres psql -U postgres -d excise_tax_db

# Verify credentials in .env
cat .env | grep POSTGRES
```

**Migration errors**
```bash
# Check migration version
migrate -path migrations -database "postgresql://postgres:<password>@localhost:5432/excise_tax_db?sslmode=disable" version

# Force to specific version (use with caution)
migrate -path migrations -database "postgresql://postgres:<password>@localhost:5432/excise_tax_db?sslmode=disable" force <version>
```

**Build errors**
```bash
# Clear Go cache
go clean -cache -modcache

# Re-download dependencies
go mod download
```

## Contributing

1. Follow Go coding standards and conventions
2. Write tests for new features
3. Run tests and linter before committing
4. Update documentation for API changes
5. Follow semantic versioning for releases

## License

Copyright (c) 2024 State Government. All rights reserved.

## Support

For questions or issues:
- Technical Documentation: [docs/](docs/)
- API Reference: [docs/swagger.yaml](docs/swagger.yaml)
- Architecture Guide: [BACKEND_INFRASTRUCTURE_SPEC.md](BACKEND_INFRASTRUCTURE_SPEC.md)
