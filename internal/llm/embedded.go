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

	// Extract basic information from issues
	var activities []string
	var topics []string
	
	for _, issue := range issues {
		// Analyze issue summary and description for key activities
		text := strings.ToLower(issue.Fields.Summary + " " + issue.Fields.Description.Text)
		
		// Extract activities based on issue content
		issueActivities := e.extractActivitiesFromText(text)
		activities = append(activities, issueActivities...)
		
		// Extract technical topics
		issueTopics := e.extractTopicsFromText(text)
		topics = append(topics, issueTopics...)
	}
	
	// Process worklog entries
	for _, worklog := range worklogs {
		if worklog.Comment != "" {
			text := strings.ToLower(worklog.Comment)
			worklogActivities := e.extractActivitiesFromText(text)
			activities = append(activities, worklogActivities...)
			
			worklogTopics := e.extractTopicsFromText(text)
			topics = append(topics, worklogTopics...)
		}
	}
	
	return e.buildSummaryFromActivities(activities, topics, len(issues), len(worklogs))
}

// GenerateStandupSummaryWithComments creates an enhanced summary using comment data
func (e *EmbeddedLLM) GenerateStandupSummaryWithComments(issues []jira.Issue, comments []jira.Comment, worklogs []jira.WorklogEntry) (string, error) {
	if len(issues) == 0 && len(comments) == 0 && len(worklogs) == 0 {
		return "No recent activity to report", nil
	}
	
	// Use enhanced data processing pipeline
	processor := NewEnhancedDataProcessor(false) // debug=false for production
	processedData, err := processor.ProcessIssuesWithComments(issues, comments)
	if err != nil {
		// Fallback to original method if enhanced processing fails
		return e.generateFallbackSummary(issues, comments, worklogs)
	}
	
	// Use technical pattern matcher for enhanced analysis
	patternMatcher := NewTechnicalPatternMatcher(false)
	
	// Generate enhanced summary using processed data
	return e.generateEnhancedStandupSummary(processedData, patternMatcher, worklogs)
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
	
	// Extract individual summaries and activities
	var activities []string
	var topics []string
	var completedWork []string
	var inProgressWork []string
	
	for _, comment := range comments {
		if comment.Body.Text == "" {
			continue
		}
		
		text := strings.ToLower(comment.Body.Text)
		
		// Categorize activities by completion status
		if e.isCompletedActivity(text) {
			activity := e.extractActivitySummary(comment.Body.Text)
			if activity != "" {
				completedWork = append(completedWork, activity)
			}
		} else if e.isInProgressActivity(text) {
			activity := e.extractActivitySummary(comment.Body.Text)
			if activity != "" {
				inProgressWork = append(inProgressWork, activity)
			}
		} else {
			// General activity
			activity := e.extractActivitySummary(comment.Body.Text)
			if activity != "" {
				activities = append(activities, activity)
			}
		}
		
		// Extract topics
		commentTopics := e.extractTopics(comment.Body.Text)
		topics = append(topics, commentTopics...)
	}
	
	// Remove duplicates and prioritize
	completedWork = e.removeDuplicateActivities(completedWork)
	inProgressWork = e.removeDuplicateActivities(inProgressWork)
	activities = e.removeDuplicateActivities(activities)
	topics = e.removeDuplicateStrings(topics)
	
	// Build prioritized summary
	return e.buildMultiCommentSummary(completedWork, inProgressWork, activities, topics)
}

// isCompletedActivity checks if the text indicates completed work
func (e *EmbeddedLLM) isCompletedActivity(text string) bool {
	completionIndicators := []string{
		"completed", "finished", "done", "resolved", "merged", "deployed", 
		"implemented", "fixed", "released", "closed", "solved",
	}
	
	for _, indicator := range completionIndicators {
		if strings.Contains(text, indicator) {
			return true
		}
	}
	return false
}

// isInProgressActivity checks if the text indicates work in progress
func (e *EmbeddedLLM) isInProgressActivity(text string) bool {
	progressIndicators := []string{
		"working on", "in progress", "currently", "investigating", 
		"looking into", "started", "beginning", "planning",
	}
	
	for _, indicator := range progressIndicators {
		if strings.Contains(text, indicator) {
			return true
		}
	}
	return false
}

