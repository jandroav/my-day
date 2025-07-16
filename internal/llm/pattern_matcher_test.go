package llm

import (
	"testing"
)

// TestTechnicalPatternMatcher tests the technical pattern matching functionality
func TestTechnicalPatternMatcher(t *testing.T) {
	matcher := NewTechnicalPatternMatcher(false)
	
	testCases := []struct {
		name            string
		text            string
		expectedPatterns []string
		patternType     string
	}{
		{
			name: "Infrastructure patterns - Terraform and AWS",
			text: "Deploy AWS Lambda function using Terraform configuration",
			expectedPatterns: []string{"terraform", "aws"},
			patternType: "infrastructure",
		},
		{
			name: "Kubernetes deployment patterns",
			text: "Fix Kubernetes pod scaling issues in production cluster",
			expectedPatterns: []string{"kubernetes"},
			patternType: "infrastructure",
		},
		{
			name: "CI/CD pipeline patterns",
			text: "Update Jenkins pipeline for automated deployment to staging",
			expectedPatterns: []string{"deploy"},
			patternType: "deployment",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := matcher.MatchAllPatterns(tc.text)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			// Check that we got some results
			if len(results) == 0 {
				t.Error("Expected some pattern matches but got none")
			}
			
			// Check for specific pattern types
			if tc.patternType == "infrastructure" {
				if infraPatterns, ok := results["infrastructure"]; ok {
					if patterns, ok := infraPatterns.([]InfrastructurePattern); ok {
						if len(patterns) == 0 {
							t.Error("Expected infrastructure patterns but got none")
						}
					}
				}
			}
			
			if tc.patternType == "deployment" {
				if deployPatterns, ok := results["deployment"]; ok {
					if patterns, ok := deployPatterns.([]DeploymentPattern); ok {
						if len(patterns) == 0 {
							t.Error("Expected deployment patterns but got none")
						}
					}
				}
			}
		})
	}
}

// TestInfrastructurePatternMatching tests infrastructure-specific pattern matching
func TestInfrastructurePatternMatching(t *testing.T) {
	matcher := NewTechnicalPatternMatcher(false)
	
	testCases := []struct {
		name         string
		text         string
		expectedType string
		expectedAction string
	}{
		{
			name:         "Terraform deployment",
			text:         "Deploy AWS infrastructure using Terraform",
			expectedType: "terraform",
			expectedAction: "deploy",
		},
		{
			name:         "Kubernetes configuration",
			text:         "Configure Kubernetes cluster for production",
			expectedType: "kubernetes",
			expectedAction: "configure",
		},
		{
			name:         "AWS setup",
			text:         "Setup AWS VPC and security groups",
			expectedType: "aws",
			expectedAction: "setup",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patterns, err := matcher.MatchInfrastructurePatterns(tc.text)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if len(patterns) == 0 {
				t.Error("Expected infrastructure patterns but got none")
			}
			
			// Check for expected type and action
			found := false
			for _, pattern := range patterns {
				if pattern.Type == tc.expectedType {
					found = true
					if pattern.Action != tc.expectedAction {
						t.Errorf("Expected action '%s', got '%s'", tc.expectedAction, pattern.Action)
					}
					break
				}
			}
			
			if !found {
				t.Errorf("Expected to find pattern type '%s'", tc.expectedType)
			}
		})
	}
}

