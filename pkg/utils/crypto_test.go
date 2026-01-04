package utils

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "mySecurePassword123"

	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if hashed == "" {
		t.Error("expected hashed password to be non-empty")
	}

	if hashed == password {
		t.Error("hashed password should not equal plain password")
	}
}

func TestHashPassword_TooShort(t *testing.T) {
	password := "short"

	_, err := HashPassword(password)
	if err != ErrPasswordTooShort {
		t.Errorf("expected ErrPasswordTooShort, got: %v", err)
	}
}

func TestHashPassword_TooLong(t *testing.T) {
	password := strings.Repeat("a", MaxPasswordLength+1)

	_, err := HashPassword(password)
	if err != ErrPasswordTooLong {
		t.Errorf("expected ErrPasswordTooLong, got: %v", err)
	}
}

func TestComparePasswords_Success(t *testing.T) {
	password := "mySecurePassword123"

	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	err = ComparePasswords(hashed, password)
	if err != nil {
		t.Errorf("expected passwords to match, got error: %v", err)
	}
}

func TestComparePasswords_Mismatch(t *testing.T) {
	password := "mySecurePassword123"
	wrongPassword := "wrongPassword456"

	hashed, err := HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	err = ComparePasswords(hashed, wrongPassword)
	if err != ErrInvalidPassword {
		t.Errorf("expected ErrInvalidPassword, got: %v", err)
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectedErr error
	}{
		{
			name:        "valid password",
			password:    "validPass123",
			expectedErr: nil,
		},
		{
			name:        "too short",
			password:    "short",
			expectedErr: ErrPasswordTooShort,
		},
		{
			name:        "too long",
			password:    strings.Repeat("a", MaxPasswordLength+1),
			expectedErr: ErrPasswordTooLong,
		},
		{
			name:        "minimum length",
			password:    "12345678",
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if err != tt.expectedErr {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	length := 32

	str1, err := GenerateRandomString(length)
	if err != nil {
		t.Fatalf("failed to generate random string: %v", err)
	}

	str2, err := GenerateRandomString(length)
	if err != nil {
		t.Fatalf("failed to generate random string: %v", err)
	}

	if str1 == str2 {
		t.Error("expected two random strings to be different")
	}

	if len(str1) != length {
		t.Errorf("expected string length %d, got %d", length, len(str1))
	}
}

func TestGenerateRandomString_InvalidLength(t *testing.T) {
	_, err := GenerateRandomString(0)
	if err == nil {
		t.Error("expected error for zero length")
	}

	_, err = GenerateRandomString(-1)
	if err == nil {
		t.Error("expected error for negative length")
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	length := 32

	bytes1, err := GenerateRandomBytes(length)
	if err != nil {
		t.Fatalf("failed to generate random bytes: %v", err)
	}

	bytes2, err := GenerateRandomBytes(length)
	if err != nil {
		t.Fatalf("failed to generate random bytes: %v", err)
	}

	if len(bytes1) != length {
		t.Errorf("expected byte slice length %d, got %d", length, len(bytes1))
	}

	// Check that they're different
	same := true
	for i := 0; i < length; i++ {
		if bytes1[i] != bytes2[i] {
			same = false
			break
		}
	}

	if same {
		t.Error("expected two random byte slices to be different")
	}
}

func TestGenerateSecureToken(t *testing.T) {
	token1, err := GenerateSecureToken(32)
	if err != nil {
		t.Fatalf("failed to generate secure token: %v", err)
	}

	token2, err := GenerateSecureToken(32)
	if err != nil {
		t.Fatalf("failed to generate secure token: %v", err)
	}

	if token1 == token2 {
		t.Error("expected two tokens to be different")
	}

	if token1 == "" {
		t.Error("expected token to be non-empty")
	}
}

func TestGenerateResetToken(t *testing.T) {
	token, err := GenerateResetToken()
	if err != nil {
		t.Fatalf("failed to generate reset token: %v", err)
	}

	if token == "" {
		t.Error("expected reset token to be non-empty")
	}
}

func TestGenerateSessionID(t *testing.T) {
	session1, err := GenerateSessionID()
	if err != nil {
		t.Fatalf("failed to generate session ID: %v", err)
	}

	session2, err := GenerateSessionID()
	if err != nil {
		t.Fatalf("failed to generate session ID: %v", err)
	}

	if session1 == session2 {
		t.Error("expected two session IDs to be different")
	}

	if session1 == "" {
		t.Error("expected session ID to be non-empty")
	}
}

func TestHashSHA256(t *testing.T) {
	input := "test input string"

	hash1 := HashSHA256(input)
	hash2 := HashSHA256(input)

	if hash1 != hash2 {
		t.Error("expected same input to produce same hash")
	}

	if hash1 == "" {
		t.Error("expected hash to be non-empty")
	}

	if len(hash1) != 64 { // SHA256 produces 64 hex characters
		t.Errorf("expected hash length 64, got %d", len(hash1))
	}

	// Different input should produce different hash
	differentHash := HashSHA256("different input")
	if hash1 == differentHash {
		t.Error("expected different inputs to produce different hashes")
	}
}

func TestGenerateCodeVerifier(t *testing.T) {
	verifier, err := GenerateCodeVerifier()
	if err != nil {
		t.Fatalf("failed to generate code verifier: %v", err)
	}

	if verifier == "" {
		t.Error("expected code verifier to be non-empty")
	}

	// Should be base64url encoded
	if strings.Contains(verifier, "+") || strings.Contains(verifier, "/") || strings.Contains(verifier, "=") {
		t.Error("expected base64url encoding (no +, /, or = characters)")
	}
}

func TestGenerateCodeChallenge(t *testing.T) {
	verifier := "test_verifier_string_for_pkce"

	challenge1 := GenerateCodeChallenge(verifier)
	challenge2 := GenerateCodeChallenge(verifier)

	if challenge1 != challenge2 {
		t.Error("expected same verifier to produce same challenge")
	}

	if challenge1 == "" {
		t.Error("expected challenge to be non-empty")
	}

	// Should be base64url encoded
	if strings.Contains(challenge1, "+") || strings.Contains(challenge1, "/") || strings.Contains(challenge1, "=") {
		t.Error("expected base64url encoding (no +, /, or = characters)")
	}

	// Different verifier should produce different challenge
	differentChallenge := GenerateCodeChallenge("different_verifier")
	if challenge1 == differentChallenge {
		t.Error("expected different verifiers to produce different challenges")
	}
}

func TestGenerateStateToken(t *testing.T) {
	token, err := GenerateStateToken()
	if err != nil {
		t.Fatalf("failed to generate state token: %v", err)
	}

	if token == "" {
		t.Error("expected state token to be non-empty")
	}
}

func TestCheckPasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		expected PasswordStrength
	}{
		{
			name:     "very strong password",
			password: "MyVeryStr0ng!P@ssw0rd",
			expected: PasswordVeryStrong,
		},
		{
			name:     "strong password",
			password: "Str0ng!Pass12",
			expected: PasswordStrong,
		},
		{
			name:     "moderate password",
			password: "Pass123!",
			expected: PasswordModerate,
		},
		{
			name:     "weak password",
			password: "password",
			expected: PasswordWeak,
		},
		{
			name:     "weak short password",
			password: "Pass1!",
			expected: PasswordWeak,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strength := CheckPasswordStrength(tt.password)
			if strength != tt.expected {
				t.Errorf("expected strength %s, got %s", tt.expected.String(), strength.String())
			}
		})
	}
}

func TestPasswordStrength_String(t *testing.T) {
	tests := []struct {
		strength PasswordStrength
		expected string
	}{
		{PasswordWeak, "weak"},
		{PasswordModerate, "moderate"},
		{PasswordStrong, "strong"},
		{PasswordVeryStrong, "very_strong"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			str := tt.strength.String()
			if str != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, str)
			}
		})
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "testPassword123!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = HashPassword(password)
	}
}

func BenchmarkComparePasswords(b *testing.B) {
	password := "testPassword123!"
	hashed, _ := HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ComparePasswords(hashed, password)
	}
}

func BenchmarkGenerateRandomString(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateRandomString(32)
	}
}

func BenchmarkHashSHA256(b *testing.B) {
	input := "test input string for hashing"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = HashSHA256(input)
	}
}
