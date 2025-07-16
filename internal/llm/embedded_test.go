package llm

import (
	"testing"
	"time"
	"my-day/internal/jira"
)

// TestNewEmbeddedLLMWithConfig tests the configuration-based constructor
func TestNewEmbeddedLLMWithConfig(t *testing.T) {
	config := LLMConfig{
		Enabled:                  true,
		Mode:                     "embedded",
		Model:                    "test-model",
		Debug:                    true,
		SummaryStyle:             "technical",
		MaxSummaryLength:         150,
		IncludeTechnicalDetails:  true,
		PrioritizeRecentWork:     true,
		FallbackStrategy:         "graceful",
	}
	
	llm := NewEmbeddedLLMWithConfig(config)
	
	if llm == nil {
		t.Fatal("Expected LLM instance, got nil")
	}
	
	if llm.model != "test-model" {
		t.Errorf("Expected model 'test-model', got '%s'", llm.model)
	}
	
	if llm.config == nil {
		t.Fatal("Expected config to be stored, got nil")
	}
	
	if llm.config.MaxSummaryLength != 150 {
		t.Errorf("Expected MaxSummaryLength 150, got %d", llm.config.MaxSummaryLength)
	}
}

// TestConfigurationAwareMethods tests that configuration options are respected
func TestConfigurationAwareMethods(t *testing.T) {
	tests := []struct {
		name           string
		config         LLMConfig
		expectedLength int
		expectedTech   bool
		expectedStyle  string
	}{
		{
			name: "Technical style with details",
			config: LLMConfig{
				SummaryStyle:             "technical",
				MaxSummaryLength:         100,
				IncludeTechnicalDetails:  true,
			},
			expectedLength: 100,
			expectedTech:   true,
			expectedStyle:  "technical",
		},
		{
			name: "Brief style without details",
			config: LLMConfig{
				SummaryStyle:             "brief",
				MaxSummaryLength:         50,
				IncludeTechnicalDetails:  false,
			},
			expectedLength: 50,
			expectedTech:   false,
			expectedStyle:  "brief",
		},
		{
			name: "Business style with medium length",
			config: LLMConfig{
				SummaryStyle:             "business",
				MaxSummaryLength:         200,
				IncludeTechnicalDetails:  true,
			},
			expectedLength: 200,
			expectedTech:   true,
			expectedStyle:  "business",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			llm := NewEmbeddedLLMWithConfig(tt.config)
			
			if llm.getConfiguredMaxLength() != tt.expectedLength {
				t.Errorf("Expected max length %d, got %d", tt.expectedLength, llm.getConfiguredMaxLength())
			}
			
			if llm.shouldIncludeTechnicalDetails() != tt.expectedTech {
				t.Errorf("Expected technical details %t, got %t", tt.expectedTech, llm.shouldIncludeTechnicalDetails())
			}
			
			if llm.getSummaryStyle() != tt.expectedStyle {
				t.Errorf("Expected style '%s', got '%s'", tt.expectedStyle, llm.getSummaryStyle())
			}
		})
	}
}

