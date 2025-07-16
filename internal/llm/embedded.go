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

// SummarizeComments generates a summary of user's comments from today using enhanced processing
func (e *EmbeddedLLM) SummarizeComments(comments []jira.Comment) (string, error) {
	if len(comments) == 0 {
		return "", nil
	}
	
	// Use enhanced data processing for better insights
	processor := NewEnhancedDataProcessor(false) // debug=false for production
	patternMatcher := NewTechnicalPatternMatcher(false)
	
	// Process comments to extract enhanced insights
	var processedComments []ProcessedComment
	for _, comment := range comments {
		// Use the public method from processor to process individual comments
		processedComment, err := e.processCommentWithProcessor(processor, comment)
		if err != nil {
			// Fallback to original processing if enhanced processing fails
			continue
		}
		processedComments = append(processedComments, processedComment)
	}
	
	// If enhanced processing failed for all comments, use fallback
	if len(processedComments) == 0 {
		if len(comments) == 1 {
			return e.summarizeSingleComment(comments[0]), nil
		}
		return e.summarizeMultipleComments(comments), nil
	}
	
	// Generate enhanced summary using processed comments and pattern matcher
	return e.generateEnhancedCommentSummary(processedComments, patternMatcher)
}

// processCommentWithProcessor processes a single comment using the enhanced data processor
func (e *EmbeddedLLM) processCommentWithProcessor(processor *EnhancedDataProcessor, comment jira.Comment) (ProcessedComment, error) {
	// Create a temporary issue to use the processor's processComment method
	// We need to use reflection or create a wrapper since processComment is private
	// For now, we'll implement the processing logic directly here
	
	if comment.ID == "" {
		return ProcessedComment{}, fmt.Errorf("comment ID is empty")
	}
	
	text := comment.Body.Text
	
	processedComment := ProcessedComment{
		Original:         comment,
		ExtractedActions: e.extractActionsFromText(text),
		TechnicalTerms:   e.extractTechnicalTermsFromText(text),
		WorkType:         e.determineCommentWorkTypeFromText(text),
		Sentiment:        e.determineSentimentFromText(text),
		Importance:       e.calculateCommentImportanceFromText(text),
		ActivityType:     e.determineActivityTypeFromText(text),
		CompletionStatus: e.determineCommentCompletionStatusFromText(text),
		KeyTopics:        e.extractKeyTopicsFromText(text),
	}
	
	return processedComment, nil
}

// extractActionsFromText extracts action verbs from text (helper for enhanced processing)
func (e *EmbeddedLLM) extractActionsFromText(text string) []string {
	lowerText := strings.ToLower(text)
	var actions []string
	
	actionVerbs := []string{
		"implemented", "created", "added", "built", "developed",
		"fixed", "resolved", "corrected", "debugged", "troubleshot",
		"updated", "modified", "changed", "improved", "enhanced",
		"deployed", "released", "pushed", "merged", "integrated",
		"tested", "verified", "validated", "checked", "confirmed",
		"configured", "setup", "installed", "initialized", "prepared",
		"investigated", "analyzed", "reviewed", "examined", "explored",
		"documented", "wrote", "recorded", "noted", "explained",
	}
	
	for _, verb := range actionVerbs {
		if strings.Contains(lowerText, verb) {
			actions = append(actions, verb)
		}
	}
	
	return e.removeDuplicateStrings(actions)
}

// extractTechnicalTermsFromText extracts technical terms from text (helper for enhanced processing)
func (e *EmbeddedLLM) extractTechnicalTermsFromText(text string) []string {
	lowerText := strings.ToLower(text)
	var terms []string
	
	technicalTerms := []string{
		"terraform", "spacelift", "aws", "kubernetes", "k8s", "docker",
		"database", "sql", "postgresql", "mysql", "mongodb",
		"api", "rest", "graphql", "endpoint", "microservice",
		"ci/cd", "pipeline", "jenkins", "github", "gitlab",
		"vpc", "ecr", "s3", "lambda", "ec2", "rds",
		"oauth", "oidc", "authentication", "authorization", "jwt",
		"ssl", "tls", "https", "security", "encryption",
		"monitoring", "logging", "metrics", "alerts",
		"redis", "elasticsearch", "kafka", "rabbitmq",
		"nginx", "apache", "load balancer", "proxy",
	}
	
	for _, term := range technicalTerms {
		if strings.Contains(lowerText, term) {
			terms = append(terms, term)
		}
	}
	
	return e.removeDuplicateStrings(terms)
}

// determineCommentWorkTypeFromText determines work type from comment content (helper for enhanced processing)
func (e *EmbeddedLLM) determineCommentWorkTypeFromText(text string) string {
	lowerText := strings.ToLower(text)
	
	if strings.Contains(lowerText, "terraform") || strings.Contains(lowerText, "aws") || strings.Contains(lowerText, "infrastructure") {
		return "infrastructure"
	}
	if strings.Contains(lowerText, "database") || strings.Contains(lowerText, "sql") {
		return "database"
	}
	if strings.Contains(lowerText, "deploy") || strings.Contains(lowerText, "release") {
		return "deployment"
	}
	if strings.Contains(lowerText, "test") || strings.Contains(lowerText, "testing") {
		return "testing"
	}
	if strings.Contains(lowerText, "review") || strings.Contains(lowerText, "pr") || strings.Contains(lowerText, "merge") {
		return "code_review"
	}
	if strings.Contains(lowerText, "fix") || strings.Contains(lowerText, "bug") || strings.Contains(lowerText, "error") {
		return "bug_fix"
	}
	if strings.Contains(lowerText, "security") || strings.Contains(lowerText, "auth") {
		return "security"
	}
	
	return "general"
}

// determineSentimentFromText determines the sentiment of a comment (helper for enhanced processing)
func (e *EmbeddedLLM) determineSentimentFromText(text string) string {
	lowerText := strings.ToLower(text)
	
	positiveWords := []string{"completed", "fixed", "resolved", "success", "working", "good", "great", "done"}
	negativeWords := []string{"blocked", "failed", "error", "issue", "problem", "broken", "stuck"}
	
	positiveCount := 0
	negativeCount := 0
	
	for _, word := range positiveWords {
		if strings.Contains(lowerText, word) {
			positiveCount++
		}
	}
	
	for _, word := range negativeWords {
		if strings.Contains(lowerText, word) {
			negativeCount++
		}
	}
	
	if positiveCount > negativeCount {
		return "positive"
	} else if negativeCount > positiveCount {
		return "negative"
	}
	
	return "neutral"
}

