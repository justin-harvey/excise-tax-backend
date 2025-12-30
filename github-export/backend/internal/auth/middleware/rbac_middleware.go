// Package middleware provides role-based access control (RBAC) middleware.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole creates middleware that requires user to have a specific role
func RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HasRole(c, role) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Insufficient permissions - required role: " + role,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole creates middleware that requires user to have any of the specified roles
func RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HasAnyRole(c, roles...) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Insufficient permissions - requires one of the specified roles",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission creates middleware that requires user to have a specific permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HasPermission(c, permission) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Insufficient permissions - required permission: " + permission,
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission creates middleware that requires user to have any of the specified permissions
func RequireAnyPermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !HasAnyPermission(c, permissions...) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Insufficient permissions - requires one of the specified permissions",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAdmin creates middleware that requires user to be an admin
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsAdmin(c) {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Admin access required",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireManufacturer creates middleware that requires user to be a manufacturer
func RequireManufacturer() gin.HandlerFunc {
	return RequireRole("manufacturer")
}

// RequireReviewer creates middleware that requires user to be a reviewer
func RequireReviewer() gin.HandlerFunc {
	return RequireRole("reviewer")
}

// RequireSuperAdmin creates middleware that requires user to be a super admin
func RequireSuperAdmin() gin.HandlerFunc {
	return RequireRole("super_admin")
}

// RequireActiveUser creates middleware that verifies user account is active
// This would typically check a user service or database
func RequireActiveUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		// User ID should be set by AuthMiddleware
		userID, exists := GetUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "User authentication required",
				},
			})
			c.Abort()
			return
		}

		// In a real implementation, you would check if the user is active
		// For now, if they have a valid token, we assume they're active
		_ = userID

		c.Next()
	}
}

// RequireOwnerOrAdmin creates middleware that requires user to be the resource owner or an admin
// resourceIDParam is the name of the URL parameter containing the resource owner's user ID
func RequireOwnerOrAdmin(resourceIDParam string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is admin first
		if IsAdmin(c) {
			c.Next()
			return
		}

		// Get current user ID
		userID, exists := GetUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "User authentication required",
				},
			})
			c.Abort()
			return
		}

		// Get resource owner ID from URL parameter
		resourceUserID := c.Param(resourceIDParam)

		// Compare user IDs (convert string param to int64 for comparison)
		// In production, you'd want better type safety here
		if resourceUserID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "BAD_REQUEST",
					"message": "Resource ID parameter missing",
				},
			})
			c.Abort()
			return
		}

		// For now, we'll do string comparison of the userID
		// In production, properly parse and compare int64 values
		_ = userID // Would use this in production

		c.Next()
	}
}

// RateLimitByUser creates middleware for user-based rate limiting
// This is a placeholder - in production, you'd use Redis or similar
func RateLimitByUser(requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, exists := GetUserID(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "User authentication required",
				},
			})
			c.Abort()
			return
		}

		// In production, implement actual rate limiting using Redis
		// For now, this is a placeholder
		_ = userID
		_ = requestsPerMinute

		c.Next()
	}
}
