package logger

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Environment != "development" {
		t.Errorf("expected environment 'development', got '%s'", cfg.Environment)
	}

	if cfg.Level != "info" {
		t.Errorf("expected level 'info', got '%s'", cfg.Level)
	}

	if len(cfg.OutputPaths) != 1 || cfg.OutputPaths[0] != "stdout" {
		t.Errorf("unexpected output paths: %v", cfg.OutputPaths)
	}
}

func TestNewLogger_Development(t *testing.T) {
	cfg := Config{
		Environment: "development",
		Level:       "debug",
	}

	log, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer log.Sync()

	if log.logger == nil {
		t.Error("logger is nil")
	}

	if log.sugar == nil {
		t.Error("sugar logger is nil")
	}
}

func TestNewLogger_Production(t *testing.T) {
	cfg := Config{
		Environment: "production",
		Level:       "info",
	}

	log, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer log.Sync()

	if log.logger == nil {
		t.Error("logger is nil")
	}
}

func TestNewLogger_InvalidLevel(t *testing.T) {
	cfg := Config{
		Environment: "development",
		Level:       "invalid",
	}

	_, err := NewLogger(cfg)
	if err == nil {
		t.Error("expected error for invalid log level, got nil")
	}
}

func TestWithRequestID(t *testing.T) {
	ctx := context.Background()
	requestID := "req-123"

	ctx = WithRequestID(ctx, requestID)

	retrieved := ctx.Value(requestIDKey)
	if retrieved == nil {
		t.Fatal("request ID not found in context")
	}

	if retrieved.(string) != requestID {
		t.Errorf("expected request ID '%s', got '%s'", requestID, retrieved.(string))
	}
}

func TestWithUserID(t *testing.T) {
	ctx := context.Background()
	userID := "user-456"

	ctx = WithUserID(ctx, userID)

	retrieved := ctx.Value(userIDKey)
	if retrieved == nil {
		t.Fatal("user ID not found in context")
	}

	if retrieved.(string) != userID {
		t.Errorf("expected user ID '%s', got '%s'", userID, retrieved.(string))
	}
}

func TestWithTraceID(t *testing.T) {
	ctx := context.Background()
	traceID := "trace-789"

	ctx = WithTraceID(ctx, traceID)

	retrieved := ctx.Value(traceIDKey)
	if retrieved == nil {
		t.Fatal("trace ID not found in context")
	}

	if retrieved.(string) != traceID {
		t.Errorf("expected trace ID '%s', got '%s'", traceID, retrieved.(string))
	}
}

func TestExtractContextFields(t *testing.T) {
	cfg := DefaultConfig()
	log, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer log.Sync()

	ctx := context.Background()
	ctx = WithRequestID(ctx, "req-123")
	ctx = WithUserID(ctx, "user-456")
	ctx = WithTraceID(ctx, "trace-789")

	fields := log.extractContextFields(ctx)

	if len(fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(fields))
	}

	// Verify field names
	expectedFields := map[string]bool{
		"request_id": false,
		"user_id":    false,
		"trace_id":   false,
	}

	for _, field := range fields {
		expectedFields[field.Key] = true
	}

	for key, found := range expectedFields {
		if !found {
			t.Errorf("expected field '%s' not found", key)
		}
	}
}

func TestWithFields(t *testing.T) {
	cfg := DefaultConfig()
	log, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer log.Sync()

	childLogger := log.WithFields(zap.String("service", "test"))

	if childLogger == nil {
		t.Error("child logger is nil")
	}

	if childLogger.logger == log.logger {
		t.Error("child logger should be a new instance")
	}
}

func TestLoggingMethods(t *testing.T) {
	cfg := Config{
		Environment: "production",
		Level:       "debug",
	}

	log, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}
	defer log.Sync()

	ctx := WithRequestID(context.Background(), "test-req")

	// These should not panic
	log.Debug("debug message")
	log.DebugContext(ctx, "debug message with context")
	log.Info("info message")
	log.InfoContext(ctx, "info message with context")
	log.Warn("warn message")
	log.WarnContext(ctx, "warn message with context")
	log.Error("error message")
	log.ErrorContext(ctx, "error message with context")

	// Sugar logger methods
	log.Debugf("debug %s", "formatted")
	log.Infof("info %s", "formatted")
	log.Warnf("warn %s", "formatted")
	log.Errorf("error %s", "formatted")
}