// extractActivitySummary extracts a concise activity summary from comment text
func (e *EmbeddedLLM) extractActivitySummary(text string) string {
	// Use the improved createIntelligentSummary but limit length
	summary := e.createIntelligentSummary(text)
	if len(summary) > 80 {
		return e.shortenText(summary, 80)
	}
	return summary
}

// removeDuplicateActivities removes similar activities from a slice
func (e *EmbeddedLLM) removeDuplicateActivities(activities []string) []string {
	if len(activities) <= 1 {
		return activities
	}
	
	var unique []string
	seen := make(map[string]bool)
	
	for _, activity := range activities {
		// Create a normalized version for comparison
		normalized := e.normalizeActivity(activity)
		if !seen[normalized] {
			seen[normalized] = true
			unique = append(unique, activity)
		}
	}
	
	return unique
}

// normalizeActivity creates a normalized version of an activity for comparison
func (e *EmbeddedLLM) normalizeActivity(activity string) string {
	// Convert to lowercase and remove common variations
	normalized := strings.ToLower(activity)
	
	// Remove common prefixes/suffixes that don't change meaning
	prefixesToRemove := []string{"working on ", "completed ", "implemented ", "fixed "}
	for _, prefix := range prefixesToRemove {
		if strings.HasPrefix(normalized, prefix) {
			normalized = strings.TrimPrefix(normalized, prefix)
			break
		}
	}
	
	// Extract key technical terms for comparison
	keyTerms := []string{
		"terraform", "aws", "database", "vpc", "ecr", "kubernetes", "docker",
		"pipeline", "deployment", "pr", "merge", "test", "configuration",
	}
	
	for _, term := range keyTerms {
		if strings.Contains(normalized, term) {
			return term // Use the key term as the normalized identifier
		}
	}
	
	// Fallback to first few words
	words := strings.Fields(normalized)
	if len(words) > 3 {
		return strings.Join(words[:3], " ")
	}
	
	return normalized
}

