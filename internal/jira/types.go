package jira

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// JiraTime represents a time field that can handle Jira's various date formats
type JiraTime struct {
	time.Time
}

// UnmarshalJSON handles Jira's date format variations
func (jt *JiraTime) UnmarshalJSON(data []byte) error {
	// Remove quotes
	timeStr := strings.Trim(string(data), `"`)
	
	if timeStr == "null" || timeStr == "" {
		jt.Time = time.Time{}
		return nil
	}
	
	// Try different Jira date formats
	formats := []string{
		"2006-01-02T15:04:05.000-0700",  // Jira format with milliseconds and timezone
		"2006-01-02T15:04:05.000Z",      // Jira format with milliseconds and Z
		"2006-01-02T15:04:05-0700",      // Jira format without milliseconds
		"2006-01-02T15:04:05Z",          // ISO format with Z
		time.RFC3339,                     // Standard RFC3339
		time.RFC3339Nano,                 // RFC3339 with nanoseconds
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			jt.Time = t
			return nil
		}
	}
	
	// If all parsing attempts fail, return error
	return &time.ParseError{
		Layout: "Jira time formats",
		Value:  timeStr,
	}
}

// MarshalJSON converts JiraTime back to JSON
func (jt JiraTime) MarshalJSON() ([]byte, error) {
	if jt.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(jt.Time.Format(time.RFC3339))
}

// Issue represents a Jira issue
type Issue struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Self   string `json:"self"`
	Fields Fields `json:"fields"`
}

// JiraDescription represents a description field that can be string or object
type JiraDescription struct {
	Text string
}

