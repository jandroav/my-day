package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"my-day/internal/config"
	"my-day/internal/github"
	"my-day/internal/jira"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync tickets from Jira and GitHub",
	Long: `Sync pulls your latest tickets from Jira and GitHub activity and stores them locally.

This command fetches tickets assigned to you or created by you from
the configured project keys and GitHub repositories, then caches them for report generation.

GitHub integration includes:
- Pull requests (authored, assigned, or reviewed)
- Commits (authored by you)
- Workflow runs and CI/CD status`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := syncTickets(cmd); err != nil {
			color.Red("Sync failed: %v", err)
			os.Exit(1)
		}
	},
}

// IssueWithComments represents an issue with today's comments
type IssueWithComments struct {
	Issue    jira.Issue     `json:"issue"`
	Comments []jira.Comment `json:"comments"`
}

// TicketCache represents the cached ticket data
type TicketCache struct {
	LastSync           time.Time              `json:"last_sync"`
	Issues             []jira.Issue           `json:"issues"`
	IssuesWithComments []IssueWithComments    `json:"issues_with_comments"`
	Worklogs           []jira.WorklogEntry    `json:"worklogs"`
	GitHubActivity     []github.Activity      `json:"github_activity"`
	LastGitHubSync     time.Time              `json:"last_github_sync"`
}

func init() {
	rootCmd.AddCommand(syncCmd)
	
	// Sync-specific flags
	syncCmd.Flags().Int("max-results", 100, "Maximum number of tickets to fetch")
	syncCmd.Flags().Bool("force", false, "Force sync even if recently synced")
	syncCmd.Flags().Bool("worklog", true, "Include worklog entries")
	syncCmd.Flags().Duration("since", 7*24*time.Hour, "Fetch tickets and worklogs updated since this duration ago")
	syncCmd.Flags().Duration("comments-since", 24*time.Hour, "Look for your comments within this duration (defaults to --since value if not specified)")
	syncCmd.Flags().StringSlice("platforms", []string{"jira", "github"}, "Platforms to sync (jira, github)")
	syncCmd.Flags().Bool("github", true, "Include GitHub activity (if connected and enabled)")
}

