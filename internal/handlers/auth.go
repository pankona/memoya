package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// HTTPAuthClient interface for making authentication requests
type HTTPAuthClient interface {
	CallTool(ctx context.Context, toolName string, args interface{}) ([]byte, error)
}

type AuthHandler struct {
	httpClient    HTTPAuthClient
	configManager *ConfigManager
}

func NewAuthHandler(httpClient HTTPAuthClient) *AuthHandler {
	return &AuthHandler{
		httpClient:    httpClient,
		configManager: NewConfigManager(),
	}
}

type AuthStartArgs struct{}

type AuthStartResult struct {
	Success         bool   `json:"success"`
	DeviceCode      string `json:"device_code,omitempty"`
	UserCode        string `json:"user_code,omitempty"`
	VerificationURL string `json:"verification_url,omitempty"`
	ExpiresIn       int    `json:"expires_in,omitempty"`
	Message         string `json:"message"`
}

func (h *AuthHandler) Start(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[AuthStartArgs]) (*mcp.CallToolResultFor[AuthStartResult], error) {
	// Call server's device auth start endpoint
	respData, err := h.httpClient.CallTool(ctx, "device_auth_start", struct{}{})
	if err != nil {
		result := AuthStartResult{
			Success: false,
			Message: fmt.Sprintf("Failed to start authentication: %v", err),
		}

		jsonBytes, _ := json.Marshal(result)
		return &mcp.CallToolResultFor[AuthStartResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(jsonBytes)},
			},
		}, nil
	}

	// Parse server response
	var serverResp struct {
		Success bool `json:"success"`
		Data    struct {
			DeviceCode      string `json:"device_code"`
			UserCode        string `json:"user_code"`
			VerificationURI string `json:"verification_uri"`
			ExpiresIn       int    `json:"expires_in"`
		} `json:"data"`
		Message string `json:"message"`
	}

	if err := json.Unmarshal(respData, &serverResp); err != nil {
		result := AuthStartResult{
			Success: false,
			Message: "Failed to parse server response",
		}

		jsonBytes, _ := json.Marshal(result)
		return &mcp.CallToolResultFor[AuthStartResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(jsonBytes)},
			},
		}, nil
	}

	if !serverResp.Success {
		result := AuthStartResult{
			Success: false,
			Message: serverResp.Message,
		}

		jsonBytes, _ := json.Marshal(result)
		return &mcp.CallToolResultFor[AuthStartResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(jsonBytes)},
			},
		}, nil
	}

	// Store device code temporarily in config
	config, err := h.configManager.Load()
	if err != nil {
		config = &Config{}
	}

	config.PendingAuth = &PendingAuth{
		DeviceCode: serverResp.Data.DeviceCode,
		StartedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(time.Duration(serverResp.Data.ExpiresIn) * time.Second),
	}

	if err := h.configManager.Save(config); err != nil {
		// Continue even if we can't save the pending auth
		fmt.Printf("Warning: Failed to save pending auth: %v\n", err)
	}

	result := AuthStartResult{
		Success:         true,
		DeviceCode:      serverResp.Data.DeviceCode,
		UserCode:        serverResp.Data.UserCode,
		VerificationURL: serverResp.Data.VerificationURI,
		ExpiresIn:       serverResp.Data.ExpiresIn,
		Message:         fmt.Sprintf("Please visit %s and enter code: %s", serverResp.Data.VerificationURI, serverResp.Data.UserCode),
	}

	jsonBytes, _ := json.Marshal(result)
	return &mcp.CallToolResultFor[AuthStartResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}

type AuthStatusArgs struct{}

type AuthStatusResult struct {
	Success       bool   `json:"success"`
	Authenticated bool   `json:"authenticated"`
	Token         string `json:"token,omitempty"`
	ExpiresAt     string `json:"expires_at,omitempty"`
	Message       string `json:"message"`
}

