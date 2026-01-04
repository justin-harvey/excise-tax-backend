package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/maxfelker/excise-tax-backend/v2/internal/api"
	"github.com/maxfelker/excise-tax-backend/v2/internal/tax"
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

	// Initialize tax calculator with sample rates
	taxRates := []tax.TaxRate{
		{
			ProductType:   tax.ProductTypeBeer,
			RatePerUnit:   0.58,
			UnitType:      tax.UnitTypeGallon,
			EffectiveDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Description:   "Federal excise tax on beer",
			Jurisdiction:  "federal",
		},
		{
			ProductType:   tax.ProductTypeWine,
			RatePerUnit:   1.07,
			UnitType:      tax.UnitTypeGallon,
			EffectiveDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Description:   "Federal excise tax on wine",
			Jurisdiction:  "federal",
		},
		{
			ProductType:   tax.ProductTypeSpirits,
			RatePerUnit:   13.50,
			UnitType:      tax.UnitTypeGallon,
			EffectiveDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Description:   "Federal excise tax on distilled spirits",
			Jurisdiction:  "federal",
		},
	}

	taxCalc := tax.NewCalculator(taxRates, tax.WithLogger(log))

	// Create application
	app := api.NewApplication(cfg, log, xrplClient, taxCalc)

	// Start server
	if err := app.Serve(); err != nil {
		log.Error("server error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	log.Info("server stopped")
}
