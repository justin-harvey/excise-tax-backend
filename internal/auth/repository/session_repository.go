// Package repository provides data access layer for session management operations.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/excise-tax-portal/backend/internal/auth/model"
	"github.com/excise-tax-portal/backend/pkg/cache"
	"github.com/excise-tax-portal/backend/pkg/utils"
)

var (
	// ErrSessionNotFound is returned when a session is not found
	ErrSessionNotFound = errors.New("session not found")
	// ErrSessionExpired is returned when a session has expired
	ErrSessionExpired = errors.New("session has expired")
)

// SessionRepository handles session storage in Redis
type SessionRepository struct {
	cache *cache.RedisClient
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(cache *cache.RedisClient) *SessionRepository {
	return &SessionRepository{cache: cache}
}

// CreateSession creates a new session in Redis
func (r *SessionRepository) CreateSession(ctx context.Context, session *model.Session) error {
	if session.ID == "" {
		sessionID, err := utils.GenerateSessionID()
		if err != nil {
			return fmt.Errorf("failed to generate session ID: %w", err)
		}
		session.ID = sessionID
	}

	if session.CreatedAt.IsZero() {
		session.CreatedAt = time.Now()
	}

	// Calculate TTL
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return errors.New("session expiration time must be in the future")
	}

	// Store session in Redis with key pattern: session:{session_id}
	key := fmt.Sprintf("session:%s", session.ID)
	if err := r.cache.Set(ctx, key, session, ttl); err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	// Also create a user sessions index for logout all functionality
	// Key pattern: user_sessions:{user_id}
	userSessionsKey := fmt.Sprintf("user_sessions:%d", session.UserID)
	if err := r.addToUserSessions(ctx, userSessionsKey, session.ID, ttl); err != nil {
		// Not critical if this fails, log but don't return error
		// In production, you'd want to log this
	}

	return nil
}

// GetSession retrieves a session from Redis
func (r *SessionRepository) GetSession(ctx context.Context, sessionID string) (*model.Session, error) {
	key := fmt.Sprintf("session:%s", sessionID)

	var session model.Session
	if err := r.cache.Get(ctx, key, &session); err != nil {
		if err.Error() == fmt.Sprintf("key %s not found", key) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		// Delete expired session
		_ = r.DeleteSession(ctx, sessionID)
		return nil, ErrSessionExpired
	}

	return &session, nil
}

// GetSessionByRefreshToken retrieves a session by refresh token
// This requires scanning all sessions for the user (using user_sessions index)
func (r *SessionRepository) GetSessionByRefreshToken(ctx context.Context, userID int64, refreshToken string) (*model.Session, error) {
	userSessionsKey := fmt.Sprintf("user_sessions:%d", userID)

	// Get all session IDs for the user
	var sessionIDs []string
	if err := r.cache.Get(ctx, userSessionsKey, &sessionIDs); err != nil {
		return nil, ErrSessionNotFound
	}

	// Check each session
	for _, sessionID := range sessionIDs {
		session, err := r.GetSession(ctx, sessionID)
		if err != nil {
			continue // Skip invalid/expired sessions
		}

		if session.RefreshToken == refreshToken {
			return session, nil
		}
	}

	return nil, ErrSessionNotFound
}

// UpdateSession updates an existing session
func (r *SessionRepository) UpdateSession(ctx context.Context, session *model.Session) error {
	// Calculate new TTL
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return errors.New("session expiration time must be in the future")
	}

	key := fmt.Sprintf("session:%s", session.ID)
	if err := r.cache.Set(ctx, key, session, ttl); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// DeleteSession deletes a session from Redis
func (r *SessionRepository) DeleteSession(ctx context.Context, sessionID string) error {
	// Get session first to remove from user sessions index
	session, err := r.GetSession(ctx, sessionID)
	if err != nil && !errors.Is(err, ErrSessionNotFound) && !errors.Is(err, ErrSessionExpired) {
		return err
	}

	// Delete the session
	key := fmt.Sprintf("session:%s", sessionID)
	if err := r.cache.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	// Remove from user sessions index if we have the session info
	if session != nil {
		userSessionsKey := fmt.Sprintf("user_sessions:%d", session.UserID)
		_ = r.removeFromUserSessions(ctx, userSessionsKey, sessionID)
	}

	return nil
}

