# Quick Start Guide

## Installation

```bash
cd backend/pkg
go mod download
```

## Running Tests

```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Specific package
go test ./database -v
```

## Common Usage Patterns

### 1. Initialize Core Services

```go
package main

import (
    "context"
    "log"

    "github.com/excise-tax-portal/backend/pkg/cache"
    "github.com/excise-tax-portal/backend/pkg/config"
    "github.com/excise-tax-portal/backend/pkg/database"
    "github.com/excise-tax-portal/backend/pkg/logger"
    "github.com/excise-tax-portal/backend/pkg/utils"
    "github.com/excise-tax-portal/backend/pkg/validator"
    "go.uber.org/zap"
)

func main() {
    ctx := context.Background()

    // 1. Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    // 2. Initialize logger
    logger, err := logger.NewLogger(cfg.Logger)
    if err != nil {
        log.Fatal(err)
    }
    defer logger.Sync()

    // 3. Initialize database
    db, err := database.NewPostgresDB(ctx, &cfg.Database)
    if err != nil {
        logger.Fatal("Database connection failed", zap.Error(err))
    }
    defer db.Close()

    // 4. Initialize Redis cache
    cache, err := cache.NewRedisClient(ctx, &cfg.Redis)
    if err != nil {
        logger.Fatal("Redis connection failed", zap.Error(err))
    }
    defer cache.Close()

    // 5. Initialize JWT manager
    jwtManager, err := utils.NewJWTManager(&utils.JWTConfig{
        SecretKey:            cfg.JWT.SecretKey,
        AccessTokenDuration:  cfg.JWT.AccessTokenDuration,
        RefreshTokenDuration: cfg.JWT.RefreshTokenDuration,
        Issuer:               cfg.JWT.Issuer,
    })
    if err != nil {
        logger.Fatal("JWT manager initialization failed", zap.Error(err))
    }

    // 6. Initialize validator
    validator := validator.New()

    logger.Info("All services initialized successfully")
}
```

### 2. Database Operations

```go
// Simple query
var count int
err := db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)

// Query multiple rows
rows, err := db.Query(ctx, "SELECT id, email FROM users WHERE active = $1", true)
defer rows.Close()

for rows.Next() {
    var id int64
    var email string
    if err := rows.Scan(&id, &email); err != nil {
        return err
    }
    // Process row
}

// Execute insert/update/delete
result, err := db.Exec(ctx,
    "INSERT INTO users (email, password) VALUES ($1, $2)",
    "user@example.com", hashedPassword,
)

// Transaction
err = db.WithTransaction(ctx, func(tx pgx.Tx) error {
    // Multiple operations
    _, err := tx.Exec(ctx, "INSERT INTO users ...")
    if err != nil {
        return err
    }

    _, err = tx.Exec(ctx, "INSERT INTO audit_log ...")
    return err
})
```

### 3. Redis Caching

```go
// Cache user data
type User struct {
    ID    int64  `json:"id"`
    Email string `json:"email"`
}

user := User{ID: 123, Email: "user@example.com"}

// Set with 1 hour TTL
err := cache.Set(ctx, "user:123", user, 1*time.Hour)

// Get
var cachedUser User
err = cache.Get(ctx, "user:123", &cachedUser)
if err != nil {
    // Cache miss or error
}

// Simple string operations
err = cache.SetString(ctx, "token:abc", "valid", 15*time.Minute)
token, err := cache.GetString(ctx, "token:abc")

// Counter operations
count, err := cache.Increment(ctx, "api:requests:user:123")
```

### 4. Logging

```go
// Basic logging
logger.Info("Server started")
logger.Error("Database error", zap.Error(err))

// With fields
logger.Info("User created",
    zap.Int64("user_id", 123),
    zap.String("email", "user@example.com"),
)

// Context-aware logging (for request tracing)
ctx := logger.WithRequestID(context.Background(), "req-abc-123")
ctx = logger.WithUserID(ctx, "user-456")

logger.InfoContext(ctx, "Processing payment",
    zap.Float64("amount", 100.50),
)
// Output: {"level":"info","request_id":"req-abc-123","user_id":"user-456","msg":"Processing payment","amount":100.5}
```

### 5. Error Handling

```go
import "github.com/excise-tax-portal/backend/pkg/errors"

// Create errors
err := errors.NewNotFoundError("user", "user not found with ID: 123")
err := errors.NewValidationError("invalid input", map[string]string{
    "email": "must be a valid email address",
})
err := errors.NewUnauthorizedError("invalid credentials")

// Check error types
if errors.IsNotFound(err) {
    // Handle not found
}

// In HTTP handler
if appErr, ok := err.(*errors.AppError); ok {
    w.WriteHeader(appErr.HTTPStatus)
    json.NewEncoder(w).Encode(errors.NewErrorResponse(appErr))
}
```

### 6. Validation

