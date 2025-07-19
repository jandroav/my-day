package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize my-day configuration",
	Long: `Initialize creates a default configuration file for my-day.
	
This will create a config.yaml file in ~/.my-day/ with default settings
that you can customize for your Jira instance and team projects.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := initializeConfig(cmd); err != nil {
			color.Red("Error initializing configuration: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	
	// Init-specific flags
	initCmd.Flags().Bool("force", false, "Overwrite existing configuration file")
	initCmd.Flags().Bool("guided", false, "Interactive guided setup")
}

func initializeConfig(cmd *cobra.Command) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".my-day")
	configFile := filepath.Join(configDir, "config.yaml")

	// Check if config already exists
	if _, err := os.Stat(configFile); err == nil {
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			color.Yellow("Configuration file already exists at: %s", configFile)
			color.Yellow("Use --force to overwrite")
			return nil
		}
	}

	// Create config directory
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Check if guided setup is requested
	guided, _ := cmd.Flags().GetBool("guided")
	
	var configContent string
	if guided {
		configContent = generateGuidedConfig()
	} else {
		configContent = generateConfigTemplate()
	}

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	color.Green("✓ Configuration file created at: %s", configFile)
	
	if guided {
		showGuidedSetupInstructions(configFile)
	} else {
		showStandardSetupInstructions(configFile)
	}

	return nil
}

// generateConfigTemplate creates a comprehensive configuration template with comments
func generateConfigTemplate() string {
	return `# my-day Configuration File
# This file contains all settings for the my-day CLI tool
# Only the Jira base_url needs to be changed for basic usage

# =============================================================================
# JIRA CONFIGURATION
# =============================================================================
jira:
  # REQUIRED: Change this to your Jira Cloud instance URL
  base_url: "https://your-instance.atlassian.net"  # env: MY_DAY_JIRA_BASE_URL
  
  # OPTIONAL: Set email/token here or use environment variables/CLI flags
  # For security, consider using environment variables instead:
  # export MY_DAY_JIRA_EMAIL="your-email@example.com"
  # export MY_DAY_JIRA_TOKEN="your-api-token"
  email: ""    # Your Jira email address (env: MY_DAY_JIRA_EMAIL)
  token: ""    # Your Jira API token (env: MY_DAY_JIRA_TOKEN)
  
  # Projects to track (customize for your organization)
  projects:    # env: MY_DAY_JIRA_PROJECTS (comma-separated)
    - "DAT"
    - "IO"
    # Add your own projects:
    # - "PROJ"
  
  # Custom Fields Configuration (for report grouping)
  # Find field IDs in Jira: Admin > Issues > Custom Fields
  custom_fields:
    squad:
      field_id: "customfield_12944"
      display_name: "Squad"
      field_type: "select"
    team:
      field_id: "customfield_12945"
      display_name: "Team"
      field_type: "select"
    component:
      field_id: "customfield_12946"
      display_name: "Component"
      field_type: "multi-select"
    epic:
      field_id: "customfield_10014"
      display_name: "Epic Link"
      field_type: "epic"
    sprint:
      field_id: "customfield_10007"
      display_name: "Sprint"
      field_type: "sprint"

# =============================================================================
# LLM (AI) CONFIGURATION
# =============================================================================
llm:
  enabled: true                                      # env: MY_DAY_LLM_ENABLED
  mode: "ollama"                                     # env: MY_DAY_LLM_MODE (ollama, embedded, disabled)
  model: "qwen2.5:3b"                                # env: MY_DAY_LLM_MODEL
  
  # LLM Behavior Settings
  debug: false                                       # env: MY_DAY_LLM_DEBUG
  summary_style: "technical"                         # env: MY_DAY_LLM_SUMMARY_STYLE (technical, business, brief)
  max_summary_length: 0                             # env: MY_DAY_LLM_MAX_SUMMARY_LENGTH (0 = no limit)
  include_technical_details: true                    # env: MY_DAY_LLM_INCLUDE_TECHNICAL_DETAILS
  prioritize_recent_work: true                       # env: MY_DAY_LLM_PRIORITIZE_RECENT_WORK
  fallback_strategy: "graceful"                      # env: MY_DAY_LLM_FALLBACK_STRATEGY (graceful, strict)
  
  # Ollama Configuration (Docker-based LLM)
  ollama:
    base_url: "http://localhost:11434"               # env: MY_DAY_LLM_OLLAMA_BASE_URL
    model: "qwen2.5:3b"                              # env: MY_DAY_LLM_OLLAMA_MODEL
    
  # Model Recommendations:
  # - qwen2.5:3b (1.9GB) - Fast, good balance (default)
  # - llama3.1:8b (4.7GB) - Better quality, slower
  # - codellama:7b (3.8GB) - Best for technical/DevOps content
  # - phi3:3.8b (2.3GB) - Good for Microsoft tech stacks

