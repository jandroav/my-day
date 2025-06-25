package jira

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
)

const (
	authorizationURL = "https://auth.atlassian.com/authorize"
	tokenURL         = "https://auth.atlassian.com/oauth/token"
	audience         = "api.atlassian.com"
	scope            = "read:jira-user read:jira-work"
)

// AuthManager handles OAuth authentication with Jira
type AuthManager struct {
	config     *oauth2.Config
	authFile   string
	serverPort string
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(clientID, clientSecret, redirectURI string) *AuthManager {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       []string{scope},
		Endpoint: oauth2.Endpoint{
			AuthURL:  authorizationURL,
			TokenURL: tokenURL,
		},
	}

	homeDir, _ := os.UserHomeDir()
	authFile := filepath.Join(homeDir, ".my-day", "auth.json")

	return &AuthManager{
		config:     config,
		authFile:   authFile,
		serverPort: "8080",
	}
}

// GetAuthURL generates the OAuth authorization URL
func (am *AuthManager) GetAuthURL() string {
	state := am.generateState()
	return am.config.AuthCodeURL(state, 
		oauth2.SetAuthURLParam("audience", audience),
		oauth2.SetAuthURLParam("prompt", "consent"),
	)
}

// StartAuthServer starts a local server to handle OAuth callback
func (am *AuthManager) StartAuthServer(ctx context.Context) (<-chan *oauth2.Token, <-chan error) {
	tokenChan := make(chan *oauth2.Token, 1)
	errChan := make(chan error, 1)

	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":" + am.serverPort,
		Handler: mux,
	}

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errChan <- fmt.Errorf("no authorization code received")
			return
		}

		token, err := am.config.Exchange(ctx, code,
			oauth2.SetAuthURLParam("audience", audience),
		)
		if err != nil {
			errChan <- fmt.Errorf("failed to exchange token: %w", err)
			return
		}

		tokenChan <- token

		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
			<html>
			<body style="font-family: Arial, sans-serif; text-align: center; margin-top: 50px;">
				<h2 style="color: green;">✓ Authentication Successful!</h2>
				<p>You can now close this tab and return to the terminal.</p>
				<script>setTimeout(() => window.close(), 3000);</script>
			</body>
			</html>
		`))
	})

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server error: %w", err)
		}
	}()

	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	return tokenChan, errChan
}

// SaveToken saves the OAuth token to disk
func (am *AuthManager) SaveToken(token *oauth2.Token) error {
	authInfo := AuthInfo{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		ExpiresAt:    token.Expiry,
		TokenType:    token.TokenType,
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

// LoadToken loads the OAuth token from disk
func (am *AuthManager) LoadToken() (*oauth2.Token, error) {
	data, err := os.ReadFile(am.authFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read auth file: %w", err)
	}

	var authInfo AuthInfo
	if err := json.Unmarshal(data, &authInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal auth info: %w", err)
	}

	token := &oauth2.Token{
		AccessToken:  authInfo.AccessToken,
		RefreshToken: authInfo.RefreshToken,
		Expiry:       authInfo.ExpiresAt,
		TokenType:    authInfo.TokenType,
	}

	return token, nil
}

// GetValidToken returns a valid token, refreshing if necessary
func (am *AuthManager) GetValidToken(ctx context.Context) (*oauth2.Token, error) {
	token, err := am.LoadToken()
	if err != nil {
		return nil, fmt.Errorf("no saved token found: %w", err)
	}

	// Check if token needs refresh
	if token.Expiry.Before(time.Now().Add(5 * time.Minute)) {
		tokenSource := am.config.TokenSource(ctx, token)
		newToken, err := tokenSource.Token()
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}

		if err := am.SaveToken(newToken); err != nil {
			return nil, fmt.Errorf("failed to save refreshed token: %w", err)
		}

		return newToken, nil
	}

	return token, nil
}

// IsAuthenticated checks if valid authentication exists
func (am *AuthManager) IsAuthenticated(ctx context.Context) bool {
	_, err := am.GetValidToken(ctx)
	return err == nil
}

// ClearAuth removes stored authentication
func (am *AuthManager) ClearAuth() error {
	if err := os.Remove(am.authFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove auth file: %w", err)
	}
	return nil
}

// generateState generates a random state parameter for OAuth
func (am *AuthManager) generateState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}