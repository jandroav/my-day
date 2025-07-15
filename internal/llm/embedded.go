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

// SummarizeComments generates a summary of user's comments from today
func (e *EmbeddedLLM) SummarizeComments(comments []jira.Comment) (string, error) {
	if len(comments) == 0 {
		return "", nil
	}
	
	if len(comments) == 1 {
		return e.summarizeSingleComment(comments[0]), nil
	}
	
	return e.summarizeMultipleComments(comments), nil
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
	text := issue.Fields.Summary + " " + issue.Fields.Description.Text
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

	// This will be called with comment content in the report generator
	// For now, just provide a simple summary
	summary := fmt.Sprintf("Worked on %d ticket(s)", len(issues))
	
	if len(worklogs) > 0 {
		summary += fmt.Sprintf(" with %d worklog entries", len(worklogs))
	}
	
	return summary, nil
}

// summarizeSingleComment creates a concise summary of a single comment
func (e *EmbeddedLLM) summarizeSingleComment(comment jira.Comment) string {
	text := comment.Body.Text
	if text == "" {
		return "Added a comment"
	}
	
	// Create intelligent summary based on comment content
	return e.createIntelligentSummary(text)
}

// summarizeMultipleComments creates a summary of multiple comments
func (e *EmbeddedLLM) summarizeMultipleComments(comments []jira.Comment) string {
	if len(comments) == 0 {
		return "No comments"
	}
	
	if len(comments) == 1 {
		return e.createIntelligentSummary(comments[0].Body.Text)
	}
	
	// Combine all comment text for analysis
	var allText []string
	for _, comment := range comments {
		if comment.Body.Text != "" {
			allText = append(allText, comment.Body.Text)
		}
	}
	
	combinedText := strings.Join(allText, " ")
	return e.createIntelligentSummary(combinedText)
}

// extractCommentSummary extracts key information from a comment
func (e *EmbeddedLLM) extractCommentSummary(text string) string {
	lowerText := strings.ToLower(text)
	
	// Look for progress indicators
	progressIndicators := []struct {
		patterns []string
		summary  string
	}{
		{[]string{"completed", "finished", "done", "resolved"}, "Completed work"},
		{[]string{"implemented", "added", "created"}, "Implemented feature"},
		{[]string{"fixed", "resolved", "corrected"}, "Fixed issue"},
		{[]string{"tested", "verified", "validated"}, "Tested changes"},
		{[]string{"deployed", "released", "pushed"}, "Deployed update"},
		{[]string{"investigating", "looking into", "checking"}, "Investigating issue"},
		{[]string{"working on", "in progress", "currently"}, "Working on task"},
		{[]string{"updated", "modified", "changed"}, "Updated configuration"},
		{[]string{"blocked", "waiting", "need"}, "Blocked or waiting"},
		{[]string{"reviewed", "code review", "pr"}, "Code review activity"},
	}
	
	for _, indicator := range progressIndicators {
		for _, pattern := range indicator.patterns {
			if strings.Contains(lowerText, pattern) {
				return indicator.summary
			}
		}
	}
	
	// Look for technical keywords
	techKeywords := []string{
		"terraform", "aws", "database", "api", "docker", "kubernetes",
		"pipeline", "ci/cd", "deployment", "configuration", "security",
		"permissions", "testing", "monitoring", "logging",
	}
	
	for _, keyword := range techKeywords {
		if strings.Contains(lowerText, keyword) {
			return "Technical progress on " + keyword
		}
	}
	
	return ""
}

// extractCommentAction identifies the main action in a comment
func (e *EmbeddedLLM) extractCommentAction(text string) string {
	lowerText := strings.ToLower(text)
	
	actions := []struct {
		patterns []string
		action   string
	}{
		{[]string{"completed", "finished", "done"}, "completed tasks"},
		{[]string{"implemented", "added"}, "implemented features"},
		{[]string{"fixed", "resolved"}, "fixed issues"},
		{[]string{"tested", "verified"}, "tested functionality"},
		{[]string{"deployed", "released"}, "deployed changes"},
		{[]string{"updated", "modified"}, "updated configurations"},
		{[]string{"reviewed", "code review"}, "reviewed code"},
		{[]string{"investigating", "looking into"}, "investigated problems"},
	}
	
	for _, actionGroup := range actions {
		for _, pattern := range actionGroup.patterns {
			if strings.Contains(lowerText, pattern) {
				return actionGroup.action
			}
		}
	}
	
	return ""
}

