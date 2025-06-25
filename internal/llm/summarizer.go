package llm

import (
	"fmt"
	"my-day/internal/jira"
)

// Summarizer defines the interface for LLM-based summarization
type Summarizer interface {
	SummarizeIssue(issue jira.Issue) (string, error)
	SummarizeIssues(issues []jira.Issue) (map[string]string, error)
	SummarizeWorklog(worklogs []jira.WorklogEntry) (string, error)
	GenerateStandupSummary(issues []jira.Issue, worklogs []jira.WorklogEntry) (string, error)
}

// ConnectionTester defines interface for testing LLM connectivity
type ConnectionTester interface {
	TestConnection() error
}

// LLMConfig represents LLM configuration options
type LLMConfig struct {
	Enabled     bool
	Mode        string // "embedded", "ollama", "disabled"
	Model       string
	OllamaURL   string
	OllamaModel string
}

// NewSummarizer creates a new summarizer based on configuration
func NewSummarizer(config LLMConfig) (Summarizer, error) {
	if !config.Enabled {
		return NewDisabledSummarizer(), nil
	}
	
	switch config.Mode {
	case "embedded":
		return NewEmbeddedLLM(config.Model), nil
	case "ollama":
		return NewOllamaClient(config.OllamaURL, config.OllamaModel), nil
	case "disabled":
		return NewDisabledSummarizer(), nil
	default:
		return nil, fmt.Errorf("unknown LLM mode: %s", config.Mode)
	}
}

// DisabledSummarizer provides fallback when LLM is disabled
type DisabledSummarizer struct{}

// NewDisabledSummarizer creates a new disabled summarizer
func NewDisabledSummarizer() *DisabledSummarizer {
	return &DisabledSummarizer{}
}

// SummarizeIssue returns basic issue information without LLM processing
func (d *DisabledSummarizer) SummarizeIssue(issue jira.Issue) (string, error) {
	return fmt.Sprintf("%s - %s", issue.Fields.Status.Name, issue.Fields.Summary), nil
}

// SummarizeIssues returns basic summaries for multiple issues
func (d *DisabledSummarizer) SummarizeIssues(issues []jira.Issue) (map[string]string, error) {
	summaries := make(map[string]string)
	
	for _, issue := range issues {
		summaries[issue.Key] = fmt.Sprintf("%s - %s", issue.Fields.Status.Name, issue.Fields.Summary)
	}
	
	return summaries, nil
}

// SummarizeWorklog returns basic worklog information
func (d *DisabledSummarizer) SummarizeWorklog(worklogs []jira.WorklogEntry) (string, error) {
	if len(worklogs) == 0 {
		return "No work logged", nil
	}
	
	return fmt.Sprintf("Work logged on %d items", len(worklogs)), nil
}

// GenerateStandupSummary returns basic activity summary
func (d *DisabledSummarizer) GenerateStandupSummary(issues []jira.Issue, worklogs []jira.WorklogEntry) (string, error) {
	return fmt.Sprintf("Recent activity: %d issues, %d worklog entries", len(issues), len(worklogs)), nil
}

// TestLLMConnection tests if the configured LLM service is available
func TestLLMConnection(config LLMConfig) error {
	if !config.Enabled || config.Mode == "disabled" {
		return nil // No connection needed
	}
	
	switch config.Mode {
	case "embedded":
		// Embedded LLM doesn't need external connection
		return nil
	case "ollama":
		client := NewOllamaClient(config.OllamaURL, config.OllamaModel)
		return client.TestConnection()
	default:
		return fmt.Errorf("unknown LLM mode: %s", config.Mode)
	}
}