package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/pankona/memoya/internal/auth"
	"github.com/pankona/memoya/internal/models"
	"github.com/pankona/memoya/internal/storage"
)

type AuthHandler struct {
	deviceFlowService *auth.DeviceFlowService
	storage           storage.Storage
}

func NewAuthHandler(storage storage.Storage, clientID, clientSecret string) *AuthHandler {
	deviceFlowService := auth.NewDeviceFlowService(storage, clientID, clientSecret)
	return &AuthHandler{
		deviceFlowService: deviceFlowService,
		storage:           storage,
	}
}

// DeviceAuthStartArgs represents arguments for starting device authentication
type DeviceAuthStartArgs struct {
	// No arguments needed for starting device flow
}

// DeviceAuthStartResult represents the result of starting device authentication
type DeviceAuthStartResult struct {
	Success         bool   `json:"success"`
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Message         string `json:"message"`
}

func (h *AuthHandler) StartDeviceAuth(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[DeviceAuthStartArgs]) (*mcp.CallToolResultFor[DeviceAuthStartResult], error) {
	if h.deviceFlowService == nil {
		return nil, fmt.Errorf("device flow service not initialized")
	}

	// Start device flow
	session, err := h.deviceFlowService.StartDeviceFlow(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start device flow: %w", err)
	}

	// Calculate expires_in from ExpiresAt
	expiresIn := int(time.Until(session.ExpiresAt).Seconds())

	result := DeviceAuthStartResult{
		Success:         true,
		DeviceCode:      session.DeviceCode,
		UserCode:        session.UserCode,
		VerificationURI: session.VerificationURI,
		ExpiresIn:       expiresIn,
		Message:         fmt.Sprintf("Please visit %s and enter code: %s", session.VerificationURI, session.UserCode),
	}

	// Convert to JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[DeviceAuthStartResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}

// DeviceAuthPollArgs represents arguments for polling device authentication
type DeviceAuthPollArgs struct {
	DeviceCode string `json:"device_code"`
}

// DeviceAuthPollResult represents the result of polling device authentication
type DeviceAuthPollResult struct {
	Success bool         `json:"success"`
	Status  string       `json:"status"` // "pending", "authorized", "expired"
	User    *models.User `json:"user,omitempty"`
	Token   string       `json:"token,omitempty"`
	Message string       `json:"message"`
}

func (h *AuthHandler) PollDeviceAuth(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[DeviceAuthPollArgs]) (*mcp.CallToolResultFor[DeviceAuthPollResult], error) {
	args := params.Arguments

	if h.deviceFlowService == nil {
		return nil, fmt.Errorf("device flow service not initialized")
	}

	// Poll for token
	user, token, err := h.deviceFlowService.PollToken(ctx, args.DeviceCode)
	if err != nil {
		// Check if it's just pending
		if err.Error() == "authorization pending" {
			result := DeviceAuthPollResult{
				Success: true,
				Status:  "pending",
				Message: "Authorization still pending",
			}
			jsonBytes, _ := json.Marshal(result)
			return &mcp.CallToolResultFor[DeviceAuthPollResult]{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(jsonBytes)},
				},
			}, nil
		}

		// Check if expired
		if err.Error() == "device auth session expired" {
			result := DeviceAuthPollResult{
				Success: false,
				Status:  "expired",
				Message: "Device authentication session has expired",
			}
			jsonBytes, _ := json.Marshal(result)
			return &mcp.CallToolResultFor[DeviceAuthPollResult]{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(jsonBytes)},
				},
			}, nil
		}

		return nil, fmt.Errorf("failed to poll device auth: %w", err)
	}

	result := DeviceAuthPollResult{
		Success: true,
		Status:  "authorized",
		User:    user,
		Token:   token,
		Message: fmt.Sprintf("Authentication successful for user %s", user.ID),
	}

	// Convert to JSON
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal result: %w", err)
	}

	return &mcp.CallToolResultFor[DeviceAuthPollResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}

// UserInfoArgs represents arguments for getting user information
type UserInfoArgs struct {
	// No arguments needed - user ID comes from JWT token context
}

// UserInfoResult represents the result of getting user information
type UserInfoResult struct {
	Success bool         `json:"success"`
	User    *models.User `json:"user,omitempty"`
	Message string       `json:"message"`
}

func (h *AuthHandler) GetUserInfo(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[UserInfoArgs]) (*mcp.CallToolResultFor[UserInfoResult], error) {
	// This would typically be called with authenticated context
	// For now, we'll return an error since we need the authentication middleware
	return nil, fmt.Errorf("user info endpoint requires authentication middleware")
}

// AccountDeleteArgs represents arguments for deleting user account
type AccountDeleteArgs struct {
	Confirm bool `json:"confirm"`
}

// AccountDeleteResult represents the result of deleting user account
type AccountDeleteResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (h *AuthHandler) DeleteAccount(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[AccountDeleteArgs]) (*mcp.CallToolResultFor[AccountDeleteResult], error) {
	args := params.Arguments

	if !args.Confirm {
		return nil, fmt.Errorf("account deletion requires explicit confirmation")
	}

	// This would typically be called with authenticated context
	// For now, we'll return an error since we need the authentication middleware
	return nil, fmt.Errorf("account deletion endpoint requires authentication middleware")
}
