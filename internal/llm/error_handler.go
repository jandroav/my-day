package llm

import (
	"fmt"
	"strings"
	"time"
	"my-day/internal/jira"
)

// ErrorHandler provides comprehensive error handling and fallback strategies for LLM processing
type ErrorHandler struct {
	fallbackStrategy string // "graceful", "strict", "minimal"
	debugLogger      *DebugLogger
}

// LLMError represents a structured error with context and suggestions
type LLMError struct {
	Type        string    `json:"type"`
	Message     string    `json:"message"`
	Context     string    `json:"context,omitempty"`
	Cause       error     `json:"cause,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	Severity    string    `json:"severity"` // "low", "medium", "high", "critical"
	Suggestions []string  `json:"suggestions,omitempty"`
	Recoverable bool      `json:"recoverable"`
}

// FallbackResult represents the result of a fallback operation
type FallbackResult struct {
	Success      bool        `json:"success"`
	Result       interface{} `json:"result"`
	FallbackUsed string      `json:"fallback_used"`
	OriginalError *LLMError  `json:"original_error"`
	Quality      string      `json:"quality"` // "high", "medium", "low"
}

// NewErrorHandler creates a new error handler with specified fallback strategy
func NewErrorHandler(fallbackStrategy string, debugLogger *DebugLogger) *ErrorHandler {
	return &ErrorHandler{
		fallbackStrategy: fallbackStrategy,
		debugLogger:      debugLogger,
	}
}

// HandleProcessingError handles errors during data processing with appropriate fallbacks
func (eh *ErrorHandler) HandleProcessingError(err error, context string, inputData interface{}) *FallbackResult {
	llmError := eh.createLLMError("processing_error", err, context)
	
	if eh.debugLogger != nil {
		eh.debugLogger.AddWarning("processing_error", llmError.Message, context, llmError.Severity)
	}
	
	switch eh.fallbackStrategy {
	case "strict":
		return &FallbackResult{
			Success:       false,
			Result:        nil,
			FallbackUsed:  "none",
			OriginalError: llmError,
			Quality:       "none",
		}
	case "minimal":
		return eh.handleMinimalFallback(llmError, inputData)
	default: // "graceful"
		return eh.handleGracefulFallback(llmError, inputData)
	}
}

// HandleSummaryError handles errors during summary generation
func (eh *ErrorHandler) HandleSummaryError(err error, inputText string) *FallbackResult {
	llmError := eh.createLLMError("summary_error", err, fmt.Sprintf("Input length: %d", len(inputText)))
	
	if eh.debugLogger != nil {
		eh.debugLogger.AddWarning("summary_error", llmError.Message, llmError.Context, llmError.Severity)
	}
	
	// Try different fallback strategies based on error type
	if strings.Contains(err.Error(), "empty") || strings.Contains(err.Error(), "no content") {
		return eh.handleEmptyContentFallback(llmError, inputText)
	}
	
	if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "processing") {
		return eh.handleProcessingTimeoutFallback(llmError, inputText)
	}
	
	return eh.handleGenericSummaryFallback(llmError, inputText)
}

// HandleCommentProcessingError handles errors during comment processing
func (eh *ErrorHandler) HandleCommentProcessingError(err error, comments []jira.Comment) *FallbackResult {
	llmError := eh.createLLMError("comment_processing_error", err, 
		fmt.Sprintf("Comment count: %d", len(comments)))
	
	if eh.debugLogger != nil {
		eh.debugLogger.AddWarning("comment_processing_error", llmError.Message, llmError.Context, llmError.Severity)
	}
	
	// Attempt to process comments with basic rule-based approach
	return eh.handleCommentFallback(llmError, comments)
}

// HandlePatternMatchingError handles errors during pattern matching
func (eh *ErrorHandler) HandlePatternMatchingError(err error, text string) *FallbackResult {
	llmError := eh.createLLMError("pattern_matching_error", err, 
		fmt.Sprintf("Text length: %d", len(text)))
	
	if eh.debugLogger != nil {
		eh.debugLogger.AddWarning("pattern_matching_error", llmError.Message, llmError.Context, llmError.Severity)
	}
	
	// Use basic keyword matching as fallback
	return eh.handlePatternMatchingFallback(llmError, text)
}

// createLLMError creates a structured LLM error with context and suggestions
func (eh *ErrorHandler) createLLMError(errorType string, cause error, context string) *LLMError {
	llmError := &LLMError{
		Type:      errorType,
		Message:   cause.Error(),
		Context:   context,
		Cause:     cause,
		Timestamp: time.Now(),
		Severity:  eh.determineSeverity(errorType, cause),
		Recoverable: eh.isRecoverable(errorType, cause),
	}
	
	// Add specific suggestions based on error type
	llmError.Suggestions = eh.generateSuggestions(errorType, cause, context)
	
	return llmError
}

// determineSeverity determines the severity of an error
func (eh *ErrorHandler) determineSeverity(errorType string, cause error) string {
	errorMsg := strings.ToLower(cause.Error())
	
	// Critical errors that prevent any processing
	if strings.Contains(errorMsg, "panic") || strings.Contains(errorMsg, "fatal") {
		return "critical"
	}
	
	// High severity errors that significantly impact functionality
	if strings.Contains(errorMsg, "timeout") || strings.Contains(errorMsg, "connection") ||
		strings.Contains(errorMsg, "authentication") {
		return "high"
	}
	
	// Medium severity errors that can be worked around
	if strings.Contains(errorMsg, "validation") || strings.Contains(errorMsg, "format") ||
		strings.Contains(errorMsg, "parsing") {
		return "medium"
	}
	
	// Low severity errors that have minimal impact
	return "low"
}

// isRecoverable determines if an error is recoverable with fallbacks
func (eh *ErrorHandler) isRecoverable(errorType string, cause error) bool {
	errorMsg := strings.ToLower(cause.Error())
	
	// Non-recoverable errors
	if strings.Contains(errorMsg, "panic") || strings.Contains(errorMsg, "fatal") ||
		strings.Contains(errorMsg, "authentication") {
		return false
	}
	
	// Most processing errors are recoverable with fallbacks
	return true
}

// generateSuggestions generates helpful suggestions based on error type
func (eh *ErrorHandler) generateSuggestions(errorType string, cause error, context string) []string {
	var suggestions []string
	errorMsg := strings.ToLower(cause.Error())
	
	switch errorType {
	case "processing_error":
		suggestions = append(suggestions, "Try running sync to refresh data")
		suggestions = append(suggestions, "Check if input data is properly formatted")
		if strings.Contains(errorMsg, "empty") {
			suggestions = append(suggestions, "Ensure there are issues or comments to process")
		}
	case "summary_error":
		suggestions = append(suggestions, "Verify that input text contains meaningful content")
		suggestions = append(suggestions, "Try using a different summary style or length")
	case "comment_processing_error":
		suggestions = append(suggestions, "Check if comments contain valid text content")
		suggestions = append(suggestions, "Verify comment permissions and access")
	case "pattern_matching_error":
		suggestions = append(suggestions, "Ensure text contains technical terminology")
		suggestions = append(suggestions, "Try with more specific technical keywords")
	}
	
	// Generic suggestions
	if strings.Contains(errorMsg, "timeout") {
		suggestions = append(suggestions, "Try again with a smaller dataset")
		suggestions = append(suggestions, "Check network connectivity")
	}
	
	return suggestions
}

// handleGracefulFallback implements graceful degradation strategy
func (eh *ErrorHandler) handleGracefulFallback(llmError *LLMError, inputData interface{}) *FallbackResult {
	// Try multiple fallback levels
	
	// Level 1: Try basic rule-based processing
	if result := eh.tryBasicProcessing(inputData); result != nil {
		return &FallbackResult{
			Success:       true,
			Result:        result,
			FallbackUsed:  "basic_processing",
			OriginalError: llmError,
			Quality:       "medium",
		}
	}
	
	// Level 2: Try metadata-based processing
	if result := eh.tryMetadataProcessing(inputData); result != nil {
		return &FallbackResult{
			Success:       true,
			Result:        result,
			FallbackUsed:  "metadata_processing",
			OriginalError: llmError,
			Quality:       "low",
		}
	}
	
	// Level 3: Return minimal safe result
	return &FallbackResult{
		Success:       true,
		Result:        eh.createMinimalResult(inputData),
		FallbackUsed:  "minimal_safe",
		OriginalError: llmError,
		Quality:       "low",
	}
}

// handleMinimalFallback implements minimal fallback strategy
func (eh *ErrorHandler) handleMinimalFallback(llmError *LLMError, inputData interface{}) *FallbackResult {
	return &FallbackResult{
		Success:       true,
		Result:        eh.createMinimalResult(inputData),
		FallbackUsed:  "minimal",
		OriginalError: llmError,
		Quality:       "low",
	}
}

// handleEmptyContentFallback handles cases where content is empty or missing
func (eh *ErrorHandler) handleEmptyContentFallback(llmError *LLMError, inputText string) *FallbackResult {
	var result string
	
	if len(inputText) == 0 {
		result = "No content available for summary"
	} else {
		// Try to extract any meaningful content
		words := strings.Fields(inputText)
		if len(words) > 0 {
			if len(words) <= 10 {
				result = strings.Join(words, " ")
			} else {
				result = strings.Join(words[:10], " ") + "..."
			}
		} else {
			result = "Content contains no readable text"
		}
	}
	
	return &FallbackResult{
		Success:       true,
		Result:        result,
		FallbackUsed:  "empty_content_fallback",
		OriginalError: llmError,
		Quality:       "low",
	}
}

// handleProcessingTimeoutFallback handles processing timeout errors
func (eh *ErrorHandler) handleProcessingTimeoutFallback(llmError *LLMError, inputText string) *FallbackResult {
	// Create a very basic summary from first few words
	words := strings.Fields(inputText)
	var result string
	
	if len(words) == 0 {
		result = "Processing timeout - no content available"
	} else if len(words) <= 5 {
		result = strings.Join(words, " ")
	} else {
		result = strings.Join(words[:5], " ") + " (processing timeout)"
	}
	
	return &FallbackResult{
		Success:       true,
		Result:        result,
		FallbackUsed:  "timeout_fallback",
		OriginalError: llmError,
		Quality:       "low",
	}
}

// handleGenericSummaryFallback handles generic summary errors
func (eh *ErrorHandler) handleGenericSummaryFallback(llmError *LLMError, inputText string) *FallbackResult {
	// Extract first sentence or meaningful chunk
	sentences := strings.Split(inputText, ".")
	var result string
	
	if len(sentences) > 0 && len(strings.TrimSpace(sentences[0])) > 0 {
		result = strings.TrimSpace(sentences[0])
		if len(result) > 100 {
			result = result[:97] + "..."
		}
	} else {
		result = "Summary generation failed - content available but not processable"
	}
	
	return &FallbackResult{
		Success:       true,
		Result:        result,
		FallbackUsed:  "generic_summary_fallback",
		OriginalError: llmError,
		Quality:       "low",
	}
}

// handleCommentFallback handles comment processing fallbacks
func (eh *ErrorHandler) handleCommentFallback(llmError *LLMError, comments []jira.Comment) *FallbackResult {
	if len(comments) == 0 {
		return &FallbackResult{
			Success:       true,
			Result:        "No comments to process",
			FallbackUsed:  "empty_comments",
			OriginalError: llmError,
			Quality:       "low",
		}
	}
	
	// Create basic summary from comment count and first comment
	var result string
	if len(comments) == 1 {
		if comments[0].Body.Text != "" {
			words := strings.Fields(comments[0].Body.Text)
			if len(words) > 10 {
				result = strings.Join(words[:10], " ") + "..."
			} else {
				result = comments[0].Body.Text
			}
		} else {
			result = "Comment added"
		}
	} else {
		result = fmt.Sprintf("%d comments added", len(comments))
		if comments[0].Body.Text != "" {
			words := strings.Fields(comments[0].Body.Text)
			if len(words) > 5 {
				result += " - " + strings.Join(words[:5], " ") + "..."
			}
		}
	}
	
	return &FallbackResult{
		Success:       true,
		Result:        result,
		FallbackUsed:  "basic_comment_summary",
		OriginalError: llmError,
		Quality:       "medium",
	}
}

// handlePatternMatchingFallback handles pattern matching fallbacks
func (eh *ErrorHandler) handlePatternMatchingFallback(llmError *LLMError, text string) *FallbackResult {
	// Use basic keyword matching
	lowerText := strings.ToLower(text)
	var matchedKeywords []string
	
	// Basic technical keywords
	keywords := []string{
		"terraform", "aws", "kubernetes", "docker", "database", "api",
		"deployment", "security", "testing", "ci/cd", "pipeline",
	}
	
	for _, keyword := range keywords {
		if strings.Contains(lowerText, keyword) {
			matchedKeywords = append(matchedKeywords, keyword)
		}
	}
	
	var result interface{}
	if len(matchedKeywords) > 0 {
		result = map[string]interface{}{
			"matched_keywords": matchedKeywords,
			"confidence":       0.5, // Low confidence for basic matching
		}
	} else {
		result = map[string]interface{}{
			"matched_keywords": []string{},
			"confidence":       0.0,
		}
	}
	
	return &FallbackResult{
		Success:       true,
		Result:        result,
		FallbackUsed:  "basic_keyword_matching",
		OriginalError: llmError,
		Quality:       "low",
	}
}

// tryBasicProcessing attempts basic rule-based processing
func (eh *ErrorHandler) tryBasicProcessing(inputData interface{}) interface{} {
	switch data := inputData.(type) {
	case []jira.Issue:
		if len(data) > 0 {
			return fmt.Sprintf("Basic processing of %d issues", len(data))
		}
	case []jira.Comment:
		if len(data) > 0 {
			return fmt.Sprintf("Basic processing of %d comments", len(data))
		}
	case string:
		if len(data) > 0 {
			words := strings.Fields(data)
			if len(words) > 10 {
				return strings.Join(words[:10], " ") + "..."
			}
			return data
		}
	}
	return nil
}

// tryMetadataProcessing attempts metadata-based processing
func (eh *ErrorHandler) tryMetadataProcessing(inputData interface{}) interface{} {
	switch data := inputData.(type) {
	case []jira.Issue:
		if len(data) > 0 {
			// Extract basic metadata
			statuses := make(map[string]int)
			for _, issue := range data {
				statuses[issue.Fields.Status.Name]++
			}
			
			var statusParts []string
			for status, count := range statuses {
				statusParts = append(statusParts, fmt.Sprintf("%d %s", count, status))
			}
			
			return fmt.Sprintf("Issues: %s", strings.Join(statusParts, ", "))
		}
	case []jira.Comment:
		if len(data) > 0 {
			return fmt.Sprintf("Activity with %d comments", len(data))
		}
	}
	return nil
}

// createMinimalResult creates a minimal safe result
func (eh *ErrorHandler) createMinimalResult(inputData interface{}) interface{} {
	switch inputData.(type) {
	case []jira.Issue:
		return "Issue activity detected"
	case []jira.Comment:
		return "Comment activity detected"
	case string:
		return "Text content available"
	default:
		return "Activity detected"
	}
}

// GetErrorStatistics returns statistics about handled errors
func (eh *ErrorHandler) GetErrorStatistics() map[string]interface{} {
	return map[string]interface{}{
		"fallback_strategy": eh.fallbackStrategy,
		"debug_enabled":     eh.debugLogger != nil,
	}
}

// ValidateInput validates input data before processing
func (eh *ErrorHandler) ValidateInput(inputData interface{}) *LLMError {
	switch data := inputData.(type) {
	case []jira.Issue:
		for i, issue := range data {
			if issue.Key == "" {
				return &LLMError{
					Type:        "validation_error",
					Message:     fmt.Sprintf("Issue %d has empty key", i),
					Context:     "input_validation",
					Timestamp:   time.Now(),
					Severity:    "high",
					Recoverable: false,
					Suggestions: []string{"Ensure all issues have valid keys"},
				}
			}
		}
	case []jira.Comment:
		for i, comment := range data {
			if comment.ID == "" {
				return &LLMError{
					Type:        "validation_error",
					Message:     fmt.Sprintf("Comment %d has empty ID", i),
					Context:     "input_validation",
					Timestamp:   time.Now(),
					Severity:    "medium",
					Recoverable: true,
					Suggestions: []string{"Filter out comments with missing IDs"},
				}
			}
		}
	case string:
		if len(strings.TrimSpace(data)) == 0 {
			return &LLMError{
				Type:        "validation_error",
				Message:     "Input text is empty or contains only whitespace",
				Context:     "input_validation",
				Timestamp:   time.Now(),
				Severity:    "medium",
				Recoverable: true,
				Suggestions: []string{"Provide non-empty text content"},
			}
		}
	}
	
	return nil // No validation errors
}