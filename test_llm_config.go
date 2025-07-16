package main

import (
	"fmt"
	"my-day/internal/jira"
	"my-day/internal/llm"
)

func main() {
	// Test different LLM configurations
	testConfigurations()
}

func testConfigurations() {
	// Create a test issue
	testIssue := jira.Issue{
		Key: "TEST-123",
		Fields: jira.Fields{
			Summary:     "Fix authentication timeout in user login API with AWS Lambda deployment",
			Description: jira.JiraDescription{Text: "Users are experiencing timeouts when logging in through the API. Investigation shows the OAuth token validation is taking too long. Need to deploy fix to production environment using Terraform."},
			Status: jira.Status{
				Name: "In Progress",
			},
			Priority: jira.Priority{
				Name: "High",
			},
			IssueType: jira.IssueType{
				Name: "Bug",
			},
			Project: jira.Project{
				Key:  "TEST",
				Name: "Test Project",
			},
		},
	}

	// Test configuration 1: Technical style with full details
	fmt.Println("=== Configuration 1: Technical Style with Full Details ===")
	config1 := llm.LLMConfig{
		Enabled:                  true,
		Mode:                     "embedded",
		Model:                    "tinyllama",
		Debug:                    true,
		SummaryStyle:             "technical",
		MaxSummaryLength:         200,
		IncludeTechnicalDetails:  true,
		PrioritizeRecentWork:     true,
		FallbackStrategy:         "graceful",
	}
	
	llm1 := llm.NewEmbeddedLLMWithConfig(config1)
	summary1, err := llm1.SummarizeIssue(testIssue)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Summary: %s\n", summary1)
	}
	fmt.Println()

	// Test configuration 2: Brief style without technical details
	fmt.Println("=== Configuration 2: Brief Style without Technical Details ===")
	config2 := llm.LLMConfig{
		Enabled:                  true,
		Mode:                     "embedded",
		Model:                    "tinyllama",
		Debug:                    false,
		SummaryStyle:             "brief",
		MaxSummaryLength:         100,
		IncludeTechnicalDetails:  false,
		PrioritizeRecentWork:     true,
		FallbackStrategy:         "graceful",
	}
	
	llm2 := llm.NewEmbeddedLLMWithConfig(config2)
	summary2, err := llm2.SummarizeIssue(testIssue)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Summary: %s\n", summary2)
	}
	fmt.Println()

	// Test configuration 3: Business style with medium length
	fmt.Println("=== Configuration 3: Business Style with Medium Length ===")
	config3 := llm.LLMConfig{
		Enabled:                  true,
		Mode:                     "embedded",
		Model:                    "tinyllama",
		Debug:                    false,
		SummaryStyle:             "business",
		MaxSummaryLength:         150,
		IncludeTechnicalDetails:  true,
		PrioritizeRecentWork:     false,
		FallbackStrategy:         "graceful",
	}
	
	llm3 := llm.NewEmbeddedLLMWithConfig(config3)
	summary3, err := llm3.SummarizeIssue(testIssue)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Summary: %s\n", summary3)
	}
	fmt.Println()

	// Test debug functionality
	fmt.Println("=== Debug Report Test ===")
	debugReport, err := llm1.GetDebugReport()
	if err != nil {
		fmt.Printf("Debug report error: %v\n", err)
	} else if debugReport != nil {
		fmt.Printf("Debug report available with %d steps\n", len(debugReport.Steps))
	} else {
		fmt.Println("No debug report available")
	}
}