# =============================================================================
# REPORT CONFIGURATION
# =============================================================================
report:
  format: "console"                                  # env: MY_DAY_REPORT_FORMAT (console, markdown)
  include_yesterday: true                            # env: MY_DAY_REPORT_INCLUDE_YESTERDAY
  include_today: true                                # env: MY_DAY_REPORT_INCLUDE_TODAY
  include_in_progress: true                          # env: MY_DAY_REPORT_INCLUDE_IN_PROGRESS
  
  # Obsidian Export Settings
  export:
    enabled: false                                   # env: MY_DAY_REPORT_EXPORT_ENABLED
    folder_path: "~/Documents/my-day-reports"        # env: MY_DAY_REPORT_EXPORT_FOLDER_PATH
    filename_date: "2006-01-02"                      # env: MY_DAY_REPORT_EXPORT_FILENAME_DATE
    tags: ["report", "my-day", "standup"]            # env: MY_DAY_REPORT_EXPORT_TAGS (comma-separated)

# =============================================================================
# ADVANCED SETTINGS
# =============================================================================
# Most users won't need to modify these settings

# Global settings
verbose: false                                       # env: MY_DAY_VERBOSE
quiet: false                                         # env: MY_DAY_QUIET

# =============================================================================
# USAGE EXAMPLES
# =============================================================================
# After setup, try these commands:
#
# Basic daily report:
#   my-day report
#
# Report with AI analysis:
#   my-day report --debug --show-quality
#
# Group by squad/team:
#   my-day report --field squad
#
# Export to Obsidian:
#   my-day report --export
#
# Switch LLM models:
#   my-day llm models
#   my-day llm switch codellama:7b
#
# Test different models:
#   my-day report --ollama-model llama3.1:8b
#   my-day report --llm-style business
#
# =============================================================================
# For more information:
# - Documentation: https://github.com/jandroav/my-day
# - Report issues: https://github.com/jandroav/my-day/issues
# - Check config: my-day config show
# - Test LLM: my-day llm test
# =============================================================================
`
}

// generateGuidedConfig creates a simpler config with prompts for essential settings
func generateGuidedConfig() string {
	return `# my-day Configuration File
# Generated with guided setup - edit the values marked with TODO

# =============================================================================
# JIRA CONFIGURATION (REQUIRED)
# =============================================================================
jira:
  # TODO: Change this to your Jira Cloud URL
  base_url: "https://your-instance.atlassian.net"   # env: MY_DAY_JIRA_BASE_URL
  
  # Authentication (recommended: use environment variables for security)
  # Set these via environment or use 'my-day auth' command
  email: ""    # Your Jira email (env: MY_DAY_JIRA_EMAIL)
  token: ""    # Your Jira API token (env: MY_DAY_JIRA_TOKEN)
  
  # TODO: Update these project keys to match your organization
  projects:    # env: MY_DAY_JIRA_PROJECTS (comma-separated)
    - "DAT"
    - "IO"

# =============================================================================
# LLM (AI) CONFIGURATION
# =============================================================================
llm:
  enabled: true                                      # env: MY_DAY_LLM_ENABLED
  mode: "ollama"                                     # env: MY_DAY_LLM_MODE (ollama, embedded, disabled)
  model: "qwen2.5:3b"                                # env: MY_DAY_LLM_MODEL
  
  # AI Behavior
  summary_style: "technical"                         # env: MY_DAY_LLM_SUMMARY_STYLE (technical, business, brief)
  include_technical_details: true                    # env: MY_DAY_LLM_INCLUDE_TECHNICAL_DETAILS
  
  # Docker LLM Settings
  ollama:
    base_url: "http://localhost:11434"               # env: MY_DAY_LLM_OLLAMA_BASE_URL
    model: "qwen2.5:3b"                              # env: MY_DAY_LLM_OLLAMA_MODEL

