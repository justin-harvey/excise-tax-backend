// Package main is the entry point for the authentication service.
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/excise-tax-portal/backend/internal/auth/handler"
	"github.com/excise-tax-portal/backend/internal/auth/middleware"
	"github.com/excise-tax-portal/backend/internal/auth/repository"
	"github.com/excise-tax-portal/backend/internal/auth/service"
	"github.com/excise-tax-portal/backend/pkg/database"
	"github.com/excise-tax-portal/backend/pkg/utils"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration from environment variables
	config := loadConfig()

	// Initialize context
	ctx := context.Background()

	// Initialize database connection
	dbConfig := &database.Config{
		Host:              config.DBHost,
		Port:              config.DBPort,
		User:              config.DBUser,
		Password:          config.DBPassword,
		DBName:            config.DBName,
		SSLMode:           config.DBSSLMode,
		MaxConns:          25,
		MinConns:          5,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   30 * time.Minute,
		HealthCheckPeriod: 1 * time.Minute,
	}

	db, err := database.NewPostgresDB(ctx, dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connection established")

	// Initialize JWT manager
	jwtConfig := &utils.JWTConfig{
		SecretKey:            config.JWTSecret,
		AccessTokenDuration:  time.Hour,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "excise-tax-portal",
	}

	jwtManager, err := utils.NewJWTManager(jwtConfig)
	if err != nil {
		log.Fatalf("Failed to initialize JWT manager: %v", err)
	}
	log.Println("JWT manager initialized")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)

	// Initialize services (no session repository - using stateless JWT only)
	authService := service.NewAuthService(userRepo, nil, jwtManager)

	// Initialize OAuth service (optional)
	var oauthService *service.OAuthService
	if config.OAuthEnabled {
		oauthConfig := &service.OAuthConfig{
			Provider:         service.ProviderStateDotGov,
			ClientID:         config.OAuthClientID,
			ClientSecret:     config.OAuthClientSecret,
			RedirectURI:      config.OAuthRedirectURI,
			AuthorizationURL: config.OAuthAuthURL,
			TokenURL:         config.OAuthTokenURL,
			UserInfoURL:      config.OAuthUserInfoURL,
			Scopes:           []string{"openid", "profile", "email"},
		}

		oauthService, err = service.NewOAuthService(oauthConfig)
		if err != nil {
			log.Printf("Warning: Failed to initialize OAuth service: %v", err)
		} else {
			log.Println("OAuth service initialized")
		}
	}

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	var oauthHandler *handler.OAuthHandler
	if oauthService != nil {
		oauthHandler = handler.NewOAuthHandler(oauthService, authService)
	}

	// Setup HTTP server
	router := setupRouter(authHandler, oauthHandler, authService, config)

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Starting auth service on port %d", config.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// setupRouter configures the Gin router with routes and middleware
func setupRouter(
	authHandler *handler.AuthHandler,
	oauthHandler *handler.OAuthHandler,
	authService *service.AuthService,
	config *Config,
) *gin.Engine {
	// Set Gin mode
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     config.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "auth-service",
			"time":    time.Now().UTC(),
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public auth routes (no authentication required)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/forgot-password", authHandler.ForgotPassword)
			auth.POST("/reset-password", authHandler.ResetPassword)

			// OAuth routes (if enabled)
			if oauthHandler != nil {
				oauth := auth.Group("/oauth")
				{
					oauth.GET("/authorize", oauthHandler.OAuthAuthorize)
					oauth.GET("/callback", oauthHandler.OAuthCallback)
					oauth.POST("/token", oauthHandler.OAuthToken)
					oauth.POST("/refresh", oauthHandler.OAuthRefresh)
				}
			}
		}

		// Protected routes (require authentication)
		protected := v1.Group("/auth")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			protected.POST("/logout", authHandler.Logout)
			protected.GET("/me", authHandler.GetCurrentUser)
			protected.POST("/change-password", authHandler.ChangePassword)
		}
	}

	return router
}

// Config holds application configuration
type Config struct {
	Environment       string
	Port              int
	DBHost            string
	DBPort            int
	DBUser            string
	DBPassword        string
	DBName            string
	DBSSLMode         string
	JWTSecret         string
	CORSOrigins       []string
	OAuthEnabled      bool
	OAuthClientID     string
	OAuthClientSecret string
	OAuthRedirectURI  string
	OAuthAuthURL      string
	OAuthTokenURL     string
	OAuthUserInfoURL  string
}

// loadConfig loads configuration from environment variables
func loadConfig() *Config {
	return &Config{
		Environment:       getEnv("ENVIRONMENT", "development"),
		Port:              getEnvAsInt("PORT", 8080),
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnvAsInt("DB_PORT", 5432),
		DBUser:            getEnv("DB_USER", "postgres"),
		DBPassword:        getEnv("DB_PASSWORD", ""),
		DBName:            getEnv("DB_NAME", "excise_tax"),
		DBSSLMode:         getEnv("DB_SSL_MODE", "disable"),
		JWTSecret:         getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		CORSOrigins:       []string{getEnv("CORS_ORIGIN", "http://localhost:3000")},
		OAuthEnabled:      getEnv("OAUTH_ENABLED", "false") == "true",
		OAuthClientID:     getEnv("OAUTH_CLIENT_ID", ""),
		OAuthClientSecret: getEnv("OAUTH_CLIENT_SECRET", ""),
		OAuthRedirectURI:  getEnv("OAUTH_REDIRECT_URI", ""),
		OAuthAuthURL:      getEnv("OAUTH_AUTH_URL", ""),
		OAuthTokenURL:     getEnv("OAUTH_TOKEN_URL", ""),
		OAuthUserInfoURL:  getEnv("OAUTH_USER_INFO_URL", ""),
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets an environment variable as an integer with a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	var value int
	if _, err := fmt.Sscanf(valueStr, "%d", &value); err != nil {
		return defaultValue
	}
	return value
}
