// Package logger provides structured logging with context-aware request tracking.
//
// Example usage:
//
//	// Initialize logger
//	log, err := logger.NewLogger(logger.Config{
//		Environment: "production",
//		Level:       "info",
//	})
//	if err != nil {
//		panic(err)
//	}
//	defer log.Sync()
//
//	// Basic logging
//	log.Info("Application started")
//	log.Error("Failed to process request", zap.Error(err))
//
//	// Context-aware logging
//	ctx := logger.WithRequestID(context.Background(), "req-123")
//	log.InfoContext(ctx, "Processing request")
package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const (
	// requestIDKey is the context key for request ID.
	requestIDKey contextKey = "request_id"
	// userIDKey is the context key for user ID.
	userIDKey contextKey = "user_id"
	// traceIDKey is the context key for trace ID.
	traceIDKey contextKey = "trace_id"
)

// Config holds the logger configuration parameters.
type Config struct {
	// Environment determines the logging format ("development" or "production")
	Environment string
	// Level is the minimum log level (debug, info, warn, error, fatal)
	Level string
	// OutputPaths is a list of paths to write logs to
	OutputPaths []string
	// ErrorOutputPaths is a list of paths to write internal logger errors to
	ErrorOutputPaths []string
}

// DefaultConfig returns a Config with sensible default values.
func DefaultConfig() Config {
	return Config{
		Environment:      "development",
		Level:            "info",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

// Logger wraps zap.Logger to provide additional functionality.
type Logger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// NewLogger creates a new logger with the provided configuration.
func NewLogger(cfg Config) (*Logger, error) {
	var zapConfig zap.Config

	// Configure based on environment
	if cfg.Environment == "production" {
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set log level
	level, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level %s: %w", cfg.Level, err)
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// Set output paths
	if len(cfg.OutputPaths) > 0 {
		zapConfig.OutputPaths = cfg.OutputPaths
	}
	if len(cfg.ErrorOutputPaths) > 0 {
		zapConfig.ErrorOutputPaths = cfg.ErrorOutputPaths
	}

	// Build logger
	baseLogger, err := zapConfig.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return &Logger{
		logger: baseLogger,
		sugar:  baseLogger.Sugar(),
	}, nil
}

// Sync flushes any buffered log entries.
// Applications should call Sync before exiting.
func (l *Logger) Sync() error {
	return l.logger.Sync()
}

// WithRequestID adds a request ID to the context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// WithUserID adds a user ID to the context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// WithTraceID adds a trace ID to the context.
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// extractContextFields extracts logging fields from the context.
func (l *Logger) extractContextFields(ctx context.Context) []zap.Field {
	fields := make([]zap.Field, 0, 3)

	if requestID, ok := ctx.Value(requestIDKey).(string); ok && requestID != "" {
		fields = append(fields, zap.String("request_id", requestID))
	}

	if userID, ok := ctx.Value(userIDKey).(string); ok && userID != "" {
		fields = append(fields, zap.String("user_id", userID))
	}

	if traceID, ok := ctx.Value(traceIDKey).(string); ok && traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	return fields
}

// Debug logs a message at debug level.
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

// DebugContext logs a message at debug level with context fields.
func (l *Logger) DebugContext(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := l.extractContextFields(ctx)
	allFields := append(contextFields, fields...)
	l.logger.Debug(msg, allFields...)
}

// Info logs a message at info level.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// InfoContext logs a message at info level with context fields.
func (l *Logger) InfoContext(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := l.extractContextFields(ctx)
	allFields := append(contextFields, fields...)
	l.logger.Info(msg, allFields...)
}

// Warn logs a message at warn level.
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

// WarnContext logs a message at warn level with context fields.
func (l *Logger) WarnContext(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := l.extractContextFields(ctx)
	allFields := append(contextFields, fields...)
	l.logger.Warn(msg, allFields...)
}

// Error logs a message at error level.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// ErrorContext logs a message at error level with context fields.
func (l *Logger) ErrorContext(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := l.extractContextFields(ctx)
	allFields := append(contextFields, fields...)
	l.logger.Error(msg, allFields...)
}

// Fatal logs a message at fatal level and then calls os.Exit(1).
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

// FatalContext logs a message at fatal level with context fields and then calls os.Exit(1).
func (l *Logger) FatalContext(ctx context.Context, msg string, fields ...zap.Field) {
	contextFields := l.extractContextFields(ctx)
	allFields := append(contextFields, fields...)
	l.logger.Fatal(msg, allFields...)
}

// With creates a child logger with the given fields.
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		logger: l.logger.With(fields...),
		sugar:  l.sugar.With(fields...),
	}
}

// WithFields is an alias for With that provides a more descriptive name.
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return l.With(fields...)
}

// Debugf uses fmt.Sprintf to log a templated message at debug level.
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.sugar.Debugf(template, args...)
}

// Infof uses fmt.Sprintf to log a templated message at info level.
func (l *Logger) Infof(template string, args ...interface{}) {
	l.sugar.Infof(template, args...)
}

// Warnf uses fmt.Sprintf to log a templated message at warn level.
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.sugar.Warnf(template, args...)
}

// Errorf uses fmt.Sprintf to log a templated message at error level.
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.sugar.Errorf(template, args...)
}

// Fatalf uses fmt.Sprintf to log a templated message at fatal level and then calls os.Exit(1).
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.sugar.Fatalf(template, args...)
}

// GetZapLogger returns the underlying zap.Logger for advanced usage.
func (l *Logger) GetZapLogger() *zap.Logger {
	return l.logger
}

// GetSugarLogger returns the underlying zap.SugaredLogger for advanced usage.
func (l *Logger) GetSugarLogger() *zap.SugaredLogger {
	return l.sugar
}
