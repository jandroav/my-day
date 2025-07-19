package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"my-day/internal/config"
	"my-day/internal/jira"
	"my-day/internal/llm"
)

// llmCmd represents the llm command
var llmCmd = &cobra.Command{
	Use:   "llm",
	Short: "Manage LLM integration",
	Long: `Manage LLM (Large Language Model) integration for ticket summarization.

Test connectivity, check status, and manage LLM configuration.`,
}

var llmTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test LLM connectivity",
	Long:  "Test if the configured LLM service is available and working.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := testLLMConnection(); err != nil {
			color.Red("LLM test failed: %v", err)
			os.Exit(1)
		}
	},
}

var llmStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show LLM status",
	Long:  "Display current LLM configuration and status.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := showLLMStatus(); err != nil {
			color.Red("Error showing LLM status: %v", err)
			os.Exit(1)
		}
	},
}

var llmStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Docker LLM container",
	Long:  "Start the Docker LLM container for better summarization.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := startDockerLLM(); err != nil {
			color.Red("Failed to start Docker LLM: %v", err)
			os.Exit(1)
		}
	},
}

var llmStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop Docker LLM container",
	Long:  "Stop the Docker LLM container to free up resources.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := stopDockerLLM(); err != nil {
			color.Red("Failed to stop Docker LLM: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(llmCmd)
	llmCmd.AddCommand(llmTestCmd)
	llmCmd.AddCommand(llmStatusCmd)
	llmCmd.AddCommand(llmStartCmd)
	llmCmd.AddCommand(llmStopCmd)
}

func testLLMConnection() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	llmConfig := llm.LLMConfig{
		Enabled:                  cfg.LLM.Enabled,
		Mode:                     cfg.LLM.Mode,
		Model:                    cfg.LLM.Model,
		Debug:                    cfg.LLM.Debug,
		SummaryStyle:             cfg.LLM.SummaryStyle,
		MaxSummaryLength:         cfg.LLM.MaxSummaryLength,
		IncludeTechnicalDetails:  cfg.LLM.IncludeTechnicalDetails,
		PrioritizeRecentWork:     cfg.LLM.PrioritizeRecentWork,
		FallbackStrategy:         cfg.LLM.FallbackStrategy,
		OllamaURL:                cfg.LLM.Ollama.BaseURL,
		OllamaModel:              cfg.LLM.Ollama.Model,
	}

	color.Cyan("🧠 Testing LLM connectivity...")
	color.White("Mode: %s", llmConfig.Mode)

	if !llmConfig.Enabled {
		color.Yellow("⚠️  LLM is disabled in configuration")
		return nil
	}

	if err := llm.TestLLMConnection(llmConfig); err != nil {
		return fmt.Errorf("LLM connection test failed: %w", err)
	}

	color.Green("✅ LLM connection successful!")

	// Test summarization capability
	color.White("Testing summarization...")
	summarizer, err := llm.NewSummarizer(llmConfig)
	if err != nil {
		return fmt.Errorf("failed to create summarizer: %w", err)
	}

	// Create a test issue
	testIssue := createTestIssue()
	summary, err := summarizer.SummarizeIssue(testIssue)
	if err != nil {
		color.Yellow("⚠️  Summarization test failed: %v", err)
	} else {
		color.Green("✅ Summarization working!")
		color.White("Test summary: %s", summary)
	}

	return nil
}

