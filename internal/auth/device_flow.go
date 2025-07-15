package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/pankona/memoya/internal/models"
	"github.com/pankona/memoya/internal/storage"
)

const (
	GoogleDeviceAuthURL = "https://oauth2.googleapis.com/device/code"
	GoogleTokenURL      = "https://oauth2.googleapis.com/token"
	GoogleUserInfoURL   = "https://www.googleapis.com/oauth2/v2/userinfo"
)

type DeviceFlowService struct {
	storage      storage.Storage
	clientID     string
	clientSecret string
	httpClient   *http.Client
}

func NewDeviceFlowService(storage storage.Storage, clientID, clientSecret string) *DeviceFlowService {
	return &DeviceFlowService{
		storage:      storage,
		clientID:     clientID,
		clientSecret: clientSecret,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// DeviceAuthResponse represents the response from Google's device auth endpoint
type DeviceAuthResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// StartDeviceFlow initiates the OAuth device flow
func (s *DeviceFlowService) StartDeviceFlow(ctx context.Context) (*models.DeviceAuthSession, error) {
	// Request device code from Google
	data := url.Values{}
	data.Set("client_id", s.clientID)
	data.Set("scope", "openid email profile")

	req, err := http.NewRequestWithContext(ctx, "POST", GoogleDeviceAuthURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create device auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "memoya-server/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device auth request failed with status: %d", resp.StatusCode)
	}

	var authResp DeviceAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("failed to decode device auth response: %w", err)
	}

	// Create session
	session := &models.DeviceAuthSession{
		DeviceCode:      authResp.DeviceCode,
		UserCode:        authResp.UserCode,
		VerificationURI: authResp.VerificationURL,
		ExpiresAt:       time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second),
		Status:          models.DeviceAuthStatusPending,
		CreatedAt:       time.Now(),
	}

	// Store session
	if err := s.storage.CreateDeviceAuthSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to store device auth session: %w", err)
	}

	return session, nil
}

// TokenResponse represents the response from Google's token endpoint
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	IDToken      string `json:"id_token,omitempty"`
	Error        string `json:"error,omitempty"`
}

// GoogleUserInfo represents user information from Google
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// PollToken polls for authorization completion
func (s *DeviceFlowService) PollToken(ctx context.Context, deviceCode string) (*models.User, string, error) {
	// Get session
	session, err := s.storage.GetDeviceAuthSession(ctx, deviceCode)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get device auth session: %w", err)
	}

	// Check if expired
	if time.Now().After(session.ExpiresAt) {
		session.Status = models.DeviceAuthStatusExpired
		s.storage.UpdateDeviceAuthSession(ctx, session)
		return nil, "", fmt.Errorf("device auth session expired")
	}

	// Check if already authorized
	if session.Status == models.DeviceAuthStatusAuthorized && session.UserID != "" {
		user, err := s.storage.GetUser(ctx, session.UserID)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get user: %w", err)
		}
		// Generate JWT token
		token, err := GenerateJWT(user.ID)
		if err != nil {
			return nil, "", fmt.Errorf("failed to generate JWT: %w", err)
		}
		return user, token, nil
	}

	// Poll Google for token
	data := url.Values{}
	data.Set("client_id", s.clientID)
	data.Set("client_secret", s.clientSecret)
	data.Set("device_code", session.DeviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	req, err := http.NewRequestWithContext(ctx, "POST", GoogleTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, "", fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to poll token: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, "", fmt.Errorf("failed to decode token response: %w", err)
	}

	// Handle errors
	if tokenResp.Error != "" {
		if tokenResp.Error == "authorization_pending" {
			return nil, "", fmt.Errorf("authorization pending")
		}
		return nil, "", fmt.Errorf("token error: %s", tokenResp.Error)
	}

	// Get user info from Google
	userInfo, err := s.getUserInfo(ctx, tokenResp.AccessToken)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get user info: %w", err)
	}

	// Find or create user
	user, err := s.findOrCreateUser(ctx, userInfo)
	if err != nil {
		return nil, "", fmt.Errorf("failed to find or create user: %w", err)
	}

	// Update session
	session.Status = models.DeviceAuthStatusAuthorized
	session.UserID = user.ID
	if err := s.storage.UpdateDeviceAuthSession(ctx, session); err != nil {
		return nil, "", fmt.Errorf("failed to update session: %w", err)
	}

	// Generate JWT token
	token, err := GenerateJWT(user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate JWT: %w", err)
	}

	return user, token, nil
}

func (s *DeviceFlowService) getUserInfo(ctx context.Context, accessToken string) (*GoogleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", GoogleUserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("userinfo request failed with status: %d", resp.StatusCode)
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

func (s *DeviceFlowService) findOrCreateUser(ctx context.Context, userInfo *GoogleUserInfo) (*models.User, error) {
	// Try to find existing user by Google ID
	user, err := s.storage.GetUserByGoogleID(ctx, userInfo.ID)
	if err == nil {
		// User exists, update last login and return
		user.IsActive = true
		if err := s.storage.UpdateUser(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
		return user, nil
	}

	// User doesn't exist, create new one
	user = &models.User{
		ID:        uuid.New().String(), // Privacy-focused: random UUID, not tied to Google ID
		GoogleID:  userInfo.ID,         // Only store Google ID for authentication
		CreatedAt: time.Now(),
		IsActive:  true,
	}

	if err := s.storage.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// CleanupExpiredSessions removes expired device auth sessions
func (s *DeviceFlowService) CleanupExpiredSessions(ctx context.Context) error {
	// This would need to be implemented based on storage capabilities
	// For now, we'll leave it as a placeholder since Firestore doesn't have direct collection-wide operations
	return nil
}

// generateSecureCode generates a cryptographically secure random code
func generateSecureCode(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)[:length]
}
