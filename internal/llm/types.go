package llm

import (
	"fmt"
	"strings"
	"time"
	"my-day/internal/jira"
)

// ProcessedData represents enhanced processed data from Jira issues and comments
type ProcessedData struct {
	Issues           []EnhancedIssue   `json:"issues"`
	TechnicalContext *TechnicalContext `json:"technical_context"`
	WorkPatterns     []WorkPattern     `json:"work_patterns"`
	TimelineEvents   []TimelineEvent   `json:"timeline_events"`
	ProcessedAt      time.Time         `json:"processed_at"`
}

// EnhancedIssue represents a Jira issue with enhanced processing and analysis
type EnhancedIssue struct {
	Issue             jira.Issue         `json:"issue"`
	Comments          []jira.Comment     `json:"comments"`
	ProcessedComments []ProcessedComment `json:"processed_comments"`
	TechnicalContext  *TechnicalContext  `json:"technical_context"`
	WorkSummary       string             `json:"work_summary"`
	Priority          int                `json:"priority"`
	WorkType          string             `json:"work_type"`
	CompletionStatus  string             `json:"completion_status"`
	KeyActivities     []string           `json:"key_activities"`
}

// ProcessedComment represents an analyzed comment with extracted insights
type ProcessedComment struct {
	Original         jira.Comment `json:"original"`
	ExtractedActions []string     `json:"extracted_actions"`
	TechnicalTerms   []string     `json:"technical_terms"`
	WorkType         string       `json:"work_type"`
	Sentiment        string       `json:"sentiment"`
	Importance       int          `json:"importance"`
	ActivityType     string       `json:"activity_type"`
	CompletionStatus string       `json:"completion_status"`
	KeyTopics        []string     `json:"key_topics"`
}

// TechnicalContext represents technical context extracted from issues and comments
type TechnicalContext struct {
	Technologies     []string               `json:"technologies"`
	Environments     []string               `json:"environments"`
	Actions          []string               `json:"actions"`
	Deployments      []DeploymentActivity   `json:"deployments"`
	Infrastructure   []InfrastructureWork   `json:"infrastructure"`
	DatabaseWork     []DatabaseActivity     `json:"database_work"`
	SecurityWork     []SecurityActivity     `json:"security_work"`
	TestingWork      []TestingActivity      `json:"testing_work"`
	CodeReviewWork   []CodeReviewActivity   `json:"code_review_work"`
}

// WorkPattern represents identified patterns in work activities
type WorkPattern struct {
	Type        string    `json:"type"`
	Pattern     string    `json:"pattern"`
	Frequency   int       `json:"frequency"`
	Confidence  float64   `json:"confidence"`
	Examples    []string  `json:"examples"`
	LastSeen    time.Time `json:"last_seen"`
}

// TimelineEvent represents a chronological event in the work timeline
type TimelineEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	IssueKey    string    `json:"issue_key"`
	Source      string    `json:"source"` // "comment", "status_change", "worklog"
	Importance  int       `json:"importance"`
}

// DeploymentActivity represents deployment-related work
type DeploymentActivity struct {
	Type        string    `json:"type"` // "deploy", "rollback", "prepare", "validate"
	Environment string    `json:"environment"`
	Status      string    `json:"status"` // "completed", "in_progress", "planned", "blocked"
	Component   string    `json:"component"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

// InfrastructureWork represents infrastructure-related activities
type InfrastructureWork struct {
	Type        string    `json:"type"` // "terraform", "aws", "kubernetes", "networking"
	Action      string    `json:"action"` // "configure", "deploy", "update", "troubleshoot"
	Component   string    `json:"component"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

// DatabaseActivity represents database-related work
type DatabaseActivity struct {
	Type        string    `json:"type"` // "migration", "permissions", "configuration", "troubleshooting"
	Action      string    `json:"action"`
	Database    string    `json:"database"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

// SecurityActivity represents security-related work
type SecurityActivity struct {
	Type        string    `json:"type"` // "authentication", "authorization", "secrets", "compliance"
	Action      string    `json:"action"`
	Component   string    `json:"component"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

// TestingActivity represents testing-related work
type TestingActivity struct {
	Type        string    `json:"type"` // "unit", "integration", "e2e", "performance", "security"
	Action      string    `json:"action"` // "create", "run", "fix", "update"
	Status      string    `json:"status"`
	Results     string    `json:"results"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

// CodeReviewActivity represents code review activities
type CodeReviewActivity struct {
	Type        string    `json:"type"` // "create_pr", "review", "approve", "merge", "address_feedback"
	Action      string    `json:"action"`
	PRNumber    string    `json:"pr_number"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}

// ValidationError represents validation errors for malformed data
type ValidationError struct {
	Field   string `json:"field"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// ProcessingError represents errors that occur during data processing
type ProcessingError struct {
	Type        string            `json:"type"`
	Message     string            `json:"message"`
	Source      string            `json:"source"`
	Validations []ValidationError `json:"validations,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
}

// Validate validates the ProcessedData structure
func (pd *ProcessedData) Validate() []ValidationError {
	var errors []ValidationError
	
	// Validate issues
	for i, issue := range pd.Issues {
		if issue.Issue.Key == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("issues[%d].issue.key", i),
				Value:   issue.Issue.Key,
				Message: "Issue key cannot be empty",
			})
		}
		
		// Validate processed comments
		for j, comment := range issue.ProcessedComments {
			if comment.Original.ID == "" {
				errors = append(errors, ValidationError{
					Field:   fmt.Sprintf("issues[%d].processed_comments[%d].original.id", i, j),
					Value:   comment.Original.ID,
					Message: "Comment ID cannot be empty",
				})
			}
		}
	}
	
	// Validate technical context
	if pd.TechnicalContext != nil {
		if err := pd.TechnicalContext.Validate(); len(err) > 0 {
			errors = append(errors, err...)
		}
	}
	
	return errors
}