// calculateCommentImportanceFromText calculates importance score for a comment (helper for enhanced processing)
func (e *EmbeddedLLM) calculateCommentImportanceFromText(text string) int {
	lowerText := strings.ToLower(text)
	importance := 50 // Base importance
	
	// High importance indicators
	highImportanceWords := []string{"critical", "urgent", "blocked", "production", "outage", "security"}
	for _, word := range highImportanceWords {
		if strings.Contains(lowerText, word) {
			importance += 30
		}
	}
	
	// Medium importance indicators
	mediumImportanceWords := []string{"completed", "deployed", "merged", "resolved", "implemented"}
	for _, word := range mediumImportanceWords {
		if strings.Contains(lowerText, word) {
			importance += 20
		}
	}
	
	// Technical complexity indicators
	technicalWords := []string{"terraform", "kubernetes", "database", "api", "infrastructure"}
	for _, word := range technicalWords {
		if strings.Contains(lowerText, word) {
			importance += 10
		}
	}
	
	// Cap at 100
	if importance > 100 {
		importance = 100
	}
	
	return importance
}

// determineActivityTypeFromText determines the type of activity from comment (helper for enhanced processing)
func (e *EmbeddedLLM) determineActivityTypeFromText(text string) string {
	lowerText := strings.ToLower(text)
	
	if strings.Contains(lowerText, "completed") || strings.Contains(lowerText, "finished") || strings.Contains(lowerText, "done") {
		return "completion"
	}
	if strings.Contains(lowerText, "started") || strings.Contains(lowerText, "beginning") || strings.Contains(lowerText, "working on") {
		return "initiation"
	}
	if strings.Contains(lowerText, "updated") || strings.Contains(lowerText, "modified") || strings.Contains(lowerText, "changed") {
		return "update"
	}
	if strings.Contains(lowerText, "investigating") || strings.Contains(lowerText, "looking into") || strings.Contains(lowerText, "debugging") {
		return "investigation"
	}
	if strings.Contains(lowerText, "blocked") || strings.Contains(lowerText, "waiting") || strings.Contains(lowerText, "stuck") {
		return "blocker"
	}
	
	return "progress"
}

// determineCommentCompletionStatusFromText determines completion status from comment (helper for enhanced processing)
func (e *EmbeddedLLM) determineCommentCompletionStatusFromText(text string) string {
	lowerText := strings.ToLower(text)
	
	if strings.Contains(lowerText, "completed") || strings.Contains(lowerText, "finished") || strings.Contains(lowerText, "done") || strings.Contains(lowerText, "resolved") {
		return "completed"
	}
	if strings.Contains(lowerText, "working on") || strings.Contains(lowerText, "in progress") || strings.Contains(lowerText, "currently") {
		return "in_progress"
	}
	if strings.Contains(lowerText, "blocked") || strings.Contains(lowerText, "waiting") || strings.Contains(lowerText, "stuck") {
		return "blocked"
	}
	if strings.Contains(lowerText, "planning") || strings.Contains(lowerText, "will") || strings.Contains(lowerText, "next") {
		return "planned"
	}
	
	return "unknown"
}

// extractKeyTopicsFromText extracts key topics from comment text (helper for enhanced processing)
func (e *EmbeddedLLM) extractKeyTopicsFromText(text string) []string {
	lowerText := strings.ToLower(text)
	var topics []string
	
	topicMap := map[string]string{
		"terraform":     "Terraform",
		"spacelift":     "Spacelift",
		"aws":          "AWS",
		"database":     "Database",
		"kubernetes":   "Kubernetes",
		"k8s":          "Kubernetes",
		"docker":       "Docker",
		"api":          "API",
		"security":     "Security",
		"authentication": "Authentication",
		"deployment":   "Deployment",
		"testing":      "Testing",
		"monitoring":   "Monitoring",
		"ci/cd":        "CI/CD",
		"pipeline":     "Pipeline",
	}
	
	for keyword, topic := range topicMap {
		if strings.Contains(lowerText, keyword) {
			topics = append(topics, topic)
		}
	}
	
	return e.removeDuplicateStrings(topics)
}

// ProcessIssuesWithComments processes issues with comments using the new enhanced pipeline
func (e *EmbeddedLLM) ProcessIssuesWithComments(issues []jira.Issue, comments []jira.Comment) (*ProcessedData, error) {
	if len(issues) == 0 {
		return nil, fmt.Errorf("no issues provided for processing")
	}
	
	// Create enhanced data processor with debug disabled for production
	processor := NewEnhancedDataProcessor(false)
	
	// Use the enhanced processing pipeline
	processedData, err := processor.ProcessIssuesWithComments(issues, comments)
	if err != nil {
		// Fallback handling when enhanced processing fails
		return e.createFallbackProcessedData(issues, comments, err)
	}
	
	// Enhance with technical pattern matching
	patternMatcher := NewTechnicalPatternMatcher(false)
	err = e.enhanceWithTechnicalPatterns(processedData, patternMatcher)
	if err != nil {
		// Log warning but continue with basic processed data
		// In a real implementation, this would use proper logging
		fmt.Printf("Warning: Failed to enhance with technical patterns: %v\n", err)
	}
	
	return processedData, nil
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
	
	// Process issues with comments using enhanced pipeline
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
		"database":     "Database",
		"kubernetes":   "Kubernetes",
		"k8s":          "Kubernetes",
		"docker":       "Docker",
		"api":          "API",
		"security":     "Security",
		"authentication": "Authentication",
		"deployment":   "Deployment",
		"testing":      "Testing",
		"monitoring":   "Monitoring",
		"ci/cd":        "CI/CD",
		"pipeline":     "Pipeline",
	}
	
	for keyword, topic := range topicMap {
		if strings.Contains(lowerText, keyword) {
			topics = append(topics, topic)
		}
	}
	
	return e.removeDuplicateStrings(topics)
}

// createIssuesWithCommentsMapping creates a proper mapping between issues and comments
func (e *EmbeddedLLM) createIssuesWithCommentsMapping(issues []jira.Issue, comments []jira.Comment) []jira.Issue {
	// For now, return issues as-is since the processor will handle comment association
	// In a real implementation, we would need proper issue-comment mapping from the API
	return issues
}

// generateFallbackSummary generates a fallback summary when enhanced processing fails
func (e *EmbeddedLLM) generateFallbackSummary(issues []jira.Issue, comments []jira.Comment, worklogs []jira.WorklogEntry) (string, error) {
	// Fall back to the original GenerateStandupSummary method
	return e.GenerateStandupSummary(issues, worklogs)
}

