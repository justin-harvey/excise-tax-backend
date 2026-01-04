# Excise Tax Portal - Backend Infrastructure Specification
## Production-Ready Golang Microservices Architecture

**Version:** 2.0.0  
**Date:** December 29, 2025  
**Status:** Ready for Development

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Architecture Overview](#architecture-overview)
3. [Technology Stack](#technology-stack)
4. [System Requirements](#system-requirements)
5. [Project Structure](#project-structure)
6. [Service Specifications](#service-specifications)
7. [Database Schema](#database-schema)
8. [API Specifications](#api-specifications)
9. [Authentication & Authorization](#authentication--authorization)
10. [XRPL Blockchain Integration](#xrpl-blockchain-integration)
11. [Docker Configuration](#docker-configuration)
12. [CI/CD Pipeline](#cicd-pipeline)
13. [Cloud Deployment](#cloud-deployment)
14. [Security Requirements](#security-requirements)
15. [Monitoring & Logging](#monitoring--logging)
16. [Development Workflow](#development-workflow)

---

## Executive Summary

This document specifies the production-ready backend infrastructure for the Excise Tax Portal, transitioning from a static Netlify site to a fully orchestrated, cloud-native application. The system will support state government excise tax collection with blockchain payment integration.

### Key Objectives:
- **Replace static frontend** with production React application
- **Build Golang backend** for high performance and cloud-native deployment
- **Implement XRPL blockchain** payment processing
- **Deploy via Docker/Kubernetes** for multi-cloud compatibility
- **Enable white-label theming** for state-specific customization
- **Achieve enterprise scale** with low latency

### Business Impact:
- **Cost Reduction:** 99.9% lower payment processing fees ($3-4B → ~$4M annually)
- **Speed:** 3-5 second settlement vs 30+ days
- **Market:** $250-300B annual US excise tax volume
- **Expansion:** Foundation for $16T+ government payments under EO 14247

---

## Architecture Overview

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Load Balancer (Nginx)                    │
│                    SSL/TLS Termination                       │
└────────────────────────────┬────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────┐
│                  React Frontend (Docker)                     │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  • State-specific theming                            │  │
│  │  • Multi-route SPA                                   │  │
│  │  • Real-time WebSocket updates                       │  │
│  │  • Responsive design                                 │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────┘
                             │ REST API / WebSocket
                             ▼
┌─────────────────────────────────────────────────────────────┐
│              API Gateway (Golang) - Port 8080               │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  • Request routing                                    │  │
│  │  • Rate limiting                                      │  │
│  │  • Authentication middleware                          │  │
│  │  • Request/response logging                           │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────┘
                             │
            ┌────────────────┼────────────────┐
            │                │                │
            ▼                ▼                ▼
┌──────────────────┐ ┌──────────────┐ ┌──────────────────┐
│  Payment Service │ │ Tax Service  │ │ Reporting Service│
│    (Golang)      │ │  (Golang)    │ │    (Golang)      │
│   Port 8081      │ │  Port 8082   │ │   Port 8083      │
└────────┬─────────┘ └──────┬───────┘ └────────┬─────────┘
         │                  │                  │
         └──────────────────┼──────────────────┘
                            │
                            ▼
         ┌──────────────────────────────────┐
         │  PostgreSQL Database (Primary)   │
         │  + Read Replicas (2x)            │
         └──────────────────────────────────┘
                            │
         ┌──────────────────┼──────────────────┐
         │                  │                  │
         ▼                  ▼                  ▼
┌──────────────────┐ ┌──────────────┐ ┌──────────────────┐
│   Redis Cache    │ │ XRPL Network │ │  S3/GCS Storage  │
│   (Sessions)     │ │ (Blockchain) │ │  (Documents)     │
└──────────────────┘ └──────────────┘ └──────────────────┘
```

### Microservices Architecture

**Core Services:**
1. **API Gateway** - Request routing, authentication, rate limiting
2. **Payment Service** - XRPL blockchain payments, ACH, credit cards
3. **Tax Service** - Tax calculation, reporting, compliance
4. **Reporting Service** - Admin dashboards, analytics, exports
5. **Notification Service** - Email, SMS, webhooks
6. **Auth Service** - JWT tokens, OAuth, CDB integration

**Supporting Infrastructure:**
- **PostgreSQL** - Primary data store with replication
- **Redis** - Session management, caching, pub/sub
- **RabbitMQ** - Async job queue (payment processing, emails)
- **MinIO/S3** - Document storage (receipts, reports)
- **Prometheus + Grafana** - Metrics and monitoring
- **ELK Stack** - Centralized logging

---

## Technology Stack

### Backend

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| **Primary Language** | Go (Golang) | 1.21+ | Core services |
| **Web Framework** | Gin / Chi | Latest | HTTP routing |
| **Database** | PostgreSQL | 15+ | Primary data store |
| **Cache** | Redis | 7+ | Sessions, caching |
| **Message Queue** | RabbitMQ | 3.12+ | Async processing |
| **Blockchain SDK** | XRPL Go SDK | Latest | XRP Ledger integration |
| **ORM** | GORM | Latest | Database abstraction |
| **Migration Tool** | golang-migrate | Latest | DB migrations |
| **Testing** | Testify | Latest | Unit/integration tests |
| **Validation** | validator/v10 | Latest | Request validation |
| **Logging** | Logrus / Zap | Latest | Structured logging |

### Frontend

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| **Framework** | React | 18+ | UI framework |
| **Build Tool** | Vite | 5+ | Fast builds |
| **Styling** | TailwindCSS | 3+ | Utility-first CSS |
| **State Management** | Zustand | Latest | Global state |
| **HTTP Client** | Axios | Latest | API calls |
| **WebSocket** | Socket.io-client | Latest | Real-time updates |
| **Forms** | React Hook Form | Latest | Form handling |
| **Charts** | Recharts | Latest | Data visualization |
| **Testing** | Vitest + RTL | Latest | Component testing |

### DevOps

| Component | Technology | Version | Purpose |
|-----------|-----------|---------|---------|
| **Containerization** | Docker | 24+ | Application packaging |
| **Orchestration** | Kubernetes | 1.28+ | Container orchestration |
| **CI/CD** | GitHub Actions | Latest | Automated pipeline |
| **IaC** | Terraform | 1.6+ | Infrastructure as code |
| **Monitoring** | Prometheus | Latest | Metrics collection |
| **Visualization** | Grafana | Latest | Dashboards |
| **Logging** | Elasticsearch | 8+ | Log aggregation |
| **Log Shipping** | Logstash/Fluentd | Latest | Log processing |
| **APM** | Jaeger | Latest | Distributed tracing |

---

## System Requirements

### Development Environment

**Minimum:**
- CPU: 4 cores
- RAM: 8 GB
- Disk: 50 GB SSD
- OS: Linux, macOS, or Windows with WSL2

**Recommended:**
- CPU: 8 cores
- RAM: 16 GB
- Disk: 100 GB NVMe SSD
- OS: Linux (Ubuntu 22.04 LTS) or macOS

### Production Environment

**Per Service Instance:**
- CPU: 2-4 vCPUs
- RAM: 4-8 GB
- Network: 1 Gbps
- Storage: 50 GB SSD (+ database volumes)

**Database Server:**
- CPU: 8-16 vCPUs
- RAM: 32-64 GB
- Storage: 500 GB - 2 TB NVMe SSD
- IOPS: 10,000+

**Scalability Targets:**
- Handle 10,000 concurrent users
- Process 1,000 transactions per second
- 99.9% uptime SLA
- < 100ms average API response time
- < 5 second XRPL payment confirmation

---

## Project Structure

```
excise-tax-portal/
├── backend/
│   ├── cmd/
│   │   ├── api-gateway/          # API Gateway entry point
│   │   │   └── main.go
│   │   ├── payment-service/       # Payment service entry point
│   │   │   └── main.go
│   │   ├── tax-service/           # Tax service entry point
│   │   │   └── main.go
│   │   ├── reporting-service/     # Reporting service entry point
│   │   │   └── main.go
│   │   ├── notification-service/  # Notification service entry point
│   │   │   └── main.go
│   │   └── auth-service/          # Auth service entry point
│   │       └── main.go
│   │
│   ├── internal/                  # Private application code
│   │   ├── api-gateway/
│   │   │   ├── handler/           # HTTP handlers
│   │   │   ├── middleware/        # Auth, logging, rate limiting
│   │   │   └── router/            # Route definitions
│   │   │
│   │   ├── payment/
│   │   │   ├── handler/           # Payment endpoints
│   │   │   ├── service/           # Business logic
│   │   │   ├── repository/        # Data access
│   │   │   ├── xrpl/              # XRPL blockchain integration
│   │   │   │   ├── client.go      # XRPL client wrapper
│   │   │   │   ├── payment.go     # Payment processing
│   │   │   │   ├── monitor.go     # Transaction monitoring
│   │   │   │   └── oracle.go      # Price oracle (XRP/USD)
│   │   │   └── model/             # Domain models
│   │   │
│   │   ├── tax/
│   │   │   ├── handler/
│   │   │   ├── service/           # Tax calculation logic
│   │   │   ├── repository/
│   │   │   └── model/
│   │   │
│   │   ├── reporting/
│   │   │   ├── handler/
│   │   │   ├── service/           # Report generation
│   │   │   ├── repository/
│   │   │   └── model/
│   │   │
│   │   ├── notification/
│   │   │   ├── handler/
│   │   │   ├── service/           # Email, SMS service
│   │   │   └── template/          # Email templates
│   │   │
│   │   └── auth/
│   │       ├── handler/
│   │       ├── service/           # JWT, OAuth logic
│   │       ├── middleware/        # Auth middleware
│   │       └── model/
│   │
│   ├── pkg/                       # Public shared libraries
│   │   ├── database/              # DB connection pool
│   │   │   ├── postgres.go
│   │   │   └── migrations/
│   │   ├── cache/                 # Redis client
│   │   │   └── redis.go
│   │   ├── queue/                 # RabbitMQ client
│   │   │   └── rabbitmq.go
│   │   ├── storage/               # S3/MinIO client
│   │   │   └── s3.go
│   │   ├── logger/                # Structured logging
│   │   │   └── logger.go
│   │   ├── config/                # Configuration loading
│   │   │   └── config.go
│   │   ├── errors/                # Custom error types
│   │   │   └── errors.go
│   │   ├── validator/             # Request validation
│   │   │   └── validator.go
│   │   └── utils/                 # Helper functions
│   │       ├── crypto.go
│   │       ├── jwt.go
│   │       └── time.go
│   │
│   ├── migrations/                # Database migrations
│   │   ├── 000001_init_schema.up.sql
│   │   ├── 000001_init_schema.down.sql
│   │   ├── 000002_add_xrpl_tables.up.sql
│   │   ├── 000002_add_xrpl_tables.down.sql
│   │   └── ...
│   │
│   ├── scripts/                   # Utility scripts
│   │   ├── setup.sh               # Local setup script
│   │   ├── migrate.sh             # Migration runner
│   │   ├── seed.sh                # Seed test data
│   │   └── test.sh                # Run tests
│   │
│   ├── tests/                     # Integration tests
│   │   ├── api/                   # API integration tests
│   │   ├── e2e/                   # End-to-end tests
│   │   └── fixtures/              # Test data
│   │
│   ├── configs/                   # Configuration files
│   │   ├── config.yaml            # Default config
│   │   ├── config.dev.yaml        # Development config
│   │   ├── config.staging.yaml    # Staging config
│   │   └── config.prod.yaml       # Production config
│   │
│   ├── docs/                      # API documentation
│   │   ├── swagger.yaml           # OpenAPI spec
│   │   └── architecture.md        # Architecture docs
│   │
│   ├── go.mod                     # Go module definition
│   ├── go.sum                     # Go dependencies checksum
│   ├── Makefile                   # Build automation
│   ├── .env.example               # Environment template
│   └── README.md                  # Backend README
│
├── frontend/
│   ├── src/
│   │   ├── components/            # React components
│   │   │   ├── common/            # Shared components
│   │   │   ├── company/           # Company portal components
│   │   │   └── admin/             # Admin portal components
│   │   ├── pages/                 # Page components
│   │   │   ├── Company/
│   │   │   │   ├── Dashboard.jsx
│   │   │   │   ├── Payment.jsx
│   │   │   │   ├── Reports.jsx
│   │   │   │   └── Profile.jsx
│   │   │   └── Admin/
│   │   │       ├── Dashboard.jsx
│   │   │       ├── ReviewReports.jsx
│   │   │       ├── Users.jsx
│   │   │       └── Settings.jsx
│   │   ├── services/              # API service layer
│   │   │   ├── api.js             # Axios config
│   │   │   ├── auth.js            # Auth API calls
│   │   │   ├── payment.js         # Payment API calls
│   │   │   ├── tax.js             # Tax API calls
│   │   │   └── xrpl.js            # XRPL API calls
│   │   ├── store/                 # State management
│   │   │   ├── auth.js            # Auth store
│   │   │   ├── payment.js         # Payment store
│   │   │   └── config.js          # App config store
│   │   ├── hooks/                 # Custom React hooks
│   │   │   ├── useAuth.js
│   │   │   ├── usePayment.js
│   │   │   └── useWebSocket.js
│   │   ├── utils/                 # Utility functions
│   │   │   ├── formatting.js
│   │   │   ├── validation.js
│   │   │   └── constants.js
│   │   ├── themes/                # State-specific themes
│   │   │   ├── default.js         # Default theme
│   │   │   ├── nebraska.js        # Nebraska theme
│   │   │   └── maine.js           # Maine theme
│   │   ├── App.jsx                # Root component
│   │   ├── main.jsx               # Entry point
│   │   └── router.jsx             # Route configuration
│   │
│   ├── public/                    # Static assets
│   │   ├── logos/                 # State logos
│   │   └── favicons/              # Favicons
│   │
│   ├── tests/                     # Frontend tests
│   │   ├── unit/
│   │   ├── integration/
│   │   └── e2e/
│   │
│   ├── package.json               # npm dependencies
│   ├── vite.config.js             # Vite configuration
│   ├── tailwind.config.js         # Tailwind configuration
│   ├── .env.example               # Environment template
│   └── README.md                  # Frontend README
│
├── infrastructure/
│   ├── docker/                    # Docker configurations
│   │   ├── Dockerfile.backend     # Backend multi-stage
│   │   ├── Dockerfile.frontend    # Frontend multi-stage
│   │   └── docker-compose.yml     # Local development
│   │
│   ├── kubernetes/                # K8s manifests
│   │   ├── base/                  # Base configs
│   │   │   ├── namespace.yaml
│   │   │   ├── configmap.yaml
│   │   │   ├── secrets.yaml
│   │   │   └── services.yaml
│   │   ├── deployments/           # Service deployments
│   │   │   ├── api-gateway.yaml
│   │   │   ├── payment-service.yaml
│   │   │   ├── tax-service.yaml
│   │   │   └── frontend.yaml
│   │   ├── ingress/               # Ingress rules
│   │   │   └── ingress.yaml
│   │   └── monitoring/            # Monitoring stack
│   │       ├── prometheus.yaml
│   │       └── grafana.yaml
│   │
│   ├── terraform/                 # Infrastructure as Code
│   │   ├── modules/               # Reusable modules
│   │   │   ├── eks/               # AWS EKS
│   │   │   ├── gke/               # GCP GKE
│   │   │   ├── aks/               # Azure AKS
│   │   │   ├── rds/               # AWS RDS
│   │   │   └── networking/        # VPC, subnets
│   │   ├── environments/          # Environment-specific
│   │   │   ├── dev/
│   │   │   ├── staging/
│   │   │   └── production/
│   │   ├── main.tf                # Main configuration
│   │   ├── variables.tf           # Input variables
│   │   ├── outputs.tf             # Output values
│   │   └── terraform.tfvars       # Variable values
│   │
│   └── helm/                      # Helm charts (alternative to K8s)
│       └── excise-tax-portal/
│           ├── Chart.yaml
│           ├── values.yaml
│           └── templates/
│
├── .github/
│   └── workflows/                 # GitHub Actions
│       ├── backend-ci.yml         # Backend CI/CD
│       ├── frontend-ci.yml        # Frontend CI/CD
│       ├── deploy-dev.yml         # Deploy to dev
│       ├── deploy-staging.yml     # Deploy to staging
│       └── deploy-prod.yml        # Deploy to production
│
├── .gitignore                     # Git ignore rules
├── LICENSE                        # License file
└── README.md                      # Project README
```

---

## Service Specifications

### 1. API Gateway Service

**Purpose:** Central entry point for all client requests

**Responsibilities:**
- Route requests to appropriate microservices
- Authenticate and authorize requests
- Rate limiting and throttling
- Request/response logging
- CORS handling
- API versioning
- Health checks

**Endpoints:**
```
GET  /health                    # Health check
GET  /metrics                   # Prometheus metrics
POST /api/v1/auth/login         # Forward to auth service
POST /api/v1/auth/logout        # Forward to auth service
*    /api/v1/payments/*          # Forward to payment service
*    /api/v1/tax/*               # Forward to tax service
*    /api/v1/reports/*           # Forward to reporting service
```

**Configuration:**
```yaml
server:
  port: 8080
  read_timeout: 30s
  write_timeout: 30s
  
rate_limit:
  requests_per_minute: 100
  burst: 20
  
cors:
  allowed_origins:
    - https://tax.state.gov
    - https://admin.tax.state.gov
  allowed_methods:
    - GET
    - POST
    - PUT
    - DELETE
  allowed_headers:
    - Content-Type
    - Authorization
```

---

### 2. Payment Service

**Purpose:** Handle all payment processing including XRPL blockchain

**Responsibilities:**
- XRPL payment creation and monitoring
- Traditional payment methods (ACH, credit card)
- Payment status tracking
- Transaction verification
- Payment reconciliation
- Refund processing

**Key Functions:**

```go
// CreateXRPLPayment creates a new XRPL payment request
func (s *PaymentService) CreateXRPLPayment(ctx context.Context, req CreatePaymentRequest) (*Payment, error)

// MonitorXRPLTransaction monitors an XRPL transaction until completion
func (s *PaymentService) MonitorXRPLTransaction(ctx context.Context, paymentID string) error

// GetExchangeRate fetches current XRP/USD exchange rate from multiple sources
func (s *PaymentService) GetExchangeRate(ctx context.Context) (*ExchangeRate, error)

// VerifyPayment verifies a payment on the XRPL blockchain
func (s *PaymentService) VerifyPayment(ctx context.Context, txHash string) (*VerificationResult, error)

// ProcessRefund processes a refund for a completed payment
func (s *PaymentService) ProcessRefund(ctx context.Context, paymentID string, reason string) error
```

**Database Tables:**
- `payments` - All payment records
- `xrpl_transactions` - XRPL-specific transaction data
- `xrpl_wallets` - Manufacturer and state wallet addresses
- `exchange_rates` - Historical exchange rate data
- `payment_methods` - Supported payment methods per manufacturer

**XRPL Integration:**
```go
type XRPLClient struct {
    client      *xrpl.Client
    stateWallet *xrpl.Wallet
    network     string // testnet, mainnet
}

type PaymentRequest struct {
    ManufacturerID   int64
    TaxAmountUSD     float64
    ReportID         string
    Description      string
    CallbackURL      string
}

type PaymentResponse struct {
    PaymentID        string
    XRPAmount        float64
    DestinationTag   uint32
    QRCode           string
    ExpiresAt        time.Time
    WalletAddress    string
}
```

---

### 3. Tax Service

**Purpose:** Tax calculation, report submission, and compliance

**Responsibilities:**
- Tax amount calculation based on production data
- Report submission and validation
- Compliance checking
- Tax rate management
- Form generation (PDF)
- Historical tax data

**Key Functions:**

```go
// CalculateTaxAmount calculates tax owed based on production data
func (s *TaxService) CalculateTaxAmount(ctx context.Context, report ProductionReport) (*TaxCalculation, error)

// SubmitReport submits a manufacturer's tax report
func (s *TaxService) SubmitReport(ctx context.Context, report TaxReport) (*SubmissionResult, error)

// ValidateReport validates report data against regulations
func (s *TaxService) ValidateReport(ctx context.Context, report TaxReport) ([]ValidationError, error)

// GenerateTaxForm generates a PDF tax form
func (s *TaxService) GenerateTaxForm(ctx context.Context, reportID string) ([]byte, error)
```

**Database Tables:**
- `tax_reports` - All submitted reports
- `tax_rates` - Current and historical tax rates
- `production_data` - Detailed production information
- `compliance_rules` - Business rules and validation logic
- `form_templates` - PDF form templates

---

### 4. Reporting Service

**Purpose:** Admin dashboards, analytics, and data exports

**Responsibilities:**
- Report review and approval workflow
- Dashboard metrics and KPIs
- Data visualization
- Export functionality (CSV, Excel, PDF)
- Audit trails
- Compliance reporting

**Key Functions:**

```go
// GetPendingReports retrieves reports awaiting review
func (s *ReportingService) GetPendingReports(ctx context.Context, filter ReportFilter) ([]*Report, error)

// ApproveReport approves a manufacturer's report
func (s *ReportingService) ApproveReport(ctx context.Context, reportID string, approverID int64) error

// RequestCorrection requests corrections to a report
func (s *ReportingService) RequestCorrection(ctx context.Context, reportID string, feedback string) error

// GetDashboardMetrics retrieves metrics for admin dashboard
func (s *ReportingService) GetDashboardMetrics(ctx context.Context, dateRange DateRange) (*DashboardMetrics, error)

// ExportData exports data in specified format
func (s *ReportingService) ExportData(ctx context.Context, query ExportQuery) ([]byte, error)
```

---

### 5. Notification Service

**Purpose:** Email, SMS, and webhook notifications

**Responsibilities:**
- Email notifications (payment confirmations, report approvals)
- SMS alerts for critical events
- Webhook delivery to external systems
- Template management
- Delivery tracking

**Key Functions:**

```go
// SendEmail sends an email using a template
func (s *NotificationService) SendEmail(ctx context.Context, to string, template string, data map[string]interface{}) error

// SendSMS sends an SMS message
func (s *NotificationService) SendSMS(ctx context.Context, phoneNumber string, message string) error

// SendWebhook sends a webhook to external URL
func (s *NotificationService) SendWebhook(ctx context.Context, url string, payload interface{}) error
```

---

### 6. Auth Service

**Purpose:** Authentication and authorization

**Responsibilities:**
- User login/logout
- JWT token generation and validation
- OAuth 2.0 integration with State.gov CDB
- Role-based access control (RBAC)
- Session management
- Password reset

**Key Functions:**

```go
// Login authenticates a user and returns JWT token
func (s *AuthService) Login(ctx context.Context, credentials Credentials) (*TokenPair, error)

// ValidateToken validates a JWT token and returns claims
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*Claims, error)

// RefreshToken refreshes an access token using refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)

// Logout invalidates a user's session
func (s *AuthService) Logout(ctx context.Context, userID int64, token string) error

// CheckPermission checks if user has required permission
func (s *AuthService) CheckPermission(ctx context.Context, userID int64, resource string, action string) (bool, error)
```

**JWT Claims Structure:**
```go
type Claims struct {
    UserID       int64    `json:"user_id"`
    Email        string   `json:"email"`
    Roles        []string `json:"roles"`
    Permissions  []string `json:"permissions"`
    IsAdmin      bool     `json:"is_admin"`
    StateID      string   `json:"state_id,omitempty"`
    jwt.StandardClaims
}
```

---

## Database Schema

### Core Tables

```sql
-- Users Table
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    role VARCHAR(50) NOT NULL, -- 'manufacturer', 'admin', 'super_admin'
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);

-- Manufacturers Table
CREATE TABLE manufacturers (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    company_name VARCHAR(255) NOT NULL,
    tax_id VARCHAR(50) UNIQUE NOT NULL,
    license_number VARCHAR(100) UNIQUE NOT NULL,
    business_type VARCHAR(50), -- 'brewery', 'winery', 'distillery'
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(2),
    zip_code VARCHAR(10),
    phone VARCHAR(20),
    website VARCHAR(255),
    is_approved BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_manufacturers_tax_id ON manufacturers(tax_id);
CREATE INDEX idx_manufacturers_user_id ON manufacturers(user_id);

-- Payments Table
CREATE TABLE payments (
    id BIGSERIAL PRIMARY KEY,
    manufacturer_id BIGINT REFERENCES manufacturers(id),
    report_id BIGINT,
    payment_method VARCHAR(50) NOT NULL, -- 'xrpl', 'ach', 'credit_card'
    amount_usd DECIMAL(15, 2) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'processing', 'completed', 'failed', 'refunded'
    transaction_id VARCHAR(255),
    confirmation_number VARCHAR(100),
    payment_date TIMESTAMP,
    processed_at TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payments_manufacturer_id ON payments(manufacturer_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_payment_date ON payments(payment_date);

-- XRPL Payments Table
CREATE TABLE xrpl_payments (
    id BIGSERIAL PRIMARY KEY,
    payment_id BIGINT REFERENCES payments(id),
    xrp_amount DECIMAL(20, 6) NOT NULL,
    exchange_rate DECIMAL(10, 4) NOT NULL,
    destination_address VARCHAR(100) NOT NULL,
    destination_tag INT,
    source_address VARCHAR(100),
    tx_hash VARCHAR(100) UNIQUE,
    ledger_index BIGINT,
    fee_xrp DECIMAL(10, 6),
    status VARCHAR(50) DEFAULT 'pending',
    qr_code TEXT,
    expires_at TIMESTAMP,
    confirmed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_xrpl_payments_payment_id ON xrpl_payments(payment_id);
CREATE INDEX idx_xrpl_payments_tx_hash ON xrpl_payments(tx_hash);
CREATE INDEX idx_xrpl_payments_status ON xrpl_payments(status);

-- Tax Reports Table
CREATE TABLE tax_reports (
    id BIGSERIAL PRIMARY KEY,
    manufacturer_id BIGINT REFERENCES manufacturers(id),
    report_type VARCHAR(50) NOT NULL, -- 'monthly', 'annual'
    form_number VARCHAR(20), -- '35-7136', '35-7131', etc.
    report_period_start DATE NOT NULL,
    report_period_end DATE NOT NULL,
    submission_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) DEFAULT 'pending', -- 'pending', 'under_review', 'approved', 'needs_correction', 'rejected'
    reviewer_id BIGINT REFERENCES users(id),
    reviewed_at TIMESTAMP,
    review_notes TEXT,
    production_data JSONB,
    tax_amount_calculated DECIMAL(15, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tax_reports_manufacturer_id ON tax_reports(manufacturer_id);
CREATE INDEX idx_tax_reports_status ON tax_reports(status);
CREATE INDEX idx_tax_reports_report_period ON tax_reports(report_period_start, report_period_end);

-- Tax Rates Table
CREATE TABLE tax_rates (
    id BIGSERIAL PRIMARY KEY,
    product_type VARCHAR(50) NOT NULL, -- 'beer', 'wine', 'spirits'
    rate_per_unit DECIMAL(10, 4) NOT NULL,
    unit_type VARCHAR(50), -- 'gallon', 'barrel', 'case'
    effective_date DATE NOT NULL,
    expiration_date DATE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tax_rates_product_type ON tax_rates(product_type);
CREATE INDEX idx_tax_rates_effective_date ON tax_rates(effective_date);

-- Exchange Rates Table (XRPL)
CREATE TABLE exchange_rates (
    id BIGSERIAL PRIMARY KEY,
    source VARCHAR(50) NOT NULL, -- 'coinbase', 'binance', 'kraken', 'bitstamp'
    xrp_usd_rate DECIMAL(10, 6) NOT NULL,
    bid DECIMAL(10, 6),
    ask DECIMAL(10, 6),
    volume_24h DECIMAL(20, 2),
    timestamp TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_exchange_rates_timestamp ON exchange_rates(timestamp DESC);
CREATE INDEX idx_exchange_rates_source ON exchange_rates(source);

-- Audit Log Table
CREATE TABLE audit_log (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    action VARCHAR(100) NOT NULL, -- 'login', 'payment_created', 'report_approved', etc.
    resource_type VARCHAR(50), -- 'payment', 'report', 'user'
    resource_id BIGINT,
    ip_address INET,
    user_agent TEXT,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_log_user_id ON audit_log(user_id);
CREATE INDEX idx_audit_log_action ON audit_log(action);
CREATE INDEX idx_audit_log_created_at ON audit_log(created_at DESC);

-- Sessions Table (for Redis backup)
CREATE TABLE sessions (
    id VARCHAR(255) PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    expires_at TIMESTAMP NOT NULL,
    data JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
```

### Views

```sql
-- Dashboard Metrics View
CREATE VIEW dashboard_metrics AS
SELECT 
    COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_reports,
    COUNT(CASE WHEN status = 'needs_correction' THEN 1 END) as needs_correction,
    COUNT(CASE WHEN status = 'approved' AND DATE_TRUNC('month', reviewed_at) = DATE_TRUNC('month', CURRENT_DATE) THEN 1 END) as approved_this_month,
    COUNT(CASE WHEN status = 'approved' AND EXTRACT(YEAR FROM reviewed_at) = EXTRACT(YEAR FROM CURRENT_DATE) THEN 1 END) as approved_this_year,
    SUM(CASE WHEN status = 'approved' AND DATE_TRUNC('month', reviewed_at) = DATE_TRUNC('month', CURRENT_DATE) THEN tax_amount_calculated ELSE 0 END) as tax_collected_this_month,
    SUM(CASE WHEN status = 'approved' AND EXTRACT(YEAR FROM reviewed_at) = EXTRACT(YEAR FROM CURRENT_DATE) THEN tax_amount_calculated ELSE 0 END) as tax_collected_this_year
FROM tax_reports;

-- Payment Summary View
CREATE VIEW payment_summary AS
SELECT 
    p.manufacturer_id,
    m.company_name,
    COUNT(p.id) as total_payments,
    SUM(CASE WHEN p.status = 'completed' THEN p.amount_usd ELSE 0 END) as total_paid,
    SUM(CASE WHEN p.status = 'pending' THEN p.amount_usd ELSE 0 END) as pending_amount,
    COUNT(CASE WHEN p.payment_method = 'xrpl' THEN 1 END) as xrpl_payment_count,
    AVG(CASE WHEN p.payment_method = 'xrpl' AND p.status = 'completed' THEN EXTRACT(EPOCH FROM (p.processed_at - p.created_at)) END) as avg_xrpl_processing_time_seconds
FROM payments p
JOIN manufacturers m ON p.manufacturer_id = m.id
GROUP BY p.manufacturer_id, m.company_name;
```

---

## API Specifications

### Authentication

All API requests (except `/api/v1/auth/login` and `/api/v1/auth/register`) must include an Authorization header:

```
Authorization: Bearer <JWT_TOKEN>
```

### Common Response Structure

**Success Response:**
```json
{
  "success": true,
  "data": {
    // Response data
  },
  "meta": {
    "timestamp": "2025-12-29T18:30:00Z",
    "request_id": "req_abc123"
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "error": {
    "code": "INVALID_INPUT",
    "message": "Invalid tax amount",
    "details": {
      "field": "amount_usd",
      "reason": "must be greater than 0"
    }
  },
  "meta": {
    "timestamp": "2025-12-29T18:30:00Z",
    "request_id": "req_abc123"
  }
}
```

### API Endpoints

#### Authentication Endpoints

```
POST /api/v1/auth/register
POST /api/v1/auth/login
POST /api/v1/auth/logout
POST /api/v1/auth/refresh
POST /api/v1/auth/forgot-password
POST /api/v1/auth/reset-password
GET  /api/v1/auth/me
```

**Example: Login**
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "manufacturer@example.com",
  "password": "SecurePassword123!"
}

Response 200:
{
  "success": true,
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "token_type": "Bearer",
    "expires_in": 3600,
    "user": {
      "id": 123,
      "email": "manufacturer@example.com",
      "first_name": "John",
      "last_name": "Doe",
      "role": "manufacturer"
    }
  }
}
```

#### Payment Endpoints

```
POST   /api/v1/payments/xrpl/create
GET    /api/v1/payments/xrpl/:paymentId/status
POST   /api/v1/payments/xrpl/:paymentId/verify
GET    /api/v1/payments/xrpl/exchange-rate
POST   /api/v1/payments/xrpl/convert
GET    /api/v1/payments
GET    /api/v1/payments/:paymentId
POST   /api/v1/payments/:paymentId/refund
```

**Example: Create XRPL Payment**
```http
POST /api/v1/payments/xrpl/create
Authorization: Bearer <token>
Content-Type: application/json

{
  "manufacturer_id": 123,
  "tax_amount_usd": 17500.00,
  "report_id": "RPT-2025-12-001",
  "description": "December 2025 Excise Tax Payment",
  "callback_url": "https://manufacturer.com/webhook/payment"
}

Response 201:
{
  "success": true,
  "data": {
    "payment_id": "PAY-XRP-20251229-ABC123",
    "xrp_amount": 25000.00,
    "exchange_rate": 0.7000,
    "destination_address": "rN7n7otQDd6FczFgLdCqZLYjmHG3A9LKr7",
    "destination_tag": 123456,
    "qr_code": "data:image/png;base64,iVBORw0KGgoA...",
    "expires_at": "2025-12-30T18:30:00Z",
    "status": "pending",
    "instructions": "Send exactly 25000.00 XRP to the address above with destination tag 123456"
  }
}
```

**Example: Get Exchange Rate**
```http
GET /api/v1/payments/xrpl/exchange-rate
Authorization: Bearer <token>

Response 200:
{
  "success": true,
  "data": {
    "xrp_usd": 0.7000,
    "usd_xrp": 1.4286,
    "sources": [
      {
        "name": "coinbase",
        "rate": 0.7005,
        "last_updated": "2025-12-29T18:29:45Z"
      },
      {
        "name": "binance",
        "rate": 0.6998,
        "last_updated": "2025-12-29T18:29:50Z"
      },
      {
        "name": "kraken",
        "rate": 0.6997,
        "last_updated": "2025-12-29T18:29:48Z"
      }
    ],
    "median_rate": 0.7000,
    "last_updated": "2025-12-29T18:29:50Z"
  }
}
```

#### Tax Report Endpoints

```
POST   /api/v1/reports
GET    /api/v1/reports
GET    /api/v1/reports/:reportId
PUT    /api/v1/reports/:reportId
DELETE /api/v1/reports/:reportId
POST   /api/v1/reports/:reportId/submit
GET    /api/v1/reports/:reportId/pdf
POST   /api/v1/reports/:reportId/validate
```

**Example: Submit Tax Report**
```http
POST /api/v1/reports
Authorization: Bearer <token>
Content-Type: application/json

{
  "manufacturer_id": 123,
  "report_type": "monthly",
  "form_number": "35-7136",
  "report_period_start": "2025-12-01",
  "report_period_end": "2025-12-31",
  "production_data": {
    "total_gallons_produced": 50000,
    "total_gallons_sold": 45000,
    "total_gallons_inventory": 25000,
    "products": [
      {
        "name": "Pale Ale",
        "sku": "PA-001",
        "gallons_produced": 30000,
        "abv": 5.5,
        "container_size": "12oz"
      },
      {
        "name": "IPA",
        "sku": "IPA-001",
        "gallons_produced": 20000,
        "abv": 6.8,
        "container_size": "16oz"
      }
    ]
  }
}

Response 201:
{
  "success": true,
  "data": {
    "report_id": "RPT-2025-12-001",
    "status": "pending",
    "tax_amount_calculated": 17500.00,
    "submission_date": "2025-12-29T18:30:00Z",
    "next_steps": "Your report is pending review. You will be notified once it has been reviewed."
  }
}
```

#### Admin Endpoints

```
GET    /api/v1/admin/reports/pending
GET    /api/v1/admin/reports/:reportId
POST   /api/v1/admin/reports/:reportId/approve
POST   /api/v1/admin/reports/:reportId/request-correction
POST   /api/v1/admin/reports/:reportId/reject
GET    /api/v1/admin/dashboard
GET    /api/v1/admin/users
GET    /api/v1/admin/manufacturers
POST   /api/v1/admin/manufacturers/:id/approve
GET    /api/v1/admin/payments
GET    /api/v1/admin/audit-log
POST   /api/v1/admin/export
```

---

## Authentication & Authorization

### JWT Token Structure

**Access Token (Short-lived: 1 hour):**
```json
{
  "sub": "123",
  "email": "user@example.com",
  "roles": ["manufacturer"],
  "permissions": ["payment:create", "report:submit", "report:read:own"],
  "iat": 1703865600,
  "exp": 1703869200
}
```

**Refresh Token (Long-lived: 7 days):**
```json
{
  "sub": "123",
  "type": "refresh",
  "iat": 1703865600,
  "exp": 1704470400
}
```

### Role-Based Access Control (RBAC)

**Roles:**
- `manufacturer` - Can create payments, submit reports, view own data
- `admin` - Can review reports, approve manufacturers, view all data
- `super_admin` - Full system access including user management

**Permissions Matrix:**

| Resource | Action | Manufacturer | Admin | Super Admin |
|----------|--------|--------------|-------|-------------|
| Payments | Create | ✅ (own) | ❌ | ✅ |
| Payments | Read | ✅ (own) | ✅ | ✅ |
| Payments | Refund | ❌ | ✅ | ✅ |
| Reports | Create | ✅ | ❌ | ✅ |
| Reports | Read | ✅ (own) | ✅ | ✅ |
| Reports | Approve | ❌ | ✅ | ✅ |
| Users | Read | ✅ (self) | ✅ | ✅ |
| Users | Create | ❌ | ❌ | ✅ |
| Users | Update | ✅ (self) | ✅ (non-admin) | ✅ |
| Users | Delete | ❌ | ❌ | ✅ |

### OAuth 2.0 Integration (State.gov CDB)

**Flow:** Authorization Code Grant with PKCE

**Configuration:**
```yaml
oauth:
  provider: state_gov_cdb
  authorization_url: https://login.state.gov/oauth2/authorize
  token_url: https://login.state.gov/oauth2/token
  userinfo_url: https://login.state.gov/oauth2/userinfo
  client_id: ${CDB_CLIENT_ID}
  client_secret: ${CDB_CLIENT_SECRET}
  redirect_uri: https://tax.state.gov/auth/callback
  scopes:
    - openid
    - profile
    - email
```

**Implementation:**
```go
func (s *AuthService) InitiateOAuthLogin(ctx context.Context) (string, error) {
    // Generate state and PKCE verifier
    state := generateRandomString(32)
    verifier := generateRandomString(64)
    challenge := base64URLEncode(sha256(verifier))
    
    // Store state and verifier in Redis
    s.cache.Set(ctx, "oauth:state:"+state, verifier, 10*time.Minute)
    
    // Build authorization URL
    authURL := fmt.Sprintf(
        "%s?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s&code_challenge=%s&code_challenge_method=S256",
        s.config.OAuth.AuthorizationURL,
        s.config.OAuth.ClientID,
        url.QueryEscape(s.config.OAuth.RedirectURI),
        url.QueryEscape(strings.Join(s.config.OAuth.Scopes, " ")),
        state,
        challenge,
    )
    
    return authURL, nil
}

func (s *AuthService) HandleOAuthCallback(ctx context.Context, code string, state string) (*TokenPair, error) {
    // Retrieve and validate state
    verifier, err := s.cache.Get(ctx, "oauth:state:"+state)
    if err != nil {
        return nil, errors.New("invalid state")
    }
    
    // Exchange code for token
    resp, err := s.httpClient.PostForm(s.config.OAuth.TokenURL, url.Values{
        "grant_type":    {"authorization_code"},
        "code":          {code},
        "redirect_uri":  {s.config.OAuth.RedirectURI},
        "client_id":     {s.config.OAuth.ClientID},
        "client_secret": {s.config.OAuth.ClientSecret},
        "code_verifier": {verifier},
    })
    // ... handle token response and create user session
}
```

---

## XRPL Blockchain Integration

### Overview

The XRPL integration enables near-zero-cost payment processing by leveraging the XRP Ledger blockchain. Payments settle in 3-5 seconds with fees of approximately $0.0003.

### Architecture

```
┌────────────────────────────────────────────────────────────┐
│                  Payment Service (Golang)                   │
├────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────────────────────────────┐  │
│  │            XRPL Payment Manager                      │  │
│  │  • CreatePaymentRequest                              │  │
│  │  • MonitorTransaction                                │  │
│  │  • VerifyPayment                                     │  │
│  └─────────────┬────────────────────────────────────────┘  │
│                │                                            │
│  ┌─────────────▼──────────┐  ┌────────────────────────┐   │
│  │  XRPL Client Wrapper   │  │  Price Oracle Service  │   │
│  │  • Connect to network  │  │  • Multi-source rates  │   │
│  │  • Submit transactions │  │  • Outlier detection   │   │
│  │  • Subscribe to events │  │  • 30s refresh cycle   │   │
│  └─────────────┬──────────┘  └───────────┬────────────┘   │
└────────────────┼─────────────────────────┼────────────────┘
                 │                         │
                 │                         │
                 ▼                         ▼
        ┌────────────────┐      ┌──────────────────┐
        │  XRPL Network  │      │ Exchange APIs     │
        │  (Mainnet)     │      │ • Coinbase        │
        │                │      │ • Binance         │
        │  State Wallet  │      │ • Kraken          │
        │  rN7n7otQDd... │      │ • Bitstamp        │
        └────────────────┘      └──────────────────┘
```

### XRPL Client Implementation

```go
package xrpl

import (
    "context"
    "time"
    
    "github.com/rubblelabs/ripple/websockets"
    "github.com/rubblelabs/ripple/data"
)

type Client struct {
    ws          *websockets.Remote
    stateWallet *data.Account
    network     string
    logger      *logger.Logger
}

type Config struct {
    WSEndpoint      string
    StateAddress    string
    StateSecret     string
    Network         string // "testnet", "mainnet"
    ReconnectDelay  time.Duration
}

func NewClient(cfg Config) (*Client, error) {
    // Connect to XRPL network
    ws, err := websockets.NewRemote(cfg.WSEndpoint)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to XRPL: %w", err)
    }
    
    // Load state wallet
    wallet, err := data.NewAccountFromSecret(cfg.StateSecret, data.SECP256K1)
    if err != nil {
        return nil, fmt.Errorf("failed to load wallet: %w", err)
    }
    
    client := &Client{
        ws:          ws,
        stateWallet: wallet,
        network:     cfg.Network,
        logger:      logger.New("xrpl-client"),
    }
    
    // Verify connection
    if err := client.Ping(context.Background()); err != nil {
        return nil, fmt.Errorf("connection check failed: %w", err)
    }
    
    client.logger.Info("XRPL client connected successfully",
        "network", cfg.Network,
        "wallet", wallet.Address(),
    )
    
    return client, nil
}

func (c *Client) CreatePaymentRequest(ctx context.Context, req PaymentRequest) (*PaymentResponse, error) {
    // Get current exchange rate
    rate, err := c.GetExchangeRate(ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to get exchange rate: %w", err)
    }
    
    // Convert USD to XRP
    xrpAmount := req.AmountUSD / rate.XRPUSD
    
    // Generate unique destination tag
    destinationTag := c.generateDestinationTag(req.ManufacturerID, req.ReportID)
    
    // Create payment response
    response := &PaymentResponse{
        PaymentID:       c.generatePaymentID(),
        XRPAmount:       xrpAmount,
        ExchangeRate:    rate.XRPUSD,
        WalletAddress:   c.stateWallet.Address().String(),
        DestinationTag:  destinationTag,
        ExpiresAt:       time.Now().Add(24 * time.Hour),
        Status:          "pending",
    }
    
    // Generate QR code for payment
    response.QRCode = c.generateQRCode(response)
    
    return response, nil
}

func (c *Client) MonitorTransaction(ctx context.Context, paymentID string, destinationTag uint32) error {
    // Subscribe to account transactions
    subscription := &websockets.SubscribeCommand{
        Accounts: []data.Account{*c.stateWallet},
    }
    
    if err := c.ws.Subscribe(subscription); err != nil {
        return fmt.Errorf("failed to subscribe: %w", err)
    }
    
    // Monitor for incoming transactions
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
            
        case msg := <-c.ws.Incoming:
            if tx, ok := msg.(*websockets.TransactionStreamMsg); ok {
                if c.isMatchingPayment(tx, destinationTag) {
                    // Verify and process payment
                    return c.processIncomingPayment(ctx, paymentID, tx)
                }
            }
        }
    }
}

func (c *Client) VerifyPayment(ctx context.Context, txHash string) (*VerificationResult, error) {
    // Get transaction from ledger
    result, err := c.ws.Tx(txHash)
    if err != nil {
        return nil, fmt.Errorf("failed to get transaction: %w", err)
    }
    
    // Verify transaction details
    verification := &VerificationResult{
        TxHash:       txHash,
        Confirmed:    result.Validated,
        LedgerIndex:  result.LedgerIndex,
        Amount:       result.Transaction.GetAmount(),
        Fee:          result.Transaction.GetFee(),
        Timestamp:    result.CloseTime,
    }
    
    return verification, nil
}

type PaymentRequest struct {
    ManufacturerID int64
    ReportID       string
    AmountUSD      float64
    Description    string
}

type PaymentResponse struct {
    PaymentID      string
    XRPAmount      float64
    ExchangeRate   float64
    WalletAddress  string
    DestinationTag uint32
    QRCode         string
    ExpiresAt      time.Time
    Status         string
}

type VerificationResult struct {
    TxHash      string
    Confirmed   bool
    LedgerIndex uint32
    Amount      data.Amount
    Fee         data.Value
    Timestamp   time.Time
}
```

### Price Oracle Implementation

```go
package xrpl

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "sort"
    "sync"
    "time"
)

type PriceOracle struct {
    sources map[string]ExchangeSource
    cache   *ExchangeRate
    mu      sync.RWMutex
    logger  *logger.Logger
}

type ExchangeSource interface {
    GetRate(ctx context.Context) (*SourceRate, error)
    Name() string
}

type SourceRate struct {
    Source      string
    Rate        float64
    Bid         float64
    Ask         float64
    Volume24h   float64
    LastUpdated time.Time
}

type ExchangeRate struct {
    XRPUSD     float64
    USDXRP     float64
    Sources    []SourceRate
    MedianRate float64
    Updated    time.Time
}

func NewPriceOracle(cfg OracleConfig) *PriceOracle {
    oracle := &PriceOracle{
        sources: make(map[string]ExchangeSource),
        logger:  logger.New("price-oracle"),
    }
    
    // Register exchange sources
    oracle.sources["coinbase"] = NewCoinbaseSource(cfg.Coinbase)
    oracle.sources["binance"] = NewBinanceSource(cfg.Binance)
    oracle.sources["kraken"] = NewKrakenSource(cfg.Kraken)
    oracle.sources["bitstamp"] = NewBitstampSource(cfg.Bitstamp)
    
    // Start refresh cycle
    go oracle.startRefreshCycle(cfg.RefreshInterval)
    
    return oracle
}

func (o *PriceOracle) GetExchangeRate(ctx context.Context) (*ExchangeRate, error) {
    o.mu.RLock()
    defer o.mu.RUnlock()
    
    if o.cache == nil {
        return nil, fmt.Errorf("exchange rate not initialized")
    }
    
    // Check if cache is stale
    if time.Since(o.cache.Updated) > 2*time.Minute {
        return nil, fmt.Errorf("exchange rate is stale")
    }
    
    return o.cache, nil
}

func (o *PriceOracle) startRefreshCycle(interval time.Duration) {
    ticker := time.NewTicker(interval)
    defer ticker.Stop()
    
    // Initial fetch
    o.refreshRates()
    
    for range ticker.C {
        o.refreshRates()
    }
}

func (o *PriceOracle) refreshRates() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    var wg sync.WaitGroup
    sourceRates := make(chan SourceRate, len(o.sources))
    
    // Fetch from all sources concurrently
    for _, source := range o.sources {
        wg.Add(1)
        go func(s ExchangeSource) {
            defer wg.Done()
            rate, err := s.GetRate(ctx)
            if err != nil {
                o.logger.Error("failed to get rate from source",
                    "source", s.Name(),
                    "error", err,
                )
                return
            }
            sourceRates <- *rate
        }(source)
    }
    
    // Wait for all sources
    go func() {
        wg.Wait()
        close(sourceRates)
    }()
    
    // Collect rates
    var rates []float64
    var sources []SourceRate
    for rate := range sourceRates {
        rates = append(rates, rate.Rate)
        sources = append(sources, rate)
    }
    
    if len(rates) < 2 {
        o.logger.Error("insufficient exchange sources",
            "available", len(rates),
            "required", 2,
        )
        return
    }
    
    // Calculate median rate (more resistant to outliers than average)
    sort.Float64s(rates)
    var medianRate float64
    mid := len(rates) / 2
    if len(rates)%2 == 0 {
        medianRate = (rates[mid-1] + rates[mid]) / 2
    } else {
        medianRate = rates[mid]
    }
    
    // Remove outliers (rates that deviate >5% from median)
    var filteredRates []float64
    var filteredSources []SourceRate
    for i, rate := range rates {
        deviation := abs(rate-medianRate) / medianRate
        if deviation <= 0.05 { // 5% threshold
            filteredRates = append(filteredRates, rate)
            filteredSources = append(filteredSources, sources[i])
        } else {
            o.logger.Warn("outlier rate detected",
                "source", sources[i].Source,
                "rate", rate,
                "median", medianRate,
                "deviation", deviation*100,
            )
        }
    }
    
    // Update cache
    o.mu.Lock()
    o.cache = &ExchangeRate{
        XRPUSD:     medianRate,
        USDXRP:     1 / medianRate,
        Sources:    filteredSources,
        MedianRate: medianRate,
        Updated:    time.Now(),
    }
    o.mu.Unlock()
    
    o.logger.Info("exchange rate updated",
        "rate", medianRate,
        "sources", len(filteredSources),
    )
}

// Coinbase exchange source
type CoinbaseSource struct {
    apiKey string
    client *http.Client
}

func (c *CoinbaseSource) GetRate(ctx context.Context) (*SourceRate, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", "https://api.coinbase.com/v2/exchange-rates?currency=XRP", nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := c.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    var data struct {
        Data struct {
            Currency string
            Rates    map[string]string
        }
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
        return nil, err
    }
    
    rate, err := parseFloat(data.Data.Rates["USD"])
    if err != nil {
        return nil, err
    }
    
    return &SourceRate{
        Source:      "coinbase",
        Rate:        rate,
        LastUpdated: time.Now(),
    }, nil
}

func (c *CoinbaseSource) Name() string {
    return "coinbase"
}
```

### Payment Monitoring Service

```go
package xrpl

import (
    "context"
    "sync"
    "time"
)

type MonitorService struct {
    xrplClient *Client
    db         *database.DB
    cache      *cache.Redis
    logger     *logger.Logger
    monitors   map[string]context.CancelFunc
    mu         sync.RWMutex
}

func NewMonitorService(xrplClient *Client, db *database.DB, cache *cache.Redis) *MonitorService {
    return &MonitorService{
        xrplClient: xrplClient,
        db:         db,
        cache:      cache,
        logger:     logger.New("payment-monitor"),
        monitors:   make(map[string]context.CancelFunc),
    }
}

func (s *MonitorService) StartMonitoring(ctx context.Context, paymentID string, destinationTag uint32, timeout time.Duration) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    // Check if already monitoring
    if _, exists := s.monitors[paymentID]; exists {
        return fmt.Errorf("already monitoring payment %s", paymentID)
    }
    
    // Create cancellable context
    monitorCtx, cancel := context.WithTimeout(ctx, timeout)
    s.monitors[paymentID] = cancel
    
    // Start monitoring in goroutine
    go func() {
        defer func() {
            s.mu.Lock()
            delete(s.monitors, paymentID)
            s.mu.Unlock()
        }()
        
        err := s.xrplClient.MonitorTransaction(monitorCtx, paymentID, destinationTag)
        if err != nil {
            s.logger.Error("payment monitoring failed",
                "payment_id", paymentID,
                "error", err,
            )
            
            // Update payment status to expired/failed
            s.updatePaymentStatus(context.Background(), paymentID, "expired")
            return
        }
        
        s.logger.Info("payment received and verified",
            "payment_id", paymentID,
        )
        
        // Update payment status to completed
        s.updatePaymentStatus(context.Background(), paymentID, "completed")
        
        // Send notification
        s.sendPaymentNotification(context.Background(), paymentID)
    }()
    
    return nil
}

func (s *MonitorService) StopMonitoring(paymentID string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    if cancel, exists := s.monitors[paymentID]; exists {
        cancel()
        delete(s.monitors, paymentID)
    }
}

func (s *MonitorService) updatePaymentStatus(ctx context.Context, paymentID string, status string) error {
    query := `
        UPDATE xrpl_payments
        SET status = $1, updated_at = NOW()
        WHERE payment_id = (SELECT id FROM payments WHERE id = $2)
    `
    _, err := s.db.ExecContext(ctx, query, status, paymentID)
    return err
}

func (s *MonitorService) sendPaymentNotification(ctx context.Context, paymentID string) error {
    // Fetch payment details
    payment, err := s.getPaymentDetails(ctx, paymentID)
    if err != nil {
        return err
    }
    
    // Send email notification
    // Send webhook notification
    // Update dashboard
    
    return nil
}
```

---

## Docker Configuration

### Multi-Stage Dockerfile for Backend

```dockerfile
# infrastructure/docker/Dockerfile.backend

# Stage 1: Build
FROM golang:1.21-alpine AS builder

# Install dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY backend/go.mod backend/go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY backend/ ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/api-gateway ./cmd/api-gateway
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/payment-service ./cmd/payment-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/tax-service ./cmd/tax-service
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/bin/reporting-service ./cmd/reporting-service

# Stage 2: Runtime
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 app && \
    adduser -D -u 1000 -G app app

# Set working directory
WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/bin/ ./
COPY --from=builder /app/migrations/ ./migrations/
COPY --from=builder /app/configs/ ./configs/

# Change ownership
RUN chown -R app:app /app

# Switch to non-root user
USER app

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Default command (can be overridden)
CMD ["./api-gateway"]
```

### Multi-Stage Dockerfile for Frontend

```dockerfile
# infrastructure/docker/Dockerfile.frontend

# Stage 1: Build
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files
COPY frontend/package*.json ./

# Install dependencies
RUN npm ci

# Copy source code
COPY frontend/ ./

# Build application
RUN npm run build

# Stage 2: Runtime (Nginx)
FROM nginx:alpine

# Copy custom nginx config
COPY infrastructure/docker/nginx.conf /etc/nginx/nginx.conf

# Copy built assets
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy entrypoint script for runtime config injection
COPY infrastructure/docker/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/health || exit 1

# Start nginx
ENTRYPOINT ["/docker-entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
```

### Docker Compose for Local Development

```yaml
# infrastructure/docker/docker-compose.yml

version: '3.9'

services:
  # PostgreSQL Database
  postgres:
    image: postgres:15-alpine
    container_name: excise-tax-postgres
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${DB_PASSWORD:-postgres}
      POSTGRES_DB: excise_tax_portal
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ../backend/migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - excise-tax-network

  # Redis Cache
  redis:
    image: redis:7-alpine
    container_name: excise-tax-redis
    ports:
      - "6379:6379"
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - excise-tax-network

  # RabbitMQ Message Queue
  rabbitmq:
    image: rabbitmq:3-management-alpine
    container_name: excise-tax-rabbitmq
    environment:
      RABBITMQ_DEFAULT_USER: ${RABBITMQ_USER:-admin}
      RABBITMQ_DEFAULT_PASS: ${RABBITMQ_PASSWORD:-admin}
    ports:
      - "5672:5672"
      - "15672:15672"
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
    healthcheck:
      test: ["CMD", "rabbitmq-diagnostics", "ping"]
      interval: 30s
      timeout: 10s
      retries: 5
    networks:
      - excise-tax-network

  # API Gateway
  api-gateway:
    build:
      context: ../../
      dockerfile: infrastructure/docker/Dockerfile.backend
    container_name: excise-tax-api-gateway
    command: ./api-gateway
    environment:
      - APP_ENV=development
      - APP_PORT=8080
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=${DB_PASSWORD:-postgres}
      - DB_NAME=excise_tax_portal
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - RABBITMQ_HOST=rabbitmq
      - RABBITMQ_PORT=5672
      - RABBITMQ_USER=${RABBITMQ_USER:-admin}
      - RABBITMQ_PASSWORD=${RABBITMQ_PASSWORD:-admin}
      - JWT_SECRET=${JWT_SECRET:-your-secret-key-change-in-production}
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
      rabbitmq:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    networks:
      - excise-tax-network

  # Payment Service
  payment-service:
    build:
      context: ../../
      dockerfile: infrastructure/docker/Dockerfile.backend
    container_name: excise-tax-payment-service
    command: ./payment-service
    environment:
      - APP_ENV=development
      - APP_PORT=8081
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=${DB_PASSWORD:-postgres}
      - DB_NAME=excise_tax_portal
      - REDIS_HOST=redis
      - REDIS_PORT=6379
      - XRPL_NETWORK=${XRPL_NETWORK:-testnet}
      - XRPL_WS_ENDPOINT=${XRPL_WS_ENDPOINT:-wss://s.altnet.rippletest.net:51233}
      - XRPL_STATE_ADDRESS=${XRPL_STATE_ADDRESS}
      - XRPL_STATE_SECRET=${XRPL_STATE_SECRET}
    ports:
      - "8081:8081"
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - excise-tax-network

  # Tax Service
  tax-service:
    build:
      context: ../../
      dockerfile: infrastructure/docker/Dockerfile.backend
    container_name: excise-tax-tax-service
    command: ./tax-service
    environment:
      - APP_ENV=development
      - APP_PORT=8082
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=${DB_PASSWORD:-postgres}
      - DB_NAME=excise_tax_portal
    ports:
      - "8082:8082"
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - excise-tax-network

  # Reporting Service
  reporting-service:
    build:
      context: ../../
      dockerfile: infrastructure/docker/Dockerfile.backend
    container_name: excise-tax-reporting-service
    command: ./reporting-service
    environment:
      - APP_ENV=development
      - APP_PORT=8083
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=${DB_PASSWORD:-postgres}
      - DB_NAME=excise_tax_portal
    ports:
      - "8083:8083"
    depends_on:
      postgres:
        condition: service_healthy
    networks:
      - excise-tax-network

  # Frontend
  frontend:
    build:
      context: ../../
      dockerfile: infrastructure/docker/Dockerfile.frontend
    container_name: excise-tax-frontend
    environment:
      - VITE_API_BASE_URL=http://localhost:8080/api/v1
      - VITE_WS_URL=ws://localhost:8080/ws
    ports:
      - "3000:80"
    depends_on:
      - api-gateway
    networks:
      - excise-tax-network

volumes:
  postgres_data:
  redis_data:
  rabbitmq_data:

networks:
  excise-tax-network:
    driver: bridge
```

---

## CI/CD Pipeline

### GitHub Actions Workflow

```yaml
# .github/workflows/backend-ci.yml

name: Backend CI/CD

on:
  push:
    branches: [main, develop]
    paths:
      - 'backend/**'
      - 'infrastructure/**'
      - '.github/workflows/backend-ci.yml'
  pull_request:
    branches: [main, develop]
    paths:
      - 'backend/**'

env:
  GO_VERSION: '1.21'
  DOCKER_REGISTRY: ghcr.io
  IMAGE_NAME: excise-tax-portal

jobs:
  # Test Job
  test:
    name: Test
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: excise_tax_test
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432
      
      redis:
        image: redis:7
        options: >-
          --health-cmd "redis-cli ping"
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 6379:6379

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ env.GO_VERSION }}
          cache-dependency-path: backend/go.sum

      - name: Install dependencies
        working-directory: backend
        run: |
          go mod download
          go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

      - name: Run linter
        working-directory: backend
        run: golangci-lint run --timeout=5m

      - name: Run tests
        working-directory: backend
        env:
          DB_HOST: localhost
          DB_PORT: 5432
          DB_USER: postgres
          DB_PASSWORD: postgres
          DB_NAME: excise_tax_test
          REDIS_HOST: localhost
          REDIS_PORT: 6379
        run: |
          go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
          go tool cover -func=coverage.out

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./backend/coverage.out
          flags: backend

  # Build Job
  build:
    name: Build Docker Images
    runs-on: ubuntu-latest
    needs: test
    if: github.event_name == 'push'
    
    strategy:
      matrix:
        service: [api-gateway, payment-service, tax-service, reporting-service]

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Login to GitHub Container Registry
        uses: docker/login-action@v3
        with:
          registry: ${{ env.DOCKER_REGISTRY }}
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}

      - name: Extract metadata
        id: meta
        uses: docker/metadata-action@v5
        with:
          images: ${{ env.DOCKER_REGISTRY }}/${{ github.repository }}/${{ matrix.service }}
          tags: |
            type=ref,event=branch
            type=ref,event=pr
            type=semver,pattern={{version}}
            type=semver,pattern={{major}}.{{minor}}
            type=sha,prefix={{branch}}-

      - name: Build and push
        uses: docker/build-push-action@v5
        with:
          context: .
          file: infrastructure/docker/Dockerfile.backend
          push: true
          tags: ${{ steps.meta.outputs.tags }}
          labels: ${{ steps.meta.outputs.labels }}
          cache-from: type=gha
          cache-to: type=gha,mode=max
          build-args: |
            SERVICE=${{ matrix.service }}

  # Deploy to Development
  deploy-dev:
    name: Deploy to Development
    runs-on: ubuntu-latest
    needs: build
    if: github.ref == 'refs/heads/develop' && github.event_name == 'push'
    environment:
      name: development
      url: https://dev.tax.state.gov

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Configure kubectl
        uses: azure/k8s-set-context@v3
        with:
          method: kubeconfig
          kubeconfig: ${{ secrets.DEV_KUBECONFIG }}

      - name: Deploy to Kubernetes
        run: |
          kubectl apply -f infrastructure/kubernetes/base/
          kubectl apply -f infrastructure/kubernetes/deployments/
          kubectl rollout status deployment/api-gateway -n excise-tax-dev
          kubectl rollout status deployment/payment-service -n excise-tax-dev

      - name: Run smoke tests
        run: |
          ./scripts/smoke-test.sh https://dev.tax.state.gov

  # Deploy to Production
  deploy-prod:
    name: Deploy to Production
    runs-on: ubuntu-latest
    needs: build
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    environment:
      name: production
      url: https://tax.state.gov

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Configure kubectl
        uses: azure/k8s-set-context@v3
        with:
          method: kubeconfig
          kubeconfig: ${{ secrets.PROD_KUBECONFIG }}

      - name: Deploy to Kubernetes
        run: |
          kubectl apply -f infrastructure/kubernetes/base/
          kubectl apply -f infrastructure/kubernetes/deployments/
          kubectl rollout status deployment/api-gateway -n excise-tax-prod
          kubectl rollout status deployment/payment-service -n excise-tax-prod

      - name: Run smoke tests
        run: |
          ./scripts/smoke-test.sh https://tax.state.gov

      - name: Notify deployment
        if: always()
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          webhook_url: ${{ secrets.SLACK_WEBHOOK }}
          text: 'Production deployment ${{ job.status }}'
```

---

## Cloud Deployment

### Kubernetes Deployment Manifests

```yaml
# infrastructure/kubernetes/deployments/api-gateway.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
  namespace: excise-tax-portal
  labels:
    app: api-gateway
    version: v1
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-gateway
  template:
    metadata:
      labels:
        app: api-gateway
        version: v1
    spec:
      containers:
      - name: api-gateway
        image: ghcr.io/your-org/excise-tax-portal/api-gateway:latest
        ports:
        - containerPort: 8080
          name: http
        env:
        - name: APP_ENV
          value: production
        - name: APP_PORT
          value: "8080"
        - name: DB_HOST
          valueFrom:
            secretKeyRef:
              name: database-credentials
              key: host
        - name: DB_PORT
          valueFrom:
            secretKeyRef:
              name: database-credentials
              key: port
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: database-credentials
              key: username
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: database-credentials
              key: password
        - name: REDIS_HOST
          value: redis-service
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: jwt-secret
        resources:
          requests:
            memory: "512Mi"
            cpu: "500m"
          limits:
            memory: "1Gi"
            cpu: "1000m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: api-gateway-service
  namespace: excise-tax-portal
spec:
  type: LoadBalancer
  selector:
    app: api-gateway
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
---
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: api-gateway-hpa
  namespace: excise-tax-portal
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: api-gateway
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### Terraform AWS Deployment

```hcl
# infrastructure/terraform/modules/eks/main.tf

resource "aws_eks_cluster" "excise_tax" {
  name     = var.cluster_name
  role_arn = aws_iam_role.eks_cluster.arn
  version  = "1.28"

  vpc_config {
    subnet_ids              = var.subnet_ids
    endpoint_private_access = true
    endpoint_public_access  = true
    public_access_cidrs     = var.allowed_cidrs
    security_group_ids      = [aws_security_group.eks_cluster.id]
  }

  encryption_config {
    provider {
      key_arn = aws_kms_key.eks.arn
    }
    resources = ["secrets"]
  }

  enabled_cluster_log_types = ["api", "audit", "authenticator", "controllerManager", "scheduler"]

  depends_on = [
    aws_iam_role_policy_attachment.eks_cluster_policy,
    aws_iam_role_policy_attachment.eks_vpc_resource_controller,
  ]

  tags = var.tags
}

resource "aws_eks_node_group" "main" {
  cluster_name    = aws_eks_cluster.excise_tax.name
  node_group_name = "${var.cluster_name}-node-group"
  node_role_arn   = aws_iam_role.eks_node_group.arn
  subnet_ids      = var.subnet_ids

  scaling_config {
    desired_size = var.desired_nodes
    max_size     = var.max_nodes
    min_size     = var.min_nodes
  }

  instance_types = var.instance_types
  disk_size      = 50

  update_config {
    max_unavailable = 1
  }

  labels = {
    role = "general"
  }

  depends_on = [
    aws_iam_role_policy_attachment.eks_worker_node_policy,
    aws_iam_role_policy_attachment.eks_cni_policy,
    aws_iam_role_policy_attachment.eks_container_registry_policy,
  ]

  tags = var.tags
}

# RDS PostgreSQL
resource "aws_db_instance" "excise_tax" {
  identifier                = "${var.environment}-excise-tax-db"
  engine                    = "postgres"
  engine_version            = "15.4"
  instance_class            = var.db_instance_class
  allocated_storage         = 100
  max_allocated_storage     = 1000
  storage_encrypted         = true
  kms_key_id                = aws_kms_key.rds.arn
  
  db_name  = "excise_tax_portal"
  username = var.db_master_username
  password = var.db_master_password

  multi_az                    = true
  publicly_accessible         = false
  vpc_security_group_ids      = [aws_security_group.rds.id]
  db_subnet_group_name        = aws_db_subnet_group.main.name
  
  backup_retention_period     = 30
  backup_window               = "03:00-04:00"
  maintenance_window          = "Mon:04:00-Mon:05:00"
  
  enabled_cloudwatch_logs_exports = ["postgresql", "upgrade"]
  performance_insights_enabled    = true
  
  deletion_protection = true
  skip_final_snapshot = false
  final_snapshot_identifier = "${var.environment}-excise-tax-final-snapshot-${formatdate("YYYY-MM-DD-hhmm", timestamp())}"

  tags = var.tags
}

# ElastiCache Redis
resource "aws_elasticache_replication_group" "excise_tax" {
  replication_group_id       = "${var.environment}-excise-tax-redis"
  replication_group_description = "Redis cluster for Excise Tax Portal"
  engine                     = "redis"
  engine_version             = "7.0"
  node_type                  = var.redis_node_type
  number_cache_clusters      = 2
  port                       = 6379
  parameter_group_name       = "default.redis7"
  automatic_failover_enabled = true
  at_rest_encryption_enabled = true
  transit_encryption_enabled = true
  auth_token                 = var.redis_auth_token

  subnet_group_name  = aws_elasticache_subnet_group.main.name
  security_group_ids = [aws_security_group.redis.id]

  snapshot_retention_limit = 5
  snapshot_window         = "05:00-06:00"

  tags = var.tags
}
```

---

## Security Requirements

### 1. Authentication & Authorization
- JWT tokens with RS256 signing
- OAuth 2.0 integration with State.gov CDB
- Multi-factor authentication (MFA) for admin users
- Role-based access control (RBAC)
- Session expiration and refresh tokens

### 2. Data Encryption
- TLS 1.3 for all API communications
- Database encryption at rest (AES-256)
- Encrypted backups
- Secrets management via AWS Secrets Manager / HashiCorp Vault

### 3. XRPL Security
- State wallet private key stored in HSM or KMS
- Transaction signing in secure enclave
- Destination tags for payment tracking
- Transaction amount verification
- Multi-signature wallet support (optional)

### 4. Input Validation
- Server-side validation for all inputs
- SQL injection prevention (parameterized queries)
- XSS prevention (output encoding)
- CSRF protection
- Rate limiting per user/IP

### 5. Audit Logging
- All authentication attempts
- Payment transactions
- Report approvals
- Admin actions
- Failed authorization attempts
- Data exports

### 6. Compliance
- PCI DSS compliance for credit card payments
- GDPR compliance for EU users
- State data protection regulations
- Regular security audits
- Penetration testing (annual)

---

## Monitoring & Logging

### Prometheus Metrics

```go
// pkg/metrics/metrics.go

package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    // HTTP metrics
    HTTPRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    HTTPRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )

    // Payment metrics
    PaymentsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "payments_total",
            Help: "Total number of payments",
        },
        []string{"method", "status"},
    )

    XRPLPaymentDuration = promauto.NewHistogram(
        prometheus.HistogramOpts{
            Name:    "xrpl_payment_duration_seconds",
            Help:    "XRPL payment processing duration",
            Buckets: []float64{1, 3, 5, 10, 30, 60},
        },
    )

    XRPLTransactionFee = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "xrpl_transaction_fee_xrp",
            Help: "Current XRPL transaction fee in XRP",
        },
    )

    // Database metrics
    DBConnectionsActive = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "db_connections_active",
            Help: "Number of active database connections",
        },
    )

    DBQueryDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "db_query_duration_seconds",
            Help:    "Database query duration",
            Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 5},
        },
        []string{"query"},
    )

    // Business metrics
    ReportsSubmittedTotal = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "reports_submitted_total",
            Help: "Total number of reports submitted",
        },
    )

    ReportsApprovedTotal = promauto.NewCounter(
        prometheus.CounterOpts{
            Name: "reports_approved_total",
            Help: "Total number of reports approved",
        },
    )

    TaxCollectedUSD = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "tax_collected_usd",
            Help: "Total tax collected in USD",
        },
    )
)
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Excise Tax Portal - Overview",
    "panels": [
      {
        "title": "HTTP Requests per Second",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])"
          }
        ]
      },
      {
        "title": "XRPL Payment Success Rate",
        "targets": [
          {
            "expr": "sum(rate(payments_total{method=\"xrpl\",status=\"completed\"}[5m])) / sum(rate(payments_total{method=\"xrpl\"}[5m]))"
          }
        ]
      },
      {
        "title": "Average Payment Processing Time",
        "targets": [
          {
            "expr": "histogram_quantile(0.5, rate(xrpl_payment_duration_seconds_bucket[5m]))"
          }
        ]
      },
      {
        "title": "Tax Collected (24h)",
        "targets": [
          {
            "expr": "tax_collected_usd"
          }
        ]
      }
    ]
  }
}
```

---

## Development Workflow

### 1. Local Development Setup

```bash
# Clone repository
git clone https://github.com/your-org/excise-tax-portal.git
cd excise-tax-portal

# Start infrastructure
cd infrastructure/docker
docker-compose up -d postgres redis rabbitmq

# Setup backend
cd ../../backend
cp .env.example .env
# Edit .env with local config

# Run migrations
make migrate-up

# Install dependencies
go mod download

# Run tests
make test

# Start API gateway (development mode with hot reload)
make dev-api-gateway

# In another terminal, start payment service
make dev-payment-service
```

### 2. Making Changes

```bash
# Create feature branch
git checkout -b feature/add-payment-method

# Make changes
# ... edit code ...

# Run linter
make lint

# Run tests
make test

# Commit changes
git add .
git commit -m "feat: add credit card payment method"

# Push and create PR
git push origin feature/add-payment-method
```

### 3. Code Review Process

1. Developer creates PR
2. Automated tests run (GitHub Actions)
3. Code review by 2+ team members
4. Approval required before merge
5. Merge to `develop` branch
6. Auto-deploy to development environment
7. QA testing
8. Merge to `main` for production

### 4. Database Migrations

```bash
# Create new migration
make migrate-create name=add_payment_methods_table

# This creates:
# migrations/000003_add_payment_methods_table.up.sql
# migrations/000003_add_payment_methods_table.down.sql

# Run migration
make migrate-up

# Rollback migration
make migrate-down
```

### 5. Testing Strategy

**Unit Tests:**
```go
// internal/payment/service/xrpl_test.go

func TestCreatePaymentRequest(t *testing.T) {
    // Setup
    mockOracle := &mocks.MockPriceOracle{}
    mockOracle.On("GetExchangeRate", mock.Anything).Return(&ExchangeRate{
        XRPUSD: 0.7000,
    }, nil)
    
    service := NewPaymentService(mockOracle, nil, nil)
    
    // Execute
    req := CreatePaymentRequest{
        ManufacturerID: 123,
        AmountUSD:      1000.00,
        ReportID:       "RPT-001",
    }
    
    response, err := service.CreateXRPLPayment(context.Background(), req)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, response)
    assert.Equal(t, 1428.57, response.XRPAmount) // 1000 / 0.7
    assert.Equal(t, 0.7000, response.ExchangeRate)
}
```

**Integration Tests:**
```go
// tests/api/payment_test.go

func TestCreatePaymentEndToEnd(t *testing.T) {
    // Setup test server
    server := setupTestServer(t)
    defer server.Close()
    
    // Create test user
    token := createTestUser(t, server)
    
    // Create payment request
    payload := map[string]interface{}{
        "tax_amount_usd": 1000.00,
        "report_id":      "TEST-001",
    }
    
    resp := makeAuthRequest(t, server, "POST", "/api/v1/payments/xrpl/create", token, payload)
    
    // Assert
    assert.Equal(t, http.StatusCreated, resp.StatusCode)
    
    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)
    
    assert.True(t, result["success"].(bool))
    assert.NotEmpty(t, result["data"].(map[string]interface{})["payment_id"])
}
```

---

## Deployment Checklist

### Pre-Deployment

- [ ] All tests passing
- [ ] Code review completed
- [ ] Security scan completed
- [ ] Performance testing completed
- [ ] Database migrations tested
- [ ] Backup and rollback plan documented
- [ ] Monitoring dashboards configured
- [ ] Alerts configured
- [ ] Documentation updated
- [ ] Change request approved

### Deployment Steps

1. [ ] Announce maintenance window
2. [ ] Create database backup
3. [ ] Deploy backend services (canary first)
4. [ ] Run database migrations
5. [ ] Deploy frontend
6. [ ] Run smoke tests
7. [ ] Monitor metrics for 30 minutes
8. [ ] Gradual traffic ramp-up
9. [ ] Final verification
10. [ ] Announce deployment complete

### Post-Deployment

- [ ] Monitor error rates
- [ ] Check payment success rates
- [ ] Verify XRPL integration
- [ ] Review logs for issues
- [ ] Update status page
- [ ] Document any issues
- [ ] Schedule post-mortem (if needed)

---

## Appendix A: Environment Variables

### Backend Services

```bash
# Application
APP_ENV=production
APP_NAME=excise-tax-portal
APP_PORT=8080
APP_DEBUG=false

# Database
DB_HOST=postgres.example.com
DB_PORT=5432
DB_USER=excise_tax_app
DB_PASSWORD=<secret>
DB_NAME=excise_tax_portal
DB_SSL_MODE=require
DB_MAX_CONNECTIONS=100
DB_MAX_IDLE_CONNECTIONS=10

# Redis
REDIS_HOST=redis.example.com
REDIS_PORT=6379
REDIS_PASSWORD=<secret>
REDIS_DB=0
REDIS_MAX_RETRIES=3

# RabbitMQ
RABBITMQ_HOST=rabbitmq.example.com
RABBITMQ_PORT=5672
RABBITMQ_USER=excise_tax_app
RABBITMQ_PASSWORD=<secret>
RABBITMQ_VHOST=/excise-tax

# JWT
JWT_SECRET=<secret>
JWT_ACCESS_TOKEN_EXPIRY=1h
JWT_REFRESH_TOKEN_EXPIRY=168h

# OAuth
OAUTH_PROVIDER=state_gov_cdb
OAUTH_CLIENT_ID=<client-id>
OAUTH_CLIENT_SECRET=<secret>
OAUTH_REDIRECT_URI=https://tax.state.gov/auth/callback
OAUTH_AUTHORIZATION_URL=https://login.state.gov/oauth2/authorize
OAUTH_TOKEN_URL=https://login.state.gov/oauth2/token

# XRPL
XRPL_NETWORK=mainnet
XRPL_WS_ENDPOINT=wss://xrplcluster.com
XRPL_STATE_ADDRESS=rN7n7otQDd6FczFgLdCqZLYjmHG3A9LKr7
XRPL_STATE_SECRET=<secret>
XRPL_PAYMENT_TIMEOUT=24h
XRPL_PRICE_ORACLE_INTERVAL=30s

# Email
SMTP_HOST=smtp.example.com
SMTP_PORT=587
SMTP_USER=notifications@tax.state.gov
SMTP_PASSWORD=<secret>
SMTP_FROM=Excise Tax Portal <noreply@tax.state.gov>

# AWS
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=<key-id>
AWS_SECRET_ACCESS_KEY=<secret>
S3_BUCKET=excise-tax-documents

# Monitoring
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
JAEGER_ENABLED=true
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
```

---

## Appendix B: API Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `AUTH_001` | 401 | Invalid credentials |
| `AUTH_002` | 401 | Token expired |
| `AUTH_003` | 401 | Invalid token |
| `AUTH_004` | 403 | Insufficient permissions |
| `PAY_001` | 400 | Invalid payment amount |
| `PAY_002` | 400 | Invalid payment method |
| `PAY_003` | 404 | Payment not found |
| `PAY_004` | 409 | Payment already processed |
| `PAY_005` | 500 | XRPL connection error |
| `PAY_006` | 408 | Payment timeout |
| `TAX_001` | 400 | Invalid report data |
| `TAX_002` | 404 | Report not found |
| `TAX_003` | 409 | Report already submitted |
| `TAX_004` | 400 | Invalid tax calculation |
| `SYS_001` | 500 | Internal server error |
| `SYS_002` | 503 | Service unavailable |
| `SYS_003` | 429 | Rate limit exceeded |

---

## Appendix C: Makefile

```makefile
# backend/Makefile

.PHONY: help build test lint migrate-up migrate-down dev clean

# Variables
APP_NAME := excise-tax-portal
DOCKER_REGISTRY := ghcr.io/your-org
GO_VERSION := 1.21

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

dev-api-gateway: ## Run API gateway in development mode
	@echo "Starting API Gateway..."
	@air -c .air.api-gateway.toml

dev-payment-service: ## Run payment service in development mode
	@echo "Starting Payment Service..."
	@air -c .air.payment-service.toml

test: ## Run tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

lint: ## Run linter
	@echo "Running linter..."
	@golangci-lint run --timeout=5m

##@ Database

migrate-up: ## Run database migrations
	@echo "Running migrations..."
	@migrate -path=./migrations -database="$(DB_URL)" up

migrate-down: ## Rollback database migrations
	@echo "Rolling back migrations..."
	@migrate -path=./migrations -database="$(DB_URL)" down 1

migrate-create: ## Create new migration (usage: make migrate-create name=add_users_table)
	@echo "Creating migration: $(name)"
	@migrate create -ext sql -dir ./migrations -seq $(name)

##@ Build

build: ## Build all services
	@echo "Building services..."
	@CGO_ENABLED=0 go build -o bin/api-gateway ./cmd/api-gateway
	@CGO_ENABLED=0 go build -o bin/payment-service ./cmd/payment-service
	@CGO_ENABLED=0 go build -o bin/tax-service ./cmd/tax-service
	@CGO_ENABLED=0 go build -o bin/reporting-service ./cmd/reporting-service

build-docker: ## Build Docker images
	@echo "Building Docker images..."
	@docker build -t $(DOCKER_REGISTRY)/api-gateway:latest -f ../infrastructure/docker/Dockerfile.backend --build-arg SERVICE=api-gateway ..
	@docker build -t $(DOCKER_REGISTRY)/payment-service:latest -f ../infrastructure/docker/Dockerfile.backend --build-arg SERVICE=payment-service ..

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

##@ Tools

install-tools: ## Install development tools
	@echo "Installing tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/cosmtrek/air@latest
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

---

## Contact & Support

**Project Lead:** Justin Harvey  
**Email:** jharvmail@gmail.com  
**Repository:** https://github.com/your-org/excise-tax-portal

**Documentation Updates:** This document should be updated as the project evolves. All changes should be reviewed by the technical lead.

**Version History:**
- v2.0.0 (2025-12-29) - Production architecture specification
- v1.0.0 (2025-12-17) - Initial static demo

---

*End of Specification Document*
