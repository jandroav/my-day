package report

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"my-day/internal/jira"
)

// ReportCache represents a cached report
type ReportCache struct {
	ID                string                   `json:"id"`
	Date              time.Time                `json:"date"`
	Config            *Config                  `json:"config"`
	Content           string                   `json:"content"`
	Format            string                   `json:"format"`
	GeneratedAt       time.Time                `json:"generated_at"`
	InputHash         string                   `json:"input_hash"`
	IssueCount        int                      `json:"issue_count"`
	CommentCount      int                      `json:"comment_count"`
	WorklogCount      int                      `json:"worklog_count"`
	LLMUsed           bool                     `json:"llm_used"`
	GenerationTimeMs  int64                    `json:"generation_time_ms"`
	ExportPaths       map[string]string        `json:"export_paths,omitempty"` // format -> file path
}

// ReportCacheIndex maintains an index of all cached reports
type ReportCacheIndex struct {
	Reports []ReportCacheEntry `json:"reports"`
}

// ReportCacheEntry represents a summary entry in the cache index
type ReportCacheEntry struct {
	ID           string            `json:"id"`
	Date         string            `json:"date"`         // YYYY-MM-DD format
	Format       string            `json:"format"`
	GeneratedAt  time.Time         `json:"generated_at"`
	InputHash    string            `json:"input_hash"`
	IssueCount   int               `json:"issue_count"`
	CommentCount int               `json:"comment_count"`
	WorklogCount int               `json:"worklog_count"`
	LLMUsed      bool              `json:"llm_used"`
	ExportPaths  map[string]string `json:"export_paths,omitempty"`
}

// CacheManager handles report caching operations
type CacheManager struct {
	cacheDir string
}

// NewCacheManager creates a new cache manager
func NewCacheManager() (*CacheManager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(homeDir, ".my-day", "reports")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &CacheManager{cacheDir: cacheDir}, nil
}

// GenerateReportID creates a unique ID for a report based on input parameters
func (cm *CacheManager) GenerateReportID(config *Config, issues []jira.Issue, comments map[string][]jira.Comment, worklogs []jira.WorklogEntry, targetDate time.Time) string {
	// Create a hash based on all input parameters that affect the report
	hasher := sha256.New()
	
	// Include date
	hasher.Write([]byte(targetDate.Format("2006-01-02")))
	
	// Include config parameters that affect output
	configData := fmt.Sprintf("format:%s|llm:%t|mode:%s|model:%s|detailed:%t|debug:%t|quality:%t|verbose:%t|field:%s",
		config.Format, config.LLMEnabled, config.LLMMode, config.LLMModel, 
		config.Detailed, config.Debug, config.ShowQuality, config.Verbose, config.GroupByField)
	hasher.Write([]byte(configData))
	
	// Include issue IDs and update times (sorted for consistency)
	var issueData []string
	for _, issue := range issues {
		issueData = append(issueData, fmt.Sprintf("%s:%s", issue.Key, issue.Fields.Updated.Time.Format(time.RFC3339)))
	}
	sort.Strings(issueData)
	hasher.Write([]byte(strings.Join(issueData, "|")))
	
	// Include comment data (sorted for consistency)
	var commentData []string
	for issueKey, issueComments := range comments {
		for _, comment := range issueComments {
			commentData = append(commentData, fmt.Sprintf("%s:%s:%s", issueKey, comment.ID, comment.Created.Time.Format(time.RFC3339)))
		}
	}
	sort.Strings(commentData)
	hasher.Write([]byte(strings.Join(commentData, "|")))
	
	// Include worklog data (sorted for consistency)
	var worklogData []string
	for _, worklog := range worklogs {
		worklogData = append(worklogData, fmt.Sprintf("%s:%s", worklog.IssueID, worklog.Started.Time.Format(time.RFC3339)))
	}
	sort.Strings(worklogData)
	hasher.Write([]byte(strings.Join(worklogData, "|")))
	
	hash := hex.EncodeToString(hasher.Sum(nil))
	
	// Create a readable ID with date and hash prefix
	return fmt.Sprintf("%s_%s", targetDate.Format("2006-01-02"), hash[:12])
}

