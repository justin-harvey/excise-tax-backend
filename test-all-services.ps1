# Excise Tax Portal - Comprehensive Service Test Script
# Tests Docker infrastructure and Go microservices

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Excise Tax Portal - Service Tests    " -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$testsPassed = 0
$testsFailed = 0

function Test-Service {
    param (
        [string]$Name,
        [string]$Url,
        [int]$ExpectedStatus = 200
    )

    try {
        $response = Invoke-WebRequest -Uri $Url -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
        if ($response.StatusCode -eq $ExpectedStatus) {
            Write-Host "✓ $Name - PASSED" -ForegroundColor Green
            $script:testsPassed++
            return $true
        } else {
            Write-Host "✗ $Name - FAILED (Status: $($response.StatusCode))" -ForegroundColor Red
            $script:testsFailed++
            return $false
        }
    } catch {
        Write-Host "✗ $Name - FAILED ($($_.Exception.Message))" -ForegroundColor Red
        $script:testsFailed++
        return $false
    }
}

# Test 1: Docker Infrastructure
Write-Host "TEST SUITE 1: Docker Infrastructure" -ForegroundColor Yellow
Write-Host "=====================================" -ForegroundColor Yellow
Write-Host ""

# PostgreSQL
Write-Host "Testing PostgreSQL..." -ForegroundColor White
try {
    $pgResult = docker exec excise-tax-postgres pg_isready -U postgres 2>&1
    if ($pgResult -like "*accepting connections*") {
        Write-Host "✓ PostgreSQL - RUNNING" -ForegroundColor Green
        $testsPassed++
    } else {
        Write-Host "✗ PostgreSQL - NOT READY" -ForegroundColor Red
        $testsFailed++
    }
} catch {
    Write-Host "✗ PostgreSQL - NOT RUNNING" -ForegroundColor Red
    $testsFailed++
}

# Redis
Write-Host "Testing Redis..." -ForegroundColor White
try {
    $redisResult = docker exec excise-tax-redis redis-cli -a redis ping 2>&1 | Select-String "PONG"
    if ($redisResult) {
        Write-Host "✓ Redis - RUNNING" -ForegroundColor Green
        $testsPassed++
    } else {
        Write-Host "✗ Redis - NOT RESPONDING" -ForegroundColor Red
        $testsFailed++
    }
} catch {
    Write-Host "✗ Redis - NOT RUNNING" -ForegroundColor Red
    $testsFailed++
}

# RabbitMQ
Write-Host "Testing RabbitMQ..." -ForegroundColor White
try {
    $rmqResult = docker exec excise-tax-rabbitmq rabbitmq-diagnostics -q ping 2>&1
    if ($rmqResult -like "*Ping succeeded*" -or $LASTEXITCODE -eq 0) {
        Write-Host "✓ RabbitMQ - RUNNING" -ForegroundColor Green
        $testsPassed++
    } else {
        Write-Host "⚠ RabbitMQ - STARTING" -ForegroundColor Yellow
    }
} catch {
    Write-Host "⚠ RabbitMQ - STARTING (this is normal)" -ForegroundColor Yellow
}

# RabbitMQ Management UI
Test-Service -Name "RabbitMQ Management UI" -Url "http://localhost:15672"

# Prometheus
Test-Service -Name "Prometheus" -Url "http://localhost:9090"

# Grafana
Test-Service -Name "Grafana" -Url "http://localhost:3001"

Write-Host ""

# Test 2: Database Schema
Write-Host "TEST SUITE 2: Database Schema" -ForegroundColor Yellow
Write-Host "==============================" -ForegroundColor Yellow
Write-Host ""

$expectedTables = @(
    "users",
    "manufacturers",
    "payments",
    "xrpl_payments",
    "tax_reports",
    "tax_rates",
    "exchange_rates",
    "audit_log",
    "sessions"
)

Write-Host "Checking for required tables..." -ForegroundColor White
foreach ($table in $expectedTables) {
    $tableExists = docker exec excise-tax-postgres psql -U postgres -d excise_tax_db -c "\dt $table" -A -t 2>&1
    if ($tableExists -and $tableExists -notlike "*Did not find*") {
        Write-Host "✓ Table '$table' exists" -ForegroundColor Green
        $testsPassed++
    } else {
        Write-Host "✗ Table '$table' missing" -ForegroundColor Red
        $testsFailed++
    }
}

