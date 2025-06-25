package config

import "github.com/spf13/viper"

// SetDefaults sets default configuration values
func SetDefaults() {
	// Jira defaults
	viper.SetDefault("jira.oauth.redirect_uri", "http://localhost:8080/callback")
	
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
	viper.SetDefault("llm.ollama.base_url", "http://localhost:11434")
	viper.SetDefault("llm.ollama.model", "llama3.1")

	// Report defaults
	viper.SetDefault("report.format", "console")
	viper.SetDefault("report.include_yesterday", true)
	viper.SetDefault("report.include_today", true)
	viper.SetDefault("report.include_in_progress", true)

	// Application defaults
	viper.SetDefault("verbose", false)
	viper.SetDefault("quiet", false)
}