// Package middleware provides HTTP middleware for authentication and authorization.
package middleware

import (
	"net/http"
	"strings"

	"github.com/excise-tax-portal/backend/internal/auth/service"
	"github.com/excise-tax-portal/backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ContextKey represents context key type
type ContextKey string

const (
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// UserEmailKey is the context key for user email
	UserEmailKey ContextKey = "user_email"
	// UserRolesKey is the context key for user roles
	UserRolesKey ContextKey = "user_roles"
	// UserPermissionsKey is the context key for user permissions
	UserPermissionsKey ContextKey = "user_permissions"
	// IsAdminKey is the context key for admin status
	IsAdminKey ContextKey = "is_admin"
	// ClaimsKey is the context key for full JWT claims
	ClaimsKey ContextKey = "claims"
)

// AuthMiddleware creates a Gin middleware for JWT authentication
func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "Authorization header is required",
				},
			})
			c.Abort()
			return
		}

		// Extract JWT token from Bearer format
		token, err := utils.ExtractTokenFromBearer(authHeader)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "Invalid authorization header format",
				},
			})
			c.Abort()
			return
		}

		// Validate token
		claims, err := authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			// Check if token is expired
			if err == utils.ErrExpiredToken {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "TOKEN_EXPIRED",
						"message": "Token has expired",
					},
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "INVALID_TOKEN",
						"message": "Invalid or malformed token",
					},
				})
			}
			c.Abort()
			return
		}

		// Store claims in context
		c.Set(string(UserIDKey), claims.UserID)
		c.Set(string(UserEmailKey), claims.Email)
		c.Set(string(UserRolesKey), claims.Roles)
		c.Set(string(UserPermissionsKey), claims.Permissions)
		c.Set(string(IsAdminKey), claims.IsAdmin)
		c.Set(string(ClaimsKey), claims)

		c.Next()
	}
}

// OptionalAuthMiddleware creates middleware that allows but doesn't require authentication
func OptionalAuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No auth provided, continue without setting user context
			c.Next()
			return
		}

		// Try to extract and validate token
		token, err := utils.ExtractTokenFromBearer(authHeader)
		if err != nil {
			c.Next()
			return
		}

		claims, err := authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.Next()
			return
		}

		// Store claims in context
		c.Set(string(UserIDKey), claims.UserID)
		c.Set(string(UserEmailKey), claims.Email)
		c.Set(string(UserRolesKey), claims.Roles)
		c.Set(string(UserPermissionsKey), claims.Permissions)
		c.Set(string(IsAdminKey), claims.IsAdmin)
		c.Set(string(ClaimsKey), claims)

		c.Next()
	}
}

// GetUserID retrieves user ID from Gin context
func GetUserID(c *gin.Context) (int64, bool) {
	userID, exists := c.Get(string(UserIDKey))
	if !exists {
		return 0, false
	}

	id, ok := userID.(int64)
	return id, ok
}

// GetUserEmail retrieves user email from Gin context
func GetUserEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get(string(UserEmailKey))
	if !exists {
		return "", false
	}

	e, ok := email.(string)
	return e, ok
}

// GetUserRoles retrieves user roles from Gin context
func GetUserRoles(c *gin.Context) ([]string, bool) {
	roles, exists := c.Get(string(UserRolesKey))
	if !exists {
		return nil, false
	}

	r, ok := roles.([]string)
	return r, ok
}

// GetUserPermissions retrieves user permissions from Gin context
func GetUserPermissions(c *gin.Context) ([]string, bool) {
	permissions, exists := c.Get(string(UserPermissionsKey))
	if !exists {
		return nil, false
	}

	p, ok := permissions.([]string)
	return p, ok
}

// IsAdmin checks if the current user is an admin
func IsAdmin(c *gin.Context) bool {
	isAdmin, exists := c.Get(string(IsAdminKey))
	if !exists {
		return false
	}

	admin, ok := isAdmin.(bool)
	return ok && admin
}

// GetClaims retrieves full JWT claims from Gin context
func GetClaims(c *gin.Context) (*utils.Claims, bool) {
	claims, exists := c.Get(string(ClaimsKey))
	if !exists {
		return nil, false
	}

	userClaims, ok := claims.(*utils.Claims)
	return userClaims, ok
}

// HasRole checks if user has a specific role
func HasRole(c *gin.Context, role string) bool {
	roles, ok := GetUserRoles(c)
	if !ok {
		return false
	}

	for _, r := range roles {
		if strings.EqualFold(r, role) {
			return true
		}
	}
	return false
}

// HasPermission checks if user has a specific permission
func HasPermission(c *gin.Context, permission string) bool {
	permissions, ok := GetUserPermissions(c)
	if !ok {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyRole checks if user has any of the specified roles
func HasAnyRole(c *gin.Context, roles ...string) bool {
	for _, role := range roles {
		if HasRole(c, role) {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if user has any of the specified permissions
func HasAnyPermission(c *gin.Context, permissions ...string) bool {
	for _, permission := range permissions {
		if HasPermission(c, permission) {
			return true
		}
	}
	return false
}
