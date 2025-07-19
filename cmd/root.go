package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"my-day/internal/config"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "my-day",
	Short: "A DevOps daily standup report generator",
	Long: `my-day is a colorful CLI tool that helps DevOps team members 
track Jira tickets across multiple teams and generate daily standup reports.

It integrates with Jira Cloud and optionally uses embedded LLM for ticket summarization.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.my-day/config.yaml)")
	rootCmd.PersistentFlags().String("jira-url", "", "Jira base URL")
	rootCmd.PersistentFlags().String("jira-email", "", "Jira email address for API token authentication")
	rootCmd.PersistentFlags().String("jira-token", "", "Jira API token")
	rootCmd.PersistentFlags().StringSlice("projects", []string{}, "Jira project keys to track")
	rootCmd.PersistentFlags().String("llm-mode", "ollama", "LLM mode: embedded, ollama, disabled")
	rootCmd.PersistentFlags().String("llm-model", "qwen2.5:3b", "LLM model name")
	rootCmd.PersistentFlags().Bool("llm-enabled", true, "Enable LLM features")
	rootCmd.PersistentFlags().String("ollama-url", "http://localhost:11434", "Ollama base URL")
	rootCmd.PersistentFlags().String("ollama-model", "qwen2.5:3b", "Ollama model name")
	rootCmd.PersistentFlags().Bool("llm-debug", false, "Enable LLM debug mode")
	rootCmd.PersistentFlags().String("llm-style", "technical", "LLM summary style: technical, business, brief")
	rootCmd.PersistentFlags().Int("llm-max-length", 0, "Maximum LLM summary length (0 for no limit)")
	rootCmd.PersistentFlags().Bool("llm-technical-details", true, "Include technical details in summaries")
	rootCmd.PersistentFlags().String("llm-fallback", "graceful", "LLM fallback strategy: graceful, strict")
	rootCmd.PersistentFlags().String("report-format", "console", "Report format: console, markdown")
	rootCmd.PersistentFlags().Bool("include-yesterday", true, "Include yesterday's work in report")
	rootCmd.PersistentFlags().Bool("include-today", true, "Include today's work in report")
	rootCmd.PersistentFlags().Bool("include-in-progress", true, "Include in-progress tickets in report")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Enable quiet output")

	// Bind flags to viper
	viper.BindPFlag("jira.base_url", rootCmd.PersistentFlags().Lookup("jira-url"))
	viper.BindPFlag("jira.email", rootCmd.PersistentFlags().Lookup("jira-email"))
	viper.BindPFlag("jira.token", rootCmd.PersistentFlags().Lookup("jira-token"))
	viper.BindPFlag("jira.projects", rootCmd.PersistentFlags().Lookup("projects"))
	viper.BindPFlag("llm.mode", rootCmd.PersistentFlags().Lookup("llm-mode"))
	viper.BindPFlag("llm.model", rootCmd.PersistentFlags().Lookup("llm-model"))
	viper.BindPFlag("llm.enabled", rootCmd.PersistentFlags().Lookup("llm-enabled"))
	viper.BindPFlag("llm.debug", rootCmd.PersistentFlags().Lookup("llm-debug"))
	viper.BindPFlag("llm.summary_style", rootCmd.PersistentFlags().Lookup("llm-style"))
	viper.BindPFlag("llm.max_summary_length", rootCmd.PersistentFlags().Lookup("llm-max-length"))
	viper.BindPFlag("llm.include_technical_details", rootCmd.PersistentFlags().Lookup("llm-technical-details"))
	viper.BindPFlag("llm.fallback_strategy", rootCmd.PersistentFlags().Lookup("llm-fallback"))
	viper.BindPFlag("llm.ollama.base_url", rootCmd.PersistentFlags().Lookup("ollama-url"))
	viper.BindPFlag("llm.ollama.model", rootCmd.PersistentFlags().Lookup("ollama-model"))
	viper.BindPFlag("report.format", rootCmd.PersistentFlags().Lookup("report-format"))
	viper.BindPFlag("report.include_yesterday", rootCmd.PersistentFlags().Lookup("include-yesterday"))
	viper.BindPFlag("report.include_today", rootCmd.PersistentFlags().Lookup("include-today"))
	viper.BindPFlag("report.include_in_progress", rootCmd.PersistentFlags().Lookup("include-in-progress"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".my-day" (without extension).
		viper.AddConfigPath(home + "/.my-day")
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// Environment variables
	viper.SetEnvPrefix("MY_DAY")
	viper.AutomaticEnv()
	
	// Bind environment variables explicitly for nested keys
	// Jira configuration
	viper.BindEnv("jira.email", "MY_DAY_JIRA_EMAIL")
	viper.BindEnv("jira.token", "MY_DAY_JIRA_TOKEN")
	viper.BindEnv("jira.base_url", "MY_DAY_JIRA_BASE_URL")
	viper.BindEnv("jira.projects", "MY_DAY_JIRA_PROJECTS")
	
	// LLM configuration
	viper.BindEnv("llm.mode", "MY_DAY_LLM_MODE")
	viper.BindEnv("llm.model", "MY_DAY_LLM_MODEL")
	viper.BindEnv("llm.enabled", "MY_DAY_LLM_ENABLED")
	viper.BindEnv("llm.debug", "MY_DAY_LLM_DEBUG")
	viper.BindEnv("llm.summary_style", "MY_DAY_LLM_SUMMARY_STYLE")
	viper.BindEnv("llm.max_summary_length", "MY_DAY_LLM_MAX_SUMMARY_LENGTH")
	viper.BindEnv("llm.include_technical_details", "MY_DAY_LLM_INCLUDE_TECHNICAL_DETAILS")
	viper.BindEnv("llm.prioritize_recent_work", "MY_DAY_LLM_PRIORITIZE_RECENT_WORK")
	viper.BindEnv("llm.fallback_strategy", "MY_DAY_LLM_FALLBACK_STRATEGY")
	viper.BindEnv("llm.ollama.base_url", "MY_DAY_LLM_OLLAMA_BASE_URL")
	viper.BindEnv("llm.ollama.model", "MY_DAY_LLM_OLLAMA_MODEL")
	
	// Report configuration
	viper.BindEnv("report.format", "MY_DAY_REPORT_FORMAT")
	viper.BindEnv("report.include_yesterday", "MY_DAY_REPORT_INCLUDE_YESTERDAY")
	viper.BindEnv("report.include_today", "MY_DAY_REPORT_INCLUDE_TODAY")
	viper.BindEnv("report.include_in_progress", "MY_DAY_REPORT_INCLUDE_IN_PROGRESS")
	viper.BindEnv("report.export.enabled", "MY_DAY_REPORT_EXPORT_ENABLED")
	viper.BindEnv("report.export.folder_path", "MY_DAY_REPORT_EXPORT_FOLDER_PATH")
	viper.BindEnv("report.export.filename_date", "MY_DAY_REPORT_EXPORT_FILENAME_DATE")
	viper.BindEnv("report.export.tags", "MY_DAY_REPORT_EXPORT_TAGS")

	// Set defaults
	config.SetDefaults()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}