// Validate validates the TechnicalContext structure
func (tc *TechnicalContext) Validate() []ValidationError {
	var errors []ValidationError
	
	// Validate deployment activities
	for i, deployment := range tc.Deployments {
		if deployment.Type == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("technical_context.deployments[%d].type", i),
				Value:   deployment.Type,
				Message: "Deployment type cannot be empty",
			})
		}
		if deployment.Status == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("technical_context.deployments[%d].status", i),
				Value:   deployment.Status,
				Message: "Deployment status cannot be empty",
			})
		}
	}
	
	// Validate infrastructure work
	for i, infra := range tc.Infrastructure {
		if infra.Type == "" {
			errors = append(errors, ValidationError{
				Field:   fmt.Sprintf("technical_context.infrastructure[%d].type", i),
				Value:   infra.Type,
				Message: "Infrastructure type cannot be empty",
			})
		}
	}
	
	return errors
}

// GetSummary returns a brief summary of the processed data
func (pd *ProcessedData) GetSummary() string {
	if len(pd.Issues) == 0 {
		return "No issues processed"
	}
	
	var summaryParts []string
	
	// Count different types of work
	completedCount := 0
	inProgressCount := 0
	
	for _, issue := range pd.Issues {
		switch strings.ToLower(issue.CompletionStatus) {
		case "completed", "done", "resolved":
			completedCount++
		case "in_progress", "working", "active":
			inProgressCount++
		}
	}
	
	if completedCount > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d completed", completedCount))
	}
	if inProgressCount > 0 {
		summaryParts = append(summaryParts, fmt.Sprintf("%d in progress", inProgressCount))
	}
	
	// Add technical context summary
	if pd.TechnicalContext != nil {
		if len(pd.TechnicalContext.Technologies) > 0 {
			summaryParts = append(summaryParts, fmt.Sprintf("using %s", strings.Join(pd.TechnicalContext.Technologies[:min(3, len(pd.TechnicalContext.Technologies))], ", ")))
		}
	}
	
	if len(summaryParts) == 0 {
		return fmt.Sprintf("Processed %d issues", len(pd.Issues))
	}
	
	return strings.Join(summaryParts, ", ")
}

// GetKeyActivities returns the most important activities across all issues
func (pd *ProcessedData) GetKeyActivities() []string {
	var activities []string
	activityCount := make(map[string]int)
	
	// Collect activities from all issues
	for _, issue := range pd.Issues {
		for _, activity := range issue.KeyActivities {
			if activity != "" {
				activityCount[activity]++
				if activityCount[activity] == 1 { // First occurrence
					activities = append(activities, activity)
				}
			}
		}
	}
	
	// Sort by frequency and return top activities
	// For now, return first 5 unique activities
	if len(activities) > 5 {
		return activities[:5]
	}
	
	return activities
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NewProcessedData creates a new ProcessedData instance with validation
func NewProcessedData() *ProcessedData {
	return &ProcessedData{
		Issues:           make([]EnhancedIssue, 0),
		TechnicalContext: &TechnicalContext{
			Technologies:   make([]string, 0),
			Environments:   make([]string, 0),
			Actions:        make([]string, 0),
			Deployments:    make([]DeploymentActivity, 0),
			Infrastructure: make([]InfrastructureWork, 0),
			DatabaseWork:   make([]DatabaseActivity, 0),
			SecurityWork:   make([]SecurityActivity, 0),
			TestingWork:    make([]TestingActivity, 0),
			CodeReviewWork: make([]CodeReviewActivity, 0),
		},
		WorkPatterns:   make([]WorkPattern, 0),
		TimelineEvents: make([]TimelineEvent, 0),
		ProcessedAt:    time.Now(),
	}
}

// AddIssue adds an enhanced issue to the processed data with validation
func (pd *ProcessedData) AddIssue(issue EnhancedIssue) error {
	// Validate the issue
	if issue.Issue.Key == "" {
		return fmt.Errorf("issue key cannot be empty")
	}
	
	// Check for duplicates
	for _, existingIssue := range pd.Issues {
		if existingIssue.Issue.Key == issue.Issue.Key {
			return fmt.Errorf("issue %s already exists", issue.Issue.Key)
		}
	}
	
	pd.Issues = append(pd.Issues, issue)
	return nil
}

// AddTimelineEvent adds a timeline event with validation
func (pd *ProcessedData) AddTimelineEvent(event TimelineEvent) error {
	if event.EventType == "" {
		return fmt.Errorf("event type cannot be empty")
	}
	if event.Description == "" {
		return fmt.Errorf("event description cannot be empty")
	}
	
	pd.TimelineEvents = append(pd.TimelineEvents, event)
	return nil
}