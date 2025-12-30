# Excise Tax Portal - Start All Services
# This script starts Docker infrastructure and tests connectivity

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Excise Tax Portal - Service Startup  " -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if Docker is running
Write-Host "Step 1: Checking Docker Desktop..." -ForegroundColor Yellow
try {
    $dockerVersion = docker --version
    Write-Host "✓ Docker found: $dockerVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ Docker not found in PATH" -ForegroundColor Red
    Write-Host ""
    Write-Host "SOLUTION:" -ForegroundColor Yellow
    Write-Host "1. Make sure Docker Desktop is running (check system tray)" -ForegroundColor White
    Write-Host "2. Close and reopen PowerShell/Terminal" -ForegroundColor White
    Write-Host "3. Or restart your computer to refresh PATH" -ForegroundColor White
    Write-Host ""
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Host ""

# Check Docker Compose
Write-Host "Step 2: Checking Docker Compose..." -ForegroundColor Yellow
try {
    $composeVersion = docker compose version
    Write-Host "✓ Docker Compose found: $composeVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ Docker Compose not available" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Navigate to Docker directory
Write-Host "Step 3: Starting Docker services..." -ForegroundColor Yellow
Set-Location "C:\Users\justin.harvey\excise-tax-portal\infrastructure\docker"

# Start services
Write-Host "  - Starting PostgreSQL, Redis, RabbitMQ, Prometheus, Grafana..." -ForegroundColor White
docker compose up -d

if ($LASTEXITCODE -ne 0) {
    Write-Host "✗ Failed to start services" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Wait for services to initialize
Write-Host "Step 4: Waiting for services to initialize (30 seconds)..." -ForegroundColor Yellow
Start-Sleep -Seconds 30

Write-Host ""

# Check service status
Write-Host "Step 5: Checking service status..." -ForegroundColor Yellow
docker compose ps

Write-Host ""

# Test PostgreSQL
Write-Host "Step 6: Testing PostgreSQL..." -ForegroundColor Yellow
try {
    $pgTest = docker exec excise-tax-postgres pg_isready -U postgres
    if ($pgTest -like "*accepting connections*") {
        Write-Host "✓ PostgreSQL is ready" -ForegroundColor Green
    } else {
        Write-Host "⚠ PostgreSQL may not be ready yet" -ForegroundColor Yellow
    }
} catch {
    Write-Host "✗ PostgreSQL check failed" -ForegroundColor Red
}

Write-Host ""

# Test Redis
Write-Host "Step 7: Testing Redis..." -ForegroundColor Yellow
try {
    $redisTest = docker exec excise-tax-redis redis-cli -a redis ping
    if ($redisTest -eq "PONG") {
        Write-Host "✓ Redis is ready" -ForegroundColor Green
    } else {
        Write-Host "⚠ Redis may not be ready yet" -ForegroundColor Yellow
    }
} catch {
    Write-Host "✗ Redis check failed" -ForegroundColor Red
}

Write-Host ""

# Test RabbitMQ
Write-Host "Step 8: Testing RabbitMQ..." -ForegroundColor Yellow
try {
    $rmqTest = docker exec excise-tax-rabbitmq rabbitmq-diagnostics -q ping
    if ($rmqTest -eq "Ping succeeded") {
        Write-Host "✓ RabbitMQ is ready" -ForegroundColor Green
    } else {
        Write-Host "⚠ RabbitMQ may not be ready yet" -ForegroundColor Yellow
    }
} catch {
    Write-Host "✗ RabbitMQ check failed (this is normal if still starting)" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "         Service URLs                   " -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "PostgreSQL:     localhost:5432" -ForegroundColor White
Write-Host "  Database:     excise_tax_db" -ForegroundColor Gray
Write-Host "  Username:     postgres" -ForegroundColor Gray
Write-Host "  Password:     postgres" -ForegroundColor Gray
Write-Host ""
Write-Host "Redis:          localhost:6379" -ForegroundColor White
Write-Host "  Password:     redis" -ForegroundColor Gray
Write-Host ""
Write-Host "RabbitMQ:       localhost:5672" -ForegroundColor White
Write-Host "  Management:   http://localhost:15672" -ForegroundColor Cyan
Write-Host "  Username:     admin" -ForegroundColor Gray
Write-Host "  Password:     admin" -ForegroundColor Gray
Write-Host ""
Write-Host "Prometheus:     http://localhost:9090" -ForegroundColor Cyan
Write-Host "Grafana:        http://localhost:3001" -ForegroundColor Cyan
Write-Host "  Username:     admin" -ForegroundColor Gray
Write-Host "  Password:     admin" -ForegroundColor Gray
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "✓ Infrastructure started successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Run database migrations: .\migrate-database.ps1" -ForegroundColor White
Write-Host "2. Build Go services: cd ..\backend; make build" -ForegroundColor White
Write-Host "3. Run services: make run-all" -ForegroundColor White
Write-Host ""
Write-Host "To view logs: docker compose logs -f" -ForegroundColor Gray
Write-Host "To stop services: docker compose down" -ForegroundColor Gray
Write-Host ""