func (h *AuthHandler) Status(ctx context.Context, ss *mcp.ServerSession, params *mcp.CallToolParamsFor[AuthStatusArgs]) (*mcp.CallToolResultFor[AuthStatusResult], error) {
	config, err := h.configManager.Load()
	if err != nil {
		result := AuthStatusResult{
			Success:       false,
			Authenticated: false,
			Message:       "No authentication information found",
		}

		jsonBytes, _ := json.Marshal(result)
		return &mcp.CallToolResultFor[AuthStatusResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(jsonBytes)},
			},
		}, nil
	}

	// Check if there's a pending auth first (prioritize new authentication)
	if config.PendingAuth != nil && config.PendingAuth.ExpiresAt.After(time.Now()) {
		// Poll server for the result
		pollArgs := struct {
			DeviceCode string `json:"device_code"`
		}{
			DeviceCode: config.PendingAuth.DeviceCode,
		}

		respData, err := h.httpClient.CallTool(ctx, "device_auth_poll", pollArgs)
		if err != nil {
			result := AuthStatusResult{
				Success:       false,
				Authenticated: false,
				Message:       fmt.Sprintf("Failed to check authentication status: %v", err),
			}

			jsonBytes, _ := json.Marshal(result)
			return &mcp.CallToolResultFor[AuthStatusResult]{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(jsonBytes)},
				},
			}, nil
		}

		// Parse server response
		var serverResp struct {
			Success bool `json:"success"`
			Data    *struct {
				AccessToken string `json:"access_token"`
				User        struct {
					ID       string `json:"id"`
					GoogleID string `json:"google_id"`
				} `json:"user"`
			} `json:"data,omitempty"`
			Message string `json:"message"`
		}

		if err := json.Unmarshal(respData, &serverResp); err != nil {
			result := AuthStatusResult{
				Success:       false,
				Authenticated: false,
				Message:       "Failed to parse server response",
			}

			jsonBytes, _ := json.Marshal(result)
			return &mcp.CallToolResultFor[AuthStatusResult]{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(jsonBytes)},
				},
			}, nil
		}

		// Check if authentication failed
		if !serverResp.Success {
			// Check for specific error messages
			if strings.Contains(serverResp.Message, "authorization pending") {
				result := AuthStatusResult{
					Success:       true,
					Authenticated: false,
					Message:       "Waiting for user authorization. Please complete the authentication in your browser.",
				}

				jsonBytes, _ := json.Marshal(result)
				return &mcp.CallToolResultFor[AuthStatusResult]{
					Content: []mcp.Content{
						&mcp.TextContent{Text: string(jsonBytes)},
					},
				}, nil
			}

			// Clear the pending auth
			config.PendingAuth = nil
			h.configManager.Save(config)

			result := AuthStatusResult{
				Success:       false,
				Authenticated: false,
				Message:       serverResp.Message,
			}

			jsonBytes, _ := json.Marshal(result)
			return &mcp.CallToolResultFor[AuthStatusResult]{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(jsonBytes)},
				},
			}, nil
		}

		// Check if we got data with access token
		if serverResp.Data == nil || serverResp.Data.AccessToken == "" {
			result := AuthStatusResult{
				Success:       false,
				Authenticated: false,
				Message:       "Poll completed successfully",
			}

			jsonBytes, _ := json.Marshal(result)
			return &mcp.CallToolResultFor[AuthStatusResult]{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(jsonBytes)},
				},
			}, nil
		}

		// Success! Save the token
		config.AuthToken = serverResp.Data.AccessToken
		expiresAt := time.Now().Add(7 * 24 * time.Hour) // JWT tokens expire in 7 days (1 week)
		config.TokenExpiresAt = &expiresAt
		config.PendingAuth = nil

		if err := h.configManager.Save(config); err != nil {
			result := AuthStatusResult{
				Success:       false,
				Authenticated: false,
				Message:       fmt.Sprintf("Failed to save authentication token: %v", err),
			}

			jsonBytes, _ := json.Marshal(result)
			return &mcp.CallToolResultFor[AuthStatusResult]{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(jsonBytes)},
				},
			}, nil
		}

		result := AuthStatusResult{
			Success:       true,
			Authenticated: true,
			Token:         serverResp.Data.AccessToken,
			ExpiresAt:     expiresAt.Format(time.RFC3339),
			Message:       "Authentication successful!",
		}

		jsonBytes, _ := json.Marshal(result)
		return &mcp.CallToolResultFor[AuthStatusResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(jsonBytes)},
			},
		}, nil
	}

	// Check if we have a valid existing token (fallback)
	if config.AuthToken != "" && config.TokenExpiresAt != nil && config.TokenExpiresAt.After(time.Now()) {
		result := AuthStatusResult{
			Success:       true,
			Authenticated: true,
			Token:         config.AuthToken,
			ExpiresAt:     config.TokenExpiresAt.Format(time.RFC3339),
			Message:       "Authenticated",
		}

		jsonBytes, _ := json.Marshal(result)
		return &mcp.CallToolResultFor[AuthStatusResult]{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(jsonBytes)},
			},
		}, nil
	}

	// No valid token and no pending auth
	result := AuthStatusResult{
		Success:       true,
		Authenticated: false,
		Message:       "Not authenticated. Use auth_start to begin authentication.",
	}

	jsonBytes, _ := json.Marshal(result)
	return &mcp.CallToolResultFor[AuthStatusResult]{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(jsonBytes)},
		},
	}, nil
}

// ConfigManager handles reading and writing configuration to XDG config directory
type ConfigManager struct {
	configPath string
}

type Config struct {
	AuthToken      string       `json:"auth_token,omitempty"`
	TokenExpiresAt *time.Time   `json:"token_expires_at,omitempty"`
	PendingAuth    *PendingAuth `json:"pending_auth,omitempty"`
}

type PendingAuth struct {
	DeviceCode string    `json:"device_code"`
	StartedAt  time.Time `json:"started_at"`
	ExpiresAt  time.Time `json:"expires_at"`
}

func NewConfigManager() *ConfigManager {
	return &ConfigManager{}
}

func (c *ConfigManager) getConfigPath() string {
	if c.configPath != "" {
		return c.configPath
	}

	// Get XDG config directory
	configDir := getXDGConfigDir()
	c.configPath = filepath.Join(configDir, "memoya", "auth.json")
	return c.configPath
}

func (c *ConfigManager) Load() (*Config, error) {
	configPath := c.getConfigPath()

	data, err := readFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

func (c *ConfigManager) Save(config *Config) error {
	configPath := c.getConfigPath()

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := createDir(dir); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := writeFile(configPath, data); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Platform-specific functions (will be implemented in separate files)
var (
	getXDGConfigDir = getXDGConfigDirDefault
	readFile        = readFileDefault
	writeFile       = writeFileDefault
	createDir       = createDirDefault
)
