// Package main is the entry point for the API Gateway service.
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"github.com/excise-tax-portal/backend/internal/api-gateway/handler"
	"github.com/excise-tax-portal/backend/internal/api-gateway/middleware"
	"github.com/excise-tax-portal/backend/internal/api-gateway/proxy"
	"github.com/excise-tax-portal/backend/internal/api-gateway/router"
	"github.com/excise-tax-portal/backend/pkg/cache"
	"github.com/excise-tax-portal/backend/pkg/logger"
)

// Config represents the application configuration
type Config struct {
	Server struct {
		Port         int           `yaml:"port"`
		ReadTimeout  time.Duration `yaml:"read_timeout"`
		WriteTimeout time.Duration `yaml:"write_timeout"`
	} `yaml:"server"`

	RateLimit struct {
		RequestsPerMinute int `yaml:"requests_per_minute"`
		Burst             int `yaml:"burst"`
	} `yaml:"rate_limit"`

	CORS struct {
		AllowedOrigins []string `yaml:"allowed_origins"`
		AllowedMethods []string `yaml:"allowed_methods"`
		AllowedHeaders []string `yaml:"allowed_headers"`
	} `yaml:"cors"`

	Services struct {
		AuthURL      string `yaml:"auth_url"`
		PaymentURL   string `yaml:"payment_url"`
		TaxURL       string `yaml:"tax_url"`
		ReportingURL string `yaml:"reporting_url"`
	} `yaml:"services"`

	// JWT holds the secret the gateway uses to validate access tokens locally
	// (see middleware.Auth). It was missing from this struct even though the
	// auth middleware has always required it - set via JWT_SECRET_KEY, not
	// committed here, same as every other secret in this repo.
	JWT struct {
		SecretKey string `yaml:"secret_key"`
	} `yaml:"jwt"`

	Redis struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	} `yaml:"redis"`

	Logging struct {
		Level       string `yaml:"level"`
		Environment string `yaml:"environment"`
	} `yaml:"logging"`
}

// RedisHealthAdapter adapts cache.RedisClient to handler.RedisHealthChecker
type RedisHealthAdapter struct {
	client *cache.RedisClient
}

func (r *RedisHealthAdapter) HealthCheck(ctx context.Context) error {
	return r.client.HealthCheck(ctx)
}

