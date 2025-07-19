package report

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"my-day/internal/jira"
	"my-day/internal/llm"
)

// IssueWithComments represents an issue with today's comments
type IssueWithComments struct {
	Issue    jira.Issue     `json:"issue"`
	Comments []jira.Comment `json:"comments"`
}

// Generator handles report generation
type Generator struct {
	config     *Config
	summarizer llm.Summarizer
}

// Config represents report generation configuration
type Config struct {
	Format            string
	LLMEnabled        bool
	LLMMode           string
	LLMModel          string
	OllamaURL         string
	OllamaModel       string
	IncludeYesterday  bool
	IncludeToday      bool
	IncludeInProgress bool
	Detailed          bool
	Debug             bool
	ShowQuality       bool
	Verbose           bool
	GroupByField      string
	ExportEnabled     bool
	ExportFolderPath  string
	ExportFileDate    string
	ExportTags        []string
}

// NewGenerator creates a new report generator
func NewGenerator(config *Config) *Generator {
	// Initialize LLM summarizer based on configuration
	llmConfig := llm.LLMConfig{
		Enabled:                  config.LLMEnabled,
		Mode:                     config.LLMMode,
		Model:                    config.LLMModel,
		Debug:                    config.Debug,
		SummaryStyle:             "technical", // Default to technical style for DevOps context
		MaxSummaryLength:         200,
		IncludeTechnicalDetails:  true,
		PrioritizeRecentWork:     true,
		FallbackStrategy:         "graceful",
		OllamaURL:                config.OllamaURL,
		OllamaModel:              config.OllamaModel,
	}
	
	summarizer, err := llm.NewSummarizer(llmConfig)
	if err != nil {
		// Fallback to disabled summarizer if initialization fails
		summarizer = llm.NewDisabledSummarizer()
	}
	
	return &Generator{
		config:     config,
		summarizer: summarizer,
	}
}

// Generate creates a daily standup report
func (g *Generator) Generate(issues []jira.Issue, worklogs []jira.WorklogEntry, targetDate time.Time) (string, error) {
	// Filter issues based on configuration and target date
	filteredIssues := g.filterIssues(issues, targetDate)
	filteredWorklogs := g.filterWorklogs(worklogs, targetDate)

	switch g.config.Format {
	case "markdown":
		return g.generateMarkdown(filteredIssues, filteredWorklogs, targetDate)
	default:
		return g.generateConsole(filteredIssues, filteredWorklogs, targetDate)
	}
}

// GenerateWithComments creates a daily standup report with comment summaries
func (g *Generator) GenerateWithComments(issuesWithComments []IssueWithComments, worklogs []jira.WorklogEntry, targetDate time.Time) (string, error) {
	// Extract just the issues for filtering
	var issues []jira.Issue
	for _, iwc := range issuesWithComments {
		issues = append(issues, iwc.Issue)
	}

	// Filter issues and worklogs
	filteredIssues := g.filterIssues(issues, targetDate)
	filteredWorklogs := g.filterWorklogs(worklogs, targetDate)

	// Create a map of issue key to comments for quick lookup
	commentsMap := make(map[string][]jira.Comment)
	for _, iwc := range issuesWithComments {
		commentsMap[iwc.Issue.Key] = iwc.Comments
	}

	if g.config.GroupByField != "" {
		return g.generateFieldGroupedReport(filteredIssues, commentsMap, filteredWorklogs, targetDate, g.config.GroupByField)
	}

	switch g.config.Format {
	case "markdown":
		return g.generateMarkdownWithComments(filteredIssues, commentsMap, filteredWorklogs, targetDate)
	default:
		return g.generateConsoleWithComments(filteredIssues, commentsMap, filteredWorklogs, targetDate)
	}
}

func (g *Generator) filterIssues(issues []jira.Issue, targetDate time.Time) []jira.Issue {
	var filtered []jira.Issue
	
	today := targetDate.Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)

	for _, issue := range issues {
		issueDate := issue.Fields.Updated.Time.Truncate(24 * time.Hour)
		
		include := false
		
		// Check if issue should be included based on date
		if g.config.IncludeToday && issueDate.Equal(today) {
			include = true
		}
		if g.config.IncludeYesterday && issueDate.Equal(yesterday) {
			include = true
		}
		
		// Always include in-progress issues if configured
		if g.config.IncludeInProgress && isInProgress(issue) {
			include = true
		}

		if include {
			filtered = append(filtered, issue)
		}
	}

	// Sort by priority and last updated
	sort.Slice(filtered, func(i, j int) bool {
		// First sort by status category (In Progress > To Do > Done)
		iCategory := getStatusCategory(filtered[i])
		jCategory := getStatusCategory(filtered[j])
		
		if iCategory != jCategory {
			return iCategory < jCategory
		}
		
		// Then by update time (most recent first)
		return filtered[i].Fields.Updated.Time.After(filtered[j].Fields.Updated.Time)
	})

	return filtered
}

func (g *Generator) filterWorklogs(worklogs []jira.WorklogEntry, targetDate time.Time) []jira.WorklogEntry {
	var filtered []jira.WorklogEntry
	
	today := targetDate.Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)

	for _, worklog := range worklogs {
		worklogDate := worklog.Started.Time.Truncate(24 * time.Hour)
		
		include := false
		if g.config.IncludeToday && worklogDate.Equal(today) {
			include = true
		}
		if g.config.IncludeYesterday && worklogDate.Equal(yesterday) {
			include = true
		}

		if include {
			filtered = append(filtered, worklog)
		}
	}

	// Sort by start time (most recent first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Started.Time.After(filtered[j].Started.Time)
	})

	return filtered
}

func (g *Generator) generateConsole(issues []jira.Issue, worklogs []jira.WorklogEntry, targetDate time.Time) (string, error) {
	var report strings.Builder
	
	// Header
	report.WriteString(fmt.Sprintf("🚀 Daily Standup Report - %s\n", targetDate.Format("January 2, 2006")))
	report.WriteString(strings.Repeat("=", 50) + "\n")
	report.WriteString("📝 Issues with your comments today\n\n")

	// AI Summary if enabled
	if g.config.LLMEnabled {
		standupSummary, err := g.summarizer.GenerateStandupSummary(issues, worklogs)
		if err == nil && standupSummary != "" {
			report.WriteString("🤖 AI SUMMARY\n")
			report.WriteString(fmt.Sprintf("%s\n\n", standupSummary))
		}
	}

	// Summary
	report.WriteString("📊 SUMMARY\n")
	report.WriteString(fmt.Sprintf("• Issues with comments today: %d\n", len(issues)))
	report.WriteString(fmt.Sprintf("• Worklog entries: %d\n", len(worklogs)))
	report.WriteString("\n")

	// Group issues by status
	statusGroups := groupIssuesByStatus(issues)
	
	// In Progress section
	if inProgress, exists := statusGroups["In Progress"]; exists && len(inProgress) > 0 {
		report.WriteString("🔄 CURRENTLY WORKING ON\n")
		for _, issue := range inProgress {
			report.WriteString(g.formatIssueConsole(issue))
		}
		report.WriteString("\n")
	}

	// Recently completed section
	if done, exists := statusGroups["Done"]; exists && len(done) > 0 {
		report.WriteString("✅ RECENTLY COMPLETED\n")
		for _, issue := range done {
			report.WriteString(g.formatIssueConsole(issue))
		}
		report.WriteString("\n")
	}

	// To Do section
	if todo, exists := statusGroups["To Do"]; exists && len(todo) > 0 {
		report.WriteString("📋 TO DO\n")
		for _, issue := range todo {
			report.WriteString(g.formatIssueConsole(issue))
		}
		report.WriteString("\n")
	}

	// Worklog section
	if len(worklogs) > 0 {
		report.WriteString("⏰ WORK LOG\n")
		for _, worklog := range worklogs {
			report.WriteString(g.formatWorklogConsole(worklog))
		}
		report.WriteString("\n")
	}

	// Footer
	report.WriteString("---\n")
	report.WriteString("Generated by my-day CLI 🤖\n")

	return report.String(), nil
}

