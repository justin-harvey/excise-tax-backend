# Quick Docker Status Checker
# Run this to verify Docker is ready

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "   Docker Desktop Status Checker       " -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if Docker Desktop process is running
Write-Host "Checking if Docker Desktop is running..." -ForegroundColor Yellow
$dockerProcess = Get-Process -Name "Docker Desktop" -ErrorAction SilentlyContinue

if ($dockerProcess) {
    Write-Host "✓ Docker Desktop process found (PID: $($dockerProcess.Id))" -ForegroundColor Green
} else {
    Write-Host "✗ Docker Desktop is NOT running" -ForegroundColor Red
    Write-Host ""
    Write-Host "ACTION REQUIRED:" -ForegroundColor Yellow
    Write-Host "1. Press Windows Key" -ForegroundColor White
    Write-Host "2. Type 'Docker Desktop'" -ForegroundColor White
    Write-Host "3. Click to launch it" -ForegroundColor White
    Write-Host "4. Wait for the whale icon to be steady in system tray" -ForegroundColor White
    Write-Host "5. Run this script again" -ForegroundColor White
    Write-Host ""
    exit 1
}

Write-Host ""

# Check if docker command is available
Write-Host "Checking if docker command is available..." -ForegroundColor Yellow
try {
    $dockerVersion = docker --version 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Docker CLI found: $dockerVersion" -ForegroundColor Green
    } else {
        throw "Docker command failed"
    }
} catch {
    Write-Host "✗ Docker command not found in PATH" -ForegroundColor Red
    Write-Host ""
    Write-Host "SOLUTION:" -ForegroundColor Yellow
    Write-Host "Close this window and open a NEW PowerShell as Administrator" -ForegroundColor White
    Write-Host "Or run: `$env:Path = [System.Environment]::GetEnvironmentVariable('Path','Machine') + ';' + [System.Environment]::GetEnvironmentVariable('Path','User')" -ForegroundColor Gray
    Write-Host ""
    exit 1
}

Write-Host ""

# Check if docker compose is available
Write-Host "Checking if docker compose is available..." -ForegroundColor Yellow
try {
    $composeVersion = docker compose version 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Docker Compose found: $composeVersion" -ForegroundColor Green
    } else {
        throw "Docker compose failed"
    }
} catch {
    Write-Host "✗ Docker Compose not available" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Try to connect to Docker daemon
Write-Host "Checking if Docker Engine is ready..." -ForegroundColor Yellow
try {
    $dockerInfo = docker info 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✓ Docker Engine is ready" -ForegroundColor Green
    } else {
        Write-Host "⚠ Docker Engine is starting..." -ForegroundColor Yellow
        Write-Host "  Wait 30 seconds and try again" -ForegroundColor Gray
        exit 1
    }
} catch {
    Write-Host "⚠ Docker Engine is starting..." -ForegroundColor Yellow
    Write-Host "  Wait 30 seconds and try again" -ForegroundColor Gray
    exit 1
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "✓ Docker is READY!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "You can now run:" -ForegroundColor Yellow
Write-Host "  .\start-services.ps1     - Start all Docker services" -ForegroundColor White
Write-Host "  .\migrate-database.ps1   - Create database schema" -ForegroundColor White
Write-Host "  .\test-all-services.ps1  - Test everything" -ForegroundColor White
Write-Host ""
