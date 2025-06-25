package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"my-day/internal/config"
	"my-day/internal/report"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate daily standup report",
	Long: `Generate generates a daily standup report based on your recent Jira activity.

The report includes tickets you've worked on, their current status, and
optionally AI-generated summaries of your progress.`,
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

	// Determine LLM settings
	llmEnabled := cfg.LLM.Enabled
	if noLLM, _ := cmd.Flags().GetBool("no-llm"); noLLM {
		llmEnabled = false
	}

	detailed, _ := cmd.Flags().GetBool("detailed")

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
	})

	color.Cyan("📋 Generating daily standup report...")
	if dateStr, _ := cmd.Flags().GetString("date"); dateStr != "" {
		color.White("Report date: %s", targetDate.Format("2006-01-02"))
	} else {
		color.White("Report date: %s (today)", targetDate.Format("2006-01-02"))
	}

	// Generate report
	reportContent, err := generator.Generate(cache.Issues, cache.Worklogs, targetDate)
	if err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
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