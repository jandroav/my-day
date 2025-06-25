package llm

import (
	"fmt"
	"strings"
	"my-day/internal/jira"
)

// EmbeddedLLM provides lightweight text summarization without external dependencies
type EmbeddedLLM struct {
	model string
}

// NewEmbeddedLLM creates a new embedded LLM instance
func NewEmbeddedLLM(model string) *EmbeddedLLM {
	return &EmbeddedLLM{
		model: model,
	}
}

// SummarizeIssue generates a summary for a Jira issue
func (e *EmbeddedLLM) SummarizeIssue(issue jira.Issue) (string, error) {
	// For now, implement rule-based summarization
	// This can be enhanced later with actual LLM integration
	return e.generateRuleBasedSummary(issue), nil
}

// SummarizeIssues generates summaries for multiple issues
func (e *EmbeddedLLM) SummarizeIssues(issues []jira.Issue) (map[string]string, error) {
	summaries := make(map[string]string)
	
	for _, issue := range issues {
		summary, err := e.SummarizeIssue(issue)
		if err != nil {
			continue // Skip issues that can't be summarized
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
	
	// Group by issue and summarize
	issueWork := make(map[string][]string)
	for _, worklog := range worklogs {
		if worklog.Comment != "" {
			issueWork[worklog.IssueID] = append(issueWork[worklog.IssueID], worklog.Comment)
		}
	}
	
	var summaryParts []string
	for issueID, comments := range issueWork {
		issueSummary := e.summarizeWorklogComments(comments)
		summaryParts = append(summaryParts, fmt.Sprintf("%s: %s", issueID, issueSummary))
	}
	
	return strings.Join(summaryParts, "; "), nil
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
	
	return context + " " + e.shortenText(issue.Fields.Summary, 60)
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
	text := issue.Fields.Summary + " " + issue.Fields.Description
	text = strings.ToLower(text)
	
	var points []string
	
	// Technical keywords that indicate specific work
	techKeywords := []string{
		"api", "database", "migration", "deployment", "ci/cd", "pipeline",
		"security", "authentication", "oauth", "ssl", "encryption",
		"performance", "optimization", "scaling", "monitoring",
		"docker", "kubernetes", "k8s", "terraform", "ansible",
		"aws", "azure", "gcp", "cloud", "serverless",
		"microservice", "integration", "endpoint", "service",
		"bug", "error", "exception", "crash", "timeout",
		"test", "testing", "unit test", "integration test",
		"documentation", "readme", "guide", "tutorial",
	}
	
	// Action keywords that indicate what's being done
	actionKeywords := []string{
		"implement", "fix", "update", "upgrade", "refactor",
		"optimize", "improve", "enhance", "add", "remove",
		"configure", "setup", "install", "deploy", "migrate",
		"investigate", "debug", "troubleshoot", "analyze",
		"review", "test", "validate", "verify",
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
	
	// Find action context
	for _, keyword := range actionKeywords {
		if strings.Contains(text, keyword) {
			points = append(points, keyword)
			break // One action keyword is enough
		}
	}
	
	// Extract environment mentions
	environments := []string{"production", "staging", "development", "test", "prod", "dev"}
	for _, env := range environments {
		if strings.Contains(text, env) {
			points = append(points, env+" environment")
			break
		}
	}
	
	return points
}

// summarizeWorklogComments creates a brief summary of worklog comments
func (e *EmbeddedLLM) summarizeWorklogComments(comments []string) string {
	if len(comments) == 0 {
		return "work logged"
	}
	
	if len(comments) == 1 {
		return e.shortenText(comments[0], 50)
	}
	
	// For multiple comments, find common themes
	allText := strings.ToLower(strings.Join(comments, " "))
	
	// Look for common action words
	actions := []string{"implemented", "fixed", "updated", "tested", "deployed", "configured", "investigated"}
	for _, action := range actions {
		if strings.Contains(allText, action) {
			return action + " and updated progress"
		}
	}
	
	// Default to first comment if no patterns found
	return e.shortenText(comments[0], 40) + " (and more)"
}

// shortenText truncates text to a maximum length while preserving word boundaries
func (e *EmbeddedLLM) shortenText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	
	// Find the last space before maxLength
	shortened := text[:maxLength]
	lastSpace := strings.LastIndex(shortened, " ")
	
	if lastSpace > maxLength/2 {
		return text[:lastSpace] + "..."
	}
	
	return text[:maxLength-3] + "..."
}

// GenerateStandupSummary creates an overall summary for standup reporting
func (e *EmbeddedLLM) GenerateStandupSummary(issues []jira.Issue, worklogs []jira.WorklogEntry) (string, error) {
	if len(issues) == 0 && len(worklogs) == 0 {
		return "No recent activity to report", nil
	}
	
	var summaryParts []string
	
	// Categorize issues by status
	inProgress := 0
	completed := 0
	blocked := 0
	
	for _, issue := range issues {
		status := strings.ToLower(issue.Fields.Status.Name)
		switch {
		case strings.Contains(status, "progress") || strings.Contains(status, "development"):
			inProgress++
		case strings.Contains(status, "done") || strings.Contains(status, "closed") || strings.Contains(status, "resolved"):
			completed++
		case strings.Contains(status, "blocked"):
			blocked++
		}
	}
	
	// Generate summary based on activity
	if inProgress > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("actively working on %d items", inProgress))
	}
	
	if completed > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("completed %d items", completed))
	}
	
	if blocked > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d items blocked", blocked))
	}
	
	if len(worklogs) > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("logged work on %d items", len(worklogs)))
	}
	
	if len(summaryParts) == 0 {
		return "Recent activity includes ticket updates and progress", nil
	}
	
	return "Currently " + strings.Join(summaryParts, ", "), nil
}