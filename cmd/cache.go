package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"my-day/internal/report"
)

// cacheCmd represents the cache command
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage report cache",
	Long: `Cache manages the report cache system.

The cache stores generated reports to avoid calling the LLM repeatedly for the same data.
This command provides subcommands to list, clear, and inspect cached reports.`,
}

// cacheListCmd lists cached reports
var cacheListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cached reports",
	Long: `List displays all cached reports with their metadata.

Shows information about each cached report including date, format, generation time,
LLM usage, and export paths.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := listCache(cmd); err != nil {
			color.Red("Failed to list cache: %v", err)
			os.Exit(1)
		}
	},
}

// cacheClearCmd clears the report cache
var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear cached reports",
	Long: `Clear removes cached reports from the system.

You can clear all reports or specify criteria to clear specific reports.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := clearCache(cmd); err != nil {
			color.Red("Failed to clear cache: %v", err)
			os.Exit(1)
		}
	},
}

// cacheStatsCmd shows cache statistics
var cacheStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show cache statistics",
	Long: `Stats displays detailed statistics about the report cache.

Shows information about cache size, number of reports, LLM usage patterns,
and storage details.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := showCacheStats(cmd); err != nil {
			color.Red("Failed to show cache stats: %v", err)
			os.Exit(1)
		}
	},
}

// cacheDeleteCmd deletes specific cached reports
var cacheDeleteCmd = &cobra.Command{
	Use:   "delete [report-id...]",
	Short: "Delete specific cached reports",
	Long: `Delete removes specific cached reports by their ID.

You can specify one or more report IDs to delete. Use 'my-day cache list' 
to see available report IDs.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := deleteCache(cmd, args); err != nil {
			color.Red("Failed to delete cache: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	
	// Add subcommands
	cacheCmd.AddCommand(cacheListCmd)
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheStatsCmd)
	cacheCmd.AddCommand(cacheDeleteCmd)
	
	// Flags for list command
	cacheListCmd.Flags().String("from", "", "List reports from this date (YYYY-MM-DD)")
	cacheListCmd.Flags().String("to", "", "List reports to this date (YYYY-MM-DD)")
	cacheListCmd.Flags().String("format", "", "Filter by format (markdown, console)")
	cacheListCmd.Flags().Bool("llm-only", false, "Show only reports generated with LLM")
	
	// Flags for clear command
	cacheClearCmd.Flags().Bool("all", false, "Clear all cached reports")
	cacheClearCmd.Flags().String("before", "", "Clear reports before this date (YYYY-MM-DD)")
	cacheClearCmd.Flags().Bool("force", false, "Skip confirmation prompt")
	
	// Flags for delete command
	cacheDeleteCmd.Flags().Bool("force", false, "Skip confirmation prompt")
}

func listCache(cmd *cobra.Command) error {
	cacheManager, err := report.NewCacheManager()
	if err != nil {
		return fmt.Errorf("failed to initialize cache manager: %w", err)
	}

	// Parse date filters
	var fromDate, toDate *time.Time
	if fromStr, _ := cmd.Flags().GetString("from"); fromStr != "" {
		date, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			return fmt.Errorf("invalid from date format. Use YYYY-MM-DD: %w", err)
		}
		fromDate = &date
	}
	if toStr, _ := cmd.Flags().GetString("to"); toStr != "" {
		date, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			return fmt.Errorf("invalid to date format. Use YYYY-MM-DD: %w", err)
		}
		toDate = &date
	}

	reports, err := cacheManager.ListReports(fromDate, toDate)
	if err != nil {
		return fmt.Errorf("failed to list cached reports: %w", err)
	}

	// Apply additional filters
	formatFilter, _ := cmd.Flags().GetString("format")
	llmOnly, _ := cmd.Flags().GetBool("llm-only")
	
	var filteredReports []report.ReportCacheEntry
	for _, r := range reports {
		if formatFilter != "" && r.Format != formatFilter {
			continue
		}
		if llmOnly && !r.LLMUsed {
			continue
		}
		filteredReports = append(filteredReports, r)
	}

	if len(filteredReports) == 0 {
		color.Yellow("No cached reports found matching the criteria")
		return nil
	}

	color.Cyan("📋 Cached Reports (%d total, %d matching filters)", len(reports), len(filteredReports))
	color.White(strings.Repeat("=", 80))

	for _, reportEntry := range filteredReports {
		statusIcon := "✅"
		if !reportEntry.LLMUsed {
			statusIcon = "📝"
		}

		color.White("%s %s [%s] - %d issues, %d comments", 
			statusIcon, 
			reportEntry.Date, 
			reportEntry.Format,
			reportEntry.IssueCount,
			reportEntry.CommentCount)
		
		color.Cyan("   ID: %s", reportEntry.ID)
		color.Cyan("   Generated: %s", reportEntry.GeneratedAt.Format("2006-01-02 15:04:05"))
		
		if len(reportEntry.ExportPaths) > 0 {
			color.Green("   Exported to:")
			for format, path := range reportEntry.ExportPaths {
				color.Green("     %s: %s", format, path)
			}
		}
		
		fmt.Println()
	}

	return nil
}

