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
	rootCmd.PersistentFlags().String("jira-client-id", "", "Jira OAuth client ID")
	rootCmd.PersistentFlags().String("jira-client-secret", "", "Jira OAuth client secret")
	rootCmd.PersistentFlags().String("jira-redirect-uri", "http://localhost:8080/callback", "OAuth redirect URI")
	rootCmd.PersistentFlags().StringSlice("projects", []string{}, "Jira project keys to track")
	rootCmd.PersistentFlags().String("llm-mode", "embedded", "LLM mode: embedded, ollama, disabled")
	rootCmd.PersistentFlags().String("llm-model", "tinyllama", "LLM model name")
	rootCmd.PersistentFlags().Bool("llm-enabled", true, "Enable LLM features")
	rootCmd.PersistentFlags().String("ollama-url", "http://localhost:11434", "Ollama base URL")
	rootCmd.PersistentFlags().String("ollama-model", "llama3.1", "Ollama model name")
	rootCmd.PersistentFlags().String("report-format", "console", "Report format: console, markdown")
	rootCmd.PersistentFlags().Bool("include-yesterday", true, "Include yesterday's work in report")
	rootCmd.PersistentFlags().Bool("include-today", true, "Include today's work in report")
	rootCmd.PersistentFlags().Bool("include-in-progress", true, "Include in-progress tickets in report")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Enable quiet output")

	// Bind flags to viper
	viper.BindPFlag("jira.base_url", rootCmd.PersistentFlags().Lookup("jira-url"))
	viper.BindPFlag("jira.oauth.client_id", rootCmd.PersistentFlags().Lookup("jira-client-id"))
	viper.BindPFlag("jira.oauth.client_secret", rootCmd.PersistentFlags().Lookup("jira-client-secret"))
	viper.BindPFlag("jira.oauth.redirect_uri", rootCmd.PersistentFlags().Lookup("jira-redirect-uri"))
	viper.BindPFlag("jira.projects", rootCmd.PersistentFlags().Lookup("projects"))
	viper.BindPFlag("llm.mode", rootCmd.PersistentFlags().Lookup("llm-mode"))
	viper.BindPFlag("llm.model", rootCmd.PersistentFlags().Lookup("llm-model"))
	viper.BindPFlag("llm.enabled", rootCmd.PersistentFlags().Lookup("llm-enabled"))
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

	// Set defaults
	config.SetDefaults()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		if viper.GetBool("verbose") {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}