// generateEnhancedStandupSummary generates an enhanced summary using processed data
func (e *EmbeddedLLM) generateEnhancedStandupSummary(processedData *ProcessedData, patternMatcher *TechnicalPatternMatcher, worklogs []jira.WorklogEntry) (string, error) {
	if processedData == nil || len(processedData.Issues) == 0 {
		return "No recent activity to report", nil
	}
	
	// Extract key activities from processed data
	keyActivities := processedData.GetKeyActivities()
	
	// Get technical context summary
	var techSummary string
	if processedData.TechnicalContext != nil {
		techSummary = e.buildTechnicalContextSummary(processedData.TechnicalContext)
	}
	
	// Categorize work by completion status
	var completedWork []string
	var inProgressWork []string
	var blockedWork []string
	
	for _, issue := range processedData.Issues {
		switch strings.ToLower(issue.CompletionStatus) {
		case "completed", "done", "resolved":
			if issue.WorkSummary != "" {
				completedWork = append(completedWork, issue.WorkSummary)
			}
		case "in_progress", "working", "active":
			if issue.WorkSummary != "" {
				inProgressWork = append(inProgressWork, issue.WorkSummary)
			}
		case "blocked":
			if issue.WorkSummary != "" {
				blockedWork = append(blockedWork, issue.WorkSummary)
			}
		}
	}
	
	// Build enhanced summary
	return e.buildEnhancedSummary(completedWork, inProgressWork, blockedWork, keyActivities, techSummary)
}

// generateEnhancedCommentSummary generates an enhanced summary from processed comments
func (e *EmbeddedLLM) generateEnhancedCommentSummary(processedComments []ProcessedComment, patternMatcher *TechnicalPatternMatcher) (string, error) {
	if len(processedComments) == 0 {
		return "No comments to summarize", nil
	}
	
	// Categorize comments by activity type and completion status
	var completedActivities []string
	var inProgressActivities []string
	var technicalTopics []string
	
	for _, comment := range processedComments {
		// Extract activities based on completion status
		switch comment.CompletionStatus {
		case "completed":
			if len(comment.ExtractedActions) > 0 {
				activity := e.buildActivityFromComment(comment)
				if activity != "" {
					completedActivities = append(completedActivities, activity)
				}
			}
		case "in_progress":
			if len(comment.ExtractedActions) > 0 {
				activity := e.buildActivityFromComment(comment)
				if activity != "" {
					inProgressActivities = append(inProgressActivities, activity)
				}
			}
		}
		
		// Collect technical topics
		technicalTopics = append(technicalTopics, comment.KeyTopics...)
	}
	
	// Remove duplicates
	completedActivities = e.removeDuplicateStrings(completedActivities)
	inProgressActivities = e.removeDuplicateStrings(inProgressActivities)
	technicalTopics = e.removeDuplicateStrings(technicalTopics)
	
	// Build summary from enhanced data
	return e.buildEnhancedCommentSummary(completedActivities, inProgressActivities, technicalTopics)
}

// buildTechnicalContextSummary builds a summary from technical context
func (e *EmbeddedLLM) buildTechnicalContextSummary(context *TechnicalContext) string {
	var parts []string
	
	// Add deployment activities
	if len(context.Deployments) > 0 {
		deploymentCount := 0
		for _, deployment := range context.Deployments {
			if deployment.Status == "completed" {
				deploymentCount++
			}
		}
		if deploymentCount > 0 {
			parts = append(parts, fmt.Sprintf("%d deployment(s)", deploymentCount))
		}
	}
	
	// Add infrastructure work
	if len(context.Infrastructure) > 0 {
		infraCount := 0
		for _, infra := range context.Infrastructure {
			if infra.Status == "completed" {
				infraCount++
			}
		}
		if infraCount > 0 {
			parts = append(parts, fmt.Sprintf("%d infrastructure update(s)", infraCount))
		}
	}
	
	// Add key technologies
	if len(context.Technologies) > 0 {
		topTech := context.Technologies
		if len(topTech) > 3 {
			topTech = topTech[:3]
		}
		parts = append(parts, "using "+strings.Join(topTech, ", "))
	}
	
	if len(parts) == 0 {
		return ""
	}
	
	return strings.Join(parts, ", ")
}

// buildActivityFromComment builds an activity description from a processed comment
func (e *EmbeddedLLM) buildActivityFromComment(comment ProcessedComment) string {
	if len(comment.ExtractedActions) == 0 {
		return ""
	}
	
	action := comment.ExtractedActions[0] // Take the first action
	
	// Add technical context if available
	if len(comment.TechnicalTerms) > 0 {
		tech := comment.TechnicalTerms[0] // Take the first technical term
		return fmt.Sprintf("%s %s work", strings.Title(action), tech)
	}
	
	// Add work type context
	if comment.WorkType != "general" {
		return fmt.Sprintf("%s %s", strings.Title(action), strings.ReplaceAll(comment.WorkType, "_", " "))
	}
	
	return strings.Title(action) + " development work"
}

// buildEnhancedSummary builds the final enhanced summary
func (e *EmbeddedLLM) buildEnhancedSummary(completed, inProgress, blocked, keyActivities []string, techSummary string) (string, error) {
	var parts []string
	
	// Add completed work (highest priority)
	if len(completed) > 0 {
		if len(completed) == 1 {
			parts = append(parts, "Completed: "+completed[0])
		} else if len(completed) <= 3 {
			parts = append(parts, "Completed: "+strings.Join(completed, ", "))
		} else {
			parts = append(parts, fmt.Sprintf("Completed: %s and %d other items", strings.Join(completed[:2], ", "), len(completed)-2))
		}
	}
	
	// Add in-progress work
	if len(inProgress) > 0 && len(parts) < 2 {
		if len(inProgress) == 1 {
			parts = append(parts, "Working on: "+inProgress[0])
		} else {
			parts = append(parts, "Working on: "+strings.Join(inProgress[:min(2, len(inProgress))], ", "))
		}
	}
	
	// Add blocked work (important to highlight)
	if len(blocked) > 0 && len(parts) < 3 {
		parts = append(parts, "Blocked: "+strings.Join(blocked[:min(2, len(blocked))], ", "))
	}
	
	// Add technical context if we don't have enough content
	if len(parts) < 2 && techSummary != "" {
		parts = append(parts, "Technical work: "+techSummary)
	}
	
	// Add key activities as fallback
	if len(parts) == 0 && len(keyActivities) > 0 {
		if len(keyActivities) <= 3 {
			parts = append(parts, "Activities: "+strings.Join(keyActivities, ", "))
		} else {
			parts = append(parts, "Activities: "+strings.Join(keyActivities[:3], ", ")+" and more")
		}
	}
	
	// Final fallback
	if len(parts) == 0 {
		return "Multiple development activities completed", nil
	}
	
	return strings.Join(parts, "; "), nil
}

