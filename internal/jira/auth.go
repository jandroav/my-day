package jira

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AuthManager handles API token authentication with Jira
type AuthManager struct {
	authFile string
	apiToken *APITokenAuth
}

// APITokenAuth represents API token authentication
type APITokenAuth struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

// NewAuthManager creates a new API token authentication manager
func NewAuthManager(email, token string) *AuthManager {
	homeDir, _ := os.UserHomeDir()
	authFile := filepath.Join(homeDir, ".my-day", "auth.json")

	return &AuthManager{
		authFile: authFile,
		apiToken: &APITokenAuth{
			Email: email,
			Token: token,
		},
	}
}

// SaveAPIToken saves the API token credentials to disk
func (am *AuthManager) SaveAPIToken() error {
	if am.apiToken == nil {
		return fmt.Errorf("no API token configured")
	}

	authInfo := AuthInfo{
		AuthType:  "token",
		APIToken:  am.apiToken,
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour), // API tokens don't expire, but we set a far future date
	}

	data, err := json.MarshalIndent(authInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal auth info: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(am.authFile)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create auth directory: %w", err)
	}

	if err := os.WriteFile(am.authFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write auth file: %w", err)
	}

	return nil
}

// LoadAPIToken loads the API token from disk
func (am *AuthManager) LoadAPIToken() (*APITokenAuth, error) {
	data, err := os.ReadFile(am.authFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read auth file: %w", err)
	}

	var authInfo AuthInfo
	if err := json.Unmarshal(data, &authInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal auth info: %w", err)
	}

	if authInfo.APIToken == nil {
		return nil, fmt.Errorf("no API token found in auth file")
	}

	return authInfo.APIToken, nil
}

// IsAuthenticated checks if valid API token authentication exists
func (am *AuthManager) IsAuthenticated() bool {
	_, err := am.LoadAPIToken()
	return err == nil
}

// ClearAuth removes stored authentication
func (am *AuthManager) ClearAuth() error {
	if err := os.Remove(am.authFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove auth file: %w", err)
	}
	return nil
}