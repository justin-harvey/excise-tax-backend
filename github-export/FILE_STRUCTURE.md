# File Structure Documentation

Complete directory structure and file organization for the Excise Tax Portal.

---

## 📁 **Root Directory**

```
excise-tax-portal/
├── README.md                           # Main project documentation
├── LICENSE                             # MIT License
├── .gitignore                          # Git ignore rules
├── CONTRIBUTING.md                     # Contribution guidelines
├── FILE_STRUCTURE.md                   # This file
├── STARTUP_GUIDE.md                    # Docker startup instructions
├── BUILD_COMPLETION_REPORT.md          # Build documentation
├── GITHUB_UPLOAD_INSTRUCTIONS.md       # GitHub upload guide
├── STATUS_AND_NEXT_STEPS.md            # Current status
├── README_START_HERE.md                # Quick start guide
├── README_FRONTEND_LEGACY.md           # Legacy frontend README
│
├── start-services.ps1                  # Start all Docker services
├── migrate-database.ps1                # Run database migrations
├── test-all-services.ps1               # Test infrastructure
├── check-docker.ps1                    # Verify Docker status
├── prepare-for-github.ps1              # Package for GitHub upload
│
├── backend/                            # Go microservices (see below)
├── infrastructure/                     # Docker & deployment configs
├── .github/                            # GitHub Actions workflows
├── deployment/                         # Frontend deployment versions
└── v2.0/                               # Frontend v2.0 versions
```

---

## 🔧 **Backend Directory** (`backend/`)

Complete Go microservices implementation for blockchain-powered tax collection.

```
backend/
├── README.md                           # Backend documentation
├── BACKEND_INFRASTRUCTURE_SPEC.md      # Architecture specification
├── go.mod                              # Go module definition
├── go.sum                              # Dependency checksums
├── Makefile                            # Build automation
├── .env.example                        # Environment template
│
├── cmd/                                # Service entry points
│   ├── api-gateway/
│   │   └── main.go                     # API Gateway service (Port 8080)
│   ├── payment-service/
│   │   └── main.go                     # Payment service (Port 8081)
│   ├── auth-service/
│   │   └── main.go                     # Authentication service (Port 8084)
│   ├── tax-service/
│   │   └── main.go                     # Tax calculation service (Port 8082)
│   ├── reporting-service/
│   │   └── main.go                     # Reporting service (Port 8083)
│   └── notification-service/
│       └── main.go                     # Notification service (Port 8085)
│
├── internal/                           # Private application code
│   ├── api-gateway/
│   │   ├── README.md
│   │   ├── middleware/                 # Request middleware
│   │   │   ├── auth.go                 # JWT authentication
│   │   │   ├── cors.go                 # CORS configuration
│   │   │   ├── ratelimit.go            # Rate limiting
│   │   │   └── logging.go              # Request logging
│   │   ├── handlers/                   # HTTP handlers
│   │   │   ├── health.go               # Health checks
│   │   │   └── proxy.go                # Service proxy
│   │   └── router/
│   │       └── router.go               # Route configuration
│   │
│   ├── payment/
│   │   ├── README.md
│   │   ├── models/
│   │   │   └── payment.go              # Payment data models
│   │   ├── handlers/
│   │   │   └── payment_handler.go      # Payment endpoints
│   │   ├── service/
│   │   │   ├── payment_service.go      # Business logic
│   │   │   ├── xrpl_client.go          # XRPL integration
│   │   │   ├── price_oracle.go         # Multi-source price oracle
│   │   │   └── payment_monitor.go      # Blockchain monitoring
│   │   └── repository/
│   │       └── payment_repository.go   # Database layer
│   │
│   ├── auth/
│   │   ├── models/
│   │   │   ├── user.go                 # User models
│   │   │   └── session.go              # Session models
│   │   ├── handlers/
│   │   │   └── auth_handler.go         # Auth endpoints
│   │   ├── service/
│   │   │   ├── auth_service.go         # Auth business logic
│   │   │   ├── jwt_service.go          # JWT management
│   │   │   ├── oauth_service.go        # OAuth integration
│   │   │   └── rbac_service.go         # Role-based access control
│   │   └── repository/
│   │       └── user_repository.go      # User database layer
│   │
│   ├── tax/
│   │   ├── models/
│   │   │   └── tax.go                  # Tax data models
│   │   ├── handlers/
│   │   │   └── tax_handler.go          # Tax endpoints
│   │   ├── service/
│   │   │   ├── tax_service.go          # Tax business logic
│   │   │   ├── calculator.go           # Tax calculations
│   │   │   └── validator.go            # Tax validation
│   │   └── repository/
│   │       └── tax_repository.go       # Tax database layer
│   │
│   ├── reporting/
│   │   ├── models/
│   │   │   └── report.go               # Report models
│   │   ├── handlers/
│   │   │   └── reporting_handler.go    # Report endpoints
│   │   ├── service/
│   │   │   ├── reporting_service.go    # Report generation
│   │   │   ├── dashboard_service.go    # Dashboard data
│   │   │   ├── analytics_service.go    # Analytics
│   │   │   └── export_service.go       # Report export (CSV, PDF)
│   │   └── repository/
│   │       └── reporting_repository.go # Report database layer
│   │
│   └── notification/
│       ├── models/
│       │   └── notification.go         # Notification models
│       ├── handlers/
│       │   └── notification_handler.go # Notification endpoints
│       ├── service/
│       │   ├── notification_service.go # Notification logic
│       │   ├── email_service.go        # Email delivery
│       │   ├── sms_service.go          # SMS delivery
│       │   └── webhook_service.go      # Webhook delivery
│       └── repository/
│           └── notification_repository.go
│
├── pkg/                                # Shared packages
│   ├── README.md
│   ├── database/
│   │   └── postgres.go                 # PostgreSQL connection pool
│   ├── cache/
│   │   └── redis.go                    # Redis client wrapper
│   ├── logger/
│   │   └── logger.go                   # Structured logging (Zap)
│   ├── config/
│   │   └── config.go                   # Configuration management
│   ├── errors/
│   │   └── errors.go                   # Custom error types
│   ├── validator/
│   │   └── validator.go                # Request validation
│   ├── utils/
│   │   ├── jwt.go                      # JWT utilities
│   │   └── crypto.go                   # Cryptography helpers
│   └── queue/
│       └── rabbitmq.go                 # Message queue client
│
├── migrations/                         # Database migrations
│   ├── README.md
│   ├── 000001_init_schema.up.sql       # Initial schema
│   ├── 000001_init_schema.down.sql
│   ├── 000002_add_xrpl_tables.up.sql   # XRPL payment tables
│   ├── 000002_add_xrpl_tables.down.sql
│   ├── 000003_add_tax_tables.up.sql    # Tax reporting tables
│   ├── 000003_add_tax_tables.down.sql
│   ├── 000004_add_views.up.sql         # Database views
│   └── 000004_add_views.down.sql
│
├── configs/                            # Configuration files
│   ├── config.yaml                     # Default configuration
│   ├── config.dev.yaml                 # Development config
│   └── config.prod.yaml                # Production config
│
├── scripts/                            # Build & utility scripts
│   ├── build.sh                        # Build all services
│   ├── test.sh                         # Run all tests
│   └── deploy.sh                       # Deployment script
│
├── tests/                              # Test files
│   ├── integration/                    # Integration tests
│   └── e2e/                            # End-to-end tests
│
└── docs/                               # Additional documentation
    ├── api/                            # API documentation
    ├── architecture/                   # Architecture diagrams
    └── deployment/                     # Deployment guides
```

