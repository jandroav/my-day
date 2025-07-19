package config

import "github.com/spf13/viper"

// SetDefaults sets default configuration values
func SetDefaults() {
	// Jira defaults (API token authentication)
	viper.SetDefault("jira.email", "")
	viper.SetDefault("jira.token", "")
	
	// Default projects for DevOps teams
	viper.SetDefault("jira.projects", []map[string]string{
		{"key": "DEVOPS", "name": "DevOps Team"},
		{"key": "INTEROP", "name": "Interop Team"},
		{"key": "FOUND", "name": "Foundation Team"},
		{"key": "ENT", "name": "Enterprise Team"},
		{"key": "LBIO", "name": "LBIO Team"},
	})

	// LLM defaults
	viper.SetDefault("llm.enabled", true)
	viper.SetDefault("llm.mode", "embedded")
	viper.SetDefault("llm.model", "tinyllama")
	viper.SetDefault("llm.debug", false)
	viper.SetDefault("llm.summary_style", "technical")
	viper.SetDefault("llm.max_summary_length", 200)
	viper.SetDefault("llm.include_technical_details", true)
	viper.SetDefault("llm.prioritize_recent_work", true)
	viper.SetDefault("llm.fallback_strategy", "graceful")
	viper.SetDefault("llm.ollama.base_url", "http://localhost:11434")
	viper.SetDefault("llm.ollama.model", "llama3.1")

	// Report defaults
	viper.SetDefault("report.format", "console")
	viper.SetDefault("report.include_yesterday", true)
	viper.SetDefault("report.include_today", true)
	viper.SetDefault("report.include_in_progress", true)
	
	// Export defaults
	viper.SetDefault("report.export.enabled", false)
	viper.SetDefault("report.export.folder_path", "~/Documents/my-day-reports")
	viper.SetDefault("report.export.filename_date", "2006-01-02")
	viper.SetDefault("report.export.tags", []string{"report", "my-day"})

	// Application defaults
	viper.SetDefault("verbose", false)
	viper.SetDefault("quiet", false)
}