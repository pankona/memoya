package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient handles HTTP communication with the Cloud Run server
type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
	authToken  string
}

// NewHTTPClient creates a new HTTP client instance
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetAuthToken sets the bearer token for authentication
func (c *HTTPClient) SetAuthToken(token string) {
	c.authToken = token
}

// CallTool makes an HTTP POST request to the specified MCP tool endpoint
func (c *HTTPClient) CallTool(ctx context.Context, toolName string, args interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s/mcp/%s", c.baseURL, toolName)

	// Marshal arguments to JSON
	jsonData, err := json.Marshal(args)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal arguments: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "memoya-mcp-client/1.0")

	// Add authentication header if token is available
	if c.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check HTTP status
	if resp.StatusCode >= 400 {
		// Try to parse error response
		var errorResp struct {
			Success bool   `json:"success"`
			Error   string `json:"error"`
			Code    string `json:"code"`
		}

		if json.Unmarshal(body, &errorResp) == nil && !errorResp.Success {
			return nil, fmt.Errorf("server error [%s]: %s", errorResp.Code, errorResp.Error)
		}

		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// Ping checks if the server is reachable
func (c *HTTPClient) Ping(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create ping request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to ping server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-OK status: %d", resp.StatusCode)
	}

	return nil
}
