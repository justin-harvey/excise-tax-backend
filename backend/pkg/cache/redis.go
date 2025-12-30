// Package cache provides Redis client wrapper with generic type support for caching operations.
//
// Example usage:
//
//	config := &cache.Config{
//		Host:     "localhost",
//		Port:     6379,
//		Password: "",
//		DB:       0,
//	}
//
//	client, err := cache.NewRedisClient(ctx, config)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer client.Close()
//
//	// Set a value with TTL
//	err = client.Set(ctx, "user:123", userData, 1*time.Hour)
//
//	// Get a value
//	var user User
//	err = client.Get(ctx, "user:123", &user)
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds the Redis client configuration parameters.
type Config struct {
	Host            string
	Port            int
	Password        string
	DB              int
	MaxRetries      int
	MinIdleConns    int
	PoolSize        int
	ConnMaxIdleTime time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

// DefaultConfig returns a Config with sensible default values.
func DefaultConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            6379,
		Password:        "",
		DB:              0,
		MaxRetries:      3,
		MinIdleConns:    5,
		PoolSize:        10,
		ConnMaxIdleTime: 5 * time.Minute,
		DialTimeout:     5 * time.Second,
		ReadTimeout:     3 * time.Second,
		WriteTimeout:    3 * time.Second,
	}
}

// RedisClient wraps redis.Client to provide additional functionality and type-safe operations.
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new Redis client with the provided configuration.
// It returns an error if the connection cannot be established or if the initial health check fails.
func NewRedisClient(ctx context.Context, cfg *Config) (*RedisClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("redis config cannot be nil")
	}

	client := redis.NewClient(&redis.Options{
		Addr:            fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:        cfg.Password,
		DB:              cfg.DB,
		MaxRetries:      cfg.MaxRetries,
		MinIdleConns:    cfg.MinIdleConns,
		PoolSize:        cfg.PoolSize,
		ConnMaxIdleTime: cfg.ConnMaxIdleTime,
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
	})

	redisClient := &RedisClient{client: client}

	// Perform initial health check
	if err := redisClient.HealthCheck(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("initial health check failed: %w", err)
	}

	return redisClient, nil
}

// HealthCheck verifies that the Redis connection is alive and functional.
func (r *RedisClient) HealthCheck(ctx context.Context) error {
	if r.client == nil {
		return fmt.Errorf("redis client is nil")
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	if result != "PONG" {
		return fmt.Errorf("unexpected ping response: %s", result)
	}

	return nil
}

// Close gracefully closes the Redis client connection.
func (r *RedisClient) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// Set stores a value in Redis with the specified key and TTL.
// The value is automatically serialized to JSON.
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	err = r.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

// Get retrieves a value from Redis and deserializes it into the provided destination.
func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key %s not found", key)
		}
		return fmt.Errorf("failed to get key %s: %w", key, err)
	}

	err = json.Unmarshal(data, dest)
	if err != nil {
		return fmt.Errorf("failed to unmarshal value for key %s: %w", key, err)
	}

	return nil
}

// GetString retrieves a string value from Redis.
func (r *RedisClient) GetString(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", fmt.Errorf("key %s not found", key)
		}
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return val, nil
}

// SetString stores a string value in Redis with the specified key and TTL.
func (r *RedisClient) SetString(ctx context.Context, key string, value string, ttl time.Duration) error {
	err := r.client.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

// Delete removes one or more keys from Redis.
func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	err := r.client.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}

	return nil
}

// Exists checks if a key exists in Redis.
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of key %s: %w", key, err)
	}
	return count > 0, nil
}

// Expire sets a timeout on a key.
func (r *RedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	err := r.client.Expire(ctx, key, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiry on key %s: %w", key, err)
	}
	return nil
}

// TTL returns the remaining time to live of a key.
func (r *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get TTL for key %s: %w", key, err)
	}
	return ttl, nil
}

// Increment increments the integer value of a key by one.
func (r *RedisClient) Increment(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}
	return val, nil
}

// IncrementBy increments the integer value of a key by the given amount.
func (r *RedisClient) IncrementBy(ctx context.Context, key string, value int64) (int64, error) {
	val, err := r.client.IncrBy(ctx, key, value).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s by %d: %w", key, value, err)
	}
	return val, nil
}

// Decrement decrements the integer value of a key by one.
func (r *RedisClient) Decrement(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Decr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to decrement key %s: %w", key, err)
	}
	return val, nil
}

// Publish publishes a message to the specified channel.
func (r *RedisClient) Publish(ctx context.Context, channel string, message interface{}) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = r.client.Publish(ctx, channel, data).Err()
	if err != nil {
		return fmt.Errorf("failed to publish to channel %s: %w", channel, err)
	}

	return nil
}

// Subscribe subscribes to one or more channels and returns a PubSub instance.
// The caller is responsible for closing the PubSub when done.
//
// Example usage:
//
//	pubsub := client.Subscribe(ctx, "notifications")
//	defer pubsub.Close()
//
//	ch := pubsub.Channel()
//	for msg := range ch {
//		fmt.Println(msg.Payload)
//	}
func (r *RedisClient) Subscribe(ctx context.Context, channels ...string) *redis.PubSub {
	return r.client.Subscribe(ctx, channels...)
}

// GetClient returns the underlying redis.Client for advanced operations.
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// FlushDB removes all keys from the current database.
// WARNING: This should only be used in testing environments.
func (r *RedisClient) FlushDB(ctx context.Context) error {
	err := r.client.FlushDB(ctx).Err()
	if err != nil {
		return fmt.Errorf("failed to flush database: %w", err)
	}
	return nil
}
