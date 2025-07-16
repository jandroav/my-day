package llm

import (
	"fmt"
	"strings"
	"testing"
	"time"
	"my-day/internal/jira"
)

// TestFullLLMPipelineIntegration tests the complete LLM pipeline with realistic data
func TestFullLLMPipelineIntegration(t *testing.T) {
	// Create realistic DevOps Jira issues and comments
	issues, comments, worklogs := createRealisticDevOpsData()
	
	// Test with different LLM configurations
	testConfigs := []struct {
		name   string
		config LLMConfig
	}{
		{
			name: "Technical style with full details",
			config: LLMConfig{
				Enabled:                  true,
				Mode:                     "embedded",
				Model:                    "test-model",
				Debug:                    false,
				SummaryStyle:             "technical",
				MaxSummaryLength:         200,
				IncludeTechnicalDetails:  true,
				PrioritizeRecentWork:     true,
				FallbackStrategy:         "graceful",
			},
		},
		{
			name: "Business style without technical details",
			config: LLMConfig{
				Enabled:                  true,
				Mode:                     "embedded",
				Model:                    "test-model",
				Debug:                    false,
				SummaryStyle:             "business",
				MaxSummaryLength:         150,
				IncludeTechnicalDetails:  false,
				PrioritizeRecentWork:     true,
				FallbackStrategy:         "graceful",
			},
		},
		{
			name: "Brief style with minimal details",
			config: LLMConfig{
				Enabled:                  true,
				Mode:                     "embedded",
				Model:                    "test-model",
				Debug:                    false,
				SummaryStyle:             "brief",
				MaxSummaryLength:         100,
				IncludeTechnicalDetails:  false,
				PrioritizeRecentWork:     true,
				FallbackStrategy:         "graceful",
			},
		},
	}
	
	for _, tc := range testConfigs {
		t.Run(tc.name, func(t *testing.T) {
			llm := NewEmbeddedLLMWithConfig(tc.config)
			
			// Test individual issue summarization
			for _, issue := range issues {
				summary, err := llm.SummarizeIssue(issue)
				if err != nil {
					t.Errorf("Failed to summarize issue %s: %v", issue.Key, err)
					continue
				}
				
				if summary == "" {
					t.Errorf("Empty summary for issue %s", issue.Key)
					continue
				}
				
				// Validate summary length
				if len(summary) > tc.config.MaxSummaryLength {
					t.Errorf("Summary for %s exceeds max length: %d > %d", 
						issue.Key, len(summary), tc.config.MaxSummaryLength)
				}
				
				// Validate technical details inclusion
				if tc.config.IncludeTechnicalDetails {
					if !containsTechnicalTerms(summary) && containsTechnicalTerms(issue.Fields.Summary+" "+issue.Fields.Description.Text) {
						t.Logf("Warning: Expected technical terms in summary for %s: %s", issue.Key, summary)
					}
				}
			}
			
			// Test comment summarization
			commentSummary, err := llm.SummarizeComments(comments)
			if err != nil {
				t.Errorf("Failed to summarize comments: %v", err)
			} else if commentSummary != "" {
				// Allow a small tolerance (5% or 10 characters, whichever is smaller)
				tolerance := min(10, tc.config.MaxSummaryLength/20)
				if len(commentSummary) > tc.config.MaxSummaryLength + tolerance {
					t.Errorf("Comment summary exceeds max length by more than tolerance: %d > %d (tolerance: %d)", 
						len(commentSummary), tc.config.MaxSummaryLength, tolerance)
				}
			}
			
			// Test full standup summary generation
			standupSummary, err := llm.GenerateStandupSummaryWithComments(issues, comments, worklogs)
			if err != nil {
				t.Errorf("Failed to generate standup summary: %v", err)
			} else {
				if standupSummary == "" {
					t.Error("Empty standup summary")
				}
				
				// Validate summary quality
				validateSummaryQuality(t, standupSummary, tc.config, "standup")
			}
		})
	}
}

