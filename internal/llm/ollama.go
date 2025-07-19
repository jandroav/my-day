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

// OllamaError represents a structured error from Ollama operations
type OllamaError struct {
	Type    string                 `json:"type"`
	Message string                 `json:"message"`
	Cause   error                  `json:"cause,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *OllamaError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
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

// NewOllamaClientWithDockerManagement creates an Ollama client with automatic Docker management
func NewOllamaClientWithDockerManagement(config LLMConfig) (Summarizer, error) {
	dockerManager := NewDockerLLMManager()
	
	// Try to ensure Docker LLM is ready
	if err := dockerManager.EnsureReady(); err != nil {
		// If Docker setup fails, fall back to embedded LLM with a warning
		fmt.Printf("⚠️  Docker LLM setup failed (%v), falling back to embedded model\n", err)
		return NewEmbeddedLLMWithConfig(config), nil
	}
	
	// Use the Docker-managed Ollama instance
	dockerConfig := config
	dockerConfig.OllamaURL = dockerManager.GetBaseURL()
	dockerConfig.OllamaModel = dockerManager.GetModel()
	
	return NewOllamaClientWithConfig(dockerConfig), nil
}

// SummarizeIssue generates a summary for a Jira issue using Ollama with fallback
func (o *OllamaClient) SummarizeIssue(issue jira.Issue) (string, error) {
	prompt := o.buildIssuePrompt(issue)
	result, err := o.generate(prompt)
	
	// If Ollama fails, fallback to embedded LLM
	if err != nil && o.shouldFallbackToEmbedded(err) {
		return o.fallbackToEmbedded().SummarizeIssue(issue)
	}
	
	return result, err
}

// SummarizeComments generates a summary of user's comments from today using Ollama with fallback
func (o *OllamaClient) SummarizeComments(comments []jira.Comment) (string, error) {
	if len(comments) == 0 {
		return "", nil
	}
	
	prompt := o.buildCommentsPrompt(comments)
	result, err := o.generate(prompt)
	
	// If Ollama fails, fallback to embedded LLM
	if err != nil && o.shouldFallbackToEmbedded(err) {
		return o.fallbackToEmbedded().SummarizeComments(comments)
	}
	
	return result, err
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
	result, err := o.generate(prompt)
	
	// If Ollama fails, fallback to embedded LLM
	if err != nil && o.shouldFallbackToEmbedded(err) {
		return o.fallbackToEmbedded().SummarizeWorklog(worklogs)
	}
	
	return result, err
}

// GenerateStandupSummary creates an overall summary for standup reporting
func (o *OllamaClient) GenerateStandupSummary(issues []jira.Issue, worklogs []jira.WorklogEntry) (string, error) {
	prompt := o.buildStandupPrompt(issues, worklogs)
	result, err := o.generate(prompt)
	
	// If Ollama fails, fallback to embedded LLM
	if err != nil && o.shouldFallbackToEmbedded(err) {
		return o.fallbackToEmbedded().GenerateStandupSummary(issues, worklogs)
	}
	
	return result, err
}

// GenerateStandupSummaryWithComments creates an enhanced summary using comment data
func (o *OllamaClient) GenerateStandupSummaryWithComments(issues []jira.Issue, comments []jira.Comment, worklogs []jira.WorklogEntry) (string, error) {
	prompt := o.buildStandupPromptWithComments(issues, comments, worklogs)
	result, err := o.generate(prompt)
	
	// If Ollama fails, fallback to embedded LLM
	if err != nil && o.shouldFallbackToEmbedded(err) {
		return o.fallbackToEmbedded().GenerateStandupSummaryWithComments(issues, comments, worklogs)
	}
	
	return result, err
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

// generate sends a prompt to Ollama and returns the response with retry logic
func (o *OllamaClient) generate(prompt string) (string, error) {
	return o.generateWithRetry(prompt, 3) // Default 3 retries
}

// generateWithRetry sends a prompt to Ollama with retry logic and enhanced error handling
func (o *OllamaClient) generateWithRetry(prompt string, maxRetries int) (string, error) {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: wait 1s, 2s, 4s between retries
			waitTime := time.Duration(1<<(attempt-1)) * time.Second
			time.Sleep(waitTime)
		}
		
		result, err := o.attemptGenerate(prompt)
		if err == nil {
			return result, nil
		}
		
		lastErr = err
		
		// Check if error is retryable
		if !o.isRetryableError(err) {
			break
		}
		
		// Log retry attempt if debug is enabled
		if o.config != nil && o.config.Debug {
			fmt.Printf("Ollama request failed (attempt %d/%d): %v\n", attempt+1, maxRetries+1, err)
		}
	}
	
	// All retries failed, return enhanced error message
	return "", o.enhanceErrorMessage(lastErr, maxRetries)
}

// attemptGenerate makes a single attempt to generate a response from Ollama
func (o *OllamaClient) attemptGenerate(prompt string) (string, error) {
	// Use longer timeout if debug is enabled
	timeout := 30 * time.Second
	if o.config != nil && o.config.Debug {
		timeout = 60 * time.Second
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	request := OllamaRequest{
		Model:  o.model,
		Prompt: prompt,
		Stream: false,
	}
	
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", &OllamaError{
			Type:    "marshal_error",
			Message: "Failed to prepare request data",
			Cause:   err,
		}
	}
	
	req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/api/generate", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", &OllamaError{
			Type:    "request_creation_error",
			Message: "Failed to create HTTP request",
			Cause:   err,
		}
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := o.client.Do(req)
	if err != nil {
		// Check if it's a timeout or connection error
		if ctx.Err() == context.DeadlineExceeded {
			return "", &OllamaError{
				Type:    "timeout_error",
				Message: fmt.Sprintf("Request timed out after %v", timeout),
				Cause:   err,
			}
		}
		return "", &OllamaError{
			Type:    "connection_error",
			Message: "Failed to connect to Ollama service",
			Cause:   err,
		}
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		// Read response body for more detailed error information
		bodyBytes, _ := json.Marshal(resp.Body)
		return "", &OllamaError{
			Type:    "api_error",
			Message: fmt.Sprintf("Ollama API returned status %d", resp.StatusCode),
			Details: map[string]interface{}{
				"status_code": resp.StatusCode,
				"response_body": string(bodyBytes),
			},
		}
	}
	
	var response OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", &OllamaError{
			Type:    "decode_error",
			Message: "Failed to decode Ollama response",
			Cause:   err,
		}
	}
	
	return strings.TrimSpace(response.Response), nil
}

// isRetryableError determines if an error should trigger a retry
func (o *OllamaClient) isRetryableError(err error) bool {
	if ollamaErr, ok := err.(*OllamaError); ok {
		switch ollamaErr.Type {
		case "timeout_error", "connection_error":
			return true
		case "api_error":
			// Retry on 5xx server errors, but not 4xx client errors
			if details, ok := ollamaErr.Details["status_code"].(int); ok {
				return details >= 500 && details < 600
			}
			return false
		default:
			return false
		}
	}
	return false
}

// enhanceErrorMessage creates a user-friendly error message with suggestions
func (o *OllamaClient) enhanceErrorMessage(err error, retries int) error {
	if ollamaErr, ok := err.(*OllamaError); ok {
		switch ollamaErr.Type {
		case "connection_error":
			return fmt.Errorf("unable to connect to Ollama service at %s after %d retries. Please ensure Ollama is running and accessible. You can start Ollama with 'ollama serve' or check if it's running on a different port", o.baseURL, retries+1)
		case "timeout_error":
			return fmt.Errorf("Ollama request timed out after %d retries. The model '%s' might be too large or the system is under heavy load. Consider using a smaller model or increasing timeout", retries+1, o.model)
		case "api_error":
			if details, ok := ollamaErr.Details["status_code"].(int); ok {
				switch details {
				case 404:
					return fmt.Errorf("model '%s' not found in Ollama. Please pull the model first with 'ollama pull %s'", o.model, o.model)
				case 500:
					return fmt.Errorf("Ollama server error after %d retries. The service might be overloaded or experiencing issues", retries+1)
				default:
					return fmt.Errorf("Ollama API error (status %d) after %d retries: %s", details, retries+1, ollamaErr.Message)
				}
			}
			return fmt.Errorf("Ollama API error after %d retries: %s", retries+1, ollamaErr.Message)
		default:
			return fmt.Errorf("Ollama error after %d retries: %s", retries+1, ollamaErr.Message)
		}
	}
	return fmt.Errorf("Ollama error after %d retries: %v", retries+1, err)
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
	
	prompt += "\n\nIMPORTANT: Write the summary in first person (using 'I' statements) as if you are the person working on this ticket.\n"
	prompt += "Provide a 1-2 sentence summary suitable for a standup report:"
	
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
	
	prompt += "\nIMPORTANT: Write the summary in first person (using 'I' statements) as if you are the person who did this work.\n"
	prompt += "Provide a brief summary of the work accomplished:"
	
	return prompt
}

// buildStandupPrompt creates a prompt for generating an overall standup summary
func (o *OllamaClient) buildStandupPrompt(issues []jira.Issue, worklogs []jira.WorklogEntry) string {
	// Use enhanced prompt generation with configuration-aware templates
	return o.buildEnhancedStandupPrompt(issues, nil, worklogs)
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
	
	prompt += "\nIMPORTANT: Write the summary in first person (using 'I' statements) as if you are the person who did the work.\n"
	prompt += "Provide a 1-2 sentence summary of the work progress described in these comments:"
	
	return prompt
}

// buildStandupPromptWithComments creates a comprehensive prompt for standup summary with comments
func (o *OllamaClient) buildStandupPromptWithComments(issues []jira.Issue, comments []jira.Comment, worklogs []jira.WorklogEntry) string {
	// Use enhanced prompt generation with configuration-aware templates
	return o.buildEnhancedStandupPrompt(issues, comments, worklogs)
}

// buildEnhancedStandupPrompt creates an enhanced standup prompt with configuration-aware templates
func (o *OllamaClient) buildEnhancedStandupPrompt(issues []jira.Issue, comments []jira.Comment, worklogs []jira.WorklogEntry) string {
	// Get summary style from configuration
	summaryStyle := o.getSummaryStyle()
	maxLength := o.getMaxSummaryLength()
	includeTechnicalDetails := o.shouldIncludeTechnicalDetails()
	
	// Build context-rich prompt based on style
	var prompt string
	
	switch summaryStyle {
	case "business":
		prompt = o.buildBusinessStylePrompt(issues, comments, worklogs, maxLength)
	case "brief":
		prompt = o.buildBriefStylePrompt(issues, comments, worklogs, maxLength)
	default: // "technical" or fallback
		prompt = o.buildTechnicalStylePrompt(issues, comments, worklogs, maxLength, includeTechnicalDetails)
	}
	
	return prompt
}

// buildTechnicalStylePrompt creates a technical-focused prompt for DevOps teams
func (o *OllamaClient) buildTechnicalStylePrompt(issues []jira.Issue, comments []jira.Comment, worklogs []jira.WorklogEntry, maxLength int, includeTechnicalDetails bool) string {
	prompt := "You are summarizing work for a DevOps team standup. Focus on technical implementation details, infrastructure work, and deployment status.\n\n"
	
	// Add technical context guidance
	if includeTechnicalDetails {
		prompt += "Pay special attention to these technical areas:\n"
		prompt += "- Infrastructure: Terraform, AWS services, Kubernetes, Docker\n"
		prompt += "- Deployments: CI/CD pipelines, releases, environment changes\n"
		prompt += "- Database: Migrations, permissions, configuration changes\n"
		prompt += "- Security: Authentication, authorization, SSL/TLS, secrets management\n"
		prompt += "- Monitoring: Logging, metrics, alerts, observability\n\n"
	}
	
	// Add structured data
	prompt += o.buildStructuredDataSection(issues, comments, worklogs, true)
	
	// Add technical-focused instructions
	prompt += fmt.Sprintf("Generate a technical standup summary (max %d words) that includes:\n", maxLength/5) // Rough word estimate
	prompt += "1. Specific technical work completed (mention technologies used)\n"
	prompt += "2. Infrastructure or deployment changes\n"
	prompt += "3. Any technical blockers or dependencies\n"
	prompt += "4. Next technical steps or work ready for deployment\n\n"
	
	if includeTechnicalDetails {
		prompt += "Use technical terminology appropriately and mention specific tools, services, or technologies involved.\n\n"
	}
	
	prompt += "IMPORTANT: Write the summary in first person (using 'I' statements) as if you are the person who did the work. This should sound natural when read aloud in a standup meeting.\n\n"
	prompt += "Technical Summary:"
	
	return prompt
}

// buildBusinessStylePrompt creates a business-focused prompt for management reporting
func (o *OllamaClient) buildBusinessStylePrompt(issues []jira.Issue, comments []jira.Comment, worklogs []jira.WorklogEntry, maxLength int) string {
	prompt := "You are summarizing work progress for a business stakeholder standup. Focus on deliverables, progress toward goals, and business impact.\n\n"
	
	// Add business context guidance
	prompt += "Focus on these business aspects:\n"
	prompt += "- Feature delivery and user-facing improvements\n"
	prompt += "- Progress toward project milestones\n"
	prompt += "- Risk mitigation and issue resolution\n"
	prompt += "- Timeline and delivery commitments\n\n"
	
	// Add structured data
	prompt += o.buildStructuredDataSection(issues, comments, worklogs, false)
	
	// Add business-focused instructions
	prompt += fmt.Sprintf("Generate a business-focused standup summary (max %d words) that includes:\n", maxLength/5)
	prompt += "1. Key deliverables completed or progressed\n"
	prompt += "2. Impact on project timeline or business goals\n"
	prompt += "3. Any risks or blockers affecting delivery\n"
	prompt += "4. Next steps toward project milestones\n\n"
	
	prompt += "Avoid technical jargon and focus on business value and outcomes.\n\n"
	prompt += "IMPORTANT: Write the summary in first person (using 'I' statements) as if you are the person who did the work. This should sound natural when read aloud in a standup meeting.\n\n"
	prompt += "Business Summary:"
	
	return prompt
}

// buildBriefStylePrompt creates a concise prompt for quick updates
func (o *OllamaClient) buildBriefStylePrompt(issues []jira.Issue, comments []jira.Comment, worklogs []jira.WorklogEntry, maxLength int) string {
	prompt := "Create a very brief, concise standup summary. Focus only on the most important activities and current status.\n\n"
	
	// Add structured data (simplified)
	prompt += o.buildStructuredDataSection(issues, comments, worklogs, false)
	
	// Add brief-focused instructions
	prompt += fmt.Sprintf("Generate a brief standup summary (max %d words) with:\n", maxLength/5)
	prompt += "1. Most important work completed\n"
	prompt += "2. Current focus/priority\n"
	prompt += "3. Any immediate blockers\n\n"
	
	prompt += "Keep it concise and focus on high-impact activities only.\n\n"
	prompt += "IMPORTANT: Write the summary in first person (using 'I' statements) as if you are the person who did the work. This should sound natural when read aloud in a standup meeting.\n\n"
	prompt += "Brief Summary:"
	
	return prompt
}

// buildStructuredDataSection creates a structured data section for prompts
func (o *OllamaClient) buildStructuredDataSection(issues []jira.Issue, comments []jira.Comment, worklogs []jira.WorklogEntry, includeTechnicalContext bool) string {
	var section strings.Builder
	
	section.WriteString("=== WORK DATA ===\n")
	
	// Add issues with enhanced context
	if len(issues) > 0 {
		section.WriteString("Recent Issues:\n")
		for i, issue := range issues {
			if i >= 5 { // Limit to most important issues
				break
			}
			
			// Add priority and type context
			priorityEmoji := o.getPriorityEmoji(issue.Fields.Priority.Name)
			typeContext := o.getIssueTypeContext(issue.Fields.IssueType.Name)
			
			section.WriteString(fmt.Sprintf("- %s %s [%s] %s: %s\n", 
				priorityEmoji,
				issue.Key,
				issue.Fields.Project.Key,
				typeContext,
				issue.Fields.Summary))
			
			section.WriteString(fmt.Sprintf("  Status: %s", issue.Fields.Status.Name))
			
			// Add technical context if enabled
			if includeTechnicalContext {
				techTerms := o.extractTechnicalTerms(issue.Fields.Summary + " " + issue.Fields.Description.Text)
				if len(techTerms) > 0 {
					section.WriteString(fmt.Sprintf(" | Tech: %s", strings.Join(techTerms, ", ")))
				}
			}
			section.WriteString("\n")
		}
		section.WriteString("\n")
	}
	
	// Add comments with enhanced analysis
	if len(comments) > 0 {
		section.WriteString("Today's Activity Comments:\n")
		for i, comment := range comments {
			if i >= 8 { // Show more comments since they're the main data source
				break
			}
			
			timeStr := comment.Created.Time.Format("15:04")
			activityType := o.determineActivityType(comment.Body.Text)
			
			section.WriteString(fmt.Sprintf("- [%s] %s: %s\n", 
				timeStr, 
				activityType,
				comment.Body.Text))
		}
		section.WriteString("\n")
	}
	
	// Add worklog summary
	if len(worklogs) > 0 {
		section.WriteString(fmt.Sprintf("Work Logged: %d entries\n\n", len(worklogs)))
	}
	
	section.WriteString("=== END DATA ===\n\n")
	
	return section.String()
}

// Configuration helper methods
func (o *OllamaClient) getSummaryStyle() string {
	if o.config != nil && o.config.SummaryStyle != "" {
		return o.config.SummaryStyle
	}
	return "technical" // Default style
}

func (o *OllamaClient) getMaxSummaryLength() int {
	if o.config != nil && o.config.MaxSummaryLength > 0 {
		return o.config.MaxSummaryLength
	}
	return 200 // Default max length
}

func (o *OllamaClient) shouldIncludeTechnicalDetails() bool {
	if o.config != nil {
		return o.config.IncludeTechnicalDetails
	}
	return true // Default to including technical details
}

// Helper methods for enhanced context
func (o *OllamaClient) getPriorityEmoji(priority string) string {
	switch strings.ToLower(priority) {
	case "critical", "highest":
		return "🔥"
	case "high":
		return "⚡"
	case "medium":
		return "📋"
	case "low", "lowest":
		return "📝"
	default:
		return "📋"
	}
}

func (o *OllamaClient) getIssueTypeContext(issueType string) string {
	switch strings.ToLower(issueType) {
	case "bug":
		return "🐛 Bug Fix"
	case "feature", "story":
		return "✨ Feature"
	case "task":
		return "📋 Task"
	case "epic":
		return "🎯 Epic"
	case "improvement":
		return "🔧 Improvement"
	default:
		return "📋 Work"
	}
}

func (o *OllamaClient) extractTechnicalTerms(text string) []string {
	lowerText := strings.ToLower(text)
	var terms []string
	
	technicalTerms := []string{
		"terraform", "aws", "kubernetes", "k8s", "docker",
		"database", "sql", "postgresql", "mysql", "mongodb",
		"api", "rest", "graphql", "microservice",
		"ci/cd", "pipeline", "jenkins", "github", "gitlab",
		"vpc", "ecr", "s3", "lambda", "ec2", "rds",
		"oauth", "authentication", "ssl", "tls", "security",
		"monitoring", "logging", "metrics", "alerts",
	}
	
	for _, term := range technicalTerms {
		if strings.Contains(lowerText, term) {
			terms = append(terms, term)
			if len(terms) >= 3 { // Limit to avoid clutter
				break
			}
		}
	}
	
	return terms
}

func (o *OllamaClient) determineActivityType(text string) string {
	lowerText := strings.ToLower(text)
	
	if strings.Contains(lowerText, "completed") || strings.Contains(lowerText, "finished") || strings.Contains(lowerText, "done") {
		return "✅ Completed"
	}
	if strings.Contains(lowerText, "deployed") || strings.Contains(lowerText, "released") {
		return "🚀 Deployed"
	}
	if strings.Contains(lowerText, "blocked") || strings.Contains(lowerText, "stuck") {
		return "🚫 Blocked"
	}
	if strings.Contains(lowerText, "testing") || strings.Contains(lowerText, "test") {
		return "🧪 Testing"
	}
	if strings.Contains(lowerText, "investigating") || strings.Contains(lowerText, "debugging") {
		return "🔍 Investigating"
	}
	if strings.Contains(lowerText, "working on") || strings.Contains(lowerText, "implementing") {
		return "⚙️ Working"
	}
	
	return "📝 Update"
}

// shouldFallbackToEmbedded determines if we should fallback to embedded LLM based on the error
func (o *OllamaClient) shouldFallbackToEmbedded(err error) bool {
	if ollamaErr, ok := err.(*OllamaError); ok {
		switch ollamaErr.Type {
		case "connection_error", "timeout_error":
			return true // Fallback on connection issues
		case "api_error":
			// Fallback on server errors but not client errors
			if details, ok := ollamaErr.Details["status_code"].(int); ok {
				return details >= 500 && details < 600
			}
			return false
		default:
			return false
		}
	}
	return true // Fallback on unknown errors
}

// fallbackToEmbedded creates an embedded LLM instance for fallback
func (o *OllamaClient) fallbackToEmbedded() *EmbeddedLLM {
	if o.config != nil {
		return NewEmbeddedLLMWithConfig(*o.config)
	}
	return NewEmbeddedLLM(o.model)
}