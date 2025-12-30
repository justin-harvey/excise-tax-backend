# Docker Infrastructure Troubleshooting Guide

Common issues and solutions for the Excise Tax Portal Docker infrastructure.

## Table of Contents

- [General Issues](#general-issues)
- [PostgreSQL Issues](#postgresql-issues)
- [Redis Issues](#redis-issues)
- [RabbitMQ Issues](#rabbitmq-issues)
- [Prometheus Issues](#prometheus-issues)
- [Grafana Issues](#grafana-issues)
- [Network Issues](#network-issues)
- [Performance Issues](#performance-issues)
- [Data Issues](#data-issues)

## General Issues

### Docker Desktop Not Running

**Symptoms:**
- Error: "Cannot connect to the Docker daemon"
- Error: "docker: command not found"

**Solutions:**

1. **Check if Docker Desktop is running:**
   - Look for Docker icon in Windows system tray
   - If not present, start Docker Desktop from Start menu

2. **Restart Docker Desktop:**
   ```powershell
   # Right-click Docker icon > Restart
   # Or from PowerShell:
   Stop-Service docker
   Start-Service docker
   ```

3. **Verify Docker is working:**
   ```powershell
   docker --version
   docker ps
   ```

### Port Already in Use

**Symptoms:**
- Error: "Bind for 0.0.0.0:5432 failed: port is already allocated"

**Solutions:**

1. **Find process using the port:**
   ```powershell
   # For PostgreSQL (5432)
   netstat -ano | findstr :5432

   # For Redis (6379)
   netstat -ano | findstr :6379

   # For RabbitMQ (5672, 15672)
   netstat -ano | findstr :5672
   netstat -ano | findstr :15672
   ```

2. **Stop the conflicting process:**
   ```powershell
   # Replace PID with actual process ID from netstat output
   taskkill /PID <PID> /F
   ```

3. **Or change the port in docker-compose.override.yml:**
   ```yaml
   services:
     postgres:
       ports:
         - "5433:5432"  # Use 5433 instead of 5432
   ```

### Services Won't Start

**Symptoms:**
- Container immediately exits
- Container status shows "Exited (1)"

**Solutions:**

1. **Check logs:**
   ```powershell
   docker-compose logs [service-name]
   ```

2. **Check Docker Desktop resources:**
   - Open Docker Desktop
   - Settings > Resources
   - Ensure at least:
     - 8 GB RAM
     - 4 CPUs
     - 20 GB disk space

3. **Remove and recreate containers:**
   ```powershell
   docker-compose down
   docker-compose up -d
   ```

4. **Check for disk space:**
   ```powershell
   docker system df
   ```

### Out of Memory Errors

**Symptoms:**
- Container crashes randomly
- Error: "Cannot allocate memory"

**Solutions:**

1. **Increase Docker memory limit:**
   - Docker Desktop > Settings > Resources
   - Increase Memory to 8-16 GB
   - Click Apply & Restart

2. **Reduce resource limits in docker-compose.yml:**
   ```yaml
   deploy:
     resources:
       limits:
         memory: 1G  # Reduce from 2G
   ```

3. **Free up system memory:**
   ```powershell
   # Clean Docker cache
   docker system prune -a
   ```

## PostgreSQL Issues

### Cannot Connect to PostgreSQL

**Symptoms:**
- Error: "could not connect to server"
- Error: "FATAL: password authentication failed"

**Solutions:**

1. **Check if PostgreSQL is running:**
   ```powershell
   docker exec excise-tax-postgres pg_isready -U postgres
   ```

2. **Check PostgreSQL logs:**
   ```powershell
   docker-compose logs postgres
   ```

3. **Verify credentials:**
   - Check `.env` file for correct POSTGRES_USER and POSTGRES_PASSWORD
   - Default: postgres/postgres

4. **Wait for PostgreSQL to fully start:**
   ```powershell
   # PostgreSQL can take 30-60 seconds to start
   Start-Sleep -Seconds 30
   .\healthcheck.ps1
   ```

5. **Test connection:**
   ```powershell
   docker exec -it excise-tax-postgres psql -U postgres -d excise_tax_db
   ```

### Database Not Found

**Symptoms:**
- Error: "database 'excise_tax_db' does not exist"

**Solutions:**

1. **Check if database was created:**
   ```powershell
   docker exec excise-tax-postgres psql -U postgres -c "\l"
   ```

2. **Manually create database:**
   ```powershell
   docker exec excise-tax-postgres psql -U postgres -c "CREATE DATABASE excise_tax_db;"
   ```

3. **Recreate container with volume reset:**
   ```powershell
   docker-compose down -v
   docker-compose up -d
   ```

### Extensions Not Installed

**Symptoms:**
- Error: "extension 'uuid-ossp' is not available"

**Solutions:**

1. **Check installed extensions:**
   ```powershell
   docker exec excise-tax-postgres psql -U postgres -d excise_tax_db -c "SELECT * FROM pg_extension;"
   ```

2. **Manually install extensions:**
   ```powershell
   docker exec excise-tax-postgres psql -U postgres -d excise_tax_db -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";"
   ```

3. **Re-run initialization script:**
   ```powershell
   docker exec -i excise-tax-postgres psql -U postgres -d excise_tax_db < postgres/init.sql
   ```

### PostgreSQL Performance Issues

**Symptoms:**
- Slow queries
- High CPU usage

**Solutions:**

1. **Check active connections:**
   ```powershell
   docker exec excise-tax-postgres psql -U postgres -d excise_tax_db -c "SELECT count(*) FROM pg_stat_activity;"
   ```

2. **View slow queries:**
   ```powershell
   docker exec excise-tax-postgres psql -U postgres -d excise_tax_db -c "SELECT query, calls, total_time FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;"
   ```

3. **Increase shared_buffers:**
   ```powershell
   docker exec excise-tax-postgres psql -U postgres -c "ALTER SYSTEM SET shared_buffers = '512MB';"
   docker-compose restart postgres
   ```

## Redis Issues

### Cannot Connect to Redis

**Symptoms:**
- Error: "Could not connect to Redis"
- Error: "NOAUTH Authentication required"

**Solutions:**

1. **Check if Redis is running:**
   ```powershell
   docker exec excise-tax-redis redis-cli -a redis ping
   ```

2. **Check Redis logs:**
   ```powershell
   docker-compose logs redis
   ```

3. **Verify password:**
   - Check `.env` file for REDIS_PASSWORD
   - Default: redis

4. **Test connection with password:**
   ```powershell
   docker exec excise-tax-redis redis-cli -a redis INFO server
   ```

### Redis Out of Memory

**Symptoms:**
- Error: "OOM command not allowed when used memory"

**Solutions:**

1. **Check memory usage:**
   ```powershell
   docker exec excise-tax-redis redis-cli -a redis INFO memory
   ```

2. **Clear cache:**
   ```powershell
   docker exec excise-tax-redis redis-cli -a redis FLUSHALL
   ```

3. **Increase memory limit:**
   ```yaml
   # In docker-compose.yml
   redis:
     deploy:
       resources:
         limits:
           memory: 1G  # Increase from 512M
   ```

4. **Set maxmemory policy:**
   ```powershell
   docker exec excise-tax-redis redis-cli -a redis CONFIG SET maxmemory-policy allkeys-lru
   ```

### Redis Data Loss

**Symptoms:**
- Cache data missing after restart

**Solutions:**

1. **Check AOF persistence:**
   ```powershell
   docker exec excise-tax-redis redis-cli -a redis CONFIG GET appendonly
   ```

2. **Enable AOF if disabled:**
   ```powershell
   docker exec excise-tax-redis redis-cli -a redis CONFIG SET appendonly yes
   docker exec excise-tax-redis redis-cli -a redis BGREWRITEAOF
   ```

3. **Check volume mount:**
   ```powershell
   docker volume inspect excise-tax-redis-data
   ```

## RabbitMQ Issues

### Cannot Access RabbitMQ Management UI

**Symptoms:**
- http://localhost:15672 not accessible
- Error: "Connection refused"

**Solutions:**

1. **Check if RabbitMQ is running:**
   ```powershell
   docker exec excise-tax-rabbitmq rabbitmqctl status
   ```

2. **Check if management plugin is enabled:**
   ```powershell
   docker exec excise-tax-rabbitmq rabbitmq-plugins list
   ```

3. **Enable management plugin:**
   ```powershell
   docker exec excise-tax-rabbitmq rabbitmq-plugins enable rabbitmq_management
   docker-compose restart rabbitmq
   ```

4. **Wait for RabbitMQ to fully start:**
   - RabbitMQ can take 60-90 seconds to start
   - Check logs: `docker-compose logs rabbitmq`

5. **Verify credentials:**
   - Username: admin
   - Password: admin (from .env file)

### RabbitMQ Connection Failed

**Symptoms:**
- Error: "Cannot connect to AMQP broker"

**Solutions:**

1. **Check if port 5672 is accessible:**
   ```powershell
   Test-NetConnection -ComputerName localhost -Port 5672
   ```

2. **Check RabbitMQ logs:**
   ```powershell
   docker-compose logs rabbitmq
   ```

3. **Verify virtual host exists:**
   ```powershell
   docker exec excise-tax-rabbitmq rabbitmqctl list_vhosts
   ```

4. **Create virtual host if missing:**
   ```powershell
   docker exec excise-tax-rabbitmq rabbitmqctl add_vhost excise_tax
   docker exec excise-tax-rabbitmq rabbitmqctl set_permissions -p excise_tax admin ".*" ".*" ".*"
   ```

### Message Queue Buildup

**Symptoms:**
- High memory usage
- Slow message processing

**Solutions:**

1. **Check queue status:**
   ```powershell
   docker exec excise-tax-rabbitmq rabbitmqctl list_queues name messages consumers
   ```

2. **Purge specific queue:**
   ```powershell
   docker exec excise-tax-rabbitmq rabbitmqctl purge_queue queue_name
   ```

3. **Increase consumer count in your application**

4. **Check for dead letter exchanges:**
   ```powershell
   docker exec excise-tax-rabbitmq rabbitmqctl list_exchanges
   ```

## Prometheus Issues

### Prometheus Not Collecting Metrics

**Symptoms:**
- No data in Prometheus UI
- Targets showing "Down"

**Solutions:**

1. **Check Prometheus targets:**
   - Open http://localhost:9090/targets
   - Check which targets are up/down

2. **Check Prometheus configuration:**
   ```powershell
   docker exec excise-tax-prometheus cat /etc/prometheus/prometheus.yml
   ```

3. **Validate configuration:**
   ```powershell
   docker exec excise-tax-prometheus promtool check config /etc/prometheus/prometheus.yml
   ```

4. **Restart Prometheus:**
   ```powershell
   docker-compose restart prometheus
   ```

5. **Check if services expose metrics endpoints:**
   - Services need to expose /metrics endpoint
   - Verify in application configuration

### Prometheus High Memory Usage

**Symptoms:**
- Container crashes
- Slow query performance

**Solutions:**

1. **Reduce retention time:**
   ```yaml
   # In docker-compose.yml, change command:
   - '--storage.tsdb.retention.time=7d'  # Reduce from 30d
   ```

2. **Increase memory limit:**
   ```yaml
   prometheus:
     deploy:
       resources:
         limits:
           memory: 2G  # Increase from 1G
   ```

3. **Reduce scrape frequency:**
   ```yaml
   # In prometheus/prometheus.yml
   global:
     scrape_interval: 30s  # Increase from 15s
   ```

## Grafana Issues

### Cannot Login to Grafana

**Symptoms:**
- Invalid username/password
- Login page not loading

**Solutions:**

1. **Check Grafana logs:**
   ```powershell
   docker-compose logs grafana
   ```

2. **Verify credentials:**
   - Default: admin/admin
   - Check .env file for GRAFANA_ADMIN_USER and GRAFANA_ADMIN_PASSWORD

3. **Reset admin password:**
   ```powershell
   docker exec excise-tax-grafana grafana-cli admin reset-admin-password newpassword
   ```

4. **Check if Grafana is healthy:**
   ```powershell
   Invoke-WebRequest -Uri "http://localhost:3001/api/health" -UseBasicParsing
   ```

### Prometheus Datasource Not Working

**Symptoms:**
- "Data source not found"
- No data in dashboards

**Solutions:**

1. **Check datasource configuration:**
   - Grafana > Configuration > Data Sources
   - Verify Prometheus URL: http://prometheus:9090

2. **Test datasource connection:**
   - Click "Test" button in datasource settings

3. **Verify Prometheus is accessible from Grafana:**
   ```powershell
   docker exec excise-tax-grafana wget -O- http://prometheus:9090/api/v1/status/config
   ```

4. **Recreate datasource:**
   ```powershell
   docker-compose down
   docker volume rm excise-tax-grafana-data
   docker-compose up -d
   ```

### Dashboard Not Loading Data

**Symptoms:**
- "No data" in panels
- Empty graphs

**Solutions:**

1. **Check time range:**
   - Ensure time range in top-right corner includes recent data

2. **Verify query syntax:**
   - Click "Edit" on panel
   - Check Prometheus query is valid

3. **Check Prometheus has data:**
   - Open Prometheus UI: http://localhost:9090
   - Run same query in Prometheus

4. **Refresh dashboard:**
   - Click refresh button
   - Set auto-refresh interval

## Network Issues

### Containers Cannot Communicate

**Symptoms:**
- Application cannot connect to database
- Services cannot reach each other

**Solutions:**

1. **Check network exists:**
   ```powershell
   docker network ls | findstr excise-tax
   ```

2. **Inspect network:**
   ```powershell
   docker network inspect excise-tax-network
   ```

3. **Verify containers are on same network:**
   ```powershell
   docker inspect excise-tax-postgres --format='{{json .NetworkSettings.Networks}}'
   ```

4. **Recreate network:**
   ```powershell
   docker-compose down
   docker network rm excise-tax-network
   docker-compose up -d
   ```

5. **Test connectivity between containers:**
   ```powershell
   docker exec excise-tax-grafana ping -c 4 prometheus
   ```

### DNS Resolution Not Working

**Symptoms:**
- Error: "Could not resolve host"

**Solutions:**

1. **Use container name instead of localhost:**
   - From container: use `postgres` not `localhost`
   - From host: use `localhost`

2. **Check Docker DNS settings:**
   - Docker Desktop > Settings > Network
   - Try changing DNS server

3. **Restart Docker Desktop**

## Performance Issues

### Slow Container Startup

**Symptoms:**
- Containers take several minutes to start

**Solutions:**

1. **Check disk space:**
   ```powershell
   docker system df
   ```

2. **Clean up unused resources:**
   ```powershell
   docker system prune -a
   ```

3. **Increase Docker Desktop resources:**
   - Settings > Resources
   - Increase CPU and RAM allocation

4. **Use SSD instead of HDD**

5. **Disable antivirus scanning on Docker directories:**
   - Add Docker data directory to antivirus exclusions

### High CPU Usage

**Symptoms:**
- System becomes slow
- Docker Desktop shows high CPU

**Solutions:**

1. **Check which container is using CPU:**
   ```powershell
   docker stats
   ```

2. **Reduce resource limits:**
   ```yaml
   # In docker-compose.yml
   deploy:
     resources:
       limits:
         cpus: '1.0'  # Reduce from 2.0
   ```

3. **Check for infinite loops in application logs**

4. **Restart problematic container:**
   ```powershell
   docker-compose restart [service-name]
   ```

### Slow Query Performance

**Symptoms:**
- Database queries are slow
- Application timeouts

**Solutions:**

1. **Check PostgreSQL query performance:**
   ```powershell
   docker exec excise-tax-postgres psql -U postgres -d excise_tax_db -c "SELECT query, mean_time, calls FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"
   ```

2. **Add indexes:**
   ```sql
   CREATE INDEX idx_column_name ON table_name(column_name);
   ```

3. **Analyze tables:**
   ```powershell
   docker exec excise-tax-postgres psql -U postgres -d excise_tax_db -c "ANALYZE;"
   ```

4. **Increase PostgreSQL memory:**
   ```yaml
   postgres:
     deploy:
       resources:
         limits:
           memory: 4G  # Increase from 2G
   ```

## Data Issues

### Data Lost After Restart

**Symptoms:**
- Database empty after restart
- Cache cleared

**Solutions:**

1. **Check if volumes exist:**
   ```powershell
   docker volume ls
   ```

2. **Inspect volume:**
   ```powershell
   docker volume inspect excise-tax-postgres-data
   ```

3. **Don't use `down -v` flag:**
   ```powershell
   # This DELETES volumes
   docker-compose down -v  # DON'T DO THIS unless intentional

   # Instead use:
   docker-compose down  # Keeps volumes
   ```

4. **Restore from backup:**
   ```powershell
   .\manage.ps1 restore
   ```

### Volume Permission Issues

**Symptoms:**
- Error: "Permission denied"
- Cannot write to volume

**Solutions:**

1. **Check volume permissions:**
   ```powershell
   docker exec excise-tax-postgres ls -la /var/lib/postgresql/data
   ```

2. **Fix permissions:**
   ```powershell
   docker exec -u root excise-tax-postgres chown -R postgres:postgres /var/lib/postgresql/data
   ```

3. **Recreate volume:**
   ```powershell
   docker-compose down
   docker volume rm excise-tax-postgres-data
   docker-compose up -d
   ```

### Backup/Restore Failed

**Symptoms:**
- Backup script fails
- Cannot restore data

**Solutions:**

1. **Check if container is running:**
   ```powershell
   docker-compose ps
   ```

2. **Manually backup:**
   ```powershell
   # PostgreSQL
   docker exec excise-tax-postgres pg_dump -U postgres excise_tax_db > backup.sql

   # Redis
   docker exec excise-tax-redis redis-cli -a redis SAVE
   docker cp excise-tax-redis:/data/dump.rdb backup.rdb
   ```

3. **Manually restore:**
   ```powershell
   # PostgreSQL
   Get-Content backup.sql | docker exec -i excise-tax-postgres psql -U postgres -d excise_tax_db

   # Redis
   docker cp backup.rdb excise-tax-redis:/data/dump.rdb
   docker-compose restart redis
   ```

## Advanced Troubleshooting

### Enable Debug Logging

**PostgreSQL:**
```powershell
docker exec excise-tax-postgres psql -U postgres -c "ALTER SYSTEM SET log_min_messages = 'DEBUG1';"
docker-compose restart postgres
```

**Redis:**
```powershell
docker exec excise-tax-redis redis-cli -a redis CONFIG SET loglevel debug
```

### Check Container Resource Usage

```powershell
# Real-time stats
docker stats

# Container inspect
docker inspect excise-tax-postgres

# System-wide info
docker system info
docker system df
```

### Export Container Logs

```powershell
# All logs to file
docker-compose logs > all-logs.txt

# Specific service
docker-compose logs postgres > postgres-logs.txt

# With timestamps
docker-compose logs -t > logs-with-timestamps.txt
```

### Reset Everything (Nuclear Option)

**WARNING: This deletes ALL data!**

```powershell
# Stop and remove everything
docker-compose down -v

# Remove all containers
docker rm -f $(docker ps -aq)

# Remove all volumes
docker volume prune -f

# Remove all networks
docker network prune -f

# Remove all images
docker image prune -a -f

# Clean everything
docker system prune -a -f --volumes

# Start fresh
docker-compose up -d
```

## Getting Help

If you can't resolve the issue:

1. **Run comprehensive health check:**
   ```powershell
   .\healthcheck.ps1 -Detailed
   ```

2. **Collect diagnostic information:**
   ```powershell
   docker-compose ps > diagnostic.txt
   docker-compose logs >> diagnostic.txt
   docker system info >> diagnostic.txt
   docker system df >> diagnostic.txt
   ```

3. **Check Docker Desktop status:**
   - Open Docker Desktop
   - Check for updates
   - View Troubleshoot > Get Support

4. **Contact team support with:**
   - Error messages
   - Service logs
   - Health check output
   - Steps to reproduce

## Useful Commands Reference

```powershell
# Health checks
.\healthcheck.ps1
.\healthcheck.ps1 -Detailed
.\manage.ps1 health

# Service management
.\manage.ps1 start
.\manage.ps1 stop
.\manage.ps1 restart [service]

# Logs
.\manage.ps1 logs -Follow
docker-compose logs -f [service]

# Status
.\manage.ps1 status
docker-compose ps

# Cleanup
.\manage.ps1 clean
docker system prune

# Backup/Restore
.\manage.ps1 backup
.\manage.ps1 restore

# Complete reset
docker-compose down -v
docker-compose up -d
```

## Prevention Tips

1. **Always use named volumes for persistence**
2. **Regular backups with `.\manage.ps1 backup`**
3. **Monitor resource usage with `docker stats`**
4. **Keep Docker Desktop updated**
5. **Allocate sufficient resources (8+ GB RAM)**
6. **Don't use `docker-compose down -v` unless intentional**
7. **Check logs regularly for warnings**
8. **Run health checks before reporting issues**
9. **Use `.env` file for configuration**
10. **Test in development before production**

---

**Last Updated:** 2025-12-29
**Version:** 1.0
