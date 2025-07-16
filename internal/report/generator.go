package report

import (
	"fmt"
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
}

// NewGenerator creates a new report generator
func NewGenerator(config *Config) *Generator {
	// Initialize LLM summarizer based on configuration
	llmConfig := llm.LLMConfig{
		Enabled:     config.LLMEnabled,
		Mode:        config.LLMMode,
		Model:       config.LLMModel,
		OllamaURL:   config.OllamaURL,
		OllamaModel: config.OllamaModel,
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
		
		if len(allComments) > 0 {
			// Use the enhanced LLM method for intelligent summary
			summary, err := g.summarizer.GenerateStandupSummaryWithComments(issues, allComments, worklogs)
			if err == nil && summary != "" {
				report.WriteString("🤖 AI SUMMARY OF TODAY'S WORK\n")
				report.WriteString(fmt.Sprintf("%s\n\n", summary))
			}
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
		
		if len(allComments) > 0 {
			// Use the enhanced LLM method for intelligent summary
			summary, err := g.summarizer.GenerateStandupSummaryWithComments(issues, allComments, worklogs)
			if err == nil && summary != "" {
				report.WriteString("## 🤖 AI Summary of Today's Work\n\n")
				report.WriteString(fmt.Sprintf("%s\n\n", summary))
			}
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