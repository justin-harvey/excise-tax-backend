// Package middleware provides HTTP middleware for the API Gateway.
package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"
	// RequestIDKey is the context key for request ID
	RequestIDKey = "request_id"
)

// RequestID generates a unique request ID for each request and adds it to the context.
// If a request ID is already present in the X-Request-ID header, it will be used instead.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID already exists in header
		requestID := c.GetHeader(RequestIDHeader)

		// Generate new UUID if not present
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add to context for downstream use
		c.Set(RequestIDKey, requestID)

		// Add to response header
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

// GetRequestID retrieves the request ID from the Gin context.
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
