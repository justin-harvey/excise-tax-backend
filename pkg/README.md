# Shared Packages

This directory contains shared packages used across all microservices in the excise tax portal backend.

## Packages

### config

Configuration management with support for YAML files and environment variables.

**Features:**
- YAML configuration files
- Environment variable overrides
- Validation
- Default values
- Environment-specific configs (dev, staging, prod)

**Example:**
```go
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Server running on %s\n", cfg.GetServerAddress())
```

### database

PostgreSQL connection management with connection pooling and transaction support.

**Features:**
- Connection pooling using pgxpool
- Context-aware queries
- Transaction wrapper functions
- Health checks
- Graceful shutdown

**Example:**
```go
db, err := database.NewPostgresDB(ctx, &database.Config{
    Host:   "localhost",
    Port:   5432,
    User:   "postgres",
    DBName: "excise_tax",
})
defer db.Close()

// Use transaction
err = db.WithTransaction(ctx, func(tx pgx.Tx) error {
    _, err := tx.Exec(ctx, "INSERT INTO users (email) VALUES ($1)", email)
    return err
})
```

### errors

Custom error types with HTTP status code mapping for consistent error handling.

**Features:**
- Pre-defined error types (NotFound, Validation, Unauthorized, etc.)
- HTTP status code mapping
- Error wrapping
- JSON serialization for API responses
- Type checking helpers

**Example:**
```go
err := errors.NewNotFoundError("user", "user not found with ID: 123")

if errors.IsNotFound(err) {
    // Handle not found error
}

// Get HTTP status
status := errors.GetHTTPStatus(err)
```

### logger

Structured logging with context-aware request tracking using zap.

**Features:**
- JSON formatting for production
- Console formatting for development
- Context-aware logging (request ID, user ID, trace ID)
- Multiple log levels
- High performance

**Example:**
```go
log, err := logger.NewLogger(logger.Config{
    Environment: "production",
    Level:       "info",
})
defer log.Sync()

// Add request ID to context
ctx := logger.WithRequestID(context.Background(), "req-123")

// Log with context
log.InfoContext(ctx, "Processing request", zap.String("user", "john"))
```

### validator

Request validation using go-playground/validator with custom validation rules.

**Features:**
- Struct validation using tags
- Custom validators (phone, FEIN, state code, ZIP, strong password)
- Error formatting for API responses
- Field-level validation

**Example:**
```go
type CreateUserRequest struct {
    Email    string `validate:"required,email"`
    Password string `validate:"required,strong_password"`
    Phone    string `validate:"omitempty,phone"`
}

v := validator.New()
if err := v.Validate(req); err != nil {
    errors := v.FormatErrors(err)
    // Return formatted errors to client
}
```

### utils

Utility functions for JWT token management and cryptography.

#### JWT (utils/jwt.go)

**Features:**
- Access and refresh token generation
- Token validation
- Claims extraction
- Role and permission checking
- Token expiration handling

**Example:**
```go
manager, err := utils.NewJWTManager(&utils.JWTConfig{
    SecretKey:            "your-secret-key",
    AccessTokenDuration:  15 * time.Minute,
    RefreshTokenDuration: 7 * 24 * time.Hour,
    Issuer:               "excise-tax-portal",
})

claims := &utils.Claims{
    UserID:      123,
    Email:       "user@example.com",
    Roles:       []string{"user"},
    Permissions: []string{"read", "write"},
    IsAdmin:     false,
}

token, err := manager.GenerateAccessToken(claims)
```

#### Crypto (utils/crypto.go)

**Features:**
- Password hashing with bcrypt
- Secure random string/token generation
- SHA256 hashing
- Password strength checking
- PKCE support for OAuth2

**Example:**
```go
// Hash password
hashed, err := utils.HashPassword("myPassword123")

// Compare passwords
err = utils.ComparePasswords(hashed, "myPassword123")

// Generate secure token
token, err := utils.GenerateSecureToken(32)

// Check password strength
strength := utils.CheckPasswordStrength("MyP@ssw0rd123")
```

## Testing

Run tests for all packages:

```bash
cd backend/pkg
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

Run tests for a specific package:

```bash
go test ./database
go test ./cache
go test ./logger
```

## Dependencies

All dependencies are managed in `go.mod`. Main dependencies include:

- **github.com/jackc/pgx/v5** - PostgreSQL driver and toolkit
- **github.com/redis/go-redis/v9** - Redis client
- **go.uber.org/zap** - Structured logging
- **github.com/spf13/viper** - Configuration management
- **github.com/go-playground/validator/v10** - Validation
- **github.com/golang-jwt/jwt/v5** - JWT tokens
- **golang.org/x/crypto** - Cryptography

## Usage in Microservices

To use these packages in your microservice:

1. Import the package:
```go
import "github.com/excise-tax-portal/backend/pkg/database"
```

2. Initialize in your service's main function:
```go
func main() {
    // Load config
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    // Initialize logger
    logger, err := logger.NewLogger(cfg.Logger)
    if err != nil {
        log.Fatal(err)
    }
    defer logger.Sync()

    // Initialize database
    db, err := database.NewPostgresDB(ctx, &cfg.Database)
    if err != nil {
        logger.Fatal("Failed to connect to database", zap.Error(err))
    }
    defer db.Close()

    // Initialize Redis
    cache, err := cache.NewRedisClient(ctx, &cfg.Redis)
    if err != nil {
        logger.Fatal("Failed to connect to Redis", zap.Error(err))
    }
    defer cache.Close()
}
```

## Best Practices

1. **Error Handling**: Always use the custom error types from `pkg/errors` for consistent error handling
2. **Logging**: Use context-aware logging to track requests across services
3. **Configuration**: Store sensitive data in environment variables, not in config files
4. **Validation**: Validate all incoming requests using the validator package
5. **Caching**: Use Redis for session storage, rate limiting, and frequently accessed data
6. **Database**: Always use transactions for multi-step operations
7. **Security**: Never log sensitive data (passwords, tokens, etc.)

## Development

When adding new shared functionality:

1. Create a new package directory under `pkg/`
2. Add godoc comments for all exported functions and types
3. Write comprehensive unit tests
4. Update this README with usage examples
5. Add dependencies to `go.mod`
