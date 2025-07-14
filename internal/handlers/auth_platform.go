package handlers

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Default implementations for platform-specific functions

func getXDGConfigDirDefault() string {
	// Check XDG_CONFIG_HOME first
	if configHome := os.Getenv("XDG_CONFIG_HOME"); configHome != "" {
		return configHome
	}

	// Fall back to $HOME/.config
	if home := os.Getenv("HOME"); home != "" {
		return filepath.Join(home, ".config")
	}

	// Last resort
	return ".config"
}

func readFileDefault(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return io.ReadAll(file)
}

func writeFileDefault(path string, data []byte) error {
	// Write to a temporary file first
	tmpPath := path + ".tmp"

	file, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}

	_, err = file.Write(data)
	if closeErr := file.Close(); closeErr != nil && err == nil {
		err = closeErr
	}

	if err != nil {
		os.Remove(tmpPath)
		return err
	}

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return nil
}

func createDirDefault(path string) error {
	return os.MkdirAll(path, 0700)
}

// GetAuthToken is a helper function to retrieve the saved auth token
func GetAuthToken() (string, error) {
	cm := NewConfigManager()
	config, err := cm.Load()
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // No config file yet
		}
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	if config.AuthToken == "" {
		return "", nil
	}

	// Check if token is expired
	if config.TokenExpiresAt != nil && config.TokenExpiresAt.Before(time.Now()) {
		return "", nil // Token expired
	}

	return config.AuthToken, nil
}