// TestLLMPipelineWithDifferentDataSizes tests performance with various data sizes
func TestLLMPipelineWithDifferentDataSizes(t *testing.T) {
	config := LLMConfig{
		Enabled:                  true,
		Mode:                     "embedded",
		Model:                    "test-model",
		SummaryStyle:             "technical",
		MaxSummaryLength:         200,
		IncludeTechnicalDetails:  true,
	}
	llm := NewEmbeddedLLMWithConfig(config)
	
	testCases := []struct {
		name         string
		issueCount   int
		commentCount int
		maxDuration  time.Duration
	}{
		{
			name:         "Small dataset",
			issueCount:   2,
			commentCount: 3,
			maxDuration:  100 * time.Millisecond,
		},
		{
			name:         "Medium dataset",
			issueCount:   5,
			commentCount: 10,
			maxDuration:  500 * time.Millisecond,
		},
		{
			name:         "Large dataset",
			issueCount:   10,
			commentCount: 25,
			maxDuration:  2 * time.Second,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			issues := createTestIssues(tc.issueCount)
			comments := createTestComments(tc.commentCount)
			worklogs := createTestWorklogs(tc.issueCount)
			
			start := time.Now()
			summary, err := llm.GenerateStandupSummaryWithComments(issues, comments, worklogs)
			duration := time.Since(start)
			
			if err != nil {
				t.Errorf("Failed to generate summary: %v", err)
			}
			
			if summary == "" {
				t.Error("Empty summary generated")
			}
			
			if duration > tc.maxDuration {
				t.Errorf("Processing took too long: %v (max: %v)", duration, tc.maxDuration)
			}
			
			t.Logf("Processed %d issues, %d comments in %v", tc.issueCount, tc.commentCount, duration)
		})
	}
}

// TestLLMPipelineErrorHandling tests error handling in the full pipeline
func TestLLMPipelineErrorHandling(t *testing.T) {
	config := LLMConfig{
		Enabled:          true,
		Mode:             "embedded",
		Model:            "test-model",
		FallbackStrategy: "graceful",
	}
	llm := NewEmbeddedLLMWithConfig(config)
	
	testCases := []struct {
		name     string
		issues   []jira.Issue
		comments []jira.Comment
		worklogs []jira.WorklogEntry
		expectError bool
	}{
		{
			name:     "Empty data",
			issues:   []jira.Issue{},
			comments: []jira.Comment{},
			worklogs: []jira.WorklogEntry{},
			expectError: false, // Should handle gracefully
		},
		{
			name: "Malformed issues",
			issues: []jira.Issue{
				{
					Key: "", // Empty key
					Fields: jira.Fields{
						Summary: "Test issue",
					},
				},
			},
			comments: []jira.Comment{},
			worklogs: []jira.WorklogEntry{},
			expectError: false, // Should handle gracefully
		},
		{
			name:   "Empty comments",
			issues: createTestIssues(1),
			comments: []jira.Comment{
				{
					ID: "1",
					Body: jira.JiraDescription{Text: ""}, // Empty text
					Created: jira.JiraTime{Time: time.Now()},
				},
			},
			worklogs: []jira.WorklogEntry{},
			expectError: false, // Should handle gracefully
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			summary, err := llm.GenerateStandupSummaryWithComments(tc.issues, tc.comments, tc.worklogs)
			
			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			// Even with errors, should get some kind of summary
			if !tc.expectError && summary == "" {
				t.Error("Expected some summary even with malformed data")
			}
		})
	}
}

// TestSummaryQualityValidation tests the quality of generated summaries
func TestSummaryQualityValidation(t *testing.T) {
	config := LLMConfig{
		Enabled:                  true,
		Mode:                     "embedded",
		Model:                    "test-model",
		SummaryStyle:             "technical",
		MaxSummaryLength:         200,
		IncludeTechnicalDetails:  true,
	}
	llm := NewEmbeddedLLMWithConfig(config)
	
	// Create realistic scenarios
	scenarios := []struct {
		name     string
		issues   []jira.Issue
		comments []jira.Comment
		expectedQualities []string
	}{
		{
			name: "Infrastructure deployment scenario",
			issues: []jira.Issue{
				createInfrastructureIssue("DEVOPS-123", "Deploy AWS Lambda using Terraform", "In Progress", "High"),
				createInfrastructureIssue("DEVOPS-124", "Configure VPC security groups", "Done", "Medium"),
			},
			comments: []jira.Comment{
				createTechnicalComment("1", "Completed Terraform configuration for Lambda deployment"),
				createTechnicalComment("2", "VPC security groups updated and tested in staging"),
			},
			expectedQualities: []string{"technical_terms", "completion_status", "coherent"},
		},
		{
			name: "Database migration scenario",
			issues: []jira.Issue{
				createDatabaseIssue("DB-456", "PostgreSQL migration to version 14", "In Progress", "Critical"),
				createDatabaseIssue("DB-457", "Update connection strings", "Done", "High"),
			},
			comments: []jira.Comment{
				createTechnicalComment("3", "Database migration completed successfully"),
				createTechnicalComment("4", "All microservices updated with new connection strings"),
			},
			expectedQualities: []string{"technical_terms", "completion_status", "coherent"},
		},
	}
	
	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			summary, err := llm.GenerateStandupSummaryWithComments(scenario.issues, scenario.comments, nil)
			if err != nil {
				t.Errorf("Failed to generate summary: %v", err)
				return
			}
			
			if summary == "" {
				t.Error("Empty summary generated")
				return
			}
			
			// Validate expected qualities
			for _, quality := range scenario.expectedQualities {
				if !validateSummaryHasQuality(summary, quality) {
					t.Errorf("Summary lacks expected quality '%s': %s", quality, summary)
				}
			}
			
			t.Logf("Generated summary: %s", summary)
		})
	}
}

