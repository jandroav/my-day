package report

import (
	"strings"
	"testing"
	"time"

	"my-day/internal/jira"
)

func TestHasMeaningfulComments(t *testing.T) {
	tests := []struct {
		name     string
		comments []jira.Comment
		expected bool
	}{
		{
			name:     "No comments",
			comments: []jira.Comment{},
			expected: false,
		},
		{
			name: "Empty comment",
			comments: []jira.Comment{
				{Body: jira.JiraDescription{Text: ""}},
			},
			expected: false,
		},
		{
			name: "Short comment",
			comments: []jira.Comment{
				{Body: jira.JiraDescription{Text: "ok"}},
			},
			expected: false,
		},
		{
			name: "Whitespace only",
			comments: []jira.Comment{
				{Body: jira.JiraDescription{Text: "   \n\t  "}},
			},
			expected: false,
		},
		{
			name: "Meaningful comment",
			comments: []jira.Comment{
				{Body: jira.JiraDescription{Text: "Implemented the authentication feature"}},
			},
			expected: true,
		},
		{
			name: "Mix of short and meaningful comments",
			comments: []jira.Comment{
				{Body: jira.JiraDescription{Text: "ok"}},
				{Body: jira.JiraDescription{Text: "Reviewed the pull request and added suggestions"}},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasMeaningfulComments(tt.comments)
			if result != tt.expected {
				t.Errorf("hasMeaningfulComments() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateConsoleWithCommentsSkipsAI(t *testing.T) {
	// Create a generator with LLM enabled
	config := &Config{
		Format:     "console",
		LLMEnabled: true,
		LLMMode:    "disabled", // Use disabled mode to avoid external dependencies
	}
	
	generator := NewGenerator(config)
	
	// Create test data with non-meaningful comments
	issues := []jira.Issue{
		{
			Key: "TEST-123",
			Fields: jira.Fields{
				Summary: "Test issue",
				Status:  jira.Status{Name: "In Progress"},
				Project: jira.Project{Key: "TEST"},
				Priority: jira.Priority{Name: "Medium"},
				Updated: jira.JiraTime{Time: time.Now()},
			},
		},
	}
	
	commentsMap := map[string][]jira.Comment{
		"TEST-123": {
			{Body: jira.JiraDescription{Text: "ok"}}, // Non-meaningful comment
		},
	}
	
	var worklogs []jira.WorklogEntry
	targetDate := time.Now()
	
	// Generate report
	reportContent, err := generator.generateConsoleWithComments(issues, commentsMap, worklogs, targetDate)
	if err != nil {
		t.Fatalf("generateConsoleWithComments() error = %v", err)
	}
	
	// Check that AI summary is skipped
	if !strings.Contains(reportContent, "⚠️  AI SUMMARY SKIPPED") {
		t.Error("Expected AI summary to be skipped for non-meaningful comments")
	}
	
	if strings.Contains(reportContent, "🤖 AI SUMMARY OF TODAY'S WORK") {
		t.Error("AI summary should not be generated for non-meaningful comments")
	}
}

func TestGenerateConsoleWithCommentsGeneratesAI(t *testing.T) {
	// Create a generator with LLM enabled
	config := &Config{
		Format:     "console",
		LLMEnabled: true,
		LLMMode:    "disabled", // Use disabled mode to avoid external dependencies
	}
	
	generator := NewGenerator(config)
	
	// Create test data with meaningful comments
	issues := []jira.Issue{
		{
			Key: "TEST-123",
			Fields: jira.Fields{
				Summary: "Test issue",
				Status:  jira.Status{Name: "In Progress"},
				Project: jira.Project{Key: "TEST"},
				Priority: jira.Priority{Name: "Medium"},
				Updated: jira.JiraTime{Time: time.Now()},
			},
		},
	}
	
	commentsMap := map[string][]jira.Comment{
		"TEST-123": {
			{Body: jira.JiraDescription{Text: "Implemented the authentication feature and fixed the login bug"}}, // Meaningful comment
		},
	}
	
	var worklogs []jira.WorklogEntry
	targetDate := time.Now()
	
	// Generate report
	reportContent, err := generator.generateConsoleWithComments(issues, commentsMap, worklogs, targetDate)
	if err != nil {
		t.Fatalf("generateConsoleWithComments() error = %v", err)
	}
	
	// Check that AI summary is NOT skipped
	if strings.Contains(reportContent, "⚠️  AI SUMMARY SKIPPED") {
		t.Error("AI summary should not be skipped for meaningful comments")
	}
	
	// The disabled summarizer should still generate a basic summary
	if !strings.Contains(reportContent, "Recent activity:") {
		t.Error("Expected some form of AI summary to be generated for meaningful comments")
	}
}