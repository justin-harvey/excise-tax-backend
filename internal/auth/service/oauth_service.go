// Package service implements OAuth2 authentication logic for third-party providers.
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/excise-tax-portal/backend/pkg/utils"
)

// OAuthProvider represents supported OAuth providers
type OAuthProvider string

const (
	// ProviderStateDotGov represents State.gov CDB OAuth provider
	ProviderStateDotGov OAuthProvider = "state_gov"
	// ProviderGoogle represents Google OAuth provider
	ProviderGoogle OAuthProvider = "google"
)

// OAuthConfig holds OAuth2 configuration
type OAuthConfig struct {
	Provider         OAuthProvider
	ClientID         string
	ClientSecret     string
	RedirectURI      string
	AuthorizationURL string
	TokenURL         string
	UserInfoURL      string
	Scopes           []string
}

// OAuthService handles OAuth2 authentication flows
type OAuthService struct {
	config     *OAuthConfig
	httpClient *http.Client
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(config *OAuthConfig) (*OAuthService, error) {
	if config == nil {
		return nil, errors.New("OAuth config cannot be nil")
	}

	if config.ClientID == "" || config.RedirectURI == "" {
		return nil, errors.New("OAuth client ID and redirect URI are required")
	}

	return &OAuthService{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// AuthorizationURLRequest represents request for authorization URL
type AuthorizationURLRequest struct {
	State       string
	UsePKCE     bool
	ExtraParams map[string]string
}

// AuthorizationURLResponse contains the authorization URL and PKCE verifier
type AuthorizationURLResponse struct {
	URL          string
	State        string
	CodeVerifier string // Only set if PKCE is used
}

// GetAuthorizationURL generates an OAuth authorization URL with optional PKCE
func (s *OAuthService) GetAuthorizationURL(req *AuthorizationURLRequest) (*AuthorizationURLResponse, error) {
	params := url.Values{}
	params.Set("client_id", s.config.ClientID)
	params.Set("redirect_uri", s.config.RedirectURI)
	params.Set("response_type", "code")

	// Add scopes
	if len(s.config.Scopes) > 0 {
		scopes := ""
		for i, scope := range s.config.Scopes {
			if i > 0 {
				scopes += " "
			}
			scopes += scope
		}
		params.Set("scope", scopes)
	}

	// Generate or use provided state
	state := req.State
	if state == "" {
		var err error
		state, err = utils.GenerateStateToken()
		if err != nil {
			return nil, fmt.Errorf("failed to generate state token: %w", err)
		}
	}
	params.Set("state", state)

	// PKCE support
	var codeVerifier string
	if req.UsePKCE {
		var err error
		codeVerifier, err = utils.GenerateCodeVerifier()
		if err != nil {
			return nil, fmt.Errorf("failed to generate code verifier: %w", err)
		}

		codeChallenge := utils.GenerateCodeChallenge(codeVerifier)
		params.Set("code_challenge", codeChallenge)
		params.Set("code_challenge_method", "S256")
	}

	// Add extra parameters
	for key, value := range req.ExtraParams {
		params.Set(key, value)
	}

	authURL := s.config.AuthorizationURL + "?" + params.Encode()

	return &AuthorizationURLResponse{
		URL:          authURL,
		State:        state,
		CodeVerifier: codeVerifier,
	}, nil
}

// ExchangeCodeRequest represents token exchange request
type ExchangeCodeRequest struct {
	Code         string
	CodeVerifier string // For PKCE
	State        string // For validation
}

// TokenResponse represents OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
	IDToken      string `json:"id_token,omitempty"` // For OpenID Connect
}

// ExchangeCodeForToken exchanges an authorization code for an access token
func (s *OAuthService) ExchangeCodeForToken(ctx context.Context, req *ExchangeCodeRequest) (*TokenResponse, error) {
	if req.Code == "" {
		return nil, errors.New("authorization code is required")
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", req.Code)
	data.Set("redirect_uri", s.config.RedirectURI)
	data.Set("client_id", s.config.ClientID)

	// Add client secret if configured (not used in PKCE flow for public clients)
	if s.config.ClientSecret != "" {
		data.Set("client_secret", s.config.ClientSecret)
	}

	// Add PKCE code verifier if provided
	if req.CodeVerifier != "" {
		data.Set("code_verifier", req.CodeVerifier)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.config.TokenURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.URL.RawQuery = data.Encode()
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// UserInfo represents user information from OAuth provider
type UserInfo struct {
	ID            string                 `json:"id"`
	Email         string                 `json:"email"`
	EmailVerified bool                   `json:"email_verified"`
	Name          string                 `json:"name"`
	GivenName     string                 `json:"given_name"`
	FamilyName    string                 `json:"family_name"`
	Picture       string                 `json:"picture"`
	Locale        string                 `json:"locale"`
	Extra         map[string]interface{} `json:"-"` // For provider-specific fields
}

// GetUserInfo retrieves user information using an access token
func (s *OAuthService) GetUserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	if accessToken == "" {
		return nil, errors.New("access token is required")
	}

	if s.config.UserInfoURL == "" {
		return nil, errors.New("user info URL not configured")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", s.config.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user info request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &userInfo, nil
}

// ValidateState validates an OAuth state parameter
func (s *OAuthService) ValidateState(received, expected string) error {
	if received == "" {
		return errors.New("state parameter is missing")
	}

	if received != expected {
		return errors.New("state parameter mismatch - possible CSRF attack")
	}

	return nil
}

// RefreshAccessToken refreshes an access token using a refresh token
func (s *OAuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	if refreshToken == "" {
		return nil, errors.New("refresh token is required")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", s.config.ClientID)

	if s.config.ClientSecret != "" {
		data.Set("client_secret", s.config.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.config.TokenURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.URL.RawQuery = data.Encode()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to decode token response: %w", err)
	}

	return &tokenResp, nil
}

// DefaultStateDotGovConfig returns default OAuth config for State.gov CDB
func DefaultStateDotGovConfig(clientID, clientSecret, redirectURI string) *OAuthConfig {
	return &OAuthConfig{
		Provider:         ProviderStateDotGov,
		ClientID:         clientID,
		ClientSecret:     clientSecret,
		RedirectURI:      redirectURI,
		AuthorizationURL: "https://cdb.state.gov/oauth/authorize", // Example URL
		TokenURL:         "https://cdb.state.gov/oauth/token",
		UserInfoURL:      "https://cdb.state.gov/oauth/userinfo",
		Scopes:           []string{"openid", "profile", "email"},
	}
}

// DefaultGoogleConfig returns default OAuth config for Google
func DefaultGoogleConfig(clientID, clientSecret, redirectURI string) *OAuthConfig {
	return &OAuthConfig{
		Provider:         ProviderGoogle,
		ClientID:         clientID,
		ClientSecret:     clientSecret,
		RedirectURI:      redirectURI,
		AuthorizationURL: "https://accounts.google.com/o/oauth2/v2/auth",
		TokenURL:         "https://oauth2.googleapis.com/token",
		UserInfoURL:      "https://www.googleapis.com/oauth2/v2/userinfo",
		Scopes:           []string{"openid", "profile", "email"},
	}
}
