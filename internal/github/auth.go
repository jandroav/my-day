package github

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AuthManager handles GitHub token authentication
type AuthManager struct {
	authFile string
	token    string
}

// NewAuthManager creates a new GitHub authentication manager
func NewAuthManager(token string) *AuthManager {
	homeDir, _ := os.UserHomeDir()
	authFile := filepath.Join(homeDir, ".my-day", "github-auth.json")

	return &AuthManager{
		authFile: authFile,
		token:    token,
	}
}

// SaveToken saves the GitHub token credentials to disk
func (am *AuthManager) SaveToken() error {
	if am.token == "" {
		return fmt.Errorf("no GitHub token configured")
	}

	authInfo := AuthInfo{
		Token:     am.token,
		ExpiresAt: time.Now().Add(365 * 24 * time.Hour), // GitHub tokens don't expire unless revoked
	}

	data, err := json.MarshalIndent(authInfo, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal auth info: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(am.authFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create auth directory: %w", err)
	}

	// Write auth file with restricted permissions
	if err := os.WriteFile(am.authFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write auth file: %w", err)
	}

	return nil
}

// LoadToken loads the GitHub token from disk
func (am *AuthManager) LoadToken() (*AuthInfo, error) {
	data, err := os.ReadFile(am.authFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("GitHub not authenticated. Run 'my-day github connect --token your-token' first")
		}
		return nil, fmt.Errorf("failed to read auth file: %w", err)
	}

	var authInfo AuthInfo
	if err := json.Unmarshal(data, &authInfo); err != nil {
		return nil, fmt.Errorf("failed to parse auth file: %w", err)
	}

	// Check if token has expired (though GitHub tokens typically don't expire)
	if time.Now().After(authInfo.ExpiresAt) {
		return nil, fmt.Errorf("GitHub token has expired. Please re-authenticate")
	}

	return &authInfo, nil
}

// IsAuthenticated checks if GitHub authentication is available
func (am *AuthManager) IsAuthenticated() bool {
	_, err := am.LoadToken()
	return err == nil
}

// ClearAuthentication removes stored GitHub authentication
func (am *AuthManager) ClearAuthentication() error {
	if err := os.Remove(am.authFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove auth file: %w", err)
	}
	return nil
}

// GetAuthFile returns the path to the auth file
func (am *AuthManager) GetAuthFile() string {
	return am.authFile
}