package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/excise-tax-portal/backend/internal/payment/handler"
	"github.com/excise-tax-portal/backend/internal/payment/repository"
	"github.com/excise-tax-portal/backend/internal/payment/service"
	"github.com/excise-tax-portal/backend/internal/payment/xrpl"
	"github.com/excise-tax-portal/backend/pkg/cache"
	"github.com/excise-tax-portal/backend/pkg/database"
	"github.com/excise-tax-portal/backend/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logConfig := logger.Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		Level:       getEnv("LOG_LEVEL", "info"),
	}

	log, err := logger.NewLogger(logConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("starting payment service")

	// Create context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database
	db, err := initDatabase(ctx, log.GetZapLogger())
	if err != nil {
		log.Fatal("failed to initialize database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Redis
	redisClient, err := initRedis(ctx, log.GetZapLogger())
	if err != nil {
		log.Fatal("failed to initialize Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// Initialize XRPL client
	xrplClient, err := initXRPLClient(ctx, log.GetZapLogger())
	if err != nil {
		log.Fatal("failed to initialize XRPL client", zap.Error(err))
	}
	defer xrplClient.Close()

	// Initialize price oracle
	oracle := initPriceOracle(ctx, log.GetZapLogger())
	oracle.Start(ctx)
	defer oracle.Stop()

	// Initialize payment monitor
	monitor := xrpl.NewMonitorService(xrplClient, log.GetZapLogger())
	if err := monitor.Start(ctx); err != nil {
		log.Fatal("failed to start payment monitor", zap.Error(err))
	}
	defer monitor.Stop(ctx)

	// Initialize services
	paymentRepo := repository.NewPaymentRepository(db)
	processor := xrpl.NewPaymentProcessor(xrplClient, oracle, monitor, log.GetZapLogger())
	paymentService := service.NewPaymentService(paymentRepo, redisClient, xrplClient, oracle, monitor, processor, log.GetZapLogger())

	// Initialize HTTP server
	srv := initHTTPServer(paymentService, log.GetZapLogger())

	// Start server
	go func() {
		port := getEnv("PORT", "8082")
		log.Info("starting HTTP server", zap.String("port", port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down payment service")

	// Graceful shutdown with 30 second timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("server forced to shutdown", zap.Error(err))
	}

	log.Info("payment service stopped")
}

func initDatabase(ctx context.Context, log *zap.Logger) (*database.PostgresDB, error) {
	cfg := &database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnvInt("DB_PORT", 5432),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "excise_tax"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		MaxConns: int32(getEnvInt("DB_MAX_CONNS", 25)),
		MinConns: int32(getEnvInt("DB_MIN_CONNS", 5)),
	}

	db, err := database.NewPostgresDB(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Info("database connected",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.DBName),
	)

	return db, nil
}

func initRedis(ctx context.Context, log *zap.Logger) (*cache.RedisClient, error) {
	cfg := &cache.Config{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnvInt("REDIS_PORT", 6379),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       getEnvInt("REDIS_DB", 0),
	}

	client, err := cache.NewRedisClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info("Redis connected",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
	)

	return client, nil
}

func initXRPLClient(ctx context.Context, log *zap.Logger) (*xrpl.Client, error) {
	network := getEnv("XRPL_NETWORK", "testnet")

	var serverURL string
	if network == "mainnet" {
		serverURL = getEnv("XRPL_MAINNET_SERVER", "wss://xrplcluster.com")
	} else {
		serverURL = getEnv("XRPL_TESTNET_SERVER", "wss://s.altnet.rippletest.net:51233")
	}

	cfg := &xrpl.Config{
		ServerURL:            serverURL,
		Network:              network,
		StateAddress:         getEnv(fmt.Sprintf("XRPL_%s_ADDRESS", network), ""),
		TransactionTimeout:   60 * time.Second,
		MaxReconnectAttempts: 10,
		ReconnectDelay:       5 * time.Second,
		BackoffMultiplier:    1.5,
	}

	client, err := xrpl.NewClient(cfg, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create XRPL client: %w", err)
	}

	if err := client.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize XRPL client: %w", err)
	}

	log.Info("XRPL client initialized",
		zap.String("network", network),
		zap.String("server", serverURL),
	)

	return client, nil
}

func initPriceOracle(ctx context.Context, log *zap.Logger) *xrpl.PriceOracle {
	cfg := &xrpl.OracleConfig{
		UpdateInterval:   30 * time.Second,
		CacheDuration:    60 * time.Second,
		OutlierThreshold: 0.10,
		MinSources:       2,
		HTTPTimeout:      10 * time.Second,
	}

	oracle := xrpl.NewPriceOracle(cfg, log)

	log.Info("price oracle initialized",
		zap.Duration("update_interval", cfg.UpdateInterval),
		zap.Duration("cache_duration", cfg.CacheDuration),
	)

	return oracle
}

func initHTTPServer(paymentService *service.PaymentService, log *zap.Logger) *http.Server {
	// Set Gin mode
	if getEnv("ENVIRONMENT", "development") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Recovery())
	router.Use(loggingMiddleware(log))
	router.Use(corsMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"service":   "payment-service",
			"timestamp": time.Now(),
		})
	})

	// Metrics endpoint (placeholder)
	router.GET("/metrics", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "payment-service",
			"version": "1.0.0",
		})
	})

	// API routes
	api := router.Group("/api/v1")
	paymentHandler := handler.NewPaymentHandler(paymentService, log)
	paymentHandler.RegisterRoutes(api)

	port := getEnv("PORT", "8082")
	return &http.Server{
		Addr:         ":" + port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func loggingMiddleware(log *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		log.Info("HTTP request",
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", c.ClientIP()),
		)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var intValue int
		if _, err := fmt.Sscanf(value, "%d", &intValue); err == nil {
			return intValue
		}
	}
	return defaultValue
}
