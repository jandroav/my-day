package config

import (
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Jira   JiraConfig   `mapstructure:"jira" yaml:"jira"`
	GitHub GitHubConfig `mapstructure:"github" yaml:"github"`
	LLM    LLMConfig    `mapstructure:"llm" yaml:"llm"`
	Report ReportConfig `mapstructure:"report" yaml:"report"`
}

// JiraConfig represents Jira configuration
type JiraConfig struct {
	BaseURL      string                 `mapstructure:"base_url" yaml:"base_url"`
	Email        string                 `mapstructure:"email" yaml:"email"`
	Token        string                 `mapstructure:"token" yaml:"token"`
	Projects     []string               `mapstructure:"projects" yaml:"projects"`
	CustomFields map[string]CustomField `mapstructure:"custom_fields" yaml:"custom_fields"`
}

// CustomField represents a custom field configuration
type CustomField struct {
	FieldID     string `mapstructure:"field_id" yaml:"field_id"`
	DisplayName string `mapstructure:"display_name" yaml:"display_name"`
	FieldType   string `mapstructure:"field_type" yaml:"field_type"`
}

// GitHubConfig represents GitHub configuration
type GitHubConfig struct {
	Enabled      bool     `mapstructure:"enabled" yaml:"enabled"`
	Token        string   `mapstructure:"token" yaml:"token"`
	Repositories []string `mapstructure:"repositories" yaml:"repositories"`
	IncludePRs   bool     `mapstructure:"include_prs" yaml:"include_prs"`
	IncludeCommits bool   `mapstructure:"include_commits" yaml:"include_commits"`
	IncludeWorkflows bool `mapstructure:"include_workflows" yaml:"include_workflows"`
}

// LLMConfig represents LLM configuration
type LLMConfig struct {
	Enabled                  bool         `mapstructure:"enabled" yaml:"enabled"`
	Mode                     string       `mapstructure:"mode" yaml:"mode"`
	Model                    string       `mapstructure:"model" yaml:"model"`
	Debug                    bool         `mapstructure:"debug" yaml:"debug"`
	SummaryStyle             string       `mapstructure:"summary_style" yaml:"summary_style"`
	MaxSummaryLength         int          `mapstructure:"max_summary_length" yaml:"max_summary_length"`
	IncludeTechnicalDetails  bool         `mapstructure:"include_technical_details" yaml:"include_technical_details"`
	PrioritizeRecentWork     bool         `mapstructure:"prioritize_recent_work" yaml:"prioritize_recent_work"`
	FallbackStrategy         string       `mapstructure:"fallback_strategy" yaml:"fallback_strategy"`
	Ollama                   OllamaConfig `mapstructure:"ollama" yaml:"ollama"`
}

// OllamaConfig represents Ollama-specific configuration
type OllamaConfig struct {
	BaseURL string `mapstructure:"base_url" yaml:"base_url"`
	Model   string `mapstructure:"model" yaml:"model"`
}

// ReportConfig represents report generation configuration
type ReportConfig struct {
	Format            string       `mapstructure:"format" yaml:"format"`
	IncludeYesterday  bool         `mapstructure:"include_yesterday" yaml:"include_yesterday"`
	IncludeToday      bool         `mapstructure:"include_today" yaml:"include_today"`
	IncludeInProgress bool         `mapstructure:"include_in_progress" yaml:"include_in_progress"`
	Export            ExportConfig `mapstructure:"export" yaml:"export"`
}

// ExportConfig represents export configuration
type ExportConfig struct {
	Enabled       bool   `mapstructure:"enabled" yaml:"enabled"`
	FolderPath    string `mapstructure:"folder_path" yaml:"folder_path"`
	FileNameDate  string `mapstructure:"filename_date" yaml:"filename_date"`
	Tags          []string `mapstructure:"tags" yaml:"tags"`
}

// Load loads the configuration from viper
func Load() (*Config, error) {
	var config Config
	
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetString returns a string configuration value
func GetString(key string) string {
	return viper.GetString(key)
}

// GetBool returns a boolean configuration value
func GetBool(key string) bool {
	return viper.GetBool(key)
}

// GetStringSlice returns a string slice configuration value
func GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}