// createIntelligentSummary analyzes comment content and creates a meaningful summary
func (e *EmbeddedLLM) createIntelligentSummary(text string) string {
	if text == "" {
		return "Empty comment"
	}
	
	// Clean up the text
	text = strings.TrimSpace(text)
	lowerText := strings.ToLower(text)
	
	// Analyze content patterns and create contextual summaries
	
	// PR/Merge related activities
	if strings.Contains(lowerText, "merged") && (strings.Contains(lowerText, "pr") || strings.Contains(lowerText, "pull request")) {
		if strings.Contains(lowerText, "ready to be deployed") || strings.Contains(lowerText, "deploy") {
			return "Merged PR - ready for deployment"
		}
		return "Merged pull request with changes"
	}
	
	if strings.Contains(lowerText, "created") && (strings.Contains(lowerText, "pr") || strings.Contains(lowerText, "pull request")) {
		if strings.Contains(lowerText, "foundation") || strings.Contains(lowerText, "initial") {
			return "Created foundational PR for upcoming work"
		}
		return "Created new pull request"
	}
	
	// Infrastructure and configuration work
	if strings.Contains(lowerText, "terraform") || strings.Contains(lowerText, "spacelift") {
		if strings.Contains(lowerText, "secrets") && strings.Contains(lowerText, "aws") {
			return "Configured AWS secrets management infrastructure"
		}
		if strings.Contains(lowerText, "applied") || strings.Contains(lowerText, "plan") {
			return "Updated Terraform infrastructure configuration"
		}
		return "Infrastructure configuration work"
	}
	
	// Database related work
	if strings.Contains(lowerText, "database") || strings.Contains(lowerText, "sql") {
		if strings.Contains(lowerText, "permissions") || strings.Contains(lowerText, "user") {
			return "Investigated database permission configuration"
		}
		if strings.Contains(lowerText, "lbcat") || strings.Contains(lowerText, "lbuser") {
			return "Worked on database user setup and permissions"
		}
		return "Database configuration and setup"
	}
	
	// AWS/Cloud work
	if strings.Contains(lowerText, "aws") || strings.Contains(lowerText, "oidc") {
		if strings.Contains(lowerText, "role") || strings.Contains(lowerText, "authentication") {
			return "Set up AWS OIDC authentication and roles"
		}
		return "AWS cloud configuration work"
	}
	
	// VPC/Networking
	if strings.Contains(lowerText, "vpc") || strings.Contains(lowerText, "endpoint") {
		if strings.Contains(lowerText, "ecr") || strings.Contains(lowerText, "private") {
			return "Fixed VPC endpoint configuration for ECR"
		}
		return "Network and VPC configuration"
	}
	
	// Testing and validation
	if strings.Contains(lowerText, "test") && (strings.Contains(lowerText, "performed") || strings.Contains(lowerText, "validated")) {
		return "Performed testing and validation"
	}
	
	// Questions and discussions
	if strings.Contains(lowerText, "how are you proposing") || strings.Contains(lowerText, "do you remember") || strings.Contains(lowerText, "?") {
		if strings.Contains(lowerText, "triggering") || strings.Contains(lowerText, "liquibase") {
			return "Discussed Liquibase implementation approach and alternatives"
		}
		return "Technical discussion and problem-solving"
	}
	
	// Simple completion indicators
	if strings.TrimSpace(lowerText) == "added!" {
		return "Completed requested changes"
	}
	
	// Documentation work
	if strings.Contains(lowerText, "documentation") || strings.Contains(lowerText, "hands-on") {
		return "Created documentation and guides"
	}
	
	// Security improvements
	if strings.Contains(lowerText, "security improvement") {
		return "Implemented security improvements"
	}
	
	// Extract first meaningful sentence that contains action verbs
	sentences := strings.Split(text, ".")
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 15 && len(sentence) < 120 {
			sentenceLower := strings.ToLower(sentence)
			actionVerbs := []string{"created", "implemented", "configured", "set up", "fixed", "updated", "deployed", "merged", "added", "performed"}
			for _, verb := range actionVerbs {
				if strings.Contains(sentenceLower, verb) {
					return sentence
				}
			}
		}
	}
	
	// Fallback: extract key topics if no action found
	topics := e.extractTopics(text)
	if len(topics) > 0 {
		return "Work on " + strings.Join(topics, ", ")
	}
	
	// Final fallback: first part of comment
	if len(text) > 100 {
		firstPart := text[:100]
		if lastSpace := strings.LastIndex(firstPart, " "); lastSpace > 50 {
			return firstPart[:lastSpace] + "..."
		}
		return firstPart + "..."
	}
	
	return text
}

