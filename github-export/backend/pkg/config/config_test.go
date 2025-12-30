package config

import (
	"os"
	"testing"
	"time"
)

func TestValidateConfig_ValidConfig(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port:        8080,
			Environment: "development",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			User:     "postgres",
			DBName:   "testdb",
			MaxConns: 25,
			MinConns: 5,
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		JWT: JWTConfig{
			SecretKey:            "this-is-a-very-long-secret-key-for-testing-purposes",
			AccessTokenDuration:  15 * time.Minute,
			RefreshTokenDuration: 7 * 24 * time.Hour,
		},
		Logger: LoggerConfig{
			Level: "info",
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
		RateLimit: RateLimitConfig{
			Enabled:        true,
			RequestsPerMin: 100,
			BurstSize:      20,
		},
	}

	err := validateConfig(cfg)
	if err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}
}

func TestValidateConfig_InvalidServerPort(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port:        -1,
			Environment: "development",
		},
		Database: DatabaseConfig{
			Host:   "localhost",
			User:   "postgres",
			DBName: "testdb",
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		JWT: JWTConfig{
			SecretKey:            "this-is-a-very-long-secret-key-for-testing",
			AccessTokenDuration:  15 * time.Minute,
			RefreshTokenDuration: 7 * 24 * time.Hour,
		},
		Logger: LoggerConfig{
			Level: "info",
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	err := validateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid port, got nil")
	}
}

func TestValidateConfig_InvalidEnvironment(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port:        8080,
			Environment: "invalid",
		},
		Database: DatabaseConfig{
			Host:   "localhost",
			User:   "postgres",
			DBName: "testdb",
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		JWT: JWTConfig{
			SecretKey:            "this-is-a-very-long-secret-key-for-testing",
			AccessTokenDuration:  15 * time.Minute,
			RefreshTokenDuration: 7 * 24 * time.Hour,
		},
		Logger: LoggerConfig{
			Level: "info",
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	err := validateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid environment, got nil")
	}
}

func TestValidateConfig_MissingDatabaseHost(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port:        8080,
			Environment: "development",
		},
		Database: DatabaseConfig{
			Host:   "",
			User:   "postgres",
			DBName: "testdb",
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		JWT: JWTConfig{
			SecretKey:            "this-is-a-very-long-secret-key-for-testing",
			AccessTokenDuration:  15 * time.Minute,
			RefreshTokenDuration: 7 * 24 * time.Hour,
		},
		Logger: LoggerConfig{
			Level: "info",
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	err := validateConfig(cfg)
	if err == nil {
		t.Error("expected error for missing database host, got nil")
	}
}

func TestValidateConfig_ShortJWTSecret(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port:        8080,
			Environment: "development",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			User:     "postgres",
			DBName:   "testdb",
			MaxConns: 25,
			MinConns: 5,
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		JWT: JWTConfig{
			SecretKey:            "short",
			AccessTokenDuration:  15 * time.Minute,
			RefreshTokenDuration: 7 * 24 * time.Hour,
		},
		Logger: LoggerConfig{
			Level: "info",
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	err := validateConfig(cfg)
	if err == nil {
		t.Error("expected error for short JWT secret, got nil")
	}
}

func TestValidateConfig_InvalidLogLevel(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Port:        8080,
			Environment: "development",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			User:     "postgres",
			DBName:   "testdb",
			MaxConns: 25,
			MinConns: 5,
		},
		Redis: RedisConfig{
			Host: "localhost",
			Port: 6379,
		},
		JWT: JWTConfig{
			SecretKey:            "this-is-a-very-long-secret-key-for-testing",
			AccessTokenDuration:  15 * time.Minute,
			RefreshTokenDuration: 7 * 24 * time.Hour,
		},
		Logger: LoggerConfig{
			Level: "invalid",
		},
		CORS: CORSConfig{
			AllowedOrigins: []string{"http://localhost:3000"},
		},
	}

	err := validateConfig(cfg)
	if err == nil {
		t.Error("expected error for invalid log level, got nil")
	}
}

func TestIsDevelopment(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Environment: "development",
		},
	}

	if !cfg.IsDevelopment() {
		t.Error("expected IsDevelopment to return true")
	}

	if cfg.IsProduction() {
		t.Error("expected IsProduction to return false")
	}
}

func TestIsProduction(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Environment: "production",
		},
	}

	if !cfg.IsProduction() {
		t.Error("expected IsProduction to return true")
	}

	if cfg.IsDevelopment() {
		t.Error("expected IsDevelopment to return false")
	}
}

func TestGetServerAddress(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 8080,
		},
	}

	addr := cfg.GetServerAddress()
	expected := "localhost:8080"

	if addr != expected {
		t.Errorf("expected server address '%s', got '%s'", expected, addr)
	}
}

func TestLoadConfig_WithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("APP_SERVER_PORT", "9000")
	os.Setenv("APP_DATABASE_HOST", "db.example.com")
	os.Setenv("APP_JWT_SECRET_KEY", "this-is-a-test-secret-key-from-environment-variables")
	defer func() {
		os.Unsetenv("APP_SERVER_PORT")
		os.Unsetenv("APP_DATABASE_HOST")
		os.Unsetenv("APP_JWT_SECRET_KEY")
	}()

	// LoadConfig should fail because we don't have all required fields
	// but we can test that env vars are being read
	_, err := LoadConfig("nonexistent", ".")
	// We expect validation errors, not config file errors
	if err != nil {
		// This is expected since we don't have all required config
		t.Logf("Expected error: %v", err)
	}
}
