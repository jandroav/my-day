package llm

import (
	"fmt"
	"strings"
	"my-day/internal/jira"
)

// EmbeddedLLM provides lightweight text summarization without external dependencies
type EmbeddedLLM struct {
	model        string
	debugLogger  *DebugLogger
	errorHandler *ErrorHandler
	config       *LLMConfig
}

// NewEmbeddedLLM creates a new embedded LLM instance
func NewEmbeddedLLM(model string) *EmbeddedLLM {
	debugLogger := NewDebugLogger(false, false) // Debug disabled by default
	return &EmbeddedLLM{
		model:        model,
		debugLogger:  debugLogger,
		errorHandler: NewErrorHandler("graceful", debugLogger),
	}
}

// NewEmbeddedLLMWithDebug creates a new embedded LLM instance with debug logging enabled
func NewEmbeddedLLMWithDebug(model string, verbose bool) *EmbeddedLLM {
	debugLogger := NewDebugLogger(true, verbose)
	return &EmbeddedLLM{
		model:        model,
		debugLogger:  debugLogger,
		errorHandler: NewErrorHandler("graceful", debugLogger),
	}
}

// NewEmbeddedLLMWithConfig creates a new embedded LLM instance with full configuration
func NewEmbeddedLLMWithConfig(config LLMConfig) *EmbeddedLLM {
	debugLogger := NewDebugLogger(config.Debug, config.Debug) // Use debug flag for verbose too
	return &EmbeddedLLM{
		model:        config.Model,
		debugLogger:  debugLogger,
		errorHandler: NewErrorHandler(config.FallbackStrategy, debugLogger),
		config:       &config, // Store config for use in processing
	}
}

// GetDebugReport returns the current debug report
func (e *EmbeddedLLM) GetDebugReport() (*DebugReport, error) {
	if e.debugLogger == nil {
		return nil, fmt.Errorf("debug logger not initialized")
	}
	return e.debugLogger.GetDebugReport()
}

// SummarizeIssue generates a summary for a Jira issue
func (e *EmbeddedLLM) SummarizeIssue(issue jira.Issue) (string, error) {
	return e.generateRuleBasedSummary(issue), nil
}

// SummarizeComments generates a summary of user's comments
func (e *EmbeddedLLM) SummarizeComments(comments []jira.Comment) (string, error) {
	if len(comments) == 0 {
		return "", nil
	}
	
	if len(comments) == 1 {
		return e.createIntelligentSummary(comments[0].Body.Text), nil
	}
	
	// For multiple comments, create a combined summary
	var activities []string
	for _, comment := range comments {
		if comment.Body.Text != "" {
			summary := e.createIntelligentSummary(comment.Body.Text)
			if len(summary) > 0 {
				activities = append(activities, summary)
			}
		}
	}
	
	if len(activities) == 0 {
		return "Multiple comments added", nil
	}
	
	// Combine and limit length
	combined := strings.Join(activities, "; ")
	maxLength := e.getConfiguredMaxLength()
	if len(combined) > maxLength {
		return e.shortenText(combined, maxLength), nil
	}
	
	return combined, nil
}

// SummarizeIssues generates summaries for multiple issues
func (e *EmbeddedLLM) SummarizeIssues(issues []jira.Issue) (map[string]string, error) {
	summaries := make(map[string]string)
	
	for _, issue := range issues {
		summary, err := e.SummarizeIssue(issue)
		if err != nil {
			continue
		}
		summaries[issue.Key] = summary
	}
	
	return summaries, nil
}

// SummarizeWorklog generates a summary for worklog entries
func (e *EmbeddedLLM) SummarizeWorklog(worklogs []jira.WorklogEntry) (string, error) {
	if len(worklogs) == 0 {
		return "No work logged", nil
	}
	
	return fmt.Sprintf("Work logged on %d items", len(worklogs)), nil
}

// GenerateStandupSummary creates an overall summary for standup reporting
func (e *EmbeddedLLM) GenerateStandupSummary(issues []jira.Issue, worklogs []jira.WorklogEntry) (string, error) {
	if len(issues) == 0 && len(worklogs) == 0 {
		return "No recent activity to report", nil
	}

	var parts []string
	
	if len(issues) > 0 {
		parts = append(parts, fmt.Sprintf("%d issues", len(issues)))
	}
	
	if len(worklogs) > 0 {
		parts = append(parts, fmt.Sprintf("%d worklog entries", len(worklogs)))
	}
	
	return "Recent activity: " + strings.Join(parts, ", "), nil
}