// Helper functions for creating realistic test data

func createRealisticDevOpsData() ([]jira.Issue, []jira.Comment, []jira.WorklogEntry) {
	issues := []jira.Issue{
		createInfrastructureIssue("DEVOPS-101", "Deploy microservice to AWS ECS using Terraform", "In Progress", "High"),
		createInfrastructureIssue("DEVOPS-102", "Configure Kubernetes ingress for new service", "Done", "Medium"),
		createDatabaseIssue("DEVOPS-103", "PostgreSQL database migration for user service", "In Progress", "Critical"),
		createSecurityIssue("DEVOPS-104", "Implement OAuth 2.0 authentication flow", "Code Review", "High"),
		createBugIssue("DEVOPS-105", "Fix memory leak in payment processing service", "Done", "Critical"),
	}
	
	comments := []jira.Comment{
		createTechnicalComment("1", "Completed Terraform configuration for ECS deployment. Service is now running in staging environment."),
		createTechnicalComment("2", "Kubernetes ingress controller configured with SSL termination. Load balancing working correctly."),
		createTechnicalComment("3", "Database migration scripts tested in development. Ready for production deployment."),
		createTechnicalComment("4", "OAuth implementation completed. JWT tokens working with refresh mechanism."),
		createTechnicalComment("5", "Memory leak identified in payment processor. Fix deployed and monitoring shows stable memory usage."),
		createTechnicalComment("6", "All integration tests passing. Code review feedback addressed."),
		createTechnicalComment("7", "Production deployment scheduled for tomorrow morning. Rollback plan prepared."),
	}
	
	worklogs := []jira.WorklogEntry{
		{
			IssueID: "DEVOPS-101",
			Started: jira.JiraTime{Time: time.Now().Add(-2 * time.Hour)},
			Comment: "Worked on Terraform configuration",
		},
		{
			IssueID: "DEVOPS-102",
			Started: jira.JiraTime{Time: time.Now().Add(-1 * time.Hour)},
			Comment: "Configured Kubernetes ingress",
		},
		{
			IssueID: "DEVOPS-103",
			Started: jira.JiraTime{Time: time.Now().Add(-30 * time.Minute)},
			Comment: "Database migration testing",
		},
	}
	
	return issues, comments, worklogs
}

func createTestIssues(count int) []jira.Issue {
	issues := make([]jira.Issue, count)
	for i := 0; i < count; i++ {
		issues[i] = createInfrastructureIssue(
			fmt.Sprintf("TEST-%d", i+1),
			fmt.Sprintf("Test issue %d with AWS and Terraform", i+1),
			"In Progress",
			"Medium",
		)
	}
	return issues
}

func createTestComments(count int) []jira.Comment {
	comments := make([]jira.Comment, count)
	for i := 0; i < count; i++ {
		comments[i] = createTechnicalComment(
			fmt.Sprintf("%d", i+1),
			fmt.Sprintf("Completed technical work on infrastructure component %d", i+1),
		)
	}
	return comments
}

func createTestWorklogs(count int) []jira.WorklogEntry {
	worklogs := make([]jira.WorklogEntry, count)
	for i := 0; i < count; i++ {
		worklogs[i] = jira.WorklogEntry{
			IssueID: fmt.Sprintf("TEST-%d", i+1),
			Started: jira.JiraTime{Time: time.Now().Add(-time.Duration(i) * time.Hour)},
			Comment: fmt.Sprintf("Worked on test issue %d", i+1),
		}
	}
	return worklogs
}

func createInfrastructureIssue(key, summary, status, priority string) jira.Issue {
	return jira.Issue{
		Key: key,
		Fields: jira.Fields{
			Summary:     summary,
			Description: jira.JiraDescription{Text: "Infrastructure work involving AWS services, Terraform configuration, and Kubernetes deployment."},
			Status:      jira.Status{Name: status},
			Priority:    jira.Priority{Name: priority},
			IssueType:   jira.IssueType{Name: "Task"},
			Project:     jira.Project{Key: "DEVOPS", Name: "DevOps Team"},
			Created:     jira.JiraTime{Time: time.Now().Add(-24 * time.Hour)},
			Updated:     jira.JiraTime{Time: time.Now().Add(-1 * time.Hour)},
		},
	}
}

