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
	"my-day/internal/jira"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync tickets from Jira",
	Long: `Sync pulls your latest tickets from Jira and stores them locally.

This command fetches tickets assigned to you or created by you from
the configured project keys and caches them for report generation.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := syncTickets(cmd); err != nil {
			color.Red("Sync failed: %v", err)
			os.Exit(1)
		}
	},
}

// TicketCache represents the cached ticket data
type TicketCache struct {
	LastSync time.Time   `json:"last_sync"`
	Issues   []jira.Issue `json:"issues"`
	Worklogs []jira.WorklogEntry `json:"worklogs"`
}

func init() {
	rootCmd.AddCommand(syncCmd)
	
	// Sync-specific flags
	syncCmd.Flags().Int("max-results", 100, "Maximum number of tickets to fetch")
	syncCmd.Flags().Bool("force", false, "Force sync even if recently synced")
	syncCmd.Flags().Bool("worklog", true, "Include worklog entries")
	syncCmd.Flags().Duration("since", 7*24*time.Hour, "Sync tickets updated since this duration ago")
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

	client := jira.NewClient(
		cfg.Jira.BaseURL,
		cfg.Jira.OAuth.ClientID,
		cfg.Jira.OAuth.ClientSecret,
		cfg.Jira.OAuth.RedirectURI,
	)

	// Check authentication
	ctx := context.Background()
	if !client.GetAuthManager().IsAuthenticated(ctx) {
		return fmt.Errorf("not authenticated with Jira. Run 'my-day auth' first")
	}

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
	
	// Extract project keys from configuration
	var projectKeys []string
	for _, project := range cfg.Jira.Projects {
		projectKeys = append(projectKeys, project.Key)
	}

	if len(projectKeys) == 0 {
		color.Yellow("No project keys configured. Add projects to your config file.")
		return nil
	}

	color.White("Fetching tickets from projects: %v", projectKeys)

	// Fetch recent issues
	searchResponse, err := client.GetMyRecentIssues(ctx, projectKeys, maxResults)
	if err != nil {
		return fmt.Errorf("failed to fetch issues: %w", err)
	}

	color.Green("✓ Fetched %d issues", len(searchResponse.Issues))

	// Fetch worklog if enabled
	var worklogs []jira.WorklogEntry
	if includeWorklog, _ := cmd.Flags().GetBool("worklog"); includeWorklog {
		since, _ := cmd.Flags().GetDuration("since")
		sinceTime := time.Now().Add(-since)
		
		color.White("Fetching worklog entries since %s...", sinceTime.Format("2006-01-02"))
		
		worklogs, err = client.GetMyWorklog(ctx, sinceTime)
		if err != nil {
			color.Yellow("Warning: Failed to fetch worklog: %v", err)
			worklogs = []jira.WorklogEntry{} // Continue without worklog
		} else {
			color.Green("✓ Fetched %d worklog entries", len(worklogs))
		}
	}

	// Create cache
	cache := TicketCache{
		LastSync: time.Now(),
		Issues:   searchResponse.Issues,
		Worklogs: worklogs,
	}

	// Save to cache file
	if err := saveCache(cacheFile, &cache); err != nil {
		return fmt.Errorf("failed to save cache: %w", err)
	}

	color.Green("✓ Sync completed successfully")
	color.White("Issues: %d", len(cache.Issues))
	color.White("Worklog entries: %d", len(cache.Worklogs))
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
	if len(cache.Issues) == 0 {
		return
	}

	fmt.Println()
	color.Cyan("📊 Recent Activity Summary")

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
	count := 0
	for _, issue := range cache.Issues {
		if count >= 5 {
			break
		}
		
		timeSince := time.Since(issue.Fields.Updated)
		color.White("  %s - %s (%v ago)", 
			issue.Key, 
			truncateString(issue.Fields.Summary, 50),
			timeSince.Round(time.Hour))
		count++
	}

	if len(cache.Issues) > 5 {
		color.White("  ... and %d more", len(cache.Issues)-5)
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}