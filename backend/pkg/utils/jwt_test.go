package utils

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestNewJWTManager(t *testing.T) {
	config := &JWTConfig{
		SecretKey:            "test-secret-key-that-is-long-enough",
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "test-issuer",
	}

	manager, err := NewJWTManager(config)
	if err != nil {
		t.Fatalf("failed to create JWT manager: %v", err)
	}

	if manager == nil {
		t.Error("expected manager to be created")
	}

	if manager.config != config {
		t.Error("expected config to be set")
	}
}

func TestNewJWTManager_NilConfig(t *testing.T) {
	_, err := NewJWTManager(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
}

func TestNewJWTManager_EmptySecretKey(t *testing.T) {
	config := &JWTConfig{
		SecretKey: "",
	}

	_, err := NewJWTManager(config)
	if err == nil {
		t.Error("expected error for empty secret key")
	}
}

func TestGenerateAccessToken(t *testing.T) {
	config := &JWTConfig{
		SecretKey:            "test-secret-key-that-is-long-enough",
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "test-issuer",
	}

	manager, err := NewJWTManager(config)
	if err != nil {
		t.Fatalf("failed to create JWT manager: %v", err)
	}

	claims := &Claims{
		UserID:      123,
		Email:       "test@example.com",
		Roles:       []string{"user"},
		Permissions: []string{"read"},
		IsAdmin:     false,
		StateID:     "CA",
	}

	token, err := manager.GenerateAccessToken(claims)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	if token == "" {
		t.Error("expected token to be generated")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	config := &JWTConfig{
		SecretKey:            "test-secret-key-that-is-long-enough",
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "test-issuer",
	}

	manager, err := NewJWTManager(config)
	if err != nil {
		t.Fatalf("failed to create JWT manager: %v", err)
	}

	claims := &Claims{
		UserID: 123,
		Email:  "test@example.com",
	}

	token, err := manager.GenerateRefreshToken(claims)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	if token == "" {
		t.Error("expected token to be generated")
	}
}

func TestValidateAccessToken(t *testing.T) {
	config := &JWTConfig{
		SecretKey:            "test-secret-key-that-is-long-enough",
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "test-issuer",
	}

	manager, err := NewJWTManager(config)
	if err != nil {
		t.Fatalf("failed to create JWT manager: %v", err)
	}

	claims := &Claims{
		UserID:      123,
		Email:       "test@example.com",
		Roles:       []string{"user", "admin"},
		Permissions: []string{"read", "write"},
		IsAdmin:     true,
		StateID:     "CA",
	}

	token, err := manager.GenerateAccessToken(claims)
	if err != nil {
		t.Fatalf("failed to generate access token: %v", err)
	}

	validatedClaims, err := manager.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}

	if validatedClaims.UserID != claims.UserID {
		t.Errorf("expected user ID %d, got %d", claims.UserID, validatedClaims.UserID)
	}

	if validatedClaims.Email != claims.Email {
		t.Errorf("expected email %s, got %s", claims.Email, validatedClaims.Email)
	}

	if validatedClaims.TokenType != "access" {
		t.Errorf("expected token type 'access', got '%s'", validatedClaims.TokenType)
	}
}

func TestValidateRefreshToken(t *testing.T) {
	config := &JWTConfig{
		SecretKey:            "test-secret-key-that-is-long-enough",
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "test-issuer",
	}

	manager, err := NewJWTManager(config)
	if err != nil {
		t.Fatalf("failed to create JWT manager: %v", err)
	}

	claims := &Claims{
		UserID: 123,
		Email:  "test@example.com",
	}

	token, err := manager.GenerateRefreshToken(claims)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	validatedClaims, err := manager.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("failed to validate refresh token: %v", err)
	}

	if validatedClaims.TokenType != "refresh" {
		t.Errorf("expected token type 'refresh', got '%s'", validatedClaims.TokenType)
	}
}

