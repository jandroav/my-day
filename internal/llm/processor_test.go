package llm

import (
	"testing"
	"time"
	"my-day/internal/jira"
)

// TestEnhancedDataProcessor tests the enhanced data processor functionality
func TestEnhancedDataProcessor(t *testing.T) {
	processor := NewEnhancedDataProcessor(false)
	
	// Create test data
	issues := []jira.Issue{
		{
			Key: "DEV-123",
			Fields: jira.Fields{
				Summary:     "Deploy AWS Lambda using Terraform",
				Description: jira.JiraDescription{Text: "Configure VPC and security groups for microservice deployment"},
				Status:      jira.Status{Name: "In Progress"},
				Priority:    jira.Priority{Name: "High"},
				IssueType:   jira.IssueType{Name: "Task"},
				Project:     jira.Project{Key: "DEVOPS"},
				Created:     jira.JiraTime{Time: time.Now().Add(-24 * time.Hour)},
				Updated:     jira.JiraTime{Time: time.Now().Add(-1 * time.Hour)},
			},
		},
		{
			Key: "DEV-124",
			Fields: jira.Fields{
				Summary:     "Database migration for authentication service",
				Description: jira.JiraDescription{Text: "Update PostgreSQL schema and user data migration"},
				Status:      jira.Status{Name: "Done"},
				Priority:    jira.Priority{Name: "Medium"},
				IssueType:   jira.IssueType{Name: "Task"},
				Project:     jira.Project{Key: "DEVOPS"},
				Created:     jira.JiraTime{Time: time.Now().Add(-48 * time.Hour)},
				Updated:     jira.JiraTime{Time: time.Now().Add(-2 * time.Hour)},
			},
		},
	}
	
	comments := []jira.Comment{
		{
			ID:      "1",
			Body:    jira.JiraDescription{Text: "Completed Terraform configuration for Lambda deployment"},
			Created: jira.JiraTime{Time: time.Now().Add(-2 * time.Hour)},
			Updated: jira.JiraTime{Time: time.Now().Add(-2 * time.Hour)},
		},
		{
			ID:      "2",
			Body:    jira.JiraDescription{Text: "Database migration scripts tested successfully"},
			Created: jira.JiraTime{Time: time.Now().Add(-1 * time.Hour)},
			Updated: jira.JiraTime{Time: time.Now().Add(-1 * time.Hour)},
		},
	}
	
	// Test processing
	processedData, err := processor.ProcessIssuesWithComments(issues, comments)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	
	if processedData == nil {
		t.Fatal("Expected processed data but got nil")
	}
	
	// Validate processed data
	if len(processedData.Issues) != 2 {
		t.Errorf("Expected 2 issues, got %d", len(processedData.Issues))
	}
	
	if processedData.TechnicalContext == nil {
		t.Error("Expected technical context but got nil")
	}
	
	// Check for technical terms extraction
	if len(processedData.TechnicalContext.Technologies) == 0 {
		t.Error("Expected technical terms but got none")
	}
	
	// Check for activities extraction
	if len(processedData.TechnicalContext.Actions) == 0 {
		t.Error("Expected activities but got none")
	}
}

// TestProcessedDataValidation tests data validation
func TestProcessedDataValidation(t *testing.T) {
	processor := NewEnhancedDataProcessor(false)
	
	// Test with valid data
	validIssues := []jira.Issue{
		{
			Key: "VALID-123",
			Fields: jira.Fields{
				Summary:  "Valid issue",
				Status:   jira.Status{Name: "In Progress"},
				Priority: jira.Priority{Name: "High"},
			},
		},
	}
	
	processedData, err := processor.ProcessIssuesWithComments(validIssues, []jira.Comment{})
	if err != nil {
		t.Errorf("Unexpected error with valid data: %v", err)
	}
	
	errors := processedData.Validate()
	if len(errors) > 0 {
		t.Errorf("Expected no validation errors but got: %v", errors)
	}
}

