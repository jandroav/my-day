package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"my-day/internal/config"
	"my-day/internal/jira"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Jira using API token",
	Long: `Authenticate with Jira using API token (recommended for CLI usage).

Create an API token:
1. Go to https://id.atlassian.com/manage-profile/security/api-tokens
2. Click "Create API token"
3. Give it a name and copy the generated token
4. Use this command: my-day auth --email your-email@example.com --token your-api-token`,
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
	authCmd.Flags().String("email", "", "Email address for API token authentication (can be set in config)")
	authCmd.Flags().String("token", "", "API token for authentication (can be set in config)")
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

	// Handle clear flag
	if clear, _ := cmd.Flags().GetBool("clear"); clear {
		// Create a temporary auth manager to clear auth
		authManager := jira.NewAuthManager("", "")
		if err := authManager.ClearAuth(); err != nil {
			return fmt.Errorf("failed to clear authentication: %w", err)
		}
		color.Green("✓ Authentication cleared")
		return nil
	}

	// Handle test flag
	if test, _ := cmd.Flags().GetBool("test"); test {
		// Create a temporary auth manager to test auth
		authManager := jira.NewAuthManager("", "")
		if !authManager.IsAuthenticated() {
			return fmt.Errorf("no authentication found. Run 'my-day auth --email your-email --token your-token' first")
		}
		
		// Test with a real client
		apiToken, err := authManager.LoadAPIToken()
		if err != nil {
			return fmt.Errorf("failed to load API token: %w", err)
		}
		
		client := jira.NewClient(cfg.Jira.BaseURL, apiToken.Email, apiToken.Token)
		return testAuthentication(client)
	}

	// Get email and token from flags or config
	email, _ := cmd.Flags().GetString("email")
	token, _ := cmd.Flags().GetString("token")
	
	// If not provided as flags, use config values
	if email == "" {
		email = cfg.Jira.Email
	}
	if token == "" {
		token = cfg.Jira.Token
	}

	if email == "" || token == "" {
		return fmt.Errorf("email and token are required. Use --email and --token flags or set them in config file")
	}

	// Create client and save credentials
	color.Cyan("🔑 Configuring API token authentication...")
	client := jira.NewClient(cfg.Jira.BaseURL, email, token)

	// Save the API token
	if err := client.GetAuthManager().SaveAPIToken(); err != nil {
		return fmt.Errorf("failed to save API token: %w", err)
	}

	color.Green("✓ API token authentication configured!")

	// Test the connection
	if err := testAuthentication(client); err != nil {
		color.Yellow("Warning: Authentication saved but connection test failed: %v", err)
		color.Yellow("Please verify your Jira base URL and API token are correct")
	} else {
		color.Green("✓ Connection to Jira verified")
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