package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"my-day/internal/github"
)

// githubCmd represents the github command
var githubCmd = &cobra.Command{
	Use:   "github",
	Short: "Manage GitHub integration",
	Long: `Manage GitHub integration for tracking pull requests, commits, and other code activities.

Connect to GitHub to include code activity in your daily reports alongside Jira tickets.`,
}

// githubConnectCmd represents the github connect command
var githubConnectCmd = &cobra.Command{
	Use:   "connect",
	Short: "Connect to GitHub",
	Long: `Connect to GitHub using a personal access token.

To create a GitHub personal access token:
1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Select these scopes: repo, user, workflow
4. Copy the generated token

Example:
  my-day github connect --token ghp_xxxxxxxxxxxxxxxxxxxx`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := connectGitHub(cmd); err != nil {
			color.Red("Failed to connect to GitHub: %v", err)
			os.Exit(1)
		}
	},
}

// githubStatusCmd represents the github status command
var githubStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show GitHub connection status",
	Long:  `Show the current GitHub connection status and user information.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := showGitHubStatus(cmd); err != nil {
			color.Red("Failed to get GitHub status: %v", err)
			os.Exit(1)
		}
	},
}

// githubTestCmd represents the github test command
var githubTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test GitHub connection",
	Long:  `Test the GitHub API connection and display user information.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := testGitHubConnection(cmd); err != nil {
			color.Red("GitHub connection test failed: %v", err)
			os.Exit(1)
		}
	},
}

// githubDisconnectCmd represents the github disconnect command
var githubDisconnectCmd = &cobra.Command{
	Use:   "disconnect",
	Short: "Disconnect from GitHub",
	Long:  `Remove GitHub authentication and disconnect from the service.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := disconnectGitHub(cmd); err != nil {
			color.Red("Failed to disconnect from GitHub: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(githubCmd)
	githubCmd.AddCommand(githubConnectCmd)
	githubCmd.AddCommand(githubStatusCmd)
	githubCmd.AddCommand(githubTestCmd)
	githubCmd.AddCommand(githubDisconnectCmd)

	// Flags for connect command
	githubConnectCmd.Flags().String("token", "", "GitHub personal access token")
	githubConnectCmd.Flags().Bool("test", true, "Test connection after connecting")
}

func connectGitHub(cmd *cobra.Command) error {
	token, _ := cmd.Flags().GetString("token")
	test, _ := cmd.Flags().GetBool("test")

	// Check for token in environment if not provided
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	if token == "" {
		return fmt.Errorf("GitHub token is required. Use --token flag or set GITHUB_TOKEN environment variable")
	}

	color.Cyan("🔗 Connecting to GitHub...")

	// Create auth manager and save token
	authManager := github.NewAuthManager(token)
	if err := authManager.SaveToken(); err != nil {
		return fmt.Errorf("failed to save GitHub token: %w", err)
	}

	color.Green("✓ GitHub token saved")

	// Test connection if requested
	if test {
		client := github.NewClient(token)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		user, err := client.GetCurrentUser(ctx)
		if err != nil {
			color.Yellow("⚠️  GitHub token saved, but connection test failed: %v", err)
			return nil
		}

		color.Green("✓ GitHub connection successful")
		color.White("Connected as: %s (%s)", user.Name, user.Login)
		if user.Email != "" {
			color.White("Email: %s", user.Email)
		}
	}

	return nil
}

func showGitHubStatus(cmd *cobra.Command) error {
	authManager := github.NewAuthManager("")
	
	if !authManager.IsAuthenticated() {
		color.Yellow("❌ GitHub not connected")
		color.White("Run 'my-day github connect --token your-token' to connect")
		return nil
	}

	authInfo, err := authManager.LoadToken()
	if err != nil {
		return fmt.Errorf("failed to load GitHub auth: %w", err)
	}

	color.Green("✅ GitHub connected")
	color.White("Token expires: %s", authInfo.ExpiresAt.Format("2006-01-02 15:04:05"))
	
	if authInfo.Username != "" {
		color.White("Username: %s", authInfo.Username)
	}

	// Test current connection
	client := github.NewClient(authInfo.Token)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := client.GetCurrentUser(ctx)
	if err != nil {
		color.Yellow("⚠️  Connection test failed: %v", err)
		color.White("You may need to reconnect: my-day github connect --token your-token")
	} else {
		color.Green("✓ Connection test successful")
		color.White("User: %s (%s)", user.Name, user.Login)
	}

	return nil
}

func testGitHubConnection(cmd *cobra.Command) error {
	authManager := github.NewAuthManager("")
	
	if !authManager.IsAuthenticated() {
		return fmt.Errorf("GitHub not connected. Run 'my-day github connect' first")
	}

	authInfo, err := authManager.LoadToken()
	if err != nil {
		return fmt.Errorf("failed to load GitHub auth: %w", err)
	}

	color.Cyan("🧪 Testing GitHub connection...")

	client := github.NewClient(authInfo.Token)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test basic connection
	user, err := client.GetCurrentUser(ctx)
	if err != nil {
		return fmt.Errorf("connection test failed: %w", err)
	}

	color.Green("✓ GitHub API connection successful")
	color.White("Authenticated as: %s (%s)", user.Name, user.Login)
	if user.Email != "" {
		color.White("Email: %s", user.Email)
	}
	color.White("User Type: %s", user.Type)

	// Test repositories access
	color.Cyan("🔍 Testing repository access...")
	repos, err := client.GetUserRepositories(ctx, time.Time{})
	if err != nil {
		color.Yellow("⚠️  Repository access test failed: %v", err)
	} else {
		color.Green("✓ Repository access successful")
		color.White("Found %d accessible repositories", len(repos))
		
		// Show a few recent repositories
		if len(repos) > 0 {
			color.White("Recent repositories:")
			for i, repo := range repos {
				if i >= 3 { // Show only first 3
					break
				}
				visibility := "public"
				if repo.Private {
					visibility = "private"
				}
				color.White("  - %s (%s)", repo.FullName, visibility)
			}
		}
	}

	return nil
}

func disconnectGitHub(cmd *cobra.Command) error {
	authManager := github.NewAuthManager("")
	
	if !authManager.IsAuthenticated() {
		color.Yellow("GitHub is not connected")
		return nil
	}

	color.Cyan("🔌 Disconnecting from GitHub...")

	if err := authManager.ClearAuthentication(); err != nil {
		return fmt.Errorf("failed to clear GitHub authentication: %w", err)
	}

	color.Green("✓ GitHub disconnected successfully")
	color.White("GitHub authentication has been removed")

	return nil
}