```go
import "github.com/excise-tax-portal/backend/pkg/validator"

type CreateUserRequest struct {
    Email       string `json:"email" validate:"required,email"`
    Password    string `json:"password" validate:"required,strong_password"`
    Phone       string `json:"phone" validate:"omitempty,phone"`
    CompanyFEIN string `json:"company_fein" validate:"required,fein"`
    State       string `json:"state" validate:"required,state_code"`
    ZipCode     string `json:"zip_code" validate:"required,zip"`
}

v := validator.New()
req := CreateUserRequest{...}

if err := v.Validate(req); err != nil {
    fieldErrors := v.FormatErrors(err)
    // Return to client
    // {"email": "Email is required", "password": "Password must contain..."}
}
```

### 7. JWT Authentication

```go
import "github.com/excise-tax-portal/backend/pkg/utils"

// Generate token
claims := &utils.Claims{
    UserID:      123,
    Email:       "user@example.com",
    Roles:       []string{"user", "taxpayer"},
    Permissions: []string{"read:payments", "write:payments"},
    IsAdmin:     false,
    StateID:     "CA",
}

accessToken, err := jwtManager.GenerateAccessToken(claims)
refreshToken, err := jwtManager.GenerateRefreshToken(claims)

// Validate token
claims, err := jwtManager.ValidateAccessToken(tokenString)
if err != nil {
    // Invalid or expired token
}

// Check permissions
if claims.HasPermission("write:payments") {
    // Allow action
}

if claims.HasAnyRole("admin", "state_admin") {
    // Admin action
}
```

### 8. Password Management

```go
import "github.com/excise-tax-portal/backend/pkg/utils"

// Hash password on registration
hashedPassword, err := utils.HashPassword("userPassword123!")

// Verify password on login
err = utils.ComparePasswords(hashedPassword, "userPassword123!")
if err != nil {
    // Invalid password
}

// Check password strength
strength := utils.CheckPasswordStrength("MyP@ssw0rd123")
// Returns: PasswordWeak, PasswordModerate, PasswordStrong, PasswordVeryStrong

// Generate secure tokens
resetToken, err := utils.GenerateResetToken()
sessionID, err := utils.GenerateSessionID()
randomString, err := utils.GenerateRandomString(32)
```

## Configuration

### Environment Variables

```bash
# Override config file values with environment variables
export APP_SERVER_PORT=9000
export APP_DATABASE_HOST=db.example.com
export APP_DATABASE_PASSWORD=secret
export APP_JWT_SECRET_KEY=your-super-secret-key-at-least-32-chars
export APP_REDIS_HOST=redis.example.com
```

### Config File (config.yaml)

```yaml
server:
  port: 8080
  environment: production

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: excise_tax_portal

jwt:
  secret_key: "your-secret-key-at-least-32-characters-long"
  access_token_duration: 15m
  refresh_token_duration: 168h
```

## Common Patterns

### Middleware Example

```go
func AuthMiddleware(jwtManager *utils.JWTManager, logger *logger.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract token
            authHeader := r.Header.Get("Authorization")
            token, err := utils.ExtractTokenFromBearer(authHeader)
            if err != nil {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            // Validate token
            claims, err := jwtManager.ValidateAccessToken(token)
            if err != nil {
                logger.Warn("Invalid token", zap.Error(err))
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }

            // Add to context
            ctx := context.WithValue(r.Context(), "claims", claims)
            ctx = logger.WithUserID(ctx, fmt.Sprintf("%d", claims.UserID))

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Request Logging Middleware

```go
func LoggingMiddleware(logger *logger.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Generate request ID
            requestID, _ := utils.GenerateRandomString(16)
            ctx := logger.WithRequestID(r.Context(), requestID)

            // Log request
            logger.InfoContext(ctx, "Incoming request",
                zap.String("method", r.Method),
                zap.String("path", r.URL.Path),
            )

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### Health Check Handler

```go
func HealthCheckHandler(db *database.PostgresDB, cache *cache.RedisClient) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        // Check database
        if err := db.HealthCheck(ctx); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            json.NewEncoder(w).Encode(map[string]string{
                "status": "unhealthy",
                "database": "down",
            })
            return
        }

        // Check Redis
        if err := cache.HealthCheck(ctx); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            json.NewEncoder(w).Encode(map[string]string{
                "status": "unhealthy",
                "cache": "down",
            })
            return
        }

        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "status": "healthy",
        })
    }
}
```

## Troubleshooting

### Common Issues

1. **Database connection fails:**
   - Check `DATABASE_HOST` and `DATABASE_PASSWORD` environment variables
   - Verify PostgreSQL is running
   - Check firewall rules

2. **Redis connection fails:**
   - Verify Redis is running
   - Check `REDIS_HOST` configuration
   - Ensure Redis is accepting connections

3. **JWT validation fails:**
   - Ensure `JWT_SECRET_KEY` is consistent across services
   - Check token expiration
   - Verify token format (Bearer prefix)

4. **Configuration not loading:**
   - Ensure `config.yaml` is in the correct path
   - Check environment variable names (must be prefixed with `APP_`)
   - Verify YAML syntax

## Next Steps

1. Copy `config.example.yaml` to `config.yaml`
2. Update with your environment values
3. Install dependencies: `go mod download`
4. Run tests: `go test ./...`
5. Import packages in your service
6. Follow the patterns above