func clearCache(cmd *cobra.Command) error {
	cacheManager, err := report.NewCacheManager()
	if err != nil {
		return fmt.Errorf("failed to initialize cache manager: %w", err)
	}

	clearAll, _ := cmd.Flags().GetBool("all")
	beforeStr, _ := cmd.Flags().GetString("before")
	force, _ := cmd.Flags().GetBool("force")

	if !clearAll && beforeStr == "" {
		return fmt.Errorf("must specify either --all or --before date")
	}

	// Get list of reports to delete
	var reportsToDelete []report.ReportCacheEntry
	
	if clearAll {
		reports, err := cacheManager.ListReports(nil, nil)
		if err != nil {
			return fmt.Errorf("failed to list reports: %w", err)
		}
		reportsToDelete = reports
	} else {
		beforeDate, err := time.Parse("2006-01-02", beforeStr)
		if err != nil {
			return fmt.Errorf("invalid before date format. Use YYYY-MM-DD: %w", err)
		}
		
		reports, err := cacheManager.ListReports(nil, &beforeDate)
		if err != nil {
			return fmt.Errorf("failed to list reports: %w", err)
		}
		
		for _, r := range reports {
			if r.GeneratedAt.Before(beforeDate) {
				reportsToDelete = append(reportsToDelete, r)
			}
		}
	}

	if len(reportsToDelete) == 0 {
		color.Yellow("No reports found to delete")
		return nil
	}

	// Confirm deletion
	if !force {
		color.Yellow("This will delete %d cached reports:", len(reportsToDelete))
		for _, r := range reportsToDelete {
			color.White("  - %s [%s] (generated %s)", r.Date, r.Format, r.GeneratedAt.Format("2006-01-02"))
		}
		
		fmt.Print("\nAre you sure? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			color.Yellow("Cancelled")
			return nil
		}
	}

	// Delete reports
	deleteCount := 0
	for _, r := range reportsToDelete {
		if err := cacheManager.DeleteReport(r.ID); err != nil {
			color.Yellow("Warning: Failed to delete report %s: %v", r.ID, err)
		} else {
			deleteCount++
		}
	}

	color.Green("✓ Deleted %d cached reports", deleteCount)
	return nil
}

func showCacheStats(cmd *cobra.Command) error {
	cacheManager, err := report.NewCacheManager()
	if err != nil {
		return fmt.Errorf("failed to initialize cache manager: %w", err)
	}

	stats, err := cacheManager.GetCacheStats()
	if err != nil {
		return fmt.Errorf("failed to get cache statistics: %w", err)
	}

	color.Cyan("📊 Cache Statistics")
	color.White(strings.Repeat("=", 50))

	// Basic stats
	color.White("Total reports: %d", stats["total_reports"])
	color.White("Cache directory: %s", stats["cache_directory"])
	
	cacheSizeBytes := stats["cache_size_bytes"].(int64)
	if cacheSizeBytes < 1024 {
		color.White("Cache size: %d bytes", cacheSizeBytes)
	} else if cacheSizeBytes < 1024*1024 {
		color.White("Cache size: %.2f KB", float64(cacheSizeBytes)/1024)
	} else {
		color.White("Cache size: %.2f MB", float64(cacheSizeBytes)/(1024*1024))
	}

	color.White("LLM usage: %d reports", stats["llm_usage_count"])

	// Reports by date
	if dateGroups, ok := stats["reports_by_date"].(map[string]int); ok && len(dateGroups) > 0 {
		color.Cyan("\n📅 Reports by Date:")
		for date, count := range dateGroups {
			color.White("  %s: %d reports", date, count)
		}
	}

	// Reports by format
	if formatGroups, ok := stats["reports_by_format"].(map[string]int); ok && len(formatGroups) > 0 {
		color.Cyan("\n📄 Reports by Format:")
		for format, count := range formatGroups {
			color.White("  %s: %d reports", format, count)
		}
	}

	return nil
}

func deleteCache(cmd *cobra.Command, reportIDs []string) error {
	cacheManager, err := report.NewCacheManager()
	if err != nil {
		return fmt.Errorf("failed to initialize cache manager: %w", err)
	}

	force, _ := cmd.Flags().GetBool("force")

	// Verify reports exist
	var reportsToDelete []string
	for _, id := range reportIDs {
		if _, err := cacheManager.LoadReport(id); err != nil {
			color.Yellow("Warning: Report %s not found, skipping", id)
		} else {
			reportsToDelete = append(reportsToDelete, id)
		}
	}

	if len(reportsToDelete) == 0 {
		color.Yellow("No valid reports found to delete")
		return nil
	}

	// Confirm deletion
	if !force {
		color.Yellow("This will delete %d cached reports:", len(reportsToDelete))
		for _, id := range reportsToDelete {
			color.White("  - %s", id)
		}
		
		fmt.Print("\nAre you sure? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			color.Yellow("Cancelled")
			return nil
		}
	}

	// Delete reports
	deleteCount := 0
	for _, id := range reportsToDelete {
		if err := cacheManager.DeleteReport(id); err != nil {
			color.Yellow("Warning: Failed to delete report %s: %v", id, err)
		} else {
			color.Green("✓ Deleted report %s", id)
			deleteCount++
		}
	}

	color.Green("✓ Deleted %d cached reports", deleteCount)
	return nil
}