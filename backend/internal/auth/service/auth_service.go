// Package service implements authentication business logic.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/excise-tax-portal/backend/internal/auth/model"
	"github.com/excise-tax-portal/backend/internal/auth/repository"
	"github.com/excise-tax-portal/backend/pkg/utils"
)

var (
	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")
	// ErrUnauthorized is returned when authentication fails
	ErrUnauthorized = errors.New("unauthorized")
	// ErrPermissionDenied is returned when user lacks required permissions
	ErrPermissionDenied = errors.New("permission denied")
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo    *repository.UserRepository
	sessionRepo *repository.SessionRepository
	jwtManager  *utils.JWTManager
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo *repository.UserRepository,
	sessionRepo *repository.SessionRepository,
	jwtManager *utils.JWTManager,
) *AuthService {
	return &AuthService{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		jwtManager:  jwtManager,
	}
}

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Email        string
	Password     string
	FirstName    string
	LastName     string
	Phone        string
	Role         string
	Manufacturer *model.Manufacturer // Optional manufacturer profile
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Email     string
	Password  string
	IPAddress string
	UserAgent string
}

// ChangePasswordRequest represents password change request
type ChangePasswordRequest struct {
	UserID      int64
	OldPassword string
	NewPassword string
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*model.User, error) {
	// Validate input
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("%w: email and password are required", ErrInvalidInput)
	}

	if err := utils.ValidatePassword(req.Password); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Default role to manufacturer if not specified
	if req.Role == "" {
		req.Role = model.RoleManufacturer
	}

	// Validate role
	if !model.IsValidRole(req.Role) {
		return nil, fmt.Errorf("%w: invalid role", ErrInvalidInput)
	}

	// Create user
	user := &model.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Role:      req.Role,
		IsActive:  true,
	}

	createdUser, err := s.userRepo.CreateUser(ctx, user, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, fmt.Errorf("%w: email already in use", ErrInvalidInput)
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create manufacturer profile if provided
	if req.Manufacturer != nil {
		req.Manufacturer.UserID = createdUser.ID
		_, err := s.userRepo.CreateManufacturer(ctx, req.Manufacturer)
		if err != nil {
			// Log error but don't fail registration
			// In production, you'd want proper error handling here
		}
	}

	return createdUser, nil
}

// Login authenticates a user and returns JWT tokens
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*model.TokenPair, *model.User, error) {
	// Validate credentials
	user, err := s.userRepo.ValidateUserCredentials(ctx, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidCredentials) {
			return nil, nil, ErrUnauthorized
		}
		return nil, nil, fmt.Errorf("failed to validate credentials: %w", err)
	}

	// Generate JWT claims
	claims := &utils.Claims{
		UserID:      user.ID,
		Email:       user.Email,
		Roles:       []string{user.Role},
		Permissions: model.GetRolePermissions(user.Role),
		IsAdmin:     model.IsAdmin(user.Role),
	}

	// Generate access token
	accessToken, err := s.jwtManager.GenerateAccessToken(claims)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := s.jwtManager.GenerateRefreshToken(claims)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Create session
	session := &model.Session{
		UserID:       user.ID,
		Email:        user.Email,
		Role:         user.Role,
		RefreshToken: refreshToken,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
		ExpiresAt:    time.Now().Add(7 * 24 * time.Hour), // 7 days
		CreatedAt:    time.Now(),
	}

	if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
		return nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail login
	}

	tokenPair := &model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600, // 1 hour in seconds
	}

	return tokenPair, user, nil
}

// Logout invalidates a user's session
func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("%w: session ID is required", ErrInvalidInput)
	}

	if err := s.sessionRepo.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// LogoutAll invalidates all sessions for a user
func (s *AuthService) LogoutAll(ctx context.Context, userID int64) error {
	if err := s.sessionRepo.DeleteAllUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete all user sessions: %w", err)
	}

	return nil
}

// RefreshToken renews an access token using a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*model.TokenPair, error) {
	// Validate refresh token
	claims, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid refresh token", ErrUnauthorized)
	}

	// Verify session exists
	session, err := s.sessionRepo.GetSessionByRefreshToken(ctx, claims.UserID, refreshToken)
	if err != nil {
		return nil, fmt.Errorf("%w: session not found", ErrUnauthorized)
	}

	// Get updated user information
	user, err := s.userRepo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is still active
	if !user.IsActive {
		return nil, fmt.Errorf("%w: user account is inactive", ErrUnauthorized)
	}

	// Generate new tokens with updated user information
	newClaims := &utils.Claims{
		UserID:      user.ID,
		Email:       user.Email,
		Roles:       []string{user.Role},
		Permissions: model.GetRolePermissions(user.Role),
		IsAdmin:     model.IsAdmin(user.Role),
	}

	accessToken, err := s.jwtManager.GenerateAccessToken(newClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(newClaims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Update session with new refresh token
	session.RefreshToken = newRefreshToken
	session.ExpiresAt = time.Now().Add(7 * 24 * time.Hour)
	if err := s.sessionRepo.UpdateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	tokenPair := &model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}

	return tokenPair, nil
}

// ValidateToken validates a JWT token and returns claims
func (s *AuthService) ValidateToken(ctx context.Context, token string) (*utils.Claims, error) {
	claims, err := s.jwtManager.ValidateAccessToken(token)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrUnauthorized, err)
	}

	return claims, nil
}

// GetCurrentUser retrieves user information from token
func (s *AuthService) GetCurrentUser(ctx context.Context, userID int64) (*model.UserWithManufacturer, error) {
	userWithManufacturer, err := s.userRepo.GetUserWithManufacturer(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return userWithManufacturer, nil
}

// ChangePassword changes a user's password
func (s *AuthService) ChangePassword(ctx context.Context, req *ChangePasswordRequest) error {
	// Validate new password
	if err := utils.ValidatePassword(req.NewPassword); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Get user
	user, err := s.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify old password
	if err := utils.ComparePasswords(user.PasswordHash, req.OldPassword); err != nil {
		return fmt.Errorf("%w: current password is incorrect", ErrUnauthorized)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, req.UserID, req.NewPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate all sessions (force re-login)
	_ = s.sessionRepo.DeleteAllUserSessions(ctx, req.UserID)

	return nil
}

// ResetPassword resets a user's password (for forgot password flow)
// In a real implementation, this would verify a reset token first
func (s *AuthService) ResetPassword(ctx context.Context, email, newPassword string, resetToken string) error {
	// TODO: Implement token verification
	// For now, this is a simplified version

	// Validate new password
	if err := utils.ValidatePassword(newPassword); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Get user by email
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Update password
	if err := s.userRepo.UpdatePassword(ctx, user.ID, newPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Invalidate all sessions
	_ = s.sessionRepo.DeleteAllUserSessions(ctx, user.ID)

	return nil
}

// CheckPermission checks if a user has a specific permission
func (s *AuthService) CheckPermission(ctx context.Context, userID int64, permission string) (bool, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	permissions := model.GetRolePermissions(user.Role)
	for _, p := range permissions {
		if p == permission {
			return true, nil
		}
	}

	return false, nil
}

// CheckRole checks if a user has a specific role
func (s *AuthService) CheckRole(ctx context.Context, userID int64, role string) (bool, error) {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("failed to get user: %w", err)
	}

	return user.Role == role, nil
}

// GetActiveSessions returns all active sessions for a user
func (s *AuthService) GetActiveSessions(ctx context.Context, userID int64) ([]*model.Session, error) {
	sessions, err := s.sessionRepo.GetActiveSessions(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}

	return sessions, nil
}
