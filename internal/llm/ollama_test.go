package llm

import (
	"testing"
	"time"
	"my-day/internal/jira"
)

// TestNewOllamaClientWithConfig tests the configuration-based constructor
func TestNewOllamaClientWithConfig(t *testing.T) {
	config := LLMConfig{
		Enabled:                  true,
		Mode:                     "ollama",
		Model:                    "test-model",
		Debug:                    true,
		SummaryStyle:             "technical",
		MaxSummaryLength:         150,
		IncludeTechnicalDetails:  true,
		PrioritizeRecentWork:     true,
		FallbackStrategy:         "graceful",
		OllamaURL:                "http://localhost:11434",
		OllamaModel:              "llama3.1",
	}
	
	client := NewOllamaClientWithConfig(config)
	
	if client == nil {
		t.Fatal("Expected Ollama client instance, got nil")
	}
	
	if client.model != "llama3.1" {
		t.Errorf("Expected model 'llama3.1', got '%s'", client.model)
	}
	
	if client.baseURL != "http://localhost:11434" {
		t.Errorf("Expected baseURL 'http://localhost:11434', got '%s'", client.baseURL)
	}
	
	if client.config == nil {
		t.Fatal("Expected config to be stored, got nil")
	}
	
	if client.config.MaxSummaryLength != 150 {
		t.Errorf("Expected MaxSummaryLength 150, got %d", client.config.MaxSummaryLength)
	}
}

// TestOllamaErrorTypes tests the OllamaError type and its methods
func TestOllamaErrorTypes(t *testing.T) {
	testCases := []struct {
		name          string
		ollamaError   *OllamaError
		expectedError string
	}{
		{
			name: "Connection error without cause",
			ollamaError: &OllamaError{
				Type:    "connection_error",
				Message: "Failed to connect to Ollama service",
			},
			expectedError: "connection_error: Failed to connect to Ollama service",
		},
		{
			name: "Timeout error with cause",
			ollamaError: &OllamaError{
				Type:    "timeout_error",
				Message: "Request timed out",
				Cause:   &OllamaError{Type: "network", Message: "network timeout"},
			},
			expectedError: "timeout_error: Request timed out (caused by: network: network timeout)",
		},
		{
			name: "API error with details",
			ollamaError: &OllamaError{
				Type:    "api_error",
				Message: "API returned error",
				Details: map[string]interface{}{
					"status_code": 404,
					"model":       "missing-model",
				},
			},
			expectedError: "api_error: API returned error",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errorStr := tc.ollamaError.Error()
			if errorStr != tc.expectedError {
				t.Errorf("Expected error '%s', got '%s'", tc.expectedError, errorStr)
			}
		})
	}
}

// TestRetryableErrorDetection tests the retry logic for different error types
func TestRetryableErrorDetection(t *testing.T) {
	config := LLMConfig{
		OllamaURL:   "http://localhost:11434",
		OllamaModel: "test-model",
	}
	client := NewOllamaClientWithConfig(config)
	
	testCases := []struct {
		name           string
		error          error
		shouldRetry    bool
		shouldFallback bool
	}{
		{
			name: "Connection error - should retry and fallback",
			error: &OllamaError{
				Type:    "connection_error",
				Message: "Failed to connect",
			},
			shouldRetry:    true,
			shouldFallback: true,
		},
		{
			name: "Timeout error - should retry and fallback",
			error: &OllamaError{
				Type:    "timeout_error",
				Message: "Request timed out",
			},
			shouldRetry:    true,
			shouldFallback: true,
		},
		{
			name: "Server error (500) - should retry and fallback",
			error: &OllamaError{
				Type:    "api_error",
				Message: "Server error",
				Details: map[string]interface{}{
					"status_code": 500,
				},
			},
			shouldRetry:    true,
			shouldFallback: true,
		},
		{
			name: "Client error (404) - should not retry but no fallback",
			error: &OllamaError{
				Type:    "api_error",
				Message: "Model not found",
				Details: map[string]interface{}{
					"status_code": 404,
				},
			},
			shouldRetry:    false,
			shouldFallback: false,
		},
		{
			name: "Marshal error - should not retry or fallback",
			error: &OllamaError{
				Type:    "marshal_error",
				Message: "Failed to marshal request",
			},
			shouldRetry:    false,
			shouldFallback: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shouldRetry := client.isRetryableError(tc.error)
			if shouldRetry != tc.shouldRetry {
				t.Errorf("Expected retry %t, got %t", tc.shouldRetry, shouldRetry)
			}
			
			shouldFallback := client.shouldFallbackToEmbedded(tc.error)
			if shouldFallback != tc.shouldFallback {
				t.Errorf("Expected fallback %t, got %t", tc.shouldFallback, shouldFallback)
			}
		})
	}
}