func showLLMStatus() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	color.Cyan("🧠 LLM Status")
	fmt.Println()

	// Basic configuration
	color.Yellow("Configuration:")
	color.White("  Enabled: %t", cfg.LLM.Enabled)
	color.White("  Mode: %s", cfg.LLM.Mode)
	color.White("  Model: %s", cfg.LLM.Model)
	color.White("  Debug: %t", cfg.LLM.Debug)
	color.White("  Summary Style: %s", cfg.LLM.SummaryStyle)
	color.White("  Max Summary Length: %d", cfg.LLM.MaxSummaryLength)
	color.White("  Include Technical Details: %t", cfg.LLM.IncludeTechnicalDetails)
	color.White("  Prioritize Recent Work: %t", cfg.LLM.PrioritizeRecentWork)
	color.White("  Fallback Strategy: %s", cfg.LLM.FallbackStrategy)

	if cfg.LLM.Mode == "ollama" {
		color.White("  Ollama URL: %s", cfg.LLM.Ollama.BaseURL)
		color.White("  Ollama Model: %s", cfg.LLM.Ollama.Model)
	}

	fmt.Println()

	// Status indicators
	if !cfg.LLM.Enabled {
		color.Yellow("Status: ⚠️  Disabled")
		color.White("LLM features are disabled. Reports will use basic text processing.")
		return nil
	}

	switch cfg.LLM.Mode {
	case "embedded":
		color.Green("Status: ✅ Embedded mode active")
		color.White("Using built-in lightweight summarization.")
	case "ollama":
		color.White("Status: Testing Ollama connection...")
		llmConfig := llm.LLMConfig{
			Enabled:                  cfg.LLM.Enabled,
			Mode:                     cfg.LLM.Mode,
			Model:                    cfg.LLM.Model,
			Debug:                    cfg.LLM.Debug,
			SummaryStyle:             cfg.LLM.SummaryStyle,
			MaxSummaryLength:         cfg.LLM.MaxSummaryLength,
			IncludeTechnicalDetails:  cfg.LLM.IncludeTechnicalDetails,
			PrioritizeRecentWork:     cfg.LLM.PrioritizeRecentWork,
			FallbackStrategy:         cfg.LLM.FallbackStrategy,
			OllamaURL:                cfg.LLM.Ollama.BaseURL,
			OllamaModel:              cfg.LLM.Ollama.Model,
		}
		
		if err := llm.TestLLMConnection(llmConfig); err != nil {
			color.Red("Status: ❌ Ollama connection failed")
			color.White("Error: %v", err)
			color.White("Make sure Ollama is running and the model is available.")
		} else {
			color.Green("Status: ✅ Ollama connected")
		}
	case "disabled":
		color.Yellow("Status: ⚠️  Explicitly disabled")
	default:
		color.Red("Status: ❌ Unknown mode: %s", cfg.LLM.Mode)
	}

	fmt.Println()
	color.White("💡 Tips:")
	if cfg.LLM.Mode == "ollama" {
		color.White("  • Test connection: my-day llm test")
		color.White("  • Check Ollama status: ollama list")
		color.White("  • Pull model: ollama pull %s", cfg.LLM.Ollama.Model)
	}
	color.White("  • Disable LLM: my-day report --no-llm")
	color.White("  • Change mode: edit config file or use --llm-mode flag")

	return nil
}

func createTestIssue() jira.Issue {
	return jira.Issue{
		Key: "TEST-123",
		Fields: jira.Fields{
			Summary:     "Fix authentication timeout in user login API",
			Description: jira.JiraDescription{Text: "Users are experiencing timeouts when logging in through the API. Investigation shows the OAuth token validation is taking too long."},
			Status: jira.Status{
				Name: "In Progress",
				Category: struct {
					ID   string `json:"id"`
					Key  string `json:"key"`
					Name string `json:"name"`
				}{
					Key: "indeterminate",
				},
			},
			Priority: jira.Priority{
				Name: "High",
			},
			IssueType: jira.IssueType{
				Name: "Bug",
			},
			Project: jira.Project{
				Key:  "TEST",
				Name: "Test Project",
			},
		},
	}
}

func startDockerLLM() error {
	color.Cyan("🐳 Starting Docker LLM...")
	
	dockerManager := llm.NewDockerLLMManager()
	return dockerManager.EnsureReady()
}

func stopDockerLLM() error {
	color.Cyan("🛑 Stopping Docker LLM...")
	
	dockerManager := llm.NewDockerLLMManager()
	if err := dockerManager.StopContainer(); err != nil {
		return err
	}
	
	color.Green("✅ Docker LLM stopped successfully")
	return nil
}