// TestDeploymentPatternMatching tests deployment-specific pattern matching
func TestDeploymentPatternMatching(t *testing.T) {
	matcher := NewTechnicalPatternMatcher(false)
	
	testCases := []struct {
		name               string
		text               string
		expectedType       string
		expectedEnvironment string
	}{
		{
			name:               "Production deployment",
			text:               "Deploy application to production environment",
			expectedType:       "deploy",
			expectedEnvironment: "production",
		},
		{
			name:               "Staging deployment",
			text:               "Release new version to staging",
			expectedType:       "deploy",
			expectedEnvironment: "staging",
		},
		{
			name:               "Development deployment",
			text:               "Deploy changes to development environment",
			expectedType:       "deploy",
			expectedEnvironment: "development",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patterns, err := matcher.MatchDeploymentPatterns(tc.text)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if len(patterns) == 0 {
				t.Error("Expected deployment patterns but got none")
			}
			
			// Check for expected type and environment
			found := false
			for _, pattern := range patterns {
				if pattern.Type == tc.expectedType {
					found = true
					if pattern.Environment != tc.expectedEnvironment {
						t.Errorf("Expected environment '%s', got '%s'", tc.expectedEnvironment, pattern.Environment)
					}
					break
				}
			}
			
			if !found {
				t.Errorf("Expected to find pattern type '%s'", tc.expectedType)
			}
		})
	}
}

// TestDevelopmentPatternMatching tests development-specific pattern matching
func TestDevelopmentPatternMatching(t *testing.T) {
	matcher := NewTechnicalPatternMatcher(false)
	
	testCases := []struct {
		name         string
		text         string
		expectedType string
		expectedAction string
	}{
		{
			name:         "Code review",
			text:         "Create pull request for code review",
			expectedType: "code_review",
			expectedAction: "create",
		},
		{
			name:         "Bug fix",
			text:         "Fix authentication bug in production",
			expectedType: "bug_fix",
			expectedAction: "fix",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			patterns, err := matcher.MatchDevelopmentPatterns(tc.text)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if len(patterns) == 0 {
				t.Error("Expected development patterns but got none")
			}
			
			// Check for expected type and action
			found := false
			for _, pattern := range patterns {
				if pattern.Type == tc.expectedType {
					found = true
					if pattern.Action != tc.expectedAction {
						t.Errorf("Expected action '%s', got '%s'", tc.expectedAction, pattern.Action)
					}
					break
				}
			}
			
			if !found {
				t.Errorf("Expected to find pattern type '%s'", tc.expectedType)
			}
		})
	}
}

// TestPatternMatchingEdgeCases tests edge cases in pattern matching
func TestPatternMatchingEdgeCases(t *testing.T) {
	matcher := NewTechnicalPatternMatcher(false)
	
	testCases := []struct {
		name        string
		text        string
		expectError bool
		expectEmpty bool
	}{
		{
			name:        "Empty text",
			text:        "",
			expectError: false,
			expectEmpty: true,
		},
		{
			name:        "Whitespace only",
			text:        "   \n\t  ",
			expectError: false,
			expectEmpty: true,
		},
		{
			name:        "No technical terms",
			text:        "This is a general business discussion about quarterly goals",
			expectError: false,
			expectEmpty: true,
		},
		{
			name:        "Mixed case technical terms",
			text:        "TERRAFORM deployment with AWS and Kubernetes",
			expectError: false,
			expectEmpty: false,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := matcher.MatchAllPatterns(tc.text)
			
			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			if tc.expectEmpty {
				// Check if all pattern types are empty
				isEmpty := true
				if infraPatterns, ok := results["infrastructure"]; ok {
					if patterns, ok := infraPatterns.([]InfrastructurePattern); ok && len(patterns) > 0 {
						isEmpty = false
					}
				}
				if deployPatterns, ok := results["deployment"]; ok {
					if patterns, ok := deployPatterns.([]DeploymentPattern); ok && len(patterns) > 0 {
						isEmpty = false
					}
				}
				if devPatterns, ok := results["development"]; ok {
					if patterns, ok := devPatterns.([]DevelopmentPattern); ok && len(patterns) > 0 {
						isEmpty = false
					}
				}
				
				if !isEmpty {
					t.Error("Expected empty results but got some patterns")
				}
			}
		})
	}
}

