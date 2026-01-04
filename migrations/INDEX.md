# Database Migrations - File Index

## Quick Navigation

- [Migration Files](#migration-files) - SQL migration files
- [Documentation](#documentation) - Guides and references
- [Scripts](#scripts) - Automation and helper scripts
- [Usage Guide](#usage-guide) - Quick start instructions

## Migration Files

### Migration 1: Core Schema (Init)
| File | Path | Purpose | Lines |
|------|------|---------|-------|
| **000001_init_schema.up.sql** | C:\Users\justin.harvey\excise-tax-portal\backend\migrations\000001_init_schema.up.sql | Creates core authentication and user management tables | 157 |
| **000001_init_schema.down.sql** | C:\Users\justin.harvey\excise-tax-portal\backend\migrations\000001_init_schema.down.sql | Rollback for init schema | 16 |

**Tables Created:**
- users (11 columns, 4 indexes)
- manufacturers (16 columns, 6 indexes)
- audit_log (9 columns, 6 indexes)
- sessions (5 columns, 2 indexes)

**Also Creates:**
- update_updated_at_column() function
- 2 triggers for automatic timestamps

---

### Migration 2: XRPL Payment Infrastructure
| File | Path | Purpose | Lines |
|------|------|---------|-------|
| **000002_add_xrpl_tables.up.sql** | C:\Users\justin.harvey\excise-tax-portal\backend\migrations\000002_add_xrpl_tables.up.sql | Creates XRPL payment processing infrastructure | 227 |
| **000002_add_xrpl_tables.down.sql** | C:\Users\justin.harvey\excise-tax-portal\backend\migrations\000002_add_xrpl_tables.down.sql | Rollback for XRPL tables | 15 |

**Tables Created:**
- payments (13 columns, 7 indexes)
- xrpl_payments (16 columns, 9 indexes)
- exchange_rates (8 columns, 4 indexes)
- xrpl_transaction_log (13 columns, 6 indexes)
- xrpl_settlement_batches (15 columns, 3 indexes)

**Also Creates:**
- get_pending_xrpl_payment_by_tag() function
- 2 triggers for automatic timestamps

---

### Migration 3: Tax Reporting
| File | Path | Purpose | Lines |
|------|------|---------|-------|
| **000003_add_tax_tables.up.sql** | C:\Users\justin.harvey\excise-tax-portal\backend\migrations\000003_add_tax_tables.up.sql | Creates tax reporting and rate management tables | 180 |
| **000003_add_tax_tables.down.sql** | C:\Users\justin.harvey\excise-tax-portal\backend\migrations\000003_add_tax_tables.down.sql | Rollback for tax tables | 15 |

**Tables Created:**
- tax_reports (16 columns, 9 indexes)
- tax_rates (8 columns, 5 indexes)

**Also Creates:**
- get_current_tax_rate() function
- calculate_report_tax() function
- 1 trigger for automatic timestamps
- Default tax rate seed data

---

### Migration 4: Database Views
| File | Path | Purpose | Lines |
|------|------|---------|-------|
| **000004_add_views.up.sql** | C:\Users\justin.harvey\excise-tax-portal\backend\migrations\000004_add_views.up.sql | Creates analytical views for reporting | 157 |
| **000004_add_views.down.sql** | C:\Users\justin.harvey\excise-tax-portal\backend\migrations\000004_add_views.down.sql | Rollback for views | 12 |

**Views Created:**
- dashboard_metrics (admin dashboard summary)
- payment_summary (manufacturer payment statistics)
- xrpl_payment_stats (daily XRPL statistics)
- manufacturer_report_summary (report statistics by manufacturer)
- recent_activity (last 100 audit log entries)
- tax_rate_history (current and historical rates)
- daily_settlement_summary (settlement batch metrics)

---

## Documentation

| File | Path | Purpose | Size |
|------|------|---------|------|
| **README.md** | C:\Users\justin.harvey\excise-tax-portal\backend\migrations\README.md | Comprehensive migration documentation | ~450 lines |
| **QUICK_START.md** | C:\Users\justin.harvey\excise-tax-portal\backend\migrations\QUICK_START.md | Quick reference guide | ~285 lines |
| **MIGRATION_SUMMARY.md** | C:\Users\justin.harvey\excise-tax-portal\backend\migrations\MIGRATION_SUMMARY.md | Complete summary of all migrations | ~350 lines |
| **INDEX.md** | C:\Users\justin.harvey\excise-tax-portal\backend\migrations\INDEX.md | This file - complete file index | Current |

### Documentation Contents

**README.md** includes:
- Installation instructions
- Usage examples
- Migration best practices
- Troubleshooting guide
- Production deployment checklist
- Schema documentation

**QUICK_START.md** includes:
- Prerequisites
- Quick setup steps
- Common commands
- Development workflow
- Verification steps

**MIGRATION_SUMMARY.md** includes:
- Complete file list
- Database schema overview
- Key features
- Testing instructions
- Technical specifications

---

## Scripts

### Migration Scripts

| File | Path | Platform | Purpose |
|------|------|----------|---------|
| **migrate.sh** | C:\Users\justin.harvey\excise-tax-portal\backend\scripts\migrate.sh | Linux/macOS | Bash migration wrapper |
| **migrate.ps1** | C:\Users\justin.harvey\excise-tax-portal\backend\scripts\migrate.ps1 | Windows | PowerShell migration wrapper |
| **verify.sh** | C:\Users\justin.harvey\excise-tax-portal\backend\scripts\verify.sh | Linux/macOS | Run verification script |
| **verify_migrations.sql** | C:\Users\justin.harvey\excise-tax-portal\backend\scripts\verify_migrations.sql | All | SQL verification script |

### Script Commands

**migrate.sh / migrate.ps1** support:
- `up [N]` - Apply migrations
- `down [N]` - Rollback migrations
- `drop` - Drop all tables
- `force VERSION` - Fix dirty state
- `version` - Show current version
- `create NAME` - Create new migration

**verify.sh** runs comprehensive verification including:
- Migration version check
- Object count validation
- Table structure verification
- Index validation
- Foreign key check
- Expected vs actual comparison

---

### Makefile

| File | Path | Purpose |
|------|------|---------|
| **Makefile** | C:\Users\justin.harvey\excise-tax-portal\backend\Makefile | Make targets for automation |

**Available Make Targets:**
```
Database Migrations:
  migrate-up          Apply all pending migrations
  migrate-up-1        Apply next migration
  migrate-down        Rollback last migration
  migrate-down-all    Rollback all migrations
  migrate-reset       Reset database (down all, then up all)
  migrate-version     Show current migration version
  migrate-create      Create new migration
  migrate-force       Force migration version

Database Setup:
  db-create          Create database
  db-drop            Drop database
  db-setup           Create database and run migrations
  db-reset           Drop, recreate, and migrate
  db-test            Set up test database
  db-verify          Verify database schema
  db-console         Open PostgreSQL console
  db-backup          Backup database
  db-restore         Restore from backup

Development:
  build              Build Go application
  run                Run Go application
  test               Run tests
  test-coverage      Run tests with coverage
  lint               Run linter
  fmt                Format Go code
  tidy               Tidy Go modules
  dev                Setup development environment
  clean              Clean build artifacts
  install-tools      Install development tools
```

---

## Usage Guide

### Initial Setup

1. **Install golang-migrate**
   ```bash
   # macOS
   brew install golang-migrate

   # Windows
   choco install migrate

   # Linux
   curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
   sudo mv migrate /usr/local/bin/
   ```

2. **Set Environment Variables**
   ```bash
   # Linux/macOS
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=postgres
   export DB_PASSWORD=your_password
   export DB_NAME=excise_tax_portal

   # Or use complete URL
   export DATABASE_URL="postgresql://postgres:password@localhost:5432/excise_tax_portal?sslmode=disable"
   ```

3. **Run Migrations**
   ```bash
   # Using script
   cd C:\Users\justin.harvey\excise-tax-portal\backend
   ./scripts/migrate.sh up

   # Using Make
   make migrate-up
   ```

4. **Verify Installation**
   ```bash
   # Using script
   ./scripts/verify.sh

   # Or manually
   psql -U postgres -d excise_tax_portal -f scripts/verify_migrations.sql
   ```

### Common Operations

**Apply all migrations:**
```bash
./scripts/migrate.sh up          # Bash
.\scripts\migrate.ps1 up         # PowerShell
make migrate-up                  # Make
```

**Rollback last migration:**
```bash
./scripts/migrate.sh down 1      # Bash
.\scripts\migrate.ps1 down 1     # PowerShell
make migrate-down                # Make
```

**Check version:**
```bash
./scripts/migrate.sh version     # Bash
.\scripts\migrate.ps1 version    # PowerShell
make migrate-version             # Make
```

**Create new migration:**
```bash
./scripts/migrate.sh create add_notifications      # Bash
.\scripts\migrate.ps1 create add_notifications     # PowerShell
make migrate-create name=add_notifications         # Make
```

**Verify migrations:**
```bash
./scripts/verify.sh              # Bash
make db-verify                   # Make
```

### Development Workflow

1. **Start fresh:**
   ```bash
   make db-setup    # Creates DB and runs migrations
   ```

2. **Make schema changes:**
   ```bash
   make migrate-create name=add_new_feature
   # Edit generated .up.sql and .down.sql files
   ```

3. **Test migration:**
   ```bash
   make migrate-up-1    # Apply new migration
   make migrate-down    # Test rollback
   make migrate-up-1    # Reapply
   ```

4. **Verify changes:**
   ```bash
   make db-verify
   ./scripts/verify.sh
   ```

5. **Reset if needed:**
   ```bash
   make migrate-reset   # Rollback all and reapply
   ```

### Production Deployment

1. **Backup:**
   ```bash
   make db-backup
   ```

2. **Test on staging:**
   ```bash
   DATABASE_URL="postgresql://user:pass@staging:5432/db" make migrate-up
   ```

3. **Deploy to production:**
   ```bash
   DATABASE_URL="postgresql://user:pass@prod:5432/db" make migrate-up
   ```

4. **Verify:**
   ```bash
   DATABASE_URL="postgresql://user:pass@prod:5432/db" make migrate-version
   ```

---

## File Summary

### Total Files Created: 16

**Migration Files:** 8
- 4 .up.sql files
- 4 .down.sql files

**Documentation:** 4
- README.md
- QUICK_START.md
- MIGRATION_SUMMARY.md
- INDEX.md (this file)

**Scripts:** 4
- migrate.sh (Bash)
- migrate.ps1 (PowerShell)
- verify.sh (Bash)
- verify_migrations.sql (SQL)

**Build Files:** 1
- Makefile

### Database Objects Created: 80+

**Tables:** 11
- users, manufacturers, audit_log, sessions
- payments, xrpl_payments, exchange_rates, xrpl_transaction_log, xrpl_settlement_batches
- tax_reports, tax_rates

**Views:** 7
- dashboard_metrics, payment_summary, xrpl_payment_stats
- manufacturer_report_summary, recent_activity
- tax_rate_history, daily_settlement_summary

**Functions:** 4
- update_updated_at_column()
- get_pending_xrpl_payment_by_tag()
- get_current_tax_rate()
- calculate_report_tax()

**Triggers:** 5
- update_users_updated_at
- update_manufacturers_updated_at
- update_payments_updated_at
- update_xrpl_payments_updated_at
- update_tax_reports_updated_at

**Indexes:** 60+
- Foreign key indexes
- Query optimization indexes
- Partial indexes for filtered queries
- Composite indexes for complex queries

**Constraints:** 30+
- Primary keys (11)
- Foreign keys (15+)
- Check constraints (20+)
- Unique constraints (10+)

---

## Quick Reference

### File Locations

All migration files:
```
C:\Users\justin.harvey\excise-tax-portal\backend\migrations\
├── 000001_init_schema.up.sql
├── 000001_init_schema.down.sql
├── 000002_add_xrpl_tables.up.sql
├── 000002_add_xrpl_tables.down.sql
├── 000003_add_tax_tables.up.sql
├── 000003_add_tax_tables.down.sql
├── 000004_add_views.up.sql
├── 000004_add_views.down.sql
├── README.md
├── QUICK_START.md
├── MIGRATION_SUMMARY.md
└── INDEX.md
```

All scripts:
```
C:\Users\justin.harvey\excise-tax-portal\backend\scripts\
├── migrate.sh
├── migrate.ps1
├── verify.sh
└── verify_migrations.sql
```

Build automation:
```
C:\Users\justin.harvey\excise-tax-portal\backend\
└── Makefile
```

### Environment Variables

Required for migration scripts:
```
DB_HOST        Database host (default: localhost)
DB_PORT        Database port (default: 5432)
DB_USER        Database user (default: postgres)
DB_PASSWORD    Database password (default: postgres)
DB_NAME        Database name (default: excise_tax_portal)
DATABASE_URL   Complete connection string (optional override)
```

### Dependencies

**Required:**
- PostgreSQL 14+ (recommended: 16)
- golang-migrate v4.17.0+

**Optional:**
- Make (for Makefile usage)
- bash (for .sh scripts on Linux/macOS)
- PowerShell (for .ps1 scripts on Windows)

---

## Support & Resources

- **Quick Start:** [QUICK_START.md](./QUICK_START.md)
- **Full Documentation:** [README.md](./README.md)
- **Migration Summary:** [MIGRATION_SUMMARY.md](./MIGRATION_SUMMARY.md)
- **Backend Spec:** [../BACKEND_INFRASTRUCTURE_SPEC.md](../BACKEND_INFRASTRUCTURE_SPEC.md)
- **golang-migrate:** https://github.com/golang-migrate/migrate
- **PostgreSQL:** https://www.postgresql.org/docs/

---

**Created:** 2025-12-29
**Last Updated:** 2025-12-29
**Version:** 1.0.0
