// Package utils provides utility functions for JWT token generation, validation,
// and cryptographic operations used throughout the application.
package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// ErrInvalidToken is returned when token validation fails
	ErrInvalidToken = errors.New("invalid token")
	// ErrExpiredToken is returned when token has expired
	ErrExpiredToken = errors.New("token has expired")
	// ErrTokenNotValid is returned when token signature is invalid
	ErrTokenNotValid = errors.New("token signature invalid")
)

// JWTConfig holds JWT configuration parameters
type JWTConfig struct {
	SecretKey            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	Issuer               string
}

// DefaultJWTConfig returns a JWTConfig with default values
func DefaultJWTConfig() *JWTConfig {
	return &JWTConfig{
		AccessTokenDuration:  1 * time.Hour,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "excise-tax-portal",
	}
}

// Claims represents JWT claims structure with custom fields
type Claims struct {
	UserID      int64    `json:"user_id"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	IsAdmin     bool     `json:"is_admin"`
	StateID     string   `json:"state_id,omitempty"`
	TokenType   string   `json:"token_type"` // "access" or "refresh"
	jwt.RegisteredClaims
}

// JWTManager handles JWT token operations
type JWTManager struct {
	config *JWTConfig
}

// NewJWTManager creates a new JWT manager with the provided configuration
func NewJWTManager(config *JWTConfig) (*JWTManager, error) {
	if config == nil {
		return nil, errors.New("JWT config cannot be nil")
	}
	if config.SecretKey == "" {
		return nil, errors.New("JWT secret key cannot be empty")
	}
	return &JWTManager{config: config}, nil
}

// GenerateAccessToken generates a new JWT access token with the provided claims
func (m *JWTManager) GenerateAccessToken(claims *Claims) (string, error) {
	now := time.Now()

	claims.TokenType = "access"
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(m.config.AccessTokenDuration)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    m.config.Issuer,
		Subject:   fmt.Sprintf("%d", claims.UserID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(m.config.SecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a new JWT refresh token with the provided claims
func (m *JWTManager) GenerateRefreshToken(claims *Claims) (string, error) {
	now := time.Now()

	// Refresh tokens have minimal claims
	refreshClaims := &Claims{
		UserID:    claims.UserID,
		Email:     claims.Email,
		TokenType: "refresh",
	}

	refreshClaims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(now.Add(m.config.RefreshTokenDuration)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    m.config.Issuer,
		Subject:   fmt.Sprintf("%d", claims.UserID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	tokenString, err := token.SignedString([]byte(m.config.SecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenNotValid
	}

	return claims, nil
}

// ValidateAccessToken validates an access token and ensures it's the correct type
func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "access" {
		return nil, errors.New("token is not an access token")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token and ensures it's the correct type
func (m *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != "refresh" {
		return nil, errors.New("token is not a refresh token")
	}

	return claims, nil
}

// ExtractTokenFromBearer extracts the JWT token from a Bearer authorization header
func ExtractTokenFromBearer(bearerToken string) (string, error) {
	if bearerToken == "" {
		return "", errors.New("authorization header is empty")
	}

	const prefix = "Bearer "
	if len(bearerToken) < len(prefix) {
		return "", errors.New("invalid authorization header format")
	}

	if bearerToken[:len(prefix)] != prefix {
		return "", errors.New("authorization header must start with 'Bearer '")
	}

	token := bearerToken[len(prefix):]
	if token == "" {
		return "", errors.New("token is empty")
	}

	return token, nil
}

// IsExpired checks if the token has expired
func (c *Claims) IsExpired() bool {
	if c.ExpiresAt == nil {
		return true
	}
	return time.Now().After(c.ExpiresAt.Time)
}

// HasRole checks if the claims contain a specific role
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// HasPermission checks if the claims contain a specific permission
func (c *Claims) HasPermission(permission string) bool {
	for _, p := range c.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyRole checks if the claims contain any of the specified roles
func (c *Claims) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}

// HasAllRoles checks if the claims contain all of the specified roles
func (c *Claims) HasAllRoles(roles ...string) bool {
	for _, role := range roles {
		if !c.HasRole(role) {
			return false
		}
	}
	return true
}