func main() {
	// Load configuration
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.NewLogger(logger.Config{
		Environment: cfg.Logging.Environment,
		Level:       cfg.Logging.Level,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting API Gateway service",
		zap.Int("port", cfg.Server.Port),
		zap.String("environment", cfg.Logging.Environment),
	)

	// Initialize Redis client
	redisConfig := &cache.Config{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}

	ctx := context.Background()
	redisClient, err := cache.NewRedisClient(ctx, redisConfig)
	if err != nil {
		log.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	log.Info("Connected to Redis",
		zap.String("host", cfg.Redis.Host),
		zap.Int("port", cfg.Redis.Port),
	)

	// Initialize service proxy
	proxyConfig := proxy.ProxyConfig{
		Services: map[string]proxy.ServiceConfig{
			"auth": {
				Name:            "auth",
				BaseURL:         cfg.Services.AuthURL,
				Timeout:         30 * time.Second,
				HealthCheckPath: "/health",
			},
			"payment": {
				Name:            "payment",
				BaseURL:         cfg.Services.PaymentURL,
				Timeout:         30 * time.Second,
				HealthCheckPath: "/health",
			},
			"tax": {
				Name:            "tax",
				BaseURL:         cfg.Services.TaxURL,
				Timeout:         30 * time.Second,
				HealthCheckPath: "/health",
			},
			"reporting": {
				Name:            "reporting",
				BaseURL:         cfg.Services.ReportingURL,
				Timeout:         30 * time.Second,
				HealthCheckPath: "/health",
			},
		},
		DefaultTimeout:  30 * time.Second,
		MaxIdleConns:    100,
		IdleConnTimeout: 90 * time.Second,
	}

	serviceProxy := proxy.NewServiceProxy(proxyConfig, log.GetZapLogger())

	// Start periodic health checks for backend services
	healthCheckCtx, healthCheckCancel := context.WithCancel(context.Background())
	defer healthCheckCancel()
	go serviceProxy.StartHealthChecks(healthCheckCtx, 30*time.Second)

	// Initialize rate limiter
	rateLimitConfig := middleware.DefaultRateLimitConfig()
	rateLimitConfig.RequestsPerWindow = cfg.RateLimit.RequestsPerMinute
	rateLimitConfig.WindowSize = 15 * time.Minute // As per spec
	rateLimitConfig.BurstSize = cfg.RateLimit.Burst

	rateLimiter := middleware.NewRateLimiter(redisClient.GetClient(), rateLimitConfig, log.GetZapLogger())

	// Initialize health handler
	redisHealthAdapter := &RedisHealthAdapter{client: redisClient}
	healthHandler := handler.NewHealthHandler(serviceProxy, redisHealthAdapter, log.GetZapLogger())

	// Configure CORS
	corsConfig := middleware.CORSConfig{
		AllowedOrigins:   cfg.CORS.AllowedOrigins,
		AllowedMethods:   cfg.CORS.AllowedMethods,
		AllowedHeaders:   cfg.CORS.AllowedHeaders,
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           3600,
	}

	// Configure authentication
	authConfig := middleware.AuthConfig{
		JWTSecretKey: cfg.JWT.SecretKey,
		AnonymousRoutes: []string{
			"/api/v1/auth/login",
			"/api/v1/auth/register",
			"/api/v1/auth/forgot-password",
			"/api/v1/auth/reset-password",
			"/health",
			"/health/liveness",
			"/health/readiness",
			"/metrics",
		},
		OptionalAuthRoutes: []string{},
	}

	// Setup router
	routerConfig := router.Config{
		Logger:          log.GetZapLogger(),
		ServiceProxy:    serviceProxy,
		RateLimiter:     rateLimiter,
		HealthHandler:   healthHandler,
		CORSConfig:      corsConfig,
		AuthConfig:      authConfig,
		EnableMetrics:   true,
		EnableDebugMode: cfg.Logging.Environment == "development",
	}

	r := router.Setup(routerConfig)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Info("API Gateway server listening",
			zap.String("address", srv.Addr),
		)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down API Gateway server...")

	// Cancel health check goroutine
	healthCheckCancel()

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}

	log.Info("API Gateway server stopped")
}

// loadConfig loads configuration from file and environment variables
func loadConfig() (*Config, error) {
	cfg := &Config{}

	// Set defaults
	cfg.Server.Port = 8080
	cfg.Server.ReadTimeout = 30 * time.Second
	cfg.Server.WriteTimeout = 30 * time.Second
	cfg.RateLimit.RequestsPerMinute = 100
	cfg.RateLimit.Burst = 20
	cfg.Logging.Level = "info"
	cfg.Logging.Environment = "development"

	// Try to load from config file
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Config file is optional, use defaults and environment variables
		fmt.Printf("Config file not found, using defaults and environment variables\n")
	} else {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Override with environment variables
	if port := os.Getenv("PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Server.Port)
	}

	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		cfg.Redis.Host = redisHost
	} else if cfg.Redis.Host == "" {
		cfg.Redis.Host = "localhost"
	}

	if redisPort := os.Getenv("REDIS_PORT"); redisPort != "" {
		fmt.Sscanf(redisPort, "%d", &cfg.Redis.Port)
	} else if cfg.Redis.Port == 0 {
		cfg.Redis.Port = 6379
	}

	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		cfg.Redis.Password = redisPassword
	}

	// Service URLs
	if authURL := os.Getenv("AUTH_SERVICE_URL"); authURL != "" {
		cfg.Services.AuthURL = authURL
	} else if cfg.Services.AuthURL == "" {
		cfg.Services.AuthURL = "http://localhost:8084"
	}

	if paymentURL := os.Getenv("PAYMENT_SERVICE_URL"); paymentURL != "" {
		cfg.Services.PaymentURL = paymentURL
	} else if cfg.Services.PaymentURL == "" {
		cfg.Services.PaymentURL = "http://localhost:8081"
	}

	if taxURL := os.Getenv("TAX_SERVICE_URL"); taxURL != "" {
		cfg.Services.TaxURL = taxURL
	} else if cfg.Services.TaxURL == "" {
		cfg.Services.TaxURL = "http://localhost:8082"
	}

	if reportingURL := os.Getenv("REPORTING_SERVICE_URL"); reportingURL != "" {
		cfg.Services.ReportingURL = reportingURL
	} else if cfg.Services.ReportingURL == "" {
		cfg.Services.ReportingURL = "http://localhost:8083"
	}

	if jwtSecret := os.Getenv("JWT_SECRET_KEY"); jwtSecret != "" {
		cfg.JWT.SecretKey = jwtSecret
	}

	// Set default CORS if not configured
	if len(cfg.CORS.AllowedOrigins) == 0 {
		cfg.CORS.AllowedOrigins = []string{"http://localhost:3000", "https://tax.state.gov"}
	}
	if len(cfg.CORS.AllowedMethods) == 0 {
		cfg.CORS.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(cfg.CORS.AllowedHeaders) == 0 {
		cfg.CORS.AllowedHeaders = []string{"Content-Type", "Authorization"}
	}

	return cfg, nil
}
