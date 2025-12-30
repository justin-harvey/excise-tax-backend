package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Logging logs all incoming HTTP requests with detailed information.
// It logs method, path, status code, duration, IP, and user ID if available.
func Logging(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip logging for health check endpoints
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		// Start timer
		start := time.Now()

		// Get request ID
		requestID := GetRequestID(c)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Get user ID if available from context
		userID := GetUserID(c)

		// Build log fields
		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("query", c.Request.URL.RawQuery),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("response_size", c.Writer.Size()),
		}

		// Add user ID if present
		if userID != "" {
			fields = append(fields, zap.String("user_id", userID))
		}

		// Add error if present
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}

		// Log based on status code
		switch {
		case c.Writer.Status() >= 500:
			logger.Error("HTTP request completed with server error", fields...)
		case c.Writer.Status() >= 400:
			logger.Warn("HTTP request completed with client error", fields...)
		default:
			logger.Info("HTTP request completed", fields...)
		}
	}
}

// GetUserID retrieves the user ID from the Gin context if available.
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}
