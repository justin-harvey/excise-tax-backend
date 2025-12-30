package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthConfig holds the authentication configuration
type AuthConfig struct {
	AuthServiceURL    string
	AnonymousRoutes   []string // Routes that don't require authentication
	OptionalAuthRoutes []string // Routes where auth is optional
}

// UserClaims represents the authenticated user's claims
type UserClaims struct {
	UserID    string   `json:"user_id"`
	Email     string   `json:"email"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Role      string   `json:"role"`
	Roles     []string `json:"roles"`
}

// Auth returns a middleware that validates JWT tokens
func Auth(config AuthConfig, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := GetRequestID(c)
		path := c.Request.URL.Path

		// Check if route allows anonymous access
		if isAnonymousRoute(path, config.AnonymousRoutes) {
			c.Next()
			return
		}

		// Extract token from Authorization header
		token, err := extractToken(c)
		if err != nil {
			// Check if auth is optional for this route
			if isOptionalAuthRoute(path, config.OptionalAuthRoutes) {
				c.Next()
				return
			}

			logger.Warn("Missing or invalid authorization header",
				zap.String("request_id", requestID),
				zap.String("path", path),
				zap.Error(err),
			)

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authentication required",
				},
			})
			c.Abort()
			return
		}

		// Validate token by calling auth service
		claims, err := validateToken(c.Request.Context(), token, config.AuthServiceURL, logger)
		if err != nil {
			logger.Warn("Token validation failed",
				zap.String("request_id", requestID),
				zap.String("path", path),
				zap.Error(err),
			)

			c.JSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "Invalid or expired authentication token",
				},
			})
			c.Abort()
			return
		}

		// Add user claims to context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Set("user_roles", claims.Roles)
		c.Set("user_claims", claims)

		logger.Debug("Request authenticated",
			zap.String("request_id", requestID),
			zap.String("user_id", claims.UserID),
			zap.String("role", claims.Role),
		)

		c.Next()
	}
}

// RequireRole returns a middleware that checks if user has required role
func RequireRole(requiredRole string, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := GetRequestID(c)

		// Get user role from context
		userRole, exists := c.Get("user_role")
		if !exists {
			logger.Warn("User role not found in context",
				zap.String("request_id", requestID),
			)

			c.JSON(http.StatusForbidden, gin.H{
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Insufficient permissions",
				},
			})
			c.Abort()
			return
		}

		// Check if user has required role
		role, ok := userRole.(string)
		if !ok || role != requiredRole {
			// Also check roles array for more flexible role checking
			userRoles, rolesExist := c.Get("user_roles")
			if !rolesExist {
				logger.Warn("Access denied - insufficient role",
					zap.String("request_id", requestID),
					zap.String("user_role", role),
					zap.String("required_role", requiredRole),
				)

				c.JSON(http.StatusForbidden, gin.H{
					"error": gin.H{
						"code":    "FORBIDDEN",
						"message": "Insufficient permissions",
					},
				})
				c.Abort()
				return
			}

			// Check if required role is in roles array
			roles, ok := userRoles.([]string)
			if !ok || !containsRole(roles, requiredRole) {
				logger.Warn("Access denied - insufficient role",
					zap.String("request_id", requestID),
					zap.Strings("user_roles", roles),
					zap.String("required_role", requiredRole),
				)

				c.JSON(http.StatusForbidden, gin.H{
					"error": gin.H{
						"code":    "FORBIDDEN",
						"message": "Insufficient permissions",
					},
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// extractToken extracts the JWT token from the Authorization header
func extractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header missing")
	}

	// Check if it's a Bearer token
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", fmt.Errorf("token is empty")
	}

	return token, nil
}

// validateToken validates the JWT token by calling the auth service
func validateToken(ctx context.Context, token string, authServiceURL string, logger *zap.Logger) (*UserClaims, error) {
	// TODO: Implement actual token validation by calling auth service
	// For now, this is a placeholder that should be replaced with actual HTTP call to auth service

	// This would typically:
	// 1. Make POST request to auth service /api/v1/auth/validate endpoint
	// 2. Send token in request body
	// 3. Parse response to get user claims
	// 4. Handle errors (invalid token, expired token, network errors)

	// Placeholder implementation - in production, replace with actual HTTP call
	// Example:
	// req, err := http.NewRequestWithContext(ctx, "POST", authServiceURL+"/api/v1/auth/validate",
	//     strings.NewReader(`{"token":"`+token+`"}`))
	// if err != nil {
	//     return nil, err
	// }
	// resp, err := http.DefaultClient.Do(req)
	// ... parse response into UserClaims

	return nil, fmt.Errorf("token validation not implemented - integrate with auth service")
}

// isAnonymousRoute checks if the path is in the anonymous routes list
func isAnonymousRoute(path string, anonymousRoutes []string) bool {
	for _, route := range anonymousRoutes {
		if matchRoute(path, route) {
			return true
		}
	}
	return false
}

// isOptionalAuthRoute checks if the path is in the optional auth routes list
func isOptionalAuthRoute(path string, optionalRoutes []string) bool {
	for _, route := range optionalRoutes {
		if matchRoute(path, route) {
			return true
		}
	}
	return false
}

// matchRoute checks if a path matches a route pattern (supports wildcards)
func matchRoute(path, pattern string) bool {
	// Exact match
	if path == pattern {
		return true
	}

	// Wildcard match (e.g., /api/v1/auth/*)
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(path, prefix)
	}

	return false
}

// containsRole checks if a role exists in the roles array
func containsRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// GetUserClaims retrieves the user claims from the Gin context
func GetUserClaims(c *gin.Context) (*UserClaims, error) {
	claims, exists := c.Get("user_claims")
	if !exists {
		return nil, fmt.Errorf("user claims not found in context")
	}

	userClaims, ok := claims.(*UserClaims)
	if !ok {
		return nil, fmt.Errorf("invalid user claims type")
	}

	return userClaims, nil
}
