package api

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maxfelker/excise-tax-backend/v2/internal/tax"
	"github.com/maxfelker/excise-tax-backend/v2/pkg/xrpl"
)

// Application holds the dependencies for the HTTP handlers
type Application struct {
	config        *Config
	logger        *slog.Logger
	xrpl          *xrpl.Client
	taxCalculator *tax.Calculator
}

// NewApplication creates a new Application instance
func NewApplication(cfg *Config, logger *slog.Logger, xrplClient *xrpl.Client, taxCalc *tax.Calculator) *Application {
	return &Application{
		config:        cfg,
		logger:        logger,
		xrpl:          xrplClient,
		taxCalculator: taxCalc,
	}
}

// Serve starts the HTTP server with graceful shutdown
func (app *Application) Serve() error {
	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.Port),
		Handler:      app.routes(),
		ReadTimeout:  app.config.Timeouts.Read,
		WriteTimeout: app.config.Timeouts.Write,
		IdleTimeout:  app.config.Timeouts.Idle,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	// Channel to receive server errors
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		app.logger.Info("starting server",
			slog.String("env", app.config.Env),
			slog.Int("port", app.config.Port),
			slog.String("xrpl_url", app.config.XRPL.URL),
		)

		serverErrors <- srv.ListenAndServe()
	}()

	// Channel to listen for interrupt signals
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive a signal or server error
	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server error: %w", err)
		}

	case sig := <-shutdown:
		app.logger.Info("received shutdown signal",
			slog.String("signal", sig.String()),
		)

		// Give outstanding requests a deadline for completion
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Gracefully shutdown the server
		if err := srv.Shutdown(ctx); err != nil {
			// Force close if graceful shutdown fails
			if closeErr := srv.Close(); closeErr != nil {
				return fmt.Errorf("failed to close server: %w", closeErr)
			}
			return fmt.Errorf("failed to gracefully shutdown server: %w", err)
		}

		app.logger.Info("server stopped gracefully")
	}

	return nil
}