// SaveReport saves a generated report to cache
func (cm *CacheManager) SaveReport(reportID string, config *Config, content string, targetDate time.Time, 
	issueCount, commentCount, worklogCount int, generationTimeMs int64, inputHash string) error {
	
	cache := &ReportCache{
		ID:               reportID,
		Date:             targetDate,
		Config:           config,
		Content:          content,
		Format:           config.Format,
		GeneratedAt:      time.Now(),
		InputHash:        inputHash,
		IssueCount:       issueCount,
		CommentCount:     commentCount,
		WorklogCount:     worklogCount,
		LLMUsed:          config.LLMEnabled,
		GenerationTimeMs: generationTimeMs,
		ExportPaths:      make(map[string]string),
	}
	
	// Save the full report cache
	cacheFile := filepath.Join(cm.cacheDir, fmt.Sprintf("%s.json", reportID))
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report cache: %w", err)
	}
	
	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write report cache: %w", err)
	}
	
	// Update the index
	if err := cm.updateIndex(cache); err != nil {
		return fmt.Errorf("failed to update cache index: %w", err)
	}
	
	return nil
}

// LoadReport loads a cached report by ID
func (cm *CacheManager) LoadReport(reportID string) (*ReportCache, error) {
	cacheFile := filepath.Join(cm.cacheDir, fmt.Sprintf("%s.json", reportID))
	
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read report cache: %w", err)
	}
	
	var cache ReportCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("failed to unmarshal report cache: %w", err)
	}
	
	return &cache, nil
}

// FindReport finds a cached report for the given parameters
func (cm *CacheManager) FindReport(config *Config, issues []jira.Issue, comments map[string][]jira.Comment, worklogs []jira.WorklogEntry, targetDate time.Time) (*ReportCache, error) {
	reportID := cm.GenerateReportID(config, issues, comments, worklogs, targetDate)
	return cm.LoadReport(reportID)
}

// ListReports returns all cached reports, optionally filtered by date range
func (cm *CacheManager) ListReports(fromDate, toDate *time.Time) ([]ReportCacheEntry, error) {
	index, err := cm.loadIndex()
	if err != nil {
		return nil, err
	}
	
	var filtered []ReportCacheEntry
	for _, entry := range index.Reports {
		entryDate, err := time.Parse("2006-01-02", entry.Date)
		if err != nil {
			continue // Skip invalid entries
		}
		
		include := true
		if fromDate != nil && entryDate.Before(*fromDate) {
			include = false
		}
		if toDate != nil && entryDate.After(*toDate) {
			include = false
		}
		
		if include {
			filtered = append(filtered, entry)
		}
	}
	
	// Sort by date (most recent first)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].GeneratedAt.After(filtered[j].GeneratedAt)
	})
	
	return filtered, nil
}

// DeleteReport removes a cached report
func (cm *CacheManager) DeleteReport(reportID string) error {
	// Remove the cache file
	cacheFile := filepath.Join(cm.cacheDir, fmt.Sprintf("%s.json", reportID))
	if err := os.Remove(cacheFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove cache file: %w", err)
	}
	
	// Update the index
	index, err := cm.loadIndex()
	if err != nil {
		return err
	}
	
	// Remove entry from index
	for i, entry := range index.Reports {
		if entry.ID == reportID {
			index.Reports = append(index.Reports[:i], index.Reports[i+1:]...)
			break
		}
	}
	
	return cm.saveIndex(index)
}

