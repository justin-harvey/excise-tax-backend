package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/excise-tax-portal/backend/pkg/config"
	"github.com/excise-tax-portal/backend/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger. This service is still a scaffold (see the TODOs
	// below), so there is no loaded config yet to source these from -
	// matching the sensible defaults DefaultJWTConfig uses elsewhere.
	log, err := logger.NewLogger(logger.Config{Environment: "development", Level: "info"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Tax Service starting...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", zap.Error(err))
	}

	// TODO: Initialize database connection
	// TODO: Initialize Redis cache
	// TODO: Initialize tax calculation engine
	// TODO: Initialize router and handlers

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.TaxService.Port),
		Handler:      nil, // TODO: Add router
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("Tax Service listening", zap.Int("port", cfg.TaxService.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Tax Service...")

	// Graceful shutdown with 10 second timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Tax Service stopped")
}