func TestValidateAccessToken_WithRefreshToken(t *testing.T) {
	config := &JWTConfig{
		SecretKey:            "test-secret-key-that-is-long-enough",
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "test-issuer",
	}

	manager, err := NewJWTManager(config)
	if err != nil {
		t.Fatalf("failed to create JWT manager: %v", err)
	}

	claims := &Claims{
		UserID: 123,
		Email:  "test@example.com",
	}

	refreshToken, err := manager.GenerateRefreshToken(claims)
	if err != nil {
		t.Fatalf("failed to generate refresh token: %v", err)
	}

	// Should fail because it's a refresh token, not an access token
	_, err = manager.ValidateAccessToken(refreshToken)
	if err == nil {
		t.Error("expected error when validating refresh token as access token")
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	config := &JWTConfig{
		SecretKey:            "test-secret-key-that-is-long-enough",
		AccessTokenDuration:  15 * time.Minute,
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "test-issuer",
	}

	manager, err := NewJWTManager(config)
	if err != nil {
		t.Fatalf("failed to create JWT manager: %v", err)
	}

	_, err = manager.ValidateToken("invalid.token.here")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	config := &JWTConfig{
		SecretKey:            "test-secret-key-that-is-long-enough",
		AccessTokenDuration:  -1 * time.Hour, // Already expired
		RefreshTokenDuration: 7 * 24 * time.Hour,
		Issuer:               "test-issuer",
	}

	manager, err := NewJWTManager(config)
	if err != nil {
		t.Fatalf("failed to create JWT manager: %v", err)
	}

	claims := &Claims{
		UserID: 123,
		Email:  "test@example.com",
	}

	token, err := manager.GenerateAccessToken(claims)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	_, err = manager.ValidateToken(token)
	if err != ErrExpiredToken {
		t.Errorf("expected ErrExpiredToken, got: %v", err)
	}
}

func TestExtractTokenFromBearer(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		{
			name:        "valid bearer token",
			input:       "Bearer abc123",
			expected:    "abc123",
			shouldError: false,
		},
		{
			name:        "empty string",
			input:       "",
			expected:    "",
			shouldError: true,
		},
		{
			name:        "missing bearer prefix",
			input:       "abc123",
			expected:    "",
			shouldError: true,
		},
		{
			name:        "bearer without token",
			input:       "Bearer ",
			expected:    "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := ExtractTokenFromBearer(tt.input)

			if tt.shouldError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if token != tt.expected {
				t.Errorf("expected token '%s', got '%s'", tt.expected, token)
			}
		})
	}
}

func TestClaims_HasRole(t *testing.T) {
	claims := &Claims{
		Roles: []string{"user", "admin"},
	}

	if !claims.HasRole("user") {
		t.Error("expected HasRole to return true for 'user'")
	}

	if !claims.HasRole("admin") {
		t.Error("expected HasRole to return true for 'admin'")
	}

	if claims.HasRole("superuser") {
		t.Error("expected HasRole to return false for 'superuser'")
	}
}

func TestClaims_HasPermission(t *testing.T) {
	claims := &Claims{
		Permissions: []string{"read", "write"},
	}

	if !claims.HasPermission("read") {
		t.Error("expected HasPermission to return true for 'read'")
	}

	if claims.HasPermission("delete") {
		t.Error("expected HasPermission to return false for 'delete'")
	}
}

func TestClaims_HasAnyRole(t *testing.T) {
	claims := &Claims{
		Roles: []string{"user"},
	}

	if !claims.HasAnyRole("user", "admin") {
		t.Error("expected HasAnyRole to return true")
	}

	if claims.HasAnyRole("admin", "superuser") {
		t.Error("expected HasAnyRole to return false")
	}
}

func TestClaims_HasAllRoles(t *testing.T) {
	claims := &Claims{
		Roles: []string{"user", "admin"},
	}

	if !claims.HasAllRoles("user", "admin") {
		t.Error("expected HasAllRoles to return true")
	}

	if claims.HasAllRoles("user", "admin", "superuser") {
		t.Error("expected HasAllRoles to return false")
	}
}

func TestClaims_IsExpired(t *testing.T) {
	// jwt/v5's RegisteredClaims.ExpiresAt is *jwt.NumericDate, not *time.Time -
	// this test predates the v5 upgrade and used to construct it as a raw
	// *time.Time, which no longer compiles (go vet caught it).
	claims := &Claims{}

	// Not expired
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(1 * time.Hour))

	if claims.IsExpired() {
		t.Error("expected IsExpired to return false for future expiration")
	}

	// Expired
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-1 * time.Hour))

	if !claims.IsExpired() {
		t.Error("expected IsExpired to return true for past expiration")
	}
}

func TestDefaultJWTConfig(t *testing.T) {
	config := DefaultJWTConfig()

	if config.AccessTokenDuration != 1*time.Hour {
		t.Errorf("expected access token duration 1h, got %v", config.AccessTokenDuration)
	}

	if config.RefreshTokenDuration != 7*24*time.Hour {
		t.Errorf("expected refresh token duration 168h, got %v", config.RefreshTokenDuration)
	}

	if config.Issuer != "excise-tax-portal" {
		t.Errorf("expected issuer 'excise-tax-portal', got '%s'", config.Issuer)
	}
}
