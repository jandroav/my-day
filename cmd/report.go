package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"my-day/internal/config"
	"my-day/internal/jira"
	"my-day/internal/report"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate daily standup report",
	Long: `Generate generates a daily standup report based on your recent Jira activity.

The report includes tickets you've worked on, their current status, and
optionally AI-generated summaries of your progress.

Data comes from the local cache populated by 'my-day sync'. Use --since to 
filter which tickets are included based on their last update time.

Reports are automatically cached to improve performance and reduce LLM API calls.
Use --no-cache to disable caching or --cache-only to use only cached reports.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := generateReport(cmd); err != nil {
			color.Red("Report generation failed: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
	
	// Report-specific flags
	reportCmd.Flags().String("date", "", "Generate report for specific date (YYYY-MM-DD)")
	reportCmd.Flags().String("output", "", "Output file path (default: stdout)")
	reportCmd.Flags().Bool("no-llm", false, "Disable LLM summarization for this report")
	reportCmd.Flags().Bool("detailed", false, "Include detailed ticket information")
	reportCmd.Flags().Bool("debug", false, "Enable debug output for LLM processing")
	reportCmd.Flags().Bool("show-quality", false, "Show summary quality indicators")
	reportCmd.Flags().Bool("verbose", false, "Show verbose LLM processing information")
	
	// Cache-specific flags
	reportCmd.Flags().Bool("no-cache", false, "Disable report caching (always generate fresh report)")
	reportCmd.Flags().Bool("cache-only", false, "Only use cached reports (fail if no cache exists)")
	
	// Data filtering flags
	reportCmd.Flags().Duration("since", 7*24*time.Hour, "Include tickets and worklogs updated since this duration ago")
	
	// Field grouping flags
	reportCmd.Flags().String("field", "", "Group report by specified Jira custom field (e.g., 'squad', 'team', 'component')")
	
	// Export-specific flags
	reportCmd.Flags().Bool("export", false, "Export report to markdown file")
	reportCmd.Flags().String("export-folder", "", "Folder path for exported reports (overrides config)")
	reportCmd.Flags().StringSlice("export-tags", []string{}, "Additional tags for exported report (overrides config)")
}

func generateReport(cmd *cobra.Command) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Get cache file
	cacheFile, err := getCacheFilePath()
	if err != nil {
		return fmt.Errorf("failed to get cache file path: %w", err)
	}

	// Load cached data
	cache, err := loadCache(cacheFile)
	if err != nil {
		color.Yellow("No cached data found. Run 'my-day sync' first.")
		return fmt.Errorf("failed to load cache: %w", err)
	}

	// Check cache age
	if time.Since(cache.LastSync) > 24*time.Hour {
		color.Yellow("Cache is older than 24 hours. Consider running 'my-day sync' for fresh data.")
	}

	// Parse date flag
	var targetDate time.Time
	if dateStr, _ := cmd.Flags().GetString("date"); dateStr != "" {
		targetDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return fmt.Errorf("invalid date format. Use YYYY-MM-DD: %w", err)
		}
	} else {
		targetDate = time.Now()
	}
	
	// Get flags for feedback
	debug, _ := cmd.Flags().GetBool("debug")
	verbose, _ := cmd.Flags().GetBool("verbose")
	
	// Filter cached data based on --since flag
	since, _ := cmd.Flags().GetDuration("since")
	sinceTime := time.Now().Add(-since)
	originalIssueCount := len(cache.IssuesWithComments)
	cache = filterCacheDataBySince(cache, sinceTime, targetDate)
	
	if verbose || debug {
		color.White("Filtered from %d to %d issues using --since %v", originalIssueCount, len(cache.IssuesWithComments), since)
	}

	// Determine LLM settings
	llmEnabled := cfg.LLM.Enabled
	if noLLM, _ := cmd.Flags().GetBool("no-llm"); noLLM {
		llmEnabled = false
	}

	detailed, _ := cmd.Flags().GetBool("detailed")
	showQuality, _ := cmd.Flags().GetBool("show-quality")
	groupByField, _ := cmd.Flags().GetString("field")
	
	// Cache flags
	noCache, _ := cmd.Flags().GetBool("no-cache")
	cacheOnly, _ := cmd.Flags().GetBool("cache-only")
	useCache := !noCache
	
	
	// Export flags
	exportEnabled, _ := cmd.Flags().GetBool("export")
	exportFolder, _ := cmd.Flags().GetString("export-folder")
	exportTags, _ := cmd.Flags().GetStringSlice("export-tags")
	
	// Override export settings if flags are provided
	if exportEnabled {
		cfg.Report.Export.Enabled = true
	}
	if exportFolder != "" {
		cfg.Report.Export.FolderPath = exportFolder
	}
	if len(exportTags) > 0 {
		cfg.Report.Export.Tags = exportTags
	}

	// Create report generator
	generator := report.NewGenerator(&report.Config{
		Format:            cfg.Report.Format,
		LLMEnabled:        llmEnabled,
		LLMMode:           cfg.LLM.Mode,
		LLMModel:          cfg.LLM.Model,
		OllamaURL:         cfg.LLM.Ollama.BaseURL,
		OllamaModel:       cfg.LLM.Ollama.Model,
		IncludeYesterday:  cfg.Report.IncludeYesterday,
		IncludeToday:      cfg.Report.IncludeToday,
		IncludeInProgress: cfg.Report.IncludeInProgress,
		Detailed:          detailed,
		Debug:             debug,
		ShowQuality:       showQuality,
		Verbose:           verbose,
		GroupByField:      groupByField,
		ExportEnabled:     cfg.Report.Export.Enabled,
		ExportFolderPath:  cfg.Report.Export.FolderPath,
		ExportFileDate:    cfg.Report.Export.FileNameDate,
		ExportTags:        cfg.Report.Export.Tags,
	})

	color.Cyan("📋 Generating daily standup report...")
	color.White("Showing tickets with your comments today")
	if dateStr, _ := cmd.Flags().GetString("date"); dateStr != "" {
		color.White("Report date: %s", targetDate.Format("2006-01-02"))
	} else {
		color.White("Report date: %s (today)", targetDate.Format("2006-01-02"))
	}
	color.White("Including tickets updated since: %s (last %v)", sinceTime.Format("2006-01-02 15:04"), since)

	// Generate report with comments if available, using caching
	var reportContent string
	
	if len(cache.IssuesWithComments) > 0 {
		// Convert to report package type
		var reportIssuesWithComments []report.IssueWithComments
		for _, iwc := range cache.IssuesWithComments {
			reportIssuesWithComments = append(reportIssuesWithComments, report.IssueWithComments{
				Issue:    iwc.Issue,
				Comments: iwc.Comments,
			})
		}
		
		// Check if cache-only mode and no cache exists
		if cacheOnly {
			cacheManager := generator.GetCacheManager()
			if cacheManager != nil {
				// Create comments map for cache lookup
				commentsMap := make(map[string][]jira.Comment)
				var issues []jira.Issue
				for _, iwc := range reportIssuesWithComments {
					issues = append(issues, iwc.Issue)
					commentsMap[iwc.Issue.Key] = iwc.Comments
				}
				
				cachedReport, cacheErr := cacheManager.FindReport(generator.GetConfig(), issues, commentsMap, cache.Worklogs, targetDate)
				if cacheErr != nil || cachedReport == nil {
					return fmt.Errorf("no cached report found for %s (cache-only mode)", targetDate.Format("2006-01-02"))
				}
				reportContent = cachedReport.Content
			} else {
				return fmt.Errorf("cache manager not available (cache-only mode)")
			}
		} else {
			// Use the new caching-aware generation method
			reportContent, err = generator.GenerateWithCommentsAndCache(reportIssuesWithComments, cache.Worklogs, targetDate, useCache)
		}
	} else {
		// Fallback to basic report generation with caching
		if cacheOnly {
			cacheManager := generator.GetCacheManager()
			if cacheManager != nil {
				commentsMap := make(map[string][]jira.Comment)
				cachedReport, cacheErr := cacheManager.FindReport(generator.GetConfig(), cache.Issues, commentsMap, cache.Worklogs, targetDate)
				if cacheErr != nil || cachedReport == nil {
					return fmt.Errorf("no cached report found for %s (cache-only mode)", targetDate.Format("2006-01-02"))
				}
				reportContent = cachedReport.Content
			} else {
				return fmt.Errorf("cache manager not available (cache-only mode)")
			}
		} else {
			reportContent, err = generator.GenerateWithCache(cache.Issues, cache.Worklogs, targetDate, useCache)
		}
	}
	
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Handle export to Obsidian if enabled
	if err := generator.ExportToObsidian(reportContent, targetDate); err != nil {
		color.Yellow("⚠️  Export to Obsidian failed: %v", err)
	} else if cfg.Report.Export.Enabled || exportEnabled {
		exportPath := cfg.Report.Export.FolderPath
		if exportFolder != "" {
			exportPath = exportFolder
		}
		filename := targetDate.Format(cfg.Report.Export.FileNameDate) + ".md"
		color.Green("✓ Report exported to Obsidian: %s/%s", exportPath, filename)
	}

	// Handle output
	if outputFile, _ := cmd.Flags().GetString("output"); outputFile != "" {
		if err := os.WriteFile(outputFile, []byte(reportContent), 0644); err != nil {
			return fmt.Errorf("failed to write report to file: %w", err)
		}
		color.Green("✓ Report saved to: %s", outputFile)
	} else {
		fmt.Print(reportContent)
	}

	return nil
}

// filterCacheDataBySince filters cached data based on the since duration
func filterCacheDataBySince(cache *TicketCache, sinceTime time.Time, targetDate time.Time) *TicketCache {
	// Create a new cache with filtered data
	filteredCache := &TicketCache{
		LastSync:           cache.LastSync,
		Issues:             []jira.Issue{},
		IssuesWithComments: []IssueWithComments{},
		Worklogs:           []jira.WorklogEntry{},
	}
	
	// Filter issues based on update time
	for _, issue := range cache.Issues {
		if issue.Fields.Updated.Time.After(sinceTime) {
			filteredCache.Issues = append(filteredCache.Issues, issue)
		}
	}
	
	// Filter issues with comments based on issue update time and comment creation time
	for _, iwc := range cache.IssuesWithComments {
		// Check if the issue itself was updated within the since period
		if iwc.Issue.Fields.Updated.Time.After(sinceTime) {
			// Filter comments to only include those within the since period or target date
			var filteredComments []jira.Comment
			todayStart := targetDate.Truncate(24 * time.Hour)
			todayEnd := todayStart.Add(24 * time.Hour)
			
			for _, comment := range iwc.Comments {
				// Include comments from target date or within since period
				if (comment.Created.Time.After(todayStart) && comment.Created.Time.Before(todayEnd)) ||
				   comment.Created.Time.After(sinceTime) {
					filteredComments = append(filteredComments, comment)
				}
			}
			
			// Only include the issue if it has filtered comments or was recently updated
			if len(filteredComments) > 0 || iwc.Issue.Fields.Updated.Time.After(sinceTime) {
				filteredCache.IssuesWithComments = append(filteredCache.IssuesWithComments, IssueWithComments{
					Issue:    iwc.Issue,
					Comments: filteredComments,
				})
			}
		}
	}
	
	// Filter worklogs based on start time
	for _, worklog := range cache.Worklogs {
		if worklog.Started.Time.After(sinceTime) {
			filteredCache.Worklogs = append(filteredCache.Worklogs, worklog)
		}
	}
	
	return filteredCache
}
