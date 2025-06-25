package jira

import "time"

// Issue represents a Jira issue
type Issue struct {
	ID     string `json:"id"`
	Key    string `json:"key"`
	Self   string `json:"self"`
	Fields Fields `json:"fields"`
}

// Fields represents Jira issue fields
type Fields struct {
	Summary     string      `json:"summary"`
	Description string      `json:"description"`
	Status      Status      `json:"status"`
	Priority    Priority    `json:"priority"`
	IssueType   IssueType   `json:"issuetype"`
	Project     Project     `json:"project"`
	Assignee    *User       `json:"assignee"`
	Reporter    User        `json:"reporter"`
	Created     time.Time   `json:"created"`
	Updated     time.Time   `json:"updated"`
	Resolution  *Resolution `json:"resolution"`
	Labels      []string    `json:"labels"`
}

// Status represents issue status
type Status struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    struct {
		ID   string `json:"id"`
		Key  string `json:"key"`
		Name string `json:"name"`
	} `json:"statusCategory"`
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
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// WorklogEntry represents a worklog entry
type WorklogEntry struct {
	ID      string    `json:"id"`
	Author  User      `json:"author"`
	Comment string    `json:"comment"`
	Started time.Time `json:"started"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
	IssueID string    `json:"issueId"`
}