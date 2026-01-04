package database

import (
	"context"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Host != "localhost" {
		t.Errorf("expected host 'localhost', got '%s'", cfg.Host)
	}

	if cfg.Port != 5432 {
		t.Errorf("expected port 5432, got %d", cfg.Port)
	}

	if cfg.MaxConns != 25 {
		t.Errorf("expected MaxConns 25, got %d", cfg.MaxConns)
	}

	if cfg.MinConns != 5 {
		t.Errorf("expected MinConns 5, got %d", cfg.MinConns)
	}
}

func TestNewPostgresDB_NilConfig(t *testing.T) {
	ctx := context.Background()
	_, err := NewPostgresDB(ctx, nil)

	if err == nil {
		t.Error("expected error when config is nil, got nil")
	}

	if err.Error() != "database config cannot be nil" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestNewPostgresDB_InvalidConfig(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := &Config{
		Host:     "invalid-host-that-does-not-exist",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		DBName:   "testdb",
		SSLMode:  "disable",
		MaxConns: 5,
		MinConns: 1,
	}

	_, err := NewPostgresDB(ctx, cfg)

	if err == nil {
		t.Error("expected error when connecting to invalid host, got nil")
	}
}

func TestPostgresDB_HealthCheck_NilPool(t *testing.T) {
	db := &PostgresDB{Pool: nil}
	ctx := context.Background()

	err := db.HealthCheck(ctx)

	if err == nil {
		t.Error("expected error when pool is nil, got nil")
	}

	if err.Error() != "database pool is nil" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestPostgresDB_Close_NilPool(t *testing.T) {
	db := &PostgresDB{Pool: nil}

	// Should not panic
	db.Close()
}

// Integration tests would require a real database connection
// These should be run with build tags or in a separate test suite
// Example:
// func TestNewPostgresDB_Integration(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("skipping integration test")
// 	}
//
// 	cfg := &Config{
// 		Host:     "localhost",
// 		Port:     5432,
// 		User:     "postgres",
// 		Password: "password",
// 		DBName:   "testdb",
// 		SSLMode:  "disable",
// 	}
//
// 	ctx := context.Background()
// 	db, err := NewPostgresDB(ctx, cfg)
// 	if err != nil {
// 		t.Fatalf("failed to connect to database: %v", err)
// 	}
// 	defer db.Close()
//
// 	err = db.HealthCheck(ctx)
// 	if err != nil {
// 		t.Errorf("health check failed: %v", err)
// 	}
// }