// TestPatternMatcherStatistics tests the statistics functionality
func TestPatternMatcherStatistics(t *testing.T) {
	matcher := NewTechnicalPatternMatcher(false)
	
	stats := matcher.GetPatternStatistics()
	
	if stats == nil {
		t.Fatal("Expected statistics but got nil")
	}
	
	// Check for expected statistics fields
	expectedFields := []string{
		"infrastructure_patterns",
		"deployment_patterns", 
		"development_patterns",
		"total_patterns",
	}
	
	for _, field := range expectedFields {
		if _, ok := stats[field]; !ok {
			t.Errorf("Expected statistics field '%s' but not found", field)
		}
	}
	
	// Check that total patterns is reasonable
	if totalPatterns, ok := stats["total_patterns"].(int); ok {
		if totalPatterns <= 0 {
			t.Error("Expected positive number of total patterns")
		}
	} else {
		t.Error("Expected total_patterns to be an integer")
	}
}

// TestPatternMatcherDebugMode tests debug mode functionality
func TestPatternMatcherDebugMode(t *testing.T) {
	debugMatcher := NewTechnicalPatternMatcher(true)
	normalMatcher := NewTechnicalPatternMatcher(false)
	
	text := "Deploy AWS Lambda using Terraform"
	
	// Test debug mode
	debugResults, err := debugMatcher.MatchAllPatterns(text)
	if err != nil {
		t.Errorf("Unexpected error in debug mode: %v", err)
	}
	
	// Test normal mode
	normalResults, err := normalMatcher.MatchAllPatterns(text)
	if err != nil {
		t.Errorf("Unexpected error in normal mode: %v", err)
	}
	
	// Both should find patterns
	if len(debugResults) == 0 {
		t.Error("Expected patterns in debug mode")
	}
	
	if len(normalResults) == 0 {
		t.Error("Expected patterns in normal mode")
	}
}

// TestRealWorldPatternMatching tests with realistic DevOps scenarios
func TestRealWorldPatternMatching(t *testing.T) {
	matcher := NewTechnicalPatternMatcher(false)
	
	realWorldCases := []struct {
		name        string
		text        string
		expectInfra bool
		expectDeploy bool
		expectDev   bool
	}{
		{
			name: "Infrastructure deployment",
			text: "Deployed new VPC configuration using Terraform. Updated security groups and configured ALB with SSL certificates. Kubernetes cluster is now ready for microservice deployment.",
			expectInfra: true,
			expectDeploy: true,
			expectDev: false,
		},
		{
			name: "CI/CD pipeline setup",
			text: "Configured Jenkins pipeline for automated testing and deployment. Docker images are built and pushed to ECR. Kubernetes manifests updated for rolling deployment strategy.",
			expectInfra: true,
			expectDeploy: true,
			expectDev: false,
		},
		{
			name: "Code review process",
			text: "Created pull request for authentication feature. Code review completed and approved. Ready to merge into main branch.",
			expectInfra: false,
			expectDeploy: false,
			expectDev: true,
		},
	}
	
	for _, tc := range realWorldCases {
		t.Run(tc.name, func(t *testing.T) {
			results, err := matcher.MatchAllPatterns(tc.text)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			
			// Check infrastructure patterns
			if tc.expectInfra {
				if infraPatterns, ok := results["infrastructure"]; ok {
					if patterns, ok := infraPatterns.([]InfrastructurePattern); ok {
						if len(patterns) == 0 {
							t.Error("Expected infrastructure patterns but got none")
						}
					}
				} else {
					t.Error("Expected infrastructure patterns in results")
				}
			}
			
			// Check deployment patterns
			if tc.expectDeploy {
				if deployPatterns, ok := results["deployment"]; ok {
					if patterns, ok := deployPatterns.([]DeploymentPattern); ok {
						if len(patterns) == 0 {
							t.Error("Expected deployment patterns but got none")
						}
					}
				} else {
					t.Error("Expected deployment patterns in results")
				}
			}
			
			// Check development patterns
			if tc.expectDev {
				if devPatterns, ok := results["development"]; ok {
					if patterns, ok := devPatterns.([]DevelopmentPattern); ok {
						if len(patterns) == 0 {
							t.Error("Expected development patterns but got none")
						}
					}
				} else {
					t.Error("Expected development patterns in results")
				}
			}
		})
	}
}