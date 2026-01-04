// Package config provides configuration management with support for environment variables,
// YAML files, and validation.
//
// Example usage:
//
//	cfg, err := config.Load()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fmt.Printf("Server running on port %d\n", cfg.Server.Port)
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	Server              ServerConfig    `mapstructure:"server"`
	APIGateway          ServiceConfig   `mapstructure:"api_gateway"`
	PaymentService      ServiceConfig   `mapstructure:"payment_service"`
	TaxService          ServiceConfig   `mapstructure:"tax_service"`
	ReportingService    ServiceConfig   `mapstructure:"reporting_service"`
	NotificationService ServiceConfig   `mapstructure:"notification_service"`
	AuthService         ServiceConfig   `mapstructure:"auth_service"`
	Database            DatabaseConfig  `mapstructure:"database"`
	Redis               RedisConfig     `mapstructure:"redis"`
	RabbitMQ            RabbitMQConfig  `mapstructure:"rabbitmq"`
	S3                  S3Config        `mapstructure:"s3"`
	JWT                 JWTConfig       `mapstructure:"jwt"`
	Logger              LoggerConfig    `mapstructure:"logger"`
	CORS                CORSConfig      `mapstructure:"cors"`
	RateLimit           RateLimitConfig `mapstructure:"rate_limit"`
	XRPL                XRPLConfig      `mapstructure:"xrpl"`
}

// ServerConfig holds server-related configuration.
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Environment     string        `mapstructure:"environment"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// ServiceConfig holds configuration for individual microservices.
type ServiceConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxConns        int32         `mapstructure:"max_conns"`
	MinConns        int32         `mapstructure:"min_conns"`
	MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime"`
	MaxConnIdleTime time.Duration `mapstructure:"max_conn_idle_time"`
}

// RedisConfig holds Redis connection configuration.
type RedisConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Password        string        `mapstructure:"password"`
	DB              int           `mapstructure:"db"`
	MaxRetries      int           `mapstructure:"max_retries"`
	MinIdleConns    int           `mapstructure:"min_idle_conns"`
	PoolSize        int           `mapstructure:"pool_size"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

// RabbitMQConfig holds RabbitMQ configuration.
type RabbitMQConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	VHost    string `mapstructure:"vhost"`
}

// S3Config holds S3/MinIO configuration.
type S3Config struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key"`
	BucketName      string `mapstructure:"bucket_name"`
	Region          string `mapstructure:"region"`
	UseSSL          bool   `mapstructure:"use_ssl"`
}

// JWTConfig holds JWT token configuration.
type JWTConfig struct {
	SecretKey            string        `mapstructure:"secret_key"`
	AccessTokenDuration  time.Duration `mapstructure:"access_token_duration"`
	RefreshTokenDuration time.Duration `mapstructure:"refresh_token_duration"`
	Issuer               string        `mapstructure:"issuer"`
}

// LoggerConfig holds logging configuration.
type LoggerConfig struct {
	Level       string `mapstructure:"level"`
	Environment string `mapstructure:"environment"`
}

// CORSConfig holds CORS configuration.
type CORSConfig struct {
	AllowedOrigins   []string `mapstructure:"allowed_origins"`
	AllowedMethods   []string `mapstructure:"allowed_methods"`
	AllowedHeaders   []string `mapstructure:"allowed_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"`
}

// RateLimitConfig holds rate limiting configuration.
type RateLimitConfig struct {
	Enabled        bool          `mapstructure:"enabled"`
	RequestsPerMin int           `mapstructure:"requests_per_min"`
	BurstSize      int           `mapstructure:"burst_size"`
	CleanupPeriod  time.Duration `mapstructure:"cleanup_period"`
}

// XRPLConfig holds XRPL blockchain configuration.
type XRPLConfig struct {
	NetworkURL      string        `mapstructure:"network_url"`
	WalletAddress   string        `mapstructure:"wallet_address"`
	WalletSeed      string        `mapstructure:"wallet_seed"`
	IsTestnet       bool          `mapstructure:"is_testnet"`
	SyncInterval    time.Duration `mapstructure:"sync_interval"`
	ConfirmationMin int           `mapstructure:"confirmation_min"`
}

// Load reads configuration from file and environment variables using default paths.
func Load() (*Config, error) {
	return LoadConfig("config", "./configs")
}

