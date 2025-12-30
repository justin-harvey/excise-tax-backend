// Package router provides HTTP routing configuration for the API Gateway.
package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"excise-tax-portal/backend/internal/api-gateway/handler"
	"excise-tax-portal/backend/internal/api-gateway/middleware"
	"excise-tax-portal/backend/internal/api-gateway/proxy"
)

// Config holds the router configuration
type Config struct {
	Logger            *zap.Logger
	ServiceProxy      *proxy.ServiceProxy
	RateLimiter       *middleware.RateLimiter
	HealthHandler     *handler.HealthHandler
	CORSConfig        middleware.CORSConfig
	AuthConfig        middleware.AuthConfig
	EnableMetrics     bool
	EnableDebugMode   bool
}

// Setup configures and returns the Gin router with all middleware and routes
func Setup(config Config) *gin.Engine {
	// Set Gin mode
	if !config.EnableDebugMode {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.New()

	// Apply global middleware in order:
	// 1. Request ID generation (first, so all logs have request ID)
	router.Use(middleware.RequestID())

	// 2. Panic recovery (catch panics early)
	router.Use(middleware.Recovery(config.Logger.GetZapLogger()))

	// 3. CORS handling (before authentication)
	router.Use(middleware.CORS(config.CORSConfig))

	// 4. Request logging (after request ID and recovery)
	router.Use(middleware.Logging(config.Logger.GetZapLogger()))

	// Health check endpoints (no authentication required)
	router.GET("/health", config.HealthHandler.Health)
	router.GET("/health/liveness", config.HealthHandler.Liveness)
	router.GET("/health/readiness", config.HealthHandler.Readiness)

	// Metrics endpoint (Prometheus format)
	if config.EnableMetrics {
		router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	}

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth service routes (no authentication required for login/register)
		authGroup := v1.Group("/auth")
		{
			// Anonymous routes (login, register, etc.)
			authGroup.POST("/login", config.ServiceProxy.ProxyRequest("auth", "/api/v1/auth"))
			authGroup.POST("/register", config.ServiceProxy.ProxyRequest("auth", "/api/v1/auth"))
			authGroup.POST("/forgot-password", config.ServiceProxy.ProxyRequest("auth", "/api/v1/auth"))
			authGroup.POST("/reset-password", config.ServiceProxy.ProxyRequest("auth", "/api/v1/auth"))

			// Authenticated routes
			authProtected := authGroup.Group("")
			authProtected.Use(middleware.Auth(config.AuthConfig, config.Logger.GetZapLogger()))
			{
				authProtected.POST("/logout", config.ServiceProxy.ProxyRequest("auth", "/api/v1/auth"))
				authProtected.GET("/me", config.ServiceProxy.ProxyRequest("auth", "/api/v1/auth"))
				authProtected.PUT("/password", config.ServiceProxy.ProxyRequest("auth", "/api/v1/auth"))
				authProtected.PUT("/profile", config.ServiceProxy.ProxyRequest("auth", "/api/v1/auth"))
			}
		}

		// Payment service routes (authentication required)
		paymentGroup := v1.Group("/payments")
		paymentGroup.Use(middleware.Auth(config.AuthConfig, config.Logger.GetZapLogger()))
		paymentGroup.Use(config.RateLimiter.RateLimit())
		{
			paymentGroup.POST("", config.ServiceProxy.ProxyRequest("payment", "/api/v1/payments"))
			paymentGroup.GET("/:id", config.ServiceProxy.ProxyRequest("payment", "/api/v1/payments"))
			paymentGroup.GET("", config.ServiceProxy.ProxyRequest("payment", "/api/v1/payments"))
			paymentGroup.PUT("/:id/status", config.ServiceProxy.ProxyRequest("payment", "/api/v1/payments"))
			paymentGroup.POST("/:id/refund", config.ServiceProxy.ProxyRequest("payment", "/api/v1/payments"))
		}

		// Tax service routes (authentication required)
		taxGroup := v1.Group("/tax")
		taxGroup.Use(middleware.Auth(config.AuthConfig, config.Logger.GetZapLogger()))
		taxGroup.Use(config.RateLimiter.RateLimit())
		{
			// Tax calculations
			taxGroup.POST("/calculate", config.ServiceProxy.ProxyRequest("tax", "/api/v1/tax"))
			taxGroup.POST("/validate", config.ServiceProxy.ProxyRequest("tax", "/api/v1/tax"))

			// Tax filings
			taxGroup.POST("/filings", config.ServiceProxy.ProxyRequest("tax", "/api/v1/tax"))
			taxGroup.GET("/filings/:id", config.ServiceProxy.ProxyRequest("tax", "/api/v1/tax"))
			taxGroup.GET("/filings", config.ServiceProxy.ProxyRequest("tax", "/api/v1/tax"))
			taxGroup.PUT("/filings/:id", config.ServiceProxy.ProxyRequest("tax", "/api/v1/tax"))
			taxGroup.POST("/filings/:id/submit", config.ServiceProxy.ProxyRequest("tax", "/api/v1/tax"))

			// Tax rates
			taxGroup.GET("/rates", config.ServiceProxy.ProxyRequest("tax", "/api/v1/tax"))
			taxGroup.GET("/rates/:commodity", config.ServiceProxy.ProxyRequest("tax", "/api/v1/tax"))
		}

		// Reporting service routes (authentication required)
		reportGroup := v1.Group("/reports")
		reportGroup.Use(middleware.Auth(config.AuthConfig, config.Logger.GetZapLogger()))
		reportGroup.Use(config.RateLimiter.RateLimit())
		{
			reportGroup.POST("/generate", config.ServiceProxy.ProxyRequest("reporting", "/api/v1/reports"))
			reportGroup.GET("/:id", config.ServiceProxy.ProxyRequest("reporting", "/api/v1/reports"))
			reportGroup.GET("", config.ServiceProxy.ProxyRequest("reporting", "/api/v1/reports"))
			reportGroup.GET("/:id/download", config.ServiceProxy.ProxyRequest("reporting", "/api/v1/reports"))
			reportGroup.DELETE("/:id", config.ServiceProxy.ProxyRequest("reporting", "/api/v1/reports"))
		}

		// Admin routes (authentication + admin role required)
		adminGroup := v1.Group("/admin")
		adminGroup.Use(middleware.Auth(config.AuthConfig, config.Logger.GetZapLogger()))
		adminGroup.Use(middleware.RequireRole("admin", config.Logger.GetZapLogger()))
		adminGroup.Use(config.RateLimiter.RateLimit())
		{
			// User management
			adminGroup.GET("/users", config.ServiceProxy.ProxyRequest("auth", "/api/v1/admin"))
			adminGroup.GET("/users/:id", config.ServiceProxy.ProxyRequest("auth", "/api/v1/admin"))
			adminGroup.PUT("/users/:id", config.ServiceProxy.ProxyRequest("auth", "/api/v1/admin"))
			adminGroup.DELETE("/users/:id", config.ServiceProxy.ProxyRequest("auth", "/api/v1/admin"))
			adminGroup.PUT("/users/:id/role", config.ServiceProxy.ProxyRequest("auth", "/api/v1/admin"))

			// System configuration
			adminGroup.GET("/config", config.ServiceProxy.ProxyRequest("auth", "/api/v1/admin"))
			adminGroup.PUT("/config", config.ServiceProxy.ProxyRequest("auth", "/api/v1/admin"))

			// Analytics and monitoring
			adminGroup.GET("/analytics", config.ServiceProxy.ProxyRequest("reporting", "/api/v1/admin"))
			adminGroup.GET("/audit-logs", config.ServiceProxy.ProxyRequest("auth", "/api/v1/admin"))
		}
	}

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "The requested endpoint does not exist",
			},
		})
	})

	return router
}
