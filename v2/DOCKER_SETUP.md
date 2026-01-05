# Docker Setup for Excise Tax Backend v2

## Overview

The v2 API runs in Docker with PostgreSQL for data persistence. The setup includes:

- **PostgreSQL 15** - Database with JSONB support for payments
- **XRPL API (Development)** - Hot-reload development server
- **XRPL API (Production)** - Optimized production build
- **Swagger UI** - API documentation viewer

## Quick Start

### 1. Set Up Environment Variables

Copy the example environment file:

```bash
cp .env.example .env
```

Edit `.env` and set secure passwords:

```bash
# Generate secure password
openssl rand -base64 32

# Update .env file with generated values
POSTGRES_PASSWORD=<generated_password>
JWT_SECRET=<generated_secret>
```

### 2. Start Development Environment

Start API with PostgreSQL:

```bash
docker compose --profile dev up
```

This starts:
- PostgreSQL on port 5432
- API on port 8080
- Swagger UI on port 8081

### 3. Run Database Migrations

Once PostgreSQL is running, apply migrations:

```bash
# Connect to database
docker exec -it excise-tax-postgres psql -U postgres -d excise_tax_db

# Or run migrations from host
psql -h localhost -U postgres -d excise_tax_db -f migrations/000001_init_schema.up.sql
psql -h localhost -U postgres -d excise_tax_db -f migrations/000005_add_payments.up.sql
```

### 4. Test the API

```bash
# Health check
curl http://localhost:8080/health

# Create a payment
curl -X POST http://localhost:8080/payments \
  -H "Content-Type: application/json" \
  -d '{
    "type": "xrpl",
    "amount": "100",
    "currency": "XRP",
    "from": "rSenderAddress",
    "to": "rReceiverAddress"
  }'

# List payments
curl http://localhost:8080/payments

# View API docs
open http://localhost:8081
```

## Docker Profiles

### Development (`dev`)
```bash
docker compose --profile dev up
```
- Hot reload enabled
- Debug logging
- Text log format
- Connected to testnet

### Production (`prod`)
```bash
docker compose --profile prod up
```
- Optimized build
- JSON logging
- Connected to mainnet
- SSL required for DB

### Test
```bash
docker compose --profile test up
```
- Runs test suite with race detector
- Generates coverage report
- Isolated environment

## Database Access

### Connect to PostgreSQL

```bash
# Via Docker
docker exec -it excise-tax-postgres psql -U postgres -d excise_tax_db

# From host (requires psql installed)
psql -h localhost -U postgres -d excise_tax_db
```

### Useful Commands

```sql
-- List tables
\dt

-- Describe payments table
\d payments

-- View all payments
SELECT id, type, state, created_at FROM payments;

-- Query payment events
SELECT 
  id,
  jsonb_array_length(events) as event_count
FROM payments;

-- Find pending payments
SELECT * FROM payments WHERE state = 'pending';
```

### Backup and Restore

```bash
# Backup
docker exec excise-tax-postgres pg_dump -U postgres excise_tax_db > backup.sql

# Restore
cat backup.sql | docker exec -i excise-tax-postgres psql -U postgres -d excise_tax_db
```

## Volumes

Data is persisted in Docker volumes:

```bash
# View volumes
docker volume ls | grep excise-tax

# Inspect volume
docker volume inspect excise-tax-postgres-data

# Remove volumes (WARNING: deletes all data)
docker compose down -v
```

## Networking

All services run on the `excise-tax-network` bridge network:

- Subnet: 172.28.0.0/16
- Services can communicate by container name
- API connects to PostgreSQL at `postgres:5432`

## Healthchecks

### PostgreSQL
```bash
docker exec excise-tax-postgres pg_isready -U postgres
```

### API
```bash
curl http://localhost:8080/health
```

### View Health Status
```bash
docker compose ps
```

## Logs

### View All Logs
```bash
docker compose logs -f
```

### View Specific Service
```bash
docker compose logs -f api-dev
docker compose logs -f postgres
```

### Limit Log Lines
```bash
docker compose logs --tail=100 api-dev
```

## Troubleshooting

### PostgreSQL Won't Start
```bash
# Check logs
docker compose logs postgres

# Verify .env file has POSTGRES_PASSWORD set
cat .env | grep POSTGRES_PASSWORD

# Check volume permissions
docker volume inspect excise-tax-postgres-data
```

### API Can't Connect to Database
```bash
# Verify PostgreSQL is healthy
docker compose ps postgres

# Check network connectivity
docker exec xrpl-api-dev ping -c 3 postgres

# Verify environment variables
docker exec xrpl-api-dev env | grep DB_
```

### Reset Everything
```bash
# Stop and remove everything
docker compose down -v

# Remove images
docker compose down --rmi all

# Start fresh
docker compose --profile dev up --build
```

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `POSTGRES_DB` | Database name | `excise_tax_db` | No |
| `POSTGRES_USER` | Database user | `postgres` | No |
| `POSTGRES_PASSWORD` | Database password | - | **Yes** |
| `POSTGRES_PORT` | Database port | `5432` | No |
| `API_PORT` | API server port | `8080` | No |
| `XRPL_URL` | XRPL node URL | testnet | No |
| `LOG_LEVEL` | Logging level | `debug` | No |
| `LOG_FORMAT` | Log format | `text` | No |

## Production Deployment

### Security Checklist

- [ ] Set strong `POSTGRES_PASSWORD`
- [ ] Set unique `JWT_SECRET`
- [ ] Use SSL for database (`DB_SSLMODE=require`)
- [ ] Use mainnet XRPL URL
- [ ] Set `LOG_FORMAT=json`
- [ ] Set `LOG_LEVEL=info` or `warn`
- [ ] Configure firewall rules
- [ ] Enable rate limiting
- [ ] Set up monitoring
- [ ] Configure backups

### Production Start

```bash
# Set production environment
export ENV=production

# Start services
docker compose --profile prod up -d

# Verify health
docker compose ps
curl http://localhost:8080/health
```

## Maintenance

### Update API

```bash
# Rebuild and restart
docker compose --profile dev up --build -d

# View logs
docker compose logs -f api-dev
```

### Database Migrations

```bash
# Apply new migration
psql -h localhost -U postgres -d excise_tax_db -f migrations/000006_new_migration.up.sql

# Rollback migration
psql -h localhost -U postgres -d excise_tax_db -f migrations/000006_new_migration.down.sql
```

### Monitor Resources

```bash
# View resource usage
docker stats

# View disk usage
docker system df
```

## Next Steps

1. ✅ Docker environment is set up
2. ✅ PostgreSQL is running
3. ✅ API connects to database
4. Run migrations: `psql -h localhost -U postgres -d excise_tax_db -f migrations/000005_add_payments.up.sql`
5. Test payment creation via API
6. Set up monitoring and alerting
7. Configure production deployment