func (g *Generator) generateConsoleWithComments(issues []jira.Issue, commentsMap map[string][]jira.Comment, worklogs []jira.WorklogEntry, targetDate time.Time) (string, error) {
	var report strings.Builder
	
	// Header
	report.WriteString(fmt.Sprintf("🚀 Daily Standup Report - %s\n", targetDate.Format("January 2, 2006")))
	report.WriteString(strings.Repeat("=", 50) + "\n")
	report.WriteString("📝 Issues with your comments today\n\n")

	// AI Summary if enabled - based on comments
	if g.config.LLMEnabled {
		allComments := []jira.Comment{}
		for _, comments := range commentsMap {
			allComments = append(allComments, comments...)
		}
		
		if hasMeaningfulComments(allComments) {
			// Use the enhanced LLM method for intelligent summary
			summary, err := g.summarizer.GenerateStandupSummaryWithComments(issues, allComments, worklogs)
			if err == nil && summary != "" {
				report.WriteString("🤖 AI SUMMARY OF TODAY'S WORK\n")
				report.WriteString(fmt.Sprintf("%s\n\n", summary))
			}
		} else if len(allComments) > 0 {
			// Show warning when there are comments but they're not meaningful enough for AI summary
			report.WriteString("⚠️  AI SUMMARY SKIPPED\n")
			report.WriteString("No meaningful comment content found for AI summarization.\n")
			report.WriteString("Consider adding more detailed comments to your Jira tickets for better AI insights.\n\n")
		}
	}

	// Summary
	report.WriteString("📊 SUMMARY\n")
	report.WriteString(fmt.Sprintf("• Issues with comments today: %d\n", len(issues)))
	
	totalComments := 0
	for _, comments := range commentsMap {
		totalComments += len(comments)
	}
	report.WriteString(fmt.Sprintf("• Total comments added: %d\n", totalComments))
	report.WriteString(fmt.Sprintf("• Worklog entries: %d\n", len(worklogs)))
	report.WriteString("\n")

	// Group issues by status
	statusGroups := groupIssuesByStatus(issues)
	
	// In Progress section
	if inProgress, exists := statusGroups["In Progress"]; exists && len(inProgress) > 0 {
		report.WriteString("🔄 CURRENTLY WORKING ON\n")
		for _, issue := range inProgress {
			report.WriteString(g.formatIssueConsoleWithComments(issue, commentsMap[issue.Key]))
		}
		report.WriteString("\n")
	}

	// Recently completed section
	if done, exists := statusGroups["Done"]; exists && len(done) > 0 {
		report.WriteString("✅ RECENTLY COMPLETED\n")
		for _, issue := range done {
			report.WriteString(g.formatIssueConsoleWithComments(issue, commentsMap[issue.Key]))
		}
		report.WriteString("\n")
	}

	// To Do section
	if todo, exists := statusGroups["To Do"]; exists && len(todo) > 0 {
		report.WriteString("📋 TO DO\n")
		for _, issue := range todo {
			report.WriteString(g.formatIssueConsoleWithComments(issue, commentsMap[issue.Key]))
		}
		report.WriteString("\n")
	}

	// Worklog section
	if len(worklogs) > 0 {
		report.WriteString("⏰ WORK LOG\n")
		for _, worklog := range worklogs {
			report.WriteString(g.formatWorklogConsole(worklog))
		}
		report.WriteString("\n")
	}

	// Footer
	report.WriteString("---\n")
	report.WriteString("Generated by my-day CLI 🤖\n")

	return report.String(), nil
}

func (g *Generator) generateMarkdown(issues []jira.Issue, worklogs []jira.WorklogEntry, targetDate time.Time) (string, error) {
	var report strings.Builder
	
	// Header
	report.WriteString(fmt.Sprintf("# Daily Standup Report - %s\n\n", targetDate.Format("January 2, 2006")))
	report.WriteString("*Issues with your comments today*\n\n")

	// AI Summary if enabled
	if g.config.LLMEnabled {
		standupSummary, err := g.summarizer.GenerateStandupSummary(issues, worklogs)
		if err == nil && standupSummary != "" {
			report.WriteString("## 🤖 AI Summary\n\n")
			report.WriteString(fmt.Sprintf("%s\n\n", standupSummary))
		}
	}

	// Summary
	report.WriteString("## Summary\n\n")
	report.WriteString(fmt.Sprintf("- **Issues with comments today**: %d\n", len(issues)))
	report.WriteString(fmt.Sprintf("- **Worklog entries**: %d\n\n", len(worklogs)))

	// Group issues by status
	statusGroups := groupIssuesByStatus(issues)
	
	// In Progress section
	if inProgress, exists := statusGroups["In Progress"]; exists && len(inProgress) > 0 {
		report.WriteString("## 🔄 Currently Working On\n\n")
		for _, issue := range inProgress {
			report.WriteString(g.formatIssueMarkdown(issue))
		}
		report.WriteString("\n")
	}

	// Recently completed section
	if done, exists := statusGroups["Done"]; exists && len(done) > 0 {
		report.WriteString("## ✅ Recently Completed\n\n")
		for _, issue := range done {
			report.WriteString(g.formatIssueMarkdown(issue))
		}
		report.WriteString("\n")
	}

	// To Do section
	if todo, exists := statusGroups["To Do"]; exists && len(todo) > 0 {
		report.WriteString("## 📋 To Do\n\n")
		for _, issue := range todo {
			report.WriteString(g.formatIssueMarkdown(issue))
		}
		report.WriteString("\n")
	}

	// Worklog section
	if len(worklogs) > 0 {
		report.WriteString("## ⏰ Work Log\n\n")
		for _, worklog := range worklogs {
			report.WriteString(g.formatWorklogMarkdown(worklog))
		}
		report.WriteString("\n")
	}

	// Footer
	report.WriteString("---\n")
	report.WriteString("*Generated by my-day CLI*\n")

	return report.String(), nil
}

func (g *Generator) formatIssueConsole(issue jira.Issue) string {
	var result strings.Builder
	
	statusIcon := getStatusIcon(issue.Fields.Status.Name)
	priorityIcon := getPriorityIcon(issue.Fields.Priority.Name)
	
	result.WriteString(fmt.Sprintf("  %s %s [%s] %s\n", 
		statusIcon, 
		issue.Key, 
		issue.Fields.Project.Key,
		issue.Fields.Summary))
	
	// Add AI summary if enabled and detailed mode
	if g.config.LLMEnabled && g.config.Detailed {
		if summary, err := g.summarizer.SummarizeIssue(issue); err == nil && summary != "" {
			result.WriteString(fmt.Sprintf("    🤖 %s\n", summary))
		}
	}
	
	if g.config.Detailed {
		result.WriteString(fmt.Sprintf("    Priority: %s %s | Status: %s\n", 
			priorityIcon,
			issue.Fields.Priority.Name,
			issue.Fields.Status.Name))
		result.WriteString(fmt.Sprintf("    Updated: %s\n", 
			issue.Fields.Updated.Time.Format("Jan 2, 15:04")))
		
		if issue.Fields.Description.Text != "" {
			result.WriteString(fmt.Sprintf("    %s\n", issue.Fields.Description.Text))
		}
	}
	
	result.WriteString("\n")
	return result.String()
}

func (g *Generator) formatIssueMarkdown(issue jira.Issue) string {
	statusIcon := getStatusIcon(issue.Fields.Status.Name)
	priorityIcon := getPriorityIcon(issue.Fields.Priority.Name)
	
	result := fmt.Sprintf("- %s **[%s]** %s\n", statusIcon, issue.Key, issue.Fields.Summary)
	
	// Add AI summary if enabled and detailed mode
	if g.config.LLMEnabled && g.config.Detailed {
		if summary, err := g.summarizer.SummarizeIssue(issue); err == nil && summary != "" {
			result += fmt.Sprintf("  - 🤖 **AI Summary**: %s\n", summary)
		}
	}
	
	if g.config.Detailed {
		result += fmt.Sprintf("  - Priority: %s %s\n", priorityIcon, issue.Fields.Priority.Name)
		result += fmt.Sprintf("  - Status: %s\n", issue.Fields.Status.Name)
		result += fmt.Sprintf("  - Updated: %s\n", issue.Fields.Updated.Time.Format("Jan 2, 15:04"))
		
		if issue.Fields.Description.Text != "" {
			result += fmt.Sprintf("  - %s\n", issue.Fields.Description.Text)
		}
	}
	
	result += "\n"
	return result
}