// removeDuplicateStrings removes duplicate strings from a slice
func (e *EmbeddedLLM) removeDuplicateStrings(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// buildMultiCommentSummary builds a coherent summary from categorized activities
func (e *EmbeddedLLM) buildMultiCommentSummary(completed, inProgress, general []string, topics []string) string {
	var parts []string
	
	// Prioritize completed work
	if len(completed) > 0 {
		if len(completed) == 1 {
			parts = append(parts, completed[0])
		} else if len(completed) <= 3 {
			parts = append(parts, strings.Join(completed, ", "))
		} else {
			parts = append(parts, strings.Join(completed[:2], ", ")+" and other completed work")
		}
	}
	
	// Add in-progress work
	if len(inProgress) > 0 && len(parts) < 2 {
		if len(inProgress) == 1 {
			parts = append(parts, inProgress[0])
		} else {
			parts = append(parts, strings.Join(inProgress[:1], ", ")+" (in progress)")
		}
	}
	
	// Add general activities if we don't have enough content
	if len(parts) == 0 && len(general) > 0 {
		if len(general) == 1 {
			parts = append(parts, general[0])
		} else {
			parts = append(parts, strings.Join(general[:2], ", "))
		}
	}
	
	// If we still don't have content, use topics
	if len(parts) == 0 && len(topics) > 0 {
		if len(topics) <= 3 {
			return "Technical work on " + strings.Join(topics, ", ")
		} else {
			return "Technical work on " + strings.Join(topics[:3], ", ") + " and more"
		}
	}
	
	// Combine parts into final summary
	if len(parts) == 0 {
		return "Multiple development activities"
	}
	
	return strings.Join(parts, "; ")
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
	
	// First, try to extract structured technical activities
	if summary := e.extractTechnicalActivity(lowerText, text); summary != "" {
		return summary
	}
	
	// Then try to identify development workflow activities
	if summary := e.extractDevelopmentActivity(lowerText, text); summary != "" {
		return summary
	}
	
	// Try to extract infrastructure and deployment activities
	if summary := e.extractInfrastructureActivity(lowerText, text); summary != "" {
		return summary
	}
	
	// Look for general progress indicators
	if summary := e.extractProgressActivity(lowerText, text); summary != "" {
		return summary
	}
	
	// Extract first meaningful sentence that contains action verbs
	if summary := e.extractActionSentence(text); summary != "" {
		return summary
	}
	
	// Fallback: extract key topics if no action found
	topics := e.extractTopics(text)
	if len(topics) > 0 {
		return "Work on " + strings.Join(topics, ", ")
	}
	
	// Final fallback: first part of comment
	return e.shortenText(text, 100)
}

// extractTechnicalActivity identifies specific technical work patterns
func (e *EmbeddedLLM) extractTechnicalActivity(lowerText, originalText string) string {
	// PR/Merge related activities with enhanced context
	if strings.Contains(lowerText, "merged") {
		if strings.Contains(lowerText, "pr") || strings.Contains(lowerText, "pull request") {
			if strings.Contains(lowerText, "ready to be deployed") || strings.Contains(lowerText, "deploy") {
				return "Merged PR - ready for deployment"
			}
			if strings.Contains(lowerText, "fix") || strings.Contains(lowerText, "bug") {
				return "Merged bug fix PR"
			}
			if strings.Contains(lowerText, "feature") || strings.Contains(lowerText, "enhancement") {
				return "Merged feature PR"
			}
			return "Merged pull request with changes"
		}
		if strings.Contains(lowerText, "branch") {
			return "Merged development branch"
		}
	}
	
	if strings.Contains(lowerText, "created") {
		if strings.Contains(lowerText, "pr") || strings.Contains(lowerText, "pull request") {
			if strings.Contains(lowerText, "foundation") || strings.Contains(lowerText, "initial") {
				return "Created foundational PR for upcoming work"
			}
			if strings.Contains(lowerText, "draft") {
				return "Created draft PR for review"
			}
			return "Created new pull request"
		}
		if strings.Contains(lowerText, "branch") {
			return "Created new development branch"
		}
		if strings.Contains(lowerText, "ticket") || strings.Contains(lowerText, "issue") {
			return "Created new ticket/issue"
		}
	}
	
	// Code review activities
	if strings.Contains(lowerText, "review") {
		if strings.Contains(lowerText, "approved") || strings.Contains(lowerText, "lgtm") {
			return "Approved code review"
		}
		if strings.Contains(lowerText, "requested changes") || strings.Contains(lowerText, "feedback") {
			return "Provided code review feedback"
		}
		if strings.Contains(lowerText, "addressed") || strings.Contains(lowerText, "fixed comments") {
			return "Addressed code review comments"
		}
		return "Conducted code review"
	}
	
	return ""
}

// extractDevelopmentActivity identifies development-specific work
func (e *EmbeddedLLM) extractDevelopmentActivity(lowerText, originalText string) string {
	// Testing activities
	if strings.Contains(lowerText, "test") {
		if strings.Contains(lowerText, "passed") || strings.Contains(lowerText, "successful") {
			return "Tests passed successfully"
		}
		if strings.Contains(lowerText, "failed") || strings.Contains(lowerText, "failing") {
			return "Investigated test failures"
		}
		if strings.Contains(lowerText, "added") || strings.Contains(lowerText, "wrote") {
			return "Added new tests"
		}
		if strings.Contains(lowerText, "unit") {
			return "Worked on unit tests"
		}
		if strings.Contains(lowerText, "integration") {
			return "Worked on integration tests"
		}
		if strings.Contains(lowerText, "performed") || strings.Contains(lowerText, "validated") {
			return "Performed testing and validation"
		}
		return "Testing work"
	}
	
	// Bug fixing
	if strings.Contains(lowerText, "fix") || strings.Contains(lowerText, "bug") {
		if strings.Contains(lowerText, "resolved") || strings.Contains(lowerText, "solved") {
			return "Resolved bug/issue"
		}
		if strings.Contains(lowerText, "investigating") || strings.Contains(lowerText, "debugging") {
			return "Investigating bug/issue"
		}
		if strings.Contains(lowerText, "hotfix") {
			return "Applied hotfix"
		}
		return "Bug fix work"
	}
	
	// Implementation work
	if strings.Contains(lowerText, "implement") {
		if strings.Contains(lowerText, "feature") {
			return "Implemented new feature"
		}
		if strings.Contains(lowerText, "api") {
			return "Implemented API functionality"
		}
		if strings.Contains(lowerText, "endpoint") {
			return "Implemented API endpoint"
		}
		return "Implementation work"
	}
	
	// Refactoring
	if strings.Contains(lowerText, "refactor") {
		return "Code refactoring work"
	}
	
	return ""
}

// extractInfrastructureActivity identifies infrastructure and deployment activities
func (e *EmbeddedLLM) extractInfrastructureActivity(lowerText, originalText string) string {
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
	
	// Kubernetes/Container work
	if strings.Contains(lowerText, "kubernetes") || strings.Contains(lowerText, "k8s") || strings.Contains(lowerText, "docker") {
		if strings.Contains(lowerText, "deploy") || strings.Contains(lowerText, "rollout") {
			return "Deployed Kubernetes applications"
		}
		if strings.Contains(lowerText, "config") || strings.Contains(lowerText, "manifest") {
			return "Updated Kubernetes configuration"
		}
		return "Kubernetes/container work"
	}
	
	// CI/CD Pipeline work
	if strings.Contains(lowerText, "pipeline") || strings.Contains(lowerText, "ci/cd") || strings.Contains(lowerText, "jenkins") {
		if strings.Contains(lowerText, "fix") || strings.Contains(lowerText, "broke") {
			return "Fixed CI/CD pipeline issues"
		}
		if strings.Contains(lowerText, "deploy") || strings.Contains(lowerText, "build") {
			return "Updated deployment pipeline"
		}
		return "CI/CD pipeline work"
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
	
	return ""
}

// extractProgressActivity identifies general progress indicators
func (e *EmbeddedLLM) extractProgressActivity(lowerText, originalText string) string {
	// Progress completion indicators
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
	
	return ""
}

// extractActionSentence extracts meaningful sentences with action verbs
func (e *EmbeddedLLM) extractActionSentence(text string) string {
	sentences := strings.Split(text, ".")
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 15 && len(sentence) < 120 {
			sentenceLower := strings.ToLower(sentence)
			actionVerbs := []string{
				"created", "implemented", "configured", "set up", "fixed", 
				"updated", "deployed", "merged", "added", "performed",
				"completed", "resolved", "tested", "investigated", "reviewed",
			}
			for _, verb := range actionVerbs {
				if strings.Contains(sentenceLower, verb) {
					return sentence
				}
			}
		}
	}
	return ""
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

// generateFallbackSummary provides fallback when enhanced processing fails
func (e *EmbeddedLLM) generateFallbackSummary(issues []jira.Issue, comments []jira.Comment, worklogs []jira.WorklogEntry) (string, error) {
	// Collect all activities and categorize them using original method
	var activities []string
	var topics []string
	
	// Process comments (most detailed source of information)
	for _, comment := range comments {
		text := strings.ToLower(comment.Body.Text)
		
		// Extract specific activities from comments
		commentActivities := e.extractActivitiesFromText(text)
		activities = append(activities, commentActivities...)
		
		// Extract technical topics
		commentTopics := e.extractTopicsFromText(text)
		topics = append(topics, commentTopics...)
	}
	
	// Process issues for additional context
	for _, issue := range issues {
		text := strings.ToLower(issue.Fields.Summary + " " + issue.Fields.Description.Text)
		issueActivities := e.extractActivitiesFromText(text)
		activities = append(activities, issueActivities...)
		
		issueTopics := e.extractTopicsFromText(text)
		topics = append(topics, issueTopics...)
	}
	
	// Process worklog entries
	for _, worklog := range worklogs {
		if worklog.Comment != "" {
			text := strings.ToLower(worklog.Comment)
			worklogActivities := e.extractActivitiesFromText(text)
			activities = append(activities, worklogActivities...)
			
			worklogTopics := e.extractTopicsFromText(text)
			topics = append(topics, worklogTopics...)
		}
	}
	
	return e.buildSummaryFromActivities(activities, topics, len(issues), len(worklogs))
}

// generateEnhancedStandupSummary creates summary using ProcessedData and TechnicalPatternMatcher
func (e *EmbeddedLLM) generateEnhancedStandupSummary(processedData *ProcessedData, patternMatcher *TechnicalPatternMatcher, worklogs []jira.WorklogEntry) (string, error) {
	if len(processedData.Issues) == 0 && len(worklogs) == 0 {
		return "No recent activity to report", nil
	}
	
	// Extract enhanced activities from processed data
	var completedWork []string
	var inProgressWork []string
	var technicalTopics []string
	var keyActivities []string
	
	// Process enhanced issues
	for _, issue := range processedData.Issues {
		// Add work summary if available
		if issue.WorkSummary != "" {
			switch issue.CompletionStatus {
			case "completed", "done", "resolved":
				completedWork = append(completedWork, issue.WorkSummary)
			case "in_progress", "working", "active":
				inProgressWork = append(inProgressWork, issue.WorkSummary)
			default:
				keyActivities = append(keyActivities, issue.WorkSummary)
			}
		}
		
		// Extract key activities
		keyActivities = append(keyActivities, issue.KeyActivities...)
		
		// Process comments with enhanced analysis
		for _, comment := range issue.ProcessedComments {
			// Use technical pattern matcher for enhanced insights
			if comment.Original.Body.Text != "" {
				patterns, err := patternMatcher.MatchAllPatterns(comment.Original.Body.Text)
				if err == nil {
					// Extract insights from pattern matches
					e.extractInsightsFromPatterns(patterns, &technicalTopics, &keyActivities)
				}
			}
			
			// Add technical terms and topics
			technicalTopics = append(technicalTopics, comment.TechnicalTerms...)
			technicalTopics = append(technicalTopics, comment.KeyTopics...)
		}
	}
	
	// Extract technical context
	if processedData.TechnicalContext != nil {
		technicalTopics = append(technicalTopics, processedData.TechnicalContext.Technologies...)
		
		// Add deployment activities
		for _, deployment := range processedData.TechnicalContext.Deployments {
			activity := fmt.Sprintf("%s %s (%s)", strings.Title(deployment.Action), deployment.Component, deployment.Status)
			switch deployment.Status {
			case "completed":
				completedWork = append(completedWork, activity)
			case "in_progress":
				inProgressWork = append(inProgressWork, activity)
			default:
				keyActivities = append(keyActivities, activity)
			}
		}
		
		// Add infrastructure work
		for _, infra := range processedData.TechnicalContext.Infrastructure {
			activity := fmt.Sprintf("%s %s work (%s)", strings.Title(infra.Action), infra.Type, infra.Status)
			switch infra.Status {
			case "completed":
				completedWork = append(completedWork, activity)
			case "in_progress":
				inProgressWork = append(inProgressWork, activity)
			default:
				keyActivities = append(keyActivities, activity)
			}
		}
	}
	
	// Process worklog entries with enhanced analysis
	for _, worklog := range worklogs {
		if worklog.Comment != "" {
			patterns, err := patternMatcher.MatchAllPatterns(worklog.Comment)
			if err == nil {
				e.extractInsightsFromPatterns(patterns, &technicalTopics, &keyActivities)
			}
		}
	}
	
	// Remove duplicates and build final summary
	completedWork = e.removeDuplicateStrings(completedWork)
	inProgressWork = e.removeDuplicateStrings(inProgressWork)
	keyActivities = e.removeDuplicateStrings(keyActivities)
	technicalTopics = e.removeDuplicateStrings(technicalTopics)
	
	return e.buildEnhancedSummary(completedWork, inProgressWork, keyActivities, technicalTopics)
}

// extractInsightsFromPatterns extracts insights from pattern matching results
func (e *EmbeddedLLM) extractInsightsFromPatterns(patterns map[string]interface{}, technicalTopics *[]string, keyActivities *[]string) {
	// Extract infrastructure patterns
	if infraPatterns, ok := patterns["infrastructure"].([]InfrastructurePattern); ok {
		for _, pattern := range infraPatterns {
			if pattern.Confidence > 0.7 { // Only high-confidence matches
				*technicalTopics = append(*technicalTopics, pattern.Type)
				if pattern.Action != "unknown" && pattern.Component != "general" {
					activity := fmt.Sprintf("%s %s %s", strings.Title(pattern.Action), pattern.Type, pattern.Component)
					*keyActivities = append(*keyActivities, activity)
				}
			}
		}
	}
	
	// Extract deployment patterns
	if deployPatterns, ok := patterns["deployment"].([]DeploymentPattern); ok {
		for _, pattern := range deployPatterns {
			if pattern.Confidence > 0.7 {
				*technicalTopics = append(*technicalTopics, "deployment")
				if pattern.Environment != "unknown" {
					activity := fmt.Sprintf("%s to %s", strings.Title(pattern.Type), pattern.Environment)
					*keyActivities = append(*keyActivities, activity)
				}
			}
		}
	}
	
	// Extract development patterns
	if devPatterns, ok := patterns["development"].([]DevelopmentPattern); ok {
		for _, pattern := range devPatterns {
			if pattern.Confidence > 0.7 {
				*technicalTopics = append(*technicalTopics, pattern.Type)
				if pattern.Action != "unknown" {
					activity := fmt.Sprintf("%s %s", strings.Title(pattern.Action), pattern.Type)
					*keyActivities = append(*keyActivities, activity)
				}
			}
		}
	}
}

// buildEnhancedSummary builds the final enhanced summary
func (e *EmbeddedLLM) buildEnhancedSummary(completed, inProgress, activities, topics []string) string {
	var summaryParts []string
	
	// Prioritize completed work
	if len(completed) > 0 {
		if len(completed) == 1 {
			summaryParts = append(summaryParts, "Completed: "+completed[0])
		} else if len(completed) <= 3 {
			summaryParts = append(summaryParts, "Completed: "+strings.Join(completed, ", "))
		} else {
			summaryParts = append(summaryParts, fmt.Sprintf("Completed: %s and %d other items", strings.Join(completed[:2], ", "), len(completed)-2))
		}
	}
	
	// Add in-progress work
	if len(inProgress) > 0 && len(summaryParts) < 2 {
		if len(inProgress) == 1 {
			summaryParts = append(summaryParts, "Working on: "+inProgress[0])
		} else {
			summaryParts = append(summaryParts, "Working on: "+strings.Join(inProgress[:min(2, len(inProgress))], ", "))
		}
	}
	
	// Add other key activities if we need more content
	if len(summaryParts) < 2 && len(activities) > 0 {
		filteredActivities := e.filterUniqueActivities(activities, append(completed, inProgress...))
		if len(filteredActivities) > 0 {
			if len(filteredActivities) == 1 {
				summaryParts = append(summaryParts, "Also: "+filteredActivities[0])
			} else {
				summaryParts = append(summaryParts, "Also: "+strings.Join(filteredActivities[:min(2, len(filteredActivities))], ", "))
			}
		}
	}
	
	// Add technical context if summary is still short
	if len(summaryParts) <= 1 && len(topics) > 0 {
		uniqueTopics := e.removeDuplicateStrings(topics)
		if len(uniqueTopics) <= 3 {
			summaryParts = append(summaryParts, "Focus areas: "+strings.Join(uniqueTopics, ", "))
		} else {
			summaryParts = append(summaryParts, "Focus areas: "+strings.Join(uniqueTopics[:3], ", ")+" and more")
		}
	}
	
	// Fallback if no content
	if len(summaryParts) == 0 {
		return "Various development and infrastructure activities"
	}
	
	return strings.Join(summaryParts, ". ")
}

// filterUniqueActivities filters out activities that are similar to already included ones
func (e *EmbeddedLLM) filterUniqueActivities(activities, existing []string) []string {
	var unique []string
	existingNormalized := make(map[string]bool)
	
	// Normalize existing activities
	for _, activity := range existing {
		normalized := e.normalizeActivity(activity)
		existingNormalized[normalized] = true
	}
	
	// Filter unique activities
	for _, activity := range activities {
		normalized := e.normalizeActivity(activity)
		if !existingNormalized[normalized] {
			unique = append(unique, activity)
			existingNormalized[normalized] = true
		}
	}
	
	return unique
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

// extractActivitiesFromText identifies key activities from text content
func (e *EmbeddedLLM) extractActivitiesFromText(text string) []string {
	var activities []string
	
	// PR/Merge related activities
	if strings.Contains(text, "merged") && strings.Contains(text, "pr") {
		if strings.Contains(text, "ready to be deployed") || strings.Contains(text, "deploy") {
			activities = append(activities, "Merged PRs for deployment")
		} else {
			activities = append(activities, "Merged pull requests")
		}
	} else if strings.Contains(text, "created") && (strings.Contains(text, "pr") || strings.Contains(text, "pull request")) {
		if strings.Contains(text, "foundation") || strings.Contains(text, "initial") {
			activities = append(activities, "Created foundational PRs")
		} else {
			activities = append(activities, "Created pull requests")
		}
	}
	
	// Infrastructure and configuration work
	if strings.Contains(text, "terraform") || strings.Contains(text, "spacelift") {
		if strings.Contains(text, "secrets") && strings.Contains(text, "aws") {
			activities = append(activities, "Configured AWS secrets management infrastructure")
		} else if strings.Contains(text, "applied") || strings.Contains(text, "plan") {
			activities = append(activities, "Updated Terraform infrastructure configuration")
		} else {
			activities = append(activities, "Infrastructure configuration work")
		}
	}
	
	// Database related work
	if strings.Contains(text, "database") || strings.Contains(text, "sql") {
		if strings.Contains(text, "permissions") || strings.Contains(text, "user") {
			activities = append(activities, "Worked on database permissions")
		} else if strings.Contains(text, "migration") {
			activities = append(activities, "Database migration work")
		} else {
			activities = append(activities, "Database configuration and setup")
		}
	}
	
	// AWS/Cloud work
	if strings.Contains(text, "aws") || strings.Contains(text, "oidc") {
		if strings.Contains(text, "role") || strings.Contains(text, "authentication") {
			activities = append(activities, "Set up AWS OIDC authentication")
		} else {
			activities = append(activities, "AWS cloud configuration work")
		}
	}
	
	// VPC/Networking
	if strings.Contains(text, "vpc") || strings.Contains(text, "endpoint") {
		if strings.Contains(text, "ecr") || strings.Contains(text, "private") {
			activities = append(activities, "Fixed VPC endpoint configuration")
		} else {
			activities = append(activities, "Network and VPC configuration")
		}
	}
	
	// Testing and validation
	if strings.Contains(text, "test") && (strings.Contains(text, "performed") || strings.Contains(text, "validated")) {
		activities = append(activities, "Performed testing and validation")
	}
	
	// Deployment activities
	if strings.Contains(text, "deployed") || strings.Contains(text, "deployment") {
		if strings.Contains(text, "ready") {
			activities = append(activities, "Prepared deployments")
		} else {
			activities = append(activities, "Deployed changes")
		}
	}
	
	// Development activities
	if strings.Contains(text, "implemented") || strings.Contains(text, "developed") {
		activities = append(activities, "Implemented features")
	}
	
	if strings.Contains(text, "fixed") || strings.Contains(text, "resolved") {
		activities = append(activities, "Fixed issues")
	}
	
	if strings.Contains(text, "configured") || strings.Contains(text, "setup") {
		activities = append(activities, "Configuration work")
	}
	
	// Documentation work
	if strings.Contains(text, "documentation") || strings.Contains(text, "guide") {
		activities = append(activities, "Created documentation")
	}
	
	return activities
}

// extractTopicsFromText identifies technical topics from text content
func (e *EmbeddedLLM) extractTopicsFromText(text string) []string {
	var topics []string
	
	topicMap := map[string]string{
		"terraform":     "Terraform",
		"spacelift":     "Spacelift",
		"aws":          "AWS",
		"database":     "Database",
		"permissions":  "Permissions",
		"secrets":      "Secrets Management",
		"oidc":         "OIDC",
		"vpc":          "VPC",
		"ecr":          "ECR",
		"liquibase":    "Liquibase",
		"pr":           "Pull Requests",
		"testing":      "Testing",
		"deployment":   "Deployment",
		"kubernetes":   "Kubernetes",
		"k8s":          "Kubernetes",
		"docker":       "Docker",
		"ci/cd":        "CI/CD",
		"pipeline":     "Pipeline",
		"monitoring":   "Monitoring",
		"security":     "Security",
		"authentication": "Authentication",
		"api":          "API",
		"microservice": "Microservices",
		"integration":  "Integration",
	}
	
	for keyword, topic := range topicMap {
		if strings.Contains(text, keyword) {
			topics = append(topics, topic)
			if len(topics) >= 5 { // Limit topics to avoid too many
				break
			}
		}
	}
	
	return topics
}

// buildSummaryFromActivities creates a coherent summary from activities and topics
func (e *EmbeddedLLM) buildSummaryFromActivities(activities, topics []string, issueCount, worklogCount int) (string, error) {
	// Remove duplicates
	activities = e.removeDuplicates(activities)
	topics = e.removeDuplicates(topics)
	
	var summaryParts []string
	
	// Add activity summary
	if len(activities) > 0 {
		if len(activities) == 1 {
			summaryParts = append(summaryParts, activities[0])
		} else if len(activities) <= 3 {
			summaryParts = append(summaryParts, strings.Join(activities, ", "))
		} else {
			// Prioritize most important activities
			prioritizedActivities := e.prioritizeActivities(activities)
			summaryParts = append(summaryParts, strings.Join(prioritizedActivities[:3], ", ")+" and other development work")
		}
	}
	
	// Add technical focus areas if we have topics but few activities
	if len(topics) > 0 && len(activities) <= 1 {
		if len(topics) <= 3 {
			summaryParts = append(summaryParts, fmt.Sprintf("Focus areas: %s", strings.Join(topics, ", ")))
		} else {
			summaryParts = append(summaryParts, fmt.Sprintf("Focus areas: %s", strings.Join(topics[:3], ", ")))
		}
	}
	
	// Fallback if no meaningful activities or topics found
	if len(summaryParts) == 0 {
		if issueCount > 0 {
			summaryParts = append(summaryParts, fmt.Sprintf("Worked on %d development task(s)", issueCount))
		}
		if worklogCount > 0 {
			summaryParts = append(summaryParts, fmt.Sprintf("with %d worklog entries", worklogCount))
		}
	}
	
	if len(summaryParts) == 0 {
		return "No recent activity to report", nil
	}
	
	return strings.Join(summaryParts, ". "), nil
}

// prioritizeActivities sorts activities by importance for standup summaries
func (e *EmbeddedLLM) prioritizeActivities(activities []string) []string {
	// Define priority order (higher priority first)
	priorityKeywords := []struct {
		keywords []string
		priority int
	}{
		{[]string{"deployed", "deployment", "ready for deployment"}, 10},
		{[]string{"merged", "pull request"}, 9},
		{[]string{"fixed", "resolved", "bug"}, 8},
		{[]string{"implemented", "feature"}, 7},
		{[]string{"configured", "infrastructure", "terraform", "aws"}, 6},
		{[]string{"tested", "testing", "validation"}, 5},
		{[]string{"created", "documentation"}, 4},
	}
	
	// Score each activity
	type scoredActivity struct {
		activity string
		score    int
	}
	
	var scored []scoredActivity
	for _, activity := range activities {
		score := 0
		lowerActivity := strings.ToLower(activity)
		
		for _, priorityGroup := range priorityKeywords {
			for _, keyword := range priorityGroup.keywords {
				if strings.Contains(lowerActivity, keyword) {
					score = priorityGroup.priority
					break
				}
			}
			if score > 0 {
				break
			}
		}
		
		scored = append(scored, scoredActivity{activity, score})
	}
	
	// Sort by score (highest first)
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[j].score > scored[i].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}
	
	// Extract sorted activities
	var result []string
	for _, item := range scored {
		result = append(result, item.activity)
	}
	
	return result
}

// removeDuplicates removes duplicate strings from a slice
func (e *EmbeddedLLM) removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}