// buildEnhancedCommentSummary builds a summary from enhanced comment data
func (e *EmbeddedLLM) buildEnhancedCommentSummary(completed, inProgress, topics []string) (string, error) {
	var parts []string
	
	// Prioritize completed activities
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
			parts = append(parts, inProgress[0]+" (in progress)")
		} else {
			parts = append(parts, strings.Join(inProgress[:1], ", ")+" (in progress)")
		}
	}
	
	// Add technical topics if we don't have enough content
	if len(parts) == 0 && len(topics) > 0 {
		if len(topics) <= 3 {
			return "Technical work on " + strings.Join(topics, ", "), nil
		} else {
			return "Technical work on " + strings.Join(topics[:3], ", ") + " and more", nil
		}
	}
	
	if len(parts) == 0 {
		return "Development activities", nil
	}
	
	return strings.Join(parts, "; "), nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// generateEnhancedCommentSummary creates a summary using processed comments and pattern matching
func (e *EmbeddedLLM) generateEnhancedCommentSummary(processedComments []ProcessedComment, patternMatcher *TechnicalPatternMatcher) (string, error) {
	if len(processedComments) == 0 {
		return "No comments to summarize", nil
	}
	
	if len(processedComments) == 1 {
		return e.generateSingleProcessedCommentSummary(processedComments[0], patternMatcher)
	}
	
	// Categorize activities by completion status and importance
	var completedWork []string
	var inProgressWork []string
	var generalActivities []string
	var technicalTopics []string
	
	for _, comment := range processedComments {
		// Use pattern matcher for enhanced insights
		if comment.Original.Body.Text != "" {
			patterns, err := patternMatcher.MatchAllPatterns(comment.Original.Body.Text)
			if err == nil {
				e.extractInsightsFromPatterns(patterns, &technicalTopics, &generalActivities)
			}
		}
		
		// Categorize based on completion status
		switch comment.CompletionStatus {
		case "completed":
			if len(comment.ExtractedActions) > 0 {
				activity := e.buildActivityFromComment(comment)
				if activity != "" {
					completedWork = append(completedWork, activity)
				}
			}
		case "in_progress":
			if len(comment.ExtractedActions) > 0 {
				activity := e.buildActivityFromComment(comment)
				if activity != "" {
					inProgressWork = append(inProgressWork, activity)
				}
			}
		default:
			if len(comment.ExtractedActions) > 0 {
				activity := e.buildActivityFromComment(comment)
				if activity != "" {
					generalActivities = append(generalActivities, activity)
				}
			}
		}
		
		// Add technical topics
		technicalTopics = append(technicalTopics, comment.TechnicalTerms...)
		technicalTopics = append(technicalTopics, comment.KeyTopics...)
	}
	
	// Remove duplicates and build final summary
	completedWork = e.removeDuplicateStrings(completedWork)
	inProgressWork = e.removeDuplicateStrings(inProgressWork)
	generalActivities = e.removeDuplicateStrings(generalActivities)
	technicalTopics = e.removeDuplicateStrings(technicalTopics)
	
	return e.buildEnhancedCommentSummary(completedWork, inProgressWork, generalActivities, technicalTopics), nil
}

// generateSingleProcessedCommentSummary creates a summary for a single processed comment
func (e *EmbeddedLLM) generateSingleProcessedCommentSummary(comment ProcessedComment, patternMatcher *TechnicalPatternMatcher) (string, error) {
	// Use pattern matching for enhanced insights
	if comment.Original.Body.Text != "" {
		patterns, err := patternMatcher.MatchAllPatterns(comment.Original.Body.Text)
		if err == nil {
			// Extract high-confidence patterns
			if infraPatterns, ok := patterns["infrastructure"].([]InfrastructurePattern); ok {
				for _, pattern := range infraPatterns {
					if pattern.Confidence > 0.8 {
						return fmt.Sprintf("%s %s %s (%s)", 
							strings.Title(pattern.Action), 
							pattern.Type, 
							pattern.Component, 
							pattern.Status), nil
					}
				}
			}
			
			if deployPatterns, ok := patterns["deployment"].([]DeploymentPattern); ok {
				for _, pattern := range deployPatterns {
					if pattern.Confidence > 0.8 {
						return fmt.Sprintf("%s to %s (%s)", 
							strings.Title(pattern.Type), 
							pattern.Environment, 
							pattern.Status), nil
					}
				}
			}
			
			if devPatterns, ok := patterns["development"].([]DevelopmentPattern); ok {
				for _, pattern := range devPatterns {
					if pattern.Confidence > 0.8 {
						return fmt.Sprintf("%s %s (%s)", 
							strings.Title(pattern.Action), 
							pattern.Type, 
							pattern.Status), nil
					}
				}
			}
		}
	}
	
	// Fallback to processed comment data
	if len(comment.ExtractedActions) > 0 && len(comment.TechnicalTerms) > 0 {
		return fmt.Sprintf("%s %s work", 
			strings.Title(comment.ExtractedActions[0]), 
			comment.TechnicalTerms[0]), nil
	}
	
	if len(comment.ExtractedActions) > 0 {
		return fmt.Sprintf("%s work", strings.Title(comment.ExtractedActions[0])), nil
	}
	
	// Final fallback to original intelligent summary
	return e.createIntelligentSummary(comment.Original.Body.Text), nil, nil
}

// buildActivityFromComment builds an activity description from a processed comment
func (e *EmbeddedLLM) buildActivityFromComment(comment ProcessedComment) string {
	if len(comment.ExtractedActions) == 0 {
		return ""
	}
	
	action := comment.ExtractedActions[0]
	
	// Add technical context if available
	if len(comment.TechnicalTerms) > 0 {
		return fmt.Sprintf("%s %s", strings.Title(action), comment.TechnicalTerms[0])
	}
	
	// Add work type context
	if comment.WorkType != "general" && comment.WorkType != "" {
		return fmt.Sprintf("%s %s work", strings.Title(action), comment.WorkType)
	}
	
	// Add key topics if available
	if len(comment.KeyTopics) > 0 {
		return fmt.Sprintf("%s %s", strings.Title(action), comment.KeyTopics[0])
	}
	
	return strings.Title(action) + " work"
}

