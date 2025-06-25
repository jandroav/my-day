package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"my-day/internal/config"
	"my-day/internal/jira"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Jira",
	Long: `Authenticate establishes OAuth connection with your Jira instance.

This will open your browser to complete the OAuth flow and store
authentication tokens securely for future use.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := authenticateWithJira(cmd); err != nil {
			color.Red("Authentication failed: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	
	// Auth-specific flags
	authCmd.Flags().Bool("clear", false, "Clear existing authentication")
	authCmd.Flags().Bool("test", false, "Test existing authentication")
	authCmd.Flags().Bool("no-browser", false, "Don't automatically open browser")
}

func authenticateWithJira(cmd *cobra.Command) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate required configuration
	if cfg.Jira.BaseURL == "" {
		return fmt.Errorf("Jira base URL not configured. Run 'my-day init' first")
	}
	if cfg.Jira.OAuth.ClientID == "" {
		return fmt.Errorf("Jira OAuth client ID not configured")
	}
	if cfg.Jira.OAuth.ClientSecret == "" {
		return fmt.Errorf("Jira OAuth client secret not configured")
	}

	client := jira.NewClient(
		cfg.Jira.BaseURL,
		cfg.Jira.OAuth.ClientID,
		cfg.Jira.OAuth.ClientSecret,
		cfg.Jira.OAuth.RedirectURI,
	)

	authManager := client.GetAuthManager()

	// Handle flags
	if clear, _ := cmd.Flags().GetBool("clear"); clear {
		if err := authManager.ClearAuth(); err != nil {
			return fmt.Errorf("failed to clear authentication: %w", err)
		}
		color.Green("✓ Authentication cleared")
		return nil
	}

	if test, _ := cmd.Flags().GetBool("test"); test {
		return testAuthentication(client)
	}

	// Check if already authenticated
	ctx := context.Background()
	if authManager.IsAuthenticated(ctx) {
		color.Yellow("Already authenticated with Jira")
		if err := testAuthentication(client); err != nil {
			color.Yellow("Authentication appears invalid, re-authenticating...")
		} else {
			color.Green("✓ Authentication is valid")
			return nil
		}
	}

	color.Cyan("🔐 Starting Jira OAuth authentication...")
	color.White("This will open your browser to complete the OAuth flow.")

	// Start local server for OAuth callback
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	tokenChan, errChan := authManager.StartAuthServer(ctx)

	// Generate and display auth URL
	authURL := authManager.GetAuthURL()
	color.White("\nAuthentication URL: %s", authURL)

	// Open browser automatically unless disabled
	if noBrowser, _ := cmd.Flags().GetBool("no-browser"); !noBrowser {
		color.White("Opening browser automatically...")
		if err := openBrowser(authURL); err != nil {
			color.Yellow("Failed to open browser automatically: %v", err)
			color.White("Please open the URL above manually")
		}
	}

	color.White("\nWaiting for authentication...")

	select {
	case token := <-tokenChan:
		if err := authManager.SaveToken(token); err != nil {
			return fmt.Errorf("failed to save authentication token: %w", err)
		}
		color.Green("✓ Authentication successful!")
		
		// Test the connection
		if err := testAuthentication(client); err != nil {
			color.Yellow("Warning: Authentication succeeded but connection test failed: %v", err)
		} else {
			color.Green("✓ Connection to Jira verified")
		}

	case err := <-errChan:
		return fmt.Errorf("authentication failed: %w", err)
		
	case <-ctx.Done():
		return fmt.Errorf("authentication timed out")
	}

	return nil
}

func testAuthentication(client *jira.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.TestConnection(ctx); err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	return nil
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}