// GenerateStandupSummaryWithComments creates an enhanced summary using comment data
func (e *EmbeddedLLM) GenerateStandupSummaryWithComments(issues []jira.Issue, comments []jira.Comment, worklogs []jira.WorklogEntry) (string, error) {
	if len(issues) == 0 && len(comments) == 0 && len(worklogs) == 0 {
		return "No recent activity to report", nil
	}
	
	var parts []string
	
	if len(issues) > 0 {
		parts = append(parts, fmt.Sprintf("%d issues", len(issues)))
	}
	
	if len(comments) > 0 {
		parts = append(parts, fmt.Sprintf("%d comments", len(comments)))
	}
	
	if len(worklogs) > 0 {
		parts = append(parts, fmt.Sprintf("%d worklog entries", len(worklogs)))
	}
	
	return "Recent activity: " + strings.Join(parts, ", "), nil
}

// generateRuleBasedSummary creates a concise summary using rule-based approach
func (e *EmbeddedLLM) generateRuleBasedSummary(issue jira.Issue) string {
	// Start with the status and priority context
	context := e.getContextualPrefix(issue)
	
	// Extract key information from summary and description
	keyPoints := e.extractKeyPoints(issue)
	
	// Combine context with key points
	if len(keyPoints) > 0 {
		return fmt.Sprintf("%s %s", context, strings.Join(keyPoints, ", "))
	}
	
	return context + " " + e.shortenText(issue.Fields.Summary, e.getConfiguredMaxLength()/3)
}

// getContextualPrefix generates a contextual prefix based on issue status and type
func (e *EmbeddedLLM) getContextualPrefix(issue jira.Issue) string {
	status := strings.ToLower(issue.Fields.Status.Name)
	issueType := strings.ToLower(issue.Fields.IssueType.Name)
	priority := strings.ToLower(issue.Fields.Priority.Name)
	
	var prefix string
	
	// Status-based prefixes
	switch {
	case strings.Contains(status, "progress") || strings.Contains(status, "development"):
		prefix = "Working on"
	case strings.Contains(status, "review"):
		prefix = "Under review:"
	case strings.Contains(status, "done") || strings.Contains(status, "closed"):
		prefix = "Completed"
	case strings.Contains(status, "blocked"):
		prefix = "Blocked:"
	default:
		prefix = "Planning"
	}
	
	// Add urgency for high priority items
	if strings.Contains(priority, "high") || strings.Contains(priority, "critical") {
		prefix = "🔥 " + prefix
	}
	
	// Add type context for specific issue types
	if strings.Contains(issueType, "bug") {
		prefix += " bug fix:"
	} else if strings.Contains(issueType, "feature") || strings.Contains(issueType, "story") {
		prefix += " feature:"
	} else if strings.Contains(issueType, "task") {
		prefix += " task:"
	}
	
	return prefix
}

// extractKeyPoints identifies important keywords and phrases
func (e *EmbeddedLLM) extractKeyPoints(issue jira.Issue) []string {
	text := issue.Fields.Summary + " " + issue.Fields.Description.Text
	text = strings.ToLower(text)
	
	var points []string
	
	// Technical keywords that indicate specific work (only include if technical details are enabled)
	if e.shouldIncludeTechnicalDetails() {
		techKeywords := []string{
			"api", "database", "migration", "deployment", "ci/cd", "pipeline",
			"security", "authentication", "oauth", "ssl", "encryption",
			"performance", "optimization", "scaling", "monitoring",
			"docker", "kubernetes", "k8s", "terraform", "ansible",
			"aws", "azure", "gcp", "cloud", "serverless",
			"microservice", "integration", "endpoint", "service",
		}
		
		// Find technical context
		for _, keyword := range techKeywords {
			if strings.Contains(text, keyword) {
				points = append(points, keyword)
				if len(points) >= 2 { // Limit technical keywords
					break
				}
			}
		}
	}
	
	// Action keywords that indicate what's being done (always include)
	actionKeywords := []string{
		"implement", "fix", "update", "upgrade", "refactor",
		"optimize", "improve", "enhance", "add", "remove",
		"configure", "setup", "install", "deploy", "migrate",
		"investigate", "debug", "troubleshoot", "analyze",
		"review", "test", "validate", "verify",
	}
	
	// Find action context
	for _, keyword := range actionKeywords {
		if strings.Contains(text, keyword) {
			points = append(points, keyword)
			break // One action keyword is enough
		}
	}
	
	return points
}

// Configuration-aware helper methods
func (e *EmbeddedLLM) getConfiguredMaxLength() int {
	if e.config != nil && e.config.MaxSummaryLength > 0 {
		return e.config.MaxSummaryLength
	}
	return 200 // Default max length
}

func (e *EmbeddedLLM) shouldIncludeTechnicalDetails() bool {
	if e.config != nil {
		return e.config.IncludeTechnicalDetails
	}
	return true // Default to including technical details
}