// TestWorkTypeDetection tests work type detection
func TestWorkTypeDetection(t *testing.T) {
	processor := NewEnhancedDataProcessor(false)
	
	testCases := []struct {
		name         string
		issue        jira.Issue
		expectedType string
	}{
		{
			name: "Infrastructure work",
			issue: jira.Issue{
				Key: "INFRA-1",
				Fields: jira.Fields{
					Summary:     "Deploy AWS resources using Terraform",
					Description: jira.JiraDescription{Text: "Setup VPC and security groups"},
					IssueType:   jira.IssueType{Name: "Task"},
				},
			},
			expectedType: "infrastructure",
		},
		{
			name: "Database work",
			issue: jira.Issue{
				Key: "DB-1",
				Fields: jira.Fields{
					Summary:     "PostgreSQL migration",
					Description: jira.JiraDescription{Text: "Update database schema"},
					IssueType:   jira.IssueType{Name: "Task"},
				},
			},
			expectedType: "database",
		},
		{
			name: "Bug fix work",
			issue: jira.Issue{
				Key: "BUG-1",
				Fields: jira.Fields{
					Summary:     "Fix authentication error",
					Description: jira.JiraDescription{Text: "User login failing"},
					IssueType:   jira.IssueType{Name: "Bug"},
				},
			},
			expectedType: "bug_fix",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			workType := processor.determineWorkType(tc.issue)
			if workType != tc.expectedType {
				t.Errorf("Expected work type '%s', got '%s'", tc.expectedType, workType)
			}
		})
	}
}

// TestTechnicalTermExtraction tests technical term extraction
func TestTechnicalTermExtraction(t *testing.T) {
	processor := NewEnhancedDataProcessor(false)
	
	testCases := []struct {
		name          string
		text          string
		expectedTerms []string
	}{
		{
			name:          "AWS and Terraform",
			text:          "Deploy AWS Lambda using Terraform configuration",
			expectedTerms: []string{"terraform", "aws"},
		},
		{
			name:          "Database terms",
			text:          "PostgreSQL database migration with SQL scripts",
			expectedTerms: []string{"postgresql", "sql", "database"},
		},
		{
			name:          "Kubernetes deployment",
			text:          "Deploy Docker containers to Kubernetes cluster",
			expectedTerms: []string{"kubernetes", "docker"},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			terms := processor.extractTechnicalTerms(tc.text)
			
			for _, expected := range tc.expectedTerms {
				found := false
				for _, term := range terms {
					if term == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find term '%s' in %v", expected, terms)
				}
			}
		})
	}
}

// TestActionExtraction tests action extraction from comments
func TestActionExtraction(t *testing.T) {
	processor := NewEnhancedDataProcessor(false)
	
	testCases := []struct {
		name            string
		text            string
		expectedActions []string
	}{
		{
			name:            "Deployment actions",
			text:            "Implemented new service and deployed to production",
			expectedActions: []string{"implemented", "deployed"},
		},
		{
			name:            "Fix actions",
			text:            "Fixed authentication bug and updated security settings",
			expectedActions: []string{"fixed", "updated"},
		},
		{
			name:            "Configuration actions",
			text:            "Configured database connection and tested migration",
			expectedActions: []string{"configured", "tested"},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actions := processor.extractActions(tc.text)
			
			for _, expected := range tc.expectedActions {
				found := false
				for _, action := range actions {
					if action == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find action '%s' in %v", expected, actions)
				}
			}
		})
	}
}

// TestSentimentAnalysis tests sentiment analysis
func TestSentimentAnalysis(t *testing.T) {
	processor := NewEnhancedDataProcessor(false)
	
	testCases := []struct {
		name              string
		text              string
		expectedSentiment string
	}{
		{
			name:              "Positive sentiment",
			text:              "Successfully completed deployment. Everything working great!",
			expectedSentiment: "positive",
		},
		{
			name:              "Negative sentiment",
			text:              "Deployment failed. Major issues with configuration. Service is broken.",
			expectedSentiment: "negative",
		},
		{
			name:              "Neutral sentiment",
			text:              "Updated configuration settings for the service",
			expectedSentiment: "neutral",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sentiment := processor.determineSentiment(tc.text)
			if sentiment != tc.expectedSentiment {
				t.Errorf("Expected sentiment '%s', got '%s'", tc.expectedSentiment, sentiment)
			}
		})
	}
}