---

## 🐳 **Infrastructure Directory** (`infrastructure/`)

Docker Compose configuration and deployment infrastructure.

```
infrastructure/
└── docker/
    ├── README.md                       # Docker documentation
    ├── QUICKSTART.md                   # Quick start guide
    ├── TROUBLESHOOTING.md              # Troubleshooting guide
    ├── docker-compose.yml              # Main compose file
    ├── .env.docker                     # Docker environment variables
    │
    ├── postgres/                       # PostgreSQL configuration
    │   ├── init-db.sql                 # Database initialization
    │   └── postgresql.conf             # PostgreSQL settings
    │
    ├── prometheus/                     # Prometheus monitoring
    │   ├── prometheus.yml              # Prometheus configuration
    │   └── alerts.yml                  # Alert rules
    │
    └── grafana/                        # Grafana dashboards
        ├── datasources/
        │   └── prometheus.yml          # Prometheus datasource
        └── dashboards/
            └── backend-dashboard.json  # Backend metrics dashboard
```

---

## 🤖 **GitHub Actions** (`.github/`)

CI/CD automation workflows.

```
.github/
└── workflows/
    └── backend-ci.yml                  # Backend CI/CD pipeline
                                        # - Linting (golangci-lint)
                                        # - Testing (with PostgreSQL/Redis)
                                        # - Building (6 microservices)
                                        # - Docker image publishing
```

---

## 🌐 **Frontend Deployment** (`deployment/` & `v2.0/`)

Legacy frontend versions (PayPal, Circle, Ripple integrations).

```
deployment/
├── paypal-professional/                # PayPal payment integration
│   ├── index.html
│   ├── admin-dashboard.html
│   └── README.md
│
├── circle-professional/                # Circle payment integration
│   ├── index.html
│   └── README.md
│
└── ripple-professional/                # Ripple payment integration
    ├── index.html
    └── README.md

v2.0/
├── paypal-professional/                # Version 2.0 PayPal
│   ├── index.html
│   ├── admin-dashboard.html
│   └── README.md
│
└── README_CRYPTO_THEME.md              # Crypto theme documentation
```

---

## 📝 **Key Files Explained**

### Root Level Scripts

| File | Purpose | Usage |
|------|---------|-------|
| `start-services.ps1` | Start all Docker services | `.\start-services.ps1` |
| `migrate-database.ps1` | Run database migrations | `.\migrate-database.ps1` |
| `test-all-services.ps1` | Test infrastructure | `.\test-all-services.ps1` |
| `check-docker.ps1` | Verify Docker Desktop | `.\check-docker.ps1` |
| `prepare-for-github.ps1` | Export for GitHub | `.\prepare-for-github.ps1` |

