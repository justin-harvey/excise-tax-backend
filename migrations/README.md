# Database Migrations

This directory contains PostgreSQL database migrations for the Excise Tax Portal using [golang-migrate](https://github.com/golang-migrate/migrate).

## Directory Structure

```
migrations/
├── 000001_init_schema.up.sql           # Core tables (users, manufacturers, audit_log, sessions)
├── 000001_init_schema.down.sql         # Rollback for core tables
├── 000002_add_xrpl_tables.up.sql       # XRPL payment tables
├── 000002_add_xrpl_tables.down.sql     # Rollback for XRPL tables
├── 000003_add_tax_tables.up.sql        # Tax reporting tables
├── 000003_add_tax_tables.down.sql      # Rollback for tax tables
├── 000004_add_views.up.sql             # Database views for reporting
├── 000004_add_views.down.sql           # Rollback for views
└── README.md                            # This file
```

## Migration Overview

### Migration 000001: Init Schema
**Purpose:** Creates core authentication and user management tables

**Tables Created:**
- `users` - User accounts with roles (manufacturer, admin, super_admin)
- `manufacturers` - Manufacturer business registration details
- `audit_log` - Complete audit trail of system actions
- `sessions` - Session storage for authentication

**Key Features:**
- Foreign key relationships with appropriate CASCADE/SET NULL
- Indexed columns for performance
- Triggers for automatic `updated_at` timestamps
- Email uniqueness constraint
- Active user filtering support

### Migration 000002: XRPL Tables
**Purpose:** Adds XRP Ledger payment processing infrastructure

**Tables Created:**
- `payments` - Generic payment records (all payment methods)
- `xrpl_payments` - XRPL-specific payment details
- `exchange_rates` - Historical XRP/USD exchange rates
- `xrpl_transaction_log` - Complete XRPL transaction audit trail
- `xrpl_settlement_batches` - Daily XRP to USD settlement tracking

**Key Features:**
- Support for multiple payment methods (XRPL, ACH, credit card, wire)
- Exchange rate tracking from multiple sources
- Destination tag-based payment matching
- QR code storage for payment requests
- Settlement batch processing

### Migration 000003: Tax Tables
**Purpose:** Adds tax reporting and rate management

**Tables Created:**
- `tax_reports` - Tax reports submitted by manufacturers
- `tax_rates` - Product-specific tax rates with effective dates

**Key Features:**
- Multiple report types (monthly, quarterly, annual)
- Production data stored as JSONB
- Tax calculation support
- Report review workflow (draft, pending, under_review, approved, etc.)
- Historical tax rate tracking

### Migration 000004: Views
**Purpose:** Creates database views for reporting and analytics

**Views Created:**
- `dashboard_metrics` - Aggregated metrics for admin dashboard
- `payment_summary` - Payment statistics by manufacturer
- `xrpl_payment_stats` - Daily XRPL payment statistics
- `manufacturer_report_summary` - Report statistics by manufacturer
- `recent_activity` - Recent system activity (last 100 entries)
- `tax_rate_history` - Current and historical tax rates
- `daily_settlement_summary` - XRPL settlement batch metrics

## Installation

### Install golang-migrate

**macOS (Homebrew):**
```bash
brew install golang-migrate
```

**Linux (curl):**
```bash
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/
```

**Windows (Chocolatey):**
```bash
choco install migrate
```

**Go install:**
```bash
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Usage

### Using the Migration Script

The `scripts/migrate.sh` script provides a convenient wrapper around golang-migrate:

```bash
# Apply all pending migrations
./scripts/migrate.sh up

# Apply next N migrations
./scripts/migrate.sh up 2

# Rollback last migration
./scripts/migrate.sh down 1

# Rollback all migrations (with confirmation)
./scripts/migrate.sh down

# Check current version
./scripts/migrate.sh version

# Force version (fix dirty state)
./scripts/migrate.sh force 3

# Create new migration
./scripts/migrate.sh create add_new_feature

# Drop all tables (with confirmation)
./scripts/migrate.sh drop
```

### Using golang-migrate Directly

```bash
# Set database URL
export DATABASE_URL="postgresql://user:password@localhost:5432/excise_tax_portal?sslmode=disable"

# Apply all migrations
migrate -path ./migrations -database $DATABASE_URL up

# Rollback last migration
migrate -path ./migrations -database $DATABASE_URL down 1

# Check version
migrate -path ./migrations -database $DATABASE_URL version
```

## Environment Variables

Configure database connection using environment variables:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=your_password
export DB_NAME=excise_tax_portal

# Or use a complete connection string
export DATABASE_URL="postgresql://user:password@localhost:5432/excise_tax_portal?sslmode=disable"
```

## Creating New Migrations

### Using the Script
```bash
./scripts/migrate.sh create add_notification_tables
```

This creates:
- `000005_add_notification_tables.up.sql`
- `000005_add_notification_tables.down.sql`

### Manual Creation

1. Find the next migration number (e.g., `000005`)
2. Create two files:
   - `000005_your_migration_name.up.sql` - Forward migration
   - `000005_your_migration_name.down.sql` - Rollback migration
3. Write SQL for schema changes in the `.up.sql` file
4. Write rollback SQL in the `.down.sql` file

### Migration Best Practices

**DO:**
- Always create matching up and down migrations
- Test migrations on a development database first
- Use transactions when possible
- Add comments explaining complex changes
- Include indexes for foreign keys and frequently queried columns
- Use CHECK constraints for data validation
- Add table and column comments using `COMMENT ON`

**DON'T:**
- Modify existing migration files after they've been applied
- Create migrations that can't be rolled back
- Drop tables without CASCADE or checking dependencies
- Forget to handle NULL values appropriately
- Use application-specific logic in migrations

## Schema Documentation

### Core Concepts

1. **Users & Authentication**
   - Users have roles: manufacturer, admin, super_admin
   - Manufacturers link to users via `user_id`
   - Sessions table provides PostgreSQL backup for Redis

2. **Payment Processing**
   - Generic `payments` table for all payment types
   - XRPL-specific details in `xrpl_payments`
   - Exchange rates tracked from multiple sources
   - Settlement batches for daily XRP to USD conversion

3. **Tax Reporting**
   - Reports have status workflow
   - Production data stored as JSONB for flexibility
   - Tax rates with effective date ranges
   - Automatic tax calculation support

4. **Auditing**
   - All actions logged to `audit_log`
   - IP address and user agent tracking
   - JSON metadata for additional context

### Data Types Used

- `BIGSERIAL` - Auto-incrementing 64-bit integers for IDs
- `TIMESTAMP` - Date and time (not DATETIME)
- `DECIMAL(p, s)` - Precise numeric values for money
- `JSONB` - Binary JSON for flexible data structures
- `INET` - IP addresses
- `VARCHAR(n)` - Variable-length strings with limit

### Foreign Key Policies

- `ON DELETE CASCADE` - Child records deleted with parent (sessions, manufacturers)
- `ON DELETE RESTRICT` - Prevents deletion if children exist (payments, reports)
- `ON DELETE SET NULL` - Sets to NULL when parent deleted (audit_log, reviewers)

## Troubleshooting

### Dirty Database State

If a migration fails mid-execution, the database may be in a "dirty" state:

```bash
# Check current version (will show "dirty" status)
./scripts/migrate.sh version

# Fix by forcing to last known good version
./scripts/migrate.sh force 2
```

### Connection Issues

Verify database connection:
```bash
psql -h localhost -U postgres -d excise_tax_portal -c "SELECT version();"
```

### Permission Errors

Ensure database user has proper permissions:
```sql
GRANT ALL PRIVILEGES ON DATABASE excise_tax_portal TO your_user;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO your_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO your_user;
```

## Testing Migrations

### Development Environment

```bash
# Create test database
createdb excise_tax_portal_test

# Run migrations
DATABASE_URL="postgresql://postgres:postgres@localhost:5432/excise_tax_portal_test?sslmode=disable" \
  ./scripts/migrate.sh up

# Test rollback
DATABASE_URL="postgresql://postgres:postgres@localhost:5432/excise_tax_portal_test?sslmode=disable" \
  ./scripts/migrate.sh down

# Drop test database
dropdb excise_tax_portal_test
```

### Automated Testing

```bash
#!/bin/bash
# test_migrations.sh

set -e

DB_NAME="excise_tax_portal_test_$(date +%s)"
export DATABASE_URL="postgresql://postgres:postgres@localhost:5432/$DB_NAME?sslmode=disable"

echo "Creating test database: $DB_NAME"
createdb "$DB_NAME"

echo "Running migrations up..."
./scripts/migrate.sh up

echo "Running migrations down..."
./scripts/migrate.sh down

echo "Cleaning up..."
dropdb "$DB_NAME"

echo "Migration tests passed!"
```

## Production Deployment

### Pre-deployment Checklist

- [ ] Test migrations on staging environment
- [ ] Backup production database
- [ ] Review all SQL for potential issues
- [ ] Verify rollback procedures
- [ ] Check for long-running migrations (locks)
- [ ] Plan for downtime if needed

### Deployment Steps

1. **Backup Database**
   ```bash
   pg_dump -h production-host -U postgres excise_tax_portal > backup_$(date +%Y%m%d_%H%M%S).sql
   ```

2. **Apply Migrations**
   ```bash
   ./scripts/migrate.sh up
   ```

3. **Verify Success**
   ```bash
   ./scripts/migrate.sh version
   psql -h production-host -U postgres -d excise_tax_portal -c "SELECT COUNT(*) FROM users;"
   ```

4. **Monitor Application**
   - Check application logs
   - Verify API endpoints
   - Test critical user flows

## Additional Resources

- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Backend Infrastructure Spec](../BACKEND_INFRASTRUCTURE_SPEC.md)

## Support

For issues or questions:
1. Check migration error messages carefully
2. Review PostgreSQL logs
3. Verify database connection settings
4. Test migrations on development environment first