// buildEnhancedCommentSummary builds the final enhanced comment summary
func (e *EmbeddedLLM) buildEnhancedCommentSummary(completed, inProgress, general, topics []string) string {
	var summaryParts []string
	
	// Prioritize completed work
	if len(completed) > 0 {
		if len(completed) == 1 {
			summaryParts = append(summaryParts, "Completed: "+completed[0])
		} else if len(completed) <= 3 {
			summaryParts = append(summaryParts, "Completed: "+strings.Join(completed, ", "))
		} else {
			summaryParts = append(summaryParts, fmt.Sprintf("Completed: %s and %d other items", 
				strings.Join(completed[:2], ", "), len(completed)-2))
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
	
	// Add general activities if needed
	if len(summaryParts) < 2 && len(general) > 0 {
		filteredGeneral := e.filterUniqueActivities(general, append(completed, inProgress...))
		if len(filteredGeneral) > 0 {
			if len(filteredGeneral) == 1 {
				summaryParts = append(summaryParts, "Also: "+filteredGeneral[0])
			} else {
				summaryParts = append(summaryParts, "Also: "+strings.Join(filteredGeneral[:min(2, len(filteredGeneral))], ", "))
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
		return "Various development activities"
	}
	
	return strings.Join(summaryParts, ". ")
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
			activity := fmt.Sprintf("%s %s (%s)", strings.Title(deployment.Type), deployment.Component, deployment.Status)
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

// createIssuesWithCommentsMapping creates a proper mapping between issues and comments
func (e *EmbeddedLLM) createIssuesWithCommentsMapping(issues []jira.Issue, comments []jira.Comment) []jira.Issue {
	// For now, return issues as-is since the processor will handle comment association
	// In a real implementation, we would need proper issue-comment mapping from the API
	return issues
}

// generateFallbackSummary generates a fallback summary when enhanced processing fails
func (e *EmbeddedLLM) generateFallbackSummary(issues []jira.Issue, comments []jira.Comment, worklogs []jira.WorklogEntry) (string, error) {
	// Fall back to the original GenerateStandupSummary method
	return e.GenerateStandupSummary(issues, worklogs)
}

// generateEnhancedStandupSummary generates an enhanced summary using processed data
func (e *EmbeddedLLM) generateEnhancedStandupSummary(processedData *ProcessedData, patternMatcher *TechnicalPatternMatcher, worklogs []jira.WorklogEntry) (string, error) {
	if processedData == nil || len(processedData.Issues) == 0 {
		return "No recent activity to report", nil
	}
	
	// Extract key activities from processed data
	keyActivities := processedData.GetKeyActivities()
	
	// Get technical context summary
	var techSummary string
	if processedData.TechnicalContext != nil {
		techSummary = e.buildTechnicalContextSummary(processedData.TechnicalContext)
	}
	
	// Categorize work by completion status
	var completedWork []string
	var inProgressWork []string
	var blockedWork []string
	
	for _, issue := range processedData.Issues {
		switch strings.ToLower(issue.CompletionStatus) {
		case "completed", "done", "resolved":
			if issue.WorkSummary != "" {
				completedWork = append(completedWork, issue.WorkSummary)
			}
		case "in_progress", "working", "active":
			if issue.WorkSummary != "" {
				inProgressWork = append(inProgressWork, issue.WorkSummary)
			}
		case "blocked":
			if issue.WorkSummary != "" {
				blockedWork = append(blockedWork, issue.WorkSummary)
			}
		}
	}
	
	// Build enhanced summary
	return e.buildEnhancedSummary(completedWork, inProgressWork, blockedWork, keyActivities, techSummary)
}

// generateEnhancedCommentSummary generates an enhanced summary from processed comments
func (e *EmbeddedLLM) generateEnhancedCommentSummary(processedComments []ProcessedComment, patternMatcher *TechnicalPatternMatcher) (string, error) {
	if len(processedComments) == 0 {
		return "No comments to summarize", nil
	}
	
	// Categorize comments by activity type and completion status
	var completedActivities []string
	var inProgressActivities []string
	var technicalTopics []string
	
	for _, comment := range processedComments {
		// Extract activities based on completion status
		switch comment.CompletionStatus {
		case "completed":
			if len(comment.ExtractedActions) > 0 {
				activity := e.buildActivityFromComment(comment)
				if activity != "" {
					completedActivities = append(completedActivities, activity)
				}
			}
		case "in_progress":
			if len(comment.ExtractedActions) > 0 {
				activity := e.buildActivityFromComment(comment)
				if activity != "" {
					inProgressActivities = append(inProgressActivities, activity)
				}
			}
		}
		
		// Collect technical topics
		technicalTopics = append(technicalTopics, comment.KeyTopics...)
	}
	
	// Remove duplicates
	completedActivities = e.removeDuplicateStrings(completedActivities)
	inProgressActivities = e.removeDuplicateStrings(inProgressActivities)
	technicalTopics = e.removeDuplicateStrings(technicalTopics)
	
	// Build summary from enhanced data
	return e.buildEnhancedCommentSummary(completedActivities, inProgressActivities, technicalTopics)
}

// buildTechnicalContextSummary builds a summary from technical context
func (e *EmbeddedLLM) buildTechnicalContextSummary(context *TechnicalContext) string {
	var parts []string
	
	// Add deployment activities
	if len(context.Deployments) > 0 {
		deploymentCount := 0
		for _, deployment := range context.Deployments {
			if deployment.Status == "completed" {
				deploymentCount++
			}
		}
		if deploymentCount > 0 {
			parts = append(parts, fmt.Sprintf("%d deployment(s)", deploymentCount))
		}
	}
	
	// Add infrastructure work
	if len(context.Infrastructure) > 0 {
		infraCount := 0
		for _, infra := range context.Infrastructure {
			if infra.Status == "completed" {
				infraCount++
			}
		}
		if infraCount > 0 {
			parts = append(parts, fmt.Sprintf("%d infrastructure update(s)", infraCount))
		}
	}
	
	// Add key technologies
	if len(context.Technologies) > 0 {
		topTech := context.Technologies
		if len(topTech) > 3 {
			topTech = topTech[:3]
		}
		parts = append(parts, "using "+strings.Join(topTech, ", "))
	}
	
	if len(parts) == 0 {
		return ""
	}
	
	return strings.Join(parts, ", ")
}

// buildActivityFromComment builds an activity description from a processed comment
func (e *EmbeddedLLM) buildActivityFromComment(comment ProcessedComment) string {
	if len(comment.ExtractedActions) == 0 {
		return ""
	}
	
	action := comment.ExtractedActions[0] // Take the first action
	
	// Add technical context if available
	if len(comment.TechnicalTerms) > 0 {
		tech := comment.TechnicalTerms[0] // Take the first technical term
		return fmt.Sprintf("%s %s work", strings.Title(action), tech)
	}
	
	// Add work type context
	if comment.WorkType != "general" {
		return fmt.Sprintf("%s %s", strings.Title(action), strings.ReplaceAll(comment.WorkType, "_", " "))
	}
	
	return strings.Title(action) + " development work"
}

// buildEnhancedSummary builds the final enhanced summary
func (e *EmbeddedLLM) buildEnhancedSummary(completed, inProgress, blocked, keyActivities []string, techSummary string) (string, error) {
	var parts []string
	
	// Add completed work (highest priority)
	if len(completed) > 0 {
		if len(completed) == 1 {
			parts = append(parts, "Completed: "+completed[0])
		} else if len(completed) <= 3 {
			parts = append(parts, "Completed: "+strings.Join(completed, ", "))
		} else {
			parts = append(parts, fmt.Sprintf("Completed: %s and %d other items", strings.Join(completed[:2], ", "), len(completed)-2))
		}
	}
	
	// Add in-progress work
	if len(inProgress) > 0 && len(parts) < 2 {
		if len(inProgress) == 1 {
			parts = append(parts, "Working on: "+inProgress[0])
		} else {
			parts = append(parts, "Working on: "+strings.Join(inProgress[:min(2, len(inProgress))], ", "))
		}
	}
	
	// Add blocked work (important to highlight)
	if len(blocked) > 0 && len(parts) < 3 {
		parts = append(parts, "Blocked: "+strings.Join(blocked[:min(2, len(blocked))], ", "))
	}
	
	// Add technical context if we don't have enough content
	if len(parts) < 2 && techSummary != "" {
		parts = append(parts, "Technical work: "+techSummary)
	}
	
	// Add key activities as fallback
	if len(parts) == 0 && len(keyActivities) > 0 {
		if len(keyActivities) <= 3 {
			parts = append(parts, "Activities: "+strings.Join(keyActivities, ", "))
		} else {
			parts = append(parts, "Activities: "+strings.Join(keyActivities[:3], ", ")+" and more")
		}
	}
	
	// Final fallback
	if len(parts) == 0 {
		return "Multiple development activities completed", nil
	}
	
	return strings.Join(parts, "; "), nil
}

// buildEnhancedCommentSummary builds a summary from enhanced comment data
func (e *EmbeddedLLM) buildEnhancedCommentSummary(completed, inProgress, topics []string) (string, error) {
	var parts []string
	
	// Prioritize completed activities
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
			parts = append(parts, inProgress[0]+" (in progress)")
		} else {
			parts = append(parts, strings.Join(inProgress[:1], ", ")+" (in progress)")
		}
	}
	
	// Add technical topics if we don't have enough content
	if len(parts) == 0 && len(topics) > 0 {
		if len(topics) <= 3 {
			return "Technical work on " + strings.Join(topics, ", "), nil
		} else {
			return "Technical work on " + strings.Join(topics[:3], ", ") + " and more", nil
		}
	}
	
	if len(parts) == 0 {
		return "Development activities", nil
	}
	
	return strings.Join(parts, "; "), nil
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// createFallbackProcessedData creates basic processed data when enhanced processing fails
func (e *EmbeddedLLM) createFallbackProcessedData(issues []jira.Issue, comments []jira.Comment, originalError error) (*ProcessedData, error) {
	// Create basic processed data structure
	processedData := NewProcessedData()
	
	// Group comments by issue (basic implementation)
	commentsByIssue := e.groupCommentsByIssueKey(comments, issues)
	
	// Process each issue with basic logic
	for _, issue := range issues {
		issueComments := commentsByIssue[issue.Key]
		
		// Create basic enhanced issue
		enhancedIssue := EnhancedIssue{
			Issue:             issue,
			Comments:          issueComments,
			ProcessedComments: make([]ProcessedComment, 0),
			TechnicalContext:  &TechnicalContext{},
			Priority:          e.calculateBasicPriority(issue),
			WorkType:          e.determineBasicWorkType(issue),
			CompletionStatus:  e.determineBasicCompletionStatus(issue),
			KeyActivities:     e.extractBasicKeyActivities(issue, issueComments),
		}
		
		// Process comments with fallback logic
		for _, comment := range issueComments {
			processedComment, err := e.processCommentWithFallback(comment)
			if err == nil {
				enhancedIssue.ProcessedComments = append(enhancedIssue.ProcessedComments, processedComment)
			}
		}
		
		// Generate basic work summary
		enhancedIssue.WorkSummary = e.generateBasicWorkSummary(enhancedIssue)
		
		// Add to processed data
		if err := processedData.AddIssue(enhancedIssue); err != nil {
			continue // Skip issues that can't be added
		}
		
		// Extract basic technical context
		e.extractBasicTechnicalContext(processedData.TechnicalContext, enhancedIssue)
	}
	
	return processedData, nil
}

// groupCommentsByIssueKey groups comments by issue key using basic matching
func (e *EmbeddedLLM) groupCommentsByIssueKey(comments []jira.Comment, issues []jira.Issue) map[string][]jira.Comment {
	commentsByIssue := make(map[string][]jira.Comment)
	
	// Initialize map with issue keys
	for _, issue := range issues {
		commentsByIssue[issue.Key] = make([]jira.Comment, 0)
	}
	
	// For now, we'll assume comments need to be associated externally
	// In a real implementation, this would use proper issue-comment relationships
	// This is a placeholder that distributes comments evenly for demonstration
	if len(issues) > 0 && len(comments) > 0 {
		commentsPerIssue := len(comments) / len(issues)
		remainder := len(comments) % len(issues)
		
		commentIndex := 0
		for i, issue := range issues {
			count := commentsPerIssue
			if i < remainder {
				count++
			}
			
			for j := 0; j < count && commentIndex < len(comments); j++ {
				commentsByIssue[issue.Key] = append(commentsByIssue[issue.Key], comments[commentIndex])
				commentIndex++
			}
		}
	}
	
	return commentsByIssue
}

// calculateBasicPriority calculates basic priority for fallback processing
func (e *EmbeddedLLM) calculateBasicPriority(issue jira.Issue) int {
	priority := strings.ToLower(issue.Fields.Priority.Name)
	
	switch priority {
	case "critical", "highest":
		return 100
	case "high":
		return 80
	case "medium":
		return 60
	case "low":
		return 40
	case "lowest":
		return 20
	default:
		return 50
	}
}

// determineBasicWorkType determines basic work type for fallback processing
func (e *EmbeddedLLM) determineBasicWorkType(issue jira.Issue) string {
	issueType := strings.ToLower(issue.Fields.IssueType.Name)
	summary := strings.ToLower(issue.Fields.Summary)
	
	if strings.Contains(issueType, "bug") || strings.Contains(summary, "fix") {
		return "bug_fix"
	}
	if strings.Contains(summary, "deploy") {
		return "deployment"
	}
	if strings.Contains(summary, "terraform") || strings.Contains(summary, "aws") {
		return "infrastructure"
	}
	if strings.Contains(summary, "database") {
		return "database"
	}
	if strings.Contains(issueType, "feature") || strings.Contains(issueType, "story") {
		return "feature_development"
	}
	
	return "general"
}

// determineBasicCompletionStatus determines basic completion status for fallback processing
func (e *EmbeddedLLM) determineBasicCompletionStatus(issue jira.Issue) string {
	status := strings.ToLower(issue.Fields.Status.Name)
	
	if strings.Contains(status, "done") || strings.Contains(status, "closed") {
		return "completed"
	}
	if strings.Contains(status, "progress") || strings.Contains(status, "development") {
		return "in_progress"
	}
	if strings.Contains(status, "blocked") {
		return "blocked"
	}
	
	return "planned"
}

// extractBasicKeyActivities extracts basic key activities for fallback processing
func (e *EmbeddedLLM) extractBasicKeyActivities(issue jira.Issue, comments []jira.Comment) []string {
	var activities []string
	
	// Extract from issue summary
	summary := strings.ToLower(issue.Fields.Summary)
	if strings.Contains(summary, "implement") {
		activities = append(activities, "Implementation work")
	}
	if strings.Contains(summary, "fix") {
		activities = append(activities, "Bug fixing")
	}
	if strings.Contains(summary, "deploy") {
		activities = append(activities, "Deployment work")
	}
	
	// Extract from comments
	for _, comment := range comments {
		text := strings.ToLower(comment.Body.Text)
		if strings.Contains(text, "completed") {
			activities = append(activities, "Completed tasks")
			break
		}
		if strings.Contains(text, "working on") {
			activities = append(activities, "Active development")
			break
		}
	}
	
	if len(activities) == 0 {
		activities = append(activities, "General development work")
	}
	
	return e.removeDuplicateStrings(activities)
}

// processCommentWithFallback processes a comment with basic fallback logic
func (e *EmbeddedLLM) processCommentWithFallback(comment jira.Comment) (ProcessedComment, error) {
	if comment.ID == "" {
		return ProcessedComment{}, fmt.Errorf("comment ID is empty")
	}
	
	text := comment.Body.Text
	
	// Basic processing without enhanced features
	processedComment := ProcessedComment{
		Original:         comment,
		ExtractedActions: e.extractBasicActions(text),
		TechnicalTerms:   e.extractBasicTechnicalTerms(text),
		WorkType:         e.determineBasicCommentWorkType(text),
		Sentiment:        e.determineBasicSentiment(text),
		Importance:       e.calculateBasicImportance(text),
		ActivityType:     e.determineBasicActivityType(text),
		CompletionStatus: e.determineBasicCompletionStatusFromComment(text),
		KeyTopics:        e.extractBasicTopics(text),
	}
	
	return processedComment, nil
}

// extractBasicActions extracts basic actions for fallback processing
func (e *EmbeddedLLM) extractBasicActions(text string) []string {
	lowerText := strings.ToLower(text)
	var actions []string
	
	basicActions := []string{"implemented", "fixed", "updated", "deployed", "tested", "reviewed", "created"}
	for _, action := range basicActions {
		if strings.Contains(lowerText, action) {
			actions = append(actions, action)
		}
	}
	
	return actions
}

// extractBasicTechnicalTerms extracts basic technical terms for fallback processing
func (e *EmbeddedLLM) extractBasicTechnicalTerms(text string) []string {
	lowerText := strings.ToLower(text)
	var terms []string
	
	basicTerms := []string{"terraform", "aws", "database", "api", "docker", "kubernetes"}
	for _, term := range basicTerms {
		if strings.Contains(lowerText, term) {
			terms = append(terms, term)
		}
	}
	
	return terms
}

// determineBasicCommentWorkType determines basic work type from comment for fallback processing
func (e *EmbeddedLLM) determineBasicCommentWorkType(text string) string {
	lowerText := strings.ToLower(text)
	
	if strings.Contains(lowerText, "terraform") || strings.Contains(lowerText, "aws") {
		return "infrastructure"
	}
	if strings.Contains(lowerText, "database") {
		return "database"
	}
	if strings.Contains(lowerText, "deploy") {
		return "deployment"
	}
	if strings.Contains(lowerText, "test") {
		return "testing"
	}
	
	return "general"
}

// determineBasicSentiment determines basic sentiment for fallback processing
func (e *EmbeddedLLM) determineBasicSentiment(text string) string {
	lowerText := strings.ToLower(text)
	
	if strings.Contains(lowerText, "completed") || strings.Contains(lowerText, "fixed") {
		return "positive"
	}
	if strings.Contains(lowerText, "blocked") || strings.Contains(lowerText, "failed") {
		return "negative"
	}
	
	return "neutral"
}

// calculateBasicImportance calculates basic importance for fallback processing
func (e *EmbeddedLLM) calculateBasicImportance(text string) int {
	lowerText := strings.ToLower(text)
	importance := 50
	
	if strings.Contains(lowerText, "critical") || strings.Contains(lowerText, "urgent") {
		importance += 30
	}
	if strings.Contains(lowerText, "completed") || strings.Contains(lowerText, "deployed") {
		importance += 20
	}
	
	return importance
}

// determineBasicActivityType determines basic activity type for fallback processing
func (e *EmbeddedLLM) determineBasicActivityType(text string) string {
	lowerText := strings.ToLower(text)
	
	if strings.Contains(lowerText, "completed") {
		return "completion"
	}
	if strings.Contains(lowerText, "working on") {
		return "initiation"
	}
	if strings.Contains(lowerText, "updated") {
		return "update"
	}
	
	return "progress"
}

// determineBasicCompletionStatusFromComment determines basic completion status from comment for fallback processing
func (e *EmbeddedLLM) determineBasicCompletionStatusFromComment(text string) string {
	lowerText := strings.ToLower(text)
	
	if strings.Contains(lowerText, "completed") || strings.Contains(lowerText, "done") {
		return "completed"
	}
	if strings.Contains(lowerText, "working on") {
		return "in_progress"
	}
	if strings.Contains(lowerText, "blocked") {
		return "blocked"
	}
	
	return "unknown"
}

// extractBasicTopics extracts basic topics for fallback processing
func (e *EmbeddedLLM) extractBasicTopics(text string) []string {
	lowerText := strings.ToLower(text)
	var topics []string
	
	basicTopics := map[string]string{
		"terraform": "Terraform",
		"aws":       "AWS",
		"database":  "Database",
		"api":       "API",
		"docker":    "Docker",
	}
	
	for keyword, topic := range basicTopics {
		if strings.Contains(lowerText, keyword) {
			topics = append(topics, topic)
		}
	}
	
	return topics
}

// generateBasicWorkSummary generates basic work summary for fallback processing
func (e *EmbeddedLLM) generateBasicWorkSummary(issue EnhancedIssue) string {
	if len(issue.ProcessedComments) == 0 {
		return fmt.Sprintf("%s: %s", issue.CompletionStatus, issue.Issue.Fields.Summary)
	}
	
	// Simple summary based on completion status and work type
	return fmt.Sprintf("%s %s work", strings.Title(issue.CompletionStatus), issue.WorkType)
}

// extractBasicTechnicalContext extracts basic technical context for fallback processing
func (e *EmbeddedLLM) extractBasicTechnicalContext(context *TechnicalContext, issue EnhancedIssue) {
	// Extract basic technologies from processed comments
	for _, comment := range issue.ProcessedComments {
		context.Technologies = append(context.Technologies, comment.TechnicalTerms...)
		context.Actions = append(context.Actions, comment.ExtractedActions...)
	}
	
	// Remove duplicates
	context.Technologies = e.removeDuplicateStrings(context.Technologies)
	context.Actions = e.removeDuplicateStrings(context.Actions)
}

// enhanceWithTechnicalPatterns enhances processed data with technical pattern matching
func (e *EmbeddedLLM) enhanceWithTechnicalPatterns(processedData *ProcessedData, patternMatcher *TechnicalPatternMatcher) error {
	if processedData == nil || patternMatcher == nil {
		return fmt.Errorf("processed data or pattern matcher is nil")
	}
	
	// Enhance each issue with technical pattern insights
	for i := range processedData.Issues {
		issue := &processedData.Issues[i]
		
		// Analyze issue summary and description
		issueText := issue.Issue.Fields.Summary + " " + issue.Issue.Fields.Description.Text
		if issueText != "" {
			patterns, err := patternMatcher.MatchAllPatterns(issueText)
			if err == nil {
				e.applyPatternInsights(issue, patterns)
			}
		}
		
		// Analyze comments
		for j := range issue.ProcessedComments {
			comment := &issue.ProcessedComments[j]
			if comment.Original.Body.Text != "" {
				patterns, err := patternMatcher.MatchAllPatterns(comment.Original.Body.Text)
				if err == nil {
					e.applyCommentPatternInsights(comment, patterns)
				}
			}
		}
	}
	
	return nil
}

// applyPatternInsights applies pattern matching insights to an enhanced issue
func (e *EmbeddedLLM) applyPatternInsights(issue *EnhancedIssue, patterns map[string]interface{}) {
	// Extract infrastructure patterns
	if infraPatterns, ok := patterns["infrastructure"].([]InfrastructurePattern); ok {
		for _, pattern := range infraPatterns {
			if pattern.Confidence > 0.7 {
				// Add to technical context
				if issue.TechnicalContext == nil {
					issue.TechnicalContext = &TechnicalContext{}
				}
				
				// Add infrastructure work
				infraWork := InfrastructureWork{
					Type:        pattern.Type,
					Action:      pattern.Action,
					Component:   pattern.Component,
					Status:      pattern.Status,
					Description: fmt.Sprintf("%s %s %s", pattern.Action, pattern.Type, pattern.Component),
					Timestamp:   pattern.Timestamp,
				}
				issue.TechnicalContext.Infrastructure = append(issue.TechnicalContext.Infrastructure, infraWork)
				
				// Update key activities
				activity := fmt.Sprintf("%s %s", strings.Title(pattern.Action), pattern.Type)
				issue.KeyActivities = append(issue.KeyActivities, activity)
			}
		}
	}
	
	// Extract deployment patterns
	if deployPatterns, ok := patterns["deployment"].([]DeploymentPattern); ok {
		for _, pattern := range deployPatterns {
			if pattern.Confidence > 0.7 {
				if issue.TechnicalContext == nil {
					issue.TechnicalContext = &TechnicalContext{}
				}
				
				// Add deployment activity
				deployActivity := DeploymentActivity{
					Type:        pattern.Type,
					Environment: pattern.Environment,
					Status:      pattern.Status,
					Component:   pattern.Component,
					Description: fmt.Sprintf("%s to %s", pattern.Type, pattern.Environment),
					Timestamp:   pattern.Timestamp,
				}
				issue.TechnicalContext.Deployments = append(issue.TechnicalContext.Deployments, deployActivity)
				
				// Update key activities
				activity := fmt.Sprintf("%s to %s", strings.Title(pattern.Type), pattern.Environment)
				issue.KeyActivities = append(issue.KeyActivities, activity)
			}
		}
	}
	
	// Remove duplicate activities
	issue.KeyActivities = e.removeDuplicateStrings(issue.KeyActivities)
}

// applyCommentPatternInsights applies pattern matching insights to a processed comment
func (e *EmbeddedLLM) applyCommentPatternInsights(comment *ProcessedComment, patterns map[string]interface{}) {
	// Extract additional technical terms from patterns
	if infraPatterns, ok := patterns["infrastructure"].([]InfrastructurePattern); ok {
		for _, pattern := range infraPatterns {
			if pattern.Confidence > 0.7 {
				comment.TechnicalTerms = append(comment.TechnicalTerms, pattern.Type)
				if pattern.Action != "unknown" {
					comment.ExtractedActions = append(comment.ExtractedActions, pattern.Action)
				}
			}
		}
	}
	
	if deployPatterns, ok := patterns["deployment"].([]DeploymentPattern); ok {
		for _, pattern := range deployPatterns {
			if pattern.Confidence > 0.7 {
				comment.TechnicalTerms = append(comment.TechnicalTerms, "deployment")
				comment.KeyTopics = append(comment.KeyTopics, "Deployment")
				if pattern.Environment != "unknown" {
					comment.KeyTopics = append(comment.KeyTopics, strings.Title(pattern.Environment))
				}
			}
		}
	}
	
	if devPatterns, ok := patterns["development"].([]DevelopmentPattern); ok {
		for _, pattern := range devPatterns {
			if pattern.Confidence > 0.7 {
				comment.TechnicalTerms = append(comment.TechnicalTerms, pattern.Type)
				if pattern.Action != "unknown" {
					comment.ExtractedActions = append(comment.ExtractedActions, pattern.Action)
				}
			}
		}
	}
	
	// Remove duplicates
	comment.TechnicalTerms = e.removeDuplicateStrings(comment.TechnicalTerms)
	comment.ExtractedActions = e.removeDuplicateStrings(comment.ExtractedActions)
	comment.KeyTopics = e.removeDuplicateStrings(comment.KeyTopics)
}