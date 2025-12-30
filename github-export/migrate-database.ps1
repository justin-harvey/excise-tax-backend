# Excise Tax Portal - Database Migration Script
# This script runs all database migrations

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "    Database Migration Runner          " -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if migrations directory exists
$migrationsPath = "C:\Users\justin.harvey\excise-tax-portal\backend\migrations"
if (-not (Test-Path $migrationsPath)) {
    Write-Host "✗ Migrations directory not found: $migrationsPath" -ForegroundColor Red
    exit 1
}

# Database connection string
$dbHost = "localhost"
$dbPort = "5432"
$dbUser = "postgres"
$dbPassword = "postgres"
$dbName = "excise_tax_db"
$dbUrl = "postgresql://${dbUser}:${dbPassword}@${dbHost}:${dbPort}/${dbName}?sslmode=disable"

Write-Host "Checking database connectivity..." -ForegroundColor Yellow
try {
    $pgTest = docker exec excise-tax-postgres psql -U postgres -d excise_tax_db -c "SELECT 1"
    Write-Host "✓ Database is accessible" -ForegroundColor Green
} catch {
    Write-Host "✗ Cannot connect to database" -ForegroundColor Red
    Write-Host "  Make sure Docker services are running: .\start-services.ps1" -ForegroundColor Yellow
    exit 1
}

Write-Host ""

# Check if golang-migrate is installed
Write-Host "Checking for golang-migrate..." -ForegroundColor Yellow
$migrateCmd = Get-Command migrate -ErrorAction SilentlyContinue

if (-not $migrateCmd) {
    Write-Host "✗ golang-migrate not found" -ForegroundColor Red
    Write-Host ""
    Write-Host "Installing golang-migrate via Chocolatey..." -ForegroundColor Yellow
    Write-Host ""

    # Check if chocolatey is installed
    $chocoCmd = Get-Command choco -ErrorAction SilentlyContinue
    if (-not $chocoCmd) {
        Write-Host "Chocolatey not found. Installing migrations manually via Docker..." -ForegroundColor Yellow
        Write-Host ""

        # Run migrations using Docker
        Write-Host "Running migrations via Docker container..." -ForegroundColor Yellow

        # Run each migration file manually via psql
        $migrationFiles = Get-ChildItem -Path $migrationsPath -Filter "*.up.sql" | Sort-Object Name

        foreach ($file in $migrationFiles) {
            Write-Host "  Applying: $($file.Name)" -ForegroundColor White
            $content = Get-Content $file.FullName -Raw

            # Execute SQL via docker exec
            $content | docker exec -i excise-tax-postgres psql -U postgres -d excise_tax_db

            if ($LASTEXITCODE -eq 0) {
                Write-Host "  ✓ Applied successfully" -ForegroundColor Green
            } else {
                Write-Host "  ✗ Failed to apply migration" -ForegroundColor Red
            }
        }

    } else {
        Write-Host "Installing golang-migrate..." -ForegroundColor Yellow
        choco install golang-migrate -y

        if ($LASTEXITCODE -ne 0) {
            Write-Host "✗ Failed to install golang-migrate" -ForegroundColor Red
            exit 1
        }

        # Refresh PATH
        $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")

        Write-Host "✓ golang-migrate installed" -ForegroundColor Green
        Write-Host ""
        Write-Host "Running migrations..." -ForegroundColor Yellow
        migrate -path $migrationsPath -database $dbUrl up
    }
} else {
    Write-Host "✓ golang-migrate found" -ForegroundColor Green
    Write-Host ""
    Write-Host "Running migrations..." -ForegroundColor Yellow
    Write-Host "  Path: $migrationsPath" -ForegroundColor Gray
    Write-Host "  Database: $dbName" -ForegroundColor Gray
    Write-Host ""

    migrate -path $migrationsPath -database $dbUrl up

    if ($LASTEXITCODE -eq 0) {
        Write-Host ""
        Write-Host "✓ Migrations completed successfully!" -ForegroundColor Green
    } else {
        Write-Host ""
        Write-Host "✗ Migration failed" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""

# Verify tables were created
Write-Host "Verifying database schema..." -ForegroundColor Yellow
$tables = docker exec excise-tax-postgres psql -U postgres -d excise_tax_db -c "\dt" -A -t

if ($tables) {
    Write-Host "✓ Database tables created:" -ForegroundColor Green
    docker exec excise-tax-postgres psql -U postgres -d excise_tax_db -c "\dt"
} else {
    Write-Host "⚠ No tables found - migrations may not have run correctly" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "✓ Database setup complete!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Build Go services: cd backend; make build" -ForegroundColor White
Write-Host "2. Run services: make run-all" -ForegroundColor White
Write-Host ""
