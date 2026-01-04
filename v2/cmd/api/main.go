package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/maxfelker/excise-tax-backend/v2/internal/api"
	"github.com/maxfelker/excise-tax-backend/v2/pkg/logger"
	"github.com/maxfelker/excise-tax-backend/v2/pkg/xrpl"
)

func main() {
	// Load configuration
	cfg, err := api.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logLevel := logger.LevelInfo
	switch cfg.Logging.Level {
	case "debug":
		logLevel = logger.LevelDebug
	case "warn":
		logLevel = logger.LevelWarn
	case "error":
		logLevel = logger.LevelError
	}

	log := logger.New(logger.Options{
		Level: logLevel,
		JSON:  cfg.Logging.Format == "json",
	})

	log.Info("starting XRPL API server",
		slog.String("env", cfg.Env),
		slog.Int("port", cfg.Port),
		slog.String("xrpl_url", cfg.XRPL.URL),
	)

	// Initialize XRPL client
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	xrplClient := xrpl.New(cfg.XRPL.URL,
		xrpl.WithLogger(log),
		xrpl.WithTimeout(cfg.XRPL.Timeout),
		xrpl.WithMaxRetries(cfg.XRPL.RetryAttempts),
	)

	// Connect to XRPL
	if err := xrplClient.Connect(ctx); err != nil {
		log.Error("failed to connect to XRPL", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer xrplClient.Close()

	log.Info("connected to XRPL", slog.String("url", cfg.XRPL.URL))

	// Create application
	app := api.NewApplication(cfg, log, xrplClient)

	// Start server
	if err := app.Serve(); err != nil {
		log.Error("server error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info("server stopped")
}
