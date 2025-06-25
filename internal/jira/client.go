package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

// Client represents a Jira API client
type Client struct {
	baseURL     string
	httpClient  *http.Client
	authManager *AuthManager
}

// NewClient creates a new Jira client
func NewClient(baseURL, clientID, clientSecret, redirectURI string) *Client {
	authManager := NewAuthManager(clientID, clientSecret, redirectURI)
	
	return &Client{
		baseURL:     strings.TrimSuffix(baseURL, "/"),
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		authManager: authManager,
	}
}

// GetAuthManager returns the authentication manager
func (c *Client) GetAuthManager() *AuthManager {
	return c.authManager
}

// SearchIssues searches for issues using JQL
func (c *Client) SearchIssues(ctx context.Context, jql string, maxResults int) (*SearchResponse, error) {
	token, err := c.authManager.GetValidToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	// Build search URL
	searchURL := fmt.Sprintf("%s/rest/api/3/search", c.baseURL)
	params := url.Values{
		"jql":        {jql},
		"maxResults": {fmt.Sprintf("%d", maxResults)},
		"fields":     {"summary,description,status,priority,issuetype,project,assignee,reporter,created,updated,resolution,labels"},
	}

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var searchResponse SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &searchResponse, nil
}

// GetIssuesByProjects retrieves issues for specific projects
func (c *Client) GetIssuesByProjects(ctx context.Context, projectKeys []string, maxResults int) (*SearchResponse, error) {
	if len(projectKeys) == 0 {
		return &SearchResponse{Issues: []Issue{}}, nil
	}

	// Build JQL query for project keys
	projectFilter := strings.Join(projectKeys, ",")
	jql := fmt.Sprintf("project in (%s) ORDER BY updated DESC", projectFilter)

	return c.SearchIssues(ctx, jql, maxResults)
}

// GetMyRecentIssues retrieves issues assigned to or reported by the current user
func (c *Client) GetMyRecentIssues(ctx context.Context, projectKeys []string, maxResults int) (*SearchResponse, error) {
	var jqlParts []string

	// Add project filter if specified
	if len(projectKeys) > 0 {
		projectFilter := strings.Join(projectKeys, ",")
		jqlParts = append(jqlParts, fmt.Sprintf("project in (%s)", projectFilter))
	}

	// Add user filter (assignee or reporter)
	jqlParts = append(jqlParts, "(assignee = currentUser() OR reporter = currentUser())")

	// Add time filter for recent activity
	jqlParts = append(jqlParts, "updated >= -7d")

	jql := strings.Join(jqlParts, " AND ") + " ORDER BY updated DESC"

	return c.SearchIssues(ctx, jql, maxResults)
}

// GetMyWorklog retrieves worklog entries for the current user
func (c *Client) GetMyWorklog(ctx context.Context, since time.Time) ([]WorklogEntry, error) {
	token, err := c.authManager.GetValidToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	// Get current user info first
	userInfo, err := c.getCurrentUser(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// Search for issues with worklog by current user
	sinceStr := since.Format("2006-01-02")
	jql := fmt.Sprintf("worklogAuthor = currentUser() AND worklogDate >= %s ORDER BY updated DESC", sinceStr)

	searchResponse, err := c.SearchIssues(ctx, jql, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to search worklog issues: %w", err)
	}

	var worklogEntries []WorklogEntry
	for _, issue := range searchResponse.Issues {
		worklogs, err := c.getIssueWorklogs(ctx, token, issue.Key, userInfo.AccountID, since)
		if err != nil {
			continue // Skip issues where we can't get worklogs
		}
		worklogEntries = append(worklogEntries, worklogs...)
	}

	return worklogEntries, nil
}

// getCurrentUser gets information about the current authenticated user
func (c *Client) getCurrentUser(ctx context.Context, token *oauth2.Token) (*User, error) {
	url := fmt.Sprintf("%s/rest/api/3/myself", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// getIssueWorklogs retrieves worklog entries for a specific issue
func (c *Client) getIssueWorklogs(ctx context.Context, token *oauth2.Token, issueKey, userAccountID string, since time.Time) ([]WorklogEntry, error) {
	url := fmt.Sprintf("%s/rest/api/3/issue/%s/worklog", c.baseURL, issueKey)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get worklogs: status %d", resp.StatusCode)
	}

	var worklogResponse struct {
		Worklogs []WorklogEntry `json:"worklogs"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&worklogResponse); err != nil {
		return nil, err
	}

	// Filter worklogs by user and date
	var userWorklogs []WorklogEntry
	for _, worklog := range worklogResponse.Worklogs {
		if worklog.Author.AccountID == userAccountID && worklog.Started.After(since) {
			worklog.IssueID = issueKey
			userWorklogs = append(userWorklogs, worklog)
		}
	}

	return userWorklogs, nil
}

// TestConnection tests the connection to Jira
func (c *Client) TestConnection(ctx context.Context) error {
	token, err := c.authManager.GetValidToken(ctx)
	if err != nil {
		return fmt.Errorf("authentication required: %w", err)
	}

	url := fmt.Sprintf("%s/rest/api/3/myself", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to Jira: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Jira API returned status %d", resp.StatusCode)
	}

	return nil
}