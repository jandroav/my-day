package llm

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// DebugLogger provides comprehensive debugging capabilities for LLM processing
type DebugLogger struct {
	enabled    bool
	verbose    bool
	logFile    *os.File
	steps      []DebugStep
	warnings   []DebugWarning
	startTime  time.Time
}

// DebugStep represents a single processing step with timing and data
type DebugStep struct {
	Step        string      `json:"step"`
	Timestamp   time.Time   `json:"timestamp"`
	Duration    time.Duration `json:"duration"`
	InputData   interface{} `json:"input_data,omitempty"`
	OutputData  interface{} `json:"output_data,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Success     bool        `json:"success"`
	Error       string      `json:"error,omitempty"`
}

// DebugWarning represents validation warnings and quality issues
type DebugWarning struct {
	Type        string    `json:"type"`
	Message     string    `json:"message"`
	Context     string    `json:"context,omitempty"`
	Severity    string    `json:"severity"` // "low", "medium", "high"
	Timestamp   time.Time `json:"timestamp"`
	Suggestion  string    `json:"suggestion,omitempty"`
}

// DebugReport contains comprehensive debugging information
type DebugReport struct {
	SessionID       string         `json:"session_id"`
	StartTime       time.Time      `json:"start_time"`
	EndTime         time.Time      `json:"end_time"`
	TotalDuration   time.Duration  `json:"total_duration"`
	Steps           []DebugStep    `json:"steps"`
	Warnings        []DebugWarning `json:"warnings"`
	Summary         DebugSummary   `json:"summary"`
	Configuration   map[string]interface{} `json:"configuration"`
}

// DebugSummary provides high-level debugging insights
type DebugSummary struct {
	TotalSteps      int           `json:"total_steps"`
	SuccessfulSteps int           `json:"successful_steps"`
	FailedSteps     int           `json:"failed_steps"`
	TotalWarnings   int           `json:"total_warnings"`
	ProcessingTime  time.Duration `json:"processing_time"`
	QualityScore    float64       `json:"quality_score"`
	Recommendations []string      `json:"recommendations"`
}

// NewDebugLogger creates a new debug logger instance
func NewDebugLogger(enabled, verbose bool) *DebugLogger {
	logger := &DebugLogger{
		enabled:   enabled,
		verbose:   verbose,
		steps:     make([]DebugStep, 0),
		warnings:  make([]DebugWarning, 0),
		startTime: time.Now(),
	}
	
	// Create log file if verbose mode is enabled
	if verbose {
		logFileName := fmt.Sprintf("llm_debug_%s.log", time.Now().Format("20060102_150405"))
		if file, err := os.Create(logFileName); err == nil {
			logger.logFile = file
			logger.logToFile(fmt.Sprintf("Debug session started at %s", time.Now().Format(time.RFC3339)))
		}
	}
	
	return logger
}

// LogProcessingStep logs a processing step with input/output data
func (d *DebugLogger) LogProcessingStep(step string, data interface{}) {
	if !d.enabled {
		return
	}
	
	stepStart := time.Now()
	
	debugStep := DebugStep{
		Step:      step,
		Timestamp: stepStart,
		InputData: data,
		Success:   true,
		Metadata:  make(map[string]interface{}),
	}
	
	// Add step metadata
	debugStep.Metadata["step_number"] = len(d.steps) + 1
	debugStep.Metadata["session_duration"] = time.Since(d.startTime)
	
	d.steps = append(d.steps, debugStep)
	
	if d.verbose {
		d.logVerbose("STEP", fmt.Sprintf("Starting step: %s", step))
		if data != nil {
			d.logVerbose("INPUT", fmt.Sprintf("Input data: %+v", data))
		}
	}
}

// LogStepCompletion logs the completion of a processing step
func (d *DebugLogger) LogStepCompletion(step string, outputData interface{}, err error) {
	if !d.enabled {
		return
	}
	
	// Find the most recent step with this name
	for i := len(d.steps) - 1; i >= 0; i-- {
		if d.steps[i].Step == step {
			d.steps[i].Duration = time.Since(d.steps[i].Timestamp)
			d.steps[i].OutputData = outputData
			
			if err != nil {
				d.steps[i].Success = false
				d.steps[i].Error = err.Error()
				
				if d.verbose {
					d.logVerbose("ERROR", fmt.Sprintf("Step %s failed: %v", step, err))
				}
			} else {
				if d.verbose {
					d.logVerbose("SUCCESS", fmt.Sprintf("Step %s completed in %v", step, d.steps[i].Duration))
					if outputData != nil {
						d.logVerbose("OUTPUT", fmt.Sprintf("Output data: %+v", outputData))
					}
				}
			}
			break
		}
	}
}

// LogPatternMatches logs pattern matching results with confidence scores
func (d *DebugLogger) LogPatternMatches(patterns []PatternMatch) {
	if !d.enabled {
		return
	}
	
	d.LogProcessingStep("pattern_matching", map[string]interface{}{
		"pattern_count": len(patterns),
		"patterns":      patterns,
	})
	
	if d.verbose {
		for _, pattern := range patterns {
			d.logVerbose("PATTERN", fmt.Sprintf("Matched: %s (confidence: %.2f)", 
				pattern.Pattern.Name, pattern.Confidence))
		}
	}
	
	// Validate pattern quality
	d.validatePatternQuality(patterns)
	
	d.LogStepCompletion("pattern_matching", patterns, nil)
}

// LogSummaryGeneration logs summary generation with input and output
func (d *DebugLogger) LogSummaryGeneration(input, output string) {
	if !d.enabled {
		return
	}
	
	d.LogProcessingStep("summary_generation", map[string]interface{}{
		"input_length":  len(input),
		"output_length": len(output),
		"input_preview": d.truncateText(input, 200),
	})
	
	if d.verbose {
		d.logVerbose("SUMMARY_INPUT", fmt.Sprintf("Input text (%d chars): %s", 
			len(input), d.truncateText(input, 500)))
		d.logVerbose("SUMMARY_OUTPUT", fmt.Sprintf("Generated summary (%d chars): %s", 
			len(output), output))
	}
	
	// Validate summary quality
	d.validateSummaryQuality(input, output)
	
	d.LogStepCompletion("summary_generation", output, nil)
}

// LogDataProcessing logs data processing steps with validation
func (d *DebugLogger) LogDataProcessing(stepName string, inputData, outputData interface{}) {
	if !d.enabled {
		return
	}
	
	d.LogProcessingStep(stepName, inputData)
	
	if d.verbose {
		d.logVerbose("PROCESSING", fmt.Sprintf("Processing step: %s", stepName))
	}
	
	// Validate processed data if it's ProcessedData
	if processedData, ok := outputData.(*ProcessedData); ok {
		d.validateProcessedData(processedData)
	}
	
	d.LogStepCompletion(stepName, outputData, nil)
}

// AddWarning adds a validation warning or quality issue
func (d *DebugLogger) AddWarning(warningType, message, context, severity string) {
	if !d.enabled {
		return
	}
	
	warning := DebugWarning{
		Type:      warningType,
		Message:   message,
		Context:   context,
		Severity:  severity,
		Timestamp: time.Now(),
	}
	
	// Add suggestions based on warning type
	switch warningType {
	case "empty_summary":
		warning.Suggestion = "Check if input data contains meaningful content"
	case "low_confidence":
		warning.Suggestion = "Consider using more specific technical terms or patterns"
	case "processing_failure":
		warning.Suggestion = "Verify input data format and try fallback processing"
	case "validation_error":
		warning.Suggestion = "Check data structure and required fields"
	}
	
	d.warnings = append(d.warnings, warning)
	
	if d.verbose {
		d.logVerbose("WARNING", fmt.Sprintf("[%s] %s: %s", severity, warningType, message))
		if context != "" {
			d.logVerbose("CONTEXT", context)
		}
		if warning.Suggestion != "" {
			d.logVerbose("SUGGESTION", warning.Suggestion)
		}
	}
}

// GetDebugReport generates a comprehensive debug report
func (d *DebugLogger) GetDebugReport() (*DebugReport, error) {
	if !d.enabled {
		return nil, fmt.Errorf("debug logging is not enabled")
	}
	
	endTime := time.Now()
	
	// Calculate summary statistics
	summary := d.calculateSummary()
	
	report := &DebugReport{
		SessionID:     fmt.Sprintf("llm_debug_%d", d.startTime.Unix()),
		StartTime:     d.startTime,
		EndTime:       endTime,
		TotalDuration: endTime.Sub(d.startTime),
		Steps:         d.steps,
		Warnings:      d.warnings,
		Summary:       summary,
		Configuration: map[string]interface{}{
			"debug_enabled": d.enabled,
			"verbose_mode":  d.verbose,
			"log_file":      d.logFile != nil,
		},
	}
	
	return report, nil
}

// PrintDebugSummary prints a human-readable debug summary to stdout
func (d *DebugLogger) PrintDebugSummary() {
	if !d.enabled {
		fmt.Println("Debug logging is not enabled")
		return
	}
	
	report, err := d.GetDebugReport()
	if err != nil {
		fmt.Printf("Error generating debug report: %v\n", err)
		return
	}
	
	fmt.Println("\n=== LLM Processing Debug Summary ===")
	fmt.Printf("Session Duration: %v\n", report.TotalDuration)
	fmt.Printf("Total Steps: %d\n", report.Summary.TotalSteps)
	fmt.Printf("Successful Steps: %d\n", report.Summary.SuccessfulSteps)
	fmt.Printf("Failed Steps: %d\n", report.Summary.FailedSteps)
	fmt.Printf("Warnings: %d\n", report.Summary.TotalWarnings)
	fmt.Printf("Quality Score: %.2f/100\n", report.Summary.QualityScore)
	
	if len(report.Summary.Recommendations) > 0 {
		fmt.Println("\nRecommendations:")
		for _, rec := range report.Summary.Recommendations {
			fmt.Printf("  • %s\n", rec)
		}
	}
	
	if len(d.warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, warning := range d.warnings {
			fmt.Printf("  [%s] %s: %s\n", strings.ToUpper(warning.Severity), warning.Type, warning.Message)
		}
	}
	
	if d.verbose && len(d.steps) > 0 {
		fmt.Println("\nProcessing Steps:")
		for i, step := range d.steps {
			status := "✓"
			if !step.Success {
				status = "✗"
			}
			fmt.Printf("  %d. %s %s (%v)\n", i+1, status, step.Step, step.Duration)
		}
	}
	
	fmt.Println("=====================================\n")
}

// Close closes the debug logger and any open files
func (d *DebugLogger) Close() {
	if d.logFile != nil {
		d.logToFile(fmt.Sprintf("Debug session ended at %s", time.Now().Format(time.RFC3339)))
		d.logFile.Close()
		d.logFile = nil
	}
}

// validatePatternQuality validates the quality of pattern matches
func (d *DebugLogger) validatePatternQuality(patterns []PatternMatch) {
	if len(patterns) == 0 {
		d.AddWarning("no_patterns", "No technical patterns detected in text", "", "medium")
		return
	}
	
	// Check for low confidence patterns
	lowConfidenceCount := 0
	for _, pattern := range patterns {
		if pattern.Confidence < 0.5 {
			lowConfidenceCount++
		}
	}
	
	if lowConfidenceCount > len(patterns)/2 {
		d.AddWarning("low_confidence", 
			fmt.Sprintf("%d out of %d patterns have low confidence", lowConfidenceCount, len(patterns)),
			"", "medium")
	}
}

// validateSummaryQuality validates the quality of generated summaries
func (d *DebugLogger) validateSummaryQuality(input, output string) {
	if output == "" {
		d.AddWarning("empty_summary", "Generated summary is empty", 
			fmt.Sprintf("Input length: %d", len(input)), "high")
		return
	}
	
	if len(output) < 10 {
		d.AddWarning("short_summary", "Generated summary is very short", 
			fmt.Sprintf("Summary: %s", output), "medium")
	}
	
	if len(input) > 0 && len(output) > len(input) {
		d.AddWarning("summary_too_long", "Summary is longer than input text", 
			fmt.Sprintf("Input: %d chars, Output: %d chars", len(input), len(output)), "medium")
	}
	
	// Check for generic/meaningless summaries
	genericPhrases := []string{
		"no recent activity",
		"multiple development activities",
		"technical work",
		"general progress",
	}
	
	lowerOutput := strings.ToLower(output)
	for _, phrase := range genericPhrases {
		if strings.Contains(lowerOutput, phrase) {
			d.AddWarning("generic_summary", "Summary appears to be generic", 
				fmt.Sprintf("Contains phrase: %s", phrase), "low")
			break
		}
	}
}

// validateProcessedData validates ProcessedData structure
func (d *DebugLogger) validateProcessedData(data *ProcessedData) {
	if data == nil {
		d.AddWarning("validation_error", "ProcessedData is nil", "", "high")
		return
	}
	
	if len(data.Issues) == 0 {
		d.AddWarning("no_issues", "No issues found in processed data", "", "medium")
	}
	
	// Validate individual issues
	for i, issue := range data.Issues {
		if issue.Issue.Key == "" {
			d.AddWarning("validation_error", 
				fmt.Sprintf("Issue %d has empty key", i), "", "high")
		}
		
		if issue.WorkSummary == "" {
			d.AddWarning("missing_summary", 
				fmt.Sprintf("Issue %s has no work summary", issue.Issue.Key), "", "medium")
		}
		
		if len(issue.KeyActivities) == 0 {
			d.AddWarning("no_activities", 
				fmt.Sprintf("Issue %s has no key activities", issue.Issue.Key), "", "low")
		}
	}
	
	// Check technical context
	if data.TechnicalContext == nil {
		d.AddWarning("missing_context", "Technical context is missing", "", "medium")
	} else if len(data.TechnicalContext.Technologies) == 0 {
		d.AddWarning("no_technologies", "No technologies detected", "", "low")
	}
}

// calculateSummary calculates summary statistics for the debug report
func (d *DebugLogger) calculateSummary() DebugSummary {
	successfulSteps := 0
	failedSteps := 0
	totalProcessingTime := time.Duration(0)
	
	for _, step := range d.steps {
		if step.Success {
			successfulSteps++
		} else {
			failedSteps++
		}
		totalProcessingTime += step.Duration
	}
	
	// Calculate quality score based on success rate and warnings
	qualityScore := 100.0
	if len(d.steps) > 0 {
		successRate := float64(successfulSteps) / float64(len(d.steps))
		qualityScore = successRate * 100
		
		// Reduce score based on warnings
		highSeverityWarnings := 0
		mediumSeverityWarnings := 0
		for _, warning := range d.warnings {
			switch warning.Severity {
			case "high":
				highSeverityWarnings++
			case "medium":
				mediumSeverityWarnings++
			}
		}
		
		qualityScore -= float64(highSeverityWarnings) * 20
		qualityScore -= float64(mediumSeverityWarnings) * 10
		
		if qualityScore < 0 {
			qualityScore = 0
		}
	}
	
	// Generate recommendations
	recommendations := d.generateRecommendations()
	
	return DebugSummary{
		TotalSteps:      len(d.steps),
		SuccessfulSteps: successfulSteps,
		FailedSteps:     failedSteps,
		TotalWarnings:   len(d.warnings),
		ProcessingTime:  totalProcessingTime,
		QualityScore:    qualityScore,
		Recommendations: recommendations,
	}
}

// generateRecommendations generates recommendations based on debug data
func (d *DebugLogger) generateRecommendations() []string {
	var recommendations []string
	
	// Analyze warnings to generate recommendations
	warningTypes := make(map[string]int)
	for _, warning := range d.warnings {
		warningTypes[warning.Type]++
	}
	
	if warningTypes["empty_summary"] > 0 {
		recommendations = append(recommendations, "Improve input data quality or use fallback summarization")
	}
	
	if warningTypes["low_confidence"] > 0 {
		recommendations = append(recommendations, "Add more specific technical patterns to improve recognition")
	}
	
	if warningTypes["no_patterns"] > 0 {
		recommendations = append(recommendations, "Ensure input text contains technical terminology")
	}
	
	if warningTypes["validation_error"] > 0 {
		recommendations = append(recommendations, "Check data structure validation and error handling")
	}
	
	// Analyze processing performance
	if len(d.steps) > 0 {
		avgDuration := time.Duration(0)
		for _, step := range d.steps {
			avgDuration += step.Duration
		}
		avgDuration = avgDuration / time.Duration(len(d.steps))
		
		if avgDuration > time.Second {
			recommendations = append(recommendations, "Consider optimizing processing performance")
		}
	}
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Processing completed successfully with good quality")
	}
	
	return recommendations
}

// logVerbose logs verbose debug information
func (d *DebugLogger) logVerbose(level, message string) {
	if !d.verbose {
		return
	}
	
	timestamp := time.Now().Format("15:04:05.000")
	logMessage := fmt.Sprintf("[%s] %s: %s", timestamp, level, message)
	
	// Log to console
	log.Println(logMessage)
	
	// Log to file if available
	if d.logFile != nil {
		d.logToFile(logMessage)
	}
}

// logToFile writes a message to the log file
func (d *DebugLogger) logToFile(message string) {
	if d.logFile != nil {
		d.logFile.WriteString(fmt.Sprintf("%s\n", message))
		d.logFile.Sync()
	}
}

// truncateText truncates text to a maximum length for logging
func (d *DebugLogger) truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	
	return text[:maxLength-3] + "..."
}

// SaveDebugReport saves the debug report to a JSON file
func (d *DebugLogger) SaveDebugReport(filename string) error {
	if !d.enabled {
		return fmt.Errorf("debug logging is not enabled")
	}
	
	report, err := d.GetDebugReport()
	if err != nil {
		return err
	}
	
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal debug report: %v", err)
	}
	
	return os.WriteFile(filename, jsonData, 0644)
}