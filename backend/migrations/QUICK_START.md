# Database Migrations Quick Start

## Prerequisites

1. **Install golang-migrate**
   - Windows: `choco install migrate` or `scoop install migrate`
   - macOS: `brew install golang-migrate`
   - Linux: See README.md for download instructions

2. **PostgreSQL Database**
   - Install PostgreSQL 14+ (recommended: 16)
   - Create database: `createdb excise_tax_portal`

## Quick Setup

### 1. Configure Database Connection

**Option A: Environment Variables (Recommended)**
```bash
# Windows (PowerShell)
$env:DB_HOST="localhost"
$env:DB_PORT="5432"
$env:DB_USER="postgres"
$env:DB_PASSWORD="your_password"
$env:DB_NAME="excise_tax_portal"

# Linux/macOS (Bash)
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your_password
export DB_NAME=excise_tax_portal
```

**Option B: Complete Connection String**
```bash
# Windows (PowerShell)
$env:DATABASE_URL="postgresql://postgres:password@localhost:5432/excise_tax_portal?sslmode=disable"

# Linux/macOS (Bash)
export DATABASE_URL="postgresql://postgres:password@localhost:5432/excise_tax_portal?sslmode=disable"
```

### 2. Run Migrations

**Windows (PowerShell)**
```powershell
cd C:\Users\justin.harvey\excise-tax-portal\backend

# Apply all migrations
.\scripts\migrate.ps1 up

# Check current version
.\scripts\migrate.ps1 version
```

**Linux/macOS (Bash)**
```bash
cd ~/excise-tax-portal/backend

# Make script executable
chmod +x scripts/migrate.sh

# Apply all migrations
./scripts/migrate.sh up

# Check current version
./scripts/migrate.sh version
```

## Common Commands

### Apply Migrations
```bash
# Apply all pending migrations
./scripts/migrate.sh up              # Linux/macOS
.\scripts\migrate.ps1 up             # Windows

# Apply specific number of migrations
./scripts/migrate.sh up 2            # Linux/macOS
.\scripts\migrate.ps1 up 2           # Windows
```

### Rollback Migrations
```bash
# Rollback last migration
./scripts/migrate.sh down 1          # Linux/macOS
.\scripts\migrate.ps1 down 1         # Windows

# Rollback all migrations (requires confirmation)
./scripts/migrate.sh down            # Linux/macOS
.\scripts\migrate.ps1 down           # Windows
```

### Check Status
```bash
# Show current migration version
./scripts/migrate.sh version         # Linux/macOS
.\scripts\migrate.ps1 version        # Windows
```

### Create New Migration
```bash
# Create new migration files
./scripts/migrate.sh create add_feature       # Linux/macOS
.\scripts\migrate.ps1 create add_feature      # Windows
```

## Migration Sequence

The migrations will be applied in this order:

1. **000001_init_schema** - Creates core tables
   - users
   - manufacturers
   - audit_log
   - sessions

2. **000002_add_xrpl_tables** - Adds XRPL payment infrastructure
   - payments
   - xrpl_payments
   - exchange_rates
   - xrpl_transaction_log
   - xrpl_settlement_batches

3. **000003_add_tax_tables** - Adds tax reporting
   - tax_reports
   - tax_rates

4. **000004_add_views** - Creates database views
   - dashboard_metrics
   - payment_summary
   - xrpl_payment_stats
   - manufacturer_report_summary
   - recent_activity
   - tax_rate_history
   - daily_settlement_summary

## Verify Installation

After running migrations, verify the setup:

```sql
-- Connect to database
psql -U postgres -d excise_tax_portal

-- List all tables
\dt

-- List all views
\dv

-- Check migration version
SELECT * FROM schema_migrations;

-- Count tables (should be 11)
SELECT COUNT(*) FROM information_schema.tables
WHERE table_schema = 'public' AND table_type = 'BASE TABLE';

-- Count views (should be 7)
SELECT COUNT(*) FROM information_schema.views
WHERE table_schema = 'public';
```

Expected output:
- 11 tables (users, manufacturers, audit_log, sessions, payments, xrpl_payments, exchange_rates, xrpl_transaction_log, xrpl_settlement_batches, tax_reports, tax_rates)
- 7 views
- schema_migrations table showing version 4

## Troubleshooting

### Error: "Dirty database version"
```bash
# Check current version
./scripts/migrate.sh version

# Output shows: 3 (dirty)
# This means migration 3 failed mid-execution

# Fix by forcing to last good version
./scripts/migrate.sh force 2

# Then try again
./scripts/migrate.sh up
```

### Error: "migrate: command not found"
```bash
# Verify installation
which migrate                    # Linux/macOS
where.exe migrate                # Windows

# If not found, install golang-migrate (see Prerequisites)
```

### Error: "connection refused"
```bash
# Verify PostgreSQL is running
psql -U postgres -c "SELECT version();"

# Check connection parameters
echo $DATABASE_URL               # Linux/macOS
echo $env:DATABASE_URL          # Windows
```

### Error: "permission denied"
```bash
# Make script executable (Linux/macOS only)
chmod +x scripts/migrate.sh

# Or run with bash directly
bash scripts/migrate.sh up
```

## Development Workflow

### 1. Starting Fresh
```bash
# Create database
createdb excise_tax_portal

# Run all migrations
./scripts/migrate.sh up

# Verify
./scripts/migrate.sh version
```

### 2. Making Changes
```bash
# Create new migration
./scripts/migrate.sh create add_notifications

# Edit the generated files:
# - migrations/000005_add_notifications.up.sql
# - migrations/000005_add_notifications.down.sql

# Test migration
./scripts/migrate.sh up 1

# Test rollback
./scripts/migrate.sh down 1

# Apply again if successful
./scripts/migrate.sh up 1
```

### 3. Resetting Database
```bash
# Option 1: Rollback all migrations
./scripts/migrate.sh down

# Option 2: Drop everything
./scripts/migrate.sh drop

# Then reapply
./scripts/migrate.sh up
```

## Production Deployment

### Pre-deployment
```bash
# 1. Backup database
pg_dump -h production-host -U postgres excise_tax_portal > backup.sql

# 2. Test on staging
DATABASE_URL="postgresql://user:pass@staging-host:5432/excise_tax_portal" \
  ./scripts/migrate.sh up

# 3. Verify staging
DATABASE_URL="postgresql://user:pass@staging-host:5432/excise_tax_portal" \
  ./scripts/migrate.sh version
```

### Deployment
```bash
# Apply migrations to production
DATABASE_URL="postgresql://user:pass@production-host:5432/excise_tax_portal" \
  ./scripts/migrate.sh up

# Verify
DATABASE_URL="postgresql://user:pass@production-host:5432/excise_tax_portal" \
  ./scripts/migrate.sh version
```

## Next Steps

1. Review the full [README.md](./README.md) for detailed documentation
2. Check [BACKEND_INFRASTRUCTURE_SPEC.md](../BACKEND_INFRASTRUCTURE_SPEC.md) for schema details
3. Set up your Go application to use these database tables
4. Configure connection pooling and environment variables
5. Implement database models in your application

## Support

- Detailed documentation: [README.md](./README.md)
- golang-migrate docs: https://github.com/golang-migrate/migrate
- PostgreSQL docs: https://www.postgresql.org/docs/
