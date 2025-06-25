package config

import (
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Jira   JiraConfig   `mapstructure:"jira" yaml:"jira"`
	LLM    LLMConfig    `mapstructure:"llm" yaml:"llm"`
	Report ReportConfig `mapstructure:"report" yaml:"report"`
}

// JiraConfig represents Jira configuration
type JiraConfig struct {
	BaseURL  string        `mapstructure:"base_url" yaml:"base_url"`
	OAuth    OAuthConfig   `mapstructure:"oauth" yaml:"oauth"`
	Projects []ProjectInfo `mapstructure:"projects" yaml:"projects"`
}

// OAuthConfig represents OAuth configuration
type OAuthConfig struct {
	ClientID     string `mapstructure:"client_id" yaml:"client_id"`
	ClientSecret string `mapstructure:"client_secret" yaml:"client_secret"`
	RedirectURI  string `mapstructure:"redirect_uri" yaml:"redirect_uri"`
}

// ProjectInfo represents a Jira project
type ProjectInfo struct {
	Key  string `mapstructure:"key" yaml:"key"`
	Name string `mapstructure:"name" yaml:"name"`
}

// LLMConfig represents LLM configuration
type LLMConfig struct {
	Enabled bool         `mapstructure:"enabled" yaml:"enabled"`
	Mode    string       `mapstructure:"mode" yaml:"mode"`
	Model   string       `mapstructure:"model" yaml:"model"`
	Ollama  OllamaConfig `mapstructure:"ollama" yaml:"ollama"`
}

// OllamaConfig represents Ollama-specific configuration
type OllamaConfig struct {
	BaseURL string `mapstructure:"base_url" yaml:"base_url"`
	Model   string `mapstructure:"model" yaml:"model"`
}

// ReportConfig represents report generation configuration
type ReportConfig struct {
	Format            string `mapstructure:"format" yaml:"format"`
	IncludeYesterday  bool   `mapstructure:"include_yesterday" yaml:"include_yesterday"`
	IncludeToday      bool   `mapstructure:"include_today" yaml:"include_today"`
	IncludeInProgress bool   `mapstructure:"include_in_progress" yaml:"include_in_progress"`
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