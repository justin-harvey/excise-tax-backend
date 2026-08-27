// Package handler provides HTTP handlers for OAuth2 authentication endpoints.
package handler

import (
	"net/http"
	"time"

	"github.com/excise-tax-portal/backend/internal/auth/service"

	"github.com/gin-gonic/gin"
)

// OAuthHandler handles OAuth authentication HTTP requests
type OAuthHandler struct {
	oauthService *service.OAuthService
	authService  *service.AuthService
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(oauthService *service.OAuthService, authService *service.AuthService) *OAuthHandler {
	return &OAuthHandler{
		oauthService: oauthService,
		authService:  authService,
	}
}

// OAuthAuthorize initiates the OAuth authorization flow
// GET /api/v1/auth/oauth/authorize
func (h *OAuthHandler) OAuthAuthorize(c *gin.Context) {
	// Get optional state from query params
	state := c.Query("state")

	// Generate authorization URL with PKCE
	authURL, err := h.oauthService.GetAuthorizationURL(&service.AuthorizationURLRequest{
		State:   state,
		UsePKCE: true, // Enable PKCE for enhanced security
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "AUTHORIZATION_URL_FAILED",
				"message": "Failed to generate authorization URL",
				"details": err.Error(),
			},
		})
		return
	}

	// Store code verifier in session/cookie for the callback
	// In production, you'd store this in Redis or a secure session
	c.SetCookie(
		"oauth_code_verifier",
		authURL.CodeVerifier,
		300, // 5 minutes
		"/",
		"",
		true, // Secure
		true, // HttpOnly
	)

	// Store state for validation
	c.SetCookie(
		"oauth_state",
		authURL.State,
		300, // 5 minutes
		"/",
		"",
		true, // Secure
		true, // HttpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"authorization_url": authURL.URL,
			"state":             authURL.State,
		},
		"meta": gin.H{
			"timestamp": time.Now().UTC(),
		},
	})
}

// OAuthCallback handles the OAuth callback
// GET /api/v1/auth/oauth/callback
func (h *OAuthHandler) OAuthCallback(c *gin.Context) {
	// Get authorization code and state from query params
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")

	// Check for OAuth errors
	if errorParam != "" {
		errorDesc := c.Query("error_description")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "OAUTH_ERROR",
				"message": errorDesc,
			},
		})
		return
	}

	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_INPUT",
				"message": "Authorization code is missing",
			},
		})
		return
	}

	// Retrieve and validate state
	expectedState, err := c.Cookie("oauth_state")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_STATE",
				"message": "State parameter missing",
			},
		})
		return
	}

	if err := h.oauthService.ValidateState(state, expectedState); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "STATE_MISMATCH",
				"message": err.Error(),
			},
		})
		return
	}

	// Retrieve code verifier
	codeVerifier, err := c.Cookie("oauth_code_verifier")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INVALID_REQUEST",
				"message": "Code verifier missing",
			},
		})
		return
	}

	// Exchange authorization code for access token
	tokenResp, err := h.oauthService.ExchangeCodeForToken(c.Request.Context(), &service.ExchangeCodeRequest{
		Code:         code,
		CodeVerifier: codeVerifier,
		State:        state,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "TOKEN_EXCHANGE_FAILED",
				"message": "Failed to exchange authorization code",
				"details": err.Error(),
			},
		})
		return
	}

	// Get user info from OAuth provider
	userInfo, err := h.oauthService.GetUserInfo(c.Request.Context(), tokenResp.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "USER_INFO_FAILED",
				"message": "Failed to retrieve user information",
				"details": err.Error(),
			},
		})
		return
	}

	// TODO: Create or update user in database using userInfo
	// For now, we'll return the OAuth tokens and user info

	// Clear OAuth cookies
	c.SetCookie("oauth_code_verifier", "", -1, "/", "", true, true)
	c.SetCookie("oauth_state", "", -1, "/", "", true, true)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"oauth_access_token": tokenResp.AccessToken,
			"token_type":         tokenResp.TokenType,
			"expires_in":         tokenResp.ExpiresIn,
			"user_info":          userInfo,
		},
		"meta": gin.H{
			"timestamp": time.Now().UTC(),
		},
	})
}

// OAuthToken exchanges an authorization code for tokens (alternative endpoint)
// POST /api/v1/auth/oauth/token
func (h *OAuthHandler) OAuthToken(c *gin.Context) {
	type TokenRequest struct {
		Code         string `json:"code" binding:"required"`
		CodeVerifier string `json:"code_verifier"`
		State        string `json:"state"`
	}

	var req TokenRequest
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

	// Exchange authorization code for access token
	tokenResp, err := h.oauthService.ExchangeCodeForToken(c.Request.Context(), &service.ExchangeCodeRequest{
		Code:         req.Code,
		CodeVerifier: req.CodeVerifier,
		State:        req.State,
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "TOKEN_EXCHANGE_FAILED",
				"message": "Failed to exchange authorization code",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"access_token":  tokenResp.AccessToken,
			"refresh_token": tokenResp.RefreshToken,
			"token_type":    tokenResp.TokenType,
			"expires_in":    tokenResp.ExpiresIn,
			"id_token":      tokenResp.IDToken,
		},
		"meta": gin.H{
			"timestamp": time.Now().UTC(),
		},
	})
}

// OAuthRefresh refreshes an OAuth access token
// POST /api/v1/auth/oauth/refresh
func (h *OAuthHandler) OAuthRefresh(c *gin.Context) {
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var req RefreshRequest
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

	tokenResp, err := h.oauthService.RefreshAccessToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "REFRESH_FAILED",
				"message": "Failed to refresh access token",
				"details": err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"access_token":  tokenResp.AccessToken,
			"refresh_token": tokenResp.RefreshToken,
			"token_type":    tokenResp.TokenType,
			"expires_in":    tokenResp.ExpiresIn,
		},
		"meta": gin.H{
			"timestamp": time.Now().UTC(),
		},
	})
}
