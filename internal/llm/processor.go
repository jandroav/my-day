package llm

import (
	"fmt"
	"strings"
	"time"
	"my-day/internal/jira"
)

// EnhancedDataProcessor handles transformation from raw Jira data to processed format
type EnhancedDataProcessor struct {
	debug bool
}

// NewEnhancedDataProcessor creates a new data processor
func NewEnhancedDataProcessor(debug bool) *EnhancedDataProcessor {
	return &EnhancedDataProcessor{
		debug: debug,
	}
}

// ProcessIssuesWithComments transforms raw Jira issues and comments into ProcessedData
func (p *EnhancedDataProcessor) ProcessIssuesWithComments(issues []jira.Issue, comments []jira.Comment) (*ProcessedData, error) {
	processedData := NewProcessedData()
	
	// Group comments by issue
	commentsByIssue := p.groupCommentsByIssue(comments)
	
	// Process each issue
	for _, issue := range issues {
		enhancedIssue, err := p.processIssue(issue, commentsByIssue[issue.Key])
		if err != nil {
			if p.debug {
				fmt.Printf("Warning: Failed to process issue %s: %v\n", issue.Key, err)
			}
			// Continue processing other issues even if one fails
			continue
		}
		
		if err := processedData.AddIssue(enhancedIssue); err != nil {
			if p.debug {
				fmt.Printf("Warning: Failed to add issue %s: %v\n", issue.Key, err)
			}
			continue
		}
		
		// Extract technical context from this issue
		p.extractTechnicalContextFromIssue(processedData.TechnicalContext, enhancedIssue)
		
		// Create timeline events
		p.createTimelineEvents(processedData, enhancedIssue)
	}
	
	// Validate the processed data
	if validationErrors := processedData.Validate(); len(validationErrors) > 0 {
		if p.debug {
			fmt.Printf("Validation warnings: %+v\n", validationErrors)
		}
		// Continue despite validation warnings for now
	}
	
	return processedData, nil
}

// processIssue converts a single Jira issue to an EnhancedIssue
func (p *EnhancedDataProcessor) processIssue(issue jira.Issue, comments []jira.Comment) (EnhancedIssue, error) {
	if issue.Key == "" {
		return EnhancedIssue{}, fmt.Errorf("issue key is empty")
	}
	
	enhancedIssue := EnhancedIssue{
		Issue:             issue,
		Comments:          comments,
		ProcessedComments: make([]ProcessedComment, 0),
		TechnicalContext:  &TechnicalContext{},
		Priority:          p.calculateIssuePriority(issue),
		WorkType:          p.determineWorkType(issue),
		CompletionStatus:  p.determineCompletionStatus(issue),
		KeyActivities:     make([]string, 0),
	}
	
	// Process comments
	for _, comment := range comments {
		processedComment, err := p.processComment(comment)
		if err != nil {
			if p.debug {
				fmt.Printf("Warning: Failed to process comment %s: %v\n", comment.ID, err)
			}
			continue
		}
		enhancedIssue.ProcessedComments = append(enhancedIssue.ProcessedComments, processedComment)
	}
	
	// Generate work summary
	enhancedIssue.WorkSummary = p.generateWorkSummary(enhancedIssue)
	
	// Extract key activities
	enhancedIssue.KeyActivities = p.extractKeyActivities(enhancedIssue)
	
	return enhancedIssue, nil
}

// processComment converts a Jira comment to a ProcessedComment
func (p *EnhancedDataProcessor) processComment(comment jira.Comment) (ProcessedComment, error) {
	if comment.ID == "" {
		return ProcessedComment{}, fmt.Errorf("comment ID is empty")
	}
	
	text := comment.Body.Text
	
	processedComment := ProcessedComment{
		Original:         comment,
		ExtractedActions: p.extractActions(text),
		TechnicalTerms:   p.extractTechnicalTerms(text),
		WorkType:         p.determineCommentWorkType(text),
		Sentiment:        p.determineSentiment(text),
		Importance:       p.calculateCommentImportance(text),
		ActivityType:     p.determineActivityType(text),
		CompletionStatus: p.determineCommentCompletionStatus(text),
		KeyTopics:        p.extractKeyTopics(text),
	}
	
	return processedComment, nil
}

// groupCommentsByIssue groups comments by their associated issue key
func (p *EnhancedDataProcessor) groupCommentsByIssue(comments []jira.Comment) map[string][]jira.Comment {
	commentsByIssue := make(map[string][]jira.Comment)
	
	for _, comment := range comments {
		// Extract issue key from comment context or assume it's provided elsewhere
		// For now, we'll need to modify this based on how comments are associated with issues
		// This is a placeholder implementation
		issueKey := p.extractIssueKeyFromComment(comment)
		if issueKey != "" {
			commentsByIssue[issueKey] = append(commentsByIssue[issueKey], comment)
		}
	}
	
	return commentsByIssue
}

// extractIssueKeyFromComment extracts the issue key from a comment
// This is a placeholder - in practice, this association should be provided by the caller
func (p *EnhancedDataProcessor) extractIssueKeyFromComment(comment jira.Comment) string {
	// This would typically be provided by the API or calling code
	// For now, return empty string as we need the association to be provided externally
	return ""
}

