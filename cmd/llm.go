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

var llmModelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List available LLM models",
	Long:  "List available LLM models for the current LLM mode.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := listAvailableModels(); err != nil {
			color.Red("Failed to list models: %v", err)
			os.Exit(1)
		}
	},
}

var llmSwitchCmd = &cobra.Command{
	Use:   "switch [model-name]",
	Short: "Switch LLM model",
	Long:  "Switch to a different LLM model. Use 'my-day llm models' to see available models.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		modelName := args[0]
		if err := switchLLMModel(modelName); err != nil {
			color.Red("Failed to switch model: %v", err)
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
	llmCmd.AddCommand(llmModelsCmd)
	llmCmd.AddCommand(llmSwitchCmd)
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

func listAvailableModels() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	color.Cyan("🧠 Available LLM Models")
	fmt.Println()

	color.Yellow("Current Configuration:")
	color.White("  Mode: %s", cfg.LLM.Mode)
	color.White("  Current Model: %s", cfg.LLM.Model)
	if cfg.LLM.Mode == "ollama" {
		color.White("  Ollama Model: %s", cfg.LLM.Ollama.Model)
	}
	fmt.Println()

	switch cfg.LLM.Mode {
	case "ollama":
		color.Yellow("📦 Recommended Ollama Models:")
		fmt.Println()
		
		models := []struct {
			Name        string
			Size        string
			Performance string
			Description string
		}{
			{"qwen2.5:3b", "1.9GB", "Fast", "Current default - good balance of speed and quality"},
			{"llama3.2:3b", "2.0GB", "Fast", "Meta's efficient model for quick responses"},
			{"phi3:3.8b", "2.3GB", "Medium", "Microsoft's compact model, good for technical content"},
			{"llama3.1:8b", "4.7GB", "Medium", "Larger model with better understanding"},
			{"qwen2.5:7b", "4.1GB", "Medium", "Enhanced reasoning capabilities"},
			{"llama3.1:70b", "40GB", "Slow", "Highest quality but requires significant resources"},
			{"codellama:7b", "3.8GB", "Medium", "Specialized for code and technical content"},
			{"mistral:7b", "4.1GB", "Medium", "Good general-purpose model"},
		}

		for _, model := range models {
			if model.Name == cfg.LLM.Ollama.Model {
				color.Green("✅ %s (%s) - %s - %s", model.Name, model.Size, model.Performance, model.Description)
			} else {
				color.White("   %s (%s) - %s - %s", model.Name, model.Size, model.Performance, model.Description)
			}
		}

		fmt.Println()
		color.Yellow("💡 Usage:")
		color.White("  • Switch model: my-day llm switch qwen2.5:7b")
		color.White("  • Pull new model: ollama pull mistral:7b")
		color.White("  • List installed: ollama list")
		
		fmt.Println()
		color.Yellow("🔍 Checking installed models...")
		if err := showInstalledOllamaModels(); err != nil {
			color.Yellow("⚠️  Could not check installed models: %v", err)
			color.White("   Make sure Ollama is running: ollama serve")
		}

	case "embedded":
		color.Yellow("🔧 Embedded Mode Models:")
		fmt.Println()
		
		embeddedModels := []struct {
			Name        string
			Description string
		}{
			{"enhanced-embedded", "Enhanced pattern matching with technical term recognition"},
			{"basic-embedded", "Simple keyword extraction and basic summarization"},
		}

		for _, model := range embeddedModels {
			if model.Name == cfg.LLM.Model {
				color.Green("✅ %s - %s", model.Name, model.Description)
			} else {
				color.White("   %s - %s", model.Name, model.Description)
			}
		}

		fmt.Println()
		color.Yellow("💡 Usage:")
		color.White("  • Switch to Ollama for more models: my-day llm switch --mode ollama")
		color.White("  • Change embedded model: my-day report --llm-model enhanced-embedded")

	case "disabled":
		color.Yellow("⚠️  LLM is disabled")
		color.White("Enable LLM to use AI-powered summarization:")
		color.White("  • Enable embedded: my-day report --llm-enabled --llm-mode embedded")
		color.White("  • Enable Ollama: my-day report --llm-enabled --llm-mode ollama")

	default:
		color.Red("❌ Unknown LLM mode: %s", cfg.LLM.Mode)
	}

	return nil
}

func showInstalledOllamaModels() error {
	// Try to connect to Ollama and list installed models
	llmConfig := llm.LLMConfig{
		Mode:        "ollama",
		OllamaURL:   "http://localhost:11434",
		OllamaModel: "qwen2.5:3b", // dummy for connection test
	}
	
	if err := llm.TestLLMConnection(llmConfig); err != nil {
		return fmt.Errorf("Ollama not available: %w", err)
	}

	// If we can connect, show a simple message
	// In a real implementation, you'd call Ollama's API to list models
	color.Green("✅ Ollama is running")
	color.White("   Use 'ollama list' to see installed models")
	color.White("   Use 'ollama pull <model>' to install new models")
	
	return nil
}

func switchLLMModel(modelName string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	color.Cyan("🔄 Switching LLM model to: %s", modelName)

	// Validate the model name based on current mode
	switch cfg.LLM.Mode {
	case "ollama":
		if err := validateOllamaModel(modelName); err != nil {
			return fmt.Errorf("invalid Ollama model: %w", err)
		}
		color.White("✓ Model validated for Ollama")
		
	case "embedded":
		validEmbeddedModels := []string{"enhanced-embedded", "basic-embedded"}
		if !contains(validEmbeddedModels, modelName) {
			return fmt.Errorf("invalid embedded model. Valid options: %v", validEmbeddedModels)
		}
		color.White("✓ Model validated for embedded mode")
		
	case "disabled":
		return fmt.Errorf("LLM is disabled. Enable it first with --llm-enabled")
		
	default:
		return fmt.Errorf("unknown LLM mode: %s", cfg.LLM.Mode)
	}

	// Show current configuration for reference
	color.Yellow("📋 Current vs New Configuration:")
	color.White("  Mode: %s (unchanged)", cfg.LLM.Mode)
	color.White("  Current Model: %s", cfg.LLM.Model)
	color.White("  New Model: %s", modelName)
	
	fmt.Println()
	color.Yellow("💡 Model Switch Options:")
	color.White("Option 1 - Via CLI flag (temporary):")
	color.White("  my-day report --llm-model %s", modelName)
	if cfg.LLM.Mode == "ollama" {
		color.White("  my-day report --ollama-model %s", modelName)
	}
	
	fmt.Println()
	color.White("Option 2 - Via environment variable:")
	color.White("  export MY_DAY_LLM_MODEL=%s", modelName)
	if cfg.LLM.Mode == "ollama" {
		color.White("  export MY_DAY_LLM_OLLAMA_MODEL=%s", modelName)
	}
	
	fmt.Println()
	color.White("Option 3 - Update config file:")
	color.White("  Edit ~/.my-day/config.yaml:")
	color.White("  llm:")
	color.White("    model: %s", modelName)
	if cfg.LLM.Mode == "ollama" {
		color.White("    ollama:")
		color.White("      model: %s", modelName)
	}

	fmt.Println()
	color.Green("✅ Model switch information provided!")
	color.White("💡 Test the new model: my-day llm test --llm-model %s", modelName)

	return nil
}

func validateOllamaModel(modelName string) error {
	// Basic validation - check if it looks like an Ollama model name
	if modelName == "" {
		return fmt.Errorf("model name cannot be empty")
	}
	
	// Common Ollama model patterns
	validPatterns := []string{
		"llama", "qwen", "phi", "mistral", "codellama", "gemma", "neural-chat",
	}
	
	for _, pattern := range validPatterns {
		if len(modelName) >= len(pattern) && modelName[:len(pattern)] == pattern {
			return nil
		}
	}
	
	// If it doesn't match common patterns, still allow it but warn
	color.Yellow("⚠️  Warning: '%s' doesn't match common Ollama model patterns", modelName)
	color.White("   Make sure the model is available: ollama pull %s", modelName)
	
	return nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}