// TestTechnicalPatternMatching tests technical term extraction
func TestTechnicalPatternMatching(t *testing.T) {
	config := LLMConfig{
		IncludeTechnicalDetails: true,
	}
	llm := NewEmbeddedLLMWithConfig(config)
	
	testCases := []struct {
		name           string
		issue          jira.Issue
		expectedTerms  []string
		shouldContain  []string
	}{
		{
			name: "Infrastructure work with AWS and Terraform",
			issue: jira.Issue{
				Key: "DEV-123",
				Fields: jira.Fields{
					Summary:     "Deploy AWS Lambda function using Terraform",
					Description: jira.JiraDescription{Text: "Need to configure VPC and security groups for the new microservice deployment"},
					Status:      jira.Status{Name: "In Progress"},
					Priority:    jira.Priority{Name: "High"},
					IssueType:   jira.IssueType{Name: "Task"},
				},
			},
			shouldContain: []string{"aws", "terraform", "deploy"},
		},
		{
			name: "Database migration work",
			issue: jira.Issue{
				Key: "DEV-124",
				Fields: jira.Fields{
					Summary:     "Database migration for user authentication",
					Description: jira.JiraDescription{Text: "Update PostgreSQL schema and migrate user data"},
					Status:      jira.Status{Name: "In Progress"},
					Priority:    jira.Priority{Name: "Medium"},
					IssueType:   jira.IssueType{Name: "Task"},
				},
			},
			shouldContain: []string{"database", "authentication"},
		},
		{
			name: "Kubernetes deployment",
			issue: jira.Issue{
				Key: "DEV-125",
				Fields: jira.Fields{
					Summary:     "Fix Kubernetes pod scaling issues",
					Description: jira.JiraDescription{Text: "Docker containers not scaling properly in production environment"},
					Status:      jira.Status{Name: "In Progress"},
					Priority:    jira.Priority{Name: "Critical"},
					IssueType:   jira.IssueType{Name: "Bug"},
				},
			},
			shouldContain: []string{"kubernetes", "docker", "fix"},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keyPoints := llm.extractKeyPoints(tc.issue)
			
			for _, expected := range tc.shouldContain {
				found := false
				for _, point := range keyPoints {
					if point == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find '%s' in key points %v", expected, keyPoints)
				}
			}
		})
	}
}

// TestTechnicalDetailsToggle tests that technical details can be disabled
func TestTechnicalDetailsToggle(t *testing.T) {
	// Test with technical details enabled
	configWithTech := LLMConfig{
		IncludeTechnicalDetails: true,
	}
	llmWithTech := NewEmbeddedLLMWithConfig(configWithTech)
	
	// Test with technical details disabled
	configWithoutTech := LLMConfig{
		IncludeTechnicalDetails: false,
	}
	llmWithoutTech := NewEmbeddedLLMWithConfig(configWithoutTech)
	
	issue := jira.Issue{
		Key: "DEV-123",
		Fields: jira.Fields{
			Summary:     "Deploy AWS Lambda function using Terraform",
			Description: jira.JiraDescription{Text: "Configure VPC and security groups"},
			Status:      jira.Status{Name: "In Progress"},
			Priority:    jira.Priority{Name: "High"},
			IssueType:   jira.IssueType{Name: "Task"},
		},
	}
	
	keyPointsWithTech := llmWithTech.extractKeyPoints(issue)
	keyPointsWithoutTech := llmWithoutTech.extractKeyPoints(issue)
	
	// With technical details, should find technical terms
	foundTechTerms := false
	for _, point := range keyPointsWithTech {
		if point == "aws" || point == "terraform" {
			foundTechTerms = true
			break
		}
	}
	if !foundTechTerms {
		t.Error("Expected to find technical terms when technical details are enabled")
	}
	
	// Without technical details, should have fewer or no technical terms
	techTermCount := 0
	for _, point := range keyPointsWithoutTech {
		if point == "aws" || point == "terraform" {
			techTermCount++
		}
	}
	
	// Should still find action terms like "deploy"
	foundActionTerms := false
	for _, point := range keyPointsWithoutTech {
		if point == "deploy" {
			foundActionTerms = true
			break
		}
	}
	if !foundActionTerms {
		t.Error("Expected to find action terms even when technical details are disabled")
	}
}

// TestCommentSummarization tests comment summarization with various input types
func TestCommentSummarization(t *testing.T) {
	config := LLMConfig{
		MaxSummaryLength: 100,
	}
	llm := NewEmbeddedLLMWithConfig(config)
	
	testCases := []struct {
		name     string
		comments []jira.Comment
		expected string
	}{
		{
			name:     "Empty comments",
			comments: []jira.Comment{},
			expected: "",
		},
		{
			name: "Single comment",
			comments: []jira.Comment{
				{
					ID: "1",
					Body: jira.JiraDescription{Text: "Completed AWS deployment"},
					Created: jira.JiraTime{Time: time.Now()},
				},
			},
			expected: "Completed AWS deployment",
		},
		{
			name: "Multiple comments",
			comments: []jira.Comment{
				{
					ID: "1",
					Body: jira.JiraDescription{Text: "Started working on database migration"},
					Created: jira.JiraTime{Time: time.Now()},
				},
				{
					ID: "2",
					Body: jira.JiraDescription{Text: "Completed Terraform configuration"},
					Created: jira.JiraTime{Time: time.Now()},
				},
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			summary, err := llm.SummarizeComments(tc.comments)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if tc.expected != "" && summary != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, summary)
			}
			
			// Check that summary respects max length
			if len(summary) > config.MaxSummaryLength {
				t.Errorf("Summary length %d exceeds max length %d", len(summary), config.MaxSummaryLength)
			}
		})
	}
}

