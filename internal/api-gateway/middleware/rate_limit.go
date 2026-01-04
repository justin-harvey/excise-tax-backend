package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimitConfig holds the rate limiting configuration
type RateLimitConfig struct {
	RequestsPerWindow int           // Maximum requests allowed per window
	WindowSize        time.Duration // Time window for rate limiting (e.g., 15 minutes)
	BurstSize         int           // Burst capacity for sudden spikes
	EnablePerUser     bool          // Enable per-user rate limiting
	EnablePerIP       bool          // Enable per-IP rate limiting
}

// DefaultRateLimitConfig returns a rate limit configuration matching the spec
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerWindow: 100,
		WindowSize:        15 * time.Minute,
		BurstSize:         20,
		EnablePerUser:     true,
		EnablePerIP:       true,
	}
}

// RateLimiter implements distributed rate limiting using Redis with sliding window algorithm
type RateLimiter struct {
	redis  *redis.Client
	config RateLimitConfig
	logger *zap.Logger
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(redisClient *redis.Client, config RateLimitConfig, logger *zap.Logger) *RateLimiter {
	return &RateLimiter{
		redis:  redisClient,
		config: config,
		logger: logger,
	}
}

// RateLimit returns a middleware that enforces rate limiting
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting for health and metrics endpoints
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		requestID := GetRequestID(c)

		// Determine rate limit keys
		keys := rl.getRateLimitKeys(c)

		// Check rate limits for all applicable keys
		for _, key := range keys {
			allowed, remaining, resetTime, err := rl.checkRateLimit(ctx, key)

			if err != nil {
				rl.logger.Error("Rate limit check failed",
					zap.String("request_id", requestID),
					zap.String("key", key),
					zap.Error(err),
				)
				// On error, allow request but log the issue
				continue
			}

			// Set rate limit headers
			c.Header("X-RateLimit-Limit", strconv.Itoa(rl.config.RequestsPerWindow))
			c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Header("X-RateLimit-Reset", strconv.FormatInt(resetTime.Unix(), 10))

			if !allowed {
				rl.logger.Warn("Rate limit exceeded",
					zap.String("request_id", requestID),
					zap.String("key", key),
					zap.Int("limit", rl.config.RequestsPerWindow),
				)

				c.JSON(http.StatusTooManyRequests, gin.H{
					"error": gin.H{
						"code":        "RATE_LIMIT_EXCEEDED",
						"message":     "Too many requests. Please try again later.",
						"retry_after": resetTime.Unix(),
					},
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// getRateLimitKeys determines which rate limit keys to check based on configuration
func (rl *RateLimiter) getRateLimitKeys(c *gin.Context) []string {
	var keys []string

	// Per-user rate limiting
	if rl.config.EnablePerUser {
		if userID := GetUserID(c); userID != "" {
			keys = append(keys, fmt.Sprintf("rate_limit:user:%s", userID))
		}
	}

	// Per-IP rate limiting
	if rl.config.EnablePerIP {
		clientIP := c.ClientIP()
		keys = append(keys, fmt.Sprintf("rate_limit:ip:%s", clientIP))
	}

	// If no keys were generated, use IP as fallback
	if len(keys) == 0 {
		keys = append(keys, fmt.Sprintf("rate_limit:ip:%s", c.ClientIP()))
	}

	return keys
}

// checkRateLimit uses sliding window algorithm to check and update rate limit
func (rl *RateLimiter) checkRateLimit(ctx context.Context, key string) (allowed bool, remaining int, resetTime time.Time, err error) {
	now := time.Now()
	windowStart := now.Add(-rl.config.WindowSize)

	// Use Redis pipeline for atomic operations
	pipe := rl.redis.Pipeline()

	// Remove old entries outside the sliding window
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart.UnixNano(), 10))

	// Count current requests in window
	countCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})

	// Set expiration on the key
	pipe.Expire(ctx, key, rl.config.WindowSize*2)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, fmt.Errorf("failed to execute rate limit pipeline: %w", err)
	}

	// Get the count
	count, err := countCmd.Result()
	if err != nil {
		return false, 0, time.Time{}, fmt.Errorf("failed to get count: %w", err)
	}

	// Calculate remaining requests
	remaining = rl.config.RequestsPerWindow - int(count)
	if remaining < 0 {
		remaining = 0
	}

	// Calculate reset time (end of current window)
	resetTime = now.Add(rl.config.WindowSize)

	// Check if limit is exceeded (accounting for burst)
	limit := rl.config.RequestsPerWindow + rl.config.BurstSize
	allowed = int(count) <= limit

	return allowed, remaining, resetTime, nil
}

// ResetRateLimit resets the rate limit for a specific key (useful for testing)
func (rl *RateLimiter) ResetRateLimit(ctx context.Context, key string) error {
	return rl.redis.Del(ctx, key).Err()
}