// DeleteAllUserSessions deletes all sessions for a user (logout from all devices)
func (r *SessionRepository) DeleteAllUserSessions(ctx context.Context, userID int64) error {
	userSessionsKey := fmt.Sprintf("user_sessions:%d", userID)

	// Get all session IDs for the user
	var sessionIDs []string
	if err := r.cache.Get(ctx, userSessionsKey, &sessionIDs); err != nil {
		// If no sessions found, that's okay
		return nil
	}

	// Delete each session
	for _, sessionID := range sessionIDs {
		key := fmt.Sprintf("session:%s", sessionID)
		_ = r.cache.Delete(ctx, key) // Ignore errors, continue deleting
	}

	// Delete the user sessions index
	if err := r.cache.Delete(ctx, userSessionsKey); err != nil {
		return fmt.Errorf("failed to delete user sessions index: %w", err)
	}

	return nil
}

// ExtendSession extends the expiration time of a session
func (r *SessionRepository) ExtendSession(ctx context.Context, sessionID string, duration time.Duration) error {
	session, err := r.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.ExpiresAt = time.Now().Add(duration)
	return r.UpdateSession(ctx, session)
}

// GetActiveSessions returns all active sessions for a user
func (r *SessionRepository) GetActiveSessions(ctx context.Context, userID int64) ([]*model.Session, error) {
	userSessionsKey := fmt.Sprintf("user_sessions:%d", userID)

	var sessionIDs []string
	if err := r.cache.Get(ctx, userSessionsKey, &sessionIDs); err != nil {
		return []*model.Session{}, nil // No sessions is not an error
	}

	var sessions []*model.Session
	for _, sessionID := range sessionIDs {
		session, err := r.GetSession(ctx, sessionID)
		if err != nil {
			continue // Skip invalid/expired sessions
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// CleanupExpiredSessions removes expired sessions for a user (maintenance operation)
func (r *SessionRepository) CleanupExpiredSessions(ctx context.Context, userID int64) error {
	userSessionsKey := fmt.Sprintf("user_sessions:%d", userID)

	var sessionIDs []string
	if err := r.cache.Get(ctx, userSessionsKey, &sessionIDs); err != nil {
		return nil // No sessions is not an error
	}

	validSessionIDs := []string{}
	for _, sessionID := range sessionIDs {
		session, err := r.GetSession(ctx, sessionID)
		if err == nil && time.Now().Before(session.ExpiresAt) {
			validSessionIDs = append(validSessionIDs, sessionID)
		} else {
			// Delete expired session
			key := fmt.Sprintf("session:%s", sessionID)
			_ = r.cache.Delete(ctx, key)
		}
	}

	// Update user sessions index with only valid sessions
	if len(validSessionIDs) > 0 {
		// Set a reasonable TTL for the index (max of all session TTLs)
		ttl := 7 * 24 * time.Hour // Default to 7 days
		return r.cache.Set(ctx, userSessionsKey, validSessionIDs, ttl)
	} else {
		return r.cache.Delete(ctx, userSessionsKey)
	}
}

// Helper function to add session ID to user sessions index
func (r *SessionRepository) addToUserSessions(ctx context.Context, userSessionsKey, sessionID string, ttl time.Duration) error {
	var sessionIDs []string
	_ = r.cache.Get(ctx, userSessionsKey, &sessionIDs) // Ignore error, might not exist

	// Add new session ID if not already present
	found := false
	for _, id := range sessionIDs {
		if id == sessionID {
			found = true
			break
		}
	}

	if !found {
		sessionIDs = append(sessionIDs, sessionID)
	}

	return r.cache.Set(ctx, userSessionsKey, sessionIDs, ttl)
}

// Helper function to remove session ID from user sessions index
func (r *SessionRepository) removeFromUserSessions(ctx context.Context, userSessionsKey, sessionID string) error {
	var sessionIDs []string
	if err := r.cache.Get(ctx, userSessionsKey, &sessionIDs); err != nil {
		return nil // If index doesn't exist, nothing to remove
	}

	// Remove the session ID
	updatedIDs := []string{}
	for _, id := range sessionIDs {
		if id != sessionID {
			updatedIDs = append(updatedIDs, id)
		}
	}

	if len(updatedIDs) > 0 {
		return r.cache.Set(ctx, userSessionsKey, updatedIDs, 7*24*time.Hour)
	} else {
		return r.cache.Delete(ctx, userSessionsKey)
	}
}
