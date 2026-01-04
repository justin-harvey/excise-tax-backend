// Package utils provides cryptographic utility functions for password hashing,
// token generation, and secure random string generation.
package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultBcryptCost is the default cost for bcrypt hashing (12 is recommended for production)
	DefaultBcryptCost = 12
	// MinPasswordLength is the minimum password length
	MinPasswordLength = 8
	// MaxPasswordLength is the maximum password length
	MaxPasswordLength = 128
)

var (
	// ErrPasswordTooShort is returned when password is too short
	ErrPasswordTooShort = errors.New("password must be at least 8 characters long")
	// ErrPasswordTooLong is returned when password is too long
	ErrPasswordTooLong = errors.New("password exceeds maximum length")
	// ErrInvalidPassword is returned when password validation fails
	ErrInvalidPassword = errors.New("invalid password")
)

// HashPassword hashes a password using bcrypt with the default cost
func HashPassword(password string) (string, error) {
	return HashPasswordWithCost(password, DefaultBcryptCost)
}

// HashPasswordWithCost hashes a password using bcrypt with a custom cost
func HashPasswordWithCost(password string, cost int) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// ComparePasswords compares a hashed password with a plain text password
func ComparePasswords(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrInvalidPassword
		}
		return fmt.Errorf("failed to compare passwords: %w", err)
	}
	return nil
}

// ValidatePassword validates password meets minimum requirements
func ValidatePassword(password string) error {
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	if len(password) > MaxPasswordLength {
		return ErrPasswordTooLong
	}
	return nil
}

// GenerateRandomString generates a cryptographically secure random string of the specified length
func GenerateRandomString(length int) (string, error) {
	if length <= 0 {
		return "", errors.New("length must be greater than 0")
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}

// GenerateRandomBytes generates cryptographically secure random bytes
func GenerateRandomBytes(length int) ([]byte, error) {
	if length <= 0 {
		return nil, errors.New("length must be greater than 0")
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return bytes, nil
}

// GenerateSecureToken generates a cryptographically secure token of the specified byte length
// The token is returned as a base64-encoded string
func GenerateSecureToken(byteLength int) (string, error) {
	bytes, err := GenerateRandomBytes(byteLength)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateResetToken generates a password reset token (32 bytes = 256 bits)
func GenerateResetToken() (string, error) {
	return GenerateSecureToken(32)
}

// GenerateSessionID generates a unique session ID (24 bytes = 192 bits)
func GenerateSessionID() (string, error) {
	return GenerateSecureToken(24)
}

// HashSHA256 creates a SHA256 hash of the input string
func HashSHA256(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// GenerateCodeVerifier generates a PKCE code verifier (for OAuth2)
// Returns a base64url-encoded string of 32 random bytes (256 bits)
func GenerateCodeVerifier() (string, error) {
	bytes, err := GenerateRandomBytes(32)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// GenerateCodeChallenge generates a PKCE code challenge from a code verifier
// Uses SHA256 hash method
func GenerateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// GenerateStateToken generates a secure state token for OAuth2 flows
func GenerateStateToken() (string, error) {
	return GenerateSecureToken(24)
}

// PasswordStrength represents the strength level of a password
type PasswordStrength int

const (
	PasswordWeak PasswordStrength = iota
	PasswordModerate
	PasswordStrong
	PasswordVeryStrong
)

// CheckPasswordStrength evaluates password strength
func CheckPasswordStrength(password string) PasswordStrength {
	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
		hasSpecial bool
	)

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasNumber = true
		default:
			hasSpecial = true
		}
	}

	length := len(password)
	criteriaCount := 0

	if hasUpper {
		criteriaCount++
	}
	if hasLower {
		criteriaCount++
	}
	if hasNumber {
		criteriaCount++
	}
	if hasSpecial {
		criteriaCount++
	}

	// Evaluate strength
	switch {
	case length >= 16 && criteriaCount >= 3:
		return PasswordVeryStrong
	case length >= 12 && criteriaCount >= 3:
		return PasswordStrong
	case length >= 8 && criteriaCount >= 2:
		return PasswordModerate
	default:
		return PasswordWeak
	}
}

// String returns a string representation of password strength
func (s PasswordStrength) String() string {
	switch s {
	case PasswordWeak:
		return "weak"
	case PasswordModerate:
		return "moderate"
	case PasswordStrong:
		return "strong"
	case PasswordVeryStrong:
		return "very_strong"
	default:
		return "unknown"
	}
}
