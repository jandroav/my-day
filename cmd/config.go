package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"encoding/json"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"my-day/internal/config"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long: `Manage my-day configuration settings.

You can view current configuration, edit the config file,
or set individual configuration values.`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  "Display the current configuration values from all sources (file, flags, env vars).",
	Run: func(cmd *cobra.Command, args []string) {
		if err := showConfiguration(cmd); err != nil {
			color.Red("Error showing configuration: %v", err)
			os.Exit(1)
		}
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit configuration file",
	Long:  "Open the configuration file in your default editor.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := editConfiguration(); err != nil {
			color.Red("Error editing configuration: %v", err)
			os.Exit(1)
		}
	},
}

var configPathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	Long:  "Display the path to the configuration file being used.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := showConfigPath(); err != nil {
			color.Red("Error showing config path: %v", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configEditCmd)
	configCmd.AddCommand(configPathCmd)

	// Config show flags
	configShowCmd.Flags().Bool("json", false, "Output configuration as JSON")
	configShowCmd.Flags().Bool("sources", false, "Show configuration sources")
}

func showConfiguration(cmd *cobra.Command) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if JSON output is requested
	if json, _ := cmd.Flags().GetBool("json"); json {
		return showConfigurationJSON(cfg)
	}

	// Check if sources should be shown
	if sources, _ := cmd.Flags().GetBool("sources"); sources {
		return showConfigurationSources()
	}

	// Show human-readable configuration
	color.Cyan("📋 Current Configuration")
	fmt.Println()

	// Jira section
	color.Yellow("Jira:")
	color.White("  Base URL: %s", cfg.Jira.BaseURL)
	color.White("  OAuth Client ID: %s", maskSensitive(cfg.Jira.OAuth.ClientID))
	color.White("  OAuth Client Secret: %s", maskSensitive(cfg.Jira.OAuth.ClientSecret))
	color.White("  Redirect URI: %s", cfg.Jira.OAuth.RedirectURI)
	color.White("  Projects:")
	for _, project := range cfg.Jira.Projects {
		color.White("    - %s (%s)", project.Key, project.Name)
	}
	fmt.Println()

	// LLM section
	color.Yellow("LLM:")
	color.White("  Enabled: %t", cfg.LLM.Enabled)
	color.White("  Mode: %s", cfg.LLM.Mode)
	color.White("  Model: %s", cfg.LLM.Model)
	if cfg.LLM.Mode == "ollama" {
		color.White("  Ollama URL: %s", cfg.LLM.Ollama.BaseURL)
		color.White("  Ollama Model: %s", cfg.LLM.Ollama.Model)
	}
	fmt.Println()

	// Report section
	color.Yellow("Report:")
	color.White("  Format: %s", cfg.Report.Format)
	color.White("  Include Yesterday: %t", cfg.Report.IncludeYesterday)
	color.White("  Include Today: %t", cfg.Report.IncludeToday)
	color.White("  Include In Progress: %t", cfg.Report.IncludeInProgress)

	return nil
}

func showConfigurationJSON(cfg *config.Config) error {
	// Create a copy with masked sensitive values
	safeCfg := *cfg
	safeCfg.Jira.OAuth.ClientSecret = maskSensitive(cfg.Jira.OAuth.ClientSecret)

	data, err := json.MarshalIndent(safeCfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	fmt.Println(string(data))
	return nil
}

func showConfigurationSources() error {
	color.Cyan("📋 Configuration Sources")
	fmt.Println()

	configFile := viper.ConfigFileUsed()
	if configFile != "" {
		color.Yellow("Config File: %s", configFile)
	} else {
		color.Yellow("Config File: Not found")
	}

	color.Yellow("Environment Variables: MY_DAY_*")
	color.Yellow("Command Line Flags: Available on all commands")

	fmt.Println()
	color.White("Priority Order (highest to lowest):")
	color.White("1. Command line flags")
	color.White("2. Environment variables")
	color.White("3. Configuration file")
	color.White("4. Default values")

	return nil
}

func editConfiguration() error {
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		// Try to find or create config file
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configFile = filepath.Join(homeDir, ".my-day", "config.yaml")
		
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			color.Yellow("Configuration file does not exist. Run 'my-day init' first.")
			return nil
		}
	}

	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi" // Default editor
	}

	color.Cyan("Opening configuration file in %s...", editor)
	color.White("File: %s", configFile)

	cmd := exec.Command(editor, configFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open editor: %w", err)
	}

	color.Green("✓ Configuration file saved")
	return nil
}

func showConfigPath() error {
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configFile = filepath.Join(homeDir, ".my-day", "config.yaml")
		
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			color.Yellow("Configuration file does not exist")
			color.White("Expected location: %s", configFile)
			color.White("Run 'my-day init' to create it")
			return nil
		}
	}

	color.Green("Configuration file: %s", configFile)
	return nil
}

func maskSensitive(value string) string {
	if value == "" {
		return "(not set)"
	}
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + "****" + value[len(value)-2:]
}