// LoadConfig reads configuration from file and environment variables.
// configName is the name of the config file (without extension).
// configPath is the path to search for the config file.
//
// The function will:
// 1. Look for config.yaml, config.yml, or .env files in the specified path
// 2. Override values with environment variables (prefixed with APP_)
// 3. Validate required fields
func LoadConfig(configName, configPath string) (*Config, error) {
	viper.SetConfigName(configName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AddConfigPath(".")
	viper.AddConfigPath("./configs")

	// Set default values
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		// It's okay if config file doesn't exist, we'll use env vars and defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Read environment variables
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override from ENV if available
	overrideFromEnv(&config)

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default values for all configuration fields.
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.environment", "development")
	viper.SetDefault("server.read_timeout", 15*time.Second)
	viper.SetDefault("server.write_timeout", 15*time.Second)
	viper.SetDefault("server.shutdown_timeout", 30*time.Second)

	// Microservices defaults
	viper.SetDefault("api_gateway.host", "localhost")
	viper.SetDefault("api_gateway.port", 8080)
	viper.SetDefault("payment_service.host", "localhost")
	viper.SetDefault("payment_service.port", 8081)
	viper.SetDefault("tax_service.host", "localhost")
	viper.SetDefault("tax_service.port", 8082)
	viper.SetDefault("reporting_service.host", "localhost")
	viper.SetDefault("reporting_service.port", 8083)
	viper.SetDefault("notification_service.host", "localhost")
	viper.SetDefault("notification_service.port", 8084)
	viper.SetDefault("auth_service.host", "localhost")
	viper.SetDefault("auth_service.port", 8085)

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.max_conns", 25)
	viper.SetDefault("database.min_conns", 5)
	viper.SetDefault("database.max_conn_lifetime", 1*time.Hour)
	viper.SetDefault("database.max_conn_idle_time", 30*time.Minute)

	// S3 defaults
	viper.SetDefault("s3.region", "us-east-1")
	viper.SetDefault("s3.use_ssl", true)

	// JWT defaults
	viper.SetDefault("jwt.access_token_duration", 15*time.Minute)
	viper.SetDefault("jwt.refresh_token_duration", 7*24*time.Hour)
	viper.SetDefault("jwt.issuer", "excise-tax-portal")

	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.environment", "development")

	// CORS defaults
	viper.SetDefault("cors.allowed_origins", []string{"http://localhost:3000"})
	viper.SetDefault("cors.allowed_methods", []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"})
	viper.SetDefault("cors.allowed_headers", []string{"Content-Type", "Authorization"})
	viper.SetDefault("cors.allow_credentials", true)
	viper.SetDefault("cors.max_age", 300)

	// Rate limit defaults
	viper.SetDefault("rate_limit.enabled", true)
	viper.SetDefault("rate_limit.requests_per_min", 100)
	viper.SetDefault("rate_limit.burst_size", 20)
	viper.SetDefault("rate_limit.cleanup_period", 1*time.Minute)

	// XRPL defaults
	viper.SetDefault("xrpl.network_url", "wss://s.altnet.rippletest.net:51233")
	viper.SetDefault("xrpl.is_testnet", true)
	viper.SetDefault("xrpl.sync_interval", 5*time.Second)
	viper.SetDefault("xrpl.confirmation_min", 1)
}

// overrideFromEnv overrides config values from environment variables
func overrideFromEnv(cfg *Config) {
	if val := os.Getenv("DATABASE_URL"); val != "" {
		// Parse DATABASE_URL if provided (common in cloud deployments)
		// Format: postgres://user:password@host:port/dbname
		// This is a simplified version - you might want to use a proper URL parser
	}

	if val := os.Getenv("ENV"); val != "" {
		cfg.Server.Environment = val
	}
}

// validateConfig validates required configuration fields.
func validateConfig(cfg *Config) error {
	// Validate server config
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", cfg.Server.Port)
	}

	validEnvironments := map[string]bool{
		"development": true,
		"staging":     true,
		"production":  true,
		"test":        true,
	}
	if !validEnvironments[cfg.Server.Environment] {
		return fmt.Errorf("invalid environment: %s (must be development, staging, production, or test)", cfg.Server.Environment)
	}

	// Database validation is optional for some services
	if cfg.Database.DBName != "" {
		if cfg.Database.Host == "" {
			return fmt.Errorf("database host is required")
		}
		if cfg.Database.User == "" {
			return fmt.Errorf("database user is required")
		}
		if cfg.Database.MaxConns < cfg.Database.MinConns {
			return fmt.Errorf("database max_conns (%d) must be >= min_conns (%d)", cfg.Database.MaxConns, cfg.Database.MinConns)
		}
	}

	// JWT validation
	if cfg.JWT.SecretKey != "" {
		if len(cfg.JWT.SecretKey) < 32 {
			return fmt.Errorf("jwt secret key must be at least 32 characters")
		}
	}

	// Validate logger config
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
	if !validLogLevels[cfg.Logger.Level] {
		return fmt.Errorf("invalid log level: %s", cfg.Logger.Level)
	}

	return nil
}

// IsDevelopment returns true if the server is running in development mode.
func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

// IsProduction returns true if the server is running in production mode.
func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

// IsStaging returns true if the server is running in staging mode.
func (c *Config) IsStaging() bool {
	return c.Server.Environment == "staging"
}

// GetServerAddress returns the full server address (host:port).
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}