// calculateIssuePriority calculates a numeric priority for an issue
func (p *EnhancedDataProcessor) calculateIssuePriority(issue jira.Issue) int {
	priority := strings.ToLower(issue.Fields.Priority.Name)
	status := strings.ToLower(issue.Fields.Status.Name)
	
	basePriority := 0
	switch priority {
	case "critical", "highest":
		basePriority = 100
	case "high":
		basePriority = 80
	case "medium":
		basePriority = 60
	case "low":
		basePriority = 40
	case "lowest":
		basePriority = 20
	default:
		basePriority = 50
	}
	
	// Adjust based on status
	if strings.Contains(status, "progress") || strings.Contains(status, "development") {
		basePriority += 20
	} else if strings.Contains(status, "blocked") {
		basePriority += 30
	} else if strings.Contains(status, "done") || strings.Contains(status, "closed") {
		basePriority += 10
	}
	
	return basePriority
}

// determineWorkType determines the type of work based on issue content
func (p *EnhancedDataProcessor) determineWorkType(issue jira.Issue) string {
	issueType := strings.ToLower(issue.Fields.IssueType.Name)
	summary := strings.ToLower(issue.Fields.Summary)
	description := strings.ToLower(issue.Fields.Description.Text)
	
	text := summary + " " + description
	
	// Check for specific work types
	if strings.Contains(issueType, "bug") || strings.Contains(text, "fix") || strings.Contains(text, "error") {
		return "bug_fix"
	}
	if strings.Contains(text, "deploy") || strings.Contains(text, "release") {
		return "deployment"
	}
	if strings.Contains(text, "terraform") || strings.Contains(text, "infrastructure") || strings.Contains(text, "aws") {
		return "infrastructure"
	}
	if strings.Contains(text, "database") || strings.Contains(text, "migration") {
		return "database"
	}
	if strings.Contains(text, "test") || strings.Contains(text, "testing") {
		return "testing"
	}
	if strings.Contains(text, "security") || strings.Contains(text, "auth") {
		return "security"
	}
	if strings.Contains(text, "review") || strings.Contains(text, "pr") {
		return "code_review"
	}
	if strings.Contains(issueType, "feature") || strings.Contains(issueType, "story") {
		return "feature_development"
	}
	
	return "general"
}

// determineCompletionStatus determines the completion status of an issue
func (p *EnhancedDataProcessor) determineCompletionStatus(issue jira.Issue) string {
	status := strings.ToLower(issue.Fields.Status.Name)
	
	if strings.Contains(status, "done") || strings.Contains(status, "closed") || strings.Contains(status, "resolved") {
		return "completed"
	}
	if strings.Contains(status, "progress") || strings.Contains(status, "development") || strings.Contains(status, "active") {
		return "in_progress"
	}
	if strings.Contains(status, "blocked") {
		return "blocked"
	}
	if strings.Contains(status, "review") {
		return "under_review"
	}
	
	return "planned"
}

// extractActions extracts action verbs from comment text
func (p *EnhancedDataProcessor) extractActions(text string) []string {
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
	
	return p.removeDuplicateStrings(actions)
}

// extractTechnicalTerms extracts technical terms from text
func (p *EnhancedDataProcessor) extractTechnicalTerms(text string) []string {
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
	
	return p.removeDuplicateStrings(terms)
}

