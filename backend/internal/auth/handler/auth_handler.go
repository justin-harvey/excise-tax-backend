// Package handler provides HTTP handlers for authentication endpoints.
package handler

import (
	"net/http"
	"time"

	"github.com/excise-tax-portal/backend/internal/auth/middleware"
	"github.com/excise-tax-portal/backend/internal/auth/model"
	"github.com/excise-tax-portal/backend/internal/auth/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email        string               `json:"email" binding:"required,email"`
	Password     string               `json:"password" binding:"required,min=8"`
	FirstName    string               `json:"first_name" binding:"required"`
	LastName     string               `json:"last_name" binding:"required"`
	Phone        string               `json:"phone"`
	Role         string               `json:"role"`
	Manufacturer *ManufacturerRequest `json:"manufacturer,omitempty"`
}

// ManufacturerRequest represents manufacturer registration data
type ManufacturerRequest struct {
	CompanyName   string `json:"company_name" binding:"required"`
	TaxID         string `json:"tax_id" binding:"required"`
	LicenseNumber string `json:"license_number" binding:"required"`
	BusinessType  string `json:"business_type" binding:"required"`
	AddressLine1  string `json:"address_line1" binding:"required"`
	AddressLine2  string `json:"address_line2"`
	City          string `json:"city" binding:"required"`
	State         string `json:"state" binding:"required,len=2"`
	ZipCode       string `json:"zip_code" binding:"required"`
	Phone         string `json:"phone" binding:"required"`
	Website       string `json:"website"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// RefreshTokenRequest represents the refresh token request body
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// ChangePasswordRequest represents the change password request body
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// ForgotPasswordRequest represents the forgot password request body
type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// ResetPasswordRequest represents the reset password request body
type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// Register handles user registration
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	// Convert manufacturer request to model
	var manufacturer *model.Manufacturer
	if req.Manufacturer != nil {
		manufacturer = &model.Manufacturer{
			CompanyName:   req.Manufacturer.CompanyName,
			TaxID:         req.Manufacturer.TaxID,
			LicenseNumber: req.Manufacturer.LicenseNumber,
			BusinessType:  req.Manufacturer.BusinessType,
			AddressLine1:  req.Manufacturer.AddressLine1,
			AddressLine2:  req.Manufacturer.AddressLine2,
			City:          req.Manufacturer.City,
			State:         req.Manufacturer.State,
			ZipCode:       req.Manufacturer.ZipCode,
			Phone:         req.Manufacturer.Phone,
			Website:       req.Manufacturer.Website,
			IsApproved:    false, // Requires admin approval
		}
	}

	user, err := h.authService.Register(c.Request.Context(), &service.RegisterRequest{
		Email:        req.Email,
		Password:     req.Password,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		Role:         req.Role,
		Manufacturer: manufacturer,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "REGISTRATION_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"user": user.PublicUser(),
		},
		"meta": gin.H{
			"timestamp": time.Now().UTC(),
		},
	})
}

// Login handles user login
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	tokenPair, user, err := h.authService.Login(c.Request.Context(), &service.LoginRequest{
		Email:     req.Email,
		Password:  req.Password,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "AUTHENTICATION_FAILED",
				"message": "Invalid email or password",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"access_token":  tokenPair.AccessToken,
			"refresh_token": tokenPair.RefreshToken,
			"token_type":    tokenPair.TokenType,
			"expires_in":    tokenPair.ExpiresIn,
			"user":          user.PublicUser(),
		},
		"meta": gin.H{
			"timestamp": time.Now().UTC(),
		},
	})
}

// Logout handles user logout
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get session ID from request (you might store this differently)
	// For now, we'll use the user ID to logout all sessions
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	// Logout from all devices
	if err := h.authService.LogoutAll(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "LOGOUT_FAILED",
				"message": "Failed to logout",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Successfully logged out from all devices",
		},
		"meta": gin.H{
			"timestamp": time.Now().UTC(),
		},
	})
}

// RefreshToken handles token refresh
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	tokenPair, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "REFRESH_FAILED",
				"message": "Invalid or expired refresh token",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"access_token":  tokenPair.AccessToken,
			"refresh_token": tokenPair.RefreshToken,
			"token_type":    tokenPair.TokenType,
			"expires_in":    tokenPair.ExpiresIn,
		},
		"meta": gin.H{
			"timestamp": time.Now().UTC(),
		},
	})
}

// GetCurrentUser returns current user information
// GET /api/v1/auth/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	userWithManufacturer, err := h.authService.GetCurrentUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to retrieve user information",
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"user":         userWithManufacturer.User.PublicUser(),
			"manufacturer": userWithManufacturer.Manufacturer,
		},
		"meta": gin.H{
			"timestamp": time.Now().UTC(),
		},
	})
}

// ChangePassword handles password change
// POST /api/v1/auth/change-password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "UNAUTHORIZED",
				"message": "User not authenticated",
			},
		})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	err := h.authService.ChangePassword(c.Request.Context(), &service.ChangePasswordRequest{
		UserID:      userID,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "PASSWORD_CHANGE_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Password changed successfully. Please login again.",
		},
		"meta": gin.H{
			"timestamp": time.Now().UTC(),
		},
	})
}

// ForgotPassword handles password reset request
// POST /api/v1/auth/forgot-password
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	// TODO: Implement password reset token generation and email sending
	// For now, return success

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Password reset instructions sent to your email",
		},
		"meta": gin.H{
			"timestamp": time.Now().UTC(),
		},
	})
}

// ResetPassword handles password reset with token
// POST /api/v1/auth/reset-password
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": "Invalid request body",
				"details": err.Error(),
			},
		})
		return
	}

	err := h.authService.ResetPassword(c.Request.Context(), req.Email, req.NewPassword, req.Token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "PASSWORD_RESET_FAILED",
				"message": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"message": "Password reset successfully. Please login with your new password.",
		},
		"meta": gin.H{
			"timestamp": time.Now().UTC(),
		},
	})
}