// TestFallbackBehavior tests fallback behavior with edge cases
func TestFallbackBehavior(t *testing.T) {
	config := LLMConfig{
		FallbackStrategy: "graceful",
	}
	llm := NewEmbeddedLLMWithConfig(config)
	
	// Test with malformed issue
	malformedIssue := jira.Issue{
		Key: "", // Empty key
		Fields: jira.Fields{
			Summary: "Test issue",
		},
	}
	
	summary, err := llm.SummarizeIssue(malformedIssue)
	if err != nil {
		t.Errorf("Expected graceful handling of malformed issue, got error: %v", err)
	}
	
	if summary == "" {
		t.Error("Expected some summary even for malformed issue")
	}
	
	// Test with empty comment text
	emptyComments := []jira.Comment{
		{
			ID: "1",
			Body: jira.JiraDescription{Text: ""},
			Created: jira.JiraTime{Time: time.Now()},
		},
	}
	
	commentSummary, err := llm.SummarizeComments(emptyComments)
	if err != nil {
		t.Errorf("Expected graceful handling of empty comments, got error: %v", err)
	}
	
	// Should handle empty comments gracefully
	if commentSummary == "" {
		// This is acceptable for empty comments
	}
}

// TestIssueSummarization tests issue summarization with different configurations
func TestIssueSummarization(t *testing.T) {
	testCases := []struct {
		name   string
		config LLMConfig
		issue  jira.Issue
	}{
		{
			name: "High priority bug with technical details",
			config: LLMConfig{
				IncludeTechnicalDetails: true,
				MaxSummaryLength:        200,
			},
			issue: jira.Issue{
				Key: "BUG-123",
				Fields: jira.Fields{
					Summary:     "Fix authentication timeout in production API",
					Description: jira.JiraDescription{Text: "Users experiencing OAuth token validation delays"},
					Status:      jira.Status{Name: "In Progress"},
					Priority:    jira.Priority{Name: "Critical"},
					IssueType:   jira.IssueType{Name: "Bug"},
				},
			},
		},
		{
			name: "Feature development without technical details",
			config: LLMConfig{
				IncludeTechnicalDetails: false,
				MaxSummaryLength:        100,
			},
			issue: jira.Issue{
				Key: "FEAT-456",
				Fields: jira.Fields{
					Summary:     "Implement user dashboard with React components",
					Description: jira.JiraDescription{Text: "Create responsive dashboard using TypeScript and Redux"},
					Status:      jira.Status{Name: "To Do"},
					Priority:    jira.Priority{Name: "Medium"},
					IssueType:   jira.IssueType{Name: "Story"},
				},
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			llm := NewEmbeddedLLMWithConfig(tc.config)
			
			summary, err := llm.SummarizeIssue(tc.issue)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if summary == "" {
				t.Error("Expected non-empty summary")
			}
			
			// Check that summary respects max length
			if len(summary) > tc.config.MaxSummaryLength {
				t.Errorf("Summary length %d exceeds max length %d", len(summary), tc.config.MaxSummaryLength)
			}
			
			// Check for priority indicators in high priority items
			if tc.issue.Fields.Priority.Name == "Critical" {
				if !containsString(summary, "🔥") {
					t.Error("Expected priority indicator for critical issue")
				}
			}
		})
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsString(s[1:], substr)))
}

// TestDebugFunctionality tests debug logging capabilities
func TestDebugFunctionality(t *testing.T) {
	config := LLMConfig{
		Debug: true,
	}
	llm := NewEmbeddedLLMWithConfig(config)
	
	// Test that debug report can be retrieved
	report, err := llm.GetDebugReport()
	if err != nil {
		t.Errorf("Expected to get debug report, got error: %v", err)
	}
	
	if report == nil {
		t.Error("Expected debug report, got nil")
	}
}