func (g *Generator) formatWorklogConsole(worklog jira.WorklogEntry) string {
	result := fmt.Sprintf("  ⏱️  [%s] %s\n", 
		worklog.IssueID,
		worklog.Started.Time.Format("Jan 2, 15:04"))
	
	if worklog.Comment != "" {
		result += fmt.Sprintf("    %s\n", worklog.Comment)
	}
	
	result += "\n"
	return result
}

func (g *Generator) formatWorklogMarkdown(worklog jira.WorklogEntry) string {
	result := fmt.Sprintf("- ⏱️ **[%s]** %s\n", 
		worklog.IssueID,
		worklog.Started.Time.Format("Jan 2, 15:04"))
	
	if worklog.Comment != "" {
		result += fmt.Sprintf("  - %s\n", worklog.Comment)
	}
	
	result += "\n"
	return result
}

// Helper functions

func isInProgress(issue jira.Issue) bool {
	category := getStatusCategory(issue)
	return category == 1 // In Progress category
}

func getStatusCategory(issue jira.Issue) int {
	switch strings.ToLower(issue.Fields.Status.Category.Key) {
	case "indeterminate":
		return 1 // In Progress
	case "new":
		return 2 // To Do
	case "done":
		return 3 // Done
	default:
		return 2 // Default to To Do
	}
}

func groupIssuesByStatus(issues []jira.Issue) map[string][]jira.Issue {
	groups := make(map[string][]jira.Issue)
	
	for _, issue := range issues {
		statusCategory := issue.Fields.Status.Category.Key
		
		var groupName string
		switch strings.ToLower(statusCategory) {
		case "indeterminate":
			groupName = "In Progress"
		case "new":
			groupName = "To Do"
		case "done":
			groupName = "Done"
		default:
			groupName = "Other"
		}
		
		groups[groupName] = append(groups[groupName], issue)
	}
	
	return groups
}

func getStatusIcon(status string) string {
	switch strings.ToLower(status) {
	case "in progress", "in-progress":
		return "🔄"
	case "done", "closed", "resolved":
		return "✅"
	case "to do", "todo", "open", "new":
		return "📋"
	case "blocked":
		return "🚫"
	case "review", "code review":
		return "👀"
	default:
		return "📝"
	}
}