// UnmarshalJSON handles Jira's description field variations
func (jd *JiraDescription) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		jd.Text = str
		return nil
	}
	
	// If string fails, try as object with content
	var obj struct {
		Content []struct {
			Content []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"content"`
	}
	
	if err := json.Unmarshal(data, &obj); err == nil {
		// Extract text from Atlassian Document Format
		var text []string
		for _, content := range obj.Content {
			for _, innerContent := range content.Content {
				if innerContent.Text != "" {
					text = append(text, innerContent.Text)
				}
			}
		}
		jd.Text = strings.Join(text, " ")
		return nil
	}
	
	// If both fail, set empty string
	jd.Text = ""
	return nil
}

// MarshalJSON converts JiraDescription back to JSON
func (jd JiraDescription) MarshalJSON() ([]byte, error) {
	return json.Marshal(jd.Text)
}

// String returns the text content
func (jd JiraDescription) String() string {
	return jd.Text
}

// Fields represents Jira issue fields
type Fields struct {
	Summary      string                    `json:"summary"`
	Description  JiraDescription           `json:"description"`
	Status       Status                    `json:"status"`
	Priority     Priority                  `json:"priority"`
	IssueType    IssueType                 `json:"issuetype"`
	Project      Project                   `json:"project"`
	Assignee     *User                     `json:"assignee"`
	Reporter     User                      `json:"reporter"`
	Created      JiraTime                  `json:"created"`
	Updated      JiraTime                  `json:"updated"`
	Resolution   *Resolution               `json:"resolution"`
	Labels       []string                  `json:"labels"`
	CustomFields map[string]*CustomField   `json:"-"` // Store all custom fields dynamically
}

// StatusCategory represents a status category that can have string or number ID
type StatusCategory struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

// UnmarshalJSON handles both string and number IDs
func (sc *StatusCategory) UnmarshalJSON(data []byte) error {
	// Try to unmarshal into a temporary struct with interface{} for ID
	var temp struct {
		ID   interface{} `json:"id"`
		Key  string      `json:"key"`
		Name string      `json:"name"`
	}
	
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	
	// Convert ID to string regardless of original type
	switch v := temp.ID.(type) {
	case string:
		sc.ID = v
	case float64:
		sc.ID = fmt.Sprintf("%.0f", v)
	case int:
		sc.ID = fmt.Sprintf("%d", v)
	default:
		sc.ID = ""
	}
	
	sc.Key = temp.Key
	sc.Name = temp.Name
	
	return nil
}

// Status represents issue status
type Status struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Category    StatusCategory `json:"statusCategory"`
}

// Priority represents issue priority
type Priority struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// IssueType represents issue type
type IssueType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Project represents a Jira project
type Project struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Name string `json:"name"`
}

// User represents a Jira user
type User struct {
	AccountID    string `json:"accountId"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

// Resolution represents issue resolution
type Resolution struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// SearchResponse represents Jira search API response
type SearchResponse struct {
	Expand     string  `json:"expand"`
	StartAt    int     `json:"startAt"`
	MaxResults int     `json:"maxResults"`
	Total      int     `json:"total"`
	Issues     []Issue `json:"issues"`
}

// AuthResponse represents OAuth token response
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// AuthInfo represents stored authentication information
type AuthInfo struct {
	AuthType  string         `json:"auth_type"` // always "token"
	APIToken  *APITokenAuth  `json:"api_token"`
	ExpiresAt time.Time      `json:"expires_at"`
}


// WorklogEntry represents a worklog entry
type WorklogEntry struct {
	ID      string   `json:"id"`
	Author  User     `json:"author"`
	Comment string   `json:"comment"`
	Started JiraTime `json:"started"`
	Created JiraTime `json:"created"`
	Updated JiraTime `json:"updated"`
	IssueID string   `json:"issueId"`
}

// Comment represents a comment on an issue
type Comment struct {
	ID      string          `json:"id"`
	Author  User            `json:"author"`
	Body    JiraDescription `json:"body"`
	Created JiraTime        `json:"created"`
	Updated JiraTime        `json:"updated"`
}

// CustomField represents a Jira custom field that can have various value types
type CustomField struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
}

// GetStringValue returns the string representation of the custom field value
func (cf *CustomField) GetStringValue() string {
	if cf.Value == nil {
		return ""
	}
	
	switch v := cf.Value.(type) {
	case string:
		return v
	case map[string]interface{}:
		// Handle complex objects like {id: "123", value: "Squad Name"}
		if val, ok := v["value"].(string); ok {
			return val
		}
		if val, ok := v["displayName"].(string); ok {
			return val
		}
		if val, ok := v["name"].(string); ok {
			return val
		}
		return ""
	default:
		return fmt.Sprintf("%v", v)
	}
}

// UnmarshalJSON handles dynamic custom field unmarshaling for Fields
func (f *Fields) UnmarshalJSON(data []byte) error {
	// First unmarshal into a temporary map to capture all fields
	var temp map[string]interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}
	
	// Initialize custom fields map
	f.CustomFields = make(map[string]*CustomField)
	
	// Handle standard fields by unmarshaling the struct normally
	type FieldsAlias Fields // Prevent infinite recursion
	var alias FieldsAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	
	// Copy standard fields
	f.Summary = alias.Summary
	f.Description = alias.Description
	f.Status = alias.Status
	f.Priority = alias.Priority
	f.IssueType = alias.IssueType
	f.Project = alias.Project
	f.Assignee = alias.Assignee
	f.Reporter = alias.Reporter
	f.Created = alias.Created
	f.Updated = alias.Updated
	f.Resolution = alias.Resolution
	f.Labels = alias.Labels
	
	// Extract custom fields (they start with "customfield_")
	for key, value := range temp {
		if strings.HasPrefix(key, "customfield_") && value != nil {
			f.CustomFields[key] = &CustomField{
				ID:    key,
				Value: value,
			}
		}
	}
	
	return nil
}

// GetCustomFieldValue returns the value of a custom field by field ID
func (f *Fields) GetCustomFieldValue(fieldID string) string {
	if cf, exists := f.CustomFields[fieldID]; exists {
		return cf.GetStringValue()
	}
	return ""
}