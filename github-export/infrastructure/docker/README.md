# Excise Tax Portal - Docker Infrastructure

Complete Docker Compose setup for local development of the Excise Tax Portal application.

## Table of Contents

- [Services Overview](#services-overview)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Service Details](#service-details)
- [Management Commands](#management-commands)
- [Accessing Services](#accessing-services)
- [Monitoring and Observability](#monitoring-and-observability)
- [Data Persistence](#data-persistence)
- [Troubleshooting](#troubleshooting)
- [Environment Variables](#environment-variables)
- [Advanced Configuration](#advanced-configuration)

## Services Overview

This Docker Compose configuration provides the following services:

| Service | Version | Ports | Purpose |
|---------|---------|-------|---------|
| PostgreSQL | 15-alpine | 5432 | Primary relational database |
| Redis | 7-alpine | 6379 | Caching and session storage |
| RabbitMQ | 3.12-management | 5672, 15672 | Message broker for async tasks |
| Prometheus | latest | 9090 | Metrics collection and monitoring |
| Grafana | latest | 3001 | Metrics visualization and dashboards |

## Prerequisites

Before starting, ensure you have the following installed:

- **Docker Desktop** (Windows): Version 4.0 or higher
  - Download from: https://www.docker.com/products/docker-desktop
- **Docker Compose**: Version 2.0 or higher (included with Docker Desktop)
- **Minimum System Requirements**:
  - 8 GB RAM (16 GB recommended)
  - 20 GB free disk space
  - CPU with virtualization support enabled

## Quick Start

### 1. Navigate to Docker Directory

```bash
cd C:\Users\justin.harvey\excise-tax-portal\infrastructure\docker
```

### 2. Copy Environment File

```bash
copy .env.docker .env
```

### 3. Start All Services

```bash
docker-compose up -d
```

### 4. Verify Services are Running

```bash
docker-compose ps
```

All services should show "healthy" or "running" status within 1-2 minutes.

### 5. Check Service Health

```bash
# Check all logs
docker-compose logs

# Check specific service
docker-compose logs postgres
```

## Service Details

### PostgreSQL Database

**Container Name**: excise-tax-postgres
**Image**: postgres:15-alpine
**Port**: 5432

**Features**:
- Pre-configured database: `excise_tax_db`
- Installed extensions: uuid-ossp, pgcrypto, pg_trgm, citext, btree_gin
- Additional schemas: audit, reporting, integration
- Optimized performance settings
- Automatic initialization via init.sql

**Connection Details**:
```
Host: localhost
Port: 5432
Database: excise_tax_db
Username: postgres
Password: postgres (default, change in .env)
```

**Connection String**:
```
postgresql://postgres:postgres@localhost:5432/excise_tax_db
```

### Redis Cache

**Container Name**: excise-tax-redis
**Image**: redis:7-alpine
**Port**: 6379

**Features**:
- AOF (Append Only File) persistence enabled
- Password protection enabled
- Optimized for caching and session storage

**Connection Details**:
```
Host: localhost
Port: 6379
Password: redis (default, change in .env)
Database: 0
```

**Connection String**:
```
redis://:redis@localhost:6379/0
```

### RabbitMQ Message Broker

**Container Name**: excise-tax-rabbitmq
**Image**: rabbitmq:3.12-management-alpine
**Ports**: 5672 (AMQP), 15672 (Management UI)

**Features**:
- Management plugin enabled
- Default virtual host: excise_tax
- Web-based management interface
- Message persistence enabled

**Connection Details**:
```
Host: localhost
Port: 5672
Username: admin (default, change in .env)
Password: admin (default, change in .env)
Virtual Host: excise_tax
```

**Management UI**: http://localhost:15672

### Prometheus Monitoring

**Container Name**: excise-tax-prometheus
**Image**: prom/prometheus:latest
**Port**: 9090

**Features**:
- 30-day metric retention
- Pre-configured scrape targets
- Self-monitoring enabled
- Ready for application metrics

**Access**: http://localhost:9090

**Configuration**: `./prometheus/prometheus.yml`

### Grafana Dashboards

**Container Name**: excise-tax-grafana
**Image**: grafana/grafana:latest
**Port**: 3001

**Features**:
- Prometheus datasource pre-configured
- Default admin credentials
- Plugin support enabled
- Ready for custom dashboards

**Access**: http://localhost:3001
**Default Credentials**:
- Username: admin
- Password: admin (change on first login)

## Management Commands

### Starting Services

```bash
# Start all services in background
docker-compose up -d

# Start all services with logs
docker-compose up

# Start specific service
docker-compose up -d postgres
```

### Stopping Services

```bash
# Stop all services (keeps containers)
docker-compose stop

# Stop and remove all containers
docker-compose down

# Stop and remove all containers + volumes (WARNING: deletes all data)
docker-compose down -v

# Stop specific service
docker-compose stop postgres
```

### Viewing Logs

```bash
# View all logs
docker-compose logs

# Follow logs in real-time
docker-compose logs -f

# View logs for specific service
docker-compose logs -f postgres

# View last 100 lines
docker-compose logs --tail=100

# View logs since specific time
docker-compose logs --since 30m
```

### Restarting Services

```bash
# Restart all services
docker-compose restart

# Restart specific service
docker-compose restart postgres

# Rebuild and restart (after config changes)
docker-compose up -d --force-recreate
```

### Checking Service Status

```bash
# List all containers
docker-compose ps

# View resource usage
docker stats

# Check specific container health
docker inspect excise-tax-postgres --format='{{json .State.Health}}'
```

## Accessing Services

### PostgreSQL Database Access

#### Using psql CLI

```bash
# Access PostgreSQL shell
docker exec -it excise-tax-postgres psql -U postgres -d excise_tax_db

# Run SQL file
docker exec -i excise-tax-postgres psql -U postgres -d excise_tax_db < your-script.sql
```

#### Using GUI Tools

Connect with tools like pgAdmin, DBeaver, or DataGrip:
- Host: localhost
- Port: 5432
- Database: excise_tax_db
- Username: postgres
- Password: postgres

### Redis Access

```bash
# Access Redis CLI
docker exec -it excise-tax-redis redis-cli -a redis

# Test connection
docker exec -it excise-tax-redis redis-cli -a redis ping
```

### RabbitMQ Management

**Web UI**: http://localhost:15672
- Username: admin
- Password: admin

**CLI Access**:
```bash
# List queues
docker exec -it excise-tax-rabbitmq rabbitmqctl list_queues

# List exchanges
docker exec -it excise-tax-rabbitmq rabbitmqctl list_exchanges

# Check node status
docker exec -it excise-tax-rabbitmq rabbitmqctl status
```

## Monitoring and Observability

### Prometheus Metrics

1. Access Prometheus: http://localhost:9090
2. Navigate to **Status > Targets** to see all monitored services
3. Use **Graph** tab to query metrics:
   - `up{job="prometheus"}` - Service availability
   - `process_cpu_seconds_total` - CPU usage
   - `process_resident_memory_bytes` - Memory usage

### Grafana Dashboards

1. Access Grafana: http://localhost:3001
2. Login with admin/admin (change password on first login)
3. Navigate to **Configuration > Data Sources** to verify Prometheus connection
4. Import community dashboards:
   - PostgreSQL: Dashboard ID 9628
   - Redis: Dashboard ID 11835
   - RabbitMQ: Dashboard ID 10991

### Health Checks

```bash
# Check all service health
docker-compose ps

# PostgreSQL health
docker exec excise-tax-postgres pg_isready -U postgres

# Redis health
docker exec excise-tax-redis redis-cli -a redis ping

# RabbitMQ health
docker exec excise-tax-rabbitmq rabbitmq-diagnostics ping
```

## Data Persistence

All services use Docker volumes for data persistence:

| Volume | Service | Purpose |
|--------|---------|---------|
| excise-tax-postgres-data | PostgreSQL | Database files |
| excise-tax-redis-data | Redis | Cache data |
| excise-tax-rabbitmq-data | RabbitMQ | Queue data |
| excise-tax-rabbitmq-log | RabbitMQ | Log files |
| excise-tax-prometheus-data | Prometheus | Metrics data |
| excise-tax-grafana-data | Grafana | Dashboards & config |

### Volume Management

```bash
# List all volumes
docker volume ls

# Inspect volume
docker volume inspect excise-tax-postgres-data

# Backup PostgreSQL data
docker exec excise-tax-postgres pg_dump -U postgres excise_tax_db > backup.sql

# Restore PostgreSQL data
docker exec -i excise-tax-postgres psql -U postgres -d excise_tax_db < backup.sql

# Remove all volumes (WARNING: deletes all data)
docker-compose down -v
```

## Troubleshooting

### Service Won't Start

1. Check if ports are already in use:
```bash
netstat -an | findstr "5432 6379 5672 15672 9090 3001"
```

2. Check Docker logs:
```bash
docker-compose logs [service-name]
```

3. Restart Docker Desktop

4. Remove and recreate containers:
```bash
docker-compose down
docker-compose up -d
```

### PostgreSQL Connection Issues

```bash
# Check if PostgreSQL is ready
docker exec excise-tax-postgres pg_isready -U postgres

# View PostgreSQL logs
docker-compose logs postgres

# Verify PostgreSQL configuration
docker exec excise-tax-postgres cat /var/lib/postgresql/data/postgresql.conf
```

### Redis Connection Issues

```bash
# Test Redis connection
docker exec excise-tax-redis redis-cli -a redis ping

# Check Redis logs
docker-compose logs redis

# Monitor Redis commands
docker exec excise-tax-redis redis-cli -a redis monitor
```

### RabbitMQ Issues

```bash
# Check RabbitMQ status
docker exec excise-tax-rabbitmq rabbitmqctl status

# View RabbitMQ logs
docker-compose logs rabbitmq

# Reset RabbitMQ
docker-compose restart rabbitmq
```

### Performance Issues

1. Check resource usage:
```bash
docker stats
```

2. Increase Docker Desktop resources:
   - Open Docker Desktop
   - Settings > Resources
   - Increase CPU and Memory allocation

3. Check disk space:
```bash
docker system df
```

4. Clean up unused resources:
```bash
docker system prune -a
```

### Network Issues

```bash
# Inspect network
docker network inspect excise-tax-network

# Recreate network
docker-compose down
docker network rm excise-tax-network
docker-compose up -d
```

## Environment Variables

All environment variables are configured in `.env.docker`. Copy it to `.env` and modify as needed:

```bash
copy .env.docker .env
```

### Key Variables

**PostgreSQL**:
- `POSTGRES_DB`: Database name (default: excise_tax_db)
- `POSTGRES_USER`: Database user (default: postgres)
- `POSTGRES_PASSWORD`: Database password (default: postgres)

**Redis**:
- `REDIS_PASSWORD`: Redis password (default: redis)

**RabbitMQ**:
- `RABBITMQ_USER`: RabbitMQ username (default: admin)
- `RABBITMQ_PASSWORD`: RabbitMQ password (default: admin)
- `RABBITMQ_VHOST`: Virtual host (default: excise_tax)

**Grafana**:
- `GRAFANA_ADMIN_USER`: Admin username (default: admin)
- `GRAFANA_ADMIN_PASSWORD`: Admin password (default: admin)

### Security Note

For production deployments:
1. Change all default passwords
2. Use strong, randomly generated passwords
3. Store sensitive credentials in a secrets manager
4. Enable SSL/TLS for all services
5. Implement network isolation
6. Enable authentication on all services

## Advanced Configuration

### Resource Limits

Each service has resource limits defined in docker-compose.yml. Adjust as needed:

```yaml
deploy:
  resources:
    limits:
      cpus: '2.0'
      memory: 2G
    reservations:
      cpus: '0.5'
      memory: 512M
```

### Custom PostgreSQL Configuration

Modify `postgres/init.sql` to add custom initialization:
- Additional databases
- Custom extensions
- Initial data seeding
- Custom functions/procedures

### Prometheus Custom Scraping

Edit `prometheus/prometheus.yml` to add application metrics endpoints:

```yaml
- job_name: 'my-app'
  static_configs:
    - targets: ['my-app:3000']
  metrics_path: '/metrics'
```

### Grafana Custom Dashboards

1. Create dashboards in Grafana UI
2. Export as JSON
3. Place in `grafana/dashboards/` directory
4. Add provisioning config in `grafana/dashboards.yml`

### Adding New Services

To add new services to docker-compose.yml:

1. Define service under `services:` section
2. Add to `excise-tax-network`
3. Create named volume if persistence needed
4. Add health check
5. Configure resource limits
6. Update this README

## Network Architecture

All services communicate via the `excise-tax-network` bridge network:

- Subnet: 172.28.0.0/16
- Internal DNS resolution enabled
- Services can reference each other by container name
- External access via mapped ports

**Service Discovery**:
- From host: Use `localhost` and mapped ports
- Between containers: Use container name (e.g., `postgres:5432`)

## Backup and Recovery

### Automated Backup Script

Create a backup script `backup.sh`:

```bash
#!/bin/bash
BACKUP_DIR="./backups"
DATE=$(date +%Y%m%d_%H%M%S)

# Create backup directory
mkdir -p $BACKUP_DIR

# Backup PostgreSQL
docker exec excise-tax-postgres pg_dump -U postgres excise_tax_db > $BACKUP_DIR/postgres_$DATE.sql

# Backup Redis
docker exec excise-tax-redis redis-cli -a redis --rdb $BACKUP_DIR/redis_$DATE.rdb

echo "Backup completed: $DATE"
```

### Restore from Backup

```bash
# Restore PostgreSQL
docker exec -i excise-tax-postgres psql -U postgres -d excise_tax_db < backups/postgres_20250101_120000.sql

# Restore Redis
docker cp backups/redis_20250101_120000.rdb excise-tax-redis:/data/dump.rdb
docker-compose restart redis
```

## Development Workflow

### Daily Development

1. Start services:
```bash
docker-compose up -d
```

2. Develop your application (connects to services on localhost)

3. View logs if needed:
```bash
docker-compose logs -f
```

4. Stop services when done:
```bash
docker-compose stop
```

### Clean Restart

```bash
# Stop and remove containers
docker-compose down

# Start fresh
docker-compose up -d

# Or rebuild everything
docker-compose up -d --force-recreate
```

### Testing Database Migrations

```bash
# Create backup before migration
docker exec excise-tax-postgres pg_dump -U postgres excise_tax_db > pre_migration_backup.sql

# Run migration
# (your migration commands here)

# If migration fails, restore backup
docker exec -i excise-tax-postgres psql -U postgres -d excise_tax_db < pre_migration_backup.sql
```

## Support and Documentation

- **Docker Documentation**: https://docs.docker.com/
- **PostgreSQL Documentation**: https://www.postgresql.org/docs/15/
- **Redis Documentation**: https://redis.io/documentation
- **RabbitMQ Documentation**: https://www.rabbitmq.com/documentation.html
- **Prometheus Documentation**: https://prometheus.io/docs/
- **Grafana Documentation**: https://grafana.com/docs/

## Contributing

When adding new infrastructure components:

1. Update docker-compose.yml
2. Add configuration files in respective directories
3. Update .env.docker with new variables
4. Update this README with service details
5. Add health checks and resource limits
6. Test thoroughly in development environment

## License

Internal use only - Excise Tax Portal Project