# =============================================================================
# REPORT CONFIGURATION
# =============================================================================
report:
  format: "console"                                  # env: MY_DAY_REPORT_FORMAT (console, markdown)
  include_yesterday: true                            # env: MY_DAY_REPORT_INCLUDE_YESTERDAY
  include_today: true                                # env: MY_DAY_REPORT_INCLUDE_TODAY
  include_in_progress: true                          # env: MY_DAY_REPORT_INCLUDE_IN_PROGRESS
  
  # Obsidian Export (optional)
  export:
    enabled: false                                   # env: MY_DAY_REPORT_EXPORT_ENABLED
    folder_path: "~/Documents/my-day-reports"        # env: MY_DAY_REPORT_EXPORT_FOLDER_PATH
    tags: ["standup", "my-day"]                      # env: MY_DAY_REPORT_EXPORT_TAGS (comma-separated)

# =============================================================================
# Ready to use! Try these commands after setting your Jira URL:
#
# 1. Authenticate: my-day auth --email you@company.com --token YOUR_TOKEN
# 2. Sync tickets: my-day sync
# 3. Daily report: my-day report
# 4. AI status:    my-day llm status
# =============================================================================
`
}

// showStandardSetupInstructions displays detailed setup instructions for standard init
func showStandardSetupInstructions(configFile string) {
	color.Cyan("\n🚀 Setup Instructions:")
	color.White("1. Edit your Jira URL:")
	color.Yellow("   vi %s", configFile)
	color.White("   Change: https://your-instance.atlassian.net")
	color.White("   To:     https://yourcompany.atlassian.net")
	
	fmt.Println()
	color.White("2. Get your Jira API token:")
	color.Yellow("   Visit: https://id.atlassian.com/manage-profile/security/api-tokens")
	color.White("   Click 'Create API token', copy the token")
	
	fmt.Println()
	color.White("3. Authenticate with Jira:")
	color.Yellow("   my-day auth --email your-email@example.com --token YOUR_API_TOKEN")
	
	fmt.Println()
	color.White("4. Sync your tickets:")
	color.Yellow("   my-day sync")
	
	fmt.Println()
	color.White("5. Generate your first report:")
	color.Yellow("   my-day report")
	
	fmt.Println()
	color.Cyan("💡 Pro Tips:")
	color.White("• The config includes Docker LLM for better AI summaries")
	color.White("• Edit projects section to match your Jira projects")
	color.White("• Use 'my-day config show' to verify your settings")
	color.White("• Check 'my-day llm status' to test AI functionality")
}

// showGuidedSetupInstructions displays simplified instructions for guided setup
func showGuidedSetupInstructions(configFile string) {
	color.Cyan("\n🎯 Quick Setup (Guided Mode):")
	color.White("Your config is ready! Just need 2 things:")
	
	fmt.Println()
	color.Yellow("1. Update Jira URL:")
	color.White("   Edit: %s", configFile)
	color.White("   Find: https://your-instance.atlassian.net")
	color.White("   Change to: https://YOURCOMPANY.atlassian.net")
	
	fmt.Println()
	color.Yellow("2. Authenticate:")
	color.White("   Get token: https://id.atlassian.com/manage-profile/security/api-tokens")
	color.White("   Run: my-day auth --email you@company.com --token YOUR_TOKEN")
	
	fmt.Println()
	color.Green("✨ Then you're ready:")
	color.Yellow("   my-day sync    # Get your tickets")
	color.Yellow("   my-day report  # Generate report")
	
	fmt.Println()
	color.Cyan("💡 Your config includes:")
	color.White("• Docker-based LLM for smart summaries")
	color.White("• Common DevOps project templates")
	color.White("• Ready-to-use report settings")
}