# =====================================================
# Database Migration Script (PowerShell)
# Uses golang-migrate to manage PostgreSQL migrations
# =====================================================

param(
    [Parameter(Mandatory=$true, Position=0)]
    [ValidateSet('up', 'down', 'drop', 'force', 'version', 'create')]
    [string]$Command,

    [Parameter(Position=1)]
    [string]$Argument
)

# Colors for output
function Write-Success { Write-Host $args -ForegroundColor Green }
function Write-Error { Write-Host $args -ForegroundColor Red }
function Write-Warning { Write-Host $args -ForegroundColor Yellow }

# Configuration
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir
$MigrationsDir = Join-Path $ProjectRoot "migrations"

# Default database connection (can be overridden by environment variables)
$DbHost = if ($env:DB_HOST) { $env:DB_HOST } else { "localhost" }
$DbPort = if ($env:DB_PORT) { $env:DB_PORT } else { "5432" }
$DbUser = if ($env:DB_USER) { $env:DB_USER } else { "postgres" }
$DbPassword = if ($env:DB_PASSWORD) { $env:DB_PASSWORD } else { "postgres" }
$DbName = if ($env:DB_NAME) { $env:DB_NAME } else { "excise_tax_portal" }

# Construct database URL
$DatabaseUrl = if ($env:DATABASE_URL) {
    $env:DATABASE_URL
} else {
    "postgresql://${DbUser}:${DbPassword}@${DbHost}:${DbPort}/${DbName}?sslmode=disable"
}

# Check if golang-migrate is installed
if (-not (Get-Command migrate -ErrorAction SilentlyContinue)) {
    Write-Error "Error: golang-migrate is not installed."
    Write-Host ""
    Write-Host "Please install it using one of the following methods:"
    Write-Host ""
    Write-Host "Windows (Chocolatey):"
    Write-Host "  choco install migrate"
    Write-Host ""
    Write-Host "Windows (Scoop):"
    Write-Host "  scoop install migrate"
    Write-Host ""
    Write-Host "Go install:"
    Write-Host "  go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    Write-Host ""
    Write-Host "Manual download:"
    Write-Host "  Download from: https://github.com/golang-migrate/migrate/releases"
    Write-Host "  Extract and add to PATH"
    exit 1
}

# Check if migrations directory exists
if (-not (Test-Path $MigrationsDir)) {
    Write-Error "Error: Migrations directory not found at $MigrationsDir"
    exit 1
}

# Main command handler
switch ($Command) {
    'up' {
        Write-Success "Applying migrations..."
        if ($Argument) {
            migrate -path $MigrationsDir -database $DatabaseUrl up $Argument
        } else {
            migrate -path $MigrationsDir -database $DatabaseUrl up
        }
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Migrations applied successfully!"
        }
    }

    'down' {
        Write-Warning "Rolling back migrations..."
        if ($Argument) {
            migrate -path $MigrationsDir -database $DatabaseUrl down $Argument
        } else {
            Write-Error "Warning: This will rollback ALL migrations!"
            $confirm = Read-Host "Are you sure? (yes/no)"
            if ($confirm -eq "yes") {
                migrate -path $MigrationsDir -database $DatabaseUrl down
            } else {
                Write-Host "Aborted."
                exit 0
            }
        }
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Rollback completed!"
        }
    }

    'drop' {
        Write-Error "WARNING: This will drop all tables and data!"
        $confirm = Read-Host "Are you absolutely sure? Type 'DROP ALL DATA' to confirm"
        if ($confirm -eq "DROP ALL DATA") {
            migrate -path $MigrationsDir -database $DatabaseUrl drop
            if ($LASTEXITCODE -eq 0) {
                Write-Success "Database dropped successfully!"
            }
        } else {
            Write-Host "Aborted."
            exit 0
        }
    }

    'force' {
        if (-not $Argument) {
            Write-Error "Error: Version number required"
            Write-Host "Usage: .\migrate.ps1 force VERSION"
            exit 1
        }
        Write-Warning "Forcing version to $Argument..."
        migrate -path $MigrationsDir -database $DatabaseUrl force $Argument
        if ($LASTEXITCODE -eq 0) {
            Write-Success "Version forced successfully!"
        }
    }

    'version' {
        Write-Success "Current migration version:"
        migrate -path $MigrationsDir -database $DatabaseUrl version
    }

    'create' {
        if (-not $Argument) {
            Write-Error "Error: Migration name required"
            Write-Host "Usage: .\migrate.ps1 create NAME"
            exit 1
        }

        # Get the next migration number
        $LastMigration = Get-ChildItem -Path $MigrationsDir -Filter "*.up.sql" |
            ForEach-Object { $_.Name -replace '_.*' } |
            Sort-Object -Descending |
            Select-Object -First 1

        if (-not $LastMigration) {
            $NextNumber = "000001"
        } else {
            $NextNumber = ([int]$LastMigration + 1).ToString("000000")
        }

        $MigrationName = $Argument
        $UpFile = Join-Path $MigrationsDir "${NextNumber}_${MigrationName}.up.sql"
        $DownFile = Join-Path $MigrationsDir "${NextNumber}_${MigrationName}.down.sql"

        # Create up migration
        @"
-- =====================================================
-- Migration: ${MigrationName}
-- =====================================================

-- Add your migration SQL here

-- =====================================================
-- Migration Complete
-- =====================================================
"@ | Out-File -FilePath $UpFile -Encoding UTF8

        # Create down migration
        @"
-- =====================================================
-- Rollback: ${MigrationName}
-- =====================================================

-- Add your rollback SQL here

-- =====================================================
-- Rollback Complete
-- =====================================================
"@ | Out-File -FilePath $DownFile -Encoding UTF8

        Write-Success "Created migration files:"
        Write-Host "  - $UpFile"
        Write-Host "  - $DownFile"
    }
}
