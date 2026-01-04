package api

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the configuration for the API server
type Config struct {
	Port     int
	Env      string
	XRPL     XRPLConfig
	Logging  LogConfig
	CORS     CORSConfig
	Limiter  RateLimitConfig
	Timeouts TimeoutConfig
}

// XRPLConfig holds XRPL client configuration
type XRPLConfig struct {
	URL           string
	Timeout       time.Duration
	RetryAttempts int
	RetryDelay    time.Duration
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string // "debug", "info", "warn", "error"
	Format string // "json", "text"
}

// CORSConfig holds CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled bool
	RPS     float64 // requests per second
	Burst   int
}

// TimeoutConfig holds timeout configuration
type TimeoutConfig struct {
	Read  time.Duration
	Write time.Duration
	Idle  time.Duration
}

// LoadConfig loads configuration from flags, environment variables, and defaults
func LoadConfig() (*Config, error) {
	var cfg Config

	// Command-line flags
	flag.IntVar(&cfg.Port, "port", 8080, "API server port")
	flag.StringVar(&cfg.Env, "env", "development", "Environment (development|production)")
	flag.StringVar(&cfg.XRPL.URL, "xrpl-url", "", "XRPL WebSocket URL")
	flag.StringVar(&cfg.Logging.Level, "log-level", "info", "Log level (debug|info|warn|error)")
	flag.StringVar(&cfg.Logging.Format, "log-format", "json", "Log format (json|text)")
	flag.BoolVar(&cfg.Limiter.Enabled, "rate-limit", true, "Enable rate limiting")
	flag.Float64Var(&cfg.Limiter.RPS, "rate-limit-rps", 10, "Rate limit requests per second")
	flag.IntVar(&cfg.Limiter.Burst, "rate-limit-burst", 20, "Rate limit burst size")

	flag.Parse()

	// Environment variables override flags
	if envPort := os.Getenv("API_PORT"); envPort != "" {
		port, err := strconv.Atoi(envPort)
		if err != nil {
			return nil, fmt.Errorf("invalid API_PORT: %w", err)
		}
		cfg.Port = port
	}

	if envEnv := os.Getenv("ENV"); envEnv != "" {
		cfg.Env = envEnv
	}

	if envXRPLURL := os.Getenv("XRPL_URL"); envXRPLURL != "" {
		cfg.XRPL.URL = envXRPLURL
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		cfg.Logging.Level = envLogLevel
	}

	// Set defaults
	if cfg.XRPL.URL == "" {
		// Default to testnet
		cfg.XRPL.URL = "wss://s.altnet.rippletest.net:51233"
	}

	cfg.XRPL.Timeout = 30 * time.Second
	cfg.XRPL.RetryAttempts = 3
	cfg.XRPL.RetryDelay = time.Second

	// CORS defaults
	cfg.CORS.AllowedOrigins = []string{"*"}
	cfg.CORS.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	cfg.CORS.AllowedHeaders = []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"}

	// Timeout defaults
	cfg.Timeouts.Read = 10 * time.Second
	cfg.Timeouts.Write = 30 * time.Second
	cfg.Timeouts.Idle = 120 * time.Second

	// Adjust settings based on environment
	if cfg.Env == "production" {
		if cfg.Logging.Level == "debug" {
			cfg.Logging.Level = "info"
		}
		// Stricter CORS in production
		if envOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); envOrigins != "" {
			cfg.CORS.AllowedOrigins = []string{envOrigins}
		}
	} else {
		// Development mode
		if cfg.Logging.Level == "info" {
			cfg.Logging.Level = "debug"
		}
		cfg.Logging.Format = "text"
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if c.Env != "development" && c.Env != "staging" && c.Env != "production" {
		return fmt.Errorf("env must be development, staging, or production")
	}

	if c.XRPL.URL == "" {
		return fmt.Errorf("XRPL URL is required")
	}

	if c.Logging.Level != "debug" && c.Logging.Level != "info" && c.Logging.Level != "warn" && c.Logging.Level != "error" {
		return fmt.Errorf("log level must be debug, info, warn, or error")
	}

	if c.Logging.Format != "json" && c.Logging.Format != "text" {
		return fmt.Errorf("log format must be json or text")
	}

	if c.Limiter.RPS <= 0 {
		return fmt.Errorf("rate limit RPS must be positive")
	}

	if c.Limiter.Burst <= 0 {
		return fmt.Errorf("rate limit burst must be positive")
	}

	return nil
}