func syncTickets(cmd *cobra.Command) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Validate configuration
	if cfg.Jira.BaseURL == "" {
		return fmt.Errorf("Jira base URL not configured. Run 'my-day init' first")
	}

	// Create temporary auth manager to check authentication
	authManager := jira.NewAuthManager("", "")
	if !authManager.IsAuthenticated() {
		return fmt.Errorf("not authenticated with Jira. Run 'my-day auth --email your-email --token your-token' first")
	}

	// Load API token and create client
	apiToken, err := authManager.LoadAPIToken()
	if err != nil {
		return fmt.Errorf("failed to load API token: %w", err)
	}

	client := jira.NewClient(cfg.Jira.BaseURL, apiToken.Email, apiToken.Token)
	ctx := context.Background()

	// Get cache file path
	cacheFile, err := getCacheFilePath()
	if err != nil {
		return fmt.Errorf("failed to get cache file path: %w", err)
	}

	// Check if we need to sync
	force, _ := cmd.Flags().GetBool("force")
	if !force {
		if cache, err := loadCache(cacheFile); err == nil {
			if time.Since(cache.LastSync) < 10*time.Minute {
				color.Yellow("Recently synced (%v ago). Use --force to sync anyway.", 
					time.Since(cache.LastSync).Round(time.Minute))
				return nil
			}
		}
	}

	color.Cyan("🔄 Syncing tickets from Jira...")

	maxResults, _ := cmd.Flags().GetInt("max-results")
	
	// Get project keys directly from configuration (already a slice of strings)
	projectKeys := cfg.Jira.Projects

	if len(projectKeys) == 0 {
		color.Yellow("No project keys configured. Add projects to your config file.")
		return nil
	}

	color.White("Fetching tickets from projects: %v", projectKeys)

	// Fetch issues with recent updates (using --since flag)
	since, _ := cmd.Flags().GetDuration("since")
	ticketsSinceTime := time.Now().Add(-since)
	
	color.White("Searching for tickets updated since %s...", ticketsSinceTime.Format("2006-01-02"))
	searchResponse, err := client.GetMyIssuesWithTodaysComments(ctx, projectKeys, maxResults, ticketsSinceTime)
	if err != nil {
		return fmt.Errorf("failed to fetch issues: %w", err)
	}

	color.Green("✓ Found %d updated issues to check for your comments", len(searchResponse.Issues))

	// Fetch comments for each issue (using --comments-since flag)
	commentsSince, _ := cmd.Flags().GetDuration("comments-since")
	
	// If comments-since wasn't explicitly set, use the same duration as --since
	if !cmd.Flags().Changed("comments-since") {
		commentsSince = since
	}
	
	commentsSinceTime := time.Now().Add(-commentsSince)
	
	color.White("Fetching your comments from the last %v...", commentsSince)
	var issuesWithComments []IssueWithComments
	
	// Get current user info for comment filtering
	userInfo, err := client.GetCurrentUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}
	
	if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
		color.White("Looking for comments by user: %s (ID: %s)", userInfo.DisplayName, userInfo.AccountID)
		color.White("Filtering for comments after: %s", commentsSinceTime.Format("2006-01-02 15:04:05"))
	}
	
	for _, issue := range searchResponse.Issues {
		allComments, err := client.GetIssueComments(ctx, issue.Key)
		if err != nil {
			color.Yellow("Warning: Failed to fetch comments for %s: %v", issue.Key, err)
			allComments = []jira.Comment{} // Continue without comments for this issue
		}
		
		// Filter comments to only include today's comments by the current user
		var todaysComments []jira.Comment
		for _, comment := range allComments {
			if verbose, _ := cmd.Flags().GetBool("verbose"); verbose && len(allComments) > 0 {
				color.White("  Comment by %s (%s) at %s", 
					comment.Author.DisplayName, 
					comment.Author.AccountID,
					comment.Created.Time.Format("2006-01-02 15:04:05"))
			}
			
			if comment.Author.AccountID == userInfo.AccountID && 
			   comment.Created.Time.After(commentsSinceTime) {
				todaysComments = append(todaysComments, comment)
				if verbose, _ := cmd.Flags().GetBool("verbose"); verbose {
					color.Green("    ✓ This comment matches!")
				}
			}
		}
		
		// Only include issues that have comments from the current user today
		if len(todaysComments) > 0 {
			issuesWithComments = append(issuesWithComments, IssueWithComments{
				Issue:    issue,
				Comments: todaysComments,
			})
		}
	}
	
	if len(issuesWithComments) == 0 {
		color.Yellow("✓ No issues found with your comments in the last %v", commentsSince)
		color.White("  Try adding a comment to a Jira ticket or use --comments-since to look further back.")
		color.White("  Example: my-day sync --comments-since 72h")
	} else {
		color.Green("✓ Found %d issues with your comments in the last %v", len(issuesWithComments), commentsSince)
	}

	// Fetch worklog if enabled
	var worklogs []jira.WorklogEntry
	if includeWorklog, _ := cmd.Flags().GetBool("worklog"); includeWorklog {
		worklogSinceTime := time.Now().Add(-since)
		
		color.White("Fetching worklog entries since %s...", worklogSinceTime.Format("2006-01-02"))
		
		worklogs, err = client.GetMyWorklog(ctx, worklogSinceTime)
		if err != nil {
			color.Yellow("Warning: Failed to fetch worklog: %v", err)
			worklogs = []jira.WorklogEntry{} // Continue without worklog
		} else {
			color.Green("✓ Fetched %d worklog entries", len(worklogs))
		}
	}

	// Extract only the issues that have comments from the current user
	var filteredIssues []jira.Issue
	for _, iwc := range issuesWithComments {
		filteredIssues = append(filteredIssues, iwc.Issue)
	}

	// Fetch GitHub activity if enabled
	var githubActivity []github.Activity
	githubSyncTime := time.Now()
	includeGitHub, _ := cmd.Flags().GetBool("github")
	platforms, _ := cmd.Flags().GetStringSlice("platforms")
	
	if includeGitHub && containsString(platforms, "github") && cfg.GitHub.Enabled {
		color.Cyan("🐙 Syncing GitHub activity...")
		
		githubAuthManager := github.NewAuthManager("")
		if githubAuthManager.IsAuthenticated() {
			authInfo, err := githubAuthManager.LoadToken()
			if err == nil {
				githubClient := github.NewClient(authInfo.Token)
				
				// Fetch GitHub activity since the specified time
				githubSinceTime := time.Now().Add(-since)
				activity, err := githubClient.GetUserActivity(ctx, githubSinceTime, cfg.GitHub.Repositories)
				if err != nil {
					color.Yellow("Warning: Failed to fetch GitHub activity: %v", err)
					githubActivity = []github.Activity{} // Continue without GitHub
				} else {
					githubActivity = activity
					color.Green("✓ Fetched %d GitHub activities", len(githubActivity))
				}
			} else {
				color.Yellow("Warning: GitHub authentication failed: %v", err)
			}
		} else {
			color.Yellow("⚠️  GitHub not authenticated. Run 'my-day github connect' to include GitHub activity")
		}
	} else {
		color.White("GitHub sync disabled or not configured")
	}

	// Create cache
	cache := TicketCache{
		LastSync:           time.Now(),
		Issues:             filteredIssues,
		IssuesWithComments: issuesWithComments,
		Worklogs:           worklogs,
		GitHubActivity:     githubActivity,
		LastGitHubSync:     githubSyncTime,
	}

	// Save to cache file
	if err := saveCache(cacheFile, &cache); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	color.Green("✓ Sync completed successfully")
	color.White("Issues: %d", len(cache.Issues))
	color.White("Worklog entries: %d", len(cache.Worklogs))
	color.White("GitHub activities: %d", len(cache.GitHubActivity))
	color.White("Cache saved to: %s", cacheFile)

	// Show summary of recent activity
	showSyncSummary(&cache)

	return nil
}

func getCacheFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	cacheDir := filepath.Join(homeDir, ".my-day")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(cacheDir, "cache.json"), nil
}

func loadCache(filePath string) (*TicketCache, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var cache TicketCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}

	return &cache, nil
}

func saveCache(filePath string, cache *TicketCache) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

func showSyncSummary(cache *TicketCache) {
	if len(cache.Issues) == 0 && len(cache.GitHubActivity) == 0 {
		return
	}

	fmt.Println()
	color.Cyan("📊 Recent Activity Summary")

	// Show Jira activity if available
	if len(cache.Issues) > 0 {
		// Group issues by status
		statusCounts := make(map[string]int)
		for _, issue := range cache.Issues {
			statusCounts[issue.Fields.Status.Name]++
		}

		color.White("Issues by Status:")
		for status, count := range statusCounts {
			color.White("  %s: %d", status, count)
		}

		// Show recent updates
		color.White("\nRecently Updated Issues:")
		for _, issue := range cache.Issues {
			timeSince := time.Since(issue.Fields.Updated.Time)
			color.White("  %s - %s (%v ago)", 
				issue.Key, 
				truncateString(issue.Fields.Summary, 50),
				timeSince.Round(time.Hour))
		}
	}

	// Show GitHub activity if available
	if len(cache.GitHubActivity) > 0 {
		color.White("\n🐙 GitHub Activity:")
		
		// Group by activity type
		typeCounts := make(map[string]int)
		for _, activity := range cache.GitHubActivity {
			typeCounts[activity.Type]++
		}
		
		for actType, count := range typeCounts {
			color.White("  %s: %d", actType, count)
		}
		
		// Show recent activities (limit to 5)
		color.White("\nRecent GitHub Activities:")
		displayCount := len(cache.GitHubActivity)
		if displayCount > 5 {
			displayCount = 5
		}
		
		for i := 0; i < displayCount; i++ {
			activity := cache.GitHubActivity[i]
			timeSince := time.Since(activity.UpdatedAt)
			icon := getActivityIcon(activity.Type)
			color.White("  %s %s - %s (%v ago)", 
				icon,
				activity.Repository,
				truncateString(activity.Title, 40),
				timeSince.Round(time.Hour))
		}
		
		if len(cache.GitHubActivity) > 5 {
			color.White("  ... and %d more activities", len(cache.GitHubActivity)-5)
		}
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// containsString checks if a slice contains a string
func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getActivityIcon returns an icon for the activity type
func getActivityIcon(activityType string) string {
	switch activityType {
	case "pull_request":
		return "🔀"
	case "commit":
		return "💾"
	case "workflow_run":
		return "⚙️"
	case "issue":
		return "📋"
	case "release":
		return "🚀"
	default:
		return "📝"
	}
}