func createDatabaseIssue(key, summary, status, priority string) jira.Issue {
	return jira.Issue{
		Key: key,
		Fields: jira.Fields{
			Summary:     summary,
			Description: jira.JiraDescription{Text: "Database migration work involving PostgreSQL, schema updates, and data migration scripts."},
			Status:      jira.Status{Name: status},
			Priority:    jira.Priority{Name: priority},
			IssueType:   jira.IssueType{Name: "Task"},
			Project:     jira.Project{Key: "DEVOPS", Name: "DevOps Team"},
			Created:     jira.JiraTime{Time: time.Now().Add(-48 * time.Hour)},
			Updated:     jira.JiraTime{Time: time.Now().Add(-2 * time.Hour)},
		},
	}
}

func createSecurityIssue(key, summary, status, priority string) jira.Issue {
	return jira.Issue{
		Key: key,
		Fields: jira.Fields{
			Summary:     summary,
			Description: jira.JiraDescription{Text: "Security implementation involving OAuth 2.0, JWT tokens, and authentication flows."},
			Status:      jira.Status{Name: status},
			Priority:    jira.Priority{Name: priority},
			IssueType:   jira.IssueType{Name: "Story"},
			Project:     jira.Project{Key: "DEVOPS", Name: "DevOps Team"},
			Created:     jira.JiraTime{Time: time.Now().Add(-72 * time.Hour)},
			Updated:     jira.JiraTime{Time: time.Now().Add(-30 * time.Minute)},
		},
	}
}

func createBugIssue(key, summary, status, priority string) jira.Issue {
	return jira.Issue{
		Key: key,
		Fields: jira.Fields{
			Summary:     summary,
			Description: jira.JiraDescription{Text: "Critical bug fix involving memory management and performance optimization."},
			Status:      jira.Status{Name: status},
			Priority:    jira.Priority{Name: priority},
			IssueType:   jira.IssueType{Name: "Bug"},
			Project:     jira.Project{Key: "DEVOPS", Name: "DevOps Team"},
			Created:     jira.JiraTime{Time: time.Now().Add(-12 * time.Hour)},
			Updated:     jira.JiraTime{Time: time.Now().Add(-15 * time.Minute)},
		},
	}
}

func createTechnicalComment(id, text string) jira.Comment {
	return jira.Comment{
		ID:      id,
		Body:    jira.JiraDescription{Text: text},
		Created: jira.JiraTime{Time: time.Now().Add(-time.Duration(len(id)) * time.Hour)},
		Updated: jira.JiraTime{Time: time.Now().Add(-time.Duration(len(id)) * time.Hour)},
	}
}

// Helper functions for validation

func containsTechnicalTerms(text string) bool {
	technicalTerms := []string{
		"terraform", "aws", "kubernetes", "docker", "postgresql", "oauth",
		"api", "microservice", "deployment", "infrastructure", "database",
		"security", "ssl", "jwt", "migration", "configuration",
	}
	
	lowerText := strings.ToLower(text)
	for _, term := range technicalTerms {
		if strings.Contains(lowerText, term) {
			return true
		}
	}
	return false
}

func validateSummaryQuality(t *testing.T, summary string, config LLMConfig, summaryType string) {
	// Check length constraints
	if len(summary) > config.MaxSummaryLength {
		t.Errorf("%s summary exceeds max length: %d > %d", summaryType, len(summary), config.MaxSummaryLength)
	}
	
	// Check for empty or too short summaries
	if len(summary) < 10 {
		t.Errorf("%s summary too short: %s", summaryType, summary)
	}
	
	// Check for generic/meaningless content
	genericPhrases := []string{
		"no recent activity",
		"multiple development activities",
		"recent activity:",
	}
	
	lowerSummary := strings.ToLower(summary)
	for _, phrase := range genericPhrases {
		if strings.Contains(lowerSummary, phrase) && len(summary) < 50 {
			t.Logf("Warning: %s summary appears generic: %s", summaryType, summary)
		}
	}
}

func validateSummaryHasQuality(summary, quality string) bool {
	lowerSummary := strings.ToLower(summary)
	
	switch quality {
	case "technical_terms":
		return containsTechnicalTerms(summary)
	case "completion_status":
		statusWords := []string{"completed", "done", "finished", "in progress", "working", "deployed"}
		for _, word := range statusWords {
			if strings.Contains(lowerSummary, word) {
				return true
			}
		}
		return false
	case "coherent":
		// Basic coherence check - should have reasonable length and structure
		return len(summary) > 20 && len(strings.Fields(summary)) > 3
	default:
		return true
	}
}