// TestEnhancedErrorMessages tests user-friendly error message generation
func TestEnhancedErrorMessages(t *testing.T) {
	config := LLMConfig{
		OllamaURL:   "http://localhost:11434",
		OllamaModel: "test-model",
	}
	client := NewOllamaClientWithConfig(config)
	
	testCases := []struct {
		name            string
		error           error
		retries         int
		expectedContains []string
	}{
		{
			name: "Connection error",
			error: &OllamaError{
				Type:    "connection_error",
				Message: "Failed to connect",
			},
			retries: 3,
			expectedContains: []string{
				"unable to connect to Ollama service",
				"http://localhost:11434",
				"after 4 retries",
				"ollama serve",
			},
		},
		{
			name: "Timeout error",
			error: &OllamaError{
				Type:    "timeout_error",
				Message: "Request timed out",
			},
			retries: 2,
			expectedContains: []string{
				"timed out after 3 retries",
				"test-model",
				"smaller model",
			},
		},
		{
			name: "Model not found error",
			error: &OllamaError{
				Type: "api_error",
				Message: "Model not found",
				Details: map[string]interface{}{
					"status_code": 404,
				},
			},
			retries: 1,
			expectedContains: []string{
				"model 'test-model' not found",
				"ollama pull test-model",
			},
		},
		{
			name: "Server error",
			error: &OllamaError{
				Type: "api_error",
				Message: "Internal server error",
				Details: map[string]interface{}{
					"status_code": 500,
				},
			},
			retries: 2,
			expectedContains: []string{
				"server error after 3 retries",
				"overloaded",
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			enhancedErr := client.enhanceErrorMessage(tc.error, tc.retries)
			errorMsg := enhancedErr.Error()
			
			for _, expected := range tc.expectedContains {
				if !containsSubstring(errorMsg, expected) {
					t.Errorf("Expected error message to contain '%s', got: %s", expected, errorMsg)
				}
			}
		})
	}
}

// TestPromptStyleGeneration tests different prompt styles
func TestPromptStyleGeneration(t *testing.T) {
	testCases := []struct {
		name           string
		config         LLMConfig
		expectedStyle  string
		shouldContain  []string
	}{
		{
			name: "Technical style prompt",
			config: LLMConfig{
				SummaryStyle:             "technical",
				IncludeTechnicalDetails:  true,
				MaxSummaryLength:         200,
			},
			expectedStyle: "technical",
			shouldContain: []string{
				"DevOps team standup",
				"technical implementation",
				"Infrastructure:",
				"Terraform",
				"AWS services",
				"Kubernetes",
			},
		},
		{
			name: "Business style prompt",
			config: LLMConfig{
				SummaryStyle:             "business",
				IncludeTechnicalDetails:  false,
				MaxSummaryLength:         150,
			},
			expectedStyle: "business",
			shouldContain: []string{
				"business stakeholder",
				"deliverables",
				"business impact",
				"project milestones",
				"business value",
			},
		},
		{
			name: "Brief style prompt",
			config: LLMConfig{
				SummaryStyle:             "brief",
				IncludeTechnicalDetails:  false,
				MaxSummaryLength:         100,
			},
			expectedStyle: "brief",
			shouldContain: []string{
				"very brief",
				"concise",
				"most important",
				"high-impact activities",
			},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewOllamaClientWithConfig(tc.config)
			
			// Test with sample data
			issues := []jira.Issue{
				{
					Key: "TEST-123",
					Fields: jira.Fields{
						Summary:     "Deploy AWS Lambda function",
						Description: jira.JiraDescription{Text: "Configure Terraform for production deployment"},
						Status:      jira.Status{Name: "In Progress"},
						Priority:    jira.Priority{Name: "High"},
						IssueType:   jira.IssueType{Name: "Task"},
						Project:     jira.Project{Key: "TEST"},
					},
				},
			}
			
			comments := []jira.Comment{
				{
					ID: "1",
					Body: jira.JiraDescription{Text: "Completed infrastructure setup using Terraform"},
					Created: jira.JiraTime{Time: time.Now()},
				},
			}
			
			prompt := client.buildEnhancedStandupPrompt(issues, comments, nil)
			
			for _, expected := range tc.shouldContain {
				if !containsSubstring(prompt, expected) {
					t.Errorf("Expected prompt to contain '%s' for %s style", expected, tc.expectedStyle)
				}
			}
		})
	}
}

