package logger

import (
	"io"
	"log/slog"
	"os"
)

// Level represents the log level
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Options configures the logger
type Options struct {
	Level  Level
	JSON   bool
	Writer io.Writer
}

// New creates a new structured logger
func New(opts Options) *slog.Logger {
	// Default to info level
	level := slog.LevelInfo
	switch opts.Level {
	case LevelDebug:
		level = slog.LevelDebug
	case LevelInfo:
		level = slog.LevelInfo
	case LevelWarn:
		level = slog.LevelWarn
	case LevelError:
		level = slog.LevelError
	}

	// Default to stdout
	writer := opts.Writer
	if writer == nil {
		writer = os.Stdout
	}

	// Create handler options
	handlerOpts := &slog.HandlerOptions{
		Level: level,
	}

	// Choose handler based on format
	var handler slog.Handler
	if opts.JSON {
		// JSON format for production
		handler = slog.NewJSONHandler(writer, handlerOpts)
	} else {
		// Human-readable format for development
		handler = slog.NewTextHandler(writer, handlerOpts)
	}

	return slog.New(handler)
}

// NewFromEnv creates a logger configured from environment variables
func NewFromEnv() *slog.Logger {
	level := LevelInfo
	if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
		level = Level(envLevel)
	}

	json := false
	if os.Getenv("LOG_JSON") == "true" || os.Getenv("ENV") == "production" {
		json = true
	}

	return New(Options{
		Level: level,
		JSON:  json,
	})
}

// Default creates a logger with sensible defaults
func Default() *slog.Logger {
	return New(Options{
		Level: LevelInfo,
		JSON:  false,
	})
}
