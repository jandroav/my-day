package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"my-day/internal/config"
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

	// Create default configuration
	defaultConfig := config.Config{
		Jira: config.JiraConfig{
			BaseURL: "https://your-instance.atlassian.net",
			Projects: []config.ProjectInfo{
				{Key: "DEVOPS", Name: "DevOps Team"},
				{Key: "INTEROP", Name: "Interop Team"},
				{Key: "FOUND", Name: "Foundation Team"},
				{Key: "ENT", Name: "Enterprise Team"},
				{Key: "LBIO", Name: "LBIO Team"},
			},
		},
		LLM: config.LLMConfig{
			Enabled: true,
			Mode:    "embedded",
			Model:   "tinyllama",
			Ollama: config.OllamaConfig{
				BaseURL: "http://localhost:11434",
				Model:   "llama3.1",
			},
		},
		Report: config.ReportConfig{
			Format:            "console",
			IncludeYesterday:  true,
			IncludeToday:      true,
			IncludeInProgress: true,
		},
	}

	// Write configuration to file
	data, err := yaml.Marshal(&defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	color.Green("✓ Configuration file created at: %s", configFile)
	color.Cyan("\nNext steps:")
	color.White("1. Edit the configuration file to set your Jira URL and OAuth credentials")
	color.White("2. Run 'my-day auth' to authenticate with Jira")
	color.White("3. Run 'my-day sync' to pull your tickets")
	color.White("4. Run 'my-day report' to generate your daily report")

	return nil
}