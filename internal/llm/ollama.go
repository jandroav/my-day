package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"my-day/internal/jira"
)

// OllamaClient represents a client for Ollama API
type OllamaClient struct {
	baseURL string
	model   string
	client  *http.Client
	config  *LLMConfig
}

// OllamaRequest represents a request to Ollama API
type OllamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

// OllamaResponse represents a response from Ollama API
type OllamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		model:   model,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// NewOllamaClientWithConfig creates a new Ollama client with full configuration
func NewOllamaClientWithConfig(config LLMConfig) *OllamaClient {
	timeout := 30 * time.Second
	if config.Debug {
		timeout = 60 * time.Second // Longer timeout for debug mode
	}
	
	return &OllamaClient{
		baseURL: strings.TrimSuffix(config.OllamaURL, "/"),
		model:   config.OllamaModel,
		client:  &http.Client{Timeout: timeout},
		config:  &config, // Store config for prompt generation
	}
}

// SummarizeIssue generates a summary for a Jira issue using Ollama
func (o *OllamaClient) SummarizeIssue(issue jira.Issue) (string, error) {
	prompt := o.buildIssuePrompt(issue)
	return o.generate(prompt)
}

// SummarizeComments generates a summary of user's comments from today using Ollama
func (o *OllamaClient) SummarizeComments(comments []jira.Comment) (string, error) {
	if len(comments) == 0 {
		return "", nil
	}
	
	prompt := o.buildCommentsPrompt(comments)
	return o.generate(prompt)
}

// SummarizeIssues generates summaries for multiple issues
func (o *OllamaClient) SummarizeIssues(issues []jira.Issue) (map[string]string, error) {
	summaries := make(map[string]string)
	
	for _, issue := range issues {
		summary, err := o.SummarizeIssue(issue)
		if err != nil {
			// Use fallback for failed requests
			summaries[issue.Key] = fmt.Sprintf("Status: %s - %s", issue.Fields.Status.Name, issue.Fields.Summary)
			continue
		}
		summaries[issue.Key] = summary
	}
	
	return summaries, nil
}

// SummarizeWorklog generates a summary for worklog entries
func (o *OllamaClient) SummarizeWorklog(worklogs []jira.WorklogEntry) (string, error) {
	if len(worklogs) == 0 {
		return "No work logged", nil
	}
	
	prompt := o.buildWorklogPrompt(worklogs)
	return o.generate(prompt)
}

// GenerateStandupSummary creates an overall summary for standup reporting
func (o *OllamaClient) GenerateStandupSummary(issues []jira.Issue, worklogs []jira.WorklogEntry) (string, error) {
	prompt := o.buildStandupPrompt(issues, worklogs)
	return o.generate(prompt)
}

// GenerateStandupSummaryWithComments creates an enhanced summary using comment data
func (o *OllamaClient) GenerateStandupSummaryWithComments(issues []jira.Issue, comments []jira.Comment, worklogs []jira.WorklogEntry) (string, error) {
	prompt := o.buildStandupPromptWithComments(issues, comments, worklogs)
	return o.generate(prompt)
}

// TestConnection tests if Ollama is available
func (o *OllamaClient) TestConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	req, err := http.NewRequestWithContext(ctx, "GET", o.baseURL+"/api/tags", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama returned status %d", resp.StatusCode)
	}
	
	return nil
}

// generate sends a prompt to Ollama and returns the response
func (o *OllamaClient) generate(prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	request := OllamaRequest{
		Model:  o.model,
		Prompt: prompt,
		Stream: false,
	}
	
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := o.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama API returned status %d", resp.StatusCode)
	}
	
	var response OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}
	
	return strings.TrimSpace(response.Response), nil
}

// buildIssuePrompt creates a prompt for summarizing a single issue
func (o *OllamaClient) buildIssuePrompt(issue jira.Issue) string {
	prompt := fmt.Sprintf(`Summarize this Jira ticket for a daily standup report. Be concise and focus on what work is being done:

Ticket: %s
Project: %s
Status: %s
Priority: %s
Type: %s
Summary: %s`, 
		issue.Key,
		issue.Fields.Project.Key,
		issue.Fields.Status.Name,
		issue.Fields.Priority.Name,
		issue.Fields.IssueType.Name,
		issue.Fields.Summary)
	
	if issue.Fields.Description.Text != "" && len(issue.Fields.Description.Text) < 500 {
		prompt += fmt.Sprintf("\nDescription: %s", issue.Fields.Description.Text)
	}
	
	prompt += "\n\nProvide a 1-2 sentence summary suitable for a standup report:"
	
	return prompt
}