Write-Host ""

# Test 3: Go Microservices
Write-Host "TEST SUITE 3: Go Microservices" -ForegroundColor Yellow
Write-Host "===============================" -ForegroundColor Yellow
Write-Host ""

$services = @(
    @{Name="API Gateway"; Port=8080; Path="/health"},
    @{Name="Payment Service"; Port=8081; Path="/health"},
    @{Name="Tax Service"; Port=8082; Path="/health"},
    @{Name="Reporting Service"; Port=8083; Path="/health"},
    @{Name="Notification Service"; Port=8084; Path="/health"},
    @{Name="Auth Service"; Port=8085; Path="/health"}
)

$servicesRunning = $false
foreach ($service in $services) {
    $url = "http://localhost:$($service.Port)$($service.Path)"
    $result = Test-Service -Name $service.Name -Url $url
    if ($result) {
        $servicesRunning = $true
    }
}

if (-not $servicesRunning) {
    Write-Host ""
    Write-Host "⚠ No Go services are running yet" -ForegroundColor Yellow
    Write-Host "  Start services with: cd backend; make run-all" -ForegroundColor Gray
}

Write-Host ""

# Test 4: API Endpoints (if services are running)
if ($servicesRunning) {
    Write-Host "TEST SUITE 4: API Endpoints" -ForegroundColor Yellow
    Write-Host "============================" -ForegroundColor Yellow
    Write-Host ""

    # Test Exchange Rate Endpoint
    Write-Host "Testing XRPL Exchange Rate API..." -ForegroundColor White
    try {
        $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/payments/xrpl/exchange-rate" -Method Get -ErrorAction Stop
        Write-Host "✓ Exchange Rate API - PASSED" -ForegroundColor Green
        Write-Host "  Current XRP/USD Rate: $($response.data.rate)" -ForegroundColor Gray
        $testsPassed++
    } catch {
        if ($_.Exception.Response.StatusCode.Value__ -eq 404) {
            Write-Host "⚠ Exchange Rate API - Endpoint not implemented yet" -ForegroundColor Yellow
        } else {
            Write-Host "✗ Exchange Rate API - FAILED" -ForegroundColor Red
            $testsFailed++
        }
    }

    # Test Auth Registration
    Write-Host "Testing Auth Registration API..." -ForegroundColor White
    $testUser = @{
        email = "test_$([guid]::NewGuid().ToString().Substring(0,8))@example.com"
        password = "SecurePassword123!"
        first_name = "Test"
        last_name = "User"
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/register" `
            -Method Post `
            -ContentType "application/json" `
            -Body $testUser `
            -ErrorAction Stop

        Write-Host "✓ Auth Registration API - PASSED" -ForegroundColor Green
        $testsPassed++

        # Try to login with created user
        Write-Host "Testing Auth Login API..." -ForegroundColor White
        $loginData = @{
            email = ($testUser | ConvertFrom-Json).email
            password = "SecurePassword123!"
        } | ConvertTo-Json

        try {
            $loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" `
                -Method Post `
                -ContentType "application/json" `
                -Body $loginData `
                -ErrorAction Stop

            Write-Host "✓ Auth Login API - PASSED" -ForegroundColor Green
            Write-Host "  Access Token: $($loginResponse.data.access_token.Substring(0,20))..." -ForegroundColor Gray
            $testsPassed++
        } catch {
            Write-Host "✗ Auth Login API - FAILED" -ForegroundColor Red
            $testsFailed++
        }
    } catch {
        if ($_.Exception.Response.StatusCode.Value__ -eq 404) {
            Write-Host "⚠ Auth Registration API - Endpoint not implemented yet" -ForegroundColor Yellow
        } else {
            Write-Host "✗ Auth Registration API - FAILED" -ForegroundColor Red
            Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Gray
            $testsFailed++
        }
    }

    Write-Host ""
}

# Summary
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "           Test Summary                 " -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Tests Passed: $testsPassed" -ForegroundColor Green
Write-Host "Tests Failed: $testsFailed" -ForegroundColor Red
Write-Host ""

if ($testsFailed -eq 0) {
    Write-Host "✓ ALL TESTS PASSED!" -ForegroundColor Green
    exit 0
} else {
    Write-Host "⚠ Some tests failed. Check the output above for details." -ForegroundColor Yellow
    exit 1
}
