package config

import (
	"context"
	"fmt"
	"os"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

// SecretManager handles Google Cloud Secret Manager operations
type SecretManager struct {
	client    *secretmanager.Client
	projectID string
}

// NewSecretManager creates a new SecretManager instance
func NewSecretManager(ctx context.Context, projectID string) (*SecretManager, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret manager client: %w", err)
	}

	return &SecretManager{
		client:    client,
		projectID: projectID,
	}, nil
}

// Close closes the SecretManager client
func (sm *SecretManager) Close() error {
	return sm.client.Close()
}

// GetSecret retrieves a secret value from Secret Manager
func (sm *SecretManager) GetSecret(ctx context.Context, secretName string) (string, error) {
	// Build the resource name
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", sm.projectID, secretName)

	// Access the secret version
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	}

	result, err := sm.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret %s: %w", secretName, err)
	}

	return string(result.Payload.Data), nil
}

// OAuthCredentials represents OAuth configuration
type OAuthCredentials struct {
	ClientID     string
	ClientSecret string
}

// GetOAuthCredentials retrieves OAuth credentials from environment variables or Secret Manager
func GetOAuthCredentials(ctx context.Context, projectID string) (*OAuthCredentials, error) {
	// Try environment variables first (for local development)
	clientID := os.Getenv("OAUTH_CLIENT_ID")
	clientSecret := os.Getenv("OAUTH_CLIENT_SECRET")

	if clientID != "" && clientSecret != "" {
		return &OAuthCredentials{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		}, nil
	}

	// If environment variables are not set, try Secret Manager (for production)
	if projectID == "" {
		return nil, fmt.Errorf("OAuth credentials not found: OAUTH_CLIENT_ID and OAUTH_CLIENT_SECRET environment variables are not set, and PROJECT_ID is not specified for Secret Manager")
	}

	sm, err := NewSecretManager(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("OAuth credentials not found: environment variables not set and Secret Manager unavailable: %w", err)
	}
	defer sm.Close()

	// Try to get credentials from Secret Manager
	if clientID == "" {
		clientID, err = sm.GetSecret(ctx, "oauth-client-id")
		if err != nil {
			return nil, fmt.Errorf("OAuth Client ID not found in environment variables or Secret Manager: %w", err)
		}
		clientID = strings.TrimSpace(clientID) // Remove whitespace and newlines
	}

	if clientSecret == "" {
		clientSecret, err = sm.GetSecret(ctx, "oauth-client-secret")
		if err != nil {
			return nil, fmt.Errorf("OAuth Client Secret not found in environment variables or Secret Manager: %w", err)
		}
		clientSecret = strings.TrimSpace(clientSecret) // Remove whitespace and newlines
	}

	return &OAuthCredentials{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, nil
}
