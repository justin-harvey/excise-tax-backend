# Quick Start Guide - Excise Tax Portal Docker Infrastructure

Get up and running with the Excise Tax Portal local development environment in under 5 minutes.

## Prerequisites Check

1. **Docker Desktop installed and running**
   - Windows: Check system tray for Docker icon
   - Verify: Open PowerShell and run `docker --version`

2. **Minimum 8 GB RAM available**
   - Check: Task Manager > Performance tab

3. **20 GB free disk space**
   - Check: File Explorer > This PC

## Step 1: Navigate to Docker Directory

```powershell
cd C:\Users\justin.harvey\excise-tax-portal\infrastructure\docker
```

## Step 2: Copy Environment File

```powershell
copy .env.docker .env
```

## Step 3: Start All Services

### Option A: Using Docker Compose (Recommended)

```powershell
docker-compose up -d
```

### Option B: Using Management Script

```powershell
.\manage.ps1 start
```

## Step 4: Wait for Services to Start (60-90 seconds)

Check status:

```powershell
.\manage.ps1 status
```

Or:

```powershell
docker-compose ps
```

All services should show "running" or "healthy" status.

## Step 5: Verify Services are Working

### Quick Health Check

```powershell
.\manage.ps1 health
```

### Manual Verification

**PostgreSQL**:
```powershell
docker exec -it excise-tax-postgres psql -U postgres -d excise_tax_db -c "SELECT version();"
```

**Redis**:
```powershell
docker exec -it excise-tax-redis redis-cli -a redis ping
# Should return: PONG
```

**RabbitMQ**:
Open browser: http://localhost:15672
- Username: admin
- Password: admin

**Prometheus**:
Open browser: http://localhost:9090

**Grafana**:
Open browser: http://localhost:3001
- Username: admin
- Password: admin

## Step 6: Start Developing!

Your application can now connect to:

- **PostgreSQL**: `postgresql://postgres:postgres@localhost:5432/excise_tax_db`
- **Redis**: `redis://:redis@localhost:6379/0`
- **RabbitMQ**: `amqp://admin:admin@localhost:5672/excise_tax`

## Common Operations

### View Logs

```powershell
# All services
.\manage.ps1 logs -Follow

# Specific service
.\manage.ps1 logs postgres -Follow
```

### Stop Services

```powershell
# Stop (keeps data)
.\manage.ps1 stop

# Or with docker-compose
docker-compose stop
```

### Restart Services

```powershell
.\manage.ps1 restart
```

### Complete Reset (WARNING: Deletes all data)

```powershell
docker-compose down -v
docker-compose up -d
```

## Troubleshooting

### Services won't start

1. **Check if ports are available**:
   ```powershell
   netstat -an | findstr "5432 6379 5672 15672 9090 3001"
   ```
   If any ports are in use, stop the conflicting application.

2. **Restart Docker Desktop**:
   - Right-click Docker icon in system tray
   - Select "Restart"

3. **Check Docker logs**:
   ```powershell
   docker-compose logs
   ```

### "Cannot connect to database"

```powershell
# Wait a bit longer - PostgreSQL can take 30-60 seconds
Start-Sleep -Seconds 30

# Check if PostgreSQL is ready
docker exec excise-tax-postgres pg_isready -U postgres
```

### "Out of memory" errors

1. Open Docker Desktop
2. Go to Settings > Resources
3. Increase Memory to at least 8 GB
4. Click "Apply & Restart"

### Port already in use

Find and stop the conflicting service:

```powershell
# Find process using port 5432 (PostgreSQL)
netstat -ano | findstr :5432

# Kill process (replace PID with actual process ID)
taskkill /PID <PID> /F
```

## Next Steps

1. **Configure your application** to use the services:
   - Update database connection strings
   - Configure Redis cache settings
   - Set up RabbitMQ queues

2. **Set up monitoring**:
   - Import Grafana dashboards
   - Configure Prometheus alerts
   - Set up application metrics

3. **Read the full documentation**:
   - See [README.md](README.md) for detailed information
   - Check service-specific configuration

## Daily Workflow

### Morning - Start working

```powershell
cd C:\Users\justin.harvey\excise-tax-portal\infrastructure\docker
.\manage.ps1 start
```

### During development

```powershell
# Check service status
.\manage.ps1 status

# View logs if needed
.\manage.ps1 logs -Follow

# Restart specific service if needed
.\manage.ps1 restart redis
```

### Evening - End of day

```powershell
# Stop services (keeps data for tomorrow)
.\manage.ps1 stop
```

## Getting Help

### Check service health

```powershell
.\manage.ps1 health
```

### View help

```powershell
.\manage.ps1 help
```

### Check Docker resources

```powershell
docker stats
docker system df
```

## Service Credentials (Development Only)

| Service | URL | Username | Password |
|---------|-----|----------|----------|
| PostgreSQL | localhost:5432 | postgres | postgres |
| Redis | localhost:6379 | - | redis |
| RabbitMQ | http://localhost:15672 | admin | admin |
| Grafana | http://localhost:3001 | admin | admin |

**Important**: Change these credentials before deploying to production!

## Success Indicators

You'll know everything is working when:

1. All services show "healthy" status: `.\manage.ps1 status`
2. You can access all web UIs (RabbitMQ, Prometheus, Grafana)
3. Your application can connect to PostgreSQL and Redis
4. No error messages in logs: `.\manage.ps1 logs`

## Need More Help?

- Read the [full README](README.md)
- Check Docker logs: `docker-compose logs [service]`
- Inspect container: `docker inspect excise-tax-postgres`
- Join the team chat for support

---

**Congratulations!** You now have a fully functional local development environment for the Excise Tax Portal.

Happy coding!
