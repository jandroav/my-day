package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"my-day/internal/report"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export cached reports to files",
	Long: `Export exports cached reports to files without regenerating them.

This command allows you to export previously generated reports to various formats
and locations without needing to call the LLM again. You can export specific
dates, date ranges, or all cached reports.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := exportReports(cmd); err != nil {
			color.Red("Export failed: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
	
	// Export-specific flags
	exportCmd.Flags().String("date", "", "Export report for specific date (YYYY-MM-DD)")
	exportCmd.Flags().String("from", "", "Export reports from this date (YYYY-MM-DD)")
	exportCmd.Flags().String("to", "", "Export reports to this date (YYYY-MM-DD)")
	exportCmd.Flags().StringSlice("format", []string{"markdown"}, "Export formats (markdown, console)")
	exportCmd.Flags().String("output-dir", "", "Output directory (default: current directory)")
	exportCmd.Flags().Bool("list", false, "List available cached reports")
	exportCmd.Flags().Bool("force", false, "Overwrite existing files")
	exportCmd.Flags().String("filename-template", "{{.Date}}_{{.Format}}", "Filename template (supports {{.Date}}, {{.Format}}, {{.ID}})")
}

func exportReports(cmd *cobra.Command) error {

	// Initialize cache manager
	cacheManager, err := report.NewCacheManager()
	if err != nil {
		return fmt.Errorf("failed to initialize cache manager: %w", err)
	}

	// Handle list command
	if list, _ := cmd.Flags().GetBool("list"); list {
		return listCachedReports(cacheManager)
	}

	// Parse date filters
	var fromDate, toDate *time.Time
	if dateStr, _ := cmd.Flags().GetString("date"); dateStr != "" {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return fmt.Errorf("invalid date format. Use YYYY-MM-DD: %w", err)
		}
		fromDate = &date
		toDate = &date
	} else {
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
	}

	// Get export parameters
	formats, _ := cmd.Flags().GetStringSlice("format")
	outputDir, _ := cmd.Flags().GetString("output-dir")
	if outputDir == "" {
		outputDir = "."
	}
	force, _ := cmd.Flags().GetBool("force")
	filenameTemplate, _ := cmd.Flags().GetString("filename-template")

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get list of reports to export
	reports, err := cacheManager.ListReports(fromDate, toDate)
	if err != nil {
		return fmt.Errorf("failed to list cached reports: %w", err)
	}

	if len(reports) == 0 {
		color.Yellow("No cached reports found for the specified criteria")
		return nil
	}

	color.Cyan("📋 Exporting %d cached reports...", len(reports))

	exportedCount := 0
	for _, reportEntry := range reports {
		// Load the full report
		cachedReport, err := cacheManager.LoadReport(reportEntry.ID)
		if err != nil {
			color.Yellow("Warning: Failed to load report %s: %v", reportEntry.ID, err)
			continue
		}

		// Export in requested formats
		for _, format := range formats {
			// Skip if the cached report is in a different format and we can't convert
			if cachedReport.Format != format && format != "console" {
				if format == "markdown" && cachedReport.Format == "console" {
					color.Yellow("Warning: Cannot convert console format to markdown for report %s", reportEntry.ID)
					continue
				}
			}

			// Generate filename
			filename, err := generateFilename(filenameTemplate, reportEntry, format)
			if err != nil {
				color.Yellow("Warning: Failed to generate filename for report %s: %v", reportEntry.ID, err)
				continue
			}

			outputPath := filepath.Join(outputDir, filename)

			// Check if file exists and force flag
			if _, err := os.Stat(outputPath); err == nil && !force {
				color.Yellow("Skipping %s (file exists, use --force to overwrite)", outputPath)
				continue
			}

			// Export the report
			content := cachedReport.Content
			if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
				color.Yellow("Warning: Failed to write report to %s: %v", outputPath, err)
				continue
			}

			// Update cache with export path
			if err := cacheManager.UpdateExportPath(reportEntry.ID, format, outputPath); err != nil {
				color.Yellow("Warning: Failed to update export path in cache: %v", err)
			}

			color.Green("✓ Exported %s", outputPath)
			exportedCount++
		}
	}

	color.Green("✓ Export completed. %d files exported to %s", exportedCount, outputDir)
	return nil
}

func listCachedReports(cacheManager *report.CacheManager) error {
	reports, err := cacheManager.ListReports(nil, nil)
	if err != nil {
		return fmt.Errorf("failed to list cached reports: %w", err)
	}

	if len(reports) == 0 {
		color.Yellow("No cached reports found")
		return nil
	}

	color.Cyan("📋 Cached Reports (%d total)", len(reports))
	color.White(strings.Repeat("=", 80))

	for _, report := range reports {
		statusIcon := "✅"
		if !report.LLMUsed {
			statusIcon = "📝"
		}

		color.White("%s %s [%s] - %d issues, %d comments", 
			statusIcon, 
			report.Date, 
			report.Format,
			report.IssueCount,
			report.CommentCount)
		
		color.Cyan("   ID: %s", report.ID)
		color.Cyan("   Generated: %s", report.GeneratedAt.Format("2006-01-02 15:04:05"))
		
		if len(report.ExportPaths) > 0 {
			color.Green("   Exported to:")
			for format, path := range report.ExportPaths {
				color.Green("     %s: %s", format, path)
			}
		}
		
		fmt.Println()
	}

	// Show cache statistics
	stats, err := cacheManager.GetCacheStats()
	if err == nil {
		color.Cyan("\n📊 Cache Statistics")
		color.White("Total reports: %d", stats["total_reports"])
		color.White("Cache size: %.2f KB", float64(stats["cache_size_bytes"].(int64))/1024)
		color.White("LLM usage: %d reports", stats["llm_usage_count"])
	}

	return nil
}

func generateFilename(template string, report report.ReportCacheEntry, format string) (string, error) {
	filename := template
	
	// Replace template variables
	filename = strings.ReplaceAll(filename, "{{.Date}}", report.Date)
	filename = strings.ReplaceAll(filename, "{{.Format}}", format)
	filename = strings.ReplaceAll(filename, "{{.ID}}", report.ID)
	
	// Add appropriate extension
	switch format {
	case "markdown":
		if !strings.HasSuffix(filename, ".md") {
			filename += ".md"
		}
	case "console":
		if !strings.HasSuffix(filename, ".txt") {
			filename += ".txt"
		}
	}
	
	// Ensure filename is safe
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, ":", "_")
	
	return filename, nil
}