func getPriorityIcon(priority string) string {
	switch strings.ToLower(priority) {
	case "highest", "critical":
		return "🔴"
	case "high":
		return "🟠"
	case "medium":
		return "🟡"
	case "low":
		return "🟢"
	case "lowest":
		return "🔵"
	default:
		return "⚪"
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// hasMeaningfulComments checks if there are any non-empty, meaningful comments
func hasMeaningfulComments(comments []jira.Comment) bool {
	if len(comments) == 0 {
		return false
	}
	
	// Check if any comment has non-empty, meaningful text
	for _, comment := range comments {
		text := strings.TrimSpace(comment.Body.Text)
		// Consider a comment meaningful if it has at least 3 characters
		// (to filter out very short comments like "ok", "done", etc.)
		if len(text) > 3 {
			return true
		}
	}
	
	return false
}

func (g *Generator) formatIssueConsoleWithComments(issue jira.Issue, comments []jira.Comment) string {
	var result strings.Builder
	
	statusIcon := getStatusIcon(issue.Fields.Status.Name)
	priorityIcon := getPriorityIcon(issue.Fields.Priority.Name)
	
	result.WriteString(fmt.Sprintf("  %s %s [%s] %s\n", 
		statusIcon, 
		issue.Key, 
		issue.Fields.Project.Key,
		issue.Fields.Summary))
	
	// Add comment summary if enabled
	if g.config.LLMEnabled && len(comments) > 0 {
		if summary, err := g.summarizer.SummarizeComments(comments); err == nil && summary != "" {
			result.WriteString(fmt.Sprintf("    💬 Today's work: %s\n", summary))
		}
	}
	
	if g.config.Detailed {
		result.WriteString(fmt.Sprintf("    Priority: %s %s | Status: %s\n", 
			priorityIcon,
			issue.Fields.Priority.Name,
			issue.Fields.Status.Name))
		result.WriteString(fmt.Sprintf("    Updated: %s\n", 
			issue.Fields.Updated.Time.Format("Jan 2, 15:04")))
		
		// Show comment count and latest comment
		if len(comments) > 0 {
			result.WriteString(fmt.Sprintf("    Comments today: %d\n", len(comments)))
			if len(comments) > 0 {
				latestComment := comments[len(comments)-1]
				// Show full comment text without truncation
				result.WriteString(fmt.Sprintf("    Latest: %s\n", latestComment.Body.Text))
			}
		}
	}
	
	result.WriteString("\n")
	return result.String()
}

func (g *Generator) generateMarkdownWithComments(issues []jira.Issue, commentsMap map[string][]jira.Comment, worklogs []jira.WorklogEntry, targetDate time.Time) (string, error) {
	var report strings.Builder
	
	// Header
	report.WriteString(fmt.Sprintf("# Daily Standup Report - %s\n\n", targetDate.Format("January 2, 2006")))
	report.WriteString("*Issues with your comments today*\n\n")

	// AI Summary if enabled - based on comments
	if g.config.LLMEnabled {
		allComments := []jira.Comment{}
		for _, comments := range commentsMap {
			allComments = append(allComments, comments...)
		}
		
		if hasMeaningfulComments(allComments) {
			// Use the enhanced LLM method for intelligent summary
			summary, err := g.summarizer.GenerateStandupSummaryWithComments(issues, allComments, worklogs)
			if err == nil && summary != "" {
				report.WriteString("## 🤖 AI Summary of Today's Work\n\n")
				report.WriteString(fmt.Sprintf("%s\n\n", summary))
			}
		} else if len(allComments) > 0 {
			// Show warning when there are comments but they're not meaningful enough for AI summary
			report.WriteString("## ⚠️ AI Summary Skipped\n\n")
			report.WriteString("No meaningful comment content found for AI summarization.\n")
			report.WriteString("Consider adding more detailed comments to your Jira tickets for better AI insights.\n\n")
		}
	}

	// Summary
	report.WriteString("## Summary\n\n")
	report.WriteString(fmt.Sprintf("- **Issues with comments today**: %d\n", len(issues)))
	
	totalComments := 0
	for _, comments := range commentsMap {
		totalComments += len(comments)
	}
	report.WriteString(fmt.Sprintf("- **Total comments added**: %d\n", totalComments))
	report.WriteString(fmt.Sprintf("- **Worklog entries**: %d\n\n", len(worklogs)))

	// Group issues by status
	statusGroups := groupIssuesByStatus(issues)
	
	// In Progress section
	if inProgress, exists := statusGroups["In Progress"]; exists && len(inProgress) > 0 {
		report.WriteString("## 🔄 Currently Working On\n\n")
		for _, issue := range inProgress {
			report.WriteString(g.formatIssueMarkdownWithComments(issue, commentsMap[issue.Key]))
		}
		report.WriteString("\n")
	}

	// Recently completed section
	if done, exists := statusGroups["Done"]; exists && len(done) > 0 {
		report.WriteString("## ✅ Recently Completed\n\n")
		for _, issue := range done {
			report.WriteString(g.formatIssueMarkdownWithComments(issue, commentsMap[issue.Key]))
		}
		report.WriteString("\n")
	}

	// To Do section
	if todo, exists := statusGroups["To Do"]; exists && len(todo) > 0 {
		report.WriteString("## 📋 To Do\n\n")
		for _, issue := range todo {
			report.WriteString(g.formatIssueMarkdownWithComments(issue, commentsMap[issue.Key]))
		}
		report.WriteString("\n")
	}

	// Worklog section
	if len(worklogs) > 0 {
		report.WriteString("## ⏰ Work Log\n\n")
		for _, worklog := range worklogs {
			report.WriteString(g.formatWorklogMarkdown(worklog))
		}
		report.WriteString("\n")
	}

	// Footer
	report.WriteString("---\n")
	report.WriteString("*Generated by my-day CLI*\n")

	return report.String(), nil
}

func (g *Generator) formatIssueMarkdownWithComments(issue jira.Issue, comments []jira.Comment) string {
	statusIcon := getStatusIcon(issue.Fields.Status.Name)
	priorityIcon := getPriorityIcon(issue.Fields.Priority.Name)
	
	result := fmt.Sprintf("- %s **[%s]** %s\n", statusIcon, issue.Key, issue.Fields.Summary)
	
	// Add comment summary if enabled
	if g.config.LLMEnabled && len(comments) > 0 {
		if summary, err := g.summarizer.SummarizeComments(comments); err == nil && summary != "" {
			result += fmt.Sprintf("  - 💬 **Today's work**: %s\n", summary)
		}
	}
	
	if g.config.Detailed {
		result += fmt.Sprintf("  - Priority: %s %s\n", priorityIcon, issue.Fields.Priority.Name)
		result += fmt.Sprintf("  - Status: %s\n", issue.Fields.Status.Name)
		result += fmt.Sprintf("  - Updated: %s\n", issue.Fields.Updated.Time.Format("Jan 2, 15:04"))
		
		// Show comment count and latest comment
		if len(comments) > 0 {
			result += fmt.Sprintf("  - Comments today: %d\n", len(comments))
			if len(comments) > 0 {
				latestComment := comments[len(comments)-1]
				// Show full comment text without truncation
				result += fmt.Sprintf("  - Latest comment: %s\n", latestComment.Body.Text)
			}
		}
	}
	
	result += "\n"
	return result
}

// GenerateWithEnhancedContext creates a report using enhanced LLM processing with additional context
func (g *Generator) GenerateWithEnhancedContext(issuesWithComments []IssueWithComments, worklogs []jira.WorklogEntry, targetDate time.Time) (string, error) {
	// Extract just the issues for filtering
	var issues []jira.Issue
	var allComments []jira.Comment
	for _, iwc := range issuesWithComments {
		issues = append(issues, iwc.Issue)
		allComments = append(allComments, iwc.Comments...)
	}

	// Filter issues and worklogs
	filteredIssues := g.filterIssues(issues, targetDate)
	filteredWorklogs := g.filterWorklogs(worklogs, targetDate)

	// Create a map of issue key to comments for quick lookup
	commentsMap := make(map[string][]jira.Comment)
	for _, iwc := range issuesWithComments {
		commentsMap[iwc.Issue.Key] = iwc.Comments
	}

	// Pass additional context to LLM if enabled
	if g.config.LLMEnabled {
		// Prepare enhanced context for LLM processing
		enhancedContext := g.prepareEnhancedContext(filteredIssues, allComments, filteredWorklogs, targetDate)
		
		// Pass context to LLM summarizer if it supports enhanced context
		if contextualSummarizer, ok := g.summarizer.(interface{ SetEnhancedContext(map[string]interface{}) error }); ok {
			if err := contextualSummarizer.SetEnhancedContext(enhancedContext); err != nil && g.config.Debug {
				// Log error but continue processing
				fmt.Printf("Warning: Failed to set enhanced context: %v\n", err)
			}
		}
	}

	// Generate the main report
	var reportContent string
	var err error
	
	switch g.config.Format {
	case "markdown":
		reportContent, err = g.generateMarkdownWithEnhancedContext(filteredIssues, commentsMap, filteredWorklogs, targetDate)
	default:
		reportContent, err = g.generateConsoleWithEnhancedContext(filteredIssues, commentsMap, filteredWorklogs, targetDate)
	}
	
	if err != nil {
		return "", err
	}

	// Add debug information if requested
	if g.config.Debug || g.config.Verbose {
		debugInfo, debugErr := g.generateDebugInformation()
		if debugErr == nil && debugInfo != "" {
			reportContent += "\n" + debugInfo
		}
	}

	return reportContent, nil
}

// prepareEnhancedContext prepares enhanced context for LLM processing
func (g *Generator) prepareEnhancedContext(filteredIssues []jira.Issue, allComments []jira.Comment, filteredWorklogs []jira.WorklogEntry, targetDate time.Time) map[string]interface{} {
	enhancedContext := make(map[string]interface{})
	
	// Basic context information
	enhancedContext["target_date"] = targetDate.Format("2006-01-02")
	enhancedContext["issue_count"] = len(filteredIssues)
	enhancedContext["comment_count"] = len(allComments)
	enhancedContext["worklog_count"] = len(filteredWorklogs)
	
	// Issue status distribution
	statusCounts := make(map[string]int)
	for _, issue := range filteredIssues {
		statusCounts[issue.Fields.Status.Name]++
	}
	enhancedContext["status_distribution"] = statusCounts
	
	// Priority distribution
	priorityCounts := make(map[string]int)
	for _, issue := range filteredIssues {
		priorityCounts[issue.Fields.Priority.Name]++
	}
	enhancedContext["priority_distribution"] = priorityCounts
	
	// Recent activity timeline
	var recentActivities []map[string]interface{}
	for _, comment := range allComments {
		activity := map[string]interface{}{
			"timestamp": comment.Created.Time,
			"type":      "comment",
			"content":   comment.Body.Text,
		}
		recentActivities = append(recentActivities, activity)
	}
	
	for _, worklog := range filteredWorklogs {
		activity := map[string]interface{}{
			"timestamp": worklog.Started.Time,
			"type":      "worklog",
			"content":   worklog.Comment,
		}
		recentActivities = append(recentActivities, activity)
	}
	
	enhancedContext["recent_activities"] = recentActivities
	
	// Technical context hints
	var technicalTerms []string
	for _, comment := range allComments {
		text := strings.ToLower(comment.Body.Text)
		terms := []string{"terraform", "aws", "kubernetes", "database", "api", "deployment", "security", "testing"}
		for _, term := range terms {
			if strings.Contains(text, term) {
				technicalTerms = append(technicalTerms, term)
			}
		}
	}
	enhancedContext["technical_terms"] = technicalTerms
	
	return enhancedContext
}

// generateConsoleWithEnhancedContext generates console report with enhanced LLM context
func (g *Generator) generateConsoleWithEnhancedContext(issues []jira.Issue, commentsMap map[string][]jira.Comment, worklogs []jira.WorklogEntry, targetDate time.Time) (string, error) {
	var report strings.Builder
	
	// Header
	report.WriteString(fmt.Sprintf("🚀 Daily Standup Report - %s\n", targetDate.Format("January 2, 2006")))
	report.WriteString(strings.Repeat("=", 50) + "\n")
	report.WriteString("📝 Issues with your comments today (Enhanced Analysis)\n\n")

	// AI Summary if enabled - with enhanced processing
	if g.config.LLMEnabled {
		allComments := []jira.Comment{}
		for _, comments := range commentsMap {
			allComments = append(allComments, comments...)
		}
		
		if hasMeaningfulComments(allComments) {
			// Use enhanced data processor for better analysis
			processor := llm.NewEnhancedDataProcessor(g.config.Debug)
			processedData, err := processor.ProcessIssuesWithComments(issues, allComments)
			
			if err == nil && processedData != nil {
				// Generate enhanced summary using processed data
				summary := processedData.GetSummary()
				keyActivities := processedData.GetKeyActivities()
				
				report.WriteString("🤖 AI SUMMARY OF TODAY'S WORK (Enhanced)\n")
				report.WriteString(fmt.Sprintf("%s\n", summary))
				
				if len(keyActivities) > 0 {
					report.WriteString("🔑 Key Activities:\n")
					for _, activity := range keyActivities {
						report.WriteString(fmt.Sprintf("  • %s\n", activity))
					}
				}
				report.WriteString("\n")
				
				// Add quality indicators if enabled
				if g.config.ShowQuality {
					qualityInfo := g.generateSummaryQualityIndicators(summary, len(issues), len(allComments))
					if qualityInfo != "" {
						report.WriteString(qualityInfo)
						report.WriteString("\n")
					}
				}
			} else {
				// Fallback to standard summary generation
				summary, err := g.summarizer.GenerateStandupSummaryWithComments(issues, allComments, worklogs)
				if err == nil && summary != "" {
					report.WriteString("🤖 AI SUMMARY OF TODAY'S WORK\n")
					report.WriteString(fmt.Sprintf("%s\n\n", summary))
				}
			}
		} else if len(allComments) > 0 {
			// Show warning when there are comments but they're not meaningful enough for AI summary
			report.WriteString("⚠️  AI SUMMARY SKIPPED (Enhanced Mode)\n")
			report.WriteString("No meaningful comment content found for AI summarization.\n")
			report.WriteString("Consider adding more detailed comments to your Jira tickets for better AI insights.\n\n")
		}
	}

	// Summary with enhanced metrics
	report.WriteString("📊 SUMMARY\n")
	report.WriteString(fmt.Sprintf("• Issues with comments today: %d\n", len(issues)))
	
	totalComments := 0
	for _, comments := range commentsMap {
		totalComments += len(comments)
	}
	report.WriteString(fmt.Sprintf("• Total comments added: %d\n", totalComments))
	report.WriteString(fmt.Sprintf("• Worklog entries: %d\n", len(worklogs)))
	
	// Add technical context summary if available
	if g.config.LLMEnabled {
		allComments := []jira.Comment{}
		for _, comments := range commentsMap {
			allComments = append(allComments, comments...)
		}
		
		processor := llm.NewEnhancedDataProcessor(g.config.Debug)
		if processedData, err := processor.ProcessIssuesWithComments(issues, allComments); err == nil && processedData != nil {
			if processedData.TechnicalContext != nil && len(processedData.TechnicalContext.Technologies) > 0 {
				report.WriteString(fmt.Sprintf("• Technologies involved: %s\n", 
					strings.Join(processedData.TechnicalContext.Technologies[:min(5, len(processedData.TechnicalContext.Technologies))], ", ")))
			}
		}
	}
	report.WriteString("\n")

	// Group issues by status
	statusGroups := groupIssuesByStatus(issues)
	
	// In Progress section
	if inProgress, exists := statusGroups["In Progress"]; exists && len(inProgress) > 0 {
		report.WriteString("🔄 CURRENTLY WORKING ON\n")
		for _, issue := range inProgress {
			report.WriteString(g.formatIssueConsoleWithComments(issue, commentsMap[issue.Key]))
		}
		report.WriteString("\n")
	}

	// Recently completed section
	if done, exists := statusGroups["Done"]; exists && len(done) > 0 {
		report.WriteString("✅ RECENTLY COMPLETED\n")
		for _, issue := range done {
			report.WriteString(g.formatIssueConsoleWithComments(issue, commentsMap[issue.Key]))
		}
		report.WriteString("\n")
	}

	// To Do section
	if todo, exists := statusGroups["To Do"]; exists && len(todo) > 0 {
		report.WriteString("📋 TO DO\n")
		for _, issue := range todo {
			report.WriteString(g.formatIssueConsoleWithComments(issue, commentsMap[issue.Key]))
		}
		report.WriteString("\n")
	}

	// Worklog section
	if len(worklogs) > 0 {
		report.WriteString("⏰ WORK LOG\n")
		for _, worklog := range worklogs {
			report.WriteString(g.formatWorklogConsole(worklog))
		}
		report.WriteString("\n")
	}

	// Footer
	report.WriteString("---\n")
	report.WriteString("Generated by my-day CLI 🤖 (Enhanced Mode)\n")

	return report.String(), nil
}

// generateMarkdownWithEnhancedContext generates markdown report with enhanced LLM context
func (g *Generator) generateMarkdownWithEnhancedContext(issues []jira.Issue, commentsMap map[string][]jira.Comment, worklogs []jira.WorklogEntry, targetDate time.Time) (string, error) {
	var report strings.Builder
	
	// Header
	report.WriteString(fmt.Sprintf("# Daily Standup Report - %s\n\n", targetDate.Format("January 2, 2006")))
	report.WriteString("*Issues with your comments today (Enhanced Analysis)*\n\n")

	// AI Summary if enabled - with enhanced processing
	if g.config.LLMEnabled {
		allComments := []jira.Comment{}
		for _, comments := range commentsMap {
			allComments = append(allComments, comments...)
		}
		
		if hasMeaningfulComments(allComments) {
			// Use enhanced data processor for better analysis
			processor := llm.NewEnhancedDataProcessor(g.config.Debug)
			processedData, err := processor.ProcessIssuesWithComments(issues, allComments)
			
			if err == nil && processedData != nil {
				// Generate enhanced summary using processed data
				summary := processedData.GetSummary()
				keyActivities := processedData.GetKeyActivities()
				
				report.WriteString("## 🤖 AI Summary of Today's Work (Enhanced)\n\n")
				report.WriteString(fmt.Sprintf("%s\n\n", summary))
				
				if len(keyActivities) > 0 {
					report.WriteString("### 🔑 Key Activities\n\n")
					for _, activity := range keyActivities {
						report.WriteString(fmt.Sprintf("- %s\n", activity))
					}
					report.WriteString("\n")
				}
				
				// Add quality indicators if enabled
				if g.config.ShowQuality {
					qualityInfo := g.generateSummaryQualityIndicators(summary, len(issues), len(allComments))
					if qualityInfo != "" {
						report.WriteString("### 📊 Summary Quality Indicators\n\n")
						report.WriteString("```\n")
						report.WriteString(qualityInfo)
						report.WriteString("```\n\n")
					}
				}
			} else {
				// Fallback to standard summary generation
				summary, err := g.summarizer.GenerateStandupSummaryWithComments(issues, allComments, worklogs)
				if err == nil && summary != "" {
					report.WriteString("## 🤖 AI Summary of Today's Work\n\n")
					report.WriteString(fmt.Sprintf("%s\n\n", summary))
				}
			}
		} else if len(allComments) > 0 {
			// Show warning when there are comments but they're not meaningful enough for AI summary
			report.WriteString("## ⚠️ AI Summary Skipped (Enhanced Mode)\n\n")
			report.WriteString("No meaningful comment content found for AI summarization.\n")
			report.WriteString("Consider adding more detailed comments to your Jira tickets for better AI insights.\n\n")
		}
	}

	// Summary with enhanced metrics
	report.WriteString("## Summary\n\n")
	report.WriteString(fmt.Sprintf("- **Issues with comments today**: %d\n", len(issues)))
	
	totalComments := 0
	for _, comments := range commentsMap {
		totalComments += len(comments)
	}
	report.WriteString(fmt.Sprintf("- **Total comments added**: %d\n", totalComments))
	report.WriteString(fmt.Sprintf("- **Worklog entries**: %d\n", len(worklogs)))
	
	// Add technical context summary if available
	if g.config.LLMEnabled {
		allComments := []jira.Comment{}
		for _, comments := range commentsMap {
			allComments = append(allComments, comments...)
		}
		
		processor := llm.NewEnhancedDataProcessor(g.config.Debug)
		if processedData, err := processor.ProcessIssuesWithComments(issues, allComments); err == nil && processedData != nil {
			if processedData.TechnicalContext != nil && len(processedData.TechnicalContext.Technologies) > 0 {
				report.WriteString(fmt.Sprintf("- **Technologies involved**: %s\n", 
					strings.Join(processedData.TechnicalContext.Technologies[:min(5, len(processedData.TechnicalContext.Technologies))], ", ")))
			}
		}
	}
	report.WriteString("\n")

	// Group issues by status
	statusGroups := groupIssuesByStatus(issues)
	
	// In Progress section
	if inProgress, exists := statusGroups["In Progress"]; exists && len(inProgress) > 0 {
		report.WriteString("## 🔄 Currently Working On\n\n")
		for _, issue := range inProgress {
			report.WriteString(g.formatIssueMarkdownWithComments(issue, commentsMap[issue.Key]))
		}
		report.WriteString("\n")
	}

	// Recently completed section
	if done, exists := statusGroups["Done"]; exists && len(done) > 0 {
		report.WriteString("## ✅ Recently Completed\n\n")
		for _, issue := range done {
			report.WriteString(g.formatIssueMarkdownWithComments(issue, commentsMap[issue.Key]))
		}
		report.WriteString("\n")
	}

	// To Do section
	if todo, exists := statusGroups["To Do"]; exists && len(todo) > 0 {
		report.WriteString("## 📋 To Do\n\n")
		for _, issue := range todo {
			report.WriteString(g.formatIssueMarkdownWithComments(issue, commentsMap[issue.Key]))
		}
		report.WriteString("\n")
	}

	// Worklog section
	if len(worklogs) > 0 {
		report.WriteString("## ⏰ Work Log\n\n")
		for _, worklog := range worklogs {
			report.WriteString(g.formatWorklogMarkdown(worklog))
		}
		report.WriteString("\n")
	}

	// Footer
	report.WriteString("---\n")
	report.WriteString("*Generated by my-day CLI (Enhanced Mode)*\n")

	return report.String(), nil
}

// generateDebugInformation creates debug output for LLM processing
func (g *Generator) generateDebugInformation() (string, error) {
	if !g.config.LLMEnabled {
		return "", nil
	}

	var debugOutput strings.Builder
	
	debugOutput.WriteString("\n" + strings.Repeat("=", 50) + "\n")
	debugOutput.WriteString("🔍 LLM DEBUG INFORMATION\n")
	debugOutput.WriteString(strings.Repeat("=", 50) + "\n")

	// Basic LLM configuration info
	debugOutput.WriteString("Configuration:\n")
	debugOutput.WriteString(fmt.Sprintf("  • LLM Mode: %s\n", g.config.LLMMode))
	debugOutput.WriteString(fmt.Sprintf("  • Model: %s\n", g.config.LLMModel))
	if g.config.LLMMode == "ollama" {
		debugOutput.WriteString(fmt.Sprintf("  • Ollama URL: %s\n", g.config.OllamaURL))
		debugOutput.WriteString(fmt.Sprintf("  • Ollama Model: %s\n", g.config.OllamaModel))
	}
	debugOutput.WriteString(fmt.Sprintf("  • Debug Mode: %t\n", g.config.Debug))
	debugOutput.WriteString(fmt.Sprintf("  • Verbose Mode: %t\n", g.config.Verbose))
	debugOutput.WriteString(fmt.Sprintf("  • Show Quality: %t\n", g.config.ShowQuality))
	debugOutput.WriteString("\n")

	// Try to get debug report from LLM if it supports it
	// This is a type assertion to check if the summarizer supports debug reporting
	if debuggable, ok := g.summarizer.(interface{ GetDebugReport() (*llm.DebugReport, error) }); ok {
		report, err := debuggable.GetDebugReport()
		if err == nil && report != nil {
			debugOutput.WriteString("LLM Processing Report:\n")
			debugOutput.WriteString(fmt.Sprintf("  • Session ID: %s\n", report.SessionID))
			debugOutput.WriteString(fmt.Sprintf("  • Start Time: %s\n", report.StartTime.Format("15:04:05")))
			debugOutput.WriteString(fmt.Sprintf("  • End Time: %s\n", report.EndTime.Format("15:04:05")))
			debugOutput.WriteString(fmt.Sprintf("  • Total Duration: %v\n", report.TotalDuration))
			debugOutput.WriteString(fmt.Sprintf("  • Processing Steps: %d\n", len(report.Steps)))
			debugOutput.WriteString(fmt.Sprintf("  • Successful Steps: %d\n", report.Summary.SuccessfulSteps))
			debugOutput.WriteString(fmt.Sprintf("  • Failed Steps: %d\n", report.Summary.FailedSteps))
			debugOutput.WriteString(fmt.Sprintf("  • Success Rate: %.1f%%\n", float64(report.Summary.SuccessfulSteps)/float64(report.Summary.TotalSteps)*100))
			debugOutput.WriteString(fmt.Sprintf("  • Quality Score: %.1f/100\n", report.Summary.QualityScore*100))
			
			if len(report.Warnings) > 0 {
				debugOutput.WriteString(fmt.Sprintf("  • Warnings: %d\n", len(report.Warnings)))
				for _, warning := range report.Warnings {
					debugOutput.WriteString(fmt.Sprintf("    - %s: %s\n", warning.Type, warning.Message))
				}
			}
			
			if len(report.Summary.Recommendations) > 0 {
				debugOutput.WriteString("  • Recommendations:\n")
				for _, rec := range report.Summary.Recommendations {
					debugOutput.WriteString(fmt.Sprintf("    - %s\n", rec))
				}
			}
			
			if g.config.Verbose && len(report.Steps) > 0 {
				debugOutput.WriteString("\nDetailed Processing Steps:\n")
				for i, step := range report.Steps {
					debugOutput.WriteString(fmt.Sprintf("  %d. %s\n", i+1, step.Step))
					debugOutput.WriteString(fmt.Sprintf("     Time: %s (Duration: %v)\n", step.Timestamp.Format("15:04:05.000"), step.Duration))
					if step.Error != "" {
						debugOutput.WriteString(fmt.Sprintf("     ❌ Error: %s\n", step.Error))
					} else {
						debugOutput.WriteString("     ✅ Success\n")
					}
					if step.OutputData != nil && g.config.Verbose {
						debugOutput.WriteString(fmt.Sprintf("     Output: %v\n", step.OutputData))
					}
				}
			}
		} else if err != nil {
			debugOutput.WriteString(fmt.Sprintf("❌ Failed to get debug report: %v\n", err))
		}
	} else {
		debugOutput.WriteString("⚠️  LLM summarizer does not support debug reporting\n")
		debugOutput.WriteString("   (This is normal for some LLM implementations)\n")
	}

	// Additional debug information about the report generation process
	debugOutput.WriteString("\nReport Generation Info:\n")
	debugOutput.WriteString(fmt.Sprintf("  • Report Format: %s\n", g.config.Format))
	debugOutput.WriteString(fmt.Sprintf("  • Include Yesterday: %t\n", g.config.IncludeYesterday))
	debugOutput.WriteString(fmt.Sprintf("  • Include Today: %t\n", g.config.IncludeToday))
	debugOutput.WriteString(fmt.Sprintf("  • Include In Progress: %t\n", g.config.IncludeInProgress))
	debugOutput.WriteString(fmt.Sprintf("  • Detailed Mode: %t\n", g.config.Detailed))

	return debugOutput.String(), nil
}

// generateSummaryQualityIndicators creates quality metrics for the generated summary
func (g *Generator) generateSummaryQualityIndicators(summary string, issueCount int, commentCount int) string {
	if !g.config.ShowQuality {
		return ""
	}

	var quality strings.Builder
	
	quality.WriteString("\n📊 SUMMARY QUALITY INDICATORS\n")
	quality.WriteString(strings.Repeat("-", 30) + "\n")

	// Calculate basic quality metrics
	summaryLength := len(summary)
	_ = len(strings.Fields(summary))
	
	// Quality scoring (simple heuristic)
	var qualityScore float64 = 0
	var qualityFactors []string

	// Length appropriateness (50-300 characters is good)
	if summaryLength >= 50 && summaryLength <= 300 {
		qualityScore += 25
		qualityFactors = append(qualityFactors, "✓ Appropriate length")
	} else if summaryLength < 50 {
		qualityFactors = append(qualityFactors, "⚠ Summary might be too brief")
	} else {
		qualityFactors = append(qualityFactors, "⚠ Summary might be too verbose")
	}

	// Content richness (more than just counts)
	if !strings.Contains(summary, "issues") || !strings.Contains(summary, "comments") {
		qualityScore += 25
		qualityFactors = append(qualityFactors, "✓ Contains meaningful content")
	} else {
		qualityFactors = append(qualityFactors, "⚠ May be too generic")
	}

	// Technical context (contains technical terms)
	technicalTerms := []string{"deploy", "config", "test", "fix", "update", "implement", "review"}
	technicalCount := 0
	for _, term := range technicalTerms {
		if strings.Contains(strings.ToLower(summary), term) {
			technicalCount++
		}
	}
	
	if technicalCount > 0 {
		qualityScore += 25
		qualityFactors = append(qualityFactors, fmt.Sprintf("✓ Contains %d technical terms", technicalCount))
	} else {
		qualityFactors = append(qualityFactors, "⚠ Limited technical context")
	}
	
	// Data completeness (has both issues and comments)
	if issueCount > 0 && commentCount > 0 {
		qualityScore += 25
		qualityFactors = append(qualityFactors, "✓ Complete data available")
	} else {
		qualityFactors = append(qualityFactors, "⚠ Limited data available")
	}

	quality.WriteString(fmt.Sprintf("Overall Quality Score: %.0f/100\n", qualityScore))
	quality.WriteString("\nQuality Factors:\n")
	for _, factor := range qualityFactors {
		quality.WriteString(fmt.Sprintf("  %s\n", factor))
	}
	
	// Recommendations based on score
	quality.WriteString("\nRecommendations:\n")
	if qualityScore < 50 {
		quality.WriteString("  • Consider adding more detailed comments to Jira tickets\n")
		quality.WriteString("  • Include technical terms and specific actions in comments\n")
		quality.WriteString("  • Ensure tickets are updated regularly\n")
	} else if qualityScore < 75 {
		quality.WriteString("  • Good summary quality, consider adding more technical details\n")
		quality.WriteString("  • Include deployment status and environment information\n")
	} else {
		quality.WriteString("  • Excellent summary quality! Keep up the detailed documentation\n")
	}

	return quality.String()
}


// removeDuplicates removes duplicate strings from a slice
func removeDuplicates(slice []string) []string {
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

// ExportToObsidian exports the report content to Obsidian-compatible markdown
func (g *Generator) ExportToObsidian(reportContent string, targetDate time.Time) error {
	if !g.config.ExportEnabled {
		return nil
	}

	// Expand tilde in folder path
	folderPath := g.config.ExportFolderPath
	if strings.HasPrefix(folderPath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		folderPath = filepath.Join(homeDir, folderPath[2:])
	}

	// Create folder if it doesn't exist
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return fmt.Errorf("failed to create export folder: %w", err)
	}

	// Generate filename with date
	filename := targetDate.Format(g.config.ExportFileDate) + ".md"
	filePath := filepath.Join(folderPath, filename)

	// Create Obsidian-compatible content with frontmatter
	obsidianContent := g.generateObsidianMarkdown(reportContent, targetDate)

	// Write to file
	if err := os.WriteFile(filePath, []byte(obsidianContent), 0644); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	return nil
}

// generateObsidianMarkdown creates Obsidian-compatible markdown with proper frontmatter and tags
func (g *Generator) generateObsidianMarkdown(reportContent string, targetDate time.Time) string {
	var content strings.Builder

	// Add YAML frontmatter
	content.WriteString("---\n")
	content.WriteString(fmt.Sprintf("date: %s\n", targetDate.Format("2006-01-02")))
	content.WriteString(fmt.Sprintf("title: Daily Standup Report - %s\n", targetDate.Format("January 2, 2006")))
	content.WriteString("type: daily-report\n")
	
	// Add tags from config plus the date tag
	allTags := append(g.config.ExportTags, targetDate.Format("2006-01-02"))
	content.WriteString("tags:\n")
	for _, tag := range allTags {
		content.WriteString(fmt.Sprintf("  - %s\n", tag))
	}
	
	// Add creation timestamp
	content.WriteString(fmt.Sprintf("created: %s\n", time.Now().Format("2006-01-02T15:04:05-07:00")))
	content.WriteString("---\n\n")

	// Add linking to previous and next reports
	yesterday := targetDate.Add(-24 * time.Hour)
	tomorrow := targetDate.Add(24 * time.Hour)
	
	content.WriteString("## Navigation\n\n")
	content.WriteString(fmt.Sprintf("← [[%s]] | [[%s]] →\n\n", 
		yesterday.Format(g.config.ExportFileDate), 
		tomorrow.Format(g.config.ExportFileDate)))

	// Add the main report content
	content.WriteString(reportContent)

	// Add footer with backlinks and tags
	content.WriteString("\n\n---\n\n")
	content.WriteString("## Tags\n\n")
	for _, tag := range allTags {
		content.WriteString(fmt.Sprintf("#%s ", tag))
	}
	content.WriteString("\n\n")
	
	content.WriteString("## Related Notes\n\n")
	content.WriteString("*This section will be automatically populated by Obsidian's backlinks*\n")

	return content.String()
}

// generateFieldGroupedReport creates a report grouped by the specified custom field
func (g *Generator) generateFieldGroupedReport(issues []jira.Issue, commentsMap map[string][]jira.Comment, worklogs []jira.WorklogEntry, targetDate time.Time, fieldName string) (string, error) {
	// Group issues by the specified field value
	fieldGroups := g.groupIssuesByField(issues, fieldName)
	
	switch g.config.Format {
	case "markdown":
		return g.generateMarkdownFieldGrouped(fieldGroups, commentsMap, worklogs, targetDate, fieldName)
	default:
		return g.generateConsoleFieldGrouped(fieldGroups, commentsMap, worklogs, targetDate, fieldName)
	}
}

// groupIssuesByField groups issues by the value of the specified custom field
func (g *Generator) groupIssuesByField(issues []jira.Issue, fieldName string) map[string][]jira.Issue {
	groups := make(map[string][]jira.Issue)
	
	for _, issue := range issues {
		fieldValue := g.getFieldValueByName(issue, fieldName)
		if fieldValue == "" {
			fieldValue = "Unassigned"
		}
		groups[fieldValue] = append(groups[fieldValue], issue)
	}
	
	return groups
}

// getFieldValueByName gets the value of a field by its configured name
func (g *Generator) getFieldValueByName(issue jira.Issue, fieldName string) string {
	// Try to find the field ID from configuration
	// For now, we'll use a simple heuristic to find common field mappings
	fieldMapping := map[string]string{
		"squad":     "customfield_12944", // Common Squad field
		"team":      "customfield_12945", // Common Team field
		"component": "customfield_12946", // Common Component field
		"epic":      "customfield_10014", // Common Epic Link field
		"sprint":    "customfield_10007", // Common Sprint field
	}
	
	if fieldID, exists := fieldMapping[strings.ToLower(fieldName)]; exists {
		return issue.Fields.GetCustomFieldValue(fieldID)
	}
	
	// If no mapping found, try the field name as-is (might be a field ID)
	if strings.HasPrefix(fieldName, "customfield_") {
		return issue.Fields.GetCustomFieldValue(fieldName)
	}
	
	// Try as a standard field
	switch strings.ToLower(fieldName) {
	case "project":
		return issue.Fields.Project.Name
	case "priority":
		return issue.Fields.Priority.Name
	case "status":
		return issue.Fields.Status.Name
	case "issuetype", "issue_type":
		return issue.Fields.IssueType.Name
	case "assignee":
		if issue.Fields.Assignee != nil {
			return issue.Fields.Assignee.DisplayName
		}
		return "Unassigned"
	case "reporter":
		return issue.Fields.Reporter.DisplayName
	}
	
	return ""
}

// generateConsoleFieldGrouped generates console output grouped by field
func (g *Generator) generateConsoleFieldGrouped(fieldGroups map[string][]jira.Issue, commentsMap map[string][]jira.Comment, worklogs []jira.WorklogEntry, targetDate time.Time, fieldName string) (string, error) {
	var report strings.Builder
	
	// Header
	report.WriteString(fmt.Sprintf("🚀 Daily Standup Report - %s\n", targetDate.Format("January 2, 2006")))
	report.WriteString(strings.Repeat("=", 50) + "\n")
	report.WriteString(fmt.Sprintf("📝 Issues grouped by %s\n\n", strings.Title(fieldName)))

	// AI Summary if enabled
	if g.config.LLMEnabled {
		allComments := []jira.Comment{}
		for _, comments := range commentsMap {
			allComments = append(allComments, comments...)
		}
		
		var allIssues []jira.Issue
		for _, groupIssues := range fieldGroups {
			allIssues = append(allIssues, groupIssues...)
		}
		
		if hasMeaningfulComments(allComments) {
			summary, err := g.summarizer.GenerateStandupSummaryWithComments(allIssues, allComments, worklogs)
			if err == nil && summary != "" {
				report.WriteString("🤖 AI SUMMARY OF TODAY'S WORK\n")
				report.WriteString(fmt.Sprintf("%s\n\n", summary))
			}
		} else if len(allComments) > 0 {
			// Show warning when there are comments but they're not meaningful enough for AI summary
			report.WriteString("⚠️  AI SUMMARY SKIPPED\n")
			report.WriteString("No meaningful comment content found for AI summarization.\n")
			report.WriteString("Consider adding more detailed comments to your Jira tickets for better AI insights.\n\n")
		}
	}

	// Summary
	totalIssues := 0
	for _, groupIssues := range fieldGroups {
		totalIssues += len(groupIssues)
	}
	
	totalComments := 0
	for _, comments := range commentsMap {
		totalComments += len(comments)
	}
	
	report.WriteString("📊 SUMMARY\n")
	report.WriteString(fmt.Sprintf("• Total issues: %d\n", totalIssues))
	report.WriteString(fmt.Sprintf("• Groups by %s: %d\n", fieldName, len(fieldGroups)))
	report.WriteString(fmt.Sprintf("• Total comments added: %d\n", totalComments))
	report.WriteString(fmt.Sprintf("• Worklog entries: %d\n\n", len(worklogs)))

	// Sort groups by name for consistent output
	var groupNames []string
	for groupName := range fieldGroups {
		groupNames = append(groupNames, groupName)
	}
	sort.Strings(groupNames)

	// Generate each group section
	for _, groupName := range groupNames {
		groupIssues := fieldGroups[groupName]
		report.WriteString(fmt.Sprintf("🏷️  %s (%d issues)\n", strings.ToUpper(groupName), len(groupIssues)))
		report.WriteString(strings.Repeat("-", 30) + "\n")
		
		// Group issues within each field group by status
		statusGroups := groupIssuesByStatus(groupIssues)
		
		// In Progress section
		if inProgress, exists := statusGroups["In Progress"]; exists && len(inProgress) > 0 {
			report.WriteString("🔄 Currently Working On:\n")
			for _, issue := range inProgress {
				report.WriteString(g.formatIssueConsoleWithComments(issue, commentsMap[issue.Key]))
			}
		}

		// Recently completed section
		if done, exists := statusGroups["Done"]; exists && len(done) > 0 {
			report.WriteString("✅ Recently Completed:\n")
			for _, issue := range done {
				report.WriteString(g.formatIssueConsoleWithComments(issue, commentsMap[issue.Key]))
			}
		}

		// To Do section
		if todo, exists := statusGroups["To Do"]; exists && len(todo) > 0 {
			report.WriteString("📋 To Do:\n")
			for _, issue := range todo {
				report.WriteString(g.formatIssueConsoleWithComments(issue, commentsMap[issue.Key]))
			}
		}
		
		report.WriteString("\n")
	}

	// Worklog section
	if len(worklogs) > 0 {
		report.WriteString("⏰ WORK LOG\n")
		for _, worklog := range worklogs {
			report.WriteString(g.formatWorklogConsole(worklog))
		}
		report.WriteString("\n")
	}

	// Footer
	report.WriteString("---\n")
	report.WriteString("Generated by my-day CLI 🤖\n")

	return report.String(), nil
}

// generateMarkdownFieldGrouped generates markdown output grouped by field
func (g *Generator) generateMarkdownFieldGrouped(fieldGroups map[string][]jira.Issue, commentsMap map[string][]jira.Comment, worklogs []jira.WorklogEntry, targetDate time.Time, fieldName string) (string, error) {
	var report strings.Builder
	
	// Header
	report.WriteString(fmt.Sprintf("# Daily Standup Report - %s\n\n", targetDate.Format("January 2, 2006")))
	report.WriteString(fmt.Sprintf("*Issues grouped by %s*\n\n", strings.Title(fieldName)))

	// AI Summary if enabled
	if g.config.LLMEnabled {
		allComments := []jira.Comment{}
		for _, comments := range commentsMap {
			allComments = append(allComments, comments...)
		}
		
		var allIssues []jira.Issue
		for _, groupIssues := range fieldGroups {
			allIssues = append(allIssues, groupIssues...)
		}
		
		if hasMeaningfulComments(allComments) {
			summary, err := g.summarizer.GenerateStandupSummaryWithComments(allIssues, allComments, worklogs)
			if err == nil && summary != "" {
				report.WriteString("## 🤖 AI Summary of Today's Work\n\n")
				report.WriteString(fmt.Sprintf("%s\n\n", summary))
			}
		} else if len(allComments) > 0 {
			// Show warning when there are comments but they're not meaningful enough for AI summary
			report.WriteString("## ⚠️ AI Summary Skipped\n\n")
			report.WriteString("No meaningful comment content found for AI summarization.\n")
			report.WriteString("Consider adding more detailed comments to your Jira tickets for better AI insights.\n\n")
		}
	}

	// Summary
	totalIssues := 0
	for _, groupIssues := range fieldGroups {
		totalIssues += len(groupIssues)
	}
	
	totalComments := 0
	for _, comments := range commentsMap {
		totalComments += len(comments)
	}
	
	report.WriteString("## Summary\n\n")
	report.WriteString(fmt.Sprintf("- **Total issues**: %d\n", totalIssues))
	report.WriteString(fmt.Sprintf("- **Groups by %s**: %d\n", fieldName, len(fieldGroups)))
	report.WriteString(fmt.Sprintf("- **Total comments added**: %d\n", totalComments))
	report.WriteString(fmt.Sprintf("- **Worklog entries**: %d\n\n", len(worklogs)))

	// Sort groups by name for consistent output
	var groupNames []string
	for groupName := range fieldGroups {
		groupNames = append(groupNames, groupName)
	}
	sort.Strings(groupNames)

	// Generate each group section
	for _, groupName := range groupNames {
		groupIssues := fieldGroups[groupName]
		report.WriteString(fmt.Sprintf("## 🏷️ %s (%d issues)\n\n", strings.Title(groupName), len(groupIssues)))
		
		// Group issues within each field group by status
		statusGroups := groupIssuesByStatus(groupIssues)
		
		// In Progress section
		if inProgress, exists := statusGroups["In Progress"]; exists && len(inProgress) > 0 {
			report.WriteString("### 🔄 Currently Working On\n\n")
			for _, issue := range inProgress {
				report.WriteString(g.formatIssueMarkdownWithComments(issue, commentsMap[issue.Key]))
			}
			report.WriteString("\n")
		}

		// Recently completed section
		if done, exists := statusGroups["Done"]; exists && len(done) > 0 {
			report.WriteString("### ✅ Recently Completed\n\n")
			for _, issue := range done {
				report.WriteString(g.formatIssueMarkdownWithComments(issue, commentsMap[issue.Key]))
			}
			report.WriteString("\n")
		}

		// To Do section
		if todo, exists := statusGroups["To Do"]; exists && len(todo) > 0 {
			report.WriteString("### 📋 To Do\n\n")
			for _, issue := range todo {
				report.WriteString(g.formatIssueMarkdownWithComments(issue, commentsMap[issue.Key]))
			}
			report.WriteString("\n")
		}
	}

	// Worklog section
	if len(worklogs) > 0 {
		report.WriteString("## ⏰ Work Log\n\n")
		for _, worklog := range worklogs {
			report.WriteString(g.formatWorklogMarkdown(worklog))
		}
		report.WriteString("\n")
	}

	// Footer
	report.WriteString("---\n")
	report.WriteString("*Generated by my-day CLI*\n")

	return report.String(), nil
}