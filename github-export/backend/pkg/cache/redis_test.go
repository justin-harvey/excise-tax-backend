package cache

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

	if cfg.Port != 6379 {
		t.Errorf("expected port 6379, got %d", cfg.Port)
	}

	if cfg.MaxRetries != 3 {
		t.Errorf("expected MaxRetries 3, got %d", cfg.MaxRetries)
	}

	if cfg.PoolSize != 10 {
		t.Errorf("expected PoolSize 10, got %d", cfg.PoolSize)
	}
}

func TestNewRedisClient_NilConfig(t *testing.T) {
	ctx := context.Background()
	_, err := NewRedisClient(ctx, nil)

	if err == nil {
		t.Error("expected error when config is nil, got nil")
	}

	if err.Error() != "redis config cannot be nil" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestNewRedisClient_InvalidConfig(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg := &Config{
		Host:         "invalid-host-that-does-not-exist",
		Port:         6379,
		Password:     "",
		DB:           0,
		DialTimeout:  1 * time.Second,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	_, err := NewRedisClient(ctx, cfg)

	if err == nil {
		t.Error("expected error when connecting to invalid host, got nil")
	}
}

func TestRedisClient_HealthCheck_NilClient(t *testing.T) {
	client := &RedisClient{client: nil}
	ctx := context.Background()

	err := client.HealthCheck(ctx)

	if err == nil {
		t.Error("expected error when client is nil, got nil")
	}

	if err.Error() != "redis client is nil" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestRedisClient_Close_NilClient(t *testing.T) {
	client := &RedisClient{client: nil}

	// Should not panic
	err := client.Close()
	if err != nil {
		t.Errorf("expected no error when closing nil client, got: %v", err)
	}
}

// Integration tests would require a real Redis connection
// These should be run with build tags or in a separate test suite
// Example:
// func TestRedisClient_SetGet_Integration(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("skipping integration test")
// 	}
//
// 	cfg := &Config{
// 		Host: "localhost",
// 		Port: 6379,
// 	}
//
// 	ctx := context.Background()
// 	client, err := NewRedisClient(ctx, cfg)
// 	if err != nil {
// 		t.Fatalf("failed to connect to Redis: %v", err)
// 	}
// 	defer client.Close()
//
// 	// Test Set and Get
// 	type testData struct {
// 		Name  string
// 		Value int
// 	}
//
// 	data := testData{Name: "test", Value: 123}
// 	err = client.Set(ctx, "test:key", data, 1*time.Minute)
// 	if err != nil {
// 		t.Errorf("failed to set value: %v", err)
// 	}
//
// 	var retrieved testData
// 	err = client.Get(ctx, "test:key", &retrieved)
// 	if err != nil {
// 		t.Errorf("failed to get value: %v", err)
// 	}
//
// 	if retrieved.Name != data.Name || retrieved.Value != data.Value {
// 		t.Errorf("expected %+v, got %+v", data, retrieved)
// 	}
//
// 	// Cleanup
// 	client.Delete(ctx, "test:key")
// }