func (e *EmbeddedLLM) getSummaryStyle() string {
	if e.config != nil && e.config.SummaryStyle != "" {
		return e.config.SummaryStyle
	}
	return "technical" // Default style
}

func (e *EmbeddedLLM) shouldPrioritizeRecentWork() bool {
	if e.config != nil {
		return e.config.PrioritizeRecentWork
	}
	return true // Default to prioritizing recent work
}

// Helper methods
func (e *EmbeddedLLM) shortenText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	
	shortened := text[:maxLength]
	lastSpace := strings.LastIndex(shortened, " ")
	
	if lastSpace > maxLength/2 {
		return text[:lastSpace] + "..."
	}
	
	return text[:maxLength-3] + "..."
}

func (e *EmbeddedLLM) createIntelligentSummary(text string) string {
	// Enhanced intelligent summary creation with actual AI-like processing
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	
	// Extract key actions and technical details
	summary := e.extractKeyInformation(text)
	
	// Apply configured max length
	maxLength := e.getConfiguredMaxLength()
	if len(summary) > maxLength && maxLength > 0 {
		return e.shortenText(summary, maxLength)
	}
	
	return summary
}

// extractKeyInformation performs intelligent extraction of key information from comment text
func (e *EmbeddedLLM) extractKeyInformation(text string) string {
	// Convert to lowercase for analysis
	lowerText := strings.ToLower(text)
	
	// Define action patterns and their summarized forms
	actionPatterns := map[string]string{
		"implemented":   "✅ Implemented",
		"created":       "🔧 Created", 
		"fixed":         "🐛 Fixed",
		"updated":       "📝 Updated",
		"resolved":      "✅ Resolved",
		"deployed":      "🚀 Deployed",
		"configured":    "⚙️ Configured",
		"integrated":    "🔗 Integrated",
		"refactored":    "♻️ Refactored",
		"optimized":     "⚡ Optimized",
		"tested":        "🧪 Tested",
		"merged":        "🔀 Merged",
		"reviewed":      "👀 Reviewed",
		"investigating": "🔍 Investigating",
		"working on":    "🔄 Working on",
	}
	
	// Find the primary action
	var primaryAction string
	for pattern, action := range actionPatterns {
		if strings.Contains(lowerText, pattern) {
			primaryAction = action
			break
		}
	}
	
	// Extract technical terms and components
	technicalTerms := e.extractTechnicalTerms(text)
	
	// Extract URLs and references
	urls := e.extractURLs(text)
	
	// Build intelligent summary
	var parts []string
	
	if primaryAction != "" {
		parts = append(parts, primaryAction)
	}
	
	// Add key technical context
	if len(technicalTerms) > 0 {
		if len(technicalTerms) <= 3 {
			parts = append(parts, strings.Join(technicalTerms, ", "))
		} else {
			parts = append(parts, fmt.Sprintf("%s and %d other components", 
				strings.Join(technicalTerms[:2], ", "), len(technicalTerms)-2))
		}
	}
	
	// Add reference if URL present
	if len(urls) > 0 {
		parts = append(parts, "with references")
	}
	
	// Always include the original text for full context
	if len(parts) > 0 {
		return strings.Join(parts, " ") + " - " + text
	}
	
	// If no patterns matched, return the full original text
	return text
}

// extractTechnicalTerms identifies technical components, tools, and systems
func (e *EmbeddedLLM) extractTechnicalTerms(text string) []string {
	terms := []string{}
	
	// Technical patterns to look for
	patterns := []string{
		"GitHub", "GitLab", "Jira", "AWS", "Docker", "Kubernetes", "Terraform",
		"Jenkins", "CI/CD", "API", "REST", "GraphQL", "database", "PostgreSQL", 
		"MySQL", "MongoDB", "Redis", "Elasticsearch", "microservice", "lambda",
		"workflow", "pipeline", "deployment", "infrastructure", "vault", "secret",
		"OAuth", "authentication", "authorization", "SSL", "TLS", "certificate",
		"Liquibase", "Snowflake", "MSSQL", "localstack", "Ollama", "LLM", "gradle",
		"maven", "npm", "yarn", "golang", "python", "java", "javascript", "typescript",
	}
	
	lowerText := strings.ToLower(text)
	for _, pattern := range patterns {
		if strings.Contains(lowerText, strings.ToLower(pattern)) {
			terms = append(terms, pattern)
		}
	}
	
	return terms
}

// extractURLs finds URLs in the text
func (e *EmbeddedLLM) extractURLs(text string) []string {
	urls := []string{}
	words := strings.Fields(text)
	
	for _, word := range words {
		if strings.HasPrefix(word, "http://") || strings.HasPrefix(word, "https://") {
			urls = append(urls, word)
		}
	}
	
	return urls
}

