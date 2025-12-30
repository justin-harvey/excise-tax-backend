package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Recovery recovers from panics and logs the error with stack trace.
// It returns a 500 Internal Server Error response to the client.
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get stack trace
				stack := debug.Stack()

				// Get request ID for correlation
				requestID := GetRequestID(c)

				// Log the panic with full details
				logger.Error("Panic recovered",
					zap.String("request_id", requestID),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()),
					zap.Any("error", err),
					zap.ByteString("stack", stack),
				)

				// Return error response
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": gin.H{
						"code":       "INTERNAL_SERVER_ERROR",
						"message":    "An internal server error occurred",
						"request_id": requestID,
					},
				})

				// Abort further processing
				c.Abort()
			}
		}()

		c.Next()
	}
}

// PanicError is a custom error type for panics
type PanicError struct {
	Value      interface{}
	StackTrace string
}

func (p *PanicError) Error() string {
	return fmt.Sprintf("panic: %v", p.Value)
}