// extractTopics identifies main technical topics in the text
func (e *EmbeddedLLM) extractTopics(text string) []string {
	lowerText := strings.ToLower(text)
	var topics []string
	
	topicMap := map[string]string{
		"terraform":     "Terraform",
		"spacelift":     "Spacelift",
		"aws":          "AWS",
		"database":     "database",
		"permissions":  "permissions",
		"secrets":      "secrets management",
		"oidc":         "OIDC",
		"vpc":          "VPC",
		"ecr":          "ECR",
		"liquibase":    "Liquibase",
		"pr":           "pull requests",
		"testing":      "testing",
		"deployment":   "deployment",
	}
	
	for keyword, topic := range topicMap {
		if strings.Contains(lowerText, keyword) {
			topics = append(topics, topic)
			if len(topics) >= 3 { // Limit topics
				break
			}
		}
	}
	
	return topics
}

// extractDetailedCommentSummary creates a more detailed summary that includes specific technical context
func (e *EmbeddedLLM) extractDetailedCommentSummary(text string) string {
	lowerText := strings.ToLower(text)
	
	// Look for specific technical actions with more context
	if strings.Contains(lowerText, "merged") && strings.Contains(lowerText, "pr") {
		return "Merged PR - ready for deployment"
	}
	
	if strings.Contains(lowerText, "created") && (strings.Contains(lowerText, "pr") || strings.Contains(lowerText, "pull request")) {
		return "Created initial PR with foundation work"
	}
	
	if strings.Contains(lowerText, "terraform") && strings.Contains(lowerText, "apply") {
		return "Updated Terraform configuration"
	}
	
	if strings.Contains(lowerText, "deployed") || strings.Contains(lowerText, "ready to be deployed") {
		return "Completed development - ready for deployment"
	}
	
	if strings.Contains(lowerText, "spacelift") && strings.Contains(lowerText, "secrets") {
		return "Configured secrets management infrastructure"
	}
	
	if strings.Contains(lowerText, "aws") && strings.Contains(lowerText, "oidc") {
		return "Set up AWS OIDC authentication"
	}
	
	if strings.Contains(lowerText, "database") && strings.Contains(lowerText, "permissions") {
		return "Investigated database permission configuration"
	}
	
	if strings.Contains(lowerText, "ecr") && strings.Contains(lowerText, "vpc") {
		return "Fixed VPC endpoint configuration"
	}
	
	// Look for questions or discussions
	if strings.Contains(lowerText, "how are you proposing") || strings.Contains(lowerText, "do you remember") {
		return "Discussed implementation approach and alternatives"
	}
	
	// Look for completion indicators
	if strings.Contains(lowerText, "added!") {
		return "Completed requested changes"
	}
	
	// Extract first meaningful sentence if it contains technical terms
	sentences := strings.Split(text, ".")
	if len(sentences) > 0 {
		firstSentence := strings.TrimSpace(sentences[0])
		if len(firstSentence) > 20 && len(firstSentence) < 100 {
			// Check if it contains technical terms
			techTerms := []string{"terraform", "aws", "database", "pr", "merged", "deployed", "configured", "implemented"}
			for _, term := range techTerms {
				if strings.Contains(strings.ToLower(firstSentence), term) {
					return firstSentence
				}
			}
		}
	}
	
	return ""
}