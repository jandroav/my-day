package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client represents a Jira API client
type Client struct {
	baseURL     string
	httpClient  *http.Client
	authManager *AuthManager
}

// NewClient creates a new Jira client with API token authentication
func NewClient(baseURL, email, token string) *Client {
	authManager := NewAuthManager(email, token)
	
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

// getAuthenticatedClient returns an HTTP client with API token authentication
func (c *Client) getAuthenticatedClient(ctx context.Context) (*http.Client, error) {
	apiToken, err := c.authManager.LoadAPIToken()
	if err != nil {
		return nil, fmt.Errorf("API token authentication required: %w", err)
	}
	
	// Create HTTP client with basic auth transport
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &apiTokenTransport{
			email:    apiToken.Email,
			token:    apiToken.Token,
			base:     http.DefaultTransport,
		},
	}
	return client, nil
}

// apiTokenTransport implements HTTP transport with API token authentication
type apiTokenTransport struct {
	email string
	token string
	base  http.RoundTripper
}

func (t *apiTokenTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Set basic auth header
	req.SetBasicAuth(t.email, t.token)
	return t.base.RoundTrip(req)
}

// SearchIssues searches for issues using JQL
func (c *Client) SearchIssues(ctx context.Context, jql string, maxResults int) (*SearchResponse, error) {
	client, err := c.getAuthenticatedClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	// Build search URL using direct Jira instance URL
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

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
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

// GetMyWorklog retrieves worklog entries for the current user
func (c *Client) GetMyWorklog(ctx context.Context, since time.Time) ([]WorklogEntry, error) {
	// Get current user info first
	userInfo, err := c.getCurrentUser(ctx)
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
		worklogs, err := c.getIssueWorklogs(ctx, issue.Key, userInfo.AccountID, since)
		if err != nil {
			continue // Skip issues where we can't get worklogs
		}
		worklogEntries = append(worklogEntries, worklogs...)
	}

	return worklogEntries, nil
}

// GetCurrentUser gets information about the current authenticated user
func (c *Client) GetCurrentUser(ctx context.Context) (*User, error) {
	return c.getCurrentUser(ctx)
}

// getCurrentUser gets information about the current authenticated user
func (c *Client) getCurrentUser(ctx context.Context) (*User, error) {
	client, err := c.getAuthenticatedClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	url := fmt.Sprintf("%s/rest/api/3/myself", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
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

// GetMyIssuesWithTodaysComments retrieves issues where the current user added comments recently
func (c *Client) GetMyIssuesWithTodaysComments(ctx context.Context, projectKeys []string, maxResults int, since time.Time) (*SearchResponse, error) {
	var jqlParts []string
	
	// Add project filter if specified
	if len(projectKeys) > 0 {
		projectFilter := strings.Join(projectKeys, ",")
		jqlParts = append(jqlParts, fmt.Sprintf("project in (%s)", projectFilter))
	}
	
	// Search for issues updated since the specified time - we'll filter comments afterward
	sinceDate := since.Format("2006-01-02")
	jqlParts = append(jqlParts, fmt.Sprintf("updated >= %s", sinceDate))
	
	jql := strings.Join(jqlParts, " AND ")
	jql += " ORDER BY updated DESC"

	return c.SearchIssues(ctx, jql, maxResults)
}

// GetIssueComments retrieves comments for a specific issue
func (c *Client) GetIssueComments(ctx context.Context, issueKey string) ([]Comment, error) {
	client, err := c.getAuthenticatedClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	url := fmt.Sprintf("%s/rest/api/3/issue/%s/comment", c.baseURL, issueKey)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get comments: status %d", resp.StatusCode)
	}

	var response struct {
		Comments []Comment `json:"comments"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return response.Comments, nil
}

// getIssueWorklogs retrieves worklog entries for a specific issue
func (c *Client) getIssueWorklogs(ctx context.Context, issueKey string, userAccountID string, since time.Time) ([]WorklogEntry, error) {
	client, err := c.getAuthenticatedClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	url := fmt.Sprintf("%s/rest/api/3/issue/%s/worklog", c.baseURL, issueKey)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get worklogs: status %d", resp.StatusCode)
	}

	var response struct {
		Worklogs []WorklogEntry `json:"worklogs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	// Filter worklogs by user and date
	var filteredWorklogs []WorklogEntry
	for _, worklog := range response.Worklogs {
		if worklog.Author.AccountID == userAccountID && worklog.Started.Time.After(since) {
			filteredWorklogs = append(filteredWorklogs, worklog)
		}
	}

	return filteredWorklogs, nil
}

// TestConnection tests the connection to Jira
func (c *Client) TestConnection(ctx context.Context) error {
	_, err := c.getCurrentUser(ctx)
	return err
}