### Backend Configuration

| File | Purpose |
|------|---------|
| `backend/go.mod` | Go module dependencies |
| `backend/Makefile` | Build automation targets |
| `backend/.env.example` | Environment variable template |
| `backend/configs/config.yaml` | Default configuration |

### Infrastructure Services

| Service | Port | Configuration |
|---------|------|---------------|
| PostgreSQL | 5432 | `infrastructure/docker/postgres/` |
| Redis | 6379 | `docker-compose.yml` |
| RabbitMQ | 5672, 15672 | `docker-compose.yml` |
| Prometheus | 9090 | `infrastructure/docker/prometheus/` |
| Grafana | 3000 | `infrastructure/docker/grafana/` |

### Microservices

| Service | Port | Entry Point | Purpose |
|---------|------|-------------|---------|
| API Gateway | 8080 | `cmd/api-gateway/main.go` | Request routing, auth, rate limiting |
| Payment Service | 8081 | `cmd/payment-service/main.go` | XRPL blockchain payments |
| Tax Service | 8082 | `cmd/tax-service/main.go` | Tax calculations |
| Reporting Service | 8083 | `cmd/reporting-service/main.go` | Reports & analytics |
| Auth Service | 8084 | `cmd/auth-service/main.go` | JWT authentication, RBAC |
| Notification Service | 8085 | `cmd/notification-service/main.go` | Email, SMS, webhooks |

---

## 📊 **Statistics**

- **Total Backend Files:** ~150
- **Lines of Go Code:** ~15,000
- **Microservices:** 6
- **Database Migrations:** 4 sets
- **Shared Packages:** 8
- **Docker Services:** 5
- **Documentation Files:** 12+
- **PowerShell Scripts:** 5

---

## 🔍 **Finding Files**

### By Functionality

**Authentication & Authorization:**
- `backend/internal/auth/` - Auth service
- `backend/pkg/utils/jwt.go` - JWT utilities
- `backend/internal/api-gateway/middleware/auth.go` - Auth middleware

**XRPL Blockchain:**
- `backend/internal/payment/service/xrpl_client.go` - XRPL integration
- `backend/internal/payment/service/price_oracle.go` - Price aggregation
- `backend/internal/payment/service/payment_monitor.go` - Blockchain monitoring
- `backend/migrations/000002_add_xrpl_tables.up.sql` - Payment schema

**Tax Calculations:**
- `backend/internal/tax/service/calculator.go` - Tax formulas
- `backend/internal/tax/service/validator.go` - Tax validation
- `backend/migrations/000003_add_tax_tables.up.sql` - Tax schema

**Database:**
- `backend/migrations/` - All database migrations
- `backend/pkg/database/postgres.go` - Database connection pool
- `infrastructure/docker/postgres/` - PostgreSQL configuration

**Monitoring & Observability:**
- `infrastructure/docker/prometheus/` - Prometheus config
- `infrastructure/docker/grafana/` - Grafana dashboards
- `backend/pkg/logger/logger.go` - Structured logging

**Testing:**
- `backend/tests/` - Test suites
- `test-all-services.ps1` - Infrastructure tests
- `.github/workflows/backend-ci.yml` - CI tests

**Documentation:**
- `README.md` - Main documentation
- `STARTUP_GUIDE.md` - Setup instructions
- `BUILD_COMPLETION_REPORT.md` - Build details
- `CONTRIBUTING.md` - Contribution guidelines
- `GITHUB_UPLOAD_INSTRUCTIONS.md` - GitHub upload guide

---

## 🛠️ **Development Workflow**

### Typical File Modification Paths

1. **Adding a New API Endpoint:**
   - Add handler: `backend/internal/{service}/handlers/`
   - Add business logic: `backend/internal/{service}/service/`
   - Update router: `backend/internal/api-gateway/router/router.go`
   - Add tests: `backend/tests/`

2. **Database Changes:**
   - Create migration: `backend/migrations/`
   - Run migration: `.\migrate-database.ps1`
   - Update models: `backend/internal/{service}/models/`

3. **Configuration Changes:**
   - Update: `backend/configs/config.yaml`
   - Update example: `backend/.env.example`
   - Update Docker: `infrastructure/docker/docker-compose.yml`

4. **Adding Shared Functionality:**
   - Add to: `backend/pkg/{category}/`
   - Update: `backend/pkg/README.md`

---

## 📚 **Further Reading**

- [README.md](README.md) - Project overview and quick start
- [BACKEND_INFRASTRUCTURE_SPEC.md](backend/BACKEND_INFRASTRUCTURE_SPEC.md) - Complete architecture specification
- [STARTUP_GUIDE.md](STARTUP_GUIDE.md) - Detailed setup instructions
- [CONTRIBUTING.md](CONTRIBUTING.md) - How to contribute
- [BUILD_COMPLETION_REPORT.md](BUILD_COMPLETION_REPORT.md) - Build documentation

---

**Last Updated:** 2025-12-29
**Repository:** https://github.com/YOUR_USERNAME/excise-tax-portal