// determineCommentWorkType determines work type from comment content
func (p *EnhancedDataProcessor) determineCommentWorkType(text string) string {
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

// determineSentiment determines the sentiment of a comment
func (p *EnhancedDataProcessor) determineSentiment(text string) string {
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

// calculateCommentImportance calculates importance score for a comment
func (p *EnhancedDataProcessor) calculateCommentImportance(text string) int {
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

// determineActivityType determines the type of activity from comment
func (p *EnhancedDataProcessor) determineActivityType(text string) string {
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

// determineCommentCompletionStatus determines completion status from comment
func (p *EnhancedDataProcessor) determineCommentCompletionStatus(text string) string {
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

// extractKeyTopics extracts key topics from comment text
func (p *EnhancedDataProcessor) extractKeyTopics(text string) []string {
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
	
	return p.removeDuplicateStrings(topics)
}

// generateWorkSummary generates a work summary for an enhanced issue
func (p *EnhancedDataProcessor) generateWorkSummary(issue EnhancedIssue) string {
	if len(issue.ProcessedComments) == 0 {
		return fmt.Sprintf("%s: %s", issue.CompletionStatus, issue.Issue.Fields.Summary)
	}
	
	// Collect key activities from comments
	var activities []string
	for _, comment := range issue.ProcessedComments {
		if len(comment.ExtractedActions) > 0 {
			activities = append(activities, comment.ExtractedActions[0]) // Take first action
		}
	}
	
	if len(activities) == 0 {
		return fmt.Sprintf("%s work on %s", issue.WorkType, issue.Issue.Fields.Summary)
	}
	
	// Create summary from activities
	uniqueActivities := p.removeDuplicateStrings(activities)
	if len(uniqueActivities) == 1 {
		return fmt.Sprintf("%s %s", strings.Title(uniqueActivities[0]), strings.ToLower(issue.Issue.Fields.Summary))
	}
	
	return fmt.Sprintf("Multiple activities: %s", strings.Join(uniqueActivities[:min(3, len(uniqueActivities))], ", "))
}

// extractKeyActivities extracts key activities from an enhanced issue
func (p *EnhancedDataProcessor) extractKeyActivities(issue EnhancedIssue) []string {
	var activities []string
	
	// Extract from issue summary and description
	text := issue.Issue.Fields.Summary + " " + issue.Issue.Fields.Description.Text
	issueActivities := p.extractActions(text)
	activities = append(activities, issueActivities...)
	
	// Extract from processed comments
	for _, comment := range issue.ProcessedComments {
		activities = append(activities, comment.ExtractedActions...)
	}
	
	// Remove duplicates and limit to top activities
	uniqueActivities := p.removeDuplicateStrings(activities)
	if len(uniqueActivities) > 5 {
		return uniqueActivities[:5]
	}
	
	return uniqueActivities
}

// extractTechnicalContextFromIssue extracts technical context from an issue and adds to overall context
func (p *EnhancedDataProcessor) extractTechnicalContextFromIssue(context *TechnicalContext, issue EnhancedIssue) {
	// Extract technologies
	for _, comment := range issue.ProcessedComments {
		context.Technologies = append(context.Technologies, comment.TechnicalTerms...)
	}
	
	// Extract actions
	for _, comment := range issue.ProcessedComments {
		context.Actions = append(context.Actions, comment.ExtractedActions...)
	}
	
	// Create specific activity records based on work type
	timestamp := time.Now()
	if !issue.Issue.Fields.Updated.Time.IsZero() {
		timestamp = issue.Issue.Fields.Updated.Time
	}
	
	switch issue.WorkType {
	case "deployment":
		context.Deployments = append(context.Deployments, DeploymentActivity{
			Type:        "deploy",
			Environment: p.extractEnvironment(issue),
			Status:      issue.CompletionStatus,
			Component:   issue.Issue.Fields.Summary,
			Description: issue.WorkSummary,
			Timestamp:   timestamp,
		})
	case "infrastructure":
		context.Infrastructure = append(context.Infrastructure, InfrastructureWork{
			Type:        "terraform",
			Action:      "configure",
			Component:   issue.Issue.Fields.Summary,
			Status:      issue.CompletionStatus,
			Description: issue.WorkSummary,
			Timestamp:   timestamp,
		})
	case "database":
		context.DatabaseWork = append(context.DatabaseWork, DatabaseActivity{
			Type:        "configuration",
			Action:      "setup",
			Database:    "postgresql", // Default, could be extracted
			Status:      issue.CompletionStatus,
			Description: issue.WorkSummary,
			Timestamp:   timestamp,
		})
	}
	
	// Remove duplicates
	context.Technologies = p.removeDuplicateStrings(context.Technologies)
	context.Actions = p.removeDuplicateStrings(context.Actions)
}

// extractEnvironment extracts environment information from issue
func (p *EnhancedDataProcessor) extractEnvironment(issue EnhancedIssue) string {
	text := strings.ToLower(issue.Issue.Fields.Summary + " " + issue.Issue.Fields.Description.Text)
	
	environments := []string{"production", "staging", "development", "test", "prod", "dev", "stage"}
	for _, env := range environments {
		if strings.Contains(text, env) {
			return env
		}
	}
	
	return "unknown"
}

// createTimelineEvents creates timeline events from an enhanced issue
func (p *EnhancedDataProcessor) createTimelineEvents(processedData *ProcessedData, issue EnhancedIssue) {
	// Create event for issue creation
	if !issue.Issue.Fields.Created.Time.IsZero() {
		event := TimelineEvent{
			Timestamp:   issue.Issue.Fields.Created.Time,
			EventType:   "issue_created",
			Description: fmt.Sprintf("Created %s: %s", issue.Issue.Key, issue.Issue.Fields.Summary),
			IssueKey:    issue.Issue.Key,
			Source:      "issue",
			Importance:  issue.Priority,
		}
		processedData.AddTimelineEvent(event)
	}
	
	// Create events for comments
	for _, comment := range issue.ProcessedComments {
		if !comment.Original.Created.Time.IsZero() {
			event := TimelineEvent{
				Timestamp:   comment.Original.Created.Time,
				EventType:   "comment_added",
				Description: fmt.Sprintf("Comment on %s: %s", issue.Issue.Key, p.shortenText(comment.Original.Body.Text, 100)),
				IssueKey:    issue.Issue.Key,
				Source:      "comment",
				Importance:  comment.Importance,
			}
			processedData.AddTimelineEvent(event)
		}
	}
}

// removeDuplicateStrings removes duplicate strings from a slice
func (p *EnhancedDataProcessor) removeDuplicateStrings(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, item := range slice {
		if item != "" && !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}

// shortenText truncates text to a maximum length
func (p *EnhancedDataProcessor) shortenText(text string, maxLength int) string {
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