// TestCommentImportanceCalculation tests comment importance calculation
func TestCommentImportanceCalculation(t *testing.T) {
	processor := NewEnhancedDataProcessor(false)
	
	testCases := []struct {
		name              string
		text              string
		expectedMinScore  int
		expectedMaxScore  int
	}{
		{
			name:              "Critical issue",
			text:              "Critical security vulnerability in production requires urgent fix",
			expectedMinScore:  80,
			expectedMaxScore:  100,
		},
		{
			name:              "Completed deployment",
			text:              "Deployment completed successfully. Service is running in production.",
			expectedMinScore:  60,
			expectedMaxScore:  90,
		},
		{
			name:              "Regular update",
			text:              "Updated configuration file with new settings",
			expectedMinScore:  40,
			expectedMaxScore:  70,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			importance := processor.calculateCommentImportance(tc.text)
			if importance < tc.expectedMinScore || importance > tc.expectedMaxScore {
				t.Errorf("Expected importance between %d and %d, got %d", 
					tc.expectedMinScore, tc.expectedMaxScore, importance)
			}
		})
	}
}

// TestActivityTypeDetection tests activity type detection
func TestActivityTypeDetection(t *testing.T) {
	processor := NewEnhancedDataProcessor(false)
	
	testCases := []struct {
		name             string
		text             string
		expectedActivity string
	}{
		{
			name:             "Completion activity",
			text:             "Completed the database migration successfully",
			expectedActivity: "completion",
		},
		{
			name:             "Initiation activity",
			text:             "Started working on the new authentication service",
			expectedActivity: "initiation",
		},
		{
			name:             "Update activity",
			text:             "Updated the configuration files for the service",
			expectedActivity: "update",
		},
		{
			name:             "Investigation activity",
			text:             "Investigating performance issues in the production environment",
			expectedActivity: "investigation",
		},
		{
			name:             "Blocker activity",
			text:             "Blocked by database connectivity issues",
			expectedActivity: "blocker",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			activity := processor.determineActivityType(tc.text)
			if activity != tc.expectedActivity {
				t.Errorf("Expected activity '%s', got '%s'", tc.expectedActivity, activity)
			}
		})
	}
}

// TestProcessorWithDebugMode tests processor with debug mode enabled
func TestProcessorWithDebugMode(t *testing.T) {
	debugProcessor := NewEnhancedDataProcessor(true)
	
	issues := []jira.Issue{
		{
			Key: "DEBUG-1",
			Fields: jira.Fields{
				Summary:  "Test issue for debug mode",
				Status:   jira.Status{Name: "In Progress"},
				Priority: jira.Priority{Name: "Medium"},
			},
		},
	}
	
	comments := []jira.Comment{
		{
			ID:   "1",
			Body: jira.JiraDescription{Text: "Test comment for debug mode"},
		},
	}
	
	// Should process successfully even in debug mode
	processedData, err := debugProcessor.ProcessIssuesWithComments(issues, comments)
	if err != nil {
		t.Errorf("Unexpected error in debug mode: %v", err)
	}
	
	if processedData == nil {
		t.Error("Expected processed data but got nil")
	}
}

// TestProcessorWithEmptyData tests processor with empty data
func TestProcessorWithEmptyData(t *testing.T) {
	processor := NewEnhancedDataProcessor(false)
	
	// Test with empty issues and comments
	processedData, err := processor.ProcessIssuesWithComments([]jira.Issue{}, []jira.Comment{})
	if err != nil {
		t.Errorf("Unexpected error with empty data: %v", err)
	}
	
	if processedData == nil {
		t.Error("Expected processed data but got nil")
	}
	
	if len(processedData.Issues) != 0 {
		t.Errorf("Expected 0 issues, got %d", len(processedData.Issues))
	}
}

// TestProcessorWithMalformedData tests processor with malformed data
func TestProcessorWithMalformedData(t *testing.T) {
	processor := NewEnhancedDataProcessor(false)
	
	// Test with malformed issues
	malformedIssues := []jira.Issue{
		{
			Key: "", // Empty key
			Fields: jira.Fields{
				Summary: "Test issue",
			},
		},
	}
	
	malformedComments := []jira.Comment{
		{
			ID:   "", // Empty ID
			Body: jira.JiraDescription{Text: "Test comment"},
		},
	}
	
	// Should handle malformed data gracefully
	processedData, err := processor.ProcessIssuesWithComments(malformedIssues, malformedComments)
	if err != nil {
		t.Errorf("Expected graceful handling of malformed data, got error: %v", err)
	}
	
	if processedData == nil {
		t.Error("Expected processed data but got nil")
	}
}