// buildWorklogPrompt creates a prompt for summarizing worklog entries
func (o *OllamaClient) buildWorklogPrompt(worklogs []jira.WorklogEntry) string {
	prompt := "Summarize the following work log entries for a daily standup report:\n\n"
	
	for i, worklog := range worklogs {
		if i >= 10 { // Limit to avoid too long prompts
			break
		}
		prompt += fmt.Sprintf("- %s (%s): %s\n", 
			worklog.IssueID, 
			worklog.Started.Format("Jan 2"), 
			worklog.Comment)
	}
	
	prompt += "\nProvide a brief summary of the work accomplished:"
	
	return prompt
}

// buildStandupPrompt creates a prompt for generating an overall standup summary
func (o *OllamaClient) buildStandupPrompt(issues []jira.Issue, worklogs []jira.WorklogEntry) string {
	prompt := "Create a brief standup summary based on this Jira activity:\n\n"
	
	if len(issues) > 0 {
		prompt += "Recent Issues:\n"
		for i, issue := range issues {
			if i >= 5 { // Limit to most important issues
				break
			}
			prompt += fmt.Sprintf("- %s [%s]: %s (Status: %s)\n", 
				issue.Key, 
				issue.Fields.Project.Key,
				issue.Fields.Summary,
				issue.Fields.Status.Name)
		}
		prompt += "\n"
	}
	
	if len(worklogs) > 0 {
		prompt += fmt.Sprintf("Work logged on %d items\n\n", len(worklogs))
	}
	
	prompt += "Provide a 2-3 sentence summary for daily standup covering what was worked on and current status:"
	
	return prompt
}

// buildCommentsPrompt creates a prompt for summarizing user's comments
func (o *OllamaClient) buildCommentsPrompt(comments []jira.Comment) string {
	prompt := "Summarize the following comments made today for a daily standup report. Focus on what work was accomplished:\n\n"
	
	for i, comment := range comments {
		if i >= 5 { // Limit to avoid too long prompts
			break
		}
		
		timeStr := comment.Created.Time.Format("15:04")
		prompt += fmt.Sprintf("Comment at %s: %s\n", timeStr, comment.Body.Text)
	}
	
	prompt += "\nProvide a 1-2 sentence summary of the work progress described in these comments:"
	
	return prompt
}

// buildStandupPromptWithComments creates a comprehensive prompt for standup summary with comments
func (o *OllamaClient) buildStandupPromptWithComments(issues []jira.Issue, comments []jira.Comment, worklogs []jira.WorklogEntry) string {
	prompt := "Create a comprehensive standup summary based on this Jira activity with detailed comment analysis:\n\n"
	
	if len(issues) > 0 {
		prompt += "Recent Issues:\n"
		for i, issue := range issues {
			if i >= 5 { // Limit to most important issues
				break
			}
			prompt += fmt.Sprintf("- %s [%s]: %s (Status: %s)\n", 
				issue.Key, 
				issue.Fields.Project.Key,
				issue.Fields.Summary,
				issue.Fields.Status.Name)
		}
		prompt += "\n"
	}
	
	if len(comments) > 0 {
		prompt += "Today's Comments (showing actual work done):\n"
		for i, comment := range comments {
			if i >= 8 { // Show more comments since they're the main data source
				break
			}
			timeStr := comment.Created.Time.Format("15:04")
			prompt += fmt.Sprintf("- %s: %s\n", timeStr, comment.Body.Text)
		}
		prompt += "\n"
	}
	
	if len(worklogs) > 0 {
		prompt += fmt.Sprintf("Work logged on %d items\n\n", len(worklogs))
	}
	
	prompt += "Provide a 2-3 sentence summary for daily standup that focuses on:\n"
	prompt += "1. Key technical work accomplished (infrastructure, deployments, fixes)\n"
	prompt += "2. Current status and any blockers\n"
	prompt += "3. Next steps or work ready for deployment\n\n"
	prompt += "Summary:"
	
	return prompt
}