// TestOllamaTechnicalTermExtraction tests technical term extraction in prompts
func TestOllamaTechnicalTermExtraction(t *testing.T) {
	config := LLMConfig{
		IncludeTechnicalDetails: true,
	}
	client := NewOllamaClientWithConfig(config)
	
	testCases := []struct {
		name          string
		text          string
		expectedTerms []string
	}{
		{
			name: "AWS and Terraform terms",
			text: "Deploy AWS Lambda function using Terraform and configure VPC",
			expectedTerms: []string{"terraform", "aws"},
		},
		{
			name: "Database and API terms",
			text: "Update PostgreSQL database schema and REST API endpoints",
			expectedTerms: []string{"postgresql", "api", "rest"},
		},
		{
			name: "Kubernetes and Docker terms",
			text: "Fix Kubernetes deployment issues with Docker containers",
			expectedTerms: []string{"kubernetes", "docker"},
		},
		{
			name: "CI/CD and monitoring terms",
			text: "Setup CI/CD pipeline with monitoring and logging",
			expectedTerms: []string{"ci/cd", "monitoring", "logging"},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			terms := client.extractTechnicalTerms(tc.text)
			
			for _, expected := range tc.expectedTerms {
				found := false
				for _, term := range terms {
					if term == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find term '%s' in extracted terms %v", expected, terms)
				}
			}
		})
	}
}

// TestActivityTypeDetection tests activity type detection from comments
func TestOllamaActivityTypeDetection(t *testing.T) {
	config := LLMConfig{}
	client := NewOllamaClientWithConfig(config)
	
	testCases := []struct {
		name         string
		text         string
		expectedType string
	}{
		{
			name:         "Completed work",
			text:         "Completed the database migration successfully",
			expectedType: "✅ Completed",
		},
		{
			name:         "Deployment activity",
			text:         "Deployed the new version to production environment",
			expectedType: "🚀 Deployed",
		},
		{
			name:         "Blocked work",
			text:         "Blocked waiting for security team approval",
			expectedType: "🚫 Blocked",
		},
		{
			name:         "Testing activity",
			text:         "Running integration tests for the new API",
			expectedType: "🧪 Testing",
		},
		{
			name:         "Investigation work",
			text:         "Investigating the performance issues in production",
			expectedType: "🔍 Investigating",
		},
		{
			name:         "Active work",
			text:         "Working on implementing the authentication service",
			expectedType: "⚙️ Working",
		},
		{
			name:         "General update",
			text:         "Updated the documentation with new examples",
			expectedType: "📝 Update",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			activityType := client.determineActivityType(tc.text)
			if activityType != tc.expectedType {
				t.Errorf("Expected activity type '%s', got '%s'", tc.expectedType, activityType)
			}
		})
	}
}

// TestPriorityAndTypeEmojis tests emoji generation for priorities and issue types
func TestPriorityAndTypeEmojis(t *testing.T) {
	config := LLMConfig{}
	client := NewOllamaClientWithConfig(config)
	
	priorityTests := []struct {
		priority string
		expected string
	}{
		{"Critical", "🔥"},
		{"Highest", "🔥"},
		{"High", "⚡"},
		{"Medium", "📋"},
		{"Low", "📝"},
		{"Lowest", "📝"},
		{"Unknown", "📋"},
	}
	
	for _, tt := range priorityTests {
		t.Run("Priority_"+tt.priority, func(t *testing.T) {
			emoji := client.getPriorityEmoji(tt.priority)
			if emoji != tt.expected {
				t.Errorf("Expected emoji '%s' for priority '%s', got '%s'", tt.expected, tt.priority, emoji)
			}
		})
	}
	
	typeTests := []struct {
		issueType string
		expected  string
	}{
		{"Bug", "🐛 Bug Fix"},
		{"Feature", "✨ Feature"},
		{"Story", "✨ Feature"},
		{"Task", "📋 Task"},
		{"Epic", "🎯 Epic"},
		{"Improvement", "🔧 Improvement"},
		{"Unknown", "📋 Work"},
	}
	
	for _, tt := range typeTests {
		t.Run("Type_"+tt.issueType, func(t *testing.T) {
			context := client.getIssueTypeContext(tt.issueType)
			if context != tt.expected {
				t.Errorf("Expected context '%s' for type '%s', got '%s'", tt.expected, tt.issueType, context)
			}
		})
	}
}

// TestFallbackToEmbedded tests the fallback mechanism
func TestFallbackToEmbedded(t *testing.T) {
	config := LLMConfig{
		OllamaURL:   "http://localhost:11434",
		OllamaModel: "test-model",
		Model:       "embedded-model",
	}
	client := NewOllamaClientWithConfig(config)
	
	// Test fallback creation
	embeddedLLM := client.fallbackToEmbedded()
	if embeddedLLM == nil {
		t.Fatal("Expected embedded LLM instance, got nil")
	}
	
	// Test that config is passed to embedded LLM
	if embeddedLLM.config == nil {
		t.Error("Expected config to be passed to embedded LLM")
	} else if embeddedLLM.config.Model != config.Model {
		t.Errorf("Expected embedded LLM model '%s', got '%s'", config.Model, embeddedLLM.config.Model)
	}
}

// Helper function to check if a string contains a substring
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || 
		 (len(s) > len(substr) && 
		  (s[:len(substr)] == substr || 
		   s[len(s)-len(substr):] == substr || 
		   containsSubstring(s[1:], substr))))
}