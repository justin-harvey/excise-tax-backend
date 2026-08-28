# Excise Tax Payment Platform

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.21%2B-blue)](https://golang.org/dl/)
[![Docker](https://img.shields.io/badge/Docker-24.0%2B-blue)](https://www.docker.com/)
[![XRPL](https://img.shields.io/badge/XRPL-Blockchain-green)](https://xrpl.org/)
[![Security](https://img.shields.io/badge/Security-Audited-green)](SECURITY_FIXES_APPLIED.md)

> **Blockchain-enabled excise tax collection platform designed to dramatically reduce processing costs and enable near-instant settlement for government agencies.**

---

## Overview

The Excise Tax Payment Platform is a complete, full-stack system for government excise tax collection, featuring:

* **Frontend Portal** – Production web application for manufacturers, distributors, and government administrators

  * Live at: [https://github.com/G00DTECH/Excise-Tax-Payment-Platform](https://github.com/G00DTECH/Excise-Tax-Payment-Platform)
  * Company registration and secure authentication
  * Payment processing (ACH, Credit Card, Wire Transfer)
  * Manufacturing report submissions (13 LCC report types)
  * Administrative review and approval workflows
  * Payment history and tax calculators

* **Backend Infrastructure (this repository)** – A prototype cloud-native microservices system integrating XRP Ledger (XRPL) settlement to materially reduce payment processing costs while providing 3–5 second settlement times versus traditional monthly or multi-week batch processes

---

## Alignment with the Federal Digital Payments Executive Order

This platform is designed to directly support and operationalize recent federal executive guidance on digital payment modernization, including mandates to:

* Expand acceptance of digital wallets and modern payment methods
* Reduce reliance on paper checks and legacy batch ACH processes
* Improve settlement speed, transparency, and reconciliation
* Increase resilience, auditability, and security of government payment systems

### Patent Utility in This Context

The underlying patent associated with this platform covers a set of optimizations for government payment processing that enable:

* Real-time or near-real-time settlement without changing existing taxpayer-facing workflows
* Cost-competitive processing by aggregating transaction volume, optimizing internal payment flows, and accessing institutional pricing
* Compatibility with digital assets, tokenized dollars, and blockchain settlement layers while preserving compliance with existing treasury and accounting requirements
* Incremental modernization, allowing agencies to meet executive order requirements without full system replacement

In practical terms, the patent allows government agencies and payment partners to layer modern digital settlement rails (including blockchain and regulated digital dollars) beneath existing portals and accounting systems, accelerating compliance with federal digital payments policy while minimizing operational disruption.

---

## Key Features

* XRPL-based settlement for low-cost, near-instant payments
* Government-grade security including JWT authentication, role-based access control, and full audit trails
* High-throughput microservices architecture capable of handling 1,000+ transactions per second
* Real-time monitoring via Prometheus and Grafana
* Multi-source XRP/USD price oracle with aggregated exchange data
* Mobile-friendly payment flows including QR-based wallet payments
* White-label architecture suitable for reuse across states and jurisdictions
* Enterprise-scale deployment with Kubernetes readiness

---

## Business Impact

| Metric             | Traditional Systems                   | Platform Architecture | Improvement                |
| ------------------ | ------------------------------------- | --------------------- | -------------------------- |
| Processing Cost    | $3–4B annually                        | ~$4M annually         | ~99.9% reduction           |
| Settlement Time    | 30+ days                              | 3–5 seconds           | Orders of magnitude faster |
| Transaction Fee    | $15–$280                              | ~$0.0003              | Materially lower           |
| Addressable Market | $250–300B annual US excise tax volume | —                     | —                          |

Projected annual savings scale with transaction volume and jurisdiction adoption.

---

## System Components

This repository contains the backend infrastructure that powers the Excise Tax Payment Platform.

### Frontend (Separate Repository)

* Repository: [https://github.com/G00DTECH/Excise-Tax-Payment-Platform](https://github.com/G00DTECH/Excise-Tax-Payment-Platform)
* Technology: HTML5, JavaScript, responsive web design
* Features include taxpayer portals, administrative review tools, payment interfaces, report submission, calculators, and status tracking

### Backend (This Repository)

* Technology: Go microservices, PostgreSQL, Redis, RabbitMQ, XRPL
* Features:

  * RESTful API gateway for frontend integration
  * Blockchain-backed payment settlement
  * Authentication and authorization services
  * Tax calculation and validation engines
  * Reporting, analytics, and notification services

### High-Level Flow

Frontend clients communicate with the backend via JSON-based REST APIs over HTTPS. The API gateway routes requests to dedicated payment, tax, and reporting services, which interact with databases, message queues, monitoring tools, and the XRPL settlement layer.

---

## Backend Architecture

The backend follows a microservices architecture with an API gateway, dedicated domain services, and shared infrastructure services. It is containerized using Docker and designed for horizontal scaling via Kubernetes.

### Technology Stack

**Backend**

* Go 1.21+
* Gin (HTTP framework)
* GORM (ORM)
* PostgreSQL 15
* Redis 7
* RabbitMQ 3.12
* XRP Ledger (XRPL)

**Infrastructure**

* Docker and Docker Compose
* Kubernetes (deployment-ready)
* Prometheus and Grafana for monitoring
* GitHub Actions for CI/CD
* Multi-cloud compatibility (AWS, GCP, Azure)

---

## Quick Start

### Prerequisites

* Docker Desktop 24.0+
* Go 1.21+
* 8 GB RAM minimum (16 GB recommended)

### Security Setup (Required)

Before starting services, configure strong passwords for all infrastructure components.

1. Copy the environment template:

   ```bash
   cd infrastructure/docker
   cp .env.docker.example .env.docker
   ```

2. Generate four strong passwords and update the following values:

   * POSTGRES_PASSWORD
   * REDIS_PASSWORD
   * RABBITMQ_PASSWORD
   * GRAFANA_ADMIN_PASSWORD

Do not commit `.env.docker` to version control.

### Installation

```bash
# Clone repository
git clone https://github.com/YOUR_USERNAME/excise-tax-portal.git
cd excise-tax-portal

# Start infrastructure
docker compose up -d

# Run migrations and services
make build
make run-all
```

---

## Documentation

* Security fixes and audit notes: SECURITY_FIXES_APPLIED.md
* Startup and deployment guides
* Backend architecture and file structure documentation
* REST API specifications

Frontend documentation is maintained in the separate frontend repository.

---

## License

This project is licensed under the MIT License. See the LICENSE file for details.

---

Built to support modern, secure, and policy-aligned government payment infrastructure.