// ClearCache removes all cached reports
func (cm *CacheManager) ClearCache() error {
	// Remove all cache files
	files, err := filepath.Glob(filepath.Join(cm.cacheDir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to glob cache files: %w", err)
	}
	
	for _, file := range files {
		if filepath.Base(file) == "index.json" {
			continue // Don't remove index yet
		}
		if err := os.Remove(file); err != nil {
			return fmt.Errorf("failed to remove cache file %s: %w", file, err)
		}
	}
	
	// Clear the index
	index := &ReportCacheIndex{Reports: []ReportCacheEntry{}}
	return cm.saveIndex(index)
}

// UpdateExportPath updates the export path for a specific format in a cached report
func (cm *CacheManager) UpdateExportPath(reportID, format, path string) error {
	cache, err := cm.LoadReport(reportID)
	if err != nil {
		return err
	}
	
	if cache.ExportPaths == nil {
		cache.ExportPaths = make(map[string]string)
	}
	cache.ExportPaths[format] = path
	
	// Save updated cache
	cacheFile := filepath.Join(cm.cacheDir, fmt.Sprintf("%s.json", reportID))
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated report cache: %w", err)
	}
	
	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write updated report cache: %w", err)
	}
	
	// Update index
	return cm.updateIndex(cache)
}

// GetCacheStats returns statistics about the cache
func (cm *CacheManager) GetCacheStats() (map[string]interface{}, error) {
	index, err := cm.loadIndex()
	if err != nil {
		return nil, err
	}
	
	stats := make(map[string]interface{})
	stats["total_reports"] = len(index.Reports)
	
	// Group by date
	dateGroups := make(map[string]int)
	formatGroups := make(map[string]int)
	llmUsageCount := 0
	
	for _, entry := range index.Reports {
		dateGroups[entry.Date]++
		formatGroups[entry.Format]++
		if entry.LLMUsed {
			llmUsageCount++
		}
	}
	
	stats["reports_by_date"] = dateGroups
	stats["reports_by_format"] = formatGroups
	stats["llm_usage_count"] = llmUsageCount
	stats["cache_directory"] = cm.cacheDir
	
	// Calculate cache size
	var totalSize int64
	files, _ := filepath.Glob(filepath.Join(cm.cacheDir, "*.json"))
	for _, file := range files {
		if info, err := os.Stat(file); err == nil {
			totalSize += info.Size()
		}
	}
	stats["cache_size_bytes"] = totalSize
	
	return stats, nil
}

// loadIndex loads the cache index
func (cm *CacheManager) loadIndex() (*ReportCacheIndex, error) {
	indexFile := filepath.Join(cm.cacheDir, "index.json")
	
	data, err := os.ReadFile(indexFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Return empty index if file doesn't exist
			return &ReportCacheIndex{Reports: []ReportCacheEntry{}}, nil
		}
		return nil, fmt.Errorf("failed to read cache index: %w", err)
	}
	
	var index ReportCacheIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache index: %w", err)
	}
	
	return &index, nil
}

// saveIndex saves the cache index
func (cm *CacheManager) saveIndex(index *ReportCacheIndex) error {
	indexFile := filepath.Join(cm.cacheDir, "index.json")
	
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache index: %w", err)
	}
	
	return os.WriteFile(indexFile, data, 0644)
}

// updateIndex updates the cache index with a new or updated report
func (cm *CacheManager) updateIndex(cache *ReportCache) error {
	index, err := cm.loadIndex()
	if err != nil {
		return err
	}
	
	// Create entry for index
	entry := ReportCacheEntry{
		ID:           cache.ID,
		Date:         cache.Date.Format("2006-01-02"),
		Format:       cache.Format,
		GeneratedAt:  cache.GeneratedAt,
		InputHash:    cache.InputHash,
		IssueCount:   cache.IssueCount,
		CommentCount: cache.CommentCount,
		WorklogCount: cache.WorklogCount,
		LLMUsed:      cache.LLMUsed,
		ExportPaths:  cache.ExportPaths,
	}
	
	// Remove existing entry with same ID if it exists
	for i, existing := range index.Reports {
		if existing.ID == cache.ID {
			index.Reports = append(index.Reports[:i], index.Reports[i+1:]...)
			break
		}
	}
	
	// Add new entry
	index.Reports = append(index.Reports, entry)
	
	return cm.saveIndex(index)
}